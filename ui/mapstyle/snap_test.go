package mapstyle

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

var (
	earthThing = mapgen.PlanetThing(2)
	marsThing  = mapgen.PlanetThing(3)
)

func spaceView(t *testing.T, age string) (*view, Frame) {
	t.Helper()
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: age, Seed: 7})
	return New().(*view), Frame{Model: m}
}

func press(v *view, f Frame, k tcell.Key) { v.HandleKey(tcell.NewEventKey(k, 0, tcell.ModNone), f) }

func shot(v *view, f Frame) tcell.SimulationScreen { return drawFrame(v, f, 140, 45, false) }

// onBody puts the cursor on a body of the overview at the given zoom.
func onBody(t *testing.T, v *view, f Frame, zoom int, th mapgen.Thing) {
	t.Helper()
	v.zoomTo(v.scene(f), zoom)
	shot(v, f).Fini()
	for _, b := range bodiesOf(v.ov.Grid) {
		if b.th == th {
			v.cur = v.ov.Pick(b.at.X, b.at.Y)
			return
		}
	}
	t.Fatalf("no %v in the overview", th)
}

// bodyUnder is the thing the cursor stands on in the shown grid.
func bodyUnder(v *view, f Frame) mapgen.Thing {
	g, x, y, _ := v.shownCell(v.scene(f))
	b, _ := bodyAt(g, x, y, nil)
	return b.th
}

func TestSpaceArrowsJumpBetweenBodies(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	onBody(t, v, f, zInner, sunThing)
	shot(v, f).Fini()
	if got := bodyUnder(v, f); got != sunThing {
		t.Fatalf("start on %v, want the sun", got)
	}
	g, x, _, _ := v.shownCell(v.scene(f))
	_ = g
	press(v, f, tcell.KeyRight)
	shot(v, f).Fini()
	g2, x2, _, _ := v.shownCell(v.scene(f))
	_ = g2
	if x2 <= x {
		t.Errorf("right moved the cursor from column %d to %d", x, x2)
	}
	if got := bodyUnder(v, f); got == sunThing || got == (mapgen.Thing{}) {
		t.Errorf("right landed on %v, want another body", got)
	}
}

func TestSpaceArrowIntoNothingStays(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	onBody(t, v, f, zInner, sunThing)
	shot(v, f).Fini()
	for i := 0; i < 20; i++ { // run off the last body to the right
		press(v, f, tcell.KeyRight)
		shot(v, f).Fini()
	}
	end, th := v.cur, bodyUnder(v, f)
	press(v, f, tcell.KeyRight)
	shot(v, f).Fini()
	if v.cur != end || bodyUnder(v, f) != th {
		t.Errorf("right with nothing that way moved the cursor from %v to %v", end, v.cur)
	}
}

func TestBracketsSurroundTheBodyAndNotItsPixels(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	onBody(t, v, f, zInner, marsThing)
	v.inspect = false
	bare := shot(v, f)
	defer bare.Fini()
	v.inspect = true
	marked := shot(v, f)
	defer marked.Fini()
	a, _, _ := bare.GetContents()
	b, _, _ := marked.GetContents()
	corners, changed := 0, 0
	for i := 1 * 140; i < 41*140; i++ { // the picture, inside the bars
		if a[i].Runes[0] == b[i].Runes[0] {
			continue
		}
		changed++
		switch b[i].Runes[0] {
		case '┌', '┐', '└', '┘':
			corners++
		}
	}
	if corners == 0 || corners > 4 || changed != corners {
		t.Errorf("the cursor changed %d cells, %d of them corners; want only the corners (1 to 4)", changed, corners)
	}
	var sunBox mapgen.Rect
	for _, bd := range bodiesOf(v.ov.Grid) {
		if bd.th == marsThing {
			sunBox = bd.box
		}
	}
	if sunBox.W == 0 {
		t.Fatal("Mars is not in the overview")
	}
}

func reds(scr tcell.SimulationScreen) (n int) {
	cells, w, _ := scr.GetContents()
	_ = w
	for _, c := range cells {
		fg, bg, _ := c.Style.Decompose()
		for _, col := range []tcell.Color{fg, bg} {
			r, g, b := col.RGB()
			if r > g+40 && r > b+40 {
				n++
			}
		}
	}
	return n
}

func hasThing(g *mapgen.Grid, key string) bool {
	for _, th := range g.Things {
		if th.Key == key {
			return true
		}
	}
	return false
}

func TestZoomOnMarsDrawsMarsWithNoStation(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	onBody(t, v, f, zInner, marsThing)
	press(v, f, tcell.KeyPgDn)
	mars := shot(v, f)
	defer mars.Fini()
	if v.shownZoom() != zCloseUp {
		t.Fatalf("zoom %d, want the close-up", v.shownZoom())
	}
	if hasThing(v.scene(f).Grid, "space_station") {
		t.Error("Mars has a station")
	}
	onBody(t, v, f, zInner, earthThing)
	press(v, f, tcell.KeyPgDn)
	earth := shot(v, f)
	defer earth.Fini()
	if !hasThing(v.scene(f).Grid, "home") {
		t.Error("Earth's close-up lost the home world")
	}
	if reds(mars) <= reds(earth) {
		t.Errorf("Mars is no redder than Earth: %d against %d", reds(mars), reds(earth))
	}
	// a second zoom in on another planet goes nowhere
	onBody(t, v, f, zInner, marsThing)
	press(v, f, tcell.KeyPgDn)
	press(v, f, tcell.KeyPgDn)
	if v.shownZoom() != zCloseUp {
		t.Errorf("zoom %d after diving past Mars, want the close-up", v.shownZoom())
	}
}

func TestZoomOutKeepsTheCursorOnThePlanet(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	onBody(t, v, f, zInner, marsThing)
	press(v, f, tcell.KeyPgDn)
	shot(v, f).Fini()
	press(v, f, tcell.KeyPgUp)
	shot(v, f).Fini()
	if got := bodyUnder(v, f); got != marsThing {
		t.Errorf("back at inner the cursor is on %v, want Mars", got)
	}
}

func TestSunRefusesTheDive(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	onBody(t, v, f, zInner, sunThing)
	press(v, f, tcell.KeyPgDn)
	shot(v, f).Fini()
	if v.shownZoom() != zInner {
		t.Errorf("zoom %d on the sun, want it to stay at inner", v.shownZoom())
	}
}

func TestSpaceFrameIsTheSameForTheSameSeed(t *testing.T) {
	text := func() string {
		v, f := spaceView(t, "interstellar_age")
		onBody(t, v, f, zInner, marsThing)
		press(v, f, tcell.KeyPgDn)
		scr := shot(v, f)
		defer scr.Fini()
		return capture.Text(scr)
	}
	if text() != text() {
		t.Error("the same seed drew two different frames")
	}
}

// Off the space ages the arrows still step one cell.
func TestLandArrowsStillStepOneCell(t *testing.T) {
	v, f := spaceView(t, "bronze_age")
	shot(v, f).Fini()
	v.zoomTo(v.scene(f), zCloseUp)
	shot(v, f).Fini()
	from := v.cur
	press(v, f, tcell.KeyRight)
	if v.cur.X != from.X+1 || v.cur.Y != from.Y {
		t.Errorf("right moved %v to %v, want one cell", from, v.cur)
	}
}

// planetRefs is the planets (by Ref, 1 Mercury .. 8 Neptune) and whether the
// home world is among the bodies of g.
func planetRefs(g *mapgen.Grid) (refs map[int]bool, home bool) {
	refs = map[int]bool{}
	for _, b := range bodiesOf(g) {
		switch b.th.Key {
		case "planet":
			refs[b.th.Ref] = true
		case "home":
			home = true
		}
	}
	return refs, home
}

// TestInterstellarOpensOnThePort: a big panel opens the Interstellar Age on
// Earth and its space port, the modules of the ring inspect as the buildings,
// and zooming in again does nothing.
func TestInterstellarOpensOnThePort(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	scr := shot(v, f)
	scr.Fini()
	if v.shownZoom() != zCloseUp || v.overview {
		t.Fatalf("opened on zoom %d (overview %v), want the close-up", v.shownZoom(), v.overview)
	}
	g := v.scene(f).Grid
	if !hasThing(g, "home") {
		t.Error("no Earth in the close-up")
	}
	mod := -1
	for i, h := range g.Hit {
		if g.Things[h].Kind == mapgen.KindBuilding {
			mod = i
			break
		}
	}
	if mod < 0 {
		t.Fatal("no port module in the close-up")
	}
	v.cur = mapgen.Pt{X: mod % g.W, Y: mod / g.W}
	if in, _ := v.Inspect(f); in.Title == "" || in.Title == f.Model.TownName {
		t.Errorf("a port module inspects as %+v", in)
	}
	press(v, f, tcell.KeyPgDn)
	if v.shownZoom() != zCloseUp {
		t.Errorf("PgDn at the port gave zoom %d", v.shownZoom())
	}
}

// TestInterstellarLadder: out from the port, INNER, OVERVIEW, COSMOS, and back.
func TestInterstellarLadder(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	for _, want := range []int{zInner, zOverview, zCosmos, zCosmos} {
		press(v, f, tcell.KeyPgUp)
		shot(v, f).Fini()
		if v.shownZoom() != want {
			t.Fatalf("PgUp gave zoom %d, want %d", v.shownZoom(), want)
		}
	}
	for _, want := range []int{zOverview, zInner, zCloseUp, zCloseUp} {
		press(v, f, tcell.KeyPgDn)
		shot(v, f).Fini()
		if v.shownZoom() != want {
			t.Fatalf("PgDn gave zoom %d, want %d", v.shownZoom(), want)
		}
	}
}

// TestInnerShowsTheInnerPlanetsOnly: Mercury to Mars and the sun, not Jupiter.
func TestInnerShowsTheInnerPlanetsOnly(t *testing.T) {
	for _, size := range [][2]int{{100, 30}, {140, 45}, {200, 55}} {
		v, f := spaceView(t, "interstellar_age")
		v.zoomTo(v.scene(f), zInner)
		drawFrame(v, f, size[0], size[1], false).Fini()
		refs, home := planetRefs(v.ov.Grid)
		if !home || !refs[1] || !refs[2] || !refs[4] || len(refs) != 3 {
			t.Errorf("%v: inner shows planets %v, home %v; want Mercury, Venus, Mars and the home world", size, refs, home)
		}
		if !hasThing(v.ov.Grid, "sun") {
			t.Errorf("%v: no sun at INNER", size)
		}
	}
}

// TestOverviewShowsAllEight: the whole of Sol.
func TestOverviewShowsAllEight(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	v.zoomTo(v.scene(f), zOverview)
	drawFrame(v, f, 140, 45, false).Fini()
	refs, home := planetRefs(v.ov.Grid)
	if !home || len(refs) != 7 {
		t.Errorf("overview shows planets %v, home %v; want all eight", refs, home)
	}
}

// TestCosmosStarsAreNotPlaces: the outer system has no star to pick, and
// zooming in on the sky steps to the overview, as ever.
func TestCosmosStarsAreNotPlaces(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	v.zoomTo(v.scene(f), zCosmos)
	drawFrame(v, f, 140, 45, false).Fini()
	if hasThing(v.ov.Grid, "star") {
		t.Error("a star is a place at COSMOS")
	}
	if refs, home := planetRefs(v.ov.Grid); !home || len(refs) != 7 {
		t.Errorf("cosmos shows planets %v, home %v; want Sol small in the middle", refs, home)
	}
	g := v.ov.Grid
	for i, c := range g.Cells {
		if g.Hit[i] == 0 && c.R != ' ' && c.R != 0 { // a dot of sky
			v.cur = v.ov.Pick(i%g.W, i/g.W)
			break
		}
	}
	press(v, f, tcell.KeyPgDn)
	drawFrame(v, f, 140, 45, false).Fini()
	if v.shownZoom() != zOverview || v.planet != 0 {
		t.Errorf("zoom %d planet %d after zooming in on the sky", v.shownZoom(), v.planet)
	}
}
