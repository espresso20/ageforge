package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
)

// writeRawSave marshals a GameSave and writes it (unsigned) to the save path for
// name, overwriting any existing file. ListSaveDetails reads JSON only and never
// verifies signatures, so an unsigned hand-built save is a valid parse fixture.
func writeRawSave(t *testing.T, name string, save GameSave) {
	t.Helper()
	data, err := json.Marshal(save)
	if err != nil {
		t.Fatalf("marshal raw save: %v", err)
	}
	if err := os.WriteFile(savePath(name), data, 0644); err != nil {
		t.Fatalf("write raw save %q: %v", name, err)
	}
}

// findDetail runs ListSaveDetails and returns the entry for name, failing if it
// is absent.
func findDetail(t *testing.T, name string) SaveInfo {
	t.Helper()
	details, err := ListSaveDetails()
	if err != nil {
		t.Fatalf("ListSaveDetails failed: %v", err)
	}
	for i := range details {
		if details[i].Name == name {
			return details[i]
		}
	}
	t.Fatalf("save %q not present in ListSaveDetails output", name)
	return SaveInfo{}
}

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

// pickBuildingKeys returns a real wonder key and a real non-wonder building key
// from config, so the rich-fields test stays valid as the building set evolves.
func pickBuildingKeys(t *testing.T) (wonder, other string) {
	t.Helper()
	for k, d := range config.BuildingByKey() {
		if d.Category == "wonder" && wonder == "" {
			wonder = k
		}
		if d.Category != "wonder" && other == "" {
			other = k
		}
	}
	if wonder == "" || other == "" {
		t.Fatalf("could not find a wonder (%q) and non-wonder (%q) building in config", wonder, other)
	}
	return wonder, other
}

// TestListSaveDetailsRichFields writes a real save, then rewrites it with a
// hand-built header carrying buildings/workers/resources/title/epoch/prestige so
// the derivation + parse logic in ListSaveDetails is exercised directly. The save
// is rewritten unsigned, which is fine: ListSaveDetails never verifies signatures.
func TestListSaveDetailsRichFields(t *testing.T) {
	name := "test_rich_details"
	writeTestSave(t, name) // ensures the dir exists and registers cleanup

	wonderKey, prodKey := pickBuildingKeys(t)

	// 5 production + 1 wonder = 6 building instances (the count-0 key is excluded
	// from both totals). 4 + 6 = 10 workers.
	save := GameSave{
		Timestamp:          time.Now(),
		Tick:               4242,
		Age:                "iron_age",
		Morale:             0.75,
		CurrentTitle:       "The Eternal",
		CurrentEpoch:       "iron_era",
		PendingCatastrophe: "iron_era", // pending_catastrophe holds an epoch key
		Milestones:         []string{"m1", "m2", "m3"},
		Resources:          map[string]float64{"soldiers": 137.9, "wood": 50},
		Buildings:          map[string]int{prodKey: 5, wonderKey: 1, "ghost_key_zero": 0},
		Workers: map[string]WorkerInfo{
			"food":     {Count: 4},
			"military": {Count: 6},
		},
		Research: ResearchSave{Researched: []string{"t1", "t2", "t3", "t4"}},
		Prestige: PrestigeSave{Level: 2, TotalEarned: 1500},
	}

	writeRawSave(t, name, save)

	found := findDetail(t, name)
	if found.Corrupt {
		t.Fatalf("rich save %q flagged Corrupt", name)
	}
	if found.Title != "The Eternal" {
		t.Errorf("Title = %q, want %q", found.Title, "The Eternal")
	}
	// Epoch should be the display name, not the raw key.
	if found.Epoch == "" || found.Epoch == "iron_era" {
		t.Errorf("Epoch = %q, want a display name resolved from %q", found.Epoch, "iron_era")
	}
	if found.Population != 10 {
		t.Errorf("Population = %d, want 10 (4 food + 6 military)", found.Population)
	}
	if found.Buildings != 6 { // 5 prod + 1 wonder; the count-0 key is excluded
		t.Errorf("Buildings = %d, want 6", found.Buildings)
	}
	if found.Wonders != 1 {
		t.Errorf("Wonders = %d, want 1", found.Wonders)
	}
	if found.Techs != 4 {
		t.Errorf("Techs = %d, want 4", found.Techs)
	}
	if found.Soldiers != 137 { // int() truncation of 137.9
		t.Errorf("Soldiers = %d, want 137", found.Soldiers)
	}
	if found.PrestigeLevel != 2 {
		t.Errorf("PrestigeLevel = %d, want 2", found.PrestigeLevel)
	}
	if found.PrestigeTotal != 1500 {
		t.Errorf("PrestigeTotal = %d, want 1500", found.PrestigeTotal)
	}
	if found.MilestonesDone != 3 {
		t.Errorf("MilestonesDone = %d, want 3", found.MilestonesDone)
	}
	if found.MilestonesTotal != len(config.Milestones()) {
		t.Errorf("MilestonesTotal = %d, want %d", found.MilestonesTotal, len(config.Milestones()))
	}
	if found.PendingCatastrophe == "" || found.PendingCatastrophe == "iron_era" {
		t.Errorf("PendingCatastrophe = %q, want a catastrophe display name resolved from %q", found.PendingCatastrophe, "iron_era")
	}
	if found.Tick != 4242 {
		t.Errorf("Tick = %d, want 4242", found.Tick)
	}
}

func TestActiveSaveNameDefaultsAndSetter(t *testing.T) {
	// A fresh engine has never saved or loaded → bare `save` targets the autosave.
	ge := NewGameEngine()
	if got := ge.ActiveSaveName(); got != AutosaveName {
		t.Errorf("fresh ActiveSaveName() = %q, want %q", got, AutosaveName)
	}
	// An explicit `save <name>` records the slot for future bare saves.
	ge.SetActiveSaveName("test")
	if got := ge.ActiveSaveName(); got != "test" {
		t.Errorf("after SetActiveSaveName(\"test\"), ActiveSaveName() = %q, want \"test\"", got)
	}
}

func TestLoadGameSetsActiveSaveName(t *testing.T) {
	name := "test_active_load"
	ge := writeTestSave(t, name) // writes name.json, registers cleanup

	// A different engine that loads the slot should adopt it as its active name,
	// so a subsequent bare `save` overwrites the loaded slot rather than autosave.
	other := NewGameEngine()
	if got := other.ActiveSaveName(); got != AutosaveName {
		t.Fatalf("pre-load ActiveSaveName() = %q, want %q", got, AutosaveName)
	}
	if err := other.LoadGame(name); err != nil {
		t.Fatalf("LoadGame(%q) failed: %v", name, err)
	}
	if got := other.ActiveSaveName(); got != name {
		t.Errorf("after LoadGame(%q), ActiveSaveName() = %q, want %q", name, got, name)
	}

	// A failed load must NOT change the active slot.
	if err := ge.LoadGame("test_active_load_missing"); err == nil {
		t.Fatal("LoadGame on missing file = nil error, want error")
	}
	if got := ge.ActiveSaveName(); got != AutosaveName {
		t.Errorf("after failed LoadGame, ActiveSaveName() = %q, want %q (unchanged)", got, AutosaveName)
	}
}

func TestParentNameRoundTrips(t *testing.T) {
	name := "test_lineage_child"
	ge := NewGameEngine()
	t.Cleanup(func() { _ = os.Remove(savePath(name)) })

	// Same-package test → set the unexported field directly (no setter needed).
	ge.activeParentName = "RootGame"
	if err := ge.SaveGame(name); err != nil {
		t.Fatalf("SaveGame(%q) failed: %v", name, err)
	}

	// A fresh engine loading the child must adopt the persisted parent.
	loaded := NewGameEngine()
	if err := loaded.LoadGame(name); err != nil {
		t.Fatalf("LoadGame(%q) failed: %v", name, err)
	}
	if loaded.activeParentName != "RootGame" {
		t.Errorf("activeParentName after load = %q, want %q", loaded.activeParentName, "RootGame")
	}

	// ListSaveDetails must surface the same parent for Phase 3's tree.
	if got := findDetail(t, name).ParentName; got != "RootGame" {
		t.Errorf("SaveInfo.ParentName = %q, want %q", got, "RootGame")
	}
}

func TestStartNewNamedGame(t *testing.T) {
	name := "Test Realm"
	t.Cleanup(func() { _ = os.Remove(savePath(name)) })

	ge := NewGameEngine()
	if err := ge.StartNewNamedGame(name); err != nil {
		t.Fatalf("StartNewNamedGame(%q) failed: %v", name, err)
	}

	if got := ge.ActiveSaveName(); got != name {
		t.Errorf("ActiveSaveName() = %q, want %q", got, name)
	}
	if got := ge.ActiveParentName(); got != "" {
		t.Errorf("ActiveParentName() = %q, want \"\" (root)", got)
	}
	if !SaveExists(name) {
		t.Fatalf("save %q not found after StartNewNamedGame", name)
	}

	// The written file must load back into a fresh engine as that named root save.
	loaded := NewGameEngine()
	if err := loaded.LoadGame(name); err != nil {
		t.Fatalf("LoadGame(%q) failed: %v", name, err)
	}
	if got := loaded.ActiveSaveName(); got != name {
		t.Errorf("after load, ActiveSaveName() = %q, want %q", got, name)
	}
	if got := loaded.ActiveParentName(); got != "" {
		t.Errorf("after load, ActiveParentName() = %q, want \"\" (root)", got)
	}
}

func TestBranchSave(t *testing.T) {
	parent := "Branch Parent"
	child := "Branch Child"
	t.Cleanup(func() {
		_ = os.Remove(savePath(parent))
		_ = os.Remove(savePath(child))
	})

	ge := NewGameEngine()
	if err := ge.StartNewNamedGame(parent); err != nil {
		t.Fatalf("StartNewNamedGame(%q) failed: %v", parent, err)
	}

	// Capture the parent file's bytes so we can prove branching leaves it untouched.
	parentBytesBefore, err := os.ReadFile(savePath(parent))
	if err != nil {
		t.Fatalf("read parent save: %v", err)
	}

	if err := ge.BranchSave(child); err != nil {
		t.Fatalf("BranchSave(%q) failed: %v", child, err)
	}

	// The branch becomes the active slot with the parent recorded as its lineage.
	if got := ge.ActiveSaveName(); got != child {
		t.Errorf("ActiveSaveName() = %q, want %q", got, child)
	}
	if got := ge.ActiveParentName(); got != parent {
		t.Errorf("ActiveParentName() = %q, want %q", got, parent)
	}

	// The child file exists and loads back with its parent recorded.
	if !SaveExists(child) {
		t.Fatalf("child save %q not found after BranchSave", child)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame(child); err != nil {
		t.Fatalf("LoadGame(%q) failed: %v", child, err)
	}
	if got := loaded.ActiveParentName(); got != parent {
		t.Errorf("after load, child ActiveParentName() = %q, want %q", got, parent)
	}

	// The parent save still exists, byte-identical (frozen at the branch point).
	if !SaveExists(parent) {
		t.Fatalf("parent save %q vanished after BranchSave", parent)
	}
	parentBytesAfter, err := os.ReadFile(savePath(parent))
	if err != nil {
		t.Fatalf("re-read parent save: %v", err)
	}
	if string(parentBytesBefore) != string(parentBytesAfter) {
		t.Errorf("parent save bytes changed after BranchSave; want unchanged")
	}

	// Branching to the same name again is rejected (already exists).
	if err := ge.BranchSave(child); err == nil {
		t.Errorf("BranchSave(%q) twice = nil error, want already-exists error", child)
	}
}

func TestActiveParentNameSetter(t *testing.T) {
	// A fresh engine is a lineage root → empty parent.
	ge := NewGameEngine()
	if got := ge.ActiveParentName(); got != "" {
		t.Errorf("fresh ActiveParentName() = %q, want \"\"", got)
	}
	// Setter/getter round-trip.
	ge.SetActiveParentName("Ancestor Realm")
	if got := ge.ActiveParentName(); got != "Ancestor Realm" {
		t.Errorf("after SetActiveParentName, ActiveParentName() = %q, want %q", got, "Ancestor Realm")
	}
}

func TestLegacySaveHasNoParentName(t *testing.T) {
	// A legacy-style save written without parent_name must load as a root ("").
	name := "test_lineage_legacy"
	writeTestSave(t, name) // ensures the dir exists + registers cleanup
	writeRawSave(t, name, GameSave{
		Timestamp: time.Now(),
		Age:       "stone_age",
		// ParentName intentionally omitted (zero value, omitempty drops it).
	})

	loaded := NewGameEngine()
	if err := loaded.LoadGame(name); err != nil {
		t.Fatalf("LoadGame(%q) failed: %v", name, err)
	}
	if loaded.activeParentName != "" {
		t.Errorf("legacy save loaded with activeParentName = %q, want \"\"", loaded.activeParentName)
	}
	if got := findDetail(t, name).ParentName; got != "" {
		t.Errorf("legacy SaveInfo.ParentName = %q, want \"\"", got)
	}
}

// writeSignedSaveWithParent creates a real, signed, loadable save under name with
// the given ParentName, via the engine's SaveGame path (so verifySave passes), and
// registers cleanup. Same-package access lets us set the unexported parent field
// directly, matching TestParentNameRoundTrips.
func writeSignedSaveWithParent(t *testing.T, name, parent string) {
	t.Helper()
	ge := NewGameEngine()
	ge.activeParentName = parent
	if err := ge.SaveGame(name); err != nil {
		t.Fatalf("SaveGame(%q) failed: %v", name, err)
	}
	t.Cleanup(func() { _ = os.Remove(savePath(name)) })
}

// readSaveFromDisk reads and JSON-parses a save by name for assertions.
func readSaveFromDisk(t *testing.T, name string) GameSave {
	t.Helper()
	data, err := os.ReadFile(savePath(name))
	if err != nil {
		t.Fatalf("read save %q: %v", name, err)
	}
	var gs GameSave
	if err := json.Unmarshal(data, &gs); err != nil {
		t.Fatalf("unmarshal save %q: %v", name, err)
	}
	return gs
}

func TestRenameSaveReparentsChildren(t *testing.T) {
	writeSignedSaveWithParent(t, "Parent", "")          // root
	writeSignedSaveWithParent(t, "Child", "Parent")     // child of Parent
	writeSignedSaveWithParent(t, "Grandchild", "Child") // child of Child
	t.Cleanup(func() { _ = os.Remove(savePath("NewParent")) })

	if err := RenameSave("Parent", "NewParent"); err != nil {
		t.Fatalf("RenameSave(Parent, NewParent) failed: %v", err)
	}

	// File move happened.
	if SaveExists("Parent") {
		t.Errorf("old save \"Parent\" still exists after rename")
	}
	if !SaveExists("NewParent") {
		t.Fatalf("renamed save \"NewParent\" missing")
	}

	// Child re-parented to the new name...
	child := readSaveFromDisk(t, "Child")
	if child.ParentName != "NewParent" {
		t.Errorf("Child.ParentName = %q, want \"NewParent\"", child.ParentName)
	}
	// ...and still verifies (re-sign worked → NOT flagged modified).
	if sigValid, _ := verifySave(&child); !sigValid {
		t.Errorf("Child failed signature verification after reparent — re-sign broken")
	}

	// Grandchild's parent ("Child") is untouched — only direct children move.
	grand := readSaveFromDisk(t, "Grandchild")
	if grand.ParentName != "Child" {
		t.Errorf("Grandchild.ParentName = %q, want \"Child\" (unaffected)", grand.ParentName)
	}
}

func TestRenameSaveDoesNotLaunderModifiedChild(t *testing.T) {
	writeSignedSaveWithParent(t, "Parent", "")      // root
	writeSignedSaveWithParent(t, "Child", "Parent") // child of Parent
	t.Cleanup(func() { _ = os.Remove(savePath("NewParent")) })

	// Tamper the child: bump a signed field but keep the stored signature, so the
	// file still parses as JSON yet fails verifySave (modified).
	child := readSaveFromDisk(t, "Child")
	child.Tick += 1000 // mutate a payload field WITHOUT re-signing
	tampered, err := json.Marshal(child)
	if err != nil {
		t.Fatalf("marshal tampered child: %v", err)
	}
	if err := os.WriteFile(savePath("Child"), tampered, 0644); err != nil {
		t.Fatalf("write tampered child: %v", err)
	}
	// Sanity: the tampered child must already read as modified.
	if sigValid, _ := verifySave(&child); sigValid {
		t.Fatalf("test setup failed — tampered child still verifies")
	}

	if err := RenameSave("Parent", "NewParent"); err != nil {
		t.Fatalf("RenameSave(Parent, NewParent) failed: %v", err)
	}

	after := readSaveFromDisk(t, "Child")
	// ParentName left untouched — we don't rewrite a modified save.
	if after.ParentName != "Parent" {
		t.Errorf("modified Child.ParentName = %q, want unchanged \"Parent\"", after.ParentName)
	}
	// Still detected as modified — no laundering of the badge.
	if sigValid, _ := verifySave(&after); sigValid {
		t.Errorf("modified Child now verifies after rename — badge was laundered")
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
	// New rich fields must all default to zero on a corrupt save (no panic, no
	// partial parse leaking through).
	if corrupt.Title != "" || corrupt.Epoch != "" || corrupt.PendingCatastrophe != "" ||
		corrupt.Population != 0 || corrupt.Buildings != 0 || corrupt.Wonders != 0 ||
		corrupt.Techs != 0 || corrupt.Soldiers != 0 || corrupt.PrestigeTotal != 0 ||
		corrupt.MilestonesDone != 0 || corrupt.MilestonesTotal != 0 {
		t.Errorf("corrupt save has non-zero rich fields: %+v", *corrupt)
	}
	// The healthy neighbour must still be listed and parsed fine — one bad file
	// must not poison the rest.
	if healthy == nil {
		t.Errorf("healthy save %q dropped from listing alongside a corrupt file", good)
	} else if healthy.Corrupt {
		t.Errorf("healthy save %q wrongly flagged Corrupt", good)
	}
}
