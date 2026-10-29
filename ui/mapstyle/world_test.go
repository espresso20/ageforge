package mapstyle

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// worldAges are an age of each scale.
var worldAges = []string{"medieval_age", "modern_age", "space_age", "interstellar_age", "galactic_age"}

// busyModel is a town with a civ at war, routes, a harbinger, idle workers and
// both kinds of party out.
func busyModel(t testing.TB, age string) *mapmodel.Model {
	t.Helper()
	st := fixture.State(fixture.Options{Age: age, Seed: 6, Tick: noon, Wars: 1, Harbinger: true, Idle: 14, Routes: 4})
	st.Military.ActiveScout = &game.ExpeditionSnapshot{Name: "Scout the hills", TicksLeft: 40}
	st.Military.ActiveMilitary = &game.ExpeditionSnapshot{Name: "March on the raiders", TicksLeft: 25}
	return testBuilder.Build(&st, nil)
}

// firstCell is the first cell of a scene that shows a thing of a kind.
func firstCell(g *mapgen.Grid, kind mapgen.Kind, key string) (mapgen.Pt, bool) {
	for i, h := range g.Hit {
		if th := g.Things[h]; th.Kind == kind && (key == "" || th.Key == key) {
			return mapgen.Pt{X: i % g.W, Y: i / g.W}, true
		}
	}
	return mapgen.Pt{}, false
}

// TestWorldRoundTheTownInspectsAsItself: the cursor on a civ, a route, the
// harbinger, a party that is out and the idle workers says what each is, in
// the game's names, with the command to type for it.
func TestWorldRoundTheTownInspectsAsItself(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range worldAges {
		m := busyModel(t, age)
		v := New().(*view)
		lane := m.AgeIdx >= m.Catalog.SpaceAge()
		var civ *mapmodel.Faction
		for i := range m.Factions {
			if m.Factions[i].Discovered {
				civ = &m.Factions[i]
				break
			}
		}
		routeCiv := ""
		for _, rt := range m.Routes {
			if rt.Civ != "" {
				routeCiv = rt.Civ
				break
			}
		}
		cases := []struct {
			what  string
			kind  mapgen.Kind
			key   string
			title string
			cmd   string
		}{
			{"civ", mapgen.KindCiv, civ.Key, civ.Name, m.FactionCommand(civ)},
			{"route", mapgen.KindRoute, routeCiv, map[bool]string{false: "Trail to the ", true: "Lane to the "}[lane] + m.Faction(routeCiv).Name, "trade route list"},
			{"harbinger", mapgen.KindHarbinger, "", m.Harbinger.Name, "harbinger"},
			{"scouts", mapgen.KindMission, "scout", "Scouts: Scout the hills", "expedition"},
			{"army", mapgen.KindMission, "military", "Army: March on the raiders", "army"},
			{"idle workers", mapgen.KindIdle, "", "Idle workers", m.IdleCommand()},
		}
		if age == "galactic_age" { // the star map: a lane to the civ, no parties or loiterers
			cases = append(cases[:1:1], cases[2], struct {
				what  string
				kind  mapgen.Kind
				key   string
				title string
				cmd   string
			}{"lane", mapgen.KindLane, routeCiv, "Lane: " + m.TownName + " to " + m.Faction(routeCiv).Name, ""})
		}
		for _, c := range cases {
			if c.what == "civ" {
				c.title = civ.Name
			}
			// a party can pass behind a building, so look through some frames
			var p mapgen.Pt
			var ok bool
			var f Frame
			for anim := 0; anim < 60 && !ok; anim += 5 {
				f = Frame{Model: m, Anim: anim, Clock: anim}
				p, ok = firstCell(v.scene(f).Grid, c.kind, c.key)
			}
			if !ok {
				t.Errorf("%s: no %s on the map", age, c.what)
				continue
			}
			v.cur = p
			in, _ := v.Inspect(f)
			if in.Title != c.title || in.Command != c.cmd || len(in.Lines) == 0 {
				t.Errorf("%s: the %s inspects as %+v, want title %q and command %q", age, c.what, in, c.title, c.cmd)
			}
		}
	}
}

// TestTheCursorCanVisitACiv: Tab goes round the buildings and wonders and on
// to the civs at the edge.
func TestTheCursorCanVisitACiv(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range worldAges {
		m := busyModel(t, age)
		v := New()
		f := Frame{Model: m}
		draw(v, m, 120, 36, false, mapmodel.TierUnicode).Fini()
		seen := map[string]bool{}
		for i := 0; i < 400 && len(seen) < m.DiscoveredCount(); i++ {
			v.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), f)
			if in, _ := v.Inspect(f); in.Command != "" {
				for _, fc := range m.Factions {
					if fc.Discovered && in.Title == fc.Name {
						seen[fc.Key] = true
					}
				}
			}
		}
		if len(seen) != m.DiscoveredCount() {
			t.Errorf("%s: Tab visited %d of %d civs", age, len(seen), m.DiscoveredCount())
		}
	}
}

// TestWorldRoundTheTownDrawsAtEverySize: every age draws it in a big panel, a
// small one and the mini map, and in plain ASCII.
func TestWorldRoundTheTownDrawsAtEverySize(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range worldAges {
		m := busyModel(t, age)
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
			for _, sz := range sizes {
				// the cursor stands on a civ, so the close-up has the edge in view
				v := New().(*view)
				f := Frame{Model: m, Anim: 17, Clock: 17, Tier: tier}
				civ, _ := firstCell(v.scene(f).Grid, mapgen.KindCiv, "")
				v.cur, v.recentre = civ, true
				scr := drawFrame(v, f, sz.w, sz.h, sz.compact)
				cells, w, h := scr.GetContents()
				for i, c := range cells[:w*h] {
					if tier == mapmodel.TierASCII && len(c.Runes) > 0 && (c.Runes[0] > 0x7e || c.Runes[0] == '?') {
						t.Fatalf("%s %s: %q at (%d,%d) is not clean ASCII", age, sz.name, c.Runes[0], i%w, i/w)
					}
				}
				if sz.name != "mini" && tier == mapmodel.TierUnicode && !strings.ContainsAny(capture.Text(scr), "▪✦") {
					t.Errorf("%s %s: no civ marker on the screen", age, sz.name)
				}
				scr.Fini()
			}
		}
	}
}

// TestWorldRoundTheTownNeverCoversTheCursor: with the cursor on a civ, the
// harbinger or the idle workers, the cell keeps the glyph and wears the
// cursor's highlight over it.
func TestWorldRoundTheTownNeverCoversTheCursor(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := busyModel(t, "modern_age")
	f := Frame{Model: m, Anim: 17, Clock: 17}
	for _, kind := range []mapgen.Kind{mapgen.KindCiv, mapgen.KindHarbinger, mapgen.KindIdle} {
		target, ok := firstCell(New().(*view).scene(f).Grid, kind, "")
		if !ok {
			t.Fatalf("nothing of kind %d to put the cursor on", kind)
		}
		draw := func(cursor bool) (rune, tcell.Style) {
			v := New().(*view)
			v.scene(f)
			v.cur, v.recentre, v.inspect = target, true, cursor
			scr := drawFrame(v, f, 120, 36, false)
			defer scr.Fini()
			cells, _, _ := scr.GetContents()
			c := cells[(1+target.Y-v.cam.Y)*120+target.X-v.cam.X]
			return c.Runes[0], c.Style
		}
		bare, bareStyle := draw(false)
		marked, markedStyle := draw(true)
		if bare != marked || bareStyle == markedStyle {
			t.Errorf("kind %d: the cursor turned %q into %q or does not show", kind, bare, marked)
		}
	}
}
