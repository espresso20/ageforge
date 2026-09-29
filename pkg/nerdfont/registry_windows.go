//go:build windows

package nerdfont

import (
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// fontsKey is the per-user font registration key (Windows 10 1809 and
// later read it; no admin rights needed).
const fontsKey = `Software\Microsoft\Windows NT\CurrentVersion\Fonts`

// registerFonts writes one HKCU font entry per installed file.
var registerFonts = func(files []string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, fontsKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	for _, f := range files {
		if err := k.SetStringValue(fontValueName(filepath.Base(f)), f); err != nil {
			return err
		}
	}
	return nil
}
