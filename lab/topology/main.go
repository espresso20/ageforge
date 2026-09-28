// Command topology is the Map Lab "topology" prototype: the empire drawn as
// a living system diagram (producers → resource bus → services → external
// hosts) instead of a geographic map. See DESIGN.md.
//
//	go run ./lab/topology -gen                  # rebuild fixtures with the smoke bot
//	go run ./lab/topology -age medieval_age     # interactive viewer
//	go run ./lab/topology -capture lab/topology/captures
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// fixtureAges are the runs -gen plays: <age>[@dwell ticks][+incident].
// "+incident" layers real player actions on the bot's run (raiding a civ
// into war, then the dev console's catastrophe) and writes <age>_incident.
var fixtureAges = []string{
	"primitive_age@120", "bronze_age", "medieval_age", "industrial_age",
	"industrial_age+incident", "electric_age", "digital_age", "cyberpunk_age",
	"galactic_age", "quantum_age",
}

func main() {
	var (
		gen      = flag.Bool("gen", false, "generate fixtures with the smoke bot, then exit")
		genAges  = flag.String("ages", "", "comma-separated runs for -gen (default: the lab set)")
		seed     = flag.Int64("seed", 7, "bot seed for -gen")
		dwell    = flag.Int("dwell", 900, "ticks to play inside the target age for -gen")
		dir      = flag.String("fixtures", "lab/topology/fixtures", "fixture directory")
		age      = flag.String("age", "medieval_age", "fixture to view")
		themeKey = flag.String("theme", "forge", "theme key")
		capture  = flag.String("capture", "", "write txt/html captures into this directory and exit")
	)
	flag.Parse()

	if *gen {
		runGen(*genAges, *seed, *dwell, *dir)
		return
	}
	if *capture != "" {
		if err := runCaptures(*dir, *capture); err != nil {
			fmt.Fprintln(os.Stderr, "capture:", err)
			os.Exit(1)
		}
		return
	}
	if err := runInteractive(*dir, *age, *themeKey); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runGen(list string, seed int64, dwell int, dir string) {
	ages := fixtureAges
	if list != "" {
		ages = strings.Split(list, ",")
	}
	for _, spec := range ages {
		a, d := spec, dwell
		incident := strings.HasSuffix(a, "+incident")
		a = strings.TrimSuffix(a, "+incident")
		if i := strings.IndexByte(a, '@'); i > 0 {
			fmt.Sscanf(a[i+1:], "%d", &d)
			a = a[:i]
		}
		target := a
		if incident {
			a += "_incident"
		}
		fmt.Printf("%s (seed %d, dwell %d)\n", a, seed, d)
		fx, err := generate(target, seed, d, false, incident, func(f string, v ...any) { fmt.Printf(f+"\n", v...) })
		if err != nil {
			fmt.Fprintln(os.Stderr, "gen:", err)
			os.Exit(1)
		}
		path := filepath.Join(dir, a+".json.gz")
		if err := saveFixture(fx, path); err != nil {
			fmt.Fprintln(os.Stderr, "save:", err)
			os.Exit(1)
		}
		fmt.Printf("  wrote %s (%d ticks)\n", path, fx.Ticks)
	}
}

func listFixtures(dir string) []string {
	ms, _ := filepath.Glob(filepath.Join(dir, "*.json.gz"))
	var out []string
	for _, m := range ms {
		out = append(out, strings.TrimSuffix(filepath.Base(m), ".json.gz"))
	}
	order := map[string]int{}
	for i, a := range fixtureAges {
		a = strings.SplitN(strings.SplitN(a, "@", 2)[0], "+", 2)[0]
		if strings.HasSuffix(fixtureAges[i], "+incident") {
			a += "_incident"
		}
		order[a] = i
	}
	sort.SliceStable(out, func(i, j int) bool { return order[out[i]] < order[out[j]] })
	return out
}

func mustTheme(key string) theme.Theme {
	t, ok := theme.ByKey(key)
	if !ok {
		t, _ = theme.ByKey(theme.DefaultKey)
	}
	return t
}

// runCaptures writes the comparison set.
func runCaptures(dir, out string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	forge := mustTheme("forge")
	for _, name := range listFixtures(dir) {
		fx, err := loadFixture(filepath.Join(dir, name+".json.gz"))
		if err != nil {
			return err
		}
		title := fmt.Sprintf("%s · %s · seed %d · 160×48 · forge", name, fx.State.AgeName, fx.Seed)
		v := newView(fx, 160, 48)
		v.frame = 3
		if err := writeCapture(out, name, title, v, forge, false); err != nil {
			return err
		}
		if err := writeAnim(out, name, title, v, forge, 16); err != nil {
			return err
		}
		fmt.Println("captured", name)
	}
	extras := []struct {
		fixture, base, theme string
		w, h                 int
		sel                  string
		mono                 bool
		world                bool
	}{
		{"medieval_age", "medieval_age-100x30", "forge", 100, 30, "", false, false},
		{"digital_age", "digital_age-100x30", "forge", 100, 30, "", false, false},
		{"medieval_age", "medieval_age-80x24", "forge", 80, 24, "", false, false},
		{"medieval_age", "medieval_age-mini-40x15", "forge", 40, 15, "", false, false},
		{"industrial_age_incident", "industrial_incident-mini-40x15", "forge", 40, 15, "", false, false},
		{"cyberpunk_age", "cyberpunk_age-mini-40x15", "cyberpunk", 40, 15, "", false, false},
		{"medieval_age", "medieval_age-light", "daylight", 160, 48, "", false, false},
		{"digital_age", "digital_age-light", "daylight", 160, 48, "", false, false},
		{"industrial_age_incident", "industrial_incident-light-100x30", "daylight", 100, 30, "", false, false},
		{"medieval_age", "medieval_age-inspect", "forge", 160, 48, "prod:food", false, false},
		{"industrial_age_incident", "industrial_incident-inspect", "forge", 160, 48, "host:civ:*war", false, false},
		{"digital_age", "digital_age-mono", "forge", 160, 48, "", true, false},
		{"cyberpunk_age", "cyberpunk_age-cyberpunk-theme", "cyberpunk", 160, 48, "", false, false},
		{"primitive_age", "primitive_age-80x24", "forge", 80, 24, "", false, false},
		{"industrial_age_incident", "industrial_incident-world", "forge", 160, 48, "", false, true},
		{"cyberpunk_age", "cyberpunk_age-world", "forge", 160, 48, "", false, true},
		{"galactic_age", "galactic_age-world", "cosmic", 160, 48, "", false, true},
		{"medieval_age", "medieval_age-world-100x30", "forge", 100, 30, "", false, true},
	}
	for _, e := range extras {
		fx, err := loadFixture(filepath.Join(dir, e.fixture+".json.gz"))
		if err != nil {
			fmt.Println("skip", e.base, err)
			continue
		}
		t := mustTheme(e.theme)
		v := newView(fx, e.w, e.h)
		v.frame = 3
		v.world = e.world
		if e.sel != "" {
			v.sel = e.sel
			if e.sel == "host:civ:*war" {
				v.sel = ""
				for id, n := range v.m.nodes {
					if strings.HasPrefix(id, "host:civ:") && n.alert == "AT WAR" {
						v.sel = id
					}
				}
			}
			v.inspect = v.sel != ""
		}
		title := fmt.Sprintf("%s · %s · %d×%d · %s", e.fixture, fx.State.AgeName, e.w, e.h, e.theme)
		if e.mono {
			title += " · monochrome"
		}
		if err := writeCapture(out, e.base, title, v, t, e.mono); err != nil {
			return err
		}
		if strings.Contains(e.base, "mini") || strings.HasSuffix(e.base, "100x30") {
			if err := writeAnim(out, e.base, title, v, t, 12); err != nil {
				return err
			}
		}
		fmt.Println("captured", e.base)
	}
	return writeIndex(out)
}

// writeIndex lists every capture on one page.
func writeIndex(out string) error {
	ms, _ := filepath.Glob(filepath.Join(out, "*.html"))
	var b strings.Builder
	b.WriteString(`<!doctype html><meta charset="utf-8"><title>topology captures</title>
<style>body{font:14px -apple-system,Segoe UI,sans-serif;background:#0d1117;color:#c9d1d9;padding:24px}a{color:#e3b341}li{margin:3px 0}</style>
<h1>Map Lab: topology</h1><p>Real terminal renders (tcell SimulationScreen cells → HTML) of fixtures the smoke bot played. <code>-anim</code> pages loop the animation.</p><ul>`)
	for _, m := range ms {
		name := filepath.Base(m)
		if name == "index.html" {
			continue
		}
		fmt.Fprintf(&b, `<li><a href="%s">%s</a></li>`, name, strings.TrimSuffix(name, ".html"))
	}
	b.WriteString("</ul>")
	return os.WriteFile(filepath.Join(out, "index.html"), []byte(b.String()), 0o644)
}

// ---- interactive -----------------------------------------------------

func runInteractive(dir, age, themeKey string) error {
	names := listFixtures(dir)
	if len(names) == 0 {
		return fmt.Errorf("no fixtures in %s (run with -gen first)", dir)
	}
	ai := 0
	for i, n := range names {
		if n == age {
			ai = i
		}
	}
	themes := theme.All()
	ti := 0
	for i, t := range themes {
		if t.Key == themeKey {
			ti = i
		}
	}
	scr, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := scr.Init(); err != nil {
		return err
	}
	defer scr.Fini()
	return loop(scr, dir, names, ai, themes, ti, nil)
}

// loop runs the viewer on scr until quit. onFrame, if set, sees each drawn
// view (tests use it).
func loop(scr tcell.Screen, dir string, names []string, ai int, themes []theme.Theme, ti int, onFrame func(*view)) error {
	load := func() (*view, error) {
		fx, err := loadFixture(filepath.Join(dir, names[ai]+".json.gz"))
		if err != nil {
			return nil, err
		}
		w, h := scr.Size()
		return newView(fx, w, h), nil
	}
	v, err := load()
	if err != nil {
		return err
	}
	pal := newPalette(themes[ti], v.s, false)
	events := make(chan tcell.Event, 16)
	go func() {
		for {
			events <- scr.PollEvent()
		}
	}()
	tick := time.NewTicker(120 * time.Millisecond)
	defer tick.Stop()
	for {
		v.draw()
		v.c.blit(scr, pal)
		scr.Show()
		if onFrame != nil {
			onFrame(v)
		}
		select {
		case <-tick.C:
			v.frame++
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				scr.Sync()
				w, h := scr.Size()
				v.resize(w, h)
			case *tcell.EventKey:
				switch {
				case ev.Key() == tcell.KeyEscape && v.inspect:
					v.inspect = false
				case ev.Key() == tcell.KeyEscape || ev.Rune() == 'q' || ev.Key() == tcell.KeyCtrlC:
					return nil
				case ev.Key() == tcell.KeyUp:
					v.move(0, -1)
				case ev.Key() == tcell.KeyDown:
					v.move(0, 1)
				case ev.Key() == tcell.KeyLeft:
					v.move(-1, 0)
				case ev.Key() == tcell.KeyRight:
					v.move(1, 0)
				case ev.Key() == tcell.KeyEnter || ev.Rune() == ' ':
					if v.sel == "" {
						v.move(0, 0)
					}
					v.inspect = !v.inspect
					v.staged = ""
				case ev.Rune() >= '1' && ev.Rune() <= '3' && v.inspect:
					if n := v.m.nodes[v.sel]; n != nil {
						_, cmds := v.inspectLines(n)
						if i := int(ev.Rune() - '1'); i < len(cmds) {
							v.staged = cmds[i]
						}
					}
				case ev.Rune() == 'w':
					v.world = !v.world
					v.inspect = false
				case ev.Rune() == 'd':
					v.diff = !v.diff
				case ev.Rune() == 't':
					ti = (ti + 1) % len(themes)
					pal = newPalette(themes[ti], v.s, false)
				case ev.Rune() == 'a' || ev.Rune() == 'A':
					if ev.Rune() == 'a' {
						ai = (ai + 1) % len(names)
					} else {
						ai = (ai + len(names) - 1) % len(names)
					}
					nv, err := load()
					if err == nil {
						nv.frame = v.frame
						v = nv
						pal = newPalette(themes[ti], v.s, false)
					}
				}
			}
		}
	}
}

// move walks the selection: dy within a column, dx across columns.
func (v *view) move(dx, dy int) {
	var cols [][]*node
	for c := 0; c < 4; c++ {
		var col []*node
		if c == int(kPort) {
			col = append(col, v.L.ports...)
		} else {
			for _, n := range v.L.nodes {
				if int(n.kind) == c && n.kind != kPort {
					col = append(col, n)
				}
			}
			if c == int(kService) && v.L.world != nil {
				col = append(col, v.L.world)
			}
		}
		sort.SliceStable(col, func(i, j int) bool { return col[i].y < col[j].y })
		if len(col) > 0 {
			cols = append(cols, col)
		}
	}
	if len(cols) == 0 {
		return
	}
	ci, ni := -1, -1
	for i, col := range cols {
		for j, n := range col {
			if n.id == v.sel {
				ci, ni = i, j
			}
		}
	}
	if ci < 0 {
		v.sel = cols[0][0].id
		return
	}
	y := cols[ci][ni].y
	switch {
	case dy != 0:
		ni = clampi(ni+dy, 0, len(cols[ci])-1)
	case dx != 0:
		ci = clampi(ci+dx, 0, len(cols)-1)
		best, bd := 0, 1<<30
		for j, n := range cols[ci] {
			d := n.y - y
			if d < 0 {
				d = -d
			}
			if d < bd {
				best, bd = j, d
			}
		}
		ni = best
	}
	v.sel = cols[ci][ni].id
	v.staged = ""
}
