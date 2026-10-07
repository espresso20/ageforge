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
	ge.GrantTechsForTest("tool_making", "fire_mastery", "stoneworking", "primitive_writing", "pottery", "bronze_working", "currency", "masonry", "mathematics", "iron_smelting")
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
func TestResearchTreeDrawsEveryTechWhole(t *testing.T) {
	st := classicalGame(t).GetState()
	for _, w := range []int{80, 120} {
		for _, far := range []bool{false, true} {
			geom := treeGeomFor(w, far)
			m := buildTree(st, geom, false, "road_building")
			g := m.canvas(st, "road_building")
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
func TestResearchTreeAtEverySize(t *testing.T) {
	st := classicalGame(t).GetState()
	for _, size := range treeSizes {
		w, h := size[0], size[1]-promptRows
		for _, far := range []bool{false, true} {
			for _, ascii := range []bool{false, true} {
				for _, sel := range []string{"tool_making", "road_building", "philosophy", "siege_warfare"} {
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
	if m.later != len(hidden) || !strings.Contains(g.String(), "53 more techs wait in later ages") {
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
		"road_building":    {"Road Building gives +8% gold production and trade routes take 15% less time.", "It builds on Masonry.", "It is item 1 in your plan."},
		"military_tactics": {"gives +15% military power.", "It opens campaigns.", "It builds on Bronze Working.", "adds it to your plan. Philosophy is running, so it starts in"},
		"mathematics":      {"It is this age's keystone: the Colosseum cannot be built without it.", "You already hold this."},
		"philosophy":       {"It is being researched:", "Type research cancel to stop it"},
		"imperial_legions": {"It builds on Siege Warfare and Iron Smelting.", "It starts once you hold what it needs."},
		"theology":         {"the Great Library cannot be built without it.", "It starts once you reach the Medieval Age and hold what it needs."},
		"tool_making":      {"gives +10% food production, +10% wood production and gathering by hand brings 2 more.", "It needs nothing before it."},
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
