package game

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// bareMissionTown is a seeded town in age with an income and stock to pay any
// mission, and no tech researched: the tiers of the board must open without
// one.
func bareMissionTown(t *testing.T, seed int64, age string) *GameEngine {
	t.Helper()
	ge := missionReady(t, seed, age)
	ge.mu.Lock()
	ge.Research.researched = map[string]bool{}
	ge.Research.rebuildBonuses()
	ge.mu.Unlock()
	return ge
}

// setExperience sets the army's experience and lets the board follow it.
func setExperience(ge *GameEngine, xp float64) {
	ge.mu.Lock()
	ge.Military.experience = xp
	ge.mu.Unlock()
	ge.EnsureMissionBoard()
}

func eraOrderOf(ge *GameEngine, age string) int {
	era, _ := ge.rules.Era(ge.rules.EraOf(age))
	return era.Order
}

// TestBoardTiersByEra: the Stone Era's board holds short offers only, the Iron
// Era's short and long, the Steel Era's and every later era's all three, three
// offers to a tier, in every age; and every offer can be sent with no tech
// researched.
func TestBoardTiersByEra(t *testing.T) {
	for _, age := range ageKeys() {
		ge := bareMissionTown(t, 5, age)
		ge.EnsureMissionBoard()
		tiers := min(3, eraOrderOf(ge, age)+1)
		slots := ge.Military.board.Slots
		if len(slots) != 3*tiers {
			t.Fatalf("%s: %d offers, want %d", age, len(slots), 3*tiers)
		}
		for i, slot := range slots {
			want := []string{TierShort, TierLong, TierGrand}[i/3]
			if slot.Offer == nil || slot.Offer.Tier != want {
				t.Fatalf("%s: slot %d is %+v, want a %s offer", age, i+1, slot.Offer, want)
			}
		}
		state := ge.GetState().Military
		if len(state.Board) != 3*tiers || len(state.Expeditions) != 3*tiers {
			t.Errorf("%s: the state shows %d slots and %d offers, want %d", age, len(state.Board), len(state.Expeditions), 3*tiers)
		}
		for _, o := range state.Expeditions {
			if o.LaunchBlockReason != "" && strings.Contains(o.LaunchBlockReason, "first") {
				t.Errorf("%s: offer %d is blocked for a tech: %q", age, o.Number, o.LaunchBlockReason)
			}
		}
		for n := 1; n <= len(slots); n++ {
			clearArmy(ge) // the last party is home, so its soldiers are too
			stockFor(ge, n)
			if err := ge.LaunchExpedition(n); err != nil {
				t.Errorf("%s: offer %d with no tech researched: %v", age, n, err)
			}
		}
	}
}

// TestBoardTiersFollowTheRank: a Veteran army sees the next tier an era early
// (the Stone Era gets long, the Iron Era grand), a Seasoned one does not, and a
// tier that closes again takes its offers with it while the tiers still open
// keep theirs.
func TestBoardTiersFollowTheRank(t *testing.T) {
	for _, c := range []struct {
		age                   string
		green, seasoned, vets int
	}{{"stone_age", 3, 3, 6}, {"bronze_age", 3, 3, 6}, {"iron_age", 6, 6, 9}, {"medieval_age", 6, 6, 9}, {"renaissance_age", 9, 9, 9}} {
		ge := bareMissionTown(t, 8, c.age)
		setExperience(ge, 0)
		if got := len(ge.Military.board.Slots); got != c.green {
			t.Fatalf("%s, Green: %d offers, want %d", c.age, got, c.green)
		}
		kept := ge.Military.board.clone()
		setExperience(ge, 29.5)
		if got := len(ge.Military.board.Slots); got != c.seasoned {
			t.Errorf("%s, Seasoned: %d offers, want %d", c.age, got, c.seasoned)
		}
		setExperience(ge, 30)
		if got := len(ge.Military.board.Slots); got != c.vets {
			t.Errorf("%s, Veteran: %d offers, want %d", c.age, got, c.vets)
		}
		for i := range kept.Slots {
			if *kept.Slots[i].Offer != *ge.Military.board.Slots[i].Offer {
				t.Errorf("%s: offer %d changed when a tier opened", c.age, i+1)
			}
		}
		setExperience(ge, 0)
		if got := len(ge.Military.board.Slots); got != c.green {
			t.Errorf("%s, Green again: %d offers, want %d", c.age, got, c.green)
		}
	}
}

// comeHome sends a mission of tier out with a certain end and brings it home.
func comeHome(t *testing.T, ge *GameEngine, tier, grand string, success bool) {
	t.Helper()
	clearArmy(ge)
	n := plantOffer(ge, tier, grand)
	stockFor(ge, n)
	if err := ge.LaunchExpedition(n); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	a := ge.Military.ActiveByTier(tier)
	a.TicksLeft = 1
	if success {
		a.Luck, a.SuccessChance = 0, 1
	} else {
		a.Luck, a.SuccessChance = 1, 0
	}
	ge.tick++
	ge.processExpeditions()
	ge.mu.Unlock()
}

// TestExperienceFromMissions: a success gives short 1, long 3 and grand 8; a
// failure takes half of that, never below 0; the return log carries the
// change.
func TestExperienceFromMissions(t *testing.T) {
	ge := bareMissionTown(t, 3, "renaissance_age")
	xp := func() float64 { return ge.Military.Experience() }
	steps := []struct {
		tier, grand string
		success     bool
		want        float64
	}{
		{TierShort, "", false, 0}, // already at the floor
		{TierShort, "", true, 1},
		{TierLong, "", true, 4},
		{TierGrand, GrandMilitary, true, 12},
		{TierGrand, GrandCivil, true, 20},
		{TierShort, "", false, 19.5},
		{TierLong, "", false, 18},
		{TierGrand, GrandMilitary, false, 14},
	}
	for i, s := range steps {
		comeHome(t, ge, s.tier, s.grand, s.success)
		if xp() != s.want {
			t.Fatalf("step %d (%s, success %v): experience %v, want %v", i+1, s.tier, s.success, xp(), s.want)
		}
		last := ge.Military.returns[0]
		wantChange := s.want
		if i > 0 {
			wantChange = s.want - steps[i-1].want
		}
		if last.Experience != wantChange {
			t.Errorf("step %d: the return log says %v experience, want %v", i+1, last.Experience, wantChange)
		}
	}
	if got := ge.GetState().Military.Experience; got != 14 {
		t.Errorf("the state shows %v experience, want 14", got)
	}
}

// TestRankUpIsTold: crossing a rank's threshold logs one line, naming the rank.
func TestRankUpIsTold(t *testing.T) {
	ge := bareMissionTown(t, 4, "bronze_age")
	setExperience(ge, 9)
	if logHas(ge, "Your army is now") {
		t.Fatal("the log already names a rank")
	}
	comeHome(t, ge, TierShort, "", true)
	if !logHas(ge, "Your army is now Seasoned.") {
		t.Errorf("no rank-up line at 10 experience; the log ends %q", lastLog(ge))
	}
	if ge.Military.Rank() != 1 {
		t.Errorf("rank %d at 10 experience, want 1", ge.Military.Rank())
	}
	// A loss below the line is not told: only a rank up is.
	n := 0
	for _, l := range ge.log {
		if strings.Contains(l.Message, "Your army is now") {
			n++
		}
	}
	comeHome(t, ge, TierShort, "", false)
	for _, l := range ge.log {
		if strings.Contains(l.Message, "Your army is now") {
			n--
		}
	}
	if n != 0 {
		t.Errorf("a loss told a rank (%d lines more)", -n)
	}
}

// TestRankEffects: each rank adds two points of odds (to at most 95%), 5% of
// haul and 25% of the Legendary offer chance.
func TestRankEffects(t *testing.T) {
	ge := bareMissionTown(t, 6, "renaissance_age")
	offer := testOffer(ge, TierLong, "")
	offer.SuccessChance = .5
	high := offer
	high.SuccessChance = .94
	setExperience(ge, 0)
	ge.mu.RLock()
	baseHaul := ge.missionHaulLocked(10)
	ge.mu.RUnlock()
	if len(baseHaul) == 0 {
		t.Fatal("no haul to compare")
	}
	for rank, floor := range []float64{0, 10, 30, 80, 200} {
		setExperience(ge, floor)
		ge.mu.RLock()
		odds, capped, haul := ge.offerSuccessLocked(offer), ge.offerSuccessLocked(high), ge.missionHaulLocked(10)
		ge.mu.RUnlock()
		if want := .5 + .02*float64(rank); !near(odds, want) {
			t.Errorf("rank %d: odds %v, want %v", rank, odds, want)
		}
		if capped > .95+1e-9 {
			t.Errorf("rank %d: odds %v pass the 95%% cap", rank, capped)
		}
		for res, base := range baseHaul {
			if want := base * (1 + .05*float64(rank)); !near(haul[res], want) {
				t.Errorf("rank %d: %s haul %v, want %v", rank, res, haul[res], want)
			}
		}
	}
	// The Legendary chance: the same roll, seeded, passes at chance x factor.
	for rank := 0; rank < 5; rank++ {
		factor := config.ArmyRareFactor(rank)
		hits := 0
		for seed := int64(1); seed <= 4000; seed++ {
			roll := rand.New(rand.NewSource(seed)).Float64()
			b := MissionBoard{Made: 100, SinceRare: 20}
			got := b.rollRare(rand.New(rand.NewSource(seed)), factor)
			if want := roll < config.RareOfferChance*factor; got != want {
				t.Fatalf("rank %d seed %d: rare=%v, roll %v against %v", rank, seed, got, roll, config.RareOfferChance*factor)
			}
			if got {
				hits++
			}
		}
		if rank == 4 && hits <= 4000*3/100 {
			t.Errorf("a Legendary army saw %d rare offers in 4000 rolls, want about twice the usual", hits)
		}
	}
}

// TestExperienceIsTheRuns: it is saved and loaded, and a prestige starts the
// army Green again.
func TestExperienceIsTheRuns(t *testing.T) {
	defer SetDataDirForTest(t.TempDir())()
	ge := bareMissionTown(t, 9, "bronze_age")
	setExperience(ge, 42.5)
	again := saveAndLoad(t, ge, "army_experience")
	if got := again.Military.Experience(); got != 42.5 {
		t.Errorf("a loaded game holds %v experience, want 42.5", got)
	}
	if got := again.GetState().Military.Rank; got != 2 {
		t.Errorf("a loaded game's army is rank %d, want 2", got)
	}

	// A prestige makes a new military manager, and the experience goes with
	// the rest of the run.
	pr := newSeededEngine(2)
	for _, a := range pr.Rules().AgeKeys()[1:] {
		if err := pr.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
		_ = pr.ForceQuietFateForTest(pr.Rules().EraOf(a))
		if a == PrestigeRunAge {
			break
		}
	}
	pr.mu.Lock()
	pr.Military.experience = 120
	pr.mu.Unlock()
	_ = pr.SummonHarbingerForTest(PrestigeRunAge)
	if err := pr.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if got := pr.Military.Experience(); got != 0 {
		t.Errorf("after a prestige the army holds %v experience, want 0", got)
	}
}
