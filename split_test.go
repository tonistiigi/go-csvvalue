package csvvalue

import (
	"errors"
	"io"
	"testing"
)

func TestSplitEarlyBreak(t *testing.T) {
	// Break after first field
	var first string
	for v, err := range Split("a,b,c") {
		if err != nil {
			t.Fatal(err)
		}
		first = v
		break
	}
	if first != "a" {
		t.Fatalf("unexpected first field: got %q, want %q", first, "a")
	}

	// Break after second field
	var fields []string
	for v, err := range Split("x,y,z") {
		if err != nil {
			t.Fatal(err)
		}
		fields = append(fields, v)
		if len(fields) == 2 {
			break
		}
	}
	if len(fields) != 2 || fields[0] != "x" || fields[1] != "y" {
		t.Fatalf("unexpected fields: got %v, want [x y]", fields)
	}
}

func TestSplitEarlyBreakQuoted(t *testing.T) {
	var first string
	for v, err := range Split(`"a,b",c,d`) {
		if err != nil {
			t.Fatal(err)
		}
		first = v
		break
	}
	if first != "a,b" {
		t.Fatalf("unexpected first field: got %q, want %q", first, "a,b")
	}
}

func TestSplitEarlyBreakUnicodeComma(t *testing.T) {
	p := NewParser()
	p.Comma = '€'
	var first string
	for v, err := range p.Split("a€b€c") {
		if err != nil {
			t.Fatal(err)
		}
		first = v
		break
	}
	if first != "a" {
		t.Fatalf("unexpected first field: got %q, want %q", first, "a")
	}
}

func TestSplitEarlyBreakQuotedUnicodeComma(t *testing.T) {
	p := NewParser()
	p.Comma = '€'
	var first string
	for v, err := range p.Split(`"a€b"€c€d`) {
		if err != nil {
			t.Fatal(err)
		}
		first = v
		break
	}
	if first != "a€b" {
		t.Fatalf("unexpected first field: got %q, want %q", first, "a€b")
	}
}

func TestSplitMultipleIterations(t *testing.T) {
	it := Split("a,b,c")

	for range 3 {
		var fields []string
		for v, err := range it {
			if err != nil {
				t.Fatal(err)
			}
			fields = append(fields, v)
		}
		if len(fields) != 3 || fields[0] != "a" || fields[1] != "b" || fields[2] != "c" {
			t.Fatalf("unexpected fields: got %v, want [a b c]", fields)
		}
	}
}

func TestSplitDefaultParser(t *testing.T) {
	out, err := Fields("foo,bar,baz", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 3 || out[0] != "foo" || out[1] != "bar" || out[2] != "baz" {
		t.Fatalf("unexpected output: %v", out)
	}

	// Test buffer reuse
	out2, err := Fields("aaa,bbb", out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out2) != 2 || out2[0] != "aaa" || out2[1] != "bbb" {
		t.Fatalf("unexpected output: %v", out2)
	}
	if &out[0] != &out2[0] {
		t.Fatalf("expected buffer reuse")
	}
}

func TestSplitEmpty(t *testing.T) {
	for _, err := range Split("") {
		if !errors.Is(err, io.EOF) {
			t.Fatalf("expected io.EOF, got %v", err)
		}
	}
}

func TestSplitInvalidDelim(t *testing.T) {
	p := NewParser()
	p.Comma = '\n'
	for _, err := range p.Split("foo") {
		if !errors.Is(err, errInvalidDelim) {
			t.Fatalf("expected errInvalidDelim, got %v", err)
		}
	}
}
