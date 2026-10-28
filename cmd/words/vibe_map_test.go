package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEveryTextOfTheMapSheetIsWiredOrLeftAlone: a text of the map (its labels,
// legends, inspect lines, place and part names, the themes' words) that can
// follow VibeMode does: it is a vibe call (Line) or an entry of a table (E)
// that carries its row id. A new label that is not fails this test, and the
// fix is one command:
//
//	go run ./cmd/words vibe-wire map.csv
//
// What is left alone has a reason that is final (a key hint, a name no screen
// draws); none is left for later ("made once", "compared").
func TestEveryTextOfTheMapSheetIsWiredOrLeftAlone(t *testing.T) {
	c := realCatalog(t)
	var loose, later, needless []string
	wired := 0
	for _, r := range c.rows {
		if r.area != areaMap {
			continue
		}
		where := fmt.Sprintf("%s (%s:%d)", r.id, r.u.f.rel, r.u.line)
		switch v := c.wireVerdict(r); {
		case r.u.vibeID != "":
			wired++
			if forGood(v.reason) {
				needless = append(needless, where+": "+v.reason)
			}
		case v.kind != wireNone:
			loose = append(loose, where)
		case !forGood(v.reason):
			later = append(later, where+": "+v.reason)
		}
	}
	if wired < 150 {
		t.Errorf("only %d texts of the map sheet are wired", wired)
	}
	if len(loose) > 0 {
		t.Errorf("%d texts of the map sheet can follow VibeMode and are not written as vibe calls; run `go run ./cmd/words vibe-wire map.csv`:\n  %s",
			len(loose), strings.Join(loose, "\n  "))
	}
	if len(later) > 0 {
		t.Errorf("%d texts of the map sheet cannot follow VibeMode as the code is written; change the code or judge them in vibe_left.go or vibe_map.go:\n  %s",
			len(later), strings.Join(later, "\n  "))
	}
	if len(needless) > 0 {
		t.Errorf("%d texts of the map sheet are wired and judged not to follow VibeMode; take the vibe call off them:\n  %s",
			len(needless), strings.Join(needless, "\n  "))
	}
}

// TestTheMapsTablesCarryTheirIds: the tables the map makes once (the lines'
// names, the town's syllables, the themes) hold their words as
// entries that carry the row id, so the map reads them when it draws; a table
// entry is the text of its row and is counted as wired.
func TestTheMapsTablesCarryTheirIds(t *testing.T) {
	c := realCatalog(t)
	w := newVibeWiring(c)
	tables := map[string]int{
		"map.lineage.":          17,
		"theme.":                40,
		"mapmodel.world.syllA.": 20,
	}
	for prefix, least := range tables {
		n := 0
		for _, r := range c.rows {
			// (the single words of a table, "water", are on the unsure sheet, and
			// stand in the table as vibe.Same until that sheet is wired)
			if !strings.HasPrefix(r.id, prefix) || r.area != areaMap {
				continue
			}
			n++
			if !w.wired(r.id) {
				t.Errorf("%s is not counted as wired", r.id)
			}
		}
		if n < least {
			t.Errorf("%d rows start with %s, want at least %d", n, prefix, least)
		}
	}
}

// TestMapLabelsCarryTheirWidths: a label the map lays out in a width of its own
// has that width as its max in the export, so vibe-lint holds the script's
// line to it.
func TestMapLabelsCarryTheirWidths(t *testing.T) {
	c := realCatalog(t)
	for id, want := range map[string]int{
		"map.lineage.housing":                       20,
		"ui.mapstyle.roguelike.chrome.flowLines.01": 5, // the flows panel's tag column
		"ui.mapstyle.roguelike.chrome.flowLines.08": 5,
		"ui.mapstyle.roguelike.chrome.flowLines.09": 0, // the text next to it has the rest of the line
		"ui.mapstyle.roguelike.chrome.header.01":    0, // a header segment: dropped from the left until the line fits
	} {
		r := c.byID[id]
		if r == nil {
			t.Errorf("%s: no such row", id)
			continue
		}
		if r.max != want {
			t.Errorf("%s (%q): max %d, want %d", id, r.current, r.max, want)
		}
	}
	// Every plain text fits its own limit: a limit below the plain English
	// would refuse the line it replaces.
	for _, r := range c.rows {
		if r.area == areaMap && r.max > 0 && len([]rune(r.current)) > r.max {
			t.Errorf("%s: %q is %d characters and the row allows %d", r.id, r.current, len([]rune(r.current)), r.max)
		}
	}
}

// TestVibeLintHoldsTheCommandsASentenceNames: the hint line of the map that ends
// in the command map flows keeps it as it is typed, however the rest is said.
func TestVibeLintHoldsTheCommandsASentenceNames(t *testing.T) {
	c := realCatalog(t)
	for id, cmds := range typedIn {
		r := c.byID[id]
		if r == nil {
			t.Fatalf("%s: no such row", id)
		}
		for _, cmd := range cmds {
			if !strings.Contains(r.current, cmd) {
				t.Errorf("%s: %q is not in %q", id, cmd, r.current)
			}
		}
	}
	const id = "ui.mapstyle.roguelike.chrome.bottom.01"
	dir := t.TempDir()
	write := func(text string) {
		body := "id,text\n" + id + ",\"" + text + "\"\n"
		if err := os.WriteFile(filepath.Join(dir, "1.csv"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	write("arrows go  Shift quick  PgUp PgDn zoom  Tab next  Home middle  flows")
	if n, err := vibeLint(c, realRoot(t), dir, &out); err != nil || n != 1 || !strings.Contains(out.String(), "is missing map flows") {
		t.Errorf("vibe-lint on a line that lost the command: %d errors, %v:\n%s", n, err, out.String())
	}
	out.Reset()
	write("arrows go  Shift quick  PgUp PgDn zoom  Tab next  Home middle  map flows")
	if n, err := vibeLint(c, realRoot(t), dir, &out); err != nil || n != 0 {
		t.Errorf("vibe-lint refused a line that keeps the command: %d errors, %v:\n%s", n, err, out.String())
	}
}
