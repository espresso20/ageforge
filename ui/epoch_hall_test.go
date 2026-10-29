package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
)

// lateHallState is a run well on its way: it has been through every age up
// to the one named, with a few badges earned.
func lateHallState(t *testing.T, age string) game.GameState {
	t.Helper()
	engine := game.NewGameEngine()
	must(t, engine.SummonHarbingerForTest(age))
	st := engine.GetState()
	st.Stats.AgesReached = nil
	st.Stats.AgeLog = nil
	for i, a := range st.Ruleset().Ages() {
		st.Stats.AgesReached = append(st.Stats.AgesReached, a.Key)
		st.Stats.AgeLog = append(st.Stats.AgeLog, game.AgeEntry{Age: a.Key, Tick: i * 900})
		if a.Key == age {
			break
		}
	}
	st.CatastrophesEndured = 2
	st.AccountStats = &game.AccountStatsView{Badges: []game.BadgeView{
		{Key: "a", Name: "Common One", Rarity: "common", Points: 1, Earned: true},
		{Key: "b", Name: "Epic Two", Rarity: "epic", Points: 20, Earned: true, Emblem: "star"},
		{Key: "c", Name: "Rare Three", Rarity: "rare", Points: 10, Earned: true},
		{Key: "d", Name: "Legend Four", Rarity: "legendary", Points: 50, Earned: true},
		{Key: "e", Name: "Uncommon Five", Rarity: "uncommon", Points: 3, Earned: true},
		{Key: "f", Name: "Unearned Six", Rarity: "mythic", Points: 99},
		{Key: "g", Name: "Common Seven", Rarity: "common", Points: 1, Earned: true},
	}}
	return st
}

// unseenAges are the names of the ages the player may not see yet.
func unseenAges(st game.GameState) []string {
	sight := game.SightOf(&st)
	var out []string
	for _, a := range st.Ruleset().Ages() {
		if !sight.Age(a.Key) {
			out = append(out, a.Name)
		}
	}
	return out
}

func TestEpochHallFrontPage(t *testing.T) {
	early := game.NewGameEngine().GetState()
	late := lateHallState(t, "galactic_age")
	for name, st := range map[string]game.GameState{"early": early, "late": late} {
		hidden := unseenAges(st)
		if len(hidden) == 0 {
			t.Fatalf("%s: setup: nothing is unseen", name)
		}
		for _, dim := range [][2]int{{200, 50}, {80, 24}} {
			txt := untag(hallFrontText(st, dim[0], dim[1], 0))
			if strings.Contains(txt, "…") {
				t.Errorf("%s %dx%d: prose cut off:\n%s", name, dim[0], dim[1], txt)
			}
			for _, a := range hidden {
				if strings.Contains(txt, a) {
					t.Errorf("%s %dx%d: names the unseen age %q", name, dim[0], dim[1], a)
				}
			}
			wantAll(t, name, txt, "Highlights", "Ages", "Top badges", "Feats", "more to come")
		}
	}
	txt := untag(hallFrontText(early, 120, 40, 0))
	wantAll(t, "early", txt, "No badges earned yet.")
	txt = untag(hallFrontText(late, 120, 40, 0))
	wantAll(t, "late", txt, "Legend Four", "Epic Two", "Rare Three", "Uncommon Five", "Common One")
	if strings.Contains(txt, "Unearned Six") || strings.Contains(txt, "Common Seven") {
		t.Errorf("late: shows more than the five rarest earned:\n%s", txt)
	}
	if strings.Index(txt, "Legend Four") > strings.Index(txt, "Epic Two") {
		t.Errorf("late: badges are not rarest first")
	}
}

func TestEpochHallTabSwitchesPages(t *testing.T) {
	h := &epochHall{}
	st := game.NewGameEngine().GetState()
	front := h.provider(st, 120)
	if !h.handleKey(wkey(tcell.KeyTab, 0)) || !h.details {
		t.Fatal("Tab did not open the details page")
	}
	if details := h.provider(st, 120); details == front || details != epochProvider(st, 120) {
		t.Error("the details page is not the old panel")
	}
	h.handleKey(wkey(tcell.KeyTab, 0))
	if h.details || h.provider(st, 120) != front {
		t.Error("Tab did not return to the front page")
	}
	h.handleKey(wkey(tcell.KeyTab, 0))
	h.open()
	if h.details {
		t.Error("open should start on the front page")
	}
	if h.handleKey(wkey(tcell.KeyDown, 0)) {
		t.Error("scroll keys belong to the text view")
	}
}

func TestEpochCard(t *testing.T) {
	late := lateHallState(t, "galactic_age")
	card := epochCardText(late)
	if !strings.HasPrefix(card, "```\n") || !strings.HasSuffix(card, "```\n") {
		t.Errorf("card is not fenced:\n%s", card)
	}
	wantAll(t, "card", card, "Highlights", "Ages", "Top badges", "Feats", "Legend Four", "Done:", "Now:", "Ahead:", "more to come")
	if cardTagRe.MatchString(card) {
		t.Errorf("card holds color tags:\n%s", card)
	}
	for _, l := range strings.Split(card, "\n") {
		if n := runeLen(l); n > epochCardWidth {
			t.Errorf("line is %d wide: %q", n, l)
		}
	}
	for _, a := range unseenAges(late) {
		if strings.Contains(card, a) {
			t.Errorf("card names the unseen age %q", a)
		}
	}
}

func TestEpochCardCommandWritesFile(t *testing.T) {
	dir := t.TempDir()
	defer game.SetDataDirForTest(dir)()
	engine := game.NewGameEngine()
	res := HandleCommand("epoch card", engine)
	if res.Type != "success" || !strings.Contains(res.Message, "epoch-card.txt") {
		t.Fatalf("reply = %+v", res)
	}
	if res := HandleCommand("epoch", engine); res.OverlayName != "epoch" {
		t.Errorf("bare epoch should open the screen: %+v", res)
	}
}
