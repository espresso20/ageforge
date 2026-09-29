package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// Every body line fits inside the box, so the precomputed height is exact
// (no wrapped overflow, no see-through rows).
func TestCatastropheModalLayoutFitsWidth(t *testing.T) {
	inner := catastropheModalWidth - 2
	for _, ep := range config.Epochs() {
		for _, already := range []bool{false, true} {
			l := buildCatastropheModalLayout(ep.Key, already, 0.50, game.DefaultEndureOutcome(), game.GameState{})
			rows := 0
			for _, block := range []string{l.header, l.endure, l.succumb, l.hint} {
				for _, line := range strings.Split(block, "\n") {
					rows++
					if w := tview.TaggedStringWidth(line); w > inner {
						t.Errorf("%s (already=%v): line %q is %d cells, box interior is %d", ep.Key, already, line, w, inner)
					}
				}
			}
			// header + endure + succumb + hint + 3 gaps + buttons + 2 border
			if want := rows + 3 + 1 + 2; l.height != want {
				t.Errorf("%s: height %d, want %d", ep.Key, l.height, want)
			}
		}
	}
}

// screenGrid renders pages at w×h and returns one rune slice per row.
func screenGrid(t *testing.T, pages *tview.Pages, w, h int) [][]rune {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(w, h)
	pages.SetRect(0, 0, w, h)
	pages.Draw(sim)
	sim.Show()
	cells, cw, ch := sim.GetContents()
	grid := make([][]rune, ch)
	for y := 0; y < ch; y++ {
		grid[y] = make([]rune, cw)
		for x := 0; x < cw; x++ {
			r := ' '
			if rs := cells[y*cw+x].Runes; len(rs) > 0 {
				r = rs[0]
			}
			grid[y][x] = r
		}
	}
	return grid
}

// The modal is a box floating over the dashboard: the page beneath stays
// visible around it, and nothing beneath shows through inside it.
func TestCatastropheModalFloatsOverDashboard(t *testing.T) {
	const w, h = 160, 50
	pages := tview.NewPages()
	filler := tview.NewTextView().SetText(strings.Repeat(strings.Repeat("X", w)+"\n", h))
	pages.AddPage("dashboard", filler, true, true)
	d := NewDashboard(tview.NewApplication(), game.NewGameEngine(), pages)
	d.showCatastropheModal("steel_era")

	grid := screenGrid(t, pages, w, h)
	title := []rune(" ☄ Steel Era Catastrophe ")
	y0, x0 := -1, -1
	for y, row := range grid {
		if idx := strings.Index(string(row), string(title)); idx >= 0 {
			y0 = y
			// Walk left from the title to the box corner.
			rx := len([]rune(string(row)[:idx]))
			for x := rx; x >= 0; x-- {
				if row[x] == '╔' || row[x] == '┌' {
					x0 = x
					break
				}
			}
			break
		}
	}
	if y0 < 0 || x0 < 0 {
		t.Fatal("modal box not found on screen")
	}
	l := buildCatastropheModalLayout("steel_era", false, 0, game.DefaultEndureOutcome(), game.GameState{})
	for y := y0; y < y0+l.height; y++ {
		for x := x0; x < x0+catastropheModalWidth; x++ {
			if grid[y][x] == 'X' {
				t.Fatalf("dashboard shows through the modal at (%d,%d)", x, y)
			}
		}
	}
	// Outside the box the dashboard is still drawn (the old full-width opaque
	// spacers blanked all of it).
	for _, p := range [][2]int{{0, 0}, {w - 1, h - 1}, {x0 - 1, y0}, {x0 + catastropheModalWidth, y0}, {x0, y0 - 1}, {x0, y0 + l.height}} {
		if grid[p[1]][p[0]] != 'X' {
			t.Errorf("dashboard blanked outside the modal at (%d,%d): %q", p[0], p[1], grid[p[1]][p[0]])
		}
	}
}

// A pending or unrecorded catastrophe must never read as Survived.
func TestEpochHistoryNeverShowsUnresolvedAsSurvived(t *testing.T) {
	state := game.NewGameEngine().GetState()
	state.EpochKey = "electric_era"
	state.EpochEventHistory = []game.EpochEventRecord{
		{EpochKey: "iron_era", EventName: "The Great Plague", EventType: "catastrophe", Outcome: game.CatastrophePending},
		{EpochKey: "steel_era", EventName: "The World War", EventType: "catastrophe"},
	}
	var sb strings.Builder
	epochProviderHistory(&sb, state)
	out := sb.String()
	if strings.Contains(out, "Endured") {
		t.Errorf("unresolved catastrophes shown as Endured:\n%s", out)
	}
	if !strings.Contains(out, "Pending") || !strings.Contains(out, "outcome not recorded") {
		t.Errorf("expected Pending and 'outcome not recorded' labels:\n%s", out)
	}

	state.EpochEventHistory[0].Outcome = game.CatastropheEndured
	state.EpochEventHistory[1].Outcome = game.CatastropheSuccumbed
	sb.Reset()
	epochProviderHistory(&sb, state)
	out = sb.String()
	if !strings.Contains(out, "Endured") || !strings.Contains(out, "Succumbed") {
		t.Errorf("resolved outcomes missing:\n%s", out)
	}
}

func TestStatsCatastropheCounts(t *testing.T) {
	state := game.NewGameEngine().GetState()
	state.CatastrophesEndured = 2
	state.CatastrophesSuccumbed = 3
	// Legacy flags no longer drive the count (a repeat succumb in one epoch has one flag).
	state.LegacyBonuses = map[string]bool{"iron_era": true}
	out := statsProvider(state, 120)
	if !strings.Contains(out, "5  (endured 2, succumbed 3)") {
		t.Errorf("catastrophe counts wrong:\n%s", out)
	}
	// The old "Epochs Survived" line duplicated the endured count; it is gone.
	if strings.Contains(out, "Epochs Survived") {
		t.Errorf("stats still shows the Epochs Survived line:\n%s", out)
	}
}

func TestCatastropheCommandReopensOrReports(t *testing.T) {
	eng := game.NewGameEngine()
	res := HandleCommand("catastrophe", eng)
	if res.OpenCatastrophe || !strings.Contains(res.Message, "No catastrophe pending") {
		t.Errorf("no pending: got %+v", res)
	}
	if !strings.Contains(res.Message, "Iron Era") {
		t.Errorf("outlook should name the next epoch:\n%s", res.Message)
	}
	// The player-facing invoke is gone: it only prints usage and triggers nothing.
	r := HandleCommand("catastrophe invoke", eng)
	if r.OpenCatastrophe || !strings.HasPrefix(r.Message, "Usage: catastrophe") || strings.Contains(r.Message, "invoke") {
		t.Errorf("catastrophe invoke: got %+v, want usage only", r)
	}
	if eng.GetState().PendingCatastrophe != "" {
		t.Error("catastrophe invoke triggered a catastrophe")
	}
	if got := NewAutoCompleter(eng)("catastrophe i"); len(got) != 0 {
		t.Errorf("autocomplete still offers catastrophe subcommands: %v", got)
	}
	if strings.Contains(helpProvider(eng.GetState(), 120), "catastrophe invoke") {
		t.Error("help still lists catastrophe invoke")
	}
}
