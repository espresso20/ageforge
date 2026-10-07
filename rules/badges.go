package rules

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// badgeSubject is one row of a table a badge family is made from.
type badgeSubject struct {
	key    string
	name   string
	age    string // the age its tier comes from (TierByEra)
	reveal config.BadgeReveal
}

// buildBadges expands the families into badges, in table order, and puts
// the hand-written ones after them. A family row that cannot be expanded
// (no Key template, a ladder without counts) makes no badges; the guard
// test reports it.
func (s *Set) buildBadges(written []config.BadgeDef, families []config.BadgeFamilyDef) {
	s.badges = nil
	for _, f := range families {
		s.badges = append(s.badges, s.expandFamily(f)...)
	}
	s.badges = append(s.badges, written...)
	s.badgeByKey = make(map[string]config.BadgeDef, len(s.badges))
	s.badgeAlias = map[string]string{}
	for _, b := range s.badges {
		if _, taken := s.badgeByKey[b.Key]; !taken {
			s.badgeByKey[b.Key] = b
		}
		for _, a := range b.Aliases {
			if _, taken := s.badgeAlias[a]; !taken {
				s.badgeAlias[a] = b.Key
			}
		}
	}
}

// expandFamily makes a family's badges: one per subject, or one per
// subject and rung.
func (s *Set) expandFamily(f config.BadgeFamilyDef) []config.BadgeDef {
	var out []config.BadgeDef
	for _, sub := range s.badgeSubjects(f.Source) {
		if len(f.Only) > 0 && !slices.Contains(f.Only, sub.key) {
			continue
		}
		if slices.Contains(f.Except, sub.key) {
			continue
		}
		if len(f.Rungs) == 0 {
			out = append(out, s.familyBadge(f, sub, config.BadgeRung{}, 0, f.Threshold))
			continue
		}
		counts := f.Ladders[sub.key]
		for i, rung := range f.Rungs {
			if i >= len(counts) {
				break
			}
			out = append(out, s.familyBadge(f, sub, rung, i+1, counts[i]))
		}
	}
	return out
}

// familyBadge is the badge of one subject (and rung n, from 1; 0 for a
// family that is not a ladder).
func (s *Set) familyBadge(f config.BadgeFamilyDef, sub badgeSubject, rung config.BadgeRung, n int, threshold float64) config.BadgeDef {
	fill := strings.NewReplacer(
		"{key}", sub.key,
		"{name}", sub.name,
		"{lname}", strings.ToLower(sub.name),
		"{rung}", rung.Name,
		"{n}", strconv.Itoa(n),
		"{count}", badgeCount(threshold),
		"{era}", strconv.Itoa(s.eraOrderOfAge(sub.age)),
	).Replace
	b := config.BadgeDef{
		Key:       fill(f.Key),
		Family:    f.Family,
		Subject:   sub.key,
		Name:      strings.TrimSpace(fill(f.Name)),
		Desc:      fill(f.Desc),
		Tier:      f.Tier,
		Rarity:    f.Rarity,
		Scope:     f.Scope,
		Counter:   fill(f.Counter),
		Threshold: threshold,
		Event:     fill(f.Event),
		InAge:     fill(f.InAge),
		Reveal:    f.Reveal,
		Proof:     f.Proof,
		Emblem:    fill(f.Emblem),
	}
	if n > 0 {
		b.Ladder = strings.TrimSpace(fill(f.Ladder))
	}
	if f.TierByEra {
		if era, ok := s.eraByKey[s.eraOfAge[sub.age]]; ok && era.Order < len(config.BadgeEraTiers) {
			b.Tier = config.BadgeEraTiers[era.Order]
		}
	}
	if n > 0 {
		b.Tier = rung.Tier
	}
	if t, ok := f.Tiers[sub.key]; ok {
		b.Tier = t
	}
	if f.RevealBySubject {
		b.Reveal = sub.reveal
	}
	if name := f.Names[b.Key]; name != "" {
		b.Name = name
	}
	if desc := f.Descs[b.Key]; desc != "" {
		b.Desc = desc
	}
	b.Aliases = slices.Clone(f.Aliases[b.Key])
	b.Reward = f.Rewards[b.Key]
	return b
}

// eraOrderOfAge is the order of the era an age belongs to, 0 for an age
// the ruleset does not have.
func (s *Set) eraOrderOfAge(age string) int {
	if era, ok := s.eraByKey[s.eraOfAge[age]]; ok {
		return era.Order
	}
	return 0
}

// BadgeTitle is the title a badge score holds and its rank (0 for the one
// every account starts with), and the next title with the score it asks
// for ("" and 0 at the top). complete is whether the account holds every
// badge that counts: that is a title of its own, above the rest.
func (s *Set) BadgeTitle(points int, complete bool) (title string, rank int, next string, nextAt int) {
	for i, t := range s.badgeTitles {
		if points >= t.Points {
			title, rank = t.Title, i
			continue
		}
		next, nextAt = t.Title, t.Points
		break
	}
	if complete {
		return config.BadgeCompleteTitle, len(s.badgeTitles), "", 0
	}
	return title, rank, next, nextAt
}

// badgeCount writes a threshold the way a description shows it: "14",
// "1,400", "1.1T".
func badgeCount(v float64) string {
	if v == float64(int64(v)) && v < 1e6 {
		return textfmt.Int(int(v))
	}
	return config.FormatAmount(v)
}

// badgeSubjects lists a source table's rows, in the table's own order.
func (s *Set) badgeSubjects(src config.BadgeSource) []badgeSubject {
	var out []badgeSubject
	switch src {
	case config.BadgeSourceNone:
		out = append(out, badgeSubject{})
	case config.BadgeSourceAges:
		for i, a := range s.ages {
			reveal := config.RevealUntilNextAge(a.Key)
			if i == 0 {
				reveal = config.BadgeReveal{}
			}
			out = append(out, badgeSubject{key: a.Key, name: a.Name, age: a.Key, reveal: reveal})
		}
	case config.BadgeSourceEras:
		for _, e := range s.eras {
			first := ""
			if len(e.Ages) > 0 {
				first = e.Ages[0]
			}
			out = append(out, badgeSubject{key: e.Key, name: e.Name, age: first, reveal: s.revealAtAge(first)})
		}
	case config.BadgeSourceWonders:
		for _, a := range s.ages {
			if w, ok := s.buildingByKey[s.wonders[a.Key]]; ok {
				out = append(out, badgeSubject{key: w.Key, name: w.Name, age: a.Key, reveal: s.revealAtAge(a.Key)})
			}
		}
	case config.BadgeSourceLineages:
		first := map[string]int{}
		for _, b := range s.buildings {
			if b.LineageKey == "" || b.LineageKey == lineageWonder || b.LineageKey == lineageMonuments {
				continue
			}
			pos, ok := s.agePos[b.RequiredAge]
			if !ok {
				continue
			}
			if cur, seen := first[b.LineageKey]; !seen || pos < cur {
				first[b.LineageKey] = pos
			}
		}
		keys := make([]string, 0, len(first))
		for k := range first {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			age := s.ageKeys[first[k]]
			out = append(out, badgeSubject{
				key: k, name: textfmt.Capitalize(strings.ReplaceAll(k, "_", " ")),
				age: age, reveal: s.revealAtAge(age),
			})
		}
	case config.BadgeSourceResources:
		for _, r := range s.resources {
			out = append(out, badgeSubject{key: r.Key, name: r.Name, age: r.Age, reveal: s.revealAtAge(r.Age)})
		}
	case config.BadgeSourceCivs:
		for _, c := range s.factions {
			out = append(out, badgeSubject{key: c.Key, name: c.Name, age: c.MinAge, reveal: config.RevealUntilCivMet(c.Key)})
		}
	case config.BadgeSourceHarbingers:
		for _, a := range s.ageKeys {
			if h, ok := s.harbinger[a]; ok {
				out = append(out, badgeSubject{key: h.Key, name: h.Name, age: a, reveal: config.RevealUntilHarbingerMet(h.Key)})
			}
		}
	case config.BadgeSourceAwakenings:
		for _, a := range s.awakenings {
			out = append(out, badgeSubject{
				key: a.Key, name: a.Name, age: a.TriggerAge,
				reveal: config.RevealUntilSeen(config.BadgeEvAwakening + "." + a.Key),
			})
		}
	}
	return out
}

// revealAtAge hides a badge until the account has reached age. Something of
// the first age shows from the start.
func (s *Set) revealAtAge(age string) config.BadgeReveal {
	if pos, ok := s.agePos[age]; !ok || pos == 0 {
		return config.BadgeReveal{}
	}
	return config.RevealUntilAge(age)
}

// ===== Badges =====

// Badges returns every badge: each family's in table order, then the
// hand-written ones.
func (s *Set) Badges() []config.BadgeDef { return slices.Clone(s.badges) }

// Badge returns a badge's definition.
func (s *Set) Badge(key string) (config.BadgeDef, bool) {
	b, ok := s.badgeByKey[key]
	return b, ok
}

// BadgeForAlias returns the key of the badge that an older account file
// knew as alias (an account achievement), and whether there is one.
func (s *Set) BadgeForAlias(alias string) (string, bool) {
	k, ok := s.badgeAlias[alias]
	return k, ok
}

// BadgeProblems lists what is wrong with the badge tables themselves: a
// family that made no badges, a key used twice, an alias two badges claim.
// Empty for a sound catalog. The guard test fails on any.
func BadgeProblems(written []config.BadgeDef, families []config.BadgeFamilyDef, set *Set) []string {
	var out []string
	for _, f := range families {
		if f.Key == "" {
			out = append(out, fmt.Sprintf("the %s family has no key template", f.Family))
			continue
		}
		if len(set.expandFamily(f)) == 0 {
			out = append(out, fmt.Sprintf("the %s family makes no badges: no subject of its table passes Only and Except, or its ladders have no counts", f.Family))
		}
		for _, k := range f.Only {
			found := false
			for _, sub := range set.badgeSubjects(f.Source) {
				found = found || sub.key == k
			}
			if !found {
				out = append(out, fmt.Sprintf("the %s family keeps %q, which its table does not have", f.Family, k))
			}
		}
	}
	seen := map[string]bool{}
	alias := map[string]string{}
	for _, b := range set.badges {
		if seen[b.Key] {
			out = append(out, fmt.Sprintf("the badge key %q is used twice", b.Key))
		}
		seen[b.Key] = true
		for _, a := range b.Aliases {
			if other, taken := alias[a]; taken && other != b.Key {
				out = append(out, fmt.Sprintf("the alias %q belongs to both %s and %s", a, other, b.Key))
			}
			alias[a] = b.Key
		}
	}
	return out
}
