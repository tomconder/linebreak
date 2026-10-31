package linebreak

import (
	"slices"
	"strings"
	"testing"
	"unicode"
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

// checkWords checks that the lines hold the words of the text, in order.
func checkWords(t *testing.T, name, text string, lines []string) {
	if got, want := strings.Fields(strings.Join(lines, " ")), strings.Fields(text); !slices.Equal(got, want) {
		t.Fatalf("%s: words changed: got %q, want %q", name, got, want)
	}
}

// checkText checks that the lines hold the text once spaces, hyphens and soft
// hyphens are removed. KnuthPlass can split a word at a soft hyphen, so its
// words do not match the input one to one.
func checkText(t *testing.T, name, text string, lines []string) {
	if !utf8.ValidString(text) {
		return // removing a space between invalid bytes can make a valid rune
	}
	strip := func(s string) string {
		return strings.Map(func(r rune) rune {
			if r == '-' || r == 0xAD || unicode.IsSpace(r) {
				return -1
			}
			return r
		}, s)
	}
	if got, want := strip(strings.Join(lines, "")), strip(text); got != want {
		t.Fatalf("%s: text changed: got %q, want %q", name, got, want)
	}
}

// checkWidth checks that no line with more than one word is wider than width.
func checkWidth(t *testing.T, name string, width int, lines []string) {
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
	f.Add(soft("an ex|tra|or|di|nary day"), 10)
	f.Add("", 5)

	f.Fuzz(func(t *testing.T, text string, width int) {
		// ponytail: keep width small so slack*slack cannot overflow and the fuzzer finds short lines
		width = width%100 + 1
		if width <= 0 {
			width += 100
		}

		greedy := linebreak.Greedy(text, width)
		minrag := linebreak.MinRaggedness(text, width)
		kp := linebreak.KnuthPlass(text, width)

		checkWords(t, "Greedy", text, greedy)
		checkWidth(t, "Greedy", width, greedy)
		checkWords(t, "MinRaggedness", text, minrag)
		checkWidth(t, "MinRaggedness", width, minrag)
		checkText(t, "KnuthPlass", text, kp)
		checkWidth(t, "KnuthPlass", width, kp)

		if mc, gc := cost(minrag, width), cost(greedy, width); mc > gc {
			t.Fatalf("MinRaggedness cost %d is greater than Greedy cost %d: minrag %q, greedy %q", mc, gc, minrag, greedy)
		}
	})
}
