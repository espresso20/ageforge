package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// VillagerPanel shows workers grouped by domain with building assignments.
type VillagerPanel struct {
	root     *tview.TextView
	lastHash uint64
}

// NewVillagerPanel creates the villager info panel
func NewVillagerPanel() *VillagerPanel {
	vp := &VillagerPanel{}
	vp.root = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	vp.root.SetBorder(true).SetTitle(" Villagers ").SetTitleColor(ColorTitle)
	return vp
}

// Primitive returns the underlying tview primitive
func (vp *VillagerPanel) Primitive() tview.Primitive {
	return vp.root
}

// domainGroups defines the display grouping and order.
var domainGroups = []struct {
	label   string
	domains []string
}{
	{"Materials", []string{"food", "lumber", "masonry"}},
	{"Knowledge", []string{"knowledge", "faith"}},
	{"Civil", []string{"trade", "engineering"}},
	{"Late-game", []string{"military", "metallurgy", "energy", "hacker", "astronaut"}},
}

// domainLabel returns a capitalized display label for a domain key.
func domainLabel(domain string) string {
	if domain == "" {
		return ""
	}
	return strings.ToUpper(domain[:1]) + domain[1:]
}

// UpdateState refreshes the villager panel display
func (vp *VillagerPanel) UpdateState(state game.GameState) {
	// Build a map: domainKey -> VillagerTypeState for present types (keys are already domain keys)
	domainToType := make(map[string]game.VillagerTypeState)
	for domainKey, vt := range state.Villagers.Types {
		domainToType[domainKey] = vt
	}

	// Hash-based dirty check — include total pop, food drain, and assignment contents
	var h uint64
	h = hashKey(state.Age)
	h ^= uint64(state.Villagers.TotalPop+1) * 31
	h ^= uint64(state.Villagers.FoodDrain*1000) * 37
	for key, vt := range state.Villagers.Types {
		if vt.Count > 0 || len(vt.Assignments) > 0 {
			h ^= hashKey(key) * 13
			h ^= uint64(vt.Count+1) * 7
			h ^= uint64(vt.IdleCount+1) * 3
			for bldKey, n := range vt.Assignments {
				h ^= hashKey(bldKey) * uint64(n+1) * 5
			}
		}
	}
	if h == vp.lastHash {
		return
	}
	vp.lastHash = h

	// Build a lookup: buildingKey -> BuildingState for assignment name resolution
	bldByKey := make(map[string]game.BuildingState, len(state.Buildings))
	for k, b := range state.Buildings {
		bldByKey[k] = b
	}

	var sb strings.Builder

	// Panel header: WORKERS  45 total · -3.2/t food
	foodDrainStr := fmt.Sprintf("%.1f", state.Villagers.FoodDrain)
	fmt.Fprintf(&sb, "[yellow]WORKERS[-]  [gray]%d total · -%s/t food[-]\n\n",
		state.Villagers.TotalPop, foodDrainStr)

	anyGroup := false
	for _, grp := range domainGroups {
		// Collect which domains in this group have entries
		type domainEntry struct {
			domainKey string
			vt        game.VillagerTypeState
		}
		var entries []domainEntry
		for _, dk := range grp.domains {
			if vt, ok := domainToType[dk]; ok {
				entries = append(entries, domainEntry{dk, vt})
			}
		}
		if len(entries) == 0 {
			continue
		}

		// Group header
		if anyGroup {
			sb.WriteByte('\n')
		}
		anyGroup = true
		fmt.Fprintf(&sb, "[yellow]── %s ──[-]\n", grp.label)

		for _, e := range entries {
			dk := e.domainKey
			vt := e.vt

			// Resolve class name from config
			className := ""
			if wcd, ok := config.WorkerClassByDomainAndAge(dk, state.Age); ok {
				className = wcd.ClassName
			}

			// Domain line
			label := domainLabel(dk)
			if className != "" {
				fmt.Fprintf(&sb, "[white]%s[-] [gray](%s)[-]  %d  idle: %d\n",
					label, className, vt.Count, vt.IdleCount)
			} else {
				fmt.Fprintf(&sb, "[white]%s[-]  %d  idle: %d\n",
					label, vt.Count, vt.IdleCount)
			}

			// Per-building assignment lines — sort for stable output
			if len(vt.Assignments) > 0 {
				type bldEntry struct {
					key      string
					name     string
					assigned int
					capacity int
				}
				var blds []bldEntry
				for bldKey, assigned := range vt.Assignments {
					if assigned <= 0 {
						continue
					}
					name := bldKey
					capacity := 0
					if bs, ok := bldByKey[bldKey]; ok {
						if bs.Name != "" {
							name = bs.Name
						}
						// WorkerCapacity is per-instance; total = capacity * count
						if bs.Count > 0 {
							capacity = bs.WorkerCapacity * bs.Count
						}
					}
					blds = append(blds, bldEntry{bldKey, name, assigned, capacity})
				}
				sort.Slice(blds, func(i, j int) bool {
					return blds[i].name < blds[j].name
				})
				if len(blds) == 0 {
					sb.WriteString("  [gray](no assignments)[-]\n")
				} else {
					for _, b := range blds {
						if b.capacity > 0 {
							fmt.Fprintf(&sb, "  %s  %d/%d\n", b.name, b.assigned, b.capacity)
						} else {
							fmt.Fprintf(&sb, "  %s  %d\n", b.name, b.assigned)
						}
					}
				}
			} else {
				sb.WriteString("  [gray](no assignments)[-]\n")
			}
		}
	}

	if !anyGroup {
		sb.WriteString("[gray]No villager types unlocked yet.[-]\n")
	}

	vp.root.SetText(sb.String())
}
