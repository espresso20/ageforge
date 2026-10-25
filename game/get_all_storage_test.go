package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestGetAllStorageFillsEveryResource: the dev console's /fill reads one entry
// per resource the set defines, each the fill amount.
func TestGetAllStorageFillsEveryResource(t *testing.T) {
	got := NewResourceManager().GetAllStorage()
	want := config.BaseResources()
	if len(got) != len(want) {
		t.Errorf("GetAllStorage holds %d resources, config defines %d", len(got), len(want))
	}
	for _, r := range want {
		if amt, ok := got[r.Key]; !ok || amt != devFillAmount {
			t.Errorf("GetAllStorage[%q] = %v, %v; want %v", r.Key, amt, ok, devFillAmount)
		}
	}
}
