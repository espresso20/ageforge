package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// shot renders one frame onto a tcell SimulationScreen and reads it back,
// so captures are what a terminal would really show.
func shot(w *World, v View, compact bool) (txt, htm string) {
	sim := tcell.NewSimulationScreen("UTF-8")
	_ = sim.Init()
	sim.SetSize(v.W, v.H)
	var fb *FB
	if compact {
		fb = renderCompact(w, v)
	} else {
		fb = render(w, v)
	}
	fb.blitTo(sim)
	sim.Show()
	return screenText(sim), screenHTML(sim)
}

type capSpec struct {
	name    string
	state   string
	w, h    int
	tod     float64
	theme   string
	frames  int
	cam     string // "end", "start", "mid", or "" (end)
	colors  int
	cursor  int
	newMk   bool
	compact bool
}

func loadWorld(dir, name string, sceneH int) (*World, error) {
	st, err := loadState(dir, name)
	if err != nil {
		return nil, err
	}
	var prev *game.GameState
	if p, err := loadState(dir, "prev_"+name); err == nil {
		prev = &p
	}
	return buildWorld(st, prev, sceneH), nil
}

func camFor(w *World, width int, mode string) int {
	switch mode {
	case "start":
		return 0
	case "mid":
		return max(0, (w.W-width)/2)
	}
	// "end": the present. If the whole town fits, centre it; otherwise end
	// the view just past the newest district, the frontier's cranes at the
	// right edge.
	if w.FrontierX-westPad+16 <= width {
		return (westPad+w.FrontierX)/2 - width/2
	}
	return w.FrontierX + 14 - width
}

func runCapture(dir, out string, c capSpec) error {
	sh := sceneRows(c.h)
	if c.compact {
		sh = c.h
	}
	w, err := loadWorld(dir, c.state, sh)
	if err != nil {
		return err
	}
	th, ok := theme.ByKey(c.theme)
	if !ok {
		return fmt.Errorf("no theme %q", c.theme)
	}
	cursor := c.cursor
	if cursor == 0 {
		cursor = -1
	}
	if c.w == 0 { // the whole panorama in one frame
		c.w = w.W
		c.cam = "start"
	}
	v := View{W: c.w, H: c.h, Cam: camFor(w, c.w, c.cam), TOD: c.tod, Theme: th, Colors: c.colors,
		Cursor: cursor, Weather: -1, ShowNew: c.newMk}
	n := max(1, c.frames)
	var txts, htms []string
	for i := 0; i < n; i++ {
		v.Frame = 40 + i
		t, h := shot(w, v, c.compact)
		txts = append(txts, t)
		htms = append(htms, h)
	}
	bg := fmt.Sprintf("#%06x", th.Color(theme.RoleBackground).Hex())
	fg := fmt.Sprintf("#%06x", th.Color(theme.RoleText).Hex())
	title := fmt.Sprintf("skyline · %s · %s · %dx%d", c.state, c.theme, c.w, c.h)
	if err := os.WriteFile(filepath.Join(out, c.name+".txt"), []byte(txts[0]), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, c.name+".html"), []byte(wrapHTML(title, bg, fg, htms, 6)), 0o644)
}

// captureSpecs is the review set: every age by day, key ages by night,
// the light and duotone themes, 16 colours, small terminals, the compact
// mini view, animation, inspect, changes, and the event overlays.
func captureSpecs() []capSpec {
	var cs []capSpec
	nightAt := map[string]bool{"medieval_age": true, "industrial_age": true, "atomic_age": true,
		"digital_age": true, "cyberpunk_age": true, "space_age": true, "quantum_age": true, "colonial_age": true}
	ages := []string{"primitive_age", "stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age",
		"renaissance_age", "colonial_age", "industrial_age", "victorian_age", "electric_age", "atomic_age",
		"modern_age", "information_age", "digital_age", "cyberpunk_age", "fusion_age", "space_age",
		"interstellar_age", "galactic_age", "quantum_age", "transcendent_age"}
	for i, a := range ages {
		// by day, with a little drift through the hours so the set is not
		// twenty-two identical noons
		tod := 0.40 + 0.03*float64(i%5)
		cs = append(cs, capSpec{name: a, state: a, w: 160, h: 45, tod: tod, theme: "forge", newMk: true})
		if nightAt[a] {
			cs = append(cs, capSpec{name: a + "_night", state: a, w: 160, h: 45, tod: 0.955, theme: "forge"})
		}
	}
	cs = append(cs,
		capSpec{name: "primitive_early", state: "primitive_early", w: 160, h: 45, tod: 0.745, theme: "forge"},
		capSpec{name: "primitive_early_night", state: "primitive_early", w: 160, h: 45, tod: 0.9, theme: "forge"},
		capSpec{name: "medieval_age_dusk", state: "medieval_age", w: 160, h: 45, tod: 0.752, theme: "forge"},
		// light themes: a daylight sky, and night as a blue hour
		capSpec{name: "light_daylight_industrial", state: "industrial_age", w: 160, h: 45, tod: 0.45, theme: "daylight"},
		capSpec{name: "light_daylight_cyberpunk_night", state: "cyberpunk_age", w: 160, h: 45, tod: 0.95, theme: "daylight"},
		capSpec{name: "light_hcl_medieval", state: "medieval_age", w: 160, h: 45, tod: 0.5, theme: "high_contrast_light"},
		capSpec{name: "theme_parchment_renaissance", state: "renaissance_age", w: 160, h: 45, tod: 0.45, theme: "parchment"},
		capSpec{name: "theme_monochrome_digital", state: "digital_age", w: 160, h: 45, tod: 0.95, theme: "monochrome"},
		capSpec{name: "theme_cyberpunk_space", state: "space_age", w: 160, h: 45, tod: 0.8, theme: "cyberpunk"},
		// the ANSI fallback
		capSpec{name: "ansi16_industrial", state: "industrial_age", w: 160, h: 45, tod: 0.45, theme: "forge", colors: 16},
		capSpec{name: "ansi16_cyberpunk_night", state: "cyberpunk_age", w: 160, h: 45, tod: 0.95, theme: "forge", colors: 16},
		// small terminals
		capSpec{name: "small_100x30_medieval", state: "medieval_age", w: 100, h: 30, tod: 0.45, theme: "forge", newMk: true},
		capSpec{name: "small_100x30_cyberpunk_night", state: "cyberpunk_age", w: 100, h: 30, tod: 0.95, theme: "forge"},
		capSpec{name: "small_80x24_industrial", state: "industrial_age", w: 80, h: 24, tod: 0.45, theme: "forge"},
		capSpec{name: "small_80x24_primitive_early", state: "primitive_early", w: 80, h: 24, tod: 0.74, theme: "forge"},
		// animation: smoke, windmills, traffic, blinking beacons
		capSpec{name: "anim_industrial", state: "industrial_age", w: 160, h: 45, tod: 0.45, theme: "forge", frames: 10},
		capSpec{name: "anim_cyberpunk_night", state: "cyberpunk_age", w: 160, h: 45, tod: 0.95, theme: "forge", frames: 10},
		capSpec{name: "anim_medieval_dusk", state: "medieval_age", w: 160, h: 45, tod: 0.75, theme: "forge", frames: 10},
		// the verb: inspect a building
		capSpec{name: "inspect_industrial", state: "industrial_age", w: 160, h: 45, tod: 0.45, theme: "forge", cursor: 118},
		capSpec{name: "inspect_medieval_night", state: "medieval_age", w: 160, h: 45, tod: 0.95, theme: "forge", cursor: 96},
		// events
		capSpec{name: "event_harbinger_iron", state: "harbinger_iron_age", w: 160, h: 45, tod: 0.47, theme: "forge"},
		capSpec{name: "event_harbinger_victorian_night", state: "harbinger_victorian_age", w: 160, h: 45, tod: 0.93, theme: "forge"},
		capSpec{name: "event_catastrophe_information", state: "catastrophe_information_age", w: 160, h: 45, tod: 0.8, theme: "forge", frames: 6},
		capSpec{name: "event_harbinger_cyberpunk_night", state: "harbinger_cyberpunk_age", w: 160, h: 45, tod: 0.95, theme: "forge", frames: 4},
	)
	// the whole city, west (oldest) to east (the present), one frame each
	cs = append(cs,
		capSpec{name: "panorama_medieval_age", state: "medieval_age", w: 0, h: 45, tod: 0.74, theme: "forge"},
		capSpec{name: "panorama_digital_age_night", state: "digital_age", w: 0, h: 45, tod: 0.95, theme: "forge"},
	)
	for _, a := range []string{"primitive_early", "medieval_age", "industrial_age", "cyberpunk_age", "galactic_age"} {
		cs = append(cs, capSpec{name: "compact_40x15_" + a, state: a, w: 40, h: 15, tod: 0.45, theme: "forge", compact: true, newMk: true})
	}
	cs = append(cs,
		capSpec{name: "compact_40x15_cyberpunk_night", state: "cyberpunk_age", w: 40, h: 15, tod: 0.95, theme: "forge", compact: true, frames: 6},
		capSpec{name: "compact_40x15_daylight_industrial", state: "industrial_age", w: 40, h: 15, tod: 0.45, theme: "daylight", compact: true},
	)
	return cs
}

func writeCaptures(dir, out string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	var done []capSpec
	for _, c := range captureSpecs() {
		if _, err := os.Stat(filepath.Join(dir, c.state+".json.gz")); err != nil {
			fmt.Println("skip", c.name, "(no state", c.state+")")
			continue
		}
		if err := runCapture(dir, out, c); err != nil {
			return fmt.Errorf("%s: %w", c.name, err)
		}
		done = append(done, c)
	}
	fmt.Println("wrote", len(done), "captures")
	return writeIndex(out, done)
}

func writeIndex(out string, cs []capSpec) error {
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8"><title>skyline captures</title>
<style>body{background:#0e0e12;color:#ccc;font:14px ui-monospace,Menlo,monospace;margin:20px}a{color:#e0b84a}li{margin:3px 0}</style>
</head><body><h1>skyline · capture index</h1><p>Every capture is a tcell SimulationScreen read back cell by cell (HTML) and as plain text (.txt).</p><ul>`)
	for _, c := range cs {
		fmt.Fprintf(&b, `<li><a href="%s.html">%s</a> · <a href="%s.txt">txt</a> · %dx%d · %s</li>`, c.name, c.name, c.name, c.w, c.h, c.theme)
	}
	b.WriteString("</ul></body></html>\n")
	return os.WriteFile(filepath.Join(out, "index.html"), []byte(b.String()), 0o644)
}
