package ui

import (
	"math"
	"strconv"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// The army's rank as the Expedition panel and the People panel's Army tab
// show it (config/army.go has the numbers).

// xpText is experience as a player reads it: whole, or with a half.
func xpText(x float64) string {
	if x == math.Trunc(x) {
		return strconv.Itoa(int(x))
	}
	return strconv.FormatFloat(x, 'f', 1, 64)
}

// rankName is the age's name for a rank.
func rankName(state game.GameState, rank int) string {
	return state.Term(config.ArmyRankKey(rank)).One
}

// rankProgress is the experience as "42/80" toward the next rank, or just the
// number at the top rank.
func rankProgress(state game.GameState) string {
	mil := state.Military
	if mil.Rank+1 >= config.ArmyRankCount {
		return xpText(mil.Experience)
	}
	return xpText(mil.Experience) + "/" + strconv.Itoa(config.ArmyRankFloor(mil.Rank+1))
}

// rankLine is the header's line: "Army: Veteran · 42/80".
func rankLine(state game.GameState) string {
	return vibe.Line("ui.army_rank.rankLine.01", "Army: ") + rankName(state, state.Military.Rank) + " · " + rankProgress(state)
}

// rankGives is what the army's rank gives, for the Army tab: "odds +4 points,
// hauls +10%, Legendary offers +50%".
func rankGives(rank int) string {
	if rank == 0 {
		return vibe.Line("ui.army_rank.rankGives.01", "nothing yet")
	}
	return vibe.Sprintf("ui.army_rank.rankGives.02", "odds +%d points, hauls +%d%%, Legendary offers +%d%%",
		int(math.Round(config.ArmyOddsBonus(rank)*100)), int(math.Round((config.ArmyHaulFactor(rank)-1)*100)), int(math.Round((config.ArmyRareFactor(rank)-1)*100)))
}

// rankReturnText is what a returned mission did to the army's experience:
// " The army gained 3 experience."
func rankReturnText(delta float64) string {
	switch {
	case delta > 0:
		return vibe.Line("ui.army_rank.rankReturnText.01", " The army gained ") + xpText(delta) + vibe.Line("ui.army_rank.rankReturnText.02", " experience.")
	case delta < 0:
		return vibe.Line("ui.army_rank.rankReturnText.03", " The army lost ") + xpText(-delta) + vibe.Line("ui.army_rank.rankReturnText.04", " experience.")
	}
	return ""
}

// rankNext says what the next tier of the board needs: from Veteran the next
// tier opens an era early.
func rankNext(state game.GameState) string {
	if state.Military.Rank >= config.ArmyEarlyRank {
		return vibe.Line("ui.army_rank.rankNext.01", "Missions of the next tier are offered an era early.")
	}
	return rankName(state, config.ArmyEarlyRank) + vibe.Line("ui.army_rank.rankNext.02", " armies see the next tier of missions an era early.")
}
