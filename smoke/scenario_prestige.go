package smoke

import (
	"fmt"
	"math"
	"runtime/debug"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The prestige scenario has two halves.
//
// Played: the greedy bot plays for two prestige cycles. The runner checks
// every prestige it makes against the documented points formula and checks
// what must carry over (see checkPrestigeCarry). With today's pacing the bot
// may not reach the Modern Age inside the budget; the scenario then reports
// how far it got.
//
// Hooked: the same mechanics on demand, through the engine's exported test
// hooks (SummonHarbingerForTest to place the game in an age,
// ForceCatastropheForTest, ForceLastPassageForTest, EnterAgeForTest), never
// through play: a Succumb for a legacy bonus and ruins, a prestige from the
// Modern Age, the retired perks refused, each legacy kit item bought and
// checked, the shop refund from a signed level-5 save, a second prestige, a
// succumbed Last Passage for the Cosmic Legacy, and more prestiges it must
// survive.

func runPrestige(e *Env, res *Result) {
	cfg := e.Base
	// Two first-run-length cycles on the one-week curve (the second runs on
	// Era Mastery, with the legacy kit bought): 900 hours.
	cfg.Cycles, cfg.MaxSim = 2, 900*time.Hour
	sum := runBotSet(e, res, "prestige-played", "prestige", cfg, e.seeds(1))
	var played []string
	for _, r := range sum.Runs {
		for _, c := range r.Cycles {
			played = append(played, fmt.Sprintf("| %d | %d | %s | %s | %d | %d | %s |", r.Seed, c.Cycle, c.FinalAge, dur(c.Seconds), c.Points, c.Expected, orDefault(c.Ending, "plain")))
		}
		if len(r.Cycles) < cfg.Cycles {
			res.warn("prestige_unreached", "seed %d prestiged %d of %d times in %s simulated; it ended %s at %s",
				r.Seed, len(r.Cycles), cfg.Cycles, cfg.MaxSim, r.Outcome, r.FinalAge)
		}
	}
	res.section("Played cycles", "| seed | cycle | from age | 1x time | points | formula | ending |\n|---|---|---|---|---|---|---|\n%s",
		orDefault(strings.Join(played, "\n"), "| - | - | no prestige reached | - | - | - | - |"))

	steps := prestigeHooked(e, res)
	res.section("Hooked mechanics (test hooks, not play)", "| step | result |\n|---|---|\n%s", strings.Join(steps, "\n"))
	res.Summary = fmt.Sprintf("played: %d prestige(s) in %d seed(s); hooked: %d step(s)", sum.Runs[0].Stats.Prestiges, len(sum.Runs), len(steps))
}

// hookedEngine plays seed to 300 ticks into the Stone Age so there are
// buildings to ruin.
func hookedEngine(e *Env, seed int64) *game.GameEngine {
	cfg := e.Base
	cfg.MaxSim = 500 * time.Hour
	ge, _ := playUntil(cfg, seed, inAgeFor("stone_age", 300))
	return ge
}

func prestigeHooked(e *Env, res *Result) (steps []string) {
	seed := e.SeedBase
	repro := fmt.Sprintf("go run ./cmd/smoke -scenario prestige -seed-base %d -v", seed)
	check := func(name string, ok bool, format string, args ...interface{}) bool {
		if ok {
			steps = append(steps, fmt.Sprintf("| %s | ok |", name))
			return true
		}
		msg := fmt.Sprintf(format, args...)
		steps = append(steps, fmt.Sprintf("| %s | %s |", name, cell(msg)))
		f := res.fail("prestige_"+strings.ReplaceAll(name, " ", "_"), "%s", msg)
		f.Seed, f.Repro = seed, repro
		return false
	}
	defer func() {
		if rec := recover(); rec != nil {
			f := res.fail("prestige_panic", "hooked prestige mechanics panicked: %v", rec)
			f.Detail, f.Seed, f.Repro = string(debug.Stack()), seed, repro
		}
	}()

	up, ctl := hookedEngine(e, seed), hookedEngine(e, seed)
	twinSkip := func(p string) bool { return saveloadSkip(p) || p == "Stats.GameStarted" }
	if d := firstDiff(up.GetState(), ctl.GetState(), twinSkip); !check("twin engines agree", d == "", "two engines played from seed %d differ at %s", seed, d) {
		return steps
	}
	// Era Mastery twin check: the same state at mastery 1 (k = 2) in its
	// age runs every rate at exactly double and holds double the storage.
	{
		kb, kk := hookedEngine(e, seed), hookedEngine(e, seed)
		age := kb.GetState().Age
		kb.SetMasteryForTest(nil, age)
		kk.SetMasteryForTest(map[string]int{age: 1}, age)
		bs, ks := kb.GetState(), kk.GetState()
		var bad []string
		for _, r := range sortedKeys(bs.Resources) {
			b, k := bs.Resources[r], ks.Resources[r]
			if k.Rate != 2*b.Rate || k.Storage != 2*b.Storage {
				bad = append(bad, fmt.Sprintf("%s rate %g vs %g, storage %g vs %g", r, k.Rate, b.Rate, k.Storage, b.Storage))
			}
		}
		check("mastery twin: k = 2 doubles rates and storage", ks.Mastery.K == 2 && len(bad) == 0, "k %v; %s", ks.Mastery.K, strings.Join(bad, "; "))
	}

	both := func(f func(ge *game.GameEngine) error) error {
		if err := f(up); err != nil {
			return err
		}
		return f(ctl)
	}

	// A Succumb in the Iron epoch: a legacy bonus and ruins to carry.
	err := both(func(ge *game.GameEngine) error {
		_ = ge.SummonHarbingerForTest("iron_age") // places the game; a missing harbinger is fine
		if err := ge.ForceCatastropheForTest(); err != nil {
			return err
		}
		return ge.Succumb()
	})
	if !check("succumb in the iron epoch", err == nil, "%v", err) {
		return steps
	}
	st := up.GetState()
	ruins := 0
	for _, b := range st.Buildings {
		ruins += b.RuinCount
	}
	check("succumb leaves a legacy bonus", len(st.LegacyBonuses) > 0, "no legacy bonus after Succumb: %v", st.LegacyBonuses)
	check("succumb leaves ruins", ruins > 0, "no ruins after Succumb")

	// Prestige from the Modern Age, twice, with upgrades bought in between.
	prestige := func(ge *game.GameEngine, label string) (before, after game.GameState, ok bool) {
		_ = ge.SummonHarbingerForTest("modern_age")
		before = ge.GetState()
		want := PrestigePoints(before)
		if !check(label+": can prestige", before.Prestige.CanPrestige, "CanPrestige is false in %s", before.Age) {
			return before, before, false
		}
		check(label+": pending points match the formula", before.Prestige.PendingPoints == want,
			"prestige shows %d pending points, the documented formula gives %d (level %d)", before.Prestige.PendingPoints, want, before.Prestige.Level)
		if err := ge.DoPrestige(); !check(label+": prestige", err == nil, "%v", err) {
			return before, before, false
		}
		after = ge.GetState()
		check(label+": level", after.Prestige.Level == before.Prestige.Level+1, "level %d -> %d", before.Prestige.Level, after.Prestige.Level)
		check(label+": points paid", after.Prestige.TotalEarned-before.Prestige.TotalEarned == want,
			"paid %d points, formula %d", after.Prestige.TotalEarned-before.Prestige.TotalEarned, want)
		probs := prestigeCarryProblems(before, after, "plain")
		var msgs []string
		for _, p := range probs {
			msgs = append(msgs, p.msg)
		}
		check(label+": legacy, ruins, upgrades and mastery carry over", len(probs) == 0, "%s", strings.Join(msgs, "; "))
		return before, after, true
	}
	if _, _, ok := prestige(up, "prestige 1"); !ok {
		return steps
	}
	if _, _, ok := prestige(ctl, "prestige 1 (twin)"); !ok {
		return steps
	}
	// The first shop's perks are retired: none can be bought, and the legacy
	// kit sells instead (prestigeKitHooked).
	var sold []string
	for _, def := range config.PrestigeUpgrades() {
		if def.Retired && up.BuyPrestigeUpgrade(def.Key) == nil {
			sold = append(sold, def.Key)
		}
	}
	check("retired perks can't be bought", len(sold) == 0, "the shop sold retired perks: %s", strings.Join(sold, ", "))
	prestigeKitHooked(e, check)
	prestigeRefundHooked(check)
	_, _, ok1 := prestige(up, "prestige 2")
	_, _, ok2 := prestige(ctl, "prestige 2 (twin)")
	if !ok1 || !ok2 {
		return steps
	}

	// The Cosmic Legacy: a succumbed Last Passage, then prestiges it survives.
	lp := firstLastPassageAge()
	if err := up.ForceLastPassageForTest(lp); !check("last passage comes in "+lp, err == nil, "%v", err) {
		return steps
	}
	before := up.GetState()
	if err := up.SuccumbLastPassage(); !check("succumb the last passage", err == nil, "%v", err) {
		return steps
	}
	after := up.GetState()
	check("cosmic legacy granted", after.LastPassage.CosmicLegacy, "no Cosmic Legacy after a succumbed Last Passage")
	check("succumbed passage pays nothing", after.Prestige.TotalEarned == before.Prestige.TotalEarned,
		"a succumbed Last Passage paid %d points", after.Prestige.TotalEarned-before.Prestige.TotalEarned)
	check("succumbed passage completes the prestige", after.Prestige.Level == before.Prestige.Level+1, "level %d -> %d", before.Prestige.Level, after.Prestige.Level)
	_, a3, ok := prestige(up, "prestige after the legacy")
	if ok {
		check("cosmic legacy survives prestige", a3.LastPassage.CosmicLegacy, "the Cosmic Legacy was lost on the next prestige")
	}
	if err := up.ForceLastPassageForTest(lp); err == nil {
		b4 := up.GetState()
		check("a second legacy is refused", up.SuccumbLastPassage() != nil, "Succumb was accepted while already holding the Cosmic Legacy")
		if err := up.EndureLastPassage(); check("endure the last passage", err == nil, "%v", err) {
			a4 := up.GetState()
			paid, full := a4.Prestige.TotalEarned-b4.Prestige.TotalEarned, PrestigePoints(b4)
			want := int(math.Floor(float64(full) * game.LastPassageKeepFor(b4.LastPassage.BraceLevel)))
			check("endured passage pays its share", paid == want, "an endured Last Passage paid %d of %d points; the documented share is %d", paid, full, want)
			check("cosmic legacy survives an endured passage", a4.LastPassage.CosmicLegacy, "the Cosmic Legacy was lost")
			// Several prestiges apart, so mastery has moved more than one
			// level: only what must never be lost is compared.
			var lost []problem
			for _, p := range prestigeCarryProblems(before, a4, "plain") {
				if p.check != "prestige_mastery" {
					lost = append(lost, p)
				}
			}
			check("legacy bonuses and ruins survive every prestige", len(lost) == 0, "%v", lost)
		}
	}
	return steps
}

// prestigeKitHooked buys each legacy kit item after a Modern Age prestige
// and checks it does what the shop says: a run that planned a Stone Age
// building and an advance, set a worker share and met a civilization;
// then, in the next run, the shares are set again, entering the Stone Age
// adds its template slice, and entering the Bronze Age meets the
// civilization again. A tech the player planned comes along with the
// template too (their own research path; the kit picks none by itself).
func prestigeKitHooked(e *Env, check func(string, bool, string, ...interface{}) bool) {
	ge := hookedEngine(e, e.SeedBase)
	st := ge.GetState()
	planned := ""
	for _, k := range sortedKeys(config.BuildingByKey()) {
		d := config.BuildingByKey()[k]
		if d.RequiredAge == st.Age && d.Category == "production" && d.MaxCount == 0 && st.Buildings[k].Unlocked {
			planned = k
			break
		}
	}
	if !check("kit: a building to plan in "+st.Age, planned != "", "no production building is open in %s", st.Age) {
		return
	}
	if _, err := ge.PlanAddBuild(planned, 2); !check("kit: plan a build", err == nil, "%v", err) {
		return
	}
	if err := ge.PlanAddAdvance(); !check("kit: plan an advance", err == nil, "%v", err) {
		return
	}
	if _, err := ge.SetWorkerShare("food", 40); !check("kit: set a worker share", err == nil, "%v", err) {
		return
	}
	if err := ge.MeetFactionForTest("riverlands_tribes", 50); !check("kit: meet a civilization", err == nil, "%v", err) {
		return
	}
	_ = ge.SummonHarbingerForTest(game.PrestigeRunAge)
	if err := ge.DoPrestige(); !check("kit: prestige from "+game.PrestigeRunAge, err == nil, "%v", err) {
		return
	}
	pts := ge.GetState().Prestige.Available
	for _, key := range config.LegacyKit() {
		err := ge.BuyPrestigeUpgrade(key)
		check("kit: buy "+key, err == nil, "%v (with %d points)", err, pts)
	}
	after := ge.GetState()
	total := 0
	for _, key := range config.LegacyKit() {
		total += config.PrestigeUpgradeByKey()[key].Costs[0]
	}
	check("kit: the whole kit costs 99 points", total == 99 && after.Prestige.Available == pts-total, "the kit cost %d; %d points left of %d", total, after.Prestige.Available, pts)
	check("kit: worker shares carry over", after.Workers.Shares["food"] == 40, "shares after buying Worker Shares: %v", after.Workers.Shares)

	if err := ge.EnterAgeForTest(st.Age); !check("kit: enter "+st.Age, err == nil, "%v", err) {
		return
	}
	var kinds []string
	for _, it := range ge.GetState().Plan {
		kinds = append(kinds, it.Kind+":"+it.Key)
	}
	got := strings.Join(kinds, ",")
	check("kit: the template's slice is added on entering its age", strings.Contains(got, "build:"+planned) && strings.Contains(got, "advance:"),
		"the plan in %s holds [%s], want %s and an advance", st.Age, got, planned)
	if err := ge.EnterAgeForTest("bronze_age"); !check("kit: enter bronze_age", err == nil, "%v", err) {
		return
	}
	f := ge.GetState().Diplomacy.Factions["riverlands_tribes"]
	check("kit: old friends are met again at their age", f.Discovered && f.Opinion == 0,
		"Riverlands Tribes in the Bronze Age: met %v, opinion %d (want met at 0)", f.Discovered, f.Opinion)

	// The player's own research path: a tech planned in the first age is
	// planned again when the next run starts there, and nothing else is.
	tech := ""
	for _, t := range config.Technologies() {
		if t.Age == config.AgeOrder()[0] && len(t.Prerequisites) == 0 {
			tech = t.Key
			break
		}
	}
	own := game.NewGameEngine()
	own.SeedRNG(e.SeedBase)
	if err := own.PlanAddResearch(tech); !check("kit: plan a tech", err == nil, "%v", err) {
		return
	}
	_ = own.SummonHarbingerForTest(game.PrestigeRunAge)
	if err := own.DoPrestige(); !check("kit: prestige with a planned tech", err == nil, "%v", err) {
		return
	}
	if err := own.BuyPrestigeUpgrade(config.LegacyPlan); !check("kit: buy the template", err == nil, "%v", err) {
		return
	}
	var research []string
	for _, it := range own.GetState().Plan {
		if it.Kind == game.PlanResearch {
			research = append(research, it.Key)
		}
	}
	check("kit: a planned tech comes along with the template", len(research) == 1 && research[0] == tech,
		"the new run's plan holds the techs %v, want only %s (the one the player planned)", research, tech)
	check("kit: nothing researches by itself", own.GetState().Research.CurrentTech == "", "research started by itself: %s", own.GetState().Research.CurrentTech)
}

// prestigeRefundHooked loads a signed level-5 save from the first shop (the
// plan's worked example) and checks the refund: 600 points, the old tiers at
// 0 with their keys kept, the shop at its version, once.
func prestigeRefundHooked(check func(string, bool, string, ...interface{}) bool) {
	upgrades := map[string]int{
		"gather_boost": 3, "storage_bonus": 3, "research_speed": 3, "military_power": 3,
		"starting_food": 5, "starting_wood": 4, "population_cap": 3, "expedition_loot": 3,
	}
	src := game.NewGameEngine()
	src.SeedRNG(5)
	if err := src.WriteShopV1SaveForTest("level5", 5, 86, 3, upgrades); !check("refund: write a level-5 save", err == nil, "%v", err) {
		return
	}
	for i, name := range []string{"first load", "second load"} {
		ge := game.NewGameEngine()
		if err := ge.LoadGame("level5"); !check("refund: "+name, err == nil, "%v", err) {
			return
		}
		p := ge.GetState().Prestige
		var held []string
		for key := range upgrades {
			if p.Upgrades[key].Tier != 0 {
				held = append(held, key)
			}
		}
		check("refund: "+name+" gives 600 points", p.Available == 600 && p.TotalEarned == 600 && len(held) == 0 && p.ShopVersion == config.PrestigeShopVersion,
			"available %d, total %d, old tiers still held %v, shop version %d", p.Available, p.TotalEarned, held, p.ShopVersion)
		if i == 0 {
			if err := ge.SaveGame("level5"); !check("refund: save after the refund", err == nil, "%v", err) {
				return
			}
		}
	}
}
