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

// Captures are real terminal renders: the view draws into a tcell
// SimulationScreen exactly as it would into a terminal, and these helpers
// read the screen's cells back out as text or as HTML with inline colours.

type shot struct {
	w, h  int
	runes [][]rune
	fg    [][]tcell.Color
	bg    [][]tcell.Color
	bold  [][]bool
}

func renderShot(v *view, pal palette) shot {
	w, h := v.c.w, v.c.h
	scr := tcell.NewSimulationScreen("UTF-8")
	if err := scr.Init(); err != nil {
		panic(err)
	}
	defer scr.Fini()
	scr.SetSize(w, h)
	v.draw()
	v.c.blit(scr, pal)
	scr.Show()
	cells, sw, sh := scr.GetContents()
	s := shot{w: sw, h: sh}
	for y := 0; y < sh; y++ {
		var rs []rune
		var fs, bs []tcell.Color
		var bo []bool
		for x := 0; x < sw; x++ {
			c := cells[y*sw+x]
			r := ' '
			if len(c.Runes) > 0 {
				r = c.Runes[0]
			}
			fg, bg, attr := c.Style.Decompose()
			rs = append(rs, r)
			fs = append(fs, fg)
			bs = append(bs, bg)
			bo = append(bo, attr&tcell.AttrBold != 0)
		}
		s.runes = append(s.runes, rs)
		s.fg = append(s.fg, fs)
		s.bg = append(s.bg, bs)
		s.bold = append(s.bold, bo)
	}
	return s
}

func (s shot) text() string {
	var b strings.Builder
	for _, row := range s.runes {
		b.WriteString(strings.TrimRight(string(row), " "))
		b.WriteByte('\n')
	}
	return b.String()
}

func hex(c tcell.Color) string {
	r, g, b := c.RGB()
	if r < 0 {
		return "inherit"
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// styles interns inline styles as CSS classes, so an animated page of 16
// frames stays small.
type styles struct {
	ids map[string]int
	css strings.Builder
}

func newStyles() *styles { return &styles{ids: map[string]int{}} }

func (t *styles) class(st string) string {
	id, ok := t.ids[st]
	if !ok {
		id = len(t.ids)
		t.ids[st] = id
		fmt.Fprintf(&t.css, ".s%d{%s}\n", id, st)
	}
	return fmt.Sprintf("s%d", id)
}

// preHTML renders the shot as one <pre> of styled runs.
func (s shot) preHTML(id string, hidden bool, t *styles) string {
	var b strings.Builder
	disp := ""
	if hidden {
		disp = ` style="display:none"`
	}
	fmt.Fprintf(&b, `<pre class="term" id="%s"%s>`, id, disp)
	for y := range s.runes {
		x := 0
		for x < s.w {
			fg, bg, bo := s.fg[y][x], s.bg[y][x], s.bold[y][x]
			j := x
			var run strings.Builder
			for j < s.w && s.fg[y][j] == fg && s.bg[y][j] == bg && s.bold[y][j] == bo {
				run.WriteRune(s.runes[y][j])
				j++
			}
			st := fmt.Sprintf("color:%s;background:%s", hex(fg), hex(bg))
			if bo {
				st += ";font-weight:bold"
			}
			fmt.Fprintf(&b, `<span class="%s">%s</span>`, t.class(st), html.EscapeString(run.String()))
			x = j
		}
		b.WriteByte('\n')
	}
	b.WriteString("</pre>")
	return b.String()
}

func page(title string, bg tcell.Color, body string, t *styles) string {
	return fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"><title>%s</title>
<style>
body{margin:0;padding:16px;background:%s;}
.term{font-family:Menlo,"SF Mono","DejaVu Sans Mono",Consolas,monospace;font-size:13px;line-height:1.0;margin:0;display:inline-block;transform-origin:0 0}
.term span{white-space:pre}
.cap{font:12px -apple-system,Segoe UI,sans-serif;color:#888;margin:0 0 8px}
%s</style></head><body>
<p class="cap">%s</p>
<div id="wrap">%s</div>
<script>
// Scale the terminal to the window width (never up), so a 160-column
// capture fits a laptop screen.
function fit(){if(location.hash=='#full')return;const w=document.getElementById('wrap');let pw=0,ph=0;
for(const p of w.querySelectorAll('.term')){p.style.transform='';if(p.offsetWidth>pw){pw=p.offsetWidth;ph=p.offsetHeight;}}
const s=Math.min(1,(innerWidth-32)/pw);for(const p of w.querySelectorAll('.term'))p.style.transform='scale('+s+')';
w.style.height=(ph*s)+'px';}
addEventListener('resize',fit);addEventListener('load',fit);
</script>
</body></html>
`, html.EscapeString(title), hex(bg), t.css.String(), html.EscapeString(title), body)
}

// writeCapture writes <base>.txt and <base>.html for one frame.
func writeCapture(dir, base, title string, v *view, t theme.Theme, mono bool) error {
	pal := newPalette(t, v.s, mono)
	s := renderShot(v, pal)
	if err := os.WriteFile(filepath.Join(dir, base+".txt"), []byte(s.text()), 0o644); err != nil {
		return err
	}
	st := newStyles()
	pre := s.preHTML("f0", false, st)
	return os.WriteFile(filepath.Join(dir, base+".html"), []byte(page(title, t.Color(theme.RoleBackground), pre, st)), 0o644)
}

// writeAnim writes <base>-anim.html (a JS frame loop over n frames) and
// <base>-frames.txt (the first four frames, one after another).
func writeAnim(dir, base, title string, v *view, t theme.Theme, n int) error {
	pal := newPalette(t, v.s, false)
	var pres, txt strings.Builder
	st := newStyles()
	start := v.frame
	for i := 0; i < n; i++ {
		v.frame = start + i
		s := renderShot(v, pal)
		pres.WriteString(s.preHTML(fmt.Sprintf("f%d", i), i > 0, st))
		pres.WriteByte('\n')
		if i < 4 {
			fmt.Fprintf(&txt, "── frame %d ──\n%s\n", i, s.text())
		}
	}
	v.frame = start
	js := fmt.Sprintf(`<script>
let i=0;const n=%d;setInterval(()=>{document.getElementById('f'+i).style.display='none';i=(i+1)%%n;document.getElementById('f'+i).style.display='inline-block';},140);
</script>`, n)
	if err := os.WriteFile(filepath.Join(dir, base+"-anim.html"), []byte(page(title+" (animated)", t.Color(theme.RoleBackground), pres.String()+js, st)), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, base+"-frames.txt"), []byte(txt.String()), 0o644)
}
