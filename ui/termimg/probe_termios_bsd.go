//go:build darwin || freebsd || netbsd || openbsd

package termimg

import "golang.org/x/sys/unix"

const getTermios, setTermios = unix.TIOCGETA, unix.TIOCSETA
