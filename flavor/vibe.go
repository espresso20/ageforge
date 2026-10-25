package flavor

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/espresso20/ageforge/vibe"
)

// vibe.go: the stories under VibeMode.
//
// A story line is drawn from a catalog: a sentence (a skel) and, when the
// sentence has a slot, one entry of a word list. Both are read through the
// script at the one place a line is put together (tmpl.compose), by a row id
// each has by its place in the catalog:
//
//	flavor.<catalog>.<pool>.<n>     the n-th sentence of a pool, from 001
//	flavor.words.<list>.<n>         the n-th entry of a word list, from 001
//
// The ids are the ones the copy export gives the same texts (cmd/words reads
// the catalogs from the source and a test holds the two together), so the
// script's writer works from the export and the code needs no id in a catalog.
//
// Which sentence is drawn, which entry fills its slot, what the Stream
// remembers and what Result.Template names do not depend on the switch: the
// draw is made on the plain catalog and the script is read after it. A line
// that is kept (a harbinger's lines are part of its thread, saved with it)
// keeps its plain English and its Key, and is told again from the Key when it
// is read (Retell).

// rowParts splits a pool's name into its catalog's part of a row id and the
// pool's own ("exp_success_late" is expedition_success and late). Both are ""
// for a name no catalog claims.
func rowParts(prefix string) (key, bucket string) {
	for _, e := range momentTable {
		if prefix == e.prefix {
			return e.key, ""
		}
		if rest, ok := strings.CutPrefix(prefix, e.prefix+"_"); ok {
			return e.key, rest
		}
	}
	return "", ""
}

// sentenceRow is the row id of the n-th sentence (from 0) of a pool.
func sentenceRow(key, bucket string, n int) string {
	if key == "" {
		return ""
	}
	return "flavor." + key + "." + bucket + "." + place(n)
}

// wordRow is the row id of the n-th entry (from 0) of a word list.
func wordRow(list string, n int) string {
	return "flavor.words." + list + "." + place(n)
}

// place is the place of the n-th entry (from 0) in a row id: from 001.
func place(n int) string { return fmt.Sprintf("%03d", n+1) }

// sentence is the sentence a template was made from, as the export reads it:
// the slot as the slot mark and no full stop at the end (compose puts it back).
func (t tmpl) sentence() string {
	var sb strings.Builder
	for _, p := range t.Parts {
		if p.bank != "" {
			sb.WriteString(slotMark)
			continue
		}
		sb.WriteString(p.lit)
	}
	return strings.TrimSuffix(sb.String(), ".")
}

// toldParts are the template's parts in the script's words: the script's line
// for the sentence, cut at its slot. A sentence with no line in the script, or
// whose line cannot take the draw the template makes (it has no slot where the
// template has one, or has one where the template has none), is told as it was
// written.
func (t tmpl) toldParts() []part {
	if t.Row == "" {
		return t.Parts
	}
	plain := t.sentence()
	line := vibe.Line(t.Row, plain)
	if line == plain {
		return t.Parts
	}
	switch len(t.Parts) {
	case 1:
		if t.Parts[0].bank == "" && !strings.Contains(line, slotMark) {
			return []part{lit(line + ".")}
		}
	case 3:
		if t.Parts[0].bank == "" && t.Parts[1].bank != "" && t.Parts[2].bank == "" {
			if head, tail, ok := strings.Cut(line, slotMark); ok && !strings.Contains(tail, slotMark) {
				return []part{lit(head), t.Parts[1], lit(tail + ".")}
			}
		}
	}
	return t.Parts
}

// compose puts the template together from the entries drawn for its slots (the
// index into each bank, in order), as the plain catalog says it, or, when told
// is set, in the script's words. The placeholders are not filled in yet.
func (t tmpl) compose(draws []int, told bool) string {
	parts := t.Parts
	if told {
		parts = t.toldParts()
	}
	var sb strings.Builder
	n := 0
	for _, p := range parts {
		if p.bank == "" {
			sb.WriteString(p.lit)
			continue
		}
		bank := banks[p.bank]
		if len(bank) == 0 || n >= len(draws) {
			continue // an empty bank is a catalog bug; a test catches it
		}
		i := draws[n]
		n++
		word := bank[i]
		if told {
			word = vibe.Line(wordRow(p.bank, i), word)
		}
		sb.WriteString(word)
	}
	return sb.String()
}

// Key names the line a Result is, so it can be told again later in the words
// of whichever script is in force (Retell): the template and the entry drawn
// for each of its slots, such as "harb_warn_stone_age_12:7". It is "" for an
// empty Result.
func (r Result) Key() string {
	if r.Template == "" {
		return ""
	}
	if len(r.draws) == 0 {
		return r.Template
	}
	var sb strings.Builder
	sb.WriteString(r.Template)
	for i, d := range r.draws {
		if i == 0 {
			sb.WriteByte(':')
		} else {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.Itoa(d))
	}
	return sb.String()
}

// templateByID is every template of the catalog by its id.
var templateByID = sync.OnceValue(func() map[string]tmpl {
	out := map[string]tmpl{}
	for _, m := range Moments() {
		for _, t := range templatesFor(m) {
			out[t.ID] = t
		}
	}
	return out
})

// Retell is the line a Key names, for req: in the words of the script when
// told is set and VibeMode is on, and in plain English otherwise. req fills the
// placeholders as Generate did (only the fields the line uses matter). It
// reports false for a key this catalog does not know (one from a save of
// another build, or one that is edited), and the caller keeps the plain line
// it has. To be sure a key still names the line it was made for, tell it in
// plain English and compare.
func Retell(key string, req Request, told bool) (string, bool) {
	id, rest, hasDraws := strings.Cut(key, ":")
	t, ok := templateByID()[id]
	if !ok {
		return "", false
	}
	var draws []int
	if hasDraws {
		for _, s := range strings.Split(rest, ",") {
			d, err := strconv.Atoi(s)
			if err != nil || d < 0 {
				return "", false
			}
			draws = append(draws, d)
		}
	}
	slots := 0
	for _, p := range t.Parts {
		if p.bank == "" {
			continue
		}
		if slots >= len(draws) || draws[slots] >= len(banks[p.bank]) {
			return "", false
		}
		slots++
	}
	if slots != len(draws) {
		return "", false
	}
	told = told && vibe.On()
	return fillAs(t.compose(draws, told), req, told), true
}

// Row is one text of the catalog that is told through the script: its row id
// and the text in the game's own English.
type Row struct {
	ID, Plain string
}

// VibeRows lists every text the package reads through the script: each
// sentence of each catalog, then each entry of each word list. The copy export
// is checked against it (the ids and the texts must be the same), so a
// sentence cannot be told in the script's words without a row of the export.
func VibeRows() []Row {
	var rows []Row
	for _, m := range Moments() {
		for _, t := range templatesFor(m) {
			if t.Row != "" {
				rows = append(rows, Row{t.Row, t.sentence()})
			}
		}
	}
	names := make([]string, 0, len(banks))
	for name := range banks {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		for i, word := range banks[name] {
			rows = append(rows, Row{wordRow(name, i), word})
		}
	}
	return rows
}
