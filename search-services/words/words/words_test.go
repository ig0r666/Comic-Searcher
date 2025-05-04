package words

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStopWord(t *testing.T) {
	tests := []struct {
		name string
		word string
		want bool
	}{
		{"stop word", "and", true},
		{"not a stop word", "linux", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsStopWord(tt.word))
		})
	}
}

func TestSplitIntoWords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"simple sentence", "hello world", []string{"hello", "world"}},
		{"with punctuation", "hello, world!", []string{"hello", "world"}},
		{"empty string", "", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, splitIntoWords(tt.input))
		})
	}
}

func TestNormalizedString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			"simple case",
			"cats dogs",
			[]string{"cat", "dog"},
		},
		{
			"with stop words",
			"the cat and the dog",
			[]string{"cat", "dog"},
		},
		{
			"with duplicates",
			"cats cats",
			[]string{"cat"},
		},
		{
			"empty string",
			"",
			[]string{},
		},
		{
			"stop words",
			"the and or",
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizedString(tt.input))
		})
	}
}

func TestNormalizedString_ErrorHandling(t *testing.T) {
	result := NormalizedString("test")
	assert.NotEmpty(t, result, "Should return not empty slice")
}
