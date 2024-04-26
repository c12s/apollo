package configs

import (
	"apollo/configs/server"
	"apollo/configs/nats"
)

type Config interface {
	Server() server.Config
	Nats()	nats.Config
}

type config struct {
	server server.Config
	nats   nats.Config
}

func NewConfig() (Config, error) {
	return &config{
		server: server.NewConfig(),
		nats: 	nats.NewConfig(),
	}, nil
}

func (c config) Server() server.Config {
	return c.server
}

func (c config) Nats() nats.Config {
	return c.nats
}
