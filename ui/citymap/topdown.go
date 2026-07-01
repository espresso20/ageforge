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

	// streetEdge is the crisp darker edge stroked one cell out from the packed-earth
	// lane band (playtest FIX 1: BOLD packed-earth roads). A subtle darken of the lane
	// surface so the trodden path reads with a defined shoulder against the dirt ground,
	// without a hard black outline. Theme-derived like streetCol; nil → derived fallback.
	streetEdge func(tdPal) color.RGBA

	// Living-city filler accents.
	gardenCol func(tdPal) color.RGBA
	squareCol func(tdPal) color.RGBA
	treeCol   func(tdPal) color.RGBA
	propCol   func(tdPal) color.RGBA

	// pondCol is the BUILT decorative-pond water tone (playtest polish FIX 4). A pond is a
	// small MADE water feature dug into the town fabric (a garden-family ornament), NOT
	// natural terrain water — the citymap has no biome/ocean (locked #2). The recipe pulls
	// the theme background toward a muted blue/teal water anchor so it retints on a theme
	// switch and stays in-family; nil → tdPondColor derives a safe fallback. Drawn as a
	// small blue blob with a lighter rim by drawPond.
	pondCol func(tdPal) color.RGBA

	// pavedCol is the TOWN-SQUARE paved-stone ground tone (playtest FIX: dress the
	// wonder/center plaza as a made surface, not bare dirt). Distinct from the
	// era-tinted dirt — a lighter/greyer packed-or-paved tint derived from the ground
	// family so the square reads as a deliberate paved surface, and retints on a theme
	// switch like every other recipe. Drawn under the wonder roof + props.
	pavedCol func(tdPal) color.RGBA

	// Walls capability (locked #9). Primitive sets false; walls arrive in V3-B.
	hasWalls bool
	wallCol  func(tdPal) color.RGBA

	fillerDensity float64 // balanced living-city filler amount (locked #12)

	// slotSpacing is the per-era golden-spiral spacing between building instances (the
	// PACK-DENSITY knob, playtest FIX 2 groundwork). PRIMITIVE stays AIRY (the current
	// value); later-era presets set it progressively TIGHTER so V3-B/C cities and the
	// metropolis pack denser. 0 → fall back to defaultTdConfig.slotSpacing so an untuned
	// era still renders. Routed into tdConfig.slotSpacing by generateTopPlan.
	slotSpacing float64
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
	// BOLD packed-earth main ways (playtest FIX 1): the village lanes are a STRONG visual
	// element, not a thin scratch. laneWidth 1 strokes a ~2–3px band (drawTdStreet centers
	// the band, so half-width 1 → 3px). Era-scalable: later eras can widen for boulevards.
	laneWidth: 1,
	// streetJitter: a GENTLE wind — the lanes still meander like footpaths (organic), but
	// not so sharply that offsetting the lane-lining rows (FIX 2) compresses same-side
	// neighbours into each other on tight inside bends. Was 0.9 (very wavy) under the old
	// pull-to-lane model where nothing lined the lane; lining needs gentler curves.
	streetJitter: 0.35,

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
		// BOLD packed-earth lane (playtest FIX 1): a WORN, TRODDEN TAN that reads clearly
		// LIGHTER than the era-tinted dirt ground so the streets actually stand out. Start
		// from the dirt anchor, then lift toward the theme's light neutral (RoleText) and a
		// touch toward the pale stone anchor for a bleached, packed-trail cast. The old
		// recipe DARKENED the dirt (near-invisible against the ground); this inverts that to
		// a high-contrast lighter path, still earthy + fully theme-derived (retints).
		packed := blend(blend(p.bg, p.dim, 0.28), dirtAnchor, 0.42)
		return blend(blend(packed, p.text, 0.42), stoneAnchor, 0.22)
	},
	streetEdge: func(p tdPal) color.RGBA {
		// Crisp shoulder: the packed-earth surface darkened a touch so the trodden band has a
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
	// water anchor. Kept in-family (blended, never raw) so it reads as a small dug pond in
	// the village greenery rather than a cartoon puddle, and retints with the theme.
	pondCol: func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.18), waterAnchor, 0.55)
	},

	// Town-square paving: packed pale stone. Start from the dirt square tone, then lift
	// it toward the theme's light neutral (RoleText) and pull the last touch toward a
	// grey stone anchor so it reads LIGHTER + GREYER than the surrounding trodden dirt —
	// a deliberately made surface, still earthy/village in mood, still theme-derived.
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

// defaultTdStyle is the reasonable fallback for every non-primitive era until
// V3-B/C/D tunes them (locked Phasing: "Other eras use a reasonable default
// preset"). It reuses the organic recipes for a legible, theme-tinted city with a
// simple ridged-rect default roof — NOT its final silhouette. Kept as a copy so
// tuning organic later can't accidentally restyle every era.
var defaultTdStyle = func() tdEraStyle {
	s := organicVillageStyle
	s.name = "default"
	s.streetPattern = tdOrganic
	// PER-ERA DENSITY (playtest FIX 2 groundwork): later eras pack TIGHTER than the airy
	// primitive village so V3-B/C cities + the metropolis read denser. This is framework
	// only — the primitive preset keeps its airy 2.4; every not-yet-tuned era renders on
	// this default at a tighter slot spacing. V3-B/C/D can dial each era band's own value.
	s.slotSpacing = 1.7
	return s
}()

// Muted hue anchors for the earthy village mood. Blended at modest strength against
// theme roles so a light or dark theme still gets an in-family, non-cartoon palette.
var (
	earthAnchor = color.RGBA{R: 0x8a, G: 0x63, B: 0x3a, A: 0xff} // thatch/wood brown
	dirtAnchor  = color.RGBA{R: 0x7c, G: 0x66, B: 0x46, A: 0xff} // packed dirt
	grassAnchor = color.RGBA{R: 0x4f, G: 0x6f, B: 0x33, A: 0xff} // built greenery
	// stoneAnchor is the pale packed-stone anchor for TOWN-SQUARE paving — a light warm
	// grey so a plaza reads as a made surface, lighter/greyer than the dirt, blended
	// against theme roles (never used raw) so it retints and stays in-family.
	stoneAnchor = color.RGBA{R: 0xa8, G: 0xa0, B: 0x92, A: 0xff} // packed pale stone
	// waterAnchor is the muted blue-teal anchor for BUILT decorative ponds (playtest polish
	// FIX 4) — a calm pool blue, never used raw: blended against theme roles so a pond
	// retints and stays in-family with the village palette rather than reading as cartoon water.
	waterAnchor = color.RGBA{R: 0x36, G: 0x6b, B: 0x8f, A: 0xff} // muted pond blue-teal
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
	tdPond                    // a BUILT decorative pond (small made water feature, FIX 4)
	tdWall                    // a wall segment
	tdGate                    // a gate in the wall ring

	// Town-square lots (playtest FIX: the wonder/center plaza is DRESSED as a town square,
	// not left as bare dirt). tdPlaza is the paved-stone ground patch under the wonder/
	// center roof + props; the tdProp* kinds are the deliberate, seeded, era-appropriate
	// props arranged AROUND (never overlapping) the roof. Kept as distinct kinds so each
	// prop gets its own small top-down draw routine and later eras can swap the set
	// (fountain/statue/benches) without disturbing the primitive one.
	tdPlaza       // paved-stone town-square ground (drawn under the wonder/center roof)
	tdPropWell    // a stone well head (ring + dark shaft)
	tdPropFirepit // a firepit / hearth (dark ring + ember center)
	tdPropStones  // standing stones / a totem (a couple of upright dabs)
	tdPropStall   // a market stall (awning patch)
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
	// junctions are the lane-network JUNCTION points (playtest polish FIX 2: connected
	// roads): the count-INDEPENDENT lattice of crossings where the cross-streets meet main
	// A, plus the central crossroads. The fabric CLEARS a small radius around each so no
	// building sits on a junction, and the connectivity test reads them as the graph's known
	// crossings. Derived from the fixed mains geometry + config (tdJunctionLattice), never
	// from the building count, so clearing a lot near one preserves stable-incremental.
	junctions []tdPoint
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

// goldenAngle is the golden angle in radians (~137.5°). Placing item i at angle
// i*goldenAngle with radius ∝ sqrt(i) gives a stable, low-overlap phyllotaxis SPREAD:
// every item has a FIXED index, so adding a later one never moves an earlier one (locked
// #8, the anti-re-randomize guarantee). Used for the WONDER ANCHOR spread (tdAnchorPoints)
// and to fan the compact BRANCH-street anchor points evenly inside the bounded town disc
// (tdGrowNetwork) — NOT for radial spokes (the retired pinwheel). The building fabric
// itself is no longer a spiral — it LINES the lanes (FIX 2, tdPopulateIntermixed).
const goldenAngle = 2.399963229728653 // math.Pi * (3 - sqrt(5))

// slotJitter returns a small ORGANIC offset for a lane-lining slot (keyed by slot index i
// and a group id di), breaking up the perfectly-regular row so the fabric reads natural
// rather than mechanical. CRITICAL: the offset is a PURE FUNCTION of (i, di, seed) via a
// hash — it does NOT draw from the threaded rng — so a slot's jitter is identical whether
// it is the last building placed at count N or an interior one at count N+1. That is what
// keeps placement stable-incremental (locked #8) even with jitter: adding a building can
// never move an existing one. amp is the max wander in city units; the lane-lining caller
// clamps it further (longitudinally, and outward-only perpendicular) so jitter can never
// push a roof onto the road or into a neighbour.
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

	// --- line-the-lanes placement (playtest FIX 2) ---------------------------------
	// The old pull-to-lane (laneBias) yanked building CENTERS onto the lanes, so the
	// fabric squished onto the road and the streets were invisible. It is replaced by
	// LINING: each roof sits ALONGSIDE a lane, offset perpendicular to the centerline so
	// the road stays visible between two opposing rows of buildings.
	//
	// laneHalf is the lane's half-width in city units (the road's own footprint). A lot's
	// perpendicular offset from the centerline is laneHalf + its own half-extent +
	// laneMargin, so it never sits on the road. Scaled to the era street width so bolder
	// eras keep their frontage clear of the wider road.
	laneHalf float64
	// laneMargin is the small clear gap between the road edge and the building face.
	laneMargin float64
	// minGap is the minimum clear space kept between ANY two roof lots (never touching).
	// The along-lane step is (roof extent + minGap), so same-side neighbours can't overlap.
	minGap float64
	// laneStartClear is how far along a lane (from its anchor/junction end) the first lot
	// sits, in units of roofSize — it keeps the crowded lane JUNCTIONS (where lanes meet
	// the core and each other) building-free so distinct lanes have diverged before the
	// fabric starts lining them (structural cross-lane no-overlap) and the road reads out
	// of the plaza.
	laneStartClear float64
	// plazaRadius is the clear ground kept immediately around a WONDER anchor (in units
	// of roofSize) so the city fabric never buries the centerpiece (the playtest
	// complaint). Non-wonder roof lots landing inside a wonder's plaza are dropped.
	plazaRadius float64

	// --- compact bounded street network (playtest FIX: no pinwheel) ----------------
	// The village is a COMPACT, BOUNDED tangle of streets contained within a town disc —
	// NOT radial spokes flung outward (the pinwheel regression). townBaseRadius is the disc
	// radius (city units) at the smallest settlement; townGrowth scales it up with SQRT of
	// the fabric-slot count so the footprint grows only SLOWLY (compact + airy, never flung
	// out). tdTownRadius(slots) = townBaseRadius + townGrowth*sqrt(slots).
	townBaseRadius float64
	townGrowth     float64
	// branchSpacing is the target frontage (city units) each lane-lining slot consumes; the
	// network grows enough MAIN + BRANCH + LOOP street to line `slots` lots at this spacing
	// WITHIN the disc — denser internal network as the count rises, not longer arms.
	branchSpacing float64
}

var defaultTdConfig = tdConfig{
	// anchorSpread: how far the outermost wonder anchor sits from the core (∝√N). Kept well
	// INSIDE the compact town disc (tdTownRadius) so multiple wonders sit interior and the
	// fabric genuinely HUGS each anchor rather than filling a disc between two edge anchors.
	anchorSpread:   11,
	slotSpacing:    2.4,
	roofSize:       3.2,
	jitterAmp:      0.8,
	laneHalf:       1.4,
	laneMargin:     0.9,
	minGap:         1.1,
	laneStartClear: 0.9,
	// plazaRadius: the clear-ground ring around a WONDER anchor, in units of roofSize
	// (playtest polish FIX 3: MORE breathing room). Bumped 2.2 → 3.0 so the city center +
	// wonders read OPEN and prominent — the fabric hugs the anchor from a comfortable
	// distance instead of crowding its apron. The paved town square fills this radius
	// (tdPlaceSquares), so a wider plaza is a wider, more deliberate square; the wonder roof
	// (tdWonderScale, ~2.6× → half-extent 1.3× roofSize) stays far inside it so a generous
	// paved ring always shows. insideWonderPlaza drops any fabric lot that lands inside it.
	plazaRadius: 3.0,
	// Compact + bounded: a small base disc that grows only ~sqrt with the count, so a big
	// village densifies its street network WITHIN a slowly-expanding footprint instead of
	// throwing longer radial arms. Tuned so the primitive village stays airy but readable.
	townBaseRadius: 16,
	townGrowth:     3.4,
	branchSpacing:  tdConfigFabricSpacing, // = tdFabricStep(default)/2 (halfStep per slot)
}

// tdConfigFabricSpacing is the along-lane advance per lane-lining slot at the default roof
// size (= tdFabricStep/2). Hoisted to a const so defaultTdConfig can seed branchSpacing
// without a init cycle; tdGrowNetwork uses cfg.branchSpacing (era-scalable) at runtime.
const tdConfigFabricSpacing = (3.2*2.0 + 1.1) / 2 // roofSize*2 + minGap, halved

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
	// Route the PER-ERA DENSITY knob (playtest FIX 2 groundwork): the era preset owns the
	// slot spacing so later eras pack tighter than the airy primitive village. 0 → keep the
	// config default (an untuned era still renders). Primitive maps to 2.4 (unchanged look).
	if style.slotSpacing > 0 {
		cfg.slotSpacing = style.slotSpacing
	}
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

	// (c) lanes — laid FIRST so the fabric LINES them and the outline follows the streets
	// (FIX 2). Winding MAIN lanes cross near the center/anchors; then the network is DENSIFIED
	// with BRANCH streets (and a light LOOP once the town is big enough) all contained WITHIN
	// a bounded town disc whose radius grows only slowly with sqrt(count) — a compact village
	// tangle, NOT radial spokes flung outward (the retired pinwheel). Append-only so existing
	// lanes (and the lots already lining them) never move as the count rises.
	spanSlots := tdFrontageSpanSlots(blds)
	plan.streets = tdOrganicStreets(plan.anchors, plan.cx, plan.cy, cfg, style, r)
	plan.streets = tdGrowNetwork(plan.streets, plan.anchors, plan.cx, plan.cy, spanSlots, style, cfg, seed)

	// Junction lattice (playtest polish FIX 2: connected roads). The cross-streets now RUN
	// THROUGH main B (continuous, joined at each crossing), so the fabric must CLEAR those
	// crossings so no building sits on a junction. The lattice is centered on main B's actual
	// centerline (its midpoint) and steps along main B's axis (dirB = lane 1) by the same
	// street spacing tdCrossStreet uses, so each lattice point lands EXACTLY where a cross-
	// street crosses main B. It is count-independent (derived from the fixed mains geometry +
	// config), so the clearing stays stable-incremental. The crossings land on main B, leaving
	// main A (the near-1:1 low band) untouched.
	if len(plan.streets) > 1 {
		mb := plan.streets[1]
		dirB := unitDir(mb)
		mbMid := tdPolyMidpoint(mb) // main B's actual centerline center (offset off-core)
		perp := cfg.laneHalf + cfg.roofSize/2 + cfg.laneMargin
		streetGap := 2*perp + cfg.roofSize + cfg.minGap // same spacing tdGrowNetwork uses
		plan.junctions = tdJunctionLattice(mbMid, dirB, streetGap, cfg)
	}

	// (d) populate — the INTERMIXED, LANE-LINING fabric (FIX 2). Every non-wonder building
	// emits count-scaled roof lots that LINE ALONGSIDE the lanes (offset perpendicular to
	// the centerline, alternating sides so buildings flank a visible road), interleaved so
	// consecutive slots are different domains (a hut next to a camp next to a store, not a
	// blob of huts). Stability (locked #8): a lot's (lane, along-position, side, jitter) is
	// a pure function of (building type, instance index, seed) via a unique global frontage
	// slot — NOT of any other type's count or a shared cursor — so adding a building only
	// APPENDS and never moves an existing lot.
	tdPopulateIntermixed(&plan, blds, heroKey, cfg, seed)

	// Wonder centerpieces (FIX 2): each wonder sits AT its anchor as a dominant, ornate
	// complex with a CLEAR PLAZA around it (any fabric lot inside the plaza was dropped
	// in tdPopulateIntermixed), so the city hugs the wonder without ever burying it.
	tdPlaceWonders(&plan, cfg)

	// Final deterministic overlap guard: no two roof lots may sit closer than minGap. It
	// SKIPS (never nudges) a colliding fabric lot, yielding ONLY to the fixed wonder roofs
	// and to earlier SAME-DOMAIN lots — both count-stable — so a surviving lot keeps its
	// exact placed position (stable-incremental preserved). Cross-domain overlaps are
	// already prevented structurally (unique frontage slots + the cross-lane guard), so this
	// is a defensive net, not a mover. Runs AFTER the wonders exist so it can guard against
	// them.
	tdEnforceMinGap(&plan, cfg)

	// Town squares (playtest FIX): dress each cleared plaza (wonders + the wonderless
	// city-center) as a paved town square with a few seeded era props ringed around the
	// roof, so the open center reads as intentional rather than a bare-dirt donut.
	tdPlaceSquares(&plan, style, cfg, seed)

	// (e) filler — balanced gardens / squares / trees / props in the gaps.
	tdAddFiller(&plan, style, cfg, r)

	// (f) walls — a wall+gate ring IF the era has walls. Primitive: none. The ring is
	// sized to the lot bounding box so it hugs the built-up area.
	if style.hasWalls {
		tdAddWalls(&plan, r)
	}

	return plan
}

// ---- line-the-lanes placement (playtest FIX 2) ------------------------------
//
// Buildings LINE the lanes instead of being pulled onto them. Each lot sits ALONGSIDE a
// lane, offset perpendicular to the centerline so the road stays visible between two
// opposing rows. The whole lane network is flattened into ONE arc-length "frontage"
// line; each building instance is assigned a UNIQUE global frontage slot so no two lots
// ever coincide (structural no-overlap) and the assignment is a pure function of (type,
// instance index, seed) so it is exactly stable-incremental.

// tdFrontage is the lane network flattened into a single arc-length parameterisation.
// segs are the ordered city-space segments (concatenated across all lanes, contiguous in
// arc space: each seg's start = the running total before it), and total is the whole
// network's lineable length. lineAt(s) returns the point and the unit PERPENDICULAR at
// arc distance s. Space between lanes is NOT bridged — a lot never straddles two lanes.
type tdFrontage struct {
	segs  []tdFrontSeg
	total float64
}

type tdFrontSeg struct {
	a, b  tdPoint
	len   float64
	perpX float64 // unit perpendicular to the segment (points to the +side)
	perpY float64
	start float64 // cumulative arc length at a
	lane  int     // index of the source lane in streets (for the cross-lane overlap guard)
}

// buildFrontage concatenates every lane segment into one arc line. laneStartClear trims
// the crowded ends of each lane (near anchors/junctions) so the fabric only lines the
// diverged interiors — that keeps distinct lanes apart where buildings sit (structural
// cross-lane no-overlap) and leaves the junctions/plazas as clear road. Deterministic:
// depends only on the (already-stable) lane polylines and the config.
func buildFrontage(streets []tdStreet, startClear float64) tdFrontage {
	var f tdFrontage
	for laneIdx, s := range streets {
		if len(s.pts) < 2 {
			continue
		}
		// Whole-lane length, to trim startClear off EACH end.
		laneLen := 0.0
		for i := 0; i+1 < len(s.pts); i++ {
			laneLen += math.Hypot(s.pts[i+1].x-s.pts[i].x, s.pts[i+1].y-s.pts[i].y)
		}
		if laneLen <= 2*startClear+1e-6 {
			continue // too short to line once its ends are cleared
		}
		walked := 0.0
		for i := 0; i+1 < len(s.pts); i++ {
			a, b := s.pts[i], s.pts[i+1]
			segLen := math.Hypot(b.x-a.x, b.y-a.y)
			if segLen < 1e-6 {
				continue
			}
			// Clip this segment to the [startClear, laneLen-startClear] interior of the lane.
			lo := math.Max(0, startClear-walked)
			hi := math.Min(segLen, laneLen-startClear-walked)
			walked += segLen
			if hi-lo < 1e-6 {
				continue
			}
			ux, uy := (b.x-a.x)/segLen, (b.y-a.y)/segLen
			ca := tdPoint{a.x + ux*lo, a.y + uy*lo}
			cb := tdPoint{a.x + ux*hi, a.y + uy*hi}
			clen := hi - lo
			f.segs = append(f.segs, tdFrontSeg{
				a: ca, b: cb, len: clen,
				perpX: -uy, perpY: ux, // unit perpendicular
				start: f.total,
				lane:  laneIdx,
			})
			f.total += clen
		}
	}
	return f
}

// lineAt returns the point at arc distance s along the frontage, the unit perpendicular
// there, and the source LANE index (for the cross-lane overlap guard). ok is false if s is
// past the end (caller drops the lot). Pure over the fixed segments.
func (f tdFrontage) lineAt(s float64) (x, y, perpX, perpY float64, lane int, ok bool) {
	if len(f.segs) == 0 || s < 0 || s > f.total+1e-6 {
		return 0, 0, 0, 0, 0, false
	}
	// Linear scan (segment counts are small at village scale); could binary-search later.
	for _, sg := range f.segs {
		if s <= sg.start+sg.len+1e-9 {
			t := s - sg.start
			if t < 0 {
				t = 0
			}
			if t > sg.len {
				t = sg.len
			}
			ux := sg.b.x - sg.a.x
			uy := sg.b.y - sg.a.y
			l := math.Hypot(ux, uy)
			if l > 1e-9 {
				ux, uy = ux/l, uy/l
			}
			return sg.a.x + ux*t, sg.a.y + uy*t, sg.perpX, sg.perpY, sg.lane, true
		}
	}
	return 0, 0, 0, 0, 0, false
}

// tdTotalFabricSlots counts the total lane FRONTAGE SLOTS the fabric will consume — the
// sum over non-wonder types of their roof-lot counts. Used to size the town disc (tdTownRadius)
// so the footprint scales with the settlement. Pure over blds.
func tdTotalFabricSlots(blds []builtBuilding) int {
	total := 0
	for _, b := range blds {
		if b.category == "wonder" || b.category == "monument" {
			continue
		}
		total += tdRoofCount(b.count, b.role)
	}
	return total
}

// tdFrontageSpanSlots is the number of ROUND-ROBIN slot POSITIONS the fabric spans on the
// frontage line — maxN×T, where T is the number of fabric types and maxN the largest per-type
// roof-lot count. tdPopulateIntermixed places instance j of type-rank r at global slot r+j×T,
// so the LAST occupied slot is ~maxN×T even though only Σ(counts) lots actually exist (uneven
// type counts leave gaps). The lane network must provide frontage for this SPAN (not just the
// lot total) or the tail slots run off the end and their lots are dropped. Pure over blds.
func tdFrontageSpanSlots(blds []builtBuilding) int {
	T, maxN := 0, 0
	for _, b := range blds {
		if b.category == "wonder" || b.category == "monument" {
			continue
		}
		n := tdRoofCount(b.count, b.role)
		if n <= 0 {
			continue
		}
		T++
		if n > maxN {
			maxN = n
		}
	}
	return maxN * T
}

// tdFabricStep is the along-lane advance PER SLOT. Sides alternate every slot, so the two
// opposing rows each advance by 2 half-steps (= 2*tdFabricStep) between their own
// consecutive lots — i.e. same-side neighbours are 2*step apart along the lane. Sized
// generously from the roof extent + the min gap (padded for the widest roof, the elongated
// longhouse at 1.8×, PLUS extra headroom so that even where a winding lane BENDS — which
// compresses the spacing of the perpendicular-offset row on the inside of the curve — a
// same-side pair still never touches). The generous step is also what keeps the village
// AIRY (buildings loosely spaced along the streets, road clearly visible between rows).
func tdFabricStep(cfg tdConfig) float64 {
	// 2.4× the roof size gives loose, airy rows AND enough same-side headroom (same-side
	// neighbours sit 2 steps apart) to absorb both the per-slot jitter and the mild spacing
	// compression on the inside of a winding lane's bends, so a same-side pair never touches.
	return cfg.roofSize*2.0 + cfg.minGap
}

// tdPopulateIntermixed LINES the whole non-wonder fabric alongside the lanes (playtest FIX
// 2). It assigns every building instance a UNIQUE GLOBAL FRONTAGE SLOT via round-robin
// over the sorted-gather types: type rank r (0-based) instance j → slot = r + j*T (T =
// number of fabric types). Consecutive global slots are consecutive types, so the fabric
// is intermixed in BOTH placement order and space; and the slot→frontage map is a pure
// function of (type, j) — never of any other type's count or a shared cursor — so slot
// r+j*T never moves when a sibling grows: exact stable-incremental (locked #8). Growing a
// type only appends higher slots; a brand-new building TYPE is a layout event (as it
// already was under the old anchor-index scheme).
//
// Each slot maps to arc distance s = halfStep*slot along the frontage; the lot is offset
// PERPENDICULAR by (laneHalf + its half-extent + laneMargin) on the slot's side (parity),
// so it lines ALONGSIDE the road with the road cells free between opposing rows. A small
// bounded jitter keeps it organic without ever reaching the road or a neighbour. Lots
// past the frontage end, or inside a wonder plaza, are dropped (position-based → stable).
func tdPopulateIntermixed(plan *topPlan, blds []builtBuilding, heroKey string, cfg tdConfig, seed uint32) {
	if len(plan.anchors) == 0 {
		return
	}
	front := buildFrontage(plan.streets, cfg.laneStartClear*cfg.roofSize)
	plazaR := cfg.plazaRadius * cfg.roofSize
	halfStep := tdFabricStep(cfg) / 2
	// Junction clear radius (playtest polish FIX 2): drop any lot that would sit on a lane
	// crossing so the connected roads read as OPEN junctions. Sized to just clear the one roof
	// that would straddle a crossing (a roof sits ~perp off its own centerline) without eating
	// its along-lane neighbours (which sit ~2·halfStep away). Count-independent lattice → stable.
	junctionClear := cfg.laneHalf + cfg.roofSize/2 + cfg.laneMargin + cfg.minGap

	// The fabric types, in the stable sorted-gather order, each with a stable roof/size.
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
		types = append(types, fabType{b: b, roof: rt, sz: sz, n: n,
			label: b.key == heroKey && !plan.hasHero})
	}
	T := len(types)
	if T == 0 {
		return
	}

	// Max instance count over all types → how many rounds of the round-robin to run.
	maxN := 0
	for _, ft := range types {
		if ft.n > maxN {
			maxN = ft.n
		}
	}

	// ROUND-ROBIN over rounds j=0..maxN-1, types r=0..T-1: slot = r + j*T. This both
	// interleaves plan.lots (consecutive slots = consecutive domains) and gives each lot a
	// unique global slot whose frontage position is a pure fn of (type, j).
	for j := 0; j < maxN; j++ {
		for r := 0; r < T; r++ {
			ft := types[r]
			if j >= ft.n {
				continue
			}
			slot := r + j*T
			side := 1.0
			if slot%2 == 1 {
				side = -1.0
			}
			s := halfStep * float64(slot)
			x0, y0, perpX, perpY, lane, ok := front.lineAt(s)
			if !ok {
				continue // ran past the available frontage (extension sizes to avoid this)
			}
			half := ft.sz / 2
			if ft.roof == roofLong {
				half = ft.sz * 1.8 / 2
			}
			perp := cfg.laneHalf + half + cfg.laneMargin
			// Bounded ORGANIC jitter (never enough to overlap). Longitudinal wander is capped
			// so that even if two same-side neighbours BOTH wander toward each other (they sit
			// 2 half-steps apart) they still keep the min gap — with extra safety headroom to
			// also absorb the mild spacing compression on the inside of a lane's bends. The
			// perpendicular wander only pushes the lot FURTHER from the road (outward), so a
			// roof can never creep onto the lane. Pure fn of (slot, seed) → stable.
			halfStepLocal := tdFabricStep(cfg) / 2
			jlMax := (halfStepLocal - half - cfg.minGap/2) * 0.55
			if jlMax < 0 {
				jlMax = 0
			}
			jl, jp := slotJitter(slot, 0, seed, cfg.jitterAmp)
			jl = clampAbs(jl, jlMax)
			jp = math.Abs(jp) * 0.35 // outward only, gentle
			ux, uy := perpY, -perpX  // unit tangent (perp rotated back)
			x := x0 + perpX*side*(perp+jp) + ux*jl
			y := y0 + perpY*side*(perp+jp) + uy*jl
			if insideWonderPlaza(x, y, plan.anchors, plazaR) {
				continue
			}
			// Clear the lane junctions (playtest polish FIX 2): no building sits on a crossing, so
			// the connected roads read as open junctions. The crossings land on main B, so this
			// clears main B's row (lane 1) AND the cross-streets' rows (lane ≥ 2) at each junction
			// column, but LEAVES main A (lane 0) intact — main A holds the near-1:1 low band and
			// carries no crossings, so its exact count is preserved. Position-based over the count-
			// INDEPENDENT lattice → stable-incremental: the same at every building count, so a later
			// cross-street can never retroactively drop an existing lot (upper-clearing would → unstable).
			if lane > 0 && insideJunction(x, y, plan.junctions, junctionClear) {
				continue
			}
			// Cross-lane overlap guard (playtest FIX 2, stable): drop this lot if it falls too
			// close to a LOWER-INDEX lane's centerline — i.e. where its lane converges with or
			// crosses an earlier lane, only the higher-index lane's row is cleared, so two
			// lanes' opposing rows never collide at a junction. This is a pure function of the
			// lot's position and the (append-only, count-stable) lower-index lanes — NEVER of
			// any other lot — so it preserves EXACT stable-incremental placement (a sibling's
			// count can't change which lower lanes exist, hence can't change this decision),
			// which a lot-vs-lot skip could not (see tdEnforceMinGap). The MAINS are lanes 0,1
			// (always present); the compact grid's cross-streets now RUN THROUGH main B at cleared
			// junctions (tdCrossStreet + the junction lattice clears main B's row there), so a main's
			// own row is never buried at a crossing, and the lower-only asymmetry here can't leave a
			// main lot stranded on a road.
			if lane > 0 && nearLowerLane(x, y, plan.streets, lane, perp+2*half+cfg.minGap) {
				continue
			}
			lot := tdLot{
				x: x, y: y, w: ft.sz, h: ft.sz, kind: tdRoof,
				domain: ft.b.domain, category: ft.b.category, tier: ft.b.tier, roof: ft.roof,
			}
			if ft.roof == roofLong {
				lot.w = ft.sz * 1.8 // longhouses/rowhouses are elongated
			}
			if ft.label && j == 0 {
				lot.label = ft.b.name
				lot.prom = prominenceOf(ft.b)
			}
			plan.lots = append(plan.lots, lot)
		}
	}
}

// clampAbs clamps v to [-m, m].
func clampAbs(v, m float64) float64 {
	if v > m {
		return m
	}
	if v < -m {
		return -m
	}
	return v
}

// nearLowerLane reports whether (x,y) lies within dist of the centerline of any lane with
// index < uptoLane. Used by the lane-lining placement to clear the higher-index lane's row
// where it converges with an earlier lane (the cross-lane overlap guard). Only lower-index
// lanes are consulted, and those are append-only/count-stable, so the result is a pure
// function of position + fixed geometry → the placement stays stable-incremental.
func nearLowerLane(x, y float64, streets []tdStreet, uptoLane int, dist float64) bool {
	d2 := dist * dist
	for li := 0; li < uptoLane && li < len(streets); li++ {
		s := streets[li]
		for i := 0; i+1 < len(s.pts); i++ {
			if distToSegSq(x, y, s.pts[i], s.pts[i+1]) < d2 {
				return true
			}
		}
	}
	return false
}

// distToSegSq is the squared distance from p to segment a→b.
func distToSegSq(px, py float64, a, b tdPoint) float64 {
	dx, dy := b.x-a.x, b.y-a.y
	l2 := dx*dx + dy*dy
	if l2 < 1e-9 {
		ex, ey := px-a.x, py-a.y
		return ex*ex + ey*ey
	}
	t := ((px-a.x)*dx + (py-a.y)*dy) / l2
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	cx, cy := a.x+t*dx, a.y+t*dy
	ex, ey := px-cx, py-cy
	return ex*ex + ey*ey
}

// tdTownRadius is the BOUNDED town-disc radius model (city units) for a settlement lining `slots`
// lots (playtest FIX: no pinwheel): townBaseRadius + townGrowth·√slots — it grows only ~SQRT with
// the count, so the footprint expands SLOWLY. The STREET network doesn't consume it directly (its
// per-band chords are count-INDEPENDENT so growth stays stable-incremental — see tdCrossStreet);
// it is the canonical statement of the compact bound the whole model targets, and the anti-
// pinwheel test asserts the real footprint stays within it. A floor keeps a tiny hamlet honest.
func tdTownRadius(slots int, cfg tdConfig) float64 {
	if slots < 1 {
		slots = 1
	}
	r := cfg.townBaseRadius + cfg.townGrowth*math.Sqrt(float64(slots))
	if r < cfg.townBaseRadius {
		r = cfg.townBaseRadius
	}
	return r
}

// tdMainHalfSpan is the FIXED half-length of the wonderless village's crossing MAIN streets,
// in city units. Sized so the FIRST main alone holds the entire near-1:1 low band (~12 slots ×
// the along-lane halfStep of frontage, plus the junction trim off each end) — so low counts
// line one clean main and never spill into the second main's crossing zone. Count-INDEPENDENT
// (a constant of the config), which is what keeps the mains stable as the town grows: only
// branches are appended, never resized. Compact + bounded.
func tdMainHalfSpan(cfg tdConfig) float64 {
	// The main is 2×halfSpan long; its lineable interior is (2×halfSpan − 2×trim). The near-1:1
	// low band is ~12 slots, needing ~12×halfStep of lineable frontage on this one main. Solve
	// for the half-span and add a little headroom so all 12 low-band lots fit comfortably.
	half := tdFabricStep(cfg) / 2
	trim := cfg.laneStartClear * cfg.roofSize
	lowBand := half * 12                      // frontage the 12-slot low band consumes
	span := (lowBand+2*trim)/2 + cfg.roofSize // + headroom
	// Keep it at least the base disc so a tiny village still has a real crossroads.
	if span < cfg.townBaseRadius {
		span = cfg.townBaseRadius
	}
	return span
}

// tdGrowNetwork densifies the lane network into a COMPACT, BOUNDED, CONNECTED set of rows (playtest
// FIX: replaces the radial-spoke tdExtendLanes that flung buildings outward into a pinwheel; polish
// FIX 2: the rows now JOIN the mains instead of floating). Off the two crossing MAINS it appends
// CROSS-STREETS that all run PARALLEL to main A (dirA), each offset perpendicular by a multiple of a
// street spacing (alternating sign in growing offset bands) and running CONTINUOUSLY THROUGH main B
// where they cross it — so every cross-street is LINKED into the network at a T/X junction on main B,
// and the whole graph is one connected fabric radiating from the central crossroads (not a set of
// disconnected floating segments). The parallels never cross EACH OTHER (same direction), so the only
// crossings are the mains at the center + each cross-street meeting main B; buildings are kept off
// those junctions by the count-independent junction lattice (see tdJunctionLattice). Every cross-
// street is a chord of a compact, BAND-bounded disc, so the footprint grows only slowly with the
// street count and stays a bounded village of streets flanking a crossroads, not a starburst.
//
// Growth is DENSER, NOT LONGER: adding buildings appends more parallel rows within a slowly-
// expanding, bounded footprint — never longer radial arms.
//
// Determinism + stable-incremental (locked #8): cross-street i's offset band and geometry are a
// pure function of (i, seed) — NEVER of the total count — and streets are APPENDED after the fixed
// mains, so existing lanes (and the lots already lining them) never move as slots (hence the street
// count) rises. Panic-safe: guarded loop, disc-clipped chords.
func tdGrowNetwork(streets []tdStreet, anchors []tdAnchor, cx, cy float64, spanSlots int, style tdEraStyle, cfg tdConfig, seed uint32) []tdStreet {
	if spanSlots <= 0 || len(streets) < 2 {
		return streets
	}
	spacing := cfg.branchSpacing
	if spacing <= 0 {
		spacing = tdFabricStep(cfg) / 2
	}
	// Frontage the fabric needs: the max round-robin slot arc (spacing × the frontage SPAN, i.e.
	// maxN×T — see tdFrontageSpanSlots) plus headroom for the per-lane junction trims + a margin
	// so the tail slots don't have to reach the very end of a lane. Sizing to the SPAN (not the
	// lot total) is what keeps the tail lots on the frontage instead of running off the end.
	need := spacing * float64(spanSlots) * 1.3
	frontageLen := func(ss []tdStreet) float64 {
		return buildFrontage(ss, cfg.laneStartClear*cfg.roofSize).total
	}

	// The two crossing mains define the row directions. Read their unit vectors from their
	// endpoints (count-independent geometry, so the axes are stable).
	dirA := unitDir(streets[0])
	dirB := unitDir(streets[1])

	// The perpendicular STREET SPACING between two adjacent parallel rows: two opposing rows of
	// buildings (~perp each) plus a clear gap, so buildings line the corridor between parallels
	// without the rows of neighbouring streets overlapping.
	perp := cfg.laneHalf + cfg.roofSize/2 + cfg.laneMargin
	streetGap := 2*perp + cfg.roofSize + cfg.minGap

	guard := 0
	i := 0
	misses := 0
	for frontageLen(streets) < need && guard < 200 {
		guard++
		st, ok := tdCrossStreet(cx, cy, dirA, dirB, streetGap, i, style, cfg, seed)
		i++
		if !ok {
			// The outermost band hit the compact ceiling → the footprint is full; stop rather than
			// spin. Allow a couple of misses first (the ± signs of a band reach the edge one apart).
			misses++
			if misses > 3 {
				break
			}
			continue
		}
		misses = 0
		streets = append(streets, st...)
	}
	return streets
}

// tdCrossStreet builds one CROSS-STREET of the loose grid as ONE CONTINUOUS lane that CROSSES
// main A — a lane PARALLEL to main B, offset along main A by a whole number of streetGap BANDS,
// running the FULL chord THROUGH the main-A centerline so the street actually JOINS the network
// at that junction (playtest polish FIX 2: CONNECTED roads — no more split-short floating halves).
// The family, offset band and sign are a pure function of the street index i (and seed); the lane
// is a CHORD of a band-sized disc that just contains the band's offset — so inner bands are near-
// full-length streets and each outer band extends the footprint by exactly one gap: the grid
// DENSIFIES within a compact, roughly-round footprint that grows only one band at a time (never a
// radial arm). Geometry depends only on the band (i.e. i), so appending streets never moves an
// existing lane (stable-incremental). Returns ok=false past the compact ceiling (the town is full).
//
// Buildings are kept OFF the resulting T/X junction by the count-INDEPENDENT junction lattice
// (tdJunctionLattice + insideJunction), which clears a small radius around every main-A crossing
// symmetrically — including main A's OWN row, which the lower-only overlap guard could not clear
// stably. That is what lets the cross-streets reach and touch the mains (a connected graph) while
// still keeping the junction cells clear and the placement stable-incremental.
func tdCrossStreet(cx, cy float64, dirA, dirB tdPoint, streetGap float64, i int, style tdEraStyle, cfg tdConfig, seed uint32) ([]tdStreet, bool) {
	// ONE family of parallels: every cross-street runs PARALLEL to main B (dirB), offset along main
	// A by ±band·gap. Parallels never cross each other (same direction, different offset); each
	// CROSSES main A once, at the junction lattice point cleared of buildings — so the streets form
	// one CONNECTED network radiating from the central crossroads (mains 0,1) rather than a set of
	// disconnected floating segments. The band grows every OTHER street and the sign alternates, so
	// offsets march +1,-1,+2,-2,… out from the spine — a symmetric, compact set of rows flanking
	// the crossroads, each linked into the graph where it meets the main.
	band := i/2 + 1 // 1,1,2,2,3,3,…: the |offset| in street-gap units (never 0 = the main)
	sign := 1.0
	if i%2 == 1 {
		sign = -1.0
	}
	off := sign * float64(band) * streetGap
	// Cross-streets run PARALLEL to main A and cross main B (run=dirA, offAxis=dirB), so their
	// junctions land on main B — NOT on main A, which carries the near-1:1 low band (tdMainHalfSpan
	// sizes main A to hold it). Keeping the crossings off main A is what lets the low-band count stay
	// EXACT while the network is still fully connected (the cross-streets join the mains at main B).
	run, offAxis := dirA, dirB
	// The footprint disc that must contain this band: the mains' half-span, extended just enough
	// to hold this band's offset (so the town grows one gap per band, staying roughly round and
	// compact — never flung out). Count-independent (a function of the band only).
	discR := tdMainHalfSpan(cfg)
	if need := math.Abs(off) + streetGap; need > discR {
		discR = need
	}
	// A hard compact CEILING so no band can ever fling the town out (the anti-pinwheel bound).
	if ceil := cfg.townBaseRadius + cfg.townGrowth*7; discR > ceil {
		return nil, false
	}
	if math.Abs(off) >= discR-cfg.roofSize {
		return nil, false
	}
	// The lane is a chord of that disc at perpendicular distance |off|: half-length sqrt(R²−off²),
	// centered on (mx,my) = core + off·offAxis, which lies ON main A. Drawn as ONE continuous lane
	// from −half to +half along run THROUGH (mx,my), so it reaches and crosses the main (connected),
	// leaving no gap. Too short to bother → drop.
	half := math.Sqrt(discR*discR - off*off)
	if 2*half < cfg.roofSize*2 {
		return nil, false
	}
	mx, my := cx+offAxis.x*off, cy+offAxis.y*off
	p0 := tdPoint{mx - run.x*half, my - run.y*half}
	p1 := tdPoint{mx + run.x*half, my + run.y*half}
	sr := newRNG(hash2(uint32(i)*263+7, 0x5b0c, seed) | 1)
	return []tdStreet{windingLane(p0, p1, style.streetJitter, style.laneWidth, sr)}, true
}

// tdJunctionLattice returns the count-INDEPENDENT set of CROSS-STREET junctions (playtest polish
// FIX 2): every point where a cross-street band crosses main B — origin ± runDir·(band·streetGap)
// for band = 1,2,… out to the compact ceiling, where origin is main B's centerline midpoint and
// runDir is main B's axis. It enumerates ALL bands the compact bound permits, NOT just the cross-
// streets that currently exist, so the lattice is a pure function of the fixed mains geometry +
// config and NOTHING count-dependent. That matters for stability: a cross-street appears only when
// the town is big enough, but by clearing its crossing column UP FRONT (before the street exists),
// a later cross-street can never retroactively bury a main-B lot that used to sit there — the
// keep/drop stays a pure function of position, so placement is stable-incremental. The regular
// pre-cleared gaps on main B are exactly where cross-streets land; in a grown town every gap is a
// real junction. The CENTRAL crossroads (mains 0,1) is deliberately EXCLUDED — it is handled by the
// lower-only overlap guard (nearLowerLane), so a wonderless village keeps a FILLED heart (no plaza
// clear) rather than a punched-out center.
func tdJunctionLattice(origin tdPoint, runDir tdPoint, streetGap float64, cfg tdConfig) []tdPoint {
	if streetGap <= 0 {
		return nil
	}
	var out []tdPoint
	ceil := cfg.townBaseRadius + cfg.townGrowth*7 // the same hard compact bound tdCrossStreet uses
	for band := 1; ; band++ {
		off := float64(band) * streetGap
		if off >= ceil-cfg.roofSize {
			break
		}
		out = append(out,
			tdPoint{origin.x + runDir.x*off, origin.y + runDir.y*off},
			tdPoint{origin.x - runDir.x*off, origin.y - runDir.y*off},
		)
	}
	return out
}

// tdPolyMidpoint returns the midpoint of a polyline's first and last points — the center of a
// (nearly-straight) main lane. Used to anchor the junction lattice on main B's actual centerline
// (which sits a touch off-core), so each lattice point lands exactly where a cross-street crosses.
func tdPolyMidpoint(s tdStreet) tdPoint {
	if len(s.pts) == 0 {
		return tdPoint{}
	}
	a, b := s.pts[0], s.pts[len(s.pts)-1]
	return tdPoint{(a.x + b.x) / 2, (a.y + b.y) / 2}
}

// insideJunction reports whether (x,y) lies within clearR of any lane junction. Used by the
// lane-lining placement to keep the junction cells building-free so the connected roads read as
// open crossings, not buried ones. Pure over the (count-independent) lattice → stable.
func insideJunction(x, y float64, junctions []tdPoint, clearR float64) bool {
	if clearR <= 0 {
		return false
	}
	r2 := clearR * clearR
	for _, j := range junctions {
		dx, dy := x-j.x, y-j.y
		if dx*dx+dy*dy < r2 {
			return true
		}
	}
	return false
}

// unitDir returns the unit direction of a street from its first point to its last (the overall
// axis of a main), or (1,0) for a degenerate street. Used to read the loose grid's two axes off
// the crossing mains.
func unitDir(s tdStreet) tdPoint {
	if len(s.pts) < 2 {
		return tdPoint{1, 0}
	}
	a, b := s.pts[0], s.pts[len(s.pts)-1]
	dx, dy := b.x-a.x, b.y-a.y
	l := math.Hypot(dx, dy)
	if l < 1e-9 {
		return tdPoint{1, 0}
	}
	return tdPoint{dx / l, dy / l}
}

// tdEnforceMinGap is the final deterministic overlap guard (playtest FIX 2): no two roof
// lots may sit closer than the min gap. It SKIPS (never nudges) a colliding lot, and a lot
// only ever yields to lots that are guaranteed present regardless of any OTHER type's
// count — its own SAME-DOMAIN earlier lots and the fixed WONDER roofs — so a surviving
// lot keeps its exact placed position and the survivor set stays a stable per-domain
// prefix (locked #8). Cross-domain overlaps are prevented structurally by the unique
// global frontage slots + cleared junctions, so this pass has nothing cross-domain to do
// in a well-formed plan; it is a safety net for degenerate geometry, not a mover.
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
	// (anchor, seed), independent of any building count), so gather them first and check
	// every fabric lot against them regardless of slice order.
	var wonders []tdLot
	for _, lt := range plan.lots {
		if lt.kind == tdRoof && lt.roof == roofWonder {
			wonders = append(wonders, lt)
		}
	}
	kept := make([]tdLot, 0, len(plan.lots))
	byDom := map[string][]tdLot{} // per-domain already-kept fabric lots (stable same-domain check)
	for _, lt := range plan.lots {
		if lt.kind != tdRoof || lt.roof == roofWonder {
			kept = append(kept, lt) // non-roofs and the wonders themselves always stay
			continue
		}
		skip := false
		// Yield only to the fixed wonders and to earlier SAME-DOMAIN lots — both count-stable,
		// so a sibling domain growing can never change this lot's keep decision.
		for _, w := range wonders {
			if overlaps(lt, w) {
				skip = true
				break
			}
		}
		if !skip {
			for _, s := range byDom[lt.domain] {
				if overlaps(lt, s) {
					skip = true
					break
				}
			}
		}
		if skip {
			continue
		}
		byDom[lt.domain] = append(byDom[lt.domain], lt)
		kept = append(kept, lt)
	}
	plan.lots = kept
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

// tdWonderScale is the wonder roof extent as a multiple of roofSize, by anchor index: the
// grandest wonder (anchor 0) is the largest, the rest a touch smaller so the centre reads
// as the primary showpiece. Kept well UNDER the plaza radius (plazaRadius 2.2) so a paved
// ring always shows around the roof inside its cleared plaza, even after fill-frame shrinks
// the city at high building counts (the town-square-dressing invariant). Single source of
// truth shared by tdPlaceWonders (the roof) and tdPlaceSquares (the prop ring + apron).
func tdWonderScale(anchorIdx int) float64 {
	if anchorIdx == 0 {
		return 2.6
	}
	return 2.2
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

// ---- town square (playtest FIX: dress the cleared plaza) --------------------
//
// The wonder-anchored primitive village reads as a ring around an EMPTY plaza (a
// donut). Rather than SHRINK the open center, we DRESS it: each plaza-clearing anchor
// (the wonders + the wonderless city-center) gets a deliberate TOWN SQUARE — a paved-
// stone ground patch under the roof plus a few seeded, era-appropriate props arranged
// AROUND (never overlapping) the roof. The openness stays; it now reads as intentional.

// tdSquareProps is one era's town-square prop palette: the prop-lot kinds a wonder
// square is dressed with (full set) and the modest kinds a wonderless center gets (a
// couple). Kept as a per-era value so LATER eras can swap the set (fountain/statue/
// benches) by adding a case in tdSquarePropsFor — only the PRIMITIVE set is tuned now.
type tdSquareProps struct {
	wonder []tdLotKind // props ringed around a wonder centerpiece (the full square)
	center []tdLotKind // props for a wonderless city-center's modest square
}

// tdSquarePropsFor returns the town-square prop palette for an era. PRIMITIVE (organic):
// a well, a firepit, standing stones/totem, and a market stall around a wonder; a well +
// firepit for a bare center. Every other era falls back to the primitive set until
// V3-B/C/D tunes its own (fountains, statues, benches) — framework groundwork only.
func tdSquarePropsFor(style tdEraStyle) tdSquareProps {
	// Only the organic/primitive set is tuned; the default preset reuses it for now.
	return tdSquareProps{
		wonder: []tdLotKind{tdPropWell, tdPropFirepit, tdPropStones, tdPropStall},
		center: []tdLotKind{tdPropWell, tdPropFirepit},
	}
}

// tdPlaceSquares dresses every plaza-clearing anchor as a town square (playtest FIX).
// For each WONDER anchor it lays a paved-stone plaza patch (filling the cleared plaza
// radius, drawn under the wonder roof) and rings a few era props AROUND the wonder roof
// — seeded, deterministic, arranged on a circle strictly OUTSIDE the roof footprint and
// inside the plaza so none overlaps the centerpiece. The wonderless city-center anchor
// (which does NOT clear a big plaza) gets a MODEST square — a small paved patch + a
// well/firepit — so the heart still reads as a gathering place without hollowing a hut
// village into a donut. Pure + seeded → stable-incremental (positions are a function of
// (anchor index, seed), never of any building count).
func tdPlaceSquares(plan *topPlan, style tdEraStyle, cfg tdConfig, seed uint32) {
	if len(plan.anchors) == 0 {
		return
	}
	props := tdSquarePropsFor(style)
	plazaR := cfg.plazaRadius * cfg.roofSize
	for i, a := range plan.anchors {
		if a.wonder {
			// The wonder roof half-extent (matches tdPlaceWonders via tdWonderScale). Props
			// ring OUTSIDE this so the centerpiece is never covered.
			roofHalf := cfg.roofSize * tdWonderScale(i) / 2
			// Paved plaza patch: fills the whole cleared plaza radius (a square lot whose
			// half-extent is the plaza radius) so the open center reads as a made surface. The
			// wonder roof (see tdWonderScale) is kept clear of the plaza RIM so a paved ring
			// always shows around it even after fill-frame shrinks the plaza at high counts.
			plan.lots = append(plan.lots, tdLot{
				x: a.cx, y: a.cy, w: plazaR * 2, h: plazaR * 2, kind: tdPlaza,
			})
			// Ring the props around the roof, on a circle between the roof edge and the
			// plaza rim. A stable per-anchor angular phase so squares don't all align.
			ringR := (roofHalf + plazaR) / 2
			if ringR < roofHalf+cfg.roofSize*0.6 {
				ringR = roofHalf + cfg.roofSize*0.6 // keep clear of the roof even if the band is tight
			}
			tdRingProps(plan, a.cx, a.cy, ringR, props.wonder, uint32(i), seed)
		} else {
			// Wonderless center: a MODEST square (small paved patch + a couple of props),
			// kept small so a tiny village keeps a filled heart, not a donut.
			smallR := plazaR * 0.55
			plan.lots = append(plan.lots, tdLot{
				x: a.cx, y: a.cy, w: smallR * 2, h: smallR * 2, kind: tdPlaza,
			})
			// A well + firepit tucked just off-center, inside the small patch.
			tdRingProps(plan, a.cx, a.cy, smallR*0.55, props.center, uint32(i), seed)
		}
	}
}

// tdRingProps places one lot per prop kind evenly around a circle of radius ringR about
// (cx,cy), with a small seeded angular phase (from anchor index di) so different squares
// aren't rotationally identical. Each prop lot is small (well under the roof size) so it
// reads as a dab of detail, not a structure. Deterministic: positions are a pure
// function of (di, seed, prop index) — no threaded rng — so the square is stable-
// incremental like the rest of the plan.
func tdRingProps(plan *topPlan, cx, cy, ringR float64, kinds []tdLotKind, di, seed uint32) {
	n := len(kinds)
	if n == 0 || ringR <= 0 {
		return
	}
	phase := float64(hash2(di*131+5, 0x50a2, seed)) / float64(^uint32(0)) * 2 * math.Pi
	for k, kind := range kinds {
		ang := phase + 2*math.Pi*float64(k)/float64(n)
		// A whisper of per-prop radial jitter (seeded) so the ring isn't a perfect stamp.
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

// tdOrganicStreets lays the MAIN lanes the settlement grows ALONG (FIX 1) — the compact,
// count-INDEPENDENT bones tdGrowNetwork later densifies with branches (playtest FIX: no
// pinwheel). It ALWAYS lays two long, gently-meandering MAIN streets that CROSS near the core
// — a village crossroads that gives long, CLEAN frontage (so the low-count fabric lines them
// with near-zero crossings and lands near-1:1) and reads the heart at their crossing. Then,
// when the civ has wonder anchors, it weaves each anchor in with a short connector lane from
// the anchor toward the core, and chains consecutive anchors — so the anchors read as part of
// one woven settlement without turning the town into radial spokes. The crossing mains span a
// FIXED length (tdMainHalfSpan), so growing the count never resizes them; only branches are
// appended. Purely jittered polylines: NO terrain routing, no water (locked #10, V3-A organic).
// Grid/avenue patterns can slot in later; only organic is tuned now.
func tdOrganicStreets(anchors []tdAnchor, cx, cy float64, cfg tdConfig, style tdEraStyle, r *rng) []tdStreet {
	core := tdPoint{cx, cy}
	streets := make([]tdStreet, 0, len(anchors)*2+2)

	// The two crossing MAINS (lanes 0,1) — long, clean, count-independent. The fabric fills
	// the frontage in lane order, so lining these first keeps the low band near-1:1 on clean
	// lanes before any branch. The second main is offset a touch off-center so the two read as
	// a crossroads, not one lane doubled on the other.
	phase := r.f01() * 2 * math.Pi
	mainLen := tdMainHalfSpan(cfg)
	s0 := tdPoint{cx - math.Cos(phase)*mainLen, cy - math.Sin(phase)*mainLen}
	s1 := tdPoint{cx + math.Cos(phase)*mainLen, cy + math.Sin(phase)*mainLen}
	streets = append(streets, windingLane(s0, s1, style.streetJitter, style.laneWidth, r))
	ang2 := phase + math.Pi/2
	ox := math.Cos(phase) * cfg.roofSize * 1.5
	oy := math.Sin(phase) * cfg.roofSize * 1.5
	c0 := tdPoint{cx + ox - math.Cos(ang2)*mainLen, cy + oy - math.Sin(ang2)*mainLen}
	c1 := tdPoint{cx + ox + math.Cos(ang2)*mainLen, cy + oy + math.Sin(ang2)*mainLen}
	streets = append(streets, windingLane(c0, c1, style.streetJitter, style.laneWidth, r))

	// Anchor connectors: only for a wonder anchor that sits FAR from the crossing mains, a single
	// short lane from the anchor to the CORE weaves it into the network so it isn't stranded.
	// Anchors close to the mains (the common case — anchorSpread is kept well inside the town)
	// already sit on the grid and get NO connector, so we don't pile crossing lanes into the
	// compact fabric (which would drop/bury the rows). Consecutive anchors are NOT chained, for
	// the same reason. These come AFTER the mains so the mains stay lanes 0,1 (clean low band).
	//
	// The connector reaches the CORE (the central crossroads), not a point short of it (playtest
	// polish FIX 2: CONNECTED roads — a connector that stopped short was a floating stub). The
	// central junction is cleared of buildings by the lattice, so joining there buries nothing.
	_ = core
	connectThresh := cfg.roofSize * 4
	for _, a := range anchors {
		d := math.Hypot(a.cx-cx, a.cy-cy)
		if d < connectThresh {
			continue // already on/near the crossroads — no extra lane needed
		}
		streets = append(streets, windingLane(tdPoint{a.cx, a.cy}, tdPoint{cx, cy}, style.streetJitter, style.laneWidth, r))
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
	// Ponds (playtest polish FIX 4): a FEW BUILT decorative pools mixed IN-TOWN among the
	// gardens/greenery — a made water feature, not scattered natural water. Count is small
	// and sub-linear (a village gets 1–3, a big town a couple more) so ponds stay a rare
	// accent, never a lake district. Placed like gardens: seeded points within the built-up
	// disk (innerRad), so they read as part of the town fabric. Kept a touch smaller than a
	// garden so a pond nestles among the greenery rather than dominating it.
	ponds := 1 + int(dens*math.Sqrt(float64(roofN))*0.18)
	if ponds > 5 {
		ponds = 5 // a few, never many
	}
	for i := 0; i < ponds; i++ {
		x, y := tdDiskPoint(plan.cx, plan.cy, innerRad, r)
		plan.lots = append(plan.lots, tdLot{x: x, y: y, w: cfg.roofSize * 1.1, h: cfg.roofSize * 0.9, kind: tdPond})
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
	scale       float64
	offX, offY  float64
	minX, minY  float64
	roofFloorPx float64 // minimum roof extent in px (legibility floor, locked #3)
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

	// Streets (BOLD packed-earth lanes, FIX 1) under the fabric — a strong visual element
	// that stays visible because the roofs LINE alongside the lanes (FIX 2), not on top.
	streetCol := style.streetCol(pal)
	streetEdge := tdStreetEdgeColor(style, pal)
	for _, s := range plan.streets {
		drawTdStreet(img, xf, s, streetCol, streetEdge)
	}

	// Ground accents: gardens, squares, and the TOWN-SQUARE paved plazas painted before
	// roofs so a roof sits on top. The plaza is drawn as a rounded paved apron under the
	// wonder/center roof (a made surface, distinct from the era-tinted dirt).
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
			// BUILT decorative pond (FIX 4): a small water blob woven into the greenery.
			drawPond(img, xf, lt, pondCol)
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
	// The typed town-square props (well / firepit / standing stones / stall) dispatch to
	// their own small draw routines; the generic tdProp filler stays a quiet dab.
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

// drawGround fills the whole canvas with the era ground tone plus a subtle, seeded
// texture: each pixel picks base or a slightly-varied alt tone from a cheap hash of
// its coordinates, so the dirt reads as textured earth rather than a flat wash. No
// water, no biome — a neutral era-tinted ground (locked #2).
//
// CALM GROUND (playtest polish FIX 1): the texture is a WHISPER now, not a busy
// speckle. The old cut speckled ~22% of pixels and blended up to 0.6 of the way to
// the (noticeably different) alt tone, so the dirt fizzed and competed with the city
// — you couldn't cleanly SEE the village against it. Two dials pull the contrast
// down hard: (1) far fewer pixels are touched (groundTexFrac), and (2) the max blend
// toward alt is a small fraction (groundTexAmp) so even a touched pixel barely
// shifts. The result is a subtle-but-present dirt base that lets buildings + roads
// stand out. Still fully theme-derived (base/alt are style recipes) and seeded.
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
			// Cheap 2D value hash → [0,1); only a small fraction of pixels take any alt
			// tint at all, and that tint tops out at a low amplitude, so the texture is a
			// faint grain rather than a busy speckle that competes with the fabric.
			n := texHash(uint32(x), uint32(y), seed)
			t := 0.0
			if n < groundTexFrac {
				t = 0.5 + 0.5*n/groundTexFrac // 0.5..1.0 within the touched band
			}
			img.SetRGBA(px, py, blend(base, alt, t*groundTexAmp))
		}
	}
}

// Ground-texture calm dials (playtest polish FIX 1). Lowering EITHER quiets the dirt;
// together they drop the ground variance/contrast well below the old busy speckle so
// the village reads cleanly against a subtle base. groundTexFrac is the fraction of
// pixels that take ANY alt tint (was 0.22); groundTexAmp caps how far a touched pixel
// blends toward the alt tone (was 0.60). Kept as named consts so the quieter-ground
// test can reason about the reduction and a future era can retune without a magic number.
const (
	groundTexFrac = 0.12
	groundTexAmp  = 0.20
)

// texHash is a tiny deterministic 2D value hash returning a float in [0,1). Used for
// the ground texture speckle — cheap, seeded, and stable frame-to-frame.
func texHash(x, y, seed uint32) float64 {
	h := seed + x*374761393 + y*668265263
	h = (h ^ (h >> 13)) * 1274126177
	h ^= h >> 16
	return float64(h&0xffffff) / float64(0x1000000)
}

// drawTdStreet rasterizes a city-space lane polyline into pixels as a BOLD packed-earth
// path (playtest FIX 1). The band is CENTERED on the centerline and stroked PERPENDICULAR
// to each segment so it reads as an even, deliberate road of half-width s.width (width 0 →
// 1px; width 1 → ~3px), with a crisp darker EDGE one cell further out for a defined
// shoulder. The polyline is mapped through the fill-frame transform first, then drawn with
// the shared Bresenham rasterizer (reused; no terrain routing involved). Lanes draw UNDER
// the roofs but stay visible because the fabric now LINES alongside them, not on top.
func drawTdStreet(img *image.RGBA, xf tdTransform, s tdStreet, surface, edge color.RGBA) {
	if len(s.pts) < 2 {
		return
	}
	for i := 0; i+1 < len(s.pts); i++ {
		ax, ay := xf.px(s.pts[i].x, s.pts[i].y)
		bx, by := xf.px(s.pts[i+1].x, s.pts[i+1].y)
		// Perpendicular unit vector in pixel space, for a centered band (offset both ways).
		dx, dy := float64(bx-ax), float64(by-ay)
		l := math.Hypot(dx, dy)
		var pxu, pyu float64
		if l > 1e-6 {
			pxu, pyu = -dy/l, dx/l
		}
		off := func(k int) (int, int, int, int) {
			ox := int(math.Round(pxu * float64(k)))
			oy := int(math.Round(pyu * float64(k)))
			return ax + ox, ay + oy, bx + ox, by + oy
		}
		// Darker shoulder one cell PAST the band on each side (drawn first, so the surface
		// overpaints any overlap and the edge only shows on the true rim).
		for _, k := range []int{s.width + 1, -(s.width + 1)} {
			x0, y0, x1, y1 := off(k)
			drawRoad(img, roadSeg{x0, y0, x1, y1}, edge)
		}
		// The packed-earth surface band, centered: k in [-width, width].
		for k := -s.width; k <= s.width; k++ {
			x0, y0, x1, y1 := off(k)
			drawRoad(img, roadSeg{x0, y0, x1, y1}, surface)
		}
	}
}

// tdStreetEdgeColor resolves the lane shoulder tone for a style: the preset's streetEdge
// recipe when set, else a derived fallback (the surface darkened) so an era whose preset
// predates the field still gets a crisp edge. Pure theme read → retints on a switch.
func tdStreetEdgeColor(style tdEraStyle, pal tdPal) color.RGBA {
	if style.streetEdge != nil {
		return style.streetEdge(pal)
	}
	return darken(style.streetCol(pal), 0.22)
}

// ---- town-square paving + props (playtest FIX) ------------------------------

// tdPavedColor resolves the town-square paving tone for a style. It uses the preset's
// pavedCol recipe when set; otherwise it derives a safe fallback (the square tone lifted
// toward the light neutral) so an era whose preset predates this field still paves. Pure
// theme read → retints on a theme switch like every other tone.
func tdPavedColor(style tdEraStyle, pal tdPal) color.RGBA {
	if style.pavedCol != nil {
		return style.pavedCol(pal)
	}
	return blend(style.squareCol(pal), pal.text, 0.30)
}

// tdPondColor resolves the BUILT-pond water tone for a style (playtest polish FIX 4). It
// uses the preset's pondCol recipe when set; otherwise it derives a safe fallback (the
// theme background pulled toward the muted water anchor) so an era whose preset predates
// the field still paints a pond. Pure theme read → retints on a theme switch like every
// other tone. This is decorative BUILT water, never natural terrain water.
func tdPondColor(style tdEraStyle, pal tdPal) color.RGBA {
	if style.pondCol != nil {
		return style.pondCol(pal)
	}
	return blend(blend(pal.bg, pal.dim, 0.18), waterAnchor, 0.55)
}

// drawPlaza paints the paved town-square ground as a rounded apron: a filled ellipse in
// the paving tone with a faintly darker rim so the made surface reads with a soft edge
// against the surrounding dirt, rather than a hard rectangle. Drawn under the wonder/
// center roof + props. Half-extents are floored so a small center square still shows.
func drawPlaza(img *image.RGBA, cx, cy, hw, hh int, paved color.RGBA) {
	if hw < 1 {
		hw = 1
	}
	if hh < 1 {
		hh = 1
	}
	forEllipse(cx, cy, hw, hh, func(x, y int) { setPixel(img, x, y, paved) })
	// A whisper-darker rim so the paved apron reads as an edged surface, not a flat wash.
	rim := darken(paved, 0.12)
	forEllipse(cx, cy, hw, hh, func(x, y int) {
		fx := float64(x-cx) / float64(hw)
		fy := float64(y-cy) / float64(hh)
		if fx*fx+fy*fy >= 0.72 { // outer ring band only
			setPixel(img, x, y, rim)
		}
	})
}

// drawSquareProp renders one town-square prop as a small top-down dab, dispatched by its
// lot kind. Tones come from the era style (the paving + prop recipes) so props retint
// with the theme and stay in the era mood. Kept tiny — a well/firepit/totem/stall reads
// as a detail dressing the square, never competing with the wonder roof. All radii are
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
		// Stone well head: a pale stone ring (paving-toned) with a dark shaft mouth.
		fillDisc(img, cx, cy, rad, brighten(paved, 0.10))
		setPixel(img, cx, cy, darken(prop, 0.55)) // the dark shaft
	case tdPropFirepit:
		// Firepit: a charred dark ring with a warm ember center.
		fillDisc(img, cx, cy, rad, darken(prop, 0.45))
		ember := brighten(blend(prop, color.RGBA{R: 0xc8, G: 0x5a, B: 0x1e, A: 0xff}, 0.6), 0.10)
		setPixel(img, cx, cy, ember)
	case tdPropStones:
		// Standing stones / totem: two or three upright dabs in the prop (stone) tone, so
		// it reads as a little megalith cluster rather than a single block.
		stone := blend(prop, paved, 0.35)
		drawBlock(img, cx-rad, cy, 0, stone)
		drawBlock(img, cx, cy-rad, 0, brighten(stone, 0.10))
		drawBlock(img, cx+rad, cy, 0, darken(stone, 0.10))
		setPixel(img, cx, cy, stone)
	case tdPropStall:
		// Market stall: a small awning patch — a filled square in a warm cloth tone with a
		// lighter top edge (the sunlit awning ridge).
		cloth := blend(prop, pal.text, 0.20)
		fillRectC(img, cx, cy, rad, rad, cloth)
		drawHSpan(img, cx-rad, cx+rad, cy-rad, brighten(cloth, 0.16))
	}
}

// drawPond paints a BUILT decorative pond as a small water blob (playtest polish FIX 4):
// a filled blue/teal ellipse in the water tone with a subtle LIGHTER RIM one band in from
// the edge (a shallow shore / lit shallows) and a faintly darker deep center, so the pool
// reads as a made little pond from above rather than a flat dot. Kept small — a village
// ornament among the greenery, never a lake. All radii floored so it paints at village
// scale. This is decorative BUILT water; the citymap has no natural terrain water.
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
	// The water body.
	forEllipse(cx, cy, hw, hh, func(x, y int) { setPixel(img, x, y, water) })
	// A lighter shallows rim just inside the edge so the pool reads with a soft shore.
	rim := brighten(water, 0.18)
	forEllipse(cx, cy, hw, hh, func(x, y int) {
		fx := float64(x-cx) / float64(hw)
		fy := float64(y-cy) / float64(hh)
		if fx*fx+fy*fy >= 0.6 { // outer band only
			setPixel(img, x, y, rim)
		}
	})
	// A slightly deeper center pip for a hint of depth.
	setPixel(img, cx, cy, darken(water, 0.15))
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
