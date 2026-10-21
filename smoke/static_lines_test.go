package smoke

import "testing"

// TestLinesNeverSwitchResource: every building makes the resource of its
// one-resource line, by the table written out in static_lines.go.
func TestLinesNeverSwitchResource(t *testing.T) {
	for _, p := range StaticLines() {
		t.Error(p)
	}
}
