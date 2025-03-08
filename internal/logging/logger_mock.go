package logging

import (
	"go.uber.org/zap"
)

// MockLogger is a mock implementation of the logging.Logger interface for testing
type MockLogger struct {
	// Track calls to logging methods
	DebugCalls []string
	InfoCalls  []string
	WarnCalls  []string
	ErrorCalls []string
	FatalCalls []string
}

// NewMockLogger creates a new instance of MockLogger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		DebugCalls: make([]string, 0),
		InfoCalls:  make([]string, 0),
		WarnCalls:  make([]string, 0),
		ErrorCalls: make([]string, 0),
		FatalCalls: make([]string, 0),
	}
}

func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
	m.DebugCalls = append(m.DebugCalls, msg)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.InfoCalls = append(m.InfoCalls, msg)
}

func (m *MockLogger) Warn(msg string, fields ...zap.Field) {
	m.WarnCalls = append(m.WarnCalls, msg)
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.ErrorCalls = append(m.ErrorCalls, msg)
}

func (m *MockLogger) Fatal(msg string, fields ...zap.Field) {
	m.FatalCalls = append(m.FatalCalls, msg)
}

func (m *MockLogger) With(fields ...zap.Field) Logger {
	return m
}

func (m *MockLogger) Sync() error {
	return nil
}
