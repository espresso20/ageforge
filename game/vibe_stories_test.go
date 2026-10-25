package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// The stories the log tells (package flavor: expeditions, encounters, raids,
// the harbinger's lines, the end of a run) are drawn from catalogs and told in
// the script's words when VibeMode is on. These tests play every one of them,
// under a script that marks every text, and hold what is stored to English.

// storyKinds are the prefixes playStories writes its texts under.
const (
	kindExpedition = "expedition"
	kindEncounter  = "encounter"
	kindRaid       = "raid"
	kindRunEnd     = "runend"
	kindView       = "harbinger-view"
	kindLog        = "log"
)

// playStories plays a deterministic run that draws every kind of story in every
// age, the way the engine does (each through its own adapter, under the lock),
// answers the harbinger every way, and returns what the player would have read,
// in order, each text under the kind of story it is. The engine is returned as
// the run left it.
func playStories(t *testing.T, seed int64) ([]string, []LogEntry, *GameEngine) {
	t.Helper()
	var texts []string
	var entries []LogEntry
	rec := func(kind, label, text string) { texts = append(texts, fmt.Sprintf("%s|%s|%s", kind, label, text)) }
	ge := NewGameEngineWith(rules.CoreFor(false))
	ge.SeedRNG(seed)
	view := func(label string) {
		ge.mu.RLock()
		defer ge.mu.RUnlock()
		if v := ge.harbingerView(); v != nil {
			for _, l := range v.Lines {
				rec(kindView, label, l)
			}
		}
	}
	grab := func(label string) {
		for _, l := range ge.GetLogs() {
			rec(kindLog, label, l.Type+"|"+l.Message)
			entries = append(entries, l)
		}
		view(label)
	}
	for i, age := range config.AgeOrder()[1:] {
		if err := ge.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
		ge.GrantTechsForTest()
		for _, f := range rules.CoreFor(false).Factions() {
			_ = ge.MeetFactionForTest(f.Key, 40)
		}
		ge.mu.Lock()
		for j := 0; j < 12; j++ {
			for _, cat := range []string{ExpeditionScouting, ExpeditionMilitary} {
				for _, ok := range []bool{true, false} {
					res := ExpeditionResult{Name: "Scout Party", Category: cat, Success: ok, Rewards: map[string]float64{"food": float64(40 + j), "gold": 3}}
					if q := ge.expeditionFlavorLine(res); q != "" {
						rec(kindExpedition, age, q)
					}
				}
			}
		}
		for _, def := range ge.rules.Factions() {
			for _, mo := range []flavor.Moment{flavor.EncounterStandoff, flavor.EncounterAtCapacity} {
				rec(kindEncounter, age, ge.factionEncounterFlavor(mo, def))
			}
			rec(kindRaid, age, ge.raidLogLine(RaidRequest{FactionKey: def.Key, Resource: "food", Amount: 12, Message: "Raided."}))
		}
		ge.logRunEnding()
		ge.mu.Unlock()
		grab(age + "-stories")
		if err := ge.SummonHarbingerForTest(age); err != nil {
			t.Fatal(err)
		}
		view(age + "-summoned")
		_ = ge.HarbingerAppease()
		_ = ge.HarbingerBrace()
		if i%2 == 0 {
			_ = ge.HarbingerInvite()
		}
		ge.StepTicks(250)
		grab(age + "-harbinger")
		if i%3 == 0 && ge.ForceCatastropheForTest() == nil {
			if i%2 == 0 {
				_ = ge.Endure()
			} else {
				_ = ge.Succumb()
			}
			grab(age + "-doomed")
		}
	}
	for _, step := range []func() error{
		func() error { return ge.EnterAgeForTest("quantum_age") },
		func() error { return ge.SummonHarbingerForTest("quantum_age") },
		func() error { ge.StepTicks(300); return nil },
		func() error { return ge.ForceLastPassageForTest("transcendent_age") },
		func() error { ge.StepTicks(100); return nil },
		func() error { return ge.EndureLastPassage() },
	} {
		_ = step()
		grab("end")
	}
	_ = ge.Endure()
	if err := ge.DoPrestige(); err != nil {
		rec(kindLog, "prestige", err.Error())
	}
	grab("prestige")
	return texts, entries, ge
}

// playStoriesWith plays the run with VibeMode on or off and, when rewrite is
// not nil, with every text the code writes itself read through it.
func playStoriesWith(t *testing.T, on bool, rewrite func(id, plain string) string) ([]string, []LogEntry, *GameEngine) {
	t.Helper()
	vibeOff(t)
	if rewrite != nil {
		defer vibe.UseRewriterForTest(rewrite)()
	}
	vibe.Set(on)
	defer vibe.Set(false)
	return playStories(t, 21)
}

// TestStoriesReadPlainWhileVibeModeIsOff: with the switch off, a run draws the
// stories it always did, whatever the script holds.
func TestStoriesReadPlainWhileVibeModeIsOff(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	plain, _, _ := playStoriesWith(t, false, nil)
	if len(plain) < 5000 {
		t.Fatalf("the run read only %d texts", len(plain))
	}
	scripted, _, _ := playStoriesWith(t, false, mark)
	sameTexts(t, "VibeMode off, a script that rewrites every text", scripted, plain)
	kinds := map[string]int{}
	for _, s := range plain {
		if strings.Contains(s, markOpen) || strings.Contains(s, markClose) {
			t.Fatalf("a plain text has the script's mark in it: %q", s)
		}
		kinds[strings.SplitN(s, "|", 2)[0]]++
	}
	for _, k := range []string{kindExpedition, kindEncounter, kindRaid, kindView, kindLog} {
		if kinds[k] < 30 {
			t.Errorf("the run read only %d texts of the kind %s", kinds[k], k)
		}
	}
}

// TestStoriesFollowTheScriptAndTheStateDoesNot: with the switch on, a script
// that marks every text reaches the lines of every kind of story and changes
// nothing but their words: take the marks out and the run reads as it does in
// plain English, line for line; the state, its digest and the save are the
// run's own, with no text of the script in them.
func TestStoriesFollowTheScriptAndTheStateDoesNot(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	plain, _, plainGE := playStoriesWith(t, true, func(id, plain string) string { return plain })
	marked, _, markedGE := playStoriesWith(t, true, mark)
	if len(marked) != len(plain) {
		t.Fatalf("the run under the script read %d texts, in plain English %d", len(marked), len(plain))
	}
	told := map[string]int{}
	comments := 0
	for i := range plain {
		if got := unmark.Replace(marked[i]); got != plain[i] {
			t.Fatalf("text %d under the script, marks taken out, reads\n  %q\nin plain English\n  %q", i, got, plain[i])
		}
		if marked[i] == plain[i] {
			continue
		}
		kind := strings.SplitN(plain[i], "|", 2)[0]
		told[kind]++
		if kind == kindLog && strings.Contains(plain[i], "[comment]") {
			comments++ // a line of story in the log
		}
	}
	// Every story was met, and the script reached it: a line of each kind is
	// the script's, and the lines of the log that are stories are among them.
	for _, want := range []struct {
		kind string
		min  int
	}{{kindExpedition, 30}, {kindEncounter, 100}, {kindRaid, 100}, {kindView, 30}} {
		t.Logf("%d texts of the kind %s were told in the script's words", told[want.kind], want.kind)
		if told[want.kind] < want.min {
			t.Errorf("the script told only %d texts of the kind %s, want at least %d", told[want.kind], want.kind, want.min)
		}
	}
	// A resource named inside a story is the script's name for it, marked
	// like the rest (mark puts the mark after the first letter of "Food").
	named := 0
	for i := range plain {
		if marked[i] != plain[i] && (strings.HasPrefix(plain[i], kindExpedition+"|") || strings.HasPrefix(plain[i], kindRaid+"|")) && strings.Contains(marked[i], "f"+markOpen+"ood"+markClose) {
			named++
		}
	}
	t.Logf("%d stories named a resource in the script's words", named)
	if named < 20 {
		t.Errorf("only %d stories named a resource in the script's words, want at least 20", named)
	}
	t.Logf("%d lines of story in the log were told in the script's words", comments)
	if comments < 1000 {
		t.Errorf("the script told only %d lines of story in the log", comments)
	}
	// The line that ends a run is told too.
	ends := 0
	for i := range plain {
		if strings.HasPrefix(plain[i], kindLog+"|prestige|") && strings.Contains(plain[i], "[comment]") && marked[i] != plain[i] {
			ends++
		}
	}
	if ends == 0 {
		t.Error("the closing line of the run was not told in the script's words")
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
	vibe.Set(true)
	defer vibe.Set(false)
	defer vibe.UseRewriterForTest(mark)()
	if err := markedGE.SaveGame("stories"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(savePath("stories"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), markOpen) {
		t.Error("the save holds a text of the script")
	}
}

// harborLines are what the engine's thread holds and what a reader of it is
// given, for the live thread.
func harborLines(ge *GameEngine) (kept, keys, read []string) {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	kept = append(kept, ge.harbinger.Lines...)
	keys = append(keys, ge.harbinger.LineKeys...)
	if v := ge.harbingerView(); v != nil {
		read = v.Lines
	}
	return kept, keys, read
}

// TestAHarborLineIsKeptInEnglishAndToldWhenItIsRead: the arrival and the
// warning of a harbinger are part of its thread and saved with it. The thread
// holds them in plain English with a key each, the log tells them in the
// script's words as they are drawn, and the view and a loaded save tell them
// again from the keys, in whichever script is in force when they are read.
func TestAHarborLineIsKeptInEnglishAndToldWhenItIsRead(t *testing.T) {
	isolateAccountDir(t)
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	vibeOff(t)
	defer vibe.Set(false)

	// In plain English the thread keeps its lines and the view reads them as they are.
	plainGE := newSeededEngine(4)
	if err := plainGE.SummonHarbingerForTest("bronze_age"); err != nil {
		t.Fatal(err)
	}
	if kept, keys, read := harborLines(plainGE); len(kept) != 2 || len(keys) != 2 || !reflect.DeepEqual(kept, read) {
		t.Fatalf("in plain English the thread keeps %q (keys %q) and the view reads %q", kept, keys, read)
	}

	// With the switch on, the same thread (the VibeMode set names the figure,
	// so the same engine under a script that changes nothing is the twin).
	vibe.Set(true)
	// Read the twin while that script is still in force: the real script
	// has its own lines for these warnings.
	wantKept, wantKeys, wantRead := func() ([]string, []string, []string) {
		defer vibe.UseRewriterForTest(func(id, plain string) string { return plain })()
		ge := newSeededEngine(4)
		if err := ge.SummonHarbingerForTest("bronze_age"); err != nil {
			t.Fatal(err)
		}
		return harborLines(ge)
	}()
	if len(wantKept) != 2 {
		t.Fatalf("under a script that changes nothing the thread keeps %q", wantKept)
	}
	// The thread keeps the lines a game in plain English keeps, the figure's
	// name too, though the VibeMode set names the figure otherwise; the view
	// tells them with the name it has.
	if plainKept, plainKeys, _ := harborLines(plainGE); !reflect.DeepEqual(wantKept, plainKept) || !reflect.DeepEqual(wantKeys, plainKeys) {
		t.Errorf("on the VibeMode set the thread keeps %q (keys %q), in plain English %q (keys %q)", wantKept, wantKeys, plainKept, plainKeys)
	}
	if wantRead[0] == wantKept[0] && wantRead[1] == wantKept[1] {
		t.Log("the figure has the same name in both scripts here, so the view reads what the thread keeps")
	}

	defer vibe.UseRewriterForTest(mark)()
	ge := newSeededEngine(4)
	if err := ge.SummonHarbingerForTest("bronze_age"); err != nil {
		t.Fatal(err)
	}
	kept, keys, read := harborLines(ge)
	if !reflect.DeepEqual(kept, wantKept) || !reflect.DeepEqual(keys, wantKeys) {
		t.Errorf("under the script the thread keeps %q (keys %q), under one that changes nothing %q (keys %q)", kept, keys, wantKept, wantKeys)
	}
	if len(keys) != len(kept) || keys[0] == "" || keys[1] == "" {
		t.Fatalf("the thread's keys are %q", keys)
	}
	for i := range read {
		if !strings.Contains(read[i], markOpen) || unmark.Replace(read[i]) != wantRead[i] {
			t.Errorf("the view reads %q under the script, under one that changes nothing %q", read[i], wantRead[i])
		}
	}
	told := 0
	for _, l := range ge.GetLogs() {
		for i := range read {
			if strings.Contains(l.Message, "[comment]"+read[i]+"[-]") {
				told++
			}
		}
	}
	if told != len(read) {
		t.Errorf("the log told %d of the thread's %d lines as the view reads them", told, len(read))
	}

	// The save keeps English and the keys; a load tells them in the script in force.
	if err := ge.SaveGame("harbor"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(savePath("harbor"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), markOpen) {
		t.Fatal("the save holds a text of the script")
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("harbor"); err != nil {
		t.Fatal(err)
	}
	if k, _, r := harborLines(loaded); !reflect.DeepEqual(k, wantKept) || !reflect.DeepEqual(r, read) {
		t.Errorf("a load under the script keeps %q and reads %q, want %q and %q", k, r, wantKept, read)
	}
	vibe.Set(false)
	if k, keys2, r := harborLines(loaded); !reflect.DeepEqual(k, wantKept) || !reflect.DeepEqual(r, wantKept) || !reflect.DeepEqual(keys2, keys) {
		t.Errorf("with VibeMode off again the load keeps %q and reads %q", k, r)
	}
	if strings.Contains(strings.Join(wantKept, " "), "Vibe Reader") {
		t.Error("the thread keeps the VibeMode set's name for the figure")
	}
}

// TestASaveWithoutKeysStillLoadsAndReadsPlain: a save from before the keys
// has the lines and no keys. It loads, and its lines read as they were
// drawn, in any script.
func TestASaveWithoutKeysStillLoadsAndReadsPlain(t *testing.T) {
	isolateAccountDir(t)
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	vibeOff(t)
	defer vibe.Set(false)

	ge := newSeededEngine(4)
	if err := ge.SummonHarbingerForTest("iron_age"); err != nil {
		t.Fatal(err)
	}
	wantKept, _, _ := harborLines(ge)
	ge.mu.RLock()
	old := ge.buildSaveSnapshot()
	ge.mu.RUnlock()
	if old.Harbinger == nil || len(old.Harbinger.LineKeys) == 0 {
		t.Fatal("the save holds no keys to take away")
	}
	old.Harbinger.LineKeys = nil
	old.Signature = signSave(old)
	data, err := json.MarshalIndent(old, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "line_keys") {
		t.Fatal("the save still has its keys")
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(savePath("before-keys"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, on := range []bool{false, true} {
		func() {
			defer vibe.UseRewriterForTest(mark)()
			vibe.Set(on)
			defer vibe.Set(false)
			loaded := NewGameEngine()
			if err := loaded.LoadGame("before-keys"); err != nil {
				t.Fatalf("VibeMode %v: %v", on, err)
			}
			if k, keys, r := harborLines(loaded); !reflect.DeepEqual(k, wantKept) || !reflect.DeepEqual(r, wantKept) || len(keys) != 0 {
				t.Errorf("VibeMode %v: the old save keeps %q with keys %q and reads %q, want %q", on, k, keys, r, wantKept)
			}
			// It plays on: a new warning in its thread has keys again.
			loaded.StepTicks(50)
		}()
	}
}

// TestAStoryKeepsItsTagsInEveryScript: what a line of story is about is the
// tags it was written with, so a script leaves every line in the panel it was
// in: a run in plain English and the same run under a script write the same
// lines with the same tags.
func TestAStoryKeepsItsTagsInEveryScript(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	_, plain, _ := playStoriesWith(t, false, nil)
	_, marked, _ := playStoriesWith(t, true, mark)
	if len(plain) != len(marked) {
		t.Fatalf("the run under the script wrote %d lines, in plain English %d", len(marked), len(plain))
	}
	stories, raids := 0, 0
	for i := range plain {
		if got, want := marked[i].Tags, plain[i].Tags&^TagPlain; got != want {
			t.Fatalf("line %d has tags %b under the script and %b in plain English: %q", i, got, want, plain[i].Message)
		}
		if strings.Contains(plain[i].Message, "[comment]") {
			stories++
			if marked[i].Tags&TagArmy != 0 {
				raids++
			}
		}
	}
	if stories < 1000 || raids == 0 {
		t.Errorf("%d lines of story in the run, %d of them about the army", stories, raids)
	}
}

// TestAKeyThatNamesAnotherLineIsNotTrusted: a key is only used to tell a line
// again when it tells the line the thread holds, in plain English. A key from a
// build whose catalog differed names some other sentence, and the line reads
// as it was drawn.
func TestAKeyThatNamesAnotherLineIsNotTrusted(t *testing.T) {
	isolateAccountDir(t)
	rules.CoreFor(true) // the real VibeMode set, built before any script is loaned
	vibeOff(t)
	defer vibe.Set(false)
	defer vibe.UseRewriterForTest(mark)()
	vibe.Set(true)

	ge := newSeededEngine(4)
	if err := ge.SummonHarbingerForTest("bronze_age"); err != nil {
		t.Fatal(err)
	}
	kept, keys, read := harborLines(ge)
	if len(keys) != 2 || !strings.Contains(read[0], markOpen) || !strings.Contains(read[1], markOpen) {
		t.Fatalf("the thread keeps %q (keys %q) and the view reads %q", kept, keys, read)
	}
	ge.mu.Lock()
	ge.harbinger.LineKeys = []string{"nonesuch_0", keys[0]} // no such sentence, and the other line's key
	ge.mu.Unlock()
	_, _, read = harborLines(ge)
	if read[0] != kept[0] || read[1] != kept[1] {
		t.Errorf("with keys that do not name them the view reads %q, want the lines as they were drawn %q", read, kept)
	}
}
