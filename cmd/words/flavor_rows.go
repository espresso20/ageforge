package main

import (
	"fmt"
	"go/ast"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// flavor_rows.go reads the flavor catalogs: the sentences each catalog can
// say, grouped into pools by the ages they fit, and the word lists some of
// those sentences draw a phrase from.
//
// A catalog file declares its pools with pool(), agePool() and agePools().
// The reader takes the pool's name and ages from those calls, so a sentence's
// id is its catalog, its pool and its place in the pool.

// moment is one flavor catalog: one game situation the package narrates.
type moment struct {
	prefix string // what its pool names start with ("exp_fail")
	key    string // its part of an id ("expedition_fail")
	title  string // what the README calls it
	when   string // "when an expedition fails", for where sentences
	about  string // for a kind restriction: what the kind describes
	from   string // the age key before which it cannot happen, "" for the first age
	area   *area
}

// moments are the catalogs, longest prefix first where one prefix starts
// another. TestFlavorPoolsAreAllRead fails when a pool matches none.
var moments = []*moment{
	{"exp_success", "expedition_success", "Expedition comes back well", "after an expedition or campaign succeeds", "expedition", "", areaFlavorExp},
	{"exp_fail", "expedition_fail", "Expedition comes back badly", "after an expedition or campaign fails", "expedition", "", areaFlavorExp},
	{"enc_standoff", "encounter_standoff", "Standoff with a civilization at war with you", "when an expedition runs into a civilization you are at war with and nobody fights", "civilization", "bronze_age", areaFlavorEnc},
	{"enc_cap", "encounter_full", "A boon you have no room for", "when an expedition meets a civilization while you already hold as many boons as you can", "civilization", "bronze_age", areaFlavorEnc},
	{"war_raid", "raid", "Raid", "when a civilization at war with you raids", "civilization", "bronze_age", areaFlavorRaid},
	{"harb_arrival", "harbinger_arrival", "A harbinger arrives", "when the age's harbinger arrives", "risk", "", areaFlavorHarbIn},
	{"harb_warn", "harbinger_warning", "The harbinger's warning", "when the harbinger gives the warning", "risk", "", areaFlavorHarbIn},
	{"harb_appeased", "harbinger_appeased", "After you appease", "after you pay to appease the harbinger", "risk", "iron_age", areaFlavorHarbDo},
	{"harb_braced", "harbinger_braced", "After you brace", "after you pay to brace against the doom", "risk", "iron_age", areaFlavorHarbDo},
	{"harb_invited", "harbinger_invited", "After you invite the doom", "after you invite the doom", "risk", "iron_age", areaFlavorHarbDo},
	{"harb_vindicated", "harbinger_vindicated", "The warning comes true", "when the harbinger warned and the catastrophe came", "risk", "iron_age", areaFlavorHarbOut},
	{"harb_spared", "harbinger_spared", "The warning comes to nothing", "when the harbinger warned and the catastrophe did not come", "risk", "iron_age", areaFlavorHarbOut},
	{"harb_discredited", "harbinger_discredited", "A false prophet is found out", "when a harbinger turns out to have been a false prophet", "risk", "", areaFlavorHarbOut},
	{"harb_fulfilled", "harbinger_fulfilled", "An invited doom arrives", "when a doom you invited arrives", "risk", "iron_age", areaFlavorHarbOut},
	{"run_end", "run_ending", "The end of a run", "at every prestige, as the run ends", "", "medieval_age", areaFlavorEnd},
}

// The flavor package sorts the ages into five spans, and a pool fits one or
// more of them (flavor/render.go, eraOf). First and last age of each, by
// place in the age order.
var flavorSpans = map[string][2]int{
	"eraAncient":    {1, 5},
	"eraFeudal":     {6, 8},
	"eraIndustrial": {9, 13},
	"eraDigital":    {14, 17},
	"eraCosmic":     {18, 22},
}

// flavorPool is one pool of sentences.
type flavorPool struct {
	m      *moment
	bucket string   // the pool's part of an id ("any", "ancient", "stone_age")
	spans  []string // the spans it fits (flavorSpans keys); all five for an ungated pool
	ages   []string // the age keys it is held to, for a per-age pool
	size   int
	tmpl   string // what the flavor package's own ids for its sentences start with
}

type flavorReader struct {
	m       *module
	pools   map[string]*flavorPool // by the variable that holds the sentences
	setPool map[string]*moment     // by the variable that holds per-age sets (agePools)
	setPre  map[string]string      // that variable's pool-name prefix after the moment's
	spans   map[string][]string    // erasAny, erasEarly ... -> spans
	lists   map[string][]string    // named string lists (tLow, runEndPlainAges)
	bankUse map[string][]*row      // word list -> the sentences that draw on it
	bankLen map[string]int         // word list -> its longest entry
	bankN   map[string]int
	made    map[string]*flavorPool // per-age pools, by id prefix
	unread  []string               // pools no moment claims
	byTmpl  map[string]*row        // sentences by the flavor package's own id
	order   []*flavorPool          // pools in the order they are declared
}

func newFlavorReader(m *module) *flavorReader {
	fr := &flavorReader{
		m: m, pools: map[string]*flavorPool{}, setPool: map[string]*moment{}, setPre: map[string]string{},
		spans: map[string][]string{}, lists: map[string][]string{},
		bankUse: map[string][]*row{}, bankLen: map[string]int{}, bankN: map[string]int{},
		made: map[string]*flavorPool{}, byTmpl: map[string]*row{},
	}
	var files []*srcFile
	for _, f := range m.files {
		if f.mode == modeFlavor {
			files = append(files, f)
		}
	}
	// Named lists first: the era shorthands, the tier sets and the age lists.
	for _, f := range files {
		for _, d := range f.ast.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, sp := range gd.Specs {
				vs, ok := sp.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, n := range vs.Names {
					if i >= len(vs.Values) {
						if exprText(vs.Type) == "[]era" {
							fr.spans[n.Name] = nil // declared empty: every span
						}
						continue
					}
					if cl, ok := vs.Values[i].(*ast.CompositeLit); ok && exprText(cl.Type) == "[]era" {
						for _, el := range cl.Elts {
							fr.spans[n.Name] = append(fr.spans[n.Name], exprText(el))
						}
						continue
					}
					if l := fr.stringList(f, vs.Values[i]); l != nil {
						fr.lists[n.Name] = l
					}
				}
			}
		}
	}
	// Then the pool declarations.
	for _, f := range files {
		ast.Inspect(f.ast, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			id, ok := call.Fun.(*ast.Ident)
			if !ok || len(call.Args) < 2 {
				return true
			}
			prefix, isLit := stringLit(call.Args[0])
			if !isLit {
				return true
			}
			mo, bucket := momentOf(prefix)
			switch id.Name {
			case "pool", "agePool":
				if len(call.Args) != 3 {
					return true
				}
				v := exprText(call.Args[2])
				if mo == nil {
					fr.unread = append(fr.unread, prefix)
					return true
				}
				p := &flavorPool{m: mo, bucket: bucket, tmpl: prefix}
				fr.order = append(fr.order, p)
				if id.Name == "pool" {
					p.spans = fr.spans[exprText(call.Args[1])]
				} else {
					p.ages = fr.stringList(f, call.Args[1])
					p.spans = spansOfAges(p.ages)
				}
				fr.pools[v] = p
			case "agePools":
				if mo == nil {
					fr.unread = append(fr.unread, prefix)
					return true
				}
				fr.setPool[exprText(call.Args[1])] = mo
				fr.setPre[exprText(call.Args[1])] = prefix
			}
			return true
		})
	}
	// Then which sentences use which word list, and how long its entries run.
	for _, f := range files {
		for _, u := range m.units(f) {
			if name, ok := bankOf(u); ok {
				fr.bankN[name]++
				if n := len([]rune(u.text)); n > fr.bankLen[name] {
					fr.bankLen[name] = n
				}
			}
			if s := u.st(); s.typ == "flavor.skel" && s.field == "Text" {
				if p := fr.poolOf(u); p != nil {
					p.size++
				}
			}
		}
	}
	return fr
}

// momentOf finds the catalog a pool name belongs to and what is left of the
// name.
func momentOf(prefix string) (*moment, string) {
	var best *moment
	for _, mo := range moments {
		if (prefix == mo.prefix || strings.HasPrefix(prefix, mo.prefix+"_")) && (best == nil || len(mo.prefix) > len(best.prefix)) {
			best = mo
		}
	}
	if best == nil {
		return nil, ""
	}
	return best, strings.TrimPrefix(strings.TrimPrefix(prefix, best.prefix), "_")
}

// spansOfAges returns the spans a list of age keys falls in.
func spansOfAges(ages []string) []string {
	w := sharedWorld()
	seen := map[string]bool{}
	var out []string
	for _, name := range []string{"eraAncient", "eraFeudal", "eraIndustrial", "eraDigital", "eraCosmic"} {
		sp := flavorSpans[name]
		for _, a := range ages {
			if i := w.ageIdx[a]; i >= sp[0] && i <= sp[1] && !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
		}
	}
	return out
}

var theWorld *world

func sharedWorld() *world {
	if theWorld == nil {
		theWorld = newWorld()
	}
	return theWorld
}

// stringList resolves an expression to a list of strings: a []string
// literal, an ages(...) call, or a variable that holds one.
func (fr *flavorReader) stringList(f *srcFile, e ast.Expr) []string {
	switch x := e.(type) {
	case *ast.CompositeLit:
		if exprText(x.Type) != "[]string" && x.Type != nil {
			return nil
		}
		var out []string
		for _, el := range x.Elts {
			out = append(out, fr.m.keyText(f.pkg, el))
		}
		return out
	case *ast.CallExpr:
		if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "ages" {
			var out []string
			for _, a := range x.Args {
				out = append(out, fr.m.keyText(f.pkg, a))
			}
			return out
		}
	case *ast.Ident:
		return fr.lists[x.Name]
	}
	return nil
}

// bankOf reports whether a unit is an entry of a word list, and which.
func bankOf(u *unit) (string, bool) {
	if len(u.structs) > 0 || !strings.HasSuffix(u.decl, "Banks") || len(u.coll) != 2 || !u.coll[1].isMap {
		return "", false
	}
	return u.coll[1].key, true
}

// poolOf finds the pool a sentence belongs to.
func (fr *flavorReader) poolOf(u *unit) *flavorPool {
	if p, ok := fr.pools[u.decl]; ok && len(u.structs) == 1 {
		return p
	}
	mo, ok := fr.setPool[u.decl]
	if !ok || len(u.structs) != 2 {
		return nil
	}
	// A per-age set: its ages are the set's first field.
	set := u.structs[1]
	var agesExpr ast.Expr
	for i, el := range set.lit.Elts {
		if kv, ok := el.(*ast.KeyValueExpr); ok {
			if exprText(kv.Key) == "ages" {
				agesExpr = kv.Value
			}
		} else if i == 0 {
			agesExpr = el
		}
	}
	ages := fr.stringList(u.f, agesExpr)
	if len(ages) == 0 {
		return nil
	}
	key := u.decl + "/" + ages[0]
	if p, ok := fr.made[key]; ok {
		return p
	}
	p := &flavorPool{m: mo, bucket: ages[0], ages: ages, spans: spansOfAges(ages), tmpl: fr.setPre[u.decl] + "_" + ages[0]}
	fr.made[key] = p
	fr.order = append(fr.order, p)
	return p
}

// read makes the row for a unit of the flavor package, or says why not. A
// unit that is neither a catalog sentence nor a word-list entry is left to
// the code reader ("" for why).
func (fr *flavorReader) read(c *catalog, u *unit) (*row, string) {
	w := c.w
	if name, ok := bankOf(u); ok {
		r := &row{
			id:    fmt.Sprintf("flavor.words.%s.%03d", name, u.coll[0].index+1),
			kind:  kindVoice,
			area:  areaFlavorExp,
			rules: rulesBank,
			pool:  fr.bankN[name],
			seen:  seenPool,
			age:   -1,
		}
		// Filled in by bankRows once the sentences that use it are known.
		r.where = name
		return r, ""
	}
	s := u.st()
	if s.typ != "flavor.skel" {
		return nil, ""
	}
	if s.field != "Text" {
		return nil, "a setting of a flavor sentence, not text"
	}
	p := fr.poolOf(u)
	if p == nil {
		return nil, ""
	}
	r := &row{
		id:    fmt.Sprintf("flavor.%s.%s.%03d", p.m.key, p.bucket, s.coll[0].index+1),
		kind:  kindVoice,
		area:  p.m.area,
		rules: rulesSkel,
		pool:  p.size,
		seen:  seenPool,
		eras:  p.spans,
	}
	fr.byTmpl[fmt.Sprintf("%s_%d", p.tmpl, s.coll[0].index)] = r
	lo, hi := 1, len(w.ageKeys)
	if len(p.ages) > 0 {
		lo, hi = len(w.ageKeys), 1
		for _, a := range p.ages {
			if i := w.ageIdx[a]; i > 0 {
				lo, hi = min(lo, i), max(hi, i)
			}
		}
	} else if len(p.spans) > 0 {
		lo, hi = len(w.ageKeys), 1
		for _, sp := range p.spans {
			lo, hi = min(lo, flavorSpans[sp][0]), max(hi, flavorSpans[sp][1])
		}
	}
	from := lo
	if i := w.ageIdx[p.m.from]; i > from {
		from = i
	}
	r.age = from
	var sb strings.Builder
	sb.WriteString("A line of story the log may add " + p.m.when)
	switch {
	case len(p.ages) > 0 && contiguous(w, p.ages):
		sb.WriteString(", in " + w.agesText(lo, hi) + ".")
	case len(p.ages) > 0:
		var names []string
		for _, a := range p.ages {
			names = append(names, "the "+w.called(a))
		}
		sb.WriteString(", in " + joinAnd(names) + ".")
	case lo == 1 && hi == len(w.ageKeys):
		sb.WriteString(", in any age.")
	default:
		sb.WriteString(", in " + w.agesText(lo, hi) + ".")
	}
	if kinds := fr.fieldList(u, s, "Kinds"); len(kinds) > 0 {
		sb.WriteString(" " + kindSentence(p.m, kinds))
	}
	if tones := fr.fieldList(u, s, "Tones"); len(tones) > 0 {
		sb.WriteString(" Used only for a " + strings.ToLower(strings.Join(tones, " or ")) + " telling.")
	}
	r.max = c.rules.maxLine - 1 // the game adds the full stop
	if slot := s.sib["Slot"]; slot != "" {
		sb.WriteString(fmt.Sprintf(" The ~ is replaced by one entry of the word list %s (%d entries, the rows flavor.words.%s.*).", slot, fr.bankN[slot], slot))
		r.max = c.rules.maxLine - fr.bankLen[slot] // the longest entry may be the one drawn
		fr.bankUse[slot] = append(fr.bankUse[slot], r)
	}
	if strings.Contains(u.text, "{") {
		sb.WriteString(" " + slotSentence(u.text))
	}
	sb.WriteString(" The game adds the full stop.")
	if unanswerable(p) {
		// A false prophet can come in the first era, but the game refuses
		// every answer to one there.
		r.unsure = true
		sb.WriteString(" No catastrophe can strike in the first era, so the game refuses to appease, brace or invite there, and as things stand this line is never reached.")
	}
	r.where = sb.String()
	return r, ""
}

// unanswerable reports whether a pool holds the lines that follow an answer
// to a harbinger (appease, brace, invite) in ages where the game allows no
// answer: the ages of an era in which no catastrophe may strike.
func unanswerable(p *flavorPool) bool {
	switch p.m.prefix {
	case "harb_appeased", "harb_braced", "harb_invited":
	default:
		return false
	}
	if len(p.ages) == 0 {
		return false
	}
	for _, a := range p.ages {
		if config.CatastropheAllowed(config.EpochForAge(a)) {
			return false
		}
	}
	return true
}

// fieldList returns the values of a list field of a sentence (Kinds, Tones).
func (fr *flavorReader) fieldList(u *unit, s structSite, field string) []string {
	for _, el := range s.lit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok || exprText(kv.Key) != field {
			continue
		}
		if cl, ok := kv.Value.(*ast.CompositeLit); ok {
			var out []string
			for _, e := range cl.Elts {
				out = append(out, fr.m.keyText(u.f.pkg, e))
			}
			return out
		}
		return fr.stringList(u.f, kv.Value)
	}
	return nil
}

// kindSentence says in plain words what a Kinds restriction means.
func kindSentence(mo *moment, kinds []string) string {
	plain := make([]string, len(kinds))
	for i, k := range kinds {
		plain[i] = strings.ReplaceAll(k, "_", " ")
	}
	list := strings.Join(plain, " or ")
	switch mo.about {
	case "expedition":
		if list == "military" {
			return "Only for a military campaign."
		}
		return "Only for a " + list + " expedition."
	case "civilization":
		return "Only when the other civilization is " + list + "."
	case "risk":
		if list == "false prophet" {
			return "Only when the harbinger was a false prophet."
		}
		return "Only when the risk the harbinger sees is " + list + "."
	}
	return "Only for: " + list + "."
}

// slotSentence explains the placeholders a sentence carries.
func slotSentence(text string) string {
	names := map[string]string{
		"{subject}":    "{subject} is the name of the expedition, civilization or harbinger",
		"{res}":        "{res} is a resource's name",
		"{res_stores}": "{res_stores} reads “your food stores”",
		"{res_haul}":   "{res_haul} reads “a haul of food”",
		"{amt_res}":    "{amt_res} is an amount and its resource (“40 food”)",
		"{amt}":        "{amt} is an amount",
		"{n}":          "{n} is a count",
		"{ticks}":      "{ticks} is a number of ticks",
	}
	var parts []string
	seen := map[string]bool{}
	for _, m := range slotRe.FindAllString(text, -1) {
		if seen[m] {
			continue
		}
		seen[m] = true
		if n, ok := names[m]; ok {
			parts = append(parts, n)
		} else {
			parts = append(parts, m+" is filled in by the game")
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "In it, " + strings.Join(parts, "; ") + "."
}

// contiguous reports whether a list of ages is an unbroken run of the order.
func contiguous(w *world, ages []string) bool {
	idx := make([]int, 0, len(ages))
	for _, a := range ages {
		idx = append(idx, w.ageIdx[a])
	}
	sort.Ints(idx)
	for i := 1; i < len(idx); i++ {
		if idx[i] != idx[i-1]+1 {
			return false
		}
	}
	return true
}

func joinAnd(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
}

// bankRows finishes the word-list rows once every sentence is read: where
// each entry goes, the ages it can turn up in, and its sheet.
func (fr *flavorReader) bankRows(c *catalog) {
	for _, r := range c.rows {
		if r.rules != rulesBank {
			continue
		}
		name := r.where
		users := fr.bankUse[name]
		if len(users) == 0 {
			r.where = fmt.Sprintf("An entry of the word list %s, which no sentence draws on at the moment.", name)
			r.unsure = true
			r.area = areaUnsure
			continue
		}
		first := users[0]
		r.area = first.area
		r.age = first.age
		// The entry and the longest sentence that takes it must fit the
		// line together.
		longest := 0
		for _, us := range users {
			longest = max(longest, len([]rune(us.current)))
		}
		r.max = c.rules.maxLine - longest
		seen := map[string]bool{}
		for _, us := range users {
			if us.age < r.age {
				r.age = us.age
			}
			for _, e := range us.eras {
				if !seen[e] {
					seen[e] = true
					r.eras = append(r.eras, e)
				}
			}
			if len(us.eras) == 0 {
				r.eras = nil
				seen = map[string]bool{"all": true}
			}
		}
		if seen["all"] {
			r.eras = nil
		}
		r.where = fmt.Sprintf("One of the %d entries of the word list %s. The game drops one entry into the ~ of %s, such as “%s” (%s). It lands in the middle of a sentence: no capital letter to start, no full stop to end.",
			r.pool, name, plural(len(users), "sentence", "sentences"), first.current, first.id)
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}
