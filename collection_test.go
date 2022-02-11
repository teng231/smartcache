package smartcache

import (
	"context"
	"log"
	"reflect"
	"testing"
	"time"
)

func TestCollectionRun(t *testing.T) {

	testcaseCollections := map[string]*CollectionConfig{
		"col1": {Key: "col1", Capacity: 10, ExpireDuration: 2 * time.Second},
		"col2": {Key: "col2", Capacity: 100, ExpireDuration: 2 * time.Second, GCInterval: 5 * time.Second},
	}

	testcaseOfCollection := map[string]interface{}{
		"key1": "hello",
		"key2": map[string]int{"x1": 1, "x2": 2},
		"key3": []int{1, 2, 3, 4},
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

	col1 := collections["col1"]
	for k, val := range testcaseOfCollection {
		if err := col1.Upsert(context.TODO(), k, val); err != nil {
			log.Print(err)
			t.Fail()
		}
	}

	log.Print(col1.Len())

	for k, val := range testcaseOfCollection {
		_val, has := col1.Get(context.TODO(), k)
		if !has {
			log.Print("key not found")
			t.Fail()
		}

		if !reflect.DeepEqual(_val, val) {
			log.Print("_val : ", _val, " val: ", val)
			t.Fail()
		}
		log.Print("done ", k)
	}

}
func TestCollectionRunWithGCInterval(t *testing.T) {
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
	if err := col2.Upsert(context.TODO(), "k1", "done"); err != nil {
		log.Print(33, err)
	}

	if col2.Len() != 1 {
		log.Print("err, len")
		t.Fail()
	}
	log.Print("wait for GC run")
	time.Sleep(7 * time.Second)

	if col2.Len() != 0 {
		log.Print("err, len 2")
		t.Fail()
	}
}

func TestCollectionRunUpsertsAndRemove(t *testing.T) {
	testcaseOfCollection := map[string]interface{}{
		"key1": "hello",
		"key2": map[string]int{"x1": 1, "x2": 2},
		"key3": []int{1, 2, 3, 4},
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
	colKVs := make([]*CollectionKV, 0, 10)
	for k, val := range testcaseOfCollection {
		colKVs = append(colKVs, &CollectionKV{k, val})
	}

	if c, err := col2.Upserts(context.TODO(), colKVs...); err != nil || c != len(testcaseOfCollection) {
		log.Print(err)
		t.Fail()
	}
	if col2.Len() != 3 {
		t.Fail()
	}

	col2.Delete(context.TODO(), "key1")

	if col2.Len() != 2 {
		t.Fail()
	}
}

func TestCollectionIter(t *testing.T) {
	testcaseOfCollection := map[string]interface{}{
		"key1": "hello",
		"key2": map[string]int{"x1": 1, "x2": 2},
		"key3": []int{1, 2, 3, 4},
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
	log.Print(col2.Len())
	col2.Iter(context.TODO(), "key3", func(item interface{}, index int) {
		data := item.(int)
		// process
		log.Print(data)
	})
}

func TestCollectionDelete(t *testing.T) {

	testcaseCollections := map[string]*CollectionConfig{
		"col1": {Key: "col1", Capacity: 10, ExpireDuration: 2 * time.Second},
	}

	testcaseOfCollection := map[string]interface{}{
		"key1": "hello",
		"key2": map[string]int{"x1": 1, "x2": 2},
		"key3": []int{1, 2, 3, 4},
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

	col1 := collections["col1"]
	for k, val := range testcaseOfCollection {
		if err := col1.Upsert(context.TODO(), k, val); err != nil {
			log.Print(err)
			t.Fail()
		}
	}

	log.Print(col1.Len())

	col1.Delete(context.TODO(), "key1")

	_, has := col1.Get(context.TODO(), "key1")
	if has {
		log.Print("key1 can not delete")
		t.Fail()
	}
}
