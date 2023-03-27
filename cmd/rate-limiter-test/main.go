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
	log.Printf("result from test is %t", v)

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

    // ok, err :=  bucket.CheckRequestRateLimiter(ctx, client, "sam")
    // if err != nil {
    //     log.Panicf("error request rate limiter: %v\n", err)
    //     return false 
    // } 
    // return ok

    var isSuccess bool
	// send requests with 1sec interval and send each inside our reqData slice
	for i, item := range reqData {
		// send reqeuest (call rate limiter function)
		for key, val := range item {
			// create channel for each user
			success := make([]chan bool, 3)
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
				fmt.Printf("Second: %d, User: %s, Total: %d, Success: %d, Failed: %d\n", second+1, key, val, success, fail)
				if val > 20 && success == 20 {
					s <- true
				} else {
					s <- false
				}

			}(key, val, i, success[i])
			for _, r := range success {
                isSuccess = <-r
                fmt.Println(isSuccess)
			}
		}
		// sleep for 1 sec
		time.Sleep(1 * time.Second)
	}
	time.Sleep(4 * time.Second)
	return isSuccess
}

// func sendReq(ctx context.Context, client *redis.Client, bucket tokenbucket.RateLimiter, name string, req chan int) {
//         bucket.CheckRequestRateLimiter(ctx, client, name)
// }
