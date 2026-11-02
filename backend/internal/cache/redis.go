package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrMiss = errors.New("cache miss")

type Client struct {
	redis *redis.Client
}

func New(redisClient *redis.Client) *Client {
	return &Client{redis: redisClient}
}

type Entry struct {
	Data        json.RawMessage `json:"data"`
	SoftExpires time.Time       `json:"soft_expires"`
}

func (c *Client) Get(ctx context.Context, key string) (Entry, error) {
	value, err := c.redis.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return Entry{}, ErrMiss
	}
	if err != nil {
		return Entry{}, err
	}
	var entry Entry
	if err := json.Unmarshal(value, &entry); err != nil {
		return Entry{}, err
	}
	return entry, nil
}

func (c *Client) Set(ctx context.Context, key string, entry Entry, softTTL, hardTTL time.Duration) error {
	value, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, value, hardTTL).Err()
}
