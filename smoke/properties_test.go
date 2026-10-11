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
	if len(p.List) != 21 {
		t.Fatalf("%d properties, want 21", len(p.List))
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
