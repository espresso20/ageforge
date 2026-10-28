package mapstyle

import (
	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
)

// freshFrames is how long what is new since the last visit stays lit once the
// map first shows it, in frames of the world's clock (eight to the second), so
// it clears with motion off too.
const freshFrames = 200

// newsKey names the news of a model: the building types that grew and the
// wonders finished since the last visit. It is 0 when there is none.
func newsKey(m *mapmodel.Model) uint64 {
	var key uint64
	for _, b := range m.Buildings {
		if b.Delta > 0 {
			key = mapmodel.Hash(int64(key), mapmodel.HashStr(b.Key), int64(b.Delta))
		}
	}
	for _, w := range m.Wonders {
		if w.Delta {
			key = mapmodel.Hash(int64(key), mapmodel.HashStr(w.Key))
		}
	}
	return key
}

// freshLit says whether the news of the frame's model is still lit: it is for
// freshFrames from the first look, and clears for good once the map has been
// looked at that long (new news lights again).
func (v *view) freshLit(f Frame) bool {
	key := newsKey(f.Model)
	if key == 0 || !v.changes {
		return false
	}
	now := f.Clock
	if now == 0 {
		now = f.Anim
	}
	if key != v.freshKey {
		v.freshKey, v.freshFrom = key, now
	}
	return now-v.freshFrom < freshFrames
}

// focus is the building line in focus, "" outside the district zoom.
func (v *view) focus() string {
	if v.zoomSet && v.zoom == zDistrict {
		return v.district
	}
	return ""
}

// shownZoom is the zoom in force: the player's, once chosen, else the one the
// size of the panel chose.
func (v *view) shownZoom() int {
	switch {
	case v.zoomSet:
		return v.zoom
	case v.overview:
		return zOverview
	}
	return zCloseUp
}

// zoomLabel names the zoom the last frame showed, and for a district whose.
func (v *view) zoomLabel() string {
	label := zoomNames[v.shownZoom()].String()
	if v.shownZoom() == zDistrict {
		label += " · " + lineageLabel(v.district)
	}
	return label
}

// zoomTo sets the zoom the player chose. Entering the district zoom puts the
// cursor in the district of what it is on (or the nearest one); with no town
// to frame it stays a close-up.
func (v *view) zoomTo(sc *mapgen.Scene, zoom int) {
	v.zoom, v.zoomSet, v.recentre = zoom, true, true
	if zoom != zDistrict {
		return
	}
	if len(sc.Districts) == 0 {
		v.zoom = zCloseUp
		return
	}
	here := sc.Grid.ThingAt(v.cur.X, v.cur.Y).Lineage
	best, gap := 0, 1<<60
	for i, d := range sc.Districts {
		dx, dy := d.Heart.X-v.cur.X, d.Heart.Y-v.cur.Y
		if gap2 := dx*dx + 4*dy*dy; d.Line == here || gap2 < gap {
			best, gap = i, gap2
			if d.Line == here {
				break
			}
		}
	}
	v.visitDistrict(sc.Districts[best])
}

// visitDistrict frames a district: the cursor goes to its heart.
func (v *view) visitDistrict(d mapgen.District) {
	v.district, v.cur, v.recentre = d.Line, d.Heart, true
}

// cycleDistrict moves the district zoom on to the next (dir 1) or previous
// building line that has a district.
func (v *view) cycleDistrict(sc *mapgen.Scene, dir int) {
	n := len(sc.Districts)
	if n == 0 {
		return
	}
	at := 0
	for i, d := range sc.Districts {
		if d.Line == v.district {
			at = i
		}
	}
	v.visitDistrict(sc.Districts[((at+dir)%n+n)%n])
}
