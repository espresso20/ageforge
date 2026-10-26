package mapmodel

import (
	"cmp"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// recap.go is the "while you were away" digest. A Visit is a small snapshot
// of what the player last saw; the model diffs the live state against it.
// The engine captures one when a save loads, before offline catch-up runs
// (game.GameState.SessionStart), which is exactly the state the player left
// at the end of their last session; the map can also take a fresh one when
// the player closes the map ("since you last looked").

// Visit is the baseline a recap is measured from.
type Visit struct {
	Tick      int
	SavedAt   time.Time
	Age       string
	Buildings map[string]int
	Ruins     map[string]int
	Pop       int
	// Relations holds each discovered civ's standing.
	Relations map[string]Relation
	Routes    int
	Harbinger string
	Completed int // expeditions completed
}

// VisitOf snapshots a model as a baseline.
func VisitOf(m *Model) *Visit {
	v := &Visit{Tick: m.Tick, Age: m.Age, Buildings: map[string]int{}, Ruins: map[string]int{},
		Pop: m.Workers.Pop, Relations: map[string]Relation{}, Routes: len(m.Routes),
		Completed: m.Expeditions.Completed}
	for _, b := range m.Buildings {
		if b.Count > 0 {
			v.Buildings[b.Key] = b.Count
		}
		if b.Ruins > 0 {
			v.Ruins[b.Key] = b.Ruins
		}
	}
	for _, f := range m.Factions {
		if f.Discovered {
			v.Relations[f.Key] = f.Relation
		}
	}
	if m.Harbinger != nil {
		v.Harbinger = m.Harbinger.Name
	}
	return v
}

// VisitFromSession converts the engine's load-time mark (nil when the game
// was not loaded from a save) into a baseline.
func VisitFromSession(s *game.SessionMark) *Visit {
	if s == nil {
		return nil
	}
	v := &Visit{Tick: s.Tick, SavedAt: s.SavedAt, Age: s.Age, Buildings: map[string]int{}, Ruins: map[string]int{},
		Pop: s.Pop, Relations: map[string]Relation{}, Routes: s.Routes, Harbinger: s.Harbinger,
		Completed: s.Expeditions}
	maps.Copy(v.Buildings, s.Buildings)
	maps.Copy(v.Ruins, s.Ruins)
	for k, st := range s.Civs {
		v.Relations[k] = relationOf(st)
	}
	return v
}

func relationOf(status string) Relation {
	switch status {
	case "war":
		return RelWar
	case "embargo":
		return RelEmbargo
	case "rival":
		return RelRival
	case "allied":
		return RelAllied
	case "friendly":
		return RelFriendly
	}
	return RelNeutral
}

// Change is a building count that moved.
type Change struct {
	Key, Name string
	N         int
}

// NewsKind sorts news items by what they are about.
type NewsKind uint8

const (
	NewsBuilt NewsKind = iota
	NewsAge
	NewsWonder
	NewsCiv
	NewsWar
	NewsPeace
	NewsRoute
	NewsLost
	NewsHarbinger
	NewsCatastrophe
	NewsExpedition
	NewsPop
)

// NewsItem is one line of the recap.
type NewsItem struct {
	Kind NewsKind
	Text string
	Bad  bool // a setback (draw it in the warning colour)
}

// Recap is what changed since the baseline.
type Recap struct {
	HasBaseline bool
	Since       int // ticks since the baseline
	Built       []Change
	Lost        []Change
	NewCount    int // copies built
	Items       []NewsItem
}

func recap(m *Model, st *game.GameState, v *Visit) Recap {
	r := Recap{}
	if m.Catastrophe.Pending != "" {
		r.Items = append(r.Items, NewsItem{Kind: NewsCatastrophe, Text: m.Catastrophe.PendingName + vibe.Line("mapmodel.recap.recap.01", " is upon you"), Bad: true})
	}
	if v == nil {
		return r
	}
	if ai, ok := m.Catalog.AgeIdx[v.Age]; ok && ai > m.AgeIdx {
		return r // a baseline from before a prestige says nothing useful
	}
	r.HasBaseline = true
	r.Since = m.Tick - v.Tick
	if v.Age != "" && v.Age != m.Age {
		r.Items = append(r.Items, NewsItem{Kind: NewsAge, Text: vibe.Line("mapmodel.recap.recap.02", "entered the ") + m.AgeName})
	}
	cur := map[string]int{}
	for _, b := range m.Buildings {
		cur[b.Key] = b.Count
	}
	keys := maps.Clone(cur)
	maps.Copy(keys, v.Buildings)
	for _, k := range slices.Sorted(maps.Keys(keys)) {
		d := m.Catalog.Defs[k]
		if d == nil || d.Wonder {
			continue // wonders make their own news
		}
		delta := cur[k] - v.Buildings[k]
		switch {
		case delta > 0:
			r.Built = append(r.Built, Change{k, d.Name, delta})
			r.NewCount += delta
		case delta < 0:
			r.Lost = append(r.Lost, Change{k, d.Name, -delta})
		}
	}
	byN := func(a, b Change) int { return cmp.Compare(b.N, a.N) }
	slices.SortStableFunc(r.Built, byN)
	slices.SortStableFunc(r.Lost, byN)
	for _, w := range m.Wonders {
		if w.Delta {
			r.Items = append(r.Items, NewsItem{Kind: NewsWonder, Text: w.Name + vibe.Line("mapmodel.recap.recap.03", " completed")})
		}
	}
	if len(r.Built) > 0 {
		var parts []string
		for i, c := range r.Built {
			if i == 3 {
				parts = append(parts, vibe.Line("mapmodel.recap.recap.04", "and ")+strconv.Itoa(len(r.Built)-3)+vibe.Line("mapmodel.recap.recap.05", " more"))
				break
			}
			parts = append(parts, "+"+strconv.Itoa(c.N)+" "+c.Name)
		}
		r.Items = append(r.Items, NewsItem{Kind: NewsBuilt, Text: strings.Join(parts, ", ")})
	}
	for _, f := range m.Factions {
		if !f.Discovered {
			continue
		}
		was, known := v.Relations[f.Key]
		switch {
		case !known:
			r.Items = append(r.Items, NewsItem{Kind: NewsCiv, Text: vibe.Line("mapmodel.recap.recap.06", "met the ") + f.Name})
		case f.Relation == RelWar && was != RelWar:
			r.Items = append(r.Items, NewsItem{Kind: NewsWar, Text: vibe.Line("mapmodel.recap.recap.07", "the ") + f.Name + vibe.Line("mapmodel.recap.recap.08", " declared war"), Bad: true})
		case was == RelWar && f.Relation != RelWar:
			r.Items = append(r.Items, NewsItem{Kind: NewsPeace, Text: vibe.Line("mapmodel.recap.recap.09", "peace with the ") + f.Name})
		case f.Relation == RelAllied && was != RelAllied:
			r.Items = append(r.Items, NewsItem{Kind: NewsCiv, Text: vibe.Line("mapmodel.recap.recap.10", "allied with the ") + f.Name})
		}
	}
	if n := len(m.Routes) - v.Routes; n > 0 {
		r.Items = append(r.Items, NewsItem{Kind: NewsRoute, Text: plural(n, vibe.Line("mapmodel.recap.recap.11", "trade route"), vibe.Line("mapmodel.recap.recap.12", "trade routes")) + vibe.Line("mapmodel.recap.recap.13", " opened")})
	}
	if n := m.Expeditions.Completed - v.Completed; n > 0 {
		r.Items = append(r.Items, NewsItem{Kind: NewsExpedition, Text: plural(n, vibe.Line("mapmodel.recap.recap.14", "expedition"), vibe.Line("mapmodel.recap.recap.15", "expeditions")) + vibe.Line("mapmodel.recap.recap.16", " came home")})
	}
	if m.Harbinger != nil && m.Harbinger.Name != v.Harbinger {
		r.Items = append(r.Items, NewsItem{Kind: NewsHarbinger, Text: m.Harbinger.Name + vibe.Line("mapmodel.recap.recap.17", " has come"), Bad: true})
	}
	if len(r.Lost) > 0 {
		lost := 0
		for _, c := range r.Lost {
			lost += c.N
		}
		r.Items = append(r.Items, NewsItem{Kind: NewsLost, Text: plural(lost, vibe.Line("mapmodel.recap.recap.18", "building"), vibe.Line("mapmodel.recap.recap.19", "buildings")) + vibe.Line("mapmodel.recap.recap.20", " lost or upgraded"), Bad: true})
	}
	if d := m.Workers.Pop - v.Pop; d > 0 && v.Pop > 0 {
		r.Items = append(r.Items, NewsItem{Kind: NewsPop, Text: "+" + strconv.Itoa(d) + vibe.Line("mapmodel.recap.recap.21", " people")})
	}
	return r
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}

// Headline is the recap as one news line of at most max cells: "since
// your last visit (2h 08m): +2 Guildhall, +1 Castle Keep · met the Ironhold
// Clans". Empty when nothing changed.
func (r Recap) Headline(max int) string {
	if len(r.Items) == 0 {
		return ""
	}
	var parts []string
	for _, it := range r.Items {
		parts = append(parts, it.Text)
	}
	s := strings.Join(parts, " · ")
	if r.HasBaseline && r.Since > 0 {
		s = vibe.Line("mapmodel.recap.Headline.01", "since your last visit (") + Duration(r.Since) + "): " + s
	}
	return Clip(s, max)
}

// Duration renders ticks as play time: "2h 08m", "14m", "3d 4h".
func Duration(ticks int) string {
	secs := int(float64(ticks) * game.BaseTickInterval.Seconds())
	mins := secs / 60
	switch {
	case mins < 1:
		return strconv.Itoa(secs) + "s"
	case mins < 60:
		return strconv.Itoa(mins) + "m"
	case mins < 24*60:
		m := mins % 60
		pad := ""
		if m < 10 {
			pad = "0"
		}
		return strconv.Itoa(mins/60) + "h " + pad + strconv.Itoa(m) + "m"
	}
	return strconv.Itoa(mins/1440) + "d " + strconv.Itoa(mins%1440/60) + "h"
}

// Clip cuts s to at most max runes, ending in "…" when cut.
func Clip(s string, max int) string {
	if max <= 0 {
		return ""
	}
	n := 0
	for i := range s {
		if n == max-1 {
			rest := 0
			for range s[i:] {
				rest++
			}
			if rest > 1 {
				return s[:i] + "…"
			}
			return s
		}
		n++
	}
	return s
}
