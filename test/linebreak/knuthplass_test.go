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
