package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// This file loads the game's Go source for the export and the import: every
// non-test file of the packages that hold player text, parsed once, with the
// tables the reader needs to say what a string literal is part of (which
// struct type a composite literal builds, what a named constant holds).

// scanned lists the packages whose string literals the export reads, with
// the reader each one gets. The order is the order of the scan.
var scanned = []struct {
	dir  string
	mode scanMode
}{
	{"flavor", modeCode},
	{"config", modeData},
	{"boon", modeData},
	{"theme", modeData},
	{"mapmodel", modeData},
	{"rules", modeCode},
	{"pkg/textfmt", modeCode},
	{"game", modeCode},
	{"presence", modeCode},
	{"ui", modeCode},
	{"ui/mapstyle", modeCode},
	{".", modeCode},
}

// notScanned are the module's other packages, each with why no player reads
// its strings. TestEveryPackageIsAccountedFor fails on a package in neither
// list, so a new package cannot go unread by accident.
var notScanned = map[string]string{
	"cmd/words":             "this tool",
	"cmd/bannerpng":         "README banner generator, no player text",
	"smoke":                 "the smoke suite",
	"vibe":                  "the VibeMode switch, no text (its script is data, in words/vibe)",
	"detmath":               "arithmetic, no text",
	"mapgen":                "the map engine: art, hues and internal keys, no player text (the ui reads the game's own words)",
	"mapmodel/fixture":      "test fixtures",
	"pkg/nerdfont":          "the icon font installer's glyph table and system paths",
	"pkg/sprites":           "pixel art",
	"pkg/wordlist":          "reads the plain-text word lists, no text of its own",
	"ui/mapstyle/capture":   "a developer capture tool, never shown in the game",
	"ui/mapstyle/styletest": "checks shared by the map tests, never shown in the game",
}

type scanMode int

const (
	modeCode scanMode = iota // ordinary code: text is whatever reaches the screen
	modeData                 // data definitions: text is the fields the tables say
)

// srcFile is one parsed source file.
type srcFile struct {
	rel     string // slash path from the module root, "config/ages.go"
	dir     string // slash path of its package, "config" ("." for the root)
	pkg     string // package name
	mode    scanMode
	src     []byte
	ast     *ast.File       // nil for a plain-text file
	text    bool            // a plain-text file of player text (see textFile), not Go source
	imports map[string]bool // names its imports are known by
}

// module is the parsed source tree.
type module struct {
	root   string
	fset   *token.FileSet
	files  []*srcFile
	byRel  map[string]*srcFile
	types  map[string]map[string]ast.Expr // package name -> type name -> its definition
	consts map[string]map[string]string   // package name -> constant -> string value
}

// findRoot walks up from dir to the directory that holds the game's go.mod.
func findRoot(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if b, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil &&
			strings.Contains(string(b), "module github.com/espresso20/ageforge") {
			return dir, nil
		}
		up := filepath.Dir(dir)
		if up == dir {
			return "", fmt.Errorf("not inside the AgeForge repository (no go.mod found)")
		}
		dir = up
	}
}

// loadModule parses the scanned packages under root.
func loadModule(root string) (*module, error) {
	m := &module{
		root:   root,
		fset:   token.NewFileSet(),
		byRel:  map[string]*srcFile{},
		types:  map[string]map[string]ast.Expr{},
		consts: map[string]map[string]string{},
	}
	for _, sp := range scanned {
		names, err := goFiles(filepath.Join(root, filepath.FromSlash(sp.dir)))
		if os.IsNotExist(err) {
			continue // a partial tree, as the tests use
		}
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			rel := path.Join(sp.dir, name)
			src, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil {
				return nil, err
			}
			f, err := parser.ParseFile(m.fset, rel, src, parser.SkipObjectResolution)
			if err != nil {
				return nil, fmt.Errorf("%s does not parse: %w", rel, err)
			}
			sf := &srcFile{rel: rel, dir: sp.dir, pkg: f.Name.Name, mode: sp.mode, src: src, ast: f, imports: map[string]bool{}}
			for _, im := range f.Imports {
				p, _ := strconv.Unquote(im.Path.Value)
				name := path.Base(p)
				if im.Name != nil {
					name = im.Name.Name
				}
				sf.imports[name] = true
			}
			m.files = append(m.files, sf)
			m.byRel[rel] = sf
			m.index(sf)
		}
	}
	return m, nil
}

// textFile loads a plain-text file that holds player text, once: a flavor
// catalog or the save name words. It is not one of the module's Go files, so
// the readers never see it as source; its rows point into it by byte position,
// as the rows of a Go file do, and the import writes a rewrite back there.
func (m *module) textFile(rel string) (*srcFile, error) {
	if f, ok := m.byRel[rel]; ok {
		return f, nil
	}
	src, err := os.ReadFile(filepath.Join(m.root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, err
	}
	dir := path.Dir(rel)
	f := &srcFile{rel: rel, dir: dir, pkg: path.Base(dir), src: src, text: true}
	m.byRel[rel] = f
	return f, nil
}

// textUnit is the text at bytes [from, to) of a plain-text file.
func textUnit(f *srcFile, line, from, to int) *unit {
	return &unit{f: f, start: from, end: to, line: line, text: string(f.src[from:to])}
}

// goFiles lists a directory's non-test Go files, sorted.
func goFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		out = append(out, n)
	}
	sort.Strings(out)
	return out, nil
}

// index records a file's type definitions and string constants.
func (m *module) index(f *srcFile) {
	if m.types[f.pkg] == nil {
		m.types[f.pkg] = map[string]ast.Expr{}
		m.consts[f.pkg] = map[string]string{}
	}
	for _, d := range f.ast.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, sp := range gd.Specs {
			switch s := sp.(type) {
			case *ast.TypeSpec:
				m.types[f.pkg][s.Name.Name] = s.Type
			case *ast.ValueSpec:
				if gd.Tok != token.CONST {
					continue
				}
				for i, n := range s.Names {
					if i < len(s.Values) {
						if v, ok := stringLit(s.Values[i]); ok {
							m.consts[f.pkg][n.Name] = v
						}
					}
				}
			}
		}
	}
}

// stringLit returns the value of a string literal, or of a chain of string
// literals joined with +.
func stringLit(e ast.Expr) (string, bool) {
	switch x := e.(type) {
	case *ast.BasicLit:
		if x.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(x.Value)
		return s, err == nil
	case *ast.ParenExpr:
		return stringLit(x.X)
	case *ast.BinaryExpr:
		if x.Op != token.ADD {
			return "", false
		}
		a, ok := stringLit(x.X)
		if !ok {
			return "", false
		}
		b, ok := stringLit(x.Y)
		return a + b, ok
	}
	return "", false
}

// typeRef is a type expression and the package whose names it is written in.
type typeRef struct {
	pkg  string
	expr ast.Expr
}

// named returns "pkg.Name" for a reference to a named type, else "".
func (m *module) named(t typeRef) string {
	switch x := t.expr.(type) {
	case *ast.Ident:
		return t.pkg + "." + x.Name
	case *ast.SelectorExpr:
		if p, ok := x.X.(*ast.Ident); ok {
			return p.Name + "." + x.Sel.Name
		}
	case *ast.StarExpr:
		return m.named(typeRef{t.pkg, x.X})
	case *ast.ParenExpr:
		return m.named(typeRef{t.pkg, x.X})
	}
	return ""
}

// under follows named types to the definition underneath. It gives up after
// 16 steps.
func (m *module) under(t typeRef) typeRef {
	for i := 0; i < 16 && t.expr != nil; i++ {
		switch x := t.expr.(type) {
		case *ast.StarExpr:
			t = typeRef{t.pkg, x.X}
		case *ast.ParenExpr:
			t = typeRef{t.pkg, x.X}
		case *ast.Ident:
			def, ok := m.types[t.pkg][x.Name]
			if !ok {
				return t
			}
			t = typeRef{t.pkg, def}
		case *ast.SelectorExpr:
			p, ok := x.X.(*ast.Ident)
			if !ok {
				return t
			}
			def, ok := m.types[p.Name][x.Sel.Name]
			if !ok {
				return t
			}
			t = typeRef{p.Name, def}
		default:
			return t
		}
	}
	return t
}

// structField is one field of a struct type, in declaration order.
type structField struct {
	name string
	typ  ast.Expr
}

// fields lists a struct type's fields and the package that defines it, or nil
// when t is not a struct.
func (m *module) fields(t typeRef) ([]structField, string) {
	u := m.under(t)
	st, ok := u.expr.(*ast.StructType)
	if !ok {
		return nil, ""
	}
	var out []structField
	for _, f := range st.Fields.List {
		for _, n := range f.Names {
			out = append(out, structField{n.Name, f.Type})
		}
	}
	return out, u.pkg
}
