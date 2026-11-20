package rules

import (
	"slices"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Every badge of the core ruleset counts toward one worker domain.
func TestEveryBadgeHasADomain(t *testing.T) {
	doms := config.WorkerDomains()
	for _, b := range Core().Badges() {
		if d := b.Domain(); !slices.Contains(doms, d) {
			t.Errorf("badge %s (family %q, subject %q) resolves to %q, not a worker domain", b.Key, b.Family, b.Subject, d)
		}
	}
}
