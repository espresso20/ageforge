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
	// — which share this engine — can reach it via engine.Account() (the accounts design §6). We
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
	go exitOnSignal(sigs, engine, app.Stop)

	// Run UI (blocks until exit). It returns however the player leaves: the menu's
	// Quit, the quit command, Ctrl+C (a key to the terminal, not a signal) or a signal.
	err := app.Run()
	leave(engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// exitOnSignal waits for a signal that ends the program, then leaves the game and closes
// the UI, which makes Run return.
func exitOnSignal(sigs <-chan os.Signal, engine *game.GameEngine, closeUI func()) {
	<-sigs
	leave(engine)
	closeUI()
}

// leave is the one way out of the game, whatever ended the program. A game in play is
// saved to its own save (game.SaveOnExit); with no game in play nothing is written. Then
// the engine is stopped. It may run twice (a signal, then the end of main): the second
// time there is nothing left to do.
func leave(engine *game.GameEngine) {
	if name, err := engine.SaveOnExit(); err != nil {
		fmt.Fprintf(os.Stderr, "The game could not be saved to '%s': %v\n", name, err)
	}
	engine.Stop()
}
