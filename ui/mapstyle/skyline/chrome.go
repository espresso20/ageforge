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

// minimap is the whole panorama in one row: the skyline profile in eighth
// blocks, the viewport bracketed, age boundaries ticked and new buildings
// in the positive colour.
func (s *scene) minimap() {
	y := s.H - 2
	W := s.W
	prof := s.lay.prof
	ww := s.lay.width
	bg := theme.Color(theme.RoleBackground)
	chip := theme.Color(theme.RoleChip)
	dim, txt := role(theme.RoleDim, bg), role(theme.RoleText, chip)
	acc := role(theme.RoleAccent, bg)
	pos := role(theme.RolePositive, bg)
	vp0 := float64(s.cam) / float64(ww) * float64(W)
	vp1 := float64(s.cam+W) / float64(ww) * float64(W)
	maxH := 1
	for _, h := range prof {
		maxH = max(maxH, h)
	}
	isNew := make([]bool, W)
	if s.v.changes {
		for _, x := range s.lay.news {
			if c := int(float64(x) / float64(ww) * float64(W)); c >= 0 && c < W {
				isNew[c] = true
			}
		}
	}
	for x := 0; x < W; x++ {
		a, b := x*ww/W, (x+1)*ww/W
		h := 0
		for i := a; i < b && i < len(prof); i++ {
			h = max(h, prof[i])
		}
		lvl := int(math.Ceil(float64(h) / float64(maxH) * 8))
		in := float64(x) >= vp0 && float64(x) < vp1
		cbg, fg := bg, dim
		if in {
			cbg, fg = chip, txt
		}
		if isNew[x] {
			fg = theme.Legible(pos, cbg, 3)
		}
		ch := eighths[clampInt(lvl, 0, 8)]
		if ch == '█' {
			ch = '▇'
		}
		s.fb.set(x, y, ch, fg, cbg, dTop)
	}
	for _, d := range s.m.Skyline.Districts {
		if c := d.X0 * W / ww; c > 0 && c < W {
			if cl := s.fb.at(c, y); cl.ch == ' ' {
				s.fb.set(c, y, '╷', theme.Legible(dim, cl.bg, 2), cl.bg, dTop)
			}
		}
	}
	for _, c := range []struct {
		x  int
		ch rune
	}{{int(vp0), '▕'}, {min(W-1, int(vp1)), '▏'}} {
		if cl := s.fb.at(clampInt(c.x, 0, W-1), y); cl != nil {
			s.fb.set(clampInt(c.x, 0, W-1), y, c.ch, theme.Legible(acc, cl.bg, 3), cl.bg, dTop)
		}
	}
}

func (s *scene) status(bg tcell.Color) {
	y := s.H - 1
	m := s.m
	if s.v.inspect {
		if in, ok := s.v.inspection(m); ok {
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
		s.segs(0, y, bg, []seg{{" lit windows ", theme.RoleText}, {"staffed  ", theme.RoleDim},
			{"smoke ", theme.RoleText}, {"producing  ", theme.RoleDim}, {"▼ ", theme.RolePositive}, {"new  ", theme.RoleDim},
			{"IDLE LOW FULL ", theme.RoleWarning}, {"flows  ", theme.RoleDim}, {"pennants ", theme.RoleText},
			{"your standing with each civ", theme.RoleDim}})
		return
	}
	where := "the wilds"
	if di := m.Skyline.DistrictAt(s.cam + s.W/2); di >= 0 {
		where = strings.ToLower(m.Catalog.AgeNames[m.Skyline.Districts[di].Age]) + " district"
	}
	s.segs(0, y, bg, []seg{{" ◄► ", theme.RoleAccent}, {"scroll  ", theme.RoleDim}, {"PgUp PgDn ", theme.RoleAccent},
		{"half  ", theme.RoleDim}, {"Home End ", theme.RoleAccent}, {"oldest, present  ", theme.RoleDim},
		{"Tab ", theme.RoleAccent}, {"inspect  ", theme.RoleDim}, {"map flows ", theme.RoleAccent}, {"flows  ", theme.RoleDim},
		{"│ ", theme.RoleDim}, {where, theme.RoleLabel}})
}
