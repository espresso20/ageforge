.PHONY: build run clean check test vet validate all release

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

# Validate config keys (fast — catches typos in resource/building/tech/age keys)
validate:
	@echo "\033[0;36m[ageforge]\033[0m Validating config keys..."
	@go test ./config/ -count=1 -run TestConfig 2>&1 | \
		if grep -q "^ok"; then \
			echo "\033[0;32m[  OK  ]\033[0m Config keys valid"; \
		else \
			go test ./config/ -v -count=1 -run TestConfig 2>&1 | grep -A20 "FAIL\|Bad\|Orphaned\|Duplicate"; \
			echo "\033[0;31m[ FAIL ]\033[0m Config validation failed — see errors above"; \
			exit 1; \
		fi

# Build + vet + validate config (no run)
check: build vet validate
	@echo "\033[0;32m[  OK  ]\033[0m All checks passed"

# Run tests with formatted output (delegates to dev.sh)
test:
	@bash dev.sh test

# Run tests (raw go test output, for CI or piping)
test-raw: build vet
	@go test ./... -v -count=1

# Run the game
run: build
	@./ageforge

# Clean build artifacts
clean:
	@rm -f ageforge
	@rm -rf bin
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
