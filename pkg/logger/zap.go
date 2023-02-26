package logger

import (
	"log"
	"time"

	"github.com/nmtri1912/go-common/utils/timeutils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(production bool) *zap.Logger {
	config := getConfig(production)
	// AddCallerSkip to skip report wrapper as caller in log message
	zapLogger, err := config.Build(
		zap.AddCallerSkip(1),
	)
	if err != nil {
		log.Fatal("Can not create logger ", err)
	}
	return zapLogger
}

func getConfig(production bool) zap.Config {
	if production {
		config := zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = simpleLogTimeEncoder
		return config
	}
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = simpleLogTimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return config
}

func simpleLogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(timeutils.GetTimeFormat(t))
}
