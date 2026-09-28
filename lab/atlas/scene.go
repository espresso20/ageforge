package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// Scene is the atlas view: the model, the plate being shown, the camera,
// the cursor and what is selected. Draw lays out chrome around the map.

type Scene struct {
	A        *Atlas
	Plate    *Plate
	View     View
	fitted   bool
	CurX     int // cursor in map-rect cells
	CurY     int
	Frame    int
	Legend   bool
	Mono     bool
	Verb     string // the command line a verb key produced, shown in the status bar
	mapRect  Rect
	last     *earthCtx
	cosmic   *cosmicCtx
	poiIdx   int
	lastSize [2]int
}

type POI struct {
	Kind string // capital, civ, route, trail, omen, scar, wonder
	Name string
	At   Pt
	Ref  int
}

func newScene(a *Atlas) *Scene {
	return &Scene{A: a, Plate: plates[plateForAge(a.St.Age)], Legend: true}
}

// layout decides the frame: full (side panel), narrow (bottom inspector)
// or mini (a sidebar-sized locator with one status line).
func (s *Scene) layout(w, h int) (mapR, panel, insp Rect, mode string) {
	switch {
	case w <= 60 || h <= 18:
		return Rect{0, 1, w, h - 2}, Rect{}, Rect{}, "mini"
	case w >= 120 && s.Legend:
		pw := 34
		if w >= 150 {
			pw = 38
		}
		return Rect{0, 0, w - pw, h - 1}, Rect{w - pw, 0, pw, h - 1}, Rect{}, "full"
	}
	return Rect{0, 0, w, h - 5}, Rect{}, Rect{0, h - 5, w, 4}, "narrow"
}

// inner is the map rect inside its frame.
func (s *Scene) inner(r Rect, mode string) Rect {
	if mode == "mini" || s.Plate.Frame == "hide" {
		return r
	}
	return Rect{r.X + 1, r.Y + 1, r.W - 2, r.H - 2}
}

// fit points the camera at the known world (or, in the mini view, the
// realm), at a scale that fills the rect.
func (s *Scene) fit(rc Rect, mode string) {
	a := s.A
	if s.Plate.Model != "earth" {
		s.View = View{0, 0, cosmicFitScale(s.Plate, rc)}
		s.CurX, s.CurY = rc.W/2, rc.H/2
		return
	}
	minX, minY, maxX, maxY := a.Capital.X, a.Capital.Y, a.Capital.X, a.Capital.Y
	add := func(p Pt, r float64) {
		minX, maxX = math.Min(minX, p.X-r), math.Max(maxX, p.X+r)
		minY, maxY = math.Min(minY, p.Y-r), math.Max(maxY, p.Y+r)
	}
	switch {
	case mode == "mini":
		add(a.Capital, 14+a.Camp)
		for i := 0; i < a.Owned; i++ {
			add(a.W.Regions[a.Claim[i]].Seed, 14)
		}
	case a.AgeIdx >= 12:
		add(Pt{worldW / 2, worldH / 2}, 0)
		add(Pt{8, 8}, 0)
		add(Pt{worldW - 8, worldH - 8}, 0)
	case a.AgeIdx < 3:
		// the Stone Era shows the edge of the known: the dark beyond it
		add(a.Capital, a.knownR*1.15)
	default:
		add(a.Capital, a.knownR*0.85)
		for _, c := range a.Civs {
			if c.Known && !c.Celestial {
				add(c.Seat, 12)
			}
		}
		for i := 0; i < a.Owned; i++ {
			add(a.W.Regions[a.Claim[i]].Seed, 10)
		}
	}
	minX, minY = math.Max(minX, 0), math.Max(minY, 0)
	maxX, maxY = math.Min(maxX, worldW), math.Min(maxY, worldH)
	S := math.Max((maxX-minX)/float64(rc.W), (maxY-minY)/float64(2*rc.H)) * 1.04
	S = clamp(S, 0.25, 4)
	s.View = View{(minX + maxX) / 2, (minY + maxY) / 2, S}
	cx, cy := s.toCell(a.Capital, rc)
	s.CurX, s.CurY = cx, cy
}

func (s *Scene) toCell(p Pt, rc Rect) (int, int) {
	return int(math.Floor((p.X-s.View.CX)/s.View.S + float64(rc.W)/2)), int(math.Floor((p.Y-s.View.CY)/(2*s.View.S) + float64(rc.H)/2))
}

func (s *Scene) toWorld(x, y int, rc Rect) Pt {
	return Pt{s.View.CX + (float64(x)+0.5-float64(rc.W)/2)*s.View.S, s.View.CY + (float64(y)+0.5-float64(rc.H)/2)*2*s.View.S}
}

// Draw renders the whole scene into a w×h canvas.
func (s *Scene) Draw(w, h int) *Canvas {
	pal := newPal(s.Plate)
	cv := newCanvas(w, h, pal.Bg)
	mapR, panel, insp, mode := s.layout(w, h)
	in := s.inner(mapR, mode)
	if !s.fitted || s.lastSize != [2]int{w, h} {
		s.fit(in, mode)
		s.fitted = true
		s.lastSize = [2]int{w, h}
	}
	s.mapRect = in
	s.CurX = int(clamp(float64(s.CurX), 0, float64(in.W-1)))
	s.CurY = int(clamp(float64(s.CurY), 0, float64(in.H-1)))
	plate := s.Plate
	if mode == "mini" {
		plate = plate.mini()
	}
	if s.Plate.Model == "earth" {
		s.last = drawEarth(cv, in, s.A, plate, pal, s.View, s.Frame)
		s.cosmic = nil
	} else {
		s.cosmic = drawCosmic(cv, in, s.A, plate, pal, s.View, s.Frame)
		s.last = nil
	}
	if mode != "mini" {
		s.decorate(cv, in, pal)
		s.frameChrome(cv, mapR, in, pal)
	}
	s.cursor(cv, in, pal, mode)
	switch mode {
	case "full":
		s.sidePanel(cv, panel, pal)
		s.statusLine(cv, Rect{0, h - 1, w, 1}, pal, mode)
	case "narrow":
		s.inspectorStrip(cv, insp, pal)
		s.statusLine(cv, Rect{0, h - 1, w, 1}, pal, mode)
	case "mini":
		s.miniChrome(cv, w, h, pal)
	}
	if s.Mono {
		for i := range cv.C {
			c := &cv.C[i]
			if c.Fg != c.Bg {
				c.Fg = pal.Ink
			}
			c.Bg = pal.Bg
		}
	}
	return cv
}

// ---------------------------------------------------------------------------
// map furniture: cartouche, compass, scale bar

func (s *Scene) decorate(cv *Canvas, in Rect, pal *Pal) {
	p, a := s.Plate, s.A
	home := a.HomeName
	title := p.Title
	if strings.Contains(title, "%s") {
		title = fmt.Sprintf(title, home)
	}
	sub := p.Sub
	if strings.Contains(sub, "%d") {
		sub = fmt.Sprintf(sub, 1+int(hashStr(home)%40))
	} else if strings.Contains(sub, "%s") {
		sub = fmt.Sprintf(sub, a.Towns[0].Name)
	}
	third := fmt.Sprintf("%s · %s", config.AgeByKey()[a.St.Age].Name, epochName(a.Epoch))
	if p.Frame == "hide" {
		cv.Text(in.X+2, in.Y, title, pal.Ink, tcell.AttrItalic|tcell.AttrBold, zChrome)
		cv.Text(in.X+2, in.Y+1, sub, pal.Faint, tcell.AttrItalic, zChrome)
		cv.Reserve(in.X, in.Y, uniseg.StringWidth(title)+4, 2)
	} else if in.W > 50 && in.W < 100 {
		// a narrow map gets a one-line title slip instead of a cartouche
		t := " " + trunc(title, in.W/2) + " "
		cv.Fill(in.X+1, in.Y, uniseg.StringWidth(t), 1, pal.Paper, zChrome)
		cv.Text(in.X+1, in.Y, t, pal.Accent, tcell.AttrBold|tcell.AttrReverse, zChrome)
		cv.Reserve(in.X+1, in.Y, uniseg.StringWidth(t), 1)
	} else if in.W >= 100 {
		if p.Model == "earth" && p.Frame != "hud" {
			title = strings.ToUpper(title)
		}
		wd := max3(uniseg.StringWidth(title), uniseg.StringWidth(sub), uniseg.StringWidth(third)) + 4
		if wd > in.W/2 {
			wd = in.W / 2
		}
		x0, y0 := in.X+1, in.Y
		box := []rune("╭─╮│╰╯")
		if p.Frame == "hud" || p.Frame == "chart" {
			box = []rune("┌─┐│└┘")
		}
		cv.Fill(x0, y0, wd, 5, pal.Paper, zChrome)
		for x := x0; x < x0+wd; x++ {
			cv.Put(x, y0, box[1], pal.Border, zChrome)
			cv.Put(x, y0+4, box[1], pal.Border, zChrome)
		}
		for y := y0; y <= y0+4; y++ {
			cv.Put(x0, y, box[3], pal.Border, zChrome)
			cv.Put(x0+wd-1, y, box[3], pal.Border, zChrome)
		}
		cv.Put(x0, y0, box[0], pal.Border, zChrome)
		cv.Put(x0+wd-1, y0, box[2], pal.Border, zChrome)
		cv.Put(x0, y0+4, box[4], pal.Border, zChrome)
		cv.Put(x0+wd-1, y0+4, box[5], pal.Border, zChrome)
		cv.Text(x0+2, y0+1, trunc(title, wd-4), pal.Accent, tcell.AttrBold, zChrome)
		cv.Text(x0+2, y0+2, trunc(sub, wd-4), pal.Faint, tcell.AttrItalic, zChrome)
		cv.Text(x0+2, y0+3, trunc(third, wd-4), pal.Label, 0, zChrome)
		cv.Reserve(x0, y0, wd, 5)
	}
	// compass
	var art []string
	switch p.Frame {
	case "neat":
		art = []string{"  N  ", "  ▲  ", "W─✛─E", "  ▼  ", "  S  "}
	case "survey":
		art = []string{"▲", "N", "│"}
	case "hud":
		art = []string{"N", "↑"}
	case "hide":
		if x, y := in.X+in.W-4, in.Y+in.H/2; true {
			cv.PutA(x, y, '☼', pal.High, tcell.AttrBold, zChrome)
			cv.Text(x-8, y+1, "sunrise", pal.Faint, tcell.AttrItalic, zChrome)
		}
	}
	if art != nil && in.W > 40 {
		cv.Stamp(in.X+in.W-5, in.Y+3, art, pal.Ink, zChrome)
	}
	// scale bar
	if in.W > 40 && p.UnitPer > 0 {
		units := 14 * s.View.S * p.UnitPer
		nice := niceNum(units)
		cells := int(nice / (s.View.S * p.UnitPer))
		if cells >= 4 && cells < in.W/2 {
			x0, y0 := in.X+2, in.Y+in.H-2
			if p.Frame == "hide" {
				y0 = in.Y + in.H - 3
			}
			cv.Fill(x0-1, y0, cells+uniseg.StringWidth(p.Unit)+10, 1, pal.Paper, zChrome)
			half := cells / 2
			for i := 0; i <= cells; i++ {
				g := '─'
				if p.Frame == "neat" || p.Frame == "survey" {
					g = '▀'
					if i >= half {
						g = '▄'
					}
				}
				switch i {
				case 0:
					g = '├'
				case half:
					g = '┼'
				case cells:
					g = '┤'
				}
				cv.Put(x0+i, y0, g, pal.Ink, zChrome)
			}
			cv.Text(x0+cells+2, y0, fmt.Sprintf("%s %s", fmtNum(nice), p.Unit), pal.Label, 0, zChrome)
		}
	}
}

func epochName(k string) string {
	for _, e := range config.Epochs() {
		if e.Key == k {
			return e.Name
		}
	}
	return k
}

func niceNum(v float64) float64 {
	if v <= 0 {
		return 1
	}
	e := math.Pow(10, math.Floor(math.Log10(v)))
	f := v / e
	switch {
	case f < 1.5:
		return e
	case f < 3.5:
		return 2 * e
	case f < 7.5:
		return 5 * e
	}
	return 10 * e
}

func fmtNum(v float64) string {
	if v >= 1 {
		return fmt.Sprintf("%.0f", v)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
}

func max3(a, b, c int) int {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}

func trunc(s string, n int) string {
	if uniseg.StringWidth(s) <= n {
		return s
	}
	r := []rune(s)
	if n < 1 {
		return ""
	}
	return string(r[:n-1]) + "…"
}

// frameChrome draws the plate's border around the map.
func (s *Scene) frameChrome(cv *Canvas, r, in Rect, pal *Pal) {
	p := s.Plate
	put := func(x, y int, g rune, c tcell.Color) { cv.Put(x, y, g, c, zChrome) }
	switch p.Frame {
	case "hide":
		// cut the map to the shape of a stretched hide, stitched at the edge
		for j := 0; j < in.H; j++ {
			for i := 0; i < in.W; i++ {
				u := (float64(i) + 0.5 - float64(in.W)/2) / (float64(in.W) / 2)
				v := (float64(j) + 0.5 - float64(in.H)/2) / (float64(in.H) / 2)
				ang := math.Atan2(v, u)
				lim := 1 + 0.07*math.Sin(ang*5+1) + 0.05*math.Sin(ang*9)
				d := math.Pow(math.Abs(u), 2.6) + math.Pow(math.Abs(v), 2.6)
				x, y := in.X+i, in.Y+j
				if d > lim {
					c := cv.At(x, y)
					if c.Z < zChrome {
						*c = Cell{R: ' ', Fg: pal.Bg, Bg: pal.Bg, Z: zChrome - 1}
					}
				} else if d > lim*0.9 {
					c := cv.At(x, y)
					if c.Z < zMark {
						c.R, c.Fg, c.Z = '·', pal.mix(pal.Paper, pal.Relief, 0.8), zBorder
						if (i+j)%2 == 0 {
							c.R = ' '
						}
					}
				}
			}
		}
	case "double", "neat", "survey", "chart":
		h, v, tl, tr, bl, br := '═', '║', '╔', '╗', '╚', '╝'
		if p.Frame == "neat" {
			h, v, tl, tr, bl, br = '━', '┃', '┏', '┓', '┗', '┛'
		}
		if p.Frame == "survey" || p.Frame == "chart" {
			h, v, tl, tr, bl, br = '─', '│', '┌', '┐', '└', '┘'
		}
		for x := r.X; x < r.X+r.W; x++ {
			put(x, r.Y, h, pal.Border)
			put(x, r.Y+r.H-1, h, pal.Border)
		}
		for y := r.Y; y < r.Y+r.H; y++ {
			put(r.X, y, v, pal.Border)
			put(r.X+r.W-1, y, v, pal.Border)
		}
		put(r.X, r.Y, tl, pal.Border)
		put(r.X+r.W-1, r.Y, tr, pal.Border)
		put(r.X, r.Y+r.H-1, bl, pal.Border)
		put(r.X+r.W-1, r.Y+r.H-1, br, pal.Border)
		// graticule labels on the frame
		if p.Graticule > 0 && p.Model == "earth" {
			g := p.Graticule
			lastEnd := -99
			for i := 0; i < in.W; i++ {
				a := s.toWorld(i, 0, in).X - s.View.S/2
				b := a + s.View.S
				if math.Floor(a/g) != math.Floor(b/g) {
					val := math.Floor(b/g) * g
					var lab string
					if p.GridRefs {
						lab = fmt.Sprintf("%02d", int(val/g)%100)
					} else {
						lon := val - worldW/2
						hemi := "E"
						if lon < 0 {
							hemi, lon = "W", -lon
						}
						lab = fmt.Sprintf("%.0f°%s", lon, hemi)
						if lon == 0 {
							lab = "0°"
						}
					}
					x := in.X + i
					if x-1 > lastEnd && x+1+len([]rune(lab)) < r.X+r.W-1 {
						cv.Put(x, r.Y, '┬', pal.Border, zChrome)
						cv.Text(x+1, r.Y+r.H-1, lab, pal.Label, 0, zChrome)
						lastEnd = x + len(lab)
					}
				}
			}
			for j := 0; j < in.H; j++ {
				a := s.toWorld(0, j, in).Y - s.View.S
				b := a + 2*s.View.S
				if math.Floor(a/g) != math.Floor(b/g) {
					val := math.Floor(b/g) * g
					var lab string
					if p.GridRefs {
						lab = fmt.Sprintf("%02d", int(val/g)%100)
					} else {
						lat := (worldH/2 - val) * 0.8
						hemi := "N"
						if lat < 0 {
							hemi, lat = "S", -lat
						}
						lab = fmt.Sprintf("%.0f%s", lat, hemi)
					}
					cv.Put(r.X, in.Y+j, '┤', pal.Border, zChrome)
					cv.Text(r.X+r.W-1-len([]rune(lab)), in.Y+j, lab, pal.Label, 0, zChrome+1)
					_ = lab
				}
			}
		}
	case "hud":
		c := pal.Pos
		for k := 0; k < 6; k++ {
			put(r.X+k, r.Y, '─', c)
			put(r.X+r.W-1-k, r.Y, '─', c)
			put(r.X+k, r.Y+r.H-1, '─', c)
			put(r.X+r.W-1-k, r.Y+r.H-1, '─', c)
		}
		for k := 0; k < 3; k++ {
			put(r.X, r.Y+k, '│', c)
			put(r.X+r.W-1, r.Y+k, '│', c)
			put(r.X, r.Y+r.H-1-k, '│', c)
			put(r.X+r.W-1, r.Y+r.H-1-k, '│', c)
		}
		put(r.X, r.Y, '┌', c)
		put(r.X+r.W-1, r.Y, '┐', c)
		put(r.X, r.Y+r.H-1, '└', c)
		put(r.X+r.W-1, r.Y+r.H-1, '┘', c)
		stamp := fmt.Sprintf(" PASS %04d · T+%s ", s.A.St.Tick%10000, clock(s.A.St.Tick+s.Frame*20))
		cv.Text(r.X+r.W-2-uniseg.StringWidth(stamp), r.Y, stamp, pal.Pos, 0, zChrome)
		if p.Scan {
			// the downlink scanline sweeps down the image
			y := in.Y + (s.Frame*2)%in.H
			for x := in.X; x < in.X+in.W; x++ {
				cv.SetBg(x, y, pal.mix(pal.Paper, pal.Pos, 0.18))
			}
			cv.Text(r.X+1, y, "▸", pal.Pos, tcell.AttrBold, zChrome)
		}
	}
}

func clock(t int) string {
	sec := t * 2
	return fmt.Sprintf("%02d:%02d:%02d", (sec/3600)%24, (sec/60)%60, sec%60)
}

// ---------------------------------------------------------------------------
// cursor, selection, inspector

func (s *Scene) cursor(cv *Canvas, in Rect, pal *Pal, mode string) {
	if mode == "mini" {
		return
	}
	x, y := in.X+s.CurX, in.Y+s.CurY
	c := cv.At(x, y)
	if c == nil {
		return
	}
	c.Attr |= tcell.AttrReverse
	c.Z = zCursor
	if c.R == ' ' {
		c.R, c.Fg = '+', pal.Accent
	}
	for _, d := range [][3]int{{-1, 0, '['}, {1, 0, ']'}} {
		if n := cv.At(x+d[0], y); n != nil && n.Z < zMark && x+d[0] >= in.X && x+d[0] < in.X+in.W {
			n.R, n.Fg, n.Z = rune(d[2]), pal.Accent, zCursor
			n.Attr = tcell.AttrBold
		}
	}
}

// pois lists what can be selected, in tab order.
func (s *Scene) pois() []POI {
	a := s.A
	var out []POI
	if s.Plate.Model != "earth" && s.cosmic != nil {
		return s.cosmic.pois
	}
	out = append(out, POI{"capital", a.Towns[0].Name, a.Capital, 0})
	for i, c := range a.Civs {
		if c.Known && !c.Celestial {
			out = append(out, POI{"civ", c.Def.Name, c.Seat, i})
		}
	}
	for i, r := range a.Routes {
		if len(r.Path) > 0 {
			out = append(out, POI{"route", r.Name, pointAlong(r.Path, 0.5), i})
		}
	}
	for i, o := range a.Omens {
		if o.Active {
			out = append(out, POI{"omen", o.Name, o.At, i})
		}
	}
	for i, t := range a.Trails {
		if t.Active {
			out = append(out, POI{"trail", t.Name, pointAlong(t.Path, t.Progress), i})
		}
	}
	for i, sc := range a.Scars {
		out = append(out, POI{"scar", sc.Name, sc.At, i})
	}
	for i, w := range a.Wonders {
		out = append(out, POI{"wonder", w.Name, w.At, i})
	}
	for i, t := range a.Towns[1:] {
		if t.Owner == "" {
			out = append(out, POI{"town", t.Name, t.At, i + 1})
		}
	}
	for i, t := range a.Trails {
		if !t.Active && s.Plate.Relief != "shade" && !s.Plate.Night {
			out = append(out, POI{"trail", t.Name, t.Path[len(t.Path)-1], i})
		}
	}
	return out
}

// selected is the POI under (or next to) the cursor, if any.
func (s *Scene) selected() (POI, bool) {
	best, bd := POI{}, 2.6
	for _, p := range s.pois() {
		x, y := s.toCell(p.At, s.mapRect)
		d := math.Hypot(float64(x-s.CurX), float64(y-s.CurY)*2)
		if d < bd {
			best, bd = p, d
		}
	}
	return best, bd < 2.6
}

// Next moves the cursor to the next point of interest, panning if needed.
func (s *Scene) Next(dir int) {
	ps := s.pois()
	if len(ps) == 0 {
		return
	}
	s.poiIdx = (s.poiIdx + dir + len(ps)) % len(ps)
	s.focus(ps[s.poiIdx].At)
}

func (s *Scene) focus(p Pt) {
	x, y := s.toCell(p, s.mapRect)
	if x < 2 || y < 2 || x >= s.mapRect.W-2 || y >= s.mapRect.H-2 {
		s.View.CX, s.View.CY = p.X, p.Y
		x, y = s.toCell(p, s.mapRect)
	}
	s.CurX, s.CurY = x, y
}

type line struct {
	s    string
	c    tcell.Color
	attr tcell.AttrMask
}

// inspect describes the selection, ending with the commands that act on it.
func (s *Scene) inspect(pal *Pal) []line {
	a := s.A
	p, ok := s.selected()
	var L []line
	add := func(str string, c tcell.Color, at tcell.AttrMask) { L = append(L, line{str, c, at}) }
	verb := func(cmd string) { add("» "+cmd, pal.Accent, 0) }
	if ok && s.cosmic != nil {
		return s.cosmic.inspect(p, pal)
	}
	if !ok {
		if s.cosmic != nil {
			add("empty space", pal.Faint, tcell.AttrItalic)
			return L
		}
		w := s.toWorld(s.CurX, s.CurY, s.mapRect)
		if a.Known(w.X, w.Y) < 0.5 {
			add("TERRA INCOGNITA", pal.Faint, tcell.AttrBold)
			add("uncharted: no one has been", pal.Faint, tcell.AttrItalic)
			add("here yet", pal.Faint, tcell.AttrItalic)
			if len(a.St.Military.Expeditions) > 0 {
				verb("expedition " + a.St.Military.Expeditions[0].Key)
			}
			return L
		}
		h := a.W.Height(w.X, w.Y)
		if h <= 0 {
			name := "open sea"
			bd := math.Inf(1)
			for _, sea := range a.W.Seas {
				if d := sea.At.Dist(w); d < 70 && d < bd {
					name, bd = sea.Name, d
				}
			}
			add(strings.ToUpper(name), pal.Water, tcell.AttrBold)
			add(fmt.Sprintf("depth ~%d fathoms", int(-h*400)+3), pal.Faint, 0)
			return L
		}
		rg := a.W.RegionAt(w.X, w.Y)
		if rg < 0 {
			return L
		}
		r := a.W.Regions[rg]
		add(strings.ToUpper(r.Name), pal.Ink, tcell.AttrBold)
		add(fmt.Sprintf("%s · elevation %d", a.W.Biome(w.X, w.Y, h), int(h*1850)), pal.Faint, 0)
		switch o := a.OwnerAt(w.X, w.Y); {
		case o == 0 && a.Owned == 0:
			add("your hunting grounds", pal.Accent, 0)
		case o == 0:
			idx := 0
			for i, c := range a.Claim {
				if c == rg {
					idx = i
				}
			}
			add(fmt.Sprintf("province of your realm (#%d)", idx+1), pal.Accent, 0)
			if a.FreshReg[rg] {
				add("annexed since your last visit", pal.High, tcell.AttrItalic)
			}
		case o > 0:
			c := a.Civs[o-1]
			add("held by the "+c.Def.Name, pal.Rel(relation(c)), 0)
		default:
			add("unclaimed", pal.Faint, tcell.AttrItalic)
		}
		return L
	}
	switch p.Kind {
	case "capital":
		st := a.St
		add(strings.ToUpper(p.Name), pal.Accent, tcell.AttrBold)
		prov := "a camp and its hunting grounds"
		if a.Owned > 0 {
			prov = plural(a.Owned, "province", "provinces")
		}
		add(fmt.Sprintf("capital · %s", prov), pal.Label, 0)
		add(fmt.Sprintf("%d buildings · %d wonders", a.Buildings, len(a.Wonders)), pal.Ink, 0)
		idle := ""
		if st.Workers.TotalIdle > 0 {
			idle = fmt.Sprintf(" · %d idle", st.Workers.TotalIdle)
		}
		add(fmt.Sprintf("pop %d/%d%s", st.Workers.TotalPop, st.Workers.MaxPop, idle), pal.Ink, 0)
		if st.Workers.TotalIdle > 0 {
			add(fmt.Sprintf("%d hands with nothing to do", st.Workers.TotalIdle), pal.Warn, tcell.AttrItalic)
			verb("assign <building> <n>")
		}
		if st.AgeReady {
			add(st.NextAgeName+" is within reach", pal.High, 0)
			verb("advance")
		}
	case "civ":
		c := a.Civs[p.Ref]
		f := c.Info
		add(strings.ToUpper(c.Def.Name), pal.Rel(relation(c)), tcell.AttrBold)
		add("seat: "+c.SeatName, pal.Faint, 0)
		add(fmt.Sprintf("%s · %s", f.Personality, stars(c.Def.Strength)), pal.Label, 0)
		add(fmt.Sprintf("standing %+d %s %s", f.Opinion, bar(f.Opinion), f.Status), pal.Rel(relation(c)), 0)
		add(fmt.Sprintf("trades %s · +%.0f%% allied", c.Def.Specialty, c.Def.TradeBonus*100), pal.Ink, 0)
		open, taken := 0, 0
		for _, d := range f.Deals {
			if d.Taken {
				taken++
			} else {
				open++
			}
		}
		if len(f.Deals) > 0 {
			add(fmt.Sprintf("deals: %d open · %d taken", open, taken), pal.Ink, 0)
			for _, d := range f.Deals {
				if !d.Taken {
					add(fmt.Sprintf(" %d. %s %s → %s %s", d.Num, short(d.GiveAmt), d.Give, short(d.GetAmt), d.Get), pal.Faint, 0)
					break
				}
			}
		} else if f.DealsBlocked != "" {
			add(trunc(f.DealsBlocked, 30), pal.Faint, tcell.AttrItalic)
		}
		if f.LentWorkers > 0 {
			add(fmt.Sprintf("lending you %d workers", f.LentWorkers), pal.Pos, 0)
		}
		switch {
		case f.AtWar:
			add("AT WAR · raiding your stores", pal.Neg, tcell.AttrBold)
			verb("diplomacy tribute " + c.Key)
		case f.Status == "embargo":
			add("under your embargo", pal.Warn, 0)
			verb("diplomacy neutral " + c.Key)
		case f.Opinion >= 50 && f.Status != "allied":
			verb("diplomacy ally " + c.Key)
		case open > 0:
			verb(fmt.Sprintf("diplomacy accept %s %d", c.Key, f.Deals[0].Num))
		default:
			verb("diplomacy gift " + c.Key)
		}
	case "route":
		r := a.Routes[p.Ref]
		add(strings.ToUpper(r.Name), pal.High, tcell.AttrBold)
		add("to "+r.Dest, pal.Label, 0)
		add("sends "+resList(r.Export), pal.Ink, 0)
		add("brings "+resList(r.Import), pal.Ink, 0)
		if r.Disrupted {
			add("BLOCKADED · income suspended", pal.Neg, tcell.AttrBold)
		}
		verb("trade route stop " + r.Key)
	case "trail":
		t := a.Trails[p.Ref]
		add(strings.ToUpper(t.Name), pal.Label, tcell.AttrBold)
		add(fmt.Sprintf("%s · sent %d times", t.Category, t.Runs), pal.Ink, 0)
		if t.Active {
			add(fmt.Sprintf("out now · %d%% of the way", int(t.Progress*100)), pal.High, 0)
		}
		verb("expedition " + t.Key)
	case "omen":
		o := a.Omens[p.Ref]
		add(strings.ToUpper(o.Name), pal.Neg, tcell.AttrBold)
		add("a harbinger speaks of "+o.Target, pal.Ink, tcell.AttrItalic)
		if o.Numeric {
			add(fmt.Sprintf("catastrophe odds %.0f%%", o.Prob*100), pal.Warn, 0)
		}
		verb("harbinger appease")
		verb("harbinger brace")
	case "scar":
		sc := a.Scars[p.Ref]
		add(strings.ToUpper(sc.Name), pal.Neg, tcell.AttrBold)
		add(epochName(sc.Epoch)+" catastrophe", pal.Faint, 0)
		if sc.Outcome != "" {
			add("you "+sc.Outcome, pal.Ink, tcell.AttrItalic)
		}
	case "wonder":
		add(strings.ToUpper(p.Name), pal.High, tcell.AttrBold)
		add("a wonder of your realm", pal.Faint, tcell.AttrItalic)
	case "town":
		t := a.Towns[p.Ref]
		add(strings.ToUpper(t.Name), pal.Ink, tcell.AttrBold)
		add("a town of your realm", pal.Faint, 0)
		if t.Fresh {
			add("founded since your last visit", pal.High, tcell.AttrItalic)
		}
	default:
		return s.cosmic.inspect(p, pal)
	}
	return L
}

func stars(n int) string {
	return strings.Repeat("★", n) + strings.Repeat("☆", 5-n)
}

func bar(op int) string {
	n := int(math.Round(float64(op+100) / 20))
	n = int(clamp(float64(n), 0, 10))
	return "▕" + strings.Repeat("█", n) + strings.Repeat("░", 10-n) + "▏"
}

func short(v float64) string {
	switch {
	case v >= 1e12:
		return fmt.Sprintf("%.1fT", v/1e12)
	case v >= 1e9:
		return fmt.Sprintf("%.1fB", v/1e9)
	case v >= 1e6:
		return fmt.Sprintf("%.1fM", v/1e6)
	case v >= 1e3:
		return fmt.Sprintf("%.1fK", v/1e3)
	}
	return fmt.Sprintf("%.0f", v)
}

func resList(m map[string]float64) string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	var parts []string
	for _, k := range ks {
		parts = append(parts, fmt.Sprintf("%s %s", short(m[k]), k))
	}
	return strings.Join(parts, ", ")
}

// legend lists the symbols the current plate uses, only those present.
func (s *Scene) legend(pal *Pal) []line {
	p, a := s.Plate, s.A
	var L []line
	item := func(g rune, c tcell.Color, text string) {
		L = append(L, line{string(g) + "  " + text, c, 0})
	}
	if p.Model != "earth" {
		if s.cosmic != nil {
			return s.cosmic.legend
		}
		return L
	}
	if a.AgeIdx < 3 {
		item(p.Capital, pal.Accent, "your camp")
		item(p.BorderRune, pal.Accent, "hunting grounds")
	} else {
		item(p.Capital, pal.Accent, "your capital")
		item(p.Town, pal.Accent, "your towns")
		item(p.BorderRune, pal.Accent, "your border")
	}
	civs := 0
	for _, c := range a.Civs {
		if c.Known && !c.Celestial {
			civs++
		}
	}
	if civs > 0 {
		item(p.CivCapital, pal.Label, "a people's seat")
		L = append(L, line{"   +ally  !rival  ✕war", pal.Faint, 0})
	}
	if len(a.Routes) > 0 {
		item(p.Ship, pal.High, "trade route, goods moving")
	}
	if len(a.Trails) > 0 && !p.Night && !p.Scan {
		item(p.TrailRune, pal.Label, "expedition track")
	}
	for _, c := range a.Civs {
		if c.Info.AtWar && c.Known {
			item(p.Front, pal.Neg, "war front")
			break
		}
	}
	for _, c := range a.Civs {
		if c.Info.Status == "embargo" && c.Known {
			item('▫', pal.Warn, "embargo blockade")
			break
		}
	}
	if len(a.Omens) > 0 {
		item(p.Omen, pal.Neg, "harbinger's omen")
	}
	if len(a.Scars) > 0 {
		item(p.Scar, pal.Neg, "catastrophe scar")
	}
	if len(a.Wonders) > 0 {
		item(p.Wonder, pal.High, "wonder")
	}
	if p.RiverH != ' ' {
		item(p.RiverH, pal.Water, "river")
	}
	switch p.Relief {
	case "marks", "hachure":
		item(p.LandMarks["mountains"]|firstNonZero(p.LandMarks["peaks"]), pal.Relief, "mountains")
	case "contour":
		item('╭', pal.Relief, "contour, 185 m")
	case "shade":
		item('▓', pal.Forest, "forest · ▒ grass · ░ dry")
	}
	if p.Night {
		item('•', pal.Light1, "city lights")
		item('◇', pal.High, "satellite")
	}
	switch p.Fog {
	case "dark":
		item('░', pal.Faint, "where knowledge ends")
	case "dragons":
		item('∿', pal.Neg, "terra incognita")
	case "blank":
		item('·', pal.Faint, "parts unknown")
	case "unsurveyed":
		item('╱', pal.Faint, "unsurveyed")
	}
	return L
}

func firstNonZero(r rune) rune { return r }

func (s *Scene) sidePanel(cv *Canvas, r Rect, pal *Pal) {
	cv.Fill(r.X, r.Y, r.W, r.H, pal.Bg, zChrome)
	for y := r.Y; y < r.Y+r.H; y++ {
		cv.Put(r.X, y, '│', pal.Border, zChrome)
	}
	x := r.X + 2
	y := r.Y
	wd := r.W - 3
	head := func(t string) {
		cv.Text(x, y, t, pal.Accent, tcell.AttrBold, zChrome)
		cv.Text(x+uniseg.StringWidth(t)+1, y, strings.Repeat("─", max(0, wd-uniseg.StringWidth(t)-1)), pal.Border, 0, zChrome)
		y++
	}
	write := func(l line) {
		if y < r.Y+r.H {
			cv.Text(x, y, trunc(l.s, wd), l.c, l.attr, zChrome)
			y++
		}
	}
	head("ATLAS")
	write(line{fmt.Sprintf("plate %d of %d · %s", plateNum(s.Plate.Key), earnedPlates(s.A.St.Age), s.Plate.Name), pal.Label, 0})
	known, total := 0, 0
	for _, c := range s.A.Civs {
		total++
		if c.Known {
			known++
		}
	}
	write(line{fmt.Sprintf("%d of %d peoples met · %d routes", known, total, len(s.A.Routes)), pal.Faint, 0})
	y++
	if len(s.A.Changes) > 0 {
		head("SINCE YOU LAST LOOKED")
		for _, c := range s.A.Changes {
			write(line{"+ " + c, pal.High, 0})
		}
		y++
	}
	head("INSPECT")
	for _, l := range s.inspect(pal) {
		write(l)
	}
	y++
	head("LEGEND")
	for _, l := range s.legend(pal) {
		if y >= r.Y+r.H-1 {
			break
		}
		write(l)
	}
}

func (s *Scene) inspectorStrip(cv *Canvas, r Rect, pal *Pal) {
	cv.Fill(r.X, r.Y, r.W, r.H, pal.Bg, zChrome)
	L := s.inspect(pal)
	// two columns of two lines, then changes at the right
	col := 0
	for i, l := range L {
		if i >= 8 {
			break
		}
		cx := r.X + 1 + col*(r.W/3)
		cy := r.Y + i%4
		cv.Text(cx, cy, trunc(l.s, r.W/3-1), l.c, l.attr, zChrome)
		if i%4 == 3 {
			col++
		}
	}
	if len(s.A.Changes) > 0 {
		cx := r.X + 2*(r.W/3) + 1
		cv.Text(cx, r.Y, "since you last looked", pal.Accent, tcell.AttrBold, zChrome)
		for i, c := range s.A.Changes {
			if i >= 3 {
				break
			}
			cv.Text(cx, r.Y+1+i, trunc("+ "+c, r.W/3-2), pal.High, 0, zChrome)
		}
	}
}

func (s *Scene) statusLine(cv *Canvas, r Rect, pal *Pal, mode string) {
	cv.Fill(r.X, r.Y, r.W, 1, pal.Chip, zChrome)
	w := s.toWorld(s.CurX, s.CurY, s.mapRect)
	left := " " + s.readout(w)
	if s.Verb != "" {
		left = " would run: " + s.Verb
	}
	keys := "←↑↓→ move  tab next  +/- zoom  0 fit  [ ] plates  ⏎ verb  g legend  q quit"
	if r.W < 110 {
		keys = "tab next  +/- zoom  [ ] plates  ⏎ verb  q"
	}
	cv.Text(r.X, r.Y, trunc(left, r.W-uniseg.StringWidth(keys)-3), pal.Ink, 0, zChrome+1)
	cv.Text(r.X+r.W-uniseg.StringWidth(keys)-1, r.Y, keys, pal.Faint, 0, zChrome+1)
	for i := range cv.C[r.Y*cv.W : r.Y*cv.W+r.W] {
		cv.C[r.Y*cv.W+i].Bg = pal.Chip
	}
}

// readout is the cursor position in the plate's own coordinates.
func (s *Scene) readout(w Pt) string {
	p, a := s.Plate, s.A
	switch p.Model {
	case "earth":
	default:
		if s.cosmic != nil {
			return s.cosmic.readout(s.CurX, s.CurY)
		}
		return ""
	}
	lon := w.X - worldW/2
	lat := (worldH/2 - w.Y) * 0.8
	d := w.Dist(a.Capital) * p.UnitPer
	switch p.Key {
	case "hide":
		return fmt.Sprintf("%s from the fire", walk(d))
	case "portolan":
		return fmt.Sprintf("%s leagues by the %s wind from %s", fmtNum(math.Round(d)), wind(a.Capital, w), a.Towns[0].Name)
	case "survey":
		return fmt.Sprintf("grid ref %03d %03d · %s km from %s", int(w.X*5)%1000, int((worldH-w.Y)*5)%1000, fmtNum(math.Round(d)), a.Towns[0].Name)
	case "satellite", "radar":
		return fmt.Sprintf("LAT %+.2f  LON %+.2f  RNG %s KM", lat, lon, fmtNum(math.Round(d)))
	}
	return fmt.Sprintf("%s, %s · %s %s from %s", dms(lat, "N", "S"), dms(lon, "E", "W"), fmtNum(math.Round(d)), p.Unit, a.Towns[0].Name)
}

func walk(d float64) string {
	days := d / 8 * 1.0
	switch {
	case days < 0.6:
		return "less than a day's walk"
	case days < 1.5:
		return "a day's walk"
	}
	return fmt.Sprintf("%.0f days' walk", days)
}

func dms(v float64, pos, neg string) string {
	h := pos
	if v < 0 {
		h, v = neg, -v
	}
	return fmt.Sprintf("%d°%02d′%s", int(v), int((v-math.Floor(v))*60), h)
}

func wind(from, to Pt) string {
	names := []string{"East", "East by South", "Southeast", "South by East", "South", "South by West", "Southwest", "West by South",
		"West", "West by North", "Northwest", "North by West", "North", "North by East", "Northeast", "East by North"}
	ang := math.Atan2(to.Y-from.Y, to.X-from.X)
	k := int(math.Round(ang/(math.Pi/8)+16)) % 16
	return names[k]
}

func (s *Scene) miniChrome(cv *Canvas, w, h int, pal *Pal) {
	a := s.A
	cv.Fill(0, 0, w, 1, pal.Chip, zChrome)
	title := fmt.Sprintf(" %s · %s", strings.ToUpper(a.Towns[0].Name), config.AgeByKey()[a.St.Age].Name)
	cv.Text(0, 0, trunc(title, w), pal.Accent, tcell.AttrBold, zChrome+1)
	cv.Fill(0, h-1, w, 1, pal.Chip, zChrome)
	// the one thing most worth knowing, glyph first
	msg, col := "", pal.Label
	for _, c := range a.Civs {
		if c.Known && c.Info.AtWar {
			msg, col = fmt.Sprintf("%c war · %s", s.Plate.Front, c.Def.Name), pal.Neg
		}
	}
	if msg == "" {
		for _, o := range a.Omens {
			if o.Active {
				msg, col = fmt.Sprintf("%c %s", s.Plate.Omen, o.Name), pal.Neg
				if o.Numeric {
					msg += fmt.Sprintf(" %.0f%%", o.Prob*100)
				}
			}
		}
	}
	if msg == "" && len(a.Changes) > 0 {
		msg, col = "+ "+a.Changes[0], pal.High
	}
	if msg == "" {
		msg = fmt.Sprintf("%d prov · %d routes · %d met", a.Owned, len(a.Routes), countKnown(a))
	}
	cv.Text(1, h-1, trunc(msg, w-2), col, 0, zChrome+1)
	for i := 0; i < w; i++ {
		cv.C[i].Bg = pal.Chip
		cv.C[(h-1)*w+i].Bg = pal.Chip
	}
}

func countKnown(a *Atlas) int {
	n := 0
	for _, c := range a.Civs {
		if c.Known {
			n++
		}
	}
	return n
}

func plateNum(key string) int {
	for i, k := range plateOrder {
		if k == key {
			return i + 1
		}
	}
	return 0
}

// earnedPlates is how many plates the atlas holds at this age.
func earnedPlates(age string) int { return plateNum(plateForAge(age)) }
