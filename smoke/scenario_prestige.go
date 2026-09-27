package smoke

import (
	"fmt"
	"math"
	"runtime/debug"
	"strings"
	"time"

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
// ForceCatastropheForTest, ForceLastPassageForTest), never through play: a
// Succumb for a legacy bonus and ruins, a prestige from the Modern Age,
// upgrades bought and their effects measured against a twin engine that
// bought none, a second prestige, a succumbed Last Passage for the Cosmic
// Legacy, and more prestiges it must survive.

func runPrestige(e *Env, res *Result) {
	cfg := e.Base
	cfg.Cycles, cfg.MaxSim = 2, 400*time.Hour
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
		check(label+": legacy, ruins, upgrades and passive bonus carry over", len(probs) == 0, "%s", strings.Join(msgs, "; "))
		return before, after, true
	}
	if _, _, ok := prestige(up, "prestige 1"); !ok {
		return steps
	}
	if _, _, ok := prestige(ctl, "prestige 1 (twin)"); !ok {
		return steps
	}
	bought := map[string]int{}
	for _, k := range []string{"starting_food", "starting_wood", "population_cap", "storage_bonus", "tick_speed", "gather_boost"} {
		if up.BuyPrestigeUpgrade(k) == nil {
			bought[k]++
		}
	}
	check("buy upgrades", len(bought) > 0, "could not buy any upgrade with %d points", up.GetState().Prestige.Available)
	_, u2, ok1 := prestige(up, "prestige 2")
	_, c2, ok2 := prestige(ctl, "prestige 2 (twin)")
	if ok1 && ok2 {
		type effect struct {
			key       string
			got, want float64
		}
		// Everything lands at the reset: starting resources, caps and
		// rates are all in the first snapshot of the new run.
		effects := []effect{
			{"starting_food", u2.Resources["food"].Amount - c2.Resources["food"].Amount, 25 * float64(bought["starting_food"])},
			{"starting_wood", u2.Resources["wood"].Amount - c2.Resources["wood"].Amount, 25 * float64(bought["starting_wood"])},
			{"population_cap", float64(u2.Workers.MaxPop - c2.Workers.MaxPop), 2 * float64(bought["population_cap"])},
			{"storage_bonus", u2.Resources["food"].Storage - c2.Resources["food"].Storage, 20 * float64(bought["storage_bonus"])},
			{"tick_speed", u2.TickSpeedBonus - c2.TickSpeedBonus, 0.05 * float64(bought["tick_speed"])},
		}
		for _, ef := range effects {
			if bought[ef.key] == 0 {
				continue
			}
			check("upgrade "+ef.key+" applies", math.Abs(ef.got-ef.want) < 1e-6,
				"%s tier %d changed the fresh run by %.4g against the twin; the documented effect is %.4g", ef.key, bought[ef.key], ef.got, ef.want)
		}
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
			check("legacy bonuses and ruins survive every prestige", len(prestigeCarryProblems(before, a4, "plain")) == 0,
				"%v", prestigeCarryProblems(before, a4, "plain"))
		}
	}
	return steps
}
