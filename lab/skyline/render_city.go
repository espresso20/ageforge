package main

import (
	"math"
)

type emitter struct {
	x, y     int // screen col, scene row
	strength float64
	heavy    bool // industry: darker, more
	seed     int
}

// lots draws every building, back row first.
func (s *scene) lots() {
	p := s.p
	night := p.Night
	var em []emitter
	s.solid = make([]bool, s.v.W*s.v.H)
	for _, l := range s.w.Lots {
		x0 := l.X - s.v.Cam
		if x0+l.Spr.W < -8 || x0 > s.v.W+8 {
			// off screen, but its smoke might still drift in
			if x0+l.Spr.W < -20 || x0 > s.v.W+4 {
				continue
			}
		}
		y0 := s.groundY - l.Spr.H
		depth := l.Row
		sel := s.v.Cursor >= 0 && s.v.Cursor >= x0 && s.v.Cursor < x0+l.Spr.W && s.selected() == l
		neoWin := l.Age >= 15
		for sy := 0; sy < l.Spr.H; sy++ {
			for sx := 0; sx < l.Spr.W; sx++ {
				c := l.Spr.C[sy*l.Spr.W+sx]
				if c.Ch == 0 {
					continue
				}
				X, Y := x0+sx, s.Y(y0+sy)
				if !s.fb.ok(X, Y) {
					continue
				}
				s.solid[Y*s.v.W+X] = true
				fg := p.Col(l.Fam, l.Var, c.Fg, depth)
				if c.Fg == SWin {
					if night > 0.25 && s.windowLit(l, sx, sy) {
						lc := cWinLitA
						if hash(l.Seed, sx, sy)%4 == 0 {
							lc = cWinLitB
						}
						if neoWin && hash(l.Seed, sx, sy)%3 == 0 {
							lc = cWinLitNeo
						}
						fg = fg.Lerp(p.Emit(lc, depth), math.Min(1, (night-0.25)*2))
					}
				}
				if sel {
					fg = fg.Lerp(hex(0xffffff), 0.28)
				}
				if c.Bg == SNone {
					if c.Ch == '█' && textured(l, c.Fg, sx, sy) {
						// ANSI texture: thatch, tile, rubble and rock get a
						// sparse ▓ grain instead of a flat fill
						s.fb.set(X, Y, '▓', fg, fg.Mul(0.82))
					} else if c.Ch == '█' {
						s.fb.set(X, Y, '█', fg, fg)
					} else {
						s.fb.fg(X, Y, c.Ch, fg)
					}
					continue
				}
				bg := p.Col(l.Fam, l.Var, c.Bg, depth)
				if sel {
					bg = bg.Lerp(hex(0xffffff), 0.28)
				}
				s.fb.set(X, Y, c.Ch, fg, bg)
			}
		}
		for _, b := range l.Spr.Beacons {
			on := (s.v.Frame/3+int(hash(l.Seed, b.X)%7))%8 < 3
			if on || night < 0.3 {
				c := p.Emit(cBeacon, depth)
				if !on {
					c = c.Mul(0.55)
				}
				s.fb.fg(x0+b.X, s.Y(y0+b.Y), '•', c)
			}
		}
		for _, b := range l.Spr.Blades {
			s.blades(x0+b.X, y0+b.Y, l, depth)
		}
		for i, fl := range l.Spr.Flame {
			s.flame(x0+fl.X, y0+fl.Y, l.Seed+i, depth)
		}
		for i, e := range l.Spr.Smoke {
			if i > 0 && hash(l.Seed, i)%3 != 0 {
				continue // most buildings send up one plume, not one per flue
			}
			busy := l.Staff
			if !l.Producer {
				busy = 0.35
				if hash(l.Seed, 9)%2 == 0 {
					continue // not every hearth is lit
				}
			} else if busy < 0.05 {
				continue // an idle mill does not smoke: that is the point
			}
			em = append(em, emitter{x0 + e.X, y0 + e.Y, busy, l.Producer && l.Age >= 8, l.Seed + i*31})
		}
	}
	s.emitters = em
}

func textured(l *Lot, sl Slot, x, y int) bool {
	switch sl {
	case SRoof, SRoofShade, SRock, SRockDark, SLeaf, SLeafDark:
		return hash(l.Seed, x, y)%4 == 0
	case SWall, SWallShade:
		// rough stone and brick only, not glass and render
		if l.Fam <= 5 && (l.Fam != 2) {
			return hash(l.Seed, x, y)%7 == 0
		}
	}
	return false
}

func (s *scene) windowLit(l *Lot, x, y int) bool {
	on := hashf(l.Seed, x, y) < 0.15+0.8*l.Staff
	// a few windows change their minds every so often
	if hash(l.Seed, x, y, s.v.Frame/48)%29 == 0 {
		on = !on
	}
	return on
}

func (s *scene) blades(x, y int, l *Lot, depth int) {
	c := s.p.Col(l.Fam, l.Var, STrim, depth)
	ph := 0
	if l.Staff > 0.05 {
		ph = (s.v.Frame / 3) % 2
	}
	Y := func(dy int) int { return s.Y(y + dy) }
	if ph == 0 {
		s.fb.fg(x-1, Y(-1), '╲', c)
		s.fb.fg(x+1, Y(-1), '╱', c)
		s.fb.fg(x-1, Y(1), '╱', c)
		s.fb.fg(x+1, Y(1), '╲', c)
		s.fb.fg(x-2, Y(-2), '╲', c)
		s.fb.fg(x+2, Y(-2), '╱', c)
	} else {
		s.fb.fg(x, Y(-1), '│', c)
		s.fb.fg(x, Y(-2), '│', c)
		s.fb.fg(x-1, Y(0), '─', c)
		s.fb.fg(x-2, Y(0), '─', c)
		s.fb.fg(x+1, Y(0), '─', c)
		s.fb.fg(x+2, Y(0), '─', c)
		s.fb.fg(x, Y(1), '│', c)
	}
	s.fb.fg(x, Y(0), '•', c)
}

var flameGlyphs = []rune{'▲', '^', '♠', '▲'}

func (s *scene) flame(x, y, seed, depth int) {
	k := int(hash(seed, s.v.Frame/2) % 4)
	cols := []RGB{hex(0xffd23a), hex(0xff8a1a), hex(0xff5a1a), hex(0xffb03a)}
	s.fb.fg(x, s.Y(y), flameGlyphs[k], s.p.Emit(cols[k], depth))
}

// smoke rises from busy chimneys, drifting with the wind and thinning
// through the shade ramp ▓▒░· as it goes.
func (s *scene) smoke() {
	wind := 0.7 + 0.3*math.Sin(float64(s.w.St.Tick)/400)
	if s.p.D.Weather == 3 {
		wind = 1.1
	}
	for _, e := range s.emitters {
		n := 2 + int(e.strength*2)
		life := 11
		if e.heavy {
			n++
			life = 14
		}
		for k := 0; k < n; k++ {
			age := (s.v.Frame + k*life/n + int(hash(e.seed)%uint32(life))) % life
			// a plume leans downwind and spreads as it climbs
			fy := float64(e.y) - float64(age)*0.5
			fx := float64(e.x) + float64(age)*wind*1.3 + math.Sin(float64(age)*0.9+float64(e.seed%7))*0.35
			x, y := int(math.Round(fx)), int(math.Round(fy))
			u := float64(age) / float64(life)
			if u > 0.82 {
				continue // gone
			}
			base := hex(0xb8b4b0)
			if e.heavy {
				base = hex(0x6a6460)
			}
			col := s.p.Lit(base, 0)
			if e.heavy && s.p.Night > 0.4 && u < 0.3 {
				col = col.Lerp(s.p.Emit(hex(0xff7a30), 0), 0.35) // lit from the furnace below
			}
			// smoke is translucent: it tints what is behind it rather
			// than stamping a texture; only the fresh puff at the flue is
			// dense enough for a shade glyph
			spread := int(u * 3)
			alpha := (1 - u) * 0.75
			for dx := 0; dx <= spread; dx++ {
				xx := x + dx - spread/2
				px := s.fb.at(xx, s.Y(y))
				if px == nil || y < 0 || y >= s.groundY {
					continue
				}
				a := alpha
				if dx != spread/2 {
					a *= 0.6
				}
				px.Bg = px.Bg.Lerp(col, a)
				px.Fg = px.Fg.Lerp(col, a)
				if u < 0.2 && dx == spread/2 {
					s.fb.fg(xx, s.Y(y), '▓', col)
				} else if u < 0.45 && dx == spread/2 && px.Ch == ' ' {
					s.fb.fg(xx, s.Y(y), '▒', col)
				}
			}
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// ------------------------------------------------------------------ ground

func (s *scene) ground() {
	a := s.w.AgeIdx
	p := s.p
	gy := s.groundY
	for x := 0; x < s.v.W; x++ {
		wx := s.wx(x)
		di := s.w.DistrictAt(wx)
		dAge := -1
		if di >= 0 {
			dAge = s.w.Districts[di].Age
		}
		// water
		if di >= 0 {
			d := s.w.Districts[di]
			if d.Bay && wx >= d.BayX && wx < d.BayX+d.BW {
				s.water(x, wx, d)
				continue
			}
		}
		// surface: the district's own period
		var top, under RGB
		switch {
		case dAge < 0 || dAge <= 3:
			top, under = hex(0x5d8a3a), hex(0x6a4a2e)
		case dAge <= 7:
			top, under = hex(0x8a8478), hex(0x5a5448)
		case dAge <= 11:
			top, under = hex(0x6a5a50), hex(0x3e3834)
		default:
			top, under = hex(0x9aa0a8), hex(0x4a4e56)
		}
		s.fb.set(x, s.Y(gy), '▀', p.Lit(top, 0), p.Lit(under, 0))
		// the road: re-skinned to the current age, all the way through
		// the old town, so history stands on today's streets
		road, mark := s.roadStyle(a)
		rc := p.Lit(road, 0)
		ch, mc := ' ', rc
		switch {
		case a <= 1:
			if hash(wx, 3)%3 == 0 {
				ch, mc = '░', p.Lit(hex(0x7a9a4a), 0)
			}
		case a <= 7:
			if hash(wx, 4)%4 == 0 {
				ch, mc = '·', p.Lit(mark, 0)
			}
		case a <= 11:
			ch, mc = '═', p.Lit(mark, 0)
		case a <= 16:
			if (wx/3)%2 == 0 {
				ch, mc = '─', p.Lit(mark, 0)
			}
		default:
			ch, mc = '═', p.Emit(mark, 0)
		}
		s.fb.set(x, s.Y(gy+1), ch, mc, rc)
		// foreground verge
		vg := p.Lit(hex(0x3f6a2e), 0)
		vb := p.Lit(hex(0x2a4a22), 0)
		if a >= 12 && dAge >= 0 {
			vg, vb = p.Lit(hex(0x8a8e94), 0), p.Lit(hex(0x5a5e64), 0)
		}
		vch := ' '
		if h := hash(wx, 8) % 7; h < 2 {
			vch = []rune{'"', ','}[h]
		}
		s.fb.set(x, s.Y(gy+2), vch, vg, vb)
	}
	s.lamps()
	s.frontier()
}

func (s *scene) roadStyle(a int) (RGB, RGB) {
	switch {
	case a <= 1:
		return hex(0x4f7a32), hex(0x7a9a4a)
	case a <= 7:
		return hex(0x8a7050), hex(0xb09a70)
	case a <= 11:
		return hex(0x4a4440), hex(0x8a8a8a)
	case a <= 14:
		return hex(0x3a3c40), hex(0xe0d890)
	case a <= 16:
		return hex(0x24222e), hex(0xff3ea5)
	default:
		return hex(0x2a2a40), hex(0x7ff0ff)
	}
}

func (s *scene) water(x, wx int, d District) {
	p := s.p
	for y := s.groundY; y < s.S; y++ {
		depth := float64(y-s.groundY) / 3
		c := hex(0x2d5f8a).Lerp(hex(0x16304a), depth)
		c = c.Lerp(p.SkyAt(s.groundY-1-(y-s.groundY)*3, s.groundY), 0.35)
		col := p.Lit(c, 0)
		wave := hash(wx, y, s.v.Frame/4) % 9
		ch := ' '
		fg := col.Lerp(hex(0xffffff), 0.25)
		if wave == 0 {
			ch = '~'
		} else if wave == 1 {
			ch = '▀'
			fg = col.Mul(1.15)
		}
		// reflections of lit windows and the sky above the water line
		if above := s.fb.at(x, s.Y(s.groundY-1-(y-s.groundY))); above != nil && above.Fg.Luma() > col.Luma()+60 && p.Night > 0.4 {
			if hash(wx, y, s.v.Frame/3)%3 != 0 {
				ch, fg = '▒', above.Fg.Lerp(col, 0.45)
			}
		}
		s.fb.set(x, s.Y(y), ch, fg, col)
	}
	// quay edges
	if wx == d.BayX || wx == d.BayX+d.BW-1 {
		s.fb.set(x, s.Y(s.groundY), '█', p.Lit(hex(0x6a6258), 0), p.Lit(hex(0x6a6258), 0))
	}
}

func (s *scene) lamps() {
	if s.w.AgeIdx < 9 {
		return
	}
	for x := 0; x < s.v.W; x++ {
		wx := s.wx(x)
		if wx%16 != 0 || s.w.DistrictAt(wx) < 0 {
			continue
		}
		post := s.p.Lit(hex(0x3a3a40), 0)
		s.fb.fg(x, s.Y(s.groundY-1), '│', post)
		head := post
		if s.p.Night > 0.3 {
			head = s.p.Emit(hex(0xffe0a0), 0)
			if cell := s.fb.at(x, s.Y(s.groundY+1)); cell != nil {
				cell.Bg = cell.Bg.Lerp(head, 0.25)
			}
		}
		s.fb.fg(x, s.Y(s.groundY-2), '▀', head)
	}
}

// frontier is where the city is still being built: one crane per queued
// construction, scaffolding and all.
func (s *scene) frontier() {
	fx := s.w.FrontierX - s.v.Cam + 4
	q := s.w.Queue
	if len(q) == 0 {
		return
	}
	for i := range q {
		if i >= 4 {
			break
		}
		x := fx + i*7
		h := 6 + int(hash(i, 5)%4)
		c := s.p.Lit(hex(0xe0b020), 0)
		sc := s.p.Lit(hex(0x8a8070), 0)
		for y := s.groundY - h; y < s.groundY; y++ {
			s.fb.fg(x, s.Y(y), '╫', c)
		}
		for dx := -2; dx <= 4; dx++ {
			s.fb.fg(x+dx, s.Y(s.groundY-h), '═', c)
		}
		hook := s.groundY - h + 1 + (s.v.Frame/6+i)%3
		s.fb.fg(x+3, s.Y(hook), '┴', c)
		for y := s.groundY - 3; y < s.groundY; y++ {
			for dx := 1; dx <= 4; dx++ {
				s.fb.fg(x+dx, s.Y(y), '┼', sc)
			}
		}
	}
}
