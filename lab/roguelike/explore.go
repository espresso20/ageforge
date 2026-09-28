package main

import "github.com/espresso20/ageforge/game"

// explore does what the smoke bot skips and a real player does: sends
// scouts out (lifting the fog, meeting civs) and tries to keep trade routes
// running, so the map has expeditions and caravans to show.
//
// note: route starts mostly fail in practice. Every route needs a market,
// port or harbor, markets upgrade into trading posts on age-up, and the bot
// never saves up for a fresh one. The map draws only real routes.
func explore(ge *game.GameEngine, st game.GameState) {
	if st.Military.ActiveScout == nil {
		for _, e := range st.Military.Expeditions {
			if e.Category == "scouting" && e.CanLaunch {
				_ = ge.LaunchExpedition(e.Key)
				break
			}
		}
	}
	if st.Military.ActiveMilitary == nil && st.Military.SoldierCount > 40 {
		for _, e := range st.Military.Expeditions {
			if e.Category == "military" && e.CanLaunch {
				_ = ge.LaunchExpedition(e.Key)
				break
			}
		}
	}
	if len(st.Trade.ActiveRoutes) < 3 {
		for _, r := range st.Trade.AvailableRoutes {
			if r.CanStart {
				_ = ge.StartTradeRoute(r.Key)
				break
			}
			if b, ok := st.Buildings[r.RequiredBld]; ok && b.Count < r.MinCount && b.CanBuild {
				_ = ge.BuildBuilding(r.RequiredBld)
				break
			}
		}
	}
}
