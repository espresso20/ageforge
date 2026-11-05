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
	if !hasThing(v.scene(f).Grid, "space_station") {
		t.Error("the calm Earth lost its station")
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

// TestZoomIntoAStarAndBack: from COSMOS a star opens its system, a planet of
// it opens a close-up, and zooming out returns each time with the cursor on
// the body it left, ending on the star in COSMOS.
func TestZoomIntoAStarAndBack(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	v.zoomTo(v.scene(f), zCosmos)
	shot(v, f).Fini()
	var star mapgen.Thing
	for _, b := range bodiesOf(v.ov.Grid) {
		if b.th.Key == "star" {
			star = b.th
			v.cur = v.ov.Pick(b.at.X, b.at.Y)
			break
		}
	}
	if star.Key == "" {
		t.Fatal("no star to zoom into at COSMOS")
	}
	cosmosCur := v.cur
	press(v, f, tcell.KeyPgDn)
	shot(v, f).Fini()
	if v.star != star.Ref || v.shownZoom() != zOverview || bodyUnder(v, f) != star {
		t.Fatalf("into the star: star %d zoom %d on %v", v.star, v.shownZoom(), bodyUnder(v, f))
	}
	press(v, f, tcell.KeyPgDn) // inner
	shot(v, f).Fini()
	var planet mapgen.Thing
	for _, b := range bodiesOf(v.ov.Grid) {
		if b.th.Key == "planet" {
			planet = b.th
			v.cur = v.ov.Pick(b.at.X, b.at.Y)
			break
		}
	}
	if planet.Key == "" {
		t.Fatal("the system shows no planet")
	}
	press(v, f, tcell.KeyPgDn) // close-up
	shot(v, f).Fini()
	if v.planet != planet.Ref || v.star != star.Ref {
		t.Fatalf("close-up of star %d planet %d, want %d / %d", v.star, v.planet, star.Ref, planet.Ref)
	}
	press(v, f, tcell.KeyPgDn) // the district: refused
	shot(v, f).Fini()
	if v.shownZoom() != zCloseUp {
		t.Errorf("a far system opened zoom %d", v.shownZoom())
	}
	press(v, f, tcell.KeyPgUp)
	shot(v, f).Fini()
	if v.shownZoom() != zInner || bodyUnder(v, f) != planet {
		t.Errorf("out of the close-up: zoom %d on %v, want the planet", v.shownZoom(), bodyUnder(v, f))
	}
	press(v, f, tcell.KeyPgUp)
	press(v, f, tcell.KeyPgUp)
	shot(v, f).Fini()
	if v.shownZoom() != zCosmos || v.star != 0 || bodyUnder(v, f) != star || v.cur != cosmosCur {
		t.Errorf("back at COSMOS: zoom %d star %d on %v at %v, want %v at %v", v.shownZoom(), v.star, bodyUnder(v, f), v.cur, star, cosmosCur)
	}
}

func TestFarStarInspects(t *testing.T) {
	v, f := spaceView(t, "interstellar_age")
	v.zoomTo(v.scene(f), zCosmos)
	shot(v, f).Fini()
	for _, b := range bodiesOf(v.ov.Grid) {
		if b.th.Key == "star" {
			v.cur = v.ov.Pick(b.at.X, b.at.Y)
			in, ok := v.Inspect(f)
			if !ok || in.Title == "" {
				t.Fatalf("a star at COSMOS inspects as %+v", in)
			}
			return
		}
	}
	t.Fatal("no star at COSMOS")
}
