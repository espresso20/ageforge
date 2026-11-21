package ui

import (
	"bytes"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// actionGrid is a grid with one action button of each kind in a row.
func actionGrid(a ...actionSpec) *xgrid {
	g := newXGrid(80, actionRows+1)
	x := 1
	for _, s := range a {
		x = g.actionButton(x, 1, s) + 1
	}
	return g
}

func click(k *keyPresser, g keyHolder, x, y int) {
	k.mouse(g, image.Point{}, kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: x, Y: y})
	k.mouse(g, image.Point{}, kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: x, Y: y})
}

func TestAClickOnACommandButtonRunsItsLineOnceAndAnOffOneRunsNothing(t *testing.T) {
	g := actionGrid(
		actionSpec{kind: keyPrimary, label: "send", key: "Enter", cmd: "expedition mist"},
		actionSpec{kind: keyPlain, label: "stop", cmd: "trade route stop x", sign: signStop},
		actionSpec{kind: keyPrimary, label: "send", key: "Enter", cmd: "never", off: true},
	)
	var ran []string
	var pressed []kit.Key
	k := keyPresser{press: func(k kit.Key) { pressed = append(pressed, k) }, run: func(c string) { ran = append(ran, c) }}
	for _, b := range g.keys {
		click(&k, g, b.r.Min.X+1, b.r.Min.Y+1)
	}
	if len(ran) != 2 || ran[0] != "expedition mist" || ran[1] != "trade route stop x" {
		t.Errorf("clicks ran %q", ran)
	}
	if len(pressed) != 0 {
		t.Errorf("a button that carries a line also pressed %v", pressed)
	}
}

func TestAKeyButtonWithoutALinePressesItsKey(t *testing.T) {
	g := actionGrid(actionSpec{kind: keyPrimary, label: "start", key: "Enter", sign: signPlay})
	var ran []string
	var pressed []kit.Key
	k := keyPresser{press: func(k kit.Key) { pressed = append(pressed, k) }, run: func(c string) { ran = append(ran, c) }}
	click(&k, g, 3, 2)
	if len(ran) != 0 || len(pressed) != 1 || pressed[0].Code != kit.KeyEnter {
		t.Errorf("a click ran %q and pressed %v, want Enter alone", ran, pressed)
	}
}

func TestAClickSinksTheButtonUntilItIsUpEvenWithoutAKey(t *testing.T) {
	g := actionGrid(actionSpec{kind: keyPlain, label: "stop", cmd: "stop it"})
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

func TestTheActionButtonsWidthIsWhatIsDrawn(t *testing.T) {
	keepTheme(t)
	withPictures(t)
	pal := newExpPalette()
	for _, s := range []actionSpec{
		{kind: keyPrimary, label: "send", key: "Enter"},
		{kind: keyPlain, label: "cancel", key: "Esc", sign: signCross},
		{kind: keyDanger, label: "stop", sign: signStop},
	} {
		g := actionGrid(s)
		w := actionW(s.label, s.key, s.sign)
		if g.keys[0].r.Dx() != w || g.keys[0].r.Dy() != actionRows {
			t.Errorf("%q takes %v, want %d by %d", s.label, g.keys[0].r, w, actionRows)
		}
		b := kit.NewBuffer(80, 4)
		p := kit.NewPainter(b)
		paintExpeditionKit(p, g, image.Point{}, pal)
		var k keyPresser
		k.paint(p, g, image.Point{}, pal)
		if len(b.Pics) != 1 || b.Pics[0].Rect != g.keys[0].r || b.Pics[0].Img.Rect.Dx() != w*10 || b.Pics[0].Img.Rect.Dy() != actionRows*21 {
			t.Errorf("%q: pictures %+v for a button of %v", s.label, b.Pics, g.keys[0].r)
		}
		for x, want := range []rune(s.label) {
			if c := b.At(g.keys[0].word.Min.X+x, 2); c.R != want || c.Pic == 0 {
				t.Errorf("%q: label cell %d is %+v", s.label, x, c)
			}
		}
	}
}

func TestActionPlatesAreCachedAndKindsAndStatesDiffer(t *testing.T) {
	keepTheme(t)
	pal := newExpPalette()
	var set plateSet
	plate := func(kind keyKind, st keyState, pose int) *image.RGBA {
		return set.action(pal.look(kind, st), st, 16, 10, 21, signTick, 11, 7, pose)
	}
	if plate(keyPrimary, keyRest, 0) != plate(keyPrimary, keyRest, 0) {
		t.Error("the same button twice is two pictures")
	}
	// An unavailable button lies flat in the ground's colours, whatever its kind.
	seen := map[string]*image.RGBA{}
	for _, kind := range []keyKind{keyPrimary, keyPlain, keyDanger} {
		for _, st := range []keyState{keyRest, keyHover, keyPressed, keyOff} {
			if st == keyOff && kind != keyPrimary {
				continue
			}
			img := plate(kind, st, 0)
			for name, o := range seen {
				if bytes.Equal(img.Pix, o.Pix) {
					t.Errorf("kind %d state %d looks like %s", kind, st, name)
				}
			}
			seen[string(rune('a'+kind))+string(rune('0'+st))] = img
		}
	}
	if bytes.Equal(plate(keyPrimary, keyRest, 2).Pix, plate(keyPrimary, keyRest, 0).Pix) {
		t.Error("the glint does not show on the plate")
	}
}

func TestWithoutPicturesAnActionButtonIsThreeRowsOfCells(t *testing.T) {
	keepTheme(t)
	pal := newExpPalette()
	g := actionGrid(actionSpec{kind: keyPrimary, label: "send", key: "Enter", sign: signTick})
	b := kit.NewBuffer(80, 4)
	p := kit.NewPainter(b)
	paintExpeditionKit(p, g, image.Point{}, pal)
	var k keyPresser
	k.paint(p, g, image.Point{}, pal)
	if len(b.Pics) != 0 {
		t.Errorf("%d pictures without pictures", len(b.Pics))
	}
	r := g.keys[0].r
	var row strings.Builder
	for x := r.Min.X; x < r.Max.X; x++ {
		row.WriteRune(b.At(x, r.Min.Y+1).R)
	}
	if !strings.Contains(row.String(), "send") || !strings.Contains(row.String(), " Enter ") {
		t.Errorf("the middle row reads %q", row.String())
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		first, last := b.At(r.Min.X, y), b.At(r.Max.X-1, y)
		if first.Bg != last.Bg || first.Bg == b.At(r.Max.X+3, y).Bg {
			t.Errorf("row %d is not a block of the kind's colour", y)
		}
	}
	// Pressed is the block reversed.
	now := time.Unix(100, 0)
	k = keyPresser{press: func(kit.Key) {}, feel: &keyFeel{now: func() time.Time { return now }}}
	rest := b.At(r.Min.X, r.Min.Y).Bg
	k.feel.typed(kit.Key{Code: kit.KeyEnter})
	k.paint(p, g, image.Point{}, pal)
	if b.At(r.Min.X, r.Min.Y).Bg == rest {
		t.Error("a typed Enter did not reverse the button")
	}
}

func TestTheGlintShowsOnlyWithMotionAndRestsBetweenCrossings(t *testing.T) {
	now := time.Unix(1000, 0)
	f := &keyFeel{now: func() time.Time { return now }}
	if f.glint() != 0 {
		t.Error("a glint with no motion setting")
	}
	on := true
	f.motion = func() bool { return on }
	poses := map[int]int{}
	for i := 0; i < 2*(actGlintPoses+actGlintRest); i++ {
		poses[f.glint()]++
		now = now.Add(mapAnimStep)
	}
	for p := 1; p <= actGlintPoses; p++ {
		if poses[p] != 2 {
			t.Errorf("pose %d came %d times in two cycles", p, poses[p])
		}
	}
	if poses[0] != 2*actGlintRest {
		t.Errorf("the glint rested %d steps of %d", poses[0], 2*actGlintRest)
	}
	on = false
	if f.glint() != 0 {
		t.Error("the glint shows with motion off")
	}
}

func TestTheAnimationTickerOnlyAsksForTheGlintWhenOneIsDue(t *testing.T) {
	now := time.Unix(1000, 0)
	f := &keyFeel{now: func() time.Time { return now }, motion: func() bool { return true }}
	if f.glinting() {
		t.Error("a redraw for a glint nobody painted")
	}
	for f.glint() != 0 { // into the rest
		now = now.Add(mapAnimStep)
	}
	if f.glinting() {
		t.Error("a redraw while the glint rests")
	}
	for i := 0; i < actGlintRest+1; i++ {
		now = now.Add(mapAnimStep)
	}
	if !f.glinting() || f.glinting() {
		t.Error("the due glint asks once for a redraw")
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
	g := actionGrid(actionSpec{kind: keyPrimary, label: "send", key: "Enter", cmd: "expedition mist"},
		actionSpec{kind: keyPlain, label: "back", key: "Esc"})
	k := keyPresser{press: func(kit.Key) {}, run: func(string) {}}
	move := func(x, y int, kind kit.MouseKind) {
		k.mouse(g, image.Point{}, kit.Mouse{Kind: kind, X: x, Y: y})
	}
	move(3, 2, kit.MouseMove)
	if k.feel.preview != "expedition mist" {
		t.Fatalf("resting on the button previews %q", k.feel.preview)
	}
	k.feel.typed(kit.Key{Code: kit.KeyRune, Rune: 'a'})
	move(4, 2, kit.MouseMove)
	if k.feel.preview != "" {
		t.Errorf("a typed key left %q", k.feel.preview)
	}
	move(40, 0, kit.MouseMove)
	move(3, 2, kit.MouseMove)
	if k.feel.preview != "expedition mist" {
		t.Errorf("coming back previews %q", k.feel.preview)
	}
	move(g.keys[1].r.Min.X+1, 2, kit.MouseMove)
	if k.feel.preview != "" {
		t.Errorf("a button with no line previews %q", k.feel.preview)
	}
	move(3, 2, kit.MouseMove)
	move(0, 0, kit.MouseLeave)
	if k.feel.preview != "" {
		t.Errorf("the pointer left and %q stayed", k.feel.preview)
	}
}

func TestAButtonsLineEntersTheHistoryAndTheBarLikeATypedOne(t *testing.T) {
	d, _ := newTestDashboard(nil)
	var k keyPresser
	k.wire(d.app)
	g := actionGrid(actionSpec{kind: keyPrimary, label: "go", key: "Enter", cmd: "help"})
	click(&k, g, 3, 2)
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
	if len(p.model.keys.keys) != 1 || p.model.keys.keys[0].off || p.model.keys.keys[0].kind != keyPrimary {
		t.Fatalf("the card of a tech that can start has %+v, want one primary button", p.model.keys.keys)
	}
	b := p.model.keys.keys[0]
	at := b.r.Min.Add(image.Pt(2, 1))
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
