package logger

import (
	"context"

	"github.com/nmtri1912/go-common/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func InitLogger(lifecycle fx.Lifecycle) {
	logger.InitLogger(!viper.GetBool("debug.logger"))
	lifecycle.Append(fx.Hook{OnStop: func(ctx context.Context) error {
		_ = logger.Sync()
		return nil
	}})
}
