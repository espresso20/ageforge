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
	// Home is the home arm, core to rim, as cells.
	Home []Pt
}

func pickShape(seed int64) GalaxyShape {
	r := func(n int, salt int64) int { return roll(n, seed, 300, salt) }
	s := GalaxyShape{
		Arms:  []int{2, 2, 2, 3, 3, 3, 3, 4, 4, 4}[r(10, 1)],
		Faint: r(3, 2),
		Wind:  1500 + r(2600, 3),
		Spin:  []int{1, -1}[r(2, 4)],
		Rot:   r(turn, 5),
		Core:  7 + r(8, 7),
		Flat:  72 + r(29, 8),
		Dust:  120 + r(300, 9),
		Warm:  r(241, 11) - 120,
		Glow:  85 + r(31, 12),
		Rift:  r(turn, 13),
	}
	if r(5, 6) < 2 {
		s.Bar = 18 + r(18, 6+100)
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

// reach is two opposite corners of the box round the disc.
func (g *Galaxy) reach() [2]Pt {
	ry := g.Shape.ryMax()
	return [2]Pt{{worldW/2 - galRX, worldH/2 - ry}, {worldW/2 + galRX - 1, worldH/2 + ry - 1}}
}

func (s GalaxyShape) ryMax() int { return galRYBase * s.Flat / 100 }

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
	x := worldW/2 + r*cos4096(a)/1024
	y := worldH/2 + r*sin4096(a)*s.ryMax()/(galRX*1024)
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

// NewGalaxy draws the sky of a seed.
func NewGalaxy(seed int64) *Galaxy {
	s := pickShape(seed)
	n := worldW * worldH
	g := &Galaxy{Shape: s, Dens: make([]uint8, n), Sky: make([]Cell, n)}
	arm, dust, core := make([]uint8, n), make([]uint8, n), make([]uint8, n)
	for k := 0; k < s.Arms+s.Faint; k++ {
		amp := 255
		if k >= s.Arms {
			amp = 105
		}
		for i := 0; i <= galArmN; i++ {
			rf := i * 1000 / galArmN
			c, r := s.armAt(seed, k, rf, 0)
			edge := min(1000, (1000-rf)*8)
			splat(arm, c, 3+r/14, amp*(1000-rf/3)/1000*edge/1000)
			if k < s.Arms && rf > 250 && rf < 800 {
				d, rr := s.armAt(seed, k, rf, s.Spin*s.Dust)
				splat(dust, d, 1+rr/50, 170)
			}
		}
	}
	for i := 0; i <= galArmN; i++ {
		c, _ := s.armCell(seed, s.Home, i)
		g.Home = append(g.Home, c)
	}
	// core and bar, in the galaxy's own circle (rows stretched to columns)
	ca, sa := cos4096(s.Rot), sin4096(s.Rot)
	cx, cy, ry := worldW/2, worldH/2, s.ryMax()
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			dx, w := x-cx, (y-cy)*galRX/ry
			v := bump(dx*dx+w*w, s.Core*s.Core*4) * 235 >> 8
			if s.Bar > 0 {
				u, t := (dx*ca+w*sa)/1024, (-dx*sa+w*ca)/1024
				if b := bump(u*u*1000/(s.Bar*s.Bar)+t*t*1000/36, 1000) * 170 >> 8; b > v {
					v = b
				}
			}
			core[at(x, y)] = uint8(min(v, 255))
		}
	}
	g.paint(seed, arm, dust, core)
	return g
}

// paint turns the fields into the finished sky: a dithered gradient of deep
// tones, the glow of the arms (warm near the core, cool toward the rim), dust
// lanes that darken them, the core, three layers of stars and the far
// galaxies.
func (g *Galaxy) paint(seed int64, arm, dust, core []uint8) {
	s := g.Shape
	cx, cy, ry := worldW/2, worldH/2, s.ryMax()
	h8 := func(salt int64, x, y int) int { return int(hash(seed, salt, int64(x), int64(y)) % 256) }
	var near []Pt
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			i := at(x, y)
			dx, w := x-cx, (y-cy)*galRX/ry
			rad := isqrt((dx*1000/galRX)*(dx*1000/galRX) + (w*1000/galRX)*(w*1000/galRX))
			a := int(arm[i]) * (140 + fbm(seed+21, x*256/6, y*256/4, 3)*140/noiseMax) >> 8
			disc := bump(rad*rad, 1000*1000) * 22 >> 8
			glow := min(255, a*s.Glow*5/400+disc)
			dens := min(255, glow+int(core[i]))
			g.Dens[i] = uint8(dens)
			// the deep gradient, a few weights apart, dithered
			tone := 5 + x*8/worldW + y*6/worldH + fbm(seed+31, x*256/60, y*256/28, 3)*9/noiseMax +
				max(0, fbm(seed+33, x*256/45, y*256/20, 3)-560)*7/(noiseMax-560) + h8(41, x, y)/90
			bg := P(HNebula, tone)
			if glow > 10 {
				bg = P(HNebula, tone+glow*130/255)
			}
			if glow > 35 {
				warm := clamp((640-rad+s.Warm)*256/120+128, 0, 256)
				if h8(42, x, y) < warm {
					bg = P(HSun, 30+glow*95/255)
				}
			}
			if d := int(dust[i]) * min(255, a+40) / 255; d > 100 && h8(43, x, y) < d*2/3 {
				bg = P(HVoid, 255)
			}
			if c := int(core[i]); c*220/255 > int(bg.Weight) && (bg.Hue == HSun || bg.Hue == HNebula) {
				bg = P(HSun, c*220/255)
			}
			g.Sky[i] = Cell{R: ' ', Fg: P(HStar, 0), Bg: bg}
			// three layers of stars, thicker in the arms
			h := int(hash(seed, 44, int64(x), int64(y)) % 1000)
			mult := 256 + dens*2
			pn, pm, pf := 4*mult/256, 16*(256+dens*3)/256, 60*(256+dens*3)/256
			ink := HStar
			if dens > 60 && h8(45, x, y) < 100 {
				ink = HSun
			}
			switch {
			case h < pn:
				g.Sky[i].R, g.Sky[i].Fg = pick("✦✧", uint64(h8(46, x, y))), P(ink, 255)
				near = append(near, Pt{x, y})
			case h < pn+pm:
				g.Sky[i].R, g.Sky[i].Fg = pick("∙∙+", uint64(h8(47, x, y))), P(ink, 150+h8(48, x, y)*60/256)
			case h < pn+pm+pf:
				g.Sky[i].R, g.Sky[i].Fg = '·', P(HStar, 70+h8(49, x, y)*60/256)
			}
		}
	}
	for _, p := range near { // a near star lights the cells beside it
		for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			if x, y := p.X+d[0], p.Y+d[1]; inWorld(x, y) && g.Sky[at(x, y)].Bg.Hue != HSun {
				g.Sky[at(x, y)].Bg = P(HNebula, 60)
			}
		}
	}
	for n, c := range s.Far {
		g.smudge(seed, n, c)
	}
}

// smudge is a far galaxy: a tilted oval of faint light and dots.
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
			if e < 30 {
				continue
			}
			cell := &g.Sky[at(x, y)]
			cell.Bg = P(HCool, e*40>>8)
			switch {
			case e > 200:
				cell.R, cell.Fg = '∙', P(HStar, 200)
			case e > 90:
				cell.R, cell.Fg = '·', P(HStar, 150)
			}
		}
	}
}

// homeNodes are n places along the home arm from near the core out, evenly
// spaced and nudged a little off the line.
func (g *Galaxy) homeNodes(seed int64, n int) []Pt {
	line := g.Home[galArmN/8 : galArmN*9/10]
	total := 0
	length := []int{0}
	for i := 1; i < len(line); i++ {
		total += isqrt(sq(line[i].X-line[i-1].X) + 4*sq(line[i].Y-line[i-1].Y))
		length = append(length, total)
	}
	var out []Pt
	taken := map[Pt]bool{{worldW / 2, worldH / 2}: true}
	j := 0
	for k := 0; k < n; k++ {
		want := total * (2*k + 1) / (2 * n)
		for j < len(line)-1 && length[j] < want {
			j++
		}
		p := line[j]
		for try := 0; ; try++ {
			q := Pt{p.X + roll(5, seed, 330, int64(k), int64(try)) - 2, p.Y + roll(3, seed, 331, int64(k), int64(try)) - 1}
			if inWorld(q.X, q.Y) && !taken[q] {
				taken[q] = true
				out = append(out, q)
				break
			}
		}
	}
	return out
}

func sq(v int) int { return v * v }
