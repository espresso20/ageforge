package game

import (
	"math"
	"time"

	"github.com/espresso20/ageforge/vibe"
)

// The detail the plan panel shows for the item it has selected: one row for
// each resource the item's next start takes, what the town holds of it, what
// the items above hold back, the income, and the time until the rest is
// there. Built in planViews, where the reservations of the items above are
// in hand.

// How PlanItemView.StartsIn reads.
const (
	PlanWhenTimed   = ""        // StartsIn is a time
	PlanWhenNever   = "never"   // a resource it needs has no income
	PlanWhenHand    = "hand"    // it needs the player (a wonder's deposit)
	PlanWhenUnknown = "unknown" // it waits on something that cannot be timed
)

// PlanResourceRow is one resource of an item's next start.
type PlanResourceRow struct {
	Res string
	// Need is the price, Have what the town holds, Held what the items
	// above this one reserve of it.
	Need, Have, Held float64
	// Rate is the income per tick (Resources[k].Rate); PerSecond the same per
	// second of the game clock.
	Rate, PerSecond float64
	// ReadyIn is the time until Need is free after Held: 0 when it is
	// covered. Never is set instead when it is short and has no income.
	ReadyIn time.Duration
	Never   bool
}

// PlanPillar is one charter pillar as an advance item shows it.
type PlanPillar struct {
	Name    string
	Done    bool
	Percent int // 0 to 100; 99 at most until it is met
	// Blocker names what the pillar still lacks ("" once done); BlockerKey
	// is its building or tech key when it is one.
	Blocker, BlockerKey string
	// Item is the plan item (1-based) that builds or researches the blocker,
	// 0 when none does.
	Item int
}

// ticksToDuration is n ticks of the game clock as a time (capped so a very
// long wait cannot overflow).
func ticksToDuration(n float64, iv time.Duration) time.Duration {
	d := n * float64(iv)
	if d > float64(math.MaxInt64)/2 {
		return time.Duration(math.MaxInt64 / 2)
	}
	return time.Duration(d)
}

// planDetail fills v's rows, StartsIn and Bottleneck from cost, after the
// reservations in reserved (those of the items above v).
func (ge *GameEngine) planDetail(v *PlanItemView, cost, reserved map[string]float64) {
	iv := ge.tickIntervalLocked()
	var worst time.Duration
	never := false
	for _, res := range sortedKeys(cost) {
		need := cost[res]
		if need <= 0 {
			continue
		}
		row := PlanResourceRow{Res: res, Need: need, Have: ge.Resources.Get(res), Held: reserved[res], Rate: ge.Resources.GetRate(res)}
		if iv > 0 {
			row.PerSecond = row.Rate / iv.Seconds()
		}
		if short := row.Need + row.Held - row.Have; short > 0 {
			if row.Rate <= 0 {
				row.Never = true
			} else {
				row.ReadyIn = ticksToDuration(short/row.Rate, iv)
			}
		}
		switch {
		case row.Never && !never:
			never, v.Bottleneck = true, res
		case !never && row.ReadyIn > worst:
			worst, v.Bottleneck = row.ReadyIn, res
		}
		v.Rows = append(v.Rows, row)
	}
	v.StartsIn = worst
	if never {
		v.When = PlanWhenNever
	}
}

// planDetailFor sets the detail of a view whose check is chk: timed when it
// is ready or waits on a price, unknown when it waits on something else.
func (ge *GameEngine) planDetailFor(v *PlanItemView, chk planCheck, reserved map[string]float64) {
	ge.planDetail(v, chk.cost, reserved)
	if chk.reserve && len(chk.cost) > 0 {
		v.HeldBack = make(map[string]float64, len(chk.cost))
		for res, c := range chk.cost {
			if c > 0 {
				v.HeldBack[res] = c
			}
		}
	}
	if v.Status == PlanStatusBlocked && len(chk.cost) == 0 {
		v.When = PlanWhenUnknown
	}
	if v.Kind == PlanBuild && ge.Buildings.defs[v.Key].Category == "wonder" && v.Status != PlanStatusReady {
		v.When = PlanWhenHand
		v.BankShare = ge.wonderBankShare(v.Key)
		if v.Note == "" {
			v.Note = txtWonderByHand.String() + vibe.Line("game.plan_detail.planDetailFor.02", ". It also pays itself once the funds you hold cover the bank.")
		}
	}
}

// txtWonderByHand is a wonder item's note.
var txtWonderByHand = vibe.E("game.plan_detail.planDetailFor.01", "needs a deposit by hand")

// wonderBankShare is how full wonder w's bank is, 0 to 1.
func (ge *GameEngine) wonderBankShare(w string) float64 {
	if ge.Buildings.IsWonderBankFull(w) {
		return 1
	}
	def := ge.Buildings.defs[w]
	return WonderBankShare(ge.rules, ge.age, def.BaseCost, ge.Buildings.wonderBanks[w])
}

// planResearchTimes sets the StartsIn of the research items at idx (those
// that wait only for the slot): the first starts when the tech being
// researched is done, the next when that one is, and so on. Once a tech's
// length cannot be worked out (no scholars on research), the rest stay unknown.
func (ge *GameEngine) planResearchTimes(out []PlanItemView, idx []int) {
	iv := ge.tickIntervalLocked()
	ticks := float64(ge.researchTicksLeft())
	holds := ""
	if ge.Research.currentTech != "" {
		holds = techDefName(ge.rules, ge.Research.currentTech)
	}
	frees := ticksToDuration(ticks, iv)
	if ge.Research.currentTech != "" && ticks <= 0 {
		// Nobody is researching: the slot never frees.
		for _, i := range idx {
			out[i].When = PlanWhenUnknown
		}
		return
	}
	for n, i := range idx {
		v := &out[i]
		v.StartsIn, v.Holds, v.Frees = ticksToDuration(ticks, iv), holds, frees
		def, _ := ge.rules.Tech(v.Key)
		next := ge.researchTicksIfStarted(def.Cost)
		if next <= 0 {
			for _, j := range idx[n+1:] {
				out[j].When = PlanWhenUnknown
			}
			return
		}
		ticks += float64(next)
	}
}
