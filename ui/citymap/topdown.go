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
// design-and-architecture/city-synthesis.md). It renders a TOP-DOWN pixel-art living
// city — you look straight down at roofs, packed-earth streets, gardens and squares,
// and the fabric grows and re-skins as the civ grows.
//
// LAYOUT MODEL — VORONOI BLOCKS / WARDS (map-overhaul-citymap). The city is a compact
// cluster of BLOCKS (wards); the STREETS are the GAPS BETWEEN the blocks (the Watabou
// Medieval Fantasy City Generator approach), NOT lines lined with buildings. The pipeline:
//
//	(1) TOWN AREA   — a bounded, compact disc sized by √(building count) (tdTownRadius).
//	(2) BLOCK SEEDS — B seed points scattered deterministically from citySeed (golden-angle
//	    phyllotaxis), B scaling modestly with the count, then a few LLOYD RELAXATION passes
//	    (each seed → its region centroid) for even, organic block sizes.
//	(3) REGIONS     — a RASTER nearest-seed partition of the town disc (NOT computational-
//	    geometry Voronoi): each town cell is assigned to its nearest seed. Simple + robust.
//	(4) STREETS     — the REGION BOUNDARIES: cells where the nearest and second-nearest seed
//	    are ~equidistant (within a width band) are STREET cells, painted in the bold packed-
//	    earth tone. A connected organic network with real junctions BY CONSTRUCTION.
//	(5) BLOCKS      — the region INTERIORS (inset from the boundary): each block is FILLED
//	    with buildings around its perimeter facing the surrounding streets, INTERMIXING types
//	    across blocks, count-driven, with no roof overlap.
//	(6) PLAZA+WONDER— the central region(s) stay OPEN as the plaza/town square (building-
//	    free, dressed with paving + props); wonders occupy central regions, drawn prominent.
//	(7) FILLER      — gardens / ponds / trees / props fill leftover in-block space, kept
//	    IN-TOWN (not scattered across the empty map).
//
// SCOPE: this is the citymap render path ONLY. The worldmap (worldmap.go + terrain.go's
// world funcs) is approved and untouched — the citymap simply stops calling terrain, the
// isometric drawVolume, and land-gating. There is no terrainField, no biome, no water: the
// ground is a neutral era tint and every green thing is BUILT.
//
// Determinism: everything below is pure and seeded from citySeed (a stable hash of the civ
// display name, AGE-INDEPENDENT). The seed scatter + Lloyd relaxation + raster partition are
// all pure functions of (seed, count), so the same state renders an identical town; only the
// fill-frame scale re-fits as the city densifies. Panic-safe on any canvas.
//
// STABILITY TRADEOFF (best-effort, noted for review): the block STRUCTURE is BANDED — the
// seed count B is a step function of √(roof count) (tdBlockSeedCount), so the whole block
// field is STABLE within a size band and only re-forms at a band boundary or on age-up. Within
// a band, adding a building progressively fills the existing blocks (buildings are distributed
// across blocks by a stable per-block cursor), so existing roofs mostly hold their place and a
// handful of new ones slot into open block perimeter. A real-town look is the priority over the
// strict slot-for-slot incrementality the old lane model guaranteed.

// ---- era style presets (top-down) ------------------------------------------

// tdEraStyle is the top-down era preset: the ground tone recipe, street/paving tones, roof
// material recipe, the walls capability flag, and the filler density. Colors are expressed as
// THEME-ROLE RECIPES (resolved at render time) so the whole city retints on a theme switch —
// no color is hard-coded.
//
// A recipe is a small closure over a resolved palette that returns a color, so the preset
// stays theme-agnostic while still choosing an era MOOD (earthy vs stone vs neon).
// tdResolveStyle turns the preset into concrete colors for one frame.
type tdEraStyle struct {
	name string

	// Roof material: base + shaded-slope recipes, blended per building with a subtle lineage
	// tint so a temple reads different from a hut without leaving the era mood. The
	// ridge/crown highlight is NOT a preset recipe — roofColorsFor derives it as a fixed
	// lighten of base so it can never resolve to a saturated accent (the yellow-dot
	// regression). Only the base material + the shaded slope are era recipes.
	roofBase   func(tdPal) color.RGBA // dominant roof fill (e.g. thatch brown)
	roofDark   func(tdPal) color.RGBA // shaded slope
	lineageMix float64                // how much lineage tint bleeds into the roof (~0.15–0.25)

	// Ground: base fill + a slightly-varied texture tone, both era-tinted.
	groundBase func(tdPal) color.RGBA
	groundAlt  func(tdPal) color.RGBA
	// Street surface color — the bold packed-earth tone of the block-boundary gaps.
	streetCol func(tdPal) color.RGBA

	// streetEdge is the crisp darker edge one shade below the packed-earth street surface
	// (playtest FIX 1: BOLD packed-earth roads). A subtle darken so the trodden gaps read
	// with a defined shoulder against the dirt ground, without a hard black outline.
	// Theme-derived like streetCol; nil → derived fallback.
	streetEdge func(tdPal) color.RGBA

	// Living-city filler accents.
	gardenCol func(tdPal) color.RGBA
	squareCol func(tdPal) color.RGBA
	treeCol   func(tdPal) color.RGBA
	propCol   func(tdPal) color.RGBA

	// pondCol is the BUILT decorative-pond water tone (playtest polish FIX 4). A pond is a
	// small MADE water feature dug into the town fabric (a garden-family ornament), NOT
	// natural terrain water — the citymap has no biome/ocean (locked #2). nil → tdPondColor
	// derives a safe fallback. Drawn as a small blue blob with a lighter rim by drawPond.
	pondCol func(tdPal) color.RGBA

	// pavedCol is the TOWN-SQUARE paved-stone ground tone (playtest FIX: dress the
	// wonder/center plaza as a made surface, not bare dirt). Distinct from the era-tinted
	// dirt — a lighter/greyer packed-or-paved tint derived from the ground family. nil →
	// tdPavedColor derives a safe fallback. Drawn under the wonder roof + props.
	pavedCol func(tdPal) color.RGBA

	// Walls capability (locked #9). Primitive sets false; walls arrive in V3-B.
	hasWalls bool
	wallCol  func(tdPal) color.RGBA

	fillerDensity float64 // balanced living-city filler amount (locked #12)

	// slotSpacing is the per-era PACK-DENSITY knob (playtest FIX 2 groundwork). Under the
	// Voronoi-block model it scales the in-block PERIMETER STEP + inset (tighter spacing →
	// more roofs per block, denser wards). PRIMITIVE stays AIRY (2.4, its look must not
	// change); later-era presets set it progressively TIGHTER so V3-B/C cities + the
	// metropolis pack denser. 0 → fall back to defaultTdConfig.slotSpacing. Routed into
	// tdConfig.slotSpacing by generateTopPlan.
	slotSpacing float64
}

// tdPal is the small set of resolved theme colors the style recipes draw from. Built once
// per frame from the active theme, so a theme switch re-resolves every recipe.
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

// newTdPal resolves the theme roles the era recipes need. Pure read of the active theme; no
// locks. Kept separate from the legacy terrainPalette so the top-down path never pulls a
// biome/water tone.
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

// organicVillageStyle is the fully-tuned PRIMITIVE/STONE preset (locked era table row
// "organic"): earthy dirt+grass ground, bold packed-earth street gaps, thatch/wood roofs, NO
// walls, a lived-in but not cluttered filler density. This is the V3-A showcase era.
var organicVillageStyle = tdEraStyle{
	name: "organic",

	// Thatch/wood browns, anchored to a warm earthen hue but pulled from theme roles
	// (RoleText carries the warm neutral in these themes; RoleDim grounds it) so a dark or
	// light theme still yields a legible, in-family roof.
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
		// BOLD packed-earth street gap (playtest FIX 1): a WORN, TRODDEN TAN that reads clearly
		// LIGHTER than the era-tinted dirt ground so the gaps between blocks stand out as
		// streets. Start from the dirt anchor, then lift toward the theme's light neutral
		// (RoleText) and a touch toward the pale stone anchor for a bleached, packed-trail cast.
		packed := blend(blend(p.bg, p.dim, 0.28), dirtAnchor, 0.42)
		return blend(blend(packed, p.text, 0.42), stoneAnchor, 0.22)
	},
	streetEdge: func(p tdPal) color.RGBA {
		// Crisp shoulder: the packed-earth surface darkened a touch so the trodden gap has a
		// defined edge against the dirt, without a hard black outline.
		packed := blend(blend(p.bg, p.dim, 0.28), dirtAnchor, 0.42)
		surface := blend(blend(packed, p.text, 0.42), stoneAnchor, 0.22)
		return darken(surface, 0.22)
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

	// Decorative-pond water (playtest polish FIX 4): a BUILT little pool, so a muted
	// blue/teal — the theme background grounded a touch toward dim, then pulled toward the
	// water anchor. Kept in-family (blended, never raw).
	pondCol: func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.18), waterAnchor, 0.55)
	},

	// Town-square paving: packed pale stone. Start from the dirt square tone, then lift it
	// toward the theme's light neutral (RoleText) and pull the last touch toward a grey
	// stone anchor so it reads LIGHTER + GREYER than the surrounding trodden dirt — a
	// deliberately made surface, still earthy/village in mood, still theme-derived.
	pavedCol: func(p tdPal) color.RGBA {
		earthy := blend(blend(p.bg, p.dim, 0.34), dirtAnchor, 0.30)
		return blend(blend(earthy, p.text, 0.30), stoneAnchor, 0.32)
	},

	hasWalls: false, // primitive: no walls (locked — walls arrive in V3-B)
	wallCol: func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.40), dirtAnchor, 0.35)
	},

	fillerDensity: 1.0,
	// PRIMITIVE stays AIRY: this is the current defaultTdConfig.slotSpacing value, so the
	// primitive village look does NOT change. Later eras tighten it (see defaultTdStyle).
	slotSpacing: 2.4,
}

// defaultTdStyle is the reasonable fallback for every non-primitive era until V3-B/C/D tunes
// them (locked Phasing: "Other eras use a reasonable default preset"). It reuses the organic
// recipes for a legible, theme-tinted city with a simple ridged-rect default roof — NOT its
// final silhouette. Kept as a copy so tuning organic later can't accidentally restyle every era.
var defaultTdStyle = func() tdEraStyle {
	s := organicVillageStyle
	s.name = "default"
	// PER-ERA DENSITY (playtest FIX 2 groundwork): later eras pack TIGHTER than the airy
	// primitive village so V3-B/C cities + the metropolis read denser. This is framework only
	// — the primitive preset keeps its airy 2.4; every not-yet-tuned era renders on this
	// default at a tighter slot spacing. V3-B/C/D can dial each era band's own value.
	s.slotSpacing = 1.7
	return s
}()

// Muted hue anchors for the earthy village mood. Blended at modest strength against theme
// roles so a light or dark theme still gets an in-family, non-cartoon palette.
var (
	earthAnchor = color.RGBA{R: 0x8a, G: 0x63, B: 0x3a, A: 0xff} // thatch/wood brown
	dirtAnchor  = color.RGBA{R: 0x7c, G: 0x66, B: 0x46, A: 0xff} // packed dirt
	grassAnchor = color.RGBA{R: 0x4f, G: 0x6f, B: 0x33, A: 0xff} // built greenery
	// stoneAnchor is the pale packed-stone anchor for TOWN-SQUARE paving — a light warm grey
	// so a plaza reads as a made surface, lighter/greyer than the dirt, blended against theme
	// roles (never used raw) so it retints and stays in-family.
	stoneAnchor = color.RGBA{R: 0xa8, G: 0xa0, B: 0x92, A: 0xff} // packed pale stone
	// waterAnchor is the muted blue-teal anchor for BUILT decorative ponds (playtest polish
	// FIX 4) — a calm pool blue, never used raw: blended against theme roles so a pond
	// retints and stays in-family with the village palette rather than reading as cartoon water.
	waterAnchor = color.RGBA{R: 0x36, G: 0x6b, B: 0x8f, A: 0xff} // muted pond blue-teal
)

// tdStyleForEra returns the tuned preset for an era band, or defaultTdStyle for the bands
// V3-A leaves on the fallback. Organic (primitive, stone) is the only tuned band in V3-A;
// every other band renders a legible default city.
func tdStyleForEra(e era) tdEraStyle {
	switch e {
	case eraOrganic:
		return organicVillageStyle
	default:
		return defaultTdStyle
	}
}

// ---- roof atlas -------------------------------------------------------------

// roofType is a top-down roof archetype. Each renders as a filled roof shape that FILLS its
// lot, with a ridge/texture highlight and a soft SE drop-shadow (locked #5,#6). Material comes
// from the era style; a subtle lineage tint differentiates types within the era mood.
type roofType int

const (
	roofHut    roofType = iota // small round/oval thatch roof + radial streaks
	roofRidge                  // rectangle pitched roof: center ridge + two shaded slopes
	roofLong                   // elongated ridge roof (longhouse / rowhouse)
	roofTemple                 // larger ornate symmetric roof + finial (shrine/temple)
	roofCamp                   // lean-to / open-frame / tent (gathering/forager camp)
	roofStash                  // small square store hut (storage)
	roofFlat                   // low flat structure (stone camp / workshop)
	roofWonder                 // large multi-part ornate complex (centerpiece)
)

// getRoofType maps a building's domain/category/tier to a roof archetype. Primitive
// archetypes are tuned precisely (per city-synthesis.md §roof atlas): huts round, longhouses
// elongated, shrines/temples ornate, camps as tents, stashes as store huts, stone works flat.
// Other eras fall through to roofRidge (the simple default roof) until V3-B/C specializes
// them. Pure data — safe on the render path.
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

// roofColors bundles the three material tones a roof draws with, already blended with the
// building's subtle lineage tint (~15–25%) so types differ within the era.
type roofColors struct {
	base, ridge, dark color.RGBA
}

// roofColorsFor resolves the era material for a lot's domain/category, then bleeds the lineage
// tint into it by the style's lineageMix. So every roof stays in the era mood (thatch brown)
// while a temple leans faith-violet, a forge leans forge-orange, etc. — the "subtle lineage
// tint" of locked #6.
//
// CRITICAL (yellow-dot regression guard): the ridge/highlight tone is derived from the roof
// BASE (a lighter shade of the same material), NEVER from an accent/highlight theme role. An
// earlier cut painted the ridge from style.roofRidge, whose recipe in some themes pulled
// RoleAccent/RoleHighlight (bright yellow) and stamped it at each roof's crown → the "bright
// yellow center-dot" playtest bug. Ridge is now a fixed lighten of base so a roof crown can
// only ever read as a lit patch of its own thatch. dark stays the era shaded-slope recipe.
func roofColorsFor(style tdEraStyle, pal tdPal, domain, category string) roofColors {
	base := style.roofBase(pal)
	dark := style.roofDark(pal)
	if style.lineageMix > 0 {
		// Desaturate the lineage color before bleeding it in. The raw role colors for faith
		// (violet), wonders (gold), culture (magenta) are highly saturated; blended straight
		// they poster-paint a temple's ridge into a neon cross. Pulling the tint most of the
		// way to a same-lightness gray keeps the HUE cue ("a temple leans violet") while
		// staying an earthy, muted thatch — the "subtle lineage tint" of locked #6.
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

// mutedTint desaturates a lineage color toward a gray of the same lightness, so the subtle
// roof tint nudges hue without going neon. Keeps ~35% of the original chroma.
func mutedTint(c color.RGBA) color.RGBA {
	h, s, l := rgbToHSL(c)
	return hslToRGB(h, s*0.35, l)
}

// clampRoofSat caps a roof tone's saturation so no roof — even a violet-faith temple or a
// gold-wonder — can read as a saturated poster-paint token. A hard ceiling on chroma
// guarantees the whole atlas stays in the earthy thatch family (the locked "muted & natural
// saturation" of #6), which is also what keeps the ridge safely off any accent.
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
	tdPond                    // a BUILT decorative pond (small made water feature, FIX 4)
	tdWall                    // a wall segment
	tdGate                    // a gate in the wall ring

	// Town-square lots (playtest FIX: the wonder/center plaza is DRESSED as a town square, not
	// left as bare dirt). tdPlaza is the paved-stone ground patch under the wonder/center roof
	// + props; the tdProp* kinds are the deliberate, seeded, era-appropriate props arranged
	// AROUND (never overlapping) the roof. Kept as distinct kinds so each prop gets its own
	// small top-down draw routine and later eras can swap the set (fountain/statue/benches)
	// without disturbing the primitive one.
	tdPlaza       // paved-stone town-square ground (drawn under the wonder/center roof)
	tdPropWell    // a stone well head (ring + dark shaft)
	tdPropFirepit // a firepit / hearth (dark ring + ember center)
	tdPropStones  // standing stones / a totem (a couple of upright dabs)
	tdPropStall   // a market stall (awning patch)
)

// tdLot is one placed thing, in CITY SPACE (pre-fill-frame). x,y is the lot center in city
// units; w,h its extent. Roof lots carry the building identity so the roof atlas + landmark
// labels can read them. Because generation is a pure function of (seed, count), a lot's
// city-space position is stable frame-to-frame.
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

// tdAnchor is one GROWTH ANCHOR the settlement centers on. Anchors are the built WONDERS (each
// wonder seats one, occupying a central Voronoi region drawn prominent); a civ with zero
// wonders has a single anchor at the town center. `wonder` flags an anchor that seats a wonder
// (gets a clear plaza region + the centerpiece roof); the bare city-center anchor of a
// wonderless village still clears a MODEST central plaza (a small gathering square) so the
// heart reads as a town square without hollowing a hut village into a donut. Positions are a
// stable function of (index, seed) — age-independent — so the anchor field never rearranges.
type tdAnchor struct {
	cx, cy float64
	wonder bool          // seats a wonder → plaza-cleared centerpiece
	bld    builtBuilding // the wonder building this anchor seats (wonder anchors only)
}

// tdPoint is a float city-space coordinate.
type tdPoint struct{ x, y float64 }

// topPlan is the render-ready top-down plan in CITY SPACE. renderTopDown computes the
// fill-frame transform from its lot bounding box, then paints everything mapped into canvas
// pixels. Keeping the plan in city space is what makes fill-frame work: the same relative
// layout re-fits to any canvas / density.
type topPlan struct {
	// streetCells are the centers of the STREET raster cells — the gaps between the Voronoi
	// blocks (a region boundary is a street). Painted as bold packed-earth squares of size
	// cellSize (drawStreetCells). The connectivity test unions them by grid adjacency: one
	// connected network with real junctions BY CONSTRUCTION. Derived from the raster
	// partition, not from the building count directly.
	streetCells []tdPoint
	cellSize    float64 // city-space size of one raster cell (a street square's extent = cellSize/2)
	// townR is the bounded town-disc radius (tdTownRadius) used to size the raster and bound
	// the fill-frame fit even before any roof exists.
	townR float64
	// form is the town-form archetype (organic / radial / grid / ribbon) picked for this civ+era
	// (tdPickTownForm). It selects the block-seed scatter strategy so no two towns read alike and
	// primitive villages ramble organically rather than all reading as radial wheels. Kept on the
	// plan so tests + future dressing can read which form a town took.
	form tdTownForm
	// wardSeeds are the relaxed Voronoi ward centers (the block-field seeds after Lloyd), in city
	// space, kept for tests + future ward-level dressing. The RADIAL form's wards spiral centers-out
	// (a strong seed-index↔radius correlation — the wagon-wheel signature); ORGANIC/RIBBON wards do
	// not (blue-noise / linear), which is how the anti-wheel test tells an organic town from a wheel.
	wardSeeds []tdPoint
	// anchors are the central growth seats: the built wonders (each in a central region) plus,
	// for a wonderless village, one town-center anchor. Kept on the plan so the square dressing
	// + tests can read the growth skeleton.
	anchors []tdAnchor
	lots    []tdLot
	// center is the civic heart in city space (the town-center anchor + grandest wonder seat).
	cx, cy float64
	// heroLabel is the promoted hero's identity when the civ has no civic building (locked #7).
	// Empty otherwise.
	heroLabel tdLot
	hasHero   bool
}

// ---- civ seed (stable, age-independent) -------------------------------------

// citySeed derives a STABLE-per-civ seed from the display name, the SAME way worldTerrainSeed
// does (FNV-1a over the name), but with a DISTINCT salt so the city plan and the world land
// don't share a lattice. Crucially it is AGE-INDEPENDENT (locked #8): the bones must not move
// across ages — only the era re-skin changes. An empty/anonymous name falls back to a fixed
// non-zero seed so the anonymous city is still stable.
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

// displayNameOf pulls the civ display name the citySeed hashes, matching the worldmap's rule
// (AccountStats.DisplayName, else ""). Kept local so the two seeds hash exactly the same input.
func displayNameOf(state game.GameState) string {
	if state.AccountStats != nil {
		return state.AccountStats.DisplayName
	}
	return ""
}

// ---- config -----------------------------------------------------------------

// goldenAngle is the golden angle in radians (~137.5°). Placing item i at angle i*goldenAngle
// with radius ∝ sqrt(i) gives a stable, low-overlap phyllotaxis SPREAD: every item has a FIXED
// index, so the initial block-seed scatter fills the town disc evenly (before Lloyd relaxation
// evens it further). Used for the block-seed scatter (tdScatterSeeds) and the wonder-anchor
// spread (tdAnchorPoints).
const goldenAngle = 2.399963229728653 // math.Pi * (3 - sqrt(5))

// slotJitter returns a small ORGANIC offset keyed by (i, di, seed) via a hash — a pure
// function, NOT drawn from the threaded rng — so a given slot's jitter is identical frame to
// frame. Under the Voronoi-block model it perturbs in-block roof positions off the perfectly-
// regular perimeter walk so a block reads natural rather than mechanical. amp is the max
// wander in city units; callers clamp it further so jitter never pushes a roof onto a street or
// into a neighbour.
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

// tdConfig holds the fixed generator constants (city-space units). City space is an abstract
// plane; the fill-frame transform later maps the plan's bounding box onto the canvas, so these
// are relative sizes, not pixels.
type tdConfig struct {
	slotSpacing float64 // per-era pack density (scales in-block perimeter step + inset)
	roofSize    float64 // base roof extent in city units
	jitterAmp   float64 // organic per-roof wander (city units) breaking the block lattice
	minGap      float64 // minimum clear space kept between any two roof lots (never touching)

	// plazaRadius is the clear ground kept around a WONDER anchor (in units of roofSize) so the
	// city fabric never buries the centerpiece (the playtest complaint). Roofs whose block
	// perimeter would land inside a wonder's plaza are dropped; the paved town square fills it.
	// This is the ROOMY plaza of the RADIAL/planned forms (a monument/forum-planned town's
	// identity); the ORGANIC form overrides it with a much smaller one-ward plaza (organicPlazaRadius)
	// so a rambling village never carves a dominant central void (the wagon-wheel signature).
	plazaRadius float64

	// organicPlazaRadius is the ORGANIC form's plaza clear-radius (in units of roofSize) — a MODEST,
	// one-ward-sized clearing (map-overhaul-citymap de-radialization). The organic village must NOT
	// clear a big central ring: buildings fill right up to near the center and the wonder/center gets
	// only a small dressed square, so the Voronoi mesh reads as a rambling web, not a hub-and-ring
	// wheel. Resolved per-form by tdPlazaRadius; every non-organic form keeps the roomy plazaRadius.
	organicPlazaRadius float64

	// organicEdgeAmp is how far (fraction of townR) the ORGANIC town outline is bitten INWARD by a
	// smooth, seeded, angle-varying perturbation (tdOrganicRadiusAt) so the silhouette is a rambling
	// BLOB, not a clean radial disc. INWARD-ONLY (never past townR) so every ward seed stays inside
	// the bounded footprint (the compact/anti-pinwheel guarantees hold). 0 → a plain circular disc
	// (what every non-organic form uses, unchanged). This is the DEEPEST bite of the LOW-frequency
	// lobe terms (the big bays/peninsulas that make the shape unmistakably non-circular); the
	// higher harmonics ride on a fraction of it (organicEdgeRipple) for finer wobble.
	organicEdgeAmp float64

	// organicEdgeFloor is the MINIMUM organic outline radius as a fraction of townR
	// (map-overhaul-citymap FIX 2 — ragged outline): the deepest a bay may pinch INWARD. Kept well
	// above 0 so the town stays a STAR-SHAPED domain about the core (every ray from the center is
	// ≥ this fraction of townR) → the footprint is always simply-connected, so no bay can ever carve
	// a detached island and the street web stays ONE component (streets-connected holds). Lower →
	// deeper, more dramatic pinches. 0 → fall back to a safe default.
	organicEdgeFloor float64

	// --- Voronoi block model (map-overhaul-citymap) --------------------------------
	// The town is a COMPACT, BOUNDED disc whose radius grows only ~SQRT with the fabric count —
	// NOT radial spokes (the retired pinwheel). townBaseRadius is the disc radius (city units)
	// at the smallest settlement; townGrowth scales it up with √(roof count). tdTownRadius(n) =
	// townBaseRadius + townGrowth·√n.
	townBaseRadius float64
	townGrowth     float64

	// cellSize is the RASTER cell size (city units) of the nearest-seed partition. Smaller →
	// finer block boundaries (thinner street gaps) at the cost of a denser raster. Sized so the
	// street gaps read as a couple of pixels after fill-frame.
	cellSize float64
	// streetBand is the boundary WIDTH in city units: a cell is a STREET cell if the difference
	// between its distance to the nearest and second-nearest seed is < streetBand. Wider →
	// bolder streets, smaller blocks. Scaled off cellSize so the band is always a few cells.
	streetBand float64
	// blockInset is the extra clear margin (city units) held between a street cell and a roof,
	// on TOP of the street half-width — so buildings sit inset INSIDE their block, facing the
	// surrounding street, never on it.
	blockInset float64

	// seedBase / seedGrowth drive the BANDED block-seed count: B = seedBase +
	// round(seedGrowth·√(roof count)), stepped so the block field is stable within a size band
	// (the stability tradeoff — see the file header). seedMax caps it so a metropolis stays a
	// legible ward count, not confetti.
	seedBase   int
	seedGrowth float64
	seedMax    int

	// organicSeedGrowth / organicSeedMax OVERRIDE seedGrowth / seedMax for the ORGANIC form
	// (map-overhaul-citymap de-radialization). A rambling MESH village needs MANY small wards: with
	// only ~8 wards a disc's Voronoi is inescapably a center-cell + a ring of neighbours whose
	// boundaries radiate as spokes (a wagon wheel), no matter how the seeds are scattered. Far more,
	// finer wards make the partition read as an irregular WEB with loops. Set well above the planned
	// forms' counts; other forms keep the coarser seedGrowth/seedMax (their look is unchanged).
	organicSeedGrowth float64
	organicSeedMax    int
	// lloydPasses is how many Lloyd relaxation iterations even the block sizes (each seed → its
	// region centroid). A few passes is plenty; more just converges to a hex lattice.
	lloydPasses int

	// anchorSpread is how far the outermost wonder anchor sits from the core (∝√N), kept well
	// INSIDE the town disc so wonders occupy CENTRAL regions the town hugs (FIX 2).
	anchorSpread float64
}

var defaultTdConfig = tdConfig{
	slotSpacing: 2.4,
	roofSize:    3.2,
	jitterAmp:   0.55,
	minGap:      1.1,
	// plazaRadius: the clear-ground ring around a WONDER anchor, in units of roofSize (playtest
	// polish FIX 3: roomy). The paved town square fills this radius (tdPlaceSquares), so a wider
	// plaza is a wider, more deliberate square; the wonder roof (tdWonderScale) stays far inside
	// it so a generous paved ring always shows. This is the RADIAL/planned-form plaza; ORGANIC
	// overrides it (organicPlazaRadius) so a village never carves a dominant central void.
	plazaRadius: 3.0,
	// organicPlazaRadius: the ORGANIC form's MODEST one-ward plaza (map-overhaul-citymap). Kept just
	// large enough that the grandest wonder roof (half = 2.6·roofSize/2 = 1.3·roofSize) still shows a
	// thin paved ring + its ringed props, but small enough that the fabric fills close to center and
	// the mesh never reads as a hub-and-ring wheel. Much smaller than the roomy radial 3.0.
	organicPlazaRadius: 2.0,
	// organicEdgeAmp: bite the organic outline INWARD by a big, seeded, LOW-frequency amount (up to
	// ~40% of townR at the deepest bay) so the village silhouette is an unmistakably IRREGULAR blob
	// with real bays/peninsulas — NOT the timid near-circle the first cut produced. Inward-only, and
	// the outline is floored (organicEdgeFloor) so the bite can pinch deep without ever splitting the
	// town. Different seeds get different lobe amplitudes+phases → different silhouettes.
	organicEdgeAmp: 0.40,
	// organicEdgeFloor: a bay may pinch in to ~52% of townR at the deepest — dramatic enough to read
	// as a genuine cove/waist, but the outline stays STAR-SHAPED about the core (min ray > 0), so the
	// footprint is always one connected blob (no detached islands; streets stay one component).
	organicEdgeFloor: 0.52,
	// Compact + bounded: a small base disc that grows only ~sqrt with the count.
	townBaseRadius: 16,
	townGrowth:     3.4,
	// Raster: cell ~1.3 city units so street gaps render ~1–3px after fill-frame. streetBand is the
	// boundary-classification width (map-overhaul-citymap FIX 1 — NARROWER roads): cut from the old
	// 2.1 down to ~cellSize so a block boundary is a SOLID ~1-cell web instead of a ~2-cell one — a
	// THIN village LANE, not a wide avenue. This near-halves the street-cell FRACTION (streets stop
	// dominating; the wards hold far more of the town). It is deliberately kept a hair ABOVE cellSize
	// (1.35 > 1.3) rather than pushed thinner: below ~cellSize the raster can't guarantee a contiguous
	// boundary (a Voronoi edge crosses cells at an angle, and a sub-cell band leaves gaps), which
	// fragments the street web into disconnected pieces — so ~1 cell is the thin-but-CONNECTED floor.
	// The drawn width is trimmed further in drawStreetCells so the lanes read even thinner on screen.
	cellSize:   1.3,
	streetBand: 1.35,
	blockInset: 1.0,
	// Banded block count: 2 wards at the smallest, +~0.85·√n, capped at 22 wards for a metropolis.
	seedBase:   2,
	seedGrowth: 0.85,
	seedMax:    22,
	// ORGANIC needs a much FINER ward mesh (many small blocks) so the street web reads as an
	// irregular loop network, not 8 pie-slice spokes around a hub. Bumped 2.6→3.0 alongside the
	// RAGGED outline (map-overhaul-citymap FIX 2): the deeper bays clip a couple of would-be-interior
	// wards into rim-touching ones at small towns, so a slightly finer mesh keeps enough ENCLOSED
	// faces (loops) for the anti-wheel guarantee to hold with margin even at the smallest villages.
	organicSeedGrowth: 3.0,
	organicSeedMax:    48,
	lloydPasses:       3,
	// anchorSpread kept well inside tdTownRadius so wonders sit central and the town hugs them.
	anchorSpread: 9,
}

// tdCountBand snaps a fabric roof count to the LOW EDGE of its size band — the crux of the
// STABILITY TRADEOFF (see the file header). The whole Voronoi block field (town radius + block-
// seed count + the raster) is derived from this BANDED count, not the raw one, so the block
// structure is IDENTICAL for every count within a band and only re-forms at a band boundary (or
// on age-up). Bands are geometric in √n (width bandStep in √-space): fine at low counts (a hamlet
// re-bands often but stays tiny) and coarse at high counts (a metropolis's wards are stable across
// wide count swings). n≤0 → band 0. Pure.
func tdCountBand(n int) int {
	if n <= 0 {
		return 0
	}
	const bandStep = 1.5 // √-space band width; larger → coarser (more stable) bands
	b := math.Floor(math.Sqrt(float64(n)) / bandStep)
	edge := b * bandStep
	return int(edge*edge + 0.5)
}

// ---- town-form archetypes (map-overhaul-citymap V3-A) -----------------------
//
// The Voronoi ward machinery is one pipeline, but the SEED DISTRIBUTION it starts from decides
// what KIND of town emerges — so no two towns read alike and, crucially, not every town is a
// radial "wagon wheel". A town's FORM is picked once (deterministic, era-weighted) and selects
// one of four seed-scatter strategies; everything downstream (Lloyd relaxation, the raster
// nearest-seed partition, streets = ward boundaries, buildings fill wards) is IDENTICAL across
// forms. Only tdScatterSeedsFor + a couple of per-form constraints differ.
//
//	formOrganic — POISSON-DISK / jittered-random scatter in the disc, NO radial bias and NO forced
//	    ring/spokes: rambling irregular wards, organic streets. This is what a village looks like
//	    and it is what kills the wheel (the old default). PRIMITIVE/stone lands here overwhelmingly.
//	formRadial  — the ORIGINAL golden-angle phyllotaxis + pinned center: phyllotaxis seeds put the
//	    Voronoi boundaries on radial "spokes" and the pinned-center region reads as a hub with a
//	    ring road around it — the wagon wheel. Now ONE option among four, era-weighted toward the
//	    ages that were actually planned around a monument/forum (ancient, medieval), NEVER primitive.
//	formGrid    — seeds on a JITTERED GRID over the disc: rectangular-ish wards, orthogonal-ish
//	    streets — a planned/surveyed town. Weighted toward colonial→modern.
//	formRibbon  — the town ELONGATED along an axis with seeds strung along that axis (a main road)
//	    plus lateral spread: a linear town strung along a road. A pinch of every band, dominant
//	    nowhere; the occasional primitive village that grew along a river/trail.
type tdTownForm int

const (
	formOrganic tdTownForm = iota
	formRadial
	formGrid
	formRibbon
)

// tdFormWeights is a small per-band weighting over the four forms (organic, radial, grid, ribbon),
// consumed as a discrete distribution by tdPickTownForm. A zero weight forbids a form for that
// band (e.g. primitive villages are NEVER radial or grid — they ramble, they are not planned).
// These are the V3-A defaults and are deliberately TUNABLE later (V3-B/C dials each band); the
// only hard contract V3-A tests lock is that PRIMITIVE is organic-dominant and never a wheel.
type tdFormWeights [4]float64

// tdBandFormWeights returns the form distribution for an era band (map-overhaul-citymap):
//
//	organic  (primitive, stone)      — ORGANIC-dominant, a little RIBBON, NEVER radial/grid.
//	hub-spoke(bronze, iron, class.)  — ancient: organic + radial (monument-planned cores appear).
//	castle   (medieval, renaissance) — radial + organic + some grid (market-square towns).
//	zoned    (colonial→victorian)    — GRID-heavy, some organic/ribbon (surveyed colonial towns).
//	blocks   (electric→modern)       — GRID-dominant (the planned modern city).
//	campus   (information→fusion)    — grid / organic mix (megablocks + arcology sprawl).
//	orbital  (space→transcendent)    — organic / grid (radial arcs read as neither wheel nor grid).
//
// Weights are relative (they need not sum to 1). Order: [organic, radial, grid, ribbon].
func tdBandFormWeights(e era) tdFormWeights {
	switch e {
	case eraOrganic:
		// Villages ramble; they are not planned. Organic dominates, ribbon is the rare
		// grew-along-a-trail village, and radial/grid are FORBIDDEN (0) so a primitive town can
		// never roll a wheel or a survey grid.
		return tdFormWeights{0.80, 0, 0, 0.20}
	case eraHubSpoke:
		// Ancient: the first monument/forum cores appear, so radial enters — but the countryside
		// is still mostly organic. A little ribbon; no formal grid yet.
		return tdFormWeights{0.50, 0.35, 0, 0.15}
	case eraCastle:
		// Medieval/renaissance: radial market-square towns + organic old quarters, the first
		// planned grids (bastides), a little ribbon.
		return tdFormWeights{0.32, 0.38, 0.18, 0.12}
	case eraZonedGrid:
		// Colonial→victorian: the surveyed grid takes over; organic survives in old cores, ribbon
		// along the rail/canal, radial is now the exception.
		return tdFormWeights{0.18, 0.10, 0.55, 0.17}
	case eraCityBlocks:
		// Electric→modern: the planned grid dominates the metropolis; a little organic/ribbon.
		return tdFormWeights{0.14, 0.06, 0.64, 0.16}
	case eraCampus:
		// Information→fusion: megablock grid + arcology organic sprawl, ribbon corridors.
		return tdFormWeights{0.30, 0.06, 0.50, 0.14}
	case eraOrbital:
		// Space→transcendent: organic habs + modular grid; the ring/arc look reads as neither a
		// wagon wheel nor a survey grid, so radial stays low.
		return tdFormWeights{0.44, 0.08, 0.40, 0.08}
	default:
		return tdFormWeights{0.80, 0, 0, 0.20}
	}
}

// tdPickTownForm chooses a town's FORM deterministically from (citySeed, era). It is a pure
// function — the SAME (seed, era) always yields the SAME form — and it is ERA-WEIGHTED via
// tdBandFormWeights, so different citySeeds fan out across the era-appropriate forms (no two towns
// need look alike) while PRIMITIVE reliably lands organic (never a wheel). A degenerate all-zero
// or negative weight vector falls back to organic. The seed is hashed with a distinct salt so the
// form roll is independent of the seed's other uses (anchor phase, jitter, scatter phase).
func tdPickTownForm(seed uint32, e era) tdTownForm {
	w := tdBandFormWeights(e)
	total := 0.0
	for _, x := range w {
		if x > 0 {
			total += x
		}
	}
	if total <= 0 {
		return formOrganic
	}
	// A stable [0,total) roll from a salted hash of the seed — independent of the phyllotaxis/
	// jitter hashes so changing one never shifts the form.
	roll := float64(hash2(0xF0F0, 0x0F0F, seed^0x7f4a7c15)) / float64(^uint32(0)) * total
	acc := 0.0
	for i, x := range w {
		if x <= 0 {
			continue
		}
		acc += x
		if roll < acc {
			return tdTownForm(i)
		}
	}
	// Float slop guard: return the last positive-weight form.
	for i := len(w) - 1; i >= 0; i-- {
		if w[i] > 0 {
			return tdTownForm(i)
		}
	}
	return formOrganic
}

// tdTownRadius is the BOUNDED town-disc radius model (city units) for a settlement of `n` fabric
// roof-lots (map-overhaul-citymap; playtest FIX: no pinwheel): townBaseRadius + townGrowth·√n
// over the BANDED count — it grows only ~SQRT with the count, so the footprint saturates, and it
// is a STEP function of the count so the town size is stable within a band. This bounds both the
// raster (the disc the block seeds scatter in) and the fill-frame fit; the anti-pinwheel test
// asserts the real footprint stays within it. A floor keeps a tiny hamlet honest.
func tdTownRadius(n int, cfg tdConfig) float64 {
	nb := tdCountBand(n)
	if nb < 1 {
		nb = 1
	}
	r := cfg.townBaseRadius + cfg.townGrowth*math.Sqrt(float64(nb))
	if r < cfg.townBaseRadius {
		r = cfg.townBaseRadius
	}
	return r
}

// tdPlazaRadius resolves the clear-plaza radius (city units) around a WONDER / center anchor for a
// given town FORM (map-overhaul-citymap de-radialization). The RADIAL and other planned forms keep
// the ROOMY cfg.plazaRadius (a monument/forum-planned town's identity); the ORGANIC form gets a MUCH
// smaller one-ward plaza (cfg.organicPlazaRadius) so a rambling village never clears a dominant
// central ring — the wonder/center keeps a small dressed square + a prominent roof, and the fabric
// fills close to the heart. A missing organic value falls back to the roomy radius (safe). Pure.
func tdPlazaRadius(form tdTownForm, cfg tdConfig) float64 {
	if form == formOrganic && cfg.organicPlazaRadius > 0 {
		return cfg.organicPlazaRadius * cfg.roofSize
	}
	return cfg.plazaRadius * cfg.roofSize
}

// tdOrganicRadiusAt returns the ORGANIC town's in-disc radius at a given angle (city units): the
// base townR bitten INWARD by a smooth, seeded, angle-varying perturbation so the silhouette is a
// genuinely IRREGULAR blob with real bays and peninsulas — NOT a clean radial circle, and NOT the
// timid near-circle the first cut produced (map-overhaul-citymap FIX 2 — ragged outline).
//
// Shape recipe (deterministic per city; the seed is age-independent so a civ's silhouette is stable
// across ages):
//   - A few LOW-frequency lobes (1θ, 2θ, 3θ) dominate → a couple of big bays/peninsulas, the thing
//     that makes the outline unmistakably non-circular. Each lobe gets its OWN seeded phase AND a
//     seeded amplitude weight, so two towns differ in SHAPE, not just rotation (the old version fed
//     the same two frequencies every time, so every silhouette had identical variance).
//   - A little higher-frequency RIPPLE (5θ, 7θ) at a fraction of the amplitude adds fine coastline
//     wobble on top of the big lobes.
//
// The perturbation is always SUBTRACTIVE and normalized so the deepest possible bite is
// organicEdgeAmp: radius ∈ [organicEdgeFloor, 1.0]·townR. Never exceeds townR, so no ward sits past
// the bounded footprint (the compact / anti-pinwheel guarantees + the seed-in-disc bound all hold).
// The floor keeps the footprint STAR-SHAPED about the core (every ray ≥ floor·townR > 0) → always
// simply-connected, so a deep bay can never carve a detached island (streets stay ONE component).
// Non-organic forms never call this (they keep the plain circle). Pure + bounded.
func tdOrganicRadiusAt(angle, townR float64, seed uint32) float64 {
	// Per-seed phase for each harmonic.
	ph := func(k uint32) float64 {
		return float64(hash2(0xED9E, k, seed)) / float64(^uint32(0)) * 2 * math.Pi
	}
	// Per-seed amplitude weight in [0.35,1] for a harmonic, so different towns emphasise different
	// lobes (one town a big single bay, another a three-lobed clover) rather than all sharing one
	// envelope. Keyed off a distinct salt from the phase so amplitude and phase vary independently.
	amp := func(k uint32) float64 {
		return 0.35 + 0.65*float64(hash2(0xA33A, k, seed))/float64(^uint32(0))
	}
	// LOW-frequency lobes (the big bays) carry most of the weight; the higher ripple rides on a
	// fraction. Weights are relative; the whole wave is renormalized to [-1,1] below so the bite
	// depth is controlled purely by organicEdgeAmp regardless of the weight mix.
	w := amp(1)*math.Sin(1*angle+ph(1)) +
		amp(2)*math.Sin(2*angle+ph(2)) +
		amp(3)*0.85*math.Sin(3*angle+ph(3)) +
		amp(5)*0.35*math.Sin(5*angle+ph(5)) +
		amp(7)*0.22*math.Sin(7*angle+ph(7))
	// Normalize by a TYPICAL peak (~2× the RMS of the summed sines), NOT their theoretical all-aligned
	// maximum — dividing by the worst case would make the outline barely swing because five sines
	// almost never peak together (that was the too-timid first cut). Normalizing by the typical peak
	// gives most of the rim a strong swing; the rare angle where the low lobes DO align saturates the
	// bite to the floor and reads as a dramatic cove. wn is clamped to [-1,1] so a saturated bite is
	// exactly organicEdgeAmp deep (never past townR).
	const norm = 1.9
	wn := w / norm
	if wn > 1 {
		wn = 1
	} else if wn < -1 {
		wn = -1
	}
	bite := defaultTdConfig.organicEdgeAmp * (0.5 - 0.5*wn) // 0 (no bite) .. organicEdgeAmp (deepest)
	r := townR * (1 - bite)
	floor := defaultTdConfig.organicEdgeFloor
	if floor <= 0 {
		floor = 0.55
	}
	if minR := floor * townR; r < minR {
		r = minR
	}
	return r
}

// tdInTown reports whether a city-space point lies inside the town footprint for a given FORM.
// ORGANIC uses the irregular blob outline (tdOrganicRadiusAt); every other form uses the plain
// circular disc of radius townR (unchanged). Shared by the raster partition and Lloyd relaxation so
// the town shape is consistent everywhere. Pure.
func tdInTown(x, y, townR float64, form tdTownForm, seed uint32) bool {
	d2 := x*x + y*y
	if form != formOrganic {
		return d2 <= townR*townR
	}
	rr := tdOrganicRadiusAt(math.Atan2(y, x), townR, seed)
	return d2 <= rr*rr
}

// ---- growth anchors (wonder-anchored, central) ------------------------------

// tdAnchorPoints places nWonders growth-anchor centers in city space, spread stably and
// deterministically around the core so more wonders occupy more central regions (FIX 2).
// nWonders==0 yields exactly one anchor AT the core (a cohesive wonderless village). For N≥1
// the anchors sit on a golden-angle phyllotaxis disc whose radius grows with sqrt of the anchor
// index and an overall spread that grows with sqrt(N): a few wonders sit close, many fan out —
// but all kept well INSIDE the town disc so wonders are central, not exiled. Anchor i's offset
// is a pure function of (i, seed) — NOT of N — so adding a wonder never moves the anchors placed
// before it.
func tdAnchorPoints(cx, cy float64, nWonders int, seed uint32, cfg tdConfig) []tdPoint {
	if nWonders <= 0 {
		return []tdPoint{{cx, cy}}
	}
	// A stable per-seed phase so two civs' anchor rings differ, but a given civ's is fixed
	// across ages (seed is age-independent).
	phase := float64(hash2(0xA9C, 0x37, seed)) / float64(^uint32(0)) * 2 * math.Pi
	spread := cfg.anchorSpread * math.Sqrt(float64(nWonders))
	out := make([]tdPoint, 0, nWonders)
	for i := 0; i < nWonders; i++ {
		if i == 0 {
			// The first (grandest) wonder crowns the center.
			out = append(out, tdPoint{cx, cy})
			continue
		}
		// Phyllotaxis: radius ∝ sqrt(i), angle = i*goldenAngle + phase. Normalize the radius by
		// sqrt(nWonders) so the outermost anchor lands ~spread from the core regardless of count.
		r := spread * math.Sqrt(float64(i)) / math.Sqrt(float64(maxInt(nWonders-1, 1)))
		a := float64(i)*goldenAngle + phase
		out = append(out, tdPoint{cx + math.Cos(a)*r, cy + math.Sin(a)*r})
	}
	return out
}

// ---- generate (pure, deterministic, Voronoi blocks) -------------------------

// generateTopPlan synthesizes the whole top-down city plan in CITY SPACE, purely and
// deterministically from seed, as a VORONOI BLOCK / WARD model (map-overhaul-citymap):
//
//	(a) gather   — built buildings → per-type domain/category/tier/count/role.
//	(b) anchors  — the built WONDERS become the central growth anchors (FIX 2); a wonderless
//	    village has a single town-center anchor. Wonders occupy central regions.
//	(c) blocks   — scatter B seeds (√count-scaled, banded), Lloyd-relax them, then RASTER-
//	    partition the town disc into regions. Region boundaries → STREET cells (the gaps);
//	    region interiors → BLOCKS. Central region(s) reserved as the plaza (wonders + center).
//	(d) populate — distribute all non-wonder buildings across the blocks, filling each block's
//	    PERIMETER facing the surrounding streets, intermixing types across blocks. No overlap.
//	(e) wonders  — each wonder's dominant roof at its central anchor; the plaza dressed.
//	(f) filler   — balanced gardens / ponds / trees / props in leftover in-town space.
//	(g) walls    — a wall+gate ring IF the era has walls (primitive: none).
//
// No terrain, no water, no gating: the ground is a neutral era tint.
func generateTopPlan(state game.GameState, byKey map[string]config.BuildingDef, style tdEraStyle, seed uint32) topPlan {
	cfg := defaultTdConfig
	// Route the PER-ERA DENSITY knob: the era preset owns the slot spacing so later eras pack
	// tighter than the airy primitive village. 0 → keep the config default. Primitive maps to
	// 2.4 (unchanged look).
	if style.slotSpacing > 0 {
		cfg.slotSpacing = style.slotSpacing
	}
	plan := topPlan{cx: 0, cy: 0}

	// (a) gather — reuse the pure, sorted gather from citygen.go (domain/category/tier/count/
	// role per distinct built type). Sorted, so placement order is stable.
	blds := gatherBuildings(state, byKey)

	// Split the gathered buildings into the WONDERS (the central anchors + centerpieces) and
	// the rest (the block fabric). Wonders are sorted by prominence so the grandest seats the
	// center anchor.
	var wonders []builtBuilding
	for _, b := range blds {
		if b.category == "wonder" || b.category == "monument" {
			wonders = append(wonders, b)
		}
	}
	sort.SliceStable(wonders, func(i, j int) bool { return moreProminentBld(wonders[i], wonders[j]) })

	// (b) anchors — seat one growth anchor per built wonder (FIX 2), spread stably around the
	// core (kept central). Zero wonders → a single town-center anchor. The grandest wonder
	// crowns anchor 0 (the center); the rest fan out on the seeded phyllotaxis disc.
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

	// Hero promotion (locked #7): the single most prominent LANDMARK is the labeled civic hero.
	// With no civic building AND no wonder, promote the most prominent PRODUCTION building so the
	// city still has exactly one labeled landmark.
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

	// (c) blocks — build the Voronoi block field: town disc → seeds (scattered by the town FORM) →
	// Lloyd → raster partition → street cells + block interiors. Central region(s) are reserved as
	// the plaza (wonders + the wonderless center anchor). The whole field is a pure function of
	// (seed, roof count, form). The FORM is picked once, deterministically + era-weighted, so towns
	// vary per city+era and primitive villages ramble organically instead of all reading as wheels.
	plan.form = tdPickTownForm(seed, eraForAge(state.Age))
	totalRoofs := tdTotalFabricRoofs(blds)
	plan.townR = tdTownRadius(totalRoofs, cfg)
	field := tdBuildBlockField(plan.townR, plan.anchors, totalRoofs, plan.form, cfg, seed)
	plan.streetCells = field.streetCells
	plan.cellSize = field.cellSize
	plan.wardSeeds = field.seeds

	// (d) populate — distribute all non-wonder buildings across the block interiors, filling
	// each block's perimeter facing the surrounding streets, intermixing types across blocks,
	// count-driven, with no roof overlap. Buildings inside a wonder/center plaza region are not
	// placed (those regions stay open).
	tdPopulateBlocks(&plan, field, blds, heroKey, cfg, seed)

	// Final deterministic overlap guard: no two roof lots may sit closer than minGap. It SKIPS
	// (never nudges) a colliding fabric lot, yielding to the fixed wonder roofs and to earlier
	// lots, so a surviving lot keeps its exact placed position. Runs AFTER the wonders exist so
	// it can guard against them.
	tdPlaceWonders(&plan, cfg)
	tdEnforceMinGap(&plan, cfg)

	// (e) town squares — dress each plaza region (wonders + the wonderless center) as a paved
	// town square with a few seeded era props ringed around the roof, so the open center reads
	// as intentional rather than a bare-dirt donut.
	tdPlaceSquares(&plan, style, cfg, seed)

	// (f) filler — balanced gardens / ponds / trees / props in the leftover in-town space.
	tdAddFiller(&plan, field, style, cfg, seed)

	// (g) walls — a wall+gate ring IF the era has walls. Primitive: none.
	if style.hasWalls {
		tdAddWalls(&plan, seed)
	}

	return plan
}

// tdTotalFabricRoofs is the total NON-wonder roof-lot count the fabric will place — the sum
// over non-wonder types of their count-scaled roof-lot counts. Sizes the town disc + the block
// seed count so both scale with the settlement. Pure over blds.
func tdTotalFabricRoofs(blds []builtBuilding) int {
	total := 0
	for _, b := range blds {
		if b.category == "wonder" || b.category == "monument" {
			continue
		}
		total += tdRoofCount(b.count, b.role)
	}
	return total
}

// ---- Voronoi block field (raster nearest-seed) ------------------------------

// blockField is the computed raster partition of the town disc: the seed points (block
// centers), the per-cell nearest-seed assignment, and the derived street cells + block cell
// lists. Everything is on a fixed grid covering [-townR,townR]² at cellSize resolution.
type blockField struct {
	townR    float64
	cellSize float64
	gridN    int       // cells per axis (odd, so a cell is centered on the core)
	origin   float64   // city-space coordinate of cell (0,0)'s CENTER along each axis
	seeds    []tdPoint // relaxed block-seed centers (city space)
	// nearest[c] = seed index owning cell c (row-major, c = gy*gridN+gx); -1 for out-of-disc.
	nearest []int
	// street[c] = true if cell c is a boundary (street) cell.
	street []bool
	// plazaSeed[s] = true if seed s is a reserved PLAZA region (a wonder/center anchor's region).
	plazaSeed []bool
	// blockCells[s] = the interior (non-street) cells of block s, in row-major order.
	blockCells [][]int
	// streetCells are the city-space centers of every street cell (for render + connectivity).
	streetCells []tdPoint
}

// cellCenter returns the city-space center of grid cell (gx,gy).
func (f blockField) cellCenter(gx, gy int) tdPoint {
	return tdPoint{f.origin + float64(gx)*f.cellSize, f.origin + float64(gy)*f.cellSize}
}

// tdBlockSeedCount is the BANDED block-seed count (the stability tradeoff, see the file
// header): B = seedBase + round(growth·√n), a step function of the roof count so the block
// structure is stable within a size band and only re-forms at a band boundary. Plus one seat
// per plaza anchor (each wonder / the center gets its own central region). Capped at a max
// so a metropolis stays a legible ward count. The ORGANIC form uses a FINER growth + higher cap
// (organicSeedGrowth / organicSeedMax) so its mesh has many small wards (no pie-slice wheel);
// every other form keeps the coarser seedGrowth / seedMax.
func tdBlockSeedCount(nRoofs, nAnchors int, form tdTownForm, cfg tdConfig) int {
	growth, seedMax := cfg.seedGrowth, cfg.seedMax
	if form == formOrganic && cfg.organicSeedGrowth > 0 {
		growth = cfg.organicSeedGrowth
		if cfg.organicSeedMax > 0 {
			seedMax = cfg.organicSeedMax
		}
	}
	// Band the count so the ward count re-forms only at band boundaries, in lockstep with the
	// banded town radius — the whole block field is a step function of the count.
	nb := tdCountBand(nRoofs)
	b := cfg.seedBase + int(growth*math.Sqrt(float64(nb))+0.5)
	if b < cfg.seedBase {
		b = cfg.seedBase
	}
	if b > seedMax {
		b = seedMax
	}
	// Guarantee at least one region per plaza anchor plus a couple of building wards, so a
	// wonder-heavy small civ still has interior blocks to fill.
	if min := nAnchors + 2; b < min {
		b = min
	}
	if b > seedMax+nAnchors {
		b = seedMax + nAnchors
	}
	return b
}

// tdPinnedCount returns how many of a town's anchors are PINNED as fixed Voronoi seeds through the
// scatter + Lloyd relaxation (map-overhaul-citymap de-radialization). Pinned anchors own a fixed
// central region so the streets always reach them and a WONDER ward stays put.
//
// Every non-organic form pins ALL its anchors (including the wonderless city-center) — the RADIAL
// form WANTS that pinned center as a hub with a ring road (its identity). The ORGANIC form pins only
// its WONDER anchors: a wonderless village's bare center anchor is NOT seeded at dead-center, because
// a lone seed nailed to the middle owns a round central region whose boundaries to every neighbour
// RADIATE outward as spokes — a wagon wheel — no matter how the rest is scattered. Dropping that pin
// lets the organic blue-noise seeds cover the center as ORDINARY mesh wards (none exactly central),
// so the street web through the middle is an irregular MESH with loops, not a hub-and-spoke. The
// logical town center (plan.cx,cy) is unchanged — only the block SEED at the origin goes away.
func tdPinnedCount(form tdTownForm, anchors []tdAnchor) int {
	if form != formOrganic {
		return len(anchors)
	}
	n := 0
	for _, a := range anchors {
		if a.wonder {
			n++
		}
	}
	return n
}

// tdScatterSeedsFor places B block seeds in the town area for a given town FORM, DETERMINISTICALLY
// from citySeed (map-overhaul-citymap V3-A). It pins the first tdPinnedCount(form,anchors) seeds to
// the anchors that own a fixed central region (so each wonder / a planned-form center owns a region
// and the streets reach it), then scatters the remaining FREE seeds by the form's own strategy. The
// seed COUNT is the same regardless of form (it is the banded tdBlockSeedCount), so banded stability
// is unaffected; only the arrangement + which anchors pin differs. Every form's free seeds land
// inside the town so the raster partition stays one connected boundary web (streets-connected holds
// by construction). Pure function of (form, seed, B, anchors).
func tdScatterSeedsFor(form tdTownForm, townR float64, anchors []tdAnchor, B int, cfg tdConfig, seed uint32) []tdPoint {
	nPinned := tdPinnedCount(form, anchors)
	seeds := make([]tdPoint, 0, B)
	for i := 0; i < nPinned && i < len(anchors); i++ {
		seeds = append(seeds, tdPoint{anchors[i].cx, anchors[i].cy})
	}
	need := B - len(seeds)
	if need <= 0 {
		return seeds
	}
	switch form {
	case formRadial:
		return tdScatterRadial(seeds, townR, need, cfg, seed)
	case formGrid:
		return tdScatterGrid(seeds, townR, need, cfg, seed)
	case formRibbon:
		return tdScatterRibbon(seeds, townR, need, cfg, seed)
	default: // formOrganic
		return tdScatterOrganic(seeds, townR, need, cfg, seed)
	}
}

// tdScatterRadial is the ORIGINAL golden-angle phyllotaxis scatter — now the formRadial strategy
// ONLY (map-overhaul-citymap). Free seeds fill the disc on a golden-angle spiral: this is exactly
// what puts the Voronoi ward BOUNDARIES on radial spokes and, with the pinned-center region acting
// as a hub, reads as a wagon wheel with a ring road. That look is intentional for the radial form
// (a monument/forum-planned town) but is no longer the default for every town. `seeds` already
// holds the pinned anchors; this appends the free seeds. Pure over (seed, need).
func tdScatterRadial(seeds []tdPoint, townR float64, need int, cfg tdConfig, seed uint32) []tdPoint {
	phase := float64(hash2(0x5eed, 0x1a, seed)) / float64(^uint32(0)) * 2 * math.Pi
	// The spiral fills up to ~0.92·townR so the outermost blocks sit inside the disc rim (a
	// ring of ground/greenery hugs the edge).
	for k := 0; k < need; k++ {
		frac := (float64(k) + 0.5) / float64(maxInt(need, 1))
		r := 0.92 * townR * math.Sqrt(frac)
		// Nudge the innermost few outward so they don't collide with the pinned center anchor.
		if r < cfg.townBaseRadius*0.35 {
			r = cfg.townBaseRadius*0.35 + r*0.5
		}
		ang := float64(len(seeds))*goldenAngle + phase
		seeds = append(seeds, tdPoint{math.Cos(ang) * r, math.Sin(ang) * r})
	}
	return seeds
}

// tdScatterOrganic is the POISSON-DISK / jittered-random scatter — the formOrganic strategy and
// the DEFAULT for primitive villages (map-overhaul-citymap). It drops free seeds at seeded RANDOM
// positions UNIFORMLY over the WHOLE town blob INCLUDING THE CENTER (uniform by area: radius ∝ √u,
// angle uniform) with a blue-noise REJECT pass — a candidate is rejected only if it lands within
// minDist of another FREE seed — so the wards come out irregular and RAMBLING with organic streets
// and NO radial bias and NO forced ring/spokes.
//
// DE-RADIALIZATION (map-overhaul-citymap): the reject pass DELIBERATELY IGNORES the pinned anchor
// seeds (indices < nPinned). The old scatter repelled free seeds away from the pinned center anchor,
// carving a central ANNULUS/VOID with the seeds ringed around the middle — which made the Voronoi
// boundaries radiate as spokes (a wagon wheel). Ignoring the anchors lets free seeds pack right up to
// and AROUND the center, so the center is just more mesh (many small wards near the heart), never a
// lone hub encircled by a ring. Candidates are also clamped to the ORGANIC BLOB outline
// (tdOrganicRadiusAt) so the silhouette rambles, not a clean disc.
//
// Determinism + a fixed seed COUNT: if the reject budget can't place all `need` seeds (dense packing
// is unlucky), the shortfall is filled by RELAXING the spacing (tail candidates accepted regardless),
// so exactly `need` free seeds always land and the banded stability tradeoff is preserved. Pure over
// (seed, need); bounded attempt loop → panic-safe.
func tdScatterOrganic(seeds []tdPoint, townR float64, need int, cfg tdConfig, seed uint32) []tdPoint {
	nPinned := len(seeds) // the anchor seeds already appended by tdScatterSeedsFor
	rr := 0.94 * townR    // free seeds stay just inside the (blob) rim
	// Target spacing from area: minDist ≈ 0.7·√(discArea / seeds). The 0.7 leaves the reject pass
	// room to actually place the target count before it must relax.
	area := math.Pi * rr * rr
	minDist := 0.7 * math.Sqrt(area/float64(need+len(seeds)))
	minD2 := minDist * minDist
	r := newRNG(hash2(0x0B10, uint32(need), seed) | 1)
	// sample draws one uniform-in-blob candidate: uniform by area, then clamped to the organic blob
	// outline so a candidate never lands past the rambling edge.
	sample := func() tdPoint {
		rad := rr * math.Sqrt(r.f01())
		ang := r.f01() * 2 * math.Pi
		x, y := math.Cos(ang)*rad, math.Sin(ang)*rad
		if edge := 0.94 * tdOrganicRadiusAt(ang, townR, seed); rad > edge {
			s := edge / rad
			x *= s
			y *= s
		}
		return tdPoint{x, y}
	}
	placed := 0
	// A generous, BOUNDED attempt budget: up to 40 candidates per needed seed.
	maxAttempts := (need + 1) * 40
	for attempts := 0; placed < need && attempts < maxAttempts; attempts++ {
		p := sample()
		ok := true
		// Repel against the already-placed FREE seeds only (skip the pinned anchors, so the center
		// fills instead of ringing around the anchor).
		for si := nPinned; si < len(seeds); si++ {
			dx, dy := p.x-seeds[si].x, p.y-seeds[si].y
			if dx*dx+dy*dy < minD2 {
				ok = false
				break
			}
		}
		if ok {
			seeds = append(seeds, p)
			placed++
		}
	}
	// Relax to guarantee exactly `need` free seeds (fixed count → stable banding). Any remaining
	// slots take unrejected uniform samples; they may sit a touch closer, which Lloyd then evens.
	for placed < need {
		seeds = append(seeds, sample())
		placed++
	}
	return seeds
}

// tdScatterGrid is the JITTERED-GRID scatter — the formGrid strategy (map-overhaul-citymap). It
// lays free seeds on a square lattice across the town's bounding square, keeps the ones inside the
// disc, and JITTERS each off its lattice node by a fraction of the cell — yielding rectangular-ish
// wards and orthogonal-ish streets (a surveyed/planned town), NOT a wheel. The lattice is sized so
// it yields at least `need` in-disc nodes; nodes are taken CENTERS-OUT (nearest the core first) so
// a partial fill stays central and the count is stable. A seeded phase rotates/offsets the whole
// lattice so two grid towns aren't identical. If the disc is somehow too small to host `need`
// nodes, the shortfall falls through to an organic top-up so exactly `need` free seeds land. Pure
// over (seed, need); bounded → panic-safe.
func tdScatterGrid(seeds []tdPoint, townR float64, need int, cfg tdConfig, seed uint32) []tdPoint {
	rr := 0.92 * townR
	// Choose a cell size so ~need nodes fall inside the disc: discArea ≈ need·cell² → cell ≈
	// √(π·rr² / need). A mild 1.08 loosening keeps a few spare nodes for the centers-out trim.
	cell := 1.08 * math.Sqrt(math.Pi*rr*rr/float64(need))
	if cell < 1e-3 {
		cell = 1e-3
	}
	// A seeded sub-cell phase offset so the grid origin (and thus the whole lattice) shifts per
	// city; kept within one cell so the lattice still spans the disc.
	phx := (float64(hash2(0x671D, 0x11, seed))/float64(^uint32(0)) - 0.5) * cell
	phy := (float64(hash2(0x671D, 0x12, seed))/float64(^uint32(0)) - 0.5) * cell
	// Enumerate lattice nodes across the bounding square, keep the in-disc ones with jitter, and
	// order them by distance from the core so a trim to `need` keeps the central lattice.
	type gnode struct {
		p  tdPoint
		d2 float64
	}
	var nodes []gnode
	half := int(math.Ceil(rr/cell)) + 1
	if half > 64 { // panic-safety cap on the lattice extent
		half = 64
	}
	r := newRNG(hash2(0x6D17, uint32(need), seed) | 1)
	jit := cell * 0.30 // jitter amplitude — enough to look natural, not enough to cross a lane
	for gy := -half; gy <= half; gy++ {
		for gx := -half; gx <= half; gx++ {
			nx := float64(gx)*cell + phx
			ny := float64(gy)*cell + phy
			// Jitter each node deterministically (hash keyed by lattice coords, stable per city).
			jx := (float64(hash2(uint32(gx*2+1000), uint32(gy*2+1000), seed))/float64(^uint32(0)) - 0.5) * 2 * jit
			jy := (float64(hash2(uint32(gx*2+2000), uint32(gy*2+2000), seed))/float64(^uint32(0)) - 0.5) * 2 * jit
			px, py := nx+jx, ny+jy
			d2 := px*px + py*py
			if d2 > rr*rr {
				continue
			}
			nodes = append(nodes, gnode{p: tdPoint{px, py}, d2: d2})
		}
	}
	sort.SliceStable(nodes, func(a, b int) bool { return nodes[a].d2 < nodes[b].d2 })
	for i := 0; i < need && i < len(nodes); i++ {
		seeds = append(seeds, nodes[i].p)
	}
	// Top up (rare: only if the disc couldn't host `need` in-disc nodes) so the free-seed count
	// stays exactly `need` and banded stability is preserved.
	for placed := minInt(need, len(nodes)); placed < need; placed++ {
		rad := rr * math.Sqrt(r.f01())
		ang := r.f01() * 2 * math.Pi
		seeds = append(seeds, tdPoint{math.Cos(ang) * rad, math.Sin(ang) * rad})
	}
	return seeds
}

// tdScatterRibbon is the ELONGATED-AXIS scatter — the formRibbon strategy (map-overhaul-citymap).
// The town is stretched along a seeded axis (a main road): free seeds are strung ALONG that axis
// across the disc with only a small LATERAL spread, so the Voronoi wards line up as a linear town
// strung along a road, NOT a wheel. Position-along-axis is spread evenly (index-stable, centers-
// out) with a little seeded wobble; the lateral offset is a small seeded fraction of the disc so
// the ribbon has width without becoming a blob. Every seed is clamped inside the disc so the raster
// stays connected. The pinned center anchor sits on the axis. Pure over (seed, need).
func tdScatterRibbon(seeds []tdPoint, townR float64, need int, cfg tdConfig, seed uint32) []tdPoint {
	// Axis direction: a seeded angle, so different ribbon towns run different ways.
	axis := float64(hash2(0x21B0, 0x33, seed)) / float64(^uint32(0)) * 2 * math.Pi
	ax, ay := math.Cos(axis), math.Sin(axis)
	// Perpendicular (lateral) direction.
	px, py := -ay, ax
	alongMax := 0.90 * townR // reach most of the disc along the road
	latMax := 0.28 * townR   // a modest lateral spread → a road, not a blob
	r := newRNG(hash2(0x21B1, uint32(need), seed) | 1)
	for k := 0; k < need; k++ {
		// Even, centers-out spacing along the axis in [-alongMax, alongMax], with a small wobble.
		frac := (float64(k)+0.5)/float64(need)*2 - 1 // -1..1
		wobble := (r.f01() - 0.5) * (2 * alongMax / float64(need)) * 0.8
		along := frac*alongMax + wobble
		lat := (r.f01()*2 - 1) * latMax
		x := ax*along + px*lat
		y := ay*along + py*lat
		// Clamp inside the disc so the raster partition stays whole (streets connected).
		if d := math.Hypot(x, y); d > 0.95*townR {
			s := 0.95 * townR / d
			x *= s
			y *= s
		}
		seeds = append(seeds, tdPoint{x, y})
	}
	return seeds
}

// tdBuildBlockField builds the whole Voronoi block field: it sizes the raster to the town disc,
// scatters the block seeds by the town FORM (organic / radial / grid / ribbon), Lloyd-relaxes
// them, RASTER-partitions the disc by nearest seed, then derives the street (boundary) cells and
// per-block interior cell lists. Central seeds (the plaza anchors) are flagged so their regions
// stay OPEN and stay pinned through relaxation. Pure + deterministic; panic-safe (the grid is
// capped and every loop is bounded).
func tdBuildBlockField(townR float64, anchors []tdAnchor, nRoofs int, form tdTownForm, cfg tdConfig, seed uint32) blockField {
	f := blockField{townR: townR, cellSize: cfg.cellSize}
	if townR <= 0 || cfg.cellSize <= 0 {
		return f
	}
	// Grid covering [-townR,townR]² at cellSize resolution, made ODD so a cell centers on the
	// core. Capped for panic-safety on a pathological config.
	gridN := int(math.Ceil(2*townR/cfg.cellSize)) + 1
	if gridN%2 == 0 {
		gridN++
	}
	if gridN < 3 {
		gridN = 3
	}
	if gridN > 129 {
		gridN = 129
	}
	f.gridN = gridN
	half := float64(gridN-1) / 2
	f.origin = -half * cfg.cellSize // cell (0,0) center is the disc's lower-left

	// Seeds: B block centers, first len(anchors) pinned to the plaza anchors, the free seeds
	// scattered by the town's FORM (organic / radial / grid / ribbon). Only the arrangement varies
	// by form; the COUNT is the banded tdBlockSeedCount, so banded stability is unaffected.
	B := tdBlockSeedCount(nRoofs, len(anchors), form, cfg)
	seeds := tdScatterSeedsFor(form, townR, anchors, B, cfg, seed)
	// Lloyd relaxation: a few passes move each seed toward its region's centroid for even,
	// organic block sizes. The PINNED anchor seeds are held fixed (their central regions must
	// stay put); only the free building-ward seeds relax.
	//
	// Pinned anchor seeds are held fixed through relaxation so the streets reach the central plaza
	// and WONDER wards stay put. tdPinnedCount decides how many: every planned form pins all its
	// anchors (the RADIAL hub-and-ring center is its identity), but the ORGANIC form does NOT pin a
	// wonderless center — a lone dead-center seed is the wagon-wheel HUB whose region boundaries
	// radiate as spokes. Without that pin, organic's blue-noise seeds cover the center as ordinary
	// mesh wards, so the middle is an irregular web with loops, not a hub with spokes.
	nPinned := tdPinnedCount(form, anchors)
	seeds = tdLloyd(seeds, nPinned, gridN, cfg.cellSize, f.origin, townR, form, seed, cfg.lloydPasses)
	f.seeds = seeds
	// A seed's region is a reserved PLAZA (building-free) ONLY when its anchor seats a WONDER —
	// the wonder occupies a whole central region kept clear + dressed as a square. The wonderless
	// town-center anchor's seed is a NORMAL building ward (its modest square is a small paved lot
	// at the seed, with the fabric filling the ward around it), so a hut village keeps a FILLED
	// heart and gets full block capacity for the near-1:1 low band.
	// Mark plaza (building-free) seeds by WONDER anchor directly — decoupled from nPinned, which for
	// organic no longer equals the anchor count (a wonderless center relaxes but is never a plaza).
	f.plazaSeed = make([]bool, len(seeds))
	for i := 0; i < len(anchors) && i < len(seeds); i++ {
		if anchors[i].wonder {
			f.plazaSeed[i] = true
		}
	}

	// Raster partition: nearest seed per in-town cell. The town footprint is the plain circular disc
	// for every form EXCEPT organic, which uses the irregular BLOB outline (tdInTown) so the
	// silhouette rambles rather than reading as a clean radial disc.
	f.nearest = make([]int, gridN*gridN)
	f.street = make([]bool, gridN*gridN)
	for gy := 0; gy < gridN; gy++ {
		for gx := 0; gx < gridN; gx++ {
			c := gy*gridN + gx
			p := f.cellCenter(gx, gy)
			if !tdInTown(p.x, p.y, townR, form, seed) {
				f.nearest[c] = -1 // outside the town → not built
				continue
			}
			best, second := math.Inf(1), math.Inf(1)
			bi := -1
			for si, s := range seeds {
				d := (p.x-s.x)*(p.x-s.x) + (p.y-s.y)*(p.y-s.y)
				if d < best {
					second = best
					best = d
					bi = si
				} else if d < second {
					second = d
				}
			}
			f.nearest[c] = bi
			// STREET cell: nearest and second-nearest seed ~equidistant (within the boundary
			// band). Compare the DISTANCE difference (not squared) so the band is a consistent
			// width in city units regardless of how far the cell sits from its seeds.
			if bi >= 0 && !math.IsInf(second, 1) {
				diff := math.Sqrt(second) - math.Sqrt(best)
				if diff < cfg.streetBand {
					f.street[c] = true
				}
			}
		}
	}

	// Derive street-cell centers + per-block interior cell lists.
	f.blockCells = make([][]int, len(seeds))
	for gy := 0; gy < gridN; gy++ {
		for gx := 0; gx < gridN; gx++ {
			c := gy*gridN + gx
			si := f.nearest[c]
			if si < 0 {
				continue
			}
			if f.street[c] {
				f.streetCells = append(f.streetCells, f.cellCenter(gx, gy))
				continue
			}
			// Interior cell of block si (skip plaza regions — they hold no building cells).
			if !f.plazaSeed[si] {
				f.blockCells[si] = append(f.blockCells[si], c)
			}
		}
	}
	return f
}

// tdLloyd runs `passes` Lloyd relaxation iterations over the seeds: each pass RASTER-partitions
// the town by nearest seed and moves every FREE seed (index >= nPinned) to the centroid of its
// region; pinned seeds (the plaza anchors) stay fixed. Evens out the block sizes into an
// organic, roughly-even field. The town footprint matches the raster partition's — the plain disc
// for every form except organic, which uses the irregular BLOB outline (tdInTown) — so relaxed
// centroids stay inside the same rambling town shape. Pure + deterministic; a seed that loses its
// whole region (rare) keeps its position. Panic-safe: bounded loops, guarded division.
func tdLloyd(seeds []tdPoint, nPinned, gridN int, cellSize, origin, townR float64, form tdTownForm, seed uint32, passes int) []tdPoint {
	if len(seeds) == 0 || gridN < 3 || passes <= 0 {
		return seeds
	}
	cur := make([]tdPoint, len(seeds))
	copy(cur, seeds)
	for pass := 0; pass < passes; pass++ {
		sumX := make([]float64, len(cur))
		sumY := make([]float64, len(cur))
		cnt := make([]int, len(cur))
		for gy := 0; gy < gridN; gy++ {
			for gx := 0; gx < gridN; gx++ {
				px := origin + float64(gx)*cellSize
				py := origin + float64(gy)*cellSize
				if !tdInTown(px, py, townR, form, seed) {
					continue
				}
				best := math.Inf(1)
				bi := -1
				for si, s := range cur {
					d := (px-s.x)*(px-s.x) + (py-s.y)*(py-s.y)
					if d < best {
						best = d
						bi = si
					}
				}
				if bi >= 0 {
					sumX[bi] += px
					sumY[bi] += py
					cnt[bi]++
				}
			}
		}
		for si := nPinned; si < len(cur); si++ {
			if cnt[si] > 0 {
				cur[si] = tdPoint{sumX[si] / float64(cnt[si]), sumY[si] / float64(cnt[si])}
			}
		}
	}
	return cur
}

// ---- populate the blocks (buildings face the surrounding streets) -----------

// tdPopulateBlocks distributes ALL non-wonder buildings across the Voronoi block interiors
// (map-overhaul-citymap step d). Each block is filled by walking its PERIMETER ring (the
// interior cells adjacent to a street) and dropping a roof at inset positions facing the
// surrounding street, so buildings ring the block's edge like real ward frontage; the block's
// deeper interior is left for filler. Types are INTERMIXED across blocks (a round-robin over the
// sorted-gather types assigns consecutive buildings to successive blocks), count-driven. Roofs
// that would land inside a wonder/center plaza region, on a street cell, or overlapping an
// already-placed roof are skipped (no overlap).
//
// Stability (best-effort, the banded tradeoff): within a size band the block field is fixed, so
// buildings fill a stable set of blocks by a per-block cursor; adding a building appends into an
// open perimeter slot without disturbing the ones already placed. New TYPES / a band boundary /
// an age-up re-form the field (a layout event, as the old model also had for a new building type).
func tdPopulateBlocks(plan *topPlan, field blockField, blds []builtBuilding, heroKey string, cfg tdConfig, seed uint32) {
	// The fabric types, in the stable sorted-gather order, each with a stable roof/size + count.
	type fabType struct {
		b     builtBuilding
		roof  roofType
		sz    float64
		n     int
		label bool
	}
	var types []fabType
	for _, b := range blds {
		if b.category == "wonder" || b.category == "monument" {
			continue // wonders are central anchors/centerpieces, not block fabric
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
		types = append(types, fabType{b: b, roof: rt, sz: sz, n: n,
			label: b.key == heroKey && !plan.hasHero})
	}
	if len(types) == 0 {
		return
	}

	// Build the ordered per-block SLOT lists: for every non-plaza block, the roof positions
	// inside it (perimeter-first, facing the surrounding streets). Blocks with any slots are the
	// ones we distribute buildings into.
	slotsPerBlock := tdBlockSlots(field, cfg)
	var blockOrder []int // block indices that actually have slots, in a stable order
	for si := range slotsPerBlock {
		if len(slotsPerBlock[si]) > 0 {
			blockOrder = append(blockOrder, si)
		}
	}
	if len(blockOrder) == 0 {
		return
	}
	// Sort blocks by distance of their seed from the core (inner blocks first) so the town fills
	// heart-outward and low counts sit centrally. Stable.
	sort.SliceStable(blockOrder, func(a, b int) bool {
		sa, sb := field.seeds[blockOrder[a]], field.seeds[blockOrder[b]]
		da := sa.x*sa.x + sa.y*sa.y
		db := sb.x*sb.x + sb.y*sb.y
		if da != db {
			return da < db
		}
		return blockOrder[a] < blockOrder[b]
	})

	// A per-block cursor into its slot list, so each building placed in a block takes the next
	// free frontage slot. INTERMIX across blocks: build one flat queue of building instances in
	// round-robin type order (consecutive instances are different domains), then deal them onto
	// the blocks round-robin, so a block ends up with a MIX of types and consecutive placements
	// differ.
	type inst struct {
		roof     roofType
		sz       float64
		domain   string
		category string
		tier     int
		name     string
		label    bool
	}
	maxN := 0
	for _, ft := range types {
		if ft.n > maxN {
			maxN = ft.n
		}
	}
	var queue []inst
	for j := 0; j < maxN; j++ {
		for r := 0; r < len(types); r++ {
			ft := types[r]
			if j >= ft.n {
				continue
			}
			queue = append(queue, inst{
				roof: ft.roof, sz: ft.sz, domain: ft.b.domain, category: ft.b.category,
				tier: ft.b.tier, name: ft.b.name, label: ft.label && j == 0,
			})
		}
	}

	plazaR := tdPlazaRadius(plan.form, cfg)   // form-aware: organic keeps a small one-ward plaza
	cursor := make([]int, len(slotsPerBlock)) // next free slot per block
	placed := make([]tdLot, 0, len(queue))

	// Deal the queue onto the blocks round-robin. Each instance goes to the next block (in
	// blockOrder) that still has a free frontage slot; a block that runs out is skipped.
	// Deal each queued building onto the next block in round-robin order, taking that block's next
	// free slot. If the slot is inside a wonder plaza or would overlap an already-placed roof, the
	// building RETRIES the next block/slot (it is NOT dropped) — so the near-1:1 low-count band is
	// honoured while no-overlap holds. A building is only dropped if EVERY remaining slot fails,
	// which the ample per-block capacity avoids.
	bi := 0
	guard := 0
	// Total slots is the hard cap on attempts; a generous multiple lets a building probe several
	// blocks before it lands.
	totalSlots := 0
	for _, b := range blockOrder {
		totalSlots += len(slotsPerBlock[b])
	}
	maxGuard := (totalSlots + len(queue)) * 4
	nextFreeBlock := func() (int, blockSlot, bool) {
		for tries := 0; tries < len(blockOrder); tries++ {
			blk := blockOrder[bi%len(blockOrder)]
			bi++
			if cursor[blk] < len(slotsPerBlock[blk]) {
				sl := slotsPerBlock[blk][cursor[blk]]
				cursor[blk]++
				return blk, sl, true
			}
		}
		return 0, blockSlot{}, false // every block full
	}
	for qi := 0; qi < len(queue); qi++ {
		it := queue[qi]
		landed := false
		for guard < maxGuard {
			guard++
			blk, slot, ok := nextFreeBlock()
			if !ok {
				break // all blocks full → the rest of the queue can't place
			}
			// A whisper of seeded jitter so the frontage isn't a perfect stamp; the inset already
			// holds the base slot off the street, and the collision check below is post-jitter.
			jx, jy := slotJitter(qi, blk, seed, cfg.jitterAmp)
			x := slot.x + jx
			y := slot.y + jy
			half := it.sz / 2
			if it.roof == roofLong {
				half = it.sz * 1.8 / 2
			}
			if insideWonderPlaza(x, y, plan.anchors, plazaR) {
				continue // retry another slot
			}
			collides := false
			for _, p := range placed {
				ph := math.Max(p.w, p.h) / 2
				if math.Hypot(x-p.x, y-p.y) < half+ph+cfg.minGap {
					collides = true
					break
				}
			}
			if collides {
				continue // retry another slot
			}
			lot := tdLot{
				x: x, y: y, w: it.sz, h: it.sz, kind: tdRoof,
				domain: it.domain, category: it.category, tier: it.tier, roof: it.roof,
			}
			if it.roof == roofLong {
				lot.w = it.sz * 1.8 // longhouses/rowhouses are elongated
			}
			if it.label {
				lot.label = it.name
				lot.prom = prominenceOfLot(it.tier)
			}
			placed = append(placed, lot)
			landed = true
			break
		}
		if !landed {
			// No slot took this building (all full or all collided) — stop; the remaining queue
			// can't place either. Fill-frame + count-scaling keep the town legible regardless.
			break
		}
	}
	plan.lots = append(plan.lots, placed...)
}

// blockSlot is one candidate roof position inside a block, in city space.
type blockSlot = tdPoint

// tdBlockSlots computes, for every non-plaza block, an ordered list of roof slots inside the
// block. Buildings should ring the block PERIMETER facing the surrounding streets, so slots are
// ordered PERIMETER-FIRST (by ring depth from the boundary), and within a ring by angle around
// the seed so a partly-filled block fills evenly. A perimeter cell's slot is pushed INWARD (from
// the street toward the seed) so the roof body sits fully inside the block facing the gap; deeper
// cells sit at their center. Emitting ALL interior cells (not just the perimeter) guarantees the
// block can absorb the near-1:1 low-count band — small counts fill the perimeter and read as
// frontage, and only very dense fills spill into the block core. Near-duplicate slots are thinned
// so roofs never stack. Pure + deterministic over the field.
func tdBlockSlots(field blockField, cfg tdConfig) [][]blockSlot {
	out := make([][]blockSlot, len(field.seeds))
	if field.gridN < 3 {
		return out
	}
	gridN := field.gridN
	isOpen := func(gx, gy int) bool { // a street or off-disc cell (a block edge faces it)
		if gx < 0 || gx >= gridN || gy < 0 || gy >= gridN {
			return true
		}
		c := gy*gridN + gx
		return field.nearest[c] < 0 || field.street[c]
	}
	// A slot carries its ring DEPTH (0 = perimeter, touching a street) so we can order
	// perimeter-first per block.
	type ringSlot struct {
		p     tdPoint
		depth int
	}
	perGap := cfg.streetBand/2 + cfg.blockInset + cfg.roofSize/2 // inward push for a perimeter cell
	raw := make([][]ringSlot, len(field.seeds))
	for gy := 0; gy < gridN; gy++ {
		for gx := 0; gx < gridN; gx++ {
			c := gy*gridN + gx
			si := field.nearest[c]
			if si < 0 || field.street[c] || field.plazaSeed[si] {
				continue
			}
			perim := false
			for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
				if isOpen(gx+d[0], gy+d[1]) {
					perim = true
					break
				}
			}
			p := field.cellCenter(gx, gy)
			depth := 1
			if perim {
				depth = 0
				// Push a perimeter slot inward, from the cell center toward the block seed, so the
				// roof faces the surrounding street from inside the block.
				s := field.seeds[si]
				dx, dy := s.x-p.x, s.y-p.y
				l := math.Hypot(dx, dy)
				if l < 1e-6 {
					dx, dy, l = 0.7071, 0.7071, 1
				}
				p = tdPoint{p.x + dx/l*perGap, p.y + dy/l*perGap}
			}
			raw[si] = append(raw[si], ringSlot{p: p, depth: depth})
		}
	}
	// Order each block's slots perimeter-first, then by angle around the seed; thin duplicates.
	for si := range raw {
		s := field.seeds[si]
		slots := raw[si]
		sort.SliceStable(slots, func(a, b int) bool {
			if slots[a].depth != slots[b].depth {
				return slots[a].depth < slots[b].depth // perimeter (0) before deep (1)
			}
			aa := math.Atan2(slots[a].p.y-s.y, slots[a].p.x-s.x)
			ab := math.Atan2(slots[b].p.y-s.y, slots[b].p.x-s.x)
			return aa < ab
		})
		// Thin so two kept slots in a block are at least the roof COLLISION distance apart
		// (roofSize + minGap) — then same-block slots never self-collide in the deal, so the
		// perimeter frontage packs cleanly without wasted retries.
		step := cfg.roofSize + cfg.minGap
		var kept []blockSlot
		for _, sl := range slots {
			ok := true
			for _, k := range kept {
				if math.Hypot(sl.p.x-k.x, sl.p.y-k.y) < step {
					ok = false
					break
				}
			}
			if ok {
				kept = append(kept, sl.p)
			}
		}
		out[si] = kept
	}
	return out
}

// insideWonderPlaza reports whether (x,y) falls within the clear plaza radius of any WONDER
// anchor (FIX 2). The bare city-center anchor of a wonderless village is NOT a wonder anchor,
// so a tiny village keeps a filled center — only a real wonder clears a big plaza around itself
// (the modest center square is handled separately in tdPlaceSquares).
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

// tdEnforceMinGap is the final deterministic overlap guard: no two roof lots may sit closer
// than the min gap. It SKIPS (never nudges) a colliding lot, and a lot only ever yields to lots
// EARLIER in the slice (the fixed wonder roofs are placed last, so they win against fabric via
// the pre-pass below), so a surviving lot keeps its exact placed position. In a well-formed
// Voronoi plan the block insets already prevent overlaps, so this is a safety net, not a mover.
func tdEnforceMinGap(plan *topPlan, cfg tdConfig) {
	gap := cfg.minGap
	if gap <= 0 {
		return
	}
	overlaps := func(a, b tdLot) bool {
		ah := math.Max(a.w, a.h) / 2
		bh := math.Max(b.w, b.h) / 2
		return math.Hypot(a.x-b.x, a.y-b.y) < ah+bh+gap
	}
	// Pre-pass: the wonder roofs are the FIXED obstacles (their positions are a function of
	// (anchor, seed), independent of any building count), so gather them first and check every
	// fabric lot against them regardless of slice order.
	var wonders []tdLot
	for _, lt := range plan.lots {
		if lt.kind == tdRoof && lt.roof == roofWonder {
			wonders = append(wonders, lt)
		}
	}
	kept := make([]tdLot, 0, len(plan.lots))
	for _, lt := range plan.lots {
		if lt.kind != tdRoof || lt.roof == roofWonder {
			kept = append(kept, lt) // non-roofs and the wonders themselves always stay
			continue
		}
		skip := false
		for _, w := range wonders {
			if overlaps(lt, w) {
				skip = true
				break
			}
		}
		if !skip {
			for _, k := range kept {
				if k.kind == tdRoof && k.roof != roofWonder && overlaps(lt, k) {
					skip = true
					break
				}
			}
		}
		if skip {
			continue
		}
		kept = append(kept, lt)
	}
	plan.lots = kept
}

// ---- wonders (central centerpieces) -----------------------------------------

// tdWonderScale is the wonder roof extent as a multiple of roofSize, by anchor index: the
// grandest wonder (anchor 0) is the largest, the rest a touch smaller so the centre reads as
// the primary showpiece. Kept well UNDER the plaza radius so a paved ring always shows around
// the roof inside its cleared plaza. Single source of truth shared by tdPlaceWonders (the roof)
// and tdPlaceSquares (the prop ring + apron).
func tdWonderScale(anchorIdx int) float64 {
	if anchorIdx == 0 {
		return 2.6
	}
	return 2.2
}

// tdPlaceWonders drops each wonder's dominant, ornate centerpiece roof AT its central anchor
// (locked #13, FIX 2). The grandest wonder crowns the center anchor; the rest sit at their
// spread central anchors. Each is labeled and drawn prominent; its plaza region was kept OPEN by
// the block field (a plaza seed holds no building cells), so the fabric never buries it. Only
// the grandest wonder carries the "City Center" hero label priority via its prominence.
func tdPlaceWonders(plan *topPlan, cfg tdConfig) {
	for i, a := range plan.anchors {
		if !a.wonder {
			continue
		}
		scale := tdWonderScale(i)
		prom := 1000.0
		if i > 0 {
			prom = 800
		}
		plan.lots = append(plan.lots, tdLot{
			x: a.cx, y: a.cy, w: cfg.roofSize * scale, h: cfg.roofSize * scale,
			kind: tdRoof, domain: a.bld.domain, category: a.bld.category, tier: a.bld.tier,
			roof: roofWonder, label: a.bld.name, prom: prom,
		})
	}
}

// ---- town square (dress the plaza) ------------------------------------------
//
// The wonder-anchored plaza region reads as an EMPTY center (a donut) if left bare. Rather than
// SHRINK the open center, we DRESS it: each plaza-clearing anchor (the wonders + the wonderless
// city-center) gets a deliberate TOWN SQUARE — a paved-stone ground patch under the roof plus a
// few seeded, era-appropriate props arranged AROUND (never overlapping) the roof.

// tdSquareProps is one era's town-square prop palette: the prop-lot kinds a wonder square is
// dressed with (full set) and the modest kinds a wonderless center gets (a couple). Kept as a
// per-era value so LATER eras can swap the set (fountain/statue/benches) via tdSquarePropsFor —
// only the PRIMITIVE set is tuned now.
type tdSquareProps struct {
	wonder []tdLotKind // props ringed around a wonder centerpiece (the full square)
	center []tdLotKind // props for a wonderless city-center's modest square
}

// tdSquarePropsFor returns the town-square prop palette for an era. PRIMITIVE (organic): a well,
// a firepit, standing stones/totem, and a market stall around a wonder; a well + firepit for a
// bare center. Every other era falls back to the primitive set until V3-B/C/D tunes its own.
func tdSquarePropsFor(style tdEraStyle) tdSquareProps {
	return tdSquareProps{
		wonder: []tdLotKind{tdPropWell, tdPropFirepit, tdPropStones, tdPropStall},
		center: []tdLotKind{tdPropWell, tdPropFirepit},
	}
}

// tdPlaceSquares dresses every plaza-clearing anchor as a town square. For each WONDER anchor it
// lays a paved-stone plaza patch (filling the cleared plaza radius, drawn under the wonder roof)
// and rings a few era props AROUND the wonder roof — seeded, deterministic, arranged on a circle
// strictly OUTSIDE the roof footprint and inside the plaza. The wonderless city-center anchor
// gets a MODEST square — a small paved patch + a well/firepit — so the heart still reads as a
// gathering place without hollowing a hut village into a donut. Pure + seeded → positions are a
// function of (anchor index, seed), never of any building count.
func tdPlaceSquares(plan *topPlan, style tdEraStyle, cfg tdConfig, seed uint32) {
	if len(plan.anchors) == 0 {
		return
	}
	props := tdSquarePropsFor(style)
	plazaR := tdPlazaRadius(plan.form, cfg) // form-aware: organic keeps a small one-ward plaza
	for i, a := range plan.anchors {
		if a.wonder {
			roofHalf := cfg.roofSize * tdWonderScale(i) / 2
			// Paved plaza patch filling the cleared plaza radius (a square lot whose half-extent
			// is the plaza radius). The wonder roof is kept clear of the plaza RIM so a paved
			// ring always shows around it even after fill-frame shrinks the plaza at high counts.
			plan.lots = append(plan.lots, tdLot{
				x: a.cx, y: a.cy, w: plazaR * 2, h: plazaR * 2, kind: tdPlaza,
			})
			ringR := (roofHalf + plazaR) / 2
			if ringR < roofHalf+cfg.roofSize*0.6 {
				ringR = roofHalf + cfg.roofSize*0.6 // keep clear of the roof even if the band is tight
			}
			tdRingProps(plan, a.cx, a.cy, ringR, props.wonder, uint32(i), seed)
		} else {
			// Wonderless center: a MODEST square (small paved patch + a couple of props), kept
			// small so a tiny village keeps a FILLED heart, not a donut. For the ORGANIC form this
			// is doubly important — the paved patch must be genuinely small (~one roof) so buildings
			// sit close to the true center and the village never reads as a ring around a void; cap
			// it at a small absolute size on top of the plaza-relative shrink.
			smallR := plazaR * 0.55
			if cap := cfg.roofSize * 1.4; plan.form == formOrganic && smallR > cap {
				smallR = cap
			}
			plan.lots = append(plan.lots, tdLot{
				x: a.cx, y: a.cy, w: smallR * 2, h: smallR * 2, kind: tdPlaza,
			})
			tdRingProps(plan, a.cx, a.cy, smallR*0.55, props.center, uint32(i), seed)
		}
	}
}

// tdRingProps places one lot per prop kind evenly around a circle of radius ringR about
// (cx,cy), with a small seeded angular phase (from anchor index di) so different squares aren't
// rotationally identical. Each prop lot is small (well under the roof size) so it reads as a dab
// of detail. Deterministic: positions are a pure function of (di, seed, prop index).
func tdRingProps(plan *topPlan, cx, cy, ringR float64, kinds []tdLotKind, di, seed uint32) {
	n := len(kinds)
	if n == 0 || ringR <= 0 {
		return
	}
	phase := float64(hash2(di*131+5, 0x50a2, seed)) / float64(^uint32(0)) * 2 * math.Pi
	for k, kind := range kinds {
		ang := phase + 2*math.Pi*float64(k)/float64(n)
		rj := (float64(hash2(di*131+uint32(k)+17, 0x9e37, seed))/float64(^uint32(0)) - 0.5) * ringR * 0.12
		rr := ringR + rj
		plan.lots = append(plan.lots, tdLot{
			x:    cx + math.Cos(ang)*rr,
			y:    cy + math.Sin(ang)*rr,
			w:    defaultTdConfig.roofSize * 0.55,
			h:    defaultTdConfig.roofSize * 0.55,
			kind: kind,
		})
	}
}

// moreProminentBld ranks two buildings for hero selection: higher tier, then higher count, then
// name for a stable tiebreak.
func moreProminentBld(a, b builtBuilding) bool {
	if a.tier != b.tier {
		return a.tier > b.tier
	}
	if a.count != b.count {
		return a.count > b.count
	}
	return a.name < b.name
}

// prominenceOfLot is a scalar prominence for a placed landmark lot's label priority (tier
// dominant). Used to prioritise a promoted-hero / wonder label in the overlay z-order.
func prominenceOfLot(tier int) float64 {
	return float64(tier)*100 + 1
}

// tdRoofCount maps a building's instance count to how many roof lots it emits. Near 1:1 at low
// counts (locked #4: 24 huts ≈ 24 hut-roofs) and SUB-LINEAR (sqrt) at high counts so a
// metropolis densifies without N identical clones. Residential and production share the same
// curve; both keep the near-1:1 low band. At least 1 lot so a lone building still appears.
func tdRoofCount(count int, role cityRole) int {
	if count <= 0 {
		return 0
	}
	c := float64(count)
	var n float64
	if c <= 12 {
		n = c // near-1:1 low band
	} else {
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

// ---- filler (balanced living city) ------------------------------------------

// tdAddFiller lays balanced living-city greenery into the town (map-overhaul-citymap step f).
// Gardens, ponds, street trees and props fill LEFTOVER in-block space — the DEEP interior cells
// of the blocks (the cells NOT used as a building perimeter slot) plus a few groves hugging the
// town edge. Everything in-town stays inside the built-up footprint; only the deliberate groves
// sit just past the rim. Density is balanced (locked #12) and fully seeded so it's stable.
func tdAddFiller(plan *topPlan, field blockField, style tdEraStyle, cfg tdConfig, seed uint32) {
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

	// Candidate greenery spots: the DEEP-interior block cells (interior cells whose center is
	// clear of every placed roof and every street) — leftover in-block space. Gather them, then
	// pick a seeded subset for gardens/ponds/trees so filler lands IN the blocks, woven through
	// the town, not scattered on the empty map. Ordered deterministically (row-major).
	var deep []tdPoint
	if field.gridN >= 3 {
		clearR := cfg.roofSize * 0.9
		for gy := 0; gy < field.gridN; gy++ {
			for gx := 0; gx < field.gridN; gx++ {
				c := gy*field.gridN + gx
				si := field.nearest[c]
				if si < 0 || field.street[c] || field.plazaSeed[si] {
					continue
				}
				p := field.cellCenter(gx, gy)
				near := false
				for _, lt := range plan.lots {
					if lt.kind != tdRoof {
						continue
					}
					if math.Hypot(lt.x-p.x, lt.y-p.y) < clearR+math.Max(lt.w, lt.h)/2 {
						near = true
						break
					}
				}
				if !near {
					deep = append(deep, p)
				}
			}
		}
	}

	// Counts scale with the number of roofs but stay balanced (sub-linear) so the filler seasons
	// the town rather than swamping it.
	gardens := int(dens * math.Sqrt(float64(roofN)) * 1.0)
	trees := int(dens * math.Sqrt(float64(roofN)) * 0.9)
	props := int(dens * math.Sqrt(float64(roofN)) * 0.6)
	ponds := 1 + int(dens*math.Sqrt(float64(roofN))*0.18)
	if ponds > 5 {
		ponds = 5 // a few, never many
	}

	r := newRNG(hash2(0xF111, uint32(roofN), seed) | 1)
	pick := func() (tdPoint, bool) {
		if len(deep) == 0 {
			return tdPoint{}, false
		}
		// Seeded pick without replacement so two features don't stack on one cell.
		i := int(r.next() % uint32(len(deep)))
		p := deep[i]
		deep[i] = deep[len(deep)-1]
		deep = deep[:len(deep)-1]
		return p, true
	}

	// Gardens: green plots in leftover block interior.
	for i := 0; i < gardens; i++ {
		p, ok := pick()
		if !ok {
			break
		}
		plan.lots = append(plan.lots, tdLot{x: p.x, y: p.y, w: cfg.roofSize * 1.2, h: cfg.roofSize * 1.0, kind: tdGarden})
	}
	// Ponds (FIX 4): a FEW built decorative pools mixed IN-TOWN among the greenery.
	for i := 0; i < ponds; i++ {
		p, ok := pick()
		if !ok {
			break
		}
		plan.lots = append(plan.lots, tdLot{x: p.x, y: p.y, w: cfg.roofSize * 1.1, h: cfg.roofSize * 0.9, kind: tdPond})
	}
	// Street trees: small dot clusters in leftover block interior.
	for i := 0; i < trees; i++ {
		p, ok := pick()
		if !ok {
			break
		}
		plan.lots = append(plan.lots, tdLot{x: p.x, y: p.y, w: cfg.roofSize * 0.7, h: cfg.roofSize * 0.7, kind: tdTree})
	}
	// Props: wells/stalls in leftover block interior.
	for i := 0; i < props; i++ {
		p, ok := pick()
		if !ok {
			break
		}
		plan.lots = append(plan.lots, tdLot{x: p.x, y: p.y, w: cfg.roofSize * 0.5, h: cfg.roofSize * 0.5, kind: tdProp})
	}

	// A paved square (or two) hugging a random near-core block, for a lived-in market patch.
	squares := 1 + int(dens*math.Sqrt(float64(roofN))*0.25)
	for i := 0; i < squares; i++ {
		p, ok := pick()
		if !ok {
			break
		}
		plan.lots = append(plan.lots, tdLot{x: p.x, y: p.y, w: cfg.roofSize * 1.4, h: cfg.roofSize * 1.4, kind: tdSquare})
	}

	// Groves: 2–4 deliberate stands of trees JUST OUTSIDE the town edge, each a tight knot of a
	// few tree lots so it reads as a copse, not stray dots. Seeded at spread angles so the groves
	// ring the town loosely without scattering.
	groveCount := 2 + int(r.f01()*3) // 2..4
	if groveCount > 4 {
		groveCount = 4
	}
	groveBase := r.f01() * 2 * math.Pi
	for g := 0; g < groveCount; g++ {
		ang := groveBase + 2*math.Pi*float64(g)/float64(groveCount) + (r.f01()-0.5)*0.6
		gr := rad * (1.05 + 0.08*r.f01()) // hugging just past the town edge
		gx := plan.cx + math.Cos(ang)*gr
		gy := plan.cy + math.Sin(ang)*gr
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

// tdFootprintRadius returns the max distance of any roof lot from the core — the built-up
// radius the filler + walls hug.
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

// tdRoofBBox returns the bounding box (minX,minY,maxX,maxY) of all roof lots including their
// extent. Used by tests to assert the town sits within the frame with margin. Empty roof set →
// a zero box at the core.
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

// ---- walls (capability; primitive off) --------------------------------------

// tdAddWalls rings the built-up area with a wall + a few gates in city space (locked #9). V3-A
// wires the capability but PRIMITIVE keeps hasWalls=false, so this only runs for a (future) era
// that flips the flag — the code path is complete + tested.
func tdAddWalls(plan *topPlan, seed uint32) {
	rad := tdFootprintRadius(plan) * 1.15
	if rad <= 0 {
		return
	}
	const segs = 24
	gateEvery := segs / 4 // four gates, roughly cardinal
	r := newRNG(hash2(0x3a11, uint32(len(plan.lots)), seed) | 1)
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

// tdTransform maps city space → canvas pixels: p_px = (p_city - min) * scale + offset. Computed
// from the plan's lot bounding box so the whole city FILLS the canvas with a small margin
// (locked #3). A per-axis uniform scale (the smaller of the two) preserves the city's
// proportions; the offset centers it.
type tdTransform struct {
	scale       float64
	offX, offY  float64
	minX, minY  float64
	roofFloorPx float64 // minimum roof extent in px (legibility floor, locked #3)
}

// computeTransform derives the fill-frame transform. It fits the BUILT city — the roof lots, the
// street cells, and the core — leaving a small padding, and scales so that box fills the canvas.
// Filler/greenery is deliberately EXCLUDED from the fit: a couple of edge groves must not zoom
// the whole town out. Roofs shrink as the city densifies (more roofs → larger box → smaller
// scale) but a min roof-size FLOOR keeps them legible. Panic-safe: a degenerate (empty /
// zero-extent) box yields a centered identity-ish transform.
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
	// Street cells are part of the built fabric; include them so the block-gap network stays
	// on-frame (and a plaza-only / tiny town still fits its streets).
	for _, p := range plan.streetCells {
		acc(p.x, p.y, plan.cellSize/2)
	}
	// Always include the core so an empty plan still centers sensibly.
	acc(plan.cx, plan.cy, 1)
	// Padding around the built box so the town breathes AND the near-edge groves (which sit
	// ~1.05–1.13× the town radius past the core) land inside the frame margin rather than being
	// flung off-canvas. ~15% covers the grove ring while still leaving the roofs large.
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

// ext maps a city-space extent (half-size) to a pixel half-size, enforcing the roof legibility
// floor so roofs never shrink below a visible size (locked #3).
func (t tdTransform) ext(cityExt float64) int {
	e := cityExt * t.scale
	if e < t.roofFloorPx {
		e = t.roofFloorPx
	}
	return int(e + 0.5)
}

// ---- render (top-down) ------------------------------------------------------

// renderTopDown paints the whole top-down city onto img and returns the landmark overlay
// geometry. Paint order (city-synthesis.md §Rendering, top-down):
//
//	era-tinted ground (+ subtle seeded texture noise)
//	→ streets (the bold packed-earth GAPS between the Voronoi blocks)
//	→ ground accents (plaza paving, gardens, squares, ponds)
//	→ roof sprites (roof atlas, drawn BACK-TO-FRONT by y so SE shadows layer)
//	→ trees / props
//	→ walls / gates (if the era has them)
//	→ (landmark labels are stamped by the overlay pass, from the returned geometry)
//
// Pure, panic-safe, exact output size. Every color is theme-derived (retints on a theme
// switch). NO terrain, NO biome, NO water — the ground is a neutral era tint.
func renderTopDown(img *image.RGBA, state game.GameState, w, h int, seed uint32) layoutGeometry {
	e := eraForAge(state.Age)
	style := tdStyleForEra(e)
	pal := newTdPal()

	// Ground first — a full era-tinted fill with subtle seeded texture noise.
	drawGround(img, style, pal, seed, w, h)

	plan := generateTopPlan(state, config.BuildingByKey(), style, seed)
	xf := computeTransform(&plan, w, h)

	// Streets: the BOLD packed-earth GAPS between the blocks (map-overhaul-citymap). Each street
	// cell is a filled packed-earth square with a crisp edge, drawn UNDER the roofs; because the
	// buildings sit INSET inside their blocks facing the gaps, the street network reads clearly.
	streetCol := style.streetCol(pal)
	streetEdge := tdStreetEdgeColor(style, pal)
	drawStreetCells(img, xf, plan, streetCol, streetEdge)

	// Ground accents: gardens, squares, and the TOWN-SQUARE paved plazas painted before roofs so
	// a roof sits on top. The plaza is a rounded paved apron under the wonder/center roof.
	gardenCol := style.gardenCol(pal)
	squareCol := style.squareCol(pal)
	pavedCol := tdPavedColor(style, pal)
	pondCol := tdPondColor(style, pal)
	for _, lt := range plan.lots {
		switch lt.kind {
		case tdGarden:
			cx, cy := xf.px(lt.x, lt.y)
			fillRectC(img, cx, cy, xf.ext(lt.w/2), xf.ext(lt.h/2), gardenCol)
		case tdSquare:
			cx, cy := xf.px(lt.x, lt.y)
			fillRectC(img, cx, cy, xf.ext(lt.w/2), xf.ext(lt.h/2), squareCol)
		case tdPlaza:
			cx, cy := xf.px(lt.x, lt.y)
			drawPlaza(img, cx, cy, xf.ext(lt.w/2), xf.ext(lt.h/2), pavedCol)
		case tdPond:
			drawPond(img, xf, lt, pondCol)
		}
	}

	// Roof sprites, BACK-TO-FRONT (sort by city y ascending) so a nearer roof's SE shadow lays
	// over the roof behind it. Collect roof lots, sort a copy of indices.
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

	// Trees + props on top of the ground fabric. The typed town-square props dispatch to their
	// own small draw routines; the generic tdProp filler stays a quiet dab.
	treeCol := style.treeCol(pal)
	propCol := style.propCol(pal)
	for _, lt := range plan.lots {
		switch lt.kind {
		case tdTree:
			drawTree(img, xf, lt, treeCol)
		case tdProp:
			cx, cy := xf.px(lt.x, lt.y)
			drawBlock(img, cx, cy, 0, propCol)
		case tdPropWell, tdPropFirepit, tdPropStones, tdPropStall:
			drawSquareProp(img, xf, lt, style, pal)
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

// drawGround fills the whole canvas with the era ground tone plus a subtle, seeded texture: each
// pixel picks base or a slightly-varied alt tone from a cheap hash of its coordinates, so the
// dirt reads as textured earth rather than a flat wash. No water, no biome — a neutral
// era-tinted ground (locked #2).
//
// CALM GROUND (playtest polish FIX 1): the texture is a WHISPER now, not a busy speckle. Two
// dials pull the contrast down hard: (1) far fewer pixels are touched (groundTexFrac), and (2)
// the max blend toward alt is a small fraction (groundTexAmp) so even a touched pixel barely
// shifts. The result is a subtle-but-present dirt base that lets buildings + streets stand out.
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
			n := texHash(uint32(x), uint32(y), seed)
			t := 0.0
			if n < groundTexFrac {
				t = 0.5 + 0.5*n/groundTexFrac // 0.5..1.0 within the touched band
			}
			img.SetRGBA(px, py, blend(base, alt, t*groundTexAmp))
		}
	}
}

// Ground-texture calm dials (playtest polish FIX 1). groundTexFrac is the fraction of pixels
// that take ANY alt tint (was 0.22); groundTexAmp caps how far a touched pixel blends toward the
// alt tone (was 0.60). Kept as named consts so the quieter-ground test can reason about the
// reduction and a future era can retune without a magic number.
const (
	groundTexFrac = 0.12
	groundTexAmp  = 0.20
)

// texHash is a tiny deterministic 2D value hash returning a float in [0,1). Used for the ground
// texture speckle — cheap, seeded, and stable frame-to-frame.
func texHash(x, y, seed uint32) float64 {
	h := seed + x*374761393 + y*668265263
	h = (h ^ (h >> 13)) * 1274126177
	h ^= h >> 16
	return float64(h&0xffffff) / float64(0x1000000)
}

// drawStreetCells paints the Voronoi block-boundary GAPS as bold packed-earth squares
// (map-overhaul-citymap). Each street cell (a raster boundary cell) maps through the fill-frame
// transform to a filled square in the packed-earth surface tone with a crisp darker edge one
// band out, so the connected gap network reads as trodden streets between the blocks. Reuses the
// era's packed-earth street tone (playtest FIX 1). Drawn UNDER the roofs; the buildings sit inset
// inside their blocks, so the streets stay visible. Panic-safe: floored extents, clipped setters.
func drawStreetCells(img *image.RGBA, xf tdTransform, plan topPlan, surface, edge color.RGBA) {
	if len(plan.streetCells) == 0 || plan.cellSize <= 0 {
		return
	}
	// A street cell covers cellSize×cellSize in city space; render it a hair larger than the
	// half-cell so adjacent cells' squares still meet into one continuous gap (no seams), but NOT so
	// much that the overlap fattens the thin lane back into a wide band (map-overhaul-citymap FIX 1 —
	// NARROWER roads). The old 1.15× overlap + an always-additive +1 edge padded every lane wider than
	// its cells; a ~1.05× cover keeps the web continuous while the streets stay thin. The edge is a
	// 1px-larger square drawn FIRST so the surface overpaints its interior, leaving only a thin darker
	// rim; at a floored 1px surface the edge is capped so it can't dominate a village-scale lane.
	half := plan.cellSize / 2
	surfHalf := xf.ext(half * 1.05)
	if surfHalf < 1 {
		surfHalf = 1
	}
	edgeHalf := surfHalf
	if surfHalf >= 2 {
		edgeHalf = surfHalf + 1 // a crisp shoulder only once the lane is wide enough to carry one
	}
	for _, p := range plan.streetCells {
		cx, cy := xf.px(p.x, p.y)
		fillRectC(img, cx, cy, edgeHalf, edgeHalf, edge)
	}
	for _, p := range plan.streetCells {
		cx, cy := xf.px(p.x, p.y)
		fillRectC(img, cx, cy, surfHalf, surfHalf, surface)
	}
}

// tdStreetEdgeColor resolves the street shoulder tone for a style: the preset's streetEdge
// recipe when set, else a derived fallback (the surface darkened). Pure theme read → retints on
// a switch.
func tdStreetEdgeColor(style tdEraStyle, pal tdPal) color.RGBA {
	if style.streetEdge != nil {
		return style.streetEdge(pal)
	}
	return darken(style.streetCol(pal), 0.22)
}

// ---- town-square paving + props (playtest FIX) ------------------------------

// tdPavedColor resolves the town-square paving tone for a style. It uses the preset's pavedCol
// recipe when set; otherwise it derives a safe fallback (the square tone lifted toward the light
// neutral). Pure theme read → retints on a theme switch like every other tone.
func tdPavedColor(style tdEraStyle, pal tdPal) color.RGBA {
	if style.pavedCol != nil {
		return style.pavedCol(pal)
	}
	return blend(style.squareCol(pal), pal.text, 0.30)
}

// tdPondColor resolves the BUILT-pond water tone for a style (playtest polish FIX 4). It uses
// the preset's pondCol recipe when set; otherwise it derives a safe fallback. Pure theme read →
// retints on a theme switch. This is decorative BUILT water, never natural terrain water.
func tdPondColor(style tdEraStyle, pal tdPal) color.RGBA {
	if style.pondCol != nil {
		return style.pondCol(pal)
	}
	return blend(blend(pal.bg, pal.dim, 0.18), waterAnchor, 0.55)
}

// drawPlaza paints the paved town-square ground as a rounded apron: a filled ellipse in the
// paving tone with a faintly darker rim so the made surface reads with a soft edge against the
// surrounding dirt. Drawn under the wonder/center roof + props. Half-extents floored so a small
// center square still shows.
func drawPlaza(img *image.RGBA, cx, cy, hw, hh int, paved color.RGBA) {
	if hw < 1 {
		hw = 1
	}
	if hh < 1 {
		hh = 1
	}
	forEllipse(cx, cy, hw, hh, func(x, y int) { setPixel(img, x, y, paved) })
	rim := darken(paved, 0.12)
	forEllipse(cx, cy, hw, hh, func(x, y int) {
		fx := float64(x-cx) / float64(hw)
		fy := float64(y-cy) / float64(hh)
		if fx*fx+fy*fy >= 0.72 { // outer ring band only
			setPixel(img, x, y, rim)
		}
	})
}

// drawSquareProp renders one town-square prop as a small top-down dab, dispatched by its lot
// kind. Tones come from the era style so props retint with the theme and stay in the era mood.
// Kept tiny — a well/firepit/totem/stall reads as a detail dressing the square. All radii
// floored so a prop paints at least a couple of pixels even at village scale.
func drawSquareProp(img *image.RGBA, xf tdTransform, lt tdLot, style tdEraStyle, pal tdPal) {
	cx, cy := xf.px(lt.x, lt.y)
	rad := xf.ext(math.Max(lt.w, lt.h) / 2)
	if rad < 1 {
		rad = 1
	}
	paved := tdPavedColor(style, pal)
	prop := style.propCol(pal)
	switch lt.kind {
	case tdPropWell:
		fillDisc(img, cx, cy, rad, brighten(paved, 0.10))
		setPixel(img, cx, cy, darken(prop, 0.55)) // the dark shaft
	case tdPropFirepit:
		fillDisc(img, cx, cy, rad, darken(prop, 0.45))
		ember := brighten(blend(prop, color.RGBA{R: 0xc8, G: 0x5a, B: 0x1e, A: 0xff}, 0.6), 0.10)
		setPixel(img, cx, cy, ember)
	case tdPropStones:
		stone := blend(prop, paved, 0.35)
		drawBlock(img, cx-rad, cy, 0, stone)
		drawBlock(img, cx, cy-rad, 0, brighten(stone, 0.10))
		drawBlock(img, cx+rad, cy, 0, darken(stone, 0.10))
		setPixel(img, cx, cy, stone)
	case tdPropStall:
		cloth := blend(prop, pal.text, 0.20)
		fillRectC(img, cx, cy, rad, rad, cloth)
		drawHSpan(img, cx-rad, cx+rad, cy-rad, brighten(cloth, 0.16))
	}
}

// drawPond paints a BUILT decorative pond as a small water blob (playtest polish FIX 4): a
// filled blue/teal ellipse in the water tone with a subtle LIGHTER RIM one band in from the edge
// (a shallow shore) and a faintly darker deep center, so the pool reads as a made little pond
// from above. Kept small — a village ornament. All radii floored so it paints at village scale.
func drawPond(img *image.RGBA, xf tdTransform, lt tdLot, water color.RGBA) {
	cx, cy := xf.px(lt.x, lt.y)
	hw := xf.ext(lt.w / 2)
	hh := xf.ext(lt.h / 2)
	if hw < 1 {
		hw = 1
	}
	if hh < 1 {
		hh = 1
	}
	forEllipse(cx, cy, hw, hh, func(x, y int) { setPixel(img, x, y, water) })
	rim := brighten(water, 0.18)
	forEllipse(cx, cy, hw, hh, func(x, y int) {
		fx := float64(x-cx) / float64(hw)
		fy := float64(y-cy) / float64(hh)
		if fx*fx+fy*fy >= 0.6 { // outer band only
			setPixel(img, x, y, rim)
		}
	})
	setPixel(img, cx, cy, darken(water, 0.15))
}

// drawTree paints a tree as a small dark-green dot cluster (a filled blob with a darker core),
// so a stand of trees reads as foliage from above.
func drawTree(img *image.RGBA, xf tdTransform, lt tdLot, c color.RGBA) {
	cx, cy := xf.px(lt.x, lt.y)
	rad := xf.ext(lt.w / 2)
	if rad < 1 {
		rad = 1
	}
	fillDisc(img, cx, cy, rad, c)
	img.SetRGBA(cx, cy, darken(c, 0.25))
}

// ---- roof atlas draw --------------------------------------------------------

// drawRoof renders one building lot as a TOP-DOWN roof filling its lot: a soft SE drop-shadow
// first (subtle depth, NOT isometric walls — locked #6), then the roof shape read straight from
// above. Material comes from the era style; a subtle lineage tint differentiates types. EVERY
// tone is base/dark/ridge from roofColorsFor — ridge is a base-derived lighten, so NO saturated
// theme accent ever lands on a roof (the yellow-dot fix). Dispatches on the lot's roofType.
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

	// SE drop-shadow: the roof footprint, offset down-right by ~1px, painted UNDER the roof. Soft
	// = the theme shadow tone blended into the ground — a hint of height without isometric walls.
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

// drawShadow paints a soft SE drop-shadow matching the roof's rough silhouette. It blends the
// shadow tone into whatever is beneath (so it darkens the ground, not paints a hard slab), giving
// a subtle hint of height. Kept faint (~0.28) so it reads as depth, not an outline.
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

// drawRoofHut: a small ROUND thatch roof seen from straight above — a solid thatch disc with soft
// top-down SHADING (lit toward the NW crown, shaded toward the SE eave) so it reads as a domed
// thatch cap, not a target or a pinwheel. It must read round, NOT a 4-point diamond: at tiny
// radii forEllipse degenerates to a plus, so the footprint is floored to a minimum radius.
func drawRoofHut(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	rw, rh := hutRadius(hw), hutRadius(hh)
	shade := blend(rc.base, rc.dark, 0.5)
	forEllipse(cx, cy, rw, rh, func(x, y int) {
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

// hutRadius floors a hut half-extent so a hut never collapses into a plus/diamond at small sizes
// (forEllipse includes only the cardinal cells at radius 1). Minimum 2 so the smallest hut is
// still an unmistakable little round roof.
func hutRadius(half int) int {
	if half < 2 {
		return 2
	}
	return half
}

// drawRoofRidge: a rectangular PITCHED roof read from above — a filled rectangle split by a
// center ridge line running the long axis, with the two slopes shaded slightly differently so
// the pitch reads. Serves both the house/longhouse (roofLong is just a wider lot) and the
// flat/default fallback. The ridge line is base-derived (rc.ridge), one shade lighter.
func drawRoofRidge(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	horizontalRidge := hw >= hh // ridge runs along the wider axis
	forRect(cx, cy, hw, hh, func(x, y int) {
		var slope color.RGBA
		if horizontalRidge {
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
	if horizontalRidge {
		drawHSpan(img, cx-hw, cx+hw, cy, rc.ridge)
	} else {
		for y := cy - hh; y <= cy+hh; y++ {
			img.SetRGBA(cx, y, rc.ridge)
		}
	}
}

// drawRoofTemple: the larger, grandest small building — an ornate symmetric tiered roof read
// from above. A full base footprint, a lighter stepped inner tier, and a small subtle central
// peak, all base-derived (no accent finial — that was the yellow dot). Cross ridges on both axes
// make it read symmetric and civic.
func drawRoofTemple(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	forRect(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.dark) })
	mhw := maxInt(hw*2/3, 1)
	mhh := maxInt(hh*2/3, 1)
	forRect(cx, cy, mhw, mhh, func(x, y int) { img.SetRGBA(x, y, rc.base) })
	phw := maxInt(hw/3, 1)
	phh := maxInt(hh/3, 1)
	forRect(cx, cy, phw, phh, func(x, y int) { img.SetRGBA(x, y, rc.ridge) })
	drawHSpan(img, cx-hw, cx+hw, cy, rc.ridge)
	for y := cy - hh; y <= cy+hh; y++ {
		img.SetRGBA(cx, y, rc.ridge)
	}
}

// drawRoofCamp: an OPEN lean-to / A-frame — the gathering/forager/war camp, drawn so it reads as
// a temporary open structure, clearly NOT a solid house. Only the top (north) half of the lot is
// roofed (a lean-to whose slope faces the viewer), with the open front left as ground.
func drawRoofCamp(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	for dy := -hh; dy <= 0; dy++ {
		f := float64(dy+hh) / float64(hh+1)
		rowHW := int(float64(hw) * (1 - f*0.15))
		col := blend(rc.base, rc.dark, f*0.6)
		drawHSpan(img, cx-rowHW, cx+rowHW, cy+dy, col)
	}
	drawLineC(img, cx-hw, cy-hh, cx-hw, cy, rc.dark)
	drawLineC(img, cx+hw, cy-hh, cx+hw, cy, rc.dark)
	drawHSpan(img, cx-hw, cx+hw, cy-hh, rc.ridge)
}

// drawRoofStash: a small, low, plain square roof — the storage store-hut, quieter than a
// dwelling. A compact filled square (the shaded tone so it sits low) with a single base-derived
// plank highlight across the top edge.
func drawRoofStash(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	s := minInt(hw, hh)
	if s < 1 {
		s = 1
	}
	forRect(cx, cy, s, s, func(x, y int) { img.SetRGBA(x, y, rc.dark) })
	drawHSpan(img, cx-s, cx+s, cy-s, rc.ridge)
}

// drawRoofFlat: a low flat structure — a flat slab with a thin darker rim, for the stone works.
// Reads distinctly from a pitched dwelling (no ridge line).
func drawRoofFlat(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	forRect(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.base) })
	forRectOutline(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.dark) })
}

// drawRoofWonder: the DOMINANT central complex — a large ornate multi-part roof read from above
// (locked #13). A grand elliptical base, a stepped square inner hall, and a bright base-derived
// top tier + cross ridges: unmistakably the grandest thing on the map, and STILL in the roof
// material family (no accent — the crown is a base-derived lighten, not a saturated finial).
func drawRoofWonder(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	forEllipse(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.base) })
	ihw := maxInt(hw*2/3, 1)
	ihh := maxInt(hh*2/3, 1)
	forRect(cx, cy, ihw, ihh, func(x, y int) { img.SetRGBA(x, y, blend(rc.base, rc.dark, 0.25)) })
	thw := maxInt(hw/3, 1)
	thh := maxInt(hh/3, 1)
	forRect(cx, cy, thw, thh, func(x, y int) { img.SetRGBA(x, y, rc.ridge) })
	drawHSpan(img, cx-hw, cx+hw, cy, rc.ridge)
	for y := cy - hh; y <= cy+hh; y++ {
		img.SetRGBA(cx, y, rc.ridge)
	}
}

// ---- pixel primitives (top-down) --------------------------------------------

// forRect calls fn for every pixel of the filled (2*hw+1)×(2*hh+1) rect centered on (cx,cy).
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

// forEllipse calls fn for every pixel inside the axis-aligned ellipse with radii (hw,hh)
// centered on (cx,cy).
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

// (blendPixel lives in flourish.go and is shared — it mixes a color into the existing pixel by
// t, clipped to the image; the drop-shadows use it to darken the ground rather than paint over
// it.)

// drawLineC rasterizes a solid Bresenham line in color c (reuses drawRoad's engine).
func drawLineC(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	drawRoad(img, roadSeg{x0, y0, x1, y1}, c)
}

// ---- landmark overlay geometry ----------------------------------------------

// tdGeometry extracts the LANDMARK-ONLY overlay anchors (locked #7): the city center, plus one
// buildingLabel per labeled landmark (the civic/promoted hero and the wonder). No per-house
// labels. Returned as the same layoutGeometry the overlay pass consumes, so the existing label
// pipeline stamps them unchanged. Anchors are in pixel space.
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

// buildLandmarkOverlay assembles the KEY-LANDMARKS-ONLY overlay plan for the top-down city
// (locked #7): the labeled landmark roofs (city-center hero, wonder, or a promoted hero when the
// civ has no civic building), the "City Center" label, and the corner title. It deliberately
// OMITS the old systems-weave (per-house labels, the diplomacy civ-edge ring, and trade-lane
// tags) — the city view reads by roof shape/color, not a wall of text. Reuses the existing
// overlayPlan builders + stampOverlay unchanged, so a theme switch still retints the text.
func buildLandmarkOverlay(state game.GameState, cols, rows int, geo layoutGeometry) overlayPlan {
	var plan overlayPlan
	if cols <= 0 || rows <= 0 {
		return plan
	}
	occupied := map[int]bool{}
	plan.addBuildingLabels(geo, cols, rows, occupied)
	plan.addCityCenterLabel(geo, cols, rows, occupied)
	plan.addTitle(state, cols, rows)
	return plan
}
