package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/espresso20/ageforge/config"
)

// lint.go holds a rewrite to the game's own copy rules.
//
// The rules live in the game's lint tests: the wording guards over ui, game
// and boon, the style guard over config's text, the wiki's lint, and the
// flavor catalog's checks. Their word lists and patterns are read from those
// test files when the import runs, so a term retired there is refused here
// the same day and the two cannot drift apart. What each rule does with its
// list (which text it applies to, what counts as a hit) is restated here, a
// few lines apiece. TestRulesAreRead fails if a list moves or changes shape.

// pattern is a regular expression from a lint, with what the lint says
// about it.
type pattern struct {
	re   *regexp.Regexp
	note string
}

// copyRules are the lists and patterns read from the lint tests.
type copyRules struct {
	// ui/copy_guard_test.go: the wording guards over ui, game and boon.
	retiredTerms []pattern
	rawAmount    *regexp.Regexp
	bang         *regexp.Regexp
	exemptFiles  map[string]bool // files the wording guards skip
	readBack     []string        // text the game parses back out of old saves
	// ui/stale_copy_test.go
	retiredCopy []pattern
	// ui/docs_lint_test.go: the wiki's lint.
	retiredDoc []pattern
	// config/effect_text_test.go and config/event_flavor_test.go
	styleBans []string
	rawKey    *regexp.Regexp
	verbLeft  *regexp.Regexp
	// flavor/flavor_test.go and flavor/generator_test.go
	shapes     []pattern           // banned sentence shapes
	restating  []string            // phrases that repeat what the log already said
	eraWords   map[string][]string // a word -> the spans of ages it belongs to
	eraRes     map[string]*regexp.Regexp
	minWords   int
	maxWords   int
	maxLine    int
	otherLines map[string]string // every catalog sentence -> its id, to catch a duplicate
}

// lintSource is one lint test file, parsed.
type lintSource struct {
	rel  string
	file *ast.File
}

func parseLint(root, rel string) (*lintSource, error) {
	f, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, filepath.FromSlash(rel)), nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("cannot read the copy rules in %s: %w", rel, err)
	}
	return &lintSource{rel, f}, nil
}

// value finds the expression a name is given in the file: a package-level
// var or const, or a := inside a function.
func (ls *lintSource) value(name string) (ast.Expr, error) {
	var found ast.Expr
	ast.Inspect(ls.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ValueSpec:
			for i, id := range x.Names {
				if id.Name == name && i < len(x.Values) {
					found = x.Values[i]
				}
			}
		case *ast.AssignStmt:
			for i, l := range x.Lhs {
				if id, ok := l.(*ast.Ident); ok && id.Name == name && i < len(x.Rhs) && x.Tok == token.DEFINE {
					found = x.Rhs[i]
				}
			}
		}
		return found == nil
	})
	if found == nil {
		return nil, fmt.Errorf("the copy rules in %s no longer define %s", ls.rel, name)
	}
	return found, nil
}

// regex reads a regexp.MustCompile("...") expression.
func regex(e ast.Expr) (*regexp.Regexp, bool) {
	call, ok := e.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 || exprText(call.Fun) != "regexp.MustCompile" {
		return nil, false
	}
	s, ok := stringLit(call.Args[0])
	if !ok {
		return nil, false
	}
	re, err := regexp.Compile(s)
	return re, err == nil
}

func (ls *lintSource) regex(name string) (*regexp.Regexp, error) {
	e, err := ls.value(name)
	if err != nil {
		return nil, err
	}
	re, ok := regex(e)
	if !ok {
		return nil, fmt.Errorf("%s in %s is no longer a single regular expression", name, ls.rel)
	}
	return re, nil
}

// strings reads a []string{...} literal.
func (ls *lintSource) strings(name string) ([]string, error) {
	e, err := ls.value(name)
	if err != nil {
		return nil, err
	}
	cl, ok := e.(*ast.CompositeLit)
	if !ok {
		return nil, fmt.Errorf("%s in %s is no longer a list", name, ls.rel)
	}
	var out []string
	for _, el := range cl.Elts {
		s, ok := stringLit(el)
		if !ok {
			return nil, fmt.Errorf("%s in %s holds something that is not text", name, ls.rel)
		}
		out = append(out, s)
	}
	return out, nil
}

// patterns reads a list of {pattern, note} or {note, pattern} pairs. A
// plain string where the pattern goes is matched as written, in any case.
func (ls *lintSource) patterns(name string) ([]pattern, error) {
	e, err := ls.value(name)
	if err != nil {
		return nil, err
	}
	cl, ok := e.(*ast.CompositeLit)
	if !ok {
		return nil, fmt.Errorf("%s in %s is no longer a list", name, ls.rel)
	}
	var out []pattern
	for _, el := range cl.Elts {
		pair, ok := el.(*ast.CompositeLit)
		if !ok || len(pair.Elts) != 2 {
			return nil, fmt.Errorf("%s in %s is no longer a list of pairs", name, ls.rel)
		}
		var p pattern
		var texts []string
		for _, part := range pair.Elts {
			if kv, ok := part.(*ast.KeyValueExpr); ok {
				part = kv.Value
			}
			if re, ok := regex(part); ok {
				p.re = re
			} else if s, ok := stringLit(part); ok {
				texts = append(texts, s)
			}
		}
		switch {
		case p.re != nil && len(texts) == 1:
			p.note = texts[0]
		case p.re == nil && len(texts) == 2:
			p.re = regexp.MustCompile("(?i)" + regexp.QuoteMeta(texts[0]))
			p.note = texts[1]
		default:
			return nil, fmt.Errorf("%s in %s has a pair this tool cannot read", name, ls.rel)
		}
		out = append(out, p)
	}
	return out, nil
}

// number reads an integer constant.
func (ls *lintSource) number(name string) (int, error) {
	e, err := ls.value(name)
	if err != nil {
		return 0, err
	}
	if bl, ok := e.(*ast.BasicLit); ok && bl.Kind == token.INT {
		return strconv.Atoi(bl.Value)
	}
	return 0, fmt.Errorf("%s in %s is no longer a whole number", name, ls.rel)
}

// loadRules reads the lint tests' lists from the source under root.
func loadRules(root string) (*copyRules, error) {
	cr := &copyRules{eraWords: map[string][]string{}, eraRes: map[string]*regexp.Regexp{}}
	var err error
	fail := func(e error) {
		if err == nil {
			err = e
		}
	}
	src := func(rel string) *lintSource {
		ls, e := parseLint(root, rel)
		if e != nil {
			fail(e)
			return &lintSource{rel, &ast.File{}}
		}
		return ls
	}
	get := func(v any, e error) any {
		fail(e)
		return v
	}
	guard := src("ui/copy_guard_test.go")
	cr.retiredTerms, _ = get(guard.patterns("retiredTerms")).([]pattern)
	cr.rawAmount, _ = get(guard.regex("rawAmountRe")).(*regexp.Regexp)
	cr.bang, _ = get(guard.regex("bang")).(*regexp.Regexp)
	cr.readBack, _ = get(guard.strings("copyExemptLiterals")).([]string)
	cr.exemptFiles = map[string]bool{}
	if e, e2 := guard.value("copyExemptFiles"); e2 != nil {
		fail(e2)
	} else if cl, ok := e.(*ast.CompositeLit); ok {
		for _, el := range cl.Elts {
			if kv, ok := el.(*ast.KeyValueExpr); ok {
				if rel, ok := stringLit(kv.Key); ok {
					cr.exemptFiles[rel] = true
				}
			}
		}
	}
	stale := src("ui/stale_copy_test.go")
	cr.retiredCopy, _ = get(stale.patterns("retiredCopy")).([]pattern)
	docs := src("ui/docs_lint_test.go")
	cr.retiredDoc, _ = get(docs.patterns("retiredDocText")).([]pattern)
	style := src("config/effect_text_test.go")
	cr.styleBans, _ = get(style.strings("styleBans")).([]string)
	cr.rawKey, _ = get(style.regex("rawKeyRe")).(*regexp.Regexp)
	events := src("config/event_flavor_test.go")
	cr.verbLeft, _ = get(events.regex("pctDirective")).(*regexp.Regexp)
	fl := src("flavor/flavor_test.go")
	cr.shapes, _ = get(fl.patterns("tells")).([]pattern)
	cr.restating, _ = get(fl.strings("banned")).([]string)
	cr.minWords, _ = get(fl.number("minWords")).(int)
	cr.maxWords, _ = get(fl.number("maxWords")).(int)
	gen := src("flavor/generator_test.go")
	cr.maxLine, _ = get(gen.number("maxLineRunes")).(int)
	if e, e2 := fl.value("eraMarkers"); e2 != nil {
		fail(e2)
	} else if cl, ok := e.(*ast.CompositeLit); ok {
		for _, el := range cl.Elts {
			kv, ok := el.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			word, ok := stringLit(kv.Key)
			homes, ok2 := kv.Value.(*ast.CompositeLit)
			if !ok || !ok2 {
				continue
			}
			for _, h := range homes.Elts {
				cr.eraWords[word] = append(cr.eraWords[word], exprText(h))
			}
			cr.eraRes[word] = regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
		}
	}
	if err != nil {
		return nil, err
	}
	return cr, nil
}

// spanNames are the flavor package's five spans of ages, in plain words.
var spanNames = map[string]string{
	"eraAncient": "the first five ages (Primitive to Classical)", "eraFeudal": "the Medieval to Colonial ages",
	"eraIndustrial": "the Industrial to Modern ages", "eraDigital": "the Information to Fusion ages",
	"eraCosmic": "the Space to Transcendent ages",
}

var allSpans = []string{"eraAncient", "eraFeudal", "eraIndustrial", "eraDigital", "eraCosmic"}

// check returns what is wrong with a rewrite of a row under the game's copy
// rules, one sentence a problem. A rule the current text already breaks is
// not held against the rewrite: the import does not make old text a reason
// to refuse new text.
func (cr *copyRules) check(r *row, yours string) []string {
	var out []string
	// hit reports a problem when the rewrite trips a test the original passes.
	hit := func(trips func(s string) bool, problem string) {
		if trips(yours) && !trips(r.current) {
			out = append(out, problem)
		}
	}
	has := func(sub string) func(string) bool {
		return func(s string) bool { return strings.Contains(s, sub) }
	}
	matches := func(re *regexp.Regexp) func(string) bool {
		return func(s string) bool { return re.MatchString(s) }
	}

	if r.rules&rulesUI != 0 {
		hit(has("—"), "has an em dash: use a period, colon, comma or parentheses")
		hit(matches(cr.bang), "has an exclamation mark: routine events are stated, not shouted")
		hit(matches(cr.rawAmount), "prints an amount with a bare number verb")
		if strings.Contains(yours, " ") {
			for _, p := range cr.retiredTerms {
				hit(matches(p.re), fmt.Sprintf("uses a retired term (“%s”): the game says %s", p.re.FindString(yours), p.note))
			}
		}
		if r.u != nil && strings.HasPrefix(r.u.f.rel, "ui/") {
			for _, p := range cr.retiredCopy {
				hit(matches(p.re), "brings back retired text: "+p.note)
			}
		}
	}
	if r.rules&rulesConfig != 0 {
		hit(func(s string) bool { return strings.ContainsAny(s, "—–") }, "has a dash: use a period, colon, comma or parentheses")
		hit(has("!"), "has an exclamation mark")
		hit(matches(cr.rawKey), fmt.Sprintf("has a raw key in it (“%s”): write the name a player sees", cr.rawKey.FindString(yours)))
		for _, b := range cr.styleBans {
			hit(func(s string) bool { return strings.Contains(strings.ToLower(s), b) }, fmt.Sprintf("uses “%s”, which the style guard for game data bans (US spelling, the glossary, stock words)", b))
		}
	}
	if r.rules&rulesDocs != 0 {
		for _, p := range cr.retiredDoc {
			hit(matches(p.re), fmt.Sprintf("uses “%s”, which the wiki's lint refuses, and the wiki quotes this text: %s", p.re.FindString(yours), p.note))
		}
	}
	if r.rules&rulesNoMark != 0 {
		hit(func(s string) bool { return strings.ContainsAny(s, "[]") }, "has a square bracket, which this line may not carry (the game wraps it in a style tag of its own)")
		hit(matches(cr.verbLeft), "has a % directive, which this line may not carry")
	}
	if r.rules&rulesSkel != 0 {
		out = append(out, cr.checkSentence(r, yours)...)
	}
	if r.rules&rulesBank != 0 {
		out = append(out, cr.checkBankEntry(r, yours)...)
	}
	if r.rules&(rulesSkel|rulesBank) != 0 {
		for _, b := range cr.restating {
			hit(func(s string) bool { return strings.Contains(strings.ToLower(s), b) }, fmt.Sprintf("says “%s”, which repeats what the log line above it already said", b))
		}
		hit(func(s string) bool { return strings.HasPrefix(s, "The party ") || strings.Contains(s, ". The party ") }, "uses “the party” as a bare subject: name someone or something")
		spans := r.eras
		if len(spans) == 0 {
			spans = allSpans
		}
		for word, homes := range cr.eraWords {
			if !cr.eraRes[word].MatchString(yours) || cr.eraRes[word].MatchString(r.current) {
				continue
			}
			for _, sp := range spans {
				if !contains(homes, sp) {
					out = append(out, fmt.Sprintf("has the word “%s”, which belongs to %s, but this line can also show in %s", word, spanNames[homes[0]], spanNames[sp]))
					break
				}
			}
		}
	}
	if r.rules&rulesSaveNam != 0 {
		hit(func(s string) bool {
			for _, c := range strings.TrimSpace(s) {
				if !unicode.IsLetter(c) && c != ' ' {
					return true
				}
			}
			return strings.Contains(strings.TrimSpace(s), "  ")
		}, "has something other than letters and single spaces, and a save's name becomes a file name")
	}
	out = append(out, checkTech(r, yours)...)
	return out
}

// sentenceWords counts the words of a catalog sentence the way the catalog's
// own length check does: a slot or a placeholder is one word.
func sentenceWords(s string) int {
	s = strings.ReplaceAll(s, "~", "x")
	s = slotRe.ReplaceAllString(s, "x")
	return len(strings.Fields(s))
}

// checkSentence applies the flavor catalog's per-sentence checks.
func (cr *copyRules) checkSentence(r *row, yours string) []string {
	var out []string
	text := strings.TrimSpace(yours)
	for _, p := range cr.shapes {
		if p.re.MatchString(text) && !p.re.MatchString(strings.TrimSpace(r.current)) {
			out = append(out, "is a sentence shape the catalog bans: "+p.note)
		}
	}
	if strings.Contains(text, ":") && !strings.Contains(r.current, ":") {
		out = append(out, "has a colon, which the catalog bans (it only ever introduced a verdict)")
	}
	if head, tail, ok := strings.Cut(text, ","); ok {
		first := strings.ToLower(strings.Fields(strings.TrimSpace(tail) + " x")[0])
		switch first {
		case "and", "but", "which", "both", "all", "though", "so":
			if sentenceWords(head) <= 3 {
				out = append(out, fmt.Sprintf("opens with a fragment (“%s”) and then a judgment, a shape the catalog bans", strings.TrimSpace(head)))
			}
		}
	}
	if n := sentenceWords(text); n < cr.minWords || n > cr.maxWords {
		out = append(out, fmt.Sprintf("is %d words long: a catalog sentence is %d to %d words", n, cr.minWords, cr.maxWords))
	}
	if strings.ContainsAny(text, "[]%") {
		out = append(out, "has a square bracket or a % sign: story lines carry no markup")
	}
	if strings.HasSuffix(text, ".") || strings.HasSuffix(text, "!") || strings.HasSuffix(text, "?") {
		out = append(out, "ends with its own punctuation: the game adds the full stop")
	}
	sig := text
	if i := strings.IndexAny(sig, "~{"); i >= 0 {
		sig = strings.TrimRight(sig[:i], " ")
	}
	if len(sig) < 12 {
		out = append(out, fmt.Sprintf("opens with only “%s” before its first slot: the game tells its sentences apart by their first 12 characters", sig))
	}
	if id, dup := cr.otherLines[text]; dup && id != r.id {
		out = append(out, "is word for word the sentence "+id+" already says")
	}
	return out
}

// checkBankEntry applies the catalog's checks on a word-list entry.
func (cr *copyRules) checkBankEntry(r *row, yours string) []string {
	var out []string
	text := strings.TrimSpace(yours)
	if text == "" {
		return nil
	}
	if strings.ContainsAny(text, "[]%{}~") {
		out = append(out, "is not a plain phrase: no brackets, braces, % or ~ in a word-list entry")
	}
	if strings.HasSuffix(text, ".") || strings.HasSuffix(text, ",") {
		out = append(out, "ends in punctuation: the entry lands in the middle of a sentence")
	}
	if first := []rune(text)[0]; unicode.IsUpper(first) {
		out = append(out, "starts with a capital letter: the entry lands in the middle of a sentence")
	}
	return out
}

// checkTech applies the tech tree's rule for a card code: two to
// config.TechCodeMax capital letters, no two techs alike. A tech that sets
// no code takes one from its name, so a rename can change it.
func checkTech(r *row, yours string) []string {
	if r.u == nil || r.u.st().typ != "config.TechDef" {
		return nil
	}
	s := r.u.st()
	var code string
	switch {
	case s.field == "Code":
		code = strings.TrimSpace(yours)
	case s.field == "Name" && s.sib["Code"] == "":
		code = config.TechCodeFor(strings.TrimSpace(yours))
	default:
		return nil
	}
	valid := len(code) >= 2 && len(code) <= config.TechCodeMax
	for _, c := range code {
		if c < 'A' || c > 'Z' {
			valid = false
		}
	}
	if !valid {
		if s.field == "Name" {
			return []string{fmt.Sprintf("would give the tech the card code “%s” (a tech with no code of its own takes it from the first word of its name): a code is 2 to %d capital letters", code, config.TechCodeMax)}
		}
		return []string{fmt.Sprintf("is not a card code: 2 to %d capital letters", config.TechCodeMax)}
	}
	for _, t := range config.Technologies() {
		if t.Key != s.sib["Key"] && t.Code == code {
			if s.field == "Name" {
				return []string{fmt.Sprintf("would give the tech the card code %s, which %s already has (a tech with no code of its own takes it from the first word of its name)", code, t.Name)}
			}
			return []string{fmt.Sprintf("is the card code %s already has", t.Name)}
		}
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
