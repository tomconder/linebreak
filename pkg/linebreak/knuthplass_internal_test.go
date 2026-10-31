package linebreak

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// soft replaces each "|" with a soft hyphen.
func soft(s string) string {
	return strings.ReplaceAll(s, "|", softHyphen)
}

func TestTokenize(t *testing.T) {
	assert.Equal(t, []item{
		{kind: box, text: "ex", width: 2},
		{kind: penalty, width: 1, penalty: hyphenPenalty, flagged: true},
		{kind: box, text: "tra", width: 3},
		{kind: glue, width: 1, stretch: 2},
		{kind: box, text: "day", width: 3},
		{kind: glue, infinite: true},
		{kind: penalty, forced: true},
	}, tokenize(soft("ex|tra day"), 2))

	// soft hyphens with no text next to them are dropped
	assert.Equal(t, []item{
		{kind: box, text: "a", width: 1},
		{kind: glue, infinite: true},
		{kind: penalty, forced: true},
	}, tokenize(soft("| a||"), 0))

	assert.Nil(t, tokenize("  ", 0))
	assert.Equal(t, 4, tokenize("café", 0)[0].width)
}

func TestFit(t *testing.T) {
	assert.Equal(t, lineFit{feasible: true, ratio: 1.5, badness: 337.5}, fit(8, 4, 2, 14, false))
	assert.Equal(t, lineFit{feasible: true}, fit(14, 4, 2, 14, false))
	assert.Equal(t, lineFit{feasible: true}, fit(5, 0, 1, 14, true))
	assert.Equal(t, lineFit{feasible: true, ratio: -1, badness: infBad}, fit(20, 0, 1, 14, false))
	assert.Equal(t, lineFit{}, fit(20, 0, 2, 14, false))
	assert.Equal(t, lineFit{feasible: true, ratio: math.Inf(1), badness: infBad}, fit(5, 0, 1, 14, false))
	assert.Equal(t, float64(infBad), fit(1, 1, 2, 14, false).badness) // r = 13, capped
}

func TestFitness(t *testing.T) {
	for r, want := range map[float64]int{-1: 0, -0.5: 1, 0: 1, 0.5: 1, 0.75: 2, 1: 2, 1.5: 3, math.Inf(1): 3} {
		assert.Equal(t, want, fitness(r), "r = %v", r)
	}
}

func TestDemerits(t *testing.T) {
	space := item{kind: glue, width: 1}
	hyphen := item{kind: penalty, width: 1, penalty: hyphenPenalty, flagged: true}
	end := item{kind: penalty, forced: true}

	// (10 + 337.5)², plus fitness demerits because class 3 follows class 1
	assert.Equal(t, 130756.25, demerits(lineFit{feasible: true, ratio: 1.5, badness: 337.5}, space, 3, 1, false))
	// (10 + 0 + 50)², plus flagged demerits for two hyphen lines in a row
	assert.Equal(t, 6600.0, demerits(lineFit{feasible: true}, hyphen, 1, 1, true))
	// the forced break adds no penalty: (10 + 0)²
	assert.Equal(t, 100.0, demerits(lineFit{feasible: true}, end, 1, 2, false))
	// a negative penalty is subtracted: (10 + 0)² - 20²
	assert.Equal(t, -300.0, demerits(lineFit{feasible: true}, item{kind: penalty, penalty: -20}, 1, 1, false))
}
