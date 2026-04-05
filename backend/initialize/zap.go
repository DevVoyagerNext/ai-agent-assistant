package initialize

import (
	"backend/global"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Zap() *zap.Logger {
	cfg := global.GVA_CONFIG.Log

	enabled := true
	if global.GVA_VP != nil && global.GVA_VP.IsSet("log.enabled") {
		enabled = global.GVA_VP.GetBool("log.enabled")
	}
	if !enabled {
		return zap.NewNop()
	}

	levelText := strings.TrimSpace(strings.ToLower(cfg.Level))
	if levelText == "" {
		levelText = "info"
	}

	var lvl zapcore.Level
	switch levelText {
	case "debug":
		lvl = zapcore.DebugLevel
	case "info":
		lvl = zapcore.InfoLevel
	case "warn", "warning":
		lvl = zapcore.WarnLevel
	case "error":
		lvl = zapcore.ErrorLevel
	default:
		lvl = zapcore.InfoLevel
	}

	encoding := strings.TrimSpace(strings.ToLower(cfg.Format))
	if encoding == "" {
		encoding = "console"
	}
	if encoding != "json" && encoding != "console" {
		encoding = "console"
	}

	outputPaths := cfg.OutputPaths
	if len(outputPaths) == 0 {
		outputPaths = []string{"stdout"}
	}
	errorOutputPaths := cfg.ErrorOutputPaths
	if len(errorOutputPaths) == 0 {
		errorOutputPaths = []string{"stderr"}
	}

	ensureOutputDirs := func(paths []string) {
		for _, p := range paths {
			path := strings.TrimSpace(p)
			if path == "" {
				continue
			}
			switch strings.ToLower(path) {
			case "stdout", "stderr":
				continue
			}

			dir := filepath.Dir(path)
			if dir == "" || dir == "." {
				continue
			}
			_ = os.MkdirAll(dir, 0755)
		}
	}
	ensureOutputDirs(outputPaths)
	ensureOutputDirs(errorOutputPaths)

	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(lvl),
		Development:       encoding == "console",
		Encoding:          encoding,
		DisableCaller:     cfg.DisableCaller,
		DisableStacktrace: cfg.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorOutputPaths,
	}

	logger, err := zapCfg.Build()
	if err != nil {
		fallbackCfg := zap.NewProductionConfig()
		fallbackCfg.OutputPaths = []string{"stdout"}
		fallbackCfg.ErrorOutputPaths = []string{"stderr"}
		fallbackLogger, fallbackErr := fallbackCfg.Build()
		if fallbackErr != nil {
			return zap.NewNop()
		}
		fallbackLogger.Error("日志初始化失败，已使用默认配置", zap.Error(err))
		return fallbackLogger
	}

	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		logger = logger.With(zap.String("host", hostname))
	}

	return logger
}
