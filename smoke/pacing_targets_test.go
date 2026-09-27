package smoke

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestPacingTargetsMatchConfig: the game derives its economy from
// config.AgeTargets (paybacks, build and research caps) and the harness
// grades runs against PacingTargets. Two copies of one table drift, so they
// must agree for every age the harness paces.
func TestPacingTargetsMatchConfig(t *testing.T) {
	for age, want := range PacingTargets {
		if got := config.AgeTargets[age]; got != want {
			t.Errorf("%s: config.AgeTargets = %v, smoke.PacingTargets = %v", age, got, want)
		}
	}
}
