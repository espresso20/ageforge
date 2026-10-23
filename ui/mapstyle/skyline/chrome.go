package skyline

import (
	"math"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// chrome.go: the header bar, the minimap strip and the status line. They
// are UI, so they take their colours from theme roles only.

type seg struct {
	s    string
	role theme.Role
}

// role is a theme role made legible on bg.
func role(r theme.Role, bg tcell.Color) tcell.Color { return theme.Legible(theme.Color(r), bg, 3) }

func chromeBg() tcell.Color {
	if theme.IsLight() {
		return theme.Color(theme.RoleChip)
	}
	return theme.Color(theme.RoleSurface)
}

// segs writes parts on row y from column x, clipped to the width.
func (s *scene) segs(x, y int, bg tcell.Color, parts []seg) int {
	for _, p := range parts {
		if x >= s.W {
			break
		}
		x = s.fb.text(x, y, clip(p.s, s.W-x), role(p.role, bg), bg, dTop)
	}
	return x
}

func clip(str string, n int) string { return mapmodel.Clip(str, n) }

// segsLen is how many cells parts take.
func segsLen(parts []seg) int {
	n := 0
	for _, p := range parts {
		n += len([]rune(p.s))
	}
	return n
}

// fitHints returns the first of options that fits in w cells with tail
// after it, joined to the tail: the status line names its keys as fully as
// the width allows and never clips the tail (where the view is). With none
// that fits it is the last option, and the tail clips.
func fitHints(w int, tail []seg, options ...[]seg) []seg {
	pick := options[len(options)-1]
	for _, opt := range options {
		if segsLen(opt)+segsLen(tail) <= w {
			pick = opt
			break
		}
	}
	return append(append([]seg(nil), pick...), tail...)
}

// panoramaHints is the status line at rest of a panorama (the skyline on
// the ground and in the sky): its keys, then where the view is.
func panoramaHints(w int, where string) []seg {
	key, dim := theme.RoleAccent, theme.RoleDim
	tail := []seg{{"│ ", dim}, {where, theme.RoleLabel}}
	return fitHints(w, tail,
		[]seg{{" ◄► ", key}, {"scroll  ", dim}, {"+ - ", key}, {"zoom  ", dim}, {"[ ] ", key}, {"district  ", dim},
			{"Home End ", key}, {"oldest, present  ", dim}, {"Tab ", key}, {"inspect  ", dim}, {"map flows ", key}, {"overlay  ", dim}},
		[]seg{{" ◄► ", key}, {"scroll  ", dim}, {"+ - ", key}, {"zoom  ", dim}, {"[ ] ", key}, {"district  ", dim},
			{"Home End ", key}, {"oldest, now  ", dim}, {"Tab ", key}, {"inspect  ", dim}, {"map flows  ", key}},
		[]seg{{" ◄► ", key}, {"scroll  ", dim}, {"+ - ", key}, {"zoom  ", dim}, {"[ ] ", key}, {"district  ", dim},
			{"Home End  ", key}, {"Tab ", key}, {"inspect  ", dim}},
		[]seg{{" ◄► ", key}, {"scroll  ", dim}, {"+ - ", key}, {"zoom  ", dim}, {"[ ] ", key}, {"Tab ", key}, {"inspect  ", dim}},
		[]seg{{" ◄► ", key}, {"scroll  ", dim}, {"+ - ", key}, {"zoom  ", dim}, {"Tab ", key}, {"inspect  ", dim}},
		[]seg{{" ◄► ", key}, {"scroll  ", dim}, {"Tab ", key}, {"inspect  ", dim}},
		[]seg{{" ◄► Tab  ", key}},
	)
}

func (s *scene) chrome() {
	bg := chromeBg()
	for x := 0; x < s.W; x++ {
		s.fb.set(x, 0, ' ', bg, bg, dTop)
		s.fb.set(x, s.H-1, ' ', bg, bg, dTop)
	}
	s.header(bg)
	s.minimap()
	s.status(bg)
}

// weatherSym is the clock's companion glyph: the sun or moon, or the
// weather when there is some.
func weatherSym(m *mapmodel.Model) mapmodel.Sym {
	switch m.Weather.Kind {
	case mapmodel.Rain:
		return mapmodel.SymRain
	case mapmodel.Storm:
		return mapmodel.SymBolt
	case mapmodel.Snow:
		return mapmodel.SymSnow
	case mapmodel.Cloudy, mapmodel.Smog:
		return mapmodel.SymCloud
	}
	if m.Clock.Night > 0.5 {
		return mapmodel.SymMoon
	}
	return mapmodel.SymSun
}

func idleProducers(m *mapmodel.Model) int { return len(m.Flows.Idle) }

func (s *scene) header(bg tcell.Color) {
	m := s.m
	acc := role(theme.RoleAccent, bg)
	x := 0
	for _, r := range "░▒▓" {
		s.fb.set(x, 0, r, acc, bg, dTop)
		x++
	}
	x = s.fb.text(x, 0, clip(" "+strings.ToUpper(m.AgeName)+" ", s.W-x), theme.Legible(bg, acc, 4.5), acc, dTop)
	for _, r := range "▓▒░" {
		s.fb.set(x, 0, r, acc, bg, dTop)
		x++
	}
	ws := string(mapmodel.R(weatherSym(m), s.tier))
	parts := []seg{
		{" " + m.EpochName, theme.RoleLabel},
		{"  day ", theme.RoleDim}, {strconv.Itoa(m.Clock.Day), theme.RoleHighlight},
		{"  " + m.Clock.String() + " " + ws, theme.RoleText},
		{"  pop ", theme.RoleDim}, {strconv.Itoa(m.Workers.Pop) + "/" + strconv.Itoa(m.Workers.MaxPop), theme.RoleHighlight},
		{"  bld ", theme.RoleDim}, {strconv.Itoa(m.TotalBuildings()), theme.RoleHighlight},
	}
	if n := m.Recap.NewCount; s.v.changes && n > 0 {
		parts = append(parts, seg{"  ▼ " + strconv.Itoa(n) + " new since your last visit", theme.RolePositive})
	}
	if n := idleProducers(m); n > 0 {
		parts = append(parts, seg{"  ▪ " + strconv.Itoa(n) + " idle", theme.RoleWarning})
	}
	if s.v.flows && m.Workers.Idle > 0 {
		parts = append(parts, seg{"  " + strconv.Itoa(m.Workers.Idle) + " idle workers", theme.RoleWarning})
	}
	if h := m.Harbinger; h != nil {
		parts = append(parts, seg{"  " + string(mapmodel.R(mapmodel.SymHarbinger, s.tier)) + " " + h.Name + " is on the ridge", theme.RoleWarning})
	}
	if c := m.Catastrophe; c.Pending != "" {
		parts = append(parts, seg{"  " + string(mapmodel.R(mapmodel.SymFire, s.tier)) + " " + c.PendingName, theme.RoleNegative})
	}
	s.segs(x, 0, bg, parts)
}

var eighths = [9]rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// overviewCell is the cell of an overview W wide that shows town column
// col of a town ww columns wide.
func overviewCell(col, ww, W int) int { return col * W / ww }

// overviewBox is the cells [c0, c1) of an overview W wide that show the
// part of a town ww columns wide that a view of vw columns starting at cam
// sees: from the cell of its first column to the cell of its last. A view
// that runs past the town's end ends the box at the bar's.
func overviewBox(cam, vw, ww, W int) (c0, c1 int) {
	if ww <= 0 || W <= 0 {
		return 0, 0
	}
	last := max(cam, min(cam+vw, ww)-1)
	c0 = clampInt(overviewCell(cam, ww, W), 0, W-1)
	c1 = clampInt(overviewCell(last, ww, W)+1, c0+1, W)
	return c0, c1
}

// overviewSpan is the town columns [a, b) the overview cell x stands for:
// the columns overviewCell puts there.
func overviewSpan(x, ww, W int) (a, b int) {
	a = (x*ww + W - 1) / W
	b = max(a+1, ((x+1)*ww+W-1)/W)
	return a, min(b, max(a+1, ww))
}

// ageWord is the shortest name that tells an age: "Bronze" for "Bronze Age".
func ageWord(name string) string { return strings.TrimSuffix(name, " Age") }

// minimap is the overview bar: the whole town in one row, its skyline
// as it stands (each column the height of the tallest building there in the
// colour of its top, in eighth blocks), the wonders marked, each district
// named where its name fits, buildings new since your last visit in the
// positive colour, and a bright box over the part of the town in view at
// the current zoom.
func (s *scene) minimap() {
	y := s.H - 2
	W := s.W
	lay := s.lay
	prof := lay.prof
	ww := lay.width
	bg := theme.Color(theme.RoleBackground)
	chip := theme.Color(theme.RoleChip)
	acc := role(theme.RoleAccent, bg)
	pos := role(theme.RolePositive, bg)
	// the box is the chip lifted toward the accent, so it reads in any theme
	lit := theme.Mix(chip, theme.Color(theme.RoleAccent), 0.3)
	c0, c1 := overviewBox(s.cam, s.lw, ww, W)
	s.ovBox = [2]int{c0, c1}
	maxH := 1
	for _, h := range prof {
		maxH = max(maxH, h)
	}
	isNew := make([]bool, W)
	if s.v.changes {
		for _, x := range lay.news {
			if c := overviewCell(x, ww, W); c >= 0 && c < W {
				isNew[c] = true
			}
		}
	}
	tops := make([]tcell.Color, W)
	for x := 0; x < W; x++ {
		a, b := overviewSpan(x, ww, W)
		h, lot, at := 0, int32(-1), 0
		for i := a; i < b && i < len(prof); i++ {
			if prof[i] > h {
				h, lot, at = prof[i], lay.tallest[i], i
			}
		}
		cbg := bg
		if x >= c0 && x < c1 {
			cbg = lit
		}
		fg := role(theme.RoleDim, cbg)
		if h > 0 && lot >= 0 {
			lv := &lay.lots[lot]
			sx := at - lv.x0
			if t := lv.spr.topAt(sx); t >= 0 {
				fg = theme.Legible(s.p.col(lv.fam, lv.vr, lv.spr.at(sx, t).fg, lv.ml.Row), cbg, 2)
			}
		}
		if isNew[x] {
			fg = theme.Legible(pos, cbg, 3)
		}
		tops[x] = fg
		lvl := clampInt(int(math.Ceil(float64(h)/float64(maxH)*8)), 0, 8)
		if h > 0 && lvl == 0 {
			lvl = 1
		}
		ch := eighths[lvl]
		if ch == '█' {
			ch = '▇'
		}
		s.fb.set(x, y, ch, fg, cbg, dTop)
	}
	cellBg := func(x int) tcell.Color {
		if x >= c0 && x < c1 {
			return lit
		}
		return bg
	}
	for _, d := range s.m.Skyline.Districts { // where each age begins
		if c := overviewCell(d.X0, ww, W); c > 0 && c < W {
			if cl := s.fb.at(c, y); cl.ch == ' ' {
				s.fb.set(c, y, '╷', theme.Legible(role(theme.RoleDim, cl.bg), cl.bg, 2), cl.bg, dTop)
			}
		}
	}
	wonder := make([]bool, W) // the wonders' columns, which a name keeps clear of
	for _, w := range lay.lots {
		if c := overviewCell(w.ml.X, ww, W); w.ml.Wonder && c >= 0 && c < W {
			wonder[c] = true
		}
	}
	named := make([]bool, W)
	for _, d := range s.m.Skyline.Districts { // each district's name, where it fits
		name := s.m.Catalog.AgeNames[d.Age]
		a, b := overviewCell(d.X0, ww, W), overviewCell(d.X0+d.W, ww, W)
		for _, t := range []string{name, ageWord(name)} {
			if x, ok := nameSpot(wonder, a, b, textLen(t)); ok {
				for i, r := range []rune(t) {
					cbg := theme.Mix(cellBg(x+i), tops[x+i], 0.3)
					s.fb.set(x+i, y, r, theme.Legible(theme.Color(theme.RoleText), cbg, 4.5), cbg, dTop)
					named[x+i] = true
				}
				break
			}
		}
	}
	for c, on := range wonder { // the wonders
		if on {
			cbg := cellBg(c)
			s.fb.set(c, y, '◆', theme.Legible(theme.Color(theme.RoleHighlight), cbg, 3), cbg, dTop)
		}
	}
	for _, c := range []struct {
		x  int
		ch rune
	}{{c0, '▕'}, {c1 - 1, '▏'}} { // the box's edges, where nothing else stands
		if cl := s.fb.at(c.x, y); cl != nil && !wonder[c.x] && !named[c.x] {
			s.fb.set(c.x, y, c.ch, theme.Legible(acc, cl.bg, 3), cl.bg, dTop)
		}
	}
}

// nameSpot is where a name n cells wide goes in the cells [a, b) of the
// overview with a cell of room each side and clear of the marked columns:
// the middle when it can, else the nearest spot to it.
func nameSpot(marked []bool, a, b, n int) (int, bool) {
	room := b - a - n - 2
	if room < 0 {
		return 0, false
	}
	mid := a + 1 + room/2
	for d := 0; d <= room; d++ {
		for _, x := range []int{mid - d, mid + d} {
			if x < a+1 || x > b-1-n {
				continue
			}
			clear := true
			for i := x - 1; i <= x+n && clear; i++ {
				clear = i < 0 || i >= len(marked) || !marked[i]
			}
			if clear {
				return x, true
			}
		}
	}
	return 0, false
}

func (s *scene) status(bg tcell.Color) {
	y := s.H - 1
	m := s.m
	if s.v.inspect {
		if in, ok := s.v.inspection(m, s.anim); ok {
			parts := []seg{{" ▲ ", theme.RoleAccent}, {in.title, theme.RoleText}}
			for _, l := range in.lines {
				parts = append(parts, seg{"  " + l, theme.RoleDim})
			}
			if in.cmd != "" {
				parts = append(parts, seg{"  type: ", theme.RoleDim}, seg{in.cmd, theme.RoleLabel})
			}
			s.segs(0, y, bg, parts)
			return
		}
	}
	if s.v.flows {
		w := m.Flows.Worst
		if w == "" {
			w = "nothing is stuck"
		}
		parts := []seg{{" flows: ", theme.RoleAccent}, {w, theme.RoleWarning}}
		if len(m.Flows.Full) > 0 {
			names := make([]string, 0, len(m.Flows.Full))
			for _, st := range m.Flows.Full {
				names = append(names, strings.ToLower(st.Name))
			}
			parts = append(parts, seg{"  full: ", theme.RoleDim}, seg{strings.Join(names, ", "), theme.RoleWarning})
		}
		parts = append(parts, seg{"  type: ", theme.RoleDim}, seg{m.IdleCommand(), theme.RoleLabel})
		s.segs(0, y, bg, parts)
		return
	}
	if s.v.legend {
		parts := []seg{{" lit windows ", theme.RoleText}, {"staffed  ", theme.RoleDim},
			{"smoke ", theme.RoleText}, {"producing  ", theme.RoleDim}, {"▼ ", theme.RolePositive}, {"new  ", theme.RoleDim},
			{"IDLE LOW FULL ", theme.RoleWarning}, {"flows  ", theme.RoleDim}, {"pennants ", theme.RoleText},
			{"your standing with each civ", theme.RoleDim}}
		if s.inCity { // the age's look, in a word or two (city.go)
			parts = append(parts, seg{"  │ ", theme.RoleDim}, seg{s.city.Name + ": ", theme.RoleText},
				seg{s.city.Structure, theme.RoleDim})
		}
		s.segs(0, y, bg, parts)
		return
	}
	where := "the wilds"
	if di := m.Skyline.DistrictAt(s.cam + s.lw/2); di >= 0 {
		where = strings.ToLower(m.Catalog.AgeNames[m.Skyline.Districts[di].Age]) + " district"
	}
	where = "zoom: " + s.zm.String() + "  " + where
	s.segs(0, y, bg, panoramaHints(s.W, where))
}
