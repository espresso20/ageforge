package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestBlackMarketStatusSaysWhatHappens: the status panel says a deal goes
// through five times in six and pays back the culture's minutes of income, and
// that the sixth time the smugglers vanish with the culture, which is what the
// engine's odds are. It says nothing of a big haul or a multiple of the
// culture.
func TestBlackMarketStatusSaysWhatHappens(t *testing.T) {
	ge := game.NewGameEngine()
	if err := ge.EnterAgeForTest("colonial_age"); err != nil {
		t.Fatal(err)
	}
	if st := ge.BlackMarketStatus(); st.WinChance != 5.0/6 || st.WinMult != 1 {
		t.Fatalf("the engine's odds are %v and %vx, want five in six and the same worth", st.WinChance, st.WinMult)
	}
	out := cmdBlackMarketStatus(ge).Message
	for _, want := range []string{"five in six", "one in six", "vanish with the culture", "minutes of income"} {
		if !strings.Contains(out, want) {
			t.Errorf("the Black Market status does not say %q:\n%s", want, out)
		}
	}
	for _, bad := range []string{"haul", "gamble", "%", "2.5x", "55"} {
		if strings.Contains(out, bad) {
			t.Errorf("the Black Market status says %q:\n%s", bad, out)
		}
	}
}
