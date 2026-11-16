package ui

import (
	"bytes"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
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

// The case draws every badge of the grid as a small coin and the selected
// badge's medal large in the pane.
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
	was := pictureCell
	pictureCell = func() (int, int) { return 10, 21 }
	t.Cleanup(func() { pictureCell = was })
	b := paint()
	lay := p.model.lay
	if lay.tw != coinCols || lay.th != coinRows {
		t.Fatalf("the case is laid out for tiles %d by %d, want coins", lay.tw, lay.th)
	}
	// The coins of the grid, then the medal of the pane.
	coins := 0
	for _, ln := range p.model.lines {
		if ln.kind == lineBadges && p.model.visible(ln, p.view.vy) {
			coins += len(ln.items)
		}
	}
	if coins == 0 || len(coinPics(b)) != coins+1 {
		t.Fatalf("the case painted %d pictures for %d badges in view and the pane's medal", len(coinPics(b)), coins)
	}
	looks := map[*image.RGBA]bool{}
	for _, pic := range coinPics(b)[:coins] {
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
	pic := coinPics(b)[coins]
	if pic.Rect.Dx() != medalPixCols || pic.Rect.Dy() != medalPixRows || pic.Rect.Min.X < lay.px {
		t.Errorf("the medal's room is %v, want %d by %d inside the pane (from column %d)", pic.Rect, medalPixCols, medalPixRows, lay.px)
	}
	if pic.Img.Bounds().Dx() != medalPixCols*10 || pic.Img.Bounds().Dy() != medalPixRows*21 {
		t.Errorf("the medal's picture is %v pixels, want a pixel for a pixel of its room", pic.Img.Bounds())
	}
	if c := b.At(pic.Rect.Min.X+1, pic.Rect.Min.Y+1); c.Pic == 0 || b.Pics[c.Pic-1].Img != pic.Img {
		t.Errorf("a cell of the medal's room shows picture %d", c.Pic)
	}
	// With the motion off the same case at the same size is the same images,
	// frame after frame.
	p.motion = false
	b = paint()
	again := paint()
	if len(coinPics(again)) != len(coinPics(b)) {
		t.Fatalf("a second frame painted %d pictures, the first %d", len(coinPics(again)), len(coinPics(b)))
	}
	for i := range coinPics(b) {
		if coinPics(again)[i].Img != coinPics(b)[i].Img {
			t.Fatalf("a second frame struck picture %d again", i)
		}
	}
	p.motion = true
	// A click on a coin picks its badge, at the coin's own place.
	second, want := coinPics(b)[1].Rect, p.model.coins[1].it.id
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: second.Min.X + 2, Y: second.Min.Y + 1, Clicks: 1})
	if got := p.model.selected(p.view); got == nil || got.id != want {
		t.Errorf("a click on the second coin picked %v", got)
	}
	// With the card open the grid's coins are not painted: one picture, the
	// card's medal, lies inside the card.
	p.view.card = true
	over := paint()
	if len(coinPics(over)) != 1 {
		t.Fatalf("the open card has %d pictures painted over it, want its medal alone", len(coinPics(over)))
	}
	if room := coinPics(over)[0].Rect; !room.In(p.model.cardAt) {
		t.Errorf("the card's medal is at %v, outside the card %v", room, p.model.cardAt)
	}
}

func earnedLook(tier config.BadgeTier) coinLook {
	return lookOf(medal{tier: tier, state: medalEarned}, "age", medalArt)
}

// poseCount is how many different poses a look takes in one loop of its show.
func poseCount(l coinLook, loop int) int {
	seen := map[coinPose]bool{}
	for n := 0; n < loop; n++ {
		seen[coinPoseAt(l, n)] = true
	}
	return len(seen)
}

// A coin not earned and a coin at rest have no show; every show goes round;
// and each tier's show has more in it than the tier below's.
func TestCoinPoseAt(t *testing.T) {
	tiers := []config.BadgeTier{config.BadgeBronze, config.BadgeSilver, config.BadgeGold, config.BadgePlatinum, config.BadgeLegendary}
	for _, tier := range tiers {
		l := earnedLook(tier)
		if got := coinPoseAt(l, atRest); got != (coinPose{}) {
			t.Errorf("%v at rest has the pose %+v", tier, got)
		}
		loose := lookOf(medal{tier: tier, state: medalLocked}, "age", medalArt)
		for _, n := range []int{0, 5, 77, 1 << 16} {
			if got := coinPoseAt(loose, n); got != (coinPose{}) {
				t.Errorf("a %v coin not earned has the pose %+v at %d", tier, got, n)
			}
		}
		loop := coinShows[tier].loop
		for n := 0; n < loop; n++ {
			if coinPoseAt(l, n) != coinPoseAt(l, n+loop) || coinPoseAt(l, n) != coinPoseAt(l, n+5*loop) {
				t.Fatalf("%v: the pose at %d is not the pose a loop (%d) later", tier, n, loop)
			}
		}
	}
	prev := 0
	for _, tier := range tiers[:4] {
		got := poseCount(earnedLook(tier), coinShows[tier].loop)
		if got <= prev {
			t.Errorf("%v has %d poses in a loop, no more than the tier below (%d)", tier, got, prev)
		}
		prev = got
	}
	if got := poseCount(earnedLook(config.BadgeLegendary), coinShows[config.BadgeLegendary].loop); got < hueSteps {
		t.Errorf("a legendary coin has %d poses, want at least %d", got, hueSteps)
	}
}

// A picked medal turns for turnSteps poses, then its show begins.
func TestPickedPoseAt(t *testing.T) {
	l := earnedLook(config.BadgeGold)
	for since := 0; since < turnSteps; since++ {
		if got, want := pickedPoseAt(l, since), (coinPose{turn: int8(since + 1)}); got != want {
			t.Errorf("%d poses after the pick the pose is %+v, want %+v", since, got, want)
		}
	}
	if got, want := pickedPoseAt(l, turnSteps), coinPoseAt(l, 0); got != want {
		t.Errorf("when the turn ends the pose is %+v, want the show's first %+v", got, want)
	}
	if got := pickedPoseAt(l, atRest); got != (coinPose{}) {
		t.Errorf("at rest the pose is %+v", got)
	}
}

// A pose, a tier and a hue each change the picture, and nothing else does.
func TestCoinPictureByPose(t *testing.T) {
	const w, h = 180, 189
	silver, gold, plat, legend := earnedLook(config.BadgeSilver), earnedLook(config.BadgeGold), earnedLook(config.BadgePlatinum), earnedLook(config.BadgeLegendary)
	rest := coinPicture(silver, w, h, coinPose{})
	if !bytes.Equal(rest.Pix, coinPicture(silver, w, h, coinPose{}).Pix) {
		t.Error("the same look and pose drew two different pictures")
	}
	for what, pose := range map[string]coinPose{
		"a glint": {glint: 4}, "a sparkle": {spark: [3]int8{3, 0, 0}}, "a turn": {turn: 2},
	} {
		if bytes.Equal(rest.Pix, coinPicture(silver, w, h, pose).Pix) {
			t.Errorf("%s did not change the picture", what)
		}
	}
	if bytes.Equal(coinPicture(gold, w, h, coinPose{}).Pix, coinPicture(gold, w, h, coinPose{rim: 5}).Pix) {
		t.Error("the light on a gold rim did not change the picture")
	}
	if bytes.Equal(rest.Pix, coinPicture(plat, w, h, coinPose{}).Pix) {
		t.Error("platinum at rest is drawn as silver")
	}
	if bytes.Equal(coinPicture(legend, w, h, coinPose{hue: 0}).Pix, coinPicture(legend, w, h, coinPose{hue: 6}).Pix) {
		t.Error("a legendary coin looks the same at two hues")
	}
}

// Every sign is a 13 by 13 grid of # and ., no two are alike, the special
// badges have theirs, and a badge shows the sign it is due.
func TestMedalSigns(t *testing.T) {
	keys := map[string]string{}
	for name, icon := range medalIcons {
		if len(icon) != 13 {
			t.Errorf("the %s sign has %d rows", name, len(icon))
		}
		for _, row := range icon {
			if len([]rune(row)) != 13 || strings.Trim(row, "#.") != "" {
				t.Errorf("the %s sign has the row %q", name, row)
			}
		}
		flat := strings.Join(icon, "/")
		if other, dup := keys[flat]; dup {
			t.Errorf("the signs %s and %s are the same", other, name)
		}
		keys[flat] = name
	}
	for _, name := range []string{specialCookieJar, specialLedger, specialSource, specialChains, specialBoxes, specialMuseum, specialWonders, specialRogues, specialEndings, specialUndying, specialTranscend} {
		if medalIcons[name] == nil {
			t.Errorf("no sign %s", name)
		}
		if got := lookOf(medal{tier: config.BadgeGold, state: medalEarned, special: name}, "age", coinArt).sign; got != name {
			t.Errorf("an earned %s badge carries the sign %q", name, got)
		}
		if got := lookOf(medal{tier: config.BadgeGold, state: medalLocked, special: name}, "age", coinArt).sign; got != "keyhole" {
			t.Errorf("a locked %s badge carries the sign %q, want the keyhole", name, got)
		}
		if got := lookOf(medal{tier: config.BadgeGold, state: medalHidden, special: name}, "age", coinArt).sign; got != "question" {
			t.Errorf("a hidden %s badge carries the sign %q, want the question mark", name, got)
		}
	}
}

// picturePanel is a Badges panel laid out for pictures, with a clock the test
// moves.
func picturePanel(t *testing.T) (p *badgePanel, paint func() *kit.Buffer, tick func(n int)) {
	t.Helper()
	views, sum := caseFixture()
	st := game.GameState{AccountStats: &game.AccountStatsView{Badges: views, BadgeSummary: sum}}
	clock := time.Unix(1_700_000_000, 0)
	p = newBadgePanel()
	p.now, p.start = func() time.Time { return clock }, clock
	was := pictureCell
	pictureCell = func() (int, int) { return 10, 21 }
	t.Cleanup(func() { pictureCell = was })
	p.SetRect(0, 0, 120, 30)
	p.open(st)
	paint = func() *kit.Buffer {
		b := kit.NewBuffer(120, 30)
		p.Layout(image.Rect(0, 0, 120, 30))
		p.Paint(kit.NewPainter(b))
		return b
	}
	tick = func(n int) { clock = clock.Add(time.Duration(n) * coinStep) }
	return p, paint, tick
}

// The coins move while the motion is on and an earned one is in view, stand
// still when it is off, and a coin not earned never moves.
func TestBadgePanelMotion(t *testing.T) {
	p, paint, tick := picturePanel(t)
	paint()
	earned, loose := -1, -1
	for i, c := range p.model.coins {
		if p.look(c.it, coinArt).earned() {
			earned = max(earned, i)
		} else if loose < 0 {
			loose = i
		}
	}
	if earned < 0 || loose < 0 {
		t.Fatalf("the fixture has no earned coin (%d) or none unearned (%d) in view", earned, loose)
	}
	first := paint()
	if !p.moving.Load() {
		t.Error("with the motion on and an earned coin in view the panel says nothing moves")
	}
	big := len(coinPics(first)) - 1
	tick(1)
	second := paint()
	if coinPics(first)[big].Img == coinPics(second)[big].Img {
		t.Error("the large medal's picture is the same a pose after it was picked")
	}
	for range 40 {
		tick(1)
		if coinPics(paint())[loose].Img != coinPics(first)[loose].Img {
			t.Fatal("a coin not earned changed its picture")
		}
	}
	p.motion = false
	still := paint()
	if p.moving.Load() {
		t.Error("with the motion off the panel says something moves")
	}
	tick(30)
	later := paint()
	for i := range coinPics(still) {
		if coinPics(still)[i].Img != coinPics(later)[i].Img {
			t.Fatalf("with the motion off picture %d changed with the clock", i)
		}
	}
}

// pressAt is a left press at the cell (x, y) of the panel's own cells.
func pressAt(p *badgePanel, x, y int) bool {
	b := p.Bounds().Min
	return p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: b.X + x, Y: b.Y + y, Clicks: 1})
}

// The key bar's keys and the close button are buttons: a press on one presses
// its key, the hand shows over it, and the bar fits the panel.
func TestBadgePanelKeyButtons(t *testing.T) {
	p, paint, _ := picturePanel(t)
	paint()
	var got []kit.Key
	p.buttons.press = func(k kit.Key) { got = append(got, k) }
	keys := p.model.keys.keys
	if len(keys) < 2 {
		t.Fatalf("the panel has %d key buttons", len(keys))
	}
	closed := false
	for _, k := range keys {
		got = nil
		at := k.r.Min
		if !pressAt(p, at.X, at.Y) || len(got) != 1 || got[0] != k.key {
			t.Errorf("a press at %v recorded %v, want %v", at, got, k.key)
		}
		if ptr := p.PointerAt(p.Bounds().Min.Add(at)); ptr != kit.PointerHand {
			t.Errorf("the pointer over the key at %v is %v, want the hand", at, ptr)
		}
		if k.r.Min.Y == 0 {
			closed = closed || k.key.Code == kit.KeyEsc
		}
	}
	if !closed {
		t.Error("the title bar has no close button that presses Esc")
	}
	for _, w := range []int{60, 80, 120, 160} {
		views, sum := caseFixture()
		for _, tier := range caseTiers {
			m, _, g := drawCase(views, sum, w, 30, caseView{tier: tier})
			for _, k := range m.keys.keys {
				if k.r.Min.X < 0 || k.r.Max.X > w {
					t.Errorf("%d wide, %s: a key button spans columns %d to %d", w, tier, k.r.Min.X, k.r.Max.X)
				}
			}
			if tier == mapmodel.TierASCII {
				for _, k := range m.keys.keys {
					for y := k.r.Min.Y; y < k.r.Max.Y; y++ {
						for x := k.r.Min.X; x < k.r.Max.X; x++ {
							if c := g.c[y*g.w+x]; c.r >= 0x80 {
								t.Fatalf("%d wide: the plain tier draws %q at %d,%d of the buttons", w, c.r, x, y)
							}
						}
					}
				}
			}
		}
	}
}

// coinPics is the pictures b shows but the key plates, which are a row high.
func coinPics(b *kit.Buffer) []kit.Pic {
	var out []kit.Pic
	for _, p := range b.Pics {
		if p.Rect.Dy() > 1 {
			out = append(out, p)
		}
	}
	return out
}
