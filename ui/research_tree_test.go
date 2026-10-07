package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
)

// treeSizes are the terminal sizes the Research panel must read well at.
// The panel gets the rows above the command bar.
var treeSizes = [][2]int{{80, 24}, {100, 30}, {120, 40}, {144, 46}}

// classicalGame is a game in the Classical Age with every state on the
// map at once: techs researched, one in progress, some that can start,
// some waiting for what they need, one planned, and the next age in sight.
func classicalGame(t *testing.T) *game.GameEngine {
	t.Helper()
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	for _, a := range []string{"stone_age", "bronze_age", "iron_age", "classical_age"} {
		if err := ge.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
	}
	ge.GrantTechsForTest("language", "tool_making", "fire_mastery", "stoneworking", "primitive_writing", "pottery", "woodworking", "bronze_working", "currency", "masonry", "the_wheel", "mathematics", "iron_smelting")
	ge.SetStockForTest("knowledge", 9e5)
	if err := ge.StartResearch("philosophy"); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddResearch("road_building"); err != nil {
		t.Fatal(err)
	}
	return ge
}

// drawTree lays the tree out and renders the panel as the panel does.
func drawTree(st game.GameState, w, h int, v treeView) (*treeModel, treeView, *tGrid) {
	m := buildTree(st, treeGeomFor(w, v.far), v.ascii, v.sel)
	if m.by[v.sel] == nil {
		v.sel = m.home()
		m = buildTree(st, treeGeomFor(w, v.far), v.ascii, v.sel)
	}
	m.follow(&v, w, h)
	g := renderTree(st, m, v, w, h)
	if v.ascii {
		foldPlain(g)
	}
	return m, v, g
}

// cardOf is a tech's card as plain text, a paragraph a line.
func cardOf(st game.GameState, key string) string {
	m := buildTree(st, treeGeomFor(120, false), false, key)
	var out []string
	for _, s := range m.cardSentences(st, m.by[key], "") {
		out = append(out, strings.Map(func(r rune) rune {
			if _, ok := swStyle(r); ok {
				return -1
			}
			return r
		}, s))
	}
	return strings.Join(out, "\n")
}

// TestResearchTreeDrawsEveryTechWhole is the map's picture test, on the
// whole canvas at both zooms and both badge sizes: every tech's badge is
// there cell for cell (nothing clipped it, no line or neighbour drew over
// it), its emblem sits exactly on its centre cell, its name plate is
// centred over it, and every connector cell is joined on to the next: a
// line never stops in mid air or runs under a badge.
//
// Twice: from the Classical Age, with a research running and a tech in the
// plan, and from the Electric Age, where the map holds every age a content
// batch has filled in, their either-or group, their six capstones and the
// lanes that hold three techs in one age.
func TestResearchTreeDrawsEveryTechWhole(t *testing.T) {
	treeDrawsEveryTechWhole(t, classicalGame(t).GetState(), "road_building")
	treeDrawsEveryTechWhole(t, electricGame(t).GetState(), "interchangeable_parts")
}

func treeDrawsEveryTechWhole(t *testing.T, st game.GameState, sel string) {
	t.Helper()
	for _, w := range []int{80, 120} {
		for _, far := range []bool{false, true} {
			geom := treeGeomFor(w, far)
			m := buildTree(st, geom, false, sel)
			if m.by[sel] == nil {
				t.Fatalf("%s is not on the map", sel)
			}
			g := m.canvas(st, sel)
			owned := map[[2]int]string{}
			claim := func(x, y int, key string) {
				if other, ok := owned[[2]int{x, y}]; ok && other != key {
					t.Errorf("%d far=%v: %s and %s both draw on cell %d,%d", w, far, other, key, x, y)
				}
				owned[[2]int{x, y}] = key
			}
			for _, n := range m.nodes {
				left := m.left(n)
				if left < n.lane*geom.laneW || left+geom.badgeW > (n.lane+1)*geom.laneW {
					t.Errorf("%d far=%v: %s is drawn outside its lane", w, far, n.key)
				}
				if geom.pill {
					if got := g.at(n.cx, n.top).r; got != n.emblem {
						t.Errorf("%d far=%v: %s has %q where its emblem %q goes", w, far, n.key, got, n.emblem)
					}
					for x := left; x < left+geom.badgeW; x++ {
						claim(x, n.top, n.key)
					}
					continue
				}
				rows := badgeRows(geom, n, progressOf(st))
				if len(rows) != geom.badgeH {
					t.Fatalf("%s: a badge of %d rows, want %d", n.key, len(rows), geom.badgeH)
				}
				for dy, row := range rows {
					rs := []rune(row)
					if len(rs) != geom.badgeW {
						t.Fatalf("%s: badge row %q is %d cells, want %d", n.key, row, len(rs), geom.badgeW)
					}
					for dx, want := range rs {
						x, y := left+dx, n.top+geom.nameH+dy
						claim(x, y, n.key)
						got := g.at(x, y).r
						// The marks sit on the badge's corners, by design.
						if got != want && !strings.ContainsRune("★✓⟳▸◇", got) {
							t.Errorf("%d far=%v: %s row %d cell %d is %q, want %q", w, far, n.key, dy, dx, got, want)
						}
					}
				}
				// The emblem: the middle cell of the middle row, on the grid.
				if ex, ey := n.cx, n.top+geom.nameH+geom.badgeH/2; g.at(ex, ey).r != n.emblem || ex != left+geom.badgeW/2 {
					t.Errorf("%d far=%v: %s's emblem %q is not on its centre cell (%d,%d holds %q)", w, far, n.key, n.emblem, ex, ey, g.at(ex, ey).r)
				}
				// The name plate: whole, and centred when it fits.
				nx, name := m.nameSpan(n)
				for i, r := range []rune(name) {
					claim(nx+i, n.top, n.key)
					if g.at(nx+i, n.top).r != r {
						t.Errorf("%d far=%v: %s's name plate is broken at cell %d", w, far, n.key, i)
					}
				}
				if len([]rune(name)) < geom.laneW-1 && nx != n.cx-len([]rune(name))/2 {
					t.Errorf("%d far=%v: %s's name is not centred over its badge", w, far, n.key)
				}
			}
			// The connectors: joined cell to cell, and never under a tech.
			step := map[uint8][3]int{dirN: {0, -1, dirS}, dirE: {1, 0, dirW}, dirS: {0, 1, dirN}, dirW: {-1, 0, dirE}}
			if len(m.lines) == 0 {
				t.Fatalf("%d far=%v: no connector at all", w, far)
			}
			for at, c := range m.lines {
				if key, ok := owned[at]; ok {
					t.Errorf("%d far=%v: a connector runs under %s at %v", w, far, key, at)
				}
				if got := g.at(at[0], at[1]).r; got != lineRune(c) && !strings.ContainsRune(" or", got) {
					t.Errorf("%d far=%v: the connector at %v is drawn %q, want %q", w, far, at, got, lineRune(c))
				}
				for bit, d := range step {
					if c.bits&bit == 0 {
						continue
					}
					to := [2]int{at[0] + d[0], at[1] + d[1]}
					if n, ok := m.lines[to]; ok && n.bits&uint8(d[2]) != 0 {
						continue
					}
					// A line may end only on a tech: its notch or its name plate.
					if _, ok := owned[to]; !ok && !(geom.pill && bit == dirN) {
						t.Errorf("%d far=%v: the connector at %v leads %v to nothing", w, far, at, to)
					}
				}
			}
		}
	}
}

// TestResearchTreeAtEverySize: at each size, zoom and glyph tier the panel
// fills its rectangle, the selected tech is whole on screen with its name,
// the title and the key bar are there, nothing is drawn outside the map's
// frame, and the plain tier is plain. The card fits inside the map.
//
// From the Classical Age and from the Electric Age, where the map is twice
// as tall and nine lanes wide, on its longest names and its capstones.
func TestResearchTreeAtEverySize(t *testing.T) {
	treeAtEverySize(t, classicalGame(t).GetState(), "tool_making", "road_building", "philosophy", "siege_warfare")
	treeAtEverySize(t, electricGame(t).GetState(), "patronage", "concert_of_nations", "aviation", "military_industrial_complex")
}

func treeAtEverySize(t *testing.T, st game.GameState, sels ...string) {
	t.Helper()
	for _, size := range treeSizes {
		w, h := size[0], size[1]-promptRows
		for _, far := range []bool{false, true} {
			for _, ascii := range []bool{false, true} {
				for _, sel := range sels {
					m, v, g := drawTree(st, w, h, treeView{sel: sel, far: far, ascii: ascii})
					if g.w != w || g.h != h {
						t.Fatalf("%dx%d: rendered %dx%d", w, h, g.w, g.h)
					}
					n := m.by[sel]
					ox, oy, mw, mh := mapRect(m.geom, w, h)
					left, top := m.left(n)-v.vx, n.top-v.vy
					if left < 0 || left+m.geom.badgeW > mw || top < 0 || m.bottom(n)-v.vy >= mh {
						t.Errorf("%dx%d far=%v: the selected %s is clipped (badge at %d,%d in a map of %dx%d)", w, h, far, sel, left, top, mw, mh)
					}
					if got := g.at(ox+n.cx-v.vx, oy+n.top-v.vy+m.geom.nameH+m.geom.badgeH/2).r; !ascii && got != n.emblem {
						t.Errorf("%dx%d far=%v: %s's emblem is not where the view puts it (found %q)", w, h, far, sel, got)
					}
					text := g.String()
					lines := strings.Split(text, "\n")
					if !strings.Contains(lines[0], "RESEARCH") || !strings.Contains(lines[h-1], "Esc") {
						t.Errorf("%dx%d far=%v: the title or the key bar is missing", w, h, far)
					}
					if !strings.Contains(lines[h-2], n.def.Name) {
						t.Errorf("%dx%d far=%v: the status line does not name %s: %q", w, h, far, sel, lines[h-2])
					}
					if ascii {
						for _, r := range text {
							if r >= 0x80 {
								t.Fatalf("%dx%d far=%v: the plain tier draws %q", w, h, far, r)
							}
						}
					}
					// The card: inside the map, every paragraph of it.
					_, _, cg := drawTree(st, w, h, treeView{sel: sel, far: far, ascii: ascii, card: true})
					ct := cg.String()
					if !strings.Contains(ct, strings.ToUpper(n.def.Name)) || !strings.Contains(ct, "Esc closes the card") {
						t.Errorf("%dx%d far=%v: no card for %s", w, h, far, sel)
					}
					for y, line := range strings.Split(ct, "\n") {
						if strings.ContainsAny(line, "╭╰") && strings.Contains(line, "─────────") && (y < oy || y >= oy+mh) && !far {
							t.Errorf("%dx%d: the card's frame is outside the map (row %d)", w, h, y)
						}
					}
				}
			}
		}
	}
}

// TestResearchTreeLaneRow: every lane in sight is either named over its
// column (in full, or shortened with a full stop) or counted by the arrow
// at the edge it is off, a name never runs into the arrow, and a lane off
// to the left is named in the gutter when the gutter has room.
func TestResearchTreeLaneRow(t *testing.T) {
	st := classicalGame(t).GetState()
	for _, size := range treeSizes {
		w, h := size[0], size[1]-promptRows
		for _, far := range []bool{false, true} {
			for _, sel := range []string{"primitive_writing", "philosophy", "siege_warfare"} {
				m, v, g := drawTree(st, w, h, treeView{sel: sel, far: far})
				row := []rune(strings.Split(g.String(), "\n")[1])
				ox, _, mw, _ := mapRect(m.geom, w, h)
				named, left, right := 0, 0, 0
				for i, l := range m.lanes {
					lx := ox + i*m.geom.laneW - v.vx + 1
					switch {
					case lx < ox:
						left++
					case lx+1 < ox+mw && row[lx] == []rune(l.Emblem)[0] && row[lx+1] == ' ':
						named++
						word := strings.Fields(string(row[lx+2:]))[0]
						full := laneLabel(l.Name, 99)
						if word != full && !(strings.HasSuffix(word, ".") && strings.HasPrefix(full, strings.TrimSuffix(word, "."))) {
							t.Errorf("%dx%d far=%v: lane %s reads %q", w, h, far, l.Key, word)
						}
					default:
						right++
					}
				}
				if named+left+right != len(m.lanes) || named == 0 {
					t.Fatalf("%dx%d far=%v: %d lanes, %d named", w, h, far, len(m.lanes), named)
				}
				text := string(row)
				if got := strings.Contains(text, "◂"); got != (left > 0) {
					t.Errorf("%dx%d far=%v sel=%s: %d lanes off to the left, lane row %q", w, h, far, sel, left, text)
				}
				wantRight := ""
				if right > 0 {
					wantRight = string(rune('0'+right)) + "▸"
				}
				if got := strings.Contains(text, "▸"); got != (right > 0) || !strings.Contains(text, wantRight) {
					t.Errorf("%dx%d far=%v sel=%s: %d lanes off to the right, lane row %q", w, h, far, sel, right, text)
				}
				if left == 1 && ox >= 10 && !strings.HasPrefix(text, "◂ "+laneLabel(m.lanes[0].Name, 99)) {
					t.Errorf("%dx%d far=%v: the lane off to the left is not named: %q", w, h, far, text)
				}
			}
		}
	}
	for _, c := range []struct {
		name string
		room int
		want string
	}{
		{"Knowledge", 9, "KNOWLEDGE"}, {"Knowledge", 7, "KNOWL."}, {"Materials", 7, "MATER."},
		{"Military", 7, "MILIT."}, {"Computing", 7, "COMPUT."}, {"Faith & Culture", 5, "FAITH"},
		{"Agriculture", 4, ""}, {"Agriculture", 0, ""},
	} {
		if got := laneLabel(c.name, c.room); got != c.want {
			t.Errorf("laneLabel(%q, %d) = %q, want %q", c.name, c.room, got, c.want)
		}
	}
}

// TestHelpListsResearchKeys: the Help panel lists the Research panel's keys,
// every key the panel takes, at any width (TestHelpKeepsTwoColumns holds
// the rows to the panel's two columns).
func TestHelpListsResearchKeys(t *testing.T) {
	for _, screenW := range []int{0, 80, 144} {
		help := visible(safeTags(helpProvider(game.GameState{}, screenW)))
		at := strings.Index(help, "The Research panel")
		if at < 0 {
			t.Fatal("the Help panel has no Research panel section")
		}
		section := help[at:]
		if end := strings.Index(section, "═══ Shortcuts"); end > 0 {
			section = section[:end]
		}
		for _, k := range researchKeys {
			if !strings.Contains(section, "  "+k[0]) {
				t.Errorf("%d columns: the Help panel's Research section does not list %s", screenW, k[0])
			}
		}
		for _, want := range []string{"Shift-Tab", "research tree far"} {
			if !strings.Contains(section, want) {
				t.Errorf("%d columns: the Help panel's Research section does not mention %s", screenW, want)
			}
		}
	}
	// Every key the panel routes is one the Help panel names.
	listed := ""
	for _, k := range researchKeys {
		listed += k[0] + " " + k[1] + " "
	}
	for _, name := range []string{"Arrows", "Tab", "Shift-Tab", "PgUp", "PgDn", "Home", "Enter", "Esc"} {
		if !strings.Contains(listed, name) {
			t.Errorf("the Research panel's key list leaves out %s", name)
		}
	}
}

// TestResearchTreePaintsInEveryTheme: in every theme, at every size, every
// cell the panel draws can be told from its background.
func TestResearchTreePaintsInEveryTheme(t *testing.T) {
	st := classicalGame(t).GetState()
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for _, size := range treeSizes {
			w, h := size[0], size[1]-promptRows
			scr := tcell.NewSimulationScreen("UTF-8")
			if err := scr.Init(); err != nil {
				t.Fatal(err)
			}
			scr.SetSize(w, h)
			p := newResearchPanel()
			p.view.sel = "road_building"
			p.SetRect(0, 0, w, h)
			p.update(st)
			p.Draw(scr)
			scr.Show()
			cells, cw, _ := scr.GetContents()
			seen := 0
			for i, c := range cells {
				if len(c.Runes) == 0 || c.Runes[0] == ' ' {
					continue
				}
				seen++
				fg, bg, _ := c.Style.Decompose()
				if fg == bg {
					t.Fatalf("theme %s at %dx%d: cell %d,%d (%q) is drawn in its own background colour", th.Key, w, h, i%cw, i/cw, c.Runes[0])
				}
			}
			if seen < w {
				t.Errorf("theme %s at %dx%d: the panel drew %d cells", th.Key, w, h, seen)
			}
			scr.Fini()
		}
	}
}

// TestResearchTreeTellsStatesApartByShape: each state has a mark or a
// frame of its own, so the map reads without colour. The Classical Age
// game has all of them on the map.
func TestResearchTreeTellsStatesApartByShape(t *testing.T) {
	st := classicalGame(t).GetState()
	geom := treeGeomFor(120, false)
	m := buildTree(st, geom, false, "")
	g := m.canvas(st, "")
	badge := func(key string) string {
		n := m.by[key]
		var sb strings.Builder
		for y := n.top + geom.nameH; y <= m.bottom(n); y++ {
			for x := m.left(n); x <= m.left(n)+geom.badgeW; x++ {
				sb.WriteRune(g.at(x, y).r)
			}
			sb.WriteByte('\n')
		}
		return sb.String()
	}
	for key, c := range map[string]struct {
		mark techMark
		has  string // what its badge must show
		not  string // and what it must not
	}{
		"mathematics":      {markDone, "★╔═════╗✓", "┄"}, // researched keystone: star, double frame, tick
		"pottery":          {markDone, "╭─────╮✓", "═"},  // researched, optional: rounded frame, tick
		"philosophy":       {markRunning, "░", "▾"},      // in progress: the bottom edge is a bar
		"animal_husbandry": {markReady, "╭─────╮ ", "✓"}, // can start: solid frame, no mark
		"road_building":    {markReady, "▸1", "✓"},       // planned: its place in the plan
		"imperial_legions": {markLocked, "┆", "✓"},       // waits for what it needs: dashed
		"military_tactics": {markReady, "◇", "┄"},        // opens a command
		"theology":         {markNext, "░░░░░░░░░", "╔"}, // the next age: a shaded block
	} {
		n := m.by[key]
		if n == nil {
			t.Fatalf("%s is not on the map", key)
		}
		b := badge(key)
		if n.mark != c.mark || !strings.Contains(b, c.has) || strings.Contains(b, c.not) {
			t.Errorf("%s: state %v (want %v), badge\n%swant it to show %q and not %q", key, n.mark, c.mark, b, c.has, c.not)
		}
	}
	// The spine's frame is double whatever the state.
	if b := badge("steel_forging"); !strings.Contains(b, "░") {
		t.Errorf("a next-age tech is drawn as a badge, not a block:\n%s", b)
	}
	// The far zoom keeps a mark for each state too.
	far := buildTree(st, treeGeomFor(80, true), false, "")
	fg := far.canvas(st, "")
	for key, want := range map[string]string{"mathematics": "✓π", "philosophy": "⟳Φ", "animal_husbandry": "(♞)", "imperial_legions": "┆⚑┆", "theology": "░Θ░"} {
		n := far.by[key]
		got := string([]rune{fg.at(n.cx-1, n.top).r, fg.at(n.cx, n.top).r, fg.at(n.cx+1, n.top).r})
		if !strings.HasPrefix(got, want) {
			t.Errorf("far zoom: %s reads %q, want %q", key, got, want)
		}
	}
}

// TestResearchTreeKeepsTheSpoilerRule: ages up to yours are drawn in full,
// the next age is drawn dim, and later ages are a count. No tech of a later
// age is named on the map, in the status line, on any card or by the
// research card command, and a lane has a column only once a tech of it is
// in sight.
func TestResearchTreeKeepsTheSpoilerRule(t *testing.T) {
	ge := classicalGame(t)
	st := ge.GetState()
	set := st.Ruleset()
	sight := game.SightOf(&st)
	var hidden []config.TechDef
	lanes := map[string]bool{}
	for _, def := range set.Techs() {
		if sight.Age(def.Age) {
			lanes[def.Lane] = true
		} else {
			hidden = append(hidden, def)
		}
	}
	if len(hidden) == 0 {
		t.Fatal("no tech is out of sight: the game does not test the rule")
	}
	m, _, g := drawTree(st, 144, 43, treeView{far: true})
	if m.later != len(hidden) || !strings.Contains(g.String(), "87 more techs wait in later ages") {
		t.Errorf("the map counts %d techs in later ages, want %d and a line that says so", m.later, len(hidden))
	}
	if len(m.lanes) != len(lanes) {
		t.Errorf("the map has %d lane columns, want the %d lanes with a tech in sight", len(m.lanes), len(lanes))
	}
	var all strings.Builder
	all.WriteString(g.String())
	for _, n := range m.nodes {
		all.WriteString(cardOf(st, n.key) + "\n")
		all.WriteString(m.statusLine(st, treeView{sel: n.key}) + "\n")
	}
	text := all.String()
	for _, def := range hidden {
		if m.by[def.Key] != nil {
			t.Errorf("%s, a tech of a later age, is on the map", def.Key)
		}
		if strings.Contains(text, def.Name) {
			t.Errorf("the panel names %s, a tech of a later age", def.Name)
		}
		if res := cmdResearch([]string{"card", def.Key}, ge); res.Type != "error" || strings.Contains(res.Message, def.Name) || res.OverlayName != "" {
			t.Errorf("research card %s: %+v, want a refusal that does not name it", def.Key, res)
		}
	}
	if res := cmdResearch([]string{"card", "road", "building"}, ge); res.OverlayName != "techs" || res.ResearchCard != "road_building" {
		t.Errorf("research card road building: %+v", res)
	}
}

// TestResearchCardReadsAsSentences: the card says what a tech does in the
// game's own words for its effects, what it opens, its price and time, what
// it needs and what Enter does, and never tags a tech's bonus as capped.
func TestResearchCardReadsAsSentences(t *testing.T) {
	ge := classicalGame(t)
	st := ge.GetState()
	for key, wants := range map[string][]string{
		"road_building":    {"Road Building gives +8% gold production. With it, trade routes take 15% less time.", "It builds on Masonry and The Wheel.", "It is item 1 in your plan."},
		"military_tactics": {"gives +15% military power.", "It opens campaigns.", "It builds on Bronze Working.", "adds it to your plan. Philosophy is running, so it starts in"},
		"mathematics":      {"It is this age's keystone: the Colosseum cannot be built without it.", "You already hold this."},
		"philosophy":       {"It is being researched:", "Type research cancel to stop it"},
		"imperial_legions": {"It builds on Siege Warfare and Iron Smelting.", "adds it to your plan, after the 2 techs it needs first (Military Tactics, then Siege Warfare). They run in that order, one at a time."},
		"siege_warfare":    {"adds it to your plan, after Military Tactics, which it needs first. They run in that order, one at a time."},
		"theology":         {"the Great Library cannot be built without it.", "It starts once you reach the Medieval Age and hold what it needs."},
		"tool_making":      {"gives +10% food production and +10% wood production. With it, gathering by hand brings 2 more.", "It needs nothing before it."},
	} {
		card := cardOf(st, key)
		for _, want := range wants {
			if !strings.Contains(card, want) {
				t.Errorf("%s's card is missing %q:\n%s", key, want, card)
			}
		}
		if strings.Contains(card, "capped") {
			t.Errorf("%s's card tags a tech's bonus as capped:\n%s", key, card)
		}
	}
	// The price is the one the tech starts at, research speed included.
	st.Pools["research_speed"] = game.BonusPool{Target: "research_speed", Earned: 0.25, Applied: 0.25}
	def := config.TechByKey()["siege_warfare"]
	quick := game.ResearchTicks(def.ResearchTicks, 0.25, 1, 1, 1)
	if want := "takes " + formatTicks(quick, st) + "."; quick >= def.ResearchTicks || !strings.Contains(cardOf(st, "siege_warfare"), want) {
		t.Errorf("with +25%% research speed Siege Warfare's card should say %q:\n%s", want, cardOf(st, "siege_warfare"))
	}
	if out := statsProvider(st, 140); !strings.Contains(out, "Research speed +25%: techs take 75% of their base time.") {
		t.Errorf("the Stats panel does not say what research speed does")
	}

	// An either-or group: named as one, its lines dashed, with an "or".
	src := rules.FromConfig()
	first := src.Ages[0].Key
	src.Techs = append(src.Techs,
		config.TechDef{Key: "zz_left", Name: "Zz Left", Age: first, Lane: config.LaneCraft},
		config.TechDef{Key: "zz_right", Name: "Zz Right", Age: first, Lane: config.LaneTrade},
		config.TechDef{Key: "zz_join", Name: "Zz Join", Age: first, Lane: config.LaneMilitary,
			Prerequisites: []string{"tool_making"}, AnyOf: []string{"zz_left", "zz_right"}})
	either := game.NewGameEngineWith(rules.Compile(src))
	either.GrantTechsForTest("zz_left")
	es := either.GetState()
	if card := cardOf(es, "zz_join"); !strings.Contains(card, "It builds on Tool Making. It needs one of Zz Left or Zz Right. You hold Zz Left.") {
		t.Errorf("an either-or group's card:\n%s", card)
	}
	em := buildTree(es, treeGeomFor(120, false), false, "zz_join")
	dashed := 0
	for _, c := range em.lines {
		if c.dashed && !c.solid {
			dashed++
		}
	}
	if canvas := em.canvas(es, "zz_join").String(); dashed == 0 || !strings.Contains(canvas, "┄") || !strings.Contains(canvas, " or ") {
		t.Errorf("an either-or group draws %d dashed cells and no \"or\":\n%s", dashed, canvas)
	}
}

// TestResearchPanelKeys: the arrows move and the view follows, Tab steps
// through the techs that can start, PgUp and PgDn zoom, Home returns to the
// current age, Enter opens the card and a second Enter does what it says,
// Esc closes the card first. With something typed, Tab and Enter are the
// prompt's.
func TestResearchPanelKeys(t *testing.T) {
	ge := classicalGame(t)
	p := newResearchPanel()
	p.engine = ge
	p.w, p.h = 80, 21
	p.update(ge.GetState())
	key := func(k tcell.Key, prompt string) bool {
		return p.routeKey(tcell.NewEventKey(k, 0, tcell.ModNone), prompt)
	}
	if p.view.sel != "philosophy" {
		t.Fatalf("the panel opens on %q, want the tech in progress", p.view.sel)
	}
	onScreen := func() {
		t.Helper()
		n := p.model.by[p.view.sel]
		_, _, mw, mh := mapRect(p.model.geom, p.w, p.h)
		if x, y := p.model.left(n)-p.view.vx, n.top-p.view.vy; x < 0 || x+p.model.geom.badgeW > mw || y < 0 || p.model.bottom(n)-p.view.vy >= mh {
			t.Errorf("after moving to %s the view did not follow", p.view.sel)
		}
	}
	moved := map[string]bool{}
	for _, k := range []tcell.Key{tcell.KeyUp, tcell.KeyUp, tcell.KeyRight, tcell.KeyRight, tcell.KeyDown, tcell.KeyLeft, tcell.KeyLeft, tcell.KeyUp, tcell.KeyDown} {
		if !key(k, "") {
			t.Errorf("the panel did not take an arrow key")
		}
		moved[p.view.sel] = true
		onScreen()
	}
	if len(moved) < 5 {
		t.Errorf("nine arrow keys only ever reached %v", moved)
	}
	seen := map[string]bool{}
	for i := 0; i < 12; i++ {
		key(tcell.KeyTab, "")
		if n := p.model.by[p.view.sel]; n.mark != markReady {
			t.Fatalf("Tab stopped on %s, which cannot start", n.key)
		}
		seen[p.view.sel] = true
		onScreen()
	}
	if len(seen) < 3 {
		t.Errorf("Tab only ever reached %v", seen)
	}
	key(tcell.KeyPgUp, "")
	if !p.view.far || !p.model.geom.pill {
		t.Error("PgUp did not zoom out")
	}
	key(tcell.KeyPgDn, "")
	if p.view.far {
		t.Error("PgDn did not zoom back in")
	}
	key(tcell.KeyHome, "")
	if b := p.model.bandOf(p.model.by[p.view.sel]); !b.here {
		t.Errorf("Home went to %s, not the current age", p.view.sel)
	}
	if key(tcell.KeyTab, "res") || key(tcell.KeyEnter, "res") || key(tcell.KeyRune, "") {
		t.Error("the panel took a key that belongs to the prompt")
	}

	// Enter opens the card; nothing starts or is planned on the first one.
	p.view.sel = "military_tactics"
	p.layout()
	plan := len(ge.GetState().Plan)
	key(tcell.KeyEnter, "")
	if !p.view.card || len(ge.GetState().Plan) != plan {
		t.Fatalf("the first Enter: card open %v, plan %d items (was %d)", p.view.card, len(ge.GetState().Plan), plan)
	}
	if key(tcell.KeyLeft, ""); p.view.sel != "military_tactics" {
		t.Error("an arrow moved the selection under the open card")
	}
	// Research is running, so the second Enter plans it.
	key(tcell.KeyEnter, "")
	if st := ge.GetState(); len(st.Plan) != plan+1 || st.Plan[plan].Key != "military_tactics" {
		t.Errorf("the second Enter did not add Military Tactics to the plan: %+v", st.Plan)
	}
	if !p.closeCard() || p.view.card || p.closeCard() {
		t.Error("Esc should close the card once, and then have nothing to close")
	}
	// With the slot free and the knowledge there, the card starts it.
	if err := ge.CancelResearch(); err != nil {
		t.Fatal(err)
	}
	ge.SetStockForTest("knowledge", 9e5)
	p.update(ge.GetState())
	p.view.sel = "animal_husbandry"
	p.layout()
	key(tcell.KeyEnter, "")
	key(tcell.KeyEnter, "")
	if cur := ge.GetState().Research.CurrentTech; cur != "animal_husbandry" {
		t.Errorf("Enter on the card of a tech that can start: researching %q, want animal_husbandry", cur)
	}
	// research tree far, and research card, set up the next open.
	p.request("far", "currency")
	if !p.view.far || !p.view.card || p.view.sel != "currency" {
		t.Errorf("research tree far and research card currency left the view %+v", p.view)
	}
}

// TestResearchListMarksTheKeystone: `research list` marks the tech a wonder
// needs, and only that one; on the map a keystone carries its star.
func TestResearchListMarksTheKeystone(t *testing.T) {
	ge := stoneAgeGame(t)
	ge.GrantTechsForTest("tool_making")
	mark := "★ keystone: Great Monolith"
	list := cmdResearchList(ge).Message
	for _, line := range strings.Split(list, "\n") {
		if strings.Contains(line, "stoneworking") != strings.Contains(line, mark) {
			t.Errorf("`research list` line %q: the mark belongs on Stoneworking's line and no other", line)
		}
	}
	if !strings.Contains(list, mark) {
		t.Errorf("`research list` does not mark the keystone:\n%s", list)
	}
	st := ge.GetState()
	if card := cardOf(st, "stoneworking"); !strings.Contains(card, "the Great Monolith cannot be built without it") {
		t.Errorf("Stoneworking's card does not name its wonder:\n%s", card)
	}
	_, _, g := drawTree(st, 120, 37, treeView{sel: "stoneworking"})
	if text := g.String(); strings.Count(text, "★") != 1 || !strings.Contains(text, "keystone") {
		t.Errorf("the Stone Age map should star one tech and call it a keystone in the status line:\n%s", text)
	}
}

// TestPanelsMarkLockedCommands: the Army and Trade panels and `campaign
// list` say when a command still waits for a tech, in the words of the
// refusal, once that tech's age is reached and never before.
func TestPanelsMarkLockedCommands(t *testing.T) {
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	for _, a := range []string{"stone_age", "bronze_age"} {
		if err := ge.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
	}
	const campaigns = "Campaigns need Military Tactics first. Research it to send one."
	st := ge.GetState()
	if out := militaryProvider(st, 120); !strings.Contains(out, campaigns) {
		t.Errorf("the Army panel does not say campaigns are locked:\n%s", out)
	}
	if out := cmdCampaignList(ge).Message; !strings.Contains(out, campaigns) {
		t.Errorf("campaign list does not say campaigns are locked:\n%s", out)
	}
	// Railroads and Mercantilism are techs of later ages: not a word yet.
	// The Wheel is this age's, and trade routes wait for it.
	if out := tradeProvider(st, 120); strings.Contains(out, "Railroads") || strings.Contains(out, "Mercantilism") ||
		!strings.Contains(out, "Trade routes need The Wheel first. Research it to start one.") {
		t.Errorf("the Trade panel should name The Wheel and no tech of a later age:\n%s", out)
	}
	// Envoys and Drama belong to ages not in sight: the Factions panel and
	// the festival command say nothing of them yet. The Expeditions panel
	// already lists Scout Nearby Ruins, so it names Exploration, a tech of
	// the next age, at once.
	festival := func() string { return cmdFestival(nil, ge).Message + cmdFestival([]string{"confirm"}, ge).Message }
	if out := factionsProvider(st, 120) + festival(); strings.Contains(out, "Envoys") || strings.Contains(out, "Drama") {
		t.Errorf("a panel names a tech of an age not in sight:\n%s", out)
	}
	if out := expeditionsProvider(st, 120); !strings.Contains(out, "Expeditions past the Scout Party need Exploration first. Research it to send one.") {
		t.Errorf("the Expeditions panel lists Scout Nearby Ruins and does not say it waits for Exploration:\n%s", out)
	}
	ge.GrantTechsForTest("military_tactics")
	if out := militaryProvider(ge.GetState(), 120) + cmdCampaignList(ge).Message; strings.Contains(out, "Military Tactics first") {
		t.Errorf("with Military Tactics researched the lock is still shown:\n%s", out)
	}
	for _, a := range []string{"iron_age", "classical_age", "medieval_age", "renaissance_age", "colonial_age", "industrial_age"} {
		if err := ge.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
	}
	out := tradeProvider(ge.GetState(), 120)
	for _, want := range []string{"The Rail Freight route needs Railroads first. Research it to start it.", "The black market needs Mercantilism first. Research it to deal there."} {
		if !strings.Contains(out, want) {
			t.Errorf("the Trade panel is missing %q:\n%s", want, out)
		}
	}
	// Their ages reached: each panel names the tech its command waits for,
	// and stops once it is researched.
	late := ge.GetState()
	for what, c := range map[string][2]string{
		"the Expeditions panel": {expeditionsProvider(late, 120), "Expeditions past the Scout Party need Exploration first. Research it to send one."},
		"the Factions panel":    {factionsProvider(late, 120), "Gifts, alliances, rivalries and deals need Envoys first. Research it to deal with other civilizations."},
		"festival":              {cmdFestival(nil, ge).Message, "Festivals need Drama first. Research it to hold one."},
		"festival confirm":      {cmdFestival([]string{"confirm"}, ge).Message, "Festivals need Drama first. Research it to hold one."},
	} {
		if !strings.Contains(c[0], c[1]) {
			t.Errorf("%s is missing %q:\n%s", what, c[1], c[0])
		}
	}
	ge.GrantTechsForTest("exploration", "envoys", "drama", "the_wheel")
	late = ge.GetState()
	if out := expeditionsProvider(late, 120) + factionsProvider(late, 120) + festival() + tradeProvider(late, 120); strings.Contains(out, "Exploration first") || strings.Contains(out, "Envoys first") || strings.Contains(out, "Drama first") || strings.Contains(out, "The Wheel first") {
		t.Errorf("with the techs researched a lock is still shown:\n%s", out)
	}
	if strings.Contains(out, "Interstellar Trade") || strings.Contains(out, "interstellar_trade") {
		t.Errorf("the Trade panel mentions a lock whose tech is not in the tree:\n%s", out)
	}
}

// medievalGame is a game in the Medieval Age with the Iron Era's two
// capstones in different states: Scholasticism can start (Alchemy and
// Theology are held) and Guilds waits (it has Civil Engineering and not
// Metal Casting).
func medievalGame(t *testing.T) *game.GameEngine {
	t.Helper()
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	for _, a := range []string{"stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age"} {
		if err := ge.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
	}
	ge.GrantTechsForTest("language", "tool_making", "stoneworking", "primitive_writing", "woodworking", "masonry", "the_wheel",
		"mathematics", "road_building", "philosophy", "civil_engineering", "alchemy", "theology")
	ge.SetStockForTest("knowledge", 1e9)
	return ge
}

// electricGame is a game in the Electric Age with nothing researched: the
// twelve ages a content batch has filled in are in sight (the Atomic Age is
// next).
func electricGame(t *testing.T) *game.GameEngine {
	t.Helper()
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	for _, a := range config.AgeOrder()[1:] {
		if err := ge.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
		if a == "electric_age" {
			break
		}
	}
	return ge
}

// TestEmblemsSitOnTheGrid: every tech of an age a content batch has filled
// in (the Stone, Iron, Steel and Electric Eras) wears the emblem its own
// entry sets, one glyph on the centre cell of its badge at both badge sizes
// and of its pill zoomed out, and in the plain glyph tier the first letter
// of its code sits there instead. Nothing is drawn in a cell beside the
// emblem's own. Once from the Classical Age, where the first six ages are in
// sight, and once from the Electric Age, where all twelve are.
func TestEmblemsSitOnTheGrid(t *testing.T) {
	for _, c := range []struct {
		name string
		st   game.GameState
		last string
		want int
	}{
		{"the Classical Age", classicalGame(t).GetState(), "medieval_age", 41},
		{"the Electric Age", electricGame(t).GetState(), "atomic_age", 97},
	} {
		emblemsSitOnTheGrid(t, c.name, c.st, c.last, c.want)
	}
}

func emblemsSitOnTheGrid(t *testing.T, name string, st game.GameState, last string, wantNodes int) {
	t.Helper()
	early := 0
	for _, def := range config.Technologies() {
		if config.AgePositions(config.AgeOrder())[def.Age] <= config.AgePositions(config.AgeOrder())[last] {
			early++
		}
	}
	lanes := config.TechLaneByKey()
	for _, w := range []int{80, 120} {
		for _, far := range []bool{false, true} {
			for _, ascii := range []bool{false, true} {
				geom := treeGeomFor(w, far)
				m := buildTree(st, geom, ascii, "")
				g := m.canvas(st, "")
				if ascii {
					foldPlain(g)
				}
				if len(m.nodes) != early || early != wantNodes {
					t.Fatalf("from %s, %d far=%v: %d techs on the map, the ages in sight hold %d (want %d)", name, w, far, len(m.nodes), early, wantNodes)
				}
				for _, n := range m.nodes {
					want := []rune(n.def.Emblem)[0]
					if ascii {
						want = rune(n.def.Code[0])
					}
					if n.emblem != want {
						t.Errorf("%d far=%v plain=%v: %s carries %q, want %q", w, far, ascii, n.key, n.emblem, want)
					}
					if !ascii && n.def.Emblem == lanes[n.def.Lane].Emblem && n.key != "tool_making" && n.key != "primitive_writing" && n.key != "military_tactics" {
						t.Errorf("%s wears its lane's glyph %q: a tech of a filled-in age has an emblem of its own", n.key, n.def.Emblem)
					}
					ex, ey := n.cx, n.top+geom.nameH+geom.badgeH/2
					if geom.pill {
						ex, ey = m.left(n)+1, n.top
					}
					if got := g.at(ex, ey).r; got != want {
						t.Errorf("%d far=%v plain=%v: %s has %q on its emblem's cell (%d,%d), want %q", w, far, ascii, n.key, got, ex, ey, want)
					}
					if ascii && (want < 'A' || want > 'Z') {
						t.Errorf("%s has no letter for the plain tier: %q", n.key, want)
					}
					// The cells either side belong to the badge, never to a
					// glyph that spilled over: a close badge keeps them
					// blank or shaded.
					if !geom.pill {
						for _, dx := range []int{-1, 1} {
							if r := g.at(ex+dx, ey).r; r != ' ' && r != '░' && r != '.' && r != ':' {
								t.Errorf("%d plain=%v: %s has %q beside its emblem", w, ascii, n.key, r)
							}
						}
					}
				}
			}
		}
	}
}

// TestCapstonesReadByShape: a capstone's badge is drawn in half blocks, at
// both badge sizes, and its state shows without colour like any other
// tech's: whole blocks when it can start, broken ones while it waits, a bar
// along the bottom while it is researched, a tick once it is done. Its card
// and the status line call it a capstone.
func TestCapstonesReadByShape(t *testing.T) {
	ge := medievalGame(t)
	st := ge.GetState()
	if s, g := st.Research.Techs["scholasticism"], st.Research.Techs["guilds"]; s.Kind != config.TechCapstone || g.Kind != config.TechCapstone || !s.PrereqsMet || g.PrereqsMet {
		t.Fatalf("setup: Scholasticism %+v, Guilds %+v", s, g)
	}
	for _, w := range []int{80, 120} {
		geom := treeGeomFor(w, false)
		m := buildTree(st, geom, false, "scholasticism")
		ready, waiting := m.by["scholasticism"], m.by["guilds"]
		if ready.mark != markReady || waiting.mark != markLocked {
			t.Fatalf("%d: Scholasticism is %v and Guilds %v, want ready and waiting", w, ready.mark, waiting.mark)
		}
		rr, wr := badgeRows(geom, ready, 0), badgeRows(geom, waiting, 0)
		for _, row := range append(append([]string(nil), rr...), wr...) {
			for _, r := range row {
				if strings.ContainsRune("╭╮╰╯╔╗╚╝─═│║┄┆", r) {
					t.Errorf("%d: a capstone's badge has the line %q in it: %q", w, r, row)
				}
			}
		}
		if !strings.Contains(rr[0], "▄▄▄▄▄▄▄") || !strings.Contains(wr[0], "▄ ▄ ▄ ▄") || rr[len(rr)-1] == wr[len(wr)-1] {
			t.Errorf("%d: a capstone that can start and one that waits are drawn alike:\n%s\n%s", w, strings.Join(rr, "\n"), strings.Join(wr, "\n"))
		}
		// An optional tech of the same age keeps its rounded frame.
		if rows := badgeRows(geom, m.by["feudalism"], 0); !strings.ContainsAny(rows[0], "╭┄") {
			t.Errorf("%d: Feudalism, an optional tech, is drawn %q", w, rows[0])
		}
	}
	if got := buildTree(st, treeGeomFor(120, false), false, "guilds").statusLine(st, treeView{sel: "guilds"}); !strings.Contains(got, "waits for what it needs") {
		t.Errorf("the status line for Guilds: %q", got)
	}
	if card := cardOf(st, "guilds"); !strings.Contains(card, "Guilds gives") && !strings.Contains(card, "With it, buildings cost 4% less and construction takes 8% less time") {
		t.Errorf("Guilds's card does not say what it does:\n%s", card)
	}

	// In progress: the bar; then done: the tick.
	if err := ge.StartResearch("scholasticism"); err != nil {
		t.Fatal(err)
	}
	st = ge.GetState()
	geom := treeGeomFor(120, false)
	m := buildTree(st, geom, false, "scholasticism")
	n := m.by["scholasticism"]
	if rows := badgeRows(geom, n, 0.4); n.mark != markRunning || !strings.Contains(rows[len(rows)-1], "▓▓░░░") {
		t.Errorf("a capstone in progress is drawn %q", rows[len(rows)-1])
	}
	g := m.canvas(st, "scholasticism")
	if got := g.at(m.left(n)+geom.badgeW-1, n.top+geom.nameH).r; got != '⟳' {
		t.Errorf("a capstone in progress carries %q at its top right, want the running mark", got)
	}
	// It takes the long share of the age's research cap.
	if def := config.TechByKey()["scholasticism"]; st.Research.TotalTicks <= config.TechByKey()["alchemy"].ResearchTicks || st.Research.TotalTicks > def.ResearchTicks {
		t.Errorf("Scholasticism runs %d ticks; its base is %d and an optional tech's %d", st.Research.TotalTicks, def.ResearchTicks, config.TechByKey()["alchemy"].ResearchTicks)
	}
}
