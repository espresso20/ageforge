package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// VillagerPanel shows recruitable villager types, their rates, and active bonuses
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

// UpdateState refreshes the villager panel display
func (vp *VillagerPanel) UpdateState(state game.GameState) {
	// Hash to skip redundant redraws
	var h uint64
	h = hashKey(state.Age)
	for key, vt := range state.Villagers.Types {
		if vt.Unlocked {
			h ^= hashKey(key) * 13
			h ^= uint64(vt.Count+1) * 7
			h ^= uint64(vt.IdleCount+1) * 3
			for res, n := range vt.Assignments {
				h ^= hashKey(res) * uint64(n+1) * 5
			}
		}
	}
	h ^= uint64(state.Research.Bonuses["gather_rate"]*1000) * 17
	if gu, ok := state.Prestige.Upgrades["gather_boost"]; ok {
		h ^= uint64(gu.Tier+1) * 11
	}
	h ^= uint64(state.Prestige.PassiveBonus*1000) * 19
	if h == vp.lastHash {
		return
	}
	vp.lastHash = h

	// Index default defs by key for rate lookup
	defs := make(map[string]game.VillagerTypeDef)
	for _, d := range game.DefaultVillagerTypes() {
		defs[d.Key] = d
	}

	// Gather bonus: research + prestige upgrade, both additive on base rate
	researchGather := state.Research.Bonuses["gather_rate"]
	prestigeGather := 0.0
	if gu, ok := state.Prestige.Upgrades["gather_boost"]; ok {
		prestigeGather = float64(gu.Tier) * 0.05
	}
	totalGatherBonus := researchGather + prestigeGather
	passiveBonus := state.Prestige.PassiveBonus // production_all (2% per prestige level)

	var sb strings.Builder
	anyUnlocked := false

	// Iterate in canonical definition order
	for _, def := range game.DefaultVillagerTypes() {
		vt, ok := state.Villagers.Types[def.Key]
		if !ok || !vt.Unlocked {
			continue
		}
		anyUnlocked = true

		// Name + count header
		idleColor := "green"
		if vt.Count > 0 && vt.IdleCount == 0 {
			idleColor = "yellow"
		}
		fmt.Fprintf(&sb, "[cyan::b]%s[-]  [%s]%d (idle:%d)[-]\n",
			def.Name, idleColor, vt.Count, vt.IdleCount)

		// Current assignments
		if len(vt.Assignments) > 0 {
			parts := make([]string, 0, len(vt.Assignments))
			for res, n := range vt.Assignments {
				if n > 0 {
					parts = append(parts, fmt.Sprintf("%s:%d", res, n))
				}
			}
			if len(parts) > 0 {
				fmt.Fprintf(&sb, "  [gray]→ %s[-]\n", strings.Join(parts, " "))
			}
		}

		// Gather rate (only for types that actually gather)
		if def.GatherRate > 0 {
			effective := def.GatherRate * (1.0 + totalGatherBonus) * (1.0 + passiveBonus)
			if totalGatherBonus > 0 || passiveBonus > 0 {
				fmt.Fprintf(&sb, "  [gray]%.3f[-][gray]→[-][green]%.3f[-][gray]/t[-]\n",
					def.GatherRate, effective)
			} else {
				fmt.Fprintf(&sb, "  [gray]%.3f/t[-]\n", def.GatherRate)
			}
		} else {
			fmt.Fprintf(&sb, "  [gray]combat only[-]\n")
		}

		sb.WriteByte('\n')
	}

	if !anyUnlocked {
		sb.WriteString("[gray]No villager types\\nunlocked yet.[-]\n")
	}

	// Active bonuses section
	if totalGatherBonus > 0 || passiveBonus > 0 {
		sb.WriteString("[gold::b]Bonuses[-]\n")
		if researchGather > 0 {
			fmt.Fprintf(&sb, " [cyan]+%.0f%% research[-]\n", researchGather*100)
		}
		if prestigeGather > 0 {
			fmt.Fprintf(&sb, " [cyan]+%.0f%% prestige[-]\n", prestigeGather*100)
		}
		if passiveBonus > 0 {
			fmt.Fprintf(&sb, " [yellow]+%.0f%% prod.all[-]\n", passiveBonus*100)
		}
	}

	vp.root.SetText(sb.String())
}
