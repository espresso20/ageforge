package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// tell says something the player must see when the game cannot show it on
// its own screen (it failed to start, or to save on the way out). The
// Windows build has no console (-H=windowsgui), so besides the shell, which
// is usually not there, the message goes in a box.
func tell(msg string) {
	fmt.Fprint(os.Stderr, msg)
	text, err1 := syscall.UTF16PtrFromString(strings.TrimSpace(msg))
	title, err2 := syscall.UTF16PtrFromString(boxTitle)
	if err1 != nil || err2 != nil {
		return
	}
	const okButton, errorIcon = 0x0, 0x10
	_, _, _ = syscall.NewLazyDLL(boxLibrary).NewProc(boxCall).Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), okButton|errorIcon)
}

// The message box of the system, and the name on its title bar.
const (
	boxLibrary = "user32.dll"
	boxCall    = "MessageBoxW"
	boxTitle   = "AgeForge"
)
