package main

import (
	"bytes"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// partialTree copies a few real source files, and the lint tests the copy
// rules are read from, into a temporary tree the import can write to.
func partialTree(t *testing.T) string {
	t.Helper()
	src, dst := realRoot(t), t.TempDir()
	for _, rel := range []string{
		"go.mod",
		"config/research.go",
		"ui/accounts_panel.go",
		"flavor/catalog.go", "flavor/skeleton.go", "flavor/harbinger.go", "flavor/catalog_exp_fail.go",
		"ui/copy_guard_test.go", "ui/stale_copy_test.go", "ui/docs_lint_test.go",
		"config/effect_text_test.go", "config/event_flavor_test.go",
		"flavor/flavor_test.go", "flavor/generator_test.go",
	} {
		data, err := os.ReadFile(filepath.Join(src, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		to := filepath.Join(dst, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(to, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dst
}

// fill types yours against ids in the sheets of dir.
func fill(t *testing.T, dir string, yours map[string]string) {
	t.Helper()
	names, err := sheetFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	left := map[string]bool{}
	for id := range yours {
		left[id] = true
	}
	for _, n := range names {
		if n == firstHoursFile {
			continue
		}
		recs, err := readSheet(filepath.Join(dir, n))
		if err != nil {
			t.Fatal(err)
		}
		for _, rec := range recs {
			if y, ok := yours[rec[colID]]; ok {
				rec[colYours] = y
				delete(left, rec[colID])
			}
		}
		writeSheet(t, dir, n, recs...)
	}
	if len(left) > 0 {
		t.Fatalf("no sheet has these ids: %v", left)
	}
}

func sources(t *testing.T, root string, rels ...string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	for _, rel := range rels {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		out[rel] = data
	}
	return out
}

// TestImportRoundTrip: on a copy of a few real files, a rewrite lands in the
// literal it was written against and in no other token of any file; the
// literal the code compares a button's label with moves with the label; a
// second import changes nothing; the ledger and status agree.
func TestImportRoundTrip(t *testing.T) {
	root := partialTree(t)
	dir := filepath.Join(root, "words", "export")
	c, err := read(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := export(c, dir, io.Discard); err != nil {
		t.Fatal(err)
	}
	files := []string{"config/research.go", "ui/accounts_panel.go", "flavor/catalog_exp_fail.go"}
	before := sources(t, root, files...)

	title := rowWith(t, c, "ui.accounts_panel.", "[gold]═══ Accounts ═══[-]")
	wipe := rowWith(t, c, "ui.accounts_panel.", "Wipe it")
	seven := rowWith(t, c, "flavor.expedition_fail.any.", "Seven back of nine")
	want := map[string][2]string{ // file -> literal before, literal after
		"tech.calendar.description": {`"Counting the days fixes the feasts and the seasons."`, `"Counting the days tells the farmers when to plant."`},
		title.id:                    {`"[gold]═══ Accounts ═══[-]"`, `"[gold]═══ Your accounts ═══[-]"`},
		wipe.id:                     {`"Wipe it"`, `"Erase it"`},
		seven.id:                    {`"Seven back of nine"`, `"Seven of the nine came back"`},
	}
	if len(wipe.twins) != 1 {
		t.Fatalf("%s should be tied to the one comparison that reads its label, has %d", wipe.id, len(wipe.twins))
	}
	yours := map[string]string{}
	for id, w := range want {
		yours[id] = strings.Trim(w[1], `"`)
	}
	fill(t, dir, yours)

	var out bytes.Buffer
	if err := importSheets(root, dir, importOptions{}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "applied 4, refused 0, unchanged 0") {
		t.Fatalf("the import said:\n%s", out.String())
	}

	after := sources(t, root, files...)
	moved := map[[2]string]int{}
	for _, rel := range files {
		if formatted, err := format.Source(after[rel]); err != nil || !bytes.Equal(formatted, after[rel]) {
			t.Errorf("%s is not gofmt-formatted after the import (%v)", rel, err)
		}
		a, b := tokens(t, before[rel]), tokens(t, after[rel])
		if len(a) != len(b) {
			t.Fatalf("%s has %d tokens after the import and had %d", rel, len(b), len(a))
		}
		for i := range a {
			if a[i] != b[i] {
				moved[[2]string{a[i], b[i]}]++
			}
		}
	}
	wantMoved := map[[2]string]int{}
	for id, w := range want {
		wantMoved[w]++
		if id == wipe.id {
			wantMoved[w]++ // and the comparison that reads the label
		}
	}
	if len(moved) != len(wantMoved) {
		t.Errorf("tokens that changed: %v, want %v", moved, wantMoved)
	}
	for w, n := range wantMoved {
		if moved[w] != n {
			t.Errorf("%s -> %s changed %d times, want %d", w[0], w[1], moved[w], n)
		}
	}

	// The source now reads as rewritten, and the ledger knows it.
	c2, err := read(root)
	if err != nil {
		t.Fatal(err)
	}
	led, err := loadLedger(root)
	if err != nil {
		t.Fatal(err)
	}
	for id, w := range want {
		text := strings.Trim(w[1], `"`)
		if r := c2.byID[id]; r == nil || r.current != text {
			t.Errorf("%s does not read %q after the import", id, text)
		}
		if led[id] != hashText(text) {
			t.Errorf("the ledger has %q for %s", led[id], id)
		}
	}
	if len(led) != len(want) {
		t.Errorf("the ledger holds %d rows, want %d", len(led), len(want))
	}
	_, total, gone := count(c2, led)
	if total.doneLines != len(want) || total.changed != 0 || gone != 0 {
		t.Errorf("status counts %d rewritten, %d changed since, %d gone; want %d, 0, 0", total.doneLines, total.changed, gone, len(want))
	}
	var st bytes.Buffer
	if err := status(root, &st); err != nil || !strings.Contains(st.String(), "total") {
		t.Errorf("status: %v\n%s", err, st.String())
	}

	// The same sheets again: nothing to do, nothing written.
	out.Reset()
	if err := importSheets(root, dir, importOptions{}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "applied 0, refused 0, unchanged 4") {
		t.Errorf("the second import said:\n%s", out.String())
	}
	for rel, data := range sources(t, root, files...) {
		if !bytes.Equal(data, after[rel]) {
			t.Errorf("%s changed on the second import", rel)
		}
	}

	// A rewritten line edited by hand afterwards is reported as changed since.
	path := filepath.Join(root, "config", "research.go")
	edited := bytes.Replace(after["config/research.go"], []byte("tells the farmers when to plant"), []byte("tells the farmers when to sow"), 1)
	if err := os.WriteFile(path, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	c3, err := read(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, total, _ := count(c3, led); total.changed != 1 || total.doneLines != len(want)-1 {
		t.Errorf("after a hand edit: %d changed since, %d rewritten; want 1 and %d", total.changed, total.doneLines, len(want)-1)
	}
}

// TestDryRunWritesNothing: a dry run reports and leaves the tree alone.
func TestDryRunWritesNothing(t *testing.T) {
	root := partialTree(t)
	dir := filepath.Join(root, "words", "export")
	c, err := read(root)
	if err != nil {
		t.Fatal(err)
	}
	r := c.byID["tech.calendar.description"]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeSheet(t, dir, areaTechs.file, c.record(r, "Days are counted and feasts are fixed.", ""))
	before := sources(t, root, "config/research.go")
	var out bytes.Buffer
	if err := importSheets(root, dir, importOptions{dryRun: true}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Would apply  tech.calendar.description") || !strings.Contains(out.String(), "would apply 1, refused 0") {
		t.Errorf("the dry run said:\n%s", out.String())
	}
	if after := sources(t, root, "config/research.go"); !bytes.Equal(after["config/research.go"], before["config/research.go"]) {
		t.Error("a dry run changed the source")
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(ledgerPath))); !os.IsNotExist(err) {
		t.Error("a dry run wrote the ledger")
	}
}

// TestImportRefusals: each reason the import refuses a row, against the
// real source. A refused row changes nothing, and the good row beside it
// still goes in.
func TestImportRefusals(t *testing.T) {
	c := realCatalog(t)
	built := rowWith(t, c, "game.engine.finishBuild.", "%s built (you have %d).")
	title := rowWith(t, c, "ui.accounts_panel.", "[gold]═══ Accounts ═══[-]")
	welcome := rowWith(t, c, "game.engine.", "Welcome to AgeForge. You have nothing but your hands.")
	seven := rowWith(t, c, "flavor.expedition_fail.any.", "Seven back of nine")
	slotted := rowWith(t, c, "flavor.expedition_fail.any.", "Whoever was carrying ~ has not been asked about it yet")
	rope := rowWith(t, c, "flavor.words.exp_fail_kit.", "the spare rope")
	var saveWord *row
	for _, r := range c.rows {
		if r.area == areaSaveNames {
			saveWord = r
			break
		}
	}
	desc, code, name := c.byID["tech.calendar.description"], c.byID["tech.calendar.code"], c.byID["tech.calendar.name"]
	quip := c.byID["milestone.first_shelter.flavor"]

	stale := c.record(desc, "Days are counted.", "")
	stale[colCurrent] = "Counting the days fixed the feasts."
	gone := c.record(desc, "Days are counted.", "")
	gone[colID] = "tech.no_such_tech.description"

	cases := []struct {
		name string
		rec  []string
		why  string
	}{
		{"id gone", gone, "no longer in the source"},
		{"source changed", stale, "has changed in the source since the export"},
		{"verb missing", c.record(built, "%s is built.", ""), "must keep the format verbs %s %d"},
		{"verbs out of order", c.record(built, "You have %d: %s is built.", ""), "format verbs out of order"},
		{"tag altered", c.record(title, "[yellow]═══ Accounts ═══[-]", ""), "is missing [gold]"},
		{"glyph missing", c.record(title, "[gold]Accounts[-]", ""), "is missing ═══"},
		{"tags unbalanced", c.record(title, "[-]═══ Accounts ═══[gold]", ""), "unbalanced style tags"},
		{"slot missing", c.record(slotted, "Whoever was carrying it has not been asked about it yet", ""), "is missing ~"},
		{"over max", c.record(code, "CALENDAR", ""), "the most is 5"},
		{"code taken", c.record(code, "LANG", ""), "is the card code Language already has"},
		{"em dash", c.record(welcome, "Welcome to AgeForge — you have nothing but your hands.", ""), "em dash"},
		{"exclamation", c.record(welcome, "Welcome to AgeForge! You have nothing but your hands.", ""), "exclamation mark"},
		{"retired term", c.record(welcome, "Welcome to AgeForge. Your villagers have nothing but their hands.", ""), "retired term"},
		{"config style", c.record(desc, "Counting the days fixes the colour of the seasons.", ""), "the style guard for game data bans"},
		{"config dash", c.record(name, "Calendar – Almanac", ""), "has a dash"},
		{"quip markup", c.record(quip, "A roof. [gold]Four walls.[-]", ""), "square bracket"},
		{"sentence shape", c.record(seven, "It was not the cold that took them, but the dark", ""), "sentence shape the catalog bans"},
		{"sentence length", c.record(seven, "Seven back", ""), "a catalog sentence is 3 to 45 words"},
		{"era word", c.record(seven, "The cart came home empty", ""), "which belongs to"},
		{"full stop", c.record(seven, "Seven came back of the nine.", ""), "the game adds the full stop"},
		{"restating", c.record(seven, "The expedition lost two of nine", ""), "repeats what the log line above it already said"},
		{"word-list entry", c.record(rope, "The spare rope", ""), "starts with a capital letter"},
		{"save name", c.record(saveWord, "O'Brien", ""), "letters and single spaces"},
	}
	dir := t.TempDir()
	recs := [][]string{c.record(desc, "Counting the days tells the farmers when to plant.", "")}
	recs[0][colID] = "tech.language.description"
	recs[0][colCurrent] = c.byID["tech.language.description"].current
	for i, tc := range cases {
		rec := append([]string(nil), tc.rec...)
		writeSheet(t, dir, "case.csv", recs[0], rec)
		res, err := decide(c, dir, false)
		if err != nil {
			t.Fatal(err)
		}
		if len(res.refused) != 1 || !strings.Contains(res.refused[0].why, tc.why) {
			t.Errorf("%s: want one refusal saying %q, got %+v", tc.name, tc.why, res.refused)
		}
		if len(res.changes) != 1 || res.changes[0].r.id != "tech.language.description" {
			t.Errorf("%s (case %d): the good row beside the refused one should still apply, got %d changes", tc.name, i, len(res.changes))
		}
	}

	// The same id filled two ways in two files: both are refused.
	writeSheet(t, dir, "case.csv", c.record(desc, "Days are counted.", ""))
	writeSheet(t, dir, firstHoursFile, c.record(desc, "The days are counted.", ""))
	res, err := decide(c, dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.changes) != 0 || len(res.refused) != 2 || !strings.Contains(res.refused[0].why, "filled differently in two files") {
		t.Errorf("an id filled two ways: got %d changes and %+v", len(res.changes), res.refused)
	}
	// Filled the same way in both: it counts once.
	writeSheet(t, dir, firstHoursFile, c.record(desc, "Days are counted.", ""))
	if res, _ = decide(c, dir, false); len(res.changes) != 1 || len(res.refused) != 0 {
		t.Errorf("an id filled the same way twice: got %d changes, %d refusals", len(res.changes), len(res.refused))
	}
}

// TestSameAs: a blank row takes the rewrite of the row its same_as names,
// unless it is switched off or the row has a rewrite of its own.
func TestSameAs(t *testing.T) {
	c := realCatalog(t)
	var leader *row
	var followers []*row
	for _, r := range c.rows {
		if r.current == "Cancel" && r.sameAs == "" {
			leader = r
		}
	}
	if leader == nil {
		t.Fatal("no row reads Cancel")
	}
	for _, r := range c.rows {
		if r.sameAs == leader.id {
			followers = append(followers, r)
		}
	}
	if len(followers) < 2 {
		t.Fatalf("only %d rows are the same as %s", len(followers), leader.id)
	}
	dir := t.TempDir()
	recs := [][]string{c.record(leader, "Back", "")}
	for _, f := range followers {
		recs = append(recs, c.record(f, "", ""))
	}
	writeSheet(t, dir, "case.csv", recs...)

	res, err := decide(c, dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.changes) != 1+len(followers) || len(res.refused) != 0 {
		t.Fatalf("with same_as: %d changes, %d refusals, want %d and 0", len(res.changes), len(res.refused), 1+len(followers))
	}
	for _, ch := range res.changes[1:] {
		if ch.text != "Back" || ch.after != leader.id {
			t.Errorf("%s took %q after %q", ch.r.id, ch.text, ch.after)
		}
	}
	if res, _ = decide(c, dir, true); len(res.changes) != 1 {
		t.Errorf("with --no-same: %d changes, want 1", len(res.changes))
	}
	recs[1][colYours] = "Never mind"
	writeSheet(t, dir, "case.csv", recs...)
	res, _ = decide(c, dir, false)
	for _, ch := range res.changes {
		if ch.r.id == followers[0].id && ch.text != "Never mind" {
			t.Errorf("a row with its own rewrite took %q", ch.text)
		}
	}
}

// TestRenameListsWhatQuotesTheName: after a name is applied, the import
// says where the old name is still written.
func TestRenameListsWhatQuotesTheName(t *testing.T) {
	c := realCatalog(t)
	r := c.byID["tech.calendar.name"]
	var out bytes.Buffer
	followUps(c, realRoot(t), []change{{r: r, text: "Almanac"}}, importOptions{dryRun: true}, &out)
	for _, want := range []string{"“Calendar” is now “Almanac”", "site/docs/technologies.md", "_test.go", "tech tables"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the follow-up after a rename should mention %q:\n%s", want, out.String())
		}
	}
}
