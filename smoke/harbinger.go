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
			Cycle: r.cycle, Epoch: r.rules.EraOf(v.Age), TargetEpoch: v.TargetEpochKey,
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

// Fate outcomes of a FateRow beyond the engine's own (game.FateStruck,
// game.FateSpared, game.FateRevealed).
const (
	FateQuiet   = "quiet"   // nothing fated, no false prophet
	FateCleared = "cleared" // a prestige or Succumb ended the run before it resolved
	FateOpen    = "open"    // the smoke run ended before it resolved
)

// Catastrophe baseline before fated dooms, measured on origin/master a116b82
// with the greedy bot playing a first run to the Modern Age prestige on seeds
// 1 to 49: the old transition rolls gave 0.72 expected catastrophes per first
// run (four catastrophe-capable transitions at 18%: the bot keeps faith in the
// low band) and 0.673 measured. FateChance is calibrated against it.
const (
	BaselineCatastrophesExpected = 0.72
	BaselineCatastrophesMeasured = 0.673
	BaselineSeeds                = 49
	BaselineCommit               = "a116b82"
)

// FateRow is one era's hidden fate as the run lived it. It comes from the
// engine's fate events (game.EventFateRolled, EventFateResolved), never from
// anything the bot sees.
type FateRow struct {
	Cycle        int    `json:"cycle"`
	Epoch        string `json:"epoch"`
	Fated        bool   `json:"fated"`
	FalseProphet bool   `json:"false_prophet,omitempty"`
	EntryTick    int    `json:"entry_tick"`
	Window       int    `json:"window_ticks"`
	// StrikeFrac is the drawn strike tick (a false prophet's foretold
	// moment) as a share of the era's window, from its entry.
	StrikeFrac float64 `json:"strike_frac,omitempty"`
	LeadFrac   float64 `json:"lead_frac,omitempty"`
	// Outcome is game.FateStruck, FateSpared or FateRevealed, or FateQuiet,
	// FateCleared or FateOpen.
	Outcome string `json:"outcome"`
	// AtAdvance: it resolved at an advance the player reached before the
	// strike tick (the strike could not be outrun).
	AtAdvance bool `json:"at_advance,omitempty"`
	Invited   bool `json:"invited,omitempty"`
	// WarningTicks is how long its harbinger was there before it resolved;
	// WarningAgeFrac is that as a share of the resolving age's target.
	WarningTicks   int     `json:"warning_ticks,omitempty"`
	WarningAgeFrac float64 `json:"warning_age_frac,omitempty"`
	ResolvedAge    string  `json:"resolved_age,omitempty"`
	// Expected is the era's share of the run's expected catastrophes:
	// game.FateChance × the strike chance averaged over the era's window at
	// the faith the bot kept (a strike past the time it spent there lands at
	// the transition, at the faith it left with). It does not depend on this
	// era's own roll, so it measures the rules, not the luck.
	Expected float64 `json:"expected"`

	sumChance  float64 // Σ strike chance × ticks inside the window
	lastChance float64
}

// onFateRolled starts the row for a newly rolled fate, closing the last one.
// A bus handler: it runs under the engine lock and touches only the runner.
func (r *runner) onFateRolled(e game.EventData) {
	epoch, _ := e.Payload["epoch_key"].(string)
	entry, _ := e.Payload["entry_tick"].(int)
	// A roll for the Stone Era means a prestige or a Succumb reset the run;
	// a roll for the next era means the last one was left by an advance,
	// which settles any open doom first. So does a prestige from the final
	// era (its passage); the prestige has already moved the cycle on.
	next := FateCleared
	if prev := r.fate; prev != nil {
		era, _ := r.rules.Era(epoch)
		prevEra, _ := r.rules.Era(prev.Epoch)
		advanced := era.Order == prevEra.Order+1
		prestiged := r.rules.IsFinalEra(prev.Epoch) && r.cycle > prev.Cycle
		if advanced || prestiged {
			// A strike still to come would have landed at that passage: the
			// new era's entry, or the last look before the prestige (the
			// tick counter starts again after it).
			left := entry
			if prestiged {
				left = r.fateTick
			}
			if lived := left - prev.EntryTick; lived < prev.Window && prev.Window > 0 {
				prev.sumChance += float64(prev.lastChance * float64(prev.Window-lived))
			}
		}
		if advanced {
			next = FateOpen
		}
	}
	r.closeFate(next)
	row := &FateRow{Cycle: r.cycle, Epoch: epoch, EntryTick: entry, Outcome: FateQuiet}
	row.Fated, _ = e.Payload["fated"].(bool)
	row.FalseProphet, _ = e.Payload["false_prophet"].(bool)
	row.Window, _ = e.Payload["window"].(int)
	if row.Fated || row.FalseProphet {
		row.Outcome = ""
		strike, _ := e.Payload["strike_tick"].(int)
		row.LeadFrac, _ = e.Payload["lead_frac"].(float64)
		if row.Window > 0 {
			row.StrikeFrac = float64(strike-entry) / float64(row.Window)
		}
	}
	r.fate, r.fateTick = row, entry
	r.res.Fates = append(r.res.Fates, row)
}

// onFateResolved records how the current fate ended. A bus handler.
func (r *runner) onFateResolved(e game.EventData) {
	f := r.fate
	epoch, _ := e.Payload["epoch_key"].(string)
	if f == nil || f.Epoch != epoch {
		return
	}
	f.Outcome, _ = e.Payload["outcome"].(string)
	f.AtAdvance, _ = e.Payload["at_advance"].(bool)
	f.Invited, _ = e.Payload["invited"].(bool)
	f.ResolvedAge, _ = e.Payload["age"].(string)
	tick, _ := e.Payload["tick"].(int)
	if arrived, ok := e.Payload["arrived_tick"].(int); ok && tick >= arrived {
		f.WarningTicks = tick - arrived
		if t := r.rules.TargetTicks(f.ResolvedAge); t > 0 {
			f.WarningAgeFrac = float64(f.WarningTicks) / t
		}
	}
}

// sampleFate adds the time since the last look to the current fate row's
// expected-catastrophe model, at the strike chance the faith (and any Appease
// bought) gives now.
func (r *runner) sampleFate(st game.GameState) {
	f := r.fate
	if f == nil || st.EpochKey != f.Epoch || f.Window <= 0 {
		return
	}
	appease := 0
	if h := st.Harbinger; h != nil && !h.LastPassage {
		appease = h.AppeaseLevel
	}
	chance := game.StrikeChanceAt(st.CatastropheOutlook.FaithFill, st.Resources["faith"].Storage > 0, appease)
	lo, hi := r.fateTick-f.EntryTick, st.Tick-f.EntryTick
	if hi > f.Window {
		hi = f.Window
	}
	if lo >= 0 && hi > lo {
		f.sumChance += float64(chance * float64(hi-lo))
	}
	f.lastChance, r.fateTick = chance, st.Tick
}

// closeFate settles the current row: an unresolved doom or false prophet
// becomes ifOpen (FateCleared or FateOpen), and its expected share is fixed.
func (r *runner) closeFate(ifOpen string) {
	f := r.fate
	if f == nil {
		return
	}
	r.fate = nil
	if f.Outcome == "" {
		f.Outcome = ifOpen
	}
	if f.Window > 0 && r.rules.FateAllowed(f.Epoch) {
		f.Expected = float64(game.FateChance*f.sumChance) / float64(f.Window)
	}
}

// PriceRow is the static price of one thread's harbinger answers against the
// most storage the player can build in the age its harbinger arrives in.
type PriceRow struct {
	Epoch       string `json:"epoch"`
	TargetEpoch string `json:"target_epoch"`
	// Age is the age the harbinger arrives in. Appease is priced on it (what
	// that age makes in the thread's shortest warning); the price then holds
	// for the whole thread, and storage only grows after it.
	Age string `json:"age"`
	// WarningTicks is that shortest warning in ticks at 1x: a fifth of the
	// age's pacing target for a doom, two thirds of it for the Last Passage.
	WarningTicks float64            `json:"appease_warning_ticks"`
	AppeaseL1    map[string]float64 `json:"appease_l1_cost"`
	// BraceL1 is priced by the era for a doom, and on the same warning as
	// Appease for the Last Passage.
	BraceL1 map[string]float64 `json:"brace_l1_cost"`
	// MaxStorage is -1 where an uncapped storage building covers the resource.
	MaxStorage map[string]float64 `json:"max_storage_in_age"`
}

// HarbingerPrices is the level-1 prices of every thread whose doom can be
// answered (game.HarbingerPriceTable), for each age its harbinger can arrive
// in: a fated doom's, and the Last Passage's (a TargetEpoch of ""), next to
// the most storage reachable in that age. The Stone Era is skipped: only
// false prophets come there, and nothing can strike.
func HarbingerPrices() []PriceRow {
	var rows []PriceRow
	for _, p := range game.HarbingerPriceTable() {
		row := PriceRow{
			Epoch: p.Epoch, TargetEpoch: p.Epoch, Age: p.Age, WarningTicks: p.WarningTicks,
			AppeaseL1: p.AppeaseL1, BraceL1: p.BraceL1, MaxStorage: map[string]float64{},
		}
		if p.LastPassage {
			row.TargetEpoch = ""
		}
		for _, c := range []map[string]float64{p.AppeaseL1, p.BraceL1} {
			for res := range c {
				m := MaxStorage(p.Age, res)
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
			if e.Kind == config.EffectFlatStorage && (e.Target == res || e.Target == config.AllResources) {
				total += e.Value
			}
		}
	}
	return total
}
