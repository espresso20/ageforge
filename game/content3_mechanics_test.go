package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestDigitalToCosmicMechanics: the three mechanic numbers the last content
// batch added, each where the game uses it, with its tech and without.
func TestDigitalToCosmicMechanics(t *testing.T) {
	with := func(techs ...string) *GameEngine {
		ge := newSeededEngine(13)
		learn(ge, techs...)
		ge.mu.Lock()
		ge.recalculateRates()
		ge.mu.Unlock()
		return ge
	}
	plain := with()

	// The Global Village and the Federation Charter: every civilization's
	// set of deals holds one more, each. An isolationist's lone deal too.
	slots := func(ge *GameEngine, personality string, tier int) int {
		ge.mu.Lock()
		defer ge.mu.Unlock()
		return ge.newDealEnv().slots(personality, tier)
	}
	village, both := with("global_village"), with("global_village", "federation_charter")
	for _, c := range []struct {
		personality string
		tier        int
	}{{"peaceful", 0}, {"mercantile", 2}, {"aggressive", 0}, {"isolationist", 0}} {
		base := dealSlots(c.personality, c.tier)
		if a, b, d := slots(plain, c.personality, c.tier), slots(village, c.personality, c.tier), slots(both, c.personality, c.tier); a != base || b != base+1 || d != base+2 {
			t.Errorf("a %s civilization at tier %d offers %d deals, %d with the Global Village and %d with the Federation Charter too; want %d, %d and %d",
				c.personality, c.tier, a, b, d, base, base+1, base+2)
		}
	}
	// A rolled set is as long as its slots allow: no civilization offers
	// more than them, and with the two techs the set can be two longer.
	def := plain.Diplomacy.factionDefs["riverlands_tribes"]
	roll := func(ge *GameEngine) int {
		ge.mu.Lock()
		defer ge.mu.Unlock()
		ge.age = "digital_age"
		for res, r := range ge.Resources.resources {
			ge.Resources.UnlockResource(res)
			r.Storage, r.Amount = 1e12, 1e11
		}
		most := 0
		for seed := int64(1); seed <= 40; seed++ {
			most = max(most, len(rollFactionDeals(def, FactionState{Discovered: true, Opinion: 30, Status: "friendly"}, ge.newDealEnv(), newSeededEngine(seed).gameRNG(), 1)))
		}
		return most
	}
	if a, b := roll(plain), roll(both); a > dealSlots(def.Personality, 1) || b <= a || b > dealSlots(def.Personality, 1)+2 {
		t.Errorf("the longest set of deals in 40 rolls is %d, and %d with both techs; want more with the techs, and never more than the slots (%d, and 2 more)",
			a, b, dealSlots(def.Personality, 1))
	}

	// The Global Village: an alliance costs half, in the command and on the
	// Factions panel.
	if a, b := plain.Diplomacy.AllyPrice(), village.Diplomacy.AllyPrice(); a != AllyCost || b != AllyCost/2 {
		t.Errorf("an alliance costs %v, %v with the Global Village; want %v and half", a, b, float64(AllyCost))
	}
	if got := village.GetState().Diplomacy.AllyCost; got != AllyCost/2 {
		t.Errorf("the Factions panel reads %v for an alliance with the Global Village, want %v", got, AllyCost/2)
	}
	learn(village, "envoys")
	village.mu.Lock()
	village.Diplomacy.factions["riverlands_tribes"] = &FactionState{Discovered: true, Opinion: AllyOpinion}
	village.Resources.UnlockResource("gold")
	village.Resources.resources["gold"].Storage = 1e6
	village.Resources.resources["gold"].Amount = 1000
	village.mu.Unlock()
	if err := village.SetDiplomaticStatus("riverlands_tribes", "allied"); err != nil {
		t.Fatal(err)
	}
	if got := village.Resources.Get("gold"); got != 1000-AllyCost/2 {
		t.Errorf("allying with the Global Village left %v gold of 1000, want %v", got, 1000-AllyCost/2)
	}

	// Void Contemplation: an Appease level costs a fifth less, on the
	// Harbinger panel and when it is paid.
	ge := lpEngine(t, "interstellar_age", 5)
	before := ge.GetState().Harbinger.AppeaseCost
	if len(before) != 2 {
		t.Fatalf("the Last Passage's Appease costs %v, want faith and culture", before)
	}
	learn(ge, "void_contemplation")
	after := ge.GetState().Harbinger.AppeaseCost
	for res, v := range before {
		if math.Abs(after[res]-float64(v*0.8)) > 1e-6*v {
			t.Errorf("Appease costs %v %s, %v with Void Contemplation; want a fifth less", v, res, after[res])
		}
	}
	ge.mu.Lock()
	for res, v := range after {
		ge.Resources.UnlockResource(res)
		ge.Resources.resources[res].Storage = 2 * v
		ge.Resources.resources[res].Amount = v
	}
	ge.mu.Unlock()
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatalf("appeasing at the cut price, holding exactly it: %v", err)
	}
	for res := range after {
		if got := ge.Resources.Get(res); got > 1e-6*after[res] {
			t.Errorf("after paying, %v %s is left: the price paid was not the price shown", got, res)
		}
	}
	if def := config.MechanicByKey()[config.MechanicDealSlots]; def.EffectText(1) != "every civilization offers 1 more deal" {
		t.Errorf("the effect reads %q", def.EffectText(1))
	}
}
