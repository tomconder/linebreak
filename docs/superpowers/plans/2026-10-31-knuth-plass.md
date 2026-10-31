# True Knuth-Plass Line Breaking Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development (recommended) or executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the misnamed `KnuthPlass` (really minimum raggedness) with a true Knuth-Plass line breaker, keep the old algorithm as `MinRaggedness`, and add `Justify`.

**Architecture:** Text becomes a list of Knuth-Plass items: boxes (word fragments), glue (spaces) and penalties (soft hyphens and the paragraph end). A dynamic program over (break, fitness class) pairs picks the breaks with the lowest total demerits, using the 1981 paper's badness and demerits formulas and TeX's default parameters. Two glue settings select the output: ragged-right (like TeX's `\raggedright`) for `KnuthPlass`, and stretchable spaces for `Justify`.

**Tech Stack:** Go 1.24.4 (module floor), standard library only, `github.com/stretchr/testify` for the existing table tests, Go native fuzzing.

**Spec:** No separate spec. The design decisions are in the "Design" section below. They were agreed in the planning session on 2026-10-31 and checked with a prototype (brute-force cross-check, and 90 s of fuzzing with no failures).

## Global Constraints

- The `go 1.24.4` line in `go.mod` stays. Use only APIs that exist in Go 1.24.
- No new dependencies.
- Width is counted in runes (`utf8.RuneCountInString`), the same as `Greedy` and `MinRaggedness`.
- Public signatures: `func KnuthPlass(text string, width int) []string`, `func MinRaggedness(text string, width int) []string`, `func Justify(text string, width int) []string`.
- `width <= 0` or text with no words returns `[]string{}` (empty, not nil). The existing tests check this.
- A word longer than the width gets a line of its own. It is never split, except at a soft hyphen.
- Never write a literal U+00AD in source; it is invisible. Write `string(rune(0xAD))`.
- Work on a feature branch. Each commit is GPG-signed (`git commit -S`). Never commit to `main`.
- Each task ends with `go vet ./...`, `golangci-lint run ./...` (v2.15, 0 issues) and `go test ./...` passing.
- Commands are bash (Git Bash on Windows).

## Design

| Decision | Choice |
|---|---|
| Old algorithm | Renamed `MinRaggedness`, behavior unchanged |
| `KnuthPlass` | True Knuth-Plass, ragged-right output |
| `Justify` | Same breaker with justified glue; pads each line except the last to `width` |
| Hyphenation | Soft hyphens (U+00AD) in the input only; no dictionary |
| Release | v1.1.0. `KnuthPlass` output changes, so the release notes must say so |

**Parameters** (TeX defaults): line penalty 10, hyphen penalty 50, flagged demerits 3000, fitness demerits 10000, badness cap 10000. Badness is `min(100·|r|³, 10000)`.

**Monospace glue.** A space is 1 rune wide and cannot shrink. Two settings, chosen by prototype:

| Setting | Interword stretch | Line-end stretch | Why |
|---|---|---|---|
| `ragged` | 0 | 4 | Like `\raggedright`. With stretch 1 per space, r is large on most lines, badness hits the cap, and the breaks degrade (for example "The" alone on line 1). Line-end stretch 3, 4 and 6 gave the same breaks on all test texts |
| `justified` | 2 | 0 | Stretch 1 left "The" alone on line 1. Stretch 2 and 3 gave the same output |

**Feasibility.** A line wider than `width` is infeasible, except when it holds one box. That line gets badness 10000 and ratio −1 (class tight), so an overlong word still gets a line of its own. A line with zero stretch and slack gets ratio +Inf and badness 10000. The last line ends in glue that stretches without limit, so its ratio is 0.

**Search.** Dynamic programming over (legal break, fitness class of the line ending there). The line before line 1 counts as class 1 (decent). Inner loop stops when the natural width without the break penalty is greater than `width` and the line has more than one box. This is O(n·w), and the paper's active-list form is not needed.

## File Structure

| File | Responsibility |
|---|---|
| `pkg/linebreak/minraggedness.go` | Renamed from `knuthplass.go`; the old algorithm as `MinRaggedness` |
| `pkg/linebreak/items.go` | New: parameters, `setting`, `item`, `tokenize` |
| `pkg/linebreak/knuthplass.go` | New: `fit`, `fitness`, `demerits`, `sums`, `legal`, `lineAt`, `breaks`, `segments`, `KnuthPlass` |
| `pkg/linebreak/justify.go` | New: `Justify` |
| `pkg/linebreak/knuthplass_internal_test.go` | New: white-box tests (formulas, brute-force cross-check). It must be in the package directory to reach unexported names; the other tests stay in `test/linebreak/` |
| `test/linebreak/minraggedness_test.go` | Renamed from `knuthplass_test.go` |
| `test/linebreak/knuthplass_test.go` | New: table test for the new `KnuthPlass` |
| `test/linebreak/justify_test.go` | New: table test for `Justify` |
| `test/linebreak/fuzz_test.go` | Modified: properties for all four functions |
| `cmd/linebreak/main.go` | Modified: `-algo greedy\|minrag\|kp\|justify` |
| `README.md` | Modified: algorithm sections and CLI flags |

---

### Task 1: Rename the current algorithm to MinRaggedness

**Files:**
- Rename: `pkg/linebreak/knuthplass.go` → `pkg/linebreak/minraggedness.go`
- Create: `pkg/linebreak/knuthplass.go` (temporary alias, replaced in Task 4)
- Rename: `test/linebreak/knuthplass_test.go` → `test/linebreak/minraggedness_test.go`
- Modify: `test/linebreak/fuzz_test.go`, `cmd/linebreak/main.go`, `README.md`

**Interfaces:**
- Produces: `func MinRaggedness(text string, width int) []string`, with the exact behavior of today's `KnuthPlass`.
- Produces (temporary): `func KnuthPlass(text string, width int) []string` that calls `MinRaggedness`. Task 4 deletes it.

- [ ] **Step 1: Create the branch**

```bash
git switch main && git pull && git switch -c feat/true-knuth-plass
```

- [ ] **Step 2: Rename the test, so that it fails**

```bash
git mv test/linebreak/knuthplass_test.go test/linebreak/minraggedness_test.go
sed -i 's/func TestKnuthPlass(/func TestMinRaggedness(/; s/linebreak\.KnuthPlass(/linebreak.MinRaggedness(/' test/linebreak/minraggedness_test.go
sed -i 's/the cost that KnuthPlass minimizes/the cost that MinRaggedness minimizes/; s/linebreak.KnuthPlass(text, width)/linebreak.MinRaggedness(text, width)/; s/"KnuthPlass", text, width, kp)/"MinRaggedness", text, width, kp)/; s/KnuthPlass cost %d is greater than Greedy cost %d: kp %q/MinRaggedness cost %d is greater than Greedy cost %d: minrag %q/' test/linebreak/fuzz_test.go
grep -n "KnuthPlass" test/linebreak/minraggedness_test.go test/linebreak/fuzz_test.go
```

Expected: the `grep` prints nothing.

- [ ] **Step 3: Run the tests to check that they fail**

Run: `go test ./test/...`
Expected: FAIL with `undefined: linebreak.MinRaggedness`.

- [ ] **Step 4: Rename the implementation**

```bash
git mv pkg/linebreak/knuthplass.go pkg/linebreak/minraggedness.go
```

In `pkg/linebreak/minraggedness.go`, replace the doc comment and the function name:

```go
/*
MinRaggedness wraps the text into lines that do not exceed the width. It
chooses the breaks that minimize the sum of the squared trailing spaces of
every line except the last (the minimum raggedness algorithm).
*/
func MinRaggedness(text string, width int) []string {
```

Create `pkg/linebreak/knuthplass.go`:

```go
package linebreak

// KnuthPlass is a temporary alias for MinRaggedness. The true Knuth-Plass
// algorithm replaces it.
func KnuthPlass(text string, width int) []string {
	return MinRaggedness(text, width)
}
```

- [ ] **Step 5: Add `minrag` to the CLI**

In `cmd/linebreak/main.go`, replace the flag line and the `switch`:

```go
	algo := flag.String("algo", "kp", "line-breaking algorithm: greedy, minrag or kp")
```

```go
	switch *algo {
	case "greedy":
		wrap = linebreak.Greedy
	case "minrag":
		wrap = linebreak.MinRaggedness
	case "kp":
		wrap = linebreak.KnuthPlass
	default:
		fmt.Fprintf(os.Stderr, "linebreak: unknown -algo %q, use greedy, minrag or kp\n", *algo)
		os.Exit(2)
	}
```

- [ ] **Step 6: Rename the README section**

In `README.md`, replace the `-algo` bullet:

```markdown
* `-algo`: `greedy`, `minrag` for minimum raggedness, or `kp` for Knuth-Plass (default `kp`)
```

Replace the heading `### Knuth-Plass` and the paragraph under it (lines 60-62) with:

```markdown
### Minimum raggedness

The minimum raggedness algorithm breaks a sequence of words into lines that do not exceed the given width, except for a word that is longer than the width. It uses [dynamic programming](https://en.wikipedia.org/wiki/Dynamic_programming) to choose the breaks that minimize the sum of the squared trailing spaces of every line except the last.
```

Keep its example text and result unchanged.

- [ ] **Step 7: Run all checks**

```bash
go vet ./... && golangci-lint run ./... && go test ./...
echo "The lazy yellow dog was caught by the slow red fox as he lay sleeping in the sun" | go run ./cmd/linebreak -width 14 -algo minrag
```

Expected: everything passes, 0 lint issues. The CLI prints the 7 lines of the README example: `The lazy` / `yellow dog` / `was caught by` / `the slow red` / `fox as he lay` / `sleeping in` / `the sun`.

- [ ] **Step 8: Commit**

```bash
git add -A pkg test cmd README.md
git commit -S -m "refactor(linebreak): rename KnuthPlass to MinRaggedness

The algorithm minimizes the sum of squared trailing spaces. That is
minimum raggedness, not Knuth-Plass. KnuthPlass stays as a temporary
alias until the true algorithm replaces it. The CLI gets -algo minrag."
```

---

### Task 2: Items, badness and demerits

**Files:**
- Create: `pkg/linebreak/items.go`
- Modify: `pkg/linebreak/knuthplass.go` (add functions below the alias)
- Test: `pkg/linebreak/knuthplass_internal_test.go`

**Interfaces:**
- Consumes: nothing from Task 1.
- Produces:
  - constants `linePenalty`, `hyphenPenalty`, `flaggedDemerits`, `fitnessDemerits`, `infBad` and `softHyphen`
  - `type itemKind int`, with the values `box`, `glue` and `penalty`
  - `type item struct{ kind itemKind; text string; width, stretch int; infinite bool; penalty int; flagged, forced bool }`
  - `func tokenize(text string, stretch int) []item`
  - `type lineFit struct{ feasible bool; ratio, badness float64 }`
  - `func fit(natural, stretch, boxes, width int, infinite bool) lineFit`
  - `func fitness(r float64) int`
  - `func demerits(f lineFit, p item, class, prevClass int, prevFlagged bool) float64`

- [ ] **Step 1: Write the failing tests**

Create `pkg/linebreak/knuthplass_internal_test.go`:

```go
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
		{kind: penalty, forced: true},
	}, tokenize(soft("ex|tra day"), 2))

	// soft hyphens with no text next to them are dropped
	assert.Equal(t, []item{
		{kind: box, text: "a", width: 1},
		{kind: glue, infinite: true},
		{kind: penalty, forced: true},
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
	end := item{kind: penalty, forced: true}

	// (10 + 337.5)², plus fitness demerits because class 3 follows class 1
	assert.Equal(t, 130756.25, demerits(lineFit{feasible: true, ratio: 1.5, badness: 337.5}, space, 3, 1, false))
	// (10 + 0 + 50)², plus flagged demerits for two hyphen lines in a row
	assert.Equal(t, 6600.0, demerits(lineFit{feasible: true}, hyphen, 1, 1, true))
	// the forced break adds no penalty: (10 + 0)²
	assert.Equal(t, 100.0, demerits(lineFit{feasible: true}, end, 1, 2, false))
	// a negative penalty is subtracted: (10 + 0)² - 20²
	assert.Equal(t, -300.0, demerits(lineFit{feasible: true}, item{kind: penalty, penalty: -20}, 1, 1, false))
}
```

- [ ] **Step 2: Run the tests to check that they fail**

Run: `go test ./pkg/linebreak/`
Expected: FAIL with `undefined: softHyphen`, `undefined: item`, and similar errors.

- [ ] **Step 3: Write `items.go`**

Create `pkg/linebreak/items.go`:

```go
package linebreak

import (
	"strings"
	"unicode/utf8"
)

const softHyphen = string(rune(0xAD))

// Knuth-Plass parameters, from Knuth and Plass (1981) and TeX's defaults.
const (
	linePenalty     = 10
	hyphenPenalty   = 50
	flaggedDemerits = 3000
	fitnessDemerits = 10000
	infBad          = 10000
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
	forced   bool // penalty only: the paragraph-end break
}

// tokenize turns text into items. Words are boxes, spaces are glue, and each
// soft hyphen is a flagged penalty with the width of the "-" it adds when the
// line breaks there.
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
				items = append(items, item{kind: penalty, width: 1, penalty: hyphenPenalty, flagged: true})
			}
			items = append(items, item{kind: box, text: p, width: utf8.RuneCountInString(p)})
		}
	}
	if len(items) == 0 {
		return nil
	}
	return append(items,
		item{kind: glue, infinite: true},
		item{kind: penalty, forced: true})
}
```

- [ ] **Step 4: Add `fit`, `fitness` and `demerits`**

In `pkg/linebreak/knuthplass.go`, add `import "math"` under the package line, and add this code below the alias:

```go
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
```

- [ ] **Step 5: Run all checks**

Run: `go vet ./... && golangci-lint run ./... && go test ./...`
Expected: PASS, 0 lint issues. The `unused` linter counts test files, so the new functions count as used. Do not add the `ragged` and `justified` settings yet: no test uses them before Task 3, so `unused` would report them.

- [ ] **Step 6: Commit**

```bash
git add pkg/linebreak/items.go pkg/linebreak/knuthplass.go pkg/linebreak/knuthplass_internal_test.go
git commit -S -m "feat(linebreak): add Knuth-Plass items, badness and demerits

Text becomes boxes, glue and penalties; a soft hyphen is a flagged
penalty. Badness and demerits follow Knuth and Plass (1981) with TeX's
default parameters. Glue cannot shrink in monospace text."
```

---

### Task 3: Optimal breaks

**Files:**
- Modify: `pkg/linebreak/items.go`, `pkg/linebreak/knuthplass.go`
- Test: `pkg/linebreak/knuthplass_internal_test.go`

**Interfaces:**
- Consumes: everything that Task 2 produces.
- Produces:
  - `type setting struct{ interwordStretch, lineEndStretch int }`, with the variables `ragged` (0, 4) and `justified` (2, 0)
  - `type sums struct{ width, stretch, boxes, infinite []int }`
  - `func prefixSums(items []item) sums`
  - `func legal(items []item, i int) bool`
  - `func lineAt(items []item, s sums, start, end, width, lineEndStretch int) lineFit`
  - `func breaks(items []item, width, lineEndStretch int) ([]int, float64)`, which returns the item indexes where lines end (in ascending order, the last one is the forced break) and the lowest total demerits.

- [ ] **Step 1: Write the failing test**

The brute force tries every set of breaks with the same cost functions. So this test checks the search, and Task 2 checks the formulas. Append to `pkg/linebreak/knuthplass_internal_test.go`:

```go
// bruteForce tries every set of breaks and returns the lowest total demerits.
func bruteForce(items []item, width, lineEndStretch int) float64 {
	s := prefixSums(items)
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
		total, start, prevClass, prevFlagged, ok := 0.0, 0, 1, false, true
		for _, end := range ends {
			f := lineAt(items, s, start, end, width, lineEndStretch)
			if !f.feasible {
				ok = false
				break
			}
			class := fitness(f.ratio)
			total += demerits(f, items[end], class, prevClass, prevFlagged)
			prevClass, prevFlagged, start = class, items[end].flagged, end+1
		}
		if ok && total < best {
			best = total
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
				_, got := breaks(items, width, set.lineEndStretch)
				assert.Equal(t, bruteForce(items, width, set.lineEndStretch), got, "%q %+v width %d", text, set, width)
			}
		}
	}
}

func TestBreaksEndAtForcedBreak(t *testing.T) {
	items := tokenize("The lazy yellow dog", ragged.interwordStretch)
	ends, _ := breaks(items, 10, ragged.lineEndStretch)
	assert.Equal(t, []int{3, 8}, ends) // the glue after "lazy", then the forced break
}
```

- [ ] **Step 2: Run the tests to check that they fail**

Run: `go test ./pkg/linebreak/`
Expected: FAIL with `undefined: prefixSums`, `undefined: legal`, `undefined: lineAt`, `undefined: breaks`, `undefined: setting`.

- [ ] **Step 3: Write the implementation**

In `pkg/linebreak/items.go`, add after the parameter constants:

```go
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
```

Append to `pkg/linebreak/knuthplass.go`:

```go
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
```

- [ ] **Step 4: Run all checks**

Run: `go vet ./... && golangci-lint run ./... && go test ./...`
Expected: PASS. `TestBreaksMatchesBruteForce` runs 300 cases (5 texts × 2 settings × 30 widths).

- [ ] **Step 5: Commit**

```bash
git add pkg/linebreak/items.go pkg/linebreak/knuthplass.go pkg/linebreak/knuthplass_internal_test.go
git commit -S -m "feat(linebreak): find the Knuth-Plass breaks with least demerits

Dynamic programming over (break, fitness class) pairs, so fitness and
flagged demerits between lines are exact. A test checks the result
against brute force for 300 inputs."
```

---

### Task 4: Replace KnuthPlass with the true algorithm

**Files:**
- Modify: `pkg/linebreak/knuthplass.go` (delete the alias; add `segments` and `KnuthPlass`)
- Create: `test/linebreak/knuthplass_test.go`
- Modify: `test/linebreak/fuzz_test.go` (full replacement below), `README.md`

**Interfaces:**
- Consumes: `tokenize`, `breaks`, `setting`, `ragged` and `item` from Tasks 2 and 3.
- Produces:
  - `func segments(text string, width int, set setting) [][]string`. It returns, for each line, the runs of text between its spaces. The last run of a line that ends at a soft hyphen ends in `-`. It returns nil when `width <= 0` or the text has no words.
  - `func KnuthPlass(text string, width int) []string`

- [ ] **Step 1: Write the failing test**

Create `test/linebreak/knuthplass_test.go`:

```go
package linebreak

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tomconder/linebreak/pkg/linebreak"
)

// soft replaces each "|" with a soft hyphen (U+00AD).
func soft(s string) string {
	return strings.ReplaceAll(s, "|", string(rune(0xAD)))
}

func TestKnuthPlass(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected []string
	}{
		{"Empty text", "", 10, []string{}},
		{"Zero width", "Hello", 0, []string{}},
		{"Single word", "Hello", 10, []string{"Hello"}},
		{"Two words exact fit", "Hello world", 11, []string{"Hello world"}},
		{"Multiple words with wrapping", "a b c d e", 3, []string{"a b", "c d", "e"}},
		{"Each word exactly fits", "hello world", 5, []string{"hello", "world"}},
		{
			"Multiple words with wrapping and spaces",
			"The lazy yellow dog was caught by the slow red fox as he lay sleeping in the sun", 14,
			[]string{"The lazy", "yellow dog", "was caught by", "the slow red", "fox as he lay", "sleeping in", "the sun"},
		},
		{"Word longer than width", "a supercalifragilisticexpialidocious word", 10, []string{"a", "supercalifragilisticexpialidocious", "word"}},
		{"Width counts runes not bytes", "café au lait", 7, []string{"café au", "lait"}},
		{"Breaks at a soft hyphen", soft("an ex|tra|or|di|nary day"), 10, []string{"an extra-", "ordinary", "day"}},
		{"Differs from minimum raggedness", "big big dog quickly red", 10, []string{"big big", "dog", "quickly", "red"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, linebreak.KnuthPlass(tt.text, tt.width))
		})
	}
}
```

The test table uses flat fields, not the `args` struct of the older tests, because the older tests carry dead fields.

- [ ] **Step 2: Run the test to check that it fails**

Run: `go test ./test/... -run TestKnuthPlass`
Expected: FAIL. The alias still returns the `MinRaggedness` output: `Breaks at a soft hyphen` keeps the soft hyphens in the words, and `Differs from minimum raggedness` returns `["big", "big dog", "quickly", "red"]`.

- [ ] **Step 3: Write the implementation**

In `pkg/linebreak/knuthplass.go`, delete the alias function and its comment. Change the import to:

```go
import (
	"math"
	"strings"
)
```

Append:

```go
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
```

- [ ] **Step 4: Run the test to check that it passes**

Run: `go test ./test/... -run TestKnuthPlass -v`
Expected: PASS, 11 subtests.

- [ ] **Step 5: Extend the fuzz test**

Replace `test/linebreak/fuzz_test.go` with:

```go
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
```

- [ ] **Step 6: Run the fuzzer**

Run: `go test ./test/linebreak -run '^$' -fuzz FuzzLineBreak -fuzztime 60s`
Expected: `PASS`. If it fails, `go test` saves the input in `test/linebreak/testdata/fuzz/FuzzLineBreak/`. Fix the code, and commit that input file with the fix as a regression case.

- [ ] **Step 7: Update the README**

In `README.md`, replace lines 7-10 (the overview list) with:

```markdown
This project implements three line-breaking algorithms to format text within specified width constraints:

* **Greedy Algorithm**: Fast, simple approach that fits as many words as possible on each line
* **Minimum Raggedness**: Dynamic programming that minimizes the squared trailing spaces of each line
* **Knuth-Plass Algorithm**: The algorithm from TeX. It minimizes the total demerits of the paragraph, can break words at soft hyphens, and can justify lines
```

Replace the `Multiple Algorithms` feature bullet with:

```markdown
* **Multiple Algorithms:** Choose between greedy, minimum raggedness and Knuth-Plass line breaking
```

Add this section after the minimum raggedness example:

````markdown
### Knuth-Plass

The [Knuth-Plass line breaking algorithm](https://en.wikipedia.org/wiki/Knuth%E2%80%93Plass_line-breaking_algorithm) treats text as boxes (words), glue (spaces) and penalties (possible breaks). For each possible line it computes a badness from how much the line must stretch, and it chooses the breaks with the lowest total demerits for the whole paragraph. It adds demerits when a very loose line follows a decent one, and when two lines in a row end in a hyphen.

Words can break at soft hyphens (U+00AD) in the input. There is no hyphenation dictionary.

Output is ragged-right: spaces do not stretch, and each line ends in space that can stretch, like TeX's `\raggedright`. For the example text at width 14 it gives the same lines as minimum raggedness. Ragged output differs from minimum raggedness only on some inputs, for example `big big dog quickly red` at width 10:

| Minimum raggedness | Knuth-Plass |
|---|---|
| `big` / `big dog` / `quickly` / `red` | `big big` / `dog` / `quickly` / `red` |
````

- [ ] **Step 8: Run all checks**

```bash
go vet ./... && golangci-lint run ./... && go test ./...
printf 'an ex\xc2\xadtra\xc2\xador\xc2\xaddi\xc2\xadnary day\n' | go run ./cmd/linebreak -width 10 -algo kp
```

Expected: everything passes. The CLI prints `an extra-` / `ordinary` / `day`.

- [ ] **Step 9: Commit**

```bash
git add pkg/linebreak/knuthplass.go test/linebreak/knuthplass_test.go test/linebreak/fuzz_test.go README.md
git commit -S -m "feat(linebreak): make KnuthPlass the true Knuth-Plass algorithm

KnuthPlass now minimizes Knuth-Plass demerits with ragged-right glue
and can break words at soft hyphens. Its output changes for some
input; MinRaggedness keeps the old behavior. The fuzz test checks
that KnuthPlass keeps the text and the width."
```

---

### Task 5: Justify

**Files:**
- Create: `pkg/linebreak/justify.go`
- Create: `test/linebreak/justify_test.go`
- Modify: `test/linebreak/fuzz_test.go`, `cmd/linebreak/main.go`, `README.md`

**Interfaces:**
- Consumes: `segments` and `justified` from Tasks 2 and 4, and `soft` from `test/linebreak/knuthplass_test.go`.
- Produces: `func Justify(text string, width int) []string`.

- [ ] **Step 1: Write the failing test**

Create `test/linebreak/justify_test.go`:

```go
package linebreak

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tomconder/linebreak/pkg/linebreak"
)

func TestJustify(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected []string
	}{
		{"Empty text", "", 10, []string{}},
		{"Zero width", "Hello", 0, []string{}},
		{"Last line is not padded", "Hello world", 20, []string{"Hello world"}},
		{
			"Pads every line except the last",
			"The lazy yellow dog was caught by the slow red fox as he lay sleeping in the sun", 14,
			[]string{"The       lazy", "yellow dog was", "caught  by the", "slow  red  fox", "as    he   lay", "sleeping    in", "the sun"},
		},
		{"Word longer than width", "a supercalifragilisticexpialidocious word", 10, []string{"a", "supercalifragilisticexpialidocious", "word"}},
		{"Width counts runes not bytes", "café au lait", 7, []string{"café au", "lait"}},
		{"Breaks at a soft hyphen", soft("an ex|tra|or|di|nary day"), 10, []string{"an", "extraordi-", "nary day"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, linebreak.Justify(tt.text, tt.width))
		})
	}
}
```

- [ ] **Step 2: Run the test to check that it fails**

Run: `go test ./test/... -run TestJustify`
Expected: FAIL with `undefined: linebreak.Justify`.

- [ ] **Step 3: Write the implementation**

Create `pkg/linebreak/justify.go`:

```go
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
```

- [ ] **Step 4: Run the test to check that it passes**

Run: `go test ./test/... -run TestJustify -v`
Expected: PASS, 7 subtests.

- [ ] **Step 5: Add Justify to the fuzz test**

In `test/linebreak/fuzz_test.go`, add this line after `kp := linebreak.KnuthPlass(text, width)`:

```go
		justified := linebreak.Justify(text, width)
```

Add these lines after `checkWidth(t, "KnuthPlass", width, kp)`:

```go
		checkText(t, "Justify", text, justified)
		checkWidth(t, "Justify", width, justified)
		for i, line := range justified[:max(len(justified)-1, 0)] {
			if n := utf8.RuneCountInString(line); len(strings.Fields(line)) > 1 && n != width {
				t.Fatalf("Justify: line %d %q is %d wide, want %d", i, line, n, width)
			}
		}
```

- [ ] **Step 6: Run the fuzzer**

Run: `go test ./test/linebreak -run '^$' -fuzz FuzzLineBreak -fuzztime 60s`
Expected: `PASS`. If it fails, commit the saved input with the fix, the same as in Task 4.

- [ ] **Step 7: Add `justify` to the CLI**

In `cmd/linebreak/main.go`:

```go
	algo := flag.String("algo", "kp", "line-breaking algorithm: greedy, minrag, kp or justify")
```

Replace the whole `switch` with:

```go
	switch *algo {
	case "greedy":
		wrap = linebreak.Greedy
	case "minrag":
		wrap = linebreak.MinRaggedness
	case "kp":
		wrap = linebreak.KnuthPlass
	case "justify":
		wrap = linebreak.Justify
	default:
		fmt.Fprintf(os.Stderr, "linebreak: unknown -algo %q, use greedy, minrag, kp or justify\n", *algo)
		os.Exit(2)
	}
```

- [ ] **Step 8: Update the README**

Replace the `-algo` bullet:

```markdown
* `-algo`: `greedy`, `minrag` for minimum raggedness, `kp` for Knuth-Plass, or `justify` for justified Knuth-Plass (default `kp`)
```

Add after the Knuth-Plass section:

````markdown
### Justify

`Justify` uses the Knuth-Plass algorithm with spaces that can stretch, then pads each line except the last to the full width. A line with one word is not padded. For the example text at width 14:

```
The       lazy
yellow dog was
caught  by the
slow  red  fox
as    he   lay
sleeping    in
the sun
```

Narrow widths leave wide gaps because monospace text can only add whole spaces. Justification looks better at widths of 40 or more.
````

- [ ] **Step 9: Run all checks**

```bash
go vet ./... && golangci-lint run ./... && go test ./... && pre-commit run --all-files
echo "The lazy yellow dog was caught by the slow red fox as he lay sleeping in the sun" | go run ./cmd/linebreak -width 14 -algo justify
echo "big big dog quickly red" | go run ./cmd/linebreak -width 10 -algo kp
echo x | go run ./cmd/linebreak -algo nope
```

Expected: everything passes. The first command prints the 7 lines of the README Justify example. The second prints `big big` / `dog` / `quickly` / `red`. The third prints `linebreak: unknown -algo "nope", use greedy, minrag, kp or justify` and exits with status 2.

- [ ] **Step 10: Commit**

```bash
git add pkg/linebreak/justify.go test/linebreak/justify_test.go test/linebreak/fuzz_test.go cmd/linebreak/main.go README.md
git commit -S -m "feat(linebreak): add Justify

Justify breaks lines with Knuth-Plass and stretchable spaces, then pads
each line except the last to the full width. The CLI gets
-algo justify. The fuzz test checks that justified lines are exactly
width runes wide."
```

---

## After the tasks

- Open one PR with the five commits. Merge it with a merge commit.
- Release as **v1.1.0**. The release notes must say that `KnuthPlass` output changes for some input, and that `MinRaggedness` gives the old output. They must also say that `KnuthPlass` now removes soft hyphens (U+00AD) from its output and can split words at them, where before they stayed in the words and counted toward the width.

## Known limits (not in scope)

- Rune count is not display width (East Asian wide characters, combining marks).
- No automatic hyphenation; only soft hyphens that are already in the input.
- One paragraph per call; blank lines in the input are not kept.
- Glue cannot shrink, so a line is never tighter than one space between words.
- Greedy and MinRaggedness keep soft hyphens in words and count them toward the width.
