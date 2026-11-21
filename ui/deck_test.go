package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// sampleDeck is a page with two sections: seven cards with a command, a
// locked one, and one that only shows.
func sampleDeck() deckPage {
	open := deckSection{Title: "Open routes", Meta: "7 open"}
	for _, name := range []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf"} {
		open.Cards = append(open.Cards, deckCard{
			Title: "Route " + name, Chip: "ready", ChipK: xChipCivic, Desc: "Send wood to the quarries for stone.",
			Lines:  []deckLine{{"Export", "15 wood", 0}, {"Import", "12 stone", xGood}},
			Action: "start", Command: "trade route start " + name,
		})
	}
	open.Cards = append(open.Cards, deckCard{Title: "Silk Road", Chip: "locked", ChipK: xKeyName, Why: "needs 2 Markets (have 0)", Quiet: true},
		deckCard{Title: "Sea Lane", Chip: "locked", ChipK: xKeyName, Why: "Trade routes need The Wheel first. Research it to start one, then come back.", Quiet: true})
	return deckPage{
		Title: "Trade",
		Head:  []xseg{{"Routes ", xDim}, {"7 open", xText}},
		Notes: []deckLine{{Value: "Routes importing iron are suspended.", K: xBad}},
		Sections: []deckSection{
			{Title: "Overview", Cards: []deckCard{{Title: "Caravans", Meter: true, Share: 0.5, MeterText: "1 of 2 out"}}},
			open,
		},
		Hint: "trade route start <key>",
	}
}

// deckOnStage is a deck panel laid out w by h, and the commands it ran.
func deckOnStage(t *testing.T, w, h int) (*deckPanel, *[]string) {
	t.Helper()
	keepTheme(t)
	ran := &[]string{}
	p := newDeckPanel(func(game.GameState) deckPage { return sampleDeck() })
	p.buttons.run = func(cmd string) { *ran = append(*ran, cmd) }
	p.buttons.press = func(kit.Key) {}
	p.open(game.GameState{})
	p.Layout(image.Rect(0, 0, w, h))
	return p, ran
}

// TestDeckFitsEverySize: at any size the deck draws inside its frame, every
// card is inside the content, and nothing a click can hit is outside it.
func TestDeckFitsEverySize(t *testing.T) {
	pg := sampleDeck()
	for w := 20; w <= 200; w += 7 {
		for h := 5; h <= 50; h += 3 {
			g, lay := renderDeck(&pg, 3, 4, w, h)
			for i, line := range strings.Split(strings.TrimSuffix(g.String(), "\n"), "\n") {
				if n := len([]rune(line)); n != w {
					t.Fatalf("%dx%d: row %d is %d cells", w, h, i, n)
				}
			}
			for n, r := range lay.cards {
				if r.Min.X < 0 || r.Max.X > w-4 || r.Min.Y < 0 || r.Max.Y > lay.total {
					t.Errorf("%dx%d: card %d at %v is outside the %d by %d content", w, h, n, r, w-4, lay.total)
				}
			}
			for _, ht := range g.hits {
				if !ht.r.In(image.Rect(2, 1, w-2, h-3)) {
					t.Errorf("%dx%d: a hit at %v is outside the view", w, h, ht.r)
				}
			}
		}
	}
}

// TestDeckShowsWhatACardDoes: a card shows its state, what stands in its
// way and the command its button runs; the picked card's button shows Enter.
func TestDeckShowsWhatACardDoes(t *testing.T) {
	pg := sampleDeck()
	g, lay := renderDeck(&pg, 1, 0, 153, 40)
	text := g.String()
	for _, want := range []string{"Trade", "Routes 7 open", "Routes importing iron are suspended.", "Open routes", "7 open", "Route alpha", " ready ", "Export 15 wood", " Enter ", " start ", "needs 2 Markets (have 0)", "Trade routes need The Wheel first.", "then come back.", " locked ", "1 of 2 out", "or type trade route start <key>", " Esc  close "} {
		if !strings.Contains(text, want) {
			t.Errorf("the deck does not show %q:\n%s", want, text)
		}
	}
	// The buttons carry the commands the cards used to print.
	cmds := map[string]xkey{}
	for _, k := range g.keys {
		cmds[k.cmd] = k
	}
	if k := cmds["trade route start alpha"]; k.kind != keyPrimary || k.r.Dy() != actionRows || k.key.Code != kit.KeyEnter {
		t.Errorf("the picked card's button is %+v, want the primary action button on Enter", k)
	}
	if k := cmds["trade route start bravo"]; k.kind != keyPlain || k.r.Dy() != 1 {
		t.Errorf("another card's button is %+v, want its chip", k)
	}
	if strings.Contains(text, "trade route start alpha") || strings.Contains(text, "trade route start bravo") {
		t.Errorf("a card still prints the command its button carries:\n%s", text)
	}
	if len(lay.cards) != 10 || lay.total > lay.viewH {
		t.Errorf("%d cards in %d rows of %d in view; want 10 that all fit", len(lay.cards), lay.total, lay.viewH)
	}
	if strings.Contains(text, "PgUp") {
		t.Errorf("a deck that fits offers to page:\n%s", text)
	}
}

// TestDeckArrowsMoveByPlace: Right and Left go by number; Down and Up go to
// the card under or over the pick, and stop at the ends.
func TestDeckArrowsMoveByPlace(t *testing.T) {
	pg := sampleDeck()
	_, lay := renderDeck(&pg, 0, 0, 153, 60) // three columns of 49
	if got := lay.step(1, 1, 0); got != 2 {
		t.Errorf("Right from card 1 = %d, want 2", got)
	}
	if got := lay.step(0, -1, 0); got != 0 {
		t.Errorf("Left from the first card = %d, want it to stay", got)
	}
	for _, from := range []int{1, 2, 3} {
		to := lay.step(from, 0, 1)
		if a, b := lay.cards[from], lay.cards[to]; b.Min.Y <= a.Min.Y || b.Min.X != a.Min.X {
			t.Errorf("Down from card %d (%v) = card %d (%v), want the card right under it", from, a, to, b)
		}
		if back := lay.step(to, 0, -1); back != from {
			t.Errorf("Up from card %d = %d, want back at %d", to, back, from)
		}
	}
	if got := lay.step(0, 0, -1); got != 0 {
		t.Errorf("Up from the first card = %d, want it to stay", got)
	}
	last := len(lay.cards) - 1
	if got := lay.step(last, 0, 1); got != last {
		t.Errorf("Down from the last card = %d, want it to stay", got)
	}
}

// TestDeckPanelKeysAndClicks: Enter runs the picked card's command, a click
// picks, a click on a card's button or a double click on the card runs its
// command, and a card that only shows leaves Enter to the command bar.
func TestDeckPanelKeysAndClicks(t *testing.T) {
	p, ran := deckOnStage(t, 153, 40)
	key := func(c kit.KeyCode, prompt string) bool { return p.routeKey(&kit.Key{Code: c}, prompt) }
	relay := func() { p.Layout(p.Bounds()) }
	if p.sel != 1 {
		t.Fatalf("the panel opens on card %d, want the first with a command (1)", p.sel)
	}
	if key(kit.KeyEnter, "trade") || key(kit.KeyRight, "x") {
		t.Error("with something typed, a key was taken from the command bar")
	}
	if !key(kit.KeyEnter, "") || len(*ran) != 1 || (*ran)[0] != "trade route start alpha" {
		t.Errorf("Enter ran %v, want the picked card's command", *ran)
	}
	key(kit.KeyRight, "")
	relay()
	if !key(kit.KeyEnter, "") || (*ran)[len(*ran)-1] != "trade route start bravo" {
		t.Errorf("after Right, Enter ran %v", *ran)
	}
	click := func(id, n, clicks int) bool {
		for _, ht := range p.grid.hits {
			if ht.id == id && ht.n == n {
				return p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: ht.r.Min.X, Y: ht.r.Min.Y, Clicks: clicks})
			}
		}
		t.Fatalf("no hit %d for card %d", id, n)
		return false
	}
	*ran = nil
	if !click(hitDeckCard, 4, 1) || p.sel != 4 || len(*ran) != 0 {
		t.Errorf("a click on card 4: picked %d, ran %v; want it picked and nothing run", p.sel, *ran)
	}
	relay()
	press := func(cmd string) {
		for _, k := range p.grid.keys {
			if k.cmd == cmd {
				p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: k.r.Min.X + 1, Y: k.r.Min.Y})
				p.Mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: k.r.Min.X + 1, Y: k.r.Min.Y})
				return
			}
		}
		t.Fatalf("no button carries %q", cmd)
	}
	press("trade route start echo")
	relay()
	click(hitDeckCard, 6, 2)
	if p.sel != 6 || len(*ran) != 2 || (*ran)[0] != "trade route start echo" || (*ran)[1] != "trade route start foxtrot" {
		t.Errorf("a click on a button and a double click on a card ran %v with card %d picked", *ran, p.sel)
	}
	// What a command did shows over the keys.
	p.reply("trade route start echo", CommandResult{Message: "[green]Route started.[-]\nmore", Type: "success"})
	relay()
	if text := p.grid.String(); !strings.Contains(text, "› trade route start echo  Route started.") || strings.Contains(text, "more\n") {
		t.Errorf("the panel does not say what the command did:\n%s", text)
	}
	p.reply("trade route stop echo", CommandResult{})
	relay()
	if text := p.grid.String(); !strings.Contains(text, "› trade route stop echo  ✓") {
		t.Errorf("a command with no answer of its own is not ticked:\n%s", text)
	}
	// A card with no command: picked by a click, and Enter is not the panel's.
	relay()
	click(hitDeckCard, 0, 1)
	if key(kit.KeyEnter, "") {
		t.Error("Enter was taken on a card that has no command")
	}
	if p.PointerAt(p.lay.cards[2].Min.Add(image.Pt(3, 2))) != kit.PointerHand || p.PointerAt(image.Pt(1, 1)) != kit.PointerArrow {
		t.Error("the pointer is not a hand over a card and an arrow off it")
	}
}

// TestDeckScrolls: a deck taller than the panel pages, shows where it is,
// follows the pick, and takes the wheel.
func TestDeckScrolls(t *testing.T) {
	p, _ := deckOnStage(t, 100, 20)
	if p.lay.total <= p.lay.viewH {
		t.Fatalf("the sample deck fits a 100x20 panel (%d rows in %d): the test needs it not to", p.lay.total, p.lay.viewH)
	}
	text := p.grid.String()
	if !strings.Contains(text, "┃") || !strings.Contains(text, "PgUp PgDn") {
		t.Errorf("a deck that does not fit has no scroll mark or no page keys:\n%s", text)
	}
	p.routeKey(&kit.Key{Code: kit.KeyEnd}, "")
	p.Layout(p.Bounds())
	last := p.lay.cards[len(p.lay.cards)-1]
	if p.sel != len(p.lay.cards)-1 || last.Min.Y < p.scroll || last.Max.Y > p.scroll+p.lay.viewH {
		t.Errorf("End picked card %d with rows %d to %d in view at scroll %d (%d rows): the pick is out of sight", p.sel, last.Min.Y, last.Max.Y, p.scroll, p.lay.viewH)
	}
	if !strings.Contains(p.grid.String(), "Silk Road") {
		t.Errorf("the last card is not on the screen after End:\n%s", p.grid.String())
	}
	p.routeKey(&kit.Key{Code: kit.KeyHome}, "")
	if p.scroll != 0 {
		t.Errorf("Home left the deck scrolled to %d", p.scroll)
	}
	if !p.Mouse(kit.Mouse{Kind: kit.MouseWheel, X: 10, Y: 5, Wheel: 1}) || p.scroll != 3 {
		t.Errorf("the wheel moved the deck to %d, want 3", p.scroll)
	}
	p.routeKey(&kit.Key{Code: kit.KeyPgDn}, "")
	p.routeKey(&kit.Key{Code: kit.KeyPgDn}, "")
	p.routeKey(&kit.Key{Code: kit.KeyPgDn}, "")
	if want := p.lay.total - p.lay.viewH; p.scroll != want {
		t.Errorf("PgDn past the end left the deck at %d, want %d", p.scroll, want)
	}
}
