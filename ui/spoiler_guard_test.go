package ui

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// spoiler_guard_test.go is the no-spoiler guard (playtest 2026-09-29: "don't
// show future civilizations or ages"). A fresh Primitive Age game must show
// nothing the player has not reached. The guard renders the main window,
// every text panel the dashboard registers, the read-only commands, the
// refusals a curious player can provoke and the civilization completions,
// then fails on:
//
//   - the name of any civilization (a fresh game has met none);
//   - any age after the next one. The Next Age goal names the next age on
//     purpose, so the next age is allowed everywhere: it is on screen anyway;
//   - any era after the current one.
//
// The rule behind it lives in spoilers.go (ageKnown, eraKnown).

// spoilerAllowed is the documented allow-list: a spot that may name a later
// age, era or civilization, and why. where is a panel name, "screen" or a
// command; line matches the rendered line. Keep it short, and give every
// entry a reason a player would accept.
var spoilerAllowed = []struct {
	where string
	line  *regexp.Regexp
	why   string
}{
	{"theme list", regexp.MustCompile(`^\s*Cyberpunk\s+cyberpunk\s`),
		"the Cyberpunk theme's own name: theme list names every theme, locked or not, and its lock line names no age (themeUnlockHint)"},
}

// spoilerTerm is one name a fresh game must not show.
type spoilerTerm struct {
	re   *regexp.Regexp
	what string
}

// freshSpoilerTerms lists what a Primitive Age game must not name.
func freshSpoilerTerms() []spoilerTerm {
	var out []spoilerTerm
	add := func(name, what string) {
		out = append(out, spoilerTerm{regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`), what})
	}
	for i, key := range config.AgeOrder() {
		if i < 2 { // the current age, and the next one (the Next Age goal)
			continue
		}
		name := config.AgeByKey()[key].Name
		add(name, "a later age")
	}
	// Ages named by their adjective alone ("an Embassy (Colonial)"). Only
	// adjectives that are not also resources or everyday words.
	for _, adj := range []string{"Classical", "Medieval", "Renaissance", "Colonial", "Industrial",
		"Victorian", "Atomic", "Cyberpunk", "Interstellar", "Galactic", "Transcendent"} {
		add(adj, "a later age, by its adjective")
	}
	for _, ep := range config.Epochs() {
		if ep.Order > 0 {
			add(ep.Name, "a later era")
		}
	}
	for _, f := range config.BaseFactions() {
		add(f.Name, "a civilization not met yet")
	}
	return out
}

// spoilerAllowedLine reports whether line at where is on the allow-list.
func spoilerAllowedLine(where, line string) bool {
	for _, a := range spoilerAllowed {
		if a.where == where && a.line.MatchString(line) {
			return true
		}
	}
	return false
}

// checkSpoilers fails for every term text shows at where.
func checkSpoilers(t *testing.T, where, text string, terms []spoilerTerm) {
	t.Helper()
	for _, line := range strings.Split(text, "\n") {
		if spoilerAllowedLine(where, line) {
			continue
		}
		for _, term := range terms {
			if m := term.re.FindString(line); m != "" {
				t.Errorf("%s names %s (%s): %q", where, m, term.what, strings.TrimSpace(line))
			}
		}
	}
}

// freshSpoilerEngine is a new game one tick in: the Stone Era's harbinger
// has come, and nothing else has happened.
func freshSpoilerEngine(t *testing.T) *game.GameEngine {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	engine := game.NewGameEngine()
	engine.SeedRNG(1)
	engine.StepTicks(1)
	if engine.GetState().Harbinger == nil {
		t.Fatal("no harbinger on the first tick; the guard would miss its text")
	}
	return engine
}

// TestFreshGameShowsNoSpoilers renders what a new player can see and fails
// on any name they have not reached.
func TestFreshGameShowsNoSpoilers(t *testing.T) {
	engine := freshSpoilerEngine(t)
	terms := freshSpoilerTerms()
	st := engine.GetState()

	// Every text panel.
	panels := guardPanels(t, engine)
	names := make([]string, 0, len(panels))
	for name := range panels {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		checkSpoilers(t, name+" panel", guardRendered(panels[name](st, guardPanelWidth)), terms)
	}

	// The main window.
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), engine, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	d.refresh()
	screen := screenAt(t, pages, 180, 56)
	checkSpoilers(t, "screen", screen, terms)
	if !strings.Contains(screen, "Next Age: Stone Age") {
		t.Errorf("the Next Age goal should name the next age:\n%s", screen)
	}

	// Read-only commands.
	for _, cmd := range []string{"status", "rates", "build", "upgrade", "wonder", "festival", "blackmarket",
		"prestige", "prestige shop", "speed", "catastrophe", "research list", "trade list", "trade route list",
		"plan list", "diplomacy deals", "expedition", "campaign", "wonder overflow", "theme list", "saves", "account"} {
		checkSpoilers(t, cmd, guardRendered(HandleCommand(cmd, engine).Message), terms)
	}
}

// TestFreshGameRefusalsShowNoSpoilers: commands a curious player can type
// about things they have not reached are refused without naming them: no
// civilization's name, no later age, and no key of a civilization they did
// not type.
func TestFreshGameRefusalsShowNoSpoilers(t *testing.T) {
	engine := freshSpoilerEngine(t)
	terms := freshSpoilerTerms()
	lastTech := config.Technologies()[len(config.Technologies())-1].Key
	routes := config.BaseTradeRoutes()
	lastRoute := routes[len(routes)-1].Key
	cmds := []string{
		"diplomacy deals ironhold_clans", "diplomacy deals iron", "diplomacy gift iron", "diplomacy gift ironhold_clans",
		"diplomacy ally merchant_guild", "diplomacy accept ironhold_clans 1", "plan deal ironhold_clans 1",
		"research " + lastTech, "plan research " + lastTech, "trade route start " + lastRoute,
		"expedition " + lastExpeditionKey(t), "blackmarket gold", "prestige confirm yes",
	}
	for _, cmd := range cmds {
		reply := guardRendered(HandleCommand(cmd, engine).Message)
		checkSpoilers(t, cmd, reply, terms)
		typed := strings.Fields(cmd)
		for _, f := range config.BaseFactions() {
			if strings.Contains(reply, f.Key) && !containsWord(typed, f.Key) {
				t.Errorf("%q: the reply names %s, which the player never typed: %q", cmd, f.Key, reply)
			}
		}
	}

	// Completion offers no civilization before first contact.
	comp := NewAutoCompleter(engine)
	for _, line := range []string{"diplomacy deals ", "diplomacy ally ", "diplomacy gift ", "diplomacy deals i", "plan deal ", "diplomacy accept "} {
		if got := comp(line); len(got) > 0 {
			t.Errorf("completion for %q offers %v before first contact", line, got)
		}
	}
}

// lastExpeditionKey is an expedition from the last age that has one.
func lastExpeditionKey(t *testing.T) string {
	t.Helper()
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	mm := game.NewMilitaryManager()
	ages := config.AgeOrder()
	for i := len(ages) - 1; i >= 0; i-- {
		if xs := mm.GetAvailableExpeditions(ages[i], order); len(xs) > 0 {
			return xs[len(xs)-1].Key
		}
	}
	t.Fatal("no expeditions in config")
	return ""
}

func containsWord(words []string, w string) bool {
	for _, x := range words {
		if x == w {
			return true
		}
	}
	return false
}
