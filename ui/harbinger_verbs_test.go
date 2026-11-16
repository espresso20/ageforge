package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// harbStage is the Harbinger panel with its buttons laid, and what they ran
// and asked collected.
type harbStage struct {
	d     *Dashboard
	eng   *game.GameEngine
	pages *nativePages
	ran   []string
	asked [][3]string
}

// stageHarbinger summons a harbinger and pays for both answers.
func stageHarbinger(t *testing.T, pay bool) *harbStage {
	t.Helper()
	d, eng, pages := harbDashboard(t, "classical_age")
	if pay {
		h := eng.GetState().Harbinger
		for _, c := range []map[string]float64{h.AppeaseCost, h.BraceCost} {
			for res, need := range c {
				eng.SetStockForTest(res, need*10)
			}
		}
	}
	s := &harbStage{d: d, eng: eng, pages: pages}
	pr := &d.harbPanel.view.pr
	pr.press = func(kit.Key) {}
	pr.run = func(c string) { s.ran = append(s.ran, c) }
	pr.confirm = func(q, yes, line string) { s.asked = append(s.asked, [3]string{q, yes, line}) }
	rect := image.Rect(2, 1, 100, 200)
	d.harbPanel.provider(eng.GetState(), 100)
	d.harbPanel.view.grid, d.harbPanel.view.org = d.harbPanel.view.layer(rect, 0), rect.Min
	return s
}

func (s *harbStage) button(t *testing.T, cmd string) xkey {
	t.Helper()
	if g := s.d.harbPanel.view.grid; g != nil {
		for _, k := range g.keyset().keys {
			if k.cmd == cmd {
				return k
			}
		}
	}
	t.Fatalf("no button carries %q", cmd)
	return xkey{}
}

func (s *harbStage) click(k xkey) {
	v := &s.d.harbPanel.view
	pt := v.org.Add(k.r.Min).Add(image.Pt(k.r.Dx()/2, k.r.Dy()/2))
	for _, kind := range []kit.MouseKind{kit.MousePress, kit.MouseRelease} {
		v.mouse(kit.Mouse{Kind: kind, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
	}
}

func TestHarbingerAppeaseAndBraceRunTheirCommandOnce(t *testing.T) {
	s := stageHarbinger(t, true)
	for _, sub := range []string{"appease", "brace"} {
		k := s.button(t, harbingerCmd+sub)
		if k.off || k.kind != keyPlain {
			t.Fatalf("%s is %v off=%v, want plain and on", sub, k.kind, k.off)
		}
		s.ran = nil
		s.click(k)
		if len(s.ran) != 1 || s.ran[0] != k.cmd {
			t.Errorf("%s ran %q, want once", sub, s.ran)
		}
	}
	if len(s.asked) != 0 {
		t.Errorf("a paid answer asked %q", s.asked)
	}
}

func TestHarbingerAnswersAreOffWhenUnaffordable(t *testing.T) {
	s := stageHarbinger(t, false)
	for _, sub := range []string{"appease", "brace"} {
		k := s.button(t, harbingerCmd+sub)
		if !k.off {
			t.Errorf("%s is on with nothing to pay with", sub)
		}
		s.click(k)
	}
	if len(s.ran)+len(s.asked) != 0 {
		t.Errorf("an off button ran %q and asked %q", s.ran, s.asked)
	}
}

func TestHarbingerInviteAsksFirst(t *testing.T) {
	s := stageHarbinger(t, false)
	k := s.button(t, harbingerCmd+"invite")
	if k.kind != keyDanger || k.off {
		t.Fatalf("Invite is %v off=%v, want a danger button that is on", k.kind, k.off)
	}
	s.click(k)
	if len(s.ran) != 0 || len(s.asked) != 1 {
		t.Fatalf("a click ran %q and asked %d times", s.ran, len(s.asked))
	}
	if q := s.asked[0]; !strings.Contains(q[0], s.eng.GetState().Harbinger.Name) || q[1] != txtVerbInvite.String() || q[2] != k.cmd {
		t.Errorf("asked %q", q)
	}
}

func TestHarbingerTextStillShowsTheKeys(t *testing.T) {
	s := stageHarbinger(t, true)
	if txt := untag(s.d.harbPanel.view.text); !strings.Contains(txt, "Invite (twice)") {
		t.Errorf("the key bar went:\n%s", txt)
	}
}

// With the mouse alone, through the dashboard: appease, brace, then invite
// answered by the danger button.
func TestHarbingerAnswersByMouseAlone(t *testing.T) {
	keepTheme(t)
	s := stageHarbinger(t, true)
	d, eng := s.d, s.eng
	d.harbPanel.view.pr.wire(d.overlayMgr.app)
	d.overlayMgr.screenW = 100
	show := func() {
		d.overlayMgr.Show("harbinger", eng.GetState())
		drawPrimitive(t, s.pages, 100, 60)
	}
	press := func(cmd string) {
		t.Helper()
		var k *xkey
		for i, kk := range d.harbPanel.view.grid.keyset().keys {
			if kk.cmd == cmd {
				k = &d.harbPanel.view.grid.keyset().keys[i]
			}
		}
		if k == nil {
			t.Fatalf("no button carries %q", cmd)
		}
		tv := d.overlayMgr.entries["harbinger"].tv
		pt := tv.view.Bounds().Min.Add(k.r.Min).Add(image.Pt(k.r.Dx()/2, k.r.Dy()/2))
		for _, kind := range []kit.MouseKind{kit.MousePress, kit.MouseRelease} {
			tv.Mouse(kit.Mouse{Kind: kind, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
		}
	}
	show()
	h := eng.GetState().Harbinger
	press(harbingerCmd + "appease")
	if got := eng.GetState().Harbinger.AppeaseLevel; got != h.AppeaseLevel+1 {
		t.Fatalf("appease level %d -> %d", h.AppeaseLevel, got)
	}
	show()
	press(harbingerCmd + "brace")
	if got := eng.GetState().Harbinger.BraceLevel; got != h.BraceLevel+1 {
		t.Fatalf("brace level %d -> %d", h.BraceLevel, got)
	}
	show()
	press(harbingerCmd + "invite")
	if eng.GetState().Harbinger.Invited {
		t.Fatal("invite ran before the question was answered")
	}
	_, item := s.pages.GetFrontPage()
	dlg, ok := item.(*nativeDialog)
	if !ok {
		t.Fatalf("the front page is %T, not the question", item)
	}
	d.app.SetFocus(dlg)
	keyTo(dlg, tk(kit.KeyTab, 0, 0))
	keyTo(dlg, tk(kit.KeyEnter, 0, 0))
	if !eng.GetState().Harbinger.Invited {
		t.Error("the danger button did not invite")
	}
}
