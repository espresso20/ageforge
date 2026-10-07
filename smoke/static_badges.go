package smoke

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
)

// The Badge Covenant: every badge can be earned. Each badge states how
// (config.BadgeProof), and this checks the statement against the game's
// own numbers, with the Milestone Covenant's model: a building is built
// only in its own age, so each copy must fit the most storage buildable
// there, at undiscounted prices.
//
// One run is the Primitive Age through the last age before a Modern Age
// prestige (game.PrestigeRunAge), played once. Lifetime thresholds are
// stated in runs, never in hours, so they follow the pacing targets.
//
// A badge fails when:
//
//   - it has no proof, or a proof this check does not know how to check;
//   - static "gate": its subject is not an age a run can advance into;
//   - static "copies": it asks for more than config.BadgeCopiesShare of the
//     copies of a building that can stand in the building's own age, or is
//     not tied to that age;
//   - static "lifetime": a run adds nothing to its counter, or the
//     threshold takes more than config.BadgeMaxRuns runs. A run adds one
//     prestige, and config.BadgeRunShare of a lineage's ceiling (every tier
//     up to the run's last age raised to its storage limit);
//   - static "time": the threshold is not a multiple of an age's pacing
//     target, through a predicate that reads it as one;
//   - static "run_count": it asks for more of something than a run can do;
//   - it names an event, a counter, a fact, a predicate or an age the game
//     does not have (the subject cannot occur);
//   - bot: the style is not one a smoke bot plays;
//   - spoilers: text that shows before the badge is earned names an age
//     later than the one the badge is revealed at;
//   - text: a dash or an exclamation mark in its name, description or hint;
//   - integrity: an integrity badge that carries a tier, or is not on the
//     explicit list (IntegrityBadges), or a listed key that is not one.
//
// A failure names the number that breaks.

// IntegrityBadges are the badges exempt from the proof check: they record
// something about the account or a save, are worth nothing and count toward
// nothing. Listed here so that marking a badge as integrity to skip the
// check takes an edit to this list too.
var IntegrityBadges = []string{
	"special.hand_in_the_cookie_jar",
	"special.creative_accounting",
}

// LegacyAchievements are the four account achievements of account.json
// version 1. Each must stay the alias of a badge, or accounts that hold it
// lose it on migration.
var LegacyAchievements = []string{"first_prestige", "prestige_x10", "reached_iron", "reached_modern"}

// BadgeBotStyles are the play styles a bot-proof may name.
var BadgeBotStyles = []string{"greedy", "idle", "harbinger", "succumber", "cosmic", "army", "taste"}

// BadgeProblem is a badge the static check proves out of reach, or a fault
// in the badge tables, with the reason in plain words.
type BadgeProblem struct {
	Key string `json:"key"`
	// Kind is the check that fails: "proof", "gate", "copies", "lifetime",
	// "time", "run_count", "existence", "bot", "spoiler", "text",
	// "integrity", "alias" or "table".
	Kind string `json:"kind"`
	Why  string `json:"why"`
}

// BadgeReach is one badge's verdict: what it asks for against the most the
// proof allows.
type BadgeReach struct {
	Key   string  `json:"key"`
	Name  string  `json:"name"`
	Proof string  `json:"proof"`
	Need  float64 `json:"need,omitempty"`
	Limit float64 `json:"limit,omitempty"`
	// Basis says what Limit is, in words.
	Basis string `json:"basis,omitempty"`
}

// StaticBadges checks every badge of the core ruleset against the Badge
// Covenant. It returns what breaks it and every badge's verdict, in catalog
// order.
func StaticBadges() ([]BadgeProblem, []BadgeReach) {
	return staticBadges(rules.Core(), config.Badges(), config.BadgeFamilies(), config.BuildingByKey(), game.PrestigeRunAge)
}

// staticBadges is StaticBadges over the given ruleset and tables (the tests
// hand it broken ones).
func staticBadges(set *rules.Set, written []config.BadgeDef, families []config.BadgeFamilyDef, defs map[string]config.BuildingDef, prestigeAge string) ([]BadgeProblem, []BadgeReach) {
	c := &badgeCheck{
		set:    set,
		m:      newMilestoneModel(defs, prestigeAge),
		events: config.BadgeEventKinds(),
		preds:  game.BadgePredNames(),
	}
	for _, why := range rules.BadgeProblems(written, families, set) {
		c.fail("", "table", why)
	}
	badges := set.Badges()
	seenRung := map[string]string{}
	for _, def := range badges {
		c.check(def)
		if def.Scope == config.BadgeLifetime {
			rung := fmt.Sprintf("%s at %v", def.Counter, def.Threshold)
			if other, dup := seenRung[rung]; dup {
				c.fail(def.Key, "table", fmt.Sprintf("%s (%s) and %s both ask for %s: one of them can never be the next rung.", def.Name, def.Key, other, rung))
			}
			seenRung[rung] = def.Key
		}
	}
	for _, key := range IntegrityBadges {
		if def, ok := set.Badge(key); !ok || !def.Integrity() {
			c.fail(key, "integrity", fmt.Sprintf("%s is on the list of integrity badges but is not one in the catalog.", key))
		}
	}
	for _, old := range LegacyAchievements {
		if _, ok := set.BadgeForAlias(old); !ok {
			c.fail(old, "alias", fmt.Sprintf("No badge lists the old account achievement %q as an alias, so an account that holds it would lose it.", old))
		}
	}
	return c.problems, c.reach
}

type badgeCheck struct {
	set      *rules.Set
	m        *milestoneModel
	events   []string
	preds    []string
	problems []BadgeProblem
	reach    []BadgeReach
}

func (c *badgeCheck) fail(key, kind, why string) {
	c.problems = append(c.problems, BadgeProblem{Key: key, Kind: kind, Why: why})
}

// check runs every rule on one badge.
func (c *badgeCheck) check(def config.BadgeDef) {
	label := fmt.Sprintf("%s (%s)", def.Name, def.Key)
	r := BadgeReach{Key: def.Key, Name: def.Name}
	c.text(def, label)
	c.exists(def, label)
	c.spoilers(def, label)

	switch def.Proof.Kind {
	case config.BadgeProofIntegrity:
		r.Proof = "integrity"
		if !slices.Contains(IntegrityBadges, def.Key) {
			c.fail(def.Key, "integrity", fmt.Sprintf("%s is marked as an integrity badge, which exempts it from the proof check, but it is not on the explicit list (IntegrityBadges).", label))
		}
		if def.Tier != config.BadgeNoTier || def.Points() != 0 {
			c.fail(def.Key, "integrity", fmt.Sprintf("%s is an integrity badge with a tier: integrity badges are worth nothing.", label))
		}
	case config.BadgeProofDerived:
		r.Proof = "derived"
		if def.Scope != config.BadgeLifetime || !strings.HasPrefix(def.Counter, config.BadgeEvBadge) {
			c.fail(def.Key, "proof", fmt.Sprintf("%s claims to follow from other badges, but does not count earned badges.", label))
		}
	case config.BadgeProofBot:
		r.Proof = "bot:" + def.Proof.Rule
		if !slices.Contains(BadgeBotStyles, def.Proof.Rule) {
			c.fail(def.Key, "bot", fmt.Sprintf("%s is to be proven by a bot playing the %q style, which no smoke bot plays (%s).", label, def.Proof.Rule, strings.Join(BadgeBotStyles, ", ")))
		}
	case config.BadgeProofStatic:
		r.Proof = "static:" + def.Proof.Rule
		c.static(def, label, &r)
	default:
		c.fail(def.Key, "proof", fmt.Sprintf("%s has no proof: say how it can be earned (config.BadgeDef.Proof).", label))
	}
	c.reach = append(c.reach, r)
}

// static checks a proof from config.
func (c *badgeCheck) static(def config.BadgeDef, label string, r *BadgeReach) {
	m := c.m
	switch def.Proof.Rule {
	case config.BadgeRuleGate:
		// Every gate is proven reachable by the Gate Covenant
		// (TestStaticGates); what is checked here is that the badge
		// hangs on one.
		pos, ok := m.idx[def.Subject]
		switch {
		case def.Event != config.BadgeEvAgeReached && len(def.When) == 0:
			c.fail(def.Key, "gate", fmt.Sprintf("%s rests on the age gates but is not judged on an age being reached.", label))
		case def.Event == config.BadgeEvAgeReached && (!ok || pos == 0):
			c.fail(def.Key, "gate", fmt.Sprintf("%s asks for the age %q to be reached, which no advance does.", label, def.Subject))
		}

	case config.BadgeRuleCopies:
		d, ok := m.defs[def.Subject]
		if !ok {
			c.fail(def.Key, "copies", fmt.Sprintf("%s counts copies of %q, which is not a building.", label, def.Subject))
			return
		}
		ceil := m.ceiling(def.Subject)
		if ceil.age < 0 {
			c.fail(def.Key, "copies", fmt.Sprintf("%s counts %s, which can never be built.", label, pluralName(def.Subject)))
			return
		}
		own := m.ages[ceil.age].Key
		if def.Counter != "standing."+def.Subject || def.InAge != own {
			c.fail(def.Key, "copies", fmt.Sprintf("%s must count %s standing in the %s, the only age they can be built in (Counter %q, InAge %q).",
				label, pluralName(def.Subject), m.ageName(ceil.age), def.Counter, def.InAge))
		}
		limit := math.Floor(float64(config.BadgeCopiesShare * float64(ceil.n)))
		r.Need, r.Limit = def.Threshold, limit
		r.Basis = fmt.Sprintf("%d%% of the %d %s that can stand in the %s", int(config.BadgeCopiesShare*100), ceil.n, pluralName(def.Subject), m.ageName(ceil.age))
		if def.Threshold > limit {
			why := fmt.Sprintf("%s needs %s %s in the %s, but at most %d can stand there", label, grouped(int(def.Threshold)), pluralName(def.Subject), m.ageName(ceil.age), ceil.n)
			if ceil.res != "" {
				why += fmt.Sprintf(" (copy #%d costs %s %s, over the most %s storage buildable there, %s)",
					ceil.n+1, config.FormatAmount(ceil.price), game.ResourceName(ceil.res), game.ResourceName(ceil.res), config.FormatAmount(ceil.store))
			} else if d.MaxCount > 0 {
				why += " (its max count)"
			}
			c.fail(def.Key, "copies", fmt.Sprintf("%s, and a copies badge may ask for at most %d%% of that (%s).", why, int(config.BadgeCopiesShare*100), grouped(int(limit))))
		}

	case config.BadgeRuleLifetime:
		if def.Scope != config.BadgeLifetime {
			c.fail(def.Key, "lifetime", fmt.Sprintf("%s claims a lifetime proof but is not judged on a lifetime counter.", label))
			return
		}
		perRun, basis, ok := c.perRun(def.Counter)
		if !ok {
			c.fail(def.Key, "lifetime", fmt.Sprintf("%s counts %q, and this check does not know what one run adds to it. Teach it (badgeCheck.perRun) with the proof of what a run adds.", label, def.Counter))
			return
		}
		r.Need, r.Limit, r.Basis = def.Threshold, float64(config.BadgeMaxRuns*perRun), fmt.Sprintf("%d runs of %s", config.BadgeMaxRuns, basis)
		switch {
		case perRun <= 0:
			c.fail(def.Key, "lifetime", fmt.Sprintf("%s counts %q, but a run adds nothing to it (%s).", label, def.Counter, basis))
		case def.Threshold <= 0:
			c.fail(def.Key, "lifetime", fmt.Sprintf("%s asks for a count of %v.", label, def.Threshold))
		case def.Threshold > r.Limit:
			c.fail(def.Key, "lifetime", fmt.Sprintf("%s needs %s, which is %.1f runs at %s a run (%s). A ladder's top rung may take at most %d.",
				label, config.FormatAmount(def.Threshold), def.Threshold/perRun, config.FormatAmount(perRun), basis, config.BadgeMaxRuns))
		}

	case config.BadgeRuleTime:
		target := c.set.TargetTicks(def.InAge)
		r.Need, r.Basis = def.Threshold, "times the pacing target of its age"
		switch {
		case def.Pred != config.BadgePredAgeOverstay:
			c.fail(def.Key, "time", fmt.Sprintf("%s has a time threshold, which must be read as a multiple of an age's pacing target (the predicate %q), never as hours.", label, config.BadgePredAgeOverstay))
		case def.InAge == "" || target <= 0:
			c.fail(def.Key, "time", fmt.Sprintf("%s measures time in the age %q, which has no pacing target.", label, def.InAge))
		case def.Threshold <= 0:
			c.fail(def.Key, "time", fmt.Sprintf("%s asks for %v times the target.", label, def.Threshold))
		}

	case config.BadgeRuleRunCount:
		limit, basis, ok := c.runLimit(def.Counter)
		if !ok {
			c.fail(def.Key, "run_count", fmt.Sprintf("%s counts %q in one run, and this check does not know how often a run can do that. Teach it (badgeCheck.runLimit).", label, def.Counter))
			return
		}
		r.Need, r.Limit, r.Basis = def.Threshold, limit, basis
		if def.Threshold > limit {
			c.fail(def.Key, "run_count", fmt.Sprintf("%s needs %s in one run, but the most a run can do is %s (%s).", label, grouped(int(def.Threshold)), grouped(int(limit)), basis))
		}

	default:
		c.fail(def.Key, "proof", fmt.Sprintf("%s claims the static proof %q, which this check does not have.", label, def.Proof.Rule))
	}
}

// perRun is what one run adds to a lifetime counter, and what that number
// is in words. ok is false for a counter the check has no proof for.
func (c *badgeCheck) perRun(counter string) (perRun float64, basis string, ok bool) {
	m := c.m
	switch {
	case counter == config.BadgeEvPrestige:
		return 1, "one prestige", true
	case strings.HasPrefix(counter, config.BadgeEvBuiltLineage+"."):
		lineage := strings.TrimPrefix(counter, config.BadgeEvBuiltLineage+".")
		total, deep := 0, 0
		for _, k := range sortedKeys(m.defs) {
			if m.defs[k].LineageKey != lineage {
				continue
			}
			ceil := m.ceiling(k)
			if ceil.age < 0 {
				continue
			}
			deep += ceil.n
			if ceil.age <= m.runEnd {
				total += ceil.n
			}
		}
		basis = fmt.Sprintf("%d%% of the %s %s buildings a run can hold through the %s", int(config.BadgeRunShare*100), grouped(total), lineage, m.ageName(m.runEnd))
		if total == 0 {
			// A lineage that starts after the run's last age is measured
			// in deep runs: every tier it has.
			total = deep
			basis = fmt.Sprintf("%d%% of the %s %s buildings a deep run can hold", int(config.BadgeRunShare*100), grouped(total), lineage)
		}
		return float64(config.BadgeRunShare * float64(total)), basis, true
	}
	return 0, "", false
}

// runLimit is the most of a fact one run can reach, and what the number is
// in words.
func (c *badgeCheck) runLimit(fact string) (limit float64, basis string, ok bool) {
	m := c.m
	switch fact {
	case "run." + config.BadgeEvBuildingSold:
		// A copy can be sold once for each time it is built, and selling
		// opens in the second age. Half of what a run can build, as for
		// the build-count milestones.
		if m.runEnd < 1 {
			return 0, "selling opens in the second age", true
		}
		sellable := 0
		for _, k := range sortedKeys(m.defs) {
			d := m.defs[k]
			if d.Category == "wonder" || d.Category == "storage" {
				continue
			}
			if ceil := m.ceiling(k); ceil.age >= 0 && ceil.age <= m.runEnd {
				sellable += ceil.n
			}
		}
		return math.Floor(float64(MilestoneBuildShare * float64(sellable))), fmt.Sprintf("%d%% of the %s buildings a run can build and sell", int(MilestoneBuildShare*100), grouped(sellable)), true
	}
	return 0, "", false
}

// exists checks that everything a badge names is something the game has.
func (c *badgeCheck) exists(def config.BadgeDef, label string) {
	miss := func(what string) {
		c.fail(def.Key, "existence", fmt.Sprintf("%s %s, so it can never be earned.", label, what))
	}
	kind := func(name string) bool {
		for _, ev := range c.events {
			if name == ev || strings.HasPrefix(name, ev+".") {
				return true
			}
		}
		return false
	}
	fact := func(name string) {
		switch {
		case strings.HasPrefix(name, "run."):
			tally := strings.TrimPrefix(name, "run.")
			if !kind(tally) {
				miss(fmt.Sprintf("reads the run fact %q, which counts no event the game reports", name))
			}
			for _, ev := range config.BadgeSessionEvents() {
				if tally == ev || strings.HasPrefix(tally, ev+".") {
					miss(fmt.Sprintf("reads the run fact %q, but %s is about the session, not the run, and is never tallied in one", name, ev))
				}
			}
		case strings.HasPrefix(name, "life."):
			if !kind(strings.TrimPrefix(name, "life.")) {
				miss(fmt.Sprintf("reads the counter %q, which counts no event the game reports", name))
			}
		case strings.HasPrefix(name, "standing."):
			if _, ok := c.m.defs[strings.TrimPrefix(name, "standing.")]; !ok {
				miss(fmt.Sprintf("counts the building %q, which does not exist", strings.TrimPrefix(name, "standing.")))
			}
		case strings.HasPrefix(name, "ev."):
		default:
			miss(fmt.Sprintf("reads the fact %q, which is none the game keeps (run., standing., ev. or life.)", name))
		}
	}
	if def.Key == "" || def.Name == "" || def.Desc == "" {
		miss("has no key, name or description")
	}
	switch def.Scope {
	case config.BadgeLifetime:
		if def.Counter == "" || !kind(def.Counter) {
			miss(fmt.Sprintf("counts %q, which no event the game reports adds to", def.Counter))
		}
		if def.Event != "" || def.InAge != "" || len(def.When) > 0 || def.Pred != "" {
			miss("is a lifetime badge with a condition only an event can meet; a lifetime badge is its counter and threshold")
		}
	default:
		if !slices.Contains(c.events, def.Event) {
			miss(fmt.Sprintf("is judged on the event %q, which the game does not report", def.Event))
		}
		if def.Counter != "" {
			if def.Scope == config.BadgeMoment {
				miss("is a moment badge with a counter; a badge that counts within the run is a run badge")
			}
			fact(def.Counter)
		}
	}
	for _, cond := range def.When {
		fact(cond.Fact)
	}
	if def.InAge != "" {
		if _, ok := c.m.idx[def.InAge]; !ok {
			miss(fmt.Sprintf("must be earned in the age %q, which does not exist", def.InAge))
		}
	}
	if def.Pred != "" && !slices.Contains(c.preds, def.Pred) {
		miss(fmt.Sprintf("asks for the predicate %q, which the engine does not have", def.Pred))
	}
	if t := def.Reward.Theme; t != "" {
		if _, ok := theme.ByKey(t); !ok {
			c.fail(def.Key, "existence", fmt.Sprintf("%s gives the theme %q, which does not exist.", label, t))
		}
	}
	switch def.Reveal.Kind {
	case config.BadgeRevealAtAge, config.BadgeRevealNextAge:
		if _, ok := c.m.idx[def.Reveal.Key]; !ok {
			miss(fmt.Sprintf("is revealed at the age %q, which does not exist, so it would stay hidden", def.Reveal.Key))
		}
	case config.BadgeRevealOnCounter:
		if !kind(def.Reveal.Key) {
			miss(fmt.Sprintf("is revealed by the counter %q, which no event adds to, so it would stay hidden", def.Reveal.Key))
		}
	case config.BadgeSecret:
		if def.Hint == "" && !def.Integrity() {
			miss("is secret and has no hint to list under")
		}
	}
}

// spoilers checks that text which shows before a badge is earned names no
// age later than the age the badge is revealed at. A secret badge shows
// only its hint, which may name no age past the first.
func (c *badgeCheck) spoilers(def config.BadgeDef, label string) {
	limit, text := 0, def.Name+" "+def.Desc
	switch def.Reveal.Kind {
	case config.BadgeVisible:
	case config.BadgeRevealAtAge, config.BadgeRevealNextAge:
		limit = c.m.idx[def.Reveal.Key]
	case config.BadgeSecret:
		text = def.Hint
	default:
		return // revealed by something met in play, not by an age
	}
	for i, a := range c.m.ages {
		if i > limit && strings.Contains(text, a.Name) {
			c.fail(def.Key, "spoiler", fmt.Sprintf("%s names the %s in text that shows from the %s on, before the player can have seen that age named.", label, a.Name, c.m.ageName(limit)))
		}
	}
}

// text checks a badge's text for the punctuation player text does not use.
func (c *badgeCheck) text(def config.BadgeDef, label string) {
	for what, s := range map[string]string{"name": def.Name, "description": def.Desc, "hint": def.Hint} {
		switch {
		case strings.ContainsAny(s, "—–"):
			c.fail(def.Key, "text", fmt.Sprintf("%s has a dash in its %s: %q.", label, what, s))
		case strings.Contains(s, "!"):
			c.fail(def.Key, "text", fmt.Sprintf("%s has an exclamation mark in its %s: %q.", label, what, s))
		case strings.ContainsAny(s, "{}"):
			c.fail(def.Key, "text", fmt.Sprintf("%s has an unfilled template field in its %s: %q.", label, what, s))
		}
	}
}

// writeBadges adds the Badge Covenant's table to the static report.
func writeBadges(sb *strings.Builder, problems []BadgeProblem, reach []BadgeReach) {
	sb.WriteString("\n### Badges\n\n")
	if len(problems) == 0 {
		fmt.Fprintf(sb, "All %d badges can be earned.\n\n", len(reach))
	} else {
		for _, p := range problems {
			fmt.Fprintf(sb, "- **%s** (%s): %s\n", p.Key, p.Kind, p.Why)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("| badge | proof | asks | limit | basis |\n|---|---|---|---|---|\n")
	for _, r := range reach {
		need, limit := "", ""
		if r.Need > 0 {
			need = config.FormatAmount(r.Need)
		}
		if r.Limit > 0 {
			limit = config.FormatAmount(r.Limit)
		}
		fmt.Fprintf(sb, "| %s (`%s`) | %s | %s | %s | %s |\n", r.Name, r.Key, r.Proof, need, limit, r.Basis)
	}
}
