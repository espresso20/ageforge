package smoke

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// The economy's properties: what must be true of the numbers for the game to
// be in proportion however it is played. Each is worked out from the tables
// alone, for the reference players in reference.go.
//
// A property that does not hold today is listed in KnownPropertyFailures with
// what it reads today. The check fails when a listed property starts to hold
// (take it off the list), when an unlisted one stops holding, or when a
// listed one reads differently from what the list says (a number moved:
// write the new reading down on purpose). So the list only changes when
// somebody means it to, and the work ahead is to make it shorter.

// Property is one of them, as checked.
type Property struct {
	N     int    `json:"n"`
	Says  string `json:"says"`
	Holds bool   `json:"holds"`
	// Today is the reading: the number that shows whether it holds.
	Today string `json:"today"`
}

// KnownPropertyFailures is every property that does not hold today, with
// today's reading. Later changes to the economy shorten it.
var KnownPropertyFailures = map[int]string{
	1:  "22 buildings climb at another rate (1.13)",
	2:  "fails in 5 of 22 ages; worst 0.13 of the age in the Primitive Age",
	3:  "fails in 6 of 22 ages; worst 4.6 times in the Primitive Age",
	4:  "fails in 1 of 22 ages; worst 2.7 times in the Quantum Age",
	5:  "fails in 11 of 22 ages; worst 13 techs in the Digital Age",
	8:  "fails in 17 of 22 ages; worst 0.0% eaten in the Transcendent Age",
	9:  "fails in 22 of 22 ages; worst room for 19553 people per job in the Transcendent Age",
	10: "fails in 20 of 22 ages; worst 6.4 times the age in the Electric Age",
	11: "fails in 20 of 22 ages; worst 483727 times what the buildings make in the Information Age",
	12: "27 of 27 trade route payments fall outside it",
	15: "in 11 ages the shortest warning is under 8 hours; the last of them is the Information Age",
	16: "fails in 4 of 22 ages; worst 4.6 times in the Primitive Age",
	17: "fails in 12 of 22 ages; worst 13 techs in the Digital Age",
	18: "not worked out here yet",
	19: "a full run earns 3279 points and everything for sale costs 99",
}

// Properties is the whole check: the properties, and how long each age
// lasts for each reference player.
type Properties struct {
	List      []Property `json:"list"`
	Ages      []string   `json:"ages"`
	Target    []float64  `json:"target_hours"`
	Ordinary  []float64  `json:"ordinary_hours"`
	Lingering []float64  `json:"lingering_hours"`
	CheckIn   []float64  `json:"check_in_hours"`
}

// Problems is what fails the build: a property whose state is not the one
// written down.
func (p Properties) Problems() []string {
	var out []string
	for _, pr := range p.List {
		known, listed := KnownPropertyFailures[pr.N]
		switch {
		case pr.Holds && listed:
			out = append(out, fmt.Sprintf("property %d now holds (%s): take it out of KnownPropertyFailures", pr.N, pr.Says))
		case !pr.Holds && !listed:
			out = append(out, fmt.Sprintf("property %d no longer holds (%s): %s", pr.N, pr.Says, pr.Today))
		case !pr.Holds && known != pr.Today:
			out = append(out, fmt.Sprintf("property %d reads differently than KnownPropertyFailures records (%s): recorded %q, today %q", pr.N, pr.Says, known, pr.Today))
		}
	}
	return out
}

// The shares the properties are held to.
const (
	propWallIncome   = 2.5  // income at the storage wall, over the reference town's
	propWallStore    = 2.0  // the store at the wall, over the reference store
	propTechsCarried = 2    // techs a full knowledge store carried into an age may pay for
	propFoodLow      = 0.30 // workers eat at least this share of the food grown
	propFoodHigh     = 0.80
	propRoomHigh     = 1.5  // houses hold at most this many people per job
	propResearchLow  = 0.70 // research takes this share of the ordinary player's age
	propResearchHigh = 0.90
	propRewardLow    = 1.0  // a fixed reward is worth at least this many minutes of income
	propRewardHigh   = 30.0 // and at most this many
	// warningShare is the shortest harbinger warning as a share of the
	// age's target (game/harbinger.go).
	warningShare = 0.20
)

// StaticProperties works every property out.
func StaticProperties() Properties {
	t := newRefTables()
	n := len(t.ages)
	towns := make([]map[string]int, n)
	ref := make([]townState, n)
	refIncome := make([]float64, n)
	for i := range t.ages {
		towns[i] = t.refTown(i)
		ref[i] = t.state(towns[i], i, 1)
		refIncome[i] = units(ref[i].income, t.levels[i])
	}
	ordinary := t.walk("ordinary", 1, refIncome, false)
	lingering := t.walk("lingering", 1, refIncome, false)
	checkIn := t.walk("check-in", 1, refIncome, false)
	out := Properties{Ordinary: ordinary.Hours, Lingering: lingering.Hours, CheckIn: checkIn.Hours}
	for _, a := range t.ages {
		out.Ages = append(out.Ages, a.Name)
		out.Target = append(out.Target, config.AgeTargetTicks(a.Key)/refTicksPerHour)
	}
	// The storage wall: the reference town left to buy every copy whose
	// price fits under the store, with nothing banked past it.
	wall := make([]refRun, n)
	wallTop := make([]refRun, n)
	for i := range t.ages {
		wall[i] = t.stay(i, towns[i], refPlay{rate: 1, speed: 1, refIncome: refIncome[i], horizon: 2e7})
		wallTop[i] = t.stay(i, towns[i], refPlay{rate: 1, speed: RefTopMastery, horizon: 2e7})
	}
	add := func(num int, says string, holds bool, today string) {
		out.List = append(out.List, Property{N: num, Says: says, Holds: holds, Today: today})
	}
	worst := func(count int, where string, value string) string {
		if count == 0 {
			return "holds in all 22 ages"
		}
		return fmt.Sprintf("fails in %d of %d ages; worst %s in the %s", count, n, value, where)
	}

	// 1. The price form.
	rates := map[float64]bool{1.15: true, 1.5: true, 1.75: true, 2.5: true}
	off, offRate := 0, 0.0
	for _, bs := range t.byAge {
		for _, d := range bs {
			if !rates[d.CostScale] {
				off++
				offRate = d.CostScale
			}
		}
	}
	add(1, "every building's price is its first price times a rate of 1.15, 1.5, 1.75 or 2.5 for each copy owned", off == 0,
		fmt.Sprintf("%d buildings climb at another rate (%g)", off, offRate))

	// 2. The first doubling takes at least the ordinary player's age.
	bad, where, val := 0, "", ""
	low := math.Inf(1)
	for i := range t.ages {
		if len(wall[i].doubles) == 0 {
			continue
		}
		share := wall[i].doubles[0] / float64(ordinary.Hours[i]*refTicksPerHour)
		if share < 1 {
			bad++
			if share < low {
				low, where, val = share, t.ages[i].Name, fmt.Sprintf("%.2f of the age", share)
			}
		}
	}
	add(2, "from the reference town, doubling income takes at least as long as the ordinary player's stay in that age", bad == 0, worst(bad, where, val))

	// 3 and 4. Income and store at the wall.
	bad, where, val = 0, "", ""
	high := 0.0
	bad4, where4, val4, high4 := 0, "", "", 0.0
	for i := range t.ages {
		x := units(wall[i].end.income, t.levels[i]) / refIncome[i]
		if x > propWallIncome {
			bad++
		}
		if x > high {
			high, where, val = x, t.ages[i].Name, fmt.Sprintf("%.1f times", x)
		}
		s := wall[i].end.caps["wood"] / ref[i].caps["wood"]
		if s > propWallStore {
			bad4++
		}
		if s > high4 {
			high4, where4, val4 = s, t.ages[i].Name, fmt.Sprintf("%.1f times", s)
		}
	}
	add(3, fmt.Sprintf("income at the storage wall is at most %g times the reference town's", propWallIncome), bad == 0, worst(bad, where, val))
	add(4, fmt.Sprintf("the store at the wall is at most %g times the reference store", propWallStore), bad4 == 0, worst(bad4, where4, val4))

	// 5 and 17. A full knowledge store carried in.
	carried := func(walls []refRun) (int, string, string) {
		bad, where, most := 0, "", 0
		for i := 1; i < n; i++ {
			k := walls[i-1].end.caps["knowledge"]
			if entry := config.AgeEntryCosts(t.ages[i].Key)["knowledge"]; entry > 0 {
				k = math.Min(k, 8*entry)
			} else {
				k = float64(k * 0.10)
			}
			paid := 0
			for _, price := range t.techs[i] {
				if price <= k {
					k -= price
					paid++
				}
			}
			if paid > propTechsCarried {
				bad++
			}
			if paid > most {
				most, where = paid, t.ages[i].Name
			}
		}
		return bad, where, fmt.Sprintf("%d techs", most)
	}
	bad, where, val = carried(wall)
	add(5, fmt.Sprintf("a full knowledge store carried into an age pays for at most %d of its techs", propTechsCarried), bad == 0, worst(bad, where, val))

	// 6. Every price the ordinary player meets fits a store he has reached.
	bad, where = 0, ""
	for i := range t.ages {
		var start map[string]int
		if i > 0 {
			start = towns[i-1]
		}
		only := map[string]int{}
		for _, d := range t.byAge[i] {
			only[d.Key] = t.refCount(d)
		}
		run := t.stay(i, start, refPlay{only: only, wonder: false, rate: 1, speed: 1, horizon: 2e7, leave: func(gate, now float64) bool { return true }})
		if run.gate < 0 {
			bad++
			if where == "" {
				where = t.ages[i].Name
			}
		}
	}
	today := "holds in all 22 ages"
	if bad > 0 {
		today = fmt.Sprintf("in %d of %d ages the reference town cannot be finished without banking past the store; first the %s", bad, n, where)
	}
	add(6, "every price the ordinary player meets fits under a store he has already reached", bad == 0, today)

	// 7. First copies fit the store a town arrives with.
	bad, where = 0, ""
	for i := 1; i < n; i++ {
		c := map[string]int{}
		for k, v := range towns[i-1] {
			c[k] = v
		}
		ok := true
		for _, d := range t.byAge[i] {
			if d.Category != "storage" {
				continue
			}
			for k := 0; k < RefCopies; k++ {
				if overCap(copyPrice(d, k), t.state(c, i, 1).caps) {
					ok = false
				}
				c[d.Key] = k + 1
			}
		}
		caps := t.state(c, i, 1).caps
		for _, d := range t.byAge[i] {
			if d.Category != "storage" && overCap(copyPrice(d, 0), caps) {
				ok = false
			}
		}
		if !ok {
			bad++
			if where == "" {
				where = t.ages[i].Name
			}
		}
	}
	today = "holds in all 22 ages"
	if bad > 0 {
		today = fmt.Sprintf("fails in %d of %d ages; first the %s", bad, n, where)
	}
	add(7, "the first copy of every building fits the store a reference town arrives with, once it has bought its five storage copies in order", bad == 0, today)

	// 8 and 9. Food and houses.
	bad, where, val = 0, "", ""
	low = math.Inf(1)
	bad9, where9, val9, high9 := 0, "", "", 0.0
	for i := range t.ages {
		share := ref[i].eaten / ref[i].grown
		if share < propFoodLow || share > propFoodHigh {
			bad++
		}
		if share < low {
			low, where, val = share, t.ages[i].Name, fmt.Sprintf("%.1f%% eaten", 100*share)
		}
		room := ref[i].room / ref[i].jobs
		if room < 1 || room > propRoomHigh {
			bad9++
		}
		if room > high9 {
			high9, where9, val9 = room, t.ages[i].Name, fmt.Sprintf("room for %.0f people per job", room)
		}
	}
	add(8, fmt.Sprintf("the reference town's workers eat between %.0f%% and %.0f%% of the food it grows", 100*propFoodLow, 100*propFoodHigh), bad == 0, worst(bad, where, val))
	add(9, fmt.Sprintf("the reference town's houses hold between 1 and %g times its jobs", propRoomHigh), bad9 == 0, worst(bad9, where9, val9))

	// 10. Research against the ordinary player's age.
	bad, where, val = 0, "", ""
	high = 0
	for i := range t.ages {
		k := ref[i].income["knowledge"]
		if len(t.techs[i]) == 0 || k <= 0 {
			continue
		}
		sum := 0.0
		for _, price := range t.techs[i] {
			sum += price
		}
		share := sum / k / float64(ordinary.Hours[i]*refTicksPerHour)
		if share < propResearchLow || share > propResearchHigh {
			bad++
		}
		if share > high {
			high, where, val = share, t.ages[i].Name, fmt.Sprintf("%.1f times the age", share)
		}
	}
	add(10, fmt.Sprintf("at the knowledge his own buildings make, research takes the ordinary player between %.0f%% and %.0f%% of each age", 100*propResearchLow, 100*propResearchHigh), bad == 0, worst(bad, where, val))

	// 11. The market against the buildings, for knowledge.
	bad, where, val = 0, "", ""
	high = 0
	for i, a := range t.ages {
		rate, ok := config.MarketRate("gold", "knowledge", a.Key)
		if !ok || ref[i].income["knowledge"] <= 0 {
			continue
		}
		x := float64(ref[i].income["gold"]*rate) / ref[i].income["knowledge"]
		if x > 1 {
			bad++
		}
		if x > high {
			high, where, val = x, a.Name, fmt.Sprintf("%.0f times what the buildings make", x)
		}
	}
	add(11, "selling the reference town's gold never brings in more knowledge than its own knowledge buildings make", bad == 0, worst(bad, where, val))

	// 12. Fixed rewards, for trade routes.
	routes, small := 0, 0
	for _, r := range config.BaseTradeRoutes() {
		i, ok := t.idx[r.MinAge]
		if !ok {
			continue
		}
		for res, amount := range r.Import {
			perMinute := float64(ref[i].income[res] * refTicksPerHour / 60)
			if perMinute <= 0 {
				continue
			}
			routes++
			if m := amount / perMinute; m < propRewardLow || m > propRewardHigh {
				small++
			}
		}
	}
	add(12, fmt.Sprintf("what a fixed reward pays is worth between %g and %g minutes of the reference town's income of it (trade routes; loot, flat bonuses and named events are not worked out here yet)", propRewardLow, propRewardHigh), small == 0,
		fmt.Sprintf("%d of %d trade route payments fall outside it", small, routes))

	// 13. (gone) It held every amount a gate asks for to a share of the store
	// that resource needs. No gate asks for an amount of a resource now (the
	// owner's decision of 2026-10-11), so it has nothing to measure. The
	// numbers of the others stay as they are, so the list has no 13.

	// 14. The era gift (the Ancient Cache): minutes of the town's income,
	// and in knowledge no more than the cheapest tech of the age entered.
	bad, where = 0, ""
	most := 0
	for i := 1; i < n; i++ {
		if t.ages[i].EpochKey == t.ages[i-1].EpochKey || len(t.techs[i]) == 0 {
			continue
		}
		k := math.Min(float64(ref[i-1].income["knowledge"]*config.AncientCacheMinutes*refTicksPerHour/60), t.techs[i][0])
		paid := 0
		for _, price := range t.techs[i] {
			if price <= k {
				k -= price
				paid++
			}
		}
		if paid > 1 {
			bad++
		}
		if paid > most {
			most, where = paid, t.ages[i].Name
		}
	}
	today = "holds at all six era lines"
	if bad > 0 {
		today = fmt.Sprintf("at %d of 6 era lines the Ancient Cache pays a reference town for more than one tech; most %d techs entering the %s", bad, most, where)
	}
	add(14, "the era gift pays for at most one tech of the age entered", bad == 0, today)

	// 15. Warnings against the visit gap.
	bad, where = 0, ""
	for i, a := range t.ages {
		if a.EpochKey == t.ages[0].EpochKey {
			continue // nothing is fated in the first era
		}
		if float64(out.Target[i]*warningShare) < RefVisitHours {
			bad++
			where = a.Name
		}
	}
	today = "holds wherever a doom can be fated"
	if bad > 0 {
		today = fmt.Sprintf("in %d ages the shortest warning is under %g hours; the last of them is the %s", bad, RefVisitHours, where)
	}
	add(15, "the shortest warning a harbinger can give is at least as long as the gap between a check-in player's visits", bad == 0, today)

	// 16. The check-in player.
	bad, where, val = 0, "", ""
	high = 0
	for i, a := range t.ages {
		if checkIn.Income[i] > propWallIncome {
			bad++
		}
		if checkIn.Income[i] > high {
			high, where, val = checkIn.Income[i], a.Name, fmt.Sprintf("%.1f times", checkIn.Income[i])
		}
	}
	add(16, fmt.Sprintf("the check-in player leaves each age making at most %g times what it is priced for", propWallIncome), bad == 0, worst(bad, where, val))

	// 17. The same at the top of the mastery table, where stores are larger
	// and prices are not.
	bad, where, val = carried(wallTop)
	add(17, fmt.Sprintf("at the top of the Era Mastery table a full knowledge store still pays for at most %d techs of the age entered", propTechsCarried), bad == 0, worst(bad, where, val))

	// 18 needs the game's own formulas and is stated, not worked out.
	add(18, "the reference town's faith and culture strength fall in the middle band", false, "not worked out here yet")

	// 19. Prestige points and what they buy.
	sink := 0
	for _, u := range config.ActivePrestigeUpgrades() {
		for _, c := range u.Costs {
			sink += c
		}
	}
	earned := config.DepthPoints(t.ages[n-1].Key)
	add(19, "the points a full run earns can all be spent", sink >= earned, fmt.Sprintf("a full run earns %d points and everything for sale costs %d", earned, sink))

	// 20. The largest amount any of the three players can hold in a store or
	// be quoted (the next copy of any building of the age, the age's wonder
	// and techs), at Era Mastery 1 and at the top of the
	// table, against the largest number the game can print.
	topWalks := map[string]refWalk{
		"ordinary":  t.walk("ordinary", RefTopMastery, refIncome, false),
		"lingering": t.walk("lingering", RefTopMastery, refIncome, false),
		"check-in":  t.walk("check-in", RefTopMastery, refIncome, false),
	}
	baseWalks := map[string]refWalk{"ordinary": ordinary, "lingering": lingering, "check-in": checkIn}
	wonders := map[string]map[string]float64{}
	for _, d := range config.BaseBuildings() {
		if d.Category == "wonder" {
			wonders[d.RequiredAge] = d.BaseCost
		}
	}
	biggest, biggestAt := 0.0, ""
	see := func(v float64, what, player, mastery, age string) {
		if v > biggest {
			biggest = v
			biggestAt = fmt.Sprintf("%s (%s, %s, %s, %s)", textfmt.Number(v), what, player, mastery, age)
		}
	}
	for _, player := range []string{"ordinary", "lingering", "check-in"} {
		for m, w := range []refWalk{baseWalks[player], topWalks[player]} {
			mastery := []string{"Era Mastery 1", "top of the Era Mastery table"}[m]
			for i, run := range w.Runs {
				age := t.ages[i].Name
				for r, v := range run.end.caps {
					see(v, r+" store", player, mastery, age)
				}
				for _, d := range t.byAge[i] {
					for r, v := range copyPrice(d, run.counts[d.Key]) {
						see(v, "next "+d.Name+", "+r, player, mastery, age)
					}
				}
				for r, v := range wonders[t.ages[i].Key] {
					see(v, "wonder, "+r, player, mastery, age)
				}
				if k := len(t.techs[i]); k > 0 {
					see(t.techs[i][k-1], "dearest tech", player, mastery, age)
				}
			}
		}
	}
	add(20, "nothing any of the three reference players can hold or be quoted is larger than the largest printable number", biggest <= textfmt.MaxNumber(),
		fmt.Sprintf("the largest is %s, against %s printable", biggestAt, textfmt.Number(textfmt.MaxNumber())))
	add(21, "the ordinary player's length of each age is printed beside its design target", true, "printed below")
	return out
}

func hoursText(h float64) string {
	switch {
	case h < 1:
		return fmt.Sprintf("%.0f min", h*60)
	case h < 48:
		return fmt.Sprintf("%.1f h", h)
	}
	return fmt.Sprintf("%.1f days", h/24)
}

// writeProperties renders the check for the report.
func writeProperties(sb *strings.Builder, p Properties) {
	sb.WriteString("What must be true of the economy's numbers, worked out from the game's tables alone for three reference players: an ordinary one who builds five of everything (or what the gate asks), pays the wonder and advances; a lingering one who buys everything and stays three times as long; and a check-in one who is away at half rate and advances at the first eight-hour visit after the gate opens. No bot is involved.\n\n")
	sb.WriteString("A property marked \"known failure\" does not hold today and is written down as such, with today's reading. The check fails when one of those starts to hold, when any other stops holding, or when a reading changes. There is no property 13: it was about the amounts a gate asks for, and no gate asks for one now.\n\n")
	sb.WriteString("| # | property | today | |\n|---|---|---|---|\n")
	for _, pr := range p.List {
		mark := "holds"
		if !pr.Holds {
			mark = "known failure"
			if _, ok := KnownPropertyFailures[pr.N]; !ok {
				mark = "FAILS"
			}
		} else if _, ok := KnownPropertyFailures[pr.N]; ok {
			mark = "HOLDS NOW: take it off the list"
		}
		fmt.Fprintf(sb, "| %d | %s | %s | %s |\n", pr.N, pr.Says, pr.Today, mark)
	}
	sb.WriteString("\nHow long each age lasts, by calculation: the first moment the town has made what the player bought, wonder included. Research is not timed here (property 10 sets it against the age).\n\n")
	sb.WriteString("| age | design target | ordinary | lingering | check-in |\n|---|---|---|---|---|\n")
	var tt, to, tl, tc float64
	for i, a := range p.Ages {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s |\n", a, hoursText(p.Target[i]), hoursText(p.Ordinary[i]), hoursText(p.Lingering[i]), hoursText(p.CheckIn[i]))
		tt += p.Target[i]
		to += p.Ordinary[i]
		tl += p.Lingering[i]
		tc += p.CheckIn[i]
	}
	fmt.Fprintf(sb, "| all 22 | %s | %s | %s | %s |\n", hoursText(tt), hoursText(to), hoursText(tl), hoursText(tc))
}
