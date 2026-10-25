package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// bronzeTown is a dashboard in the Bronze Age with the given techs
// researched on top of what the age has.
func bronzeTown(t *testing.T, techs ...string) (*Dashboard, *tview.Pages) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	eng := game.NewGameEngine()
	must(t, eng.EnterAgeForTest("bronze_age"))
	if len(techs) > 0 {
		eng.GrantTechsForTest(techs...)
	}
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), eng, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	d.refresh()
	return d, pages
}

// collapse is an entry's text as a reader sees it: tags gone, every run of
// spaces and line breaks one space, so a line the box wrapped still matches.
func collapse(lines []string) string {
	return strings.Join(strings.Fields(untag(strings.Join(lines, " "))), " ")
}

// entriesAt is the Economy tab's Buildings entries at w, in list order.
func entriesAt(t *testing.T, d *Dashboard, w int) [][]string {
	t.Helper()
	_, entries := buildingEntries(d.economyTab.state, w)
	return entries
}

// isLocked reports whether an entry is one a tech still locks.
func isLocked(e []string) bool {
	return strings.Contains(strings.Join(e, "\n"), lockedMarker)
}

// altarEntry is the Altar's entry in the Economy list at w, nil when the
// list does not have it.
func altarEntry(t *testing.T, d *Dashboard, w int) []string {
	t.Helper()
	for _, e := range entriesAt(t, d, w) {
		if strings.Contains(collapse(e), "Altar") {
			return e
		}
	}
	return nil
}

// scholarLine is the scholars' line of a locked entry, "" when it has none.
func scholarLine(entry []string) string {
	text := collapse(entry)
	for _, e := range lockedHintLines {
		if p := e.String(); strings.Contains(text, p) {
			return p
		}
	}
	return ""
}

// TestLockedBuildingListedAfterTheUnlocked: in the Bronze Age with Calendar
// not researched the Altar is listed after every unlocked building of the
// age, greyed out, with the research it needs and a word from the scholars.
// With Ritual still to come the need names it too.
func TestLockedBuildingListedAfterTheUnlocked(t *testing.T) {
	d, _ := bronzeTown(t, "language")
	altar := altarEntry(t, d, 60)
	if altar == nil {
		t.Fatal("the Altar is not listed while Calendar is unresearched")
	}
	if !isLocked(altar) || !strings.Contains(altar[0], "[gray]") {
		t.Errorf("the Altar's name line is not greyed out: %q", altar[0])
	}
	text := collapse(altar)
	if !strings.Contains(text, "Needs Calendar (research), after Ritual") {
		t.Errorf("the Altar does not say what it needs:\n%s", text)
	}
	if scholarLine(altar) == "" {
		t.Errorf("the Altar has no scholar line:\n%s", text)
	}
	if strings.Contains(altar[len(altar)-1], "[comment::i]") {
		t.Errorf("the Altar keeps its flavor line:\n%s", text)
	}

	// Every unlocked building of the age is listed before the Altar.
	all := entriesAt(t, d, 60)
	altarAt := -1
	for i, e := range all {
		if isLocked(e) && altarAt < 0 && strings.Contains(collapse(e), "Altar") {
			altarAt = i
		}
	}
	for key, bs := range d.economyTab.state.Buildings {
		if !bs.Unlocked || bs.AgeKey != d.economyTab.state.Age {
			continue
		}
		for i, e := range all {
			if strings.Contains(collapse(e), bs.Name+" x") && i > altarAt {
				t.Errorf("unlocked %s (%s) is listed after the Altar", bs.Name, key)
			}
		}
	}
}

// TestLockedBuildingWithRitualResearched: once Ritual is researched the
// Altar says only what Calendar needs.
func TestLockedBuildingWithRitualResearched(t *testing.T) {
	d, _ := bronzeTown(t, "language", "ritual")
	altar := altarEntry(t, d, 60)
	if altar == nil {
		t.Fatal("the Altar is not listed while Calendar is unresearched")
	}
	text := collapse(altar)
	if !strings.Contains(text, "Needs Calendar (research)") || strings.Contains(text, "after") {
		t.Errorf("the Altar should need only Calendar now:\n%s", text)
	}
}

// TestLockedBuildingStableAndGoneWhenResearched: the scholar line is the same
// on every render, and once Calendar is researched the Altar is an ordinary
// entry with no lock and no need.
func TestLockedBuildingStableAndGoneWhenResearched(t *testing.T) {
	d, _ := bronzeTown(t, "language")
	first := scholarLine(altarEntry(t, d, 60))
	second := scholarLine(altarEntry(t, d, 60))
	if first == "" || first != second {
		t.Fatalf("the scholar line changes between renders: %q then %q", first, second)
	}

	d2, _ := bronzeTown(t, "language", "ritual", "calendar")
	for _, e := range entriesAt(t, d2, 60) {
		if !strings.Contains(collapse(e), "Altar") {
			continue
		}
		if isLocked(e) || strings.Contains(collapse(e), "Needs") {
			t.Errorf("the Altar is still locked once Calendar is researched:\n%s", collapse(e))
		}
		return
	}
	t.Error("the researched Altar is not listed")
}

// TestLockedBuildingFitsTheBox: at 80x24 and 100x30 no line of a locked
// entry is wider than the Buildings box, and where the box shows the Altar
// it shows the whole entry.
func TestLockedBuildingFitsTheBox(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {100, 30}} {
		w, h := size[0], size[1]
		d, pages := bronzeTown(t, "language")
		screen := strings.Split(drawDashboard(t, d, pages, w, h), "\n")
		x, y, bw, bh := d.economyTab.buildingTV.GetInnerRect()
		for _, e := range entriesAt(t, d, bw) {
			if !isLocked(e) {
				continue
			}
			for _, line := range e {
				if vl := visibleLen(line); vl > bw {
					t.Errorf("%dx%d: a locked entry line is %d cells in a %d-cell box: %q", w, h, vl, bw, untag(line))
				}
			}
		}
		var box []string
		for j := y; j < y+bh && j < len(screen); j++ {
			line := []rune(screen[j])
			if x >= len(line) {
				box = append(box, "")
				continue
			}
			box = append(box, strings.TrimRight(string(line[x:min(x+bw, len(line))]), " "))
		}
		if strings.Contains(strings.Join(box, "\n"), "Altar") {
			wholeEntries(t, fmt.Sprintf("%dx%d", w, h), d, box)
		}
	}
}

// TestLockedBuildingInTheOverlay: the Buildings overlay lists the Altar
// after the current age's unlocked buildings, greyed out, with its need and
// no flavor line.
func TestLockedBuildingInTheOverlay(t *testing.T) {
	d, _ := bronzeTown(t, "language")
	state := d.economyTab.state
	text := buildingsProvider(state, 100)
	at := strings.Index(text, "🔒 Altar (altar)")
	if at < 0 {
		t.Fatalf("the overlay does not list the Altar:\n%s", text)
	}
	block := text[at:]
	if end := strings.Index(block, "\n\n"); end >= 0 {
		block = block[:end]
	}
	if !strings.Contains(collapse([]string{block}), "Needs Calendar (research), after Ritual") {
		t.Errorf("the overlay does not say what the Altar needs:\n%s", block)
	}
	if strings.Contains(block, "[comment::i]") {
		t.Errorf("the overlay Altar keeps its flavor line:\n%s", block)
	}
	for key, bs := range state.Buildings {
		if key == "altar" || !bs.Unlocked || bs.AgeKey != state.Age {
			continue
		}
		if i := strings.Index(text, "]"+bs.Name+"[-]"); i > at {
			t.Errorf("unlocked %s is listed after the locked Altar", bs.Name)
		}
	}
}
