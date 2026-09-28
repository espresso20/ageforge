package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The verb. Select any node and the inspector says what it is made of and
// which commands act on it, in the game's own command vocabulary. Pressing
// 1-3 stages that command in the input line (in the game: pre-fills the
// command input; the player still presses Enter).

type iline struct {
	text string
	k    cls
	bold bool
}

func (v *view) inspectLines(n *node) (lines []iline, cmds []string) {
	st := v.m.st
	add := func(k cls, f string, a ...any) { lines = append(lines, iline{fmt.Sprintf(f, a...), k, false}) }
	head := func(s string) { lines = append(lines, iline{s, cAccent, true}) }
	switch {
	case n.kind == kProducer:
		lin := strings.TrimPrefix(n.id, "prod:")
		head(n.title + " lineage")
		type row struct {
			key   string
			name  string
			count int
			st    int
			cap   int
			tier  int
			leg   bool
			up    string
		}
		var rows []row
		for key, b := range st.Buildings {
			d, ok := defs()[key]
			if !ok || lineageOf(d) != lin || b.Count == 0 {
				continue
			}
			rows = append(rows, row{key, b.Name, b.Count, b.WorkersAssigned, b.Count * b.WorkerCapacity, d.LineageTier, b.IsLegacy, b.PendingUpgrade})
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].tier != rows[j].tier {
				return rows[i].tier > rows[j].tier
			}
			return rows[i].key < rows[j].key
		})
		for _, r := range rows {
			staff := ""
			if r.cap > 0 {
				staff = fmt.Sprintf("  %d/%d staffed", r.st, r.cap)
			}
			k := cText
			if r.leg {
				k = cDim
			}
			add(k, "%3d× %s%s", r.count, r.name, staff)
		}
		for res, val := range lineageFlows(st)[lin] {
			if res == "@pop" {
				add(cLabel, "  gives %d homes", int(val))
				continue
			}
			add(cLabel, "  → %-10s %s", res, signed(val))
		}
		idle := st.Workers.TotalIdle
		upgraded := false
		for _, r := range rows {
			if r.up != "" && !upgraded {
				cmds = append(cmds, "upgrade "+r.key)
				upgraded = true
			}
			if r.cap > r.st && idle > 0 {
				cmds = append(cmds, fmt.Sprintf("assign %s %d", r.key, min(idle, r.cap-r.st)))
				idle = 0
			}
		}
		for _, r := range rows {
			if !r.leg {
				cmds = append(cmds, "build "+r.key)
				break
			}
		}
	case n.kind == kPort && n.id == "port:@pop":
		w := st.Workers
		head("Population")
		add(cText, "%d of %d housed, %d idle", w.TotalPop, w.MaxPop, w.TotalIdle)
		add(cLabel, "eats %s food", signed(w.FoodDrain))
		if w.MaxPop > w.TotalPop {
			cmds = append(cmds, "recruit max")
		}
		cmds = append(cmds, "workers")
	case n.kind == kPort:
		res := strings.TrimPrefix(n.id, "port:")
		r := st.Resources[res]
		head(fmt.Sprintf("%s store", r.Name))
		add(cText, "%s of %s  (%.0f%%)", short(r.Amount), short(r.Storage), 100*frac(r.Amount, r.Storage))
		b := r.Breakdown
		for _, p := range []struct {
			l string
			v float64
		}{{"buildings", b.BuildingRate}, {"workers", b.WorkerRate}, {"research", b.ResearchRate}, {"events", b.EventRate}, {"trade", b.TradeRate}, {"bonus", b.BonusRate}, {"eaten", -b.FoodDrain}} {
			if math.Abs(p.v) > 1e-9 {
				k := cPos
				if p.v < 0 {
					k = cNeg
				}
				add(k, "  %-10s %s", p.l, signed(p.v))
			}
		}
		switch {
		case r.Rate > 0 && r.Storage > r.Amount:
			add(cDim, "full in %s", eta((r.Storage-r.Amount)/r.Rate))
		case r.Rate < 0 && r.Amount > 0:
			add(cNeg, "empty in %s", eta(r.Amount/-r.Rate))
		}
		if n.alert == "FULL" {
			if st.WonderOverflow && st.CurrentAgeWonderKey != "" {
				add(cHi, "overflow banks into the wonder")
			} else {
				add(cWarn, "production over the cap is lost")
			}
		}
		if k := storageFor(st, res); k != "" {
			cmds = append(cmds, "build "+k)
		}
		if n.alert == "FULL" {
			cmds = append(cmds, fmt.Sprintf("trade %s <to> %s", res, short(r.Amount/4)))
		}
		if b, ok := st.Buildings[st.CurrentAgeWonderKey]; ok && b.WonderBank[res] < b.NextCost[res] {
			cmds = append(cmds, "wonder collect "+res)
		}
	case n.id == "svc:pop":
		head("Citizens")
		ws := st.Workers
		for _, k := range sortedMapKeys(ws.Types) {
			t := ws.Types[k]
			if t.Count == 0 {
				continue
			}
			add(cText, "  %-12s %3d  %d idle", t.Name, t.Count, t.IdleCount)
		}
		cmds = append(cmds, "workers")
	case n.id == "svc:lab":
		head("Research")
		rs := st.Research
		if rs.CurrentTech != "" {
			add(cText, "%s, %ds left", rs.CurrentTechName, rs.TicksLeft*2)
		}
		avail := 0
		for _, t := range rs.Techs {
			if t.Available {
				avail++
			}
		}
		add(cLabel, "%d researched, %d available", rs.TotalResearched, avail)
		cmds = append(cmds, "research list")
	case n.id == "svc:forge":
		head("Construction queue")
		for i, q := range st.BuildQueue {
			if i == 8 {
				add(cDim, "  … %d more", len(st.BuildQueue)-8)
				break
			}
			add(cText, "  %-22s %4ds", clip(q.Name, 22), q.TicksLeft*2)
		}
		cmds = append(cmds, "plan")
	case n.id == "svc:wonder":
		b := st.Buildings[st.CurrentAgeWonderKey]
		head(b.Name)
		for _, res := range sortedMapKeys(b.NextCost) {
			k := cText
			if b.WonderBank[res] >= b.NextCost[res] {
				k = cPos
			}
			add(k, "  %-10s %s / %s", res, short(b.WonderBank[res]), short(b.NextCost[res]))
		}
		if b.WonderBankFull {
			cmds = append(cmds, "build "+st.CurrentAgeWonderKey)
		}
		cmds = append(cmds, "wonder collect all", "wonder")
	case n.id == "svc:garrison":
		mil := st.Military
		head("Garrison")
		add(cText, "%d soldiers of %d, defence %.0f", mil.SoldierCount, mil.SoldierCap, mil.DefenseRating)
		cmds = append(cmds, "expedition list", "campaign list")
	case n.id == "svc:market":
		head("Market")
		for _, r := range st.Trade.ActiveRoutes {
			k := cText
			if r.Disrupted {
				k = cNeg
			}
			add(k, "  ≡ %s  ×%d", r.Name, r.CyclesDone)
		}
		for _, r := range st.Trade.AvailableRoutes {
			if r.CanStart {
				cmds = append(cmds, "trade route start "+r.Key)
				break
			}
		}
		cmds = append(cmds, "trade list")
	case strings.HasPrefix(n.id, "host:civ:"):
		key := strings.TrimPrefix(n.id, "host:civ:")
		f := st.Diplomacy.Factions[key]
		head(f.Name)
		add(cText, "%s, opinion %+d, strength %d", f.Status, f.Opinion, f.Strength)
		add(cLabel, "%s · specialty %s", f.Personality, f.Specialty)
		if f.AtWar {
			add(cNeg, "AT WAR: routes importing its goods are cut")
			cmds = append(cmds, "diplomacy tribute "+key)
		}
		if f.Opinion >= 50 && f.Status != "allied" {
			cmds = append(cmds, "diplomacy ally "+key)
		}
		cmds = append(cmds, "diplomacy gift "+key, "diplomacy deals "+key)
	case strings.HasPrefix(n.id, "host:route:"):
		key := strings.TrimPrefix(n.id, "host:route:")
		head(n.title)
		add(cText, "%s", n.sub)
		cmds = append(cmds, "trade route stop "+key)
	case n.id == "host:next":
		head("Next: " + st.NextAgeName)
		for _, res := range sortedMapKeys(st.NextAgeResReqs) {
			need := st.NextAgeResReqs[res]
			have := st.Resources[res].Amount
			k := cText
			if have >= need {
				k = cPos
			}
			add(k, "  %-10s %s / %s", res, short(have), short(need))
		}
		for _, key := range sortedMapKeys(st.NextAgeBldReqs) {
			need := st.NextAgeBldReqs[key]
			have := st.Buildings[key].Count
			k := cText
			if have >= need {
				k = cPos
			}
			add(k, "  %-18s %d / %d", clip(st.Buildings[key].Name, 18), have, need)
		}
		if st.AgeReady {
			cmds = append(cmds, "advance")
		} else {
			cmds = append(cmds, "plan advance")
		}
	default:
		head(n.title)
		if n.sub != "" {
			add(cText, "%s", n.sub)
		}
	}
	if len(cmds) > 3 {
		cmds = cmds[:3]
	}
	return lines, cmds
}

func eta(ticks float64) string {
	s := ticks * 2
	switch {
	case s < 90:
		return fmt.Sprintf("%.0fs", s)
	case s < 5400:
		return fmt.Sprintf("%.0fm", s/60)
	default:
		return fmt.Sprintf("%.1fh", s/3600)
	}
}

// storageFor is the newest buildable storage building that raises res.
func storageFor(st game.GameState, res string) string {
	best, bestTier := "", -1
	for key, b := range st.Buildings {
		d, ok := defs()[key]
		if !ok || !b.Unlocked || b.IsLegacy || b.AtMaxCount {
			continue
		}
		for _, e := range d.Effects {
			if e.Type == "storage" && (e.Target == res || e.Target == "all") {
				t := config.AgeByKey()[d.RequiredAge].Order
				if t > bestTier || (t == bestTier && key < best) {
					best, bestTier = key, t
				}
			}
		}
	}
	return best
}

func (v *view) drawInspector() {
	n := v.m.nodes[v.sel]
	if n == nil {
		return
	}
	lines, cmds := v.inspectLines(n)
	c := v.c
	w := min(48, c.w/2)
	h := len(lines) + 2
	if len(cmds) > 0 {
		h += len(cmds) + 1
	}
	h = min(h, c.h-3)
	x := c.w - w - 1
	if n.x+n.w/2 > c.w/2 {
		x = 1
	}
	y := v.L.top
	v.box(x, y, w, h, cAccent, v.s.boxes)
	c.bgRect(x, y, w, h, bgChip)
	v.title(&node{title: "INSPECT"}, x, y, w, cAccent)
	yy := y + 1
	for _, l := range lines {
		if yy >= y+h-1 {
			break
		}
		c.textB(x+2, yy, clip(l.text, w-4), l.k, w-4, l.bold)
		yy++
	}
	if len(cmds) > 0 && yy < y+h-2 {
		yy++
		for i, cmd := range cmds {
			if yy >= y+h-1 {
				break
			}
			c.textB(x+2, yy, fmt.Sprintf("%d", i+1), cHi, 1, true)
			c.text(x+4, yy, clip("› "+cmd, w-6), cText, w-6)
			yy++
		}
	}
}
