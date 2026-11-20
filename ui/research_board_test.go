package ui

import (
	"image"
	"image/png"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// boardOn is the board on a game, laid out at w by h and painted once so the
// mouse has its rows.
func boardOn(ge *game.GameEngine, w, h int) *researchPanel {
	p := newResearchPanel()
	p.engine = ge
	p.update(ge.GetState())
	p.Layout(image.Rect(0, 0, w, h))
	paintAlone(p, w, h)
	return p
}

func capital(r rune) *kit.Key { return newKey(kit.KeyRune, r, kit.ModShift) }

func TestBoardListsOnlyWhatCanStart(t *testing.T) {
	ge := classicalGame(t)
	p := boardOn(ge, 279, 74)
	st := ge.GetState()
	if len(p.ready) == 0 || len(p.coming) == 0 {
		t.Fatalf("ready %d, coming %d: both lists should have techs", len(p.ready), len(p.coming))
	}
	for _, tt := range p.ready {
		if !tt.ts.Available || tt.def.Key == st.Research.CurrentTech || p.model.queued[tt.def.Key] {
			t.Errorf("%s is listed as ready but is not startable", tt.def.Key)
		}
	}
	for _, tt := range p.coming {
		if tt.ts.Available || tt.ts.Researched {
			t.Errorf("%s is listed as coming up but is %+v", tt.def.Key, tt.ts)
		}
	}
	if p.model.queued["road_building"] != true {
		t.Errorf("the planned tech is not in the queue")
	}
}

func TestBoardKeys(t *testing.T) {
	ge := classicalGame(t)
	p := boardOn(ge, 279, 74)
	first := p.view.sel
	if first != p.ready[0].def.Key {
		t.Fatalf("opens on %q, want the first ready tech %q", first, p.ready[0].def.Key)
	}
	if !p.routeKey(newKey(kit.KeyDown, 0, 0), "") || p.view.sel == first {
		t.Fatalf("Down did not move the pick off %q", first)
	}
	if p.routeKey(newKey(kit.KeyEnter, 0, 0), "typed") {
		t.Errorf("Enter with something typed belongs to the prompt")
	}
	if p.routeKey(newKey(kit.KeyRune, 'l', 0), "") || p.routeKey(capital('L'), "typed") {
		t.Errorf("a lower case letter, or a letter with a prompt typed, belongs to the prompt")
	}
	sel := p.selected().def
	if !p.routeKey(capital('L'), "") || p.view.lane != sel.Lane {
		t.Fatalf("L did not hold the board to %q (lane %q)", sel.Lane, p.view.lane)
	}
	for _, tt := range p.all() {
		if tt.def.Lane != sel.Lane {
			t.Errorf("%s (%s) shows with the board held to %s", tt.def.Key, tt.def.Lane, sel.Lane)
		}
	}
	if !p.closeLane() || p.view.lane != "" || p.closeLane() {
		t.Errorf("Esc should show every lane once, then fall through to closing")
	}
	var opened string
	p.toTree = func(k string) { opened = k }
	p.routeKey(capital('T'), "")
	if opened != p.view.sel {
		t.Errorf("T opened the tree on %q, want %q", opened, p.view.sel)
	}
}

func TestBoardEnterQueuesWhileOneIsInWork(t *testing.T) {
	ge := classicalGame(t)
	p := boardOn(ge, 279, 74)
	key := p.view.sel
	before := len(ge.GetState().Research.Queue)
	p.routeKey(newKey(kit.KeyEnter, 0, 0), "")
	after := ge.GetState().Research.Queue
	if p.view.note != "" || len(after) != before+1 {
		t.Fatalf("Enter queued %d -> %d techs, note %q", before, len(after), p.view.note)
	}
	if got := after[len(after)-1]; got != key {
		t.Errorf("queued %q, want %q", got, key)
	}
	if p.model.queued[key] != true {
		t.Errorf("the board does not show %q as queued", key)
	}
}

func TestBoardMouse(t *testing.T) {
	ge := classicalGame(t)
	p := boardOn(ge, 279, 74)
	var row, btn, lane *boardHit
	for i := range p.hits {
		h := &p.hits[i]
		switch {
		case h.kind == hitTech && h.id != p.view.sel && row == nil:
			row = h
		case h.kind == hitQueue:
			btn = h
		case h.kind == hitLane && lane == nil:
			lane = h
		}
	}
	if row == nil || btn == nil || lane == nil {
		t.Fatalf("hits missing: row %v, queue button %v, lane %v", row, btn, lane)
	}
	press := func(r image.Rectangle, clicks int) {
		p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: r.Min.X + 1, Y: r.Min.Y + 1, Clicks: clicks})
	}
	press(row.r, 1)
	if p.view.sel != row.id {
		t.Fatalf("a click on a row picked %q, want %q", p.view.sel, row.id)
	}
	if p.PointerAt(row.r.Min.Add(image.Pt(1, 1))) != kit.PointerHand {
		t.Errorf("no hand over a row")
	}
	before := len(ge.GetState().Research.Queue)
	paintAlone(p, 279, 74)
	for _, h := range p.hits {
		if h.kind == hitQueue {
			press(h.r, 1)
		}
	}
	if n := len(ge.GetState().Research.Queue); n != before+1 {
		t.Errorf("the Queue button left %d queued, want %d", n, before+1)
	}
	paintAlone(p, 279, 74)
	for _, h := range p.hits {
		if h.kind == hitLane {
			press(h.r, 1)
			break
		}
	}
	if p.view.lane == "" {
		t.Errorf("a click on a lane row did not hold the board to it")
	}
}

func TestBoardPaintsEverySection(t *testing.T) {
	ge := classicalGame(t)
	var lines []string
	for _, size := range [][2]int{{279, 74}, {279, 68}, {120, 37}} {
		f := paintAlone(boardOn(ge, size[0], size[1]), size[0], size[1])
		lines = f.Lines()
		all := strings.Join(lines, "\n")
		for _, want := range []string{"NOW RESEARCHING", "READY NOW", "COMING UP", "TIME", "Esc"} {
			if !strings.Contains(all, want) {
				t.Errorf("%dx%d: no %q on the board", size[0], size[1], want)
			}
		}
		if size[0] >= 150 {
			for _, want := range []string{"UNLOCKS", "IN THE TREE", "Research", "Queue", "THE TEN LANES"} {
				if size[1] < 60 && want == "THE TEN LANES" {
					continue
				}
				if !strings.Contains(all, want) {
					t.Errorf("%dx%d: no %q on the board", size[0], size[1], want)
				}
			}
		}
	}
}

// TestBoardPicture writes the 279 by 74 picture when RESEARCH_SHOT names a file.
func TestBoardPicture(t *testing.T) {
	out := os.Getenv("RESEARCH_SHOT")
	if out == "" {
		t.Skip("RESEARCH_SHOT not set")
	}
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, ScreenImagePixel(paintAlone(boardOn(shotGame(t), 279, 74), 279, 74))); err != nil {
		t.Fatal(err)
	}
}

// shotGame is a made-up account in the Medieval Age with the earlier ages'
// techs in hand, one tech in work and two queued.
func shotGame(t *testing.T) *game.GameEngine {
	t.Helper()
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	for _, a := range []string{"stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age"} {
		must(t, ge.EnterAgeForTest(a))
	}
	st := ge.GetState()
	set := st.Ruleset()
	now, _ := set.Index("medieval_age")
	var keys []string
	for _, d := range set.Techs() {
		if i, _ := set.Index(d.Age); i < now {
			keys = append(keys, d.Key)
		}
	}
	ge.GrantTechsForTest(keys...)
	ge.SetTownForTest(map[string]int{"story_circle": 8, "elders_hall": 8, "scriptorium": 8, "agora": 10}, 0)
	var avail []string
	for k, ts := range ge.GetState().Research.Techs {
		if ts.Available {
			avail = append(avail, k)
		}
	}
	sort.Strings(avail)
	must(t, ge.StartResearch(avail[0]))
	must(t, ge.PlanAddResearch(avail[1]))
	must(t, ge.PlanAddResearch(avail[2]))
	return ge
}
