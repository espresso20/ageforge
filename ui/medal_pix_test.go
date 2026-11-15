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
	// Every family word has its own sign, thirteen by thirteen.
	for _, k := range append([]string{"star", "locked", "hidden"}, medalIconOrder...) {
		icon := medalIcons[k]
		if len(icon) != 13 {
			t.Errorf("the %s sign has %d rows", k, len(icon))
		}
		for _, row := range icon {
			if len(row) != 13 {
				t.Errorf("the %s sign has a row of %d", k, len(row))
			}
		}
	}
}

// Where pictures can show, the Badges pane leaves room for the selected
// badge's medal and the panel paints it as a picture; where they cannot, the
// medal is drawn in cells as before and there is no picture.
func TestBadgePaneShowsItsMedalAsAPicture(t *testing.T) {
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
	if b := paint(); len(b.Pics) != 0 {
		t.Fatalf("with no window to draw pictures the pane painted %d", len(b.Pics))
	}
	was := pictureCell
	pictureCell = func() (int, int) { return 10, 21 }
	t.Cleanup(func() { pictureCell = was })
	b := paint()
	if len(b.Pics) != 1 {
		t.Fatalf("the pane painted %d pictures, want the selected badge's medal", len(b.Pics))
	}
	pic := b.Pics[0]
	if pic.Rect.Dx() != medalPixCols || pic.Rect.Dy() != medalPixRows || pic.Rect.Min.X < p.model.lay.px {
		t.Errorf("the medal's room is %v, want %d by %d inside the pane (from column %d)", pic.Rect, medalPixCols, medalPixRows, p.model.lay.px)
	}
	if pic.Img.Bounds().Dx() != medalPixCols*10 || pic.Img.Bounds().Dy() != medalPixRows*21 {
		t.Errorf("the picture is %v pixels, want a pixel for a pixel of its room", pic.Img.Bounds())
	}
	if c := b.At(pic.Rect.Min.X+1, pic.Rect.Min.Y+1); c.Pic != 1 {
		t.Errorf("a cell of the medal's room shows picture %d", c.Pic)
	}
	// The same badge at the same size is the same image, frame after frame.
	if again := paint(); again.Pics[0].Img != pic.Img {
		t.Error("a second frame drew the medal again")
	}
}
