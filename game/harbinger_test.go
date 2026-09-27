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

// nextAge returns the age after age in the canonical order.
func nextAge(t *testing.T, age string) string {
	t.Helper()
	order := config.AgeOrder()
	for i, a := range order {
		if a == age && i+1 < len(order) {
			return order[i+1]
		}
	}
	t.Fatalf("no age after %s", age)
	return ""
}

// harbArrive returns a seeded engine that has just advanced into lastAge (the
// last age of an epoch) and been met by its harbinger.
func harbArrive(t *testing.T, lastAge string, seed int64) *GameEngine {
	t.Helper()
	ge := catEngine(t, prevAge(t, lastAge), seed)
	ge.advanceAge(lastAge)
	if ge.harbinger == nil {
		t.Fatalf("no harbinger on entering %s", lastAge)
	}
	return ge
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

// fillBraceStock gives every Brace resource a full bank of size store.
func fillBraceStock(ge *GameEngine, store float64) {
	stock := map[string][2]float64{}
	for _, k := range ge.harbingerBraceResources() {
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

// harbingerLines returns every log line from the first "⚑" line on that is a
// harbinger line or its flavor: the part of the log that must be seeded.
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

// flavorAfterVerdict returns the gray flavor line logged after the last
// verdict line ("⚑ ..."), stripped of its markup.
func flavorAfterVerdict(t *testing.T, ge *GameEngine) string {
	t.Helper()
	msgs := logMessages(ge)
	last := -1
	for i, m := range msgs {
		if strings.HasPrefix(m, "⚑") {
			last = i
		}
	}
	if last < 0 {
		t.Fatal("no harbinger verdict in the log")
	}
	for _, m := range msgs[last+1:] {
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

// --- arrival --------------------------------------------------------------------

func TestHarbingerArrivesOnEnteringLastAgeOfEpoch(t *testing.T) {
	ge := catEngine(t, "primitive_age", 1)
	var toasts []EventData
	ge.Bus.Subscribe(EventHarbingerArrived, func(e EventData) { toasts = append(toasts, e) })

	ge.advanceAge("stone_age") // middle of the Stone Era
	if ge.harbinger != nil {
		t.Fatal("harbinger arrived in stone_age, which is not the last age of its epoch")
	}
	ge.advanceAge("bronze_age")
	h := ge.harbinger
	if h == nil {
		t.Fatal("no harbinger on entering bronze_age")
	}
	if h.Age != "bronze_age" || h.EpochKey != "stone_era" || h.TargetEpoch != "iron_era" {
		t.Errorf("harbinger = %+v", h)
	}
	if len(h.Lines) != 2 {
		t.Errorf("want an arrival and a warning line, got %q", h.Lines)
	}
	if len(toasts) != 1 || toasts[0].Payload["epoch_key"] != "iron_era" {
		t.Errorf("arrival bus events = %+v", toasts)
	}
	def, _ := config.HarbingerFor("bronze_age")
	joined := strings.Join(logMessages(ge), "\n")
	if !strings.Contains(joined, capFirst(def.Name)+" has come") {
		t.Errorf("no arrival log line for %s", def.Name)
	}
	st := ge.GetState()
	if st.Harbinger == nil || st.Harbinger.Name != def.Name || st.Harbinger.TargetEpochName != "Iron Era" {
		t.Errorf("GetState harbinger = %+v", st.Harbinger)
	}
	if ge.pendingCatastrophe != "" {
		t.Error("arrival must not block anything")
	}
}

// Every epoch's last age brings its harbinger exactly when the transition out
// of it can roll a catastrophe: all of them except the final epoch's.
func TestHarbingerArrivalFollowsCatastropheGate(t *testing.T) {
	for _, ep := range config.Epochs() {
		last := ep.Ages[len(ep.Ages)-1]
		ge := catEngine(t, prevAge(t, last), 3)
		ge.advanceAge(last)
		next, hasNext := config.NextEpoch(ep.Key)
		want := hasNext && config.CatastropheAllowed(next.Key)
		if got := ge.harbinger != nil; got != want {
			t.Errorf("%s (%s): harbinger present = %v, want %v", ep.Key, last, got, want)
		}
	}
	// Never in the final epoch, even when summoned.
	ge := catEngine(t, "transcendent_age", 1)
	if err := ge.summonHarbinger(); err == nil {
		t.Error("summoning in the final epoch must be refused")
	}
}

func TestNoHarbingerWhenTransitionCannotRoll(t *testing.T) {
	ge := catEngine(t, "stone_age", 1)
	ge.epochEventFired["iron_era"] = true // the transition roll already happened
	ge.advanceAge("bronze_age")
	if ge.harbinger != nil {
		t.Error("harbinger arrived although the next transition cannot roll")
	}
}

func TestHarbingerOncePerEpochPerRun(t *testing.T) {
	ge := harbArrive(t, "bronze_age", 2)
	ge.harbinger = nil
	ge.maybeHarbingerArrive()
	if ge.harbinger != nil {
		t.Error("a second harbinger arrived in the same epoch")
	}
	// Re-entering the age (e.g. dev jump) does not bring another either.
	ge.advanceAge("bronze_age")
	if ge.harbinger != nil {
		t.Error("re-entering the last age brought a second harbinger")
	}
}

// --- false prophets ------------------------------------------------------------------

func TestFalseProphetRateOnlyBeforeIndustrial(t *testing.T) {
	rate := func(age string, n int) (float64, map[CatastropheTier]int) {
		ge := catEngine(t, age, 11)
		tiers := map[CatastropheTier]int{}
		falseCount := 0
		for i := 0; i < n; i++ {
			ge.harbinger = nil
			if !ge.harbingerArrive() {
				t.Fatalf("%s: harbinger did not arrive", age)
			}
			h := ge.harbinger
			if h.FalseProphet {
				falseCount++
				tiers[h.AnnouncedTier]++
				if h.AnnouncedTier != CatastropheTierMedium && h.AnnouncedTier != CatastropheTierHigh {
					t.Fatalf("%s: false prophet announced %q", age, h.AnnouncedTier)
				}
			} else if h.AnnouncedTier != h.ArrivalRealTier {
				t.Fatalf("%s: true harbinger announced %q, real tier %q", age, h.AnnouncedTier, h.ArrivalRealTier)
			}
		}
		return float64(falseCount) / float64(n), tiers
	}

	for _, age := range []string{"bronze_age", "medieval_age"} {
		def, _ := config.HarbingerFor(age)
		got, tiers := rate(age, 4000)
		if math.Abs(got-def.FalseProphetChance) > 0.025 {
			t.Errorf("%s: false-prophet rate %.4f, want ≈ %.4f", age, got, def.FalseProphetChance)
		}
		if tiers[CatastropheTierMedium] == 0 || tiers[CatastropheTierHigh] == 0 {
			t.Errorf("%s: false prophets should announce both medium and high, got %v", age, tiers)
		}
	}
	for _, age := range []string{"industrial_age", "atomic_age", "digital_age", "space_age"} {
		if got, _ := rate(age, 300); got != 0 {
			t.Errorf("%s: false-prophet rate %.4f, want 0 from the Industrial Age on", age, got)
		}
	}
}

// A false prophet's displayed severity moves with the real tier, so Appease
// cannot tell a false one from a true one.
func TestFalseProphetDisplayTierTracksRealTier(t *testing.T) {
	ge := catEngine(t, "medieval_age", 1)
	ge.harbinger = &HarbingerSave{Age: "medieval_age", FalseProphet: true,
		AnnouncedTier: CatastropheTierHigh, ArrivalRealTier: CatastropheTierMedium}
	if got := ge.harbingerDisplayTier(CatastropheTierMedium); got != CatastropheTierHigh {
		t.Errorf("unchanged real tier: display %q, want high", got)
	}
	if got := ge.harbingerDisplayTier(CatastropheTierLow); got != CatastropheTierMedium {
		t.Errorf("real tier fell one step: display %q, want medium", got)
	}
	ge.harbinger.AnnouncedTier = CatastropheTierMedium
	ge.harbinger.ArrivalRealTier = CatastropheTierHigh
	if got := ge.harbingerDisplayTier(CatastropheTierLow); got != CatastropheTierLow {
		t.Errorf("never below low: display %q", got)
	}
	ge.harbinger.FalseProphet = false
	if got := ge.harbingerDisplayTier(CatastropheTierLow); got != CatastropheTierLow {
		t.Errorf("true harbinger shows the real tier: display %q", got)
	}
}

// --- determinism ----------------------------------------------------------------------

func TestHarbingerSameSeedSameOutcome(t *testing.T) {
	type run struct {
		h       HarbingerSave
		history []HarbingerRecord
		lines   []string
	}
	play := func(seed int64) run {
		ge := harbArrive(t, "medieval_age", seed)
		setStock(ge, map[string][2]float64{"faith": {1000, 1000}, "culture": {1000, 1000}})
		if err := ge.HarbingerAppease(); err != nil {
			t.Fatal(err)
		}
		h := *ge.harbinger
		ge.advanceAge("renaissance_age")
		return run{h, ge.harbingerHistory, harbingerLines(ge)}
	}
	a, b := play(42), play(42)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("same seed, different harbinger runs:\n%+v\n%+v", a, b)
	}
	differs := false
	for s := int64(43); s < 53 && !differs; s++ {
		differs = !reflect.DeepEqual(a.lines, play(s).lines)
	}
	if !differs {
		t.Error("ten other seeds produced the same harbinger log as seed 42")
	}
}

// --- Appease ------------------------------------------------------------------------

func TestAppeaseCostsLevelsAndOdds(t *testing.T) {
	ge := harbArrive(t, "medieval_age", 4)
	setStock(ge, map[string][2]float64{"faith": {1000, 1000}, "culture": {2000, 2000}})

	// Level 1: 15% of each cap. Faith fill 1.0 → high band, base 12%.
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	if f, c := ge.Resources.Get("faith"), ge.Resources.Get("culture"); f != 850 || c != 1700 {
		t.Errorf("after level 1: faith %v culture %v, want 850 / 1700", f, c)
	}
	o := ge.CatastropheOutlook()
	if math.Abs(o.Probability-0.12*0.6) > 1e-9 {
		t.Errorf("level 1 probability %v, want %v", o.Probability, 0.12*0.6)
	}
	if !strings.Contains(strings.Join(logMessages(ge), "\n"), "Appease 1/2") {
		t.Error("no Appease log line")
	}

	// Level 2 costs double: 30%. Faith fill drops to 0.55 → mid band, base 15%.
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	if f, c := ge.Resources.Get("faith"), ge.Resources.Get("culture"); f != 550 || c != 1100 {
		t.Errorf("after level 2: faith %v culture %v, want 550 / 1100", f, c)
	}
	if o := ge.CatastropheOutlook(); math.Abs(o.Probability-0.15*0.36) > 1e-9 {
		t.Errorf("level 2 probability %v, want %v", o.Probability, 0.15*0.36)
	}
	if st := ge.GetState(); st.Harbinger.AppeaseLevel != 2 || st.Harbinger.AppeaseBlocked == "" || st.Harbinger.AppeaseCost != nil {
		t.Errorf("view after cap = %+v", st.Harbinger)
	}

	// Level cap.
	if err := ge.HarbingerAppease(); err == nil || !strings.Contains(err.Error(), "as far as it goes") {
		t.Errorf("third appease: err = %v", err)
	}
	if f := ge.Resources.Get("faith"); f != 550 {
		t.Errorf("refused appease deducted faith: %v", f)
	}
}

func TestAppeaseRefusesWhenUnaffordable(t *testing.T) {
	ge := harbArrive(t, "medieval_age", 4)
	setStock(ge, map[string][2]float64{"faith": {100, 1000}, "culture": {2000, 2000}})
	err := ge.HarbingerAppease()
	if err == nil || !strings.Contains(err.Error(), "50 more faith") {
		t.Fatalf("err = %v, want a shortfall of 50 faith", err)
	}
	if ge.Resources.Get("faith") != 100 || ge.Resources.Get("culture") != 2000 || ge.harbinger.AppeaseLevel != 0 {
		t.Error("a refused appease changed state")
	}
	if st := ge.GetState(); st.Harbinger.AppeaseAffordable {
		t.Error("view says affordable")
	}
}

// Before culture unlocks (the Bronze Age harbinger), Appease costs faith only.
func TestAppeaseBeforeCultureCostsFaithOnly(t *testing.T) {
	ge := harbArrive(t, "bronze_age", 4)
	cost := ge.harbingerAppeaseCost(1)
	if _, ok := cost["culture"]; ok || cost["faith"] <= 0 {
		t.Errorf("bronze age appease cost = %v, want faith only", cost)
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

// Appease lowers the escalation threshold of the real roll, not just the
// displayed odds.
func TestAppeaseChangesTheRealRoll(t *testing.T) {
	transition := func(appease bool) string {
		ge := harbArrive(t, "bronze_age", 5)
		setStock(ge, map[string][2]float64{"faith": {50, 100}})
		if appease {
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
			ge := harbArrive(t, "bronze_age", 6)
			fillBraceStock(ge, 10000)
			for i := 1; i <= tc.level; i++ {
				before := ge.harbingerBraceCost(i)
				stockBefore := ge.Resources.Get("wood")
				if err := ge.HarbingerBrace(); err != nil {
					t.Fatal(err)
				}
				if want := math.Ceil(10000 * harbingerBraceCostFrac * float64(i)); before["wood"] != want || ge.Resources.Get("wood") != stockBefore-want {
					t.Errorf("level %d wood cost %v (want %v), stock %v → %v", i, before["wood"], want, stockBefore, ge.Resources.Get("wood"))
				}
			}
			if tc.level == HarbingerMaxBrace {
				if err := ge.HarbingerBrace(); err == nil {
					t.Error("brace beyond the cap must be refused")
				}
			}
			ge.rng = badThenEscalate()
			ge.advanceAge("iron_age")
			if ge.pendingCatastrophe != "iron_era" {
				t.Fatalf("no catastrophe: pending = %q", ge.pendingCatastrophe)
			}
			if ge.pendingBraceLevel != tc.level {
				t.Fatalf("pendingBraceLevel = %d, want %d", ge.pendingBraceLevel, tc.level)
			}

			// The Brace survives a save/load before the choice.
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
	ge := harbArrive(t, "medieval_age", 7)
	setStock(ge, map[string][2]float64{"faith": {1000, 1000}, "culture": {1000, 1000}})
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
	if ge.Resources.Get("faith") != 1000 {
		t.Error("refused appease spent faith")
	}
	if err := ge.HarbingerInvite(); err == nil {
		t.Error("inviting twice must be refused")
	}
	v := ge.GetState().Harbinger
	if v.AppeaseBlocked == "" || v.InviteBlocked == "" || v.BraceBlocked != "" {
		t.Errorf("view after invite = %+v", v)
	}
	fillBraceStock(ge, 10000)
	if err := ge.HarbingerBrace(); err != nil {
		t.Errorf("brace after invite: %v", err)
	}

	ge.rng = riggedRNG(0.0) // a good roll: the invite must override it
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
			ge := harbArrive(t, "bronze_age", 9)
			ge.harbinger.FalseProphet = tc.falseProphet
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
			if ge.harbinger != nil {
				t.Error("harbinger not cleared at the transition")
			}
			if len(ge.harbingerHistory) != 1 {
				t.Fatalf("history = %+v", ge.harbingerHistory)
			}
			r := ge.harbingerHistory[0]
			if r.Outcome != tc.wantOutcome || r.FalseProphet != tc.falseProphet || r.TargetEpochKey != "iron_era" || r.Age != "bronze_age" {
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
			if st := ge.GetState(); st.Harbinger != nil || len(st.HarbingerHistory) != 1 {
				t.Errorf("state after resolution: harbinger %+v, history %d", st.Harbinger, len(st.HarbingerHistory))
			}
		})
	}
}

// --- persistence and resets ------------------------------------------------------------------

func TestHarbingerSaveLoadMidHarbinger(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := harbArrive(t, "medieval_age", 10)
	setStock(ge, map[string][2]float64{"faith": {1000, 1000}, "culture": {1000, 1000}})
	fillBraceStock(ge, 10000)
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	if err := ge.HarbingerBrace(); err != nil {
		t.Fatal(err)
	}
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	ge.harbinger.FalseProphet = true // persisted even though it cannot be seen
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
		t.Fatalf("harbinger after load = %+v, want %+v", ge2.harbinger, want)
	}
	if !ge2.catastropheInvited || !ge2.harbingerArrived["iron_era"] {
		t.Errorf("invite %v / arrived %v not restored", ge2.catastropheInvited, ge2.harbingerArrived)
	}
	if n := strings.Count(strings.Join(logMessages(ge2), "\n"), "has come"); n != 0 {
		t.Errorf("loading brought %d new harbinger arrivals", n)
	}
	// And it still resolves on the transition.
	ge2.rng = riggedRNG(0.0)
	ge2.advanceAge("renaissance_age")
	if ge2.pendingCatastrophe != "steel_era" || ge2.pendingBraceLevel != 1 {
		t.Errorf("after load and transition: pending %q brace %d", ge2.pendingCatastrophe, ge2.pendingBraceLevel)
	}
}

// Saves from before the feature, or with no harbinger, carry no harbinger keys
// and load with none.
func TestHarbingerFieldsOmittedWhenAbsent(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := catEngine(t, "iron_age", 1)
	raw, err := json.Marshal(ge.buildSaveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"harbinger", "catastrophe_invited", "pending_brace_level"} {
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
	if ge2.harbinger != nil || ge2.catastropheInvited || ge2.harbingerArrived == nil || ge2.cheaterBadge {
		t.Errorf("old-style save loaded as harbinger=%v invited=%v arrived=%v cheater=%v",
			ge2.harbinger, ge2.catastropheInvited, ge2.harbingerArrived, ge2.cheaterBadge)
	}
}

// A save sitting in the last age of an epoch with no harbinger (written before
// the feature) gets its harbinger on load.
func TestOldSaveInLastAgeGetsHarbingerOnLoad(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := catEngine(t, "medieval_age", 1)
	if ge.harbinger != nil {
		t.Fatal("catEngine should not bring a harbinger")
	}
	if err := ge.SaveGame("old_last_age"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("old_last_age"); err != nil {
		t.Fatal(err)
	}
	if ge2.harbinger == nil || ge2.harbinger.Age != "medieval_age" {
		t.Errorf("harbinger after load = %+v", ge2.harbinger)
	}
}

func TestSuccumbAndPrestigeResetHarbinger(t *testing.T) {
	// Succumb: a catastrophe pending with a Brace, another harbinger armed.
	ge := harbArrive(t, "bronze_age", 12)
	ge.rng = badThenEscalate()
	ge.advanceAge("iron_age")
	if ge.pendingCatastrophe != "iron_era" {
		t.Fatalf("setup: pending = %q", ge.pendingCatastrophe)
	}
	ge.pendingBraceLevel = 2
	ge.harbinger = &HarbingerSave{Age: "medieval_age", EpochKey: "iron_era", TargetEpoch: "steel_era", Invited: true}
	ge.catastropheInvited = true
	if err := ge.Succumb(); err != nil {
		t.Fatal(err)
	}
	if ge.harbinger != nil || ge.catastropheInvited || ge.pendingBraceLevel != 0 || len(ge.harbingerArrived) != 0 {
		t.Errorf("after Succumb: harbinger %v invited %v brace %d arrived %v",
			ge.harbinger, ge.catastropheInvited, ge.pendingBraceLevel, ge.harbingerArrived)
	}
	if len(ge.harbingerHistory) != 1 {
		t.Errorf("history should survive Succumb like the epoch history: %+v", ge.harbingerHistory)
	}

	// Prestige, from the Digital Age with its harbinger present.
	ge = harbArrive(t, "digital_age", 13)
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	ge.harbingerHistory = []HarbingerRecord{{Outcome: HarbingerOutcomeSpared}}
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if ge.harbinger != nil || ge.catastropheInvited || len(ge.harbingerArrived) != 0 || ge.harbingerHistory != nil {
		t.Errorf("after prestige: harbinger %v invited %v arrived %v history %v",
			ge.harbinger, ge.catastropheInvited, ge.harbingerArrived, ge.harbingerHistory)
	}
	// The new run can meet the Bronze Age harbinger again.
	ge.advanceAge("stone_age")
	ge.advanceAge("bronze_age")
	if ge.harbinger == nil {
		t.Error("no harbinger in the next run")
	}
}

// --- costs scale with the economy ----------------------------------------------------------------

func TestHarbingerCostsScaleWithStorage(t *testing.T) {
	ge := harbArrive(t, "industrial_age", 1)
	fillBraceStock(ge, 1000)
	small := ge.harbingerBraceCost(1)
	fillBraceStock(ge, 100000)
	big := ge.harbingerBraceCost(1)
	if len(small) == 0 {
		t.Fatal("no brace resources at industrial_age")
	}
	for k, v := range small {
		if big[k] != v*100 {
			t.Errorf("%s: cost %v at cap 1000, %v at cap 100000; want ×100", k, v, big[k])
		}
	}
	if two := ge.harbingerBraceCost(2); two[ge.harbingerBraceResources()[0]] != 2*big[ge.harbingerBraceResources()[0]] {
		t.Error("level 2 must cost double")
	}
	for _, k := range ge.harbingerBraceResources() {
		if k == "faith" || k == "culture" {
			t.Errorf("brace draws on %s, which belongs to appease", k)
		}
	}
}
