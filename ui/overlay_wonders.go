package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
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

	sb.WriteString("[gold]═══ Current Age Wonder ═══[-]\n")
	if built {
		fmt.Fprintf(&sb, " %s [gold]★ %s[-]  [green]BUILT[-]  [gray]%s[-]\n", WonderSpriteIcon(), current.name, current.ageName)
	} else {
		fmt.Fprintf(&sb, " [yellow]○ %s[-]  [gray]%s — Not yet built[-]\n", current.name, current.ageName)
	}

	// Effects
	for _, eff := range current.def.Effects {
		fmt.Fprintf(&sb, "   %s\n", formatEffect(eff))
	}
	fmt.Fprintf(&sb, "   [gold]+0.5x game speed[-]\n")

	// Bank progress if not built
	if !built && hasBs {
		fmt.Fprintf(&sb, "\n [cyan]Wonder Bank:[-]\n")
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
			fmt.Fprintf(&sb, "   [%s]%s %s %s / %s[-]\n", clr, k, bar, FormatNumber(banked), FormatNumber(need))
		}
		if bs.WonderBankFull {
			fmt.Fprintf(&sb, "   [green]✓ Bank full! Type 'build %s'[-]\n", current.key)
		} else {
			fmt.Fprintf(&sb, "   [gray]wonder collect <res> <amt|all>[-]\n")
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

// wondersProvider generates the wonders overlay text from the current game state.
// Text-only — no pixel art images. Shows wonder name, age, description, effects, and build status.
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
			maxSpeed += 0.5
		}
	}

	// Header
	fmt.Fprintf(&sb, "[gold]═══ Wonders: %d / %d ═══[-]\n", builtCount, totalCount)
	fmt.Fprintf(&sb, " [cyan]Max Speed: %.1fx[-]   [gray]Each wonder grants +0.5x game speed[-]\n\n",
		maxSpeed)

	// List each wonder
	for _, w := range wonders {
		bs, hasBs := state.Buildings[w.key]
		built := hasBs && bs.Count > 0
		unlocked := hasBs && bs.Unlocked

		if built {
			fmt.Fprintf(&sb, " %s [gold]★ %s[-]   [gray]%s[-]\n", WonderSpriteIcon(), w.name, w.ageName)
			fmt.Fprintf(&sb, "   [green]BUILT[-]\n")
			if w.def.Description != "" {
				fmt.Fprintf(&sb, "   [gray]%s[-]\n", w.def.Description)
			}
			if len(w.def.Effects) > 0 {
				sb.WriteString("   [cyan]Effects:[-]\n")
				for _, eff := range w.def.Effects {
					fmt.Fprintf(&sb, "     %s\n", formatEffect(eff))
				}
			}
			sb.WriteString("   [gold]+0.5x game speed[-]\n")
		} else if unlocked {
			if bs.WonderBankFull {
				fmt.Fprintf(&sb, " [yellow]○ %s[-]   [gray]%s[-]   [green][BANK FULL — ready to build!][-]\n",
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
							clr, k, FormatNumber(banked), FormatNumber(need))
					}
				}
				sb.WriteString("   [gray]Bank resources to build (wonder collect <res> <amt|all>)[-]\n")
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
			sb.WriteString("   [gold]+0.5x game speed when built[-]\n")
		} else {
			fmt.Fprintf(&sb, " [gray]? ???[-]   [gray]%s — locked[-]\n", w.ageName)
		}

		sb.WriteString("\n")
	}

	// Speed summary footer
	if builtCount == 0 {
		sb.WriteString("[gray]No wonders built yet. Build the wonder for your current age to unlock speed bonuses.[-]\n")
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
