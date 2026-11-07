//go:build !(linux || darwin || freebsd || netbsd || openbsd)

package termimg

// Query cannot ask the terminal here (Windows): the environment hints decide.
func Query() Probe { return Probe{} }
