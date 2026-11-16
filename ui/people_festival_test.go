package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
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
