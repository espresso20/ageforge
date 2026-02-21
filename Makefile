.PHONY: build run clean check test all release

# Default: build + vet + run
all: check run

# Build the binary
build:
	@echo "\033[0;36m[ageforge]\033[0m Building..."
	@go build -o ageforge .
	@echo "\033[0;32m[  OK  ]\033[0m Build succeeded"

# Run go vet
vet:
	@echo "\033[0;36m[ageforge]\033[0m Running go vet..."
	@go vet ./...
	@echo "\033[0;32m[  OK  ]\033[0m go vet passed"

# Build + vet (no run)
check: build vet
	@echo "\033[0;32m[  OK  ]\033[0m All checks passed"

# Run tests
test: build vet
	@echo "\033[0;36m[ageforge]\033[0m Running tests..."
	@go test ./... -v
	@echo "\033[0;32m[  OK  ]\033[0m Tests done"

# Run the game
run: build
	@./ageforge

# Clean build artifacts
clean:
	@rm -f ageforge
	@echo "\033[0;32m[  OK  ]\033[0m Cleaned"

# Build for multiple platforms
release:
	@echo "\033[0;36m[ageforge]\033[0m Building releases..."
	@mkdir -p bin
	@GOOS=darwin GOARCH=amd64 go build -o ./bin/ageforge-darwin-amd64
	@GOOS=darwin GOARCH=arm64 go build -o ./bin/ageforge-darwin-arm64
	@GOOS=linux GOARCH=amd64 go build -o ./bin/ageforge-linux-amd64
	@GOOS=windows GOARCH=amd64 go build -o ./bin/ageforge-windows-amd64.exe
	@echo "\033[0;32m[  OK  ]\033[0m Release builds complete"
