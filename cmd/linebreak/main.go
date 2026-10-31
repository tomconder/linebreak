package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tomconder/linebreak/pkg/linebreak"
)

func main() {
	width := flag.Int("width", 80, "maximum line width in runes")
	algo := flag.String("algo", "kp", "line-breaking algorithm: greedy, minrag or kp")
	flag.Parse()

	if *width <= 0 {
		fmt.Fprintln(os.Stderr, "linebreak: -width must be greater than 0")
		os.Exit(2)
	}

	var wrap func(string, int) []string
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

	text, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "linebreak:", err)
		os.Exit(1)
	}

	for _, line := range wrap(string(text), *width) {
		fmt.Println(line)
	}
}
