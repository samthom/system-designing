package rediscontainer

import (
	"context"
	"errors"
	"log"

	"github.com/go-redis/redis/v9"
	"github.com/samthom/system-designing/pkg"
)

func NewRedis(ctx context.Context) (RedisClient, error) {
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
    return &redisClient{client}, nil
}

type RedisClient interface {
    NewScript(fileName, scriptFilePath string) (Script, error)
    NewSet(key string) Set
}

type redisClient struct {
    redis.Cmdable
} 

func (r *redisClient) NewScript(fileName, scriptFilePath string) (Script, error) {
	scriptStr, err := pkg.ReadFileToString(scriptFilePath, fileName)
	if err != nil {
		log.Panicf("Unable to read file %q from %q", fileName, scriptFilePath)
		return nil, err
	}

	s := redis.NewScript(scriptStr)
    return &script{s,r},nil
}

type Script interface {
	Run(ctx context.Context, keys []string, args ...interface{}) (bool, error)
}

type script struct {
	script *redis.Script
    client redis.Cmdable
}

func (s *script) Run(ctx context.Context, keys []string, args ...interface{}) (bool, error) {
    return s.script.Run(ctx, s.client, keys, args...).Bool()
}

func (r *redisClient) NewSet(key string) Set {
    return &set{
        key: key,
        client: r.Cmdable,
    }
}

type Set interface {
    Add(ctx context.Context, val ...interface{}) (int64, error)
    Remove(ctx context.Context, val ...interface{}) (int64, error)
}

type set struct {
    key string
    client redis.Cmdable
}

func (s *set) Add(ctx context.Context, val ...interface{}) (int64, error) {
    return s.client.SAdd(ctx, s.key, val...).Result()
}

func (s *set) Remove(ctx context.Context, val ...interface{}) (int64, error) {
    return s.client.SRem(ctx, s.key, val...).Result()
}
