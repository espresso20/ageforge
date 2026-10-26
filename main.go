package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/ui"
	"github.com/espresso20/ageforge/vibe"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("ageforge %s\n", version)
		return
	}

	// Load an existing, named account before the engine exists, so the engine's first
	// words are in the script the account chose (VibeMode). The UI shares the engine and
	// reaches the account through engine.Account(). Nothing is created here: on first run
	// (no file, or an account with no name) the engine has no account and the UI asks the
	// player to name one (game.CreateAccount). A load error does not block play: the game
	// runs without an account, and the log says why.
	acct, found, loadErr := game.LoadAccount()
	loaded := found && acct.Established()
	if loaded {
		vibe.Set(acct.VibeOn())
	}

	engine := game.NewGameEngine()
	if loaded {
		engine.SetAccount(acct)
	}
	if loadErr != nil {
		engine.AddLog("error", loadErr.Error())
	}

	app := ui.NewApp(engine, version)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go exitOnSignal(sigs, engine, app.Stop)

	// Run blocks until the player leaves by any route: the menu's Quit, the quit command,
	// Ctrl+C (a key to the terminal, not a signal) or a signal.
	err := app.Run()
	leave(engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// exitOnSignal waits for a signal that ends the program, leaves the game, then closes the
// UI, which makes Run return.
func exitOnSignal(sigs <-chan os.Signal, engine *game.GameEngine, closeUI func()) {
	<-sigs
	leave(engine)
	closeUI()
}

// leave is the one way out of the game, whatever ended the program. It saves a game in
// play to its own save (game.SaveOnExit), writes nothing if no game is in play, then stops
// the engine. It can run twice (a signal, then the end of main); the second call does
// nothing.
func leave(engine *game.GameEngine) {
	if name, err := engine.SaveOnExit(); err != nil {
		fmt.Fprintf(os.Stderr, "The game could not be saved to '%s': %v\n", name, err)
	}
	engine.Stop()
}
