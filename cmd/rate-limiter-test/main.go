package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v9"
	// docker_client "github.com/samthom/system-designing/docker"
	"github.com/samthom/system-designing/rate-limiter/token-bucket"
)

func main() {
	ctx := context.TODO()
	// ctx, cancel := context.WithCancel(ctx)
	// client, err := docker_client.NewDockerClient()
	// if err != nil {
	// 	log.Panicf("NewDockerClient() failed: %v", err)
	// }
	// started, containerID := startRedis(ctx, client)
	// if started {
	// 	log.Printf("redis container started successfully.\ncontainer id: %q", containerID)
	// } else {
	// 	log.Fatalf("startRedis() failed.")
	// }

	v := testTokenBucket(ctx)
    log.Printf("test result: %t", v)

	// stopped := stopRedis(ctx, client, containerID)
	// if stopped {
	// 	log.Printf("redis container stopped successfully.")
	// 	return
	// } else {
	// 	log.Fatalf("stopRedis(%q) failed.", containerID)
	// }
}

// get a redis client to the docker continer (redis)
func redisConnect() *redis.Client {
	redisOpts := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	client := redis.NewClient(redisOpts)
	return client
}

func testTokenBucket(ctx context.Context) bool {
	// create new client
	client := redisConnect()
	pop := client.Ping(ctx)
	fmt.Printf("Ping: %s\n", pop.Val())
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

	bucket, err := tokenbucket.NewTokenBucket(10, 20, "token-bucket:", "./rate-limiter/lua")
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
					ok, _ := bucket.CheckRequestRateLimiter(ctx, client, key)
					if ok {
						success++
					} else {
						fail++
					}
				}
				log.Printf("Second: %d, User: %s, Total: %d, Success: %d, Failed: %d\n", second+1, key, val, success, fail)
				if success == 20 {
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
