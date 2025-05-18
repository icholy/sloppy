package patcher

import (
	"fmt"
	"iter"
	"strconv"
	"strings"
)

type tokenType int

const (
	StartSearchType   tokenType = iota // "<<<<<<< SEARCH line:n
	TextSeparatorType                  // "======="
	EndReplaceType                     // ">>>>>>> REPLACE
	TextType                           // any other line (incl. blank)
	InvalidType
	EOF
)

const (
	startSearchPrefix = "<<<<<<< SEARCH"
	textSeparator     = "======="
	endReplace        = ">>>>>>> REPLACE"
)

func tokenTypeString(typ tokenType) string {
	switch typ {
	case StartSearchType:
		return "StartSearchType: " + startSearchPrefix + " line:n"
	case TextSeparatorType:
		return "TextSeparatorType: " + textSeparator
	case EndReplaceType:
		return "EndReplaceType: " + endReplace
	case TextType:
		return "TextType"
	case InvalidType:
		return "InvalidType"
	case EOF:
		return "EOF"
	default:
		return "Unknown"
	}
}

type token struct {
	Type tokenType
	Line int
	Text string
}

func tokenize(input string) iter.Seq[token] {
	return func(yield func(token) bool) {
		lineNo := 0
		for line := range strings.Lines(input) {
			trim := strings.TrimRight(line, "\r\n")
			switch {
			case strings.HasPrefix(line, startSearchPrefix):
				if !yield(token{StartSearchType, lineNo, line}) {
					return
				}
			case trim == textSeparator:
				if !yield(token{TextSeparatorType, lineNo, line}) {
					return
				}
			case trim == endReplace:
				if !yield(token{EndReplaceType, lineNo, line}) {
					return
				}
			default:
				if !yield(token{TextType, lineNo, line}) {
					return
				}
			}
			lineNo++
		}
		yield(token{EOF, 0, ""})
	}
}

type parser struct {
	current token
	next    func() (token, bool)
}

func (p *parser) read() token {
	current := p.current
	tok, ok := p.next()
	if !ok {
		p.current = token{Type: InvalidType}
	} else {
		p.current = tok
	}
	return current
}

func (p *parser) expect(typ tokenType) (token, error) {
	tok := p.read()
	if tok.Type != typ {
		return token{}, fmt.Errorf("expected %s, got %s: line=%d: %q",
			tokenTypeString(typ),
			tokenTypeString(tok.Type),
			tok.Line,
			tok.Text,
		)
	}
	return tok, nil
}

func (p *parser) parseStartSearch() (int, error) {
	tok, err := p.expect(StartSearchType)
	if err != nil {
		return 0, err
	}
	suffix, _ := strings.CutPrefix(tok.Text, startSearchPrefix)
	lineStr, ok := strings.CutPrefix(strings.TrimSpace(suffix), "line:")
	if !ok {
		return 0, fmt.Errorf("expected %s, got %q", tokenTypeString(StartSearchType), tok.Text)
	}
	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return 0, fmt.Errorf("expected %s, got %q: %w", tokenTypeString(StartSearchType), tok.Text, err)
	}
	return line, nil
}

func (p *parser) parseDiff() (Diff, error) {
	var diff Diff
	var err error
	diff.Line, err = p.parseStartSearch()
	if err != nil {
		return Diff{}, err
	}
	for p.current.Type == TextType {
		diff.Search += p.current.Text
		p.read()
	}
	if _, err := p.expect(TextSeparatorType); err != nil {
		return Diff{}, err
	}
	for p.current.Type == TextType {
		diff.Replace += p.current.Text
		p.read()
	}
	if _, err := p.expect(EndReplaceType); err != nil {
		return Diff{}, nil
	}
	return diff, nil
}

func Parse(input string) ([]Diff, error) {
	next, stop := iter.Pull(tokenize(input))
	defer stop()
	p := parser{next: next}
	p.read()
	var diffs []Diff
	for p.current.Type != EOF {
		diff, err := p.parseDiff()
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, diff)
	}
	return diffs, nil
}
