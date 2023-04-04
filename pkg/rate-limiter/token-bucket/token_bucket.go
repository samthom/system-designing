package tokenbucket

import (
	"context"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v9"
)

const fileName = "token_bucket.lua"

type tokenBucket struct {
	replenishRate int
	capacity      int
	keyPrefix     string
	script        *redis.Script
}

type RateLimiter interface {
	CheckRequestRateLimiter(context.Context, redis.Cmdable, string) (bool, error)
}

func NewTokenBucket(replenishRate int, capacity int, keyPrefix string, scriptFilePath string) (RateLimiter, error) {

    scriptStr, err := readLuaScript(scriptFilePath)
	if err != nil {
		log.Panicf("Unable to read file %q from %q", fileName, scriptFilePath)
		return nil, err
	}

    script := redis.NewScript(scriptStr)
	return &tokenBucket{
		replenishRate: replenishRate,
		capacity:      capacity,
		keyPrefix:     keyPrefix,
		script:        script,
	}, nil
}

func (t *tokenBucket) CheckRequestRateLimiter(ctx context.Context, rdb redis.Cmdable, user string) (bool, error) {
	prefix := t.keyPrefix + user
    keys := []string{prefix + ":token", prefix + ":timestamp"}

    r := t.script.Run(ctx, rdb, keys, t.replenishRate, t.capacity, time.Now().Unix(), 1)
    err := r.Err()
    if err != nil {
        return false, err
    }
    ret := r.Val()
    returnValue := ret.([]interface{})
    if returnValue[0] == int64(1) {
        return true, nil
    }
    return false, nil
}

func readLuaScript(scriptFilePath string) (string, error) {
	fsys := os.DirFS(scriptFilePath)
    b, err := fs.ReadFile(fsys, fileName)
    if err != nil {
        return "", err
    }
    
    return string(b), nil
}
