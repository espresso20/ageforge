package game

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/flavor"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// The Harbinger (Phase 9 follow-up).
//
// A harbinger comes only when doom is on its way. On entering an era from the
// Iron Era on, a hidden roll decides whether a doom is fated there (fate.go);
// if it is, a harbinger thread starts some while before the strike, with the
// current age's roster figure (config.HarbingerFor), and lasts until the doom
// strikes or is lifted. The speaker changes with each age the thread lives
// through, so a doom foretold late in the Classical Age passes from the Oracle
// to the Town Crier. The thread never blocks anything and never expires. What
// the figure says about WHEN follows its config.ForecastTiming (fateWhen).
//
// The Cosmic Era has no outgoing transition and no fate: its passage is
// prestige, and its thread (TargetEpoch "") warns of the Last Passage
// (last_passage.go) from the era's first age, as it always did, resolving when
// a prestige from the Cosmic Era completes.
//
// Answers belong to the doom, not the figure, and carry across handoffs:
//
//   - Appease (faith + culture): each level multiplies the REAL strike chance
//     by harbingerAppeaseFactor. Two levels; the second costs double.
//   - Brace (the epoch's core resources): softens an Endure if the doom
//     strikes. Two levels, stored on the pending catastrophe so it still
//     applies when Endure is chosen later.
//   - Invite: makes the strike certain (it still comes when it was fated to).
//     Free and final; Appease is pointless afterwards and refuses. In an era
//     where nothing can be fated (the Stone Era) all three are refused.
//
// False prophets come from the fate too (fate.go): a false thread claims
// medium or high whatever the odds, every figure repeats the claim, and it is
// revealed once its foretold window passes. The claim is kept as a fixed
// multiple of what a real doom's chance would be (ClaimFactor), so Appease, a
// faith change or an Invite move a false warning exactly as they move a true
// one, and a false thread that reaches a figure who prints odds (the
// Industrial Age) prints the claimed figure.
//
// Randomness: a thread's start draws the arrival and warning lines from the
// engine's flavor Stream (a Last Passage thread first draws its false-prophet
// roll, always, and the claim only for a false thread). A handoff draws two
// lines, each action and the resolution one. All of it comes from the seeded
// ge.rng under the write lock.
//
// Unexported methods expect the write lock, except the read-only helpers used
// by GetState (harbingerView, harbingerDisplay and the cost functions), which
// are safe under the read lock.

const (
	// HarbingerMaxAppease and HarbingerMaxBrace are the level caps.
	HarbingerMaxAppease = 2
	HarbingerMaxBrace   = 2

	// harbingerAppeaseFactor multiplies the catastrophe chance per Appease level:
	// 0.6 at level 1, 0.36 at level 2.
	harbingerAppeaseFactor = 0.6

	// Costs belong to the thread, not the player's current caps or age, so the
	// price is the same in every age of the epoch. Level 2 costs double.
	//
	// Appease: harbingerAppeaseIncomeShare of what a player who invests
	// moderately in faith (and culture) makes over the thread's ages at their
	// pacing targets: config.FlowIncome × config.AgeTargetTicks, summed over
	// harbingerAppeaseAges. Faith is a flow resource at hand-set rates and the
	// market sells none, so the price has to follow what the ages produce; the
	// old price (15% of the passage storage) was out of reach in most threads
	// once the pacing rebalance made ages short. At a quarter, with income
	// growing through the thread, level 1 comes partway through and level 1
	// plus level 2 (three quarters) before the passage. Faith also drives the
	// roll (faith fill bands): worst case, paying drops the fill from the top
	// band to the bottom, raising the base chance from 12% to 18% (×1.5), and
	// ×0.6 still leaves 0.9× of where it started, so Appease always lowers
	// the odds.
	//
	// Brace: 12% of the most the epoch asks of each core resource across its
	// remaining advances (into its later ages and into the next epoch), for
	// resources the player has held since the epoch began. It softens a
	// catastrophe that may not come, so it is priced under what the epoch
	// asks, but it draws on several resources at once.
	harbingerAppeaseIncomeShare = 0.25
	harbingerBraceCostFrac      = 0.12
)

// Endure numbers by Brace level (index 0 = unbraced): share of destroyable
// buildings destroyed, in percent, and share of each stored resource kept.
var (
	braceDestroyPct = [HarbingerMaxBrace + 1]int{20, 15, 10}
	braceKeepFrac   = [HarbingerMaxBrace + 1]float64{endureResourceKeep, 0.30, 0.45}
)

// harbingerClaimBase is the base chance a false thread claims for its tier:
// the mid-faith and low-faith chances that the real odds produce.
var harbingerClaimBase = map[CatastropheTier]float64{
	CatastropheTierMedium: 0.15,
	CatastropheTierHigh:   0.18,
}

// Harbinger outcomes (HarbingerRecord.Outcome).
const (
	HarbingerOutcomeFulfilled   = "fulfilled"   // invited, and it came
	HarbingerOutcomeVindicated  = "vindicated"  // warned, not invited, and it came
	HarbingerOutcomeSpared      = "spared"      // a true thread, and it did not come
	HarbingerOutcomeDiscredited = "discredited" // a false thread, and it did not come
)

// HarbingerSave is the live harbinger thread, persisted as GameSave.Harbinger.
type HarbingerSave struct {
	// Age is the current speaker's age; it selects the roster entry and the
	// voice of every line.
	Age string `json:"age"`
	// Chain lists the ages whose figures have spoken in this thread, first to
	// current.
	Chain []string `json:"chain,omitempty"`
	// EpochKey is the thread's epoch. TargetEpoch is the epoch whose doom it
	// warns of: the same epoch (the doom strikes inside it), or "" for the
	// Cosmic Era's Last Passage. Saves written before fates existed hold the
	// next epoch here; restoreHarbingerState rewrites it.
	EpochKey    string `json:"epoch_key"`
	TargetEpoch string `json:"target_epoch"`
	// When is what the current figure says about when the doom falls
	// (WhenUntold, WhenThisAge, WhenThisEra).
	When string `json:"when,omitempty"`
	// FalseProphet is the thread's one roll. Never shown to the player.
	FalseProphet bool `json:"false_prophet,omitempty"`
	// AnnouncedTier is the tier of the latest warning line: the real tier for
	// a true thread, the claimed tier for a false one.
	AnnouncedTier CatastropheTier `json:"announced_tier"`
	// ClaimFactor (false threads only) is claimed ÷ real chance at the start;
	// the claimed chance is always the real one times this.
	ClaimFactor  float64 `json:"claim_factor,omitempty"`
	AppeaseLevel int     `json:"appease_level,omitempty"`
	BraceLevel   int     `json:"brace_level,omitempty"`
	// Invited is set by the Invite action. catastropheInvited is consumed when
	// the invite is honoured, so the thread keeps its own copy to tell
	// Fulfilled from Vindicated at resolution.
	Invited bool `json:"invited,omitempty"`
	// Lines are the current figure's arrival and warning lines.
	Lines       []string `json:"lines,omitempty"`
	ArrivedTick int      `json:"arrived_tick,omitempty"`
}

// HarbingerRecord is a resolved thread, for the Epoch overlay.
type HarbingerRecord struct {
	// Age and Name are the last figure's, the one who saw the passage.
	Age  string `json:"age"`
	Name string `json:"name"`
	// Chain lists every age whose figure spoke, first to last.
	Chain           []string        `json:"chain,omitempty"`
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
	// When is the last figure's timing statement.
	When string `json:"when,omitempty"`
	// The doom's timing, known once it resolved (0 for a Last Passage
	// thread): the tick the thread began, the tick the era was entered, the
	// era's expected length, and the tick the doom was fated for (a false
	// prophet's foretold moment). AtAdvance: it came at an advance the player
	// reached first.
	ArrivedTick int  `json:"arrived_tick,omitempty"`
	EntryTick   int  `json:"entry_tick,omitempty"`
	Window      int  `json:"window,omitempty"`
	StrikeTick  int  `json:"strike_tick,omitempty"`
	AtAdvance   bool `json:"at_advance,omitempty"`
}

// HarbingerView is the UI's picture of the live thread (GameState.Harbinger).
// It deliberately carries no false-prophet flag.
type HarbingerView struct {
	Key, Name, Description, Age, AgeName  string
	AppeaseLabel, BraceLabel, InviteLabel string
	// Earlier names the figures who spoke before the current one, in order.
	Earlier []string
	// Numeric is true from the Industrial Age on: the panel prints Probability.
	Numeric bool
	// LastPassage is true for the final epoch's thread, whose passage is
	// prestige: TargetEpochKey is "" and TargetEpochName is "the Last Passage".
	// PassageCame is true once that passage has struck and waits for a choice.
	LastPassage bool
	PassageCame bool
	// TargetEpochKey is the epoch whose doom the thread warns of (the current
	// one; "" for the Last Passage), for the engine's own readers (bots,
	// tests).
	TargetEpochKey string
	// TargetEpochName is what the warning names on screen: "impending doom"
	// for an era's doom, "the Last Passage" for the final epoch's. It is
	// never an era's name (harbingerWarningText).
	TargetEpochName string
	// When is what the current figure says about when the doom falls
	// (WhenUntold, WhenThisAge, WhenThisEra); WhenText says it in words
	// ("before this age is out", "before the Iron Era ends", "" for no
	// word). Always WhenUntold for the Last Passage, which comes at prestige.
	When     string
	WhenText string
	Lines    []string
	// Tier and Probability are what the warning says: the live real odds for a
	// true thread, the claimed odds for a false one. Probability is only shown
	// when Numeric.
	Tier         CatastropheTier
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
	// Endure numbers at the current and the next Brace level, the garrison
	// included (see GarrisonPct).
	EndureDestroyPct     int
	EndureKeepPct        int
	NextEndureDestroyPct int
	NextEndureKeepPct    int
	// GarrisonPct is the share (percent) the army's garrison would blunt of
	// the catastrophe's losses on top of Brace, against the threat of the age
	// it strikes in; 0 with no soldiers. GarrisonCapped reports that the
	// combined Brace + garrison cap (config.EndureReductionCap) trims it.
	GarrisonPct    int
	GarrisonCapped bool
	// The Last Passage's Endure keeps a share of the run's prestige points
	// instead: at the current and the next Brace level, in percent.
	EndurePointsPct     int
	NextEndurePointsPct int
}

// --- Thread start and handoff -----------------------------------------------------

// harbingerOnAgeAdvance runs at the end of every age advance: the new age's
// figure takes up a running thread; otherwise the Cosmic Era's Last Passage
// thread starts, or the new age's lead brings a fated doom's harbinger (a
// longer age's lead can reach back past now). An epoch transition has already
// rolled the new era's fate. Under the write lock.
func (ge *GameEngine) harbingerOnAgeAdvance() {
	if h := ge.harbinger; h != nil && h.EpochKey == ge.currentEpoch {
		if h.Age != ge.age {
			ge.harbingerHandoff()
		}
		return
	}
	if config.IsFinalEpoch(ge.currentEpoch) {
		ge.maybeLastPassageArrive()
		return
	}
	if ge.fateHarbingerDue() {
		ge.fateArrive()
	}
}

// harbingerTickCheck is the harbinger's tick hook. In the final epoch it
// starts the Last Passage thread on the first tick that finds none (a
// Succumb, a prestige, a load), one check per epoch (harbingerCheckedEpoch).
// Anywhere else it runs the era's fate: roll it if missing, bring the
// harbinger when due, strike at the fated tick (fateTick).
func (ge *GameEngine) harbingerTickCheck() {
	if !config.IsFinalEpoch(ge.currentEpoch) {
		ge.fateTick()
		return
	}
	ge.fate = nil
	if ge.harbinger != nil || ge.harbingerCheckedEpoch == ge.currentEpoch {
		return
	}
	ge.harbingerCheckedEpoch = ge.currentEpoch
	ge.maybeLastPassageArrive()
}

// maybeLastPassageArrive starts the final epoch's thread if none is running
// and none has run this epoch this run.
func (ge *GameEngine) maybeLastPassageArrive() {
	if ge.harbinger != nil || ge.harbingerArrived[ge.currentEpoch] {
		return
	}
	ge.harbingerArriveLastPassage()
}

// harbingerArriveLastPassage starts the Last Passage thread with the current
// age's figure, if prestige from here can bring it. Skips the once-per-epoch
// check; maybeLastPassageArrive owns it. Reports whether one arrived.
func (ge *GameEngine) harbingerArriveLastPassage() bool {
	out := ge.catastropheOutlook()
	if out.Passage != PassagePrestige || !out.Possible {
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
	// depend on the age's chance (zero in the Cosmic Era).
	falseProphet := rng.Float64() < def.FalseProphetChance
	announced := out.Tier
	claim := 0.0
	if falseProphet {
		announced = CatastropheTierMedium
		if rng.Intn(2) == 1 {
			announced = CatastropheTierHigh
		}
		claim = harbingerClaimBase[announced] / out.Probability
	}

	ge.harbinger = &HarbingerSave{
		Age:           def.Age,
		Chain:         []string{def.Age},
		EpochKey:      ge.currentEpoch,
		TargetEpoch:   "",
		FalseProphet:  falseProphet,
		AnnouncedTier: announced,
		ClaimFactor:   claim,
		ArrivedTick:   ge.tick,
	}
	ge.harbinger.Lines = ge.harbingerSpeak(def, announced)

	ge.addLog("event", fmt.Sprintf("⚑ %s has come, warning of %s. Type 'harbinger' to answer.",
		capFirst(def.Name), harbingerWarningText("")))
	ge.harbingerLogLines()
	ge.publishHarbinger(def, "", false)
	return true
}

// harbingerHandoff passes the running thread to the current age's figure: a
// fresh arrival and warning line in the new voice, and its own word on when.
// The claim, the levels and the invite stay with the thread.
func (ge *GameEngine) harbingerHandoff() {
	h := ge.harbinger
	def, ok := config.HarbingerFor(ge.age)
	if !ok {
		return
	}
	h.Age = def.Age
	h.Chain = append(h.Chain, def.Age)
	warning := harbingerWarningText("")
	if h.TargetEpoch != "" {
		h.When = ge.fateWhen(def)
		warning = ge.fateWarningText(h)
	}
	tier, _ := ge.harbingerDisplay()
	h.AnnouncedTier = tier
	h.Lines = ge.harbingerSpeak(def, tier)

	ge.addLog("event", fmt.Sprintf("⚑ %s takes up the warning of %s. Type 'harbinger' to answer.",
		capFirst(def.Name), warning))
	ge.harbingerLogLines()
	ge.publishHarbinger(def, h.TargetEpoch, true)
}

// harbingerWarningText is what a thread warns of, without its timing, for
// the view and the Last Passage's log lines: "impending doom" for an era's
// doom, or "the Last Passage" when targetEpoch is "" (the final epoch, whose
// passage is prestige). It never names an era: a wild man at the last fire
// could not know what anything to come will be called.
func harbingerWarningText(targetEpoch string) string {
	if targetEpoch == "" {
		name, _ := config.LastPassageInfo()
		return "the" + strings.TrimPrefix(name, "The")
	}
	return "impending doom"
}

// harbingerSpeak draws the arrival and warning lines for def at tier.
func (ge *GameEngine) harbingerSpeak(def config.HarbingerDef, tier CatastropheTier) []string {
	rng := ge.gameRNG()
	stream := ge.flavorStream()
	arrival := stream.Line(flavor.Request{Moment: flavor.HarbingerArrival, Age: def.Age, Subject: def.Name}, rng)
	warning := stream.Line(flavor.Request{Moment: flavor.HarbingerWarning, Age: def.Age, Subject: def.Name, Kind: string(tier)}, rng)
	var lines []string
	for _, l := range []string{arrival, warning} {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

// harbingerLogLines writes the current figure's lines to the log in gray.
func (ge *GameEngine) harbingerLogLines() {
	for _, l := range ge.harbinger.Lines {
		ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", l))
	}
}

// publishHarbinger tells the UI a figure has spoken (toast). Bus handlers run
// under the write lock and must not call back into the engine.
func (ge *GameEngine) publishHarbinger(def config.HarbingerDef, targetEpoch string, handoff bool) {
	ge.Bus.Publish(EventData{
		Type: EventHarbingerArrived,
		Payload: map[string]interface{}{
			"harbinger_key":  def.Key,
			"harbinger_name": def.Name,
			"epoch_key":      targetEpoch,
			"handoff":        handoff,
		},
	})
}

// summonHarbinger is the dev console's /harbinger: a harbinger arrives now
// with the current age's figure. In an era that can be fated it fates a doom
// that strikes one lead from now (the era's doom reopened if it had
// resolved); in the Stone Era, where nothing can be fated, it sends a false
// prophet; in the final epoch it starts the Last Passage thread. Takes the
// write lock.
func (ge *GameEngine) summonHarbinger() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if ge.harbinger != nil {
		return fmt.Errorf("A harbinger is already here.")
	}
	// /age jumps the age without the epoch; line them up first.
	if ep := config.EpochForAge(ge.age); ep != ge.currentEpoch {
		ge.currentEpoch = ep
	}
	if config.IsFinalEpoch(ge.currentEpoch) {
		if !ge.harbingerArriveLastPassage() {
			return fmt.Errorf("The next passage cannot bring a catastrophe.")
		}
		return nil
	}
	ge.ensureFate()
	f := ge.fate
	if f.LeadFrac <= 0 {
		f.LeadFrac = (harbingerLeadMin + harbingerLeadMax) / 2
	}
	f.Fated = config.FateAllowed(f.EpochKey)
	f.FalseProphet = !f.Fated
	if f.FalseProphet && f.Claim == "" {
		f.Claim = CatastropheTierMedium
	}
	f.Invited, f.Resolved, f.ResolvedTick, f.AtAdvance = false, "", 0, false
	f.StrikeTick = ge.tick + int(f.LeadFrac*expectedAgeTicks(ge.age))
	delete(ge.harbingerArrived, f.EpochKey)
	if !ge.fateArrive() {
		return fmt.Errorf("No harbinger could come in this age.")
	}
	return nil
}

// SummonHarbingerForTest is a test hook for other packages (the UI's panel
// and theme-sweep tests, the smoke report's price table): it places the
// engine in age, with that age's epoch and unlocks, and brings a harbinger
// there the way the dev console's /harbinger does (a fated doom, a false
// prophet in the Stone Era, the Last Passage thread in the final epoch). Not
// reachable from play. Takes the write lock.
func (ge *GameEngine) SummonHarbingerForTest(age string) error {
	ge.mu.Lock()
	if _, ok := config.AgeByKey()[age]; !ok {
		ge.mu.Unlock()
		return fmt.Errorf("Unknown age '%s'.", age)
	}
	ge.age = age
	ge.currentEpoch = config.EpochForAge(age)
	ge.harbinger = nil
	for _, a := range config.AgeOrder() {
		ge.applyAgeUnlocks(a)
		if a == age {
			break
		}
	}
	ge.mu.Unlock()
	return ge.summonHarbinger()
}

// --- Odds ---------------------------------------------------------------------

// harbingerAppeaseMultiplier is the factor Appease applies to the strike (or
// Last Passage) chance: 1, 0.6 or 0.36. Read-only.
func (ge *GameEngine) harbingerAppeaseMultiplier() float64 {
	if ge.harbinger == nil || ge.harbinger.AppeaseLevel <= 0 {
		return 1
	}
	return detmath.Pow(harbingerAppeaseFactor, float64(ge.harbinger.AppeaseLevel))
}

// harbingerDisplay is what the live thread's warning says now: for an era's
// doom the chance it strikes (strikeChance), for the Last Passage the chance
// at prestige; the real tier and chance for a true thread, the claimed ones
// (real × ClaimFactor, never "none") for a false one. With no thread it is
// the Last Passage's odds. Read-only.
func (ge *GameEngine) harbingerDisplay() (CatastropheTier, float64) {
	h := ge.harbinger
	var p float64
	tierOf := catastropheTierFor
	if h != nil && h.TargetEpoch != "" {
		p, tierOf = ge.strikeChance(), fateTierFor
	} else {
		out := ge.catastropheOutlook()
		if out.Passage != PassagePrestige {
			return CatastropheTierNone, 0
		}
		p = out.Probability
	}
	if h == nil || !h.FalseProphet {
		return tierOf(p), p
	}
	p *= h.ClaimFactor
	if p > 1 {
		p = 1
	}
	t := tierOf(p)
	if t == CatastropheTierNone {
		t = CatastropheTierLow
	}
	return t, p
}

// --- Costs --------------------------------------------------------------------

// harbingerAdvanceAges are the ages still to be entered from epochKey's first
// age through the passage: its later ages and the next epoch's first age. In
// the final epoch, whose passage is prestige, that is its own later ages.
func harbingerAdvanceAges(epochKey string) []string {
	ep, ok := config.EpochByKey()[epochKey]
	if !ok || len(ep.Ages) == 0 {
		return nil
	}
	out := append([]string(nil), ep.Ages[1:]...)
	if next, ok := config.NextEpoch(epochKey); ok && len(next.Ages) > 0 {
		out = append(out, next.Ages[0])
	}
	return out
}

// harbingerAppeaseAges are the ages whose income prices epochKey's Appease:
// every age of the epoch except the game's last, which has no advance to pace.
// In the final epoch, whose passage is prestige, a player who prestiges from
// its first age leaves before Appease is in reach; one who stays for the
// epoch gets the same timing as every other thread.
func harbingerAppeaseAges(epochKey string) []string {
	ep, ok := config.EpochByKey()[epochKey]
	if !ok {
		return nil
	}
	order := config.AgeOrder()
	last := order[len(order)-1]
	var out []string
	for _, a := range ep.Ages {
		if a != last {
			out = append(out, a)
		}
	}
	return out
}

// harbingerHeldSinceStart reports the resources unlocked by the time the
// player enters epochKey's first age (the ages' UnlockResources, cumulative),
// so a price built from them is payable in every age of the epoch.
func harbingerHeldSinceStart(epochKey string) map[string]bool {
	ep, ok := config.EpochByKey()[epochKey]
	held := map[string]bool{}
	if !ok || len(ep.Ages) == 0 {
		return held
	}
	for _, a := range config.Ages() {
		for _, r := range a.UnlockResources {
			held[r] = true
		}
		if a.Key == ep.Ages[0] {
			break
		}
	}
	return held
}

// harbingerAppeaseCost is the price of Appease level (1 or 2) in epochKey's
// thread, in faith and in culture (culture only if held since the epoch
// began): level × harbingerAppeaseIncomeShare of the resource's
// config.FlowIncome over harbingerAppeaseAges at their pacing targets, the
// level-1 figure rounded up to two significant figures. Pure.
func harbingerAppeaseCost(epochKey string, level int) map[string]float64 {
	held := harbingerHeldSinceStart(epochKey)
	cost := map[string]float64{}
	for _, k := range []string{"faith", "culture"} {
		if !held[k] {
			continue
		}
		income := 0.0
		for _, a := range harbingerAppeaseAges(epochKey) {
			// An income, not a timing window: FlowIncome is the per-tick
			// rate at the age's pacing target, so rate × target is what the
			// age makes whatever its pace (a faster age makes more per tick
			// for fewer ticks). It stays on the raw target on purpose;
			// expectedAgeTicks is for durations.
			income += float64(config.FlowIncome(k, a) * config.AgeTargetTicks(a))
		}
		if l1 := ceilSignificant(income*harbingerAppeaseIncomeShare, 2); l1 > 0 {
			cost[k] = l1 * float64(level)
		}
	}
	return cost
}

// ceilSignificant rounds v up to sig significant figures (v itself if v <= 0).
func ceilSignificant(v float64, sig int) float64 {
	if v <= 0 {
		return v
	}
	exp := int(math.Ceil(detmath.Log10(v))) - sig
	// Scale by an exact power of ten either way (see config.roundSignificant).
	if exp >= 0 {
		mag := detmath.Pow(10, float64(exp))
		return math.Ceil(v/mag-1e-9) * mag
	}
	mag := detmath.Pow(10, float64(-exp))
	return math.Ceil(float64(v*mag)-1e-9) / mag
}

// harbingerBraceBasis is, per core resource, the most any remaining advance of
// epochKey asks for it, for resources held since the epoch began, minus faith
// and culture (they belong to Appease). Pure.
func harbingerBraceBasis(epochKey string) map[string]float64 {
	held := harbingerHeldSinceStart(epochKey)
	byAge := config.AgeByKey()
	basis := map[string]float64{}
	for _, a := range harbingerAdvanceAges(epochKey) {
		for k, v := range byAge[a].ResourceReqs {
			if k == "faith" || k == "culture" || !held[k] {
				continue
			}
			if v > basis[k] {
				basis[k] = v
			}
		}
	}
	return basis
}

// harbingerBraceCost is the price of Brace level (1 or 2) in epochKey's
// thread: harbingerBraceCostFrac × level of each basis amount. Pure.
func harbingerBraceCost(epochKey string, level int) map[string]float64 {
	cost := map[string]float64{}
	for k, v := range harbingerBraceBasis(epochKey) {
		cost[k] = math.Ceil(v * harbingerBraceCostFrac * float64(level))
	}
	return cost
}

// --- Actions ------------------------------------------------------------------

// harbingerActionCheck returns the common refusal for all three actions.
func (ge *GameEngine) harbingerActionCheck() error {
	if ge.harbinger == nil {
		return fmt.Errorf("No harbinger is here. One comes only when doom is on its way.")
	}
	if ge.pendingLastPassage {
		return fmt.Errorf("The Last Passage has already come. Type 'catastrophe' to answer it.")
	}
	return nil
}

// lastPassageCame is the refusal every action shows while the Last Passage
// is pending: the answers belonged to the passage, which has been rolled.
const lastPassageCame = "the Last Passage has already come"

// harbingerPowerless is the refusal every action shows for a thread in an era
// where nothing can be fated (a false prophet in the Stone Era): no
// catastrophe can strike there, so there is nothing to answer.
func (ge *GameEngine) harbingerPowerless() string {
	h := ge.harbinger
	if h == nil || h.TargetEpoch == "" || config.FateAllowed(h.EpochKey) {
		return ""
	}
	return "no catastrophe can strike in the " + config.EpochByKey()[h.EpochKey].Name
}

// appeaseBlocked explains why Appease cannot be bought now, or "".
func (ge *GameEngine) appeaseBlocked() string {
	h := ge.harbinger
	switch {
	case h == nil:
		return "no harbinger is here"
	case ge.pendingLastPassage:
		return lastPassageCame
	case ge.harbingerPowerless() != "":
		return ge.harbingerPowerless()
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
	case ge.pendingLastPassage:
		return lastPassageCame
	case ge.harbingerPowerless() != "":
		return ge.harbingerPowerless()
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
	case ge.pendingLastPassage:
		return lastPassageCame
	case ge.harbingerPowerless() != "":
		return ge.harbingerPowerless()
	case h.Invited:
		return "already invited"
	}
	return ""
}

// shortfall lists what cost needs beyond what is stored, in config order, and
// says so when a price is larger than the resource's storage can hold.
func (ge *GameEngine) shortfall(cost map[string]float64) string {
	var parts []string
	for _, def := range config.BaseResources() {
		need, ok := cost[def.Key]
		if !ok {
			continue
		}
		have := ge.Resources.Get(def.Key)
		if have >= need {
			continue
		}
		part := textfmt.Number(math.Ceil(need-have)) + " more " + ResourceName(def.Key)
		if ge.Resources.GetStorage(def.Key) < need {
			part += fmt.Sprintf(" (your %s storage must reach %s first)", ResourceName(def.Key), textfmt.Number(need))
		}
		parts = append(parts, part)
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
		return fmt.Errorf("Cannot appease: %s.", why)
	}
	h := ge.harbinger
	level := h.AppeaseLevel + 1
	cost := harbingerAppeaseCost(h.EpochKey, level)
	if len(cost) == 0 {
		return fmt.Errorf("Cannot appease: there is nothing to offer.")
	}
	if !ge.Resources.Pay(cost) {
		return fmt.Errorf("Cannot afford to appease: you need %s.", ge.shortfall(cost))
	}
	h.AppeaseLevel = level
	def, _ := config.HarbingerFor(h.Age)
	ge.addLog("success", fmt.Sprintf("⚑ %s (Appease %d/%d): paid %s. The chance it strikes is now %.2fx its base.",
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
		return fmt.Errorf("Cannot brace: %s.", why)
	}
	h := ge.harbinger
	level := h.BraceLevel + 1
	cost := harbingerBraceCost(h.EpochKey, level)
	if len(cost) == 0 {
		return fmt.Errorf("Cannot brace: this era asks nothing you can stockpile.")
	}
	if !ge.Resources.Pay(cost) {
		return fmt.Errorf("Cannot afford to brace: you need %s.", ge.shortfall(cost))
	}
	h.BraceLevel = level
	def, _ := config.HarbingerFor(h.Age)
	if h.TargetEpoch == "" {
		ge.addLog("success", fmt.Sprintf("⚑ %s (Brace %d/%d): paid %s. If the Last Passage comes and you Endure, you keep %.0f%% of the run's prestige points (not %.0f%%).",
			def.BraceLabel, level, HarbingerMaxBrace, harbingerCostText(cost), LastPassageKeepFor(level)*100, LastPassageKeep*100))
	} else {
		ge.addLog("success", fmt.Sprintf("⚑ %s (Brace %d/%d): paid %s. If the catastrophe comes and you Endure, %d%% of buildings fall (not %d%%) and %.0f%% of stock is kept (not %.0f%%).",
			def.BraceLabel, level, HarbingerMaxBrace, harbingerCostText(cost), braceDestroyPct[level], braceDestroyPct[0], braceKeepFrac[level]*100, braceKeepFrac[0]*100))
	}
	ge.harbingerFlavorLog(flavor.HarbingerBraced, "")
	return nil
}

// HarbingerInvite makes the doom certain: an era's doom still strikes when it
// was fated to (a false prophet's invented doom becomes real), the Last
// Passage comes at the next prestige. Free and final, at most once per era:
// there is one thread per era. Takes the write lock.
func (ge *GameEngine) HarbingerInvite() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if err := ge.harbingerActionCheck(); err != nil {
		return err
	}
	if why := ge.inviteBlocked(); why != "" {
		return fmt.Errorf("Cannot invite: %s.", why)
	}
	h := ge.harbinger
	h.Invited = true
	def, _ := config.HarbingerFor(h.Age)
	if h.TargetEpoch == "" {
		ge.inviteCatastrophe()
		ge.addLog("warning", fmt.Sprintf("⚑ %s: you have invited it. Your next prestige will bring the Last Passage. This cannot be undone.",
			def.InviteLabel))
	} else {
		if f := ge.fate; f != nil && f.EpochKey == h.EpochKey {
			f.Invited = true
			f.Fated = true // a false prophet's doom is real now; the record still says it lied
		}
		// The thread's own era, never one to come.
		when := fmt.Sprintf("before the %s ends", config.EpochByKey()[h.EpochKey].Name)
		if h.When == WhenThisAge {
			when = "before this age is out"
		}
		ge.addLog("warning", fmt.Sprintf("⚑ %s: you have invited it. The catastrophe will come %s. This cannot be undone.",
			def.InviteLabel, when))
	}
	ge.harbingerFlavorLog(flavor.HarbingerInvited, "")
	return nil
}

// harbingerFlavorLog writes one gray flavor line for moment in the current
// figure's voice. Draws from ge.rng.
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

// resolveHarbinger settles the live thread: epochKey is the era whose doom it
// warned of, and came reports whether that doom struck (a catastrophe is now
// pending). Called by fateStrike under the write lock.
//
// epochKey "" is the Last Passage (the final epoch's prestige): called by
// completePrestige once the passage is settled, with came telling whether it
// struck. Its Brace has already been applied to the points, so nothing is
// handed on.
func (ge *GameEngine) resolveHarbinger(epochKey string, came bool) {
	ge.settleHarbinger(epochKey, came, "")
}

// settleHarbinger is resolveHarbinger with the verdict line given (verdict,
// when not ""): a false prophet's reveal says which window passed. It logs
// the verdict in the last figure's voice, hands the Brace level to a pending
// catastrophe, records the outcome and clears the thread.
func (ge *GameEngine) settleHarbinger(epochKey string, came bool, verdict string) {
	h := ge.harbinger
	if h == nil {
		return
	}
	def, _ := config.HarbingerFor(h.Age)
	name := capFirst(def.Name)
	targetName := config.EpochByKey()[epochKey].Name
	spared := fmt.Sprintf("⚑ The doom %s foretold passed you by. The warning was real, and you were spared.", def.Name)
	discredited := fmt.Sprintf("⚑ The doom %s foretold never came. The warning had been invented from the start.", def.Name)
	if epochKey == "" {
		targetName, _ = config.LastPassageInfo()
		spared = fmt.Sprintf("⚑ The Last Passage opens, and nothing comes through it. %s's warning was real, and you were spared.", name)
		discredited = fmt.Sprintf("⚑ The Last Passage opens, and nothing comes through it. The warning had been invented from the start, and %s was the last to repeat it.", def.Name)
	}

	var outcome, msg, kind string
	var moment flavor.Moment
	if h.FalseProphet {
		kind = flavor.KindFalseProphet
	}
	switch {
	case came && h.Invited:
		outcome, moment = HarbingerOutcomeFulfilled, flavor.HarbingerFulfilled
		msg = fmt.Sprintf("⚑ You invited it, and it came, as %s said it would.", def.Name)
		if h.FalseProphet {
			msg = "⚑ You invited it, and it came, though the warning, it emerged, had been invented from the start."
		}
	case came:
		outcome, moment = HarbingerOutcomeVindicated, flavor.HarbingerVindicated
		msg = fmt.Sprintf("⚑ %s was right.", name)
		if h.FalseProphet {
			msg = fmt.Sprintf("⚑ %s was right, though it emerged that the warning had been invented from the start.", name)
		}
	case h.FalseProphet:
		outcome, moment = HarbingerOutcomeDiscredited, flavor.HarbingerDiscredited
		msg = discredited
	default:
		outcome, moment = HarbingerOutcomeSpared, flavor.HarbingerSpared
		msg = spared
	}
	if verdict != "" {
		msg = verdict
	}
	ge.addLog("event", msg)
	if came && h.BraceLevel > 0 && epochKey != "" {
		ge.pendingBraceLevel = h.BraceLevel
		ge.addLog("info", fmt.Sprintf("  Your preparations hold: if you Endure, %d%% of buildings fall and %.0f%% of stock is kept.",
			braceDestroyPct[h.BraceLevel], braceKeepFrac[h.BraceLevel]*100))
	}
	ge.harbingerFlavorLog(moment, kind)

	rec := HarbingerRecord{
		Age: h.Age, Name: def.Name, Chain: append([]string(nil), h.Chain...),
		EpochKey: h.EpochKey, TargetEpochKey: epochKey, TargetEpochName: targetName,
		Outcome: outcome, FalseProphet: h.FalseProphet, AnnouncedTier: h.AnnouncedTier,
		AppeaseLevel: h.AppeaseLevel, BraceLevel: h.BraceLevel, Invited: h.Invited,
		Tick: ge.tick, When: h.When, ArrivedTick: h.ArrivedTick,
	}
	if f := ge.fate; epochKey != "" && f != nil && f.EpochKey == epochKey {
		rec.EntryTick, rec.Window, rec.StrikeTick, rec.AtAdvance = f.EntryTick, f.Window, f.StrikeTick, f.AtAdvance
	}
	ge.harbingerHistory = append(ge.harbingerHistory, rec)
	ge.harbinger = nil
}

// cloneHarbingerHistory copies the records and each record's Chain, for a
// snapshot that must share nothing with the engine.
func cloneHarbingerHistory(in []HarbingerRecord) []HarbingerRecord {
	if in == nil {
		return nil
	}
	out := make([]HarbingerRecord, len(in))
	for i, r := range in {
		r.Chain = append([]string(nil), r.Chain...)
		out[i] = r
	}
	return out
}

// clearHarbingerRun drops all per-run harbinger state: the live thread, the
// era's fate, the once-per-epoch record, the invite and the Brace level
// handed to a pending catastrophe. Called by Succumb, DoPrestige and Reset
// under the write lock; the next tick rolls the new run's Stone Era fate. The
// outcome history is the caller's business (it follows epochEventHistory).
func (ge *GameEngine) clearHarbingerRun() {
	ge.harbinger = nil
	ge.fate = nil
	ge.harbingerArrived = make(map[string]bool)
	ge.harbingerCheckedEpoch = ""
	ge.catastropheInvited = false
	ge.pendingBraceLevel = 0
}

// --- View ---------------------------------------------------------------------

// harbingerView builds GameState.Harbinger; nil when no thread is running.
// Read-only: safe under the read lock.
func (ge *GameEngine) harbingerView() *HarbingerView {
	h := ge.harbinger
	if h == nil {
		return nil
	}
	def, _ := config.HarbingerFor(h.Age)
	tier, prob := ge.harbingerDisplay()
	v := &HarbingerView{
		Key: def.Key, Name: def.Name, Description: def.Description,
		Age: h.Age, AgeName: config.AgeByKey()[h.Age].Name,
		AppeaseLabel: def.AppeaseLabel, BraceLabel: def.BraceLabel, InviteLabel: def.InviteLabel,
		Numeric:         def.ForecastPrecision == config.ForecastNumeric,
		LastPassage:     h.TargetEpoch == "",
		PassageCame:     h.TargetEpoch == "" && ge.pendingLastPassage,
		TargetEpochKey:  h.TargetEpoch,
		TargetEpochName: harbingerWarningText(h.TargetEpoch),
		When:            h.When,
		WhenText:        harbingerWhenText(h),
		Lines:           append([]string(nil), h.Lines...),
		Tier:            tier,
		Probability:     prob,
		AppeaseLevel:    h.AppeaseLevel,
		BraceLevel:      h.BraceLevel,
		Invited:         h.Invited,
		AppeaseBlocked:  ge.appeaseBlocked(),
		BraceBlocked:    ge.braceBlocked(),
		InviteBlocked:   ge.inviteBlocked(),
	}
	for _, a := range h.Chain {
		if a == h.Age {
			break
		}
		if d, ok := config.HarbingerFor(a); ok {
			v.Earlier = append(v.Earlier, d.Name)
		}
	}
	if v.AppeaseBlocked == "" {
		v.AppeaseCost = harbingerAppeaseCost(h.EpochKey, h.AppeaseLevel+1)
		v.AppeaseAffordable = len(v.AppeaseCost) > 0 && ge.Resources.CanAfford(v.AppeaseCost)
	}
	if v.BraceBlocked == "" {
		v.BraceCost = harbingerBraceCost(h.EpochKey, h.BraceLevel+1)
		v.BraceAffordable = len(v.BraceCost) > 0 && ge.Resources.CanAfford(v.BraceCost)
	}
	next := h.BraceLevel
	if next < HarbingerMaxBrace {
		next++
	}
	// The Endure numbers count the garrison too, measured against the threat
	// of the current age: the doom strikes in it or later in the era (the
	// Last Passage at prestige, from here).
	cur, nxt := ge.endurePreview(h.BraceLevel, ge.age), ge.endurePreview(next, ge.age)
	v.EndureDestroyPct = int(math.Round(cur.DestroyPct))
	v.EndureKeepPct = int(math.Round(cur.KeepFrac * 100))
	v.NextEndureDestroyPct = int(math.Round(nxt.DestroyPct))
	v.NextEndureKeepPct = int(math.Round(nxt.KeepFrac * 100))
	v.GarrisonPct = int(math.Round(cur.Garrison * 100))
	v.GarrisonCapped = cur.Capped || nxt.Capped
	v.EndurePointsPct = int(math.Round(LastPassageKeepFor(h.BraceLevel) * 100))
	v.NextEndurePointsPct = int(math.Round(LastPassageKeepFor(next) * 100))
	return v
}

// harbingerWhenText says a thread's When in words: "before this age is out",
// "before the Iron Era ends" (the thread's own era), or "" for no word.
func harbingerWhenText(h *HarbingerSave) string {
	switch h.When {
	case WhenThisAge:
		return "before this age is out"
	case WhenThisEra:
		return fmt.Sprintf("before the %s ends", config.EpochByKey()[h.EpochKey].Name)
	}
	return ""
}

// --- Persistence --------------------------------------------------------------

// harbingerSaveCopy returns a deep copy of the live thread for a save, or nil.
func (ge *GameEngine) harbingerSaveCopy() *HarbingerSave {
	if ge.harbinger == nil {
		return nil
	}
	h := *ge.harbinger
	h.Lines = append([]string(nil), ge.harbinger.Lines...)
	h.Chain = append([]string(nil), ge.harbinger.Chain...)
	return &h
}

// restoreHarbingerState loads the harbinger fields from a save (old saves have
// none and load clean). It starts nothing and draws nothing, so a save loads
// to exactly the state it was written in; the next tick brings whatever is
// due. A live thread from before fates existed (it warned of the transition
// out of its era) is carried into the new rules: in an era that can be fated
// it becomes the warning of that era's doom (restoreFateState fated it to
// strike at the era's end at the latest); in the Stone Era, where nothing can
// be fated now, it is dropped. Under the write lock, after the epoch,
// catastrophe and fate state are restored.
func (ge *GameEngine) restoreHarbingerState(save *GameSave) {
	ge.harbinger = nil
	ge.harbingerCheckedEpoch = ""
	if save.Harbinger != nil {
		h := *save.Harbinger
		h.Lines = append([]string(nil), save.Harbinger.Lines...)
		h.Chain = append([]string(nil), save.Harbinger.Chain...)
		if len(h.Chain) == 0 {
			h.Chain = []string{h.Age}
		}
		if h.AppeaseLevel > HarbingerMaxAppease {
			h.AppeaseLevel = HarbingerMaxAppease
		}
		if h.BraceLevel > HarbingerMaxBrace {
			h.BraceLevel = HarbingerMaxBrace
		}
		if h.FalseProphet && h.ClaimFactor <= 0 {
			h.ClaimFactor = 1
		}
		keep := true
		if save.Fate == nil && h.TargetEpoch != "" {
			// Written before fates existed.
			if ge.fate != nil && ge.fate.EpochKey == h.EpochKey {
				h.TargetEpoch = h.EpochKey
				if def, ok := config.HarbingerFor(h.Age); ok {
					h.When = ge.fateWhen(def)
				}
			} else {
				keep = false
			}
		}
		if _, ok := config.HarbingerFor(h.Age); ok && keep {
			ge.harbinger = &h
		}
	}
	ge.harbingerArrived = copyBoolMap(save.HarbingerArrived)
	if ge.harbingerArrived == nil {
		ge.harbingerArrived = make(map[string]bool)
	}
	// The invite flag only arms the Last Passage now; an era's invite lives on
	// its fate.
	ge.catastropheInvited = save.CatastropheInvited && config.IsFinalEpoch(ge.currentEpoch)
	ge.pendingBraceLevel = save.PendingBraceLevel
	if ge.pendingBraceLevel < 0 || ge.pendingBraceLevel > HarbingerMaxBrace || ge.pendingCatastrophe == "" {
		ge.pendingBraceLevel = 0
	}
	ge.harbingerHistory = append([]HarbingerRecord(nil), save.HarbingerHistory...)
}

// harbingerCostText renders a cost map in config order: "375 faith, 120 culture".
func harbingerCostText(cost map[string]float64) string {
	var parts []string
	for _, def := range config.BaseResources() {
		if v, ok := cost[def.Key]; ok {
			parts = append(parts, Amount(v, def.Key))
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
