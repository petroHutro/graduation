.PHONY: run stop-minio stop-bot

run:
	@minio server start & echo $$! > minio_pid.txt
	@go run cmd/gophermart/main.go echo $$! > go_pid.txt

test:
	go test ./...

stop:
	@-kill cat minio_pid.txt
	@-rm minio_pid.txt

	@-kill catgo_pid.txt
	@-rm go_pid.txt