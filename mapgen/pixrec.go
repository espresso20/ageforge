package mapgen

import "image/color"

// A Pixels can note its calls as it draws them (Record). The cell path folds
// the pixels into half-blocks as ever; the picture path replays the notes at
// the window's pixel size, where a disc is round, a line is smooth and a Set is
// a point of light. A position names the pixel it falls in and is drawn at that
// pixel's middle, as the cells draw it.

const (
	opSet = iota
	opLine
	opEllipse
	opDisc
	opGlow
)

// pixOp is one noted call. x, y, rx, ry are what the call took: pixels for Set
// and Line, sixteenths of a pixel for the rest (rx, ry the radii; Glow's inner
// and outer radius).
type pixOp struct {
	kind         uint8
	x, y, rx, ry int
	p            Paint
	gap, half    int
	steps        int
	shade        func(nx, ny int) Paint
}

// Record makes p note its calls in ops.
func (p *Pixels) Record(ops *[]pixOp) { p.rec = ops }

// pixRaster draws noted calls: sx and sy are picture pixels to a canvas pixel.
type pixRaster struct {
	w      int
	sx, sy float64
	hue    [NumHues]color.RGBA
}

func (r *pixRaster) rgb(h Hue) [3]float32 {
	c := r.hue[h]
	return [3]float32{float32(c.R), float32(c.G), float32(c.B)}
}

// replay draws the notes onto rows y0..y1 of rgb (r.w wide).
func (r *pixRaster) replay(ops []pixOp, y0, y1 int, rgb []float32) {
	k := r.sx / 4.5 // a point is a pixel or two wide in a cell of 9
	for _, o := range ops {
		wt := float64(o.p.Weight) / 255
		col := r.rgb(o.p.Hue)
		switch o.kind {
		case opSet:
			splat(rgb, r.w, y0, y1, (float64(o.x)+0.5)*r.sx, (float64(o.y)+0.5)*r.sy, (0.6+0.9*wt)*k, o.p.Weight >= 200, col, wt)
		case opLine:
			r.seg(rgb, y0, y1, (float64(o.x)+0.5)*r.sx, (float64(o.y)+0.5)*r.sy, (float64(o.rx)+0.5)*r.sx, (float64(o.ry)+0.5)*r.sy, 0.7, col, float32(wt))
		case opEllipse:
			r.ellipse(rgb, y0, y1, o, col, float32(wt))
		case opDisc:
			r.disc(rgb, y0, y1, o)
		case opGlow:
			r.glow(rgb, y0, y1, o, col)
		}
	}
}

// blend lays colour col over pixel (x, y) of the rows, a share a of the way.
func (r *pixRaster) blend(rgb []float32, y0, y1, x, y int, col [3]float32, a float32) {
	if x < 0 || x >= r.w || y < y0 || y >= y1 || a < 0.004 {
		return
	}
	j := 3 * ((y-y0)*r.w + x)
	a = min(a, 1)
	rgb[j] += (col[0] - rgb[j]) * a
	rgb[j+1] += (col[1] - rgb[j+1]) * a
	rgb[j+2] += (col[2] - rgb[j+2]) * a
}

// seg draws the line from (ax, ay) to (bx, by), hw pixels either side of the
// middle, with soft edges: along its long side one step at a time, coverage by
// the distance across.
func (r *pixRaster) seg(rgb []float32, y0, y1 int, ax, ay, bx, by, hw float64, col [3]float32, a float32) {
	steep := absF(by-ay) > absF(bx-ax)
	p0, m0, p1, m1 := ax, ay, bx, by // along, across
	if steep {
		p0, m0, p1, m1 = ay, ax, by, bx
	}
	if p1 < p0 {
		p0, m0, p1, m1 = p1, m1, p0, m0
	}
	slope := 0.0
	if p1 > p0 {
		slope = (m1 - m0) / (p1 - p0)
	}
	cs := 1 / sqrtF(1+slope*slope)
	for i := floorF(p0); i <= floorF(p1); i++ {
		pc := float64(i) + 0.5
		end := min(1, max(0, 0.5+min(pc-p0, p1-pc)))
		m := m0 + slope*min(max(pc, p0), p1)
		m -= slope * p0
		for j := floorF(m - hw - 1); j <= floorF(m+hw+1); j++ {
			cov := min(1, max(0, hw+0.5-absF(float64(j)+0.5-m)*cs)) * end
			if steep {
				r.blend(rgb, y0, y1, j, i, col, a*float32(cov))
			} else {
				r.blend(rgb, y0, y1, i, j, col, a*float32(cov))
			}
		}
	}
}

// ellipse draws a ring, dashed when the note says so, as short smooth lines.
func (r *pixRaster) ellipse(rgb []float32, y0, y1 int, o pixOp, col [3]float32, a float32) {
	cx, cy := float64(o.x)/pxUnit+0.5, float64(o.y)/pxUnit+0.5
	rx, ry := float64(o.rx)/pxUnit*r.sx, float64(o.ry)/pxUnit*r.sy
	cx, cy = cx*r.sx, cy*r.sy
	if int((cy-ry)-1) > y1 || int(cy+ry+1) < y0 {
		return
	}
	n := min(8192, max(24, floorF((rx+ry)*3/2)))
	px, py := cx+rx, cy
	arc := 0.0
	for i := 1; i <= n; i++ {
		t := float64(i) * turn / float64(n)
		qx, qy := cx+rx*cosF(t), cy+ry*sinF(t)
		mid := (float64(i) - 0.5) * turn / float64(n)
		l := sqrtF((qx-px)*(qx-px) + (qy-py)*(qy-py))
		s := sinF(mid)
		on := (o.half <= 0 || s >= 0) && (o.half >= 0 || s <= 0)
		if o.gap > 0 && floorF((arc+l/2)/r.sx)%(o.gap+1) != 0 {
			on = false
		}
		if on && (min(py, qy) < float64(y1)+2 && max(py, qy) > float64(y0)-2) {
			r.seg(rgb, y0, y1, px, py, qx, qy, 0.7, col, a)
		}
		arc += l
		px, py = qx, qy
	}
}

// disc draws a round body of radius r, shaded by the note's shade function.
func (r *pixRaster) disc(rgb []float32, y0, y1 int, o pixOp) {
	cx, cy := float64(o.x)/pxUnit*r.sx, float64(o.y)/pxUnit*r.sx // round: the radius sets both
	rad := float64(max(o.rx, 1)) / pxUnit * r.sx
	cy = float64(o.y) / pxUnit * r.sy
	for y := max(y0, floorF(cy-rad)-1); y <= min(y1-1, floorF(cy+rad)+1); y++ {
		for x := max(0, floorF(cx-rad)-1); x <= min(r.w-1, floorF(cx+rad)+1); x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			cov := min(1, max(0, rad-sqrtF(dx*dx+dy*dy)+0.5))
			if cov == 0 {
				continue
			}
			p := o.shade(max(-256, min(256, int(dx*256/rad))), max(-256, min(256, int(dy*256/rad))))
			if p.Hue == HNone || p.Weight == 0 {
				continue
			}
			r.blend(rgb, y0, y1, x, y, r.rgb(p.Hue), float32(cov)*float32(p.Weight)/255)
		}
	}
}

// glow draws a halo that fades from its inner radius to its outer one.
func (r *pixRaster) glow(rgb []float32, y0, y1 int, o pixOp, col [3]float32) {
	cx, cy := float64(o.x)/pxUnit*r.sx, float64(o.y)/pxUnit*r.sy
	r0, r1 := float64(o.rx)/pxUnit*r.sx, float64(o.ry)/pxUnit*r.sx
	w0 := float32(o.p.Weight) / 255 * float32(o.steps) / float32(o.steps+1)
	for y := max(y0, floorF(cy-r1)); y <= min(y1-1, floorF(cy+r1)); y++ {
		for x := max(0, floorF(cx-r1)); x <= min(r.w-1, floorF(cx+r1)); x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			d := sqrtF(dx*dx + dy*dy)
			if d >= r1 {
				continue
			}
			f := 1 - float32(max(0, d-r0)/(r1-r0))
			r.blend(rgb, y0, y1, x, y, col, w0*f*f)
		}
	}
}

// splat lays a soft point of light of width sigma on rows y0..y1: a gaussian,
// stronger the brighter wt is, with a faint halo round the brightest.
func splat(rgb []float32, w, y0, y1 int, px, py, sigma float64, glow bool, col [3]float32, wt float64) {
	reach := 3 * sigma
	if glow {
		reach = 7 * sigma / 0.9
	}
	for y := max(y0, floorF(py-reach)); y <= min(y1-1, floorF(py+reach)); y++ {
		for x := max(0, floorF(px-reach)); x <= min(w-1, floorF(px+reach)); x++ {
			dx, dy := float64(x)+0.5-px, float64(y)+0.5-py
			d2 := dx*dx + dy*dy
			a := expNeg(d2 / (2 * sigma * sigma))
			if glow {
				a += 0.22 * expNeg(d2/(2*3.5*sigma*sigma))
			}
			a = min(1, a*min(1, wt*1.5+0.1))
			if a < 0.01 {
				continue
			}
			j := 3 * ((y-y0)*w + x)
			f := float32(a)
			rgb[j] += (col[0] - rgb[j]) * f
			rgb[j+1] += (col[1] - rgb[j+1]) * f
			rgb[j+2] += (col[2] - rgb[j+2]) * f
		}
	}
}

func absF(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// sqrtF is the square root by Newton's method, scaled into 1..4 first.
func sqrtF(x float64) float64 {
	if x <= 0 {
		return 0
	}
	e := 1.0
	for x >= 4 {
		x /= 4
		e *= 2
	}
	for x < 1 {
		x *= 4
		e /= 2
	}
	g := (1 + x) / 2
	for i := 0; i < 5; i++ {
		g = (g + x/g) / 2
	}
	return g * e
}
