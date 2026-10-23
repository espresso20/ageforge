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

// TestUpgradeKeys: the upgrade table worked out at compile, per age, is
// UpgradeTarget asked of every building one at a time.
func TestUpgradeKeys(t *testing.T) {
	set := Core()
	asked := 0
	for _, age := range append(set.AgeKeys(), "no_such_age") {
		table := set.UpgradeKeys(age)
		for _, b := range set.Buildings() {
			target, ok := set.UpgradeTarget(b.Key, age)
			got, has := table[b.Key]
			if ok != has || (ok && got != target.Key) {
				t.Errorf("%s in %s: UpgradeKeys has %q (%t), UpgradeTarget says %q (%t)", b.Key, age, got, has, target.Key, ok)
			}
			if ok {
				asked++
			}
		}
		for key := range table {
			if _, known := set.Building(key); !known {
				t.Errorf("%s: UpgradeKeys names %q, which is no building", age, key)
			}
		}
	}
	if asked == 0 {
		t.Fatal("no building was older in any age; the test checked nothing")
	}
}
