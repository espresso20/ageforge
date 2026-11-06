package mapstyle

import (
	"os"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// TestZoomLadder: PgUp and PgDn step through COSMOS, OVERVIEW, INNER,
// CLOSE-UP, DISTRICT in the space scenes, the Interstellar Age without the
// district, and through OVERVIEW, CLOSE-UP, DISTRICT everywhere else, never
// landing on INNER.
func TestZoomLadder(t *testing.T) {
	fixture.WithTheme(t, "forge")
	cases := []struct {
		age  string
		want []int // from the top of the ladder down, then back up
	}{
		{"space_age", []int{zCosmos, zOverview, zInner, zCloseUp, zDistrict}},
		{"interstellar_age", []int{zCosmos, zOverview, zInner, zCloseUp}},
		{"galactic_age", []int{zCosmos, zOverview, zInner, zCloseUp, zDistrict}},
		{"medieval_age", []int{zOverview, zCloseUp, zDistrict}},
	}
	for _, c := range cases {
		m := modelFor(t, fixture.Options{Age: c.age, Seed: 7})
		v := New().(*view)
		f := Frame{Model: m}
		press := func(k tcell.Key) { v.HandleKey(tcell.NewEventKey(k, 0, tcell.ModNone), f) }
		for i := 0; i < 8; i++ { // all the way out
			press(tcell.KeyPgUp)
		}
		if v.shownZoom() != c.want[0] {
			t.Fatalf("%s: far out is zoom %d, want %d", c.age, v.shownZoom(), c.want[0])
		}
		for _, w := range c.want[1:] {
			press(tcell.KeyPgDn)
			if v.shownZoom() != w {
				t.Fatalf("%s: PgDn gave zoom %d, want %d", c.age, v.shownZoom(), w)
			}
		}
		press(tcell.KeyPgDn)
		if end := c.want[len(c.want)-1]; v.shownZoom() != end {
			t.Errorf("%s: PgDn past the deepest zoom left zoom %d", c.age, v.shownZoom())
		}
		for i := len(c.want) - 2; i >= 0; i-- {
			press(tcell.KeyPgUp)
			if v.shownZoom() != c.want[i] {
				t.Fatalf("%s: PgUp gave zoom %d, want %d", c.age, v.shownZoom(), c.want[i])
			}
		}
	}
}

// TestInnerZoomDraws: the inner system draws, names itself, differs from the
// overview, and its cursor and panning work as they do at the overview.
func TestInnerZoomDraws(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "interstellar_age", Seed: 7})
	for _, size := range [][2]int{{120, 40}, {200, 55}} {
		v := New().(*view)
		f := Frame{Model: m}
		text := func() string {
			scr := drawFrame(v, f, size[0], size[1], false)
			defer scr.Fini()
			return capture.Text(scr)
		}
		v.zoomTo(v.scene(f), zOverview)
		over := text()
		v.zoomTo(v.scene(f), zInner)
		inner := text()
		if strings.TrimSpace(inner) == "" || inner == over {
			t.Fatalf("%v: inner view empty or equal to the overview", size)
		}
		if !strings.Contains(inner, "INNER") || !v.overview || v.ov == nil {
			t.Errorf("%v: inner is not a labelled squeezed view (overview %v)", size, v.overview)
		}
		if os.Getenv("DUMP_INNER") != "" {
			t.Logf("OVERVIEW %v\n%s\nINNER %v\n%s", size, over, size, inner)
		}
		at, moved := v.cur, false
		for _, k := range []tcell.Key{tcell.KeyRight, tcell.KeyLeft, tcell.KeyDown, tcell.KeyUp} {
			v.HandleKey(tcell.NewEventKey(k, 0, tcell.ModNone), f)
			moved = moved || v.cur != at // the arrows jump to the next body that way
		}
		if !moved {
			t.Errorf("%v: no arrow moved the cursor from %v", size, at)
		}
		if _, ok := v.Inspect(f); !ok {
			t.Errorf("%v: nothing to inspect at the inner zoom", size)
		}
	}
}
