package ui

import (
	"fmt"
	"strings"
	"testing"
)

// A wonder's bank reads as what is paid toward it. The Iron Age save in the
// raw-materials tests had its Colosseum bank full at the old price (512K
// stone, 113.6K iron, 100.8K gold). The Colosseum now also asks marble and
// iron ore, so that bank is not full: the list must not say 100%, and the
// parts it lacks are the ones shown as missing.
func TestWondersListDoesNotSayFullWhenAPartIsMissing(t *testing.T) {
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "iron_age")
	st0 := eng.GetState()
	def, _ := st0.Ruleset().Building("colosseum")
	for _, res := range []string{"marble", "iron_ore"} {
		if def.BaseCost[res] <= 0 {
			t.Fatalf("staging: the Colosseum does not ask %s", res)
		}
	}
	eng.Buildings.LoadWonderBanks(map[string]map[string]float64{"colosseum": {"stone": 512000, "iron": 113600, "gold": 100800}})
	eng.StepTicks(1)
	st := eng.GetState()
	if st.Buildings["colosseum"].WonderBankFull {
		t.Fatal("staging: the bank reads full")
	}
	// What is paid: each part up to its price.
	need, paid := 0.0, 0.0
	for res, c := range def.BaseCost {
		need += c
		paid += min(st.Buildings["colosseum"].WonderBank[res], c)
	}
	want := fmt.Sprintf("(%.0f%% banked)", float64(int(paid/need*100)))
	if paid >= need || want == "(100% banked)" {
		t.Fatalf("staging: %v of %v is paid", paid, need)
	}

	text := wondersProvider(st, 100)
	if strings.Contains(untag(text), "100% banked") {
		t.Errorf("the list says 100%% banked with marble and iron ore missing:\n%s", untag(text))
	}
	if !strings.Contains(untag(text), want) {
		t.Errorf("the list does not say %s:\n%s", want, untag(text))
	}
	for _, line := range []string{
		"[red]marble: 0 / " + FormatNumber(def.BaseCost["marble"]),
		"[red]iron ore: 0 / " + FormatNumber(def.BaseCost["iron_ore"]),
		"[green]stone: 512K / " + FormatNumber(def.BaseCost["stone"]),
		"[green]iron: 114K / " + FormatNumber(def.BaseCost["iron"]),
		"[green]gold: 101K / " + FormatNumber(def.BaseCost["gold"]),
	} {
		if !strings.Contains(text, line) {
			t.Errorf("the list has no line %q:\n%s", line, text)
		}
	}
}

// A part banked past its price pays toward nothing else: a bank with only a
// surplus in one part reads the share that part is worth, not more.
func TestWondersListCapsEachPartAtItsPrice(t *testing.T) {
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "iron_age")
	st0 := eng.GetState()
	def, _ := st0.Ruleset().Building("colosseum")
	eng.Buildings.LoadWonderBanks(map[string]map[string]float64{"colosseum": {"stone": def.BaseCost["stone"] * 10}})
	eng.StepTicks(1)
	need := 0.0
	for _, c := range def.BaseCost {
		need += c
	}
	want := fmt.Sprintf("(%d%% banked)", int(def.BaseCost["stone"]/need*100))
	if text := untag(wondersProvider(eng.GetState(), 100)); !strings.Contains(text, want) {
		t.Errorf("a surplus of stone alone should read %s:\n%s", want, text)
	}
}
