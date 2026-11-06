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
	Home  int // the arm the people live on
	Rift  int // the angle at the rim where the rift opens
}

// Galaxy is the shape of a star map's galaxy and the arm the people live on;
// the sectors are laid along it.
type Galaxy struct {
	Shape GalaxyShape
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
		Rift:  r(turn, 13),
	}
	if r(3, 6) == 0 {
		s.Bar = 22 + r(20, 6+100)
	}
	s.Home = r(s.Arms, 10)
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

// NewGalaxy picks the shape of a seed's galaxy and walks its home arm.
func NewGalaxy(seed int64) *Galaxy {
	s := pickShape(seed)
	g := &Galaxy{Shape: s}
	for i := 0; i <= galArmN; i++ {
		c, _ := s.armCell(seed, s.Home, i)
		g.Home = append(g.Home, c)
	}
	return g
}

// sq is v squared.
func sq(v int) int { return v * v }
