package skyline

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

var (
	catOnce sync.Once
	testCat *mapmodel.Catalog
)

func catalog() *mapmodel.Catalog {
	catOnce.Do(func() { testCat = mapmodel.NewCatalog() })
	return testCat
}

// build makes a model from fixture options; with grow > 0 the state is
// grown that many steps past a baseline, so the recap and ▼ markers show.
func build(o fixture.Options, grow int) *mapmodel.Model {
	b := mapmodel.NewBuilder(catalog())
	st := fixture.State(o)
	if grow <= 0 {
		return b.Build(&st, nil)
	}
	base := b.Build(&st, nil)
	since := mapmodel.VisitOf(base)
	for i := 1; i <= grow; i++ {
		st = fixture.Grow(st, i)
	}
	return b.Build(&st, since)
}

// tickAt is a tick at a time of day on day 3.
// (the inverse of mapmodel.ClockAtTOD's own arithmetic).
func tickAt(tod float64) int {
	return 2*mapmodel.DayTicks + int(tod*mapmodel.DayTicks) - mapmodel.DayTicks*3/8
}

func draw(v *view, m *mapmodel.Model, w, h, anim int, tier mapmodel.GlyphTier, compact bool) tcell.SimulationScreen {
	scr := capture.NewScreen(w, h)
	f := mapstyle.Frame{Model: m, Anim: anim, Tier: tier}
	if compact {
		v.DrawCompact(scr, mapstyle.Rect{W: w, H: h}, f)
	} else {
		v.Draw(scr, mapstyle.Rect{W: w, H: h}, f)
	}
	scr.Show()
	return scr
}

// TestSpotCaptures writes review captures when SKYLINE_CAPTURE names a
// directory.
func TestSpotCaptures(t *testing.T) {
	dir := os.Getenv("SKYLINE_CAPTURE")
	if dir == "" {
		t.Skip("set SKYLINE_CAPTURE to a directory to write captures")
	}
	type shot struct {
		name    string
		o       fixture.Options
		tod     float64
		w, h    int
		compact bool
		flows   bool
		inspect bool
		tier    mapmodel.GlyphTier
		theme   string
		focus   int // centre on this district (0: the present)
	}
	shots := []shot{
		{name: "primitive_day", o: fixture.Options{Age: "primitive_age", Seed: 3}, tod: 0.5, w: 160, h: 45},
		{name: "medieval_day", o: fixture.Options{Age: "medieval_age", Seed: 3, Harbinger: true, Wars: 1}, tod: 0.45, w: 160, h: 45},
		{name: "medieval_dusk_inspect", o: fixture.Options{Age: "medieval_age", Seed: 3}, tod: 0.76, w: 160, h: 45, inspect: true},
		{name: "industrial_day", o: fixture.Options{Age: "industrial_age", Seed: 5, Routes: 4}, tod: 0.55, w: 160, h: 45},
		{name: "industrial_flows", o: fixture.Options{Age: "industrial_age", Seed: 5}, tod: 0.55, w: 160, h: 45, flows: true},
		{name: "cyberpunk_night", o: fixture.Options{Age: "cyberpunk_age", Seed: 7, Routes: 5, Harbinger: true}, tod: 0.95, w: 160, h: 45},
		{name: "cyberpunk_day", o: fixture.Options{Age: "cyberpunk_age", Seed: 7, Routes: 5}, tod: 0.5, w: 160, h: 45},
		{name: "fusion_day", o: fixture.Options{Age: "fusion_age", Seed: 2, Routes: 3}, tod: 0.5, w: 160, h: 45},
		{name: "space_night", o: fixture.Options{Age: "space_age", Seed: 2, Routes: 3}, tod: 0.05, w: 160, h: 45},
		{name: "galactic_day", o: fixture.Options{Age: "galactic_age", Seed: 4}, tod: 0.5, w: 160, h: 45},
		{name: "transcendent_night", o: fixture.Options{Age: "transcendent_age", Seed: 4, Catastrophe: true}, tod: 0.9, w: 160, h: 45},
		{name: "digital_catastrophe", o: fixture.Options{Age: "digital_age", Seed: 4, Catastrophe: true}, tod: 0.9, w: 160, h: 45},
		{name: "renaissance_80x24", o: fixture.Options{Age: "renaissance_age", Seed: 8}, tod: 0.5, w: 80, h: 24},
		{name: "compact_cyberpunk", o: fixture.Options{Age: "cyberpunk_age", Seed: 7, Routes: 5, Harbinger: true}, tod: 0.95, w: 40, h: 15, compact: true},
		{name: "compact_medieval", o: fixture.Options{Age: "medieval_age", Seed: 3}, tod: 0.5, w: 40, h: 15, compact: true},
		{name: "light_industrial_night", o: fixture.Options{Age: "industrial_age", Seed: 5}, tod: 0.95, w: 160, h: 45, theme: "daylight"},
		{name: "bronze_day", o: fixture.Options{Age: "bronze_age", Seed: 3}, tod: 0.4, w: 160, h: 45},
		{name: "classical_dusk", o: fixture.Options{Age: "classical_age", Seed: 3, Routes: 3}, tod: 0.74, w: 160, h: 45},
		{name: "victorian_night", o: fixture.Options{Age: "victorian_age", Seed: 3, Routes: 3}, tod: 0.9, w: 160, h: 45},
		{name: "electric_wonders", o: fixture.Options{Age: "modern_age", Seed: 3}, tod: 0.5, w: 200, h: 50, focus: 10},
		{name: "atomic_wonders", o: fixture.Options{Age: "modern_age", Seed: 3}, tod: 0.5, w: 200, h: 50, focus: 11},
		{name: "modern_day", o: fixture.Options{Age: "modern_age", Seed: 3, Routes: 4}, tod: 0.5, w: 160, h: 45},
		{name: "interstellar_night", o: fixture.Options{Age: "interstellar_age", Seed: 3, Routes: 4}, tod: 0.9, w: 160, h: 45},
		{name: "quantum_day", o: fixture.Options{Age: "quantum_age", Seed: 3, Routes: 4}, tod: 0.5, w: 160, h: 45},
		{name: "late_wonders", o: fixture.Options{Age: "transcendent_age", Seed: 3}, tod: 0.4, w: 200, h: 50, focus: 17},
		{name: "late_wonders2", o: fixture.Options{Age: "transcendent_age", Seed: 3}, tod: 0.4, w: 200, h: 50, focus: 19},
		{name: "ascii_colonial", o: fixture.Options{Age: "colonial_age", Seed: 5}, tod: 0.5, w: 120, h: 40, tier: mapmodel.TierASCII},
	}
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	for _, sh := range shots {
		if sh.theme != "" {
			_ = theme.SetActive(sh.theme)
		} else {
			_ = theme.SetActive(orig)
		}
		sh.o.Tick = tickAt(sh.tod) - 600 // fixture.Grow adds 600 ticks
		m := build(sh.o, 1)
		v := newView()
		v.flows = sh.flows
		v.inspect = sh.inspect
		if sh.focus > 0 && sh.focus < len(m.Skyline.Districts) {
			v.follow = false
			v.cam = m.Skyline.Districts[sh.focus].Centre - sh.w/2
		}
		var frames []string
		var txt string
		for k := 0; k < 6; k++ {
			scr := draw(v, m, sh.w, sh.h, 40+k*2, sh.tier, sh.compact)
			frames = append(frames, capture.Frame(scr))
			if k == 0 {
				txt = capture.Text(scr)
				_ = writePNG(filepath.Join(dir, sh.name+".png"), scr)
			}
		}
		bg := capture.Hex(theme.Color(theme.RoleBackground))
		fg := capture.Hex(theme.Color(theme.RoleText))
		_ = os.WriteFile(filepath.Join(dir, sh.name+".txt"), []byte(txt), 0o644)
		_ = os.WriteFile(filepath.Join(dir, sh.name+".html"), []byte(capture.HTML(sh.name, bg, fg, capture.Fonts, frames, 6)), 0o644)
	}
}
