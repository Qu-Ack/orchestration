build:
	@go build -o bin/o ./cmd/
	@sudo ./bin/o
run:
	@go run main.go
