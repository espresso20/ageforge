package ui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// festivalStage is the People overview over a town where festivals are open
// and the culture is enough, after mut.
func festivalStage(t *testing.T, mut func(*game.GameState)) *peopleStage {
	t.Helper()
	return stagePeople(t, tabOverview, func(st *game.GameState) {
		feats := map[string]game.FeatureState{}
		for k, v := range st.Features {
			feats[k] = v
		}
		f := feats[config.FeatureFestivals]
		f.Open = true
		feats[config.FeatureFestivals] = f
		st.Features = feats
		st.ActiveEvents = nil
		st.Festival = game.FestivalStatus{Cost: 5000, Culture: 9000, BuffPercent: 0.2, BuffTicks: 150, CooldownTicks: 300, Ready: true}
		if mut != nil {
			mut(st)
		}
	})
}

func TestFestivalButtonAsksFirstAndRunsOnlyOnTheYes(t *testing.T) {
	s := festivalStage(t, nil)
	k := s.key(t, festivalCmd)
	if k.off || k.ask == "" || k.yes == "" || k.kind != keyPrimary {
		t.Fatalf("the festival button: off %v, asks %q, yes %q, kind %v", k.off, k.ask, k.yes, k.kind)
	}
	if !strings.Contains(k.ask, "5K") {
		t.Errorf("the question %q does not name the cost", k.ask)
	}
	s.click(k)
	if len(s.ran) != 0 || len(s.asked) != 1 || !strings.HasSuffix(s.asked[0], "|"+festivalCmd) {
		t.Errorf("the click ran %q and asked %q, want a question and no command", s.ran, s.asked)
	}
}

func TestFestivalButtonIsOffWhenItCannotRun(t *testing.T) {
	cases := map[string]func(*game.GameState){
		"running": func(st *game.GameState) {
			st.ActiveEvents = []game.ActiveEventState{{Key: festivalEvent, TicksLeft: 60}}
			st.Festival.Ready = false
			st.Festival.CooldownLeft = 200
		},
		"cooling down": func(st *game.GameState) { st.Festival.Ready, st.Festival.CooldownLeft = false, 100 },
		"short":        func(st *game.GameState) { st.Festival.Culture = 100 },
	}
	for name, mut := range cases {
		s := festivalStage(t, mut)
		k := s.key(t, festivalCmd)
		if !k.off {
			t.Errorf("%s: the button is on", name)
		}
		s.click(k)
		if len(s.ran) != 0 || len(s.asked) != 0 {
			t.Errorf("%s: a click ran %q and asked %q", name, s.ran, s.asked)
		}
	}
}

func TestFestivalCardShowsTheCapNoteBeforeTheSpend(t *testing.T) {
	const note = "capped: no effect now"
	if s := festivalStage(t, nil); strings.Contains(s.p.grid.String(), "capped") {
		t.Fatal("the card shows a cap note when the whole buff counts")
	}
	s := festivalStage(t, func(st *game.GameState) { st.Festival.CapNote = note })
	if !strings.Contains(s.p.grid.String(), note) {
		t.Errorf("the card does not show the cap note %q", note)
	}
	if k := s.key(t, festivalCmd); k.off {
		t.Error("the button is off although the festival can be held")
	}
}

func TestFestivalCardIsNotDrawnWhileLocked(t *testing.T) {
	s := festivalStage(t, func(st *game.GameState) {
		feats := map[string]game.FeatureState{}
		for k, v := range st.Features {
			feats[k] = v
		}
		f := feats[config.FeatureFestivals]
		f.Open = false
		feats[config.FeatureFestivals] = f
		st.Features = feats
	})
	for _, k := range s.p.grid.keyset().keys {
		if k.cmd == festivalCmd {
			t.Fatal("a button exists while festivals are locked")
		}
	}
	if strings.Contains(s.p.grid.String(), txtFestival.String()) {
		t.Error("the card is drawn while festivals are locked")
	}
	if on := festivalStage(t, nil); !strings.Contains(on.p.grid.String(), txtFestival.String()) {
		t.Error("the card is not drawn while festivals are open")
	}
}

func TestFestivalChipIsDrawnOnlyWhileOneRuns(t *testing.T) {
	st := game.GameState{Festival: game.FestivalStatus{BuffPercent: 0.2}}
	if got := statusLine(st, "").Festival; got != "" {
		t.Errorf("the chip %q is drawn with no festival running", got)
	}
	st.ActiveEvents = []game.ActiveEventState{{Key: festivalEvent, TicksLeft: 60}}
	got := statusLine(st, "").Festival
	if !strings.Contains(got, "+20%") || !strings.Contains(got, "~") {
		t.Errorf("the chip %q lacks the rise or the time left", got)
	}
	v := statusLine(st, "")
	v.Age, v.People = "Age", headFact{"People 1/2", "1/2"}
	if !strings.Contains(renderHeader(v, 120, 2).String(), got) {
		t.Error("the header does not show the chip")
	}
}

func TestFestivalButtonHoldsOneFestival(t *testing.T) {
	_, _, eng := stagedDashboard(t, "bronze_age")
	eng.GrantTechsForTest("drama")
	eng.Resources.Add("culture", 1e12)
	st := eng.GetState()
	if !st.Festival.Ready || st.Festival.Culture < st.Festival.Cost {
		t.Fatalf("the staged town cannot hold a festival: %+v", st.Festival)
	}
	if got := cmdFestival([]string{"confirm", "yes"}, eng); got.Type == "error" {
		t.Fatalf("the command failed: %s", got.Message)
	}
	st = eng.GetState()
	if _, running := festivalRunning(st); !running || st.Festival.Ready {
		t.Errorf("after the command: running %v, status %+v", running, st.Festival)
	}
}

// The festival's question has a primary yes that Enter presses and a way out
// on Esc; the question for a command that cannot be undone keeps the danger
// button with the way out first.
func TestQuestionIsDangerOnlyForADangerousLine(t *testing.T) {
	keepTheme(t)
	d, pages, eng := stagedDashboard(t, "bronze_age")
	eng.GrantTechsForTest("drama")
	eng.Resources.Add("culture", 1e12)
	ask := func(yes, line string) *nativeDialog {
		t.Helper()
		d.askLine("Sure?", yes, line)
		name, item := pages.GetFrontPage()
		dlg, _ := item.(*nativeDialog)
		if name != "ask_line" || dlg == nil {
			t.Fatalf("%q opened no question (front page %q)", line, name)
		}
		d.app.SetFocus(dlg)
		return dlg
	}
	look := func(dlg *nativeDialog) ([]keyKind, []string) {
		var kinds []keyKind
		var keys []string
		for _, b := range dlg.buttons {
			kinds, keys = append(kinds, b.act), append(keys, b.key)
		}
		return kinds, keys
	}
	held := func() bool { _, on := festivalRunning(eng.GetState()); return on }

	dlg := ask("Sell", "sell hut")
	if kinds, keys := look(dlg); !reflect.DeepEqual(kinds, []keyKind{keyPlain, keyDanger}) || !reflect.DeepEqual(keys, []string{"Esc", ""}) {
		t.Errorf("a Dangerous line: kinds %v keys %v", kinds, keys)
	}
	keyTo(dlg, tk(kit.KeyEsc, 0, 0))

	dlg = ask("Hold", festivalCmd)
	if kinds, keys := look(dlg); !reflect.DeepEqual(kinds, []keyKind{keyPrimary, keyPlain}) || !reflect.DeepEqual(keys, []string{"Enter", "Esc"}) {
		t.Errorf("the festival: kinds %v keys %v", kinds, keys)
	}
	keyTo(dlg, tk(kit.KeyEsc, 0, 0))
	if held() {
		t.Fatal("Esc on the question held the festival")
	}
	keyTo(ask("Hold", festivalCmd), tk(kit.KeyEnter, 0, 0))
	if !held() {
		t.Error("Enter on the question did not hold the festival")
	}
}
