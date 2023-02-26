package tracing

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.uber.org/fx"
)

func InitTracing(lifecycle fx.Lifecycle) {
	var lg *log.Logger

	if viper.GetBool("debug.tracing") {
		lg = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Llongfile)
	}

	exporter, err := zipkin.New(viper.GetString("zipkin.url"), zipkin.WithLogger(lg))
	if err != nil {
		log.Println("Init tracing error", err)
		return
	}

	spanProcessor := trace.NewBatchSpanProcessor(exporter)

	tracerProvider := trace.NewTracerProvider(
		trace.WithSpanProcessor(spanProcessor),
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(viper.GetFloat64("zipkin.rate")))),
		trace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(viper.GetString("service.name")),
		)),
	)

	b3 := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
	otel.SetTextMapPropagator(b3)
	otel.SetTracerProvider(tracerProvider)

	lifecycle.Append(fx.Hook{
		OnStop: func(c context.Context) error {
			log.Println("Closing tracer")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return tracerProvider.Shutdown(ctx)
		},
	})
}
