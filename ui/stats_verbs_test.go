package ui

import (
	"image"
	"testing"
	"time"

	"github.com/espresso20/ageforge/window/kit"
)

// stagePrestige is a Stats page with the Prestige button in view, open or
// not, and what the button ran and asked.
func stagePrestige(t *testing.T, can bool) (p *statsPanel, k xkey, ran []string, asked *[][3]string) {
	t.Helper()
	st := statsFixture(t, 3, 4, time.Hour)
	st.Prestige.CanPrestige = can
	p = laidStats(t, st, 160, 60, true)
	a := [][3]string{}
	asked = &a
	p.pageBtn.press = func(kit.Key) {}
	p.pageBtn.run = func(c string) { ran = append(ran, c) }
	p.pageBtn.confirm = func(q, yes, line string) { a = append(a, [3]string{q, yes, line}) }
	for _, c := range p.page.g.keyset().keys {
		if c.cmd == prestigeCmd {
			k = c
		}
	}
	if k.cmd == "" {
		t.Fatal("the page has no Prestige button")
	}
	p.scroll = k.r.Min.Y
	p.scrollBy(0)
	return p, k, ran, asked
}

func clickPage(p *statsPanel, k xkey) {
	pt := p.pageOrg().Add(k.r.Min).Add(image.Pt(k.r.Dx()/2, k.r.Dy()/2))
	for _, kind := range []kit.MouseKind{kit.MousePress, kit.MouseRelease} {
		p.Mouse(kit.Mouse{Kind: kind, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
	}
}

func TestPrestigeButtonAsksFirstAndRunsNothingUntilAnswered(t *testing.T) {
	p, k, ran, asked := stagePrestige(t, true)
	if k.off {
		t.Fatal("Prestige is off with a prestige open")
	}
	clickPage(p, k)
	if len(ran) != 0 || len(*asked) != 1 {
		t.Fatalf("a click ran %q and asked %d times, want a question only", ran, len(*asked))
	}
	if got := (*asked)[0]; got[0] != txtPrestigeAsk.String() || got[1] != txtPrestigeVerb.String() || got[2] != prestigeCmd {
		t.Errorf("asked %q", got)
	}
}

func TestPrestigeButtonIsOffWhenNoPrestigeIsOpen(t *testing.T) {
	p, k, ran, asked := stagePrestige(t, false)
	if !k.off {
		t.Fatal("Prestige is on with no prestige open")
	}
	clickPage(p, k)
	if len(ran)+len(*asked) != 0 {
		t.Errorf("an off Prestige ran %q and asked %d times", ran, len(*asked))
	}
}

// Through the dashboard's question: Cancel runs nothing, the danger button
// runs the line once.
func TestPrestigeQuestionCancelAndConfirm(t *testing.T) {
	keepTheme(t)
	for _, c := range []struct {
		name    string
		answer  func(d *nativeDialog)
		wantRun int
	}{
		{"cancel", func(d *nativeDialog) { keyTo(d, tk(kit.KeyEsc, 0, 0)) }, 0},
		{"confirm", func(d *nativeDialog) {
			keyTo(d, tk(kit.KeyTab, 0, 0))
			keyTo(d, tk(kit.KeyEnter, 0, 0))
		}, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			p, k, _, _ := stagePrestige(t, true)
			d, pages := newTestDashboard(nil)
			p.pageBtn.confirm = d.askLine
			clickPage(p, k)
			_, item := pages.GetFrontPage()
			dlg, ok := item.(*nativeDialog)
			if !ok {
				t.Fatalf("the front page is %T, not the question", item)
			}
			d.app.SetFocus(dlg)
			if dlg.labels[1] != txtPrestigeVerb.String() {
				t.Fatalf("labels %q", dlg.labels)
			}
			c.answer(dlg)
			if got := len(d.cmdHistory); got != c.wantRun {
				t.Errorf("ran %d times, want %d", got, c.wantRun)
			}
		})
	}
}
