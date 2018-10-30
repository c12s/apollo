package main

import (
	"context"
	"github.com/c12s/apollo/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	service.Run(ctx, "localhost:8083")
	cancel()
}
