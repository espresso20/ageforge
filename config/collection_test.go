package config

import (
	"slices"
	"testing"
)

func TestCoinTypes(t *testing.T) {
	if len(CoinTypes) != 5 {
		t.Fatalf("%d coin types, want 5", len(CoinTypes))
	}
	for _, k := range CoinTypes {
		if !IsCoinType(k) {
			t.Errorf("%s is not a coin type", k)
		}
	}
	if IsCoinType("swift") {
		t.Error("swift is not a coin type")
	}
}

func TestFeatCoinIsNotAWell(t *testing.T) {
	if !IsCoinType(CoinFeat) || slices.Contains(CoinTypes, CoinFeat) {
		t.Error("the feat coin must be a coin type and not one of the wells")
	}
	if !CoinLevelOK(CoinFeat, CoinLegendary) || CoinLevelOK(CoinFeat, CoinLevels) || CoinLevelOK(CoinAge, CoinLegendary) || !CoinLevelOK(CoinAge, CoinLevels) {
		t.Error("levels: the feat coin has only the legendary level and the others never reach it")
	}
}
