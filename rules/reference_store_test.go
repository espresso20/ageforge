package rules

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestSetCarriesTheReferenceStore: a ruleset knows what one copy of each
// storage building held in the reference store, and the reference store on
// leaving each age (the base store and five copies of every storage building
// so far). A storage building holds nothing now; the rules that still read a
// store size (the festival and Black Market prices, deals, the Brace ceiling,
// the Maximalist badges, event stock shares) are measured on this.
func TestSetCarriesTheReferenceStore(t *testing.T) {
	set := Core()
	if got := set.ReferenceStorePerCopy("stash"); got != 150 {
		t.Errorf("a Stash held %v in the reference store, want 150", got)
	}
	if got := set.ReferenceStorePerCopy("lumber_mill"); got != 0 {
		t.Errorf("a building that is not a storage building has a reference size %v", got)
	}
	want := config.ReferenceBaseStore()
	for _, a := range set.Ages() {
		for _, d := range set.Buildings() {
			if d.Category == "storage" && d.RequiredAge == a.Key {
				want += float64(config.StorageRuleCopies * set.ReferenceStorePerCopy(d.Key))
			}
		}
		if got := set.ReferenceStore(a.Key); got != want {
			t.Errorf("%s: reference store %v, want %v", a.Key, got, want)
		}
		if typical := set.TypicalStorage("wood", a.Key); math.Abs(typical-want) > 1e-9*want {
			t.Errorf("%s: the typical store of wood is %v, the reference store %v", a.Key, typical, want)
		}
	}
	if set.ReferenceStore("no_such_age") != 0 {
		t.Error("an unknown age has a reference store")
	}
}
