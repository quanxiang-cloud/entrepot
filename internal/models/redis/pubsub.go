package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/quanxiang-cloud/entrepot/internal/models"
)

type pubSub struct {
	client *redis.ClusterClient
}

// NewPubSub NewPubSub
func NewPubSub(client *redis.ClusterClient) models.PubSubRepo {
	return &pubSub{
		client: client,
	}
}

func (ps *pubSub) Key(topic, key string) string {
	return fmt.Sprintf("entrepot:subscribe:%s:%s", topic, key)
}

func (ps *pubSub) Subscribe(ctx context.Context, pubSub *models.PubSub, ttl time.Duration) error {
	err := ps.client.HSet(
		ctx,
		ps.Key(pubSub.Topic, pubSub.Key), pubSub.UserID, pubSub.UUID,
	).Err()

	if err != nil {
		return err
	}

	return ps.client.PExpire(ctx, ps.Key(pubSub.Topic, pubSub.Key), ttl).Err()
}

func (ps *pubSub) Get(ctx context.Context, topic, key string) ([]*models.PubSub, error) {
	res := ps.client.HGetAll(
		ctx,
		ps.Key(topic, key),
	)
	if res.Err() == redis.Nil {
		return nil, nil
	}
	if res.Err() != nil {
		return nil, res.Err()
	}

	tmp, err := res.Result()
	if err != nil {
		return nil, err
	}

	ans := make([]*models.PubSub, 0, len(tmp))
	for userID, uuid := range tmp {
		ans = append(ans, &models.PubSub{
			UserID: userID,
			UUID:   uuid,
			Topic:  topic,
			Key:    key,
		})
	}

	return ans, nil
}
