package linebreak

import (
	"strings"
	"unicode/utf8"
)

/*
Justify wraps the text with the Knuth-Plass algorithm, then adds spaces between
the words of each line except the last, so that each line is exactly width runes
wide. A line with one word, or one that is wider than width, is not changed.
*/
func Justify(text string, width int) []string {
	all := segments(text, width, justified)
	lines := make([]string, 0, len(all))
	for i, segs := range all {
		gaps := len(segs) - 1
		natural := gaps
		for _, seg := range segs {
			natural += utf8.RuneCountInString(seg)
		}
		if i == len(all)-1 || gaps == 0 || natural >= width {
			lines = append(lines, strings.Join(segs, " "))
			continue
		}
		extra := width - natural
		var sb strings.Builder
		for j, seg := range segs {
			if j > 0 {
				n := 1 + extra/gaps
				if j-1 < extra%gaps {
					n++
				}
				sb.WriteString(strings.Repeat(" ", n))
			}
			sb.WriteString(seg)
		}
		lines = append(lines, sb.String())
	}
	return lines
}
