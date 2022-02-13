package smartcache

import (
	"context"
	"errors"
	"log"
	"reflect"
	"testing"
	"time"
)

type D struct {
	a string
}

func TestSessionRead(t *testing.T) {
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
	if !hit || out1 != 3 {
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
	if !hit || out3.a != "a" {
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

	var out5 int
	hit, _ = s.Get("key5", func(val interface{}, index int) bool {
		return val.(int) == 5
	}, func(i interface{}) (interface{}, error) {
		if i.(string) == "col2.key5" {
			return []int{1, 6, 5, 7}, nil
		}
		return nil, errors.New("nothing")
	}).Exec(&out5)
	log.Print(out5, hit)
	if !hit || out5 != 5 {
		t.Fail()
	}

	var out6 int
	hit, _ = s.Get("key6", func(val interface{}, index int) bool {
		return val.(int) == 4
	}, func(i interface{}) (interface{}, error) {
		if i.(string) == "col2.key5" {
			return []int{1, 6, 5, 7}, nil
		}
		return nil, errors.New("nothing")
	}).Exec(&out6)
	log.Print(out6, hit)
	if hit || out6 == 4 {
		t.Fail()
	}

}

func TestSessionRW(t *testing.T) {
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

	s := createSession(&SessionConfig{
		collection: col2,
	})

	var out5 int
	setToRedis := func(k, val interface{}) error {
		log.Print(k, val)
		return nil
	}
	err := s.Upsert("key5", 2, setToRedis)
	if err != nil {
		t.Fail()
	}
	hit, _ := s.Get("key5", nil).Exec(&out5)
	log.Print(hit, out5)
	if out5 != 2 {
		t.Fail()
	}

	var out6 []int
	err = s.Upsert("key6", []int{1, 2, 3, 4, 7}, setToRedis)
	if err != nil {
		t.Fail()
	}
	hit, _ = s.Filter("key6", func(val interface{}, index int) bool {
		return val.(int) > 3
	}).Exec(&out6)
	log.Print(hit, out6)
	if len(out6) != 2 {
		t.Fail()
	}

	var out7 int
	err = s.Upsert("key7", []int{1, 2, 3, 4, 7}, setToRedis)
	if err != nil {
		t.Fail()
	}
	hit, _ = s.Get("key7", func(val interface{}, index int) bool {
		return val.(int) > 3
	}).Exec(&out7)
	log.Print(hit, out7)
	if out7 != 4 {
		t.Fail()
	}

	var out8 int
	err = s.Upsert("key8", 9, setToRedis)
	if err != nil {
		t.Fail()
	}
	hit, _ = s.Filter("key8", func(val interface{}, index int) bool {
		return val.(int) > 3
	}).Exec(&out8)
	log.Print(hit, out8)
	if out8 == 0 && hit {
		t.Fail()
	}

	var out9 int
	err = s.Upsert("key9", 9)
	if err != nil {
		t.Fail()
	}
	hit, _ = s.Get("key9", func(val interface{}, index int) bool {
		return val.(int) >= 9
	}).Exec(&out9)
	log.Print(hit, out9)
	if out9 == 9 && hit {
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
