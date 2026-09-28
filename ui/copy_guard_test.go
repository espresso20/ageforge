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
//   - the Army PR owns the military/catastrophe files and rewrites their text
//     there. note: drop those four from this list once that PR lands.
var copyExemptFiles = map[string]string{
	"game/devmode.go":           "dev console",
	"game/devcmd.go":            "dev console",
	"game/military.go":          "owned by the Army PR",
	"game/catastrophe.go":       "owned by the Army PR",
	"ui/catastrophe_modal.go":   "owned by the Army PR",
	"ui/overlay_military.go":    "owned by the Army PR",
	"ui/citymap/debug.go":       "debug overlay",
	"game/updater.go":           "self-update plumbing, not game text",
	"ui/splash_canvas.go":       "art",
	"ui/braille.go":             "chart glyphs",
	"ui/wonder_icon.go":         "art",
	"game/savenames.go":         "generated save names",
	"ui/theme_widgets.go":       "widget chrome",
	"ui/copy_guard_test.go":     "this file",
	"game/expedition_flavor.go": "flavor catalog glue",
}

// copyExemptLiterals are literal substrings owned elsewhere (raid resolution
// in game/diplomacy.go belongs to the Army PR).
var copyExemptLiterals = []string{
	"raided you",
}

type copyLit struct {
	pos  string
	text string
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
			ast.Inspect(f, func(n ast.Node) bool {
				// Skip struct tags and import paths.
				switch x := n.(type) {
				case *ast.ImportSpec:
					return false
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
					out = append(out, copyLit{pos: rel + ":" + strconv.Itoa(p.Line), text: s})
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

// bracketWordRe finds [word] spans that tview could read as a color tag.
var bracketWordRe = regexp.MustCompile(`\[([A-Za-z][A-Za-z0-9#:\-]*)\]`)

// TestCopyNoEatenBrackets fails when a literal carries a [word] that tview
// would swallow as a color tag ("[count]", "[allied]", "[current]"). Real
// color tags ([gold], [-], [accent:chip:b]) pass. Write the brackets
// escaped ("[count[]") or pass the text through lit().
func TestCopyNoEatenBrackets(t *testing.T) {
	for _, l := range collectCopyLiterals(t) {
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
	{regexp.MustCompile(`(?i)\bESC to close\b`), "Esc to close"},
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
