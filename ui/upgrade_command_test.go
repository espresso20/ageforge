package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// `upgrade` lists the older buildings the town holds: count, target, the next
// copy's price, how many can be paid for now and any lock. `upgrade <building>`
// upgrades what the town can pay for, and the log says how many and what is next.
func TestUpgradeCommandListsAndUpgrades(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := game.NewGameEngine()
	if got := cmdUpgrade(nil, eng); !strings.Contains(got.Message, "No building upgrades available") {
		t.Fatalf("with nothing older held, upgrade says %q", got.Message)
	}
	eng.Buildings.LoadCounts(map[string]int{"gathering_camp": 3, "stash": 40})
	if err := eng.EnterAgeForTest("stone_age"); err != nil {
		t.Fatal(err)
	}
	for _, r := range []string{"wood", "stone", "food"} {
		eng.Resources.Add(r, 5000)
	}

	list := cmdUpgrade(nil, eng).Message
	for _, want := range []string{"Older buildings you can upgrade", "gathering_camp", "forager_post", "3 held", "next copy", "affordable now"} {
		if !strings.Contains(list, want) {
			t.Errorf("the list lacks %q:\n%s", want, list)
		}
	}
	if strings.Contains(list, "stash") {
		t.Errorf("the list holds storage, which is never older:\n%s", list)
	}

	if res := cmdUpgrade([]string{"gathering_camp", "1"}, eng); res.Type == "error" {
		t.Fatalf("upgrade gathering_camp 1: %q", res.Message)
	}
	if res := cmdUpgrade([]string{"gathering_camp"}, eng); res.Type == "error" {
		t.Fatalf("upgrade gathering_camp: %q", res.Message)
	}
	st := eng.GetState()
	if st.Buildings["gathering_camp"].Count != 0 || st.Buildings["forager_post"].Count != 3 {
		t.Errorf("camps %d, posts %d after upgrading; want 0 and 3", st.Buildings["gathering_camp"].Count, st.Buildings["forager_post"].Count)
	}
	if res := cmdUpgrade([]string{"stash"}, eng); res.Type != "error" || !strings.Contains(res.Message, "has no upgrade") {
		t.Errorf("upgrade stash: %+v", res)
	}
	if res := cmdUpgrade([]string{"gathering_camp", "x"}, eng); res.Type != "error" {
		t.Errorf("upgrade gathering_camp x: %+v, want a usage error", res)
	}
}
