package ui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// dispatcherCases parses ui/input.go: HandleCommand's switch gives every
// word it dispatches on (grouped per case), and each case's handler, followed
// through the input.go functions it calls, gives the literal words it
// compares arguments with.
func dispatcherCases(t *testing.T) (cases [][]string, words map[string]map[string]bool) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "input.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	funcs := map[string]*ast.FuncDecl{}
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok {
			funcs[fd.Name.Name] = fd
		}
	}
	hc := funcs["HandleCommand"]
	if hc == nil {
		t.Fatal("HandleCommand not found in ui/input.go")
	}
	words = map[string]map[string]bool{}
	var sw *ast.SwitchStmt
	ast.Inspect(hc.Body, func(n ast.Node) bool {
		if s, ok := n.(*ast.SwitchStmt); ok && sw == nil {
			sw = s
		}
		return sw == nil
	})
	if sw == nil {
		t.Fatal("HandleCommand has no switch")
	}
	for _, st := range sw.Body.List {
		cc := st.(*ast.CaseClause)
		var names []string
		for _, e := range cc.List {
			if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				s, _ := strconv.Unquote(lit.Value)
				names = append(names, s)
			}
		}
		if len(names) == 0 {
			continue // default
		}
		cases = append(cases, names)
		w := map[string]bool{}
		seen := map[string]bool{"HandleCommand": true}
		var walk func(n ast.Node)
		walk = func(n ast.Node) {
			for k := range comparedWords(n) {
				w[k] = true
			}
			ast.Inspect(n, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if id, ok := call.Fun.(*ast.Ident); ok && funcs[id.Name] != nil && !seen[id.Name] {
						seen[id.Name] = true
						walk(funcs[id.Name].Body)
					}
				}
				return true
			})
		}
		for _, st := range cc.Body {
			walk(st)
		}
		words[names[0]] = w
	}
	return cases, words
}

// comparedWords is every string literal under n used as a switch case,
// compared with == / !=, or passed to strings.EqualFold: the words a handler
// reacts to.
func comparedWords(n ast.Node) map[string]bool {
	out := map[string]bool{}
	add := func(e ast.Expr) {
		if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			s, _ := strconv.Unquote(lit.Value)
			out[strings.ToLower(s)] = true
		}
	}
	ast.Inspect(n, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CaseClause:
			for _, e := range x.List {
				add(e)
			}
		case *ast.BinaryExpr:
			if x.Op == token.EQL || x.Op == token.NEQ {
				add(x.X)
				add(x.Y)
			}
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "EqualFold" {
				for _, a := range x.Args {
					add(a)
				}
			}
		}
		return true
	})
	return out
}

// treeWords is every word the registry lets a player type after c's name:
// subcommand names and aliases and literal argument words, at every depth.
func treeWords(c *Command, out map[string]bool) {
	for _, a := range c.Args {
		for _, w := range a.Words {
			out[w] = true
		}
	}
	for _, s := range c.Subs {
		out[s.Name] = true
		for _, a := range s.Aliases {
			out[a] = true
		}
		treeWords(s, out)
	}
}

// notSubcommands are words handlers compare that are not typed after a
// command name: comparisons in helpers reached while following calls.
var notSubcommands = map[string]bool{
	"":         true, // empty-argument checks
	"allied":   true, // a faction status the diplomacy status list checks
	"worldmap": true, // cmdMap checks which alias it was typed as (worldmap opens on the world)
}

// TestRegistryMatchesDispatcher holds the registry and HandleCommand
// together: every command, alias and subcommand in the registry is one the
// dispatcher takes, and every command and subcommand word the dispatcher
// takes is in the registry.
func TestRegistryMatchesDispatcher(t *testing.T) {
	reg := registry()
	cases, handled := dispatcherCases(t)

	inSwitch := map[string]string{} // word → its case's first word
	for _, names := range cases {
		for _, n := range names {
			inSwitch[n] = names[0]
		}
	}
	inRegistry := map[string]*Command{}
	for _, c := range reg {
		for _, n := range append([]string{c.Name}, c.Aliases...) {
			if prev := inRegistry[n]; prev != nil {
				t.Errorf("registry: %q is both %s and %s", n, prev.Name, c.Name)
			}
			inRegistry[n] = c
		}
	}

	// Registry → dispatcher: names and aliases.
	for n, c := range inRegistry {
		if c.Dashboard {
			if _, ok := inSwitch[n]; ok {
				t.Errorf("%q is marked Dashboard but HandleCommand also dispatches it", n)
			}
			continue
		}
		if _, ok := inSwitch[n]; !ok {
			t.Errorf("registry has %q (%s) but HandleCommand does not dispatch it", n, c.Name)
		}
	}
	// Dispatcher → registry, and each case is one command.
	for _, names := range cases {
		c := inRegistry[names[0]]
		for _, n := range names {
			if inRegistry[n] == nil {
				t.Errorf("HandleCommand dispatches %q but the registry has no such command or alias", n)
			} else if c != nil && inRegistry[n] != c {
				t.Errorf("HandleCommand's case %v spans registry commands %s and %s", names, c.Name, inRegistry[n].Name)
			}
		}
	}

	// Subcommands and literal words, both ways.
	for _, c := range reg {
		if c.Dashboard {
			continue
		}
		reg := map[string]bool{}
		treeWords(c, reg)
		got := handled[inSwitch[c.Name]]
		for _, w := range sortedSet(reg) {
			if !got[w] {
				t.Errorf("registry: %s takes %q, but its handler never checks for it", c.Name, w)
			}
		}
		for _, w := range sortedSet(got) {
			if !reg[w] && !notSubcommands[w] {
				t.Errorf("the %s handler reacts to %q, but the registry doesn't list it under %s", c.Name, w, c.Name)
			}
		}
	}
}

// TestRegistryCommandsRun types every registry command, alias and
// subcommand path, bare, into HandleCommand on a fresh game: none may come
// back as an unknown command.
func TestRegistryCommandsRun(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	t.Chdir(t.TempDir())
	eng := game.NewGameEngine()
	var lines []string
	var walk func(prefix string, c *Command)
	walk = func(prefix string, c *Command) {
		for _, s := range c.Subs {
			for _, n := range append([]string{s.Name}, s.Aliases...) {
				lines = append(lines, prefix+" "+n)
				walk(prefix+" "+n, s)
			}
		}
	}
	for _, c := range registry() {
		if c.Dashboard {
			continue
		}
		for _, n := range append([]string{c.Name}, c.Aliases...) {
			lines = append(lines, n)
			walk(n, c)
		}
	}
	for _, line := range lines {
		if strings.Contains(line, "prestige confirm yes") || strings.HasPrefix(line, "account") && !strings.HasSuffix(line, "list") {
			continue // would reset the run or touch account files
		}
		if res := HandleCommand(line, eng); strings.Contains(res.Message, "Unknown command") {
			t.Errorf("%q: %s", line, res.Message)
		}
	}
}

// TestDashboardCommandsHandled: a registry command marked Dashboard is one
// submitInput handles itself.
func TestDashboardCommandsHandled(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "dashboard.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var words map[string]bool
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == "submitInput" {
			words = comparedWords(fd.Body)
		}
	}
	if words == nil {
		t.Fatal("Dashboard.submitInput not found")
	}
	for _, c := range registry() {
		if c.Dashboard && !words[c.Name] {
			t.Errorf("registry marks %q Dashboard, but submitInput doesn't handle it", c.Name)
		}
	}
}

// TestRegistryShape checks what the other readers rely on: every top-level
// command is listed in the Help panel (a section or a panel line) unless it
// is a bare alias view, every help form starts with its command, and the
// Dangerous list is the one the PR promises.
func TestRegistryShape(t *testing.T) {
	sections := map[string]bool{}
	for _, s := range helpSections {
		sections[s.name] = true
	}
	var dangerous []string
	var walk func(path string, c *Command)
	walk = func(path string, c *Command) {
		for _, u := range c.Help {
			if !strings.HasPrefix(u.Form, strings.Fields(path)[0]) {
				t.Errorf("%s: help form %q doesn't start with the command", path, u.Form)
			}
		}
		if c.Dangerous {
			dangerous = append(dangerous, path)
		}
		for _, s := range c.Subs {
			walk(path+" "+s.Name, s)
		}
	}
	for _, c := range registry() {
		if c.Section != "" && !sections[c.Section] {
			t.Errorf("%s: unknown help section %q", c.Name, c.Section)
		}
		if c.Section == "" && c.Panel == "" && c.Name != "techs" && c.Name != "map" {
			t.Errorf("%s is in no help section and no panel line", c.Name)
		}
		walk(c.Name, c)
	}
	sort.Strings(dangerous)
	want := "account import,account recover,account switch,account wipe,diplomacy raid,dismiss,festival confirm yes,harbinger invite,load,plan clear,prestige confirm yes,quit,research cancel,sell"
	if got := strings.Join(dangerous, ","); got != want {
		t.Errorf("Dangerous commands = %s\nwant %s", got, want)
	}
}

func sortedSet(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
