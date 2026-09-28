package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
)

// captureAll writes the comparison captures: every snapshot age at 160×48
// on the default dark theme, plus light-theme, small-terminal, mini-view,
// animated, monochrome and 16-colour variants.
func captureAll(lib, alt *library, dir string) error {
	type job struct {
		name, age, theme, plate string
		w, h, frames            int
		mono, pal16             bool
		cursorTo                string // "civ", "route", "omen": tab to the first such POI
		alt                     bool   // from the second player's states
	}
	var jobs []job
	for _, a := range lib.ages {
		jobs = append(jobs, job{name: a, age: a, theme: "forge", w: 160, h: 48, frames: 1})
	}
	jobs = append(jobs,
		job{name: "light_medieval_age", age: "medieval_age", theme: "daylight", w: 160, h: 48, frames: 1},
		job{name: "light_industrial_age", age: "industrial_age", theme: "daylight", w: 160, h: 48, frames: 1},
		job{name: "light_primitive_age", age: "primitive_age", theme: "parchment", w: 160, h: 48, frames: 1},
		job{name: "small_100x30_medieval_age", age: "medieval_age", theme: "forge", w: 100, h: 30, frames: 1},
		job{name: "small_100x30_digital_age", age: "digital_age", theme: "forge", w: 100, h: 30, frames: 1},
		job{name: "small_80x24_industrial_age", age: "industrial_age", theme: "forge", w: 80, h: 24, frames: 1},
		job{name: "mini_40x15_primitive_age", age: "primitive_age", theme: "forge", w: 40, h: 15, frames: 1},
		job{name: "mini_40x15_medieval_age", age: "medieval_age", theme: "forge", w: 40, h: 15, frames: 1},
		job{name: "mini_40x15_atomic_age", age: "atomic_age", theme: "forge", w: 40, h: 15, frames: 1},
		job{name: "mini_40x15_galactic_age", age: "galactic_age", theme: "forge", w: 40, h: 15, frames: 1},
		job{name: "anim_medieval_age", age: "medieval_age", theme: "forge", w: 160, h: 48, frames: 4},
		job{name: "anim_digital_age", age: "digital_age", theme: "forge", w: 160, h: 48, frames: 4},
		job{name: "anim_cyberpunk_age", age: "cyberpunk_age", theme: "forge", w: 160, h: 48, frames: 4},
		job{name: "anim_space_age", age: "space_age", theme: "forge", w: 160, h: 48, frames: 4},
		job{name: "mono_industrial_age", age: "industrial_age", theme: "forge", w: 160, h: 48, frames: 1, mono: true},
		job{name: "ansi16_medieval_age", age: "medieval_age", theme: "forge", w: 160, h: 48, frames: 1, pal16: true},
		job{name: "inspect_civ_industrial_age", age: "industrial_age", theme: "forge", w: 160, h: 48, frames: 1, cursorTo: "civ"},
		job{name: "plates_industrial_on_portolan", age: "industrial_age", theme: "forge", plate: "portolan", w: 160, h: 48, frames: 1},
		job{name: "inspect_route_victorian_age", age: "victorian_age", theme: "forge", w: 160, h: 48, frames: 1, cursorTo: "route"},
		job{name: "other_player_medieval_age", age: "medieval_age", theme: "forge", w: 160, h: 48, frames: 1, alt: true},
		job{name: "other_player_industrial_age", age: "industrial_age", theme: "forge", w: 160, h: 48, frames: 1, alt: true},
	)
	var index []string
	for _, j := range jobs {
		src := lib
		if j.alt {
			src = alt
		}
		if src == nil || !contains(src.ages, j.age) {
			continue
		}
		if err := theme.SetActive(j.theme); err != nil {
			return err
		}
		a, err := src.atlas(j.age)
		if err != nil {
			return err
		}
		var frames []*Canvas
		sc := newScene(a)
		if j.plate != "" {
			sc.Plate = plates[j.plate]
		}
		sc.Mono = j.mono
		sc.Draw(j.w, j.h) // lay out once so the cursor has a map to live on
		if j.cursorTo != "" {
			for i, p := range sc.pois() {
				if p.Kind == j.cursorTo {
					sc.poiIdx = i - 1
					sc.Next(1)
					break
				}
			}
		}
		for f := 0; f < j.frames; f++ {
			sc.Frame = f * 3
			frames = append(frames, sc.Draw(j.w, j.h))
		}
		title := fmt.Sprintf("atlas · %s · %s · %s · %dx%d", config.AgeByKey()[j.age].Name, sc.Plate.Name, j.theme, j.w, j.h)
		if j.alt {
			title += fmt.Sprintf(" · another player (seed %d)", a.St.Seed)
		}
		if j.frames > 1 {
			title += fmt.Sprintf(" · %d frames", j.frames)
		}
		if j.mono {
			title += " · monochrome"
		}
		if j.pal16 {
			title += " · 16 colours"
		}
		if err := writeCapture(dir, j.name, title, frames, j.pal16); err != nil {
			return err
		}
		group := "Every age · dark theme · 160×48"
		for _, g := range [][2]string{{"light_", "Light themes"}, {"small_", "Small terminals"}, {"mini_", "Mini view (sidebar)"},
			{"anim_", "Animated (4 frames)"}, {"mono_", "Degraded colour"}, {"ansi16_", "Degraded colour"}, {"inspect_", "Inspector"},
			{"plates_", "Leafing back through the plates"}, {"other_", "Another player, same age"}} {
			if strings.HasPrefix(j.name, g[0]) {
				group = g[1]
			}
		}
		if len(index) == 0 || index[len(index)-1] != "" && !strings.Contains(strings.Join(index, ""), "<h2>"+group+"</h2>") {
			index = append(index, "<h2>"+group+"</h2>")
		}
		index = append(index, fmt.Sprintf(`<li><a href="%s.html">%s</a> · <a href="%s.txt">txt</a></li>`, j.name, title, j.name))
	}
	_ = theme.SetActive("forge")
	page := fmt.Sprintf(htmlHead, "atlas captures", "#15171c", "#d8d4c8", "atlas captures") +
		"<p style=\"font-size:13px;max-width:70ch;opacity:.8\">Real tcell SimulationScreen renders of the atlas prototype from smoke-bot game states. Each page has a .txt twin. See DESIGN.md.</p>" +
		"<ul style=\"font-size:13px;line-height:1.7;list-style:none;padding:0\">" + strings.Join(index, "\n") + "</ul></body></html>\n"
	return os.WriteFile(filepath.Join(dir, "index.html"), []byte(page), 0o644)
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
