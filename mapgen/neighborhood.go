package mapgen

import (
	"cmp"
	"slices"
)

// The Galactic map opens on Sol's neighbourhood: the nearest real stars at
// their true distances from Sol, the people's colonies lit round some of them
// and joined by steady warp lanes. One scale serves every star, so a distance
// reads true; a star's direction is its galactic longitude. The close-up of
// Sol is the solar-system drawing of the Interstellar Age (overviewSystem).

// NearStar is a star of the neighbourhood.
type NearStar struct {
	Key     string // the name's key, for the inspect line
	Dist100 int    // light-years from Sol, times 100
	lon     int    // galactic longitude, degrees
	look    starLook
}

// starLook is how a star is drawn: its colour, how big it shows (0 a single
// pixel, 1 a small disc, 2 a bright disc with a glow), how bright a dim one
// is, and the companion beside it, if any.
type starLook struct {
	hue    Hue
	size   int
	weight int
	pair   Hue // a second star beside it
	extra  Hue // a third, dim one, or a white dwarf
}

// NearStars are the stars of the neighbourhood, nearest first. Distances are
// the catalogue's; longitudes are rounded to a degree.
var NearStars = [...]NearStar{
	{"alpha_centauri", 437, 316, starLook{HSun, 1, 255, HAmber, HRed}},
	{"barnards_star", 596, 31, starLook{HRed, 0, 190, 0, 0}},
	{"wolf_359", 786, 244, starLook{HRed, 0, 150, 0, 0}},
	{"lalande_21185", 831, 185, starLook{HRed, 0, 190, 0, 0}},
	{"sirius", 860, 227, starLook{HCool, 2, 255, 0, HWhite}},
	{"luyten_726_8", 870, 175, starLook{HRed, 0, 170, HRed, 0}},
	{"ross_154", 970, 11, starLook{HRed, 0, 160, 0, 0}},
	{"ross_248", 1030, 110, starLook{HRed, 0, 140, 0, 0}},
	{"epsilon_eridani", 1050, 196, starLook{HAmber, 1, 220, 0, 0}},
	{"lacaille_9352", 1070, 5, starLook{HRed, 0, 200, 0, 0}},
	{"ross_128", 1100, 270, starLook{HRed, 0, 150, 0, 0}},
	{"61_cygni", 1140, 82, starLook{HAmber, 1, 190, HAmber, 0}},
	{"procyon", 1150, 214, starLook{HSun, 2, 230, 0, HWhite}},
	{"struve_2398", 1150, 89, starLook{HRed, 0, 170, HRed, 0}},
	{"groombridge_34", 1160, 117, starLook{HRed, 0, 160, HRed, 0}},
	{"epsilon_indi", 1190, 336, starLook{HAmber, 1, 190, 0, 0}},
	{"tau_ceti", 1190, 173, starLook{HSun, 1, 210, 0, 0}},
	{"teegardens_star", 1250, 160, starLook{HRed, 0, 130, 0, 0}},
	{"kapteyns_star", 1280, 251, starLook{HRed, 0, 150, 0, 0}},
}

// ringLy is the gap between the faint distance rings, in light-years.
const ringLy = 5

// Colony is a building line with something standing, and so a lit star of the
// neighbourhood.
type Colony struct {
	Line    string
	Count   int // buildings of the line
	Systems int // systems of its sector that shine
	Tier    int // 0..3, the rank of the best building standing
}

// colonyStars says which star each colony lights: the colonies keep a star for
// good, the earlier lines the nearer ones, with a seeded shuffle among stars
// of a like distance.
func colonyStars(seed int64, cs []Colony) []int {
	order := make([]int, len(NearStars))
	for i := range order {
		order[i] = i
	}
	jitter := func(i int) int { return NearStars[i].Dist100 + roll(300, seed, 6100, int64(i)) }
	slices.SortStableFunc(order, func(a, b int) int { return cmp.Compare(jitter(a), jitter(b)) })
	out := make([]int, len(cs))
	for i, c := range cs {
		j := lineOf[c.Line]
		rank := 0 // where this line stands among all the lines, with its own jitter
		for k := 0; k < numLines; k++ {
			if k*100+roll(250, seed, 6101, int64(k)) < j*100+roll(250, seed, 6101, int64(j)) {
				rank++
			}
		}
		out[i] = order[rank]
	}
	return out
}

// OverviewNeighborhood draws Sol's neighbourhood into w by h cells: Sol at
// the middle of the window, which is the middle of the world.
func (s *Scene) OverviewNeighborhood(w, h int) *Overview {
	gw, gh := s.Grid.W, s.Grid.H
	o := &Overview{Grid: newGrid(w, h), Scale: 1, step: 256, picks: make([]Pt, w*h), X0: gw/2 - w/2, Y0: gh/2 - h/2}
	o.Grid.Wall, o.Grid.Tint, o.Grid.Space = s.Grid.Wall, s.Grid.Tint, true
	o.deepSky(s)

	pix := NewPixels(w, h)
	defer pix.Release()
	cx, cy := (2*(w/2)+1)*pxUnit, (2*(h/2)+1)*pxUnit
	reach := min(w*pxUnit, 2*h*pxUnit) * 90 / 100 // the farthest star, in 16ths of a pixel
	far := NearStars[len(NearStars)-1].Dist100
	at := func(d100, lon int) (int, int) {
		r, a := d100*reach/far, lon*turn/360
		return cx + r*cos4096(a)/1024, cy - r*sin4096(a)/2048
	}

	for ly := ringLy; ly*100 <= far; ly += ringLy { // the distance rings
		rx := ly * 100 * reach / far
		pix.Ellipse(cx, cy, rx, rx/2, P(HOrbit, 90), 0, 2, 0)
	}

	cs := s.Colonies
	slot := colonyStars(s.Seed, cs)
	colony := map[int]Colony{}
	for i, c := range cs {
		colony[slot[i]] = c
	}
	var pts [len(NearStars)][2]int
	for i, st := range NearStars {
		pts[i][0], pts[i][1] = at(st.Dist100, st.lon)
	}
	s.paintNearLanes(pix, cx, cy, &pts, colony)

	for i, st := range NearStars {
		th := Thing{Kind: KindTerrain, Key: "neighbor", Ref: i + 1}
		c, lit := colony[i]
		if lit {
			th.Lineage, th.Count = c.Line, c.Systems
		}
		own := pix.Owner(th, 5)
		x, y := pts[i][0], pts[i][1]
		if lit {
			for k, n := 0, 12; k < n; k++ { // a thin ring, a little brighter for a bigger colony
				a := turn/4 + k*turn/n
				pix.Set(floorDiv(x+4*pxUnit*cos4096(a)/1024, pxUnit), floorDiv(y+4*pxUnit*sin4096(a)/2048, pxUnit), P(HAmber, 90+25*c.Tier), own)
			}
		}
		paintNearStar(pix, x, y, st.look, own, blinkWith(s.Seed, s.Anim, 700+i, 40, 30)) // the brighter ones twinkle, slowly
	}
	sol, sx, sy := pix.Owner(sunThing, 6), cx/pxUnit, cy/pxUnit
	for dx := -2; dx <= 2; dx++ {
		pix.Set(sx+dx, sy, P([5]Hue{HSun, HWhite, HWhite, HWhite, HSun}[dx+2], 245), sol)
	}
	pix.Set(sx, sy-1, P(HSun, 210), sol)
	pix.Set(sx, sy+1, P(HSun, 210), sol)
	pix.Resolve(o.Grid)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o.picks[y*w+x] = Pt{o.X0 + x, o.Y0 + y}
		}
	}
	return o
}

// paintNearStar draws one star, with its companions, at (x, y) in 16ths: a
// dim red dwarf is one pixel, the brighter stars a short bar, the brightest a
// glint.
func paintNearStar(pix *Pixels, x, y int, l starLook, own uint8, arms bool) {
	px, py := floorDiv(x, pxUnit), floorDiv(y, pxUnit)
	pix.Set(px, py, P(l.hue, l.weight), own)
	if l.size > 0 && arms {
		pix.Set(px+1, py, P(l.hue, l.weight), own)
	}
	if l.size > 1 && arms {
		pix.Set(px-1, py, P(l.hue, 200), own)
		pix.Set(px, py-1, P(l.hue, 150), own)
		pix.Set(px, py+1, P(l.hue, 150), own)
	}
	if l.pair != 0 {
		pix.Set(px+3, py, P(l.pair, max(l.weight-30, 120)), own)
	}
	if l.extra != 0 {
		pix.Set(px-2, py+2, P(l.extra, 150), own)
	}
}

// paintNearLanes joins the colonies to Sol and to each other: every colony to
// the nearest of Sol and the colonies already joined, so the web holds
// together. A lane is a steady glow; busier ones are brighter.
func (s *Scene) paintNearLanes(pix *Pixels, cx, cy int, pts *[len(NearStars)][2]int, colony map[int]Colony) {
	type node struct {
		x, y  int
		line  string
		count int
	}
	nodes := []node{{cx, cy, "home", 0}}
	var lit []int
	for i := range NearStars { // nearest first
		if _, ok := colony[i]; ok {
			lit = append(lit, i)
		}
	}
	for _, i := range lit {
		c := colony[i]
		best, gap := 0, 1<<60
		for j, n := range nodes {
			if g := sq(n.x-pts[i][0]) + 4*sq(n.y-pts[i][1]); g < gap {
				best, gap = j, g
			}
		}
		from := nodes[best]
		m := c.Count // the lane is as busy as the smaller colony it joins
		if from.count > 0 {
			m = min(m, from.count)
		}
		traffic := 1
		if m >= 40 {
			traffic = 3
		} else if m >= 8 {
			traffic = 2
		}
		own := pix.Owner(Thing{Kind: KindLane, Key: c.Line, Lineage: from.line, Count: traffic}, 2)
		x0, y0, x1, y1 := floorDiv(from.x, pxUnit), floorDiv(from.y, pxUnit), floorDiv(pts[i][0], pxUnit), floorDiv(pts[i][1], pxUnit)
		n := max(abs(x1-x0), abs(y1-y0))
		for k := 1; k < n; k++ {
			pix.Set(x0+(x1-x0)*k/n, y0+(y1-y0)*k/n, P(HCyan, [4]int{0, 40, 60, 90}[traffic]), own)
		}
		nodes = append(nodes, node{pts[i][0], pts[i][1], c.Line, c.Count})
	}
}

// deepSky paints the near-black ground of the close views of space: sparse
// faint stars and, at the edges, a tint or two.
func (o *Overview) deepSky(s *Scene) {
	for y := 0; y < o.Grid.H; y++ {
		for x := 0; x < o.Grid.W; x++ {
			o.Grid.paint(x, y, s.deep.cell(s.Seed, o.X0+x, o.Y0+y, x, y, o.Grid.W, o.Grid.H))
		}
	}
}

// deepLayer keeps the dark ground between frames: it depends on the seed and
// the window alone.
type deepLayer struct {
	seed       int64
	w, h, x, y int
	cells      []Cell
}

// cell is deepCell of the cached window, computed once per seed and window.
func (d *deepLayer) cell(seed int64, wx, wy, x, y, w, h int) Cell {
	x0, y0 := wx-x, wy-y
	if d.cells == nil || d.seed != seed || d.w != w || d.h != h || d.x != x0 || d.y != y0 {
		d.seed, d.w, d.h, d.x, d.y = seed, w, h, x0, y0
		d.cells = make([]Cell, w*h)
		for cy := 0; cy < h; cy++ {
			for cx := 0; cx < w; cx++ {
				d.cells[cy*w+cx] = deepCell(seed, x0+cx, y0+cy, cx, cy, w, h)
			}
		}
	}
	return d.cells[y*w+x]
}

// deepCell is a cell of deep space at world (wx, wy), shown at (x, y) of a
// window w by h.
func deepCell(seed int64, wx, wy, x, y, w, h int) Cell {
	c := Cell{R: ' ', Bg: P(HVoid, 255)}
	var best int
	for k, hue := range [2]Hue{HNebula, HAir} { // each tint pools round a spot on an edge
		ax, ay := roll(w, seed, 6200, int64(k)), 0
		if roll(2, seed, 6201, int64(k)) == 0 {
			ay = h - 1
		}
		d := isqrt(sq(x-ax) + 4*sq(y-ay))
		if t := 16 - d*16/max(8, w/3); t > best {
			best, c.Bg = t, P(hue, t)
		}
	}
	if chance(3, seed, 6202, int64(wx), int64(wy)) {
		c.R, c.Fg = '·', P(HStar, 40+roll(70, seed, 6203, int64(wx), int64(wy)))
	}
	return c
}
