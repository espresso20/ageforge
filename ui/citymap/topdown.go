package citymap

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// topdown.go is the citymap v3 engine (LOCKED spec:
// design-and-architecture/city-synthesis.md). It replaces the terrain-gated,
// isometric citymap with a TOP-DOWN pixel-art living city: you look straight down
// at roofs, dirt lanes, gardens and squares, and the fabric grows and re-skins as
// the civ grows. This file owns the whole V3-A path — era-tinted ground, stable
// persistent placement (a WONDER-ANCHORED, lane-grown, type-INTERMIXED fabric — see
// city-synthesis.md §Revisions (playtest)), fill-frame scaling, the top-down roof atlas
// with SE drop-shadows, organic streets between the anchors, balanced living-city
// filler, the walls capability (off in primitive), the wonder anchors with their clear
// plazas, and landmark-only labels.
//
// SCOPE: this is the citymap render path ONLY. The worldmap (worldmap.go +
// terrain.go's world funcs) is approved and untouched — the citymap simply stops
// calling terrain, the isometric drawVolume, and land-gating. There is deliberately
// no terrainField, no biome, no water, no pathfinding on this path: the ground is a
// neutral era tint and every green thing is BUILT.
//
// Determinism: everything below is pure and seeded from citySeed (a stable hash of
// the civ display name, AGE-INDEPENDENT). Placement is a stable SEQUENCE — instance
// #N keeps slot N as #N+1 is added — so the bones never re-randomize; only the
// fill-frame scale re-fits as the city densifies. Panic-safe on any canvas.

// ---- era style presets (top-down) ------------------------------------------

// tdStreetPattern names the street topology an era grows. V3-A tunes only
// tdOrganic (winding dirt lanes); grid/avenue slot in for V3-B+.
type tdStreetPattern int

const (
	tdOrganic tdStreetPattern = iota // meandering dirt lanes cluster→core (village)
	tdGrid                           // orthogonal grid (V3-B+)
	tdAvenue                         // wide boulevards + superblocks (V3-C+)
)

// tdEraStyle is the top-down era preset: the ground tone recipe, street topology +
// width, roof material recipe, the walls capability flag, and the filler density.
// Colors are expressed as THEME-ROLE RECIPES (resolved at render time) so the whole
// city retints on a theme switch — no color is hard-coded.
//
// A recipe is a small closure over a resolved palette that returns a color, so the
// preset stays theme-agnostic while still choosing an era MOOD (earthy vs stone vs
// neon). tdResolveStyle turns the preset into concrete colors for one frame.
type tdEraStyle struct {
	name string

	streetPattern tdStreetPattern
	laneWidth     int     // half-thickness of a lane in px (0 → 1px, 1 → 3px band)
	streetJitter  float64 // lane meander amount (0 straight … 1 very wavy) — organic only

	// Roof material: base + shaded-slope recipes, blended per building with a subtle
	// lineage tint so a temple reads different from a hut without leaving the era mood.
	// The ridge/crown highlight is NOT a preset recipe — roofColorsFor derives it as a
	// fixed lighten of base so it can never resolve to a saturated accent (the yellow-
	// dot regression). Only the base material + the shaded slope are era recipes.
	roofBase   func(tdPal) color.RGBA // dominant roof fill (e.g. thatch brown)
	roofDark   func(tdPal) color.RGBA // shaded slope
	lineageMix float64                // how much lineage tint bleeds into the roof (~0.15–0.25)

	// Ground: base fill + a slightly-varied texture tone, both era-tinted.
	groundBase func(tdPal) color.RGBA
	groundAlt  func(tdPal) color.RGBA
	// Street surface color.
	streetCol func(tdPal) color.RGBA

	// Living-city filler accents.
	gardenCol func(tdPal) color.RGBA
	squareCol func(tdPal) color.RGBA
	treeCol   func(tdPal) color.RGBA
	propCol   func(tdPal) color.RGBA

	// Walls capability (locked #9). Primitive sets false; walls arrive in V3-B.
	hasWalls bool
	wallCol  func(tdPal) color.RGBA

	fillerDensity float64 // balanced living-city filler amount (locked #12)
}

// tdPal is the small set of resolved theme colors the style recipes draw from. Built
// once per frame from the active theme, so a theme switch re-resolves every recipe.
type tdPal struct {
	bg        color.RGBA // RoleBackground
	dim       color.RGBA // RoleDim
	text      color.RGBA // RoleText
	accent    color.RGBA // RoleAccent
	highlight color.RGBA // RoleHighlight
	positive  color.RGBA // RolePositive
	label     color.RGBA // RoleLabel
	shadow    color.RGBA // a soft dark for drop-shadows
}

// newTdPal resolves the theme roles the era recipes need. Pure read of the active
// theme; no locks. Kept separate from the legacy terrainPalette so the top-down
// path never pulls a biome/water tone.
func newTdPal() tdPal {
	bg := rgba(theme.Color(theme.RoleBackground))
	dim := rgba(theme.Color(theme.RoleDim))
	return tdPal{
		bg:        bg,
		dim:       dim,
		text:      rgba(theme.Color(theme.RoleText)),
		accent:    rgba(theme.Color(theme.RoleAccent)),
		highlight: rgba(theme.Color(theme.RoleHighlight)),
		positive:  rgba(theme.Color(theme.RolePositive)),
		label:     rgba(theme.Color(theme.RoleLabel)),
		// Shadow: background pushed dark, so it grounds a roof without a hard black.
		shadow: blend(dim, color.RGBA{A: 0xff}, 0.45),
	}
}

// organicVillageStyle is the fully-tuned PRIMITIVE/STONE preset (locked era table
// row "organic"): earthy dirt+grass ground, winding thin dirt lanes, thatch/wood
// roofs, NO walls, a lived-in but not cluttered filler density. This is the V3-A
// showcase era.
var organicVillageStyle = tdEraStyle{
	name:          "organic",
	streetPattern: tdOrganic,
	laneWidth:     0, // dirt paths are thin
	streetJitter:  0.9,

	// Thatch/wood browns, anchored to a warm earthen hue but pulled from theme roles
	// (RoleText carries the warm neutral in these themes; RoleDim grounds it) so a
	// dark or light theme still yields a legible, in-family roof.
	roofBase: func(p tdPal) color.RGBA {
		// warm brown thatch: background lifted toward text, warmed toward an earth anchor.
		return blend(blend(p.bg, p.text, 0.34), earthAnchor, 0.42)
	},
	roofDark: func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.text, 0.34), earthAnchor, 0.42), 0.30)
	},
	lineageMix: 0.18,

	// Ground: earthy dirt with a faint grass cast; alt tone a touch greener for texture.
	groundBase: func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.28), dirtAnchor, 0.34)
	},
	groundAlt: func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.22), grassAnchor, 0.18)
	},
	streetCol: func(p tdPal) color.RGBA {
		// packed dirt lane: dirt anchor darkened a touch so it reads as trodden earth.
		return darken(blend(blend(p.bg, p.dim, 0.30), dirtAnchor, 0.50), 0.10)
	},

	gardenCol: func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.20), grassAnchor, 0.42)
	},
	squareCol: func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.34), dirtAnchor, 0.30)
	},
	treeCol: func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.dim, 0.30), grassAnchor, 0.55), 0.12)
	},
	propCol: func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.20), earthAnchor, 0.35)
	},

	hasWalls: false, // primitive: no walls (locked — walls arrive in V3-B)
	wallCol: func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.40), dirtAnchor, 0.35)
	},

	fillerDensity: 1.0,
}

// defaultTdStyle is the reasonable fallback for every non-primitive era until
// V3-B/C/D tunes them (locked Phasing: "Other eras use a reasonable default
// preset"). It reuses the organic recipes for a legible, theme-tinted city with a
// simple ridged-rect default roof — NOT its final silhouette. Kept as a copy so
// tuning organic later can't accidentally restyle every era.
var defaultTdStyle = func() tdEraStyle {
	s := organicVillageStyle
	s.name = "default"
	s.streetPattern = tdOrganic
	return s
}()

// Muted hue anchors for the earthy village mood. Blended at modest strength against
// theme roles so a light or dark theme still gets an in-family, non-cartoon palette.
var (
	earthAnchor = color.RGBA{R: 0x8a, G: 0x63, B: 0x3a, A: 0xff} // thatch/wood brown
	dirtAnchor  = color.RGBA{R: 0x7c, G: 0x66, B: 0x46, A: 0xff} // packed dirt
	grassAnchor = color.RGBA{R: 0x4f, G: 0x6f, B: 0x33, A: 0xff} // built greenery
)

// tdStyleForEra returns the tuned preset for an era band, or defaultTdStyle for the
// bands V3-A leaves on the fallback. Organic (primitive, stone) is the only tuned
// band in V3-A; every other band renders a legible default city.
func tdStyleForEra(e era) tdEraStyle {
	switch e {
	case eraOrganic:
		return organicVillageStyle
	default:
		return defaultTdStyle
	}
}

// ---- roof atlas -------------------------------------------------------------

// roofType is a top-down roof archetype. Each renders as a filled roof shape that
// FILLS its lot, with a ridge/texture highlight and a soft SE drop-shadow (locked
// #5,#6). Material comes from the era style; a subtle lineage tint differentiates
// types within the era mood.
type roofType int

const (
	roofHut      roofType = iota // small round/oval thatch roof + radial streaks
	roofRidge                    // rectangle pitched roof: center ridge + two shaded slopes
	roofLong                     // elongated ridge roof (longhouse / rowhouse)
	roofTemple                   // larger ornate symmetric roof + finial (shrine/temple)
	roofCamp                     // lean-to / open-frame / tent (gathering/forager camp)
	roofStash                    // small square store hut (storage)
	roofFlat                     // low flat structure (stone camp / workshop)
	roofWonder                   // large multi-part ornate complex (centerpiece)
)

// getRoofType maps a building's domain/category/tier to a roof archetype. Primitive
// archetypes are tuned precisely (per city-synthesis.md §roof atlas): huts round,
// longhouses elongated, shrines/temples ornate, camps as tents, stashes as store
// huts, stone works flat. Other eras fall through to roofRidge (the simple default
// roof) until V3-B/C specializes them. Pure data — safe on the render path.
func getRoofType(domain, category string, tier int) roofType {
	switch category {
	case "wonder", "monument":
		return roofWonder
	case "housing":
		// tier 0 hut → round thatch; tier 1+ longhouse → elongated ridge.
		if tier <= 0 {
			return roofHut
		}
		return roofLong
	case "storage":
		return roofStash
	case "research", "diplomacy":
		// faith/knowledge/culture civic buildings read as shrines/temples.
		return roofTemple
	}
	switch domain {
	case "faith", "knowledge", "culture_arts":
		return roofTemple
	case "food", "military":
		// camps: open-frame tents / lean-tos.
		return roofCamp
	case "geological_extraction":
		// stone works: low flat structures.
		return roofFlat
	case "organic_extraction":
		// wood camps: open-frame like the food camps but a touch sturdier — reuse camp.
		return roofCamp
	}
	// Default (other production, other eras): a simple ridged rectangle.
	return roofRidge
}

// roofColors bundles the three material tones a roof draws with, already blended
// with the building's subtle lineage tint (~15–25%) so types differ within the era.
type roofColors struct {
	base, ridge, dark color.RGBA
}

// roofColorsFor resolves the era material for a lot's domain/category, then bleeds
// the lineage tint into it by the style's lineageMix. So every roof stays in the era
// mood (thatch brown) while a temple leans faith-violet, a forge leans forge-orange,
// etc. — the "subtle lineage tint" of locked #6.
//
// CRITICAL (yellow-dot regression guard): the ridge/highlight tone is derived from
// the roof BASE (a lighter shade of the same material), NEVER from an accent/highlight
// theme role. An earlier cut painted the ridge from style.roofRidge, whose recipe in
// some themes pulled RoleAccent/RoleHighlight (bright yellow) and stamped it at each
// roof's crown → the "bright yellow center-dot" playtest bug. Ridge is now a fixed
// lighten of base so a roof crown can only ever read as a lit patch of its own thatch.
// dark stays the era shaded-slope recipe (a darken of the same family, still safe).
func roofColorsFor(style tdEraStyle, pal tdPal, domain, category string) roofColors {
	base := style.roofBase(pal)
	dark := style.roofDark(pal)
	if style.lineageMix > 0 {
		// Desaturate the lineage color before bleeding it in. The raw role colors for
		// faith (violet), wonders (gold), culture (magenta) are highly saturated; blended
		// straight they poster-paint a temple's ridge into a neon cross. Pulling the tint
		// most of the way to a same-lightness gray keeps the HUE cue ("a temple leans
		// violet") while staying an earthy, muted thatch — the "subtle lineage tint" of
		// locked #6.
		tint := mutedTint(lineageColor(domain, category))
		base = clampRoofSat(blend(base, tint, style.lineageMix))
		dark = clampRoofSat(blend(dark, tint, style.lineageMix))
	}
	// Ridge = a lighter shade of the (already lineage-tinted, saturation-capped) base.
	// Base-derived by construction, so it can never be a saturated accent no matter the
	// theme/recipe, and the cap keeps even a civic roof's crown earthy, not neon.
	ridge := clampRoofSat(brighten(base, 0.14))
	return roofColors{base: base, ridge: ridge, dark: dark}
}

// mutedTint desaturates a lineage color toward a gray of the same lightness, so the
// subtle roof tint nudges hue without going neon. Keeps ~35% of the original chroma.
func mutedTint(c color.RGBA) color.RGBA {
	h, s, l := rgbToHSL(c)
	return hslToRGB(h, s*0.35, l)
}

// clampRoofSat caps a roof tone's saturation so no roof — even a violet-faith temple or
// a gold-wonder — can read as a saturated poster-paint token. A hard ceiling on chroma
// guarantees the whole atlas stays in the earthy thatch family (the locked "muted &
// natural saturation" of #6), which is also what keeps the ridge safely off any accent.
func clampRoofSat(c color.RGBA) color.RGBA {
	const maxSat = 0.35
	h, s, l := rgbToHSL(c)
	if s > maxSat {
		s = maxSat
	}
	return hslToRGB(h, s, l)
}

// ---- city-space plan model --------------------------------------------------

// tdLotKind classifies a placed top-down parcel.
type tdLotKind int

const (
	tdRoof   tdLotKind = iota // a building roof (the count-scaled fabric + landmarks)
	tdGarden                  // a built green plot
	tdSquare                  // a paved open square
	tdTree                    // a tree / dot cluster
	tdProp                    // a well / stall / statue prop
	tdWall                    // a wall segment
	tdGate                    // a gate in the wall ring
)

// tdLot is one placed thing, in CITY SPACE (pre-fill-frame). x,y is the lot center
// in city units; w,h its extent. Roof lots carry the building identity so the roof
// atlas + landmark labels can read them. Because placement is a stable sequence,
// a lot's city-space position never moves when later lots are added.
type tdLot struct {
	x, y float64 // center, city space
	w, h float64 // extent, city space
	kind tdLotKind

	// Building identity (roof lots only).
	domain   string
	category string
	tier     int
	roof     roofType
	label    string  // set for labeled landmarks (city center hero, wonder, promoted hero)
	prom     float64 // prominence (for landmark z-order + label priority)
}

// tdAnchor is one GROWTH ANCHOR the settlement grows around (playtest revision, FIX
// 2). Anchors are the built WONDERS (each wonder seats one anchor); a civ with zero
// wonders has a single anchor at the city center. The lanes wind between the anchors
// and the intermixed roof fabric grows around each. `wonder` flags an anchor that
// seats a wonder (gets a clear plaza + the centerpiece roof); the bare city-center
// anchor of a wonderless village does NOT clear a plaza (a tiny village must not be
// hollowed out). Positions are a stable function of (index, seed) — age-independent —
// so the anchor field never rearranges across renders/ages.
type tdAnchor struct {
	cx, cy float64
	wonder bool          // seats a wonder → plaza-cleared centerpiece
	bld    builtBuilding // the wonder building this anchor seats (wonder anchors only)
}

// tdStreet is one lane polyline in city space, with a width class.
type tdStreet struct {
	pts   []tdPoint
	width int
}

// tdPoint is a float city-space coordinate.
type tdPoint struct{ x, y float64 }

// topPlan is the render-ready top-down plan in CITY SPACE. renderTopDown computes
// the fill-frame transform from its lot bounding box, then paints everything mapped
// into canvas pixels. Keeping the plan in city space is what makes fill-frame work:
// the same relative layout re-fits to any canvas / density.
type topPlan struct {
	streets []tdStreet
	// anchors are the wonder-anchored growth seats (FIX 2): the lanes wind between
	// them and the intermixed fabric grows around each. A wonderless village has a
	// single city-center anchor. Kept on the plan so the street generator + tests can
	// read the growth skeleton.
	anchors []tdAnchor
	lots    []tdLot
	// center is the civic heart in city space (the City Center anchor + wonder seat).
	cx, cy float64
	// heroLabel is the promoted hero's identity when the civ has no civic building
	// (locked #7). Empty otherwise.
	heroLabel tdLot
	hasHero   bool
}

// ---- civ seed (stable, age-independent) -------------------------------------

// citySeed derives a STABLE-per-civ seed from the display name, the SAME way
// worldTerrainSeed does (FNV-1a over the name), but with a DISTINCT salt so the city
// plan and the world land don't share a lattice. Crucially it is AGE-INDEPENDENT
// (locked #8): the bones must not move across ages — only the era re-skin changes.
// An empty/anonymous name falls back to a fixed non-zero seed so the anonymous city
// is still stable.
func citySeed(displayName string) uint32 {
	if displayName == "" {
		return 0x1d3a5c07 // stable anonymous city
	}
	var h uint32 = 2166136261
	for i := 0; i < len(displayName); i++ {
		h ^= uint32(displayName[i])
		h *= 16777619
	}
	h ^= 0x2f6e10bd // salt: keep city seeds off the world/age seed lattices
	return h | 1
}

// displayNameOf pulls the civ display name the citySeed hashes, matching the
// worldmap's rule (AccountStats.DisplayName, else ""). Kept local so the two seeds
// hash exactly the same input.
func displayNameOf(state game.GameState) string {
	if state.AccountStats != nil {
		return state.AccountStats.DisplayName
	}
	return ""
}

// ---- growth anchors (wonder-anchored, playtest FIX 2) -----------------------
//
// The settlement grows around ANCHORS, not hard per-domain districts. Anchors are the
// built WONDERS (each seats one), spread as a stable, seeded set across the settlement
// area; a wonderless village has a single city-center anchor. More wonders → more
// anchors, wider spread. See tdAnchorPoints.

// tdAnchorPoints places nWonders growth-anchor centers in city space, spread stably
// and deterministically around the core so more wonders push the town wider (FIX 2).
// nWonders==0 yields exactly one anchor AT the core (a cohesive wonderless village).
// For N≥1 the anchors sit on a golden-angle phyllotaxis disc whose radius grows with
// sqrt of the anchor index and an overall spread that grows with sqrt(N): a few
// wonders sit close, many fan out across the whole map. Anchor i's offset is a pure
// function of (i, seed) — NOT of N — so adding a wonder never moves the anchors placed
// before it (the same stable-incremental discipline the fabric uses).
func tdAnchorPoints(cx, cy float64, nWonders int, seed uint32, cfg tdConfig) []tdPoint {
	if nWonders <= 0 {
		return []tdPoint{{cx, cy}}
	}
	// A stable per-seed phase so two civs' anchor rings differ, but a given civ's is
	// fixed across ages (seed is age-independent).
	phase := float64(hash2(0xA9C, 0x37, seed)) / float64(^uint32(0)) * 2 * math.Pi
	// Spread grows with the count so a metropolis of wonders occupies more ground while
	// a lone wonder stays central. sqrt keeps it sub-linear (fill-frame re-fits anyway).
	spread := cfg.anchorSpread * math.Sqrt(float64(nWonders))
	out := make([]tdPoint, 0, nWonders)
	for i := 0; i < nWonders; i++ {
		if i == 0 {
			// The first (grandest) wonder crowns the center.
			out = append(out, tdPoint{cx, cy})
			continue
		}
		// Phyllotaxis: radius ∝ sqrt(i), angle = i*goldenAngle + phase — a low-overlap,
		// index-stable spread. Normalize the radius by sqrt(nWonders) so the outermost
		// anchor lands ~spread from the core regardless of count.
		r := spread * math.Sqrt(float64(i)) / math.Sqrt(float64(maxInt(nWonders-1, 1)))
		a := float64(i)*goldenAngle + phase
		out = append(out, tdPoint{cx + math.Cos(a)*r, cy + math.Sin(a)*r})
	}
	return out
}

// ---- generate (pure, deterministic, stable-incremental) ---------------------

// goldenAngle is the golden angle in radians (~137.5°). Placing slot i at angle
// i*goldenAngle with radius ∝ sqrt(i) fills a disk with a stable, low-overlap
// phyllotaxis spiral: every slot has a FIXED index, so adding a later slot never moves
// an earlier one (locked #8, the anti-re-randomize guarantee). The intermixed lane
// placement (FIX 1) grows each anchor's fabric on this spiral — interleaving the
// per-type queues into it so consecutive slots are DIFFERENT domains — then pulls each
// slot toward the nearest lane so the fabric grows along the streets.
const goldenAngle = 2.399963229728653 // math.Pi * (3 - sqrt(5))

// slotJitter returns a small ORGANIC offset for slot i within anchor group di, breaking
// up the visible golden-angle diamond-lattice so the fabric reads natural rather than
// crystalline. CRITICAL: the offset is a PURE FUNCTION of (i, di, seed) via a hash —
// it does NOT draw from the threaded rng — so slot i's jitter is identical whether it
// is the last building placed at count N or an interior one at count N+1. That is what
// keeps placement stable-incremental (locked #8) even with jitter: adding a building
// can never move an existing one. amp is the max wander in city units (~a fraction of
// slot spacing) so buildings still pack close, just off the perfect lattice.
func slotJitter(i, di int, seed uint32, amp float64) (dx, dy float64) {
	if amp <= 0 {
		return 0, 0
	}
	h1 := hash2(uint32(i)*2+1, uint32(di)*131+7, seed^0x9e3779b9)
	h2 := hash2(uint32(i)*2+2, uint32(di)*131+11, seed^0x85ebca6b)
	// Map each hash to [-1,1).
	jx := float64(h1)/float64(^uint32(0))*2 - 1
	jy := float64(h2)/float64(^uint32(0))*2 - 1
	return jx * amp, jy * amp
}

// tdConfig holds the fixed generator constants (city-space units). City space is an
// abstract plane; the fill-frame transform later maps the plan's bounding box onto
// the canvas, so these are relative sizes, not pixels.
type tdConfig struct {
	anchorSpread float64 // how far the outermost wonder-anchor sits from the core (FIX 2)
	slotSpacing  float64 // golden-spiral spacing between instances (tight fabric)
	roofSize     float64 // base roof extent in city units
	jitterAmp    float64 // organic per-slot wander (city units) breaking the lattice

	// laneBias is how hard each roof slot is pulled toward its nearest lane (0 = free
	// spiral, 1 = snapped onto the lane) so the intermixed fabric grows ALONG the
	// streets and the town OUTLINE follows the lanes rather than reading as a disc.
	laneBias float64
	// plazaRadius is the clear ground kept immediately around a WONDER anchor (in units
	// of roofSize) so the city fabric never buries the centerpiece (the playtest
	// complaint). Non-wonder roof lots landing inside a wonder's plaza are dropped.
	plazaRadius float64
}

var defaultTdConfig = tdConfig{
	anchorSpread: 20,
	slotSpacing:  2.4,
	roofSize:     3.2,
	jitterAmp:    0.8,
	laneBias:     0.35,
	plazaRadius:  2.2,
}

// generateTopPlan synthesizes the whole top-down city plan in CITY SPACE, purely and
// deterministically from seed. Pipeline (city-synthesis.md §Pipeline + §Revisions
// (playtest), top-down):
//
//	(a) gather   — built buildings → per-type domain/category/tier/count/role.
//	(b) anchors  — the built WONDERS become the growth anchors (FIX 2); a wonderless
//	    village has a single city-center anchor. More wonders → more, wider anchors.
//	(c) lanes    — winding lanes wind between/around the anchors (laid FIRST so the
//	    fabric can grow along them and the town outline follows the streets).
//	(d) populate — ALL buildings placed in one stable, type-INTERMIXED sequence of
//	    slots along the lane network around the nearest anchor (FIX 1): consecutive
//	    slots are different domains (no round same-type blob). Each wonder sits AT its
//	    anchor with a clear plaza. Stable-incremental: slot i is a pure fn of (i, seed).
//	(e) filler   — balanced gardens / squares / trees / props in the gaps.
//	(f) walls    — a wall+gate ring IF the era has walls (primitive: none).
//
// No terrain, no water, no gating: the ground is neutral era tint. w,h are only used
// to bound wall geometry loosely; the fill-frame transform is computed at render.
func generateTopPlan(state game.GameState, byKey map[string]config.BuildingDef, style tdEraStyle, seed uint32) topPlan {
	cfg := defaultTdConfig
	plan := topPlan{cx: 0, cy: 0}
	r := newRNG(seed)

	// (a) gather — reuse the pure, sorted gather from citygen.go (domain/category/
	// tier/count/role per distinct built type). Sorted, so placement order is stable.
	blds := gatherBuildings(state, byKey)

	// Split the gathered buildings into the WONDERS (the growth anchors + centerpieces)
	// and the rest (the intermixed fabric). Wonders are sorted by prominence so the
	// grandest seats the center anchor.
	var wonders []builtBuilding
	for _, b := range blds {
		if b.category == "wonder" || b.category == "monument" {
			wonders = append(wonders, b)
		}
	}
	sort.SliceStable(wonders, func(i, j int) bool { return moreProminentBld(wonders[i], wonders[j]) })

	// (b) anchors — seat one growth anchor per built wonder (FIX 2), spread stably around
	// the core so more wonders push the town wider. Zero wonders → a single city-center
	// anchor (a cohesive wonderless village). The grandest wonder crowns anchor 0 (the
	// center); the rest fan out on the seeded phyllotaxis disc.
	anchorPts := tdAnchorPoints(plan.cx, plan.cy, len(wonders), seed, cfg)
	plan.anchors = make([]tdAnchor, len(anchorPts))
	for i, p := range anchorPts {
		a := tdAnchor{cx: p.x, cy: p.y}
		if i < len(wonders) {
			a.wonder = true
			a.bld = wonders[i]
		}
		plan.anchors[i] = a
	}

	// Hero promotion (locked #7): the single most prominent LANDMARK is the labeled
	// civic hero. With no civic building AND no wonder, promote the most prominent
	// PRODUCTION building so the city still has exactly one labeled landmark.
	var landmarks, production []builtBuilding
	for _, b := range blds {
		if b.category == "wonder" || b.category == "monument" {
			continue
		}
		if b.role == roleLandmark {
			landmarks = append(landmarks, b)
		} else if b.role == roleProduction {
			production = append(production, b)
		}
	}
	heroKey := ""
	if len(landmarks) > 0 {
		best := 0
		for i := 1; i < len(landmarks); i++ {
			if moreProminentBld(landmarks[i], landmarks[best]) {
				best = i
			}
		}
		heroKey = landmarks[best].key
	} else if len(wonders) == 0 && len(production) > 0 {
		best := 0
		for i := 1; i < len(production); i++ {
			if moreProminentBld(production[i], production[best]) {
				best = i
			}
		}
		plan.heroLabel = tdLot{
			domain: production[best].domain, category: production[best].category,
			tier: production[best].tier, label: production[best].name,
		}
		plan.hasHero = true
		heroKey = production[best].key
	}

	// (c) lanes — laid FIRST so the fabric grows along them and the outline follows the
	// streets (FIX 1). Winding lanes wind between the anchors + a couple of cross paths.
	plan.streets = tdOrganicStreets(plan.anchors, plan.cx, plan.cy, style, r)

	// (d) populate — the INTERMIXED lane-grown fabric (FIX 1). Every non-wonder building
	// emits count-scaled roof lots, and the per-type queues are INTERLEAVED into one
	// stable sequence per anchor so consecutive slots are different domains (a hut next
	// to a camp next to a store, not one big blob of huts). Each type is assigned to an
	// anchor (round-robin, deterministic) and grows a golden-angle spiral around it,
	// pulled toward the nearest lane. Stability (locked #8): a type's j-th lot position
	// is a pure function of (typePhase, j, seed) — NOT of any other type's count or of a
	// shared global cursor — so adding a building never moves an existing one, yet the
	// per-type phases stagger the shared spiral so different types interleave along it.
	tdPopulateIntermixed(&plan, blds, heroKey, cfg, seed)

	// Wonder centerpieces (FIX 2): each wonder sits AT its anchor as a dominant, ornate
	// complex with a CLEAR PLAZA around it (any fabric lot inside the plaza was dropped
	// in tdPopulateIntermixed), so the city hugs the wonder without ever burying it.
	tdPlaceWonders(&plan, cfg)

	// (e) filler — balanced gardens / squares / trees / props in the gaps.
	tdAddFiller(&plan, style, cfg, r)

	// (f) walls — a wall+gate ring IF the era has walls. Primitive: none. The ring is
	// sized to the lot bounding box so it hugs the built-up area.
	if style.hasWalls {
		tdAddWalls(&plan, r)
	}

	return plan
}

// ---- intermixed lane-grown placement (FIX 1) --------------------------------

// tdLotSpec is one pending fabric roof — the building identity for a single instance,
// plus its stable anchor binding, before it is emitted. Built per-type, then ROUND-ROBIN
// emitted across types into one intermixed slice (FIX 1). di + phase01 make the lot's
// position a pure function of (type, j) so it never moves when a sibling grows.
type tdLotSpec struct {
	b       builtBuilding
	j       int // this building's instance index (0-based)
	roof    roofType
	sz      float64
	di      int     // the anchor this type is bound to (typeIdx % numAnchors)
	phase01 float64 // stable per-type phase into the anchor's spiral (hash of the key)
	label   bool    // the civic hero's headline instance
}

// tdPopulateIntermixed lays the whole non-wonder fabric as a stable, type-INTERMIXED,
// lane-grown settlement (FIX 1). It (1) builds a per-type queue of instance specs — each
// type bound to ONE anchor (typeIdx % A) with a STABLE per-type phase (hash of its key)
// into that anchor's golden-angle spiral — then (2) ROUND-ROBIN emits the queues (one
// instance per type per round, in the stable sorted-gather order) so CONSECUTIVE slots
// in plan.lots are different domains (a hut next to a camp next to a store, not a blob of
// huts). A type's j-th instance takes spiral index round(stride*(phase+j)); since the
// phases differ per type, different types INTERLEAVE around the shared anchor spiral, so
// the fabric is intermixed spatially too. Each slot is pulled toward its nearest lane so
// the fabric grows ALONG the streets and the outline follows them. Lots inside a WONDER
// anchor's clear plaza are dropped so the centerpiece is never buried (FIX 2).
//
// Stability (locked #8): a lot's anchor and spiral index — hence its position and jitter
// — are a pure function of (building type, instance index, seed), NEVER of another type's
// count, the emit order, or the threaded rng. So growing any building's count only
// APPENDS its new instances and can never move an existing lot. (The round-robin emit
// only interleaves the slice; it does not feed into any position.)
func tdPopulateIntermixed(plan *topPlan, blds []builtBuilding, heroKey string, cfg tdConfig, seed uint32) {
	if len(plan.anchors) == 0 {
		return
	}
	plazaR := cfg.plazaRadius * cfg.roofSize
	nAnchors := len(plan.anchors)
	// spiralStride > 1 spaces one type's own instances apart in its anchor's spiral so
	// OTHER types' phased slots fall between them — the spatial interleave. A small
	// integer near the number of types sharing an anchor, so a run of distinct types
	// fills roughly one lap before a type places its next instance.
	const spiralStride = 3.0

	// (1) Per-type queues of instance specs, in the stable sorted-gather order. Each type
	// is bound to ONE anchor (typeIdx % A) and given a STABLE fractional phase (hash of
	// its key) into that anchor's spiral index space. CRUCIAL for stability (locked #8):
	// a lot's (anchor, spiral-index) is a pure function of (type, j) — never of any other
	// type's count or a shared global cursor — so a sibling growing can never move it. The
	// per-type phase differs by type, so different types INTERLEAVE around the anchor.
	queues := make([][]tdLotSpec, 0, len(blds))
	typeIdx := 0
	for _, b := range blds {
		if b.category == "wonder" || b.category == "monument" {
			continue // wonders are anchors/centerpieces, not fabric
		}
		n := tdRoofCount(b.count, b.role)
		if n <= 0 {
			continue
		}
		rt := getRoofType(b.domain, b.category, b.tier)
		sz := cfg.roofSize
		if rt == roofLong {
			sz *= 1.15
		}
		di := typeIdx % nAnchors
		typeIdx++
		phase01 := float64(hash2(fnvKey(b.key), 0x5bd1e995, seed)) / float64(^uint32(0))
		q := make([]tdLotSpec, n)
		for j := 0; j < n; j++ {
			q[j] = tdLotSpec{b: b, j: j, roof: rt, sz: sz, di: di, phase01: phase01,
				label: b.key == heroKey && j == 0 && !plan.hasHero}
		}
		queues = append(queues, q)
	}
	if len(queues) == 0 {
		return
	}

	// (2) ROUND-ROBIN emit: one instance per type per round, cycling types in the fixed
	// queue order, so CONSECUTIVE slots in plan.lots are different domains (a hut, a camp,
	// a store — not a blob of huts). The emit ORDER does not affect positions (those are
	// pure fns of (type, j) computed below), it only interleaves the slice so the fabric
	// reads intermixed both spatially and in placement order.
	for round := 0; ; round++ {
		placed := false
		for _, q := range queues {
			if round >= len(q) {
				continue
			}
			placed = true
			spec := q[round]
			anc := plan.anchors[spec.di]
			// The type's j-th instance takes spiral index round(stride*(phase+j)): a pure
			// function of (type, j), so it is fixed no matter what any sibling's count is.
			m := int(spiralStride*(spec.phase01+float64(spec.j)) + 0.5)
			// Grow on the anchor's spiral; a stable per-anchor angular phase so anchors'
			// spirals don't all align. A WONDER anchor floors the spiral radius just past its
			// clear plaza so the fabric HUGS the plaza edge (the town hugs the wonder) instead
			// of spawning inside it and being culled — which would leave a big empty gap
			// around a centered wonder.
			anchorPhase := float64(spec.di) * 1.7
			rad := cfg.slotSpacing * math.Sqrt(float64(m))
			if anc.wonder {
				rad += plazaR + cfg.roofSize
			}
			ang := float64(m)*goldenAngle + anchorPhase
			dx, dy := math.Cos(ang)*rad, math.Sin(ang)*rad
			jx, jy := slotJitter(m, spec.di, seed, cfg.jitterAmp)
			x := anc.cx + dx + jx
			y := anc.cy + dy + jy
			// Pull toward the nearest lane so the fabric grows ALONG the streets and the
			// outline follows the lanes, not a disc. Pure over the fixed lane polylines.
			x, y = pullToLane(x, y, plan.streets, cfg.laneBias)
			// Drop lots inside a wonder's clear plaza (FIX 2). Position-based, so the same
			// lot is dropped at any count — the surviving sequence stays stable-incremental.
			if insideWonderPlaza(x, y, plan.anchors, plazaR) {
				continue
			}
			lot := tdLot{
				x: x, y: y, w: spec.sz, h: spec.sz, kind: tdRoof,
				domain: spec.b.domain, category: spec.b.category, tier: spec.b.tier, roof: spec.roof,
			}
			if spec.roof == roofLong {
				lot.w = spec.sz * 1.8 // longhouses/rowhouses are elongated
			}
			if spec.label {
				lot.label = spec.b.name
				lot.prom = prominenceOf(spec.b)
			}
			plan.lots = append(plan.lots, lot)
		}
		if !placed {
			break
		}
	}
}

// fnvKey is the FNV-1a hash of a building key, used to derive a stable per-type spiral
// phase so different types interleave around a shared anchor (FIX 1).
func fnvKey(key string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619
	}
	return h
}

// pullToLane blends a point toward the nearest point on the lane network by bias in
// [0,1] (0 = unchanged, 1 = snapped onto the lane). This is what makes the intermixed
// fabric grow ALONG the streets so the town outline follows the lanes rather than
// reading as a round disc. Pure over the fixed lane polylines → stable-incremental. If
// there are no lanes it returns the point unchanged.
func pullToLane(x, y float64, streets []tdStreet, bias float64) (float64, float64) {
	if bias <= 0 || len(streets) == 0 {
		return x, y
	}
	nx, ny, ok := nearestOnStreets(x, y, streets)
	if !ok {
		return x, y
	}
	return x + (nx-x)*bias, y + (ny-y)*bias
}

// nearestOnStreets returns the closest point on any lane polyline to (x,y).
func nearestOnStreets(x, y float64, streets []tdStreet) (float64, float64, bool) {
	best := math.Inf(1)
	var bx, by float64
	found := false
	for _, s := range streets {
		for i := 0; i+1 < len(s.pts); i++ {
			px, py, d := nearestOnSeg(x, y, s.pts[i], s.pts[i+1])
			if d < best {
				best, bx, by, found = d, px, py, true
			}
		}
	}
	return bx, by, found
}

// nearestOnSeg returns the closest point on segment a→b to p, and the squared distance.
func nearestOnSeg(px, py float64, a, b tdPoint) (float64, float64, float64) {
	dx, dy := b.x-a.x, b.y-a.y
	l2 := dx*dx + dy*dy
	if l2 < 1e-9 {
		ex, ey := px-a.x, py-a.y
		return a.x, a.y, ex*ex + ey*ey
	}
	t := ((px-a.x)*dx + (py-a.y)*dy) / l2
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	cx, cy := a.x+t*dx, a.y+t*dy
	ex, ey := px-cx, py-cy
	return cx, cy, ex*ex + ey*ey
}

// insideWonderPlaza reports whether (x,y) falls within the clear plaza radius of any
// WONDER anchor (FIX 2). The bare city-center anchor of a wonderless village is NOT a
// wonder anchor, so a tiny village keeps a filled center — only a real wonder clears a
// plaza around itself.
func insideWonderPlaza(x, y float64, anchors []tdAnchor, plazaR float64) bool {
	if plazaR <= 0 {
		return false
	}
	for _, a := range anchors {
		if !a.wonder {
			continue
		}
		if math.Hypot(x-a.cx, y-a.cy) < plazaR {
			return true
		}
	}
	return false
}

// tdPlaceWonders drops each wonder's dominant, ornate centerpiece roof AT its anchor
// (locked #13, FIX 2). The grandest wonder crowns the center anchor; the rest sit at
// their spread anchors. Each is labeled and drawn prominent; the clear plaza around it
// was already carved by tdPopulateIntermixed, so the fabric never buries it. Only the
// grandest wonder carries the "City Center" hero label priority via its prominence.
func tdPlaceWonders(plan *topPlan, cfg tdConfig) {
	for i, a := range plan.anchors {
		if !a.wonder {
			continue
		}
		// The grandest wonder (anchor 0) is the largest; the rest are a touch smaller so
		// the center reads as the primary showpiece.
		scale := 3.0
		prom := 1000.0
		if i > 0 {
			scale = 2.4
			prom = 800
		}
		plan.lots = append(plan.lots, tdLot{
			x: a.cx, y: a.cy, w: cfg.roofSize * scale, h: cfg.roofSize * scale,
			kind: tdRoof, domain: a.bld.domain, category: a.bld.category, tier: a.bld.tier,
			roof: roofWonder, label: a.bld.name, prom: prom,
		})
	}
}

// moreProminentBld ranks two buildings for hero selection: higher tier, then higher
// count, then name for a stable tiebreak.
func moreProminentBld(a, b builtBuilding) bool {
	if a.tier != b.tier {
		return a.tier > b.tier
	}
	if a.count != b.count {
		return a.count > b.count
	}
	return a.name < b.name
}

// prominenceOf is a scalar prominence for label priority (tier dominant, count minor).
func prominenceOf(b builtBuilding) float64 {
	return float64(b.tier)*100 + math.Sqrt(float64(b.count))
}

// tdRoofCount maps a building's instance count to how many roof lots it emits. Near
// 1:1 at low counts (locked #4: 24 huts ≈ 24 hut-roofs) and SUB-LINEAR (sqrt) at
// high counts so a metropolis densifies without N identical clones. Residential and
// production share the same curve; both keep the near-1:1 low band. At least 1 lot so
// a lone building still appears.
func tdRoofCount(count int, role cityRole) int {
	if count <= 0 {
		return 0
	}
	// Near-1:1 up to ~12, then blend toward sqrt so high counts sub-scale. The blend
	// point + cap keep the fabric legible (locked #3 fill-frame handles the shrink).
	c := float64(count)
	var n float64
	if c <= 12 {
		n = c // near-1:1 low band
	} else {
		// 12 one-to-one, then sqrt-scaled growth beyond.
		n = 12 + 3.2*(math.Sqrt(c)-math.Sqrt(12))
	}
	out := int(n + 0.5)
	if out < 1 {
		out = 1
	}
	if out > 80 {
		out = 80 // legibility cap
	}
	return out
}

// ---- streets (organic, city space) ------------------------------------------

// tdOrganicStreets grows the lane network the settlement grows ALONG (FIX 1), winding
// BETWEEN and AROUND the growth anchors (FIX 2). It lays a winding lane from each anchor
// back to the core (spokes so every anchor is reached), plus a chain of lanes linking
// consecutive anchors (so the anchors are woven together, not isolated spokes), plus —
// for a wonderless single-anchor village — a couple of short radial stub lanes so the
// lone hamlet still has streets to grow along. Purely jittered polylines: NO terrain
// routing, no water (locked #10, V3-A organic). The generator is shaped so grid/avenue
// patterns can slot in later; only organic is tuned now.
func tdOrganicStreets(anchors []tdAnchor, cx, cy float64, style tdEraStyle, r *rng) []tdStreet {
	core := tdPoint{cx, cy}
	streets := make([]tdStreet, 0, len(anchors)*2+2)

	// Spokes: a winding lane from each anchor to the core. (For the single center anchor
	// this is a zero-length no-op, handled by the stub lanes below.)
	for _, a := range anchors {
		if math.Hypot(a.cx-cx, a.cy-cy) < 1e-6 {
			continue
		}
		streets = append(streets, windingLane(tdPoint{a.cx, a.cy}, core, style.streetJitter, style.laneWidth, r))
	}
	// Chain: link consecutive anchors so they read as one woven settlement.
	for i := 0; i+1 < len(anchors); i++ {
		a, b := anchors[i], anchors[i+1]
		streets = append(streets, windingLane(tdPoint{a.cx, a.cy}, tdPoint{b.cx, b.cy}, style.streetJitter, style.laneWidth, r))
	}
	// Stub lanes: a lone hamlet (single center anchor, no chain/spokes) still needs a few
	// streets to grow ALONG so it doesn't collapse to a disc. Grow 3 short radial lanes
	// from the core at seeded angles. Also give a little extra structure to any tiny town.
	if len(streets) == 0 {
		phase := r.f01() * 2 * math.Pi
		const stubs = 3
		stubLen := defaultTdConfig.anchorSpread * 0.6
		for i := 0; i < stubs; i++ {
			ang := phase + 2*math.Pi*float64(i)/float64(stubs)
			end := tdPoint{cx + math.Cos(ang)*stubLen, cy + math.Sin(ang)*stubLen}
			streets = append(streets, windingLane(core, end, style.streetJitter, style.laneWidth, r))
		}
	}
	return streets
}

// windingLane builds a jittered polyline from a to b: it walks a handful of segments,
// offsetting each interior waypoint perpendicular to the line by a seeded amount so
// the lane meanders like a footpath. jitter scales the wander; width is the lane
// class. Deterministic from the threaded rng.
func windingLane(a, b tdPoint, jitter float64, width int, r *rng) tdStreet {
	dx := b.x - a.x
	dy := b.y - a.y
	length := math.Hypot(dx, dy)
	if length < 1e-6 {
		return tdStreet{pts: []tdPoint{a, b}, width: width}
	}
	// Perpendicular unit vector for the wander offset.
	px, py := -dy/length, dx/length
	// Segment count scales with length so long lanes wind more; clamped small (village).
	segs := int(length/8) + 2
	if segs > 6 {
		segs = 6
	}
	pts := make([]tdPoint, 0, segs+1)
	pts = append(pts, a)
	for i := 1; i < segs; i++ {
		t := float64(i) / float64(segs)
		// Base point along the straight line.
		bx := a.x + dx*t
		by := a.y + dy*t
		// Wander: a signed offset that fades to 0 at the endpoints (bell-ish), scaled by
		// jitter and the lane length so the meander is proportional.
		amp := jitter * length * 0.10 * math.Sin(math.Pi*t)
		off := amp * (r.f01()*2 - 1)
		pts = append(pts, tdPoint{bx + px*off, by + py*off})
	}
	pts = append(pts, b)
	return tdStreet{pts: pts, width: width}
}

// ---- filler (balanced living city) ------------------------------------------

// tdAddFiller lays balanced living-city greenery into the town, then a few deliberate
// groves just outside its edge (FIX 3 / playtest). The old cut scattered trees across
// the whole empty canvas; this keeps ALL in-town greenery (gardens, squares, props, and
// street trees) strictly within the built-up footprint, and adds a SMALL number (2–4)
// of grove clusters hugging the town edge for a wooded fringe. Density is balanced
// (locked #12) — alive, not burying the buildings — and fully seeded so it's stable.
func tdAddFiller(plan *topPlan, style tdEraStyle, cfg tdConfig, r *rng) {
	// The built-up footprint: its center (the roof centroid, ~the core) and radius. All
	// in-town filler stays inside this disk; groves sit just past its edge.
	rad := tdFootprintRadius(plan)
	if rad < cfg.roofSize {
		rad = cfg.roofSize * 2
	}
	roofN := 0
	for _, lt := range plan.lots {
		if lt.kind == tdRoof {
			roofN++
		}
	}
	if roofN == 0 {
		return
	}

	dens := style.fillerDensity
	if dens <= 0 {
		dens = 1
	}
	// Counts scale with the number of roofs but stay balanced (sub-linear) so the
	// filler seasons the town rather than swamping it.
	gardens := int(dens * math.Sqrt(float64(roofN)) * 1.1)
	trees := int(dens * math.Sqrt(float64(roofN)) * 1.2)
	props := int(dens * math.Sqrt(float64(roofN)) * 0.7)
	squares := 1 + int(dens*math.Sqrt(float64(roofN))*0.3)

	// innerRad keeps filler off the very rim so it reads as woven THROUGH the town, not
	// ringing it. Groves live outside; everything below stays within innerRad.
	innerRad := rad * 0.9

	// A paved square (or two) hugging the civic core.
	for i := 0; i < squares; i++ {
		ang := r.f01() * 2 * math.Pi
		rr := r.f01() * cfg.roofSize * 2
		plan.lots = append(plan.lots, tdLot{
			x: plan.cx + math.Cos(ang)*rr, y: plan.cy + math.Sin(ang)*rr,
			w: cfg.roofSize * 1.6, h: cfg.roofSize * 1.6, kind: tdSquare,
		})
	}
	// Gardens: green plots woven through the built-up disk (in-town only).
	for i := 0; i < gardens; i++ {
		x, y := tdDiskPoint(plan.cx, plan.cy, innerRad, r)
		plan.lots = append(plan.lots, tdLot{x: x, y: y, w: cfg.roofSize * 1.3, h: cfg.roofSize * 1.1, kind: tdGarden})
	}
	// Street trees: small dot clusters sprinkled INSIDE the town (not a scatter past it).
	for i := 0; i < trees; i++ {
		x, y := tdDiskPoint(plan.cx, plan.cy, innerRad, r)
		plan.lots = append(plan.lots, tdLot{x: x, y: y, w: cfg.roofSize * 0.7, h: cfg.roofSize * 0.7, kind: tdTree})
	}
	// Props: wells/stalls, sprinkled inside the built-up area.
	for i := 0; i < props; i++ {
		x, y := tdDiskPoint(plan.cx, plan.cy, innerRad, r)
		plan.lots = append(plan.lots, tdLot{x: x, y: y, w: cfg.roofSize * 0.5, h: cfg.roofSize * 0.5, kind: tdProp})
	}

	// Groves: 2–4 deliberate stands of trees JUST OUTSIDE the town edge, each a tight
	// knot of a few tree lots so it reads as a copse, not stray dots. Placement is
	// seeded at spread angles so the groves ring the town loosely without scattering.
	groveCount := 2 + int(r.f01()*3) // 2..4
	if groveCount > 4 {
		groveCount = 4
	}
	groveBase := r.f01() * 2 * math.Pi
	for g := 0; g < groveCount; g++ {
		// Spread the groves around the compass with a little seeded wobble.
		ang := groveBase + 2*math.Pi*float64(g)/float64(groveCount) + (r.f01()-0.5)*0.6
		gr := rad * (1.05 + 0.08*r.f01()) // hugging just past the town edge
		gx := plan.cx + math.Cos(ang)*gr
		gy := plan.cy + math.Sin(ang)*gr
		// A copse: a few trees clustered tightly around (gx,gy).
		trunks := 3 + int(r.f01()*3) // 3..5
		for t := 0; t < trunks; t++ {
			ox := (r.f01()*2 - 1) * cfg.roofSize * 0.9
			oy := (r.f01()*2 - 1) * cfg.roofSize * 0.9
			plan.lots = append(plan.lots, tdLot{
				x: gx + ox, y: gy + oy,
				w: cfg.roofSize * 0.8, h: cfg.roofSize * 0.8, kind: tdTree,
			})
		}
	}
}

// tdFootprintRadius returns the max distance of any roof lot from the core — the
// built-up radius the filler + walls hug.
func tdFootprintRadius(plan *topPlan) float64 {
	max := 0.0
	for _, lt := range plan.lots {
		if lt.kind != tdRoof {
			continue
		}
		d := math.Hypot(lt.x-plan.cx, lt.y-plan.cy) + math.Max(lt.w, lt.h)/2
		if d > max {
			max = d
		}
	}
	return max
}

// tdRoofBBox returns the bounding box (minX,minY,maxX,maxY) of all roof lots including
// their extent. Used by tests to assert the town sits within the frame with margin and
// that greenery lands within/near the town footprint. Empty roof set → a zero box at
// the core.
func tdRoofBBox(plan *topPlan) (minX, minY, maxX, maxY float64) {
	minX, minY = math.Inf(1), math.Inf(1)
	maxX, maxY = math.Inf(-1), math.Inf(-1)
	any := false
	for _, lt := range plan.lots {
		if lt.kind != tdRoof {
			continue
		}
		any = true
		ex := math.Max(lt.w, lt.h) / 2
		if lt.x-ex < minX {
			minX = lt.x - ex
		}
		if lt.y-ex < minY {
			minY = lt.y - ex
		}
		if lt.x+ex > maxX {
			maxX = lt.x + ex
		}
		if lt.y+ex > maxY {
			maxY = lt.y + ex
		}
	}
	if !any {
		return plan.cx, plan.cy, plan.cx, plan.cy
	}
	return minX, minY, maxX, maxY
}

// tdDiskPoint returns a seeded point uniformly-ish within radius rad of (cx,cy).
func tdDiskPoint(cx, cy, rad float64, r *rng) (float64, float64) {
	ang := r.f01() * 2 * math.Pi
	rr := rad * math.Sqrt(r.f01())
	return cx + math.Cos(ang)*rr, cy + math.Sin(ang)*rr
}

// ---- walls (capability; primitive off) --------------------------------------

// tdAddWalls rings the built-up area with a wall + a few gates in city space (locked
// #9). V3-A wires the capability but PRIMITIVE keeps hasWalls=false, so this only
// runs for a (future) era that flips the flag — the code path is complete + tested.
func tdAddWalls(plan *topPlan, r *rng) {
	rad := tdFootprintRadius(plan) * 1.15
	if rad <= 0 {
		return
	}
	const segs = 24
	gateEvery := segs / 4 // four gates, roughly cardinal
	phase := r.f01() * 2 * math.Pi
	for i := 0; i < segs; i++ {
		ang := phase + 2*math.Pi*float64(i)/float64(segs)
		x := plan.cx + math.Cos(ang)*rad
		y := plan.cy + math.Sin(ang)*rad
		kind := tdWall
		if i%gateEvery == 0 {
			kind = tdGate
		}
		plan.lots = append(plan.lots, tdLot{x: x, y: y, w: 1.4, h: 1.4, kind: kind})
	}
}

// ---- fill-frame transform ---------------------------------------------------

// tdTransform maps city space → canvas pixels: p_px = (p_city - min) * scale +
// offset. Computed from the plan's lot bounding box so the whole city FILLS the
// canvas with a small margin (locked #3). A per-axis uniform scale (the smaller of
// the two) preserves the city's proportions; the offset centers it.
type tdTransform struct {
	scale        float64
	offX, offY   float64
	minX, minY   float64
	roofFloorPx  float64 // minimum roof extent in px (legibility floor, locked #3)
}

// computeTransform derives the fill-frame transform. It fits the BUILT city — the roof
// lots, the streets, and the core — leaving a small padding, and scales so that box
// fills the canvas. Filler/greenery is deliberately EXCLUDED from the fit: a couple of
// edge groves must not zoom the whole town out (that left the roofs tiny with a sea of
// empty ground in the playtest). Greenery lands within the town or just past its edge,
// clipping cleanly at the margin. Roofs shrink as the city densifies (more roofs →
// larger box → smaller scale) but a min roof-size FLOOR keeps them legible. Panic-safe:
// a degenerate (empty / zero-extent) box yields a centered identity-ish transform.
func computeTransform(plan *topPlan, w, h int) tdTransform {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	acc := func(x, y, ext float64) {
		if x-ext < minX {
			minX = x - ext
		}
		if y-ext < minY {
			minY = y - ext
		}
		if x+ext > maxX {
			maxX = x + ext
		}
		if y+ext > maxY {
			maxY = y + ext
		}
	}
	roofN := 0
	for _, lt := range plan.lots {
		if lt.kind != tdRoof {
			continue // greenery/props are seasoning; don't let them drive the fill-frame
		}
		roofN++
		acc(lt.x, lt.y, math.Max(lt.w, lt.h)/2)
	}
	// Streets are part of the built fabric; include them so lanes stay on-frame.
	for _, s := range plan.streets {
		for _, p := range s.pts {
			acc(p.x, p.y, 0)
		}
	}
	// Always include the core so an empty plan still centers sensibly.
	acc(plan.cx, plan.cy, 1)
	// Padding around the built box so the town breathes AND the near-edge groves (which
	// sit ~1.05–1.13× the town radius past the core) land inside the frame margin rather
	// than being flung off-canvas. ~15% covers the grove ring while still leaving the
	// roofs large (locked #4 framing: breathing room at village scale, not fill-to-edge).
	if roofN > 0 && maxX > minX && maxY > minY {
		padX := (maxX - minX) * 0.15
		padY := (maxY - minY) * 0.15
		minX -= padX
		minY -= padY
		maxX += padX
		maxY += padY
	}

	if math.IsInf(minX, 1) || maxX <= minX || maxY <= minY {
		// Degenerate: center a unit box.
		return tdTransform{scale: 1, offX: float64(w) / 2, offY: float64(h) / 2, minX: -0.5, minY: -0.5, roofFloorPx: 1}
	}

	// Margin: a few pixels so the city doesn't touch the frame edge.
	margin := 2.0
	availW := float64(w) - 2*margin
	availH := float64(h) - 2*margin
	if availW < 1 {
		availW = 1
	}
	if availH < 1 {
		availH = 1
	}
	spanX := maxX - minX
	spanY := maxY - minY
	sx := availW / spanX
	sy := availH / spanY
	scale := math.Min(sx, sy)
	if scale <= 0 || math.IsInf(scale, 0) || math.IsNaN(scale) {
		scale = 1
	}
	// Center the scaled box in the canvas.
	usedW := spanX * scale
	usedH := spanY * scale
	offX := margin + (availW-usedW)/2 - minX*scale
	offY := margin + (availH-usedH)/2 - minY*scale

	return tdTransform{scale: scale, offX: offX, offY: offY, minX: minX, minY: minY, roofFloorPx: 1}
}

// px maps a city-space point to integer canvas pixels.
func (t tdTransform) px(x, y float64) (int, int) {
	return int(math.Round(x*t.scale + t.offX)), int(math.Round(y*t.scale + t.offY))
}

// ext maps a city-space extent (half-size) to a pixel half-size, enforcing the roof
// legibility floor so roofs never shrink below a visible size (locked #3).
func (t tdTransform) ext(cityExt float64) int {
	e := cityExt * t.scale
	if e < t.roofFloorPx {
		e = t.roofFloorPx
	}
	return int(e + 0.5)
}

// ---- render (top-down) ------------------------------------------------------

// renderTopDown paints the whole top-down city onto img and returns the landmark
// overlay geometry. Paint order (city-synthesis.md §Rendering, top-down):
//
//	era-tinted ground (+ subtle seeded texture noise)
//	→ streets (dirt lanes)
//	→ district ground accents (already folded into filler squares/gardens below)
//	→ roof sprites (roof atlas, drawn BACK-TO-FRONT by y so SE shadows layer)
//	→ filler (gardens / squares / trees / props)
//	→ walls / gates (if the era has them)
//	→ (landmark labels are stamped by the overlay pass, from the returned geometry)
//
// Pure, panic-safe, exact output size. Every color is theme-derived (retints on a
// theme switch). NO terrain, NO biome, NO water — the ground is a neutral era tint.
func renderTopDown(img *image.RGBA, state game.GameState, w, h int, seed uint32) layoutGeometry {
	e := eraForAge(state.Age)
	style := tdStyleForEra(e)
	pal := newTdPal()

	// Ground first — a full era-tinted fill with subtle seeded texture noise.
	drawGround(img, style, pal, seed, w, h)

	plan := generateTopPlan(state, config.BuildingByKey(), style, seed)
	xf := computeTransform(&plan, w, h)

	// Streets (dirt lanes) under the fabric.
	streetCol := style.streetCol(pal)
	for _, s := range plan.streets {
		drawTdStreet(img, xf, s, streetCol)
	}

	// Ground accents: gardens + squares painted before roofs so a roof sits on top.
	gardenCol := style.gardenCol(pal)
	squareCol := style.squareCol(pal)
	for _, lt := range plan.lots {
		switch lt.kind {
		case tdGarden:
			cx, cy := xf.px(lt.x, lt.y)
			fillRectC(img, cx, cy, xf.ext(lt.w/2), xf.ext(lt.h/2), gardenCol)
		case tdSquare:
			cx, cy := xf.px(lt.x, lt.y)
			fillRectC(img, cx, cy, xf.ext(lt.w/2), xf.ext(lt.h/2), squareCol)
		}
	}

	// Roof sprites, BACK-TO-FRONT (sort by city y ascending) so a nearer roof's SE
	// shadow lays over the roof behind it. Collect roof lots, sort a copy of indices.
	roofIdx := make([]int, 0, len(plan.lots))
	for i, lt := range plan.lots {
		if lt.kind == tdRoof {
			roofIdx = append(roofIdx, i)
		}
	}
	sort.SliceStable(roofIdx, func(a, b int) bool {
		la, lb := plan.lots[roofIdx[a]], plan.lots[roofIdx[b]]
		if la.y != lb.y {
			return la.y < lb.y
		}
		return la.x < lb.x
	})
	for _, i := range roofIdx {
		drawRoof(img, xf, plan.lots[i], style, pal)
	}

	// Trees + props on top of the ground fabric (small, so they read among the roofs).
	treeCol := style.treeCol(pal)
	propCol := style.propCol(pal)
	for _, lt := range plan.lots {
		switch lt.kind {
		case tdTree:
			drawTree(img, xf, lt, treeCol)
		case tdProp:
			cx, cy := xf.px(lt.x, lt.y)
			drawBlock(img, cx, cy, 0, propCol)
		}
	}

	// Walls + gates last among pixels (if any) so the ring crowns the built-up edge.
	if style.hasWalls {
		wallCol := style.wallCol(pal)
		gateCol := brighten(wallCol, 0.25)
		for _, lt := range plan.lots {
			switch lt.kind {
			case tdWall:
				cx, cy := xf.px(lt.x, lt.y)
				drawBlock(img, cx, cy, xf.ext(lt.w/2), wallCol)
			case tdGate:
				cx, cy := xf.px(lt.x, lt.y)
				drawBlock(img, cx, cy, xf.ext(lt.w/2), gateCol)
			}
		}
	}

	// Landmark geometry: the city center + labeled landmark roofs (locked #7).
	return tdGeometry(&plan, xf, w, h)
}

// drawGround fills the whole canvas with the era ground tone plus a subtle, seeded
// texture: each pixel picks base or a slightly-varied alt tone from a cheap hash of
// its coordinates, so the dirt reads as textured earth rather than a flat wash. No
// water, no biome — a neutral era-tinted ground (locked #2).
func drawGround(img *image.RGBA, style tdEraStyle, pal tdPal, seed uint32, w, h int) {
	base := style.groundBase(pal)
	alt := style.groundAlt(pal)
	b := img.Bounds()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			px := b.Min.X + x
			py := b.Min.Y + y
			if px >= b.Max.X || py >= b.Max.Y {
				continue
			}
			// Cheap 2D value hash → [0,1); a fraction of pixels take the alt tone, and a
			// touch of that blend varies, so the texture is grainy but subtle.
			n := texHash(uint32(x), uint32(y), seed)
			t := 0.0
			if n < 0.22 {
				t = 0.5 + 0.5*n/0.22 // 0.5..1.0 toward alt on the speckled pixels
			}
			img.SetRGBA(px, py, blend(base, alt, t*0.6))
		}
	}
}

// texHash is a tiny deterministic 2D value hash returning a float in [0,1). Used for
// the ground texture speckle — cheap, seeded, and stable frame-to-frame.
func texHash(x, y, seed uint32) float64 {
	h := seed + x*374761393 + y*668265263
	h = (h ^ (h >> 13)) * 1274126177
	h ^= h >> 16
	return float64(h&0xffffff) / float64(0x1000000)
}

// drawTdStreet rasterizes a city-space lane polyline into pixels as a dirt path.
// width 0 draws a single-pixel line; width>=1 strokes a thicker band. The polyline
// is mapped through the fill-frame transform first, then drawn with the shared
// Bresenham road rasterizer (reused; no terrain routing involved).
func drawTdStreet(img *image.RGBA, xf tdTransform, s tdStreet, c color.RGBA) {
	if len(s.pts) < 2 {
		return
	}
	for i := 0; i+1 < len(s.pts); i++ {
		ax, ay := xf.px(s.pts[i].x, s.pts[i].y)
		bx, by := xf.px(s.pts[i+1].x, s.pts[i+1].y)
		drawRoad(img, roadSeg{ax, ay, bx, by}, c)
		for wstep := 1; wstep <= s.width; wstep++ {
			drawRoad(img, roadSeg{ax + wstep, ay, bx + wstep, by}, c)
			drawRoad(img, roadSeg{ax, ay + wstep, bx, by + wstep}, c)
		}
	}
}

// drawTree paints a tree as a small dark-green dot cluster (a filled blob with a
// darker core), so a stand of trees reads as foliage from above.
func drawTree(img *image.RGBA, xf tdTransform, lt tdLot, c color.RGBA) {
	cx, cy := xf.px(lt.x, lt.y)
	rad := xf.ext(lt.w / 2)
	if rad < 1 {
		rad = 1
	}
	fillDisc(img, cx, cy, rad, c)
	// A darker center pip for a hint of canopy depth.
	img.SetRGBA(cx, cy, darken(c, 0.25))
}

// ---- roof atlas draw --------------------------------------------------------

// drawRoof renders one building lot as a TOP-DOWN roof filling its lot: a soft SE
// drop-shadow first (subtle depth, NOT isometric walls — locked #6), then the roof
// shape read straight from above (Stardew / top-down-village style). Material comes
// from the era style; a subtle lineage tint differentiates types. EVERY tone is
// base/dark/ridge from roofColorsFor — ridge is a base-derived lighten, so NO
// saturated theme accent ever lands on a roof (the yellow-dot fix). Dispatches on the
// lot's roofType archetype.
func drawRoof(img *image.RGBA, xf tdTransform, lt tdLot, style tdEraStyle, pal tdPal) {
	cx, cy := xf.px(lt.x, lt.y)
	hw := xf.ext(lt.w / 2)
	hh := xf.ext(lt.h / 2)
	if hw < 1 {
		hw = 1
	}
	if hh < 1 {
		hh = 1
	}
	rc := roofColorsFor(style, pal, lt.domain, lt.category)

	// SE drop-shadow: the roof footprint, offset down-right by ~1px (subtle, scaled a
	// hair for big roofs), painted UNDER the roof. Soft = the theme shadow tone blended
	// into the ground, not a hard black slab — a hint of height without isometric walls.
	sh := 1 + hw/8
	drawShadow(img, cx+sh, cy+sh, hw, hh, lt.roof, pal.shadow)

	switch lt.roof {
	case roofHut:
		drawRoofHut(img, cx, cy, hw, hh, rc)
	case roofRidge:
		drawRoofRidge(img, cx, cy, hw, hh, rc)
	case roofLong:
		drawRoofRidge(img, cx, cy, hw, hh, rc)
	case roofTemple:
		drawRoofTemple(img, cx, cy, hw, hh, rc)
	case roofCamp:
		drawRoofCamp(img, cx, cy, hw, hh, rc)
	case roofStash:
		drawRoofStash(img, cx, cy, hw, hh, rc)
	case roofFlat:
		drawRoofFlat(img, cx, cy, hw, hh, rc)
	case roofWonder:
		drawRoofWonder(img, cx, cy, hw, hh, rc)
	default:
		drawRoofRidge(img, cx, cy, hw, hh, rc)
	}
}

// drawShadow paints a soft SE drop-shadow matching the roof's rough silhouette. It
// blends the shadow tone into whatever is beneath (so it darkens the ground, not
// paints a hard slab), giving a subtle hint of height. Kept faint (~0.28) so it reads
// as depth, not an outline.
func drawShadow(img *image.RGBA, cx, cy, hw, hh int, rt roofType, shadow color.RGBA) {
	blendFn := func(x, y int) {
		blendPixel(img, x, y, shadow, 0.28)
	}
	switch rt {
	case roofHut, roofWonder:
		forEllipse(cx, cy, hw, hh, blendFn)
	default:
		forRect(cx, cy, hw, hh, blendFn)
	}
}

// drawRoofHut: a small ROUND thatch roof seen from straight above — a solid thatch disc
// with soft top-down SHADING (lit toward the NW crown, shaded toward the SE eave) so it
// reads as a domed thatch cap, not a target or a pinwheel. No concentric rings, no
// radial spokes (both read as abstract tokens at this scale). It must read round, NOT a
// 4-point diamond: at tiny radii forEllipse degenerates to a plus, so the footprint is
// floored to a minimum radius (see hutRadius) before filling.
func drawRoofHut(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	rw, rh := hutRadius(hw), hutRadius(hh)
	shade := blend(rc.base, rc.dark, 0.5)
	forEllipse(cx, cy, rw, rh, func(x, y int) {
		// Radial position: shade grows toward the SE eave, base holds toward the NW crown.
		fx := float64(x-cx) / float64(rw)
		fy := float64(y-cy) / float64(rh)
		d := (fx + fy) // -2..2, negative = NW (lit), positive = SE (shaded)
		switch {
		case d > 0.55:
			img.SetRGBA(x, y, shade)
		case d < -0.55:
			img.SetRGBA(x, y, rc.ridge) // lit crown side
		default:
			img.SetRGBA(x, y, rc.base)
		}
	})
}

// hutRadius floors a hut half-extent so a hut never collapses into a plus/diamond at
// small sizes (forEllipse includes only the cardinal cells at radius 1). Minimum 2 so
// the smallest hut is still an unmistakable little round roof.
func hutRadius(half int) int {
	if half < 2 {
		return 2
	}
	return half
}

// drawRoofRidge: a rectangular PITCHED roof read from above — a filled rectangle split
// by a center ridge line running the long axis, with the two slopes shaded slightly
// differently (the ridge-ward light side = base, the far side = dark) so the pitch
// reads. Serves both the house/longhouse (roofLong is just a wider lot) and the flat/
// default fallback. The ridge line is base-derived (rc.ridge), one shade lighter.
func drawRoofRidge(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	horizontalRidge := hw >= hh // ridge runs along the wider axis
	forRect(cx, cy, hw, hh, func(x, y int) {
		var slope color.RGBA
		if horizontalRidge {
			// Upper slope catches light (base), lower slope falls into shade (dark).
			if y <= cy {
				slope = rc.base
			} else {
				slope = rc.dark
			}
		} else {
			if x <= cx {
				slope = rc.base
			} else {
				slope = rc.dark
			}
		}
		img.SetRGBA(x, y, slope)
	})
	// The ridge line down the center of the long axis: one shade lighter than base.
	if horizontalRidge {
		drawHSpan(img, cx-hw, cx+hw, cy, rc.ridge)
	} else {
		for y := cy - hh; y <= cy+hh; y++ {
			img.SetRGBA(cx, y, rc.ridge)
		}
	}
}

// drawRoofTemple: the larger, grandest small building — an ornate symmetric tiered
// roof read from above. A full base footprint, a lighter stepped inner tier, and a
// small subtle central peak, all base-derived (no accent finial — that was the yellow
// dot). Cross ridges on both axes make it read symmetric and civic.
func drawRoofTemple(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Base tier: full footprint in the shaded tone so the brighter inner tiers step up.
	forRect(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.dark) })
	// Middle tier: the base material, stepped in.
	mhw := maxInt(hw*2/3, 1)
	mhh := maxInt(hh*2/3, 1)
	forRect(cx, cy, mhw, mhh, func(x, y int) { img.SetRGBA(x, y, rc.base) })
	// A small central peak: the lighter base-derived ridge tone (grandest, but in-family).
	phw := maxInt(hw/3, 1)
	phh := maxInt(hh/3, 1)
	forRect(cx, cy, phw, phh, func(x, y int) { img.SetRGBA(x, y, rc.ridge) })
	// Cross ridges (both axes) so it reads symmetric.
	drawHSpan(img, cx-hw, cx+hw, cy, rc.ridge)
	for y := cy - hh; y <= cy+hh; y++ {
		img.SetRGBA(cx, y, rc.ridge)
	}
}

// drawRoofCamp: an OPEN lean-to / A-frame — the gathering/forager/war camp, drawn so
// it reads as a temporary open structure, clearly NOT a solid house. Only the top
// (north) half of the lot is roofed (a lean-to whose slope faces the viewer), with the
// open front left as ground: two ridge poles frame the pitch and the roof plane tapers
// from a full-width ridge at the back to nothing at the open front.
func drawRoofCamp(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Lean-to roof plane: full width at the back edge (cy-hh), tapering to the mid-line.
	for dy := -hh; dy <= 0; dy++ {
		// f: 0 at the back ridge, 1 at the open mid-line front.
		f := float64(dy+hh) / float64(hh+1)
		rowHW := int(float64(hw) * (1 - f*0.15)) // barely tapers — a broad slope
		// Slope shading: lighter toward the back ridge, darker toward the eave.
		col := blend(rc.base, rc.dark, f*0.6)
		drawHSpan(img, cx-rowHW, cx+rowHW, cy+dy, col)
	}
	// The two frame poles as darker edges, and the back ridge as the lighter crown line.
	drawLineC(img, cx-hw, cy-hh, cx-hw, cy, rc.dark)
	drawLineC(img, cx+hw, cy-hh, cx+hw, cy, rc.dark)
	drawHSpan(img, cx-hw, cx+hw, cy-hh, rc.ridge)
}

// drawRoofStash: a small, low, plain square roof — the storage store-hut, quieter than
// a dwelling. A compact filled square (the shaded tone so it sits low) with a single
// base-derived plank highlight across the top edge.
func drawRoofStash(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Keep it small + squat: clamp to a tight square.
	s := minInt(hw, hh)
	if s < 1 {
		s = 1
	}
	forRect(cx, cy, s, s, func(x, y int) { img.SetRGBA(x, y, rc.dark) })
	// One plank highlight across the top (base-derived).
	drawHSpan(img, cx-s, cx+s, cy-s, rc.ridge)
}

// drawRoofFlat: a low flat structure — a flat slab with a thin darker rim, for the
// stone works. Reads distinctly from a pitched dwelling (no ridge line).
func drawRoofFlat(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	forRect(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.base) })
	// A darker rim around the slab so it reads as a low walled platform.
	forRectOutline(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.dark) })
}

// drawRoofWonder: the DOMINANT central complex — a large ornate multi-part roof read
// from above (locked #13). A grand elliptical base, a stepped square inner hall, and a
// bright base-derived top tier + cross ridges: unmistakably the grandest thing on the
// map, and STILL in the roof material family (no accent — the crown is a base-derived
// lighten, not a saturated finial).
func drawRoofWonder(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Grand elliptical base.
	forEllipse(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.base) })
	// A square inner hall, stepped in and shaded a touch for the tier.
	ihw := maxInt(hw*2/3, 1)
	ihh := maxInt(hh*2/3, 1)
	forRect(cx, cy, ihw, ihh, func(x, y int) { img.SetRGBA(x, y, blend(rc.base, rc.dark, 0.25)) })
	// A smaller bright top tier — base-derived ridge tone (the grand crown, in-family).
	thw := maxInt(hw/3, 1)
	thh := maxInt(hh/3, 1)
	forRect(cx, cy, thw, thh, func(x, y int) { img.SetRGBA(x, y, rc.ridge) })
	// Cross ridges across the whole footprint so the complex reads ornate + symmetric.
	drawHSpan(img, cx-hw, cx+hw, cy, rc.ridge)
	for y := cy - hh; y <= cy+hh; y++ {
		img.SetRGBA(cx, y, rc.ridge)
	}
}

// ---- pixel primitives (top-down) --------------------------------------------

// forRect calls fn for every pixel of the filled (2*hw+1)×(2*hh+1) rect centered on
// (cx,cy). Clipping is the caller's setter's job (all setters clip via SetRGBA bounds
// checks in the wrappers below); here we just enumerate.
func forRect(cx, cy, hw, hh int, fn func(x, y int)) {
	for y := cy - hh; y <= cy+hh; y++ {
		for x := cx - hw; x <= cx+hw; x++ {
			fn(x, y)
		}
	}
}

// forRectOutline calls fn for the border pixels of the rect only.
func forRectOutline(cx, cy, hw, hh int, fn func(x, y int)) {
	for x := cx - hw; x <= cx+hw; x++ {
		fn(x, cy-hh)
		fn(x, cy+hh)
	}
	for y := cy - hh; y <= cy+hh; y++ {
		fn(cx-hw, y)
		fn(cx+hw, y)
	}
}

// forEllipse calls fn for every pixel inside the axis-aligned ellipse with radii
// (hw,hh) centered on (cx,cy).
func forEllipse(cx, cy, hw, hh int, fn func(x, y int)) {
	if hw < 1 {
		hw = 1
	}
	if hh < 1 {
		hh = 1
	}
	for y := cy - hh; y <= cy+hh; y++ {
		for x := cx - hw; x <= cx+hw; x++ {
			dx := float64(x-cx) / float64(hw)
			dy := float64(y-cy) / float64(hh)
			if dx*dx+dy*dy <= 1.0 {
				fn(x, y)
			}
		}
	}
}

// fillRectC fills a centered rect with a solid color, clipped to the image.
func fillRectC(img *image.RGBA, cx, cy, hw, hh int, c color.RGBA) {
	forRect(cx, cy, hw, hh, func(x, y int) { setPixel(img, x, y, c) })
}

// fillDisc fills a centered disc (ellipse with equal radii) with a solid color.
func fillDisc(img *image.RGBA, cx, cy, rad int, c color.RGBA) {
	forEllipse(cx, cy, rad, rad, func(x, y int) { setPixel(img, x, y, c) })
}

// setPixel writes a color at (x,y) if in bounds.
func setPixel(img *image.RGBA, x, y int, c color.RGBA) {
	b := img.Bounds()
	if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
		return
	}
	img.SetRGBA(x, y, c)
}

// (blendPixel lives in flourish.go and is shared — it mixes a color into the
// existing pixel by t, clipped to the image; the drop-shadows use it to darken the
// ground rather than paint over it.)

// drawLineC rasterizes a solid Bresenham line in color c (reuses drawRoad's engine).
func drawLineC(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	drawRoad(img, roadSeg{x0, y0, x1, y1}, c)
}

// ---- landmark overlay geometry ----------------------------------------------

// tdGeometry extracts the LANDMARK-ONLY overlay anchors (locked #7): the city center,
// plus one buildingLabel per labeled landmark (the civic/promoted hero and the wonder).
// No per-house labels. Returned as the same layoutGeometry the overlay pass consumes,
// so the existing label pipeline stamps them unchanged. Anchors are in pixel space.
func tdGeometry(plan *topPlan, xf tdTransform, w, h int) layoutGeometry {
	ccx, ccy := xf.px(plan.cx, plan.cy)
	geo := layoutGeometry{
		palaceX: clampInt(ccx, 0, maxInt(w-1, 0)),
		palaceY: clampInt(ccy, 0, maxInt(h-1, 0)),
	}
	for _, lt := range plan.lots {
		if lt.kind != tdRoof || lt.label == "" {
			continue
		}
		px, py := xf.px(lt.x, lt.y)
		geo.buildings = append(geo.buildings, buildingLabel{
			px: clampInt(px, 0, maxInt(w-1, 0)), py: clampInt(py, 0, maxInt(h-1, 0)),
			name: lt.label, lineageKey: lt.domain, category: lt.category, tier: lt.tier,
			size: xf.ext(math.Max(lt.w, lt.h) / 2),
		})
	}
	return geo
}

// buildLandmarkOverlay assembles the KEY-LANDMARKS-ONLY overlay plan for the
// top-down city (locked #7): the labeled landmark roofs (city-center hero, wonder,
// or a promoted hero when the civ has no civic building), the "City Center" label,
// and the corner title. It deliberately OMITS the old systems-weave (per-house
// labels, the diplomacy civ-edge ring, and trade-lane tags) — the city view reads by
// roof shape/color, not a wall of text. Reuses the existing overlayPlan builders +
// stampOverlay unchanged, so a theme switch still retints the text.
func buildLandmarkOverlay(state game.GameState, cols, rows int, geo layoutGeometry) overlayPlan {
	var plan overlayPlan
	if cols <= 0 || rows <= 0 {
		return plan
	}
	occupied := map[int]bool{}
	// Landmark building labels (already limited to the hero + wonder in tdGeometry).
	plan.addBuildingLabels(geo, cols, rows, occupied)
	// The City Center label under the central marker.
	plan.addCityCenterLabel(geo, cols, rows, occupied)
	// Corner title last so it crowns its corner.
	plan.addTitle(state, cols, rows)
	return plan
}
