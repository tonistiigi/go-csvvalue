package csvvalue

import (
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"testing"
)

type fieldsFunc func(string, []string) ([]string, error)

type fieldsTestFunc func(tc tcase) fieldsFunc

func stdlibFields(s string, _ []string) ([]string, error) {
	return csv.NewReader(strings.NewReader(s)).Read()
}

func stdlibTest(tc tcase) fieldsFunc {
	rdr := csv.NewReader(strings.NewReader(tc.Input))
	if tc.Comma != 0 {
		rdr.Comma = tc.Comma
	}
	rdr.LazyQuotes = tc.LazyQuotes
	rdr.TrimLeadingSpace = tc.TrimLeadingSpace
	return func(string, []string) ([]string, error) {
		return rdr.Read()
	}
}

func csvValueTest(tc tcase) fieldsFunc {
	return func(s string, _ []string) ([]string, error) {
		rdr := NewParser()
		if tc.Comma != 0 {
			rdr.Comma = tc.Comma
		}
		rdr.LazyQuotes = tc.LazyQuotes
		rdr.TrimLeadingSpace = tc.TrimLeadingSpace
		return rdr.Fields(s, nil)
	}
}

var testFuncs = map[string]fieldsTestFunc{
	"stdlib":   stdlibTest,
	"csvvalue": csvValueTest,
}

type tcase struct {
	Name   string
	Input  string
	Output []string
	Error  error

	Comma            rune
	LazyQuotes       bool
	TrimLeadingSpace bool
}

var testCases = []tcase{{
	Name:   "Simple",
	Input:  `foo,bar,baz`,
	Output: []string{"foo", "bar", "baz"},
}, {
	Name:   "Simple with NL",
	Input:  `foo,bar,baz` + "\n",
	Output: []string{"foo", "bar", "baz"},
}, {
	Name:   "Simple with CRNL",
	Input:  `foo,bar,baz` + "\r\n",
	Output: []string{"foo", "bar", "baz"},
}, {
	Name:   "Comma",
	Input:  "a;b;c\n",
	Output: []string{"a", "b", "c"},
	Comma:  ';',
}, {
	Name:   "Empty spaces",
	Input:  " ",
	Output: []string{" "},
}, {
	Name:  "Empty",
	Input: "",
	Error: io.EOF,
}, {
	Name:             "Trimmed spaces only",
	Input:            "   ",
	TrimLeadingSpace: true,
	Output:           []string{""},
}, {
	Name:             "Leading spaces",
	Input:            "   foo,bar,baz",
	TrimLeadingSpace: true,
	Output:           []string{"foo", "bar", "baz"},
}, {
	Name:   "Unhandled leading spaces",
	Input:  "   foo,bar,baz",
	Output: []string{"   foo", "bar", "baz"},
}, {
	Name:       "LazyQuotes",
	Input:      `a "word","1"2",a","b`,
	Output:     []string{`a "word"`, `1"2`, `a"`, `b`},
	LazyQuotes: true,
}, {
	Name:       "BareQuotes",
	Input:      `a "word","1"2",a"`,
	Output:     []string{`a "word"`, `1"2`, `a"`},
	LazyQuotes: true,
}, {
	Name:  "Newline only",
	Input: "\n",
	Error: io.EOF,
}, {
	Name:   "Quotes",
	Input:  `a,"b,cc ,ddd",e`,
	Output: []string{"a", "b,cc ,ddd", "e"},
}, {
	Name:   "Quotes only",
	Input:  `""`,
	Output: []string{""},
}, {
	Name:       "Single quote",
	Input:      `"`,
	Output:     []string{""},
	LazyQuotes: true,
}, {
	Name:  "Single quote strict",
	Input: `"`,
	Error: &csv.ParseError{StartLine: 1, Line: 1, Column: 2, Err: csv.ErrQuote},
}, {
	Name:  "Too many quotes only",
	Input: `"""`,
	Error: &csv.ParseError{StartLine: 1, Line: 1, Column: 4, Err: csv.ErrQuote},
}, {
	Name:  "One quote",
	Input: `foo,aa"bb,cc`,
	Error: &csv.ParseError{StartLine: 1, Line: 1, Column: 7, Err: csv.ErrBareQuote},
}, {
	Name:       "BareDoubleQuotes",
	Input:      `a""b,c`,
	Output:     []string{`a""b`, `c`},
	LazyQuotes: true,
}, {
	Name:  "BadDoubleQuotes",
	Input: `a""b,c`,
	Error: &csv.ParseError{Err: csv.ErrBareQuote, StartLine: 1, Line: 1, Column: 2},
}, {
	Name:             "TrimQuote",
	Input:            ` "a"," b",c`,
	Output:           []string{"a", " b", "c"},
	TrimLeadingSpace: true,
}, {
	Name:             "TrimAll",
	Input:            `  aa,   b,  c`,
	Output:           []string{"aa", "b", "c"},
	TrimLeadingSpace: true,
}, {
	Name:  "BadBareQuote",
	Input: `a "word","b"`,
	Error: &csv.ParseError{Err: csv.ErrBareQuote, StartLine: 1, Line: 1, Column: 3},
}, {
	Name:  "BadTrailingQuote",
	Input: `"a word",b"`,
	Error: &csv.ParseError{Err: csv.ErrBareQuote, StartLine: 1, Line: 1, Column: 11},
}, {
	Name:  "ExtraneousQuote",
	Input: `"a "word","b"`,
	Error: &csv.ParseError{Err: csv.ErrQuote, StartLine: 1, Line: 1, Column: 4},
}, {
	Name:   "Unicode comma",
	Input:  "a€b€c\n",
	Output: []string{"a", "b", "c"},
	Comma:  '€',
}, {
	Name:   "OCI config",
	Input:  `type=docker,name=test.docker,"containerimage.config={""Config"":{""Entrypoint"":[""/sbin/init"", ""--log-level=err""], ""StopSignal"":""37""},""os"":""linux"", ""architecture"":""amd64""}"`,
	Output: []string{"type=docker", "name=test.docker", `containerimage.config={"Config":{"Entrypoint":["/sbin/init", "--log-level=err"], "StopSignal":"37"},"os":"linux", "architecture":"amd64"}`},
}, {
	Name:   "End double quote",
	Input:  `a,"foo""",c`,
	Output: []string{"a", `foo"`, "c"},
},
}

func testFieldsCase(t *testing.T, tc tcase, f fieldsFunc) {
	t.Helper()

	got, err := f(tc.Input, nil)
	if err != nil {
		if tc.Error == nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if peExpected, ok := tc.Error.(*csv.ParseError); ok {
			var perr *csv.ParseError
			if !errors.As(err, &perr) {
				t.Fatalf("unexpected error: got %v, want %v", err, tc.Error)
			}
			if peExpected.StartLine != perr.StartLine {
				t.Fatalf("unexpected error start line: got %#v, want %#v", err, tc.Error)
			}
			if peExpected.Line != perr.Line {
				t.Fatalf("unexpected error line: got %v, want %v", err, tc.Error)
			}
			if peExpected.Column != perr.Column {
				t.Fatalf("unexpected error column: got %#v, want %#v", err, tc.Error)
			}
			if !errors.Is(peExpected.Err, perr.Err) {
				t.Fatalf("unexpected wrapped error: got %q, want %q", perr.Err.Error(), peExpected.Err.Error())
			}
		} else if !errors.Is(err, tc.Error) {
			t.Fatalf("unexpected error: got %v, want %v", err, tc.Error)
		}
		return
	}

	if tc.Error != nil {
		t.Fatalf("unexpected output, expected error: got %v, want %v", got, tc.Error)
	}

	if len(got) != len(tc.Output) {
		t.Fatalf("unexpected output: got %v, want %v", got, tc.Output)
	}
	for i := range got {
		if got[i] != tc.Output[i] {
			t.Fatalf("unexpected output: got %v, want %v", got, tc.Output)
		}
	}
}

func TestFields(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			for name, f := range testFuncs {
				t.Run(name, func(t *testing.T) {
					rdr := f(tc)
					testFieldsCase(t, tc, rdr)
				})
			}
		})
	}
}

func TestInvalidDelimeter(t *testing.T) {
	p := NewParser()
	p.Comma = '\n'
	_, err := p.Fields("foo\nbar\nbaz", nil)
	if !errors.Is(err, errInvalidDelim) {
		t.Fatalf("unexpected error: got %v, want %v", err, errInvalidDelim)
	}
}

func TestDefaultParser(t *testing.T) {
	out, err := Fields("foo,bar,baz", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("unexpected output: %v", out)
	}
	if out[0] != "foo" || out[1] != "bar" || out[2] != "baz" {
		t.Fatalf("unexpected output: %v", out)
	}

	out2, err := Fields("aaa,bbb", out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out2) != 2 {
		t.Fatalf("unexpected output: %v", out2)
	}
	if out2[0] != "aaa" || out2[1] != "bbb" {
		t.Fatalf("unexpected output: %v", out2)
	}

	// check that buffer is reused
	if &out[0] != &out2[0] {
		t.Fatalf("unexpected output: %v", out2)
	}
}
