package game

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
)

// --- helpers ------------------------------------------------------------------

// prevAge returns the age before age in the canonical order.
func prevAge(t *testing.T, age string) string {
	t.Helper()
	order := config.AgeOrder()
	for i, a := range order {
		if a == age && i > 0 {
			return order[i-1]
		}
	}
	t.Fatalf("no age before %s", age)
	return ""
}

// epochAges returns the ages of epochKey.
func epochAges(t *testing.T, epochKey string) []string {
	t.Helper()
	ep, ok := config.EpochByKey()[epochKey]
	if !ok {
		t.Fatalf("unknown epoch %s", epochKey)
	}
	return ep.Ages
}

// threadEngine returns a seeded engine standing in epochKey's first age with
// the epoch's thread just started: by the first tick for the Stone Era (a new
// game), by the age advance into the epoch otherwise.
func threadEngine(t *testing.T, epochKey string, seed int64) *GameEngine {
	t.Helper()
	first := epochAges(t, epochKey)[0]
	var ge *GameEngine
	if first == "primitive_age" {
		ge = catEngine(t, first, seed)
		ge.harbingerTickCheck()
	} else {
		ge = catEngine(t, prevAge(t, first), seed)
		// Unlocks are cumulative in play; catEngine only has the Primitive
		// Age's.
		for _, a := range config.AgeOrder() {
			if a == first {
				break
			}
			ge.applyAgeUnlocks(a)
		}
		ge.advanceAge(first)
		// The transition into the epoch may have rolled its own catastrophe;
		// clear it so it does not block later rolls in the test.
		ge.pendingCatastrophe = ""
	}
	if ge.harbinger == nil {
		t.Fatalf("no harbinger thread in %s", first)
	}
	return ge
}

// walkTo advances ge age by age until it stands in age.
func walkTo(t *testing.T, ge *GameEngine, age string) {
	t.Helper()
	order := config.AgeOrder()
	for guard := 0; ge.age != age; guard++ {
		if guard > len(order) {
			t.Fatalf("could not walk to %s", age)
		}
		for i, a := range order {
			if a == ge.age {
				ge.advanceAge(order[i+1])
				break
			}
		}
	}
}

// setStock sets storage and amount for resources: key → {amount, storage}.
func setStock(ge *GameEngine, stock map[string][2]float64) {
	amounts := map[string]float64{}
	storage := map[string]float64{}
	for k, v := range stock {
		amounts[k] = v[0]
		storage[k] = v[1]
		ge.Resources.UnlockResource(k)
	}
	ge.Resources.LoadStorage(storage)
	ge.Resources.LoadAmounts(amounts)
}

// fillBraceStock gives every Brace resource of the running thread a full bank.
func fillBraceStock(ge *GameEngine, store float64) {
	stock := map[string][2]float64{}
	for k := range harbingerBraceBasis(ge.harbinger.EpochKey) {
		stock[k] = [2]float64{store, store}
	}
	setStock(ge, stock)
}

func logMessages(ge *GameEngine) []string {
	out := make([]string, len(ge.log))
	for i, e := range ge.log {
		out[i] = e.Message
	}
	return out
}

func countLogs(ge *GameEngine, substr string) int {
	return strings.Count(strings.Join(logMessages(ge), "\n"), substr)
}

// harbingerLines returns every "⚑" log line and the indented lines under it:
// the part of the log that must be seeded.
func harbingerLines(ge *GameEngine) []string {
	var out []string
	inHarb := false
	for _, m := range logMessages(ge) {
		switch {
		case strings.HasPrefix(m, "⚑"):
			inHarb = true
			out = append(out, m)
		case inHarb && strings.HasPrefix(m, "  "):
			out = append(out, m)
		default:
			inHarb = false
		}
	}
	return out
}

// flavorAfterVerdict returns the gray flavor line logged after the verdict
// (the last "⚑" line before the new epoch's thread starts), without markup.
func flavorAfterVerdict(t *testing.T, ge *GameEngine) string {
	t.Helper()
	msgs := logMessages(ge)
	verdict := -1
	for i, m := range msgs {
		if strings.HasPrefix(m, "⚑") && !strings.Contains(m, "has come") && !strings.Contains(m, "takes up") {
			verdict = i
		}
	}
	if verdict < 0 {
		t.Fatal("no harbinger verdict in the log")
	}
	for _, m := range msgs[verdict+1:] {
		if strings.HasPrefix(m, "  [gray]") {
			return strings.TrimSuffix(strings.TrimPrefix(m, "  [gray]"), "[-]")
		}
	}
	t.Fatal("no flavor line after the verdict")
	return ""
}

// fromMoment reports whether line carries one of m's catalog signatures.
func fromMoment(line string, m flavor.Moment) bool {
	for _, s := range flavor.Signatures(m) {
		if strings.Contains(line, s) {
			return true
		}
	}
	return false
}

func figureName(t *testing.T, age string) string {
	t.Helper()
	def, ok := config.HarbingerFor(age)
	if !ok {
		t.Fatalf("no roster entry for %s", age)
	}
	return def.Name
}

// --- thread start ------------------------------------------------------------------

// A new game's first tick starts the Stone Era thread with the Wild Man.
func TestHarbingerThreadStartsOnFirstTick(t *testing.T) {
	ge := NewGameEngine()
	var events []EventData
	ge.Bus.Subscribe(EventHarbingerArrived, func(e EventData) { events = append(events, e) })
	ge.doTick()
	h := ge.harbinger
	if h == nil || h.Age != "primitive_age" || h.EpochKey != "stone_era" || h.TargetEpoch != "iron_era" {
		t.Fatalf("thread after first tick = %+v", h)
	}
	if len(h.Lines) != 2 || !reflect.DeepEqual(h.Chain, []string{"primitive_age"}) {
		t.Errorf("lines %q chain %v", h.Lines, h.Chain)
	}
	if len(events) != 1 || events[0].Payload["handoff"] != false {
		t.Errorf("bus events = %+v", events)
	}
	if countLogs(ge, capFirst(figureName(t, "primitive_age"))+" has come") != 1 {
		t.Error("no arrival log line for the Wild Man")
	}
	ge.doTick()
	if len(events) != 1 {
		t.Error("a second tick started another thread")
	}
}

// Entering an epoch's first age starts its thread with that age's figure.
func TestHarbingerThreadStartsOnEnteringFirstAge(t *testing.T) {
	ge := catEngine(t, "bronze_age", 1)
	ge.advanceAge("iron_age")
	h := ge.harbinger
	if h == nil || h.Age != "iron_age" || h.EpochKey != "iron_era" || h.TargetEpoch != "steel_era" {
		t.Fatalf("thread = %+v", h)
	}
	st := ge.GetState()
	if st.Harbinger == nil || st.Harbinger.Name != figureName(t, "iron_age") || st.Harbinger.TargetEpochName != "Steel Era" {
		t.Errorf("GetState harbinger = %+v", st.Harbinger)
	}
}

// Each age advance within the epoch hands the warning to that age's figure:
// new lines and a toast, same thread, same levels.
func TestHarbingerSpeakerChangesEachAge(t *testing.T) {
	for _, tc := range []struct {
		epoch string
	}{{"stone_era"}, {"neon_era"}} {
		t.Run(tc.epoch, func(t *testing.T) {
			ge := threadEngine(t, tc.epoch, 2)
			ages := epochAges(t, tc.epoch)
			var events []EventData
			ge.Bus.Subscribe(EventHarbingerArrived, func(e EventData) { events = append(events, e) })

			// Levels bought from the first figure.
			ge.harbinger.AppeaseLevel, ge.harbinger.BraceLevel = 1, 2
			prevLines := ge.harbinger.Lines
			for i, age := range ages[1:] {
				ge.advanceAge(age)
				h := ge.harbinger
				if h == nil || h.Age != age || h.EpochKey != tc.epoch {
					t.Fatalf("after entering %s: thread = %+v", age, h)
				}
				if !reflect.DeepEqual(h.Chain, ages[:i+2]) {
					t.Errorf("chain after %s = %v", age, h.Chain)
				}
				if h.AppeaseLevel != 1 || h.BraceLevel != 2 {
					t.Errorf("levels lost at the handoff to %s: appease %d brace %d", age, h.AppeaseLevel, h.BraceLevel)
				}
				if len(h.Lines) != 2 || reflect.DeepEqual(h.Lines, prevLines) {
					t.Errorf("handoff to %s kept the old lines: %q", age, h.Lines)
				}
				prevLines = h.Lines
				if countLogs(ge, capFirst(figureName(t, age))+" takes up the warning") != 1 {
					t.Errorf("no handoff log line for %s", figureName(t, age))
				}
				if v := ge.GetState().Harbinger; v.Name != figureName(t, age) || len(v.Earlier) != i+1 {
					t.Errorf("view after %s: name %q earlier %v", age, v.Name, v.Earlier)
				}
			}
			if len(events) != len(ages)-1 {
				t.Errorf("handoff toasts = %d, want %d", len(events), len(ages)-1)
			}
			for _, e := range events {
				if e.Payload["handoff"] != true {
					t.Errorf("handoff event = %+v", e.Payload)
				}
			}
		})
	}
}

// Every epoch whose passage can bring a catastrophe has a thread, in every
// age: the transition into an epoch past the Iron gate, or in the final epoch
// prestige itself (the Last Passage).
func TestHarbingerThreadFollowsCatastropheGate(t *testing.T) {
	for _, ep := range config.Epochs() {
		next, hasNext := config.NextEpoch(ep.Key)
		want := (hasNext && config.CatastropheAllowed(next.Key)) ||
			(config.IsFinalEpoch(ep.Key) && config.CatastropheAllowed(ep.Key))
		var ge *GameEngine
		if ep.Ages[0] == "primitive_age" {
			ge = catEngine(t, "primitive_age", 3)
			ge.harbingerTickCheck()
		} else {
			ge = catEngine(t, prevAge(t, ep.Ages[0]), 3)
			ge.advanceAge(ep.Ages[0])
		}
		for _, age := range ep.Ages {
			walkTo(t, ge, age)
			ge.harbingerTickCheck()
			if got := ge.harbinger != nil; got != want {
				t.Errorf("%s (%s): thread present = %v, want %v", ep.Key, age, got, want)
			}
		}
	}
	ge := catEngine(t, "transcendent_age", 1)
	if err := ge.summonHarbinger(); err != nil || ge.harbinger.TargetEpoch != "" {
		t.Errorf("summoning in the final epoch: err %v thread %+v, want a Last Passage thread", err, ge.harbinger)
	}
}

func TestNoHarbingerWhenTransitionCannotRoll(t *testing.T) {
	ge := catEngine(t, "primitive_age", 1)
	ge.epochEventFired["iron_era"] = true // the transition roll already happened
	ge.harbingerTickCheck()
	ge.advanceAge("stone_age")
	if ge.harbinger != nil {
		t.Error("thread started although the next transition cannot roll")
	}
}

func TestHarbingerOncePerEpochPerRun(t *testing.T) {
	ge := threadEngine(t, "stone_era", 2)
	ge.harbinger = nil
	ge.harbingerTickCheck()
	ge.maybeHarbingerArrive()
	ge.advanceAge("stone_age")
	if ge.harbinger != nil {
		t.Error("a second thread started in the same epoch")
	}
}

// --- false prophets ------------------------------------------------------------------

// The thread rolls once, at its first figure, with that age's chance.
func TestFalseProphetRollPerThread(t *testing.T) {
	rate := func(age string, n int) (float64, map[CatastropheTier]int) {
		ge := catEngine(t, age, 11)
		tiers := map[CatastropheTier]int{}
		falseCount := 0
		for i := 0; i < n; i++ {
			ge.harbinger = nil
			if !ge.harbingerArrive() {
				t.Fatalf("%s: no thread", age)
			}
			h := ge.harbinger
			if h.FalseProphet {
				falseCount++
				tiers[h.AnnouncedTier]++
				if h.AnnouncedTier != CatastropheTierMedium && h.AnnouncedTier != CatastropheTierHigh {
					t.Fatalf("%s: false thread claimed %q", age, h.AnnouncedTier)
				}
			} else if h.AnnouncedTier != ge.catastropheOutlook().Tier {
				t.Fatalf("%s: true thread announced %q", age, h.AnnouncedTier)
			}
		}
		return float64(falseCount) / float64(n), tiers
	}
	for _, age := range []string{"primitive_age", "iron_age", "renaissance_age"} {
		def, _ := config.HarbingerFor(age)
		if def.FalseProphetChance == 0 {
			t.Fatalf("%s should have a false-prophet chance", age)
		}
		got, tiers := rate(age, 4000)
		if math.Abs(got-def.FalseProphetChance) > 0.025 {
			t.Errorf("%s: false rate %.4f, want ≈ %.4f", age, got, def.FalseProphetChance)
		}
		if tiers[CatastropheTierMedium] == 0 || tiers[CatastropheTierHigh] == 0 {
			t.Errorf("%s: want both medium and high claims, got %v", age, tiers)
		}
	}
	for _, age := range []string{"victorian_age", "modern_age", "cyberpunk_age"} {
		if got, _ := rate(age, 300); got != 0 {
			t.Errorf("%s: false rate %.4f, want 0", age, got)
		}
	}

	// Handoffs never re-roll: a false thread stays false, a true one true.
	for seed := int64(1); seed <= 60; seed++ {
		ge := threadEngine(t, "stone_era", seed)
		was := ge.harbinger.FalseProphet
		ge.advanceAge("stone_age")
		ge.advanceAge("bronze_age")
		if ge.harbinger.FalseProphet != was {
			t.Fatalf("seed %d: the handoff changed the false-prophet flag", seed)
		}
	}
}

// A false thread's claim is a fixed multiple of the real chance, repeated by
// every figure, moved by Appease like a real one, and printed as the claimed
// figure once a numeric figure takes over.
func TestFalseThreadClaim(t *testing.T) {
	ge := threadEngine(t, "steel_era", 1)
	setFaith(ge, 5e6, 1e7)
	real := ge.catastropheOutlook().Probability // mid faith: 15%
	ge.harbinger.FalseProphet = true
	ge.harbinger.AnnouncedTier = CatastropheTierHigh
	ge.harbinger.ClaimFactor = harbingerClaimBase[CatastropheTierHigh] / real

	if tier, p := ge.harbingerDisplay(); tier != CatastropheTierHigh || math.Abs(p-0.18) > 1e-9 {
		t.Errorf("claim = %s %.3f, want high 0.18", tier, p)
	}
	ge.advanceAge("colonial_age")
	if ge.harbinger.AnnouncedTier != CatastropheTierHigh {
		t.Errorf("colonial figure announced %q, want the same false high", ge.harbinger.AnnouncedTier)
	}
	ge.advanceAge("industrial_age")
	setStock(ge, map[string][2]float64{"faith": {5e6, 1e7}, "culture": {1e7, 1e7}})
	v := ge.GetState().Harbinger
	if !v.Numeric || v.Tier != CatastropheTierHigh || math.Abs(v.Probability-0.18) > 1e-9 {
		t.Errorf("industrial view of a false thread = numeric %v %s %.3f, want the claimed high 18%%", v.Numeric, v.Tier, v.Probability)
	}
	// Appease ×0.6 moves the claim as it moves the real odds (read live, so a
	// band change from the faith spend would count too).
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	wantP := ge.catastropheOutlook().Probability * ge.harbinger.ClaimFactor
	if _, p := ge.harbingerDisplay(); math.Abs(p-wantP) > 1e-9 || p >= 0.18 {
		t.Errorf("claim after appease = %.4f, want %.4f (< 0.18)", p, wantP)
	}
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	if tier, p := ge.harbingerDisplay(); tier != CatastropheTierHigh || p != 1 {
		t.Errorf("claim after invite = %s %.3f, want high 1", tier, p)
	}
}

// --- determinism ----------------------------------------------------------------------

func TestHarbingerSameSeedSameOutcome(t *testing.T) {
	type run struct {
		history []HarbingerRecord
		lines   []string
	}
	play := func(seed int64) run {
		ge := threadEngine(t, "stone_era", seed)
		ge.advanceAge("stone_age")
		ge.advanceAge("bronze_age")
		ge.advanceAge("iron_age")
		return run{ge.harbingerHistory, harbingerLines(ge)}
	}
	a, b := play(42), play(42)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("same seed, different threads:\n%+v\n%+v", a, b)
	}
	if len(a.history) != 1 || !reflect.DeepEqual(a.history[0].Chain, []string{"primitive_age", "stone_age", "bronze_age"}) {
		t.Errorf("history = %+v", a.history)
	}
	differs := false
	for s := int64(43); s < 53 && !differs; s++ {
		differs = !reflect.DeepEqual(a.lines, play(s).lines)
	}
	if !differs {
		t.Error("ten other seeds produced the same harbinger log as seed 42")
	}
}

// --- costs ---------------------------------------------------------------------------

// The price belongs to the passage: the same in every age of the epoch, and
// whatever the player's storage.
func TestHarbingerCostSameAcrossEpoch(t *testing.T) {
	for _, ep := range config.Epochs() {
		if next, ok := config.NextEpoch(ep.Key); !ok || !config.CatastropheAllowed(next.Key) {
			continue
		}
		ge := threadEngine(t, ep.Key, 1)
		first := ge.GetState().Harbinger
		if len(first.AppeaseCost) == 0 || len(first.BraceCost) == 0 {
			t.Fatalf("%s: empty costs %v / %v", ep.Key, first.AppeaseCost, first.BraceCost)
		}
		for k := range first.BraceCost {
			if !ge.Resources.IsUnlocked(k) || k == "faith" || k == "culture" {
				t.Errorf("%s: brace charges %s, not held since the epoch began", ep.Key, k)
			}
		}
		for k := range first.AppeaseCost {
			if !ge.Resources.IsUnlocked(k) {
				t.Errorf("%s: appease charges locked %s", ep.Key, k)
			}
		}
		for _, age := range ep.Ages[1:] {
			ge.advanceAge(age)
			fillBraceStock(ge, 1e15) // caps no longer matter
			v := ge.GetState().Harbinger
			if !reflect.DeepEqual(v.AppeaseCost, first.AppeaseCost) || !reflect.DeepEqual(v.BraceCost, first.BraceCost) {
				t.Errorf("%s at %s: costs %v / %v, first age %v / %v", ep.Key, age, v.AppeaseCost, v.BraceCost, first.AppeaseCost, first.BraceCost)
			}
		}
		if two := harbingerBraceCost(ep.Key, 2); len(two) != len(first.BraceCost) {
			t.Errorf("%s: level 2 brace resources differ", ep.Key)
		} else {
			for k, v := range first.BraceCost {
				if math.Abs(two[k]-2*v) > 1 {
					t.Errorf("%s: level 2 %s = %v, want double %v", ep.Key, k, two[k], v)
				}
			}
		}
	}
}

func TestHarbingerCostExamples(t *testing.T) {
	cases := []struct {
		epoch   string
		appease map[string]float64
		brace   map[string]float64
	}{
		// Appease is a quarter of what FlowIncome makes over the thread's
		// ages at their targets. Stone Era faith: 0.01, 0.03 and 0.07 a tick
		// over 450, 1,350 and 2,700 ticks = 234; a quarter is 58.5 → 59.
		{"stone_era", map[string]float64{"faith": 59}, map[string]float64{"food": 9600, "wood": 4800, "knowledge": 2400}},
		{"steel_era", map[string]float64{"faith": 74000, "culture": 770000}, map[string]float64{"knowledge": 3600000, "gold": 1800000, "steel": 288000}},
		// The Cosmic Era's passage is prestige; Appease counts its ages but
		// the last (Interstellar, Galactic, Quantum). Brace is priced off
		// the era's own advances, for resources held from Interstellar (dark
		// matter 13T, titanium 630B → 12%). Antimatter and quantum flux
		// arrive later.
		{"cosmic_era", map[string]float64{"faith": 1200000000, "culture": 19000000000}, map[string]float64{"dark_matter": 1560000000000, "titanium": 75600000000}},
	}
	for _, c := range cases {
		if got := harbingerAppeaseCost(c.epoch, 1); !reflect.DeepEqual(got, c.appease) {
			t.Errorf("%s appease = %v, want %v", c.epoch, got, c.appease)
		}
		if got := harbingerBraceCost(c.epoch, 1); !reflect.DeepEqual(got, c.brace) {
			t.Errorf("%s brace = %v, want %v", c.epoch, got, c.brace)
		}
	}
}

// --- Appease ------------------------------------------------------------------------

func TestAppeaseCostsLevelsAndOdds(t *testing.T) {
	ge := threadEngine(t, "steel_era", 4)
	setStock(ge, map[string][2]float64{"faith": {6e5, 6e5}, "culture": {1e7, 1e7}})

	// Level 1: 74K faith, 770K culture. Faith fill 0.877 → high band, base 12%.
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	if f, c := ge.Resources.Get("faith"), ge.Resources.Get("culture"); f != 526000 || c != 9.23e6 {
		t.Errorf("after level 1: faith %v culture %v", f, c)
	}
	if o := ge.CatastropheOutlook(); math.Abs(o.Probability-0.12*0.6) > 1e-9 {
		t.Errorf("level 1 probability %v, want %v", o.Probability, 0.12*0.6)
	}
	if countLogs(ge, "Appease 1/2") != 1 {
		t.Error("no Appease log line")
	}

	// Level 2 costs double. Fill 0.63 → mid band, base 15%.
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	if f := ge.Resources.Get("faith"); f != 378000 {
		t.Errorf("after level 2: faith %v", f)
	}
	if o := ge.CatastropheOutlook(); math.Abs(o.Probability-0.15*0.36) > 1e-9 {
		t.Errorf("level 2 probability %v, want %v", o.Probability, 0.15*0.36)
	}
	if st := ge.GetState(); st.Harbinger.AppeaseLevel != 2 || st.Harbinger.AppeaseBlocked == "" || st.Harbinger.AppeaseCost != nil {
		t.Errorf("view after cap = %+v", st.Harbinger)
	}
	if err := ge.HarbingerAppease(); err == nil || !strings.Contains(err.Error(), "as far as it goes") {
		t.Errorf("third appease: err = %v", err)
	}
	if f := ge.Resources.Get("faith"); f != 378000 {
		t.Errorf("refused appease deducted faith: %v", f)
	}
}

func TestAppeaseRefusesWhenUnaffordable(t *testing.T) {
	ge := threadEngine(t, "stone_era", 4) // appease: 59 faith
	setStock(ge, map[string][2]float64{"faith": {20, 40}})
	err := ge.HarbingerAppease()
	if err == nil || !strings.Contains(err.Error(), "39 more faith") || !strings.Contains(err.Error(), "storage must reach 59") {
		t.Fatalf("err = %v, want the shortfall and the storage it needs", err)
	}
	if ge.Resources.Get("faith") != 20 || ge.harbinger.AppeaseLevel != 0 {
		t.Error("a refused appease changed state")
	}
	if st := ge.GetState(); st.Harbinger.AppeaseAffordable {
		t.Error("view says affordable")
	}
}

func TestNoHarbingerActionsWithoutOne(t *testing.T) {
	ge := catEngine(t, "iron_age", 1)
	for name, act := range map[string]func() error{
		"appease": ge.HarbingerAppease, "brace": ge.HarbingerBrace, "invite": ge.HarbingerInvite,
	} {
		if err := act(); err == nil || !strings.Contains(err.Error(), "no harbinger") {
			t.Errorf("%s without a harbinger: err = %v", name, err)
		}
	}
	if ge.catastropheInvited {
		t.Error("invite armed without a harbinger")
	}
}

// Appease lowers the escalation threshold of the real roll.
func TestAppeaseChangesTheRealRoll(t *testing.T) {
	transition := func(appease bool) string {
		ge := threadEngine(t, "stone_era", 5)
		ge.advanceAge("stone_age")
		ge.advanceAge("bronze_age")
		// Faith fill ends at 0.5 either way (mid band), so only Appease differs.
		setStock(ge, map[string][2]float64{"faith": {15000, 30000}})
		if appease {
			setStock(ge, map[string][2]float64{"faith": {15059, 30000}}) // 59 paid → 15000
			if err := ge.HarbingerAppease(); err != nil {
				t.Fatal(err)
			}
		}
		// Bad roll, then an escalation draw of 0.2: under the unappeased 0.30
		// threshold, over the appeased 0.18.
		ge.rng = riggedRNG(0.99, 0.2)
		ge.advanceAge("iron_age")
		return ge.pendingCatastrophe
	}
	if got := transition(false); got != "iron_era" {
		t.Errorf("unappeased: pending = %q, want iron_era", got)
	}
	if got := transition(true); got != "" {
		t.Errorf("appeased: pending = %q, want none", got)
	}
}

// --- Brace ---------------------------------------------------------------------------

func TestBraceCostsAndSoftensALaterEndure(t *testing.T) {
	for _, tc := range []struct {
		level          int
		wantBuildings  int // of 20
		wantWoodOf1000 float64
	}{
		{0, 16, 150},
		{1, 17, 300},
		{2, 18, 450},
	} {
		t.Run(string(rune('0'+tc.level)), func(t *testing.T) {
			t.Cleanup(SetDataDirForTest(t.TempDir()))
			ge := threadEngine(t, "stone_era", 6)
			fillBraceStock(ge, 100000)
			for i := 1; i <= tc.level; i++ {
				stockBefore := ge.Resources.Get("wood")
				if err := ge.HarbingerBrace(); err != nil {
					t.Fatal(err)
				}
				want := 4800 * float64(i) // 12% of the 40000 wood the passage asks, × level
				if got := stockBefore - ge.Resources.Get("wood"); got != want {
					t.Errorf("level %d took %v wood, want %v", i, got, want)
				}
			}
			if tc.level == HarbingerMaxBrace {
				if err := ge.HarbingerBrace(); err == nil {
					t.Error("brace beyond the cap must be refused")
				}
			}
			ge.advanceAge("stone_age")
			ge.advanceAge("bronze_age")
			ge.rng = badThenEscalate()
			ge.advanceAge("iron_age")
			if ge.pendingCatastrophe != "iron_era" {
				t.Fatalf("no catastrophe: pending = %q", ge.pendingCatastrophe)
			}
			if ge.pendingBraceLevel != tc.level {
				t.Fatalf("pendingBraceLevel = %d, want %d", ge.pendingBraceLevel, tc.level)
			}

			if err := ge.SaveGame("braced"); err != nil {
				t.Fatal(err)
			}
			ge2 := NewGameEngine()
			if err := ge2.LoadGame("braced"); err != nil {
				t.Fatal(err)
			}
			if ge2.pendingBraceLevel != tc.level {
				t.Fatalf("pendingBraceLevel after load = %d, want %d", ge2.pendingBraceLevel, tc.level)
			}
			b := pickWorkerBuildings(t, "food", "knowledge")
			ge2.Buildings.counts = map[string]int{b[0]: 10, b[1]: 10}
			setStock(ge2, map[string][2]float64{"wood": {1000, 10000}})
			if err := ge2.Endure(); err != nil {
				t.Fatal(err)
			}
			if got := nonWonderTotal(ge2); got != tc.wantBuildings {
				t.Errorf("buildings after Endure = %d, want %d", got, tc.wantBuildings)
			}
			if got := ge2.Resources.Get("wood"); math.Abs(got-tc.wantWoodOf1000) > 1e-6 {
				t.Errorf("wood after Endure = %v, want %v", got, tc.wantWoodOf1000)
			}
			if ge2.pendingBraceLevel != 0 {
				t.Error("Endure did not consume the Brace")
			}
		})
	}
}

// --- Invite ----------------------------------------------------------------------------

func TestInviteForcesCatastropheAndDisablesAppease(t *testing.T) {
	ge := threadEngine(t, "iron_era", 7)
	setStock(ge, map[string][2]float64{"faith": {1e6, 1e6}})
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	if !ge.catastropheInvited || !ge.harbinger.Invited {
		t.Fatal("invite not armed")
	}
	if o := ge.CatastropheOutlook(); o.Probability != 1 || o.Tier != CatastropheTierHigh {
		t.Errorf("invited outlook = %+v", o)
	}
	if err := ge.HarbingerAppease(); err == nil || !strings.Contains(err.Error(), "invited") {
		t.Errorf("appease after invite: err = %v", err)
	}
	if ge.Resources.Get("faith") != 1e6 {
		t.Error("refused appease spent faith")
	}
	if err := ge.HarbingerInvite(); err == nil {
		t.Error("inviting twice must be refused")
	}
	v := ge.GetState().Harbinger
	if v.AppeaseBlocked == "" || v.InviteBlocked == "" || v.BraceBlocked != "" {
		t.Errorf("view after invite = %+v", v)
	}
	fillBraceStock(ge, 1e7)
	if err := ge.HarbingerBrace(); err != nil {
		t.Errorf("brace after invite: %v", err)
	}
	ge.advanceAge("classical_age")
	ge.advanceAge("medieval_age")
	if !ge.harbinger.Invited || !ge.catastropheInvited {
		t.Fatal("the invite did not survive the handoffs")
	}
	ge.rng = riggedRNG(0.0) // a good roll: the invite overrides it
	ge.advanceAge("renaissance_age")
	if ge.pendingCatastrophe != "steel_era" {
		t.Fatalf("invited catastrophe did not come: pending = %q", ge.pendingCatastrophe)
	}
	if ge.pendingBraceLevel != 1 {
		t.Errorf("Brace not handed to the invited catastrophe: %d", ge.pendingBraceLevel)
	}
	if n := len(ge.harbingerHistory); n != 1 || ge.harbingerHistory[0].Outcome != HarbingerOutcomeFulfilled {
		t.Errorf("history = %+v", ge.harbingerHistory)
	}
}

// --- resolution --------------------------------------------------------------------------

func TestHarbingerResolutionOutcomes(t *testing.T) {
	cases := []struct {
		name         string
		falseProphet bool
		invite       bool
		rng          []float64
		wantOutcome  string
		wantMoment   flavor.Moment
		wantPending  bool
	}{
		{"vindicated", false, false, []float64{0.99, 0.01}, HarbingerOutcomeVindicated, flavor.HarbingerVindicated, true},
		{"spared", false, false, []float64{0.0}, HarbingerOutcomeSpared, flavor.HarbingerSpared, false},
		{"discredited", true, false, []float64{0.0}, HarbingerOutcomeDiscredited, flavor.HarbingerDiscredited, false},
		{"false prophet vindicated by chance", true, false, []float64{0.99, 0.01}, HarbingerOutcomeVindicated, flavor.HarbingerVindicated, true},
		{"fulfilled", false, true, []float64{0.0}, HarbingerOutcomeFulfilled, flavor.HarbingerFulfilled, true},
		{"fulfilled false prophet", true, true, []float64{0.0}, HarbingerOutcomeFulfilled, flavor.HarbingerFulfilled, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ge := threadEngine(t, "stone_era", 9)
			ge.advanceAge("stone_age")
			ge.advanceAge("bronze_age")
			ge.harbinger.FalseProphet = tc.falseProphet
			ge.harbinger.ClaimFactor = 0
			if tc.falseProphet {
				ge.harbinger.ClaimFactor = 1.2
			}
			if tc.invite {
				if err := ge.HarbingerInvite(); err != nil {
					t.Fatal(err)
				}
			}
			ge.rng = riggedRNG(tc.rng...)
			ge.advanceAge("iron_age")
			if (ge.pendingCatastrophe != "") != tc.wantPending {
				t.Fatalf("pending = %q", ge.pendingCatastrophe)
			}
			if len(ge.harbingerHistory) != 1 {
				t.Fatalf("history = %+v", ge.harbingerHistory)
			}
			r := ge.harbingerHistory[0]
			if r.Outcome != tc.wantOutcome || r.FalseProphet != tc.falseProphet || r.TargetEpochKey != "iron_era" ||
				r.Age != "bronze_age" || len(r.Chain) != 3 {
				t.Errorf("record = %+v", r)
			}
			line := flavorAfterVerdict(t, ge)
			if !fromMoment(line, tc.wantMoment) {
				t.Errorf("flavor line %q is not from %s", line, tc.wantMoment)
			}
			for _, other := range []flavor.Moment{flavor.HarbingerVindicated, flavor.HarbingerSpared, flavor.HarbingerDiscredited, flavor.HarbingerFulfilled} {
				if other != tc.wantMoment && fromMoment(line, other) {
					t.Errorf("flavor line %q also matches %s", line, other)
				}
			}
			// The Iron Era's own thread has started.
			if ge.harbinger == nil || ge.harbinger.EpochKey != "iron_era" {
				t.Errorf("iron era thread after resolution = %+v", ge.harbinger)
			}
		})
	}
}

// --- persistence and resets ------------------------------------------------------------------

func TestHarbingerSaveLoadMidThread(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := threadEngine(t, "iron_era", 10)
	ge.advanceAge("classical_age")
	setStock(ge, map[string][2]float64{"faith": {1e6, 1e6}})
	fillBraceStock(ge, 1e7)
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	if err := ge.HarbingerBrace(); err != nil {
		t.Fatal(err)
	}
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	ge.harbinger.FalseProphet = true
	ge.harbinger.ClaimFactor = 1.1
	want := *ge.harbinger
	if err := ge.SaveGame("harb_mid"); err != nil {
		t.Fatal(err)
	}

	ge2 := NewGameEngine()
	if err := ge2.LoadGame("harb_mid"); err != nil {
		t.Fatal(err)
	}
	if ge2.cheaterBadge {
		t.Error("save with a harbinger failed signature verification")
	}
	if ge2.harbinger == nil || !reflect.DeepEqual(*ge2.harbinger, want) {
		t.Fatalf("thread after load = %+v, want %+v", ge2.harbinger, want)
	}
	if !ge2.catastropheInvited || !ge2.harbingerArrived["iron_era"] {
		t.Errorf("invite %v / arrived %v not restored", ge2.catastropheInvited, ge2.harbingerArrived)
	}
	if countLogs(ge2, "has come")+countLogs(ge2, "takes up") != 0 {
		t.Error("loading spoke again")
	}
	ge2.harbingerTickCheck()
	ge2.advanceAge("medieval_age")
	if !reflect.DeepEqual(ge2.harbinger.Chain, []string{"iron_age", "classical_age", "medieval_age"}) {
		t.Errorf("chain after load and handoff = %v", ge2.harbinger.Chain)
	}
	ge2.rng = riggedRNG(0.0)
	ge2.advanceAge("renaissance_age")
	if ge2.pendingCatastrophe != "steel_era" || ge2.pendingBraceLevel != 1 {
		t.Errorf("after transition: pending %q brace %d", ge2.pendingCatastrophe, ge2.pendingBraceLevel)
	}
}

// A save with no harbinger keys (written before the feature) sitting in a
// middle age of a qualifying epoch gets its thread on load, starting there.
func TestOldSaveInMiddleAgeGetsThreadOnLoad(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := catEngine(t, "classical_age", 1)
	if ge.harbinger != nil {
		t.Fatal("catEngine should not start a thread")
	}
	if err := ge.SaveGame("old_mid_age"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("old_mid_age"); err != nil {
		t.Fatal(err)
	}
	h := ge2.harbinger
	if h == nil || h.Age != "classical_age" || h.EpochKey != "iron_era" || !reflect.DeepEqual(h.Chain, []string{"classical_age"}) {
		t.Errorf("thread after load = %+v", h)
	}
}

// Saves with no thread carry no harbinger or Last Passage keys. A Cosmic Era
// save gets its Last Passage thread on load.
func TestHarbingerFieldsOmittedWhenAbsent(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := catEngine(t, "quantum_age", 1)
	raw, err := json.Marshal(ge.buildSaveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"harbinger", "catastrophe_invited", "pending_brace_level", "pending_last_passage", "cosmic_legacy"} {
		if strings.Contains(string(raw), `"`+k) {
			t.Errorf("save without a harbinger contains %q", k)
		}
	}
	if err := ge.SaveGame("no_harb"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("no_harb"); err != nil {
		t.Fatal(err)
	}
	ge2.harbingerTickCheck()
	if h := ge2.harbinger; h == nil || h.TargetEpoch != "" || h.Age != "quantum_age" || ge2.catastropheInvited || ge2.cheaterBadge {
		t.Errorf("cosmic save loaded as harbinger=%+v invited=%v cheater=%v",
			ge2.harbinger, ge2.catastropheInvited, ge2.cheaterBadge)
	}
}

func TestSuccumbAndPrestigeResetHarbinger(t *testing.T) {
	ge := threadEngine(t, "stone_era", 12)
	ge.advanceAge("stone_age")
	ge.advanceAge("bronze_age")
	ge.rng = badThenEscalate()
	ge.advanceAge("iron_age")
	if ge.pendingCatastrophe != "iron_era" || ge.harbinger == nil {
		t.Fatalf("setup: pending %q thread %v", ge.pendingCatastrophe, ge.harbinger)
	}
	ge.pendingBraceLevel = 2
	if err := ge.HarbingerInvite(); err != nil { // the Iron Era thread
		t.Fatal(err)
	}
	if err := ge.Succumb(); err != nil {
		t.Fatal(err)
	}
	if ge.harbinger != nil || ge.catastropheInvited || ge.pendingBraceLevel != 0 || len(ge.harbingerArrived) != 0 {
		t.Errorf("after Succumb: thread %v invited %v brace %d arrived %v",
			ge.harbinger, ge.catastropheInvited, ge.pendingBraceLevel, ge.harbingerArrived)
	}
	if len(ge.harbingerHistory) != 1 {
		t.Errorf("history should survive Succumb like the epoch history: %+v", ge.harbingerHistory)
	}
	ge.harbingerTickCheck()
	if ge.harbinger == nil || ge.harbinger.Age != "primitive_age" {
		t.Errorf("new run's first tick: thread = %+v", ge.harbinger)
	}

	ge = threadEngine(t, "digital_era", 13)
	walkTo(t, ge, "digital_age")
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	ge.harbingerHistory = []HarbingerRecord{{Outcome: HarbingerOutcomeSpared}}
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if ge.harbinger != nil || ge.catastropheInvited || len(ge.harbingerArrived) != 0 || ge.harbingerHistory != nil {
		t.Errorf("after prestige: thread %v invited %v arrived %v history %v",
			ge.harbinger, ge.catastropheInvited, ge.harbingerArrived, ge.harbingerHistory)
	}
	ge.harbingerTickCheck()
	if ge.harbinger == nil || ge.harbinger.Age != "primitive_age" {
		t.Errorf("after prestige, first tick: thread = %+v", ge.harbinger)
	}
}

// The Appease price is sized so a moderate faith (and culture) economy can
// pay it: modelled at config.FlowIncome through the thread's ages at their
// targets, level 1 must come before the thread's last age ends and level 2
// (on top of it) by the passage, in every epoch.
func TestAppeasePayableWithinThread(t *testing.T) {
	for _, ep := range config.Epochs() {
		cost1, cost2 := harbingerAppeaseCost(ep.Key, 1), harbingerAppeaseCost(ep.Key, 2)
		if len(cost1) == 0 {
			continue
		}
		for res, l1 := range cost1 {
			if math.Abs(cost2[res]-2*l1) > 1e-6 {
				t.Errorf("%s: level 2 %s = %v, want double %v", ep.Key, res, cost2[res], l1)
			}
			made, total, l1At := 0.0, 0.0, -1.0
			ages := harbingerAppeaseAges(ep.Key)
			for _, a := range ages {
				total += config.AgeTargetTicks(a)
			}
			elapsed := 0.0
			for _, a := range ages {
				rate, span := config.FlowIncome(res, a), config.AgeTargetTicks(a)
				if l1At < 0 && rate > 0 && made+rate*span >= l1 {
					l1At = elapsed + (l1-made)/rate
				}
				made += rate * span
				elapsed += span
			}
			if l1At < 0 || l1At >= total {
				t.Errorf("%s: level 1 %s (%v) is never made at a moderate income", ep.Key, res, l1)
			}
			if made < l1+cost2[res] {
				t.Errorf("%s: levels 1 and 2 of %s (%v) exceed the thread's moderate income %v", ep.Key, res, l1+cost2[res], made)
			}
		}
	}
}
