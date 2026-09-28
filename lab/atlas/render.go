package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// The one renderer for the planet plates. It walks the same layers for
// every epoch — ground, water, coast, relief, borders, lines, marks,
// labels — and the plate's dials decide what each layer looks like.

// Rect is a screen rectangle.
type Rect struct{ X, Y, W, H int }

// View is the camera: a centre in world units and a scale (world units per
// cell across; a cell is two units tall at the same scale).
type View struct{ CX, CY, S float64 }

// z-order of the layers
const (
	zGround = 1
	zGrid   = 2
	zWater  = 3
	zRelief = 4
	zRiver  = 5
	zCoast  = 6
	zBorder = 7
	zLine   = 8
	zMover  = 10
	zMark   = 12
	zLabel  = 14
	zChrome = 20
	zCursor = 30
)

type earthCtx struct {
	cv    *Canvas
	rc    Rect
	a     *Atlas
	p     *Plate
	pal   *Pal
	v     View
	frame int
	land  []bool
	known []float64
}

func (e *earthCtx) world(i, j float64) Pt {
	return Pt{e.v.CX + (i-float64(e.rc.W)/2)*e.v.S, e.v.CY + (j-float64(e.rc.H)/2)*2*e.v.S}
}

// screen maps a world point to fractional rect coordinates.
func (e *earthCtx) screen(p Pt) (float64, float64) {
	return (p.X-e.v.CX)/e.v.S + float64(e.rc.W)/2, (p.Y-e.v.CY)/(2*e.v.S) + float64(e.rc.H)/2
}

func (e *earthCtx) cell(p Pt) (int, int, bool) {
	fx, fy := e.screen(p)
	x, y := int(math.Floor(fx)), int(math.Floor(fy))
	return e.rc.X + x, e.rc.Y + y, x >= 0 && y >= 0 && x < e.rc.W && y < e.rc.H
}

func (e *earthCtx) isLand(x, y int) bool {
	i, j := x-e.rc.X, y-e.rc.Y
	if i < 0 || j < 0 || i >= e.rc.W || j >= e.rc.H {
		return false
	}
	return e.land[j*e.rc.W+i]
}

func (e *earthCtx) knownAt(x, y int) float64 {
	i, j := x-e.rc.X, y-e.rc.Y
	if i < 0 || j < 0 || i >= e.rc.W || j >= e.rc.H {
		return 0
	}
	return e.known[j*e.rc.W+i]
}

func drawEarth(cv *Canvas, rc Rect, a *Atlas, p *Plate, pal *Pal, v View, frame int) *earthCtx {
	e := &earthCtx{cv: cv, rc: rc, a: a, p: p, pal: pal, v: v, frame: frame,
		land: make([]bool, rc.W*rc.H), known: make([]float64, rc.W*rc.H)}
	e.ground()
	if p.Rhumbs {
		e.rhumbs()
	}
	e.rivers()
	e.lines()
	e.marks()
	e.labels()
	return e
}

// ground: water, coast, relief, borders and fog, cell by cell.
func (e *earthCtx) ground() {
	w, a, p, pal, cv := e.a.W, e.a, e.p, e.pal, e.cv
	W, H := e.rc.W, e.rc.H
	exact := e.v.S < 0.9
	hc := make([]float64, (W+1)*(H+1))
	cd := make([]float64, (W+1)*(H+1))
	for j := 0; j <= H; j++ {
		for i := 0; i <= W; i++ {
			pt := e.world(float64(i), float64(j))
			if exact {
				hc[j*(W+1)+i] = w.Height(pt.X, pt.Y)
			} else {
				hc[j*(W+1)+i] = w.H(pt.X, pt.Y)
			}
			if len(p.WaterLines) > 0 {
				cd[j*(W+1)+i] = w.CoastDist(pt.X, pt.Y)
			}
		}
	}
	corner := func(arr []float64, i, j int) (tl, tr, br, bl float64) {
		return arr[j*(W+1)+i], arr[j*(W+1)+i+1], arr[(j+1)*(W+1)+i+1], arr[(j+1)*(W+1)+i]
	}
	owners := make([]int, W*H)
	for j := 0; j < H; j++ {
		for i := 0; i < W; i++ {
			c := e.world(float64(i)+0.5, float64(j)+0.5)
			owners[j*W+i] = -2
			k := a.Known(c.X, c.Y)
			e.known[j*W+i] = k
			tl, tr, br, bl := corner(hc, i, j)
			h := (tl + tr + br + bl) / 4
			e.land[j*W+i] = h > 0
			if k >= 0.5 {
				owners[j*W+i] = a.OwnerAt(c.X, c.Y)
			}
		}
	}
	for j := 0; j < H; j++ {
		for i := 0; i < W; i++ {
			x, y := e.rc.X+i, e.rc.Y+j
			c := e.world(float64(i)+0.5, float64(j)+0.5)
			cx, cy := int(math.Floor(c.X/e.v.S)), int(math.Floor(c.Y/(2*e.v.S)))
			r := rnd01(w.Seed+7, cx, cy)
			cv.SetBg(x, y, pal.Paper)
			k := e.known[j*W+i]
			tl, tr, br, bl := corner(hc, i, j)
			h := (tl + tr + br + bl) / 4
			if k < 0.5 {
				e.fog(x, y, k, r, h > 0)
				continue
			}
			// graticule under everything
			if p.Graticule > 0 {
				e.graticule(x, y, i, j)
			}
			bits := 0
			if tl > 0 {
				bits |= 8
			}
			if tr > 0 {
				bits |= 4
			}
			if br > 0 {
				bits |= 2
			}
			if bl > 0 {
				bits |= 1
			}
			if h <= 0 && bits == 0 {
				e.sea(x, y, i, j, c, r, cd, corner)
				continue
			}
			if bits != 0 && bits != 15 && p.Coast != "none" {
				g := msRound[bits]
				if p.Coast == "square" {
					g = msSquare[bits]
				}
				col, attr := pal.Ink, tcell.AttrMask(0)
				if p.CoastBold {
					attr = tcell.AttrBold
				}
				switch {
				case p.Coast == "dim":
					col = pal.mix(pal.Paper, pal.Water, 0.55)
				case p.Relief == "contour":
					col = pal.Water
				}
				cv.PutA(x, y, g, col, attr, zCoast)
			}
			if h <= 0 {
				if p.Relief == "shade" || p.Relief == "contour" {
					cv.SetBg(x, y, pal.WaterWash)
				}
				continue
			}
			// land
			o := owners[j*W+i]
			if p.Wash > 0 && o >= 0 {
				if o == 0 {
					bg := pal.PlayerWash
					if rg := w.RegionAt(c.X, c.Y); rg >= 0 && a.FreshReg[rg] {
						bg = pal.FreshWash
					}
					cv.SetBg(x, y, bg)
				} else {
					cv.SetBg(x, y, pal.RelWash[relation(a.Civs[o-1])])
				}
			}
			e.relief(x, y, i, j, c, h, r, hc, corner)
			if p.Night && o >= 0 {
				// from orbit a realm is its lights: settled land glows
				r2 := rnd01(w.Seed+8, cx, cy)
				near := math.Inf(1)
				for _, t := range a.Towns {
					near = math.Min(near, t.At.Dist(c))
				}
				col := pal.Light2
				if o > 0 {
					col = pal.mix(pal.Paper, pal.Rel(relation(a.Civs[o-1])), 0.75)
				}
				switch {
				case near < 5*e.v.S && r2 < 0.8:
					cv.Put(x, y, '•', pal.Light1, zBorder+1)
				case near < 9*e.v.S && r2 < 0.5:
					cv.Put(x, y, '∙', pal.Light1, zBorder+1)
				case r2 < 0.22:
					cv.Put(x, y, '·', col, zBorder+1)
				}
			}
			// borders: where the owner changes to the right or below
			if o >= -1 && !p.Night {
				for _, d := range [][2]int{{1, 0}, {0, 1}} {
					ii, jj := i+d[0], j+d[1]
					if ii >= W || jj >= H || !e.land[jj*W+ii] {
						continue
					}
					o2 := owners[jj*W+ii]
					if o2 != o && o2 >= -1 && (o >= 0 || o2 >= 0) {
						who := o
						if who < 0 {
							who = o2
						}
						col := pal.Accent
						if who > 0 {
							col = pal.Rel(relation(a.Civs[who-1]))
						}
						g := p.BorderRune
						if g != '·' {
							g = '┊'
							if d[1] == 1 {
								g = '┈'
							}
						}
						cv.Put(x, y, g, col, zBorder)
					}
				}
			}
		}
	}
	// spot heights on the survey sheet: the highest point in each block
	if p.Relief == "contour" && !p.Mini {
		const bw, bh = 18, 7
		for by := 0; by < H; by += bh {
			for bx := 0; bx < W; bx += bw {
				best, bi, bj := 0.0, -1, -1
				for j := by + 1; j < by+bh-1 && j < H; j++ {
					for i := bx + 1; i < bx+bw-5 && i < W; i++ {
						if e.known[j*W+i] < 0.5 {
							continue
						}
						if v := hc[j*(W+1)+i]; v > best {
							best, bi, bj = v, i, j
						}
					}
				}
				if bi >= 0 && best > 0.25 {
					x, y := e.rc.X+bi, e.rc.Y+bj
					lab := fmt.Sprintf("%d", int(best*1850))
					cv.Put(x, y, '▴', pal.Relief, zMark-1)
					cv.Text(x+1, y, lab, pal.Relief, 0, zMark-1)
					cv.Reserve(x, y, len(lab)+1, 1)
				}
			}
		}
	}
}

func (e *earthCtx) fog(x, y int, k, r float64, land bool) {
	p, pal, cv := e.p, e.pal, e.cv
	switch p.Fog {
	case "dark":
		// a charcoal smudge where knowledge gives out
		if k > 0.12 && r < 0.55 {
			g := '·'
			if k > 0.3 && r < 0.25 {
				g = '░'
			}
			cv.Put(x, y, g, pal.mix(pal.Paper, pal.Faint, 0.7), zGround)
		}
	case "dragons":
		if k > 0.15 && r < 0.12 {
			cv.Put(x, y, '·', pal.Faint, zGround)
		}
	case "blank":
		if k > 0.2 && r < 0.3 {
			cv.Put(x, y, '·', pal.mix(pal.Paper, pal.Faint, 0.6), zGround)
		}
	case "unsurveyed":
		if k > 0.2 && r < 0.4 {
			cv.Put(x, y, '·', pal.mix(pal.Paper, pal.Faint, 0.6), zGround)
		} else if (x+2*y)%9 == 0 {
			cv.Put(x, y, '╱', pal.mix(pal.Paper, pal.Faint, 0.3), zGround)
		}
	}
}

func (e *earthCtx) graticule(x, y, i, j int) {
	g := e.p.Graticule
	a := e.world(float64(i), float64(j))
	b := e.world(float64(i+1), float64(j+1))
	vx := math.Floor(a.X/g) != math.Floor(b.X/g)
	hy := math.Floor(a.Y/g) != math.Floor(b.Y/g)
	col := e.pal.mix(e.pal.Paper, e.pal.Faint, 0.55)
	if e.p.GridRefs {
		col = e.pal.mix(e.pal.Paper, e.pal.Water, 0.55)
	}
	if e.p.Night || e.p.Scan {
		col = e.pal.mix(e.pal.Paper, e.pal.Faint, 0.3)
	}
	switch {
	case vx && hy:
		e.cv.Put(x, y, '┼', col, zGrid)
	case vx:
		e.cv.Put(x, y, '┊', col, zGrid)
	case hy:
		e.cv.Put(x, y, '┈', col, zGrid)
	}
}

func (e *earthCtx) sea(x, y, i, j int, c Pt, r float64, cd []float64, corner func([]float64, int, int) (float64, float64, float64, float64)) {
	p, pal, cv := e.p, e.pal, e.cv
	d := e.a.W.CoastDist(c.X, c.Y)
	switch {
	case p.Relief == "shade":
		switch {
		case d < 3:
			cv.SetBg(x, y, pal.mix(pal.Paper, pal.Water, 0.34))
		case d < 9:
			cv.SetBg(x, y, pal.mix(pal.Paper, pal.Water, 0.22))
		default:
			cv.SetBg(x, y, pal.mix(pal.Paper, pal.WaterDeep, 0.26))
		}
	case p.Relief == "contour":
		cv.SetBg(x, y, pal.WaterWash)
	}
	for li, L := range p.WaterLines {
		tl, tr, br, bl := corner(cd, i, j)
		bits := 0
		if tl < L {
			bits |= 8
		}
		if tr < L {
			bits |= 4
		}
		if br < L {
			bits |= 2
		}
		if bl < L {
			bits |= 1
		}
		if bits != 0 && bits != 15 {
			col := pal.mix(pal.Paper, pal.Water, 0.85-float64(li)*0.22)
			cv.Put(x, y, msRound[bits], col, zWater)
			return
		}
	}
	if p.SeaDensity > 0 && r < p.SeaDensity {
		g := p.SeaGlyphs[int(r*1000)%len(p.SeaGlyphs)]
		cv.Put(x, y, g, pal.mix(pal.Paper, pal.Water, 0.7), zWater)
	}
	if p.Night && d < 1.2 {
		cv.Put(x, y, '·', pal.mix(pal.Paper, pal.Water, 0.35), zWater)
	}
}

func (e *earthCtx) relief(x, y, i, j int, c Pt, h, r float64, hc []float64, corner func([]float64, int, int) (float64, float64, float64, float64)) {
	p, pal, cv, w := e.p, e.pal, e.cv, e.a.W
	biome := w.Biome(c.X, c.Y, h)
	tl, tr, br, bl := corner(hc, i, j)
	switch p.Relief {
	case "marks":
		dens := p.MarkDens
		if biome == "mountains" || biome == "peaks" {
			dens *= 2.2
		}
		if biome == "grassland" || biome == "steppe" {
			dens *= 0.35
		}
		if biome == "hills" {
			dens *= 0.4
			if p.Frame == "hide" {
				dens = 0
			}
		}
		if g, ok := p.LandMarks[biome]; ok && r < dens {
			cv.Put(x, y, g, e.biomeInk(biome), zRelief)
		}
	case "hachure":
		gx := (tr + br - tl - bl) / 2
		gy := (bl + br - tl - tr) / 2
		mag := math.Hypot(gx, gy) / e.v.S
		switch {
		case h > 0.78 && r < 0.5:
			cv.Put(x, y, '▲', pal.Relief, zRelief)
		case mag > 0.035 && r < clamp(mag*9, 0, 0.85):
			cv.Put(x, y, dirGlyph(gx, gy, '-', '╎', '╱', '╲'), pal.Relief, zRelief)
		case biome == "forest" && r < 0.3:
			cv.Put(x, y, '♣', pal.Forest, zRelief)
		}
	case "contour":
		for L := math.Floor(math.Max(math.Max(tl, tr), math.Max(br, bl))*10) / 10; L >= 0.1; L -= 0.1 {
			bits := 0
			if tl > L {
				bits |= 8
			}
			if tr > L {
				bits |= 4
			}
			if br > L {
				bits |= 2
			}
			if bl > L {
				bits |= 1
			}
			if bits != 0 && bits != 15 {
				col, attr := pal.mix(pal.Paper, pal.Relief, 0.75), tcell.AttrMask(0)
				if int(math.Round(L*10))%5 == 0 {
					col, attr = pal.Relief, tcell.AttrBold
				}
				cv.PutA(x, y, msRound[bits], col, attr, zRelief)
				return
			}
		}
		if biome == "forest" && r < 0.34 {
			cv.Put(x, y, '♧', pal.Forest, zRelief)
		}
	case "none":
		if p.Night {
			cv.SetBg(x, y, pal.NightLand)
		}
	case "shade":
		cv.SetBg(x, y, pal.BiomeWash[biome])
		gx := (tr + br - tl - bl) / 2
		gy := (bl + br - tl - tr) / 2
		lit := -(gx + gy) / e.v.S * 6 // light from the north-west
		// the colour is the photograph; glyphs only where the relief or
		// the canopy would show, so it still reads without colour
		g, col := ' ', e.biomeInk(biome)
		switch {
		case biome == "peaks" || biome == "ice":
			g, col = '▲', pal.Snow
			if r > 0.4 {
				g = '░'
			}
		case lit < -0.3:
			g, col = '░', pal.mix(pal.Paper, pal.Rock, 0.8)
		case lit > 0.45:
			g, col = '▒', pal.mix(pal.Paper, e.biomeInk(biome), 0.6)
		case biome == "forest" && r < 0.3:
			g = '♣'
		case biome == "desert" && r < 0.2:
			g = '∴'
		case (biome == "mountains" || biome == "hills") && r < 0.25:
			g = '^'
		}
		cv.Put(x, y, g, col, zRelief)
	}
}

func (e *earthCtx) biomeInk(b string) tcell.Color {
	switch b {
	case "mountains", "hills":
		return e.pal.Relief
	case "peaks", "ice":
		return e.pal.Snow
	case "forest":
		return e.pal.Forest
	case "desert":
		return e.pal.Sand
	case "grassland", "steppe":
		return e.pal.Grass
	}
	return e.pal.Faint
}

// rhumbs: the portolan's web of wind lines from two roses at sea.
func (e *earthCtx) rhumbs() {
	a := e.a
	var roses []Pt
	for _, off := range []Pt{{-40, 26}, {50, -20}, {60, 40}, {-60, -30}} {
		best, bd := Pt{}, math.Inf(1)
		for k := 0; k < 300; k++ {
			p := Pt{a.Capital.X + off.X + (rnd01(a.W.Seed, k, 71)-0.5)*40, a.Capital.Y + off.Y + (rnd01(a.W.Seed, k, 72)-0.5)*30}
			if a.W.CoastDist(p.X, p.Y) < 5 {
				continue
			}
			if d := p.Dist(Pt{a.Capital.X + off.X, a.Capital.Y + off.Y}); d < bd {
				best, bd = p, d
			}
		}
		if !math.IsInf(bd, 1) {
			roses = append(roses, best)
		}
		if len(roses) == 2 {
			break
		}
	}
	for _, rose := range roses {
		sx, sy := e.screen(rose)
		for k := 0; k < 16; k++ {
			ang := float64(k) * math.Pi / 8
			col := e.pal.Faint
			switch {
			case k%4 == 0:
				col = e.pal.mix(e.pal.Paper, e.pal.Ink, 0.45)
			case k%2 == 0:
				col = e.pal.mix(e.pal.Paper, e.pal.Pos, 0.6)
			default:
				col = e.pal.mix(e.pal.Paper, e.pal.Neg, 0.6)
			}
			ex, ey := sx+math.Cos(ang)*400, sy+math.Sin(ang)*200
			e.cv.Line(float64(e.rc.X)+sx, float64(e.rc.Y)+sy, float64(e.rc.X)+ex, float64(e.rc.Y)+ey, func(i, x, y int, dx, dy float64) (rune, bool) {
				return '·', i > 2 && i%2 == 0 && !e.isLand(x, y) && e.knownAt(x, y) >= 0.5
			}, col, zGrid)
		}
		x, y, ok := e.cell(rose)
		if ok {
			e.cv.Stamp(x, y, []string{"  N  ", "╲ │ ╱", "─ ✺ ─", "╱ │ ╲", "  S  "}, e.pal.Ink, zMark)
			e.cv.Put(x, y, '✺', e.pal.Neg, zMark+1)
		}
	}
	// the dragons live in the unknown sea
	if e.p.Fog == "dragons" {
		placed := 0
		for k := 0; k < 400 && placed < 2; k++ {
			i := int(rnd01(a.W.Seed, k, 81) * float64(e.rc.W-14))
			j := int(rnd01(a.W.Seed, k, 82) * float64(e.rc.H-3))
			if e.known[j*e.rc.W+i] > 0.05 || e.known[j*e.rc.W+i+12] > 0.05 || e.known[(j+2)*e.rc.W+i] > 0.05 {
				continue
			}
			x, y := e.rc.X+i, e.rc.Y+j
			if e.cv.Reserved(x, y) || e.cv.Reserved(x+12, y+2) {
				continue
			}
			e.cv.Stamp(x+6, y+1, []string{"  ,~⌒~,  ", "∿∿∿∿∿ɕ∿∿∿", ""}, e.pal.mix(e.pal.Paper, e.pal.Neg, 0.7), zMark)
			if placed == 0 {
				e.cv.Text(x, y+3, "hic svnt dracones", e.pal.mix(e.pal.Paper, e.pal.Faint, 0.8), tcell.AttrItalic, zLabel)
				e.cv.Reserve(x, y+3, 17, 1)
			}
			placed++
		}
	}
}

func (e *earthCtx) polyline(path []Pt, pick func(k int, land bool, dx, dy float64) (rune, bool), col func(k int) tcell.Color, z int8, knownOnly bool) {
	for k := 0; k+1 < len(path); k++ {
		x0, y0 := e.screen(path[k])
		x1, y1 := e.screen(path[k+1])
		if (x0 < -2 && x1 < -2) || (y0 < -2 && y1 < -2) || (x0 > float64(e.rc.W)+2 && x1 > float64(e.rc.W)+2) || (y0 > float64(e.rc.H)+2 && y1 > float64(e.rc.H)+2) {
			continue
		}
		kk := k
		n := int(math.Max(math.Abs(x1-x0), math.Abs(y1-y0)*2)) + 1
		lx, ly := -9999, -9999
		for s := 0; s <= n; s++ {
			t := float64(s) / float64(n)
			x := e.rc.X + int(math.Floor(x0+(x1-x0)*t))
			y := e.rc.Y + int(math.Floor(y0+(y1-y0)*t))
			if x == lx && y == ly {
				continue
			}
			lx, ly = x, y
			if x < e.rc.X || y < e.rc.Y || x >= e.rc.X+e.rc.W || y >= e.rc.Y+e.rc.H {
				continue
			}
			if knownOnly && e.knownAt(x, y) < 0.5 {
				continue
			}
			if r, ok := pick(kk*1000+s, e.isLand(x, y), x1-x0, y1-y0); ok {
				e.cv.Put(x, y, r, col(kk), z)
			}
		}
	}
}

func (e *earthCtx) rivers() {
	if e.p.RiverH == ' ' {
		return
	}
	col := e.pal.Water
	for _, rv := range e.a.W.Rivers {
		e.polyline(rv, func(k int, land bool, dx, dy float64) (rune, bool) {
			if !land {
				return 0, false
			}
			return dirGlyph(dx, dy, e.p.RiverH, e.p.RiverV, e.p.RiverH, e.p.RiverH), true
		}, func(int) tcell.Color { return col }, zRiver, true)
	}
}

// pointAlong returns the point at fraction t of a polyline's length.
func pointAlong(path []Pt, t float64) Pt {
	if len(path) == 0 {
		return Pt{}
	}
	total := 0.0
	for i := 1; i < len(path); i++ {
		total += path[i].Dist(path[i-1])
	}
	want := t * total
	for i := 1; i < len(path); i++ {
		d := path[i].Dist(path[i-1])
		if want <= d && d > 0 {
			f := want / d
			return Pt{lerp(path[i-1].X, path[i].X, f), lerp(path[i-1].Y, path[i].Y, f)}
		}
		want -= d
	}
	return path[len(path)-1]
}

func (e *earthCtx) lines() {
	a, p, pal := e.a, e.p, e.pal
	// expedition trails (the explorers' age: not from orbit)
	if !p.Night && !p.Scan && !p.Mini {
		recent := map[string]bool{}
		// the three most recently travelled
		for n := 0; n < 3; n++ {
			best, bt := "", -1
			for _, t := range a.Trails {
				if !recent[t.Key] && t.LastTick > bt {
					best, bt = t.Key, t.LastTick
				}
			}
			recent[best] = true
		}
		for _, t := range a.Trails {
			col := pal.mix(pal.Paper, pal.Faint, 0.8)
			if recent[t.Key] {
				col = pal.Label
			}
			e.polyline(t.Path, func(k int, land bool, dx, dy float64) (rune, bool) {
				return p.TrailRune, k%2 == 0
			}, func(int) tcell.Color { return col }, zLine-1, false)
			if end := t.Path[len(t.Path)-1]; true {
				if x, y, ok := e.cell(end); ok {
					e.cv.Put(x, y, '×', col, zLine)
				}
			}
		}
	}
	for _, t := range a.Trails {
		if !t.Active {
			continue
		}
		n := int(float64(len(t.Path)-1) * t.Progress)
		e.polyline(t.Path[:n+1], func(k int, land bool, dx, dy float64) (rune, bool) {
			return '·', true
		}, func(int) tcell.Color { return pal.High }, zLine+1, false)
		if x, y, ok := e.cell(t.Path[n]); ok {
			glyph := '⚑'
			if e.frame%2 == 1 {
				glyph = '⚐'
			}
			e.cv.PutA(x, y, glyph, pal.High, tcell.AttrBold, zMover)
		}
	}
	// trade routes, with goods moving along them
	for ri, r := range a.Routes {
		col := pal.High
		if r.Partner != "" {
			for _, c := range a.Civs {
				if c.Key == r.Partner {
					col = pal.Rel(relation(c))
				}
			}
		}
		if r.Disrupted {
			col = pal.mix(pal.Paper, pal.Neg, 0.6)
		}
		e.polyline(r.Path, func(k int, land bool, dx, dy float64) (rune, bool) {
			if land {
				return p.RouteLand, k%2 == 0 || p.RouteLand == '┿'
			}
			return p.RouteSea, k%2 == 0
		}, func(int) tcell.Color { return col }, zLine, false)
		if r.Disrupted {
			if x, y, ok := e.cell(pointAlong(r.Path, 0.5)); ok {
				e.cv.PutA(x, y, '✕', pal.Neg, tcell.AttrBold, zMark)
			}
			continue
		}
		for k := 0; k < 2; k++ {
			ph := math.Mod(float64(e.frame)*0.045+float64(ri)*0.37+float64(k)*0.5, 1)
			if k == 1 {
				ph = 1 - ph // the return cargo
			}
			pt := pointAlong(r.Path, ph)
			if x, y, ok := e.cell(pt); ok {
				g := p.Caravan
				if !e.isLand(x, y) {
					g = p.Ship
				}
				e.cv.PutA(x, y, g, col, tcell.AttrBold, zMover)
			}
		}
	}
	// war fronts and embargo blockades
	for _, c := range a.Civs {
		if c.Celestial || !c.Known {
			continue
		}
		for k, f := range c.Front {
			if k%2 == 1 {
				continue
			}
			if x, y, ok := e.cell(f); ok {
				e.cv.PutA(x, y, p.Front, pal.Neg, tcell.AttrBold, zMark)
			}
		}
		if c.Info.Status == "embargo" {
			for k := 0; k < 28; k++ {
				ang := float64(k) / 28 * 2 * math.Pi
				pt := Pt{c.Seat.X + math.Cos(ang)*9, c.Seat.Y + math.Sin(ang)*6}
				if x, y, ok := e.cell(pt); ok {
					e.cv.Put(x, y, '▫', pal.Warn, zLine+1)
				}
			}
		}
	}
	// satellite ground tracks and the radar sweep
	if p.Tracks {
		for k := 0; k < 2; k++ {
			amp := float64(e.rc.H) * (0.3 + 0.1*float64(k))
			ph := float64(k) * 2.1
			for i := 0; i < e.rc.W; i += 2 {
				y := float64(e.rc.H)/2 + amp*math.Sin(float64(i)/float64(e.rc.W)*2*math.Pi+ph)
				e.cv.Put(e.rc.X+i, e.rc.Y+int(y), '·', pal.mix(pal.Paper, pal.Accent, 0.45), zGrid)
			}
			si := (e.frame*3 + k*37) % e.rc.W
			y := float64(e.rc.H)/2 + amp*math.Sin(float64(si)/float64(e.rc.W)*2*math.Pi+ph)
			e.cv.PutA(e.rc.X+si, e.rc.Y+int(y), '◇', pal.High, tcell.AttrBold, zMover)
		}
	}
	if p.Sweep {
		cx, cy := e.screen(a.Capital)
		R := math.Min(float64(e.rc.W)*0.32, float64(e.rc.H)*0.62)
		for _, rr := range []float64{R / 2, R} {
			for k := 0; k < 72; k++ {
				ang := float64(k) / 72 * 2 * math.Pi
				x := e.rc.X + int(cx+math.Cos(ang)*rr)
				y := e.rc.Y + int(cy+math.Sin(ang)*rr/2)
				e.cv.Put(x, y, '·', pal.mix(pal.Paper, pal.Pos, 0.5), zGrid)
			}
		}
		th := float64(e.frame) * 0.4
		for k, fade := range []float64{1, 0.7, 0.45, 0.25} {
			ang := th - float64(k)*0.07
			col := pal.mix(pal.Paper, pal.Pos, fade)
			e.cv.Line(float64(e.rc.X)+cx, float64(e.rc.Y)+cy, float64(e.rc.X)+cx+math.Cos(ang)*R, float64(e.rc.Y)+cy+math.Sin(ang)*R/2,
				func(i, x, y int, dx, dy float64) (rune, bool) {
					return dirGlyph(dx, dy, '─', '│', '╱', '╲'), i > 0
				}, col, zLine)
		}
		// blips: bright just after the beam passes
		for _, c := range a.Civs {
			if c.Celestial || !c.Known {
				continue
			}
			sx, sy := e.screen(c.Seat)
			ang := math.Atan2((sy-cy)*2, sx-cx)
			behind := math.Mod(th-ang+8*math.Pi, 2*math.Pi)
			if x, y, ok := e.cell(c.Seat); ok && behind < 2.2 {
				e.cv.PutA(x, y, '◆', pal.mix(pal.Paper, pal.Pos, 1-behind/2.6), tcell.AttrBold, zMover)
			}
		}
	}
}

func (e *earthCtx) marks() {
	a, p, pal, cv := e.a, e.p, e.pal, e.cv
	// city lights on the night pass
	if p.Night {
		for _, t := range a.Towns {
			n, col := 26, pal.Light1
			if t.Capital && t.Owner == "" {
				n = 60 + a.Buildings/12
			}
			if t.Owner != "" {
				for _, c := range a.Civs {
					if c.Key == t.Owner {
						if !c.Known {
							n = 0
						}
						col = pal.mix(pal.Paper, pal.Rel(relation(c)), 0.8)
					}
				}
			}
			h := hashStr(t.Name)
			for k := 0; k < n; k++ {
				rr := math.Sqrt(rnd01(h, k, 1)) * (3 + math.Sqrt(float64(n)))
				ang := rnd01(h, k, 2) * 2 * math.Pi
				pt := Pt{t.At.X + math.Cos(ang)*rr*1.3, t.At.Y + math.Sin(ang)*rr}
				if x, y, ok := e.cell(pt); ok && e.isLand(x, y) {
					g := '·'
					c2 := col
					switch {
					case rr < 2:
						g = '•'
					case rr < 4:
						g = '∙'
					default:
						c2 = pal.mix(pal.Paper, col, 0.6)
						if t.Owner == "" {
							c2 = pal.mix(pal.Paper, pal.Light2, 0.8)
						}
					}
					if (e.frame+k)%23 == 0 {
						c2 = pal.Paper // a light going out, briefly
					}
					cv.Put(x, y, g, c2, zBorder+1)
				}
			}
		}
	}
	// scars
	for _, s := range a.Scars {
		x, y, ok := e.cell(s.At)
		if !ok {
			continue
		}
		col := pal.Neg
		if s.Outcome == "endured" {
			col = pal.mix(pal.Paper, pal.Neg, 0.75)
		}
		rr := s.R / e.v.S
		for k := 0; k < 20; k++ {
			ang := float64(k) / 20 * 2 * math.Pi
			cv.Put(x+int(math.Cos(ang)*rr), y+int(math.Sin(ang)*rr/2), '·', col, zMark-1)
		}
		cv.PutA(x, y, p.Scar, col, tcell.AttrBold, zMark)
	}
	// omens: the harbinger's origin, pulsing while the thread is live
	for _, o := range a.Omens {
		x, y, ok := e.cell(o.At)
		if !ok {
			continue
		}
		if o.Active {
			col := pal.Neg
			if e.frame%2 == 1 {
				col = pal.Warn
			}
			cv.PutA(x, y, p.Omen, col, tcell.AttrBold, zMark+1)
			for k := 1; k <= 4; k++ { // a tail pointing away from home
				ax, ay := e.screen(a.Capital)
				ox, oy := e.screen(o.At)
				dx, dy := ox-ax, oy-ay
				l := math.Hypot(dx, dy)
				if l > 0 {
					cv.Put(x+int(dx/l*float64(k)*1.2), y+int(dy/l*float64(k)*0.6), '·', pal.mix(pal.Paper, col, 1-float64(k)*0.2), zMark)
				}
			}
		} else {
			cv.Put(x, y, p.Omen, pal.mix(pal.Paper, pal.Faint, 0.8), zMark)
		}
	}
	// wonders
	for _, wd := range a.Wonders {
		if x, y, ok := e.cell(wd.At); ok {
			cv.PutA(x, y, p.Wonder, pal.High, tcell.AttrBold, zMark)
		}
	}
	// the Stone Era camp: one mark per kind of building
	for _, m := range a.Camps {
		if x, y, ok := e.cell(m.At); ok {
			cv.PutA(x, y, campGlyph(a.St.Buildings[m.Key].Category), pal.Accent, 0, zMark)
		}
	}
	// towns and capitals
	for _, t := range a.Towns {
		x, y, ok := e.cell(t.At)
		if !ok {
			continue
		}
		if t.Owner == "" {
			col, g := pal.Accent, p.Town
			attr := tcell.AttrMask(0)
			if t.Capital {
				g, attr = p.Capital, tcell.AttrBold
			}
			if t.Fresh && !t.Capital {
				col = pal.High
			}
			cv.PutA(x, y, g, col, attr, zMark+2)
			continue
		}
		for _, c := range a.Civs {
			if c.Key == t.Owner && c.Known {
				cv.PutA(x, y, p.CivCapital, pal.Rel(relation(c)), tcell.AttrBold, zMark+2)
			}
		}
	}
}

func campGlyph(cat string) rune {
	switch cat {
	case "housing":
		return '⌂'
	case "production":
		return '⚒'
	case "research":
		return '◎'
	case "military":
		return '†'
	case "storage":
		return '▪'
	case "wonder":
		return '✦'
	}
	return '•'
}

var nearOffsets = [][2]int{{2, 0}, {-998, 0}, {-999, -1}, {-999, 1}, {2, -1}, {2, 1}, {-998, -1}, {-998, 1}}

func spaced(s string) string {
	var b strings.Builder
	for i, r := range strings.ToUpper(s) {
		if i > 0 {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return strings.ReplaceAll(b.String(), "   ", "  ")
}

func (e *earthCtx) caps(s string) string {
	if e.p.Caps {
		return strings.ToUpper(s)
	}
	return s
}

func (e *earthCtx) labels() {
	a, p, pal, cv := e.a, e.p, e.pal, e.cv
	b := [4]int{e.rc.X, e.rc.Y, e.rc.X + e.rc.W, e.rc.Y + e.rc.H}
	// reserve every mark so no label covers one
	for i := range cv.C {
		if cv.C[i].Z >= zMark {
			cv.lab[i] = true
		}
	}
	// the capital
	if x, y, ok := e.cell(a.Capital); ok {
		cv.Label(x, y, a.Towns[0].Name, pal.Accent, tcell.AttrBold, zLabel, nearOffsets, b)
	}
	// omens
	for _, o := range a.Omens {
		if !o.Active {
			continue
		}
		if x, y, ok := e.cell(o.At); ok {
			s := o.Name
			if o.Numeric && o.Prob > 0 {
				s = fmt.Sprintf("%s %d%%", o.Name, int(o.Prob*100+0.5))
			}
			cv.Label(x, y, s, pal.Neg, tcell.AttrItalic, zLabel, nearOffsets, b)
		}
	}
	// civilizations
	for _, c := range a.Civs {
		if c.Celestial || !c.Known {
			continue
		}
		x, y, ok := e.cell(c.Seat)
		if !ok {
			continue
		}
		name := c.Def.Name
		if p.Caps {
			name = spaced(name)
			if uniseg.StringWidth(name) > e.rc.W/3 {
				name = strings.ToUpper(c.Def.Name)
			}
		}
		// the standing is spelled out too, so it survives without colour
		switch relation(c) {
		case "war":
			name += " ✕war"
		case "allied":
			name += " +ally"
		case "rival":
			name += " !" + c.Info.Status
		}
		col := pal.Rel(relation(c))
		if !cv.Label(x, y, name, col, tcell.AttrBold, zLabel, [][2]int{{-999, 1}, {-999, -1}, {2, 0}, {-998, 0}, {-999, 2}, {-999, -2}}, b) {
			cv.Label(x, y, strings.ToUpper(c.Def.Name), col, tcell.AttrBold, zLabel, nearOffsets, b)
		}
		if c.Fresh {
			cv.Label(x, y, "new", pal.High, tcell.AttrReverse, zLabel, [][2]int{{2, 0}, {-998, 0}, {2, -1}}, b)
		}
		if c.Info.AtWar && len(c.Front) > 0 {
			if fx, fy, ok := e.cell(c.Front[len(c.Front)/2]); ok {
				cv.Label(fx, fy, "front", pal.Neg, tcell.AttrItalic, zLabel, nearOffsets, b)
			}
		}
	}
	// the camp's buildings, while there are few enough to name
	if len(a.Camps) <= 6 {
		for _, m := range a.Camps {
			if x, y, ok := e.cell(m.At); ok {
				cv.Label(x, y, strings.ToLower(m.Name), pal.Label, tcell.AttrItalic, zLabel, nearOffsets, b)
			}
		}
	}
	// scars
	for _, s := range a.Scars {
		if p.Mini {
			break
		}
		if x, y, ok := e.cell(s.At); ok {
			cv.Label(x, y, s.Name, pal.mix(pal.Paper, pal.Neg, 0.85), tcell.AttrItalic, zLabel, nearOffsets, b)
		}
	}
	// regions: big enough on screen, known; not from orbit
	for _, r := range a.W.Regions {
		if p.Night || p.Relief == "shade" || p.Mini {
			break
		}
		x, y, ok := e.cell(r.Centroid)
		if !ok || e.knownAt(x, y) < 0.9 || r.Area == 0 {
			continue
		}
		width := math.Sqrt(float64(r.Area)) / e.v.S
		name := r.Name
		if p.Caps && width > float64(len(name))*2+2 {
			name = spaced(name)
		} else if p.Caps {
			name = strings.ToUpper(name)
		}
		if width < float64(uniseg.StringWidth(name))*0.7 {
			continue
		}
		col := pal.mix(pal.Paper, pal.Label, 0.75)
		if o := a.Owner[r.ID]; o == 0 && a.Owned > 0 {
			col = pal.mix(pal.Paper, pal.Accent, 0.8)
		}
		cv.Label(x, y, name, col, 0, zLabel-1, [][2]int{{-999, 0}, {-999, 1}, {-999, -1}}, b)
	}
	// seas
	for _, s := range a.W.Seas {
		x, y, ok := e.cell(s.At)
		if !ok || e.knownAt(x, y) < 0.9 || p.Mini {
			continue
		}
		cv.Label(x, y, e.caps(s.Name), pal.mix(pal.Paper, pal.Water, 0.9), tcell.AttrItalic, zLabel-1, [][2]int{{-999, 0}, {-999, 1}, {-999, -1}}, b)
	}
	// towns, when there is room to name them
	for _, t := range a.Towns[1:] {
		if t.Owner != "" || (e.v.S > 1.3 && !t.Fresh) || p.Mini {
			continue
		}
		if x, y, ok := e.cell(t.At); ok {
			col := pal.mix(pal.Paper, pal.Ink, 0.8)
			if t.Fresh {
				col = pal.High
			}
			cv.Label(x, y, t.Name, col, 0, zLabel-2, nearOffsets, b)
		}
	}
}
