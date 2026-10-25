package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/vibe"
)

// screenPanels are the panels a player opens by typing their name. The map is
// left out: it moves on the clock.
var screenPanels = []string{"help", "milestones", "research", "badges", "army", "expedition", "trade",
	"factions", "stats", "wonders", "people", "charter", "logs", "epoch", "harbinger", "plan", "history", "buildings"}

// TestTheScreensSwitchAtOnce: every text of the screens is read when it is
// drawn, so the switch is not a restart. With a rewriter in place of the script
// (every text the code writes itself reads in guillemets), the dashboard and
// each panel are drawn with VibeMode off, then on, then off again, from the one
// game that is never rebuilt. On changes every one of them; off again gives back
// exactly what the first draw gave.
func TestTheScreensSwitchAtOnce(t *testing.T) {
	vibeOffAfter(t)
	defer vibe.UseRewriterForTest(func(id, plain string) string { return "«" + plain + "»" })()
	d, pages, eng := badgeDashboard(t)
	d.runForTest("motion off")
	// The switch is thrown on the engine, not typed: a typed command writes a line
	// in the log, which the screens would then differ by.
	set := func(on bool) {
		t.Helper()
		must(t, eng.SetVibeMode(on))
		d.refresh()
	}

	draw := func(cmd string) string {
		t.Helper()
		if cmd != "" {
			d.runForTest(cmd)
		}
		d.refresh()
		got := drawDashboard(t, d, pages, 160, 50)
		d.overlayMgr.Hide()
		return got
	}
	shots := func() map[string]string {
		out := map[string]string{"": draw("")}
		for _, p := range screenPanels {
			out[p] = draw(p)
		}
		return out
	}

	shots() // the first frames learn the screen's size and where the focus goes; the ones after it are what is compared
	plain := shots()
	for p, text := range plain {
		if strings.Contains(text, "«") {
			t.Errorf("%q: the screen has a rewritten text with VibeMode off", p)
		}
	}

	set(true)
	if !vibe.On() {
		t.Fatal("SetVibeMode(true) did not turn it on")
	}
	on := shots()
	for p, text := range on {
		name := p
		if name == "" {
			name = "the dashboard"
		}
		if text == plain[p] || !strings.Contains(text, "«") {
			t.Errorf("%s: VibeMode is on and the screen reads as it did", name)
		}
	}

	set(false)
	if vibe.On() {
		t.Fatal("SetVibeMode(false) did not turn it off")
	}
	back := shots()
	for p, text := range back {
		if text != plain[p] {
			t.Errorf("%q: with VibeMode off again the screen is not what it was: %s", p, firstDifference(plain[p], text))
		}
	}
}

// TestTheBoxTitlesFollowTheSwitch: the titles on the borders of the boxes the
// dashboard makes once (the Panels list, the log, the command line, the three
// boxes of the economy tab, every panel's overlay) are written again when the
// switch moves, without a panel being opened.
func TestTheBoxTitlesFollowTheSwitch(t *testing.T) {
	vibeOffAfter(t)
	defer vibe.UseRewriterForTest(func(id, plain string) string { return "«" + plain + "»" })()
	d, _, eng := badgeDashboard(t)
	set := func(on bool) {
		t.Helper()
		must(t, eng.SetVibeMode(on))
		d.refresh()
	}
	titles := func() []string {
		out := []string{d.sidebar.GetTitle(), d.logTV.GetTitle(), d.inputField.GetTitle(),
			d.economyTab.resourceTV.GetTitle(), d.economyTab.buildingTV.GetTitle(), d.economyTab.constructionTV.GetTitle()}
		for _, name := range []string{"help", "milestones", "army", "trade", "factions", "stats", "wonders", "logs", "epoch", "harbinger", "plan", "history", "buildings"} {
			out = append(out, d.overlayMgr.entries[name].tv.GetTitle())
		}
		return out
	}
	plain := titles()
	for i, s := range plain {
		if s == "" || strings.Contains(s, "«") {
			t.Errorf("title %d reads %q with VibeMode off", i, s)
		}
	}
	set(true)
	for i, s := range titles() {
		if !strings.Contains(s, "«") {
			t.Errorf("title %d (%q) did not change with VibeMode on", i, plain[i])
		}
	}
	set(false)
	for i, s := range titles() {
		if s != plain[i] {
			t.Errorf("title %d reads %q with VibeMode off again, was %q", i, s, plain[i])
		}
	}
}

// firstDifference names the first line two screens differ in.
func firstDifference(was, now string) string {
	a, b := strings.Split(was, "\n"), strings.Split(now, "\n")
	for i := 0; i < len(a) || i < len(b); i++ {
		var x, y string
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		if x != y {
			ra, rb := []rune(x), []rune(y)
			at := 0
			for at < len(ra) && at < len(rb) && ra[at] == rb[at] {
				at++
			}
			from := max(0, at-12)
			return fmt.Sprintf("line %d, from column %d: was %q, now %q", i+1, at+1, string(ra[from:min(len(ra), at+30)]), string(rb[from:min(len(rb), at+30)]))
		}
	}
	return "no difference"
}
