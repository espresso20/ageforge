package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
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
		fmt.Fprint(&sb, "[gray]No workers yet. They come on their own once you have housing and a building with worker slots,\n")
		fmt.Fprint(&sb, "or recruit by hand with: [cyan]recruit [count|max][-]\n\n")
		sb.WriteString(workerSection("Shares"))
		writeShares(&sb, state)
		return sb.String()
	}

	idle := wt.IdleCount
	housingLeft := maxPop - total

	// ── Morale ───────────────────────────────────
	// Banded model: bar + status recolour by band (green high / neutral mid /
	// red low) off state.MoraleMultiplier; no "penalty" copy in the neutral band.
	sb.WriteString(workerSection("Morale"))
	band := computeMoraleBand(state.Morale, state.MoraleMultiplier)
	moraleBar := moraleBandBar(int(state.Morale*20), 20, 20, band.Color)
	capStr := ""
	if state.MoraleCap > 1.0 {
		capStr = fmt.Sprintf("  [gray]cap %s[-]", textfmt.Percent(state.MoraleCap))
	}
	fmt.Fprintf(&sb, "  [white]Morale:[white] [%s]%.0f%%[-]%s  %s\n", band.Color, state.Morale*100, capStr, moraleBar)
	switch {
	case band.Bonus:
		fmt.Fprintf(&sb, "  [green]%s[-]\n", band.Status)
	case band.Penalty:
		fmt.Fprintf(&sb, "  [red]%s[-]\n", band.Status)
	default:
		fmt.Fprintf(&sb, "  [white]%s[-]\n", band.Status)
	}
	sb.WriteString("\n")

	// ── Summary ──────────────────────────────────
	sb.WriteString(workerSection("Summary"))
	fmt.Fprintf(&sb, "  [white]Population:[white] [yellow]%d[white] / [green]%d[-]   [white]Idle:[white] [cyan]%d[-]   [white]Housing left:[white] [green]%d[-]\n",
		total, maxPop, idle, housingLeft)

	// Food sustainability. Every worker eats the same amount, whatever it
	// does: the age's food-class cost (game.WorkerManager.FoodDrain), so the
	// per-worker figure is the total drain over the headcount.
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
		if netFood < 0 {
			netColor = "red"
		}
		fmt.Fprintf(&sb, "  [white]Food use:[white] [red]%s[-] [gray](%s food each)[-]   [white]Net food:[white] [%s]%s[-]\n",
			textfmt.Rate(-foodDrain), strings.TrimPrefix(textfmt.RateValue(drainPerWorker), "+"), netColor, textfmt.Rate(netFood))
		if netFood >= 0 && breakEven > 0 {
			fmt.Fprintf(&sb, "  [gray]Current food production feeds up to [white]%s[-][gray].[-]\n", textfmt.Count(breakEven, "worker", "workers"))
		} else if netFood < 0 {
			sb.WriteString("  [red]⚠ Food is falling. Workers starve when it runs out.[-]\n")
		}
	}
	sb.WriteString("\n")

	// ── Shares ───────────────────────────────────
	sb.WriteString(workerSection("Shares"))
	writeShares(&sb, state)
	sb.WriteString("\n")

	// ── Building slots ───────────────────────────
	sb.WriteString(workerSection("Building slots"))
	totalSlots := 0
	filledSlots := 0
	type openSlot struct {
		Name string
		Open int
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
	// Ties break on the name: state.Buildings is a map, so without it equal
	// counts would swap places between refreshes.
	sort.Slice(openSlots, func(i, j int) bool {
		if openSlots[i].Open != openSlots[j].Open {
			return openSlots[i].Open > openSlots[j].Open
		}
		return openSlots[i].Name < openSlots[j].Name
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
		sb.WriteString("  [gray]No worker buildings built yet.[-]\n")
	}
	sb.WriteString("\n")

	// ── By domain ────────────────────────────────
	sb.WriteString(workerSection("By domain"))

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
		fmt.Fprintf(&sb, "  [cyan]Idle: %d[white]. Assign with: [cyan]assign <building> [count|all][-]\n", idle)
	} else {
		for _, domain := range groupOrder {
			grp := groupMap[domain]
			label, ok := panelDomainLabels[domain]
			if !ok {
				label = capitalize(domain)
			}
			// note: no per-domain food figure here. The class table carries one,
			// but the game charges every worker the food class's cost (see the
			// summary above), so showing the domain's would be a number nobody pays.
			classInfo := ""
			if cls, found := config.WorkerClassByDomainAndAge(domain, state.Age); found && cls.ClassName != "" {
				classInfo = fmt.Sprintf(" %s × %d", cls.ClassName, grp.Total)
			}
			fmt.Fprintf(&sb, "  [cyan]%s[-][white]%s[-]\n", label, classInfo)
			for _, row := range grp.Rows {
				bar := assignBar(row.WorkersAssigned, row.Capacity, 10)
				fmt.Fprintf(&sb, "    %-26s %s [cyan]%d[white]/[green]%d[-]\n",
					row.Name, bar, row.WorkersAssigned, row.Capacity)
			}
			fmt.Fprintln(&sb)
		}
		if idle > 0 {
			fmt.Fprintf(&sb, "  [yellow]Idle: %d[white]. Assign with: [cyan]assign <building> [count|all][-]\n", idle)
		}
	}

	return sb.String()
}

// writeShares is the Workers panel's Shares section: what auto-recruit is
// doing, then each domain's share of the workforce (set, or auto: by its
// buildings' slots) with its workers and slots.
func writeShares(sb *strings.Builder, state game.GameState) {
	fmt.Fprintf(sb, "  [white]Auto-recruit:[-] %s\n", recruitLine(state))
	rows := game.ShareRows(state)
	if len(rows) == 0 {
		sb.WriteString("  [gray]No worker buildings yet: every domain is on auto.[-]\n")
		return
	}
	for _, r := range rows {
		share := fmt.Sprintf("[white]%4s[-] [gray]auto[-]", textfmt.Percent(r.Percent/100))
		if r.Set {
			share = fmt.Sprintf("[yellow]%4s[-] [yellow]set[-] ", game.SharePercent(math.Round(r.Percent)))
		}
		fmt.Fprintf(sb, "  [cyan]%-12s[-] %s  %s [cyan]%d[white]/[green]%d[-]\n",
			r.Name, share, assignBar(r.Workers, r.Slots, 10), r.Workers, r.Slots)
	}
	sb.WriteString("  [gray]Set a share with:[-] [cyan]workers share <domain> <percent|auto>[-]\n")
}

// recruitLine says what auto-recruit is doing (game.RecruitStatus).
func recruitLine(state game.GameState) string {
	switch game.RecruitStatus(state) {
	case game.RecruitOff:
		return "[gray]off (turn it on with[-] [cyan]workers auto-recruit on[-][gray])[-]"
	case game.RecruitHeld:
		return "[yellow]waiting[-] [gray]after your last worker command, " + formatTicks(state.Workers.HoldTicks, state) + " more[-]"
	case game.RecruitHousing:
		return "[green]on[-] [gray](no housing left: build housing for more workers)[-]"
	case game.RecruitSlots:
		return "[green]on[-] [gray](every worker slot is filled)[-]"
	case game.RecruitFood:
		return "[green]on[-] [yellow](waiting for food: build or staff food buildings)[-]"
	case game.RecruitStarve:
		return "[green]on[-] [red](paused: food has run out)[-]"
	}
	return "[green]on[-] [gray](recruiting as slots open)[-]"
}
