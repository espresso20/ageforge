.PHONY: build run clean all

# Default target
all: build

# Build the game
build:
	@echo "Building AgeForge..."
	@go build -o ageforge

# Run the game
run: build
	@echo "Running AgeForge..."
	@./ageforge

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f ageforge

# Build for multiple platforms
release:
	@echo "Building releases for multiple platforms..."
	@GOOS=darwin GOARCH=amd64 go build -o ./bin/ageforge-darwin-amd64
	@GOOS=darwin GOARCH=arm64 go build -o ./bin/ageforge-darwin-arm64
	@GOOS=linux GOARCH=amd64 go build -o ./bin/ageforge-linux-amd64
	@GOOS=windows GOARCH=amd64 go build -o ./bin/ageforge-windows-amd64.exe
	@echo "Release builds complete."
