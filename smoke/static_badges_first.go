package smoke

import (
	"fmt"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// static_badges_first.go holds the Badge Covenant's rule for the first rung
// of a lifetime ladder: an ordinary first run must hold it by the time it
// reaches the Bronze Age, or the second age after the one the ladder's
// subject first appears in, whichever comes later. A ladder whose first
// rung is one occurrence (the first prestige, the first festival) passes:
// no rung can ask for less.
//
// "An ordinary first run" is the model the lifetime proof already counts
// runs in (badgeCheck.added): typical incomes over the pacing targets, a
// tenth of what storage lets a run build, the wonder of every age.

// firstRungAge is the age an ordinary first run must hold the first rung of
// a ladder by, for a subject of the first ages.
const firstRungAge = "bronze_age"

// firstRungGrace is how many ages after its subject first appears a ladder
// may take to give its first rung.
const firstRungGrace = 2

// FirstRung is one ladder's first rung against what an ordinary first run
// has by its deadline.
type FirstRung struct {
	Ladder string `json:"ladder"`
	Key    string `json:"key"`
	// Need is the first rung's count.
	Need float64 `json:"need"`
	// From is the age the ladder's subject first appears in, By the age the
	// run must hold the rung on reaching, Have what it has added by then.
	From string  `json:"from"`
	By   string  `json:"by"`
	Have float64 `json:"have"`
	// Rungs is every rung's count, lowest first.
	Rungs []float64 `json:"rungs"`
}

// firstRungThrough is the place of the last age an ordinary first run has
// finished when it must hold the first rung of a ladder whose subject first
// appears in the age at place first.
func (c *badgeCheck) firstRungThrough(first int) int {
	return min(max(c.m.idx[firstRungAge], first+firstRungGrace)-1, len(c.m.ages)-1)
}

// firstAge is the place of the age the thing a lifetime counter counts
// first appears in.
func (c *badgeCheck) firstAge(counter string) int {
	m := c.m
	if res, ok := strings.CutPrefix(counter, config.BadgeEvProduced+"."); ok {
		if age, _ := c.set.FirstProduction(res); age != "" {
			return m.idx[age]
		}
		return m.resAge[res]
	}
	if lineage, ok := strings.CutPrefix(counter, config.BadgeEvBuiltLineage+"."); ok {
		first := len(m.ages) - 1
		for _, k := range sortedKeys(m.defs) {
			if m.defs[k].LineageKey != lineage {
				continue
			}
			if ceil := m.ceiling(k); ceil.age >= 0 {
				first = min(first, ceil.age)
			}
		}
		return first
	}
	if civ, ok := strings.CutPrefix(counter, config.BadgeEvDeal+"."); ok {
		f, _ := c.set.Faction(civ)
		return m.idx[f.MinAge]
	}
	switch counter {
	case config.BadgeEvExpedition:
		order := c.set.Indexes()
		mm := game.NewMilitaryManager()
		for i, a := range m.ages {
			if len(mm.GetAvailableExpeditions(a.Key, order)) > 0 {
				return i
			}
		}
	case config.BadgeEvDeal:
		first := len(m.ages) - 1
		for _, f := range c.set.Factions() {
			if pos, known := m.idx[f.MinAge]; known {
				first = min(first, pos)
			}
		}
		return first
	case config.BadgeEvRaidBlunted:
		// A raid is blunted once a garrison stands: the first age a
		// building that trains soldiers can be built in, soldiers allowed.
		first := len(m.ages) - 1
		for _, k := range sortedKeys(m.defs) {
			if d := m.defs[k]; d.Category == "military" {
				if ceil := m.ceiling(k); ceil.age >= 0 {
					first = min(first, ceil.age)
				}
			}
		}
		return max(first, m.resAge["soldiers"])
	case config.BadgeEvBuildingUpgraded:
		// The first upgrade: the first age a second tier of a lineage can
		// be built in.
		first := len(m.ages) - 1
		for _, k := range sortedKeys(m.defs) {
			if d := m.defs[k]; d.LineageKey != "" && d.LineageTier > 0 {
				if ceil := m.ceiling(k); ceil.age >= 0 {
					first = min(first, ceil.age)
				}
			}
		}
		return first
	case config.BadgeEvMarketTrade:
		first := len(m.ages) - 1
		for _, x := range c.set.ExchangeRates() {
			if pos, known := m.idx[x.MinAge]; known {
				first = min(first, pos)
			}
		}
		return first
	case config.BadgeEvFestival:
		return m.resAge["culture"]
	case config.BadgeEvBlackMarket:
		return m.idx["colonial_age"]
	case config.BadgeEvEraEvent + ".bad_challenging":
		// An era event comes on entering an era: the second era's first age.
		for i, a := range m.ages {
			if c.set.EraOf(a.Key) != c.set.EraOf(m.ages[0].Key) {
				return i
			}
		}
	case config.BadgeEvPrestige, config.BadgeEvLegacyPrestige, config.BadgeEvMemoryAccepted:
		return m.runEnd
	}
	return 0
}

// firstRungs checks the first rung of every lifetime ladder in badges and
// returns each ladder's verdict, in catalog order.
func (c *badgeCheck) firstRungs(badges []config.BadgeDef) []FirstRung {
	type ladder struct {
		name, counter string
		rungs         []config.BadgeDef
	}
	var ladders []*ladder
	byKey := map[string]*ladder{}
	for _, def := range badges {
		if def.Ladder == "" || def.Scope != config.BadgeLifetime || def.Integrity() {
			continue
		}
		k := def.Ladder + "\x00" + def.Counter
		l, ok := byKey[k]
		if !ok {
			l = &ladder{name: def.Ladder, counter: def.Counter}
			byKey[k] = l
			ladders = append(ladders, l)
		}
		l.rungs = append(l.rungs, def)
	}

	// The badges an ordinary first run is proven to hold on reaching
	// firstRungAge: the ages reached, the wonders raised, and every first
	// rung this rule holds to that deadline. It is what the ladder over
	// earned badges is measured against, so that ladder is checked last.
	early := c.m.idx[firstRungAge]
	proven := 0
	for _, def := range badges {
		switch {
		case def.Proof.Rule == config.BadgeRuleGate && def.Event == config.BadgeEvAgeReached:
			if pos, ok := c.m.idx[def.Subject]; ok && pos <= early {
				proven++
			}
		case def.Proof.Rule == config.BadgeRuleWonder:
			for i := 0; i < early; i++ {
				if c.set.Wonder(c.m.ages[i].Key) == def.Subject {
					proven++
				}
			}
		}
	}

	var out, counted []FirstRung
	check := func(l *ladder, have float64, first, through int, basis string) FirstRung {
		slices.SortStableFunc(l.rungs, func(a, b config.BadgeDef) int {
			switch {
			case a.Threshold < b.Threshold:
				return -1
			case a.Threshold > b.Threshold:
				return 1
			}
			return 0
		})
		low := l.rungs[0]
		r := FirstRung{Ladder: l.name, Key: low.Key, Need: low.Threshold, Have: have,
			From: c.m.ageName(first), By: c.m.ageName(min(through+1, len(c.m.ages)-1))}
		for _, d := range l.rungs {
			r.Rungs = append(r.Rungs, d.Threshold)
		}
		if low.Threshold > 1 && low.Threshold > have {
			c.fail(low.Key, "first_rung", fmt.Sprintf("%s (%s) is the first rung of the %s ladder and asks for %s, but an ordinary first run has %s by the time it reaches the %s (%s). A first rung must be earned by the %s, or by the second age after its subject first appears (the %s).",
				low.Name, low.Key, l.name, config.FormatAmount(low.Threshold), config.FormatAmount(have), r.By, basis,
				c.m.ageName(early), r.From))
		}
		return r
	}
	for _, l := range ladders {
		if strings.HasPrefix(l.counter, config.BadgeEvBadge) {
			continue
		}
		first := c.firstAge(l.counter)
		through := c.firstRungThrough(first)
		have, basis, ok := c.added(l.counter, through)
		if !ok {
			// The lifetime proof reports a counter it cannot count.
			continue
		}
		r := check(l, have, first, through, basis)
		out = append(out, r)
		// A rung of one occurrence passes the rule without being something
		// an ordinary run does by then (a Succumb, an Appease): only a
		// count the run is shown to reach is counted on.
		if through < early && r.Need > 1 && r.Need <= have {
			proven++
		}
	}
	for _, l := range ladders {
		if l.counter != config.BadgeEvBadge {
			continue
		}
		counted = append(counted, check(l, float64(proven), 0, early-1,
			"the ages reached, the wonders raised and the first rungs due by then"))
	}
	return append(out, counted...)
}

// StaticFirstRungs is every lifetime ladder's first rung against what an
// ordinary first run holds by its deadline, for the core ruleset.
func StaticFirstRungs() []FirstRung {
	_, _, first := staticBadgesFull()
	return first
}
