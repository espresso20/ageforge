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
// persistent placement (golden-angle spirals per district), fill-frame scaling,
// the top-down roof atlas with SE drop-shadows, loose districts, organic streets,
// balanced living-city filler, the walls capability (off in primitive), the wonder
// centerpiece, and landmark-only labels.
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

// tdDistrictKind is a loose district cluster (locked #11): buildings gravitate to
// same-kind clusters that BLUR into each other, not hard zones.
type tdDistrictKind int

const (
	distResidential tdDistrictKind = iota
	distProduction
	distCivic
	distMarket
	distGarrison
)

// tdDistrict is one loose cluster: a center in city space and the running count of
// instances placed into it (drives the golden-angle spiral radius). Center angles
// are stable per seed so the district map doesn't rearrange across renders/ages.
type tdDistrict struct {
	kind    tdDistrictKind
	cx, cy  float64
	placed  int // how many roof lots have been slotted here (spiral index cursor)
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
	streets   []tdStreet
	districts []tdDistrict
	lots      []tdLot
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

// ---- district classification ------------------------------------------------

// districtKindFor sorts a building into a loose district by its category/domain
// (locked #11): housing→residential, faith/knowledge/culture + research/wonder→civic,
// military→garrison, storage/trade/harbor→market, everything else→production.
func districtKindFor(domain, category string) tdDistrictKind {
	switch category {
	case "housing":
		return distResidential
	case "research", "wonder", "monument", "diplomacy":
		return distCivic
	case "military":
		return distGarrison
	case "storage":
		return distMarket
	}
	switch domain {
	case "faith", "knowledge", "culture_arts":
		return distCivic
	case "military":
		return distGarrison
	case "storage", "trade", "harbor":
		return distMarket
	}
	return distProduction
}

// ---- generate (pure, deterministic, stable-incremental) ---------------------

// goldenAngle is the golden angle in radians (~137.5°). Placing instance i at angle
// i*goldenAngle with radius ∝ sqrt(i) fills a disk with a stable, low-overlap
// phyllotaxis spiral: every instance has a FIXED slot index, so adding the (N+1)th
// building never moves the first N (locked #8, the anti-re-randomize guarantee).
const goldenAngle = 2.399963229728653 // math.Pi * (3 - sqrt(5))

// spiralSlot returns the city-space offset for the i-th instance in a district's
// golden-angle spiral: angle = i*goldenAngle (+ a stable per-district phase), radius
// = spacing*sqrt(i). Deterministic and index-stable — the whole point of the spiral.
func spiralSlot(i int, spacing, phase float64) (dx, dy float64) {
	r := spacing * math.Sqrt(float64(i))
	a := float64(i)*goldenAngle + phase
	return math.Cos(a) * r, math.Sin(a) * r
}

// slotJitter returns a small ORGANIC offset for the i-th slot of district di, breaking
// up the visible golden-angle diamond-lattice so the hamlet reads natural rather than
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
	districtRadius float64 // how far district centers sit from the core (large city)
	slotSpacing    float64 // golden-spiral spacing between instances (tight hamlet)
	roofSize       float64 // base roof extent in city units
	jitterAmp      float64 // organic per-slot wander (city units) breaking the lattice

	// clusterThreshold: below this many total roof lots the settlement stays a SINGLE
	// cohesive cluster (districts pull in to a tight core radius); at/above it the loose
	// multi-district ring emerges (locked #11, but a village is one hamlet — locked spec
	// + playtest). tightRadiusScale is how far the sub-clusters sit from the core while
	// the settlement is small — small enough that they blur into one town, not spokes.
	clusterThreshold int
	tightRadiusScale float64
}

var defaultTdConfig = tdConfig{
	districtRadius:   22,
	slotSpacing:      2.4,
	roofSize:         3.2,
	jitterAmp:        0.8,
	clusterThreshold: 35,
	tightRadiusScale: 0.20,
}

// generateTopPlan synthesizes the whole top-down city plan in CITY SPACE, purely and
// deterministically from seed. Pipeline (city-synthesis.md §Pipeline, top-down):
//
//	(a) gather   — built buildings → per-type domain/category/tier/count/role.
//	(b) districts — seed loose cluster centers at stable angles around the core.
//	(c) populate — each building emits count-scaled roof lots into its district via
//	    the STABLE golden-angle spiral, in a fixed order; wonders → central complex;
//	    the civic hero (or a promoted production hero) is labeled.
//	(d) streets  — winding dirt lanes linking each district center to the core.
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

	// Total roof lots (excluding the centerpiece wonder) drives the settlement mode: a
	// small settlement is ONE tight hamlet, a large one spreads into the loose district
	// ring. Computed up front from the same tdRoofCount the populate loop uses, so the
	// threshold decision is exact.
	totalRoofs := 0
	for _, b := range blds {
		if b.category == "wonder" || b.category == "monument" {
			continue
		}
		totalRoofs += tdRoofCount(b.count, b.role)
	}
	// Below the threshold: pull the district centers WAY in so the whole thing reads as
	// one cohesive cluster around the core (a hamlet), not spokes to distant clumps
	// (FIX 2 / playtest). At/above it: the full loose multi-district ring (locked #11).
	radiusScale := 1.0
	if totalRoofs < cfg.clusterThreshold {
		radiusScale = cfg.tightRadiusScale
	}

	// (b) districts — five loose clusters seeded around the core at STABLE angles. A
	// stable per-seed phase rotates the whole ring so two civs differ, but a given civ
	// is fixed across ages (seed is age-independent). Clusters may overlap (blur). While
	// the settlement is small, radiusScale collapses the ring toward the core.
	kinds := []tdDistrictKind{distResidential, distProduction, distCivic, distMarket, distGarrison}
	ringPhase := r.f01() * 2 * math.Pi
	dmap := make(map[tdDistrictKind]int, len(kinds)) // kind → index into plan.districts
	for i, k := range kinds {
		ang := ringPhase + 2*math.Pi*float64(i)/float64(len(kinds))
		// Civic sits nearest the core (its hero anchors the center); the rest ring out.
		dr := cfg.districtRadius * radiusScale
		if k == distCivic {
			dr *= 0.35
		}
		// A small stable jitter so the ring isn't a perfect pentagon.
		jr := dr * (0.85 + 0.30*r.f01())
		plan.districts = append(plan.districts, tdDistrict{
			kind: k,
			cx:   math.Cos(ang) * jr,
			cy:   math.Sin(ang) * jr,
		})
		dmap[k] = i
	}

	// Split gathered buildings by role for hero promotion + wonder handling.
	var landmarks, production []builtBuilding
	var wonder *builtBuilding
	for i := range blds {
		b := blds[i]
		if b.category == "wonder" || b.category == "monument" {
			// Keep the highest-tier wonder as THE centerpiece (locked #13).
			if wonder == nil || b.tier > wonder.tier || (b.tier == wonder.tier && b.count > wonder.count) {
				w := b
				wonder = &w
			}
			continue
		}
		if b.role == roleLandmark {
			landmarks = append(landmarks, b)
		} else if b.role == roleProduction {
			production = append(production, b)
		}
	}

	// The single most prominent LANDMARK is the labeled civic hero at the center. If
	// the civ has no civic building at all (a village of only huts + camps), promote
	// its most prominent PRODUCTION building to the hero so the city still has exactly
	// one labeled landmark (locked #7). This mirrors the old promote-a-hero rule.
	heroKey := ""
	if len(landmarks) > 0 {
		best := 0
		for i := 1; i < len(landmarks); i++ {
			if moreProminentBld(landmarks[i], landmarks[best]) {
				best = i
			}
		}
		heroKey = landmarks[best].key
	} else if wonder == nil && len(production) > 0 {
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

	// (c) populate — each building emits count-scaled roof lots into its district via
	// the stable golden-angle spiral. The spiral index is per-district and advances in
	// a fixed order (sorted gather, then per-building loop), so slot N is stable.
	for _, b := range blds {
		if (b.category == "wonder" || b.category == "monument") && wonder != nil && b.key == wonder.key {
			continue // the centerpiece is placed separately, dead-center
		}
		n := tdRoofCount(b.count, b.role)
		if n <= 0 {
			continue
		}
		dk := districtKindFor(b.domain, b.category)
		di := dmap[dk]
		d := &plan.districts[di]
		rt := getRoofType(b.domain, b.category, b.tier)
		// A stable per-district phase so different districts' spirals don't align.
		phase := float64(di) * 1.7
		sz := cfg.roofSize
		if rt == roofLong {
			sz *= 1.15
		}
		for j := 0; j < n; j++ {
			slot := d.placed
			dx, dy := spiralSlot(slot, cfg.slotSpacing, phase)
			// Organic wander: a pure function of (slot, di, seed), so it never moves an
			// existing building when a later one is added (stable-incremental, locked #8).
			jx, jy := slotJitter(slot, di, seed, cfg.jitterAmp)
			d.placed++
			lot := tdLot{
				x: d.cx + dx + jx, y: d.cy + dy + jy,
				w: sz, h: sz, kind: tdRoof,
				domain: b.domain, category: b.category, tier: b.tier, roof: rt,
			}
			// Longhouses/rowhouses are elongated.
			if rt == roofLong {
				lot.w = sz * 1.8
			}
			// Label the civic hero on its first (innermost) instance only.
			if b.key == heroKey && j == 0 && !plan.hasHero {
				lot.label = b.name
				lot.prom = prominenceOf(b)
			}
			plan.lots = append(plan.lots, lot)
		}
	}

	// Wonder centerpiece: a dominant, ornate complex dead-center (locked #13). Placed
	// last in the roof list order-wise but drawn back-to-front by y, so it crowns.
	if wonder != nil {
		plan.lots = append(plan.lots, tdLot{
			x: plan.cx, y: plan.cy, w: cfg.roofSize * 3.0, h: cfg.roofSize * 3.0,
			kind: tdRoof, domain: wonder.domain, category: wonder.category, tier: wonder.tier,
			roof: roofWonder, label: wonder.name, prom: 1000,
		})
	}

	// (d) streets — winding dirt lanes from each district center to the core, plus a
	// couple of meandering cross paths, all in city space (no terrain routing).
	plan.streets = tdOrganicStreets(plan, style, r)

	// (e) filler — balanced gardens / squares / trees / props in the gaps.
	tdAddFiller(&plan, style, cfg, r)

	// (f) walls — a wall+gate ring IF the era has walls. Primitive: none. The ring is
	// sized to the lot bounding box so it hugs the built-up area.
	if style.hasWalls {
		tdAddWalls(&plan, r)
	}

	return plan
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

// tdOrganicStreets grows the village lane network in city space: a winding dirt lane
// from each district center back to the core, plus a couple of meandering cross
// paths between adjacent districts. Purely jittered polylines — NO terrain routing,
// no water (locked #10, V3-A organic). The generator is shaped so grid/avenue
// patterns can slot in later; only organic is tuned now.
func tdOrganicStreets(plan topPlan, style tdEraStyle, r *rng) []tdStreet {
	streets := make([]tdStreet, 0, len(plan.districts)+2)
	core := tdPoint{plan.cx, plan.cy}
	for _, d := range plan.districts {
		streets = append(streets, windingLane(tdPoint{d.cx, d.cy}, core, style.streetJitter, style.laneWidth, r))
	}
	// A couple of cross paths linking consecutive district centers so the village reads
	// as connected lanes, not just spokes. Bounded so it stays a village, not a grid.
	n := len(plan.districts)
	for i := 0; i < n && i < 3; i++ {
		a := plan.districts[i]
		b := plan.districts[(i+1)%n]
		streets = append(streets, windingLane(tdPoint{a.cx, a.cy}, tdPoint{b.cx, b.cy}, style.streetJitter, style.laneWidth, r))
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
