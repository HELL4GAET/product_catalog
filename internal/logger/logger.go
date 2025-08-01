package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(env string, levelStr string) {
	cfg := zap.NewProductionConfig()
	if env == "local" || env == "dev" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(levelStr)); err == nil {
		cfg.Level = zap.NewAtomicLevelAt(lvl)
	}
	
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l, err := cfg.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic("cannot initialize logger: " + err.Error())
	}
	Log = l
}
