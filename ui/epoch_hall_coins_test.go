package ui

import (
	"image"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/window/kit"
)

// The page records a coin for each badge it shows: large in the full
// layout, small where there is little room.
func TestEpochHallRecordsCoins(t *testing.T) {
	st := lateHallState(t, "galactic_age")
	for _, tc := range []struct {
		name       string
		cols, rows int
		w, h, art  int
	}{
		{"full", 200, 50, hallCoinW, hallCoinH, medalArt},
		{"small", 200, 30, hallMiniCoinW, hallMiniCoinH, coinArt},
	} {
		page := hallFrontPage(st, tc.cols, tc.rows, mapmodel.TierUnicode, hallMotion{}, true)
		if len(page.coins) != hallBadgeCount {
			t.Fatalf("%s: %d coins, want %d", tc.name, len(page.coins), hallBadgeCount)
		}
		for i, c := range page.coins {
			if c.r.Dx() != tc.w || c.r.Dy() != tc.h || c.look.art != tc.art || c.i != i {
				t.Errorf("%s: coin %d is %v art %d order %d, want %dx%d art %d", tc.name, i, c.r, c.look.art, c.i, tc.w, tc.h, tc.art)
			}
		}
	}
}

// Painted in a text view, the page's coins are one picture each, on the
// cells the page set aside, and not drawn in cells.
func TestEpochHallPaintsCoinsOverTheText(t *testing.T) {
	was := pictureCell
	pictureCell = func() (int, int) { return 10, 21 }
	t.Cleanup(func() { pictureCell = was })
	st := lateHallState(t, "galactic_age")
	page := hallFrontPage(st, 120, 50, mapmodel.TierUnicode, hallMotion{}, true)
	tv := newNativeText().SetDynamicColors(true).SetWrap(true)
	tv.SetText(safeTags(page.text))
	pics := &hallCoins{list: page.coins}
	tv.over = func(p *kit.Painter, text image.Rectangle, top int) { pics.paint(p, text, top, atRest) }
	b := paintAlone(tv, 120, 200)
	if len(b.Pics) != len(page.coins) || len(b.Pics) == 0 {
		t.Fatalf("%d pictures for %d coins", len(b.Pics), len(page.coins))
	}
	org := tv.view.Bounds().Min
	for i, c := range page.coins {
		if got, want := b.Pics[i].Rect, c.r.Add(org); got != want {
			t.Errorf("coin %d is on %v, want %v", i, got, want)
		}
		if b.Pics[i].Img.Bounds().Dx() != c.r.Dx()*10 {
			t.Errorf("coin %d picture is %v wide", i, b.Pics[i].Img.Bounds())
		}
	}
	if pics.moving.Load() {
		t.Error("coins at rest say they move")
	}
}

// The card's picture shows the coins: inside each coin's room it differs
// from the same card with no coins painted.
func TestEpochPictureShowsCoins(t *testing.T) {
	st := lateHallState(t, "galactic_age")
	img := epochPicture(st)
	coins := hallFrontPage(st, epochPicCols, epochLayoutRows, mapmodel.TierUnicode, hallMotion{}, false).coins
	if len(coins) == 0 {
		t.Fatal("no coins on the card's page")
	}
	frame, tv, rows := epochCardFrame(st)
	tv.over = nil
	bare := ScreenImage(paintAlone(frame, epochPicCols, rows), shotCellW, shotCellH)
	org := tv.view.Bounds().Min
	for i, c := range coins {
		r := c.r.Add(org)
		found := false
		for y := shotPad + r.Min.Y*shotCellH; y < shotPad+r.Max.Y*shotCellH && !found; y++ {
			for x := shotPad + r.Min.X*shotCellW; x < shotPad+r.Max.X*shotCellW; x++ {
				if img.RGBAAt(x, y) != bare.RGBAAt(x, y) {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("coin %d: its room on the card is the same as with no coin", i)
		}
	}
}

// With a window to draw pictures in, the Badges panel draws coins in the
// plain glyph tier as it does in the others.
func TestBadgePanelDrawsCoinsInThePlainTier(t *testing.T) {
	p, paint, _ := picturePanel(t)
	p.view.tier = mapmodel.TierASCII
	b := paint()
	if !p.pix() || len(b.Pics) == 0 || !p.model.lay.pix {
		t.Fatalf("the plain tier painted %d pictures (pix %v)", len(b.Pics), p.model.lay.pix)
	}
}
