package procmap

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// Draw renders the full view: header, picture, side panel, news, inspector and
// status line. Below 60 by 16 it draws
// the compact view instead.
func (v *view) Draw(scr tcell.Screen, r mapstyle.Rect, f mapstyle.Frame) {
	if r.W < 60 || r.H < 16 {
		v.DrawCompact(scr, r, f)
		return
	}
	ink := v.begin(f)
	cv := mapstyle.NewCanvas(scr, r, f.Tier, tcell.StyleDefault)
	if ink == nil {
		drawNoMap(cv)
		return
	}
	cv.Fill(0, 0, r.W, r.H, ' ', ink.Style(mapmodel.CText))
	sc := v.scene(f)
	side := 0
	if r.W >= 110 && (v.legend || v.flows) {
		side = sideWidth
	}
	mapW, mapH := r.W-side, r.H-5
	v.overview = v.wantsOverview(r)
	zoom := zoomNames[zCloseUp].String()
	if v.overview {
		zoom = zoomNames[zOverview].String()
	}
	v.drawPicture(cv, ink.Palette, 0, 1, mapW, mapH, sc, f.Anim)
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
func (v *view) drawPicture(cv *mapstyle.Canvas, mp *mapstyle.Palette, x, y, w, h int, sc *mapgen.Scene, anim int) {
	var cx, cy int
	var on bool
	if v.overview {
		v.ov = sc.Overview(w, h)
		v.blit(cv, x, y, v.ov.Grid)
		cx, cy, on = v.ov.CellOf(v.cur)
	} else {
		v.ov = nil
		v.fit(w, h)
		v.blit(cv, x, y, sc.Grid.Window(v.cam.X, v.cam.Y, w, h))
		cx, cy = v.cur.X-v.cam.X, v.cur.Y-v.cam.Y
		on = cx >= 0 && cy >= 0 && cx < w && cy < h
	}
	if on && v.inspect {
		v.putCursor(cv, mp, x+cx, y+cy, anim)
	}
}

// putCursor highlights the cell (x, y) of the canvas, keeping its glyph.
func (v *view) putCursor(cv *mapstyle.Canvas, mp *mapstyle.Palette, x, y, anim int) {
	r, under := cv.Get(x, y)
	_, ground, _ := under.Decompose()
	cv.Put(x, y, r, mp.CursorOver(ground, anim))
}

// fit places the close-up's window: it stays put and scrolls only when the
// cursor nears an edge, and jumps only when the cursor is off screen.
func (v *view) fit(w, h int) {
	key := [2]int{w, h}
	if key != v.camKey || v.recentre || v.cur.X < v.cam.X || v.cur.X >= v.cam.X+w || v.cur.Y < v.cam.Y || v.cur.Y >= v.cam.Y+h {
		v.camKey, v.recentre = key, false
		v.cam = mapgen.Pt{X: v.cur.X - w/2, Y: v.cur.Y - h/2}
	}
	mx, my := min(8, w/4), min(4, h/4)
	v.cam.X = clamp(v.cam.X, v.cur.X-w+mx+1, v.cur.X-mx)
	v.cam.Y = clamp(v.cam.Y, v.cur.Y-h+my+1, v.cur.Y-my)
	v.cam.X = clamp(v.cam.X, 0, max(0, mapgen.WorldW-w))
	v.cam.Y = clamp(v.cam.Y, 0, max(0, mapgen.WorldH-h))
}

func clamp(v, lo, hi int) int { return min(max(v, lo), hi) }

// DrawCompact renders the glanceable mini view: a header line, the whole town
// squeezed into the space, and one news line.
func (v *view) DrawCompact(scr tcell.Screen, r mapstyle.Rect, f mapstyle.Frame) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	ink := v.begin(f)
	cv := mapstyle.NewCanvas(scr, r, f.Tier, tcell.StyleDefault)
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
	ov := v.scene(f).Overview(r.W, mh)
	v.blit(cv, 0, 1, ov.Grid)
	if cx, cy, ok := ov.CellOf(mapgen.Pt{X: v.mark.X, Y: v.mark.Y}); v.marked && ok {
		v.putCursor(cv, ink.Palette, cx, 1+cy, f.Anim)
	}
}
