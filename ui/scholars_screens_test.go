package ui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The screens and the scholars: a tech in progress shows a bar that fills to
// 100 percent, its percent and its time left; a tech not started shows the
// time the scholars researching now would take; nothing shows knowledge as
// something held or spent. Each test reads the real state and the real screen.

// scholarGame is a game in the Stone Age with a few scholars, Language known
// and Primitive Writing in progress.
func scholarGame(t *testing.T) *game.GameEngine {
	t.Helper()
	ge := gameIn(t, "stone_age")
	ge.GrantTechsForTest("language")
	ge.SetTownForTest(map[string]int{"story_circle": 2, "elders_hall": 1}, 0)
	if err := ge.StartResearch("primitive_writing"); err != nil {
		t.Fatal(err)
	}
	return ge
}

// headerRow is the first line of the Research panel at w by h.
func headerRow(st game.GameState, w, h int) string {
	_, _, g := drawTree(st, w, h, treeView{sel: "primitive_writing"})
	return strings.Split(g.String(), "\n")[0]
}

// The bar in the panel's header grows with the work done and stands at 100
// percent when the tech is done, at both sizes the panel must read at.
func TestResearchBarFillsToOneHundred(t *testing.T) {
	ge := scholarGame(t)
	for _, sz := range [][2]int{{80, 24}, {100, 30}} {
		st := ge.GetState()
		if st.Research.TicksLeft <= 0 || st.Research.NeedsScholars {
			t.Fatalf("staging: %+v", st.Research)
		}
		row := headerRow(st, sz[0], sz[1])
		for _, want := range []string{"Primitive Writing", fmt.Sprintf("%d%%", researchPercent(st)), "left"} {
			if !strings.Contains(row, want) {
				t.Errorf("%dx%d: the header lacks %q: %q", sz[0], sz[1], want, row)
			}
		}
	}
	last := -1
	for i := 0; i < 4000 && ge.GetState().Research.CurrentTech != ""; i++ {
		ge.StepTicks(5)
		st := ge.GetState()
		if st.Research.CurrentTech == "" {
			break
		}
		p := researchPercent(st)
		if p < last {
			t.Fatalf("the bar went back: %d%% after %d%%", p, last)
		}
		last = p
		row := headerRow(st, 100, 30)
		want := strings.Repeat("█", p/10) + strings.Repeat("░", 10-p/10)
		if !strings.Contains(row, want) {
			t.Fatalf("at %d%% the bar should read %q: %q", p, want, row)
		}
	}
	if last <= 0 {
		t.Fatal("the research never moved")
	}
	st := ge.GetState()
	if st.Research.CurrentTech != "" || !st.Research.Techs["primitive_writing"].Researched {
		t.Fatalf("Primitive Writing did not finish: %+v", st.Research.CurrentTech)
	}
	if got := researchPercent(st); got != 0 {
		t.Errorf("with nothing in progress the percent reads %d", got)
	}
}

// The percent comes from the work done over the price, so it reads right
// with no scholar researching: the bar stays where it is, and the screens
// say the tech needs scholars instead of a time.
func TestResearchWithNoScholarsResearchingSaysSo(t *testing.T) {
	ge := scholarGame(t)
	ge.StepTicks(40)
	if _, err := ge.SetResearchShare(0); err != nil {
		t.Fatal(err)
	}
	st := ge.GetState()
	if !st.Research.NeedsScholars || st.Research.TicksLeft != 0 {
		t.Fatalf("staging: %+v", st.Research)
	}
	if researchPercent(st) <= 0 {
		t.Fatalf("the percent is %d after work was done", researchPercent(st))
	}
	card := cardOf(st, "primitive_writing")
	if !strings.Contains(card, "It needs scholars researching to go on.") || !strings.Contains(card, fmt.Sprintf("%d%% done", researchPercent(st))) {
		t.Errorf("the card of a tech nobody researches:\n%s", card)
	}
	if row := headerRow(st, 100, 30); !strings.Contains(row, "needs scholars researching") || strings.Contains(row, "~0s") {
		t.Errorf("the header: %q", row)
	}
	list := cmdResearchList(ge).Message
	if !strings.Contains(list, "needs scholars researching") || strings.Contains(list, "~0s") {
		t.Errorf("research list:\n%s", list)
	}
	// Each tech in the list reads that it needs scholars, not a time.
	for _, line := range strings.Split(list, "\n") {
		if strings.Contains(line, "[cyan]ritual[-]") && !strings.Contains(line, "needs scholars researching") {
			t.Errorf("a tech with nobody researching reads %q", line)
		}
	}
}

// A tech not started shows the time the scholars researching now would take,
// on its card, in the list and on the status line; fewer researching scholars
// make it longer.
func TestTechTimeIsComputedFromTheScholarsResearching(t *testing.T) {
	ge := scholarGame(t)
	full := ge.GetState()
	ts := full.Research.Techs["ritual"]
	if ts.TicksNeeded <= 0 {
		t.Fatalf("Ritual takes %d ticks", ts.TicksNeeded)
	}
	check := func(st game.GameState, ticks int) {
		t.Helper()
		want := formatTicks(ticks, st)
		if card := cardOf(st, "ritual"); !strings.Contains(card, "It takes "+want+".") {
			t.Errorf("the card should say it takes %s:\n%s", want, card)
		}
		if list := cmdResearchList(ge).Message; !strings.Contains(list, "Ritual ("+want+")") {
			t.Errorf("the list should give %s:\n%s", want, list)
		}
		m := buildTree(st, treeGeomFor(120, false), false, "ritual")
		if line := m.statusLine(st, treeView{sel: "ritual"}); !strings.Contains(line, want) {
			t.Errorf("the status line should give %s: %q", want, line)
		}
	}
	check(full, ts.TicksNeeded)
	if _, err := ge.SetResearchShare(50); err != nil {
		t.Fatal(err)
	}
	half := ge.GetState()
	if got := half.Research.Techs["ritual"].TicksNeeded; got <= ts.TicksNeeded {
		t.Errorf("with half the scholars researching Ritual takes %d ticks, not more than %d", got, ts.TicksNeeded)
	}
	check(half, half.Research.Techs["ritual"].TicksNeeded)
}

// With nothing in progress and nothing queued every scholar teaches, yet a
// tech shows the time it would take the scholars set to research. With the
// share at 0 the screens say it needs scholars.
func TestIdleScholarsStillShowATechTime(t *testing.T) {
	ge := scholarGame(t)
	if err := ge.CancelResearch(); err != nil {
		t.Fatal(err)
	}
	st := ge.GetState()
	if st.Research.CurrentTech != "" || st.Scholars.Researching != 0 || st.Scholars.Total == 0 {
		t.Fatalf("staging: %+v", st.Scholars)
	}
	ticks := st.Research.Techs["ritual"].TicksNeeded
	if ticks <= 0 {
		t.Fatalf("an idle town shows Ritual at %d ticks", ticks)
	}
	want := formatTicks(ticks, st)
	if card := cardOf(st, "ritual"); !strings.Contains(card, "It takes "+want+".") || strings.Contains(card, "needs scholars") {
		t.Errorf("idle card:\n%s", card)
	}
	if list := cmdResearchList(ge).Message; !strings.Contains(list, "Ritual ("+want+")") || strings.Contains(list, "needs scholars") {
		t.Errorf("idle list:\n%s", list)
	}
	if _, err := ge.SetResearchShare(0); err != nil {
		t.Fatal(err)
	}
	st = ge.GetState()
	if card := cardOf(st, "ritual"); !strings.Contains(card, needsScholars) {
		t.Errorf("share 0 card:\n%s", card)
	}
	if list := cmdResearchList(ge).Message; !strings.Contains(list, "Ritual ("+needsScholarsShort+")") {
		t.Errorf("share 0 list:\n%s", list)
	}
}

// The queue is shown in order, and the scholars line says how many research
// and teach and what teaching adds.
func TestResearchListShowsQueueAndScholars(t *testing.T) {
	ge := scholarGame(t)
	if err := ge.PlanAddResearch("ritual"); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddResearch("pottery"); err != nil {
		t.Fatal(err)
	}
	st := ge.GetState()
	list := cmdResearchList(ge).Message
	i, j := strings.Index(list, "1. "), strings.Index(list, "2. ")
	if i < 0 || j < i || !strings.Contains(list, "Queued:") {
		t.Errorf("the queue is not in order:\n%s", list)
	}
	for n, key := range st.Research.Queue {
		if want := fmt.Sprintf("%d. %s", n+1, st.Research.Techs[key].Name); !strings.Contains(list, want) {
			t.Errorf("the list lacks %q:\n%s", want, list)
		}
	}
	for _, out := range []string{list, cmdResearchList(ge).Message} {
		if want := scholarsLine(st); !strings.Contains(out, want) {
			t.Errorf("no scholars line %q:\n%s", want, out)
		}
	}
	s := st.Scholars
	want := fmt.Sprintf("%d research, %d teach. Teaching adds", s.Researching, s.Teaching)
	if !strings.Contains(scholarsLine(st), want) {
		t.Errorf("scholars line = %q", scholarsLine(st))
	}
}

// The scholars line is on the card and in the header at both sizes, and the
// Activity panel shows the bar with the percent.
func TestScholarsLineWhereResearchIsShown(t *testing.T) {
	ge := scholarGame(t)
	st := ge.GetState()
	if card := cardOf(st, "ritual"); !strings.Contains(card, scholarsLine(st)) {
		t.Errorf("the card lacks the scholars line:\n%s", card)
	}
	// The header gives the bar first and the scholars what room is left: at
	// 100 columns there is room; at 80 the card and the list carry the line.
	row := headerRow(st, 100, 30)
	if want := fmt.Sprintf("%d research, %d teach", st.Scholars.Researching, st.Scholars.Teaching); !strings.Contains(row, want) {
		t.Errorf("100x30: the header lacks %q: %q", want, row)
	}
	if row := headerRow(st, 80, 24); !strings.Contains(row, "Primitive Writing") || !strings.Contains(row, "░") {
		t.Errorf("80x24: the header lost the bar: %q", row)
	}
	logs := logsProvider(st, 100)
	if !strings.Contains(logs, fmt.Sprintf("%d%%", researchPercent(st))) || !strings.Contains(logs, "█") && !strings.Contains(logs, "░") {
		t.Errorf("the Activity panel does not show the bar:\n%s", logs)
	}
}

// With nothing to research every scholar teaches, and the bonus is the whole
// of the age's: the stats and the rates both give it its own line, read off
// the state.
func TestTeachingBonusHasItsOwnLine(t *testing.T) {
	ge := scholarGame(t)
	if err := ge.CancelResearch(); err != nil {
		t.Fatal(err)
	}
	st := ge.GetState()
	if st.Scholars.Teaching != st.Scholars.Total || st.Scholars.TeachingBonus <= 0 {
		t.Fatalf("staging: %+v", st.Scholars)
	}
	pct := fmt.Sprintf("+%.0f%%", st.Scholars.TeachingBonus*100)
	stats := statsProvider(st, 120)
	if want := fmt.Sprintf("[green]%-7s[-] Worker output from teaching", pct); !strings.Contains(stats, want) {
		t.Errorf("the stats have no teaching line for %s:\n%s", pct, section(t, stats, "Research bonuses", "Resource rates"))
	}
	rates := HandleCommand("rates", ge).Message
	if want := "Teaching: " + pct + " to what workers make (" + fmt.Sprint(st.Scholars.Total) + " of " + fmt.Sprint(st.Scholars.Total) + " scholars teach)"; !strings.Contains(rates, want) {
		t.Errorf("rates lack %q:\n%s", want, rates)
	}
	// Half research: the bonus is half, and both screens say so.
	if err := ge.StartResearch("primitive_writing"); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.SetResearchShare(50); err != nil {
		t.Fatal(err)
	}
	st = ge.GetState()
	if st.Scholars.TeachingBonus >= st.Scholars.TeachingFull || st.Scholars.TeachingBonus <= 0 {
		t.Fatalf("staging: %+v", st.Scholars)
	}
	half := fmt.Sprintf("+%.1f%%", st.Scholars.TeachingBonus*100)
	half = strings.Replace(half, ".0%", "%", 1)
	if rates := HandleCommand("rates", ge).Message; !strings.Contains(rates, "Teaching: "+half) {
		t.Errorf("rates at half: want %s:\n%s", half, rates)
	}
}

// A knowledge building shows its scholars a copy and its copies against its
// cap, and at the cap says it can be built no further in the words the
// Buildings panel gives a building at its limit.
func TestKnowledgeBuildingShowsScholarsAndCap(t *testing.T) {
	ge := scholarGame(t)
	st := ge.GetState()
	bs := st.Buildings["story_circle"]
	if bs.Scholars != 2 || bs.ScholarCap <= 0 {
		t.Fatalf("staging: %+v", bs)
	}
	want := fmt.Sprintf("%d of %d built.", bs.Count, bs.ScholarCap)
	if got := buildingText(bs); !strings.HasSuffix(got, want) || !strings.Contains(got, "2 scholars") {
		t.Errorf("buildingText = %q, want 2 scholars and it to end %q", got, want)
	}
	// A description that does not name the scholars gets them added.
	if got := buildingText(game.BuildingState{Description: "A hall.", Scholars: 3, ScholarCap: 5, Count: 1}); got != "A hall. 3 scholars a copy, 1 of 5 built." {
		t.Errorf("buildingText without the scholars in the description = %q", got)
	}
	if !strings.Contains(buildingsProvider(st, 120), want) {
		t.Errorf("the Buildings overlay lacks %q", want)
	}
	if !strings.Contains(cmdBuild(nil, ge).Message, want) {
		t.Errorf("the build list lacks %q", want)
	}
	if atLimit(bs) != "" {
		t.Errorf("a building under its cap says %q", atLimit(bs))
	}
	ge.SetTownForTest(map[string]int{"story_circle": bs.ScholarCap}, 0)
	st = ge.GetState()
	bs = st.Buildings["story_circle"]
	if !bs.AtMaxCount || atLimit(bs) != "Building limit reached." {
		t.Fatalf("at the cap: %+v", bs)
	}
	if !strings.Contains(buildingsProvider(st, 120), "Building limit reached.") || !strings.Contains(cmdBuild(nil, ge).Message, "Building limit reached.") {
		t.Errorf("a knowledge building at its cap is not said to be at its limit")
	}
	// A building that is not a knowledge building adds nothing.
	if got := buildingText(game.BuildingState{Description: "A hut."}); got != "A hut." {
		t.Errorf("buildingText of a plain building = %q", got)
	}
}

// The scholars command: kinds and counts, the split, the bonus; a percent
// sets the share through the engine and prints its reply.
func TestScholarsCommand(t *testing.T) {
	ge := scholarGame(t)
	st := ge.GetState()
	res := HandleCommand("scholars", ge)
	if res.Type != "info" {
		t.Fatalf("scholars: %+v", res)
	}
	for _, g := range st.Scholars.Kinds {
		if want := fmt.Sprintf("%s × %d", g.Kind, g.Count); !strings.Contains(res.Message, want) {
			t.Errorf("scholars lacks %q:\n%s", want, res.Message)
		}
	}
	for _, want := range []string{fmt.Sprintf("Scholars[-] %d", st.Scholars.Total), scholarsLine(st), "100% of your scholars research. Set it with 'scholars <percent>'."} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("scholars lacks %q:\n%s", want, res.Message)
		}
	}
	for _, in := range []string{"scholars 50", "scholars 50%"} {
		res = HandleCommand(in, ge)
		want, _ := ge.SetResearchShare(50)
		if res.Message != want || res.Type == "error" {
			t.Errorf("%s = %+v, want the engine's reply %q", in, res, want)
		}
	}
	if st := ge.GetState(); st.Scholars.Share != 50 {
		t.Errorf("the share is %d", st.Scholars.Share)
	}
	if res := HandleCommand("scholars 150", ge); res.Type != "error" || !strings.Contains(res.Message, "between 0 and 100 percent") {
		t.Errorf("scholars 150 = %+v", res)
	}
	usage := usageFor("scholars")
	for _, in := range []string{"scholars lots", "scholars 1 2", "scholars 12.5"} {
		if res := HandleCommand(in, ge); res.Type != "error" || res.Message != usage {
			t.Errorf("%s = %+v, want the usage line %q", in, res, usage)
		}
	}
	if !strings.Contains(usage, "scholars [percent]") && !strings.Contains(usage, "scholars") {
		t.Errorf("usage = %q", usage)
	}
	// Completion and help know it like its neighbors.
	if got := NewAutoCompleter(ge)("schol"); !containsWord(got, "scholars") {
		t.Errorf("completing schol = %v", got)
	}
	if help := HandleCommand("help scholars", ge); !strings.Contains(help.Message+help.OverlayName, "scholars") {
		t.Logf("help scholars = %+v", help)
	}
	// A town with no scholars says so in the engine's words.
	fresh := game.NewGameEngine()
	if res := HandleCommand("scholars", fresh); !strings.Contains(res.Message, "no scholars yet: knowledge buildings come with them") {
		t.Errorf("a new game: %q", res.Message)
	}
}

// The knowledge worker domain is gone from every list the player is offered.
func TestKnowledgeIsNotAWorkerDomain(t *testing.T) {
	ge := scholarGame(t)
	st := ge.GetState()
	for _, d := range st.Ruleset().WorkerDomains() {
		if d == "knowledge" {
			t.Fatal("knowledge is still a worker domain")
		}
	}
	if n := len(st.Ruleset().WorkerDomains()); n != 11 {
		t.Errorf("%d worker domains, want 11", n)
	}
	if res := HandleCommand("roster knowledge 40", ge); res.Type != "error" || !strings.Contains(res.Message, "Unknown worker domain") {
		t.Errorf("roster knowledge 40 = %+v", res)
	}
	for _, line := range NewAutoCompleter(ge)("roster ") {
		if strings.Contains(line, "knowledge") {
			t.Errorf("roster offers %q", line)
		}
	}
	if roster := HandleCommand("roster", ge).Message; strings.Contains(strings.ToLower(roster), "knowledge") {
		t.Errorf("the roster names knowledge:\n%s", roster)
	}
}

// No screen shows knowledge as something held, spent, stored or priced.
var knowledgeHeld = regexp.MustCompile(`(?i)knowledge (storage|store|price|cost|rate|production|amount)|of your [^ ]* ?knowledge|\bkp\b|knowledge/tick|spent|Knowledge:`)

func TestNoScreenShowsKnowledgeAsHeldOrSpent(t *testing.T) {
	ge := scholarGame(t)
	if err := ge.PlanAddResearch("ritual"); err != nil {
		t.Fatal(err)
	}
	st := ge.GetState()
	screens := map[string]string{
		"research list":   cmdResearchList(ge).Message,
		"rates":           HandleCommand("rates", ge).Message,
		"stats":           statsProvider(st, 120),
		"activity":        logsProvider(st, 100),
		"buildings":       buildingsProvider(st, 120),
		"build list":      cmdBuild(nil, ge).Message,
		"scholars":        HandleCommand("scholars", ge).Message,
		"research cancel": usageFor("research cancel"),
		"history labels":  historyProvider(st, 100),
		"economy buildings": func() string {
			_, e := buildingEntries(st, 60)
			var sb strings.Builder
			for _, x := range e {
				sb.WriteString(strings.Join(x, "\n"))
			}
			return sb.String()
		}(),
	}
	m := buildTree(st, treeGeomFor(120, false), false, "ritual")
	for _, n := range m.nodes {
		screens["card "+n.key] = cardOf(st, n.key)
		screens["status "+n.key] = m.statusLine(st, treeView{sel: n.key})
	}
	for _, sz := range [][2]int{{80, 24}, {100, 30}} {
		_, _, g := drawTree(st, sz[0], sz[1], treeView{sel: "ritual"})
		screens[fmt.Sprintf("panel %dx%d", sz[0], sz[1])] = g.String()
	}
	for where, text := range screens {
		if loc := knowledgeHeld.FindStringIndex(text); loc != nil {
			lo, hi := max(loc[0]-30, 0), min(loc[1]+30, len(text))
			t.Errorf("%s: %q", where, text[lo:hi])
		}
	}
}

// The resource boxes' resources are the 25 the game has: knowledge is no row.
func TestResourcesBoxHasNoKnowledgeRow(t *testing.T) {
	st := scholarGame(t).GetState()
	for _, r := range resourceRows(st) {
		if strings.EqualFold(r.name, "knowledge") {
			t.Errorf("the Resources box has a %q row", r.name)
		}
	}
}

// `research` completes to the techs that can start now, and to no other.
func TestSuggestResearchGameAware(t *testing.T) {
	ge := scholarGame(t)
	if err := ge.CancelResearch(); err != nil {
		t.Fatal(err)
	}
	st := ge.GetState()
	got := NewAutoCompleter(ge)("research ")
	techs := 0
	for _, line := range got {
		key := strings.TrimPrefix(line, "research ")
		ts, ok := st.Research.Techs[key]
		if !ok {
			continue // list, cancel, card, tree
		}
		techs++
		if !ts.Available {
			t.Errorf("research offers %s, which can't start now", key)
		}
	}
	if techs == 0 {
		t.Errorf("research offers %v; want the available techs", got)
	}
	for key, ts := range st.Research.Techs {
		if ts.Available && !containsWord(got, "research "+key) {
			t.Errorf("research does not offer %s, which can start", key)
		}
	}
}

// A wonder's work for the scholars reads as what it is, not as a bare number.
func TestWonderResearchEffectIsNotANumber(t *testing.T) {
	if got := formatTechEffect(config.Effect{Type: "research", Value: 0.05}); got != "speeds research" {
		t.Errorf("a wonder's research effect reads %q", got)
	}
	st := scholarGame(t).GetState()
	for _, l := range strings.Split(wondersProvider(st, 120), "\n") {
		if strings.TrimSpace(l) == "[green]+0.05[-]" || strings.TrimSpace(l) == "+0.05" {
			t.Errorf("a bare number in the Wonders panel: %q", l)
		}
	}
}

// An active Enlightenment boon reads as the scholars' work, not as a bonus to
// a resource that no longer exists.
func TestEnlightenmentLineNamesScholarOutput(t *testing.T) {
	plain, _ := effectMagnitude(game.EventEffectInfo{Type: "knowledge_rate", Target: "knowledge", Value: 0.2})
	if plain != "+20% scholar output" {
		t.Errorf("the active Enlightenment reads %q", plain)
	}
}
