package smoke

import (
	"strings"
	"testing"
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
