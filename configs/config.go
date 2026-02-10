package configs

import (
	"apollo/configs/nats"
	"apollo/configs/server"
	"apollo/configs/tracing"
)

type Config interface {
	Server() server.Config
	Nats() nats.Config
	Tracing() tracing.Config
}

type config struct {
	server  server.Config
	nats    nats.Config
	tracing tracing.Config
}

func NewConfig() (Config, error) {
	return &config{
		server:  server.NewConfig(),
		nats:    nats.NewConfig(),
		tracing: tracing.NewConfig(),
	}, nil
}

func (c config) Server() server.Config {
	return c.server
}

func (c config) Nats() nats.Config {
	return c.nats
}

func (c config) Tracing() tracing.Config {
	return c.tracing
}
