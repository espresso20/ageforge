package rules

import (
	"sort"
	"testing"
)

// TestChargedMaterials: the two most-charged materials of each age, worked out
// once at compile, are what counting the age's buildings' costs gives, the
// most charged first and a tie broken by key.
func TestChargedMaterials(t *testing.T) {
	set := Core()
	for _, age := range set.AgeKeys() {
		def, _ := set.Age(age)
		charges := map[string]int{}
		for _, key := range def.UnlockBuildings {
			if b, ok := set.Building(key); ok {
				for res := range b.BaseCost {
					charges[res]++
				}
			}
		}
		var want []string
		for res := range charges {
			want = append(want, res)
		}
		sort.Slice(want, func(i, j int) bool {
			if charges[want[i]] != charges[want[j]] {
				return charges[want[i]] > charges[want[j]]
			}
			return want[i] < want[j]
		})
		var two [2]string
		copy(two[:], want)
		if got := set.ChargedMaterials(age); got != two {
			t.Errorf("%s: ChargedMaterials = %v, want %v", age, got, two)
		}
		for res, level := range set.PriceLevels(age) {
			if got := set.PriceLevel(age, res); got != level {
				t.Errorf("%s %s: PriceLevel = %v, PriceLevels says %v", age, res, got, level)
			}
		}
	}
	if got := set.ChargedMaterials("no_such_age"); got != ([2]string{}) {
		t.Errorf("an unknown age charges %v", got)
	}
}
