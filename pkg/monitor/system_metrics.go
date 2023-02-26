package monitor

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SystemMetrics struct {
	goroutineCountDesc *prometheus.Desc
	systemCPUCoreDesc  *prometheus.Desc
	systemUptimeDesc   *prometheus.Desc
	startTime          time.Time
}

func (m *SystemMetrics) Describe(c chan<- *prometheus.Desc) {
	c <- m.goroutineCountDesc
	c <- m.systemCPUCoreDesc
	c <- m.systemUptimeDesc
}

func (m *SystemMetrics) Collect(c chan<- prometheus.Metric) {
	metric, err := prometheus.NewConstMetric(m.goroutineCountDesc, prometheus.GaugeValue, float64(runtime.NumGoroutine()))
	if err == nil {
		c <- metric
	}
	metric, err = prometheus.NewConstMetric(m.systemCPUCoreDesc, prometheus.GaugeValue, float64(runtime.NumCPU()))
	if err == nil {
		c <- metric
	}
	metric, err = prometheus.NewConstMetric(m.systemUptimeDesc, prometheus.GaugeValue, time.Since(m.startTime).Seconds())
	if err == nil {
		c <- metric
	}
}
