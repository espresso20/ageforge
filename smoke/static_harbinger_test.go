package smoke

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestHarbingerPricesFitStorage: every Appease and Brace price, both levels,
// fits the most storage buildable in the first age of its era. Appease grows
// with the pacing targets while faith and culture storage stay as typed, so a
// longer curve is what would break this.
func TestHarbingerPricesFitStorage(t *testing.T) {
	rows := HarbingerPrices()
	if len(rows) == 0 {
		t.Fatal("no harbinger prices read")
	}
	for _, p := range StaticHarbingerPrices() {
		t.Errorf("%s %s costs %v %s, over the %v buildable in %s", p.Epoch, p.Answer, p.Price, p.Resource, p.MaxStorage, p.Age)
	}
	// The check sees a price that cannot fit: faith storage in the Iron Era's
	// first age is finite, so no row may claim an uncapped faith store there.
	if m := MaxStorage(config.EpochByKey()["iron_era"].Ages[0], "faith"); m <= 0 || m > 1e300 {
		t.Errorf("iron-age faith storage %v: the check would never fire", m)
	}
}
