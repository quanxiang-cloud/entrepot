package config

import (
	"github.com/quanxiang-cloud/cabin/tailormade/client"
	mysql2 "github.com/quanxiang-cloud/cabin/tailormade/db/mysql"
	redis2 "github.com/quanxiang-cloud/cabin/tailormade/db/redis"
	"io/ioutil"

	"github.com/quanxiang-cloud/cabin/logger"
	"gopkg.in/yaml.v2"
)

// Conf Global configuration
var Conf *Config

// DefaultPath default configuration path
var DefaultPath = "./configs/config.yml"

// Config configuration struct
type Config struct {
	Port          string        `yaml:"port"`
	InternalNet   client.Config `yaml:"internalNet"`
	FormHost      string        `yaml:"formHost"`
	FlowHost      string        `yaml:"flowHost"`
	PolyAPIHost   string        `yaml:"polyAPIHost"`
	StructorHost  string        `yaml:"structorHost"`
	AppCenterHost string        `yaml:"appCenterHost"`
	PersonaHost   string        `yaml:"personaHost"`
	Model         string        `yaml:"model"`
	Mysql         mysql2.Config `yaml:"mysql"`
	Log           logger.Config `yaml:"log"`
	ProcessorNum  int           `yaml:"processorNum"`
	Redis         redis2.Config `yaml:"redis"`
}

// NewConfig new a configuration
func NewConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath
	}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, &Conf)
	if err != nil {
		return nil, err
	}

	return Conf, nil
}
