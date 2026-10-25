package smoke

import (
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// TestAccountsScenario runs the accounts scenario in a temp data directory,
// as cmd/smoke does, and checks that it leaves the core ruleset alone.
func TestAccountsScenario(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	set := rules.Core()
	before := set.Digest()
	res := &Findings{}
	runAccounts(1, res)
	for _, f := range res.Failures {
		t.Errorf("accounts: %s: %s\n%s\n%s", f.Check, f.Message, f.Repro, f.Detail)
	}
	if after := set.Digest(); after != before {
		t.Errorf("the accounts scenario changed the core ruleset (digest %s, was %s)", after, before)
	}
}
