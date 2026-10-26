package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// Template is a wonder's hand-drawn shape. Art is drawn as written; Tone has
// the same shape and names the material of each cell by a letter the age
// style colours (Tones). A missing Tone means stone ('s' throughout).
type Template struct {
	Art, Tone []string
}

// templates are the wonders drawn so far, by wonder key. A wonder without
// one is drawn as the generic monument.
var templates = map[string]Template{
	"stonehenge": {
		Art: []string{
			"  Π Π Π  ",
			" Π     Π ",
			"Π   ▄   Π",
			" Π     Π ",
			"  Π Π Π  ",
		},
		Tone: []string{
			"  s s s  ",
			" s     s ",
			"s   a   s",
			" s     s ",
			"  s s s  ",
		},
	},
	"space_program": {
		Art: []string{
			"    ▲    ",
			"   ╱█╲   ",
			"  ╟│█│╢  ",
			" ╤═╧█╧═╤ ",
			"▄▄▄▄▄▄▄▄▄",
		},
		Tone: []string{
			"    a    ",
			"   sgs   ",
			"  msgsm  ",
			" mmmsmmm ",
			"sssssssss",
		},
	},
	"dyson_scaffold": {
		Art: []string{
			" ╭─┬─┬─╮ ",
			"╱╲ ╲ ╱ ╱╲",
			"│ ╲ ☼ ╱ │",
			"╲╱ ╱ ╲ ╲╱",
			" ╰─┴─┴─╯ ",
		},
		Tone: []string{
			" mmmmmmm ",
			"mmm m mmm",
			"m m f m m",
			"mmm m mmm",
			" mmmmmmm ",
		},
	},
}

// monuments stand in for wonders with no template of their own; one is
// picked by the wonder's key, so neighbours do not all look alike.
var monuments = []Template{
	{
		Art:  []string{"  ▲  ", " ╱█╲ ", " │█│ ", "▄███▄"},
		Tone: []string{"  a  ", " sas ", " sss ", "sssss"},
	},
	{
		Art:  []string{"▄▄▄▄▄▄▄", "█     █", "█     █", "█     █"},
		Tone: []string{"aaaaaaa", "s     s", "s     s", "s     s"},
	},
	{
		Art:  []string{"  ▄▄▄  ", " ▟███▙ ", " █▪▪▪█ ", " ▀▀▀▀▀ "},
		Tone: []string{"  aaa  ", " sssss ", " sgggs ", " sssss "},
	},
	{
		Art:  []string{"   ▲   ", "  ▟█▙  ", " ▟███▙ ", "▟█████▙"},
		Tone: []string{"   a   ", "  sss  ", " sssss ", "sssssss"},
	},
}

func (t Template) size() (w, h int) { return Art(t.Art).size() }

// paintWonder stamps a wonder on its plot, finished or rising: an unfinished
// wonder shows from the ground up as far as its progress, under scaffolding.
func (f *frame) paintWonder(ws wonderSpot) {
	t, ok := templates[ws.w.Key]
	if !ok {
		t = monuments[roll(len(monuments), mapmodel.HashStr(ws.w.Key))]
	}
	w, h := t.size()
	r := ws.plot.Rect
	x0, y0 := anchor(Lot{Rect: r, Door: Pt{r.X + r.W/2, r.Y + r.H}}, w, h)
	built := h
	if !ws.w.Built {
		built = (h*int(ws.w.Progress*100) + 99) / 100
	}
	thing := Thing{Kind: KindWonder, Key: ws.w.Key}
	f.targets = append(f.targets, middleOf(t.Art, x0, y0))
	for dy := h - built; dy < h; dy++ {
		for dx, ch := range []rune(t.Art[dy]) {
			if ch == ' ' {
				continue
			}
			tone := byte('s')
			if t.Tone != nil && dx < len(t.Tone[dy]) {
				tone = t.Tone[dy][dx]
			}
			ink, ok := f.st.Tones[tone]
			if !ok {
				ink = P(f.st.Wall, 235)
			}
			f.g.draw(x0+dx, y0+dy, ch, ink)
			f.g.mark(x0+dx, y0+dy, thing)
			if tone == 'g' || tone == 'f' {
				f.glow = append(f.glow, Pt{x0 + dx, y0 + dy})
			}
		}
	}
	if built < h {
		for dx := 0; dx < w; dx += 2 {
			f.g.draw(x0+dx, y0+h-built-1, '╫', f.st.Tones['a'])
			f.g.mark(x0+dx, y0+h-built-1, thing)
		}
	}
}
