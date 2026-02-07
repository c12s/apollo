package tracing

import (
	"fmt"
	"os"
)

type Config interface {
	Endpoint() string
}

type config struct {
	jaegerHost     string
	jaegerGRPCPort string
}

func NewConfig() Config {
	return config{
		jaegerHost:     os.Getenv("JAEGER_HOST"),
		jaegerGRPCPort: os.Getenv("JAEGER_GRPC_PORT"),
	}
}

func (c config) Endpoint() string {
	return fmt.Sprintf("%s:%s", c.jaegerHost, c.jaegerGRPCPort)
}
