package flavor

import (
	"fmt"
	"sort"
	"strings"
)

// rowid.go: the row ids of the flavor texts, the ids the copy export gives the
// same sentences and word-list entries (cmd/words reads the catalogs from the
// source).

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

// Row is one text of the catalog that is told through the script: its row id
// and the text in the game's own English.
type Row struct {
	ID, Plain string
}

// catalogRows lists every text the package reads: each sentence of each
// catalog, then each entry of each word list, with its row id and its plain
// text. The copy export is checked against it (the ids and the texts must be
// the same).
func catalogRows() []Row {
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
