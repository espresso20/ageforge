package game

import (
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// wantCoin fails the test unless the account holds exactly one copy of typ in
// age struck in run, at a level in [lo, hi], and returns it.
func wantCoin(t *testing.T, acct *Account, typ, age, run string, lo, hi int) CoinCopy {
	t.Helper()
	for _, c := range acct.CoinStackOf(typ, age).Copies {
		if c.Run != run {
			continue
		}
		if c.Level < lo || c.Level > hi || c.Type != typ || c.Age != age || c.Seed == 0 {
			t.Errorf("%s in %s: %+v, want a level from %d to %d", typ, age, c, lo, hi)
		}
		return c
	}
	t.Fatalf("no %s coin for %s in run %q (holds %+v)", typ, age, run, acct.CoinStackOf(typ, age))
	return CoinCopy{}
}

func noCoin(t *testing.T, acct *Account, typ, age string) {
	t.Helper()
	if s := acct.CoinStackOf(typ, age); len(s.Copies) != 0 {
		t.Errorf("a %s coin for %s was struck: %+v", typ, age, s)
	}
}

func TestAgeCoinIsStruckWhenTheAgeIsLeft(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	ge.StepTicks(5)
	noCoin(t, acct, config.CoinAge, "primitive_age")
	advanceTo(ge, "stone_age")
	c := wantCoin(t, acct, config.CoinAge, "primitive_age", "run", 6, 6)
	if c.Figure != ge.tick {
		t.Errorf("the figure is %d, want the %d ticks spent in the age", c.Figure, ge.tick)
	}
	noCoin(t, acct, config.CoinAge, "stone_age")
	// A long stay gives the lowest level.
	ge.mu.Lock()
	ge.tick += int(3 * config.AgeTargetTicks("stone_age"))
	ge.mu.Unlock()
	advanceTo(ge, "bronze_age")
	wantCoin(t, acct, config.CoinAge, "stone_age", "run", 1, 1)
}

func TestWonderCoinBelongsToTheWondersAge(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	wonder := rules.Core().Wonder("primitive_age")
	finishOne(ge, wonder)
	wantCoin(t, acct, config.CoinWonder, "primitive_age", "run", 6, 6)

	// Finished in a later age, after a long time: still struck, at a coarse level.
	ge.mu.Lock()
	ge.tick += int(5 * config.AgeTargetTicks("primitive_age"))
	ge.mu.Unlock()
	advanceTo(ge, "stone_age")
	finishOne(ge, rules.Core().Wonder("stone_age"))
	noCoin(t, acct, config.CoinWonder, "bronze_age")
	wantCoin(t, acct, config.CoinWonder, "stone_age", "run", 5, 6)

	// The same wonder in a second run adds a copy.
	ge.mu.Lock()
	ge.activeSaveName = "run2"
	ge.mu.Unlock()
	finishOne(ge, wonder)
	wantCoin(t, acct, config.CoinWonder, "primitive_age", "run2", 1, 1)
	if s := acct.CoinStackOf(config.CoinWonder, "primitive_age"); len(s.Copies) != 2 || s.Count != 2 {
		t.Errorf("two runs: %d copies, count %d", len(s.Copies), s.Count)
	}
}

func TestTechsCoinNeedsEveryTechOfTheAge(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	techs := rules.Core().TechsOf("primitive_age")
	if len(techs) == 0 {
		t.Fatal("the Primitive Age has no techs")
	}
	for i, tech := range techs {
		ge.mu.Lock()
		ge.Research.researched[tech.Key] = true
		ge.note(config.BadgeEvResearchDone, tech.Key)
		ge.mu.Unlock()
		if i < len(techs)-1 {
			noCoin(t, acct, config.CoinTechs, "primitive_age")
		}
	}
	wantCoin(t, acct, config.CoinTechs, "primitive_age", "run", 6, 6)
}

func TestMaximalistCoinNeedsTheBadgesCount(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	def, ok := rules.Core().Badge("maximalist.primitive_age")
	if !ok {
		t.Fatal("no maximalist row for the Primitive Age")
	}
	ge.mu.Lock()
	ge.Buildings.counts["hut"] = int(def.Threshold) - 2
	ge.mu.Unlock()
	finishOne(ge, "gathering_camp")
	noCoin(t, acct, config.CoinMaximalist, "primitive_age")
	finishOne(ge, rules.Core().Wonder("primitive_age")) // a wonder does not count
	noCoin(t, acct, config.CoinMaximalist, "primitive_age")
	finishOne(ge, "stash")
	wantCoin(t, acct, config.CoinMaximalist, "primitive_age", "run", 6, 6)
}

func TestHarbingerCoinIsStruckWhenTheFigureIsMet(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	set := rules.Core()
	var age, key string
	for _, a := range set.AgeKeys() {
		if h, ok := set.Harbinger(a); ok {
			age, key = a, h.Key
			break
		}
	}
	if key == "" {
		t.Fatal("no harbinger in the ruleset")
	}
	ge.mu.Lock()
	ge.note(config.BadgeEvHarbingerMet, key)
	ge.mu.Unlock()
	wantCoin(t, acct, config.CoinHarbinger, age, "run", 6, 6)
}

// A developer command never stops a strike, and an engine with no account strikes nothing.
func TestCoinsAreStruckInDevModeAndNeedAnAccount(t *testing.T) {
	isolateAccountDir(t)
	withDevMode(t)
	ge, acct := devEngine(t, "Ada")
	if out := DevExecCommand("/age stone_age", ge); out == "" {
		t.Fatal("/age gave no answer")
	}
	finishOne(ge, rules.Core().Wonder("stone_age"))
	wantCoin(t, acct, config.CoinWonder, "stone_age", "run", 1, 6)

	bare := NewGameEngine()
	bare.mu.Lock()
	bare.note(config.BadgeEvHarbingerMet, "anything")
	bare.mu.Unlock()
}

func TestLastAgeCoinIsStruckWhenARunEndsByPrestigeFromIt(t *testing.T) {
	isolateAccountDir(t)
	last := "transcendent_age"
	for _, c := range []struct {
		outcome, age string
		want         bool
	}{
		{RunPrestige, last, true},
		{RunSuccumbed, last, false},
		{RunLeft, last, false},
		{RunPrestige, "quantum_age", false},
	} {
		ge, acct := devEngine(t, "Ada")
		ge.mu.Lock()
		ge.age = c.age
		ge.resetRunLocked(0.7, c.outcome, 0)
		ge.mu.Unlock()
		if got := len(acct.CoinStackOf(config.CoinAge, c.age).Copies) == 1; got != c.want {
			t.Errorf("%s from %s: struck %v, want %v", c.outcome, c.age, got, c.want)
		}
	}
}

func TestCoinCopyCarriesTheEditedMark(t *testing.T) {
	isolateAccountDir(t)
	a := testAccount(t)
	if c, _ := strikeAt(a, "age", "stone_age", 1.5, "r1", 10); c.Flags != 0 {
		t.Errorf("a clean strike is marked: %+v", c)
	}
	if c, _ := a.StrikeCoin(rules.Core(), "age", "iron_age", 1, 1, "r1", time.Unix(10, 0), true); c.Flags&BadgeFlagCrossed == 0 {
		t.Errorf("a strike on an edited run is not marked: %+v", c)
	}
	a.Tampered = true
	if c, _ := strikeAt(a, "age", "bronze_age", 1.5, "r1", 10); c.Flags&BadgeFlagCrossed == 0 {
		t.Errorf("a strike on an edited account is not marked: %+v", c)
	}
	// The earlier copy of a run keeps its mark through a merge, as a badge does.
	theirs := &Account{AccountID: a.AccountID}
	theirs.StrikeCoin(rules.Core(), "age", "bronze_age", 1, 1, "r1", time.Unix(5, 0), false)
	mine := testAccount(t)
	mine.StrikeCoin(rules.Core(), "age", "bronze_age", 1, 1, "r1", time.Unix(9, 0), true)
	mine.mu.Lock()
	mine.takeCollectionLocked(theirs.collectionLocked(), true)
	mine.mu.Unlock()
	if c := mine.CoinStackOf("age", "bronze_age").Copies[0]; c.At != 5 || c.Flags != 0 {
		t.Errorf("merge kept %+v, want the earlier, unmarked copy", c)
	}
	// A save round trip keeps the mark.
	b := &Account{AccountID: a.AccountID}
	b.coins = coinsFromFile(a.collectionLocked())
	if c := b.CoinStackOf("age", "iron_age").Copies[0]; c.Flags&BadgeFlagCrossed == 0 {
		t.Errorf("round trip lost the mark: %+v", c)
	}
}
