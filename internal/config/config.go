package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
}

func New(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	err = yaml.Unmarshal(file, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
