package smoke

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
)

// The Storage check. Stock has no cap, so a storage building holds nothing and
// no store is sized: each copy adds config.StorageBoostPercent percent to all
// production instead. What is left to hold, from the tables alone (it does not
// call config's own rules, so a rule that drifted would not agree with itself):
//
//   - the storage building climbs at storageClimb a copy and has no copy
//     limit: the price is the limit;
//   - it stores nothing and does one thing, boosts all production by
//     storageBoost a copy;
//   - the gate out of the age asks for storageCopies copies of its storage
//     building, so every town that passes arrives with five of each;
//   - the copies the gate asks for are payable from the age's income: every
//     price part of each of them is at most what a moderate town of the age
//     (config.Incomes) makes of that resource in one pacing target of the
//     age. This is the intent of the old rule that the first copy fits the
//     store a town arrives with: an age must not open with a storage
//     building nobody can pay for.
const (
	storageCopies = 5
	storageClimb  = 1.75
	storageBoost  = 0.01
)

// StorageProblem is a place the tables part from the storage check.
type StorageProblem struct {
	Age  string `json:"age"`
	Rule string `json:"rule"` // curve, boost, gate_copies, afford
	Why  string `json:"why"`
}

func (p StorageProblem) String() string { return fmt.Sprintf("%s: %s", p.Age, p.Why) }

// StorageRow is one age's storage building as the tables give it.
type StorageRow struct {
	Age      string `json:"age"`
	Building string `json:"building"`
	// First and Last are the dearest price part of the first copy and of the
	// last copy the gate asks for; Share is the larger over what the age's
	// moderate town makes of that resource in one pacing target (1 or under
	// is payable inside it).
	First float64 `json:"first_copy"`
	Last  float64 `json:"last_gate_copy"`
	Share float64 `json:"share_of_target_income"`
}

// StaticStorage checks every age's storage building against the storage
// check and returns what breaks it, with one row per age that has one.
func StaticStorage() ([]StorageProblem, []StorageRow) {
	defs := config.BuildingByKey()
	income := config.Incomes(config.BaseBuildings(), config.Technologies(), config.AgeOrder(), config.AnyResource)
	return staticStorage(defs, config.Ages(), income)
}

// storageCost is the price of copy k (1 for the first) of d in res.
func storageCost(d config.BuildingDef, res string, k int) float64 {
	return math.Max(1, math.Floor(float64(d.BaseCost[res]*detmath.Pow(d.CostScale, float64(k-1)))))
}

// staticStorage is StaticStorage over the given tables and incomes (age ->
// resource -> per tick): the check's broken-number tests feed in others.
func staticStorage(defs map[string]config.BuildingDef, ages []config.AgeDef, income map[string]map[string]float64) ([]StorageProblem, []StorageRow) {
	var problems []StorageProblem
	var rows []StorageRow
	for i, age := range ages {
		var b config.BuildingDef
		found := false
		for _, k := range sortedKeys(defs) {
			if d := defs[k]; d.Category == "storage" && d.RequiredAge == age.Key {
				b, found = d, true
			}
		}
		if !found {
			continue
		}
		fail := func(rule, format string, args ...any) {
			problems = append(problems, StorageProblem{Age: age.Key, Rule: rule, Why: fmt.Sprintf(format, args...)})
		}
		if b.CostScale != storageClimb || b.MaxCount != 0 {
			fail("curve", "%s climbs at %g a copy with a copy limit of %d; storage climbs at %g and has no copy limit (the price is the limit)", b.Key, b.CostScale, b.MaxCount, storageClimb)
		}
		boost := 0.0
		for _, e := range b.Effects {
			switch {
			case e.Type == "bonus" && e.Target == "production_all":
				boost += e.Value
			default:
				fail("boost", "%s has an effect %s on %s: a storage building stores nothing and only boosts all production", b.Key, e.Type, e.Target)
			}
		}
		if math.Abs(boost-storageBoost) > 1e-12 {
			fail("boost", "%s boosts all production by %g a copy, want %g", b.Key, boost, storageBoost)
		}
		if i+1 < len(ages) {
			if got := ages[i+1].BuildingReqs[b.Key]; got != storageCopies {
				fail("gate_copies", "the gate into %s asks for %d %s, want %d: whoever passes it must arrive with five copies", ages[i+1].Key, got, b.Key, storageCopies)
			}
		}
		row := StorageRow{Age: age.Key, Building: b.Key}
		target := config.LatePriceTargetTicks(age.Key)
		for _, res := range sortedKeys(b.BaseCost) {
			made := float64(income[age.Key][res] * target)
			first, last := storageCost(b, res, 1), storageCost(b, res, storageCopies)
			row.First, row.Last = math.Max(row.First, first), math.Max(row.Last, last)
			if made <= 0 {
				fail("afford", "%s costs %s, which a moderate town of the age makes none of", b.Key, res)
				continue
			}
			row.Share = math.Max(row.Share, last/made)
			if last > made {
				fail("afford", "copy %d of %s costs %s %s, more than the %s a moderate town makes of it in one pacing target of the age", storageCopies, b.Key, num(last), res, num(made))
			}
		}
		rows = append(rows, row)
	}
	return problems, rows
}

// writeStorage renders the storage check.
func writeStorage(sb *strings.Builder, problems []StorageProblem, rows []StorageRow) {
	sb.WriteString("Stock has no cap, so a storage building holds nothing: each copy adds 1% to all production. Checked here from the tables alone: it climbs at 1.75 a copy with no copy limit; it does nothing else; each gate asks for five copies of the storage building of the age it leaves; and the five copies are payable from the age's income, each price part inside what a moderate town of the age makes of that resource in one pacing target. `go test ./smoke` fails on any row below.\n\n")
	if len(problems) == 0 {
		sb.WriteString("No problems.\n\n")
	} else {
		sb.WriteString("| age | rule | |\n|---|---|---|\n")
		for _, p := range problems {
			fmt.Fprintf(sb, "| %s | %s | %s |\n", p.Age, p.Rule, p.Why)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("| age | storage building | dearest part, first copy | dearest part, fifth copy | fifth copy over one target of income |\n|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %.2fx |\n", r.Age, r.Building, num(r.First), num(r.Last), r.Share)
	}
}
