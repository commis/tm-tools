package config

type BaseConfig struct {
	// Output level for logging
	LogLevel string `mapstructure:"log_level"`

	// Database directory
	DBPath string `mapstructure:"db_dir"`
}

func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		LogLevel: DefaultLogLevel(),
		DBPath:   "data",
	}
}

func DefaultLogLevel() string {
	return "error"
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
