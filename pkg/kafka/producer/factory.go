package producer

import (
	"context"

	"go.uber.org/fx"
)

func NewKafkaTemplate(lifecycle fx.Lifecycle, brokers []string) KafkaProducer {
	p := NewKafkaProducer(brokers)
	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.Close()
			return nil
		}},
	)
	return p
}
