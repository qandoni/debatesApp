package core_transport_websocket

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AllowedOrigins []string      `envconfig:"ALLOWED_ORIGINS" default:"http://localhost:3000"`
	ReadLimit      int64         `envconfig:"READ_LIMIT"      default:"4096"`
	WriteWait      time.Duration `envconfig:"WRITE_WAIT"      default:"10s"`
	PongWait       time.Duration `envconfig:"PONG_WAIT"       default:"60s"`
	PingPeriod     time.Duration `envconfig:"PING_PERIOD"     default:"54s"`
	SendBufferSize int           `envconfig:"SEND_BUFFER_SIZE" default:"64"`
	AuthWait       time.Duration `envconfig:"AUTH_WAIT"       default:"10s"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("WS", &config); err != nil {
		return Config{}, fmt.Errorf("proccess envconfig: %w", err)
	}

	config.AllowedOrigins = normalizeOrigins(config.AllowedOrigins)
	if len(config.AllowedOrigins) == 0 {
		return Config{}, fmt.Errorf("WS_ALLOWED_ORIGINS is empty: no client origin will be accepted")
	}
	if config.ReadLimit <= 0 {
		return Config{}, fmt.Errorf("WS_READ_LIMIT must be positive, got: %d", config.ReadLimit)
	}
	if config.SendBufferSize <= 0 {
		return Config{}, fmt.Errorf("WS_SEND_BUFFER_SIZE must be positive, got: %d", config.SendBufferSize)
	}
	if config.WriteWait <= 0 || config.PongWait <= 0 || config.AuthWait <= 0 {
		return Config{}, fmt.Errorf(
			"WS_WRITE_WAIT, WS_PONG_WAIT and WS_AUTH_WAIT must be positive, got: %s, %s, %s",
			config.WriteWait, config.PongWait, config.AuthWait,
		)
	}
	if config.PingPeriod >= config.PongWait {
		return Config{}, fmt.Errorf("WS_PING_PERIOD (%s) must be less than WS_PONG_WAIT (%s), otherwise server drops alive connections",
			config.PingPeriod, config.PongWait,
		)
	}
	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get WS config")
		panic(err)
	}
	return config
}

func normalizeOrigins(origins []string) []string {
	normalized := make([]string, 0, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		normalized = append(normalized, origin)
	}
	return normalized
}
