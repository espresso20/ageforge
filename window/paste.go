package window

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Ebitengine has no clipboard, so paste asks the OS the same way
// ui/clipboard.go does for pictures: by running its own tool.

// pasteTools are the readers to try, in order, for a GOOS.
func pasteTools(goos string) [][]string {
	switch goos {
	case "darwin":
		return [][]string{{"pbpaste"}}
	case "windows":
		return [][]string{{"powershell", "-NoProfile", "-Command", "Get-Clipboard -Raw"}}
	}
	return [][]string{{"wl-paste", "-n"}, {"xclip", "-selection", "clipboard", "-o"}, {"xsel", "-ob"}}
}

// clipboardText is the text on the clipboard as one line ("" if none).
func clipboardText() string {
	for _, argv := range pasteTools(runtime.GOOS) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		out, err := exec.CommandContext(ctx, argv[0], argv[1:]...).Output()
		cancel()
		if err == nil {
			return oneLine(string(out))
		}
	}
	return ""
}

// oneLine folds line breaks to spaces: the command line takes a single line.
func oneLine(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\r", "\n")), " ")
}
