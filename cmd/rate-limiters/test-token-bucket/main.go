package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v9"
	docker_client "github.com/samthom/system-designing/pkg/docker"
    redis_container "github.com/samthom/system-designing/lib/redis"
	tokenbucket "github.com/samthom/system-designing/pkg/rate-limiter/token-bucket"
)

func main() {
    ctx := createAContext()
	dockerClient, err := docker_client.NewDockerClient()
	if err != nil {
		log.Panicf("NewDockerClient() failed: %v", err)
	}

    redisContainer, _ := setupRedisContainer(ctx, dockerClient)
    redisClient, err := redis_container.RedisConnect(ctx)
    if err != nil {
        log.Fatal(err)
    }
	v := testTokenBucket(ctx, redisClient)
    log.Printf("test result: %t", v)

    stopped := redisContainer.StopRedis(ctx)
	if stopped {
		log.Printf("redis container stopped successfully.")
		return
	} else {
		log.Fatal("stopRedis() failed.")
	}
}

func createAContext() context.Context {
    ctx := context.TODO()
	// ctx, _ = context.WithCancel(ctx)
    return ctx
}

func setupRedisContainer(ctx context.Context, dockerClient docker_client.DockerClient) (redis_container.RedisContainer, error) {
    redisContainer := redis_container.GetRedis(ctx, dockerClient)
    containerID, err := redisContainer.StartRedis(ctx)
	if err == nil {
		log.Printf("redis container started successfully. container id: %q", containerID)
        return redisContainer, nil
	} else {
		log.Fatalf("startRedis() failed. %v", err)
        return redisContainer, err
	}
}

func testTokenBucket(ctx context.Context, redisClient redis.Cmdable) bool {
	// create new client
	reqData := [3]map[string]int{}
	reqData[0] = map[string]int{
		"achu": 25,
		"bob":  25,
		"tom":  20,
	}
	reqData[1] = map[string]int{
		"achu": 20,
		"bob":  21,
		"tom":  26,
	}
	reqData[2] = map[string]int{
		"achu": 15,
		"bob":  18,
		"tom":  30,
	}

	bucket, err := tokenbucket.NewTokenBucket(10, 20, "token-bucket:", "./pkg/rate-limiter/lua")
	if err != nil {
		log.Fatalf("Unable create NewTokenBucket instance: %v", err)
	}

	var success bool
	// send requests with 1sec interval and send each inside our reqData slice
	for i, item := range reqData {
		// send reqeuest (call rate limiter function)
		for key, val := range item {
			result := make(chan bool)
			go func(key string, val int, second int, s chan bool) {
				var (
					success int = 0
					fail    int = 0
				)
				for i := 0; i < val; i++ {
					// go sendReq(ctx, client, bucket, key)
					ok, _ := bucket.CheckRequestRateLimiter(ctx, redisClient, key)
					if ok {
						success++
					} else {
						fail++
					}
				}
				log.Printf("Second: %d, User: %s, Total: %d, Success: %d, Failed: %d\n", second+1, key, val, success, fail)
				if success <= 20 {
					s <- true
				} else {
					s <- false
				}

			}(key, val, i, result)
			success = <-result
            success = success && true
		}
		// sleep for 1 sec
		time.Sleep(1 * time.Second)
	}
	return success
}
