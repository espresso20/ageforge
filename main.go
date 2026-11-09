package main

import (
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/presence"
	"github.com/espresso20/ageforge/ui"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window"
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

	// The default is the window; --terminal keeps the terminal for now.
	var err error
	var win *window.Game
	var done chan struct{}
	if host, herr := window.NewHost(); herr != nil {
		err = herr
	} else if !slices.Contains(os.Args[1:], "--terminal") {
		done = make(chan struct{})
		win = window.New(window.Options{
			Host: host, Dir: game.RootDataDir(), Done: done,
			// Closing the window leaves the game the way Ctrl+C does.
			OnClose: func() { leave(engine); app.Stop() },
			Script:  os.Getenv("AGEFORGE_WINDOW_SCRIPT"),
		})
		app.SetScreen(host)
		app.SetWindowHost(win)
	}

	// Discord Rich Presence: off unless the build carries an application id. It follows
	// the menu and the game by reading the engine on its own goroutine.
	status := presence.Start(presence.Options{AppID: presence.AppID, Activity: presence.Provider(engine, time.Now())})
	app.OnDiscordChange(status.Refresh)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go exitOnSignal(sigs, engine, app.Stop)

	// Run blocks until the player leaves by any route: the menu's Quit, the quit command,
	// Ctrl+C (a key to the terminal, not a signal) or a signal.
	switch {
	case err != nil: // the window could not be set up: nothing to run
	case win == nil:
		err = app.Run()
	default:
		// The window owns the main thread (macOS needs that); the game runs beside it.
		// When the game ends the window closes, and when the window closes the game stops.
		errc := make(chan error, 1)
		go func() { errc <- app.Run(); close(done) }()
		werr := win.Run()
		app.Stop()
		if err = <-errc; err == nil {
			err = werr
		}
	}
	status.Close()
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
