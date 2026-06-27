package game

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestSave creates a real, signed, loadable save under the given base name
// via the engine, and registers cleanup to remove it (and any "-copy" variants a
// test may produce). It returns the engine so callers can reuse it for loads.
func writeTestSave(t *testing.T, name string) *GameEngine {
	t.Helper()
	ge := NewGameEngine()
	if err := ge.SaveGame(name); err != nil {
		t.Fatalf("SaveGame(%q) failed: %v", name, err)
	}
	if !SaveExists(name) {
		t.Fatalf("save %q not found after SaveGame", name)
	}
	t.Cleanup(func() {
		// Best-effort cleanup; ignore errors (file may already be gone/renamed).
		_ = os.Remove(savePath(name))
		_ = os.Remove(savePath(name + "-copy"))
		_ = os.Remove(savePath(name + "-copy-2"))
	})
	return ge
}

func TestSanitizeSaveName(t *testing.T) {
	ok := []struct{ in, want string }{
		{"my_save", "my_save"},
		{"  spaced  ", "spaced"},
		{"game.json", "game"}, // trailing .json stripped
		{"save-copy-2", "save-copy-2"},
	}
	for _, tc := range ok {
		got, err := sanitizeSaveName(tc.in)
		if err != nil {
			t.Errorf("sanitizeSaveName(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("sanitizeSaveName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}

	bad := []string{
		"",
		"   ",
		"../escape",
		"foo/bar",
		`foo\bar`,
		"..",
		"a..b",
		".hidden",
		string([]byte{'a', 0, 'b'}),
		strings.Repeat("x", maxSaveNameLen+1),
	}
	for _, in := range bad {
		if _, err := sanitizeSaveName(in); err == nil {
			t.Errorf("sanitizeSaveName(%q) = nil error, want rejection", in)
		}
	}
}

func TestDeleteSave(t *testing.T) {
	name := "test_delete_save"
	writeTestSave(t, name)

	if err := DeleteSave(name); err != nil {
		t.Fatalf("DeleteSave(%q) failed: %v", name, err)
	}
	if SaveExists(name) {
		t.Errorf("save %q still exists after DeleteSave", name)
	}

	// Missing file → clear error.
	if err := DeleteSave("test_delete_does_not_exist"); err == nil {
		t.Error("DeleteSave on missing file = nil error, want error")
	}
}

func TestDeleteSaveRejectsTraversal(t *testing.T) {
	// Sentinel file in a temp dir that a traversal name could target. It must
	// survive a DeleteSave call with a crafted path-traversal name.
	tmp := t.TempDir()
	sentinel := filepath.Join(tmp, "passwd")
	if err := os.WriteFile(sentinel, []byte("do-not-delete"), 0644); err != nil {
		t.Fatalf("failed to write sentinel: %v", err)
	}

	traversal := "../../../../../../../../" + tmp + "/passwd"
	if err := DeleteSave(traversal); err == nil {
		t.Errorf("DeleteSave(%q) = nil error, want rejection", traversal)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Errorf("sentinel file was touched/removed by traversal name: %v", err)
	}

	// A couple more crafted names for good measure.
	for _, bad := range []string{"../../etc/passwd", `..\..\secrets`} {
		if err := DeleteSave(bad); err == nil {
			t.Errorf("DeleteSave(%q) = nil error, want rejection", bad)
		}
	}
}

func TestRenameSave(t *testing.T) {
	oldName := "test_rename_old"
	newName := "test_rename_new"
	writeTestSave(t, oldName)
	t.Cleanup(func() { _ = os.Remove(savePath(newName)) })

	if err := RenameSave(oldName, newName); err != nil {
		t.Fatalf("RenameSave(%q,%q) failed: %v", oldName, newName, err)
	}
	if SaveExists(oldName) {
		t.Errorf("old save %q still exists after rename", oldName)
	}
	if !SaveExists(newName) {
		t.Fatalf("new save %q missing after rename", newName)
	}
	// Renamed file must still load (signature preserved — bytes unchanged).
	if err := NewGameEngine().LoadGame(newName); err != nil {
		t.Errorf("LoadGame(%q) after rename failed: %v", newName, err)
	}

	// Collision: renaming onto an existing save must be rejected.
	other := "test_rename_other"
	writeTestSave(t, other)
	if err := RenameSave(other, newName); err == nil {
		t.Errorf("RenameSave onto existing %q = nil error, want rejection", newName)
	}
	if !SaveExists(other) {
		t.Errorf("source %q should be untouched after a rejected collision", other)
	}

	// Invalid target name.
	if err := RenameSave(newName, "../escape"); err == nil {
		t.Error("RenameSave to invalid name = nil error, want rejection")
	}
	// Missing source.
	if err := RenameSave("test_rename_missing", "whatever"); err == nil {
		t.Error("RenameSave with missing source = nil error, want error")
	}
}

func TestDuplicateSave(t *testing.T) {
	name := "test_dup_save"
	writeTestSave(t, name)

	first, err := DuplicateSave(name)
	if err != nil {
		t.Fatalf("DuplicateSave(%q) failed: %v", name, err)
	}
	if first != name+"-copy" {
		t.Errorf("first duplicate name = %q, want %q", first, name+"-copy")
	}
	if !SaveExists(first) {
		t.Fatalf("first duplicate %q missing", first)
	}
	// Both original and copy must load.
	if err := NewGameEngine().LoadGame(name); err != nil {
		t.Errorf("LoadGame(original %q) failed: %v", name, err)
	}
	if err := NewGameEngine().LoadGame(first); err != nil {
		t.Errorf("LoadGame(copy %q) failed: %v", first, err)
	}

	// Second duplicate must bump to "-copy-2".
	second, err := DuplicateSave(name)
	if err != nil {
		t.Fatalf("second DuplicateSave(%q) failed: %v", name, err)
	}
	if second != name+"-copy-2" {
		t.Errorf("second duplicate name = %q, want %q", second, name+"-copy-2")
	}
	if !SaveExists(second) {
		t.Errorf("second duplicate %q missing", second)
	}

	// Missing source.
	if _, err := DuplicateSave("test_dup_missing"); err == nil {
		t.Error("DuplicateSave on missing source = nil error, want error")
	}
}

func TestListSaveDetailsPopulatesFields(t *testing.T) {
	name := "test_details_save"
	ge := writeTestSave(t, name)
	// Give the save a non-trivial tick so we can assert it round-trips.
	ge.mu.Lock()
	ge.tick = 1234
	ge.mu.Unlock()
	if err := ge.SaveGame(name); err != nil {
		t.Fatalf("re-SaveGame failed: %v", err)
	}

	details, err := ListSaveDetails()
	if err != nil {
		t.Fatalf("ListSaveDetails failed: %v", err)
	}
	var found *SaveInfo
	for i := range details {
		if details[i].Name == name {
			found = &details[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("save %q not present in ListSaveDetails output", name)
	}
	if found.Corrupt {
		t.Errorf("freshly written save %q flagged Corrupt", name)
	}
	if found.Tick != 1234 {
		t.Errorf("Tick = %d, want 1234", found.Tick)
	}
	if found.Age == "" {
		t.Errorf("Age is empty, want the engine's starting age")
	}
	if found.Timestamp.IsZero() {
		t.Errorf("Timestamp is zero, want the save's timestamp")
	}
	// A fresh engine has full morale; ensure the field was parsed (non-zero).
	if found.Morale <= 0 {
		t.Errorf("Morale = %v, want > 0", found.Morale)
	}
}

func TestListSaveDetailsFlagsCorrupt(t *testing.T) {
	// Ensure the dir exists, then drop a non-JSON .json file into it.
	good := "test_corrupt_neighbor"
	writeTestSave(t, good) // guarantees the save dir exists

	dir := saveDirectory()
	corruptName := "test_corrupt_file"
	corruptPath := filepath.Join(dir, corruptName+".json")
	if err := os.WriteFile(corruptPath, []byte("this is not json {{{"), 0644); err != nil {
		t.Fatalf("failed to write corrupt save: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(corruptPath) })

	details, err := ListSaveDetails()
	if err != nil {
		t.Fatalf("ListSaveDetails failed (should not error on a corrupt file): %v", err)
	}

	var corrupt, healthy *SaveInfo
	for i := range details {
		switch details[i].Name {
		case corruptName:
			corrupt = &details[i]
		case good:
			healthy = &details[i]
		}
	}
	if corrupt == nil {
		t.Fatalf("corrupt save %q not present in output (should be included, not dropped)", corruptName)
	}
	if !corrupt.Corrupt {
		t.Errorf("corrupt save %q not flagged Corrupt", corruptName)
	}
	if corrupt.Timestamp.IsZero() {
		t.Errorf("corrupt save Timestamp is zero, want file mtime fallback")
	}
	// The healthy neighbour must still be listed and parsed fine — one bad file
	// must not poison the rest.
	if healthy == nil {
		t.Errorf("healthy save %q dropped from listing alongside a corrupt file", good)
	} else if healthy.Corrupt {
		t.Errorf("healthy save %q wrongly flagged Corrupt", good)
	}
}
