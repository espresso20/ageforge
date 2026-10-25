package all

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
	"github.com/espresso20/ageforge/vibe"
)

// accented writes every vowel of a text with an accent: the same letters, in
// as many cells, in more bytes (as a script in another alphabet is).
var accented = strings.NewReplacer("a", "á", "e", "é", "i", "í", "o", "ó", "u", "ú", "A", "Á", "E", "É", "I", "Í", "O", "Ó", "U", "Ú")

// unaccented undoes accented on a picture.
var unaccented = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U")

// TestVibeTextNeverMovesWhatIsAroundIt: a text of the map is drawn by cells, never
// by bytes. With every text the code writes itself read with accents (the
// same letters in the same number of cells, but more bytes each), both styles
// draw the same picture as with VibeMode off, once the accents are taken off
// again: the cursor walks the same way, the inspector says the same, the
// tags and the signs land in the same cells. A text indexed by bytes or
// measured with len shows here, in a place no one reads the English of.
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
	// A ground age of each look, the cyberpunk megacity with its signs and tags,
	// and the sky scenes.
	ages := []string{"stone_age", "medieval_age", "modern_age", "cyberpunk_age", "fusion_age", "space_age", "galactic_age", "quantum_age"}

	// pictures walks one style over one model and returns what it drew and said.
	pictures := func(style string, m *mapmodel.Model) string {
		var b strings.Builder
		for _, size := range [][2]int{{160, 48}, {80, 24}} {
			v, _ := Registry().New(style)
			for _, o := range []mapstyle.Option{mapstyle.OptFlows, mapstyle.OptInspect, mapstyle.OptLegend, mapstyle.OptChanges} {
				v.SetOption(o, true)
			}
			scr := capture.NewScreen(size[0], size[1])
			r := mapstyle.Rect{W: size[0], H: size[1]}
			for step := 0; step < 2*len(keys); step++ {
				f := mapstyle.Frame{Model: m, Anim: step, Clock: step * 3, Tier: mapmodel.TierUnicode}
				v.HandleKey(keys[step%len(keys)], f)
				in, ok := v.Inspect(f)
				fmt.Fprintf(&b, "inspect %v %q %q %q %q\n", ok, in.Title, in.Lines, in.Command, in.Kind)
				if step%2 == 0 {
					v.Draw(scr, r, f)
					b.WriteString(capture.Text(scr))
				}
			}
			v2, _ := Registry().New(style)
			scr.SetSize(40, 15)
			v2.DrawCompact(scr, mapstyle.Rect{W: 40, H: 15}, mapstyle.Frame{Model: m, Tier: mapmodel.TierUnicode})
			b.WriteString(capture.Text(scr))
			scr.Fini()
		}
		return b.String()
	}

	for _, style := range Registry().Names() {
		for ai, age := range ages {
			st := fixture.State(fixture.Options{Age: age, Seed: int64(7 + ai%3), Scale: 1.5, Routes: 2, Wars: ai % 3,
				Harbinger: ai%2 == 0, Catastrophe: ai == 3, Ruins: ai%4 == 0})
			prev := fixture.State(fixture.Options{Age: age, Seed: int64(7 + ai%3), Scale: 1, Routes: 1})
			since := mapmodel.VisitOf(mapmodel.NewBuilder(nil).Build(&prev, nil))

			vibe.Set(false)
			plain := pictures(style, mapmodel.NewBuilder(nil).Build(&st, since))

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
			on := pictures(style, mapmodel.NewBuilder(nil).Build(&st, since))
			vibe.Set(false)
			restoreRewriter()
			restoreScript()

			if !strings.ContainsAny(on, "áéíóú") {
				t.Errorf("%s %s: with VibeMode on nothing of the picture was read with accents", style, age)
			}
			if got := unaccented.Replace(on); got != plain {
				t.Errorf("%s %s: with VibeMode on (accents taken off again) the picture is not the plain one: %s",
					style, age, firstDifferingLine(plain, got))
			}
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
