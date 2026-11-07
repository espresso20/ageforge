//go:build linux || darwin || freebsd || netbsd || openbsd

package termimg

import (
	"bytes"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// Query asks the terminal what it can show. It talks to the controlling
// terminal directly and must run before the screen takes the input over. It
// puts the terminal in raw mode for the moment (with a read that times out), and
// gives up on a terminal that stays silent.
func Query() Probe {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		return Probe{}
	}
	defer f.Close()
	fd := int(f.Fd())
	old, err := unix.IoctlGetTermios(fd, getTermios)
	if err != nil {
		return Probe{}
	}
	raw := *old
	raw.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	raw.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	raw.Cflag &^= unix.CSIZE | unix.PARENB
	raw.Cflag |= unix.CS8
	raw.Cc[unix.VMIN], raw.Cc[unix.VTIME] = 0, 1 // a read returns after a tenth of a second of silence
	if unix.IoctlSetTermios(fd, setTermios, &raw) != nil {
		return Probe{}
	}
	defer unix.IoctlSetTermios(fd, setTermios, old)
	if _, err := f.WriteString(probeQueries); err != nil {
		return Probe{}
	}
	var got []byte
	buf := make([]byte, 256)
	for quiet, end := 0, time.Now().Add(time.Second); quiet < 3 && time.Now().Before(end); {
		n, err := unix.Read(fd, buf)
		if err != nil && err != unix.EINTR {
			break
		}
		if n == 0 {
			quiet++
			continue
		}
		quiet = 0
		got = append(got, buf[:n]...)
		if i := bytes.Index(got, []byte("\x1b[?")); i >= 0 && bytes.IndexByte(got[i:], 'c') > 0 {
			break // the device attributes come last
		}
	}
	return ParseReplies(got)
}
