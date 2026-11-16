package ui

import (
	"fmt"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
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

// paintLive paints the pages at w by h into a kit buffer, as the window does.
func paintLive(pages *nativePages, w, h int) *kit.Buffer {
	pages.SetRect(0, 0, w, h)
	root := kit.NewRoot(pages)
	root.Resize(w, h)
	var b kit.Buffer
	root.Paint(&b)
	return &b
}

// bufText is the text of the rectangle r of b, a line per row.
func bufText(b *kit.Buffer, r image.Rectangle) string {
	var sb strings.Builder
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			sb.WriteRune(b.At(x, y).R)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// TestBadgeMomentOnTheLiveScreen: an earned badge shows its card in the top
// right of the dashboard, under the header's buttons, at every size, and not
// in the header's notice; the card moves with motion on and is whole and
// still with it off; it is gone after its life, and badges earned together
// take it one after another.
func TestBadgeMomentOnTheLiveScreen(t *testing.T) {
	for _, size := range treeSizes {
		w, h := size[0], size[1]
		d, pages, eng := badgeDashboard(t)
		drawnDashboard(t, d, pages, w, h) // every box gets its size
		clock := time.Unix(1_700_000_000, 0)
		d.fxNow = func() time.Time { return clock }
		d.miniMap.now = d.fxNow
		must(t, eng.Account().SetMotion(true))
		where := fmt.Sprintf("%dx%d", w, h)

		// What the dashboard shows with no card painted over it.
		bare := func() *kit.Buffer {
			over := d.rootFx.over
			d.rootFx.over = nil
			defer func() { d.rootFx.over = over }()
			return paintLive(pages, w, h)
		}
		card := image.Rect(w-momentW-1, 1, w-1, 1+momentH)

		// Two badges together: the Stone Age and a prestige rung. The age is
		// reported alone: an age entered for real raises its splash.
		eng.ReportForTest(config.BadgeEvAgeReached, "stone_age")
		for i := 0; i < 10; i++ {
			eng.ReportForTest(config.BadgeEvPrestige, "modern_age")
		}
		d.refresh()
		var names []string
		views, _ := eng.Badges()
		for _, l := range eng.GetLogs() {
			for _, v := range views {
				if l.Message == game.BadgeLogLine(v) {
					names = append(names, v.Name)
				}
			}
		}
		if len(names) < 2 {
			t.Fatalf("%s: the log carries %d badge lines, want two", where, len(names))
		}

		// The card opens, then plays whole.
		b := paintLive(pages, w, h)
		if !d.momentOn.Load() {
			t.Errorf("%s: the card does not move with motion on", where)
		}
		if row := bufText(b, image.Rect(0, 0, w, 1)); strings.Contains(row, "Badge earned") {
			t.Errorf("%s: the header's notice has a badge toast: %q", where, row)
		}
		clock = clock.Add(momentOpen * coinStep)
		b = paintLive(pages, w, h)
		if !d.momentOn.Load() {
			t.Errorf("%s: the whole card is still with motion on", where)
		}
		if len(b.Pics) == 0 {
			t.Fatalf("%s: the card has no coin", where)
		}
		var coin kit.Pic
		for _, p := range b.Pics {
			if p.Rect.Min == image.Pt(card.Min.X+2, card.Min.Y+1) {
				coin = p
			}
		}
		if coin.Img == nil {
			t.Fatalf("%s: no picture at the card's coin place %v; pictures: %v", where, image.Pt(card.Min.X+2, card.Min.Y+1), b.Pics)
		}
		if coin.Rect.Min.Y != 2 {
			t.Errorf("%s: the coin starts on row %d, want 2", where, coin.Rect.Min.Y)
		}
		if b.At(coin.Rect.Min.X, coin.Rect.Min.Y).Pic == 0 || oneColour(coin.Img) {
			t.Errorf("%s: the coin's cells do not own it, or it is one flat colour", where)
		}
		if text := bufText(b, card); !strings.Contains(text, names[0]) || strings.Contains(text, names[1]) {
			t.Errorf("%s: the card shows %q, want the first badge %q alone", where, text, names[0])
		}
		if row := bufText(b, image.Rect(0, 0, w, 1)); strings.Contains(row, "Badge earned") {
			t.Errorf("%s: the header's notice has a badge toast: %q", where, row)
		}

		// With motion off the card is whole and still.
		must(t, eng.Account().SetMotion(false))
		still := paintLive(pages, w, h)
		clock = clock.Add(20 * coinStep)
		later := paintLive(pages, w, h)
		if d.momentOn.Load() {
			t.Errorf("%s: the card moves with motion off", where)
		}
		if bufText(still, card) != bufText(later, card) || bufText(later, card) == bufText(bare(), card) {
			t.Errorf("%s: the still card changed, or is not on show", where)
		}
		if bottom := later.At(card.Min.X, card.Max.Y-1).R; bottom == ' ' {
			t.Errorf("%s: the still card is not whole", where)
		}
		must(t, eng.Account().SetMotion(true))

		// After its life the first card is gone and the second takes over.
		clock = clock.Add(momentLife * coinStep)
		gone := paintLive(pages, w, h)
		if bufText(gone, card) != bufText(bare(), card) || len(gone.Pics) != len(bare().Pics) {
			t.Errorf("%s: a card is left after its life", where)
		}
		paintLive(pages, w, h) // the next card begins at this paint
		clock = clock.Add(momentOpen * coinStep)
		b = paintLive(pages, w, h)
		if text := bufText(b, card); !strings.Contains(text, names[1]) || strings.Contains(text, names[0]) {
			t.Errorf("%s: the second card shows %q, want %q", where, text, names[1])
		}

		// And after the last, the dashboard shows again.
		for range names {
			clock = clock.Add(momentLife * coinStep)
			paintLive(pages, w, h)
			paintLive(pages, w, h) // the next card begins at this paint
		}
		clock = clock.Add(momentLife * coinStep)
		paintLive(pages, w, h)
		end := paintLive(pages, w, h)
		if ref := bare(); bufText(end, card) != bufText(ref, card) || len(end.Pics) != len(ref.Pics) || d.momentOn.Load() {
			t.Errorf("%s: a card is left after the last", where)
		}
	}
}

// TestBadgeMomentWaitsUnderAPanel: a badge earned while a panel covers the
// dashboard keeps its card until the panel closes, then shows it.
func TestBadgeMomentWaitsUnderAPanel(t *testing.T) {
	const w, h = 160, 48
	d, pages, eng := badgeDashboard(t)
	drawnDashboard(t, d, pages, w, h)
	clock := time.Unix(1_700_000_000, 0)
	d.fxNow = func() time.Time { return clock }
	d.miniMap.now = d.fxNow
	must(t, eng.Account().SetMotion(true))

	if !d.overlayMgr.Show("techs", eng.GetState()) {
		t.Fatal("the Research panel did not open")
	}
	eng.ReportForTest(config.BadgeEvAgeReached, "stone_age")
	d.refresh()
	if len(d.moment.queue) != 1 {
		t.Fatalf("%d badges wait for the card, want 1", len(d.moment.queue))
	}
	name := d.moment.queue[0].Name
	paintLive(pages, w, h)
	clock = clock.Add(2 * momentLife * coinStep)
	paintLive(pages, w, h)
	if !d.moment.start.IsZero() || len(d.moment.queue) != 1 || d.momentOn.Load() {
		t.Fatalf("the card ran under the panel: start %v, %d queued", d.moment.start, len(d.moment.queue))
	}

	d.overlayMgr.Hide()
	paintLive(pages, w, h)
	clock = clock.Add(momentOpen * coinStep)
	b := paintLive(pages, w, h)
	card := image.Rect(w-momentW-1, 1, w-1, 1+momentH)
	if text := bufText(b, card); !strings.Contains(text, name) {
		t.Errorf("the card does not show once the panel closes: %q", text)
	}
}
