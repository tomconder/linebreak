package linebreak

import (
	"math"
	"strings"
)

// lineFit is the result of setting one line.
type lineFit struct {
	feasible bool
	ratio    float64
	badness  float64
}

// fit computes the adjustment ratio and badness of a line. A line wider than
// width is only feasible when it holds one box, so an overlong word gets a
// line of its own.
func fit(natural, stretch, boxes, width int, infinite bool) lineFit {
	switch {
	case natural > width:
		if boxes == 1 {
			return lineFit{feasible: true, ratio: -1, badness: infBad}
		}
		return lineFit{}
	case natural == width || infinite:
		return lineFit{feasible: true}
	case stretch == 0:
		return lineFit{feasible: true, ratio: math.Inf(1), badness: infBad}
	}
	r := float64(width-natural) / float64(stretch)
	return lineFit{feasible: true, ratio: r, badness: math.Min(100*r*r*r, infBad)}
}

// fitness is the Knuth-Plass fitness class: 0 tight, 1 decent, 2 loose,
// 3 very loose.
func fitness(r float64) int {
	switch {
	case r < -0.5:
		return 0
	case r <= 0.5:
		return 1
	case r <= 1:
		return 2
	default:
		return 3
	}
}

// demerits is the cost of one line that ends at break item p, given the
// fitness class and flag of the line before it.
func demerits(f lineFit, p item, class, prevClass int, prevFlagged bool) float64 {
	d := linePenalty + f.badness
	switch {
	case p.kind == penalty && !p.forced && p.penalty >= 0:
		d = (d + float64(p.penalty)) * (d + float64(p.penalty))
	case p.kind == penalty && !p.forced:
		d = d*d - float64(p.penalty*p.penalty)
	default:
		d *= d
	}
	if p.kind == penalty && p.flagged && prevFlagged {
		d += flaggedDemerits
	}
	if class-prevClass > 1 || prevClass-class > 1 {
		d += fitnessDemerits
	}
	return d
}

// sums holds prefix sums over items, so the totals of a line are two lookups.
type sums struct {
	width, stretch, boxes, infinite []int
}

func prefixSums(items []item) sums {
	n := len(items)
	s := sums{make([]int, n+1), make([]int, n+1), make([]int, n+1), make([]int, n+1)}
	for i, it := range items {
		s.width[i+1], s.stretch[i+1], s.boxes[i+1], s.infinite[i+1] = s.width[i], s.stretch[i], s.boxes[i], s.infinite[i]
		switch it.kind {
		case box:
			s.width[i+1] += it.width
			s.boxes[i+1]++
		case glue:
			s.width[i+1] += it.width
			s.stretch[i+1] += it.stretch
			if it.infinite {
				s.infinite[i+1]++
			}
		}
	}
	return s
}

// legal reports whether a line can end at item i: at a penalty, or at glue
// that follows a box.
func legal(items []item, i int) bool {
	switch items[i].kind {
	case penalty:
		return true
	case glue:
		return i > 0 && items[i-1].kind == box
	}
	return false
}

// lineAt sets the line from item start up to the break at item end.
func lineAt(items []item, s sums, start, end, width, lineEndStretch int) lineFit {
	natural := s.width[end] - s.width[start]
	if items[end].kind == penalty {
		natural += items[end].width
	}
	stretch := s.stretch[end] - s.stretch[start] + lineEndStretch
	return fit(natural, stretch, s.boxes[end]-s.boxes[start], width, s.infinite[end]-s.infinite[start] > 0)
}

// breaks returns the item indexes where lines end, chosen to minimize the
// total demerits, and that total.
//
// ponytail: dynamic programming over (break, fitness class) pairs. It is
// O(n·w) because a line stops growing once it is wider than width. Switch to
// the active-list form from the paper only if paragraphs get long.
func breaks(items []item, width, lineEndStretch int) ([]int, float64) {
	s := prefixSums(items)

	// node 0 is the paragraph start; node k > 0 is a legal break.
	pos := []int{-1}
	for i := range items {
		if legal(items, i) {
			pos = append(pos, i)
		}
	}

	type state struct {
		demerits  float64
		prevNode  int
		prevClass int
	}
	best := make([][4]state, len(pos))
	for k := range best {
		for c := range best[k] {
			best[k][c].demerits = math.Inf(1)
		}
	}
	best[0][1].demerits = 0 // the line before the first line counts as decent

	for a := range pos {
		for prevClass := 0; prevClass < 4; prevClass++ {
			cur := best[a][prevClass]
			if math.IsInf(cur.demerits, 1) {
				continue
			}
			start := pos[a] + 1
			prevFlagged := a > 0 && items[pos[a]].flagged
			for b := a + 1; b < len(pos); b++ {
				end := pos[b]
				if s.width[end]-s.width[start] > width && s.boxes[end]-s.boxes[start] > 1 {
					break
				}
				f := lineAt(items, s, start, end, width, lineEndStretch)
				if !f.feasible {
					continue
				}
				class := fitness(f.ratio)
				d := cur.demerits + demerits(f, items[end], class, prevClass, prevFlagged)
				if d < best[b][class].demerits {
					best[b][class] = state{d, a, prevClass}
				}
			}
		}
	}

	last := len(pos) - 1
	class := 0
	for c := 1; c < 4; c++ {
		if best[last][c].demerits < best[last][class].demerits {
			class = c
		}
	}
	total := best[last][class].demerits

	var result []int
	for k := last; k > 0; {
		result = append(result, pos[k])
		st := best[k][class]
		k, class = st.prevNode, st.prevClass
	}
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result, total
}

// segments breaks the text into lines, and splits each line into the runs of
// text between its spaces. A line that ends at a soft hyphen gets a "-".
func segments(text string, width int, set setting) [][]string {
	items := tokenize(text, set.interwordStretch)
	if width <= 0 || items == nil {
		return nil
	}
	ends, _ := breaks(items, width, set.lineEndStretch)
	lines := make([][]string, 0, len(ends))
	start := 0
	for _, end := range ends {
		var segs []string
		var sb strings.Builder
		for _, it := range items[start:end] {
			switch {
			case it.kind == box:
				sb.WriteString(it.text)
			case it.kind == glue && !it.infinite:
				segs = append(segs, sb.String())
				sb.Reset()
			}
		}
		if items[end].flagged {
			sb.WriteString("-")
		}
		lines = append(lines, append(segs, sb.String()))
		start = end + 1
	}
	return lines
}

/*
KnuthPlass wraps the text into lines that do not exceed the width. It uses the
Knuth-Plass algorithm, which chooses the breaks with the lowest total demerits
for the whole paragraph. Words can also break at soft hyphens (U+00AD). The
lines are ragged-right; use Justify for justified lines.
*/
func KnuthPlass(text string, width int) []string {
	lines := []string{}
	for _, segs := range segments(text, width, ragged) {
		lines = append(lines, strings.Join(segs, " "))
	}
	return lines
}
