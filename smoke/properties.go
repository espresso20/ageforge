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
	2:  "fails in 2 of 22 ages; worst 0.29 of the age in the Primitive Age",
	8:  "fails in 5 of 22 ages; worst 21.4% eaten in the Atomic Age",
	9:  "fails in 22 of 22 ages; worst room for 22534 people per job in the Transcendent Age",
	12: "27 of 27 trade route payments fall outside it",
	15: "in 11 ages the shortest warning is under 8 hours; the last of them is the Information Age",
	16: "fails in 5 of 22 ages; worst 4.6 times in the Primitive Age",
	18: "not worked out here yet",
}

// KnownPropertyReasons is why a known failure fails, where the reason is a
// decision and not a table to fix.
var KnownPropertyReasons = map[int]string{}

// researchBand is the band of property 10 for age: the share of the ordinary
// player's stay that the age's techs take with every scholar at the cap
// researching. The Primitive Age (config.ResearchSimmerExemptAge) keeps 0.7 to
// 0.9; every other age's band is that scaled by config.ResearchSimmer, 1.05 to
// 1.35 (constant products, exact).
func researchBand(age string) (low, high float64) {
	if age == config.ResearchSimmerExemptAge {
		return propResearchLow, propResearchHigh
	}
	return propResearchLow * config.ResearchSimmer, propResearchHigh * config.ResearchSimmer
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
	propIncomeHigh   = 2.5  // income a player leaves an age with, over the reference town's
	propFoodLow      = 0.30 // workers eat at least this share of the food grown
	propFoodHigh     = 0.80
	propRoomHigh     = 1.5  // houses hold at most this many people per job
	propResearchLow  = 0.70 // research takes this share of the ordinary player's age (the exempt age's band)
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
	ordinary := t.walk("ordinary", 1, refIncome)
	lingering := t.walk("lingering", 1, refIncome)
	checkIn := t.walk("check-in", 1, refIncome)
	out := Properties{Ordinary: ordinary.Hours, Lingering: lingering.Hours, CheckIn: checkIn.Hours}
	for _, a := range t.ages {
		out.Ages = append(out.Ages, a.Name)
		out.Target = append(out.Target, config.AgeTargetTicks(a.Key)/refTicksPerHour)
	}
	// The reference town left to buy the cheapest thing it can, with no limit,
	// until its income has doubled.
	grown := make([]refRun, n)
	for i := range t.ages {
		grown[i] = t.stay(i, towns[i], refPlay{rate: 1, speed: 1, refIncome: refIncome[i], horizon: 2e7, stopAtDouble: true})
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
	high := 0.0
	today := ""
	for i := range t.ages {
		if len(grown[i].doubles) == 0 {
			continue
		}
		share := grown[i].doubles[0] / float64(ordinary.Hours[i]*refTicksPerHour)
		if share < 1 {
			bad++
			if share < low {
				low, where, val = share, t.ages[i].Name, fmt.Sprintf("%.2f of the age", share)
			}
		}
	}
	add(2, "from the reference town, doubling income takes at least as long as the ordinary player's stay in that age", bad == 0, worst(bad, where, val))

	// 5. (gone) It held what a full knowledge store carried into an age
	// could pay for. Knowledge is not stored now: a tech is work the scholars
	// do and there is no stock to carry. The numbers of the others stay as
	// they are, so the list has no 5.

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

	// 10. Research against the ordinary player's age: with every scholar of
	// the copies the gates ask for (the cap) researching at a full crew.
	bad, where, val = 0, "", ""
	high = 0
	for i := range t.ages {
		k := ref[i].income[config.ResearchTarget]
		if len(t.techs[i]) == 0 || k <= 0 {
			continue
		}
		sum := 0.0
		for _, price := range t.techs[i] {
			sum += price
		}
		share := sum / k / float64(ordinary.Hours[i]*refTicksPerHour)
		lo, hi := researchBand(t.ages[i].Key)
		if share < lo || share > hi {
			bad++
		}
		if share > high {
			high, where, val = share, t.ages[i].Name, fmt.Sprintf("%.1f times the age", share)
		}
	}
	add(10, fmt.Sprintf("with every scholar at the cap researching, the Primitive Age's techs take between %.0f%% and %.0f%% of the ordinary player's stay, and every later age's between %.0f%% and %.0f%% (research simmers: %g times as long)", 100*propResearchLow, 100*propResearchHigh, 100*propResearchLow*config.ResearchSimmer, 100*propResearchHigh*config.ResearchSimmer, config.ResearchSimmer), bad == 0, worst(bad, where, val))

	// 11. (gone) It held the gold sold for knowledge against what the
	// buildings make. The market no longer swaps, so there is nothing to
	// sell. The numbers of the others stay as they are, so the list has no 11.

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
		k := math.Min(float64(ref[i-1].income[config.ResearchTarget]*config.AncientCacheMinutes*refTicksPerHour/60), t.techs[i][0])
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
		if checkIn.Income[i] > propIncomeHigh {
			bad++
		}
		if checkIn.Income[i] > high {
			high, where, val = checkIn.Income[i], a.Name, fmt.Sprintf("%.1f times", checkIn.Income[i])
		}
	}
	add(16, fmt.Sprintf("the check-in player leaves each age making at most %g times what it is priced for", propIncomeHigh), bad == 0, worst(bad, where, val))

	// 17. (gone) The same as 5, at the top of the mastery table. It goes
	// with 5, and the list has no 17.

	// 22. The cap is the cap: no town makes more research than the ceiling
	// property 10 reads. The most copies of a knowledge building the game lets
	// a town build is the count its gate asks for (MaxCount), so the reference
	// town, which holds exactly those, is the best a town can do by building.
	bad, where, val = 0, "", ""
	for i := range t.ages {
		most := map[string]int{}
		for k, c := range towns[i] {
			most[k] = c
		}
		for j := 0; j <= i; j++ {
			for _, d := range t.byAge[j] {
				if config.IsKnowledgeBuilding(d) {
					most[d.Key] = d.MaxCount
				}
			}
		}
		if got, ceiling := t.state(most, i, 1).income[config.ResearchTarget], ref[i].income[config.ResearchTarget]; got > ceiling*(1+1e-9) {
			bad++
			where, val = t.ages[i].Name, fmt.Sprintf("%.2f times the ceiling", got/ceiling)
		}
	}
	add(22, "building every knowledge building the game allows makes no more research than the ceiling of property 10", bad == 0, worst(bad, where, val))

	// 18 needs the game's own formulas and is stated, not worked out.
	add(18, "the reference town's faith and culture strength fall in the middle band", false, "not worked out here yet")

	// 19. Prestige points no longer have a shop sink.
	add(19, "prestige points have no shop sink", len(config.ActivePrestigeUpgrades()) == 0, "the prestige shop has no items")

	// 20. The largest amount any of the three players can be quoted (the next
	// copy of any building of the age, the age's wonder and techs), at Era
	// Mastery 1 and at the top of the table, against the largest number the
	// game can print. (Stock has no cap, so what a town holds has no bound
	// to read here: it follows what it is quoted.)
	topWalks := map[string]refWalk{
		"ordinary":  t.walk("ordinary", RefTopMastery, refIncome),
		"lingering": t.walk("lingering", RefTopMastery, refIncome),
		"check-in":  t.walk("check-in", RefTopMastery, refIncome),
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
	add(20, "nothing any of the three reference players can be quoted is larger than the largest printable number", biggest <= textfmt.MaxNumber(),
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
	sb.WriteString("What must be true of the economy's numbers, worked out from the game's tables alone for three reference players: an ordinary one who builds eight of everything but storage and five of each storage building (or what the gate asks), pays the wonder and advances; a lingering one who buys everything and stays three times as long; and a check-in one who is away at half rate and advances at the first eight-hour visit after the gate opens. No bot is involved.\n\n")
	sb.WriteString("A property marked \"known failure\" does not hold today and is written down as such, with today's reading. The check fails when one of those starts to hold, when any other stops holding, or when a reading changes. There is no property 13: it was about the amounts a gate asks for, and no gate asks for one now. There are no properties 3, 4, 6 and 7 either: they were about the storage wall (income and store at the wall, prices and first copies against the stores), and stock has no cap.\n\n")
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
