package ui

import (
	"bytes"
	"image"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// A medal's picture is a coin on a clear ground: solid in the middle, clear
// in the corners, the same every time for the same medal, and different for
// a medal not earned, another tier and another family.
func TestMedalPicture(t *testing.T) {
	bronze := medal{tier: config.BadgeBronze, state: medalEarned}
	const w, h = 180, 189
	img := medalPicture(bronze, "age", w, h)
	if img.Bounds() != image.Rect(0, 0, w, h) {
		t.Fatalf("the picture is %v, want %dx%d", img.Bounds(), w, h)
	}
	if img.RGBAAt(w/2, h/2).A != 255 || img.RGBAAt(2, 2).A != 0 || img.RGBAAt(w-3, h-3).A != 0 {
		t.Errorf("the coin should be solid in the middle and clear in the corners: %v %v %v", img.RGBAAt(w/2, h/2), img.RGBAAt(2, 2), img.RGBAAt(w-3, h-3))
	}
	if again := medalPicture(bronze, "age", w, h); !bytes.Equal(img.Pix, again.Pix) {
		t.Error("the same medal drew two different pictures")
	}
	for what, other := range map[string]*image.RGBA{
		"a medal not earned": medalPicture(medal{tier: config.BadgeBronze, state: medalLocked}, "age", w, h),
		"a silver medal":     medalPicture(medal{tier: config.BadgeSilver, state: medalEarned}, "age", w, h),
		"another family":     medalPicture(bronze, "research", w, h),
		"a legendary medal":  medalPicture(medal{tier: config.BadgeLegendary, state: medalEarned}, "age", w, h),
	} {
		if bytes.Equal(img.Pix, other.Pix) {
			t.Errorf("%s drew the same picture as an earned bronze one", what)
		}
	}
	// Every sign is thirteen by thirteen.
	for k, icon := range medalIcons {
		if len(icon) != 13 {
			t.Errorf("the %s sign has %d rows", k, len(icon))
		}
		for _, row := range icon {
			if len(row) != 13 {
				t.Errorf("the %s sign has a row of %d", k, len(row))
			}
		}
	}
	// Every family of the case has a sign of its own, by its key: no two
	// families share one, and only Specials carry the star.
	seen := map[string]string{}
	for _, f := range badgeFamilyTitles {
		sign := medalSign(f.key)
		if medalIcons[sign] == nil {
			t.Errorf("family %s has the sign %q, which is not drawn", f.key, sign)
		}
		if sign == "star" && f.key != "special" {
			t.Errorf("family %s has no sign of its own", f.key)
		}
		if other, dup := seen[sign]; dup {
			t.Errorf("families %s and %s share the %s sign", other, f.key, sign)
		}
		seen[sign] = f.key
	}
	// A medal earned in a modified game is struck through.
	earned := medal{tier: config.BadgeBronze, state: medalEarned}
	crossed := earned
	crossed.crossed = true
	if bytes.Equal(coinPicture(lookOf(crossed, "age", coinArt), 84, 72, coinPose{}).Pix, coinPicture(lookOf(earned, "age", coinArt), 84, 72, coinPose{}).Pix) {
		t.Error("a crossed medal draws the same coin as one earned fairly")
	}
	// The small coin is the medal at half the art: whole, centred, and not
	// the large one shrunk.
	small := coinPicture(lookOf(earned, "age", coinArt), 70, 63, coinPose{})
	if small.Rect.Dx() != 70 || small.Rect.Dy() != 63 || small.RGBAAt(0, 0).A != 0 || small.RGBAAt(35, 31).A == 0 {
		t.Errorf("the small coin is %v with corner alpha %d and centre alpha %d", small.Rect, small.RGBAAt(0, 0).A, small.RGBAAt(35, 31).A)
	}
}

// Where pictures can show, the case draws every badge of the grid as a small
// coin and the selected badge's medal large in the pane; where they cannot,
// the medals are drawn in cells as before and there is no picture.
func TestBadgeCaseShowsItsMedalsAsPictures(t *testing.T) {
	views, sum := caseFixture()
	st := game.GameState{AccountStats: &game.AccountStatsView{Badges: views, BadgeSummary: sum}}
	clock := time.Unix(1_700_000_000, 0)
	p := newBadgePanel()
	p.now, p.start = func() time.Time { return clock }, clock
	p.SetRect(0, 0, 120, 30)
	p.open(st)
	paint := func() *kit.Buffer {
		b := kit.NewBuffer(120, 30)
		p.Layout(image.Rect(0, 0, 120, 30))
		p.Paint(kit.NewPainter(b))
		return b
	}
	if b := paint(); len(b.Pics) != 0 || p.model.lay.tw != miniW {
		t.Fatalf("with no window to draw pictures the case painted %d, on tiles %d wide", len(b.Pics), p.model.lay.tw)
	}
	was := pictureCell
	pictureCell = func() (int, int) { return 10, 21 }
	t.Cleanup(func() { pictureCell = was })
	b := paint()
	lay := p.model.lay
	if !lay.pix || lay.tw != coinCols || lay.th != coinRows {
		t.Fatalf("the case is laid out for tiles %d by %d (pix %v), want coins", lay.tw, lay.th, lay.pix)
	}
	// The coins of the grid, then the medal of the pane.
	coins := 0
	for _, ln := range p.model.lines {
		if ln.kind == lineBadges && p.model.visible(ln, p.view.vy) {
			coins += len(ln.items)
		}
	}
	if coins == 0 || len(b.Pics) != coins+1 {
		t.Fatalf("the case painted %d pictures for %d badges in view and the pane's medal", len(b.Pics), coins)
	}
	looks := map[*image.RGBA]bool{}
	for _, pic := range b.Pics[:coins] {
		if pic.Rect.Dx() != coinCols || pic.Rect.Dy() != coinRows || pic.Rect.Max.X > lay.gx+lay.gw {
			t.Errorf("a coin's room is %v, want %d by %d inside the grid", pic.Rect, coinCols, coinRows)
		}
		if pic.Img.Bounds().Dx() != coinCols*10 || pic.Img.Bounds().Dy() != coinRows*21 {
			t.Errorf("a coin's picture is %v pixels, want a pixel for a pixel of its room", pic.Img.Bounds())
		}
		looks[pic.Img] = true
	}
	if len(looks) >= coins {
		t.Errorf("%d coins are %d pictures: coins that look alike should share one", coins, len(looks))
	}
	pic := b.Pics[coins]
	if pic.Rect.Dx() != medalPixCols || pic.Rect.Dy() != medalPixRows || pic.Rect.Min.X < lay.px {
		t.Errorf("the medal's room is %v, want %d by %d inside the pane (from column %d)", pic.Rect, medalPixCols, medalPixRows, lay.px)
	}
	if pic.Img.Bounds().Dx() != medalPixCols*10 || pic.Img.Bounds().Dy() != medalPixRows*21 {
		t.Errorf("the medal's picture is %v pixels, want a pixel for a pixel of its room", pic.Img.Bounds())
	}
	if c := b.At(pic.Rect.Min.X+1, pic.Rect.Min.Y+1); int(c.Pic) != coins+1 {
		t.Errorf("a cell of the medal's room shows picture %d", c.Pic)
	}
	// The same case at the same size is the same images, frame after frame.
	again := paint()
	for i := range b.Pics {
		if again.Pics[i].Img != b.Pics[i].Img {
			t.Fatalf("a second frame struck picture %d again", i)
		}
	}
	// A click on a coin picks its badge, at the coin's own place.
	second, want := b.Pics[1].Rect, p.model.coins[1].it.id
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: second.Min.X + 2, Y: second.Min.Y + 1, Clicks: 1})
	if got := p.model.selected(p.view); got == nil || got.id != want {
		t.Errorf("a click on the second coin picked %v", got)
	}
	// With the card open nothing is painted over it.
	p.view.card = true
	if over := paint(); len(over.Pics) != 0 {
		t.Errorf("the open card has %d pictures painted over it", len(over.Pics))
	}
}
