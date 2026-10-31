# linebreak

linebreak is a Go library and command-line tool that breaks text into lines of a given width. It has four functions, `Greedy`, `MinRaggedness`, `KnuthPlass` and `Justify`, and they use three algorithms:

* Greedy puts as many words on each line as fit.
* Minimum raggedness uses dynamic programming to choose the breaks that make the line lengths most even.
* Knuth-Plass is the algorithm that TeX uses. It chooses the breaks with the lowest total cost for the whole paragraph, and it can break words at soft hyphens and justify lines.

Width is counted in runes (Unicode code points), not bytes. The package has unit tests and fuzz tests.

## Install and run

The command reads text from standard input and writes the wrapped lines to standard output. For example:

```bash
echo "The lazy yellow dog was caught by the slow red fox" | go run github.com/tomconder/linebreak/cmd/linebreak -width 14 -algo kp
```

The command has two flags:

* `-width` sets the maximum line width in runes. The default is 80.
* `-algo` sets the algorithm. It can be `greedy`, `minrag` for minimum raggedness, `kp` for Knuth-Plass, or `justify` for justified Knuth-Plass. The default is `kp`.

The command treats all of standard input as one paragraph, so it doesn't keep blank lines.

To use the package in a Go project, add it with `go get`:

```bash
go get github.com/tomconder/linebreak/pkg/linebreak
```

## Algorithms

### Greedy

The [greedy algorithm](https://en.wikipedia.org/wiki/Greedy_algorithm) goes through the words in order, and it puts each word on the current line if the word fits. When a word doesn't fit, it starts a new line. For example, here is the following text at width 14:

```
The lazy yellow dog was caught by the slow red fox as he lay sleeping in the sun
```

The greedy algorithm gives the following lines:

```
The lazy
yellow dog was
caught by the
slow red fox
as he lay
sleeping in
the sun
```

### Minimum raggedness

The minimum raggedness algorithm uses [dynamic programming](https://en.wikipedia.org/wiki/Dynamic_programming) to choose the breaks that minimize the sum of the squared trailing spaces of every line except the last. No line is wider than the width, unless one word is wider than the width. For the same text at width 14, the line lengths are more even:

```
The lazy
yellow dog
was caught by
the slow red
fox as he lay
sleeping in
the sun
```

### Knuth-Plass

The [Knuth-Plass algorithm](https://en.wikipedia.org/wiki/Knuth%E2%80%93Plass_line-breaking_algorithm) models text as boxes, glue and penalties. Boxes are words, glue is the space between words, and penalties are other places where a line can break, such as soft hyphens.

For each possible line, the algorithm computes a badness score from how much the glue must stretch to fill the width. It then turns badness and penalties into demerits, which are the cost of each line, and it chooses the breaks with the lowest total demerits for the paragraph. It adds extra demerits in two cases:

* Two adjacent lines are in fitness classes that are more than one class apart. A fitness class groups lines by how loose or tight they are, so a very loose line next to a normal line costs more.
* Two lines in a row end in a hyphen.

Words can break at soft hyphens (U+00AD) in the input, and there is no hyphenation dictionary. Greedy and minimum raggedness treat U+00AD as an ordinary character, so they keep it in the words and count it toward the width.

`KnuthPlass` gives ragged-right output, like TeX's `\raggedright`. The spaces between words don't stretch, and each line instead ends in space that can stretch. For the example text at width 14, the output is the same as minimum raggedness. The two algorithms give different lines only on some inputs, e.g., `big big dog quickly red` at width 10:

```
Minimum raggedness   Knuth-Plass
------------------   -----------
big                  big big
big dog              dog
quickly              quickly
red                  red
```

### Justify

`Justify` runs the Knuth-Plass algorithm with spaces that can stretch, and then it pads every line except the last to the full width. It doesn't pad a line that has one word. For the example text at width 14, it gives the following lines:

```
The       lazy
yellow dog was
caught  by the
slow  red  fox
as    he   lay
sleeping    in
the sun
```

Monospace text can only add whole spaces, so narrow widths leave wide gaps. Justified output looks better at widths of 40 or more.
