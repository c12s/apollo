package server


type Config interface {
	Port() string
}

type config struct {
	port string
}

func NewConfig() Config {
	return config{
		port: "8002",
	}
}

func (c config) Port() string {
	return c.port
}
