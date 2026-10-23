package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// renderCurrentWonderSummary returns a short text block for the current-age wonder,
// prepended at the top of the wonders overlay so the player can see their active
// wonder status without scrolling.
func renderCurrentWonderSummary(state game.GameState) string {
	var sb strings.Builder
	var current *wonderInfo
	for _, w := range getWonderList(state.Ruleset()) {
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

	sb.WriteString(vibe.Line("ui.overlay_wonders.renderCurrentWonderSummary.01", "[gold]═══ This age's wonder ═══[-]\n"))
	if built {
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_wonders.renderCurrentWonderSummary.02", " %s [gold]★ %s[-]  [green]Built[-]  [gray]%s[-]\n", WonderSpriteIcon(current.key), current.name, current.ageName))
	} else {
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_wonders.renderCurrentWonderSummary.03", " [yellow]○ %s[-]  [gray]%s, not built yet[-]\n", current.name, current.ageName))
	}

	// Effects
	for _, eff := range current.def.Effects {
		fmt.Fprintf(&sb, "   %s%s\n", formatEffect(eff), capTag(state, eff, built, "-"))
	}

	// Its keystone tech, and whether it is researched
	if line := wonderKeystoneLine(state, current.key); line != "" && !built {
		fmt.Fprintf(&sb, "\n   %s\n", line)
	}

	// Bank progress if not built
	if !built && hasBs {
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_wonders.renderCurrentWonderSummary.04", "\n [cyan]Wonder bank:[-]\n"))
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
			fmt.Fprintf(&sb, "   %s\n", wonderBuildLine(state, current.key))
		} else {
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", wonderCollectHint())
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

// wonderBankPercent returns the floored share of a wonder's price that has
// been paid, in whole percent: game.WonderBankShare, the one reading the Next
// Age strip, this panel, the Map and the Advance panel all share. A surplus
// in one resource cannot pay another part of the price, and each part counts
// for what it is worth in the age's prices.
func wonderBankPercent(state game.GameState, wonder string) float64 {
	bs := state.Buildings[wonder]
	if bs.WonderBankFull {
		return 100
	}
	def, ok := state.Ruleset().Building(wonder)
	if !ok {
		return 0
	}
	return math.Floor(game.WonderBankShare(state.Ruleset(), def.RequiredAge, def.BaseCost, bs.WonderBank) * 100)
}

// wondersProvider generates the Wonders panel text from the current game state.
// Text only (no pixel art). Shows wonder name, age, description, effects, and build status.
func wondersProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	sb.WriteString(renderCurrentWonderSummary(state))

	wonders := getWonderList(state.Ruleset())
	builtCount := 0
	totalCount := len(wonders)

	// Pre-count built wonders for header
	for _, w := range wonders {
		if bs, ok := state.Buildings[w.key]; ok && bs.Count > 0 {
			builtCount++
		}
	}

	// Header
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_wonders.wondersProvider.01", "[gold]═══ Wonders: %d / %d ═══[-]\n", builtCount, totalCount))
	if builtCount > 0 {
		sb.WriteString(vibe.Line("ui.overlay_wonders.wondersProvider.15", "[gray]Press G for a gallery of the wonders you have built.[-]\n"))
	}

	// List each wonder. A locked wonder of an age the player cannot see named
	// yet is only counted (spoilers.go).
	sight := game.SightOf(&state)
	laterLocked := 0
	for _, w := range wonders {
		bs, hasBs := state.Buildings[w.key]
		built := hasBs && bs.Count > 0
		unlocked := hasBs && bs.Unlocked
		if !built && !unlocked && !sight.Age(w.ageKey) {
			laterLocked++
			continue
		}

		if built {
			fmt.Fprintf(&sb, " %s [gold]★ %s[-]   [gray]%s[-]\n", WonderSpriteIcon(w.key), w.name, w.ageName)
			fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_wonders.wondersProvider.02", "   [green]Built[-]\n"))
			if w.def.Description != "" {
				fmt.Fprintf(&sb, "   [gray]%s[-]\n", w.def.Description)
			}
			if len(w.def.Effects) > 0 {
				sb.WriteString(vibe.Line("ui.overlay_wonders.wondersProvider.03", "   [cyan]Effects:[-]\n"))
				for _, eff := range w.def.Effects {
					fmt.Fprintf(&sb, "     %s%s\n", formatEffect(eff), capTag(state, eff, true, "-"))
				}
			}
		} else if unlocked {
			if bs.WonderBankFull {
				ready := vibe.Line("ui.overlay_wonders.wondersProvider.04", "[green](bank full, ready to build)[-]")
				if bs.NeedsTech != "" {
					ready = vibe.Line("ui.overlay_wonders.wondersProvider.05", "[yellow](bank full, waiting for its keystone)[-]")
				}
				fmt.Fprintf(&sb, " [yellow]○ %s[-]   [gray]%s[-]   %s\n", w.name, w.ageName, ready)
			} else {
				pct := wonderBankPercent(state, w.key)
				if pct > 0 {
					fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_wonders.wondersProvider.06", " [yellow]○ %s[-]   [gray]%s[-]   [yellow](%.0f%% banked)[-]\n",
						w.name, w.ageName, pct))
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
				sb.WriteString(vibe.Line("ui.overlay_wonders.wondersProvider.07", "   [gray]Bank resources to build it: ") + wonderCollectHint() + "[-]\n")
			}
			if line := wonderKeystoneLine(state, w.key); line != "" {
				fmt.Fprintf(&sb, "   %s\n", line)
			}
			if w.def.Description != "" {
				fmt.Fprintf(&sb, "   [gray]%s[-]\n", w.def.Description)
			}
			if len(w.def.Effects) > 0 {
				sb.WriteString(vibe.Line("ui.overlay_wonders.wondersProvider.08", "   [cyan]Effects when built:[-]\n"))
				for _, eff := range w.def.Effects {
					fmt.Fprintf(&sb, "     %s%s\n", formatEffect(eff), capTag(state, eff, false, "-"))
				}
			}
		} else {
			fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_wonders.wondersProvider.09", " [gray]? ???[-]   [gray]%s, locked[-]\n", w.ageName))
		}

		sb.WriteString("\n")
	}
	if laterLocked > 0 {
		sb.WriteString(" " + theme.Paint(theme.RoleDim, vibe.Sprintf("ui.overlay_wonders.wondersProvider.10", "? ???   %s in ages still to come, locked", textfmt.Count(laterLocked, vibe.Line("ui.overlay_wonders.wondersProvider.11", "more wonder"), vibe.Line("ui.overlay_wonders.wondersProvider.12", "more wonders")))) + "\n\n")
	}

	// Summary footer
	if builtCount == 0 {
		sb.WriteString(vibe.Line("ui.overlay_wonders.wondersProvider.13", "[gray]No wonders built yet.[-]\n"))
	} else {
		var builtNames []string
		for _, w := range wonders {
			if bs, ok := state.Buildings[w.key]; ok && bs.Count > 0 {
				builtNames = append(builtNames, w.name)
			}
		}
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_wonders.wondersProvider.14", "[gold]Built:[-] %s\n", strings.Join(builtNames, ", ")))
	}

	return sb.String()
}
