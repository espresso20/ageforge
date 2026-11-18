package game

import (
	"testing"

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
