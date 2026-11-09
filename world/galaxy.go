package world

// The galaxy of the three space ages: a seeded spiral of arms, a core and
// maybe a bar. Integers only. Cells are the 240x96 raster the terminal map
// and the native map both lay the galaxy out on.

// Pt is a cell of the galaxy raster.
type Pt struct{ X, Y int }

const (
	CellsW = 240 // the raster the galaxy is laid on, in columns
	CellsH = 96  // and in rows
)

const (
	GalRX     = 110 // the galaxy's reach either way, in columns
	GalRYBase = 44  // and in rows, before it is flattened
	GalArmN   = 260 // steps along an arm
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

func PickShape(seed int64) GalaxyShape {
	r := func(n int, salt int64) int { return Roll(n, seed, 300, salt) }
	s := GalaxyShape{
		Arms:  []int{2, 2, 2, 2, 2, 3, 3, 3, 4, 4}[r(10, 1)],
		Tilt:  r(Turn/6, 15) - Turn/12,
		Faint: r(3, 2),
		Wind:  900 + r(4200, 3),
		Spin:  []int{1, -1}[r(2, 4)],
		Rot:   r(Turn, 5),
		Core:  7 + r(8, 7),
		Flat:  40 + r(61, 8),
		Rift:  r(Turn, 13),
	}
	if r(3, 6) == 0 {
		s.Bar = 22 + r(20, 6+100)
	}
	s.Home = r(s.Arms, 10)
	return s
}

// reach is two opposite corners of the box round the disc, which is an oval
// turned by its tilt.
func (g *Galaxy) Reach() [2]Pt {
	s := g.Shape
	cs, sn := Cos4096(s.Tilt), Sin4096(s.Tilt)
	b := GalRX * s.Flat / 100
	hw := Isqrt(Sq(GalRX*cs/16)+Sq(b*sn/16)) / 64
	hh := Isqrt(Sq(GalRX*sn/16)+Sq(b*cs/16)) / 64 * GalRYBase / GalRX
	return [2]Pt{{CellsW/2 - hw, CellsH/2 - hh}, {CellsW/2 + hw - 1, CellsH/2 + hh - 1}}
}

// ArmPhase is where arm k begins, major arms first and then the faint ones
// between them.
func (s GalaxyShape) ArmPhase(k int) int {
	if k < s.Arms {
		return s.Rot + k*Turn/s.Arms
	}
	return s.Rot + (2*(k-s.Arms)+1)*Turn/(2*s.Arms)
}

// ArmAngle is the angle of arm k at rf (0 to 1000 along it): it winds fast
// near the core and slower out at the rim.
func (s GalaxyShape) ArmAngle(seed int64, k, rf int) int {
	wind := s.Wind + Roll(301, seed, 301, int64(k)) - 150
	return s.ArmPhase(k) + s.Spin*wind*(rf*3/4+Isqrt(rf*1000)/4)/1000
}

// ArmCell is the cell of arm k at step i of GalArmN, and its radius in
// columns.
func (s GalaxyShape) ArmCell(seed int64, k, i int) (Pt, int) {
	return s.ArmAt(seed, k, i*1000/GalArmN, 0)
}

// ArmAt is the cell of arm k at rf, with its angle moved by shift.
func (s GalaxyShape) ArmAt(seed int64, k, rf, shift int) (Pt, int) {
	rIn := max(s.Core*3/2, s.Bar)
	r := rIn + (GalRX-rIn)*rf/1000
	a := s.ArmAngle(seed, k, rf) + shift
	gx, gy := r*Cos4096(a)/1024, r*Sin4096(a)/1024*s.Flat/100
	cs, sn := Cos4096(s.Tilt), Sin4096(s.Tilt)
	x := CellsW/2 + (gx*cs-gy*sn)/1024
	y := CellsH/2 + (gx*sn+gy*cs)/1024*GalRYBase/GalRX
	return Pt{x, y}, r
}

// NewGalaxy picks the shape of a seed's galaxy and walks its home arm.
func NewGalaxy(seed int64) *Galaxy {
	s := PickShape(seed)
	g := &Galaxy{Shape: s}
	for i := 0; i <= GalArmN; i++ {
		c, _ := s.ArmCell(seed, s.Home, i)
		g.Home = append(g.Home, c)
	}
	return g
}
