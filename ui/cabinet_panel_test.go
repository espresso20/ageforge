package ui

import (
	"image"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/cabinet"
	"github.com/espresso20/ageforge/window/kit"
)

// cabinetRig is the cabinet of a fresh game on a fresh account, laid out at
// 160x44 cells of 12x24 pixels, with a way to paint it.
func cabinetRig(t *testing.T) (*Dashboard, *cabinetPanel, func() *kit.Buffer) {
	t.Helper()
	d, _, _ := badgeDashboard(t)
	was := pictureCell
	pictureCell = func() (int, int) { return 12, 24 }
	t.Cleanup(func() { pictureCell = was })
	p := d.cabinet
	p.update(d.engine.GetState())
	p.Layout(image.Rect(0, 0, 160, 44))
	return d, p, func() *kit.Buffer {
		b := kit.NewBuffer(160, 44)
		p.Paint(kit.NewPainter(b))
		return b
	}
}

// plateText is the words on the cells of one plate.
func plateText(b *kit.Buffer, r image.Rectangle) string {
	var s strings.Builder
	for x := r.Min.X; x < r.Max.X; x++ {
		s.WriteRune(b.At(x, r.Min.Y).R)
	}
	return strings.TrimSpace(s.String())
}

// The commands open the cabinet with the command bar keeping the keyboard,
// and an age or a drawer's name selects its drawer; a name the player cannot
// see yet finds nothing.
func TestBadgesCommandOpensTheCabinet(t *testing.T) {
	d, p, _ := cabinetRig(t)
	for _, cmd := range []string{"badges", "achievements"} {
		d.runForTest(cmd)
		if got := d.overlayMgr.ActiveName(); got != "badges" {
			t.Fatalf("%s opened %q", cmd, got)
		}
		d.overlayMgr.Hide()
	}
	for cmd, want := range map[string]int{"badges stone age": 1, "badges Stone": 1, "badges themes": 24, "badges ledgers": 22, "badges curios": 23} {
		d.runForTest("badges themes")
		d.runForTest(cmd)
		if d.overlayMgr.ActiveName() != "badges" || p.sel != want {
			t.Errorf("%s: panel %q, drawer %d, want %d", cmd, d.overlayMgr.ActiveName(), p.sel, want)
		}
	}
	d.overlayMgr.Hide()
	for _, cmd := range []string{"badges modern", "badges zzz", "badges next"} {
		if res := HandleCommand(cmd, d.engine); res.OverlayName != "" || res.Type != "error" {
			t.Errorf("%s: %+v", cmd, res)
		}
	}
}

// The arrows move across the grid without wrapping, Home and End go to the
// ends, and the keys that print are left to the prompt.
func TestCabinetKeys(t *testing.T) {
	_, p, _ := cabinetRig(t)
	for _, c := range []struct {
		key  kit.KeyCode
		want int
	}{
		{kit.KeyLeft, 0}, {kit.KeyUp, 0}, {kit.KeyRight, 1}, {kit.KeyDown, 6}, {kit.KeyEnd, 24},
		{kit.KeyRight, 24}, {kit.KeyDown, 24}, {kit.KeyLeft, 23}, {kit.KeyUp, 18}, {kit.KeyHome, 0},
	} {
		if !p.routeKey(&kit.Key{Code: c.key}, "") || p.sel != c.want {
			t.Fatalf("key %v: drawer %d, want %d", c.key, p.sel, c.want)
		}
	}
	if p.routeKey(&kit.Key{Code: kit.KeyRune, Rune: 'x'}, "") || p.routeKey(&kit.Key{Code: kit.KeyEnter}, "") {
		t.Error("the cabinet took a key that is not its own")
	}
}

// A click on a plate selects its drawer, the hand shows over a drawer, the
// wheel steps one drawer, and an age out of sight keeps its name off its plate.
func TestCabinetMouseAndPlates(t *testing.T) {
	_, p, paint := cabinetRig(t)
	b := paint()
	lay, _ := p.plan()
	at := lay.Plates[7].Min
	if !p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: at.X, Y: at.Y, Clicks: 1}) || p.sel != 7 {
		t.Errorf("a click on drawer 7's plate selected %d", p.sel)
	}
	if p.PointerAt(at) != kit.PointerHand || p.PointerAt(image.Pt(0, 0)) == kit.PointerHand {
		t.Error("the hand does not show over a drawer only")
	}
	p.Mouse(kit.Mouse{Kind: kit.MouseWheel, Wheel: 1})
	p.Mouse(kit.Mouse{Kind: kit.MouseWheel, Wheel: 1})
	p.Mouse(kit.Mouse{Kind: kit.MouseWheel, Wheel: -1})
	if p.sel != 8 {
		t.Errorf("the wheel left the selection on %d, want 8", p.sel)
	}
	ages := config.Ages()
	if got := plateText(b, lay.Plates[0]); got != ages[0].Name {
		t.Errorf("the first plate says %q, want %q", got, ages[0].Name)
	}
	last := len(ages) - 1
	if got := plateText(b, lay.Plates[last]); strings.Contains(got, ages[last].Name) || got == "" {
		t.Errorf("the plate of an age out of sight says %q", got)
	}
	if got := plateText(b, lay.Plates[24]); got == "" {
		t.Error("the last drawer has no name")
	}
}

// What a player holds shows in the totals; the picture is drawn again for a
// change in what the drawers hold or in the size, and never for a move.
func TestCabinetPictureIsKept(t *testing.T) {
	d, p, paint := cabinetRig(t)
	paint()
	n := p.renders
	for _, k := range []kit.KeyCode{kit.KeyRight, kit.KeyDown, kit.KeyEnd, kit.KeyHome} {
		p.routeKey(&kit.Key{Code: k}, "")
		paint()
	}
	d.refresh()
	paint()
	if p.renders != n || n != 1 {
		t.Fatalf("the front was drawn %d times after %d, for moves only", p.renders, n)
	}
	acct := d.engine.Account()
	acct.StrikeCoin(config.CoinAge, config.Ages()[0].Key, 10, 1, "run", time.Unix(1700000000, 0), false)
	p.update(d.engine.GetState())
	if b := paint(); p.renders != n+1 || !strings.Contains(plateText(b, mustPlan(t, p).Title), "Pieces 1 of 132") {
		t.Errorf("a coin held: %d draws, title %q", p.renders, plateText(b, mustPlan(t, p).Title))
	}
	p.Layout(image.Rect(0, 0, 150, 42))
	paint()
	if p.renders != n+2 {
		t.Errorf("a new size drew the front %d times", p.renders-n-1)
	}
	p.Layout(image.Rect(0, 0, 40, 10))
	if b := paint(); !strings.Contains(plateText(b, image.Rect(0, 0, 40, 1)), "too small") {
		t.Error("a window too small for the cabinet says nothing")
	}
}

func mustPlan(t *testing.T, p *cabinetPanel) cabinet.Layout {
	t.Helper()
	lay, ok := p.plan()
	if !ok {
		t.Fatal("the cabinet does not fit")
	}
	return lay
}
