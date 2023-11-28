.PHONY: run stop-minio stop-bot

run:
	@minio server start & 
	@go run cmd/bot/main.go

test:
	go test ./...
