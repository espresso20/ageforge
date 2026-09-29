package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// newWonderTestEngine is newCaseTestEngine with food set to food: in the
// Primitive Age the Sacred Grove needs 500 food and 1,000 wood; in the Bronze
// Age Stonehenge needs no food.
func newWonderTestEngine(t *testing.T, age string, food float64) *game.GameEngine {
	t.Helper()
	ge := newCaseTestEngine(t, age)
	ge.Resources.LoadAmounts(map[string]float64{"food": food})
	return ge
}

// TestWonderBankCommand drives every deposit form through HandleCommand
// (card AWr6Nu6U). `wonder bank food all` used to fall through to the bank
// status screen and deposit nothing, since only `collect` was recognised;
// each form must now either deposit, saying how much, or say exactly why not.
func TestWonderBankCommand(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	const bronze = "bronze_age"
	forms := []string{"wonder bank food all", "wonder bank food 100", "wonder bank food max", "wonder bank food", "wonder collect food all"}
	cases := []struct {
		name  string
		age   string
		food  float64
		setup func(t *testing.T, ge *game.GameEngine)
		want  map[string]float64 // form → food banked; absent: refused
		err   string             // what a refusal must say
	}{
		{name: "needs food", food: 50000,
			want: map[string]float64{"wonder bank food all": 500, "wonder bank food 100": 100, "wonder bank food max": 500, "wonder bank food": 500, "wonder collect food all": 500}},
		{name: "some food", food: 60,
			want: map[string]float64{"wonder bank food all": 60, "wonder bank food max": 60, "wonder bank food": 60, "wonder collect food all": 60},
			err:  "not enough food: need 100, have 60"},
		{name: "no food", food: 0, err: "you have no food to bank"},
		{name: "wonder doesn't need food", age: bronze, food: 50000, err: "Stonehenge doesn't need food (it needs iron, stone, wood)"},
		{name: "food already banked", food: 50000, err: "Sacred Grove already has all the food it needs",
			setup: func(t *testing.T, ge *game.GameEngine) {
				if res := HandleCommand("wonder bank food all", ge); res.Type != "success" {
					t.Fatalf("setup: %+v", res)
				}
			}},
		{name: "wonder built", food: 50000, err: "Sacred Grove is already built: there is nothing left to bank this age.",
			setup: func(t *testing.T, ge *game.GameEngine) {
				if msg := game.DevExecCommand("/build sacred_grove", ge); msg != "built sacred_grove" {
					t.Fatalf("setup: %q", msg)
				}
			}},
	}
	for _, c := range cases {
		for _, form := range forms {
			t.Run(c.name+"/"+form, func(t *testing.T) {
				ge := newWonderTestEngine(t, c.age, c.food)
				if c.setup != nil {
					c.setup(t, ge)
				}
				before := ge.GetState()
				res := HandleCommand(form, ge)
				after := ge.GetState()
				key := "sacred_grove"
				if c.age == bronze {
					key = "stonehenge"
				}
				gotBanked := after.Buildings[key].WonderBank["food"] - before.Buildings[key].WonderBank["food"]
				spent := before.Resources["food"].Amount - after.Resources["food"].Amount
				want, ok := c.want[form]
				if !ok {
					if res.Type != "error" || !strings.Contains(res.Message, c.err) {
						t.Errorf("%q = %+v, want an error saying %q", form, res, c.err)
					}
					if gotBanked != 0 || spent != 0 {
						t.Errorf("%q was refused but banked %v food and spent %v", form, gotBanked, spent)
					}
					return
				}
				if res.Type != "success" {
					t.Fatalf("%q = %+v, want a deposit", form, res)
				}
				if gotBanked != want || spent != want {
					t.Errorf("%q banked %v food and spent %v, want %v", form, gotBanked, spent, want)
				}
				if prefix := "Banked " + FormatNumber(want) + " food into Sacred Grove"; !strings.HasPrefix(res.Message, prefix) {
					t.Errorf("%q = %q, want it to start %q", form, res.Message, prefix)
				}
			})
		}
	}
}

// TestWonderBankAllResources is `wonder bank all`: every resource the wonder
// still needs, each as far as what is on hand goes, naming what was skipped.
func TestWonderBankAllResources(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()

	ge := newWonderTestEngine(t, "", 0)
	res := HandleCommand("wonder bank all", ge)
	if res.Type != "success" || !strings.HasPrefix(res.Message, "Banked 1.00K wood into Sacred Grove. Not banked: you have no food to bank.") {
		t.Errorf("no food: %+v", res)
	}
	if res := HandleCommand("wonder bank all", ge); res.Type != "error" ||
		res.Message != "Nothing banked: you have no food to bank; Sacred Grove already has all the wood it needs." {
		t.Errorf("again: %+v", res)
	}
	ge.Resources.LoadAmounts(map[string]float64{"food": 800})
	res = HandleCommand("wonder bank all max", ge)
	if res.Type != "success" || !strings.Contains(res.Message, "Banked 500 food into Sacred Grove.") || !strings.Contains(res.Message, "Bank full! Type 'build sacred_grove'") {
		t.Errorf("rest of the food: %+v", res)
	}
	if got := ge.GetState().Resources["food"].Amount; got != 300 {
		t.Errorf("food left = %v, want 300", got)
	}

	bronze := newWonderTestEngine(t, "bronze_age", 50000)
	if res := HandleCommand("wonder bank all", bronze); res.Type != "success" || strings.Contains(res.Message, "food") {
		t.Errorf("Stonehenge takes no food, and banking all shouldn't mention it: %+v", res)
	}
}

// TestWonderBankRefusals covers the malformed forms: each is an error that
// says what to type, and none shows the status screen instead.
func TestWonderBankRefusals(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	for _, c := range []struct{ cmd, want string }{
		{"wonder bank", wonderCollectUsage},
		{"wonder bank food 1 2", wonderCollectUsage},
		{"wonder bank all 100", "give an amount for one resource at a time"},
		{"wonder bank food lots", "the amount must be a positive number"},
		{"wonder bank food -5", "the amount must be a positive number"},
		{"wonder bank unobtainium all", "Unknown resource: unobtainium"},
		{"wonder deposit food all", `Unknown wonder command "deposit"`},
	} {
		ge := newWonderTestEngine(t, "", 50000)
		if res := HandleCommand(c.cmd, ge); res.Type != "error" || !strings.Contains(res.Message, c.want) {
			t.Errorf("%q = %+v, want an error saying %q", c.cmd, res, c.want)
		}
	}
}

// TestWonderBankCompletion: the registry takes every form, so Enter runs it
// as typed, and completion offers the resources the wonder still needs and
// the amount words.
func TestWonderBankCompletion(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	ge := newWonderTestEngine(t, "", 50000)
	c := newCompleter(ge, nil)
	for _, line := range []string{"wonder bank food all", "wonder bank food 100", "wonder bank food max", "wonder bank food", "wonder bank all", "wonder collect food all"} {
		if !c.complete(line) {
			t.Errorf("%q is not a whole command to the registry", line)
		}
	}
	complete := NewAutoCompleter(ge)
	if got := strings.Join(complete("wonder bank "), ","); got != "wonder bank food,wonder bank wood,wonder bank all" {
		t.Errorf("wonder bank completes to %s", got)
	}
	if got := strings.Join(complete("wonder bank food "), ","); got != "wonder bank food all,wonder bank food max" {
		t.Errorf("wonder bank food completes to %s", got)
	}
	HandleCommand("wonder bank food", ge)
	if got := strings.Join(NewAutoCompleter(ge)("wonder bank "), ","); got != "wonder bank wood,wonder bank all" {
		t.Errorf("with the food banked, wonder bank completes to %s", got)
	}
	bronze := NewAutoCompleter(newWonderTestEngine(t, "bronze_age", 50000))
	if got := strings.Join(bronze("wonder bank "), ","); got != "wonder bank iron,wonder bank stone,wonder bank wood,wonder bank all" {
		t.Errorf("in the Bronze Age, wonder bank completes to %s", got)
	}
}
