package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/ageforge/game"
	"github.com/user/ageforge/ui"
)

func main() {
	// Create game engine
	engine := game.NewGameEngine()

	// Create UI
	app := ui.NewApp(engine)

	// Handle OS signals for clean exit
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		engine.Stop()
		app.Stop()
	}()

	// Run UI (blocks until exit)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
