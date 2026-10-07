package game

import (
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// badges.go is the one place the game says what happened, and where both
// record layers are judged from it:
//
//   - milestones belong to the run: they pay out inside it and reset with
//     it. They are judged on the tick event, from the run's state, exactly
//     where the tick always checked them.
//   - badges belong to the account: earned once, kept for good, worth
//     nothing inside a run. They are judged on every event, from the
//     event, the run's facts and the account's lifetime counters.
//
// The engine calls report at each hook site, under its write lock. Nothing
// here does I/O, draws from the run's RNG or calls back into the engine's
// locking methods, so a report can sit anywhere in a tick. This is not the
// Bus: bus handlers exist for the UI and run in subscription order, and
// what earns a badge must not depend on who subscribed first.
//
// A badge never changes a run's numbers, so a seed plays the same run on
// any account.

// Event is one thing that happened in the game.
type Event struct {
	// Kind is one of config's BadgeEv constants.
	Kind string
	// Subject is what it happened to: an age, a building, an outcome.
	Subject string
	// N is how much; 0 reads as 1.
	N float64
	// Attrs are the event's own numbers, read by a badge as ev.<name>.
	Attrs map[string]float64
}

// boolFact is a yes or no as an event attribute: 1 or 0.
func boolFact(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func (e Event) amount() float64 {
	if e.N == 0 {
		return 1
	}
	return e.N
}

// The fact prefixes a badge row reads (config.BadgeCond).
const (
	factRun      = "run."
	factStanding = "standing."
	factEvent    = "ev."
	factLife     = "life."
)

// badgeBook is a ruleset's badges, indexed for judging. It never changes;
// the engine builds one for its ruleset.
type badgeBook struct {
	set  *rules.Set
	defs []config.BadgeDef
	// byKey is a badge's place in defs.
	byKey map[string]int
	// byEvent is the Run and Moment badges judged on an event kind.
	byEvent map[string][]int
	// byCounter is the Lifetime badges on an account counter, lowest
	// threshold first.
	byCounter map[string][]int
	// counters is the account counters kept: the ones a badge names, as its
	// counter, in a condition or in its reveal rule. No other is stored.
	counters map[string]bool
	// runTallies is the per-subject run tallies kept ("building_built.hut").
	// A tally by kind alone is always kept.
	runTallies map[string]bool
}

// newBadgeBook indexes set's badges.
func newBadgeBook(set *rules.Set) *badgeBook {
	b := &badgeBook{
		set:        set,
		defs:       set.Badges(),
		byKey:      map[string]int{},
		byEvent:    map[string][]int{},
		byCounter:  map[string][]int{},
		counters:   map[string]bool{},
		runTallies: map[string]bool{},
	}
	fact := func(name string) {
		switch {
		case strings.HasPrefix(name, factLife):
			b.counters[strings.TrimPrefix(name, factLife)] = true
		case strings.HasPrefix(name, factRun):
			b.runTallies[strings.TrimPrefix(name, factRun)] = true
		}
	}
	for i, d := range b.defs {
		if _, taken := b.byKey[d.Key]; taken {
			continue
		}
		b.byKey[d.Key] = i
		switch d.Scope {
		case config.BadgeLifetime:
			if d.Counter != "" {
				b.byCounter[d.Counter] = append(b.byCounter[d.Counter], i)
				b.counters[d.Counter] = true
			}
		default:
			if d.Event != "" {
				b.byEvent[d.Event] = append(b.byEvent[d.Event], i)
			}
			fact(d.Counter)
		}
		for _, c := range d.When {
			fact(c.Fact)
		}
		if d.Reveal.Kind == config.BadgeRevealOnCounter && d.Reveal.Key != "" {
			b.counters[d.Reveal.Key] = true
		}
	}
	for _, list := range b.byCounter {
		sort.SliceStable(list, func(x, y int) bool { return b.defs[list[x]].Threshold < b.defs[list[y]].Threshold })
	}
	return b
}

// def returns a badge's definition, nil for a key the ruleset does not have.
func (b *badgeBook) def(key string) *config.BadgeDef {
	if i, ok := b.byKey[key]; ok {
		return &b.defs[i]
	}
	return nil
}

// listens reports whether anything is judged on, or counted from, an event:
// when nothing is, a report costs two map reads and no lock.
func (b *badgeBook) listens(kind, subject string) bool {
	if len(b.byEvent[kind]) > 0 || b.counters[kind] {
		return true
	}
	return subject != "" && b.counters[kind+"."+subject]
}

// tracksBuilt reports whether a badge counts first-time builds of def (by
// building or by lineage), so the run must keep its build marks.
func (b *badgeBook) tracksBuilt(def config.BuildingDef) bool {
	if b.counters[config.BadgeEvBuilt] || b.counters[config.BadgeEvBuilt+"."+def.Key] {
		return true
	}
	if def.LineageKey == "" {
		return false
	}
	return b.counters[config.BadgeEvBuiltLineage] || b.counters[config.BadgeEvBuiltLineage+"."+def.LineageKey]
}

// RunFacts is what a run remembers about itself for the records: how often
// each event happened, and the marks that keep a sold and rebuilt building
// from counting twice. It is saved with the run and starts over with it
// (prestige, Succumb, a new game).
type RunFacts struct {
	// Whole is true when the facts were kept from the run's first tick. A
	// save from before run facts loads without it, and a badge that asks
	// for something never to have happened in the run cannot be earned
	// there: the run's past is not known.
	Whole bool `json:"whole,omitempty"`
	// Counts is how often each event was reported, by kind, and by kind and
	// subject for the subjects a badge names.
	Counts map[string]float64 `json:"counts,omitempty"`
	// Net is the copies of a building the run has built, less the copies
	// sold; High is the most Net has been. A build counts toward a lifetime
	// build counter only when it takes Net past High. Kept for the
	// buildings a badge counts.
	Net  map[string]int `json:"net,omitempty"`
	High map[string]int `json:"high,omitempty"`
	// AgeTick is the tick the run entered its current age, when AgeKnown.
	AgeTick  int  `json:"age_tick,omitempty"`
	AgeKnown bool `json:"age_known,omitempty"`
}

// newRunFacts is the facts of a run at its first tick.
func newRunFacts() RunFacts { return RunFacts{Whole: true, AgeKnown: true} }

// empty reports whether nothing is recorded: such facts are left out of the
// save, so a save from before them keeps its bytes when written again.
func (f *RunFacts) empty() bool {
	return !f.Whole && !f.AgeKnown && f.AgeTick == 0 && len(f.Counts) == 0 && len(f.Net) == 0 && len(f.High) == 0
}

// clone returns a copy that shares no map with f.
func (f RunFacts) clone() RunFacts {
	f.Counts = cloneFloatMap(f.Counts)
	f.Net = cloneIntMap(f.Net)
	f.High = cloneIntMap(f.High)
	return f
}

func cloneFloatMap(m map[string]float64) map[string]float64 {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]float64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneIntMap(m map[string]int) map[string]int {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]int, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// runFactsSaveCopy is the run's facts for the save: a copy, or nil when
// there is nothing to write. Callers hold ge.mu.
func (ge *GameEngine) runFactsSaveCopy() *RunFacts {
	if ge.runFacts.empty() {
		return nil
	}
	f := ge.runFacts.clone()
	return &f
}

// note tallies an event in the run.
func (f *RunFacts) note(book *badgeBook, ev Event) {
	if f.Counts == nil {
		f.Counts = map[string]float64{}
	}
	f.Counts[ev.Kind] += ev.amount()
	if ev.Subject != "" {
		if name := ev.Kind + "." + ev.Subject; book.runTallies[name] {
			f.Counts[name] += ev.amount()
		}
	}
}

// built records n copies of a building finished and returns how many took
// it past the most the run had built before: the copies that count.
func (f *RunFacts) built(key string, n int) int {
	if f.Net == nil {
		f.Net, f.High = map[string]int{}, map[string]int{}
	}
	f.Net[key] += n
	fresh := f.Net[key] - f.High[key]
	if fresh <= 0 {
		return 0
	}
	f.High[key] = f.Net[key]
	return fresh
}

// sold records n copies of a building sold.
func (f *RunFacts) sold(key string, n int) {
	if f.Net == nil {
		f.Net, f.High = map[string]int{}, map[string]int{}
	}
	f.Net[key] -= n
}

// report says that something happened. It tallies the event in the run,
// judges the run's milestones (on the tick) and the account's badges.
// Callers hold ge.mu for writing.
func (ge *GameEngine) report(ev Event) {
	switch ev.Kind {
	case config.BadgeEvTick:
		// The run layer. Milestones read the run's state, and completing
		// one reports it in turn.
		ge.checkMilestones()
	case config.BadgeEvAgeReached:
		ge.runFacts.AgeTick, ge.runFacts.AgeKnown = ge.tick, true
		ge.runFacts.note(ge.badges, ev)
	default:
		ge.runFacts.note(ge.badges, ev)
	}
	ge.judgeBadges(ev)
}

// note is report for an event of one, with no attributes.
func (ge *GameEngine) note(kind, subject string) { ge.report(Event{Kind: kind, Subject: subject}) }

// noteBuilt reports n finished copies of a building: the completion itself,
// a wonder raised, and, for a building a badge counts, the copies that took
// it past the most the run had built (so selling and rebuilding adds
// nothing to a lifetime count). Callers hold ge.mu for writing.
func (ge *GameEngine) noteBuilt(key string, n int) {
	if n <= 0 {
		return
	}
	def, ok := ge.rules.Building(key)
	if !ok {
		return
	}
	if ge.badges.tracksBuilt(def) {
		if fresh := ge.runFacts.built(key, n); fresh > 0 {
			ge.report(Event{Kind: config.BadgeEvBuilt, Subject: key, N: float64(fresh)})
			if def.LineageKey != "" {
				ge.report(Event{Kind: config.BadgeEvBuiltLineage, Subject: def.LineageKey, N: float64(fresh)})
			}
		}
	}
	if def.Category == "wonder" {
		ge.report(Event{Kind: config.BadgeEvWonderRaised, Subject: key, N: float64(n)})
	}
	ge.report(Event{Kind: config.BadgeEvBuildingBuilt, Subject: key, N: float64(n)})
}

// noteSold reports n copies of a building sold.
func (ge *GameEngine) noteSold(key string, n int) {
	if n <= 0 {
		return
	}
	if def, ok := ge.rules.Building(key); ok && ge.badges.tracksBuilt(def) {
		ge.runFacts.sold(key, n)
	}
	ge.report(Event{Kind: config.BadgeEvBuildingSold, Subject: key, N: float64(n)})
}

// judgeBadges judges the held account's badges on an event. A run that
// does not record to the account (the developer console changed it) earns
// nothing and moves no counter; only the integrity badges are judged then.
// A run that belongs to another account is not this account's at all.
// Callers hold ge.mu.
func (ge *GameEngine) judgeBadges(ev Event) {
	acct := ge.account
	if acct == nil || ge.badges == nil {
		return
	}
	if ge.runAccountID != "" && ge.runAccountID != acct.AccountID {
		return
	}
	if !ge.badges.listens(ev.Kind, ev.Subject) {
		return
	}
	acct.judge(ge.badges, ev, ge.badgeCtxLocked(ev))
}

// badgeCtx is what a badge is judged against besides the account itself.
type badgeCtx struct {
	// clean is false for a run the developer console has changed.
	clean bool
	// crossed marks what is earned in a save edited outside the game.
	crossed bool
	// run is the save the badge is earned in.
	run string
	age string
	// fact reads a run, standing or event fact; known is false for a fact
	// the run cannot vouch for.
	fact func(name string) (value float64, known bool)
	// pred runs a named predicate.
	pred func(def *config.BadgeDef) bool
}

// badgeCtxLocked is the engine's side of a judgment. Callers hold ge.mu.
func (ge *GameEngine) badgeCtxLocked(ev Event) badgeCtx {
	return badgeCtx{
		clean:   ge.accountForRecordsLocked() != nil,
		crossed: ge.cheaterBadge,
		run:     ge.activeSaveName,
		age:     ge.age,
		fact: func(name string) (float64, bool) {
			switch {
			case strings.HasPrefix(name, factRun):
				return ge.runFacts.Counts[strings.TrimPrefix(name, factRun)], ge.runFacts.Whole
			case strings.HasPrefix(name, factStanding):
				return float64(ge.Buildings.GetCount(strings.TrimPrefix(name, factStanding))), true
			case strings.HasPrefix(name, factEvent):
				v, ok := ev.Attrs[strings.TrimPrefix(name, factEvent)]
				return v, ok
			}
			return 0, false
		},
		pred: func(def *config.BadgeDef) bool {
			p := badgePreds[def.Pred]
			return p != nil && p(ge, def)
		},
	}
}

// badgePreds are the named predicates a badge row can ask for
// (config.BadgeDef.Pred). Each reads the engine under its lock and changes
// nothing.
var badgePreds = map[string]func(ge *GameEngine, def *config.BadgeDef) bool{
	// The run has spent Threshold times its current age's pacing target in
	// that age.
	config.BadgePredAgeOverstay: func(ge *GameEngine, def *config.BadgeDef) bool {
		target := ge.rules.TargetTicks(ge.age)
		if !ge.runFacts.AgeKnown || target <= 0 || def.Threshold <= 0 {
			return false
		}
		return float64(ge.tick-ge.runFacts.AgeTick) >= float64(def.Threshold*target)
	},
}

// BadgePredNames lists the predicates the engine knows, for the guard: a
// badge that names another can never be earned.
func BadgePredNames() []string { return sortedKeys(badgePreds) }

// ReportForTest reports an event as the game would, for tests outside this
// package (the UI's): a prestige, an age reached, a building built. It
// changes nothing else about the run.
func (ge *GameEngine) ReportForTest(kind, subject string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.note(kind, subject)
}

// noteDayLocked records today as a day the account was played on, when a
// game is started or loaded, and reports the day when it is a new one. The
// date is the player's own calendar day. A run that does not record to the
// account leaves no day.
//
// The day belongs to the account, not to the run: it goes to the account's
// badges only and is not tallied in the run's facts, which must be the same
// on any account and on any date. Callers hold ge.mu for writing.
func (ge *GameEngine) noteDayLocked() {
	acct := ge.accountForRecordsLocked()
	if acct == nil {
		return
	}
	if acct.noteDay(time.Now().Format(accountDayLayout)) {
		ge.judgeBadges(Event{Kind: config.BadgeEvDayPlayed})
	}
}

// accountDayLayout is how a day is written in account.json: "2026-10-07".
const accountDayLayout = "2006-01-02"

// NoteDevUnlocked reports that the developer console was unlocked. The UI
// calls it when the passphrase is accepted. It is a fact about the account's
// session, not the run, so it is not tallied in the run's facts.
func (ge *GameEngine) NoteDevUnlocked() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.judgeBadges(Event{Kind: config.BadgeEvDevUnlocked})
}

// DrainEarnedBadges returns the badges earned since the last call, as they
// now show, and forgets them. The dashboard calls it from its refresh,
// outside the engine's lock, and writes the toast and the log line.
func (ge *GameEngine) DrainEarnedBadges() []BadgeView {
	ge.mu.RLock()
	acct, book := ge.account, ge.badges
	ge.mu.RUnlock()
	if acct == nil || book == nil {
		return nil
	}
	keys := acct.drainEarned()
	if len(keys) == 0 {
		return nil
	}
	out := make([]BadgeView, 0, len(keys))
	for _, k := range keys {
		if v, ok := acct.badgeView(book, k); ok {
			out = append(out, v)
		}
	}
	return out
}

// BadgeLogLine is the log line for a badge just earned. It is fixed per
// badge: no quip, so earning one draws nothing from the run's streams.
func BadgeLogLine(v BadgeView) string {
	line := "Badge earned: " + v.Name
	if v.Tier != "" {
		line += " (" + v.Tier + ")"
	}
	return line + ". " + v.Desc
}

// Badges returns the held account's badges as the player may see them, and
// the totals. Locked badges the spoiler rules withhold come back as
// silhouettes: the text is not in the view at all. nil without an account.
func (ge *GameEngine) Badges() ([]BadgeView, BadgeSummary) {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.badgesLocked()
}

// badgesLocked is Badges under ge.mu (read or write). It takes the
// account's lock; never call it from a bus handler.
func (ge *GameEngine) badgesLocked() ([]BadgeView, BadgeSummary) {
	if ge.account == nil || ge.badges == nil {
		return nil, BadgeSummary{}
	}
	return ge.account.badgeViews(ge.badges, ge.ageSightLocked())
}
