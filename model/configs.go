package model

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Apollo struct {
	Conf Config `yaml:"apollo"`
}

type Config struct {
	ConfVersion    string            `yaml:"version"`
	Address        string            `yaml:"address"`
	Meridian       string            `yaml:"meridian"`
	DB             []string          `yaml:"db"`
	InstrumentConf map[string]string `yaml:"instrument"`
}

func ConfigFile(n ...string) (*Config, error) {
	path := "config.yml"
	if len(n) > 0 {
		path = n[0]
	}

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var conf Apollo
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return nil, err
	}
	return &conf.Conf, nil
}
