//go:build mapcapture

package all

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// TestWriteMandalaCaptures draws the Transcendent Age's mandala in both
// styles from the smoke bot's transcendent_age state (run
// TestGenerateStates in ui/mapstyle/capture first): a run through all
// seven eras, the same run as if it had passed through only the first
// four (its record of ages reached cut down; the map shows a ring only for
// an era reached), and the same run with buildings of the Transcendent Age
// itself (the bot's state has none: the crown is still to be raised). Each goes full size and as the mini map, as an animated
// HTML file, its first frame as text and as cells (decimal rune and hex
// colours per cell, for an outside renderer), with four frames of the wave
// of light and the full view again with the cursor out and with the legend
// on, into MANDALA_CAPTURE_DIR
// (default map_captures/mandala/after at the repo root, which git
// ignores). It uses only the public styles, so the same file run on an
// older checkout writes the "before" set.
//
//	go test -tags mapcapture -run TestWriteMandalaCaptures ./ui/mapstyle/all
func TestWriteMandalaCaptures(t *testing.T) {
	root := filepath.Join("..", "..", "..", "map_captures")
	states := envOr("MAP_STATES_DIR", filepath.Join(root, "states"))
	out := envOr("MANDALA_CAPTURE_DIR", filepath.Join(root, "mandala", "after"))
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	th := envOr("SKY_THEME", "forge")
	if err := theme.SetActive(th); err != nil {
		t.Fatal(err)
	}
	st, err := capture.LoadState(filepath.Join(states, "transcendent_age.json.gz"))
	if err != nil {
		t.Skipf("no state: %v", err)
	}
	four := st
	four.Stats.AgesReached = nil
	for _, e := range config.Epochs()[:4] {
		four.Stats.AgesReached = append(four.Stats.AgesReached, e.Ages...)
	}
	four.Stats.AgesReached = append(four.Stats.AgesReached, st.Age)
	crown := st // the bot stops the moment it arrives: give the age some buildings of its own
	for i := 1; i <= 3; i++ {
		crown = fixture.Grow(crown, i)
	}
	crown.Tick = st.Tick
	for _, run := range []struct {
		name string
		m    *mapmodel.Model
	}{{"7eras", mapmodel.NewBuilder(nil).Build(&st, nil)}, {"4eras", mapmodel.NewBuilder(nil).Build(&four, nil)},
		{"crown", mapmodel.NewBuilder(nil).Build(&crown, nil)}} {
		for _, style := range Registry().Names() {
			for _, mini := range []bool{false, true} {
				w, h := 160, 48
				name := style + "_" + run.name
				if mini {
					w, h, name = 40, 15, name+"_mini"
				}
				v, _ := Registry().New(style)
				scr := capture.NewScreen(w, h)
				var frames []string
				for i := 0; i < 32; i++ {
					f := mapstyle.Frame{Model: run.m, Anim: 2000 + i*6, Tier: mapmodel.TierUnicode}
					if mini {
						v.DrawCompact(scr, mapstyle.Rect{W: w, H: h}, f)
					} else {
						v.Draw(scr, mapstyle.Rect{W: w, H: h}, f)
					}
					frames = append(frames, capture.Frame(scr))
					if i == 0 {
						_ = os.WriteFile(filepath.Join(out, name+".txt"), []byte(capture.Text(scr)), 0o644)
						_ = os.WriteFile(filepath.Join(out, name+".cells"), []byte(mandalaCells(scr)), 0o644)
					}
					if i%2 == 0 && i < 8 && !mini && run.name == "7eras" { // the wave of light, a quarter of a breath apart
						_ = os.WriteFile(filepath.Join(out, name+"_pulse"+strconv.Itoa(i/2)+".cells"), []byte(mandalaCells(scr)), 0o644)
					}
				}
				if !mini && run.name == "7eras" {
					// the cursor on a building of a ring, and the legend line
					for i := 0; i < 60; i++ {
						v.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), mapstyle.Frame{Model: run.m, Anim: 2000})
					}
					v.Draw(scr, mapstyle.Rect{W: w, H: h}, mapstyle.Frame{Model: run.m, Anim: 2000, Tier: mapmodel.TierUnicode})
					_ = os.WriteFile(filepath.Join(out, name+"_inspect.cells"), []byte(mandalaCells(scr)), 0o644)
					v.HandleKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), mapstyle.Frame{Model: run.m, Anim: 2000})
					v.SetOption(mapstyle.OptInspect, false)
					v.SetOption(mapstyle.OptLegend, true)
					v.Draw(scr, mapstyle.Rect{W: w, H: h}, mapstyle.Frame{Model: run.m, Anim: 2000, Tier: mapmodel.TierUnicode})
					_ = os.WriteFile(filepath.Join(out, name+"_legend.cells"), []byte(mandalaCells(scr)), 0o644)
				}
				bg := capture.Hex(theme.Color(theme.RoleBackground))
				fg := capture.Hex(theme.Color(theme.RoleText))
				title := style + " · transcendent_age · " + run.name + " · " + strconv.Itoa(w) + "x" + strconv.Itoa(h)
				_ = os.WriteFile(filepath.Join(out, name+".html"), []byte(capture.HTML(title, bg, fg, capture.Fonts, frames, 6)), 0o644)
				scr.Fini()
			}
		}
	}
}

// mandalaCells writes a screen's cells for an outside renderer: "W H",
// then a line per row of tab-separated "rune fg bg" (decimal rune, hex
// colours).
func mandalaCells(scr tcell.SimulationScreen) string {
	scr.Show()
	cells, w, h := scr.GetContents()
	var b strings.Builder
	b.WriteString(strconv.Itoa(w) + " " + strconv.Itoa(h) + "\n")
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			r := ' '
			if len(c.Runes) > 0 {
				r = c.Runes[0]
			}
			fg, bg, _ := c.Style.Decompose()
			if x > 0 {
				b.WriteByte('\t')
			}
			b.WriteString(strconv.Itoa(int(r)) + " " + capture.Hex(fg) + " " + capture.Hex(bg))
		}
		b.WriteByte('\n')
	}
	return b.String()
}
