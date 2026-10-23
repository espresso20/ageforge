package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// eachTheme runs fn once per registered theme, with that theme active, and
// puts the active theme back afterwards.
func eachTheme(t *testing.T, fn func(t *testing.T)) {
	t.Helper()
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		t.Run(th.Key, fn)
	}
}

// pinScene is the Ready Room drawn at w by h, with the highlight on the
// first offer and the animation frame at frame.
type pinScene struct {
	state game.GameState
	v     expView
	l     expLayout
	sc    *expScene
	g     *xgrid
	hi    int
	ox    int
	oy    int
}

func drawPinScene(t *testing.T, w, h, frame int) pinScene {
	t.Helper()
	engine, p := expFixture(t)
	p.open(engine.GetState())
	p.setSize(w, h)
	v := p.view
	v.frame = frame
	state := p.state
	l := layoutFor(w, h)
	sc := buildScene(state, l.mapW-2, l.mainH-2)
	if len(sc.offers) != 9 {
		t.Fatalf("the board has %d pins, want 9", len(sc.offers))
	}
	return pinScene{state: state, v: v, l: l, sc: sc, g: renderExpedition(state, v, w, h),
		hi: sc.selected(v.sel), ox: 2, oy: l.mainY + 1}
}

// cellAt is the grid cell over a map cell of the scene.
func (s pinScene) cellAt(p pt) xcell { return s.g.at(s.ox+p.x, s.oy+p.y) }

// TestPinChipsAreThreeWideWithTheRightRoles checks every offer's pin is a
// chip three cells wide, its glyph in the middle, in the role of its tier,
// in every theme.
func TestPinChipsAreThreeWideWithTheRightRoles(t *testing.T) {
	eachTheme(t, func(t *testing.T) {
		s := drawPinScene(t, 100, 30, 0)
		pal := newExpPalette()
		roles := map[chipTone]theme.Role{toneShort: theme.RolePositive, toneLong: theme.RoleLabel,
			toneMil: theme.RoleNegative, toneCivic: theme.RoleAccent, toneGoal: theme.RolePositive}
		for i, o := range s.sc.offers {
			if !o.placed {
				t.Fatalf("offer %d found no room", i)
			}
			tone := o.tone()
			want := chipKind(tone, i == s.hi && pulseBright(s.v.frame))
			for dx := -1; dx <= 1; dx++ {
				c := s.g.at(s.ox+o.at.x+dx, s.oy+o.at.y)
				if c.k != want {
					t.Errorf("offer %d (%s): cell %+d is kind %d, want the chip %d", i, o.offer.Name, dx, c.k, want)
				}
				if dx == 0 && c.r != o.glyph() {
					t.Errorf("offer %d: the chip's middle is %q, want %q", i, c.r, o.glyph())
				}
				if dx != 0 && c.r != ' ' {
					t.Errorf("offer %d: the chip's side is %q, want a blank", i, c.r)
				}
			}
			if i == s.hi {
				continue // the bright phase has no role of its own
			}
			_, bgc, _ := pal.styles[want].Decompose()
			if tone == toneLegend {
				gold := theme.Mix(theme.Color(theme.RoleHighlight), theme.MapHueColor(theme.HueGold), 0.6)
				if bgc != gold {
					t.Errorf("legendary chip is %v, want the gold role %v", bgc, gold)
				}
				continue
			}
			if bgc != theme.Color(roles[tone]) {
				t.Errorf("%s chip is %v, want the %s role %v", o.offer.Tier, bgc, roles[tone], theme.Color(roles[tone]))
			}
		}
	})
}

// TestPinMapInksReadInEveryTheme holds the chips, routes, names and the
// selection to 4.5:1 on what they sit on, and the terrain inks to the map
// floor, in every theme.
func TestPinMapInksReadInEveryTheme(t *testing.T) {
	eachTheme(t, func(t *testing.T) {
		pal := newExpPalette()
		bg := theme.Color(theme.RoleBackground)
		for k := xChipShort; k <= xChipGoalHi; k++ {
			fg, chip, _ := pal.styles[k].Decompose()
			if r := theme.ContrastRatio(fg, chip); r < 4.5 {
				t.Errorf("chip kind %d reads %.2f:1, want >= 4.5", k, r)
			}
		}
		for _, k := range []xk{xRoute, xPlace, xGood, xBright} {
			fg, _, _ := pal.styles[k].Decompose()
			if r := theme.ContrastRatio(fg, bg); r < 4.5 {
				t.Errorf("kind %d reads %.2f:1 on the map, want >= 4.5", k, r)
			}
		}
		if fg, sel, _ := pal.styles[xSel].Decompose(); theme.ContrastRatio(fg, sel) < 4.5 {
			t.Errorf("the selected label reads %.2f:1, want >= 4.5", theme.ContrastRatio(fg, sel))
		}
		for _, k := range []xk{xWater, xGrass, xHill, xDust} {
			fg, _, _ := pal.styles[k].Decompose()
			if r := theme.ContrastRatio(fg, bg); r < mapstyle.TerrainFloor {
				t.Errorf("terrain kind %d reads %.2f:1 on the map, want >= %.1f", k, r, mapstyle.TerrainFloor)
			}
		}
	})
}

// TestPinHaloLeavesTerrainClear checks that no terrain is drawn on the cells
// next to a pin or the town, and that the terrain is drawn everywhere else.
func TestPinHaloLeavesTerrainClear(t *testing.T) {
	for _, sz := range [][2]int{{100, 30}, {80, 24}, {120, 40}} {
		s := drawPinScene(t, sz[0], sz[1], 0)
		t0 := s.sc.terrain
		ring := []pt{s.sc.town}
		for _, it := range append(append([]expPinItem{}, s.sc.acts...), s.sc.offers...) {
			if it.placed {
				ring = append(ring, it.at)
			}
		}
		for _, c := range ring {
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					x, y := c.x+dx, c.y+dy
					if x < 0 || y < 0 || x >= t0.w || y >= t0.h {
						continue
					}
					if !t0.halo[y*t0.w+x] {
						t.Errorf("%dx%d: %v is next to a pin or the town but not in the halo", sz[0], sz[1], pt{x, y})
					}
				}
			}
		}
		drawn := 0
		for y := 0; y < t0.h; y++ {
			for x := 0; x < t0.w; x++ {
				k := s.cellAt(pt{x, y}).k
				if t0.halo[y*t0.w+x] {
					switch k {
					case xWater, xHill, xGrass, xDust:
						t.Errorf("%dx%d: terrain drawn in the halo at %v", sz[0], sz[1], pt{x, y})
					}
					continue
				}
				if k == xWater || k == xHill || k == xGrass || k == xDust {
					drawn++
				}
			}
		}
		if drawn == 0 {
			t.Errorf("%dx%d: no terrain drawn outside the halo", sz[0], sz[1])
		}
	}
}

// TestPinNamesSitBesideTheirChips checks every pin's place name is drawn
// beside its chip in the dim role, the selected one bright on the selection,
// and that no name lies on a chip, the town or another name.
func TestPinNamesSitBesideTheirChips(t *testing.T) {
	for _, sz := range [][2]int{{100, 30}, {80, 24}, {120, 40}} {
		s := drawPinScene(t, sz[0], sz[1], 0)
		w := s.l.mapW - 2
		row := func(y int) []xcell {
			out := make([]xcell, w)
			for x := 0; x < w; x++ {
				out[x] = s.cellAt(pt{x, y})
			}
			return out
		}
		var named, skipped int
		for i, o := range s.sc.offers {
			if !o.placed {
				continue
			}
			cells := row(o.at.y)
			runes := make([]rune, w)
			for x, c := range cells {
				runes[x] = c.r
			}
			idx := strings.Index(string(runes), o.offer.Place)
			if idx < 0 {
				skipped++
				continue
			}
			named++
			want := xPlace
			if i == s.hi {
				want = xSel
			}
			start := len([]rune(string(runes)[:idx]))
			for x := start; x < start+len([]rune(o.offer.Place)); x++ {
				if cells[x].k != want {
					t.Errorf("%dx%d: %q cell %d is kind %d, want %d", sz[0], sz[1], o.offer.Place, x, cells[x].k, want)
				}
			}
		}
		if named+skipped != 9 {
			t.Errorf("%dx%d: %d names and %d skipped, want 9 pins", sz[0], sz[1], named, skipped)
		}
		if skipped > 0 {
			t.Logf("%dx%d: %d of 9 names skipped for want of room", sz[0], sz[1], skipped)
		}
		if s.hi >= 0 && s.sc.offers[s.hi].placed && !strings.Contains(s.g.String(), s.sc.offers[s.hi].offer.Place) {
			t.Errorf("%dx%d: the selected pin's name is not drawn", sz[0], sz[1])
		}
	}
}

// TestSelectedChipPulsesAndHoldsStillWithMotionOff checks the highlighted
// chip alternates between its two phases on the redraw tick, sits in its
// bright phase when the motion is off, and the others never change.
func TestSelectedChipPulsesAndHoldsStillWithMotionOff(t *testing.T) {
	selected := func(frame int) xk {
		s := drawPinScene(t, 100, 30, frame)
		o := s.sc.offers[s.hi]
		return s.cellAt(o.at).k
	}
	tone := drawPinScene(t, 100, 30, 0).sc.offers[0].tone()
	if got, want := selected(0), chipKind(tone, true); got != want {
		// the first offer is the highlighted one, so it is the selected pin
		t.Errorf("frame 0: selected chip kind %d, want %d", got, want)
	}
	if got, want := selected(1), chipKind(tone, false); got != want {
		t.Errorf("frame 1: selected chip kind %d, want %d", got, want)
	}
	if got, want := selected(atRest), chipKind(tone, true); got != want {
		t.Errorf("motion off: selected chip kind %d, want the bright phase %d", got, want)
	}
	s := drawPinScene(t, 100, 30, 1)
	for i, o := range s.sc.offers {
		if i == s.hi || !o.placed {
			continue
		}
		if got, want := s.cellAt(o.at).k, chipKind(o.tone(), false); got != want {
			t.Errorf("pin %d kind %d, want its steady chip %d", i, got, want)
		}
	}
}

// TestPinPlacementLeavesRoomForChips checks that, over many maps, every pin
// stands on land, off the town, and its three-cell chip overlaps no other
// chip or the town; pins in one row stand at least four columns apart.
func TestPinPlacementLeavesRoomForChips(t *testing.T) {
	tiers := []string{game.TierShort, game.TierLong, game.TierGrand}
	for _, age := range []string{"stone_age", "bronze_age", "modern_age"} {
		for seed := int64(1); seed <= 25; seed++ {
			for _, sz := range [][2]int{{60, 19}, {74, 26}, {100, 28}} {
				w, h := sz[0], sz[1]
				terrain := buildTerrain(seed, age, w, h)
				town := townAt(w, h)
				var specs []pinSpec
				for k := 0; k < 9; k++ {
					specs = append(specs, pinSpec{pinHash(fmt.Sprintf("offer %d", k), int(seed)*100+k), tiers[k%3]})
				}
				at, ok := placePins(terrain, specs)
				taken := map[pt]int{town: -1}
				for i, p := range at {
					if !ok[i] {
						continue
					}
					if terrain.isWater(p.x, p.y) || p == town {
						t.Errorf("%s seed %d %dx%d: pin %d on water or the town at %v", age, seed, w, h, i, p)
					}
					for dx := -1; dx <= 1; dx++ {
						c := pt{p.x + dx, p.y}
						if c.x < 0 || c.x >= w {
							t.Errorf("%s seed %d %dx%d: pin %d's chip leaves the map at %v", age, seed, w, h, i, c)
							continue
						}
						if j, dup := taken[c]; dup {
							t.Errorf("%s seed %d %dx%d: pin %d's chip overlaps %d at %v", age, seed, w, h, i, j, c)
						}
						taken[c] = i
					}
					for j, q := range at {
						if j > i && ok[j] && q.y == p.y && abs(q.x-p.x) < 4 {
							t.Errorf("%s seed %d %dx%d: pins %d and %d are %d columns apart", age, seed, w, h, i, j, abs(q.x-p.x))
						}
					}
				}
			}
		}
	}
}

// TestMissionUnderWayShowsItsGoalAsAChip checks a party out is drawn in the
// Report: its goal is a chip, its name beside it, and the highlighted party's
// goal is in the bright phase on the selected pin.
func TestMissionUnderWayShowsItsGoalAsAChip(t *testing.T) {
	engine, p := expFixture(t)
	p.open(engine.GetState())
	want := *p.scene().offers[p.scene().selected(p.view.sel)].offer
	if !want.CanLaunch {
		t.Fatalf("the first offer cannot be sent: %s", want.LaunchBlockReason)
	}
	p.handleKey(tcell.KeyEnter)
	p.handleKey(tcell.KeyEnter)
	p.setReport(true)
	p.view.frame = 0
	state := p.state
	l := reportLayoutFor(100, 30)
	sc := buildScene(state, l.mapW-2, l.mainH-2)
	if len(sc.acts) != 1 || !sc.acts[0].placed {
		t.Fatalf("want one party out and placed, have %d", len(sc.acts))
	}
	g := renderExpedition(state, p.view, 100, 30)
	goal := sc.acts[0].at
	k := g.at(2+goal.x, l.mainY+1+goal.y).k
	if want := chipKind(toneGoal, true); k != want {
		t.Errorf("the highlighted party's goal is kind %d, want the bright goal chip %d", k, want)
	}
	if !strings.Contains(g.String(), sc.acts[0].act.Place) {
		t.Errorf("the party's place name is not on the map")
	}
}
