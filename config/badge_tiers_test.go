package config

import "testing"

// TestBadgeTierWords pins what each tier says and is worth: its name, its
// points and the rarity a badge takes from it. A tier outside the table says
// nothing and is worth nothing.
func TestBadgeTierWords(t *testing.T) {
	cases := []struct {
		tier   BadgeTier
		name   string
		points int
		rarity string
	}{
		{BadgeNoTier, "", 0, ""},
		{BadgeBronze, "bronze", 5, "common"},
		{BadgeSilver, "silver", 10, "uncommon"},
		{BadgeGold, "gold", 25, "rare"},
		{BadgePlatinum, "platinum", 50, "epic"},
		{BadgeLegendary, "legendary", 100, "legendary"},
		{BadgeTier(6), "", 0, ""},
		{BadgeTier(255), "", 0, ""},
	}
	for _, c := range cases {
		if got := c.tier.Name(); got != c.name {
			t.Errorf("tier %d: Name() = %q, want %q", c.tier, got, c.name)
		}
		if got := c.tier.Points(); got != c.points {
			t.Errorf("tier %d: Points() = %d, want %d", c.tier, got, c.points)
		}
		d := BadgeDef{Tier: c.tier}
		if got := d.RarityName(); got != c.rarity {
			t.Errorf("tier %d: RarityName() = %q, want %q", c.tier, got, c.rarity)
		}
		d.Proof.Kind = BadgeProofStatic
		if got := d.Points(); got != c.points {
			t.Errorf("tier %d: a proven badge's Points() = %d, want %d", c.tier, got, c.points)
		}
		d.Proof.Kind = BadgeProofIntegrity
		if got := d.Points(); got != 0 {
			t.Errorf("tier %d: an integrity badge's Points() = %d, want 0", c.tier, got)
		}
	}
}

// TestBadgeRarityWords pins each rarity's name, and that a rarity a badge
// sets wins over the one its tier would give.
func TestBadgeRarityWords(t *testing.T) {
	names := []struct {
		rarity BadgeRarity
		name   string
	}{
		{BadgeRarityByTier, ""},
		{BadgeCommon, "common"},
		{BadgeUncommon, "uncommon"},
		{BadgeRare, "rare"},
		{BadgeEpic, "epic"},
		{BadgeMythic, "legendary"},
		{BadgeRarity(6), ""},
		{BadgeRarity(255), ""},
	}
	for _, c := range names {
		if got := c.rarity.Name(); got != c.name {
			t.Errorf("rarity %d: Name() = %q, want %q", c.rarity, got, c.name)
		}
	}
	for _, tier := range []BadgeTier{BadgeNoTier, BadgeBronze, BadgeLegendary} {
		for _, c := range names {
			if c.rarity == BadgeRarityByTier {
				continue
			}
			d := BadgeDef{Tier: tier, Rarity: c.rarity}
			if got := d.RarityName(); got != c.name {
				t.Errorf("tier %d with rarity %d: RarityName() = %q, want %q", tier, c.rarity, got, c.name)
			}
		}
	}
}
