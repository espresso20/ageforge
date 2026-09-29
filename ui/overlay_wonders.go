package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// renderCurrentWonderSummary returns a short text block for the current-age wonder,
// prepended at the top of the wonders overlay so the player can see their active
// wonder status without scrolling.
func renderCurrentWonderSummary(state game.GameState) string {
	var sb strings.Builder
	var current *wonderInfo
	for _, w := range getWonderList() {
		if w.ageKey == state.Age {
			wCopy := w
			current = &wCopy
			break
		}
	}
	if current == nil {
		return ""
	}

	bs, hasBs := state.Buildings[current.key]
	built := hasBs && bs.Count > 0

	sb.WriteString("[gold]═══ This age's wonder ═══[-]\n")
	if built {
		fmt.Fprintf(&sb, " %s [gold]★ %s[-]  [green]Built[-]  [gray]%s[-]\n", WonderSpriteIcon(current.key), current.name, current.ageName)
	} else {
		fmt.Fprintf(&sb, " [yellow]○ %s[-]  [gray]%s, not built yet[-]\n", current.name, current.ageName)
	}

	// Effects
	for _, eff := range current.def.Effects {
		fmt.Fprintf(&sb, "   %s\n", formatEffect(eff))
	}
	fmt.Fprintf(&sb, "   [gold]%s[-]\n", wonderSpeedCapText)

	// Bank progress if not built
	if !built && hasBs {
		fmt.Fprintf(&sb, "\n [cyan]Wonder bank:[-]\n")
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
			fmt.Fprintf(&sb, "   [%s]%s %s %s / %s[-]\n", clr, game.ResourceName(k), bar, FormatNumber(banked), FormatNumber(need))
		}
		if bs.WonderBankFull {
			fmt.Fprintf(&sb, "   [green]✓ The bank is full. Build it with: build %s[-]\n", current.key)
		} else {
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", wonderCollectHint())
		}
		fmt.Fprintf(&sb, "   %s\n", wonderOverflowLine(state.WonderOverflow))
	}
	sb.WriteString("\n")
	return sb.String()
}

// wonderOverflowLine describes the wonder overflow switch for the wonder
// panels: whether production a full store would waste goes into the bank.
func wonderOverflowLine(on bool) string {
	if on {
		return theme.Paint(theme.RoleLabel, "Overflow") + " " + theme.Paint(theme.RolePositive, "on") +
			theme.Paint(theme.RoleDim, ": what a full store would waste is banked here (wonder overflow off)")
	}
	return theme.Paint(theme.RoleLabel, "Overflow") + " " + theme.Paint(theme.RoleWarning, "off") +
		theme.Paint(theme.RoleDim, ": production over a cap is lost (wonder overflow on)")
}

// wondersProvider generates the Wonders panel text from the current game state.
// Text only (no pixel art). Shows wonder name, age, description, effects, and build status.
func wondersProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	sb.WriteString(renderCurrentWonderSummary(state))

	wonders := getWonderList()
	builtCount := 0
	totalCount := len(wonders)
	maxSpeed := 1.0

	// Pre-count built wonders for header
	for _, w := range wonders {
		if bs, ok := state.Buildings[w.key]; ok && bs.Count > 0 {
			builtCount++
			maxSpeed += config.WonderSpeedCapStep
		}
	}

	// Header
	fmt.Fprintf(&sb, "[gold]═══ Wonders: %d / %d ═══[-]\n", builtCount, totalCount)
	fmt.Fprintf(&sb, " [cyan]Speed cap: %.1fx[-]   [gray]Each wonder raises the speed cap by %sx (set it with: speed %.1f).[-]\n\n",
		maxSpeed, config.FormatAmount(config.WonderSpeedCapStep), 1.0+config.WonderSpeedCapStep)

	// List each wonder
	for _, w := range wonders {
		bs, hasBs := state.Buildings[w.key]
		built := hasBs && bs.Count > 0
		unlocked := hasBs && bs.Unlocked

		if built {
			fmt.Fprintf(&sb, " %s [gold]★ %s[-]   [gray]%s[-]\n", WonderSpriteIcon(w.key), w.name, w.ageName)
			fmt.Fprintf(&sb, "   [green]Built[-]\n")
			if w.def.Description != "" {
				fmt.Fprintf(&sb, "   [gray]%s[-]\n", w.def.Description)
			}
			if len(w.def.Effects) > 0 {
				sb.WriteString("   [cyan]Effects:[-]\n")
				for _, eff := range w.def.Effects {
					fmt.Fprintf(&sb, "     %s\n", formatEffect(eff))
				}
			}
			fmt.Fprintf(&sb, "   [gold]%s[-]\n", wonderSpeedCapText)
		} else if unlocked {
			if bs.WonderBankFull {
				fmt.Fprintf(&sb, " [yellow]○ %s[-]   [gray]%s[-]   [green](bank full, ready to build)[-]\n",
					w.name, w.ageName)
			} else {
				// Compute fill percentage
				totalNeed, totalBanked := 0.0, 0.0
				for res, need := range w.def.BaseCost {
					totalNeed += need
					totalBanked += bs.WonderBank[res]
				}
				pct := 0.0
				if totalNeed > 0 {
					pct = totalBanked / totalNeed * 100
					if pct > 100 {
						pct = 100
					}
				}
				if pct > 0 {
					fmt.Fprintf(&sb, " [yellow]○ %s[-]   [gray]%s[-]   [yellow](%.0f%% banked)[-]\n",
						w.name, w.ageName, pct)
				} else {
					fmt.Fprintf(&sb, " [yellow]○ %s[-]   [gray]%s[-]\n", w.name, w.ageName)
				}

				// Per-resource bank breakdown
				if len(w.def.BaseCost) > 0 {
					costKeys := make([]string, 0, len(w.def.BaseCost))
					for k := range w.def.BaseCost {
						costKeys = append(costKeys, k)
					}
					sort.Strings(costKeys)
					for _, k := range costKeys {
						need := w.def.BaseCost[k]
						banked := bs.WonderBank[k]
						resPct := 0.0
						if need > 0 {
							resPct = banked / need
							if resPct > 1 {
								resPct = 1
							}
						}
						clr := "red"
						if resPct >= 1.0 {
							clr = "green"
						} else if resPct > 0 {
							clr = "yellow"
						}
						fmt.Fprintf(&sb, "     [%s]%s: %s / %s[-]\n",
							clr, game.ResourceName(k), FormatNumber(banked), FormatNumber(need))
					}
				}
				sb.WriteString("   [gray]Bank resources to build it: " + wonderCollectHint() + "[-]\n")
			}
			if w.def.Description != "" {
				fmt.Fprintf(&sb, "   [gray]%s[-]\n", w.def.Description)
			}
			if len(w.def.Effects) > 0 {
				sb.WriteString("   [cyan]Effects when built:[-]\n")
				for _, eff := range w.def.Effects {
					fmt.Fprintf(&sb, "     %s\n", formatEffect(eff))
				}
			}
			fmt.Fprintf(&sb, "   [gold]%s once built[-]\n", wonderSpeedCapText)
		} else {
			fmt.Fprintf(&sb, " [gray]? ???[-]   [gray]%s, locked[-]\n", w.ageName)
		}

		sb.WriteString("\n")
	}

	// Speed summary footer
	if builtCount == 0 {
		sb.WriteString("[gray]No wonders built yet. Each one you build raises the speed cap.[-]\n")
	} else {
		var builtNames []string
		for _, w := range wonders {
			if bs, ok := state.Buildings[w.key]; ok && bs.Count > 0 {
				builtNames = append(builtNames, w.name)
			}
		}
		fmt.Fprintf(&sb, "[gold]Built:[-] %s\n", strings.Join(builtNames, ", "))
	}

	return sb.String()
}
