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
		{kind: penalty},
	}, tokenize(soft("ex|tra day"), 2))

	// soft hyphens with no text next to them are dropped
	assert.Equal(t, []item{
		{kind: box, text: "a", width: 1},
		{kind: glue, infinite: true},
		{kind: penalty},
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
	end := item{kind: penalty}

	// (10 + 337.5)², plus fitness demerits because class 3 follows class 1
	assert.Equal(t, 130756.25, demerits(lineFit{feasible: true, ratio: 1.5, badness: 337.5}, space, 3, 1, false))
	// (10 + 0)² + 50², plus flagged demerits for two hyphen lines in a row
	assert.Equal(t, 5600.0, demerits(lineFit{feasible: true}, hyphen, 1, 1, true))
	// the forced break adds no penalty: (10 + 0)²
	assert.Equal(t, 100.0, demerits(lineFit{feasible: true}, end, 1, 2, false))
}

// total returns the demerits of the lines that end at ends, or false if a
// line is not feasible.
func total(items []item, ends []int, width, lineEndStretch int) (float64, bool) {
	s := prefixSums(items)
	sum, start, prevClass, prevFlagged := 0.0, 0, 1, false
	for _, end := range ends {
		f := lineAt(items, s, start, end, width, lineEndStretch)
		if !f.feasible {
			return 0, false
		}
		class := fitness(f.ratio)
		sum += demerits(f, items[end], class, prevClass, prevFlagged)
		prevClass, prevFlagged, start = class, items[end].flagged, end+1
	}
	return sum, true
}

// bruteForce tries every set of breaks and returns the lowest total demerits.
func bruteForce(items []item, width, lineEndStretch int) float64 {
	var optional []int
	for i := 0; i < len(items)-1; i++ {
		if legal(items, i) {
			optional = append(optional, i)
		}
	}
	best := math.Inf(1)
	for mask := 0; mask < 1<<len(optional); mask++ {
		var ends []int
		for k, p := range optional {
			if mask&(1<<k) != 0 {
				ends = append(ends, p)
			}
		}
		ends = append(ends, len(items)-1)
		if sum, ok := total(items, ends, width, lineEndStretch); ok && sum < best {
			best = sum
		}
	}
	return best
}

func TestBreaksMatchesBruteForce(t *testing.T) {
	texts := []string{
		"The lazy yellow dog was caught by the slow red fox",
		"a b c d e f g h i j",
		soft("hy|phen|ation is ex|tra|or|di|nary"),
		"a supercalifragilisticexpialidocious word here",
		"aaa bb c dddd ee f gggggg",
	}
	for _, text := range texts {
		for _, set := range []setting{ragged, justified} {
			for width := 1; width <= 30; width++ {
				items := tokenize(text, set.interwordStretch)
				ends, got := breaks(items, width, set.lineEndStretch)
				assert.Equal(t, bruteForce(items, width, set.lineEndStretch), got, "%q %+v width %d", text, set, width)
				// the returned breaks must give the total that breaks reports
				sum, ok := total(items, ends, width, set.lineEndStretch)
				assert.True(t, ok)
				assert.Equal(t, got, sum, "%q %+v width %d: ends %v", text, set, width, ends)
			}
		}
	}
}

func TestBreaksEndAtForcedBreak(t *testing.T) {
	items := tokenize("The lazy yellow dog", ragged.interwordStretch)
	ends, _ := breaks(items, 10, ragged.lineEndStretch)
	assert.Equal(t, []int{3, 8}, ends) // the glue after "lazy", then the forced break
}

func TestLegalSkipsParagraphEndGlue(t *testing.T) {
	items := tokenize("a b", ragged.interwordStretch)
	// items: box "a", glue, box "b", infinite glue, forced penalty
	assert.True(t, legal(items, 1))
	assert.False(t, legal(items, 3))
}
