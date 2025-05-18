package patcher

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestSearch(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		diff      Diff
		threshold float64
		found     bool
		want      Edit
	}{
		{
			name:      "exact match",
			source:    "hello world\n",
			diff:      Diff{Line: 1, Search: "hello world\n", Replace: "goodbye world\n"},
			threshold: 1.0,
			found:     true,
			want:      Edit{Start: 0, End: 12, Text: "goodbye world\n"},
		},
		{
			name:      "no match",
			source:    "foo bar\n",
			diff:      Diff{Line: 1, Search: "hello world\n", Replace: "goodbye world\n"},
			threshold: 1.0,
			found:     false,
			want:      Edit{},
		},
		{
			name:      "partial match with lower threshold",
			source:    "hello wurld\n",
			diff:      Diff{Line: 1, Search: "hello world\n", Replace: "goodbye world\n"},
			threshold: 0.9, // allow one char difference
			found:     true,
			want:      Edit{Start: 0, End: 12, Text: "goodbye world\n"},
		},
		{
			name:      "multi-line match",
			source:    "foo\nbar\nbaz\n",
			diff:      Diff{Line: 2, Search: "bar\nbaz\n", Replace: "qux\nquux\n"},
			threshold: 1.0,
			found:     true,
			want:      Edit{Start: 4, End: 12, Text: "qux\nquux\n"},
		},
		{
			name:      "match at different line (hint off by one)",
			source:    "foo\nbar\nbaz\n",
			diff:      Diff{Line: 1, Search: "bar\nbaz\n", Replace: "qux\nquux\n"},
			threshold: 1.0,
			found:     true,
			want:      Edit{Start: 4, End: 12, Text: "qux\nquux\n"},
		},
		{
			name:      "out-of-bounds line hint",
			source:    "foo\nbar\nbaz\n",
			diff:      Diff{Line: 10, Search: "baz\n", Replace: "qux\n"},
			threshold: 1.0,
			found:     true,
			want:      Edit{Start: 8, End: 12, Text: "qux\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edit, ok := Search(tt.source, tt.diff, tt.threshold)
			assert.Equal(t, ok, tt.found)
			if tt.found {
				assert.DeepEqual(t, edit, tt.want)
			}
		})
	}
}

func TestApply(t *testing.T) {
	tests := []struct {
		name   string
		source string
		edits  []Edit
		want   string
		err    bool
	}{
		{
			name:   "no edits",
			source: "hello world\n",
			edits:  nil,
			want:   "hello world\n",
			err:    false,
		},
		{
			name:   "single edit",
			source: "hello world\n",
			edits:  []Edit{{Start: 0, End: 5, Text: "goodbye"}},
			want:   "goodbye world\n",
			err:    false,
		},
		{
			name:   "multiple non-overlapping edits",
			source: "foo bar baz\n",
			edits: []Edit{
				{Start: 0, End: 3, Text: "FOO"},
				{Start: 4, End: 7, Text: "BAR"},
			},
			want: "FOO BAR baz\n",
			err:  false,
		},
		{
			name:   "overlapping edits",
			source: "abcdef",
			edits: []Edit{
				{Start: 0, End: 3, Text: "ABC"},
				{Start: 2, End: 5, Text: "XYZ"}, // overlaps with previous
			},
			want: "",
			err:  true,
		},
		{
			name:   "invalid range",
			source: "abcdef",
			edits:  []Edit{{Start: 5, End: 2, Text: "oops"}},
			want:   "",
			err:    true,
		},
		{
			name:   "edit at end",
			source: "abcdef",
			edits:  []Edit{{Start: 6, End: 6, Text: "!"}},
			want:   "abcdef!",
			err:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Apply(tt.source, tt.edits)
			if tt.err {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
				assert.Equal(t, got, tt.want)
			}
		})
	}
}
