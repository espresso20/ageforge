package mapstyle

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// lit counts the cells of a view's scene that wear the tint of the news.
func lit(v *view, f Frame) int {
	n := 0
	for _, c := range v.scene(f).Grid.Cells {
		if c.Bg.Hue == mapgen.HFresh {
			n++
		}
	}
	return n
}

// TestNewSinceLastVisitLightsThenClears: what is new since the last visit is
// lit when the map first shows it, stays lit for a while, and is clear for
// good once the map has been looked at that long; with the news setting off it
// is never lit, and news of another kind lights again. Lit or not, the
// picture draws at every size.
func TestNewSinceLastVisitLightsThenClears(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range []string{"medieval_age", "modern_age", "space_age", "galactic_age"} {
		m := grown(t, fixture.Options{Age: age, Seed: 7})
		more := grown(t, fixture.Options{Age: age, Seed: 9})
		at := func(model *mapmodel.Model, clock int) Frame { return Frame{Model: model, Anim: clock, Clock: clock} }
		v, off := New().(*view), New().(*view)
		off.SetOption(OptChanges, false)
		if lit(v, at(m, 1000)) == 0 {
			t.Errorf("%s: the news is not lit at the first look", age)
		}
		if lit(v, at(m, 1000+freshFrames/2)) == 0 {
			t.Errorf("%s: the news is clear before the map has been looked at", age)
		}
		if lit(v, at(m, 1000+freshFrames)) != 0 || lit(v, at(m, 1000+3*freshFrames)) != 0 {
			t.Errorf("%s: the news is still lit after the map has been looked at", age)
		}
		if lit(off, at(m, 1000)) != 0 {
			t.Errorf("%s: the news is lit with the setting off", age)
		}
		if lit(v, at(more, 5000)) == 0 {
			t.Errorf("%s: new news does not light", age)
		}
		for _, sz := range sizes {
			scr := drawFrame(New(), at(m, 1000), sz.w, sz.h, sz.compact)
			if strings.TrimSpace(capture.Text(scr)) == "" {
				t.Errorf("%s %s: blank screen", age, sz.name)
			}
			scr.Fini()
		}
	}
}

// TestNothingNewLightsNothing: a town with no news is not lit.
func TestNothingNewLightsNothing(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "modern_age", Seed: 7})
	if n := lit(New().(*view), Frame{Model: m, Anim: 3, Clock: 3}); n != 0 {
		t.Errorf("a town with no news has %d cells lit", n)
	}
}

// TestDistrictZoom: PgDn from the close-up frames a district and names it,
// Tab and Shift-Tab go round the districts, PgUp steps back out, and the
// district zoom draws at every size.
func TestDistrictZoom(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "modern_age", Seed: 7})
	f := Frame{Model: m}
	v := New().(*view)
	text := func() string {
		scr := drawFrame(v, f, 120, 36, false)
		defer scr.Fini()
		return capture.Text(scr)
	}
	press := func(k tcell.Key) { v.HandleKey(tcell.NewEventKey(k, 0, tcell.ModNone), f) }
	if !strings.Contains(text(), "SETTLEMENT") {
		t.Fatal("a big panel does not open on the close-up")
	}
	press(tcell.KeyPgDn)
	first := text()
	label := func(s string) string { return strings.TrimSpace(s[strings.Index(s, "DISTRICT"):strings.Index(s, "DISTRICT")+40]) }
	if !strings.Contains(first, "DISTRICT · ") || v.shownZoom() != zDistrict {
		t.Fatalf("PgDn did not frame a district:\n%s", first)
	}
	sc := v.scene(f)
	here := sc.Grid.ThingAt(v.cur.X, v.cur.Y)
	if here.Lineage != v.district || here.Kind != 3 {
		t.Errorf("the cursor is on %+v, not a building of %s", here, v.district)
	}
	seen := map[string]bool{v.district: true}
	for i := 0; i < len(sc.Districts); i++ {
		press(tcell.KeyTab)
		seen[v.district] = true
		if in, _ := v.Inspect(f); !strings.Contains(in.Title, "×") {
			t.Errorf("Tab left the cursor on %+v", in)
		}
	}
	if len(seen) != len(sc.Districts) {
		t.Errorf("Tab went to %d of %d districts", len(seen), len(sc.Districts))
	}
	was := v.district
	press(tcell.KeyBacktab)
	if v.district == was || label(text()) == "" {
		t.Error("Shift-Tab stayed in the same district")
	}
	press(tcell.KeyPgUp)
	if v.shownZoom() != zCloseUp || !strings.Contains(text(), "SETTLEMENT") {
		t.Error("PgUp did not step back to the close-up")
	}
	press(tcell.KeyPgDn)
	press(tcell.KeyPgDn)
	if v.shownZoom() != zDistrict {
		t.Error("PgDn at the district zoom left it")
	}
	for _, sz := range [][2]int{{200, 60}, {120, 36}, {100, 30}, {80, 24}} {
		scr := drawOffset(v, m, sz[0], sz[1], 3, mapmodel.TierUnicode, false)
		checkRect(t, scr, Rect{X: 2, Y: 1, W: sz[0], H: sz[1]}, "district zoom")
		scr.Fini()
	}
}

// TestDistrictZoomNeedsATown: with nothing built there is no district to
// frame, and PgDn stays on the close-up.
func TestDistrictZoomNeedsATown(t *testing.T) {
	fixture.WithTheme(t, "forge")
	st := fixture.State(fixture.Options{Age: "primitive_age", Seed: 7})
	st.Buildings = nil
	m := testBuilder.Build(&st, nil)
	v := New().(*view)
	f := Frame{Model: m}
	drawFrame(v, f, 120, 36, false).Fini()
	v.HandleKey(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone), f)
	if v.shownZoom() != zCloseUp {
		t.Errorf("PgDn with no town framed a district: zoom %d", v.shownZoom())
	}
}
