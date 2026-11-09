package window

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// gameRunes is every non-ASCII printable rune the game can print: the string
// and rune literals of its Go source, and the data files it loads. A rune the
// window has no renderer for would show as "?", so TestEveryRuneHasARenderer
// fails when a new one appears.
func gameRunes(t *testing.T) map[rune]string {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	out := map[rune]string{}
	note := func(s, where string) {
		for _, r := range s {
			if r >= 0x80 && unicode.IsPrint(r) {
				if _, ok := out[r]; !ok {
					out[r] = where
				}
			}
		}
	}
	fs := token.NewFileSet()
	skip := map[string]bool{"window": true, "site": true, "smoke": true, "docs": true, "scripts": true, ".git": true, "design-private": true, "data": true, "reports": true, "research_notes": true, "smoke-report": true}
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		if d.IsDir() {
			if rel != "." && skip[strings.Split(rel, string(filepath.Separator))[0]] {
				return filepath.SkipDir
			}
			return nil
		}
		switch {
		case strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go"):
			f, err := parser.ParseFile(fs, p, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				if bl, ok := n.(*ast.BasicLit); ok && (bl.Kind == token.STRING || bl.Kind == token.CHAR) {
					if s, err := strconv.Unquote(bl.Value); err == nil {
						note(s, rel+":"+strconv.Itoa(fs.Position(bl.Pos()).Line))
					}
				}
				return true
			})
		case strings.HasSuffix(p, ".csv"), strings.HasSuffix(p, ".json"), strings.HasSuffix(p, ".txt") && (strings.HasPrefix(rel, "flavor") || strings.HasPrefix(rel, "game")):
			b, _ := os.ReadFile(p)
			note(string(b), rel)
		}
		return nil
	})
	return out
}

func TestEveryRuneHasARenderer(t *testing.T) {
	fc := newFaces()
	var bad []string
	// Every other chain starts with a font and ends in the same fallbacks, so
	// the smooth font's chain is the one that has to cover everything.
	for r, where := range gameRunes(t) {
		if len(fc.missing(FontMono, []rune{r})) > 0 {
			bad = append(bad, fmt.Sprintf("%c U+%04X (%s)", r, r, where))
		}
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		t.Errorf("%d runes have no font and no drawn shape (add them to the subsets in fonts/README.md):\n%s", len(bad), strings.Join(bad, "\n"))
	}
}
