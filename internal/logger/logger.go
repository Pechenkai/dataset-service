package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"ppo/internal/config"
)

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
		// открываем (или создаём) файл в режиме append
		f, err := os.OpenFile(logCfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, nil, err
		}
		writer = zapcore.AddSync(f)
	} else {
		writer = zapcore.AddSync(os.Stdout)
	}
	core := zapcore.NewCore(encoder, writer, level)

	logger := zap.New(core,
		zap.AddCaller(),                       // включаем вывод caller
		zap.AddStacktrace(zapcore.ErrorLevel), // стектрейс, если уровень ≥ ERROR
	)

	closeFn := func() {
		_ = logger.Sync()
	}

	return logger, closeFn, nil
}
