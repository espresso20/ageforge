package mapstyle

import (
	"slices"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// Draw renders the full view: header, picture, side panel, news, inspector and
// status line. Below 60 by 16 it draws
// the compact view instead.
func (v *view) Draw(scr tcell.Screen, r Rect, f Frame) {
	if r.W < 60 || r.H < 16 {
		v.DrawCompact(scr, r, f)
		return
	}
	ink := v.begin(f)
	cv := NewCanvas(scr, r, f.Tier, tcell.StyleDefault)
	if ink == nil {
		drawNoMap(cv)
		return
	}
	cv.Fill(0, 0, r.W, r.H, ' ', ink.Style(mapmodel.CText))
	v.overview = v.wantsOverview(r)
	sc := v.scene(f)
	side := 0
	if r.W >= 110 && (v.legend || v.flows) {
		side = sideWidth
	}
	mapW, mapH := r.W-side, r.H-5
	v.drawPicture(cv, ink.Palette, 0, 1, mapW, mapH, sc, f)
	zoom := v.zoomLabel()
	v.drawHeader(cv, r.W, f.Model, zoom)
	switch {
	case side > 0:
		if y := v.drawSidebar(cv, f.Model, mapW, 1, side, mapH); v.legend && y < 1+mapH {
			v.drawLegend(cv, ink.Palette, mapW+2, y, side-3, 1+mapH-y, f.Model)
		}
	case v.flows:
		boxW := min(sideWidth, mapW/2)
		v.drawFlowsBox(cv, f.Model, mapW-boxW, 1, boxW, mapH)
	}
	in, ok := v.Inspect(f)
	v.drawBottom(cv, r.W, r.H, f.Model, in, ok, zoom)
}

// drawPicture draws the world in the box (x, y, w, h) of the canvas: the whole
// town squeezed to fit, or the close-up round the cursor.
func (v *view) drawPicture(cv *Canvas, mp *Palette, x, y, w, h int, sc *mapgen.Scene, f Frame) {
	anim := f.Anim
	var cx, cy int
	var on bool
	if v.overview {
		switch z := v.shownZoom(); {
		case v.stars && z == zCosmos:
			v.ov = sc.OverviewVoid(w, h)
		case v.stars && z == zOverview:
			v.ov = sc.OverviewArm(w, h)
		case z == zCosmos:
			v.ov = sc.OverviewCosmos(w, h)
		case v.stars && z == zInner:
			v.ov = sc.OverviewNeighborhood(w, h)
		case v.stars && z == zCloseUp:
			v.ov = sc.OverviewSol(w, h)
		case z == zInner:
			v.ov = sc.SystemOverview(w, h, mapgen.SystemViewInner)
		default:
			v.ov = sc.SystemOverview(w, h, mapgen.SystemViewFull)
		}
		if v.seek != (mapgen.Thing{}) { // zoomed between bodies: the cursor stays on it
			bs := bodiesOf(v.ov.Grid)
			for _, want := range [2]mapgen.Thing{v.seek, sunThing} { // else the sun, if the body is out of the frame
				if i := slices.IndexFunc(bs, func(b body) bool { return b.th == want }); i >= 0 {
					v.cur = v.ov.Pick(bs[i].at.X, bs[i].at.Y)
					break
				}
			}
			v.seek = mapgen.Thing{}
		}
		if z := v.shownZoom(); f.Stage != nil && v.imaged(z) {
			v.pal = cellPaletteFor(v.pal, v.ov.Grid)
			v.placeVoid(cv, f.Stage, f.Slot, x, y, w, h, sc.Seed)
		} else {
			if f.Stage != nil {
				f.Stage.Clear(f.Slot)
			}
			v.blit(cv, x, y, v.ov.Grid)
		}
		cx, cy, on = v.ov.CellOf(v.cur)
	} else {
		if f.Stage != nil {
			f.Stage.Clear(f.Slot)
		}
		v.ov = nil
		v.fit(w, h, sc)
		v.win = mapgen.Rect{X: v.cam.X, Y: v.cam.Y, W: w, H: h}
		v.blit(cv, x, y, windowFilled(sc, v.cam.X, v.cam.Y, w, h))
		cx, cy = v.cur.X-v.cam.X, v.cur.Y-v.cam.Y
		on = cx >= 0 && cy >= 0 && cx < w && cy < h
	}
	if !on || !v.inspect {
		return
	}
	if g, gx, gy, ok := v.shownCell(sc); ok && sc.Grid.Space {
		if b, ok := v.bodyAt(g, gx, gy); ok {
			// the corners stand outside the body, so the body shows whole
			dx, dy := x, y // the grid's cell to the canvas's
			if !v.overview {
				dx, dy = x-v.cam.X, y-v.cam.Y
			}
			v.putBrackets(cv, mp, mapgen.Rect{X: b.box.X + dx - 1, Y: b.box.Y + dy - 1, W: b.box.W + 2, H: b.box.H + 2}, mapgen.Rect{X: x, Y: y, W: w, H: h}, anim)
			return
		}
	}
	v.putCursor(cv, mp, x+cx, y+cy, anim)
}

// putBrackets draws the four corners of box on the canvas in the cursor's
// colour, those that fall inside the picture clip.
func (v *view) putBrackets(cv *Canvas, mp *Palette, box, clip mapgen.Rect, anim int) {
	for _, c := range [4]struct {
		x, y int
		r    rune
	}{{box.X, box.Y, '┌'}, {box.X + box.W - 1, box.Y, '┐'}, {box.X, box.Y + box.H - 1, '└'}, {box.X + box.W - 1, box.Y + box.H - 1, '┘'}} {
		if c.x < clip.X || c.y < clip.Y || c.x >= clip.X+clip.W || c.y >= clip.Y+clip.H {
			continue
		}
		_, under := cv.Get(c.x, c.y)
		ground := v.groundOf(under)
		_, fill, _ := mp.CursorOver(ground, anim).Decompose()
		cv.Put(c.x, c.y, c.r, tcell.StyleDefault.Foreground(fill).Background(ground).Bold(true))
	}
}

// putCursor highlights the cell (x, y) of the canvas, keeping its glyph.
func (v *view) putCursor(cv *Canvas, mp *Palette, x, y, anim int) {
	r, under := cv.Get(x, y)
	cv.Put(x, y, r, mp.CursorOver(v.groundOf(under), anim))
}

// groundOf is the colour a cell stands on; over an image it is the palette's
// ground, which the image is drawn on.
func (v *view) groundOf(st tcell.Style) tcell.Color {
	_, ground, _ := st.Decompose()
	if ground == theme.Transparent && v.pal != nil {
		return v.pal.bg
	}
	return ground
}

// fit places the close-up's window: it stays put and scrolls only when the
// cursor nears an edge, and jumps only when the cursor is off screen.
func (v *view) fit(w, h int, sc *mapgen.Scene) {
	gw, gh := mapgen.WorldW, mapgen.WorldH
	if sc != nil && sc.Grid != nil {
		gw, gh = sc.Grid.W, sc.Grid.H
	}
	key := [2]int{w, h}
	if key != v.camKey || v.recentre || v.cur.X < v.cam.X || v.cur.X >= v.cam.X+w || v.cur.Y < v.cam.Y || v.cur.Y >= v.cam.Y+h {
		v.camKey, v.recentre = key, false
		v.cam = mapgen.Pt{X: v.cur.X - w/2, Y: v.cur.Y - h/2}
	}
	mx, my := min(8, w/4), min(4, h/4)
	v.cam.X = clamp(v.cam.X, v.cur.X-w+mx+1, v.cur.X-mx)
	v.cam.Y = clamp(v.cam.Y, v.cur.Y-h+my+1, v.cur.Y-my)
	v.cam.X = clamp(v.cam.X, 0, max(0, gw-w))
	v.cam.Y = clamp(v.cam.Y, 0, max(0, gh-h))
	// A window bigger than the world centres it; windowFilled paints the
	// rest.
	if w > gw {
		v.cam.X = -(w - gw) / 2
	}
	if h > gh {
		v.cam.Y = -(h - gh) / 2
	}
}

// windowFilled is sc.Grid.Window for a window that may reach past the world:
// in space scales, procedural space cells fill beyond the world edge without mirroring;
// in ground scales, land is mirrored to preserve coastlines, but unique landmarks,
// homes, and buildings are never duplicated.
func windowFilled(sc *mapgen.Scene, x0, y0, w, h int) *mapgen.Grid {
	g := sc.Grid
	out := g.Window(x0, y0, w, h)
	if x0 >= 0 && y0 >= 0 && x0+w <= g.W && y0+h <= g.H {
		return out
	}
	outCell := sc.SpaceCell
	if outCell == nil {
		outCell = sc.GroundCell
	}
	if outCell != nil {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				wx, wy := x0+x, y0+y
				if wx >= 0 && wy >= 0 && wx < g.W && wy < g.H {
					continue
				}
				out.Cells[y*w+x] = outCell(wx, wy)
				out.Hit[y*w+x] = 0
			}
		}
		return out
	}
	mirror := func(v, n int) int {
		v = ((v % (2 * n)) + 2*n) % (2 * n)
		if v >= n {
			v = 2*n - 1 - v
		}
		return v
	}
	terrain := func(i int) bool {
		th := g.Things[g.Hit[i]]
		if th.Key == "home" {
			return false
		}
		return th.Kind == mapgen.KindNone || th.Kind == mapgen.KindTerrain
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			wx, wy := x0+x, y0+y
			if wx >= 0 && wy >= 0 && wx < g.W && wy < g.H {
				continue
			}
			sx, sy := mirror(wx, g.W), mirror(wy, g.H)
			i := sy*g.W + sx
			// Walk inward off a building or road to the ground under it.
			dx := 1
			if sx >= g.W/2 {
				dx = -1
			}
			for n := 0; n < 64 && !terrain(i) && sx+dx >= 0 && sx+dx < g.W; n++ {
				sx += dx
				i = sy*g.W + sx
			}
			if !terrain(i) {
				dx = -dx
				for n := 0; n < 64 && !terrain(i) && sx+dx >= 0 && sx+dx < g.W; n++ {
					sx += dx
					i = sy*g.W + sx
				}
			}
			if terrain(i) {
				out.Cells[y*w+x] = g.Cells[i]
				out.Hit[y*w+x] = g.Hit[i]
			} else {
				out.Cells[y*w+x] = mapgen.Cell{R: ' '}
				out.Hit[y*w+x] = 0
			}
		}
	}
	return out
}

func clamp(v, lo, hi int) int { return min(max(v, lo), hi) }

// DrawCompact renders the glanceable mini view: a header line, the whole town
// squeezed into the space, and one news line.
func (v *view) DrawCompact(scr tcell.Screen, r Rect, f Frame) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	ink := v.begin(f)
	cv := NewCanvas(scr, r, f.Tier, tcell.StyleDefault)
	if ink == nil {
		drawNoMap(cv)
		return
	}
	cv.Fill(0, 0, r.W, r.H, ' ', ink.Style(mapmodel.CText))
	v.drawCompactBars(cv, r, f.Model)
	mh := r.H - 2
	if mh <= 0 {
		return
	}
	v.overview = true
	ov := v.scene(f).Overview(r.W, mh)
	v.blit(cv, 0, 1, ov.Grid)
	if cx, cy, ok := ov.CellOf(mapgen.Pt{X: v.mark.X, Y: v.mark.Y}); v.marked && ok {
		v.putCursor(cv, ink.Palette, cx, 1+cy, f.Anim)
	}
}
