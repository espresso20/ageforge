package ui

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// clipboard.go puts a picture on the system clipboard by way of the tool the OS
// ships with. There is no clipboard package in go.mod on purpose.

// clipboardTimeout keeps a stuck helper from hanging the screen.
const clipboardTimeout = 2 * time.Second

// errNoClipboard says no clipboard tool answered.
var errNoClipboard = errors.New("no clipboard tool available")

// imageCommand is one way to put a PNG file on the clipboard: the command to
// run, and whether it reads the file from its standard input.
type imageCommand struct {
	argv  []string
	stdin bool
}

// clipboardImageCommands are the commands to try, in order, to put the PNG
// at path on the clipboard. The path is handed over as an argument or on
// stdin, or quoted for the one script language that has to carry it inside a
// string, so a data folder with spaces or quotes in its name is safe.
func clipboardImageCommands(goos, path string) []imageCommand {
	switch goos {
	case "darwin":
		script := `set the clipboard to (read (POSIX file "` +
			strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(path) +
			`") as ` + "\u00abclass PNGf\u00bb)"
		return []imageCommand{{argv: []string{"osascript", "-e", script}}}
	case "windows":
		script := "Add-Type -AssemblyName System.Windows.Forms,System.Drawing; " +
			"[System.Windows.Forms.Clipboard]::SetImage([System.Drawing.Image]::FromFile('" +
			strings.ReplaceAll(path, "'", "''") + "'))"
		return []imageCommand{{argv: []string{"powershell", "-NoProfile", "-STA", "-Command", script}}}
	}
	return []imageCommand{
		{argv: []string{"wl-copy", "--type", "image/png"}, stdin: true},
		{argv: []string{"xclip", "-selection", "clipboard", "-t", "image/png", "-i", path}},
	}
}

// copyImageToClipboard puts the PNG at path on the clipboard, trying each
// tool the OS may have. It returns errNoClipboard when none is installed, or
// the last error.
func copyImageToClipboard(path string) error {
	return copyImageWith(clipboardImageCommands(runtime.GOOS, path), path)
}

func copyImageWith(cmds []imageCommand, path string) error {
	err := errNoClipboard
	for _, c := range cmds {
		bin, lerr := exec.LookPath(c.argv[0])
		if lerr != nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), clipboardTimeout)
		cmd := exec.CommandContext(ctx, bin, c.argv[1:]...)
		var f *os.File
		if c.stdin {
			var oerr error
			if f, oerr = os.Open(path); oerr != nil {
				cancel()
				err = oerr
				continue
			}
			cmd.Stdin = f
		}
		cmd.WaitDelay = 200 * time.Millisecond
		rerr := cmd.Run()
		cancel()
		if f != nil {
			f.Close()
		}
		if rerr == nil {
			return nil
		}
		err = rerr
	}
	return err
}
