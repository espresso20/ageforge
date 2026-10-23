//go:build !windows

package nerdfont

// registerFonts is nil off Windows, where the build has no registry support.
var registerFonts func(files []string) error
