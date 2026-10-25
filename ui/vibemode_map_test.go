package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/ui/mapstyle/all"
	"github.com/espresso20/ageforge/vibe"
)

// TestTheMapSwitchesAtOnce: every text of the map is read when the map is drawn,
// or when its model is built, and the model is built again when the switch
// moves, so the switch is not a restart. With a rewriter in place of the script
// (every text the code writes itself reads in guillemets), each map style is
// drawn in the Map panel, with the flows overlay, then the mini map on the
// dashboard, with VibeMode off, then on, then off again, from the one game that
// is never rebuilt. On changes the picture; off again gives back exactly what
// the first draw gave.
func TestTheMapSwitchesAtOnce(t *testing.T) {
	vibeOffAfter(t)
	defer vibe.UseRewriterForTest(func(id, plain string) string { return "«" + plain + "»" })()
	d, pages, eng := badgeDashboard(t)
	d.runForTest("motion off")
	// The map's clock is pinned, so a frame is a function of the game.
	pinned := time.Now()
	d.mapPanel.start, d.mapPanel.now = pinned, func() time.Time { return pinned }
	set := func(on bool) {
		t.Helper()
		must(t, eng.SetVibeMode(on))
		d.refresh()
	}

	// shots draws the panel of each style as the player would meet it: as
	// opened, with the cursor on a few things (Tab), and with the flows overlay.
	shots := func() map[string]string {
		out := map[string]string{}
		for _, style := range all.Registry().Names() {
			d.runForTest("style " + style)
			d.runForTest("map")
			if d.overlayMgr.ActiveName() != "map" {
				t.Fatalf("map did not open the Map panel (style %s)", style)
			}
			d.refresh()
			out[style+" as opened"] = drawMapPanel(t, d, 160, 48)
			// the cursor home, then on the third thing Tab reaches: the inspector
			// lines of whatever stands there
			for _, k := range []tcell.Key{tcell.KeyHome, tcell.KeyTab, tcell.KeyTab, tcell.KeyTab} {
				d.mapPanel.routeKey(tcell.NewEventKey(k, 0, tcell.ModNone), "")
			}
			out[style+" inspecting"] = drawMapPanel(t, d, 160, 48)
			d.runForTest("map flows on")
			out[style+" with flows"] = drawMapPanel(t, d, 160, 48)
			d.runForTest("map flows off")
			d.overlayMgr.Hide()
		}
		d.runForTest("minimap on")
		d.refresh()
		out["the mini map"] = drawDashboard(t, d, pages, 160, 50)
		d.runForTest("minimap off")
		return out
	}

	shots() // the first frames learn the screen's size and where the focus goes; the ones after it are what is compared
	plain := shots()
	for name, text := range plain {
		if strings.Contains(text, "«") {
			t.Errorf("%s: the map has a rewritten text with VibeMode off", name)
		}
	}

	set(true)
	on := shots()
	// The words the model itself holds, such as the town's name made of the
	// script's syllables, are in the script now in force: the builder is not
	// the cache that keeps the other one's.
	if name := d.mapPanel.model.Town.World.Name; !strings.Contains(name, "«") {
		t.Errorf("the town is called %q with VibeMode on", name)
	}
	// The registry the panel holds was made before the switch; the style's name
	// on the key bar is read from its entry when the bar is drawn.
	for style, title := range map[string]string{"roguelike": "«Roguelike»", "skyline": "«Skyline»"} {
		if !strings.Contains(on[style+" as opened"], title) {
			t.Errorf("the key bar of the %s map does not name its style %s with VibeMode on", style, title)
		}
	}
	for name, text := range on {
		if text == plain[name] || !strings.Contains(text, "«") {
			t.Errorf("%s: VibeMode is on and the map reads as it did", name)
		}
	}

	set(false)
	back := shots()
	if name := d.mapPanel.model.Town.World.Name; strings.Contains(name, "«") {
		t.Errorf("the town is called %q with VibeMode off again", name)
	}
	for name, text := range back {
		if text != plain[name] {
			t.Errorf("%s: with VibeMode off again the map is not what it was: %s", name, firstDifference(plain[name], text))
		}
	}
}
