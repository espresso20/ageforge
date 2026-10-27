package mapgen

import (
	"maps"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
)

// What a screen layer needs to know of the engine, besides its grids.

// WorldW and WorldH are the size of the world, in cells.
const (
	WorldW = worldW
	WorldH = worldH
)

// Sign is the sign hung on the buildings of a line: its symbol on the map.
func Sign(line string) rune { return lineSigns[line] }

// LineHue is the hue that tints the buildings of a line.
func LineHue(line string) Hue { return lineHue(line) }

// terrainOf is the model's terrain a biome stands for.
var terrainOf = [numBiomes]mapmodel.Terrain{
	Deep: mapmodel.TDeep, Shallows: mapmodel.TShallow, Sand: mapmodel.TBeach, Scrub: mapmodel.TGrass,
	Grass: mapmodel.TGrass, Forest: mapmodel.TForest, Hills: mapmodel.THills, Mountain: mapmodel.TMountain,
	River: mapmodel.TRiver,
}

// TerrainOf is the terrain a terrain Thing's key names; ok is false for a key
// that is no biome (space, a planet).
func TerrainOf(key string) (t mapmodel.Terrain, ok bool) {
	for b, name := range biomeNames {
		if name == key {
			return terrainOf[b], true
		}
	}
	return 0, false
}

// Runes lists every glyph the styles, the wonder templates and the signs can
// draw, each once, for a check that none lacks a plain ASCII shape.
func Runes() []rune {
	seen := map[rune]bool{}
	add := func(s string) {
		for _, r := range s {
			seen[r] = true
		}
	}
	for _, st := range styles {
		looks := append([]Look{st.Yard, st.Crops, st.Plaza, st.Rubble, st.Smoke, st.Sky}, st.Land[:]...)
		for _, look := range looks {
			add(look.Glyphs)
		}
		for _, r := range st.Roads {
			add(r.Joins)
		}
		for _, fam := range st.Shapes {
			for _, tier := range fam {
				for _, variants := range tier {
					for _, row := range variants {
						add(row)
					}
				}
			}
		}
		for _, a := range st.Filler {
			for _, row := range a {
				add(row)
			}
		}
		for _, s := range st.Scatter {
			add(s.Glyphs)
		}
		for _, a := range st.Agents {
			add(a.Glyphs)
		}
		add(st.Windows)
	}
	for _, tpl := range templates {
		add(strings.Join(tpl.Art, ""))
	}
	for _, tpl := range monuments {
		add(strings.Join(tpl.Art, ""))
	}
	for _, r := range lineSigns {
		seen[r] = true
	}
	add(joinsLight + joinsHeavy + joinsDouble + joinsRounded + "★∙")
	delete(seen, ' ')
	return slices.Sorted(maps.Keys(seen))
}

// HasTemplate reports whether a wonder has a template drawn for it.
func HasTemplate(key string) bool {
	_, ok := templates[key]
	return ok
}
