package smoke

import "testing"

// TestEndedResourcesStayEnded: no building is priced in a resource that
// nothing in its age makes any more (see static_ended.go).
func TestEndedResourcesStayEnded(t *testing.T) {
	for _, p := range StaticEnded() {
		t.Error(p)
	}
}
