package cabinet

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/espresso20/ageforge/pkg/coins"
)

var (
	baize  = ramp(0x0b1a12, 0x112619, 0x173323, 0x1f4230, 0x2a543d)
	velvet = ramp(0x0a0f20, 0x101933, 0x172646, 0x203660, 0x2b4778)
	felt   = ramp(0x101113, 0x191b1e, 0x24272b, 0x32363b, 0x434850)
	foam   = ramp(0x07080a, 0x0d0f12, 0x13161a, 0x1a1e23, 0x232830)
	alu    = ramp(0x14171c, 0x21262d, 0x303741, 0x424b57, 0x59626f)
)

// Copy is one held piece.
type Copy struct {
	Level int
	Seed  uint32
}

// Tray is what an open drawer shows. Stacks hold the copies finest first; an
// empty stack is a well with the outline of its piece. A label under an empty
// well (or the legendary's, when not held) takes the dim plate colours.
type Tray struct {
	Slot   int // drawer index 0..21; picks the lining and the frame
	Money  string
	Types  [Wells]string
	Stacks [Wells][]Copy
	Legend bool
}

// mask is the set of pixels a piece covers.
type mask struct {
	w, h int
	a    []bool
}

func (m mask) at(x, y int) bool { return x >= 0 && y >= 0 && x < m.w && y < m.h && m.a[y*m.w+x] }

// shape is the opaque part of im (a disc of n pixels where there is none)
// laid at each offset, which must not be negative.
func shape(im *image.RGBA, n int, offs []image.Point) mask {
	var mx, my int
	for _, p := range offs {
		mx, my = max(mx, p.X), max(my, p.Y)
	}
	m := mask{n + mx, n + my, make([]bool, (n+mx)*(n+my))}
	for _, p := range offs {
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				if im == nil {
					m.a[(y+p.Y)*m.w+x+p.X] = m.a[(y+p.Y)*m.w+x+p.X] || math.Hypot(float64(x-n/2)+0.5, float64(y-n/2)+0.5) <= float64(n)/2
				} else if im.Pix[(y*n+x)*4+3] > 0 {
					m.a[(y+p.Y)*m.w+x+p.X] = true
				}
			}
		}
	}
	return m
}

// grow widens the mask by r in every direction; the seed varies how round.
func (m mask) grow(r, seed int) mask {
	o := mask{m.w + 2*r, m.h + 2*r, make([]bool, (m.w+2*r)*(m.h+2*r))}
	lim := r*r + r/2 + r*(seed%4)/2
	for y := 0; y < m.h; y++ {
		for x := 0; x < m.w; x++ {
			if !m.at(x, y) {
				continue
			}
			o.a[(y+r)*o.w+x+r] = true
			if m.at(x-1, y) && m.at(x+1, y) && m.at(x, y-1) && m.at(x, y+1) {
				continue
			}
			for dy := -r; dy <= r; dy++ {
				for dx := -r; dx <= r; dx++ {
					if dx*dx+dy*dy <= lim {
						o.a[(y+r+dy)*o.w+x+r+dx] = true
					}
				}
			}
		}
	}
	return o
}

// par runs f on horizontal bands of r at once.
func par(r image.Rectangle, f func(image.Rectangle)) {
	var wg sync.WaitGroup
	for k := 0; k < 8 && !r.Empty(); k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f(image.Rect(r.Min.X, r.Min.Y+k*r.Dy()/8, r.Max.X, r.Min.Y+(k+1)*r.Dy()/8))
		}()
	}
	wg.Wait()
}

// surface fills r with lining (stippled felt, or faceted foam) or brushed
// metal, from the pixel's place alone, so a patch matches the whole tray.
func (c *canvas) surface(r, lin image.Rectangle, rp []color.RGBA, kind, seed int) {
	r = r.Intersect(c.img.Rect.Add(c.org))
	p, wd := max(8, c.sc(14)), max(4, c.sc(20))
	for y := r.Min.Y; y < r.Max.Y; y++ {
		row := hash(0, y, seed)
		for x := r.Min.X; x < r.Max.X; x++ {
			v := 0.56 + 0.07*(hash(x, y, seed)-0.5) + 0.14*(valueNoise(float64(x)/55, float64(y)/55, seed+3)-0.5)
			switch kind {
			case 1: // brushed metal
				v = 0.46 + 0.22*(row-0.5) + 0.18*(valueNoise(float64(x)/90, float64(y)*3, seed+1)-0.5)
			case 2: // foam facets: lit on top, dark below
				a, b := (x%p)-p/2, (y%p)-p/2
				v = 0.4
				if abs(a) < abs(b) {
					v = 0.3 + 0.48*float64(btoi(b < 0))
				} else if a < 0 {
					v = 0.64
				}
			}
			if kind != 1 { // the inner walls throw a hard shadow
				v -= 0.3 * float64(max(0, 2-min(x-lin.Min.X, y-lin.Min.Y)/wd))
			}
			c.px(x, y, tone(rp, v-0.3*c.fall(x, y), x, y))
		}
	}
}

func abs(v int) int { return max(v, -v) }

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// cut paints a well, the mask m with its corner at wx, wy, into the lining:
// a floor, a dark cut edge upper left and a lip lower right (brass when
// selected). The seed varies the rim, the lip and the grain of the floor.
func (c *canvas) cut(m mask, wx, wy int, rp []color.RGBA, sel, recess bool, seed int) {
	floor, d1, d2, lip, lipN := mul(rp[1], 0.72), mul(rp[0], 0.28), mul(rp[0], 0.55), rp[3], 2+seed%2
	if recess {
		floor, d1, d2 = mul(rp[1], 1.1), mul(rp[0], 0.45), mul(rp[0], 0.95)
	}
	if sel {
		lip, lipN = brass[4], 3
	}
	rim, spk := 2+seed/3%2, 0.8+0.03*float64(seed%5)
	for y := -3; y < m.h+3; y++ {
		for x := -3; x < m.w+3; x++ {
			if m.at(x, y) {
				s := 0
				for t := 1; t <= 6 && s == 0; t++ {
					if !m.at(x-t, y) || !m.at(x, y-t) || !m.at(x-t, y-t) {
						s = t
					}
				}
				col := floor
				if s > rim {
					col = d2
				} else if s > 0 {
					col = d1
				} else if hash(wx+x, wy+y, seed) > spk {
					col = mul(floor, 1.1)
				}
				c.px(wx+x, wy+y, col)
				continue
			}
			for t := 1; t <= lipN; t++ {
				if m.at(x-t, y) || m.at(x, y-t) || m.at(x-t, y-t) {
					c.px(wx+x, wy+y, lip)
					break
				}
			}
		}
	}
}

// well paints type well i: cut to the shape of its piece, the stack in it.
func (c *canvas) well(l TrayLayout, t Tray, i int, mint *coins.Mint, sel bool, rp []color.RGBA) {
	st, cp, seed := t.Stacks[i], l.ctr[i], t.Slot*16+i+1
	n := min(len(st), 5)
	ref := get(mint, t.Money, t.Types[i], coins.Levels, 0, l.piece)
	offs := make([]image.Point, max(n, 1))
	for k := 1; k < len(offs); k++ {
		offs[k] = image.Pt(k*l.step+int(hash(k, seed, 31)*3), k*l.step+int(hash(k, seed, 32)*3))
	}
	half := offs[len(offs)-1].Div(2)
	m := shape(ref, l.piece, offs).grow(l.grow, seed)
	c.cut(m, cp.X-l.piece/2-half.X-l.grow, cp.Y-l.piece/2-half.Y-l.grow, rp, sel, n == 0, seed)
	if n == 0 {
		if ref != nil {
			c.outline(cp.X-half.X, cp.Y-half.Y, ref)
		}
		return
	}
	lift, sh := 0, max(2, c.sc(3))
	if sel {
		lift = max(4, c.sc(8))
	}
	for k := n - 1; k >= 0; k-- {
		im := get(mint, t.Money, t.Types[i], st[k].Level, st[k].Seed, l.piece)
		if im == nil {
			continue
		}
		x, y := cp.X-half.X+offs[k].X, cp.Y-half.Y+offs[k].Y
		d, lf := sh, 0
		if k == 0 {
			x, y, d, lf = x-lift, y-lift, sh+lift, lift
		}
		c.cast(im, x, y, d, lf)
	}
}

// cast lays a piece over a hard shadow, long when the piece is lifted.
func (c *canvas) cast(im *image.RGBA, x, y, d, lift int) {
	c.tint = color.RGBA{}
	if lift > 0 {
		c.tint = c.dark
	}
	for e := d; e <= d+lift; e += max(1, lift/2) {
		c.lay(im, x, y, e, true)
	}
	c.lay(im, x, y, 0, false)
}

// get is the mint's piece, or nil where there is no mint.
func get(m *coins.Mint, age, typ string, level int, seed uint32, size int) *image.RGBA {
	if m == nil {
		return nil
	}
	return m.Piece(age, typ, level, seed, size)
}

// legend paints the round central well: the level 7 piece when held, its
// outline when not.
func (c *canvas) legend(l TrayLayout, t Tray, mint *coins.Mint, sel bool, rp []color.RGBA) {
	cp := l.ctr[Wells]
	c.pit(cp.X, cp.Y, l.dia, rp)
	im := get(mint, t.Money, "feat", coins.Legendary, 0, l.dia*6/7)
	switch {
	case im == nil:
	case !t.Legend:
		c.outline(cp.X, cp.Y, im)
	case sel:
		d := max(4, c.sc(8))
		c.cast(im, cp.X-d, cp.Y-d, d+max(2, c.sc(3)), d)
	default:
		c.cast(im, cp.X, cp.Y, max(2, c.sc(3)), 0)
	}
}

// lining is the ramp and the kind of surface of a drawer's tray.
func lining(slot int) ([]color.RGBA, int) {
	switch {
	case slot >= 15 && slot < 22:
		return foam, 2
	case slot > 12:
		return felt, 0
	case slot > 7:
		return velvet, 0
	}
	return baize, 0
}

// tray paints what lies in the canvas: the frame, the lining, the wells and
// the plates. sel is the selected well, Wells for the legendary, -1 for none.
func (c *canvas) tray(l TrayLayout, t Tray, mint *coins.Mint, sel int) {
	whole, lin := c.img.Rect.Add(c.org), l.lining
	w, h := l.Cells.X*l.Cell.X, l.Cells.Y*l.Cell.Y
	rp, kind := lining(t.Slot)
	c.dark = mul(rp[1], 0.55)
	for _, r := range []image.Rectangle{image.Rect(0, 0, w, lin.Min.Y), image.Rect(0, lin.Max.Y, w, h),
		image.Rect(0, lin.Min.Y, lin.Min.X, lin.Max.Y), image.Rect(lin.Max.X, lin.Min.Y, w, lin.Max.Y)} {
		par(r.Intersect(whole), func(b image.Rectangle) {
			if kind == 2 {
				c.surface(b, lin, alu, 1, 9+t.Slot)
			} else {
				c.board(b, 700+t.Slot, 0, false)
			}
		})
	}
	par(lin.Intersect(whole), func(b image.Rectangle) { c.surface(b, lin, rp, kind, 7) })
	groove, st := max(2, c.sc(4)), strip*l.Cell.Y
	c.fill(image.Rect(0, st-groove, w, st), black)
	c.fill(image.Rect(0, h-st, w, h-st+groove), black)
	c.edge(image.Rect(0, 0, w, st-groove), false)
	c.edge(image.Rect(0, h-st+groove, w, h), false)
	c.edge(lin.Inset(-1), true)
	for i := 0; i < Wells; i++ {
		if l.Wells[i].Overlaps(whole) {
			c.well(l, t, i, mint, sel == i, rp)
		}
	}
	if l.Legend.Overlaps(whole) {
		c.legend(l, t, mint, sel == Wells, rp)
	}
	seed := t.Slot*20 + 1
	for k, r := range []image.Rectangle{l.Name, l.Count, l.Hint, l.Above, l.Below} {
		if (k == 3 && t.Slot == 0) || (k == 4 && t.Slot == Slots-1) {
			continue
		}
		c.plate(l.pix(r), false, seed+k)
	}
	c.plate(l.pix(l.Ticket), false, seed+12)
	for i, r := range l.Labels {
		c.plate(l.pix(r), i < Wells && len(t.Stacks[i]) == 0 || i == Wells && !t.Legend, seed+5+i)
	}
}

// Open paints the whole panel for an open drawer, nothing selected.
func Open(l TrayLayout, t Tray, mint *coins.Mint) *image.RGBA {
	c := newCanvas(image.Rect(0, 0, l.Cells.X*l.Cell.X, l.Cells.Y*l.Cell.Y), l.s)
	c.tray(l, t, mint, -1)
	return c.img
}

// Lift paints well i selected (Wells for the legendary): the top piece raised
// over a hard shadow, a brass line round the lip. The picture covers the
// well's cells and one more cell each side, to lay over Open's picture.
func Lift(l TrayLayout, t Tray, i int, mint *coins.Mint) (cells image.Rectangle, img *image.RGBA) {
	r := l.Legend
	if i >= 0 && i < Wells {
		r = l.Wells[i]
	}
	cells = image.Rect(r.Min.X/l.Cell.X-1, r.Min.Y/l.Cell.Y-1, (r.Max.X+l.Cell.X-1)/l.Cell.X+1, (r.Max.Y+l.Cell.Y-1)/l.Cell.Y+1)
	c := newCanvas(l.pix(cells), l.s)
	c.tray(l, t, mint, i)
	return cells, c.img
}

// pit paints the legendary's round well: a floor a few steps darker than the
// lining and dithered lighter towards the rim, a hard shadow along the upper
// edge and a thin gold lip.
func (c *canvas) pit(cx, cy, dia int, rp []color.RGBA) {
	r, cut := float64(dia)/2, float64(max(3, c.sc(10)))
	for y := -dia/2 - 1; y <= dia/2+1; y++ {
		for x := -dia/2 - 1; x <= dia/2+1; x++ {
			d := math.Hypot(float64(x), float64(y))
			col := mul(rp[1], 0.5)
			switch {
			case d > r:
				continue
			case d > r-2:
				col = brass[4]
			case float64(-x-y)/math.Sqrt2 > 0 && r-d < cut:
				col = mul(rp[0], 0.35)
			case (d-r*0.75)/(r*0.25) > bayer[(cy+y)&3][(cx+x)&3]:
				col = mul(rp[1], 0.8)
			}
			c.px(cx+x, cy+y, col)
		}
	}
}
