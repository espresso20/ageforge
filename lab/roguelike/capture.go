package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// frame is one captured screen: the real tcell cells, not an image.
type frame struct {
	W, H  int
	Cells []tcell.SimCell
}

// snap renders a view into a SimulationScreen of w×h and returns its cells.
func snap(v *View, w, h int, compact bool) frame {
	scr := tcell.NewSimulationScreen("UTF-8")
	if err := scr.Init(); err != nil {
		panic(err)
	}
	scr.SetSize(w, h)
	if compact {
		v.DrawCompact(scr, w, h)
	} else {
		v.Draw(scr, w, h)
	}
	scr.Show()
	cells, cw, ch := scr.GetContents()
	out := make([]tcell.SimCell, len(cells))
	copy(out, cells)
	scr.Fini()
	return frame{cw, ch, out}
}

func (f frame) text() string {
	var b strings.Builder
	for y := 0; y < f.H; y++ {
		var line strings.Builder
		for x := 0; x < f.W; x++ {
			c := f.Cells[y*f.W+x]
			if len(c.Runes) == 0 {
				line.WriteRune(' ')
				continue
			}
			line.WriteString(string(c.Runes))
		}
		b.WriteString(strings.TrimRight(line.String(), " "))
		b.WriteByte('\n')
	}
	return b.String()
}

func hexOf(c tcell.Color, fallback string) string {
	if c == tcell.ColorDefault || !c.Valid() {
		return fallback
	}
	return fmt.Sprintf("#%06x", c.Hex())
}

// htmlBody renders the cells as runs of inline-styled spans.
func (f frame) htmlBody() string {
	bg := hexOf(theme.Color(theme.RoleBackground), "#000000")
	fgDef := hexOf(theme.Color(theme.RoleText), "#cccccc")
	var b strings.Builder
	for y := 0; y < f.H; y++ {
		curStyle := ""
		open := false
		for x := 0; x < f.W; x++ {
			c := f.Cells[y*f.W+x]
			fg, cbg, attr := c.Style.Decompose()
			sty := "color:" + hexOf(fg, fgDef)
			if hb := hexOf(cbg, bg); hb != bg {
				sty += ";background:" + hb
			}
			if attr&tcell.AttrBold != 0 {
				sty += ";font-weight:bold"
			}
			if attr&tcell.AttrReverse != 0 {
				sty = "color:" + hexOf(cbg, bg) + ";background:" + hexOf(fg, fgDef) + ";font-weight:bold"
			}
			if sty != curStyle {
				if open {
					b.WriteString("</span>")
				}
				b.WriteString(`<span style="` + sty + `">`)
				curStyle, open = sty, true
			}
			r := " "
			if len(c.Runes) > 0 {
				r = string(c.Runes)
			}
			b.WriteString(html.EscapeString(r))
		}
		if open {
			b.WriteString("</span>")
		}
		b.WriteByte('\n')
	}
	return b.String()
}

const pageHead = `<!doctype html><html><head><meta charset="utf-8"><title>%s</title>
<style>
body{margin:0;padding:16px;background:%s;color:%s;font-family:system-ui,sans-serif}
pre{font-family:"JetBrains Mono","DejaVu Sans Mono",Menlo,Consolas,monospace;font-size:13px;line-height:1.18;margin:0;display:inline-block;padding:6px;border:1px solid %s}
h1{font-size:13px;font-weight:600;margin:0 0 8px;opacity:.8}
</style></head><body><h1>%s</h1>
`

func writeCapture(dir, name, title string, frames []frame) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, name+".txt"), []byte(frames[0].text()), 0o644); err != nil {
		return err
	}
	bg := hexOf(theme.Color(theme.RoleBackground), "#000")
	fg := hexOf(theme.Color(theme.RoleText), "#ccc")
	bd := hexOf(theme.Color(theme.RoleBorder), "#444")
	var b strings.Builder
	fmt.Fprintf(&b, pageHead, html.EscapeString(title), bg, fg, bd, html.EscapeString(title))
	if len(frames) == 1 {
		b.WriteString("<pre>" + frames[0].htmlBody() + "</pre>")
	} else {
		b.WriteString(`<pre id="f"></pre><script>const F=[`)
		for i, f := range frames {
			if i > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(&b, "%q", f.htmlBody())
		}
		b.WriteString(`];let i=0;const e=document.getElementById('f');function t(){e.innerHTML=F[i];i=(i+1)%F.length}t();setInterval(t,280);</script>`)
	}
	b.WriteString("</body></html>\n")
	return os.WriteFile(filepath.Join(dir, name+".html"), []byte(b.String()), 0o644)
}
