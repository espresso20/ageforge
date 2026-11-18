package game

import (
	"bytes"
	"encoding/json"
	"testing"

)

// A badge remembers the bar it was earned at (card s3NHfReX). Every test
// runs in a temp data root; no real account, badge file or save is read.

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
	acct.Badges = map[string]BadgeEarned{badgeCollector: {At: 100, Count: 15}}
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
