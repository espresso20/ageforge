// Package capture turns map frames drawn on a tcell SimulationScreen into
// review files: plain text (the monochrome check) and HTML built cell by
// cell from the screen (a real terminal render, not an image), animated when
// there are several frames. It also loads and saves the game states the
// captures are drawn from.
package capture

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
)

// NewScreen returns an initialised simulation screen of w x h.
func NewScreen(w, h int) tcell.SimulationScreen {
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		panic(err)
	}
	s.SetSize(w, h)
	return s
}

// Text shows the screen and reads it back as plain text, trailing spaces
// trimmed.
func Text(s tcell.SimulationScreen) string {
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
func Frame(s tcell.SimulationScreen) string {
	s.Show()
	cells, w, h := s.GetContents()
	var b strings.Builder
	for y := 0; y < h; y++ {
		var run strings.Builder
		curFg, curBg := tcell.ColorDefault, tcell.ColorDefault
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
			bd := attr&tcell.AttrBold != 0
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

// Fonts is the CSS font stack for ordinary captures; NerdFonts puts Nerd
// Font Mono faces first for the nerd-tier captures.
const (
	Fonts     = `ui-monospace,Menlo,Consolas,"DejaVu Sans Mono",monospace`
	NerdFonts = `"JetBrainsMono Nerd Font Mono","Hack Nerd Font Mono","FiraCode Nerd Font Mono","MesloLGS NF","Symbols Nerd Font Mono",ui-monospace,Menlo,monospace`
)

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
func Hex(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r&0xff, g&0xff, b&0xff)
}

// trimState drops the parts of a snapshot the maps never read.
func trimState(st game.GameState) game.GameState {
	st.History = nil
	st.AccountStats = nil
	st.Modifiers = nil
	st.Milestones = game.MilestoneState{}
	st.Research.Techs = nil
	if len(st.Log) > 12 {
		st.Log = st.Log[len(st.Log)-12:]
	}
	return st
}

// SaveState writes a snapshot as gzipped JSON.
func SaveState(path string, st game.GameState) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	zw := gzip.NewWriter(f)
	if err := json.NewEncoder(zw).Encode(trimState(st)); err != nil {
		f.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// LoadState reads a snapshot written by SaveState.
func LoadState(path string) (game.GameState, error) {
	var st game.GameState
	f, err := os.Open(path)
	if err != nil {
		return st, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return st, err
	}
	err = json.NewDecoder(zr).Decode(&st)
	return st, err
}
