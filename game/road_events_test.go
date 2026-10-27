package game

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

func TestRoadEventRollsAreSeededAndBounded(t *testing.T) {
	o := MissionOffer{Tier: TierLong, LengthMinutes: 120}
	w, _ := config.MissionWordsFor("bronze_age")
	a := rollRoadEvents(rand.New(rand.NewSource(71)), o, 120, w)
	b := rollRoadEvents(rand.New(rand.NewSource(71)), o, 120, w)
	if len(a) != len(b) {
		t.Fatalf("event counts differ: %d and %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("event %d differs: %+v / %+v", i, a[i], b[i])
		}
		if a[i].Tick < 12 || a[i].Tick > 108 {
			t.Fatalf("event tick outside mission bounds: %+v", a[i])
		}
	}
}

func TestRoadEventPoissonMean(t *testing.T) {
	for _, tc := range []struct {
		tier         string
		length, mean float64
	}{{TierLong, 240, 6}, {TierShort, 240, 3}} {
		total := 0
		for seed := int64(0); seed < 10000; seed++ {
			w, _ := config.MissionWordsFor("bronze_age")
			total += len(rollRoadEvents(rand.New(rand.NewSource(seed)), MissionOffer{Tier: tc.tier, LengthMinutes: tc.length}, 240, w))
		}
		got := float64(total) / 10000
		if got < tc.mean-.15 || got > tc.mean+.15 {
			t.Errorf("%s mean %.3f, want near %.1f", tc.tier, got, tc.mean)
		}
	}
}

func TestRoadEventEffectsAndSaveCopy(t *testing.T) {
	ge := NewGameEngine()
	w, ok := ge.rules.MissionWords("bronze_age")
	if !ok {
		t.Fatal("missing Bronze Age words")
	}
	a := &ActiveExpedition{MissionOffer: MissionOffer{Age: "bronze_age", Place: "Tin Road", SuccessChance: .50, HaulMinutes: 20}, TotalTicks: 100, TicksLeft: 50, BaseRewards: map[string]float64{"wood": 10}, SuccessRewards: map[string]float64{"wood": 10}, FailureRewards: map[string]float64{"wood": 3}}
	storm := RoadEvent{Kind: "storm", Story: w.RoadTemplate("storm", 0)}
	ge.applyRoadEvent(a, &storm)
	if a.TicksLeft != 65 {
		t.Fatalf("storm left %d ticks, want 65", a.TicksLeft)
	}
	find := RoadEvent{Kind: "find", Story: w.RoadTemplate("find", 0)}
	ge.applyRoadEvent(a, &find)
	if a.SuccessRewards["wood"] <= 10 {
		t.Fatalf("find did not grow haul: %v", a.SuccessRewards)
	}
	ge.Resources.Remove("gold", ge.Resources.Get("gold"))
	setIncome(ge, "gold", .1)
	toll := RoadEvent{Kind: "toll", Story: w.RoadTemplate("toll", 0)}
	story := ge.applyRoadEvent(a, &toll)
	if toll.Kind != "foes" || a.SuccessChance >= .50 || strings.Contains(story, "<") {
		t.Fatalf("unpaid toll did not become foes: kind=%s odds=%f story=%q", toll.Kind, a.SuccessChance, story)
	}
	a.RoadEvents = []RoadEvent{{Tick: 4, Kind: "fair", Story: "fair", Fired: true}}
	a.RoadEventsSeen = []string{"Fair skies."}
	ge.Military.setActive(TierLong, a)
	saved := ge.Military.ActiveForSave()
	if len(saved) != 1 || len(saved[0].Mission.RoadEvents) != 1 || len(saved[0].Mission.RoadEventsSeen) != 1 {
		t.Fatal("road events missing from save copy")
	}
	loaded := NewMilitaryManagerWith(ge.rules)
	loaded.LoadState(MilitarySave{Active: saved})
	if got := loaded.ActiveByTier(TierLong); got == nil || !got.RoadEvents[0].Fired || got.RoadEventsSeen[0] != "Fair skies." {
		t.Fatal("road event state did not survive save/load")
	}
}

func TestRoadEventOddsLossesAndRuins(t *testing.T) {
	ge := NewGameEngine()
	w, _ := ge.rules.MissionWords("bronze_age")
	a := &ActiveExpedition{MissionOffer: MissionOffer{Age: "bronze_age", Place: "Tin Road", SuccessChance: .50}, TicksLeft: 40, TotalTicks: 100,
		SuccessRewards: map[string]float64{"wood": 100}, FailureRewards: map[string]float64{"wood": 30}}
	ge.applyRoadEvent(a, &RoadEvent{Kind: "helpers"})
	if a.TicksLeft != 30 {
		t.Fatalf("helpers left %d ticks, want 30", a.TicksLeft)
	}
	ge.applyRoadEvent(a, &RoadEvent{Kind: "foes"})
	if a.SuccessChance != .42 {
		t.Fatalf("foes odds %.2f, want .42", a.SuccessChance)
	}
	ge.applyRoadEvent(a, &RoadEvent{Kind: "fair"})
	if a.SuccessChance != .48 {
		t.Fatalf("fair odds %.2f, want .48", a.SuccessChance)
	}
	ge.applyRoadEvent(a, &RoadEvent{Kind: "ambush"})
	if a.SuccessRewards["wood"] != 85 || a.FailureRewards["wood"] != 25.5 {
		t.Fatalf("nonmilitary ambush haul: success=%v failure=%v", a.SuccessRewards, a.FailureRewards)
	}
	military := &ActiveExpedition{MissionOffer: MissionOffer{Age: "bronze_age", Tier: TierGrand, Grand: GrandMilitary}, Soldiers: 10}
	ge.applyRoadEvent(military, &RoadEvent{Kind: "ambush", Story: w.RoadTemplate("ambush", 0)})
	if military.Soldiers != 9 {
		t.Fatalf("military ambush left %d soldiers, want 9", military.Soldiers)
	}
	ge.Military.board = MissionBoard{Age: "bronze_age", Slots: make([]boardSlot, boardSizeFor(3))}
	ge.applyRoadEvent(a, &RoadEvent{Kind: "ruins", Story: w.RoadTemplate("ruins", 0)})
	if offer := ge.Military.board.Slots[0].Offer; offer == nil || offer.Tier != TierShort || !strings.Contains(offer.Description, "Found by this mission") {
		t.Fatalf("ruins did not add a found short offer: %+v", offer)
	}
}

// applyRoadEvent carries a road event out and returns its story as read now.
func (ge *GameEngine) applyRoadEvent(a *ActiveExpedition, e *RoadEvent) string {
	words, _ := ge.rules.MissionWords(a.Age)
	return tellRoadStory(words, a.roadPlace(), *e, ge.roadEventTold(a, e))
}
