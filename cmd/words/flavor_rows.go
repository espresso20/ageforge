package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
)

// This file reads the flavor catalogs: the sentences each catalog can say,
// grouped into pools by the ages they fit, and the word lists some of those
// sentences draw a phrase from.
//
// The catalogs are the text files of flavor/text. The flavor package reads
// them and says, for each sentence and word, where it is written and what
// pool it is in; a sentence's id is its catalog, its pool and its place in the
// pool.

// moment is one flavor catalog: one game situation the package narrates.
type moment struct {
	prefix string // what its pool names start with ("exp_fail")
	key    string // its part of an id ("expedition_fail")
	title  string // what the README calls it
	when   string // "after an expedition or campaign fails", for where sentences
	about  string // for a kind restriction: what the kind describes
	from   string // the age key before which it cannot happen, "" for the first age
	area   *area
}

// moments are the catalogs, in the order of the guide. momentOf picks the
// longest prefix that matches a pool name, so one prefix may start another.
// The read fails when a pool matches none.
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

// flavorSpans gives the first and last age of each of the flavor package's
// five spans (flavor/render.go, eraOf), by place in the age order from 1. A
// pool fits one or more of them.
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
	spans  []string // the spans it fits (flavorSpans keys); empty for a pool open to every age
	ages   []string // the age keys it is held to, for a per-age pool
	size   int
	tmpl   string // what the flavor package's own ids for its sentences start with
}

type flavorReader struct {
	pools   map[string]*flavorPool // by the pool's name
	order   []*flavorPool          // pools in the order the files list them
	bankUse map[string][]*row      // word list -> the sentences that draw on it
	bankLen map[string]int         // word list -> its longest entry
	bankN   map[string]int         // word list -> its number of entries
	byTmpl  map[string]*row        // sentences by the flavor package's own id
}

func newFlavorReader() *flavorReader {
	return &flavorReader{
		pools:   map[string]*flavorPool{},
		bankUse: map[string][]*row{},
		bankLen: map[string]int{},
		bankN:   map[string]int{},
		byTmpl:  map[string]*row{},
	}
}

// flavorFolder is where the catalogs are written, below the module root.
const flavorFolder = "flavor/text"

// read makes the rows of every text in the catalog files, in the order the
// files list them. A word-list row says only its list until bankRows has seen
// the sentences that use it.
func (fr *flavorReader) read(c *catalog) ([]*row, error) {
	folder := filepath.Join(c.m.root, filepath.FromSlash(flavorFolder))
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return nil, nil // a partial tree, as the tests use
	}
	entries, err := flavor.ReadEntries(os.DirFS(folder))
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.Bank {
			fr.bankN[e.Pool] = e.Size
			fr.bankLen[e.Pool] = max(fr.bankLen[e.Pool], utf8.RuneCountInString(e.Text))
		}
	}
	rows := make([]*row, 0, len(entries))
	for _, e := range entries {
		file, err := c.m.textFile(path.Join(flavorFolder, e.File))
		if err != nil {
			return nil, err
		}
		var r *row
		if e.Bank {
			r = bankEntryRow(e)
		} else if r, err = fr.sentenceRow(c, e); err != nil {
			return nil, err
		}
		r.u = textUnit(file, e.Line, e.From, e.To)
		r.current = e.Text
		if r.unsure {
			r.area = areaUnsure
		}
		rows = append(rows, r)
	}
	return rows, nil
}

// momentOf finds the catalog a pool name belongs to.
func momentOf(pool string) *moment {
	var best *moment
	for _, mo := range moments {
		if (pool == mo.prefix || strings.HasPrefix(pool, mo.prefix+"_")) && (best == nil || len(mo.prefix) > len(best.prefix)) {
			best = mo
		}
	}
	return best
}

// spansOfEras gives the flavorSpans keys of the eras a pool is held to.
func spansOfEras(eras []string) []string {
	var spans []string
	for _, name := range eras {
		spans = append(spans, "era"+strings.ToUpper(name[:1])+name[1:])
	}
	return spans
}

// poolOf finds the pool an entry belongs to, and makes it the first time.
func (fr *flavorReader) poolOf(e flavor.Entry) (*flavorPool, error) {
	if p, ok := fr.pools[e.Pool]; ok {
		return p, nil
	}
	mo := momentOf(e.Pool)
	if mo == nil || !strings.HasPrefix(e.ID, "flavor."+mo.key+".") {
		return nil, fmt.Errorf("the flavor pool %s matches no catalog in moments (cmd/words/flavor_rows.go)", e.Pool)
	}
	p := &flavorPool{m: mo, bucket: e.Bucket, spans: spansOfEras(e.Eras), ages: e.Ages, size: e.Size, tmpl: e.Pool}
	fr.pools[e.Pool] = p
	fr.order = append(fr.order, p)
	return p, nil
}

// bankEntryRow makes the row for an entry of a word list.
func bankEntryRow(e flavor.Entry) *row {
	return &row{
		id:    e.ID,
		kind:  kindVoice,
		area:  areaFlavorExp,
		rules: rulesBank,
		pool:  e.Size,
		seen:  seenPool,
		age:   -1,
		where: e.Pool,
	}
}

// sentenceRow makes the row for a catalog sentence.
func (fr *flavorReader) sentenceRow(c *catalog, e flavor.Entry) (*row, error) {
	p, err := fr.poolOf(e)
	if err != nil {
		return nil, err
	}
	w := c.w
	r := &row{
		id:    e.ID,
		kind:  kindVoice,
		area:  p.m.area,
		rules: rulesSkel,
		pool:  p.size,
		seen:  seenPool,
		eras:  p.spans,
	}
	fr.byTmpl[fmt.Sprintf("%s_%d", p.tmpl, e.Index)] = r
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
	if len(e.Kinds) > 0 {
		sb.WriteString(" " + kindSentence(p.m, e.Kinds))
	}
	if len(e.Tones) > 0 {
		sb.WriteString(" Used only for a " + strings.Join(e.Tones, " or ") + " telling.")
	}
	r.max = c.rules.maxLine - 1 // the game adds the full stop
	if e.Slot != "" {
		sb.WriteString(fmt.Sprintf(" The ~ is replaced by one entry of the word list %s (%d entries, the rows flavor.words.%s.*).", e.Slot, fr.bankN[e.Slot], e.Slot))
		r.max = c.rules.maxLine - fr.bankLen[e.Slot] // the longest entry may be the one drawn
		fr.bankUse[e.Slot] = append(fr.bankUse[e.Slot], r)
	}
	if strings.Contains(e.Text, "{") {
		sb.WriteString(" " + slotSentence(e.Text))
	}
	sb.WriteString(" The game adds the full stop.")
	if unanswerable(p) {
		// A false prophet can come in the first era (the Stone Era), but the game
		// refuses every answer to one there.
		r.unsure = true
		sb.WriteString(" No catastrophe can strike in the first era, so the game refuses to appease, brace or invite there, and as things stand this line is never reached.")
	}
	r.where = sb.String()
	return r, nil
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

var theWorld *world

func sharedWorld() *world {
	if theWorld == nil {
		theWorld = newWorld()
	}
	return theWorld
}
