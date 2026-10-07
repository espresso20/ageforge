package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// lastPassageDashboard builds a dashboard over an engine standing in age (a
// Cosmic Era age) with its Last Passage thread, optionally carrying the Cosmic
// Legacy, and the Last Passage pending.
func lastPassageDashboard(t *testing.T, age string, legacy bool) (*Dashboard, *game.GameEngine, *tview.Pages) {
	t.Helper()
	d, engine, pages := harbDashboard(t, age)
	engine.SetCosmicLegacyForTest(legacy)
	if err := engine.ForceLastPassageForTest(age); err != nil {
		t.Fatal(err)
	}
	return d, engine, pages
}

// The Last Passage variant: its own title and text, the modal pops once on
// refresh, Esc hides it, the badge stays, and `catastrophe` reopens it.
func TestLastPassageModalBadgeAndReopen(t *testing.T) {
	d, engine, pages := lastPassageDashboard(t, "galactic_age", false)
	d.refresh()
	if !pages.HasPage(catastrophePage) {
		t.Fatal("refresh did not show the Last Passage modal")
	}
	screen := renderPages(t, pages, 160, 50)
	t.Logf("Last Passage modal:\n%s", modalCapture(screen, "The Last Passage"))
	for _, want := range []string{
		"✦ The Last Passage", "☄ The Last Passage", "Prestige waits on your answer",
		"ENDURE: pass through, diminished", "You keep 50% of this run's prestige points",
		"SUCCUMB: let it take the run", "no points from this run", "Cosmic Legacy: production +10%",
		"[E] ENDURE", "[S] SUCCUMB", "prestige waits · type 'catastrophe' to reopen",
	} {
		if !strings.Contains(screen, want) {
			t.Errorf("modal missing %q\n%s", want, screen)
		}
	}
	if strings.Contains(screen, "buildings destroyed") {
		t.Error("the Last Passage modal shows the epoch catastrophe's Endure terms")
	}
	if !strings.Contains(d.statusTV.GetText(true), "☄ Last Passage. Type catastrophe to choose.") {
		t.Errorf("no Last Passage badge: %q", d.statusTV.GetText(true))
	}

	// Esc hides it; the next refresh does not re-pop it; the badge stays.
	_, front := pages.GetFrontPage()
	front.InputHandler()(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone), func(tview.Primitive) {})
	if pages.HasPage(catastrophePage) {
		t.Fatal("Esc did not close the modal")
	}
	d.refresh()
	if pages.HasPage(catastrophePage) {
		t.Error("refresh re-popped a modal the player closed")
	}
	if !strings.Contains(d.statusTV.GetText(true), "Last Passage") {
		t.Error("badge gone after Esc")
	}

	// Nothing but prestige is blocked: the command reopens, prestige refuses.
	r := HandleCommand("catastrophe", engine)
	if !r.OpenCatastrophe {
		t.Fatalf("catastrophe → %+v, want reopen", r)
	}
	if !d.reopenCatastropheModal() || !pages.HasPage(catastrophePage) {
		t.Fatal("reopen did not show the modal")
	}
	if r := HandleCommand("prestige confirm yes", engine); r.Type != "error" || !strings.Contains(r.Message, "Last Passage") {
		t.Errorf("prestige while pending → %+v", r)
	}

	// E endures: prestige completes and the modal is gone.
	_, front = pages.GetFrontPage()
	front.InputHandler()(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone), func(tview.Primitive) {})
	st := engine.GetState()
	if st.LastPassage.Pending || st.Prestige.Level != 1 || st.Age != "primitive_age" {
		t.Fatalf("after E: pending %v level %d age %s", st.LastPassage.Pending, st.Prestige.Level, st.Age)
	}
	d.refresh()
	if pages.HasPage(catastrophePage) || strings.Contains(d.statusTV.GetText(true), "Last Passage") {
		t.Error("modal or badge survived the choice")
	}
}

// With the Cosmic Legacy held, Succumb is shown closed, with the reason, and
// neither S nor Tab can reach it.
func TestLastPassageSuccumbClosedWhenLegacyHeld(t *testing.T) {
	d, engine, pages := lastPassageDashboard(t, "quantum_age", true)
	d.refresh()
	screen := renderPages(t, pages, 160, 50)
	t.Logf("Last Passage modal, legacy held:\n%s", modalCapture(screen, "The Last Passage"))
	if !strings.Contains(screen, "You already carry the Cosmic Legacy. Succumb is closed to you.") {
		t.Errorf("no reason for the closed Succumb\n%s", screen)
	}
	if strings.Contains(screen, "Cosmic Legacy: production +10%") {
		t.Error("offers the legacy the player already holds")
	}
	_, front := pages.GetFrontPage()
	front.InputHandler()(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone), func(tview.Primitive) {})
	if !engine.GetState().LastPassage.Pending || !pages.HasPage(catastrophePage) {
		t.Fatal("S acted on a closed Succumb")
	}
	front.InputHandler()(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), func(tview.Primitive) {})
	if b, ok := d.catFocus.(*tview.Button); !ok || b.GetLabel() != tview.Escape("[E] ENDURE") {
		t.Errorf("Tab moved focus off Endure: %v", d.catFocus)
	}
}

// modalCapture returns the rows of screen from the one holding title down to
// the box's bottom border, for the test log.
func modalCapture(screen, title string) string {
	rows := strings.Split(screen, "\n")
	start := -1
	for i, r := range rows {
		if strings.Contains(r, title) {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	// Cut each row to the box's columns: from the top-left corner to the
	// matching right edge.
	top := []rune(rows[start])
	left, right := -1, -1
	for i, r := range top {
		if r == '╔' && left < 0 {
			left = i
		}
		if r == '╗' {
			right = i
		}
	}
	if left < 0 || right < 0 {
		return ""
	}
	var out []string
	for _, row := range rows[start:] {
		rr := []rune(row)
		if len(rr) <= right {
			break
		}
		out = append(out, string(rr[left:right+1]))
		if rr[left] == '╚' {
			break
		}
	}
	return strings.Join(out, "\n")
}

// The layout fits the box at every Brace level, legacy held or not.
func TestLastPassageModalLayoutFitsWidth(t *testing.T) {
	inner := catastropheModalWidth - 2
	for brace := 0; brace <= game.HarbingerMaxBrace; brace++ {
		for _, legacy := range []bool{false, true} {
			lp := game.LastPassageState{Pending: true, CosmicLegacy: legacy, BraceLevel: brace, KeepPct: 85, PointsNow: 12345, PointsIfEndured: 10493}
			l := buildLastPassageModalLayout(lp)
			rows := 0
			for _, block := range []string{l.header, l.endure, l.succumb, l.hint} {
				for _, line := range strings.Split(block, "\n") {
					rows++
					if w := tview.TaggedStringWidth(line); w > inner {
						t.Errorf("brace %d legacy %v: %q is %d cells, box interior is %d", brace, legacy, line, w, inner)
					}
				}
			}
			if want := rows + 3 + 1 + 2; l.height != want {
				t.Errorf("brace %d legacy %v: height %d, want %d", brace, legacy, l.height, want)
			}
		}
	}
}

// The Cosmic Era panel warns of the Last Passage, talks in points, and prints
// the odds; once it has come, the answers are closed.
func TestHarbingerPanelLastPassage(t *testing.T) {
	d, engine, pages := harbDashboard(t, "interstellar_age")
	txt := untag(harbingerPanelText(engine.GetState(), "", false, false))
	t.Logf("Cosmic Era harbinger panel:\n%s", txt)
	for _, want := range []string{
		"The Distress Beacon", "Warning of the Last Passage: the end of this civilization, when you next prestige.",
		"Published odds:", "%", "If it comes and you Endure: you keep 50% of the run's prestige points.",
		"Next level: 70% kept.", "Guarantees the Last Passage at your next prestige",
		"Next level costs: 1.2B faith (have 0), 18B culture (have 0)",
		"Next level costs: 24Q titanium (have 0), 20Q dark matter (have 0)",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("panel missing %q:\n%s", want, txt)
		}
	}
	if strings.Contains(txt, "passage into the") || strings.Contains(txt, "buildings fall") {
		t.Errorf("panel talks about an epoch passage:\n%s", txt)
	}

	if err := engine.ForceLastPassageForTest("interstellar_age"); err != nil {
		t.Fatal(err)
	}
	txt = untag(harbingerPanelText(engine.GetState(), "", false, false))
	t.Logf("Cosmic Era harbinger panel, Last Passage pending:\n%s", txt)
	if !strings.Contains(txt, "The Last Passage has come. Prestige waits") ||
		!strings.Contains(txt, "Unavailable: the Last Passage has already come") {
		t.Errorf("pending panel:\n%s", txt)
	}
	if r := HandleCommand("harbinger brace", engine); r.Type != "error" || !strings.Contains(r.Message, "already come") {
		t.Errorf("brace while pending → %+v", r)
	}

	// On screen, through the overlay, like a player opens it.
	d.refresh()
	d.closeCatastropheModal()
	if !d.overlayMgr.Show("harbinger", engine.GetState()) {
		t.Fatal("harbinger overlay not registered")
	}
	screen := renderText(t, pages, 160, 50)
	if !strings.Contains(screen, "Warning of the Last Passage") {
		t.Errorf("overlay not on screen:\n%s", screen)
	}
}

// `prestige` mentions the Last Passage risk only in the Cosmic Era, with the
// figure; `prestige confirm` explains what happens if it goes against you.
func TestPrestigeCommandLastPassageText(t *testing.T) {
	_, engine, _ := harbDashboard(t, "modern_age")
	for _, cmd := range []string{"prestige", "prestige confirm"} {
		if r := HandleCommand(cmd, engine); strings.Contains(r.Message, "Last Passage") {
			t.Errorf("%q before the Cosmic Era mentions the Last Passage:\n%s", cmd, r.Message)
		}
	}
	if r := HandleCommand("prestige", engine); !strings.Contains(r.Message, "Reach the Modern Age") && !strings.Contains(r.Message, "You can prestige now") {
		t.Errorf("status:\n%s", r.Message)
	}

	_, engine, _ = harbDashboard(t, "galactic_age")
	status := untag(HandleCommand("prestige", engine).Message)
	t.Logf("prestige (Cosmic Era):\n%s", status)
	if !strings.Contains(status, "☄ The Last Passage: 18% chance (high) when you prestige.") || !strings.Contains(status, "The Elder Relay is warning of it.") {
		t.Errorf("status:\n%s", status)
	}
	confirm := untag(HandleCommand("prestige confirm", engine).Message)
	t.Logf("prestige confirm (Cosmic Era):\n%s", confirm)
	for _, want := range []string{
		"In the Cosmic Era prestige can bring the Last Passage", "If it comes, prestige waits for your choice",
		"Endure: keep 50% of this run's points", "Succumb: no points from this run, and the Cosmic Legacy (+10% production, permanent).",
	} {
		if !strings.Contains(confirm, want) {
			t.Errorf("confirm missing %q:\n%s", want, confirm)
		}
	}

	engine.SetCosmicLegacyForTest(true)
	if r := untag(HandleCommand("prestige confirm", engine).Message); !strings.Contains(r, "Succumb is closed: you already carry the Cosmic Legacy.") {
		t.Errorf("confirm with the legacy held:\n%s", r)
	}
	if r := untag(HandleCommand("prestige", engine).Message); !strings.Contains(r, "Cosmic Legacy: +10% production") {
		t.Errorf("status with the legacy held:\n%s", r)
	}

	// Invited: prestige confirm yes brings it and asks the dashboard to open the choice.
	if r := HandleCommand("harbinger invite", engine); r.Type != "success" {
		t.Fatalf("invite → %+v", r)
	}
	if r := untag(HandleCommand("prestige confirm", engine).Message); !strings.Contains(r, "You invited it. It will come.") {
		t.Errorf("confirm after invite:\n%s", r)
	}
	r := HandleCommand("prestige confirm yes", engine)
	if !r.OpenCatastrophe || !strings.Contains(r.Message, "The Last Passage has come") || !engine.GetState().LastPassage.Pending {
		t.Errorf("confirm yes → %+v", r)
	}
	if st := untag(HandleCommand("prestige", engine).Message); !strings.Contains(st, "The Last Passage has come. Prestige waits") {
		t.Errorf("status while pending:\n%s", st)
	}
}

// The outlook surfaces name the prestige passage in the Cosmic Era.
func TestOutlookSurfacesNameThePrestigePassage(t *testing.T) {
	_, engine, _ := harbDashboard(t, "quantum_age")
	if r := HandleCommand("catastrophe", engine); !strings.Contains(r.Message, "Next passage (prestige, the Last Passage)") {
		t.Errorf("catastrophe outlook:\n%s", r.Message)
	}
	state := engine.GetState()
	var sb strings.Builder
	epochProviderCurrentEpoch(&sb, state)
	if out := untag(sb.String()); !strings.Contains(out, "Next passage (prestige, the Last Passage)") || !strings.Contains(out, "waiting for your prestige") {
		t.Errorf("epoch overlay:\n%s", out)
	}
	if got := multiplierSourceLabel("cosmic_legacy"); got != "Cosmic Legacy" {
		t.Errorf("multiplier label = %q", got)
	}
	engine.SetCosmicLegacyForTest(true)
	out := untag(statsProvider(engine.GetState(), 140))
	if !strings.Contains(out, "Cosmic Legacy:") || !strings.Contains(out, "Cosmic Legacy") {
		t.Errorf("stats overlay without the Cosmic Legacy:\n%s", out)
	}
	// It is applied after the caps: a multiplier of its own beside the
	// all-production pool, and never tagged as capped, however full the
	// pool is (here milestone rewards put it far past the cap).
	engine.GrantTechsForTest()
	engine.GrantBonusForTest("production_all", 4)
	st := engine.GetState()
	if p := st.Pools["production_all"]; !p.Limited {
		t.Fatalf("the all-production pool is not capped in the Quantum Age with +400%% of milestone rewards: %+v", p)
	}
	out = untag(statsProvider(st, 140))
	for _, want := range []string{
		"Cosmic Legacy:   all production +10%, counted after the caps (permanent, through every prestige)\n",
		"Cosmic Legacy ×1.10",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the Stats panel does not show %q:\n%s", want, out)
		}
	}
	if r := untag(HandleCommand("rates", engine).Message); !strings.Contains(r, "Cosmic Legacy: +") {
		t.Errorf("rates has no Cosmic Legacy part:\n%s", r)
	}
}
