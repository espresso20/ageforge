package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The topology model: a four-column system diagram built from one
// game.GameState. Column 0 holds producers (one node per building lineage),
// column 1 the resource "bus" (one port per resource, drawn as a single
// chip), column 2 the city's services (population, lab, forge, wonder,
// garrison, market), column 3 the external hosts (civilizations, trade
// partners, expeditions, the next age).

type kind int

const (
	kProducer kind = iota
	kPort
	kService
	kHost
)

// severity of a node or edge signal.
type severity int

const (
	sevOK severity = iota
	sevInfo
	sevWarn
	sevCrit
)

type node struct {
	id    string
	kind  kind
	title string // box title / port label
	sub   string // secondary label (top building, tech name, ...)
	badge string // right side of the title: count, status
	value float64
	mass  float64 // layout weight: building count for producers

	gauge      float64 // 0..1, <0 = none
	gaugeLabel string
	spark      []float64
	stat       string // numeric readout, e.g. "+4.2/s"
	sev        severity
	alert      string // short alert tag, e.g. "FULL"

	// Port specifics.
	amount, storage, rate float64

	// Layout.
	x, y, w, h int
	anchor     map[string]int // edge id -> row offset of its pin
	hidden     bool
}

type edgeState int

const (
	eFlow     edgeState = iota // normal flow
	eBackflow                  // target full: flow piles up
	eStarve                    // draining source
	eIdle                      // link up, nothing moving
	eSevered                   // war / blockade
)

type edge struct {
	id       string
	from, to string
	label    string
	rate     float64 // units per tick (sign ignored; direction is from->to)
	state    edgeState
	// path is the routed polyline, cells from source pin to target pin.
	path []pt
}

type pt struct{ x, y int }

type model struct {
	st     game.GameState
	fx     *Fixture
	nodes  map[string]*node
	cols   [4][]*node
	edges  []*edge
	alarms []string // header ticker, most severe first
	catast string   // pending catastrophe name, "" if none
}

var lineageLabel = map[string]string{
	"food":                  "Food",
	"organic_extraction":    "Forestry",
	"geological_extraction": "Quarry & Mine",
	"metallurgy":            "Metallurgy",
	"knowledge":             "Scholarship",
	"faith":                 "Faith",
	"culture_arts":          "Culture",
	"energy":                "Power",
	"engineering":           "Engineering",
	"trade":                 "Commerce",
	"harbor":                "Harbor",
	"hacker":                "Netrunners",
	"housing":               "Housing",
	"military":              "Military",
	"monument":              "Monuments",
	"astronaut":             "Astronauts",
}

// lineageOf returns the display lineage of a building definition, "" for
// buildings that are not flow producers (storage, wonders).
func lineageOf(d config.BuildingDef) string {
	switch d.LineageKey {
	case "storage", "wonder", "":
		if d.Category == "housing" {
			return "housing"
		}
		return ""
	}
	return d.LineageKey
}

// lineageOutput attributes each lineage's nominal output per resource, using
// the engine's worker-fill formula: base × count × (0.20 + 0.80 × fill),
// ruins at half base with no workers. Keyed "lineage" → total.
func lineageOutput(st game.GameState) map[string]float64 {
	out := map[string]float64{}
	for lin, byRes := range lineageFlows(st) {
		for _, v := range byRes {
			out[lin] += v
		}
	}
	return out
}

var defsCache map[string]config.BuildingDef

func defs() map[string]config.BuildingDef {
	if defsCache == nil {
		defsCache = config.BuildingByKey()
	}
	return defsCache
}

// lineageFlows is lineage → resource → nominal units/tick.
func lineageFlows(st game.GameState) map[string]map[string]float64 {
	out := map[string]map[string]float64{}
	for key, b := range st.Buildings {
		if b.Count == 0 && b.RuinCount == 0 {
			continue
		}
		d, ok := defs()[key]
		if !ok {
			continue
		}
		lin := lineageOf(d)
		if lin == "" {
			continue
		}
		fill := 1.0
		if cap := b.Count * b.WorkerCapacity; cap > 0 {
			fill = math.Min(1, float64(b.WorkersAssigned)/float64(cap))
		}
		mult := float64(b.Count)*(0.20+0.80*fill) + 0.5*float64(b.RuinCount)
		for _, e := range d.Effects {
			switch e.Type {
			case "production":
				if out[lin] == nil {
					out[lin] = map[string]float64{}
				}
				out[lin][e.Target] += e.Value * mult
			case "capacity":
				if e.Target == "population" {
					if out[lin] == nil {
						out[lin] = map[string]float64{}
					}
					out[lin]["@pop"] += e.Value * float64(b.Count)
				}
			}
		}
	}
	return out
}

func buildModel(fx *Fixture) *model {
	st := fx.State
	m := &model{st: st, fx: fx, nodes: map[string]*node{}}
	add := func(n *node) *node {
		if n.gauge == 0 && n.gaugeLabel == "" {
			n.gauge = -1
		}
		m.nodes[n.id] = n
		m.cols[n.kind] = append(m.cols[n.kind], n)
		return n
	}
	flows := lineageFlows(st)

	// ---- ports (column 1): population pseudo-port first, then resources in
	// the game's canonical order so the bus reads the same every session.
	disrupted := map[string]bool{}
	for _, r := range st.Trade.DisruptedResources {
		disrupted[r] = true
	}
	w := st.Workers
	pop := add(&node{id: "port:@pop", kind: kPort, title: "pop", amount: float64(w.TotalPop), storage: float64(w.MaxPop)})
	pop.gauge = frac(pop.amount, pop.storage)
	pop.stat = short(float64(w.TotalPop)) + "/" + short(float64(w.MaxPop))
	if w.TotalIdle > 0 {
		pop.alert, pop.sev = "IDLE", sevWarn
	} else if w.TotalPop >= w.MaxPop && w.MaxPop > 0 {
		pop.alert, pop.sev = "HOUSED", sevInfo
	}
	for _, rd := range config.BaseResources() {
		r, ok := st.Resources[rd.Key]
		if !ok || !r.Unlocked {
			continue
		}
		if r.Amount < 0.5 && math.Abs(r.Rate) < 1e-6 {
			continue
		}
		n := add(&node{id: "port:" + rd.Key, kind: kPort, title: rd.Key, amount: r.Amount, storage: r.Storage, rate: r.Rate})
		n.gauge = frac(r.Amount, r.Storage)
		n.spark = fx.Hist["fill:"+rd.Key]
		n.stat = signed(r.Rate)
		switch {
		case disrupted[rd.Key]:
			n.alert, n.sev = "CUT", sevCrit
		case r.Rate < -1e-9 && r.Amount < -r.Rate*60:
			n.alert, n.sev = "EMPTY", sevCrit
		case r.Rate < -1e-9:
			n.alert, n.sev = "DRAIN", sevWarn
		case r.Storage > 0 && r.Amount >= r.Storage*0.985 && r.Rate > 0:
			n.alert, n.sev = "FULL", sevWarn
		}
	}

	// ---- producers (column 0): one node per lineage with output.
	var lins []string
	feeders := map[string]int{}
	for lin, byRes := range flows {
		lins = append(lins, lin)
		for res, v := range byRes {
			if v > 0 {
				feeders[res]++
			}
		}
	}
	sort.Strings(lins)
	for _, lin := range lins {
		byRes := flows[lin]
		n := &node{id: "prod:" + lin, kind: kProducer, title: lineageLabel[lin]}
		if n.title == "" {
			n.title = strings.ToUpper(lin[:1]) + lin[1:]
		}
		count, staffed, capa := 0, 0, 0
		topTier, top := -1, ""
		for key, b := range st.Buildings {
			d, ok := defs()[key]
			if !ok || lineageOf(d) != lin || (b.Count == 0 && b.RuinCount == 0) {
				continue
			}
			count += b.Count
			staffed += b.WorkersAssigned
			capa += b.Count * b.WorkerCapacity
			if !b.IsLegacy && b.Count > 0 && (d.LineageTier > topTier || (d.LineageTier == topTier && b.Name < top)) {
				topTier, top = d.LineageTier, b.Name
			}
		}
		if count == 0 {
			continue
		}
		n.sub = top
		n.badge = fmt.Sprintf("×%d", count)
		n.mass = float64(count)
		n.spark = fx.Hist["lin:"+lin]
		for res, v := range byRes {
			if res != "@pop" {
				n.value += v
			}
		}
		if capa > 0 {
			n.gauge = float64(staffed) / float64(capa)
			n.gaugeLabel = fmt.Sprintf("%d/%d", staffed, capa)
			if n.gauge < 0.5 {
				n.alert, n.sev = "UNDERSTAFFED", sevWarn
			}
		}
		if byRes["@pop"] > 0 && n.value == 0 {
			n.value = byRes["@pop"] / 20 // housing sized by the homes it gives
			n.stat = "+" + short(byRes["@pop"]) + " pop"
		} else {
			n.stat = signed(n.value)
		}
		add(n)
		for _, res := range sortedMapKeys(byRes) {
			v := byRes[res]
			to := "port:" + res
			if _, ok := m.nodes[to]; !ok || v <= 0 {
				continue
			}
			// Side outputs (under a tenth of the lineage's output, into a
			// store someone else also feeds) stay in the inspector: every
			// wire drawn must be worth reading.
			if res != "@pop" && v < 0.1*n.value && feeders[res] > 1 {
				continue
			}
			e := &edge{id: n.id + ">" + to, from: n.id, to: to, rate: v, label: res}
			tp := m.nodes[to]
			switch {
			case tp.sev == sevWarn && tp.alert == "FULL":
				e.state = eBackflow
			case res == "@pop":
				e.state = eFlow
				e.rate = v / 50
			}
			m.edges = append(m.edges, e)
		}
	}

	// ---- services (column 2).
	link := func(from, to string, rate float64, st edgeState) {
		if m.nodes[from] == nil || m.nodes[to] == nil {
			return
		}
		m.edges = append(m.edges, &edge{id: from + ">" + to, from: from, to: to, rate: math.Abs(rate), state: st})
	}

	if r, ok := st.Resources["food"]; ok {
		n := add(&node{id: "svc:pop", kind: kService, title: "Citizens", badge: fmt.Sprintf("%d", w.TotalPop)})
		n.gauge = frac(float64(w.TotalPop-w.TotalIdle), float64(w.TotalPop))
		n.gaugeLabel = fmt.Sprintf("%d working", w.TotalPop-w.TotalIdle)
		n.stat = fmt.Sprintf("eats %s", short(w.FoodDrain))
		n.sub = domainsLine(w)
		if w.TotalIdle > 0 {
			n.alert, n.sev = fmt.Sprintf("%d IDLE", w.TotalIdle), sevWarn
		}
		if r.Amount <= 1 && r.Rate < 0 {
			n.alert, n.sev = "STARVING", sevCrit
		}
		link("port:food", n.id, w.FoodDrain, eFlow)
	}
	{
		rs := st.Research
		n := add(&node{id: "svc:lab", kind: kService, title: "Research", badge: fmt.Sprintf("%d done", rs.TotalResearched)})
		if rs.CurrentTech != "" && rs.TotalTicks > 0 {
			n.sub = rs.CurrentTechName
			n.gauge = 1 - float64(rs.TicksLeft)/float64(rs.TotalTicks)
			n.gaugeLabel = fmt.Sprintf("%ds", rs.TicksLeft*2)
			link("port:knowledge", n.id, st.Resources["knowledge"].Rate, eFlow)
		} else {
			n.sub = "no tech queued"
			n.alert, n.sev = "IDLE", sevInfo
			link("port:knowledge", n.id, 0, eIdle)
		}
	}
	if len(st.BuildQueue) > 0 {
		q := st.BuildQueue[0]
		n := add(&node{id: "svc:forge", kind: kService, title: "Construction", badge: fmt.Sprintf("%d queued", len(st.BuildQueue))})
		n.sub = q.Name
		if q.TotalTicks > 0 {
			n.gauge = 1 - float64(q.TicksLeft)/float64(q.TotalTicks)
			n.gaugeLabel = fmt.Sprintf("%ds", q.TicksLeft*2)
		}
	}
	if wk := st.CurrentAgeWonderKey; wk != "" {
		b := st.Buildings[wk]
		n := add(&node{id: "svc:wonder", kind: kService, title: "Wonder", sub: b.Name})
		have, need := 0.0, 0.0
		for res, c := range b.NextCost {
			need += c
			have += math.Min(c, b.WonderBank[res])
		}
		n.gauge = frac(have, need)
		n.gaugeLabel = fmt.Sprintf("%.0f%% banked", 100*n.gauge)
		if b.WonderBankFull {
			n.alert, n.sev = "READY", sevInfo
		}
		// The wonder draws on the resources it still needs.
		var needs []string
		for res, c := range b.NextCost {
			if b.WonderBank[res] < c {
				needs = append(needs, res)
			}
		}
		sort.Strings(needs)
		for _, res := range needs {
			link("port:"+res, n.id, st.Resources[res].Rate*0.25, eFlow)
		}
	}
	mil := st.Military
	if mil.SoldierCap > 0 || mil.SoldierCount > 0 {
		n := add(&node{id: "svc:garrison", kind: kService, title: "Garrison", badge: "def " + short(mil.DefenseRating)})
		n.gauge = frac(float64(mil.SoldierCount), float64(mil.SoldierCap))
		n.gaugeLabel = short(float64(mil.SoldierCount)) + "/" + short(float64(mil.SoldierCap))
		n.sub = fmt.Sprintf("%d expeditions won", mil.CompletedCount)
		link("port:soldiers", n.id, mil.SoldierRate, eFlow)
	}
	tr := st.Trade
	if len(tr.ActiveRoutes) > 0 || len(tr.AvailableRoutes) > 0 || anyDiscovered(st) {
		n := add(&node{id: "svc:market", kind: kService, title: "Market", badge: fmt.Sprintf("%d routes", len(tr.ActiveRoutes))})
		exported := map[string]float64{}
		imported := map[string]float64{}
		bad := 0
		for _, r := range tr.ActiveRoutes {
			for k, v := range r.Export {
				exported[k] += v
			}
			for k, v := range r.Import {
				imported[k] += v
			}
			if r.Disrupted {
				bad++
			}
		}
		switch open := countStartable(tr.AvailableRoutes); {
		case len(tr.ActiveRoutes) == 0 && open > 0:
			n.alert, n.sev = fmt.Sprintf("%d ROUTES UNUSED", open), sevInfo
		case len(tr.ActiveRoutes) == 0:
			n.sub = "no routes running"
		case open > 0:
			n.sub = fmt.Sprintf("%d more could start", open)
		}
		if bad > 0 {
			n.alert, n.sev = fmt.Sprintf("%d CUT", bad), sevCrit
		}
		for _, res := range sortedMapKeys(exported) {
			link("port:"+res, n.id, exported[res]/20, eFlow)
		}
		for _, res := range sortedMapKeys(imported) {
			s := eFlow
			if disrupted[res] {
				s = eSevered
			}
			link(n.id, "port:"+res, imported[res]/20, s)
		}
	}

	// ---- hosts (column 3).
	if st.NextAge != "" {
		n := add(&node{id: "host:next", kind: kHost, title: "→ " + st.NextAgeName})
		have, need := 0.0, 0.0
		for res, c := range st.NextAgeResReqs {
			need += 1
			have += math.Min(1, frac(st.Resources[res].Amount, c))
		}
		for key, c := range st.NextAgeBldReqs {
			need += 1
			have += math.Min(1, frac(float64(st.Buildings[key].Count), float64(c)))
		}
		n.gauge = frac(have, need)
		n.gaugeLabel = fmt.Sprintf("%.0f%%", 100*n.gauge)
		if st.CurrentAgeWonderName != "" {
			n.sub = "needs " + st.CurrentAgeWonderName
		} else {
			n.sub = "requirements"
		}
		if st.AgeReady {
			n.alert, n.sev, n.sub = "READY", sevInfo, "type advance"
		}
		if m.nodes["svc:wonder"] != nil {
			link("svc:wonder", n.id, 0.2*n.gauge+0.05, eFlow)
		} else {
			link("svc:lab", n.id, 0.2*n.gauge+0.05, eFlow)
		}
	}
	for _, r := range tr.ActiveRoutes {
		n := add(&node{id: "host:route:" + r.Key, kind: kHost, title: r.Name, badge: fmt.Sprintf("×%d", r.CyclesDone)})
		n.sub = flowLine(r.Export, "▸") + " " + flowLine(r.Import, "◂")
		s := eFlow
		if r.Disrupted {
			s = eSevered
			n.alert, n.sev = "BLOCKADED", sevCrit
		}
		link("svc:market", n.id, 0.3, s)
	}
	for _, key := range sortedMapKeys(st.Diplomacy.Factions) {
		f := st.Diplomacy.Factions[key]
		if !f.Discovered {
			continue
		}
		n := add(&node{id: "host:civ:" + key, kind: kHost, title: f.Name, badge: f.Status})
		n.gauge = float64(f.Opinion+100) / 200
		n.gaugeLabel = fmt.Sprintf("%+d", f.Opinion)
		n.sub = f.Specialty
		s := eIdle
		switch {
		case f.AtWar:
			s = eSevered
			n.alert, n.sev = "AT WAR", sevCrit
		case f.Status == "embargo":
			s = eSevered
			n.alert, n.sev = "EMBARGO", sevWarn
		case f.TradeCount > 0 || f.Status == "allied":
			s = eFlow
		}
		if f.LentWorkers > 0 {
			n.sub = fmt.Sprintf("lends %d workers", f.LentWorkers)
		}
		link(n.id, "svc:market", 0.15+0.05*float64(f.TradeCount), s)
	}
	for _, ex := range []*game.ExpeditionSnapshot{mil.ActiveScout, mil.ActiveMilitary} {
		if ex == nil {
			continue
		}
		n := add(&node{id: "host:exp:" + ex.Name, kind: kHost, title: ex.Name, badge: fmt.Sprintf("%d⚔", ex.Soldiers)})
		n.sub = fmt.Sprintf("returns in %ds", ex.TicksLeft*2)
		n.alert, n.sev = "AWAY", sevInfo
		link("svc:garrison", n.id, 0.4, eFlow)
	}
	if h := st.Harbinger; h != nil {
		n := add(&node{id: "host:harbinger", kind: kHost, title: h.Name, badge: "harbinger"})
		n.sub = "foretells " + h.TargetEpochName
		n.alert, n.sev = "OMEN", sevWarn
		link(n.id, "svc:pop", 0.1, eIdle)
	}

	// ---- global alarms.
	if st.PendingCatastrophe != "" {
		name, _ := config.CatastropheInfo(st.PendingCatastrophe)
		m.catast = name
		m.alarms = append(m.alarms, "CATASTROPHE: "+strings.ToUpper(name))
	}
	for _, id := range m.order() {
		n := m.nodes[id]
		if n.sev >= sevWarn && n.alert != "" {
			label := n.title
			m.alarms = append(m.alarms, fmt.Sprintf("%s %s", label, n.alert))
		}
	}
	return m
}

// order is a stable listing of node ids: crit before warn, then by column.
func (m *model) order() []string {
	var ids []string
	for c := 0; c < 4; c++ {
		for _, n := range m.cols[c] {
			ids = append(ids, n.id)
		}
	}
	sort.SliceStable(ids, func(i, j int) bool { return m.nodes[ids[i]].sev > m.nodes[ids[j]].sev })
	return ids
}

func anyDiscovered(st game.GameState) bool {
	for _, f := range st.Diplomacy.Factions {
		if f.Discovered {
			return true
		}
	}
	return false
}

func countStartable(rs []game.TradeRouteInfo) int {
	n := 0
	for _, r := range rs {
		if r.CanStart {
			n++
		}
	}
	return n
}

func domainsLine(w game.WorkerState) string {
	type d struct {
		k string
		n int
	}
	var ds []d
	for k, t := range w.Types {
		if t.Count > 0 {
			ds = append(ds, d{k, t.Count})
		}
	}
	sort.Slice(ds, func(i, j int) bool {
		if ds[i].n != ds[j].n {
			return ds[i].n > ds[j].n
		}
		return ds[i].k < ds[j].k
	})
	var parts []string
	for i, x := range ds {
		if i == 3 {
			break
		}
		parts = append(parts, fmt.Sprintf("%s %d", x.k, x.n))
	}
	return strings.Join(parts, " · ")
}

func flowLine(m map[string]float64, arrow string) string {
	var parts []string
	for _, k := range sortedMapKeys(m) {
		parts = append(parts, arrow+k)
	}
	return strings.Join(parts, " ")
}

func sortedMapKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func frac(a, b float64) float64 {
	if b <= 0 {
		return 0
	}
	return math.Max(0, math.Min(1, a/b))
}

// short formats a magnitude compactly: 950, 1.2k, 34M.
func short(v float64) string {
	a := math.Abs(v)
	switch {
	case a >= 1e15:
		return fmt.Sprintf("%.1fP", v/1e15)
	case a >= 1e12:
		return fmt.Sprintf("%.1fT", v/1e12)
	case a >= 1e9:
		return fmt.Sprintf("%.1fG", v/1e9)
	case a >= 1e6:
		return fmt.Sprintf("%.1fM", v/1e6)
	case a >= 1e4:
		return fmt.Sprintf("%.0fk", v/1e3)
	case a >= 1e3:
		return fmt.Sprintf("%.1fk", v/1e3)
	case a >= 10:
		return fmt.Sprintf("%.0f", v)
	case a >= 1:
		return fmt.Sprintf("%.1f", v)
	case a == 0:
		return "0"
	default:
		return fmt.Sprintf("%.2f", v)
	}
}

// signed is a per-tick rate readout, in the game's own "/t" unit.
func signed(v float64) string {
	if math.Abs(v) < 0.005 {
		return "±0/t"
	}
	s := short(v)
	if v > 0 {
		s = "+" + s
	}
	return s + "/t"
}
