package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/ui"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("ageforge %s\n", version)
		return
	}

	// Create game engine
	engine := game.NewGameEngine()

	// Load (or first-run create) the per-player account and hand it to the engine, so
	// the UI/dashboard — which share this engine — can reach it via engine.Account()
	// in later phases (accounts.md §6). The account is non-critical: if loading fails,
	// log to stderr and run without one rather than blocking play.
	acct, err := game.LoadOrCreate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load account, continuing without one: %v\n", err)
	} else if acct != nil {
		engine.SetAccount(acct)
		if acct.FreshlyCreated {
			// Non-blocking first-run notice: surfaces in the in-game log, never gates
			// play (accounts.md §6). Player-facing, no jargon.
			engine.AddLog("info", "Welcome! A local account was created on this machine to track your unlocks across all your games.")
		}
	}

	// Create UI
	app := ui.NewApp(engine, version)

	// Handle OS signals for clean exit
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		engine.SaveGame("autosave")
		engine.Stop()
		app.Stop()
	}()

	// Run UI (blocks until exit)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
