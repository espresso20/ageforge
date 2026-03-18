package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

const workerSectionWidth = 44

func workerSection(title string) string {
	bar := strings.Repeat("─", workerSectionWidth-len(title)-3)
	return fmt.Sprintf("[gray]── %s %s[-]\n", title, bar)
}

func workersProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	sb.WriteString("\n[gold]═══ Workers ═══[-]\n\n")

	wt := state.Workers.Types["worker"]
	total := state.Workers.TotalPop
	maxPop := state.Workers.MaxPop
	foodDrain := state.Workers.FoodDrain

	if !wt.Unlocked || total == 0 {
		fmt.Fprintf(&sb, "[white]Workers  [yellow]0[white] / [green]%d[white]\n\n", maxPop)
		fmt.Fprint(&sb, "[gray]No workers yet.\n")
		fmt.Fprint(&sb, "Recruit with: [cyan]recruit [count|max][-]\n")
		return sb.String()
	}

	idle := wt.IdleCount
	housingLeft := maxPop - total

	// ── Summary ──────────────────────────────────
	sb.WriteString(workerSection("Summary"))
	fmt.Fprintf(&sb, "  [white]Pop:[white] [yellow]%d[white] / [green]%d[-]   [white]Idle:[white] [cyan]%d[-]   [white]Housing left:[white] [green]%d[-]\n",
		total, maxPop, idle, housingLeft)

	// Food sustainability
	foodRS, hasFoodRS := state.Resources["food"]
	if hasFoodRS {
		netFood := foodRS.Rate // already net (includes worker drain)
		drainPerWorker := 0.0
		if total > 0 {
			drainPerWorker = foodDrain / float64(total)
		}
		breakEven := 0
		if drainPerWorker > 0 && foodRS.Rate+foodDrain > 0 {
			breakEven = int((foodRS.Rate + foodDrain) / drainPerWorker)
		}

		netColor := "green"
		netPrefix := "+"
		if netFood < 0 {
			netColor = "red"
			netPrefix = ""
		}
		fmt.Fprintf(&sb, "  [white]Food drain:[white] [red]%.2f/tick[-]   [white]Net food:[white] [%s]%s%.2f/tick[-]\n",
			foodDrain, netColor, netPrefix, netFood)
		if netFood >= 0 && breakEven > 0 {
			fmt.Fprintf(&sb, "  [gray]Sustains up to [white]%d[-][gray] workers at current food rate[-]\n", breakEven)
		} else if netFood < 0 {
			sb.WriteString("  [red]⚠ Food deficit — workers may starve[-]\n")
		}
	}
	sb.WriteString("\n")

	// ── Slot Utilization ─────────────────────────
	sb.WriteString(workerSection("Slot Utilization"))
	totalSlots := 0
	filledSlots := 0
	type openSlot struct {
		Name  string
		Open  int
	}
	var openSlots []openSlot

	for _, bs := range state.Buildings {
		if bs.WorkerDomain == "" || bs.Count <= 0 {
			continue
		}
		cap := bs.WorkerCapacity * bs.Count
		totalSlots += cap
		filledSlots += bs.WorkersAssigned
		open := cap - bs.WorkersAssigned
		if open > 0 {
			openSlots = append(openSlots, openSlot{Name: bs.Name, Open: open})
		}
	}
	sort.Slice(openSlots, func(i, j int) bool {
		return openSlots[i].Open > openSlots[j].Open
	})

	if totalSlots > 0 {
		pct := float64(filledSlots) / float64(totalSlots)
		bar := assignBar(filledSlots, totalSlots, 16)
		fmt.Fprintf(&sb, "  [white]Filled:[white] [cyan]%d[white] / [green]%d[-]  %s  [gray]%.0f%%[-]\n",
			filledSlots, totalSlots, bar, pct*100)
		if len(openSlots) > 0 {
			parts := make([]string, 0, 4)
			for i, s := range openSlots {
				if i >= 4 {
					break
				}
				parts = append(parts, fmt.Sprintf("[cyan]%s[white] (+%d)[-]", s.Name, s.Open))
			}
			fmt.Fprintf(&sb, "  [gray]Open slots:[white] %s[-]\n", strings.Join(parts, "[gray],[white] "))
		} else {
			sb.WriteString("  [green]✓ All slots filled[-]\n")
		}
	} else {
		sb.WriteString("  [gray]No worker buildings built yet[-]\n")
	}
	sb.WriteString("\n")

	// ── Domain Breakdown ─────────────────────────
	sb.WriteString(workerSection("Domain Breakdown"))

	type domainGroup struct {
		Domain string
		Rows   []buildingRow
		Total  int
	}
	groupMap := map[string]*domainGroup{}
	groupOrder := []string{}

	for bKey, bs := range state.Buildings {
		if bs.WorkersAssigned <= 0 || bs.WorkerDomain == "" {
			continue
		}
		domain := bs.WorkerDomain
		if _, exists := groupMap[domain]; !exists {
			groupMap[domain] = &domainGroup{Domain: domain}
			groupOrder = append(groupOrder, domain)
		}
		count := bs.Count
		if count <= 0 {
			count = 1
		}
		cap := bs.WorkerCapacity * count
		groupMap[domain].Rows = append(groupMap[domain].Rows, buildingRow{
			Key:             bKey,
			Name:            bs.Name,
			WorkersAssigned: bs.WorkersAssigned,
			Capacity:        cap,
		})
		groupMap[domain].Total += bs.WorkersAssigned
	}

	sort.Strings(groupOrder)
	for _, grp := range groupMap {
		sort.Slice(grp.Rows, func(i, j int) bool {
			return grp.Rows[i].Key < grp.Rows[j].Key
		})
	}

	if len(groupOrder) == 0 {
		fmt.Fprintf(&sb, "  [cyan]Idle: %d[white] — assign with: [cyan]assign <building> [count|all][-]\n", idle)
	} else {
		for _, domain := range groupOrder {
			grp := groupMap[domain]
			label, ok := panelDomainLabels[domain]
			if !ok {
				label = capitalize(domain)
			}
			classInfo := ""
			foodCostStr := ""
			if cls, found := config.WorkerClassByDomainAndAge(domain, state.Age); found && cls.ClassName != "" {
				classInfo = fmt.Sprintf(" %s × %d", cls.ClassName, grp.Total)
				if cls.FoodCost > 0 {
					foodCostStr = fmt.Sprintf("  [gray]%.3f food/tick each[-]", cls.FoodCost)
				}
			}
			fmt.Fprintf(&sb, "  [cyan][%s][-][white]%s[-]%s\n", label, classInfo, foodCostStr)
			for _, row := range grp.Rows {
				bar := assignBar(row.WorkersAssigned, row.Capacity, 10)
				fmt.Fprintf(&sb, "    %-26s %s [cyan]%d[white]/[green]%d[-]\n",
					row.Name, bar, row.WorkersAssigned, row.Capacity)
			}
			fmt.Fprintln(&sb)
		}
		if idle > 0 {
			fmt.Fprintf(&sb, "  [yellow]Idle: %d[white] — assign with: [cyan]assign <building> [count|all][-]\n", idle)
		}
	}

	return sb.String()
}
