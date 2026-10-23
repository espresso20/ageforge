package game

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// The log lines and the refusals are written by the code, not read from a
// ruleset: vibe.Line, Sprintf and Errorf swap them for the script's line while
// VibeMode is on. These tests play one run, with refusals asked for at every
// stage, and read every text the player would have read.

// playForText plays a deterministic run on set and returns the text of it in
// order: every log line (as the log held it after each stage) and every
// refusal, each with the stage it came at. The engine is returned as the run
// left it.
func playForText(t *testing.T, set *rules.Set) ([]string, *GameEngine) {
	t.Helper()
	texts, _, ge := playRun(t, set)
	return texts, ge
}

// playRun is playForText that also returns the log entries it read, in the
// order the texts have them (one entry for each "log" text).
func playRun(t *testing.T, set *rules.Set) ([]string, []LogEntry, *GameEngine) {
	t.Helper()
	var texts []string
	var entries []LogEntry
	rec := func(format string, args ...any) { texts = append(texts, fmt.Sprintf(format, args...)) }
	ge := NewGameEngineWith(set)
	ge.SeedRNG(7)
	grab := func(label string) {
		for _, l := range ge.GetLogs() {
			rec("%s|log|%s|%s", label, l.Type, l.Message)
			entries = append(entries, l)
		}
		// What else the player reads and the code writes itself: the names
		// of the timed effects under way and the civilization log.
		for _, ae := range ge.Events.active {
			rec("%s|timed|%s|%s", label, ae.Key, ae.Name)
		}
		for _, h := range ge.historyLines() {
			rec("%s|history|%s", label, h)
		}
		for _, a := range ge.Military.active {
			for _, story := range a.RoadEventsSeen {
				rec("%s|road|%s", label, story)
			}
		}
	}
	refuse := func(label string) {
		errs := []struct {
			name string
			do   func() error
		}{
			{"build0", func() error { _, e := ge.BuildMultiple("hut", 0); return e }},
			{"buildunknown", func() error { _, e := ge.BuildMultiple("hutt", 1); return e }},
			{"buildnope", func() error { _, e := ge.BuildMultiple("zzz", 1); return e }},
			{"buildmany", func() error { _, e := ge.BuildMultiple("hut", 100000); return e }},
			{"buildwonder", func() error { _, e := ge.BuildMultiple("great_pyramid", 1); return e }},
			{"recruit0", func() error { return ge.RecruitWorker("worker", 0) }},
			{"recruitbig", func() error { return ge.RecruitWorker("worker", 1000000) }},
			{"recruitodd", func() error { return ge.RecruitWorker("nonesuch", 1) }},
			{"assign0", func() error { return ge.AssignWorker("hut", 0) }},
			{"assignbig", func() error { return ge.AssignWorker("hut", 100000) }},
			{"assignnone", func() error { return ge.AssignWorker("zzz", 1) }},
			{"assignall", func() error { _, e := ge.AssignAll("zzz"); return e }},
			{"assignallhut", func() error { _, e := ge.AssignAll("hut"); return e }},
			{"unassign0", func() error { return ge.UnassignWorker("hut", 0) }},
			{"unassignbig", func() error { return ge.UnassignWorker("hut", 1000) }},
			{"unassignall", func() error { _, e := ge.UnassignAll("hut"); return e }},
			{"dismiss", func() error { return ge.DismissWorkers("hut", 5, false) }},
			{"sell0", func() error { return ge.SellBuilding("hut", 0) }},
			{"sellhut", func() error { return ge.SellBuilding("hut", 1) }},
			{"sellnone", func() error { return ge.SellBuilding("zzz", 1) }},
			{"research", func() error { return ge.StartResearch("zzz") }},
			{"researchfar", func() error { return ge.StartResearch("quantum_computing") }},
			{"cancelresearch", func() error { return ge.CancelResearch() }},
			{"gather", func() error { _, e := ge.GatherResource("gold", 5); return e }},
			{"gatherneg", func() error { _, e := ge.GatherResource("food", -1); return e }},
			{"festival", func() error { return ge.DoFestival() }},
			{"blackmarket", func() error { _, _, e := ge.DoBlackMarket("zzz"); return e }},
			{"blackmarket2", func() error { _, _, e := ge.DoBlackMarket("gold"); return e }},
			{"advance", func() error { return ge.AdvanceAge() }},
			{"prestige", func() error { return ge.DoPrestige() }},
			{"upgrade", func() error { return ge.UpgradeBuilding("hut", 1, false) }},
			{"upgradeall", func() error { return ge.UpgradeBuilding("zzz", 1, true) }},
			{"route", func() error { return ge.StartTradeRoute("zzz") }},
			{"routestop", func() error { return ge.StopTradeRoute("zzz") }},
			{"status", func() error { return ge.SetDiplomaticStatus("zzz", "allied") }},
			{"statusbad", func() error { return ge.SetDiplomaticStatus("zzz", "bogus") }},
			{"gift", func() error { return ge.SendGift("zzz") }},
			{"tribute", func() error { return ge.SendTribute("zzz") }},
			{"deal", func() error { _, e := ge.AcceptFactionDeal("zzz", 1); return e }},
			{"dealnone", func() error { _, e := ge.AcceptFactionDeal("", 0); return e }},
			{"expedition", func() error { return ge.LaunchExpedition(99) }},
			{"expedition0", func() error { return ge.LaunchExpedition(0) }},
			{"memory", func() error { return ge.AcceptAncientMemory() }},
			{"endure", func() error { return ge.Endure() }},
			{"succumb", func() error { return ge.Succumb() }},
			{"endurelp", func() error { return ge.EndureLastPassage() }},
			{"succumblp", func() error { return ge.SuccumbLastPassage() }},
			{"appease", func() error { return ge.HarbingerAppease() }},
			{"brace", func() error { return ge.HarbingerBrace() }},
			{"invite", func() error { return ge.HarbingerInvite() }},
			{"planupgrade", func() error { _, e := ge.PlanAddUpgrade("hut", 1); return e }},
			{"planupgrade0", func() error { _, e := ge.PlanAddUpgrade("zzz", 0); return e }},
			{"planbuild", func() error { _, e := ge.PlanAddBuild("zzz", 1); return e }},
			{"planbuildhut", func() error { _, e := ge.PlanAddBuild("hutt", 1); return e }},
			{"planremove", func() error { _, e := ge.PlanRemove(99); return e }},
			{"planresearch", func() error { _, e := ge.PlanAddResearchChain("zzz"); return e }},
			{"planadvance", func() error { return ge.PlanAddAdvance() }},
			{"share", func() error { _, e := ge.SetWorkerShare("zzz", 50); return e }},
			{"share2", func() error { _, e := ge.SetWorkerShare("food", 500); return e }},
			{"scholarshare", func() error { _, e := ge.SetResearchShare(500); return e }},
			{"scholarshare2", func() error { _, e := ge.SetResearchShare(50); return e }},
			{"switch", func() error { _, e := ge.SwitchAccount("zzz"); return e }},
			{"branch", func() error { return ge.BranchSave("") }},
			{"load", func() error { return ge.LoadGame("zzz-nosuch") }},
			{"bank", func() error { _, e := ge.BankWonderResource("zzz", "gold", 5); return e }},
			{"bank2", func() error { _, e := ge.BankWonderResource("great_pyramid", "gold", 5); return e }},
		}
		for _, e := range errs {
			func() {
				defer func() {
					if r := recover(); r != nil {
						rec("%s|refuse|%s|PANIC %v", label, e.name, r)
					}
				}()
				err := e.do()
				if err == nil {
					rec("%s|refuse|%s|<nil>", label, e.name)
				} else {
					rec("%s|refuse|%s|%s", label, e.name, err.Error())
				}
			}()
		}
		grab(label + "-after-refusals")
	}
	refuse("start")
	for i, age := range config.AgeOrder()[1:] {
		if err := ge.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
		ge.GrantTechsForTest()
		for _, f := range rules.CoreFor(false).Factions() {
			_ = ge.MeetFactionForTest(f.Key, 40)
		}
		_ = ge.SummonHarbingerForTest(age)
		if i%2 == 0 {
			sendRoadMission(t, ge, i)
		}
		ge.StepTicks(350)
		grab(age + "-a")
		if i%4 == 0 {
			refuse(age)
		}
		if i%3 == 0 && ge.ForceCatastropheForTest() == nil {
			grab(age + "-doom")
			if i%2 == 0 {
				_ = ge.Endure()
			} else {
				_ = ge.Succumb()
			}
			grab(age + "-doomed")
		}
		ge.StepTicks(350)
		grab(age + "-b")
	}
	steps := []struct {
		name string
		do   func() error
	}{
		{"enter the quantum age", func() error { return ge.EnterAgeForTest("quantum_age") }},
		{"a harbinger", func() error { return ge.SummonHarbingerForTest("quantum_age") }},
		{"step", func() error { ge.StepTicks(300); return nil }},
		{"clear the doom", func() error { _ = ge.Endure(); return nil }},
		{"the Last Passage", func() error { return ge.ForceLastPassageForTest("transcendent_age") }},
		{"step", func() error { ge.StepTicks(100); return nil }},
		{"endure it", ge.EndureLastPassage},
		{"step", func() error { ge.StepTicks(300); return nil }},
		{"an earlier age", func() error { return ge.EnterAgeForTest(config.AgeOrder()[8]) }},
		{"step", func() error { ge.StepTicks(300); return nil }},
	}
	for _, st := range steps {
		if err := st.do(); err != nil {
			t.Fatalf("%s: %v", st.name, err)
		}
		grab("end:" + st.name)
	}
	refuse("end")
	_ = ge.Endure()
	if err := ge.DoPrestige(); err != nil {
		rec("prestige|refuse|%s", err.Error())
	}
	grab("prestige")
	ge.StepTicks(300)
	grab("after-prestige")
	return texts, entries, ge
}

// sendRoadMission sends a short mission that meets a road event of every kind,
// each told from a template of the table, so the run reads every story in the
// words of the script in force. The templates are the ones the event drew
// (a mission is launched and rolls its own events; they are replaced here by
// events at known ticks).
func sendRoadMission(t *testing.T, ge *GameEngine, turn int) {
	t.Helper()
	setSoldiers(ge, 500)
	n := plantOffer(ge, TierShort, "")
	if err := ge.LaunchExpedition(n); err != nil {
		return // the mission could not be paid: the run goes on without it
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	a := ge.Military.ActiveByTier(TierShort)
	if a == nil {
		return
	}
	words, _ := ge.rules.MissionWords(a.Age)
	a.RoadEvents = a.RoadEvents[:0]
	for j, kind := range []string{"storm", "helpers", "find", "toll", "foes", "ambush", "fair", "ruins", "toll", "ambush"} {
		k := (turn + j) % words.RoadTemplateCount(kind)
		a.RoadEvents = append(a.RoadEvents, RoadEvent{
			Tick: 2 + j, Kind: kind,
			Story:       words.RoadTemplate(kind, k),
			AuxStory:    words.RoadTemplate("foes", (k+1)%3),
			AmbushStory: words.RoadTemplate("ambush_haul", (k+2)%3),
			FoeIndex:    j % 3, HelperIndex: (j + 1) % 3, WeatherIndex: (j + 2) % 3, FindIndex: j % 3,
		})
	}
	a.TotalTicks, a.TicksLeft = max(a.TotalTicks, 60), max(a.TicksLeft, 60)
}

// playWith plays the run with VibeMode set to on and, when rewrite is not
// nil, with every text the code writes itself read through it.
func playWith(t *testing.T, on bool, rewrite func(id, plain string) string) ([]string, *GameEngine) {
	t.Helper()
	vibeOff(t)
	if rewrite != nil {
		defer vibe.UseRewriterForTest(rewrite)()
	}
	vibe.Set(on)
	defer vibe.Set(false)
	return playForText(t, rules.CoreFor(false))
}

// mark is a script that changes every text of the code and keeps what is in
// it: the verbs, the slots and the words. Taking the marks out gives the text
// back. A text that begins with a letter keeps it ahead of the mark, so the
// code that capitalizes a line it is given still finds a letter to capitalize.
func mark(id, plain string) string {
	if r, n := utf8.DecodeRuneInString(plain); n > 0 && unicode.IsLetter(r) {
		return plain[:n] + markOpen + plain[n:] + markClose
	}
	return markOpen + plain + markClose
}

// The marks mark runs the script has written. No text of the game has them.
const (
	markOpen  = "⟦"
	markClose = "⟧"
)

var unmark = strings.NewReplacer(markOpen, "", markClose, "")

// idShaped matches a row id: dotted words that begin with a package.
var idShaped = regexp.MustCompile(`\b(game|config|rules|pkg)\.[a-z_]+\.[A-Za-z_]+\.[0-9]{2}\b`)

func sameTexts(t *testing.T, what string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: %d texts, want %d", what, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s: text %d reads\n  %q\nwant\n  %q", what, i, got[i], want[i])
			return
		}
	}
}

// TestLogTextIsPlainWhileVibeModeIsOff: with VibeMode off, the log lines and
// the refusals of a whole run read as they always did, whatever the script
// holds: a script that rewrites every text does not reach them.
func TestLogTextIsPlainWhileVibeModeIsOff(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	plain, _ := playWith(t, false, nil)
	if len(plain) < 5000 {
		t.Fatalf("the run read only %d texts", len(plain))
	}
	scripted, _ := playWith(t, false, mark)
	sameTexts(t, "VibeMode off, a script that rewrites every text", scripted, plain)
	for _, s := range plain {
		if strings.Contains(s, markOpen) || strings.Contains(s, markClose) {
			t.Fatalf("a plain text has the script's mark in it: %q", s)
		}
	}
}

// TestLogTextFallsBackToPlain: with VibeMode on and no line in the script for
// a text, it reads plain. It never reads as an id, and never as nothing.
func TestLogTextFallsBackToPlain(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	// The run in the words of the shipped script for everything the ruleset
	// writes, and the code's own texts as they are written.
	want, _ := playWith(t, true, func(id, plain string) string { return plain })
	for _, script := range []map[string]string{
		{}, // no row at all
		{"building.hut.name": "Crash Pad", "game.engine.nonesuch.01": "A row no text has"}, // rows for other things
	} {
		restore := vibe.UseScriptForTest(script)
		got, _ := playWith(t, true, nil)
		restore()
		sameTexts(t, fmt.Sprintf("VibeMode on, a script of %d rows with no line for the texts", len(script)), got, want)
		for _, s := range got {
			parts := strings.Split(s, "|")
			if strings.TrimSpace(parts[len(parts)-1]) == "" || idShaped.MatchString(s) {
				t.Fatalf("a text reads as an id or as nothing: %q", s)
			}
		}
	}
}

// TestLogTextFollowsTheScriptAndTheStateDoesNot: with VibeMode on, a script
// that rewrites every text of the code changes the words of the log and of the
// refusals and nothing else. Take the script's marks out and the run reads as
// it does in plain English, line for line; the state and the save are the
// run's own, to the last bit, with no text of the script in either.
func TestLogTextFollowsTheScriptAndTheStateDoesNot(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	plain, plainGE := playWith(t, true, func(id, plain string) string { return plain })
	marked, markedGE := playWith(t, true, mark)
	if len(marked) != len(plain) {
		t.Fatalf("the run under the script read %d texts, in plain English %d", len(marked), len(plain))
	}
	changed := map[string]bool{}
	for i := range plain {
		if marked[i] != plain[i] {
			changed[marked[i]] = true
		}
		if got := unmark.Replace(marked[i]); got != plain[i] {
			t.Fatalf("text %d under the script, marks taken out, reads\n  %q\nin plain English\n  %q", i, got, plain[i])
		}
	}
	if len(changed) < 400 {
		t.Errorf("the script reached only %d different texts of the run", len(changed))
	}
	// The things the code names itself and the save keeps as keys or in
	// English are told in the script's words too, and every kind was met.
	for _, what := range []struct {
		kind string
		min  int
	}{{"|road|", 10}, {"|timed|", 5}, {"|history|", 5}} {
		told := map[string]bool{}
		for i := range plain {
			if marked[i] != plain[i] && strings.Contains(plain[i], what.kind) {
				told[marked[i]] = true
			}
		}
		t.Logf("%d different texts of the kind %s were told in the script's words", len(told), what.kind)
		if len(told) < what.min {
			t.Errorf("the script told only %d different texts of the kind %s, want at least %d", len(told), what.kind, what.min)
		}
	}

	a, b := stateMap(t, plainGE), stateMap(t, markedGE)
	for k := range a {
		if !reflect.DeepEqual(a[k], b[k]) {
			t.Errorf("under the script, %q is not what it is in plain English: %.300v vs %.300v", k, a[k], b[k])
		}
	}
	if plainGE.StateDigest() != markedGE.StateDigest() {
		t.Error("the state digests differ")
	}
	js, err := markedGE.StateJSON()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(js), markOpen) {
		t.Error("the state holds a text of the script")
	}
	if strings.Contains(string(js), `"Tags"`) {
		t.Error("the state holds the tags of a log line")
	}
	vibe.Set(true)
	defer vibe.Set(false)
	defer vibe.UseRewriterForTest(mark)()
	if err := markedGE.SaveGame("marked"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(savePath("marked"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), markOpen) {
		t.Error("the save holds a text of the script")
	}
	if strings.Contains(string(raw), `"Tags"`) || strings.Contains(string(raw), `"log"`) {
		t.Error("the save holds the log or the tags of its lines")
	}
}

// TestALineThatStartsWithASymbolIsNeverCut: a script's line may begin with
// anything (an emoji, an accented letter), and the code that works on the
// first letter of a line it is handed (capitalizing it, lower-casing it) must
// keep it whole. Every text of a run under a script that puts a symbol first
// is valid text, with nothing cut out of it.
func TestALineThatStartsWithASymbolIsNeverCut(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	texts, _ := playWith(t, true, func(id, plain string) string { return "💀 " + plain })
	reached := 0
	for _, s := range texts {
		if !utf8.ValidString(s) || strings.ContainsRune(s, utf8.RuneError) {
			t.Fatalf("a text has a letter cut out of it: %q", s)
		}
		if strings.Contains(s, "💀") {
			reached++
		}
	}
	if reached < 400 {
		t.Errorf("the script reached only %d texts", reached)
	}
}

// TestALogLineFollowsTheSwitch: one line and one refusal, end to end. They
// are written in the words of the script while VibeMode is on and the script
// has a line for them, in plain English when it is off, and the arguments go
// into the script's format.
func TestALogLineFollowsTheSwitch(t *testing.T) {
	vibeOff(t)
	gathered := func() string {
		ge := NewGameEngine()
		if _, err := ge.GatherResource("food", 1); err != nil {
			t.Fatal(err)
		}
		logs := ge.GetLogs()
		return logs[len(logs)-1].Message
	}
	refused := func() string {
		_, err := NewGameEngine().BuildMultiple("hut", 0)
		if err == nil {
			t.Fatal("a build of none was not refused")
		}
		return err.Error()
	}
	plainLine, plainRefusal := gathered(), refused()
	if !strings.HasPrefix(plainLine, "Gathered ") || plainRefusal != "Build count must be positive (got 0)." {
		t.Fatalf("plain English reads %q and %q", plainLine, plainRefusal)
	}
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	defer vibe.UseScriptForTest(map[string]string{
		"game.engine.GatherResource.04": "Snagged %s, no cap.",
		"game.engine.BuildMultiple.01":  "Zero builds? Bet. (got %d)",
	})()
	vibe.Set(true)
	defer vibe.Set(false)
	// What the amount reads as is the ruleset's word for it (the script
	// has one for food); the rest of the line is the script's.
	middle := strings.TrimSuffix(strings.TrimPrefix(func() string {
		defer vibe.UseRewriterForTest(func(id, plain string) string { return plain })()
		return gathered()
	}(), "Gathered "), ".")
	if got := gathered(); got != "Snagged "+middle+", no cap." {
		t.Errorf("with VibeMode on, the line reads %q, want %q", got, "Snagged "+middle+", no cap.")
	}
	if got := refused(); got != "Zero builds? Bet. (got 0)" {
		t.Errorf("with VibeMode on, the refusal reads %q", got)
	}
	vibe.Set(false)
	if gathered() != plainLine || refused() != plainRefusal {
		t.Error("with VibeMode off again, the text is not plain English")
	}
}

// TestAScriptDoesNotMoveALineOutOfAPanel: what a line is about is the tags it
// was written with, so a script that rewrites every text of the code leaves
// every line in the panel it was in. Played plain and played under a script
// that marks every text, a run writes the same lines with the same tags, a
// panel that asks a line what it is about picks the same lines in both, and
// a line under the script is read by its tags and never by its words.
func TestAScriptDoesNotMoveALineOutOfAPanel(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	vibeOff(t)
	run := func(on bool, rewrite func(id, plain string) string) []LogEntry {
		if rewrite != nil {
			defer vibe.UseRewriterForTest(rewrite)()
		}
		vibe.Set(on)
		defer vibe.Set(false)
		_, entries, _ := playRun(t, rules.CoreFor(false))
		return entries
	}
	plain := run(false, nil)
	marked := run(true, mark)
	if len(marked) != len(plain) {
		t.Fatalf("the run under the script wrote %d lines, in plain English %d", len(marked), len(plain))
	}

	subjects := []struct {
		name string
		tag  LogTag
		min  int
	}{{"people", TagPeople, 100}, {"army", TagArmy, 5}, {"bad news", TagBad, 100}, {"good news", TagGood, 20}}
	picked := map[LogTag]int{}
	for i := range plain {
		if plain[i].Tags&TagPlain == 0 {
			t.Fatalf("line %d was written in plain English and is not marked so: %+v", i, plain[i])
		}
		if marked[i].Tags&TagPlain != 0 {
			t.Fatalf("line %d was written under a script and is marked plain: %+v", i, marked[i])
		}
		if got, want := marked[i].Tags, plain[i].Tags&^TagPlain; got != want {
			t.Fatalf("line %d has tags %b under the script and %b in plain English: %q", i, got, want, plain[i].Message)
		}
		for _, sub := range subjects {
			// The panel under the script picks what the plain run's tags pick.
			want := plain[i].Tags&sub.tag != 0
			if got := marked[i].About(sub.tag); got != want {
				t.Fatalf("line %d: the %s panel picks it %v under the script and %v in plain English: %q", i, sub.name, got, want, plain[i].Message)
			}
			if want && plain[i].Type != "debug" {
				picked[sub.tag]++
			}
		}
	}
	for _, sub := range subjects {
		t.Logf("%d lines of the run are about %s", picked[sub.tag], sub.name)
		if picked[sub.tag] < sub.min {
			t.Errorf("only %d lines of the run are about %s, want at least %d", picked[sub.tag], sub.name, sub.min)
		}
	}
}

// TestALineIsReadByItsWordsOnlyInPlainEnglish: a line written in plain English
// is picked by its words as it always was, with no tag needed; a line written
// under a script is picked by its tags, whatever its words say.
func TestALineIsReadByItsWordsOnlyInPlainEnglish(t *testing.T) {
	plain := LogEntry{Message: "Recruited 3 workers (population 4/9).", Tags: TagPlain}
	if !plain.About(TagPeople) || !plain.About(TagGood) || plain.About(TagArmy) || plain.About(TagBad) {
		t.Errorf("a plain line is read wrong: people %v good %v army %v bad %v", plain.About(TagPeople), plain.About(TagGood), plain.About(TagArmy), plain.About(TagBad))
	}
	quiet := LogEntry{Message: "Built a Hut (you have 4).", Tags: TagPlain | TagPeople}
	if quiet.About(TagPeople) {
		t.Error("a plain line is picked by a tag its words do not carry: plain lines are read by their words")
	}
	// The same line in a script's words, with its tags.
	slang := LogEntry{Message: "Squad of 3 pulled up, no cap.", Tags: TagPeople | TagGood}
	if !slang.About(TagPeople) || !slang.About(TagGood) || slang.About(TagArmy) || slang.About(TagBad) {
		t.Error("a line under a script is not read by its tags")
	}
	// Words that read like a loss do not make a line under a script one.
	trap := LogEntry{Message: "The soldiers lost the plot.", Tags: TagPeople}
	if trap.About(TagArmy) || trap.About(TagBad) {
		t.Error("a line under a script is read by its words")
	}
}

// TestTheLinesOfTheCommandsAreTagged: a command that recruits, assigns,
// unassigns or dismisses writes a line about the people, in whatever script;
// a loss is bad news, and an arrival good.
func TestTheLinesOfTheCommandsAreTagged(t *testing.T) {
	vibeOff(t)
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	defer vibe.Set(false)
	for _, on := range []bool{false, true} {
		vibe.Set(on)
		ge := newSeededEngine(1)
		ge.Workers.UnlockType("worker")
		ge.Buildings.counts["gathering_camp"] = 2
		ge.Buildings.counts["hut"] = 2
		ge.recalculateRates()
		last := func() LogEntry { logs := ge.GetLogs(); return logs[len(logs)-1] }
		if err := ge.RecruitWorker("worker", 1); err != nil {
			t.Fatalf("VibeMode %v: a recruit was refused: %v", on, err)
		}
		e := last()
		if e.Tags&(TagPeople|TagGood) != TagPeople|TagGood {
			t.Errorf("VibeMode %v: a recruit is written %+v", on, e)
		}
		if marked := e.Tags&TagPlain != 0; marked == on {
			t.Errorf("VibeMode %v: the line is marked plain %v", on, marked)
		}
		if err := ge.AssignWorker("gathering_camp", 1); err != nil {
			t.Fatalf("VibeMode %v: an assign was refused: %v", on, err)
		}
		if e := last(); e.Tags&TagPeople == 0 {
			t.Errorf("VibeMode %v: an assign is written %+v", on, e)
		}
		if err := ge.UnassignWorker("gathering_camp", 1); err != nil {
			t.Fatalf("VibeMode %v: an unassign was refused: %v", on, err)
		}
		if e := last(); e.Tags&TagPeople == 0 {
			t.Errorf("VibeMode %v: an unassign is written %+v", on, e)
		}
	}
}
