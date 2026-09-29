package main

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Px is one terminal cell of the composed frame.
type Px struct {
	Ch     rune
	Fg, Bg RGB
	Bold   bool
}

// FB is a frame buffer the size of the terminal. The scene composes into
// it back to front; it is then written to a tcell screen (or captured).
type FB struct {
	W, H int
	C    []Px
}

func newFB(w, h int) *FB { return &FB{W: w, H: h, C: make([]Px, w*h)} }

func (f *FB) ok(x, y int) bool { return x >= 0 && y >= 0 && x < f.W && y < f.H }

func (f *FB) at(x, y int) *Px {
	if !f.ok(x, y) {
		return nil
	}
	return &f.C[y*f.W+x]
}

func (f *FB) set(x, y int, ch rune, fg, bg RGB) {
	if p := f.at(x, y); p != nil {
		*p = Px{Ch: ch, Fg: fg, Bg: bg}
	}
}

// fg draws a glyph keeping the cell's background (partial glyphs let the
// scene behind show through).
func (f *FB) fg(x, y int, ch rune, fg RGB) {
	if p := f.at(x, y); p != nil {
		if p.Ch == '█' { // a full block's colour is the visible background
			p.Bg = p.Fg
		}
		p.Ch, p.Fg, p.Bold = ch, fg, false
	}
}

func (f *FB) fill(x, y int, c RGB) { f.set(x, y, ' ', c, c) }

func (f *FB) text(x, y int, s string, fg, bg RGB) int {
	for _, r := range s {
		f.set(x, y, r, fg, bg)
		x++
	}
	return x
}

// visible returns the colour a viewer sees in the cell's upper half, used
// for reflections and the compact view.
func (p Px) visible() RGB {
	switch p.Ch {
	case '█', '▀', '▓':
		return p.Fg
	case '▒':
		return p.Fg.Lerp(p.Bg, 0.5)
	}
	return p.Bg
}

// cellOut is what goes to a terminal: a full block becomes a space on its
// colour (no font seams between rows).
func (p Px) cellOut() (rune, RGB, RGB) {
	if p.Ch == '█' {
		return ' ', p.Fg, p.Fg
	}
	if p.Ch == 0 {
		return ' ', p.Fg, p.Bg
	}
	return p.Ch, p.Fg, p.Bg
}

func (f *FB) blitTo(s tcell.Screen) {
	for y := 0; y < f.H; y++ {
		for x := 0; x < f.W; x++ {
			ch, fg, bg := f.C[y*f.W+x].cellOut()
			st := tcell.StyleDefault.Foreground(fg.TC()).Background(bg.TC())
			s.SetContent(x, y, ch, nil, st)
		}
	}
}

// ----------------------------------------------------------- captures

// screenText reads a tcell screen back as plain text.
func screenText(s tcell.SimulationScreen) string {
	cells, w, h := s.GetContents()
	var b strings.Builder
	for y := 0; y < h; y++ {
		var line strings.Builder
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			r := ' '
			if len(c.Runes) > 0 {
				r = c.Runes[0]
			}
			line.WriteRune(r)
		}
		b.WriteString(strings.TrimRight(line.String(), " "))
		b.WriteByte('\n')
	}
	return b.String()
}

// screenHTML renders a tcell SimulationScreen's cells to an HTML <pre>
// with inline colours: a real terminal render, not an image.
func screenHTML(s tcell.SimulationScreen) string {
	cells, w, h := s.GetContents()
	var b strings.Builder
	for y := 0; y < h; y++ {
		var run strings.Builder
		curFg, curBg := tcell.ColorDefault, tcell.ColorDefault
		flush := func() {
			if run.Len() == 0 {
				return
			}
			fr, fgc, fb := curFg.RGB()
			br, bgc, bb := curBg.RGB()
			fmt.Fprintf(&b, `<span style="color:#%02x%02x%02x;background:#%02x%02x%02x">%s</span>`,
				fr, fgc, fb, br, bgc, bb, html.EscapeString(run.String()))
			run.Reset()
		}
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			fg, bg, _ := c.Style.Decompose()
			if fg != curFg || bg != curBg {
				flush()
				curFg, curBg = fg, bg
			}
			r := ' '
			if len(c.Runes) > 0 {
				r = c.Runes[0]
			}
			run.WriteRune(r)
		}
		flush()
		b.WriteByte('\n')
	}
	return b.String()
}

const htmlHead = `<!doctype html><html><head><meta charset="utf-8"><title>%s</title>
<style>
body{margin:0;background:%s;color:#ccc;font-family:ui-monospace,Menlo,Consolas,"DejaVu Sans Mono",monospace}
h1{font-size:13px;font-weight:normal;margin:10px 14px;color:%s;opacity:.8}
pre{margin:0 14px 14px;font-size:13px;line-height:1.0;letter-spacing:0;display:inline-block}
pre span{display:inline}
.f{display:none}.f.on{display:block}
</style></head><body><h1>%s</h1>
`

var styleRe = regexp.MustCompile(`<span style="color:#([0-9a-f]{6});background:#([0-9a-f]{6})">`)

// classify swaps the inline colour pairs for short CSS classes: the same
// cells, a third of the bytes.
func classify(frames []string) ([]string, string) {
	names := map[string]string{}
	var css strings.Builder
	out := make([]string, len(frames))
	for i, f := range frames {
		out[i] = styleRe.ReplaceAllStringFunc(f, func(m string) string {
			sm := styleRe.FindStringSubmatch(m)
			k := sm[1] + sm[2]
			n, ok := names[k]
			if !ok {
				n = fmt.Sprintf("c%x", len(names))
				names[k] = n
				fmt.Fprintf(&css, ".%s{color:#%s;background:#%s}", n, sm[1], sm[2])
			}
			return `<span class="` + n + `">`
		})
	}
	return out, css.String()
}

func wrapHTML(title, bg, fg string, frames []string, fps int) string {
	var b strings.Builder
	frames, css := classify(frames)
	fmt.Fprintf(&b, htmlHead, html.EscapeString(title), bg, fg, html.EscapeString(title))
	fmt.Fprintf(&b, "<style>%s</style>\n", css)
	if len(frames) == 1 {
		b.WriteString("<pre>")
		b.WriteString(frames[0])
		b.WriteString("</pre></body></html>\n")
		return b.String()
	}
	for i, fr := range frames {
		cls := "f"
		if i == 0 {
			cls = "f on"
		}
		fmt.Fprintf(&b, `<pre class="%s">%s</pre>`, cls, fr)
	}
	fmt.Fprintf(&b, `<script>var f=document.querySelectorAll('.f'),i=0;setInterval(function(){f[i].classList.remove('on');i=(i+1)%%f.length;f[i].classList.add('on')},%d);</script></body></html>
`, 1000/max(1, fps))
	return b.String()
}
