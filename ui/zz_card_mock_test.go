package ui

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// SCRATCH: pictures of sample decks. Not for commit.

func TestZZCardMock(t *testing.T) {
	out := os.Getenv("CARD_MOCK_DIR")
	if out == "" {
		t.Skip("set CARD_MOCK_DIR")
	}
	keepTheme(t)
	trade := deckPage{
		Title: "Trade · Medieval Age",
		Head:  []xseg{{"Routes ", xDim}, {"2 running", xGood}, {" · ", xFaint}, {"4 open", xText}, {" · ", xFaint}, {"2 locked", xDim}, {"     Allies ", xDim}, {"Riverlands Tribes +10% food", xGood}},
		Notes: []deckLine{{Value: "⚠ Routes importing iron are suspended (a war, or your embargo). Lift your embargo with: diplomacy neutral <civ>", K: xBad}},
		Sections: []deckSection{
			{Title: "Running", Meta: "2 of 3 caravans out", Cards: []deckCard{
				{Title: "Local Barter", Chip: "running", ChipK: xChipShort, Lines: []deckLine{{"Export", "10 food", 0}, {"Import", "8 wood", xGood}}, Meter: true, Share: 0.4, MeterText: "1m 12s left · 14 cycles done", Action: "stop", Command: "trade route stop local_barter"},
				{Title: "Iron Road", Chip: "suspended", ChipK: xChipMil, Lines: []deckLine{{"Export", "40 stone", 0}, {"Import", "12 iron", xBad}}, Why: "Suspended: iron imports are blockaded", Action: "stop", Command: "trade route stop iron_road"},
			}},
			{Title: "Open routes", Cards: []deckCard{
				{Title: "Stone Trade", Chip: "ready", ChipK: xChipCivic, Desc: "Send wood to the quarries for stone.", Lines: []deckLine{{"Export", "15 wood", 0}, {"Import", "12 stone", xGood}}, Action: "start", Command: "trade route start stone_trade"},
				{Title: "Gold Caravan", Chip: "ready", ChipK: xChipCivic, Desc: "Send stone caravans out for gold.", Lines: []deckLine{{"Export", "50 stone", 0}, {"Import", "5 gold", xGood}}, Action: "start", Command: "trade route start gold_caravan"},
				{Title: "Silk Road", Chip: "locked", ChipK: xKeyName, Desc: "Send gold along the Silk Road for culture.", Lines: []deckLine{{"Export", "30 gold", 0}, {"Import", "80 culture", xGood}}, Why: "needs 2 Markets (have 0)", Quiet: true},
				{Title: "Spice Fleet", Chip: "locked", ChipK: xKeyName, Desc: "Ships carry gold east and come home heavy with spice and stories.", Lines: []deckLine{{"Export", "120 gold", 0}, {"Import", "40 culture, 10 faith", xGood}}, Why: "needs 1 Harbor (have 0)", Quiet: true},
			}},
		},
		Hint: "trade route start <key>",
	}
	miles := deckPage{
		Title: "Milestones · Medieval Age",
		Head:  []xseg{{"Progress ", xDim}, {"5 of 77", xBright}, {"   Title ", xDim}, {"\"Aspiring\"", xTitle}, {"   Next title at ", xDim}, {"8", xText}},
		Sections: []deckSection{
			{Title: "Settlement", Meta: "1 of 6", MinW: 34, Cards: []deckCard{
				{Title: "First Shelter", Chip: "✓ done", ChipK: xChipShort, Desc: "Build your first home.", Lines: []deckLine{{"Reward", "+10 food", xGood}}, Quiet: true},
				{Title: "Small Village", Chip: "now", ChipK: xChipCivic, Desc: "Reach a population of 250.", Meter: true, Share: 0.08, MeterText: "20/250 people", Lines: []deckLine{{"Reward", "+50 wood", xGood}}},
				{Title: "Bustling Town", Desc: "Reach a population of 1,500.", Meter: true, Share: 0.013, MeterText: "20/1.5K people", Lines: []deckLine{{"Reward", "+5% food production", xGood}}},
				{Title: "Growing City", Desc: "Reach a population of 5,000.", Meter: true, Share: 0.004, MeterText: "20/5K people", Lines: []deckLine{{"Reward", "+10% food production", xGood}}},
				{Title: "Metropolis", Desc: "Reach a population of 15,000.", Meter: true, Share: 0.001, MeterText: "20/15K people", Lines: []deckLine{{"Reward", "+10% all production", xGood}}},
				{Title: "Urban Sprawl", Desc: "Reach a population of 50,000.", Meter: true, Share: 0, MeterText: "20/50K people", Lines: []deckLine{{"Reward", "+15% all production", xGood}}},
			}},
			{Title: "Builder", Meta: "0 of 5", MinW: 34, Cards: []deckCard{
				{Title: "Early Builder", Chip: "now", ChipK: xChipCivic, Desc: "Build 10 buildings.", Meter: true, Share: 0.3, MeterText: "3/10 buildings", Lines: []deckLine{{"Reward", "+20 stone", xGood}}},
				{Title: "Master Mason", Desc: "Build 100 buildings.", Meter: true, Share: 0.03, MeterText: "3/100 buildings", Lines: []deckLine{{"Reward", "+5% build speed", xGood}}},
				{Title: "???", Chip: "hidden", ChipK: xKeyName, Desc: "A hidden milestone.", Quiet: true},
			}},
		},
		Hint: "milestones",
	}
	army := deckPage{
		Title: "Army · Medieval Age",
		Notes: []deckLine{{Value: "You have no garrison: raids hit you with full force. Soldiers blunt raids, war raids and what an Endure takes (up to 45%).", K: xBad}},
		Sections: []deckSection{
			{Title: "Overview", MinW: 30, Cards: []deckCard{
				{Title: "Soldiers", Lines: []deckLine{{"", "0 posted of 0", xBright}, {"", "none away", xDim}}},
				{Title: "Posts", Lines: []deckLine{{"", "none yet", xBright}, {"", "a military building comes with its soldiers", xDim}}},
				{Title: "Defense", Lines: []deckLine{{"", "0", xBright}, {"", "blunts up to 45% of a raid", xDim}}},
				{Title: "Threat", Chip: "high", ChipK: xChipMil, Lines: []deckLine{{"", "2.81K", xBad}, {"", "raids in the Medieval Age", xDim}}},
			}},
			{Title: "Campaigns", Meta: "0 done", Cards: []deckCard{
				{Title: "Border Skirmish", Chip: "ready", ChipK: xChipCivic, Desc: "Push the raiders back past the river and keep what they drop.", Lines: []deckLine{{"Needs", "40 soldiers", 0}, {"Odds", "72%", xGood}, {"Haul", "about 20 min of income", xGood}}, Action: "send", Command: "expedition border_skirmish"},
				{Title: "Siege of Ashford", Chip: "locked", ChipK: xKeyName, Desc: "A walled town that has it coming.", Lines: []deckLine{{"Needs", "200 soldiers", 0}}, Why: "needs 2 Barracks (have 0)", Quiet: true},
			}},
		},
		Hint: "expedition <name>",
	}
	shots := []struct {
		name      string
		pg        deckPage
		w, h, sel int
	}{
		{"trade-153", trade, 153, 38, 2}, {"miles-153", miles, 153, 38, 1}, {"army-153", army, 153, 38, 4},
		{"trade-100", trade, 100, 30, 0}, {"miles-80", miles, 80, 21, 1},
	}
	for _, s := range shots {
		g, lay := renderDeck(&s.pg, s.sel, 0, s.w, s.h)
		total := lay.total
		b := kit.NewBuffer(s.w, s.h)
		paintExpeditionKit(kit.NewPainter(b), g, image.Point{}, newExpPalette())
		f, err := os.Create(filepath.Join(out, "CD-"+s.name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, ScreenImage(b, shotCellW, shotCellH)); err != nil {
			t.Fatal(err)
		}
		f.Close()
		t.Logf("%s: %d rows of content", s.name, total)
	}
}
