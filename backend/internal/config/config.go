package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Log        LogConfig        `mapstructure:"log"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

type PrometheusConfig struct {
	Timeout   int `mapstructure:"timeout"`
	MaxRetries int `mapstructure:"max_retries"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

var GlobalConfig *Config

func Init(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return err
	}

	return nil
}
