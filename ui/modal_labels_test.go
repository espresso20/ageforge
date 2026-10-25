package ui

import (
	"strings"
	"testing"
)

// Button labels like "[ENDURE]" parse as tview style tags and render blank
// unless escaped. These check what actually lands on screen, shortcuts included.
func TestCatastropheModalButtonLabelsVisible(t *testing.T) {
	d, pages := newTestDashboard(nil)
	d.showCatastropheModal("iron_era")
	screen := drawPrimitive(t, pages, 160, 50).Text()
	wantAll(t, "catastrophe modal", screen, "[E] ENDURE", "[S] SUCCUMB", "type 'catastrophe' to reopen")
	// Defer is gone: the choice waits behind Esc + the `catastrophe` command.
	if strings.Contains(screen, "Defer") {
		t.Errorf("catastrophe modal still offers Defer\n%s", screen)
	}
}

func TestAncientMemoryModalButtonLabelsVisible(t *testing.T) {
	d, pages := newTestDashboard(nil)
	d.showAncientMemoryModal("fire_mastery", "")
	screen := drawPrimitive(t, pages, 160, 50).Text()
	wantAll(t, "ancient memory modal", screen, "[A] Accept", "[D] Decline")
}
