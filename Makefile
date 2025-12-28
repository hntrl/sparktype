.PHONY: build install test clean run

# Build the sparktype binary
build:
	go build -o bin/sparktype ./cmd/sparktype

# Install the binary to $GOPATH/bin
install:
	go install ./cmd/sparktype

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf dist/

# Run the CLI
run:
	go run ./cmd/sparktype $(ARGS)

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	go vet ./...

# Download dependencies
deps:
	go mod tidy
	go mod download

