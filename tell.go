//go:build !windows

package main

import (
	"fmt"
	"os"
)

// tell says something the player must see when the game cannot show it on
// its own screen (it failed to start, or to save on the way out). Here that
// is the shell the game was started from.
func tell(msg string) { fmt.Fprint(os.Stderr, msg) }
