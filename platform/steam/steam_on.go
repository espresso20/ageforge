//go:build steam

// Package steam is the game's one door to Steamworks. A build with the steam
// tag talks to Steam; any other build gets the no-ops in steam_off.go and
// imports nothing from Steam.
package steam

import (
	"os"
	"path/filepath"
	"sync"

	sw "github.com/hajimehoshi/go-steamworks"
)

// AppID is Spacewar, Valve's test app, until AgeForge has its own App ID.
const AppID = 480

// Build is true in a build made with the steam tag.
const Build = true

// DataDir is the folder a Steam build keeps its data in: AgeForge under the
// OS user config folder, where Steam Auto-Cloud can sync from. Empty if the
// folder cannot be found.
func DataDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "AgeForge")
}

// achievements maps a badge id to its Steam API name. Empty until the
// shortlist of Steam achievements is decided.
var achievements = map[string]string{}

// api is the part of Steamworks this package calls, so tests can stand in.
type api interface {
	runCallbacks()
	setAchievement(name string)
	storeStats()
}

type live struct{}

func (live) runCallbacks()           { sw.RunCallbacks() }
func (live) setAchievement(n string) { sw.SteamUserStats().SetAchievement(n) }
func (live) storeStats()             { sw.SteamUserStats().StoreStats() }

var (
	mu      sync.Mutex
	backend api // nil until Start succeeds
	queue   []string
)

// Start opens Steamworks. It must run before the window is created, so the
// overlay can hook the graphics device. If Steam wants to launch the game
// itself it exits and lets it. An error means the game carries on without
// Steam.
func Start() error {
	if sw.RestartAppIfNecessary(AppID) {
		os.Exit(0)
	}
	if err := sw.Init(); err != nil {
		return err
	}
	mu.Lock()
	backend = live{}
	mu.Unlock()
	return nil
}

// Frame runs Steam's callbacks, then sends the queued unlocks. Call it once
// per Update, from the game loop.
func Frame() {
	mu.Lock()
	b, q := backend, queue
	if b != nil {
		queue = nil
	}
	mu.Unlock()
	if b == nil {
		return
	}
	b.runCallbacks()
	flush(b, q)
}

func flush(b api, names []string) {
	for _, n := range names {
		b.setAchievement(n)
	}
	if len(names) > 0 {
		b.storeStats()
	}
}

// Unlock queues the Steam achievement for a badge id; an id with no entry in
// achievements is ignored. It only queues, so any goroutine may call it,
// bus handlers included.
func Unlock(badge string) {
	name, ok := achievements[badge]
	if !ok {
		return
	}
	mu.Lock()
	queue = append(queue, name)
	mu.Unlock()
}
