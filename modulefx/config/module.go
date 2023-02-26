package config

import (
	"go.uber.org/fx"
)

var Module = fx.Invoke(LoadConfiguration)
