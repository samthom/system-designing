package main

import (
	"context"
	"log"
	"time"

	rediscontainer "github.com/samthom/system-designing/lib/redis"
	docker_client "github.com/samthom/system-designing/pkg/docker"
	ratelimiter "github.com/samthom/system-designing/lib/rate-limiter"
)

func main() {
	ctx := createAContext()
	dockerClient, err := docker_client.NewDockerClient()
	if err != nil {
		log.Panicf("NewDockerClient() failed: %v", err)
	}

	redisContainer, _ := rediscontainer.SetupRedisContainer(ctx, dockerClient)
    redisClient, err := rediscontainer.NewRedis(ctx)
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

var reqData = [3]map[string]int{
	{
		"achu": 25,
		"bob":  25,
		"tom":  20,
	},
	{

		"achu": 20,
		"bob":  21,
		"tom":  26,
	},
	{
		"achu": 15,
		"bob":  18,
		"tom":  30,
	},
}

func testTokenBucket(ctx context.Context, redisClient rediscontainer.RedisClient) bool {
    script, err := redisClient.NewScript(ratelimiter.TOKENBUCKET, "./lib/rate-limiter")
    if err != nil {
        log.Fatalf("Unable to create script: %v", err)
    }
	bucket := ratelimiter.NewTokenBucket(10, 20, "token-bucket:", script)

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
					ok, _ := bucket.CheckRequestRateLimiter(ctx, key)
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
		time.Sleep(1 * time.Second)
	}
	return success
}
