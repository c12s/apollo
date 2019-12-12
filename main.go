package main

import (
	"github.com/c12s/apollo/model"
	"github.com/c12s/apollo/service"
	"log"
)

func main() {
	conf, err := model.ConfigFile()
	if err != nil {
		log.Fatal(err)
		return
	}

	service.Run(conf)
}
