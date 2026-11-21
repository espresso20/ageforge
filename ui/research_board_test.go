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
		if !tt.ts.Available && tt.def.Key != st.Research.CurrentTech {
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
	// Nothing leaves the board for being queued or in work.
	listed := map[string]bool{}
	for _, tt := range p.all() {
		listed[tt.def.Key] = true
	}
	if !listed["road_building"] || !listed[st.Research.CurrentTech] {
		t.Errorf("a queued tech or the one in work is missing from the table")
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
	if p.view.bad || p.view.note == "" || len(after) != before+1 {
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
	p := boardOn(ge, 279, 110)
	var row, lane *boardHit
	for i := range p.hits {
		h := &p.hits[i]
		switch {
		case h.kind == hitTech && h.id != p.view.sel && row == nil:
			row = h
		case h.kind == hitLane && lane == nil:
			lane = h
		}
	}
	if row == nil || lane == nil {
		t.Fatalf("hits missing: row %v, lane %v", row, lane)
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
	paintAlone(p, 279, 110)
	p.buttons.press = func(k kit.Key) { p.routeKey(&k, "") }
	queued := false
	for _, k := range p.keys.keyset().keys {
		if k.key.Rune == 'Q' && !k.r.Empty() {
			press(k.r, 1)
			queued = true
		}
	}
	if !queued {
		t.Fatal("no Queue button on the board")
	}
	if n := len(ge.GetState().Research.Queue); n != before+1 {
		t.Errorf("the Queue button left %d queued, want %d", n, before+1)
	}
	paintAlone(p, 279, 110)
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
	for _, size := range [][2]int{{279, 74}, {279, 110}, {100, 30}} {
		all := strings.Join(paintAlone(boardOn(ge, size[0], size[1]), size[0], size[1]).Lines(), "\n")
		want := []string{"NOW RESEARCHING", "READY NOW", "COMING UP", "TIME", "Esc"}
		if size[0] >= 150 {
			want = append(want, "UNLOCKS", "IN THE TREE", "Research", "Queue", "QUEUE", "more techs wait in later ages")
		}
		if size[1] >= 100 {
			want = append(want, "THE TEN LANES")
		}
		for _, w := range want {
			if !strings.Contains(all, w) {
				t.Errorf("%dx%d: no %q on the board", size[0], size[1], w)
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

func TestBoardQueueKeepsTechsOnTheBoard(t *testing.T) {
	ge := classicalGame(t)
	p := boardOn(ge, 279, 74)
	cur := ge.GetState().Research.CurrentTech
	out := strings.Join(paintAlone(p, 279, 74).Lines(), "\n")
	for _, want := range []string{"in work", "#2", "starts in", "READY NOW"} {
		if !strings.Contains(out, want) {
			t.Errorf("no %q on a board with a queue", want)
		}
	}
	// A locked tech queues with what it needs, and the note says so.
	var locked string
	for _, tt := range p.coming {
		if len(p.model.missing(tt, "")) > 0 && !p.model.queued[tt.def.Key] && tt.def.Key != cur {
			locked = tt.def.Key
			break
		}
	}
	p.pick(locked)
	p.act(false)
	if p.view.bad || !strings.Contains(p.view.note, "queued") {
		t.Fatalf("note %q after queueing %s", p.view.note, locked)
	}
	if !p.model.queued[locked] || p.model.pos[locked] == 0 {
		t.Fatalf("%s is not marked queued", locked)
	}
	// Delete takes it out again; the tech stays on the board.
	if !p.routeKey(newKey(kit.KeyDelete, 0, 0), "") || p.model.queued[locked] {
		t.Errorf("Delete left %s queued", locked)
	}
	if p.routeKey(newKey(kit.KeyDelete, 0, 0), "typed") {
		t.Errorf("Delete was taken with something typed")
	}
	if p.model.byKey[locked] == nil || p.index() < 0 {
		t.Errorf("the pick left the board")
	}
}

func TestBoardEmptyQueueSaysSo(t *testing.T) {
	ge := classicalGame(t)
	ge.PlanClear()
	p := boardOn(ge, 279, 74)
	if out := strings.Join(paintAlone(p, 279, 74).Lines(), "\n"); !strings.Contains(out, "Q queues it") {
		t.Errorf("an empty queue says nothing")
	}
}
