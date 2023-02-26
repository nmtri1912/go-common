package consumer

import (
	"context"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/nmtri1912/go-common/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type KafkaConsumer struct {
	Brokers     []string
	GroupId     string
	Topics      []string
	ready       chan bool
	Handler     func(message *sarama.ConsumerMessage)
	client      sarama.ConsumerGroup
	quitConsume chan bool
	quitWorker  chan bool
	waitConsume *sync.WaitGroup
	waitWorker  *sync.WaitGroup
	messages    chan *sarama.ConsumerMessage
	numWorker   int
}

func NewKafkaConsumer(brokers []string, groupId string, topics []string, handler func(message *sarama.ConsumerMessage), numWorker int) *KafkaConsumer {
	if numWorker < 1 {
		logger.L().Fatal("Number of workers invalid")
	}

	return &KafkaConsumer{
		Brokers:     brokers,
		GroupId:     groupId,
		Topics:      topics,
		Handler:     handler,
		ready:       make(chan bool),
		quitConsume: make(chan bool),
		quitWorker:  make(chan bool),
		waitConsume: &sync.WaitGroup{},
		waitWorker:  &sync.WaitGroup{},
		messages:    make(chan *sarama.ConsumerMessage, 8*numWorker),
		numWorker:   numWorker,
	}
}

func (c *KafkaConsumer) Start() {
	logger.L().Info("Starting Sarama consumer")

	// https://github.com/Shopify/sarama/blob/main/examples/consumergroup/main.go#L68
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Metadata.Timeout = time.Second * 3

	ctx, cancel := context.WithCancel(context.Background())
	var err error
	c.client, err = sarama.NewConsumerGroup(c.Brokers, c.GroupId, config)
	if err != nil {
		logger.L().Fatal("Error creating consumer group client", zap.Error(err))
	}

	logger.L().Info("Start worker", zap.Int("numbers", c.numWorker))
	for i := 0; i < c.numWorker; i++ {
		c.waitWorker.Add(1)
		go func() {
			defer func() {
				c.waitWorker.Done()
			}()

			for {
				select {
				case msg, ok := <-c.messages:
					if ok {
						c.Handler(msg)
					}
				case <-c.quitWorker:
					if len(c.messages) < 1 {
						return
					}
				}
			}
		}()
	}

	logger.L().Info("Consumer start consume")
	c.waitConsume.Add(1)
	go func() {
		defer c.waitConsume.Done()
		for {
			if err := c.client.Consume(ctx, c.Topics, c); err != nil {
				logger.L().Error("Error from consumer", zap.Error(err))
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			c.ready = make(chan bool)
		}
	}()

	<-c.ready // Await till the consumer has been set up

	select {
	case <-ctx.Done():
	case <-c.quitConsume:
	}
	cancel()
}

func (c *KafkaConsumer) Close() {
	logger.L().Info("Stopping consumer")
	close(c.quitConsume)
	c.waitConsume.Wait()
	if err := c.client.Close(); err != nil {
		logger.L().Error("Error closing client", zap.Error(err))
	}
	close(c.quitWorker)
	c.waitWorker.Wait()
	close(c.messages)
	logger.L().Info("All workers have exited")
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			wMsg := otelsarama.NewConsumerMessageCarrier(message)
			nCtx := otel.GetTextMapPropagator().Extract(context.Background(), wMsg)
			spanCtx := trace.SpanContextFromContext(nCtx)

			logger.L().Info("Receive kafka message",
				zap.String("topic", message.Topic),
				zap.Int32("partition", message.Partition),
				zap.Int64("offset", message.Offset),
				zap.String("key", string(message.Key)),
				zap.String("value", string(message.Value)),
				zap.String("trace_id", string(spanCtx.TraceID().String())),
			)
			c.messages <- message
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func ExtractSpanAndTracingContext(message *sarama.ConsumerMessage) (context.Context, trace.Span) {
	wMsg := otelsarama.NewConsumerMessageCarrier(message)
	nCtx := otel.GetTextMapPropagator().Extract(context.Background(), wMsg)
	nCtx, span := otel.Tracer("kafka:"+message.Topic).Start(nCtx, "consume:kafka:"+message.Topic)
	return nCtx, span
}
