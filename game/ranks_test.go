package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// giveDomainPoints puts badges worth exactly pts points in a domain on the account.
func giveDomainPoints(t *testing.T, ge *GameEngine, domain string, pts int) {
	t.Helper()
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if ge.account == nil {
		ge.account = &Account{}
	}
	if ge.account.Badges == nil {
		ge.account.Badges = map[string]BadgeEarned{}
	}
	got := 0
	for i := range ge.badges.defs {
		def := &ge.badges.defs[i]
		if got >= pts {
			break
		}
		if def.Domain() == domain && def.Points() > 0 && got+def.Points() <= pts {
			ge.account.Badges[def.Key] = BadgeEarned{}
			got += def.Points()
		}
	}
	ge.account.badgeRev++
	if got < pts {
		t.Fatalf("the ruleset holds only %d points in %s", got, domain)
	}
}

func TestDomainRankRaisesWorkerOutput(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	ge.Buildings.counts["hut"] = 10
	ge.Buildings.counts["gathering_camp"] = 3
	ge.Buildings.unlocked["gathering_camp"] = true
	ge.mu.Unlock()
	ge.RecruitWorker("food", 9)
	if _, err := ge.AssignAll("gathering_camp"); err != nil {
		t.Fatal(err)
	}
	rate := func() float64 {
		ge.mu.Lock()
		defer ge.mu.Unlock()
		ge.recalculateRates()
		return ge.Resources.resources["food"].Breakdown.BuildingRate
	}
	base := rate()
	if base <= 0 {
		t.Fatalf("no food production to measure: %v", base)
	}
	// Needs a rank 2 account: 100 points of food badges.
	giveDomainPoints(t, ge, "food", 100)
	if got := ge.DomainRank("food"); got != 2 {
		t.Fatalf("DomainRank(food) = %d, want 2", got)
	}
	if got := ge.DomainRank("lumber"); got != 0 {
		t.Fatalf("DomainRank(lumber) = %d, want 0", got)
	}
	ge.mu.Lock()
	byWorkers := 0.0
	for _, eff := range ge.Buildings.defs["gathering_camp"].Effects {
		if eff.Type == "production" && eff.Target == "food" {
			byWorkers = eff.Value * 3 * staffedShare
		}
	}
	ge.mu.Unlock()
	want := base + byWorkers*config.RankBonus*2
	if got := rate(); math.Abs(got-want) > 1e-9 {
		t.Errorf("food building rate with rank 2 = %v, want %v (base %v)", got, want, base)
	}
}
