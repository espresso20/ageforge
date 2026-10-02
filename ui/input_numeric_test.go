package ui

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// hostileNumbers are the numeric arguments the smoke fuzzer found (or could
// find) trouble with: integer overflow, negatives, NaN and infinities.
var hostileNumbers = []string{
	"9223372036854775807", "-9223372036854775808", "18446744073709551616",
	"-1", "0", "-0", "NaN", "nan", "Inf", "+Inf", "-Inf", "1e308", "1e309", "-1e308",
	"1000001", "0x1F", "1_000", "3.5",
}

// TestRecruitMaxIntRejected is the smoke fuzzer's repro: this recruit used to
// overflow the pop-cap check and leave the population at about -9.2e18.
func TestRecruitMaxIntRejected(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	res := HandleCommand("recruit 9223372036854775807", ge)
	if res.Type != "error" {
		t.Fatalf("recruit MaxInt: got %q (%s), want an error", res.Type, res.Message)
	}
	if !strings.Contains(res.Message, "1 to 1000000") {
		t.Errorf("recruit MaxInt error %q does not say what a valid count is", res.Message)
	}
	if pop := ge.GetState().Workers.TotalPop; pop != 0 {
		t.Errorf("population = %d after a refused recruit, want 0", pop)
	}
}

// TestHostileNumericArguments types every hostile number into every command
// that takes a count or an amount and checks the game is still sane after
// each one: population in [0, cap], no negative assignments, every resource
// finite and non-negative.
func TestHostileNumericArguments(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	for _, c := range []string{"gather food 25", "gather wood 25", "build hut", "recruit max", "assign gathering_camp"} {
		HandleCommand(c, ge)
	}
	ge.StepTicks(5)

	templates := []string{
		"recruit %s", "assign gathering_camp %s", "unassign gathering_camp %s",
		"dismiss gathering_camp %s", "sell hut %s", "build hut %s", "upgrade hut %s",
		"gather food %s", "trade food wood %s", "wonder collect food %s",
	}
	for _, tpl := range templates {
		for _, n := range hostileNumbers {
			cmd := strings.Replace(tpl, "%s", n, 1)
			HandleCommand(cmd, ge)
			ge.StepTicks(1)
			st := ge.GetState()
			w := st.Workers
			if w.TotalPop < 0 || w.TotalPop > w.MaxPop && w.MaxPop > 0 {
				t.Fatalf("after %q: population %d outside [0, %d]", cmd, w.TotalPop, w.MaxPop)
			}
			if w.TotalIdle < 0 || w.TotalIdle > w.TotalPop {
				t.Fatalf("after %q: idle %d outside [0, %d]", cmd, w.TotalIdle, w.TotalPop)
			}
			for key, rs := range st.Resources {
				if math.IsNaN(rs.Amount) || math.IsInf(rs.Amount, 0) || rs.Amount < 0 {
					t.Fatalf("after %q: resource %s = %v", cmd, key, rs.Amount)
				}
			}
		}
	}
}

func TestParseCountAndAmount(t *testing.T) {
	for _, s := range []string{"1", "42", "1000000"} {
		if _, err := parseCount(s); err != nil {
			t.Errorf("parseCount(%q): %v", s, err)
		}
	}
	for _, s := range append([]string{"", "abc", "1.5"}, hostileNumbers[:14]...) {
		if s == "1e308" || s == "1e309" || s == "-1e308" {
			continue // not integers either way; covered by the float list
		}
		if n, err := parseCount(s); err == nil {
			t.Errorf("parseCount(%q) = %d, want an error", s, n)
		}
	}
	for _, s := range []string{"1", "0.5", "25", "1e3"} {
		if _, err := parseAmount(s); err != nil {
			t.Errorf("parseAmount(%q): %v", s, err)
		}
	}
	for _, s := range []string{"", "abc", "-1", "0", "-0", "NaN", "nan", "Inf", "+Inf", "-Inf", "1e309", "-1e308"} {
		if n, err := parseAmount(s); err == nil {
			t.Errorf("parseAmount(%q) = %v, want an error", s, n)
		}
	}
}
