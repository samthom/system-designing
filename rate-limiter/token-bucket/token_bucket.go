package tokenbucket

import (
	"context"
	"io/fs"
	"os"
	"time"

	"github.com/go-redis/redis/v9"
)

// How many tokens to be added per second ?
// How many requests per second do you want a user to be allowed to do ?
const REPLENISH_RATE = 10

// How much throttling to be allowed ?
const CAPACITY = 2 * REPLENISH_RATE

const KEY_PREFIX = `request_rate_limiter:`

func CheckRequestRateLimiter(ctx context.Context,rdb redis.Client, user string) (bool, error) {

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
