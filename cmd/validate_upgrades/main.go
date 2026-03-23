package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// upgradeCost computes total cost to upgrade upgradeCount copies of oldDef to newDef,
// assuming oldCount copies of old exist and newCount copies of new already exist.
// Replicates the logic in game/buildings.go UpgradeCost exactly.
func upgradeCost(oldDef, newDef config.BuildingDef, oldCount, newCount, upgradeCount int) map[string]float64 {
	total := make(map[string]float64)
	for i := 0; i < upgradeCount; i++ {
		oldExp := float64(oldCount - 1 - i)
		if oldExp < 0 {
			oldExp = 0
		}
		newExp := float64(newCount + i)
		for res, base := range newDef.BaseCost {
			newCopyCost := math.Floor(base * math.Pow(newDef.CostScale, newExp))
			oldBase := oldDef.BaseCost[res]
			oldCopyCost := math.Floor(oldBase * math.Pow(oldDef.CostScale, oldExp))
			oldSellValue := math.Floor(oldCopyCost * 0.5)
			delta := newCopyCost - oldSellValue
			if delta < 0 {
				delta = 0
			}
			total[res] += delta
		}
	}
	return total
}

// freshBuildCost computes the total cost to build n copies of def from scratch,
// starting at newCount already built (no trade-in).
func freshBuildCost(def config.BuildingDef, startCount, n int) map[string]float64 {
	total := make(map[string]float64)
	for i := 0; i < n; i++ {
		exp := float64(startCount + i)
		for res, base := range def.BaseCost {
			total[res] += math.Floor(base * math.Pow(def.CostScale, exp))
		}
	}
	return total
}

// sumCost sums all resource values in a cost map (for ratio calculations).
func sumCost(m map[string]float64) float64 {
	s := 0.0
	for _, v := range m {
		s += v
	}
	return s
}

// productionRate returns the total production/tick from a building's effects (base rate × count).
func productionRate(def config.BuildingDef, count int) float64 {
	rate := 0.0
	for _, eff := range def.Effects {
		if eff.Type == "production" {
			rate += eff.Value * float64(count)
		}
	}
	return rate
}

// formatCost formats a cost map as "res:val res:val" sorted for readability.
func formatCost(m map[string]float64) string {
	var parts []string
	for res, v := range m {
		parts = append(parts, fmt.Sprintf("%s:%.0f", res, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, " ")
}

// isFreeUpgrade returns true if all resources in upgradeCost are 0.
func isFreeUpgrade(cost map[string]float64) bool {
	for _, v := range cost {
		if v > 0 {
			return false
		}
	}
	return true
}

// isOverpriced returns true if any resource in upgradeCost exceeds the corresponding fresh cost.
func isOverpriced(upgCost, freshCost map[string]float64) bool {
	for res, uv := range upgCost {
		if fv, ok := freshCost[res]; ok {
			if uv > fv {
				return true
			}
		} else {
			// New building needs a resource the old one doesn't provide sell credit for.
			// Compare against fresh cost sum as a proxy: if upgrade for this resource alone
			// exceeds total fresh, flag it.
			if uv > sumCost(freshCost) && sumCost(freshCost) > 0 {
				return true
			}
		}
	}
	return false
}

// hasNegativeDelta returns true if any resource would yield negative delta (should be caught by max(0)).
// This verifies the guard in UpgradeCost is never actually needed (delta already ≥ 0 before max).
func hasNegativeDeltaBeforeGuard(oldDef, newDef config.BuildingDef, oldCount, newCount, upgradeCount int) bool {
	for i := 0; i < upgradeCount; i++ {
		oldExp := float64(oldCount - 1 - i)
		if oldExp < 0 {
			oldExp = 0
		}
		newExp := float64(newCount + i)
		for res, base := range newDef.BaseCost {
			newCopyCost := math.Floor(base * math.Pow(newDef.CostScale, newExp))
			oldBase := oldDef.BaseCost[res]
			oldCopyCost := math.Floor(oldBase * math.Pow(oldDef.CostScale, oldExp))
			oldSellValue := math.Floor(oldCopyCost * 0.5)
			delta := newCopyCost - oldSellValue
			if delta < 0 {
				return true
			}
		}
	}
	return false
}

type lineageSummary struct {
	lineage           string
	tierCount         int
	minBoostRatio     float64
	maxBoostRatio     float64
	anyFreeUpgrade    bool
	anyOverpriced     bool
	anyNegativeDelta  bool
	overpricedDetails []string
	freeDetails       []string
	highBoostDetails  []string
}

func main() {
	buildings := config.BaseBuildings()

	// Group production buildings by lineage
	lineageMap := make(map[string][]config.BuildingDef)
	for _, b := range buildings {
		if b.LineageKey == "" || b.Category != "production" {
			continue
		}
		lineageMap[b.LineageKey] = append(lineageMap[b.LineageKey], b)
	}

	// Sort each lineage by tier
	for k := range lineageMap {
		sort.Slice(lineageMap[k], func(i, j int) bool {
			return lineageMap[k][i].LineageTier < lineageMap[k][j].LineageTier
		})
	}

	// Sorted lineage keys for deterministic output
	var lineageKeys []string
	for k := range lineageMap {
		lineageKeys = append(lineageKeys, k)
	}
	sort.Strings(lineageKeys)

	oldCounts := []int{1, 3, 5, 10}
	summaries := make(map[string]*lineageSummary)

	for _, lineage := range lineageKeys {
		tiers := lineageMap[lineage]
		s := &lineageSummary{
			lineage:       lineage,
			tierCount:     len(tiers),
			minBoostRatio: math.MaxFloat64,
			maxBoostRatio: -math.MaxFloat64,
		}
		summaries[lineage] = s

		fmt.Printf("\n%s\n%s\n", strings.Repeat("=", 72), strings.Repeat("=", 72))
		fmt.Printf("LINEAGE: %s  (%d tiers)\n", strings.ToUpper(lineage), len(tiers))
		fmt.Printf("%s\n\n", strings.Repeat("=", 72))

		for pairIdx := 0; pairIdx < len(tiers)-1; pairIdx++ {
			oldDef := tiers[pairIdx]
			newDef := tiers[pairIdx+1]

			fmt.Printf("  %s (tier %d) → %s (tier %d)\n",
				oldDef.Key, oldDef.LineageTier, newDef.Key, newDef.LineageTier)
			fmt.Printf("  Old base cost: %s  scale=%.2f\n", formatCost(oldDef.BaseCost), oldDef.CostScale)
			fmt.Printf("  New base cost: %s  scale=%.2f\n", formatCost(newDef.BaseCost), newDef.CostScale)
			fmt.Printf("  %-8s  %-14s  %-14s  %-8s  %-8s  %-8s  %-8s  %s\n",
				"old_cnt", "upgrade_cost", "fresh_cost", "savings%", "prod_bfr", "prod_aft", "boost", "flags")
			fmt.Printf("  %s\n", strings.Repeat("-", 100))

			pairHasIssue := false

			for _, oldCount := range oldCounts {
				// Upgrade all oldCount copies
				upgCost := upgradeCost(oldDef, newDef, oldCount, 0, oldCount)
				freshCost := freshBuildCost(newDef, 0, oldCount)

				upgSum := sumCost(upgCost)
				freshSum := sumCost(freshCost)

				savingsPct := 0.0
				if freshSum > 0 {
					savingsPct = (freshSum - upgSum) / freshSum * 100.0
				}

				prodBefore := productionRate(oldDef, oldCount)
				prodAfter := productionRate(newDef, oldCount)

				boostRatio := 0.0
				if prodBefore > 0 {
					boostRatio = prodAfter / prodBefore
				}

				// Track summary stats
				if boostRatio > 0 {
					if boostRatio < s.minBoostRatio {
						s.minBoostRatio = boostRatio
					}
					if boostRatio > s.maxBoostRatio {
						s.maxBoostRatio = boostRatio
					}
				}

				// Check flags
				flags := []string{}
				isFree := isFreeUpgrade(upgCost)
				isOver := isOverpriced(upgCost, freshCost)
				hasNeg := hasNegativeDeltaBeforeGuard(oldDef, newDef, oldCount, 0, oldCount)
				boostHigh := boostRatio > 4.0

				if isFree {
					flags = append(flags, "FREE!")
					s.anyFreeUpgrade = true
					detail := fmt.Sprintf("%s→%s (old_cnt=%d)", oldDef.Key, newDef.Key, oldCount)
					s.freeDetails = append(s.freeDetails, detail)
					pairHasIssue = true
				}
				if isOver {
					flags = append(flags, "OVERPRICED!")
					s.anyOverpriced = true
					detail := fmt.Sprintf("%s→%s (old_cnt=%d, upg=%.0f fresh=%.0f)", oldDef.Key, newDef.Key, oldCount, upgSum, freshSum)
					s.overpricedDetails = append(s.overpricedDetails, detail)
					pairHasIssue = true
				}
				if hasNeg {
					flags = append(flags, "NEG_DELTA!")
					s.anyNegativeDelta = true
					pairHasIssue = true
				}
				if boostHigh {
					flags = append(flags, fmt.Sprintf("HIGH_BOOST(%.1fx)", boostRatio))
					detail := fmt.Sprintf("%s→%s (old_cnt=%d, boost=%.2fx)", oldDef.Key, newDef.Key, oldCount, boostRatio)
					s.highBoostDetails = append(s.highBoostDetails, detail)
					pairHasIssue = true
				}
				if savingsPct < 0 {
					flags = append(flags, fmt.Sprintf("NEG_SAVINGS(%.1f%%)", savingsPct))
					pairHasIssue = true
				}

				flagStr := strings.Join(flags, " ")

				fmt.Printf("  %-8d  %-14.0f  %-14.0f  %-8.1f  %-8.1f  %-8.1f  %-8.2f  %s\n",
					oldCount, upgSum, freshSum, savingsPct, prodBefore, prodAfter, boostRatio, flagStr)
			}

			_ = pairHasIssue
			fmt.Println()
		}
	}

	// Summary table
	fmt.Printf("\n%s\n", strings.Repeat("=", 100))
	fmt.Println("SUMMARY TABLE")
	fmt.Printf("%s\n", strings.Repeat("=", 100))
	fmt.Printf("%-20s  %-6s  %-10s  %-10s  %-12s  %-12s  %-12s\n",
		"Lineage", "Tiers", "MinBoost", "MaxBoost", "FreeUpgrades", "Overpriced", "NegDelta")
	fmt.Printf("%s\n", strings.Repeat("-", 100))

	totalIssues := 0
	for _, lineage := range lineageKeys {
		s := summaries[lineage]
		minB := "N/A"
		maxB := "N/A"
		if s.minBoostRatio != math.MaxFloat64 {
			minB = fmt.Sprintf("%.2fx", s.minBoostRatio)
		}
		if s.maxBoostRatio != -math.MaxFloat64 {
			maxB = fmt.Sprintf("%.2fx", s.maxBoostRatio)
		}
		freeStr := "no"
		if s.anyFreeUpgrade {
			freeStr = fmt.Sprintf("YES(%d)", len(s.freeDetails))
			totalIssues++
		}
		overStr := "no"
		if s.anyOverpriced {
			overStr = fmt.Sprintf("YES(%d)", len(s.overpricedDetails))
			totalIssues++
		}
		negStr := "no"
		if s.anyNegativeDelta {
			negStr = "YES"
			totalIssues++
		}
		fmt.Printf("%-20s  %-6d  %-10s  %-10s  %-12s  %-12s  %-12s\n",
			lineage, s.tierCount, minB, maxB, freeStr, overStr, negStr)
	}
	fmt.Printf("%s\n", strings.Repeat("-", 100))

	// Print detail sections for any issues
	fmt.Printf("\n--- ISSUE DETAILS ---\n")
	hasAny := false
	for _, lineage := range lineageKeys {
		s := summaries[lineage]
		if len(s.freeDetails) > 0 {
			hasAny = true
			fmt.Printf("\n[%s] FREE UPGRADES:\n", lineage)
			for _, d := range s.freeDetails {
				fmt.Printf("  * %s\n", d)
			}
		}
		if len(s.overpricedDetails) > 0 {
			hasAny = true
			fmt.Printf("\n[%s] OVERPRICED UPGRADES (upgrade > fresh build):\n", lineage)
			for _, d := range s.overpricedDetails {
				fmt.Printf("  * %s\n", d)
			}
		}
		if len(s.highBoostDetails) > 0 {
			hasAny = true
			fmt.Printf("\n[%s] HIGH BOOST RATIO (>4x production gain per resource spent):\n", lineage)
			for _, d := range s.highBoostDetails {
				fmt.Printf("  * %s\n", d)
			}
		}
		if s.anyNegativeDelta {
			hasAny = true
			fmt.Printf("\n[%s] NEGATIVE DELTA BEFORE GUARD: some upgrade deltas would be negative without the max(0) guard\n", lineage)
		}
	}
	if !hasAny {
		fmt.Println("  No issues found.")
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 100))
	if totalIssues == 0 {
		fmt.Println("RESULT: ALL CHECKS PASSED — no pricing anomalies found.")
	} else {
		fmt.Printf("RESULT: %d issue category(ies) detected across lineages — review details above.\n", totalIssues)
	}
	fmt.Printf("%s\n", strings.Repeat("=", 100))
}
