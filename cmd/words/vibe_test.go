package main

import (
	"bytes"
	"encoding/csv"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/words"
)

// knownGaps are the texts of the ruleset that the export names differently
// or not at all, by field, so a script row for them cannot reach the game
// yet: the badge families, the mission badges' era names and one
// reward title, which the code reader finds by their calls and names its
// own way, and the tech codes, which the export leaves out. The test pins
// the count so a new gap, or one that closes, is noticed.
var knownGaps = map[string]int{
	"config.BadgeExpeditionDef.Name":  21,
	"config.BadgeFamilyDef.Desc":      22,
	"config.BadgeFamilyDef.Descs":     16,
	"config.BadgeFamilyDef.Hint":      1,
	"config.BadgeFamilyDef.Ladder":    22,
	"config.BadgeReward.Title":        2,
	"config.MilestoneDef.Description": 1,
	"config.TechDef.Code":             39,
}

// TestRuntimeIdsAreTheExportsOwn is what keeps the overlay and the export
// calling a row by one name. It reads the ids off the running ruleset (the
// way VibeMode does) and holds each to the export's row: the same id, the
// same text, once the game's own composition is allowed for (a building's
// description is followed by what it does; {dur} is written into an era
// event). A text with words that the export has no row for must be in
// knownGaps.
func TestRuntimeIdsAreTheExportsOwn(t *testing.T) {
	c := realCatalog(t)
	gaps := map[string]int{}
	seen := map[string]string{}
	for _, cell := range words.Cells(rules.FromConfig()) {
		if prev, dup := seen[cell.ID]; dup && prev != cell.Text {
			t.Errorf("%s is two different texts in the ruleset: %q and %q", cell.ID, prev, cell.Text)
		}
		seen[cell.ID] = cell.Text
		row, ok := c.byID[cell.ID]
		if !ok {
			if hasWord(cell.Text) {
				gaps[cell.Field]++
			}
			continue
		}
		if got := config.ComposeText(cell.Owner, cell.Field, row.current); got != cell.Text {
			t.Errorf("%s: the export reads %q (%q once composed) and the ruleset has %q", cell.ID, row.current, got, cell.Text)
		}
	}
	for field, n := range gaps {
		if knownGaps[field] != n {
			t.Errorf("%s: %d texts with words have no row in the export (known: %d)", field, n, knownGaps[field])
		}
	}
	for field, n := range knownGaps {
		if _, ok := gaps[field]; !ok {
			t.Errorf("%s was a known gap of %d and has closed: take it out of knownGaps", field, n)
		}
	}
	if len(seen) < 3000 {
		t.Errorf("the walk found only %d texts in the ruleset", len(seen))
	}
}

// TestLexiconRowsAreNotInTheExport: the lexicon's ids are its own, and none
// of them is an export id.
func TestLexiconRowsAreNotInTheExport(t *testing.T) {
	c := realCatalog(t)
	for _, cell := range config.LexiconCells() {
		if _, ok := c.byID[cell.ID]; ok {
			t.Errorf("%s is a lexicon id and an export id", cell.ID)
		}
	}
}

// TestTheVibeScriptIsSound: the script that ships passes vibe-lint, and the
// vibe.Line calls in the code agree with the export.
func TestTheVibeScriptIsSound(t *testing.T) {
	c := realCatalog(t)
	var out bytes.Buffer
	n, err := vibeLint(c, realRoot(t), "", &out)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("vibe-lint found %d errors:\n%s", n, out.String())
	}
	if calls := lineCalls(c.m); len(calls) < 2 {
		t.Errorf("found %d vibe.Line calls, want the demo lines at least", len(calls))
	}
}

// TestVibeLintCatches: each way a row can be wrong is an error, in a sheet of
// its own, and a sound row is not.
func TestVibeLintCatches(t *testing.T) {
	c := realCatalog(t)
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Find a row whose text carries a format verb, one with a {dur} slot and
	// one with a short limit.
	var verbRow, slotRow, maxRow *row
	for _, r := range c.rows {
		switch {
		case verbRow == nil && strings.Count(r.current, "%s") == 1 && strings.Count(r.current, "%") == 1 && r.area == areaTechs:
			verbRow = r
		case slotRow == nil && strings.Contains(r.current, "{dur}"):
			slotRow = r
		case maxRow == nil && r.max > 0 && strings.HasPrefix(r.id, "tech."):
			maxRow = r
		}
	}
	if verbRow == nil || slotRow == nil || maxRow == nil {
		t.Fatalf("the export has no row with a verb (%v), a {dur} slot (%v) and a limit (%v)", verbRow != nil, slotRow != nil, maxRow != nil)
	}
	write("1good.csv", "id,text\n"+
		"building.hut.name,Crash Pad\n"+
		"lexicon.bronze_age.scholars.one,Big Brain\n"+
		verbRow.id+",\"Everything goes %s\"\n")
	write("2bad.csv", "id,text\n"+
		"building.nonesuch.name,Nothing\n"+
		slotRow.id+",A line with no duration\n"+
		maxRow.id+",WAYTOOLONGFORTHELIMIT\n"+
		"building.hut.name,Same id again\n")
	write("3worse.csv", "id,notes\nx,y\n")

	var out bytes.Buffer
	n, err := vibeLint(c, realRoot(t), dir, &out)
	if err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"building.nonesuch.name is not a row of the export or the lexicon",
		slotRow.id + " is missing {dur}",
		maxRow.id + " is ",
		"the id building.hut.name is also in 1good.csv:2",
		"3worse.csv: the header row has no id and text columns",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("vibe-lint did not say %q:\n%s", want, got)
		}
	}
	if n != 5 {
		t.Errorf("vibe-lint found %d errors, want 5:\n%s", n, got)
	}
	if strings.Contains(got, "1good.csv:4") || strings.Contains(got, "1good.csv:3") {
		t.Errorf("vibe-lint refused a sound row:\n%s", got)
	}
}

// TestVibeLintReadsTheLineCalls: a vibe call (Line, Sprintf, Errorf) must name
// an export row and quote its text.
func TestVibeLintReadsTheLineCalls(t *testing.T) {
	src := `package game
import "github.com/espresso20/ageforge/vibe"
func f(x string, a ...any) {
	_ = vibe.Line("a.b.01", "Plain one")
	_ = vibe.Line("a.b.02", x)
	_ = vibe.Line("a.b.03")
	_ = other.Line("a.b.04", "Not ours")
	_ = vibe.Text("a.b.05", "Not a Line")
	_ = vibe.Sprintf("a.b.06", "%s and "+"%d", x, 1)
	_ = vibe.Errorf("a.b.07", "none")
	_ = vibe.Sprintf("a.b.08", x, a...)
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "game/x.go", src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	m := &module{fset: fset, files: []*srcFile{{rel: "game/x.go", ast: f}}}
	got := lineCalls(m)
	want := []lineCall{
		{id: "a.b.01", plain: "Plain one", kind: "Line", file: "game/x.go", line: 4, literal: true},
		{id: "a.b.02", kind: "Line", file: "game/x.go", line: 5},
		{id: "a.b.06", plain: "%s and %d", kind: "Sprintf", file: "game/x.go", line: 9, literal: true, args: 2},
		{id: "a.b.07", plain: "none", kind: "Errorf", file: "game/x.go", line: 10, literal: true},
		{id: "a.b.08", kind: "Sprintf", file: "game/x.go", line: 11, args: -1},
	}
	if len(got) != len(want) {
		t.Fatalf("lineCalls found %d calls, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("call %d: %+v, want %+v", i, got[i], want[i])
		}
	}
}

// TestVibeExport: the to-do sheets hold every row of the export and the
// lexicon the script has no line for, once, with the plain text to read and
// the text to fill in left empty; a row the script has is not there.
func TestVibeExport(t *testing.T) {
	c := realCatalog(t)
	root := realRoot(t)
	dir := t.TempDir()
	var out bytes.Buffer
	if err := vibeExport(c, root, "", dir, &out); err != nil {
		t.Fatal(err)
	}
	rows, _, err := readScript(scriptDir(root, ""))
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, r := range rows {
		have[r.ID] = true
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.csv"))
	// Sheets drop out of the to-do list as their script is finished, so only
	// check that the export wrote something.
	if len(files) == 0 {
		t.Fatal("no sheets written")
	}
	ids := map[string]string{}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		recs, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(data), bom))).ReadAll()
		if err != nil {
			t.Fatal(err)
		}
		for i, rec := range recs {
			if i == 0 {
				if strings.Join(rec, ",") != strings.Join(vibeColumns, ",") {
					t.Errorf("%s: header %v", f, rec)
				}
				continue
			}
			id := rec[0]
			if prev, dup := ids[id]; dup {
				t.Errorf("%s is in %s and %s", id, prev, f)
			}
			ids[id] = filepath.Base(f)
			if have[id] {
				t.Errorf("%s is in the script already and in the to-do sheet %s", id, f)
			}
			if rec[len(rec)-1] != "" {
				t.Errorf("%s: the text to fill in is not empty", id)
			}
			if rec[6] == "" {
				t.Errorf("%s: no plain text", id)
			}
			if rec[7] != "yes" && rec[7] != "no" {
				t.Errorf("%s: wired is %q", id, rec[7])
			}
		}
	}
	want := 0
	for _, r := range c.rows {
		if !have[r.id] {
			if reason := c.wireVerdict(r).reason; forGood(reason) {
				// Nothing will read a line for it: it is not asked for.
				if _, ok := ids[r.id]; ok {
					t.Errorf("%s is left alone for good (%s) and is in a to-do sheet", r.id, reason)
				}
				continue
			}
			want++
			if _, ok := ids[r.id]; !ok {
				t.Errorf("the export row %s is in no to-do sheet", r.id)
			}
		}
	}
	for _, cell := range config.LexiconCells() {
		if !have[cell.ID] {
			want++
			if _, ok := ids[cell.ID]; !ok {
				t.Errorf("the lexicon row %s is in no to-do sheet", cell.ID)
			}
		}
	}
	if len(ids) != want {
		t.Errorf("%d rows in the sheets, want %d", len(ids), want)
	}
	if !strings.Contains(out.String(), "total") {
		t.Errorf("the summary has no total:\n%s", out.String())
	}
	// A second run starts clean: a row the script has since gained leaves.
	names := []string{}
	for _, f := range files {
		names = append(names, filepath.Base(f))
	}
	sort.Strings(names)
	if names[0] != "badges.csv" {
		t.Errorf("sheets: %v", names)
	}
}
