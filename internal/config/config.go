package config

import (
	"github.com/spf13/viper"
	"time"
)

type Database struct {
	DSN               string        `mapstructure:"dsn"`
	MaxConns          int32         `mapstructure:"max_conns"`
	MinConns          int32         `mapstructure:"min_conns"`
	MaxConnIdleTime   time.Duration `mapstructure:"max_conn_idle_time"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
	ConnectTimeout    time.Duration `mapstructure:"connect_timeout"`
}

type HTTP struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type Storage struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Region    string `mapstructure:"region"`
	Bucket    string `mapstructure:"bucket"`
}

type TechUI struct {
	Host string `env:"TECHUI_HOST" envDefault:"0.0.0.0"`
	Port int    `env:"TECHUI_PORT" envDefault:"8090"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	TimeFormat string `mapstructure:"time_format"`
	FilePath   string `mapstructure:"file"`
}

type Config struct {
	Database Database  `mapstructure:"database"`
	HTTP     HTTP      `mapstructure:"http"`
	Storage  Storage   `mapstructure:"storage"`
	LogCfg   LogConfig `mapstructure:"log"`
	TechUI   TechUI    `mapstructure:"techui"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AutomaticEnv()

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("log.time_format", "2006-01-02T15:04:05.000Z07:00")
	v.SetDefault("log.file", "")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
