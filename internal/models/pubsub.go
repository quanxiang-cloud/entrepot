package models

import (
	"context"
	"time"
)

// PubSub PubSub
type PubSub struct {
	UserID string
	UUID   string
	Topic  string
	Key    string
}

//PubSubRepo PubSubRepo
type PubSubRepo interface {
	Subscribe(ctx context.Context, pubsub *PubSub, ttl time.Duration) error

	Get(ctx context.Context, topic, key string) ([]*PubSub, error)
}
