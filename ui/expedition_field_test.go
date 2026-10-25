package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// fieldState is the Bronze Age town with one party out: eight soldiers set
// out, two fell on the road, its odds fell from 72% to 64%, a find scaled its
// loot by 1.4, and three road events have fired.
func fieldState(t *testing.T) game.GameState {
	t.Helper()
	engine, _ := expFixture(t)
	state := engine.GetState()
	state.Military.Active = []*game.ExpeditionSnapshot{{
		Name: "Scout: Count the herds near the Copper Hills", Place: "Copper Hills", ExpiresTick: 900,
		Label: "Scout", Tier: game.TierShort, Party: 8, Soldiers: 6,
		TotalTicks: 100, TicksLeft: 40, SuccessChance: .64, LaunchChance: .72,
		HaulMinutes: 20, HaulScale: 1.4, SuccessHaul: map[string]float64{"food": 120, "wood": 40},
		RoadStories: []string{"A herder showed the party a dry ford.", "Bandits shadowed the party for days.", "The party found an amber hoard."},
		RoadKinds:   []string{"helpers", "foes", "find"},
	}}
	return state
}

// colourOf is the colour of the first cell of word on the grid's row that
// holds it.
func colourOf(t *testing.T, g *xgrid, word string) xk {
	t.Helper()
	for y, line := range g.Lines() {
		if x := strings.Index(line, word); x >= 0 {
			return g.at(runeLen(line[:x]), y).k
		}
	}
	t.Fatalf("%q is not on the grid:\n%s", word, g.String())
	return 0
}

func TestReportSplitsTheWidthNearFiftyFiveForTheMap(t *testing.T) {
	for _, w := range []int{100, 160, 240} {
		l := reportLayoutFor(w, 30)
		if !l.split {
			t.Fatalf("%d columns: no split", w)
		}
		if l.mapW+l.cardW != w-3 || l.cardX != l.mapW+2 {
			t.Errorf("%d columns: map %d and card %d at %d do not share the width", w, l.mapW, l.cardW, l.cardX)
		}
		if share := float64(l.mapW) / float64(w-3); share < .52 || share > .58 {
			t.Errorf("%d columns: the map takes %.0f%%, want about 55%%", w, share*100)
		}
		if l.cardW < 36 || l.mapW < 40 {
			t.Errorf("%d columns: map %d, card %d, below the minimums", w, l.mapW, l.cardW)
		}
	}
	if l := reportLayoutFor(80, 24); l.split {
		t.Errorf("80 columns split; the compact layout should hold")
	}
}

func TestFieldReportShowsTheWholeParty(t *testing.T) {
	state := fieldState(t)
	g := renderExpedition(state, expView{report: true}, 160, 45)
	text := g.String()
	wantAll(t, "the field report", text, "Field report", "Count the herds near the Copper Hills", "Scout · short",
		"Party  8 soldiers · 2 fell so far", "64% (was 72%)", "x1.4 from the road", "On the road",
		"to go", "about 20 min of income", "+120 food", "+40 wood")
	// Every fired road event, newest first.
	last := -1
	for _, story := range []string{"amber hoard", "Bandits shadowed", "dry ford"} {
		i := strings.Index(text, story)
		if i < 0 {
			t.Fatalf("the field report omits %q:\n%s", story, text)
		}
		if i < last {
			t.Errorf("%q is out of order; road events run newest first", story)
		}
		last = i
	}
	if k := colourOf(t, g, "amber"); k != xGood {
		t.Errorf("a find is coloured %v, want the good colour", k)
	}
	if k := colourOf(t, g, "Bandits"); k != xBad {
		t.Errorf("an attack is coloured %v, want the bad colour", k)
	}
	// A mission from before the launch odds were kept shows no change.
	state.Military.Active[0].LaunchChance = 0
	if text := renderExpedition(state, expView{report: true}, 160, 45).String(); strings.Contains(text, "(was") {
		t.Errorf("odds with no launch value show a change:\n%s", text)
	}
}

func TestOutListSelectsTheMissionItShows(t *testing.T) {
	state := fieldState(t)
	second := *state.Military.Active[0]
	second.Name, second.Place, second.Label, second.ExpiresTick = "Trade: Haul the salt from the Grey Flats", "Grey Flats", "Trade Trek", 1200
	second.RoadStories, second.RoadKinds = nil, nil
	state.Military.Active = append(state.Military.Active, &second)
	text := renderExpedition(state, expView{report: true, party: 1}, 160, 45).String()
	var marked string
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "▸") {
			marked = line
		}
	}
	if !strings.Contains(marked, "to the Grey Flats") {
		t.Errorf("the highlighted line is %q, want the second party", marked)
	}
	if !strings.Contains(text, "Haul the salt from the Grey Flats") || strings.Contains(text, "Count the herds") {
		t.Errorf("the field report does not show the highlighted party:\n%s", text)
	}
}

func TestFieldReportEmptyState(t *testing.T) {
	engine, _ := expFixture(t)
	state := engine.GetState()
	offers := 0
	for _, slot := range state.Military.Board {
		if slot.Offer != nil {
			offers++
		}
	}
	want := fmt.Sprintf("Nobody is out. The Ready Room has %d %s waiting.", offers, pluralize("offer", offers))
	// On the narrower card the sentence wraps, so only its first words are
	// checked there.
	for _, sz := range [][2]int{{100, 30}, {160, 45}} {
		text := renderExpedition(state, expView{report: true}, sz[0], sz[1]).String()
		if !strings.Contains(text, "Nobody is out.") {
			t.Errorf("%dx%d: the empty state lacks its first sentence:\n%s", sz[0], sz[1], text)
		}
		if sz[0] == 160 && !strings.Contains(text, want) {
			t.Errorf("%dx%d: the empty state lacks %q:\n%s", sz[0], sz[1], want, text)
		}
	}
}

func TestReportLinesFitEverySize(t *testing.T) {
	state := fieldState(t)
	rare := state
	rare.Military.Active = []*game.ExpeditionSnapshot{{}}
	first := *state.Military.Active[0]
	first.Rare = true
	rare.Military.Active[0] = &first
	for _, sz := range [][2]int{{80, 24}, {100, 30}, {160, 45}, {240, 60}} {
		for _, st := range []game.GameState{state, rare} {
			g := renderExpedition(st, expView{report: true, marked: 1}, sz[0], sz[1])
			name := fmt.Sprintf("%dx%d rare=%v", sz[0], sz[1], st.Military.Active[0].Rare)
			lines := g.Lines()
			if len(lines) != sz[1] {
				t.Errorf("%s: %d rows, want %d", name, len(lines), sz[1])
			}
			for y := 1; y < sz[1]-1; y++ {
				if g.at(0, y).r != '│' || g.at(sz[0]-1, y).r != '│' {
					t.Errorf("%s: row %d runs through the frame: %q", name, y, lines[y])
				}
			}
			for y, line := range lines {
				if runeLen(line) > sz[0] {
					t.Errorf("%s: row %d is %d cells wide", name, y, runeLen(line))
				}
			}
			if sz[0] >= 160 && !strings.Contains(g.String(), "dry ford") {
				t.Errorf("%s: the oldest road event is cut", name)
			}
		}
	}
	if text := renderExpedition(state, expView{report: true}, 80, 24).String(); !strings.Contains(text, "Out · 1") {
		t.Errorf("80x24 lost the compact Out card:\n%s", text)
	}
}
