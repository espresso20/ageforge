package ui

import (
	"bytes"
	"image"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// actionGrid is a grid with one action button of each kind in a row.
func actionGrid(a ...chipSpec) *xgrid {
	g := newXGrid(80, chipRows+1)
	x := 1
	for _, s := range a {
		x = g.chip(x, 1, s) + 1
	}
	return g
}

func click(k *keyPresser, g keyHolder, x, y int) {
	k.mouse(g, image.Point{}, kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: x, Y: y})
	k.mouse(g, image.Point{}, kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: x, Y: y})
}

func TestAClickOnACommandButtonRunsItsLineOnceAndAnOffOneRunsNothing(t *testing.T) {
	g := actionGrid(
		chipSpec{kind: keyAccent, label: "send", key: "Enter", cmd: "expedition mist"},
		chipSpec{kind: keySteel, label: "stop", cmd: "trade route stop x"},
		chipSpec{kind: keyAccent, label: "send", key: "Enter", cmd: "never", off: true},
	)
	var ran []string
	var pressed []kit.Key
	k := keyPresser{press: func(k kit.Key) { pressed = append(pressed, k) }, run: func(c string) { ran = append(ran, c) }}
	for _, b := range g.keys {
		click(&k, g, b.r.Min.X+1, b.r.Min.Y)
	}
	if len(ran) != 2 || ran[0] != "expedition mist" || ran[1] != "trade route stop x" {
		t.Errorf("clicks ran %q", ran)
	}
	if len(pressed) != 0 {
		t.Errorf("a button that carries a line also pressed %v", pressed)
	}
}

func TestAKeyButtonWithoutALinePressesItsKey(t *testing.T) {
	g := actionGrid(chipSpec{kind: keyAccent, label: "start", key: "Enter"})
	var ran []string
	var pressed []kit.Key
	k := keyPresser{press: func(k kit.Key) { pressed = append(pressed, k) }, run: func(c string) { ran = append(ran, c) }}
	click(&k, g, 3, 1)
	if len(ran) != 0 || len(pressed) != 1 || pressed[0].Code != kit.KeyEnter {
		t.Errorf("a click ran %q and pressed %v, want Enter alone", ran, pressed)
	}
}

func TestAClickSinksTheButtonUntilItIsUpEvenWithoutAKey(t *testing.T) {
	g := actionGrid(chipSpec{kind: keySteel, label: "stop", cmd: "stop it"})
	now := time.Unix(100, 0)
	k := keyPresser{press: func(kit.Key) {}, run: func(string) {}, feel: &keyFeel{now: func() time.Time { return now }}}
	b := g.keys[0]
	k.mouse(g, image.Point{}, kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: b.r.Min.X + 1, Y: b.r.Min.Y})
	now = now.Add(time.Second)
	if !k.sunkButton(b) {
		t.Error("a held button came up")
	}
	k.mouse(g, image.Point{}, kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: b.r.Min.X + 1, Y: b.r.Min.Y})
	if k.sunkButton(b) {
		t.Error("a released button stayed down")
	}
}

func TestAChipsWidthIsItsTextAndItIsOneRow(t *testing.T) {
	for _, s := range []chipSpec{
		{kind: keyAccent, label: "send", key: "Enter"},
		{kind: keySteel, label: "cancel", key: "Esc"},
		{kind: keyAlert, label: "stop"},
	} {
		g := actionGrid(s)
		if g.keys[0].r.Dx() != chipW(s.label, s.key) || g.keys[0].r.Dy() != chipRows {
			t.Errorf("%q takes %v, want %d by %d", s.label, g.keys[0].r, chipW(s.label, s.key), chipRows)
		}
	}
}

func TestChipPlatesAreCachedAndKindsAndStatesDiffer(t *testing.T) {
	keepTheme(t)
	pal := newExpPalette()
	var set plateSet
	plate := func(kind keyKind, st keyState) *image.RGBA {
		return set.plate(pal.look(kind, st), st, 8, 10, 21)
	}
	if plate(keyAccent, keyRest) != plate(keyAccent, keyRest) {
		t.Error("the same button twice is two pictures")
	}
	seen := map[string]*image.RGBA{}
	for _, kind := range []keyKind{keyAccent, keySteel, keyAlert} {
		for _, st := range []keyState{keyRest, keyHover, keyPressed, keyOff} {
			if st == keyOff && kind != keyAccent {
				continue
			}
			img := plate(kind, st)
			for name, o := range seen {
				// Under the pointer every kind takes the accent fill.
				if st != keyRest && name[1] == byte('0'+st) && name[0] != byte('a'+kind) {
					continue
				}
				if bytes.Equal(img.Pix, o.Pix) {
					t.Errorf("kind %d state %d looks like %s", kind, st, name)
				}
			}
			seen[string(rune('a'+kind))+string(rune('0'+st))] = img
		}
	}
}

// echoBar is a command bar on a clock the test sets.
func echoBar(now *time.Time) *commandInput {
	c := newCommandInput(newCompleter(nil, nil))
	c.now = func() time.Time { return *now }
	return c
}

func TestTheCommandBarEchoesALineAndLetsGo(t *testing.T) {
	now := time.Unix(100, 0)
	c := echoBar(&now)
	if c.shadow() != "" {
		t.Fatalf("an idle bar shows %q", c.shadow())
	}
	c.echoLine("trade route start salt")
	if c.shadow() != "trade route start salt" || c.GetText() != "" {
		t.Errorf("echo %q, text %q", c.shadow(), c.GetText())
	}
	now = now.Add(echoFor - time.Millisecond)
	if c.shadow() == "" {
		t.Error("the echo went early")
	}
	now = now.Add(2 * time.Millisecond)
	if c.shadow() != "" || c.echoing() {
		t.Errorf("the echo stayed: %q", c.shadow())
	}
}

func TestThePreviewShowsOnlyOnAnEmptyPromptAndLeavesTheTextAlone(t *testing.T) {
	now := time.Unix(100, 0)
	c := echoBar(&now)
	line := ""
	c.preview = func() string { return line }
	line = "army recruit"
	if c.in.Ghost("") != "army recruit" {
		t.Errorf("the empty prompt shows %q", c.in.Ghost(""))
	}
	if c.GetText() != "" {
		t.Errorf("the preview was typed: %q", c.GetText())
	}
	if g := c.in.Ghost("zzz"); g == "army recruit" {
		t.Error("the preview shows over typed text")
	}
}

func TestAPreviewFollowsThePointerAndATypedKeyTakesItAway(t *testing.T) {
	g := actionGrid(chipSpec{kind: keyAccent, label: "send", key: "Enter", cmd: "expedition mist"},
		chipSpec{kind: keySteel, label: "back", key: "Esc"})
	k := keyPresser{press: func(kit.Key) {}, run: func(string) {}}
	move := func(x, y int, kind kit.MouseKind) {
		k.mouse(g, image.Point{}, kit.Mouse{Kind: kind, X: x, Y: y})
	}
	move(3, 1, kit.MouseMove)
	if k.feel.preview != "expedition mist" {
		t.Fatalf("resting on the button previews %q", k.feel.preview)
	}
	k.feel.typed(kit.Key{Code: kit.KeyRune, Rune: 'a'})
	move(4, 1, kit.MouseMove)
	if k.feel.preview != "" {
		t.Errorf("a typed key left %q", k.feel.preview)
	}
	move(40, 0, kit.MouseMove)
	move(3, 1, kit.MouseMove)
	if k.feel.preview != "expedition mist" {
		t.Errorf("coming back previews %q", k.feel.preview)
	}
	move(g.keys[1].r.Min.X+1, 1, kit.MouseMove)
	if k.feel.preview != "" {
		t.Errorf("a button with no line previews %q", k.feel.preview)
	}
	move(3, 1, kit.MouseMove)
	move(0, 0, kit.MouseLeave)
	if k.feel.preview != "" {
		t.Errorf("the pointer left and %q stayed", k.feel.preview)
	}
}

func TestAButtonsLineEntersTheHistoryAndTheBarLikeATypedOne(t *testing.T) {
	d, _ := newTestDashboard(nil)
	var k keyPresser
	k.wire(d.app)
	g := actionGrid(chipSpec{kind: keyAccent, label: "go", key: "Enter", cmd: "help"})
	click(&k, g, 3, 1)
	if n := len(d.cmdHistory); n != 1 || d.cmdHistory[0] != "help" {
		t.Errorf("history %q", d.cmdHistory)
	}
	if d.inputField.shadow() != "help" || d.inputField.GetText() != "" {
		t.Errorf("the bar shows %q over text %q", d.inputField.shadow(), d.inputField.GetText())
	}
}

// TestAClickOnAnOpenTechCardsStartStartsTheResearch: the card of a tech that
// can start has a start button, and a click on it starts the research as the
// second Enter does; the card of one that cannot has the button off.
func TestAClickOnAnOpenTechCardsStartStartsTheResearch(t *testing.T) {
	keepTheme(t)
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	for _, a := range []string{"stone_age", "bronze_age", "iron_age", "classical_age"} {
		must(t, ge.EnterAgeForTest(a))
	}
	ge.GrantTechsForTest("language", "tool_making", "fire_mastery", "stoneworking", "primitive_writing", "pottery", "woodworking", "bronze_working", "currency", "masonry", "the_wheel", "mathematics", "iron_smelting")
	ge.SetTownForTest(map[string]int{"story_circle": 5, "elders_hall": 5, "scriptorium": 5, "agora": 7}, 0)
	ge.SetStockForTest("knowledge", 1e6)
	p := newTreePanel()
	p.engine = ge
	p.w, p.h = 120, 37
	p.wireButtons(newAppLoop())
	var start, wait *treeNode
	st := ge.GetState()
	m := buildTree(st, treeGeomFor(120, false), "")
	for _, n := range m.nodes {
		switch {
		case start == nil && cardActionFor(st, n) == actStart:
			start = n
		case wait == nil && cardActionFor(st, n) == actPlan:
			wait = n
		}
	}
	if start == nil || wait == nil {
		t.Fatalf("staging: a tech to start (%v) and one that waits (%v)", start, wait)
	}
	p.request("close", wait.key)
	p.update(st)
	renderLines(p)
	if len(p.model.keys.keys) != 1 || !p.model.keys.keys[0].off {
		t.Fatalf("the card of a tech that waits has %+v, want one button, off", p.model.keys.keys)
	}
	p.request("close", start.key)
	p.update(st)
	renderLines(p)
	if len(p.model.keys.keys) != 1 || p.model.keys.keys[0].off || p.model.keys.keys[0].kind != keyAccent {
		t.Fatalf("the card of a tech that can start has %+v, want one primary button", p.model.keys.keys)
	}
	b := p.model.keys.keys[0]
	at := b.r.Min.Add(image.Pt(2, 0))
	if p.PointerAt(at) != kit.PointerHand {
		t.Error("no hand over the start button")
	}
	if !p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: at.X, Y: at.Y}) {
		t.Fatal("the click was not taken")
	}
	if got := ge.GetState().Research.CurrentTech; got != start.key {
		t.Errorf("researching %q after the click, want %q", got, start.key)
	}
}
