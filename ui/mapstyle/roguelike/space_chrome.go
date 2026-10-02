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

// space_chrome.go is everything round a sky scene: the header, the news
// line, the inspector, the status line, the flows panel, the legend and the
// compact view. It keeps the town's layout, so the panel reads the same in
// every age.

// skyViewNames name each scene's view on the header and the status line.
var skyViewNames = [mapmodel.NumSkyScenes]string{"", "ORBIT", "DEEP SPACE", "STARBASE", "SUPERPOSITION", "MANDALA"}

func (v *skyView) viewName() string {
	n := skyViewNames[v.scene]
	if v.g.zoom == zDistrict {
		n += " CLOSE"
	}
	return n
}

// Draw renders the full view: header, map, side panel, news, inspector and
// status line. Below 60x16 it draws the compact view instead.
func (v *skyView) Draw(scr tcell.Screen, r mapstyle.Rect, f mapstyle.Frame) {
	if r.W < 60 || r.H < 16 {
		v.DrawCompact(scr, r, f)
		return
	}
	s := v.begin(f)
	cv := mapstyle.NewCanvas(scr, r, f.Tier, v.g.bgStyle())
	if s == nil {
		cv.Text(1, 0, 0, "No map yet.", v.g.cls(mapmodel.CDim))
		return
	}
	g := v.g
	side := 0
	if r.W >= 110 && (g.legend || g.flows) {
		side = sideW
	}
	mapW, mapH := r.W-side, r.H-5
	cv.Fill(0, 1, mapW, mapH, ' ', tcell.StyleDefault.Background(v.sp.void))
	v.drawSkyMap(cv, 0, 1, mapW, mapH)
	v.header(cv, r.W)
	if side > 0 {
		v.sidebar(cv, mapW, 1, side, mapH)
	} else if g.flows {
		v.box(cv, mapW-min(sideW, mapW/2), 1, min(sideW, mapW/2), mapH)
	}
	v.bottom(cv, r.W, r.H)
}

// header: the settlement, age, epoch, the clock, people, idle, civs, view.
func (v *skyView) header(cv *mapstyle.Canvas, W int) {
	m, ch := v.sc.m, v.g.pal.Chrome
	_, cbg, _ := ch.Decompose()
	cv.Fill(0, 0, W, 1, ' ', ch)
	name := ""
	if w := m.Town.World; w != nil {
		name = w.Name
	}
	x := cv.Text(1, 0, W-2, "◆ "+name, ch.Foreground(theme.Legible(theme.Color(theme.RoleAccent), cbg, 3)).Bold(true))
	x = cv.Text(x, 0, W-x-1, " · "+m.AgeName+" · "+m.EpochName, ch)
	idle := ch
	if m.Workers.Idle > 0 {
		idle = ch.Foreground(theme.Legible(v.g.pal.Fg[mapmodel.CIdle], cbg, 3)).Bold(true)
	}
	segs := []struct {
		s  string
		st tcell.Style
	}{
		{string(weatherRune(m, v.tier)) + " " + m.Clock.String() + "  ", ch},
		{"pop " + strconv.Itoa(m.Workers.Pop) + "  ", ch},
		{"idle " + strconv.Itoa(m.Workers.Idle), idle},
		{"  civs " + strconv.Itoa(m.DiscoveredCount()) + "/" + strconv.Itoa(len(m.Factions)), ch},
		{"  " + v.viewName() + " ", ch.Bold(true)},
	}
	for ; len(segs) > 0; segs = segs[1:] {
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

// news is the one line worth knowing: the recap headline, the pending
// catastrophe, or what is stuck.
func (v *skyView) news(width int) (string, tcell.Style) {
	m := v.sc.m
	g := v.g
	danger := g.role(theme.RoleNegative).Bold(true)
	h := ""
	if g.changes {
		h = m.Recap.Headline(width - 2)
	}
	if width < 60 && h != "" {
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
		return h, g.role(theme.RolePositive)
	case m.Catastrophe.Pending != "":
		return mapmodel.Clip("DANGER: "+m.Catastrophe.PendingName+" is upon you", width), danger
	case m.Flows.Worst != "" && g.changes:
		return mapmodel.Clip("stuck: "+m.Flows.Worst, width), g.role(theme.RoleWarning)
	}
	return "all quiet", g.cls(mapmodel.CDim)
}

// bottom: the news, the two inspector lines and the status line.
func (v *skyView) bottom(cv *mapstyle.Canvas, W, H int) {
	g := v.g
	n, st := v.news(W - 2)
	cv.Text(1, H-4, W-2, n, st)
	dim := g.cls(mapmodel.CDim)
	hints := "arrows move  Shift fast  PgUp world  PgDn close  Tab next  Home hub  map flows"
	if W < 100 {
		hints = "arrows move  PgUp PgDn zoom  Tab next"
	}
	switch in, ok := v.Inspect(mapstyle.Frame{Model: v.sc.m, Anim: v.anim, Tier: v.tier}); {
	case !ok:
		cv.Text(1, H-3, W-2, "cursor hidden; move it with the arrow keys", g.bgStyle())
		cv.Text(3, H-2, W-4, hints, dim)
	default:
		cv.Text(1, H-3, W-2, "▸ "+strings.Join(append([]string{in.Title}, in.Lines...), " · "), g.role(theme.RoleBright).Bold(true))
		if in.Command == "" {
			cv.Text(3, H-2, W-4, "nothing to type here; Tab jumps to the next building", dim)
		} else {
			x := cv.Text(3, H-2, W-4, "type: ", dim)
			cv.Text(x, H-2, W-x-1, in.Command, g.cls(mapmodel.CAccent).Bold(true))
		}
	}
	left := v.viewName()
	if g.flows {
		left += " · flows"
	}
	x := cv.Text(1, H-1, W-2, left, g.cls(mapmodel.CLabel))
	if hw := mapstyle.TextLen(hints); x+2+hw <= W-1 {
		cv.Text(W-1-hw, H-1, hw, hints, dim)
	}
}

// sidebar: the flows panel on top when it is on, the legend below.
func (v *skyView) sidebar(cv *mapstyle.Canvas, x0, y0, w, h int) {
	g := v.g
	for y := y0; y < y0+h; y++ {
		cv.Put(x0, y, '│', g.pal.border)
	}
	y := y0
	if g.flows {
		cv.Text(x0+2, y, w-3, "FLOWS", g.cls(mapmodel.CAccent).Bold(true))
		for _, l := range v.flowLines() {
			if y++; y < y0+h {
				g.flowLine(cv, x0+2, y, w-3, l)
			}
		}
		y += 2
	}
	if g.legend && y < y0+h {
		v.drawSkyLegend(cv, x0+2, y, w-3, y0+h-y)
	}
}

// flowLines is the flows panel: what is full, draining, idle or short, and
// the largest lineages by share, in each one's sky symbol.
func (v *skyView) flowLines() []flowLine {
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
		p := mapmodel.SkyPartOf(v.sc.sky, sh.Lineage)
		out = append(out, flowLine{text: p.Name, bar: sh.Share, c: mapmodel.LineageClass(sh.Lineage),
			glyph: mapmodel.R(p.Sym, v.tier)})
	}
	return out
}

// box draws the flows panel in a frame over the map.
func (v *skyView) box(cv *mapstyle.Canvas, x, y, w, maxH int) {
	g := v.g
	lines := v.flowLines()
	h := min(len(lines)+2, maxH)
	if w < 8 || h < 3 {
		return
	}
	b := g.pal.border
	cv.Fill(x, y, w, h, ' ', g.bgStyle())
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
	cv.Text(x+2, y, w-4, "FLOWS", g.cls(mapmodel.CAccent).Bold(true))
	for i := 0; i < len(lines) && i < h-2; i++ {
		g.flowLine(cv, x+1, y+1+i, w-2, lines[i])
	}
}

// DrawCompact renders the glanceable mini view: a header line, the whole
// scene fitted to the space (down-sampled, the most telling cell of each
// block winning) with its traffic, and one news line. The rare visitor
// never shows here.
func (v *skyView) DrawCompact(scr tcell.Screen, r mapstyle.Rect, f mapstyle.Frame) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	s := v.begin(f)
	g := v.g
	cv := mapstyle.NewCanvas(scr, r, f.Tier, g.bgStyle())
	if s == nil {
		cv.Text(0, 0, 0, "No map yet.", g.cls(mapmodel.CDim))
		return
	}
	m := s.m
	x := cv.Text(0, 0, r.W, strings.TrimSuffix(m.AgeName, " Age")+" · pop "+strconv.Itoa(m.Workers.Pop), g.cls(mapmodel.CAccent).Bold(true))
	if m.Workers.Idle > 0 {
		cv.Text(x, 0, r.W-x, " · idle "+strconv.Itoa(m.Workers.Idle), g.cls(mapmodel.CIdle).Bold(true))
	}
	if r.H >= 2 {
		n, st := v.news(r.W)
		cv.Text(0, r.H-1, r.W, n, st)
	}
	mh := r.H - 2
	if mh <= 0 {
		return
	}
	v.compact = true
	defer func() { v.compact = false }()
	if s.sky == mapmodel.SkyMandala {
		v.compactMandala(cv, r.W, mh)
		return
	}
	sx := int(math.Ceil(float64(skyW) / float64(r.W)))
	sy := int(math.Ceil(float64(skyH) / float64(mh)))
	sc := max(1, sx, sy)
	ox := skyW/2 - r.W*sc/2
	oy := skyH/2 - mh*sc/2
	for cy := 0; cy < mh; cy++ {
		for cx := 0; cx < r.W; cx++ {
			best, bs := skyCell{r: ' '}, -1
			bx, by := 0, 0
			for j := 0; j < sc*sc; j++ {
				px, py := ox+cx*sc+j%sc, oy+cy*sc+j/sc
				c := s.at(px, py)
				if sal := c.salience(); sal > bs {
					best, bs, bx, by = c, sal, px, py
				}
			}
			l := v.resolve(s, best, bx, by)
			l = v.stateOf(s, best, l, bx, by)
			cv.Put(cx, 1+cy, l.r, v.style(l))
		}
	}
	// traffic, placed by the same down-sampling
	for i := range s.movers {
		mv := &s.movers[i]
		at := mv.where(v.anim)
		if at.i < 0 {
			continue
		}
		p := mv.lane[at.i]
		cx, cy := floorDiv(p.X-ox, sc), floorDiv(p.Y-oy, sc)
		if cx < 0 || cy < 0 || cx >= r.W || cy >= mh || blocks(s.at(p.X, p.Y)) {
			continue
		}
		ink := mapmodel.InkStarBright
		if mv.kind == smAlien {
			ink = inkAlien + mapmodel.SkyInk(mv.alien%len(mapmodel.Aliens))
		}
		cv.Put(cx, 1+cy, mapmodel.R(mv.info.Sym, v.tier), v.style(look{ink: ink, lv: 3, bold: true}))
	}
}
