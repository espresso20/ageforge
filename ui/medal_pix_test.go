package ui

import (
	"bytes"
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
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
