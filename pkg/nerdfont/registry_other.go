//go:build !windows

package nerdfont

// registerFonts is Windows only.
var registerFonts func(files []string) error
