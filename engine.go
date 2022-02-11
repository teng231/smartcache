package smartcache

import (
	"context"
	"errors"
	"log"
	"sync"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

/**
Engine is define a structure for config and save all data.
When engine run will create a session to execute with condition.
*/
type Engine struct {
	mCollection       map[string]*Collection
	mConfigCollection map[string]*CollectionConfig
	lock              *sync.RWMutex
}

type IEngine interface {
	Select(ctx context.Context, collectionKey string) ICollection
	AddCollection(cf ...*CollectionConfig) error
	Collection() map[string]*Collection
	CollectionConfig() map[string]*CollectionConfig
	Info()
}

func Start(cfs ...*CollectionConfig) *Engine {
	engine := &Engine{
		lock:              &sync.RWMutex{},
		mCollection:       make(map[string]*Collection),
		mConfigCollection: make(map[string]*CollectionConfig),
	}
	if err := engine.AddCollection(cfs...); err != nil {
		log.Panic(err)
	}
	return engine
}

func (e *Engine) Select(ctx context.Context, collectionKey string) *Session {
	col, has := e.mCollection[collectionKey]
	if !has {
		// log.Print(E_not_found_any_collection_key)
		return createSession(&SessionConfig{ctx: ctx, err: errors.New(E_not_found_any_collection_key)})
	}
	return createSession(&SessionConfig{collection: col, ctx: ctx})
}

func (e *Engine) AddCollection(cfs ...*CollectionConfig) error {
	if len(cfs) == 0 {
		return nil
	}
	for _, cf := range cfs {
		col, err := CreateCollection(cf)
		if err != nil {
			return err
		}
		e.lock.RLock()
		e.mCollection[col.key] = col
		e.mConfigCollection[col.key] = cf
		e.lock.RUnlock()
	}
	return nil
}

func (e *Engine) Collection() map[string]*Collection {
	return e.mCollection
}

func (e *Engine) CollectionConfig() map[string]*CollectionConfig {
	return e.mConfigCollection
}

func (e *Engine) Info() {
	log.Print(len(e.mCollection), e.mCollection)
	log.Print(len(e.mConfigCollection), e.mConfigCollection)
}
