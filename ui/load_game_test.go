package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

func TestCommafy(t *testing.T) {
	cases := map[int]string{
		0:       "0",
		7:       "7",
		42:      "42",
		999:     "999",
		1000:    "1,000",
		12400:   "12,400",
		100000:  "100,000",
		1234567: "1,234,567",
		-1500:   "-1,500",
		-12:     "-12",
	}
	for in, want := range cases {
		if got := commafy(in); got != want {
			t.Errorf("commafy(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		t    time.Time
		want string
	}{
		{"zero", time.Time{}, "unknown"},
		{"just now", now.Add(-10 * time.Second), "just now"},
		{"minutes", now.Add(-5 * time.Minute), "5m ago"},
		{"hours", now.Add(-2 * time.Hour), "2h ago"},
		{"yesterday", now.Add(-30 * time.Hour), "yesterday"},
		{"days", now.Add(-72 * time.Hour), "3 days ago"},
		{"future", now.Add(10 * time.Minute), "just now"},
	}
	for _, c := range cases {
		if got := relativeTime(c.t); got != c.want {
			t.Errorf("relativeTime(%s) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestRowTagPrecedence(t *testing.T) {
	// Autosave wins even if modified/elite flags are also set.
	auto := game.SaveInfo{Name: game.AutosaveName, Modified: true, Elite: true}
	if got := rowTag(auto); got != "[gold]★ auto[-]" {
		t.Errorf("rowTag(autosave) = %q, want auto tag", got)
	}
	// Modified outranks elite for a player-named save.
	mod := game.SaveInfo{Name: "hero", Modified: true, Elite: true}
	if got := rowTag(mod); got != "[red]⚠ modified[-]" {
		t.Errorf("rowTag(modified) = %q, want modified tag", got)
	}
	// Elite when neither auto nor modified.
	elite := game.SaveInfo{Name: "hero", Elite: true}
	if got := rowTag(elite); got != "[gold]⭐ elite[-]" {
		t.Errorf("rowTag(elite) = %q, want elite tag", got)
	}
	// Plain save → no tag.
	plain := game.SaveInfo{Name: "hero"}
	if got := rowTag(plain); got != "" {
		t.Errorf("rowTag(plain) = %q, want empty", got)
	}
}

func TestLegendTextCoversSymbolsWithMatchingColors(t *testing.T) {
	legend := legendText()
	// Every player-facing row symbol must be explained in the Key box.
	for _, sym := range []string{"★ auto", "⚠ modified", "⚠ corrupt"} {
		if !strings.Contains(legend, sym) {
			t.Errorf("legendText() missing symbol %q\n%s", sym, legend)
		}
	}
	// Colours must match the row tags exactly: gold for ★, red for ⚠.
	if goldTag := "[gold]★ auto[-]"; !strings.Contains(legend, goldTag) {
		t.Errorf("legendText() missing gold-tagged %q\n%s", goldTag, legend)
	}
	for _, redTag := range []string{"[red]⚠ modified[-]", "[red]⚠ corrupt[-]"} {
		if !strings.Contains(legend, redTag) {
			t.Errorf("legendText() missing red-tagged %q\n%s", redTag, legend)
		}
	}
	// The elite easter-egg badge must NOT be advertised in the legend, even
	// though a real elite save still renders its ⭐ row tag.
	if strings.Contains(legend, "elite") {
		t.Errorf("legendText() leaks the elite easter egg\n%s", legend)
	}
	// Four symbols → four content lines (sizing assumes this for the height-6 box).
	if lines := strings.Count(legend, "\n") + 1; lines != 4 {
		t.Errorf("legendText() has %d lines, want 4", lines)
	}
}

func TestDetailTextCorrupt(t *testing.T) {
	s := game.SaveInfo{Name: "broken", Corrupt: true, Timestamp: time.Now()}
	got := detailText(s, false)
	if want := "Corrupt save"; !contains(got, want) {
		t.Errorf("detailText(corrupt) = %q, should mention %q", got, want)
	}
}

func TestDetailTextPopulated(t *testing.T) {
	s := game.SaveInfo{
		Name:               "hero",
		Timestamp:          time.Now(),
		Age:                "iron_age",
		Tick:               4242,
		PrestigeLevel:      2,
		PrestigeTotal:      1500,
		Morale:             0.75,
		Title:              "The Eternal",
		Epoch:              "Iron Era",
		Population:         10,
		Buildings:          120,
		Wonders:            3,
		Techs:              4,
		Soldiers:           137,
		PendingCatastrophe: "The Great Plague",
		MilestonesDone:     7,
		MilestonesTotal:    33,
	}
	got := detailText(s, false)

	// Every stat label and the title must render.
	for _, want := range []string{
		"The Eternal", "Iron Era",
		"Population", "Buildings", "Wonders",
		"Milestones", "7/33", "Techs", "Soldiers",
		"Prestige", "Lv 2", "1,500 pts", "Morale", "75%",
		"Pending", "The Great Plague",
	} {
		if !contains(got, want) {
			t.Errorf("detailText(populated) missing %q\ngot: %q", want, got)
		}
	}
}

func TestDetailTextOmitsEmptyTitleAndCatastrophe(t *testing.T) {
	s := game.SaveInfo{
		Name:       "hero",
		Timestamp:  time.Now(),
		Age:        "stone_age",
		Population: 5,
		Buildings:  3,
		// No Title, no PendingCatastrophe, no MilestonesTotal.
		MilestonesDone: 1,
	}
	got := detailText(s, false)

	// Empty title → no quoted segment leaking through.
	if contains(got, "\"\"") {
		t.Errorf("detailText with empty Title rendered empty quotes\ngot: %q", got)
	}
	// No looming catastrophe → the whole warning line is omitted.
	if contains(got, "Pending") {
		t.Errorf("detailText with empty PendingCatastrophe still shows a Pending line\ngot: %q", got)
	}
	// With no total known, Milestones shows a bare count (no slash).
	if contains(got, "Milestones") && contains(got, "1/") {
		t.Errorf("detailText showed Milestones with a total when none was set\ngot: %q", got)
	}
	// The age must still anchor line 1.
	if !contains(got, ageDisplay("stone_age")) {
		t.Errorf("detailText missing the age display for an untitled save\ngot: %q", got)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestFooterBarShowsHotkeyButtons(t *testing.T) {
	bar := footerBar()
	// Every action label must be present.
	for _, label := range []string{"Navigate", "Load", "Delete", "Rename", "Duplicate", "Back"} {
		if !contains(bar, label) {
			t.Errorf("footerBar() missing action label %q", label)
		}
	}
	// Every hotkey must render as a keycap (between the gold cap tag and the label tag).
	for _, cap := range []string{"] ↑↓ [", "] Enter [", "] D [", "] R [", "] C [", "] Esc ["} {
		if !contains(bar, cap) {
			t.Errorf("footerBar() missing keycap %q", cap)
		}
	}
	// Buttons must be styled with a real color tag, not plain text.
	if !contains(bar, "[black:gold:b]") {
		t.Errorf("footerBar() should style keycaps with a color tag, got %q", bar)
	}
	// Regression guard: the old footer used bare [Enter]/[d]/[r]/[c]/[Esc], which
	// tview parses as color tags and SWALLOWS — the hotkeys vanished on screen.
	for _, naked := range []string{"[Enter]", "[Esc]", "[d]", "[r]", "[c]"} {
		if contains(bar, naked) {
			t.Errorf("footerBar() contains swallowable tag %q — keys will vanish in tview", naked)
		}
	}
}
