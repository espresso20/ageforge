package main

import (
	"math"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
)

// render.go composes one frame, back to front: sky, sky objects, far
// ridge, near hills or the distant city, the districts, smoke, ground,
// water, traffic, weather, event overlays, then the chrome.

type View struct {
	W, H    int
	Cam     int     // world column at the scene's left edge
	TOD     float64 // time of day; <0 derives it from the game tick
	Frame   int     // animation frame (advances ~8/s)
	Theme   theme.Theme
	Colors  int
	Cursor  int // screen column of the inspect cursor, -1 for none
	Weather int // -1 derives it from events
	ShowNew bool
}

// dayTicks is one in-game day: an hour of play at 1x.
const dayTicks = 1800

func todFromTick(tick int) float64 { return float64((tick+dayTicks*3/8)%dayTicks) / dayTicks }

// sceneRows is how many rows the scene gets under the chrome.
func sceneRows(h int) int { return h - 3 }

type scene struct {
	w        *World
	v        View
	p        *Pal
	fb       *FB
	S        int // scene rows
	groundY  int
	top      int // screen row where the scene starts
	farTops  []float64
	prof     []int
	emitters []emitter
	sel      *Lot
	solid    []bool // cells a building covers (screen coords)
	labels   []pendingLabel
}

type pendingLabel struct {
	x, y int
	text string
	col  RGB
}

func render(w *World, v View) *FB {
	fb := newFB(v.W, v.H)
	S := sceneRows(v.H)
	tod := v.TOD
	if tod < 0 {
		tod = todFromTick(w.St.Tick)
	}
	weather := v.Weather
	if weather < 0 {
		weather = weatherFor(w)
	}
	d := Dials{AgeIdx: w.AgeIdx, Epoch: config.EpochForAge(w.St.Age), TOD: tod, Theme: v.Theme,
		Colors: v.Colors, Pressure: pressure(w), Weather: weather}
	sc := &scene{w: w, v: v, p: newPal(d), fb: fb, S: S, groundY: S - 3, top: 1}
	sc.sky()
	sc.celestial()
	sc.skyStructures()
	sc.clouds()
	sc.farRidge()
	sc.harbinger()
	sc.nearLayer()
	sc.lots()
	sc.smoke()
	sc.ground()
	sc.traffic()
	sc.weather()
	sc.catastrophe()
	sc.markers()
	sc.chrome()
	return fb
}

// Y converts a scene row to a screen row.
func (s *scene) Y(y int) int { return y + s.top }

func (s *scene) wx(x int) int { return x + s.v.Cam } // screen col -> world col

// ------------------------------------------------------------------ sky

// shade glyphs used to dither between two sky bands, BBS style.
var dither = []rune{' ', '░', '▒', '▓'}

func (s *scene) sky() {
	bands := 7
	for y := 0; y < s.groundY; y++ {
		f := float64(y) / float64(s.groundY-1) * float64(bands)
		b := int(f)
		t := f - float64(b)
		a := s.p.SkyAt(int(float64(b)/float64(bands)*float64(s.groundY-1)), s.groundY)
		c := s.p.SkyAt(int(float64(b+1)/float64(bands)*float64(s.groundY-1)), s.groundY)
		a, c = s.p.final(a), s.p.final(c)
		for x := 0; x < s.v.W; x++ {
			// the transition row between two bands gets a dither; others
			// are flat. A little per-column jitter keeps the bands from
			// reading as ruled lines.
			// solid for the first half of a band, then ░ ▒ ▓ into the next
			di := 0
			if t >= 0.52 {
				di = min(3, 1+int((t-0.52)/0.48*3))
			}
			switch {
			case di <= 0:
				s.fb.set(x, s.Y(y), ' ', c, a)
			case di >= 4:
				s.fb.set(x, s.Y(y), ' ', c, c)
			default:
				s.fb.set(x, s.Y(y), dither[di], c, a)
			}
		}
	}
}

func (s *scene) celestial() {
	p := s.p
	W := s.v.W
	horizon := float64(s.groundY - 2)
	// stars
	if p.Stars > 0.05 {
		for y := 0; y < int(horizon*0.8); y++ {
			for x := 0; x < W; x++ {
				hv := hashf(x, y, 991)
				if hv > 0.022*p.Stars*(1-float64(y)/horizon) {
					continue
				}
				tw := hash(x, y, s.v.Frame/6) % 9
				ch := '·'
				switch {
				case hv < 0.0015:
					ch = '*'
				case hv < 0.004:
					ch = '+'
				case tw == 0:
					ch = '∙'
				}
				c := hex(0xfff6e0).Lerp(hex(0x9ab8ff), hashf(x, y, 5))
				if tw == 1 {
					c = c.Mul(0.6)
				}
				bgc := s.fb.at(x, s.Y(y)).Bg
				s.fb.fg(x, s.Y(y), ch, p.final(bgc.Lerp(c, 0.4+0.6*p.Stars)))
			}
		}
	}
	// the sun rides an arc from dawn (east... left here) to dusk
	t := s.v.TOD
	if t < 0 {
		t = todFromTick(s.w.St.Tick)
	}
	arc := func(u float64) (int, int) {
		x := int(float64(W) * (0.08 + 0.84*u))
		y := int(horizon - math.Sin(u*math.Pi)*horizon*0.85)
		return x, y
	}
	disc := func(x, y int, core RGB, big bool) {
		c := p.final(core)
		if big { // a five-wide sun
			s.fb.fg(x-1, s.Y(y-1), '▄', c)
			s.fb.fg(x, s.Y(y-1), '▄', c)
			s.fb.fg(x+1, s.Y(y-1), '▄', c)
			s.fb.fg(x-2, s.Y(y), '▐', c)
			s.fb.set(x-1, s.Y(y), '█', c, c)
			s.fb.set(x, s.Y(y), '█', c, c)
			s.fb.set(x+1, s.Y(y), '█', c, c)
			s.fb.fg(x+2, s.Y(y), '▌', c)
			s.fb.fg(x-1, s.Y(y+1), '▀', c)
			s.fb.fg(x, s.Y(y+1), '▀', c)
			s.fb.fg(x+1, s.Y(y+1), '▀', c)
			return
		}
		s.fb.fg(x-1, s.Y(y), '▄', c)
		s.fb.fg(x, s.Y(y), '█', c)
		s.fb.fg(x+1, s.Y(y), '▄', c)
		s.fb.fg(x-1, s.Y(y+1), '▀', c)
		s.fb.fg(x, s.Y(y+1), '█', c)
		s.fb.fg(x+1, s.Y(y+1), '▀', c)
	}
	if u := (t - 0.23) / 0.54; u >= 0 && u <= 1 {
		x, y := arc(u)
		low := 1 - math.Sin(u*math.Pi)
		disc(x, y, hex(0xfffbe8).Lerp(hex(0xffa040), low*low), true)
	}
	// the moon, opposite
	mt := t + 0.5
	if mt >= 1 {
		mt--
	}
	if u := (mt - 0.23) / 0.54; u >= 0 && u <= 1 {
		x, y := arc(u)
		if y >= 1 {
			mc := hex(0xe8ecf8)
			if p.light {
				mc = hex(0xfaf8f0)
			}
			disc(x, y, mc, false)
		}
	}
}

// clouds drift slowly across; weather decides how many.
func (s *scene) clouds() {
	n := []int{3, 6, 9, 11}[s.p.D.Weather]
	if s.w.AgeIdx >= 18 {
		n /= 2
	}
	span := s.v.W + 60
	top := s.groundY / 2
	for i := 0; i < n; i++ {
		cw := 12 + int(hash(i, 41)%18)
		ch := 1 + int(hash(i, 43)%2)
		speed := 0.03 + hashf(i, 47)*0.05
		x0 := int(float64(hash(i, 49)%uint32(span))+float64(s.v.Frame)*speed-float64(s.v.Cam)*0.08) % span
		if x0 < 0 {
			x0 += span
		}
		x0 -= 30
		y0 := 1 + int(hash(i, 53)%uint32(max(1, top)))
		lit := s.p.Light
		base := hex(0xf4f4f8).Tint(lit)
		if s.p.Twilight > 0.3 {
			base = base.Lerp(hex(0xff9a80), s.p.Twilight*0.6)
		}
		if s.p.D.Weather >= 2 {
			base = base.Mul(0.7)
		}
		under := base.Mul(0.78)
		for dx := 0; dx < cw; dx++ {
			u := float64(dx) / float64(cw-1)
			bump := math.Sin(u*math.Pi) * (0.9 + 0.8*noise1(float64(dx)*0.5, i))
			hgt := bump * (1.1 + float64(ch)*0.55)
			x := x0 + dx
			for k := 0; k < int(math.Ceil(hgt)); k++ {
				y := y0 + ch - k
				if y < 0 || y >= s.groundY {
					continue
				}
				c := base
				if k == 0 {
					c = under
				}
				frac := hgt - float64(k)
				if frac < 0.5 {
					s.fb.fg(x, s.Y(y), '▄', s.p.final(c))
				} else {
					s.fb.set(x, s.Y(y), '█', s.p.final(c), s.p.final(c))
				}
			}
			// wispy fringe
			if bump > 0.2 {
				y := y0 + ch + 1
				if y < s.groundY && hash(x, i, 3)%2 == 0 {
					s.fb.fg(x, s.Y(y), '▀', s.p.final(under))
				}
			}
		}
	}
}

// ------------------------------------------------------------ landscape

func (s *scene) ridge(layer int, par float64, amp float64, freq float64, base int) []float64 {
	h := make([]float64, s.v.W)
	for x := range h {
		fx := float64(x) + float64(s.v.Cam)*par
		n := fbm(fx*freq, 300+layer*77+s.w.Seed%1000, 4)
		h[x] = float64(base) - math.Pow(n, 1.4)*amp
	}
	return h
}

func (s *scene) paintRidge(tops []float64, col, rim RGB) {
	for x, t := range tops {
		ti := int(math.Floor(t))
		for y := max(0, ti); y < s.groundY; y++ {
			s.fb.set(x, s.Y(y), '█', col, col)
		}
		if t-float64(ti) < 0.5 && ti-1 >= 0 {
			// the half-block rim anti-aliases the ridgeline
			s.fb.fg(x, s.Y(ti-1), '▄', rim)
		}
		_ = rim
	}
}

func (s *scene) farRidge() {
	amp := float64(s.groundY) * 0.42
	tops := s.ridge(0, 0.15, amp, 1.0/30, s.groundY-1)
	col := s.p.Hill(hex(0x4a5a7e), 0.42)
	if s.w.AgeIdx >= 18 {
		col = s.p.Hill(hex(0x5a4a8a), 0.42)
	}
	s.paintRidge(tops, col, col)
	// snow on the high peaks
	for x, t := range tops {
		if t < float64(s.groundY)-amp*0.72 {
			y := int(t)
			snow := s.p.Hill(hex(0xf0f4ff), 0.3)
			s.fb.set(x, s.Y(y), '▀', snow, col)
		}
	}
	s.farTops = tops
	s.factions(tops)
}

func (s *scene) nearLayer() {
	a := s.w.AgeIdx
	amp := float64(s.groundY) * 0.16
	tops := s.ridge(1, 0.4, amp, 1.0/18, s.groundY-1)
	var col RGB
	switch {
	case a <= 7:
		col = s.p.Hill(hex(0x2f5234), 0.22)
	case a <= 11:
		col = s.p.Hill(hex(0x4a4e4a), 0.22)
	default:
		col = s.p.Hill(hex(0x3a4a5a), 0.22)
	}
	s.paintRidge(tops, col, col)
	// forest on the hills while the world is young
	if a <= 9 {
		tc := s.p.Hill(hex(0x1e3a26), 0.15)
		for x, t := range tops {
			fx := x + int(float64(s.v.Cam)*0.4)
			if hash(fx, 5)%3 == 0 {
				y := int(t) - 1
				if y > 0 {
					s.fb.fg(x, s.Y(y), '▲', tc)
				}
			}
		}
	}
	s.midCity()
}

// midCity is the distant city: flat silhouettes whose height follows the
// real skyline, so a big economy has a big shadow behind it.
func (s *scene) midCity() {
	a := s.w.AgeIdx
	if a < 3 {
		return
	}
	prof := s.profileCache()
	par := 0.6
	base := materials[familyOf(a)][0].Wall.Lerp(hex(0x404858), 0.55)
	winc := s.p.Emit(cWinLitA, 3)
	if a >= 15 {
		winc = s.p.Emit(hex(0xff6ac8), 3)
	}
	for x := 0; x < s.v.W; x++ {
		mx := float64(x) + float64(s.v.Cam)*par
		// blocks of 3..7 columns, found by walking fixed 8-column cells
		cell := int(math.Floor(mx / 8))
		split := 3 + int(hash(cell, 7)%3)
		blk := cell * 2
		off := int(mx) - cell*8
		if off >= split {
			blk++
		}
		first := off == 0 || off == split
		last := off == split-1 || off == 7
		wx := int(mx / par)
		if wx < 0 || wx >= len(prof) {
			continue
		}
		// local density: the real skyline around this column
		dh := 0
		for k := -12; k <= 12; k++ {
			if i := wx + k*3; i >= 0 && i < len(prof) && prof[i] > dh {
				dh = prof[i]
			}
		}
		if dh == 0 {
			continue
		}
		col := s.p.Hill(base.Mul(0.9+0.2*hashf(blk, 5)), 0.45)
		hgt := int(float64(dh) * (0.4 + 0.5*hashf(blk, 61)))
		if a <= 5 { // walled town: a curtain wall with towers
			hgt = int(float64(s.groundY) * 0.12)
			if blk%5 == 0 {
				hgt += 2
			}
		}
		top := s.groundY - hgt
		for y := top; y < s.groundY; y++ {
			s.fb.set(x, s.Y(y), '█', col, col)
		}
		// the roofline says which century the far town is in
		switch {
		case a <= 5:
			if int(mx)%2 == 0 {
				s.fb.fg(x, s.Y(top-1), '▄', col)
			}
		case a <= 7:
			if first || last {
				s.fb.set(x, s.Y(top), '▄', col, s.fb.at(x, s.Y(top)-1).Bg)
				s.fb.at(x, s.Y(top)).Bg = s.fb.at(x, s.Y(top-1)).Bg
			} else {
				s.fb.fg(x, s.Y(top-1), '▄', col)
			}
		case a <= 11:
			if first && hash(blk, 9)%2 == 0 { // chimneys
				s.fb.fg(x, s.Y(top-1), '█', col)
				s.fb.fg(x, s.Y(top-2), '▄', col)
			}
		default:
			if off == 1 && hash(blk, 9)%3 == 0 { // masts
				s.fb.fg(x, s.Y(top-1), '│', col)
				s.fb.fg(x, s.Y(top-2), '│', col)
			}
			if a >= 15 && !first && !last && hash(blk, 11)%3 == 0 {
				s.fb.set(x, s.Y(top), '▀', s.p.Emit(hex(0x29f0ff), 3), col)
			}
		}
		if last && hgt > 2 { // a sliver of shade between blocks
			for y := top; y < s.groundY; y++ {
				s.fb.set(x, s.Y(y), '█', col.Mul(0.85), col.Mul(0.85))
			}
		}
		if s.p.Night > 0.4 && a > 5 {
			for y := top + 1; y < s.groundY; y += 2 {
				if hash(int(mx), y, 71)%4 == 0 && !last {
					s.fb.set(x, s.Y(y), '▪', winc, col)
				}
			}
		}
	}
}

func (s *scene) profileCache() []int {
	if s.prof == nil {
		s.prof = s.w.profile()
	}
	return s.prof
}
