package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Url        string `yaml:"url"`
	Token      string `yaml:"token"`
	configFile string
}

func GetConfigPath(defaultConfigFile string) (string, error) {
	if defaultConfigFile == "" {
		home, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("error fetching user config dir, error is '%s'", err)
		}
		configPath := path.Join(home, "trnotes")
		os.MkdirAll(configPath, os.ModePerm)
		configFilePath := path.Join(configPath, "config.yaml")
		return configFilePath, nil
	}
	stat, err := os.Stat(defaultConfigFile)
	if err != nil {
		return "", err
	}
	if stat.IsDir() {
		return "", fmt.Errorf("config should be a file, got dir")
	}
	parent := filepath.Dir(defaultConfigFile)
	os.MkdirAll(parent, os.ModePerm)
	return defaultConfigFile, nil
}

func GetExistingConfig(defaultConfigFile string) (*Config, error) {
	configFilePath, err := GetConfigPath(defaultConfigFile)
	if err != nil {
		return nil, err
	}
	conf, err := Parse(configFilePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error reading config, error is '%s'", err)
		}
		return nil, nil
	}
	conf.configFile = configFilePath
	return conf, err

}

func New(url, token, defaultConfigFile string) (*Config, error) {
	c := Config{Url: url, Token: token}
	configFile, err := GetConfigPath(defaultConfigFile)
	if err != nil {
		return nil, err
	}
	c.configFile = configFile
	return &c, err

}

func Parse(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	err = yaml.Unmarshal(file, conf)
	if err != nil {
		return nil, fmt.Errorf("error parsing config, error is '%s'", err)
	}
	conf.configFile = filename
	return conf, nil
}

func (c *Config) Save() error {
	if c.configFile == "" {
		return fmt.Errorf("missing config path")
	}
	content, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error saving config, error is '%s'", err)
	}
	return os.WriteFile(c.configFile, []byte(content), 0666)
}

func GetComputerName() (string, error) {
	return os.Hostname()
}
