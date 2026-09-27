package game

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/flavor"
)

// LoadGame used to call SeedRNG(save.Seed), so every load restarted the
// random stream from the run's first draw (the smoke suite's
// known_bug_rng_restarts_on_load). Saves now carry each stream's position.

// TestCountingSourceMatchesStockSource: the counter must not change a single
// draw, or every seed's stream (and every seed-pinned test) would shift.
func TestCountingSourceMatchesStockSource(t *testing.T) {
	for _, seed := range []int64{0, 1, 42, -7, 1 << 40} {
		stock := rand.New(rand.NewSource(seed))
		src := newCountingSource(seed)
		counted := rand.New(src)
		for i := 0; i < 200; i++ {
			if a, b := stock.Intn(1000), counted.Intn(1000); a != b {
				t.Fatalf("seed %d draw %d: Intn %d vs %d", seed, i, a, b)
			}
			if a, b := stock.Float64(), counted.Float64(); a != b {
				t.Fatalf("seed %d draw %d: Float64 %v vs %v", seed, i, a, b)
			}
			if a, b := stock.Uint64(), counted.Uint64(); a != b {
				t.Fatalf("seed %d draw %d: Uint64 %v vs %v", seed, i, a, b)
			}
		}
		if a, b := stock.Perm(20), counted.Perm(20); !slices.Equal(a, b) {
			t.Fatalf("seed %d: Perm %v vs %v", seed, a, b)
		}
		if src.draws == 0 {
			t.Fatalf("seed %d: no draws counted", seed)
		}
	}
}

// TestCountingSourceSkip: skipping n steps lands where n draws would.
func TestCountingSourceSkip(t *testing.T) {
	a, b := newCountingSource(9), newCountingSource(9)
	ra := rand.New(a)
	for i := 0; i < 137; i++ {
		ra.Intn(50)
		ra.Float64()
	}
	b.skip(a.draws)
	if a.draws != b.draws {
		t.Fatalf("draws %d vs %d", a.draws, b.draws)
	}
	rb := rand.New(b)
	for i := 0; i < 50; i++ {
		if x, y := ra.Int63(), rb.Int63(); x != y {
			t.Fatalf("after skip, draw %d: %d vs %d", i, x, y)
		}
	}
}

// TestLoadGame_ContinuesRNGStream plays the determinism scenario (events,
// expeditions, raids and encounters all rolling dice), saves part way, loads
// into a fresh engine and plays both on: the loaded game must match the
// uninterrupted one to the bit, log included.
func TestLoadGame_ContinuesRNGStream(t *testing.T) {
	isolateAccountDir(t)
	const split = 1200
	a, sc := determinismSetup(t, 42)
	sc.play(a, 0, split)
	if err := a.SaveGame("rng-continuity"); err != nil {
		t.Fatal(err)
	}
	b := NewGameEngine()
	if err := b.LoadGame("rng-continuity"); err != nil {
		t.Fatal(err)
	}
	if ga, gb := a.GetState(), b.GetState(); ga.RNGDraws != gb.RNGDraws || ga.QuipDraws != gb.QuipDraws {
		t.Fatalf("stream positions after load: gameplay %d vs %d, quip %d vs %d", ga.RNGDraws, gb.RNGDraws, ga.QuipDraws, gb.QuipDraws)
	}
	if a.GetState().RNGDraws == 0 {
		t.Fatal("the scenario drew nothing before the save; the test would pass vacuously")
	}
	// The log is not saved, so only lines written after the save can match.
	saved := a.GetState().Tick
	after := func(transcript string) string {
		var out strings.Builder
		for _, line := range strings.SplitAfter(transcript, "\n") {
			var tick int
			if _, err := fmt.Sscanf(line, "log %d", &tick); err == nil && tick > saved {
				out.WriteString(line)
			}
		}
		return out.String()
	}
	ta := after(sc.play(a, split, determinismTicks))
	tb := after(sc.play(b, split, determinismTicks))
	if ta == "" {
		t.Fatal("no log lines after the save; the comparison would be vacuous")
	}
	fa, fb := determinismFingerprint(a)+ta, determinismFingerprint(b)+tb
	if fa != fb {
		t.Fatalf("loaded game diverged from the uninterrupted one:\n%s", firstDiff(fa, fb))
	}
}

// TestRestoreRNG covers the edges: an old save with no positions restarts
// both streams from the seed, and a position past the replay cap (an edited
// save) does too instead of hanging the load.
func TestRestoreRNG(t *testing.T) {
	fresh := newSeededEngine(5)
	want := fresh.rng.Int63()

	ge := newSeededEngine(99)
	if !ge.restoreRNG(5, 0, 0) {
		t.Fatal("restoreRNG with zero positions reported failure")
	}
	if got := ge.rng.Int63(); got != want {
		t.Errorf("zero positions: first draw %d, want the seed's first draw %d", got, want)
	}

	ge = newSeededEngine(99)
	if ge.restoreRNG(5, maxRNGReplay+1, 0) {
		t.Error("restoreRNG past the cap reported success")
	}
	if got := ge.rng.Int63(); got != want {
		t.Errorf("past the cap: first draw %d, want the seed's first draw %d", got, want)
	}
}

// TestLoadGame_RestoresProseMemory: how often a flavour line is redrawn
// depends on the Stream's memory of recent lines, and redraws spend gameplay
// draws, so the memory rides in the save with the stream position.
func TestLoadGame_RestoresProseMemory(t *testing.T) {
	isolateAccountDir(t)
	a := newSeededEngine(3)
	a.StepTicks(1) // a real game has ticked: the epoch's harbinger check has run
	req := flavor.Request{Moment: flavor.ExpeditionSuccess, Age: "iron_age", Subject: "the hills"}
	for i := 0; i < 20; i++ {
		a.flavorStream().Line(req, a.gameRNG())
	}
	if err := a.SaveGame("prose"); err != nil {
		t.Fatal(err)
	}
	b := NewGameEngine()
	if err := b.LoadGame("prose"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 40; i++ {
		la, lb := a.flavorStream().Line(req, a.gameRNG()), b.flavorStream().Line(req, b.gameRNG())
		if la != lb {
			t.Fatalf("line %d after load differs:\n  live:   %s\n  loaded: %s", i, la, lb)
		}
	}
	if da, db := a.rngSrc.draws, b.rngSrc.draws; da != db {
		t.Errorf("draws after 40 lines: live %d, loaded %d", da, db)
	}
}
