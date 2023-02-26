package monitor

import (
	"sync"
	"time"

	"github.com/nmtri1912/go-common/pkg/logger"
	"github.com/nmtri1912/go-common/utils/errorutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

var mu = sync.Mutex{}

type MonitorRecorder struct {
	invocationErrorCounter *prometheus.CounterVec
	durationBuckets        *prometheus.HistogramVec
}

var GlobalRecorder *MonitorRecorder = nil

func InitMonitorMetrics() {
	mu.Lock()
	defer mu.Unlock()
	if GlobalRecorder != nil {
		return
	}
	logger.L().Info("Init monitor metrics...")
	GlobalRecorder = NewMonitorRecorder()
	logger.L().Info("Inited monitor metrics")
}

func NewMonitorRecorder() *MonitorRecorder {
	constLabels := prometheus.Labels{
		"application": viper.GetString("service.name"),
	}

	labels := []string{"name", "type", "method"}
	durationBuckets := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "invocations_seconds",
		Help:        "Duration buckets",
		Buckets:     GetHistorgramBuckets(),
		ConstLabels: constLabels,
	}, labels)

	labels = []string{"name", "type", "method", "error_code"}
	invocationErrorCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "invocations_errors_total",
		Help:        "Number of invocation errors",
		ConstLabels: constLabels,
	}, labels)

	systemMetrics := &SystemMetrics{
		goroutineCountDesc: prometheus.NewDesc("goroutine_count", "Number of active goroutine", nil, constLabels),
		systemCPUCoreDesc:  prometheus.NewDesc("system_cpu_count", "CPU core count", nil, constLabels),
		systemUptimeDesc:   prometheus.NewDesc("process_uptime_seconds", "System uptime in seconds", nil, constLabels),
		startTime:          time.Now(),
	}

	prometheus.MustRegister(invocationErrorCounter, durationBuckets, systemMetrics)
	return &MonitorRecorder{
		invocationErrorCounter: invocationErrorCounter,
		durationBuckets:        durationBuckets,
	}
}

func RecordMetrics(service, metricType, method string, start time.Time, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.L().Warn("recovered in RecordMetrics")
		}
	}()

	duration := time.Since(start).Seconds()
	GlobalRecorder.durationBuckets.WithLabelValues(service, metricType, method).Observe(duration)
	if err != nil {
		_, errorReason, _ := errorutils.ExtractReasonAndDomainFromError(err, "")
		GlobalRecorder.invocationErrorCounter.
			WithLabelValues(service, metricType, method, errorReason).
			Inc()
	}

}
