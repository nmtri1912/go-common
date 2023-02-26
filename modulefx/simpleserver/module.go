package simpleserver

import "go.uber.org/fx"

var Module = fx.Invoke(RunServer)
