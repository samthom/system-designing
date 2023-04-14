package ratelimiter

import (
	"context"

	rediscontainer "github.com/samthom/system-designing/lib/redis"
)

const LEAKYBUCKET = "leaky_bucket.lua"

type leakyBucket struct {
	capacity  int
	keyPrefix string
	script    rediscontainer.Script
}

func NewLeakyBucket(capacity int, keyPrefix string, script rediscontainer.Script) RateLimiter {
    return &leakyBucket{
        capacity: capacity,
        keyPrefix: keyPrefix,
        script: script,
    } 
}

func (t *leakyBucket) CheckRequestRateLimiter(ctx context.Context, id string) (bool, error) {
	keys := []string{t.keyPrefix}

	return t.script.Run(ctx, keys, id)
}
