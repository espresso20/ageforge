package all

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// mapText draws every view of every style for a state (full views at two
// sizes, the region zoom, the mini map at its old and new sizes, and each
// style's cursor stepped through every target) and returns everything a
// player could read: the screens and the inspections.
func mapText(t *testing.T, st game.GameState) string {
	t.Helper()
	m := mapmodel.NewBuilder(nil).Build(&st, nil)
	reg := Registry()
	var b strings.Builder
	f := mapstyle.Frame{Model: m, Tier: mapmodel.TierUnicode}
	for _, name := range reg.Names() {
		for _, world := range []bool{false, true} {
			v, _ := reg.New(name)
			v.SetOption(mapstyle.OptWorld, world)
			v.SetOption(mapstyle.OptInspect, true)
			v.SetOption(mapstyle.OptFlows, true)
			for _, sz := range [][2]int{{160, 48}, {120, 40}} {
				scr := capture.NewScreen(sz[0], sz[1])
				v.Draw(scr, mapstyle.Rect{W: sz[0], H: sz[1]}, f)
				b.WriteString(capture.Text(scr))
				scr.Fini()
			}
			scr := capture.NewScreen(160, 48)
			for i := 0; i < 80; i++ {
				v.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), f)
				v.Draw(scr, mapstyle.Rect{W: 160, H: 48}, f)
				if in, ok := v.Inspect(f); ok {
					b.WriteString(in.Title + "\n" + strings.Join(in.Lines, "\n") + "\n" + in.Command + "\n")
				}
				if i%10 == 0 {
					b.WriteString(capture.Text(scr))
				}
			}
			scr.Fini()
		}
		for _, sz := range [][2]int{{40, 15}, {65, 9}} {
			v, _ := reg.New(name)
			scr := capture.NewScreen(sz[0], sz[1])
			v.DrawCompact(scr, mapstyle.Rect{W: sz[0], H: sz[1]}, f)
			b.WriteString(capture.Text(scr))
			scr.Fini()
		}
	}
	return b.String()
}

// unreached lists the names a state must not show, by the game's one
// no-spoiler rule (game/spoilers.go): every civilization not yet met (name
// and key), every era not reached, and every age past those the player may
// see named.
func unreached(st game.GameState) []string {
	sight := game.SightOf(&st)
	var out []string
	for _, f := range config.BaseFactions() {
		if fi, ok := st.Diplomacy.Factions[f.Key]; !ok || !fi.Discovered {
			out = append(out, f.Name, f.Key)
		}
	}
	for _, e := range config.Epochs() {
		if !sight.Era(e.Key) {
			out = append(out, e.Name)
		}
	}
	for _, a := range config.Ages() {
		if !sight.Age(a.Key) {
			out = append(out, a.Name)
		}
	}
	return out
}

// TestNoSpoilers: the maps name no civilization the player has not met and
// no age or era not yet reached: not on screen, not on the inspect line,
// not in the mini map. Adam asked for no spoilers after his first playtest.
func TestNoSpoilers(t *testing.T) {
	cases := []struct {
		name string
		st   func() game.GameState
	}{
		{"a new game in the Primitive Age", func() game.GameState {
			st := fixture.State(fixture.Options{Age: "primitive_age", Seed: 7, Harbinger: true, Idle: 4})
			for k, f := range st.Diplomacy.Factions {
				f.Discovered = false
				st.Diplomacy.Factions[k] = f
			}
			st.Harbinger.TargetEpochName = "impending doom" // as the game words it
			return st
		}},
		{"the Primitive Age, ready to advance", func() game.GameState {
			st := fixture.State(fixture.Options{Age: "primitive_age", Seed: 3, Harbinger: true})
			for k, f := range st.Diplomacy.Factions {
				f.Discovered = false
				st.Diplomacy.Factions[k] = f
			}
			st.AgeReady = true
			return st
		}},
		{"the Medieval Age with two civs met", func() game.GameState {
			st := fixture.State(fixture.Options{Age: "medieval_age", Seed: 7, Harbinger: true, Wars: 1, Routes: 2})
			met := 0
			for _, f := range config.BaseFactions() {
				fi := st.Diplomacy.Factions[f.Key]
				fi.Discovered = met < 2 && fi.Discovered
				if fi.Discovered {
					met++
				} else {
					fi.AtWar = false
				}
				st.Diplomacy.Factions[f.Key] = fi
			}
			return st
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := c.st()
			txt := mapText(t, st)
			for _, word := range unreached(st) {
				if i := strings.Index(txt, word); i >= 0 {
					lo, hi := max(0, i-60), min(len(txt), i+len(word)+60)
					t.Errorf("the map shows %q:\n...%s...", word, txt[lo:hi])
				}
			}
			if st.AgeReady && !strings.Contains(txt, st.NextAgeName) {
				t.Errorf("ready to advance, but the map never names the next age %q", st.NextAgeName)
			}
			if h := st.Harbinger; h != nil && !strings.Contains(txt, "warns of "+h.TargetEpochName) {
				t.Errorf("the harbinger's line does not say it warns of %q", h.TargetEpochName)
			}
		})
	}
}
