package ui

import (
	"image"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/window/kit"
)

// The earned-badge card draws its coin as a picture in every glyph tier, and
// in the plain tier folds its frame and text to ASCII.
func TestBadgeMomentDrawsItsCoinInEveryTier(t *testing.T) {
	views, _ := caseFixture()
	var earned game.BadgeView
	for _, v := range views {
		if v.Earned {
			earned = v
			break
		}
	}
	if earned.Key == "" {
		t.Fatal("the fixture has no earned badge")
	}
	for _, tier := range caseTiers {
		m := &badgeMoment{}
		m.add(earned)
		b := kit.NewBuffer(80, 24)
		now := time.Unix(1_700_000_000, 0)
		m.paint(kit.NewPainter(b), image.Rect(0, 0, 80, 24), now, false, tier)
		if len(b.Pics) != 1 {
			t.Fatalf("%s: the card painted %d pictures, want its coin", tier, len(b.Pics))
		}
		pic := b.Pics[0]
		if int(b.At(pic.Rect.Min.X, pic.Rect.Min.Y).Pic) != 1 || oneColour(pic.Img) {
			t.Errorf("%s: the coin does not own its cells or is one flat colour", tier)
		}
		if tier == mapmodel.TierASCII {
			for y := 0; y < 24; y++ {
				for x := 0; x < 80; x++ {
					if c := b.At(x, y); c.R >= 0x80 {
						t.Fatalf("the plain tier draws %q at %d,%d", c.R, x, y)
					}
				}
			}
		}
	}
}
