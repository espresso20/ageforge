package game

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// An early prestige (before the Modern Age) pays little. The run's last lines
// say plainly what it paid and that a deeper run pays far more, without
// naming an age the player has not seen.
func TestEarlyPrestigeSaysWhatItPaid(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	prestigeFrom := func(age string, seen ...string) *GameEngine {
		ge := newSeededEngine(8)
		for _, a := range ageKeys() {
			ge.Prestige.NoteAgeEntered(a)
			ge.Stats.AgesReached = append(ge.Stats.AgesReached, a)
			if a == age {
				break
			}
		}
		ge.Stats.AgesReached = append(ge.Stats.AgesReached, seen...)
		ge.age, ge.currentEpoch = age, config.EpochForAge(age)
		if err := ge.DoPrestige(); err != nil {
			t.Fatalf("prestige from %s: %v", age, err)
		}
		return ge
	}

	// A first run: the Modern Age has not been seen, so it is not named.
	ge := prestigeFrom("medieval_age")
	want := "That was an early prestige, from the Medieval Age: it paid 9 prestige points for the 5 ages the run completed. " +
		"Going deeper pays far more: each era's ages are worth 3 times the era before, and a full run, 7 ages further on, pays 120 prestige points."
	done, early := indexOfLog(ge, "Prestige complete. Level 1, 9 prestige points earned."), indexOfLog(ge, want)
	if done < 0 || early != done+1 {
		t.Errorf("the log after a Medieval Age prestige should carry, right after \"Prestige complete\":\n%s\ngot:\n%s", want, strings.Join(logMessages(ge), "\n"))
	}
	if indexOfLog(ge, "Modern Age") >= 0 {
		t.Errorf("a first run's early prestige names the Modern Age, which the player has not seen:\n%s", strings.Join(logMessages(ge), "\n"))
	}

	// A player who has seen the Modern Age is told it by name.
	ge = prestigeFrom("atomic_age", "modern_age")
	want = "That was an early prestige, from the Atomic Age: it paid 93 prestige points for the 11 ages the run completed. " +
		"Going deeper pays far more: each era's ages are worth 3 times the era before, and a run to the Modern Age pays 120 prestige points."
	if indexOfLog(ge, want) < 0 {
		t.Errorf("the log after an Atomic Age prestige should carry:\n%s\ngot:\n%s", want, strings.Join(logMessages(ge), "\n"))
	}

	// A full run carries no such line.
	ge = prestigeFrom("modern_age")
	if indexOfLog(ge, "early prestige") >= 0 || indexOfLog(ge, "Prestige complete. Level 1, 120 prestige points earned.") < 0 {
		t.Errorf("a Modern Age prestige is a full run:\n%s", strings.Join(logMessages(ge), "\n"))
	}

	// The line itself, both tenses, and nothing from the Modern Age on.
	sight := newAgeSight("medieval_age", nil, "")
	if got, want := EarlyPrestigeLine(sight, "medieval_age", 9, false),
		"This is an early prestige, from the Medieval Age: it pays 9 prestige points for the 5 ages this run completed. "+
			"Going deeper pays far more: each era's ages are worth 3 times the era before, and a full run, 7 ages further on, pays 120 prestige points."; got != want {
		t.Errorf("EarlyPrestigeLine before the prestige:\n got %q\nwant %q", got, want)
	}
	for _, a := range ageKeys()[ageOrders()[PrestigeRunAge]:] {
		if got := EarlyPrestigeLine(sight, a, 1, true); got != "" {
			t.Errorf("EarlyPrestigeLine(%s) = %q, want none", a, got)
		}
	}
	// What it quotes is what the engine pays.
	if got := config.DepthPoints("medieval_age"); got != 9 {
		t.Errorf("a Medieval Age prestige pays %d, the text says 9", got)
	}
	if got := config.DepthPoints(PrestigeRunAge); got != 120 {
		t.Errorf("a Modern Age prestige pays %d, the text says 120", got)
	}
	if w := config.DepthWeight("iron_age") / config.DepthWeight("bronze_age"); w != 3 {
		t.Errorf("an era's ages are worth %d times the era before, the text says 3", w)
	}
}
