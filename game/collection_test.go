package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
)

// strikeAt strikes a coin whose figure is the given share of the age's reference time.
func strikeAt(a *Account, typ, age string, share float64, run string, at int64) (CoinCopy, bool) {
	ticks := int(share * config.AgeTargetTicks(age))
	return a.StrikeCoin(typ, age, ticks, 1, run, time.Unix(at, 0))
}

func TestCoinOneCopyForEachRun(t *testing.T) {
	isolateAccountDir(t)
	a := testAccount(t)
	if c, ok := strikeAt(a, "wonder", "stone_age", 1.5, "r1", 100); !ok || c.Level != 2 {
		t.Fatalf("first strike: %+v, %v", c, ok)
	}
	// Worse and equal strikes in the same run change nothing.
	if _, ok := strikeAt(a, "wonder", "stone_age", 1.9, "r1", 200); ok {
		t.Error("a worse strike changed the stack")
	}
	if c, ok := strikeAt(a, "wonder", "stone_age", 1.5, "r1", 300); ok || c.At != 100 {
		t.Errorf("an equal strike replaced the first copy: %+v, %v", c, ok)
	}
	s := a.CoinStackOf("wonder", "stone_age")
	if len(s.Copies) != 1 || s.Count != 1 {
		t.Fatalf("stack has %d copies, count %d, want 1 and 1", len(s.Copies), s.Count)
	}
	// Another run adds a copy.
	strikeAt(a, "wonder", "stone_age", 1.5, "r2", 400)
	if s := a.CoinStackOf("wonder", "stone_age"); len(s.Copies) != 2 || s.Count != 2 {
		t.Errorf("a second run: %d copies, count %d", len(s.Copies), s.Count)
	}
	// A different type or age is its own stack.
	strikeAt(a, "techs", "stone_age", 1.5, "r1", 100)
	strikeAt(a, "wonder", "iron_age", 1.5, "r1", 100)
	if len(a.CoinStackOf("wonder", "stone_age").Copies) != 2 {
		t.Error("another stack changed this one")
	}
	if got := a.CoinStacksOfAge("stone_age"); len(got) != 2 || got[0].Type != "wonder" || got[1].Type != "techs" {
		t.Errorf("stacks of the age: %+v", got)
	}
	// Bad input strikes nothing.
	if _, ok := a.StrikeCoin("swift", "stone_age", 1, 1, "r1", time.Now()); ok {
		t.Error("an unknown type was struck")
	}
}

func TestCoinBetterLevelWinsInARun(t *testing.T) {
	isolateAccountDir(t)
	a := testAccount(t)
	strikeAt(a, "age", "bronze_age", 1.5, "r1", 100)
	c, ok := strikeAt(a, "age", "bronze_age", 0.5, "r1", 200)
	if !ok || c.Level != 6 {
		t.Fatalf("a better strike: %+v, %v", c, ok)
	}
	s := a.CoinStackOf("age", "bronze_age")
	if len(s.Copies) != 1 || s.Count != 1 || s.Copies[0].Level != 6 || s.Copies[0].At != 200 {
		t.Errorf("stack after a better strike: %+v", s)
	}
	if a.CoinBest("age", "bronze_age") != 6 || a.CoinBest("age", "iron_age") != 0 {
		t.Error("best level wrong")
	}
}

func TestCoinStackCapKeepsTheCountAndTheBest(t *testing.T) {
	isolateAccountDir(t)
	a := testAccount(t)
	// 30 runs at level 1, oldest first, then one at level 4.
	for i := 0; i < 30; i++ {
		strikeAt(a, "maximalist", "iron_age", 3, fmt.Sprintf("r%02d", i), int64(1000+i))
	}
	s := a.CoinStackOf("maximalist", "iron_age")
	if len(s.Copies) != coinStackMax || s.Count != 30 {
		t.Fatalf("%d copies, count %d, want %d and 30", len(s.Copies), s.Count, coinStackMax)
	}
	// The oldest went first.
	for _, c := range s.Copies {
		if c.At < 1006 {
			t.Errorf("a copy from %d outlived newer ones", c.At)
		}
	}
	if c, ok := strikeAt(a, "maximalist", "iron_age", 1, "best", 5000); !ok || c.Level != 4 {
		t.Fatalf("strike on a full stack: %+v, %v", c, ok)
	}
	s = a.CoinStackOf("maximalist", "iron_age")
	if len(s.Copies) != coinStackMax || s.Count != 31 || s.Copies[0].Level != 4 {
		t.Errorf("after the best: %d copies, count %d, top level %d", len(s.Copies), s.Count, s.Copies[0].Level)
	}
	// A worse coin into a full stack of finer ones is not kept, but is counted.
	for i := 0; i < coinStackMax; i++ {
		strikeAt(a, "maximalist", "iron_age", 0.5, fmt.Sprintf("fine%02d", i), int64(6000+i))
	}
	if _, ok := strikeAt(a, "maximalist", "iron_age", 3, "late", 9000); ok {
		t.Error("a coarse coin was kept in a stack of finer ones")
	}
	if s := a.CoinStackOf("maximalist", "iron_age"); s.Count != 31+coinStackMax+1 {
		t.Errorf("count %d does not include every strike", s.Count)
	}
}

func TestCoinSeedIsStable(t *testing.T) {
	a := coinSeed("acct", "run", "wonder", "stone_age")
	if a != coinSeed("acct", "run", "wonder", "stone_age") {
		t.Error("the same inputs gave different seeds")
	}
	for _, other := range []uint32{
		coinSeed("acct2", "run", "wonder", "stone_age"), coinSeed("acct", "run2", "wonder", "stone_age"),
		coinSeed("acct", "run", "techs", "stone_age"), coinSeed("acct", "run", "wonder", "iron_age"),
	} {
		if other == a {
			t.Error("a changed input kept the seed")
		}
	}
	if coinSeed("ab", "c", "t", "a") == coinSeed("a", "bc", "t", "a") {
		t.Error("the fields run together")
	}
	isolateAccountDir(t)
	acct := testAccount(t)
	c, _ := strikeAt(acct, "age", "stone_age", 1, "r1", 1)
	if c.Seed != coinSeed(acct.AccountID, "r1", "age", "stone_age") {
		t.Error("a struck copy carries another seed")
	}
}

func TestCoinFileRoundTripAndDamage(t *testing.T) {
	isolateAccountDir(t)
	acct := testAccount(t)
	strikeAt(acct, "age", "stone_age", 0.5, "a", 10)
	strikeAt(acct, "harbinger", "iron_age", 1.5, "b", 20)
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(accountDir(acct.AccountID), collectionFileName)
	good, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := reloadAccount(t)
	if len(got.coins) != 2 || got.CoinBest("age", "stone_age") != 6 || got.CoinStackOf("harbinger", "iron_age").Count != 1 {
		t.Fatalf("the coins changed in a save and load: %+v", got.coins)
	}
	if c := got.CoinStackOf("age", "stone_age").Copies[0]; c != acct.CoinStackOf("age", "stone_age").Copies[0] {
		t.Errorf("the copy changed in a round trip: %+v", c)
	}
	var f collectionFile
	if err := json.Unmarshal(good, &f); err != nil {
		t.Fatal(err)
	}
	f.Stacks[0].Copies[0].Level = 1 // edited, signature kept
	edited, _ := json.Marshal(&f)
	foreign := collectionFile{Version: collectionVersion, AccountID: "ffffffffffffffffffffffffffffffff", Stacks: f.Stacks}
	foreign.Signature = signCollection(&foreign)
	foreignData, _ := json.Marshal(&foreign)
	for name, data := range map[string][]byte{"garbage": []byte("{not json"), "bad signature": edited, "another account's": foreignData, "empty": {}} {
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		got := reloadAccount(t)
		if len(got.coins) != 0 || got.Tampered || got.BadgesTampered {
			t.Errorf("%s: %d stacks, tampered %v/%v", name, len(got.coins), got.Tampered, got.BadgesTampered)
		}
		// Overwritten at the next strike.
		strikeAt(got, "wonder", "stone_age", 1, "c", 30)
		if err := got.FlushIfDirty(); err != nil {
			t.Fatal(err)
		}
		if again := reloadAccount(t); len(again.coins) != 1 || again.CoinBest("wonder", "stone_age") != 4 {
			t.Errorf("%s: the file was not rewritten: %d stacks", name, len(again.coins))
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if got := reloadAccount(t); len(got.coins) != 0 {
		t.Errorf("a missing file loaded %d stacks", len(got.coins))
	}
}

func TestCoinMergeUnionsByRunAndTakesTheFiner(t *testing.T) {
	isolateAccountDir(t)
	mine := testAccount(t)
	strikeAt(mine, "age", "stone_age", 1.5, "r1", 10) // level 2
	strikeAt(mine, "age", "stone_age", 1.5, "r2", 20)
	theirs := &Account{AccountID: mine.AccountID}
	strikeAt(theirs, "age", "stone_age", 0.5, "r1", 15) // level 6, same run
	strikeAt(theirs, "age", "stone_age", 1.5, "r3", 30)
	strikeAt(theirs, "wonder", "iron_age", 1.5, "r1", 40)
	f := theirs.collectionLocked()
	if !verifyCollection(f) {
		t.Fatal("the file does not verify")
	}
	mine.mu.Lock()
	mine.takeCollectionLocked(f, true)
	mine.mu.Unlock()
	s := mine.CoinStackOf("age", "stone_age")
	if len(s.Copies) != 3 || s.Copies[0].Level != 6 || s.Copies[0].Run != "r1" {
		t.Errorf("merged stack: %+v", s)
	}
	if s.Count != 3 || mine.CoinBest("wonder", "iron_age") != 2 {
		t.Errorf("count %d, wonder best %d", s.Count, mine.CoinBest("wonder", "iron_age"))
	}
	// Merging the same file again adds nothing.
	mine.mu.Lock()
	mine.takeCollectionLocked(f, true)
	mine.mu.Unlock()
	if s := mine.CoinStackOf("age", "stone_age"); len(s.Copies) != 3 || s.Count != 3 {
		t.Errorf("a repeat merge: %d copies, count %d", len(s.Copies), s.Count)
	}
	// A replace takes the export's, and an export with none changes nothing.
	mine.mu.Lock()
	mine.takeCollectionLocked(f, false)
	mine.takeCollectionLocked(nil, false)
	mine.mu.Unlock()
	if s := mine.CoinStackOf("age", "stone_age"); len(s.Copies) != 2 {
		t.Errorf("a replace left %d copies", len(s.Copies))
	}
}

func TestCoinExportCarriesTheCollection(t *testing.T) {
	isolateAccountDir(t)
	acct := testAccount(t)
	strikeAt(acct, "techs", "iron_age", 0.5, "r1", 10)
	acct.mu.Lock()
	f := acct.collectionLocked()
	acct.mu.Unlock()
	if !verifyCollection(f) || f.AccountID != acct.AccountID || len(f.Stacks) != 1 {
		t.Errorf("export file: %+v", f)
	}
	f.Stacks[0].Copies[0].Level = 1
	if verifyCollection(f) {
		t.Error("an edited file verifies")
	}
}
