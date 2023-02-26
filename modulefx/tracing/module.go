package tracing

import "go.uber.org/fx"

var Module = fx.Invoke(InitTracing)
