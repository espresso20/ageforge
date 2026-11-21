.PHONY: banner build run clean check vet steam all release release-patch release-minor release-major commit

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

# The Steam builds, into bin/steam/. The tag turns Steamworks on and the
# updater and font download off. The Windows flags stop the Go runtime
# freezing the game under Steam's overlay and hide the console window.
# See CONTRIBUTING.md, "Steam build".
STEAM_VERSION ?= dev
steam:
	@mkdir -p bin/steam
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags steam -ldflags "-X main.version=$(STEAM_VERSION) -X=runtime.godebugDefault=asyncpreemptoff=1 -H=windowsgui" -o bin/steam/ageforge-windows-amd64.exe .
	GOOS=darwin GOARCH=arm64 go build -tags steam -ldflags "-X main.version=$(STEAM_VERSION)" -o bin/steam/ageforge-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -tags steam -ldflags "-X main.version=$(STEAM_VERSION)" -o bin/steam/ageforge-linux-amd64 .

# Run the game
run: build
	@./ageforge

# Generate the README banner PNG and Discord icon
banner:
	@echo "\033[0;36m[ageforge]\033[0m Generating banner and icon PNGs..."
	@go run ./cmd/bannerpng
	@go run ./cmd/bannerpng -icon-af
	@echo "\033[0;32m[  OK  ]\033[0m Banner and icon generated to docs/"

# Clean build artifacts
clean:
	@rm -f ageforge
	@rm -rf bin
	@echo "\033[0;32m[  OK  ]\033[0m Cleaned"

# Build for multiple platforms (local, no version injection)
release:
	@echo "\033[0;36m[ageforge]\033[0m Building releases..."
	@mkdir -p bin
	@GOOS=darwin  GOARCH=amd64 go build -o ./bin/ageforge-darwin-amd64
	@GOOS=darwin  GOARCH=arm64 go build -o ./bin/ageforge-darwin-arm64
	@GOOS=linux   GOARCH=amd64 go build -o ./bin/ageforge-linux-amd64
	@GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o ./bin/ageforge-windows-amd64.exe
	@echo "\033[0;32m[  OK  ]\033[0m Release builds complete"

# Bump version, update CHANGELOG, commit, tag, and push to trigger CI release
release-patch:
	@bash scripts/release.sh patch

release-minor:
	@bash scripts/release.sh minor

release-major:
	@bash scripts/release.sh major

# Interactive conventional commit — stages everything, prompts for type + description, commits + pushes
commit:
	@bash scripts/commit.sh
