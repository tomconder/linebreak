// Package linebreak wraps text into lines of a given width.
//
// Width is counted in runes. A word longer than the width is placed on a
// line by itself, so that line is wider than the width.
//
// In KnuthPlass and Justify, a line that ends at a soft hyphen includes the
// added "-". A word fragment that fills the width exactly makes that line one
// rune wider. A word with soft hyphens is broken at them even when it is
// wider than the width.
package linebreak

import (
	"strings"
	"unicode/utf8"
)

/*
Greedy wraps the text into lines that do not exceed the width. It uses the greedy line-breaking algorithm.
*/
func Greedy(text string, width int) []string {
	if width <= 0 {
		return []string{}
	}

	words := strings.Fields(text)
	n := len(words)

	if n == 0 {
		return []string{}
	}

	lineLength := 0

	var groups []string
	var sb strings.Builder

	for _, word := range words {
		wordLength := utf8.RuneCountInString(word)

		if lineLength == 0 {
			sb.WriteString(word)
			lineLength = wordLength
		} else if lineLength+1+wordLength <= width {
			sb.WriteString(" ")
			sb.WriteString(word)
			lineLength += 1 + wordLength
		} else {
			groups = append(groups, sb.String())

			sb.Reset()
			sb.WriteString(word)
			lineLength = wordLength
		}
	}

	if sb.Len() > 0 {
		groups = append(groups, sb.String())
	}

	return groups
}
