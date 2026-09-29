package ui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// copy_guard_test.go holds the systemic wording guards. Each one scans the
// string literals of every non-test Go file that produces player text and
// fails on a pattern the wording audit retired. They read source, not
// rendered output, so they catch a bad string before anything calls it.

// copyDirs are the packages whose string literals reach the player.
var copyDirs = []string{"ui", "game", "boon"}

// copyExemptFiles are skipped by every guard in this file.
//   - dev-only consoles and dumps are read by us, not players;
//   - art, glyphs and generated names are not sentences.
var copyExemptFiles = map[string]string{
	"game/devmode.go":                "dev console",
	"game/devcmd.go":                 "dev console",
	"ui/citymap/debug.go":            "debug overlay",
	"game/updater.go":                "self-update plumbing, not game text",
	"ui/splash_canvas.go":            "art",
	"ui/braille.go":                  "chart glyphs",
	"ui/wonder_icon.go":              "art",
	"game/savenames.go":              "generated save names",
	"ui/theme_widgets.go":            "widget chrome",
	"ui/copy_guard_test.go":          "this file",
	"game/expedition_flavor.go":      "flavor catalog glue",
	"ui/mapstyle/capture/capture.go": "dev capture tool: writes HTML/JS for review pages, never shown in game",
}

// copyExemptLiterals are literal substrings exempt from every guard.
//   - The civilization-log markers in game/catastrophe.go are parsed back out
//     of saved history by countCatastropheOutcomes, so rewording them would
//     break the stats of existing saves. note: store the outcome as data
//     first (audit systemic 15), then reword and drop these.
var copyExemptLiterals = []string{
	" — Endured ",
	" — Succumbed to ",
}

type copyLit struct {
	pos  string
	text string
	// usageForm: a registry help form in ui/commands.go ("sell <building>
	// [count]"). Forms reach the screen only through helpForm (tview.Escape)
	// or a usage line in the log (safeTags); TestUsageFormsSurviveTview
	// proves both, so the bracket guard skips them.
	usageForm bool
	// escaped: the literal is an argument to tview.Escape or lit(), so its
	// brackets print as written.
	escaped bool
}

// isEscapeCall reports whether c is tview.Escape(...) or lit(...).
func isEscapeCall(c *ast.CallExpr) bool {
	switch f := c.Fun.(type) {
	case *ast.Ident:
		return f.Name == "lit"
	case *ast.SelectorExpr:
		if pkg, ok := f.X.(*ast.Ident); ok {
			return pkg.Name == "tview" && f.Sel.Name == "Escape"
		}
	}
	return false
}

// collectCopyLiterals returns every string literal in the non-test Go files
// of copyDirs, minus the exemptions.
func collectCopyLiterals(t *testing.T) []copyLit {
	t.Helper()
	root := ".."
	var out []copyLit
	fset := token.NewFileSet()
	for _, dir := range copyDirs {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			rel = filepath.ToSlash(rel)
			if _, skip := copyExemptFiles[rel]; skip {
				return nil
			}
			f, perr := parser.ParseFile(fset, path, nil, 0)
			if perr != nil {
				return perr
			}
			var forms map[token.Pos]bool
			if rel == "ui/commands.go" {
				forms = registryFormLits(f)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				// Skip struct tags and import paths.
				switch x := n.(type) {
				case *ast.ImportSpec:
					return false
				case *ast.CallExpr:
					// Text passed through tview.Escape or lit() prints its
					// brackets literally, so it is not checked for them. The
					// other guards still see it via the literal itself.
					if isEscapeCall(x) {
						for _, a := range x.Args {
							if bl, ok := a.(*ast.BasicLit); ok && bl.Kind == token.STRING {
								if s, err := strconv.Unquote(bl.Value); err == nil {
									p := fset.Position(bl.Pos())
									out = append(out, copyLit{pos: rel + ":" + strconv.Itoa(p.Line), text: s, escaped: true})
								}
							}
						}
						return false
					}
				case *ast.Field:
					if x.Tag != nil {
						ast.Inspect(x.Type, func(ast.Node) bool { return true })
						return false
					}
				case *ast.BasicLit:
					if x.Kind != token.STRING {
						return true
					}
					s, uerr := strconv.Unquote(x.Value)
					if uerr != nil {
						return true
					}
					for _, ex := range copyExemptLiterals {
						if strings.Contains(s, ex) {
							return true
						}
					}
					p := fset.Position(x.Pos())
					out = append(out, copyLit{pos: rel + ":" + strconv.Itoa(p.Line), text: s, usageForm: forms[x.Pos()]})
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return out
}

// registryFormLits finds the help-form literals of the command registry: the
// Form of each Usage{form, text}, the form argument of sub(name, form, ...)
// and of confirmYes(form, text).
func registryFormLits(f *ast.File) map[token.Pos]bool {
	out := map[token.Pos]bool{}
	mark := func(e ast.Expr) {
		if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			out[lit.Pos()] = true
		}
	}
	isUsage := func(e ast.Expr) bool {
		id, ok := e.(*ast.Ident)
		return ok && id.Name == "Usage"
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CompositeLit:
			if at, ok := x.Type.(*ast.ArrayType); ok && isUsage(at.Elt) {
				for _, el := range x.Elts {
					if cl, ok := el.(*ast.CompositeLit); ok && len(cl.Elts) > 0 {
						mark(cl.Elts[0])
					}
				}
			}
			if isUsage(x.Type) && len(x.Elts) > 0 {
				mark(x.Elts[0])
			}
		case *ast.CallExpr:
			if id, ok := x.Fun.(*ast.Ident); ok {
				switch {
				case id.Name == "sub" && len(x.Args) > 1:
					mark(x.Args[1])
				case id.Name == "confirmYes" && len(x.Args) > 0:
					mark(x.Args[0])
				}
			}
		}
		return true
	})
	return out
}

// bracketWordRe finds [word] spans that tview could read as a color tag.
var bracketWordRe = regexp.MustCompile(`\[([A-Za-z][A-Za-z0-9#:\-]*)\]`)

// TestCopyNoEatenBrackets fails when a literal carries a [word] that tview
// would swallow as a color tag ("[count]", "[allied]", "[current]"). Real
// color tags ([gold], [-], [accent:chip:b]) pass. Write the brackets
// escaped ("[count[]") or pass the text through lit().
func TestCopyNoEatenBrackets(t *testing.T) {
	for _, l := range collectCopyLiterals(t) {
		if l.usageForm || l.escaped {
			continue
		}
		for _, m := range bracketWordRe.FindAllStringSubmatchIndex(l.text, -1) {
			content := l.text[m[2]:m[3]]
			// "[word[]" is tview's escape; the regex stops at the first ']'
			// so check the byte after the match for the escape form.
			if m[1] < len(l.text) && m[1] >= 1 && l.text[m[1]-1] == ']' && m[1]-2 >= 0 && l.text[m[1]-2] == '[' {
				continue
			}
			if tviewEatsTag(content) && !isStyleTag(content) {
				t.Errorf("%s: %q would vanish in tview (write it as [%s[] or wrap in lit())", l.pos, "["+content+"]", content)
			}
		}
	}
}

// TestCopyNoEmDashes fails on an em dash in player text. Use a period, colon,
// comma or parentheses.
func TestCopyNoEmDashes(t *testing.T) {
	for _, l := range collectCopyLiterals(t) {
		if strings.Contains(l.text, "—") {
			t.Errorf("%s: em dash in player text: %q", l.pos, l.text)
		}
	}
}

// retiredTerms are glossary variants the wording audit replaced. Matching is
// case-insensitive on whole words inside literals that read as prose (they
// contain a space), so map keys and identifiers are not flagged.
var retiredTerms = []struct {
	re  *regexp.Regexp
	use string
}{
	{regexp.MustCompile(`(?i)\bstandings\b|%d standing\b|[0-9] standing\b|faction standing`), "opinion"},
	{regexp.MustCompile(`(?i)\bfavours?\b`), "boon (or Goodwill for the deal kind)"},
	{regexp.MustCompile(`(?i)\bvillagers?\b`), "worker"},
	{regexp.MustCompile(`(?i)\bcolour`), "color"},
	{regexp.MustCompile(`(?i)\bharbour`), "harbor"},
	{regexp.MustCompile(`(?i)\bcentre`), "center"},
	{regexp.MustCompile(`(?i)\bcivilisation`), "civilization"},
	{regexp.MustCompile(`(?i)\bdefence\b`), "defense"},
	{regexp.MustCompile(`(?i)\blabour\b`), "labor"},
	{regexp.MustCompile(`(?i)\bcancelled\b`), "canceled"},
	{regexp.MustCompile(`(?i)\bpop cap\b`), "housing"},
	{regexp.MustCompile(`(?i)\bhouse left\b`), "housing left"},
	{regexp.MustCompile(`\bproduction_all\b`), "all production"},
	{regexp.MustCompile(`\d/t\b`), "/tick"},
	{regexp.MustCompile(`(?i)\bmilitary cap\b`), "soldier storage"},
	{regexp.MustCompile(`(?i)unlocks \+0\.5x`), "raises the speed cap by 0.5x"},
	{regexp.MustCompile(`\bESC\b`), "Esc (e.g. \"(Esc to close)\")"},
	{regexp.MustCompile(`\(s\)`), "a plural helper (textfmt.Count)"},
}

// TestCopyGlossaryTerms fails when a retired term shows up in player text.
func TestCopyGlossaryTerms(t *testing.T) {
	for _, l := range collectCopyLiterals(t) {
		if !strings.Contains(l.text, " ") {
			continue
		}
		for _, rt := range retiredTerms {
			if rt.re.MatchString(l.text) {
				t.Errorf("%s: retired term in %q; use %s", l.pos, l.text, rt.use)
			}
		}
	}
}

// rawAmountRe finds a bare float verb followed by a resource slot, the shape
// of an amount printed without the shared formatter ("%.0f %s").
var rawAmountRe = regexp.MustCompile(`%[+]?\.[0-9]f %s`)

// TestCopyOneNumberFormatter fails when a format string prints an amount with
// a raw float verb instead of textfmt.Number / game.Amount / FormatNumber.
func TestCopyOneNumberFormatter(t *testing.T) {
	for _, l := range collectCopyLiterals(t) {
		if rawAmountRe.MatchString(l.text) {
			t.Errorf("%s: raw amount format %q; use textfmt.Number or game.Amount", l.pos, l.text)
		}
	}
}

// TestCopyNoRoutineExclamations fails on an exclamation mark in player text.
// Routine events are stated, not shouted.
func TestCopyNoRoutineExclamations(t *testing.T) {
	bang := regexp.MustCompile(`[A-Za-z0-9)\]]!(\s|$|\[|")`)
	for _, l := range collectCopyLiterals(t) {
		if bang.MatchString(l.text) {
			t.Errorf("%s: exclamation mark in %q", l.pos, l.text)
		}
	}
}
