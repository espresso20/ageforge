package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

// The wiki's tech tables (site/docs/technologies.md, "Tech Tree by Age") are
// written from config: one section an age, a row a tech, in table order.
// TestTechWikiTables fails when a section drifts from what config says, and
// UPDATE_WIKI=1 rewrites them:
//
//	UPDATE_WIKI=1 go test ./config -run TestTechWikiTables
const techWikiPage = "../site/docs/technologies.md"

// techWikiSpan is how long an age's techs take, for its heading: seconds
// under ten minutes, whole minutes after ("2m 16s", "46m", "1h 10m").
func techWikiSpan(ticks int) string {
	secs := int(float64(ticks) * TickSeconds)
	switch {
	case secs < 60:
		return fmt.Sprintf("%ds", secs)
	case secs < 600:
		if secs%60 == 0 {
			return fmt.Sprintf("%dm", secs/60)
		}
		return fmt.Sprintf("%dm %ds", secs/60, secs%60)
	case secs < 3600:
		return fmt.Sprintf("%dm", secs/60)
	}
	if m := secs / 60 % 60; m > 0 {
		return fmt.Sprintf("%dh %dm", secs/3600, m)
	}
	return fmt.Sprintf("%dh", secs/3600)
}

// techWikiSections is every age's section as the wiki should hold it, by
// age name: the heading with the span of its techs' research times, the
// table's head, and a row for each tech.
func techWikiSections() map[string]string {
	techs := Technologies()
	blds := BaseBuildings()
	kinds := TechKinds(techs, blds)
	wonderOf := map[string]string{}
	opens := map[string][]string{}
	for _, b := range blds {
		switch {
		case b.RequiredTech == "":
		case b.Category == "wonder":
			wonderOf[b.RequiredTech] = b.Name
		default:
			opens[b.RequiredTech] = append(opens[b.RequiredTech], "the "+b.Name)
		}
	}
	comma := func(n int) string {
		s := fmt.Sprint(n)
		for i := len(s) - 3; i > 0; i -= 3 {
			s = s[:i] + "," + s[i:]
		}
		return s
	}
	quote := func(keys []string) []string {
		out := make([]string, len(keys))
		for i, k := range keys {
			out[i] = "`" + k + "`"
		}
		return out
	}
	out := map[string]string{}
	ages := Ages()
	for i, age := range ages {
		lo, hi := 0, 0
		var rows []string
		for _, tc := range techs {
			if tc.Age != age.Key {
				continue
			}
			if lo == 0 || tc.ResearchTicks < lo {
				lo = tc.ResearchTicks
			}
			hi = max(hi, tc.ResearchTicks)
			kind := string(kinds[tc.Key])
			if w := wonderOf[tc.Key]; w != "" {
				kind = "**keystone** (" + w + ")"
			}
			needs := strings.Join(quote(tc.Prerequisites), ", ")
			if len(tc.AnyOf) > 0 {
				if needs != "" {
					needs += ", and "
				}
				needs += "one of " + strings.Join(quote(tc.AnyOf), " or ")
			}
			if needs == "" {
				needs = "none"
			}
			var eff []string
			for _, l := range FeaturesOpenedBy(tc.Key) {
				eff = append(eff, l.Opens)
			}
			for _, o := range opens[tc.Key] {
				eff = append(eff, "opens "+o)
			}
			for _, e := range tc.Effects {
				eff = append(eff, e.Text())
			}
			text := strings.Join(eff, ", ")
			if text != "" {
				text = strings.ToUpper(text[:1]) + text[1:]
			}
			rows = append(rows, fmt.Sprintf("| `%s` | %s | %s | %s kp | %s | %s | %s |", tc.Key, tc.Name, kind, FormatRateValue(tc.Cost), comma(tc.ResearchTicks), needs, text))
		}
		if len(rows) == 0 {
			continue
		}
		span := "~" + techWikiSpan(lo)
		if hi != lo {
			span += " to " + techWikiSpan(hi)
		}
		if i < len(ages)-1 {
			span += "/tech"
		}
		out[age.Name] = fmt.Sprintf("### %s (%s)\n\n| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |\n|---|---|---|---|---|---|---|\n%s\n", age.Name, span, strings.Join(rows, "\n"))
	}
	return out
}

// techWikiSection finds an age's section in the page: its heading, the
// table's head and its rows.
func techWikiSection(page, age string) (start, end int, ok bool) {
	re := regexp.MustCompile(`(?m)^### ` + regexp.QuoteMeta(age) + ` \([^)]*\)\n\n\| Key \|[^\n]*\n\|---[^\n]*\n(?:\| ` + "`" + `[^\n]*\n)+`)
	loc := re.FindStringIndex(page)
	if loc == nil {
		return 0, 0, false
	}
	return loc[0], loc[1], true
}

func TestTechWikiTables(t *testing.T) {
	raw, err := os.ReadFile(techWikiPage)
	if err != nil {
		t.Fatal(err)
	}
	page := string(raw)
	want := techWikiSections()
	if len(want) != len(Ages()) {
		t.Fatalf("%d ages have techs, of %d", len(want), len(Ages()))
	}
	update := os.Getenv("UPDATE_WIKI") != ""
	drifted := 0
	for _, age := range Ages() {
		a, z, ok := techWikiSection(page, age.Name)
		if !ok {
			t.Errorf("the wiki has no tech table for the %s", age.Name)
			continue
		}
		if page[a:z] == want[age.Name] {
			continue
		}
		drifted++
		if update {
			page = page[:a] + want[age.Name] + page[z:]
			continue
		}
		got, exp := strings.Split(page[a:z], "\n"), strings.Split(want[age.Name], "\n")
		for i := 0; i < len(got) || i < len(exp); i++ {
			g, e := "", ""
			if i < len(got) {
				g = got[i]
			}
			if i < len(exp) {
				e = exp[i]
			}
			if g != e {
				t.Errorf("%s, line %d of its table:\n wiki   %s\n config %s", age.Name, i+1, g, e)
				break
			}
		}
	}
	if update && drifted > 0 {
		if err := os.WriteFile(techWikiPage, []byte(page), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("rewrote %d of the wiki's tech tables", drifted)
	} else if drifted > 0 {
		t.Errorf("%d of the wiki's tech tables have drifted from config: UPDATE_WIKI=1 go test ./config -run TestTechWikiTables", drifted)
	}
}
