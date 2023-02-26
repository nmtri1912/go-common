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
)

func InitializeTracing() func() {
	var logger *log.Logger

	if viper.GetBool("log.tracing") {
		logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Llongfile)
	}

	exporter, err := zipkin.New(viper.GetString("zipkin.url"), zipkin.WithLogger(logger))
	if err != nil {
		log.Println("Failed to init tracer")
	}

	spanProcessor := trace.NewBatchSpanProcessor(exporter)

	tracerProvider := trace.NewTracerProvider(
		trace.WithSpanProcessor(spanProcessor),
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(viper.GetFloat64("zipkin.rate")))),
		trace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(viper.GetString("service.name")+"-"+viper.GetString("service.env")),
		)),
	)

	b3 := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
	otel.SetTextMapPropagator(b3)

	otel.SetTracerProvider(tracerProvider)

	closeTracer := func() {
		log.Println("Closing tracer")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err = tracerProvider.Shutdown(ctx); err != nil {
			log.Println("Error when close tracer ", err)
		}
	}
	return closeTracer
}
