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

	// Load an EXISTING, established account and hand it to the engine so the UI/dashboard
	// — which share this engine — can reach it via engine.Account() (accounts.md §6). We
	// no longer auto-create here: on first run (no file, or a legacy unnamed account) the
	// engine is left accountless and the UI prompts the player to name their account, which
	// derives the identity (game.CreateNamedAccount). Loading is non-critical: ignore errors
	// and run accountless rather than block play.
	if acct, found, _ := game.LoadAccount(); found && acct.Established() {
		engine.SetAccount(acct)
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
