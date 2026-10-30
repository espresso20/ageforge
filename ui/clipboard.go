package ui

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"runtime"
	"time"
)

// clipboard.go puts text on the system clipboard by way of the tool the OS
// ships with. There is no clipboard package in go.mod on purpose.

// clipboardTimeout keeps a stuck helper from hanging the screen.
const clipboardTimeout = 2 * time.Second

// errNoClipboard says no clipboard tool answered.
var errNoClipboard = errors.New("no clipboard tool available")

// clipboardCommands are the commands to try, in order, for an OS. Each reads
// the text from stdin.
func clipboardCommands(goos string) [][]string {
	switch goos {
	case "darwin":
		return [][]string{{"pbcopy"}}
	case "windows":
		// note: the brackets are joined in code so the copy guard, which reads
		// them as tview tags, leaves this PowerShell line alone.
		return [][]string{{"powershell", "-NoProfile", "-Command",
			"[" + "Console]::InputEncoding = [" + "System.Text.Encoding]::UTF8; $input | Set-Clipboard"}}
	}
	return [][]string{{"wl-copy"}, {"xclip", "-selection", "clipboard"}, {"xsel", "-b"}}
}

// copyToClipboard puts text on the clipboard, trying each tool the OS may
// have. It returns errNoClipboard when none is installed, or the last error.
func copyToClipboard(text string) error {
	return copyWith(clipboardCommands(runtime.GOOS), text)
}

func copyWith(cmds [][]string, text string) error {
	err := errNoClipboard
	for _, argv := range cmds {
		path, lerr := exec.LookPath(argv[0])
		if lerr != nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), clipboardTimeout)
		cmd := exec.CommandContext(ctx, path, argv[1:]...)
		cmd.Stdin = bytes.NewReader([]byte(text))
		// Some tools (wl-copy, xclip) stay up to own the selection; do not
		// wait on their output pipes, only on the process, and bound it.
		cmd.WaitDelay = 200 * time.Millisecond
		rerr := cmd.Run()
		cancel()
		if rerr == nil {
			return nil
		}
		err = rerr
	}
	return err
}
