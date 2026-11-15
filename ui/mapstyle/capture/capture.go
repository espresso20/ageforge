// Package capture turns map frames drawn on a Screen into
// review files: plain text (the monochrome check) and HTML built cell by
// cell from the screen (a real terminal render, not an image), animated when
// there are several frames.
package capture

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// A Cell is one square of a Screen: what was drawn there, as it was drawn.
type Cell struct {
	Runes []rune
	Style kit.Style
}

// A Screen is a grid of cells for a map to be drawn on and read back from
// (it is a mapstyle.Cells). Nothing is resolved: a cell holds the style it
// was given, theme roles and all.
type Screen struct {
	w, h  int
	cells []Cell
}

// NewScreen returns a blank screen of w x h.
func NewScreen(w, h int) *Screen {
	s := &Screen{}
	s.SetSize(w, h)
	return s
}

// SetSize makes the screen w x h and blanks it.
func (s *Screen) SetSize(w, h int) {
	s.w, s.h = w, h
	s.cells = make([]Cell, w*h)
	s.Clear()
}

func (s *Screen) Size() (w, h int) { return s.w, s.h }

// Clear blanks every cell.
func (s *Screen) Clear() {
	for i := range s.cells {
		s.cells[i] = Cell{Runes: []rune{' '}}
	}
}

func (s *Screen) SetContent(x, y int, r rune, _ []rune, st kit.Style) {
	if x >= 0 && y >= 0 && x < s.w && y < s.h {
		s.cells[y*s.w+x] = Cell{Runes: []rune{r}, Style: st}
	}
}

func (s *Screen) GetContent(x, y int) (rune, []rune, kit.Style, int) {
	if x < 0 || y < 0 || x >= s.w || y >= s.h {
		return ' ', nil, kit.Style{}, 1
	}
	c := s.cells[y*s.w+x]
	return c.Runes[0], nil, c.Style, 1
}

// GetContents is every cell, row by row, and the screen's size.
func (s *Screen) GetContents() ([]Cell, int, int) { return s.cells, s.w, s.h }

// Show and Fini are here for the code that was written for a terminal
// screen: there is nothing to flush and nothing to close.
func (s *Screen) Show() {}
func (s *Screen) Fini() {}

// Text shows the screen and reads it back as plain text, trailing spaces
// trimmed.
func Text(s *Screen) string {
	s.Show()
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

// Frame renders a screen's cells as HTML spans with inline colours.
func Frame(s *Screen) string {
	s.Show()
	cells, w, h := s.GetContents()
	var b strings.Builder
	for y := 0; y < h; y++ {
		var run strings.Builder
		curFg, curBg := theme.ColorDefault, theme.ColorDefault
		bold := false
		flush := func() {
			if run.Len() == 0 {
				return
			}
			fr, fg, fb := curFg.RGB()
			br, bg, bb := curBg.RGB()
			wt := ""
			if bold {
				wt = ";font-weight:bold"
			}
			fmt.Fprintf(&b, `<span style="color:#%02x%02x%02x;background:#%02x%02x%02x%s">%s</span>`,
				fr&0xff, fg&0xff, fb&0xff, br&0xff, bg&0xff, bb&0xff, wt, html.EscapeString(run.String()))
			run.Reset()
		}
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			fg, bg, attr := c.Style.Decompose()
			bd := attr&kit.Bold != 0
			if fg != curFg || bg != curBg || bd != bold {
				flush()
				curFg, curBg, bold = fg, bg, bd
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
body{margin:0;background:%s;color:#ccc;font-family:%s}
h1{font-size:13px;font-weight:normal;margin:10px 14px;color:%s;opacity:.8}
pre{margin:0 14px 14px;font-size:13px;line-height:1.0;letter-spacing:0;display:inline-block;font-family:inherit}
.f{display:none}.f.on{display:block}
</style></head><body><h1>%s</h1>
`

// Fonts is the CSS font stack for captures.
const Fonts = `ui-monospace,Menlo,Consolas,"DejaVu Sans Mono",monospace`

var styleRe = regexp.MustCompile(`<span style="([^"]+)">`)

// classify swaps inline styles for short CSS classes: the same cells, a
// third of the bytes.
func classify(frames []string) ([]string, string) {
	names := map[string]string{}
	var css strings.Builder
	out := make([]string, len(frames))
	for i, f := range frames {
		out[i] = styleRe.ReplaceAllStringFunc(f, func(m string) string {
			k := styleRe.FindStringSubmatch(m)[1]
			n, ok := names[k]
			if !ok {
				n = fmt.Sprintf("c%x", len(names))
				names[k] = n
				fmt.Fprintf(&css, ".%s{%s}", n, k)
			}
			return `<span class="` + n + `">`
		})
	}
	return out, css.String()
}

// HTML wraps frames into a page; several frames play as an animation at
// fps. bg and fg are the page colours, fonts the CSS font stack.
func HTML(title, bg, fg, fonts string, frames []string, fps int) string {
	var b strings.Builder
	frames, css := classify(frames)
	fmt.Fprintf(&b, htmlHead, html.EscapeString(title), bg, fonts, fg, html.EscapeString(title))
	fmt.Fprintf(&b, "<style>%s</style>\n", css)
	if len(frames) == 1 {
		b.WriteString("<pre>" + frames[0] + "</pre></body></html>\n")
		return b.String()
	}
	for i, f := range frames {
		cls := "f"
		if i == 0 {
			cls = "f on"
		}
		fmt.Fprintf(&b, "<pre class=\"%s\">%s</pre>\n", cls, f)
	}
	if fps <= 0 {
		fps = 6
	}
	fmt.Fprintf(&b, `<script>const f=document.querySelectorAll('.f');let i=0;setInterval(()=>{f[i].classList.remove('on');i=(i+1)%%f.length;f[i].classList.add('on')},%d)</script>`, 1000/fps)
	b.WriteString("</body></html>\n")
	return b.String()
}

// Hex renders a colour as #rrggbb for page chrome.
func Hex(c theme.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r&0xff, g&0xff, b&0xff)
}
