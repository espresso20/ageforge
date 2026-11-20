//go:build mapcapture

package ui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// The wiki's screens (site/docs/screens/) are real ones: the game's own
// panels, drawn on a simulated terminal from one staged game, and written as
// rows of text with color runs. site/docs/screens.js draws them in the wiki,
// and TestDocScreens (docs_screens_test.go) holds the files to the figures
// that name them. Regenerate them all with
//
//	go test -tags mapcapture -run TestWriteSiteScreens ./ui
//
// It takes under a minute and writes the same bytes every time: the seed is
// fixed, the map's animation clock is pinned, toasts are held, and nothing
// in frame depends on the wall clock (the Statistics page's play time tile
// shows the whole seconds of the short run, the same on each pass).
// SITE_SCREENS_OUT picks another folder, and
// SITE_SCREENS_TEXT a folder for plain-text copies to read in a terminal.
//
// The game behind the pictures is played, not painted: one seeded run from
// the Primitive Age to the Medieval Age through the engine's own methods and
// the prompt's own commands, with every building paid for and waited for,
// every wonder banked and every age advanced for real. Each gate is met as the
// game lists it (meetGate reads the Next Age list), not from counts typed here. Staging takes three
// liberties, and no others:
//
//   - stores are topped up between purchases (stock), which stands in for
//     the hours of waiting, and set to mixed levels before a picture (level);
//   - the Iron Era's secret fate is kept quiet while its town is built, then
//     a doom is fated on cue, so the harbinger and the catastrophe arrive
//     through the game's own path when the pictures need them;
//   - if that doom's seeded strike misses, the catastrophe is brought on by
//     the engine's test hook, so the page still has its picture.

const (
	// A panel is drawn at 100x30: the size the game's own small-terminal
	// checks use, and room enough for every panel to read as it should.
	sitePanelW, sitePanelH = 100, 30
	// The dashboard is drawn at 120x40, the size the wiki calls a large
	// terminal: every box writes its rows for the width it is drawn at
	// (resources_box.go, tab_economy.go), so nothing wraps there, the mini
	// map shows, and the Bronze Age's seven resources each have a row.
	siteDashW, siteDashH = 120, 40
	// The themes page shows two screens side by side, so those two are the
	// game at its smallest size.
	sitePairW, sitePairH = 80, 24

	siteScreensSeed = 20261007

	// siteMenuVersion is the version the main menu's pictures carry: the
	// release the menu was drawn for. (A test build has none of its own.)
	siteMenuVersion = "v4.0.0"
	// siteMenuGame is the name the menu's picture gives the staged game.
	siteMenuGame = "Meridia"
)

// siteStage is one staged game and the dashboard that draws it.
type siteStage struct {
	// onTree takes the tech tree's pictures, part way through the Classical
	// Age (toMedieval calls it).
	onTree func()
	// onIron takes the picture of an epoch's arrival, as the advance into
	// the Iron Age opens the Iron Era (toIron calls it).
	onIron func()
	t      *testing.T
	eng    *game.GameEngine
	d      *Dashboard
	app    *appLoop
	pages  *nativePages
}

func newSiteStage(t *testing.T) *siteStage {
	t.Helper()
	eng := game.NewGameEngine()
	eng.SeedRNG(siteScreensSeed)
	d, pages := newTestDashboard(eng)
	// The maps animate on the clock: hold both at their first frame.
	still := time.Unix(0, 0)
	d.mapPanel.start, d.mapPanel.now = still, func() time.Time { return still }
	d.miniMap.start, d.miniMap.now = still, func() time.Time { return still }
	d.expPanel.start, d.expPanel.now = still, func() time.Time { return still }
	// The arrival screen reads the game through the engine, as it does in
	// the app (app.go). Its timers are held: nothing here waits for them.
	d.overlayMgr.engine = eng
	d.overlayMgr.after = func(time.Duration, func()) func() { return func() {} }
	return &siteStage{t: t, eng: eng, d: d, app: d.app, pages: pages}
}

// say types a command at the prompt and presses Enter, as a player would
// (Dashboard.submitInput), and fails if the game refuses it.
func (s *siteStage) say(line string) {
	s.t.Helper()
	before := len(s.eng.GetLogs())
	s.d.inputField.SetText(line)
	s.d.submitInput()
	logs := s.eng.GetLogs()
	if before > len(logs) {
		before = 0
	}
	for _, l := range logs[before:] {
		if l.Type == "error" {
			s.t.Fatalf("%q was refused: %s", line, l.Message)
		}
	}
}

// reach is the stock the staging tops res up to: the age's stock scale (a
// good stock of it), or twice the dearest next copy of anything if that is
// more.
func (s *siteStage) reach(st game.GameState, res string) float64 {
	top := config.StockScale(st.Age)
	for _, b := range st.Buildings {
		top = max(top, 2*b.NextCost[res])
	}
	return top
}

// stock tops every unlocked resource up to its reach. It stands in for the
// waiting a player does between purchases: everything else is bought, built
// and waited for.
func (s *siteStage) stock() {
	st := s.eng.GetState()
	for k, rs := range st.Resources {
		if rs.Unlocked {
			if top := s.reach(st, k); top > rs.Amount {
				s.eng.Resources.Add(k, top-rs.Amount)
			}
		}
	}
}

// level sets each named resource to a share of its reach, so the picture does
// not show every resource at the top after stock.
func (s *siteStage) level(shares map[string]float64) {
	st := s.eng.GetState()
	for _, k := range sortedKeysOf(shares) {
		rs, ok := st.Resources[k]
		if !ok || !rs.Unlocked {
			s.t.Fatalf("level: no unlocked resource %q", k)
		}
		want := float64(s.reach(st, k) * shares[k])
		if want < rs.Amount {
			s.eng.Resources.Remove(k, rs.Amount-want)
		} else {
			s.eng.Resources.Add(k, want-rs.Amount)
		}
	}
}

// wait runs the game n ticks.
func (s *siteStage) wait(n int) { s.eng.StepTicks(n) }

// finish runs the game until nothing is under construction.
func (s *siteStage) finish() {
	s.t.Helper()
	for i := 0; len(s.eng.GetState().BuildQueue) > 0; i++ {
		if i > 200000 {
			s.t.Fatal("the build queue never emptied")
		}
		s.eng.StepTicks(1)
	}
}

// raise builds n copies of a building one after another, each paid for from
// topped-up stock and waited for.
func (s *siteStage) raise(key string, n int) {
	s.t.Helper()
	for i := 0; i < n; i++ {
		s.stock()
		if err := s.eng.BuildBuilding(key); err != nil {
			s.t.Fatalf("building %s (copy %d of %d): %v", key, i+1, n, err)
		}
		s.finish()
	}
}

// wonder banks and builds the current age's wonder.
func (s *siteStage) wonder(key string) {
	s.t.Helper()
	for i := 0; i < 400 && !s.eng.GetState().Buildings[key].WonderBankFull; i++ {
		s.stock()
		for res := range s.eng.GetState().Buildings[key].NextCost {
			_, _ = s.eng.BankWonderMax(key, res)
		}
	}
	s.raise(key, 1)
}

// advance takes the game into the next age through the real advance.
func (s *siteStage) advance() {
	s.t.Helper()
	s.stock()
	if err := s.eng.AdvanceAge(); err != nil {
		s.t.Fatalf("advancing from %s: %v", s.eng.GetState().Age, err)
	}
}

// learn researches techs one after another, each to the end.
func (s *siteStage) learn(keys ...string) {
	s.t.Helper()
	for _, k := range keys {
		s.stock()
		if err := s.eng.StartResearch(k); err != nil {
			s.t.Fatalf("researching %s: %v", k, err)
		}
		for i := 0; s.eng.GetState().Research.CurrentTech != ""; i++ {
			if i > 200000 {
				s.t.Fatalf("research of %s never finished", k)
			}
			s.eng.StepTicks(1)
		}
	}
}

// quiet clears the toast line, so what a picture shows there comes from what
// happens next and not from the staging before.
func (s *siteStage) quiet() {
	s.d.toastMgr = NewToastManager()
}

// holdToast keeps the newest toast (if any) on screen: toasts queue up and
// expire on the wall clock, which the picture must not depend on.
func (s *siteStage) holdToast() {
	tm := s.d.toastMgr
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if n := len(tm.queue); n > 0 {
		tm.current = &tm.queue[n-1]
	}
	tm.queue = nil
	if tm.current != nil {
		tm.current.Expiry = time.Now().Add(24 * time.Hour)
	}
}

// open shows a panel through its command. With a heading it then scrolls
// the panel down, as the arrow keys would, until the line that holds the
// heading is the panel's first.
func (s *siteStage) open(cmd, heading string) {
	s.t.Helper()
	s.say(cmd)
	name := s.d.overlayMgr.ActiveName()
	if name == "" {
		s.t.Fatalf("%q opened no panel", cmd)
	}
	if heading == "" {
		return
	}
	e, ok := s.d.overlayMgr.entries[name]
	if !ok {
		s.t.Fatalf("the %s panel does not scroll", name)
	}
	_, y, _, _ := e.tv.GetInnerRect()
	line := func(cells []simCell, y int) string {
		var b strings.Builder
		for x := 0; x < sitePanelW; x++ {
			if c := cells[y*sitePanelW+x]; len(c.Runes) > 0 {
				b.WriteRune(c.Runes[0])
			}
		}
		return b.String()
	}
	last := ""
	for row := 0; row < 2000; row++ {
		e.tv.ScrollTo(row, 0)
		cells, _, _, _ := s.draw(sitePanelW, sitePanelH)
		_, y, _, _ = e.tv.GetInnerRect()
		if strings.Contains(line(cells, y), heading) {
			return
		}
		var all strings.Builder
		for yy := 0; yy < sitePanelH; yy++ {
			all.WriteString(line(cells, yy))
		}
		if all.String() == last {
			// The panel is at its end: the heading only has to be in sight.
			if strings.Contains(last, heading) {
				return
			}
			break
		}
		last = all.String()
	}
	s.t.Fatalf("the %s panel has no line with %q", name, heading)
}

// close closes whatever panel is open.
func (s *siteStage) close() { s.d.overlayMgr.Hide() }

// draw refreshes the dashboard from the engine and draws the whole screen at
// w x h, the way the app does on its own screen. It returns the cells and the
// pictures the frame drew.
func (s *siteStage) draw(w, h int) ([]simCell, []kit.Pic, theme.Color, theme.Color) {
	s.t.Helper()
	sim := newSim()
	must(s.t, sim.Init())
	defer sim.Fini()
	sim.SetSize(w, h)
	screen := sim
	s.d.overlayMgr.screenW = w
	// Twice: the first draw lays the panels out, and a refresh after it
	// fills in what depends on their sizes (the mini map, wrapped text).
	var cells []simCell
	var pics []kit.Pic
	for i := 0; i < 2; i++ {
		s.d.refresh()
		screen.Clear()
		s.pages.SetRect(0, 0, w, h)
		drawOn(screen, s.pages)
		screen.Show()
		got, _, _ := sim.GetContents()
		cells = append(cells[:0], got...)
		pics = append(pics[:0], sim.buf.Pics...)
	}
	return cells, pics, theme.Of(theme.RoleText), theme.Of(theme.RoleBackground)
}

// --- the staged game ------------------------------------------------------------

type siteBuild struct {
	key string
	n   int
}

func (s *siteStage) raiseAll(list ...siteBuild) {
	s.t.Helper()
	for _, b := range list {
		s.raise(b.key, b.n)
	}
}

// meetGate builds what the gate out of the current age asks for, as the game
// lists it (NextAgeBldReqs): every building to the count it asks. The wonder
// is not in that list, and each age's keystone must be learned before this
// runs, so every building on the list can be built. What the staging builds
// for the pictures comes first; this tops it up to the gate and no further.
func (s *siteStage) meetGate() {
	s.t.Helper()
	st := s.eng.GetState()
	for _, k := range sortedKeysOf(st.NextAgeBldReqs) {
		if need := st.Buildings[k].NeedsTech; need != "" {
			s.t.Fatalf("the gate out of %s asks for %s, which waits for %s: learn it before meetGate", st.Age, k, need)
		}
		s.raiseTo(k, st.NextAgeBldReqs[k])
	}
}

// raiseTo builds a building until there are n of it.
func (s *siteStage) raiseTo(key string, n int) {
	s.t.Helper()
	if have := s.eng.GetState().Buildings[key].Count; have < n {
		s.raise(key, n-have)
	}
}

// until runs the game, stores kept full, until done reports true.
func (s *siteStage) until(what string, done func(st game.GameState) bool) {
	s.t.Helper()
	for i := 0; !done(s.eng.GetState()); i++ {
		if i > 20000 {
			st := s.eng.GetState()
			var short []string
			for _, k := range sortedKeysOf(st.NextAgeBldReqs) {
				if st.Buildings[k].Count < st.NextAgeBldReqs[k] {
					short = append(short, fmt.Sprintf("%s %d of %d", k, st.Buildings[k].Count, st.NextAgeBldReqs[k]))
				}
			}
			s.t.Fatalf("staging never got to: %s (in %s; wonder left %q; short of: %s)", what, st.Age, st.CurrentAgeWonderKey, strings.Join(short, ", "))
		}
		s.stock()
		s.eng.StepTicks(2)
	}
}

// dismiss lets the dashboard put up whatever it has waiting (the arrival
// screen of an advance) and closes it.
func (s *siteStage) dismiss() {
	s.d.refresh()
	s.close()
}

// arrive lets the dashboard put up the arrival screen it has waiting, and
// holds the screen's clock at a frame of its celebration.
func (s *siteStage) arrive(frame int) {
	s.t.Helper()
	s.d.refresh()
	a := s.d.overlayMgr.arrival
	if a == nil {
		s.t.Fatal("no arrival screen came up")
	}
	still := time.Unix(0, 0)
	a.start, a.base = still, 0
	a.now = func() time.Time { return still.Add(time.Duration(frame) * mapAnimStep) }
}

// endFilm lets the dashboard put up the film of the run that just ended,
// and holds its clock at a frame: at is given the film's beats and says
// which.
func (s *siteStage) endFilm(at func(spans []endSpan) int) {
	s.t.Helper()
	s.d.refresh()
	e := s.d.overlayMgr.film
	if e == nil {
		s.t.Fatal("no film came up")
	}
	frame := at(e.spans())
	still := time.Unix(0, 0)
	e.start, e.base, e.sc = still, 0, nil
	e.now = func() time.Time { return still.Add(time.Duration(frame) * mapAnimStep) }
}

// inform moves the arrival screen on from its celebration to what the age
// opens, as any key does.
func (s *siteStage) inform() {
	s.t.Helper()
	a := s.d.overlayMgr.arrival
	if a == nil {
		s.t.Fatal("no arrival screen is up")
	}
	a.key()
	if a.stage != arrInforming {
		s.t.Fatal("a key did not move the arrival screen on")
	}
}

// toBronze plays a new game through the Primitive and Stone Ages, each age's
// buildings bought and built in turn and its wonder raised, and makes both
// advances for real. It stops on the Bronze Age's first tick.
func (s *siteStage) toBronze() {
	s.raiseAll(siteBuild{"stash", 5}, siteBuild{"gathering_camp", 4}, siteBuild{"wood_camp", 4},
		siteBuild{"hut", 10}, siteBuild{"story_circle", 5}, siteBuild{"shrine", 2})
	s.learn("tool_making", "fire_mastery", "language")
	s.meetGate()
	s.wonder("sacred_grove")
	s.advance()
	s.dismiss() // the Stone Age's arrival, seen and closed, as in play
	// The Standing Stones wait for Ritual: research comes first, as it
	// would in play.
	s.raiseAll(siteBuild{"storage_pit", 5}, siteBuild{"longhouse", 15}, siteBuild{"forager_post", 4},
		siteBuild{"woodcutter_camp", 4}, siteBuild{"stone_camp", 3}, siteBuild{"stone_pit", 5},
		siteBuild{"elders_hall", 5}, siteBuild{"war_camp", 2})
	s.learn("ritual")
	s.raiseAll(siteBuild{"standing_stones", 2})
	s.learn("stoneworking", "pottery", "primitive_writing", "woodworking")
	s.meetGate()
	s.wonder("great_monolith")
	s.advance()
}

// bronzeTown builds the Bronze Age town the wiki's main pictures show, then
// plays a short session at the prompt: a tech finished and the next one well
// under way, a trade route, scouts on the road, two farms going up and a plan
// that is part started, part waiting.
func (s *siteStage) bronzeTown() {
	s.raiseAll(siteBuild{"warehouse", 3}, siteBuild{"house", 8}, siteBuild{"farm", 5}, siteBuild{"lumber_mill", 4},
		siteBuild{"quarry", 3}, siteBuild{"scriptorium", 3}, siteBuild{"market", 2}, siteBuild{"smithy", 2})
	// Calendar opens the Altar and The Wheel opens trade routes. The
	// Barracks wait for Military Tactics, which this town does not hold
	// yet (toIron builds them).
	s.learn("calendar", "the_wheel")
	s.raiseAll(siteBuild{"altar", 2})
	s.wait(300)
	s.stock()
	s.quiet()
	s.say("research bronze_working")
	for s.eng.GetState().Research.CurrentTech != "" {
		s.wait(1)
	}
	s.say("research masonry")
	s.say("trade route start local_barter")
	s.wait(310)
	s.level(map[string]float64{"food": 0.71, "wood": 0.34, "stone": 0.83, "gold": 0.37, "iron": 0.22, "faith": 0.41})
	// The Scout Party needs no tech; scouting past it waits for Exploration.
	s.say("expedition scouting_party")
	s.say("wonder collect stone")
	s.say("wonder collect wood 9000")
	s.say("build farm 2")
	s.say("plan build house 6")
	s.say("plan research currency")
	s.say("plan build lumber_mill 4")
	s.say("plan build quarry 5")
	s.say("plan advance")
	// Long enough for the plan to start what it can afford and be part way
	// to the next house.
	s.wait(28)
	s.holdToast()
}

// toIron finishes the Bronze Age (its plan, its wonder, the buildings the
// next age asks for) and advances into the Iron Age and the Iron Era, then
// builds that age's town, trains a garrison and sends scouts out until three
// civilizations have been met.
func (s *siteStage) toIron() {
	s.until("the plan's builds and research done", func(st game.GameState) bool {
		return len(st.BuildQueue) == 0 && st.Research.CurrentTech == "" && len(st.Plan) == 1
	})
	s.raiseTo("scriptorium", 5)
	s.raiseTo("warehouse", 5)
	s.learn("animal_husbandry", "agriculture", "military_tactics", "map_making")
	s.raiseAll(siteBuild{"barracks", 2})
	s.meetGate()
	s.wonder("stonehenge")
	// The last item of the plan is the advance: it goes by itself.
	s.until("the planned advance", func(st game.GameState) bool { return st.Age == "iron_age" })
	if s.onIron != nil {
		s.onIron()
	}
	s.dismiss()
	// From the Iron Era on a doom may be fated in secret. The era is kept
	// quiet while the town is built, and the doom the pictures need is fated
	// when they need it.
	must(s.t, s.eng.ForceQuietFateForTest("iron_era"))
	s.raiseAll(siteBuild{"granary", 5}, siteBuild{"townhouse", 6}, siteBuild{"field_works", 4}, siteBuild{"timber_yard", 4},
		siteBuild{"marble_quarry", 3}, siteBuild{"agora", 7}, siteBuild{"temple", 2}, siteBuild{"hunting_lodge", 4},
		siteBuild{"trading_post", 3}, siteBuild{"ironworks", 2})
	// The Smelter and the Legion Fort wait for their techs, and scouting
	// past the Scout Party for Exploration.
	s.learn("iron_smelting", "siege_warfare", "exploration")
	s.raiseAll(siteBuild{"legion_fort", 3}, siteBuild{"smelter", 2})
	s.scout(1)
}

// toMedieval endures the catastrophe, rebuilds, and plays through the
// Classical Age into the Medieval Age, where prestige opens and a second
// civilization can be met.
func (s *siteStage) toMedieval() {
	if err := s.eng.Endure(); err != nil {
		s.t.Fatalf("enduring: %v", err)
	}
	s.raiseTo("granary", 5)
	s.raiseTo("agora", 7)
	s.raiseTo("hunting_lodge", 15)
	s.raiseTo("trading_post", 10)
	s.learn("mathematics", "road_building")
	s.meetGate()
	s.wonder("colosseum")
	s.advance()
	s.dismiss()
	s.raiseAll(siteBuild{"classical_vault", 5}, siteBuild{"villa", 6}, siteBuild{"estate_farm", 4}, siteBuild{"wood_workshop", 3},
		siteBuild{"marble_works", 3}, siteBuild{"library", 8}, siteBuild{"oracle_house", 2}, siteBuild{"military_academy", 15},
		siteBuild{"merchant_quarter", 5}, siteBuild{"aqueduct", 2}, siteBuild{"amphitheater", 2})
	// The Forge waits for Metal Casting.
	s.learn("metal_casting")
	s.raiseAll(siteBuild{"forge", 2})
	// The tech tree's pictures are taken here, with Philosophy under way:
	// the map then has techs researched, in progress, ready to start and
	// waiting for what they need on screen at once, and the next age dim.
	s.stock()
	if err := s.eng.StartResearch("philosophy"); err != nil {
		s.t.Fatalf("researching philosophy: %v", err)
	}
	s.wait(300)
	if s.onTree != nil {
		s.onTree()
	}
	s.until("Philosophy researched", func(st game.GameState) bool { return st.Research.CurrentTech == "" })
	s.learn("civil_engineering")
	s.meetGate()
	s.wonder("parthenon")
	s.advance()
	s.dismiss()
	// Five keeps, as the gate asks: Feudalism's price (the age sets it) has to fit the stores.
	s.raiseAll(siteBuild{"keep", 5}, siteBuild{"manor", 5}, siteBuild{"demesne", 3}, siteBuild{"sawmill", 3},
		siteBuild{"stonemasons_guild", 2}, siteBuild{"monastery_library", 3},
		siteBuild{"castle_keep", 2}, siteBuild{"guildhall", 3}, siteBuild{"workshop", 2}, siteBuild{"ironmonger", 2})
	// The Cathedral waits for Theology, the age's keystone, and Feudalism
	// stands on The Plough.
	s.learn("theology", "irrigation", "the_plough", "feudalism")
	s.raiseAll(siteBuild{"cathedral", 2})
	s.scout(2)
}

// scout sends scouting parties out, one after another, until n civilizations
// have been met.
func (s *siteStage) scout(n int) {
	s.t.Helper()
	met := func() int {
		c := 0
		for _, f := range s.eng.GetState().Diplomacy.Factions {
			if f.Discovered {
				c++
			}
		}
		return c
	}
	for i := 0; met() < n; i++ {
		if i > 120 {
			s.t.Fatalf("120 scouting parties met only %d civilizations", met())
		}
		s.stock()
		if err := s.eng.LaunchExpedition(1); err != nil {
			s.t.Fatalf("scouting: %v", err)
		}
		for s.eng.GetState().Military.ActiveScout != nil {
			s.wait(1)
		}
	}
}

// --- the screens ------------------------------------------------------------------

// siteShots writes the pictures.
type siteShots struct {
	t       *testing.T
	out     string
	textDir string
	bytes   map[string]int
	// town is the Bronze Age town's save, as the game writes it: the main
	// menu's picture shows it behind the menu.
	town []byte
}

// saveBytes saves the stage's game as the game does, returns the file's
// bytes and takes the file away again.
func (s *siteStage) saveBytes(name string) []byte {
	s.t.Helper()
	must(s.t, s.eng.SaveGame(name))
	path := filepath.Join(game.DataDir(), "saves", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		s.t.Fatal(err)
	}
	must(s.t, os.Remove(path))
	return data
}

// takeMenu draws the main menu of eng's account as it stands, at the
// dashboard's size and in the default theme, and writes it as name. The
// menu is the real page (menu.go), at rest: its clock is held, so the
// picture is the frame a still menu shows.
func (w *siteShots) takeMenu(eng *game.GameEngine, name string) {
	w.t.Helper()
	must(w.t, theme.SetActive(theme.DefaultKey))
	app, pages := newAppLoop(), newNativePages()
	m := newMainMenu(app, pages, eng, siteMenuVersion)
	still := time.Unix(0, 0)
	m.start, m.now = still, func() time.Time { return still }
	m.eliteAt, m.view.elite = -1, [3]string{}
	defer m.leave()

	cols, rows := siteDashW, siteDashH
	f := drawWrapped(w.t, m, cols, rows, 1)
	raw, text, err := encodeSiteScreen(f.Cells(), f.Pics(), cols, rows, theme.Of(theme.RoleText), theme.Of(theme.RoleBackground), false)
	if err != nil {
		w.t.Fatalf("%s: %v", name, err)
	}
	if _, dup := w.bytes[name]; dup {
		w.t.Fatalf("two screens are named %q", name)
	}
	w.bytes[name] = len(raw)
	must(w.t, os.WriteFile(filepath.Join(w.out, name+".json"), raw, 0o644))
	if w.textDir != "" {
		must(w.t, os.WriteFile(filepath.Join(w.textDir, name+".txt"), []byte(text), 0o644))
	}
	w.t.Logf("%-18s %3dx%-2d %6d bytes", name, cols, rows, len(raw))
}

// take draws the stage as it stands, at the capture size and in the default
// theme, and writes it as name.
func (w *siteShots) take(s *siteStage, name string) {
	w.t.Helper()
	w.takeIn(s, name, theme.DefaultKey, sitePanelW, sitePanelH)
}

// takeIn is take in another theme or size.
func (w *siteShots) takeIn(s *siteStage, name, themeKey string, cols, rows int) {
	w.t.Helper()
	must(w.t, theme.SetActive(themeKey))
	cells, pics, fg, bg := s.draw(cols, rows)
	raw, text, err := encodeSiteScreen(cells, pics, cols, rows, fg, bg, themeKey != theme.DefaultKey)
	if err != nil {
		w.t.Fatalf("%s: %v", name, err)
	}
	must(w.t, theme.SetActive(theme.DefaultKey))
	if _, dup := w.bytes[name]; dup {
		w.t.Fatalf("two screens are named %q", name)
	}
	w.bytes[name] = len(raw)
	must(w.t, os.WriteFile(filepath.Join(w.out, name+".json"), raw, 0o644))
	if w.textDir != "" {
		must(w.t, os.WriteFile(filepath.Join(w.textDir, name+".txt"), []byte(text), 0o644))
	}
	w.t.Logf("%-18s %3dx%-2d %6d bytes", name, cols, rows, len(raw))
}

func TestWriteSiteScreens(t *testing.T) {
	out := os.Getenv("SITE_SCREENS_OUT")
	if out == "" {
		out = filepath.Join("..", "site", "docs", "screens")
	}
	must(t, os.MkdirAll(out, 0o755))
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	keepTheme(t)
	prevDev := game.DevModeActive
	game.DevModeActive = false
	t.Cleanup(func() { game.DevModeActive = prevDev })
	must(t, theme.SetActive(theme.DefaultKey))
	shots := &siteShots{t: t, out: out, textDir: os.Getenv("SITE_SCREENS_TEXT"), bytes: map[string]int{}}

	s := newSiteStage(t)
	shots.takeIn(s, "new-game", theme.DefaultKey, siteDashW, siteDashH)
	s.toBronze()
	// The first refresh after an advance puts up the arrival screen: the
	// celebration, held here just after the strike with the town part way
	// through its turn, then what the age opens.
	s.arrive(arrAgeStrike + 3)
	shots.takeIn(s, "age-arrival", theme.DefaultKey, siteDashW, siteDashH)
	s.inform()
	shots.take(s, "age-advance")
	s.close()

	s.bronzeTown()
	shots.takeIn(s, "dashboard", theme.DefaultKey, siteDashW, siteDashH)
	// The town as it stands, for the main menu's picture (writeBadgeScreens).
	shots.town = s.saveBytes("town")
	for _, p := range []struct {
		name, cmd, heading string
	}{
		{"buildings", "buildings", "Bronze Age (current)"},
		{"plan", "plan", ""},
		{"rates", "stats", ""},
		{"people", "people", ""},
		{"advance", "charter", ""},
		{"trade", "trade", ""},
		{"wonders", "wonders", ""},
		{"milestones", "milestones", ""},
		{"expeditions", "expedition", ""},
		{"history", "history", ""},
		{"help", "help", ""},
	} {
		s.open(p.cmd, p.heading)
		shots.take(s, p.name)
		s.close()
	}
	// The Map: the close-up, as a large terminal draws it, and the whole town
	// in one view, as a panel of the usual size draws it.
	s.open("map", "")
	shots.takeIn(s, "map", theme.DefaultKey, siteDashW, siteDashH)
	shots.take(s, "map-region")
	s.close()
	// The Wonders panel's gallery: G on the list opens it.
	s.open("wonders", "")
	keyTo(s.d.overlayMgr.entries["wonders"].tv, tk(kit.KeyRune, 'g', 0))
	if !s.d.wonders.gallery {
		t.Fatal("G did not open the gallery")
	}
	shots.take(s, "wonders-gallery")
	s.close()
	// One panel in two themes, at the game's smallest size, for the themes page.
	s.open("people", "")
	shots.takeIn(s, "theme-forge", "forge", sitePairW, sitePairH)
	shots.takeIn(s, "theme-daylight", "daylight", sitePairW, sitePairH)
	s.close()

	// The advance into the Iron Age opens the Iron Era: an epoch's arrival,
	// held just after the heavy blow, with the screen still lit.
	s.onIron = func() {
		s.arrive(arrEraStrike + 1)
		shots.takeIn(s, "epoch-arrival", theme.DefaultKey, siteDashW, siteDashH)
	}
	s.toIron()
	s.quiet()
	s.level(map[string]float64{"food": 0.58, "wood": 0.77, "stone": 0.49, "gold": 0.66, "iron": 0.52, "marble": 0.28, "iron_ore": 0.44, "faith": 0.63})
	s.wait(30)

	// A doom is fated for this era, due now: its harbinger comes on the next
	// tick, and the doom waits out the shortest warning the game allows.
	fate := s.eng.FateForTest()
	if fate == nil {
		t.Fatal("the Iron Era has no fate")
	}
	must(t, s.eng.ForceFateForTest("iron_era", s.eng.GetState().Tick-fate.EntryTick))
	s.wait(1)
	if s.eng.GetState().Harbinger == nil {
		t.Fatal("no harbinger came")
	}
	s.open("harbinger", "")
	shots.take(s, "harbinger")
	s.close()

	s.quiet()
	s.until("the doom resolving", func(st game.GameState) bool { return st.Harbinger == nil })
	if s.eng.GetState().PendingCatastrophe == "" {
		// The seeded strike missed (it lands six to nine times in ten). The
		// page still needs its picture, so the catastrophe is brought on.
		t.Log("the fated doom was spared; forcing the catastrophe")
		must(t, s.eng.ForceCatastropheForTest())
	}
	s.level(map[string]float64{"food": 0.64, "wood": 0.81, "stone": 0.57, "gold": 0.73, "iron": 0.6, "marble": 0.35, "iron_ore": 0.5, "faith": 0.63})
	s.holdToast()
	// The dashboard puts the choice up by itself on its next refresh.
	shots.take(s, "catastrophe")
	s.d.closeCatastropheModal()

	s.onTree = func() {
		for _, p := range [][2]string{{"research", "research tree close"}, {"research-card", "research card civil_engineering"}, {"research-far", "research tree far"}} {
			var queued []string
			if p[0] == "research-far" {
				// The zoomed-out picture shows a queue: Fortification, a
				// tech of the next age, planned with the tech it still
				// needs. The engine is asked directly so the log stays as it
				// was, and the chain comes out again before the game goes
				// on.
				chain, err := s.eng.PlanAddResearchChain("fortification")
				if err != nil || len(chain) < 2 {
					t.Fatalf("queuing Fortification for the picture: %v, %v", chain, err)
				}
				queued = chain
			}
			s.open(p[1], "")
			shots.take(s, p[0])
			s.close()
			for range queued {
				plan := s.eng.GetState().Plan
				for i := len(plan) - 1; i >= 0; i-- {
					if plan[i].Kind == game.PlanResearch && slices.Contains(queued, plan[i].Key) {
						if _, err := s.eng.PlanRemove(i + 1); err != nil {
							t.Fatal(err)
						}
						break
					}
				}
			}
		}
		s.say("research tree close")
		s.close()
	}
	s.toMedieval()
	s.quiet()
	s.level(map[string]float64{"food": 0.47, "wood": 0.69, "stone": 0.74, "gold": 0.55, "iron": 0.61, "marble": 0.43,
		"iron_ore": 0.52, "steel": 0.26, "culture": 0.33, "faith": 0.72})
	s.wait(3)
	s.open("factions", "")
	shots.take(s, "factions")
	s.close()
	// Prestige has no panel of its own: the command answers in the log.
	s.say("prestige")
	s.open("logs", "")
	shots.take(s, "prestige")
	s.close()
	// The prestige itself: the staged game is given up for real, and its
	// film is held at two of its beats, the town part burnt down and the
	// word just struck.
	s.say("prestige confirm yes")
	s.endFilm(func([]endSpan) int { return endBurnHold + 9 })
	shots.takeIn(s, "prestige-ending", theme.DefaultKey, siteDashW, siteDashH)
	s.endFilm(func(spans []endSpan) int {
		for _, sp := range spans {
			if sp.beat == beatStrike {
				return sp.from + endBlow + 2
			}
		}
		t.Fatal("the film of a prestige has no strike")
		return 0
	})
	shots.takeIn(s, "prestige-strike", theme.DefaultKey, siteDashW, siteDashH)
	s.close()

	writeBadgeScreens(t, shots)

	// A screen dropped from the list leaves no file behind.
	old, err := filepath.Glob(filepath.Join(out, "*.json"))
	must(t, err)
	total := 0
	for _, f := range old {
		n, ok := shots.bytes[strings.TrimSuffix(filepath.Base(f), ".json")]
		if !ok {
			must(t, os.Remove(f))
		}
		total += n
	}
	t.Logf("wrote %d screens to %s: %d bytes", len(shots.bytes), out, total)
}

// writeBadgeScreens takes the Ranks screen's pictures: the ranks, and the
// all-badges list. Badges belong to an account, which the game above is played
// without, so these have a stage of their own: a new account (in the test's
// own data folder, never the player's) and a named game on it, played through
// the Primitive and Stone Ages the same way. One liberty, and no other: the
// account's ten prestiges are reported through the engine's test hook, which
// stands in for ten runs of play.
//
// The main menu's two pictures are taken on the same account: the first
// page before it has a game, and the town page at the end. One liberty
// there: the town behind the menu is the Bronze Age town of the first
// stage (the one the dashboard and the map pictures show), whose save is
// put in the account's slot under a name and marked as its main game. The
// menu itself is the real page, reading that save as it reads any other.
func writeBadgeScreens(t *testing.T, shots *siteShots) {
	t.Helper()
	day := time.Date(2026, time.September, 29, 12, 0, 0, 0, time.UTC)
	t.Cleanup(game.SetBadgeClockForTest(func() time.Time { return day }))

	s := newSiteStage(t)
	acct, err := game.CreateAccount("Ada")
	must(t, err)
	s.eng.SetAccount(acct)
	// A first visit: an account, and no game yet.
	shots.takeMenu(s.eng, "menu-first")
	must(t, s.eng.StartNewNamedGame("Rome"))
	s.eng.SeedRNG(siteScreensSeed)

	s.toBronze()
	s.dismiss() // the Bronze Age's splash
	s.level(map[string]float64{"food": 0.62, "wood": 0.48, "stone": 0.71})
	s.wait(3)
	for i := 0; i < 10; i++ {
		s.eng.ReportForTest(config.BadgeEvPrestige, "modern_age")
	}
	s.d.refresh() // the dashboard announces what was earned

	s.open("badges", "")
	if s.d.badgePanel.view.tab != tabRanks {
		t.Fatalf("badges opened the tab %d", s.d.badgePanel.view.tab)
	}
	shots.take(s, "badge-ranks")
	s.close()

	s.open("badges all", "")
	if s.d.badgePanel.view.tab != tabAll {
		t.Fatalf("badges all opened the tab %d", s.d.badgePanel.view.tab)
	}
	shots.take(s, "badge-list")
	s.close()

	// The menu with a game: the town of the first stage, as this account's
	// main game.
	if len(shots.town) == 0 {
		t.Fatal("the first stage left no town for the menu's picture")
	}
	must(t, os.WriteFile(filepath.Join(game.DataDir(), "saves", siteMenuGame+".json"), shots.town, 0o644))
	must(t, acct.SetMainGame(siteMenuGame))
	shots.takeMenu(s.eng, "menu")
}

// --- the file format ----------------------------------------------------------

// siteScreenFile is one screen as the wiki loads it: its size in cells, the
// colors it uses, its styles (foreground and background as palette indexes,
// then 1 bold + 2 italic + 4 underline), and its rows: the text and the style
// runs (style, length, style, length, ...). Style 0 is plain text on the
// screen's own background, and a row stops where only blank background is
// left. Bg is
// set for a screen drawn in another theme than the default, whose ground the
// wiki's plate does not already have.
type siteScreenFile struct {
	W       int             `json:"w"`
	H       int             `json:"h"`
	Bg      string          `json:"bg,omitempty"`
	Palette []string        `json:"palette"`
	Styles  [][3]int        `json:"styles"`
	Rows    []siteScreenRow `json:"rows"`
	Pics    []siteScreenPic `json:"pics,omitempty"`   // where each picture sits, in cells
	Images  []string        `json:"images,omitempty"` // distinct images, base64 PNG
}

type siteScreenRow struct {
	T string `json:"t"`
	R []int  `json:"r"`
}

// A siteScreenPic is a picture over the grid: its cells from (X, Y), W by H
// of them, showing Images[I].
type siteScreenPic struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
	I int `json:"i"`
}

func siteHex(c theme.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r&0xff, g&0xff, b&0xff)
}

// encodeSiteScreen turns drawn cells and their pictures into the file's bytes
// and a plain-text copy. It refuses a cell the wiki's grid could not hold: a
// wide rune or a combining one, and a picture that some other ink drew over.
func encodeSiteScreen(cells []simCell, pics []kit.Pic, w, h int, defFg, defBg theme.Color, ownGround bool) ([]byte, string, error) {
	f := siteScreenFile{W: w, H: h}
	if ownGround {
		f.Bg = siteHex(defBg)
	}
	kept, err := placePictures(cells, pics, w, h)
	if err != nil {
		return nil, "", err
	}
	imgIdx := map[string]int{} // encoded PNG -> its index in f.Images
	for _, p := range kept {
		enc, err := encodeSitePNG(reduceSiteImage(p.Img))
		if err != nil {
			return nil, "", err
		}
		i, ok := imgIdx[string(enc)]
		if !ok {
			i = len(f.Images)
			imgIdx[string(enc)] = i
			f.Images = append(f.Images, base64.StdEncoding.EncodeToString(enc))
		}
		r := p.Rect
		f.Pics = append(f.Pics, siteScreenPic{X: r.Min.X, Y: r.Min.Y, W: r.Dx(), H: r.Dy(), I: i})
	}
	palIdx := map[string]int{}
	color := func(c theme.Color) int {
		hx := siteHex(c)
		if i, ok := palIdx[hx]; ok {
			return i
		}
		palIdx[hx] = len(f.Palette)
		f.Palette = append(f.Palette, hx)
		return palIdx[hx]
	}
	styIdx := map[[3]int]int{}
	style := func(fg, bg theme.Color, attr int) int {
		k := [3]int{color(fg), color(bg), attr}
		if i, ok := styIdx[k]; ok {
			return i
		}
		styIdx[k] = len(f.Styles)
		f.Styles = append(f.Styles, k)
		return styIdx[k]
	}
	style(defFg, defBg, 0)

	var plain strings.Builder
	for y := 0; y < h; y++ {
		type cell struct {
			r      rune
			fg, bg theme.Color
			attr   int
		}
		row := make([]cell, w)
		end := 0 // one past the last cell that is not a plain blank
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			r := ' '
			if len(c.Runes) > 0 {
				r = c.Runes[0]
			}
			if len(c.Runes) > 1 || uniseg.StringWidth(string(r)) != 1 {
				return nil, "", fmt.Errorf("cell (%d,%d) holds %q, which is not one column wide", x, y, string(c.Runes))
			}
			fg, bg, attrs := decompose(c.Style)
			if fg == theme.ColorDefault {
				fg = defFg
			}
			if bg == theme.ColorDefault {
				bg = defBg
			}
			if attrs&kit.Reverse != 0 {
				fg, bg = bg, fg
			}
			attr := 0
			if attrs&kit.Bold != 0 {
				attr |= 1
			}
			if attrs&kit.Italic != 0 {
				attr |= 2
			}
			if attrs&kit.Underline != 0 {
				attr |= 4
			}
			row[x] = cell{r, fg, bg, attr}
			if r != ' ' || bg != defBg || attr&4 != 0 {
				end = x + 1
			}
		}
		var text []rune
		runs := []int{}
		cur, n := -1, 0
		var last cell
		for x := 0; x < end; x++ {
			c := row[x]
			if c.r == ' ' && c.attr&4 == 0 {
				// A space shows only its background: it rides in the run
				// before it when that has the same one, and is plain otherwise.
				if cur >= 0 && last.bg == c.bg && last.attr&4 == 0 {
					c.fg, c.attr = last.fg, last.attr
				} else {
					c.fg, c.attr = defFg, 0
				}
			}
			s := style(c.fg, c.bg, c.attr)
			text = append(text, c.r)
			if s != cur {
				if n > 0 {
					runs = append(runs, cur, n)
				}
				cur, n = s, 0
			}
			n++
			last = c
		}
		if n > 0 {
			runs = append(runs, cur, n)
		}
		plain.WriteString(string(text) + "\n")
		f.Rows = append(f.Rows, siteScreenRow{T: string(text), R: runs})
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(f); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), plain.String(), nil
}

// placePictures checks that each picture still owns every cell of its
// rectangle: the cell's Pic is the picture's index plus one. A picture no cell
// owns is left out. One that some cells own and others do not means something
// was drawn over part of it, and the wiki must not show it.
func placePictures(cells []simCell, pics []kit.Pic, w, h int) ([]kit.Pic, error) {
	var kept []kit.Pic
	for i, p := range pics {
		r := p.Rect
		owned, area := 0, 0
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				area++
				if x < 0 || y < 0 || x >= w || y >= h {
					continue
				}
				if int(cells[y*w+x].Pic) == i+1 {
					owned++
				}
			}
		}
		switch {
		case owned == 0:
			continue
		case owned < area:
			return nil, fmt.Errorf("picture over %v: %d of its %d cells were drawn over", r, area-owned, area)
		}
		kept = append(kept, p)
	}
	return kept, nil
}

// reduceSiteImage shrinks img by the largest whole factor k that divides both
// its sides and leaves every k×k block one colour. Coin art is drawn at a
// whole multiple of its pixels, so this gives the art at its own size. The
// result always starts at 0,0.
func reduceSiteImage(img *image.RGBA) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	k := max(siteGCD(w, h), 1)
	for ; k > 1 && (w%k != 0 || h%k != 0 || !siteBlocksUniform(img, k)); k-- {
	}
	out := image.NewRGBA(image.Rect(0, 0, w/k, h/k))
	for y := 0; y < h/k; y++ {
		for x := 0; x < w/k; x++ {
			out.SetRGBA(x, y, img.RGBAAt(b.Min.X+x*k, b.Min.Y+y*k))
		}
	}
	return out
}

// siteBlocksUniform reports whether every k×k block of img is one colour.
func siteBlocksUniform(img *image.RGBA, k int) bool {
	b := img.Bounds()
	for by := b.Min.Y; by < b.Max.Y; by += k {
		for bx := b.Min.X; bx < b.Max.X; bx += k {
			c := img.RGBAAt(bx, by)
			for y := by; y < by+k; y++ {
				for x := bx; x < bx+k; x++ {
					if img.RGBAAt(x, y) != c {
						return false
					}
				}
			}
		}
	}
	return true
}

// siteGCD is the greatest common divisor of two positive numbers.
func siteGCD(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// encodeSitePNG is img as PNG bytes.
func encodeSitePNG(img *image.RGBA) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
