package httpserver

import "go.uber.org/fx"

var Module = fx.Invoke(RunServer)
