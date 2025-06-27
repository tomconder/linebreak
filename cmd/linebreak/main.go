package main

import (
	"fmt"

	"github.com/tomconder/linebreak/pkg/linebreak"
)

func main() {
	text := "The lazy yellow dog was caught by the slow red fox as he lay sleeping in the sun"

	result := linebreak.Greedy(text, 14)
	for _, group := range result {
		fmt.Println(group)
	}

	fmt.Println("-------")

	result = linebreak.KnuthPlass(text, 14)
	for _, group := range result {
		fmt.Println(group)
	}
}
