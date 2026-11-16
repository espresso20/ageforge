package ui

import (
	"bytes"
	"image"
	"testing"
	"time"

	"github.com/espresso20/ageforge/window/kit"
)

// withPictures lets the buttons draw plates for a 10 by 21 cell.
func withPictures(t *testing.T) {
	t.Helper()
	was := pictureCell
	pictureCell = func() (int, int) { return 10, 21 }
	t.Cleanup(func() { pictureCell = was })
}

func buttonGrid() *xgrid {
	g := newXGrid(60, 2)
	x := g.keyButton(1, 1, "Enter", "build") + 1
	g.keyButton(x, 1, "↑↓", "move")
	return g
}

func TestPlatesAreCachedAndEachStateDiffers(t *testing.T) {
	keepTheme(t)
	pal := newExpPalette()
	var set plateSet
	plate := func(kind keyKind, st keyState) *image.RGBA {
		return set.plate(pal.look(kind, st), st, 7, 10, 21)
	}
	if plate(keyAccent, keyRest) != plate(keyAccent, keyRest) {
		t.Error("the same key twice is two pictures")
	}
	seen := map[keyState]*image.RGBA{}
	for _, st := range []keyState{keyRest, keyHover, keyPressed, keyOff} {
		img := plate(keyAccent, st)
		if img.Rect.Dx() != 70 || img.Rect.Dy() != 21 {
			t.Errorf("state %d plate is %v, want the cap's cells in pixels", st, img.Rect)
		}
		for o, other := range seen {
			if bytes.Equal(img.Pix, other.Pix) {
				t.Errorf("states %d and %d look alike", st, o)
			}
		}
		seen[st] = img
	}
	if bytes.Equal(plate(keySteel, keyRest).Pix, plate(keyAccent, keyRest).Pix) {
		t.Error("a steel key looks like an accent key")
	}
	set.plate(pal.look(keyAccent, keyRest), keyRest, 7, 12, 24)
	if len(set.m) != 1 {
		t.Errorf("a new cell size left %d plates", len(set.m))
	}
}

func TestPaintLaysPlatesOnlyWithPictures(t *testing.T) {
	keepTheme(t)
	pal := newExpPalette()
	g := buttonGrid()
	paintWith := func() *kit.Buffer {
		b := kit.NewBuffer(60, 2)
		p := kit.NewPainter(b)
		var k keyPresser
		k.press = func(kit.Key) {}
		paintExpeditionKit(p, g, image.Point{}, pal)
		k.paint(p, g, image.Point{}, pal)
		return b
	}
	plain := paintWith()
	for x := 1; x < 8; x++ {
		if c := plain.At(x, 1); c.Pic != 0 || c.R != []rune(" Enter ")[x-1] {
			t.Errorf("without pictures cell %d is %+v", x, c)
		}
	}
	if len(plain.Pics) != 0 {
		t.Errorf("without pictures %d pictures", len(plain.Pics))
	}
	withPictures(t)
	pix := paintWith()
	if len(pix.Pics) != 2 {
		t.Fatalf("two keys painted %d pictures", len(pix.Pics))
	}
	for x := 1; x < 8; x++ {
		if c := pix.At(x, 1); c.Pic == 0 || c.R != []rune(" Enter ")[x-1] {
			t.Errorf("with pictures cell %d is %+v, want the plate and the cap text", x, c)
		}
	}
	g.fold()
	if len(paintWith().Pics) != 0 {
		t.Error("a folded grid still got plates")
	}
}

func TestAMousePressSinksTheKeyUntilReleaseAndATypedKeyForTwoSteps(t *testing.T) {
	keepTheme(t)
	g := buttonGrid()
	now := time.Unix(100, 0)
	var pressed []kit.Key
	k := keyPresser{press: func(k kit.Key) { pressed = append(pressed, k) }, feel: &keyFeel{now: func() time.Time { return now }}}
	enter := kit.Key{Code: kit.KeyEnter}
	k.mouse(g, image.Point{}, kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: 3, Y: 1})
	if len(pressed) != 1 || pressed[0] != enter {
		t.Fatalf("the press sent %v", pressed)
	}
	now = now.Add(10 * coinStep)
	if !k.feel.sunk(enter) {
		t.Error("the key came up while the button was down")
	}
	k.mouse(g, image.Point{}, kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: 3, Y: 1})
	now = now.Add(coinStep)
	if k.feel.sunk(enter) {
		t.Error("a released key stayed sunk past its steps")
	}
	k.feel.typed(enter)
	now = now.Add(coinStep)
	if !k.feel.sunk(enter) || k.feel.sunk(kit.Key{Code: kit.KeyEsc}) {
		t.Error("a typed key should sink its own button for a step")
	}
	now = now.Add(coinStep + coinStep/2)
	if k.feel.sunk(enter) {
		t.Error("a typed key stayed sunk after two steps")
	}
	if !k.feel.busy() {
		t.Error("one more repaint is due after the sink ends")
	}
	now = now.Add(2 * coinStep)
	if k.feel.busy() {
		t.Error("still busy long after")
	}
}

func TestPressedKeysShowSunkInPixels(t *testing.T) {
	keepTheme(t)
	withPictures(t)
	pal := newExpPalette()
	g := buttonGrid()
	now := time.Unix(100, 0)
	k := keyPresser{press: func(kit.Key) {}, feel: &keyFeel{now: func() time.Time { return now }}}
	paint := func() *image.RGBA {
		b := kit.NewBuffer(60, 2)
		p := kit.NewPainter(b)
		paintExpeditionKit(p, g, image.Point{}, pal)
		k.paint(p, g, image.Point{}, pal)
		return b.Pics[0].Img
	}
	rest := paint()
	k.feel.typed(kit.Key{Code: kit.KeyEnter})
	if bytes.Equal(rest.Pix, paint().Pix) {
		t.Error("a typed Enter did not sink the Enter button")
	}
	now = now.Add(5 * coinStep)
	if again := paint(); again != rest {
		t.Error("after the sink the button is not the rest picture again")
	}
}

func TestKeysOffAreDrawnAndNotPressed(t *testing.T) {
	g := buttonGrid()
	g.keysOff(0)
	if _, ok := g.keyAt(image.Pt(3, 1)); !ok {
		t.Error("an unavailable key is still a button")
	}
	if !g.keys[0].off {
		t.Error("keysOff left a key on")
	}
}

func TestFootLayoutDoesNotChangeWithTheButtons(t *testing.T) {
	keepTheme(t)
	keys := [][2]string{{"↑↓", "scroll"}, {"PgUp PgDn", "page"}, {"Esc", "close"}}
	for _, w := range []int{60, 80, 120, 160} {
		g := renderPanelFrame(w, 10, "T", keys, "", "")
		x := 2
		for i, kv := range keys {
			if i > 0 {
				x++
			}
			x += keyButtonW(kv[0], kv[1])
		}
		line := g.String()
		want := " ↑↓  scroll   PgUp PgDn  page   Esc  close "
		if !bytes.Contains([]byte(line), []byte(want)) {
			t.Errorf("at %d wide the foot lacks %q:\n%s", w, want, line)
		}
		if got := keyButtonW("Esc", "close"); got != 12 {
			t.Errorf("keyButtonW = %d", got)
		}
		_ = x
	}
}
