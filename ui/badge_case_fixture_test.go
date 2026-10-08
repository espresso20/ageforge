package ui

import (
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// caseFixture is a badge list with every tier and every state in it, as
// the engine would hand it over: earned badges of each tier, a ladder half
// climbed, badges in sight and not earned, withheld badges (their text
// already gone), a secret with its hint, a crossed badge and the three
// integrity badges.
func caseFixture() ([]game.BadgeView, game.BadgeSummary) {
	at := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	tierOf := map[config.BadgeTier]string{}
	for _, t := range []config.BadgeTier{config.BadgeBronze, config.BadgeSilver, config.BadgeGold, config.BadgePlatinum, config.BadgeLegendary} {
		tierOf[t] = t.Name()
	}
	rarity := map[config.BadgeTier]string{
		config.BadgeBronze: "common", config.BadgeSilver: "uncommon", config.BadgeGold: "rare",
		config.BadgePlatinum: "epic", config.BadgeLegendary: "legendary",
	}
	mk := func(key, family, name, desc, emblem string, tier config.BadgeTier) game.BadgeView {
		return game.BadgeView{
			Key: key, Family: family, Name: name, Desc: desc, Emblem: emblem,
			Level: tier, Tier: tierOf[tier], Rarity: rarity[tier], Points: tier.Points(),
		}
	}
	earned := func(v game.BadgeView) game.BadgeView {
		v.Earned, v.At, v.Run = true, at, "Rome"
		return v
	}
	hidden := func(key, family string, tier config.BadgeTier) game.BadgeView {
		return game.BadgeView{Key: key, Family: family, Name: game.BadgeHiddenName, Hidden: true, Level: tier, Tier: tierOf[tier], Rarity: rarity[tier], Points: tier.Points()}
	}
	ladder := func(v game.BadgeView, name string, rung, rungs int, progress, target float64) game.BadgeView {
		v.Ladder, v.Rung, v.Rungs = name, rung, rungs
		if !v.Earned {
			v.Progress, v.Target = progress, target
		}
		return v
	}
	views := []game.BadgeView{
		earned(mk("age.stone_age", "age", "Rock Solid", "Reach the Stone Age.", "centre.0", config.BadgeBronze)),
		earned(mk("age.iron_age", "age", "Age of Iron", "Reach the Iron Age.", "centre.1", config.BadgeBronze)),
		earned(mk("age.industrial_age", "age", "Full Steam", "Reach the Industrial Age.", "centre.2", config.BadgeSilver)),
		mk("age.modern_age", "age", "Into the Modern Age", "Reach the Modern Age.", "centre.4", config.BadgeGold),
		hidden("age.space_age", "age", config.BadgeGold),
		hidden("age.galactic_age", "age", config.BadgePlatinum),

		ladder(earned(mk("lineage.housing.1", "lineage", "Housing Hobbyist", "Build 14 housing buildings across all your runs. Sold and rebuilt copies count once.", "lineage.housing", config.BadgeBronze)), "Housing", 1, 5, 0, 0),
		ladder(earned(mk("lineage.housing.2", "lineage", "Housing Contractor", "Build 57 housing buildings across all your runs. Sold and rebuilt copies count once.", "lineage.housing", config.BadgeSilver)), "Housing", 2, 5, 0, 0),
		ladder(mk("lineage.housing.3", "lineage", "Housing Magnate", "Build 180 housing buildings across all your runs. Sold and rebuilt copies count once.", "lineage.housing", config.BadgeGold), "Housing", 3, 5, 131, 180),
		ladder(mk("lineage.housing.4", "lineage", "Housing Tycoon", "Build 520 housing buildings across all your runs. Sold and rebuilt copies count once.", "lineage.housing", config.BadgePlatinum), "Housing", 4, 5, 131, 520),
		ladder(mk("lineage.housing.5", "lineage", "Housing Dynasty", "Build 1,400 housing buildings across all your runs. Sold and rebuilt copies count once.", "lineage.housing", config.BadgeLegendary), "Housing", 5, 5, 131, 1400),
		ladder(mk("lineage.military.1", "lineage", "Military Hobbyist", "Build 9 military buildings across all your runs. Sold and rebuilt copies count once.", "lineage.military", config.BadgeBronze), "Military", 1, 2, 4, 9),
		ladder(mk("lineage.military.2", "lineage", "Military Contractor", "Build 40 military buildings across all your runs. Sold and rebuilt copies count once.", "lineage.military", config.BadgeSilver), "Military", 2, 2, 4, 40),
		hidden("lineage.hacker.1", "lineage", config.BadgeBronze),
		hidden("lineage.hacker.2", "lineage", config.BadgeSilver),

		ladder(earned(mk("ladder.prestiges.1", "ladder", "First Prestige", "Prestige for the first time.", "star", config.BadgeBronze)), "Prestiges", 1, 4, 0, 0),
		ladder(earned(mk("ladder.prestiges.2", "ladder", "Creature of Habit", "Prestige 3 times.", "star", config.BadgeSilver)), "Prestiges", 2, 4, 0, 0),
		ladder(earned(mk("ladder.prestiges.3", "ladder", "Serial Reincarnator", "Prestige 10 times.", "star", config.BadgeGold)), "Prestiges", 3, 4, 0, 0),
		ladder(earned(mk("ladder.prestiges.4", "ladder", "Eternal Return", "Prestige 25 times.", "star", config.BadgeLegendary)), "Prestiges", 4, 4, 0, 0),

		earned(mk("special.hut_hoarder", "special", "Hut Hoarder", "Have 60 huts standing before you leave the Primitive Age. The 60th costs 18,956 wood, 1,354 times the first.", "hut", config.BadgeGold)),
		earned(mk("special.small_world", "special", "Small World", "Meet every civilization.", "star", config.BadgePlatinum)),
		mk("special.fashionably_late", "special", "Fashionably Late", "Spend ten times the Primitive Age's pacing target in the Primitive Age.", "sun", config.BadgeBronze),
	}
	// Hut Hoarder was earned in a modified game.
	for i := range views {
		if views[i].Key == "special.hut_hoarder" {
			views[i].Rarity = "epic"
		}
		if views[i].Key == "age.industrial_age" {
			views[i].Crossed = true
		}
	}
	secret := hidden("special.liquidation_sale", "special", config.BadgeBronze)
	secret.Secret, secret.Desc = true, "Something about a clearance."
	views = append(views, secret)
	for _, in := range [][3]string{
		{"special.hand_in_the_cookie_jar", "Hand in the Cookie Jar", "cookie_jar"},
		{"special.touched_by_the_source", "Touched by the Source", "source"},
		{"special.creative_accounting", "Creative Accounting", "ledger"},
	} {
		v := earned(game.BadgeView{Key: in[0], Family: "special", Name: in[1], Desc: "An integrity badge.", Emblem: in[2], Secret: true, Integrity: true})
		if in[2] == "source" {
			v.RewardTheme = "source"
		}
		if in[2] == "ledger" {
			v.RewardTheme, v.Crossed = "glitch", true
		}
		views = append(views, v)
	}
	var sum game.BadgeSummary
	for _, v := range views {
		switch {
		case v.Integrity:
		case v.Earned:
			sum.Earned++
			sum.Shown++
			if !v.Crossed {
				sum.Points += v.Points
			}
		case v.Hidden:
			sum.Hidden++
		default:
			sum.Shown++
		}
	}
	sum.Title, sum.TitleRank, sum.NextTitle, sum.NextTitleAt = "Headman", 1, "Magistrate", 1000
	return views, sum
}

// drawCase lays the case out and renders it as the panel does.
func drawCase(views []game.BadgeView, sum game.BadgeSummary, w, h int, v caseView, n int) (*caseModel, caseView, *tGrid) {
	m := buildCase(views, sum, true, v.tab, caseLayoutFor(w, h, v.tab))
	if m.selected(v) == nil {
		v.sel, v.card = m.home(), false
	}
	m.follow(&v)
	g := renderCase(m, v, n)
	if v.plain() {
		foldBadges(g)
	}
	return m, v, g
}
