package ui

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/logroll"
	"github.com/espresso20/ageforge/window/kit"
)

// logsGame is a made-up town with a spread of lines across the offices.
func logsGame(t *testing.T) *game.GameEngine {
	t.Helper()
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	for _, a := range []string{"stone_age", "bronze_age"} {
		must(t, ge.EnterAgeForTest(a))
	}
	ge.AddLog(game.OfficeGame, game.LevelHeadline, "event", "The Bronze Age began.")
	for i := 0; i < 40; i++ {
		ge.AddLog(game.OfficeBuilding, game.LevelTown, "success", fmt.Sprintf("Built clay pit number %d.", i))
		ge.AddLog(game.OfficeTrade, game.LevelTown, "info", fmt.Sprintf("Traders came with pots %d.", i))
		ge.AddLog(game.OfficeStores, game.LevelEverything, "warning", fmt.Sprintf("Stone ran low %d.", i))
	}
	ge.AddLog(game.OfficeResearch, game.LevelTown, "success", "Pottery was learned.")
	return ge
}

func logsOn(ge *game.GameEngine, w, h int) *logsPanel {
	p := newLogsPanel()
	p.open(ge.GetState())
	p.Layout(image.Rect(0, 0, w, h))
	paintAlone(p, w, h)
	return p
}

// Every rect the plan gives lies inside the panel and the rects that hold
// pictures or words do not overlap, at the sizes the window gives.
func TestLogsLayoutFits(t *testing.T) {
	ge := logsGame(t)
	for _, sz := range [][2]int{{279, 71}, {279, 68}, {157, 43}, {157, 40}} {
		p := logsOn(ge, sz[0], sz[1])
		l := p.lay
		if l.Mode == logroll.Placeholder {
			t.Fatalf("%dx%d: the placeholder showed", sz[0], sz[1])
		}
		all := map[string]logroll.Rect{"title": l.Title, "find": l.Find, "rail": l.Rail, "roll": l.Roll, "slate": l.Slate, "lever": l.Lever, "hint": l.Hint}
		for i, s := range l.Stamps {
			all[fmt.Sprint("stamp", i)] = s
		}
		bounds := logroll.Rect{W: sz[0], H: sz[1]}
		names := make([]string, 0, len(all))
		for n, r := range all {
			names = append(names, n)
			if r.X < 0 || r.Y < 0 || r.Right() > bounds.Right() || r.Bottom() > bounds.Bottom() {
				t.Errorf("%dx%d: %s %v is outside the panel", sz[0], sz[1], n, r)
			}
		}
		for i, a := range names {
			for _, b := range names[i+1:] {
				if all[a].Overlaps(all[b]) && !(a == "title" && b == "find" || a == "find" && b == "title") && !(a == "roll" || b == "roll") && !(a == "slate" || b == "slate") {
					t.Errorf("%dx%d: %s and %s overlap", sz[0], sz[1], a, b)
				}
			}
		}
	}
	if p := logsOn(ge, 100, 30); p.lay.Mode != logroll.Placeholder {
		t.Errorf("100x30 did not give the placeholder")
	}
	if p := logsOn(ge, 150, 38); p.lay.Mode != logroll.Compact {
		t.Errorf("150x38 is not compact")
	}
}

// Find matches whole words, in the message, the amounts and the lines a fold
// holds.
func TestLogsFindWholeWords(t *testing.T) {
	p := logsOn(logsGame(t), 279, 68)
	p.on[game.OfficeStores] = true
	p.setLevel(game.LevelEverything)
	p.find = "pots"
	p.build()
	n := 0
	for _, r := range p.rows {
		if r.kind == rowLine {
			n++
			if !strings.Contains(r.text, "pots") || len(r.hits) == 0 {
				t.Errorf("%q matched without a mark", r.text)
			}
		}
	}
	if n == 0 {
		t.Fatalf("Find pots found nothing")
	}
	p.find = "pot"
	p.build()
	for _, r := range p.rows {
		if r.kind == rowLine {
			t.Errorf("Find pot matched %q: part of a word is not a word", r.text)
		}
	}
}

// Keys, clicks and commands reach the state.
func TestLogsInput(t *testing.T) {
	ge := logsGame(t)
	p := logsOn(ge, 279, 68)
	if p.level != game.LevelTown || p.on[game.OfficeGame] {
		t.Fatalf("opened at level %d, Game office %v", p.level, p.on[game.OfficeGame])
	}
	p.Key(*newKey(kit.KeyRune, 'T', 0))
	if p.on[game.OfficeTrade] {
		t.Errorf("T did not stamp Trade off")
	}
	p.Key(*newKey(kit.KeyRune, 'A', 0))
	if !p.on[game.OfficeTrade] || !p.on[game.OfficeGame] {
		t.Errorf("A did not turn every office on")
	}
	p.Key(*newKey(kit.KeyRune, 'F', 0))
	for _, r := range "pottery" {
		p.Key(*newKey(kit.KeyRune, r, 0))
	}
	p.Key(*newKey(kit.KeyEnter, 0, 0))
	if p.find != "pottery" || p.typing {
		t.Errorf("Find is %q, typing %v", p.find, p.typing)
	}
	// A click on the second stop of the lever sets the level.
	p.find = ""
	p.build()
	paintAlone(p, 279, 68)
	st := p.lay.Stops[2]
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: st.X + 1, Y: st.Y + 1, Clicks: 1})
	if p.level != game.LevelEverything {
		t.Errorf("a click on the third stop left the level at %d", p.level)
	}
	// A click on a stamp turns an office off.
	s := p.lay.Stamps[game.OfficeBuilding-1]
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: s.X + 1, Y: s.Y + 1, Clicks: 1})
	if p.on[game.OfficeBuilding] {
		t.Errorf("a click on the Building stamp left it on")
	}
	// Commands.
	for cmd, want := range map[string]logsRequest{
		"logs headlines":   {level: 1},
		"logs everything":  {level: 3},
		"logs all":         {all: true},
		"logs trade":       {office: game.OfficeTrade},
		"logs find clay x": {setFind: true, find: "clay x"},
	} {
		res := HandleCommand(cmd, ge)
		if res.OverlayName != "logs" || res.Logs != want {
			t.Errorf("%q gave overlay %q, request %+v; want %+v", cmd, res.OverlayName, res.Logs, want)
		}
	}
	if res := HandleCommand("logs nonsense", ge); res.Type != "error" {
		t.Errorf("logs nonsense was not refused")
	}
}

// The newest line is at the foot of the roll, and the slate puts what is under
// way first.
func TestLogsRollAndSlate(t *testing.T) {
	ge := logsGame(t)
	p := logsOn(ge, 279, 68)
	if last := p.rows[len(p.rows)-1]; last.text != "Pottery was learned." {
		t.Errorf("the newest row is %q", last.text)
	}
	if err := ge.StartResearch("pottery"); err == nil {
		p.update(ge.GetState())
		if d := p.slate(); len(d.under) == 0 || d.under[len(d.under)-1].kind != srBar {
			t.Errorf("the slate's Under way lacks the research: %v", d.under)
		}
	}
	if d := p.slate(); len(d.tallies) != 8 || d.tallies[0].n == "0" {
		t.Errorf("the sitting's tally is %+v", d.tallies)
	}
}

// A saved log comes back with a seam, which the roll labels.
func TestLogsSeamLabel(t *testing.T) {
	p := newLogsPanel()
	st := game.GameState{Log: []game.LogEntry{
		{Tick: 1, At: 1000, Office: game.OfficeBuilding, Level: game.LevelTown, Message: "Built a hut."},
		{Seam: true, At: 1000, Count: 1},
		{Tick: 2, At: 2000, Office: game.OfficeBuilding, Level: game.LevelTown, Message: "Built a hall."},
	}}
	p.open(st)
	p.Layout(image.Rect(0, 0, 279, 68))
	p.build()
	found := false
	for _, r := range p.rows {
		found = found || r.kind == rowSeam && strings.Contains(r.text, "1 lines")
	}
	if !found {
		t.Errorf("no seam row in %+v", p.rows)
	}
}

// shotLog is a made-up town's log: a sitting that ended, headlines, a fold
// with its lines, builds in every office and troubles.
func shotLog(now int64) []game.LogEntry {
	var log []game.LogEntry
	at := now - 3*3600
	add := func(o game.Office, lv int, typ, msg string, paid, gained []game.LogAmt) {
		at += 40
		log = append(log, game.LogEntry{Tick: int(at / 4), At: at, Office: o, Level: lv, Type: typ, Message: msg, Paid: paid, Gained: gained})
	}
	for i := 0; i < 20; i++ {
		add(game.OfficeBuilding, game.LevelTown, "success", fmt.Sprintf("Mill %d built (you have %d)", i+1, i+1), []game.LogAmt{{Res: "wood", N: 120 * float64(i+1)}}, nil)
		if i%4 == 0 {
			add(game.OfficeTrade, game.LevelTown, "info", "Trade route Local Barter is running again after 9m.", nil, []game.LogAmt{{Res: "gold", N: 45}})
			add(game.OfficePeople, game.LevelTown, "info", fmt.Sprintf("%d villagers joined the town.", 3+i), nil, nil)
			add(game.OfficeStores, game.LevelEverything, "warning", "Wood is nearly full.", nil, nil)
		}
	}
	add(game.OfficeWorld, game.LevelHeadline, "success", "Advanced from the Dark Age to the Medieval Age.", nil, nil)
	log[len(log)-1].Banner = true
	log = append(log, game.LogEntry{Seam: true, At: at, Count: len(log)})
	at += 3600 - 1200
	for i := 0; i < 8; i++ {
		add(game.OfficeResearch, game.LevelTown, "success", fmt.Sprintf("Research %d was learned.", i+1), nil, nil)
		add(game.OfficeArmy, game.LevelTown, "info", "The scouts came back from the ridge.", nil, nil)
		add(game.OfficeDiplomacy, game.LevelTown, "info", "The Riverlands Tribes sent word.", nil, nil)
		add(game.OfficePlan, game.LevelTown, "info", "A plan step was finished.", nil, nil)
	}
	add(game.OfficeStores, game.LevelTown, "warning", "Food is falling: gone in about 38m.", nil, nil)
	log[len(log)-1].Trouble = true
	f := game.LogEntry{Tick: int(at / 4), At: at + 300, FirstAt: at + 40, Office: game.OfficeBuilding, Level: game.LevelTown, Type: "success", Message: "8 Manors built (you have 15)", Count: 8, ID: "m", Subject: "manor",
		Paid: []game.LogAmt{{Res: "stone", N: 6400}}, Gained: []game.LogAmt{{Res: "workers", N: 90}}}
	for i := 0; i < 3; i++ {
		f.Members = append(f.Members, game.LogMember{At: at + int64(40*i), Tick: f.Tick, Text: fmt.Sprintf("Manor built (you have %d)", 13+i), Paid: []game.LogAmt{{Res: "stone", N: 800}}})
	}
	log = append(log, f)
	add(game.OfficeWorld, game.LevelHeadline, "success", "Wonder available: Great Library", nil, nil)
	add(game.OfficeTrade, game.LevelTown, "info", "The Riverlands Tribes offer 3 new deals.", nil, nil)
	return log
}

// TestLogsPicture writes pictures of the whole screen, the command box
// included, when LOGS_SHOT names a directory.
func TestLogsPicture(t *testing.T) {
	dir := os.Getenv("LOGS_SHOT")
	if dir == "" {
		t.Skip("LOGS_SHOT not set")
	}
	restoreForge(t)
	was := pictureCell
	pictureCell = func() (int, int) { return 7, 15 }
	t.Cleanup(func() { pictureCell = was })
	now := time.Now().Unix()
	shot := func(name string, w, h int, setup func(*logsPanel)) {
		d, pages, eng := stagedDashboard(t, "medieval_age")
		pages.SwitchToPage("dashboard")
		d.overlayMgr.Show("logs", eng.GetState())
		st := stageOf(d, pages, w, h)
		p := d.logsPanel
		p.state.Log = shotLog(now)
		if rs, ok := p.state.Resources["food"]; ok {
			rs.Rate = -4.2
			p.state.Resources["food"] = rs
		}
		p.build()
		if setup != nil {
			setup(p)
			p.Layout(p.Bounds())
		}
		st.paint()
		f, err := os.Create(dir + "/2026-11-20-logs-" + name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, ScreenImagePixel(&st.buf)); err != nil {
			t.Fatal(err)
		}
	}
	shot("main", 279, 74, nil)
	shot("headlines", 279, 74, func(p *logsPanel) { p.setLevel(game.LevelHeadline) })
	shot("find", 279, 74, func(p *logsPanel) { p.find = "built"; p.build() })
	shot("fold", 279, 74, func(p *logsPanel) {
		for i, r := range p.rows {
			if r.kind == rowLine && r.e.Count > 1 {
				p.sel = i
				p.follow = false
				p.toggleFold()
			}
		}
	})
	shot("compact", 157, 49, nil)
}
