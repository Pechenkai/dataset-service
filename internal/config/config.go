package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
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
	TwoFA          TwoFA         `mapstructure:"twofa"`
}

type CLI struct {
	APIBaseURL string `mapstructure:"api_base_url"`
	TokenFile  string `mapstructure:"token_file"`
}

type AdminAccount struct {
	Username string `mapstructure:"username"`
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
	Country  string `mapstructure:"country"`
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

type Broker struct {
	URL                   string        `mapstructure:"url"`
	GatewayToCoreQueue    string        `mapstructure:"gateway_to_core_queue"`
	CoreToGatewayQueue    string        `mapstructure:"core_to_gateway_queue"`
	CoreToDataQueue       string        `mapstructure:"core_to_data_queue"`
	DataToCoreQueue       string        `mapstructure:"data_to_core_queue"`
	Prefetch              int           `mapstructure:"prefetch"`
	ReconnectDelay        time.Duration `mapstructure:"reconnect_delay"`
	MessageProcessTimeout time.Duration `mapstructure:"message_process_timeout"`
	EnableDLQ             bool          `mapstructure:"enable_dlq"`
}

type TwoFA struct {
	CodeTTL       time.Duration `mapstructure:"code_ttl"`
	MaxAttempts   int           `mapstructure:"max_attempts"`
	BlockDuration time.Duration `mapstructure:"block_duration"`
	DebugSecret   string        `mapstructure:"debug_secret"`
	Delivery      string        `mapstructure:"delivery"`
}

type Config struct {
	Database Database     `mapstructure:"database"`
	HTTP     HTTP         `mapstructure:"http"`
	Storage  Storage      `mapstructure:"storage"`
	LogCfg   LogConfig    `mapstructure:"log"`
	TechUI   TechUI       `mapstructure:"techui"`
	Mongo    Mongo        `mapstructure:"mongo"`
	Auth     Auth         `mapstructure:"auth"`
	CLI      CLI          `mapstructure:"cli"`
	Admin    AdminAccount `mapstructure:"admin"`
	Broker   Broker       `mapstructure:"broker"`
}

func Load() (*Config, error) {
	// best-effort: load .env in current and module root so env overrides yaml defaults
	_ = godotenv.Load()
	if root := findModuleRoot(); root != "" {
		_ = godotenv.Load(filepath.Join(root, ".env"))
	}

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
	v.SetDefault("auth.twofa.code_ttl", "5m")
	v.SetDefault("auth.twofa.max_attempts", 3)
	v.SetDefault("auth.twofa.block_duration", "2m")
	v.SetDefault("auth.twofa.debug_secret", "")
	v.SetDefault("auth.twofa.delivery", "email")
	v.SetDefault("cli.api_base_url", "http://localhost:8080/api/v2")
	v.SetDefault("cli.token_file", "")
	v.SetDefault("admin.username", "admin")
	v.SetDefault("admin.email", "admin@example.com")
	v.SetDefault("admin.password", "admin123")
	v.SetDefault("admin.country", "RU")
	v.SetDefault("broker.url", "amqp://guest:guest@localhost:5672/")
	v.SetDefault("broker.gateway_to_core_queue", "gateway.core.cmd")
	v.SetDefault("broker.core_to_gateway_queue", "core.gateway.evt")
	v.SetDefault("broker.core_to_data_queue", "core.data.cmd")
	v.SetDefault("broker.data_to_core_queue", "data.core.evt")
	v.SetDefault("broker.prefetch", 10)
	v.SetDefault("broker.reconnect_delay", "2s")
	v.SetDefault("broker.message_process_timeout", "10s")
	v.SetDefault("broker.enable_dlq", true)

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
