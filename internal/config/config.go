package config

type Database struct {
	DSN               string
	MaxConns          int32
	MinConns          int32
	MaxConnIdleTime   uint32
	HealthCheckPeriod uint32
	ConnectTimeout    uint32
}
