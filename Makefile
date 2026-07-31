.PHONY: all build test race bench lint clean proto docker

BINARY_NAME=prefixos
BUILD_DIR=bin

all: build

proto:
	@echo "Generating Protobuf stubs..."
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/v1/prefixos.proto

build: proto
	@echo "Building PrefixOS server binary..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/server/main.go

test:
	@echo "Running unit tests..."
	go test -v ./...

race:
	@echo "Running tests with race detector..."
	go test -v -race ./...

bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./benchmarks/...

lint:
	@echo "Running linter..."
	golangci-lint run ./... || true

docker:
	docker build -t prefixos:latest .

clean:
	rm -rf $(BUILD_DIR)
