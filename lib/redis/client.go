package redis_container

import (
	"context"
	"errors"
	"log"

	"github.com/go-redis/redis/v9"
)

// get a redis client to the docker continer (redis)
func RedisConnect(ctx context.Context) (redis.Cmdable, error) {
	redisOpts := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	client := redis.NewClient(redisOpts)
	pop := client.Ping(ctx)
	result := pop.Val()
	log.Printf("redis PING: %s\n", result)
	if len(result) == 0 {
		return nil, errors.New("unable to connect to the redis server.")
	}
	return client, nil
}
