package linebreak

import (
	"strings"
)

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
		wordLength := len(word)

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
