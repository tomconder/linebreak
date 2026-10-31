package linebreak

import "math"

// KnuthPlass is a temporary alias for MinRaggedness. The true Knuth-Plass
// algorithm replaces it.
func KnuthPlass(text string, width int) []string {
	return MinRaggedness(text, width)
}

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
