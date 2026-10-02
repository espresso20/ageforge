//go:build mapcapture

package all

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// TestWriteVisitorCapture draws the Space Age town in both styles during its
// first visit of each kind (a saucer flying over, one hovering, a figure
// walking; the skyline shows a saucer for all three), a second apart, for
// reviewing the rare visitor:
//
//	go test -tags mapcapture -run TestWriteVisitorCapture ./ui/mapstyle/all
func TestWriteVisitorCapture(t *testing.T) {
	root := filepath.Join("..", "..", "..", "map_captures")
	states := envOr("MAP_STATES_DIR", filepath.Join(root, "states"))
	out := envOr("MOVERS_CAPTURE_DIR", filepath.Join(root, "movers"))
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	if err := theme.SetActive("forge"); err != nil {
		t.Fatal(err)
	}
	st, err := capture.LoadState(filepath.Join(states, "space_age.json.gz"))
	if err != nil {
		t.Skipf("no space_age state: %v", err)
	}
	seen := map[mapmodel.SightingKind]bool{}
	for a := 0; len(seen) < 3 && a < 400*mapmodel.SightingBlock; {
		sg := mapmodel.NextSighting(st.Seed, true, a)
		a = sg.Start + sg.Frames
		if seen[sg.Kind] {
			continue
		}
		seen[sg.Kind] = true
		name := map[mapmodel.SightingKind]string{mapmodel.SightFlyby: "visitor_flyby", mapmodel.SightHover: "visitor_hover",
			mapmodel.SightWalker: "visitor_walker"}[sg.Kind]
		var frames []int
		for f := sg.Start + sg.Frames*2/5; f < sg.Start+sg.Frames*3/5 && len(frames) < 12; f += 8 {
			frames = append(frames, f)
		}
		note := fmt.Sprintf("a visit from frame %d for %d frames", sg.Start, sg.Frames)
		moverShot(t, states, out, "roguelike", "space_age", name, mapmodel.TierUnicode, frames, note)
		moverShot(t, states, out, "skyline", "space_age", "skyline_"+name, mapmodel.TierUnicode, frames, note)
	}
}
