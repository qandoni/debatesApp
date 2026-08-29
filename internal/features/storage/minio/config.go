package minio

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("MINIO", &config); err != nil {
		return Config{}, fmt.Errorf("proccess envconfig: %w", err)
	}
	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get MinIO storage config: %w", err)
		panic(err)
	}
	return config
}

type Config struct {
	Endpoint  string `envconfig:"ENDPOINT" required:"true"`
	AccessKey string `envconfig:"ROOT_USER" required:"true"`
	SecretKey string `envconfig:"ROOT_PASSWORD" required:"true"`
	Bucket    string `envconfig:"BUCKET" required:"true"`
	Secure    bool   `envconfig:"SECURE" required:"true"`
}
