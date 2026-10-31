# linebreak

A Go library demonstrating line-breaking algorithms with optimal text formatting.

## Overview

This project implements two line-breaking algorithms to format text within specified width constraints:

* **Greedy Algorithm**: Fast, simple approach that fits as many words as possible on each line
* **Knuth-Plass Algorithm**: Sophisticated dynamic programming solution that optimizes line breaks based on penalty functions

## Features

* **Multiple Algorithms:** Choose between greedy and Knuth-Plass line breaking
* **Dynamic Programming:** Efficiently computes optimal line breaks. It uses a penalty for trailing spaces to promote balanced text.
* **Unicode Aware:** Width is counted in runes, not bytes.
* **Unit Tested:** Includes unit and fuzz tests to ensure quality and to allow future changes with confidence.

## Installation and Running

The app reads text from standard input and writes the wrapped lines to standard output:

```bash
echo "The lazy yellow dog was caught by the slow red fox" | go run github.com/tomconder/linebreak/cmd/linebreak -width 14 -algo kp
```

* `-width`: maximum line width in runes (default 80)
* `-algo`: `greedy` or `kp` for Knuth-Plass (default `kp`)

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

### Knuth-Plass

The [Knuth-Plass line breaking algorithm](https://en.wikipedia.org/wiki/Knuth%E2%80%93Plass_line-breaking_algorithm) breaks a sequence of words into lines that do not exceed the given width, except for a word that is longer than the width. It uses [dynamic programming](https://en.wikipedia.org/wiki/Dynamic_programming) to determine optimal breakpoints based on a penalty function that discourages trailing empty spaces.

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
