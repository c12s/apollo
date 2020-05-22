package main

import (
	"github.com/c12s/apollo/model"
	"github.com/c12s/apollo/service"
	"github.com/c12s/apollo/storage/etcd"
	"github.com/c12s/apollo/storage/redis"
	"github.com/c12s/apollo/storage/vault"
	"log"
	"time"
)

func main() {
	conf, err := model.ConfigFile()
	if err != nil {
		log.Fatal(err)
		return
	}

	cache, err := redis.New(conf.Cache)
	if err != nil {
		log.Fatal(err)
		return
	}

	secrets, err := vault.New(conf.SEndpoints)
	if err != nil {
		log.Fatal(err)
		return
	}

	db, err := etcd.New(conf, cache, secrets, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	service.Run(db, conf)
}
