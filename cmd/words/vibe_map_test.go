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
// What is left alone has a reason that is final (a key hint, a drawing, a name
// no screen draws); none is left for later ("made once", "compared").
func TestEveryTextOfTheMapSheetIsWiredOrLeftAlone(t *testing.T) {
	c := realCatalog(t)
	var loose, later, needless []string
	wired := 0
	left := map[string]int{}
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
		default:
			left[v.reason]++
		}
	}
	if wired < 1000 {
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
	for _, reason := range []string{leftKeycap, leftNeverShown, leftArt} {
		if left[reason] == 0 {
			t.Errorf("no text of the map sheet is left alone as %q: the rule or the list is dead", reason)
		}
	}
}

// TestTheMapsTablesCarryTheirIds: the tables the map makes once (the legend's
// entries, the city's parts, the movers, the themes) hold their words as
// entries that carry the row id, so the map reads them when it draws; a table
// entry is the text of its row and is counted as wired.
func TestTheMapsTablesCarryTheirIds(t *testing.T) {
	c := realCatalog(t)
	w := newVibeWiring(c)
	tables := map[string]int{
		"map.city.":                           120, // the parts of the late-game city: name, title, lines
		"map.mover.":                          120,
		"map.sky.":                            60,
		"map.lineage.":                        17,
		"theme.":                              40,
		"ui.mapstyle.roguelike.glyph.lgInfo.": 10,
		"mapmodel.world.syllA.":               20,
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
		"ui.mapstyle.roguelike.glyph.lgInfo.18":           20, // a legend label: "new since last visit" fills the 20 cells
		"ui.mapstyle.roguelike.space_legend.skyLgInfo.30": 20,
		"ui.mapstyle.roguelike.glyph.lgGroups.03":         23,
		"map.city.cooling.name":                           20,
		"map.city.cooling.title":                          0, // the heading of an inspect line, as long as the line
		"map.city.cooling.line.01":                        0,
		"map.mover.starship.name":                         20,
		"map.mover.starship.title":                        0,
		"map.sky.SkyQuantum.trade":                        20,
		"map.lineage.housing":                             20,
		"ui.mapstyle.roguelike.chrome.flowLines.01":       5, // the flows panel's tag column
		"ui.mapstyle.roguelike.chrome.flowLines.08":       5,
		"ui.mapstyle.roguelike.chrome.flowLines.09":       0, // the text next to it has the rest of the line
		"ui.mapstyle.roguelike.space_chrome.flowLines.05": 5,
		"ui.mapstyle.roguelike.city.holoAds.01":           4, // one letter to a cell, four to a block
		"ui.mapstyle.roguelike.city.megacorps.02":         3, // a tower's initials
		"ui.mapstyle.roguelike.city.megacorps.01":         0, // its name is an inspect title
		"ui.mapstyle.skyline.city.holoAds.01":             8,
		"ui.mapstyle.skyline.wonders_late.citadel.01":     7,
		"ui.mapstyle.roguelike.chrome.header.01":          0, // a header segment: dropped from the left until the line fits
		"ui.mapstyle.roguelike.inspect.describe.01":       0,
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

// TestAStringWithNoRowBecomesAnEntryWhenItGetsOne: a table of entries holds the
// strings that have no row as vibe.Same, and vibe-wire turns one into vibe.E when
// a sheet gives it a row, so the sheet that does (the unsure one) needs no
// hand work.
func TestAStringWithNoRowBecomesAnEntryWhenItGetsOne(t *testing.T) {
	c := realCatalog(t)
	n := 0
	for _, r := range c.rows {
		if r.u.f.rel != "ui/mapstyle/roguelike/glyph.go" || r.current != "water" {
			continue
		}
		n++
		if v := c.wireVerdict(r); v.kind != wireSame {
			t.Errorf("%s (%q) is judged %v, want it turned from vibe.Same into vibe.E", r.id, r.current, v.kind)
		}
	}
	if n == 0 {
		t.Fatal("no row for the legend's word water")
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
