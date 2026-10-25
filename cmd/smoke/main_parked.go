//go:build !smokebot

// Command smoke is AgeForge's smoke suite runner. Its test bot is parked: build
// with -tags smokebot to run it (see cmd/smoke/main.go and CONTRIBUTING.md).
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "the smoke suite's test bot is parked: build with -tags smokebot, e.g. go run -tags smokebot ./cmd/smoke -list")
	os.Exit(2)
}
