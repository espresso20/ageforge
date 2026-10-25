// Command balanceaudit prints a cost table for every building in the base config, then
// a summary of buildings per CostScale and per age band. It uses only the standard
// library and the config package.
//
// The cost of the N-th copy (1-indexed) follows game/buildings.go GetCost:
//
//	floor(BaseCost[res] * CostScale^(N-1))
//
// It leaves out the global build-cost multiplier that GetCost applies.
//
// Usage:
//
//	go run ./cmd/balanceaudit
package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/espresso20/ageforge/config"
)

// costAt returns the cost of the N-th copy (1-indexed) of a base amount. math.Pow with
// an integer exponent gives the same bits as the game's detmath.Pow.
func costAt(base, scale float64, n int) float64 {
	return math.Floor(base * math.Pow(scale, float64(n-1)))
}

// primaryCost returns the resource with the largest base amount, and that amount.
// Ties go to the alphabetically first resource key, so the output is stable. An empty
// cost returns "-" and 0.
func primaryCost(cost map[string]float64) (string, float64) {
	res := ""
	var amt float64 = -1
	for r, v := range cost {
		if v > amt || (v == amt && (res == "" || r < res)) {
			res, amt = r, v
		}
	}
	if res == "" {
		return "-", 0
	}
	return res, amt
}

// prodRate returns the value of the building's first "production" effect, or false if it
// has none.
func prodRate(b config.BuildingDef) (float64, bool) {
	for _, e := range b.Effects {
		if e.Type == "production" {
			return e.Value, true
		}
	}
	return 0, false
}

// fmtNum renders large numbers compactly with K/M/B/T suffixes.
func fmtNum(v float64) string {
	abs := math.Abs(v)
	switch {
	case abs >= config.Trillion:
		return trim(v/config.Trillion) + "T"
	case abs >= config.Billion:
		return trim(v/config.Billion) + "B"
	case abs >= config.Million:
		return trim(v/config.Million) + "M"
	case abs >= 1000:
		return trim(v/1000) + "K"
	default:
		if v == math.Trunc(v) {
			return fmt.Sprintf("%d", int64(v))
		}
		return trim(v)
	}
}

// trim formats a float with up to 2 decimals and strips trailing zeros.
func trim(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// ageBand puts an age key in the "early", "mid" or "late" band.
func ageBand(age string) string {
	switch age {
	case "primitive_age", "stone_age", "bronze_age", "iron_age":
		return "early"
	}
	// By position in the age order: first third "early", second third "mid", the rest
	// "late". An age missing from the order is "late".
	order := ageIndex()
	idx, ok := order[age]
	if !ok {
		return "late"
	}
	n := len(order)
	if idx < n/3 {
		return "early"
	}
	if idx < 2*n/3 {
		return "mid"
	}
	return "late"
}

// ageIndexCache is the age key -> position in config.AgeOrder() lookup that ageIndex
// builds on first use.
var ageIndexCache map[string]int

func ageIndex() map[string]int {
	if ageIndexCache != nil {
		return ageIndexCache
	}
	ageIndexCache = make(map[string]int)
	for i, k := range config.AgeOrder() {
		ageIndexCache[k] = i
	}
	return ageIndexCache
}

func main() {
	buildings := config.BaseBuildings()
	idx := ageIndex()

	// Sort by age order, then tier, then key.
	sort.SliceStable(buildings, func(i, j int) bool {
		ai, aj := idx[buildings[i].RequiredAge], idx[buildings[j].RequiredAge]
		if ai != aj {
			return ai < aj
		}
		if buildings[i].LineageTier != buildings[j].LineageTier {
			return buildings[i].LineageTier < buildings[j].LineageTier
		}
		return buildings[i].Key < buildings[j].Key
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "age\ttier\tkey\tcostScale\tprimaryRes\tbase\tcost@1\tcost@10\tcost@20\tcost@30\tmaxCount\tprodRate")

	for _, b := range buildings {
		res, base := primaryCost(b.BaseCost)
		maxStr := "-"
		if b.MaxCount > 0 {
			maxStr = fmt.Sprintf("%d", b.MaxCount)
		} else if b.FullRateCopies > 0 {
			maxStr = fmt.Sprintf("%d soft", b.FullRateCopies) // half rate past it
		}
		rateStr := ""
		if r, ok := prodRate(b); ok {
			rateStr = trim(r)
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			b.RequiredAge,
			b.LineageTier,
			b.Key,
			trim(b.CostScale),
			res,
			fmtNum(base),
			fmtNum(costAt(base, b.CostScale, 1)),
			fmtNum(costAt(base, b.CostScale, 10)),
			fmtNum(costAt(base, b.CostScale, 20)),
			fmtNum(costAt(base, b.CostScale, 30)),
			maxStr,
			rateStr,
		)
	}
	w.Flush()

	printSummary(buildings)
}

func printSummary(buildings []config.BuildingDef) {
	fmt.Println()
	fmt.Println("==== SUMMARY ====")

	byScale := map[string]int{}
	// Age band -> formatted CostScale -> building count.
	byBandScale := map[string]map[string]int{}
	bands := []string{"early", "mid", "late"}
	for _, band := range bands {
		byBandScale[band] = map[string]int{}
	}

	scaleSet := map[string]bool{}
	for _, b := range buildings {
		s := trim(b.CostScale)
		byScale[s]++
		band := ageBand(b.RequiredAge)
		byBandScale[band][s]++
		scaleSet[s] = true
	}

	scales := make([]string, 0, len(scaleSet))
	for s := range scaleSet {
		scales = append(scales, s)
	}
	sort.Slice(scales, func(i, j int) bool {
		return parseFloat(scales[i]) < parseFloat(scales[j])
	})

	fmt.Println()
	fmt.Println("-- buildings per CostScale (all ages) --")
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "costScale\tcount")
	for _, s := range scales {
		fmt.Fprintf(w, "%s\t%d\n", s, byScale[s])
	}
	w.Flush()

	fmt.Println()
	fmt.Println("-- buildings per (ageBand x CostScale) --")
	w = tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	header := "costScale"
	for _, band := range bands {
		header += "\t" + band
	}
	fmt.Fprintln(w, header)
	for _, s := range scales {
		row := s
		for _, band := range bands {
			row += fmt.Sprintf("\t%d", byBandScale[band][s])
		}
		fmt.Fprintln(w, row)
	}
	w.Flush()
}

// parseFloat reads a number formatted by trim. It returns 0 if s does not parse.
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%g", &f)
	return f
}
