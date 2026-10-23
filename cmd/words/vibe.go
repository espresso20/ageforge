package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/words"
)

// vibe.go: the two commands that look after VibeMode's script.
//
//	vibe-export   what the script still has no line for, sheet by sheet
//	vibe-lint     whether what the script has is sound
//
// The script is words/vibe/*.csv: a row id and a line. The ids are the
// export's. A row reaches the screen by one of three ways, and only those
// are "wired": the ruleset overlay (a text of a definition), the lexicon
// (a word for a kind of person) or a vibe.Line call in the code. A row in
// the script that nothing reads yet is harmless and lint says so.

// vibeTodoDir is where vibe-export writes, from the repository root. It is in
// .gitignore: the sheets are working files for filling the script in.
const vibeTodoDir = "words/vibe-todo"

// vibeColumns are the columns of a to-do sheet. id and text are the ones the
// script reads back; text is left empty to be filled in.
var vibeColumns = []string{"id", "where", "kind", "age", "keep", "max", "plain", "wired", "text"}

// lineCall is a vibe call in the source: vibe.Line(id, plain), or
// vibe.Sprintf or vibe.Errorf(id, format, args...).
type lineCall struct {
	id, plain string
	kind      string // Line, E, Sprintf or Errorf
	file      string
	line      int
	literal   bool // plain is a string literal (or a constant sum of them), which the lint can compare
	args      int  // for a format, how many arguments follow it; -1 when they are spread (args...)
}

// lineCalls finds every vibe call in the parsed source.
func lineCalls(m *module) []lineCall {
	var calls []lineCall
	for _, f := range m.files {
		ast.Inspect(f.ast, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || !isVibeCall(call) {
				return true
			}
			kind := call.Fun.(*ast.SelectorExpr).Sel.Name
			if (kind == "Line" || kind == "E") && len(call.Args) != 2 {
				return true
			}
			id, ok := stringOf(call.Args[0])
			if !ok {
				return true
			}
			lc := lineCall{id: id, kind: kind, file: f.rel, line: m.fset.Position(call.Pos()).Line}
			lc.plain, lc.literal = stringLit(call.Args[1])
			if kind != "Line" && kind != "E" {
				lc.args = len(call.Args) - 2
				if call.Ellipsis.IsValid() {
					lc.args = -1
				}
			}
			calls = append(calls, lc)
			return true
		})
	}
	return calls
}

// stringOf is the value of a string literal.
func stringOf(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return s, true
}

// vibeWiring is what the game reads from the script.
type vibeWiring struct {
	data    map[string]words.Cell // the texts of the ruleset, by row id
	lexicon map[string]words.Cell // the words of the lexicon, by row id
	code    map[string]lineCall   // the vibe.Line calls, by row id
	calls   []lineCall            // all of them, in file order
	tables  map[string]bool       // the row ids of table entries the code reads through vibe.Line when it draws them
}

func newVibeWiring(c *catalog) *vibeWiring {
	w := &vibeWiring{data: map[string]words.Cell{}, lexicon: map[string]words.Cell{}, code: map[string]lineCall{}, tables: map[string]bool{}}
	for _, r := range c.rows {
		if r.u.vibeTable {
			w.tables[r.id] = true
		}
	}
	for _, cell := range words.Cells(rules.FromConfig()) {
		if _, dup := w.data[cell.ID]; !dup {
			w.data[cell.ID] = cell
		}
	}
	for _, cell := range config.LexiconCells() {
		w.lexicon[cell.ID] = cell
	}
	w.calls = lineCalls(c.m)
	for _, lc := range w.calls {
		if _, dup := w.code[lc.id]; !dup {
			w.code[lc.id] = lc
		}
	}
	return w
}

// wired reports whether the game reads the script's line for the row id.
func (w *vibeWiring) wired(id string) bool {
	if _, ok := w.data[id]; ok {
		return true
	}
	if _, ok := w.lexicon[id]; ok {
		return true
	}
	if w.tables[id] {
		return true
	}
	_, ok := w.code[id]
	return ok
}

// scriptDir is the folder of the script's sheets under root, or dir when one
// was given.
func scriptDir(root, dir string) string {
	if dir != "" {
		return dir
	}
	return filepath.Join(root, filepath.FromSlash(words.VibeDir))
}

// readScript reads every sheet in dir. It returns the rows with text, in file
// order, and what was wrong with the sheets.
func readScript(dir string) ([]words.Row, []string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	var rows []words.Row
	var problems []string
	seen := map[string]words.Row{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".csv") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, nil, err
		}
		got, probs := words.ParseRows(e.Name(), f)
		f.Close()
		problems = append(problems, probs...)
		for _, r := range got {
			if first, dup := seen[r.ID]; dup {
				problems = append(problems, fmt.Sprintf("%s:%d: the id %s is also in %s:%d", r.File, r.Line, r.ID, first.File, first.Line))
				continue
			}
			seen[r.ID] = r
			rows = append(rows, r)
		}
	}
	return rows, problems, nil
}

// vibeExport writes a sheet for each area of the export, and one for the
// lexicon, of the rows the script has no line for yet.
func vibeExport(c *catalog, root, scriptDirArg, dir string, out io.Writer) error {
	rows, _, err := readScript(scriptDir(root, scriptDirArg))
	if err != nil {
		return err
	}
	have := map[string]bool{}
	for _, r := range rows {
		have[r.ID] = true
	}
	wiring := newVibeWiring(c)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// A sheet that is not written this time must not stay from the last.
	old, _ := filepath.Glob(filepath.Join(dir, "*.csv"))
	for _, f := range old {
		_ = os.Remove(f)
	}
	yes := func(id string) string {
		if wiring.wired(id) {
			return "yes"
		}
		return "no"
	}

	type tally struct{ rows, wired, left int }
	counts := map[string]*tally{}
	leftBy := map[string]int{} // rows left alone for good, by reason
	var total, totalWired int
	var names []string
	write := func(file string, recs [][]string) error {
		if len(recs) == 0 {
			return nil
		}
		names = append(names, file)
		return os.WriteFile(filepath.Join(dir, file), sheetWith(vibeColumns, recs), 0o644)
	}
	for _, a := range areas {
		var recs [][]string
		left := 0
		for _, r := range c.rows {
			if r.area != a || have[r.id] {
				continue
			}
			if reason := c.wireVerdict(r).reason; forGood(reason) {
				left++ // not asked of the script: nothing will ever read a line for it
				leftBy[reason]++
				continue
			}
			rec := make([]string, len(vibeColumns))
			rec[0] = r.id
			rec[1] = r.where
			rec[2] = r.kind
			rec[3] = c.ageText(r)
			rec[4] = keepColumn(r.kept())
			if r.max > 0 {
				rec[5] = strconv.Itoa(r.max)
			}
			rec[6] = r.current
			rec[7] = yes(r.id)
			recs = append(recs, rec)
		}
		if err := write(a.file, recs); err != nil {
			return err
		}
		t := &tally{rows: len(recs), left: left}
		for _, rec := range recs {
			if rec[7] == "yes" {
				t.wired++
			}
		}
		counts[a.file] = t
		total += t.rows
		totalWired += t.wired
	}
	var recs [][]string
	for _, cell := range config.LexiconCells() {
		if have[cell.ID] {
			continue
		}
		parts := strings.Split(cell.ID, ".") // lexicon.<scope>.<key>.<form>
		form := "singular"
		if parts[len(parts)-1] == "many" {
			form = "plural"
		}
		where := "The " + form + " of the lexicon's word " + parts[2] + " in "
		age := ""
		if parts[1] == "plain" {
			where += "every age that has no word of its own."
		} else {
			where += "the " + strings.ReplaceAll(parts[1], "_", " ") + "."
			if n := c.w.ageOf(parts[1]); n > 0 {
				age = fmt.Sprintf("%02d %s", n, c.w.ageNames[n-1])
			}
		}
		recs = append(recs, []string{cell.ID, where, kindName, age, "", "", cell.Text, "yes", ""})
	}
	if err := write("lexicon.csv", recs); err != nil {
		return err
	}
	counts["lexicon.csv"] = &tally{rows: len(recs), wired: len(recs)}
	total += len(recs)
	totalWired += len(recs)

	fmt.Fprintf(out, "Wrote %d sheets to %s\n\n%-34s %7s %7s %7s\n", len(names), dir, "sheet", "rows", "wired", "left")
	files := make([]string, 0, len(counts))
	for f := range counts {
		files = append(files, f)
	}
	sort.Strings(files)
	totalLeft := 0
	for _, f := range files {
		if t := counts[f]; t.rows > 0 || t.left > 0 {
			fmt.Fprintf(out, "%-34s %7d %7d %7d\n", f, t.rows, t.wired, t.left)
			totalLeft += t.left
		}
	}
	fmt.Fprintf(out, "%-34s %7d %7d %7d\n", "total", total, totalWired, totalLeft)
	fmt.Fprintf(out, "\n%d rows of the script are written already. wired = yes: the game reads that row once it has a line.\n", len(have))
	if totalLeft > 0 {
		fmt.Fprintf(out, "left: rows the sheets leave out, because the game will never read a line for them:\n")
		reasons := make([]string, 0, len(leftBy))
		for reason := range leftBy {
			reasons = append(reasons, reason)
		}
		sort.Strings(reasons)
		for _, reason := range reasons {
			fmt.Fprintf(out, "  %5d  %s\n", leftBy[reason], reason)
		}
	}
	return nil
}

// vibeLint checks the script in dir against the export and the source, and
// writes what it finds to out. It returns the number of errors; a row that is
// sound but that nothing in the game reads yet is a note, not an error.
func vibeLint(c *catalog, root, dir string, out io.Writer) (int, error) {
	rows, problems, err := readScript(scriptDir(root, dir))
	if err != nil {
		return 0, err
	}
	wiring := newVibeWiring(c)
	errs := 0
	fail := func(format string, args ...any) {
		errs++
		fmt.Fprintf(out, "error: "+format+"\n", args...)
	}
	for _, p := range problems {
		fail("%s", p)
	}

	notWired := map[string]int{}
	for _, r := range rows {
		where := fmt.Sprintf("%s:%d: %s", r.File, r.Line, r.ID)
		if cell, ok := wiring.lexicon[r.ID]; ok {
			_ = cell
			continue
		}
		row, ok := c.byID[r.ID]
		if !ok {
			fail("%s is not a row of the export or the lexicon", where)
			continue
		}
		if reason := c.wireVerdict(row).reason; forGood(reason) {
			fail("%s is left alone for good (%s): nothing reads a line for it", where, reason)
			continue
		}
		for _, p := range checkKeep(row, r.Text) {
			fail("%s %s", where, p)
		}
		// The spaces and line breaks a text starts and ends with are part of how the
		// code joins it to what is next to it, so a line keeps them as they are.
		lead, trail := edges(row.current)
		if l, tr := edges(r.Text); l != lead || tr != trail {
			fail("%s must start with %q and end with %q, as the original does, and starts with %q and ends with %q", where, lead, trail, l, tr)
		}
		if row.max > 0 && utf8.RuneCountInString(r.Text) > row.max {
			fail("%s is %d characters and the row allows %d", where, utf8.RuneCountInString(r.Text), row.max)
		}
		if !wiring.wired(r.ID) {
			notWired[row.area.file]++
		}
	}
	// A vibe.Line call must name a row of the export, and quote its text.
	for _, lc := range wiring.calls {
		where := fmt.Sprintf("%s:%d: vibe.%s(%q)", lc.file, lc.line, lc.kind, lc.id)
		row, ok := c.byID[lc.id]
		switch {
		case !ok:
			fail("%s names a row the export does not have", where)
		case !lc.literal:
			fail("%s must give its plain text as a string literal, so the lint can compare it", where)
		case row.current != lc.plain:
			fail("%s gives %q and the export reads %q at that id", where, lc.plain, row.current)
		}
	}

	var files []string
	for f := range notWired {
		files = append(files, f)
	}
	sort.Strings(files)
	n := 0
	for _, f := range files {
		n += notWired[f]
	}
	if n > 0 {
		fmt.Fprintf(out, "note: %d rows have a line and nothing reads them yet:", n)
		for _, f := range files {
			fmt.Fprintf(out, " %s %d", f, notWired[f])
		}
		fmt.Fprintln(out)
	}
	fmt.Fprintf(out, "%d rows in the script, %d errors\n", len(rows), errs)
	return errs, nil
}
