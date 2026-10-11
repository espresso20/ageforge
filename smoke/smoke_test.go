package smoke

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

// TestBotClearsPrimitiveAge keeps the autoplayer honest in the regular test
// suite: one seed, legitimate play only, must get out of the Primitive Age
// without a panic or an invariant violation. The full run lives behind
// `make smoke`.
func TestBotClearsPrimitiveAge(t *testing.T) {
	needBotPlay(t)
	if testing.Short() {
		t.Skip("plays ~12k ticks")
	}
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	cfg := DefaultConfig()
	cfg.MaxSim = 9 * time.Hour
	cfg.StopAge = "stone_age"
	res := Run(cfg, 1)
	for _, a := range res.Anomalies {
		if a.Kind != KindSoftlock || a.Check != "run_budget" {
			t.Errorf("%s/%s: %s\n%s", a.Kind, a.Check, a.Message, a.Dump)
		}
	}
	if len(res.Ages) == 0 || res.Ages[0].Age != "primitive_age" {
		t.Fatalf("bot never left the Primitive Age in %s simulated (outcome %s, ended in %s)", cfg.MaxSim, res.Outcome, res.FinalAge)
	}
	if res.Stats.BuildingsCompleted == 0 || res.Stats.Actions["build_required"] == 0 {
		t.Errorf("suspiciously idle bot: %+v", res.Stats.Actions)
	}
}

func TestReportRendersAnomaliesAndPacing(t *testing.T) {
	runs := []*RunResult{{
		Seed: 7, Outcome: OutcomeSoftlock, FinalAge: "stone_age",
		Ages: []AgeSplit{{Cycle: 1, Age: "primitive_age", Ticks: 100, Seconds: 200}},
		Anomalies: []*Anomaly{{
			Kind: KindSoftlock, Check: "no_progress", Message: "stuck", Seed: 7, Cycle: 1,
			Age: "stone_age", Count: 1, Dump: "state",
		}},
	}}
	s := NewSummary("quick", DefaultConfig(), time.Now(), runs)
	if !s.Failed || s.Anomalies != 1 {
		t.Fatalf("summary should fail with one anomaly: %+v", s)
	}
	var md, js bytes.Buffer
	if err := s.WriteMarkdown(&md); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteJSON(&js); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"FAIL", "| 1 | primitive_age | 1 |", "softlock: no_progress", "state"} {
		if !strings.Contains(md.String(), want) {
			t.Errorf("markdown missing %q:\n%s", want, md.String())
		}
	}
	if !strings.Contains(js.String(), `"check": "no_progress"`) {
		t.Errorf("json missing anomaly:\n%s", js.String())
	}
}
