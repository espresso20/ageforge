package flavor

import (
	"strconv"
	"strings"
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
