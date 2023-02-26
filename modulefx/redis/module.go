package redis

import "go.uber.org/fx"

var Module = fx.Provide(NewCache)
