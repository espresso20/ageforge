package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// loadState loads a save written by generate through the engine's own
// LoadGame, so what the organism draws is what a player's save holds.
func loadState(dir, name string) (game.GameState, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return game.GameState{}, err
	}
	restore := game.SetDataDirForTest(abs)
	defer restore()
	ge := game.NewGameEngine()
	if err := ge.LoadGame(name); err != nil {
		return game.GameState{}, err
	}
	return ge.GetState(), nil
}

// loadPair loads <name> and, when present, <base>_prev for "since your
// last visit".
func loadPair(dir, name string) (*Organism, *Organism, error) {
	st, err := loadState(dir, name)
	if err != nil {
		return nil, nil, err
	}
	o := Build(st)
	base := strings.TrimSuffix(strings.TrimSuffix(name, "_catastrophe"), "_seedling")
	var prev *Organism
	if pst, err := loadState(dir, base+"_prev"); err == nil && name == base {
		prev = Build(pst)
	}
	return o, prev, nil
}

// frame renders one frame on a tcell SimulationScreen: the same cells a
// terminal would get.
func frame(o, prev *Organism, v View, w, h int) tcell.SimulationScreen {
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		panic(err)
	}
	s.SetSize(w, h)
	Render(o, prev, v, w, h).Blit(s, 0, 0)
	s.Show()
	return s
}

func screenText(s tcell.SimulationScreen) string {
	cells, w, h := s.GetContents()
	var b strings.Builder
	for y := 0; y < h; y++ {
		var line strings.Builder
		for x := 0; x < w; x++ {
			rs := cells[y*w+x].Runes
			if len(rs) == 0 {
				line.WriteByte(' ')
			} else {
				line.WriteRune(rs[0])
			}
		}
		b.WriteString(strings.TrimRight(line.String(), " "))
		b.WriteByte('\n')
	}
	return b.String()
}

func hexOf(c tcell.Color, fallback tcell.Color) string {
	if c == tcell.ColorDefault || c.Hex() < 0 {
		c = fallback
	}
	return fmt.Sprintf("#%06x", c.Hex())
}

// quantize16 maps a colour to the nearest of the 16 ANSI colours (xterm
// values), to prove the picture degrades.
func quantize16(c tcell.Color) tcell.Color {
	ansi := []int32{0x000000, 0x800000, 0x008000, 0x808000, 0x000080, 0x800080, 0x008080, 0xc0c0c0,
		0x808080, 0xff0000, 0x00ff00, 0xffff00, 0x0000ff, 0xff00ff, 0x00ffff, 0xffffff}
	r, g, b := c.RGB()
	best, bd := ansi[0], int64(1<<62)
	for _, a := range ansi {
		ar, ag, ab := int64(a>>16&0xff), int64(a>>8&0xff), int64(a&0xff)
		d := (ar-int64(r))*(ar-int64(r)) + (ag-int64(g))*(ag-int64(g)) + (ab-int64(b))*(ab-int64(b))
		if d < bd {
			best, bd = a, d
		}
	}
	return tcell.NewHexColor(best)
}

// screenHTML renders the cells as <pre> rows with inline colours.
func screenHTML(s tcell.SimulationScreen, q16 bool) []string {
	cells, w, h := s.GetContents()
	bg := theme.Color(theme.RoleBackground)
	fgDef := theme.Color(theme.RoleText)
	rows := make([]string, h)
	for y := 0; y < h; y++ {
		var b strings.Builder
		cur := ""
		open := false
		for x := 0; x < w; x++ {
			ce := cells[y*w+x]
			fg, bgc, attr := ce.Style.Decompose()
			if q16 {
				if fg != tcell.ColorDefault {
					fg = quantize16(fg)
				}
				if bgc != tcell.ColorDefault {
					bgc = quantize16(bgc)
				}
			}
			st := fmt.Sprintf("color:%s;background:%s", hexOf(fg, fgDef), hexOf(bgc, bg))
			if attr&tcell.AttrBold != 0 {
				st += ";font-weight:bold"
			}
			if attr&tcell.AttrItalic != 0 {
				st += ";font-style:italic"
			}
			if st != cur {
				if open {
					b.WriteString("</span>")
				}
				b.WriteString(`<span style="` + st + `">`)
				cur, open = st, true
			}
			r := ' '
			if len(ce.Runes) > 0 {
				r = ce.Runes[0]
			}
			b.WriteString(html.EscapeString(string(r)))
		}
		if open {
			b.WriteString("</span>")
		}
		rows[y] = b.String()
	}
	return rows
}

const pageHead = `<!doctype html><html><head><meta charset="utf-8"><title>%s</title>
<style>
body{margin:0;padding:24px;background:%s;color:%s;font-family:ui-monospace,"SF Mono",Menlo,"DejaVu Sans Mono",monospace}
h1{font:600 14px/1.4 ui-monospace,Menlo,monospace;margin:0 0 12px;opacity:.75}
pre{margin:0;font-size:13px;line-height:1.18;letter-spacing:0}
pre span{display:inline}
</style></head><body><h1>%s</h1>
`

func writeHTML(path, title string, frames [][]string, fps float64) error {
	bg := hexOf(theme.Color(theme.RoleBackground), tcell.ColorBlack)
	fg := hexOf(theme.Color(theme.RoleText), tcell.ColorWhite)
	var b strings.Builder
	fmt.Fprintf(&b, pageHead, html.EscapeString(title), bg, fg, html.EscapeString(title))
	for i, f := range frames {
		disp := ""
		if i > 0 {
			disp = ` style="display:none"`
		}
		fmt.Fprintf(&b, "<pre class=\"frame\"%s>%s</pre>\n", disp, strings.Join(f, "\n"))
	}
	if len(frames) > 1 {
		fmt.Fprintf(&b, `<script>
const fr=[...document.querySelectorAll('.frame')];let i=0;
setInterval(()=>{fr[i].style.display='none';i=(i+1)%%fr.length;fr[i].style.display='block'},%d);
</script>
`, int(1000/fps))
	}
	b.WriteString("</body></html>\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// captureSpec is one entry on the comparison page.
type captureSpec struct {
	State  string
	Out    string
	Theme  string
	W, H   int
	Frames int
	Sel    string // lineage key to inspect, "" for none
	Q16    bool
}

func runCapture(states, outDir string, cs captureSpec) error {
	if err := theme.SetActive(cs.Theme); err != nil {
		return err
	}
	o, prev, err := loadPair(states, cs.State)
	if err != nil {
		return fmt.Errorf("%s: %w", cs.State, err)
	}
	v := View{Sel: -1, Since: prev != nil}
	if cs.Sel != "" {
		for i, l := range o.Limbs {
			if l.Spec.Key == cs.Sel {
				v.Sel = i
			}
		}
	}
	nf := max(1, cs.Frames)
	var frames [][]string
	var txt string
	for f := 0; f < nf; f++ {
		fv := v
		if nf > 1 {
			fv.T = 1 + float64(f)*0.25
		}
		s := frame(o, prev, fv, cs.W, cs.H)
		if f == 0 {
			txt = screenText(s)
		}
		frames = append(frames, screenHTML(s, cs.Q16))
		s.Fini()
	}
	if err := os.WriteFile(filepath.Join(outDir, cs.Out+".txt"), []byte(txt), 0o644); err != nil {
		return err
	}
	title := fmt.Sprintf("organism · %s · %s · %dx%d", o.AgeName, cs.Theme, cs.W, cs.H)
	if nf > 1 {
		title += fmt.Sprintf(" · %d frames", nf)
	}
	return writeHTML(filepath.Join(outDir, cs.Out+".html"), title, frames, 4)
}
