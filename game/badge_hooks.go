package game

import (
	"github.com/espresso20/ageforge/config"
)

// badge_hooks.go holds the reports that need a little arithmetic before
// they are made: the pace of a run and of an era, what the account is
// wearing at an advance, the production batches and the census. Like every
// report they run under the engine's write lock, read the game and change
// nothing in it: no resource, no roll of the run's RNG, no log line. A run
// plays the same with them or without.

// tell reports an event to the account alone: it is judged, and never
// tallied in the run's facts (config.BadgeSessionEvents). Callers hold
// ge.mu for writing.
func (ge *GameEngine) tell(ev Event) { ge.judgeBadges(ev) }

// targetsBefore is the sum of the pacing targets, in ticks, of every age
// before age.
func (ge *GameEngine) targetsBefore(age string) float64 {
	pos, ok := ge.rules.Index(age)
	if !ok {
		return 0
	}
	sum := 0.0
	for _, k := range ge.rules.AgeKeys()[:pos] {
		sum += ge.rules.TargetTicks(k)
	}
	return sum
}

// eraPace is the ticks the run has spent in era over the pacing targets of
// the era's ages. known is false when the run does not know when it entered
// the era (a save from before that was kept).
func (ge *GameEngine) eraPace(era string) (pace float64, known bool) {
	e, ok := ge.rules.Era(era)
	if !ok || !ge.runFacts.EraKnown {
		return 0, false
	}
	sum := 0.0
	for _, a := range e.Ages {
		sum += ge.rules.TargetTicks(a)
	}
	if sum <= 0 {
		return 0, false
	}
	return float64(ge.tick-ge.runFacts.EraTick) / sum, true
}

// reportEraLeft reports that the run has left era, with its pace when the
// run knows it.
func (ge *GameEngine) reportEraLeft(era string) {
	ev := Event{Kind: config.BadgeEvEraLeft, Subject: era}
	if pace, ok := ge.eraPace(era); ok {
		ev.Attrs = map[string]float64{"pace": pace}
	}
	ge.report(ev)
}

// reportAgeReached reports an advance from oldAge into newAge: the age
// itself, with the run's pace so far (its ticks over the pacing targets of
// the ages behind it) when the run was kept from its first tick; the era
// left, when the advance crosses into a new one; and what the account is
// wearing.
func (ge *GameEngine) reportAgeReached(oldAge, newAge string) {
	ev := Event{Kind: config.BadgeEvAgeReached, Subject: newAge}
	if sum := ge.targetsBefore(newAge); ge.runFacts.Whole && sum > 0 {
		ev.Attrs = map[string]float64{"pace": float64(ge.tick) / sum}
	}
	if from, to := ge.rules.EraOf(oldAge), ge.rules.EraOf(newAge); from != to && from != "" {
		ge.reportEraLeft(from)
		ge.runFacts.EraTick, ge.runFacts.EraKnown = ge.tick, true
	}
	ge.report(ev)
	ge.ageTold = newAge
	ge.tellLooks()
}

// noteAgeJump tells the account of an age the run is in without having
// advanced into it: the developer console's jump, which sets the age and
// reports nothing. Every age the jump passed is told, in order, so a run
// that jumps ahead earns what it would have earned on the way (the age
// badges, and the themes they give). The run's own facts are left alone: no
// advance happened in it. A run that has just been loaded or begun is where
// it is, and a move back to an earlier age tells nothing.
func (ge *GameEngine) noteAgeJump() {
	if ge.age == ge.ageTold {
		return
	}
	told := ge.ageTold
	ge.ageTold = ge.age
	from, known := ge.rules.Index(told)
	to, ok := ge.rules.Index(ge.age)
	if ge.account == nil || !known || !ok || to <= from {
		return
	}
	for _, age := range ge.rules.AgeKeys()[from+1 : to+1] {
		ge.tell(Event{Kind: config.BadgeEvAgeReached, Subject: age})
	}
}

// tellLooks tells the account what it is wearing as the run advances: its
// theme, its map style and its map glyphs. They are the account's own
// settings, so the run's facts never hold them.
func (ge *GameEngine) tellLooks() {
	acct := ge.account
	if acct == nil {
		return
	}
	theme, style, glyphs := acct.looks()
	ge.tell(Event{Kind: config.BadgeEvAdvancedTheme, Subject: theme})
	ge.tell(Event{Kind: config.BadgeEvAdvancedStyle, Subject: style})
	ge.tell(Event{Kind: config.BadgeEvAdvancedGlyphs, Subject: glyphs})
}

// The looks an account has when it has chosen none.
const (
	defaultThemeKey  = "forge"
	defaultMapStyle  = "roguelike"
	defaultMapGlyphs = "unicode"
)

// looks is the theme, map style and map glyphs the account wears, with the
// defaults for what it never chose.
func (a *Account) looks() (theme, style, glyphs string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	theme, style, glyphs = a.Prefs.ActiveTheme, a.Prefs.MapStyle, a.Prefs.MapGlyphs
	if theme == "" {
		theme = defaultThemeKey
	}
	if style == "" {
		style = defaultMapStyle
	}
	if glyphs == "" {
		glyphs = defaultMapGlyphs
	}
	return theme, style, glyphs
}

// doomInvited reports whether era's doom was invited.
func (ge *GameEngine) doomInvited(era string) bool {
	return ge.fate != nil && ge.fate.EpochKey == era && ge.fate.Invited
}

// ----- production -----

// noteProduced adds what the tick's rates produce to the batch, and sends
// the batch every config.BadgeProducedTicks ticks. Only a held account
// counts production, so a game without one does no work here. scale is the
// ticks the rates stand for (1 on a live tick).
func (ge *GameEngine) noteProduced(scale float64) {
	if ge.account == nil || ge.badges == nil || !ge.badges.countsProduction {
		return
	}
	if ge.produced == nil {
		ge.produced = make(map[string]float64, len(ge.Resources.order))
	}
	for _, key := range ge.Resources.order {
		if r := ge.Resources.resources[key]; r.Rate > 0 && ge.Resources.unlocked[key] {
			ge.produced[key] += float64(r.Rate * scale)
		}
	}
	if ge.tick-ge.producedTick >= config.BadgeProducedTicks {
		ge.flushProduced()
	}
}

// flushProduced reports the batch and empties it.
func (ge *GameEngine) flushProduced() {
	ge.producedTick = ge.tick
	for _, key := range ge.Resources.order {
		if n := ge.produced[key]; n > 0 {
			ge.produced[key] = 0
			ge.tell(Event{Kind: config.BadgeEvProduced, Subject: key, N: n})
		}
	}
}

// ----- the census -----

// noteCensus reports the state of the run every config.BadgeCensusTicks
// ticks, for the badges that ask how things stand rather than what
// happened. Only a held account is told.
func (ge *GameEngine) noteCensus() {
	if ge.account == nil || ge.badges == nil || ge.tick%config.BadgeCensusTicks != 0 {
		return
	}
	if !ge.badges.listens(config.BadgeEvCensus, "") {
		return
	}
	attrs := make(map[string]float64, len(ge.badges.domains)+2)
	staffed := map[string]int{}
	rt := ge.Workers.domains["worker"]
	if rt != nil {
		for building, n := range rt.assignments {
			if def, ok := ge.Workers.buildingDefs[building]; ok && def.WorkerDomain != "" {
				staffed[def.WorkerDomain] += n
			}
		}
	}
	for _, d := range ge.badges.domains {
		attrs["staffed."+d] = float64(staffed[d])
	}
	full, capped := true, 0
	for _, key := range ge.Resources.order {
		r := ge.Resources.resources[key]
		if !ge.Resources.unlocked[key] || r.Storage <= 0 || ge.rules.IsFlowResource(key) {
			continue
		}
		capped++
		if r.Amount < r.Storage {
			full = false
		}
	}
	attrs["full"] = boolFact(full && capped > 0)
	if pos, ok := ge.rules.Index(ge.age); ok {
		attrs["age"] = float64(pos)
	}
	ge.tell(Event{Kind: config.BadgeEvCensus, Attrs: attrs})
}

// notePlanSize reports how many items wait in the build plan, after one
// was added.
func (ge *GameEngine) notePlanSize() {
	ge.report(Event{Kind: config.BadgeEvPlanQueued, Attrs: map[string]float64{"size": float64(len(ge.plan))}})
}

// noteWars reports the wars that started and ended since the last call, in
// the order they happened. tribute marks the endings a tribute bought.
func (ge *GameEngine) noteWars(tribute bool) {
	for _, w := range ge.Diplomacy.takeWarEvents() {
		if w.started {
			ge.report(Event{Kind: config.BadgeEvWarStarted, Subject: w.faction, Attrs: map[string]float64{"wars": float64(w.wars)}})
			continue
		}
		ge.report(Event{Kind: config.BadgeEvWarEnded, Subject: w.faction, Attrs: map[string]float64{"tribute": boolFact(tribute)}})
	}
}

// NoteVisitorInspected reports that the player inspected a visitor on the
// map (kind names it: "saucer"). The UI calls it; it takes the write lock.
// A visitor is the map's own, not the run's, so only the account is told.
func (ge *GameEngine) NoteVisitorInspected(kind string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.tell(Event{Kind: config.BadgeEvVisitor, Subject: kind})
}
