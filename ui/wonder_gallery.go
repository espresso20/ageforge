package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// wonderProgressBar returns a simple fill bar with colored segments.
// Filled portion uses BarFillColor (purple), empty portion uses BarEmptyColor (dark gray).
func wonderProgressBar(pct float64, width int) string {
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	return "[" + BarFillColor + "]" + strings.Repeat("█", filled) + "[" + BarEmptyColor + "]" + strings.Repeat("░", empty) + "[-]"
}

// wonderInfo holds config + state for a wonder
type wonderInfo struct {
	key     string
	name    string
	ageKey  string
	ageName string
	def     config.BuildingDef
}

// getWonderList returns all wonders in age order
func getWonderList() []wonderInfo {
	ages := config.Ages()
	buildings := config.BuildingByKey()

	var wonders []wonderInfo
	for _, age := range ages {
		for _, bKey := range age.UnlockBuildings {
			if def, ok := buildings[bKey]; ok && def.Category == "wonder" {
				wonders = append(wonders, wonderInfo{
					key:     bKey,
					name:    def.Name,
					ageKey:  age.Key,
					ageName: age.Name,
					def:     def,
				})
				break // one wonder per age
			}
		}
	}
	return wonders
}

// ─── Current Age Wonder Panel (dashboard strip) ───

// WonderPanel shows the current age's wonder with pixel art + perks
type WonderPanel struct {
	root     *tview.Flex
	lastHash uint64
}

// NewWonderPanel creates the single-wonder panel for the dashboard
func NewWonderPanel() *WonderPanel {
	wp := &WonderPanel{}
	wp.root = tview.NewFlex().SetDirection(tview.FlexColumn)
	wp.root.SetBorder(true).SetTitle(" Wonder ").SetTitleColor(ColorTitle)
	return wp
}

// Primitive returns the underlying tview primitive
func (wp *WonderPanel) Primitive() tview.Primitive {
	return wp.root
}

// UpdateState refreshes the current-age wonder display
func (wp *WonderPanel) UpdateState(state game.GameState) {
	// Find the wonder for the current age
	var current *wonderInfo
	for _, w := range getWonderList() {
		if w.ageKey == state.Age {
			wCopy := w
			current = &wCopy
			break
		}
	}

	// Hash to detect changes
	var h uint64
	h = hashKey(state.Age)
	if current != nil {
		if bs, ok := state.Buildings[current.key]; ok {
			h ^= uint64(bs.Count)*7 + hashKey(current.key)
			if bs.Unlocked {
				h ^= 13
			}
			if bs.WonderBankFull {
				h ^= 997
			}
			for res, amt := range bs.WonderBank {
				h ^= hashKey(res) * uint64(amt+1)
			}
		}
	}
	// Also hash wonder count for speed display
	for _, w := range getWonderList() {
		if bs, ok := state.Buildings[w.key]; ok && bs.Count > 0 {
			h ^= hashKey(w.key) * 5
		}
	}
	if h == wp.lastHash {
		return
	}
	wp.lastHash = h

	wp.root.Clear()

	_, _, totalW, ht := wp.root.GetInnerRect()
	if totalW < 10 || ht < 3 {
		return
	}

	if current == nil {
		// No wonder for this age (shouldn't happen, but handle gracefully)
		tv := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
		tv.SetText("[gray]No wonder available this age[-]")
		wp.root.AddItem(tv, 0, 1, false)
		return
	}

	built := false
	unlocked := false
	if bs, ok := state.Buildings[current.key]; ok {
		built = bs.Count > 0
		unlocked = bs.Unlocked
	}
	_ = unlocked

	// Text-only layout
	infoTV := tview.NewTextView().SetDynamicColors(true)

	var sb strings.Builder
	if built {
		fmt.Fprintf(&sb, "%s [gold::b]★ %s[-]\n", WonderSpriteIcon(current.key), current.name)
		fmt.Fprintf(&sb, "[green]BUILT[-] — [gray]%s[-]\n\n", current.ageName)
	} else {
		fmt.Fprintf(&sb, "[yellow::b]%s[-]\n", current.name)
		fmt.Fprintf(&sb, "[gray]%s — Not yet built[-]\n\n", current.ageName)
	}

	// Show effects/perks
	fmt.Fprintf(&sb, "[cyan]Perks:[-]\n")
	for _, eff := range current.def.Effects {
		fmt.Fprintf(&sb, "  %s\n", formatEffect(eff))
	}
	fmt.Fprintf(&sb, "  [gold]+0.5x game speed[-]\n")

	// Bank progress if not built
	if !built {
		if bs, ok := state.Buildings[current.key]; ok {
			fmt.Fprintf(&sb, "\n[cyan]Wonder Bank:[-]\n")
			costKeys := make([]string, 0, len(current.def.BaseCost))
			for k := range current.def.BaseCost {
				costKeys = append(costKeys, k)
			}
			sort.Strings(costKeys)
			for _, k := range costKeys {
				need := current.def.BaseCost[k]
				banked := bs.WonderBank[k]
				pct := 0.0
				if need > 0 {
					pct = banked / need
					if pct > 1 {
						pct = 1
					}
				}
				clr := "red"
				if pct >= 1.0 {
					clr = "green"
				} else if pct > 0 {
					clr = "yellow"
				}
				bar := wonderProgressBar(pct, 8)
				fmt.Fprintf(&sb, "  [%s]%s %s %s / %s[-]\n", clr, k, bar, FormatNumber(banked), FormatNumber(need))
			}
			if bs.WonderBankFull {
				fmt.Fprintf(&sb, "  [green]✓ Bank full! Type 'build %s'[-]\n", current.key)
			} else {
				fmt.Fprintf(&sb, "  [gray]wonder collect <res> <amt|all>[-]\n")
			}
		}
		fmt.Fprintf(&sb, "\n[gray]Build ticks: %d[-]\n", current.def.BuildTicks)
	}

	// Wonder count / speed
	wonderCount := 0
	for _, w := range getWonderList() {
		if bs, ok := state.Buildings[w.key]; ok && bs.Count > 0 {
			wonderCount++
		}
	}
	maxSpeed := 1.0 + float64(wonderCount)*0.5
	fmt.Fprintf(&sb, "\n[gold]Wonders built: %d[-] — [cyan]Max speed: %.1fx[-]", wonderCount, maxSpeed)

	infoTV.SetText(sb.String())

	wp.root.AddItem(infoTV, 0, 1, false)
}

// formatEffect formats a building effect for display
func formatEffect(eff config.Effect) string {
	switch eff.Type {
	case "production":
		return fmt.Sprintf("[green]+%g %s/tick[-]", eff.Value, eff.Target)
	case "capacity":
		return fmt.Sprintf("[yellow]+%.0f %s cap[-]", eff.Value, eff.Target)
	case "storage":
		if eff.Target == "all" {
			return fmt.Sprintf("[yellow]+%.0f all storage[-]", eff.Value)
		}
		return fmt.Sprintf("[yellow]+%.0f %s storage[-]", eff.Value, eff.Target)
	case "bonus":
		return fmt.Sprintf("[cyan]+%.0f%% %s[-]", eff.Value*100, eff.Target)
	default:
		return fmt.Sprintf("%s %s: %g", eff.Type, eff.Target, eff.Value)
	}
}

