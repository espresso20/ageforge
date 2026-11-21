package ui

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/rules"
)

// next_step.go is the rule that names the next step: one value, worked out
// once a refresh from the last snapshot on the UI path (never in a bus
// handler), that feeds the card, the next step plate, Tab and Show me. It
// decides nothing and does nothing: a step that is ignored just stays.

// nextDrySeconds is how soon a falling resource must run out to count as
// running dry.
const nextDrySeconds = 300

// nextThen is how many items follow the step on the card.
const nextThen = 3

// The rules, in the order they are tried; none is the step of a town with
// nothing pressing.
const (
	ruleChoice   = "choice"
	ruleAdvance  = "advance"
	ruleResearch = "research"
	ruleIdle     = "idle"
	ruleDry      = "dry"
	ruleIncome   = "income"
	ruleScholars = "scholars"
	ruleHomes    = "homes"
	ruleGate     = "gate"
	ruleNone     = "none"
)

// nextItem is a thing that follows the step: what it is, one line on it, and
// where Show me takes the player: the card of a building or a panel.
type nextItem struct {
	title, detail string
	panel, card   string
}

// nextStep is what to do next. card is the building to build ("" for a step
// that is not one) and cmd the command that does it ("" for none); panel is
// the panel word for a step that lives in a panel ("" for one on the main
// view). cost is a card's price, have what the town holds of it (enough) or
// what it is short of.
type nextStep struct {
	rule, title, detail string
	why                 []string
	cost, have          string
	enough              bool
	panel, card, cmd    string
	then                []nextItem
	view                panelView // the panels found and opened, which the list shows
	// onLine is set by the view, not the rule: the command line holds cmd, so
	// Enter runs the step.
	onLine bool
}

// item is the step as an item that follows another.
func (s nextStep) item() nextItem {
	return nextItem{title: s.title, detail: s.detail, panel: s.panel, card: s.card}
}

// computeNext works the next step out of a snapshot: the first rule that
// matches is NOW, then come the next three things. What cannot wait comes
// first, then what makes the town run by itself, then the gate.
func computeNext(st game.GameState, v panelView) nextStep {
	n := nextCtx{st: st, cards: buildCards(st, buyOne), set: st.Ruleset()}
	var steps []nextStep
	for _, rule := range []func(*nextCtx) (nextStep, bool){
		(*nextCtx).choice, (*nextCtx).advance, (*nextCtx).research, (*nextCtx).idle,
		(*nextCtx).dry, (*nextCtx).income, (*nextCtx).scholars, (*nextCtx).homes, (*nextCtx).gate,
	} {
		s, ok := rule(&n)
		if !ok || s.panel != "" && !v.has(s.panel) {
			continue
		}
		// Two rules that ask for the same building are one step.
		if s.card == "" || !slices.ContainsFunc(steps, func(o nextStep) bool { return o.card == s.card }) {
			steps = append(steps, s)
		}
	}
	now := nextStep{rule: ruleNone, title: rowid.Line("ui.next_step.computeNext.01", "Nothing is pressing")}
	var items []nextItem
	if len(steps) > 0 {
		now = steps[0]
		for _, s := range steps[1:] {
			items = append(items, s.item())
		}
	}
	for _, it := range n.passive() {
		if it.panel == "" || v.has(it.panel) {
			items = append(items, it)
		}
	}
	now.then, now.view = items[:min(len(items), nextThen)], v
	return now
}

// nextCtx is a snapshot with the cards of its buildings and its ruleset.
type nextCtx struct {
	st    game.GameState
	cards []buildCard
	set   *rules.Set
}

// pick is the cheapest card of a building the player can still build that
// the test accepts, affordable or not.
func (n *nextCtx) pick(ok func(bs game.BuildingState, def config.BuildingDef) bool) (buildCard, bool) {
	var best buildCard
	var bestCost float64
	found := false
	for _, c := range n.cards {
		bs, def := n.st.Buildings[c.key], config.BuildingDef{}
		if d, has := n.set.Building(c.key); has {
			def = d
		}
		if !bs.Unlocked || bs.AtMaxCount || !ok(bs, def) {
			continue
		}
		cost := 0.0
		for _, p := range c.price {
			cost += p.need
		}
		if !found || cost < bestCost {
			best, bestCost, found = c, cost, true
		}
	}
	return best, found
}

// going reports whether a building the test accepts is going up.
func (n *nextCtx) going(ok func(bs game.BuildingState, def config.BuildingDef) bool) bool {
	return slices.ContainsFunc(n.st.BuildQueue, func(q game.BuildQueueSnapshot) bool {
		def, _ := n.set.Building(q.Key)
		return ok(n.st.Buildings[q.Key], def)
	})
}

// stands reports whether a building the test accepts stands.
func (n *nextCtx) stands(ok func(bs game.BuildingState, def config.BuildingDef) bool) bool {
	for key, bs := range n.st.Buildings {
		if def, has := n.set.Building(key); has && bs.Count > 0 && ok(bs, def) {
			return true
		}
	}
	return false
}

// copies is how many of key are going up.
func (n *nextCtx) copies(key string) int {
	c := 0
	for _, q := range n.st.BuildQueue {
		if q.Key == key {
			c++
		}
	}
	return c
}

// want is pick for a rule that answers a need: nothing while a building the
// test accepts is going up, since the need is being answered.
func (n *nextCtx) want(ok func(bs game.BuildingState, def config.BuildingDef) bool) (buildCard, bool) {
	if n.going(ok) {
		return buildCard{}, false
	}
	return n.pick(ok)
}

// makes accepts a building that brings res in.
func makes(res string) func(game.BuildingState, config.BuildingDef) bool {
	return func(_ game.BuildingState, def config.BuildingDef) bool {
		return slices.ContainsFunc(def.Effects, func(e config.Effect) bool {
			return e.Type == "production" && e.Target == res && e.Value > 0
		})
	}
}

// isHome accepts a building that houses people.
func isHome(_ game.BuildingState, def config.BuildingDef) bool {
	return slices.ContainsFunc(def.Effects, func(e config.Effect) bool {
		return e.Type == "capacity" && e.Target == "population" && e.Value > 0
	})
}

// gatherFor is the resource to gather by hand toward card c's price and how
// much of it is missing: the part the town is furthest short of that hands
// can gather and nothing brings in. "" when the card can be built, gathering
// has ended or no part is one.
func (n *nextCtx) gatherFor(c buildCard) (res string, short float64) {
	if c.canBuild || !n.st.CanGatherByHand() {
		return "", 0
	}
	worst := 0.0
	for _, p := range c.price {
		if p.ok || p.need <= 0 || !handGathered(p.res) || n.st.Resources[p.res].Rate > 0 {
			continue
		}
		if f := (p.need - p.have) / p.need; f > worst {
			res, short, worst = p.res, p.need-p.have, f
		}
	}
	return res, short
}

// build is a step that builds card c: its price, and what the town holds of
// it or what it is short of. A price that only hands can fill makes the
// step's command the gather that fills it.
func (n *nextCtx) build(rule, title, detail string, why []string, c buildCard) nextStep {
	s := nextStep{rule: rule, title: title, detail: detail, why: why, card: c.key, cmd: buildLine(c.key), enough: c.canBuild}
	for i, p := range c.price {
		if i > 0 {
			s.cost += ", "
		}
		s.cost += p.text
	}
	switch {
	case c.canBuild && len(c.price) > 0:
		s.have = rowid.Sprintf("ui.next_step.build.01", "you have %s", textfmt.Number(c.price[0].have))
	case c.reason != "":
		s.have = c.reason
		if c.wait != "" {
			s.have += ", " + c.wait
		}
	}
	if res, short := n.gatherFor(c); res != "" {
		s.cmd = command("gather", res)
		s.have = rowid.Sprintf("ui.next_step.build.02", "gather %s by hand", game.Amount(short, res))
	}
	return s
}

// choice: a choice blocks play. The offer of a memory is a modal that can be
// left waiting, so it is not a step.
func (n *nextCtx) choice() (nextStep, bool) {
	if n.st.PendingCatastrophe == "" {
		return nextStep{}, false
	}
	return nextStep{rule: ruleChoice, title: rowid.Line("ui.next_step.choice.01", "A choice is waiting"),
		detail: rowid.Line("ui.next_step.choice.02", "the town waits for it"),
		why:    []string{rowid.Line("ui.next_step.choice.03", "A catastrophe has come and needs an answer.")},
		panel:  "catastrophe", cmd: "catastrophe"}, true
}

// advance: the age is ready.
func (n *nextCtx) advance() (nextStep, bool) {
	if !n.st.AgeReady {
		return nextStep{}, false
	}
	return nextStep{rule: ruleAdvance, title: rowid.Sprintf("ui.next_step.advance.01", "Advance to the %s", n.st.NextAgeName),
		detail: rowid.Line("ui.next_step.advance.02", "every pillar stands"),
		why:    []string{rowid.Line("ui.next_step.advance.03", "The gate to the next age is open.")},
		panel:  "charter", cmd: "advance"}, true
}

// research: nothing is being researched, a tech can start and there are
// scholars to work it. The keystone if it can start, else the nearest tech on
// the way to it that can, else the cheapest that can.
func (n *nextCtx) research() (nextStep, bool) {
	rs := n.st.Research
	if rs.CurrentTech != "" || rs.PerTick <= 0 {
		return nextStep{}, false
	}
	key, keystone := "", false
	if n.st.CurrentAgeWonderKey != "" {
		if def, ok := n.set.Building(n.st.CurrentAgeWonderKey); ok && def.RequiredTech != "" && !rs.Techs[def.RequiredTech].Researched {
			key, keystone = nearestStartable(rs.Techs, def.RequiredTech), true
		}
	}
	if key == "" {
		keystone = false
		for _, k := range sortedKeysOf(rs.Techs) {
			if t := rs.Techs[k]; t.Available && (key == "" || t.Cost < rs.Techs[key].Cost) {
				key = k
			}
		}
	}
	if key == "" {
		return nextStep{}, false
	}
	why := []string{rowid.Line("ui.next_step.research.01", "A tech can start."), rowid.Line("ui.next_step.research.02", "Nothing is being researched.")}
	if keystone {
		why[0] = rowid.Sprintf("ui.next_step.research.03", "It leads to the keystone for the %s.", n.st.NextAgeName)
	}
	return nextStep{rule: ruleResearch, title: rowid.Sprintf("ui.next_step.research.04", "Research %s", rs.Techs[key].Name),
		detail: rowid.Line("ui.next_step.research.05", "nothing is being researched"), why: why, panel: "research", cmd: "research " + key}, true
}

// nearestStartable is the tech to start for the target: the target itself if
// it can start, else the first one that can, looking back through what it
// needs breadth first; "" when none can.
func nearestStartable(techs map[string]game.TechState, target string) string {
	seen := map[string]bool{target: true}
	for queue := []string{target}; len(queue) > 0; queue = queue[1:] {
		t := techs[queue[0]]
		if t.Available && !t.Researched {
			return queue[0]
		}
		for _, need := range append(slices.Clone(t.Prerequisites), t.AnyOf...) {
			if !seen[need] && !techs[need].Researched {
				seen[need] = true
				queue = append(queue, need)
			}
		}
	}
	return ""
}

// scholars: research stands still for want of scholars, with a tech in work
// or one that could start. A town with no home builds that first.
func (n *nextCtx) scholars() (nextStep, bool) {
	rs := n.st.Research
	waiting := rs.CurrentTech != ""
	for _, t := range rs.Techs {
		waiting = waiting || t.Available
	}
	if rs.PerTick > 0 || !waiting || n.st.Workers.MaxPop == 0 && !n.going(isHome) {
		return nextStep{}, false
	}
	c, ok := n.want(func(bs game.BuildingState, _ config.BuildingDef) bool { return bs.Scholars > 0 })
	if !ok {
		return nextStep{}, false
	}
	return n.build(ruleScholars, rowid.Sprintf("ui.next_step.scholars.01", "Build a %s", c.name),
		rowid.Line("ui.next_step.scholars.02", "research needs scholars"),
		[]string{rowid.Line("ui.next_step.scholars.03", "Scholars come with it."), rowid.Line("ui.next_step.scholars.04", "Research has none to work it.")}, c), true
}

// idle: people wait and every job is taken.
func (n *nextCtx) idle() (nextStep, bool) {
	w := n.st.Workers
	if w.TotalIdle <= 0 {
		return nextStep{}, false
	}
	for _, bs := range n.st.Buildings {
		if bs.Count*bs.WorkerCapacity > bs.WorkersAssigned {
			return nextStep{}, false
		}
	}
	c, ok := n.want(func(bs game.BuildingState, _ config.BuildingDef) bool { return bs.WorkerCapacity > 0 })
	if !ok {
		return nextStep{}, false
	}
	bs := n.st.Buildings[c.key]
	return n.build(ruleIdle, rowid.Sprintf("ui.next_step.idle.01", "Give %d waiting people work", w.TotalIdle),
		rowid.Line("ui.next_step.idle.02", "every job is taken"),
		[]string{rowid.Sprintf("ui.next_step.idle.03", "A %s has jobs for %d.", c.name, bs.WorkerCapacity), rowid.Line("ui.next_step.idle.04", "Every job is taken.")}, c), true
}

// homes: the homes are full.
func (n *nextCtx) homes() (nextStep, bool) {
	w := n.st.Workers
	if w.HeldBy != game.HeldHousing {
		return nextStep{}, false
	}
	c, ok := n.want(isHome)
	if !ok {
		return nextStep{}, false
	}
	if w.MaxPop == 0 {
		return n.build(ruleHomes, rowid.Sprintf("ui.next_step.homes.01", "Build a %s", c.name),
			rowid.Line("ui.next_step.homes.04", "the town has no homes yet"),
			[]string{rowid.Line("ui.next_step.homes.05", "People need a home before they come.")}, c), true
	}
	return n.build(ruleHomes, rowid.Sprintf("ui.next_step.homes.01", "Build a %s", c.name),
		rowid.Sprintf("ui.next_step.homes.02", "homes are full: %d of %d", w.TotalPop, w.MaxPop),
		[]string{rowid.Sprintf("ui.next_step.homes.03", "Homes are full: %d of %d.", w.TotalPop, w.MaxPop)}, c), true
}

// gate: the first row of the gate to the next age the player can act on.
func (n *nextCtx) gate() (nextStep, bool) {
	v := n.st.AdvanceView()
	for _, p := range v.Pillars {
		for _, r := range p.Rows {
			if r.Met || r.Kind == game.GateRowCount && r.Have+n.copies(r.Key) >= r.Need {
				continue
			}
			detail := rowid.Sprintf("ui.next_step.gate.01", "the %s asks for %d", v.NextName, r.Need)
			why := []string{rowid.Sprintf("ui.next_step.gate.02", "%s: %d of %d.", r.Label, r.Have, r.Need)}
			if r.Kind == game.GateRowCount {
				if i := slices.IndexFunc(n.cards, func(c buildCard) bool { return c.key == r.Key }); i >= 0 {
					if bs := n.st.Buildings[r.Key]; bs.Unlocked && !bs.AtMaxCount {
						c := n.cards[i]
						return n.build(ruleGate, rowid.Sprintf("ui.next_step.gate.03", "Build a %s", c.name), detail, why, c), true
					}
				}
			}
			if r.Kind == game.GateRowDone && n.st.Research.Techs[r.Key].Available {
				return nextStep{rule: ruleGate, title: rowid.Sprintf("ui.next_step.gate.04", "Research %s", n.st.Research.Techs[r.Key].Name),
					detail: detail, why: why, panel: "research", cmd: "research " + r.Key}, true
			}
			return nextStep{rule: ruleGate, title: r.Label, detail: detail, why: why, panel: "charter", cmd: "charter"}, true
		}
	}
	return nextStep{}, false
}

// dry: a resource falls and runs out within nextDrySeconds.
func (n *nextCtx) dry() (nextStep, bool) {
	secs := tickSeconds(n.st.TickIntervalMs)
	var res string
	soon := math.Inf(1)
	for _, k := range sortedKeysOf(n.st.Resources) {
		r := n.st.Resources[k]
		if !r.Unlocked || r.Rate >= 0 {
			continue
		}
		if left := r.Amount / -r.Rate * secs; left < nextDrySeconds && left < soon {
			res, soon = k, left
		}
	}
	if res == "" {
		return nextStep{}, false
	}
	c, ok := n.want(makes(res))
	if !ok {
		return nextStep{}, false
	}
	name := game.ResourceName(res)
	left := textfmt.Duration(time.Duration(soon * float64(time.Second)))
	return n.build(ruleDry, rowid.Sprintf("ui.next_step.dry.01", "%s is running out", name),
		rowid.Sprintf("ui.next_step.dry.02", "none left in %s", left),
		[]string{rowid.Sprintf("ui.next_step.dry.03", "%s runs out in %s.", name, left), rowid.Sprintf("ui.next_step.dry.04", "A %s makes more.", c.name)}, c), true
}

// income: a resource the town still gathers by hand has no building that
// brings it in, standing or going up. A town with no home builds that first.
func (n *nextCtx) income() (nextStep, bool) {
	if !n.st.CanGatherByHand() || n.st.Workers.MaxPop == 0 && !n.going(isHome) {
		return nextStep{}, false
	}
	for _, res := range handRes {
		if r := n.st.Resources[res]; !r.Unlocked || r.Rate > 0 || n.stands(makes(res)) {
			continue
		}
		c, ok := n.want(makes(res))
		if !ok {
			continue
		}
		name := game.ResourceName(res)
		return n.build(ruleIncome, rowid.Sprintf("ui.next_step.income.01", "Build a %s", c.name),
			rowid.Sprintf("ui.next_step.income.02", "nothing brings %s in yet", name),
			[]string{rowid.Sprintf("ui.next_step.income.03", "Only hands bring %s in so far.", name), rowid.Sprintf("ui.next_step.income.04", "A %s brings it in by itself.", c.name)}, c), true
	}
	return nextStep{}, false
}

// passive lists the things that follow the matching steps, in a fixed order:
// research in progress, age progress, the next milestone, the head of the
// plan, a festival, a harbinger.
func (n *nextCtx) passive() []nextItem {
	st := n.st
	var out []nextItem
	if rs := st.Research; rs.CurrentTech != "" {
		title := rs.CurrentTechName
		if rs.PerTick <= 0 {
			title += " " + needsScholarsShort.String()
		} else if rs.TicksLeft > 0 {
			left := textfmt.Duration(time.Duration(float64(rs.TicksLeft) * tickSeconds(st.TickIntervalMs) * float64(time.Second)))
			title = rowid.Sprintf("ui.next_step.passive.01", "%s done in %s", rs.CurrentTechName, left)
		}
		out = append(out, nextItem{title: title,
			detail: rowid.Line("ui.next_step.passive.02", "research goes on after that"), panel: "research"})
	}
	if v := st.AdvanceView(); !v.Final {
		out = append(out, nextItem{title: rowid.Sprintf("ui.next_step.passive.03", "%s: %s of the way", v.AgeName, fmt.Sprintf("%.0f%%", v.Progress*100)),
			detail: rowid.Line("ui.next_step.passive.04", "four pillars to fill"), panel: "charter"})
	}
	var near []string
	for k, m := range st.Milestones.Milestones {
		if !m.Completed && m.Visible && len(m.Progress) > 0 {
			near = append(near, k)
		}
	}
	frac := func(m game.MilestoneInfo) float64 {
		sum := 0.0
		for _, p := range m.Progress {
			sum += math.Min(1, p.Current/math.Max(p.Target, 1))
		}
		return sum / float64(len(m.Progress))
	}
	sort.Slice(near, func(i, j int) bool {
		a, b := st.Milestones.Milestones[near[i]], st.Milestones.Milestones[near[j]]
		if fa, fb := frac(a), frac(b); fa != fb {
			return fa > fb
		}
		return near[i] < near[j]
	})
	if len(near) > 0 {
		m := st.Milestones.Milestones[near[0]]
		out = append(out, nextItem{title: m.Name, detail: m.Description, panel: "milestones"})
	}
	if len(st.Plan) > 0 {
		out = append(out, nextItem{title: st.Plan[0].Name, detail: rowid.Line("ui.next_step.passive.05", "first in the plan"), panel: "plan"})
	}
	if f := st.Festival; f.Ready && f.Culture >= f.Cost {
		out = append(out, nextItem{title: rowid.Line("ui.next_step.passive.06", "Hold a festival"),
			detail: rowid.Line("ui.next_step.passive.07", "the town can afford one"), panel: "festival"})
	}
	if h := st.Harbinger; h != nil {
		out = append(out, nextItem{title: h.Name, detail: rowid.Line("ui.next_step.passive.08", "a harbinger is here"), panel: "harbinger"})
	}
	return out
}
