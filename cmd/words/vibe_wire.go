package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// vibe_wire.go: vibe-wire, which puts a text of the code under VibeMode.
//
// A text of the code is wired when it is written as a vibe call that carries
// its row id:
//
//	vibe.Line("game.engine.Reset.01", "Game wiped. Starting fresh.")
//	vibe.Sprintf("game.engine.finishBuild.02", "%s built (you have %d).", name, n)
//	vibe.Errorf("game.engine.RecruitMax.02", "No housing left (%d/%d).", a, b)
//
// The call returns the string itself while VibeMode is off or the script has
// no line for the id, and the script's line while it is on and has one. A
// format goes through Sprintf or Errorf, so the script's line is the format
// the arguments go into, and vet reads the call as it read fmt's.
//
// Nobody writes these by hand for a sheet of a thousand lines: vibe-wire
// reads the export, decides for each text of the sheet whether it can follow
// the switch, and rewrites the source. It decides from the syntax and from
// the lists of texts judged by hand (vibe_left.go, vibe_screens.go,
// vibe_map.go), so it can be run again after the code has moved. What it will
// not wire (a text the code compares, one that ends up in a save, a key) it
// leaves as it is, and says why. A text that cannot follow the switch as it
// is written (it is made once, at start-up) it reports and leaves for the code
// to change.

// Why a text of the code is not wired. A reason that is "for good" is not
// asked of the script's writer; the others (compared, once) are work for
// whoever changes the code around the text.
const (
	leftCovered  = "covered: its words are written as lexicon.* rows"
	leftCompared = "compared: the code matches the text it is given against it"
	leftState    = "state: it ends up in the state or a save, which never holds a script's text"
	leftKey      = "key: a key, a file name, a header or a format the code reads back, not words"
	leftDev      = "developers: a line of the debug log, a dev command or an audit, which a player does not read"
	leftUnits    = "units: the suffix of a number (Qi, Oc, No), which reads the same in any script"
	leftGlue     = "glue: a joining word the code also splits a name by or looks for in it, put back as it was found"
	leftSearch   = "search: a word the code looks for in a line written in plain English (a line under a script is read by its tags)"
	leftOnce     = "once: made when the game starts or in a constant, so it cannot follow the switch"
)

// forGood says whether a reason is final. A text the code compares, or makes
// once, can be wired once the code is changed to match on a key or to make
// the text when it is wanted; the others never can.
func forGood(reason string) bool {
	return reason != "" && reason != leftOnce && reason != leftCompared
}

// leftAlone are the texts a person has judged are not to follow the switch,
// by row id, with the reason. The rest are decided from the syntax.
var leftAlone = map[string]string{}

// leftAloneFor lists the ids a reason holds for.
func leftAloneFor(reason string, ids ...string) {
	for _, id := range ids {
		leftAlone[id] = reason
	}
}

// wireKind is the vibe call a text is written as.
type wireKind int

const (
	wireNone    wireKind = iota
	wireLine             // vibe.Line(id, "text")
	wireSprintf          // fmt.Sprintf("format", ...) becomes vibe.Sprintf(id, "format", ...)
	wireErrorf           // fmt.Errorf likewise
	wireEntry            // a text of a table made once (a package-level var) becomes vibe.E(id, "text")
	wireFprintf          // fmt.Fprintf(w, "format", ...) becomes fmt.Fprint(w, vibe.Sprintf(id, "format", ...))
	wireSame             // vibe.Same("text"), the entry of a string that had no row, becomes vibe.E(id, "text")
)

// verdict is what vibe-wire makes of a row.
type verdict struct {
	kind   wireKind
	reason string // when kind is wireNone: why
	call   *ast.CallExpr
}

// nodeCtx is what the syntax around a text says about wiring it.
type nodeCtx struct {
	inFunc  bool          // inside a function body (or a literal of one), so it runs when the game does
	inConst bool          // inside a const declaration
	inVar   bool          // inside a package-level var declaration
	call    *ast.CallExpr // the call the text is the format of (fmt.Sprintf, Errorf or Fprintf), or the vibe.Same it is the argument of
	kind    wireKind
}

// nodeCtxs reads the context of every text unit of a file, by its place.
func (m *module) nodeCtxs(f *srcFile) map[[2]int]nodeCtx {
	out := map[[2]int]nodeCtx{}
	var stack []ast.Node
	ast.Inspect(f.ast, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		stack = append(stack, n)
		record := func(e ast.Expr) {
			ctx := nodeCtx{}
			for _, a := range stack[:len(stack)-1] {
				switch x := a.(type) {
				case *ast.FuncDecl:
					ctx.inFunc = x.Body != nil
				case *ast.FuncLit:
					ctx.inFunc = true
				case *ast.GenDecl:
					if x.Tok == token.CONST {
						ctx.inConst = true
					}
					if x.Tok == token.VAR {
						ctx.inVar = true
					}
				}
			}
			j := len(stack) - 2
			for j >= 0 {
				if _, ok := stack[j].(*ast.ParenExpr); !ok {
					break
				}
				j--
			}
			if j >= 0 {
				if call, ok := stack[j].(*ast.CallExpr); ok && len(call.Args) > 0 {
					if isVibeSame(call) && call.Args[0] == e {
						ctx.call, ctx.kind = call, wireSame
					}
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if x, ok := sel.X.(*ast.Ident); ok && x.Name == "fmt" && f.imports["fmt"] {
							switch {
							case call.Args[0] == e && sel.Sel.Name == "Sprintf":
								ctx.call, ctx.kind = call, wireSprintf
							case call.Args[0] == e && sel.Sel.Name == "Errorf":
								ctx.call, ctx.kind = call, wireErrorf
							case len(call.Args) > 1 && call.Args[1] == e && sel.Sel.Name == "Fprintf":
								ctx.call, ctx.kind = call, wireFprintf
							}
						}
					}
				}
			}
			out[[2]int{m.fset.Position(e.Pos()).Offset, m.fset.Position(e.End()).Offset}] = ctx
		}
		switch x := n.(type) {
		case *ast.ImportSpec, *ast.Field:
			stack = stack[:len(stack)-1]
			return false
		case *ast.BinaryExpr:
			if x.Op == token.ADD {
				if _, ok := stringLit(x); ok {
					record(x)
					stack = stack[:len(stack)-1]
					return false
				}
			}
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				record(x)
			}
		}
		return true
	})
	return out
}

// wireVerdict decides how the text of a row follows the switch, or why it
// does not.
func (c *catalog) wireVerdict(r *row) verdict {
	if flavorRead(r) {
		// A sentence or a word of a flavor catalog is told through the script by
		// the flavor package when it draws a line, under the id its place in the
		// catalog gives it. There is nothing in the source to rewrite.
		return verdict{}
	}
	if reason, ok := leftAlone[r.id]; ok {
		return verdict{reason: reason}
	}
	if reason := screenReason(r); reason != "" {
		return verdict{reason: reason}
	}
	if strings.HasPrefix(r.id, "config.lexicon.lexiconRows.") || strings.HasPrefix(r.id, "config.lexicon.plainTerms.") {
		return verdict{reason: leftCovered}
	}
	if len(r.twins) > 0 && !wireDespiteTwins[r.id] {
		return verdict{reason: leftCompared}
	}
	for _, cs := range r.u.calls {
		if logCalls[cs.fn] && cs.pkg == "" && len(cs.args) > 0 && cs.args[0] == "debug" {
			return verdict{reason: leftDev}
		}
	}
	if c.ctxs == nil {
		c.ctxs = map[*srcFile]map[[2]int]nodeCtx{}
	}
	idx, ok := c.ctxs[r.u.f]
	if !ok {
		idx = c.m.nodeCtxs(r.u.f)
		c.ctxs[r.u.f] = idx
	}
	ctx, ok := idx[[2]int{r.u.start, r.u.end}]
	if ok && ctx.kind == wireSame {
		// A vibe.Same entry whose string has a row of the export: it becomes
		// vibe.E.
		return verdict{kind: wireSame, call: ctx.call}
	}
	if ok && (r.area == areaScreens || r.area == areaMap) && !ctx.inFunc && !ctx.inConst && ctx.inVar && tableEntry(r.u.near) {
		// A table made once: its entries carry their ids and read when used.
		return verdict{kind: wireEntry}
	}
	if !ok || !ctx.inFunc || ctx.inConst {
		return verdict{reason: leftOnce}
	}
	if ctx.kind != wireNone {
		return verdict{kind: ctx.kind, call: ctx.call}
	}
	return verdict{kind: wireLine}
}

// flavorRead says whether a row is a sentence or a word-list entry of a flavor
// catalog, which the flavor package tells through the script itself.
func flavorRead(r *row) bool { return r.rules&(rulesSkel|rulesBank) != 0 }

// edit replaces the bytes [lo, hi) of a file.
type edit struct {
	lo, hi int
	text   string
}

// vibeWire writes the vibe calls for the rows of one sheet that can have
// them and are not wired yet. With dry set it only reports. With only set, it
// takes just the rows whose id starts with it ("ui.input."), so a big sheet can
// be wired a panel at a time.
func vibeWire(c *catalog, root, sheet, only string, dry bool, out io.Writer) error {
	var target *area
	for _, a := range areas {
		if a.file == sheet {
			target = a
		}
	}
	if target == nil {
		return fmt.Errorf("no sheet is called %q", sheet)
	}
	perFile := map[*srcFile][]edit{}
	var files []*srcFile
	count := map[wireKind]int{}
	left := map[string][]string{}
	already := 0
	for _, r := range c.rows {
		if r.area != target {
			continue
		}
		if r.u.vibeID != "" || flavorRead(r) {
			already++
			continue
		}
		if !strings.HasPrefix(r.id, only) {
			continue
		}
		v := c.wireVerdict(r)
		if v.kind == wireNone {
			left[v.reason] = append(left[v.reason], r.id)
			continue
		}
		count[v.kind]++
		f := r.u.f
		if _, seen := perFile[f]; !seen {
			files = append(files, f)
			perFile[f] = nil
		}
		id := strconv.Quote(r.id)
		switch v.kind {
		case wireSame:
			// vibe.Same("text") becomes vibe.E("id", "text"): the call's name is
			// replaced and the id goes in front of the text.
			sel := v.call.Fun.(*ast.SelectorExpr)
			lo := c.m.fset.Position(sel.Sel.Pos()).Offset
			hi := c.m.fset.Position(sel.Sel.End()).Offset
			perFile[f] = append(perFile[f],
				edit{lo, hi, "E"},
				edit{r.u.start, r.u.start, id + ", "})
		case wireLine, wireEntry:
			name := "vibe.Line("
			if v.kind == wireEntry {
				name = "vibe.E("
			}
			perFile[f] = append(perFile[f],
				edit{r.u.start, r.u.start, name + id + ", "},
				edit{r.u.end, r.u.end, ")"})
		case wireFprintf:
			// The writer keeps its place; the format becomes the vibe call's and the
			// whole is written with Fprint, so vet reads the format and its arguments
			// at vibe.Sprintf as it read them at Fprintf.
			sel := v.call.Fun.(*ast.SelectorExpr)
			lo := c.m.fset.Position(sel.Sel.Pos()).Offset
			hi := c.m.fset.Position(sel.Sel.End()).Offset
			rp := c.m.fset.Position(v.call.Rparen).Offset
			perFile[f] = append(perFile[f],
				edit{lo, hi, "Fprint"},
				edit{r.u.start, r.u.start, "vibe.Sprintf(" + id + ", "},
				edit{rp, rp, ")"})
		case wireSprintf, wireErrorf:
			name := "vibe.Sprintf"
			if v.kind == wireErrorf {
				name = "vibe.Errorf"
			}
			lo := c.m.fset.Position(v.call.Fun.Pos()).Offset
			hi := c.m.fset.Position(v.call.Fun.End()).Offset
			perFile[f] = append(perFile[f],
				edit{lo, hi, name},
				edit{r.u.start, r.u.start, id + ", "})
		}
	}

	fmt.Fprintf(out, "%s: %d rows already wired, %d to wire now (%d vibe.Line, %d vibe.Sprintf, %d vibe.Errorf, %d vibe.E) in %d files\n",
		sheet, already, count[wireLine]+count[wireSprintf]+count[wireFprintf]+count[wireErrorf]+count[wireEntry]+count[wireSame],
		count[wireLine], count[wireSprintf]+count[wireFprintf], count[wireErrorf], count[wireEntry]+count[wireSame], len(files))
	reasons := make([]string, 0, len(left))
	for reason := range left {
		reasons = append(reasons, reason)
	}
	sort.Strings(reasons)
	for _, reason := range reasons {
		fmt.Fprintf(out, "left: %4d  %s\n", len(left[reason]), reason)
		if dry {
			for _, id := range left[reason] {
				fmt.Fprintf(out, "        %s\n", id)
			}
		}
	}
	if dry {
		return nil
	}
	edited := map[*srcFile]bool{}
	for _, f := range files {
		edited[f] = true
		src, err := applyEdits(f.src, perFile[f])
		if err != nil {
			return fmt.Errorf("%s: %w", f.rel, err)
		}
		src, err = fixImports(f.rel, src)
		if err != nil {
			return fmt.Errorf("%s: %w", f.rel, err)
		}
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(f.rel)), src, 0o644); err != nil {
			return err
		}
	}
	// A file that already uses vibe keeps its imports tidy too.
	for _, f := range c.m.files {
		if edited[f] || !bytes.Contains(f.src, []byte("vibe.")) {
			continue
		}
		src, err := fixImports(f.rel, f.src)
		if err != nil {
			return fmt.Errorf("%s: %w", f.rel, err)
		}
		if !bytes.Equal(src, f.src) {
			if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(f.rel)), src, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(out, "tidied the imports of %s\n", f.rel)
		}
	}
	return nil
}

// applyEdits makes the edits in a source, last first so the offsets hold.
func applyEdits(src []byte, edits []edit) ([]byte, error) {
	sort.SliceStable(edits, func(i, j int) bool { return edits[i].lo > edits[j].lo })
	out := append([]byte(nil), src...)
	prev := len(src) + 1
	for _, e := range edits {
		if e.hi > prev {
			return nil, fmt.Errorf("two edits overlap near byte %d", e.lo)
		}
		out = append(out[:e.lo], append([]byte(e.text), out[e.hi:]...)...)
		prev = e.lo
	}
	return out, nil
}

const vibeImport = `"github.com/espresso20/ageforge/vibe"`

// stdImport reports whether an import path (quoted) is the standard library's.
func stdImport(quoted string) bool {
	path, _ := strconv.Unquote(quoted)
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".")
}

// fixImports gives a file the vibe import it needs, in the group of the
// module's imports and not the standard library's, drops the fmt import if
// nothing in the file uses fmt, then formats it.
func fixImports(rel string, src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, rel, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	usesVibe, usesFmt := false, false
	ast.Inspect(f, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok {
				switch x.Name {
				case "vibe":
					usesVibe = true
				case "fmt":
					usesFmt = true
				}
			}
		}
		return true
	})
	var fmtSpec, vibeSpec *ast.ImportSpec
	var block *ast.GenDecl
	for _, im := range f.Imports {
		switch im.Path.Value {
		case `"fmt"`:
			fmtSpec = im
		case vibeImport:
			vibeSpec = im
		}
	}
	for _, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			block = gd
			break
		}
	}
	var edits []edit
	off := func(p token.Pos) int { return fset.Position(p).Offset }
	line := func(p token.Pos) int { return fset.Position(p).Line }
	// dropLine is the edit that takes a spec's whole line away.
	dropLine := func(spec *ast.ImportSpec) edit {
		lo, hi := off(spec.Pos()), off(spec.End())
		if spec.Doc != nil {
			lo = off(spec.Doc.Pos())
		}
		for lo > 0 && (src[lo-1] == ' ' || src[lo-1] == '\t') {
			lo--
		}
		if hi < len(src) && src[hi] == '\n' {
			hi++
		}
		return edit{lo, hi, ""}
	}

	if usesVibe && block == nil {
		// A file with no imports at all: the import goes in after the package clause.
		edits = append(edits, edit{off(f.Name.End()), off(f.Name.End()), "\n\nimport " + vibeImport})
	} else if usesVibe && block != nil && !block.Lparen.IsValid() {
		// A single import with no parentheses: "fmt" alone, or something else.
		spec := block.Specs[0].(*ast.ImportSpec)
		switch {
		case vibeSpec != nil:
		case spec == fmtSpec && !usesFmt:
			edits = append(edits, edit{off(block.Pos()), off(block.End()), "import " + vibeImport})
			fmtSpec = nil
		default:
			edits = append(edits, edit{off(block.Pos()), off(block.End()),
				"import (\n\t" + spec.Path.Value + "\n\n\t" + vibeImport + "\n)"})
		}
	} else if usesVibe && block != nil {
		// Where the vibe import belongs: with the imports of the module and
		// of other people's packages, or in a group of its own after the
		// standard library's.
		specs := make([]*ast.ImportSpec, 0, len(block.Specs))
		for _, sp := range block.Specs {
			specs = append(specs, sp.(*ast.ImportSpec))
		}
		misplaced := vibeSpec == nil
		if vibeSpec != nil {
			// Its group: the specs on the lines around it with no blank line between.
			for i, sp := range specs {
				if sp != vibeSpec {
					continue
				}
				alone := true
				for j := i - 1; j >= 0 && line(specs[j].End()) >= line(specs[j+1].Pos())-1; j-- {
					alone = alone && !stdImport(specs[j].Path.Value)
				}
				for j := i + 1; j < len(specs) && line(specs[j].Pos()) <= line(specs[j-1].End())+1; j++ {
					alone = alone && !stdImport(specs[j].Path.Value)
				}
				misplaced = !alone
			}
		}
		if misplaced {
			if vibeSpec != nil {
				edits = append(edits, dropLine(vibeSpec))
			}
			last := specs[len(specs)-1]
			if last == vibeSpec && len(specs) > 1 {
				last = specs[len(specs)-2]
			}
			text := "\t" + vibeImport + "\n"
			if stdImport(last.Path.Value) {
				text = "\n" + text // the standard library's group ends: a new one
			}
			edits = append(edits, edit{off(block.Rparen), off(block.Rparen), text})
		}
	}
	if !usesFmt && fmtSpec != nil {
		if block != nil && !block.Lparen.IsValid() {
			edits = append(edits, edit{off(block.Pos()), off(block.End()), ""})
		} else {
			edits = append(edits, dropLine(fmtSpec))
		}
	}
	if len(edits) > 0 {
		if src, err = applyEdits(src, edits); err != nil {
			return nil, err
		}
	}
	return format.Source(src)
}
