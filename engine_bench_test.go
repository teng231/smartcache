package smartcache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

func BenchmarkEngineCacheWrite(b *testing.B) {
	e := Start()
	var benchmarks = []struct {
		input int
	}{
		{input: 100},
		{input: 1000},
		{input: 74382},
	}

	for _, bm := range benchmarks {
		err := e.AddCollection(
			&CollectionConfig{
				Key:            fmt.Sprintf("casetest_%d", bm.input),
				Capacity:       100000,
				ExpireDuration: 10 * time.Second,
			},
		)
		if err != nil {
			log.Print(err)
		}
		ctx := context.TODO()
		b.Run(fmt.Sprintf("input_size_%d", bm.input), func(b *testing.B) {
			for i := 0; i < bm.input; i++ {
				if err := e.Select(ctx, fmt.Sprintf("casetest_%d", bm.input)).Upsert(i, fmt.Sprintf("value %d", i)); err != nil {
					log.Print(err)
				}
			}
		})
	}

}

func BenchmarkEngineCachex100000W10g(b *testing.B) {
	e := Start()
	e.AddCollection(
		&CollectionConfig{
			Key:            "bench100000w10g",
			Capacity:       100000,
			ExpireDuration: 10 * time.Second,
		},
	)
	buf := make(chan int, 100000)
	wg := &sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		buf <- i
		wg.Add(1)
	}
	ctx := context.TODO()

	for i := 0; i < 10; i++ {
		go func() {
			for {
				ivalue := <-buf
				if err := e.Select(ctx, "bench100000w10g").Upsert(ivalue, fmt.Sprintf("value %d", ivalue)); err != nil {
					log.Print(err)
				}
				if ivalue%3 == 0 {
					var k string
					hit, _ := e.Select(ctx, "bench100000w10g").Get(ivalue, nil).Exec(&k)
					if !hit {
						b.Fail()
					}
					if k != fmt.Sprintf("value %d", ivalue) {
						b.Fail()
					}
				}
				wg.Done()
			}
		}()
	}
	wg.Wait()
}
