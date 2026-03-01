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
