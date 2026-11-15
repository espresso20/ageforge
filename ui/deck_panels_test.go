package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// deckText is every word a page shows, one per line, for a test to look in.
func deckText(pg deckPage) string {
	var sb strings.Builder
	line := func(parts ...string) { sb.WriteString(strings.Join(parts, " ") + "\n") }
	line(pg.Title, pg.Hint)
	for _, s := range pg.Head {
		sb.WriteString(s.s)
	}
	sb.WriteString("\n")
	for _, n := range pg.Notes {
		line(n.Label, n.Value)
	}
	for _, s := range pg.Sections {
		line(s.Title, s.Meta)
		for _, n := range s.Notes {
			line(n.Label, n.Value)
		}
		for _, c := range s.Cards {
			line(c.Title, c.Chip, c.Desc, c.MeterText, c.Why, c.Action, c.Command)
			for _, l := range c.Lines {
				line(l.Label, l.Value)
			}
		}
	}
	return sb.String()
}

// deckCardNamed is the card called title.
func deckCardNamed(t *testing.T, pg deckPage, title string) (deckCard, int) {
	t.Helper()
	for n := 0; n < pg.count(); n++ {
		if c, _ := pg.card(n); c.Title == title {
			return c, n
		}
	}
	t.Fatalf("no card %q in:\n%s", title, deckText(pg))
	return deckCard{}, 0
}

// deckSectionNamed is the section called title.
func deckSectionNamed(t *testing.T, pg deckPage, title string) deckSection {
	t.Helper()
	for _, s := range pg.Sections {
		if s.Title == title {
			return s
		}
	}
	t.Fatalf("no section %q in:\n%s", title, deckText(pg))
	return deckSection{}
}

// deckOf is a deck panel for page, open on state and laid out w by h, and
// the commands it ran.
func deckOf(t *testing.T, page func(game.GameState) deckPage, state game.GameState, w, h int) (*deckPanel, *[]string) {
	t.Helper()
	keepTheme(t)
	ran := &[]string{}
	p := newDeckPanel(page)
	p.run = func(cmd string) { *ran = append(*ran, cmd) }
	p.buttons.press = func(kit.Key) {}
	p.open(state)
	p.Layout(image.Rect(0, 0, w, h))
	return p, ran
}

// wantDeckEnter: with card n picked, Enter runs exactly command.
func wantDeckEnter(t *testing.T, page func(game.GameState) deckPage, state game.GameState, title, command string) {
	t.Helper()
	pg := page(state)
	_, n := deckCardNamed(t, pg, title)
	p, ran := deckOf(t, page, state, 153, 41)
	p.pick(n)
	p.Layout(image.Rect(0, 0, 153, 41))
	if !p.routeKey(&kit.Key{Code: kit.KeyEnter}, "") || len(*ran) != 1 || (*ran)[0] != command {
		t.Errorf("Enter on %q ran %q, want %q", title, *ran, command)
	}
}

// wantDeckFits: the panel lays out and draws at the three sizes the game is
// played at, every row the width of the frame and every card in the content.
func wantDeckFits(t *testing.T, page func(game.GameState) deckPage, state game.GameState) {
	t.Helper()
	pg := page(state)
	for _, sz := range [][2]int{{80, 21}, {100, 30}, {153, 41}} {
		w, h := sz[0], sz[1]
		p, _ := deckOf(t, page, state, w, h)
		for i, line := range strings.Split(strings.TrimSuffix(p.grid.String(), "\n"), "\n") {
			if n := len([]rune(line)); n != w {
				t.Fatalf("%dx%d: row %d is %d cells", w, h, i, n)
			}
		}
		if len(p.lay.cards) != pg.count() {
			t.Fatalf("%dx%d: %d card places for %d cards", w, h, len(p.lay.cards), pg.count())
		}
		for n, r := range p.lay.cards {
			if r.Max.X > w-4 || r.Max.Y > p.lay.total {
				t.Fatalf("%dx%d: card %d at %v is outside the content (%d rows)", w, h, n, r, p.lay.total)
			}
		}
	}
}

// TestDeckPanelsOnTheDashboard: each deck opens from its command, shows a
// foot that names Esc and offers the command bar, takes Esc to close, and has
// its overlay in the bar's list so a command's answer reaches it.
func TestDeckPanelsOnTheDashboard(t *testing.T) {
	d, pages := poisonedStage(t)
	if len(d.decks) == 0 {
		t.Fatal("no deck panels registered")
	}
	for name := range d.decks {
		openPanel(t, d, name, 100, 30)
		if d.overlayMgr.ActiveName() != name {
			t.Errorf("%s opened %q", name, d.overlayMgr.ActiveName())
		}
		cells := drawnDashboard(t, d, pages, 100, 30)
		foot := rectText(cells, 100, 0, 30-promptRows-2, 100, 1)[0]
		for _, want := range []string{"Esc", "or type"} {
			if !strings.Contains(foot, want) {
				t.Errorf("%s: the foot does not say %q: %q", name, want, foot)
			}
		}
		if o, ok := d.barOverlays[name]; !ok || o.reply == nil || o.routeKey == nil || o.update == nil {
			t.Errorf("%s: no bar overlay with a reply: %+v", name, o)
		}
		pressKey(d, pages, kit.KeyEsc, 0)
		if d.overlayMgr.HasActive() {
			t.Errorf("%s: Esc did not close the panel", name)
		}
	}
}

// TestDeckShowsWhatACommandDid: after a command runs over the deck its line
// and the first line of the answer are on the panel.
func TestDeckShowsWhatACommandDid(t *testing.T) {
	p, _ := deckOf(t, tradePage, tradeState(), 100, 30)
	p.reply("trade route stop local_barter", CommandResult{Type: "success", Message: "Stopped Local Barter.\nmore"})
	p.Layout(image.Rect(0, 0, 100, 30))
	wantAll(t, "deck", p.grid.String(), "› trade route stop local_barter", "Stopped Local Barter.")
}
