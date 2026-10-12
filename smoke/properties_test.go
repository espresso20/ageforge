package smoke

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestProperties holds every property of the economy to the state written
// down for it: a known failure that starts to hold, a property that stops
// holding, or a reading that moves, fails here.
func TestProperties(t *testing.T) {
	p := StaticProperties()
	// Numbered 1 to 21 without 13, which was about gate amounts and is gone.
	if len(p.List) != 20 {
		t.Fatalf("%d properties, want 20", len(p.List))
	}
	for _, pr := range p.List {
		if pr.N == 13 {
			t.Error("property 13 is back: no gate asks for an amount of a resource, so it has nothing to measure")
		}
	}
	for _, problem := range p.Problems() {
		t.Error(problem)
	}
	if t.Failed() || testing.Verbose() {
		var sb strings.Builder
		writeProperties(&sb, p)
		t.Log("\n" + sb.String())
	}
}

// TestAcceptedShortagesFitTheLingeringPlayer: a shortage the owner accepts
// (config.MakerAcceptedShortages) is covered, in the end, by a few more of the
// older buildings; it must not take longer to make from the carried required
// copies, at listed output, than the lingering reference player stays in the
// age (RefLinger times as long as it takes the ordinary player to open the
// gate). The reading is in targets of the age.
func TestAcceptedShortagesFitTheLingeringPlayer(t *testing.T) {
	p := StaticProperties()
	pos := map[string]int{}
	for i, a := range p.Ages {
		pos[a] = i
	}
	for _, a := range config.MakerAcceptedShortages() {
		i, ok := pos[config.AgeByKey()[a.Age].Name]
		if !ok {
			t.Fatalf("accepted shortage in unknown age %q", a.Age)
		}
		hours := a.Targets * config.AgeTargets[a.Age].Hours()
		if hours > p.Lingering[i] {
			t.Errorf("%s %s: the carried copies need %.1f targets, %.1f h, more than the lingering player's stay of %.1f h (RefLinger %.0f times the gate)", a.Age, a.Material, a.Targets, hours, p.Lingering[i], RefLinger)
		}
		t.Logf("%s %s: %.2f targets = %.1f h against the lingering player's %.1f h", a.Age, a.Material, a.Targets, hours, p.Lingering[i])
	}
}
