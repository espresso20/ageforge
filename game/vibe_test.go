package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

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
	if err := acct.SetVibe(true); err != nil {
		t.Fatal(err)
	}
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
	if err := loaded.Save(); err != nil {
		t.Fatal(err)
	}
	if !slotAccount(t, acct.AccountID).VibeOn() {
		t.Error("saving the account lost the setting")
	}
	// Motion and VibeMode sit side by side in the one file.
	if err := loaded.SetMotion(false); err != nil {
		t.Fatal(err)
	}
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
	if err := slotAccount(t, acct.AccountID).SetVibe(false); err != nil {
		t.Fatal(err)
	}
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

	if err := ge.SetVibeMode(true); err != nil {
		t.Fatal(err)
	}
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

	if err := ge.SetVibeMode(false); err != nil {
		t.Fatal(err)
	}
	wantManagersOn(t, ge, rules.CoreFor(false), "after vibemode off")
	if got, _ := ge.Rules().Building("hut"); got.Name != "Hut" {
		t.Errorf("the hut reads %q with VibeMode off again", got.Name)
	}

	// An engine with a set of its own keeps it.
	own := NewGameEngineWith(secondSet())
	set := own.Rules()
	if err := own.SetVibeMode(true); err != nil {
		t.Fatal(err)
	}
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
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
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
	moved.StepTicks(300)
	control.StepTicks(300)
	a, b := stateMap(t, moved), stateMap(t, control)
	if !reflect.DeepEqual(a, b) {
		for k := range a {
			if !reflect.DeepEqual(a[k], b[k]) {
				t.Errorf("after 300 ticks under VibeMode, %q differs from the twin that never switched", k)
			}
		}
	}

	// A save made under VibeMode is the save made in plain English: loaded
	// into a fresh engine each, the two are in one state, and play on alike.
	if err := moved.SaveGame("vibe"); err != nil {
		t.Fatal(err)
	}
	moved.followVibe(false)
	if err := moved.SaveGame("plain"); err != nil {
		t.Fatal(err)
	}
	fromVibe, fromPlain := NewGameEngine(), NewGameEngine()
	if err := fromVibe.LoadGame("vibe"); err != nil {
		t.Fatalf("a save made under VibeMode does not load: %v", err)
	}
	if err := fromPlain.LoadGame("plain"); err != nil {
		t.Fatal(err)
	}
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

// TestEveryTextCanChangeWithoutTouchingANumber is the strongest form of
// "words only": a script that rewrites every text of the ruleset that has an
// id, the names of ages, resources, buildings and techs included, and a run
// under it comes out in the state of the same run in plain English, to the
// last bit.
func TestEveryTextCanChangeWithoutTouchingANumber(t *testing.T) {
	vibeOff(t)
	script := map[string]string{}
	for _, c := range words.Cells(rules.FromConfig()) {
		if c.Field == "config.TechDef.Code" {
			continue // a code is five capital letters, not a line
		}
		script[c.ID] = "~" + c.Text + "~"
	}
	restore := vibe.UseScriptForTest(script)
	set := rules.Compile(rules.Vibed(rules.FromConfig()))
	restore()
	if set.Digest() == rules.CoreFor(false).Digest() {
		t.Fatal("the rewritten set reads the same as the plain one")
	}
	play := func(set *rules.Set) *GameEngine {
		ge := NewGameEngineWith(set)
		ge.SeedRNG(42)
		ge.age = "industrial_age"
		ageOrder := ge.progress.GetAgeOrder()
		for key, def := range ge.Buildings.defs {
			if ageOrder[def.RequiredAge] > ageOrder[ge.age] || def.Category == "wonder" {
				continue
			}
			ge.Buildings.UnlockBuilding(key)
			ge.Buildings.counts[key] = 3
		}
		ge.GrantTechsForTest()
		ge.recalculateRates()
		ge.StepTicks(400)
		return ge
	}
	plain, vibed := play(rules.CoreFor(false)), play(set)
	a, b := stateMap(t, plain), stateMap(t, vibed)
	// An offer on the board and a party on the road keep the title, place and
	// story they were generated with, in the save as in play: they are words
	// frozen into the state when the offer was made, and the one place where a
	// script reaches a save (open: the board's text should be drawn from the
	// script in force, not stored).
	dropText(a["military"])
	dropText(b["military"])
	// The title is written to a save in plain English for the two core sets
	// (TestASaveHoldsThePlainTitle); this set is a set of its own.
	delete(a, "current_title")
	delete(b, "current_title")
	for k := range a {
		if !reflect.DeepEqual(a[k], b[k]) {
			t.Errorf("under a script that rewrites every text, %q is not what it is in plain English: %.300v vs %.300v", k, a[k], b[k])
		}
	}
}

// dropText removes the keys that hold a generated line from a decoded JSON
// value, wherever they sit in it.
func dropText(v any) {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			switch k {
			case "title", "place", "description", "story", "ambush_story", "aux_story":
				delete(x, k)
			default:
				dropText(e)
			}
		}
	case []any:
		for _, e := range x {
			dropText(e)
		}
	}
}

// TestASaveHoldsThePlainTitle: the title a civilization wears is the one
// text a save writes besides names the player gave, so a save made under
// VibeMode holds the plain title, and the same save read in either script
// shows the title of that script.
func TestASaveHoldsThePlainTitle(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	vibeOff(t)
	titleOf := func(ge *GameEngine) (worn, saved string) {
		ge.Milestones.recalculateTitle()
		var save GameSave
		b, err := ge.StateJSON()
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(b, &save); err != nil {
			t.Fatal(err)
		}
		return ge.Milestones.GetCurrentTitle(), save.CurrentTitle
	}
	ge := NewGameEngine()
	for i, def := range ge.Milestones.defs {
		if i < 5 {
			ge.Milestones.completed[def.Key] = true
		}
	}
	if worn, saved := titleOf(ge); worn != "Aspiring" || saved != "Aspiring" {
		t.Fatalf("in plain English the title is %q and the save holds %q", worn, saved)
	}
	ge.followVibe(true)
	if worn, saved := titleOf(ge); worn != "Lowkey Cooking" || saved != "Aspiring" {
		t.Errorf("under VibeMode the title worn is %q (want Lowkey Cooking) and the save holds %q (want Aspiring)", worn, saved)
	}
	ge.followVibe(false)
	if worn, saved := titleOf(ge); worn != "Aspiring" || saved != "Aspiring" {
		t.Errorf("back in plain English the title is %q and the save holds %q", worn, saved)
	}
}
