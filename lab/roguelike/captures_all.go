package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/theme"
)

var captureAges = []string{
	"primitive_age", "stone_age", "bronze_age", "medieval_age", "industrial_age", "cyberpunk_age", "galactic_age",
}

type shot struct {
	name, title string
	age         string
	seed        int64
	theme       string
	zoom        int
	w, h        int
	frames      int
	compact     bool
	cata        bool
}

// focus puts the cursor on the most interesting building: the one that
// grew most since the last visit, else the newest non-housing one.
func focus(v *View) {
	best, score := -1, -1.0
	for i, b := range v.S.Blds {
		if len(b.Tiles) == 0 || b.Wonder || b.V.Lineage == "housing" || b.V.Lineage == "food" {
			continue
		}
		if rdist(b.Tiles[0].X, b.Tiles[0].Y, v.S.W.CX, v.S.W.CY) > 9 {
			continue // keep the capture's camera on the town
		}
		sc := float64(b.Delta)*10 + float64(b.V.Tier) + 0.01*float64(b.V.Count)
		if b.V.Legacy {
			sc -= 50
		}
		if b.V.Workers > 0 {
			sc += 3
		}
		if sc > score {
			best, score = i, sc
		}
	}
	if best >= 0 {
		t := v.S.Blds[best].Tiles[0]
		v.CurX, v.CurY = t.X, t.Y
	}
}

func writeAllCaptures(stateDir, out string) error {
	var shots []shot
	for _, a := range captureAges {
		shots = append(shots,
			shot{name: a, age: a, seed: 7, theme: "forge", zoom: ZSettlement, w: 160, h: 48, frames: 1},
			shot{name: a + "_region", age: a, seed: 7, theme: "forge", zoom: ZRegion, w: 160, h: 48, frames: 1},
			shot{name: a + "_mini", age: a, seed: 7, theme: "forge", w: 40, h: 15, frames: 1, compact: true},
		)
	}
	shots = append(shots,
		shot{name: "medieval_age_anim", age: "medieval_age", seed: 7, theme: "forge", zoom: ZSettlement, w: 160, h: 48, frames: 8},
		shot{name: "primitive_age_anim", age: "primitive_age", seed: 7, theme: "forge", zoom: ZSettlement, w: 120, h: 36, frames: 8},
		shot{name: "galactic_age_region_anim", age: "galactic_age", seed: 7, theme: "forge", zoom: ZRegion, w: 160, h: 48, frames: 8},
		shot{name: "medieval_age_district", age: "medieval_age", seed: 7, theme: "forge", zoom: ZDistrict, w: 160, h: 48, frames: 1},
		shot{name: "cyberpunk_age_district", age: "cyberpunk_age", seed: 7, theme: "forge", zoom: ZDistrict, w: 160, h: 48, frames: 1},
		shot{name: "medieval_age_light", age: "medieval_age", seed: 7, theme: "daylight", zoom: ZSettlement, w: 160, h: 48, frames: 1},
		shot{name: "industrial_age_light", age: "industrial_age", seed: 7, theme: "daylight", zoom: ZSettlement, w: 160, h: 48, frames: 1},
		shot{name: "galactic_age_region_light", age: "galactic_age", seed: 7, theme: "daylight", zoom: ZRegion, w: 160, h: 48, frames: 1},
		shot{name: "medieval_age_parchment", age: "medieval_age", seed: 7, theme: "parchment", zoom: ZSettlement, w: 160, h: 48, frames: 1},
		shot{name: "cyberpunk_age_cyberpunk_theme", age: "cyberpunk_age", seed: 7, theme: "cyberpunk", zoom: ZSettlement, w: 160, h: 48, frames: 1},
		shot{name: "industrial_age_100x30", age: "industrial_age", seed: 7, theme: "forge", zoom: ZSettlement, w: 100, h: 30, frames: 1},
		shot{name: "medieval_age_80x24", age: "medieval_age", seed: 7, theme: "forge", zoom: ZSettlement, w: 80, h: 24, frames: 1},
		shot{name: "medieval_age_mini_light", age: "medieval_age", seed: 7, theme: "daylight", w: 40, h: 15, frames: 1, compact: true},
		shot{name: "medieval_age_seed11", age: "medieval_age", seed: 11, theme: "forge", zoom: ZSettlement, w: 160, h: 48, frames: 1},
		shot{name: "industrial_age_catastrophe", age: "industrial_age", seed: 7, theme: "forge", zoom: ZSettlement, w: 160, h: 48, frames: 6, cata: true},
	)
	worlds := map[int64]*World{}
	plans := map[int64]*Plan{}
	var index []string
	for _, sh := range shots {
		mv, err := loadView(stateDir, sh.age, sh.seed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", sh.name, err)
			continue
		}
		if worlds[sh.seed] == nil {
			worlds[sh.seed] = NewWorld(sh.seed, worldW, worldH)
			plans[sh.seed] = NewPlan(worlds[sh.seed])
		}
		if err := theme.SetActive(sh.theme); err != nil {
			return err
		}
		v := NewView(Build(worlds[sh.seed], plans[sh.seed], mv, sh.cata))
		v.Zoom = sh.zoom
		focus(v)
		var frames []frame
		for f := 0; f < sh.frames; f++ {
			v.Frame = f
			frames = append(frames, snap(v, sh.w, sh.h, sh.compact))
		}
		title := fmt.Sprintf("%s · %s · %s · %d×%d · theme %s", sh.name, mv.AgeName, modeName(sh), sh.w, sh.h, sh.theme)
		if sh.frames > 1 {
			title += fmt.Sprintf(" · %d frames", sh.frames)
		}
		if sh.cata {
			title += " · catastrophe overlay forced on"
		}
		if err := writeCapture(out, sh.name, title, frames); err != nil {
			return err
		}
		index = append(index, fmt.Sprintf(`<li><a href="%s.html">%s</a> <small>(<a href="%s.txt">txt</a>)</small></li>`,
			sh.name, html.EscapeString(title), sh.name))
	}
	sort.Strings(index)
	_ = theme.SetActive("forge")
	page := `<!doctype html><meta charset="utf-8"><title>Glyph World captures</title>
<style>body{font-family:system-ui,sans-serif;background:#111;color:#ddd;padding:24px}a{color:#e0b040}li{margin:4px 0}</style>
<h1>Glyph World captures</h1><p>Real tcell SimulationScreen cells rendered to HTML. Every state comes from the smoke bot.</p><ul>` +
		strings.Join(index, "\n") + "</ul>\n"
	return os.WriteFile(filepath.Join(out, "index.html"), []byte(page), 0o644)
}

func modeName(sh shot) string {
	if sh.compact {
		return "mini view"
	}
	return strings.ToLower(zoomNames[sh.zoom]) + " zoom"
}
