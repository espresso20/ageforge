package game

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// A badge remembers the bar it was earned at (card s3NHfReX). Every test
// runs in a temp data root; no real account, badge file or save is read.

const (
	stoneRung1 = "resource.stone.1"
	stoneRung2 = "resource.stone.2"
	stoneRung3 = "resource.stone.3"
	stoneCount = "produced.stone"
)

func coreDef(t *testing.T, key string) config.BadgeDef {
	t.Helper()
	def, ok := rules.Core().Badge(key)
	if !ok {
		t.Fatalf("no badge %s", key)
	}
	return def
}

// writeOldFile makes a slot whose badge file is as the code from before this
// change wrote it: rungs with no remembered count, signed. The account holds
// the two lower stone rungs and the third, and a counter that meets the first
// two only; an age badge sits beside them.
func writeOldFile(t *testing.T) (id string) {
	t.Helper()
	acct, err := CreateAccount("Ada")
	if err != nil {
		t.Fatal(err)
	}
	r2 := coreDef(t, stoneRung2)
	acct.mu.Lock()
	acct.Badges = map[string]BadgeEarned{
		stoneRung1:          {At: 100},
		stoneRung2:          {At: 200},
		stoneRung3:          {At: 300, Run: "old"},
		"age.stone_age":     {At: 50},
		"lineage.housing.1": {At: 60},
	}
	acct.Counters = map[string]float64{stoneCount: r2.Threshold * 2, "built.lineage.housing": config.Billion}
	acct.dirty = true
	acct.mu.Unlock()
	if err := acct.Save(); err != nil {
		t.Fatal(err)
	}
	return acct.AccountID
}

func TestOldRungsAreSettledOnce(t *testing.T) {
	isolateAccountDir(t)
	id := writeOldFile(t)
	if raw := slotBadgeFile(t, id); bytes.Contains(raw, []byte(`"count"`)) {
		t.Fatalf("the old-style file already remembers counts:\n%s", raw)
	}

	ge, acct := openIn(t, id)
	_ = ge
	if acct.BadgesTampered {
		t.Fatal("a badge file from before this change loads flagged as edited")
	}
	if hasBadge(acct, stoneRung3) {
		t.Error("the third stone rung, earned under a lower bar, is still earned")
	}
	for _, key := range []string{stoneRung1, stoneRung2} {
		if !hasBadge(acct, key) {
			t.Errorf("%s was lost", key)
		}
		if got, want := acct.Badges[key].Count, coreDef(t, key).Threshold; got != want {
			t.Errorf("%s remembers %v, want today's %v", key, got, want)
		}
	}
	for _, key := range []string{"age.stone_age", "lineage.housing.1"} {
		if !hasBadge(acct, key) {
			t.Errorf("%s was touched", key)
		}
	}
	if acct.Badges["age.stone_age"].Count != 0 {
		t.Error("an age badge remembers a count")
	}

	// The cleanup was saved, and the file is valid and says nothing wrong.
	var bf badgeFile
	raw := slotBadgeFile(t, id)
	if err := json.Unmarshal(raw, &bf); err != nil || !verifyBadgeFile(&bf) || bf.Tampered {
		t.Fatalf("the file after the cleanup: err %v, verifies %v, tampered %v", err, verifyBadgeFile(&bf), bf.Tampered)
	}
	if _, ok := bf.Badges[stoneRung3]; ok {
		t.Error("the file still holds the un-earned rung")
	}

	// Once: another load changes nothing, in memory or on disk.
	_, again := openIn(t, id)
	if string(slotBadgeFile(t, id)) != string(raw) {
		t.Error("a second load rewrote the badge file")
	}
	if len(again.Badges) != len(acct.Badges) || again.BadgesTampered {
		t.Errorf("a second load: %d badges (was %d), tampered %v", len(again.Badges), len(acct.Badges), again.BadgesTampered)
	}
}

// TestRungKeepsItsBarWhenTheBarRises: a rung earned after this change keeps
// its count, and the words that go with it, when the ruleset's bar goes up.
func TestRungKeepsItsBarWhenTheBarRises(t *testing.T) {
	isolateAccountDir(t)
	acct, err := CreateAccount("Bea")
	if err != nil {
		t.Fatal(err)
	}
	book := coreBook()
	r3 := coreDef(t, stoneRung3)
	acct.judge(book, Event{Kind: config.BadgeEvProduced, Subject: "stone", N: r3.Threshold}, badgeCtx{})
	if !hasBadge(acct, stoneRung3) {
		t.Fatalf("precondition: producing %v stone did not earn %s: %v", r3.Threshold, stoneRung3, acct.EarnedBadges())
	}
	if got := acct.Badges[stoneRung3].Count; got != r3.Threshold {
		t.Fatalf("the earned rung remembers %v, want %v", got, r3.Threshold)
	}
	wasText, _ := rules.Core().BadgeDescAt(stoneRung3, r3.Threshold)
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}

	// Raise every production ladder's top through the ruleset.
	src := rules.FromConfig()
	for i := range src.BadgeFamilies {
		if src.BadgeFamilies[i].Measure == config.BadgeMeasureProduction {
			src.BadgeFamilies[i].TopRuns *= 10
		}
	}
	raised := rules.Compile(src)
	rd, ok := raised.Badge(stoneRung3)
	if !ok || rd.Threshold <= r3.Threshold {
		t.Fatalf("precondition: the raised bar is %v, today's is %v", rd.Threshold, r3.Threshold)
	}
	book2 := newBadgeBook(raised)

	// It loads under the raised ruleset still earned, and says what it did.
	reloaded := slotAccount(t, acct.AccountID)
	NewGameEngineWith(raised).SetAccount(reloaded)
	if !hasBadge(reloaded, stoneRung3) {
		t.Fatal("the rung was un-earned when the bar rose")
	}
	v, _ := reloaded.badgeView(book2, stoneRung3)
	if v.Desc != wasText {
		t.Errorf("the earned rung says %q, want the text it was earned at: %q", v.Desc, wasText)
	}
	// A rung not earned shows the raised bar.
	r4, _ := raised.Badge("resource.stone.4")
	if _, has := reloaded.Badges["resource.stone.4"]; has {
		t.Fatal("precondition: stone.4 is earned")
	}
	if v4, _ := reloaded.badgeView(book2, "resource.stone.4"); !strings.Contains(v4.Desc, badgeCountText(raised, "resource.stone.4", r4.Threshold)) {
		t.Errorf("an unearned rung says %q, want today's count", v4.Desc)
	}
	if reloaded.BadgesTampered {
		t.Error("a file with a remembered count loads flagged as edited")
	}
}

func badgeCountText(set *rules.Set, key string, count float64) string {
	d, _ := set.BadgeDescAt(key, count)
	return d
}

// TestRememberedCountIsNotSigned: a build from before the field drops it on
// reading, and must still find the signature it expects; the new build reads
// a file with the field and a file without it alike.
func TestRememberedCountIsNotSigned(t *testing.T) {
	isolateAccountDir(t)
	id := writeOldFile(t)
	openIn(t, id) // settles and saves with counts
	raw := slotBadgeFile(t, id)
	if !bytes.Contains(raw, []byte(`"count"`)) {
		t.Fatalf("no remembered count in the file:\n%s", raw)
	}
	var bf badgeFile
	if err := json.Unmarshal(raw, &bf); err != nil || !verifyBadgeFile(&bf) {
		t.Fatalf("a file with counts does not verify: %v", err)
	}
	// What a build from before the field holds after reading the file.
	type oldEarned struct {
		At    int64  `json:"at"`
		Run   string `json:"run,omitempty"`
		Flags int    `json:"f,omitempty"`
	}
	var old struct {
		Version   int                  `json:"version"`
		AccountID string               `json:"account_id"`
		Badges    map[string]oldEarned `json:"badges,omitempty"`
		Counters  map[string]float64   `json:"counters,omitempty"`
		Days      []string             `json:"days,omitempty"`
		Tampered  bool                 `json:"tampered,omitempty"`
		Signature string               `json:"_sig,omitempty"`
	}
	if err := json.Unmarshal(raw, &old); err != nil {
		t.Fatal(err)
	}
	out, _ := json.Marshal(old)
	var asOld badgeFile
	if err := json.Unmarshal(out, &asOld); err != nil {
		t.Fatal(err)
	}
	for _, e := range asOld.Badges {
		if e.Count != 0 {
			t.Fatal("the stand-in for the old build kept a count")
		}
	}
	if !verifyBadgeFile(&asOld) {
		t.Error("a build from before the field would find the signature wrong")
	}
	// Edits still show: change a badge's time and the file is flagged.
	editBadgeFile(t, id, func(m map[string]json.RawMessage) {
		m["counters"] = json.RawMessage(`{"produced.stone": 1}`)
	})
	if again := slotAccount(t, id); !again.BadgesTampered {
		t.Error("editing the counters of a file with counts was not noticed")
	}
}

// TestRungsThatGiveSomethingAreLeftAlone: the cleanup does not take back a
// rung a theme or a title comes from, even when its counter is under today's
// bar; whether it should is the owner's call.
func TestRungsThatGiveSomethingAreLeftAlone(t *testing.T) {
	isolateAccountDir(t)
	set := rules.Core()
	acct, err := CreateAccount("Cy")
	if err != nil {
		t.Fatal(err)
	}
	var given []string
	acct.mu.Lock()
	acct.Badges = map[string]BadgeEarned{}
	for _, def := range set.Badges() {
		if _, counted := set.BadgeDescAt(def.Key, def.Threshold); !counted || def.Ladder == "" || def.Scope != config.BadgeLifetime {
			continue
		}
		if def.Reward.Theme != "" || def.Reward.Title != "" {
			given = append(given, def.Key)
			acct.Badges[def.Key] = BadgeEarned{At: 10}
		}
	}
	acct.dirty = true
	acct.mu.Unlock()
	if len(given) == 0 {
		t.Fatal("no counted ladder rung gives a theme or a title")
	}
	if err := acct.Save(); err != nil {
		t.Fatal(err)
	}
	_, loaded := openIn(t, acct.AccountID)
	for _, key := range given {
		if !hasBadge(loaded, key) {
			t.Errorf("%s gives %+v and was un-earned", key, coreDef(t, key).Reward)
		}
	}
}
