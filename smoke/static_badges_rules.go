package smoke

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// static_badges_rules.go holds the proofs the full catalog added to the
// Badge Covenant (static_badges.go): the last copy a store allows, most of
// an age's buildings at once, a wonder, an age's techs, a domain's
// payroll, a habit, and the things the game simply does. It also knows what
// one run adds to each counter a ladder climbs, and how often one run can
// do what a run badge counts.
//
// A number here is computed from the game's tables wherever the tables
// have it. Where they do not (how often a bot's garrison blunts a raid),
// the number is the nightly bot's, and its basis says so.

// The rates the tables cannot give: what the nightly greedy bot did in a
// run, as the badge design recorded it.
const (
	botRaidsBluntedPerRun   = 30.0
	botUpgradesPerRun       = 850.0
	botHardTimesPerRun      = 1.4
	botMemoriesPerPrestige  = 0.5
	humanMarketTradesPerRun = 200.0
)

// moreRules runs one of the catalog's static proofs and reports whether
// the rule is one of them.
func (c *badgeCheck) moreRules(def config.BadgeDef, label string, r *BadgeReach) bool {
	m := c.m
	switch def.Proof.Rule {
	case config.BadgeRuleCeiling:
		ceil := m.ceiling(def.Subject)
		if _, ok := m.defs[def.Subject]; !ok || ceil.age < 0 {
			c.fail(def.Key, "ceiling", fmt.Sprintf("%s counts copies of %q, which can never be built.", label, def.Subject))
			return true
		}
		if own := m.ages[ceil.age].Key; def.Counter != "standing."+def.Subject || def.InAge != own {
			c.fail(def.Key, "ceiling", fmt.Sprintf("%s must count %s standing in the %s, the only age they can be built in.", label, pluralName(def.Subject), m.ageName(ceil.age)))
		}
		r.Need, r.Limit = def.Threshold, float64(ceil.n)
		r.Basis = fmt.Sprintf("every one of the %d %s that can stand in the %s", ceil.n, pluralName(def.Subject), m.ageName(ceil.age))
		if def.Threshold > float64(ceil.n) {
			c.fail(def.Key, "ceiling", fmt.Sprintf("%s needs %s %s in the %s, but at most %d can stand there.", label, grouped(int(def.Threshold)), pluralName(def.Subject), m.ageName(ceil.age), ceil.n))
		}

	case config.BadgeRuleAgeCopies:
		pos, ok := m.idx[def.InAge]
		if !ok || def.Counter != "standing_age."+def.InAge {
			c.fail(def.Key, "age_copies", fmt.Sprintf("%s must count the buildings of its own age standing (Counter %q, InAge %q).", label, def.Counter, def.InAge))
			return true
		}
		total := 0
		for _, k := range sortedKeys(m.defs) {
			if d := m.defs[k]; d.Category == "wonder" || d.RequiredAge != def.InAge {
				continue
			}
			if ceil := m.ceiling(k); ceil.age == pos {
				total += ceil.n
			}
		}
		limit := math.Floor(float64(config.BadgeAgeShare * float64(total)))
		r.Need, r.Limit = def.Threshold, limit
		r.Basis = fmt.Sprintf("%d%% of the %s buildings the %s's storage allows", int(config.BadgeAgeShare*100), grouped(total), m.ageName(pos))
		if def.Threshold <= 0 || def.Threshold > limit {
			c.fail(def.Key, "age_copies", fmt.Sprintf("%s needs %s of the %s's buildings standing, and may ask for at most %s (%s).", label, grouped(int(def.Threshold)), m.ageName(pos), grouped(int(limit)), r.Basis))
		}

	case config.BadgeRuleWonder:
		found := false
		for _, a := range m.ages {
			found = found || c.set.Wonder(a.Key) == def.Subject
		}
		ceil := m.ceiling(def.Subject)
		if !found || ceil.age < 0 || ceil.n < 1 || def.Event != config.BadgeEvWonderRaised {
			c.fail(def.Key, "wonder", fmt.Sprintf("%s asks for %q to be raised, which is no age's wonder or cannot be paid for in its age.", label, def.Subject))
		}
		r.Basis = "the wonder of its age"

	case config.BadgeRuleTechs:
		n := len(c.set.TechsOf(def.Subject))
		r.Need, r.Limit, r.Basis = def.Threshold, float64(n), "every tech of its age"
		if n == 0 || def.Threshold != float64(n) || def.Counter != "researched_age."+def.Subject {
			c.fail(def.Key, "techs", fmt.Sprintf("%s asks for %v techs of the age %q, which has %d.", label, def.Threshold, def.Subject, n))
		}

	case config.BadgeRuleStaffing:
		slots, deep := 0, 0
		for _, k := range sortedKeys(m.defs) {
			d := m.defs[k]
			if d.WorkerDomain != def.Subject || d.WorkerCapacity <= 0 {
				continue
			}
			ceil := m.ceiling(k)
			if ceil.age < 0 {
				continue
			}
			deep += ceil.n * d.WorkerCapacity
			if ceil.age <= m.runEnd {
				slots += ceil.n * d.WorkerCapacity
			}
		}
		housing, where := m.housing[m.runEnd], "a run"
		if slots == 0 {
			slots, housing, where = deep, m.housing[len(m.housing)-1], "a deep run"
		}
		limit := math.Floor(float64(config.BadgeStaffShare * float64(slots)))
		r.Need, r.Limit = def.Threshold, limit
		r.Basis = fmt.Sprintf("%d%% of the %s %s slots %s's buildings hold", int(config.BadgeStaffShare*100), grouped(slots), def.Subject, where)
		switch {
		case def.Counter != "ev.staffed."+def.Subject || def.Event != config.BadgeEvCensus || !def.AnySubject:
			c.fail(def.Key, "staffing", fmt.Sprintf("%s must count the census's workers in its own domain.", label))
		case def.Threshold <= 0 || def.Threshold > limit:
			c.fail(def.Key, "staffing", fmt.Sprintf("%s needs %s workers in %s, and may ask for at most %s (%s).", label, grouped(int(def.Threshold)), def.Subject, grouped(int(limit)), r.Basis))
		case def.Threshold > housing:
			c.fail(def.Key, "staffing", fmt.Sprintf("%s needs %s workers, more than the %s people %s can house.", label, grouped(int(def.Threshold)), config.FormatAmount(housing), where))
		}

	case config.BadgeRuleHabit:
		r.Need, r.Basis = def.Threshold, "a habit of play, not a number of runs"
		habit := def.Counter == config.BadgeEvDayPlayed || def.Counter == config.BadgeEvReturned || def.Event == config.BadgeEvDayPlayed
		if !habit {
			c.fail(def.Key, "habit", fmt.Sprintf("%s claims to measure a habit, but counts neither days played nor returns.", label))
		}

	case config.BadgeRuleOccurs:
		r.Basis = def.Proof.Why
		if strings.TrimSpace(def.Proof.Why) == "" {
			c.fail(def.Key, "occurs", fmt.Sprintf("%s says only that it can happen. Say why a player can bring it about (config.Occurs).", label))
		}

	default:
		return false
	}
	return true
}

// subjectMissing checks the subject of a badge judged on an event against
// what that event is about, and says what is wrong ("" when nothing is): a
// subject the game does not have, or a subject on an event that has none.
// Either way the event would never match the row, and the badge could
// never be earned. A row that sets AnySubject is not held to its event's
// subject, so it is not checked here.
func (c *badgeCheck) subjectMissing(def config.BadgeDef) string {
	if def.Subject == "" || def.AnySubject {
		return ""
	}
	building := func() string {
		if _, ok := c.m.defs[def.Subject]; !ok {
			return fmt.Sprintf("names the building %q, which does not exist", def.Subject)
		}
		return ""
	}
	switch def.Event {
	case config.BadgeEvBuildingBuilt, config.BadgeEvBuilt, config.BadgeEvBuildingSold, config.BadgeEvBuildingUpgraded:
		return building()
	case config.BadgeEvWonderRaised, config.BadgeEvWonderBanked:
		for _, a := range c.m.ages {
			if c.set.Wonder(a.Key) == def.Subject {
				return ""
			}
		}
		return fmt.Sprintf("names the wonder %q, which is no age's wonder", def.Subject)
	case config.BadgeEvBuiltLineage:
		for _, d := range c.m.defs {
			if d.LineageKey == def.Subject {
				return ""
			}
		}
		return fmt.Sprintf("names the lineage %q, which no building belongs to", def.Subject)
	case config.BadgeEvResearchDone, config.BadgeEvMemoryAccepted, config.BadgeEvMemoryDeclined:
		if _, ok := c.set.Tech(def.Subject); !ok {
			return fmt.Sprintf("names the tech %q, which does not exist", def.Subject)
		}
	case config.BadgeEvMilestone:
		if _, ok := c.set.Milestone(def.Subject); !ok {
			return fmt.Sprintf("names the milestone %q, which does not exist", def.Subject)
		}
	case config.BadgeEvChain:
		for _, ch := range c.set.MilestoneChains() {
			if ch.Key == def.Subject {
				return ""
			}
		}
		return fmt.Sprintf("names the milestone chain %q, which does not exist", def.Subject)
	case config.BadgeEvGathered, config.BadgeEvBlackMarket, config.BadgeEvMarketTrade, config.BadgeEvProduced:
		if _, ok := c.m.resAge[def.Subject]; !ok {
			return fmt.Sprintf("names the resource %q, which does not exist", def.Subject)
		}
	case config.BadgeEvCivMet, config.BadgeEvCivAllied, config.BadgeEvWarEnded, config.BadgeEvWarStarted,
		config.BadgeEvWorkersLent, config.BadgeEvGiftSent, config.BadgeEvDeal:
		if _, ok := c.set.Faction(def.Subject); !ok {
			return fmt.Sprintf("names the civilization %q, which does not exist", def.Subject)
		}
	case config.BadgeEvCivStatus:
		if !slices.Contains([]string{"friendly", "allied", "embargo", "neutral"}, def.Subject) {
			return fmt.Sprintf("names the standing %q, which no civilization takes", def.Subject)
		}
	case config.BadgeEvHarbingerMet, config.BadgeEvAppeased, config.BadgeEvBraced, config.BadgeEvInvited:
		for _, a := range c.m.ages {
			if h, ok := c.set.Harbinger(a.Key); ok && h.Key == def.Subject {
				return ""
			}
		}
		return fmt.Sprintf("names the harbinger %q, which does not exist", def.Subject)
	case config.BadgeEvHarbingerResolved:
		verdicts := []string{game.HarbingerOutcomeFulfilled, game.HarbingerOutcomeVindicated, game.HarbingerOutcomeSpared, game.HarbingerOutcomeDiscredited}
		if !slices.Contains(verdicts, def.Subject) {
			return fmt.Sprintf("names the verdict %q, which no thread ends in (%v)", def.Subject, verdicts)
		}
	case config.BadgeEvEndured, config.BadgeEvSuccumbed, config.BadgeEvDoomNamed:
		if !c.set.CatastropheAllowed(def.Subject) {
			return fmt.Sprintf("asks for a doom in the era %q, where none can strike", def.Subject)
		}
	case config.BadgeEvEraLeft:
		if _, ok := c.set.Era(def.Subject); !ok {
			return fmt.Sprintf("names the era %q, which does not exist", def.Subject)
		}
	case config.BadgeEvExpedition:
		order := c.set.Indexes()
		mm := game.NewMilitaryManager()
		for _, a := range c.m.ages {
			for _, x := range mm.GetAvailableExpeditions(a.Key, order) {
				if x.Key == def.Subject {
					return ""
				}
			}
		}
		return fmt.Sprintf("names the expedition %q, which does not exist", def.Subject)
	case config.BadgeEvAwakening:
		for _, a := range c.m.ages {
			if aw, ok := c.set.Awakening(a.Key); ok && aw.Key == def.Subject {
				return ""
			}
		}
		return fmt.Sprintf("names the awakening %q, which does not exist", def.Subject)
	case config.BadgeEvEraEvent:
		for _, e := range append(c.set.GoodEraEvents(), c.set.ChallengingEraEvents()...) {
			if e.Type == def.Subject {
				return ""
			}
		}
		return fmt.Sprintf("names the kind of era event %q, which none is", def.Subject)
	case config.BadgeEvAdvancedTheme:
		if _, ok := theme.ByKey(def.Subject); !ok {
			return fmt.Sprintf("names the theme %q, which does not exist", def.Subject)
		}
	case config.BadgeEvAdvancedStyle:
		if !slices.Contains(config.BadgeMapStyles, def.Subject) {
			return fmt.Sprintf("names the map style %q, which does not exist", def.Subject)
		}
	case config.BadgeEvAdvancedGlyphs:
		if !slices.Contains(config.BadgeMapGlyphs, def.Subject) {
			return fmt.Sprintf("names the glyph set %q, which does not exist", def.Subject)
		}
	case config.BadgeEvLastPassage:
		if !slices.Contains([]string{"endured", "succumbed", "spared"}, def.Subject) {
			return fmt.Sprintf("asks for the Last Passage to end %q, which it cannot", def.Subject)
		}
	case config.BadgeEvAgeReached, config.BadgeEvPrestige, config.BadgeEvLegacyPrestige:
		if _, ok := c.m.idx[def.Subject]; !ok {
			return fmt.Sprintf("asks for the age %q, which does not exist", def.Subject)
		}
	case config.BadgeEvUpgradeBought, config.BadgeEvKit:
		if !slices.Contains(c.set.LegacyKit(), def.Subject) {
			return fmt.Sprintf("names the legacy kit item %q, which does not exist", def.Subject)
		}
	case config.BadgeEvBadge:
		// The event's subject is the family of the badge just earned.
		for _, d := range c.set.Badges() {
			if d.Family == def.Subject {
				return ""
			}
		}
		return fmt.Sprintf("names the badge family %q, which has no badge", def.Subject)
	case config.BadgeEvVisitor:
		// The map's own word for what was looked at: not checked here.
	default:
		// Everything else is reported with no subject: a row that names
		// one would wait for a subject that never comes.
		return fmt.Sprintf("names the subject %q, but the event %q is reported with none (set AnySubject if the badge is about %q and judged on every %s)", def.Subject, def.Event, def.Subject, def.Event)
	}
	return ""
}

// runTicks is the pacing targets, in ticks, of the run's ages from the age
// at place from on.
func (c *badgeCheck) runTicks(from int) float64 {
	sum := 0.0
	for i := max(from, 0); i <= c.m.runEnd; i++ {
		sum += c.set.TargetTicks(c.m.ages[i].Key)
	}
	return sum
}

// runEras is the eras a run plays through, and how many of them a doom can
// strike in.
func (c *badgeCheck) runEras() (eras, dooms int) {
	seen := map[string]bool{}
	for i := 0; i <= c.m.runEnd; i++ {
		if era := c.set.EraOf(c.m.ages[i].Key); !seen[era] {
			seen[era] = true
			eras++
			if c.set.CatastropheAllowed(era) {
				dooms++
			}
		}
	}
	return eras, dooms
}

// every counts how often something with a cooldown (in unstretched ticks)
// fits into the run's ages from place from on, each age's cooldown
// stretched as the game stretches it.
func (c *badgeCheck) every(from, cooldown int) float64 {
	n := 0.0
	for i := max(from, 0); i <= c.m.runEnd; i++ {
		age := c.m.ages[i].Key
		if wait := c.set.StretchTicks(age, cooldown); wait > 0 {
			n += float64(c.set.TargetTicks(age) / float64(wait))
		}
	}
	return math.Floor(n)
}

// The cooldowns of the game's own constants, in unstretched ticks. They
// are unexported there; the tests hold these to them
// (TestBadgeCooldownsMatchTheGame).
const (
	festivalCooldown    = 300
	blackMarketCooldown = 240
	dealRefresh         = 1800
)

// morePerRun is what one run adds to the counters the catalog's ladders
// climb.
func (c *badgeCheck) morePerRun(counter string) (perRun float64, basis string, ok bool) {
	m := c.m
	eras, dooms := c.runEras()
	ages := m.runEnd + 1
	if res, found := strings.CutPrefix(counter, config.BadgeEvProduced+"."); found {
		p, deep := c.set.RunProduction(res)
		basis = "one run's typical income of it, age by age, over each age's pacing target"
		if deep {
			basis = "its typical income over the pacing targets of its first two ages (it comes after a run's last age)"
		}
		return p, basis, true
	}
	if civ, found := strings.CutPrefix(counter, config.BadgeEvDeal+"."); found {
		if _, exists := c.set.Faction(civ); !exists {
			return 0, "no such civilization", true
		}
		return 1, "one deal a run from a civilization that has been met: its offers come round again every " + grouped(dealRefresh) + " ticks", true
	}
	switch counter {
	case config.BadgeEvEndured:
		return float64(dooms), fmt.Sprintf("the %d eras of a run a doom can strike in, each of which Invite makes certain", dooms), true
	case config.BadgeEvSuccumbed:
		return 1, "one Succumb ends a run", true
	case config.BadgeEvHarbingerResolved:
		return float64(eras), fmt.Sprintf("one thread in each of a run's %d eras", eras), true
	case config.BadgeEvAppeased, config.BadgeEvBraced:
		return float64(2 * eras), fmt.Sprintf("two levels in each of a run's %d threads", eras), true
	case config.BadgeEvInvited:
		return float64(dooms), fmt.Sprintf("one Invite in each of the %d eras of a run a doom can strike in", dooms), true
	case config.BadgeEvHarbingerResolved + ".discredited":
		// A false prophet comes to an era whose doom is not fated, at the
		// chance of its first age's harbinger.
		expect := 0.0
		seen := map[string]bool{}
		for i := 0; i <= m.runEnd; i++ {
			era := c.set.EraOf(m.ages[i].Key)
			if seen[era] {
				continue
			}
			seen[era] = true
			h, _ := c.set.Harbinger(m.ages[i].Key)
			quiet := 1.0
			if c.set.FateAllowed(era) {
				quiet = 1 - game.FateChance
			}
			expect += float64(quiet * h.FalseProphetChance)
		}
		return expect, "the chance of a false prophet in each of a run's eras, added up", true
	case config.BadgeEvWonderRaised:
		return float64(ages), fmt.Sprintf("the wonder of each of a run's %d ages", ages), true
	case config.BadgeEvResearchDone:
		return float64(m.techs[m.runEnd]), "the techs of a run's ages", true
	case config.BadgeEvMilestone:
		n := 0
		for _, ms := range c.set.Milestones() {
			if pos, known := m.idx[ms.MinAge]; ms.MinAge == "" || known && pos <= m.runEnd {
				n++
			}
		}
		return float64(n), "the milestones a run's ages open", true
	case config.BadgeEvChain:
		n := 0
		for _, ch := range c.set.MilestoneChains() {
			within := len(ch.MilestoneKeys) > 0
			for _, k := range ch.MilestoneKeys {
				ms, known := c.set.Milestone(k)
				pos, aged := m.idx[ms.MinAge]
				within = within && known && (ms.MinAge == "" || aged && pos <= m.runEnd)
			}
			if within {
				n++
			}
		}
		return float64(n), "the milestone chains a run's ages can finish", true
	case config.BadgeEvExpedition:
		// One party out at a time; the slowest expedition an age offers,
		// coming back with a success half the time.
		order := c.set.Indexes()
		mm := game.NewMilitaryManager()
		n := 0.0
		for i := 0; i <= m.runEnd; i++ {
			age := m.ages[i].Key
			slowest := 0
			for _, x := range mm.GetAvailableExpeditions(age, order) {
				slowest = max(slowest, x.DurationMax)
			}
			if slowest > 0 {
				n += float64(c.set.TargetTicks(age) / float64(c.set.StretchTicks(age, slowest)) / 2)
			}
		}
		return math.Floor(n), "one party out at a time on an age's slowest expedition, back with a success half the time", true
	case config.BadgeEvDeal:
		met := 0
		for _, f := range c.set.Factions() {
			if pos, known := m.idx[f.MinAge]; known && pos <= m.runEnd {
				met++
			}
		}
		return float64(met), fmt.Sprintf("one deal a run from each of the %d civilizations a run meets", met), true
	case config.BadgeEvRaidBlunted:
		return botRaidsBluntedPerRun, "about 30 a run, as the nightly bot's garrison did", true
	case config.BadgeEvBuildingUpgraded:
		return botUpgradesPerRun, "about 850 copies a run, as the nightly bot upgraded", true
	case config.BadgeEvPlanStarted:
		n := float64(config.BadgeRunShare * float64(m.builds[m.runEnd]))
		return n, fmt.Sprintf("%d%% of the %s buildings a run can hold, as for the lineage ladders", int(config.BadgeRunShare*100), grouped(m.builds[m.runEnd])), true
	case config.BadgeEvFestival:
		from := m.resAge["culture"]
		return c.every(from, festivalCooldown), "a festival every cooldown from the age culture comes in", true
	case config.BadgeEvBlackMarket:
		return c.every(m.idx["colonial_age"], blackMarketCooldown), "a deal every cooldown from the age the black market opens in", true
	case config.BadgeEvMemoryAccepted:
		return botMemoriesPerPrestige, "about one Ancient Memory every other prestige, as the nightly bot was offered", true
	case config.BadgeEvEraEvent + ".bad_challenging":
		return botHardTimesPerRun, "about 1.4 a run, as the nightly bot met", true
	case config.BadgeEvMarketTrade:
		return humanMarketTradesPerRun, "a few trades an age, typed by hand", true
	case config.BadgeEvLegacyPrestige:
		return 1, "one prestige a run, once Cosmic Legacy is held", true
	}
	return 0, "", false
}

// moreRunLimits is the most one run can do of what the catalog's run
// badges count.
func (c *badgeCheck) moreRunLimits(fact string) (limit float64, basis string, ok bool) {
	m := c.m
	switch fact {
	case "run." + config.BadgeEvWonderRaised:
		return float64(m.runEnd + 1), "the wonder of each of a run's ages", true
	case "run." + config.BadgeEvChain:
		return float64(len(c.set.MilestoneChains())), "the milestone chains there are: a milestone stays completed, so a run can finish them one after another", true
	case "run." + config.BadgeEvMilestone:
		// Every milestone the milestone check passes; a milestone stays
		// completed, so a run that goes to the last age can hold them all.
		n := 0
		for _, ms := range c.set.Milestones() {
			if _, problems := m.check(ms); len(problems) == 0 {
				n++
			}
		}
		return float64(n), "the milestones the milestone check proves, which a run completes one after another", true
	case "run." + config.BadgeEvDeal + ".*":
		// A civilization's offers come round every dealRefresh ticks, from
		// the age it is met in. The earliest of them has the most turns.
		best := 0.0
		for _, f := range c.set.Factions() {
			if pos, known := m.idx[f.MinAge]; known && pos <= m.runEnd {
				best = math.Max(best, c.every(pos, dealRefresh))
			}
		}
		return best, "one deal each time a civilization's offers come round, from the age it is met in", true
	case "run." + config.BadgeEvFestival:
		return c.every(m.resAge["culture"], festivalCooldown), "a festival every cooldown from the age culture comes in", true
	case "run." + config.BadgeEvGathered:
		return math.MaxInt32, "a hand gather is a command, and nothing limits it", true
	case "run." + config.BadgeEvEraEvent + ".bad_challenging":
		return float64(len(c.set.Eras())), "every era opens with an event, and a run can play every era", true
	}
	return 0, "", false
}
