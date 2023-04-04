.PHONY: rate-limiter
rate-limiter:
	go run cmd/rate-limiters/test-token-bucket/main.go
