package smoke

import (
	"os"
	"testing"
)

// botTestsEnv is the one switch for the unit tests in this package that need
// the test bot to play well under the economy's current numbers. The owner
// switched the bot off on 2026-10-10 ("until i say turn it back on") because
// its play was being used to choose numbers; the same note, and the way to
// switch the bot back on in CI, is in .github/workflows/go.yml. To turn these
// tests back on, set AGEFORGE_BOT_TESTS=1 (or delete needBotPlay and its
// calls). Tests that only use the bot's helpers without judging its play do
// not call it and keep running.
const botTestsEnv = "AGEFORGE_BOT_TESTS"

// needBotPlay skips t unless the bot's tests are switched on.
func needBotPlay(t *testing.T) {
	t.Helper()
	if os.Getenv(botTestsEnv) == "" {
		t.Skip("the test bot is switched off while the economy's numbers are re-derived by calculation (owner, 2026-10-10); set " + botTestsEnv + "=1 to run the tests that need it to play")
	}
}
