package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// retiredTradeRungs are the rungs of the market-trades ladder, which the
// catalog no longer holds, and retiredTradeCounter is the counter that fed it.
var retiredTradeRungs = []string{
	"ladder.trades.1", "ladder.trades.2", "ladder.trades.3",
	"special.maxed_out", "special.fully_invested",
}

const retiredTradeCounter = "market_trade"

// retiredMapStyleBadges are the two badges for wearing a map style, from
// when the map had two.
var retiredMapStyleBadges = []string{"map.style.roguelike", "map.style.skyline"}

// retiredMissionBadges are the sixteen per-mission badges of the old
// mission list. The catalog now has one per era and kind.
var retiredMissionBadges = []string{
	"expedition.scout_party", "expedition.scout_ruins", "expedition.raid_bandits", "expedition.trade_escort",
	"expedition.conquer_territory", "expedition.siege_castle", "expedition.naval_expedition", "expedition.colonial_campaign",
	"expedition.world_domination", "expedition.cyber_raid", "expedition.neon_heist", "expedition.fusion_assault",
	"expedition.orbital_strike", "expedition.warp_invasion", "expedition.galactic_conquest", "expedition.quantum_incursion",
}

// TestAccountWithRetiredTradeRungsLoadsClean: an account that earned the
// market-trades rungs before the ladder was retired loads with no tamper
// flag. Its rungs and counter stay in the file, signed as they were; the
// rungs stop showing and stop scoring.
func TestAccountWithRetiredTradeRungsLoadsClean(t *testing.T) {
	isolateAccountDir(t)
	ge0, acct := earnSome(t, "Ada")
	id := acct.AccountID
	_, before := ge0.Badges()
	retired := slices.Concat(retiredTradeRungs, retiredMissionBadges, retiredMapStyleBadges)

	// The badge file as the game that still had the ladder signed it.
	var bf badgeFile
	noErr(t, json.Unmarshal(slotBadgeFile(t, id), &bf))
	if bf.Badges == nil {
		bf.Badges = map[string]BadgeEarned{}
	}
	if bf.Counters == nil {
		bf.Counters = map[string]float64{}
	}
	for _, key := range retired {
		bf.Badges[key] = BadgeEarned{At: 1760000000, Run: "old run"}
	}
	bf.Counters[retiredTradeCounter] = 5000
	bf.Signature = signBadgeFile(&bf)
	out, err := json.MarshalIndent(&bf, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(accountDir(id), badgeFileName)
	noErr(t, os.WriteFile(path, out, 0644))

	ge, loaded := openIn(t, id)
	if loaded.BadgesTampered || loaded.Tampered {
		t.Fatalf("an account with the retired rungs loads flagged: badge file tampered %v, account tampered %v", loaded.BadgesTampered, loaded.Tampered)
	}
	for _, key := range retired {
		e, ok := loaded.Badges[key]
		if !ok || e.Flags&BadgeFlagCrossed != 0 {
			t.Errorf("%s after the load: held %v, flags %d; want held and not crossed", key, ok, e.Flags)
		}
	}
	if loaded.Counters[retiredTradeCounter] != 5000 {
		t.Errorf("the %s counter after the load is %v, want 5000", retiredTradeCounter, loaded.Counters[retiredTradeCounter])
	}

	// The rungs are gone from the case and from the score.
	views, after := ge.Badges()
	for _, v := range views {
		if strings.HasPrefix(v.Key, "ladder.trades.") {
			t.Errorf("the case still lists %s", v.Key)
		}
	}
	if after.Earned != before.Earned || after.Points != before.Points {
		t.Errorf("the retired rungs count: %d earned and %d points, were %d and %d without them", after.Earned, after.Points, before.Earned, before.Points)
	}

	// A save that writes the file anew keeps them in it, and the file still
	// verifies.
	diskBefore := slotBadgeFile(t, id)
	recordPrestige(loaded, "stone_age")
	noErr(t, loaded.FlushIfDirty())
	if string(slotBadgeFile(t, id)) == string(diskBefore) {
		t.Fatal("precondition: the save did not write the badge file")
	}
	var saved badgeFile
	noErr(t, json.Unmarshal(slotBadgeFile(t, id), &saved))
	if !verifyBadgeFile(&saved) || saved.Tampered {
		t.Errorf("the badge file after a save: verifies %v, tampered %v", verifyBadgeFile(&saved), saved.Tampered)
	}
	for _, key := range retired {
		if _, ok := saved.Badges[key]; !ok {
			t.Errorf("%s was dropped from the badge file by a save", key)
		}
	}
	if saved.Counters[retiredTradeCounter] != 5000 {
		t.Errorf("the %s counter in the file after a save is %v, want 5000", retiredTradeCounter, saved.Counters[retiredTradeCounter])
	}
	if again := slotAccount(t, id); again.BadgesTampered || again.Tampered {
		t.Errorf("after a save and a load: badge file tampered %v, account tampered %v", again.BadgesTampered, again.Tampered)
	}
}
