package config

import "testing"

func TestBuildingMetaKeysExistInBaseBuildingsRaw(t *testing.T) {
	rawKeys := make(map[string]bool)
	for _, building := range baseBuildingsRaw() {
		rawKeys[building.Key] = true
	}

	for key := range buildingMeta() {
		if !rawKeys[key] {
			t.Errorf("buildingMeta contains key %q absent from baseBuildingsRaw", key)
		}
	}
}
