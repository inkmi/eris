package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseLine tests the parseLine function with various input scenarios.
func TestParseLine(t *testing.T) {
	tests := []struct {
		line  string
		key   string
		value string
	}{
		{"title: Example Title", "title", "Example Title"},
		{"page-break: yes", "page-break", "yes"},
		{"sample: no", "sample", "no"},
		{"body content without key", "body", ""},
	}

	for _, test := range tests {
		key, value := parseLine(test.line)

		if key != test.key || value != test.value {
			t.Errorf("parseLine(%q) = %q, %q; want %q, %q", test.line, key, value, test.key, test.value)
		}
	}
}

// TestParseContent tests the parseContent function with a comprehensive example.
func TestParseContent(t *testing.T) {
	input := `title: Example Title
page-break: yes
sample: yes
backmatter: no
frontmatter: yes
`

	scanner := bufio.NewScanner(strings.NewReader(input))
	result := parseContent(scanner)

	// Expected content based on the input
	expected := content{
		title:       "Example Title",
		pageBreak:   true,
		sample:      true,
		backmatter:  false,
		frontmatter: true,
	}

	assert.Equal(t, expected, result)

}
