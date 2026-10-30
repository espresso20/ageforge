//go:build !windows

package presence

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

// dialIPC connects to the first Discord socket that answers, discord-ipc-0 to -9.
func dialIPC() (io.ReadWriteCloser, error) {
	for _, p := range socketPaths() {
		c, err := net.DialTimeout("unix", p, time.Second)
		if err == nil {
			return c, nil
		}
	}
	return nil, fmt.Errorf("presence: no Discord socket found")
}

// socketPaths lists the places Discord puts its sockets, including the snap and flatpak
// sandboxes, in the order to try them.
func socketPaths() []string {
	var dirs []string
	seen := map[string]bool{}
	add := func(d string) {
		if d != "" && !seen[d] {
			seen[d] = true
			dirs = append(dirs, d)
		}
	}
	for _, env := range []string{"XDG_RUNTIME_DIR", "TMPDIR", "TMP", "TEMP"} {
		add(os.Getenv(env))
	}
	add("/tmp")
	var out []string
	for _, d := range dirs {
		for _, sub := range []string{"", "snap.discord", filepath.Join("app", "com.discordapp.Discord")} {
			for i := 0; i < 10; i++ {
				out = append(out, filepath.Join(d, sub, fmt.Sprintf("discord-ipc-%d", i)))
			}
		}
	}
	return out
}
