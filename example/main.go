package main

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/teng231/smartcache"
)

type D struct {
	A string
	B int
}

func getDataFromRedis(cond interface{}) (interface{}, error) {
	condI := cond.(string)
	items := strings.Split(condI, ".")
	id := items[len(items)-1]
	idInt, _ := strconv.ParseInt(id, 10, 64)
	if idInt%2 == 0 {
		return D{B: int(idInt), A: "redis hit"}, nil
	}
	return nil, errors.New("not found")
}

func setDataRedis(key, val interface{}) error {
	log.Print(key, val)
	return nil
}
func getDataFromMysql(cond interface{}) (interface{}, error) {
	condI := cond.(string)
	items := strings.Split(condI, ".")
	id := items[len(items)-1]
	idInt, _ := strconv.ParseInt(id, 10, 64)
	log.Print(strconv.ParseInt(id, 16, 64))
	if idInt%3 == 0 {
		return D{B: int(idInt), A: "db hit"}, nil
	}
	return nil, errors.New("not found")
}

func main() {
	cache := smartcache.Start(
		&smartcache.CollectionConfig{
			Key:      "key1",
			Capacity: 100,
		},
	)

	if err := cache.Select(context.TODO(), "key1").Upsert("abc", 1); err != nil {
		log.Print(err)
	}
	var out1 int
	cache.Select(context.TODO(), "key1").Get("abc", nil).Exec(&out1)
	log.Print(out1)
	hit, err := cache.Select(context.TODO(), "key2").Get("cde", nil).Exec(&out1)
	log.Print(err)
	if hit {
		log.Print("nothing hit")
	}
	// simple reader
	out2 := &D{}
	cache.Select(context.TODO(), "key1").Get(2, nil, getDataFromRedis, getDataFromMysql).Exec(out2)
	log.Print(2, out2)
	cache.Select(context.TODO(), "key1").Get(3, nil, getDataFromRedis, getDataFromMysql).Exec(out2)
	log.Print(3, out2)
	cache.Select(context.TODO(), "key1").Get(6, nil, getDataFromRedis, getDataFromMysql).Exec(out2)
	log.Print(6, out2)

	cache.Select(context.TODO(), "key1").Get(6, nil, getDataFromRedis, getDataFromMysql).Exec(out2)
	log.Print(6, out2)

	// insert to set redis
	out3 := &D{}
	cache.Select(context.TODO(), "key1").Upsert(10, &D{A: "setter done", B: 10}, setDataRedis)
	cache.Select(context.TODO(), "key1").Get(10, nil).Exec(out3)
	if out3.B != 10 {
		panic("fail")
	}
}
