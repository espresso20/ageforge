package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// panelSection returns the lines of the milestones panel section headed by
// name, up to the next heading.
func panelSection(shown, name string) []string {
	var out []string
	in := false
	for _, line := range strings.Split(shown, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "◆ ") || strings.HasPrefix(trimmed, "★ ") {
			if in {
				break
			}
			in = strings.HasPrefix(trimmed[len("◆ "):], name)
			continue
		}
		if in && trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// TestMilestonesPanelListsEveryCategory: the panel's Progress total counts
// every milestone, so every milestone has to be listed under a heading. The
// epoch milestones were counted and never listed.
func TestMilestonesPanelListsEveryCategory(t *testing.T) {
	inOrder := map[string]bool{}
	names := config.MilestoneCategoryNames()
	for _, cat := range config.MilestoneCategoryOrder() {
		inOrder[cat] = true
		if names[cat] == "" {
			t.Errorf("milestone category %q has no display name", cat)
		}
	}

	var st game.GameState
	st.Milestones.Milestones = map[string]game.MilestoneInfo{}
	for _, m := range config.Milestones() {
		if !inOrder[m.Category] {
			t.Errorf("milestone %s is in category %q, which config.MilestoneCategoryOrder leaves out", m.Key, m.Category)
		}
		st.Milestones.Milestones[m.Key] = game.MilestoneInfo{
			Name: m.Name, Category: m.Category, Visible: true, Completed: true,
		}
	}
	st.Milestones.TotalCount = len(st.Milestones.Milestones)
	st.Milestones.CompletedCount = st.Milestones.TotalCount

	shown := guardRendered(milestonesProvider(st, guardPanelWidth))
	for _, m := range config.Milestones() {
		found := false
		for _, line := range panelSection(shown, names[m.Category]) {
			if line == "✓ "+m.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s (%s) is not listed under %s:\n%s", m.Name, m.Key, names[m.Category], shown)
		}
	}
}

// TestMilestonesPanelEpochGroup: a new game lists the epoch milestones it can
// see in their own group and counts the rest as hidden without naming them.
// Age Hopper needs the Classical Age, which a Primitive Age player cannot see
// named yet (game/spoilers.go), so it stays hidden though it is not marked
// Hidden.
func TestMilestonesPanelEpochGroup(t *testing.T) {
	engine := freshSpoilerEngine(t)
	shown := guardRendered(milestonesProvider(engine.GetState(), guardPanelWidth))

	epoch := panelSection(shown, "Epoch")
	if len(epoch) == 0 {
		t.Fatalf("no Epoch group in the milestones panel:\n%s", shown)
	}
	body := strings.Join(epoch, "\n")
	for _, name := range []string{"First Farmers", "Survivor"} {
		if !strings.Contains(body, "○ "+name) {
			t.Errorf("the Epoch group should list %s:\n%s", name, body)
		}
	}
	for _, name := range []string{"Age Hopper", "Enduring Civilization", "Industrial Titan", "Power Grid"} {
		if strings.Contains(shown, name) {
			t.Errorf("a new game's milestones panel names %s:\n%s", name, shown)
		}
	}
	if !strings.Contains(body, "+ 4 hidden milestones") {
		t.Errorf("the Epoch group should count its 4 hidden milestones:\n%s", body)
	}
	checkSpoilers(t, "milestones panel", shown, freshSpoilerTerms())
}
