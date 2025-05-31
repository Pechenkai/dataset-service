package config

import (
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
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

type LogConfig struct {
	Level      string `env:"LOG_LEVEL" envDefault:"info"`
	Format     string `env:"LOG_FORMAT" envDefault:"console"` // "console" или "json"
	TimeFormat string `env:"LOG_TIME_FORMAT" envDefault:"2006-01-02T15:04:05.000Z07:00"`
}
type Config struct {
	Database Database
	HTTP     HTTP
	Storage  Storage
	LogCfg   LogConfig
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
