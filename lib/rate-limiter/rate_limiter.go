package ratelimiter

import "context"

type RateLimiter interface {
	CheckRequestRateLimiter(ctx context.Context, id string) (ok bool, err error)
}
