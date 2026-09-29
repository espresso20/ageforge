package game

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// testDealEnv is a dealEnv for age with every resource the ages up to it
// unlock, roomy stores, and each store a different amount full (so the
// fullest-first asks have a fixed order).
func testDealEnv(age string) dealEnv {
	order := config.AgeOrder()
	ages := config.AgeByKey()
	env := dealEnv{age: age, unlocked: map[string]bool{}, amount: map[string]float64{}, storage: map[string]float64{}}
	for i, a := range order {
		for _, r := range ages[a].UnlockResources {
			env.unlocked[r] = true
		}
		if a == age {
			if i+1 < len(order) {
				env.next = order[i+1]
			}
			break
		}
	}
	for i, r := range config.BaseResources() {
		env.storage[r.Key] = 1e21
		env.amount[r.Key] = float64(i+1) * 1e18
	}
	return env
}

func factionDef(t *testing.T, key string) config.FactionDef {
	t.Helper()
	def, ok := config.FactionByKey()[key]
	if !ok {
		t.Fatalf("no civ %q", key)
	}
	return def
}

// dealAge is the first age at which civ def's specialty can be traded: its
// MinAge, or later when the specialty is locked there.
func dealAge(def config.FactionDef) string {
	order := config.AgeOrder()
	for i, a := range order {
		if a != def.MinAge {
			continue
		}
		for _, b := range order[i:] {
			if testDealEnv(b).unlocked[def.Specialty] {
				return b
			}
		}
	}
	return def.MinAge
}

var (
	neutralState = FactionState{Discovered: true, Status: "neutral"}
	alliedState  = FactionState{Discovered: true, Status: "allied", Opinion: 60}
)

// TestDeals_PerPersonality rolls every civ at the age its specialty first
// trades, neutral and allied, and checks what each personality promises.
func TestDeals_PerPersonality(t *testing.T) {
	for _, def := range config.BaseFactions() {
		age := dealAge(def)
		env := testDealEnv(age)
		for _, fs := range []FactionState{neutralState, alliedState} {
			ds := rollFactionDeals(def, fs, env, testRNG(), 1)
			tier := dealTier(fs)
			if len(ds) == 0 || len(ds) > dealSlots(def.Personality, tier) {
				t.Errorf("%s %s in %s: %d deals, want 1..%d", def.Key, fs.Status, age, len(ds), dealSlots(def.Personality, tier))
			}
			ids := map[int]bool{}
			for _, d := range ds {
				if ids[d.ID] {
					t.Errorf("%s: duplicate deal ID %d", def.Key, d.ID)
				}
				ids[d.ID] = true
				checkDeal(t, def, fs, env, d)
			}
		}
	}

	// Slot counts: mercantile > peaceful > aggressive; isolationists offer
	// one, two when allied; standing adds slots.
	for _, c := range []struct {
		pers string
		tier int
		want int
	}{{"peaceful", 0, 2}, {"peaceful", 1, 3}, {"peaceful", 2, 4}, {"mercantile", 0, 3}, {"mercantile", 2, 5},
		{"aggressive", 0, 1}, {"aggressive", 2, 3}, {"isolationist", 0, 1}, {"isolationist", 1, 1}, {"isolationist", 2, 2}} {
		if got := dealSlots(c.pers, c.tier); got != c.want {
			t.Errorf("dealSlots(%s, %d) = %d, want %d", c.pers, c.tier, got, c.want)
		}
	}
	// Rates rise with standing, mercantile pays best, aggressive worst, and
	// every one beats the market (0.8 of parity) without passing parity.
	for _, pers := range []string{"peaceful", "mercantile", "aggressive", "isolationist"} {
		for _, kind := range []string{DealSell, DealWant} {
			prev := 0.0
			for tier := 0; tier <= 2; tier++ {
				m := dealMult(pers, tier, kind)
				if m <= 1-config.ExchangeFee || m > 1 || m < prev {
					t.Errorf("dealMult(%s, %d, %s) = %v (previous tier %v)", pers, tier, kind, m, prev)
				}
				prev = m
			}
		}
	}
	if !(dealMult("mercantile", 0, DealSell) > dealMult("peaceful", 0, DealSell) && dealMult("peaceful", 0, DealSell) > dealMult("aggressive", 0, DealSell)) {
		t.Error("mercantile should pay best and aggressive worst")
	}
	// Lots grow with standing and strength; mercantile lots are bigger.
	merchant := factionDef(t, "merchant_guild")
	if !(dealLot(merchant, 2, DealSell, 0.5) > dealLot(merchant, 1, DealSell, 0.5) && dealLot(merchant, 1, DealSell, 0.5) > dealLot(merchant, 0, DealSell, 0.5)) {
		t.Error("lots should grow with standing")
	}
	weak, strong := factionDef(t, "riverlands_tribes"), factionDef(t, "ironhold_clans")
	if dealLot(strong, 0, DealSell, 0.5) <= dealLot(weak, 0, DealSell, 0.5) {
		t.Error("lots should grow with strength")
	}

	// The isolationist Directorate offers the next age's goods: oil, which
	// the Atomic Age prices nowhere and the Modern Age builds with.
	dir := factionDef(t, "atomic_directorate")
	ds := rollFactionDeals(dir, neutralState, testDealEnv("atomic_age"), testRNG(), 1)
	if len(ds) != 1 || ds[0].Kind != DealRare || ds[0].Get != "oil" {
		t.Fatalf("Directorate in the Atomic Age offers %+v, want one rare oil deal", ds)
	}

	// A mercantile civ offers more than an aggressive one at the same standing.
	guild := rollFactionDeals(merchant, neutralState, testDealEnv(dealAge(merchant)), testRNG(), 1)
	clans := rollFactionDeals(strong, neutralState, testDealEnv(dealAge(strong)), testRNG(), 1)
	if len(guild) <= len(clans) {
		t.Errorf("Merchant Guild offers %d deals, Ironhold Clans %d; mercantile should offer more", len(guild), len(clans))
	}
}

// checkDeal asserts one deal's terms: its kind fits, amounts are positive and
// fit the stores, a goods deal beats the market where the market trades the
// pair and never beats parity, a favor pays standing.
func checkDeal(t *testing.T, def config.FactionDef, fs FactionState, env dealEnv, d FactionDeal) {
	t.Helper()
	if !(d.GiveAmt >= 1) || d.GiveAmt > env.storage[d.Give] || !env.unlocked[d.Give] {
		t.Errorf("%s %+v: bad price", def.Key, d)
	}
	switch d.Kind {
	case DealFavor:
		if d.Get != "" || d.Standing != dealFavorStanding || fs.Opinion >= dealStandingCap {
			t.Errorf("%s %+v: bad favor", def.Key, d)
		}
		return
	case DealRare:
		if def.Personality != "isolationist" || config.DealPriceLevel(d.Get, env.age) > 0 {
			t.Errorf("%s %+v: rare deal from a non-isolationist or for goods this age prices", def.Key, d)
		}
		return
	case DealSell, DealWant:
		if d.Get != def.Specialty {
			t.Errorf("%s %+v: goods deal not in the specialty %s", def.Key, d, def.Specialty)
		}
	default:
		t.Errorf("%s: unknown kind %q", def.Key, d.Kind)
		return
	}
	if !(d.GetAmt >= 1) || d.GetAmt > dealGetShare*env.storage[d.Get]*1.001 {
		t.Errorf("%s %+v: bad goods amount", def.Key, d)
	}
	rate := d.GetAmt / d.GiveAmt
	if m, ok := config.MarketOffers(d.Give, d.Get, env.age); ok {
		if rate <= m {
			t.Errorf("%s %+v: rate %v does not beat the market's %v", def.Key, d, rate, m)
		}
		if edge := rate/m - 1; edge < 0.04 || edge > 0.26 {
			t.Errorf("%s %+v: edge %.3f outside 5-25%%", def.Key, d, edge)
		}
	}
	lf, lt := config.DealPriceLevel(d.Give, env.age), config.DealPriceLevel(d.Get, env.age)
	if _, market := config.MarketOffers(d.Give, d.Get, env.age); !market && lf > 0 && lt > 0 && rate > lt/lf*1.001 {
		t.Errorf("%s %+v: rate %v beats parity %v", def.Key, d, rate, lt/lf)
	}
}

// TestDeals_NoneWhenHostile: at war, embargoed, rival or hostile civs offer
// nothing, draw nothing, and refuse to trade.
func TestDeals_NoneWhenHostile(t *testing.T) {
	def := factionDef(t, "riverlands_tribes")
	env := testDealEnv("bronze_age")
	for name, fs := range map[string]FactionState{
		"war":     {Discovered: true, Status: "neutral", AtWar: true},
		"embargo": {Discovered: true, Status: "embargo"},
		"rival":   {Discovered: true, Status: "rival"},
		"hostile": {Discovered: true, Status: "neutral", Opinion: dealHostileOpinion},
	} {
		rng := testRNG()
		if ds := rollFactionDeals(def, fs, env, rng, 1); ds != nil {
			t.Errorf("%s: offered %v", name, ds)
		}
		if rng.Int63() != testRNG().Int63() {
			t.Errorf("%s: a blocked civ drew from the rng", name)
		}
		if dealBlocked(fs) == "" {
			t.Errorf("%s: not blocked", name)
		}
	}
	if dealBlocked(FactionState{Discovered: true, Status: "neutral", Opinion: dealHostileOpinion + 1}) != "" {
		t.Error("opinion just above the hostile line should trade")
	}

	// In the engine: offers rolled while friendly are hidden and refused
	// once the civ is embargoed, and neither the panel nor accept shows them.
	ge := dealEngine(t, "bronze_age")
	meet(ge, "riverlands_tribes", 0)
	if n := len(ge.GetState().Diplomacy.Factions["riverlands_tribes"].Deals); n == 0 {
		t.Fatal("no offers while neutral")
	}
	if err := ge.SetDiplomaticStatus("riverlands_tribes", "embargo"); err != nil {
		t.Fatal(err)
	}
	f := ge.GetState().Diplomacy.Factions["riverlands_tribes"]
	if len(f.Deals) != 0 || f.DealsBlocked == "" {
		t.Errorf("embargoed civ still shows %d deals (blocked %q)", len(f.Deals), f.DealsBlocked)
	}
	before := ge.GetState().Resources
	if _, err := ge.AcceptFactionDeal("riverlands_tribes", 1); err == nil || !strings.Contains(err.Error(), "embargo") {
		t.Errorf("accept under embargo: %v", err)
	}
	if !reflect.DeepEqual(before, ge.GetState().Resources) {
		t.Error("a refused deal changed resources")
	}
	// At war, the next roll is empty.
	ge.mu.Lock()
	ge.Diplomacy.factions["riverlands_tribes"].AtWar = true
	ge.Diplomacy.factions["riverlands_tribes"].DealTicks = dealRefreshTicks
	ge.tickFactionDeals()
	n := len(ge.Diplomacy.factions["riverlands_tribes"].Deals)
	ge.mu.Unlock()
	if n != 0 {
		t.Errorf("a civ at war rolled %d deals", n)
	}
}

// dealEngine is a seeded engine in age with every earlier unlock and roomy,
// half-full stores.
func dealEngine(t *testing.T, age string) *GameEngine {
	t.Helper()
	ge := newSeededEngine(11)
	for _, a := range config.AgeOrder() {
		ge.applyAgeUnlocks(a)
		if a == age {
			break
		}
	}
	ge.advanceAge(age)
	ge.pendingCatastrophe = ""
	for _, r := range ge.Resources.resources {
		r.Storage = 1e15
		r.Amount = 5e14
	}
	return ge
}

// meet discovers civ key at opinion and rolls its offers.
func meet(ge *GameEngine, key string, opinion int) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.Diplomacy.DiscoverFaction(key)
	ge.Diplomacy.factions[key].Opinion = opinion
	ge.tickFactionDeals()
}

// TestDeals_AcceptOnce: taking a deal pays its price and delivers its goods
// or standing exactly once.
func TestDeals_AcceptOnce(t *testing.T) {
	ge := dealEngine(t, "colonial_age")
	meet(ge, "merchant_guild", 0)
	ge.mu.RLock()
	d := ge.Diplomacy.factions["merchant_guild"].Deals[0]
	give0, get0 := ge.Resources.Get(d.Give), ge.Resources.Get(d.Get)
	ge.mu.RUnlock()

	got, err := ge.AcceptFactionDeal("merchant_guild", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != d.ID || !got.Taken {
		t.Errorf("accepted %+v, want deal %d taken", got, d.ID)
	}
	ge.mu.RLock()
	fs := ge.Diplomacy.factions["merchant_guild"]
	give1, get1, op, trades := ge.Resources.Get(d.Give), ge.Resources.Get(d.Get), fs.Opinion, fs.TradeCount
	ge.mu.RUnlock()
	if give1 != give0-d.GiveAmt || get1 != get0+d.GetAmt {
		t.Errorf("%s %v→%v (price %v), %s %v→%v (goods %v)", d.Give, give0, give1, d.GiveAmt, d.Get, get0, get1, d.GetAmt)
	}
	if op != dealTradeStanding || trades != 1 {
		t.Errorf("opinion %d, trades %d after one deal; want %d, 1", op, trades, dealTradeStanding)
	}
	if _, err := ge.AcceptFactionDeal("merchant_guild", 1); err == nil || !strings.Contains(err.Error(), "already taken") {
		t.Errorf("second accept: %v", err)
	}
	ge.mu.RLock()
	if ge.Resources.Get(d.Give) != give1 || ge.Resources.Get(d.Get) != get1 {
		t.Error("the second accept moved resources")
	}
	ge.mu.RUnlock()
	if f := ge.GetState().Diplomacy.Factions["merchant_guild"]; !f.Deals[0].Taken {
		t.Error("the panel does not show the deal taken")
	}

	// Short of the price: refused, nothing moves.
	ge.mu.Lock()
	d2 := ge.Diplomacy.factions["merchant_guild"].Deals[1]
	ge.Resources.resources[d2.Give].Amount = d2.GiveAmt / 2
	ge.mu.Unlock()
	if _, err := ge.AcceptFactionDeal("merchant_guild", 2); err == nil || !strings.Contains(err.Error(), "Not enough") {
		t.Errorf("unaffordable accept: %v", err)
	}
	for _, bad := range []int{0, 99} {
		if _, err := ge.AcceptFactionDeal("merchant_guild", bad); err == nil {
			t.Errorf("accept %d: no error", bad)
		}
	}
	if _, err := ge.AcceptFactionDeal("nobody", 1); err == nil {
		t.Error("accept from an unknown civ: no error")
	}
}

// TestDeals_StandingCap: deals raise opinion to dealStandingCap and no
// further, and never lower one a gift already raised past it.
func TestDeals_StandingCap(t *testing.T) {
	ge := dealEngine(t, "bronze_age")
	meet(ge, "riverlands_tribes", 0)
	favor := func(id, opinion int) int {
		ge.mu.Lock()
		fs := ge.Diplomacy.factions["riverlands_tribes"]
		fs.Opinion = opinion
		fs.Deals = []FactionDeal{{ID: id, Kind: DealFavor, Give: "wood", GiveAmt: 10, Standing: dealFavorStanding}}
		ge.mu.Unlock()
		if _, err := ge.AcceptFactionDeal("riverlands_tribes", 1); err != nil {
			t.Fatal(err)
		}
		ge.mu.RLock()
		defer ge.mu.RUnlock()
		return ge.Diplomacy.factions["riverlands_tribes"].Opinion
	}
	if got := favor(100, 10); got != 10+dealFavorStanding {
		t.Errorf("favor from 10: opinion %d", got)
	}
	if got := favor(101, dealStandingCap-2); got != dealStandingCap {
		t.Errorf("favor near the cap: opinion %d, want %d", got, dealStandingCap)
	}
	if got := favor(102, 70); got != 70 {
		t.Errorf("favor above the cap: opinion %d, want 70 untouched", got)
	}
	// Friendly status follows the opinion, as for a gift.
	if st := ge.GetState().Diplomacy.Factions["riverlands_tribes"].Status; st != "friendly" {
		t.Errorf("status %q after deals past 25 opinion", st)
	}
}

// TestDeals_Refresh: offers roll on first contact, hold for dealRefreshTicks
// ticks of play, re-roll then and on an age advance, and do not move while
// the player is offline.
func TestDeals_Refresh(t *testing.T) {
	ge := dealEngine(t, "colonial_age")
	meet(ge, "merchant_guild", 20)
	state := func() FactionState {
		ge.mu.RLock()
		defer ge.mu.RUnlock()
		fs := *ge.Diplomacy.factions["merchant_guild"]
		fs.Deals = append([]FactionDeal(nil), fs.Deals...)
		return fs
	}
	first := state()
	if first.DealRound != 1 || first.DealTicks != 0 || first.DealsFor != "colonial_age" || len(first.Deals) == 0 {
		t.Fatalf("after first contact: round %d ticks %d for %q, %d deals", first.DealRound, first.DealTicks, first.DealsFor, len(first.Deals))
	}
	tick := func(n int) {
		ge.mu.Lock()
		for i := 0; i < n; i++ {
			ge.tickFactionDeals()
		}
		ge.mu.Unlock()
	}
	tick(dealRefreshTicks - 1)
	if s := state(); s.DealRound != 1 || !reflect.DeepEqual(s.Deals, first.Deals) {
		t.Fatalf("offers rotated early: round %d", s.DealRound)
	}
	if left := ge.GetState().Diplomacy.Factions["merchant_guild"].DealRefreshIn; left != 1 {
		t.Errorf("refresh in %d ticks, want 1", left)
	}

	// Offline: a day away moves neither the timer nor the offers.
	ge.SimulateOffline(20 * time.Hour)
	if s := state(); s.DealRound != 1 || s.DealTicks != dealRefreshTicks-1 || !reflect.DeepEqual(s.Deals, first.Deals) {
		t.Fatalf("offline moved the offers: round %d, ticks %d", s.DealRound, s.DealTicks)
	}

	tick(1)
	second := state()
	if second.DealRound != 2 || second.DealTicks != 0 {
		t.Fatalf("after %d ticks: round %d, ticks %d", dealRefreshTicks, second.DealRound, second.DealTicks)
	}
	for _, d := range second.Deals {
		if d.ID <= dealIDStride*2 || d.ID > dealIDStride*3 {
			t.Errorf("round 2 deal ID %d outside its round", d.ID)
		}
	}

	// An age advance re-rolls on the next tick, whatever the timer says.
	ge.mu.Lock()
	ge.advanceAge("industrial_age")
	ge.pendingCatastrophe = ""
	ge.mu.Unlock()
	if n := len(ge.GetState().Diplomacy.Factions["merchant_guild"].Deals); n != 0 {
		t.Errorf("colonial offers still shown in the Industrial Age (%d)", n)
	}
	if _, err := ge.AcceptFactionDeal("merchant_guild", 1); err == nil {
		t.Error("took a colonial offer in the Industrial Age")
	}
	tick(1)
	if s := state(); s.DealRound != 3 || s.DealsFor != "industrial_age" {
		t.Errorf("after the advance: round %d for %q", s.DealRound, s.DealsFor)
	}
}

// TestDeals_SaveLoadRoundTrip: offers, taken flags, the round and the timer
// survive a save and load exactly.
func TestDeals_SaveLoadRoundTrip(t *testing.T) {
	isolateAccountDir(t)
	ge := dealEngine(t, "colonial_age")
	meet(ge, "merchant_guild", 30)
	meet(ge, "riverlands_tribes", 10)
	if _, err := ge.AcceptFactionDeal("merchant_guild", 1); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	for i := 0; i < 123; i++ {
		ge.tickFactionDeals()
	}
	want := ge.Diplomacy.GetFactionsForSave()
	ge.mu.Unlock()
	if err := ge.SaveGame("deals-roundtrip"); err != nil {
		t.Fatal(err)
	}
	b := NewGameEngine()
	if err := b.LoadGame("deals-roundtrip"); err != nil {
		t.Fatal(err)
	}
	b.mu.RLock()
	got := b.Diplomacy.GetFactionsForSave()
	b.mu.RUnlock()
	if !reflect.DeepEqual(got, want) {
		gj, _ := json.Marshal(got)
		wj, _ := json.Marshal(want)
		t.Fatalf("round trip:\n got %s\nwant %s", gj, wj)
	}
	if !want["merchant_guild"].Deals[0].Taken || want["merchant_guild"].DealTicks == 0 {
		t.Fatal("the scenario saved no taken deal or timer; the test would pass vacuously")
	}

	// A civ saved before deals existed carries none of the new fields and
	// rolls a fresh set on its first tick.
	old, _ := json.Marshal(FactionStateSave{Discovered: true, Status: "neutral"})
	if strings.Contains(string(old), "deal") {
		t.Errorf("empty deal fields are not omitted: %s", old)
	}
}

// TestDeals_Deterministic: the same seed rolls the same offers, a different
// seed different ones.
func TestDeals_Deterministic(t *testing.T) {
	roll := func(seed int64) string {
		ge := dealEngine(t, "colonial_age")
		ge.SeedRNG(seed)
		for _, key := range []string{"riverlands_tribes", "ironhold_clans", "merchant_guild"} {
			meet(ge, key, 30)
		}
		ge.mu.RLock()
		defer ge.mu.RUnlock()
		j, _ := json.Marshal(ge.Diplomacy.GetFactionsForSave())
		return string(j)
	}
	if a, b := roll(3), roll(3); a != b {
		t.Fatalf("same seed, different offers:\n%s\n%s", a, b)
	}
	if roll(3) == roll(4) {
		t.Error("different seeds rolled identical offers")
	}
}

// TestDeals_Plan: a planned deal waits for its price (holding it from items
// below), is taken once, and drops out when its offer rotates away.
func TestDeals_Plan(t *testing.T) {
	ge := dealEngine(t, "colonial_age")
	meet(ge, "merchant_guild", 30)
	ge.mu.Lock()
	d := ge.Diplomacy.factions["merchant_guild"].Deals[0]
	ge.Resources.resources[d.Give].Amount = d.GiveAmt / 2
	ge.mu.Unlock()
	if err := ge.PlanAddDeal("merchant_guild", 1); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddDeal("merchant_guild", 1); err == nil {
		t.Error("planned the same deal twice")
	}
	if err := ge.PlanAddDeal("merchant_guild", 99); err == nil {
		t.Error("planned a deal that doesn't exist")
	}
	ge.mu.Lock()
	ge.runPlanTick()
	views := ge.planViews()
	ge.mu.Unlock()
	if len(views) != 1 || views[0].Status != PlanStatusWaiting || views[0].Short != d.Give || math.Abs(views[0].Progress-0.5) > 1e-9 {
		t.Fatalf("waiting deal item: %+v", views)
	}
	ge.mu.Lock()
	ge.Resources.resources[d.Give].Amount = d.GiveAmt * 3
	get0 := ge.Resources.Get(d.Get)
	ge.runPlanTick()
	taken := ge.Diplomacy.factions["merchant_guild"].Deals[0].Taken
	left, get1 := len(ge.plan), ge.Resources.Get(d.Get)
	ge.mu.Unlock()
	if !taken || left != 0 || get1 != get0+d.GetAmt {
		t.Fatalf("after the price came in: taken %v, %d items left, %s %v→%v", taken, left, d.Get, get0, get1)
	}

	// A planned deal whose offers rotate away drops out with a log line.
	if err := ge.PlanAddDeal("merchant_guild", 2); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	ge.Diplomacy.factions["merchant_guild"].DealTicks = dealRefreshTicks
	ge.tickFactionDeals()
	ge.runPlanTick()
	left = len(ge.plan)
	ge.mu.Unlock()
	if left != 0 {
		t.Error("a deal item outlived its offer")
	}
	found := false
	for _, l := range ge.GetState().Log {
		if strings.Contains(l.Message, "the offer is gone") {
			found = true
		}
	}
	if !found {
		t.Error("no log line for the dropped deal item")
	}
}

// TestDeals_Wording: every surface words a deal from the player's side,
// "<Kind>: give X → get Y", with the civ selling its specialty as your Buy
// and the civ wanting your goods as your Sell; the log line and the plan
// label carry the same terms.
func TestDeals_Wording(t *testing.T) {
	num := textfmt.Number
	for _, c := range []struct {
		d    FactionDeal
		want string
	}{
		{FactionDeal{Kind: DealSell, Give: "coal", GiveAmt: 876e6, Get: "food", GetAmt: 966e3}, "Buy: give 876M coal → get 966K food"},
		{FactionDeal{Kind: DealWant, Give: "iron_ore", GiveAmt: 440e6, Get: "food", GetAmt: 877e3}, "Sell: give 440M iron ore → get 877K food"},
		{FactionDeal{Kind: DealFavor, Give: "steel", GiveAmt: 899e6, Standing: 5}, "Goodwill: give 899M steel → get +5 opinion"},
		{FactionDeal{Kind: DealRare, Give: "electricity", GiveAmt: 25.7e9, Get: "oil", GetAmt: 3.3e9}, "Rare: give 25.7B electricity → get 3.3B oil"},
	} {
		if got := DealTerms(c.d.Kind, c.d.Give, c.d.GiveAmt, c.d.Get, c.d.GetAmt, c.d.Standing, num); got != c.want {
			t.Errorf("DealTerms(%s) = %q, want %q", c.d.Kind, got, c.want)
		}
	}
	// The log and the plan label word the same terms as a clause.
	if got, want := dealLogTerms(FactionDeal{Kind: DealSell, Give: "iron_ore", GiveAmt: 876e6, Get: "food", GetAmt: 966e3}), "Buy: give 876M iron ore, get 966K food"; got != want {
		t.Errorf("dealLogTerms = %q, want %q", got, want)
	}

	ge := dealEngine(t, "colonial_age")
	meet(ge, "merchant_guild", 30)
	if err := ge.PlanAddDeal("merchant_guild", 2); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	d1, d2 := ge.Diplomacy.factions["merchant_guild"].Deals[0], ge.Diplomacy.factions["merchant_guild"].Deals[1]
	ge.Resources.resources[d2.Give].Amount = 0
	views := ge.planViews()
	ge.mu.Unlock()
	if want := "deal with the Merchant Guild (" + dealLogTerms(d2) + ")"; len(views) != 1 || views[0].Name != want {
		t.Errorf("plan label %+v, want %q", views, want)
	}
	if _, err := ge.AcceptFactionDeal("merchant_guild", 1); err != nil {
		t.Fatal(err)
	}
	want := "Deal with the Merchant Guild. " + dealLogTerms(d1) + "."
	found := false
	for _, l := range ge.GetState().Log {
		found = found || l.Message == want
	}
	if !found {
		t.Errorf("no log line %q", want)
	}
}
