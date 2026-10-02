package roguelike

import (
	"math"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// chrome.go is everything round the map: the header, the news line, the
// inspector, the status line, the legend, the flows panel and the compact
// view.

const sideW = 26

// begin readies the per-frame state and returns the scene (nil without a
// model).
func (v *view) begin(f mapstyle.Frame) *scene {
	v.anim, v.tier, v.seen = f.Anim, f.Tier, [numLg]lgEntry{}
	s := v.sceneFor(f.Model)
	ep := 0
	if s != nil {
		ep = s.epoch
	}
	v.palette(ep)
	return s
}

func (v *view) bgStyle() tcell.Style { return v.cls(mapmodel.CText) }

// Draw renders the full view: header, map, side panel, news, inspector and
// status line. Below 60x16 it draws the compact view instead.
func (v *view) Draw(scr tcell.Screen, r mapstyle.Rect, f mapstyle.Frame) {
	if r.W < 60 || r.H < 16 {
		v.DrawCompact(scr, r, f)
		return
	}
	s := v.begin(f)
	cv := mapstyle.NewCanvas(scr, r, f.Tier, v.bgStyle())
	if s == nil {
		cv.Text(1, 0, 0, "No map yet.", v.cls(mapmodel.CDim))
		return
	}
	side := 0
	if r.W >= 110 && (v.legend || v.flows) {
		side = sideW
	}
	mapW, mapH := r.W-side, r.H-5
	v.drawMap(cv, 0, 1, mapW, mapH)
	v.header(cv, r.W)
	if side > 0 {
		v.sidebar(cv, mapW, 1, side, mapH)
	} else if v.flows { // no room for a sidebar: a box over the map
		v.box(cv, mapW-min(sideW, mapW/2), 1, min(sideW, mapW/2), mapH)
	}
	v.bottom(cv, r.W, r.H)
}

// header: settlement, age, epoch, weather, people, idle, civs met, zoom.
func (v *view) header(cv *mapstyle.Canvas, W int) {
	s, m, ch := v.sc, v.sc.m, v.pal.Chrome
	_, cbg, _ := ch.Decompose()
	cv.Fill(0, 0, W, 1, ' ', ch)
	x := cv.Text(1, 0, W-2, "◆ "+s.w.Name, ch.Foreground(theme.Legible(theme.Color(theme.RoleAccent), cbg, 3)).Bold(true))
	x = cv.Text(x, 0, W-x-1, " · "+m.AgeName+" · "+m.EpochName, ch)
	idle := ch
	if m.Workers.Idle > 0 {
		idle = ch.Foreground(theme.Legible(v.pal.Fg[mapmodel.CIdle], cbg, 3)).Bold(true)
	}
	segs := []struct {
		s  string
		st tcell.Style
	}{
		{string(weatherRune(m, v.tier)) + " " + m.Clock.String() + "  ", ch},
		{"pop " + strconv.Itoa(m.Workers.Pop) + "  ", ch},
		{"idle " + strconv.Itoa(m.Workers.Idle), idle},
		{"  civs " + strconv.Itoa(m.DiscoveredCount()) + "/" + strconv.Itoa(len(m.Factions)), ch},
		{"  " + zoomNames[v.zoom] + " ", ch.Bold(true)},
	}
	for ; len(segs) > 0; segs = segs[1:] { // drop from the left until it fits
		n := 0
		for _, sg := range segs {
			n += mapstyle.TextLen(sg.s)
		}
		if x+n+2 <= W {
			for rx, i := W-n, 0; i < len(segs); i++ {
				rx = cv.Text(rx, 0, 0, segs[i].s, segs[i].st)
			}
			return
		}
	}
}

func weatherRune(m *mapmodel.Model, tier mapmodel.GlyphTier) rune {
	sym := mapmodel.SymSun
	switch m.Weather.Kind {
	case mapmodel.Cloudy, mapmodel.Smog:
		sym = mapmodel.SymCloud
	case mapmodel.Rain:
		sym = mapmodel.SymRain
	case mapmodel.Storm:
		sym = mapmodel.SymBolt
	case mapmodel.Snow:
		sym = mapmodel.SymSnow
	default:
		if m.Clock.Night > 0.5 {
			sym = mapmodel.SymMoon
		}
	}
	return mapmodel.R(sym, tier)
}

// news is the one line worth knowing: the recap headline, the pending
// catastrophe, or what is stuck.
func (v *view) news(width int) (string, tcell.Style) {
	m := v.sc.m
	danger := v.role(theme.RoleNegative).Bold(true)
	h := ""
	if v.changes {
		h = m.Recap.Headline(width - 2)
	}
	if width < 60 && h != "" { // narrow: the items without the preamble
		var parts []string
		for _, it := range m.Recap.Items {
			parts = append(parts, it.Text)
		}
		h = mapmodel.Clip(strings.Join(parts, " · "), width)
	}
	switch {
	case h != "" && m.Recap.Items[0].Bad:
		return h, danger
	case h != "":
		return h, v.role(theme.RolePositive)
	case m.Catastrophe.Pending != "":
		return mapmodel.Clip("DANGER: "+m.Catastrophe.PendingName+" is upon you", width), danger
	case m.Flows.Worst != "" && v.changes:
		return mapmodel.Clip("stuck: "+m.Flows.Worst, width), v.role(theme.RoleWarning)
	}
	return "all quiet", v.cls(mapmodel.CDim)
}

// bottom: news, the two inspector lines and the status line.
func (v *view) bottom(cv *mapstyle.Canvas, W, H int) {
	n, st := v.news(W - 2)
	cv.Text(1, H-4, W-2, n, st)
	dim := v.cls(mapmodel.CDim)
	hints := "arrows move  Shift fast  PgUp PgDn zoom  Tab next  Home center  map flows"
	if W < 100 {
		hints = "arrows move  PgUp PgDn zoom  Tab next"
	}
	switch in, ok := v.Inspect(mapstyle.Frame{Model: v.sc.m, Anim: v.anim, Tier: v.tier}); {
	case !ok:
		cv.Text(1, H-3, W-2, "cursor hidden; move it with the arrow keys", v.bgStyle())
		cv.Text(3, H-2, W-4, hints, dim)
	default:
		cv.Text(1, H-3, W-2, "▸ "+strings.Join(append([]string{in.Title}, in.Lines...), " · "), v.role(theme.RoleBright).Bold(true))
		if in.Command == "" {
			cv.Text(3, H-2, W-4, "nothing to type here; Tab jumps to the next building", dim)
		} else {
			x := cv.Text(3, H-2, W-4, "type: ", dim)
			cv.Text(x, H-2, W-x-1, in.Command, v.cls(mapmodel.CAccent).Bold(true))
		}
	}
	left := zoomNames[v.zoom]
	if v.zoom == zRegion {
		left += " · " + plates[v.sc.epoch].name
	}
	if v.flows {
		left += " · flows"
	}
	x := cv.Text(1, H-1, W-2, left, v.cls(mapmodel.CLabel))
	if hw := mapstyle.TextLen(hints); x+2+hw <= W-1 {
		cv.Text(W-1-hw, H-1, hw, hints, dim)
	}
}

// sidebar: the flows panel on top when it is on, the legend below.
func (v *view) sidebar(cv *mapstyle.Canvas, x0, y0, w, h int) {
	for y := y0; y < y0+h; y++ {
		cv.Put(x0, y, '│', v.pal.border)
	}
	y := y0
	if v.flows {
		cv.Text(x0+2, y, w-3, "FLOWS", v.cls(mapmodel.CAccent).Bold(true))
		for _, l := range v.flowLines() {
			if y++; y < y0+h {
				v.flowLine(cv, x0+2, y, w-3, l)
			}
		}
		y += 2
	}
	if v.legend && y < y0+h {
		v.drawLegend(cv, x0+2, y, w-3, y0+h-y)
	}
}

// flowLine is one flows row: a tag and text, or a share bar.
type flowLine struct {
	tag, text string
	c         mapmodel.Class
	bar       float64 // > 0: a share bar
	glyph     rune
}

func (v *view) flowLines() []flowLine {
	fl := v.sc.m.Flows
	var out []flowLine
	names := func(ss []mapmodel.Store) string {
		var n []string
		for _, s := range ss {
			n = append(n, strings.ToLower(s.Name))
		}
		return strings.Join(n, ", ")
	}
	add := func(ok bool, tag, text string, c mapmodel.Class) {
		if ok {
			out = append(out, flowLine{tag: tag, text: text, c: c})
		}
	}
	add(len(fl.Full) > 0, "FULL ", names(fl.Full), mapmodel.CIdle)
	add(len(fl.Draining) > 0, "DRAIN", names(fl.Draining), mapmodel.CDanger)
	add(fl.IdleWorkers > 0, "IDLE ", plural(fl.IdleWorkers, "worker", "workers"), mapmodel.CIdle)
	short := len(fl.Understaffed) + len(fl.Idle)
	add(short > 0, "SHORT", plural(short, "building", "buildings"), mapmodel.CIdle)
	add(len(out) == 0, "OK   ", "nothing stuck", mapmodel.CFresh)
	for i, sh := range fl.Shares {
		if i == 5 || sh.Share <= 0 {
			break
		}
		out = append(out, flowLine{text: lineageLabel(sh.Lineage), bar: sh.Share, c: mapmodel.LineageClass(sh.Lineage),
			glyph: mapmodel.R(mapmodel.LineageSym(sh.Lineage, v.sc.epoch), v.tier)})
	}
	return out
}

func (v *view) flowLine(cv *mapstyle.Canvas, x, y, w int, l flowLine) {
	if l.bar <= 0 {
		xx := cv.Text(x, y, w, l.tag, v.cls(l.c).Bold(true))
		cv.Text(xx+1, y, x+w-xx-1, l.text, v.cls(mapmodel.CText))
		return
	}
	cv.Put(x, y, l.glyph, v.cls(l.c))
	fill := max(1, int(math.Round(float64(l.bar*6))))
	for i := 0; i < 6 && i+2 < w; i++ {
		if i < fill {
			cv.Put(x+2+i, y, '█', v.cls(l.c))
		} else {
			cv.Put(x+2+i, y, '░', v.cls(mapmodel.CMemory))
		}
	}
	xx := cv.Text(x+9, y, w-9, strconv.Itoa(int(float64(l.bar*100)+0.5))+"% ", v.cls(mapmodel.CText))
	cv.Text(xx, y, x+w-xx, l.text, v.cls(mapmodel.CDim))
}

// box draws the flows panel in a frame over the map.
func (v *view) box(cv *mapstyle.Canvas, x, y, w, maxH int) {
	lines := v.flowLines()
	h := min(len(lines)+2, maxH)
	if w < 8 || h < 3 {
		return
	}
	b := v.pal.border
	cv.Fill(x, y, w, h, ' ', v.bgStyle())
	for i := 0; i < w; i++ {
		cv.Put(x+i, y, '─', b)
		cv.Put(x+i, y+h-1, '─', b)
	}
	for j := 0; j < h; j++ {
		cv.Put(x, y+j, '│', b)
		cv.Put(x+w-1, y+j, '│', b)
	}
	for k, c := range [4][2]int{{x, y}, {x + w - 1, y}, {x, y + h - 1}, {x + w - 1, y + h - 1}} {
		cv.Put(c[0], c[1], []rune("┌┐└┘")[k], b)
	}
	cv.Text(x+2, y, w-4, "FLOWS", v.cls(mapmodel.CAccent).Bold(true))
	for i := 0; i < len(lines) && i < h-2; i++ {
		v.flowLine(cv, x+1, y+1+i, w-2, lines[i])
	}
}

// drawLegend lists what this frame actually drew, by group.
func (v *view) drawLegend(cv *mapstyle.Canvas, x, y, w, h int) {
	cv.Text(x, y, w, "LEGEND", v.cls(mapmodel.CAccent).Bold(true))
	row, end := y+1, y+h
	for grp := uint8(0); grp < uint8(len(lgGroups)); grp++ {
		head := false
		for id := lgID(0); id < numLg; id++ {
			label, g := lgLabel(id)
			if e := v.seen[id]; e.on && g == grp && label != "" {
				if !head && row+1 < end {
					cv.Text(x, row, w, lgGroups[grp], v.cls(mapmodel.CDim))
					row, head = row+1, true
				}
				if !head || row >= end {
					return
				}
				cv.Put(x+1, row, e.r, e.st)
				cv.Text(x+3, row, w-3, label, v.cls(mapmodel.CText))
				row++
			}
		}
	}
}

// CompactShowsNews: the compact view's last row is the news line.
func (v *view) CompactShowsNews() bool { return true }

// DrawCompact renders the glanceable mini view (designed for 40x15): a
// header line, the settlement fitted to the space, and one news line.
func (v *view) DrawCompact(scr tcell.Screen, r mapstyle.Rect, f mapstyle.Frame) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	s := v.begin(f)
	cv := mapstyle.NewCanvas(scr, r, f.Tier, v.bgStyle())
	if s == nil {
		cv.Text(0, 0, 0, "No map yet.", v.cls(mapmodel.CDim))
		return
	}
	m := s.m
	x := cv.Text(0, 0, r.W, strings.TrimSuffix(m.AgeName, " Age")+" · pop "+strconv.Itoa(m.Workers.Pop), v.cls(mapmodel.CAccent).Bold(true))
	if m.Workers.Idle > 0 {
		cv.Text(x, 0, r.W-x, " · idle "+strconv.Itoa(m.Workers.Idle), v.cls(mapmodel.CIdle).Bold(true))
	}
	if r.H >= 2 {
		n, st := v.news(r.W)
		cv.Text(0, r.H-1, r.W, n, st)
	}
	if mh := r.H - 2; mh > 0 { // the town, auto-fitted; down-sampled if it must be
		cx, cy := (s.x0+s.x1)/2, (s.y0+s.y1)/2
		sc := max(int(ceilDiv(float64(s.x1-s.x0+3), float64(r.W))), int(ceilDiv(float64(s.y1-s.y0+3), float64(mh))))
		g := geom{x: 0, y: 1, w: r.W, h: mh, zoom: zSettlement, scale: 1, cellW: 1, vx: cx - r.W/2, vy: cy - mh/2}
		if sc > 1 {
			g.zoom, g.scale, g.vx, g.vy = zRegion, sc, cx-r.W*sc/2, cy-mh*sc/2
		}
		saved := v.g
		v.g, v.compact = g, true
		v.drawTiles(cv, g)
		v.drawLife(cv)
		v.g, v.compact = saved, false
	}
}
