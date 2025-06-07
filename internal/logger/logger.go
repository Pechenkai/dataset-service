package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"ppo/internal/config"
)

type disablingSyncer struct {
	inner zapcore.WriteSyncer
	mu    sync.Mutex
	dead  bool
}

func (d *disablingSyncer) Write(p []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.dead {
		return len(p), nil
	}

	n, err := d.inner.Write(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logger: отключен после первой ошибки записи: %v\n", err)
		d.dead = true
		return len(p), nil
	}
	return n, nil
}

func (d *disablingSyncer) Sync() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.dead {
		return nil
	}

	if err := d.inner.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "Logger: отключен после первой ошибки Sync(): %v\n", err)
		d.dead = true
		return nil
	}
	return nil
}

func NewLogger(logCfg config.LogConfig) (*zap.Logger, func(), error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logCfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(logCfg.TimeFormat))
		},
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	switch logCfg.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	default:
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	var writer zapcore.WriteSyncer
	if logCfg.FilePath != "" {
		f, err := os.OpenFile(logCfg.FilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("нет прав окрыть log file %q: %w", logCfg.FilePath, err)
		}
		writer = &disablingSyncer{inner: zapcore.AddSync(f)}
	} else {
		writer = &disablingSyncer{inner: zapcore.AddSync(os.Stdout)}
	}

	core := zapcore.NewCore(encoder, writer, level)
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	closeFn := func() {
		_ = logger.Sync()
	}

	return logger, closeFn, nil
}
