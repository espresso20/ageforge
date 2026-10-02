package smoke

import (
	"slices"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/ui"
)

// The fuzz corpus reaches the worker shares commands: the registry gives it
// `workers`, and the completions its walks follow lead into `workers share
// <domain> [percent|auto]` and `workers auto-recruit [on|off]`.
func TestFuzzCorpusCoversWorkerShares(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	ge := game.NewGameEngine()
	gen := newFuzzer(11, ge)
	found := false
	for _, c := range gen.commands {
		found = found || c == "workers"
	}
	if !found {
		t.Fatalf("the fuzz corpus has no workers command: %v", gen.commands)
	}
	comp := ui.NewAutoCompleter(ge)
	for line, want := range map[string][]string{
		"workers ":              {"workers share", "workers auto-recruit"},
		"workers share ":        {"workers share food", "workers share astronaut", "workers share auto"},
		"workers share food ":   {"workers share food auto"},
		"workers auto-recruit ": {"workers auto-recruit on", "workers auto-recruit off"},
	} {
		got := comp(line)
		for _, w := range want {
			if !slices.Contains(got, w) {
				t.Errorf("a fuzz walk from %q is offered %v, missing %q", line, got, w)
			}
		}
	}
	// And the lines it builds from them run without breaking anything.
	for _, line := range []string{"workers share food 30", "workers share knowledge 1e308", "workers share auto", "workers share military NaN", "workers auto-recruit off", "workers auto-recruit on"} {
		ui.HandleCommand(line, ge)
	}
	ge.StepTicks(50)
	for _, p := range invariantProblems(ge.GetState(), config.BuildingByKey()) {
		t.Errorf("after the worker shares commands: %s: %s", p.check, p.msg)
	}
}

// The idle style leaves its workers to worker shares: it never recruits or
// assigns by hand, and the game staffs its buildings between visits. With
// -no-shares it recruits and assigns by hand, with auto-recruit off.
func TestIdleBotLeavesWorkersToShares(t *testing.T) {
	if testing.Short() {
		t.Skip("plays ~20k ticks")
	}
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	play := func(noShares bool) (*runner, game.GameState) {
		cfg, err := ApplyStyle(DefaultConfig(), StyleIdle)
		if err != nil {
			t.Fatal(err)
		}
		cfg.MaxSim = 13 * time.Hour
		cfg.StopAge = "stone_age"
		cfg.NoShares = noShares
		ge := game.NewGameEngine()
		ge.SeedRNG(1)
		r := newRunner(cfg, 1, ge)
		r.play()
		return r, ge.GetState()
	}
	r, st := play(false)
	for _, kind := range []string{"recruit", "assign", "unassign"} {
		if n := r.bot.Actions[kind]; n > 0 {
			t.Errorf("the shares bot made %d %s actions by hand", n, kind)
		}
	}
	if st.Workers.TotalPop == 0 || !st.Workers.AutoRecruit {
		t.Errorf("shares bot: population %d, auto-recruit %v; want the game to have recruited", st.Workers.TotalPop, st.Workers.AutoRecruit)
	}
	r, st = play(true)
	if r.bot.Actions["recruit"] == 0 || r.bot.Actions["assign"] == 0 {
		t.Errorf("-no-shares bot: actions %v, want hand recruits and assigns", r.bot.Actions)
	}
	if st.Workers.AutoRecruit {
		t.Error("-no-shares left auto-recruit on")
	}
}
