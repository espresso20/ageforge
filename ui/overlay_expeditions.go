package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// expeditionsProvider renders the current mission board and missions underway.
func expeditionsProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	mil := state.Military
	fmt.Fprintf(&sb, " [gold]═══ Expeditions · %s ═══[-]\n", state.AgeName)
	tier := ""
	for _, slot := range mil.Board {
		if slot.Tier != tier {
			tier = slot.Tier
			words, _ := state.Ruleset().MissionWords(state.Age)
			kind := map[string]int{game.TierShort: config.WordsShort, game.TierLong: config.WordsLong, game.TierGrand: config.WordsGrandMilitary}[tier]
			heading := words.Tiers[kind]
			if tier == game.TierGrand {
				heading = "Grand"
			}
			fmt.Fprintf(&sb, "\n [gold]═══ %s ═══[-]\n", heading)
		}
		if slot.Offer == nil {
			fmt.Fprintf(&sb, " %d. Taken · new offer in %s\n", slot.Number, formatTicks(slot.RefillTicks, state))
			continue
		}
		e := *slot.Offer
		name := e.Name
		if e.Rare {
			name = strings.TrimPrefix(name, "Legendary ")
			name = "[gold]Legendary[-] " + name
		}
		fmt.Fprintf(&sb, " %d. %s\n", slot.Number, name)
		fmt.Fprintf(&sb, "   [gray]%.0f min · %.0f%% odds · haul %s · cost %s[-]\n", e.LengthMinutes, e.SuccessChance*100, expeditionHaul(e, state), expeditionOfferCost(e, state))
		if e.Tier == game.TierGrand {
			if e.Grand == game.GrandMilitary {
				fmt.Fprintf(&sb, "   military · %d %s\n", e.SoldiersNeeded, pluralize("soldier", e.SoldiersNeeded))
			} else {
				fmt.Fprintf(&sb, "   civil · %s\n", expeditionOfferCost(e, state))
			}
		}
		fmt.Fprintf(&sb, "   %s\n", e.Description)
		if !e.CanLaunch && e.LaunchBlockReason != "" {
			fmt.Fprintf(&sb, "   [red]%s[-]\n", e.LaunchBlockReason)
		}
	}
	sb.WriteString("\n [gold]═══ Under way ═══[-]\n")
	for _, active := range []*game.ExpeditionSnapshot{mil.ActiveShort, mil.ActiveLong, mil.ActiveGrand} {
		if active != nil {
			fmt.Fprintf(&sb, " %s · %s · %.0f%% odds · %s left\n", active.Tier, active.Name, active.SuccessChance*100, formatTicks(active.TicksLeft, state))
		}
	}
	if auto := mil.AutoExpedition; auto.Active {
		if auto.Starved {
			sb.WriteString("\n [yellow]Geographic Society: cannot pay the expedition cost.[-]\n")
		} else {
			fmt.Fprintf(&sb, "\n [cyan]Geographic Society:[-] next party in %s\n", formatTicks(auto.TicksLeft, state))
		}
	} else {
		hint := "A later age brings it."
		if state.Age == "colonial_age" {
			hint = "The Industrial Age brings it."
		}
		sb.WriteString("\n" + theme.Paint(theme.RoleDim, " A Geographic Society scouts automatically. "+hint) + "\n")
	}
	sb.WriteString("\n [gray]expedition <n> to send · expedition list[-]\n")
	return sb.String()
}
