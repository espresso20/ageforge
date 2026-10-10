package rules

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
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
// test reports it. Thresholds that belong to the game's tables are filled
// in last (config.BadgeMeasure).
func (s *Set) buildBadges(written []config.BadgeDef, families []config.BadgeFamilyDef) {
	s.badges = nil
	for _, f := range families {
		s.badges = append(s.badges, s.expandFamily(f)...)
	}
	s.badges = append(s.badges, written...)
	inSet := map[string]int{}
	for _, b := range s.badges {
		if b.Set != "" && !b.Integrity() {
			inSet[b.Set]++
		}
	}
	for i := range s.badges {
		b := &s.badges[i]
		switch b.Measure {
		case config.BadgeMeasureMilestones:
			b.Threshold = float64(len(s.milestones))
		case config.BadgeMeasureSet:
			b.Threshold = float64(inSet[strings.TrimPrefix(b.Counter, config.BadgeEvBadge+".set.")])
		}
	}
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
			threshold := f.Threshold
			if t, ok := f.Thresholds[sub.key]; ok {
				threshold = t
			}
			if f.Measure == config.BadgeMeasureTechs {
				threshold = float64(len(s.techsByAge[sub.key]))
				if threshold == 0 {
					continue // an age with no techs has no syllabus
				}
			}
			out = append(out, s.familyBadge(f, sub, config.BadgeRung{}, 0, threshold))
			continue
		}
		rungs, counts := f.Rungs, f.Ladders[sub.key]
		if f.Measure == config.BadgeMeasureProduction {
			rungs, counts = s.productionLadder(f, sub)
		}
		for i, rung := range rungs {
			if i >= len(counts) {
				break
			}
			out = append(out, s.familyBadge(f, sub, rung, i+1, counts[i]))
		}
	}
	return out
}

// RunProduction is what one run produces of a resource: the sum, over the
// run's ages, of the resource's typical income times the age's pacing
// target in ticks. A run is the ages before the one a full prestige is made
// from (runEnd is the last of them); deep is true for a resource that comes
// after them, which is measured over its own first two ages instead.
func (s *Set) RunProduction(res string) (perRun float64, deep bool) {
	def, ok := s.resourceByKey[res]
	if !ok {
		return 0, false
	}
	first, last := 0, s.runEndPos()
	if pos, ok := s.agePos[def.Age]; ok && pos > last {
		first, last, deep = pos, min(pos+1, len(s.ageKeys)-1), true
	}
	for i := first; i <= last; i++ {
		age := s.ageKeys[i]
		perRun += float64(s.typIncome[age][res] * s.targetTicks[age])
	}
	return perRun, deep
}

// runEndPos is the place of a run's last age: the age before the Digital
// Era's first, which is where a full prestige is made from.
func (s *Set) runEndPos() int {
	for _, e := range s.eras {
		if e.Order == badgeRunEras && len(e.Ages) > 0 {
			if pos, ok := s.agePos[e.Ages[0]]; ok && pos > 0 {
				return pos - 1
			}
		}
	}
	return len(s.ageKeys) - 1
}

// badgeRunEras is how many eras a run plays through before a full
// prestige: the Stone, Iron, Steel and Electric Eras.
const badgeRunEras = 4

// ProductionThrough is what an ordinary first run has produced of a
// resource by the end of age: the sum, over the ages up to and including it,
// of the resource's typical income times the age's pacing target in ticks.
func (s *Set) ProductionThrough(res, age string) float64 {
	last, ok := s.agePos[age]
	if !ok {
		return 0
	}
	sum := 0.0
	for i := 0; i <= last; i++ {
		a := s.ageKeys[i]
		sum += float64(s.typIncome[a][res] * s.targetTicks[a])
	}
	return sum
}

// FirstProduction is the first age an ordinary town makes a resource in
// (the first with a typical income of it, from the age the resource unlocks
// in; from the first age for a resource its own buildings unlock) and what
// it makes of it there: that income over the age's pacing target. "" and 0
// for a resource no age makes.
func (s *Set) FirstProduction(res string) (age string, made float64) {
	def, ok := s.resourceByKey[res]
	if !ok {
		return "", 0
	}
	from := s.agePos[def.Age]
	if def.BuiltUnlocks {
		from = 0
	}
	for _, a := range s.ageKeys[from:] {
		if inc := s.typIncome[a][res]; inc > 0 {
			return a, float64(inc * s.targetTicks[a])
		}
	}
	return "", 0
}

// productionLadder is a resource's rungs and their counts. The first rung
// is what an ordinary town makes of the resource in the first age it makes
// any, so a first run earns it there or soon after. The top rung is
// f.TopRuns runs of what one run produces. The rungs between climb in even
// multiplicative steps. A resource that comes after a run's last age has
// one rung fewer (it skips the first tier), on the same rule.
func (s *Set) productionLadder(f config.BadgeFamilyDef, sub badgeSubject) ([]config.BadgeRung, []float64) {
	perRun, deep := s.RunProduction(sub.key)
	_, first := s.FirstProduction(sub.key)
	if perRun <= 0 || first <= 0 || f.TopRuns <= 0 {
		return nil, nil
	}
	rungs := f.Rungs
	if deep && len(rungs) > 1 {
		rungs = rungs[1:]
	}
	return rungs, evenLadder(first, float64(perRun*f.TopRuns), len(rungs))
}

// evenLadder is n counts from first to top in even multiplicative steps,
// each rounded down to two figures. A top at or under first leaves the one
// rung.
func evenLadder(first, top float64, n int) []float64 {
	first = twoFigures(first)
	if n <= 1 || top <= first {
		return []float64{first}
	}
	counts := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		v := first
		switch {
		case i == n-1:
			v = top
		case i > 0:
			v = float64(first * detmath.Pow(top/first, float64(i)/float64(n-1)))
		}
		counts = append(counts, twoFigures(v))
	}
	return counts
}

// twoFigures rounds v down to two significant figures, so a rung reads as
// a round number and never asks for more runs than it says.
func twoFigures(v float64) float64 {
	if v <= 0 {
		return 0
	}
	// The power of ten that leaves two figures before the point, found by
	// stepping: no logarithm, so the rung is the same on every machine.
	mag := 1.0
	for v/mag >= 100 {
		mag *= 10
	}
	for v/mag < 10 {
		mag /= 10
	}
	return math.Floor(v/mag) * mag
}

// familyBadge is the badge of one subject (and rung n, from 1; 0 for a
// family that is not a ladder).
func (s *Set) familyBadge(f config.BadgeFamilyDef, sub badgeSubject, rung config.BadgeRung, n int, threshold float64) config.BadgeDef {
	mid := sub.name
	if strings.HasPrefix(mid, "The ") {
		mid = "the " + strings.TrimPrefix(mid, "The ")
	}
	fill := strings.NewReplacer(
		"{key}", sub.key,
		"{name}", sub.name,
		"{Name}", textfmt.Capitalize(sub.name),
		"{mid}", mid,
		"{short}", strings.TrimSuffix(sub.name, " Age"),
		"{lname}", strings.ToLower(sub.name),
		"{rung}", rung.Name,
		"{n}", strconv.Itoa(n),
		"{count}", badgeCount(threshold),
		"{era}", strconv.Itoa(s.eraOrderOfAge(sub.age)),
		"{techs}", techsLine(int(threshold), sub.name),
	).Replace
	b := config.BadgeDef{
		Key:        fill(f.Key),
		Family:     f.Family,
		Subject:    sub.key,
		AnySubject: f.AnySubject,
		Name:       strings.TrimSpace(fill(f.Name)),
		Desc:       fill(f.Desc),
		Hint:       f.Hint,
		Tier:       f.Tier,
		Rarity:     f.Rarity,
		Scope:      f.Scope,
		Counter:    fill(f.Counter),
		Threshold:  threshold,
		Event:      fill(f.Event),
		InAge:      fill(f.InAge),
		When:       slices.Clone(f.When),
		Reveal:     f.Reveal,
		Proof:      f.Proof,
		Emblem:     fill(f.Emblem),
		Set:        fill(f.Set),
	}
	if e, ok := f.Emblems[sub.key]; ok {
		b.Emblem = e
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
	if f.RevealAtAge {
		b.Reveal = s.revealAtAge(sub.age)
	}
	if name := f.Names[b.Key]; name != "" {
		b.Name = name
	}
	if desc := f.Descs[b.Key]; desc != "" {
		b.Desc = fill(desc)
	}
	b.Aliases = slices.Clone(f.Aliases[b.Key])
	b.Reward = f.Rewards[b.Key]
	return b
}

// techsLine is the description of an age's syllabus: every tech of the
// age, said the way the number reads.
func techsLine(n int, age string) string {
	switch {
	case n == 1:
		return "Research the one " + age + " tech."
	case n == 2:
		return "Research both " + age + " techs in one run."
	}
	return "Research all " + strconv.Itoa(n) + " " + age + " techs in one run."
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
	if v >= 1e18 {
		// Past the last suffix: the count in Q, its thousands set off.
		q := strconv.FormatInt(int64(math.Round(v/1e15)), 10)
		for i := len(q) - 3; i > 0; i -= 3 {
			q = q[:i] + "," + q[i:]
		}
		return q + "Q"
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
	case config.BadgeSourceDooms:
		for _, e := range s.eras {
			if !s.CatastropheAllowed(e.Key) || len(e.Ages) == 0 {
				continue
			}
			name, _ := s.Catastrophe(e.Key)
			out = append(out, badgeSubject{key: e.Key, name: name, age: e.Ages[0], reveal: config.RevealUntilDoomNamed(e.Key)})
		}
	case config.BadgeSourceDomains:
		// A domain a building can be staffed in, with the age of the first
		// such building.
		first := map[string]int{}
		for _, b := range s.buildings {
			if b.WorkerDomain == "" || b.WorkerCapacity <= 0 {
				continue
			}
			pos, ok := s.agePos[b.RequiredAge]
			if !ok {
				continue
			}
			if cur, seen := first[b.WorkerDomain]; !seen || pos < cur {
				first[b.WorkerDomain] = pos
			}
		}
		for _, d := range s.domains {
			pos, ok := first[d]
			if !ok {
				continue
			}
			age := s.ageKeys[pos]
			out = append(out, badgeSubject{key: d, name: textfmt.Capitalize(d), age: age, reveal: s.revealAtAge(age)})
		}
	case config.BadgeSourceExpeditions:
		for _, x := range s.badgeExpeditions {
			out = append(out, badgeSubject{key: x.Key, name: x.Name, age: x.MinAge, reveal: s.revealAtAge(x.MinAge)})
		}
	case config.BadgeSourceThemes:
		for _, t := range s.badgeThemes {
			sub := badgeSubject{key: t.Key, name: t.Name}
			if t.Badge != "" {
				sub.reveal = config.RevealUntilBadge(t.Badge)
			}
			out = append(out, sub)
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

// BadgeWornTitles returns the titles badges give (config.BadgeTitles): each
// is held with its badge, or with every badge of its set.
func (s *Set) BadgeWornTitles() []config.BadgeTitleDef { return slices.Clone(s.badgeWornTitles) }

// BadgesInSet returns the keys of the badges of a set, in catalog order.
func (s *Set) BadgesInSet(set string) []string {
	var out []string
	for _, b := range s.badges {
		if b.Set == set && !b.Integrity() {
			out = append(out, b.Key)
		}
	}
	return out
}

// BadgeSessionEvents is the events that are told to the account and never
// tallied in a run's facts (config.BadgeSessionEvents).
func (s *Set) BadgeSessionEvents() []string { return config.BadgeSessionEvents() }

// IsFlowResource reports whether res is a flow resource: one that is spent
// as it is made rather than saved up (config.IsFlowResource).
func (s *Set) IsFlowResource(res string) bool { return config.IsFlowResource(res) }
