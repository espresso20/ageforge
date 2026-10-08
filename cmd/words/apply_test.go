package main

import (
	"bytes"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/espresso20/ageforge/config"
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

// fill types yours against ids in the area sheets of dir.
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

// beforeTag puts extra words in front of a line's closing style tag.
func beforeTag(text, extra string) string {
	return strings.TrimSuffix(text, "[-]") + extra + "[-]"
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

	desc := c.byID["tech.calendar.description"]
	title := pick(t, c, "a tagged interface string", func(r *row) bool { return r.u.f.rel == "ui/accounts_panel.go" && tagged(r) })
	label := pick(t, c, "a label the code matches on", func(r *row) bool { return len(r.twins) == 1 && plain(r) })
	line := c.byID["flavor.expedition_fail.any.001"]
	if desc == nil || line == nil {
		t.Fatal("the partial tree is missing the tech description or the flavor line")
	}
	yours := map[string]string{
		desc.id:  desc.current + " Every year.",
		title.id: beforeTag(title.current, " here"),
		label.id: label.current + " now",
		line.id:  line.current + " again",
	}
	// What should change, as source text: each row's literal, and for the
	// label the comparison that reads it.
	wantMoved := map[[2]string]int{}
	for id, text := range yours {
		r := c.byID[id]
		was := string(before[r.u.f.rel][r.u.start:r.u.end])
		wantMoved[[2]string{was, quote(r.u, text)}] += 1 + len(r.twins)
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
	for id, text := range yours {
		if r := c2.byID[id]; r == nil || r.current != text {
			t.Errorf("%s does not read %q after the import", id, text)
		}
		if led[id] != hashText(text) {
			t.Errorf("the ledger has %q for %s", led[id], id)
		}
	}
	if len(led) != len(yours) {
		t.Errorf("the ledger holds %d rows, want %d", len(led), len(yours))
	}
	_, total, gone := count(c2, led)
	if total.doneLines != len(yours) || total.changed != 0 || gone != 0 {
		t.Errorf("status counts %d rewritten, %d changed since, %d gone; want %d, 0, 0", total.doneLines, total.changed, gone, len(yours))
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
	edited := bytes.Replace(after["config/research.go"], []byte(" Every year."), []byte(" Every single year."), 1)
	if err := os.WriteFile(path, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	c3, err := read(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, total, _ := count(c3, led); total.changed != 1 || total.doneLines != len(yours)-1 {
		t.Errorf("after a hand edit: %d changed since, %d rewritten; want 1 and %d", total.changed, total.doneLines, len(yours)-1)
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
	writeSheet(t, dir, areaTechs.file, c.record(r, r.current+" Every year.", ""))
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
	// The rows are chosen by what they are; each rewrite is built from the
	// row's own text, so the cases hold whatever the game says today.
	two := pick(t, c, "a log line with two different format verbs", func(r *row) bool {
		v := verbsOf(r)
		return r.rules&rulesUI != 0 && len(v) == 2 && v[0] != v[1] && len(r.kept()) == 2
	})
	v := verbsOf(two)
	swapped := strings.Replace(strings.Replace(strings.Replace(two.current, v[0], "\x00", 1), v[1], v[0], 1), "\x00", v[1], 1)
	title := pick(t, c, "a tagged interface string with a glyph", func(r *row) bool {
		if !tagged(r) || r.kept()[0].text == "[yellow]" {
			return false
		}
		for _, k := range r.kept() {
			if k.kind == tokGlyph {
				return true
			}
		}
		return false
	})
	open := title.kept()[0].text
	var glyph string
	for _, k := range title.kept() {
		if k.kind == tokGlyph {
			glyph = k.text
		}
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(title.current, open), "[-]")
	prose := pick(t, c, "a plain log sentence", func(r *row) bool {
		return r.rules&rulesUI != 0 && plain(r) && r.kind == kindMessage && strings.Contains(r.current, " ") &&
			strings.HasPrefix(r.u.f.rel, "ui/") && !strings.ContainsAny(r.current, "—!")
	})
	sentence := pick(t, c, "a flavor sentence for any age with no slot", func(r *row) bool {
		return r.rules&rulesSkel != 0 && plain(r) && len(r.eras) == 0
	})
	slotted := pick(t, c, "a flavor sentence with a slot", func(r *row) bool {
		return r.rules&rulesSkel != 0 && strings.Contains(r.current, "~") && !strings.Contains(r.current, "{")
	})
	entry := pick(t, c, "a word-list entry", func(r *row) bool {
		return r.rules&rulesBank != 0 && unicode.IsLower([]rune(r.current)[0])
	})
	saveWord := pick(t, c, "a save-name word", func(r *row) bool { return r.area == areaSaveNames })
	quip := pick(t, c, "a milestone quip", func(r *row) bool { return r.rules&rulesNoMark != 0 && plain(r) })
	desc, code := c.byID["tech.calendar.description"], c.byID["tech.calendar.code"]
	var otherCode string
	for _, tech := range config.Technologies() {
		if tech.Key != "calendar" {
			otherCode = tech.Code
			break
		}
	}
	// A word the catalog holds to one span of ages, and a phrase it bans.
	var eraWord string
	words := make([]string, 0, len(c.rules.eraWords))
	for w := range c.rules.eraWords {
		words = append(words, w)
	}
	sort.Strings(words)
	for _, w := range words {
		if len(c.rules.eraWords[w]) == 1 && !strings.Contains(w, " ") {
			eraWord = w
			break
		}
	}
	upper := string(unicode.ToUpper([]rune(entry.current)[0])) + string([]rune(entry.current)[1:])

	stale := c.record(desc, desc.current+" Every year.", "")
	stale[colCurrent] = "What this line used to say."
	gone := c.record(desc, desc.current+" Every year.", "")
	gone[colID] = "tech.no_such_tech.description"

	cases := []struct {
		name string
		rec  []string
		why  string
	}{
		{"id gone", gone, "no longer in the source"},
		{"source changed", stale, "has changed in the source since the export"},
		{"verb missing", c.record(two, strings.Replace(two.current, v[1], "it", 1), ""), "must keep the format verbs " + v[0] + " " + v[1]},
		{"verbs out of order", c.record(two, swapped, ""), "format verbs out of order"},
		{"tag altered", c.record(title, "[yellow]"+inner+"[-]", ""), "is missing " + open},
		{"glyph missing", c.record(title, strings.ReplaceAll(title.current, glyph, ""), ""), "is missing " + glyph},
		{"tags unbalanced", c.record(title, "[-]"+inner+open, ""), "unbalanced style tags"},
		{"slot missing", c.record(slotted, strings.Replace(slotted.current, "~", "it", 1), ""), "is missing ~"},
		{"over max", c.record(code, "TOOLONGACODE", ""), "the most is 5"},
		{"code taken", c.record(code, otherCode, ""), "is the card code"},
		{"em dash", c.record(prose, prose.current+" — and more", ""), "em dash"},
		{"exclamation", c.record(prose, "Stop! "+prose.current, ""), "exclamation mark"},
		{"retired term", c.record(prose, prose.current+" The villagers agree.", ""), "retired term"},
		{"config style", c.record(desc, desc.current+" In full colour.", ""), "the style guard for game data bans"},
		{"config dash", c.record(desc, desc.current+" – more.", ""), "has a dash"},
		{"quip markup", c.record(quip, quip.current+" [x]", ""), "square bracket"},
		{"sentence shape", c.record(sentence, "It was not the cold that took them, but the dark", ""), "sentence shape the catalog bans"},
		{"sentence length", c.record(sentence, "Two words", ""), "a catalog sentence is"},
		{"era word", c.record(sentence, "Somebody left the "+eraWord+" out in the rain again", ""), "which belongs to"},
		{"full stop", c.record(sentence, sentence.current+".", ""), "the game adds the full stop"},
		{"restating", c.record(sentence, "Everybody said "+c.rules.restating[0]+" and went home early", ""), "repeats what the log line above it already said"},
		{"word-list entry", c.record(entry, upper, ""), "starts with a capital letter"},
		{"save name", c.record(saveWord, "O'Brien", ""), "letters and single spaces"},
	}
	dir := t.TempDir()
	good := c.byID["tech.language.description"]
	goodRec := c.record(good, good.current+" More so.", "")
	for _, tc := range cases {
		writeSheet(t, dir, "case.csv", goodRec, tc.rec)
		res, err := decide(c, dir, false)
		if err != nil {
			t.Fatal(err)
		}
		if len(res.refused) != 1 || !strings.Contains(res.refused[0].why, tc.why) {
			t.Errorf("%s: want one refusal saying %q, got %+v", tc.name, tc.why, res.refused)
		}
		if len(res.changes) != 1 || res.changes[0].r.id != good.id {
			t.Errorf("%s: the good row beside the refused one should still apply, got %d changes", tc.name, len(res.changes))
		}
	}

	// The same id filled two ways in two files: both are refused.
	writeSheet(t, dir, "case.csv", c.record(desc, desc.current+" Every year.", ""))
	writeSheet(t, dir, firstHoursFile, c.record(desc, desc.current+" Each year.", ""))
	res, err := decide(c, dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.changes) != 0 || len(res.refused) != 2 || !strings.Contains(res.refused[0].why, "filled differently in two files") {
		t.Errorf("an id filled two ways: got %d changes and %+v", len(res.changes), res.refused)
	}
	// Filled the same way in both: it counts once.
	writeSheet(t, dir, firstHoursFile, c.record(desc, desc.current+" Every year.", ""))
	if res, _ = decide(c, dir, false); len(res.changes) != 1 || len(res.refused) != 0 {
		t.Errorf("an id filled the same way twice: got %d changes, %d refusals", len(res.changes), len(res.refused))
	}
}

// TestSameAs: a blank row takes the rewrite of the row its same_as names,
// unless it is switched off or the row has a rewrite of its own.
func TestSameAs(t *testing.T) {
	c := realCatalog(t)
	// The label with the most rows that are the same as it.
	follow := map[string][]*row{}
	for _, r := range c.rows {
		if r.sameAs != "" {
			follow[r.sameAs] = append(follow[r.sameAs], r)
		}
	}
	var leader *row
	for _, r := range c.rows {
		if r.kind == kindLabel && plain(r) && r.rules == rulesUI && !strings.Contains(r.current, " ") && len(follow[r.id]) >= 2 &&
			(leader == nil || len(follow[r.id]) > len(follow[leader.id])) {
			leader = r
		}
	}
	if leader == nil {
		t.Fatal("no label has two rows that are the same as it")
	}
	followers := follow[leader.id]
	rewrite := leader.current + " now"
	dir := t.TempDir()
	recs := [][]string{c.record(leader, rewrite, "")}
	for _, f := range followers {
		recs = append(recs, c.record(f, "", ""))
	}
	writeSheet(t, dir, "case.csv", recs...)

	res, err := decide(c, dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.changes) != 1+len(followers) || len(res.refused) != 0 {
		t.Fatalf("with same_as: %d changes, %d refusals (%+v), want %d and 0", len(res.changes), len(res.refused), res.refused, 1+len(followers))
	}
	for _, ch := range res.changes[1:] {
		if ch.text != rewrite || ch.after != leader.id {
			t.Errorf("%s took %q after %q", ch.r.id, ch.text, ch.after)
		}
	}
	if res, _ = decide(c, dir, true); len(res.changes) != 1 {
		t.Errorf("with --no-same: %d changes, want 1", len(res.changes))
	}
	own := leader.current + " then"
	recs[1][colYours] = own
	writeSheet(t, dir, "case.csv", recs...)
	res, _ = decide(c, dir, false)
	for _, ch := range res.changes {
		if ch.r.id == followers[0].id && ch.text != own {
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
	followUps(c, realRoot(t), []change{{r: r, text: "Daybook"}}, importOptions{dryRun: true}, &out)
	for _, want := range []string{"“" + r.current + "” is now “Daybook”", "site/docs/technologies.md", "tech tables"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the follow-up after a rename should mention %q:\n%s", want, out.String())
		}
	}
	// A line that is not a name, which a wiki page quotes word for word.
	root := realRoot(t)
	for _, q := range loadQuoters(root) {
		if q.kind != "wiki" {
			continue
		}
		for _, row := range c.rows {
			text := strings.TrimSpace(row.current)
			if row.kind == kindName || len([]rune(text)) < 12 || !strings.Contains(q.text, text) {
				continue
			}
			got := strings.Join(wikiQuotes(root, []change{{r: row, text: text + " More."}}), "\n")
			if !strings.Contains(got, q.rel+": "+row.id) {
				t.Errorf("%s quotes %s and should be listed, got %q", q.rel, row.id, got)
			}
			return
		}
	}
}
