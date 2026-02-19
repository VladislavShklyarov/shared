package kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"time"
)

type Consumer struct {
	reader           *kafka.Reader
	deliverySemantic DeliverySemantic
}

func NewConsumer(cfg ConsumerConfig) *Consumer {
	readerCfg := kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: cfg.MinBytes,
		MaxBytes: cfg.MaxBytes,
	}

	if cfg.DeliverySemantic == AtMostOnce {
		readerCfg.CommitInterval = time.Second // авто-коммит
	} else {
		readerCfg.CommitInterval = 0 // ручной commit
	}

	reader := kafka.NewReader(readerCfg)

	return &Consumer{
		reader:           reader,
		deliverySemantic: cfg.DeliverySemantic,
	}
}

func (c *Consumer) Read(ctx context.Context) (*Event, kafka.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return nil, kafka.Message{}, err
	}

	var event Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return nil, kafka.Message{}, err
	}

	// AtMostOnce → коммитим сразу, после чтения
	if c.deliverySemantic == AtMostOnce {
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return nil, kafka.Message{}, err
		}
	}

	return &event, msg, nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
