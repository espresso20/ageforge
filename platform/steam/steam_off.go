//go:build !steam

// Package steam is the game's one door to Steamworks. This build has no
// Steam: every call does nothing. See steam_on.go.
package steam

// Build is false in a build made without the steam tag.
const Build = false

// DataDir is empty: this build keeps its data beside the binary.
func DataDir() string { return "" }

// Start does nothing.
func Start() error { return nil }

// Frame does nothing.
func Frame() {}

// Unlock does nothing.
func Unlock(string) {}
