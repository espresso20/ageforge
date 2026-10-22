package smoke

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// refHours is the ordinary player's hours in each age, with the model's
// upgrades on or off.
func refHours(upgrades bool) []float64 {
	t := newRefTables()
	t.upgrades = upgrades
	refIncome := make([]float64, len(t.ages))
	for i := range t.ages {
		refIncome[i] = units(t.state(t.refTown(i), i, 1).income, t.levels[i])
	}
	return t.walk("ordinary", 1, refIncome, false).Hours
}

// TestReferenceUpgradesShortenTheGame: buying older copies up at half price is
// cheaper than buying new, so the model's players finish sooner with upgrades
// on, and an age with nothing older to upgrade is the same either way.
func TestReferenceUpgradesShortenTheGame(t *testing.T) {
	with, without := refHours(true), refHours(false)
	if with[0] != without[0] {
		t.Errorf("the Primitive Age has nothing to upgrade from: %g hours with upgrades, %g without", with[0], without[0])
	}
	sumWith, sumWithout := 0.0, 0.0
	for i := range with {
		sumWith += with[i]
		sumWithout += without[i]
	}
	if sumWith >= sumWithout {
		t.Errorf("upgrading took %g hours over the 22 ages, buying new %g: upgrades should shorten the game", sumWith, sumWithout)
	}
}

// TestReferenceUpgradePlan: the plan is the game's own. In an age, a rung is
// upgraded into from the older rungs of its line (config.UpgradeTarget), as
// many copies as the reference count asks or as are held, whichever is fewer,
// oldest first; storage and knowledge buildings are in no line.
func TestReferenceUpgradePlan(t *testing.T) {
	tb := newRefTables()
	i := tb.idx["iron_age"]
	start := tb.refTown(i - 1)
	ref := map[string]int{}
	for _, d := range tb.byAge[i] {
		ref[d.Key] = tb.refCount(d)
	}
	older, nUp := tb.upgradePlan(i, start, ref)
	if len(nUp) == 0 {
		t.Fatal("no rung of the Iron Age is upgraded into from the reference town of the age before")
	}
	for key, n := range nUp {
		d := tb.defs[key]
		if !config.InLine(d) || config.IsKnowledgeBuilding(d) || d.Category == "storage" {
			t.Errorf("%s is upgraded into and is in no line", key)
		}
		held, prev := 0, -1
		for _, od := range older[key] {
			if tg, ok := config.UpgradeTarget(tb.rungs[d.Line()], tb.pos, od, "iron_age"); !ok || tg.Key != key {
				t.Errorf("%s is not an older building of %s", od.Key, key)
			}
			if od.LineageTier < prev {
				t.Errorf("%s: older rungs are not oldest first", key)
			}
			prev = od.LineageTier
			held += start[od.Key]
		}
		if want := min(ref[key], held); n != want {
			t.Errorf("%s: %d upgrades, want min(reference count %d, %d older copies held) = %d", key, n, ref[key], held, want)
		}
	}
	tb.upgrades = false
	if o, n := tb.upgradePlan(i, start, ref); len(o) != 0 || len(n) != 0 {
		t.Errorf("with upgrades off the plan is %v, %v, want none", o, n)
	}
}

// TestUpgradePriceIsTheGames: half of the new copy's price, a whole number,
// never nothing.
func TestUpgradePriceIsTheGames(t *testing.T) {
	got := upgradePrice(map[string]float64{"wood": 101, "stone": 1})
	if got["wood"] != 50 || got["stone"] != 1 {
		t.Errorf("upgradePrice = %v, want wood 50 (floor of half) and stone 1 (never nothing)", got)
	}
}
