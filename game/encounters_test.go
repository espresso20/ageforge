package game

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"testing"

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

// TestEncounter_SpecialtyBuffAppliesAndExpires proves the proof buff rides the
// existing timed-modifier machinery: it lands in the resolver's "<specialty>_rate"
// additive pool as a percentage, then is removed once its rolled duration elapses.
func TestEncounter_SpecialtyBuffAppliesAndExpires(t *testing.T) {
	ge := NewGameEngine()
	setAge(ge, "medieval_age")
	ge.mu.Lock()
	ge.SeedRNG(999)

	def := config.FactionByKey()["riverlands_tribes"] // specialty: food
	ge.Diplomacy.DiscoverFaction(def.Key)

	buff, ok := ge.rollFactionBuff(def)
	if !ok {
		ge.mu.Unlock()
		t.Fatal("rollFactionBuff returned false for a faction with a specialty")
	}
	eff := buff.event.Effects[0]
	if eff.Type != "food_rate" {
		ge.mu.Unlock()
		t.Fatalf("buff effect Type = %q, want food_rate (rides the <resource>_rate pool)", eff.Type)
	}
	pct := eff.Value
	dur := buff.event.TicksLeft
	if pct < factionBuffPctMin || pct > factionBuffPctMax {
		ge.mu.Unlock()
		t.Errorf("rolled magnitude %.4f outside [%.2f, %.2f]", pct, factionBuffPctMin, factionBuffPctMax)
	}
	if dur < factionBuffTicksMin || dur > factionBuffTicksMax {
		ge.mu.Unlock()
		t.Errorf("rolled duration %d outside [%d, %d]", dur, factionBuffTicksMin, factionBuffTicksMax)
	}

	// APPLY: inject the boon; the resolver's food_rate additive pool now carries it.
	ge.Events.InjectEvent(buff.event)
	if got := ge.buildResolver().AddTotal("food_rate"); math.Abs(got-pct) > 1e-9 {
		ge.mu.Unlock()
		t.Fatalf("buff not applied: food_rate pool = %.6f, want %.6f", got, pct)
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
		t.Errorf("buff expired early: food_rate pool = %.6f one tick before duration elapsed (want %.6f)", stillActive, pct)
	}
	if afterExpiry != 0 {
		t.Errorf("buff did not expire: food_rate pool still %.6f after %d ticks", afterExpiry, dur)
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
