package shields

import "math"

// Domain is the kind of work a badge is for; each has its own emblem.
type Domain int

const (
	Food Domain = iota
	Lumber
	Masonry
	Faith
	Trade
	Engineering
	Energy
	Hacker
	Astronaut
	Military
	Metallurgy
	NumDomains
)

var domainNames = [NumDomains]string{"Food", "Lumber", "Masonry", "Faith", "Trade", "Engineering", "Energy", "Hacker", "Astronaut", "Military", "Metallurgy"}

func (d Domain) String() string { return domainNames[d] }

// The large emblem is drawn on a 20 by 22 mask in tones; each has its own
// shape and its own way of being built.
const (
	emW = 20
	emH = 22
)

var bigEmblems = [NumDomains]func(m *mask){
	Food: func(m *mask) {
		// three ears, each a stalk with grains stepping down it in pairs
		ear := func(bx, by, tx, ty int) {
			m.line(bx, by, tx, ty, tMain)
			m.set(tx, ty-1, tMain)
			for j := 0; j < 3; j++ {
				y := ty + 2 + 3*j
				x := tx + (bx-tx)*(y-ty)/(by-ty)
				m.set(x-1, y, tMain)
				m.set(x-2, y-1, tLight)
				m.set(x+1, y, tMain)
				m.set(x+2, y-1, tLight)
			}
		}
		ear(10, 13, 10, 2)
		ear(9, 13, 4, 4)
		ear(11, 13, 16, 4)
		m.poly(tMain, 8, 15, 12, 15, 17, 21.5, 3, 21.5)
		m.line(10, 16, 10, 21, tShade)
		m.line(8, 16, 6, 21, tShade)
		m.line(12, 16, 14, 21, tShade)
		m.rect(7, 12, 6, 3, tShade)
		m.line(7, 12, 12, 12, tLight)
		m.set(12, 13, tLight)
	},
	Lumber: func(m *mask) {
		// a long haft, a broad head that bites to the right, a chip of wood
		m.poly(tNoBevl, 2, 19.5, 4, 21.5, 14.5, 7, 12.5, 5.5)
		m.poly(tMain, 8, 5, 10.5, 1.5, 15.5, 0, 19, 2.5, 19.5, 8, 16.5, 11, 13.5, 11.5, 12, 8.5)
		m.line(16, 2, 18, 8, tLight)
		m.line(15, 2, 17, 9, tLight)
		m.rect(10, 5, 3, 3, tDeep)
		m.line(3, 19, 12, 7, tShade)
		m.line(6, 17, 7, 19, tDeep)
		m.line(8, 14, 9, 16, tDeep)
		m.poly(tMain, 1, 4, 4, 3, 5, 6, 2, 7)
		m.line(2, 5, 4, 5, tLight)
	},
	Masonry: func(m *mask) {
		const cx, cy = 10.0, 11.0
		seg := func(x, y int) int {
			return int(math.Floor(math.Atan2(cy-(float64(y)+.5), float64(x)+.5-cx) / (math.Pi / 7)))
		}
		ring := func(x, y int) bool {
			dx, dy := float64(x)+.5-cx, float64(y)+.5-cy
			r := math.Hypot(dx, dy)
			return dy <= 0.5 && r >= 5.2 && r <= 9.8
		}
		for y := 0; y < 12; y++ {
			for x := 0; x < emW; x++ {
				if !ring(x, y) {
					continue
				}
				m.set(x, y, tMain)
				if ring(x-1, y) && seg(x-1, y) != seg(x, y) {
					m.set(x, y, tDeep)
				}
				if s := seg(x, y); s == 3 {
					m.set(x, y, tLight)
				}
			}
		}
		for _, x0 := range []int{0, 14} {
			m.rect(x0, 11, 6, 9, tMain)
			for r := 0; r < 4; r++ {
				m.line(x0, 12+2*r, x0+5, 12+2*r, tDeep)
				m.set(x0+1+(r%2)*3, 11+2*r, tDeep)
			}
		}
		m.rect(0, 20, 20, 2, tShade)
		m.line(0, 20, 19, 20, tLight)
	},
	Faith: func(m *mask) {
		m.poly(tMain, 10, 0, 12.5, 4, 15.5, 8, 16, 12.5, 14, 16, 10, 17.5, 6, 16, 4, 12.5, 4.5, 8.5, 7, 5.5, 8, 2.5)
		m.poly(tMain, 4.5, 8.5, 2.5, 4.5, 6, 7)
		m.poly(tMain, 15.5, 8.5, 17.5, 5, 14, 7)
		m.poly(tGem, 10, 6, 12.5, 10, 13, 13.5, 10, 16.5, 7, 13.5, 7.5, 10)
		m.poly(tGemHi, 10, 10, 11.5, 12.5, 10, 15, 8.5, 12.5)
		m.poly(tNoBevl, 1, 18, 19, 18, 15.5, 21.5, 4.5, 21.5)
		m.line(1, 18, 18, 18, tLight)
		m.line(5, 20, 14, 20, tShade)
	},
	Trade: func(m *mask) {
		m.disc(10, 2, 1.7, tLight)
		m.rect(9, 3, 2, 16, tMain)
		m.rect(0, 4, 20, 2, tMain)
		m.line(0, 4, 19, 4, tLight)
		m.rect(5, 19, 10, 3, tShade)
		m.rect(5, 19, 10, 1, tLight)
		for _, c := range []int{3, 16} {
			m.line(c, 6, c-3, 13, tShade)
			m.line(c, 6, c+3, 13, tShade)
			cf := float64(c)
			m.poly(tMain, cf-4.5, 13, cf+4.5, 13, cf+3, 16.5, cf-3, 16.5)
			m.line(c-4, 13, c+3, 13, tLight)
		}
	},
	Engineering: func(m *mask) {
		const cx, cy = 10.0, 11.0
		for y := 0; y < emH; y++ {
			for x := 0; x < emW; x++ {
				dx, dy := float64(x)+.5-cx, float64(y)+.5-cy
				r, a := math.Hypot(dx, dy), math.Atan2(dy, dx)
				ph := math.Mod(a+2*math.Pi+math.Pi/16, math.Pi/4)
				tooth := math.Abs(ph-math.Pi/8) < 0.23
				switch {
				case r < 2.7:
				case r <= 6.6, r <= 9.8 && tooth && r > 6:
					m.set(x, y, tMain)
					if r >= 4.2 && r < 5 {
						m.set(x, y, tDeep)
					}
				}
			}
		}
		m.disc(cx, cy, 2.7, 0)
		m.poly(tLight, 8.5, 10, 9.5, 9, 11.5, 9, 12, 10)
	},
	Energy: func(m *mask) {
		m.poly(tMain, 12, 0, 3, 12, 8.8, 12, 6, 21.5, 17, 8, 11, 8, 15, 0)
		m.poly(tGem, 12, 3, 7.5, 10.5, 11, 10.5, 9.5, 17, 14, 9.5, 10, 9.5, 12.5, 3)
		for _, p := range [][2]int{{18, 13}, {2, 4}, {17, 2}, {3, 17}} {
			m.set(p[0], p[1], tLight)
			m.set(p[0]-1, p[1], tLight)
			m.set(p[0]+1, p[1], tLight)
			m.set(p[0], p[1]-1, tLight)
			m.set(p[0], p[1]+1, tLight)
		}
	},
	Hacker: func(m *mask) {
		m.rect(0, 2, 20, 17, tMain)
		m.rect(1, 5, 18, 13, tDeep)
		for _, x := range []int{2, 4, 6} {
			m.set(x, 3, tDeep)
		}
		m.rect(15, 3, 3, 1, tShade)
		m.line(4, 8, 8, 11, tGem)
		m.line(8, 11, 4, 14, tGem)
		m.line(5, 8, 9, 11, tGem)
		m.line(9, 11, 5, 14, tGem)
		m.rect(11, 14, 5, 2, tLight)
		m.rect(10, 8, 7, 1, tShade)
		m.rect(11, 10, 4, 1, tShade)
	},
	Astronaut: func(m *mask) {
		m.poly(tMain, 10, 0, 13, 4.5, 13.6, 13.5, 6.4, 13.5, 7, 4.5)
		m.poly(tMain, 6.5, 9.5, 2.5, 14.5, 2.5, 18.5, 6.5, 16)
		m.poly(tMain, 13.5, 9.5, 17.5, 14.5, 17.5, 18.5, 13.5, 16)
		m.disc(10, 7.4, 2.6, tLight)
		m.disc(10, 7.4, 1.6, tDeep)
		m.rect(8, 13, 4, 2, tShade)
		m.poly(tGem, 8, 15, 12, 15, 10, 21.5)
		m.poly(tGemHi, 9, 15, 11, 15, 10, 19)
		for _, p := range [][2]int{{1, 4}, {18, 2}, {17, 20}, {1, 11}} {
			m.set(p[0], p[1], tLight)
		}
	},
	Military: func(m *mask) {
		blade := func(f func(x float64) float64) {
			m.poly(tMain, f(0.2), 0.2, f(2.6), 0.2, f(15.5), 13.4, f(13.4), 15.5, f(0.2), 2.6)
			m.line(int(f(2)), 2, int(f(13)), 13, tLight)
			m.line(int(f(1)), 3, int(f(12.4)), 14, tShade)
		}
		blade(func(x float64) float64 { return x })
		blade(func(x float64) float64 { return 19.8 - x })
		m.line(10, 17, 17, 10, tLight)
		m.line(10, 18, 17, 11, tShade)
		m.line(9, 17, 2, 10, tLight)
		m.line(9, 18, 2, 11, tShade)
		m.line(15, 16, 17, 18, tShade)
		m.line(16, 15, 18, 17, tShade)
		m.line(4, 16, 2, 18, tShade)
		m.line(3, 15, 1, 17, tShade)
		m.disc(18.3, 19, 1.3, tLight)
		m.disc(1.5, 19, 1.3, tLight)
	},
	Metallurgy: func(m *mask) {
		m.rect(7, 2, 6, 2, tGem)
		m.rect(8, 2, 3, 1, tGemHi)
		m.rect(4, 4, 15, 4, tMain)
		m.poly(tMain, 0, 5.2, 4, 4, 4, 8, 1.5, 7)
		m.poly(tMain, 7, 8, 16, 8, 15, 13, 8, 13)
		m.poly(tMain, 5, 13, 18, 13, 19.5, 17, 3.5, 17)
		m.rect(3, 17, 17, 3, tShade)
		m.line(3, 17, 19, 17, tLight)
		m.line(4, 4, 18, 4, tLight)
		for _, p := range [][2]int{{2, 1}, {16, 0}, {17, 2}, {5, 3}} {
			m.set(p[0], p[1], tLight)
		}
	},
}
