package smoke

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/ui"
)

// commandCorpus is the kept corpus of FuzzCommand: one of each shape of
// thing a player types, plus the malformed ones that have mattered. go test
// runs these; go test -fuzz=FuzzCommand ./smoke grows from them.
var commandCorpus = []string{
	"", " ", "help", "status", "build", "build hut", "build hut 5", "build hut -1", "build hut 99999999999999999999",
	"build nosuchthing", "sell hut", "upgrade hut", "research", "research language", "research cancel",
	"advance", "wonder collect all", "wonder collect wood 1e309", "recruit max", "recruit 5", "assign war_camp 5",
	"workers share food 40", "workers share food 400%", "workers share knowledge -5",
	"trade gold knowledge 10", "trade wood wood 1", "trade list", "trade route start local_barter",
	"plan add build hut 3", "plan research language", "plan trade wood stone 10", "plan remove 1", "plan up 0", "plan clear",
	"expedition scout_party", "campaign raid_bandits", "prestige", "prestige confirm", "prestige buy legacy_plan",
	"catastrophe", "appease", "brace", "invite", "endure", "succumb", "rates", "stats", "milestones", "badges",
	"build hut\x00", "build\thut", "BUILD HUT", "build  hut   5  ", "plan add build " + strings.Repeat("a", 512),
	strings.Repeat("build hut ", 64), "🙂", "build 🙂 5", "research \"language\"", "trade gold knowledge NaN", "trade gold knowledge Inf",
}

// FuzzCommand puts anything at all through the game's command handler on a
// fresh game and lets a few ticks pass. Nothing typed may panic the game or
// stop its clock.
func FuzzCommand(f *testing.F) {
	for _, s := range commandCorpus {
		f.Add(s)
	}
	restore := game.SetDataDirForTest(f.TempDir())
	f.Cleanup(restore)
	f.Fuzz(func(t *testing.T, cmd string) {
		ge := game.NewGameEngine()
		ge.SeedRNG(1)
		ge.StepTicks(3)
		before := ge.GetState().Tick
		ui.HandleCommand(cmd, ge)
		ge.StepTicks(3)
		if after := ge.GetState().Tick; after != before+3 {
			t.Fatalf("after %q the clock went from tick %d to %d, want %d", cmd, before, after, before+3)
		}
	})
}
