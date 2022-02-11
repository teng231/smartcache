package smartcache

import (
	"context"
	"log"
	"reflect"
	"testing"
	"time"
)

type D struct {
	a string
}

func TestSessionRun(t *testing.T) {
	// create session
	testcaseOfCollection := map[string]interface{}{
		// "key1": "hello",
		// "key2": map[string]int{"x1": 1, "x2": 2},
		"key3": []int{1, 2, 3, 4},
		"key4": []*D{{"a"}, {"b"}},
	}
	testcaseCollections := map[string]*CollectionConfig{
		"col2": {Key: "col2", Capacity: 100, ExpireDuration: 2 * time.Second, GCInterval: 5 * time.Second},
	}
	collections := map[string]*Collection{}

	for k, config := range testcaseCollections {
		col, err := CreateCollection(config)
		if err != nil {
			log.Print(err)
			t.Fail()
		}
		collections[k] = col
	}
	col2 := collections["col2"]
	for k, val := range testcaseOfCollection {
		if err := col2.Upsert(context.TODO(), k, val); err != nil {
			log.Print(err)
			t.Fail()
		}
	}
	log.Print("abc", col2.Len())
	s := createSession(&SessionConfig{
		collection: col2,
	})

	var out1 int
	hit, _ := s.Get("key3", func(val interface{}, index int) bool {
		return val.(int) == 3
	}).Exec(&out1)
	log.Print(out1, hit)
	if !hit {
		t.Fail()
	}

	var out2 int
	hit, _ = s.Get("key3", func(val interface{}, index int) bool {
		return val.(int) == 10
	}).Exec(&out2)
	log.Print(out2, hit)
	if hit {
		t.Fail()
	}

	var out3 = &D{}
	hit, _ = s.Get("key4", func(val interface{}, index int) bool {
		return val.(*D).a == "a"
	}).Exec(out3)
	log.Print(out3, hit)
	if !hit {
		t.Fail()
	}

	var out4 = &D{}
	hit, _ = s.Get("key4", func(val interface{}, index int) bool {
		return val.(*D).a == "g"
	}).Exec(out4)
	log.Print(out4, hit)
	if hit {
		t.Fail()
	}

}

func x(ptr interface{}, key string) bool {
	log.Print(ptr)
	cases := map[string]interface{}{
		"c1": 1,
		"c2": D{"a"},
		"c3": &D{"b"},
	}
	val, ok := cases[key]
	if !ok {
		return false
	}
	structValue := reflect.Indirect(reflect.ValueOf(ptr))
	structValue.Set(reflect.Indirect(reflect.ValueOf(val)))
	return true
}

func TestX(t *testing.T) {
	for _, item := range []string{"c1", "c2", "c3"} {
		var out interface{}
		log.Print("out", &out)
		hit := x(&out, item)
		log.Print(hit, out, &out)
	}
	// var out interface{}
	// log.Print("out", &out)
	// hit := x(&out, "c1")
	// log.Print(hit, out, &out)
}
