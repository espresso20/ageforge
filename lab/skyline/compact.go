package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// compact.go: the glanceable form for a sidebar (e.g. 40x15). The whole
// panorama is squeezed into the box as a skyline "sparkline": eighth-block
// tops (▁▂▃▄▅▆▇) give each column 8 steps of height, the facade below is
// the tallest building's own material, and the same signals as the full
// view ride on it: lit windows = staffed, smoke = producing, ▼ = new since
// the last check-in, the harbinger on the horizon, fire in a catastrophe.

var eighths = []rune(" ▁▂▃▄▅▆▇█")

func renderCompact(w *World, v View) *FB {
	fb := newFB(v.W, v.H)
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
	s := &scene{w: w, v: v, p: newPal(d), fb: fb, S: v.H - 1, groundY: v.H - 2, top: 1}
	p := s.p
	S := s.groundY // sky+city rows: 0..S-1 (scene coords)
	for y := 0; y < S; y++ {
		c := p.final(p.SkyAt(y, S))
		for x := 0; x < v.W; x++ {
			fb.fill(x, s.Y(y), c)
		}
	}
	if p.Stars > 0.2 {
		for y := 0; y < S/2; y++ {
			for x := 0; x < v.W; x++ {
				if hashf(x, y, 991) < 0.03*p.Stars {
					fb.fg(x, s.Y(y), '·', p.final(hex(0xfff6e0)))
				}
			}
		}
	}
	// sun or moon
	sx := int(float64(v.W) * (0.1 + 0.8*math.Mod(tod+0.27, 0.5)/0.5))
	if p.Day > 0.3 {
		fb.fg(sx, s.Y(1), '●', p.final(hex(0xfff0b0)))
	} else {
		fb.fg(sx, s.Y(1), '☾', p.final(hex(0xe8ecf8)))
	}
	// a far ridge for depth
	for x := 0; x < v.W; x++ {
		n := fbm(float64(x)/9, 300+w.Seed%1000, 3)
		top := float64(S) - 1 - n*float64(S)*0.35
		col := p.Hill(hex(0x4a5a7e), 0.45)
		for y := int(math.Ceil(top)); y < S; y++ {
			fb.fill(x, s.Y(y), col)
		}
		if top-math.Floor(top) > 0.01 {
			fb.set(x, s.Y(int(math.Floor(top))), eighths[int((math.Ceil(top)-top)*8)], col, fb.at(x, s.Y(int(math.Floor(top)))).Bg)
		}
	}
	// the skyline
	prof := s.profileCache()
	maxH := 1
	for _, h := range prof {
		maxH = max(maxH, h)
	}
	span := float64(w.W) / float64(v.W)
	top := float64(S - 1)
	for x := 0; x < v.W; x++ {
		a, b := int(float64(x)*span), int(float64(x+1)*span)
		h := 0
		for i := a; i < b && i < len(prof); i++ {
			h = max(h, prof[i])
		}
		if h == 0 {
			continue
		}
		l := w.LotAt((a+b)/2, -1, 0)
		for i := a; l == nil && i < b; i++ {
			l = w.LotAt(i, -1, 0)
		}
		fam, vr, staff := familyOf(w.AgeIdx), 0, 0.5
		if l != nil {
			fam, vr, staff = l.Fam, l.Var, l.Staff
		}
		wall := p.Col(fam, vr, SWall, 0)
		hh := float64(h) / float64(maxH) * top
		full := int(hh)
		for k := 0; k < full; k++ {
			y := S - 1 - k
			fb.fill(x, s.Y(y), wall)
			if p.Night > 0.35 && k < full-1 && hash(x, y, w.Seed)%100 < uint32(20+60*staff) {
				fb.set(x, s.Y(y), '·', p.Emit(cWinLitA, 0), wall)
			}
		}
		if f := hh - float64(full); f > 0.06 {
			y := S - 1 - full
			fb.set(x, s.Y(y), eighths[max(1, int(f*8))], wall, fb.at(x, s.Y(y)).Bg)
		}
		// signals above the roofline
		ay := S - 2 - full
		if ay >= 0 && l != nil {
			switch {
			case v.ShowNew && anyNew(w, a, b):
				fb.fg(x, s.Y(ay), '▼', s.role(rolePositive))
			case l.Wonder:
				fb.fg(x, s.Y(ay), '♦', p.Emit(hex(0xffd479), 0))
			case l.Producer && l.Staff > 0.05 && hash(x, 3)%3 == 0:
				fb.fg(x, s.Y(ay), []rune("░▒")[(v.Frame/4+x)%2], p.Lit(hex(0xb0aca8), 0))
			}
		}
		if w.St.PendingCatastrophe != "" && hash(x, 5)%5 == 0 && ay >= 0 {
			fb.fg(x, s.Y(ay+1), flameGlyphs[int(hash(x, v.Frame/2)%4)], p.Emit(cFire, 0))
		}
	}
	if h := w.St.Harbinger; h != nil {
		hx := v.W/6 + int(hash(len(h.Key))%uint32(max(1, v.W/2)))
		fb.fg(hx, s.Y(S/2), 'Ω', s.role(roleWarn))
	}
	// the road, with the caravans on it
	road, mark := s.roadStyle(w.AgeIdx)
	for x := 0; x < v.W; x++ {
		fb.set(x, s.Y(S), ' ', p.Lit(mark, 0), p.Lit(road, 0))
	}
	for i := 0; i < 1+len(w.St.Trade.ActiveRoutes); i++ {
		x := (int(hash(i, 23)%uint32(v.W)) + v.Frame/3) % v.W
		fb.fg(x, s.Y(S), '▪', p.Lit(hex(0xe0b020), 0))
	}
	// one line of header: age, clock, what changed
	bg := s.role(roleSurface)
	for x := 0; x < v.W; x++ {
		fb.set(x, 0, ' ', bg, bg)
	}
	name := strings.ToUpper(strings.TrimSuffix(w.St.AgeName, " Age"))
	x := fb.text(0, 0, " "+name+" ", s.role(roleBg), s.role(roleAccent))
	x = fb.text(x, 0, " "+clock(tod), s.role(roleText), bg)
	if v.ShowNew && w.NewCount > 0 {
		x = fb.text(x, 0, fmt.Sprintf(" ▼%d", w.NewCount), s.role(rolePositive), bg)
	}
	idle := idleProducers(w)
	if idle > 0 {
		fb.text(x, 0, fmt.Sprintf(" %d idle", idle), s.role(roleWarn), bg)
	}
	return fb
}

func anyNew(w *World, a, b int) bool {
	for _, l := range w.Lots {
		if l.New && l.X+l.Spr.W/2 >= a && l.X+l.Spr.W/2 < b {
			return true
		}
	}
	return false
}

// idleProducers counts producer types with slots and no one working them.
func idleProducers(w *World) int {
	n := 0
	for _, b := range w.St.Buildings {
		if b.Count > 0 && b.WorkerCapacity > 0 && b.WorkersAssigned == 0 {
			n++
		}
	}
	return n
}
