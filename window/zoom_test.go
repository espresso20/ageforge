package window

import "testing"

// The Settings page and the keys change the text size through one call.
func TestZoomByChangesTheTextSize(t *testing.T) {
	g := &Game{set: defaultSettings(), scale: 1}
	was := g.TextSize()
	g.ZoomBy(2)
	if got := g.TextSize(); got != was+2 {
		t.Errorf("size %d after two steps from %d", got, was)
	}
	g.ZoomBy(-1)
	if got := g.TextSize(); got != was+1 {
		t.Errorf("size %d after one step back", got)
	}
}
