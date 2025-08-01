package logger_test

import (
	"testing"

	"go.uber.org/zap/zapcore"
	"product-catalog/internal/logger"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		levelStr string
		wantLvl  zapcore.Level
	}{
		{"production info", "prod", "info", zapcore.InfoLevel},
		{"development debug", "dev", "debug", zapcore.DebugLevel},
		{"invalid level fallback", "local", "invalid", zapcore.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Init panicked: %v", r)
				}
			}()

			logger.Init(tt.env, tt.levelStr)

			if logger.Log == nil {
				t.Fatal("expected logger to be initialized, got nil")
			}

			if !logger.Log.Core().Enabled(tt.wantLvl) {
				t.Errorf("expected log level %s to be enabled", tt.wantLvl)
			}
		})
	}
}
