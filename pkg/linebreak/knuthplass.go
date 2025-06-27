package linebreak

import (
	"math"
	"strings"
)

/*
KnuthPlass wraps the text into lines that not exceed the width. It uses the Knuth-Plass line-breaking algorithm.
*/
func KnuthPlass(text string, width int) []string {
	if width <= 0 {
		return []string{}
	}

	words := strings.Fields(text)
	n := len(words)
	if n == 0 {
		return []string{}
	}

	lengths := make([]int, n+1)
	for i, word := range words {
		lengths[i+1] = lengths[i] + len(word)
	}

	minCost := make([]float64, n+1)
	for i := 0; i <= n; i++ {
		minCost[i] = math.Inf(1)
	}
	minCost[0] = 0

	breakpoints := make([]int, n+1)

	for i := 0; i < n; i++ {
		for j := i + 1; j <= n; j++ {
			lineLength := lengths[j] - lengths[i]
			spaces := j - i - 1
			totalLen := lineLength + spaces

			if totalLen > width {
				break
			}

			slack := width - totalLen
			var cost float64
			if j == n {
				cost = 0
			} else {
				cost = float64(slack * slack)
			}

			if minCost[i]+cost < minCost[j] {
				minCost[j] = minCost[i] + cost
				breakpoints[j] = i
			}
		}
	}

	var breakList []int
	for j := n; j > 0; j = breakpoints[j] {
		breakList = append(breakList, j)
	}
	breakList = append(breakList, 0)

	for i, j := 0, len(breakList)-1; i < j; i, j = i+1, j-1 {
		breakList[i], breakList[j] = breakList[j], breakList[i]
	}

	groups := make([]string, 0, len(breakList)-1)
	for i := 0; i < len(breakList)-1; i++ {
		start, end := breakList[i], breakList[i+1]
		line := strings.Join(words[start:end], " ")
		groups = append(groups, line)
	}

	return groups
}
