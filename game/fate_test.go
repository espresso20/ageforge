package game

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
)

// --- helpers ------------------------------------------------------------------

// fateEngine returns a seeded engine standing in age, with every age's unlocks
// up to it and the era's fate rolled, as the first tick in it would.
func fateEngine(t *testing.T, age string, seed int64) *GameEngine {
	t.Helper()
	ge := catEngine(t, age, seed)
	for _, a := range config.AgeOrder() {
		ge.applyAgeUnlocks(a)
		if a == age {
			break
		}
	}
	ge.harbingerTickCheck()
	return ge
}

// forceFate fates a doom in ge's era striking offset ticks after it began.
func forceFate(t *testing.T, ge *GameEngine, offset int) {
	t.Helper()
	if err := ge.ForceFateForTest(ge.currentEpoch, offset); err != nil {
		t.Fatal(err)
	}
}

// tickTo moves the clock to tick and runs the fate's tick hook there.
func tickTo(ge *GameEngine, tick int) {
	ge.tick = tick
	ge.harbingerTickCheck()
}

// arrivalTick is when ge's fated harbinger is due in the current age.
func arrivalTick(ge *GameEngine) int {
	return ge.fate.StrikeTick - int(ge.fate.LeadFrac*expectedAgeTicks(ge.age))
}

// readyToAdvance makes ge's next advance allowed without meeting the
// requirements (AdvanceAge trusts ageReady).
func readyToAdvance(ge *GameEngine) { ge.ageReady = true }

// --- the roll -------------------------------------------------------------------

// The roll is seeded, draws five values whatever it decides, and its outcomes
// have the documented shape: FateChance of eras fated, a strike tick across
// the era's expected length, a lead inside its bounds, false prophets only
// with nothing fated.
func TestFateRollIsSeededAndShaped(t *testing.T) {
	roll := func(age string, seed int64) FateSave {
		ge := catEngine(t, age, seed)
		ge.harbingerTickCheck()
		return *ge.fate
	}
	if a, b := roll("iron_age", 42), roll("iron_age", 42); a != b {
		t.Fatalf("same seed, different fates:\n%+v\n%+v", a, b)
	}

	// The shape, over many rolls of one seeded stream (one engine: building
	// thousands is slow under -race).
	window := int(math.Round(expectedEraTicks("iron_era")))
	const n = 6000
	fated, falses, offsets := 0, 0, 0.0
	ge := catEngine(t, "iron_age", 7)
	for seed := int64(1); seed <= n; seed++ {
		ge.rollFate()
		f := *ge.fate
		if f.EpochKey != "iron_era" || f.Window != window || f.EntryTick != 0 || f.Resolved != "" {
			t.Fatalf("roll %d: fate = %+v", seed, f)
		}
		switch {
		case f.Fated:
			fated++
			if f.FalseProphet {
				t.Fatalf("seed %d: fated and a false prophet", seed)
			}
			offsets += float64(f.StrikeTick - f.EntryTick)
		case f.FalseProphet:
			falses++
			if f.Claim != CatastropheTierMedium && f.Claim != CatastropheTierHigh {
				t.Fatalf("seed %d: false prophet claims %q", seed, f.Claim)
			}
		default:
			if f.StrikeTick != 0 || f.LeadFrac != 0 {
				t.Fatalf("seed %d: a quiet era with a strike: %+v", seed, f)
			}
		}
		if f.Fated || f.FalseProphet {
			if off := f.StrikeTick - f.EntryTick; off < 0 || off >= window {
				t.Fatalf("seed %d: strike offset %d outside [0, %d)", seed, off, window)
			}
			if f.LeadFrac < harbingerLeadMin || f.LeadFrac >= harbingerLeadMax {
				t.Fatalf("seed %d: lead %v outside [%v, %v)", seed, f.LeadFrac, harbingerLeadMin, harbingerLeadMax)
			}
		}
	}
	if got := float64(fated) / n; math.Abs(got-FateChance) > 0.02 {
		t.Errorf("fated rate %.3f, want ≈ %.2f", got, FateChance)
	}
	def, _ := config.HarbingerFor("iron_age")
	if got, want := float64(falses)/n, (1-FateChance)*def.FalseProphetChance; math.Abs(got-want) > 0.015 {
		t.Errorf("false prophet rate %.4f, want ≈ %.4f", got, want)
	}
	if mean := offsets / float64(fated); math.Abs(mean/float64(window)-0.5) > 0.03 {
		t.Errorf("mean strike offset is %.3f of the era, want ≈ 0.5 (uniform)", mean/float64(window))
	}

	// Five draws, whatever the outcome.
	for seed := int64(1); seed <= 50; seed++ {
		ge := catEngine(t, "iron_age", seed)
		before, _ := ge.rngDraws()
		ge.rollFate()
		if after, _ := ge.rngDraws(); after-before != 5 {
			t.Fatalf("seed %d: the roll drew %d values, want 5", seed, after-before)
		}
	}
}

// False prophets only before the Industrial Age, only with nothing fated; the
// Stone Era can only ever have a false prophet; the final epoch has no fate.
func TestFalseProphetsOnlyBeforeIndustrialWithNothingFated(t *testing.T) {
	count := func(age string, n int) (fated, falses int) {
		ge := catEngine(t, age, 3) // one seeded stream: engines are slow to build under -race
		for i := 0; i < n; i++ {
			ge.rollFate()
			if f := ge.fate; f.Fated {
				fated++
			} else if f.FalseProphet {
				falses++
			}
		}
		return fated, falses
	}
	if fated, falses := count("primitive_age", 3000); fated != 0 || math.Abs(float64(falses)/3000-0.125) > 0.02 {
		t.Errorf("stone era: %d fated, %d false prophets in 3000; want none fated, about 1 in 8 false", fated, falses)
	}
	if _, falses := count("renaissance_age", 3000); falses == 0 {
		t.Error("steel era: no false prophet in 3000 rolls")
	}
	for _, age := range []string{"victorian_age", "modern_age", "cyberpunk_age"} {
		if _, falses := count(age, 800); falses != 0 {
			t.Errorf("%s: %d false prophets, want none from the Industrial Age on", age, falses)
		}
	}
	ge := catEngine(t, "interstellar_age", 1)
	ge.harbingerTickCheck()
	if ge.fate != nil {
		t.Errorf("the final epoch rolled a fate: %+v", ge.fate)
	}
}

// Entering an era rolls its fate at the transition, after the epoch event.
func TestFateRolledOnEnteringAnEra(t *testing.T) {
	ge := fateEngine(t, "medieval_age", 4)
	ge.tick = 777
	var rolled []EventData
	ge.Bus.Subscribe(EventFateRolled, func(e EventData) { rolled = append(rolled, e) })
	ge.advanceAge("renaissance_age")
	if ge.fate == nil || ge.fate.EpochKey != "steel_era" || ge.fate.EntryTick != 777 {
		t.Fatalf("fate after entering the Steel Era = %+v", ge.fate)
	}
	if want := int(math.Round(expectedEraTicks("steel_era"))); ge.fate.Window != want {
		t.Errorf("window %d, want %d (the era's ages at their targets)", ge.fate.Window, want)
	}
	if len(rolled) != 1 || rolled[0].Payload["epoch_key"] != "steel_era" {
		t.Errorf("fate events = %+v", rolled)
	}
	if n := len(ge.epochEventHistory); n == 0 || ge.epochEventHistory[n-1].EpochKey != "steel_era" {
		t.Errorf("no epoch event for the transition: %+v", ge.epochEventHistory)
	}
}

// --- save and load ---------------------------------------------------------------

// The fate is persisted: a reload restores it exactly and never re-rolls it,
// and the stream carries on where it stopped.
func TestFateSurvivesSaveLoadWithoutReroll(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	for _, tc := range []struct {
		name  string
		setup func(ge *GameEngine)
	}{
		{"fated", func(ge *GameEngine) { forceFate(t, ge, 21000) }},
		{"quiet", func(ge *GameEngine) {
			if err := ge.ForceQuietFateForTest(ge.currentEpoch); err != nil {
				t.Fatal(err)
			}
		}},
		{"false prophet", func(ge *GameEngine) {
			if err := ge.ForceFalseProphetForTest(ge.currentEpoch, 21000); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ge := fateEngine(t, "classical_age", 6)
			tc.setup(ge)
			ge.tick = 100
			want := *ge.fate
			draws, _ := ge.rngDraws()
			if err := ge.SaveGame("fate_" + strings.ReplaceAll(tc.name, " ", "_")); err != nil {
				t.Fatal(err)
			}
			ge2 := NewGameEngine()
			if err := ge2.LoadGame("fate_" + strings.ReplaceAll(tc.name, " ", "_")); err != nil {
				t.Fatal(err)
			}
			if ge2.cheaterBadge {
				t.Error("a save with a fate failed its signature check")
			}
			if ge2.fate == nil || *ge2.fate != want {
				t.Fatalf("fate after load = %+v, want %+v", ge2.fate, want)
			}
			if d, _ := ge2.rngDraws(); d != draws {
				t.Errorf("stream position after load %d, want %d", d, draws)
			}
			ge2.harbingerTickCheck()
			if *ge2.fate != want {
				t.Errorf("the first tick after load re-rolled the fate: %+v", ge2.fate)
			}
		})
	}
}

// A save written before fates existed, standing mid-era with no thread, gets
// its era's fate on the first tick (entered then). One with a live thread from
// the old rules keeps it as the warning of this era's doom, fated to strike at
// the era's end at the latest; a Stone Era one is dropped.
func TestOldSavesMigrateToFates(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := catEngine(t, "classical_age", 1)
	ge.tick = 5000
	if err := ge.SaveGame("old_no_thread"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("old_no_thread"); err != nil {
		t.Fatal(err)
	}
	if ge2.fate != nil || ge2.harbinger != nil {
		t.Fatalf("load started something: fate %+v thread %+v", ge2.fate, ge2.harbinger)
	}
	ge2.harbingerTickCheck()
	if ge2.fate == nil || ge2.fate.EpochKey != "iron_era" || ge2.fate.EntryTick != 5000 {
		t.Errorf("first tick after an old load: fate %+v", ge2.fate)
	}

	// An old thread in the Iron Era warned of the transition into the Steel
	// Era; it becomes this era's doom.
	old := catEngine(t, "classical_age", 2)
	old.harbinger = &HarbingerSave{Age: "classical_age", Chain: []string{"iron_age", "classical_age"}, EpochKey: "iron_era",
		TargetEpoch: "steel_era", AnnouncedTier: CatastropheTierMedium, AppeaseLevel: 1, Invited: true}
	old.harbingerArrived["iron_era"] = true
	old.catastropheInvited = true
	snap := old.buildSaveSnapshot()
	snap.Fate = nil
	snap.Signature = signSave(snap, saveHMACKey)
	writeRawSave(t, "old_thread", snap)
	ge3 := NewGameEngine()
	if err := ge3.LoadGame("old_thread"); err != nil {
		t.Fatal(err)
	}
	h, f := ge3.harbinger, ge3.fate
	if h == nil || h.TargetEpoch != "iron_era" || h.AppeaseLevel != 1 || !h.Invited || h.When != WhenThisEra {
		t.Fatalf("migrated thread = %+v", h)
	}
	if f == nil || !f.Fated || !f.Invited || f.StrikeTick != math.MaxInt32 || ge3.catastropheInvited {
		t.Fatalf("migrated fate = %+v (invite flag %v)", f, ge3.catastropheInvited)
	}
	// It strikes at the era's end at the latest.
	ge3.age = "medieval_age"
	readyToAdvance(ge3)
	if err := ge3.AdvanceAge(); err == nil || ge3.pendingCatastrophe != "iron_era" {
		t.Errorf("leaving the era: err %v, pending %q; want the invited doom first", err, ge3.pendingCatastrophe)
	}

	// An old Stone Era thread has no doom to warn of now.
	stone := catEngine(t, "stone_age", 3)
	stone.harbinger = &HarbingerSave{Age: "stone_age", Chain: []string{"stone_age"}, EpochKey: "stone_era", TargetEpoch: "iron_era"}
	stone.harbingerArrived["stone_era"] = true
	snap = stone.buildSaveSnapshot()
	snap.Fate = nil
	snap.Signature = signSave(snap, saveHMACKey)
	writeRawSave(t, "old_stone", snap)
	ge4 := NewGameEngine()
	if err := ge4.LoadGame("old_stone"); err != nil {
		t.Fatal(err)
	}
	if ge4.harbinger != nil {
		t.Errorf("an old Stone Era thread survived: %+v", ge4.harbinger)
	}
}

// Saves with no fate write no fate key (a final-epoch save), and a quiet era's
// fate carries no strike.
func TestFateSaveFields(t *testing.T) {
	cosmic := catEngine(t, "galactic_age", 1)
	cosmic.harbingerTickCheck()
	raw, err := json.Marshal(cosmic.buildSaveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"fate"`) {
		t.Errorf("a final-epoch save carries a fate: %s", raw)
	}
	quiet := fateEngine(t, "iron_age", 1)
	if err := quiet.ForceQuietFateForTest("iron_era"); err != nil {
		t.Fatal(err)
	}
	raw, _ = json.Marshal(quiet.buildSaveSnapshot().Fate)
	for _, k := range []string{"fated", "strike_tick", "lead_frac", "false_prophet"} {
		if strings.Contains(string(raw), `"`+k+`"`) {
			t.Errorf("a quiet fate writes %q: %s", k, raw)
		}
	}
}

// --- no leak -----------------------------------------------------------------------

// Until its harbinger comes, a fated doom changes nothing the player can see:
// two engines alike but for the fate show the same snapshot tick after tick.
func TestNoFateLeaksBeforeTheHarbinger(t *testing.T) {
	mk := func(fated bool) *GameEngine {
		ge := fateEngine(t, "classical_age", 9)
		if fated {
			forceFate(t, ge, 40000)
		} else if err := ge.ForceQuietFateForTest("iron_era"); err != nil {
			t.Fatal(err)
		}
		return ge
	}
	a, b := mk(true), mk(false)
	norm := func(st GameState) GameState {
		st.Stats.GameStarted, st.Stats.PlayTime = time.Time{}, 0 // wall clock
		return st
	}
	for i := 0; i < 400; i++ {
		a.doTick()
		b.doTick()
		if i%50 != 0 {
			continue
		}
		sa, sb := norm(a.GetState()), norm(b.GetState())
		if sa.Harbinger != nil {
			t.Fatalf("tick %d: the harbinger came early", sa.Tick)
		}
		if !reflect.DeepEqual(sa, sb) {
			va, vb := reflect.ValueOf(sa), reflect.ValueOf(sb)
			for i := 0; i < va.NumField(); i++ {
				if !reflect.DeepEqual(va.Field(i).Interface(), vb.Field(i).Interface()) {
					t.Errorf("tick %d: GameState.%s differs between a fated era and a quiet one:\n%+v\n%+v",
						sa.Tick, va.Type().Field(i).Name, va.Field(i).Interface(), vb.Field(i).Interface())
				}
			}
			t.FailNow()
		}
	}
}

// --- arrival -------------------------------------------------------------------------

// The harbinger comes a lead before the strike (LeadFrac of the current age's
// target), never before the era began, with the current age's figure.
func TestHarbingerArrivesALeadBeforeTheStrike(t *testing.T) {
	ge := fateEngine(t, "classical_age", 3)
	forceFate(t, ge, 30000)
	arrive := arrivalTick(ge)
	tickTo(ge, arrive-1)
	if ge.harbinger != nil {
		t.Fatalf("harbinger at tick %d, due at %d", ge.tick, arrive)
	}
	tickTo(ge, arrive)
	h := ge.harbinger
	if h == nil || h.Age != "classical_age" || h.EpochKey != "iron_era" || h.TargetEpoch != "iron_era" || h.ArrivedTick != arrive || h.FalseProphet {
		t.Fatalf("thread at arrival = %+v", h)
	}
	if countLogs(ge, "The Oracle has come, warning of impending doom") != 1 {
		t.Errorf("no arrival line:\n%s", strings.Join(logMessages(ge), "\n"))
	}

	// A strike sooner than the lead: the harbinger comes at once, and the
	// strike waits the shortest lead after it (a doom is always foretold).
	early := fateEngine(t, "iron_age", 3)
	early.tick = 10
	forceFate(t, early, 5)
	tickTo(early, 10)
	if early.harbinger == nil || early.harbinger.ArrivedTick != 10 {
		t.Fatalf("early strike: thread %+v, want one arrived at the current tick", early.harbinger)
	}
	if want := 10 + int(harbingerLeadMin*expectedAgeTicks("iron_age")); early.fate.StrikeTick != want || early.pendingCatastrophe != "" {
		t.Errorf("early strike: fated for %d (pending %q), want held to %d", early.fate.StrikeTick, early.pendingCatastrophe, want)
	}

	// Entering a longer age can bring the lead past now: the new age's
	// figure comes at the advance. The Iron Age's lead (half of 4,500 ticks)
	// falls short; the Classical Age's (half of 6,300) reaches.
	adv := fateEngine(t, "iron_age", 5)
	forceFate(t, adv, int(expectedAgeTicks("iron_age"))+int(0.45*expectedAgeTicks("classical_age")))
	adv.fate.LeadFrac = 0.5
	adv.tick = int(expectedAgeTicks("iron_age")) - 10
	adv.harbingerTickCheck()
	if adv.harbinger != nil {
		t.Fatal("setup: the harbinger came in the Iron Age")
	}
	adv.advanceAge("classical_age")
	if adv.harbinger == nil || adv.harbinger.Age != "classical_age" {
		t.Errorf("after entering the Classical Age: thread %+v, want the Oracle", adv.harbinger)
	}
}

// --- timing forecasts --------------------------------------------------------------

// What a figure says about WHEN sharpens by era: nothing before the Classical
// Age; from it, "before this age is out" when the doom falls in the current
// age on the era's schedule (or in its last age), else "before the era ends".
// From the Industrial Age the medium also prints the odds.
func TestForecastTiersByEra(t *testing.T) {
	for _, tc := range []struct {
		age     string
		offset  func(ages []string) int
		want    string
		numeric bool
	}{
		{"iron_age", func([]string) int { return 100 }, WhenUntold, false},
		// The strike falls in the Classical Age's span of the schedule.
		{"classical_age", func(a []string) int { return int(expectedAgeTicks(a[0]) + 0.5*expectedAgeTicks(a[1])) }, WhenThisAge, false},
		// It falls in the Medieval Age's span, the Oracle can only say "this era".
		{"classical_age", func(a []string) int {
			return int(expectedAgeTicks(a[0]) + expectedAgeTicks(a[1]) + 0.9*expectedAgeTicks(a[2]))
		}, WhenThisEra, false},
		// The era's last age cannot be outlasted.
		{"medieval_age", func(a []string) int { return int(expectedEraTicks("iron_era")) - 1 }, WhenThisAge, false},
		{"renaissance_age", func(a []string) int { return int(expectedAgeTicks(a[0]) / 2) }, WhenThisAge, false},
		{"industrial_age", func(a []string) int { return int(expectedEraTicks("steel_era")) - 1 }, WhenThisAge, true},
		{"victorian_age", func(a []string) int { return int(expectedEraTicks("electric_era")) - 1 }, WhenThisEra, true},
	} {
		t.Run(tc.age+"_"+tc.want, func(t *testing.T) {
			ge := fateEngine(t, tc.age, 1)
			forceFate(t, ge, tc.offset(config.EpochByKey()[ge.currentEpoch].Ages))
			ge.fate.LeadFrac = harbingerLeadMin
			delete(ge.harbingerArrived, ge.currentEpoch)
			if !ge.fateArrive() {
				t.Fatal("no harbinger")
			}
			v := ge.GetState().Harbinger
			if v.When != tc.want || v.Numeric != tc.numeric {
				t.Errorf("%s: when %q numeric %v, want %q %v", tc.age, v.When, v.Numeric, tc.want, tc.numeric)
			}
			wantText := map[string]string{WhenUntold: "", WhenThisAge: "before this age is out",
				WhenThisEra: "before the " + config.EpochByKey()[ge.currentEpoch].Name + " ends"}[tc.want]
			if v.WhenText != wantText {
				t.Errorf("when text %q, want %q", v.WhenText, wantText)
			}
		})
	}
	// The figures' timing follows the roster: none before the Classical Age.
	for i, h := range config.Harbingers() {
		want := config.TimingNone
		if i >= 4 {
			want = config.TimingAge
		}
		if h.ForecastTiming != want {
			t.Errorf("%s: timing %v, want %v", h.Age, h.ForecastTiming, want)
		}
	}
}

// --- the strike ------------------------------------------------------------------------

// The doom strikes at its tick: pending, the thread vindicated, the Brace
// handed on, the fate settled there and then.
func TestFatedDoomStrikesAtItsTick(t *testing.T) {
	ge := fateEngine(t, "classical_age", 7)
	forceFate(t, ge, 26000)
	strike := ge.fate.StrikeTick
	tickTo(ge, arrivalTick(ge))
	if ge.harbinger == nil {
		t.Fatal("no harbinger")
	}
	ge.harbinger.BraceLevel = 1
	var cat []EventData
	ge.Bus.Subscribe(EventEpochEventFired, func(e EventData) { cat = append(cat, e) })
	tickTo(ge, strike-1)
	if ge.pendingCatastrophe != "" || ge.fate.Resolved != "" {
		t.Fatalf("struck early: pending %q fate %+v", ge.pendingCatastrophe, ge.fate)
	}
	ge.rng = riggedRNG(0.01) // under any strike chance
	tickTo(ge, strike)
	if ge.pendingCatastrophe != "iron_era" || ge.fate.Resolved != FateStruck || ge.fate.ResolvedTick != strike || ge.fate.AtAdvance {
		t.Fatalf("at the strike tick: pending %q fate %+v", ge.pendingCatastrophe, ge.fate)
	}
	if len(cat) != 1 || cat[0].Payload["event_type"] != "catastrophe" {
		t.Errorf("catastrophe events = %+v", cat)
	}
	if ge.harbinger != nil || len(ge.harbingerHistory) != 1 || ge.harbingerHistory[0].Outcome != HarbingerOutcomeVindicated {
		t.Errorf("thread %+v history %+v; want it resolved vindicated", ge.harbinger, ge.harbingerHistory)
	}
	if r := ge.harbingerHistory[0]; r.StrikeTick != strike || r.EntryTick != ge.fate.EntryTick || r.Window != ge.fate.Window || r.ArrivedTick == 0 {
		t.Errorf("record timing = %+v", r)
	}
	if ge.pendingBraceLevel != 1 {
		t.Errorf("Brace not handed to the pending catastrophe: %d", ge.pendingBraceLevel)
	}
	// One doom per era: nothing more comes.
	tickTo(ge, strike+50000)
	if n := len(cat); n != 1 {
		t.Errorf("%d catastrophes in one era", n)
	}
}

// A struck doom mid-age waits as a pending catastrophe that blocks advancing
// and prestige until it is answered.
func TestMidAgeStrikeBlocksAdvanceAndPrestige(t *testing.T) {
	ge := fateEngine(t, "digital_age", 2)
	forceFate(t, ge, 1000)
	tickTo(ge, arrivalTick(ge))
	ge.rng = riggedRNG(0.01)
	tickTo(ge, ge.fate.StrikeTick)
	if ge.pendingCatastrophe != "digital_era" {
		t.Fatalf("setup: pending %q", ge.pendingCatastrophe)
	}
	readyToAdvance(ge)
	if err := ge.AdvanceAge(); err == nil || !strings.Contains(err.Error(), "catastrophe") {
		t.Errorf("advance while pending: %v", err)
	}
	if err := ge.DoPrestige(); err == nil || !strings.Contains(err.Error(), "catastrophe") {
		t.Errorf("prestige while pending: %v", err)
	}
	if st := ge.GetState(); st.PendingCatastrophe != "digital_era" {
		t.Errorf("snapshot pending %q; the popup reads it", st.PendingCatastrophe)
	}
	if err := ge.Endure(); err != nil {
		t.Fatal(err)
	}
	if ge.ageReady {
		t.Error("Endure left the old readiness standing: the requirements must be checked afresh")
	}
	if err := ge.DoPrestige(); err != nil {
		t.Errorf("prestige after Endure: %v", err)
	}
}

// Appease lowers the chance the strike hits (read at the strike tick, by the
// faith band then); a miss is Spared, logged and recorded.
func TestAppeaseLowersTheStrikeAndSparedIsHandled(t *testing.T) {
	strike := func(appease int, roll float64) *GameEngine {
		ge := fateEngine(t, "classical_age", 11)
		forceFate(t, ge, 26000)
		tickTo(ge, arrivalTick(ge))
		ge.harbinger.AppeaseLevel = appease
		setFaith(ge, 50, 100) // mid band: 75% before Appease
		ge.rng = riggedRNG(roll)
		tickTo(ge, ge.fate.StrikeTick)
		return ge
	}
	for _, tc := range []struct {
		appease int
		chance  float64
	}{{0, 0.75}, {1, 0.45}, {2, 0.27}} {
		ge := strike(tc.appease, tc.chance-0.01)
		if ge.fate.Resolved != FateStruck {
			t.Errorf("appease %d: a roll just under %.2f missed", tc.appease, tc.chance)
		}
		ge = strike(tc.appease, tc.chance+0.01)
		if ge.fate.Resolved != FateSpared || ge.pendingCatastrophe != "" {
			t.Fatalf("appease %d: a roll just over %.2f struck: %+v", tc.appease, tc.chance, ge.fate)
		}
		if len(ge.harbingerHistory) != 1 || ge.harbingerHistory[0].Outcome != HarbingerOutcomeSpared {
			t.Errorf("appease %d: history %+v, want spared", tc.appease, ge.harbingerHistory)
		}
		if countLogs(ge, "passed you by") != 1 {
			t.Errorf("appease %d: no Spared line:\n%s", tc.appease, strings.Join(logMessages(ge), "\n"))
		}
		if o := ge.CatastropheOutlook(); o.Possible {
			t.Errorf("appease %d: outlook after Spared = %+v, want the era done", tc.appease, o)
		}
	}
	// The faith band is the one at the strike tick.
	ge := fateEngine(t, "classical_age", 11)
	forceFate(t, ge, 26000)
	tickTo(ge, arrivalTick(ge))
	setFaith(ge, 90, 100)
	ge.rng = riggedRNG(0.65) // over the high band's 60%, under the mid band's 75%
	tickTo(ge, ge.fate.StrikeTick)
	if ge.fate.Resolved != FateSpared {
		t.Errorf("high faith at the strike: %+v, want spared", ge.fate)
	}
}

// StrikeChanceAt, the pure rule the smoke report models with, gives what the
// engine's strike rolls against.
func TestStrikeChanceAtMatchesTheEngine(t *testing.T) {
	for _, fill := range []struct {
		amount, storage float64
	}{{0, 0}, {10, 100}, {50, 100}, {90, 100}} {
		for appease := 0; appease <= HarbingerMaxAppease; appease++ {
			ge := fateEngine(t, "classical_age", 1)
			forceFate(t, ge, 26000)
			tickTo(ge, arrivalTick(ge))
			ge.harbinger.AppeaseLevel = appease
			setFaith(ge, fill.amount, fill.storage)
			f, ok := ge.faithFill()
			if got, want := StrikeChanceAt(f, ok, appease), ge.strikeChance(); math.Abs(got-want) > 1e-12 {
				t.Errorf("faith %v/%v appease %d: StrikeChanceAt %v, engine %v", fill.amount, fill.storage, appease, got, want)
			}
		}
	}
}

// --- can't outrun it -------------------------------------------------------------------

// Reaching the era's final transition before the strike tick brings the
// strike there, before the advance; with no harbinger yet it comes first and
// the advance waits for one more try.
func TestStrikeLandsAtTheTransitionWhenOutrun(t *testing.T) {
	t.Run("harbinger present", func(t *testing.T) {
		ge := fateEngine(t, "medieval_age", 3)
		forceFate(t, ge, 18000)
		tickTo(ge, arrivalTick(ge))
		if ge.harbinger == nil {
			t.Fatal("no harbinger")
		}
		ge.rng = riggedRNG(0.01)
		readyToAdvance(ge)
		err := ge.AdvanceAge()
		if err == nil || ge.age != "medieval_age" || ge.pendingCatastrophe != "iron_era" {
			t.Fatalf("advance: err %v age %s pending %q; want the strike first", err, ge.age, ge.pendingCatastrophe)
		}
		if !ge.fate.AtAdvance || ge.fate.Resolved != FateStruck {
			t.Errorf("fate = %+v, want struck at the advance", ge.fate)
		}
		if err := ge.Endure(); err != nil {
			t.Fatal(err)
		}
		readyToAdvance(ge)
		if err := ge.AdvanceAge(); err != nil || ge.age != "renaissance_age" {
			t.Errorf("advance after Endure: %v, age %s", err, ge.age)
		}
	})
	t.Run("spared at the gate", func(t *testing.T) {
		ge := fateEngine(t, "medieval_age", 3)
		forceFate(t, ge, 18000)
		tickTo(ge, arrivalTick(ge))
		ge.rng = riggedRNG(0.999)
		readyToAdvance(ge)
		if err := ge.AdvanceAge(); err != nil || ge.age != "renaissance_age" {
			t.Fatalf("advance: %v, age %s; want spared and through", err, ge.age)
		}
		if n := len(ge.harbingerHistory); n != 1 || ge.harbingerHistory[0].Outcome != HarbingerOutcomeSpared || !ge.harbingerHistory[0].AtAdvance {
			t.Errorf("history %+v", ge.harbingerHistory)
		}
	})
	t.Run("no harbinger yet", func(t *testing.T) {
		ge := fateEngine(t, "medieval_age", 3)
		forceFate(t, ge, 18800)
		tickTo(ge, 100)
		if ge.harbinger != nil {
			t.Fatal("setup: the harbinger came early")
		}
		readyToAdvance(ge)
		err := ge.AdvanceAge()
		if err == nil || !strings.Contains(err.Error(), "stands in your way") || ge.harbinger == nil || ge.age != "medieval_age" {
			t.Fatalf("first advance: err %v thread %+v age %s; want the harbinger and a wait", err, ge.harbinger, ge.age)
		}
		if ge.pendingCatastrophe != "" {
			t.Fatal("struck without a warning")
		}
		ge.rng = riggedRNG(0.01)
		if err := ge.AdvanceAge(); err == nil || ge.pendingCatastrophe != "iron_era" {
			t.Errorf("second advance: err %v pending %q; want the strike", err, ge.pendingCatastrophe)
		}
	})
	t.Run("within this age", func(t *testing.T) {
		// The Oracle said "before this age is out": leaving the age early
		// brings the doom first.
		ge := fateEngine(t, "classical_age", 3)
		ages := config.EpochByKey()["iron_era"].Ages
		forceFate(t, ge, int(expectedAgeTicks(ages[0])+0.8*expectedAgeTicks(ages[1])))
		tickTo(ge, arrivalTick(ge))
		if ge.harbinger == nil || ge.harbinger.When != WhenThisAge {
			t.Fatalf("setup: thread %+v", ge.harbinger)
		}
		ge.rng = riggedRNG(0.01)
		readyToAdvance(ge)
		if err := ge.AdvanceAge(); err == nil || ge.pendingCatastrophe != "iron_era" || ge.age != "classical_age" {
			t.Errorf("advance within the promised age: err %v pending %q age %s", err, ge.pendingCatastrophe, ge.age)
		}
	})
	t.Run("before the era ends", func(t *testing.T) {
		// "Before the era ends": an advance inside the era does not bring it.
		ge := fateEngine(t, "classical_age", 3)
		forceFate(t, ge, int(expectedEraTicks("iron_era"))-10)
		ge.fate.LeadFrac = harbingerLeadMax
		delete(ge.harbingerArrived, "iron_era")
		ge.fateArrive()
		if ge.harbinger.When != WhenThisEra {
			t.Fatalf("setup: when %q", ge.harbinger.When)
		}
		readyToAdvance(ge)
		if err := ge.AdvanceAge(); err != nil || ge.age != "medieval_age" || ge.fate.Resolved != "" {
			t.Errorf("advance inside the era: err %v age %s fate %+v", err, ge.age, ge.fate)
		}
		if ge.harbinger == nil || ge.harbinger.Age != "medieval_age" || ge.harbinger.When != WhenThisAge {
			t.Errorf("handoff: %+v, want the Town Crier saying this age", ge.harbinger)
		}
	})
}

// The build plan's advance obeys the same gate.
func TestPlanAdvanceCannotOutrunTheDoom(t *testing.T) {
	ge := fateEngine(t, "medieval_age", 3)
	forceFate(t, ge, 18800)
	ge.plan = []PlanItem{{Kind: PlanAdvance, Count: 1}}
	ge.pendingCatastrophe = ""
	// Meet the requirements the easy way: the plan asks CheckAdvancement.
	ge.progress = newProgressManagerMet(t, ge)
	var s planStarts
	ge.runPlan(&s)
	if ge.age != "medieval_age" || ge.harbinger == nil || len(ge.plan) != 1 {
		t.Fatalf("first plan run: age %s thread %+v plan %+v; want the harbinger and the item kept", ge.age, ge.harbinger, ge.plan)
	}
	ge.rng = riggedRNG(0.01)
	ge.runPlan(&s)
	if ge.age != "medieval_age" || ge.pendingCatastrophe != "iron_era" || len(ge.plan) != 1 {
		t.Errorf("second plan run: age %s pending %q plan %+v; want the strike and the item waiting", ge.age, ge.pendingCatastrophe, ge.plan)
	}
}

// --- invite ----------------------------------------------------------------------------

// Invite makes the strike certain (it still comes at its tick), is only there
// while a harbinger is, once per era; in the Stone Era it is refused.
func TestInviteGuaranteesTheStrike(t *testing.T) {
	ge := fateEngine(t, "classical_age", 13)
	if err := ge.ForceQuietFateForTest("iron_era"); err != nil {
		t.Fatal(err)
	}
	if err := ge.HarbingerInvite(); err == nil || !strings.Contains(err.Error(), "No harbinger") {
		t.Errorf("invite without a harbinger: %v", err)
	}
	forceFate(t, ge, 26000)
	tickTo(ge, arrivalTick(ge))
	setStock(ge, map[string][2]float64{"faith": {1e6, 1e6}})
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	if !ge.fate.Invited || !ge.harbinger.Invited || ge.catastropheInvited {
		t.Errorf("invite state: fate %+v thread invited %v flag %v", ge.fate, ge.harbinger.Invited, ge.catastropheInvited)
	}
	if o := ge.CatastropheOutlook(); o.Probability != 1 || o.Tier != CatastropheTierHigh {
		t.Errorf("invited outlook = %+v", o)
	}
	if err := ge.HarbingerInvite(); err == nil {
		t.Error("a second invite must be refused")
	}
	if err := ge.HarbingerAppease(); err == nil || !strings.Contains(err.Error(), "invited") {
		t.Errorf("appease after invite: %v", err)
	}
	ge.rng = riggedRNG(0.999) // a miss under any odds: the invite overrides it
	tickTo(ge, ge.fate.StrikeTick-1)
	if ge.pendingCatastrophe != "" {
		t.Fatal("an invite made it strike early")
	}
	tickTo(ge, ge.fate.StrikeTick)
	if ge.pendingCatastrophe != "iron_era" || ge.harbingerHistory[len(ge.harbingerHistory)-1].Outcome != HarbingerOutcomeFulfilled {
		t.Fatalf("invited doom: pending %q history %+v", ge.pendingCatastrophe, ge.harbingerHistory)
	}
	last := ge.epochEventHistory[len(ge.epochEventHistory)-1]
	if !strings.Contains(last.EventName, "invited") {
		t.Errorf("record = %+v", last)
	}

	// A false prophet's invented doom becomes real.
	fp := fateEngine(t, "classical_age", 14)
	if err := fp.ForceFalseProphetForTest("iron_era", 26000); err != nil {
		t.Fatal(err)
	}
	tickTo(fp, arrivalTick(fp))
	if err := fp.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	tickTo(fp, fp.fate.StrikeTick)
	if fp.pendingCatastrophe != "iron_era" {
		t.Fatalf("invited false prophet: pending %q", fp.pendingCatastrophe)
	}
	if r := fp.harbingerHistory[len(fp.harbingerHistory)-1]; r.Outcome != HarbingerOutcomeFulfilled || !r.FalseProphet {
		t.Errorf("record = %+v, want fulfilled and false", r)
	}

	// Nothing can strike in the Stone Era: every answer is refused there.
	stone := fateEngine(t, "stone_age", 15)
	if err := stone.ForceFalseProphetForTest("stone_era", 100); err != nil {
		t.Fatal(err)
	}
	tickTo(stone, 100)
	if stone.harbinger == nil {
		t.Fatal("no stone era harbinger")
	}
	for name, act := range map[string]func() error{"invite": stone.HarbingerInvite, "brace": stone.HarbingerBrace, "appease": stone.HarbingerAppease} {
		if err := act(); err == nil || !strings.Contains(err.Error(), "no catastrophe can strike in the Stone Era") {
			t.Errorf("stone era %s: %v", name, err)
		}
	}
}

// --- false prophets ------------------------------------------------------------------------

// A false prophet arrives like a real harbinger, claims medium or high, lets
// its foretold moment pass, and is revealed when its window does: at the age's
// end for "before this age is out", at the era's end otherwise.
func TestFalseProphetRevealedAfterItsWindow(t *testing.T) {
	t.Run("untimed, revealed at the era's end", func(t *testing.T) {
		ge := fateEngine(t, "iron_age", 16)
		if err := ge.ForceFalseProphetForTest("iron_era", 2000); err != nil {
			t.Fatal(err)
		}
		tickTo(ge, arrivalTick(ge))
		h := ge.harbinger
		if h == nil || !h.FalseProphet || h.When != WhenUntold || (h.AnnouncedTier != CatastropheTierHigh && h.AnnouncedTier != CatastropheTierMedium) {
			t.Fatalf("false thread = %+v", h)
		}
		v := ge.GetState().Harbinger
		if v.Tier != CatastropheTierHigh || math.Abs(v.Probability-0.9) > 1e-9 {
			t.Errorf("claim = %s %.2f, want the high claim (90%%)", v.Tier, v.Probability)
		}
		tickTo(ge, ge.fate.StrikeTick+10)
		if ge.pendingCatastrophe != "" || ge.harbinger == nil || ge.fate.Resolved != "" {
			t.Fatalf("after the foretold moment: pending %q thread %v fate %+v; want nothing yet", ge.pendingCatastrophe, ge.harbinger, ge.fate)
		}
		readyToAdvance(ge)
		if err := ge.AdvanceAge(); err != nil || ge.age != "classical_age" {
			t.Fatalf("advance inside the era: %v", err)
		}
		if ge.harbinger == nil || ge.harbinger.Age != "classical_age" || ge.fate.Resolved != "" {
			t.Fatalf("the Oracle should take up the lie: %+v / %+v", ge.harbinger, ge.fate)
		}
		ge.harbinger.When = WhenThisEra
		ge.advanceAge("medieval_age")
		ge.harbinger.When = WhenUntold
		readyToAdvance(ge)
		if err := ge.AdvanceAge(); err != nil || ge.age != "renaissance_age" {
			t.Fatalf("leaving the era: %v, age %s", err, ge.age)
		}
		r := ge.harbingerHistory[len(ge.harbingerHistory)-1]
		if r.Outcome != HarbingerOutcomeDiscredited || !r.FalseProphet || ge.pendingCatastrophe != "" {
			t.Errorf("record = %+v pending %q", r, ge.pendingCatastrophe)
		}
		if countLogs(ge, "The Iron Era ends without the doom") != 1 {
			t.Errorf("no era-end reveal line:\n%s", strings.Join(logMessages(ge), "\n"))
		}
	})
	t.Run("timed, revealed at the age's end", func(t *testing.T) {
		ge := fateEngine(t, "classical_age", 17)
		ages := config.EpochByKey()["iron_era"].Ages
		if err := ge.ForceFalseProphetForTest("iron_era", int(expectedAgeTicks(ages[0])+0.7*expectedAgeTicks(ages[1]))); err != nil {
			t.Fatal(err)
		}
		tickTo(ge, arrivalTick(ge))
		if ge.harbinger == nil || ge.harbinger.When != WhenThisAge {
			t.Fatalf("setup: %+v", ge.harbinger)
		}
		readyToAdvance(ge)
		if err := ge.AdvanceAge(); err != nil || ge.age != "medieval_age" {
			t.Fatalf("advance: %v", err)
		}
		if ge.harbinger != nil || ge.fate.Resolved != FateRevealed {
			t.Errorf("after the promised age: thread %+v fate %+v; want revealed", ge.harbinger, ge.fate)
		}
		if countLogs(ge, "This age ends without the doom") != 1 {
			t.Errorf("no age-end reveal line:\n%s", strings.Join(logMessages(ge), "\n"))
		}
		// Revealed, the era reads quiet again: nothing says it is safe.
		if o := ge.CatastropheOutlook(); !o.Possible || o.Warned {
			t.Errorf("outlook after a reveal = %+v, want quiet", o)
		}
	})
	t.Run("stone era, never came", func(t *testing.T) {
		// In the Stone Era nothing can strike: a false prophet who has not
		// come by its end never comes, and the advance is not held up.
		ge := fateEngine(t, "bronze_age", 18)
		if err := ge.ForceFalseProphetForTest("stone_era", int(expectedEraTicks("stone_era"))-1); err != nil {
			t.Fatal(err)
		}
		tickTo(ge, 10)
		if ge.harbinger != nil {
			t.Fatal("setup: the false prophet came early")
		}
		readyToAdvance(ge)
		if err := ge.AdvanceAge(); err != nil || ge.age != "iron_age" || ge.harbinger != nil {
			t.Fatalf("leaving the Stone Era: err %v age %s thread %+v; want through, no harbinger", err, ge.age, ge.harbinger)
		}
	})
	t.Run("stone era", func(t *testing.T) {
		ge := fateEngine(t, "primitive_age", 18)
		if err := ge.ForceFalseProphetForTest("stone_era", 200); err != nil {
			t.Fatal(err)
		}
		tickTo(ge, arrivalTick(ge))
		if ge.harbinger == nil || ge.harbinger.Age != "primitive_age" || ge.harbinger.When != WhenUntold {
			t.Fatalf("the Wild Man should come, untimed: %+v", ge.harbinger)
		}
		walkTo(t, ge, "bronze_age")
		readyToAdvance(ge)
		if err := ge.AdvanceAge(); err != nil || ge.age != "iron_age" {
			t.Fatalf("leaving the Stone Era: %v", err)
		}
		if r := ge.harbingerHistory[len(ge.harbingerHistory)-1]; r.Outcome != HarbingerOutcomeDiscredited {
			t.Errorf("record = %+v", r)
		}
	})
}

// --- offline ---------------------------------------------------------------------------

// Offline catch-up keeps the doom's hour: the harbinger comes and the doom
// strikes at its tick while the player is away, and it waits, pending.
func TestOfflineCatchUpAcrossAStrike(t *testing.T) {
	ge := fateEngine(t, "classical_age", 19)
	forceFate(t, ge, 3000)
	ge.rng = riggedRNG(0.01)
	var resolved []EventData
	ge.Bus.Subscribe(EventFateResolved, func(e EventData) { resolved = append(resolved, e) })
	ge.SimulateOffline(8 * time.Hour)
	if ge.pendingCatastrophe != "iron_era" {
		t.Fatalf("after 8h away: pending %q, fate %+v", ge.pendingCatastrophe, ge.fate)
	}
	if ge.fate.ResolvedTick != ge.fate.StrikeTick {
		t.Errorf("struck at tick %d, fated for %d", ge.fate.ResolvedTick, ge.fate.StrikeTick)
	}
	if len(resolved) != 1 || resolved[0].Payload["tick"] != ge.fate.StrikeTick {
		t.Fatalf("resolve events = %+v", resolved)
	}
	if a, ok := resolved[0].Payload["arrived_tick"].(int); !ok || a >= ge.fate.StrikeTick {
		t.Errorf("resolve event arrival %v, want the harbinger's tick before the strike", resolved[0].Payload["arrived_tick"])
	}
	if countLogs(ge, "has come, warning of impending doom") != 1 || countLogs(ge, "was right") != 1 {
		t.Errorf("the away log should show the warning and the strike:\n%s", strings.Join(logMessages(ge), "\n"))
	}
	if ge.tick < ge.fate.StrikeTick+100 {
		t.Errorf("the catch-up stopped at the strike: tick %d", ge.tick)
	}
}

// --- the Last Passage -----------------------------------------------------------------------

// The Cosmic Era keeps its Last Passage thread from its first age and rolls no
// fate: prestige there still brings the Last Passage as before.
func TestLastPassageUnchangedByFates(t *testing.T) {
	ge := catEngine(t, "space_age", 20)
	for _, a := range config.AgeOrder() {
		ge.applyAgeUnlocks(a)
		if a == "space_age" {
			break
		}
	}
	ge.harbingerTickCheck()
	if err := ge.ForceQuietFateForTest("neon_era"); err != nil {
		t.Fatal(err)
	}
	readyToAdvance(ge)
	if err := ge.AdvanceAge(); err != nil {
		t.Fatal(err)
	}
	if ge.fate != nil {
		t.Errorf("the Cosmic Era rolled a fate: %+v", ge.fate)
	}
	if h := ge.harbinger; h == nil || h.TargetEpoch != "" || h.Age != "interstellar_age" || h.When != WhenUntold {
		t.Fatalf("cosmic thread on entry = %+v", h)
	}
	if o := ge.CatastropheOutlook(); o.Passage != PassagePrestige || !o.Possible || o.Warned {
		t.Errorf("cosmic outlook = %+v", o)
	}
}

// newProgressManagerMet is ge's progress manager with the current age's
// requirements met (every resource and building requirement removed).
func newProgressManagerMet(t *testing.T, ge *GameEngine) *ProgressManager {
	t.Helper()
	pm := NewProgressManager()
	for i := range pm.ages {
		pm.ages[i].ResourceReqs = nil
		pm.ages[i].BuildingReqs = nil
	}
	pm.ageWonders = map[string]string{}
	return pm
}
