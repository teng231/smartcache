package smartcache

import (
	"context"
	"errors"
	"log"
	"reflect"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

/**
Collection is define a structure of data
Save data inmem here
*/
type CollectionValue struct {
	Created int64       `json:"created"`
	Value   interface{} `json:"value"`
}

type CollectionKV struct {
	Key   interface{} `json:"key"`
	Value interface{} `json:"value"`
}

type ICollection interface {
	Upsert(ctx context.Context, key, value interface{}) error
	Upserts(ctx context.Context, in ...*CollectionKV) (int, error)
	Delete(ctx context.Context, key interface{}) error
	Get(ctx context.Context, key interface{}) (interface{}, bool)
	Iter(ctx context.Context, key interface{}, filtering func(item interface{}, index int))
	Key() string
	Len() int
	IsKeyExisted(key interface{}) bool
	GC()
}

type Collection struct {
	key            string
	data           *lru.Cache
	expireDuration time.Duration
}

type CollectionConfig struct {
	Key            string
	Capacity       int
	ExpireDuration time.Duration
	GCInterval     time.Duration
}

func CreateCollection(config *CollectionConfig) (*Collection, error) {
	if config.Key == "" {
		return nil, errors.New(E_not_found_any_collection_key)
	}
	if config.Capacity == 0 {
		config.Capacity = 100
	}
	c, err := lru.New(config.Capacity)
	if err != nil {
		log.Print(config.Capacity, err)
		return nil, err
	}
	s := &Collection{
		data:           c,
		expireDuration: config.ExpireDuration,
		key:            config.Key,
	}
	if config.GCInterval != 0 {
		tick := time.NewTicker(config.GCInterval)
		go func() {
			<-tick.C
			s.GC()
		}()
	}
	return s, nil
}

func (c *Collection) IsKeyExisted(key interface{}) bool {
	has := c.data.Contains(key)
	if !has {
		return has
	}
	value, _ := c.data.Get(key)
	colValue := value.(*CollectionValue)
	if colValue.Created+int64(c.expireDuration) < time.Now().Unix() {
		c.data.Remove(key)
		return false
	}
	return true
}

func (c *Collection) Key() string {
	return c.key
}

func (c *Collection) Len() int {
	return c.data.Len()
}

// GC remove key expired you need slow run it
func (c *Collection) GC() error {
	if c.data.Len() == 0 {
		return nil
	}
	tombs := make([]interface{}, 0, 10)
	for _, key := range c.data.Keys() {
		value, _ := c.data.Get(key)
		sval := value.(*CollectionValue)
		if time.Now().Unix() > sval.Created+int64(c.expireDuration.Seconds()) {
			tombs = append(tombs, key)
		}
	}

	for _, tomb := range tombs {
		c.data.Remove(tomb)
	}
	return nil
}

func (c *Collection) Upsert(ctx context.Context, key interface{}, value interface{}) error {
	cvalue := &CollectionValue{
		Created: time.Now().Unix(),
		Value:   value,
	}
	ef := c.data.Add(key, cvalue)
	if !ef {
		return nil
	}
	return errors.New(E_upsert_problem)
}

func (c *Collection) Upserts(ctx context.Context, in ...*CollectionKV) (int, error) {
	count := 0
	for _, item := range in {
		cvalue := &CollectionValue{
			Created: time.Now().Unix(),
			Value:   item.Value,
		}
		ef := c.data.Add(item.Key, cvalue)
		if !ef {
			count++
		}
	}
	return count, nil
}

func (c *Collection) Delete(ctx context.Context, key interface{}) error {
	ef := c.data.Remove(key)
	if ef {
		return nil
	}
	return errors.New(E_remove_problem)
}

func (c *Collection) Get(ctx context.Context, key interface{}) (interface{}, bool) {
	value, has := c.data.Get(key)
	if !has {
		return nil, has
	}
	colValue := value.(*CollectionValue)
	if colValue.Created+int64(c.expireDuration) < time.Now().Unix() {
		c.data.Remove(key)
		return nil, false
	}
	return colValue.Value, has
}

func (c *Collection) Iter(ctx context.Context, key interface{}, filtering func(item interface{}, index int)) {
	value, has := c.data.Get(key)
	if !has {
		return
	}
	colValue := value.(*CollectionValue)
	if colValue.Created+int64(c.expireDuration) < time.Now().Unix() {
		c.data.Remove(key)
		return
	}
	rv := reflect.ValueOf(colValue.Value)
	if rv.Kind() != reflect.Slice {
		return
	}

	for i := 0; i < rv.Len(); i++ {
		filtering(rv.Index(i).Interface(), i)
	}
}
