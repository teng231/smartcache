# SmartCache
> Upgrade from my dumbcache

[![Go Reference](https://pkg.go.dev/badge/github.com/princjef/gomarkdoc.svg)](https://pkg.go.dev/github.com/teng231/smartcache)
[![Go](https://github.com/teng231/smartcache/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/teng231/smartcache/actions/workflows/go.yml)

Hope it can be tool to help your works.

### Installation

```bash
go get github.com/teng231/smartcache
```

### How it work?

flow like that:


```bash
┌────────┐       ┌────────┐      ┌───────────────┐
│        │       │        │      │               │
│ Engine ├─────►│ Session├─────►│   Collection  │
│        │       │        │      │               │
└────────┘       └────────┘      └───────────────┘
```

* When start app you create engine. Engine define structure of collection.
* Engine provider `select` to access module cache.
* Session process cache or set, get value to cache, from local module or remote module
* Collection local cache module

### Example
You can see full example here: [Example](https://github.com/teng231/smartcache/tree/master/example)

```go

cache := smartcache.Start(
    &smartcache.CollectionConfig{
        Key:      "key1",
        Capacity: 100,
    },
)

if err := cache.Select(context.TODO(), "key1").Upsert("abc", 1); err != nil {
    log.Print(err)
}
var out1 int
cache.Select(context.TODO(), "key1").Get("abc", nil).Exec(&out1)
log.Print(out1)

```

### Benmark Result:
follow link
[https://www.cloudbees.com/blog/real-life-go-benchmarking](https://www.cloudbees.com/blog/real-life-go-benchmarking)
```bash
$ go test -cpuprofile=cpu.out -benchmem -memprofile=mem.out -bench=BenchmarkEngineCacheWrite -run=^a

goos: linux
goarch: amd64
pkg: github.com/teng231/smartcache
cpu: Intel(R) Core(TM) i5-10400 CPU @ 2.90GHz
BenchmarkEngineCacheWrite/input_size_100-12         	1000000000	         0.0000460 ns/op	       0 B/op	       0 allocs/op
BenchmarkEngineCacheWrite/input_size_1000-12        	1000000000	         0.0004365 ns/op	       0 B/op	       0 allocs/op
BenchmarkEngineCacheWrite/input_size_74382-12       	1000000000	         0.05490 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/teng231/smartcache	0.513s
```

```bash
$ go test -cpuprofile=cpu.out -benchmem -memprofile=mem.out -bench=BenchmarkEngineCachex100000W10g -run=^a

goos: linux
goarch: amd64
pkg: github.com/teng231/smartcache
cpu: Intel(R) Core(TM) i5-10400 CPU @ 2.90GHz
BenchmarkEngineCachex100000W10g-12    	1000000000	         0.008111 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/teng231/smartcache	0.229s
```


Contributed:
    [Quang Anh](github.com/)