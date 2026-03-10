package csvvalue

import (
	"encoding/csv"
	"io"
	"iter"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Split returns an iterator over the fields in a CSV line using the default parser.
func Split(inp string) iter.Seq2[string, error] {
	return defaultParser.Split(inp)
}

// Split returns an iterator over the fields in a CSV line.
// The iterator yields (field, nil) for each field, or ("", error) on parse error.
// Iteration stops after the first error. If the input is empty, the iterator
// yields ("", io.EOF) immediately.
func (r *Parser) Split(line string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		if !validDelim(r.Comma) {
			yield("", errInvalidDelim)
			return
		}

		// allow trailing newline for compatibility
		if n := len(line); n > 0 && line[n-1] == '\n' {
			if n > 1 && line[n-2] == '\r' {
				line = line[:n-2]
			} else {
				line = line[:n-1]
			}
		}

		if len(line) == 0 {
			yield("", io.EOF)
			return
		}

		comma := r.Comma
		lazyQuotes := r.LazyQuotes
		trim := r.TrimLeadingSpace
		commaLen := utf8.RuneLen(comma)

		// Save original for pos computation on error only
		orig := line

		if comma < utf8.RuneSelf {
			commaByte := byte(comma) //nolint:gosec // guarded by comma < utf8.RuneSelf
			splitByte(line, orig, commaByte, commaLen, trim, lazyQuotes, yield)
		} else {
			splitRune(line, orig, comma, commaLen, trim, lazyQuotes, yield)
		}
	}
}

func splitByte(line, orig string, comma byte, commaLen int, trim, lazyQuotes bool, yield func(string, error) bool) {
	const quoteLen = len(`"`)

	for {
		if trim {
			i := strings.IndexFunc(line, func(r rune) bool {
				return !unicode.IsSpace(r)
			})
			if i < 0 {
				i = len(line)
			}
			line = line[i:]
		}
		if len(line) == 0 || line[0] != '"' {
			// Non-quoted string field
			i := strings.IndexByte(line, comma)
			var field string
			if i >= 0 {
				field = line[:i]
			} else {
				field = line
			}
			if !lazyQuotes {
				if j := strings.IndexByte(field, '"'); j >= 0 {
					yield("", parseErr(len(orig)-len(line)+j, csv.ErrBareQuote))
					return
				}
			}
			if !yield(field, nil) {
				return
			}
			if i >= 0 {
				line = line[i+commaLen:]
				continue
			}
			return
		}
		// Quoted string field
		pos := len(orig) - len(line)
		line = line[quoteLen:]
		pos += quoteLen
		var field string
		for {
			i := strings.IndexByte(line, '"')
			if i >= 0 {
				field += line[:i]
				line = line[i+quoteLen:]
				pos += i + quoteLen
				switch {
				case len(line) > 0 && line[0] == '"':
					// `""` sequence (append quote).
					field += "\""
					line = line[quoteLen:]
					pos += quoteLen
				case len(line) > 0 && line[0] == comma:
					// `",` sequence (end of field).
					line = line[commaLen:]
					if !yield(field, nil) {
						return
					}
					goto nextField
				case len(line) == 0:
					yield(field, nil)
					return
				case lazyQuotes:
					// `"` sequence (bare quote).
					field += "\""
				default:
					// `"*` sequence (invalid non-escaped quote).
					yield("", parseErr(pos-quoteLen, csv.ErrQuote))
					return
				}
			} else {
				if !lazyQuotes {
					yield("", parseErr(pos, csv.ErrQuote))
					return
				}
				// Hit end of line (copy all data so far).
				yield(field+line, nil)
				return
			}
		}
	nextField:
	}
}

func splitRune(line, orig string, comma rune, commaLen int, trim, lazyQuotes bool, yield func(string, error) bool) {
	const quoteLen = len(`"`)

	for {
		if trim {
			i := strings.IndexFunc(line, func(r rune) bool {
				return !unicode.IsSpace(r)
			})
			if i < 0 {
				i = len(line)
			}
			line = line[i:]
		}
		if len(line) == 0 || line[0] != '"' {
			// Non-quoted string field
			i := strings.IndexRune(line, comma)
			var field string
			if i >= 0 {
				field = line[:i]
			} else {
				field = line
			}
			if !lazyQuotes {
				if j := strings.IndexByte(field, '"'); j >= 0 {
					yield("", parseErr(len(orig)-len(line)+j, csv.ErrBareQuote))
					return
				}
			}
			if !yield(field, nil) {
				return
			}
			if i >= 0 {
				line = line[i+commaLen:]
				continue
			}
			return
		}
		// Quoted string field
		pos := len(orig) - len(line)
		line = line[quoteLen:]
		pos += quoteLen
		var field string
		for {
			i := strings.IndexByte(line, '"')
			if i >= 0 {
				field += line[:i]
				line = line[i+quoteLen:]
				pos += i + quoteLen
				switch rn := nextRune(line); {
				case rn == '"':
					// `""` sequence (append quote).
					field += "\""
					line = line[quoteLen:]
					pos += quoteLen
				case rn == comma:
					// `",` sequence (end of field).
					line = line[commaLen:]
					if !yield(field, nil) {
						return
					}
					goto nextField
				case len(line) == 0:
					yield(field, nil)
					return
				case lazyQuotes:
					// `"` sequence (bare quote).
					field += "\""
				default:
					// `"*` sequence (invalid non-escaped quote).
					yield("", parseErr(pos-quoteLen, csv.ErrQuote))
					return
				}
			} else {
				if !lazyQuotes {
					yield("", parseErr(pos, csv.ErrQuote))
					return
				}
				// Hit end of line (copy all data so far).
				yield(field+line, nil)
				return
			}
		}
	nextField:
	}
}
