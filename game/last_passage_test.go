package game

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
)

// --- helpers ------------------------------------------------------------------

// lpEngine returns a seeded engine standing in age (a Cosmic Era age) with the
// Last Passage thread started, as the first tick of the epoch would.
func lpEngine(t *testing.T, age string, seed int64) *GameEngine {
	t.Helper()
	ge := catEngine(t, age, seed)
	ge.harbingerTickCheck()
	if ge.harbinger == nil || ge.harbinger.TargetEpoch != "" {
		t.Fatalf("no Last Passage thread at %s: %+v", age, ge.harbinger)
	}
	return ge
}

// makeLastPassagePending prestiges ge with a roll that always hits.
func makeLastPassagePending(t *testing.T, ge *GameEngine) {
	t.Helper()
	ge.rng = riggedRNG(0.0, 0.5, 0.25, 0.75)
	if err := ge.DoPrestige(); err != nil {
		t.Fatalf("DoPrestige: %v", err)
	}
	if !ge.pendingLastPassage {
		t.Fatalf("the Last Passage did not come on a roll of 0")
	}
}

// endureLastPassageIfPending lets a test that prestiges from the final epoch
// get past a Last Passage the seed happened to roll.
func endureLastPassageIfPending(t *testing.T, ge *GameEngine) {
	t.Helper()
	if !ge.GetState().LastPassage.Pending {
		return
	}
	if err := ge.EndureLastPassage(); err != nil {
		t.Fatalf("EndureLastPassage: %v", err)
	}
}

// hasCosmicLegacyModifier reports whether the resolver carries the Cosmic
// Legacy's +10% production_all.
func hasCosmicLegacyModifier(ge *GameEngine) bool {
	for _, m := range ge.buildResolver().All() {
		if m.Source == cosmicLegacySource && m.Target == "production_all" && m.Op == OpAdd && m.Value == CosmicLegacyProductionBonus {
			return true
		}
	}
	return false
}

// indexOfLog returns the index of the first log line containing substr, or -1.
func indexOfLog(ge *GameEngine, substr string) int {
	for i, e := range ge.log {
		if strings.Contains(e.Message, substr) {
			return i
		}
	}
	return -1
}

// --- the cosmic thread ----------------------------------------------------------

func TestCosmicThreadStartsAndHandsOff(t *testing.T) {
	ge := threadEngine(t, "cosmic_era", 21)
	h := ge.harbinger
	if h.EpochKey != "cosmic_era" || h.TargetEpoch != "" || h.Age != "interstellar_age" || h.FalseProphet {
		t.Fatalf("cosmic thread = %+v", h)
	}
	if countLogs(ge, "warning of the Last Passage") != 1 {
		t.Errorf("arrival log should name the Last Passage:\n%s", strings.Join(logMessages(ge), "\n"))
	}
	if v := ge.GetState().Harbinger; v == nil || !v.LastPassage || v.TargetEpochName != "the Last Passage" || v.TargetEpochKey != "" {
		t.Errorf("view = %+v", v)
	}

	// Answers carry across the handoffs.
	setStock(ge, map[string][2]float64{"faith": {5e10, 5e10}, "culture": {5e10, 5e10}})
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatalf("appease: %v", err)
	}
	fillBraceStock(ge, 2e12)
	if err := ge.HarbingerBrace(); err != nil {
		t.Fatalf("brace: %v", err)
	}
	if countLogs(ge, "you keep 70% of the run's prestige points") != 1 {
		t.Errorf("brace log should talk about points:\n%s", strings.Join(logMessages(ge), "\n"))
	}

	wantNames := map[string]string{
		"interstellar_age": "the Distress Beacon",
		"galactic_age":     "the Elder Relay",
		"quantum_age":      "your future self",
		"transcendent_age": "your unmade self",
	}
	chain := []string{"interstellar_age"}
	for _, age := range []string{"galactic_age", "quantum_age", "transcendent_age"} {
		walkTo(t, ge, age)
		chain = append(chain, age)
		h := ge.harbinger
		if h == nil || h.Age != age || h.TargetEpoch != "" || strings.Join(h.Chain, ",") != strings.Join(chain, ",") {
			t.Fatalf("at %s: thread = %+v", age, h)
		}
		if h.AppeaseLevel != 1 || h.BraceLevel != 1 {
			t.Errorf("at %s: levels appease %d brace %d, want 1 and 1", age, h.AppeaseLevel, h.BraceLevel)
		}
		def, _ := config.HarbingerFor(age)
		if def.Name != wantNames[age] {
			t.Errorf("at %s: speaker %q, want %q", age, def.Name, wantNames[age])
		}
		if v := ge.GetState().Harbinger; v == nil || v.Name != wantNames[age] {
			t.Errorf("at %s: view %+v", age, v)
		}
	}
	if n := countLogs(ge, "takes up the warning of the Last Passage"); n != 3 {
		t.Errorf("handoff logs = %d, want 3", n)
	}
	if def, _ := config.HarbingerFor("interstellar_age"); def.Name != wantNames["interstellar_age"] {
		t.Errorf("interstellar speaker = %q", def.Name)
	}
}

func TestCosmicThreadCostsAreFixedAcrossTheEpoch(t *testing.T) {
	for _, age := range epochAges(t, "cosmic_era") {
		ge := lpEngine(t, age, 5)
		v := ge.GetState().Harbinger
		if v.AppeaseCost["faith"] != 1200000000 || v.AppeaseCost["culture"] != 19000000000 {
			t.Errorf("%s: appease = %v", age, v.AppeaseCost)
		}
		if v.BraceCost["dark_matter"] != 1560000000000 || v.BraceCost["titanium"] != 75600000000 || len(v.BraceCost) != 2 {
			t.Errorf("%s: brace = %v", age, v.BraceCost)
		}
		if v.EndurePointsPct != 50 || v.NextEndurePointsPct != 70 {
			t.Errorf("%s: endure points %d%% next %d%%", age, v.EndurePointsPct, v.NextEndurePointsPct)
		}
	}
}

// --- the outlook -------------------------------------------------------------------

func TestOutlookPassageKind(t *testing.T) {
	for _, ep := range config.Epochs() {
		ge := catEngine(t, ep.Ages[0], 1)
		o := ge.CatastropheOutlook()
		want := PassageEpoch
		if config.IsFinalEpoch(ep.Key) {
			want = PassagePrestige
		}
		if o.Passage != want {
			t.Errorf("%s: passage %q, want %q", ep.Key, o.Passage, want)
		}
		if want == PassagePrestige && (o.NextEpochKey != "" || !o.Possible) {
			t.Errorf("%s: prestige outlook = %+v", ep.Key, o)
		}
		if want == PassageEpoch && o.NextEpochKey == "" {
			t.Errorf("%s: epoch outlook has no next epoch: %+v", ep.Key, o)
		}
	}
	// Invited: certain. Appeased twice: ×0.36.
	ge := lpEngine(t, "galactic_age", 1)
	ge.harbinger.AppeaseLevel = 2
	if p := ge.CatastropheOutlook().Probability; p < 0.0647 || p > 0.0649 {
		t.Errorf("appeased twice: %v, want 0.18 × 0.36", p)
	}
	ge.catastropheInvited = true
	if p := ge.CatastropheOutlook().Probability; p != 1 {
		t.Errorf("invited: %v, want 1", p)
	}
}

// --- the roll ------------------------------------------------------------------------

func TestLastPassageOnlyRollsFromTheCosmicEra(t *testing.T) {
	// A stray invite before the Cosmic Era is not a Last Passage.
	ge := catEngine(t, "space_age", 3)
	ge.catastropheInvited = true
	ge.rng = riggedRNG(0.0)
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if ge.pendingLastPassage || ge.Prestige.GetLevel() != 1 || ge.age != "primitive_age" {
		t.Errorf("space-age prestige: pending %v level %d age %s", ge.pendingLastPassage, ge.Prestige.GetLevel(), ge.age)
	}

	ge = lpEngine(t, "interstellar_age", 3)
	makeLastPassagePending(t, ge)
	if ge.Prestige.GetLevel() != 0 || ge.age != "interstellar_age" {
		t.Errorf("pending Last Passage completed the prestige: level %d age %s", ge.Prestige.GetLevel(), ge.age)
	}
	st := ge.GetState()
	if !st.LastPassage.Pending || st.CatastropheOutlook.Possible || st.PendingCatastrophe != "" {
		t.Errorf("pending state = %+v, outlook %+v", st.LastPassage, st.CatastropheOutlook)
	}
	if countLogs(ge, "The Last Passage has come") != 1 {
		t.Errorf("no arrival log:\n%s", strings.Join(logMessages(ge), "\n"))
	}
}

func TestLastPassageSparedCompletesPrestige(t *testing.T) {
	ge := lpEngine(t, "transcendent_age", 4)
	full := ge.GetState().Prestige.PendingPoints
	ge.rng = riggedRNG(0.99, 0.3, 0.6, 0.1)
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if ge.pendingLastPassage || ge.Prestige.GetLevel() != 1 || ge.Prestige.Snapshot().Available != full || ge.age != "primitive_age" {
		t.Fatalf("spared: pending %v level %d points %d (want %d) age %s", ge.pendingLastPassage,
			ge.Prestige.GetLevel(), ge.Prestige.Snapshot().Available, full, ge.age)
	}
	verdict := indexOfLog(ge, "you were spared")
	done := indexOfLog(ge, "Prestige complete.")
	if verdict < 0 || done < 0 || verdict > done {
		t.Fatalf("verdict %d, prestige line %d:\n%s", verdict, done, strings.Join(logMessages(ge), "\n"))
	}
	if !strings.Contains(ge.log[verdict].Message, "The Last Passage opens") {
		t.Errorf("verdict = %q", ge.log[verdict].Message)
	}
	if ge.harbinger != nil || ge.harbingerHistory != nil || ge.catastropheInvited {
		t.Errorf("thread not cleared: %+v", ge.harbinger)
	}
	if len(ge.catastropheHistory) != 0 {
		t.Errorf("a spared passage wrote the civilization log: %v", ge.catastropheHistory)
	}
}

func TestLastPassageEndureKeepsAShareOfThePoints(t *testing.T) {
	for brace, keep := range []float64{0.50, 0.70, 0.85} {
		ge := lpEngine(t, "transcendent_age", 6)
		ge.harbinger.BraceLevel = brace
		full := ge.GetState().Prestige.PendingPoints
		want := int(float64(full) * keep)
		if lp := ge.GetState().LastPassage; lp.KeepPct != int(keep*100+0.5) || lp.PointsIfEndured != want || lp.BraceLevel != brace {
			t.Errorf("brace %d: state %+v, want keep %v and %d points", brace, lp, keep, want)
		}
		makeLastPassagePending(t, ge)
		if err := ge.EndureLastPassage(); err != nil {
			t.Fatalf("brace %d: %v", brace, err)
		}
		if got := ge.Prestige.Snapshot().Available; got != want || ge.Prestige.GetLevel() != 1 || ge.age != "primitive_age" {
			t.Errorf("brace %d: %d points level %d age %s, want %d of %d", brace, got, ge.Prestige.GetLevel(), ge.age, want, full)
		}
		if ge.cosmicLegacy || ge.pendingLastPassage {
			t.Errorf("brace %d: Endure granted the legacy or left it pending", brace)
		}
		if indexOfLog(ge, "was right") < 0 || indexOfLog(ge, "ENDURE: The Last Passage") < 0 {
			t.Errorf("brace %d: missing verdict or outcome:\n%s", brace, strings.Join(logMessages(ge), "\n"))
		}
		if e, _ := countCatastropheOutcomes(ge.catastropheHistory); e != 1 {
			t.Errorf("brace %d: endured count %d, want 1", brace, e)
		}
	}
	// A run too short to keep anything keeps nothing, never less.
	ge := lpEngine(t, "interstellar_age", 6)
	if got := ge.lastPassagePoints(lastPassageEndured, 1); got != 0 {
		t.Errorf("Endure of 1 point = %d, want 0", got)
	}
}

func TestLastPassageSuccumbGrantsTheCosmicLegacy(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := lpEngine(t, "quantum_age", 7)
	ge.Prestige.LoadState(2, 10, 4, nil) // a veteran: level 2, 4 points in the bank
	if hasCosmicLegacyModifier(ge) {
		t.Fatal("legacy before Succumb")
	}
	makeLastPassagePending(t, ge)
	if err := ge.SuccumbLastPassage(); err != nil {
		t.Fatal(err)
	}
	snap := ge.Prestige.Snapshot()
	if snap.Level != 3 || snap.Available != 4 || snap.TotalEarned != 10 || !ge.cosmicLegacy || ge.age != "primitive_age" {
		t.Fatalf("after Succumb: %+v legacy %v age %s", snap, ge.cosmicLegacy, ge.age)
	}
	if indexOfLog(ge, "Cosmic Legacy: all production +10%") < 0 || indexOfLog(ge, "SUCCUMB: The Last Passage") < 0 {
		t.Errorf("log:\n%s", strings.Join(logMessages(ge), "\n"))
	}
	if _, s := countCatastropheOutcomes(ge.catastropheHistory); s != 1 {
		t.Errorf("succumbed count %d, want 1", s)
	}

	check := func(when string) {
		t.Helper()
		if !ge.cosmicLegacy || !hasCosmicLegacyModifier(ge) {
			t.Fatalf("%s: legacy %v, modifier %v", when, ge.cosmicLegacy, hasCosmicLegacyModifier(ge))
		}
		found := false
		for _, m := range ge.GetState().Modifiers {
			if m.Source == cosmicLegacySource && m.Target == "production_all" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: the Cosmic Legacy is missing from GameState.Modifiers", when)
		}
		if _, ok := ge.permanentBonuses["production_all"]; ok {
			t.Errorf("%s: the legacy was stored as a bonus value", when)
		}
	}
	check("after Succumb")
	base := ge.buildResolver().AddTotal("production_all")

	// Two more prestiges, one of them from the Cosmic Era.
	ge.age, ge.currentEpoch = "modern_age", config.EpochForAge("modern_age")
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	check("after a Modern Age prestige")
	ge.age, ge.currentEpoch = "galactic_age", config.EpochForAge("galactic_age")
	ge.harbingerTickCheck()
	makeLastPassagePending(t, ge)
	// Owned already: Succumb is refused, Endure is the only answer.
	if err := ge.SuccumbLastPassage(); err == nil || !strings.Contains(err.Error(), "already carry the Cosmic Legacy") {
		t.Fatalf("second Succumb: err = %v", err)
	}
	if err := ge.Succumb(); err == nil || !ge.pendingLastPassage {
		t.Fatalf("Succumb() routed past the refusal: err %v pending %v", err, ge.pendingLastPassage)
	}
	if err := ge.EndureLastPassage(); err != nil {
		t.Fatal(err)
	}
	check("after a Cosmic Era prestige")
	// The epoch catastrophe's Succumb resets the run too.
	ge.age, ge.currentEpoch = "iron_age", "iron_era"
	ge.pendingCatastrophe = "iron_era"
	if err := ge.Succumb(); err != nil {
		t.Fatal(err)
	}
	check("after an epoch Succumb")

	if err := ge.SaveGame("legacy"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("legacy"); err != nil {
		t.Fatal(err)
	}
	ge = ge2
	check("after save/load")
	if got := ge.buildResolver().AddTotal("production_all"); got < CosmicLegacyProductionBonus {
		t.Errorf("production_all after load = %v, want at least the legacy's %v (was %v)", got, CosmicLegacyProductionBonus, base)
	}
	if ge.cheaterBadge {
		t.Error("save with the Cosmic Legacy failed its signature")
	}

	// Only a full wipe clears it.
	ge.Reset()
	if ge.cosmicLegacy || hasCosmicLegacyModifier(ge) {
		t.Error("Reset kept the Cosmic Legacy")
	}
}

func TestInviteGuaranteesTheLastPassage(t *testing.T) {
	ge := lpEngine(t, "quantum_age", 8)
	ge.harbinger.BraceLevel = 2
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	if countLogs(ge, "Your next prestige will bring the Last Passage") != 1 {
		t.Errorf("invite log:\n%s", strings.Join(logMessages(ge), "\n"))
	}
	if err := ge.HarbingerAppease(); err == nil {
		t.Error("Appease after an invite should refuse")
	}
	full := ge.GetState().Prestige.PendingPoints
	ge.rng = riggedRNG(0.999) // would miss any real odds
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if !ge.pendingLastPassage || ge.catastropheInvited {
		t.Fatalf("invited: pending %v invite still armed %v", ge.pendingLastPassage, ge.catastropheInvited)
	}
	if !ge.GetState().LastPassage.Invited {
		t.Error("state does not say the Last Passage was invited")
	}
	if err := ge.EndureLastPassage(); err != nil {
		t.Fatal(err)
	}
	if indexOfLog(ge, "You invited it, and it came") < 0 {
		t.Errorf("no Fulfilled verdict:\n%s", strings.Join(logMessages(ge), "\n"))
	}
	if got, want := ge.Prestige.Snapshot().Available, int(float64(full)*0.85); got != want {
		t.Errorf("braced twice: %d points, want %d of %d", got, want, full)
	}
}

func TestPendingLastPassageBlocksOnlyPrestigeAndSurvivesSaveLoad(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := lpEngine(t, "galactic_age", 9)
	ge.harbinger.BraceLevel = 1
	makeLastPassagePending(t, ge)

	if err := ge.DoPrestige(); err == nil || !strings.Contains(err.Error(), "The Last Passage is upon you") {
		t.Fatalf("second prestige: err = %v", err)
	}
	if err := ge.HarbingerBrace(); err == nil || !strings.Contains(err.Error(), "already come") {
		t.Errorf("brace while pending: err = %v", err)
	}
	if v := ge.GetState().Harbinger; v == nil || !v.PassageCame || v.BraceBlocked == "" || v.AppeaseBlocked == "" || v.InviteBlocked == "" {
		t.Errorf("view while pending = %+v", v)
	}
	if err := ge.forceCatastrophe(); err == nil {
		t.Error("an epoch catastrophe was stacked on the pending Last Passage")
	}

	if err := ge.SaveGame("pending_lp"); err != nil {
		t.Fatal(err)
	}
	infos, err := ListSaveDetails()
	if err != nil || len(infos) != 1 || infos[0].PendingCatastrophe != "The Last Passage" {
		t.Errorf("save details = %+v, err %v", infos, err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("pending_lp"); err != nil {
		t.Fatal(err)
	}
	if ge2.cheaterBadge {
		t.Error("pending save failed its signature")
	}
	st := ge2.GetState()
	if !st.LastPassage.Pending || st.LastPassage.BraceLevel != 1 || st.PendingCatastrophe != "" {
		t.Fatalf("after load: %+v", st.LastPassage)
	}
	if err := ge2.DoPrestige(); err == nil {
		t.Error("prestige after load was not blocked")
	}

	// Advancing is not blocked; the next figure takes up the thread.
	ge2.mu.Lock()
	ge2.ageReady = true
	ge2.mu.Unlock()
	if err := ge2.AdvanceAge(); err != nil {
		t.Fatalf("AdvanceAge while the Last Passage waits: %v", err)
	}
	if ge2.age != "quantum_age" || ge2.harbinger == nil || ge2.harbinger.Age != "quantum_age" || !ge2.pendingLastPassage {
		t.Fatalf("after advancing: age %s thread %+v pending %v", ge2.age, ge2.harbinger, ge2.pendingLastPassage)
	}

	full := ge2.GetState().Prestige.PendingPoints
	if err := ge2.EndureLastPassage(); err != nil {
		t.Fatal(err)
	}
	if got := ge2.Prestige.Snapshot().Available; got != int(float64(full)*0.70) {
		t.Errorf("Endure after load: %d points, want 70%% of %d", got, full)
	}
}

func TestLastPassageIsDeterministic(t *testing.T) {
	run := func(seed int64) (bool, []string) {
		ge := lpEngine(t, "transcendent_age", seed)
		if err := ge.DoPrestige(); err != nil {
			t.Fatal(err)
		}
		came := ge.pendingLastPassage
		if came {
			if err := ge.EndureLastPassage(); err != nil {
				t.Fatal(err)
			}
		}
		return came, logMessages(ge)
	}
	came, spared := 0, 0
	for seed := int64(1); seed <= 60; seed++ {
		a, la := run(seed)
		b, lb := run(seed)
		if a != b || strings.Join(la, "\n") != strings.Join(lb, "\n") {
			t.Fatalf("seed %d: outcome or log differs between runs", seed)
		}
		if a {
			came++
		} else {
			spared++
		}
	}
	if came == 0 || spared == 0 {
		t.Errorf("60 seeds gave %d Last Passages and %d spared; want both", came, spared)
	}
}

// --- the ending lines -------------------------------------------------------------------

// runEndingCorpus is every line RunEnding can say at age with the Subject the
// engine sends (the pools are slotless, so a few thousand draws see them all).
func runEndingCorpus(age string) map[string]bool {
	req := flavor.Request{Moment: flavor.RunEnding, Age: age}
	if config.IsFinalEpoch(config.EpochForAge(age)) {
		def, _ := config.HarbingerFor(age)
		req.Subject = def.Name
	}
	rng := rand.New(rand.NewSource(1))
	out := map[string]bool{}
	for i := 0; i < 4000; i++ {
		out[flavor.Line(req, rng)] = true
	}
	return out
}

func TestRunEndingLineAtEveryPrestige(t *testing.T) {
	for _, age := range []string{"modern_age", "information_age", "cyberpunk_age", "space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"} {
		ge := catEngine(t, age, 10)
		ge.harbingerTickCheck()
		ge.rng = rand.New(&seqSource{vals: []float64{0.99, 0.41, 0.07, 0.63, 0.29, 0.85}})
		if err := ge.DoPrestige(); err != nil {
			t.Fatal(err)
		}
		if ge.pendingLastPassage {
			t.Fatalf("%s: rolled a Last Passage on 0.99", age)
		}
		done := indexOfLog(ge, "Prestige complete.")
		if done < 1 {
			t.Fatalf("%s: no line before the prestige summary:\n%s", age, strings.Join(logMessages(ge), "\n"))
		}
		ending := strings.TrimSuffix(strings.TrimPrefix(ge.log[done-1].Message, "  [gray]"), "[-]")
		corpus := runEndingCorpus(age)
		if !corpus[ending] {
			t.Errorf("%s: the line before the summary is not a RunEnding line for this age: %q", age, ending)
		}
		if ge.log[done-1].Tick != 0 {
			t.Errorf("%s: carried line kept its old tick %d", age, ge.log[done-1].Tick)
		}
		// The plain register before the Cosmic Era: no Cosmic Era line can be
		// drawn there, and no figure is named.
		if !config.IsFinalEpoch(config.EpochForAge(age)) {
			for line := range runEndingCorpus("galactic_age") {
				if corpus[line] && strings.Contains(line, "Elder Relay") {
					t.Errorf("%s reaches a Cosmic Era line: %q", age, line)
				}
			}
		}
	}
}
