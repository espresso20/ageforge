package main

import (
	"fmt"
	"go/ast"
	"path"
	"regexp"
	"strings"
	"unicode"
)

// code_rows.go reads text out of ordinary code: the panels, the command
// replies, the log lines, the refusals. There is no table of fields to go
// by here, so the reader works from what a literal is used for. It leaves
// out what the code only matches on or looks up (keys, identifiers, file
// names, the patterns of text operations), what only a developer reads, and
// what has no words in it; it keeps the rest, and marks what it cannot be
// sure reaches the screen.

// Why a unit is left out. The export reports the counts.
const (
	whyReadBack = "text the game reads back out of old saves, so it has to stay as it is"
	whyNoWords  = "no words in it: layout, a number format or a glyph"
	whyMatched  = "text the code compares or looks things up by, not text it shows"
	whyKey      = "a key or identifier"
	whyOp       = "the pattern of a text operation (what to find, cut or replace), not shown"
	whyDev      = "read by developers only: the dev console, a debugging dump or an internal failure"
	whyPath     = "a file name, path or web address"
	whySyntax   = "a command as the player types it, which must match the command parser"
	whyArt      = "art or a map glyph, not words"
)

// keyish matches a single token that reads as a key: lower case, digits and
// the punctuation keys are written with. capsKey is the same in capitals
// with a digit or an underscore in it ("UTF-8", "TERM_PROGRAM"): a word in
// capitals alone ("FLOWS") is a label.
var (
	keyish  = regexp.MustCompile(`^[a-z0-9_.:/#*<>=+,@&|\\-]+$`)
	capsKey = regexp.MustCompile(`^[A-Z0-9_-]*[0-9_][A-Z0-9_-]*$`)
)

// keyLike reports whether text reads as a key or identifier rather than as
// words for a player.
func keyLike(s string) bool {
	if s == "" || strings.ContainsAny(s, " \t\n") {
		return false
	}
	if keyish.MatchString(s) || capsKey.MatchString(s) {
		return true
	}
	if strings.ContainsAny(s, "_/\\") {
		return true
	}
	// camelCase, PascalCase with a second capital, or a dotted name.
	prevLower := false
	for i, r := range s {
		if unicode.IsUpper(r) && prevLower {
			return true
		}
		if r == '.' && i > 0 && i < len(s)-1 {
			rest := []rune(s[i+1:])
			if unicode.IsLetter(rest[0]) {
				return true
			}
		}
		prevLower = unicode.IsLower(r)
	}
	return false
}

// Calls whose string arguments are never shown.
var (
	// Package functions that work on text: every argument after the first is
	// a pattern. NewReplacer and the Sscan family take patterns throughout.
	opPkgs     = map[string]bool{"strings": true, "bytes": true, "regexp": true, "strconv": true, "utf8": true, "unicode": true}
	opAllArgs  = map[string]bool{"strings.NewReplacer": true, "fmt.Sscanf": true, "fmt.Sscan": true, "fmt.Fscanf": true, "regexp.MustCompile": true, "regexp.Compile": true, "time.Parse": true}
	pathPkgs   = map[string]bool{"filepath": true, "path": true, "os": true, "exec": true, "http": true, "url": true, "json": true, "hex": true, "base64": true, "ioutil": true, "io": true, "runtime": true, "nerdfont": true}
	devCalls   = map[string]bool{"panic": true, "log.Printf": true, "log.Println": true, "log.Fatalf": true, "log.Fatal": true, "debugf": true}
	syntaxArgs = map[string]bool{"sub#0": true, "sub#1": true, "panel#0": true, "confirmYes#0": true, "usageFor#0": true, "subUsage#0": true, "helpRow#0": true}
	// Calls that take a noun and put it in a sentence ("Unknown %s '%s'.").
	nounArgs = map[string]bool{"unknownKeyError#0": true}
	// Calls that put a single lower-case word in front of the player.
	wordCalls = map[string]bool{"Count": true, "Plural": true, "plural": true, "pluralize": true}
	// Calls that write a line to the game log; the text is their last
	// argument or the format after the category.
	logCalls = map[string]bool{"addLog": true, "AddLog": true, "log": true, "logf": true, "addLogRoutine": true, "logRoutine": true}
	errCalls = map[string]bool{"fmt.Errorf": true, "errors.New": true}
)

// fileNote says what a source file is to a player.
type fileNote struct {
	place string // "the Trade panel": where its text shows
	area  *area
	age   string // an age key, or a key world.ageOf understands; "menu" for before the first age
	seen  int
	voice bool   // its sentences are story rather than messages
	skip  string // set to leave the whole file out, with the reason
}

type codeReader struct {
	c      *catalog
	f      *srcFile
	note   fileNote
	n      map[string]int              // rows made so far, by function
	dupFn  map[string]bool             // function names more than one receiver has
	prose  map[*ast.CompositeLit]bool  // lists and maps that hold worded text
	parent map[*unit]*ast.CompositeLit // the list or map a unit is an element of
}

func newCodeReader(c *catalog, f *srcFile, units []*unit) *codeReader {
	cr := &codeReader{
		c: c, f: f, note: noteFor(f), n: map[string]int{}, dupFn: map[string]bool{},
		prose: map[*ast.CompositeLit]bool{}, parent: map[*unit]*ast.CompositeLit{},
	}
	seen := map[string]string{}
	for _, d := range f.ast.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		recv := ""
		if fd.Recv != nil && len(fd.Recv.List) > 0 {
			recv = strings.TrimPrefix(exprText(fd.Recv.List[0].Type), "*")
		}
		if prev, ok := seen[fd.Name.Name]; ok && prev != recv {
			cr.dupFn[fd.Name.Name] = true
		}
		seen[fd.Name.Name] = recv
	}
	// Which lists and maps hold worded text: a single lower-case word in one
	// of those is text too, not a key.
	lits := map[[2]int]*ast.CompositeLit{}
	ast.Inspect(f.ast, func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			for _, el := range cl.Elts {
				v := el
				if kv, ok := el.(*ast.KeyValueExpr); ok {
					v = kv.Value
				}
				lits[[2]int{c.m.fset.Position(v.Pos()).Offset, c.m.fset.Position(v.End()).Offset}] = cl
			}
		}
		return true
	})
	for _, u := range units {
		if u.near != "elem" {
			continue
		}
		if cl, ok := lits[[2]int{u.start, u.end}]; ok {
			cr.parent[u] = cl
			if strings.Contains(strings.TrimSpace(u.text), " ") && hasWord(u.text) {
				cr.prose[cl] = true
			}
		}
	}
	return cr
}

// read makes the row for a unit in ordinary code, or says why not.
func (cr *codeReader) read(u *unit) (*row, string) {
	if cr.note.skip != "" {
		return nil, cr.note.skip
	}
	if !hasWord(u.text) {
		return nil, whyNoWords
	}
	for _, marker := range cr.c.rules.readBack {
		if strings.Contains(u.text, marker) {
			return nil, whyReadBack
		}
	}
	if r, why, done := cr.special(u); done {
		if r != nil {
			cr.defaults(u, r)
		}
		return r, why
	}
	switch u.near {
	case "compare", "case", "index", "mapkey":
		return nil, whyMatched
	}
	isErr, isLog, single, noun := false, false, false, false
	for i, cs := range u.calls {
		q := cs.fn
		if cs.pkg != "" {
			q = cs.pkg + "." + cs.fn
		}
		switch {
		case devCalls[q]:
			return nil, whyDev
		case opAllArgs[q]:
			return nil, whyOp
		case i == 0 && u.near == "call" && opPkgs[cs.pkg] && cs.arg >= 1:
			return nil, whyOp
		case i == 0 && u.near == "call" && pathPkgs[cs.pkg]:
			return nil, whyPath
		case i == 0 && u.near == "call" && syntaxArgs[fmt.Sprintf("%s#%d", cs.fn, cs.arg)]:
			return nil, whySyntax
		case i == 0 && u.near == "call" && cs.fn == "Format" && cs.pkg == "":
			return nil, whyOp // a date layout
		case errCalls[q]:
			isErr = true
		case logCalls[cs.fn] && cs.pkg == "":
			if i == 0 && cs.arg == 0 && len(cs.args) > 1 {
				return nil, whyKey // the log line's category
			}
			isLog = true
		case i == 0 && wordCalls[cs.fn] && cs.arg >= 1:
			single = true
		case i == 0 && u.near == "call" && nounArgs[fmt.Sprintf("%s#%d", cs.fn, cs.arg)]:
			noun = true
		}
	}
	if keyLike(u.text) && !single && !noun && !cr.prose[cr.parent[u]] {
		return nil, whyKey
	}
	if looksLikePath(u.text) {
		return nil, whyPath
	}
	if cr.c.isCommand(u.text) && !single && !noun {
		return nil, whySyntax
	}
	if strings.IndexFunc(u.text, func(r rune) bool { return r < 0x20 && r != '\n' && r != '\t' }) >= 0 {
		return nil, whyKey
	}

	r := &row{age: -1}
	fn := u.decl
	if u.fn != "" && cr.dupFn[u.fn] && u.recv != "" {
		fn = u.recv + "." + u.fn
	}
	if fn == "" {
		fn = "top"
	}
	cr.n[fn]++
	r.id = fmt.Sprintf("%s.%s.%s.%02d", strings.ReplaceAll(strings.Trim(cr.f.dir, "."), "/", "."), strings.TrimSuffix(path.Base(cr.f.rel), ".go"), fn, cr.n[fn])
	r.id = strings.TrimPrefix(r.id, ".")
	if cr.f.dir == "." {
		r.id = "main" + "." + strings.TrimPrefix(r.id, ".")
	}

	words := countWords(u.text)
	sentence := words >= 5 || (words >= 2 && endsLikeSentence(u.text))
	what := ""
	switch {
	case isErr:
		r.kind = kindMessage
		what = "A refusal or error message"
		r.seen = seenNews
		low := strings.TrimSpace(u.text)
		if low != "" && unicode.IsLower([]rune(low)[0]) && (strings.Contains(low, "%w") || strings.HasPrefix(low, "failed to")) {
			r.unsure = true
			what = "An error from the plumbing under the game (files, the network), which may never be shown as written"
		}
	case isLog:
		r.kind = kindMessage
		what = "A line in the game log"
		r.seen = seenNews
	case single:
		r.kind = kindLabel
		what = "The word the game uses after a count of one (“1 " + strings.TrimSpace(u.text) + "”),"
		if u.call().arg >= 2 {
			what = "The word the game uses after a count of two or more (“3 " + strings.TrimSpace(u.text) + "”),"
		}
	case sentence && cr.note.voice:
		r.kind = kindVoice
		what = "A line of story"
	case sentence:
		r.kind = kindMessage
		what = "A line of text"
	default:
		r.kind = kindLabel
		what = "A label or short piece of text"
	}
	if u.fn == "String" && u.recv != "" && keyish.MatchString(strings.TrimSpace(u.text)) {
		// One lower-case word that names a value: other code may match on
		// it as well as show it.
		r.unsure = true
		what = "The name of a setting or state (" + u.recv + "), which the code may also match on, so it may not be safe to change,"
	}
	where := what + " " + cr.note.place + "."
	if u.chain != nil && u.chain.parts >= 1 && strings.TrimSpace(u.chain.pattern) != strings.TrimSpace(u.text) {
		where += " It is one piece of a line the game puts together, which reads: “" + oneLine(u.chain.pattern) + "” (… is what the game fills in)."
	}
	r.where = where
	cr.defaults(u, r)
	return r, ""
}

// defaults fills what the file's note decides, where a row has not set it.
func (cr *codeReader) defaults(u *unit, r *row) {
	if r.area == nil {
		r.area = cr.note.area
	}
	if r.seen == 0 {
		r.seen = cr.note.seen
	}
	if r.age < 0 && cr.note.age != "" {
		if cr.note.age == "menu" {
			r.age = 0
		} else {
			r.age = cr.c.w.ageOf(cr.note.age)
		}
	}
	switch cr.f.dir {
	case "ui", "game", "boon":
		if !cr.c.rules.exemptFiles[cr.f.rel] {
			r.rules |= rulesUI
		}
	case "config":
		r.rules |= rulesConfig | rulesDocs
	}
}

// endsLikeSentence reports whether text ends the way a sentence does.
func endsLikeSentence(s string) bool {
	t := strings.TrimSpace(bare(s))
	t = strings.TrimRight(t, `"')]`)
	return strings.HasSuffix(t, ".") || strings.HasSuffix(t, "?") || strings.HasSuffix(t, "!")
}

var pathish = regexp.MustCompile(`^(https?://|\.{0,2}/)|\.(json|go|md|txt|tmp|exe|corrupt|ttf|zip|csv|html)$`)

func looksLikePath(s string) bool {
	return !strings.Contains(s, " ") && pathish.MatchString(s)
}

// oneLine writes a pattern on one line for a where sentence.
func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " / ")
	return strings.TrimSpace(s)
}

// isCommand reports whether text is a command a player types ("gather wood
// 5", "account badges", "plan build <building> [count]") and not a sentence:
// it starts with a command, and every word after is a subcommand, a word an
// argument takes, a key, a number or an argument in brackets. Such text must
// match the command parser, so it is not for rewriting.
func (c *catalog) isCommand(s string) bool {
	words := strings.Fields(s)
	if len(words) < 2 || len(words) > 6 || !c.commands[words[0]] {
		return false // one word alone is a key or a label, and is judged as one
	}
	for _, w := range words[1:] {
		_, isKey := c.w.name[w]
		switch {
		case c.cmdWords[w], isKey, strings.Contains(w, "_"):
		case commandArg.MatchString(w):
		default:
			return false
		}
	}
	return true
}

// commandArg matches an argument as a command's form writes it: a number,
// or a name in angle or square brackets.
var commandArg = regexp.MustCompile(`^([0-9]+|<[a-z0-9_|-]+>|\[[a-z0-9_|<>-]+\])$`)

// readCommands lists the words of the command registry: the commands a line
// can start with, and the subcommands and argument words that can follow.
func (c *catalog) readCommands() {
	c.commands, c.cmdWords = map[string]bool{}, map[string]bool{}
	f := c.m.byRel["ui/commands.go"]
	if f == nil {
		return
	}
	for _, u := range c.m.units(f) {
		s := u.st()
		cs := u.call()
		inCall := len(u.calls) > 0 && cs.out == 0
		depth := 0
		for _, st := range u.structs {
			if st.typ == "ui.Command" {
				depth++
			}
		}
		switch {
		case !inCall && s.typ == "ui.Command" && (s.field == "Name" || s.field == "Aliases"):
			c.cmdWords[u.text] = true
			if depth == 1 && u.fn == "registry" && u.text != "confirm" && u.text != "yes" {
				c.commands[u.text] = true
			}
		case !inCall && s.typ == "ui.Arg" && s.field == "Words":
			c.cmdWords[u.text] = true
		case inCall && cs.fn == "sub" && cs.arg == 0:
			c.cmdWords[u.text] = true
		case inCall && cs.fn == "panel" && cs.arg == 0:
			c.cmdWords[u.text] = true
			c.commands[u.text] = true
		case inCall && cs.fn == "panel" && cs.arg >= 2:
			c.cmdWords[u.text] = true
			c.commands[u.text] = true
		}
	}
}
