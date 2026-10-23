package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/flavor"
)

// flavorSheets are the seven sheets of the stories, with the fewest rows each
// has today. The flavor package tells every row of them through the script, so
// a sentence or a word added to a catalog counts here without a line of code.
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
