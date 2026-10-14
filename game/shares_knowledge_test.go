package game

import (
	"slices"
	"testing"

	"github.com/espresso20/ageforge/rules"
)

// The knowledge worker domain is gone: it is not offered, and a share an old
// save set on it is ignored on load while the other shares stay.
func TestKnowledgeIsNoWorkerDomainAndAnOldShareIsIgnored(t *testing.T) {
	set := rules.Core()
	if slices.Contains(set.WorkerDomains(), "knowledge") || len(set.WorkerDomains()) != 11 {
		t.Fatalf("worker domains = %v, want 11 and no knowledge", set.WorkerDomains())
	}
	if IsWorkerDomain(set, "knowledge") {
		t.Error("knowledge is still a worker domain")
	}
	got := cleanSharesIn(set, map[string]float64{"knowledge": 40, "food": 30})
	if _, ok := got["knowledge"]; ok || got["food"] != 30 || len(got) != 1 {
		t.Errorf("shares from an old save = %v, want food 30 only", got)
	}
}
