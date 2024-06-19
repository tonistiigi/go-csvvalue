package csvvalue

import (
	"testing"
)

var cacheMatrix = map[string]func(*testing.B, fieldsFunc, string){
	"withcache": benchFieldsWithCache,
	"nocache":   benchFieldsNoCache,
}

var fieldsFuncs = map[string]fieldsFunc{
	"stdlib":   stdlibFields,
	"csvvalue": Fields,
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
