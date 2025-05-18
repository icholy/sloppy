package patcher

import (
	"slices"
	"testing"

	"gotest.tools/v3/assert"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		tokens []token
	}{
		{
			name:   "empty input",
			input:  "",
			tokens: []token{{Type: EOF}},
		},
		{
			name:  "single text line",
			input: "hello world",
			tokens: []token{
				{Type: TextType, Line: 0, Text: "hello world"},
				{Type: EOF},
			},
		},
		{
			name:  "single text line with newline",
			input: "hello world\n",
			tokens: []token{
				{Type: TextType, Line: 0, Text: "hello world\n"},
				{Type: EOF},
			},
		},
		{
			name:  "start search",
			input: "<<<<<<< SEARCH line:1\n",
			tokens: []token{
				{Type: StartSearchType, Line: 0, Text: "<<<<<<< SEARCH line:1\n"},
				{Type: EOF},
			},
		},
		{
			name:  "separator",
			input: "=======\n",
			tokens: []token{
				{Type: TextSeparatorType, Line: 0, Text: "=======\n"},
				{Type: EOF},
			},
		},
		{
			name:  "end replace",
			input: ">>>>>>> REPLACE\n",
			tokens: []token{
				{Type: EndReplaceType, Line: 0, Text: ">>>>>>> REPLACE\n"},
				{Type: EOF},
			},
		},
		{
			name:  "mixed lines with newlines",
			input: "<<<<<<< SEARCH line:1\nfoo\n=======\nbar\n>>>>>>> REPLACE\n",
			tokens: []token{
				{Type: StartSearchType, Line: 0, Text: "<<<<<<< SEARCH line:1\n"},
				{Type: TextType, Line: 1, Text: "foo\n"},
				{Type: TextSeparatorType, Line: 2, Text: "=======\n"},
				{Type: TextType, Line: 3, Text: "bar\n"},
				{Type: EndReplaceType, Line: 4, Text: ">>>>>>> REPLACE\n"},
				{Type: EOF},
			},
		},
		{
			name:  "mixed lines without trailing newline",
			input: "<<<<<<< SEARCH line:1\nfoo\n=======\nbar\n>>>>>>> REPLACE",
			tokens: []token{
				{Type: StartSearchType, Line: 0, Text: "<<<<<<< SEARCH line:1\n"},
				{Type: TextType, Line: 1, Text: "foo\n"},
				{Type: TextSeparatorType, Line: 2, Text: "=======\n"},
				{Type: TextType, Line: 3, Text: "bar\n"},
				{Type: EndReplaceType, Line: 4, Text: ">>>>>>> REPLACE"},
				{Type: EOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := slices.Collect(tokenize(tt.input))
			assert.DeepEqual(t, tokens, tt.tokens)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		diffs []Diff
		err   bool
	}{
		{
			name:  "empty input",
			input: "",
			diffs: nil,
			err:   false,
		},
		{
			name:  "single diff block",
			input: "<<<<<<< SEARCH line:1\nfoo\n=======\nbar\n>>>>>>> REPLACE\n",
			diffs: []Diff{
				{
					Line:    1,
					Search:  "foo\n",
					Replace: "bar\n",
				},
			},
			err: false,
		},
		{
			name:  "multiple diff blocks",
			input: "<<<<<<< SEARCH line:1\nfoo\n=======\nbar\n>>>>>>> REPLACE\n<<<<<<< SEARCH line:3\nbaz\n=======\nqux\n>>>>>>> REPLACE\n",
			diffs: []Diff{
				{
					Line:    1,
					Search:  "foo\n",
					Replace: "bar\n",
				}, {
					Line:    3,
					Search:  "baz\n",
					Replace: "qux\n",
				},
			},
			err: false,
		},
		{
			name:  "invalid missing separator",
			input: "<<<<< SEARCH line:1\nfoo\nbar\n>>>>>>> REPLACE\n",
			diffs: nil,
			err:   true,
		},
		{
			name:  "invalid missing end",
			input: "<<<<< SEARCH line:1\nfoo\n=======\nbar\n",
			diffs: nil,
			err:   true,
		},
		{
			name:  "multi-line search and replace",
			input: "<<<<<<< SEARCH line:34\nfoo\nbar\n=======\nbaz\nqux\n>>>>>>> REPLACE\n",
			diffs: []Diff{{
				Line:    34,
				Search:  "foo\nbar\n",
				Replace: "baz\nqux\n",
			}},
			err: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs, err := Parse(tt.input)
			if tt.err {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, diffs, tt.diffs)
		})
	}
}
