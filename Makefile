.PHONY: build test lint fmt docker-build

build:
	go build -o slackliner .

test:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run

fmt:
	go fmt ./...

docker-build:
	docker build -t slackliner .
