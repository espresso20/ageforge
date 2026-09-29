package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// render_life.go: everything that moves or reacts: traffic on the road,
// ships in the bay, aircraft, weather, the harbinger on the ridge, the
// catastrophe overlay, and the "new since you were last here" markers.

func (s *scene) selected() *Lot {
	if s.v.Cursor < 0 {
		return nil
	}
	if s.sel == nil {
		s.sel = s.w.LotAt(s.wx(s.v.Cursor), -1, s.groundY)
	}
	return s.sel
}

// ----------------------------------------------------------------- traffic

type vehicle struct {
	art   string
	slots []Slot
	mat   int // material family
	lane  int // 0 road, 1 sky, 2 high sky
	speed float64
	seed  int
	trade bool
	war   bool
	smoke int // index in art of a funnel, -1 none
}

func vt(art string, slots ...Slot) vehicle {
	return vehicle{art: art, slots: slots, smoke: -1}
}

func roadVehicle(a, i int) vehicle {
	switch {
	case a <= 1:
		return vt("☻", SWallDark)
	case a <= 7:
		v := vt("▄█▄ ▄▀", SRoof, SWall, SRoof, SNone, STrunk, STrunk)
		return v
	case a <= 11:
		if i%3 == 0 {
			v := vt("▐▀▀▌▐▀▀▌▐█▄▄▀", SRoof, SRoof, SRoof, SRoof, SRoof, SRoof, SRoof, SRoof, SMetalDark, SMetalDark, SMetalDark, SMetalDark, SBeacon)
			v.smoke = 10
			return v
		}
		return vt("▄█▄", SNeon3, SNeon3, SNeon3)
	case a <= 14:
		if i%4 == 0 {
			return vt("▄▄▄█▄", SMetal, SMetal, SMetal, SNeon1, SNeon1)
		}
		return vt("▄█▄", []Slot{SNeon1, SNeon2, SNeon3, SMetal}[i%4], SGlassHi, []Slot{SNeon1, SNeon2, SNeon3, SMetal}[i%4])
	case a <= 16:
		return vt("▄▀▄", SNeon1, SGlassHi, SNeon2)
	default:
		return vt("▄▀▀▀▀▀▀▀▄", SGlow, STrim, STrim, STrim, STrim, STrim, STrim, STrim, SGlow)
	}
}

func (s *scene) traffic() {
	st := s.w.St
	a := s.w.AgeIdx
	var vs []vehicle
	n := 2 + min(6, st.Workers.TotalPop/180)
	for i := 0; i < n; i++ {
		v := roadVehicle(a, i)
		v.speed = 0.25 + hashf(i, 13)*0.35
		if hash(i, 17)%2 == 0 {
			v.speed = -v.speed
		}
		v.seed = i
		v.mat = familyOf(a)
		vs = append(vs, v)
	}
	// every active trade route is a caravan in the partner's gold
	for i := range st.Trade.ActiveRoutes {
		v := roadVehicle(a, i+7)
		v.trade = true
		v.speed = 0.3 + 0.1*float64(i%3)
		v.seed = 100 + i
		v.mat = familyOf(a)
		vs = append(vs, v)
	}
	// a war is a warband marching in from the west
	for _, k := range sortedFactionKeys(st) {
		if st.Diplomacy.Factions[k].AtWar {
			v := vt("►☻☻☻", SBeacon, SInk, SInk, SInk)
			v.war = true
			v.speed = 0.22
			v.seed = int(hash(len(k), int(k[0])))
			vs = append(vs, v)
		}
	}
	span := s.w.W + 20
	for _, v := range vs {
		pos := float64(hash(v.seed, 23)%uint32(span)) + float64(s.v.Frame)*v.speed
		wx := int(math.Mod(pos, float64(span)))
		if wx < 0 {
			wx += span
		}
		wx -= 10
		x := wx - s.v.Cam
		art := []rune(v.art)
		slots := v.slots
		if v.speed < 0 {
			art = reverseArt(art)
			slots = reverseSlots(slots)
		}
		y := s.Y(s.groundY + 1)
		for i, r := range art {
			if r == ' ' {
				continue
			}
			var c RGB
			switch {
			case v.trade && slots[i] == SRoof:
				c = s.p.Lit(hex(0xe0b020), 0)
			case slots[i] == SBeacon && v.war:
				c = s.p.Emit(hex(0xff2a2a), 0)
			default:
				c = s.p.Col(v.mat, v.seed, slots[i], 0)
			}
			s.fb.fg(x+i, y, r, c)
		}
		if v.smoke >= 0 {
			si := v.smoke
			if v.speed < 0 {
				si = len(art) - 1 - si
			}
			for k := 1; k <= 3; k++ {
				ch := []rune{'▓', '▒', '░'}[k-1]
				dx := -k
				if v.speed < 0 {
					dx = k
				}
				s.fb.fg(x+si+dx, s.Y(s.groundY-k), ch, s.p.Lit(hex(0xa0a0a0), 0))
			}
		}
		// headlights at night
		if s.p.Night > 0.4 && a >= 12 && !v.war {
			hx := x + len(art)
			if v.speed < 0 {
				hx = x - 1
			}
			if c := s.fb.at(hx, y); c != nil {
				c.Bg = c.Bg.Lerp(s.p.Emit(hex(0xfff0c0), 0), 0.3)
			}
		}
	}
	s.ships()
	s.aircraft()
	s.birds()
}

func reverseArt(r []rune) []rune {
	o := make([]rune, len(r))
	sw := map[rune]rune{'►': '◄', '▌': '▐', '▐': '▌'}
	for i, c := range r {
		if m, ok := sw[c]; ok {
			c = m
		}
		o[len(r)-1-i] = c
	}
	return o
}

func reverseSlots(s []Slot) []Slot {
	o := make([]Slot, len(s))
	for i, c := range s {
		o[len(s)-1-i] = c
	}
	return o
}

func sortedFactionKeys(st game.GameState) []string {
	var ks []string
	for k, f := range st.Diplomacy.Factions {
		if f.Discovered {
			ks = append(ks, k)
		}
	}
	sort.Strings(ks)
	return ks
}

func (s *scene) ships() {
	a := s.w.AgeIdx
	n := 1 + len(s.w.St.Trade.ActiveRoutes)
	for di, d := range s.w.Districts {
		if !d.Bay {
			continue
		}
		for i := 0; i < n; i++ {
			seed := di*10 + i
			sp := 0.08 + hashf(seed, 3)*0.1
			span := float64(d.BW - 8)
			ph := math.Mod(float64(hash(seed, 5)%100)+float64(s.v.Frame)*sp, span*2)
			off := ph
			dir := 1
			if ph > span {
				off, dir = span*2-ph, -1
			}
			x := d.BayX + 2 + int(off) - s.v.Cam
			y := s.groundY
			bob := 0
			if (s.v.Frame/5+i)%4 == 0 {
				bob = 0
			}
			_ = bob
			hull := s.p.Lit(hex(0x4a3020), 0)
			sail := s.p.Lit(hex(0xf0ead8), 0)
			switch {
			case a <= 7:
				s.fb.fg(x+1, s.Y(y-2), '▲', sail)
				s.fb.fg(x+1, s.Y(y-1), '█', sail)
				s.fb.fg(x+2, s.Y(y-1), '▌', sail)
				s.fb.fg(x+2, s.Y(y-2), '▄', sail)
				s.fb.fg(x, s.Y(y), '▀', hull)
				s.fb.fg(x+1, s.Y(y), '█', hull)
				s.fb.fg(x+2, s.Y(y), '█', hull)
				s.fb.fg(x+3, s.Y(y), '▀', hull)
			case a <= 11:
				hull = s.p.Lit(hex(0x2a2a2e), 0)
				s.fb.text(x, s.Y(y), "", hull, hull)
				for k, r := range []rune("▀████▀") {
					s.fb.fg(x+k, s.Y(y), r, hull)
				}
				fx := x + 2
				if dir < 0 {
					fx = x + 3
				}
				s.fb.fg(fx, s.Y(y-1), '█', s.p.Lit(hex(0xa03020), 0))
				s.fb.fg(fx-dir, s.Y(y-2), '▒', s.p.Lit(hex(0x9a9a9a), 0))
				s.fb.fg(fx-2*dir, s.Y(y-3), '░', s.p.Lit(hex(0x9a9a9a), 0))
			default:
				hull = s.p.Lit(hex(0x3a4a5a), 0)
				for k, r := range []rune("▀██████▀") {
					s.fb.fg(x+k, s.Y(y), r, hull)
				}
				cols := []RGB{hex(0xc04a2a), hex(0x2a7ac0), hex(0xe0b020), hex(0x3a9a4a)}
				for k := 1; k < 6; k++ {
					s.fb.fg(x+k, s.Y(y-1), '▄', s.p.Lit(cols[int(hash(seed, k))%4], 0))
				}
			}
			if s.p.Night > 0.4 {
				s.fb.fg(x+3, s.Y(y-1), '·', s.p.Emit(hex(0x60ff60), 0))
			}
		}
	}
}

func (s *scene) aircraft() {
	a := s.w.AgeIdx
	if a < 10 {
		return
	}
	n := 1 + min(3, (a-10)/2)
	for i := 0; i < n; i++ {
		sp := 0.5 + hashf(i, 31)*0.6
		span := s.v.W + 30
		x := int(float64(hash(i, 33)%uint32(span))+float64(s.v.Frame)*sp)%span - 15
		y := 2 + int(hash(i, 37)%uint32(max(1, s.groundY/3)))
		c := s.p.Lit(hex(0xd8dce4), 1)
		art := "─═►"
		if a >= 17 {
			art = "·─▄█►"
			c = s.p.Emit(hex(0xe8f8ff), 1)
		} else if a >= 15 && i%2 == 0 {
			// hover traffic threads between the towers
			art = "▀▄▀"
			y = s.groundY/2 + int(hash(i, 39)%uint32(max(1, s.groundY/4)))
			c = s.p.Emit([]RGB{hex(0xff3ea5), hex(0x29f0ff)}[i%2], 1)
		}
		for k, r := range []rune(art) {
			s.fb.fg(x+k, s.Y(y), r, c)
		}
		if s.p.Night > 0.4 && (s.v.Frame/3+i)%4 == 0 {
			s.fb.fg(x+len([]rune(art)), s.Y(y), '•', s.p.Emit(cBeacon, 1))
		}
	}
	// satellites crossing the night sky
	if a >= 12 && s.p.Night > 0.5 {
		for i := 0; i < 2; i++ {
			x := int(float64(hash(i, 91)%uint32(s.v.W))+float64(s.v.Frame)*0.2) % s.v.W
			s.fb.fg(x, s.Y(1+i*2), '·', s.p.Emit(hex(0xffffff), 2))
		}
	}
}

func (s *scene) birds() {
	if s.w.AgeIdx > 9 || s.p.Day < 0.5 || s.p.D.Weather >= 2 {
		return
	}
	for i := 0; i < 3; i++ {
		span := s.v.W + 20
		x := int(float64(hash(i, 71)%uint32(span))+float64(s.v.Frame)*0.35) % span
		y := 3 + int(hash(i, 73)%5)
		c := s.p.Lit(hex(0x2a2a30), 1)
		for k := 0; k < 3; k++ {
			ch := 'v'
			if (s.v.Frame/2+k+i)%2 == 0 {
				ch = '-'
			}
			s.fb.fg(x+k*2, s.Y(y+(k%2)), ch, c)
		}
	}
}

// ------------------------------------------------------------ weather

func weatherFor(w *World) int {
	for _, e := range w.St.ActiveEvents {
		k := e.Key
		switch {
		case strings.Contains(k, "storm"):
			return 3
		case strings.Contains(k, "flood"), strings.Contains(k, "plague"), strings.Contains(k, "blight"):
			return 2
		case strings.Contains(k, "drought"):
			return 0
		}
	}
	day := w.St.Tick / dayTicks
	switch hash(day, w.Seed) % 10 {
	case 7, 8:
		return 1
	case 9:
		return 2
	}
	return 0
}

func (s *scene) weather() {
	wt := s.p.D.Weather
	if wt < 2 {
		return
	}
	drops := s.v.W * s.S / 22
	if wt == 3 {
		drops = s.v.W * s.S / 10
	}
	c := s.p.Lit(hex(0xa8b8d0), 0).Lerp(s.p.Sky[2], 0.3)
	for i := 0; i < drops; i++ {
		y := int((hash(i, 1)%uint32(s.S+10) + uint32(s.v.Frame*2)) % uint32(s.S+2))
		x := int(hash(i, 2)%uint32(s.v.W+s.S)) - y/2
		if y >= s.S || x < 0 || x >= s.v.W {
			continue
		}
		s.fb.fg(x, s.Y(y), '╱', c)
	}
	// lightning in a storm, now and then
	if wt == 3 && (s.v.Frame/4)%23 == 0 {
		x := int(hash(s.v.Frame/92, 7) % uint32(s.v.W))
		bolt := s.p.Emit(hex(0xffffff), 0)
		for y := 0; y < s.groundY-4; y++ {
			ch := '╲'
			if hash(y, x)%2 == 0 {
				ch = '╱'
				x++
			} else {
				x--
			}
			s.fb.fg(x, s.Y(y), ch, bolt)
		}
	}
}

// ------------------------------------------------------------ harbinger

func pressure(w *World) float64 {
	if w.St.PendingCatastrophe != "" {
		return 1
	}
	if h := w.St.Harbinger; h != nil {
		return math.Min(1, 0.35+h.Probability*2)
	}
	return 0
}

// harbinger: a lone figure standing on the far ridge, facing the city.
// The closer the catastrophe, the nearer the hill it stands on.
func (s *scene) harbinger() {
	h := s.w.St.Harbinger
	if h == nil || s.farTops == nil {
		return
	}
	x := s.v.W/6 + int(hash(len(h.Key), int(h.Key[0]))%uint32(max(1, s.v.W*2/3)))
	top := int(s.farTops[min(max(0, x), len(s.farTops)-1)])
	ink := s.p.final(hex(0x0c0a10))
	if s.p.light {
		ink = s.p.final(hex(0x1a1420))
	}
	if s.w.AgeIdx >= 13 {
		// from the Digital Era the harbinger is a signal, not a man: a
		// hologram that flickers on the ridge
		if hash(s.v.Frame/2, 3)%5 == 0 {
			return
		}
		ink = s.p.Emit(hex(0x7ff0ff), 1)
	}
	y := top - 1
	s.fb.fg(x, s.Y(y-2), '▄', ink)
	s.fb.fg(x-1, s.Y(y-1), '▐', ink)
	s.fb.set(x, s.Y(y-1), '█', ink, ink)
	s.fb.fg(x+1, s.Y(y-1), '▌', ink)
	s.fb.fg(x-1, s.Y(y), '▐', ink)
	s.fb.fg(x+1, s.Y(y), '▌', ink)
	s.fb.set(x, s.Y(y), '█', ink, ink)
	s.fb.fg(x+2, s.Y(y-2), '│', ink)
	s.fb.fg(x+2, s.Y(y-1), '│', ink)
	s.fb.fg(x+2, s.Y(y), '│', ink)
	// his eyes catch the light at night
	if s.p.Night > 0.4 {
		s.fb.fg(x, s.Y(y-2), '▄', s.p.Emit(hex(0xff4a3a), 2))
	}
	// the name goes on the grid above him, drawn after the city so it is
	// only shown where no building stands in front of it
	s.labels = append(s.labels, pendingLabel{x - len([]rune(h.Name))/2, y - 4, h.Name, s.role(roleWarn)})
}

// ------------------------------------------------------------ neighbours

// relation colours are fixed signals, the same in every theme: war red,
// ally green, trade gold, neutral steel.
func relationColour(f game.FactionInfo) RGB {
	switch {
	case f.AtWar:
		return hex(0xff3a30)
	case f.Status == "rival" || f.Status == "embargo":
		return hex(0xe07a3a)
	case f.Status == "allied":
		return hex(0x4ad06a)
	case f.Status == "friendly" || f.TradeCount > 0:
		return hex(0xf0c040)
	}
	return hex(0x8aa0c0)
}

// factions puts every civilisation you have met on the far ridge as a small
// town of its own, pennant in the colour of your relations: the world view
// and the city view as one landscape.
func (s *scene) factions(tops []float64) {
	keys := sortedFactionKeys(s.w.St)
	n := len(keys)
	if n == 0 {
		return
	}
	for i, k := range keys {
		f := s.w.St.Diplomacy.Factions[k]
		// spaced along the ridge, drifting with its parallax
		x0 := (i+1)*s.v.W/(n+1) - int(float64(s.v.Cam)*0.15)%max(1, s.v.W/(n+1)) + int(hash(len(k), i)%9) - 4
		x0 = ((x0 % s.v.W) + s.v.W) % s.v.W
		w := 5 + f.Strength
		col := s.p.Hill(hex(0x3a4668), 0.3)
		for dx := 0; dx < w; dx++ {
			x := x0 + dx
			if x < 0 || x >= len(tops) {
				continue
			}
			top := int(tops[x])
			hgt := 1 + int(hash(i, dx, 3)%uint32(2+f.Strength/2))
			for y := top - hgt; y < top; y++ {
				if y >= 0 {
					s.fb.set(x, s.Y(y), '█', col, col)
				}
			}
			if s.p.Night > 0.4 && hash(i, dx)%2 == 0 && top-1 >= 0 {
				s.fb.set(x, s.Y(top-1), '·', s.p.Emit(cWinLitA, 4), col)
			}
		}
		px := x0 + w/2
		if px >= 0 && px < len(tops) {
			py := int(tops[px]) - 4
			rc := s.p.final(relationColour(f))
			s.fb.fg(px, s.Y(py+1), '│', col)
			s.fb.fg(px, s.Y(py+2), '│', col)
			s.fb.fg(px+1, s.Y(py+1), '►', rc)
			s.labels = append(s.labels, pendingLabel{px - len([]rune(f.Name))/2, py, f.Name, rc})
		}
	}
}

// ------------------------------------------------------------ catastrophe

func (s *scene) catastrophe() {
	ep := s.w.St.PendingCatastrophe
	if ep == "" {
		return
	}
	switch config.EpochByKey()[ep].CatastropheKey {
	case "digital_collapse", "solar_event":
		s.glitch()
		s.fires(8)
	case "reality_fracture":
		s.skyCracks()
		s.glitch()
	case "nuclear_meltdown":
		s.skyCracks()
		s.fires(5)
	default:
		s.fires(3)
	}
}

// fires: flames on the roofs of one lot in every n, with black smoke.
func (s *scene) fires(every int) {
	for _, l := range s.w.Lots {
		if int(hash(l.Seed, 77)%uint32(every)) != 0 || l.Wonder {
			continue
		}
		x0 := l.X - s.v.Cam
		y0 := s.groundY - l.Spr.H
		for sx := 0; sx < l.Spr.W; sx++ {
			// the first solid cell in this column is the roof
			for sy := 0; sy < l.Spr.H; sy++ {
				if l.Spr.at(sx, sy).Ch != 0 {
					if hash(l.Seed, sx)%2 == 0 {
						s.flame(x0+sx, y0+sy-1, l.Seed+sx*7, 0)
					}
					break
				}
			}
		}
		for k := 0; k < 7; k++ {
			age := (s.v.Frame + k*3) % 20
			x := x0 + l.Spr.W/2 + age/3
			y := y0 - 2 - age/2
			ch := []rune{'█', '▓', '▒', '░'}[min(3, age/5)]
			if y >= 0 {
				s.fb.fg(x, s.Y(y), ch, s.p.final(hex(0x1e1a1a)))
			}
		}
	}
}

func (s *scene) glitch() {
	for i := 0; i < s.S/2; i++ {
		y := int(hash(i, s.v.Frame/2) % uint32(s.groundY))
		x := int(hash(i, 5, s.v.Frame/2) % uint32(s.v.W))
		n := 4 + int(hash(i, 9)%16)
		c := s.p.Emit([]RGB{hex(0xff3ea5), hex(0x29f0ff), hex(0xffffff)}[i%3], 0)
		for k := 0; k < n; k++ {
			if px := s.fb.at(x+k, s.Y(y)); px != nil {
				s.fb.fg(x+k, s.Y(y), []rune("▓▒░▀▄")[int(hash(k, i))%5], c)
			}
		}
	}
}

func (s *scene) skyCracks() {
	c := s.p.Emit(hex(0xf0e8ff), 0)
	for i := 0; i < 3; i++ {
		x := s.v.W/4 + i*s.v.W/4
		for y := 0; y < s.groundY*2/3; y++ {
			if hash(i, y)%2 == 0 {
				x++
				s.fb.fg(x, s.Y(y), '╲', c)
			} else {
				x--
				s.fb.fg(x, s.Y(y), '╱', c)
			}
		}
	}
}

// ------------------------------------------------------------ sky megastructures

func (s *scene) skyStructures() {
	a := s.w.AgeIdx
	if a >= 18 { // a ringed planet hangs over the late city
		cx, cy := s.v.W*3/4-int(float64(s.v.Cam)*0.02), 3
		pc := s.p.Emit(hex(0xc8a0e0).Lerp(s.p.Sky[0], 0.3), 4)
		rc := s.p.Emit(hex(0xf0d8a0).Lerp(s.p.Sky[0], 0.3), 4)
		for dy := -1; dy <= 1; dy++ {
			for dx := -3; dx <= 3; dx++ {
				if math.Hypot(float64(dx)/2, float64(dy)) <= 1.6 {
					s.fb.set(cx+dx, s.Y(cy+dy), '█', pc, pc)
				}
			}
		}
		s.fb.fg(cx-2, s.Y(cy-2), '▄', pc)
		s.fb.fg(cx-1, s.Y(cy-2), '▄', pc)
		s.fb.fg(cx, s.Y(cy-2), '▄', pc)
		for dx := -6; dx <= 6; dx++ {
			if dx >= -3 && dx <= 3 {
				s.fb.set(cx+dx, s.Y(cy), '▀', rc, pc)
			} else {
				s.fb.fg(cx+dx, s.Y(cy), '─', rc)
			}
		}
	}
	if a >= 17 { // the orbital ring: a shallow arc across the whole sky
		rc := s.p.Emit(hex(0xb8c8e0), 4).Lerp(s.p.Sky[1], 0.35)
		cx := float64(s.v.W)/2 - float64(s.v.Cam)*0.03
		for x := 0; x < s.v.W; x++ {
			d := (float64(x) - cx) / float64(s.v.W)
			fy := 1.5 + d*d*float64(s.groundY)*0.9
			y := int(fy)
			if y >= s.groundY/2 {
				continue
			}
			ch := '▀'
			if fy-float64(y) >= 0.5 {
				ch = '▄'
			}
			s.fb.fg(x, s.Y(y), ch, rc)
			if s.p.Night > 0.4 && (x+s.v.Frame/4)%9 == 0 {
				s.fb.fg(x, s.Y(y), '•', s.p.Emit(hex(0xfff0a0), 3))
			}
		}
	}
	if a >= 17 { // space elevators rise from the spaceport district
		for _, di := range []int{17} {
			if di >= len(s.w.Districts) {
				continue
			}
			d := s.w.Districts[di]
			x := d.X0 + d.W/2 + 3 - s.v.Cam
			if x < 0 || x >= s.v.W {
				continue
			}
			tc := s.p.Col(9, 0, SMetal, 0)
			for y := 0; y < s.groundY-1; y++ {
				s.fb.fg(x, s.Y(y), '│', tc)
			}
			cy := (s.v.Frame / 2) % (s.groundY * 2)
			if cy >= s.groundY {
				cy = s.groundY*2 - cy
			}
			if cy < s.groundY-1 {
				s.fb.fg(x, s.Y(cy), '◘', s.p.Emit(hex(0x7ff0ff), 0))
			}
		}
	}
}

// ------------------------------------------------------------ markers

// markers flag what is new since the last check-in: a ▼ over every new
// silhouette, in the theme's positive colour.
func (s *scene) markers() {
	for _, lb := range s.labels {
		if lb.y < 0 {
			continue
		}
		free := true
		for i := range []rune(lb.text) {
			if x := lb.x + i; x >= 0 && x < s.v.W && s.solid[s.Y(lb.y)*s.v.W+x] {
				free = false
			}
		}
		if free {
			for i, r := range []rune(lb.text) {
				if c := s.fb.at(lb.x+i, s.Y(lb.y)); c != nil {
					s.fb.fg(lb.x+i, s.Y(lb.y), r, lb.col.Lerp(c.visible(), 0.2))
				}
			}
		}
	}
	if s.v.Cursor >= 0 {
		acc := s.role(roleAccent)
		s.fb.fg(s.v.Cursor, s.Y(s.groundY+1), '▲', acc)
		if l := s.selected(); l != nil {
			// bracket the selection on the road, and name it over its roof
			x0, x1 := l.X-s.v.Cam, l.X-s.v.Cam+l.Spr.W-1
			s.fb.fg(x0, s.Y(s.groundY+2), '└', acc)
			s.fb.fg(x1, s.Y(s.groundY+2), '┘', acc)
			for x := x0 + 1; x < x1; x++ {
				s.fb.fg(x, s.Y(s.groundY+2), '─', acc)
			}
			label := fmt.Sprintf(" %s ×%d ", l.Name, l.Count)
			ly := max(0, s.groundY-l.Spr.H-2)
			lx := max(0, min(s.v.W-len([]rune(label)), (x0+x1)/2-len([]rune(label))/2))
			s.fb.text(lx, s.Y(ly), label, s.role(roleBg), acc)
		}
	}
	if !s.v.ShowNew || s.w.Prev == nil {
		return
	}
	c := s.role(rolePositive)
	for _, l := range s.w.Lots {
		if !l.New {
			continue
		}
		x := l.X - s.v.Cam + l.Spr.W/2
		y := s.groundY - l.Spr.H - 1
		if y < 0 {
			y = 0
		}
		if (s.v.Frame/4)%2 == 0 || true {
			s.fb.fg(x, s.Y(y), '▼', c)
		}
	}
}
