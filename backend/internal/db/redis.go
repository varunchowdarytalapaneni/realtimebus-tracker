package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"bus-tracker/internal/domain"
)

const ChannelPrefix = "telemetry:bus:"

func Channel(busID int64) string   { return fmt.Sprintf("%s%d", ChannelPrefix, busID) }
func latestKey(busID int64) string { return fmt.Sprintf("bus:%d:latest", busID) }

// Atomic compare-and-set: only store the ping if it is newer than what's cached.
// Protects against late/out-of-order pings (e.g. an offline batch arriving).
var setIfNewer = redis.NewScript(`
local cur = redis.call('HGET', KEYS[1], 'ts')
if cur and tonumber(cur) >= tonumber(ARGV[1]) then return 0 end
redis.call('HSET', KEYS[1], 'ts', ARGV[1], 'data', ARGV[2])
redis.call('EXPIRE', KEYS[1], ARGV[3])
return 1
`)

type TelemetryCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewTelemetryCache(rdb *redis.Client, ttl time.Duration) *TelemetryCache {
	return &TelemetryCache{rdb: rdb, ttl: ttl}
}

func (c *TelemetryCache) SetLatestIfNewer(ctx context.Context, t domain.Telemetry) (bool, error) {
	payload, err := json.Marshal(t)
	if err != nil {
		return false, err
	}
	n, err := setIfNewer.Run(ctx, c.rdb, []string{latestKey(t.BusID)},
		t.Timestamp.UnixMilli(), payload, int(c.ttl.Seconds())).Int()
	return n == 1, err
}

func (c *TelemetryCache) GetLatest(ctx context.Context, busID int64) (*domain.Telemetry, error) {
	s, err := c.rdb.HGet(ctx, latestKey(busID), "data").Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var t domain.Telemetry
	if err := json.Unmarshal([]byte(s), &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *TelemetryCache) Publish(ctx context.Context, t domain.Telemetry) error {
	payload, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return c.rdb.Publish(ctx, Channel(t.BusID), payload).Err()
}

// PSubscribe subscribes ONCE to every bus channel (pattern subscribe).
// go-redis automatically reconnects and resubscribes.
func (c *TelemetryCache) PSubscribe(ctx context.Context) *redis.PubSub {
	return c.rdb.PSubscribe(ctx, ChannelPrefix+"*")
}
