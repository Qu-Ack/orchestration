build:
	@go build -o bin/o ~/projects/orchestration/cmd/
	@sudo ./bin/o
run:
	@go run main.go
