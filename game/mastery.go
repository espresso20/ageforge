package game

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/espresso20/ageforge/config"
)

// Era Mastery (Pacing v2, PR 5). Each age remembers how many runs completed
// it (its mastery m, up to config.MasteryCap), and an age you know runs k
// times faster, k = 1 + √m (config.AgeSpeed, catch-up included):
//
//   - production: every resource's net rate × k, after the ×3 cap and
//     everything else (recalculateRates), shown as its own breakdown line;
//   - storage: × k, so the hours a store holds are the same at any k;
//   - build and research times: ÷ k, rounded up, at least one tick
//     (MasteryTicks);
//   - the fate window and the harbinger's lead: ÷ k (expectedAgeTicks).
//
// Events, raids, trade routes and the other real-clock timers are never
// divided. Mastery is fixed for a run: it changes only when a run ends, at a
// prestige or a Succumb (every age below the run's furthest gains one
// level), so k never changes mid-run and a fate rolled at an era's entry
// stays consistent. The record (the
// deepest age ever entered) can move mid-run, but only ages already behind
// the player feel it (catch-up), never the current one.
//
// The grace rule: when k drops (entering new ground, or leaving catch-up),
// storage shrinks with it. Stock already above the new cap is not cut: it
// stays until spent, production stops adding to it, and the grace ends once
// the resource is under its cap (ResourceManager.grace, saved as
// GameSave.OverCapGrace).

// ageKeys is every age key in order, built once (config rebuilds its tables
// on every call).
var ageKeys = sync.OnceValue(config.AgeOrder)

// MasteryTicks divides ticks by k, rounded up, never below one tick. k ≤ 1
// (the frontier, or a manager no engine has set) leaves ticks as they are.
func MasteryTicks(ticks int, k float64) int {
	if k <= 1 || ticks <= 0 {
		return ticks
	}
	return max(1, int(math.Ceil(float64(ticks)/k)))
}

// SpeedText formats a mastery speed for player text: "2x", "2.4x", "4.2x".
func SpeedText(k float64) string {
	return strconv.FormatFloat(math.Round(k*10)/10, 'f', -1, 64) + "x"
}

// shortAgeName is an age's name without " Age" ("Atomic"), for ranges.
func shortAgeName(key string) string {
	return strings.TrimSuffix(AgeName(key), " Age")
}

// ===== PrestigeManager: the mastery map =====

// rebuildSpeeds recomputes every age's k from the mastery map and the
// record. Called whenever either changes, so AgeSpeed (hit every tick and
// every snapshot) is one map lookup.
func (pm *PrestigeManager) rebuildSpeeds() {
	order := ageKeys()
	rec := -1
	if o, ok := ageOrders()[pm.record]; ok {
		rec = o
	}
	if pm.speeds == nil {
		pm.speeds = make(map[string]float64, len(order))
	}
	for i, a := range order {
		behind := -1
		if rec >= 0 {
			behind = rec - i
		}
		pm.speeds[a] = config.AgeSpeed(pm.mastery[a], behind)
	}
}

// AgeSpeed is how many times faster age runs (1 for an unknown age).
func (pm *PrestigeManager) AgeSpeed(age string) float64 {
	if k, ok := pm.speeds[age]; ok {
		return k
	}
	return 1
}

// Mastery is age's mastery level.
func (pm *PrestigeManager) Mastery(age string) int { return pm.mastery[age] }

// Record is the deepest age ever entered ("" before any).
func (pm *PrestigeManager) Record() string { return pm.record }

// RunFurthest is the deepest age entered this run.
func (pm *PrestigeManager) RunFurthest() string { return pm.runFurthest }

// NoteAgeEntered records that age was entered this run: it may become the
// run's furthest age and the record. Reports whether the record moved.
func (pm *PrestigeManager) NoteAgeEntered(age string) bool {
	o, ok := ageOrders()[age]
	if !ok {
		return false
	}
	if cur, ok := ageOrders()[pm.runFurthest]; !ok || o > cur {
		pm.runFurthest = age
	}
	if cur, ok := ageOrders()[pm.record]; !ok || o > cur {
		pm.record = age
		pm.rebuildSpeeds()
		return true
	}
	return false
}

// masteryGains is the ages a prestige now would raise: every age below the
// run's furthest that is not yet at the cap, in order.
func (pm *PrestigeManager) masteryGains() []string {
	far, ok := ageOrders()[pm.runFurthest]
	if !ok {
		return nil
	}
	var out []string
	for i, a := range ageKeys() {
		if i >= far {
			break
		}
		if pm.mastery[a] < config.MasteryCap {
			out = append(out, a)
		}
	}
	return out
}

// CommitRun is the mastery step of a run's end (a prestige or a Succumb):
// every age below this run's furthest gains one level, up to the cap, and
// the next run starts from the Primitive Age. Returns the ages that gained.
func (pm *PrestigeManager) CommitRun() []string {
	gained := pm.masteryGains()
	for _, a := range gained {
		pm.mastery[a] = config.ClampMastery(pm.mastery[a] + 1)
	}
	pm.runFurthest = ageKeys()[0]
	pm.rebuildSpeeds()
	return gained
}

// SetMastery sets age's mastery (clamped), for the dev console and tests.
func (pm *PrestigeManager) SetMastery(age string, m int) {
	if m = config.ClampMastery(m); m == 0 {
		delete(pm.mastery, age)
	} else {
		pm.mastery[age] = m
	}
	pm.rebuildSpeeds()
}

// SetRecord sets the record to age ("" clears it), for the dev console and
// tests.
func (pm *PrestigeManager) SetRecord(age string) {
	pm.record = age
	pm.rebuildSpeeds()
}

// LoadMastery restores the mastery state from a save. Unknown ages and
// out-of-range levels are dropped or clamped, as a hand-edited save could
// carry them.
func (pm *PrestigeManager) LoadMastery(mastery map[string]int, record, runFurthest string, seeded bool) {
	pm.mastery = make(map[string]int, len(mastery))
	for a, m := range mastery {
		if _, ok := ageOrders()[a]; ok && config.ClampMastery(m) > 0 {
			pm.mastery[a] = config.ClampMastery(m)
		}
	}
	pm.record, pm.runFurthest = "", ""
	if _, ok := ageOrders()[record]; ok {
		pm.record = record
	}
	if _, ok := ageOrders()[runFurthest]; ok {
		pm.runFurthest = runFurthest
	}
	pm.masterySeeded = seeded
	pm.rebuildSpeeds()
}

// masterySave is the mastery map for a save (nil when empty, so omitempty
// drops it).
func (pm *PrestigeManager) masterySave() map[string]int {
	if len(pm.mastery) == 0 {
		return nil
	}
	out := make(map[string]int, len(pm.mastery))
	for a, m := range pm.mastery {
		out[a] = m
	}
	return out
}

// MasterySnapshot is the mastery state for the UI, with current the age the
// player is in.
func (pm *PrestigeManager) MasterySnapshot(current string) MasteryState {
	st := MasteryState{
		K:           pm.AgeSpeed(current),
		Level:       pm.mastery[current],
		Record:      pm.record,
		RunFurthest: pm.runFurthest,
		Ages:        make(map[string]int, len(pm.mastery)),
		Speeds:      make(map[string]float64, len(pm.speeds)),
		NextGains:   pm.masteryGains(),
	}
	st.CatchUp = st.K > config.MasteryK(st.Level)
	for a, m := range pm.mastery {
		st.Ages[a] = m
	}
	for a, k := range pm.speeds {
		st.Speeds[a] = k
	}
	return st
}

// MasteryState is the Era Mastery view in a snapshot. The UI shows ages only
// up to Record (the no-spoiler rule): it never names an age past it.
type MasteryState struct {
	// K is the current age's speed and Level its mastery; CatchUp marks an
	// age running at catch-up speed rather than its own.
	K       float64
	Level   int
	CatchUp bool
	// Record is the deepest age ever entered; RunFurthest this run's.
	Record      string
	RunFurthest string
	// Ages is every age's mastery (ages at 0 left out); Speeds every age's
	// k, catch-up included.
	Ages   map[string]int
	Speeds map[string]float64
	// NextGains is the ages a prestige now would raise one level, in order.
	NextGains []string
}

// ===== Engine side =====

// speedK is the current age's speed. Must be called with the lock held.
func (ge *GameEngine) speedK() float64 { return ge.Prestige.AgeSpeed(ge.age) }

// noteGraceLocked applies the grace rule when the current age's k has
// dropped since the last rates pass: every resource is marked, and the
// storage pass that follows (recalculateRates) keeps the grace only where
// stock now sits above the cap. Must be called with the write lock held.
func (ge *GameEngine) noteGraceLocked(k float64) {
	if ge.lastK > 0 && k < ge.lastK {
		ge.Resources.markGraceAll()
	}
	ge.lastK = k
}

// masteryEntryLine is the log line on entering age from an age that ran at
// prevK: the speed on known ground, or that new ground runs at 1x when the
// age before it ran faster. "" when there is nothing to say.
func (ge *GameEngine) masteryEntryLine(age string, prevK float64) string {
	k := ge.Prestige.AgeSpeed(age)
	name := AgeName(age)
	switch {
	case k > config.MasteryK(ge.Prestige.Mastery(age)):
		return fmt.Sprintf("Known ground: the %s runs %s faster while you catch up to your record.", name, SpeedText(k))
	case k > 1:
		return fmt.Sprintf("Known ground: the %s runs %s faster (mastery %d).", name, SpeedText(k), ge.Prestige.Mastery(age))
	case prevK > 1:
		return fmt.Sprintf("New ground: the %s runs at 1x until a run completes it.", name)
	}
	return ""
}

// masteryCommitLine is the prestige's log line for the ages that gained a
// mastery level ("" when none did).
func masteryCommitLine(gained []string) string {
	switch len(gained) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("Era Mastery: the %s gained a mastery level and will run faster.", AgeName(gained[0]))
	}
	return fmt.Sprintf("Era Mastery: the %s to the %s gained a mastery level each and will run faster.",
		AgeName(gained[0]), AgeName(gained[len(gained)-1]))
}

// seedMasteryLocked is the one-time step that gives a save from before Era
// Mastery the mastery its past prestiges earned: each one was a Modern Age
// prestige, which completes the Primitive to the Atomic Age, so those ages
// get mastery min(10, level). The record becomes the deepest of the current
// age, the Modern Age (any prestige reached it), the Interstellar Age with
// the Cosmic Legacy held, and the first age of every epoch succumbed to.
// Runs once, after the signature check, and marks the save
// (PrestigeSave.MasterySeeded). A level-0 save gains nothing and logs
// nothing. Must be called with the write lock held.
func (ge *GameEngine) seedMasteryLocked() {
	pm := ge.Prestige
	if pm.masterySeeded {
		return
	}
	pm.masterySeeded = true
	if pm.level <= 0 {
		return
	}
	m := config.ClampMastery(pm.level)
	last := ""
	for _, a := range ageKeys() {
		if a == PrestigeRunAge {
			break
		}
		if pm.mastery[a] < m {
			pm.mastery[a] = m
		}
		last = a
	}
	pm.NoteAgeEntered(PrestigeRunAge)
	pm.runFurthest = "" // the run's own furthest comes from the run (noteRunAgesLocked)
	if ge.cosmicLegacy {
		pm.NoteAgeEntered("interstellar_age")
		pm.runFurthest = ""
	}
	for _, ep := range sortedKeys(ge.legacyBonuses) {
		if ages := config.EpochByKey()[ep].Ages; ge.legacyBonuses[ep] && len(ages) > 0 {
			pm.NoteAgeEntered(ages[0])
			pm.runFurthest = ""
		}
	}
	pm.rebuildSpeeds()
	ge.noteRunAgesLocked()
	ge.addLog("info", fmt.Sprintf("Era Mastery: ages you have completed now run faster. %s to %s: mastery %d (%s).",
		shortAgeName(ageKeys()[0]), shortAgeName(last), m, SpeedText(config.MasteryK(m))))
}

// noteRunAgesLocked makes sure the run's furthest age and the record cover
// the ages this run has reached (GameStats.AgesReached) and the current
// one. Idempotent; on load it fills them in for saves from before Era
// Mastery. Must be called with the write lock held.
func (ge *GameEngine) noteRunAgesLocked() {
	for _, a := range ge.Stats.AgesReached {
		ge.Prestige.NoteAgeEntered(a)
	}
	ge.Prestige.NoteAgeEntered(ge.age)
}

// SetMasteryForTest sets the mastery of every age in mastery and the record,
// as a veteran's save would carry them, and starts the run's grace tracking
// at the new speed. A test hook for other packages (the smoke suite's
// veteran preset); not reachable from play.
func (ge *GameEngine) SetMasteryForTest(mastery map[string]int, record string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	for _, a := range sortedKeys(mastery) {
		ge.Prestige.SetMastery(a, mastery[a])
	}
	ge.Prestige.SetRecord(record)
	ge.Prestige.NoteAgeEntered(ge.age)
	ge.recalculateRates()
}

// buildTicksLocked is def's construction time in the current age: its
// BuildTicks ÷ k on known ground. Must be called with the lock held.
func (ge *GameEngine) buildTicksLocked(def config.BuildingDef) int {
	return MasteryTicks(def.BuildTicks, ge.speedK())
}
