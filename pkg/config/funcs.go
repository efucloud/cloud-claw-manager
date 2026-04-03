package config

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

const (
	defaultLogLevel      = "info"
	defaultLogFilename   = "./log/cloud-claw-manager.log"
	defaultLogMaxSize    = 100
	defaultLogMaxAge     = 30
	defaultLogMaxBackups = 10
)

func boolPtr(v bool) *bool {
	return &v
}

func normalizeLogConfig(conf *LogConfig) *LogConfig {
	if conf == nil {
		conf = &LogConfig{}
	}
	conf.Level = strings.ToLower(strings.TrimSpace(conf.Level))
	if conf.Level == "" {
		conf.Level = defaultLogLevel
	}
	if strings.TrimSpace(conf.Filename) == "" {
		conf.Filename = defaultLogFilename
	}
	if conf.MaxSize <= 0 {
		conf.MaxSize = defaultLogMaxSize
	}
	if conf.MaxAge <= 0 {
		conf.MaxAge = defaultLogMaxAge
	}
	if conf.MaxBackups <= 0 {
		conf.MaxBackups = defaultLogMaxBackups
	}
	if conf.Production == nil {
		conf.Production = boolPtr(true)
	}
	return conf
}

func logConfig(conf *LogConfig) {
	conf = normalizeLogConfig(conf)
	writeSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   conf.Filename,
		MaxSize:    conf.MaxSize,
		MaxBackups: conf.MaxBackups,
		MaxAge:     conf.MaxAge,
		LocalTime:  conf.LocalTime,
		Compress:   conf.Compress,
	})
	var encoderConfig zapcore.EncoderConfig
	if conf.Production != nil && *conf.Production {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	var level zapcore.Level
	switch strings.ToLower(strings.TrimSpace(conf.Level)) {
	case "info":
		level = zapcore.InfoLevel
	case "debug":
		level = zapcore.DebugLevel
	case "warn", "warning":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}
	core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writeSyncer, zapcore.AddSync(os.Stdout)), level)
	logger := zap.New(core, zap.AddCaller())
	Logger = logger.Sugar()

}

func (c *Config) Init() {
	c.LogConfig = normalizeLogConfig(c.LogConfig)
	logConfig(c.LogConfig)
	c.OidcConfig.Issuer = strings.TrimSuffix(c.OidcConfig.Issuer, "/")
	c.Default(context.Background())

}
func GetLangFromCtx(ctx context.Context) (lang string) {
	lang = "zh"
	lan := ctx.Value(RequestLanguage)
	if lan != nil {
		lang = lan.(string)
	}
	return lang
}
func GetOperatorFromCtx(ctx context.Context) (operator string) {
	operator = "unknown"
	requesterId := ctx.Value(RequestUserId)
	if requesterId != nil {
		operator = requesterId.(string)
	}
	return operator
}
