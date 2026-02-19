package kafka

import "time"

type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Service   string    `json:"service"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"createdAt"`
}
