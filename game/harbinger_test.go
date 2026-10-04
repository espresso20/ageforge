package game

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
	"github.com/espresso20/ageforge/pkg/textfmt"
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
// the epoch's thread just started: a doom fated for the era's last tick (so
// the thread lives through every age) whose harbinger has come; a false
// prophet in the Stone Era, where nothing can be fated; the Last Passage
// thread in the final epoch.
func threadEngine(t *testing.T, epochKey string, seed int64) *GameEngine {
	t.Helper()
	first := epochAges(t, epochKey)[0]
	ge := fateEngine(t, first, seed)
	late := int(baseEraTicks(epochKey)) - 1
	switch {
	case config.IsFinalEpoch(epochKey):
		// The Last Passage thread started on the first tick; keep the era's
		// own doom out of the way.
		if err := ge.ForceQuietFateForTest(epochKey); err != nil {
			t.Fatal(err)
		}
	case config.FateAllowed(epochKey):
		forceFate(t, ge, late)
		ge.fateArrive()
	default:
		if err := ge.ForceFalseProphetForTest(epochKey, late); err != nil {
			t.Fatal(err)
		}
		ge.fateArrive()
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
// (the last "⚑" line that is not an arrival or a handoff), without markup.
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

// --- threads ------------------------------------------------------------------

// A harbinger never names an era to come, nor any age past its own epoch: not
// in the arrival and handoff lines, the flavor lines, the answers or the view
// the panel draws. The player has not reached them (playtest 2026-09-29:
// "warning of the passage into the Iron Era"). Its own era it may name.
func TestHarbingerNeverNamesTheEraToCome(t *testing.T) {
	for _, ep := range config.Epochs() {
		t.Run(ep.Key, func(t *testing.T) {
			ge := threadEngine(t, ep.Key, 3)
			var forbidden []string
			for _, other := range config.Epochs() {
				if other.Order > ep.Order {
					forbidden = append(forbidden, other.Name)
				}
			}
			last := ep.Ages[len(ep.Ages)-1]
			past := false
			for _, a := range config.AgeOrder() {
				if past {
					forbidden = append(forbidden, config.AgeByKey()[a].Name)
				}
				past = past || a == last
			}
			var seen []string
			look := func() {
				if v := ge.GetState().Harbinger; v != nil {
					seen = append(seen, v.TargetEpochName, v.AgeName, v.WhenText, v.AppeaseBlocked, v.BraceBlocked, v.InviteBlocked)
					seen = append(seen, v.Lines...)
				}
			}
			look()
			for _, a := range ep.Ages[1:] {
				ge.advanceAge(a)
				look()
			}
			stock := map[string][2]float64{}
			for k := range eraAppeaseCost(ep.Key, HarbingerMaxAppease) {
				stock[k] = [2]float64{1e12, 1e12}
			}
			for k := range harbingerBraceBasis(ep.Key) {
				stock[k] = [2]float64{1e12, 1e12}
			}
			setStock(ge, stock)
			for _, act := range []func() error{ge.HarbingerAppease, ge.HarbingerBrace, ge.HarbingerInvite} {
				if err := act(); err != nil {
					seen = append(seen, err.Error())
				}
			}
			look()
			seen = append(seen, harbingerLines(ge)...)
			if len(harbingerLines(ge)) == 0 {
				t.Fatal("no harbinger lines were logged")
			}
			for _, line := range seen {
				for _, name := range forbidden {
					if strings.Contains(line, name) {
						t.Errorf("%q names %s", line, name)
					}
				}
			}
		})
	}
}

// Each age advance within the epoch hands the warning to that age's figure:
// new lines and a toast, same thread, same levels.
func TestHarbingerSpeakerChangesEachAge(t *testing.T) {
	for _, tc := range []struct {
		epoch string
	}{{"stone_era"}, {"iron_era"}, {"neon_era"}} {
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

// A quiet era sees no harbinger in any of its ages: harbingers come only when
// a doom is on its way.
func TestNoHarbingerInAQuietEra(t *testing.T) {
	for _, ep := range config.Epochs() {
		if !config.FateAllowed(ep.Key) {
			continue
		}
		ge := fateEngine(t, ep.Ages[0], 3)
		if err := ge.ForceQuietFateForTest(ep.Key); err != nil {
			t.Fatal(err)
		}
		tick := 0
		for _, age := range ep.Ages {
			walkTo(t, ge, age)
			for step := 0; step < 10; step++ {
				tick += int(baseAgeTicks(age)) / 10
				tickTo(ge, tick)
			}
			// (The Cosmic Era's Last Passage thread is no doom's harbinger.)
			if h := ge.harbinger; h != nil && h.TargetEpoch != "" {
				t.Errorf("%s (%s): a harbinger came in a quiet era: %+v", ep.Key, age, h)
			}
		}
	}
}

// One thread per era per run: once the era's doom has resolved, no other
// harbinger comes in it.
func TestHarbingerOncePerEpochPerRun(t *testing.T) {
	ge := threadEngine(t, "iron_era", 2)
	ge.rng = riggedRNG(0.999)
	ge.fate.StrikeTick = ge.tick
	tickTo(ge, ge.tick)
	if ge.harbinger != nil || ge.fate.Resolved != FateSpared {
		t.Fatalf("setup: thread %+v fate %+v", ge.harbinger, ge.fate)
	}
	for tick := 0; tick < int(baseEraTicks("iron_era")); tick += 500 {
		tickTo(ge, tick)
	}
	ge.advanceAge("classical_age")
	if ge.harbinger != nil {
		t.Errorf("a second thread started in the same era: %+v", ge.harbinger)
	}
}

// --- false prophets ------------------------------------------------------------------

// Handoffs never re-roll: a false thread stays false, a true one true.
func TestHandoffKeepsTheFalseProphetFlag(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		for _, falseProphet := range []bool{false, true} {
			ge := fateEngine(t, "iron_age", seed)
			late := int(baseEraTicks("iron_era")) - 1
			var err error
			if falseProphet {
				err = ge.ForceFalseProphetForTest("iron_era", late)
			} else {
				err = ge.ForceFateForTest("iron_era", late)
			}
			if err != nil {
				t.Fatal(err)
			}
			ge.fateArrive()
			ge.advanceAge("classical_age")
			ge.advanceAge("medieval_age")
			if ge.harbinger.FalseProphet != falseProphet {
				t.Fatalf("seed %d: the handoff changed the false-prophet flag", seed)
			}
		}
	}
}

// A false thread's claim is a fixed multiple of what a real doom's strike
// chance would be, repeated by every figure, moved by Appease like a real
// one, and printed as the claimed figure once a numeric figure takes over.
func TestFalseThreadClaim(t *testing.T) {
	ge := fateEngine(t, "renaissance_age", 1)
	setFaith(ge, 5e6, 1e7) // mid faith: a real doom would strike 75% of the time
	if err := ge.ForceFalseProphetForTest("steel_era", int(baseEraTicks("steel_era"))-1); err != nil {
		t.Fatal(err)
	}
	ge.fateArrive()
	if tier, p := ge.harbingerDisplay(); tier != CatastropheTierHigh || math.Abs(p-0.90) > 1e-9 {
		t.Errorf("claim = %s %.3f, want high 0.90", tier, p)
	}
	ge.advanceAge("colonial_age")
	if ge.harbinger.AnnouncedTier != CatastropheTierHigh {
		t.Errorf("colonial figure announced %q, want the same false high", ge.harbinger.AnnouncedTier)
	}
	ge.advanceAge("industrial_age")
	setStock(ge, map[string][2]float64{"faith": {5e6, 1e7}, "culture": {1e7, 1e7}})
	v := ge.GetState().Harbinger
	if !v.Numeric || v.Tier != CatastropheTierHigh || math.Abs(v.Probability-0.90) > 1e-9 {
		t.Errorf("industrial view of a false thread = numeric %v %s %.3f, want the claimed high 90%%", v.Numeric, v.Tier, v.Probability)
	}
	// Appease ×0.6 moves the claim as it would move the real odds (read
	// live, so a band change from the faith spend counts too).
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	wantP := ge.strikeChance() * ge.harbinger.ClaimFactor
	if _, p := ge.harbingerDisplay(); math.Abs(p-wantP) > 1e-9 || p >= 0.90 {
		t.Errorf("claim after appease = %.4f, want %.4f (< 0.90)", p, wantP)
	}
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	if tier, p := ge.harbingerDisplay(); tier != CatastropheTierHigh || p != 1 {
		t.Errorf("claim after invite = %s %.3f, want high 1", tier, p)
	}
}

// --- determinism ----------------------------------------------------------------------

// The same seed plays the same era: the same fate, the same harbinger lines,
// the same verdict.
func TestHarbingerSameSeedSameOutcome(t *testing.T) {
	type run struct {
		fate    FateSave
		history []HarbingerRecord
		lines   []string
	}
	play := func(seed int64) run {
		ge := catEngine(t, "bronze_age", seed)
		for _, a := range config.AgeOrder() {
			ge.applyAgeUnlocks(a)
			if a == "bronze_age" {
				break
			}
		}
		ge.advanceAge("iron_age") // rolls the Iron Era's fate
		forceFate(t, ge, 9000)    // same doom every seed; the lines still differ
		for tick := 0; tick <= 9000; tick += 100 {
			if tick == 8900 {
				ge.advanceAge("classical_age")
			}
			tickTo(ge, tick)
		}
		return run{*ge.fate, ge.harbingerHistory, harbingerLines(ge)}
	}
	a, b := play(42), play(42)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("same seed, different eras:\n%+v\n%+v", a, b)
	}
	if len(a.history) != 1 || a.fate.Resolved == "" {
		t.Errorf("the doom did not resolve: fate %+v history %+v", a.fate, a.history)
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

// The price belongs to the doom: the same in every age of the epoch, and
// whatever the player's storage.
func TestHarbingerCostSameAcrossEpoch(t *testing.T) {
	for _, ep := range config.Epochs() {
		if !config.FateAllowed(ep.Key) && !config.IsFinalEpoch(ep.Key) {
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
		epoch string
		// appease is the Last Passage's price (nil where the era has no
		// Last Passage): a quarter of what FlowIncome makes over the era's
		// ages at their targets.
		appease map[string]float64
		brace   map[string]float64
	}{
		{"stone_era", nil, map[string]float64{"food": 9600, "wood": 4800, "knowledge": 2400}},
		{"steel_era", nil, map[string]float64{"knowledge": 3600000, "gold": 1800000, "steel": 288000}},
		// The Cosmic Era's passage is prestige; its Last Passage thread's
		// Appease counts the era's ages but the last (Interstellar, Galactic,
		// Quantum). Brace is priced off the era's own advances, for
		// resources held from Interstellar (dark matter 13T, titanium 630B
		// → 12%). Antimatter and quantum flux arrive later.
		{"cosmic_era", map[string]float64{"faith": 3100000000, "culture": 48000000000}, map[string]float64{"dark_matter": 1560000000000, "titanium": 75600000000}},
	}
	for _, c := range cases {
		if c.appease != nil {
			if got := eraAppeaseCost(c.epoch, 1); !reflect.DeepEqual(got, c.appease) {
				t.Errorf("%s Last Passage appease = %v, want %v", c.epoch, got, c.appease)
			}
		}
		if got := harbingerBraceCost(c.epoch, 1); !reflect.DeepEqual(got, c.brace) {
			t.Errorf("%s brace = %v, want %v", c.epoch, got, c.brace)
		}
	}

	// A doom's Appease is priced on its warning: three quarters of what the
	// age its harbinger arrives in makes in the shortest warning, 15% of the
	// age. Iron Age faith: 0.75 a tick over 11,700 ticks = 8,775; 15% is
	// 1,316 → 1,400. The Iron Era asks no culture (it arrives in the
	// Classical Age, after the era began).
	dooms := []struct {
		epoch, age string
		appease    map[string]float64
	}{
		{"iron_era", "iron_age", map[string]float64{"faith": 1400}},
		{"iron_era", "medieval_age", map[string]float64{"faith": 4900}},
		{"steel_era", "renaissance_age", map[string]float64{"faith": 9200, "culture": 95000}},
		{"steel_era", "industrial_age", map[string]float64{"faith": 79000, "culture": 870000}},
		{"cosmic_era", "interstellar_age", map[string]float64{"faith": 270000000, "culture": 4100000000}},
	}
	for _, c := range dooms {
		if got := doomAppeaseCost(c.epoch, c.age, 1); !reflect.DeepEqual(got, c.appease) {
			t.Errorf("a doom of the %s foretold in the %s: appease = %v, want %v", c.epoch, c.age, got, c.appease)
		}
	}
}

// A thread's Appease price is its own: a doom's is priced on the age its
// harbinger arrived in and stays the same in every age the thread lives
// through; the Last Passage's is priced on the era.
func TestThreadAppeaseCost(t *testing.T) {
	ge := threadEngine(t, "electric_era", 6)
	ages := epochAges(t, "electric_era")
	want := doomAppeaseCost("electric_era", ages[0], 1)
	if got := threadAppeaseCost(ge.harbinger, 1); !reflect.DeepEqual(got, want) || len(want) != 2 {
		t.Fatalf("thread started in %s: appease %v, want %v", ages[0], got, want)
	}
	for _, a := range ages[1:] {
		ge.advanceAge(a)
		if ge.harbinger == nil || ge.harbinger.Age != a || ge.harbinger.startAge() != ages[0] {
			t.Fatalf("no handoff to %s (thread %+v)", a, ge.harbinger)
		}
		if got := ge.GetState().Harbinger.AppeaseCost; !reflect.DeepEqual(got, want) {
			t.Errorf("in %s the price is %v; it was set at %v when the harbinger arrived", a, got, want)
		}
	}
	// A harbinger that arrives later in the era is priced on that age: more,
	// since the age makes more.
	later := doomAppeaseCost("electric_era", ages[len(ages)-1], 1)
	for res, v := range want {
		if later[res] <= v {
			t.Errorf("%s: a doom foretold in %s costs %v, in %s %v; the later age makes more", res, ages[len(ages)-1], later[res], ages[0], v)
		}
	}
	// A save from before chains were kept prices on the current speaker.
	old := &HarbingerSave{Age: ages[1], EpochKey: "electric_era", TargetEpoch: "electric_era"}
	if got := threadAppeaseCost(old, 1); !reflect.DeepEqual(got, doomAppeaseCost("electric_era", ages[1], 1)) {
		t.Errorf("a thread with no chain: appease %v", got)
	}
	// The Last Passage's thread keeps the era's price whatever age it is in.
	lp := &HarbingerSave{Age: "galactic_age", Chain: []string{"interstellar_age", "galactic_age"}, EpochKey: "cosmic_era", TargetEpoch: ""}
	if got := threadAppeaseCost(lp, 2); !reflect.DeepEqual(got, eraAppeaseCost("cosmic_era", 2)) {
		t.Errorf("the Last Passage's thread: appease %v, want the era's %v", got, eraAppeaseCost("cosmic_era", 2))
	}
	// Era Mastery does not move the price: a mastered age makes k times as
	// much per tick for a warning k times shorter.
	vet := threadEngine(t, "electric_era", 6)
	vet.Prestige.SetMastery(ages[0], 4)
	if got := vet.GetState().Harbinger.AppeaseCost; !reflect.DeepEqual(got, want) {
		t.Errorf("on known ground the price is %v, on new ground %v", got, want)
	}
	// The table the smoke suite checks against storage has every thread.
	rows := HarbingerPriceTable()
	dooms, passages := 0, 0
	for _, r := range rows {
		if r.LastPassage {
			passages++
			if !reflect.DeepEqual(r.AppeaseL1, eraAppeaseCost(r.Epoch, 1)) {
				t.Errorf("table: Last Passage row %+v", r)
			}
			continue
		}
		dooms++
		if !reflect.DeepEqual(r.AppeaseL1, doomAppeaseCost(r.Epoch, r.Age, 1)) || len(r.BraceL1) == 0 {
			t.Errorf("table: row %+v", r)
		}
	}
	if want := len(config.AgeOrder()) - len(epochAges(t, "stone_era")); dooms != want || passages != 1 {
		t.Errorf("table has %d doom rows and %d Last Passage rows, want %d (every age from the Iron Era on) and 1", dooms, passages, want)
	}
}

// --- Appease ------------------------------------------------------------------------

func TestAppeaseCostsLevelsAndOdds(t *testing.T) {
	ge := threadEngine(t, "steel_era", 4)
	// The thread's own price: what the age its harbinger arrived in makes in
	// the shortest warning.
	price := threadAppeaseCost(ge.harbinger, 1)
	pf, pc := price["faith"], price["culture"]
	if pf <= 0 || pc <= 0 || !reflect.DeepEqual(price, doomAppeaseCost("steel_era", ge.harbinger.startAge(), 1)) {
		t.Fatalf("steel era level 1 = %v", price)
	}
	// Faith: 9.6 prices in a store of 10, so the fill bands below come out.
	setStock(ge, map[string][2]float64{"faith": {9.6 * pf, 10 * pf}, "culture": {13 * pc, 13 * pc}})

	// Level 1. Faith fill 0.86 → high band: a 60% strike, ×0.6.
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	if f, c := ge.Resources.Get("faith"), ge.Resources.Get("culture"); math.Abs(f-8.6*pf) > 1e-6*pf || math.Abs(c-12*pc) > 1e-6*pc {
		t.Errorf("after level 1: faith %v culture %v, want %v and %v", f, c, 8.6*pf, 12*pc)
	}
	if o := ge.CatastropheOutlook(); math.Abs(o.Probability-0.60*0.6) > 1e-9 {
		t.Errorf("level 1 probability %v, want %v", o.Probability, 0.60*0.6)
	}
	if countLogs(ge, "Appease 1/2") != 1 {
		t.Error("no Appease log line")
	}

	// Level 2 costs double. Fill 0.66 → mid band: 75%, ×0.36.
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	left := 6.6 * pf
	if f := ge.Resources.Get("faith"); math.Abs(f-left) > 1e-6*pf {
		t.Errorf("after level 2: faith %v, want %v (level 2 costs double)", f, left)
	}
	if o := ge.CatastropheOutlook(); math.Abs(o.Probability-0.75*0.36) > 1e-9 {
		t.Errorf("level 2 probability %v, want %v", o.Probability, 0.75*0.36)
	}
	if st := ge.GetState(); st.Harbinger.AppeaseLevel != 2 || st.Harbinger.AppeaseBlocked == "" || st.Harbinger.AppeaseCost != nil {
		t.Errorf("view after cap = %+v", st.Harbinger)
	}
	if err := ge.HarbingerAppease(); err == nil || !strings.Contains(err.Error(), "as far as it goes") {
		t.Errorf("third appease: err = %v", err)
	}
	if f := ge.Resources.Get("faith"); math.Abs(f-left) > 1e-6*pf {
		t.Errorf("refused appease deducted faith: %v", f)
	}
}

func TestAppeaseRefusesWhenUnaffordable(t *testing.T) {
	ge := threadEngine(t, "iron_era", 4)
	need := threadAppeaseCost(ge.harbinger, 1)["faith"]
	if need <= 40 {
		t.Fatalf("iron era appease costs %v faith", need)
	}
	setStock(ge, map[string][2]float64{"faith": {20, 40}})
	err := ge.HarbingerAppease()
	want := textfmt.Number(need-20) + " more faith"
	if err == nil || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "storage must reach "+textfmt.Number(need)) {
		t.Fatalf("err = %v, want %q and the storage it needs", err, want)
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
		if err := act(); err == nil || !strings.Contains(err.Error(), "No harbinger") {
			t.Errorf("%s without a harbinger: err = %v", name, err)
		}
	}
	if ge.catastropheInvited {
		t.Error("invite armed without a harbinger")
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
			ge := threadEngine(t, "iron_era", 6)
			fillBraceStock(ge, 1e9)
			price := harbingerBraceCost("iron_era", 1)
			res := sortedKeys(price)[0]
			for i := 1; i <= tc.level; i++ {
				stockBefore := ge.Resources.Get(res)
				if err := ge.HarbingerBrace(); err != nil {
					t.Fatal(err)
				}
				if got, want := stockBefore-ge.Resources.Get(res), harbingerBraceCost("iron_era", i)[res]; got != want {
					t.Errorf("level %d took %v %s, want %v", i, got, res, want)
				}
			}
			if tc.level == HarbingerMaxBrace {
				if err := ge.HarbingerBrace(); err == nil {
					t.Error("brace beyond the cap must be refused")
				}
			}
			ge.rng = riggedRNG(0.01)
			ge.fate.StrikeTick = ge.tick
			tickTo(ge, ge.tick)
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

// Invite makes the doom certain, blocks Appease, leaves Brace open, survives
// the handoffs, and the doom still comes when it must: here at the era's end,
// which the player reaches first.
func TestInviteForcesCatastropheAndDisablesAppease(t *testing.T) {
	ge := threadEngine(t, "iron_era", 7)
	setStock(ge, map[string][2]float64{"faith": {1e6, 1e6}})
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	if !ge.fate.Invited || !ge.harbinger.Invited || ge.catastropheInvited {
		t.Fatal("invite not armed on the fate")
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
	fillBraceStock(ge, 1e9)
	if err := ge.HarbingerBrace(); err != nil {
		t.Errorf("brace after invite: %v", err)
	}
	ge.advanceAge("classical_age")
	ge.advanceAge("medieval_age")
	if !ge.harbinger.Invited || !ge.fate.Invited {
		t.Fatal("the invite did not survive the handoffs")
	}
	ge.rng = riggedRNG(0.999) // a miss under any odds: the invite overrides it
	readyToAdvance(ge)
	if err := ge.AdvanceAge(); err == nil || ge.pendingCatastrophe != "iron_era" {
		t.Fatalf("invited doom at the era's end: err %v pending %q", err, ge.pendingCatastrophe)
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
		roll         float64
		wantOutcome  string
		wantMoment   flavor.Moment
		wantPending  bool
	}{
		{"vindicated", false, false, 0.01, HarbingerOutcomeVindicated, flavor.HarbingerVindicated, true},
		{"spared", false, false, 0.999, HarbingerOutcomeSpared, flavor.HarbingerSpared, false},
		{"discredited", true, false, 0.01, HarbingerOutcomeDiscredited, flavor.HarbingerDiscredited, false},
		{"fulfilled", false, true, 0.999, HarbingerOutcomeFulfilled, flavor.HarbingerFulfilled, true},
		{"fulfilled false prophet", true, true, 0.999, HarbingerOutcomeFulfilled, flavor.HarbingerFulfilled, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ge := fateEngine(t, "medieval_age", 9)
			late := int(baseEraTicks("iron_era")) - 1
			if tc.falseProphet {
				if err := ge.ForceFalseProphetForTest("iron_era", late); err != nil {
					t.Fatal(err)
				}
			} else {
				forceFate(t, ge, late)
			}
			ge.fateArrive()
			if tc.invite {
				if err := ge.HarbingerInvite(); err != nil {
					t.Fatal(err)
				}
			}
			// The player reaches the era's end first: the doom settles there.
			ge.rng = riggedRNG(tc.roll)
			readyToAdvance(ge)
			_ = ge.AdvanceAge()
			if (ge.pendingCatastrophe != "") != tc.wantPending {
				t.Fatalf("pending = %q", ge.pendingCatastrophe)
			}
			if len(ge.harbingerHistory) != 1 {
				t.Fatalf("history = %+v", ge.harbingerHistory)
			}
			r := ge.harbingerHistory[0]
			if r.Outcome != tc.wantOutcome || r.FalseProphet != tc.falseProphet || r.TargetEpochKey != "iron_era" ||
				r.Age != "medieval_age" || !r.AtAdvance || r.Window == 0 {
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
		})
	}
}

// --- persistence and resets ------------------------------------------------------------------

func TestHarbingerSaveLoadMidThread(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := threadEngine(t, "iron_era", 10)
	ge.advanceAge("classical_age")
	setStock(ge, map[string][2]float64{"faith": {1e6, 1e6}})
	fillBraceStock(ge, 1e9)
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	if err := ge.HarbingerBrace(); err != nil {
		t.Fatal(err)
	}
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	want, wantFate := *ge.harbinger, *ge.fate
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
	if ge2.harbinger == nil || !reflect.DeepEqual(*ge2.harbinger, want) || *ge2.fate != wantFate {
		t.Fatalf("thread after load = %+v fate %+v, want %+v / %+v", ge2.harbinger, ge2.fate, want, wantFate)
	}
	if !ge2.fate.Arrived || ge2.catastropheInvited {
		t.Errorf("arrived %v / invite flag %v", ge2.fate.Arrived, ge2.catastropheInvited)
	}
	if countLogs(ge2, "has come")+countLogs(ge2, "takes up") != 0 {
		t.Error("loading spoke again")
	}
	ge2.harbingerTickCheck()
	ge2.advanceAge("medieval_age")
	if !reflect.DeepEqual(ge2.harbinger.Chain, []string{"iron_age", "classical_age", "medieval_age"}) {
		t.Errorf("chain after load and handoff = %v", ge2.harbinger.Chain)
	}
	readyToAdvance(ge2)
	if err := ge2.AdvanceAge(); err == nil || ge2.pendingCatastrophe != "iron_era" || ge2.pendingBraceLevel != 1 {
		t.Errorf("at the era's end: err %v pending %q brace %d", err, ge2.pendingCatastrophe, ge2.pendingBraceLevel)
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
	for _, k := range []string{"harbinger", "catastrophe_invited", "pending_brace_level", "pending_last_passage", "cosmic_legacy", "fate"} {
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
	ge := threadEngine(t, "iron_era", 12)
	ge.rng = riggedRNG(0.01)
	ge.fate.StrikeTick = ge.tick
	tickTo(ge, ge.tick)
	if ge.pendingCatastrophe != "iron_era" {
		t.Fatalf("setup: pending %q", ge.pendingCatastrophe)
	}
	ge.pendingBraceLevel = 2
	if err := ge.Succumb(); err != nil {
		t.Fatal(err)
	}
	if ge.harbinger != nil || ge.fate != nil || ge.catastropheInvited || ge.pendingBraceLevel != 0 || len(ge.harbingerArrived) != 0 {
		t.Errorf("after Succumb: thread %v fate %v invited %v brace %d arrived %v",
			ge.harbinger, ge.fate, ge.catastropheInvited, ge.pendingBraceLevel, ge.harbingerArrived)
	}
	if len(ge.harbingerHistory) != 1 {
		t.Errorf("history should survive Succumb like the epoch history: %+v", ge.harbingerHistory)
	}
	ge.harbingerTickCheck()
	if ge.fate == nil || ge.fate.EpochKey != "stone_era" {
		t.Errorf("new run's first tick: fate = %+v, want the Stone Era's", ge.fate)
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
	if ge.harbinger != nil || ge.fate != nil || ge.catastropheInvited || len(ge.harbingerArrived) != 0 || ge.harbingerHistory != nil {
		t.Errorf("after prestige: thread %v fate %v invited %v arrived %v history %v",
			ge.harbinger, ge.fate, ge.catastropheInvited, ge.harbingerArrived, ge.harbingerHistory)
	}
	ge.harbingerTickCheck()
	if ge.fate == nil || ge.fate.EpochKey != "stone_era" {
		t.Errorf("after prestige, first tick: fate = %+v", ge.fate)
	}
}

// A doom's Appease is sized to its warning, not to the era. In every age a
// harbinger can arrive in, a moderate faith (and culture) economy
// (config.FlowIncome) makes level 1 inside the shortest warning (in three
// quarters of it: 15% of the age), so no full warning is too short for it
// and nothing has to be saved beforehand. Level 2 costs double, so both
// levels together cost more than the shortest warning makes: the stretch a
// longer warning, or faith kept beforehand, pays for. It is never more than
// the longest warning makes.
func TestAppeasePayableWithinTheWarning(t *testing.T) {
	if harbingerAppeaseWindowShare > 1 || 3*harbingerAppeaseWindowShare <= 1 {
		t.Fatalf("window share %v: level 1 must fit the shortest warning, and both levels must not", harbingerAppeaseWindowShare)
	}
	checked := 0
	for _, ep := range config.Epochs() {
		if !config.FateAllowed(ep.Key) {
			continue
		}
		for _, age := range ep.Ages {
			cost1, cost2 := doomAppeaseCost(ep.Key, age, 1), doomAppeaseCost(ep.Key, age, 2)
			if cost1["faith"] <= 0 {
				t.Errorf("%s in %s: level 1 asks no faith (%v)", ep.Key, age, cost1)
			}
			for res, l1 := range cost1 {
				checked++
				if math.Abs(cost2[res]-2*l1) > 1e-6 {
					t.Errorf("%s in %s: level 2 %s = %v, want double %v", ep.Key, age, res, cost2[res], l1)
				}
				rate, span := config.FlowIncome(res, age), config.AgeTargetTicks(age)
				shortest, longest := rate*span*harbingerLeadMin, rate*span*harbingerLeadMax
				// Rounded up to two significant figures: within 10% of the share.
				if l1 > 1.1*shortest*harbingerAppeaseWindowShare || l1 > shortest {
					t.Errorf("%s in %s: level 1 %s (%v) is more than its share of what the shortest warning makes (%v)", ep.Key, age, res, l1, shortest*harbingerAppeaseWindowShare)
				}
				if l1+cost2[res] <= shortest {
					t.Errorf("%s in %s: both levels of %s (%v) fit the shortest warning (%v): level 2 is no stretch", ep.Key, age, res, l1+cost2[res], shortest)
				}
				if l1+cost2[res] > 1.1*longest {
					t.Errorf("%s in %s: both levels of %s (%v) are more than the longest warning makes (%v)", ep.Key, age, res, l1+cost2[res], longest)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no price checked")
	}
}

// The Last Passage's Appease is sized to its era, which the thread lasts: a
// moderate faith (and culture) economy modelled at config.FlowIncome through
// the Cosmic Era's ages at their targets makes level 1 before the era's last
// age ends and level 2 (on top of it) by the era's end.
func TestLastPassageAppeasePayableWithinItsEra(t *testing.T) {
	for _, ep := range config.Epochs() {
		if !config.IsFinalEpoch(ep.Key) {
			continue
		}
		cost1, cost2 := eraAppeaseCost(ep.Key, 1), eraAppeaseCost(ep.Key, 2)
		if len(cost1) == 0 {
			t.Fatalf("%s: the Last Passage's appease asks nothing", ep.Key)
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
