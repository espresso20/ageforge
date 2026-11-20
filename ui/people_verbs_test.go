package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// peopleStage is the People panel on a tab, laid out at 100x30 over the
// snapshot after mut, with what its buttons run and ask collected.
type peopleStage struct {
	p     *peoplePanel
	ran   []string
	asked []string
}

func stagePeople(t *testing.T, tab int, mut func(*game.GameState)) *peopleStage {
	t.Helper()
	st := peopleTown(t)
	// Copy the map the town shares with its engine before a test edits it.
	bs := make(map[string]game.BuildingState, len(st.Buildings))
	for k, v := range st.Buildings {
		bs[k] = v
	}
	st.Buildings = bs
	if mut != nil {
		mut(&st)
	}
	s := &peopleStage{p: newPeoplePanel()}
	s.p.open(st)
	s.p.view.tab = tab
	s.p.buttons.press = func(k kit.Key) { s.p.take(k) }
	s.p.buttons.run = func(c string) { s.ran = append(s.ran, c) }
	s.p.buttons.confirm = func(q, _, c string) { s.asked = append(s.asked, q+"|"+c) }
	s.p.Layout(s.rect())
	return s
}

// rect is the panel's rectangle: the Overview takes the large layout.
func (s *peopleStage) rect() image.Rectangle {
	if s.p.view.tab == tabOverview {
		return image.Rect(0, 0, peopleBigW, peopleBigH)
	}
	return image.Rect(0, 0, 157, 43)
}

// key is the button that carries cmd.
func (s *peopleStage) key(t *testing.T, cmd string) xkey {
	t.Helper()
	for _, k := range s.p.grid.keyset().keys {
		if k.cmd == cmd {
			return k
		}
	}
	t.Fatalf("no button carries %q", cmd)
	return xkey{}
}

// click presses and releases the left button on the middle of k.
func (s *peopleStage) click(k xkey) {
	pt := k.r.Min.Add(image.Pt(k.r.Dx()/2, 0))
	for _, kind := range []kit.MouseKind{kit.MousePress, kit.MouseRelease} {
		s.p.Mouse(kit.Mouse{Kind: kind, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
	}
	s.p.Layout(s.rect())
}

// setBuilding edits one building of the snapshot.
func setBuilding(st *game.GameState, key string, f func(*game.BuildingState)) {
	b := st.Buildings[key]
	f(&b)
	st.Buildings[key] = b
}

func TestPeopleWorkStepperRunsAssignAndUnassignOnce(t *testing.T) {
	s := stagePeople(t, tabWork, func(st *game.GameState) {
		st.Workers.TotalIdle = 5
		setBuilding(st, "farm", func(b *game.BuildingState) { b.WorkersAssigned = 3 })
	})
	minus, plus := s.key(t, "unassign farm"), s.key(t, "assign farm")
	if minus.off || plus.off {
		t.Fatalf("minus off %v, plus off %v, with workers in and idle people", minus.off, plus.off)
	}
	s.click(minus)
	s.click(plus)
	if len(s.ran) != 2 || s.ran[0] != "unassign farm" || s.ran[1] != "assign farm" {
		t.Errorf("the stepper ran %q", s.ran)
	}
}

func TestPeopleWorkStepperIsOffAtTheEnds(t *testing.T) {
	s := stagePeople(t, tabWork, func(st *game.GameState) {
		st.Workers.TotalIdle = 0
		setBuilding(st, "farm", func(b *game.BuildingState) { b.WorkersAssigned = 0 })
	})
	minus, plus := s.key(t, "unassign farm"), s.key(t, "assign farm")
	if !minus.off {
		t.Error("minus is on with no workers in the building")
	}
	if !plus.off {
		t.Error("plus is on with nobody idle")
	}
	s.click(minus)
	s.click(plus)
	if len(s.ran) != 0 {
		t.Errorf("an off button ran %q", s.ran)
	}
}

func TestPeopleDismissAsksFirstAndRunsOnlyOnTheDangerButton(t *testing.T) {
	s := stagePeople(t, tabWork, func(st *game.GameState) {
		setBuilding(st, "farm", func(b *game.BuildingState) { b.WorkersAssigned = 3 })
	})
	k := s.key(t, "dismiss farm")
	if k.off || k.ask == "" || k.yes == "" {
		t.Fatalf("the dismiss chip: off %v, asks %q, yes %q", k.off, k.ask, k.yes)
	}
	s.click(k)
	if len(s.ran) != 0 || len(s.asked) != 1 {
		t.Errorf("the click ran %q and asked %q, want a question and no command", s.ran, s.asked)
	}
	// With nobody to dismiss the chip is off and asks nothing.
	s = stagePeople(t, tabWork, func(st *game.GameState) {
		setBuilding(st, "farm", func(b *game.BuildingState) { b.WorkersAssigned = 0 })
	})
	s.click(s.key(t, "dismiss farm"))
	if len(s.ran)+len(s.asked) != 0 {
		t.Errorf("an off dismiss ran %q and asked %q", s.ran, s.asked)
	}
}

// The question itself, through the dashboard: the keyboard starts on Cancel,
// Enter there and Esc cancel, and the danger button runs the line once.
func TestPeopleDismissQuestionKeysAndButtons(t *testing.T) {
	keepTheme(t)
	for _, step := range []struct {
		name    string
		answer  func(d *nativeDialog)
		wantRun int
	}{
		{"Enter on the opening button", func(d *nativeDialog) { keyTo(d, tk(kit.KeyEnter, 0, 0)) }, 0},
		{"Esc", func(d *nativeDialog) { keyTo(d, tk(kit.KeyEsc, 0, 0)) }, 0},
		{"the danger button", func(d *nativeDialog) {
			keyTo(d, tk(kit.KeyTab, 0, 0))
			keyTo(d, tk(kit.KeyEnter, 0, 0))
		}, 1},
	} {
		t.Run(step.name, func(t *testing.T) {
			d, pages := newTestDashboard(nil)
			d.askLine("Dismiss a worker from Farm?", txtDismiss.String(), "dismiss farm")
			// A yes goes through pressLine, which the history records.
			_, item := pages.GetFrontPage()
			dlg, ok := item.(*nativeDialog)
			if !ok {
				t.Fatalf("the front page is %T, not the question", item)
			}
			d.app.SetFocus(dlg)
			if !dlg.buttons[0].HasFocus() || dlg.labels[0] != txtCancel.String() || dlg.labels[1] != txtDismiss.String() {
				t.Fatalf("the keyboard starts off Cancel: labels %q", dlg.labels)
			}
			step.answer(dlg)
			if got := len(d.cmdHistory); got != step.wantRun {
				t.Errorf("the answer ran the line %d times, want %d", got, step.wantRun)
			}
			if name, _ := pages.GetFrontPage(); name == "ask_line" {
				t.Error("the question is still up after the answer")
			}
		})
	}
}

func TestPeopleRecruitRunsTheChosenCountAndIsOffWithoutRoom(t *testing.T) {
	for _, c := range []struct {
		buy  buyChoice
		line string
	}{{buyOne, "recruit"}, {buyTen, "recruit 10"}, {buyMax, "recruit max"}} {
		s := stagePeople(t, tabWork, func(st *game.GameState) {
			st.Workers.TotalPop, st.Workers.MaxPop = 100, 200
		})
		s.p.view.buy = c.buy
		s.p.Layout(s.rect())
		s.click(s.key(t, c.line))
		if len(s.ran) != 1 || s.ran[0] != c.line {
			t.Errorf("choice %d ran %q, want %q once", c.buy, s.ran, c.line)
		}
	}
	// Room for five: one is fine, ten is off.
	s := stagePeople(t, tabWork, func(st *game.GameState) {
		st.Workers.TotalPop, st.Workers.MaxPop = 195, 200
	})
	if s.key(t, "recruit").off {
		t.Error("recruit one is off with room for five")
	}
	s.p.view.buy = buyTen
	s.p.Layout(s.rect())
	ten := s.key(t, "recruit 10")
	if !ten.off {
		t.Error("recruit ten is on with room for five")
	}
	s.click(ten)
	if len(s.ran) != 0 {
		t.Errorf("an off recruit ran %q", s.ran)
	}
	// No room at all: every choice is off.
	s = stagePeople(t, tabWork, func(st *game.GameState) {
		st.Workers.TotalPop, st.Workers.MaxPop = 200, 200
	})
	if !s.key(t, "recruit").off {
		t.Error("recruit is on with the homes full")
	}
}

func TestPeopleSplitStepsByTenAndStopsAtTheEnds(t *testing.T) {
	for _, c := range []struct {
		share, total      int
		minus, plus       string
		minusOff, plusOff bool
	}{
		{50, 18, "scholars 40", "scholars 60", false, false},
		{5, 18, "scholars 0", "scholars 15", false, false},
		{95, 18, "scholars 85", "scholars 100", false, false},
		{0, 18, "scholars 0", "scholars 10", true, false},
		{100, 18, "scholars 90", "scholars 100", false, true},
		{50, 0, "scholars 40", "scholars 60", true, true},
	} {
		s := stagePeople(t, tabScholars, func(st *game.GameState) {
			st.Scholars.Share, st.Scholars.Total = c.share, c.total
		})
		minus, plus := s.key(t, c.minus), s.key(t, c.plus)
		if minus.off != c.minusOff || plus.off != c.plusOff {
			t.Errorf("share %d of %d: minus off %v, plus off %v", c.share, c.total, minus.off, plus.off)
		}
		s.click(minus)
		s.click(plus)
		var want []string
		if !c.minusOff {
			want = append(want, c.minus)
		}
		if !c.plusOff {
			want = append(want, c.plus)
		}
		if len(s.ran) != len(want) || (len(want) > 0 && s.ran[0] != want[0]) || (len(want) > 1 && s.ran[1] != want[1]) {
			t.Errorf("share %d of %d ran %q, want %q", c.share, c.total, s.ran, want)
		}
	}
}

func TestPeopleOverviewLinksChangeTheTab(t *testing.T) {
	s := stagePeople(t, tabOverview, nil)
	seen := map[int]bool{}
	for _, k := range s.p.grid.keyset().keys {
		if tab, ok := goOf(k.key); ok && k.cmd == "" {
			seen[tab] = true
		}
	}
	for _, tab := range []int{tabWork, tabArmy} {
		if !seen[tab] {
			t.Fatalf("the Overview has no link to tab %d", tab)
		}
		s = stagePeople(t, tabOverview, nil)
		for _, k := range s.p.grid.keyset().keys {
			if to, ok := goOf(k.key); ok && to == tab {
				s.click(k)
				break
			}
		}
		if s.p.view.tab != tab {
			t.Errorf("a click on the link to tab %d left the panel on tab %d", tab, s.p.view.tab)
		}
	}
}

func TestPeopleShareStepsByTenAndStopsAtTheEnds(t *testing.T) {
	probe := stagePeople(t, tabWork, nil)
	dom := game.ShareRows(probe.p.state)[0].Domain
	for _, c := range []struct {
		set               bool
		pct               float64
		minus, plus       string
		minusOff, plusOff bool
	}{
		{true, 50, "roster " + dom + " 40", "roster " + dom + " 60", false, false},
		{true, 0, "roster " + dom + " 0", "roster " + dom + " 10", true, false},
		{true, 100, "roster " + dom + " 90", "roster " + dom + " 100", false, true},
		{true, 5, "roster " + dom + " 0", "roster " + dom + " 15", false, false},
	} {
		s := stagePeople(t, tabWork, func(st *game.GameState) {
			st.Workers.Shares = map[string]float64{dom: c.pct}
		})
		minus, plus := s.key(t, c.minus), s.key(t, c.plus)
		if minus.off != c.minusOff || plus.off != c.plusOff {
			t.Errorf("share %v: minus off %v, plus off %v", c.pct, minus.off, plus.off)
		}
		s.click(minus)
		s.click(plus)
		want := 0
		for _, off := range []bool{c.minusOff, c.plusOff} {
			if !off {
				want++
			}
		}
		if len(s.ran) != want {
			t.Errorf("share %v ran %q, want %d lines", c.pct, s.ran, want)
		}
	}
}

func TestPeopleAutoChipRunsAutoAndIsAccentWhileAuto(t *testing.T) {
	probe := stagePeople(t, tabWork, nil)
	dom := game.ShareRows(probe.p.state)[0].Domain
	s := stagePeople(t, tabWork, nil)
	k := s.key(t, "roster "+dom+" auto")
	if k.kind != keyAccent {
		t.Errorf("auto chip kind %v while on auto", k.kind)
	}
	s.click(k)
	if len(s.ran) != 1 || s.ran[0] != "roster "+dom+" auto" {
		t.Errorf("ran %q", s.ran)
	}
	set := stagePeople(t, tabWork, func(st *game.GameState) { st.Workers.Shares = map[string]float64{dom: 40} })
	if set.key(t, "roster "+dom+" auto").kind == keyAccent {
		t.Errorf("auto chip is accent on a set share")
	}
}

func TestPeopleWorkRowsStayPutWhateverIsPicked(t *testing.T) {
	var base []xkey
	var names [][]int
	for sub := 0; sub < 3; sub++ {
		s := stagePeople(t, tabWork, nil)
		bs := peopleBody{st: s.p.state}.areaList(game.ShareRows(s.p.state)[s.p.view.row].Domain)
		if len(bs) < 3 {
			t.Skipf("only %d buildings in the first area", len(bs))
		}
		s.p.view.sub = sub
		s.p.Layout(s.rect())
		var ys []int
		var steps []xkey
		var dismiss *xkey
		for _, k := range s.p.grid.keyset().keys {
			if strings.HasPrefix(k.cmd, "assign ") || strings.HasPrefix(k.cmd, "unassign ") {
				steps = append(steps, k)
				ys = append(ys, k.r.Min.Y)
			}
			if strings.HasPrefix(k.cmd, "dismiss ") {
				kk := k
				dismiss = &kk
			}
		}
		names = append(names, ys)
		if sub == 0 {
			base = steps
		}
		if dismiss == nil || dismiss.r.Min.Y != ys[2*sub] || dismiss.r.Max.X+1 != steps[2*sub].r.Min.X {
			t.Errorf("pick %d: dismiss %v not on the picked row (%v)", sub, dismiss, ys)
		}
		if len(steps) != len(base) {
			t.Fatalf("pick %d: %d steppers, want %d", sub, len(steps), len(base))
		}
	}
	for i := 1; i < len(names); i++ {
		for j := range names[0] {
			if names[i][j] != names[0][j] {
				t.Errorf("rows moved with the pick: %v vs %v", names[0], names[i])
			}
		}
	}
}
