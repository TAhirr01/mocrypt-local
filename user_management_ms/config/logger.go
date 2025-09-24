package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",    // The main log message
		LevelKey:       "level",  // Log level
		TimeKey:        "time",   // Timestamp key
		NameKey:        "logger", // Logger name (optional)
		CallerKey:      "",       // Disable caller info
		FunctionKey:    "",       // Disable function info
		StacktraceKey:  "",       // Disable stacktrace
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // INFO, WARN, ERROR
		EncodeTime:     zapcore.ISO8601TimeEncoder,  // human-readable timestamp
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // JSON logs
		zapcore.AddSync(zapcore.Lock(os.Stdout)),
		zap.InfoLevel, // Minimum log level
	)

	return zap.New(core)
}
