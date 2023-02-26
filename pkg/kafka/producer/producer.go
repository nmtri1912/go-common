package producer

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	"github.com/nmtri1912/go-common/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type KafkaProducer interface {
	Close()
	Produce(topic string, key string, value []byte)
	ProduceCtx(ctx context.Context, topic string, key string, value []byte)
}

type kafkaProducerImpl struct {
	producer sarama.AsyncProducer
}

func NewKafkaProducer(brokers []string) KafkaProducer {
	// https://github.com/Shopify/sarama/blob/main/examples/http_server/http_server.go#L219
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Metadata.Timeout = time.Second * 3

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		logger.L().Fatal("Error creating kafka producer", zap.Error(err))
	}

	logResultMessage(producer)

	return &kafkaProducerImpl{producer: producer}
}

func logResultMessage(producer sarama.AsyncProducer) {
	// log error message
	go func() {
		for err := range producer.Errors() {
			key, _ := err.Msg.Key.Encode()
			value, _ := err.Msg.Value.Encode()
			logger.L().Info("Failed to push kafka message",
				zap.String("topic", err.Msg.Topic),
				zap.Int32("partition", err.Msg.Partition),
				zap.Int64("offset", err.Msg.Offset),
				zap.String("key", string(key)),
				zap.String("value", string(value)),
				zap.Error(err),
			)
		}
	}()

	// log success message
	go func() {
		for result := range producer.Successes() {
			wResult := otelsarama.NewProducerMessageCarrier(result)
			ctx := otel.GetTextMapPropagator().Extract(context.Background(), wResult)
			spanCtx := trace.SpanContextFromContext(ctx)

			key, _ := result.Key.Encode()
			value, _ := result.Value.Encode()
			logger.L().Info("Push kafka message successfully",
				zap.String("topic", result.Topic),
				zap.Int32("partition", result.Partition),
				zap.Int64("offset", result.Offset),
				zap.String("key", string(key)),
				zap.String("value", string(value)),
				zap.String("trace_id", spanCtx.TraceID().String()),
			)
		}
	}()
}

// Close https://github.com/Shopify/sarama/blob/main/examples/http_server/http_server.go#L91
func (p *kafkaProducerImpl) Close() {
	logger.L().Info("Closing kafka producer")
	if err := p.producer.Close(); err != nil {
		logger.L().Error("Failed to shut down access log producer cleanly", zap.Error(err))
	}
}

// Produce https://github.com/Shopify/sarama/blob/main/examples/http_server/http_server.go#L181
func (p *kafkaProducerImpl) Produce(topic string, key string, value []byte) {
	p.producer.Input() <- &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.StringEncoder(value),
		Timestamp: time.Now(),
	}
}

func (p *kafkaProducerImpl) ProduceCtx(ctx context.Context, topic string, key string, value []byte) {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.StringEncoder(value),
		Timestamp: time.Now(),
	}
	//start a span and inject header
	wrapperMsg := otelsarama.NewProducerMessageCarrier(msg)
	nCtx, span := otel.Tracer("kafka:"+topic).Start(ctx, "produce:kafka:"+topic)
	defer span.End()
	otel.GetTextMapPropagator().Inject(nCtx, wrapperMsg)

	p.producer.Input() <- msg
}
