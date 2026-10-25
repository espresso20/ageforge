package game

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// forgeProof is the elite proof worked out here from the key, apart from proofOf.
func forgeProof(sig string) string {
	mac := hmac.New(sha256.New, []byte("ageforge-master-forger-v1"))
	mac.Write([]byte(sig))
	return hex.EncodeToString(mac.Sum(nil))
}

// TestSaveGame_EliteProof covers that a forge master's save is written with a proof that
// verifies, that moving the save under a new parent re-signs it and keeps the proof valid,
// and that a plain save gets none.
func TestSaveGame_EliteProof(t *testing.T) {
	isolateAccountDir(t)
	elite := NewGameEngine()
	elite.eliteBadge = true
	if err := elite.SaveGame("forged"); err != nil {
		t.Fatal(err)
	}
	plain := NewGameEngine()
	if err := plain.SaveGame("plain"); err != nil {
		t.Fatal(err)
	}

	gs := readSaveFromDisk(t, "forged")
	if gs.Proof == "" || gs.Proof != forgeProof(gs.Signature) {
		t.Fatalf("proof %q for signature %q, want %q", gs.Proof, gs.Signature, forgeProof(gs.Signature))
	}
	if sigOK, eliteOK := verifySave(&gs); !sigOK || !eliteOK {
		t.Errorf("forged save verifies as sig=%v elite=%v", sigOK, eliteOK)
	}
	if gs := readSaveFromDisk(t, "plain"); gs.Proof != "" {
		t.Errorf("a plain save carries the proof %q", gs.Proof)
	}

	for _, name := range []string{"forged", "plain"} {
		before := readSaveFromDisk(t, name)
		if err := reparentSaveFile(name, "Elder"); err != nil {
			t.Fatal(err)
		}
		after := readSaveFromDisk(t, name)
		sigOK, eliteOK := verifySave(&after)
		if after.ParentName != "Elder" || !sigOK || eliteOK != (before.Proof != "") {
			t.Errorf("%s after reparent: parent %q sig=%v elite=%v (proof before %q)", name, after.ParentName, sigOK, eliteOK, before.Proof)
		}
		if after.Signature == before.Signature {
			t.Errorf("%s: reparenting left the signature alone", name)
		}
		if (before.Proof == "") != (after.Proof == "") {
			t.Errorf("%s: proof before %q, after %q", name, before.Proof, after.Proof)
		}
	}
}

// TestDuplicateSave_LeavesNoTempFile covers that a copy is written whole and the temp file
// is gone.
func TestDuplicateSave_LeavesNoTempFile(t *testing.T) {
	isolateAccountDir(t)
	if err := NewGameEngine().SaveGame("orig"); err != nil {
		t.Fatal(err)
	}
	dup, err := DuplicateSave("orig")
	if err != nil {
		t.Fatal(err)
	}
	a, _ := os.ReadFile(savePath("orig"))
	b, _ := os.ReadFile(savePath(dup))
	if len(a) == 0 || string(a) != string(b) {
		t.Errorf("the copy is not the original's bytes (%d and %d bytes)", len(a), len(b))
	}
	if matches, _ := filepath.Glob(filepath.Join(filepath.Dir(savePath("orig")), "*.tmp")); len(matches) != 0 {
		t.Errorf("temp files left behind: %v", matches)
	}
}

// TestProofOf pins the elite proof to fixed pairs worked out outside the code.
func TestProofOf(t *testing.T) {
	for sig, want := range map[string]string{
		"abc":              "66cdd7e2f86ca7f033f19ec67dfe30a372c24fa67aed5f9ac3bc0073af313192",
		"0123456789abcdef": "a0a081676bfe779994e2688b0ed60ad6313843d98d861ca01db0ca88a1189ea5",
	} {
		if got := proofOf(sig); got != want {
			t.Errorf("proofOf(%q) = %s, want %s", sig, got, want)
		}
	}
}

// TestWriteFileAtomic covers a file written whole with no temp file left, an old file
// replaced, and the temp file removed when either step fails.
func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.json")
	for _, content := range []string{"first", "second, longer"} {
		if err := writeFileAtomic(path, []byte(content)); err != nil {
			t.Fatal(err)
		}
		if got, _ := os.ReadFile(path); string(got) != content {
			t.Errorf("file holds %q, want %q", got, content)
		}
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temp file left after a good write: %v", err)
	}

	// The rename fails when the target is a folder: the temp file must go.
	blocked := filepath.Join(dir, "blocked.json")
	if err := os.Mkdir(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(blocked, []byte("x")); err == nil {
		t.Error("writing over a folder did not fail")
	}
	if _, err := os.Stat(blocked + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temp file left after a failed rename: %v", err)
	}

	// The write fails when the folder is missing.
	if err := writeFileAtomic(filepath.Join(dir, "missing", "b.json"), []byte("x")); err == nil {
		t.Error("writing into a missing folder did not fail")
	}
}

// TestSaveFiles covers that only .json files come back, folders and other files are left
// out, a missing folder holds none, and the lists built on it agree.
func TestSaveFiles(t *testing.T) {
	isolateAccountDir(t)
	if files, err := saveFiles(filepath.Join(t.TempDir(), "nope")); files != nil || err != nil {
		t.Errorf("missing folder: %v, %v", files, err)
	}
	if names, err := ListSaves(); len(names) != 0 || err != nil {
		t.Errorf("no saves yet: %v, %v", names, err)
	}

	dir := saveDirectory()
	if err := os.MkdirAll(filepath.Join(dir, "folder.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"b.json", "a.json", "notes.txt", "a.json.tmp"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	files, err := saveFiles(dir)
	if err != nil || len(files) != 2 || files[0].Name() != "a.json" || files[1].Name() != "b.json" {
		t.Fatalf("saveFiles = %v, %v; want a.json, b.json", files, err)
	}
	if names, _ := ListSaves(); len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("ListSaves = %v, want [a b]", names)
	}
	if details, _ := ListSaveDetails(); len(details) != 2 || details[0].Name != "a" || details[1].Name != "b" {
		t.Errorf("ListSaveDetails = %+v, want a then b", details)
	}
	if err := WipeAllSaves(); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"folder.json", "notes.txt", "a.json.tmp"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("WipeAllSaves removed %s: %v", f, err)
		}
	}
	if names, _ := ListSaves(); len(names) != 0 {
		t.Errorf("saves left after WipeAllSaves: %v", names)
	}
}
