package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
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

type Auth struct {
	Secret         string        `mapstructure:"secret"`
	AccessTokenTTL time.Duration `mapstructure:"access_token_ttl"`
}

type CLI struct {
	APIBaseURL string `mapstructure:"api_base_url"`
	TokenFile  string `mapstructure:"token_file"`
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

type Mongo struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

type Config struct {
	Database Database  `mapstructure:"database"`
	HTTP     HTTP      `mapstructure:"http"`
	Storage  Storage   `mapstructure:"storage"`
	LogCfg   LogConfig `mapstructure:"log"`
	TechUI   TechUI    `mapstructure:"techui"`
	Mongo    Mongo     `mapstructure:"mongo"`
	Auth     Auth      `mapstructure:"auth"`
	CLI      CLI       `mapstructure:"cli"`
}

func Load() (*Config, error) {
	v := viper.New()

	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")

		if configDir := os.Getenv("CONFIG_DIR"); configDir != "" {
			v.AddConfigPath(configDir)
		}
		v.AddConfigPath(".")

		if root := findModuleRoot(); root != "" {
			v.AddConfigPath(root)
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("log.time_format", "2006-01-02T15:04:05.000Z07:00")
	v.SetDefault("log.file", "")
	v.SetDefault("auth.secret", "change-me")
	v.SetDefault("auth.access_token_ttl", "24h")
	v.SetDefault("cli.api_base_url", "http://localhost:8080/api/v2")
	v.SetDefault("cli.token_file", "")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func findModuleRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			return ""
		}
		cwd = parent
	}
}
