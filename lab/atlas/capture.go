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

// Captures go through a real tcell SimulationScreen: the canvas is blitted
// exactly as the interactive view blits it, then read back cell by cell.

func blit(scr tcell.Screen, cv *Canvas) {
	for y := 0; y < cv.H; y++ {
		for x := 0; x < cv.W; x++ {
			c := cv.C[y*cv.W+x]
			st := tcell.StyleDefault.Foreground(c.Fg).Background(c.Bg).Attributes(c.Attr)
			scr.SetContent(x, y, c.R, nil, st)
		}
	}
}

type simCell struct {
	r      rune
	fg, bg tcell.Color
	attr   tcell.AttrMask
}

func simulate(cv *Canvas) ([]simCell, error) {
	scr := tcell.NewSimulationScreen("UTF-8")
	if err := scr.Init(); err != nil {
		return nil, err
	}
	defer scr.Fini()
	scr.SetSize(cv.W, cv.H)
	blit(scr, cv)
	scr.Show()
	cells, w, h := scr.(tcell.SimulationScreen).GetContents()
	if w != cv.W || h != cv.H {
		return nil, fmt.Errorf("simulation screen is %dx%d, want %dx%d", w, h, cv.W, cv.H)
	}
	out := make([]simCell, len(cells))
	for i, c := range cells {
		fg, bg, at := c.Style.Decompose()
		r := ' '
		if len(c.Runes) > 0 {
			r = c.Runes[0]
		}
		out[i] = simCell{r, fg, bg, at}
	}
	return out, nil
}

func toText(cells []simCell, w, h int) string {
	var b strings.Builder
	for y := 0; y < h; y++ {
		var l strings.Builder
		for x := 0; x < w; x++ {
			l.WriteRune(cells[y*w+x].r)
		}
		b.WriteString(strings.TrimRight(l.String(), " "))
		b.WriteByte('\n')
	}
	return b.String()
}

func cssColor(c tcell.Color, pal16 bool) string {
	if pal16 {
		c = tcell.FindColor(c, ansi16)
	}
	r, g, b := c.RGB()
	if r < 0 {
		return "inherit"
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// ansi16 is the xterm 16-colour palette, for the degraded capture.
var ansi16 = []tcell.Color{
	tcell.NewRGBColor(0, 0, 0), tcell.NewRGBColor(205, 0, 0), tcell.NewRGBColor(0, 205, 0), tcell.NewRGBColor(205, 205, 0),
	tcell.NewRGBColor(0, 0, 238), tcell.NewRGBColor(205, 0, 205), tcell.NewRGBColor(0, 205, 205), tcell.NewRGBColor(229, 229, 229),
	tcell.NewRGBColor(127, 127, 127), tcell.NewRGBColor(255, 0, 0), tcell.NewRGBColor(0, 255, 0), tcell.NewRGBColor(255, 255, 0),
	tcell.NewRGBColor(92, 92, 255), tcell.NewRGBColor(255, 0, 255), tcell.NewRGBColor(0, 255, 255), tcell.NewRGBColor(255, 255, 255),
}

// htmlPre renders cells as a <pre> of styled spans, runs of equal style
// merged.
func htmlPre(cells []simCell, w, h int, pal16 bool) string {
	var b strings.Builder
	b.WriteString("<pre class=\"term\">")
	for y := 0; y < h; y++ {
		b.WriteString("<div>")
		var cur string
		var run strings.Builder
		flush := func() {
			if run.Len() > 0 {
				fmt.Fprintf(&b, "<span style=\"%s\">%s</span>", cur, cellText(run.String()))
				run.Reset()
			}
		}
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			st := fmt.Sprintf("color:%s;background:%s", cssColor(c.fg, pal16), cssColor(c.bg, pal16))
			if c.attr&tcell.AttrBold != 0 {
				st += ";font-weight:bold"
			}
			if c.attr&tcell.AttrItalic != 0 {
				st += ";font-style:italic"
			}
			if c.attr&tcell.AttrReverse != 0 {
				st = fmt.Sprintf("color:%s;background:%s", cssColor(c.bg, pal16), cssColor(c.fg, pal16))
			}
			if st != cur {
				flush()
				cur = st
			}
			run.WriteRune(c.r)
		}
		flush()
		b.WriteString("</div>")
	}
	b.WriteString("</pre>")
	return b.String()
}

const htmlHead = `<!doctype html><html><head><meta charset="utf-8"><title>%s</title>
<style>
body{margin:0;padding:16px;background:%s;color:%s;font-family:ui-monospace,"JetBrains Mono","SF Mono",Menlo,Consolas,monospace}
h1{font-size:13px;font-weight:600;margin:0 0 8px;opacity:.75}
h2{font-size:13px;font-weight:600;margin:18px 0 4px}
a{color:inherit}
pre.term{font-family:inherit;font-size:13px;margin:0;display:inline-block;padding:0;letter-spacing:0}
pre.term div{height:16px;line-height:16px;white-space:pre}
pre.term span{white-space:pre;display:inline-block;height:16px;line-height:16px;vertical-align:top}
pre.term i{font-style:inherit;display:inline-block;width:1ch;overflow:visible;text-align:center}
.frames pre{display:none}.frames pre.on{display:inline-block}
</style></head><body><h1>%s</h1>
`

func writeCapture(dir, name, title string, frames []*Canvas, pal16 bool) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	bg := cssColor(theme.Color(theme.RoleBackground), false)
	fg := cssColor(theme.Color(theme.RoleText), false)
	var body strings.Builder
	fmt.Fprintf(&body, htmlHead, html.EscapeString(title), bg, fg, html.EscapeString(title))
	var txt string
	if len(frames) > 1 {
		body.WriteString(`<div class="frames">`)
	}
	for i, cv := range frames {
		cells, err := simulate(cv)
		if err != nil {
			return err
		}
		if i == 0 {
			txt = toText(cells, cv.W, cv.H)
		}
		pre := htmlPre(cells, cv.W, cv.H, pal16)
		if len(frames) > 1 {
			on := ""
			if i == 0 {
				on = " on"
			}
			pre = strings.Replace(pre, `<pre class="term">`, `<pre class="term`+on+`">`, 1)
		}
		body.WriteString(pre)
	}
	if len(frames) > 1 {
		body.WriteString(`</div><script>
(function(){var f=document.querySelectorAll('.frames pre'),i=0;setInterval(function(){f[i].classList.remove('on');i=(i+1)%f.length;f[i].classList.add('on');},420);})();
</script>`)
	}
	body.WriteString("</body></html>\n")
	if err := os.WriteFile(filepath.Join(dir, name+".html"), []byte(body.String()), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".txt"), []byte(txt), 0o644)
}

// cellText escapes a run, pinning every symbol outside ASCII and the box and
// block ranges to one character cell, so a browser font with wide symbols
// cannot break the grid the terminal keeps.
func cellText(s string) string {
	var b strings.Builder
	for _, r := range s {
		e := html.EscapeString(string(r))
		if r > 0x7f {
			b.WriteString("<i>" + e + "</i>")
		} else {
			b.WriteString(e)
		}
	}
	return b.String()
}
