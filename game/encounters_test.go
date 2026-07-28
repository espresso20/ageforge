package game

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/boon"
	"github.com/espresso20/ageforge/config"
)

// containsStr reports whether s is in xs.
func containsStr(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// encounterSignature captures everything a run of encounters produced — the
// discovered roster, the active timed boons (key/duration/effect), and the log
// lines — into a single comparable string. Two runs with the same seed and the
// same expedition sequence must produce byte-identical signatures.
func encounterSignature(ge *GameEngine, log []string) string {
	var b strings.Builder

	var disc []string
	for _, def := range config.BaseFactions() {
		if ge.Diplomacy.IsDiscovered(def.Key) {
			disc = append(disc, def.Key)
		}
	}
	sort.Strings(disc)
	b.WriteString("disc:")
	b.WriteString(strings.Join(disc, ","))

	b.WriteString("|events:")
	for _, ae := range ge.Events.GetActiveForSave() {
		fmt.Fprintf(&b, "[%s@%d", ae.Key, ae.TicksLeft)
		for _, eff := range ae.Effects {
			fmt.Fprintf(&b, " %s=%.6f", eff.Type, eff.Value)
		}
		b.WriteString("]")
	}

	b.WriteString("|log:")
	b.WriteString(strings.Join(log, "¦"))
	return b.String()
}

// encounterSeq is a fixed sequence of expedition resolutions exercising both
// categories and both outcomes.
var encounterSeq = []struct {
	cat     string
	success bool
}{
	{ExpeditionScouting, true}, {ExpeditionScouting, true}, {ExpeditionMilitary, true},
	{ExpeditionScouting, false}, {ExpeditionMilitary, false}, {ExpeditionScouting, true},
	{ExpeditionScouting, true}, {ExpeditionMilitary, true}, {ExpeditionScouting, true},
	{ExpeditionScouting, false}, {ExpeditionScouting, true}, {ExpeditionMilitary, true},
	{ExpeditionScouting, true}, {ExpeditionScouting, true}, {ExpeditionScouting, true},
	{ExpeditionMilitary, true}, {ExpeditionScouting, true}, {ExpeditionScouting, false},
	{ExpeditionScouting, true}, {ExpeditionScouting, true}, {ExpeditionMilitary, true},
	{ExpeditionScouting, true}, {ExpeditionScouting, true}, {ExpeditionScouting, true},
}

// TestSeededEncounter_Deterministic is the determinism guard: the same seed + the
// same expedition sequence yield identical encounter/buff outcomes across two
// independent engines.
func TestSeededEncounter_Deterministic(t *testing.T) {
	const seed int64 = 0x00A6EF0_1234_5678 // fixed, arbitrary

	run := func() string {
		ge := NewGameEngine()
		setAge(ge, "space_age") // many factions eligible → rich encounters
		ge.mu.Lock()
		defer ge.mu.Unlock()
		ge.SeedRNG(seed)
		var log []string
		for _, s := range encounterSeq {
			log = append(log, ge.rollExpeditionEncounter(s.cat, s.success)...)
		}
		return encounterSignature(ge, log)
	}

	a := run()
	b := run()
	if a != b {
		t.Fatalf("seeded encounter outcomes diverged:\n a = %s\n b = %s", a, b)
	}
	// Sanity: the run must actually have produced encounters, else determinism is
	// vacuously true.
	if !strings.Contains(a, "faction_boon_") {
		t.Fatalf("determinism run produced no faction boons — signature is trivial:\n%s", a)
	}
}

// TestEncounter_DiscoversEligibleFaction confirms the primary discovery path:
// running scouting expeditions eventually turns up an eligible (age-floor-met)
// faction that reaching its MinAge alone did NOT auto-discover.
func TestEncounter_DiscoversEligibleFaction(t *testing.T) {
	ge := NewGameEngine()
	setAge(ge, "bronze_age") // only riverlands_tribes (MinAge bronze) is eligible
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.SeedRNG(12345)

	// Floor, not trigger: no expedition has run, so nobody is discovered yet.
	if ge.Diplomacy.IsDiscovered("riverlands_tribes") {
		t.Fatal("faction discovered merely on reaching its MinAge — floor should not auto-discover")
	}

	found := false
	for i := 0; i < 200 && !found; i++ {
		ge.rollExpeditionEncounter(ExpeditionScouting, true)
		found = ge.Diplomacy.IsDiscovered("riverlands_tribes")
	}
	if !found {
		t.Fatal("running scouting expeditions never discovered the eligible faction")
	}
}

// TestDiscoverFactions_AgeFallbackAndFloor verifies the anti-softlock fallback:
// reaching MinAge (or one age past) does NOT auto-discover, but ageFallbackGap ages
// past does — with no expeditions run at all.
func TestDiscoverFactions_AgeFallbackAndFloor(t *testing.T) {
	order := fullAgeOrder()

	// At exactly the floor: no fallback.
	if got := NewDiplomacyManager().DiscoverFactions("bronze_age", order); len(got) != 0 {
		t.Fatalf("DiscoverFactions at MinAge discovered %v; want none (age is a floor)", got)
	}
	// One age past (< gap): still no fallback.
	if got := NewDiplomacyManager().DiscoverFactions("iron_age", order); len(got) != 0 {
		t.Fatalf("DiscoverFactions one age past MinAge discovered %v; want none (gap is %d)", got, ageFallbackGap)
	}
	// ageFallbackGap ages past MinAge: the fallback fires with no expeditions.
	dm := NewDiplomacyManager()
	got := dm.DiscoverFactions("classical_age", order) // bronze + 2
	if !containsStr(got, "riverlands_tribes") {
		t.Fatalf("age fallback did not auto-discover the long-eligible founding civ; got %v", got)
	}
	fs := dm.factions["riverlands_tribes"]
	if fs == nil || !fs.Discovered || fs.Opinion != 0 || fs.Status != "neutral" {
		t.Errorf("fallback discovery did not seed neutral/0 state: %+v", fs)
	}
	// A civ still below its own floor is untouched.
	if dm.IsDiscovered("ironhold_clans") { // MinAge medieval, above classical
		t.Error("ironhold_clans discovered before reaching its MinAge floor")
	}
}

// TestBoonApplier_InjectsAndExpires proves an encounter's TIMED boon still rides
// the existing timed-modifier machinery when routed through the game-side
// boon.Applier: a RateBuff lands in the resolver's "<resource>_rate" additive pool
// as a percentage, then is removed once its duration elapses. (Phase 1 rolled this
// buff inline; Phase 2b constructs it in the boon engine and applies it here.)
func TestBoonApplier_InjectsAndExpires(t *testing.T) {
	ge := NewGameEngine()
	setAge(ge, "medieval_age")
	ge.mu.Lock()

	const pct = 0.13
	const dur = 4000
	// A concrete timed boon, as the boon engine would hand it to Apply.
	b := boon.Boon{
		Kind:          boon.RateBuff,
		Resource:      "food",
		Magnitude:     pct,
		DurationTicks: dur,
		Name:          "Specialty Windfall",
		Flavor:        "Their artisans share a trade secret — +13% food.",
	}
	def := config.FactionByKey()["riverlands_tribes"] // food specialist
	applier := boonApplier{ge: ge, name: def.Name, key: def.Key}

	// APPLY through the real game Applier: the resolver's food_rate pool now carries it.
	line := boon.Apply(b, applier)
	if line == "" {
		ge.mu.Unlock()
		t.Fatal("boon.Apply returned an empty flavor line")
	}
	if got := ge.buildResolver().AddTotal("food_rate"); math.Abs(got-pct) > 1e-9 {
		ge.mu.Unlock()
		t.Fatalf("boon not applied: food_rate pool = %.6f, want %.6f", got, pct)
	}

	// Isolate expiry from random events so only the boon drives the pool.
	ge.Events.nextEventTick = 1 << 40
	order := ge.progress.GetAgeOrder()

	// One tick before expiry it is still active; the duration-th tick removes it.
	for i := 0; i < dur-1; i++ {
		ge.Events.Tick(i, "medieval_age", order, ge.currentEpoch)
	}
	stillActive := ge.buildResolver().AddTotal("food_rate")
	ge.Events.Tick(dur, "medieval_age", order, ge.currentEpoch)
	afterExpiry := ge.buildResolver().AddTotal("food_rate")
	ge.mu.Unlock()

	if math.Abs(stillActive-pct) > 1e-9 {
		t.Errorf("boon expired early: food_rate pool = %.6f one tick before duration elapsed (want %.6f)", stillActive, pct)
	}
	if afterExpiry != 0 {
		t.Errorf("boon did not expire: food_rate pool still %.6f after %d ticks", afterExpiry, dur)
	}
}

// neutralState / statusState are FactionState fixtures for profile tests.
func statusState(status string) FactionState {
	return FactionState{Discovered: true, Status: status}
}

// TestFactionProfile_Derivation checks that a faction's DATA + standing map onto
// the boon-profile knobs the way factionProfile promises: personality leans the
// per-kind weights, strength scales magnitude, allied standing lifts magnitude AND
// rarity, and an at-war faction yields no positive boon.
func TestFactionProfile_Derivation(t *testing.T) {
	const age = "space_age"

	peaceful := config.FactionByKey()["riverlands_tribes"]      // peaceful, Str 1
	aggressive := config.FactionByKey()["ironhold_clans"]       // aggressive, Str 3
	isolationist := config.FactionByKey()["stellar_federation"] // isolationist, Str 4

	pPeace := factionProfile(peaceful, statusState("neutral"), age)
	pAggr := factionProfile(aggressive, statusState("neutral"), age)

	// Specialty windfall (RateBuff) is always favoured.
	if pPeace.WeightMult[boon.RateBuff] <= 1.0 || pAggr.WeightMult[boon.RateBuff] <= 1.0 {
		t.Errorf("RateBuff not up-weighted for both: peaceful=%.2f aggressive=%.2f",
			pPeace.WeightMult[boon.RateBuff], pAggr.WeightMult[boon.RateBuff])
	}
	// Peaceful up-weights the gentle kinds (RateBuff + AllProduction).
	if pPeace.WeightMult[boon.AllProduction] <= 1.0 {
		t.Errorf("peaceful did not up-weight AllProduction: %.2f", pPeace.WeightMult[boon.AllProduction])
	}
	// Aggressive up-weights TickSpeed + InstantResource ("spoils").
	if pAggr.WeightMult[boon.TickSpeed] <= 1.0 || pAggr.WeightMult[boon.InstantResource] <= 1.0 {
		t.Errorf("aggressive did not up-weight spoils: TickSpeed=%.2f InstantResource=%.2f",
			pAggr.WeightMult[boon.TickSpeed], pAggr.WeightMult[boon.InstantResource])
	}
	// Contrast: aggressive favours TickSpeed far more than peaceful does; peaceful
	// favours the gentle broad buff far more than aggressive does.
	if pAggr.WeightMult[boon.TickSpeed] <= pPeace.WeightMult[boon.TickSpeed] {
		t.Errorf("aggressive TickSpeed weight %.2f not > peaceful %.2f",
			pAggr.WeightMult[boon.TickSpeed], pPeace.WeightMult[boon.TickSpeed])
	}
	if pPeace.WeightMult[boon.AllProduction] <= pAggr.WeightMult[boon.AllProduction] {
		t.Errorf("peaceful AllProduction weight %.2f not > aggressive %.2f",
			pPeace.WeightMult[boon.AllProduction], pAggr.WeightMult[boon.AllProduction])
	}

	// Isolationist raises RarityScale so the rare tier surfaces.
	pIso := factionProfile(isolationist, statusState("neutral"), age)
	if pIso.RarityScale <= pPeace.RarityScale {
		t.Errorf("isolationist RarityScale %.2f not > a non-isolationist's %.2f",
			pIso.RarityScale, pPeace.RarityScale)
	}

	// Standing: allied raises BOTH MagnitudeScale and RarityScale vs neutral.
	neutral := factionProfile(peaceful, statusState("neutral"), age)
	allied := factionProfile(peaceful, statusState("allied"), age)
	if allied.MagnitudeScale <= neutral.MagnitudeScale {
		t.Errorf("allied MagnitudeScale %.3f not > neutral %.3f", allied.MagnitudeScale, neutral.MagnitudeScale)
	}
	if allied.RarityScale <= neutral.RarityScale {
		t.Errorf("allied RarityScale %.3f not > neutral %.3f", allied.RarityScale, neutral.RarityScale)
	}

	// Strength: a Str-5 civ gifts harder than a Str-1 civ (same personality/standing).
	str1 := factionProfile(config.FactionDef{Personality: "peaceful", Specialty: "food", Strength: 1}, statusState("neutral"), age)
	str5 := factionProfile(config.FactionDef{Personality: "peaceful", Specialty: "food", Strength: 5}, statusState("neutral"), age)
	if str5.MagnitudeScale <= str1.MagnitudeScale {
		t.Errorf("Str-5 MagnitudeScale %.3f not > Str-1 %.3f", str5.MagnitudeScale, str1.MagnitudeScale)
	}

	// At war: the profile disables every kind, so RollBoon yields the zero Boon.
	war := factionProfile(peaceful, FactionState{Discovered: true, Status: "neutral", AtWar: true}, age)
	if b := boon.RollBoon(war, rand.New(rand.NewSource(1))); b != (boon.Boon{}) {
		t.Errorf("at-war faction yielded a positive boon: %+v", b)
	}
}

// captureApplier records the boon side effects it is handed, so a test can assert
// exactly one effect landed without touching real game state.
type captureApplier struct {
	injects  int
	grants   int
	loans    int
	drains   int
	kills    int
	lastName string
}

func (c *captureApplier) InjectTimedEffects(_ []config.Effect, _ int, name string) {
	c.injects++
	c.lastName = name
}
func (c *captureApplier) GrantResource(_ string, _ float64) { c.grants++ }
func (c *captureApplier) GrantTempWorkers(_, _ int)         { c.loans++ }
func (c *captureApplier) DrainResource(_ string, _ float64) { c.drains++ }
func (c *captureApplier) LoseWorkers(_ int)                 { c.kills++ }

// TestFactionEncounter_RollsAppliesOneEffect_Deterministic drives the real
// encounter roll (factionProfile → boon.RollBoon → boon.Apply) against a capturing
// Applier: a discovered faction produces exactly ONE applied effect, and the same
// seed reproduces the same boon.
func TestFactionEncounter_RollsAppliesOneEffect_Deterministic(t *testing.T) {
	def := config.FactionByKey()["merchant_guild"] // mercantile, Str 2

	roll := func(seed int64) (boon.Boon, *captureApplier) {
		ge := NewGameEngine()
		setAge(ge, "space_age")
		ge.mu.Lock()
		defer ge.mu.Unlock()
		ge.SeedRNG(seed)
		ge.Diplomacy.DiscoverFaction(def.Key)
		state, _ := ge.Diplomacy.StateOf(def.Key)
		prof := factionProfile(def, state, ge.age)
		b := boon.RollBoon(prof, ge.rng)
		cap := &captureApplier{}
		boon.Apply(b, cap)
		return b, cap
	}

	b1, cap1 := roll(0x5EED)
	b2, cap2 := roll(0x5EED)

	if b1 == (boon.Boon{}) {
		t.Fatal("a discovered, non-hostile faction rolled the zero boon")
	}
	if b1 != b2 {
		t.Fatalf("same seed produced different boons:\n b1=%+v\n b2=%+v", b1, b2)
	}
	// Exactly one Applier effect landed (each kind maps to one call).
	total := cap1.injects + cap1.grants + cap1.loans
	if total != 1 {
		t.Fatalf("expected exactly one applied effect, got %d (injects=%d grants=%d loans=%d)",
			total, cap1.injects, cap1.grants, cap1.loans)
	}
	if (cap1.injects != cap2.injects) || (cap1.grants != cap2.grants) || (cap1.loans != cap2.loans) {
		t.Errorf("effect dispatch diverged across identical seeds: %+v vs %+v", cap1, cap2)
	}
}

// TestSeedPersists_RoundTrip confirms the run's master seed survives a save/load so
// the run stays reproducible from its start.
func TestSeedPersists_RoundTrip(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	ge.SeedRNG(0x1234_5678_9abc)
	want := ge.seed
	ge.mu.Unlock()

	if err := ge.SaveGame("test_seed_roundtrip"); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}
	defer os.Remove("data/saves/test_seed_roundtrip.json")

	ge2 := NewGameEngine()
	if err := ge2.LoadGame("test_seed_roundtrip"); err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}
	if got := ge2.Seed(); got != want {
		t.Errorf("seed did not round-trip: got %d, want %d", got, want)
	}
}
