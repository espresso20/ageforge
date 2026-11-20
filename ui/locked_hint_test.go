package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// bronzeTown is a dashboard in the Bronze Age with the given techs
// researched on top of what the age has.
func bronzeTown(t *testing.T, techs ...string) (*Dashboard, *nativePages) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	eng := game.NewGameEngine()
	must(t, eng.EnterAgeForTest("bronze_age"))
	if len(techs) > 0 {
		eng.GrantTechsForTest(techs...)
	}
	d, pages := newTestDashboard(eng)
	d.refresh()
	return d, pages
}

// collapse is an entry's text as a reader sees it: tags gone, every run of
// spaces and line breaks one space, so a line the box wrapped still matches.
func collapse(lines []string) string {
	return strings.Join(strings.Fields(untag(strings.Join(lines, " "))), " ")
}

// altarCard is the Altar's card in the Buildings box, nil when it has none.
func altarCard(d *Dashboard) *buildCard {
	for _, c := range d.economyTab.buildBox.cards {
		if c.key == "altar" {
			return &c
		}
	}
	return nil
}

// TestLockedBuildingIsACardWithItsNeed: in the Bronze Age with Calendar not
// researched the Altar is a card that cannot be built, whose reason is the need
// the overlay names, with no wait, in the last group and with its Build off.
func TestLockedBuildingIsACardWithItsNeed(t *testing.T) {
	d, pages := bronzeTown(t, "language")
	altar := altarCard(d)
	if altar == nil {
		t.Fatal("the Altar has no card while Calendar is unresearched")
	}
	if altar.canBuild || altar.group != groupLater || altar.wait != "" {
		t.Errorf("the Altar card: can build %v, group %d, wait %q; want a card that cannot be built, in the last group, waiting on nothing", altar.canBuild, altar.group, altar.wait)
	}
	if want := "Needs Calendar (research), after Ritual"; altar.reason != want {
		t.Errorf("the Altar's reason is %q, want %q", altar.reason, want)
	}
	// Opened, the quiet row lists it, its reason on the screen and its button off.
	d.economyTab.buildBox.expanded = true
	drawDashboard(t, d, pages, 160, 60)
	for _, pc := range d.economyTab.buildBox.placed {
		if pc.card.key == "altar" {
			if !d.economyTab.buildBox.cardKeys.keys[pc.build].off {
				t.Error("the Altar's Build button is on")
			}
			return
		}
	}
	t.Error("the opened list does not show the Altar")
}

// TestLockedBuildingGoneWhenResearched: once Calendar is researched the Altar
// is an ordinary card with no need.
func TestLockedBuildingGoneWhenResearched(t *testing.T) {
	d, _ := bronzeTown(t, "language", "ritual", "calendar")
	altar := altarCard(d)
	if altar == nil {
		t.Fatal("the researched Altar has no card")
	}
	if strings.Contains(altar.reason, "Needs") {
		t.Errorf("the Altar still has a need once Calendar is researched: %q", altar.reason)
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
	// With Language researched, the path is [Ritual], so it uses the "after" format
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

// TestLockedNeedsLineFormats: the needs line has different formats based on
// the path length and next techs.
func TestLockedNeedsLineFormats(t *testing.T) {
	tests := []struct {
		name    string
		techs   []string
		wantIn  string
		notWant string
	}{
		// Path empty: "Needs X (research)"
		// (Ritual and Language researched, so path = [])
		{
			name:    "path empty",
			techs:   []string{"language", "ritual"},
			wantIn:  "Needs Calendar (research)",
			notWant: "after",
		},
		// Path length 1: "Needs X (research), after Y"
		// (Language researched, so path = [Ritual])
		{
			name:    "path length 1",
			techs:   []string{"language"},
			wantIn:  "Needs Calendar (research), after Ritual",
			notWant: "Start with",
		},
		// Path length 2: "Needs X (research), 2 techs before it. Start with Y"
		// (Nothing researched, so path = [Ritual, Language])
		{
			name:    "path length 2",
			techs:   []string{},
			wantIn:  "Needs Calendar (research), 2 techs before it. Start with Language",
			notWant: "after",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := bronzeTown(t, tt.techs...)
			altar := altarCard(d)
			if altar == nil {
				t.Fatal("the Altar has no card")
			}
			text := altar.reason
			if !strings.Contains(text, tt.wantIn) {
				t.Errorf("got %q, want substring %q", text, tt.wantIn)
			}
			if tt.notWant != "" && strings.Contains(text, tt.notWant) {
				t.Errorf("got %q, should not contain %q", text, tt.notWant)
			}
		})
	}
}

// TestLockedNeedsLineJoinNames: joinNames formats tech names correctly.
func TestLockedNeedsLineJoinNames(t *testing.T) {
	tests := []struct {
		names   []string
		want    string
		wantRem int
	}{
		{[]string{}, "", 0},
		{[]string{"A"}, "A", 0},
		{[]string{"A", "B"}, "A or B", 0},
		{[]string{"A", "B", "C"}, "A, B or C", 0},
		{[]string{"A", "B", "C", "D"}, "A, B or C", 1},
		{[]string{"A", "B", "C", "D", "E", "F", "G"}, "A, B or C", 4},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d names", len(tt.names)), func(t *testing.T) {
			got, rem := joinNames(tt.names)
			if got != tt.want || rem != tt.wantRem {
				t.Errorf("joinNames(%v) = %q, %d; want %q, %d", tt.names, got, rem, tt.want, tt.wantRem)
			}
		})
	}
}
