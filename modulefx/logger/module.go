package logger

import "go.uber.org/fx"

var Module = fx.Invoke(InitLogger)
