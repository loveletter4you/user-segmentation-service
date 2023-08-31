package config

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Server struct {
		Port string `envconfig:"SERVER_PORT" yaml:"Port"`
	} `yaml:"Server"`

	Database struct {
		Name     string `envconfig:"DB_NAME"`
		Host     string `envconfig:"DB_HOST"`
		Port     string `envconfig:"DB_PORT"`
		Username string `envconfig:"DB_ROOT"`
		Password string `envconfig:"DB_PASSWORD"`
	}

	DatabaseConnection struct {
		Attempts int `yaml:"Attempts"`
		Timeout  int `yaml:"Timeout"`
	} `yaml:"DatabaseConnection"`
}

func NewConfig(configPath string) (*Config, error) {
	var cfg Config
	f, err := os.Open("/" + configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	err = envconfig.Process("", &cfg)
	return &cfg, err
}
