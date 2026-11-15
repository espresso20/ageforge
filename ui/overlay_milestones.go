package ui

import (
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// milestonesPage is the Milestones panel: the progress and title, and a
// section for each chain, a card for each milestone you can see.
func milestonesPage(state game.GameState) deckPage {
	ms := state.Milestones
	pg := deckPage{Title: vibe.Line("ui.overlay_milestones.milestonesPage.01", "Milestones")}
	pg.Head = []xseg{{vibe.Sprintf("ui.overlay_milestones.milestonesPage.02", "Progress: %d/%d", ms.CompletedCount, ms.TotalCount), xBright}}
	if ms.CurrentTitle != "" {
		pg.Head = append(pg.Head, xseg{"  \"" + ms.CurrentTitle + "\"", xTitle})
	}

	// Build chain lookup by category
	chainByCategory := make(map[string]game.ChainInfo)
	for _, chain := range ms.Chains {
		chainByCategory[chain.Category] = chain
	}

	// Group milestones by category
	categoryMilestones := make(map[string][]game.MilestoneInfo)
	for _, m := range ms.Milestones {
		categoryMilestones[m.Category] = append(categoryMilestones[m.Category], m)
	}

	// Categories in config order. A hand-kept copy here once dropped the
	// epoch milestones, which the header still counted.
	catNames := config.MilestoneCategoryNames()

	for _, cat := range config.MilestoneCategoryOrder() {
		milestones := categoryMilestones[cat]
		if len(milestones) == 0 {
			continue
		}
		sec := deckSection{Title: catNames[cat], MinW: 34}
		if chain, ok := chainByCategory[cat]; ok {
			sec.Meta = vibe.Sprintf("ui.overlay_milestones.milestonesPage.03", "%d of %d", chain.CompletedCount, chain.TotalCount)
			if chain.Complete {
				sec.Notes = append(sec.Notes, deckLine{Value: "✓ " + chain.Title, K: xGood})
				if chain.BoostActive {
					sec.Notes = append(sec.Notes, deckLine{Value: vibe.Line("ui.overlay_milestones.milestonesPage.04", "⚡ boost active"), K: xGood})
				}
			}
		}

		// Sort: completed first, then by name
		sort.Slice(milestones, func(i, j int) bool {
			if milestones[i].Completed != milestones[j].Completed {
				return milestones[i].Completed
			}
			return milestones[i].Name < milestones[j].Name
		})

		hiddenCount, now, nowShare := 0, -1, 0.0
		for _, m := range milestones {
			if !m.Visible {
				hiddenCount++
				continue
			}
			c := deckCard{Title: m.Name}
			reward := m.RewardText + rewardCapNotes(state, m.Rewards, m.Completed)
			if m.Completed {
				c.Chip, c.ChipK, c.Quiet = vibe.Line("ui.overlay_milestones.milestonesPage.05", "✓ done"), xChipShort, true
			} else {
				c.Desc = m.Description
				// The first condition not yet met is the meter; the rest are lines.
				share, n := 0.0, 0
				for _, p := range m.Progress {
					n++
					f := 1.0
					if !p.Met && p.Target > 0 {
						f = max(0, min(1, p.Current/p.Target))
					}
					share += f
					switch {
					case p.Met:
						c.Lines = append(c.Lines, deckLine{Value: "✓ " + p.Label, K: xGood})
					case !c.Meter:
						c.Meter, c.Share = true, f
						c.MeterText = textfmt.Number(p.Current) + "/" + textfmt.Number(p.Target) + " " + p.Label
					default:
						c.Lines = append(c.Lines, deckLine{Value: textfmt.Number(p.Current) + "/" + textfmt.Number(p.Target) + " " + p.Label})
					}
				}
				// The card closest to done, with some progress, is the one in hand.
				if n > 0 && share/float64(n) > nowShare {
					now, nowShare = len(sec.Cards), share/float64(n)
				}
			}
			if m.RewardText != "" {
				c.Lines = append(c.Lines, deckLine{vibe.Line("ui.overlay_milestones.milestonesPage.06", "Reward"), reward, xGood})
			}
			sec.Cards = append(sec.Cards, c)
		}
		if now >= 0 {
			sec.Cards[now].Chip, sec.Cards[now].ChipK = vibe.Line("ui.overlay_milestones.milestonesPage.07", "now"), xChipCivic
		}
		if hiddenCount > 0 {
			sec.Notes = append(sec.Notes, deckLine{Value: "+ " + textfmt.Count(hiddenCount, vibe.Line("ui.overlay_milestones.milestonesPage.08", "hidden milestone"), vibe.Line("ui.overlay_milestones.milestonesPage.09", "hidden milestones")), K: xDim})
		}
		pg.Sections = append(pg.Sections, sec)
	}
	return pg
}

// rewardCapNotes is the "capped" notes for a milestone's rewards, one per
// reward a limit holds back, to follow the line that lists them: ""
// when every reward counts in full.
func rewardCapNotes(state game.GameState, rewards []config.Effect, held bool) string {
	var sb strings.Builder
	for _, eff := range rewards {
		if note := capNote(state, eff, held); note != "" {
			sb.WriteString(" (" + note + ")")
		}
	}
	return sb.String()
}
