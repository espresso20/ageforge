package config

import "testing"

// TestEndedResourcesStayEnded: no building is priced in a resource that
// nothing in its age makes any more (see ended_resources.go for the rule).
func TestEndedResourcesStayEnded(t *testing.T) {
	parts := EndedParts(BaseBuildings(), Ages())
	for _, p := range parts {
		t.Error(p)
	}
	if len(parts) > 0 {
		t.Fatalf("%d price parts charge a resource that has ended", len(parts))
	}
}
