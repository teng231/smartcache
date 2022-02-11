benchmark-subscribe-10000:
	go test -cpuprofile=cpu.out -benchmem -memprofile=mem.out -bench=BenchmarkEngineCacheWrite -run=^a
profiler-cpu:
	go tool pprof smartcache.test cpu.out
profiler-mem:
	go tool pprof smartcache.test mem.out