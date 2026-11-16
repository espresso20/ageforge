package ui

import (
	"fmt"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/window/kit"
)

// cardTown is a dashboard on a fresh game in its first age with a big stock
// of what the first buildings cost, laid out on a 160 by 45 screen.
func cardTown(t *testing.T) (*Dashboard, *stage, *game.GameEngine) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	eng := game.NewGameEngine()
	for _, r := range []string{"food", "wood", "stone"} {
		eng.SetStockForTest(r, 1e6)
	}
	d, pages := newTestDashboard(eng)
	d.refresh()
	st := newStage(pages, 160, 45)
	d.refresh()
	st.paint()
	return d, st, eng
}

// medievalTown is a dashboard in the Medieval Age, where the prices take
// several resources, on a screen w by h.
func medievalTown(t *testing.T, w, h int) (*Dashboard, *stage, *game.GameEngine) {
	t.Helper()
	d, pages, eng := stagedDashboard(t, "medieval_age")
	d.refresh()
	st := newStage(pages, w, h)
	d.refresh()
	st.paint()
	return d, st, eng
}

// setStock sets the stock of every resource to the same amount.
func setStock(eng *game.GameEngine, amount float64) {
	for r := range eng.GetState().Resources {
		eng.SetStockForTest(r, amount)
	}
}

// placedOf is the placed card of key on the screen now.
func placedOf(t *testing.T, d *Dashboard, key string) placedCard {
	t.Helper()
	for _, pc := range d.economyTab.buildBox.placed {
		if pc.card.key == key {
			return pc
		}
	}
	t.Fatalf("the Buildings box shows no card for %s", key)
	return placedCard{}
}

// centre is the screen cell at the middle of a rectangle given in the box's cells.
func centre(d *Dashboard, r image.Rectangle) image.Point {
	x, y, _, _ := d.economyTab.buildBox.GetRect()
	return image.Pt(x+(r.Min.X+r.Max.X-1)/2, y+(r.Min.Y+r.Max.Y-1)/2)
}

// buildPoint is the cell of a card's Build button.
func buildPoint(t *testing.T, d *Dashboard, key string) image.Point {
	t.Helper()
	return centre(d, d.economyTab.buildBox.cardKeys.keys[placedOf(t, d, key).build].r)
}

// buyPoint is the cell of a buy choice in the strip.
func buyPoint(t *testing.T, d *Dashboard, c buyChoice) image.Point {
	t.Helper()
	for _, k := range d.economyTab.buildBox.strip.keys {
		if k.key == buyKey(c) {
			return centre(d, k.r)
		}
	}
	t.Fatalf("the strip has no key for choice %d", c)
	return image.Point{}
}

// ownedAndQueued is the copies of key built or under way.
func ownedAndQueued(eng *game.GameEngine, key string) int {
	st := eng.GetState()
	n := st.Buildings[key].Count
	for _, q := range st.BuildQueue {
		if q.Key == key {
			n++
		}
	}
	return n
}

// A click on Build runs exactly the build command for the amount chosen,
// once, through the command path (the history has it), and builds that many.
// An off button runs nothing and its card says why.
func TestAClickOnBuildRunsTheBuildCommand(t *testing.T) {
	d, st, eng := cardTown(t)
	const key = "gathering_camp"
	for _, c := range []struct {
		buy  buyChoice
		cmd  string
		want func(game.GameState) int
	}{
		{buyOne, "build gathering_camp", func(game.GameState) int { return 1 }},
		{buyTen, "build gathering_camp 10", func(game.GameState) int { return 10 }},
		{buyMax, "build gathering_camp max", func(s game.GameState) int { return s.MaxBuildable(key) }},
	} {
		for _, r := range []string{"food", "wood", "stone"} {
			eng.SetStockForTest(r, 1e6)
		}
		d.refresh()
		st.paint()
		st.click(buyPoint(t, d, c.buy).X, buyPoint(t, d, c.buy).Y)
		d.refresh()
		st.paint()
		if d.economyTab.buildBox.buy != c.buy {
			t.Fatalf("a click on the buy key left the choice at %d, want %d", d.economyTab.buildBox.buy, c.buy)
		}
		if got := placedOf(t, d, key).card.cmd; got != c.cmd {
			t.Fatalf("with choice %d the card carries %q, want %q", c.buy, got, c.cmd)
		}
		before, runs, want := ownedAndQueued(eng, key), len(d.cmdHistory), c.want(eng.GetState())
		at := buildPoint(t, d, key)
		st.click(at.X, at.Y)
		if len(d.cmdHistory) != runs+1 || d.cmdHistory[len(d.cmdHistory)-1] != c.cmd {
			t.Errorf("a click on Build ran %v, want exactly one run of %q", d.cmdHistory[runs:], c.cmd)
		}
		if got := ownedAndQueued(eng, key) - before; got != want || want == 0 {
			t.Errorf("%q built %d copies, want %d", c.cmd, got, want)
		}
	}

	// A card that cannot be built: its button is off, a click runs nothing,
	// and the card says what is short.
	eng.SetStockForTest("wood", eng.GetState().Buildings["hut"].NextCost["wood"]*0.6)
	d.refresh()
	st.paint()
	pc := placedOf(t, d, "hut")
	if pc.card.canBuild || pc.card.reason == "" || !strings.HasPrefix(pc.card.reason, "short of ") {
		t.Fatalf("staging: the hut card is %+v", pc.card)
	}
	if !d.economyTab.buildBox.cardKeys.keys[pc.build].off {
		t.Error("the Build button of a card that cannot be built is on")
	}
	runs, before := len(d.cmdHistory), ownedAndQueued(eng, "hut")
	at := buildPoint(t, d, "hut")
	st.click(at.X, at.Y)
	if len(d.cmdHistory) != runs || ownedAndQueued(eng, "hut") != before {
		t.Errorf("a click on an off Build ran %v", d.cmdHistory[runs:])
	}
}

// A click on a card, anywhere but its button, picks it; the picked card's
// command is the dim line of an empty prompt, and Enter on the empty prompt
// runs it. With text in the prompt Enter runs the text; with the pointer on
// another button, whose own command the prompt shows, Enter runs nothing.
func TestThePickedCardIsThePromptsPreview(t *testing.T) {
	d, st, eng := cardTown(t)
	box := d.economyTab.buildBox
	if line := d.inputField.shadow(); line != "" {
		t.Fatalf("with no card picked the empty prompt shows %q", line)
	}
	pc := placedOf(t, d, "gathering_camp")
	x, y, _, _ := box.GetRect()
	st.click(x+3, y+pc.y) // the card's left edge, on its first row
	if box.picked != "gathering_camp" {
		t.Fatalf("a click on the card picked %q", box.picked)
	}
	if got := d.inputField.shadow(); got != "build gathering_camp" {
		t.Errorf("the empty prompt shows %q for the picked card", got)
	}
	// Typing replaces the preview with the typist's line.
	d.inputField.SetText("roster")
	if got := d.inputField.in.Ghost("roster"); strings.Contains(got, "gathering_camp") {
		t.Errorf("with text in the prompt the ghost is %q", got)
	}
	runs, before := len(d.cmdHistory), ownedAndQueued(eng, "gathering_camp")
	d.submitInput()
	if len(d.cmdHistory) != runs+1 || d.cmdHistory[runs] != "roster" || ownedAndQueued(eng, "gathering_camp") != before {
		t.Errorf("Enter with text ran %v and built %d", d.cmdHistory[runs:], ownedAndQueued(eng, "gathering_camp")-before)
	}
	// The pointer on another card's Build button: its command is the preview,
	// and Enter on the empty prompt runs neither.
	other := buildPoint(t, d, "hut")
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: other.X, Y: other.Y})
	if got := d.inputField.shadow(); got != "build hut" {
		t.Errorf("with the pointer on a Build button the prompt shows %q", got)
	}
	runs = len(d.cmdHistory)
	d.submitInput()
	if len(d.cmdHistory) != runs {
		t.Errorf("Enter with another button's command shown ran %v", d.cmdHistory[runs:])
	}
	// The pointer away, Enter runs the picked card's command.
	st.mouse(kit.Mouse{Kind: kit.MouseLeave, X: 0, Y: 0})
	if got := d.inputField.shadow(); got != "build gathering_camp" {
		t.Fatalf("with the pointer away the prompt shows %q", got)
	}
	d.submitInput()
	if len(d.cmdHistory) != runs+1 || d.cmdHistory[runs] != "build gathering_camp" || ownedAndQueued(eng, "gathering_camp") != before+1 {
		t.Errorf("Enter on the empty prompt ran %v and built %d, want one build of the picked card", d.cmdHistory[runs:], ownedAndQueued(eng, "gathering_camp")-before)
	}
	// Only the picked card's button carries a key.
	d.refresh()
	st.paint()
	keys := 0
	for _, pc := range box.placed {
		if k := box.cardKeys.keys[pc.build]; !k.cap.Empty() {
			keys++
			if pc.card.key != "gathering_camp" {
				t.Errorf("the button of %s carries a key", pc.card.key)
			}
		}
	}
	if keys != 1 {
		t.Errorf("%d Build buttons carry a key, want the picked card's alone", keys)
	}
	// A panel over the dashboard takes the preview away.
	d.inputField.echo = ""
	d.runForTest("help")
	if got := d.inputField.shadow(); got != "" {
		t.Errorf("with a panel open the prompt shows %q", got)
	}
}

// The gather buttons are there while the age gathers, one for each resource
// that can be gathered, and a click gathers.
func TestGatherButtonsInTheAgesThatGather(t *testing.T) {
	d, st, eng := cardTown(t)
	box := d.economyTab.buildBox
	var cmds []string
	for _, k := range box.strip.keys {
		if strings.HasPrefix(k.cmd, "gather ") {
			cmds = append(cmds, k.cmd)
		}
	}
	if len(cmds) < 2 || cmds[0] != "gather food" || cmds[1] != "gather wood" {
		t.Fatalf("the first age's strip has the gather commands %v", cmds)
	}
	before := eng.GetState().Resources["food"].Amount
	for _, k := range box.strip.keys {
		if k.cmd == "gather food" {
			p := centre(d, k.r)
			runs := len(d.cmdHistory)
			st.click(p.X, p.Y)
			if len(d.cmdHistory) != runs+1 || d.cmdHistory[runs] != "gather food" {
				t.Errorf("a click on Gather ran %v", d.cmdHistory[runs:])
			}
		}
	}
	if eng.GetState().Resources["food"].Amount <= before {
		t.Error("a click on Gather food gathered nothing")
	}
	// Past the Medieval Age the command no longer works and the buttons go.
	d2, _, _ := stagedDashboard(t, "renaissance_age")
	d2.refresh()
	drawnDashboard(t, d2, d2.pages, 160, 45)
	for _, k := range d2.economyTab.buildBox.strip.keys {
		if strings.HasPrefix(k.cmd, "gather ") {
			t.Errorf("the Renaissance Age strip has %q", k.cmd)
		}
	}
}

// A card's price lights each resource the player holds enough of and leaves
// the rest dull; the gauge is how much of the price is in hand; a card that
// cannot be built says what it is furthest from and how long that takes at the
// current rate, or that there is no income.
func TestCardPriceGaugeAndWait(t *testing.T) {
	_, _, eng := stagedDashboard(t, "medieval_age")
	state := eng.GetState()
	var key string
	var cost map[string]float64
	for _, k := range sortedKeysOf(state.Buildings) {
		if b := state.Buildings[k]; b.AgeKey == state.Age && b.Unlocked && b.Category != "wonder" && len(b.NextCost) >= 3 {
			key, cost = k, b.NextCost
			break
		}
	}
	if key == "" {
		t.Fatal("no building of the Medieval Age costs three resources")
	}
	res := sortedKeysOf(cost)
	// Hold the first resource in full, the second at a quarter, rising two a
	// second; the rest in full.
	state.TickIntervalMs = 1000
	set := func(r string, amount, rate float64) {
		rs := state.Resources[r]
		rs.Amount, rs.Rate = amount, rate
		state.Resources[r] = rs
	}
	set(res[0], cost[res[0]], 0)
	set(res[1], cost[res[1]]/4, 2)
	for _, r := range res[2:] {
		set(r, cost[r], 0)
	}
	canBuild := func(can bool) {
		b := state.Buildings[key]
		b.CanBuild = can
		state.Buildings[key] = b
	}
	canBuild(false)
	c := cardFor(t, state, key, buyOne)
	if c.canBuild {
		t.Fatal("staging: the card can be built")
	}
	for i, p := range c.price {
		if want := i != 1; p.ok != want {
			t.Errorf("price part %d (%s) lit = %v, want %v", i, p.text, p.ok, want)
		}
	}
	if want := (1 + 0.25 + float64(len(res)-2)) / float64(len(res)); c.frac < want-1e-9 || c.frac > want+1e-9 {
		t.Errorf("the gauge reads %v, want %v", c.frac, want)
	}
	deficit := cost[res[1]] * 3 / 4
	if want := "short of " + game.Amount(deficit, res[1]); c.reason != want {
		t.Errorf("the reason is %q, want %q", c.reason, want)
	}
	if want := "in " + durationOf(deficit/2); c.wait != want {
		t.Errorf("the wait is %q, want %q", c.wait, want)
	}
	// Not rising: no income.
	set(res[1], cost[res[1]]/4, 0)
	if c := cardFor(t, state, key, buyOne); c.wait != txtNoIncome.String() {
		t.Errorf("with no income the wait is %q, want %q", c.wait, txtNoIncome.String())
	}
	// The resource that takes longest is the one named.
	set(res[0], 0, 1)
	set(res[1], 0, 1000)
	if c := cardFor(t, state, key, buyOne); !strings.Contains(c.reason, game.ResourceName(res[0])) {
		t.Errorf("the reason %q does not name %s, which is furthest away", c.reason, game.ResourceName(res[0]))
	}
	// Enough of all of it: nothing is short.
	for _, r := range res {
		set(r, cost[r], 0)
	}
	canBuild(true)
	if c := cardFor(t, state, key, buyOne); c.reason != "" || c.wait != "" || c.group != groupBuild {
		t.Errorf("a price in full has the reason %q, the wait %q and the group %d", c.reason, c.wait, c.group)
	}
}

// cardFor is the card of key in a state.
func cardFor(t *testing.T, state game.GameState, key string, buy buyChoice) buildCard {
	t.Helper()
	for _, c := range buildCards(state, buy) {
		if c.key == key {
			return c
		}
	}
	t.Fatalf("no card for %s", key)
	return buildCard{}
}

// durationOf is a wait of secs seconds as the cards write it.
func durationOf(secs float64) string {
	return textfmt.Duration(time.Duration(secs * float64(time.Second)))
}

// The price of x10 is that of ten copies, and of max that of the copies the
// player can pay for now, at least one.
func TestCardPriceForTheAmountChosen(t *testing.T) {
	d, _, eng := cardTown(t)
	state := eng.GetState()
	const key = "gathering_camp"
	one, _ := state.BuildBatch(key, 1)
	ten, _ := state.BuildBatch(key, 10)
	n := state.MaxBuildable(key)
	most, _ := state.BuildBatch(key, n)
	for _, c := range []struct {
		buy  buyChoice
		cost map[string]float64
	}{{buyOne, one}, {buyTen, ten}, {buyMax, most}} {
		card := cardFor(t, state, key, c.buy)
		if len(card.price) != len(c.cost) {
			t.Fatalf("choice %d prices %d resources, want %d", c.buy, len(card.price), len(c.cost))
		}
		for i, r := range sortedKeysOf(c.cost) {
			if want := game.Amount(c.cost[r], r); card.price[i].text != want {
				t.Errorf("choice %d prices %q, want %q", c.buy, card.price[i].text, want)
			}
		}
	}
	// Not even one affordable: max still shows the price of one.
	for _, r := range []string{"food", "wood", "stone"} {
		eng.SetStockForTest(r, 0)
	}
	if got, want := cardFor(t, eng.GetState(), key, buyMax).price[0].text, game.Amount(one[sortedKeysOf(one)[0]], sortedKeysOf(one)[0]); got != want {
		t.Errorf("max with nothing affordable prices %q, want the price of one, %q", got, want)
	}
	_ = d
}

// Order: what can be built first, cheapest first; then what has at least half
// its price in hand, closest first; the rest wait behind one quiet row that
// counts them. Opened, the row lists them all in the same rule.
func TestCardOrderAndTheQuietRow(t *testing.T) {
	cards := []buildCard{
		{key: "far", frac: 0.1, group: groupLater},
		{key: "dear", canBuild: true, share: 0.6, group: groupBuild},
		{key: "close", frac: 0.9, group: groupNear},
		{key: "cheap", canBuild: true, share: 0.1, group: groupBuild},
		{key: "half", frac: 0.5, group: groupNear},
		{key: "nearer", frac: 0.3, group: groupLater},
	}
	keys, shown := orderCards(cards)
	want := []string{"cheap", "dear", "close", "half", "nearer", "far"}
	if strings.Join(keys, " ") != strings.Join(want, " ") || shown != 4 {
		t.Errorf("order %v with %d before the quiet row, want %v and 4", keys, shown, want)
	}

	d, st, eng := cardTown(t)
	box := d.economyTab.buildBox
	later := len(box.order) - box.shown
	if later > 0 {
		bx, by, bw, bh := box.GetRect()
		if text := strings.Join(rectText(drawnDashboard(t, d, d.pages, 160, 45), 160, bx, by, bw, bh), "\n"); box.more < 0 || !strings.Contains(text, "more later") {
			t.Errorf("%d cards wait and the box shows no quiet row:\n%s", later, text)
		}
		if want := later; box.hidden != want {
			t.Errorf("the quiet row counts %d, want %d", box.hidden, want)
		}
		// A click on it opens the rest, listed in the same order.
		x, y, _, _ := box.GetRect()
		st.click(x+10, y+box.more)
		if !box.expanded {
			t.Fatal("a click on the quiet row did not open it")
		}
		d.refresh()
		st.paint()
		var got []string
		for _, pc := range box.placed {
			got = append(got, pc.card.key)
		}
		if len(got) != len(box.order) || strings.Join(got, " ") != strings.Join(box.order, " ") {
			t.Errorf("the opened list shows %v, want %v", got, box.order)
		}
	}
	_ = eng
}

// The order does not change while the pointer is over the box, and is made
// again when the pointer leaves, or a building is built.
func TestCardsDoNotMoveUnderThePointer(t *testing.T) {
	d, st, eng := medievalTown(t, 160, 45)
	box := d.economyTab.buildBox
	x, y, _, _ := box.GetRect()
	// A stock that puts the cards in another order than the one they are in.
	reorder := func() {
		t.Helper()
		held := strings.Join(box.order, " ")
		for _, level := range []float64{0, 1e2, 1e4, 1e5, 1e6, 1e7, 1e9} {
			setStock(eng, level)
			fresh, _ := orderCards(buildCards(eng.GetState(), box.buy))
			if strings.Join(fresh, " ") != held {
				return
			}
		}
		t.Fatalf("staging: no stock puts the cards in another order than %v", held)
	}
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: x + 20, Y: y + 8})
	if !box.in {
		t.Fatal("the pointer is not over the box")
	}
	held := append([]string(nil), box.order...)
	reorder()
	d.refresh()
	if strings.Join(box.order, " ") != strings.Join(held, " ") {
		t.Errorf("the order changed under the pointer: %v then %v", held, box.order)
	}
	st.mouse(kit.Mouse{Kind: kit.MouseLeave})
	d.refresh()
	if fresh, _ := orderCards(box.cards); strings.Join(box.order, " ") != strings.Join(fresh, " ") || strings.Join(box.order, " ") == strings.Join(held, " ") {
		t.Errorf("with the pointer away the order is %v, want it made again (was %v)", box.order, held)
	}
	// A build sorts again even under the pointer.
	setStock(eng, 1e9)
	d.refresh()
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: x + 20, Y: y + 8})
	held = append(held[:0], box.order...)
	reorder()
	d.refresh()
	if strings.Join(box.order, " ") != strings.Join(held, " ") {
		t.Fatalf("the order changed under the pointer before the build: %v then %v", held, box.order)
	}
	setStock(eng, 1e9)
	if err := eng.BuildBuilding(held[0]); err != nil {
		t.Fatal(err)
	}
	reorder()
	d.refresh()
	if fresh, _ := orderCards(box.cards); strings.Join(box.order, " ") != strings.Join(fresh, " ") {
		t.Errorf("a build did not sort the cards again: %v, want %v", box.order, fresh)
	}
}

// The wheel and PgDn turn the cards a listed thing or a page at a time and
// stop at either end; turned past the last card before the quiet row they
// open it.
func TestTheWheelAndPageKeysTurnTheCards(t *testing.T) {
	d, st, eng := medievalTown(t, 160, 30)
	setStock(eng, 1e9)
	d.refresh()
	st.paint()
	box := d.economyTab.buildBox
	x, y, _, _ := box.GetRect()
	if box.hidden == 0 {
		t.Fatal("staging: every card fits a 30-row screen")
	}
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: x + 20, Y: y + 8})
	wheel := func(rows int) bool {
		moved := st.mouse(kit.Mouse{Kind: kit.MouseWheel, X: x + 20, Y: y + 8, Wheel: rows})
		st.paint()
		return moved
	}
	if wheel(-1) || box.first != 0 {
		t.Errorf("a turn up at the top moved the list to %d", box.first)
	}
	if !wheel(3) || box.first != 1 {
		t.Errorf("a turn down left the list at %d, want 1 (one a turn, however far)", box.first)
	}
	if !wheel(-1) || box.first != 0 {
		t.Errorf("a turn back up left the list at %d, want 0", box.first)
	}
	for i := 0; i < 2*len(box.order); i++ {
		wheel(1)
	}
	if box.first == 0 || box.first+box.shownItems < box.items {
		t.Errorf("turning down to the end shows %d to %d of %d", box.first, box.first+box.shownItems, box.items)
	}
	if wheel(1) && !box.expanded {
		t.Error("a turn down at the end still moved the list")
	}
	box.first, box.expanded = 0, false
	d.refresh()
	st.paint()
	if captured(d.root, tk(kit.KeyPgDn, 0, 0)) != nil {
		t.Error("PgDn was not taken by the dashboard")
	}
	if box.first == 0 {
		t.Error("PgDn did not turn the cards")
	}
	if captured(d.root, tk(kit.KeyPgUp, 0, 0)) != nil {
		t.Error("PgUp was not taken by the dashboard")
	}
	if box.first != 0 {
		t.Errorf("PgUp from the second page left the list at %d", box.first)
	}

	// With cards behind the quiet row, the wheel past the cards opens it.
	setStock(eng, 0)
	box.first, box.expanded = 0, false
	st.mouse(kit.Mouse{Kind: kit.MouseLeave})
	d.refresh()
	st.paint()
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: x + 20, Y: y + 8})
	if box.shown != 0 || len(box.order) == 0 {
		t.Fatalf("staging: %d cards before the quiet row, %d in all", box.shown, len(box.order))
	}
	if !wheel(1) || !box.expanded {
		t.Error("a turn down at the quiet row did not open it")
	}
	d.refresh()
	st.paint()
	if len(box.placed) == 0 {
		t.Error("the opened list shows no card")
	}
}

// Every building maps to one of the 18 kinds; every sign is 13 by 13 of # and
// . and no two are alike; each kind has a tint and a glyph.
func TestEveryBuildingHasASign(t *testing.T) {
	if len(signKinds) != 18 {
		t.Fatalf("%d kinds, want 18", len(signKinds))
	}
	isKind := map[string]bool{}
	for _, k := range signKinds {
		isKind[k] = true
	}
	used := map[string]int{}
	for key, def := range config.BuildingByKey() {
		k := signKindOf(def)
		if !isKind[k] {
			t.Errorf("%s has the kind %q, which has no sign", key, k)
		}
		used[k]++
	}
	for _, k := range signKinds {
		if used[k] == 0 {
			t.Errorf("no building has the kind %q", k)
		}
		pat := signPix[k]
		if len(pat) != 13 {
			t.Errorf("the %s sign has %d rows", k, len(pat))
		}
		struck := 0
		for _, row := range pat {
			if len([]rune(row)) != 13 || strings.Trim(row, "#.") != "" {
				t.Errorf("the %s sign has the row %q", k, row)
			}
			struck += strings.Count(row, "#")
		}
		if struck < 20 {
			t.Errorf("the %s sign has only %d struck pixels", k, struck)
		}
		if _, ok := signTints[k]; !ok {
			t.Errorf("no tint for %s", k)
		}
		if signGlyphs[k] == 0 || signGlyphsPlain[k] == 0 || signGlyphsPlain[k] > 0x7e {
			t.Errorf("the %s glyphs are %q and %q", k, signGlyphs[k], signGlyphsPlain[k])
		}
	}
	for i, a := range signKinds {
		for _, b := range signKinds[i+1:] {
			if strings.Join(signPix[a], "") == strings.Join(signPix[b], "") {
				t.Errorf("the %s and %s signs are the same", a, b)
			}
		}
	}
}

// The left column is 46 cells at 160 columns and the cards have the rest of
// the economy tab; a narrower screen shares them as it always did. The cards
// hold at every width: inside the box, one under another, every price whole,
// the Build buttons in one column inside the box.
func TestCardsAtEveryWidth(t *testing.T) {
	restoreForge(t)
	for _, c := range []struct {
		w, h        int
		left, cards int // the box widths; cards 0 where only the layout rules are checked
	}{{100, 30, 0, 0}, {120, 40, 43, 44}, {160, 45, 46, 81}, {240, 60, 46, 161}} {
		d, pages, _ := stagedDashboard(t, "medieval_age")
		cells := drawnDashboard(t, d, pages, c.w, c.h)
		_, _, lw, _ := d.economyTab.leftCol.GetRect()
		_, _, bw, _ := d.economyTab.buildBox.GetRect()
		if c.cards != 0 && (lw != c.left || bw != c.cards) {
			t.Errorf("%dx%d: the left column is %d wide and the cards' box %d, want %d and %d", c.w, c.h, lw, bw, c.left, c.cards)
		}
		checkCards(t, fmt.Sprintf("the Medieval Age at %dx%d", c.w, c.h), d, cells, c.w)
	}
}

// With pictures the cards are drawn: the slab, the gauge and each Build button
// are pictures over their cells, and what a card says is still text.
func TestCardsDrawPictures(t *testing.T) {
	was := pictureCell
	pictureCell = func() (int, int) { return 10, 21 }
	t.Cleanup(func() { pictureCell = was })
	d, st, _ := cardTown(t)
	st.paint()
	pics := len(st.buf.Pics)
	// One picture a card, and one more for each button, the buy keys and the gather buttons.
	if want := 2 * len(d.economyTab.buildBox.placed); pics < want {
		t.Errorf("%d pictures for %d cards, want at least %d", pics, len(d.economyTab.buildBox.placed), want)
	}
}
