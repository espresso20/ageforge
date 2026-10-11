package smoke

import (
	"slices"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/ui"
)

// The fuzz corpus reaches the roster and the workers commands: the registry
// gives it `roster` and `workers`, and the completions its walks follow lead
// into `roster <domain> [percent|auto]` and `workers auto-recruit [on|off]`.
// The roster's old name (`workers share`) is not offered, and still runs.
func TestFuzzCorpusCoversWorkerShares(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	ge := game.NewGameEngine()
	gen := newFuzzer(11, ge)
	for _, want := range []string{"workers", "roster"} {
		if !slices.Contains(gen.commands, want) {
			t.Fatalf("the fuzz corpus has no %s command: %v", want, gen.commands)
		}
	}
	comp := ui.NewAutoCompleter(ge)
	for line, want := range map[string][]string{
		"workers ":              {"workers auto-recruit"},
		"roster ":               {"roster food", "roster astronaut", "roster auto"},
		"roster food ":          {"roster food auto"},
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
	for _, line := range []string{"roster food 30", "roster knowledge 1e308", "roster auto", "roster military NaN",
		"workers share food 30", "workers share knowledge 1e308", "workers share auto", "workers share military NaN", "workers auto-recruit off", "workers auto-recruit on"} {
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
