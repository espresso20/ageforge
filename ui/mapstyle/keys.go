package mapstyle

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
)

var keyMoves = map[tcell.Key][2]int{tcell.KeyLeft: {-1, 0}, tcell.KeyRight: {1, 0}, tcell.KeyUp: {0, -1}, tcell.KeyDown: {0, 1}}

// HandleKey takes only keys that print nothing, so typing reaches the prompt:
// the arrows move the cursor (Shift by 8), PgUp and PgDn zoom out and in
// (the whole town, the close-up, a district), Tab and Shift-Tab jump between
// buildings, wonders and civs (between districts, at the district zoom), Home
// centres on the town square.
func (v *view) HandleKey(ev *tcell.EventKey, f Frame) bool {
	sc := v.scene(f)
	if ev == nil || sc == nil {
		return false
	}
	switch ev.Key() {
	case tcell.KeyTab, tcell.KeyBacktab:
		dir := map[bool]int{true: 1, false: -1}[ev.Key() == tcell.KeyTab]
		if v.shownZoom() == zDistrict {
			v.cycleDistrict(sc, dir)
		} else {
			v.jump(sc, dir)
		}
		return true
	case tcell.KeyPgUp:
		v.zoomTo(sc, max(v.shownZoom()-1, zOverview))
		return true
	case tcell.KeyPgDn:
		v.zoomTo(sc, min(v.shownZoom()+1, zDistrict))
		return true
	case tcell.KeyHome:
		v.cur, v.recentre = sc.Centre, true
		return true
	}
	d, ok := keyMoves[ev.Key()]
	if !ok {
		return false
	}
	if ev.Modifiers()&tcell.ModShift != 0 {
		d = [2]int{d[0] * 8, d[1] * 8}
	}
	if v.overview && v.ov != nil {
		d = [2]int{d[0] * v.ov.Scale, d[1] * v.ov.Scale}
	}
	v.inspect = true
	v.cur.X = clamp(v.cur.X+d[0], 0, mapgen.WorldW-1)
	v.cur.Y = clamp(v.cur.Y+d[1], 0, mapgen.WorldH-1)
	return true
}

// jump moves the cursor to the next (dir 1) or previous building or wonder.
func (v *view) jump(sc *mapgen.Scene, dir int) {
	ts := sc.Targets
	if len(ts) == 0 {
		return
	}
	v.inspect = true
	n := len(ts) - 1 // a step back from nowhere lands on the last
	if dir > 0 {
		n = 0
	}
	for i, t := range ts {
		if t == v.cur {
			n = ((i+dir)%len(ts) + len(ts)) % len(ts)
			break
		}
	}
	v.cur = ts[n]
}
