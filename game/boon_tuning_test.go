package game

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/boon"
	"github.com/espresso20/ageforge/config"
)

// Measurement harness for the faction-encounter BOON loop.
//
// boon_soak_test.go asks "is the system BOUNDED?". This file asks the different
// and softer question "does the system FEEL right?" — it drives realistic play
// and REPORTS distributions rather than gating on them. Its assertions are loose
// sanity checks only (something happened, the outcome buckets add up); the value
// is in the t.Logf tables. Run it with:
//
//	go test ./game/ -run TestBoonTuning -v
//
// What it simulates, per scenario:
//   - the real expedition cadence for that age — the SHORTEST available
//     expedition per category (the spam-optimal choice a continuous explorer
//     actually makes), re-launched the moment it resolves (or after a gap, for
//     the casual scenario);
//   - the real success roll (successRoll > DifficultyBase, no military bonus);
//   - the real encounter path: rollExpeditionEncounter → applyFactionBoon /
//     applyFactionMalus → boon.RollBoon → boon.Apply → InjectEvent;
//   - real per-tick expiry via EventManager.Tick.
//
// The economy is deliberately NOT simulated (no building production, no
// expedition rewards, no launch costs): stockpiles are pre-filled and rates left
// alone so that every observed resource DELTA around an encounter is attributable
// to the boon engine. That is what makes setback severity measurable.
//
// Outcome classification is done off the returned player-facing lines, matched
// against signatures derived from the catalogs themselves (see
// boonFlavorSignatures), so it does not rot when flavour text is edited.

const (
	// tuningTicks is the driven-tick budget per scenario. At the fastest real
	// cadence this is ~20k expedition resolutions and several thousand
	// encounters — enough that the outcome split is stable to well under a
	// percentage point.
	tuningTicks = 200_000
	// tuningShortTicks is the budget under -short: enough to prove the harness
	// runs end to end, not enough for stable statistics.
	tuningShortTicks = 20_000
	// tuningSampleEvery is the tick stride for the (expensive) pool samples used
	// for the uplift distribution. The %-of-time counters are sampled EVERY tick.
	tuningSampleEvery = 25
	// tuningStockFill is the amount every resource is pre-loaded with (and the
	// storage cap raised to) so a ResourceDrain has something to bite and an
	// InstantResource grant is not silently clamped away.
	tuningStockFill = 1e9
	// tuningStartPop is the starting worker head-count, big enough that repeated
	// WorkerLoss maluses cannot floor the population and hide their own severity.
	tuningStartPop = 5000
)

// --- flavour-line classification -------------------------------------------

// boonFlavorSignatures maps a distinctive literal fragment of every catalog
// flavour template to that entry's Name. Built from boon.Catalog() /
// boon.MalusCatalog() at run time, so editing flavour text cannot silently break
// classification: the longest brace-free run of each template is the signature.
func boonFlavorSignatures(defs []boon.Def) map[string]string {
	sigs := make(map[string]string)
	for _, d := range defs {
		for _, tmpl := range d.Flavors {
			if sig := longestLiteralRun(tmpl); len(sig) >= 8 {
				sigs[sig] = d.Name
			}
		}
	}
	return sigs
}

// longestLiteralRun returns the longest substring of tmpl that contains no
// {placeholder}. Trailing/leading spaces are kept — they add specificity.
func longestLiteralRun(tmpl string) string {
	best := ""
	cur := strings.Builder{}
	inBrace := false
	flush := func() {
		if s := cur.String(); len(s) > len(best) {
			best = s
		}
		cur.Reset()
	}
	for _, r := range tmpl {
		switch {
		case r == '{':
			inBrace = true
			flush()
		case r == '}':
			inBrace = false
		case !inBrace:
			cur.WriteRune(r)
		}
	}
	flush()
	return best
}

// classifyLine buckets one player-facing encounter line. Returns the bucket and,
// for boon/setback lines, the catalog entry Name that produced it.
func classifyLine(line string, boonSigs, malusSigs map[string]string) (bucket, name string) {
	if strings.Contains(line, "First contact:") {
		return "discovery", ""
	}
	for _, flavor := range atCapacityFlavors {
		if strings.Contains(line, flavor) {
			return "bounce", ""
		}
	}
	for sig, n := range boonSigs {
		if strings.Contains(line, sig) {
			return "boon", n
		}
	}
	for sig, n := range malusSigs {
		if strings.Contains(line, sig) {
			return "setback", n
		}
	}
	return "unclassified", ""
}

// --- scenario definition ----------------------------------------------------

// tuningStanding is the diplomatic state a scenario installs for one faction.
type tuningStanding struct {
	status string
	atWar  bool
}

// tuningScenario describes one representative style of play.
type tuningScenario struct {
	name string
	age  string
	// standings keys factions by roster key. Any eligible faction absent from
	// the map is left UNDISCOVERED (so the run opens with discovery encounters).
	standings map[string]tuningStanding
	// scoutGap / milGap are the idle ticks between one expedition resolving and
	// the next launching. 0 = back-to-back. Negative = that category is unused.
	scoutGap int
	milGap   int
}

func tuningScenarios() []tuningScenario {
	mid := map[string]tuningStanding{
		"riverlands_tribes":  {status: "allied"},
		"ironhold_clans":     {status: "friendly"},
		"merchant_guild":     {status: "neutral"},
		"artisan_league":     {status: "friendly"},
		"atomic_directorate": {status: "neutral"},
	}
	midAtWar := map[string]tuningStanding{
		"riverlands_tribes":  {status: "allied"},
		"ironhold_clans":     {status: "rival", atWar: true},
		"merchant_guild":     {status: "rival", atWar: true},
		"artisan_league":     {status: "friendly"},
		"atomic_directorate": {status: "neutral"},
	}
	allied := map[string]tuningStanding{}
	for _, def := range config.BaseFactions() {
		allied[def.Key] = tuningStanding{status: "allied"}
	}

	return []tuningScenario{
		{
			name: "A early-explorer (iron, neutral, back-to-back)",
			age:  "iron_age",
			// Nothing pre-discovered: the run opens on first contact, exactly
			// like a real early game.
			standings: nil,
			scoutGap:  0,
			milGap:    0,
		},
		{
			name:      "B mid-game (atomic, 5 civs, mixed standing)",
			age:       "atomic_age",
			standings: mid,
			scoutGap:  0,
			milGap:    0,
		},
		{
			name:      "C late allied (quantum, all civs allied)",
			age:       "quantum_age",
			standings: allied,
			scoutGap:  0,
			milGap:    0,
		},
		{
			name:      "D casual (atomic, 5 civs, scout every ~1000 ticks)",
			age:       "atomic_age",
			standings: mid,
			scoutGap:  1000,
			milGap:    -1,
		},
		{
			name:      "E at-war (atomic, 2 of 5 civs at war)",
			age:       "atomic_age",
			standings: midAtWar,
			scoutGap:  0,
			milGap:    0,
		},
	}
}

// --- measured output --------------------------------------------------------

type tuningStats struct {
	name string

	ticks       int
	resolutions int
	encounters  int
	discoveries int

	boons        int
	bounces      int
	setbacks     int
	unclassified int

	ticksAnyBoon int
	ticksFullCap int

	kindCounts  map[string]int
	malusCounts map[string]int

	// Uplift samples, taken only on ticks where at least one boon is live.
	prodAll  []float64 // 1 + Σ production_all from faction boons
	bestRate []float64 // 1 + max Σ <res>_rate from faction boons
	combined []float64 // product of the two, i.e. what the best resource feels

	// Setback severity.
	workersLost int
	drainFracs  []float64 // fraction of a stockpile actually removed
	negMags     []float64 // magnitude of timed negative effects (absolute)
	negDurs     []int     // duration of timed negative effects
}

func newTuningStats(name string) *tuningStats {
	return &tuningStats{
		name:        name,
		kindCounts:  map[string]int{},
		malusCounts: map[string]int{},
	}
}

func (s *tuningStats) per10k(n int) float64 {
	if s.ticks == 0 {
		return 0
	}
	return float64(n) * 10000 / float64(s.ticks)
}

func (s *tuningStats) pctOfEncounters(n int) float64 {
	if s.encounters == 0 {
		return 0
	}
	return float64(n) * 100 / float64(s.encounters)
}

func (s *tuningStats) pctOfTicks(n int) float64 {
	if s.ticks == 0 {
		return 0
	}
	return float64(n) * 100 / float64(s.ticks)
}

func meanOf(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

func medianOf(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	c := append([]float64(nil), xs...)
	sort.Float64s(c)
	return c[len(c)/2]
}

func maxOf(xs []float64) float64 {
	m := 0.0
	for _, x := range xs {
		if x > m {
			m = x
		}
	}
	return m
}

func meanOfInts(xs []int) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0
	for _, x := range xs {
		sum += x
	}
	return float64(sum) / float64(len(xs))
}

// --- the driver -------------------------------------------------------------

// expDriver models one expedition category's launch/resolve loop.
type expDriver struct {
	duration   int
	difficulty float64
	gap        int
	ticksLeft  int
	cooldown   int
	category   string
	live       bool
}

// newExpDriver picks the SHORTEST expedition available in a category at an age —
// the spam-optimal choice, and therefore the honest worst case for encounter
// cadence. Returns live=false when the category has nothing available or the
// scenario disabled it (gap < 0).
func newExpDriver(ge *GameEngine, category, age string, gap int) expDriver {
	if gap < 0 {
		return expDriver{category: category}
	}
	defs := ge.Military.GetAvailableExpeditionsByCategory(category, age, ge.progress.GetAgeOrder())
	if len(defs) == 0 {
		return expDriver{category: category}
	}
	best := defs[0]
	for _, d := range defs[1:] {
		if d.Duration < best.Duration {
			best = d
		}
	}
	return expDriver{
		duration:   best.Duration,
		difficulty: best.DifficultyBase,
		gap:        gap,
		ticksLeft:  best.Duration,
		category:   category,
		live:       true,
	}
}

// step advances one tick. resolved reports whether the expedition finished this
// tick; success carries its outcome.
func (d *expDriver) step(rng *rand.Rand) (resolved, success bool) {
	if !d.live {
		return false, false
	}
	if d.ticksLeft > 0 {
		d.ticksLeft--
		if d.ticksLeft > 0 {
			return false, false
		}
		// Same success rule as MilitaryManager.tickCategory with no military
		// bonus: successRoll > DifficultyBase (floored at 0.05).
		diff := math.Max(d.difficulty, 0.05)
		d.cooldown = d.gap
		if d.cooldown == 0 {
			d.ticksLeft = d.duration
		}
		return true, rng.Float64() > diff
	}
	d.cooldown--
	if d.cooldown <= 0 {
		d.ticksLeft = d.duration
	}
	return false, false
}

// factionBoonPools sums the additive pools contributed by live faction BOON
// events only (setbacks are measured separately). Returns 1+Σ for production_all
// and the largest 1+Σ over the per-resource "<res>_rate" pools.
func factionBoonPools(ge *GameEngine) (prodAll, bestRate float64) {
	prodAll, bestRate = 1.0, 1.0
	rates := map[string]float64{}
	for _, ae := range ge.Events.active {
		if !strings.HasPrefix(ae.Key, factionBuffKeyPrefix) {
			continue
		}
		for _, eff := range ae.Effects {
			switch {
			case eff.Type == "production_all":
				prodAll += eff.Value
			case strings.HasSuffix(eff.Type, "_rate"):
				rates[eff.Type] += eff.Value
			}
		}
	}
	for _, v := range rates {
		if 1+v > bestRate {
			bestRate = 1 + v
		}
	}
	return prodAll, bestRate
}

// prepareTuningEngine builds an engine parked at a scenario's age with its
// standings installed, stockpiles filled, a worker pool seeded, and random
// events suppressed. Caller must hold no lock; the returned engine is left
// UNLOCKED.
func prepareTuningEngine(sc tuningScenario, seed int64) *GameEngine {
	ge := NewGameEngine()
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ge.SeedRNG(seed)
	ge.age = sc.age
	ge.currentEpoch = config.EpochForAge(ge.age)
	ge.Events.nextEventTick = 1 << 40 // suppress random events: active list is faction-only

	for key, st := range sc.standings {
		ge.Diplomacy.factions[key] = &FactionState{
			Discovered: true,
			Opinion:    100,
			Status:     st.status,
			AtWar:      st.atWar,
		}
	}

	for _, def := range config.BaseResources() {
		ge.Resources.UnlockResource(def.Key)
		if r, ok := ge.Resources.resources[def.Key]; ok {
			r.Storage = tuningStockFill * 10
			r.Amount = tuningStockFill
			r.Rate = 0
		}
	}
	ge.Workers.AddLentWorkers(tuningStartPop)

	return ge
}

// runTuningScenario drives one scenario and returns its measured stats.
func runTuningScenario(t *testing.T, sc tuningScenario, ticks int, seed int64) *tuningStats {
	t.Helper()

	boonSigs := boonFlavorSignatures(boon.Catalog())
	malusSigs := boonFlavorSignatures(boon.MalusCatalog())

	ge := prepareTuningEngine(sc, seed)
	stats := newTuningStats(sc.name)
	stats.ticks = ticks

	expRNG := rand.New(rand.NewSource(seed ^ 0x5EED))

	ge.mu.Lock()
	defer ge.mu.Unlock()

	order := ge.progress.GetAgeOrder()
	drivers := []*expDriver{}
	if d := newExpDriver(ge, ExpeditionScouting, sc.age, sc.scoutGap); d.live {
		drivers = append(drivers, &d)
	}
	if d := newExpDriver(ge, ExpeditionMilitary, sc.age, sc.milGap); d.live {
		drivers = append(drivers, &d)
	}

	resSnapshot := make(map[string]float64, len(config.BaseResources()))

	for tick := 0; tick < ticks; tick++ {
		for _, d := range drivers {
			resolved, success := d.step(expRNG)
			if !resolved {
				continue
			}
			stats.resolutions++

			// Snapshot everything a setback could touch, so severity is a real
			// measured delta rather than a re-roll.
			popBefore := ge.Workers.TotalPop()
			for _, def := range config.BaseResources() {
				resSnapshot[def.Key] = ge.Resources.Get(def.Key)
			}
			activeBefore := len(ge.Events.active)

			lines := ge.rollExpeditionEncounter(d.category, success)
			if len(lines) == 0 {
				continue
			}

			var sawOutcome bool
			for _, line := range lines {
				bucket, name := classifyLine(line, boonSigs, malusSigs)
				switch bucket {
				case "discovery":
					stats.discoveries++
					sawOutcome = true
				case "boon":
					stats.boons++
					stats.kindCounts[name]++
					sawOutcome = true
				case "bounce":
					stats.bounces++
					sawOutcome = true
				case "setback":
					stats.setbacks++
					stats.malusCounts[name]++
					sawOutcome = true
					if lost := popBefore - ge.Workers.TotalPop(); lost > 0 {
						stats.workersLost += lost
					}
					for _, def := range config.BaseResources() {
						before := resSnapshot[def.Key]
						now := ge.Resources.Get(def.Key)
						if before > 0 && now < before {
							stats.drainFracs = append(stats.drainFracs, (before-now)/before)
						}
					}
					for _, ae := range ge.Events.active[activeBefore:] {
						if !strings.HasPrefix(ae.Key, factionMalusKeyPrefix) {
							continue
						}
						for _, eff := range ae.Effects {
							if eff.Value < 0 {
								stats.negMags = append(stats.negMags, math.Abs(eff.Value))
								stats.negDurs = append(stats.negDurs, ae.TicksLeft)
							}
						}
					}
				default:
					stats.unclassified++
				}
			}
			if sawOutcome {
				stats.encounters++
			}
		}

		ge.Events.Tick(tick, ge.age, order, ge.currentEpoch)

		if n := ge.activeFactionBoonCount(); n > 0 {
			stats.ticksAnyBoon++
			if n >= maxConcurrentFactionBoons {
				stats.ticksFullCap++
			}
			if tick%tuningSampleEvery == 0 {
				prodAll, bestRate := factionBoonPools(ge)
				stats.prodAll = append(stats.prodAll, prodAll)
				stats.bestRate = append(stats.bestRate, bestRate)
				stats.combined = append(stats.combined, prodAll*bestRate)
			}
		}
	}

	return stats
}

// --- reporting --------------------------------------------------------------

func reportTuningStats(t *testing.T, s *tuningStats) {
	t.Helper()

	var b strings.Builder
	fmt.Fprintf(&b, "\n=== %s ===\n", s.name)
	fmt.Fprintf(&b, "  ticks=%d  expedition resolutions=%.1f/10k  encounters=%.1f/10k (%d total, %d discoveries)\n",
		s.ticks, s.per10k(s.resolutions), s.per10k(s.encounters), s.encounters, s.discoveries)
	fmt.Fprintf(&b, "  OUTCOME SPLIT: boon %5.1f%%  |  bounce %5.1f%%  |  setback %5.1f%%  (unclassified %d)\n",
		s.pctOfEncounters(s.boons), s.pctOfEncounters(s.bounces), s.pctOfEncounters(s.setbacks), s.unclassified)
	fmt.Fprintf(&b, "  TIME:  >=1 boon active %5.1f%%  |  at FULL capacity %5.1f%%\n",
		s.pctOfTicks(s.ticksAnyBoon), s.pctOfTicks(s.ticksFullCap))

	fmt.Fprintf(&b, "  BOON KINDS (of %d granted):\n", s.boons)
	for _, d := range boon.Catalog() {
		n := s.kindCounts[d.Name]
		flag := ""
		if s.boons > 0 && float64(n)/float64(s.boons) < 0.01 {
			flag = "   <-- effectively never fires"
		}
		fmt.Fprintf(&b, "    %-20s %6d  %5.1f%%%s\n", d.Name, n,
			float64(n)*100/math.Max(float64(s.boons), 1), flag)
	}

	fmt.Fprintf(&b, "  UPLIFT while boons live (%d samples):\n", len(s.prodAll))
	fmt.Fprintf(&b, "    production_all pool  mean x%.3f  median x%.3f  max x%.3f\n",
		meanOf(s.prodAll), medianOf(s.prodAll), maxOf(s.prodAll))
	fmt.Fprintf(&b, "    best <res>_rate pool mean x%.3f  median x%.3f  max x%.3f\n",
		meanOf(s.bestRate), medianOf(s.bestRate), maxOf(s.bestRate))
	fmt.Fprintf(&b, "    combined (best res)  mean x%.3f  median x%.3f  max x%.3f\n",
		meanOf(s.combined), medianOf(s.combined), maxOf(s.combined))

	fmt.Fprintf(&b, "  SETBACKS: %.1f/10k ticks (%d total)\n", s.per10k(s.setbacks), s.setbacks)
	if s.setbacks > 0 {
		fmt.Fprintf(&b, "    workers lost total=%d  (mean %.2f per setback)\n",
			s.workersLost, float64(s.workersLost)/float64(s.setbacks))
		fmt.Fprintf(&b, "    stockpile drains: n=%d  mean %.1f%%  median %.1f%%  max %.1f%%\n",
			len(s.drainFracs), meanOf(s.drainFracs)*100, medianOf(s.drainFracs)*100, maxOf(s.drainFracs)*100)
		fmt.Fprintf(&b, "    timed dips: n=%d  mean %.1f%%  max %.1f%%  mean duration %.0f ticks\n",
			len(s.negMags), meanOf(s.negMags)*100, maxOf(s.negMags)*100, meanOfInts(s.negDurs))
		for _, d := range boon.MalusCatalog() {
			fmt.Fprintf(&b, "    %-20s %6d  %5.1f%%\n", d.Name, s.malusCounts[d.Name],
				float64(s.malusCounts[d.Name])*100/float64(s.setbacks))
		}
	}
	t.Log(b.String())
}

// --- the tests ---------------------------------------------------------------

// TestBoonTuning_Scenarios is the instrument. It reports the distribution tables
// the tuning pass is judged against and asserts only that the harness actually
// exercised the system and that the outcome buckets are self-consistent.
func TestBoonTuning_Scenarios(t *testing.T) {
	ticks := tuningTicks
	if testing.Short() {
		ticks = tuningShortTicks
	}

	var summary strings.Builder
	fmt.Fprintf(&summary, "\n%-46s %9s %8s %8s %8s %8s %8s %9s\n",
		"SCENARIO", "enc/10k", "boon%", "bounce%", "setbk%", "anyboon%", "fullcap%", "uplift~x")

	for i, sc := range tuningScenarios() {
		stats := runTuningScenario(t, sc, ticks, 0xB00_1234+int64(i))
		reportTuningStats(t, stats)

		fmt.Fprintf(&summary, "%-46s %9.1f %8.1f %8.1f %8.1f %8.1f %8.1f %9.2f\n",
			sc.name,
			stats.per10k(stats.encounters),
			stats.pctOfEncounters(stats.boons),
			stats.pctOfEncounters(stats.bounces),
			stats.pctOfEncounters(stats.setbacks),
			stats.pctOfTicks(stats.ticksAnyBoon),
			stats.pctOfTicks(stats.ticksFullCap),
			medianOf(stats.combined))

		// Loose sanity only — this is an instrument, not a gate.
		if stats.encounters == 0 {
			t.Errorf("%s: no encounters fired at all — the driver is inert", sc.name)
		}
		if stats.unclassified > 0 {
			t.Errorf("%s: %d encounter lines could not be classified — the flavour "+
				"signature map is stale", sc.name, stats.unclassified)
		}
		if got := stats.boons + stats.bounces + stats.setbacks; got < stats.encounters {
			t.Errorf("%s: outcome buckets (%d) do not cover encounters (%d)",
				sc.name, got, stats.encounters)
		}
		if p := stats.pctOfEncounters(stats.boons); p < 0 || p > 100 {
			t.Errorf("%s: nonsensical boon share %.1f%%", sc.name, p)
		}
	}

	t.Log(summary.String())
}

// TestBoonTuning_HarnessClassifies is a fast guard on the measurement machinery
// itself: every catalog flavour template must yield a signature that maps back to
// its own entry, so a tuning run can never mis-bucket an outcome.
func TestBoonTuning_HarnessClassifies(t *testing.T) {
	boonSigs := boonFlavorSignatures(boon.Catalog())
	malusSigs := boonFlavorSignatures(boon.MalusCatalog())

	if len(boonSigs) < len(boon.Catalog()) {
		t.Fatalf("built %d boon signatures for %d catalog entries — a template has no "+
			"literal run of 8+ chars", len(boonSigs), len(boon.Catalog()))
	}
	if len(malusSigs) < len(boon.MalusCatalog()) {
		t.Fatalf("built %d malus signatures for %d entries", len(malusSigs), len(boon.MalusCatalog()))
	}

	check := func(defs []boon.Def, sigsA, sigsB map[string]string, want string) {
		for _, d := range defs {
			for _, tmpl := range d.Flavors {
				bucket, name := classifyLine(tmpl, sigsA, sigsB)
				if bucket != want || name != d.Name {
					t.Errorf("template %q classified as (%s,%s), want (%s,%s)",
						tmpl, bucket, name, want, d.Name)
				}
			}
		}
	}
	check(boon.Catalog(), boonSigs, malusSigs, "boon")
	check(boon.MalusCatalog(), boonSigs, malusSigs, "setback")

	for _, flavor := range atCapacityFlavors {
		if bucket, _ := classifyLine(flavor, boonSigs, malusSigs); bucket != "bounce" {
			t.Errorf("at-capacity flavour classified as %q, want bounce", bucket)
		}
	}
}
