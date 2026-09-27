package game

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
)

// The Harbinger (Phase 9 follow-up).
//
// When the player enters the last age of an epoch and the coming transition can
// roll a catastrophe, that age's harbinger (config.HarbingerFor) turns up and
// warns of it. The harbinger never blocks anything and never expires: it stays
// until the epoch transition, which resolves it.
//
// While it is present the player may answer it once or more:
//
//   - Appease (faith + culture): each level multiplies the REAL catastrophe
//     chance by harbingerAppeaseFactor. Two levels; the second costs double.
//   - Brace (the age's core resources): softens an Endure if the catastrophe
//     comes. Two levels, stored on the pending catastrophe so it still applies
//     when Endure is chosen later.
//   - Invite: arms catastropheInvited, guaranteeing the catastrophe. Free and
//     final; Appease is pointless afterwards and refuses.
//
// Before the Industrial Age a harbinger may be a false prophet
// (HarbingerDef.FalseProphetChance). A false prophet announces medium or high
// whatever the odds, and reads exactly like a real one; the transition reveals
// it. The odds themselves are only printed from the Industrial Age on
// (ForecastNumeric), where false prophets no longer exist.
//
// Randomness: arrival draws, in order, the false-prophet roll (always), the
// false tier (only for a false prophet), then the arrival and warning lines from
// the engine's flavor Stream. Resolution draws one line. All of it comes from
// the seeded ge.rng under the write lock.
//
// Unexported methods expect the write lock, except the read-only helpers used
// by GetState (harbingerView and the cost functions), which are safe under the
// read lock.

const (
	// HarbingerMaxAppease and HarbingerMaxBrace are the level caps.
	HarbingerMaxAppease = 2
	HarbingerMaxBrace   = 2

	// harbingerAppeaseFactor multiplies the catastrophe chance per Appease level:
	// 0.6 at level 1, 0.36 at level 2.
	harbingerAppeaseFactor = 0.6

	// Cost fractions of the CURRENT storage cap, for level 1. Level 2 costs
	// double (level × fraction).
	//
	// Why storage caps: they are the one number that tracks a player's economy
	// at every age without ever reading zero. Production rates can be zero or
	// negative (faith drains), which would make a rate-based price free or
	// undefined, and fixed per-age tables go stale whenever the economy is
	// retuned. A cap is also what the player must build up to advance at all,
	// since the next age's requirements have to fit in storage, so a fraction
	// of it is always "a slice of what it takes to move on": it matters in the
	// Bronze Age and in the Space Age alike. And because the harbinger never
	// expires, an idle player can simply let the bank fill and pay later.
	//
	// 15% of faith and culture for Appease: faith is also what keeps the
	// epoch roll kind (faith fill bands), so the price has to stay small enough
	// that paying it never undoes the benefit. Worst case, the spend drops the
	// fill from the top band to the bottom one, raising the base chance from
	// 12% to 18% (×1.5); ×0.6 still leaves it at 0.9× of where it started, so
	// Appease always lowers the real odds.
	//
	// 12% of each core resource for Brace: it softens a catastrophe that may
	// not come, so it is priced a little under Appease per resource, but it
	// draws on several resources at once, the ones the next age asks for.
	harbingerAppeaseCostFrac = 0.15
	harbingerBraceCostFrac   = 0.12
)

// Endure numbers by Brace level (index 0 = unbraced): share of destroyable
// buildings destroyed, in percent, and share of each stored resource kept.
var (
	braceDestroyPct = [HarbingerMaxBrace + 1]int{20, 15, 10}
	braceKeepFrac   = [HarbingerMaxBrace + 1]float64{endureResourceKeep, 0.30, 0.45}
)

// Harbinger outcomes (HarbingerRecord.Outcome).
const (
	HarbingerOutcomeFulfilled   = "fulfilled"   // invited, and it came
	HarbingerOutcomeVindicated  = "vindicated"  // warned, not invited, and it came
	HarbingerOutcomeSpared      = "spared"      // a true harbinger, and it did not come
	HarbingerOutcomeDiscredited = "discredited" // a false prophet, and it did not come
)

// HarbingerSave is the live harbinger, persisted as GameSave.Harbinger. Fields
// are only ever added (the save is signed by re-marshalling).
type HarbingerSave struct {
	// Age is the age the harbinger arrived in; it selects the roster entry and
	// the voice of every line.
	Age string `json:"age"`
	// EpochKey is the epoch the harbinger arrived in; TargetEpoch is the epoch
	// whose transition it warns about.
	EpochKey    string `json:"epoch_key"`
	TargetEpoch string `json:"target_epoch"`
	// FalseProphet is the arrival roll. Never shown to the player.
	FalseProphet bool `json:"false_prophet,omitempty"`
	// AnnouncedTier is what the harbinger said at arrival: the real tier for a
	// true harbinger, medium or high for a false prophet.
	AnnouncedTier CatastropheTier `json:"announced_tier"`
	// ArrivalRealTier is the real tier at arrival. A false prophet's displayed
	// severity moves by however far the real tier has moved since, so that
	// Appease or a change in faith shifts it just as it shifts a real one.
	ArrivalRealTier CatastropheTier `json:"arrival_real_tier,omitempty"`
	AppeaseLevel    int             `json:"appease_level,omitempty"`
	BraceLevel      int             `json:"brace_level,omitempty"`
	// Invited is set by the Invite action. catastropheInvited is consumed when
	// the invite is honoured, so the harbinger keeps its own copy to tell
	// Fulfilled from Vindicated at resolution.
	Invited bool `json:"invited,omitempty"`
	// Lines are the arrival and warning lines, shown in the panel.
	Lines       []string `json:"lines,omitempty"`
	ArrivedTick int      `json:"arrived_tick,omitempty"`
}

// HarbingerRecord is a resolved harbinger, for the Epoch overlay.
type HarbingerRecord struct {
	Age             string          `json:"age"`
	Name            string          `json:"name"`
	EpochKey        string          `json:"epoch_key"`
	TargetEpochKey  string          `json:"target_epoch_key"`
	TargetEpochName string          `json:"target_epoch_name"`
	Outcome         string          `json:"outcome"`
	FalseProphet    bool            `json:"false_prophet,omitempty"`
	AnnouncedTier   CatastropheTier `json:"announced_tier,omitempty"`
	AppeaseLevel    int             `json:"appease_level,omitempty"`
	BraceLevel      int             `json:"brace_level,omitempty"`
	Invited         bool            `json:"invited,omitempty"`
	Tick            int             `json:"tick,omitempty"`
}

// HarbingerView is the UI's picture of the live harbinger (GameState.Harbinger).
// It deliberately carries no false-prophet flag.
type HarbingerView struct {
	Key, Name, Description, Age, AgeName  string
	AppeaseLabel, BraceLabel, InviteLabel string
	// Numeric is true from the Industrial Age on: the panel prints Probability.
	Numeric         bool
	TargetEpochKey  string
	TargetEpochName string
	Lines           []string
	// Tier is the severity the panel shows: the live real tier for a true
	// harbinger, the announced tier (shifted with the real one) for a false one.
	Tier CatastropheTier
	// Probability is the real catastrophe chance at the transition, after
	// Appease. Only shown when Numeric.
	Probability  float64
	AppeaseLevel int
	BraceLevel   int
	Invited      bool
	// Costs of the NEXT level; nil when maxed out or not allowed.
	AppeaseCost map[string]float64
	BraceCost   map[string]float64
	// Why an action is refused regardless of cost ("" when it is allowed).
	AppeaseBlocked string
	BraceBlocked   string
	InviteBlocked  string
	// Affordability of the next level (false when blocked).
	AppeaseAffordable bool
	BraceAffordable   bool
	// Endure numbers at the current and the next Brace level.
	EndureDestroyPct     int
	EndureKeepPct        int
	NextEndureDestroyPct int
	NextEndureKeepPct    int
}

// --- Arrival -----------------------------------------------------------------

// isLastAgeOfEpoch reports whether age is the final age of epochKey.
func isLastAgeOfEpoch(age, epochKey string) bool {
	ep, ok := config.EpochByKey()[epochKey]
	return ok && len(ep.Ages) > 0 && ep.Ages[len(ep.Ages)-1] == age
}

// maybeHarbingerArrive brings the current age's harbinger if the player stands
// in the last age of an epoch whose transition can roll a catastrophe, and no
// harbinger has come this epoch this run. Called after an age advance and after
// a load, under the write lock.
func (ge *GameEngine) maybeHarbingerArrive() {
	if ge.harbinger != nil || ge.harbingerArrived[ge.currentEpoch] {
		return
	}
	if !isLastAgeOfEpoch(ge.age, ge.currentEpoch) {
		return
	}
	ge.harbingerArrive()
}

// harbingerArrive makes the current age's harbinger present, if the next
// transition can roll a catastrophe. Skips the last-age and once-per-epoch
// checks; maybeHarbingerArrive owns those. Reports whether one arrived.
func (ge *GameEngine) harbingerArrive() bool {
	out := ge.catastropheOutlook()
	if !out.Possible {
		return false
	}
	def, ok := config.HarbingerFor(ge.age)
	if !ok {
		return false
	}
	if ge.harbingerArrived == nil {
		ge.harbingerArrived = make(map[string]bool)
	}
	ge.harbingerArrived[ge.currentEpoch] = true

	rng := ge.gameRNG()
	// Draw the false-prophet roll every time so the stream's shape does not
	// depend on the age's chance.
	falseProphet := rng.Float64() < def.FalseProphetChance
	announced := out.Tier
	if falseProphet {
		announced = CatastropheTierMedium
		if rng.Intn(2) == 1 {
			announced = CatastropheTierHigh
		}
	}

	stream := ge.flavorStream()
	arrival := stream.Line(flavor.Request{Moment: flavor.HarbingerArrival, Age: def.Age, Subject: def.Name}, rng)
	warning := stream.Line(flavor.Request{Moment: flavor.HarbingerWarning, Age: def.Age, Subject: def.Name, Kind: string(announced)}, rng)
	var lines []string
	for _, l := range []string{arrival, warning} {
		if l != "" {
			lines = append(lines, l)
		}
	}

	ge.harbinger = &HarbingerSave{
		Age:             def.Age,
		EpochKey:        ge.currentEpoch,
		TargetEpoch:     out.NextEpochKey,
		FalseProphet:    falseProphet,
		AnnouncedTier:   announced,
		ArrivalRealTier: out.Tier,
		Lines:           lines,
		ArrivedTick:     ge.tick,
	}

	target := config.EpochByKey()[out.NextEpochKey]
	ge.addLog("event", fmt.Sprintf("⚑ %s has come, warning of the passage into the %s. Type 'harbinger' to answer.",
		capFirst(def.Name), target.Name))
	for _, l := range lines {
		ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", l))
	}
	ge.Bus.Publish(EventData{
		Type: EventHarbingerArrived,
		Payload: map[string]interface{}{
			"harbinger_key":  def.Key,
			"harbinger_name": def.Name,
			"epoch_key":      out.NextEpochKey,
		},
	})
	return true
}

// summonHarbinger is the dev console's /harbinger: bring the current age's
// harbinger now, ignoring the last-age and once-per-epoch rules (the transition
// must still be able to roll a catastrophe). Takes the write lock.
func (ge *GameEngine) summonHarbinger() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if ge.harbinger != nil {
		return fmt.Errorf("a harbinger is already here")
	}
	// /age jumps the age without the epoch; line them up first.
	if ep := config.EpochForAge(ge.age); ep != ge.currentEpoch {
		ge.currentEpoch = ep
	}
	if !ge.harbingerArrive() {
		return fmt.Errorf("the next epoch transition cannot bring a catastrophe")
	}
	return nil
}

// --- Odds ---------------------------------------------------------------------

// harbingerAppeaseMultiplier is the factor Appease applies to the catastrophe
// chance: 1, 0.6 or 0.36. Read-only.
func (ge *GameEngine) harbingerAppeaseMultiplier() float64 {
	if ge.harbinger == nil || ge.harbinger.AppeaseLevel <= 0 {
		return 1
	}
	return math.Pow(harbingerAppeaseFactor, float64(ge.harbinger.AppeaseLevel))
}

var tierRank = map[CatastropheTier]int{
	CatastropheTierNone: 0, CatastropheTierLow: 1, CatastropheTierMedium: 2, CatastropheTierHigh: 3,
}

var tierByRank = [...]CatastropheTier{CatastropheTierNone, CatastropheTierLow, CatastropheTierMedium, CatastropheTierHigh}

// harbingerDisplayTier is the severity the panel shows. Read-only.
func (ge *GameEngine) harbingerDisplayTier(real CatastropheTier) CatastropheTier {
	h := ge.harbinger
	if h == nil || !h.FalseProphet {
		return real
	}
	r := tierRank[h.AnnouncedTier] + tierRank[real] - tierRank[h.ArrivalRealTier]
	if r < 1 {
		r = 1 // a warning is never "nothing to fear"
	}
	if r > 3 {
		r = 3
	}
	return tierByRank[r]
}

// --- Costs --------------------------------------------------------------------

// harbingerLevelCost prices level (1 or 2) of an action as frac × level of the
// current storage cap of each key, rounded up. Keys that are locked or have no
// storage are skipped. Read-only.
func (ge *GameEngine) harbingerLevelCost(keys []string, frac float64, level int) map[string]float64 {
	cost := make(map[string]float64)
	for _, k := range keys {
		if !ge.Resources.IsUnlocked(k) {
			continue
		}
		store := ge.Resources.GetStorage(k)
		if store <= 0 {
			continue
		}
		cost[k] = math.Ceil(store * frac * float64(level))
	}
	return cost
}

// harbingerAppeaseCost is the price of Appease level (1 or 2): faith and
// culture (culture only once it is unlocked, in the Classical Age). Read-only.
func (ge *GameEngine) harbingerAppeaseCost(level int) map[string]float64 {
	return ge.harbingerLevelCost([]string{"faith", "culture"}, harbingerAppeaseCostFrac, level)
}

// harbingerBraceResources are the core resources Brace draws on: what the next
// age asks for (its ResourceReqs), minus faith and culture, which belong to
// Appease. Falls back to the epoch's primary resource. Config order. Read-only.
func (ge *GameEngine) harbingerBraceResources() []string {
	want := map[string]bool{}
	order := config.AgeOrder()
	for i, a := range order {
		if a == ge.age && i+1 < len(order) {
			for k := range config.AgeByKey()[order[i+1]].ResourceReqs {
				if k != "faith" && k != "culture" {
					want[k] = true
				}
			}
			break
		}
	}
	if len(want) == 0 {
		if ep, ok := config.EpochByKey()[ge.currentEpoch]; ok {
			want[ep.PrimaryResource] = true
		}
	}
	var keys []string
	for _, def := range config.BaseResources() {
		if want[def.Key] {
			keys = append(keys, def.Key)
		}
	}
	return keys
}

// harbingerBraceCost is the price of Brace level (1 or 2). Read-only.
func (ge *GameEngine) harbingerBraceCost(level int) map[string]float64 {
	return ge.harbingerLevelCost(ge.harbingerBraceResources(), harbingerBraceCostFrac, level)
}

// --- Actions ------------------------------------------------------------------

// harbingerActionCheck returns the common refusal for all three actions.
func (ge *GameEngine) harbingerActionCheck() error {
	if ge.harbinger == nil {
		return fmt.Errorf("no harbinger is here — one comes in the last age of an epoch, when the next transition can bring a catastrophe")
	}
	return nil
}

// appeaseBlocked explains why Appease cannot be bought now, or "".
func (ge *GameEngine) appeaseBlocked() string {
	h := ge.harbinger
	switch {
	case h == nil:
		return "no harbinger is here"
	case h.Invited:
		return "you invited the catastrophe; it will come whatever you offer"
	case h.AppeaseLevel >= HarbingerMaxAppease:
		return "already appeased as far as it goes"
	}
	return ""
}

// braceBlocked explains why Brace cannot be bought now, or "".
func (ge *GameEngine) braceBlocked() string {
	h := ge.harbinger
	switch {
	case h == nil:
		return "no harbinger is here"
	case h.BraceLevel >= HarbingerMaxBrace:
		return "already braced as far as it goes"
	}
	return ""
}

// inviteBlocked explains why Invite cannot be chosen now, or "".
func (ge *GameEngine) inviteBlocked() string {
	h := ge.harbinger
	switch {
	case h == nil:
		return "no harbinger is here"
	case h.Invited:
		return "already invited"
	}
	return ""
}

// shortfall lists what cost needs beyond what is stored, in config order.
func (ge *GameEngine) shortfall(cost map[string]float64) string {
	var parts []string
	for _, def := range config.BaseResources() {
		need, ok := cost[def.Key]
		if !ok {
			continue
		}
		if have := ge.Resources.Get(def.Key); have < need {
			parts = append(parts, fmt.Sprintf("%.0f more %s", math.Ceil(need-have), def.Key))
		}
	}
	return strings.Join(parts, ", ")
}

// HarbingerAppease buys the next Appease level. Takes the write lock.
func (ge *GameEngine) HarbingerAppease() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if err := ge.harbingerActionCheck(); err != nil {
		return err
	}
	if why := ge.appeaseBlocked(); why != "" {
		return fmt.Errorf("cannot appease: %s", why)
	}
	h := ge.harbinger
	level := h.AppeaseLevel + 1
	cost := ge.harbingerAppeaseCost(level)
	if len(cost) == 0 {
		return fmt.Errorf("cannot appease: you have no faith storage to offer from")
	}
	if !ge.Resources.Pay(cost) {
		return fmt.Errorf("cannot afford to appease: need %s", ge.shortfall(cost))
	}
	h.AppeaseLevel = level
	def, _ := config.HarbingerFor(h.Age)
	ge.addLog("success", fmt.Sprintf("⚑ %s (Appease %d/%d): paid %s. The catastrophe chance is now ×%.2f.",
		def.AppeaseLabel, level, HarbingerMaxAppease, harbingerCostText(cost), ge.harbingerAppeaseMultiplier()))
	ge.harbingerFlavorLog(flavor.HarbingerAppeased, "")
	return nil
}

// HarbingerBrace buys the next Brace level. Takes the write lock.
func (ge *GameEngine) HarbingerBrace() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if err := ge.harbingerActionCheck(); err != nil {
		return err
	}
	if why := ge.braceBlocked(); why != "" {
		return fmt.Errorf("cannot brace: %s", why)
	}
	h := ge.harbinger
	level := h.BraceLevel + 1
	cost := ge.harbingerBraceCost(level)
	if len(cost) == 0 {
		return fmt.Errorf("cannot brace: none of this age's core resources have storage yet")
	}
	if !ge.Resources.Pay(cost) {
		return fmt.Errorf("cannot afford to brace: need %s", ge.shortfall(cost))
	}
	h.BraceLevel = level
	def, _ := config.HarbingerFor(h.Age)
	ge.addLog("success", fmt.Sprintf("⚑ %s (Brace %d/%d): paid %s. If the catastrophe comes and you Endure, %d%% of buildings fall (not 20%%) and %.0f%% of stock is kept (not 15%%).",
		def.BraceLabel, level, HarbingerMaxBrace, harbingerCostText(cost), braceDestroyPct[level], braceKeepFrac[level]*100))
	ge.harbingerFlavorLog(flavor.HarbingerBraced, "")
	return nil
}

// HarbingerInvite arms the invite: the coming transition brings the
// catastrophe. Free and final. Takes the write lock.
func (ge *GameEngine) HarbingerInvite() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if err := ge.harbingerActionCheck(); err != nil {
		return err
	}
	if why := ge.inviteBlocked(); why != "" {
		return fmt.Errorf("cannot invite: %s", why)
	}
	h := ge.harbinger
	h.Invited = true
	ge.inviteCatastrophe()
	def, _ := config.HarbingerFor(h.Age)
	target := config.EpochByKey()[h.TargetEpoch]
	ge.addLog("warning", fmt.Sprintf("⚑ %s: you have invited it. The passage into the %s will bring the catastrophe. This cannot be undone.",
		def.InviteLabel, target.Name))
	ge.harbingerFlavorLog(flavor.HarbingerInvited, "")
	return nil
}

// harbingerFlavorLog writes one gray flavor line for moment in the live
// harbinger's voice. Draws from ge.rng.
func (ge *GameEngine) harbingerFlavorLog(moment flavor.Moment, kind string) {
	h := ge.harbinger
	if h == nil {
		return
	}
	def, _ := config.HarbingerFor(h.Age)
	if l := ge.flavorStream().Line(flavor.Request{Moment: moment, Age: h.Age, Subject: def.Name, Kind: kind}, ge.gameRNG()); l != "" {
		ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", l))
	}
}

// --- Resolution ---------------------------------------------------------------

// resolveHarbinger settles the live harbinger at the transition into epochKey.
// came reports whether this transition made a catastrophe pending. Logs the
// verdict, hands the Brace level to the pending catastrophe, records the
// outcome and clears the harbinger. Called by detectEpochTransition right
// after the roll, under the write lock.
func (ge *GameEngine) resolveHarbinger(epochKey string, came bool) {
	h := ge.harbinger
	if h == nil {
		return
	}
	def, _ := config.HarbingerFor(h.Age)
	target := config.EpochByKey()[epochKey]
	name := capFirst(def.Name)

	var outcome, msg, kind string
	var moment flavor.Moment
	switch {
	case came && h.Invited:
		outcome, moment = HarbingerOutcomeFulfilled, flavor.HarbingerFulfilled
		msg = fmt.Sprintf("⚑ You invited it, and it came, as %s said it would.", def.Name)
		if h.FalseProphet {
			kind = flavor.KindFalseProphet
			msg = fmt.Sprintf("⚑ You invited it, and it came, though %s, it emerged, had seen nothing at all.", def.Name)
		}
	case came:
		outcome, moment = HarbingerOutcomeVindicated, flavor.HarbingerVindicated
		msg = fmt.Sprintf("⚑ %s was right.", name)
		if h.FalseProphet {
			msg = fmt.Sprintf("⚑ %s was right, though it emerged that the warning had been invented.", name)
		}
	case h.FalseProphet:
		outcome, moment = HarbingerOutcomeDiscredited, flavor.HarbingerDiscredited
		msg = fmt.Sprintf("⚑ The %s dawns untouched. %s was a false prophet: the warning had been invented.", target.Name, name)
	default:
		outcome, moment = HarbingerOutcomeSpared, flavor.HarbingerSpared
		msg = fmt.Sprintf("⚑ The %s dawns untouched. %s's warning was real, and you were spared.", target.Name, name)
	}
	ge.addLog("event", msg)
	if came && h.BraceLevel > 0 {
		ge.pendingBraceLevel = h.BraceLevel
		ge.addLog("info", fmt.Sprintf("  Your preparations hold: if you Endure, %d%% of buildings fall and %.0f%% of stock is kept.",
			braceDestroyPct[h.BraceLevel], braceKeepFrac[h.BraceLevel]*100))
	}
	ge.harbingerFlavorLog(moment, kind)

	ge.harbingerHistory = append(ge.harbingerHistory, HarbingerRecord{
		Age: h.Age, Name: def.Name,
		EpochKey: h.EpochKey, TargetEpochKey: epochKey, TargetEpochName: target.Name,
		Outcome: outcome, FalseProphet: h.FalseProphet, AnnouncedTier: h.AnnouncedTier,
		AppeaseLevel: h.AppeaseLevel, BraceLevel: h.BraceLevel, Invited: h.Invited,
		Tick: ge.tick,
	})
	ge.harbinger = nil
}

// clearHarbingerRun drops all per-run harbinger state: the live harbinger, the
// once-per-epoch arrivals, the invite and the Brace level handed to a pending
// catastrophe. Called by Succumb, DoPrestige and Reset under the write lock.
// The outcome history is the caller's business (it follows epochEventHistory).
func (ge *GameEngine) clearHarbingerRun() {
	ge.harbinger = nil
	ge.harbingerArrived = make(map[string]bool)
	ge.catastropheInvited = false
	ge.pendingBraceLevel = 0
}

// --- View ---------------------------------------------------------------------

// harbingerView builds GameState.Harbinger; nil when none is present.
// Read-only: safe under the read lock.
func (ge *GameEngine) harbingerView() *HarbingerView {
	h := ge.harbinger
	if h == nil {
		return nil
	}
	def, _ := config.HarbingerFor(h.Age)
	out := ge.catastropheOutlook()
	v := &HarbingerView{
		Key: def.Key, Name: def.Name, Description: def.Description,
		Age: h.Age, AgeName: config.AgeByKey()[h.Age].Name,
		AppeaseLabel: def.AppeaseLabel, BraceLabel: def.BraceLabel, InviteLabel: def.InviteLabel,
		Numeric:         def.ForecastPrecision == config.ForecastNumeric,
		TargetEpochKey:  h.TargetEpoch,
		TargetEpochName: config.EpochByKey()[h.TargetEpoch].Name,
		Lines:           append([]string(nil), h.Lines...),
		Tier:            ge.harbingerDisplayTier(out.Tier),
		Probability:     out.Probability,
		AppeaseLevel:    h.AppeaseLevel,
		BraceLevel:      h.BraceLevel,
		Invited:         h.Invited,
		AppeaseBlocked:  ge.appeaseBlocked(),
		BraceBlocked:    ge.braceBlocked(),
		InviteBlocked:   ge.inviteBlocked(),
	}
	if v.AppeaseBlocked == "" {
		v.AppeaseCost = ge.harbingerAppeaseCost(h.AppeaseLevel + 1)
		v.AppeaseAffordable = len(v.AppeaseCost) > 0 && ge.Resources.CanAfford(v.AppeaseCost)
	}
	if v.BraceBlocked == "" {
		v.BraceCost = ge.harbingerBraceCost(h.BraceLevel + 1)
		v.BraceAffordable = len(v.BraceCost) > 0 && ge.Resources.CanAfford(v.BraceCost)
	}
	v.EndureDestroyPct = braceDestroyPct[h.BraceLevel]
	v.EndureKeepPct = int(math.Round(braceKeepFrac[h.BraceLevel] * 100))
	next := h.BraceLevel
	if next < HarbingerMaxBrace {
		next++
	}
	v.NextEndureDestroyPct = braceDestroyPct[next]
	v.NextEndureKeepPct = int(math.Round(braceKeepFrac[next] * 100))
	return v
}

// --- Persistence --------------------------------------------------------------

// harbingerSaveCopy returns a deep copy of the live harbinger for a save, or nil.
func (ge *GameEngine) harbingerSaveCopy() *HarbingerSave {
	if ge.harbinger == nil {
		return nil
	}
	h := *ge.harbinger
	h.Lines = append([]string(nil), ge.harbinger.Lines...)
	return &h
}

// restoreHarbingerState loads the harbinger fields from a save (old saves have
// none and load clean), then lets a harbinger arrive if the save sits in the
// last age of an epoch without one. Under the write lock, after the epoch and
// catastrophe state are restored.
func (ge *GameEngine) restoreHarbingerState(save *GameSave) {
	ge.harbinger = nil
	if save.Harbinger != nil {
		h := *save.Harbinger
		h.Lines = append([]string(nil), save.Harbinger.Lines...)
		if h.AppeaseLevel > HarbingerMaxAppease {
			h.AppeaseLevel = HarbingerMaxAppease
		}
		if h.BraceLevel > HarbingerMaxBrace {
			h.BraceLevel = HarbingerMaxBrace
		}
		if _, ok := config.HarbingerFor(h.Age); ok {
			ge.harbinger = &h
		}
	}
	ge.harbingerArrived = copyBoolMap(save.HarbingerArrived)
	if ge.harbingerArrived == nil {
		ge.harbingerArrived = make(map[string]bool)
	}
	ge.catastropheInvited = save.CatastropheInvited
	ge.pendingBraceLevel = save.PendingBraceLevel
	if ge.pendingBraceLevel < 0 || ge.pendingBraceLevel > HarbingerMaxBrace || ge.pendingCatastrophe == "" {
		ge.pendingBraceLevel = 0
	}
	ge.harbingerHistory = append([]HarbingerRecord(nil), save.HarbingerHistory...)
	ge.maybeHarbingerArrive()
}

// harbingerCostText renders a cost map in config order: "375 faith, 120 culture".
func harbingerCostText(cost map[string]float64) string {
	var parts []string
	for _, def := range config.BaseResources() {
		if v, ok := cost[def.Key]; ok {
			parts = append(parts, fmt.Sprintf("%.0f %s", v, def.Key))
		}
	}
	return strings.Join(parts, ", ")
}

// capFirst upper-cases the first letter: "the Oracle" → "The Oracle".
func capFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
