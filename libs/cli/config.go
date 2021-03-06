package cli

import "github.com/commis/tm-tools/libs/log"

type BaseConfig struct {
	// Output level for logging
	LogLevel int `mapstructure:"log_level"`
}

func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		LogLevel: DefaultLogLevel(),
	}
}

func DefaultLogLevel() int {
	return log.InfoLog
}

type Config struct {
	// Top level options use an anonymous struct
	BaseConfig `mapstructure:",squash"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: DefaultBaseConfig(),
	}
}
