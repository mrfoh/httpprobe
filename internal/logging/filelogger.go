package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

type FileLoggerOptions struct {
	LogDirectory  string
	LogFilePrefix string
	MaxSize       int
	MaxAge        int
}

func NewFileLogger(opts *FileLoggerOptions) (*lumberjack.Logger, error) {
	if opts.LogDirectory == "" {
		return nil, fmt.Errorf("log directory is required, ensure LogDirectory is set")
	}

	if opts.LogFilePrefix == "" {
		return nil, fmt.Errorf("log file prefix is required, ensure LogFilePrefix is set")
	}

	if opts.MaxSize <= 0 {
		return nil, fmt.Errorf("max size must be greater than 0, ensure MaxSize is set")
	}

	if opts.MaxAge <= 0 {
		return nil, fmt.Errorf("max age must be greater than 0, ensure MaxAge is set")
	}

	if err := os.MkdirAll(opts.LogDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory - %s: %v", opts.LogDirectory, err)
	}

	logFilename := fmt.Sprintf("%s-%s.log", opts.LogFilePrefix, "app")
	filename := filepath.Join(opts.LogDirectory, logFilename)

	logger := &lumberjack.Logger{
		Filename: filename,
		MaxSize:  opts.MaxSize,
		MaxAge:   opts.MaxAge,
		Compress: true,
	}

	return logger, nil
}
