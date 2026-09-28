package smoke

import (
	"math"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// Harbinger policies for the bot (--harbinger).
const (
	HarbingerIgnore  = "ignore"
	HarbingerAppease = "appease"
	HarbingerBrace   = "brace"
	HarbingerBoth    = "both"
)

func (b *Bot) wantsAppease() bool {
	return b.Harbinger == HarbingerAppease || b.Harbinger == HarbingerBoth
}

func (b *Bot) wantsBrace() bool {
	return b.Harbinger == HarbingerBrace || b.Harbinger == HarbingerBoth
}

// harbingerTargets folds the next Appease level (up to level 2) and the
// level-1 Brace price the policy wants into the plan, so the bot invests
// toward them like any requirement.
func (b *Bot) harbingerTargets(p *plan) {
	v := p.st.Harbinger
	if v == nil {
		return
	}
	add := func(cost map[string]float64) {
		for r, c := range cost {
			p.target[r] += c
			p.capNeed[r] = math.Max(p.capNeed[r], c)
		}
	}
	if b.wantsAppease() && v.AppeaseBlocked == "" {
		add(v.AppeaseCost) // the next level's price; nil once maxed
	}
	if b.wantsBrace() && v.BraceLevel == 0 && v.BraceBlocked == "" {
		add(v.BraceCost)
	}
}

// answerHarbinger buys the next Appease level (both levels) and level 1 of
// Brace, whichever the policy wants, as soon as each is affordable. Reports
// whether it spent anything.
func (b *Bot) answerHarbinger(st game.GameState) bool {
	v := st.Harbinger
	if v == nil {
		return false
	}
	spent := false
	if b.wantsAppease() && v.AppeaseLevel < game.HarbingerMaxAppease && v.AppeaseAffordable && b.spareFor(st, v.AppeaseCost) {
		spent = b.act("harbinger_appease", v.TargetEpochKey, b.ge.HarbingerAppease()) || spent
	}
	if b.wantsBrace() && v.BraceLevel == 0 && v.BraceAffordable {
		spent = b.act("harbinger_brace", v.TargetEpochKey, b.ge.HarbingerBrace()) || spent
	}
	return spent
}

// spareFor reports whether cost leaves what the next advance still needs of
// each resource: its requirement and what the age's wonder still has to be
// banked. A player saves faith for the Sistine Chapel before appeasing with
// it; the bot used to spend it first and stall the Renaissance.
func (b *Bot) spareFor(st game.GameState, cost map[string]float64) bool {
	for res, c := range cost {
		keep := st.NextAgeResReqs[res]
		if w := st.CurrentAgeWonderKey; w != "" {
			if left := b.defs[w].BaseCost[res] - st.Buildings[w].WonderBank[res]; left > 0 {
				keep += left
			}
		}
		if st.Resources[res].Amount < c+keep {
			return false
		}
	}
	return true
}

// ResourcePoint is one resource's stock and cap at a moment.
type ResourcePoint struct {
	Stock   float64 `json:"stock"`
	Storage float64 `json:"storage"`
}

// HarbingerThread is one harbinger thread as the bot lived through it.
type HarbingerThread struct {
	Cycle       int      `json:"cycle"`
	Epoch       string   `json:"epoch"`
	TargetEpoch string   `json:"target_epoch"`
	Speakers    []string `json:"speakers"` // ages whose figure spoke, first to last
	StartTick   int      `json:"start_tick"`
	Outcome     string   `json:"outcome"` // a game verdict, or unresolved/cleared
	// FalseProphet is only known once the thread resolves (the live view
	// hides it by design); nil while unresolved.
	FalseProphet *bool `json:"false_prophet,omitempty"`

	AppeaseL1 map[string]float64 `json:"appease_l1_cost"`
	// AppeaseL2 is the level-2 price, read once level 1 is bought.
	AppeaseL2 map[string]float64 `json:"appease_l2_cost,omitempty"`
	BraceL1   map[string]float64 `json:"brace_l1_cost"`
	// Stock and storage of every priced resource when the thread started and
	// at the last decision before it resolved (the passage).
	AtStart   map[string]ResourcePoint `json:"at_start"`
	AtPassage map[string]ResourcePoint `json:"at_passage"`

	AppeaseAffordable bool `json:"appease_ever_affordable"`
	BraceAffordable   bool `json:"brace_ever_affordable"`
	AppeaseFits       bool `json:"appease_ever_fits_storage"`
	BraceFits         bool `json:"brace_ever_fits_storage"`
	// Seconds (1x) from the thread's start to first affordability; -1 if never.
	AppeaseAfterSecs float64 `json:"appease_affordable_after_seconds"`
	BraceAfterSecs   float64 `json:"brace_affordable_after_seconds"`
	// AppeaseL2AfterSecs is when level 2 first became affordable with level 1
	// already bought (-1 if never): the second payment, not the sum.
	AppeaseL2AfterSecs float64 `json:"appease_l2_affordable_after_seconds"`
	// LengthSecs is the thread's span so far: start to its last decision.
	LengthSecs   float64 `json:"length_seconds"`
	AppeaseLevel int     `json:"appease_level_bought"`
	BraceLevel   int     `json:"brace_level_bought"`

	startSim time.Duration
}

func affordableIn(st game.GameState, cost map[string]float64) (afford, fits bool) {
	if len(cost) == 0 {
		return false, false
	}
	afford, fits = true, true
	for r, c := range cost {
		rs := st.Resources[r]
		if rs.Amount < c {
			afford = false
		}
		if rs.Storage < c {
			fits = false
		}
	}
	return afford, fits
}

func pricedPoints(st game.GameState, costs ...map[string]float64) map[string]ResourcePoint {
	out := map[string]ResourcePoint{}
	for _, c := range costs {
		for r := range c {
			rs := st.Resources[r]
			out[r] = ResourcePoint{Stock: rs.Amount, Storage: rs.Storage}
		}
	}
	return out
}

// trackHarbinger follows the live thread from one snapshot to the next.
func (r *runner) trackHarbinger(st game.GameState) {
	v := st.Harbinger
	th := r.thread
	if th != nil && (v == nil || v.TargetEpochKey != th.TargetEpoch || th.Cycle != r.cycle || r.restarted(th, v.Age)) {
		r.closeThread(st)
		th = nil
	}
	if v == nil {
		return
	}
	if th == nil {
		th = &HarbingerThread{
			Cycle: r.cycle, Epoch: config.EpochForAge(v.Age), TargetEpoch: v.TargetEpochKey,
			StartTick: r.ticks, Outcome: "unresolved", startSim: r.sim,
			AppeaseAfterSecs: -1, BraceAfterSecs: -1, AppeaseL2AfterSecs: -1,
		}
		// The live view prices the NEXT level; at the start that is level 1.
		if v.AppeaseLevel == 0 {
			th.AppeaseL1 = v.AppeaseCost
		}
		if v.BraceLevel == 0 {
			th.BraceL1 = v.BraceCost
		}
		th.AtStart = pricedPoints(st, th.AppeaseL1, th.BraceL1)
		r.thread = th
		r.res.Harbingers = append(r.res.Harbingers, th)
	}
	if n := len(th.Speakers); n == 0 || th.Speakers[n-1] != v.Age {
		th.Speakers = append(th.Speakers, v.Age)
	}
	since := (r.sim - th.startSim).Seconds()
	if a, f := affordableIn(st, th.AppeaseL1); a || f {
		th.AppeaseFits = th.AppeaseFits || f
		if a && !th.AppeaseAffordable {
			th.AppeaseAffordable, th.AppeaseAfterSecs = true, since
		}
	}
	if a, f := affordableIn(st, th.BraceL1); a || f {
		th.BraceFits = th.BraceFits || f
		if a && !th.BraceAffordable {
			th.BraceAffordable, th.BraceAfterSecs = true, since
		}
	}
	if v.AppeaseLevel == 1 && th.AppeaseL2 == nil {
		th.AppeaseL2 = v.AppeaseCost
	}
	if v.AppeaseLevel == 1 && v.AppeaseAffordable && th.AppeaseL2AfterSecs < 0 {
		th.AppeaseL2AfterSecs = since
	}
	// A level bought since the last look (the bot plays right after an age
	// advance, which this snapshot predates) was affordable by now at the
	// latest.
	if v.AppeaseLevel >= 1 && !th.AppeaseAffordable {
		th.AppeaseAffordable, th.AppeaseAfterSecs = true, since
	}
	if v.AppeaseLevel >= 2 && th.AppeaseL2AfterSecs < 0 {
		th.AppeaseL2AfterSecs = since
	}
	if v.BraceLevel >= 1 && !th.BraceAffordable {
		th.BraceAffordable, th.BraceAfterSecs = true, since
	}
	th.LengthSecs = since
	th.AppeaseLevel = max(th.AppeaseLevel, v.AppeaseLevel)
	th.BraceLevel = max(th.BraceLevel, v.BraceLevel)
	th.AtPassage = pricedPoints(st, th.AppeaseL1, th.BraceL1)
}

// restarted reports a thread that began again from an earlier age than its
// last speaker: a Succumb reset the run inside the same cycle.
func (r *runner) restarted(th *HarbingerThread, age string) bool {
	n := len(th.Speakers)
	return n > 0 && r.ageIdx[age] < r.ageIdx[th.Speakers[n-1]]
}

// closeThread settles the tracked thread against the game's record of
// resolved harbingers. A prestige or Succumb clears threads without a
// verdict.
func (r *runner) closeThread(st game.GameState) {
	th := r.thread
	r.thread = nil
	for i := len(st.HarbingerHistory) - 1; i >= 0; i-- {
		h := st.HarbingerHistory[i]
		if h.TargetEpochKey == th.TargetEpoch && h.EpochKey == th.Epoch {
			th.Outcome = h.Outcome
			fp := h.FalseProphet
			th.FalseProphet = &fp
			r.res.Stats.HarbingerVerdicts[h.Outcome]++
			if fp {
				r.res.Stats.FalseProphets++
			}
			th.AppeaseLevel = max(th.AppeaseLevel, h.AppeaseLevel)
			th.BraceLevel = max(th.BraceLevel, h.BraceLevel)
			return
		}
	}
	th.Outcome = "cleared"
}

// PriceRow is the static price of one epoch's harbinger answers against the
// most storage the player can build before the passage.
type PriceRow struct {
	Epoch       string             `json:"epoch"`
	TargetEpoch string             `json:"target_epoch"`
	LastAge     string             `json:"last_age"`
	AppeaseL1   map[string]float64 `json:"appease_l1_cost"`
	BraceL1     map[string]float64 `json:"brace_l1_cost"`
	// MaxStorage is -1 where an uncapped storage building covers the resource.
	MaxStorage map[string]float64 `json:"max_storage_by_passage"`
}

// HarbingerPrices summons a harbinger on a scratch engine in each epoch that
// has one (game.SummonHarbingerForTest; never on the played engine) and reads
// the level-1 prices from its view, next to the most storage reachable by the
// epoch's last age.
func HarbingerPrices() []PriceRow {
	var rows []PriceRow
	for _, ep := range config.Epochs() {
		if len(ep.Ages) == 0 {
			continue
		}
		ge := game.NewGameEngine()
		if ge.SummonHarbingerForTest(ep.Ages[0]) != nil {
			continue
		}
		v := ge.GetState().Harbinger
		if v == nil {
			continue
		}
		last := ep.Ages[len(ep.Ages)-1]
		row := PriceRow{
			Epoch: ep.Key, TargetEpoch: v.TargetEpochKey, LastAge: last,
			AppeaseL1: v.AppeaseCost, BraceL1: v.BraceCost, MaxStorage: map[string]float64{},
		}
		for _, c := range []map[string]float64{v.AppeaseCost, v.BraceCost} {
			for res := range c {
				m := MaxStorage(last, res)
				if math.IsInf(m, 1) {
					m = -1 // JSON has no Inf; -1 means no cap
				}
				row.MaxStorage[res] = m
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// MaxStorage is the most storage for res a player can have in ageKey: base
// storage, every capped storage building up to that age at its MaxCount, and
// every storage tech up to that age. +Inf if an uncapped storage building
// covers it. It assumes every earlier storage copy was built in its own age,
// which is the best case.
func MaxStorage(ageKey, res string) float64 {
	return maxStorageIn(config.BuildingByKey(), ageKey, res)
}

// maxStorageIn is MaxStorage over defs (the static checks' broken-number
// tests feed in altered storage buildings).
func maxStorageIn(defs map[string]config.BuildingDef, ageKey, res string) float64 {
	order := map[string]int{}
	for i, k := range config.AgeOrder() {
		order[k] = i
	}
	limit := order[ageKey]
	total := 0.0
	for _, r := range config.BaseResources() {
		if r.Key == res {
			total = r.BaseStorage
		}
	}
	// Sorted, not map order: the float sum below must come out the same on
	// every run or the report's storage caps wobble in the last digit.
	for _, key := range sortedKeys(defs) {
		d := defs[key]
		if d.RequiredAge == "" || order[d.RequiredAge] > limit || d.Category == "wonder" {
			continue
		}
		for _, e := range d.Effects {
			if e.Type != "storage" || (e.Target != res && e.Target != "all") {
				continue
			}
			if d.MaxCount == 0 {
				return math.Inf(1)
			}
			total += float64(e.Value * float64(d.MaxCount))
		}
	}
	for _, t := range config.Technologies() {
		if order[t.Age] > limit {
			continue
		}
		for _, e := range t.Effects {
			if e.Type == "storage" && (e.Target == res || e.Target == "all") {
				total += e.Value
			}
		}
	}
	return total
}
