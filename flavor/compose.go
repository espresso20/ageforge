package flavor

import (
	"strconv"
	"strings"
	"sync"
)

// compose puts the template together from the entries drawn for its slots (the
// index into each bank, in order). The placeholders are not filled in yet.
func (t tmpl) compose(draws []int) string {
	var sb strings.Builder
	n := 0
	for _, p := range t.Parts {
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
		sb.WriteString(bank[i])
	}
	return sb.String()
}

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

// Retell is the line a Key names, for req, in plain English. req fills the
// placeholders as Generate did (only the fields the line uses matter). It
// reports false for a key this catalog does not know (one from a save of
// another build, or one that is edited), and the caller keeps the plain line
// it has.
func Retell(key string, req Request) (string, bool) {
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
	return fillAs(t.compose(draws), req), true
}
