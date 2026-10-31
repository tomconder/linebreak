# linebreak

A Go library demonstrating line-breaking algorithms with optimal text formatting.

## Overview

This project implements three algorithms to format text within specified width constraints. Knuth-Plass can also justify lines:

* **Greedy Algorithm**: Fast, simple approach that fits as many words as possible on each line
* **Minimum Raggedness**: Dynamic programming that minimizes the squared trailing spaces of each line
* **Knuth-Plass Algorithm**: The algorithm from TeX. It minimizes the total demerits of the paragraph, can break words at soft hyphens, and can justify lines

## Features

* **Multiple Algorithms:** Choose between greedy, minimum raggedness and Knuth-Plass line breaking
* **Dynamic Programming:** Efficiently computes optimal line breaks. Minimum raggedness penalizes trailing spaces to promote balanced text. Knuth-Plass minimizes demerits.
* **Unicode Aware:** Width is counted in runes, not bytes.
* **Unit Tested:** Includes unit and fuzz tests to ensure quality and to allow future changes with confidence.

## Installation and Running

The app reads text from standard input and writes the wrapped lines to standard output:

```bash
echo "The lazy yellow dog was caught by the slow red fox" | go run github.com/tomconder/linebreak/cmd/linebreak -width 14 -algo kp
```

* `-width`: maximum line width in runes (default 80)
* `-algo`: `greedy`, `minrag` for minimum raggedness, `kp` for Knuth-Plass, or `justify` for justified Knuth-Plass (default `kp`)

The app treats all of standard input as one paragraph, so blank lines are not kept.

To use the package in your project, you can get it using Go modules:

```bash
go get github.com/tomconder/linebreak/pkg/linebreak
```

## Algorithms

### Greedy

The [greedy algorithm](https://en.wikipedia.org/wiki/Greedy_algorithm) breaks a sequence of words into lines. At each step it fits as many words as possible on each line within the given width.

For example, the following text with a given width of 14:

```
The lazy yellow dog was caught by the slow red fox as he lay sleeping in the sun
```

gives the following result

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

The minimum raggedness algorithm breaks a sequence of words into lines that do not exceed the given width, except for a word that is longer than the width. It uses [dynamic programming](https://en.wikipedia.org/wiki/Dynamic_programming) to choose the breaks that minimize the sum of the squared trailing spaces of every line except the last.

For example, the following text with a given width of 14:

```
The lazy yellow dog was caught by the slow red fox as he lay sleeping in the sun
```

gives the following result

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

The [Knuth-Plass line breaking algorithm](https://en.wikipedia.org/wiki/Knuth%E2%80%93Plass_line-breaking_algorithm) treats text as boxes (words), glue (spaces) and penalties (possible breaks). For each possible line it computes a badness from how much the line must stretch, and it chooses the breaks with the lowest total demerits for the whole paragraph. It adds demerits when adjacent lines differ by more than one fitness class (for example a very loose line next to a decent one), and when two lines in a row end in a hyphen.

Words can break at soft hyphens (U+00AD) in the input. There is no hyphenation dictionary. Greedy and minimum raggedness do not treat soft hyphens specially; they keep U+00AD in the words and count it toward the width.

Output is ragged-right: spaces do not stretch, and each line ends in space that can stretch, like TeX's `\raggedright`. For the example text at width 14 it gives the same lines as minimum raggedness. Ragged output differs from minimum raggedness only on some inputs, for example `big big dog quickly red` at width 10:

| Minimum raggedness | Knuth-Plass |
|---|---|
| `big` / `big dog` / `quickly` / `red` | `big big` / `dog` / `quickly` / `red` |

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
