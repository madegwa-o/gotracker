package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	HTTPAddr       string
	CloudWSURL     string
	BufferSize     int
	ReconnectDelay time.Duration
	FlushInterval  time.Duration
	WriteTimeout   time.Duration
	LogLevel       string
}

func Load() Config {
	v := viper.New()
	v.SetEnvPrefix("EDGE")
	v.AutomaticEnv()

	v.SetDefault("HTTP_ADDR", ":8080")
	v.SetDefault("CLOUD_WS_URL", "ws://localhost:9090/ws/cloud")
	v.SetDefault("BUFFER_SIZE", 10000)
	v.SetDefault("RECONNECT_DELAY", "3s")
	v.SetDefault("FLUSH_INTERVAL", "1s")
	v.SetDefault("WRITE_TIMEOUT", "5s")
	v.SetDefault("LOG_LEVEL", "info")

	return Config{
		HTTPAddr:       v.GetString("HTTP_ADDR"),
		CloudWSURL:     v.GetString("CLOUD_WS_URL"),
		BufferSize:     v.GetInt("BUFFER_SIZE"),
		ReconnectDelay: mustParseDuration(v.GetString("RECONNECT_DELAY"), 3*time.Second),
		FlushInterval:  mustParseDuration(v.GetString("FLUSH_INTERVAL"), time.Second),
		WriteTimeout:   mustParseDuration(v.GetString("WRITE_TIMEOUT"), 5*time.Second),
		LogLevel:       v.GetString("LOG_LEVEL"),
	}
}

func mustParseDuration(raw string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return d
}
