package config

import "testing"

func TestMissionCatalog(t *testing.T) {
	missions := Missions()
	if len(missions) != 21 {
		t.Fatalf("got %d missions, want 21", len(missions))
	}
	want := map[string]struct{ min, max, success, yield, cost float64 }{
		MissionScout:      {5, 30, .85, 2, 5},
		MissionExpedition: {30, 60, .75, 3, 15},
		MissionCampaign:   {60, 120, .65, 4, 10},
	}
	counts := make(map[string]map[string]int)
	for _, m := range missions {
		if counts[m.Era] == nil {
			counts[m.Era] = make(map[string]int)
		}
		counts[m.Era][m.Kind]++
		w, ok := want[m.Kind]
		if !ok {
			t.Errorf("%s has unknown kind %q", m.Key, m.Kind)
			continue
		}
		if m.MinMinutes != w.min || m.MaxMinutes != w.max || m.BaseSuccess != w.success || m.Yield != w.yield || m.CostMinutes != w.cost {
			t.Errorf("%s has wrong tuning: %+v", m.Key, m)
		}
		if m.Key == "" || m.Name == "" || m.Description == "" {
			t.Errorf("incomplete mission: %+v", m)
		}
	}
	for _, era := range Epochs() {
		if got := counts[era.Key]; len(got) != 3 || got[MissionScout] != 1 || got[MissionExpedition] != 1 || got[MissionCampaign] != 1 {
			t.Errorf("%s kinds: %v", era.Key, got)
		}
	}
}
