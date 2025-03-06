.PHONY: build clean test run

# Build variables
BINARY_NAME=snagbot
BUILD_DIR=bin

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run:
	@go run ./cmd/server

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

lint:
	@echo "Linting..."
	@golangci-lint run

# Build and run the application
dev: build run
