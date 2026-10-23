package main

import (
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// widths.go: the longest a label of the screens may be, from the width the
// code lays it out in. A VibeMode line can be longer than the plain one, and
// where the code gives a label a fixed width it does not make room: the value
// after it moves, or a table's column spills. So the export carries the width
// as the row's max, and vibe-lint holds a line to it.
//
// The width comes from the code, one of three ways:
//   - the label is padded by the code to a width (padRight(label, 12)), or by
//     its own trailing spaces ("Odds   " before a value in a column);
//   - the label is a table entry limited by a constant of the table
//     (panelInfo.job by panelJobMax);
//   - the label is the argument of a format that sets its field to a width
//     (Fprintf(w, " %-20s ...", "Current epoch:")).

// padCalls are the functions that pad a text to the width of their second argument.
var padCalls = map[string]bool{"padRight": true, "padLeft": true}

// widthConsts name the constant that limits a field of a table, by "package.Type.field".
var widthConsts = map[string]string{
	"ui.panelInfo.job": "panelJobMax",
}

// padded is a label with spaces after it, to the column its value is in: words
// only, so a hint or a sentence with spaces in it is not taken for one.
var padded = regexp.MustCompile(`^[A-Za-z][A-Za-z' ]*[A-Za-z] {2,}$`)

// fieldVerb is a format verb with a width: %-20s, %14s.
var fieldVerb = regexp.MustCompile(`%[-+# 0]*(\d+)?(?:\.\d+)?[a-zA-Z%]`)

// fixedWidths gives every row of the screens sheet that is laid out in a fixed
// width its max.
func (c *catalog) fixedWidths() {
	ints := map[*srcFile]map[string]int{}
	for _, r := range c.rows {
		if r.max > 0 {
			continue
		}
		switch r.area {
		case areaScreens:
			r.max = c.widthOf(r, ints)
		case areaMap:
			r.max = mapWidthOf(r)
		}
	}
}

// widthOf is the width the code gives the row's text, or 0 when it gives none.
func (c *catalog) widthOf(r *row, ints map[*srcFile]map[string]int) int {
	u := r.u
	if n := c.padWidth(u, ints); n > 0 {
		return n
	}
	if n := c.verbWidth(u); n > 0 {
		return n
	}
	if len(u.structs) > 0 {
		st := u.structs[0]
		if name, ok := widthConsts[u.f.pkg+"."+strings.TrimPrefix(st.typ, u.f.pkg+".")+"."+st.field]; ok {
			if n := c.intConst(u.f, name, ints); n > 0 {
				return n
			}
		}
	}
	if padded.MatchString(r.current) {
		return utf8.RuneCountInString(r.current) // a label padded to its column by its own spaces
	}
	return 0
}

// verbWidth is the width of the field a text is the argument of, when the
// format of the call gives its verb one (%-20s) and the text is the whole
// argument: a label set in a column.
func (c *catalog) verbWidth(u *unit) int {
	n := 0
	ast.Inspect(u.f.ast, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || n > 0 {
			return n == 0
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		format := -1
		switch pkg.Name + "." + sel.Sel.Name {
		case "fmt.Sprintf", "fmt.Errorf":
			format = 0
		case "fmt.Fprintf", "vibe.Sprintf", "vibe.Errorf":
			format = 1
		}
		if format < 0 || format >= len(call.Args) {
			return true
		}
		spec, ok := stringOf(call.Args[format])
		if !ok {
			return true
		}
		for i := format + 1; i < len(call.Args); i++ {
			arg := call.Args[i]
			lo, hi := c.m.fset.Position(arg.Pos()).Offset, c.m.fset.Position(arg.End()).Offset
			if u.start < lo || u.end > hi {
				continue
			}
			if !whollyTheText(arg, lo, hi, u) {
				return true
			}
			var verbs [][]string
			for _, v := range fieldVerb.FindAllStringSubmatch(spec, -1) {
				if !strings.HasSuffix(v[0], "%") { // %% takes no argument
					verbs = append(verbs, v)
				}
			}
			k := i - format - 1
			if k < len(verbs) && strings.HasSuffix(verbs[k][0], "s") && verbs[k][1] != "" {
				n, _ = strconv.Atoi(verbs[k][1])
			}
		}
		return true
	})
	return n
}

// whollyTheText says whether an argument is the unit's text and nothing else:
// the literal, or a vibe call that carries it.
func whollyTheText(arg ast.Expr, lo, hi int, u *unit) bool {
	if lo == u.start && hi == u.end {
		return true
	}
	if call, ok := arg.(*ast.CallExpr); ok && isVibeCall(call) && len(call.Args) == 2 {
		return true
	}
	return false
}

// padWidth is the width a text is padded to when it is the first argument of
// padRight or padLeft with a width the code writes as a number or a constant.
func (c *catalog) padWidth(u *unit, ints map[*srcFile]map[string]int) int {
	if !u.inCall("padRight", "padLeft") {
		return 0
	}
	n := 0
	ast.Inspect(u.f.ast, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || n > 0 || len(call.Args) != 2 {
			return n == 0
		}
		id, ok := call.Fun.(*ast.Ident)
		if !ok || !padCalls[id.Name] {
			return true
		}
		lo, hi := c.m.fset.Position(call.Args[0].Pos()).Offset, c.m.fset.Position(call.Args[0].End()).Offset
		if u.start < lo || u.end > hi {
			return true
		}
		switch w := call.Args[1].(type) {
		case *ast.BasicLit:
			if w.Kind == token.INT {
				n, _ = strconv.Atoi(w.Value)
			}
		case *ast.Ident:
			n = c.intConst(u.f, w.Name, ints)
		}
		return true
	})
	return n
}

// intConst is the value of an integer constant of the file's package.
func (c *catalog) intConst(f *srcFile, name string, ints map[*srcFile]map[string]int) int {
	for _, g := range c.m.files {
		if g.pkg != f.pkg || g.dir != f.dir {
			continue
		}
		if _, ok := ints[g]; !ok {
			ints[g] = map[string]int{}
			for _, d := range g.ast.Decls {
				gd, ok := d.(*ast.GenDecl)
				if !ok || gd.Tok != token.CONST {
					continue
				}
				for _, sp := range gd.Specs {
					vs := sp.(*ast.ValueSpec)
					for i, n := range vs.Names {
						if i < len(vs.Values) {
							if lit, ok := vs.Values[i].(*ast.BasicLit); ok && lit.Kind == token.INT {
								if v, err := strconv.Atoi(lit.Value); err == nil {
									ints[g][n.Name] = v
								}
							}
						}
					}
				}
			}
		}
		if v, ok := ints[g][name]; ok {
			return v
		}
	}
	return 0
}
