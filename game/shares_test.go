package game

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// sharesEngine is a Primitive Age game with big stores, housing for huts×10
// and the given buildings, and no workers yet.
func sharesEngine(t *testing.T, huts int, counts map[string]int) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	ge.Buildings.counts["stash"] = 50
	ge.Buildings.counts["hut"] = huts
	for k, n := range counts {
		ge.Buildings.counts[k] = n
	}
	ge.recalculateRates()
	setAmount(ge, "food", 1000)
	return ge
}

// slotsOf is how many workers key's built copies take.
func slotsOf(ge *GameEngine, key string) int {
	return ge.Buildings.defs[key].WorkerCapacity * ge.Buildings.GetCount(key)
}

// A fresh run: a hut and two camps built by command, and once they stand the
// routine recruits into every slot and puts everyone to work, with the food
// income still positive and a line in the main log saying it happens on its
// own.
func TestShares_FreshRunGetsStaffed(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["stash"] = 20
	ge.recalculateRates()
	setAmount(ge, "wood", 2000)
	setAmount(ge, "food", 200)
	for _, k := range []string{"hut", "gathering_camp", "wood_camp"} {
		if err := ge.BuildBuilding(k); err != nil {
			t.Fatalf("build %s: %v", k, err)
		}
	}
	for i := 0; i < 500 && len(ge.buildQueue) > 0; i++ {
		ge.StepTicks(1)
	}
	ge.StepTicks(staffEveryTicks)

	st := ge.GetState()
	if st.Workers.TotalPop == 0 {
		t.Fatal("nobody was recruited into the new camps")
	}
	if st.Workers.TotalIdle != 0 {
		t.Errorf("%d workers left idle", st.Workers.TotalIdle)
	}
	for _, k := range []string{"gathering_camp", "wood_camp"} {
		if bs := st.Buildings[k]; bs.WorkersAssigned != bs.WorkerCapacity*bs.Count {
			t.Errorf("%s has %d of its %d workers", k, bs.WorkersAssigned, bs.WorkerCapacity*bs.Count)
		}
	}
	if r := st.Resources["food"].Rate; r <= 0 {
		t.Errorf("food rate %v after the recruits", r)
	}
	if !logHas(ge, "Shares: recruited") {
		t.Error("no routine line for the recruits")
	}
	found := false
	for _, l := range ge.log {
		found = found || l.Type == "info" && strings.HasPrefix(l.Message, "Workers arrive on their own")
	}
	if !found {
		t.Error("the first recruits came without the main-log line that says why")
	}
}

// Shares hold as the population grows: with knowledge set to 50% and the
// housing opening a hut at a time, every batch of recruits keeps knowledge
// at half the workforce, as near as whole workers go.
func TestShares_HeldAsPopulationGrows(t *testing.T) {
	ge := sharesEngine(t, 1, map[string]int{"gathering_camp": 10, "wood_camp": 10, "story_circle": 15})
	if _, err := ge.SetWorkerShare("knowledge", 50); err != nil {
		t.Fatal(err)
	}
	for huts := 1; huts <= 6; huts++ {
		ge.Buildings.counts["hut"] = huts
		ge.recalculateRates()
		ge.keepSharesLive()
		pop, know := ge.Workers.TotalPop(), ge.Workers.GetDomainCount("knowledge")
		if pop != 10*huts {
			t.Fatalf("%d huts: population %d, want the housing's %d", huts, pop, 10*huts)
		}
		if d := 2*know - pop; d < -1 || d > 1 {
			t.Errorf("%d huts: knowledge has %d of %d workers, want half", huts, know, pop)
		}
		// The auto domains split the other half by their slots (30 each).
		if f, l := ge.Workers.GetDomainCount("food"), ge.Workers.GetDomainCount("lumber"); f-l < -1 || f-l > 1 {
			t.Errorf("%d huts: food %d and lumber %d workers, want an even split", huts, f, l)
		}
	}
	if n := ge.Workers.IdleCount("worker"); n != 0 {
		t.Errorf("%d idle workers", n)
	}
}

// Offline catch-up recruits and staffs as the time passes, and the welcome
// back says so.
func TestShares_OfflineRecruitsAndAssigns(t *testing.T) {
	ge := sharesEngine(t, 5, map[string]int{"gathering_camp": 5, "wood_camp": 5, "story_circle": 3})
	ge.SimulateOffline(time.Hour)
	slots := slotsOf(ge, "gathering_camp") + slotsOf(ge, "wood_camp") + slotsOf(ge, "story_circle")
	if pop := ge.Workers.TotalPop(); pop != slots {
		t.Errorf("population after an hour away = %d, want the %d slots filled", pop, slots)
	}
	if n := ge.Workers.IdleCount("worker"); n != 0 {
		t.Errorf("%d idle workers after an hour away", n)
	}
	if r := ge.Resources.GetRate("food"); r <= 0 {
		t.Errorf("food rate %v after the offline recruits", r)
	}
	if !logHas(ge, "While you were away, your worker shares recruited") {
		t.Error("the welcome back doesn't say the shares recruited")
	}
}

// A building that finishes gets its workers within a few ticks, recruited as
// the housing allows.
func TestShares_NewBuildingsGetStaffed(t *testing.T) {
	ge := sharesEngine(t, 3, map[string]int{"gathering_camp": 2, "wood_camp": 2})
	ge.keepSharesLive()
	if pop := ge.Workers.TotalPop(); pop != 12 {
		t.Fatalf("setup: population %d, want the camps' 12", pop)
	}
	setAmount(ge, "wood", 1000)
	if err := ge.BuildBuilding("story_circle"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 500 && len(ge.buildQueue) > 0; i++ {
		ge.StepTicks(1)
	}
	ge.StepTicks(staffEveryTicks)
	if got, want := ge.Workers.GetAssignedCount("worker", "story_circle"), slotsOf(ge, "story_circle"); got != want {
		t.Errorf("the new story circle has %d of its %d workers", got, want)
	}
}

// After a worker command the routine waits staffHoldTicks, so workers the
// player is moving by hand stay where the player left them; then it puts the
// idle ones back to work.
func TestShares_WaitsAfterAWorkerCommand(t *testing.T) {
	ge := sharesEngine(t, 3, map[string]int{"gathering_camp": 2, "wood_camp": 2})
	ge.keepSharesLive()
	if err := ge.UnassignWorker("wood_camp", 3); err != nil {
		t.Fatal(err)
	}
	if st := ge.GetState(); st.Workers.HoldTicks != staffHoldTicks {
		t.Errorf("HoldTicks = %d right after the command, want %d", st.Workers.HoldTicks, staffHoldTicks)
	}
	ge.StepTicks(staffHoldTicks - 1)
	if n := ge.Workers.IdleCount("worker"); n != 3 {
		t.Fatalf("%d idle workers during the wait, want the 3 the player freed", n)
	}
	ge.StepTicks(staffEveryTicks + 1)
	if n := ge.Workers.IdleCount("worker"); n != 0 {
		t.Errorf("%d idle workers after the wait, want them back at work", n)
	}
}

// Auto-recruit never takes the food income below zero, and stops at the
// housing.
func TestShares_RecruitKeepsFoodAndStopsAtHousing(t *testing.T) {
	// Food-poor: one gathering camp against many other slots.
	ge := sharesEngine(t, 10, map[string]int{"gathering_camp": 1, "wood_camp": 10, "story_circle": 10})
	for i := 0; i < 20; i++ {
		ge.keepSharesLive()
	}
	pop := ge.Workers.TotalPop()
	if pop == 0 || pop >= 100 {
		t.Fatalf("population %d: want some recruits, short of the housing (food binds first)", pop)
	}
	if r := ge.Resources.GetRate("food"); r <= 0 {
		t.Errorf("food rate %v after recruiting up to the food", r)
	}
	if got := ge.Workers.GetAssignedCount("worker", "gathering_camp"); got != slotsOf(ge, "gathering_camp") {
		t.Errorf("gathering camp has %d of its %d workers: a food worker grows more than they eat, so its slots fill first", got, slotsOf(ge, "gathering_camp"))
	}
	if st := ge.GetState(); RecruitStatus(st) != RecruitFood {
		t.Errorf("RecruitStatus = %q, want %q", RecruitStatus(st), RecruitFood)
	}

	// Housing-bound: plenty of food, two huts.
	ge = sharesEngine(t, 2, map[string]int{"gathering_camp": 20, "wood_camp": 10})
	ge.keepSharesLive()
	ge.keepSharesLive()
	if pop := ge.Workers.TotalPop(); pop != 20 {
		t.Errorf("population %d, want the housing's 20", pop)
	}
	if st := ge.GetState(); RecruitStatus(st) != RecruitHousing {
		t.Errorf("RecruitStatus = %q, want %q", RecruitStatus(st), RecruitHousing)
	}

	// Nothing recruits while the food store is empty.
	ge = sharesEngine(t, 5, map[string]int{"gathering_camp": 5})
	setAmount(ge, "food", 0)
	ge.keepSharesLive()
	if pop := ge.Workers.TotalPop(); pop != 0 {
		t.Errorf("recruited %d while workers were starving", pop)
	}
}

// With auto-recruit off nothing is recruited, but idle workers still go to
// work by the shares.
func TestShares_AutoRecruitOff(t *testing.T) {
	ge := sharesEngine(t, 5, map[string]int{"gathering_camp": 2, "wood_camp": 2})
	ge.SetAutoRecruit(false)
	ge.keepSharesLive()
	if pop := ge.Workers.TotalPop(); pop != 0 {
		t.Fatalf("recruited %d with auto-recruit off", pop)
	}
	ge.Workers.Recruit("worker", 5, 50)
	ge.keepSharesLive()
	if n := ge.Workers.IdleCount("worker"); n != 0 {
		t.Errorf("%d idle workers with free slots and auto-recruit off", n)
	}
	if ge.AutoRecruit() || ge.GetState().Workers.AutoRecruit {
		t.Error("auto-recruit reads on")
	}
	ge.SetAutoRecruit(true)
	if pop := ge.Workers.TotalPop(); pop != 12 {
		t.Errorf("turning auto-recruit on brought the population to %d, want the camps' 12 at once", pop)
	}
}

// A share of 0 keeps a domain empty: its workers move to free slots
// elsewhere, and no recruit goes there.
func TestShares_ZeroKeepsADomainEmpty(t *testing.T) {
	ge := sharesEngine(t, 1, map[string]int{"gathering_camp": 4, "wood_camp": 4, "story_circle": 4, "shrine": 2})
	ge.keepSharesLive()
	if ge.Workers.GetDomainCount("faith") == 0 {
		t.Fatal("setup: auto put nobody in the shrines")
	}
	reply, err := ge.SetWorkerShare("faith", 0)
	if err != nil {
		t.Fatal(err)
	}
	if n := ge.Workers.GetDomainCount("faith"); n != 0 {
		t.Errorf("faith still has %d workers after its share went to 0 (%s)", n, reply.Line)
	}
	if reply.Warning {
		t.Error("a faith share of 0 came back as a warning")
	}
	ge.Buildings.counts["hut"] = 5
	ge.recalculateRates()
	ge.keepSharesLive()
	if n := ge.Workers.GetDomainCount("faith"); n != 0 {
		t.Errorf("recruits went to faith (%d) at a share of 0", n)
	}
	if ge.Workers.TotalPop() <= 10 {
		t.Error("no recruits for the new housing")
	}
	// Food at 0 warns.
	if reply, _ := ge.SetWorkerShare("food", 0); !reply.Warning || !strings.Contains(reply.Line, "watch your food") {
		t.Errorf("food at 0: %+v, want a warning", reply)
	}
}

// Setting a share moves workers once to match it. After that the player's own
// moves stick: the routine only places new and idle workers.
func TestShares_SettingAShareMovesWorkersOnce(t *testing.T) {
	ge := sharesEngine(t, 2, map[string]int{"gathering_camp": 4, "wood_camp": 4, "story_circle": 6})
	ge.keepSharesLive()
	reply, err := ge.SetWorkerShare("knowledge", 60)
	if err != nil {
		t.Fatal(err)
	}
	if know := ge.Workers.GetDomainCount("knowledge"); know != 12 {
		t.Errorf("knowledge has %d of 20 workers after a 60%% share, want 12 (%s)", know, reply.Line)
	}
	if !strings.Contains(reply.Line, "Knowledge 60%") || !strings.Contains(reply.Line, "Moved") {
		t.Errorf("reply %q doesn't name the share and the move", reply.Line)
	}
	if err := ge.UnassignWorker("story_circle", 2); err != nil {
		t.Fatal(err)
	}
	if err := ge.AssignWorker("wood_camp", 2); err != nil {
		t.Fatal(err)
	}
	ge.StepTicks(staffHoldTicks + staffEveryTicks)
	if know := ge.Workers.GetDomainCount("knowledge"); know != 10 {
		t.Errorf("knowledge has %d workers after the player moved 2 out, want the 10 they left", know)
	}
}

// A share change never moves the last food workers out when the food would
// run out: with knowledge at 100% the food buildings keep enough workers for
// the food rate to stay at or above zero.
func TestShares_RebalanceKeepsFood(t *testing.T) {
	ge := sharesEngine(t, 2, map[string]int{"gathering_camp": 2, "story_circle": 10})
	ge.keepSharesLive()
	if _, err := ge.SetWorkerShare("knowledge", 100); err != nil {
		t.Fatal(err)
	}
	if ge.Workers.GetDomainCount("food") == 0 {
		t.Error("every food worker was moved out")
	}
	if r := ge.Resources.GetRate("food"); r < 0 {
		t.Errorf("food rate %v after the share change", r)
	}
}

// Within a domain this age's buildings fill before superseded ones, the
// higher tier first, the same way every time.
func TestShares_FillOrder(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["gathering_camp"] = 2 // superseded by the forager post
	ge.advanceAge("stone_age")
	ge.pendingCatastrophe = ""
	ge.Buildings.counts["stash"] = 50
	ge.Buildings.counts["hut"] = 1
	ge.Buildings.counts["forager_post"] = 2
	ge.recalculateRates()
	setAmount(ge, "food", 1000)
	ge.keepSharesLive()
	if got := ge.Workers.GetAssignedCount("worker", "forager_post"); got != slotsOf(ge, "forager_post") {
		t.Errorf("forager posts have %d of %d workers: this age's food buildings fill first", got, slotsOf(ge, "forager_post"))
	}
	if !ge.Buildings.IsLegacy("gathering_camp") {
		t.Fatal("setup: the gathering camp is not superseded")
	}
	// The same state staffs the same way.
	other := newSeededEngine(1)
	other.Buildings.counts["gathering_camp"] = 2
	other.advanceAge("stone_age")
	other.pendingCatastrophe = ""
	for k, n := range ge.Buildings.counts {
		other.Buildings.counts[k] = n
	}
	other.recalculateRates()
	setAmount(other, "food", 1000)
	other.keepSharesLive()
	if !reflect.DeepEqual(other.Workers.GetAll(), ge.Workers.GetAll()) {
		t.Errorf("two equal games staffed differently: %v and %v", other.Workers.GetAll(), ge.Workers.GetAll())
	}
}

// The shares, auto-recruit and the wait survive a save and a load, the save
// still verifies, and a save from before them loads on auto with
// auto-recruit on and no new keys written.
func TestShares_SurviveSaveAndLoad(t *testing.T) {
	isolateAccountDir(t)
	ge := sharesEngine(t, 3, map[string]int{"gathering_camp": 2, "wood_camp": 2, "story_circle": 2})
	if _, err := ge.SetWorkerShare("knowledge", 40); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.SetWorkerShare("faith", 0); err != nil {
		t.Fatal(err)
	}
	ge.SetAutoRecruit(false)
	if err := ge.RecruitWorker("worker", 1); err != nil {
		t.Fatal(err)
	}
	if err := ge.SaveGame("shares-roundtrip"); err != nil {
		t.Fatal(err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("shares-roundtrip"); err != nil {
		t.Fatal(err)
	}
	if loaded.cheaterBadge {
		t.Error("the save failed its signature check")
	}
	if got, want := loaded.WorkerShares(), ge.WorkerShares(); !reflect.DeepEqual(got, want) {
		t.Errorf("shares after load = %v, want %v", got, want)
	}
	if loaded.AutoRecruit() {
		t.Error("auto-recruit came back on")
	}
	if loaded.staffHoldUntil != ge.staffHoldUntil || loaded.staffHoldUntil == 0 {
		t.Errorf("wait after load ends at tick %d, want %d", loaded.staffHoldUntil, ge.staffHoldUntil)
	}
	if st := loaded.GetState(); !reflect.DeepEqual(st.Workers.Shares, ge.WorkerShares()) || st.Workers.AutoRecruit {
		t.Errorf("state after load: shares %v, auto-recruit %v", st.Workers.Shares, st.Workers.AutoRecruit)
	}

	// The defaults write nothing new into a save.
	fresh := sharesEngine(t, 1, nil)
	if err := fresh.SaveGame("shares-default"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(savePath("shares-default"))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"worker_shares", "auto_recruit_off", "staff_hold_until"} {
		if strings.Contains(string(raw), `"`+key+`"`) {
			t.Errorf("a default game wrote %q", key)
		}
	}
	var gs GameSave
	if err := json.Unmarshal(raw, &gs); err != nil {
		t.Fatal(err)
	}
	if gs.WorkerShares != nil || gs.AutoRecruitOff || gs.StaffHoldUntil != 0 {
		t.Errorf("default save carries shares %v, auto-recruit off %v, wait %d", gs.WorkerShares, gs.AutoRecruitOff, gs.StaffHoldUntil)
	}
}

// A hand-edited save keeps only what the game could have written: known
// domains, percents from 0 to 100, a wait no longer than one command's.
func TestShares_LoadCleansSavedShares(t *testing.T) {
	got := cleanShares(map[string]float64{"food": 30, "nonsense": 50, "lumber": math.NaN(), "knowledge": 250, "faith": -5, "trade": 12.345})
	want := map[string]float64{"food": 30, "knowledge": 100, "faith": 0, "trade": 12.3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("cleanShares = %v, want %v", got, want)
	}
	if cleanShares(map[string]float64{"nonsense": 1}) != nil {
		t.Error("a map of unknown domains came back non-nil")
	}
}

// Prestige and Succumb put the shares back on auto and keep auto-recruit as
// the player left it; a new game resets both.
func TestShares_ResetByPrestigeSuccumbAndNewGame(t *testing.T) {
	setup := func() *GameEngine {
		ge := sharesEngine(t, 1, map[string]int{"story_circle": 1})
		if _, err := ge.SetWorkerShare("knowledge", 40); err != nil {
			t.Fatal(err)
		}
		ge.SetAutoRecruit(false)
		return ge
	}
	ge := setup()
	ge.mu.Lock()
	ge.completePrestige(prestigePlain)
	ge.mu.Unlock()
	if ge.WorkerShares() != nil || ge.AutoRecruit() {
		t.Errorf("after prestige: shares %v, auto-recruit %v; want auto, and auto-recruit still off", ge.WorkerShares(), ge.AutoRecruit())
	}
	ge = setup()
	succumbIn(t, ge, "iron_age")
	if ge.WorkerShares() != nil || ge.AutoRecruit() {
		t.Errorf("after Succumb: shares %v, auto-recruit %v; want auto, and auto-recruit still off", ge.WorkerShares(), ge.AutoRecruit())
	}
	ge = setup()
	ge.Reset()
	if ge.WorkerShares() != nil || !ge.AutoRecruit() || ge.staffHoldUntil != 0 {
		t.Errorf("after a new game: shares %v, auto-recruit %v, wait %d", ge.WorkerShares(), ge.AutoRecruit(), ge.staffHoldUntil)
	}
}

// ShareRows is what the Workers panel shows: auto domains split what the set
// shares leave by their slots, and set shares over 100% are scaled to fit.
func TestShares_ShareRows(t *testing.T) {
	ge := sharesEngine(t, 1, map[string]int{"gathering_camp": 2, "wood_camp": 2, "story_circle": 3})
	rows := func() map[string]ShareRow {
		out := map[string]ShareRow{}
		for _, r := range ShareRows(ge.GetState()) {
			out[r.Domain] = r
		}
		return out
	}
	r := rows()
	if len(r) != 3 || r["food"].Slots != 6 || r["knowledge"].Slots != 6 {
		t.Fatalf("rows = %+v", r)
	}
	for _, d := range []string{"food", "lumber", "knowledge"} {
		if r[d].Set || math.Abs(r[d].Percent-100.0/3) > 1e-9 {
			t.Errorf("%s on auto = %+v, want a third by slots", d, r[d])
		}
	}
	if _, err := ge.SetWorkerShare("knowledge", 50); err != nil {
		t.Fatal(err)
	}
	r = rows()
	if !r["knowledge"].Set || r["knowledge"].Percent != 50 || math.Abs(r["food"].Percent-25) > 1e-9 {
		t.Errorf("rows with knowledge at 50%%: %+v", r)
	}
	if _, err := ge.SetWorkerShare("food", 100); err != nil {
		t.Fatal(err)
	}
	r = rows()
	if math.Abs(r["food"].Percent-100.0/1.5) > 1e-9 || r["lumber"].Percent != 0 {
		t.Errorf("rows with 150%% set: %+v, want each set share scaled to fit and auto left nothing", r)
	}
	if _, ok := rows()["faith"]; ok {
		t.Error("a domain with no buildings and no share got a row")
	}
}

// The routine allocates nothing once every slot is filled: it runs every few
// ticks, and GetState and the tick have tight budgets.
func TestShares_SettledRunAllocatesNothing(t *testing.T) {
	ge := newLateGameEngine(t)
	ge.keepSharesLive()
	ge.keepSharesLive()
	if allocs := testing.AllocsPerRun(50, ge.keepSharesLive); allocs != 0 {
		t.Errorf("a settled routine run allocates %v times", allocs)
	}
}

// With shares set, a plan copy takes workers only from its own domain's
// superseded buildings, so the split holds. On auto it takes them from any
// superseded producer, as before.
func TestShares_PlanStaffingFollowsTheShares(t *testing.T) {
	game := func(shares bool) *GameEngine {
		ge := newSeededEngine(1)
		ge.Buildings.counts["stash"] = 50
		ge.Buildings.counts["hut"] = 3
		ge.Buildings.counts["wood_camp"] = 10
		ge.recalculateRates()
		setAmount(ge, "food", 1000)
		ge.SetAutoRecruit(false)
		if err := ge.RecruitWorker("worker", 30); err != nil {
			t.Fatal(err)
		}
		if err := ge.AssignWorker("wood_camp", 30); err != nil {
			t.Fatal(err)
		}
		if shares {
			if _, err := ge.SetWorkerShare("masonry", 50); err != nil {
				t.Fatal(err)
			}
		}
		ge.advanceAge("stone_age")
		ge.pendingCatastrophe = ""
		if !ge.Buildings.IsLegacy("wood_camp") {
			t.Fatal("setup: wood camps are not superseded in the Stone Age")
		}
		ge.Buildings.counts["stone_camp"]++
		ge.staffPlanCopy("stone_camp")
		return ge
	}
	if got := game(true).Workers.GetAssignedCount("worker", "stone_camp"); got != 0 {
		t.Errorf("with shares set the stone camp took %d workers from the wood camps (another domain)", got)
	}
	ge := game(false)
	if got := ge.Workers.GetAssignedCount("worker", "stone_camp"); got != slotsOf(ge, "stone_camp") {
		t.Errorf("on auto the stone camp has %d of its %d workers", got, slotsOf(ge, "stone_camp"))
	}
}
