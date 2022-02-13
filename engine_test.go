package smartcache

import (
	"context"
	"log"
	"reflect"
	"testing"
	"time"
)

func TestCreateEngine(t *testing.T) {
	e := Start()

	e.AddCollection(
		&CollectionConfig{
			Key:            "short_dur",
			Capacity:       100,
			ExpireDuration: 300 * time.Millisecond,
		},
		&CollectionConfig{
			Key:            "short_cap",
			Capacity:       5,
			ExpireDuration: 10 * time.Second,
		},
		&CollectionConfig{
			Key:            "normal",
			Capacity:       100,
			ExpireDuration: 10 * time.Second,
		},
	)
	e.Info()
	if len(e.Collection()) != 3 {
		t.Fail()
	}
}

func TestFeatureEngine(t *testing.T) {
	e := Start()

	e.AddCollection(
		&CollectionConfig{
			Key:            "short_dur",
			Capacity:       100,
			ExpireDuration: 300 * time.Millisecond,
		},
		&CollectionConfig{
			Key:            "short_cap",
			Capacity:       5,
			ExpireDuration: 10 * time.Second,
		},
		&CollectionConfig{
			Key:            "normal",
			Capacity:       100,
			ExpireDuration: 10 * time.Second,
		},
	)
	testcaseOfNormal := map[string]interface{}{
		"key1": "hello",
		"key2": map[string]int{"x1": 1, "x2": 2},
		"key3": []int{1, 2, 3, 4},
		"key4": []*D{{"a"}, {"b"}, {"c"}, {"d"}},
	}
	for key, value := range testcaseOfNormal {
		if err := e.Select(context.TODO(), "normal").Upsert(key, value); err != nil {
			log.Print(err)
			t.Fail()
		}
	}

	for key, value := range testcaseOfNormal {
		var data interface{}
		hit, _ := e.Select(context.TODO(), "normal").Get(key, nil).Exec(&data)
		if !hit {
			t.Fail()
			log.Print(data, value)
		}
		if !reflect.DeepEqual(data, value) {
			log.Print(data, value)
		}
	}
	var out int
	hit, _ := e.Select(context.TODO(), "normal").Get("key3", func(data interface{}, index int) bool {
		return data.(int) == 3
	}).Exec(&out)
	if !hit || out != 3 {
		log.Print("fail ", out)
		t.Fail()
	}

	sliceOut := make([]int, 0)
	hit, _ = e.Select(context.TODO(), "normal").Filter("key3", func(data interface{}, index int) bool {
		return data.(int) >= 2
	}).Exec(&sliceOut)
	log.Print(sliceOut)
	if !hit {
		log.Print("fail ", sliceOut)
		t.Fail()
	}
	sliceOut3 := make([]*D, 0)
	hit, _ = e.Select(context.TODO(), "normal").Filter("key4", func(data interface{}, index int) bool {
		return data.(*D).a == "a"
	}).Exec(&sliceOut3)
	log.Print(sliceOut3)
	if !hit {
		log.Print("fail ", sliceOut3)
		t.Fail()
	}

	sliceOut2 := make([]int, 0)
	hit, _ = e.Select(context.TODO(), "normal").Filter("key5", func(data interface{}, index int) bool {
		return data.(int) >= 2
	}).Exec(&sliceOut2)
	log.Print(hit, sliceOut2)
	if hit {
		log.Print("fail ", sliceOut2)
		t.Fail()
	}
}

func TestEngineWithHook(t *testing.T) {
	e := Start()

	e.AddCollection(
		&CollectionConfig{
			Key:            "normal",
			Capacity:       100,
			ExpireDuration: 10 * time.Second,
		},
	)
	testcaseOfNormal := map[string]interface{}{
		"key1": "hello",
		"key2": map[string]int{"x1": 1, "x2": 2},
		"key3": []int{1, 2, 3, 4},
		"key4": []*D{{"a"}, {"b"}, {"c"}, {"d"}},
	}
	for key, value := range testcaseOfNormal {
		if err := e.Select(context.TODO(), "normal").Upsert(key, value); err != nil {
			log.Print(err)
			t.Fail()
		}
	}

	var data interface{}
	hit, _ := e.Select(context.TODO(), "normal").Get("key5", nil, func(condition interface{}) (interface{}, error) {
		return 10, nil
	}).Exec(&data)
	log.Print(hit, data)

	e.Select(context.TODO(), "normal").Get("key5", nil).Exec(&data)
	log.Print(hit, data)
}

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	defer cancel()
	task(ctx, 10)
}

func task(ctx context.Context, num int) {
	done := make(chan bool)
	now := time.Now()
	go func() {
		time.Sleep(1 * time.Second)
		log.Print("done 3s")
		done <- true
	}()
	log.Print(time.Since(now))
	select {
	case <-done:
		log.Print("done task")
	case <-ctx.Done():
		log.Print("timeout")
	}
}
