package ui

import (
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

// at returns a SaveInfo with a Timestamp offset minutes into the past, so
// "most-recent-first" ordering is deterministic: a larger min == older.
func at(name, parent string, min int) game.SaveInfo {
	return game.SaveInfo{
		Name:       name,
		ParentName: parent,
		Timestamp:  time.Now().Add(-time.Duration(min) * time.Minute),
	}
}

// rowByName returns the treeRow whose Info.Name matches, and a found flag.
func rowByName(rows []treeRow, name string) (treeRow, bool) {
	for _, r := range rows {
		if r.Info.Name == name {
			return r, true
		}
	}
	return treeRow{}, false
}

// names extracts the render-order name slice for order assertions.
func names(rows []treeRow) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Info.Name
	}
	return out
}

func TestBuildSaveTreeForestAndConnectors(t *testing.T) {
	// Forest:
	//   root_a (newest)
	//     ├─ child_a1 (newer child)
	//     │  └─ grandchild      ← 3rd level
	//     └─ child_a2 (older child, last sibling)
	//   root_b (older independent root)
	//   orphan      (parent "ghost" not present → promoted to root)
	//   selfloop    (parent == own name → root, no loop)
	saves := []game.SaveInfo{
		at("root_a", "", 1),
		at("child_a1", "root_a", 2),
		at("grandchild", "child_a1", 3),
		at("child_a2", "root_a", 4),
		at("root_b", "", 10),
		at("orphan", "ghost", 5),
		at("selfloop", "selfloop", 6),
	}

	rows := buildSaveTree(saves)

	if len(rows) != len(saves) {
		t.Fatalf("buildSaveTree returned %d rows, want %d", len(rows), len(saves))
	}

	// DFS pre-order, roots most-recent-first. root_a (1m) is newest root; its
	// children newest-first (child_a1 2m before child_a2 4m), grandchild nested
	// under child_a1. Then root_b, orphan, selfloop by recency (root_b 10m is
	// oldest so it sorts after orphan 5m and selfloop 6m among the roots).
	wantOrder := []string{
		"root_a",
		"child_a1",
		"grandchild",
		"child_a2",
		"orphan",   // 5m
		"selfloop", // 6m
		"root_b",   // 10m
	}
	got := names(rows)
	if len(got) != len(wantOrder) {
		t.Fatalf("render order length %d, want %d (%v)", len(got), len(wantOrder), got)
	}
	for i := range wantOrder {
		if got[i] != wantOrder[i] {
			t.Errorf("render order[%d] = %q, want %q\nfull: %v", i, got[i], wantOrder[i], got)
		}
	}

	// Roots carry no connector prefix.
	for _, name := range []string{"root_a", "root_b", "orphan", "selfloop"} {
		r, _ := rowByName(rows, name)
		if r.Prefix != "" {
			t.Errorf("root %q prefix = %q, want \"\" (no connector)", name, r.Prefix)
		}
	}

	// child_a1 is a non-last sibling → "├─ "; child_a2 is the last → "└─ ".
	if r, _ := rowByName(rows, "child_a1"); r.Prefix != "├─ " {
		t.Errorf("child_a1 prefix = %q, want %q", r.Prefix, "├─ ")
	}
	if r, _ := rowByName(rows, "child_a2"); r.Prefix != "└─ " {
		t.Errorf("child_a2 prefix = %q, want %q", r.Prefix, "└─ ")
	}

	// grandchild sits under child_a1, which still has a later sibling (child_a2),
	// so the pillar persists: "│  " + its own last-child segment "└─ ".
	if r, _ := rowByName(rows, "grandchild"); r.Prefix != "│  └─ " {
		t.Errorf("grandchild prefix = %q, want %q", r.Prefix, "│  └─ ")
	}
}

func TestBuildSaveTreeOrphanAndSelfParentAreRoots(t *testing.T) {
	saves := []game.SaveInfo{
		at("orphan", "does_not_exist", 1),
		at("me", "me", 2),
	}
	rows := buildSaveTree(saves)
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d (%v)", len(rows), names(rows))
	}
	for _, name := range []string{"orphan", "me"} {
		r, ok := rowByName(rows, name)
		if !ok {
			t.Fatalf("%q missing from tree", name)
		}
		if r.Prefix != "" {
			t.Errorf("%q should be a top-level root (empty prefix), got %q", name, r.Prefix)
		}
	}
}

func TestBuildSaveTreeCycleTerminates(t *testing.T) {
	// A.parent = B, B.parent = A — a 2-node cycle. Neither has an external root,
	// so both are mutual children. buildSaveTree must terminate and emit each
	// exactly once (no duplicates, no infinite recursion).
	saves := []game.SaveInfo{
		at("A", "B", 1),
		at("B", "A", 2),
	}

	done := make(chan []treeRow, 1)
	go func() { done <- buildSaveTree(saves) }()

	var rows []treeRow
	select {
	case rows = <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("buildSaveTree did not terminate on a 2-node cycle")
	}

	if len(rows) != 2 {
		t.Fatalf("cycle: want 2 rows total, got %d (%v)", len(rows), names(rows))
	}
	seen := map[string]int{}
	for _, r := range rows {
		seen[r.Info.Name]++
	}
	for _, name := range []string{"A", "B"} {
		if seen[name] != 1 {
			t.Errorf("cycle: %q appears %d times, want exactly 1", name, seen[name])
		}
	}
}
