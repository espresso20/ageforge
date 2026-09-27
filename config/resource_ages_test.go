package config

import "testing"

// TestResourceAgesMatchAgeUnlocks keeps the two statements of "when does this
// resource unlock" in agreement: ResourceDef.Age and the UnlockResources list
// of the age that unlocks it. Live play unlocks from the age lists; they
// disagreed for coal (Iron Age here, Renaissance Age there), and loading any
// Iron Age save unlocked coal three ages early.
func TestResourceAgesMatchAgeUnlocks(t *testing.T) {
	unlockedBy := map[string]string{}
	for _, age := range Ages() {
		for _, key := range age.UnlockResources {
			if prev, dup := unlockedBy[key]; dup {
				t.Errorf("resource %s is unlocked by both %s and %s", key, prev, age.Key)
			}
			unlockedBy[key] = age.Key
		}
	}
	defs := ResourceByKey()
	for key, age := range unlockedBy {
		if _, ok := defs[key]; !ok {
			t.Errorf("age %s unlocks unknown resource %q", age, key)
		}
	}
	for _, r := range BaseResources() {
		age, ok := unlockedBy[r.Key]
		if !ok {
			t.Errorf("resource %s (Age %s) is not in any age's UnlockResources", r.Key, r.Age)
			continue
		}
		if r.Age != age {
			t.Errorf("resource %s declares Age %q but %s unlocks it", r.Key, r.Age, age)
		}
	}
}
