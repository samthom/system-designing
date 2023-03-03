package main

import (
	"context"
	"log"
	"time"

	docker_client "github.com/samthom/system-designing/docker"
)

func main() {
    ctx := context.TODO()
    // ctx, cancel := context.WithCancel(ctx)
	client, err := docker_client.NewDockerClient()
	if err != nil {
		log.Panicf("NewDockerClient() failed: %v", err)
	}
    started, containerID := startRedis(ctx, client)
    if started {
        log.Printf("redis container started successfully.\ncontainer id: %q", containerID)
    } else {
        log.Fatalf("startRedis() failed.")
    }

    v := testTokenBucket(ctx)
    log.Printf("result from test is %t", v)

    stopped := stopRedis(ctx, client, containerID)
    if stopped {
        log.Printf("redis container stopped successfully.")
        return
    } else {
        log.Fatalf("stopRedis(%q) failed.", containerID)
    }
}

func testTokenBucket(ctx context.Context) bool {
    time.Sleep(10 * time.Second)
    return true
}
