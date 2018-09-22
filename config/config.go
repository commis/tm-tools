package config

type BaseConfig struct {
	// Output level for logging
	LogLevel string `mapstructure:"log_level"`
}

func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		LogLevel: DefaultLogLevel(),
	}
}

func DefaultLogLevel() string {
	return "debug"
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
