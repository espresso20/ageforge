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
	noErr(t, acct.FlushIfDirty())

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
	acct, err := CreateAccount("Ada")
	if err != nil {
		t.Fatal(err)
	}
	acct.mu.Lock()
	acct.Badges = map[string]BadgeEarned{stoneRung1: {At: 100, Count: coreDef(t, stoneRung1).Threshold}}
	acct.dirty = true
	acct.mu.Unlock()
	noErr(t, acct.Save())
	id := acct.AccountID
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
	noErr(t, json.Unmarshal(raw, &old))
	out, _ := json.Marshal(old)
	var asOld badgeFile
	noErr(t, json.Unmarshal(out, &asOld))
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
