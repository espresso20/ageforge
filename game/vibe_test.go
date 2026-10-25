package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/words"
)

// vibeOff puts the process's VibeMode switch back where a test found it.
func vibeOff(t *testing.T) {
	t.Helper()
	was := vibe.On()
	t.Cleanup(func() { vibe.Set(was) })
	vibe.Set(false)
}

// vibeSession opens a test of the script: the real VibeMode set is built
// before any script is loaned (so a loaned script never gets into it), the
// switch is off, and it goes back where the test found it when the test ends.
func vibeSession(t *testing.T) {
	t.Helper()
	rules.CoreFor(true)
	vibeOff(t)
}

// vibeRead sets the switch to on and, when rewrite is not nil, has every text
// the code writes itself read through it. Defer what it returns: it puts the
// switch off and the script back.
func vibeRead(on bool, rewrite func(id, plain string) string) (restore func()) {
	undo := func() {}
	if rewrite != nil {
		undo = vibe.UseRewriterForTest(rewrite)
	}
	vibe.Set(on)
	return func() {
		vibe.Set(false)
		undo()
	}
}

// TestVibeSettingLivesBesideTheAccountFile: VibeMode is off by default.
// Turning it on writes settings.json in the account's slot and leaves
// account.json byte for byte as it was. A fresh load reads it, a save of the
// account keeps it, and another account has its own.
func TestVibeSettingLivesBesideTheAccountFile(t *testing.T) {
	isolateAccountDir(t)
	vibeOff(t)
	_, acct := earnSome(t, "Ada")
	if acct.VibeOn() {
		t.Fatal("VibeMode is on for a new account")
	}
	settings := filepath.Join(accountDir(acct.AccountID), settingsFileName)
	if data, err := os.ReadFile(settings); err == nil {
		var unset map[string]any
		if json.Unmarshal(data, &unset) != nil || unset["vibe"] != nil {
			t.Fatalf("an account that never set VibeMode has it on file: %s", data)
		}
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	before := slotFile(t, acct.AccountID)
	badges := slotBadgeFile(t, acct.AccountID)
	noErr(t, acct.SetVibe(true))
	if !acct.VibeOn() {
		t.Error("VibeMode is still off")
	}
	if got := slotFile(t, acct.AccountID); string(got) != string(before) {
		t.Error("setting VibeMode rewrote account.json")
	}
	if got := slotBadgeFile(t, acct.AccountID); string(got) != string(badges) {
		t.Error("setting VibeMode rewrote the badge file")
	}
	assertOldShape(t, slotFile(t, acct.AccountID))
	var file map[string]any
	data, err := os.ReadFile(settings)
	if err != nil || json.Unmarshal(data, &file) != nil || file["vibe"] != "on" {
		t.Fatalf("settings.json: %s (%v)", data, err)
	}

	loaded := slotAccount(t, acct.AccountID)
	if !loaded.VibeOn() || loaded.Tampered {
		t.Errorf("after a load: VibeMode %v, flagged %v", loaded.VibeOn(), loaded.Tampered)
	}
	noErr(t, loaded.Save())
	if !slotAccount(t, acct.AccountID).VibeOn() {
		t.Error("saving the account lost the setting")
	}
	// Motion and VibeMode sit side by side in the one file.
	noErr(t, loaded.SetMotion(false))
	if again := slotAccount(t, acct.AccountID); !again.VibeOn() || again.MotionOn() {
		t.Errorf("motion and VibeMode in one file: VibeMode %v, motion %v", again.VibeOn(), again.MotionOn())
	}

	// A second account is its own.
	other, err := CreateAccount("Grace")
	if err != nil {
		t.Fatal(err)
	}
	if other.VibeOn() {
		t.Error("a new account inherited VibeMode from another")
	}
	noErr(t, slotAccount(t, acct.AccountID).SetVibe(false))
	if slotAccount(t, acct.AccountID).VibeOn() {
		t.Error("turning VibeMode off did not stick")
	}
}

// TestVibeModeFollowsTheAccount: the engine plays by the core set of the
// account's setting: on an account's load, on a switch to another account,
// and when the command turns it on or off. An engine on a set of its own is
// not moved.
func TestVibeModeFollowsTheAccount(t *testing.T) {
	isolateAccountDir(t)
	vibeOff(t)
	ge, alpha := devEngine(t, "Alpha")
	beta, err := CreateAccount("Beta")
	if err != nil {
		t.Fatal(err)
	}
	ge.SetAccount(alpha)
	if ge.Rules() != rules.CoreFor(false) || vibe.On() {
		t.Fatal("an account with VibeMode off did not leave the engine on the plain set")
	}

	noErr(t, ge.SetVibeMode(true))
	wantManagersOn(t, ge, rules.CoreFor(true), "after vibemode on")
	if !vibe.On() || !alpha.VibeOn() || !slotAccount(t, alpha.AccountID).VibeOn() {
		t.Error("vibemode on was not kept in the switch and the account")
	}
	if got, _ := ge.Rules().Building("hut"); got.Name != "Crash Pad" {
		t.Errorf("the hut reads %q with VibeMode on", got.Name)
	}

	// Another account has its own setting: the words go back to plain.
	if _, err := ge.SwitchAccount(beta.AccountID); err != nil {
		t.Fatal(err)
	}
	wantManagersOn(t, ge, rules.CoreFor(false), "after a switch to an account with it off")
	if vibe.On() {
		t.Error("the switch stayed on after a switch to an account with VibeMode off")
	}
	// And back: Alpha still has it on.
	if _, err := ge.SwitchAccount(alpha.AccountID); err != nil {
		t.Fatal(err)
	}
	wantManagersOn(t, ge, rules.CoreFor(true), "after a switch back to the account with it on")

	noErr(t, ge.SetVibeMode(false))
	wantManagersOn(t, ge, rules.CoreFor(false), "after vibemode off")
	if got, _ := ge.Rules().Building("hut"); got.Name != "Hut" {
		t.Errorf("the hut reads %q with VibeMode off again", got.Name)
	}

	// An engine with a set of its own keeps it.
	own := NewGameEngineWith(secondSet())
	set := own.Rules()
	noErr(t, own.SetVibeMode(true))
	if own.Rules() != set {
		t.Error("VibeMode moved an engine off its own ruleset")
	}
	// With no account the setting lasts for the session.
	bare := NewGameEngine()
	if err := bare.SetVibeMode(false); err != nil || bare.Rules() != rules.CoreFor(false) {
		t.Errorf("no account: %v", err)
	}
}

// stateMap is an engine's state as the save writes it, key by key. The save
// carries no log and no name from the script, so VibeMode has nothing in it
// to change.
func stateMap(t *testing.T, ge *GameEngine) map[string]any {
	t.Helper()
	b, err := ge.StateJSON()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	noErr(t, json.Unmarshal(b, &m))
	if len(m) < 20 {
		t.Fatalf("the state has only %d keys", len(m))
	}
	return m
}

// TestVibeModeChangesWordsNotNumbers: switching VibeMode on and off in the
// middle of a run touches no number. The state is the same to the last bit
// right after each switch, a run that plays on under the VibeMode set ends
// in the state of its twin that never switched, and a save made in one script
// loads in the other.
func TestVibeModeChangesWordsNotNumbers(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	vibeOff(t)
	play := func() *GameEngine {
		ge := newLateGameEngine(t)
		ge.GrantTechsForTest()
		ge.StepTicks(200)
		return ge
	}
	moved, control := play(), play()
	if moved.StateDigest() != control.StateDigest() {
		t.Fatal("two engines with one seed and one script are in different states")
	}
	was := moved.StateDigest()
	moved.followVibe(true)
	wantManagersOn(t, moved, rules.CoreFor(true), "after the switch")
	if moved.StateDigest() != was {
		t.Error("turning VibeMode on changed the run's state")
	}
	// Long enough for the board to expire and make new offers, and for
	// missions to be sent, meet the road and come home, under the script.
	moved.StepTicks(2500)
	control.StepTicks(2500)
	a, b := stateMap(t, moved), stateMap(t, control)
	if !reflect.DeepEqual(a, b) {
		for k := range a {
			if !reflect.DeepEqual(a[k], b[k]) {
				t.Errorf("after 2,500 ticks under VibeMode, %q differs from the twin that never switched", k)
			}
		}
	}

	// A save made under VibeMode is the save made in plain English: loaded
	// into a fresh engine each, the two are in one state, and play on alike.
	noErr(t, moved.SaveGame("vibe"))
	moved.followVibe(false)
	noErr(t, moved.SaveGame("plain"))
	fromVibe, fromPlain := NewGameEngine(), NewGameEngine()
	if err := fromVibe.LoadGame("vibe"); err != nil {
		t.Fatalf("a save made under VibeMode does not load: %v", err)
	}
	noErr(t, fromPlain.LoadGame("plain"))
	if fromVibe.StateDigest() != fromPlain.StateDigest() {
		t.Error("a save made under VibeMode loads as another state than the same run saved in plain English")
	}
	fromVibe.StepTicks(50)
	fromPlain.StepTicks(50)
	if fromVibe.StateDigest() != fromPlain.StateDigest() {
		t.Error("a game loaded from a VibeMode save plays on differently")
	}
}

// TestVibeDemoLines: the welcome line of the log reads the script while
// VibeMode is on at the engine's making, and plain English otherwise.
func TestVibeDemoLines(t *testing.T) {
	vibeOff(t)
	first := func() string {
		ge := NewGameEngine()
		logs := ge.GetState().Log
		if len(logs) == 0 {
			t.Fatal("a new engine logged nothing")
		}
		return logs[0].Message
	}
	if got := first(); !strings.HasPrefix(got, "Welcome to AgeForge. You have nothing but your hands.") {
		t.Errorf("the plain welcome is %q", got)
	}
	vibe.Set(true)
	if got := first(); !strings.Contains(got, "lock in") {
		t.Errorf("the VibeMode welcome is %q", got)
	}
}

// TestVibeModeSwitchesWhileTheGameRuns: the switch may be thrown from the
// UI while the tick loop is going. Run under -race, this is the check that
// nothing reads the ruleset or the flag unguarded; without it, it checks the
// run comes through whole.
func TestVibeModeSwitchesWhileTheGameRuns(t *testing.T) {
	vibeOff(t)
	ge := newLateGameEngine(t)
	ge.GrantTechsForTest()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 150; i++ {
			ge.StepTicks(1)
			_ = ge.GetState()
		}
	}()
	for i := 0; i < 60; i++ {
		ge.followVibe(i%2 == 0)
		_ = rules.Core()
	}
	<-done
	ge.followVibe(false)
	wantManagersOn(t, ge, rules.CoreFor(false), "after the switching")
	if ge.StateDigest() == "unencodable" {
		t.Error("the state cannot be encoded after the switching")
	}
}

// wholeRun plays a run on set that reaches the places a text could be kept
// in the state: the board and missions on the road, every age's awakening and
// era event, the civilizations met, a harbinger's thread, dooms endured and
// succumbed to and the Last Passage.
func wholeRun(t *testing.T, set *rules.Set) *GameEngine {
	t.Helper()
	ge := NewGameEngineWith(set)
	ge.SeedRNG(7)
	for i, age := range config.AgeOrder()[1:] {
		noErr(t, ge.EnterAgeForTest(age))
		ge.GrantTechsForTest()
		for _, f := range rules.CoreFor(false).Factions() {
			_ = ge.MeetFactionForTest(f.Key, 40)
		}
		_ = ge.SummonHarbingerForTest(age)
		ge.StepTicks(350)
		if i%3 == 0 && ge.ForceCatastropheForTest() == nil {
			if i%2 == 0 {
				_ = ge.Endure()
			} else {
				_ = ge.Succumb()
			}
		}
		ge.StepTicks(350)
	}
	for _, st := range endingSteps(ge) {
		if err := st.do(); err != nil {
			t.Fatalf("%s: %v", st.name, err)
		}
	}
	return ge
}

// runStep is one step of the end of a whole run.
type runStep struct {
	name string
	do   func() error
}

// endingSteps are the steps wholeRun and playRun both end on, once the ages are
// played: the quantum age, a harbinger, the Last Passage endured and a step
// back to an earlier age. Making the list does nothing; each step acts when it
// is called, so the two runs keep their own calls in their own order.
func endingSteps(ge *GameEngine) []runStep {
	return []runStep{
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
}

// thenPrestige ends the run wholeRun played with a prestige and plays on in
// the new one.
func thenPrestige(t *testing.T, ge *GameEngine) {
	t.Helper()
	_ = ge.Endure()
	noErr(t, ge.DoPrestige())
	ge.StepTicks(300)
}

// rewrittenSet is the core ruleset with every text that has an id rewritten
// ("~Hut~"): a script of a different size from the shipped one, which changes
// everything there is to change.
func rewrittenSet(t *testing.T) *rules.Set {
	t.Helper()
	script := map[string]string{}
	for _, c := range words.Cells(rules.FromConfig()) {
		if c.Field == "config.TechDef.Code" {
			continue // a code is five capital letters, not a line
		}
		script[c.ID] = "~" + c.Text + "~"
	}
	defer vibe.UseScriptForTest(script)()
	return rules.Compile(rules.Vibed(rules.FromConfig()))
}

// TestEveryTextCanChangeWithoutTouchingANumber is the strongest form of
// "words only": a script that rewrites every text of the ruleset that has an
// id, the names of ages, resources, buildings and techs included, and a whole
// run under it comes out in the state of the same run in plain English, to
// the last bit. Nothing a script writes is kept in the state.
func TestEveryTextCanChangeWithoutTouchingANumber(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	vibeOff(t)
	set := rewrittenSet(t)
	if set.Digest() == rules.CoreFor(false).Digest() {
		t.Fatal("the rewritten set reads the same as the plain one")
	}
	plain, vibed := wholeRun(t, rules.CoreFor(false)), wholeRun(t, set)
	for stage := 0; stage < 2; stage++ {
		if stage == 1 {
			thenPrestige(t, plain)
			thenPrestige(t, vibed)
		}
		sameStateButTheTitle(t, plain, vibed, "under a script that rewrites every text")
	}
}

// sameState fails the test for every key of the two engines' states that
// differs, but the keys in skip.
func sameState(t *testing.T, plain, vibed *GameEngine, when string, skip ...string) {
	t.Helper()
	a, b := stateMap(t, plain), stateMap(t, vibed)
	for _, k := range skip {
		delete(a, k)
		delete(b, k)
	}
	for k := range a {
		if !reflect.DeepEqual(a[k], b[k]) {
			t.Errorf("%s, %q is not what it is in plain English: %.300v vs %.300v", when, k, a[k], b[k])
		}
	}
}

// sameStateButTheTitle is sameState but for current_title, which the core sets
// alone write in plain English (TestASaveHoldsThePlainTitle).
func sameStateButTheTitle(t *testing.T, plain, vibed *GameEngine, when string) {
	t.Helper()
	// The title is written to a save in plain English for the two core sets
	// (TestASaveHoldsThePlainTitle); this set is a set of its own.
	sameState(t, plain, vibed, when, "current_title")
}

// TestAWholeRunUnderTheShippedScript is the same on the two core sets: a
// whole run read in VibeMode ends in the state of the run read in plain
// English, and a save of it is the same save.
func TestAWholeRunUnderTheShippedScript(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	vibeOff(t)
	plain, vibed := wholeRun(t, rules.CoreFor(false)), wholeRun(t, rules.CoreFor(true))
	sameState(t, plain, vibed, "under the shipped script")
	if plain.StateDigest() != vibed.StateDigest() {
		t.Error("the state digests differ")
	}
	thenPrestige(t, plain)
	thenPrestige(t, vibed)
	if plain.StateDigest() != vibed.StateDigest() {
		t.Error("the state digests differ after a prestige")
	}
	noErr(t, vibed.SaveGame("vibed"))
	noErr(t, plain.SaveGame("plain"))
	read := func(name string) map[string]any {
		raw, err := os.ReadFile(savePath(name))
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		noErr(t, json.Unmarshal(raw, &m))
		for _, k := range []string{"timestamp", "_sig"} {
			delete(m, k)
		}
		if stats, ok := m["stats"].(map[string]any); ok {
			delete(stats, "game_started") // the wall clock
		}
		return m
	}
	rv, rp := read("vibed"), read("plain")
	for k := range rv {
		if !reflect.DeepEqual(rv[k], rp[k]) {
			t.Errorf("the save of a run read in VibeMode is not the save of the run read in plain English: %q is %.200v and %.200v", k, rv[k], rp[k])
		}
	}
}

// TestSwitchingRedrawsEveryText: texts the state keeps by index or by key
// (offers, missions, returns, event names, the histories) read in the script
// in force. A run played in plain English and then moved onto a ruleset that
// rewrites every text reads as the run played on it all along, text for text;
// and back. (The moves are Rebind, which is what the VibeMode switch calls.)
func TestSwitchingRedrawsEveryText(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	vibeOff(t)
	set := rewrittenSet(t)
	texts := func(ge *GameEngine) map[string]any {
		st := ge.GetState()
		return map[string]any{
			"military": st.Military, "events": st.ActiveEvents, "epoch": st.EpochEventHistory,
			"harbingers": st.HarbingerHistory, "dooms": st.CatastropheHistory, "title": st.Milestones.CurrentTitle,
		}
	}
	rewrittenAllAlong := wholeRun(t, set)
	switched := wholeRun(t, rules.CoreFor(false))
	plainTexts := texts(switched)
	switched.Rebind(set)
	got, want := texts(switched), texts(rewrittenAllAlong)
	changed := 0
	for k := range want {
		if !reflect.DeepEqual(got[k], want[k]) {
			t.Errorf("after the switch, %q reads %.300v, want %.300v", k, got[k], want[k])
		}
		if !reflect.DeepEqual(got[k], plainTexts[k]) {
			changed++
		}
	}
	if changed < 5 {
		t.Errorf("the switch changed the words of only %d of %d kinds of kept text", changed, len(want))
	}
	switched.Rebind(rules.CoreFor(false))
	if back := texts(switched); !reflect.DeepEqual(back, plainTexts) {
		for k := range back {
			if !reflect.DeepEqual(back[k], plainTexts[k]) {
				t.Errorf("after switching back, %q reads %.300v, want %.300v", k, back[k], plainTexts[k])
			}
		}
	}
}
