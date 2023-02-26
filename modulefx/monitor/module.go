package monitor

import (
	"github.com/nmtri1912/go-common/pkg/monitor"
	"go.uber.org/fx"
)

var Module = fx.Invoke(monitor.InitMonitorMetrics)
