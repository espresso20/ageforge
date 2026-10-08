package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// The tests read the real source once and share the result. Reading is the
// slow part (a second or two under the race detector), so only the
// determinism test reads it a second time.

var (
	catOnce sync.Once
	theCat  *catalog
	catErr  error
)

func realRoot(t *testing.T) string {
	t.Helper()
	root, err := findRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func realCatalog(t *testing.T) *catalog {
	t.Helper()
	catOnce.Do(func() {
		root, err := findRoot(".")
		if err != nil {
			catErr = err
			return
		}
		theCat, catErr = read(root)
	})
	if catErr != nil {
		t.Fatal(catErr)
	}
	return theCat
}

// pick returns the first row that passes ok. The tests choose their rows by
// what they are (a log line with two format verbs, a label the code matches
// on), never by what they say: the wording is what this tool is for
// changing, and a test that quoted it would break on the first import.
func pick(t *testing.T, c *catalog, what string, ok func(*row) bool) *row {
	t.Helper()
	for _, r := range c.rows {
		if ok(r) {
			return r
		}
	}
	t.Fatalf("the export has no row that is %s", what)
	return nil
}

// plain reports whether a row has no pieces a rewrite must keep.
func plain(r *row) bool { return len(r.kept()) == 0 }

// verbsOf lists a row's format verbs in order.
func verbsOf(r *row) []string {
	var out []string
	for _, k := range r.kept() {
		if k.kind == tokVerb {
			out = append(out, k.text)
		}
	}
	return out
}

// tagged reports whether a row opens with a style tag, closes it at the
// end, and has nothing else to keep but glyphs.
func tagged(r *row) bool {
	ks := r.kept()
	if len(ks) < 2 || !isStyleTag(ks[0].text) || ks[len(ks)-1].text != "[-]" || !strings.HasSuffix(r.current, "[-]") {
		return false
	}
	for _, k := range ks[1 : len(ks)-1] {
		if k.kind != tokGlyph {
			return false
		}
	}
	return true
}

// ----- the export -----

// TestExportIsDeterministic: two exports of the same source, each from its
// own reading of it, are the same bytes.
func TestExportIsDeterministic(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	if err := export(realCatalog(t), a, io.Discard); err != nil {
		t.Fatal(err)
	}
	again, err := read(realRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := export(again, b, io.Discard); err != nil {
		t.Fatal(err)
	}
	names, _ := os.ReadDir(a)
	if len(names) != len(areas)+2 {
		t.Errorf("the export wrote %d files, want %d sheets, %s and %s", len(names), len(areas), firstHoursFile, readmeFile)
	}
	for _, n := range names {
		x, _ := os.ReadFile(filepath.Join(a, n.Name()))
		y, err := os.ReadFile(filepath.Join(b, n.Name()))
		if err != nil || !bytes.Equal(x, y) {
			t.Errorf("%s differs between two exports", n.Name())
		}
		if strings.HasSuffix(n.Name(), ".csv") {
			if !bytes.HasPrefix(x, []byte(bom)) {
				t.Errorf("%s does not start with a byte-order mark", n.Name())
			}
			recs, err := readSheet(filepath.Join(a, n.Name()))
			if err != nil {
				t.Errorf("%s does not read back: %v", n.Name(), err)
			}
			for _, rec := range recs {
				if rec[colYours] != "" || rec[colNotes] != "" {
					t.Errorf("%s: %s is exported with yours or notes filled", n.Name(), rec[colID])
				}
			}
		}
	}
}

// TestExportCoversKnownLines: one known line from each kind of place, with
// the kind, the sheet and the kept pieces it should have.
func TestExportCoversKnownLines(t *testing.T) {
	c := realCatalog(t)

	tech := c.byID["tech.calendar.description"]
	if tech == nil || tech.kind != kindVoice || tech.area != areaTechs || !strings.Contains(tech.where, c.w.called("tech:calendar")) {
		t.Errorf("tech.calendar.description: got %+v", tech)
	}
	if code := c.byID["tech.calendar.code"]; code == nil || code.max != 5 {
		t.Errorf("tech.calendar.code should carry the tree's five-letter limit, got %+v", code)
	}
	badge := c.byID["badge.age_stone_age.name"]
	if badge == nil || badge.kind != kindName || badge.area != areaBadges {
		t.Errorf("badge.age_stone_age.name: got %+v", badge)
	}
	if age := c.byID["age.stone_age.name"]; age == nil || age.age != 2 || !strings.HasPrefix(c.ageText(age), "02 ") {
		t.Errorf("age.stone_age.name: got %+v", age)
	}

	line := c.byID["flavor.expedition_fail.any.001"]
	if line == nil || line.kind != kindVoice || line.area != areaFlavorExp || line.rules&rulesSkel == 0 || line.max == 0 {
		t.Errorf("a flavor sentence (flavor.expedition_fail.any.001): got %+v", line)
	}
	slotted := pick(t, c, "a flavor sentence with a slot", func(r *row) bool {
		return r.rules&rulesSkel != 0 && strings.Contains(r.current, "~") && !strings.Contains(r.current, "{")
	})
	if got := keepColumn(slotted.kept()); got != "~" {
		t.Errorf("%s keeps %q, want ~", slotted.id, got)
	}
	if !strings.Contains(slotted.where, "flavor.words.") {
		t.Errorf("%s should name the word list that fills its slot: %s", slotted.id, slotted.where)
	}
	frag := c.byID["flavor.words.exp_fail_kit.001"]
	if frag == nil || frag.rules&rulesBank == 0 || !strings.Contains(frag.where, "~") || !strings.Contains(frag.where, "flavor.expedition_fail.") {
		t.Errorf("a word-list entry (flavor.words.exp_fail_kit.001) must say how it is used, got %+v", frag)
	}

	logLine := pick(t, c, "a log line with two format verbs", func(r *row) bool {
		return r.u.f.rel == "game/engine.go" && r.u.inCall("addLog") && len(verbsOf(r)) == 2 && len(r.kept()) == 2
	})
	if logLine.kind != kindMessage || logLine.area != areaLog || keepColumn(logLine.kept()) != strings.Join(verbsOf(logLine), " ") {
		t.Errorf("%s: kind %s, sheet %s, keep %q", logLine.id, logLine.kind, logLine.area.file, keepColumn(logLine.kept()))
	}
	title := pick(t, c, "a tagged interface string", func(r *row) bool { return r.u.f.rel == "ui/accounts_panel.go" && tagged(r) })
	if title.area != areaScreens || title.age != 0 || !strings.HasPrefix(keepColumn(title.kept()), title.kept()[0].text+" ") {
		t.Errorf("%s: sheet %s, keep %q, age %d", title.id, title.area.file, keepColumn(title.kept()), title.age)
	}
	help := c.byID["ui.help.commands.plan.build"]
	if help == nil || help.kind != kindMessage || !strings.Contains(help.where, "plan build") {
		t.Errorf("ui.help.commands.plan.build: got %+v", help)
	}

	// What must stay out: keys, command syntax, the dev console, struct tags.
	for _, r := range c.rows {
		switch {
		case r.current == "stone_age", r.current == "plan build <building> [count]", strings.HasPrefix(r.current, "json:"):
			t.Errorf("%s exports %q, which no player reads as text", r.id, r.current)
		case strings.HasPrefix(r.u.f.rel, "game/dev"):
			t.Errorf("%s comes from the dev console (%s)", r.id, r.u.f.rel)
		case strings.TrimSpace(r.where) == "" || r.kind == "" || r.area == nil:
			t.Errorf("%s is missing its where, kind or sheet", r.id)
		}
	}
}

// TestFirstHours: first-hours.csv holds rows of the main menu and the first
// six ages only, and every one of them is also in its own sheet.
func TestFirstHours(t *testing.T) {
	c := realCatalog(t)
	n := 0
	for _, r := range c.rows {
		if !inFirstHours(r) {
			continue
		}
		n++
		if r.age > firstHoursAges || r.unsure {
			t.Fatalf("%s (age %d) does not belong in %s", r.id, r.age, firstHoursFile)
		}
	}
	if n < 1000 {
		t.Errorf("%s would hold only %d rows", firstHoursFile, n)
	}
	menu := pick(t, c, "a main menu entry", func(r *row) bool { return r.u.f.rel == "ui/menu.go" && r.kind == kindLabel })
	hut := c.byID["building.hut.name"]
	if !inFirstHours(menu) || !inFirstHours(hut) || menu.seen > hut.seen {
		t.Errorf("the main menu should come before a building's name in %s", firstHoursFile)
	}
	if late := c.byID["age.transcendent_age.name"]; late == nil || inFirstHours(late) {
		t.Errorf("the last age's name should not be in %s", firstHoursFile)
	}
}

// TestExportCarriesFilledCells: exporting again into a folder keeps what
// was typed there, and sets aside what no longer has a row.
func TestExportCarriesFilledCells(t *testing.T) {
	c := realCatalog(t)
	dir := t.TempDir()
	r := c.byID["tech.calendar.description"]
	gone := c.record(r, "A line for a row that is gone.", "")
	gone[colID] = "tech.no_such_tech.description"
	writeSheet(t, dir, areaTechs.file, c.record(r, "Days are counted.", "check with the wiki"), gone)
	if err := export(c, dir, io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{areaTechs.file, firstHoursFile} {
		recs, err := readSheet(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, rec := range recs {
			if rec[colID] == r.id {
				found = rec[colYours] == "Days are counted." && rec[colNotes] == "check with the wiki"
			}
		}
		if !found {
			t.Errorf("%s lost the cells typed against %s", name, r.id)
		}
	}
	kept, err := readSheet(filepath.Join(dir, keptFile))
	if err != nil || len(kept) != 1 || kept[0][colID] != "tech.no_such_tech.description" {
		t.Errorf("%s should hold the one filled row with no place, got %v %v", keptFile, kept, err)
	}
}

// ----- the guards on what the export knows about -----

// TestConfigFieldsAreAllClassified fails when a struct in package config
// gains a field that can hold text and the export has not been told whether
// a player reads it. Add the field to fieldRules (data_rows.go) with its id,
// kind and where sentence, or to hiddenFields with why not.
func TestConfigFieldsAreAllClassified(t *testing.T) {
	root := realRoot(t)
	files, err := goFiles(filepath.Join(root, "config"))
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, name := range files {
		f, err := parser.ParseFile(fset, filepath.Join(root, "config", name), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return true
			}
			for _, fd := range st.Fields.List {
				if !holdsText(fd.Type) {
					continue
				}
				for _, id := range fd.Names {
					full := "config." + ts.Name.Name + "." + id.Name
					_, shown := fieldRules[full]
					_, hidden := hiddenFields[full]
					_, hiddenByName := hiddenFields[id.Name]
					if !shown && !hidden && !hiddenByName {
						t.Errorf("%s can hold text and the export does not know it (%s)", full, name)
					}
				}
			}
			return true
		})
	}
	for _, r := range realCatalog(t).rows {
		if strings.HasPrefix(r.id, "unclassified.") {
			t.Errorf("%s: %q sits in a config field the export has no rule for", r.id, r.current)
		}
	}
}

// holdsText reports whether a field's type has a string in it anywhere.
func holdsText(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == "string" {
			found = true
		}
		return !found
	})
	return found
}

// TestEveryPackageIsAccountedFor: each package of the module is either read
// for player text or listed with why it has none.
func TestEveryPackageIsAccountedFor(t *testing.T) {
	root := realRoot(t)
	read := map[string]bool{}
	for _, sp := range scanned {
		read[sp.dir] = true
	}
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if rel != "." && (strings.HasPrefix(d.Name(), ".") || d.Name() == "site" || d.Name() == "node_modules" || d.Name() == "words" || d.Name() == "data") {
			return filepath.SkipDir
		}
		names, _ := goFiles(path)
		if len(names) == 0 || read[rel] {
			return nil
		}
		if _, ok := notScanned[rel]; !ok {
			t.Errorf("package %s is neither read for player text nor listed in notScanned (source.go)", rel)
		}
		return nil
	})
}

// TestFlavorPoolsAreAllRead: every pool a catalog declares belongs to a
// catalog the export knows, and every sentence in the package is a row.
func TestFlavorPoolsAreAllRead(t *testing.T) {
	c := realCatalog(t)
	if len(c.fl.unread) > 0 {
		t.Errorf("these flavor pools match no catalog in moments (flavor_rows.go): %v", c.fl.unread)
	}
	sentences := 0
	for _, f := range c.m.files {
		if f.mode != modeFlavor {
			continue
		}
		for _, u := range c.m.units(f) {
			if s := u.st(); s.typ == "flavor.skel" && s.field == "Text" {
				sentences++
			}
		}
	}
	rows := 0
	for _, r := range c.rows {
		if r.rules&rulesSkel != 0 {
			rows++
		}
		if r.rules&rulesBank != 0 && r.unsure {
			t.Errorf("%s: a word list no sentence draws on", r.id)
		}
	}
	if rows != sentences || rows == 0 {
		t.Errorf("the flavor package has %d sentences and the export has %d rows for them", sentences, rows)
	}
	for _, mo := range moments {
		if _, ok := flavorMoments[mo.prefix]; !ok {
			t.Errorf("catalog %s has no flavor.Moment in flavorMoments (readme.go)", mo.prefix)
		}
		if exs, _ := c.examples(mo, 3); len(exs) != 3 {
			t.Errorf("catalog %s: the guide has %d assembled examples, want 3", mo.key, len(exs))
		}
	}
}

// TestRulesAreRead: the lists the import takes from the game's lint tests
// are all found, and each still has something in it.
func TestRulesAreRead(t *testing.T) {
	cr := realCatalog(t).rules
	for name, n := range map[string]int{
		"retired terms": len(cr.retiredTerms), "retired text": len(cr.retiredCopy), "the wiki's retired text": len(cr.retiredDoc),
		"config's style bans": len(cr.styleBans), "banned sentence shapes": len(cr.shapes), "restating phrases": len(cr.restating),
		"era words": len(cr.eraWords), "minimum words": cr.minWords, "maximum words": cr.maxWords, "longest line": cr.maxLine,
	} {
		if n == 0 {
			t.Errorf("the copy rules have nothing for: %s", name)
		}
	}
	if cr.bang == nil || cr.rawAmount == nil || cr.rawKey == nil || cr.verbLeft == nil {
		t.Error("a pattern of the copy rules was not read")
	}
	for word, homes := range cr.eraWords {
		for _, h := range homes {
			if _, ok := flavorSpans[h]; !ok {
				t.Errorf("era word %q names a span the export does not know: %s", word, h)
			}
		}
	}
}

// ----- small pieces -----

func TestKeptPieces(t *testing.T) {
	for _, tc := range []struct{ text, want string }{
		{"[red]Switch failed: %v[-]", "[red] %v [-]"},
		{"%d of %d%s", "%d %d %s"},
		{"50% of your gold", ""},
		{"Reach the {name}.", "{name}"},
		{"Enter: import  ·  Esc: cancel", "Enter · Esc"},
		{"Type 'catastrophe' to choose.\nOr wait.", `'catastrophe' \n`},
		{"  idle ", ""},
		{"[A[] Accept", "[A[]"},
		{"★  %s  ★", "★ %s ★"},
	} {
		if got := keepColumn(tokensOf(tc.text, false)); got != tc.want {
			t.Errorf("keep of %q = %q, want %q", tc.text, got, tc.want)
		}
	}
	if got := keepColumn(tokensOf("They came back with ~", true)); got != "~" {
		t.Errorf("a story sentence's slot: %q", got)
	}
}

func TestKeysAndWords(t *testing.T) {
	for text, key := range map[string]bool{
		"stone_age": true, "error": true, "ExpeditionSuccess": true, "account.json": true, "UTF-8": true, "TERM_PROGRAM": true,
		"Cancel": false, "FLOWS": false, "New game": false, " idle": false, "Hut": false,
	} {
		if keyLike(text) != key {
			t.Errorf("keyLike(%q) = %v", text, !key)
		}
	}
	for text, word := range map[string]bool{
		"Cancel": true, " pts": true, "%d of %d": true, "  ppppp   ppppp  ": false, "FF S S FF": false, "%s_%d": false, "[-] ": false, "═══": false,
	} {
		if hasWord(text) != word {
			t.Errorf("hasWord(%q) = %v", text, !word)
		}
	}
	c := realCatalog(t)
	for text, cmd := range map[string]bool{
		"gather wood 5": true, "plan build <building> [count]": true, "account badges": true, "build hut": true,
		"research speed": false, "trade route income": false, " workers": false, "Build a hut": false,
	} {
		if c.isCommand(text) != cmd {
			t.Errorf("isCommand(%q) = %v", text, !cmd)
		}
	}
}

// ----- helpers for the import tests -----

// writeSheet writes records as a sheet in dir.
func writeSheet(t *testing.T, dir, name string, recs ...[]string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), sheet(recs), 0o644); err != nil {
		t.Fatal(err)
	}
}

// tokens returns a Go file's tokens, comments included, as text.
func tokens(t *testing.T, src []byte) []string {
	t.Helper()
	var s scanner.Scanner
	fset := token.NewFileSet()
	s.Init(fset.AddFile("", fset.Base(), len(src)), src, nil, scanner.ScanComments)
	var out []string
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			return out
		}
		if lit == "" {
			lit = tok.String()
		}
		out = append(out, lit)
	}
}
