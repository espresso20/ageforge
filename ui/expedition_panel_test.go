package ui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
)

// offerNumber matches an offer number as a list prints it: "3. Scout".
var offerNumber = regexp.MustCompile(`(^|\W)\d+\.\s`)

// expFixture is a Bronze Age town with a board, and a panel on it.
func expFixture(t *testing.T) (*game.GameEngine, *expeditionPanel) {
	t.Helper()
	engine := newExpeditionTestEngine(t)
	engine.EnsureMissionBoard()
	p := newExpeditionPanel()
	p.engine = engine
	p.setSize(100, 27)
	now := time.Unix(1_000_000, 0)
	p.now, p.start = func() time.Time { return now }, now
	return engine, p
}

// panelText is the panel as the player sees it.
func panelText(p *expeditionPanel) string {
	return renderExpedition(p.state, p.view, p.w, p.h).String()
}

// runUntilHome steps the game until nothing is out.
func runUntilHome(t *testing.T, engine *game.GameEngine) {
	t.Helper()
	for i := 0; ; i++ {
		if len(engine.GetState().Military.Active) == 0 {
			return
		}
		if i > 200000 {
			t.Fatal("the mission never came home")
		}
		engine.StepTicks(1)
	}
}

func TestExpeditionPanelOpensOnTheRightView(t *testing.T) {
	engine, p := expFixture(t)

	p.open(engine.GetState())
	if p.view.report {
		t.Error("with nobody out and nothing new the panel opens on the Report")
	}
	if text := panelText(p); !strings.Contains(text, "The lands of the Bronze Age") {
		t.Errorf("the Ready Room has no map:\n%s", text)
	}

	if err := engine.LaunchExpedition(1); err != nil {
		t.Fatal(err)
	}
	p.open(engine.GetState())
	if !p.view.report {
		t.Error("with a party out the panel opens on the Ready Room, not the Report")
	}
	if text := panelText(p); !strings.Contains(text, "Parties in the field") {
		t.Errorf("the Report has no map of the field:\n%s", text)
	}

	runUntilHome(t, engine)
	if mil := engine.GetState().Military; mil.ReturnsMade != 1 || mil.ReturnsSeen != 0 {
		t.Fatalf("returns made %d, seen %d; want 1, 0", mil.ReturnsMade, mil.ReturnsSeen)
	}
	p.open(engine.GetState())
	if !p.view.report {
		t.Error("with a mission home since the last look the panel opens on the Ready Room, not the Report")
	}
	if mil := engine.GetState().Military; mil.ReturnsSeen != mil.ReturnsMade {
		t.Errorf("viewing the Report left %d of %d unseen", mil.ReturnsMade-mil.ReturnsSeen, mil.ReturnsMade)
	}
	p.open(engine.GetState())
	if p.view.report {
		t.Error("with everything seen and nobody out the panel opens on the Report, not the Ready Room")
	}
}

func TestExpeditionPanelTabSwitchesAndClearsTheNewCount(t *testing.T) {
	engine, p := expFixture(t)
	if err := engine.LaunchExpedition(1); err != nil {
		t.Fatal(err)
	}
	runUntilHome(t, engine)
	p.state, p.view = engine.GetState(), expView{}
	if text := panelText(p); !strings.Contains(text, "● 1 new") {
		t.Errorf("the Ready Room does not say a mission came home:\n%s", text)
	}

	p.handleKey(tcell.KeyTab)
	if !p.view.report {
		t.Fatal("Tab did not open the Report")
	}
	text := panelText(p)
	if strings.Contains(text, "new") {
		t.Errorf("the Report still counts the mission as new in the header:\n%s", text)
	}
	if line := lineContaining(text, "came back"); !strings.Contains(line, "●") {
		if line = lineContaining(text, "failed"); !strings.Contains(line, "●") {
			t.Errorf("the Report does not mark the return that is new:\n%s", text)
		}
	}
	if mil := engine.GetState().Military; mil.ReturnsSeen != mil.ReturnsMade {
		t.Error("viewing the Report did not clear the new count in the game")
	}

	p.handleKey(tcell.KeyTab)
	if p.view.report {
		t.Fatal("Tab did not go back to the Ready Room")
	}
	p.update(engine.GetState())
	if text := panelText(p); strings.Contains(text, "new") {
		t.Errorf("the new count came back:\n%s", text)
	}
}

func TestHopToPicksTheNearestPinThatWay(t *testing.T) {
	pins := []pt{{20, 10}, {30, 10}, {26, 4}, {40, 11}, {10, 10}, {21, 18}}
	for _, c := range []struct {
		dx, dy, want int
	}{
		{1, 0, 1},  // the pin 10 along the row, before the one 20 along
		{-1, 0, 4}, // the only one to the left
		{0, -1, 2}, // the only one above
		{0, 1, 5},  // the only one below
	} {
		if got := hopTo(pins, 0, c.dx, c.dy); got != c.want {
			t.Errorf("hopping (%d,%d) from pin 0 lands on %d, want %d", c.dx, c.dy, got, c.want)
		}
	}
	if got := hopTo(pins, 3, 1, 0); got != -1 {
		t.Errorf("hopping right from the last pin on the right lands on %d, want nowhere", got)
	}
	if got := hopTo(pins, 0, 1, 0); got == 3 {
		t.Error("a pin farther along the same row was picked over a nearer one")
	}
}

func TestExpeditionPanelArrowsHopBetweenPins(t *testing.T) {
	engine, p := expFixture(t)
	p.open(engine.GetState())
	sc := p.scene()
	if len(sc.offers) != 9 {
		t.Fatalf("the board has %d pins, want 9", len(sc.offers))
	}
	cells := sc.pinCells()
	for _, c := range []struct {
		key    tcell.Key
		dx, dy int
	}{{tcell.KeyRight, 1, 0}, {tcell.KeyLeft, -1, 0}, {tcell.KeyDown, 0, 1}, {tcell.KeyUp, 0, -1}} {
		for from := range cells {
			p.selIdx, p.view.sel = from, sc.offers[from].key
			p.handleKey(c.key)
			// The pin it lands on lies that way, and no pin that way is
			// nearer on the lab's measure.
			score := func(q pt) int {
				ddx, ddy := q.x-cells[from].x, (q.y-cells[from].y)*2
				along := ddx*c.dx + ddy*c.dy
				if along <= 0 {
					return -1
				}
				return along + (abs(ddx*c.dy)+abs(ddy*c.dx))*2
			}
			best := -1
			for i, q := range cells {
				if s := score(q); s >= 0 && (best < 0 || s < score(cells[best])) {
					best = i
				}
			}
			want := from
			if best >= 0 {
				want = best
			}
			if p.selIdx != want {
				t.Errorf("key %v from pin %d at %v lands on %d, want %d", c.key, from, cells[from], p.selIdx, want)
			}
		}
	}
}

func TestExpeditionPanelEnterEnterSends(t *testing.T) {
	engine, p := expFixture(t)
	p.open(engine.GetState())
	sc := p.scene()
	want := *sc.offers[sc.selected(p.view.sel)].offer
	if !want.CanLaunch {
		t.Fatalf("the first offer cannot be sent: %s", want.LaunchBlockReason)
	}

	p.handleKey(tcell.KeyEnter)
	if !p.view.confirm {
		t.Fatal("Enter did not ask before sending")
	}
	if text := panelText(p); !strings.Contains(text, "Send this "+want.Label+"?") {
		t.Errorf("the question is not on the screen:\n%s", text)
	}
	if mil := engine.GetState().Military; mil.ActiveShort != nil || mil.ActiveLong != nil || mil.ActiveGrand != nil {
		t.Fatal("the first Enter already sent the party")
	}
	if !p.closeConfirm() || p.view.confirm || p.closeConfirm() {
		t.Error("Esc does not take back the question, once")
	}

	p.handleKey(tcell.KeyEnter)
	p.handleKey(tcell.KeyEnter)
	if p.view.confirm {
		t.Error("the question is still up after the party was sent")
	}
	mil := engine.GetState().Military
	if mil.ActiveShort == nil || mil.ActiveShort.Name != want.Name {
		t.Fatalf("under way: %+v, want %q", mil.ActiveShort, want.Name)
	}
	if !strings.HasPrefix(p.view.flash, "Sent: ") || !p.view.flashGood {
		t.Errorf("the key bar says %q (good: %v)", p.view.flash, p.view.flashGood)
	}
	if text := panelText(p); !strings.Contains(text, p.view.flash) {
		t.Errorf("the key bar's line is not on the screen:\n%s", text)
	}
}

func TestExpeditionPanelShowsAnOfferThatCannotBeSentItsRefusal(t *testing.T) {
	engine, p := expFixture(t)
	if err := engine.LaunchExpedition(1); err != nil {
		t.Fatal(err)
	}
	p.open(engine.GetState())
	p.view.report = false
	p.update(engine.GetState())
	// Another short offer: only 3 soldiers are at home, and a party is 8.
	engine.SetGarrisonForTest("bronze_age", 11)
	p.update(engine.GetState())
	sc := p.scene()
	i := -1
	for j, it := range sc.offers {
		if it.tier == game.TierShort {
			i = j
		}
	}
	if i < 0 {
		t.Fatal("no short offer left on the board")
	}
	p.selIdx, p.view.sel = i, sc.offers[i].key
	o := *sc.offers[i].offer
	if o.CanLaunch || o.LaunchBlockReason == "" {
		t.Fatalf("a second short mission can be sent: %+v", o)
	}
	text := panelText(p)
	l := layoutFor(p.w, p.h)
	g := renderExpedition(p.state, p.view, p.w, p.h)
	var card []string
	for y := l.mainY; y < l.mainY+l.mainH; y++ {
		card = append(card, strings.TrimSpace(string(runesOf(g, l.cardX+1, y, l.cardW-2))))
	}
	if joined := strings.Join(strings.Fields(strings.Join(card, " ")), " "); !strings.Contains(joined, refusalText(o)) {
		t.Errorf("the card does not say why: want %q in\n%s", refusalText(o), text)
	}
	if strings.Contains(text, "send the party") {
		t.Errorf("the card offers to send an offer that cannot be sent:\n%s", text)
	}
	p.handleKey(tcell.KeyEnter)
	if p.view.confirm {
		t.Error("Enter asked to send an offer that cannot be sent")
	}
	if p.view.flash != refusalText(o) || p.view.flashGood {
		t.Errorf("the key bar says %q (good: %v), want the refusal", p.view.flash, p.view.flashGood)
	}
}

func TestExpeditionPinsAreStableAndApart(t *testing.T) {
	engine, p := expFixture(t)
	p.open(engine.GetState())
	for _, sz := range [][2]int{{100, 30}, {100, 27}, {80, 24}, {80, 21}, {120, 40}} {
		p.setSize(sz[0], sz[1])
		if a, b := panelText(p), panelText(p); a != b {
			t.Errorf("two renders at %dx%d differ", sz[0], sz[1])
		}
		l := layoutFor(sz[0], sz[1])
		w, h := l.mapW-2, l.mainH-2
		a, b := buildScene(p.state, w, h), buildScene(p.state, w, h)
		town := townAt(w, h)
		var all []expPinItem
		all = append(append(all, a.acts...), a.offers...)
		for i, it := range all {
			if !it.placed {
				t.Errorf("%dx%d: pin %d found no room", sz[0], sz[1], i)
				continue
			}
			if it.at.x < 1 || it.at.x > w-2 || it.at.y < 0 || it.at.y > h-1 {
				t.Errorf("%dx%d: pin %d at %v is off the map", sz[0], sz[1], i, it.at)
			}
			if a.terrain.isWater(it.at.x, it.at.y) {
				t.Errorf("%dx%d: pin %d stands in the sea", sz[0], sz[1], i)
			}
			if it.at == town {
				t.Errorf("%dx%d: pin %d stands on the town", sz[0], sz[1], i)
			}
			for j := i + 1; j < len(all); j++ {
				if o := all[j]; o.at.y == it.at.y && abs(o.at.x-it.at.x) <= 3 {
					t.Errorf("%dx%d: pins %d and %d overlap at %v and %v", sz[0], sz[1], i, j, it.at, o.at)
				}
			}
		}
		for i := range a.offers {
			if a.offers[i].at != b.offers[i].at {
				t.Errorf("%dx%d: pin %d moved between two builds", sz[0], sz[1], i)
			}
		}
	}
}

// TestPinsLieFartherOutByTier verifies short offers stand near the town,
// long ones in the middle and grand ones far: judged over many offers, on
// the cells' own scale.
func TestPinsLieFartherOutByTier(t *testing.T) {
	const w, h = 60, 19
	terrain := buildTerrain(7, "bronze_age", w, h)
	town := townAt(w, h)
	mean := map[string]float64{}
	for _, tier := range []string{game.TierShort, game.TierLong, game.TierGrand} {
		sum, n := 0.0, 0
		for i := 0; i < 40; i++ {
			at, ok := placePins(terrain, []pinSpec{{pinHash(fmt.Sprintf("%s offer %d", tier, i), 100+i), tier}})
			if !ok[0] {
				t.Fatalf("a lone %s pin found no room", tier)
			}
			dx, dy := float64(at[0].x-town.x)/float64(town.x), float64(at[0].y-town.y)/float64(town.y)
			sum += dx*dx + dy*dy
			n++
		}
		mean[tier] = sum / float64(n)
	}
	if !(mean[game.TierShort] < mean[game.TierLong] && mean[game.TierLong] < mean[game.TierGrand]) {
		t.Errorf("mean distance from the town by tier: %v; want short < long < grand", mean)
	}
}

func TestTerrainHoldsForTheAgeAndChangesWithIt(t *testing.T) {
	a, b := buildTerrain(5, "bronze_age", 60, 19), buildTerrain(5, "bronze_age", 60, 19)
	if string(a.r) != string(b.r) {
		t.Error("the same run and age drew different land")
	}
	if c := buildTerrain(5, "iron_age", 60, 19); string(a.r) == string(c.r) {
		t.Error("the next age drew the same land")
	}
	if c := buildTerrain(6, "bronze_age", 60, 19); string(a.r) == string(c.r) {
		t.Error("another run drew the same land")
	}
}

func TestExpeditionPanelFitsItsSizes(t *testing.T) {
	engine, _ := expFixture(t)
	for _, n := range []int{1, 4, 7} {
		if err := engine.LaunchExpedition(n); err != nil {
			t.Fatal(err)
		}
	}
	runUntilHome(t, engine)
	for _, n := range []int{1, 4, 7} {
		if err := engine.LaunchExpedition(n); err != nil {
			t.Logf("launching again: %v", err)
		}
	}
	state := engine.GetState()
	// A return with a long haul line, to test the cut.
	long := game.MissionReturn{Label: "Trade Mission", Tier: game.TierGrand, Place: "Clay Terraces", Success: true,
		Haul: strings.Repeat("1.14M wood, 408K stone, ", 8), Tick: 5}
	state.Military.Returns = append([]game.MissionReturn{long, long}, state.Military.Returns...)
	state.Military.ReturnsMade += 2
	for _, sz := range [][2]int{{80, 24}, {100, 30}, {80, 21}, {100, 27}, {120, 40}, {60, 14}, {76, 20}} {
		for _, report := range []bool{false, true} {
			for _, confirm := range []bool{false, true} {
				for _, rare := range []bool{false, true} {
					st := state
					if rare {
						st.Military.Board = append([]game.BoardSlotInfo(nil), st.Military.Board...)
						for i := range st.Military.Board {
							if o := st.Military.Board[i].Offer; o != nil {
								c := *o
								c.Rare = true
								st.Military.Board[i].Offer = &c
							}
						}
					}
					g := renderExpedition(st, expView{report: report, confirm: confirm && !report, marked: 2,
						flash: strings.Repeat("A long line for the key bar. ", 6)}, sz[0], sz[1])
					name := fmt.Sprintf("%dx%d report=%v confirm=%v rare=%v", sz[0], sz[1], report, confirm, rare)
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
						if offerNumber.MatchString(line) {
							t.Errorf("%s: row %d carries an offer number: %q", name, y, line)
						}
					}
				}
			}
		}
	}
}

func TestExpeditionRoadStoriesFitCompactReports(t *testing.T) {
	engine, _ := expFixture(t)
	state := engine.GetState()
	state.Military.ActiveShort = &game.ExpeditionSnapshot{Label: "Scout", Place: "Tin Road", TotalTicks: 100, TicksLeft: 50, SuccessChance: .65, HaulMinutes: 20,
		RoadStories: []string{"The party found an amber hoard. The haul grows by 40%.", "A friendly village showed the party a shorter way. 6 minutes saved.", "Bandits shadowed the party for days. The odds fall to 57%."}}
	state.Military.Active = []*game.ExpeditionSnapshot{state.Military.ActiveShort}
	for _, sz := range [][2]int{{80, 24}, {100, 30}} {
		g := renderExpedition(state, expView{report: true}, sz[0], sz[1])
		lines := g.Lines()
		if len(lines) != sz[1] {
			t.Fatalf("%dx%d produced %d rows", sz[0], sz[1], len(lines))
		}
		for y, line := range lines {
			if runeLen(line) > sz[0] {
				t.Errorf("%dx%d row %d is too wide: %d", sz[0], sz[1], y, runeLen(line))
			}
		}
		if !strings.Contains(g.String(), "On the road") {
			t.Errorf("%dx%d omitted the road event heading", sz[0], sz[1])
		}
		for _, storyWord := range []string{"amber", "friendly village", "Bandits"} {
			if !strings.Contains(g.String(), storyWord) {
				t.Errorf("%dx%d omitted newest road story %q:\n%s", sz[0], sz[1], storyWord, g.String())
			}
		}
	}
}

func TestExpeditionPanelNeverNumbersAnOffer(t *testing.T) {
	engine, p := expFixture(t)
	p.open(engine.GetState())
	for _, report := range []bool{false, true} {
		p.view.report = report
		if text := panelText(p); offerNumber.MatchString(text) {
			t.Errorf("report=%v: the panel numbers an offer:\n%s", report, text)
		}
	}
	if list := cmdExpedition([]string{"list"}, engine).Message; offerNumber.MatchString(list) {
		t.Errorf("expedition list numbers an offer:\n%s", list)
	}
}

func TestExpeditionPanelFoldsToASCII(t *testing.T) {
	engine, p := expFixture(t)
	p.open(engine.GetState())
	g := renderExpedition(p.state, p.view, 100, 27)
	g.fold()
	for y, line := range g.Lines() {
		for _, r := range line {
			if r > 0x7e {
				t.Fatalf("row %d keeps %q on the plain glyph tier: %q", y, r, line)
			}
		}
	}
	_ = mapmodel.TierASCII
}

func TestExpeditionPanelKeysBelongToThePromptWhenItHasText(t *testing.T) {
	_, p := expFixture(t)
	for _, key := range []tcell.Key{tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight, tcell.KeyTab, tcell.KeyEnter} {
		if p.routeKey(tcell.NewEventKey(key, 0, tcell.ModNone), "cop") {
			t.Errorf("key %v was taken with something typed", key)
		}
		if !p.routeKey(tcell.NewEventKey(key, 0, tcell.ModNone), "") {
			t.Errorf("key %v was left to the empty prompt", key)
		}
	}
	if p.routeKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone), "") {
		t.Error("a letter was taken from the prompt")
	}
}

func TestExpeditionReportListsWhatCameHome(t *testing.T) {
	engine, p := expFixture(t)
	if err := engine.LaunchExpedition(1); err != nil {
		t.Fatal(err)
	}
	runUntilHome(t, engine)
	p.open(engine.GetState())
	text := panelText(p)
	rets := engine.GetState().Military.Returns
	if len(rets) != 1 {
		t.Fatalf("returns: %+v", rets)
	}
	if !strings.Contains(text, "Came home") || !strings.Contains(text, goingTo(rets[0].Label, rets[0].Place)) {
		t.Errorf("the Report does not list the mission that came home:\n%s", text)
	}
	if line := lineContaining(text, goingTo(rets[0].Label, rets[0].Place)); !strings.Contains(line, "●") {
		t.Errorf("the new return is not marked: %q", line)
	}
	// A mission under way, then its empty state once it is home.
	if text := expeditionReportText(game.GameState{}); !strings.Contains(text, "Nobody is out.") {
		t.Errorf("the empty Report does not say nobody is out:\n%s", text)
	}
}

func TestBriefingColoursTheCostByWhatYouHave(t *testing.T) {
	state := game.GameState{Resources: map[string]game.ResourceState{"wood": {Amount: 10}, "stone": {Amount: 5000}}}
	o := game.ExpeditionInfo{Name: "Scout: count the herds", CanLaunch: true, Cost: map[string]float64{"wood": 1000, "stone": 100}}
	body, _ := briefing(state, o, 33)
	kinds := map[string]xk{}
	for _, line := range body {
		for _, part := range line.parts {
			kinds[part.s] = part.k
		}
	}
	if kinds["1000 wood"] != xBad && kinds["1K wood"] != xBad {
		t.Errorf("a cost you cannot pay is not in the refusal colour: %v", kinds)
	}
	if kinds["100 stone"] != xGood {
		t.Errorf("a cost you can pay is not green: %v", kinds)
	}
}

// runesOf is w cells of row y from column x.
func runesOf(g *xgrid, x, y, w int) []rune {
	var out []rune
	for i := 0; i < w; i++ {
		out = append(out, g.at(x+i, y).r)
	}
	return out
}

// Every offer's briefing carries a Party line: green when the soldiers at home
// cover the party, the refusal colour when they do not.
func TestBriefingShowsThePartyOfEveryTier(t *testing.T) {
	engine, _ := expFixture(t)
	partyLine := func(state game.GameState, o game.ExpeditionInfo) (text string, kind xk) {
		body, _ := briefing(state, o, 40)
		for _, ln := range body {
			if len(ln.parts) == 2 && strings.TrimSpace(ln.parts[0].s) == "Party" {
				return ln.parts[1].s, ln.parts[1].k
			}
		}
		return "", xText
	}
	state := engine.GetState()
	if len(state.Military.Expeditions) != 9 {
		t.Fatalf("%d offers on the board, want 9", len(state.Military.Expeditions))
	}
	for _, o := range state.Military.Expeditions {
		text, kind := partyLine(state, o)
		if want := FormatNumber(float64(o.SoldiersNeeded)) + " soldiers"; text != want || o.SoldiersNeeded < 1 {
			t.Errorf("offer %d (%s): Party line %q, want %q", o.Number, o.Tier, text, want)
		}
		if kind != xGood {
			t.Errorf("offer %d: the Party line is not green with %d at home for %d", o.Number, state.Military.SoldierCount, o.SoldiersNeeded)
		}
	}
	// With too few at home the same lines turn to the refusal colour.
	engine.SetGarrisonForTest("bronze_age", 7)
	state = engine.GetState()
	for _, o := range state.Military.Expeditions {
		if _, kind := partyLine(state, o); kind != xBad {
			t.Errorf("offer %d: the Party line is not red with 7 at home for %d", o.Number, o.SoldiersNeeded)
		}
	}
}

// The Ready Room header reads home over posted, the send box says how many
// stay home, the Report shows each party's soldiers and what fell, and the
// People box counts soldiers out: all at 80x24 as well as 100x30.
func TestPartiesOnEveryScreen(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {100, 30}} {
		engine, p := expFixture(t)
		if err := engine.LaunchExpedition(1); err != nil {
			t.Fatal(err)
		}
		state := engine.GetState()
		away := state.Military.SoldiersAway
		if away != 8 {
			t.Fatalf("%d soldiers away after one short mission in the Bronze Age, want 8", away)
		}
		ready := renderExpedition(state, expView{}, sz[0], sz[1]).String()
		if want := fmt.Sprintf("Soldiers %d of %d at home", 500-away, 500); !strings.Contains(ready, want) {
			t.Errorf("%dx%d: the header lacks %q:\n%s", sz[0], sz[1], want, ready)
		}
		if !strings.Contains(ready, "Party") {
			t.Errorf("%dx%d: the Ready Room shows no Party line:\n%s", sz[0], sz[1], ready)
		}
		p.setSize(sz[0], sz[1])
		p.open(state)
		p.view.report = false
		p.update(state)
		confirm := renderExpedition(state, expView{confirm: true, sel: p.view.sel}, sz[0], sz[1]).String()
		if !strings.Contains(confirm, fmt.Sprintf("stay home)")) || !strings.Contains(confirm, "Party") {
			t.Errorf("%dx%d: the send box lacks the party or who stays home:\n%s", sz[0], sz[1], confirm)
		}
		report := renderExpedition(state, expView{report: true}, sz[0], sz[1]).String()
		if !strings.Contains(report, "8 soldiers") {
			t.Errorf("%dx%d: the Report block lacks the party's soldiers:\n%s", sz[0], sz[1], report)
		}
		box := plainText(peopleBoxText(state, sidebarW-2, false))
		if !strings.Contains(strings.Join(strings.Fields(box), " "), "· 8 out") {
			t.Errorf("%dx%d: the People box lacks the soldiers out:\n%s", sz[0], sz[1], box)
		}
	}
}

// The returns lines say how many soldiers fell, or that all came home.
func TestReturnLinesCountTheFallen(t *testing.T) {
	r := game.MissionReturn{Label: "Skirmish", Place: "the Amber Coast", Party: 60, Fell: 36}
	if line := returnLine(r); !strings.Contains(line, "failed.") || !strings.Contains(line, "36 of 60 soldiers fell.") {
		t.Errorf("failed return: %q", line)
	}
	r = game.MissionReturn{Label: "Scout", Place: "the Cedar Pass", Success: true, Haul: "5 wood", Party: 8}
	if line := returnLine(r); !strings.Contains(line, "All 8 soldiers came home.") {
		t.Errorf("clean return: %q", line)
	}
	r.Fell = 1
	if line := returnLine(r); !strings.Contains(line, "1 of 8 soldiers fell.") {
		t.Errorf("a success that lost one: %q", line)
	}
}
