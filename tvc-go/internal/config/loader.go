package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("proxy.listen_addr", ":8080")
	v.SetDefault("proxy.sampling_rate", 0.1)
	v.SetDefault("proxy.buffer.queue_size", 10000)
	v.SetDefault("proxy.buffer.workers", 20)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("tvc")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.tvc")
		v.AddConfigPath("/etc/tvc")
	}

	v.SetEnvPrefix("TVC")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}
