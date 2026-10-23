package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// The texts the code writes itself and a save keeps as something else: the
// stories of the road, the names of timed effects, the civilization log, the
// notices a load gives an old save, and the words that were constants. They
// follow the switch where they are read, and the save holds them in English
// or as keys.

// reaching is a script that marks every text and records which rows it was
// asked for.
func reaching(reached map[string]bool) func(id, plain string) string {
	return func(id, plain string) string {
		reached[id] = true
		return mark(id, plain)
	}
}

// roadPlay is three missions that met a road event for every template of the
// table, as they were left.
type roadPlay struct {
	ge     *GameEngine
	active []*ActiveExpedition
}

// playRoad sends a civil party, a military one and a mission with no party at
// all down a road that meets, for every template of every kind, an event told
// from it (the other two templates an event holds are drawn from the pools
// by turn), and returns them. Nothing in it draws a random number.
func playRoad(t *testing.T) roadPlay {
	t.Helper()
	ge := newSeededEngine(3)
	setAge(ge, "bronze_age")
	setSoldiers(ge, 500)
	words, ok := ge.rules.MissionWords("bronze_age")
	if !ok {
		t.Fatal("no words for the Bronze Age")
	}
	ge.Military.board = MissionBoard{Age: "bronze_age", Slots: make([]boardSlot, boardSizeFor(3))}
	// Gold for two tolls in three: the third of each kind turns into foes.
	ge.Resources.Remove("gold", ge.Resources.Get("gold"))
	setIncome(ge, "gold", .1)
	toll := 2 * 60 / config.TickSeconds * ge.townIncome("gold")
	ge.Resources.Add("gold", toll*2.5)

	var events []RoadEvent
	for _, kind := range []string{"storm", "helpers", "find", "toll", "foes", "ambush", "fair", "ruins"} {
		for i := 0; i < words.RoadTemplateCount(kind); i++ {
			events = append(events, RoadEvent{
				Tick: 1, Kind: kind,
				Story:       words.RoadTemplate(kind, i),
				AuxStory:    words.RoadTemplate("foes", i+1),
				AmbushStory: words.RoadTemplate("ambush_haul", i+2),
				FoeIndex:    i, HelperIndex: i + 1, WeatherIndex: i + 2, FindIndex: i,
			})
		}
	}
	mission := func(tier, grand string, soldiers, total int) *ActiveExpedition {
		a := &ActiveExpedition{
			MissionOffer: MissionOffer{Age: "bronze_age", Place: "Tin Road", SuccessChance: .5, HaulMinutes: 20, Tier: tier, Grand: grand},
			TotalTicks:   total, TicksLeft: total / 2, Soldiers: soldiers, Party: soldiers,
			BaseRewards: map[string]float64{"wood": 10}, SuccessRewards: map[string]float64{"wood": 10}, FailureRewards: map[string]float64{"wood": 3},
			RoadEvents: append([]RoadEvent(nil), events...),
		}
		ge.Military.setActive(tier, a)
		return a
	}
	p := roadPlay{ge: ge, active: []*ActiveExpedition{
		mission(TierLong, "", 10, 100),             // spans of seconds
		mission(TierGrand, GrandMilitary, 10, 600), // and of minutes
		mission(TierShort, "", 0, 100),
	}}
	for _, a := range p.active {
		ge.fireRoadEvents(a, 1000)
	}
	return p
}

// read lists what the missions told: the stories as read, as the save holds
// them, and as the log said them.
func (p roadPlay) read() (seen []string, kept []string, logged []string) {
	for _, a := range p.active {
		seen = append(seen, a.RoadEventsSeen...)
		for _, rt := range a.RoadTold {
			kept = append(kept, rt.Text)
		}
	}
	for _, l := range p.ge.GetLogs() {
		if l.Type == "event" {
			logged = append(logged, l.Message)
		}
	}
	return
}

// TestRoadStoriesFollowTheSwitchAndTheSaveKeepsEnglish: a road event's story
// is told from a template of the table, kept in English with its numbers (a
// save never holds the script's words), and told again in the script's words
// whenever it is read. With VibeMode off it is what it always was, whatever
// the script holds.
func TestRoadStoriesFollowTheSwitchAndTheSaveKeepsEnglish(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	vibeOff(t)
	defer vibe.Set(false)

	plain := playRoad(t)
	wantSeen, wantKept, wantLogged := plain.read()
	if len(wantSeen) != 72 || len(wantKept) != len(wantSeen) {
		t.Fatalf("the road met %d stories and kept %d, want a story for each template on three missions", len(wantSeen), len(wantKept))
	}
	for _, s := range append(append([]string(nil), wantSeen...), wantKept...) {
		if strings.Contains(s, markOpen) {
			t.Fatalf("a plain story has the script's mark in it: %q", s)
		}
	}

	// Off, with a script that rewrites everything: nothing changes.
	restore := vibe.UseRewriterForTest(mark)
	again := playRoad(t)
	restore()
	seen, kept, logged := again.read()
	sameTexts(t, "off, a script that rewrites every text: the stories as read", seen, wantSeen)
	sameTexts(t, "off, a script that rewrites every text: the stories as kept", kept, wantKept)
	sameTexts(t, "off, a script that rewrites every text: the log", logged, wantLogged)

	// On, and the script has no line for a story: the templates read as they
	// always did (the words of the place are the script's, in the ruleset's
	// own rows), and what is kept is what is kept with VibeMode off.
	vibe.Set(true)
	restore = vibe.UseRewriterForTest(func(id, plain string) string { return plain })
	same := playRoad(t)
	restore()
	baseSeen, kept, baseLogged := same.read()
	sameTexts(t, "on, no line in the script: the stories as kept", kept, wantKept)
	if len(baseSeen) != len(wantSeen) {
		t.Fatalf("on, with no line in the script, the road met %d stories, off %d", len(baseSeen), len(wantSeen))
	}

	// On, and the script marks every text: the stories are told in it, the
	// numbers and the words of the place are where they were, and what is kept
	// is the English still.
	reached := map[string]bool{}
	restore = vibe.UseRewriterForTest(reaching(reached))
	defer restore()
	marked := playRoad(t)
	seen, kept, logged = marked.read()
	if len(seen) != len(baseSeen) || len(logged) != len(baseLogged) {
		t.Fatalf("under the script the road met %d stories and %d lines, with no line in the script %d and %d", len(seen), len(logged), len(baseSeen), len(baseLogged))
	}
	for i := range seen {
		if !strings.Contains(seen[i], markOpen) {
			t.Fatalf("story %d was not told in the script: %q", i, seen[i])
		}
		if got := unmark.Replace(seen[i]); got != baseSeen[i] {
			t.Fatalf("story %d, marks taken out, reads\n  %q\nwith no line in the script\n  %q", i, got, baseSeen[i])
		}
	}
	for i := range logged {
		if got := unmark.Replace(logged[i]); got != baseLogged[i] {
			t.Fatalf("log line %d, marks taken out, reads\n  %q\nwith no line in the script\n  %q", i, got, baseLogged[i])
		}
	}
	sameTexts(t, "on, a script that marks every text: the stories as kept", kept, wantKept)
	for _, row := range []string{"game.road_events.expRoadDuration.01", "game.road_events.expRoadDuration.02"} {
		if !reached[row] {
			t.Errorf("the span of time %s was never told in the script", row)
		}
	}
	n := 0
	for id := range reached {
		if strings.HasPrefix(id, "config.mission_words.roadEventTemplates.") {
			n++
		}
	}
	if n != 27 {
		t.Errorf("%d of the 27 road templates were told in the script", n)
	}

	// The save holds the English, the keys and none of the script.
	saved := marked.ge.Military.ActiveForSave()
	js, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(js), markOpen) {
		t.Error("the save of the missions holds a text of the script")
	}
	for _, s := range saved {
		if len(s.Mission.RoadTold) == 0 || len(s.Mission.RoadEventsSeen) != 0 {
			t.Errorf("a saved mission holds %d told stories and %d read ones", len(s.Mission.RoadTold), len(s.Mission.RoadEventsSeen))
		}
	}

	// Reading them again follows the switch: with it off the templates are
	// plain, with it on they are the script's.
	vibe.Set(false)
	marked.ge.Military.redraw()
	seen, _, _ = marked.read()
	sameTexts(t, "redrawn with VibeMode off", seen, baseSeen)
	vibe.Set(true)
	marked.ge.Military.redraw()
	seen, _, _ = marked.read()
	for i := range seen {
		if !strings.Contains(seen[i], markOpen) || unmark.Replace(seen[i]) != baseSeen[i] {
			t.Fatalf("story %d redrawn with VibeMode on reads %q", i, seen[i])
		}
	}

	// A mission loaded from the save reads in the script too.
	loaded := NewMilitaryManagerWith(marked.ge.rules)
	loaded.LoadState(MilitarySave{Active: saved})
	if got := loaded.ActiveByTier(TierLong); got == nil || len(got.RoadEventsSeen) == 0 || !strings.Contains(got.RoadEventsSeen[0], markOpen) {
		t.Errorf("a mission loaded under the script reads %+v", got)
	}
}

// TestPlainRoadSpansAreTheSpansOfPlainEnglish: a span of time a story keeps
// is the one the script would have said, in English, for any length.
func TestPlainRoadSpansAreTheSpansOfPlainEnglish(t *testing.T) {
	vibeOff(t)
	for ticks := 1; ticks < 400; ticks++ {
		kept := plainRoadDuration(ticks)
		if said := sayRoadDuration(kept); said != kept {
			t.Fatalf("%d ticks are kept as %q and said as %q with VibeMode off", ticks, kept, said)
		}
	}
	if got := sayRoadDuration("a long while"); got != "a long while" {
		t.Errorf("a span that is not one is said as %q", got)
	}
}

// TestTimedEffectsAreKeptAsKeysAndNamedInTheScript: Reconstruction Effort, a
// raid, a loan and a festival are saved as the key they are named by, and
// named again from the script in force when the game starts and loads them.
func TestTimedEffectsAreKeptAsKeysAndNamedInTheScript(t *testing.T) {
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	vibeOff(t)
	defer vibe.Set(false)
	keys := map[string]string{
		"endure_reconstruction": "Reconstruction Effort",
		"war_raid":              "Under Raid",
		"worker_lending":        "Workers on Loan",
		"cultural_festival":     "Cultural Festival",
	}
	inject := func(ge *GameEngine) {
		for key := range keys {
			ge.Events.InjectEvent(ActiveEvent{Key: key, Src: nameFromCode, Ref: key, TicksLeft: 50})
		}
		ge.Events.InjectEvent(ActiveEvent{Key: "military_chain_boost", Src: nameFromChain, Ref: "military_chain", TicksLeft: 50})
	}
	names := func(ge *GameEngine) map[string]string {
		out := map[string]string{}
		for _, ae := range ge.Events.active {
			out[ae.Key] = ae.Name
		}
		return out
	}

	ge := newSeededEngine(1)
	inject(ge)
	got := names(ge)
	for key, want := range keys {
		if got[key] != want {
			t.Errorf("%s is named %q, want %q", key, got[key], want)
		}
	}
	if got["military_chain_boost"] == "" || !strings.HasSuffix(got["military_chain_boost"], " Speed Boost") {
		t.Errorf("the chain boost is named %q", got["military_chain_boost"])
	}

	// The save holds the key and no name.
	for _, ae := range ge.Events.GetActiveForSave() {
		if ae.Name != "" || ae.Src == "" || ae.Ref == "" {
			t.Errorf("%s is saved as %+v, want a Src and a Ref and no Name", ae.Key, ae)
		}
	}

	// Under a script that marks every text.
	reached := map[string]bool{}
	defer vibe.UseRewriterForTest(reaching(reached))()
	vibe.Set(true)
	marked := newSeededEngine(1)
	inject(marked)
	for key, name := range names(marked) {
		if !strings.Contains(name, markOpen) {
			t.Errorf("%s is named %q under the script, which does not mark it", key, name)
		}
		if want, ok := keys[key]; ok && unmark.Replace(name) != want {
			t.Errorf("%s is named %q under the script, marks taken out", key, name)
		}
	}
	for _, row := range []string{"game.catastrophe.startReconstruction.01", "game.defense.applyWarRaids.01", "game.engine.processDiplomacy.01", "game.engine.DoFestival.03"} {
		if !reached[row] {
			t.Errorf("the row %s was never named in the script", row)
		}
	}
	saved := marked.Events.GetActiveForSave()
	js, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(js), markOpen) {
		t.Errorf("the save of the timed effects holds a text of the script: %s", js)
	}

	// Loaded, the names are drawn from the script in force: a save written
	// under the script reads plain with VibeMode off, and an old save with
	// the name written in it keeps that name.
	vibe.Set(false)
	back := newSeededEngine(2)
	back.Events.LoadState(nil, saved, 0, 0, 0)
	for _, ae := range back.Events.active {
		if want, ok := keys[ae.Key]; ok && ae.Name != want {
			t.Errorf("%s loaded with VibeMode off is named %q, want %q", ae.Key, ae.Name, want)
		}
	}
	old := newSeededEngine(2)
	old.Events.LoadState(nil, []ActiveEvent{{Key: "war_raid", Name: "Under Raid", TicksLeft: 9}}, 0, 0, 0)
	if len(old.Events.active) != 1 || old.Events.active[0].Name != "Under Raid" {
		t.Errorf("an old save's raid loads as %+v", old.Events.active)
	}
}

// historyOf plays a run that ends in every kind of line the civilization log
// has: a doom endured, one succumbed to and one succumbed to again (the legacy
// is held), and the Last Passage endured and succumbed to.
func historyOf(t *testing.T) *GameEngine {
	t.Helper()
	ge := newSeededEngine(7)
	var struck []string
	for _, age := range config.AgeOrder()[1:] {
		if err := ge.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
		if ge.ForceCatastropheForTest() != nil {
			continue
		}
		struck = append(struck, age)
		if len(struck) == 1 {
			if err := ge.Endure(); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := ge.Succumb(); err != nil {
			t.Fatal(err)
		}
		if len(struck) == 2 {
			// The same doom again, with its legacy already held.
			if err := ge.EnterAgeForTest(age); err != nil {
				t.Fatal(err)
			}
			if err := ge.ForceCatastropheForTest(); err != nil {
				t.Fatal(err)
			}
			if err := ge.Succumb(); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	if len(struck) < 2 {
		t.Fatalf("only %d dooms struck", len(struck))
	}
	if err := ge.ForceLastPassageForTest("transcendent_age"); err != nil {
		t.Fatal(err)
	}
	if err := ge.EndureLastPassage(); err != nil {
		t.Fatal(err)
	}
	if err := ge.ForceLastPassageForTest("transcendent_age"); err != nil {
		t.Fatal(err)
	}
	if err := ge.SuccumbLastPassage(); err != nil {
		t.Fatal(err)
	}
	return ge
}

// TestTheCivilizationLogIsToldInTheScriptAndSavedInEnglish: the log keeps its
// lines in English with the markers of the names, and tells them in the
// script's words when it is read with VibeMode on. Off, it reads as it did.
func TestTheCivilizationLogIsToldInTheScriptAndSavedInEnglish(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	vibeOff(t)
	defer vibe.Set(false)

	ge := historyOf(t)
	if len(ge.catastropheHistory) < 5 {
		t.Fatalf("the log has %d lines, want at least 5: %q", len(ge.catastropheHistory), ge.catastropheHistory)
	}
	plain := ge.historyLines()
	var shapes = map[string]bool{}
	for _, line := range plain {
		if strings.Contains(line, "{") || strings.Contains(line, markOpen) {
			t.Fatalf("a line of the log is read with a marker or a mark in it: %q", line)
		}
		for _, shape := range []string{"buildings lost.", "Legacy bonus earned.", "Legacy bonus already held.", "points.", "Cosmic Legacy earned."} {
			if strings.HasSuffix(line, shape) {
				shapes[shape] = true
			}
		}
	}
	if len(shapes) != 5 {
		t.Errorf("the run's log has only these shapes of line: %v\n%q", shapes, plain)
	}

	// On, and the script has no line: the same lines.
	vibe.Set(true)
	restore := vibe.UseRewriterForTest(func(id, plain string) string { return plain })
	sameTexts(t, "on, no line in the script", ge.historyLines(), plain)
	restore()

	// On, and the script marks every text: every line is told in it, with the
	// names and the numbers where they were.
	reached := map[string]bool{}
	restore = vibe.UseRewriterForTest(reaching(reached))
	told := ge.historyLines()
	restore()
	if len(told) != len(plain) {
		t.Fatalf("the log has %d lines under the script and %d in plain English", len(told), len(plain))
	}
	for i := range told {
		if !strings.Contains(told[i], markOpen) {
			t.Errorf("line %d of the log was not told in the script: %q", i, told[i])
		}
		if got := unmark.Replace(told[i]); got != plain[i] {
			t.Errorf("line %d, marks taken out, reads\n  %q\nin plain English\n  %q", i, got, plain[i])
		}
	}
	for i := 1; i <= 6; i++ {
		if row := "game.catastrophe.historyLines.0" + string(rune('0'+i)); !reached[row] {
			t.Errorf("the row %s was never told in the script", row)
		}
	}

	// What is kept is English with markers, whatever was read.
	vibe.Set(false)
	for _, line := range ge.catastropheHistory {
		if strings.Contains(line, markOpen) || !strings.Contains(line, "{") {
			t.Errorf("the kept line is %q", line)
		}
	}
	if endured, succumbed := countCatastropheOutcomes(ge.catastropheHistory); endured != 2 || succumbed != 3 {
		t.Errorf("the log counts %d endured and %d succumbed, want 2 and 3", endured, succumbed)
	}
}

// TestTheNoticesAndTheNamedWordsFollowTheSwitch: the lines an old save is
// given on its first load, the word a rare offer's title begins with, the old
// cache's flavor, the refusal of the Last Passage and the ruins' description
// are written by functions that read the script, and plain English with
// VibeMode off.
func TestTheNoticesAndTheNamedWordsFollowTheSwitch(t *testing.T) {
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	vibeOff(t)
	defer vibe.Set(false)
	texts := map[string]func() string{
		"pacing":      pacingNotice,
		"market":      marketSwapNotice,
		"scholars":    scholarsNotice,
		"storage":     storageWallNotice,
		"rare":        config.RareTitleWord,
		"cache":       ancientMemoryFlavor,
		"last":        lastPassageCame,
		"ruins":       ruinsFound,
		"ruins plain": func() string { return ruinsFoundPlain },
	}
	plain := map[string]string{}
	for name, f := range texts {
		if plain[name] = f(); plain[name] == "" || strings.Contains(plain[name], markOpen) {
			t.Fatalf("%s reads %q with VibeMode off", name, plain[name])
		}
	}
	if plain["rare"] != "Legendary" || plain["ruins"] != plain["ruins plain"] {
		t.Errorf("the rare word is %q, the ruins' line %q and its kept form %q", plain["rare"], plain["ruins"], plain["ruins plain"])
	}

	defer vibe.UseRewriterForTest(mark)()
	for name, f := range texts {
		if got := f(); got != plain[name] {
			t.Errorf("%s reads %q with VibeMode off and a script that marks every text, want %q", name, got, plain[name])
		}
	}
	vibe.Set(true)
	for name, f := range texts {
		got := f()
		if name == "ruins plain" {
			if got != plain[name] {
				t.Errorf("a ruins description kept as written reads %q under the script", got)
			}
			continue
		}
		if !strings.Contains(got, markOpen) || unmark.Replace(got) != plain[name] {
			t.Errorf("%s reads %q under the script, want %q marked", name, got, plain[name])
		}
	}

	// A rare offer's title begins with the script's word.
	words, _ := rules.Core().MissionWords("bronze_age")
	offer := MissionOffer{Age: "bronze_age", Tier: TierShort, Rare: true, Words: OfferWords{Made: true, Found: 1}}
	vibe.Set(false)
	offer.draw(words)
	title, desc := offer.Title, offer.Description
	vibe.Set(true)
	offer.draw(words)
	if !strings.Contains(offer.Title, markOpen) || unmark.Replace(offer.Title) != title || !strings.HasPrefix(unmark.Replace(offer.Title), "Legendary ") {
		t.Errorf("a rare offer's title reads %q under the script, %q in plain English", offer.Title, title)
	}
	if !strings.Contains(offer.Description, markOpen) || unmark.Replace(offer.Description) != desc || !strings.HasPrefix(desc, plain["ruins"]) {
		t.Errorf("a found offer's description reads %q under the script, %q in plain English", offer.Description, desc)
	}
	if forSave := offer.forSave(); forSave.Title != "" || forSave.Description != "" {
		t.Errorf("a drawn offer is saved with its words: %+v", forSave)
	}
}

// TestAnOldSaveGetsItsNoticeInTheScript: a save from before the one-week
// curve says so once on load, in the script's words while VibeMode is on, and
// the save it writes next holds nothing of the script.
func TestAnOldSaveGetsItsNoticeInTheScript(t *testing.T) {
	isolateAccountDir(t)
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	vibeOff(t)
	defer vibe.Set(false)

	ge := newSeededEngine(1)
	ge.StepTicks(20)
	ge.mu.RLock()
	old := ge.buildSaveSnapshot()
	ge.mu.RUnlock()
	old.PacingWeek = false
	old.Signature = signSave(old, saveHMACKey)
	data, err := json.MarshalIndent(old, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(savePath("old-curve"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	plainNotice := pacingNotice()
	defer vibe.UseRewriterForTest(mark)()
	vibe.Set(true)
	loaded := NewGameEngine()
	if err := loaded.LoadGame("old-curve"); err != nil {
		t.Fatal(err)
	}
	if n := countLines(loaded, mark("game.save.pacingNotice.01", plainNotice)); n != 1 {
		t.Errorf("the old save's notice was told in the script %d times, want once", n)
	}
	if n := countLines(loaded, plainNotice); n != 0 {
		t.Errorf("the old save's notice was told in plain English %d times under the script", n)
	}
	if err := loaded.SaveGame("old-curve-next"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(savePath("old-curve-next"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), markOpen) {
		t.Error("the save holds a text of the script")
	}
}
