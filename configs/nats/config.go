package nats

import (
	"fmt"
	"os"
)

type Config interface {
	Uri() string
}

type config struct {
	hostname string
	port     string
	username string
	password string
}

func NewConfig() Config {
	return config{
		hostname: os.Getenv("NATS_HOSTNAME"),
		port:     os.Getenv("NATS_PORT"),
		username: os.Getenv("NATS_USERNAME"),
		password: os.Getenv("NATS_PASSWORD"),
	}
}

func (c config) Uri() string {
	return fmt.Sprintf("nats://%s:%s@%s:%s", c.username, c.password, c.hostname, c.port)
}
