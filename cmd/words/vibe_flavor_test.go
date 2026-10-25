package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/flavor"
)

// flavorSheets are the seven sheets of the stories, with the fewest rows each
// may have. The flavor package tells every row of them through the script, so
// a sentence or a word added to a catalog counts here without a line of code.
// A catalog that loses rows fails the test; one that gains rows still passes.
var flavorSheets = []struct {
	area *area
	min  int
}{
	{areaFlavorEnd, 110},
	{areaFlavorRaid, 390},
	{areaFlavorHarbIn, 710},
	{areaFlavorHarbDo, 427},
	{areaFlavorHarbOut, 399},
	{areaFlavorEnc, 879},
	{areaFlavorExp, 1038},
}

// TestEveryTextOfTheFlavorSheetsIsToldThroughTheScript: a row of a story sheet
// (a sentence of a catalog, an entry of a word list, a piece of the frames that
// name a resource) is read through the script when its line is drawn, so it
// counts as wired and none is left alone. A catalog row the flavor package
// does not tell, or a piece of code text in these sheets that no vibe call
// carries, fails here.
func TestEveryTextOfTheFlavorSheetsIsToldThroughTheScript(t *testing.T) {
	c := realCatalog(t)
	w := newVibeWiring(c)
	wired := map[*area]int{}
	for _, r := range c.rows {
		sheet := false
		for _, s := range flavorSheets {
			sheet = sheet || r.area == s.area
		}
		if !sheet {
			continue
		}
		where := fmt.Sprintf("%s (%s:%d)", r.id, r.u.f.rel, r.u.line)
		if !w.wired(r.id) {
			t.Errorf("%s is not read through the script", where)
			continue
		}
		if reason := c.wireVerdict(r).reason; reason != "" && r.u.vibeID == "" {
			t.Errorf("%s is left alone (%s)", where, reason)
		}
		wired[r.area]++
	}
	for _, s := range flavorSheets {
		if wired[s.area] < s.min {
			t.Errorf("only %d rows of %s are wired, want at least %d", wired[s.area], s.area.file, s.min)
		}
	}
}

// TestTheFlavorPackageTellsTheExportsOwnRows: the ids the flavor package
// reads a line by are the ids the export gives the same texts, one for one, and
// the text it quotes for each is the export's.
func TestTheFlavorPackageTellsTheExportsOwnRows(t *testing.T) {
	c := realCatalog(t)
	told := map[string]string{}
	for _, fr := range flavor.VibeRows() {
		if _, dup := told[fr.ID]; dup {
			t.Errorf("the flavor package tells %s twice", fr.ID)
		}
		told[fr.ID] = fr.Plain
		r, ok := c.byID[fr.ID]
		switch {
		case !ok:
			t.Errorf("the flavor package tells %s, which is not a row of the export", fr.ID)
		case r.rules&(rulesSkel|rulesBank) == 0:
			t.Errorf("the flavor package tells %s, which the export does not read as a sentence or a word list", fr.ID)
		case r.current != fr.Plain:
			t.Errorf("%s: the flavor package quotes %q and the export reads %q", fr.ID, fr.Plain, r.current)
		}
	}
	n := 0
	for _, r := range c.rows {
		if r.rules&(rulesSkel|rulesBank) == 0 {
			continue
		}
		n++
		if _, ok := told[r.id]; !ok {
			t.Errorf("%s is a row of the export that the flavor package does not tell", r.id)
		}
	}
	if n != len(told) || n < 3900 {
		t.Errorf("the export has %d sentences and word-list entries, the flavor package tells %d (want 3,900 or more)", n, len(told))
	}
}

// TestVibeWireHasNothingToDoForTheFlavorSheets: the story sheets are wired by
// the catalogs themselves, so vibe-wire finds all of their rows wired already
// and rewrites no source.
func TestVibeWireHasNothingToDoForTheFlavorSheets(t *testing.T) {
	c := realCatalog(t)
	for _, s := range flavorSheets {
		var out bytes.Buffer
		if err := vibeWire(c, realRoot(t), s.area.file, "", true, &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), ", 0 to wire now (") || strings.Contains(out.String(), "left:") {
			t.Errorf("vibe-wire has something to do for %s:\n%s", s.area.file, out.String())
		}
	}
}

// TestVibeLintHoldsAStoryLineToItsSlots: a line of the script for a story
// sentence keeps the slot the sentence draws a word into and the placeholders
// the game fills in, and counts as read by the game (no "nothing reads it yet"
// note). A line that drops either is an error.
func TestVibeLintHoldsAStoryLineToItsSlots(t *testing.T) {
	c := realCatalog(t)
	var both, slotted *row // a sentence with a placeholder, one with a slot
	for _, r := range c.rows {
		if r.rules&rulesSkel == 0 || r.max < len(r.current)+20 {
			continue
		}
		switch {
		case strings.Contains(r.current, "{subject}") && (both == nil || len(r.current) < len(both.current)):
			both = r
		case strings.Contains(r.current, "~") && !strings.Contains(r.current, "{") && (slotted == nil || len(r.current) < len(slotted.current)):
			slotted = r
		}
	}
	var word *row
	for _, r := range c.rows {
		if r.rules&rulesBank != 0 && (word == nil || len(r.current) < len(word.current)) {
			word = r
		}
	}
	if both == nil || slotted == nil || word == nil {
		t.Fatalf("the export has no sentence with a subject (%v), none with a slot (%v), and no word-list entry (%v)", both != nil, slotted != nil, word != nil)
	}
	lint := func(sheet string) (int, string) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "story.csv"), []byte("id,text\n"+sheet), 0o644); err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		n, err := vibeLint(c, realRoot(t), dir, &out)
		if err != nil {
			t.Fatal(err)
		}
		return n, out.String()
	}
	row := func(id, text string) string { return id + "," + strconv.Quote(text) + "\n" }

	n, got := lint(row(both.id, "Fr fr: "+both.current) + row(slotted.id, "Fr fr: "+slotted.current) + row(word.id, word.current))
	if n != 0 || strings.Contains(got, "nothing reads them yet") {
		t.Errorf("a sound story line was refused or called unread (%d errors):\n%s", n, got)
	}
	n, got = lint(row(slotted.id, "No slot here, fr") + row(both.id, strings.ReplaceAll(both.current, "{subject}", "them")))
	for _, want := range []string{slotted.id, both.id, "{subject}"} {
		if !strings.Contains(got, want) {
			t.Errorf("vibe-lint did not say %q:\n%s", want, got)
		}
	}
	if n != 2 {
		t.Errorf("vibe-lint found %d errors in two story lines that drop a slot, want 2:\n%s", n, got)
	}
}
