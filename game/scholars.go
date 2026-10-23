package game

import (
	"fmt"
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
)

// Scholars. Knowledge is not a thing a town holds: a knowledge building comes
// with its scholars (config.BuildingDef.Scholars a copy), who work from the
// moment the copy stands. They are people: they live in the homes and eat like
// anyone (population.go). The rules are in config/knowledge_rule.go.
//
//   - A tech has a price (the work it needs) and a progress. The work the
//     researching scholars do each tick goes straight into the tech in
//     progress, and the tech is done when the progress reaches the price.
//     One tech is in progress at a time. The time it takes is what is left
//     of its price over the work done each tick (researchTicks).
//   - The research setting says what share of the scholars research, 0 to
//     100 percent (default 100). The rest teach. While no tech is in progress
//     and none is queued, every scholar teaches.
//   - Teaching raises what workers make: with every scholar teaching, the
//     bonus is config.TeachingBonus of the town's age, in proportion to the
//     share of the scholars that teach. It is its own source in the worker
//     output pool (gather_rate).
//   - Everything that used to give knowledge (events, loot, milestones, the
//     Ancient Cache, routes) gives work to the tech in progress (grantResearch).

// scholarTeachSource is the name of the teaching bonus in the worker output
// pool's breakdown.
const scholarTeachSource = "scholars teaching"

// ScholarGroup is the scholars of one kind: the knowledge worker class of
// the age their building belongs to.
type ScholarGroup struct {
	Kind  string
	Count int
}

// ScholarsState is what a snapshot says of the scholars.
type ScholarsState struct {
	// Total is every scholar the town's knowledge buildings come with.
	Total int
	// Kinds is the scholars by kind, oldest age first.
	Kinds []ScholarGroup
	// Share is the research setting, a percent from 0 to 100.
	Share int
	// Researching and Teaching split Total: while a tech is in progress or
	// queued, Share percent of the scholars research (rounded to a whole
	// scholar) and the rest teach; with nothing to research all of them teach.
	Researching, Teaching int
	// PerTick is the work the researching scholars do each tick, with every
	// bonus.
	PerTick float64
	// TeachingBonus is the bonus to what workers make now, a fraction.
	// TeachingFull is what it would be with every scholar teaching.
	TeachingBonus, TeachingFull float64
}

// scholarShareOf is the research setting, 100 unless the player set it.
func (ge *GameEngine) scholarShareOf() int {
	if ge.scholarShareSet {
		return ge.scholarShare
	}
	return 100
}

// scholarGroups is the scholars by kind, oldest age first, and their total.
// A scholar's kind is the knowledge worker class of the age of the building
// he comes with (none if the age has no such class).
func (ge *GameEngine) scholarGroups() ([]ScholarGroup, int) {
	type row struct {
		tier int
		kind string
		n    int
	}
	var rows []row
	total := 0
	ge.Buildings.eachBuiltOf(ge.Buildings.scholarKeys, func(key string, count int, def config.BuildingDef) {
		if def.Scholars <= 0 || count <= 0 {
			return
		}
		n := count * def.Scholars
		total += n
		kind := def.Name
		if c, ok := ge.rules.WorkerClass("knowledge", def.RequiredAge); ok {
			kind = c.ClassName
		}
		rows = append(rows, row{tier: def.LineageTier, kind: kind, n: n})
	})
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].tier < rows[j].tier })
	var groups []ScholarGroup
	for _, r := range rows {
		if len(groups) > 0 && groups[len(groups)-1].Kind == r.kind {
			groups[len(groups)-1].Count += r.n
			continue
		}
		groups = append(groups, ScholarGroup{Kind: r.kind, Count: r.n})
	}
	return groups, total
}

// scholarTotal is every scholar the town has.
func (ge *GameEngine) scholarTotal() int {
	// A plain loop, not eachBuiltOf: this runs on every pass of the staffing
	// routine, which allocates nothing once the town is settled.
	total := 0
	bm := ge.Buildings
	for _, key := range bm.scholarKeys {
		if n := bm.counts[key]; n > 0 {
			if s := bm.defs[key].Scholars; s > 0 {
				total += n * s
			}
		}
	}
	return total
}

// researchQueued reports whether the plan holds a research item whose age
// has been reached: a tech the scholars would take up next.
func (ge *GameEngine) researchQueued() bool {
	currentAge, _ := ge.progress.AgeIndex(ge.age)
	for _, it := range ge.plan {
		if it.Kind != PlanResearch {
			continue
		}
		if def, ok := ge.rules.Tech(it.Key); ok {
			techAge, _ := ge.progress.AgeIndex(def.Age)
			if techAge <= currentAge {
				return true
			}
		}
	}
	return false
}

// scholarPosts splits n scholars into those who research and those who
// teach. With a tech in progress or queued, share percent research, rounded
// to a whole scholar (halves up); with none, every scholar teaches.
func (ge *GameEngine) scholarPosts(n int) (research, teach int) {
	if n <= 0 {
		return 0, 0
	}
	if ge.Research.currentTech == "" && !ge.researchQueued() {
		return 0, n
	}
	research = ge.shareResearchers(n)
	return research, n - research
}

// shareResearchers is how many of n scholars the share setting puts on
// research, a whole scholar (halves up), whether or not there is a tech to
// research.
func (ge *GameEngine) shareResearchers(n int) int {
	if n <= 0 {
		return 0
	}
	return min(max((n*ge.scholarShareOf()+50)/100, 0), n)
}

// teachingFull is the worker-output bonus with every scholar teaching, in
// the town's age.
func (ge *GameEngine) teachingFull() float64 {
	return ge.rules.TeachingBonus(ge.progress.ageIndex[ge.age])
}

// teachingBonus is the worker-output bonus the scholars give now: the full
// bonus of the age in proportion to the share of the scholars that teach.
func (ge *GameEngine) teachingBonus() float64 {
	n := ge.scholarTotal()
	if n <= 0 {
		return 0
	}
	_, teach := ge.scholarPosts(n)
	if teach == n {
		return ge.teachingFull()
	}
	return float64(float64(ge.teachingFull()*float64(teach)) / float64(n))
}

// scholarModifiers is the teaching bonus as a source in the worker output
// pool. It reads only held state.
func (ge *GameEngine) scholarModifiers() []Modifier {
	b := ge.teachingBonus()
	if b == 0 {
		return nil
	}
	return []Modifier{{Source: scholarTeachSource, Target: "gather_rate", Op: OpAdd, Value: b}}
}

// researchWork is the work all the town's scholars (and its wonders) would do
// each tick if every one of them researched, with every bonus to it but the
// speed-ups that shorten a time: the building figure, morale, the pools, the
// techs' bonus to research, an event's change, the Cosmic Legacy and Era
// Mastery. Under the write lock, from the last pass of recalculateRates.
func (ge *GameEngine) researchWork(r *Resolver, mMult, prodAllFactor float64, soft config.SoftCap) float64 {
	base := 0.0
	ge.Buildings.eachBuilt(func(key string, count int, def config.BuildingDef) {
		for _, eff := range def.Effects {
			if eff.Type == "research" {
				base += float64(eff.Value * float64(count))
			}
		}
	})
	ge.researchBase = base
	rate := float64(base * mMult)
	rate = float64(rate * prodAllFactor)
	rate = float64(rate * poolFactor("knowledge_rate", r.AddTotal("knowledge_rate"), soft))
	rate = float64(rate * ge.Research.OutputFactor(config.ResearchTarget))
	legacy := ge.cosmicLegacyFactor()
	// What an event sizes its change on: the work with no event running.
	ge.researchIncome = float64(rate * legacy)
	boost, flat := 0.0, 0.0
	for _, eff := range ge.Events.GetActiveEffects() {
		switch eff.Type {
		case "research_boost":
			boost += eff.Value
		case "research":
			flat += eff.Value
		}
	}
	rate = float64(rate * max(1+boost, 0))
	rate += flat
	if rate < 0 {
		rate = 0
	}
	if legacy != 1 {
		rate = float64(rate * legacy)
	}
	return float64(rate * ge.speedK())
}

// researchTimeFactor is what a tech's time is multiplied by: the research
// speed pool's cut, the techs' factor and Ancient Knowledge's factor. The
// work a tick is divided by it (the same cuts, as a faster tick).
func (ge *GameEngine) researchTimeFactor() float64 {
	t := 1.0
	if speed := ge.combinedResearchSpeed(); speed > 0 {
		t = float64(t * (1.0 - speed))
	}
	if f := ge.Research.TimeFactor(); f > 0 && f < 1 {
		t = float64(t * f)
	}
	if f := ge.succumbResearchFactor(); f > 0 && f < 1 {
		t = float64(t * f)
	}
	return max(t, 0.05)
}

// updateResearchRate works out the work the researching scholars do each
// tick (ge.researchPerTick). Called by recalculateRates.
func (ge *GameEngine) updateResearchRate(r *Resolver, mMult, prodAllFactor float64, soft config.SoftCap) {
	n := ge.scholarTotal()
	research, _ := ge.scholarPosts(n)
	// The work with every scholar researching: what the posts split.
	work := ge.researchWork(r, mMult, prodAllFactor, soft)
	work = float64(work / ge.researchTimeFactor())
	ge.researchFull = work
	ge.researchPerTick = ge.workOfResearchers(work, n, research)
	// What a tech started now would get: the share setting's scholars, the
	// idle override left out.
	ge.researchIfStarted = ge.workOfResearchers(work, n, ge.shareResearchers(n))
}

// workOfResearchers is the work each tick of `research` scholars out of n,
// given the work of all n.
func (ge *GameEngine) workOfResearchers(full float64, n, research int) float64 {
	if n <= 0 || research <= 0 {
		return 0
	}
	work := float64(full * float64(research) / float64(n))
	if ge.Research.slowed {
		work = float64(work * memoryResearchSlowdown)
	}
	return work
}

// researchTicksFor is how many ticks the work still to do takes at the
// work the researching scholars do each tick: 0 when none does any.
func (ge *GameEngine) researchTicksFor(work float64) int {
	if ge.researchPerTick <= 0 || work <= 0 {
		return 0
	}
	return int(math.Ceil(work / ge.researchPerTick))
}

// researchTicksIfStarted is how many ticks the work takes at the work the
// scholars the share setting puts on research would do: what a tech not in
// progress would take if it were started. 0 with no such scholar.
func (ge *GameEngine) researchTicksIfStarted(work float64) int {
	if ge.researchIfStarted <= 0 || work <= 0 {
		return 0
	}
	return int(math.Ceil(work / ge.researchIfStarted))
}

// ResearchTicksLeft is the time the tech in progress has left, in ticks: its
// remaining work over the work the researching scholars do each tick. 0 when
// no scholar is researching (it has no time to show).
func (ge *GameEngine) researchTicksLeft() int {
	if ge.Research.currentTech == "" {
		return 0
	}
	return ge.researchTicksFor(ge.Research.currentPrice() - ge.Research.progress)
}

// processResearch adds one tick of the scholars' work to the tech in progress.
func (ge *GameEngine) processResearch() {
	if ge.Research.currentTech == "" {
		return
	}
	// A tech whose progress already covers its price completes now, work or
	// no work (a load).
	if completed := ge.Research.CompleteIfFull(); completed != "" {
		ge.finishResearch(completed)
		return
	}
	if completed := ge.workResearch(ge.researchPerTick); completed != "" {
		ge.finishResearch(completed)
	}
}

// advanceResearch moves research on by n ticks at once (offline catch-up), at
// the efficiency eff the rest of the town works at, as live ticks would: a
// tech that finishes partway through gives up the rest of the span to the
// next one the plan starts, at the work a tick that holds after the
// completion, with the finished tech's effects in place. With nothing to
// start the scholars teach for what is left. starts collects what the plan
// starts on the way. Reports whether a tech completed.
func (ge *GameEngine) advanceResearch(n int, eff float64, starts *planStarts) bool {
	done := false
	for n > 0 && ge.Research.currentTech != "" {
		per := float64(eff * ge.researchPerTick)
		if per <= 0 {
			break
		}
		// The ticks the tech in progress still needs at this rate, or the
		// whole span if that is less.
		k := n
		if need := ge.Research.currentPrice() - ge.Research.progress; need > 0 {
			k = min(n, max(1, int(math.Ceil(need/per))))
		}
		completed := ge.workResearch(float64(per * float64(k)))
		n -= k
		if completed == "" {
			continue
		}
		done = true
		ge.finishResearch(completed)
		ge.recalculateRates()
		if n <= 0 {
			break
		}
		// The next tech starts in the tick its predecessor finished, at zero
		// progress, and the plan's next one starts a tick later, as on a live
		// tick.
		ge.runPlan(starts)
		for n > 0 && ge.Research.currentTech == "" && ge.runPlan(starts) {
			n--
		}
	}
	return done
}

// workResearch adds work to the tech in progress and counts it (the
// milestone that counts the work done). Returns the tech completed, if any.
func (ge *GameEngine) workResearch(w float64) string {
	if w <= 0 || ge.Research.currentTech == "" {
		return ""
	}
	ge.Stats.ResearchWork += w
	ge.Milestones.researchDone = ge.Stats.ResearchWork
	return ge.Research.AddWork(w)
}

// grantResearch gives work that did not come from the scholars' tick: an
// event, loot, a milestone, a route, the Ancient Cache. It lands on the tech
// in progress, at most half that tech's price (Research.Grant). With no tech
// in progress it is lost, and the log says so. It returns the line to say
// what it did ("" when nothing landed): the share of the tech in progress it
// gave ("12% of Banking") or that it finished the tech; it reads after
// "Gained". When the cap cut the gift, the log says the scholars gained some
// ground.
func (ge *GameEngine) grantResearch(work float64) string {
	if work <= 0 {
		return ""
	}
	key := ge.Research.currentTech
	if key == "" {
		ge.addLog("info", "Your scholars had nothing to study, so the insight was lost.")
		return ""
	}
	def := ge.Research.defs[key]
	before := ge.Research.progress
	ge.Stats.ResearchWork += work
	ge.Milestones.researchDone = ge.Stats.ResearchWork
	landed, completed := ge.Research.Grant(work)
	if landed < work {
		ge.addLog("info", fmt.Sprintf("Your scholars gained some ground on %s.", def.Name))
	}
	if completed != "" {
		ge.finishResearch(completed)
		return fmt.Sprintf("enough to complete %s", def.Name)
	}
	if def.Cost <= 0 {
		return ""
	}
	return fmt.Sprintf("%s of %s", shareOfTech((ge.Research.progress-before)/def.Cost), def.Name)
}

// shareOfTech is a fraction as a percent, "12%" (under 1% as "under 1%").
func shareOfTech(f float64) string {
	p := f * 100
	switch {
	case p <= 0:
		return "0%"
	case p < 1:
		return "under 1%"
	}
	return fmt.Sprintf("%d%%", int(math.Round(p)))
}

// SetResearchShare sets the share of the scholars that research, a percent
// from 0 to 100. The rest teach. It returns the line that says what the
// setting comes to now.
func (ge *GameEngine) SetResearchShare(percent int) (string, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if percent < 0 || percent > 100 {
		return "", fmt.Errorf("Scholars research between 0 and 100 percent. You asked for %d.", percent)
	}
	ge.scholarShare = percent
	ge.scholarShareSet = percent != 100
	ge.recalculateRates()
	n := ge.scholarTotal()
	research, teach := ge.scholarPosts(n)
	switch {
	case n == 0:
		return fmt.Sprintf("%d%% of your scholars will research. You have no scholars yet: knowledge buildings come with them.", percent), nil
	case ge.Research.currentTech == "" && !ge.researchQueued():
		return fmt.Sprintf("%d%% of your scholars will research. Nothing is being researched, so all %s teach.", percent, textfmt.Count(n, "scholar", "scholars")), nil
	}
	return fmt.Sprintf("%d%% of your scholars research: %d research and %d teach.", percent, research, teach), nil
}

// scholarsStateLocked is the scholars for a snapshot. Must be called with at
// least the read lock held.
func (ge *GameEngine) scholarsStateLocked() ScholarsState {
	groups, total := ge.scholarGroups()
	research, teach := ge.scholarPosts(total)
	return ScholarsState{
		Total: total, Kinds: groups, Share: ge.scholarShareOf(),
		Researching: research, Teaching: teach,
		PerTick:       ge.researchPerTick,
		TeachingBonus: ge.teachingBonus(), TeachingFull: ge.teachingFull(),
	}
}

// milestoneResources is what the milestones read as held: the stores, and the
// work done on techs as "research".
func (ge *GameEngine) milestoneResources() map[string]float64 {
	out := ge.Resources.GetAll()
	out[config.ResearchTarget] = ge.Stats.ResearchWork
	return out
}

// startedResearchLine is the line for a research that has begun: its time at
// the work the researching scholars do now, or that it needs scholars.
func (ge *GameEngine) startedResearchLine(def config.TechDef) string {
	if ticks := ge.researchTicksLeft(); ticks > 0 {
		return fmt.Sprintf("Started researching %s (%s).", def.Name, ge.durationLocked(ticks))
	}
	return fmt.Sprintf("Started researching %s. It needs scholars researching to go on.", def.Name)
}

// researchQueueKeys is the research items in the plan, in order.
func (ge *GameEngine) researchQueueKeys() []string {
	var out []string
	for _, it := range ge.plan {
		if it.Kind == PlanResearch {
			out = append(out, it.Key)
		}
	}
	return out
}

// withResearchTimes adds the engine's view to a snapshot of the research:
// the work each tick, the time the tech in progress has left and takes, the
// time each tech takes, the queue.
func (ge *GameEngine) withResearchTimes(rs ResearchState) ResearchState {
	rs.PerTick = ge.researchPerTick
	rs.Queue = ge.researchQueueKeys()
	if rs.CurrentTech != "" {
		rs.TicksLeft = ge.researchTicksLeft()
		rs.TotalTicks = ge.researchTicksFor(rs.Price)
		rs.NeedsScholars = ge.researchPerTick <= 0
	}
	for key, t := range rs.Techs {
		if key == rs.CurrentTech {
			t.TicksNeeded = ge.researchTicksFor(t.Cost)
		} else {
			t.TicksNeeded = ge.researchTicksIfStarted(t.Cost)
		}
		rs.Techs[key] = t
	}
	return rs
}

// scholarShareSave is the research setting for a save: nil when it is the
// default, so a town that never set it writes nothing.
func (ge *GameEngine) scholarShareSave() *int {
	if !ge.scholarShareSet {
		return nil
	}
	v := ge.scholarShare
	return &v
}

// loadResearch restores the research from a save. A tech that was paid for
// and waiting out its research time (the save has a wait left) is done: it
// completes on load. The setting comes back as it was saved.
func (ge *GameEngine) loadResearch(save *GameSave) {
	rs := save.Research
	paidAndWaiting := ""
	current := rs.CurrentTech
	if current != "" && rs.TotalTicks > 0 {
		paidAndWaiting, current = current, ""
	}
	ge.Research.LoadState(rs.Researched, current, rs.Progress, rs.Slowed)
	if paidAndWaiting != "" {
		if _, ok := ge.Research.defs[paidAndWaiting]; ok {
			ge.Research.researched[paidAndWaiting] = true
			ge.Research.rebuildBonuses()
			ge.loadedCompletions = append(ge.loadedCompletions, paidAndWaiting)
		}
	}
	ge.scholarShare, ge.scholarShareSet = 100, false
	if save.ScholarShare != nil && *save.ScholarShare >= 0 && *save.ScholarShare < 100 {
		ge.scholarShare, ge.scholarShareSet = *save.ScholarShare, true
	}
}

// turnKnowledgeIntoWork is what a save from before the scholars does with the
// knowledge it held: the stock becomes a gift of work for the tech in progress
// (Research.Grant: at most half that tech's price, the rest is lost). With no
// tech in progress it is lost. A tech that was paid for and waiting completes
// on load (loadResearch) before this runs.
func (ge *GameEngine) turnKnowledgeIntoWork(save *GameSave) {
	stock := save.Resources["knowledge"]
	if stock > 0 {
		if ge.Research.currentTech == "" {
			ge.addLog("info", "Your scholars had nothing to study, so the insight was lost.")
		} else if landed, _ := ge.Research.Grant(stock); landed < stock {
			ge.addLog("info", fmt.Sprintf("Your scholars gained some ground on %s.", ge.Research.defs[ge.Research.currentTech].Name))
		}
	}
	for _, key := range ge.loadedCompletions {
		ge.addLog(LogRoutine, fmt.Sprintf("Research complete: %s.", ge.rules.Name(rules.KindTech, key)))
	}
	ge.loadedCompletions = nil
}
