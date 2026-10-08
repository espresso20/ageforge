package main

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// units.go finds every piece of text in a source file and records where it
// sits in the syntax tree: the function or declaration around it, the calls
// it is an argument of, the struct literal and field it fills, its place in a
// list or map. The readers (flavor, data, code) decide from that whether a
// player sees it and what to call it.

// unit is one piece of text: a string literal, or several joined with +.
type unit struct {
	f          *srcFile
	start, end int // byte offsets of the expression in f.src
	line       int
	raw        bool // written between backquotes
	text       string

	fn       string       // the function it is in ("" at package level)
	recv     string       // that function's receiver type, if it is a method
	decl     string       // the package-level name it is under: the function, or the var or const
	near     string       // the closest syntax around it: call, field, elem, return, assign, var, compare, case, index, mapkey, other
	assignTo string       // what an assignment or declaration gives it to
	calls    []callSite   // the calls around it, innermost first
	structs  []structSite // the struct literals around it, innermost first
	coll     []collStep   // its place in lists and maps, out to the innermost struct field
	chain    *chain       // set when it is joined with + to things that are not text
	caseKey  string       // the first value of the switch case it is under, if any
	ret      int          // its place among the values a return statement gives back
}

// callSite is a call the text is an argument of.
type callSite struct {
	pkg  string   // the package for a package function ("fmt"), else ""
	fn   string   // the function or method name
	on   string   // what a method is called on, as written ("ge", "d.engine")
	arg  int      // which argument
	out  int      // how many struct literals sit between the text and the call
	args []string // every argument that is a constant string ("" for the rest)
}

// structSite is a struct literal the text is inside.
type structSite struct {
	typ   string            // "config.BuildingDef"; "" for an unnamed struct type
	field string            // the field the text (or the list holding it) fills
	sib   map[string]string // the literal's fields that are constant strings
	lit   *ast.CompositeLit
	coll  []collStep // this literal's place in lists and maps, out to the next struct
}

// collStep is one level of list or map nesting.
type collStep struct {
	isMap bool
	key   string // map key, as text
	index int    // position in a list, from 0
}

// chain describes the + expression a literal is one part of.
type chain struct {
	pattern string // the whole expression, with "…" for what is not text
	part    int    // which text part this is, from 1
	parts   int    // how many text parts there are
}

// units returns every text unit of a file, in source order.
func (m *module) units(f *srcFile) []*unit {
	var out []*unit
	var stack []ast.Node
	litTypes := map[*ast.CompositeLit]typeRef{}

	ast.Inspect(f.ast, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		stack = append(stack, n)
		switch x := n.(type) {
		case *ast.ImportSpec:
			stack = stack[:len(stack)-1]
			return false
		case *ast.Field:
			// A struct tag is a string literal no player reads, and nothing
			// else in a field declaration is text.
			stack = stack[:len(stack)-1]
			return false
		case *ast.BinaryExpr:
			if x.Op == token.ADD {
				if s, ok := stringLit(x); ok {
					out = append(out, m.newUnit(f, stack, litTypes, x, s, false))
					stack = stack[:len(stack)-1]
					return false
				}
			}
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				if s, err := strconv.Unquote(x.Value); err == nil {
					out = append(out, m.newUnit(f, stack, litTypes, x, s, strings.HasPrefix(x.Value, "`")))
				}
			}
		}
		return true
	})
	return out
}

// newUnit builds the unit for the node on top of the stack.
func (m *module) newUnit(f *srcFile, stack []ast.Node, litTypes map[*ast.CompositeLit]typeRef, node ast.Expr, text string, raw bool) *unit {
	pos := m.fset.Position(node.Pos())
	u := &unit{
		f: f, start: pos.Offset, end: m.fset.Position(node.End()).Offset,
		line: pos.Line, raw: raw, text: text,
	}
	near := func(s string) {
		if u.near == "" {
			u.near = s
		}
	}
	var steps []collStep      // list and map levels passed since the last struct
	var pendingKey ast.Expr   // the key of the key: value pair just passed
	var pendingKV bool        // whether a key: value pair was just passed
	child := ast.Node(node)   // the node below the one being looked at
	inChain := ast.Expr(node) // the top of the + chain the unit is in

	for j := len(stack) - 2; j >= 0; j-- {
		parent := stack[j]
		switch p := parent.(type) {
		case *ast.ParenExpr:
			if inChain == child {
				inChain = p
			}
		case *ast.BinaryExpr:
			switch p.Op {
			case token.ADD:
				if inChain == child {
					inChain = p
				}
			case token.EQL, token.NEQ, token.LSS, token.GTR, token.LEQ, token.GEQ:
				near("compare")
			}
		case *ast.CallExpr:
			if child != p.Fun {
				near("call")
				cs := callSite{arg: -1, out: len(u.structs)}
				switch fun := p.Fun.(type) {
				case *ast.Ident:
					cs.fn = fun.Name
				case *ast.SelectorExpr:
					cs.fn = fun.Sel.Name
					if id, ok := fun.X.(*ast.Ident); ok && f.imports[id.Name] {
						cs.pkg = id.Name
					} else {
						cs.on = exprText(fun.X)
					}
				case *ast.ArrayType:
					cs.fn = "[]" + exprText(fun.Elt)
				default:
					cs.fn = exprText(p.Fun)
				}
				for i, a := range p.Args {
					if a == child {
						cs.arg = i
					}
					cs.args = append(cs.args, m.constString(f.pkg, a))
				}
				u.calls = append(u.calls, cs)
			}
		case *ast.KeyValueExpr:
			if child == p.Key {
				near("mapkey")
			} else {
				pendingKey, pendingKV = p.Key, true
			}
		case *ast.CompositeLit:
			t := m.litType(f, stack, j, litTypes)
			if fs, fpkg := m.fields(t); fs != nil {
				near("field")
				site := structSite{typ: m.named(t), lit: p, sib: map[string]string{}}
				for i, el := range p.Elts {
					name, val := "", el
					if kv, ok := el.(*ast.KeyValueExpr); ok {
						if id, ok := kv.Key.(*ast.Ident); ok {
							name = id.Name
						}
						val = kv.Value
					} else if i < len(fs) {
						name = fs[i].name
					}
					if name == "" {
						continue
					}
					if s := m.constString(fpkg, val); s != "" {
						site.sib[name] = s
					} else if s := m.constString(f.pkg, val); s != "" {
						site.sib[name] = s
					}
					if el == child || (pendingKV && val == child) {
						site.field = name
					}
				}
				if len(u.structs) == 0 {
					u.coll = steps
				} else {
					u.structs[len(u.structs)-1].coll = steps
				}
				steps = nil
				u.structs = append(u.structs, site)
			} else {
				near("elem")
				step := collStep{}
				if _, isMap := m.under(t).expr.(*ast.MapType); isMap || pendingKV {
					step.isMap = true
					step.key = m.keyText(f.pkg, pendingKey)
				} else {
					for i, el := range p.Elts {
						if el == child {
							step.index = i
						}
					}
				}
				steps = append(steps, step)
			}
			pendingKey, pendingKV = nil, false
		case *ast.CaseClause:
			inList := false
			for _, e := range p.List {
				if e == child {
					near("case")
					inList = true
				}
			}
			if !inList && u.caseKey == "" {
				u.caseKey = "default"
				if len(p.List) > 0 {
					u.caseKey = m.keyText(f.pkg, p.List[0])
				}
			}
		case *ast.IndexExpr:
			if child == p.Index {
				near("index")
			}
		case *ast.ReturnStmt:
			if u.near == "" {
				for i, e := range p.Results {
					if e == child {
						u.ret = i
					}
				}
			}
			near("return")
		case *ast.AssignStmt:
			near("assign")
			if u.assignTo == "" {
				for i, r := range p.Rhs {
					if r == child && i < len(p.Lhs) {
						u.assignTo = exprText(p.Lhs[i])
					}
				}
				if u.assignTo == "" && len(p.Lhs) > 0 {
					u.assignTo = exprText(p.Lhs[0])
				}
			}
		case *ast.ValueSpec:
			near("var")
			if u.assignTo == "" {
				for i, v := range p.Values {
					if v == child && i < len(p.Names) {
						u.assignTo = p.Names[i].Name
					}
				}
				if u.assignTo == "" && len(p.Names) > 0 {
					u.assignTo = p.Names[0].Name
				}
			}
			if j >= 1 {
				if _, ok := stack[j-1].(*ast.GenDecl); ok && j == 2 {
					u.decl = u.assignTo
				}
			}
		case *ast.FuncDecl:
			u.fn = p.Name.Name
			u.decl = p.Name.Name
			if p.Recv != nil && len(p.Recv.List) > 0 {
				u.recv = strings.TrimPrefix(exprText(p.Recv.List[0].Type), "*")
			}
		}
		child = parent
	}
	if len(u.structs) == 0 {
		u.coll = steps
	} else {
		u.structs[len(u.structs)-1].coll = steps
	}
	if u.near == "" {
		u.near = "other"
	}
	if be, ok := inChain.(*ast.BinaryExpr); ok && inChain != ast.Expr(node) {
		u.chain = chainOf(be, node)
	}
	return u
}

// chainOf describes the + expression top from the point of view of the
// literal at.
func chainOf(top ast.Expr, at ast.Expr) *chain {
	var leaves []ast.Expr
	var flatten func(e ast.Expr)
	flatten = func(e ast.Expr) {
		switch x := e.(type) {
		case *ast.ParenExpr:
			flatten(x.X)
		case *ast.BinaryExpr:
			if x.Op == token.ADD && x != at {
				flatten(x.X)
				flatten(x.Y)
				return
			}
			leaves = append(leaves, e)
		default:
			leaves = append(leaves, e)
		}
	}
	flatten(top)
	c := &chain{}
	var sb strings.Builder
	gap := false
	for _, l := range leaves {
		if s, ok := stringLit(l); ok {
			c.parts++
			if l == at {
				c.part = c.parts
			}
			sb.WriteString(s)
			gap = false
			continue
		}
		if !gap {
			sb.WriteString("…")
			gap = true
		}
	}
	c.pattern = sb.String()
	return c
}

// litType works out the type a composite literal builds, from its own type
// or, when that is left out, from the list, map or struct field it sits in.
func (m *module) litType(f *srcFile, stack []ast.Node, j int, cache map[*ast.CompositeLit]typeRef) typeRef {
	lit := stack[j].(*ast.CompositeLit)
	if t, ok := cache[lit]; ok {
		return t
	}
	var t typeRef
	if lit.Type != nil {
		t = typeRef{f.pkg, lit.Type}
	} else {
		// Find the composite literal that holds this one.
		k := j - 1
		var kv *ast.KeyValueExpr
		for k >= 0 {
			switch p := stack[k].(type) {
			case *ast.KeyValueExpr:
				kv = p
				k--
				continue
			case *ast.UnaryExpr, *ast.ParenExpr:
				k--
				continue
			}
			break
		}
		if k >= 0 {
			if outer, ok := stack[k].(*ast.CompositeLit); ok {
				ot := m.litType(f, stack, k, cache)
				u := m.under(ot)
				switch x := u.expr.(type) {
				case *ast.ArrayType:
					t = typeRef{u.pkg, x.Elt}
				case *ast.MapType:
					if kv != nil && kv.Key == ast.Expr(lit) {
						t = typeRef{u.pkg, x.Key}
					} else {
						t = typeRef{u.pkg, x.Value}
					}
				case *ast.StructType:
					fs, fpkg := m.fields(ot)
					if kv != nil {
						if id, ok := kv.Key.(*ast.Ident); ok {
							for _, fd := range fs {
								if fd.name == id.Name {
									t = typeRef{fpkg, fd.typ}
								}
							}
						}
					} else {
						for i, el := range outer.Elts {
							if el == ast.Expr(lit) && i < len(fs) {
								t = typeRef{fpkg, fs[i].typ}
							}
						}
					}
				}
			}
		}
	}
	cache[lit] = t
	return t
}

// constString returns the string an expression always holds: a literal, a
// chain of literals, or a named string constant. "" when it is anything else.
func (m *module) constString(pkg string, e ast.Expr) string {
	if s, ok := stringLit(e); ok {
		return s
	}
	switch x := e.(type) {
	case *ast.Ident:
		return m.consts[pkg][x.Name]
	case *ast.SelectorExpr:
		if p, ok := x.X.(*ast.Ident); ok {
			return m.consts[p.Name][x.Sel.Name]
		}
	}
	return ""
}

// keyText renders a map key for an id: its string value, or its name.
func (m *module) keyText(pkg string, e ast.Expr) string {
	if e == nil {
		return ""
	}
	if s := m.constString(pkg, e); s != "" {
		return s
	}
	return exprText(e)
}

// exprText renders a simple expression the way it is written.
func exprText(e ast.Expr) string {
	switch x := e.(type) {
	case nil:
		return ""
	case *ast.Ident:
		return x.Name
	case *ast.BasicLit:
		return x.Value
	case *ast.SelectorExpr:
		return exprText(x.X) + "." + x.Sel.Name
	case *ast.StarExpr:
		return "*" + exprText(x.X)
	case *ast.ParenExpr:
		return exprText(x.X)
	case *ast.IndexExpr:
		return exprText(x.X) + "[" + exprText(x.Index) + "]"
	case *ast.CallExpr:
		return exprText(x.Fun) + "()"
	case *ast.ArrayType:
		return "[]" + exprText(x.Elt)
	case *ast.UnaryExpr:
		return x.Op.String() + exprText(x.X)
	}
	return "?"
}

// call returns the innermost call around the unit, or a zero callSite.
func (u *unit) call() callSite {
	if len(u.calls) > 0 {
		return u.calls[0]
	}
	return callSite{}
}

// inCall reports whether any call around the unit is one of the named
// functions ("fn" or "pkg.fn").
func (u *unit) inCall(names ...string) bool {
	for _, c := range u.calls {
		for _, n := range names {
			if n == c.fn || (c.pkg != "" && n == c.pkg+"."+c.fn) {
				return true
			}
		}
	}
	return false
}

// st returns the innermost struct literal around the unit, or a zero site.
func (u *unit) st() structSite {
	if len(u.structs) > 0 {
		return u.structs[0]
	}
	return structSite{}
}
