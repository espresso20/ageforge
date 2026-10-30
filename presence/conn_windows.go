//go:build windows

package presence

import (
	"fmt"
	"io"
	"os"
)

// dialIPC opens the first Discord named pipe that exists, discord-ipc-0 to -9. A named
// pipe opens as a file, so the standard library is enough.
func dialIPC() (io.ReadWriteCloser, error) {
	for i := 0; i < 10; i++ {
		f, err := os.OpenFile(fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, i), os.O_RDWR, 0)
		if err == nil {
			return f, nil
		}
	}
	return nil, fmt.Errorf("presence: no Discord pipe found")
}
