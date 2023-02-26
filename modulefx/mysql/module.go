package mysql

import "go.uber.org/fx"

var Module = fx.Provide(NewDB)
