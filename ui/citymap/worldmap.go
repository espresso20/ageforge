package citymap

import (
	"image"
	"image/color"
	"math"
	"sort"
	"sync"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// worldmap.go is the second map view — the "Known World." Where CityMap renders the
// player's own settlement (terrain + the buildings they've built), WorldMap zooms out
// to the wider world: a procedural, Game-of-Life-style field of anonymous settlements
// (the backdrop) with the player's civilization and the discovered diplomacy civs
// called out on top as distinct, labeled, relationship-colored dots.
//
// It deliberately reuses CityMap's render primitives rather than reinventing them:
//   - the half-block streamer (streamHalfBlocks) and the same (cols × rows*2) pixel
//     canvas / '▄' cell model,
//   - the noise field (fbm) for a muted terrain/ocean wash under the dots,
//   - the palette helpers (rgba/blend/brighten + theme roles) so every layer retints
//     on a theme switch,
//   - the overlay text-pass (overlayPlan / overlayLabel / stampOverlay): labels store
//     GEOMETRY + a theme ROLE, never a baked color, so a theme switch retints the text
//     without re-laying-out, exactly like CityMap.
//
// Lock discipline mirrors CityMap: Refresh only stores a snapshot, all rendering reads
// the stored snapshot under the mutex, nothing here touches the engine or acquires an
// engine lock. The diplomacy data is read straight off the snapshot's Diplomacy.Factions.
//
// A note on civ-dot SIZE: the design calls for sizing each civ by "Strength," but the
// UI projection game.FactionInfo carries no Strength field (that lives on the internal
// faction def, not the snapshot). We size from what the snapshot DOES expose — a
// strength proxy built from Opinion magnitude plus a standing weight (allies/at-war
// read as larger, more consequential neighbors) — so the dots still convey relative
// weight without reaching past the snapshot contract.

// WorldMap is the overlay widget renderer for the world view. Like CityMap it satisfies
// the OverlayManager widget contract via Build (returns the primitive) and Refresh
// (stores a fresh snapshot).
type WorldMap struct {
	mu    sync.Mutex
	state game.GameState

	// Cache of the last rendered image AND its overlay plan, plus the key that produced
	// them — same scheme as CityMap. The geometry depends on age/size/building-count/
	// discovered-civ-set (NOT the theme); the overlay's colors resolve live at draw
	// time, so a theme switch retints without invalidating the cached layout. We fold a
	// cheap signature of the discovered-civ set into the key so discovering/losing a civ
	// (or a civ's status flipping) re-lays-out the dots.
	cachedImg      *image.RGBA
	cachedPlan     overlayPlan
	cachedW        int
	cachedH        int
	cachedAge      string
	cachedBld      int
	cachedCivSig   uint64
	cachedThemeKey string
}

// NewWorldMap creates a WorldMap renderer.
func NewWorldMap() *WorldMap { return &WorldMap{} }

// Build returns the tview.Primitive that draws the world map. The draw closure reads
// the latest stored snapshot each frame, (re)generates the image on demand, streams it
// as half-blocks, then stamps the crisp text overlay on top — colors resolved live so a
// theme switch retints the labels too. Mirrors CityMap.Build.
func (wm *WorldMap) Build(state game.GameState) tview.Primitive {
	wm.mu.Lock()
	wm.state = state
	wm.mu.Unlock()

	box := tview.NewBox().SetBorder(false)
	box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		if width <= 0 || height <= 0 {
			return x, y, width, height
		}
		img, plan := wm.imageFor(width, height)
		streamHalfBlocks(screen, img, x, y, width, height)
		stampOverlay(screen, plan, x, y, width, height)
		return x, y, width, height
	})
	return box
}

// Refresh stores a fresh state snapshot. Safe from any goroutine; no rendering, no
// engine locks. Mirrors CityMap.Refresh.
func (wm *WorldMap) Refresh(state game.GameState) {
	wm.mu.Lock()
	wm.state = state
	wm.mu.Unlock()
}

// imageFor returns the rendered image AND its overlay plan for a (width,height) cell
// area, regenerating only when the cache key changes. width/height are in cells; the
// image is (width × height*2) pixels (one half-block = two stacked pixels). Mirrors
// CityMap.imageFor but adds the discovered-civ signature to the key.
func (wm *WorldMap) imageFor(width, height int) (*image.RGBA, overlayPlan) {
	imgW := width
	imgH := height * 2

	wm.mu.Lock()
	defer wm.mu.Unlock()

	st := wm.state
	bld := builtBuildingCount(st)
	civSig := civSignature(st)
	themeKey := theme.Active().Key

	if wm.cachedImg != nil &&
		wm.cachedW == imgW && wm.cachedH == imgH &&
		wm.cachedAge == st.Age && wm.cachedBld == bld &&
		wm.cachedCivSig == civSig && wm.cachedThemeKey == themeKey {
		return wm.cachedImg, wm.cachedPlan
	}

	img, plan := renderWorldImage(st, imgW, imgH)
	wm.cachedImg = img
	wm.cachedPlan = plan
	wm.cachedW = imgW
	wm.cachedH = imgH
	wm.cachedAge = st.Age
	wm.cachedBld = bld
	wm.cachedCivSig = civSig
	wm.cachedThemeKey = themeKey
	return img, plan
}

// civSignature folds the discovered-civ set (names + status + at-war + a strength
// proxy bucket) into a stable hash so the cache invalidates when the diplomacy picture
// changes — a new neighbor discovered, a status flip to war, etc. Pure read of the
// snapshot; map iteration order doesn't matter because we sort first.
func civSignature(state game.GameState) uint64 {
	cs := worldCivs(state)
	var h uint64 = 1469598103934665603 // FNV-1a 64 offset
	mix := func(s string) {
		for i := 0; i < len(s); i++ {
			h ^= uint64(s[i])
			h *= 1099511628211
		}
		h ^= 0xff // separator
		h *= 1099511628211
	}
	for _, c := range cs {
		mix(c.name)
		mix(string(rune('0' + c.sizeBucket)))
		if c.atWar {
			mix("war")
		}
		mix(string(rune('A' + int(c.role)%26)))
	}
	return h
}

// renderWorldImage composites the full world map and returns the pixel image plus the
// cell-space overlay plan. Layers, in order:
//
//	terrain   — a MUTED world biome field (continents + oceans, newWorldTerrainField),
//	            blended hard toward the theme background so it reads as a dim strategic
//	            backdrop — continents-vs-sea at a glance — while staying dark enough for
//	            the bright civ dots to pop.
//	backdrop  — the sparse procedural settlement field: a scatter of small, dim distant
//	            marks in a dim land tone; count scales modestly with progress. Gated to
//	            land — no settlement sits in the ocean.
//	your civ  — one prominent accent dot at a stable spot (center), snapped onto land.
//	civ dots  — each discovered diplomacy civ, ringed around your civ deterministically,
//	            snapped onto land, sized by a strength proxy, colored by its own faction
//	            hue with a standing ring, and — at war — a hot-red override + ⚔ mark.
//
// Two SEPARATE seeds drive this. The land uses worldTerrainSeed(displayName) — stable per
// account, INDEPENDENT of age/progress, so the continents never rearrange when you age up.
// The settlement scatter + civ ring use the age seed (ageInfo) so the peopling of the world
// still grows with progress on top of the fixed land. The overlay plan (your-civ label, civ
// labels, title) is then computed from the same geometry so labels land on the dots. Pure:
// reads the snapshot + config-by-key only.
func renderWorldImage(state game.GameState, w, h int) (*image.RGBA, overlayPlan) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if w <= 0 || h <= 0 {
		return img, overlayPlan{}
	}

	seed, hueShift := ageInfo(state.Age)
	pal := buildPalette(hueShift)
	prog := worldProgress(state)

	// Seeded world MODEL, built ONCE and threaded into every layer below: a real continent
	// (island-masked heightmap) with coastline, rivers, and relief, plus the per-pixel
	// biome/passability field that paints the backdrop AND gates/snaps every dot to land.
	// Its seed is the stable-per-account world seed (display name), NOT the age seed, so the
	// land stays put across ages while the settlement scatter and civ ring keep moving with
	// progress. The field seam (wf) is the same *terrainField every consumer already used.
	displayName := worldDisplayName(state)
	model := buildWorldModel(w, h, worldTerrainSeed(displayName))
	wf := model.field

	// 1) Terrain layer, drawn in this AGE'S cartographic MEDIUM (Phase B). mediumForAge maps
	//    four validation ages to bespoke mediums (charcoal / parchment / satellite / neon) and
	//    every other age to the neutral atlas — the exact Phase-A look. The medium re-skins the
	//    LAND (ocean, coast, rivers, relief, chrome); the civ dots + labels below are shared
	//    across all mediums so the player's markers read the same everywhere. The atlas medium
	//    still uses `pal` internally, so the default remains theme-aware and hue-shifts per age.
	med := mediumForAge(state.Age)
	med.draw(img, model, med)

	// 2) Backdrop: the sparse settlement field — the wider world — gated to land.
	geo := drawSettlementField(img, pal, wf, seed, prog)

	// 3) Your civ: a single prominent dot at the stable anchor (center), snapped to land.
	geo.you = drawYourCiv(img, pal, wf, w, h)

	// 4) Discovered diplomacy civs: ringed around your civ, snapped to land, sized/colored
	//    by standing.
	geo.civs = drawWorldCivs(img, pal, state, wf, seed, geo.you.cx, geo.you.cy)

	// 5) Overlay plan (cell space): cols=w, rows=h/2.
	plan := buildWorldOverlay(state, w, h/2, geo)
	return img, plan
}

// worldDisplayName resolves the player's display name from the snapshot with the same
// fallback the your-civ label uses, so the world-terrain seed and the label agree on the
// name. Empty/anonymous accounts fall back to "" here (worldTerrainSeed maps that to its
// fixed anonymous world); the LABEL applies its own "Your Empire" text fallback separately.
func worldDisplayName(state game.GameState) string {
	if state.AccountStats != nil {
		return state.AccountStats.DisplayName
	}
	return ""
}

// ---- progress / scaling -----------------------------------------------------

// worldProgress maps the player's advancement to a [0,1] scalar that drives backdrop
// density + dot sizes: an early world is sparse + small, a late world dense + large. It
// blends two signals — the age's index in the canonical age order (broad era progress)
// and the total built-building count (empire scale within an era) — so advancing an age
// OR sprawling a bigger empire both grow the world. Deterministic and pure.
func worldProgress(state game.GameState) float64 {
	order := config.AgeOrder()
	ageFrac := 0.0
	if n := len(order); n > 1 {
		idx := -1
		for i, k := range order {
			if k == state.Age {
				idx = i
				break
			}
		}
		if idx >= 0 {
			ageFrac = float64(idx) / float64(n-1)
		}
	}
	// Empire scale: total instances built across all distinct types, saturating so a
	// late hoard doesn't blow past 1. ~40 buildings reads as "fully built out."
	var total int
	for _, bs := range state.Buildings {
		if bs.Count > 0 {
			total += bs.Count
		}
	}
	bldFrac := math.Min(1.0, float64(total)/40.0)

	// Weight era progress a bit more than raw count — the age is the headline driver.
	p := 0.62*ageFrac + 0.38*bldFrac
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	return p
}

// ---- geometry carriers ------------------------------------------------------

// worldDot is a placed foreground dot (your civ or a neighbor) in PIXEL space, carried
// out of the pixel pass so the overlay can label it at the right cell.
type worldDot struct {
	cx, cy int    // dot center in pixels
	radius int    // dot radius in pixels (drives where the label sits)
	name   string // display name for the label
	role   theme.Role
	atWar  bool
}

// worldGeometry is the layout handed from the pixel pass to the overlay pass.
type worldGeometry struct {
	you  worldDot
	civs []worldDot
}

// ---- terrain backdrop -------------------------------------------------------

// worldTerrainMix is how far each biome tone is pulled toward the theme background before
// it lands on the canvas: 0 = full biome color (the loud city view), 1 = pure background
// (invisible). We keep ~0.70 so the world reads as continents-vs-sea at a glance but stays
// a DIM strategic backdrop the bright civ dots pop off of — i.e. ~30% biome, ~70% bg.
const worldTerrainMix = 0.70

// mutedBiomeTones precomputes the MUTED strategic-backdrop tone for every biome once (a
// lerp per pixel across a full-canvas field is wasteful — there are only biomeCount of
// them). Each biome takes its theme-blended tone (biomeColor — the SAME classifier/
// palette the city map uses, so the world retints on a theme switch) pulled hard toward
// RoleBackground so nothing competes with the foreground dots. Water is pulled a touch
// LESS toward bg so the sea reads as a distinct dark basin rather than washing into land.
func mutedBiomeTones(pal terrainPalette) []color.RGBA {
	bg := rgba(theme.Color(theme.RoleBackground))
	muted := make([]color.RGBA, biomeCount)
	for bi := biome(0); bi < biomeCount; bi++ {
		mix := worldTerrainMix
		if bi == biomeDeepWater || bi == biomeShallowWater {
			mix = worldTerrainMix - 0.08 // seas a shade more saturated so coasts read
		}
		muted[bi] = blend(bg, biomeColor(bi, pal), 1.0-mix)
	}
	return muted
}

// atlasBiomeTones is the neutral render's map palette. The strategic-backdrop muting
// (blending every biome toward the dark background) greyed the land into an olive mud, so
// the atlas instead builds VIVID, ZONED, legibly-colored land — grassland green, forest a
// darker green, beach tan, hills brown-grey, mountains pale grey, peaks near-white — while
// the sea stays a dark depth-banded blue for strong land/water contrast.
//
// It stays THEME-DERIVED: each biome starts from the palette's theme-blended biome color
// (so a theme switch still retints the whole map) and is pulled toward a canonical map
// anchor + brightened to a legible level. The anchors are the muted hue targets a paper
// atlas uses; blending the live biome color toward them keeps the theme's character while
// guaranteeing the land reads as land, not grey.
func atlasBiomeTones(pal terrainPalette) []color.RGBA {
	bg := rgba(theme.Color(theme.RoleBackground))
	tones := make([]color.RGBA, biomeCount)

	// Canonical map anchors (muted, not cartoon). Land biomes blend toward these so the
	// zoning is legible regardless of how dark the theme's raw biome tone is.
	grassA := color.RGBA{R: 0x6f, G: 0x9e, B: 0x54, A: 0xff}  // inviting grassland green
	forestA := color.RGBA{R: 0x3f, G: 0x6b, B: 0x3c, A: 0xff} // deeper forest green
	sandA := color.RGBA{R: 0xd7, G: 0xc4, B: 0x8f, A: 0xff}   // warm beach tan
	hillA := color.RGBA{R: 0x9a, G: 0x8b, B: 0x6d, A: 0xff}   // brown-grey uplands
	mtnA := color.RGBA{R: 0xa9, G: 0xa4, B: 0x9c, A: 0xff}    // pale rocky grey
	snowA := color.RGBA{R: 0xe9, G: 0xed, B: 0xf0, A: 0xff}   // snow off-white

	// Blend each theme biome color toward its map anchor. 0.62 keeps a clear majority of
	// the legible anchor while still letting the theme tint through (retint on switch).
	const toward = 0.62
	tones[biomeGrass] = blend(biomeColor(biomeGrass, pal), grassA, toward)
	tones[biomeForest] = blend(biomeColor(biomeForest, pal), forestA, toward)
	tones[biomeSand] = blend(biomeColor(biomeSand, pal), sandA, toward)
	tones[biomeRock] = blend(biomeColor(biomeRock, pal), hillA, toward)
	tones[biomeMountain] = blend(biomeColor(biomeMountain, pal), mtnA, toward)
	tones[biomeSnow] = blend(biomeColor(biomeSnow, pal), snowA, toward)

	// Sea: keep it dark and blue (fuller mute toward background) so the vivid continent
	// pops against a dim basin.
	tones[biomeDeepWater] = blend(bg, biomeColor(biomeDeepWater, pal), 1.0-(worldTerrainMix-0.08))
	tones[biomeShallowWater] = blend(bg, biomeColor(biomeShallowWater, pal), 1.0-(worldTerrainMix-0.08))
	return tones
}

// drawWorldTerrain paints the world biome field as a MUTED strategic backdrop: each
// pixel takes its biome's precomputed muted tone (see mutedBiomeTones). This is the
// BASE fill — the neutral render (drawWorldModel) lays coastline / rivers / relief on
// top. Retained as its own entry point because a test drives it directly to assert the
// settlement layer paints zero pixels on water. Panic-safe: wf.at() returns grass for
// any out-of-range/empty query, so a tiny or zero field simply paints a uniform dim land.
func drawWorldTerrain(img *image.RGBA, pal terrainPalette, wf *terrainField) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return
	}
	muted := mutedBiomeTones(pal)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := wf.at(x, y)
			if bi >= biomeCount {
				bi = biomeGrass
			}
			img.SetRGBA(b.Min.X+x, b.Min.Y+y, muted[bi])
		}
	}
}

// drawWorldModelAtlas is the NEUTRAL atlas render of the seeded continent: the clean,
// readable DEFAULT medium (mediumForAge falls through to this for every age without a
// bespoke cartographic medium). It composites, in order:
//
//	ocean    — depth-banded blues: the deep basin darkens with distance below sea level,
//	           the shallow shelf near the coast lightens, so the sea has real bathymetry.
//	land     — the biome-toned muted fill (grass/forest/hill/mountain/snow/sand) from the
//	           shared palette, so the interior reads as terrain and retints on a theme switch.
//	coastline— a crisp darker stroke exactly along the land/ocean boundary (a land pixel
//	           with a water 4-neighbour), so the continent has a drawn shore.
//	rivers   — thin blue-ish polylines along each watercourse, widening toward the mouth
//	           with accumulated flow, so drainage reads as rivers reaching the sea.
//	relief   — small mountain (caret), hill (bump), and forest (dab) symbols at the sparse
//	           relief anchors, readable at half-block scale (highland body + a shadow pixel).
//
// Muted toward the background throughout so the atlas stays a dim strategic backdrop the
// bright civ dots pop off of. Theme-aware: every tone is a live palette role. Panic-safe:
// an empty model (tiny/zero canvas) has no rivers/relief and a guarded field, so this
// paints a uniform dim land and returns.
//
// It takes the atlas mediumPalette (built in atlasMedium) so the shared coast/river/relief
// tones come from one place, but the depth-banding + biome fill read the atlasBiomeTones
// directly (unchanged from before the medium refactor) so the default look is byte-for-byte
// what Phase A shipped.
func drawWorldModelAtlas(img *image.RGBA, pal terrainPalette, m *worldModel, mp mediumPalette) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	// Atlas tones: a lit, legible, vividly-zoned continent over a dark sea. Land vs water
	// is the map's primary read, so land must sit clearly above the sea.
	muted := atlasBiomeTones(pal)

	// 1) Base fill: depth-banded ocean + biome-toned land. For water pixels we shade by
	//    how far the elevation sits below sea level (deep = darker, shelf = lighter) so the
	//    sea has bathymetry instead of a flat blue. Land takes its lit atlas biome tone.
	deepTone := muted[biomeDeepWater]
	shelfTone := muted[biomeShallowWater]
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := m.field.at(x, y)
			if bi >= biomeCount {
				bi = biomeGrass
			}
			col := muted[bi]
			if bi == biomeDeepWater || bi == biomeShallowWater {
				// Depth 0 at the coast → 1 at the deepest sea. Shade deep→shelf by depth so
				// the basin darkens away from shore.
				e := m.elevAtPx(x, y)
				depth := 0.0
				if m.seaLevel > 0 {
					depth = clamp01((m.seaLevel - e) / m.seaLevel)
				}
				col = blend(shelfTone, deepTone, depth)
			}
			img.SetRGBA(b.Min.X+x, b.Min.Y+y, col)
		}
	}

	// 2) Coastline stroke: a crisp SHORE rim where land meets water. The ocean is already
	//    dark, so a dark stroke would just merge into it (a muddy moat) — instead the shore
	//    is a LIGHTER beach tone (the sand biome color lifted), which reads as a clean bright
	//    coastline outlining the continent against the dark sea. A land pixel with at least one
	//    water 4-neighbour is a shore pixel. Second pass so it reads over the base fill.
	coast := mp.coast // atlasMedium sets this to brighten(muted[biomeSand], 0.10)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			if isLandPx(m.field, x-1, y) && isLandPx(m.field, x+1, y) &&
				isLandPx(m.field, x, y-1) && isLandPx(m.field, x, y+1) {
				continue // interior land — not a shore pixel
			}
			img.SetRGBA(b.Min.X+x, b.Min.Y+y, coast)
		}
	}

	// 3) Rivers: thin blue courses widening to the mouth, drawn over the LAND (clipped to
	//    land so they stop at the coast, not splash into the sea). River tone is a clear,
	//    bright water blue — the palette's shallow-water tone pushed toward a bright blue —
	//    so a watercourse reads as a distinct blue thread over the lit land. Width grows with
	//    the sqrt of accumulated flow (thin at the source, widening toward the sea).
	riverTone := mp.river // atlasMedium sets this to the bright-blue shallow-water tone
	maxFlow := m.maxRiverFlow()
	for _, r := range m.rivers {
		drawRiver(img, r, riverTone, maxFlow, m)
	}

	// 4) Relief symbols: sparse mountain / hill / forest dabs. Highland/peak body with a
	//    single shadow pixel so each reads as a tiny raised mark at half-block scale.
	drawRelief(img, pal, muted, m)
}

// isLandPx reports whether pixel (x,y) is passable land in the field. Out-of-range is
// treated as WATER (false) so the canvas edge counts as ocean for the coastline test —
// a continent that runs to the frame edge still gets a shore stroke there.
func isLandPx(f *terrainField, x, y int) bool {
	if f == nil || x < 0 || y < 0 || x >= f.w || y >= f.h {
		return false
	}
	return f.passableAt(x, y)
}

// maxRiverFlow returns the peak accumulated flow across all rivers (drives per-river
// width normalization). 0 when there are no rivers.
func (m *worldModel) maxRiverFlow() float64 {
	mx := 0.0
	for _, r := range m.rivers {
		for _, f := range r.flow {
			if f > mx {
				mx = f
			}
		}
	}
	return mx
}

// drawRiver strokes one watercourse over the LAND: a Bresenham line between each
// consecutive point, thickened by a radius that grows with accumulated flow (thin at the
// source, widening toward the mouth). Every stamped pixel is gated to land so the river
// stops crisply at the coast instead of splashing blue into the sea (which would speckle
// the ocean). maxFlow normalizes the width so the biggest trunk tops out at ~3px.
func drawRiver(img *image.RGBA, r river, tone color.RGBA, maxFlow float64, m *worldModel) {
	if len(r.pts) < 2 {
		return
	}
	if maxFlow < 1 {
		maxFlow = 1
	}
	put := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for i := 0; i+1 < len(r.pts); i++ {
		a := r.pts[i]
		b := r.pts[i+1]
		// Width from the DOWNSTREAM point's flow so the mouth end of each segment is the
		// wider one. sqrt so growth tapers; capped at rx=1 (a 3px-wide brush) so even the
		// trunk stays a thread, not a canal.
		fl := r.flow[i+1] / maxFlow
		rad := int(math.Round(math.Sqrt(fl) * 1.4))
		if rad < 0 {
			rad = 0
		}
		if rad > 1 {
			rad = 1
		}
		strokeThickLineFunc(a.x, a.y, b.x, b.y, rad, tone, put)
	}
}

// strokeThickLine draws a Bresenham line from (x0,y0) to (x1,y1), stamping a small
// aspect-corrected dab of radius r at each step so the line has width. r==0 is a 1px
// line. The dab is drawn wider than tall (rx≥ry) so it reads round on the 1×2 half-block
// canvas. Every pixel goes through the put callback, so the caller can clip (e.g. rivers
// gate to land). r==0 stamps a single pixel per step.
func strokeThickLineFunc(x0, y0, x1, y1, r int, col color.RGBA, put func(x, y int, c color.RGBA)) {
	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	// Dab radii: wider than tall so the stroke reads round on half-blocks.
	rx, ry := r, r-1
	if ry < 0 {
		ry = 0
	}
	err := dx + dy
	for {
		if r <= 0 {
			put(x0, y0, col)
		} else {
			for ddy := -ry; ddy <= ry; ddy++ {
				for ddx := -rx; ddx <= rx; ddx++ {
					var ex, ey float64
					if rx > 0 {
						ex = float64(ddx) / float64(rx)
					}
					if ry > 0 {
						ey = float64(ddy) / float64(ry)
					}
					if ex*ex+ey*ey > 1.0 {
						continue
					}
					put(x0+ddx, y0+ddy, col)
				}
			}
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// drawRelief stamps the sparse relief symbols. Each anchor kind gets a tiny, readable
// mark at half-block scale:
//
//	mountain — a small caret/triangle: a highland/peak-toned apex over a wider base with a
//	           shadow pixel on the SE flank, so it reads as a raised peak, not a blob.
//	hill     — a low bump: a short highland dab with a shadow pixel beneath.
//	forest   — a small forest-toned dab (a couple of pixels) — tree cover, quiet.
//
// Tones come from the muted biome palette (so relief retints with the theme and sits in
// the same muted family as the fill); the shadow is the palette shadow role.
func drawRelief(img *image.RGBA, pal terrainPalette, muted []color.RGBA, m *worldModel) {
	if len(m.reliefs) == 0 {
		return
	}
	// Relief tones: lift the muted biome tone a touch so a symbol reads over its own fill,
	// with a shared shadow for the SE flank that grounds each mark.
	peakTone := brighten(muted[biomeMountain], 0.14)
	snowTone := brighten(muted[biomeSnow], 0.10)
	hillTone := brighten(muted[biomeRock], 0.12)
	forestTone := brighten(muted[biomeForest], 0.10)
	shadow := pal.shadow

	// land-clipped stamp: relief symbols spread a pixel or two off their anchor, which can
	// spill onto an adjacent sea pixel and speckle the ocean. Gate every relief pixel to
	// land so the sea stays clean.
	put := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, a := range m.reliefs {
		switch a.kind {
		case reliefMountain:
			// Apex + base + SE shadow → a little peak. Snowy peaks get the brighter cap.
			apex := peakTone
			if m.field.at(a.x, a.y) == biomeSnow {
				apex = snowTone
			}
			put(a.x, a.y-1, apex)     // summit
			put(a.x-1, a.y, apex)     // base left
			put(a.x, a.y, apex)       // base center
			put(a.x+1, a.y, apex)     // base right
			put(a.x+1, a.y+1, shadow) // SE flank shadow
		case reliefHill:
			// Low bump: a short cap over a shadowed foot.
			put(a.x, a.y, hillTone)
			put(a.x-1, a.y, hillTone)
			put(a.x+1, a.y, hillTone)
			put(a.x, a.y+1, shadow)
		case reliefForest:
			// A small tree-cover dab — kept quiet (a couple of pixels), no shadow.
			put(a.x, a.y, forestTone)
			put(a.x-1, a.y, forestTone)
			put(a.x, a.y-1, forestTone)
		}
	}
}

// ---- backdrop: Game-of-Life settlement field --------------------------------

// drawSettlementField paints the procedural backdrop: a SPARSE scatter of small, dim
// "settlement" marks — the wider world reading as a few far-off settlements in a big dark
// world, not a snowstorm of noise. Construction:
//
//  1. build a sparse boolean grid via settlementGrid (a low density threshold that climbs
//     only modestly with progress, then one de-clumping pass — see settlementGrid),
//  2. render each live cell as a SMALL mark (mostly a single pixel; a stable minority a
//     touch wider), drawn wider-than-tall via fillDotAspect so it reads round on the 1×2
//     half-block canvas instead of a fat cross, in a DIM land tone (RoleDim pulled a
//     touch toward the land) so it's unmistakably background behind the bright civ dots.
//
// Deterministic from seed+progress; theme-aware (the dim tone is a live theme role).
// Land-gated: a candidate cell is only peopled when the world field reports its center
// PASSABLE, so ZERO settlement pixels land in the ocean. Returns an (empty) worldGeometry
// the caller fills with the foreground dots — kept as the single geometry hand-off so
// future backdrop callouts can ride along.
func drawSettlementField(img *image.RGBA, pal terrainPalette, wf *terrainField, seed uint32, progress float64) worldGeometry {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return worldGeometry{}
	}

	// Coarse cell grid: each cell spans ~3px so the field is sparse and the dots have
	// room. Guard tiny canvases.
	gw := w / settlementCellPx
	gh := h / settlementCellPx
	if gw < 3 || gh < 3 {
		return worldGeometry{}
	}

	// Generate the Game-of-Life cell field (seed density scaled by progress, then a few
	// Conway generations to cluster it). Extracted so a test can count live cells across
	// progress levels without rasterizing through the wash.
	grid := settlementGrid(seed, gw, gh, progress)

	// Settlement speck tone: a small DARK mark that reads as a distant town cluster on the
	// now-vivid land — the grass tone deepened toward the shadow role, so it sits quietly on
	// the map (clearly behind the bright foreground civ dots) rather than as bright noise.
	// Lives in the theme (RoleDim/shadow) → retints on a switch.
	dim := rgba(theme.Color(theme.RoleDim))
	landDim := blend(darken(pal.bGrass, 0.28), dim, 0.35)

	// Render each live cell as a SMALL, dim mark — a distant settlement, not a blob. Most
	// are a single pixel; a stable minority get a 2-wide × 0-tall (i.e. 3px across, 1px
	// tall) speck so the field has a little variety without any cell reading as a fat
	// cross. Marks are drawn WIDER than tall on purpose: the half-block canvas is 1px wide
	// × 2px tall per cell, so an equal-radius dot would stretch vertically into a "+"; a
	// horizontal bias renders round-ish. Dim land tone keeps them clearly in the backdrop.
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			if !grid[gy*gw+gx] {
				continue
			}
			cx := gx*settlementCellPx + settlementCellPx/2
			cy := gy*settlementCellPx + settlementCellPx/2
			// Land gate: a settlement only exists where the world field is passable at the
			// cell center. Ocean/lake/mountain cells are skipped entirely — the wider world is
			// peopled on land, never at sea. (wf.passableAt is false off-canvas too, so edge
			// cells that spill past the field are dropped rather than drawn.)
			if !wf.passableAt(cx, cy) {
				continue
			}
			// Stable per-cell hash drives both a slight brightness jitter (so the field
			// isn't stamped-uniform) and the occasional wider speck.
			j := hashUnit(uint32(gx), uint32(gy), settlementSeed(seed)^0x5bd1e995)
			tone := blend(landDim, darken(landDim, 0.22), j)
			// ~30% of settlements are a touch wider (still 1px tall); the rest single dots.
			// A faint extra size lift at high progress reads as "more established" worlds
			// without ever becoming a cross.
			rx := 0
			if j > 0.70 || (progress > 0.6 && j > 0.45) {
				rx = 1
			}
			fillDotAspect(img, cx, cy, rx, 0, tone)
		}
	}
	return worldGeometry{}
}

// settlementCellPx is the pixel span of one backdrop cell. Widened from 3→6 so the
// coarse grid has FEWER cells (a quarter as many) and the settlements read as a sparse
// scatter of distant marks with dark space between them, not a dense field of noise.
const settlementCellPx = 6

// settlementSeed derives the backdrop's grid seed from the age seed with a 32-bit
// golden-ratio salt (the age seed is uint32, so no 64-bit constant). Stable per render.
func settlementSeed(ageSeed uint32) uint32 { return ageSeed ^ 0x9e3779b9 }

// settlementGrid builds the backdrop settlement field for a gw×gh coarse grid at a given
// progress. The earlier version seeded a dense field (12–38%) and ran Conway B3/S23
// generations, which over-clustered into a solid snowstorm of noise. We now use a SPARSE
// Poisson-ish scatter: each cell is live with a low probability that climbs only modestly
// with progress (~4% at progress 0 → ~15% at progress 1), sampled from a spatially
// uncorrelated hash so the live cells land as scattered distant settlements rather than a
// packed field. A single de-clumping pass then clears any cell with 3+ live Moore
// neighbours, so the rare coincidental clusters break apart and the field stays "a few
// far-off settlements in a big dark world." Pure + deterministic from (seed, progress);
// density is monotonic in progress. Extracted from the draw path so a test can count live
// cells across progress levels without rasterizing through the wash.
func settlementGrid(seed uint32, gw, gh int, progress float64) []bool {
	if gw <= 0 || gh <= 0 {
		return nil
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	// Sparse base density: a low floor plus a gentle progress ramp. Kept well under the
	// old 12–38% so early game reads as a handful of settlements, late game as a few dozen.
	seedDensity := 0.02 + 0.06*progress
	gseed := settlementSeed(seed)

	grid := make([]bool, gw*gh)
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			if hashUnit(uint32(gx), uint32(gy), gseed) < seedDensity {
				grid[gy*gw+gx] = true
			}
		}
	}
	// One light de-clumping pass: drop any live cell hemmed in by 3+ live neighbours so
	// coincidental clusters thin out to a scatter (the opposite of Conway growth). This
	// preserves the sparse, isolated look while still scaling count with progress.
	next := make([]bool, gw*gh)
	copy(next, grid)
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			if grid[gy*gw+gx] && settlementNeighbours(grid, gw, gh, gx, gy) >= 3 {
				next[gy*gw+gx] = false
			}
		}
	}
	return next
}

// settlementNeighbours counts the live Moore-neighbours of cell (gx,gy) in a gw×gh grid
// (out-of-range neighbours are dead — the canvas edge bounds the world).
func settlementNeighbours(grid []bool, gw, gh, gx, gy int) int {
	n := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := gx+dx, gy+dy
			if nx < 0 || ny < 0 || nx >= gw || ny >= gh {
				continue
			}
			if grid[ny*gw+nx] {
				n++
			}
		}
	}
	return n
}

// countLiveCells returns the number of live cells in a settlement grid — a pure measure
// of backdrop density used by the tests to show density scales with progress.
func countLiveCells(grid []bool) int {
	n := 0
	for _, c := range grid {
		if c {
			n++
		}
	}
	return n
}

// ---- foreground: your civ ---------------------------------------------------

// drawYourCiv paints the player's settlement as a single PROMINENT, ROUND dot in the bright
// accent role at a stable anchor (canvas center, snapped onto land). It is clearly the
// biggest, brightest mark on the map so your seat dominates the dim backdrop and the
// neighbour dots — but bounded: only ~2–3× a neighbour's diameter, not a canvas-eating blob.
//
// Roundness: a half-block sub-pixel is ~SQUARE on screen (one cell wide × half a cell tall),
// so an equal-radius disc (rx == ry, via the circular fillDot) reads round. The earlier
// aspect-corrected version used ry ≈ rx/2, which on this canvas overshot into a wide
// horizontal FOOTBALL — fixed by going round. There is deliberately NO smaller inner disc:
// a bright center disc drawn on top of a same-color body rasterizes a visible "+" cross at
// these radii, so we drop it and just brighten the whole body, giving a clean lit core with
// no cross artifact. A soft 1px glow ring (a dim accent halo one pixel proud of the body)
// lifts it off the field. Returns the placed dot so the overlay can label it. Theme-aware.
func drawYourCiv(img *image.RGBA, pal terrainPalette, wf *terrainField, w, h int) worldDot {
	cx, cy := w/2, h/2
	// Snap the anchor onto land so your capital never sits in the ocean. Center is the
	// stable spot; if it drowns, the nearest passable pixel is the smallest honest nudge.
	if nx, ny, ok := nearestPassablePx(wf, cx, cy); ok {
		cx, cy = nx, ny
	}
	accent := rgba(theme.Color(theme.RoleAccent))
	bg := rgba(theme.Color(theme.RoleBackground))

	// ROUND radius (rx == ry) scaling gently with canvas so your seat stays prominent on
	// big and small maps, clamped to a tight band: a neighbour dot is ~3–6px across, so a
	// radius of 3–5 (≈6–11px across) keeps you clearly the largest without ballooning into
	// a metropolis that swallows the map.
	r := clampInt((w+h)/40, 3, 5)

	// Glow ring: a dim accent halo one pixel proud of the body, so your capital reads as
	// lit against the dark field. Drawn first (underneath the body).
	glow := blend(bg, accent, 0.32)
	fillDot(img, cx, cy, r+1, glow)
	// Body: a single bright accent disc — no inner core disc (that rasterizes a "+"), just a
	// touch of extra brightness so the whole seat glows a clean, cross-free lit mark.
	fillDot(img, cx, cy, r, brighten(accent, 0.12))

	return worldDot{cx: cx, cy: cy, radius: r + 1, role: theme.RoleAccent}
}

// ---- foreground: diplomacy civs ---------------------------------------------

// worldCiv is a discovered civ resolved for the world view: its faction KEY (drives the
// stable per-faction color), a display name, a relationship role (drives the label + the
// standing ring), the at-war flag, and a size bucket (0..3) derived from strength.
type worldCiv struct {
	key        string
	name       string
	role       theme.Role
	atWar      bool
	sizeBucket int
}

// worldCivs resolves the diplomacy snapshot to a stable, sorted list of discovered civs
// with status colors + size buckets. Sorted by name so the ring is stable frame-to-frame
// (map iteration is otherwise random and would make the dots jump). Pure read of the
// snapshot — no locks, no engine calls. Undiscovered civs are omitted entirely.
func worldCivs(state game.GameState) []worldCiv {
	facs := state.Diplomacy.Factions
	if len(facs) == 0 {
		return nil
	}
	keys := make([]string, 0, len(facs))
	for k, f := range facs {
		if f.Discovered {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	out := make([]worldCiv, 0, len(keys))
	for _, k := range keys {
		f := facs[k]
		out = append(out, worldCiv{
			key:        k,
			name:       f.Name,
			role:       civWorldRole(f),
			atWar:      f.AtWar,
			sizeBucket: civStrengthBucket(f),
		})
	}
	return out
}

// factionColor returns a stable, distinct color for a discovered civ derived from its
// faction KEY — the same idea the city map's lineageColor uses for districts: hash the
// key to a hue and rotate a live theme role base by it, so every faction gets its own
// recognizable hue AND the whole set retints when the theme changes. RoleLabel is the
// base (a mid-chroma role that rotates cleanly across the wheel). This is the civ's
// IDENTITY color; relationship (ally/rival/at-war) is layered on top as a ring / hot
// override at draw time, so you can read both "who" and "how they feel" at a glance.
func factionColor(key string) color.RGBA {
	var hsh uint32 = 2166136261
	for i := 0; i < len(key); i++ {
		hsh ^= uint32(key[i])
		hsh *= 16777619
	}
	deg := float64(hsh % 360)
	return rotateHue(rgba(theme.Color(theme.RoleLabel)), deg)
}

// civStrengthBucket derives a 0..3 dot-size bucket from the civ's real Strength — the
// 1-5 power rating on the diplomacy roster (FactionDef), now projected onto
// game.FactionInfo. Strength maps 1→0, 2→1, 3→2, 4→2, 5→3; a live war makes a neighbor
// loom one size larger. Pure.
func civStrengthBucket(f game.FactionInfo) int {
	s := f.Strength
	if s <= 0 {
		s = 1 // guard for an unset/legacy snapshot
	}
	b := int(math.Round(float64(s-1) * 3.0 / 4.0))
	if f.AtWar {
		b++
	}
	return clampInt(b, 0, 3)
}

// civWorldRole maps a faction's standing to a theme role for the world dot. It mirrors
// the city map's civStatusRole (allied→Positive, rival/embargo→Negative, at-war→bright
// Negative handled at draw, friendly→Label, neutral→Dim) so the two views agree on what
// a relationship looks like. Reused directly to keep one source of truth.
func civWorldRole(f game.FactionInfo) theme.Role { return civStatusRole(f) }

// drawWorldCivs places each discovered civ as a dot ringed around your civ at a stable,
// deterministic angle + radius (seeded so the same neighbor always sits in the same
// place). Each dot carries TWO channels of meaning:
//
//	identity  — the body is the faction's own stable color (factionColor, hashed from the
//	            key so every civ is a distinct, theme-retinting hue you can tell apart).
//	standing  — layered on top: an ally gets a bright Positive ring, a (non-war) rival a
//	            Negative ring, and an AT-WAR civ overrides the body entirely with a hot
//	            bright red + a white-hot core so a threat pops instantly (the ⚔ mark rides
//	            on its label).
//
// Dots are drawn ROUND (rx == ry, via the circular fillDot — a half-block sub-pixel is
// ~square on screen, so an equal-radius disc reads round rather than as a cross/football),
// clearly LARGER + brighter than the now-sparse, dim backdrop so neighbours read as
// foreground actors, but a touch SMALLER than your own capital so your seat still dominates.
// Every dot is snapped onto land (nearestPassablePx) so no neighbour floats in the ocean.
// The label role stays the relationship role so the overlay colors names by standing.
// Returns the placed dots. Pure + theme-aware.
func drawWorldCivs(img *image.RGBA, pal terrainPalette, state game.GameState, wf *terrainField, seed uint32, yourX, yourY int) []worldDot {
	civs := worldCivs(state)
	if len(civs) == 0 {
		return nil
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	// Ring radius: keep neighbors well outside your civ's halo but inside the canvas, in
	// an elliptical band (the canvas is wider than tall in pixel terms only modestly, but
	// rows are half-height in cells, so we scale Y less to keep dots on-screen).
	baseRX := float64(w) * 0.34
	baseRY := float64(h) * 0.34

	hotRed := brighten(rgba(theme.Color(theme.RoleNegative)), 0.20) // at-war override body

	out := make([]worldDot, 0, len(civs))
	for i, c := range civs {
		// Deterministic angle: spread civs around the circle by index, jittered by a
		// stable per-civ hash so they don't sit on a perfect mechanical clock.
		frac := float64(i) / float64(len(civs))
		jitter := (hashUnit(uint32(i), 0xC17A, seed) - 0.5) * 0.20
		ang := frac*2*math.Pi + jitter*2*math.Pi
		// Radius jittered per civ so the ring has organic depth, clamped into the band.
		rj := 0.82 + 0.30*hashUnit(uint32(i), 0x5EED, seed)
		px := float64(yourX) + math.Cos(ang)*baseRX*rj
		py := float64(yourY) + math.Sin(ang)*baseRY*rj

		// ROUND dot: one radius (rx == ry). Sized from the strength bucket so a mightier
		// neighbour reads bigger, bumped so even the smallest clearly out-masses the 0–1px
		// backdrop, but capped at 3 so every neighbour stays a touch smaller than your own
		// capital (radius 3–5). r ∈ {2,2,3,3} for buckets 0..3.
		r := 2 + c.sizeBucket/2
		r = clampInt(r, 2, 3)
		// The ring is one pixel proud of the body, so inset by r+1 to keep the whole mark
		// on-canvas.
		inset := r + 1
		cx := clampInt(int(math.Round(px)), inset, w-1-inset)
		cy := clampInt(int(math.Round(py)), inset, h-1-inset)
		if w-1-inset < inset || h-1-inset < inset {
			// Canvas too small to inset; fall back to a clamped center-ish spot.
			cx = clampInt(int(math.Round(px)), 0, w-1)
			cy = clampInt(int(math.Round(py)), 0, h-1)
		}
		// Land snap: if the computed spot drowns, move to the nearest passable pixel so no
		// neighbour floats in the ocean. Done AFTER the inset clamp so the snap target is
		// already near an on-canvas spot; the nudge is the smallest honest move to shore.
		if nx, ny, ok := nearestPassablePx(wf, cx, cy); ok {
			cx, cy = nx, ny
		}

		ident := factionColor(c.key) // the civ's own identity hue
		standing := rgba(theme.Color(c.role))

		// Standing ring: a slightly larger disc of the relationship color under the body,
		// so an ally/rival reads its standing as a colored halo around its identity hue.
		// (Neutral/friendly rings are quiet — Dim/Label — so they don't shout.)
		ringCol := blend(rgba(theme.Color(theme.RoleBackground)), standing, 0.55)
		fillDot(img, cx, cy, r+1, ringCol)

		if c.atWar {
			// War overrides identity: a hot bright-red body — the loud threat read. The ring
			// underneath is already Negative (civStatusRole), so the whole dot glows red. No
			// inner core disc (it rasterizes a "+"); the body brightening carries the heat.
			fillDot(img, cx, cy, r, hotRed)
		} else {
			// Peace: the faction's own color, brightened a touch so it sits clearly above the
			// dim backdrop. Single disc — no inner core (cross artifact).
			fillDot(img, cx, cy, r, brighten(ident, 0.18))
		}

		out = append(out, worldDot{
			cx: cx, cy: cy, radius: r + 1,
			name: c.name, role: c.role, atWar: c.atWar,
		})
	}
	return out
}

// ---- dot primitive ----------------------------------------------------------

// fillDotAspect paints a filled ellipse with independent x/y pixel radii centered at
// (cx,cy), clipped to the image bounds. rx==ry==0 paints a single pixel. This is the
// aspect-correcting primitive the world map uses: the half-block canvas is 1px wide × 2px
// TALL per cell, so an equal-radius disc looks like a vertical "+"; passing rx≥ry (wider
// than tall) renders a mark that reads round-ish on screen. Used by the backdrop and the
// foreground civ dots so nothing renders as a fat cross.
func fillDotAspect(img *image.RGBA, cx, cy, rx, ry int, col color.RGBA) {
	b := img.Bounds()
	if rx <= 0 && ry <= 0 {
		if cx >= b.Min.X && cx < b.Max.X && cy >= b.Min.Y && cy < b.Max.Y {
			img.SetRGBA(cx, cy, col)
		}
		return
	}
	if rx < 0 {
		rx = 0
	}
	if ry < 0 {
		ry = 0
	}
	for dy := -ry; dy <= ry; dy++ {
		for dx := -rx; dx <= rx; dx++ {
			// Normalized ellipse test; guard the degenerate 0-radius axis (treat as a line).
			var ex, ey float64
			if rx > 0 {
				ex = float64(dx) / float64(rx)
			}
			if ry > 0 {
				ey = float64(dy) / float64(ry)
			}
			if ex*ex+ey*ey > 1.0 {
				continue
			}
			x, y := cx+dx, cy+dy
			if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
				continue
			}
			img.SetRGBA(x, y, col)
		}
	}
}

// fillDot paints a filled disc of radius r (pixels) centered at (cx,cy) into img with
// color col, clipped to the image bounds. r==0 paints a single pixel. Shared by the
// backdrop and both foreground passes so every dot uses one rasterizer.
func fillDot(img *image.RGBA, cx, cy, r int, col color.RGBA) {
	b := img.Bounds()
	if r <= 0 {
		if cx >= b.Min.X && cx < b.Max.X && cy >= b.Min.Y && cy < b.Max.Y {
			img.SetRGBA(cx, cy, col)
		}
		return
	}
	r2 := r * r
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy > r2 {
				continue
			}
			x, y := cx+dx, cy+dy
			if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
				continue
			}
			img.SetRGBA(x, y, col)
		}
	}
}

// ---- overlay ----------------------------------------------------------------

// buildWorldOverlay assembles the world map's text labels from the geometry: your civ's
// name under your dot, each discovered neighbor's name under its dot (at-war prefixed
// with a "⚔" mark), and the corner title. Pure (snapshot + geometry only). Labels store
// geometry + a theme ROLE; stampOverlay resolves colors live so a theme switch retints
// them — the same overlay contract CityMap uses.
func buildWorldOverlay(state game.GameState, cols, rows int, geo worldGeometry) overlayPlan {
	var plan overlayPlan
	if cols <= 0 || rows <= 0 {
		return plan
	}
	occupied := map[int]bool{}

	// 1) Your civ label, just beneath your dot, in the accent role (bold via labelTitle
	//    treatment? no — keep it a capital-kind so stampOverlay bolds it like a capital).
	plan.addYourCivLabel(state, geo, cols, rows, occupied)

	// 2) Neighbor labels, beneath each civ dot, colored by relationship, "⚔" if at war.
	plan.addWorldCivLabels(geo, cols, rows, occupied)

	// 3) Title last so it crowns its corner.
	plan.addWorldTitle(state, cols, rows)

	return plan
}

// addYourCivLabel stamps the player's civ name centered just below your dot. The name is
// AccountStats.DisplayName, falling back to "Your Empire" when accountless/unnamed.
// Accent role + capital kind so stampOverlay bolds it, marking it as the player's seat.
func (p *overlayPlan) addYourCivLabel(state game.GameState, geo worldGeometry, cols, rows int, occupied map[int]bool) {
	name := "Your Empire"
	if state.AccountStats != nil && state.AccountStats.DisplayName != "" {
		name = state.AccountStats.DisplayName
	}
	col := clampInt(pxToCellX(geo.you.cx), 0, cols-1)
	// One cell below the bottom edge of the dot (radius is in pixels → /2 for cells).
	row := clampInt(pxToCellY(geo.you.cy)+geo.you.radius/2+1, 0, rows-1)
	occupied[packCell(0, row)] = true
	text := truncLabel(name, maxLabelLen(col, cols, alignCenter))
	if text == "" {
		return
	}
	p.labels = append(p.labels, overlayLabel{
		cx: col, cy: row, text: text, role: theme.RoleAccent, kind: labelCapital, align: alignCenter,
	})
}

// addWorldCivLabels names each discovered neighbor beneath its dot, colored by its
// relationship role (resolved live at draw), with a "⚔" prefix when at war (and the
// bright treatment so the warring neighbor's label pops). Collision-limited per cell so
// two close dots don't overprint each other's names. Labels nudge a row if taken.
func (p *overlayPlan) addWorldCivLabels(geo worldGeometry, cols, rows int, occupied map[int]bool) {
	for _, d := range geo.civs {
		if d.name == "" {
			continue
		}
		col := clampInt(pxToCellX(d.cx), 0, cols-1)
		baseRow := pxToCellY(d.cy) + d.radius/2 + 1
		// Try the row under the dot, then a small spread so a crowded cluster keeps more
		// of its labels instead of dropping them.
		row := -1
		for _, dr := range []int{0, 1, -1, 2, -2} {
			cand := clampInt(baseRow+dr, 0, rows-1)
			if !occupied[packCell(0, cand)] {
				row = cand
				break
			}
		}
		if row < 0 {
			continue
		}
		occupied[packCell(0, row)] = true

		text := d.name
		if d.atWar {
			text = "⚔ " + text
		}
		text = truncLabel(text, maxLabelLen(col, cols, alignCenter))
		if text == "" {
			continue
		}
		p.labels = append(p.labels, overlayLabel{
			cx: col, cy: row, text: text, role: d.role, kind: labelCiv, bright: d.atWar, align: alignCenter,
		})
	}
}

// addWorldTitle stamps the corner title "Known World — <Age>" at the top-left in the
// accent role, drawn last so it crowns its corner. Uses the live age name (falling back
// to the age key). Mirrors the city map's title pattern.
func (p *overlayPlan) addWorldTitle(state game.GameState, cols, rows int) {
	age := state.AgeName
	if age == "" {
		age = state.Age
	}
	title := "Known World"
	if age != "" {
		title = "Known World — " + age
	}
	title = truncLabel(title, cols-1)
	if title == "" {
		return
	}
	p.labels = append(p.labels, overlayLabel{
		cx: 1, cy: 0, text: title, role: theme.RoleAccent, kind: labelTitle, align: alignLeft,
	})
}
