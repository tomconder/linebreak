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
