package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// ranks_model.go is what the Ranks screen shows, worked out from the badge
// views: a rank for each worker domain, and the lists behind it. Only badges
// in sight count anywhere on it; a silhouette has no domain.

// domainHex is each worker domain's colour, by place in config.WorkerDomains.
var domainHex = [...]string{
	"#8fbf5a", "#c9964a", "#b8b0a0", "#d9d27a", "#d9604f", "#e0a63c",
	"#6fa8dc", "#c08a5c", "#e6c34a", "#58c4a8", "#a88be0", "#7a95ff",
}

// rankDom is one worker domain's standing.
type rankDom struct {
	key, name, hex string
	opens          string // the age its workers arrive in
	pts, rank      int
	thr            [6]int // the points each rank starts at
	earned, all    int
	have, total    map[config.BadgeTier]int
	names          map[config.BadgeTier][]string // the earned badges' names, by tier
	near           []game.BadgeView              // not earned, nearest their goal first
}

// pct is how far a badge with a goal has come, 0 to 100.
func pct(v game.BadgeView) int {
	if v.Target <= 0 {
		return 0
	}
	return int(min(v.Progress/v.Target, 1) * 100)
}

// toNext is the points a domain lacks for its next rank; 0 at the top.
func (d *rankDom) toNext() int {
	if d.rank >= len(d.thr) {
		return 0
	}
	return d.thr[d.rank] - d.pts
}

// ranksModel is the Ranks screen's data.
type ranksModel struct {
	doms          []*rankDom // ranked first, then by points
	by            map[string]*rankDom
	views         []game.BadgeView
	earned, total int
	pts, held     int
	tierHave      map[config.BadgeTier]int // every badge in sight, by tier
	tierAll       map[config.BadgeTier]int
	titles        [6]string // the rank titles
	thr           [6]int
	now           time.Time // what "today" is for the recent list
}

func buildRanks(set *rules.Set, views []game.BadgeView) *ranksModel {
	m := &ranksModel{views: views, by: map[string]*rankDom{}, now: time.Now(),
		titles: set.RankTitles(), thr: set.RankThresholds(),
		tierHave: map[config.BadgeTier]int{}, tierAll: map[config.BadgeTier]int{}}
	for i, key := range set.WorkerDomains() {
		d := &rankDom{key: key, thr: m.thr, opens: set.DomainOpens(key), name: game.DomainName(key), hex: domainHex[i%len(domainHex)],
			have: map[config.BadgeTier]int{}, total: map[config.BadgeTier]int{},
			names: map[config.BadgeTier][]string{}}
		m.doms = append(m.doms, d)
		m.by[key] = d
	}
	for _, v := range views {
		if v.Integrity {
			continue
		}
		m.total++
		m.tierAll[v.Level]++
		if v.Earned {
			m.earned++
			m.tierHave[v.Level]++
		}
		d := m.by[v.Domain]
		if d == nil {
			continue
		}
		d.all++
		d.total[v.Level]++
		if v.Earned {
			d.earned++
			d.have[v.Level]++
			d.names[v.Level] = append(d.names[v.Level], v.Name)
			d.pts += v.Points
		} else if v.Target > 0 {
			d.near = append(d.near, v)
		}
	}
	for _, d := range m.doms {
		d.rank = config.RankFor(d.pts)
		m.pts += d.pts
		m.held += d.rank
		sort.SliceStable(d.near, func(i, j int) bool {
			return d.near[i].Progress/d.near[i].Target > d.near[j].Progress/d.near[j].Target
		})
	}
	// Ranked domains first, each group in the order of the worker domains.
	sort.SliceStable(m.doms, func(i, j int) bool { return m.doms[i].rank > 0 && m.doms[j].rank == 0 })
	return m
}

// list is the badges the all-badges tab shows: every badge in sight, by
// domain, or only one domain's.
func (m *ranksModel) list(dom string) []game.BadgeView {
	var out []game.BadgeView
	for _, d := range m.doms {
		if dom != "" && d.key != dom {
			continue
		}
		for _, v := range m.views {
			if v.Domain == d.key {
				out = append(out, v)
			}
		}
	}
	if dom == "" {
		for _, v := range m.views {
			if v.Domain == "" && !v.Integrity {
				out = append(out, v)
			}
		}
	}
	return out
}

// recent is the earned badges, the latest first; the ones with no date last.
func (m *ranksModel) recent() []game.BadgeView {
	var out []game.BadgeView
	for _, v := range m.views {
		if v.Earned && !v.Integrity {
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out
}

// findDomain reads a word as a worker domain by its key.
func findDomain(set *rules.Set, word string) (string, bool) {
	word = strings.ToLower(strings.TrimSpace(word))
	for _, k := range set.WorkerDomains() {
		if k == word {
			return k, true
		}
	}
	return "", false
}

// closest is the ranked domains, nearest their next rank first.
func (m *ranksModel) closest() []*rankDom {
	var out []*rankDom
	for _, d := range m.doms {
		if d.rank > 0 && d.toNext() > 0 {
			out = append(out, d)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].toNext() < out[j].toNext() })
	return out
}

// top is the ranked domain with the most points; nil with none ranked.
func (m *ranksModel) top() *rankDom {
	var best *rankDom
	for _, d := range m.doms {
		if d.rank > 0 && (best == nil || d.pts > best.pts) {
			best = d
		}
	}
	return best
}

// when is a badge's date as the recent list says it, from the model's today.
func (m *ranksModel) when(at time.Time) string {
	if at.IsZero() {
		return "earlier"
	}
	day := func(t time.Time) time.Time {
		y, mo, d := t.Date()
		return time.Date(y, mo, d, 0, 0, 0, 0, time.UTC)
	}
	n := int(day(m.now).Sub(day(at)).Hours() / 24)
	switch {
	case n <= 0:
		return "today"
	case n == 1:
		return "yesterday"
	case n < 7:
		return fmt.Sprintf("%d days ago", n)
	case n < 14:
		return "last week"
	}
	return at.Format("2 Jan 2006")
}
