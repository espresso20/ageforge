package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

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
	wantAll(t, "legendText()", legend, "★ auto", "● active", "◆ main game", "⚠ modified", "⚠ corrupt")
	// Colours must match the row tags exactly: gold for ★, red for ⚠.
	if goldTag := "[gold]★ auto[-]"; !strings.Contains(legend, goldTag) {
		t.Errorf("legendText() missing gold-tagged %q\n%s", goldTag, legend)
	}
	wantAll(t, "legendText()", legend, "[red]⚠ modified[-]", "[red]⚠ corrupt[-]")
	// The elite easter-egg badge must NOT be advertised in the legend, even
	// though a real elite save still renders its ⭐ row tag.
	if strings.Contains(legend, "elite") {
		t.Errorf("legendText() leaks the elite easter egg\n%s", legend)
	}
	// Five symbols, five content lines (the box is seven rows tall for them).
	if lines := strings.Count(legend, "\n") + 1; lines != 5 {
		t.Errorf("legendText() has %d lines, want 5", lines)
	}
}

func TestDetailTextCorrupt(t *testing.T) {
	s := game.SaveInfo{Name: "broken", Corrupt: true, Timestamp: time.Now()}
	got := detailText(s, false, "")
	if want := "Corrupt save"; !strings.Contains(got, want) {
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
	got := detailText(s, false, "")

	// Every stat label and the title must render.
	for _, want := range []string{
		"The Eternal", "Iron Era",
		"Population", "Buildings", "Wonders",
		"Milestones", "7/33", "Techs", "Soldiers",
		"Prestige", "level 2", "1.5K points", "Morale", "75%",
		"Pending", "The Great Plague",
	} {
		if !strings.Contains(got, want) {
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
	got := detailText(s, false, "")

	// Empty title → no quoted segment leaking through.
	if strings.Contains(got, "\"\"") {
		t.Errorf("detailText with empty Title rendered empty quotes\ngot: %q", got)
	}
	// No looming catastrophe → the whole warning line is omitted.
	if strings.Contains(got, "Pending") {
		t.Errorf("detailText with empty PendingCatastrophe still shows a Pending line\ngot: %q", got)
	}
	// With no total known, Milestones shows a bare count (no slash).
	if strings.Contains(got, "Milestones") && strings.Contains(got, "1/") {
		t.Errorf("detailText showed Milestones with a total when none was set\ngot: %q", got)
	}
	// The age must still anchor line 1.
	if !strings.Contains(got, ageDisplay("stone_age")) {
		t.Errorf("detailText missing the age display for an untitled save\ngot: %q", got)
	}
}

func TestDetailTextAccountAttribution(t *testing.T) {
	const current = "abcdef0123456789abcdef0123456789"

	// Save belongs to the current account → "this account".
	mine := game.SaveInfo{Name: "mine", Age: "iron_age", AccountID: current}
	if got := detailText(mine, false, current); !strings.Contains(got, "this account") {
		t.Errorf("detailText(own save) should say 'this account'\ngot: %q", got)
	}

	// Save from a different non-empty account → "another account" + short prefix.
	const other = "9999888877776666555544443333222"
	theirs := game.SaveInfo{Name: "theirs", Age: "iron_age", AccountID: other}
	got := detailText(theirs, false, current)
	if !strings.Contains(got, "another account") {
		t.Errorf("detailText(other save) should say 'another account'\ngot: %q", got)
	}
	if !strings.Contains(got, other[:8]) {
		t.Errorf("detailText(other save) should show the short id prefix %q\ngot: %q", other[:8], got)
	}
	if strings.Contains(got, other) {
		t.Errorf("detailText must never show the full account id\ngot: %q", got)
	}

	// Pre-account / legacy save → "pre-account".
	legacy := game.SaveInfo{Name: "legacy", Age: "iron_age", AccountID: ""}
	if got := detailText(legacy, false, current); !strings.Contains(got, "pre-account") {
		t.Errorf("detailText(legacy save) should say 'pre-account'\ngot: %q", got)
	}
}

// footText is a fitted key bar as it reads: each line's keys and words.
func footText(lines []footLine) string {
	var out []string
	for _, l := range lines {
		var parts []string
		for _, k := range l {
			parts = append(parts, k[0]+" "+k[1])
		}
		out = append(out, strings.Join(parts, " "))
	}
	return strings.Join(out, "\n")
}

func TestFooterBarShowsHotkeyButtons(t *testing.T) {
	lines, _ := fitKeys(loadGameKeys, 1<<16, 1)
	bar := footText(lines)
	// Every action label must be present.
	for _, label := range []string{"Navigate", "Load", "Delete", "Rename", "Duplicate", "Main game", "Back"} {
		if !strings.Contains(bar, label) {
			t.Errorf("the foot is missing action label %q", label)
		}
	}
	for _, cap := range []string{"↑↓ ", "Enter ", "D ", "R ", "C ", "M ", "Esc "} {
		if !strings.Contains(bar, cap) {
			t.Errorf("the foot is missing keycap %q", cap)
		}
	}
	// Every key is a button that presses its key: the foot of the page, drawn.
	f := newKeyFoot(nil, loadGameKeys)
	g := f.render(120, 1)
	pressable := 0
	for _, k := range g.keys {
		if k.key != (kit.Key{}) {
			pressable++
		}
	}
	// The arrows are two keys: seven actions, eight keys.
	if pressable != len(loadGameKeys)+1 {
		t.Errorf("the foot has %d pressable keys for %d actions", pressable, len(loadGameKeys))
	}
}

// TestKeyBarsFitTheirWidth: the Load Game browser's key bar is one line at
// every width from 80 columns, and the Accounts panel's nine keys are one
// line where they fit and two where they do not. Nothing is cut off: every
// key is on the bar at every width.
func TestKeyBarsFitTheirWidth(t *testing.T) {
	for w := 80; w <= 200; w++ {
		lines, gap := fitKeys(loadGameKeys, w, 1)
		if len(lines) != 1 || lines[0].footW(gap) > w {
			t.Fatalf("Load Game key bar at %d columns: %d lines, %d cells: %q", w, len(lines), lines[0].footW(gap), footText(lines))
		}
		if len(lines[0]) != len(loadGameKeys) {
			t.Fatalf("Load Game key bar at %d columns lost a key: %q", w, footText(lines))
		}
		acc, agap := fitKeys(accountsKeys, w, 2)
		if len(acc) > 2 {
			t.Fatalf("Accounts key bar at %d columns takes %d lines", w, len(acc))
		}
		n := 0
		for _, l := range acc {
			if l.footW(agap) > w {
				t.Fatalf("Accounts key bar at %d columns: a line of %d cells: %q", w, l.footW(agap), footText(acc))
			}
			n += len(l)
		}
		if n != len(accountsKeys) {
			t.Fatalf("Accounts key bar at %d columns shows %d of %d keys: %q", w, n, len(accountsKeys), footText(acc))
		}
	}
	if lines, _ := fitKeys(loadGameKeys, 120, 1); !strings.Contains(footText(lines), "Navigate") || !strings.Contains(footText(lines), "Main game") {
		t.Errorf("with room the Load Game bar names every key in full: %q", footText(lines))
	}
	if lines, _ := fitKeys(accountsKeys, 130, 2); len(lines) != 1 {
		t.Error("the Accounts bar fits on one line at 130 columns and took two")
	}
}
