package ui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// reply_guard_test.go holds two guards on command replies:
//
//   - submitInput (dashboard.go) drops every reply of Type "success", on the
//     rule that the engine already logged the event. A success reply whose
//     text the engine never logged is a message the player never sees, so
//     such a reply must be Type "info" (or carry no text).
//   - Lists built from Go maps come out in a new order on every call unless
//     they are sorted, so the list commands must print the same text twice.

// replyGuardCases are commands that succeed on newCaseTestEngine in the given
// age (the Primitive Age when empty).
var replyGuardCases = []struct{ age, cmd string }{
	{"", "gather food"},
	{"", "gather wood 25"},
	{"", "build hut"},
	{"", "build hut 2"},
	{"", "build hut max"},
	{"", "recruit"},
	{"", "recruit 1"},
	{"", "recruit max"},
	{"", "assign gathering_camp"},
	{"", "assign gathering_camp all"},
	{"", "unassign gathering_camp"},
	{"", "unassign gathering_camp all"},
	{"", "dismiss gathering_camp"},
	{"", "trade food wood 10"},
	{"", "plan build hut"},
	{"", "plan trade food wood"},
	{"", "plan trade food wood 50"},
	{"", "speed 1.0"},
	{"", "theme " + theme.DefaultKey},
	{"", "prestige buy gather_boost"},
	{"", "save my_branch"},
	{"stone_age", "sell hut"},
	{"bronze_age", "research tool_making"},
	{"bronze_age", "wonder collect stone 10"},
	{"bronze_age", "wonder collect all"},
	{"bronze_age", "expedition scout_party"},
	{"bronze_age", "campaign raid_bandits"},
	{"bronze_age", "trade route start local_barter"},
	{"bronze_age", "diplomacy gift riverlands_tribes"},
	{"bronze_age", "diplomacy rival riverlands_tribes"},
	{"bronze_age", "diplomacy neutral riverlands_tribes"},
	{"bronze_age", "diplomacy embargo riverlands_tribes"},
	{"bronze_age", "diplomacy raid riverlands_tribes"},
	{"colonial_age", "blackmarket gold"},
	{"colonial_age", "festival confirm yes"},
	{"", "account backup"},
}

// playerLogCount is the number of log entries a player can see (debug lines
// are for dumps only).
func playerLogCount(ge *game.GameEngine) int {
	n := 0
	for _, e := range ge.GetLogs() {
		if e.Type != "debug" {
			n++
		}
	}
	return n
}

// TestSuccessRepliesAreLogged: a reply of Type "success" with text must come
// with an engine log line written during the command; otherwise the player
// never sees it and the reply must be Type "info" instead.
func TestSuccessRepliesAreLogged(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	t.Chdir(t.TempDir())
	prevTheme := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prevTheme) })

	cases := append([]struct{ age, cmd string }{}, replyGuardCases...)
	cases = append(cases, struct{ age, cmd string }{"", "account export " + filepath.Join(t.TempDir(), "x.json")})
	for _, c := range cases {
		ge := newCaseTestEngine(t, c.age)
		before := playerLogCount(ge)
		res := HandleCommand(c.cmd, ge)
		if res.Type == "error" {
			t.Errorf("%q (age %q) was refused, so its success path goes unchecked; fix the case: %s", c.cmd, c.age, res.Message)
			continue
		}
		if res.Type != "success" || res.OverlayName != "" || res.OpenCatastrophe || playerLogCount(ge) > before {
			continue
		}
		if strings.TrimSpace(res.Message) != "" {
			t.Errorf("%q: Type success with text %q but the engine logged nothing; the player never sees it (use Type info)", c.cmd, res.Message)
		} else {
			t.Errorf("%q: Type success, no text, and the engine logged nothing: the command succeeds in silence", c.cmd)
		}
	}
}

// TestListCommandsAreStable runs each map-driven list command twice on the
// same engine: the two replies must match line for line.
func TestListCommandsAreStable(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	for _, age := range []string{"", "bronze_age", "colonial_age"} {
		ge := newCaseTestEngine(t, age)
		ge.Diplomacy.DiscoverFaction("merchant_guild")
		for _, cmd := range []string{"build", "status", "rates", "research list", "trade list", "upgrade", "prestige shop"} {
			first := HandleCommand(cmd, ge).Message
			for i := 0; i < 5; i++ {
				if again := HandleCommand(cmd, ge).Message; again != first {
					t.Errorf("%q (age %q) printed a different list on a second call:\n%s\n---\n%s", cmd, age, first, again)
					break
				}
			}
		}
	}
}

// TestUsageFormsSurviveTview backs the bracket guard's exemption for registry
// help forms (copy_guard_test.go, copyLit.usageForm): each form must render
// in full both in the Help panel (helpForm) and in a usage line in the log
// (usageFor, then the log's safeTags). "[count]" would otherwise vanish as a
// color tag.
func TestUsageFormsSurviveTview(t *testing.T) {
	for _, c := range registry() {
		for _, u := range appendHelpRows(nil, c) {
			if got := visible(helpForm(u.Form)); got != u.Form {
				t.Errorf("help panel renders %q as %q", u.Form, got)
			}
			i := strings.IndexAny(u.Form, "<[")
			if i < 0 {
				continue // no slots: never a usage line
			}
			line := usageFor(strings.TrimSpace(u.Form[:i]))
			if got := visible(safeTags(line)); got != line {
				t.Errorf("log renders %q as %q", line, got)
			}
		}
	}
}

// TestUsagePathsExist: every usageFor("...") and helpRow("...") call in
// input.go names a command path the registry has a help row for, so a usage
// line never falls back to a bare path.
func TestUsagePathsExist(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "input.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	ast.Inspect(f, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return true
		}
		id, ok := call.Fun.(*ast.Ident)
		if !ok || (id.Name != "usageFor" && id.Name != "helpRow") {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		path, _ := strconv.Unquote(lit.Value)
		n++
		if helpRow(path).Text == "" {
			t.Errorf("input.go:%d: %s(%q) matches no registry help row", fset.Position(lit.Pos()).Line, id.Name, path)
		}
		return true
	})
	if n == 0 {
		t.Fatal("found no usageFor calls in input.go")
	}
	// Paths input.go builds at run time.
	for _, p := range []string{"diplomacy ally", "diplomacy rival", "diplomacy embargo", "diplomacy neutral",
		"diplomacy gift", "diplomacy tribute", "diplomacy raid", "plan up", "plan down",
		"account list", "account switch", "account export", "account backup", "account import"} {
		if helpRow(p).Text == "" {
			t.Errorf("%q matches no registry help row", p)
		}
	}
}

// TestUnknownCommandSuggests: a mistyped command names the closest one.
func TestUnknownCommandSuggests(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	ge := game.NewGameEngine()
	for in, want := range map[string]string{
		"biuld":   "Unknown command 'biuld'. Did you mean 'build'? Type help for all commands.",
		"reserch": "Unknown command 'reserch'. Did you mean 'research'? Type help for all commands.",
		"xyzzy":   "Unknown command 'xyzzy'. Type help for all commands.",
	} {
		if res := HandleCommand(in, ge); res.Type != "error" || res.Message != want {
			t.Errorf("%q = %+v, want error %q", in, res, want)
		}
	}
}

// TestSaveNamesAreOneWord: `save my run` used to save as "my".
func TestSaveNamesAreOneWord(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	ge := game.NewGameEngine()
	for in, want := range map[string]string{
		"save my run": "Save names are one word (try 'save my_run').",
		"load my run": "Save names are one word (try 'load my_run').",
		"load nosuch": "No save named 'nosuch'. Type saves to list them.",
	} {
		if res := HandleCommand(in, ge); res.Type != "error" || res.Message != want {
			t.Errorf("%q = %+v, want error %q", in, res, want)
		}
	}
}

// TestDiplomacyPricesMatchEngine holds the prices the help rows and the
// Factions panel quote (allyGoldCost, allyMinOpinion, giftGoldCost,
// giftOpinion) to what the engine charges.
func TestDiplomacyPricesMatchEngine(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	ge := game.NewGameEngine()
	const civ = "riverlands_tribes"
	if err := ge.MeetFactionForTest(civ, allyMinOpinion-giftOpinion); err != nil {
		t.Fatal(err)
	}
	ge.Resources.LoadStorage(map[string]float64{"gold": 10000})
	ge.Resources.LoadAmounts(map[string]float64{"gold": 5000})
	opinion := func() int { return ge.GetState().Diplomacy.Factions[civ].Opinion }
	gold := func() float64 { return ge.GetState().Resources["gold"].Amount }

	if err := ge.SetDiplomaticStatus(civ, "allied"); err == nil {
		t.Errorf("allied at opinion %d, but the help says it needs %d", opinion(), allyMinOpinion)
	}
	before, op := gold(), opinion()
	if err := ge.SendGift(civ); err != nil {
		t.Fatal(err)
	}
	if spent := before - gold(); spent != giftGoldCost {
		t.Errorf("a gift cost %v gold, the help says %d", spent, giftGoldCost)
	}
	if got := opinion() - op; got != giftOpinion {
		t.Errorf("a gift raised opinion by %d, the help says %d", got, giftOpinion)
	}
	before = gold()
	if err := ge.SetDiplomaticStatus(civ, "allied"); err != nil {
		t.Fatalf("ally at opinion %d: %v", opinion(), err)
	}
	if spent := before - gold(); spent != allyGoldCost {
		t.Errorf("allying cost %v gold, the help says %d", spent, allyGoldCost)
	}
}
