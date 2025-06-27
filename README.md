# linebreak

A simple app to demonstrate line-breaking algorithms.

## Features

- **Dynamic Programming:** Efficiently computes optimal line breaks. It uses a penalty for trailing spaces to promote balanced text.
- **Unit Tested:** Includes unit tests to ensure quality and to allow future changes with confidence.

## Installation and Running

To run the app:

```bash
go run github.com/tomconder/linebreak/cmd/linebreak
```

To use the package in your project, you can get it using Go modules:

```bash
go get github.com/tomconder/linebreak/pkg/linebreak
```

## Algorithms

### Greedy

The [greedy algorithm](https://en.wikipedia.org/wiki/Greedy_algorithm) breaks a sequence of words into lines. At each step it fits as many words as possible on each line within the given width.

### Knuth-Plass

The [Knuth-Plass line breaking algorithm](https://en.wikipedia.org/wiki/Knuth%E2%80%93Plass_line-breaking_algorithm) breaks a sequence of words into lines that do not exceed the given width. It uses [dynamic programming](https://en.wikipedia.org/wiki/Dynamic_programming) to determine optimal breakpoints based on a penalty function that discourages trailing empty spaces.

