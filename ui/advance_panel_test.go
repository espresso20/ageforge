package ui

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// advanceSizes are the terminals the panel is drawn at (80x24, 100x30 and
// 160x45) and the panels they give, which are three rows shorter, the
// command bar's.
var advanceSizes = [][2]int{{80, 21}, {80, 24}, {100, 27}, {100, 30}, {160, 42}, {160, 45}}

// meetTheGate stands the town that meets the gate out of its age exactly
// (the next age's counts and this age's wonder) and quiets the era's fate, so
// the advance is not held up by a doom.
func meetTheGate(t *testing.T, eng *game.GameEngine) {
	t.Helper()
	set := eng.Rules()
	st := eng.GetState()
	age, ok := set.Age(set.Next(st.Age))
	if !ok {
		t.Fatalf("no age after %s", st.Age)
	}
	town := map[string]int{}
	for k, n := range age.BuildingReqs {
		town[k] = n
	}
	if w := set.Wonder(st.Age); w != "" {
		town[w] = 1
	}
	eng.SetTownForTest(town, 0)
	_ = eng.ForceQuietFateForTest(set.EraOf(st.Age))
	eng.StepTicks(1)
	if !eng.GetState().AgeReady {
		t.Fatal("staging: the gate is met and the engine does not say ready")
	}
}

// advanceFixture is a staged game in age with the panel wired to its engine.
func advanceFixture(t *testing.T, age string) (*Dashboard, *game.GameEngine, *advancePanel) {
	t.Helper()
	restoreForge(t)
	d, _, eng := stagedDashboard(t, age)
	p := newAdvancePanel()
	p.engine = eng
	p.open(eng.GetState())
	return d, eng, p
}

// sendKey gives the dashboard a key as the app's input capture sees it.
func sendKey(d *Dashboard, k tcell.Key) *tcell.EventKey {
	return d.root.GetInputCapture()(tcell.NewEventKey(k, 0, tcell.ModNone))
}

// charter opens the Advance panel the way a player does: by typing the command.
func TestCharterOpensTheAdvancePanelAndEscCloses(t *testing.T) {
	restoreForge(t)
	d, _, eng := stagedDashboard(t, "bronze_age")
	if res := HandleCommand("charter", eng); res.OverlayName != "charter" || res.Type == "error" {
		t.Fatalf("charter = %+v, want the Advance panel", res)
	}
	d.inputField.SetText("charter")
	d.submitInput()
	if got := d.overlayMgr.ActiveName(); got != "charter" {
		t.Fatalf("after charter the open panel is %q", got)
	}
	if !d.inputField.HasFocus() {
		t.Error("the command bar lost the keyboard to the panel")
	}
	if ev := sendKey(d, tcell.KeyEsc); ev != nil || d.overlayMgr.ActiveName() != "" {
		t.Errorf("Esc did not close the panel (open: %q)", d.overlayMgr.ActiveName())
	}
}

// Up and Down pick a pillar and stop at the ends; Enter, Up and Down belong to
// the prompt when it has text; nothing else is taken.
func TestAdvancePanelKeys(t *testing.T) {
	d, _, _ := advanceFixture(t, "bronze_age")
	d.inputField.SetText("charter")
	d.submitInput()
	p := d.advancePanel
	if p.ui.sel != 0 {
		t.Fatalf("opened on pillar %d", p.ui.sel)
	}
	for want := 1; want <= 3; want++ {
		if sendKey(d, tcell.KeyDown) != nil || p.ui.sel != want {
			t.Errorf("Down went to pillar %d, want %d", p.ui.sel, want)
		}
	}
	if sendKey(d, tcell.KeyDown); p.ui.sel != 3 {
		t.Errorf("Down past the last pillar went to %d", p.ui.sel)
	}
	for want := 2; want >= 0; want-- {
		if sendKey(d, tcell.KeyUp) != nil || p.ui.sel != want {
			t.Errorf("Up went to pillar %d, want %d", p.ui.sel, want)
		}
	}
	if sendKey(d, tcell.KeyUp); p.ui.sel != 0 {
		t.Errorf("Up past the first pillar went to %d", p.ui.sel)
	}
	// With something typed, the keys are the prompt's.
	d.inputField.SetText("bui")
	for _, k := range []tcell.Key{tcell.KeyUp, tcell.KeyDown, tcell.KeyEnter} {
		if sendKey(d, k) == nil {
			t.Errorf("key %v was taken from a prompt with text", k)
		}
	}
	d.inputField.SetText("")
	for _, k := range []tcell.Key{tcell.KeyLeft, tcell.KeyRight, tcell.KeyTab} {
		if sendKey(d, k) == nil {
			t.Errorf("key %v was taken by the panel", k)
		}
	}
	if ev := d.root.GetInputCapture()(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)); ev == nil {
		t.Error("a letter was taken from the prompt")
	}
}

// Enter advances when all four pillars stand and not before, through the
// call the advance command makes; the engine's refusal is the panel's line.
func TestAdvanceEnterAdvancesOnlyWhenReady(t *testing.T) {
	_, eng, p := advanceFixture(t, "bronze_age")
	before := eng.GetState().Age
	if p.view.AllStand {
		t.Fatal("staging: the staged town already meets the gate")
	}
	p.handleKey(tcell.KeyEnter)
	if got := eng.GetState().Age; got != before {
		t.Fatalf("Enter advanced from %s to %s with the pillars short", before, got)
	}
	if !strings.HasPrefix(p.ui.flash, "Not yet: ") || !strings.Contains(p.ui.flash, "still to stand") {
		t.Errorf("Enter on a short gate says %q", p.ui.flash)
	}

	meetTheGate(t, eng)
	p.update(eng.GetState())
	if !p.view.Ready {
		t.Fatalf("staging: the gate is met and the panel is not ready: %+v", p.view.Blockers)
	}
	if foot := renderAdvance(p.view, p.ui, 100, 27).Lines(); !strings.Contains(foot[24], "Enter") || !strings.Contains(foot[24], "Advance to the Iron Age") {
		t.Errorf("a met gate has no Enter prompt: %q", foot[24])
	}
	p.handleKey(tcell.KeyEnter)
	if got := eng.GetState().Age; got != "iron_age" {
		t.Fatalf("Enter on a met gate left the game in %s", got)
	}
	if p.ui.flash != "Advanced to the Iron Age." || !p.ui.flashGood {
		t.Errorf("after advancing the panel says %q", p.ui.flash)
	}
	if p.view.Age != "iron_age" || p.view.Here != 3 {
		t.Errorf("the panel did not move on: %s at %d", p.view.Age, p.view.Here)
	}
}

// A catastrophe pending stands in the way even with the four pillars met: the
// panel says so, and Enter shows the engine's own refusal.
func TestAdvanceBlockedByACatastrophe(t *testing.T) {
	_, eng, p := advanceFixture(t, "iron_age")
	meetTheGate(t, eng)
	if err := eng.ForceCatastropheForTest(); err != nil {
		t.Skipf("no catastrophe to force: %v", err)
	}
	p.update(eng.GetState())
	if !p.view.AllStand || p.view.Ready || len(p.view.Blockers) != 1 {
		t.Fatalf("a met gate with a catastrophe pending: stand %v ready %v blockers %v", p.view.AllStand, p.view.Ready, p.view.Blockers)
	}
	text := renderAdvance(p.view, p.ui, 100, 27).String()
	if !strings.Contains(text, "'catastrophe'") || strings.Contains(text, "Enter  Advance") {
		t.Errorf("the panel does not show the blocker in place of the prompt:\n%s", text)
	}
	age := eng.GetState().Age
	p.handleKey(tcell.KeyEnter)
	if eng.GetState().Age != age || !strings.Contains(p.ui.flash, "'catastrophe'") {
		t.Errorf("Enter with a catastrophe pending: age %s, says %q", eng.GetState().Age, p.ui.flash)
	}
}

// The advance command, when it refuses for a gate not met, says how to open
// the panel.
func TestAdvanceRefusalPointsAtTheCharter(t *testing.T) {
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "bronze_age")
	res := HandleCommand("advance", eng)
	if res.Type != "error" || !strings.Contains(res.Message, "'charter'") {
		t.Errorf("advance on a short gate: %+v", res)
	}
	// The wonder-first refusal too.
	_, _, fresh := stagedDashboard(t, "primitive_age")
	if res := HandleCommand("advance", fresh); !strings.Contains(res.Message, "'charter'") {
		t.Errorf("advance in the first age: %+v", res)
	}
}

// Opening the panel, moving about it, drawing it and closing it changes
// nothing in the game.
func TestAdvancePanelNeverChangesTheGame(t *testing.T) {
	restoreForge(t)
	d, _, eng := stagedDashboard(t, "bronze_age")
	before := eng.GetState()
	d.inputField.SetText("charter")
	d.submitInput()
	for _, k := range []tcell.Key{tcell.KeyDown, tcell.KeyDown, tcell.KeyUp} {
		sendKey(d, k)
	}
	d.advancePanel.update(eng.GetState())
	_ = renderAdvance(d.advancePanel.view, d.advancePanel.ui, 100, 27)
	sendKey(d, tcell.KeyEsc)
	if d.overlayMgr.ActiveName() != "" {
		t.Fatal("the panel is still open")
	}
	after := eng.GetState()
	// The statistics carry a play clock that moves between any two
	// snapshots; everything else must stand.
	before.Stats.PlayTime, after.Stats.PlayTime = 0, 0
	vb, va := reflect.ValueOf(before), reflect.ValueOf(after)
	for i := 0; i < vb.NumField(); i++ {
		if !reflect.DeepEqual(vb.Field(i).Interface(), va.Field(i).Interface()) {
			t.Errorf("the game changed while the panel was open: %s", vb.Type().Field(i).Name)
		}
	}
}

// The calendar does not run backwards: sell a building and the pillars show
// the loss, but the year the panel read stays.
func TestAdvanceYearHoldsWhenTheGateSlips(t *testing.T) {
	_, eng, p := advanceFixture(t, "bronze_age")
	set := eng.Rules()
	age, _ := set.Age(set.Next("bronze_age"))
	town := map[string]int{}
	for k, n := range age.BuildingReqs {
		town[k] = n
	}
	eng.SetTownForTest(town, 0)
	eng.StepTicks(1)
	p.update(eng.GetState())
	high := p.view
	if high.Progress < 0.5 {
		t.Fatalf("staging: progress %.2f", high.Progress)
	}
	for k := range town {
		town[k] = 0
	}
	eng.SetTownForTest(town, 0)
	eng.StepTicks(1)
	p.update(eng.GetState())
	low := p.view
	if low.Progress >= high.Progress {
		t.Fatalf("staging: the gate did not slip (%.2f then %.2f)", high.Progress, low.Progress)
	}
	if low.Year != high.Year || low.YearText != high.YearText || low.Share != high.Share {
		t.Errorf("the year went from %s (%.2f) to %s (%.2f)", high.YearText, high.Share, low.YearText, low.Share)
	}
	lines := renderAdvance(low, p.ui, 100, 27).String()
	if want := "Year " + high.YearText + " · "; !strings.Contains(lines, want) {
		t.Errorf("the march does not read %q:\n%s", want, lines)
	}
}

// advanceStates are the gates the panel is drawn for: partway, met, blocked
// and the last age.
func advanceStates(t *testing.T) map[string]game.AdvanceView {
	t.Helper()
	advanceStatesOnce.Do(func() { advanceStatesCache = buildAdvanceStates(t) })
	return advanceStatesCache
}

var (
	advanceStatesOnce  sync.Once
	advanceStatesCache map[string]game.AdvanceView
)

func buildAdvanceStates(t *testing.T) map[string]game.AdvanceView {
	t.Helper()
	out := map[string]game.AdvanceView{}
	_, eng, p := advanceFixture(t, "bronze_age")
	out["partway"] = p.view
	for _, age := range []string{"primitive_age", "classical_age", "quantum_age"} {
		_, e, q := advanceFixture(t, age)
		out["partway "+age] = q.view
		meetTheGate(t, e)
		q.update(e.GetState())
		out["met "+age] = q.view
	}
	meetTheGate(t, eng)
	st := eng.GetState()
	st.PendingCatastrophe = "iron_era"
	st.Harbinger = &game.HarbingerView{Name: "The Seer", When: game.WhenThisAge}
	out["blocked"] = st.AdvanceView()
	out["final"] = game.GameState{Age: "transcendent_age"}.AdvanceView()
	out["too long a gate"] = longestGate(t)
	return out
}

// longestGate is the gate that asks for the most kinds of building, nothing
// built.
func longestGate(t *testing.T) game.AdvanceView {
	t.Helper()
	ages, best, kinds := config.Ages(), "", 0
	for i := 0; i+1 < len(ages); i++ {
		if n := len(ages[i+1].BuildingReqs); n > kinds {
			best, kinds = ages[i].Key, n
		}
	}
	return game.GameState{Age: best}.AdvanceView()
}

// The panel fits every size: its frame whole, no line wider than the panel,
// the title, the march and the Charter, the foot and its keys.
func TestAdvancePanelFitsItsSizes(t *testing.T) {
	for name, v := range advanceStates(t) {
		for _, size := range advanceSizes {
			w, h := size[0], size[1]
			for sel := 0; sel < 4; sel++ {
				g := renderAdvance(v, advanceUI{sel: sel}, w, h)
				where := fmt.Sprintf("%s at %dx%d, pillar %d", name, w, h, sel)
				text := g.String()
				for y := 1; y < h-1; y++ {
					if l, r := g.at(0, y).r, g.at(w-1, y).r; l != '│' || r != '│' {
						t.Fatalf("%s: row %d has %q and %q on the frame:\n%s", where, y, l, r, text)
					}
				}
				lines := g.Lines()
				if len(lines) != h {
					t.Fatalf("%s: %d rows", where, len(lines))
				}
				for y, l := range lines {
					if runeLen(l) > w {
						t.Errorf("%s: row %d is %d cells wide", where, y, runeLen(l))
					}
				}
				title := "Advance · " + v.AgeName
				if !v.Final {
					title += " to " + v.NextName
				}
				if !strings.Contains(lines[0], title) {
					t.Errorf("%s: the title is %q, want %q", where, lines[0], title)
				}
				if !strings.Contains(text, "The march of ages") || !strings.Contains(text, v.YearText) {
					t.Errorf("%s: no march or no year %q:\n%s", where, v.YearText, text)
				}
				if !strings.Contains(lines[h-2], "Esc") {
					t.Errorf("%s: the key bar lacks Esc: %q", where, lines[h-2])
				}
				if v.Final {
					if !strings.Contains(text, "final age") {
						t.Errorf("%s: no word on the final age:\n%s", where, text)
					}
					continue
				}
				for _, p := range v.Pillars {
					if !strings.Contains(text, strings.ToUpper(p.Name)) {
						t.Errorf("%s: no %s pillar:\n%s", where, p.Name, text)
					}
				}
				if !strings.Contains(text, "% of the way to the "+v.NextName) {
					t.Errorf("%s: the year line does not say how far:\n%s", where, text)
				}
				if !strings.Contains(text, v.Pillars[sel].Name+" in detail") {
					t.Errorf("%s: the picked pillar is not in detail:\n%s", where, text)
				}
				foot := lines[h-3]
				switch {
				case v.Ready:
					if !strings.Contains(foot, "Enter") || !strings.Contains(foot, "Advance to the "+v.NextName) {
						t.Errorf("%s: a met gate's foot is %q", where, foot)
					}
				case v.AllStand:
					if !strings.Contains(foot, "'catastrophe'") {
						t.Errorf("%s: a blocked gate's foot is %q", where, foot)
					}
				default:
					if !strings.HasPrefix(strings.TrimSpace(foot), "│ Not yet:") && !strings.Contains(foot, "Not yet:") {
						t.Errorf("%s: a short gate's foot is %q", where, foot)
					}
				}
				if w >= 80 && !strings.Contains(text, " brings") {
					t.Errorf("%s: no box for what the next age brings:\n%s", where, text)
				}
			}
		}
	}
}

// The march keeps to the player's sight: a fresh game names its own age and
// the next, and no age or era past them, in any of the panel's states.
func TestAdvancePanelShowsNoSpoilers(t *testing.T) {
	engine := freshSpoilerEngine(t)
	terms := freshSpoilerTerms()
	v := engine.GetState().AdvanceView()
	for _, size := range advanceSizes {
		for sel := 0; sel < 4; sel++ {
			text := renderAdvance(v, advanceUI{sel: sel}, size[0], size[1]).String()
			checkSpoilers(t, fmt.Sprintf("Advance panel at %dx%d", size[0], size[1]), text, terms)
			if !strings.Contains(text, "Primitive") || !strings.Contains(text, "Stone") {
				t.Errorf("the march at %dx%d lacks the current and the next age:\n%s", size[0], size[1], text)
			}
		}
	}
	// A game in the Bronze Age may name the Iron Age and nothing after it.
	_, _, eng := stagedDashboard(t, "bronze_age")
	var later []spoilerTerm
	for i, key := range config.AgeOrder() {
		if i >= 4 {
			later = append(later, spoilerTerm{regexp.MustCompile(`\b` + regexp.QuoteMeta(config.AgeByKey()[key].Name) + `\b`), "a later age"})
		}
	}
	bv := eng.GetState().AdvanceView()
	for _, size := range advanceSizes {
		checkSpoilers(t, fmt.Sprintf("Advance panel, Bronze Age, %dx%d", size[0], size[1]), renderAdvance(bv, advanceUI{}, size[0], size[1]).String(), later)
	}
}

// Wording: no em dashes and no exclamation marks anywhere the panel writes.
func TestAdvancePanelWording(t *testing.T) {
	for name, v := range advanceStates(t) {
		for _, size := range advanceSizes {
			text := renderAdvance(v, advanceUI{}, size[0], size[1]).String()
			if strings.ContainsAny(text, "—!") {
				t.Errorf("%s at %dx%d has a dash or an exclamation mark:\n%s", name, size[0], size[1], text)
			}
		}
	}
	for _, s := range []string{charterStripHint.String(), advanceKeys[0].what.String(), advanceKeys[1].what.String(), advanceKeys[2].what.String()} {
		if strings.ContainsAny(s, "—!") {
			t.Errorf("wording %q", s)
		}
	}
}

// The plain glyph tier folds every cell to ASCII, drawn through the panel.
func TestAdvancePanelFoldsToASCII(t *testing.T) {
	for name, v := range advanceStates(t) {
		g := renderAdvance(v, advanceUI{sel: 1}, 100, 27)
		g.fold()
		for y, line := range g.Lines() {
			for _, r := range line {
				if r > 0x7e {
					t.Fatalf("%s: row %d keeps %q on the plain glyph tier: %q", name, y, r, line)
				}
			}
		}
	}
	_, _, p := advanceFixture(t, "bronze_age")
	p.settings = func() mapSettings { return mapSettings{Tier: mapmodel.TierASCII} }
	for _, c := range drawPanel(t, p, 100, 27) {
		for _, r := range c.Runes {
			if r > 0x7e {
				t.Fatalf("the drawn panel keeps %q on the plain glyph tier", r)
			}
		}
	}
}

// drawPanel paints the panel on a simulation screen and returns its cells.
func drawPanel(t *testing.T, p *advancePanel, w, h int) []tcell.SimCell {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(sim.Fini)
	sim.SetSize(w, h)
	screen := theme.WrapScreen(sim)
	p.SetRect(0, 0, w, h)
	p.Draw(screen)
	screen.Show()
	cells, _, _ := sim.GetContents()
	return append([]tcell.SimCell(nil), cells...)
}

// Every theme, light ones included, reads: each letter and figure the panel
// writes in a text colour has the contrast the other panels hold to, and the
// panel draws in a dark and a light theme.
func TestAdvancePanelReadsInEveryTheme(t *testing.T) {
	restoreForge(t)
	states := advanceStates(t)
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		pal := newExpPalette()
		for name, v := range states {
			g := renderAdvance(v, advanceUI{sel: 2}, 100, 27)
			for i, c := range g.c {
				if c.k == xFaint || !(c.r >= '0' && c.r <= '9' || c.r >= 'A' && c.r <= 'Z' || c.r >= 'a' && c.r <= 'z') {
					continue
				}
				fg, bg, _ := pal.styles[c.k].Decompose()
				if r := theme.ContrastRatio(fg, bg); r < 2.99 {
					t.Fatalf("theme %s, %s: %q at %d,%d has a contrast of %.2f (kind %d)", th.Key, name, c.r, i%g.w, i/g.w, r, c.k)
				}
			}
		}
	}
	for _, key := range []string{"forge", "daylight"} {
		if err := theme.SetActive(key); err != nil {
			t.Fatal(err)
		}
		_, _, p := advanceFixture(t, "bronze_age")
		for _, size := range [][2]int{{80, 21}, {100, 27}, {160, 42}} {
			cells := drawPanel(t, p, size[0], size[1])
			var text strings.Builder
			for _, c := range cells {
				if len(c.Runes) > 0 {
					text.WriteRune(c.Runes[0])
				}
			}
			for _, want := range []string{"KNOWLEDGE", "MONUMENT", "The march of ages"} {
				if !strings.Contains(text.String(), want) {
					t.Errorf("theme %s at %dx%d: the screen lacks %q", key, size[0], size[1], want)
				}
			}
		}
	}
}

// A line on the foot stays for a few seconds and then gives way to the status.
func TestAdvanceFlashGivesWay(t *testing.T) {
	_, _, p := advanceFixture(t, "bronze_age")
	clock := time.Now()
	p.now = func() time.Time { return clock }
	p.handleKey(tcell.KeyEnter)
	if !strings.Contains(strings.Join(strings.Fields(cellsText(drawPanel(t, p, 100, 27))), " "), "Not yet: ") {
		t.Fatal("the flash is not on screen")
	}
	clock = clock.Add(advFlashFor + time.Second)
	drawPanel(t, p, 100, 27)
	if p.ui.flash != "" {
		t.Errorf("the flash %q outlived its time", p.ui.flash)
	}
}

func cellsText(cells []tcell.SimCell) string {
	var b strings.Builder
	for _, c := range cells {
		if len(c.Runes) > 0 {
			b.WriteRune(c.Runes[0])
		} else {
			b.WriteRune(' ')
		}
	}
	return b.String()
}

// The Ledger's long bar says how to open the panel when it has room, and
// never at the cost of the width.
func TestLedgerPointsAtTheCharter(t *testing.T) {
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "bronze_age")
	l := ledgerOf(eng.GetState(), eng)
	wide := untag(strings.Join(ledgerLines(&l.data, ledgerFull, 160, 11, mapmodel.TierUnicode), "\n"))
	if !strings.Contains(wide, charterStripHint.String()) {
		t.Errorf("a wide Ledger has no hint:\n%s", wide)
	}
	for w := 40; w <= 200; w++ {
		for _, tier := range []ledgerTier{ledgerCompact, ledgerStandard, ledgerFull} {
			for _, line := range ledgerLines(&l.data, tier, w, 11, mapmodel.TierUnicode) {
				if n := runeLen(plainLogText(line)); n > w {
					t.Fatalf("tier %d at width %d: a line of %d cells: %q", tier, w, n, line)
				}
			}
		}
	}
	final := &ledger{}
	final.data.v = game.AdvanceView{Final: true}
	if lines := ledgerLines(&final.data, ledgerFull, 160, 11, mapmodel.TierUnicode); strings.Contains(strings.Join(lines, ""), charterStripHint.String()) {
		t.Errorf("the final age points at a panel with no gate: %v", lines)
	}
}

// The Help panel lists the command, the panel and its keys.
func TestHelpListsTheAdvancePanel(t *testing.T) {
	help := strings.Join(strings.Fields(untag(helpProvider(game.GameState{}, 100))), " ")
	for _, want := range []string{"charter", "The Advance panel", "Pick a pillar", "the same as the advance command"} {
		if !strings.Contains(help, want) {
			t.Errorf("the Help panel lacks %q", want)
		}
	}
}
