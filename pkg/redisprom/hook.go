package redisprom

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	Hook struct {
		options *Options

		commandCounter      *prometheus.CounterVec
		commandErrorCounter *prometheus.CounterVec
		commandDuration     *prometheus.HistogramVec
	}

	startKey struct{}
)

var (
	labelNames = []string{"command"}
)

// NewHook creates a new go-redis hook instance and registers Prometheus collectors.
func NewHook(opts ...Option) *Hook {
	options := DefaultOptions()
	options.Merge(opts...)

	commandCounter := register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "redis_command_total",
		Help: "Total redis command",
	}, labelNames)).(*prometheus.CounterVec)

	commandErrorCounter := register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "redis_command_error_total",
		Help: "Total error redis command",
	}, labelNames)).(*prometheus.CounterVec)

	commandDuration := register(prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: options.Namespace,
		Name:      "redis_command_duration",
		Help:      "Redis command latencies in seconds",
		Buckets:   options.DurationBuckets,
	}, labelNames)).(*prometheus.HistogramVec)

	return &Hook{
		options:             options,
		commandCounter:      commandCounter,
		commandErrorCounter: commandErrorCounter,
		commandDuration:     commandDuration,
	}
}

func (hook *Hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (hook *Hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if start, ok := ctx.Value(startKey{}).(time.Time); ok {
		duration := time.Since(start).Seconds()
		hook.commandDuration.WithLabelValues(cmd.Name()).Observe(duration)
	}

	hook.commandCounter.WithLabelValues(cmd.Name()).Inc()

	if isActualErr(cmd.Err()) {
		hook.commandErrorCounter.WithLabelValues(cmd.Name()).Inc()
	}

	return nil
}

func (hook *Hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (hook *Hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if err := hook.AfterProcess(ctx, redis.NewCmd(ctx, "pipeline")); err != nil {
		return err
	}

	return nil
}

func register(collector prometheus.Collector) prometheus.Collector {
	err := prometheus.DefaultRegisterer.Register(collector)
	if err == nil {
		return collector
	}

	if arErr, ok := err.(prometheus.AlreadyRegisteredError); ok {
		return arErr.ExistingCollector
	}

	panic(err)
}

func isActualErr(err error) bool {
	return err != nil && err != redis.Nil
}
