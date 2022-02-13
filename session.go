package smartcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
)

type GetterFn func(interface{}) (interface{}, error)

type SetterFn func(interface{}, interface{}) error

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
	Filter(iter func(interface{}, int) bool, getterFns ...GetterFn) ISession
	Get(iter func(interface{}, int) bool, getterFns ...GetterFn) error
	Exec(outptr interface{}) error
	Upsert(key, value interface{}, setterFns ...SetterFn) error
	Delete(key interface{}, setterFns ...SetterFn) error
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
func (s *Session) KeyBulder(sim interface{}) interface{} {
	return fmt.Sprintf("%v.%v", s.collection.Key(), sim)
}

func (s *Session) Filter(key interface{}, iter func(interface{}, int) bool, getterFns ...GetterFn) *Session {
	if s.err != nil {
		return s
	}
	if !s.collection.IsKeyExisted(key) {
		isok := false
		for _, f := range getterFns {
			val, err := f(s.KeyBulder(key))
			if err != nil {
				continue
			}
			if val != nil {
				if err := s.collection.Upsert(s.ctx, key, val); err != nil {
					continue
				}
			}
			isok = true
			break
		}
		if !isok {
			return s
		}
	}
	if iter == nil {
		val, ok := s.collection.Get(s.ctx, key)
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
	})
	s.out = out
	return s
}

func (s *Session) Get(key interface{}, iter func(interface{}, int) bool, getterFns ...GetterFn) *Session {
	if s.err != nil {
		return s
	}
	if !s.collection.IsKeyExisted(key) {
		isok := false
		for _, f := range getterFns {
			val, err := f(s.KeyBulder(key))
			if err != nil {
				continue
			}
			if val != nil {
				if err := s.collection.Upsert(s.ctx, key, val); err != nil {
					continue
				}
			}
			isok = true
			break
		}
		if !isok {
			return s
		}
	}
	if iter == nil {
		val, ok := s.collection.Get(s.ctx, key)
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
	})
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
			log.Print("filter need a slice to: ", err)
			return false, err
		}
		return true, nil
	}
	structValue := reflect.Indirect(reflect.ValueOf(outptr))
	structValue.Set(reflect.Indirect(reflect.ValueOf(s.out)))
	return true, nil
}

func (s *Session) Upsert(key interface{}, value interface{}, setterFns ...SetterFn) error {
	err := s.collection.Upsert(s.ctx, key, value)
	if len(setterFns) == 0 {
		return err
	}
	if err != nil {
		return err
	}
	errstr := ""
	for _, f := range setterFns {
		err := f(s.KeyBulder(key), value)
		if err != nil {
			errstr += err.Error()
		}
	}
	if errstr != "" {
		return errors.New(errstr)
	}
	return nil
}

func (s *Session) Delete(key interface{}, setterFns ...SetterFn) error {
	err := s.collection.Delete(s.ctx, key)
	if len(setterFns) == 0 {
		return err
	}
	if err != nil {
		return err
	}
	errstr := ""
	for _, f := range setterFns {
		err := f(s.KeyBulder(key), nil)
		if err != nil {
			errstr += err.Error()
		}
	}
	if errstr != "" {
		return errors.New(errstr)
	}
	return nil
}
