package config

import "time"

var Version = "dev"

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Proxy   ProxyConfig   `mapstructure:"proxy"`
	Storage StorageConfig `mapstructure:"storage"`
	Log     LogConfig     `mapstructure:"log"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type ProxyConfig struct {
	ListenAddr   string        `mapstructure:"listen_addr"`
	SamplingRate float64       `mapstructure:"sampling_rate"`
	Routes       []RouteConfig `mapstructure:"routes"`
	Buffer       BufferConfig  `mapstructure:"buffer"`
}

type RouteConfig struct {
	PathPrefix   string  `mapstructure:"path_prefix"`
	Target       string  `mapstructure:"target"`
	PIIDetection bool    `mapstructure:"pii_detection"`
	SamplingRate float64 `mapstructure:"sampling_rate"`
}

type BufferConfig struct {
	QueueSize int `mapstructure:"queue_size"`
	Workers   int `mapstructure:"workers"`
}

type StorageConfig struct {
	PostgresURL string `mapstructure:"postgres_url"`
	RedisURL    string `mapstructure:"redis_url"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type DiffConfig struct {
	IgnorePaths      []string `mapstructure:"ignore_paths"`
	TreatArraysAsSet bool     `mapstructure:"treat_arrays_as_set"`
}

type ReplayConfig struct {
	Workers   int           `mapstructure:"workers"`
	RateLimit int           `mapstructure:"rate_limit"`
	Timeout   time.Duration `mapstructure:"timeout"`
}
