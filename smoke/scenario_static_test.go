package smoke

import (
	"slices"
	"testing"
)

// TestStaticScenarioReport runs the static scenario as cmd/smoke does and
// holds it to the same standard as the tests beside it: no failures, and the
// report's sections in the order the reader finds them.
func TestStaticScenarioReport(t *testing.T) {
	res := &Findings{}
	runStatic(res)
	for _, f := range res.Failures {
		t.Errorf("[%s] %s", f.Check, f.Message)
	}
	var titles []string
	for _, s := range res.Sections {
		titles = append(titles, s.Title)
	}
	want := []string{
		"Static gate check", "Storage check", "Milestone feasibility", "Harbinger prices against the warning",
		"Faith strength", "Culture strength", "Culture prices", "Depth points", "Research Covenant", "Feature locks",
		"Bonus pools against their caps", "The economy's properties, by calculation", "Badge Covenant",
	}
	if !slices.Equal(titles, want) {
		t.Errorf("the static report's sections are %q, want %q", titles, want)
	}
	if res.Summary == "" {
		t.Error("the static scenario left no summary")
	}
}
