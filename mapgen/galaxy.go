package mapgen

// The galaxy of the three space ages: a seeded spiral of arms, a core and
// maybe a bar, drawn as a deep sky once per seed (the plan is cached). Integers
// only: angles are 4096ths of a turn and sin comes from a table.

// sinTable is sin over a quarter-turn-by-16 grid: entry i is sin(i/256 of a
// turn) scaled by 1024; the last entry closes the circle.
var sinTable = [257]int{
	0, 25, 50, 75, 100, 125, 150, 175, 200, 224, 249, 273, 297, 321, 345, 369,
	392, 415, 438, 460, 483, 505, 526, 548, 569, 590, 610, 630, 650, 669, 688, 706,
	724, 742, 759, 775, 792, 807, 822, 837, 851, 865, 878, 891, 903, 915, 926, 936,
	946, 955, 964, 972, 980, 987, 993, 999, 1004, 1009, 1013, 1016, 1019, 1021, 1023, 1024,
	1024, 1024, 1023, 1021, 1019, 1016, 1013, 1009, 1004, 999, 993, 987, 980, 972, 964, 955,
	946, 936, 926, 915, 903, 891, 878, 865, 851, 837, 822, 807, 792, 775, 759, 742,
	724, 706, 688, 669, 650, 630, 610, 590, 569, 548, 526, 505, 483, 460, 438, 415,
	392, 369, 345, 321, 297, 273, 249, 224, 200, 175, 150, 125, 100, 75, 50, 25,
	0, -25, -50, -75, -100, -125, -150, -175, -200, -224, -249, -273, -297, -321, -345, -369,
	-392, -415, -438, -460, -483, -505, -526, -548, -569, -590, -610, -630, -650, -669, -688, -706,
	-724, -742, -759, -775, -792, -807, -822, -837, -851, -865, -878, -891, -903, -915, -926, -936,
	-946, -955, -964, -972, -980, -987, -993, -999, -1004, -1009, -1013, -1016, -1019, -1021, -1023, -1024,
	-1024, -1024, -1023, -1021, -1019, -1016, -1013, -1009, -1004, -999, -993, -987, -980, -972, -964, -955,
	-946, -936, -926, -915, -903, -891, -878, -865, -851, -837, -822, -807, -792, -775, -759, -742,
	-724, -706, -688, -669, -650, -630, -610, -590, -569, -548, -526, -505, -483, -460, -438, -415,
	-392, -369, -345, -321, -297, -273, -249, -224, -200, -175, -150, -125, -100, -75, -50, -25,
	0,
}

// turn is a full circle in the angle unit.
const turn = 4096

// sin4096 is sin(a) scaled by 1024, a in 4096ths of a turn.
func sin4096(a int) int {
	a &= turn - 1
	i, f := a>>4, a&15
	return sinTable[i] + (sinTable[i+1]-sinTable[i])*f/16
}

func cos4096(a int) int { return sin4096(a + turn/4) }

const (
	galRX     = 110 // the galaxy's reach either way, in columns
	galRYBase = 44  // and in rows, before it is flattened
	galArmN   = 260 // steps along an arm
)

// GalaxyShape is everything that makes one seed's galaxy differ from
// another's. It is comparable, so a test can count the distinct ones.
type GalaxyShape struct {
	Arms  int // major arms, 2 to 4
	Faint int // faint arms between them, 0 to 2
	Wind  int // how far an arm turns from core to rim, in 4096ths of a turn
	Spin  int // 1 or -1: which way the arms wind
	Rot   int // where the first arm and the bar point
	Tilt  int // how far the whole disc is turned on the sky, in 4096ths of a turn
	Bar   int // the bar's half length in columns, 0 for none
	Core  int // radius of the core glow in columns
	Flat  int // height of the disc in percent of the full 44 rows
	Dust  int // how far the dust lane trails its arm, in 4096ths of a turn
	Home  int // the arm the people live on
	Warm  int // moves the radius where warm light turns cool
	Glow  int // strength of the arms in percent
	Rift  int // the angle at the rim where the rift opens
	Far   [3]Pt
}

// Galaxy is the sky of a star map: the shape, the density it makes, and the
// finished background of every cell.
type Galaxy struct {
	Shape GalaxyShape
	// Dens is how dense the stars are at each cell (arms, core), 0 to 255.
	Dens []uint8
	// Sky is the background and the stars of every cell of the world.
	Sky []Cell
	// Deep is the same sky for a close look: the deep gradient and a few dim
	// stars, with no arms, no dust, no core and no far galaxies, so that what
	// stands on it is the only thing with any weight.
	Deep []Cell
	// Home is the home arm, core to rim, as cells.
	Home []Pt
}

func pickShape(seed int64) GalaxyShape {
	r := func(n int, salt int64) int { return roll(n, seed, 300, salt) }
	s := GalaxyShape{
		Arms:  []int{2, 2, 2, 2, 2, 3, 3, 3, 4, 4}[r(10, 1)],
		Tilt:  r(turn/6, 15) - turn/12,
		Faint: r(3, 2),
		Wind:  900 + r(4200, 3),
		Spin:  []int{1, -1}[r(2, 4)],
		Rot:   r(turn, 5),
		Core:  7 + r(8, 7),
		Flat:  40 + r(61, 8),
		Dust:  120 + r(300, 9),
		Warm:  r(241, 11) - 120,
		Glow:  85 + r(31, 12),
		Rift:  r(turn, 13),
	}
	if r(3, 6) == 0 {
		s.Bar = 22 + r(20, 6+100)
	}
	s.Home = r(s.Arms, 10)
	// the distant galaxies sit in three of the four corners
	corners := [4]Pt{{8, 3}, {worldW - 36, 3}, {8, worldH - 16}, {worldW - 36, worldH - 16}}
	skip := r(4, 14)
	n := 0
	for i, c := range corners {
		if i == skip {
			continue
		}
		s.Far[n] = Pt{c.X + 6 + r(16, int64(20+i)), c.Y + 4 + r(6, int64(30+i))}
		n++
	}
	return s
}

// reach is two opposite corners of the box round the disc, which is an oval
// turned by its tilt.
func (g *Galaxy) reach() [2]Pt {
	s := g.Shape
	cs, sn := cos4096(s.Tilt), sin4096(s.Tilt)
	b := galRX * s.Flat / 100
	hw := isqrt(sq(galRX*cs/16)+sq(b*sn/16)) / 64
	hh := isqrt(sq(galRX*sn/16)+sq(b*cs/16)) / 64 * galRYBase / galRX
	return [2]Pt{{worldW/2 - hw, worldH/2 - hh}, {worldW/2 + hw - 1, worldH/2 + hh - 1}}
}

// circle is a cell's place in the galaxy's own frame: the offset from the
// middle undone of the tilt and the flattening, rows made as long as columns,
// so that the disc is a circle of radius galRX.
func (s GalaxyShape) circle(dx, dy int) (int, int) {
	cs, sn := cos4096(s.Tilt), sin4096(s.Tilt)
	y := dy * galRX / galRYBase
	return (dx*cs + y*sn) / 1024, (-dx*sn + y*cs) / 1024 * 100 / s.Flat
}

// armPhase is where arm k begins, major arms first and then the faint ones
// between them.
func (s GalaxyShape) armPhase(k int) int {
	if k < s.Arms {
		return s.Rot + k*turn/s.Arms
	}
	return s.Rot + (2*(k-s.Arms)+1)*turn/(2*s.Arms)
}

// armAngle is the angle of arm k at rf (0 to 1000 along it): it winds fast
// near the core and slower out at the rim.
func (s GalaxyShape) armAngle(seed int64, k, rf int) int {
	wind := s.Wind + roll(301, seed, 301, int64(k)) - 150
	return s.armPhase(k) + s.Spin*wind*(rf*3/4+isqrt(rf*1000)/4)/1000
}

// armCell is the cell of arm k at step i of galArmN, and its radius in
// columns.
func (s GalaxyShape) armCell(seed int64, k, i int) (Pt, int) {
	return s.armAt(seed, k, i*1000/galArmN, 0)
}

// armAt is the cell of arm k at rf, with its angle moved by shift.
func (s GalaxyShape) armAt(seed int64, k, rf, shift int) (Pt, int) {
	rIn := max(s.Core*3/2, s.Bar)
	r := rIn + (galRX-rIn)*rf/1000
	a := s.armAngle(seed, k, rf) + shift
	gx, gy := r*cos4096(a)/1024, r*sin4096(a)/1024*s.Flat/100
	cs, sn := cos4096(s.Tilt), sin4096(s.Tilt)
	x := worldW/2 + (gx*cs-gy*sn)/1024
	y := worldH/2 + (gx*sn+gy*cs)/1024*galRYBase/galRX
	return Pt{x, y}, r
}

// bump is a smooth hill: 256 at d2 of nothing, none at rad2.
func bump(d2, rad2 int) int {
	if d2 >= rad2 {
		return 0
	}
	t := 256 - d2*256/rad2
	return t * t >> 8
}

// splat raises field to a hill at c, in the true shape of the screen (a row is
// two columns tall).
func splat(field []uint8, c Pt, rho, amp int) {
	for dy := -rho/2 - 1; dy <= rho/2+1; dy++ {
		for dx := -rho; dx <= rho; dx++ {
			x, y := c.X+dx, c.Y+dy
			if !inWorld(x, y) {
				continue
			}
			if v := bump(dx*dx+4*dy*dy, rho*rho) * amp >> 8; v > int(field[at(x, y)]) {
				field[at(x, y)] = uint8(min(v, 255))
			}
		}
	}
}

// armWidth is how wide arm k is at rf (0 to 1000 along it) and r columns out,
// and how bright: full strength through the first half, then narrowing and
// fading to a thin faint tip.
func armWidth(r, rf int, home bool) (rho, amp int) {
	taper := 1000 - max(0, rf-420)*1000/760
	base := 5 + r/8
	if home {
		base += 2
	}
	return max(3, base*(300+taper*7/10)/1000), 120 + 135*taper/1000
}

// soft lifts the foot of a hill so a glow reaches further before it fades.
func soft(b int) int { return b * (512 - b) / 256 }

// NewGalaxy draws the sky of a seed.
func NewGalaxy(seed int64) *Galaxy {
	s := pickShape(seed)
	n := worldW * worldH
	g := &Galaxy{Shape: s, Dens: make([]uint8, n), Sky: make([]Cell, n), Deep: make([]Cell, n)}
	arm, dust, core, knot := make([]uint8, n), make([]uint8, n), make([]uint8, n), make([]uint8, n)
	for k := 0; k < s.Arms+s.Faint; k++ {
		faint := k >= s.Arms
		for i := 0; i <= galArmN; i++ {
			rf := i * 1000 / galArmN
			c, r := s.armAt(seed, k, rf, 0)
			rho, amp := armWidth(r, rf, k == s.Home)
			if faint {
				rho, amp = rho*3/4, amp*2/5
			} else if k != s.Home {
				amp = amp * 9 / 10
			}
			splat(arm, c, rho, amp*min(1000, (1000-rf)*8)/1000)
			if !faint && rf > 200 && rf < 800 {
				d, rr := s.armAt(seed, k, rf, s.Spin*s.Dust)
				splat(dust, d, 2+rr/40, 150)
			}
			// bright knots: a few clusters of young stars along the arm
			if !faint && rf > 80 && rf < 900 && i%9 == 0 && roll(100, seed, 340, int64(k), int64(i)) < 24 {
				splat(knot, c, 3+roll(3, seed, 341, int64(k), int64(i)), 255)
			}
		}
	}
	for i := 0; i <= galArmN; i++ {
		c, _ := s.armCell(seed, s.Home, i)
		g.Home = append(g.Home, c)
	}
	// core and bar, in the galaxy's own circle (rows stretched to columns): a
	// halo, a bright body and a white-hot middle
	ca, sa := cos4096(s.Rot), sin4096(s.Rot)
	cx, cy := worldW/2, worldH/2
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			dx, w := s.circle(x-cx, y-cy)
			d2 := dx*dx + w*w
			v := soft(bump(d2, s.Core*s.Core*6))*80>>8 + soft(bump(d2, s.Core*s.Core*2))*90>>8
			if s.Bar > 0 {
				u, t := (dx*ca+w*sa)/1024, (-dx*sa+w*ca)/1024
				if b := bump(u*u*1000/(s.Bar*s.Bar)+t*t*1000/49, 1000) * 110 >> 8; b > v {
					v = b
				}
			}
			core[at(x, y)] = uint8(min(v, 255))
		}
	}
	g.paint(seed, arm, dust, core, knot)
	return g
}

// paint turns the fields into the finished sky: a smooth dithered gradient of
// deep tones, the glow of the arms (warm near the core, cool toward the rim)
// graded from each arm's middle out into the dark, dust lanes that dim them,
// the core, three layers of stars that crowd the arms, and the far galaxies.
func (g *Galaxy) paint(seed int64, arm, dust, core, knot []uint8) {
	s := g.Shape
	cx, cy := worldW/2, worldH/2
	h8 := func(salt int64, x, y int) int { return int(hash(seed, salt, int64(x), int64(y)) % 256) }
	var near []Pt
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			i := at(x, y)
			dx, w := s.circle(x-cx, y-cy)
			rad := isqrt((dx*1000/galRX)*(dx*1000/galRX) + (w*1000/galRX)*(w*1000/galRX))
			// the arm's light, clumped a little by a slow noise, dimmed by dust
			clump := 150 + fbm(seed+21, x*256/9, y*256/5, 3)*106/noiseMax
			a := int(arm[i]) * clump >> 8
			if d := int(dust[i]); d > 0 {
				a = a * (256 - d*3/2) / 256
				a = max(a, 0)
			}
			kn := int(knot[i]) * (200 + h8(50, x, y)/5) >> 8
			a = min(255, a+kn/3)
			disc := bump(rad*rad, 1000*1000) * 20 >> 8
			glow := min(255, a*s.Glow/100+disc)
			dens := min(255, glow+int(core[i]))
			g.Dens[i] = uint8(dens)
			// the deep gradient: smooth, one weight of grain
			tone := 6 + x*7/worldW + y*5/worldH + fbm(seed+31, x*256/80, y*256/40, 2)*3/noiseMax + h8(41, x, y)/140
			bg := P(HNebula, tone)
			if glow > 4 {
				bg = P(HNebula, tone+glow*175/255*(215+h8(42, x, y)/6)/256)
			}
			// the core is one warm-white glow that wins over the arm where it
			// is the stronger, its weight falling smoothly with the distance
			if c := int(core[i]); c > 6 {
				if w := (c*55 + (h8(52, x, y)-128)*3) / 255; w > int(bg.Weight) {
					bg = P(HCore, w)
				}
			}
			g.Sky[i] = Cell{R: ' ', Fg: P(HStar, 0), Bg: bg}
			// the close sky: the gradient alone and a thin even scatter of dim
			// stars, a third of the faint ones and none of the bright, with
			// no crowding in the arms
			g.Deep[i] = Cell{R: ' ', Fg: P(HStar, 0), Bg: P(HNebula, tone)}
			switch dh := int(hash(seed, 63, int64(x), int64(y)) % 1000); {
			case dh < 14:
				g.Deep[i].R, g.Deep[i].Fg = '∙', P(HStar, 60+h8(61, x, y)*40/256)
			case dh < 40:
				g.Deep[i].R, g.Deep[i].Fg = '·', P(HStar, 50+h8(62, x, y)*40/256)
			}
			// three layers of stars, thicker in the arms and the core
			h := int(hash(seed, 44, int64(x), int64(y)) % 1000)
			pn, pm, pf := 3*(256+dens*3+kn)/256, 14*(256+dens*5+kn*2)/256, 55*(256+dens*6)/256
			ink := HStar
			if dens > 70 && h8(45, x, y) < 110 {
				ink = HSun
			}
			switch {
			case h < pn:
				g.Sky[i].R, g.Sky[i].Fg = pick("✦✧", uint64(h8(46, x, y))), P(ink, 255)
				near = append(near, Pt{x, y})
			case h < pn+pm:
				g.Sky[i].R, g.Sky[i].Fg = pick("∙∙+", uint64(h8(47, x, y))), P(ink, 150+h8(48, x, y)*80/256)
			case h < pn+pm+pf:
				g.Sky[i].R, g.Sky[i].Fg = '·', P(HStar, 70+h8(49, x, y)*70/256)
			}
		}
	}
	for _, p := range near { // a near star lights the cells beside it
		for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			if x, y := p.X+d[0], p.Y+d[1]; inWorld(x, y) {
				c := &g.Sky[at(x, y)]
				c.Bg = P(c.Bg.Hue, min(255, int(c.Bg.Weight)+22))
			}
		}
	}
	for n, c := range s.Far {
		g.smudge(seed, n, c)
	}
}

// smudge is a far galaxy: a small tilted oval of soft light round a dot.
func (g *Galaxy) smudge(seed int64, n int, c Pt) {
	a, b := 6+roll(3, seed, 320, int64(n)), 2+roll(2, seed, 321, int64(n))
	ang := roll(turn/2, seed, 322, int64(n))
	ca, sa := cos4096(ang), sin4096(ang)
	for dy := -a/2 - 2; dy <= a/2+2; dy++ {
		for dx := -a - 2; dx <= a+2; dx++ {
			x, y := c.X+dx, c.Y+dy
			if !inWorld(x, y) {
				continue
			}
			u, t := (dx*ca+2*dy*sa)/1024, (-dx*sa+2*dy*ca)/1024
			e := bump(u*u*1000/(a*a)+t*t*1000/(b*b), 1000)
			if e < 12 {
				continue
			}
			cell := &g.Sky[at(x, y)]
			cell.Bg = P(HNebula, min(255, int(cell.Bg.Weight)+e*50>>8))
			switch {
			case dx == 0 && dy == 0:
				cell.R, cell.Fg = '◌', P(HStar, 220)
			case e > 110:
				cell.R, cell.Fg = '∙', P(HStar, 190)
			case e > 40 && cell.R == ' ':
				cell.R, cell.Fg = '·', P(HStar, 140)
			}
		}
	}
}

func sq(v int) int { return v * v }
