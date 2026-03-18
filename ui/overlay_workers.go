package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

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
	fmt.Fprintf(&sb, "[white]Workers  [yellow]%d[white] / [green]%d[white]   Idle: [cyan]%d[white]   Food: [red]%.2f/tick[white]\n\n",
		total, maxPop, idle, foodDrain)

	// Build domain groups from buildings with assigned workers
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
		fmt.Fprintf(&sb, "[cyan]Idle: %d[white] — assign with: [cyan]assign <building> [count|all][-]\n", idle)
	} else {
		for _, domain := range groupOrder {
			grp := groupMap[domain]
			label, ok := panelDomainLabels[domain]
			if !ok {
				label = capitalize(domain)
			}
			classDisplay := ""
			if cls, found := config.WorkerClassByDomainAndAge(domain, state.Age); found && cls.ClassName != "" {
				classDisplay = fmt.Sprintf("  %s × %d", cls.ClassName, grp.Total)
			}
			fmt.Fprintf(&sb, "[cyan][%s][-][white]%s\n", label, classDisplay)
			for _, row := range grp.Rows {
				bar := assignBar(row.WorkersAssigned, row.Capacity, 10)
				fmt.Fprintf(&sb, "  %-28s %s [cyan]%d[white]/[green]%d[white]\n",
					row.Name, bar, row.WorkersAssigned, row.Capacity)
			}
			fmt.Fprintln(&sb)
		}
		if idle > 0 {
			fmt.Fprintf(&sb, "[cyan]Idle: %d[white] — assign with: [cyan]assign <building> [count|all][-]\n", idle)
		}
	}

	return sb.String()
}
