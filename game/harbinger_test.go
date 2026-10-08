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
			for k := range doomAppeaseCost(ep.Key, ep.Ages[0], HarbingerMaxAppease) {
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
	setFaithStrength(ge, 0.5) // mid faith: a real doom would strike 75% of the time
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
	setStock(ge, map[string][2]float64{"culture": {1e7, 1e7}})
	setFaithStrength(ge, 0.5)
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
		// Level 2 costs the same again, in every era.
		const wantTwo = 1.0
		if two := threadBraceCost(ge.harbinger, 2); len(two) != len(first.BraceCost) {
			t.Errorf("%s: level 2 brace resources differ", ep.Key)
		} else {
			for k, v := range first.BraceCost {
				if math.Abs(two[k]-wantTwo*v) > 1 {
					t.Errorf("%s: level 2 %s = %v, want %v times level 1's %v", ep.Key, k, two[k], wantTwo, v)
				}
			}
		}
	}
}

func TestHarbingerCostExamples(t *testing.T) {
	// An ordinary doom's Brace is priced on its shortest warning, in the age
	// its harbinger arrives in: a third of what a moderate economy makes of
	// each material in a fifth of the age. Iron Age stone: 46.5 a tick over
	// 2,340 ticks = 108.9K; a third is 36.3K → 37K. The materials are the
	// era's core resources that the age's buildings make: the Modern Age
	// buys its data at the market, and nothing but a wonder makes crypto, so
	// neither is asked for. No material asks for more than three fifths of
	// what a moderate builder's store holds (five copies of every storage
	// building so far), rounded down to two figures: the Classical Age's
	// store is 1.15M, so its gold and iron stop at 680K where a third of
	// the warning is 920K and 760K. It was 12% of the most the era's
	// advances ask, the same in every age of the era (eraBraceCost): 26.4K
	// stone, 6.36K iron and 21.6K gold for the whole Iron Era.
	dooms1 := []struct {
		epoch, age string
		brace      map[string]float64
	}{
		{"iron_era", "iron_age", map[string]float64{"gold": 140000, "iron": 170000, "stone": 37000}},
		{"iron_era", "classical_age", map[string]float64{"gold": 680000, "iron": 680000, "stone": 51000}},
		{"iron_era", "medieval_age", map[string]float64{"gold": 2800000, "iron": 2800000, "stone": 66000}},
		{"steel_era", "renaissance_age", map[string]float64{"gold": 16000000, "steel": 4500000}},
		{"steel_era", "colonial_age", map[string]float64{"gold": 190000000, "steel": 69000000}},
		{"steel_era", "industrial_age", map[string]float64{"gold": 1200000000, "steel": 1200000000}},
		{"electric_era", "victorian_age", map[string]float64{"electricity": 1700000, "oil": 1400000000, "steel": 5100000000}},
		{"electric_era", "electric_age", map[string]float64{"electricity": 21000000000, "oil": 4600000000, "steel": 21000000000}},
		{"electric_era", "atomic_age", map[string]float64{"electricity": 90000000000, "oil": 6100000000, "steel": 90000000000}},
		{"digital_era", "modern_age", map[string]float64{"electricity": 420000000000, "gold": 420000000000}},
		{"digital_era", "information_age", map[string]float64{"data": 590000000000, "electricity": 7300000000000, "gold": 7300000000000}},
		{"digital_era", "digital_age", map[string]float64{"data": 2400000000000, "electricity": 15000000000000, "gold": 15000000000000}},
		{"neon_era", "cyberpunk_age", map[string]float64{"data": 8100000000000, "electricity": 81000000000000}},
		{"neon_era", "fusion_age", map[string]float64{"data": 9400000000000, "electricity": 190000000000000}},
		{"neon_era", "space_age", map[string]float64{"data": 11000000000000, "electricity": 790000000000000}},
	}
	listed := 0
	for _, ep := range config.Epochs() {
		if config.FateAllowed(ep.Key) && !config.IsFinalEpoch(ep.Key) {
			listed += len(ep.Ages)
		}
	}
	if len(dooms1) != listed {
		t.Fatalf("an ordinary doom's brace is listed for %d ages, want every age of the eras a doom is fated in (%d)", len(dooms1), listed)
	}
	for _, c := range dooms1 {
		if got := doomBraceCost(c.epoch, c.age, 1); !reflect.DeepEqual(got, c.brace) {
			t.Errorf("a doom of the %s foretold in the %s: brace = %v, want %v", c.epoch, c.age, got, c.brace)
		}
		// Level 2 costs the same again.
		if got := doomBraceCost(c.epoch, c.age, 2); !reflect.DeepEqual(got, c.brace) {
			t.Errorf("a doom of the %s foretold in the %s: brace level 2 = %v, want level 1's %v again", c.epoch, c.age, got, c.brace)
		}
		for res := range c.brace {
			if res == "crypto" || res == "faith" || res == "culture" || res == "knowledge" || c.age == "modern_age" && res == "data" {
				t.Errorf("a doom of the %s foretold in the %s asks %s for Brace", c.epoch, c.age, res)
			}
		}
	}
	// The old price, for the record of what it did: the Electric Era's took
	// a moderate economy over an hour and a half of the Victorian Age's
	// electricity and seconds of the Atomic Age's.
	old := eraBraceCost("electric_era", 1)
	if want := map[string]float64{"steel": 56400000, "oil": 924000, "electricity": 3960000}; !reflect.DeepEqual(old, want) {
		t.Errorf("the Electric Era's old price = %v, want %v", old, want)
	}
	first, last := epochAges(t, "electric_era")[0], epochAges(t, "electric_era")[2]
	if early, late := old["electricity"]/config.TypicalIncome("electricity", first), old["electricity"]/config.TypicalIncome("electricity", last); early < 1000 || late > 1 {
		t.Errorf("the Electric Era's old price took %.0f ticks of %s electricity and %.2f of %s's; the collapse the rule replaced is gone from the numbers", early, first, late, last)
	}

	// The Cosmic Era's fated doom, the Reality Tear, is priced on its own
	// warning: five sixths of what the age its harbinger arrives in makes in
	// a doom's shortest warning, a sixth of the age. Interstellar Age dark
	// matter: 237B a tick over 22,464 ticks = 5.32Q; five sixths is 4.44Q →
	// 4.5Q. The Galactic Age makes 21 times the dark matter and no more
	// titanium. It was the era's 1.56T dark matter and 75.6B titanium.
	tears := []struct {
		age   string
		brace map[string]float64
	}{
		// What the ages make with the tech layer and the payback curve that
		// came with it counted: about twice the dark matter and titanium.
		// Then the soft cap: the all-production pool the model holds is past
		// +200% in these ages and a quarter of the rest counts, 5% more
		// income in the Interstellar Age, 10%, 15% and 20% in the three after.
		// Then the tree's last techs: Asteroid Refining and Stellar Core
		// Mining add 8% to titanium.
		{"interstellar_age", map[string]float64{"dark_matter": 5.1e+15, "titanium": 4.8e+15}},
		{"galactic_age", map[string]float64{"dark_matter": 6.1e+16, "titanium": 5.1e+15}},
		{"quantum_age", map[string]float64{"dark_matter": 6.3e+16, "titanium": 5.3e+15}},
		{"transcendent_age", map[string]float64{"dark_matter": 6.9e+16, "titanium": 5.8e+15}},
	}
	if ages := epochAges(t, "cosmic_era"); len(ages) != len(tears) {
		t.Fatalf("the Cosmic Era has ages %v; the Reality Tear's brace is listed for %d", ages, len(tears))
	}
	for _, c := range tears {
		if got := doomBraceCost("cosmic_era", c.age, 1); !reflect.DeepEqual(got, c.brace) {
			t.Errorf("the Reality Tear foretold in the %s: brace = %v, want %v", c.age, got, c.brace)
		}
		// Level 2 costs the same again.
		if got := doomBraceCost("cosmic_era", c.age, 2); !reflect.DeepEqual(got, c.brace) {
			t.Errorf("the Reality Tear foretold in the %s: brace level 2 = %v, want level 1's %v again", c.age, got, c.brace)
		}
	}

	// A doom's Appease is priced on its warning: three quarters of what the
	// age its harbinger arrives in makes in the shortest warning, 15% of the
	// age. Iron Age faith: 0.75 a tick, 0.9 with Ritual's and Calendar's
	// +10% each, over 11,700 ticks = 10,530; 15% is 1,580 → 1,600. The Iron
	// Era asks no culture (it arrives in the Classical Age, after the era
	// began).
	dooms := []struct {
		epoch, age string
		appease    map[string]float64
	}{
		{"iron_era", "iron_age", map[string]float64{"faith": 1600}},
		// The model counts the tech layer: Ritual's and Calendar's +10% faith
		// from the Bronze Age, Theology's +8% from the Medieval Age,
		// the culture of Patronage, Romanticism, Museums, Cinema and Social
		// Media, the two +5% on all production.
		{"iron_era", "medieval_age", map[string]float64{"faith": 5000}},
		{"steel_era", "renaissance_age", map[string]float64{"faith": 11000, "culture": 96000}},
		{"steel_era", "industrial_age", map[string]float64{"faith": 110000, "culture": 1100000}},
		// The Cosmic Era's row also carries the soft cap: the model's pool is
		// past +200% there and a quarter of the rest counts, 5% more income.
		{"cosmic_era", "interstellar_age", map[string]float64{"faith": 380000000, "culture": 5800000000}},
	}
	for _, c := range dooms {
		if got := doomAppeaseCost(c.epoch, c.age, 1); !reflect.DeepEqual(got, c.appease) {
			t.Errorf("a doom of the %s foretold in the %s: appease = %v, want %v", c.epoch, c.age, got, c.appease)
		}
	}

	// The Last Passage's thread is priced on its own warning, two thirds of
	// the age its harbinger arrives in (74,880 of the Interstellar Age's
	// 112,320 ticks). Appease is three quarters of what the warning makes,
	// half of what the age does: culture at 239,958 a tick makes 17.97B in
	// the warning; 75% is 13.5B → 14B. Brace is half of what the warning
	// makes, a third of the age: dark matter at 237B a tick makes 17.7Q in
	// the warning; half is 8.87Q → 8.9Q. Its thread comes with the era's
	// first age; the later rows price a thread that began there (an older
	// save, the dev console): the Galactic Age makes 21 times the dark
	// matter and no more titanium. Appease was 3.1B faith and 48B culture in
	// every age (a quarter of what the Interstellar, Galactic and Quantum
	// Ages make together), then 1.4B and 21B on the whole Interstellar Age;
	// Brace was the era's 1.56T dark matter and 75.6B titanium.
	passages := []struct {
		age            string
		appease, brace map[string]float64
	}{
		// With the tech layer in the model (culture +32%, faith +33%, dark
		// matter +9% and titanium +9% by the Interstellar Age; Transcendence
		// adds 5% to all of them): the rows above, that much dearer. Brace is
		// dearer again by what the payback curve added to what the age makes
		// of dark matter and titanium (about twice as much there). And by
		// the soft cap, as the Reality Tear's rows above: 5% to 20%.
		// And by the tree's last techs: 4% more culture each from Neural
		// Art, Galactic Memory and Reality Art, 8% more titanium.
		{"interstellar_age", map[string]float64{"faith": 1300000000, "culture": 20000000000}, map[string]float64{"dark_matter": 1.1e+16, "titanium": 9.6e+15}},
		{"galactic_age", map[string]float64{"faith": 2600000000, "culture": 42000000000}, map[string]float64{"dark_matter": 1.3e+17, "titanium": 1.1e+16}},
		{"quantum_age", map[string]float64{"faith": 5400000000, "culture": 89000000000}, map[string]float64{"dark_matter": 1.3e+17, "titanium": 1.1e+16}},
		{"transcendent_age", map[string]float64{"faith": 5900000000, "culture": 97000000000}, map[string]float64{"dark_matter": 1.4e+17, "titanium": 1.2e+16}},
	}
	if ages := epochAges(t, "cosmic_era"); len(ages) != len(passages) {
		t.Fatalf("the Cosmic Era has ages %v; the Last Passage's price is listed for %d", ages, len(passages))
	}
	for _, c := range passages {
		if got := lastPassageAppeaseCost("cosmic_era", c.age, 1); !reflect.DeepEqual(got, c.appease) {
			t.Errorf("the Last Passage foretold in the %s: appease = %v, want %v", c.age, got, c.appease)
		}
		if got := lastPassageBraceCost("cosmic_era", c.age, 1); !reflect.DeepEqual(got, c.brace) {
			t.Errorf("the Last Passage foretold in the %s: brace = %v, want %v", c.age, got, c.brace)
		}
	}
}

// A thread's Appease price is its own: it is priced on the age its harbinger
// arrived in and stays the same in every age the thread lives through, a
// doom's on a doom's warning and the Last Passage's on its own.
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
	// The Last Passage's thread keeps the price of the age it arrived in
	// whatever age it is in, and it is not a doom's price.
	lp := &HarbingerSave{Age: "galactic_age", Chain: []string{"interstellar_age", "galactic_age"}, EpochKey: "cosmic_era", TargetEpoch: ""}
	if got, want := threadAppeaseCost(lp, 2), lastPassageAppeaseCost("cosmic_era", "interstellar_age", 2); !reflect.DeepEqual(got, want) || len(want) != 2 {
		t.Errorf("the Last Passage's thread: appease %v, want the arrival age's %v", got, want)
	}
	if reflect.DeepEqual(threadAppeaseCost(lp, 2), doomAppeaseCost("cosmic_era", "interstellar_age", 2)) {
		t.Error("the Last Passage's thread is priced as a doom's")
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
			if !reflect.DeepEqual(r.AppeaseL1, lastPassageAppeaseCost(r.Epoch, r.Age, 1)) || !reflect.DeepEqual(r.AppeaseL2, lastPassageAppeaseCost(r.Epoch, r.Age, 2)) ||
				!reflect.DeepEqual(r.BraceL1, lastPassageBraceCost(r.Epoch, r.Age, 1)) || !reflect.DeepEqual(r.BraceL2, lastPassageBraceCost(r.Epoch, r.Age, 2)) ||
				len(r.BraceL1) == 0 || !r.FinalEra ||
				r.WarningTicks != lastPassageWarning*config.AgeTargetTicks(r.Age) || !config.IsFinalEpoch(r.Epoch) {
				t.Errorf("table: Last Passage row %+v", r)
			}
			continue
		}
		dooms++
		if !reflect.DeepEqual(r.AppeaseL1, doomAppeaseCost(r.Epoch, r.Age, 1)) || !reflect.DeepEqual(r.AppeaseL2, doomAppeaseCost(r.Epoch, r.Age, 2)) ||
			!reflect.DeepEqual(r.BraceL1, doomBraceCost(r.Epoch, r.Age, 1)) || !reflect.DeepEqual(r.BraceL2, doomBraceCost(r.Epoch, r.Age, 2)) ||
			len(r.BraceL1) == 0 || r.FinalEra != config.IsFinalEpoch(r.Epoch) ||
			r.WarningTicks != harbingerLeadMin*config.AgeTargetTicks(r.Age) {
			t.Errorf("table: row %+v", r)
		}
	}
	wantDooms, wantPassages := len(config.AgeOrder())-len(epochAges(t, "stone_era")), len(epochAges(t, "cosmic_era"))
	if dooms != wantDooms || passages != wantPassages {
		t.Errorf("table has %d doom rows and %d Last Passage rows, want %d (every age from the Iron Era on) and %d (every age of the Cosmic Era)", dooms, passages, wantDooms, wantPassages)
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
	// Faith: ten prices, every one of them made by the town's own faith
	// buildings, which did four and a half times what a moderate set would
	// have: full faith strength while it keeps it all. What it pays comes off
	// the share kept, so the bands below come out.
	setStock(ge, map[string][2]float64{"culture": {13 * pc, 13 * pc}})
	ge.setFaithMeasure(10*pf, FaithSave{Moderate: 10 * pf / FaithFullSets, Own: 10 * pf})

	// Level 1. Nine prices of ten are left: strength 0.9 → high band: a 60%
	// strike, ×0.6.
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	if f, c := ge.Resources.Get("faith"), ge.Resources.Get("culture"); math.Abs(f-9*pf) > 1e-6*pf || math.Abs(c-12*pc) > 1e-6*pc {
		t.Errorf("after level 1: faith %v culture %v, want %v and %v", f, c, 9*pf, 12*pc)
	}
	if o := ge.CatastropheOutlook(); math.Abs(o.Probability-0.60*0.6) > 1e-9 {
		t.Errorf("level 1 probability %v, want %v", o.Probability, 0.60*0.6)
	}
	if countLogs(ge, "Appease 1/2") != 1 {
		t.Error("no Appease log line")
	}

	// Level 2 costs the same again. Eight prices of ten are left: strength
	// 0.8 → still the high band: 60%, ×0.36. (At double, seven were left and
	// the town fell to the mid band, 75%: the second level paid for itself
	// in strength lost.)
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	left := 8 * pf
	if f := ge.Resources.Get("faith"); math.Abs(f-left) > 1e-6*pf {
		t.Errorf("after level 2: faith %v, want %v (level 2 costs the same again)", f, left)
	}
	if o := ge.CatastropheOutlook(); math.Abs(o.Probability-0.60*0.36) > 1e-9 {
		t.Errorf("level 2 probability %v, want %v", o.Probability, 0.60*0.36)
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
			price := threadBraceCost(ge.harbinger, 1)
			res := sortedKeys(price)[0]
			for i := 1; i <= tc.level; i++ {
				stockBefore := ge.Resources.Get(res)
				if err := ge.HarbingerBrace(); err != nil {
					t.Fatal(err)
				}
				if got, want := stockBefore-ge.Resources.Get(res), price[res]; got != want || want <= 0 {
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
// and nothing has to be saved beforehand. Level 2 costs the same again, so
// both levels together cost more than the shortest warning makes (one and
// a half times it): the stretch a longer warning, or faith kept
// beforehand, pays for. It is never more than the longest warning makes.
func TestAppeasePayableWithinTheWarning(t *testing.T) {
	if harbingerAppeaseWindowShare > 1 || 2*harbingerAppeaseWindowShare <= 1 {
		t.Fatalf("window share %v: level 1 must fit the shortest warning, and both levels must not", harbingerAppeaseWindowShare)
	}
	checked := 0
	for _, ep := range config.Epochs() {
		if !config.FateAllowed(ep.Key) {
			continue
		}
		const wantTwo = 1.0
		for _, age := range ep.Ages {
			cost1, cost2 := doomAppeaseCost(ep.Key, age, 1), doomAppeaseCost(ep.Key, age, 2)
			if cost1["faith"] <= 0 {
				t.Errorf("%s in %s: level 1 asks no faith (%v)", ep.Key, age, cost1)
			}
			for res, l1 := range cost1 {
				checked++
				if math.Abs(cost2[res]-wantTwo*l1) > 1e-6 {
					t.Errorf("%s in %s: level 2 %s = %v, want %v times level 1's %v", ep.Key, age, res, cost2[res], wantTwo, l1)
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

// The Last Passage's Appease is sized to its own warning, two thirds of the
// age its harbinger arrives in, by the rule a doom's is: in every age of the
// Cosmic Era the thread can begin in, a moderate faith and culture economy
// (config.FlowIncome) makes level 1 inside that warning (in three quarters of
// it, half the age), and levels 1 and 2 together (level 2 costs the same
// again) cost more than the warning makes: the stretch that playing the age
// out, or a stock kept beforehand, pays for. It is dearer than a doom's Appease foretold in the
// same age, by the ratio of the two warnings, and than the thread's own
// Brace, measured in the time a moderate economy needs to make each.
func TestLastPassageAppeasePayableWithinItsWarning(t *testing.T) {
	if lastPassageWarning <= harbingerLeadMax {
		t.Fatalf("the Last Passage's warning (%v of its age) must be longer than any doom's (up to %v)", lastPassageWarning, harbingerLeadMax)
	}
	checked := 0
	for _, ep := range config.Epochs() {
		if !config.IsFinalEpoch(ep.Key) {
			continue
		}
		for _, age := range ep.Ages {
			cost1, cost2 := lastPassageAppeaseCost(ep.Key, age, 1), lastPassageAppeaseCost(ep.Key, age, 2)
			doom := doomAppeaseCost(ep.Key, age, 1)
			if cost1["faith"] <= 0 || cost1["culture"] <= 0 || len(cost1) != len(doom) {
				t.Fatalf("%s in %s: the Last Passage's appease asks %v, a doom's %v", ep.Key, age, cost1, doom)
			}
			slowest := 0.0 // ticks a moderate economy needs for level 1
			for res, l1 := range cost1 {
				checked++
				if cost2[res] != l1 {
					t.Errorf("%s in %s: level 2 %s = %v, want level 1's %v again", ep.Key, age, res, cost2[res], l1)
				}
				rate := config.FlowIncome(res, age)
				warning := rate * config.AgeTargetTicks(age) * lastPassageWarning
				// Rounded up to two significant figures: within 10% of the share.
				if l1 > 1.1*warning*harbingerAppeaseWindowShare || l1 > warning {
					t.Errorf("%s in %s: level 1 %s (%v) is more than its share of what the warning makes (%v)", ep.Key, age, res, l1, warning*harbingerAppeaseWindowShare)
				}
				if l1+cost2[res] <= warning {
					t.Errorf("%s in %s: both levels of %s (%v) fit the warning (%v): level 2 is no stretch", ep.Key, age, res, l1+cost2[res], warning)
				}
				// Each figure is rounded up to two significant figures, so
				// the ratio lands within 10% of the warnings' ratio.
				ratio, want := l1/doom[res], lastPassageWarning/harbingerLeadMin
				if l1 <= doom[res] || ratio < want/1.1 || ratio > want*1.1 {
					t.Errorf("%s in %s: %s costs %v against a doom's %v (x%.2f), want about x%.0f", ep.Key, age, res, l1, doom[res], ratio, want)
				}
				slowest = math.Max(slowest, l1/rate)
			}
			for res, c := range lastPassageBraceCost(ep.Key, age, 1) {
				rate := config.TypicalIncome(res, age)
				if rate <= 0 {
					t.Fatalf("%s in %s: nothing makes %s, which Brace asks for", ep.Key, age, res)
				}
				if c/rate >= slowest {
					t.Errorf("%s in %s: Brace's %s takes a moderate economy %.0f ticks, Appease %.0f: Appease must be the dearer", ep.Key, age, res, c/rate, slowest)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no price checked")
	}
}

// In the final era Brace is sized to the thread's warning, as its Appease
// is: in every age of the Cosmic Era a thread can begin in, level 1 is a
// share of what a moderate economy (config.TypicalIncome) makes of each
// resource in the warning, so the warning pays for it. The Last Passage's is
// half of its warning (two thirds of the age), the Reality Tear's five
// sixths of a doom's shortest (a fifth of the age), which makes it half the
// Last Passage's. Level 2 costs the same again. Both ask for the resources
// the era's own price names, and far more of them: an age's requirement,
// which that price is a share of, is a few ticks of the Cosmic Era's income.
func TestFinalEraBracePricedOnTheWarning(t *testing.T) {
	checked := 0
	for _, ep := range config.Epochs() {
		if !config.IsFinalEpoch(ep.Key) {
			continue
		}
		era := eraBraceCost(ep.Key, 1)
		for _, age := range ep.Ages {
			for _, th := range []struct {
				name           string
				cost1, cost2   map[string]float64
				warning, share float64
			}{
				{"the Last Passage", lastPassageBraceCost(ep.Key, age, 1), lastPassageBraceCost(ep.Key, age, 2), lastPassageWarning, lastPassageBraceShare},
				{"the era's doom", doomBraceCost(ep.Key, age, 1), doomBraceCost(ep.Key, age, 2), harbingerLeadMin, finalDoomBraceShare},
			} {
				if len(th.cost1) == 0 || len(th.cost1) != len(era) || len(th.cost2) != len(th.cost1) {
					t.Fatalf("%s in %s: %s's brace asks %v (level 2 %v), the era's price %v", ep.Key, age, th.name, th.cost1, th.cost2, era)
				}
				for res, l1 := range th.cost1 {
					checked++
					if res == "faith" || res == "culture" {
						t.Errorf("%s in %s: %s's brace asks %s, which is Appease's", ep.Key, age, th.name, res)
					}
					if th.cost2[res] != l1 {
						t.Errorf("%s in %s: %s's level 2 %s = %v, want level 1's %v again", ep.Key, age, th.name, res, th.cost2[res], l1)
					}
					made := config.TypicalIncome(res, age) * config.AgeTargetTicks(age) * th.warning
					// Rounded up to two significant figures: within 10% of the share.
					if l1 < made*th.share || l1 > 1.1*made*th.share || l1 > made {
						t.Errorf("%s in %s: %s's level 1 %s (%v) is not its share of what the warning makes (%v)", ep.Key, age, th.name, res, l1, made*th.share)
					}
					if l1 <= 1000*era[res] {
						t.Errorf("%s in %s: %s's %s costs %v against the era's %v: it is priced on income, not on the era's requirements", ep.Key, age, th.name, res, l1, era[res])
					}
				}
			}
			// The doom's is half the Last Passage's, give or take the rounding.
			for res, lp := range lastPassageBraceCost(ep.Key, age, 1) {
				if ratio := doomBraceCost(ep.Key, age, 1)[res] / lp; ratio < 0.45 || ratio > 0.55 {
					t.Errorf("%s in %s: the doom's %s is %.2f of the Last Passage's, want about half", ep.Key, age, res, ratio)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no price checked")
	}
}

// A thread's Brace price is its own: by the age the thread's harbinger
// arrived in, the same in every age it lives through; an ordinary doom's a
// third of what its warning makes, and in the Cosmic Era the Last Passage's
// twice the Reality Tear's.
func TestThreadBraceCost(t *testing.T) {
	// A doom's thread: the price of the age it began in, not the era's old
	// one, and dearer the later it begins.
	ages := epochAges(t, "electric_era")
	first := doomBraceCost("electric_era", ages[0], 1)
	ge := threadEngine(t, "electric_era", 6)
	if got := ge.GetState().Harbinger.BraceCost; !reflect.DeepEqual(got, first) || len(first) == 0 || reflect.DeepEqual(first, eraBraceCost("electric_era", 1)) {
		t.Fatalf("a doom's thread started in %s: brace %v, want %v (not the era's old %v)", ages[0], got, first, eraBraceCost("electric_era", 1))
	}
	late := &HarbingerSave{Age: ages[2], Chain: []string{ages[1], ages[2]}, EpochKey: "electric_era", TargetEpoch: "electric_era"}
	second := doomBraceCost("electric_era", ages[1], 1)
	if got := threadBraceCost(late, 1); !reflect.DeepEqual(got, second) {
		t.Errorf("a doom foretold in %s and living in %s: brace %v, want the price of the age it was foretold in, %v", ages[1], ages[2], got, second)
	}
	for res, v := range first {
		if second[res] <= v {
			t.Errorf("a doom foretold in %s asks %v %s, no more than one foretold in %s (%v): the price follows what the age makes", ages[1], second[res], res, ages[0], v)
		}
	}

	// The Last Passage's thread keeps the price of the age it arrived in
	// through every handoff, at both levels.
	cosmic := epochAges(t, "cosmic_era")
	lp := threadEngine(t, "cosmic_era", 6)
	want1, want2 := lastPassageBraceCost("cosmic_era", cosmic[0], 1), lastPassageBraceCost("cosmic_era", cosmic[0], 2)
	if got := lp.GetState().Harbinger.BraceCost; !reflect.DeepEqual(got, want1) || len(want1) != 2 {
		t.Fatalf("the Last Passage's thread: brace %v, want %v", got, want1)
	}
	for _, a := range cosmic[1:] {
		lp.advanceAge(a)
		if lp.harbinger == nil || lp.harbinger.Age != a || lp.harbinger.startAge() != cosmic[0] {
			t.Fatalf("no handoff to %s (thread %+v)", a, lp.harbinger)
		}
		if got := lp.GetState().Harbinger.BraceCost; !reflect.DeepEqual(got, want1) {
			t.Errorf("in %s the price is %v; it was set at %v when the harbinger arrived", a, got, want1)
		}
	}
	// Level 1 takes level 1's price and leaves the view asking for level 2's,
	// which is the same again.
	if !reflect.DeepEqual(want2, want1) {
		t.Errorf("the Last Passage's level 2 asks %v, want level 1's %v again", want2, want1)
	}
	setStock(lp, map[string][2]float64{"dark_matter": {1e18, 1e18}, "titanium": {1e18, 1e18}})
	if err := lp.HarbingerBrace(); err != nil {
		t.Fatal(err)
	}
	for res, c := range want1 {
		if paid := 1e18 - lp.Resources.Get(res); paid != c {
			t.Errorf("level 1 took %v %s, want %v", paid, res, c)
		}
	}
	if got := lp.GetState().Harbinger.BraceCost; !reflect.DeepEqual(got, want2) || lp.harbinger.BraceLevel != 1 {
		t.Errorf("after level 1 the view asks %v (level %d), want level 2's %v", got, lp.harbinger.BraceLevel, want2)
	}
	// A thread that began later in the era (an older save, the dev console)
	// is priced on that age; one with no chain on its current speaker.
	later := &HarbingerSave{Age: cosmic[2], Chain: []string{cosmic[1], cosmic[2]}, EpochKey: "cosmic_era", TargetEpoch: ""}
	if got, want := threadBraceCost(later, 1), lastPassageBraceCost("cosmic_era", cosmic[1], 1); !reflect.DeepEqual(got, want) || reflect.DeepEqual(got, want1) {
		t.Errorf("a Last Passage thread that began in %s: brace %v, want that age's %v (not %v)", cosmic[1], got, want, want1)
	}
	old := &HarbingerSave{Age: cosmic[1], EpochKey: "cosmic_era", TargetEpoch: ""}
	if got, want := threadBraceCost(old, 2), lastPassageBraceCost("cosmic_era", cosmic[1], 2); !reflect.DeepEqual(got, want) {
		t.Errorf("a Last Passage thread with no chain: brace %v, want %v", got, want)
	}
	// Era Mastery does not move the price: a mastered age makes k times as
	// much per tick for a warning k times shorter.
	vet := threadEngine(t, "cosmic_era", 6)
	vet.Prestige.SetMastery(cosmic[0], 4)
	if got := vet.GetState().Harbinger.BraceCost; !reflect.DeepEqual(got, want1) {
		t.Errorf("on known ground the price is %v, on new ground %v", got, want1)
	}

	// The Cosmic Era's own doom is priced on its own warning, from the age
	// its harbinger arrived in, while the Last Passage's thread waits behind
	// it with its own price.
	both := cosmicDoom(t, cosmic[0], 60000)
	tickTo(both, arrivalTick(both))
	if h, parked := both.harbinger, both.parkedHarbinger; h == nil || h.TargetEpoch != "cosmic_era" || parked == nil || parked.TargetEpoch != "" {
		t.Fatalf("setup: live %+v parked %+v", h, parked)
	}
	tear := doomBraceCost("cosmic_era", cosmic[0], 1)
	if got := both.GetState().Harbinger.BraceCost; !reflect.DeepEqual(got, tear) || len(tear) != 2 || reflect.DeepEqual(tear, eraBraceCost("cosmic_era", 1)) {
		t.Errorf("the Reality Tear's thread: brace %v, want %v (not the era's %v)", got, tear, eraBraceCost("cosmic_era", 1))
	}
	if got := threadBraceCost(both.parkedHarbinger, 1); !reflect.DeepEqual(got, want1) {
		t.Errorf("the waiting Last Passage thread: brace %v, want %v", got, want1)
	}
	// The era's old price no longer pays for it.
	setStock(both, map[string][2]float64{"dark_matter": {1e13, 1e18}, "titanium": {1e13, 1e18}})
	if err := both.HarbingerBrace(); err == nil || both.harbinger.BraceLevel != 0 {
		t.Errorf("bracing against the Reality Tear with 10T of each: err %v, level %d; want refused", err, both.harbinger.BraceLevel)
	}
	setStock(both, map[string][2]float64{"dark_matter": {1e18, 1e18}, "titanium": {1e18, 1e18}})
	for level := 1; level <= HarbingerMaxBrace; level++ {
		before := both.Resources.Get("dark_matter")
		if err := both.HarbingerBrace(); err != nil {
			t.Fatalf("bracing against the Reality Tear, level %d: %v", level, err)
		}
		if paid := before - both.Resources.Get("dark_matter"); paid != tear["dark_matter"] {
			t.Errorf("the Reality Tear's brace level %d took %v dark matter, want %v each level", level, paid, tear["dark_matter"])
		}
	}
	if both.harbinger.BraceLevel != 2 || both.parkedHarbinger.BraceLevel != 0 {
		t.Errorf("brace levels: doom %d, Last Passage %d, want 2 and 0", both.harbinger.BraceLevel, both.parkedHarbinger.BraceLevel)
	}
	// The thread keeps its price when the age moves on.
	later = &HarbingerSave{Age: cosmic[1], Chain: []string{cosmic[0], cosmic[1]}, EpochKey: "cosmic_era", TargetEpoch: "cosmic_era"}
	if got := threadBraceCost(later, 1); !reflect.DeepEqual(got, tear) {
		t.Errorf("the Reality Tear's thread in %s: brace %v, want the arrival age's %v", cosmic[1], got, tear)
	}
}
