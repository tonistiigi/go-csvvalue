// Package csvvalue provides an efficient parser for a single line CSV value.
// It is more efficient than the standard library csv package for parsing many
// small values. For multi-line CSV parsing, the standard library is recommended.
package csvvalue

import (
	"encoding/csv"
	"errors"
	"strings"
	"unicode/utf8"
)

var errInvalidDelim = errors.New("csv: invalid field or comment delimiter")

var defaultParser = NewParser()

// Fields parses the line with default parser and returns
// slice of fields for the record. If dst is nil, a new slice is allocated.
func Fields(inp string, dst []string) ([]string, error) {
	return defaultParser.Fields(inp, dst)
}

// Parser is a CSV parser for a single line value.
type Parser struct {
	Comma            rune
	LazyQuotes       bool
	TrimLeadingSpace bool
}

// NewParser returns a new Parser with default settings.
func NewParser() *Parser {
	return &Parser{Comma: ','}
}

// Fields parses the line and returns slice of fields for the record.
// If dst is nil, a new slice is allocated.
func (r *Parser) Fields(line string, dst []string) ([]string, error) {
	if cap(dst) == 0 {
		// imprecise estimate, strings.Count is fast
		dst = make([]string, 0, 1+strings.Count(line, string(r.Comma)))
	} else {
		dst = dst[:0]
	}

	for field, err := range r.Split(line) {
		if err != nil {
			return nil, err
		}
		dst = append(dst, field)
	}
	return dst, nil
}

func validDelim(r rune) bool {
	return r != 0 && r != '"' && r != '\r' && r != '\n' && utf8.ValidRune(r) && r != utf8.RuneError
}

func nextRune(b string) rune {
	r, _ := utf8.DecodeRuneInString(b)
	return r
}

func parseErr(pos int, err error) error {
	return &csv.ParseError{StartLine: 1, Line: 1, Column: pos + 1, Err: err}
}
