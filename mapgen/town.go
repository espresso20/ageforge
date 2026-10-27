package mapgen

import (
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
)

// wonderSpot is a wonder on its plot.
type wonderSpot struct {
	plot Lot
	w    mapmodel.Wonder
}

// wonders lists the model's wonders that have a plot.
func (f *frame) wonders() []wonderSpot {
	var out []wonderSpot
	for _, w := range f.m.Wonders {
		if w.Age < len(f.plan.Plots) && f.plan.Plots[w.Age].Rect.W > 0 {
			out = append(out, wonderSpot{f.plan.Plots[w.Age], w})
		}
	}
	return out
}

// visible says how many cells of each road are laid: as far as the farthest
// parcel in use along it, and as far as any branch that is laid leaves it, so
// the roads grow with the town and never lead nowhere.
func (p *Plan) visible(inUse []Lot) []int {
	shown := make([]int, len(p.Roads))
	shown[0] = 1
	if p.Orbit {
		shown[1] = len(p.Roads[1].Cells) // the elevator always stands
	}
	for _, l := range inUse {
		shown[l.Road] = max(shown[l.Road], l.At+1)
	}
	for i := len(p.Roads) - 1; i > 0; i-- {
		if shown[i] > 0 {
			parent := p.Roads[i].Parent
			shown[parent] = max(shown[parent], p.Roads[i].From+1)
		}
	}
	return shown
}

// fillerLots are the parcels no building stands on that the people fill with
// ordinary houses, nearest the middle first: one for every FillPer people.
func (f *frame) fillerLots() []int {
	if len(f.st.Filler) == 0 {
		return nil
	}
	taken := map[int]bool{}
	for _, pl := range f.placed {
		taken[pl.lot] = true
	}
	var out []int
	for _, lot := range f.plan.Near {
		if len(out) >= f.m.Workers.Pop/f.st.FillPer {
			break
		}
		if !taken[lot] {
			out = append(out, lot)
		}
	}
	return out
}

// layRoads works out which road cells are laid, before anything is drawn.
func (f *frame) layRoads() {
	var inUse []Lot
	for _, pl := range f.placed {
		inUse = append(inUse, f.plan.Lots[pl.lot])
	}
	for _, lot := range f.filler {
		inUse = append(inUse, f.plan.Lots[lot])
	}
	if f.st.Pads > 0 && f.plan.Port.Rect.W > 0 {
		inUse = append(inUse, f.plan.Port)
	}
	for _, w := range f.wonders() {
		inUse = append(inUse, w.plot)
	}
	f.shown = f.plan.visible(inUse)
	for i, r := range f.plan.Roads {
		for _, c := range r.Cells[:f.shown[i]] {
			f.roadAt[at(c.X, c.Y)] = int8(r.Tier + 1)
			f.busy[at(c.X, c.Y)] = true
		}
	}
}

func (f *frame) laid(x, y int) bool { return inWorld(x, y) && f.roadAt[at(x, y)] > 0 }

// paintTown draws the made things: yards, roads, buildings, wonders.
func (f *frame) paintTown() {
	for _, pl := range f.placed {
		f.paintYard(f.plan.Lots[pl.lot].Rect, pl)
	}
	for _, lot := range f.filler {
		f.paintYardOf(f.plan.Lots[lot].Rect, f.st.Yard)
	}
	for _, w := range f.wonders() {
		w.plot.Rect.cells(func(x, y int) {
			f.g.paint(x, y, f.st.Plaza.at(x, y, 1))
			f.busy[at(x, y)] = true
		})
	}
	f.paintRoads()
	f.paintPort()
	for _, lot := range f.filler {
		f.paintHouse(lot)
	}
	for _, pl := range f.placed {
		if !pl.ruin {
			f.paintBuilding(pl)
		}
	}
	for _, w := range f.wonders() {
		f.paintWonder(w)
	}
	f.paintRelics()
}

func (f *frame) paintYard(r Rect, pl placed) {
	look := f.st.Yard
	switch {
	case pl.ruin:
		look = f.st.Rubble
	case pl.b.Lineage == mapmodel.LinFood && f.st.Crops.Glyphs != "":
		look = f.st.Crops
	}
	f.paintYardOf(r, look)
	if pl.ruin {
		r.cells(func(x, y int) {
			f.g.mark(x, y, Thing{Kind: KindBuilding, Key: pl.b.Key, Lineage: pl.b.Lineage, Count: pl.b.Ruins, Ruin: true})
		})
	}
}

// paintYardOf lays the ground of a parcel in a look.
func (f *frame) paintYardOf(r Rect, look Look) {
	r.cells(func(x, y int) {
		f.g.paint(x, y, look.at(x, y, 2))
		f.busy[at(x, y)] = true
	})
}

func (f *frame) paintRoads() {
	for i, r := range f.plan.Roads {
		for _, c := range r.Cells[:f.shown[i]] {
			look := f.st.Roads[r.Tier]
			mask := neighbourMask(c.X, c.Y, f.laid)
			cell := Cell{R: joinRune(look.Joins, mask, hash(int64(c.X), int64(c.Y))), Fg: look.Fg, Bg: look.Bg}
			if f.plan.Bridge[c] {
				cell.Bg = f.st.Deck
			}
			f.g.paint(c.X, c.Y, cell)
			key := roadNames[r.Tier]
			switch {
			case i == 0:
				key = "square"
			case f.plan.Bridge[c]:
				key = "bridge"
			}
			f.g.mark(c.X, c.Y, Thing{Kind: KindRoad, Key: key})
		}
	}
}

var roadNames = [3]string{"main road", "street", "lane"}

// anchor is where art w by h stands on a parcel: toward the road it opens on.
func anchor(l Lot, w, h int) (int, int) {
	r := l.Rect
	x, y := r.X+(r.W-w)/2, r.Y+(r.H-h)/2
	switch {
	case l.Door.Y >= r.Y+r.H:
		y = r.Y + r.H - h
	case l.Door.Y < r.Y:
		y = r.Y
	case l.Door.X < r.X:
		x = r.X
	default:
		x = r.X + r.W - w
	}
	return x, y
}

// size is the width and height of art; short rows count as padded.
func (a Art) size() (w, h int) {
	for _, row := range a {
		w = max(w, len([]rune(row)))
	}
	return w, len(a)
}

// stamp draws art on a parcel, toward its door. look says which glyph and
// colour each art rune is drawn as. Windows light up at night and chimneys
// smoke.
func (f *frame) stamp(a Art, l Lot, thing Thing, look func(r rune) (rune, Paint)) {
	w, h := a.size()
	x0, y0 := anchor(l, w, h)
	if thing.Kind == KindBuilding {
		f.targets = append(f.targets, middleOf(a, x0, y0))
	}
	for dy, row := range a {
		for dx, r := range []rune(row) {
			if r == ' ' {
				continue
			}
			p := Pt{x0 + dx, y0 + dy}
			glyph, ink := look(r)
			f.g.draw(p.X, p.Y, glyph, ink)
			f.g.mark(p.X, p.Y, thing)
			if strings.ContainsRune(f.st.Windows, r) {
				f.glow = append(f.glow, p)
			}
			if r == f.st.Chimney {
				f.vents = append(f.vents, p)
			}
		}
	}
}

// ink is the paint of an art rune: its own if the style names one, else the
// hue of the building line, or the walls' when there is none.
func (f *frame) ink(r rune, line Hue) Paint {
	if c, ok := f.st.Ink[r]; ok {
		return c
	}
	if line != HNone {
		return P(line, 235)
	}
	return P(f.st.Wall, 235)
}

// paintBuilding stamps the footprint of a building kind on its parcel.
func (f *frame) paintBuilding(pl placed) {
	fam := familyOf[pl.b.Lineage]
	tier := min(pl.tier, numTiers-1)
	for len(f.st.Shapes[fam][tier]) == 0 && tier > 0 {
		tier--
	}
	variants := f.st.Shapes[fam][tier]
	shape := variants[roll(len(variants), f.m.Seed, mapmodel.HashStr(pl.b.Key), int64(pl.instance))]
	line := lineHue(pl.b.Lineage)
	thing := Thing{Kind: KindBuilding, Key: pl.b.Key, Lineage: pl.b.Lineage, Count: pl.b.Count, Legacy: pl.b.Legacy}
	f.stamp(shape, f.plan.Lots[pl.lot], thing, func(r rune) (rune, Paint) {
		ink := f.ink(r, line)
		if r == '*' {
			r, ink = lineSigns[pl.b.Lineage], P(line, 255)
		}
		if pl.b.Legacy {
			ink = ink.fade(150)
		}
		return r, ink
	})
}

// paintHouse stamps an ordinary house, in the quiet colours of background.
func (f *frame) paintHouse(lot int) {
	shape := f.st.Filler[roll(len(f.st.Filler), f.m.Seed, 67, int64(lot))]
	f.stamp(shape, f.plan.Lots[lot], Thing{Kind: KindAmbient, Key: "houses"}, func(r rune) (rune, Paint) {
		return r, f.ink(r, HNone).fade(170)
	})
}

// middleOf is the drawn cell of art nearest its middle, where a cursor lands.
func middleOf(rows []string, x0, y0 int) Pt {
	best, bestGap := Pt{x0, y0}, 1<<30
	for dy, row := range rows {
		for dx, r := range []rune(row) {
			w, _ := Art(rows).size()
			if gap := abs(dx-w/2) + 2*abs(dy-len(rows)/2); r != ' ' && gap < bestGap {
				best, bestGap = Pt{x0 + dx, y0 + dy}, gap
			}
		}
	}
	return best
}

// paintPort lays the spaceport: paved ground with its launch pads in a row,
// as many as the age has built.
func (f *frame) paintPort() {
	port := f.plan.Port
	if f.st.Pads == 0 || port.Rect.W == 0 {
		return
	}
	thing := Thing{Kind: KindTerrain, Key: "port"}
	f.paintYardOf(port.Rect, f.st.Yard)
	port.Rect.cells(func(x, y int) { f.g.mark(x, y, thing) })
	w, h := f.st.Pad.size()
	for i := 0; i < f.st.Pads; i++ {
		x0 := port.Rect.X + 1 + i*(w+1)
		y0 := port.Rect.Y + port.Rect.H - h
		if port.Door.Y >= port.Rect.Y+port.Rect.H {
			y0 = port.Rect.Y
		}
		for dy, row := range f.st.Pad {
			for dx, r := range []rune(row) {
				if r != ' ' {
					f.g.draw(x0+dx, y0+dy, r, f.ink(r, HNone))
					f.g.mark(x0+dx, y0+dy, thing)
				}
			}
		}
		if i == 0 {
			f.targets = append(f.targets, Pt{x0 + w/2, y0 + h/2})
		}
	}
}
