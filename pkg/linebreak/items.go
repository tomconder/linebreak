package linebreak

import (
	"strings"
	"unicode/utf8"
)

const softHyphen = string(rune(0xAD))

// Knuth-Plass parameters. linePenalty, hyphenPenalty and fitnessDemerits are
// TeX's \linepenalty, \hyphenpenalty and \adjdemerits. flaggedDemerits is
// lower than TeX's \doublehyphendemerits (10000).
//
// ponytail: no \finalhyphendemerits (TeX: 5000). Add it if a hyphen at the
// end of the second-to-last line must cost more.
const (
	linePenalty     = 10
	hyphenPenalty   = 50
	flaggedDemerits = 3000
	fitnessDemerits = 10000
	infBad          = 10000
)

// setting is how much a line can stretch, in runes. Ragged text works like
// TeX's \raggedright: spaces do not stretch, and each line ends in glue that
// stretches. Justified text stretches the spaces instead.
type setting struct {
	interwordStretch int
	lineEndStretch   int
}

var (
	ragged    = setting{interwordStretch: 0, lineEndStretch: 4}
	justified = setting{interwordStretch: 2, lineEndStretch: 0}
)

type itemKind int

const (
	box itemKind = iota
	glue
	penalty
)

// item is a box, glue or penalty. In monospace text a space cannot shrink,
// so items have no shrink field.
type item struct {
	kind     itemKind
	text     string // box only
	width    int
	stretch  int  // glue only
	infinite bool // glue only: the paragraph-end glue that fills the last line
	penalty  int  // penalty only
	flagged  bool // penalty only: a hyphen break
}

// tokenize turns text into items. Words are boxes, spaces are glue, and each
// soft hyphen is a flagged penalty with the width of the "-" it adds when the
// line breaks there. A soft hyphen after a "-" adds nothing, so its width is 0.
func tokenize(text string, stretch int) []item {
	var items []item
	for _, word := range strings.Fields(text) {
		var parts []string
		for _, p := range strings.Split(word, softHyphen) {
			if p != "" {
				parts = append(parts, p)
			}
		}
		if len(parts) == 0 {
			continue
		}
		if len(items) > 0 {
			items = append(items, item{kind: glue, width: 1, stretch: stretch})
		}
		for j, p := range parts {
			if j > 0 {
				w := 1
				if strings.HasSuffix(parts[j-1], "-") {
					w = 0
				}
				items = append(items, item{kind: penalty, width: w, penalty: hyphenPenalty, flagged: true})
			}
			items = append(items, item{kind: box, text: p, width: utf8.RuneCountInString(p)})
		}
	}
	if len(items) == 0 {
		return nil
	}
	return append(items,
		item{kind: glue, infinite: true},
		item{kind: penalty})
}
