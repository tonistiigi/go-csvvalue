### go-csvvalue

![GitHub Release](https://img.shields.io/github/v/release/tonistiigi/go-csvvalue)
[![Go Reference](https://pkg.go.dev/badge/github.com/tonistiigi/go-csvvalue.svg)](https://pkg.go.dev/github.com/tonistiigi/go-csvvalue)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/tonistiigi/go-csvvalue/ci.yml)
![Codecov](https://img.shields.io/codecov/c/github/tonistiigi/go-csvvalue)
![GitHub License](https://img.shields.io/github/license/tonistiigi/go-csvvalue)


`go-csvvalue` provides an efficient parser for a single-line CSV value.

It is more efficient than the standard library `encoding/csv` package for parsing many small values. The main problem with stdlib implementation is that it calls `bufio.NewReader` internally, allocating 4KB of memory on each invocation. For multi-line CSV parsing, the standard library is still recommended. If you wish to optimize memory usage for `encoding/csv`, call `csv.NewReader` with an instance of `*bufio.Reader` that already has a 4KB buffer allocated and then reuse that buffer for all reads.

For further memory optimization, an existing string slice can be optionally passed to be reused for returning the parsed fields.

For backwards compatibility with stdlib record parser, the input may contain a trailing newline character.

### Benchmark

```
goos: linux
goarch: arm64
pkg: github.com/tonistiigi/go-csvvalue
BenchmarkFields/stdlib/nocache-16         817705     1364 ns/op    4520 B/op    14 allocs/op
BenchmarkFields/stdlib/withcache-16       824451     1334 ns/op    4520 B/op    14 allocs/op
BenchmarkFields/legacy/withcache-16     64307284    19.02 ns/op       0 B/op     0 allocs/op
BenchmarkFields/legacy/nocache-16       28163379    43.50 ns/op      48 B/op     1 allocs/op
BenchmarkFields/split/nocache-16        28516675    43.53 ns/op      48 B/op     1 allocs/op
BenchmarkFields/split/withcache-16      59064800    21.02 ns/op       0 B/op     0 allocs/op
BenchmarkRange/stdlib-16                 1000000     1367 ns/op    4520 B/op    14 allocs/op
BenchmarkRange/legacy-16                24597015    44.06 ns/op      48 B/op     1 allocs/op
BenchmarkRange/split-16                 72433822    16.37 ns/op       0 B/op     0 allocs/op
PASS
```

### Credits

This package is mostly based on `encoding/csv` implementation and also uses that package for compatibility testing.
