package game

import "slices"

// Gifts that outlast an accepted offer (deals.go). A boost adds dealBoostShare
// of the gross output of one resource until its end tick; the ticks of live
// play and of time away both count against it. Know-how is a flag on the
// civ's state (FactionState.Knowhow) that adds dealKnowhowShare of the civ's
// specialty through the same resolver pool the techs use.

// GiftBoost is one running boost, as saved.
type GiftBoost struct {
	Res string `json:"res"`
	End int    `json:"end"`
}

// addBoost starts a boost on res for ticks, or extends the one it has.
func (ge *GameEngine) addBoost(res string, ticks int) {
	for i := range ge.boosts {
		if ge.boosts[i].Res == res {
			ge.boosts[i].End += ticks
			ge.recalculateRates()
			return
		}
	}
	ge.boosts = append(ge.boosts, GiftBoost{Res: res, End: ge.tick + ticks})
	ge.recalculateRates()
}

// applyGiftBoosts adds each running boost to its resource's rate, after the
// pools and the tech layer. Its own breakdown line says why.
func (ge *GameEngine) applyGiftBoosts() {
	for _, b := range ge.boosts {
		if r := ge.Resources.resources[b.Res]; r != nil && r.Rate > 0 {
			add := float64(r.Rate * dealBoostShare)
			r.Rate += add
			r.Breakdown.GiftRate += add
		}
	}
}

// expireBoosts drops the boosts whose time is up and reports whether any was.
func (ge *GameEngine) expireBoosts() bool {
	n := len(ge.boosts)
	ge.boosts = slices.DeleteFunc(ge.boosts, func(b GiftBoost) bool { return b.End <= ge.tick })
	return len(ge.boosts) != n
}

// boostEndIn is the ticks until the next boost ends (0 with none running).
func (ge *GameEngine) boostEndIn() int {
	next := 0
	for _, b := range ge.boosts {
		if left := b.End - ge.tick; next == 0 || left < next {
			next = left
		}
	}
	return max(next, 0)
}

// knowhowModifiers emits each earned know-how as an addition to its civ's
// specialty pool.
func (ge *GameEngine) knowhowModifiers() []Modifier {
	var out []Modifier
	for _, def := range ge.Diplomacy.factionList {
		if fs := ge.Diplomacy.factions[def.Key]; fs != nil && fs.Knowhow && def.Specialty != "" {
			out = append(out, Modifier{Source: "knowhow:" + def.Key, Target: def.Specialty + "_rate", Op: OpAdd, Value: dealKnowhowShare})
		}
	}
	return out
}
