package linebreak

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/tomconder/linebreak/pkg/linebreak"
)

// cost is the sum of squared slack over every line except the last.
// It is the cost that MinRaggedness minimizes.
func cost(lines []string, width int) int {
	total := 0
	for _, line := range lines[:max(len(lines)-1, 0)] {
		slack := width - utf8.RuneCountInString(line)
		total += slack * slack
	}
	return total
}

func checkLines(t *testing.T, name, text string, width int, lines []string) {
	if got, want := strings.Fields(strings.Join(lines, " ")), strings.Fields(text); !slices.Equal(got, want) {
		t.Fatalf("%s: words changed: got %q, want %q", name, got, want)
	}
	for _, line := range lines {
		if utf8.RuneCountInString(line) > width && len(strings.Fields(line)) > 1 {
			t.Fatalf("%s: line %q is wider than %d and has more than one word", name, line, width)
		}
	}
}

func FuzzLineBreak(f *testing.F) {
	f.Add("The lazy yellow dog was caught by the slow red fox as he lay sleeping in the sun", 14)
	f.Add("a supercalifragilisticexpialidocious word", 10)
	f.Add("café au lait", 7)
	f.Add("", 5)

	f.Fuzz(func(t *testing.T, text string, width int) {
		// ponytail: keep width small so slack*slack cannot overflow and the fuzzer finds short lines
		width = width%100 + 1
		if width <= 0 {
			width += 100
		}

		greedy := linebreak.Greedy(text, width)
		kp := linebreak.MinRaggedness(text, width)

		checkLines(t, "Greedy", text, width, greedy)
		checkLines(t, "MinRaggedness", text, width, kp)

		if kc, gc := cost(kp, width), cost(greedy, width); kc > gc {
			t.Fatalf("MinRaggedness cost %d is greater than Greedy cost %d: minrag %q, greedy %q", kc, gc, kp, greedy)
		}
	})
}
