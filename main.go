package main

import (
	"apollo/configs"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config, err := configs.NewConfig()
	if err != nil {
		log.Fatalln(err)
	}

	app, err := NewAppWithConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	err = app.Start(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	<-shutdown

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	app.GracefulStop(ctx)
}
