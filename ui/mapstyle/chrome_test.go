package mapstyle

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
	"github.com/espresso20/ageforge/vibe"
)

// cursorCells returns the cells of a screen drawn in the cursor's fill
// (either half of its blink), with the position of each.
func cursorCells(scr tcell.SimulationScreen, ink *Palette) (cells []tcell.SimCell, at [][2]int) {
	_, fill, _ := ink.Cursor.Decompose()
	_, alt, _ := ink.CursorAlt.Decompose()
	all, w, h := scr.GetContents()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := all[y*w+x]
			if _, bg, _ := cell.Style.Decompose(); bg == fill || bg == alt {
				cells = append(cells, cell)
				at = append(at, [2]int{x, y})
			}
		}
	}
	return cells, at
}

// changedCells returns the cells of the rows from top to bottom whose style
// differs between two drawings of one screen, with the position of each.
func changedCells(before, after tcell.SimulationScreen, top, bottom int) (cells []tcell.SimCell, at [][2]int) {
	was, w, _ := before.GetContents()
	now, _, _ := after.GetContents()
	for i := top * w; i < bottom*w; i++ {
		if was[i].Style != now[i].Style {
			cells = append(cells, now[i])
			at = append(at, [2]int{i % w, i / w})
		}
	}
	return cells, at
}

// TestCursorIsAHighlightedCell: in every theme and at both zooms, the cursor
// is one highlighted cell, which no other cell of the map is drawn in; the
// glyph beneath it is kept; and the fill stands clear of the cell beneath it,
// whatever ground the engine laid there.
func TestCursorIsAHighlightedCell(t *testing.T) {
	for th := range fixture.Themes(t) {
		for _, age := range []string{"stone_age", "medieval_age", "modern_age", "cyberpunk_age", "space_age", "galactic_age"} {
			m := grown(t, fixture.Options{Age: age, Seed: 5})
			for _, world := range []bool{false, true} {
				v := New().(*view)
				v.SetOption(OptWorld, world)
				v.SetOption(OptInspect, false)
				plain := drawOffset(v, m, 120, 40, 0, mapmodel.TierUnicode, false)
				if cells, _ := cursorCells(plain, v.ink.Palette); len(cells) != 0 {
					t.Errorf("%s, %s: %d cells are drawn in the cursor's fill with no cursor out", th.Key, age, len(cells))
				}
				v.SetOption(OptInspect, true)
				lit := drawOffset(v, m, 120, 40, 0, mapmodel.TierUnicode, false)
				cells, at := changedCells(plain, lit, 2, 40-3) // the rows of the picture, inside the bars
				before, w, _ := plain.GetContents()
				plain.Fini()
				lit.Fini()
				if len(cells) != 1 {
					t.Errorf("%s, %s world %v: the cursor changes %d cells, want 1", th.Key, age, world, len(cells))
					continue
				}
				under := before[at[0][1]*w+at[0][0]]
				if len(cells[0].Runes) == 0 || len(under.Runes) == 0 || cells[0].Runes[0] != under.Runes[0] {
					t.Errorf("%s, %s world %v: the cursor hides the glyph beneath it", th.Key, age, world)
				}
				_, ground, _ := under.Style.Decompose()
				_, fill, _ := cells[0].Style.Decompose()
				if d, cr := theme.Distance(fill, ground), theme.ContrastRatio(fill, ground); d < CursorFromGround || cr < CursorOverGround {
					t.Errorf("%s, %s world %v: the cursor #%06x over the cell beneath it #%06x: distance %.3f, contrast %.2f", th.Key, age, world, fill.Hex(), ground.Hex(), d, cr)
				}
			}
		}
	}
}

// TestCursorBlinksWithTheMap: the cursor blinks slowly while the map moves
// (the fill, then its brighter shade, each for CursorBlinkFrames
// frames) and holds the fill while the motion setting holds the frame at 0.
func TestCursorBlinksWithTheMap(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := grown(t, fixture.Options{Age: "stone_age", Seed: 5})
	v := New().(*view)
	bgAt := func(anim int) tcell.Color {
		scr := drawOffset(v, m, 120, 40, anim, mapmodel.TierUnicode, false)
		defer scr.Fini()
		cells, _ := cursorCells(scr, v.ink.Palette)
		if len(cells) == 0 {
			t.Fatalf("frame %d: no cursor", anim)
		}
		_, bg, _ := cells[0].Style.Decompose()
		return bg
	}
	_, fill, _ := v.begin(Frame{Model: m}).Cursor.Decompose()
	_, alt, _ := v.ink.CursorAlt.Decompose()
	if got := bgAt(0); got != fill {
		t.Errorf("with motion off the cursor is #%06x, want the fill #%06x", got.Hex(), fill.Hex())
	}
	if got := bgAt(CursorBlinkFrames); got != alt {
		t.Errorf("half a blink on the cursor is #%06x, want the brighter shade #%06x", got.Hex(), alt.Hex())
	}
	if got := bgAt(2 * CursorBlinkFrames); got != fill {
		t.Errorf("a whole blink on the cursor is #%06x, want the fill #%06x again", got.Hex(), fill.Hex())
	}
}

// TestMiniMapMarksTheCursor: the compact view marks the cell another view's
// cursor stands on with the same highlighted cell, and nothing when that
// cursor is not out.
func TestMiniMapMarksTheCursor(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := grown(t, fixture.Options{Age: "stone_age", Seed: 5})
	panel, mini := New(), New().(*view)
	drawOffset(panel, m, 120, 40, 0, mapmodel.TierUnicode, false).Fini() // places the cursor
	mini.Point(panel.Pointer())
	scr := drawOffset(mini, m, 58, 9, 0, mapmodel.TierUnicode, true)
	if cells, _ := cursorCells(scr, mini.ink.Palette); len(cells) != 1 {
		t.Errorf("the mini map marks %d cells for the Map panel's cursor, want 1", len(cells))
	}
	scr.Fini()
	panel.SetOption(OptInspect, false)
	mini.Point(panel.Pointer())
	scr = drawOffset(mini, m, 58, 9, 0, mapmodel.TierUnicode, true)
	if cells, _ := cursorCells(scr, mini.ink.Palette); len(cells) != 0 {
		t.Errorf("the mini map marks %d cells with no cursor out, want none", len(cells))
	}
	scr.Fini()
}

// TestChromeInksKeepTheTerrainFloor: the night epochs dim the land the chrome
// is drawn beside, and the dimmed terrain classes still read at the map floor
// on every theme.
func TestChromeInksKeepTheTerrainFloor(t *testing.T) {
	for th := range fixture.Themes(t) {
		var ink *inks
		for epoch := 0; epoch < 7; epoch++ {
			ink = inksFor(ink, epoch)
			for _, class := range TerrainClasses {
				if cr := theme.ContrastRatio(ink.Fg[class], ink.Bg); cr < TerrainFloor {
					t.Errorf("%s epoch %d: class %d reads %.2f:1, want >= %.1f", th.Key, epoch, class, cr, TerrainFloor)
				}
			}
		}
	}
}

// mapText draws every view of the map for a state (full views at two sizes,
// the whole-town zoom, the mini map at two sizes, and the cursor stepped
// through every target) and returns everything a player could read: the
// screens and the inspections.
func mapText(t *testing.T, st game.GameState) string {
	t.Helper()
	m := mapmodel.NewBuilder(nil).Build(&st, nil)
	var out strings.Builder
	f := Frame{Model: m, Tier: mapmodel.TierUnicode}
	for _, world := range []bool{false, true} {
		v := New()
		v.SetOption(OptWorld, world)
		v.SetOption(OptInspect, true)
		v.SetOption(OptFlows, true)
		for _, size := range [][2]int{{160, 48}, {120, 40}} {
			scr := capture.NewScreen(size[0], size[1])
			v.Draw(scr, Rect{W: size[0], H: size[1]}, f)
			out.WriteString(capture.Text(scr))
			scr.Fini()
		}
		scr := capture.NewScreen(160, 48)
		for i := 0; i < 80; i++ {
			v.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), f)
			v.Draw(scr, Rect{W: 160, H: 48}, f)
			if in, ok := v.Inspect(f); ok {
				out.WriteString(in.Title + "\n" + strings.Join(in.Lines, "\n") + "\n" + in.Command + "\n")
			}
			if i%10 == 0 {
				out.WriteString(capture.Text(scr))
			}
		}
		scr.Fini()
	}
	for _, size := range [][2]int{{40, 15}, {65, 9}} {
		scr := capture.NewScreen(size[0], size[1])
		New().DrawCompact(scr, Rect{W: size[0], H: size[1]}, f)
		out.WriteString(capture.Text(scr))
		scr.Fini()
	}
	return out.String()
}

// unreached lists the names a state must not show, by the game's one
// no-spoiler rule (game/spoilers.go): every civilization not yet met (name
// and key), every era not reached, and every age past those the player may
// see named.
func unreached(st game.GameState) []string {
	sight := game.SightOf(&st)
	var names []string
	for _, faction := range config.BaseFactions() {
		if info, ok := st.Diplomacy.Factions[faction.Key]; !ok || !info.Discovered {
			names = append(names, faction.Name, faction.Key)
		}
	}
	for _, epoch := range config.Epochs() {
		if !sight.Era(epoch.Key) {
			names = append(names, epoch.Name)
		}
	}
	for _, age := range config.Ages() {
		if !sight.Age(age.Key) {
			names = append(names, age.Name)
		}
	}
	return names
}

// meetOnly forgets every civilization but the first n of those met, in the
// game's order, and ends any war with the ones forgotten.
func meetOnly(st *game.GameState, n int) {
	met := 0
	for _, faction := range config.BaseFactions() {
		info := st.Diplomacy.Factions[faction.Key]
		if info.Discovered && met < n {
			met++
		} else {
			info.Discovered, info.AtWar = false, false
		}
		st.Diplomacy.Factions[faction.Key] = info
	}
}

// TestNoSpoilers: the map names no civilization the player has not met and no
// age or era not yet reached: not on screen, not on the inspect line, not in
// the mini map.
func TestNoSpoilers(t *testing.T) {
	fixture.WithTheme(t, "forge")
	cases := []struct {
		name  string
		state func() game.GameState
	}{
		{"a new game in the Primitive Age", func() game.GameState {
			st := fixture.State(fixture.Options{Age: "primitive_age", Seed: 7, Harbinger: true, Idle: 4})
			meetOnly(&st, 0)
			st.Harbinger.TargetEpochName = "impending doom" // as the game words it
			return st
		}},
		{"the Primitive Age, ready to advance", func() game.GameState {
			st := fixture.State(fixture.Options{Age: "primitive_age", Seed: 3, Harbinger: true})
			meetOnly(&st, 0)
			st.AgeReady = true
			return st
		}},
		{"the Medieval Age with two civs met", func() game.GameState {
			st := fixture.State(fixture.Options{Age: "medieval_age", Seed: 7, Harbinger: true, Wars: 1, Routes: 2})
			meetOnly(&st, 2)
			return st
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := c.state()
			text := mapText(t, st)
			for _, word := range unreached(st) {
				if i := strings.Index(text, word); i >= 0 {
					lo, hi := max(0, i-60), min(len(text), i+len(word)+60)
					t.Errorf("the map shows %q:\n...%s...", word, text[lo:hi])
				}
			}
			if st.AgeReady && !strings.Contains(text, st.NextAgeName) {
				t.Errorf("ready to advance, but the map never names the next age %q", st.NextAgeName)
			}
		})
	}
}

// accented writes every vowel of a text with an accent: the same letters, in
// as many cells, in more bytes (as a script in another alphabet is).
var accented = strings.NewReplacer("a", "á", "e", "é", "i", "í", "o", "ó", "u", "ú", "A", "Á", "E", "É", "I", "Í", "O", "Ó", "U", "Ú")

// unaccented undoes accented on a picture.
var unaccented = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U")

// TestVibeTextNeverMovesWhatIsAroundIt: a text of the map is drawn by cells,
// never by bytes. With every text the code writes itself read with accents
// (the same letters in the same number of cells, but more bytes each), the map
// draws the same picture as with VibeMode off, once the accents are taken off
// again: the cursor walks the same way, the inspector says the same, the tags
// land in the same cells. A text indexed by bytes or measured with len shows
// here, in a place no one reads the English of.
func TestVibeTextNeverMovesWhatIsAroundIt(t *testing.T) {
	was := vibe.On()
	t.Cleanup(func() { vibe.Set(was) })
	vibe.Set(false)
	fixture.WithTheme(t, "forge")
	keys := []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyRight, 0, 0), tcell.NewEventKey(tcell.KeyDown, 0, 0),
		tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModShift), tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModShift),
		tcell.NewEventKey(tcell.KeyTab, 0, 0), tcell.NewEventKey(tcell.KeyPgDn, 0, 0), tcell.NewEventKey(tcell.KeyTab, 0, 0),
		tcell.NewEventKey(tcell.KeyPgUp, 0, 0), tcell.NewEventKey(tcell.KeyLeft, 0, 0),
	}
	ages := []string{"stone_age", "medieval_age", "modern_age", "cyberpunk_age", "fusion_age", "space_age", "galactic_age", "quantum_age"}

	// pictures walks the map over one model and returns what it drew and said.
	pictures := func(m *mapmodel.Model) string {
		var out strings.Builder
		for _, size := range [][2]int{{160, 48}, {80, 24}} {
			v := New()
			for _, o := range []Option{OptFlows, OptInspect, OptLegend, OptChanges} {
				v.SetOption(o, true)
			}
			scr := capture.NewScreen(size[0], size[1])
			r := Rect{W: size[0], H: size[1]}
			for step := 0; step < 2*len(keys); step++ {
				f := Frame{Model: m, Anim: step, Clock: step * 3, Tier: mapmodel.TierUnicode}
				v.HandleKey(keys[step%len(keys)], f)
				in, ok := v.Inspect(f)
				fmt.Fprintf(&out, "inspect %v %q %q %q %q\n", ok, in.Title, in.Lines, in.Command, in.Kind)
				if step%2 == 0 {
					v.Draw(scr, r, f)
					out.WriteString(capture.Text(scr))
				}
			}
			scr.SetSize(40, 15)
			New().DrawCompact(scr, Rect{W: 40, H: 15}, Frame{Model: m, Tier: mapmodel.TierUnicode})
			out.WriteString(capture.Text(scr))
			scr.Fini()
		}
		return out.String()
	}

	for ai, age := range ages {
		st := fixture.State(fixture.Options{Age: age, Seed: int64(7 + ai%3), Scale: 1.5, Routes: 2, Wars: ai % 3,
			Harbinger: ai%2 == 0, Catastrophe: ai == 3, Ruins: ai%4 == 0})
		prev := fixture.State(fixture.Options{Age: age, Seed: int64(7 + ai%3), Scale: 1, Routes: 1})
		since := mapmodel.VisitOf(mapmodel.NewBuilder(nil).Build(&prev, nil))

		vibe.Set(false)
		plain := pictures(mapmodel.NewBuilder(nil).Build(&st, since))

		// The script is empty, so the only words that change are the code's own.
		restoreScript := vibe.UseScriptForTest(map[string]string{})
		restoreRewriter := vibe.UseRewriterForTest(func(id, plain string) string {
			var sb strings.Builder
			after := false // the letter after a % is the verb of a format, not a word
			for _, r := range plain {
				switch {
				case after:
					sb.WriteRune(r)
					after = false
				case r == '%':
					sb.WriteRune(r)
					after = true
				default:
					sb.WriteString(accented.Replace(string(r)))
				}
			}
			return sb.String()
		})
		vibe.Set(true)
		on := pictures(mapmodel.NewBuilder(nil).Build(&st, since))
		vibe.Set(false)
		restoreRewriter()
		restoreScript()

		if !strings.ContainsAny(on, "áéíóú") {
			t.Errorf("%s: with VibeMode on nothing of the picture was read with accents", age)
		}
		if got := unaccented.Replace(on); got != plain {
			t.Errorf("%s: with VibeMode on (accents taken off again) the picture is not the plain one: %s", age, firstDifferingLine(plain, got))
		}
	}
}

// firstDifferingLine names the first line two pictures differ in.
func firstDifferingLine(was, now string) string {
	a, b := strings.Split(was, "\n"), strings.Split(now, "\n")
	for i := 0; i < len(a) || i < len(b); i++ {
		var x, y string
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		if x != y {
			return fmt.Sprintf("line %d was %q, now %q", i+1, x, y)
		}
	}
	return "no difference"
}
