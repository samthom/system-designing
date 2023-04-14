package main

import (
	"context"
	"log"

	rediscontainer "github.com/samthom/system-designing/lib/redis"
	docker_client "github.com/samthom/system-designing/pkg/docker"
	// ratelimiter "github.com/samthom/system-designing/lib/rate-limiter"
)

func main() {
	// ctx := createAContext()
	// dockerClient, err := docker_client.NewDockerClient()
	// if err != nil {
	// 	log.Panicf("NewDockerClient() failed: %v", err)
	// }
	//
	// redisContainer, _ := setupRedisContainer(ctx, dockerClient)
 //    redisClient, err := rediscontainer.NewRedis(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// v := testLeakyBucket(ctx, redisClient)
	// log.Printf("test result: %t", v)
	//
	// stopped := redisContainer.StopRedis(ctx)
	// if stopped {
	// 	log.Printf("redis container stopped successfully.")
	// 	return
	// } else {
	// 	log.Fatal("stopRedis() failed.")
	// }
}

func createAContext() context.Context {
	ctx := context.TODO()
	// ctx, _ = context.WithCancel(ctx)
	return ctx
}

func setupRedisContainer(ctx context.Context, dockerClient docker_client.DockerClient) (rediscontainer.RedisContainer, error) {
	redisContainer := rediscontainer.GetRedis(ctx, dockerClient)
	containerID, err := redisContainer.StartRedis(ctx)
	if err == nil {
		log.Printf("redis container started successfully. container id: %q", containerID)
		return redisContainer, nil
	} else {
		log.Fatalf("startRedis() failed. %v", err)
		return redisContainer, err
	}
}

func testLeakyBucket(ctx context.Context, rdb rediscontainer.RedisClient) bool {
    return false
}

func testCallback() {
    // data := []map[int] struct {
    //     time int
    //     timeout int
    // }
}
