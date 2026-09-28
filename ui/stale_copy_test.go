package ui

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// retiredCopy is player-facing text that was wrong and has been removed. Each
// entry says why, so nobody restores it from an old branch or doc.
var retiredCopy = []struct{ text, why string }{
	{"Reach Colonial Age to discover factions", "first contact comes from expeditions, and the Bronze Age civ is met long before Colonial"},
	{"Colonial Age and build an Embassy", "embassies raise opinion with civs already met; they never gated first contact"},
}

// TestNoRetiredCopy walks every non-test file under ui/ and site/ and fails if
// any retired string comes back, in any letter case. Test files are skipped:
// they name the strings in order to forbid them.
func TestNoRetiredCopy(t *testing.T) {
	root := ".."
	for _, dir := range []string{"ui", "site"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "node_modules" || strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			body := strings.ToLower(string(data))
			for _, r := range retiredCopy {
				if strings.Contains(body, strings.ToLower(r.text)) {
					rel, _ := filepath.Rel(root, path)
					t.Errorf("%s: contains retired text %q (%s)", filepath.ToSlash(rel), r.text, r.why)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
