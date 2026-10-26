package smoke

import (
	"maps"
	"slices"
	"testing"
)

func TestDocsyncParsers(t *testing.T) {
	cmds, accepts := registryCommands()
	if len(cmds) < 30 || len(accepts["trade"]) == 0 || !accepts["trade"]["route"] {
		t.Errorf("registry read looks wrong: %d commands, trade accepts %v", len(cmds), accepts["trade"])
	}
	if subs := offeredSubcommands(); !slices.Contains(subs["trade"], "route start") || slices.Contains(subs["gather"], "food") {
		t.Errorf("offered subcommands look wrong: trade %v, gather %v", subs["trade"], subs["gather"])
	}
	claims, err := docClaims("..")
	if err != nil {
		t.Fatal(err)
	}
	// The landing page states no counts, so it may carry no hero stats. If
	// it ever does, the claims read here are checked against the game by
	// TestDocsMatchTheGame; this only keeps the parser honest.
	if len(claims) == 0 {
		t.Error("docClaims found no claims in any file")
	}
	if m := heroStat.FindStringSubmatch(`<span class="hstat-n">22</span><span class="hstat-l">Ages</span>`); m == nil || m[1] != "22" || m[2] != "Ages" {
		t.Errorf("hero stat parser read %v", m)
	}
	if n, rows, err := lineageTable(".."); err != nil || n == 0 || rows == 0 {
		t.Errorf("lineage table: heading %d, rows %d, err %v", n, rows, err)
	}
	md := "| Key | Tab |\n|---|---|\n| `e` | Economy |\n\n| Shortcut | Command |\n|---|---|\n| `b` | `build` |\n| `h`, `?` | `help` |\n"
	want := map[string]string{"b": "build", "h": "help", "?": "help"}
	if got := documentedShortcuts(md); !maps.Equal(got, want) {
		t.Errorf("shortcuts table read as %v, want %v", got, want)
	}
}

// TestDocsMatchTheGame runs the docsync scenario's checks on this tree as a
// unit test: the numbers the docs quote (the badges and themes on the
// landing page among them) and the commands tables against the game. The
// scenario runs in the smoke suite too; here a stale count fails `go test`
// before a pull request's checks do.
func TestDocsMatchTheGame(t *testing.T) {
	res := &Findings{}
	runDocsync("..", res)
	for _, f := range res.Failures {
		t.Errorf("[%s] %s", f.Check, f.Message)
	}
}
