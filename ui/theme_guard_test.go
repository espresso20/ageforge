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

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/theme"
)

// TestNoRawColorsOutsideTheme is the "no raw colors outside theme/" guard
// (the theming design §3.8). It parses every non-test Go file under ui/, game/ and
// config/ and fails on:
//
//   - a tview inline color tag in a string literal whose fg or bg is a hex
//     literal or a tcell color name the theme does not own ([aqua], [black:gold],
//     [white:#30363d]) — theme-owned names are the role vocabulary ([accent],
//     [onaccent:accent]) plus the legacy aliases ([gold], [green], …);
//   - a named tcell color constant (tcell.ColorGold …) — use theme.Color(role);
//   - tcell.NewRGBColor / tcell.NewHexColor / tcell.GetColor — derive the
//     color from theme roles.
func TestNoRawColorsOutsideTheme(t *testing.T) {
	owned := theme.TagNames()

	// A tag is [fg], [fg:bg] or [fg:bg:attrs]; fg/bg may be empty or "-".
	tagRe := regexp.MustCompile(`\[([a-zA-Z#][a-zA-Z0-9#]*|-)?(?::([a-zA-Z#][a-zA-Z0-9#]*|-)?)?(?::[a-zA-Z-]*)?\]`)
	isCSS := func(name string) bool {
		_, ok := tcell.ColorNames[strings.ToLower(name)]
		return ok
	}
	badName := func(n string) bool {
		if n == "" || n == "-" {
			return false
		}
		if strings.HasPrefix(n, "#") {
			return true
		}
		if _, ok := owned[n]; ok {
			return false
		}
		return isCSS(n)
	}

	root := ".."
	fset := token.NewFileSet()
	for _, dir := range []string{"ui", "game", "config"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			f, perr := parser.ParseFile(fset, path, nil, 0)
			if perr != nil {
				return perr
			}
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.BasicLit:
					if x.Kind != token.STRING {
						return true
					}
					s, uerr := strconv.Unquote(x.Value)
					if uerr != nil {
						return true
					}
					for _, m := range tagRe.FindAllStringSubmatch(s, -1) {
						if badName(m[1]) || badName(m[2]) {
							t.Errorf("%s: raw color tag %q — use a theme role tag (theme.Tag / [accent] …)",
								fset.Position(x.Pos()), m[0])
						}
					}
				case *ast.SelectorExpr:
					pkg, ok := x.X.(*ast.Ident)
					if !ok || pkg.Name != "tcell" {
						return true
					}
					name := x.Sel.Name
					switch {
					case name == "ColorDefault" || name == "ColorReset" || name == "ColorNames":
					case strings.HasPrefix(name, "Color") && len(name) > 5 && name[5] >= 'A' && name[5] <= 'Z':
						t.Errorf("%s: tcell.%s — use theme.Color(role)", fset.Position(x.Pos()), name)
					case name == "NewRGBColor" || name == "NewHexColor" || name == "GetColor":
						t.Errorf("%s: tcell.%s — derive from theme roles", fset.Position(x.Pos()), name)
					}
				}
				return true
			})
			return nil
		})
		must(t, err)
	}
}
