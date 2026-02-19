package kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(cfg ProducerConfig, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    topic,
		Balancer: getBalancer(cfg.Balancer),
	}

	if cfg.DeliverySemantic == AtMostOnce {
		writer.RequiredAcks = kafka.RequireOne // лидер подтвердил
		writer.Async = true                    // можно писать быстро
	} else if cfg.DeliverySemantic == AtLeastOnce {
		writer.RequiredAcks = kafka.RequireAll // все реплики подтвердили
		writer.Async = false                   // синхронная запись, чтобы ждать ack
	}

	return &Producer{writer: writer}
}

func (p *Producer) Send(ctx context.Context, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.ID),
		Value: data,
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
