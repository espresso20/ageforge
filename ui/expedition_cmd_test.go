package ui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// newExpeditionTestEngine spins up a fresh engine in the Bronze Age, stocked
// with enough food, wood and soldiers at home to launch any offer on its mission
// board. The age jump uses the dev console (/age applies the age's unlocks);
// resource amounts are set via LoadAmounts so they bypass nominal storage
// caps.
func newExpeditionTestEngine(t *testing.T) *game.GameEngine {
	t.Helper()

	prevDev := game.DevModeActive
	game.DevModeActive = true
	t.Cleanup(func() { game.DevModeActive = prevDev })

	engine := game.NewGameEngine()
	engine.SeedRNG(11)
	engine.GrantTechsForTest("military_tactics", "exploration") // the techs that open campaigns and expeditions
	if msg := game.DevExecCommand("/age bronze_age", engine); msg != "jumped to bronze_age" {
		t.Fatalf("dev /age bronze_age failed: %q", msg)
	}
	engine.Resources.LoadAmounts(map[string]float64{
		"food": 500,
		"wood": 500,
	})
	// Every mission takes a party: an army at home that covers a few.
	engine.SetGarrisonForTest("bronze_age", 500)
	// The Bronze Age board holds short offers only; these tests send every tier.
	engine.OpenTiersForTest(3)
	return engine
}

// numbered matches an offer number at the start of a line ("3. Scout...").
var numbered = regexp.MustCompile(`(?m)^\W*\d+\. `)

// TestCmdExpeditionNoArgsPrintsTheBoard verifies bare `expedition` opens the
// Expeditions panel and `expedition list` prints the mission board as plain
// lines (nine offers, no numbers), neither launching anything.
func TestCmdExpeditionNoArgsPrintsTheBoard(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	if bare := cmdExpedition(nil, engine); bare.OverlayName != "expedition" {
		t.Errorf("cmdExpedition(nil) opens overlay %q, want the expedition panel", bare.OverlayName)
	}
	res := cmdExpedition([]string{"list"}, engine)
	if res.Type != "info" || res.OverlayName != "" {
		t.Errorf("cmdExpedition(list) = type %q overlay %q, want a printed board", res.Type, res.OverlayName)
	}
	if !strings.Contains(res.Message, "Mission board") {
		t.Errorf("the board has no heading:\n%s", res.Message)
	}
	if numbered.MatchString(res.Message) {
		t.Errorf("the printed board carries offer numbers:\n%s", res.Message)
	}
	for _, slot := range engine.GetState().Military.Board {
		if slot.Offer != nil && !strings.Contains(res.Message, slot.Offer.Name) {
			t.Errorf("the board does not list %q:\n%s", slot.Offer.Name, res.Message)
		}
	}
	st := engine.GetState()
	words, _ := st.Ruleset().MissionWords(st.Age)
	wantAll(t, "the board's sections", res.Message, words.Tiers[0]+":", words.Tiers[1]+":", "Grand:")
	if st.Military.ActiveShort != nil || st.Military.ActiveLong != nil || st.Military.ActiveGrand != nil {
		t.Error("printing the board launched something")
	}
	if list := cmdExpedition([]string{"list"}, engine); list.Message != res.Message {
		t.Error("`expedition list` is not the board")
	}
}

// TestCmdExpeditionNumberLaunches verifies `expedition <n>` sends offer n with
// its party; a second offer of the same tier goes too while the soldiers at
// home cover it, and is refused with the party and the soldiers at home when
// they do not.
func TestCmdExpeditionNumberLaunches(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	res := cmdExpedition([]string{"1"}, engine)
	if res.Type != "success" {
		t.Fatalf("cmdExpedition(1) Type = %q (msg %q), want %q", res.Type, res.Message, "success")
	}
	if st := engine.GetState(); st.Military.ActiveShort == nil {
		t.Errorf("after sending offer 1, ActiveShort is nil; expected a short mission under way")
	}
	if res := cmdExpedition([]string{"2"}, engine); res.Type != "success" {
		t.Errorf("a second short mission beside the first: type %q, %q", res.Type, res.Message)
	}
	if n := len(engine.GetState().Military.Active); n != 2 {
		t.Errorf("%d missions under way, want 2", n)
	}
	// Eight soldiers per short party, two out; three at home cannot make a third.
	engine.SetGarrisonForTest("bronze_age", 19)
	if res := cmdExpedition([]string{"3"}, engine); res.Type != "error" || !strings.Contains(res.Message, "needs a party of 8 soldiers, and 3 are at home") {
		t.Errorf("a third short mission with 3 at home: type %q, %q; want the party and the soldiers at home", res.Type, res.Message)
	}
	engine.SetGarrisonForTest("bronze_age", 500)
	if res := cmdExpedition([]string{"4"}, engine); res.Type != "success" {
		t.Errorf("a long mission beside the short one: type %q, %q", res.Type, res.Message)
	}
	if st := engine.GetState(); st.Military.ActiveLong == nil || st.Military.ActiveShort == nil {
		t.Error("the short and the long mission should both be out")
	}
}

// TestCmdCampaignIsTheSameCommand verifies `campaign` prints the same board
// and sends the same offers as `expedition`.
func TestCmdCampaignIsTheSameCommand(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	if a, b := cmdCampaign([]string{"list"}, engine), cmdExpedition([]string{"list"}, engine); a.Message != b.Message || a.Type != "info" {
		t.Errorf("campaign list prints %q, expedition list %q", a.Message, b.Message)
	}
	if a := cmdCampaign(nil, engine); a.OverlayName != "expedition" {
		t.Errorf("bare campaign opens overlay %q, want the expedition panel", a.OverlayName)
	}
	res := cmdCampaign([]string{"7"}, engine)
	if res.Type != "success" {
		t.Fatalf("cmdCampaign(7) Type = %q (msg %q), want %q", res.Type, res.Message, "success")
	}
	if st := engine.GetState(); st.Military.ActiveGrand == nil || st.Military.ActiveMilitary == nil {
		t.Errorf("after sending offer 7, no grand mission is out: %+v", st.Military)
	}
}

// TestCmdExpeditionBadArguments verifies a word, an offer number off the
// board and a number with more after it are refused, and launch nothing.
func TestCmdExpeditionBadArguments(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	for _, args := range [][]string{{"scouting", "party"}, {"scouting_party"}, {"12"}, {"0"}, {"1", "2"}} {
		res := cmdExpedition(args, engine)
		if res.Type != "error" {
			t.Errorf("cmdExpedition(%v) Type = %q (msg %q), want an error", args, res.Type, res.Message)
		}
	}
	st := engine.GetState()
	if st.Military.ActiveShort != nil || st.Military.ActiveLong != nil || st.Military.ActiveGrand != nil {
		t.Error("a refused command launched something")
	}
}

// TestOfferLineMarksLegendaryOffers verifies a Legendary offer is marked in
// the printed board and an ordinary one is not.
func TestOfferLineMarksLegendaryOffers(t *testing.T) {
	engine := newExpeditionTestEngine(t)
	state := engine.GetState()
	info := game.ExpeditionInfo{Number: 3, Name: "Legendary Hunt: track the herd across the Dry Creek", Rare: true, CanLaunch: true, SuccessChance: .6, DurationMin: 900}
	if line := offerLine(info, state); !strings.Contains(line, "Legendary") || !strings.Contains(line, "★") || !strings.Contains(line, "60% odds") {
		t.Errorf("a Legendary offer is not marked or lacks its odds: %q", line)
	}
	info.Rare, info.Name = false, "Hunt: track the herd across the Dry Creek"
	if line := offerLine(info, state); strings.Contains(line, "★") {
		t.Errorf("an ordinary offer is starred: %q", line)
	}
}

// TestMatchOffers verifies a typed name finds the offers it means: exact, else
// starts-with, else every word a prefix of a title word, with "the", "of",
// "a" and "an" left out and case ignored.
func TestMatchOffers(t *testing.T) {
	titles := []string{
		"Scout: count the herds near the Copper Hills",                   // 0
		"Scout: check the wells along the Cedar Pass",                    // 1
		"Trade Trek: carry copper to the Cedar Pass",                     // 2
		"Legendary Trade Trek: follow the amber road to the Amber Coast", // 3
		"Skirmish: hold the ford",                                        // 4
		"Skirmish: hold the ford near the Clay Terraces",                 // 5
		"Trade Mission: chart a trade route through the Clay Terraces",   // 6
	}
	for _, c := range []struct {
		input string
		want  []int
	}{
		{"scout count herds near copper hills", []int{0}},          // exact, once the small words are gone
		{"Scout: Count the Herds near the Copper Hills", []int{0}}, // case, colon and small words
		{"skirmish hold ford", []int{4}},                           // exact beats starts-with
		{"skirmish: hold the ford near", []int{5}},                 // starts with
		{"scout", []int{0, 1}},                                     // starts with, two of them
		{"sco", []int{0, 1}},                                       // a part of a word
		{"cedar pass", []int{1, 2}},                                // words of the title, any place
		{"pass cedar", []int{1, 2}},                                // in any order
		{"the cedar", []int{1, 2}},                                 // "the" is not a word
		{"a copper", []int{0, 2}},                                  // nor is "a"
		{"of an the", nil},                                         // nothing left to match
		{"legendary", []int{3}},                                    // the mark is part of the title
		{"amber", []int{3}},                                        // the first of the words
		{"terraces chart", []int{6}},                               // two words, one title
		{"dragon", nil},
		{"", nil},
	} {
		got := matchOffers(titles, c.input)
		if fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("matchOffers(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

// TestCmdExpeditionSendsByName verifies `expedition <part of a name>` sends the
// one offer it names, asks which when several match, and says when none does.
func TestCmdExpeditionSendsByName(t *testing.T) {
	engine := newExpeditionTestEngine(t)
	engine.EnsureMissionBoard()
	board := engine.GetState().Military.Board

	if res := cmdExpedition([]string{"scout"}, engine); res.Type == "success" || !strings.HasPrefix(res.Message, "Which one? ") || !strings.HasSuffix(res.Message, ". Type more of the name.") {
		t.Errorf("an ambiguous name: type %q, %q", res.Type, res.Message)
	} else {
		for _, slot := range board[:3] {
			if !strings.Contains(res.Message, slot.Offer.Name) {
				t.Errorf("the question leaves out %q: %q", slot.Offer.Name, res.Message)
			}
		}
	}
	if st := engine.GetState(); st.Military.ActiveShort != nil {
		t.Fatal("asking which one sent something")
	}

	if res, want := cmdExpedition([]string{"dragon", "keep"}, engine), `No offer called "dragon keep". Type expedition list to see the board.`; res.Message != want || res.Type != "error" {
		t.Errorf("no match: type %q, %q; want %q", res.Type, res.Message, want)
	}

	// The whole of one title, typed as words: it is the one offer.
	name := board[3].Offer.Name
	res := cmdExpedition(strings.Fields(strings.ToLower(name)), engine)
	if res.Type != "success" {
		t.Fatalf("sending %q by its name: type %q, %q", name, res.Type, res.Message)
	}
	if st := engine.GetState(); st.Military.ActiveLong == nil || st.Military.ActiveLong.Name != name {
		t.Errorf("under way: %+v, want %q", st.Military.ActiveLong, name)
	}
	// A name that matches an offer the army at home cannot make a party for is
	// refused by the game.
	engine.SetGarrisonForTest("bronze_age", 25) // a long party is 20; 5 are at home
	if res := cmdExpedition(strings.Fields(strings.ToLower(board[4].Offer.Name)), engine); res.Type != "error" || !strings.Contains(res.Message, "needs a party of 20 soldiers, and 5 are at home") {
		t.Errorf("a second long mission by name: type %q, %q", res.Type, res.Message)
	}
}
