package tokenbucket

import (
	"context"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v9"
)

type tokenBucket struct {
	replenishRate int
	capacity      int
	keyPrefix     string
	script        *redis.Script
}

type RateLimiter interface {
	CheckRequestRateLimiter(context.Context, *redis.Client, string) (bool, error)
}

func NewTokenBucket(replenishRate int, capacity int, keyPrefix string, scriptFilePath string) (RateLimiter, error) {
	fileName := "token_bucket.lua"
	fsys := os.DirFS(scriptFilePath)
	scriptByte, err := fs.ReadFile(fsys, fileName)
    script := redis.NewScript(string(scriptByte))
	if err != nil {
		log.Panicf("Unable to read file %q from %q", fileName, scriptFilePath)
		return nil, err
	}

	return &tokenBucket{
		replenishRate: replenishRate,
		capacity:      capacity,
		keyPrefix:     keyPrefix,
		script:        script,
	}, nil
}

func (t *tokenBucket) CheckRequestRateLimiter(ctx context.Context, rdb *redis.Client, user string) (bool, error) {
	prefix := t.keyPrefix + user
    keys := []string{prefix + ":token", prefix + ":timestamp"}

    // log.Printf("checking request limiter for user %q", user)
    r := t.script.Run(ctx, rdb, keys, t.replenishRate, t.capacity, time.Now().Unix(), 1)
    err := r.Err()
    if err != nil {
        // log.Printf("request failed for user %q", user)
        return false, err
    }
    // log.Printf("request passed for user %q", user)
    ret := r.Val()
    returnValue := ret.([]interface{})
    if returnValue[0] == int64(1) {
        return true, nil
    }
    return false, nil
}

// How many tokens to be added per second ?
// How many requests per second do you want a user to be allowed to do ?
const REPLENISH_RATE = 10

// How much throttling to be allowed ?
const CAPACITY = 2 * REPLENISH_RATE

const KEY_PREFIX = `request_rate_limiter:`

func CheckRequestRateLimiter(ctx context.Context, rdb redis.Client, user string) (bool, error) {

	// read the script
	// @TODO reading file each time is inefficient, create a struct with config and init when the app starts
	fsys := os.DirFS("./lua")
	scriptFile, err := fs.ReadFile(fsys, "token_bucket.lua")
	if err != nil {
		return false, err
	}

	prefix := KEY_PREFIX + user
	keys := []string{prefix + ":token", prefix + ":timestamp"}

	scriptStr := string(scriptFile)
	script := redis.NewScript(scriptStr)
	return script.Run(ctx, rdb, keys, REPLENISH_RATE, CAPACITY, time.Now().Unix(), 1).Bool()
}
