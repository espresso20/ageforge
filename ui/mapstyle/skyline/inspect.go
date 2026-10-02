package skyline

import (
	"sort"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// inspect.go is the cursor: what it can stand on (every silhouette, the civ
// towns on the ridge, the harbinger, and while one is in view the rare
// visitor's saucer), how the keys move it, and what it says about what it
// is on, in the model's command vocabulary.

type tkind uint8

const (
	tNone tkind = iota
	tLot
	tCiv
	tHarbinger
	tUFO    // the visitor's saucer: ↑ past the ridge reaches it, Tab never does
	tTether // the space elevator, from the Fusion Age (city.go)
)

// target names what the cursor is on, stably across model rebuilds.
type target struct {
	kind tkind
	key  string
	cp   int
}

// tgt is a target placed: its depth row (0 front … 2 back, 3 the ridge)
// and a world column to compare positions by.
type tgt struct {
	target
	row, x int
}

const ridgeRow = 3

// targets lists every target in Tab order: buildings west to east (front
// row first at a column), then the ridge west to east, the harbinger last.
func targets(m *mapmodel.Model, w, cam int) []tgt {
	out := make([]tgt, 0, len(m.Skyline.Lots)+8)
	for i := range m.Skyline.Lots {
		l := &m.Skyline.Lots[i]
		if m.Catalog.Defs[l.Key] == nil {
			continue
		}
		out = append(out, tgt{target{tLot, l.Key, l.Copy}, l.Row, l.X})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].x != out[j].x {
			return out[i].x < out[j].x
		}
		return out[i].row < out[j].row
	})
	for _, it := range ridgeItems(m, w, cam) {
		t := tgt{target: target{kind: tHarbinger}, row: ridgeRow, x: cam + it.x}
		if it.fac != nil {
			t.target = target{kind: tCiv, key: it.fac.Key}
		}
		out = append(out, t)
	}
	if x := tetherX(m); x >= 0 { // the space elevator stands as tall as the ridge
		out = append(out, tgt{target: target{kind: tTether}, row: ridgeRow, x: x})
	}
	return out
}

// targetsFor is targets, cached per (model, width, camera).
func (v *view) targetsFor(m *mapmodel.Model, w, cam int) []tgt {
	if v.ts == nil || v.tsM != m || v.tsW != w || v.tsCam != cam {
		v.ts, v.tsM, v.tsW, v.tsCam = targets(m, w, cam), m, w, cam
	}
	return v.ts
}

func indexOf(ts []tgt, t target) int {
	for i := range ts {
		if ts[i].target == t {
			return i
		}
	}
	return -1
}

// defaultTarget is a building near the present: the lot nearest the newest
// district's middle, front rows preferred.
func defaultTarget(m *mapmodel.Model, ts []tgt) int {
	if len(m.Skyline.Districts) == 0 {
		return -1
	}
	c := m.Skyline.Districts[len(m.Skyline.Districts)-1].Centre
	best, bd := -1, 1<<30
	for i, t := range ts {
		if t.kind != tLot {
			continue
		}
		d := abs(t.x-c) + t.row*4
		if d < bd {
			best, bd = i, d
		}
	}
	if best < 0 && len(ts) > 0 {
		best = 0
	}
	return best
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// resolve returns the cursor's index in ts at frame anim, putting it on the
// default target when it is on nothing that exists. The saucer is in no
// list: on it, resolve returns -1 while it is in view and lets go of it
// once it is not.
func (v *view) resolve(m *mapmodel.Model, ts []tgt, anim int) int {
	if v.cur.kind == tUFO {
		if saucerInView(m, anim, v.viewW()) {
			return -1
		}
		v.cur = target{}
	}
	if i := indexOf(ts, v.cur); i >= 0 {
		return i
	}
	i := defaultTarget(m, ts)
	if i >= 0 {
		v.cur = ts[i].target
	}
	return i
}

// nearestTop is the target in the highest row nearest world column x.
func nearestTop(ts []tgt, x int) int {
	best := 0
	for j, t := range ts {
		if b := ts[best]; t.row > b.row || t.row == b.row && abs(t.x-x) < abs(b.x-x) {
			best = j
		}
	}
	return best
}

// step moves the cursor at frame anim: dx along its row, drow between rows
// (↑ past the top row reaches the saucer while it is in view), tab through
// the Tab order.
func (v *view) step(m *mapmodel.Model, anim, dx, drow, tab int) {
	ts := v.targetsFor(m, v.viewW(), v.cam)
	if len(ts) == 0 {
		return
	}
	i := v.resolve(m, ts, anim)
	if v.cur.kind == tUFO {
		if drow >= 0 && tab == 0 {
			return // nothing beside it or above it
		}
		x, _, _ := saucerAt(m, anim, v.viewW())
		v.cur, v.reveal = ts[nearestTop(ts, v.cam+x+1)].target, true
		return
	}
	if i < 0 {
		return
	}
	cur := ts[i]
	switch {
	case tab != 0:
		i = ((i+tab)%len(ts) + len(ts)) % len(ts)
	case dx != 0:
		best, bd := -1, 1<<30
		for j, t := range ts {
			if t.row != cur.row || j == i {
				continue
			}
			d := (t.x - cur.x) * dx
			if d < 0 || (d == 0 && (j-i)*dx < 0) {
				continue
			}
			if d < bd {
				best, bd = j, d
			}
		}
		if best < 0 {
			return
		}
		i = best
	case drow != 0:
		moved := false
		for row := cur.row + drow; row >= 0 && row <= ridgeRow; row += drow {
			best, bd := -1, 1<<30
			for j, t := range ts {
				if t.row == row && abs(t.x-cur.x) < bd {
					best, bd = j, abs(t.x-cur.x)
				}
			}
			if best >= 0 {
				i, moved = best, true
				break
			}
		}
		if !moved && drow > 0 && saucerInView(m, anim, v.viewW()) {
			v.cur = target{kind: tUFO}
			return
		}
	}
	v.cur = ts[i].target
	v.reveal = true
}

func (v *view) inspection(m *mapmodel.Model, anim int) (in inspectionData, ok bool) {
	ts := v.targetsFor(m, v.viewW(), v.cam)
	i := v.resolve(m, ts, anim)
	if v.cur.kind == tUFO {
		return inspectionData{title: "Unknown craft", lines: []string{"Not one of ours."}, kind: mapstyle.KindAlien}, true
	}
	if i < 0 {
		return in, false
	}
	t := ts[i]
	switch t.kind {
	case tLot:
		return lotInspection(m, t.key, t.cp), true
	case tCiv:
		f := m.Faction(t.key)
		if f == nil {
			return in, false
		}
		in.title = f.Name
		in.lines = append(in.lines, "relation: "+f.Relation.String(),
			"strength "+strconv.Itoa(f.Strength)+" · opinion "+strconv.Itoa(f.Opinion))
		if f.Specialty != "" {
			in.lines = append(in.lines, "specialty: "+strings.ReplaceAll(f.Specialty, "_", " "))
		}
		if f.TradeCount > 0 {
			in.lines = append(in.lines, strconv.Itoa(f.TradeCount)+" trades")
		}
		in.cmd = m.FactionCommand(f)
	case tTether:
		info := mapmodel.FeatTether.Info()
		in.title = info.Title
		in.lines = append(in.lines, info.Lines...)
		in.cmd = mapmodel.CmdStatus
	case tHarbinger:
		h := m.Harbinger
		if h == nil {
			return in, false
		}
		in.title = h.Name
		in.lines = append(in.lines, h.WarningLine())
		if h.Numeric {
			in.lines = append(in.lines, "odds "+strconv.Itoa(int(h.Probability*100+0.5))+"%")
		} else if h.Tier != "" {
			in.lines = append(in.lines, "odds "+h.Tier)
		}
		in.lines = append(in.lines, "pressure "+strconv.Itoa(int(m.Catastrophe.Pressure*100+0.5))+"%")
		in.cmd = mapmodel.CmdHarbinger
	}
	return in, true
}

type inspectionData struct {
	title string
	lines []string
	cmd   string
	kind  string // mapstyle.Inspection.Kind
}

func lotInspection(m *mapmodel.Model, key string, cp int) inspectionData {
	var in inspectionData
	def := m.Catalog.Defs[key]
	b := m.Building(key)
	if def == nil {
		return in
	}
	if def.Wonder {
		in.title = def.Name
		in.lines = append(in.lines, "wonder of the "+m.Catalog.AgeNames[def.Age])
		in.cmd = "wonders"
		for i := range m.Wonders {
			if m.Wonders[i].Key == key {
				in.cmd = m.WonderCommand(&m.Wonders[i])
				if m.Wonders[i].Delta {
					in.lines = append(in.lines, "built since your last visit")
				}
			}
		}
		return in
	}
	in.title = def.Name
	if b == nil {
		return in
	}
	in.title += " ×" + strconv.Itoa(b.Count)
	in.lines = append(in.lines, mapmodel.LineageNames[b.Lineage]+" · "+m.Catalog.AgeNames[def.Age])
	switch {
	case b.Capacity <= 0:
		in.lines = append(in.lines, "no workers needed")
	case b.Workers == 0:
		in.lines = append(in.lines, "idle, 0/"+strconv.Itoa(b.Capacity)+" workers")
	default:
		in.lines = append(in.lines, "workers "+strconv.Itoa(b.Workers)+"/"+strconv.Itoa(b.Capacity))
	}
	if b.Rate > 0 && b.RateName != "" {
		in.lines = append(in.lines, "+"+humanN(b.Rate)+" "+strings.ToLower(b.RateName)+"/s")
	}
	switch {
	case b.Upgrade != "":
		name := b.Upgrade
		if d := m.Catalog.Defs[b.Upgrade]; d != nil {
			name = d.Name
		}
		in.lines = append(in.lines, "upgrade ready: "+name)
	case b.Legacy:
		in.lines = append(in.lines, "legacy: an older tier, it no longer grows")
	}
	if b.Delta > 0 {
		in.lines = append(in.lines, "+"+strconv.Itoa(b.Delta)+" since your last visit")
	}
	in.cmd = m.BuildingCommand(b)
	if in.cmd == "" {
		in.cmd = mapmodel.CmdStatus
	}
	return in
}

// humanN renders a rate compactly: 3.2, 41, 1.2k.
func humanN(n float64) string {
	switch {
	case n >= 1e9:
		return strconv.FormatFloat(n/1e9, 'f', 1, 64) + "B"
	case n >= 1e6:
		return strconv.FormatFloat(n/1e6, 'f', 1, 64) + "M"
	case n >= 1e4:
		return strconv.FormatFloat(n/1e3, 'f', 1, 64) + "k"
	case n >= 100:
		return strconv.FormatFloat(n, 'f', 0, 64)
	}
	return strconv.FormatFloat(n, 'f', 1, 64)
}
