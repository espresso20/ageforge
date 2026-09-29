package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
)

// chrome.go: the header, the minimap strip and the status line. These are
// UI, so they take their colours from theme roles only.

type roleKind int

const (
	roleText roleKind = iota
	roleDim
	roleAccent
	roleLabel
	roleHigh
	rolePositive
	roleWarn
	roleSurface
	roleChip
	roleBg
	roleNeg
)

func (s *scene) role(k roleKind) RGB {
	t := s.v.Theme
	var r theme.Role
	switch k {
	case roleText:
		r = theme.RoleText
	case roleDim:
		r = theme.RoleDim
	case roleAccent:
		r = theme.RoleAccent
	case roleLabel:
		r = theme.RoleLabel
	case roleHigh:
		r = theme.RoleHighlight
	case rolePositive:
		r = theme.RolePositive
	case roleWarn:
		r = theme.RoleWarning
	case roleSurface:
		r = theme.RoleSurface
	case roleChip:
		r = theme.RoleChip
	case roleNeg:
		r = theme.RoleNegative
	default:
		r = theme.RoleBackground
	}
	c := fromTC(t.Color(r))
	if s.v.Colors == 16 {
		c = nearestANSI(c)
	}
	return c
}

func clock(tod float64) string {
	m := int(tod*24*60) % (24 * 60)
	return fmt.Sprintf("%02d:%02d", m/60, m%60)
}

func humanN(n float64) string {
	switch {
	case n >= 1e12:
		return fmt.Sprintf("%.1fT", n/1e12)
	case n >= 1e9:
		return fmt.Sprintf("%.1fB", n/1e9)
	case n >= 1e6:
		return fmt.Sprintf("%.1fM", n/1e6)
	case n >= 1e4:
		return fmt.Sprintf("%.1fk", n/1e3)
	}
	return fmt.Sprintf("%.0f", n)
}

type seg struct {
	s string
	k roleKind
}

func (s *scene) segs(y int, bg RGB, parts []seg) int {
	x := 0
	for _, p := range parts {
		x = s.fb.text(x, y, p.s, s.role(p.k), bg)
	}
	return x
}

func (s *scene) chrome() {
	st := s.w.St
	W := s.v.W
	bg := s.role(roleSurface)
	if s.v.Theme.IsLight() {
		bg = s.role(roleChip)
	}
	for x := 0; x < W; x++ {
		s.fb.set(x, 0, ' ', bg, bg)
		s.fb.set(x, s.v.H-1, ' ', bg, bg)
	}
	tod := s.v.TOD
	if tod < 0 {
		tod = todFromTick(st.Tick)
	}
	moon := "☼"
	if s.p.Night > 0.5 {
		moon = "☾"
	}
	day := st.Tick/dayTicks + 1
	total := 0
	for _, b := range st.Buildings {
		total += b.Count
	}
	// the BBS header bar: shade ramp, the age, then the facts
	acc := s.role(roleAccent)
	x := 0
	for i, r := range []rune("░▒▓█") {
		s.fb.set(x+i, 0, r, acc, bg)
	}
	x = s.fb.text(4, 0, " "+strings.ToUpper(st.AgeName)+" ", s.role(roleBg), acc)
	for i, r := range []rune("█▓▒░") {
		s.fb.set(x+i, 0, r, acc, bg)
	}
	x += 5
	parts := []seg{
		{config.EpochByKey()[config.EpochForAge(st.Age)].Name, roleLabel},
		{"  day ", roleDim}, {fmt.Sprint(day), roleHigh},
		{"  " + clock(tod) + " " + moon, roleText},
		{"  pop ", roleDim}, {fmt.Sprintf("%d/%d", st.Workers.TotalPop, st.Workers.MaxPop), roleHigh},
		{"  bld ", roleDim}, {fmt.Sprint(total), roleHigh},
	}
	if s.v.ShowNew && s.w.Prev != nil && s.w.NewCount > 0 {
		parts = append(parts, seg{"  ▼ ", rolePositive}, seg{fmt.Sprintf("%d new since your last visit", s.w.NewCount), rolePositive})
	}
	if n := idleProducers(s.w); n > 0 {
		// dark districts are idle ones; say how many types sit unstaffed
		parts = append(parts, seg{"  ▪ ", roleWarn}, seg{fmt.Sprintf("%d idle", n), roleWarn})
	}
	if h := st.Harbinger; h != nil {
		parts = append(parts, seg{"  ☻ ", roleWarn}, seg{h.Name + " is on the ridge", roleWarn})
	}
	if st.PendingCatastrophe != "" {
		name, _ := config.CatastropheInfo(st.PendingCatastrophe)
		parts = append(parts, seg{"  ▲ " + name, roleNeg})
	}
	for _, p := range parts {
		x = s.fb.text(x, 0, p.s, s.role(p.k), bg)
		if x >= W {
			break
		}
	}
	s.minimap()
	s.status()
}

// minimap is the whole panorama in one row: the skyline profile in eighth
// blocks, the viewport bracketed, new buildings in the positive colour.
func (s *scene) minimap() {
	y := s.v.H - 2
	W := s.v.W
	prof := s.profileCache()
	bg := s.role(roleBg)
	dim := s.role(roleDim)
	acc := s.role(roleAccent)
	vp0 := float64(s.v.Cam) / float64(s.w.W) * float64(W)
	vp1 := float64(s.v.Cam+s.v.W) / float64(s.w.W) * float64(W)
	newCol := make([]bool, W)
	if s.v.ShowNew {
		for _, l := range s.w.Lots {
			if l.New {
				c := int(float64(l.X+l.Spr.W/2) / float64(s.w.W) * float64(W))
				if c >= 0 && c < W {
					newCol[c] = true
				}
			}
		}
	}
	maxH := 1
	for _, h := range prof {
		maxH = max(maxH, h)
	}
	blocks := []rune(" ▁▂▃▄▅▆▇█")
	for x := 0; x < W; x++ {
		a := int(float64(x) / float64(W) * float64(s.w.W))
		b := int(float64(x+1) / float64(W) * float64(s.w.W))
		h := 0
		for i := a; i < b && i < len(prof); i++ {
			h = max(h, prof[i])
		}
		lvl := int(math.Ceil(float64(h) / float64(maxH) * 8))
		inView := float64(x) >= vp0 && float64(x) < vp1
		cbg := bg
		if inView {
			cbg = s.role(roleChip)
		}
		fg := dim
		if inView {
			fg = s.role(roleText)
		}
		if newCol[x] {
			fg = s.role(rolePositive)
		}
		s.fb.set(x, y, blocks[lvl], fg, cbg)
	}
	// district ticks: the age boundaries along the bottom edge
	for _, d := range s.w.Districts {
		c := int(float64(d.X0) / float64(s.w.W) * float64(W))
		if c > 0 && c < W && s.fb.at(c, y).Ch == ' ' {
			s.fb.fg(c, y, '╷', dim)
		}
	}
	s.fb.fg(int(vp0), y, '▕', acc)
	s.fb.fg(min(W-1, int(vp1)), y, '▏', acc)
}

func (s *scene) status() {
	y := s.v.H - 1
	bg := s.role(roleSurface)
	if s.v.Theme.IsLight() {
		bg = s.role(roleChip)
	}
	if l := s.selected(); l != nil {
		st := s.w.St
		bs := st.Buildings[l.Key]
		parts := []seg{{" ▲ ", roleAccent}, {l.Name, roleText}, {" ×" + fmt.Sprint(bs.Count), roleHigh},
			{"  " + lineageName(l.Def), roleDim}}
		if bs.WorkerCapacity > 0 {
			parts = append(parts, seg{"  workers ", roleDim},
				seg{fmt.Sprintf("%d/%d", bs.WorkersAssigned, bs.Count*bs.WorkerCapacity), roleHigh})
			if bs.WorkersAssigned == 0 {
				parts = append(parts, seg{" idle", roleWarn})
			}
		}
		if l.New {
			parts = append(parts, seg{"  new", rolePositive})
		}
		parts = append(parts, seg{"  › ", roleDim}, seg{"build " + l.Key, roleLabel})
		if c := mainCost(bs.NextCost); c != "" {
			parts = append(parts, seg{" " + c, roleDim})
		}
		if l.Wonder {
			parts = parts[:len(parts)-1]
			parts = append(parts, seg{"  wonder of the " + config.AgeOrder()[l.Age], roleDim})
		}
		s.segs(y, bg, parts)
		return
	}
	di := s.w.DistrictAt(s.v.Cam + s.v.W/2)
	where := "the wilds"
	if di >= 0 {
		where = config.AgeOrder()[s.w.Districts[di].Age]
		where = strings.ReplaceAll(where, "_", " ") + " quarter"
	}
	s.segs(y, bg, []seg{{" ◄► ", roleAccent}, {"scroll  ", roleDim}, {"[ ] ", roleAccent}, {"age  ", roleDim},
		{"i ", roleAccent}, {"inspect  ", roleDim}, {"n ", roleAccent}, {"time  ", roleDim},
		{"w ", roleAccent}, {"weather  ", roleDim}, {"t ", roleAccent}, {"theme  ", roleDim},
		{"c ", roleAccent}, {"changes  ", roleDim}, {"│ ", roleDim}, {where, roleLabel}})
}

func lineageName(d config.BuildingDef) string {
	l := d.LineageKey
	if l == "" {
		l = d.Category
	}
	return strings.ReplaceAll(l, "_", " ")
}

func mainCost(cost map[string]float64) string {
	best, bv := "", 0.0
	for k, v := range cost {
		if v > bv {
			best, bv = k, v
		}
	}
	if best == "" {
		return ""
	}
	return humanN(bv) + " " + best
}
