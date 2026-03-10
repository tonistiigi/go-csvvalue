package csvvalue

import (
	"testing"

	"github.com/tonistiigi/go-csvvalue/legacy"
)

var cacheMatrix = map[string]func(*testing.B, fieldsFunc, string){
	"withcache": benchFieldsWithCache,
	"nocache":   benchFieldsNoCache,
}

var fieldsFuncs = map[string]fieldsFunc{
	"stdlib": stdlibFields,
	"legacy": legacy.Fields,
	"split":  Fields,
}

func benchFieldsWithCache(b *testing.B, f fieldsFunc, inp string) {
	var res []string
	for i := 0; i < b.N; i++ {
		var err error
		res, err = f(inp, res)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchFieldsNoCache(b *testing.B, f fieldsFunc, inp string) {
	for i := 0; i < b.N; i++ {
		res, err := f(inp, nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = res
	}
}

type rangeFunc func(string) error

var rangeFuncs = map[string]rangeFunc{
	"stdlib": func(inp string) error {
		res, err := stdlibFields(inp, nil)
		if err != nil {
			return err
		}
		for _, v := range res {
			_ = v
		}
		return nil
	},
	"legacy": func(inp string) error {
		res, err := legacy.Fields(inp, nil)
		if err != nil {
			return err
		}
		for _, v := range res {
			_ = v
		}
		return nil
	},
	"split": func(inp string) error {
		for v, err := range Split(inp) {
			if err != nil {
				return err
			}
			_ = v
		}
		return nil
	},
}

func BenchmarkFields(b *testing.B) {
	b.ReportAllocs()
	inp := "foo=bar,baz=bax,bay"

	for name, f := range fieldsFuncs {
		b.Run(name, func(b *testing.B) {
			for cache, m := range cacheMatrix {
				b.Run(cache, func(b *testing.B) {
					b.ReportAllocs()
					m(b, f, inp)
				})
			}
		})
	}
}

func BenchmarkRange(b *testing.B) {
	inp := "foo=bar,baz=bax,bay"

	for name, f := range rangeFuncs {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := f(inp); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
