package smoke

import (
	"fmt"
	"math"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// PrestigePoints is the prestige formula as site/docs/prestige.md documents
// it: base = age index, bonus = floor(milestones/10) + floor(techs/15) +
// floor(built/50), points = floor((base+bonus) / sqrt(level+1)), and at
// least 1 from the Medieval Age (index 5) on.
func PrestigePoints(st game.GameState) int {
	idx := -1
	for i, a := range config.AgeOrder() {
		if a == st.Age {
			idx = i
		}
	}
	if idx < 0 {
		return 0
	}
	raw := idx + st.Milestones.CompletedCount/10 + st.Research.TotalResearched/15 + st.Stats.TotalBuilt/50
	p := int(float64(raw) / math.Sqrt(float64(st.Prestige.Level+1)))
	if p < 1 && idx >= 5 {
		p = 1
	}
	return p
}

// answerLastPassage resolves a pending Last Passage by the configured policy
// and returns the ending: "endured", "succumbed", or "" if the answer failed.
func (r *runner) answerLastPassage(st game.GameState) string {
	if r.cfg.LastPassage == "succumb" && !st.LastPassage.CosmicLegacy {
		if r.bot.act("last_passage", "succumb", r.ge.SuccumbLastPassage()) {
			r.res.Stats.CatastrophesSuccumbed++
			return "succumbed"
		}
		return ""
	}
	if r.bot.act("last_passage", "endure", r.ge.EndureLastPassage()) {
		r.res.Stats.CatastrophesEndured++
		return "endured"
	}
	return ""
}

// checkPrestigeCarry checks what must survive a prestige: Succumb legacy
// bonuses, ruins, the Cosmic Legacy, bought upgrades, and the documented
// passive bonus.
func (r *runner) checkPrestigeCarry(before, after game.GameState, ending string) {
	for _, p := range prestigeCarryProblems(before, after, ending) {
		r.anomaly(KindInvariant, p.check, p.msg, after, true)
	}
}

// problem is one failed check: a stable name and a message.
type problem struct{ check, msg string }

func prestigeCarryProblems(before, after game.GameState, ending string) []problem {
	var out []problem
	for _, ep := range sortedKeys(before.LegacyBonuses) {
		if before.LegacyBonuses[ep] && !after.LegacyBonuses[ep] {
			out = append(out, problem{"prestige_lost_legacy", fmt.Sprintf("the Succumb legacy bonus for %s did not survive prestige", ep)})
		}
	}
	for _, key := range sortedKeys(before.Buildings) {
		bs := before.Buildings[key]
		if bs.RuinCount > 0 && after.Buildings[key].RuinCount < bs.RuinCount {
			out = append(out, problem{"prestige_lost_ruins",
				fmt.Sprintf("%s had %d ruins before prestige and %d after", key, bs.RuinCount, after.Buildings[key].RuinCount)})
		}
	}
	if (before.LastPassage.CosmicLegacy || ending == "succumbed") && !after.LastPassage.CosmicLegacy {
		out = append(out, problem{"prestige_lost_cosmic_legacy", "the Cosmic Legacy did not survive prestige"})
	}
	for _, key := range sortedKeys(before.Prestige.Upgrades) {
		if t := after.Prestige.Upgrades[key].Tier; t < before.Prestige.Upgrades[key].Tier {
			out = append(out, problem{"prestige_lost_upgrade",
				fmt.Sprintf("prestige upgrade %s fell from tier %d to %d", key, before.Prestige.Upgrades[key].Tier, t)})
		}
	}
	if after.Prestige.PassiveBonus != 0 {
		out = append(out, problem{"prestige_passive_bonus",
			fmt.Sprintf("passive bonus is %.4f at level %d; it retired into Era Mastery and must be 0", after.Prestige.PassiveBonus, after.Prestige.Level)})
	}
	return append(out, masteryCarryProblems(before, after)...)
}

// masteryCarryProblems checks Era Mastery across a prestige: every age below
// the run's furthest (the age prestiged from, or deeper if the run went
// deeper) gains one level up to config.MasteryCap, every other age keeps
// its level, and the record never moves back.
func masteryCarryProblems(before, after game.GameState) []problem {
	var out []problem
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	far := order[before.Age]
	if o, ok := order[before.Mastery.RunFurthest]; ok && o > far {
		far = o
	}
	for _, a := range config.AgeOrder() {
		was, now := before.Mastery.Ages[a], after.Mastery.Ages[a]
		want := was
		if order[a] < far {
			want = min(was+1, config.MasteryCap)
		}
		if now != want {
			out = append(out, problem{"prestige_mastery",
				fmt.Sprintf("%s mastery went %d -> %d at a prestige from %s; want %d", a, was, now, before.Age, want)})
		}
	}
	rec := order[before.Age]
	if o, ok := order[before.Mastery.Record]; ok && o > rec {
		rec = o
	}
	if o, ok := order[after.Mastery.Record]; !ok || o < rec {
		out = append(out, problem{"prestige_record",
			fmt.Sprintf("the record is %q after a prestige from %s (record before: %q)", after.Mastery.Record, before.Age, before.Mastery.Record)})
	}
	return out
}
