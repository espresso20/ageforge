package ui

import (
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
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
	earned, all    int
	have, total    map[config.BadgeTier]int
	near           []game.BadgeView // not earned, nearest their goal first
}

// toNext is the points a domain lacks for its next rank; 0 at the top.
func (d *rankDom) toNext() int {
	if d.rank >= len(config.RankThresholds) {
		return 0
	}
	return config.RankThresholds[d.rank] - d.pts
}

// ranksModel is the Ranks screen's data.
type ranksModel struct {
	doms          []*rankDom // ranked first, then by points
	by            map[string]*rankDom
	views         []game.BadgeView
	earned, total int
	pts, held     int
}

func buildRanks(views []game.BadgeView) *ranksModel {
	m := &ranksModel{views: views, by: map[string]*rankDom{}}
	ages := config.AgeByKey()
	for i, key := range config.WorkerDomains() {
		d := &rankDom{key: key, name: game.DomainName(key), hex: domainHex[i%len(domainHex)],
			have: map[config.BadgeTier]int{}, total: map[config.BadgeTier]int{}}
		for _, w := range config.WorkerClasses() {
			if w.Domain == key {
				d.opens = ages[w.AgeKey].Name
				break
			}
		}
		m.doms = append(m.doms, d)
		m.by[key] = d
	}
	for _, v := range views {
		if v.Integrity {
			continue
		}
		m.total++
		if v.Earned {
			m.earned++
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
	sort.SliceStable(m.doms, func(i, j int) bool {
		if a, b := m.doms[i], m.doms[j]; a.rank != b.rank {
			return a.rank > b.rank
		} else if a.pts != b.pts {
			return a.pts > b.pts
		}
		return false
	})
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
func findDomain(word string) (string, bool) {
	word = strings.ToLower(strings.TrimSpace(word))
	for _, k := range config.WorkerDomains() {
		if k == word {
			return k, true
		}
	}
	return "", false
}
