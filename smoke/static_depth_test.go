package smoke

import "testing"

// TestStaticDepth: the depth weights, the docs' table, the taste's share
// and the first kit price all hold.
func TestStaticDepth(t *testing.T) {
	if p := StaticDepth(); len(p) != 0 {
		t.Errorf("static depth problems: %v", p)
	}
}
