package linebreak

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tomconder/linebreak/pkg/linebreak"
)

func TestKnuthPlass(t *testing.T) {
	type args struct {
		text     string
		width    int
		expected []string
	}
	tests := []struct {
		name     string
		args     args
		want     []string
		expected []string
	}{
		{
			name: "Empty text",
			args: args{
				text:     "",
				width:    10,
				expected: []string{},
			},
		},
		{
			name: "Single word",
			args: args{
				text:     "Hello",
				width:    10,
				expected: []string{"Hello"},
			},
		},
		{
			name: "Two words exact fit",
			args: args{
				text:     "Hello world",
				width:    11,
				expected: []string{"Hello world"},
			},
		},
		{
			name: "Multiple words with wrapping",
			args: args{
				text:     "a b c d e",
				width:    3,
				expected: []string{"a b", "c d", "e"},
			},
		},
		{
			name: "Each word exactly fits",
			args: args{
				text:     "hello world",
				width:    5,
				expected: []string{"hello", "world"},
			},
		},
		{
			name: "Multiple words with wrapping and spaces",
			args: args{
				text:     "The lazy yellow dog was caught by the slow red fox as he lay sleeping in the sun",
				width:    14,
				expected: []string{"The lazy", "yellow dog", "was caught by", "the slow red", "fox as he lay", "sleeping in", "the sun"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := linebreak.KnuthPlass(tt.args.text, tt.args.width)
			assert.Equal(t, tt.args.expected, result)
		})
	}
}
