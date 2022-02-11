package smartcache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"reflect"
)

type Session struct {
	ctx        context.Context
	collection *Collection
	out        interface{}
	err        error
}

type SessionConfig struct {
	collection *Collection
	ctx        context.Context
	err        error
}

type ISession interface {
	Filter(iter func(interface{}, int) bool, setters ...FuncSetter) ISession
	Get(iter func(interface{}, int) bool, setters ...FuncSetter) error
	Exec(outptr interface{}) error
	Upsert(value interface{}) error
	Upserts(in ...*CollectionKV) (int, error)
	Delete(key interface{}) error
	Close()
}

func createSession(cf *SessionConfig) *Session {
	if cf.ctx == nil {
		cf.ctx = context.TODO()
	}
	return &Session{
		collection: cf.collection,
		ctx:        cf.ctx,
		err:        cf.err,
	}
}

func (s *Session) Close() {
	s.out = nil
	s.err = nil
	s.ctx = nil
}

func (s *Session) Filter(key interface{}, iter func(interface{}, int) bool, setters ...FuncSetter) *Session {
	if s.err != nil {
		return s
	}
	if iter == nil {
		val, ok := s.collection.Get(s.ctx, key, setters...)
		if ok {
			s.out = val
		}
		return s
	}
	out := make([]interface{}, 0, 10)
	s.collection.Iter(s.ctx, key, func(data interface{}, index int) {
		if ok := iter(data, index); ok {
			out = append(out, data)
		}
	}, setters...)
	s.out = out
	return s
}

func (s *Session) Get(key interface{}, iter func(interface{}, int) bool, setters ...FuncSetter) *Session {
	if s.err != nil {
		return s
	}
	if iter == nil {
		val, ok := s.collection.Get(s.ctx, key, setters...)
		if ok {
			s.out = val
		}
		return s
	}
	isdone := false
	s.collection.Iter(s.ctx, key, func(data interface{}, index int) {
		if ok := iter(data, index); ok && !isdone {
			s.out = data
			isdone = true
		}
	}, setters...)
	return s
}

func (s *Session) Exec(outptr interface{}) (bool, error) {
	defer s.Close()
	if s.err != nil {
		return false, s.err
	}
	if reflect.TypeOf(outptr).Kind() != reflect.Ptr {
		log.Print("warning: outptr is a pointer")
		return false, errors.New("warning: outptr is a pointer")
	}
	if s.out == nil {
		log.Print("warning: out is nil")
		return false, errors.New("warning: out is nil")
	}
	if reflect.TypeOf(s.out).Kind() == reflect.Slice {
		out, err := json.Marshal(s.out)
		if err != nil {
			log.Print(err)
			return false, err
		}
		if err := json.Unmarshal(out, outptr); err != nil {
			log.Print(err)
			return false, err
		}
		return true, nil
	}
	structValue := reflect.Indirect(reflect.ValueOf(outptr))
	structValue.Set(reflect.Indirect(reflect.ValueOf(s.out)))
	return true, nil
}

func (s *Session) Upsert(key interface{}, value interface{}) error {
	return s.collection.Upsert(s.ctx, key, value)
}

func (s *Session) Upserts(kvs ...*CollectionKV) (int, error) {
	return s.collection.Upserts(s.ctx, kvs...)
}

func (s *Session) Delete(key interface{}) error {
	return s.collection.Delete(s.ctx, key)
}
