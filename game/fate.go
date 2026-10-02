package game

import (
	"fmt"
	"math"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
)

// Fated dooms: when a catastrophe strikes is decided at random, in secret.
//
// On entering an era from the Iron Era on (config.FateAllowed: past the Iron
// gate, the Cosmic Era included), a hidden roll decides whether a doom is
// fated there. If it is, a
// strike tick is drawn uniformly across the era's expected length
// (expectedEraTicks: the sum of its ages' expectedAgeTicks) from the
// tick it was entered, so the doom can fall anywhere in any of its ages,
// mid-age included. The fate is persisted (GameSave.Fate): a reload cannot
// re-roll it.
//
// Every expected duration here (the era's window, the harbinger's lead, the
// "before this age is out" forecast, the shortest warning) is measured in
// expectedAgeTicks, never a hard-coded time, so a pacing change carries
// through in one place.
//
// A harbinger comes only when a doom is fated, a random lead before the
// strike: LeadFrac × the current age's expectedAgeTicks, never before the era
// began. A quiet era means safe, for now. Until the harbinger arrives nothing
// may reveal whether a doom is fated: CatastropheOutlook, the panels, the
// commands, the log and the map read the same either way.
//
// At the strike tick the strike rolls: the old passage chance by the faith
// band at that moment (epochGoodChance × catastropheChanceOnBadRoll) ×
// FateStrikeScale, times the Appease multiplier. A miss means Spared. An
// Invite makes it certain. A struck doom is a pending catastrophe: the popup
// waits for Endure or Succumb, and advancing and prestige wait with it.
//
// It cannot be outrun. The strike lands before an age advance that would
// carry the player past it: the era's final transition, or any advance once
// the figure speaking has said the doom falls before this age is out. If no
// harbinger has come by then, it comes at that advance, which waits for one
// more try: a doom is always foretold before it strikes. The Cosmic Era has
// no transition out: its passage is prestige, so there the doom lands before
// a prestige instead (fateBeforePrestige), ahead of the Last Passage.
//
// The Cosmic Era runs two threads: the Last Passage's, which starts on
// entering it (harbinger.go, last_passage.go), and its fated doom's (the
// Reality Tear). While the doom's thread speaks, the Last Passage's waits in
// ge.parkedHarbinger with its answers intact, and takes up the warning again
// when the doom resolves. If both come at once, the fated doom resolves first.
//
// False prophets. An era entered before the Industrial Age with nothing fated
// rolls its first age's FalseProphetChance (the Stone Era, where nothing can
// be fated, included). A false prophet draws its foretold moment and lead
// exactly as a real doom would and arrives the same way, claiming medium or
// high. At the foretold moment nothing happens; it is revealed once its
// foretold window passes without a doom: the end of the age if the figure
// speaking said "before this age is out", else the end of the era. Figures
// with no timing are revealed at the era's end.
//
// Randomness: every roll draws five values from ge.rng whatever the outcome,
// so the stream's shape never depends on the fate; a strike draws one more.
// Everything here expects the write lock, except the read-only helpers the
// snapshot uses (fateSettled, strikeChance, fateWhen).

const (
	// FateChance is the chance a doom is fated on entering an era past the
	// Iron gate. It is calibrated against the transition rolls it replaces:
	// a first run to the Modern Age prestige used to meet four catastrophe-
	// capable transitions at the old passage chance (12% to 18% by faith),
	// and now lives through three fated eras (the Iron, Steel and Electric
	// Eras) at FateStrikeScale times that chance: 3 × 0.27 × 5 ≈ 4. The
	// Digital Era's fate rarely strikes before that prestige.
	FateChance = 0.27

	// FateStrikeScale turns the old passage chance (a bad epoch roll that
	// escalates: 12% at high faith, 15% mid, 18% low) into the chance a fated
	// doom strikes: 60%, 75% and 90%. A harbinger means it; faith and Appease
	// can still lift it.
	FateStrikeScale = 5.0

	// harbingerLeadMin and harbingerLeadMax bound the harbinger's lead before
	// the strike, as a fraction of the current age's expectedAgeTicks.
	harbingerLeadMin = 0.20
	harbingerLeadMax = 0.60
)

// Fate outcomes (FateSave.Resolved).
const (
	FateStruck   = "struck"   // the strike rolled and hit: a catastrophe is pending
	FateSpared   = "spared"   // the strike rolled and missed
	FateRevealed = "revealed" // a false prophet's window passed without a doom
)

// Harbinger timing statements (HarbingerSave.When, HarbingerView.When): what
// the figure speaking says about when the doom falls. A figure with
// config.TimingNone says nothing (WhenUntold).
const (
	WhenUntold  = ""
	WhenThisAge = "this_age" // before this age is out
	WhenThisEra = "this_era" // later, but before the era ends
)

// FateSave is the current era's hidden fate, persisted as GameSave.Fate. It is
// never shown to the player: only the harbinger it sends is.
type FateSave struct {
	// EpochKey is the era the fate belongs to.
	EpochKey string `json:"epoch_key"`
	// Fated: a doom will strike this era (subject to the strike roll).
	Fated bool `json:"fated,omitempty"`
	// FalseProphet: nothing is fated, but a false prophet will come. An
	// Invite makes its doom real (Fated), and the record still says it lied.
	FalseProphet bool `json:"false_prophet,omitempty"`
	// Arrived: its harbinger has come (one per era per run).
	Arrived bool `json:"arrived,omitempty"`
	// EntryTick is the tick the era was entered (or the fate rolled).
	EntryTick int `json:"entry_tick,omitempty"`
	// Window is the era's expected length in ticks (expectedEraTicks).
	Window int `json:"window,omitempty"`
	// StrikeTick is when the doom strikes, or the moment a false prophet
	// foretells. Absolute tick.
	StrikeTick int `json:"strike_tick,omitempty"`
	// LeadFrac is the harbinger's lead before StrikeTick as a fraction of the
	// current age's expectedAgeTicks.
	LeadFrac float64 `json:"lead_frac,omitempty"`
	// Claim is the tier a false prophet claims (medium or high).
	Claim CatastropheTier `json:"claim,omitempty"`
	// Invited: an Invite made the strike certain.
	Invited bool `json:"invited,omitempty"`
	// Resolved is how the doom ended: "" while open, else a Fate* outcome.
	Resolved string `json:"resolved,omitempty"`
	// ResolvedTick is the tick it resolved.
	ResolvedTick int `json:"resolved_tick,omitempty"`
	// AtAdvance: it resolved at an age advance the player reached first (the
	// strike could not be outrun), not at StrikeTick.
	AtAdvance bool `json:"at_advance,omitempty"`
}

// expectedAgeTicks is how long a player is expected to spend in age, in
// ticks: the one measure of expected duration for the fate and the
// harbinger (the era's window, the lead, the timing forecast, the shortest
// warning). Today it is the pacing target (config.AgeTargetTicks); a
// per-age speed-up, such as a mastered age running faster, divides it here
// and nowhere else.
func expectedAgeTicks(age string) float64 { return config.AgeTargetTicks(age) }

// expectedEraTicks is epochKey's expected length: its ages' expectedAgeTicks
// summed. 0 for an unknown epoch.
func expectedEraTicks(epochKey string) float64 {
	total := 0.0
	for _, a := range config.EpochByKey()[epochKey].Ages {
		total += float64(expectedAgeTicks(a)) // rounded: the target is a product once inlined
	}
	return total
}

// fateSaveCopy returns a copy of the fate for a save, or nil.
func (ge *GameEngine) fateSaveCopy() *FateSave {
	if ge.fate == nil {
		return nil
	}
	f := *ge.fate
	return &f
}

// open reports whether the fate still has something to do: a doom or a false
// prophet not yet resolved.
func (f *FateSave) open() bool {
	return f != nil && f.Resolved == "" && (f.Fated || f.FalseProphet)
}

// lying reports whether the fate is a false prophet whose doom was never made
// real by an Invite.
func (f *FateSave) lying() bool { return f.FalseProphet && !f.Fated }

// rollFate rolls the current era's fate, entered now. Draws five values from
// ge.rng whatever the outcome.
func (ge *GameEngine) rollFate() {
	ep := ge.currentEpoch
	window := int(math.Round(expectedEraTicks(ep)))
	f := &FateSave{EpochKey: ep, EntryTick: ge.tick, Window: window}
	rng := ge.gameRNG()
	fatedRoll, falseRoll, offsetRoll, leadRoll, claimRoll := rng.Float64(), rng.Float64(), rng.Float64(), rng.Float64(), rng.Float64()
	f.Fated = config.FateAllowed(ep) && fatedRoll < FateChance
	if !f.Fated {
		def, _ := config.HarbingerFor(ge.age)
		f.FalseProphet = falseRoll < def.FalseProphetChance
	}
	if f.Fated || f.FalseProphet {
		f.StrikeTick = f.EntryTick + int(offsetRoll*float64(window))
		f.LeadFrac = harbingerLeadMin + float64(leadRoll*(harbingerLeadMax-harbingerLeadMin))
	}
	if f.FalseProphet {
		f.Claim = CatastropheTierMedium
		if claimRoll < 0.5 {
			f.Claim = CatastropheTierHigh
		}
	}
	ge.fate = f
	ge.publishFate(EventFateRolled, "", -1)
}

// ensureFate rolls the current era's fate if it has none: the first tick of a
// run (the Stone Era is never entered by a transition), and saves written
// before fates existed.
func (ge *GameEngine) ensureFate() {
	if ge.fate == nil || ge.fate.EpochKey != ge.currentEpoch {
		ge.rollFate()
	}
}

// fateThread is the live thread of the current era's doom, or nil (none has
// come, or the live thread is the Last Passage's). Read-only.
func (ge *GameEngine) fateThread() *HarbingerSave {
	f, h := ge.fate, ge.harbinger
	if f == nil || h == nil || h.TargetEpoch == "" || h.TargetEpoch != f.EpochKey {
		return nil
	}
	return h
}

// fateArrivalTick is when the fate's harbinger is due: the lead before the
// strike, measured in the current age's target.
func (ge *GameEngine) fateArrivalTick() int {
	f := ge.fate
	return f.StrikeTick - int(f.LeadFrac*expectedAgeTicks(ge.age))
}

// fateHarbingerDue reports whether the fate's harbinger should arrive now. A
// Last Passage thread in the way is parked when it does (fateArrive).
func (ge *GameEngine) fateHarbingerDue() bool {
	f := ge.fate
	if !f.open() || f.EpochKey != ge.currentEpoch || f.Arrived {
		return false
	}
	if h := ge.harbinger; h != nil && h.TargetEpoch != "" {
		return false
	}
	return ge.tick >= ge.fateArrivalTick()
}

// fateTick runs the fate for one tick (or one offline step): roll it if the
// era has none, bring the harbinger when it is due, and strike (or let a
// false prophet's moment pass) at StrikeTick. A strike waits while another
// catastrophe is pending: one is never overwritten.
func (ge *GameEngine) fateTick() {
	ge.ensureFate()
	f := ge.fate
	if !f.open() {
		return
	}
	if ge.fateHarbingerDue() {
		ge.fateArrive()
	}
	if ge.tick < f.StrikeTick || f.lying() || ge.pendingCatastrophe != "" {
		return
	}
	if ge.fateThread() == nil {
		return // never unwarned: the harbinger comes first
	}
	ge.fateStrike(false)
}

// fateNextEventIn is how many ticks until the fate's next event (the
// harbinger's arrival or the strike), for the offline catch-up to stop at; 0
// when nothing is due ahead.
func (ge *GameEngine) fateNextEventIn() int {
	f := ge.fate
	if !f.open() || f.EpochKey != ge.currentEpoch {
		return 0
	}
	next := 0
	consider := func(t int) {
		if d := t - ge.tick; d > 0 && (next == 0 || d < next) {
			next = d
		}
	}
	if !f.Arrived {
		consider(ge.fateArrivalTick())
	}
	if !f.lying() {
		consider(f.StrikeTick)
	}
	return next
}

// fateArrive brings the fate's harbinger: the current age's figure, with its
// timing, the claimed tier for a false prophet, and its lines. Reports whether
// one arrived.
func (ge *GameEngine) fateArrive() bool {
	f := ge.fate
	def, ok := config.HarbingerFor(ge.age)
	if !ok {
		return false
	}
	f.Arrived = true
	h := &HarbingerSave{
		Age:          def.Age,
		Chain:        []string{def.Age},
		EpochKey:     f.EpochKey,
		TargetEpoch:  f.EpochKey,
		FalseProphet: f.FalseProphet,
		ArrivedTick:  ge.tick,
	}
	if f.FalseProphet {
		// The claim is a fixed multiple of what a real doom's chance would be,
		// so faith and Appease move it the same way.
		if base := ge.strikeBase(); base > 0 {
			h.ClaimFactor = float64(FateStrikeScale*harbingerClaimBase[f.Claim]) / base
		} else {
			h.ClaimFactor = 1
		}
	}
	if lp := ge.harbinger; lp != nil && lp.TargetEpoch == "" {
		// The Last Passage's thread waits while the doom's speaks.
		ge.parkedHarbinger = lp
	}
	ge.harbinger = h
	// A doom is always foretold: never sooner than the shortest lead after
	// its harbinger comes (a strike fated for the era's first moments, before
	// any lead could reach back, is held until then).
	if soonest := ge.tick + int(harbingerLeadMin*expectedAgeTicks(ge.age)); f.StrikeTick < soonest {
		f.StrikeTick = soonest
	}
	h.When = ge.fateWhen(def)
	tier, _ := ge.harbingerDisplay()
	h.AnnouncedTier = tier
	h.Lines = ge.harbingerSpeak(def, tier)

	ge.addLog("event", fmt.Sprintf("⚑ %s has come, warning of %s. Type 'harbinger' to answer.",
		capFirst(def.Name), ge.fateWarningText(h)))
	ge.harbingerLogLines()
	ge.publishHarbinger(def, f.EpochKey, false)
	return true
}

// fateWhen is what def's figure says about when the current fate falls: nothing
// for a figure with no timing; "before this age is out" when the doom falls in
// the current age on the era's expected schedule (its ages at their targets
// from EntryTick), or when this is the era's last age, which it cannot
// outlast; otherwise "before the era ends". Read-only.
func (ge *GameEngine) fateWhen(def config.HarbingerDef) string {
	f := ge.fate
	if f == nil || def.ForecastTiming == config.TimingNone {
		return WhenUntold
	}
	ages := config.EpochByKey()[f.EpochKey].Ages
	end := float64(f.EntryTick)
	for i, a := range ages {
		end += float64(expectedAgeTicks(a))
		if a != ge.age {
			continue
		}
		if i == len(ages)-1 || float64(f.StrikeTick) < end {
			return WhenThisAge
		}
		return WhenThisEra
	}
	return WhenThisEra
}

// fateWarningText is what a fate thread warns of, with its timing: "impending
// doom", "impending doom before this age is out" or "impending doom before
// the Iron Era ends". It names only the current era.
func (ge *GameEngine) fateWarningText(h *HarbingerSave) string {
	switch h.When {
	case WhenThisAge:
		return "impending doom before this age is out"
	case WhenThisEra:
		return fmt.Sprintf("impending doom before the %s ends", config.EpochByKey()[h.EpochKey].Name)
	}
	return "impending doom"
}

// strikeBase is the chance a fated doom strikes by the faith band right now,
// before Appease: 60%, 75% or 90%. Read-only.
func (ge *GameEngine) strikeBase() float64 {
	return float64((1-ge.epochGoodChance())*catastropheChanceOnBadRoll) * FateStrikeScale
}

// StrikeChanceAt is the chance a fated doom strikes at faith fill fill
// (hasStorage false: faith has no storage yet) with appease levels of Appease
// bought: the rule the strike rolls by. Pure; for the smoke report, which
// models the expected catastrophes of a run from the faith it lived at.
func StrikeChanceAt(fill float64, hasStorage bool, appease int) float64 {
	good := goodChanceFor(fill, hasStorage)
	p := float64((1-good)*catastropheChanceOnBadRoll) * FateStrikeScale
	if appease > 0 {
		p *= detmath.Pow(harbingerAppeaseFactor, float64(appease))
	}
	return p
}

// strikeChance is the chance the current era's doom strikes if it rolled now:
// strikeBase × the Appease multiplier, or certain when invited. For a false
// prophet it is what a real doom's chance would be (its claim follows it).
// Read-only.
func (ge *GameEngine) strikeChance() float64 {
	if f := ge.fate; f != nil && f.Invited {
		return 1
	}
	return ge.strikeBase() * ge.harbingerAppeaseMultiplier()
}

// fateTierFor buckets a strike chance: the same thresholds as the Last
// Passage's odds, scaled by FateStrikeScale, so the three faith bands still
// map one-to-one (high faith → low, mid → medium, low faith → high).
func fateTierFor(p float64) CatastropheTier {
	return catastropheTierFor(p / FateStrikeScale)
}

// fateSettled reports whether the current era's doom has come or been
// lifted, which the player has seen (struck, or spared). A quiet era, a fated
// one whose harbinger has not come, and one whose false prophet was revealed
// all read false: the outlook must not tell them apart. Read-only.
func (ge *GameEngine) fateSettled() bool {
	f := ge.fate
	return f != nil && f.EpochKey == ge.currentEpoch && (f.Resolved == FateStruck || f.Resolved == FateSpared)
}

// fateStrike rolls the strike now and settles the doom and its harbinger.
// atAdvance says it came at an advance the player reached first. Reports
// whether it struck (a catastrophe is now pending). Draws one value from
// ge.rng whatever the odds.
func (ge *GameEngine) fateStrike(atAdvance bool) bool {
	f := ge.fate
	p := ge.strikeChance()
	roll := ge.gameRNG().Float64()
	struck := f.Invited || roll < p
	f.ResolvedTick, f.AtAdvance = ge.tick, atAdvance
	arrived := ge.harbingerArrivedTick()
	if struck {
		f.Resolved = FateStruck
		source := catastropheRolled
		if f.Invited {
			source = catastropheInvited
		}
		ge.triggerCatastrophe(f.EpochKey, source)
	} else {
		f.Resolved = FateSpared
	}
	ge.resolveHarbinger(f.EpochKey, struck)
	ge.publishFate(EventFateResolved, f.Resolved, arrived)
	return struck
}

// harbingerArrivedTick is the tick the current era's doom's harbinger
// arrived, or -1 when none is here.
func (ge *GameEngine) harbingerArrivedTick() int {
	if h := ge.fateThread(); h != nil {
		return h.ArrivedTick
	}
	return -1
}

// revealFalseProphet ends a false prophet whose foretold window has passed:
// eraEnds says the era is ending (else the age). The verdict names neither
// what comes next.
func (ge *GameEngine) revealFalseProphet(eraEnds bool) {
	f := ge.fate
	f.Resolved, f.ResolvedTick, f.AtAdvance = FateRevealed, ge.tick, true
	arrived := ge.harbingerArrivedTick()
	if h := ge.fateThread(); h != nil {
		def, _ := config.HarbingerFor(h.Age)
		window := "This age ends"
		if eraEnds {
			window = fmt.Sprintf("The %s ends", config.EpochByKey()[f.EpochKey].Name)
		}
		ge.settleHarbinger(f.EpochKey, false,
			fmt.Sprintf("⚑ %s without the doom %s foretold. The warning had been invented from the start.", window, def.Name))
	}
	ge.publishFate(EventFateResolved, f.Resolved, arrived)
}

// errHarbingerAtGate is the refusal of an advance (or, in the final epoch, a
// prestige) that brought a doom's harbinger: the doom cannot be outrun, so it
// is foretold first. again says what to do to meet it.
func errHarbingerAtGate(h *HarbingerSave, warning, again string) error {
	def, _ := config.HarbingerFor(h.Age)
	return fmt.Errorf("%s stands in your way, warning of %s. Type 'harbinger' to answer, or %s again to meet it.", capFirst(def.Name), warning, again)
}

// fateBeforeAdvance runs before every age advance to next (the advance
// command and the plan's): it settles a doom the advance would outrun. The
// era's final transition cannot leave an open doom behind, and nor can any
// advance once the figure speaking has said it falls before this age is out.
// If the harbinger has not come yet it comes now and the advance waits (an
// error), except in the Stone Era; a false prophet is revealed and the
// advance goes on; a real doom strikes, and if it hits the advance waits
// behind the pending catastrophe.
func (ge *GameEngine) fateBeforeAdvance(next string) error {
	f := ge.fate
	if !f.open() || f.EpochKey != ge.currentEpoch {
		return nil
	}
	leaving := config.EpochForAge(next) != ge.currentEpoch
	h := ge.fateThread()
	promised := h != nil && h.When == WhenThisAge
	if !leaving && !promised {
		return nil
	}
	return ge.fateAtPassage(leaving, "advance", "advancing")
}

// fateBeforePrestige runs before a prestige from the final epoch, whose
// passage is prestige itself: its doom cannot be outrun past that either, so
// it settles first, the same way, ahead of the Last Passage's roll. (A
// prestige from an earlier era leaves its open doom behind with the run.)
func (ge *GameEngine) fateBeforePrestige() error {
	f := ge.fate
	if !f.open() || f.EpochKey != ge.currentEpoch || !config.IsFinalEpoch(f.EpochKey) {
		return nil
	}
	return ge.fateAtPassage(true, "confirm prestige", "prestiging")
}

// fateAtPassage settles the open doom at a passage the player reached first:
// leaving says the passage ends the era; again and gerund word the refusal.
func (ge *GameEngine) fateAtPassage(leaving bool, again, gerund string) error {
	f := ge.fate
	// A doom not yet foretold is foretold now, and the passage waits. In an
	// era where nothing can strike (the Stone Era) a false prophet who has
	// not come by its end never comes: holding the advance up for a warning
	// the rules already rule out would only confuse.
	if ge.fateThread() == nil && !f.Arrived && config.FateAllowed(f.EpochKey) && ge.fateArrive() {
		return errHarbingerAtGate(ge.harbinger, ge.fateWarningText(ge.harbinger), again)
	}
	if f.lying() {
		ge.revealFalseProphet(leaving)
		return nil
	}
	if ge.pendingCatastrophe != "" {
		return ge.catastropheBlockErr(gerund)
	}
	if ge.fateStrike(true) {
		return ge.catastropheBlockErr(gerund)
	}
	return nil
}

// publishFate tells the smoke report and tests about the fate (EventFateRolled
// at a roll, EventFateResolved at a strike, a spared doom or a revealed false
// prophet; arrived is the tick its harbinger came, -1 for none). The payload
// carries the hidden fate, so the UI must never subscribe to either event
// (ui/fate_guard_test.go holds it to that). Bus handlers run under the write
// lock and must not call back into the engine.
func (ge *GameEngine) publishFate(event, outcome string, arrived int) {
	f := ge.fate
	if f == nil {
		return
	}
	payload := map[string]interface{}{
		"epoch_key":     f.EpochKey,
		"fated":         f.Fated,
		"false_prophet": f.FalseProphet,
		"entry_tick":    f.EntryTick,
		"window":        f.Window,
		"strike_tick":   f.StrikeTick,
		"lead_frac":     f.LeadFrac,
		"tick":          ge.tick,
		"age":           ge.age,
	}
	if outcome != "" {
		payload["outcome"] = outcome
		payload["at_advance"] = f.AtAdvance
		payload["invited"] = f.Invited
		if arrived >= 0 {
			payload["arrived_tick"] = arrived
		}
	}
	ge.Bus.Publish(EventData{Type: event, Payload: payload})
}

// restoreFateState loads the fate from a save. A save written before fates
// existed has none: one standing in a Stone Era or a fated era with a live
// thread from the old rules has that thread carried over (a warning of the
// era's end becomes a doom that strikes at the era's end at the latest);
// otherwise the first tick rolls the era's fate. Under the write lock, after
// the epoch is restored and before the harbinger is.
func (ge *GameEngine) restoreFateState(save *GameSave) {
	ge.fate = nil
	if save.Fate != nil {
		f := *save.Fate
		if f.EpochKey == ge.currentEpoch {
			if (f.Fated || f.FalseProphet) && f.LeadFrac <= 0 {
				f.LeadFrac = harbingerLeadMin
			}
			ge.fate = &f
		}
		return
	}
	h := save.Harbinger
	if h == nil || h.TargetEpoch == "" || h.EpochKey != ge.currentEpoch || !config.FateAllowed(h.EpochKey) {
		return
	}
	// (A Last Passage thread, TargetEpoch "", keeps its rules: the Cosmic
	// Era's own fate is rolled on the first tick.)
	window := int(math.Round(expectedEraTicks(h.EpochKey)))
	ge.fate = &FateSave{
		EpochKey:     h.EpochKey,
		Fated:        !h.FalseProphet || save.CatastropheInvited,
		FalseProphet: h.FalseProphet,
		EntryTick:    ge.tick,
		Window:       window,
		// The old thread promised the era's end: past any tick this era can
		// reach, so the gate at the final transition strikes it.
		StrikeTick: math.MaxInt32,
		LeadFrac:   harbingerLeadMin,
		Claim:      h.AnnouncedTier,
		Invited:    save.CatastropheInvited,
		Arrived:    true,
	}
}

// --- test hooks -----------------------------------------------------------------

// ForceFateForTest fates a doom in epochKey, which must be the current era
// and one that can be fated, striking strikeOffset ticks after the era was
// entered (its EntryTick, or now if it has no fate yet). The lead is kept
// (or set to its midpoint), the doom is open again, and no harbinger has come
// for it. For tests and the smoke suite: never reachable from play.
func (ge *GameEngine) ForceFateForTest(epochKey string, strikeOffset int) error {
	return ge.forceFate(epochKey, true, false, strikeOffset)
}

// ForceQuietFateForTest makes epochKey (the current era) quiet: nothing fated
// and no false prophet. For tests: never reachable from play.
func (ge *GameEngine) ForceQuietFateForTest(epochKey string) error {
	return ge.forceFate(epochKey, false, false, 0)
}

// ForceFalseProphetForTest sends epochKey (the current era, before the
// Industrial Age or not) a false prophet foretelling doom strikeOffset ticks
// after the era began, claiming high. For tests: never reachable from play.
func (ge *GameEngine) ForceFalseProphetForTest(epochKey string, strikeOffset int) error {
	return ge.forceFate(epochKey, false, true, strikeOffset)
}

func (ge *GameEngine) forceFate(epochKey string, fated, falseProphet bool, strikeOffset int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if epochKey != ge.currentEpoch {
		return fmt.Errorf("the current era is %s, not %s", ge.currentEpoch, epochKey)
	}
	if fated && !config.FateAllowed(epochKey) {
		return fmt.Errorf("nothing can be fated in %s", epochKey)
	}
	f := ge.fate
	if f == nil || f.EpochKey != epochKey {
		f = &FateSave{EpochKey: epochKey, EntryTick: ge.tick, Window: int(math.Round(expectedEraTicks(epochKey)))}
	}
	if f.LeadFrac <= 0 {
		f.LeadFrac = (harbingerLeadMin + harbingerLeadMax) / 2
	}
	f.Fated, f.FalseProphet, f.Invited, f.Arrived = fated, falseProphet, false, false
	f.Resolved, f.ResolvedTick, f.AtAdvance = "", 0, false
	f.StrikeTick = f.EntryTick + strikeOffset
	f.Claim = ""
	if falseProphet {
		f.Claim = CatastropheTierHigh
	}
	if !fated && !falseProphet {
		f.StrikeTick, f.LeadFrac = 0, 0
	}
	if ge.fateThread() != nil {
		ge.harbinger = nil
		ge.resumeLastPassageThread()
	}
	ge.fate = f
	return nil
}

// FateForTest returns a copy of the current era's fate (nil when it has none).
// It reads hidden state: for tests and the smoke report only, never the UI.
func (ge *GameEngine) FateForTest() *FateSave {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.fateSaveCopy()
}
