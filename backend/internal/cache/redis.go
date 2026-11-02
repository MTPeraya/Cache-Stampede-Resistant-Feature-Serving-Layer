package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/MTPeraya/stampede-shield/backend/internal/chaos"
	"github.com/redis/go-redis/v9"
)

var ErrMiss = errors.New("cache miss")

type Client struct {
	redis *redis.Client
	chaos *chaos.State
}

func New(redisClient *redis.Client, states ...*chaos.State) *Client {
	state := chaos.New()
	if len(states) > 0 && states[0] != nil {
		state = states[0]
	}
	return &Client{redis: redisClient, chaos: state}
}

type Entry struct {
	Data        json.RawMessage `json:"data"`
	SoftExpires time.Time       `json:"soft_expires"`
}

func (c *Client) Get(ctx context.Context, key string) (Entry, error) {
	if c.chaos.RedisDown() {
		return Entry{}, errors.New("redis outage injected")
	}
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
	if c.chaos.RedisDown() {
		return errors.New("redis outage injected")
	}
	value, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, value, hardTTL).Err()
}

func (c *Client) Flush(ctx context.Context) error {
	if c.chaos.RedisDown() {
		return errors.New("redis outage injected")
	}
	var cursor uint64
	for {
		keys, next, err := c.redis.Scan(ctx, cursor, "feature:*", 1000).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.redis.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}
