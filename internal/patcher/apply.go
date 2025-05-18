package patcher

import (
	"fmt"
	"sort"
	"strings"

	"github.com/agnivade/levenshtein"
)

type Diff struct {
	Line    int
	Search  string
	Replace string
}

type Edit struct {
	Start int
	End   int
	Text  string
}

// Search tries to locate `diff.Search` inside `source`.
// It begins at the requested line and expands alternately upward/downward
// until a slice whose similarity ≥ threshold is found.
//
// Similarity = 1 - (levenshtein distance / maxLen).
// On success it returns the byte‑offset edit [Start, End) to replace and true.
// If nothing satisfies the threshold it returns (zero Edit, false).
func Search(source string, diff Diff, threshold float64) (Edit, bool) {
	lines := trimSplit(source) // keep original EOLs
	if len(lines) == 0 {
		return Edit{}, false
	}

	// cumulative byte offsets: offsets[i] == start byte of line i
	offsets := make([]int, len(lines)+1)
	for i, l := range lines {
		offsets[i+1] = offsets[i] + len(l)
	}

	searchLines := trimSplit(diff.Search)
	nSearch := len(searchLines)
	if nSearch == 0 || nSearch > len(lines) {
		return Edit{}, false
	}

	// clamp user hint into valid range
	startIdx := diff.Line - 1
	startIdx = max(0, min(startIdx, len(lines)-1))

	for radius := 0; ; radius++ {
		tried := false

		// candidate above / at the hint
		left := startIdx - radius
		if left >= 0 && left+nSearch <= len(lines) {
			tried = true
			chunk := strings.Join(lines[left:left+nSearch], "")
			if similarity(chunk, diff.Search) >= threshold {
				return Edit{
					Start: offsets[left],
					End:   offsets[left+nSearch],
					Text:  diff.Replace,
				}, true
			}
		}

		// candidate below the hint   (skip radius 0 = duplicate check)
		right := startIdx + radius
		if radius > 0 && right+nSearch <= len(lines) {
			tried = true
			chunk := strings.Join(lines[right:right+nSearch], "")
			if similarity(chunk, diff.Search) >= threshold {
				return Edit{
					Start: offsets[right],
					End:   offsets[right+nSearch],
					Text:  diff.Replace,
				}, true
			}
		}

		if !tried { // both directions are out of range → give up
			break
		}
	}
	return Edit{}, false
}

// Apply performs all edits in one pass.
// Edits are applied back‑to‑front so earlier byte offsets remain valid.
func Apply(source string, edits []Edit) (string, error) {
	if len(edits) == 0 {
		return source, nil
	}

	// Apply from highest → lowest byte index.
	sort.Slice(edits, func(i, j int) bool { return edits[i].Start > edits[j].Start })

	data := []byte(source)
	lastEnd := len(data) // used to detect overlaps

	for _, e := range edits {
		// range sanity
		if e.Start < 0 || e.End < e.Start || e.End > len(data) {
			return "", fmt.Errorf("invalid edit range [%d,%d)", e.Start, e.End)
		}
		if e.End > lastEnd { // overlap with a previously applied edit
			return "", fmt.Errorf("overlapping edits at [%d,%d)", e.Start, e.End)
		}

		// splice: data = data[:Start] + e.Text + data[End:]
		data = append(data[:e.Start], append([]byte(e.Text), data[e.End:]...)...)
		lastEnd = e.Start // next edit must end before this
	}
	return string(data), nil
}

func similarity(a, b string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	dist := levenshtein.ComputeDistance(a, b)
	return 1 - float64(dist)/float64(max(len(a), len(b)))
}

func trimSplit(s string) []string {
	// strings.SplitAfter adds a trailing "" if s ends with '\n';
	// we drop it to avoid off‑by‑one issues.
	lines := strings.SplitAfter(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
