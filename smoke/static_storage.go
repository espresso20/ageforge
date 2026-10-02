package smoke

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// The Storage Covenant (design-and-architecture/economy.md, Law 1): the most
// storage buildable in an age must hold at least config.StorageHold(age)
// hours (4.5 from the Bronze Age on, 1.5 in the Primitive and Stone Ages) of
// the age's typical production (config.TypicalIncome) of every construction
// resource of the age. It sizes storage to time, where the Gate Covenant
// sizes it to prices: a store that fills in minutes wastes most of what a
// player makes between visits.

// StorageRow is one age's tightest construction resource: the most storage
// buildable by the end of the age against its typical income.
type StorageRow struct {
	Age        string  `json:"age"`
	Resource   string  `json:"resource"`
	MaxStorage float64 `json:"max_storage"`
	Income     float64 `json:"income_per_tick"`
	// Hours is how long MaxStorage holds Income at 1x.
	Hours float64 `json:"hours"`
}

// Want is the hours the covenant asks of the row's age.
func (r StorageRow) Want() float64 { return config.StorageHold(r.Age) }

// OK reports whether the row keeps the covenant.
func (r StorageRow) OK() bool { return r.Hours >= r.Want() }

// StaticStorage checks every age against the Storage Covenant and returns
// one row per age (its tightest resource), failing or not.
func StaticStorage() []StorageRow {
	return staticStorage(config.BuildingByKey(), config.TypicalIncome)
}

func staticStorage(defs map[string]config.BuildingDef, income func(res, age string) float64) []StorageRow {
	var out []StorageRow
	for _, age := range config.AgeOrder() {
		row := StorageRow{Age: age, Hours: math.Inf(1)}
		for _, res := range sortedKeys(config.PriceLevels(age)) {
			inc := income(res, age)
			if inc <= 0 {
				continue
			}
			m := maxStorageIn(defs, age, res)
			if h := m / (inc * 3600 / config.TickSeconds); h < row.Hours {
				row.Resource, row.MaxStorage, row.Income, row.Hours = res, m, inc, h
			}
		}
		if row.Resource != "" {
			out = append(out, row)
		}
	}
	return out
}

// writeStorage renders the Storage Covenant check.
func writeStorage(sb *strings.Builder, rows []StorageRow) {
	fmt.Fprintf(sb, "The most storage buildable in each age must hold %g hours of the age's typical production (%g hours in the Primitive and Stone Ages; config.TypicalIncome: five staffed copies of every producer so far, earlier wonders and techs, with the production bonus) of each of its construction resources (the Storage Covenant, economy.md). One row per age, its tightest resource. `go test ./smoke` fails on any row marked ✗.\n\n", config.StorageHoldHours, config.EarlyStorageHoldHours)
	sb.WriteString("| age | resource | typical income | max storage | holds | needs | |\n|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		mark := "✓"
		if !r.OK() {
			mark = "✗"
		}
		fmt.Fprintf(sb, "| %s | %s | %s/tick | %s | %.2f h | %g h | %s |\n", r.Age, r.Resource, num(r.Income), num(r.MaxStorage), r.Hours, r.Want(), mark)
	}
}
