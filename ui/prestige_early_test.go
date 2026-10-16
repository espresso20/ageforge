package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/rivo/tview"
)

// `prestige confirm` before the Modern Age says plainly what the early
// prestige pays and that a deeper run pays far more; from the Modern Age on
// it says nothing of the kind.
func TestPrestigeConfirmSaysAnEarlyPrestigePaysLittle(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	engine := game.NewGameEngine()
	engine.SeedRNG(1)
	for _, age := range []string{"stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age"} {
		if err := engine.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	confirm := untag(HandleCommand("prestige confirm", engine).Message)
	for _, want := range []string{
		"You will earn 9 prestige points.",
		"This is an early taste: a prestige from the Medieval Age pays 9 prestige points, for the 5 ages this run completed. " +
			"Going deeper pays far more: each era's ages are worth 3 times the era before, and a full run, 7 ages further on, pays 120 prestige points.",
	} {
		if !strings.Contains(confirm, want) {
			t.Errorf("prestige confirm in the Medieval Age does not say %q:\n%s", want, confirm)
		}
	}
	if strings.Contains(confirm, "Modern Age") {
		t.Errorf("prestige confirm names the Modern Age before the player has seen it:\n%s", confirm)
	}

	for _, age := range []string{"renaissance_age", "colonial_age", "industrial_age", "victorian_age", "electric_age", "atomic_age", "modern_age"} {
		if err := engine.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	confirm = untag(HandleCommand("prestige confirm", engine).Message)
	if strings.Contains(confirm, "early taste") || !strings.Contains(confirm, "You will earn 120 prestige points.") {
		t.Errorf("prestige confirm in the Modern Age is a full run's:\n%s", confirm)
	}
}

// A real engine's state, rather than a hand-written snapshot, drives every
// prestige surface. This catches a field that is present in types but missing
// from the engine snapshot. That production really scales is checked in game
// (the bonus truth tests).
func TestPrestigedStateShowsLastingBonus(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	engine := game.NewGameEngine()
	for _, age := range []string{"stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age"} {
		if err := engine.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	if err := engine.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	state := engine.GetState()
	want := fmt.Sprintf("x%.2f", state.Prestige.ProductionMultiplier)
	for _, got := range []string{untag(HandleCommand("prestige", engine).Message), untag(statsProvider(state, 0))} {
		if !strings.Contains(got, want) {
			t.Errorf("prestiged state omitted %q:\n%s", want, got)
		}
	}
	if got := untag(HandleCommand("prestige buy legacy_plan", engine).Message); !strings.Contains(got, "Usage:") {
		t.Errorf("retired prestige buy did not use the unknown-subcommand usage: %q", got)
	}

	pages := tview.NewPages()
	dashboard := NewDashboard(tview.NewApplication(), engine, pages)
	pages.AddPage("dashboard", dashboard.Root(), true, true)
	dashboard.inputField.SetText("stats")
	dashboard.submitInput()
	for _, size := range [][2]int{{80, 24}, {100, 30}} {
		dashboard.overlayMgr.entries["stats"].tv.ScrollTo(29, 0)
		if screen := screenAt(t, pages, size[0], size[1]); !strings.Contains(screen, want) {
			t.Errorf("%dx%d stats screen omitted %q:\n%s", size[0], size[1], want, screen)
		}
	}
}

func TestStatsShowsPrestigeFloorOnlyUntilPassed(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	engine := game.NewGameEngine()
	for _, age := range []string{"stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age"} {
		if err := engine.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	if err := engine.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	below := untag(statsProvider(engine.GetState(), 0))
	want := "Prestige opens once you pass the Medieval Age, the furthest an earlier run reached."
	if !strings.Contains(below, want) {
		t.Fatalf("stats omitted floor sentence %q:\n%s", want, below)
	}
	if err := engine.EnterAgeForTest("renaissance_age"); err != nil {
		t.Fatal(err)
	}
	past := untag(statsProvider(engine.GetState(), 0))
	if strings.Contains(past, want) {
		t.Fatalf("stats still shows floor sentence after passing it:\n%s", past)
	}
}

func TestPrestigeCommandExplainsFloor(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	engine := game.NewGameEngine()
	for _, age := range []string{"stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age"} {
		if err := engine.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	if err := engine.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	got := untag(HandleCommand("prestige confirm yes", engine).Message)
	if !strings.Contains(got, "Prestige opens once you pass the Medieval Age, the furthest an earlier run reached.") {
		t.Fatalf("prestige refusal omitted floor sentence:\n%s", got)
	}
}
