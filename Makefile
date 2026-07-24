.PHONY: run test build
run:
	go run .
test:
	go test -race -cover ./...
	go vet ./...
build:
	go build -o bin/receitas-api .
