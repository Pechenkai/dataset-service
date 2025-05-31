package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

type Database struct {
	DSN               string        `env:"DB_DSN,required"`
	MaxConns          int32         `env:"DB_MAX_CONNS" envDefault:"20"`
	MinConns          int32         `env:"DB_MIN_CONNS" envDefault:"2"`
	MaxConnIdleTime   time.Duration `env:"DB_MAX_CONN_IDLE_TIME" envDefault:"60s"`
	HealthCheckPeriod time.Duration `env:"DB_HEALTH_CHECK_PERIOD" envDefault:"30s"`
	ConnectTimeout    time.Duration `env:"DB_CONNECT_TIMEOUT" envDefault:"5s"`
}

type HTTP struct {
	Host         string        `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port         int           `env:"HTTP_PORT" envDefault:"8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"10s"`
}

type Storage struct {
	Endpoint  string `env:"S3_ENDPOINT,required"`
	AccessKey string `env:"S3_ACCESS_KEY,required"`
	SecretKey string `env:"S3_SECRET_KEY,required"`
	Region    string `env:"S3_REGION" envDefault:"us-east-1"`
	Bucket    string `env:"S3_BUCKET,required"`
}

type Config struct {
	Database Database
	HTTP     HTTP
	Storage  Storage
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(&cfg.Database); err != nil {
		return nil, err
	}
	if err := env.Parse(&cfg.HTTP); err != nil {
		return nil, err
	}
	if err := env.Parse(&cfg.Storage); err != nil {
		return nil, err
	}
	return cfg, nil
}
