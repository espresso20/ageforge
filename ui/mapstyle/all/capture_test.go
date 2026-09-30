//go:build mapcapture

package all

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// TestWriteCaptures draws review captures of both styles from the real
// smoke-bot states (run TestGenerateStates in ui/mapstyle/capture first):
// .txt (the monochrome check) and .html built cell by cell from a tcell
// SimulationScreen, animated where there are several frames.
//
//	go test -tags mapcapture -run TestWriteCaptures ./ui/mapstyle/all
//
// MAP_STATES_DIR and MAP_CAPTURES_DIR pick the folders (defaults under
// map_captures/ at the repo root, which git ignores).
func TestWriteCaptures(t *testing.T) {
	root := filepath.Join("..", "..", "..", "map_captures")
	states := envOr("MAP_STATES_DIR", filepath.Join(root, "states"))
	out := envOr("MAP_CAPTURES_DIR", root)
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	cp := &capturer{t: t, states: states, out: out, reg: Registry(), b: mapmodel.NewBuilder(nil)}
	cp.run()
	cp.index()
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

type capturer struct {
	t      *testing.T
	states string
	out    string
	reg    *mapstyle.Registry
	b      *mapmodel.Builder
	files  []string
}

// shot is one capture.
type shot struct {
	name    string
	style   string
	state   string // state file name without .json.gz
	since   bool   // diff against prev_<state>
	theme   string // "" = forge
	tier    mapmodel.GlyphTier
	w, h    int
	compact bool
	frames  int
	tod     float64 // < 0: the state's own tick
	opts    []mapstyle.Option
	keys    []*tcell.EventKey
	note    string
}

func key(k tcell.Key) *tcell.EventKey              { return tcell.NewEventKey(k, 0, tcell.ModNone) }
func keys(ks ...*tcell.EventKey) []*tcell.EventKey { return ks }

var reviewAges = []string{"primitive_age", "bronze_age", "medieval_age", "industrial_age", "victorian_age",
	"atomic_age", "digital_age", "cyberpunk_age", "space_age", "galactic_age"}

func short(age string) string { return strings.TrimSuffix(age, "_age") }

func (c *capturer) run() {
	U, A, N := mapmodel.TierUnicode, mapmodel.TierASCII, mapmodel.TierNerd
	tab := key(tcell.KeyTab)
	for _, age := range reviewAges {
		a := short(age)
		c.shoot(shot{name: a, style: "roguelike", state: age, since: true, w: 160, h: 48, tod: -1})
		c.shoot(shot{name: a + "_region", style: "roguelike", state: age, since: true, w: 160, h: 48, tod: -1, keys: keys(key(tcell.KeyPgUp)), note: "region zoom, epoch plate"})
		c.shoot(shot{name: a + "_mini", style: "roguelike", state: age, since: true, w: 40, h: 15, compact: true, tod: -1})
		c.shoot(shot{name: a, style: "skyline", state: age, since: true, w: 160, h: 48, tod: -1})
		c.shoot(shot{name: a + "_night", style: "skyline", state: age, since: true, w: 160, h: 48, tod: 0.95})
		c.shoot(shot{name: a + "_mini", style: "skyline", state: age, since: true, w: 40, h: 15, compact: true, tod: -1})
	}
	// animated traffic and life
	c.shoot(shot{name: "cyberpunk_night_anim", style: "skyline", state: "cyberpunk_age", since: true, w: 160, h: 48, tod: 0.95, frames: 16, note: "flying cars and maglev, 16 frames"})
	c.shoot(shot{name: "cyberpunk_day_anim", style: "skyline", state: "cyberpunk_age", since: true, w: 160, h: 48, tod: 0.5, frames: 16})
	c.shoot(shot{name: "industrial_anim", style: "skyline", state: "industrial_age", w: 160, h: 48, tod: 0.45, frames: 12})
	c.shoot(shot{name: "atomic_anim", style: "skyline", state: "atomic_age", w: 160, h: 48, tod: 0.8, frames: 12})
	c.shoot(shot{name: "medieval_anim", style: "skyline", state: "medieval_age", w: 160, h: 48, tod: 0.7, frames: 12})
	c.shoot(shot{name: "space_anim", style: "skyline", state: "space_age", w: 160, h: 48, tod: 0.9, frames: 12})
	c.shoot(shot{name: "galactic_anim", style: "skyline", state: "galactic_age", w: 160, h: 48, tod: 0.9, frames: 12})
	c.shoot(shot{name: "medieval_anim", style: "roguelike", state: "medieval_age", w: 160, h: 48, tod: -1, frames: 10})
	c.shoot(shot{name: "cyberpunk_region_anim", style: "roguelike", state: "cyberpunk_age", w: 160, h: 48, tod: -1, frames: 10, keys: keys(key(tcell.KeyPgUp))})
	c.shoot(shot{name: "cyberpunk_mini_anim", style: "skyline", state: "cyberpunk_age", w: 40, h: 15, compact: true, tod: 0.95, frames: 12})
	// small terminals
	for _, st := range []string{"primitive_early", "medieval_age", "industrial_age", "cyberpunk_age"} {
		c.shoot(shot{name: short(st) + "_80x24", style: "roguelike", state: st, since: true, w: 80, h: 24, tod: -1})
		c.shoot(shot{name: short(st) + "_80x24", style: "skyline", state: st, since: true, w: 80, h: 24, tod: -1})
	}
	c.shoot(shot{name: "industrial_100x30", style: "roguelike", state: "industrial_age", w: 100, h: 30, tod: -1})
	c.shoot(shot{name: "industrial_100x30", style: "skyline", state: "industrial_age", w: 100, h: 30, tod: -1})
	c.shoot(shot{name: "primitive_early", style: "roguelike", state: "primitive_early", w: 160, h: 48, tod: -1})
	c.shoot(shot{name: "primitive_early", style: "skyline", state: "primitive_early", w: 160, h: 48, tod: -1})
	// inspect, flows, district
	c.shoot(shot{name: "medieval_inspect", style: "roguelike", state: "medieval_age", since: true, w: 160, h: 48, tod: -1, keys: keys(tab, tab, tab)})
	c.shoot(shot{name: "medieval_district", style: "roguelike", state: "medieval_age", since: true, w: 160, h: 48, tod: -1, keys: keys(tab, key(tcell.KeyPgDn))})
	c.shoot(shot{name: "industrial_inspect", style: "skyline", state: "industrial_age", since: true, w: 160, h: 48, tod: 0.9, keys: keys(tab, tab, tab, tab)})
	c.shoot(shot{name: "cyberpunk_inspect_civ", style: "skyline", state: "cyberpunk_age", w: 160, h: 48, tod: 0.95, keys: keys(key(tcell.KeyBacktab), key(tcell.KeyBacktab))})
	for _, st := range []string{"industrial_age", "information_age"} {
		c.shoot(shot{name: short(st) + "_flows", style: "roguelike", state: st, w: 160, h: 48, tod: -1, opts: []mapstyle.Option{mapstyle.OptFlows}, note: "flows toggle"})
		c.shoot(shot{name: short(st) + "_flows", style: "skyline", state: st, w: 160, h: 48, tod: 0.9, opts: []mapstyle.Option{mapstyle.OptFlows}, note: "flows toggle"})
	}
	// events
	c.shoot(shot{name: "catastrophe_information", style: "roguelike", state: "catastrophe_information_age", w: 160, h: 48, tod: -1})
	c.shoot(shot{name: "catastrophe_information", style: "skyline", state: "catastrophe_information_age", w: 160, h: 48, tod: 0.9})
	c.shoot(shot{name: "harbinger_victorian", style: "skyline", state: "harbinger_victorian_age", w: 160, h: 48, tod: 0.85})
	c.shoot(shot{name: "harbinger_iron", style: "roguelike", state: "harbinger_iron_age", w: 160, h: 48, tod: -1})
	// light themes
	c.shoot(shot{name: "medieval_light", style: "roguelike", state: "medieval_age", since: true, theme: "daylight", w: 160, h: 48, tod: -1})
	c.shoot(shot{name: "industrial_region_parchment", style: "roguelike", state: "industrial_age", theme: "parchment", w: 160, h: 48, tod: -1, keys: keys(key(tcell.KeyPgUp))})
	c.shoot(shot{name: "industrial_light", style: "skyline", state: "industrial_age", since: true, theme: "daylight", w: 160, h: 48, tod: 0.45})
	c.shoot(shot{name: "cyberpunk_night_light", style: "skyline", state: "cyberpunk_age", theme: "daylight", w: 160, h: 48, tod: 0.95})
	c.shoot(shot{name: "medieval_mini_light", style: "roguelike", state: "medieval_age", theme: "daylight", w: 40, h: 15, compact: true, tod: -1})
	c.shoot(shot{name: "cyberpunk_mini_light", style: "skyline", state: "cyberpunk_age", theme: "daylight", w: 40, h: 15, compact: true, tod: 0.95})
	c.shoot(shot{name: "cyberpunk_theme", style: "skyline", state: "cyberpunk_age", theme: "cyberpunk", w: 160, h: 48, tod: 0.95})
	c.shoot(shot{name: "digital_mono", style: "skyline", state: "digital_age", theme: "monochrome", w: 160, h: 48, tod: 0.9})
	// glyph tiers
	for _, tr := range []mapmodel.GlyphTier{A, U, N} {
		c.shoot(shot{name: "cyberpunk_glyphs_" + tr.String(), style: "skyline", state: "cyberpunk_age", w: 160, h: 48, tod: 0.95, tier: tr, frames: 6})
		c.shoot(shot{name: "medieval_glyphs_" + tr.String(), style: "roguelike", state: "medieval_age", w: 160, h: 48, tod: -1, tier: tr})
		c.shoot(shot{name: "medieval_region_glyphs_" + tr.String(), style: "roguelike", state: "medieval_age", w: 160, h: 48, tod: -1, tier: tr, keys: keys(key(tcell.KeyPgUp))})
	}
}

func (c *capturer) load(name string) (game.GameState, bool) {
	st, err := capture.LoadState(filepath.Join(c.states, name+".json.gz"))
	if err != nil {
		c.t.Logf("skip %s: %v", name, err)
		return st, false
	}
	return st, true
}

func (c *capturer) shoot(s shot) {
	st, ok := c.load(s.state)
	if !ok {
		return
	}
	if s.tod >= 0 {
		clk := mapmodel.ClockAt(st.Tick)
		st.Tick = clockTick(clk.Day, s.tod)
	}
	var since *mapmodel.Visit
	if s.since {
		if prev, ok := c.load("prev_" + strings.TrimPrefix(strings.TrimPrefix(s.state, "harbinger_"), "catastrophe_")); ok {
			since = mapmodel.VisitOf(c.b.Build(&prev, nil))
		}
	}
	th := s.theme
	if th == "" {
		th = "forge"
	}
	if err := theme.SetActive(th); err != nil {
		c.t.Fatal(err)
	}
	m := mapmodel.NewBuilder(c.b.Catalog()).Build(&st, since)
	view, ok := c.reg.New(s.style)
	if !ok {
		c.t.Fatalf("no style %s", s.style)
	}
	for _, o := range s.opts {
		view.SetOption(o, true)
	}
	scr := capture.NewScreen(s.w, s.h)
	defer scr.Fini()
	r := mapstyle.Rect{W: s.w, H: s.h}
	f := mapstyle.Frame{Model: m, Tier: s.tier}
	draw := func() {
		if s.compact {
			view.DrawCompact(scr, r, f)
		} else {
			view.Draw(scr, r, f)
		}
	}
	draw()
	for _, k := range s.keys {
		view.HandleKey(k, f)
		draw()
	}
	n := s.frames
	if n < 1 {
		n = 1
	}
	var html []string
	var txt strings.Builder
	for i := 0; i < n; i++ {
		f.Anim = i * 3
		draw()
		html = append(html, capture.Frame(scr))
		if i == 0 {
			_ = os.MkdirAll(filepath.Join(c.out, "png"), 0o755)
			_ = capture.PNG(filepath.Join(c.out, "png", s.style+"_"+s.name+".png"), scr)
		}
		if i == 0 || n > 1 {
			if n > 1 {
				fmt.Fprintf(&txt, "--- frame %d\n", i)
			}
			txt.WriteString(capture.Text(scr))
		}
	}
	title := fmt.Sprintf("%s · %s · %s · %dx%d · theme %s · glyphs %s", s.style, s.state, s.name, s.w, s.h, th, s.tier)
	if s.note != "" {
		title += " · " + s.note
	}
	fonts := capture.Fonts
	if s.tier == mapmodel.TierNerd {
		fonts = capture.NerdFonts
	}
	bg := capture.Hex(theme.Color(theme.RoleBackground))
	fg := capture.Hex(theme.Color(theme.RoleText))
	base := s.style + "_" + s.name
	c.write(base+".html", capture.HTML(title, bg, fg, fonts, html, 6))
	c.write(base+".txt", title+"\n"+txt.String())
}

func clockTick(day int, tod float64) int {
	// ClockAtTOD gives the clock; reproduce the tick it came from.
	for t := (day - 1) * mapmodel.DayTicks; t < (day+1)*mapmodel.DayTicks; t += 10 {
		c := mapmodel.ClockAt(t)
		if c.Day == day && c.TOD >= tod {
			return t
		}
	}
	return (day - 1) * mapmodel.DayTicks
}

func (c *capturer) write(name, body string) {
	if err := os.WriteFile(filepath.Join(c.out, name), []byte(body), 0o644); err != nil {
		c.t.Fatal(err)
	}
	if strings.HasSuffix(name, ".html") {
		c.files = append(c.files, name)
	}
}

func (c *capturer) index() {
	sort.Strings(c.files)
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8"><title>Map captures</title><style>body{font-family:ui-monospace,Menlo,monospace;background:#111;color:#ccc;margin:20px}a{color:#8cf;display:block;margin:2px 0}</style></head><body><h1>Map captures</h1><p>Real tcell SimulationScreen renders from smoke-bot states. Nerd-tier pages need a Nerd Font installed (JetBrainsMono Nerd Font Mono or Symbols Nerd Font Mono).</p>`)
	for _, f := range c.files {
		fmt.Fprintf(&b, `<a href="%s">%s</a>`, f, strings.TrimSuffix(f, ".html"))
	}
	b.WriteString("</body></html>\n")
	c.write("index.html", b.String())
}
