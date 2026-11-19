package ui

import (
	"image"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/coins"
)

// The page records a coin for each badge it shows: large in the full
// layout, small where there is little room.
func TestEpochHallRecordsCoins(t *testing.T) {
	st := lateHallState(t, "galactic_age")
	for _, tc := range []struct {
		name       string
		cols, rows int
		w, h       int
	}{
		{"full", 200, 50, hallCoinW, hallCoinH},
		{"small", 200, 30, hallMiniCoinW, hallMiniCoinH},
	} {
		page := hallFrontPage(st, tc.cols, tc.rows, mapmodel.TierUnicode, hallMotion{}, true)
		if len(page.coins) != hallBadgeCount {
			t.Fatalf("%s: %d coins, want %d", tc.name, len(page.coins), hallBadgeCount)
		}
		for i, c := range page.coins {
			if c.r.Dx() != tc.w || c.r.Dy() != tc.h || c.coin.Type == "" {
				t.Errorf("%s: coin %d is %v of %q, want %dx%d with a coin", tc.name, i, c.r, c.coin.Type, tc.w, tc.h)
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
	pics := &hallCoins{list: page.coins, mint: coins.NewMint(16)}
	tv.over = pics.paint
	b := paintAlone(tv, 120, 200)
	if len(b.Pics) != len(page.coins) || len(b.Pics) == 0 {
		t.Fatalf("%d pictures for %d coins", len(b.Pics), len(page.coins))
	}
	org := tv.view.Bounds().Min
	for i, c := range page.coins {
		if got, want := b.Pics[i].Rect, c.r.Add(org); got != want {
			t.Errorf("coin %d is on %v, want %v", i, got, want)
		}
		if b.Pics[i].Img.Bounds().Dx() > c.r.Dx()*10 || b.Pics[i].Img.Bounds().Dx() != b.Pics[i].Img.Bounds().Dy() {
			t.Errorf("coin %d picture is %v, not a square that fits", i, b.Pics[i].Img.Bounds())
		}
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

// oneColour reports whether every pixel of img is the same.
func oneColour(img *image.RGBA) bool {
	first := img.RGBAAt(img.Rect.Min.X, img.Rect.Min.Y)
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			if img.RGBAAt(x, y) != first {
				return false
			}
		}
	}
	return true
}
