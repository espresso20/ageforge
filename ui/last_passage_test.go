package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// lastPassageDashboard builds a dashboard over an engine standing in age (a
// Cosmic Era age) with its Last Passage thread, optionally carrying the Cosmic
// Legacy, and the Last Passage pending.
func lastPassageDashboard(t *testing.T, age string, legacy bool) (*Dashboard, *game.GameEngine, *tview.Pages) {
	t.Helper()
	d, engine, pages := harbDashboard(t, age)
	engine.SetCosmicLegacyForTest(legacy)
	must(t, engine.ForceLastPassageForTest(age))
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
	screen := drawPrimitive(t, pages, 160, 50).Text()
	wantAll(t, "modal", screen,
		"✦ The Last Passage", "☄ The Last Passage", "Prestige waits on your answer",
		"ENDURE: pass through, diminished", "You keep 50% of this run's prestige points",
		"SUCCUMB: let it take the run", "no points from this run", "Cosmic Legacy: production +10%",
		"[E] ENDURE", "[S] SUCCUMB", "prestige waits · type 'catastrophe' to reopen",
	)
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
	screen := drawPrimitive(t, pages, 160, 50).Text()
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

// costItems are the materials of a price as the Harbinger panel prints them
// in the game's resource order, each with what is held ("9.6Qi titanium
// (have 0)"; the test's town holds none).
func costItems(state game.GameState, cost map[string]float64) []string {
	var items []string
	for _, def := range state.Ruleset().Resources() {
		if need, ok := cost[def.Key]; ok {
			items = append(items, game.Amount(need, def.Key)+" (have 0)")
		}
	}
	return items
}

// checkCostBlock opens the Harbinger panel over state at 80 by 24 and checks
// the nth "Next level costs" block: it begins on a row of its own, holds
// every material whole on a row (none wrapped or cut), in order, and runs
// on over consecutive rows.
func checkCostBlock(t *testing.T, d *Dashboard, pages *tview.Pages, state game.GameState, nth int, cost map[string]float64) {
	t.Helper()
	const w, h = 80, 24
	items := costItems(state, cost)
	if len(items) != len(cost) {
		t.Fatalf("the price has %d materials, the panel can name %d", len(cost), len(items))
	}
	d.overlayMgr.screenW = w
	if !d.overlayMgr.Show("harbinger", state) {
		t.Fatal("harbinger overlay not registered")
	}
	defer d.overlayMgr.Hide()
	tv := d.overlayMgr.entries["harbinger"].tv
	var at []int
	for i, line := range strings.Split(tv.GetText(true), "\n") {
		if strings.Contains(line, "Next level costs:") {
			at = append(at, i)
		}
	}
	if len(at) <= nth {
		t.Fatalf("the panel has %d price lines, want one at index %d", len(at), nth)
	}
	tv.ScrollTo(at[nth]-1, 0)
	rows := drawPrimitive(t, pages, w, h).Rows()
	first, last := -1, -1
	for _, item := range items {
		row := -1
		for i, r := range rows {
			if strings.Contains(r, item) {
				row = i
				break
			}
		}
		if row < 0 {
			t.Errorf("%q is not whole on one row of the %dx%d screen:\n%s", item, w, h, strings.Join(rows, "\n"))
			continue
		}
		if first < 0 {
			first = row
		}
		if row < last {
			t.Errorf("%q is out of order on the screen", item)
		}
		last = row
	}
	if first >= 0 && !strings.Contains(rows[first], "Next level costs:") {
		t.Errorf("the first material is not on the row that says what it is: %q", rows[first])
	}
	if first >= 0 {
		for _, r := range rows[first : last+1] {
			has := false
			for _, item := range items {
				has = has || strings.Contains(r, item)
			}
			if !has {
				t.Errorf("a row inside the price holds no material: %q", r)
			}
		}
	}
}

// The Cosmic Era panel warns of the Last Passage, talks in points, and prints
// the odds; once it has come, the answers are closed.
func TestHarbingerPanelLastPassage(t *testing.T) {
	d, engine, pages := harbDashboard(t, "interstellar_age")
	state := engine.GetState()
	txt := untag(harbingerPanelText(state, "", false, false))
	t.Logf("Cosmic Era harbinger panel:\n%s", txt)
	// The prices are the thread's own, as the engine has them (the game's
	// tests pin the figures): the panel prints them in full, with what is
	// held.
	h := state.Harbinger
	if h == nil || h.AppeaseCost["faith"] <= 0 || h.AppeaseCost["culture"] <= 0 || h.BraceCost["titanium"] <= 0 || h.BraceCost["dark_matter"] <= 0 {
		t.Fatalf("the Last Passage's thread has no prices to print: %+v", h)
	}
	wantAll(t, "panel", txt,
		"The Distress Beacon", "Warning of the Last Passage: the end of this civilization, when you next prestige.",
		"Published odds:", "%", "If it comes and you Endure: you keep 50% of the run's prestige points.",
		"Next level: 70% kept.", "Guarantees the Last Passage at your next prestige",
	)
	if strings.Contains(txt, "passage into the") || strings.Contains(txt, "buildings fall") {
		t.Errorf("panel talks about an epoch passage:\n%s", txt)
	}

	// The prices are printed in full whatever the number of materials: the
	// thread's own two, and four, each on one row of an 80 by 24 screen.
	four := map[string]float64{"titanium": 9.6 * config.Quintillion, "dark_matter": 11 * config.Quintillion, "plasma": 11.8 * config.Quintillion, "nanobots": 7.3 * config.Quintillion}
	for _, tc := range []struct {
		name  string
		brace map[string]float64
	}{{"two materials", h.BraceCost}, {"four materials", four}} {
		t.Run(tc.name, func(t *testing.T) {
			st := engine.GetState()
			hv := *st.Harbinger
			hv.BraceCost = tc.brace
			st.Harbinger = &hv
			checkCostBlock(t, d, pages, st, 0, hv.AppeaseCost)
			checkCostBlock(t, d, pages, st, 1, hv.BraceCost)
		})
	}

	must(t, engine.ForceLastPassageForTest("interstellar_age"))
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
	screen := drawPrimitive(t, pages, 160, 50).Text()
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
	wantAll(t, "confirm", confirm,
		"In the Cosmic Era prestige can bring the Last Passage", "If it comes, prestige waits for your choice",
		"Endure: keep 50% of this run's points", "Succumb: no points from this run, and the Cosmic Legacy (+10% production, permanent).",
	)

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
	// It is applied after the pools: a multiplier of its own beside the
	// all-production pool, however full the pool is (here milestone rewards
	// put it far past +200%).
	engine.GrantTechsForTest()
	engine.GrantBonusForTest("production_all", 4)
	st := engine.GetState()
	if p := st.Pools["production_all"]; p.Earned <= 2 || p.Limited {
		t.Fatalf("the all-production pool is not far past +200%% and in full in the Quantum Age with +400%% of milestone rewards: %+v", p)
	}
	out = untag(statsProvider(st, 140))
	wantAll(t, "the Stats panel", out,
		"Cosmic Legacy:   all production +10%, counted in full after every other bonus (permanent, through every prestige)\n",
		"Cosmic Legacy ×1.10",
	)
	if r := untag(HandleCommand("rates", engine).Message); !strings.Contains(r, "Cosmic Legacy: +") {
		t.Errorf("rates has no Cosmic Legacy part:\n%s", r)
	}
}

// harbingerPanelText renders the panel for the narrowest screen the game
// draws. Pure: state in, text out.
func harbingerPanelText(state game.GameState, note string, noteGood, inviteArmed bool) string {
	return harbingerPanelTextAt(state, narrowestScreenW, note, noteGood, inviteArmed)
}

// narrowestScreenW is the narrowest terminal the game is laid out for.
const narrowestScreenW = 80
