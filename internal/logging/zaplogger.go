package logging

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Create a new zap logger
func NewZapLogger(opts *LoggerOptions) (Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()

	encoderConfig.EncodeName = zapcore.FullNameEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.CallerKey = "caller"

	var encoder zapcore.Encoder
	switch strings.ToLower(opts.LogFormat) {
	case "text":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		return nil, fmt.Errorf("invalid log format: must be 'json' or 'text'")
	}

	// Configure output
	var writeSyncer zapcore.WriteSyncer
	switch strings.ToLower(opts.LogOutput) {
	case "console":
		writeSyncer = zapcore.AddSync(os.Stdout)
	case "file":
		fileLogger, err := NewFileLogger(&FileLoggerOptions{
			LogDirectory:  opts.LogDirectory,
			LogFilePrefix: opts.LogFilePrefix,
			MaxSize:       opts.MaxFileSize,
			MaxAge:        7,
		})
		if err != nil {
			return nil, err
		}
		writeSyncer = zapcore.AddSync(fileLogger)
	default:
		return nil, fmt.Errorf("invalid log output: must be 'console' or 'file'")
	}

	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		ParseLogLevel(opts.LogLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return logger, nil
}

func ParseLogLevel(level string) zap.AtomicLevel {
	// ensure the level is uppercase
	level = strings.ToUpper(level)

	switch level {
	case "DEBUG":
		return zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "INFO":
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "WARN":
		return zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "ERROR":
		return zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case "FATAL":
		return zap.NewAtomicLevelAt(zapcore.FatalLevel)
	default:
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
}
