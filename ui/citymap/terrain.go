package citymap

import (
	"image"
	"image/color"

	"github.com/espresso20/ageforge/config"
)

// terrain.go turns the FBM elevation field into a theme-tinted band fill on the
// RGBA canvas. Soft and ambient — broad water, lowland, and high ground, with no
// hard edges. The colors come entirely from the active-theme palette (palette.go),
// so the terrain visibly retints when the theme changes.

// Elevation band thresholds on the [0,1] FBM field. Tuned so most of the map is
// walkable land with a little water at the low end and highlights up top.
const (
	bandDeepWater    = 0.30
	bandShallowWater = 0.40
	bandLowland      = 0.55
	bandMidland      = 0.72
	bandHighland     = 0.86
	// >= bandHighland is peak.
)

// ageInfo derives a stable seed and a faint hue shift from the current age key,
// using the canonical age order as an index. Early ages warm, late ages cool —
// a subtle progression in P1; richer per-age styling is P2.
func ageInfo(ageKey string) (seed uint32, hueShift float64) {
	order := config.AgeOrder()
	idx := -1
	for i, k := range order {
		if k == ageKey {
			idx = i
			break
		}
	}
	if idx < 0 {
		// Unknown age (e.g. a test stub): derive a seed from the string so the
		// field is still deterministic, and use a neutral hue.
		var h uint32 = 2166136261
		for i := 0; i < len(ageKey); i++ {
			h ^= uint32(ageKey[i])
			h *= 16777619
		}
		return h | 1, 0
	}
	// Seed: spread ages apart so successive ages get visibly different terrain.
	seed = uint32(idx+1) * 2654435761
	// Hue: walk from +0.5 (warm) at the first age to -0.5 (cool) at the last.
	n := len(order)
	if n > 1 {
		hueShift = 0.5 - float64(idx)/float64(n-1)
	}
	return seed, hueShift
}

// colorForElevation picks the legacy band color for an elevation value. Retained
// for the band-based unit tests and as a reference; the live terrain now paints
// from the biome classifier (biomeColor) instead.
func colorForElevation(e float64, p terrainPalette) color.RGBA {
	switch {
	case e < bandDeepWater:
		return p.deepWater
	case e < bandShallowWater:
		// Soften the shore: blend shallow→lowland across the band.
		t := (e - bandDeepWater) / (bandShallowWater - bandDeepWater)
		return blend(p.shallowWater, p.lowland, t*0.5)
	case e < bandLowland:
		return p.lowland
	case e < bandMidland:
		return p.midland
	case e < bandHighland:
		return p.highland
	default:
		return p.peak
	}
}

// ---- Biome generator (P4) ---------------------------------------------------
//
// The terrain is classified into a small set of biomes from two independent FBM
// fields — elevation (fbm) and moisture (fbmFreq at a broader scale, moistureSeed).
// Each biome maps to a theme-blended color (palette.go) so the whole biome map
// retints on a theme switch. A passability classification rides alongside: deep/
// shallow water and mountain/peak are impassable; everything else is passable.
// Roads (pathfind.go) and building placement (entities.go) consult passability so
// nothing routes through a lake or sits on a peak.

// biome is one terrain class.
type biome uint8

const (
	biomeDeepWater biome = iota
	biomeShallowWater
	biomeSand
	biomeGrass
	biomeForest
	biomeRock
	biomeMountain
	biomeSnow
	biomeCount // sentinel: number of biomes; not a real biome
)

// Moisture/elevation thresholds for biome classification on the [0,1] fields.
// Water is purely elevation-driven (low ground floods regardless of moisture);
// above the shoreline, moisture splits grass/forest and the high band into bare
// rock vs. (always) snow at the very top.
const (
	moistDry = 0.42 // below → drier classes (grass/rock); above → wetter (forest)
)

// classifyBiome assigns a biome from an (elevation, moisture) pair, both in [0,1].
// Deterministic and pure. The bands reuse the elevation thresholds so the biome
// map lines up with the legacy silhouette (water at the low end, peaks up top).
func classifyBiome(elev, moist float64) biome {
	switch {
	case elev < bandDeepWater:
		return biomeDeepWater
	case elev < bandShallowWater:
		return biomeShallowWater
	case elev < bandShallowWater+0.04:
		// A thin shoreline strip just above the waterline reads as beach/sand.
		return biomeSand
	case elev < bandMidland:
		// Lowland → mid: grass when dry, forest when damp.
		if moist >= moistDry {
			return biomeForest
		}
		return biomeGrass
	case elev < bandHighland:
		// High ground: damp high ground keeps tree/scrub cover (still forest-ish);
		// dry high ground is bare rock/hills.
		if moist >= moistDry+0.08 {
			return biomeForest
		}
		return biomeRock
	case elev < bandHighland+0.06:
		// Just below the peak: exposed mountain stone.
		return biomeMountain
	default:
		// The very top is always snow/peak regardless of moisture.
		return biomeSnow
	}
}

// passableBiome reports whether a biome can carry a road or a building. Deep and
// shallow water and the bare mountain band are impassable; the snowy peak is too
// (it sits above the mountain band). Everything from sand up through rock is land.
func passableBiome(b biome) bool {
	switch b {
	case biomeDeepWater, biomeShallowWater, biomeMountain, biomeSnow:
		return false
	default:
		return true
	}
}

// biomeColor maps a biome to its theme-blended fill color from the palette.
func biomeColor(b biome, p terrainPalette) color.RGBA {
	switch b {
	case biomeDeepWater:
		return p.bDeepWater
	case biomeShallowWater:
		return p.bShallowWater
	case biomeSand:
		return p.bSand
	case biomeGrass:
		return p.bGrass
	case biomeForest:
		return p.bForest
	case biomeRock:
		return p.bRock
	case biomeMountain:
		return p.bMountain
	case biomeSnow:
		return p.bSnow
	default:
		return p.bGrass
	}
}

// terrainField is the per-render biome + passability map, computed once from the
// two noise fields and reused by the terrain paint, the road pathfinder, and the
// placement nudge. The grids are full canvas resolution (one entry per pixel),
// indexed [y*w+x]. Building it once keeps the per-render A* and nudges cheap (no
// re-sampling of FBM per query). w/h are the image pixel dimensions.
type terrainField struct {
	w, h     int
	biomes   []biome
	passable []bool
}

// newTerrainField samples elevation + moisture across the canvas, classifies every
// pixel into a biome, and records its passability. Pure + deterministic from seed.
func newTerrainField(w, h int, seed uint32) *terrainField {
	f := &terrainField{w: w, h: h}
	if w <= 0 || h <= 0 {
		return f
	}
	mSeed := moistureSeed(seed)
	f.biomes = make([]biome, w*h)
	f.passable = make([]bool, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			e := fbm(float64(x), float64(y), seed)
			// Moisture at a broader base frequency so damp/dry regions are larger than
			// the elevation features and don't merely trace the contour lines.
			m := fbmFreq(float64(x), float64(y), mSeed, 1.0/64.0)
			b := classifyBiome(e, m)
			idx := y*w + x
			f.biomes[idx] = b
			f.passable[idx] = passableBiome(b)
		}
	}
	return f
}

// at returns the biome at (x,y), or biomeGrass for out-of-range coords (treated as
// open land so callers never index out of bounds).
func (f *terrainField) at(x, y int) biome {
	if f == nil || x < 0 || y < 0 || x >= f.w || y >= f.h || len(f.biomes) == 0 {
		return biomeGrass
	}
	return f.biomes[y*f.w+x]
}

// passableAt reports passability at (x,y); out-of-range is impassable (the canvas
// edge is a wall for routing/placement so paths and slots stay on-canvas).
func (f *terrainField) passableAt(x, y int) bool {
	if f == nil || x < 0 || y < 0 || x >= f.w || y >= f.h || len(f.passable) == 0 {
		return false
	}
	return f.passable[y*f.w+x]
}

// drawTerrainField fills img from a precomputed biome field. Each pixel takes its
// biome's theme-blended color, so the terrain retints with the theme and the paint
// matches exactly the grid the roads + placement reason over.
func drawTerrainField(img *image.RGBA, p terrainPalette, f *terrainField) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			img.SetRGBA(x, y, biomeColor(f.at(x, y), p))
		}
	}
}

// drawTerrain fills img with the procedural, theme-tinted biome terrain, building
// the field internally. Retained as the single-shot entry point (and for any caller
// that doesn't need the field afterward); the render path uses newTerrainField +
// drawTerrainField directly so it can reuse the field for roads + placement.
func drawTerrain(img *image.RGBA, p terrainPalette, seed uint32) {
	b := img.Bounds()
	f := newTerrainField(b.Dx(), b.Dy(), seed)
	drawTerrainField(img, p, f)
}

// ---- World-scale field (worldmap.go) ----------------------------------------
//
// The world map needs the SAME biome vocabulary as the city map — so it retints on
// theme switch and shares classifyBiome/passableBiome/biomeColor — but sampled so it
// reads as a real CONTINENT surrounded by ocean, with coastlines, rivers, and relief,
// rather than an edge-to-edge noise field. That geography now lives in the seeded
// worldModel (worldmodel.go): a heightmap shaped by a radial island mask, downhill
// rivers, and relief anchors. newWorldTerrainField is the thin adapter that hands the
// model's per-pixel biome/passability grid back as a *terrainField, so every existing
// consumer (terrain paint, settlement land-gate, civ-dot land snap) is unchanged — it
// just now sees a continent. Callers that need the richer geometry (coastline / rivers /
// relief for the neutral render) build the full model via buildWorldModel directly.

// newWorldTerrainField builds the world's per-pixel biome + passability field by
// constructing the seeded continent model and returning its field. Continents + oceans
// with real coastlines (from the island mask), classified with the SAME
// classifyBiome/passableBiome the city map uses. Pure + deterministic from seed; the
// seed must be stable per account (see worldTerrainSeed) so the continents don't
// rearrange across ages. Panic-safe on tiny/zero canvases (the model returns an empty
// field whose at/passableAt guards cover all queries).
func newWorldTerrainField(w, h int, seed uint32) *terrainField {
	return buildWorldModel(w, h, seed).field
}

// worldTerrainSeed derives a STABLE-per-account world seed from the player's display
// name, so a given player's continents are fixed and — crucially — independent of age/
// progress (aging up must not rearrange the land). FNV-1a over the name, salted so it
// doesn't collide with the age seeds used elsewhere. An empty/anonymous name falls back
// to a fixed non-zero seed so the anonymous world is still stable and non-degenerate.
func worldTerrainSeed(displayName string) uint32 {
	if displayName == "" {
		return 0x6b8a9f01 // stable anonymous world
	}
	var h uint32 = 2166136261
	for i := 0; i < len(displayName); i++ {
		h ^= uint32(displayName[i])
		h *= 16777619
	}
	h ^= 0x7f4a7c15 // salt: keep world seeds off the age-seed lattice
	return h | 1
}
