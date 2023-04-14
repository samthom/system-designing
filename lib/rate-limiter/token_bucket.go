package ratelimiter

import (
	"context"
	"time"

	rediscontainer "github.com/samthom/system-designing/lib/redis"
)

const TOKENBUCKET = "token_bucket.lua"

type tokenBucket struct {
	replenishRate int
	capacity      int
	keyPrefix     string
	script        rediscontainer.Script
}

func NewTokenBucket(replenishRate int, capacity int, keyPrefix string, script rediscontainer.Script) RateLimiter {
	return &tokenBucket{
		replenishRate: replenishRate,
		capacity:      capacity,
		keyPrefix:     keyPrefix,
		script:        script,
	}
}

func (t *tokenBucket) CheckRequestRateLimiter(ctx context.Context, id string) (bool, error) {
	prefix := t.keyPrefix + id
	keys := []string{prefix + ":token", prefix + ":timestamp"}

	return t.script.Run(ctx, keys, t.replenishRate, t.capacity, time.Now().Unix(), 1)
}
