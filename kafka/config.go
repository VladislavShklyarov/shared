package kafka

import "github.com/segmentio/kafka-go"

type BalancerType string

const (
	BalancerRoundRobin BalancerType = "round_robin"
	BalancerLeastBytes BalancerType = "least_bytes"
	BalancerHash       BalancerType = "hash"
)

type DeliverySemantic string

const (
	AtMostOnce  DeliverySemantic = "at_most_once"
	AtLeastOnce DeliverySemantic = "at_least_once"
)

type ProducerConfig struct {
	Brokers          []string
	Topic            string
	Balancer         BalancerType
	Acks             int
	DeliverySemantic DeliverySemantic // at_least_once / at_most_once / exactly_once
	Retries          int
}

type ConsumerConfig struct {
	Brokers          []string
	Topic            string
	GroupID          string
	DeliverySemantic DeliverySemantic
	MinBytes         int
	MaxBytes         int
}

func getBalancer(t BalancerType) kafka.Balancer {
	switch t {
	case BalancerHash:
		return &kafka.Hash{}
	case BalancerLeastBytes:
		return &kafka.LeastBytes{}
	case BalancerRoundRobin:
		return &kafka.RoundRobin{}
	default:
		return &kafka.LeastBytes{}
	}
}
