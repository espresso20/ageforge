package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// paidByPrice is the share of a wonder's price that is paid, each part counted
// at what it is worth in the age's price levels (game.WonderBankShare says
// the same; this is worked out apart from it).
func paidByPrice(st game.GameState, wonder string) float64 {
	def, _ := st.Ruleset().Building(wonder)
	bank := st.Buildings[wonder].WonderBank
	need, paid := 0.0, 0.0
	for res, c := range def.BaseCost {
		level := st.Ruleset().DealPriceLevel(res, def.RequiredAge)
		need += c / level
		paid += min(bank[res], c) / level
	}
	return paid / need
}

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
	// What is paid: each part up to its price, each at its worth.
	share := paidByPrice(st, "colosseum")
	want := fmt.Sprintf("(%.0f%% banked)", float64(int(share*100)))
	if share >= 1 || want == "(100% banked)" {
		t.Fatalf("staging: %v of the price is paid", share)
	}

	text := wondersProvider(st, 100)
	if strings.Contains(untag(text), "100% banked") {
		t.Errorf("the list says 100%% banked with marble and iron ore missing:\n%s", untag(text))
	}
	if !strings.Contains(untag(text), want) {
		t.Errorf("the list does not say %s:\n%s", want, untag(text))
	}
	wantAll(t, "the list", text,
		"[red]marble: 0 / "+textfmt.Number(def.BaseCost["marble"]),
		"[red]iron ore: 0 / "+textfmt.Number(def.BaseCost["iron_ore"]),
		"[green]stone: 512K / "+textfmt.Number(def.BaseCost["stone"]),
		"[green]iron: 114K / "+textfmt.Number(def.BaseCost["iron"]),
		"[green]gold: 101K / "+textfmt.Number(def.BaseCost["gold"]),
	)
}

// A part banked past its price pays toward nothing else: a bank with only a
// surplus in one part reads the share that part is worth, not more. The worth
// is the part's price in the age's price levels, not its raw amount.
func TestWondersListCapsEachPartAtItsPrice(t *testing.T) {
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "iron_age")
	st0 := eng.GetState()
	def, _ := st0.Ruleset().Building("colosseum")
	eng.Buildings.LoadWonderBanks(map[string]map[string]float64{"colosseum": {"stone": def.BaseCost["stone"] * 10}})
	eng.StepTicks(1)
	st := eng.GetState()
	want := fmt.Sprintf("(%d%% banked)", int(paidByPrice(st, "colosseum")*100))
	if want == "(0% banked)" {
		t.Fatal("staging: the surplus of stone pays nothing")
	}
	if text := untag(wondersProvider(st, 100)); !strings.Contains(text, want) {
		t.Errorf("a surplus of stone alone should read %s:\n%s", want, text)
	}
}
