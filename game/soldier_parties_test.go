package game

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
)

// partyTown is a Stone Age town with n War Camps whose soldiers have taken
// their posts, stocked for short missions and with a spare reserve of
// reserve people, seeded so a mission's luck is the same run to run.
func partyTown(t *testing.T, camps, reserve int) *GameEngine {
	t.Helper()
	ge := peopleEngine(t)
	ge.SeedRNG(77)
	learn(ge, "exploration")
	// Food for a big army to eat.
	ge.mu.Lock()
	ge.Buildings.counts["gathering_camp"] = 600
	ge.recalculateRates()
	ge.mu.Unlock()
	standWarCamps(ge, camps)
	for i := 0; i < 20 && ge.GetState().Workers.Soldiers < camps*warCampPosts(t); i++ {
		ge.StepTicks(staffEveryTicks)
	}
	if got, want := ge.GetState().Workers.Soldiers, camps*warCampPosts(t); got != want {
		t.Fatalf("setup: %d soldiers posted, want %d", got, want)
	}
	ge.mu.Lock()
	ge.Workers.Recruit("worker", reserve, 1<<20)
	ge.recalculateRates()
	ge.mu.Unlock()
	for _, res := range []string{"stone", "wood", "food", "gold"} {
		setResource(ge, res, 10*config.Million)
		setIncome(ge, res, 1)
	}
	return ge
}

// A party takes its place out of the posts: the soldiers away are not at home
// (the soldiers resource, defense and further launches see only those at
// home), and they are still people (the population, the housing, the food).
func TestPartyLeavesItsPostsAndStaysPeople(t *testing.T) {
	ge := partyTown(t, 3, 0)
	posted := 3 * warCampPosts(t)
	party := config.PartySize(config.PartyShort, 1)
	popBefore := ge.GetState().Workers.TotalPop
	drainBefore := ge.GetState().Workers.FoodDrain

	noErr(t, ge.LaunchExpedition(plantOffer(ge, TierShort, "")))
	st := ge.GetState()
	if st.Military.SoldierCount != posted-party || st.Military.SoldiersAway != party {
		t.Errorf("%d at home and %d away, want %d and %d", st.Military.SoldierCount, st.Military.SoldiersAway, posted-party, party)
	}
	if st.Workers.Soldiers != posted || st.Workers.SoldiersAway != party {
		t.Errorf("the town has %d soldiers (%d out), want %d (%d out)", st.Workers.Soldiers, st.Workers.SoldiersAway, posted, party)
	}
	if st.Workers.TotalPop != popBefore || st.Workers.FoodDrain != drainBefore {
		t.Errorf("population %d and food drain %v, were %d and %v: soldiers away are still people", st.Workers.TotalPop, st.Workers.FoodDrain, popBefore, drainBefore)
	}
	if st.Military.Posts[0].Posted != posted {
		t.Errorf("%d posted, want all %d (the party is still posted, only away)", st.Military.Posts[0].Posted, posted)
	}
	ge.mu.RLock()
	rating := ge.defenseRating()
	ge.mu.RUnlock()
	if want := ge.Military.CalculateDefense(posted-party, ge.militaryPower()); rating != want {
		t.Errorf("defense rating %v, want %v for the %d at home", rating, want, posted-party)
	}
}

// Two missions of one tier can be out together while the soldiers at home
// cover both; a launch they cannot cover is refused with the party and the
// soldiers at home, and takes nothing.
func TestSameTierTogetherAndTheRefusalNamesTheNumbers(t *testing.T) {
	ge := partyTown(t, 1, 0)
	posted := warCampPosts(t)
	party := config.PartySize(config.PartyShort, 1)
	if posted < 2*party+1 {
		t.Fatalf("setup: %d posts cannot make two parties of %d and a refusal", posted, party)
	}
	for i := 0; i < 2; i++ {
		if err := ge.LaunchExpedition(plantOffer(ge, TierShort, "")); err != nil {
			t.Fatalf("short mission %d: %v", i+1, err)
		}
	}
	if got := len(ge.Military.ActiveOf(TierShort)); got != 2 {
		t.Fatalf("%d short missions out, want 2", got)
	}
	// Send long parties until the town cannot make up another.
	home := posted - 2*party
	long := config.PartySize(config.PartyLong, 1)
	n := plantOffer(ge, TierLong, "")
	err := ge.LaunchExpedition(n)
	if home >= long {
		if err != nil {
			t.Fatalf("a long party of %d with %d at home: %v", long, home, err)
		}
		return
	}
	want := fmt.Sprintf("needs a party of %d soldiers, and %d are at home", long, home)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("launch with %d at home: %v, want %q", home, err, want)
	}
	if reason := ge.GetState().Military.Expeditions; len(reason) == 0 {
		t.Error("no offers on the board")
	}
	if _, err := ge.Military.BoardOffer(n); err != nil {
		t.Errorf("the refused offer was taken: %v", err)
	}
	if got := ge.soldiersHome(); got != home {
		t.Errorf("%d at home after a refusal, want %d", got, home)
	}
}

// The losses stay inside their ranges and are the same for one seed.
func TestMissionLossesRanges(t *testing.T) {
	const party = 60
	var winLosses, winHits, trials int
	for seed := int64(0); seed < 4000; seed++ {
		trials++
		fail := missionLosses(rand.New(rand.NewSource(seed)), party, false, false)
		if fail < 24 || fail > 42 { // ceil(60 x 0.4) to ceil(60 x 0.7)
			t.Fatalf("seed %d: a failure lost %d of %d, want 24 to 42", seed, fail, party)
		}
		rare := missionLosses(rand.New(rand.NewSource(seed)), party, false, true)
		if rare < 36 || rare > 54 { // 0.6 to 0.9
			t.Fatalf("seed %d: a Legendary failure lost %d of %d, want 36 to 54", seed, rare, party)
		}
		win := missionLosses(rand.New(rand.NewSource(seed)), party, true, false)
		if win < 0 || win > 9 { // ceil(60 x 0.15)
			t.Fatalf("seed %d: a success lost %d of %d, want 0 to 9", seed, win, party)
		}
		if win > 0 {
			winHits++
			winLosses += win
		}
		if again := missionLosses(rand.New(rand.NewSource(seed)), party, true, false); again != win {
			t.Fatalf("seed %d: %d then %d, want the same draw", seed, win, again)
		}
	}
	// A success costs soldiers 30% of the time.
	if share := float64(winHits) / float64(trials); share < .27 || share > .33 {
		t.Errorf("%.1f%% of successes lost soldiers, want about 30%%", share*100)
	}
	if missionLosses(rand.New(rand.NewSource(1)), 0, false, false) != 0 {
		t.Error("an empty party lost soldiers")
	}
	// ceil: even a party of one loses its soldier on a failure.
	if got := missionLosses(rand.New(rand.NewSource(1)), 1, false, false); got != 1 {
		t.Errorf("a party of 1 lost %d on a failure, want 1", got)
	}
}

// forceReturn makes the mission under way come home next tick, failing or
// succeeding, and returns the losses its own stream is about to draw.
func forceReturn(ge *GameEngine, success bool) (a *ActiveExpedition, wantFell int) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	a = ge.Military.active[0]
	a.TicksLeft = 1
	if success {
		a.Luck, a.SuccessChance = 0, 1
	} else {
		a.Luck, a.SuccessChance = .999, .1
	}
	probe := rand.New(rand.NewSource(a.LuckSeed))
	probe.Float64() // the success roll, drawn at launch
	return a, missionLosses(probe, a.Soldiers, success, a.Rare)
}

// Survivors go back to their posts, the fallen leave the population and their
// posts stay empty; the returns log and the story say how many fell.
func TestFailedMissionLosesSoldiersAndSurvivorsReturn(t *testing.T) {
	ge := partyTown(t, 3, 0)
	posted := 3 * warCampPosts(t)
	noErr(t, ge.LaunchExpedition(plantOffer(ge, TierLong, "")))
	party := ge.Military.active[0].Soldiers
	popBefore := ge.GetState().Workers.TotalPop
	_, fell := forceReturn(ge, false)
	if fell < 1 || fell >= party {
		t.Fatalf("setup: %d of %d fall, want some but not all", fell, party)
	}
	ge.mu.Lock()
	ge.processExpeditions()
	ge.mu.Unlock()

	st := ge.GetState()
	if st.Military.SoldiersAway != 0 || ge.Military.HasActive() {
		t.Fatalf("%d still away, mission active %v", st.Military.SoldiersAway, ge.Military.HasActive())
	}
	if st.Military.SoldierCount != posted-fell {
		t.Errorf("%d soldiers at home, want %d (%d posted less the %d who fell)", st.Military.SoldierCount, posted-fell, posted, fell)
	}
	if st.Workers.Soldiers != posted-fell || st.Workers.TotalPop != popBefore-fell {
		t.Errorf("town has %d soldiers and %d people, want %d and %d: the fallen leave the population", st.Workers.Soldiers, st.Workers.TotalPop, posted-fell, popBefore-fell)
	}
	if got := st.Military.Posts[0]; got.Posted != posted-fell || got.Posts != posted {
		t.Errorf("posts %+v, want %d of %d", got, posted-fell, posted)
	}
	want := fmt.Sprintf("%d of %d soldiers fell.", fell, party)
	if len(st.Military.Returns) == 0 || st.Military.Returns[0].Fell != fell || st.Military.Returns[0].Party != party {
		t.Errorf("returns log %+v, want %d of %d", st.Military.Returns, fell, party)
	}
	if !logContains(ge, "failed. You kept 30% of the loot") || !logContains(ge, want) {
		t.Errorf("the log does not say %q", want)
	}
}

// A success sends everyone home or costs a few; either way the survivors are
// back at their posts and the count in the returns log is right.
func TestSucceededMissionBringsTheParty(t *testing.T) {
	for _, seed := range []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
		ge := partyTown(t, 3, 0)
		ge.SeedRNG(seed)
		posted := 3 * warCampPosts(t)
		noErr(t, ge.LaunchExpedition(plantOffer(ge, TierLong, "")))
		party := ge.Military.active[0].Soldiers
		_, fell := forceReturn(ge, true)
		ge.mu.Lock()
		ge.processExpeditions()
		ge.mu.Unlock()
		st := ge.GetState()
		if st.Military.SoldierCount != posted-fell || st.Military.SoldiersAway != 0 {
			t.Fatalf("seed %d: %d at home, %d away, want %d and 0", seed, st.Military.SoldierCount, st.Military.SoldiersAway, posted-fell)
		}
		line := fmt.Sprintf("All %d soldiers came home.", party)
		if fell > 0 {
			line = fmt.Sprintf("%d of %d soldiers fell.", fell, party)
		}
		if !logContains(ge, "succeeded. Loot:") || !logContains(ge, line) {
			t.Fatalf("seed %d: the log does not say %q", seed, line)
		}
	}
}

// An ambush on the road costs a tenth of the party of any tier.
func TestAmbushCostsAnyTiersParty(t *testing.T) {
	ge := partyTown(t, 3, 0)
	noErr(t, ge.LaunchExpedition(plantOffer(ge, TierLong, "")))
	a := ge.Military.active[0]
	party := a.Soldiers
	words, _ := ge.rules.MissionWords(a.Age)
	posted := ge.soldierTotal()
	ge.mu.Lock()
	story := ge.applyRoadEvent(a, &RoadEvent{Kind: "ambush", Story: words.RoadTemplate("ambush", 0), AmbushStory: words.RoadTemplate("ambush_haul", 0)})
	ge.mu.Unlock()
	lost := max(1, (party+9)/10)
	if a.Soldiers != party-lost || ge.soldierTotal() != posted-lost {
		t.Errorf("party %d and %d posted after the ambush, want %d and %d", a.Soldiers, ge.soldierTotal(), party-lost, posted-lost)
	}
	if strings.Contains(story, "<") {
		t.Errorf("story left a placeholder: %q", story)
	}
}

// A raid that lands against the garrison costs ceil(5%) of the soldiers at
// home, at least one; nobody home, nobody falls.
func TestRaidCostsSoldiersAtHome(t *testing.T) {
	ge := partyTown(t, 3, 0)
	posted := 3 * warCampPosts(t)
	ge.mu.Lock()
	said := ge.raidFallLocked()
	ge.mu.Unlock()
	want := (5*posted + 99) / 100
	if got := posted - ge.soldierTotal(); got != want || !strings.Contains(said, fmt.Sprint(want)) {
		t.Errorf("%d fell (%q), want %d", got, said, want)
	}
	if st := ge.GetState(); st.Military.Posts[0].Posted != posted-want || st.Workers.Soldiers != posted-want {
		t.Errorf("posts %+v, want %d posted", st.Military.Posts, posted-want)
	}

	// Everyone away: nobody at home to fall.
	empty := partyTown(t, 1, 0)
	empty.mu.Lock()
	empty.Military.active = append(empty.Military.active, &ActiveExpedition{MissionOffer: MissionOffer{Tier: TierShort}, Soldiers: warCampPosts(t), TicksLeft: 99})
	if said := empty.raidFallLocked(); said != "" {
		t.Errorf("a raid with the army away cost %q", said)
	}
	empty.mu.Unlock()

	// A raid on a town with one soldier at home takes that one.
	one := partyTown(t, 1, 0)
	one.mu.Lock()
	one.Military.active = append(one.Military.active, &ActiveExpedition{MissionOffer: MissionOffer{Tier: TierShort}, Soldiers: warCampPosts(t) - 1, TicksLeft: 99})
	one.syncSoldiers()
	if said := one.raidFallLocked(); !strings.Contains(said, "1 soldier") {
		t.Errorf("a raid with one soldier home: %q, want 1 soldier", said)
	}
	one.mu.Unlock()
}

// A war raid that lands and is blunted by the garrison costs soldiers and
// says so in the log.
func TestWarRaidLogsTheFallen(t *testing.T) {
	faction := config.BaseFactions()[0].Key
	ge := partyTown(t, 3, 0)
	posted := ge.soldierTotal()
	ge.mu.Lock()
	ge.age = "iron_age"
	ge.Resources.UnlockResource("gold")
	ge.Resources.Add("gold", 1000)
	ge.Diplomacy.pendingRaids = []RaidRequest{{FactionKey: faction, Resource: "gold", Amount: 200, Message: "raided"}}
	ge.applyWarRaids()
	ge.mu.Unlock()
	if fell := posted - ge.soldierTotal(); fell != (5*posted+99)/100 {
		t.Errorf("%d soldiers fell in the raid, want %d", fell, (5*posted+99)/100)
	}
	if !logContains(ge, "fell defending the town") {
		t.Error("the raid did not say soldiers fell")
	}
}

// The fallen are replaced from the reserve at most a pass at a time: 1% of the
// posts, at least one.
func TestFallenPostsRetrainFromTheReserveSlowly(t *testing.T) {
	ge := partyTown(t, 15, 30)
	posts := ge.soldierPostsTotal()
	budget := max(1, (posts+99)/100)
	ge.mu.Lock()
	fell := ge.soldiersFall(budget*3 + 1)
	posted := ge.soldierTotal()
	ge.mu.Unlock()
	if fell != budget*3+1 {
		t.Fatalf("setup: %d fell", fell)
	}
	setIncome(ge, "food", 100) // the food margin holds
	idle := ge.GetState().Workers.TotalIdle
	if idle < 3*budget {
		t.Fatalf("setup: %d in the reserve, want at least %d", idle, 3*budget)
	}
	for pass := 1; pass <= 3; pass++ {
		ge.mu.Lock()
		got := ge.fillSoldierPosts()
		ge.mu.Unlock()
		if got != budget {
			t.Fatalf("pass %d trained %d, want %d (1%% of %d posts)", pass, got, budget, posts)
		}
		posted += got
		if ge.soldierTotal() != posted {
			t.Fatalf("pass %d: %d posted, want %d", pass, ge.soldierTotal(), posted)
		}
	}
	ge.mu.Lock()
	got := ge.fillSoldierPosts()
	ge.mu.Unlock()
	if got != 1 || ge.soldierTotal() != posts {
		t.Errorf("the last vacancy: %d trained, %d of %d posted", got, ge.soldierTotal(), posts)
	}
	ge.mu.Lock()
	again := ge.fillSoldierPosts()
	ge.mu.Unlock()
	if again != 0 {
		t.Errorf("a full army trained %d more", again)
	}
}

// Training takes the reserve only: with nobody idle the posts stay empty, and
// no newcomer is brought in for them. And the food margin holds them too.
func TestFallenPostsNeedTheReserveAndTheFood(t *testing.T) {
	ge := partyTown(t, 3, 0)
	posted := ge.soldierTotal()
	ge.mu.Lock()
	ge.soldiersFall(2)
	ge.Workers.TakeIdle(1 << 20) // nobody in the reserve
	ge.mu.Unlock()
	for i := 0; i < 4; i++ {
		ge.mu.Lock()
		got := ge.fillSoldierPosts()
		ge.mu.Unlock()
		if got != 0 {
			t.Fatalf("trained %d with an empty reserve", got)
		}
	}
	if ge.soldierTotal() != posted-2 {
		t.Errorf("%d posted, want %d", ge.soldierTotal(), posted-2)
	}
	// A reserve, but no food: they wait.
	ge.mu.Lock()
	ge.Workers.Recruit("worker", 2, 1<<20)
	ge.Resources.resources["food"].Amount = 0
	got := ge.fillSoldierPosts()
	ge.mu.Unlock()
	if got != 0 {
		t.Errorf("trained %d with no food", got)
	}
}

// The posts the fallen left survive a save and a load.
func TestVacantPostsSurviveASaveAndLoad(t *testing.T) {
	defer SetDataDirForTest(t.TempDir())()
	ge := partyTown(t, 3, 0)
	ge.mu.Lock()
	ge.soldiersFall(4)
	want := ge.soldiersVacantSave()
	ge.mu.Unlock()
	if len(want) == 0 {
		t.Fatal("setup: no vacancies")
	}
	again := saveAndLoad(t, ge, "vacant_posts")
	again.mu.RLock()
	got := again.soldiersVacantSave()
	again.mu.RUnlock()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("vacancies after the load %v, want %v", got, want)
	}
}

// A mission with losses reaches the same end played live or caught up
// offline: the party that fell, the soldiers posted and the returns entry.
func TestMissionLossesOfflineMatchLive(t *testing.T) {
	build := func() (*GameEngine, int) {
		ge := partyTown(t, 3, 0)
		noErr(t, ge.LaunchExpedition(plantOffer(ge, TierLong, "")))
		ge.mu.Lock()
		a := ge.Military.active[0]
		a.Luck, a.SuccessChance = .999, .1 // a failure: the party loses most of its soldiers
		total := a.TotalTicks
		ge.mu.Unlock()
		return ge, total
	}
	live, total := build()
	offline, _ := build()
	party := live.Military.active[0].Soldiers

	for i := 0; i < total; i++ {
		live.doTick()
	}
	offline.SimulateOffline(time.Duration(total) * offline.getTickInterval())

	for name, ge := range map[string]*GameEngine{"live": live, "offline": offline} {
		if ge.Military.HasActive() {
			t.Fatalf("%s: the mission is still out", name)
		}
	}
	lr, or := live.Military.returns, offline.Military.returns
	if len(lr) != 1 || len(or) != 1 {
		t.Fatalf("returns: live %d, offline %d, want 1 each", len(lr), len(or))
	}
	if lr[0].Fell != or[0].Fell || lr[0].Party != party || or[0].Party != party || lr[0].Fell < 1 {
		t.Errorf("live lost %d of %d, offline %d of %d", lr[0].Fell, lr[0].Party, or[0].Fell, or[0].Party)
	}
	// Posts match to the soldier, counting the ones still to retrain.
	lp, op := live.soldierTotal()+live.vacantTotal(), offline.soldierTotal()+offline.vacantTotal()
	if lp != op {
		t.Errorf("live has %d posted+vacant, offline %d", lp, op)
	}
	if live.vacantTotal() != offline.vacantTotal() {
		t.Errorf("right after the return live has %d posts to retrain, offline %d", live.vacantTotal(), offline.vacantTotal())
	}

	// And the retraining that follows: the same posts refill in the same time.
	const more = staffEveryTicks * 4
	for _, ge := range []*GameEngine{live, offline} {
		ge.mu.Lock()
		ge.Workers.Recruit("worker", 10, 1<<20)
		ge.mu.Unlock()
		setIncome(ge, "food", 100)
	}
	for i := 0; i < more; i++ {
		live.doTick()
	}
	offline.SimulateOffline(time.Duration(more) * offline.getTickInterval())
	if live.soldierTotal() != offline.soldierTotal() || live.vacantTotal() != offline.vacantTotal() {
		t.Errorf("after %d more ticks live has %d posted (%d to retrain), offline %d (%d)",
			more, live.soldierTotal(), live.vacantTotal(), offline.soldierTotal(), offline.vacantTotal())
	}
}

// vacantTotal is the posts the fallen left empty.
func (ge *GameEngine) vacantTotal() int {
	n := 0
	for _, v := range ge.soldiersVacant {
		n += v
	}
	return n
}

// The Primitive Age has no army and no military building: its missions take
// no party, so a town with no soldiers can send an offer, with no refusal, no
// losses and no party on the briefing.
func TestPrimitiveAgeSendsWithoutAParty(t *testing.T) {
	ge := NewGameEngine()
	ge.SeedRNG(5)
	setAge(ge, "primitive_age")
	for _, res := range []string{"wood", "food", "stone"} {
		setResource(ge, res, config.Million)
		setIncome(ge, res, 1)
	}
	n := plantOffer(ge, TierShort, "")
	if got := ge.MissionCost(n)["soldiers"]; got != 0 {
		t.Fatalf("a Primitive Age mission costs %v soldiers, want none", got)
	}
	for _, o := range ge.GetState().Military.Expeditions {
		if o.SoldiersNeeded != 0 {
			t.Errorf("offer %d needs %d soldiers in the Primitive Age", o.Number, o.SoldiersNeeded)
		}
	}
	if err := ge.LaunchExpedition(n); err != nil {
		t.Fatalf("a town with no soldiers could not send an offer: %v", err)
	}
	a := ge.Military.ActiveByTier(TierShort)
	if a == nil || a.Soldiers != 0 || a.Party != 0 {
		t.Fatalf("the mission is %+v, want one with no party", a)
	}
	a.TicksLeft, a.Luck, a.SuccessChance = 1, .999, .1
	ge.mu.Lock()
	ge.processExpeditions()
	ge.mu.Unlock()
	if ge.Military.HasActive() {
		t.Fatal("the mission did not come home")
	}
	r := ge.Military.returns[0]
	if r.Party != 0 || r.Fell != 0 {
		t.Errorf("returns entry %+v, want no party and nobody fallen", r)
	}
	if logContains(ge, "soldiers fell") || logContains(ge, "soldiers came home") {
		t.Error("the log talks of soldiers for a mission that sent none")
	}
	// An ambush on a party-less mission takes haul only, with no soldiers line.
	learn(ge, "exploration")
	noErr(t, ge.LaunchExpedition(plantOffer(ge, TierLong, "")))
	b := ge.Military.ActiveByTier(TierLong)
	words, _ := ge.rules.MissionWords(b.Age)
	ge.mu.Lock()
	story := ge.applyRoadEvent(b, &RoadEvent{Kind: "ambush", Story: words.RoadTemplate("ambush", 0), AmbushStory: words.RoadTemplate("ambush_haul", 0)})
	ge.mu.Unlock()
	if strings.Contains(story, "soldiers fell") || strings.Contains(story, "<") {
		t.Errorf("ambush story %q", story)
	}
}

// The posted soldiers and the empty posts go into a save and come out of one
// capped at the posts the standing copies have, with nothing for a building
// that has none, and as nil (not an empty map) when nothing is left.
func TestSoldierPostsAreCappedOnSaveAndLoad(t *testing.T) {
	ge := newSeededEngine(3)
	ge.Buildings.counts["war_camp"] = 2
	posts := ge.soldierPostsOf("war_camp")
	if posts <= 0 {
		t.Fatal("setup: war camps have no posts")
	}
	over := func() map[string]int {
		return map[string]int{"war_camp": posts + 5, "no_such_building": 3, "hut": 2}
	}
	want := map[string]int{"war_camp": posts}

	ge.soldiersPosted, ge.soldiersVacant = over(), over()
	if got := ge.soldiersPostedSave(); !reflect.DeepEqual(got, want) {
		t.Errorf("posted in the save %v, want %v", got, want)
	}
	if got := ge.soldiersVacantSave(); !reflect.DeepEqual(got, want) {
		t.Errorf("vacant in the save %v, want %v", got, want)
	}

	save := &GameSave{SoldiersPosted: over(), SoldiersVacant: over()}
	ge.loadSoldiersPosted(save)
	if !reflect.DeepEqual(ge.soldiersPosted, want) || !reflect.DeepEqual(ge.soldiersVacant, want) {
		t.Errorf("after the load posted %v vacant %v, want %v", ge.soldiersPosted, ge.soldiersVacant, want)
	}
	ge.soldiersPosted["war_camp"] = 0
	if save.SoldiersPosted["war_camp"] != posts+5 {
		t.Error("the loaded soldiers share their map with the save")
	}

	ge.soldiersPosted, ge.soldiersVacant = map[string]int{"hut": 4, "war_camp": 0}, map[string]int{}
	if ge.soldiersPostedSave() != nil || ge.soldiersVacantSave() != nil {
		t.Error("nothing posted saved as a map, want nil")
	}
	ge.loadSoldiersPosted(&GameSave{SoldiersPosted: map[string]int{"hut": 4}, SoldiersVacant: map[string]int{}})
	if ge.soldiersPosted != nil || ge.soldiersVacant != nil {
		t.Errorf("an empty load left posted %v vacant %v, want nil", ge.soldiersPosted, ge.soldiersVacant)
	}
}
