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

	// Roof material: base + ridge accent recipes, blended per building with a subtle
	// lineage tint so a temple reads different from a hut without leaving the era mood.
	roofBase  func(tdPal) color.RGBA // dominant roof fill (e.g. thatch brown)
	roofRidge func(tdPal) color.RGBA // ridge / texture highlight
	roofDark  func(tdPal) color.RGBA // shaded slope
	lineageMix float64               // how much lineage tint bleeds into the roof (~0.15–0.25)

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
	roofRidge: func(p tdPal) color.RGBA {
		return brighten(blend(blend(p.bg, p.text, 0.34), earthAnchor, 0.42), 0.20)
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
func roofColorsFor(style tdEraStyle, pal tdPal, domain, category string) roofColors {
	base := style.roofBase(pal)
	ridge := style.roofRidge(pal)
	dark := style.roofDark(pal)
	if style.lineageMix > 0 {
		tint := lineageColor(domain, category)
		base = blend(base, tint, style.lineageMix)
		ridge = blend(ridge, tint, style.lineageMix*0.7)
		dark = blend(dark, tint, style.lineageMix)
	}
	return roofColors{base: base, ridge: ridge, dark: dark}
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

// tdConfig holds the fixed generator constants (city-space units). City space is an
// abstract plane; the fill-frame transform later maps the plan's bounding box onto
// the canvas, so these are relative sizes, not pixels.
type tdConfig struct {
	districtRadius float64 // how far district centers sit from the core
	slotSpacing    float64 // golden-spiral spacing between instances
	roofSize       float64 // base roof extent in city units
}

var defaultTdConfig = tdConfig{
	districtRadius: 22,
	slotSpacing:    3.2,
	roofSize:       3.0,
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

	// (b) districts — five loose clusters seeded around the core at STABLE angles. A
	// stable per-seed phase rotates the whole ring so two civs differ, but a given civ
	// is fixed across ages (seed is age-independent). Clusters may overlap (blur).
	kinds := []tdDistrictKind{distResidential, distProduction, distCivic, distMarket, distGarrison}
	ringPhase := r.f01() * 2 * math.Pi
	dmap := make(map[tdDistrictKind]int, len(kinds)) // kind → index into plan.districts
	for i, k := range kinds {
		ang := ringPhase + 2*math.Pi*float64(i)/float64(len(kinds))
		// Civic sits nearest the core (its hero anchors the center); the rest ring out.
		dr := cfg.districtRadius
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
			dx, dy := spiralSlot(d.placed, cfg.slotSpacing, phase)
			d.placed++
			lot := tdLot{
				x: d.cx + dx, y: d.cy + dy,
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

// tdAddFiller scatters balanced living-city filler into the gaps: green gardens,
// paved squares near the civic core, tree dot-clusters, and small props (wells/
// stalls). Density is deliberately balanced (locked #12) — alive but never burying
// the buildings — and fully seeded so it's deterministic. Everything is placed in
// city space around the built-up footprint derived from the roof lots.
func tdAddFiller(plan *topPlan, style tdEraStyle, cfg tdConfig, r *rng) {
	// Derive the footprint radius from the roof lots so filler stays within/around the
	// built-up area (not scattered across empty space).
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
	// filler seasons the city rather than swamping it.
	gardens := int(dens * math.Sqrt(float64(roofN)) * 1.2)
	trees := int(dens * math.Sqrt(float64(roofN)) * 1.6)
	props := int(dens * math.Sqrt(float64(roofN)) * 0.7)
	squares := 1 + int(dens*math.Sqrt(float64(roofN))*0.3)

	// A paved square (or two) hugging the civic core.
	for i := 0; i < squares; i++ {
		ang := r.f01() * 2 * math.Pi
		rr := r.f01() * cfg.roofSize * 2
		plan.lots = append(plan.lots, tdLot{
			x: plan.cx + math.Cos(ang)*rr, y: plan.cy + math.Sin(ang)*rr,
			w: cfg.roofSize * 1.6, h: cfg.roofSize * 1.6, kind: tdSquare,
		})
	}
	// Gardens: green plots scattered through the built-up disk.
	for i := 0; i < gardens; i++ {
		x, y := tdDiskPoint(plan.cx, plan.cy, rad, r)
		plan.lots = append(plan.lots, tdLot{x: x, y: y, w: cfg.roofSize * 1.3, h: cfg.roofSize * 1.1, kind: tdGarden})
	}
	// Trees: small dot clusters, biased to the outer ring (a village edge of woods).
	for i := 0; i < trees; i++ {
		x, y := tdRingPoint(plan.cx, plan.cy, rad*0.55, rad*1.05, r)
		plan.lots = append(plan.lots, tdLot{x: x, y: y, w: cfg.roofSize * 0.7, h: cfg.roofSize * 0.7, kind: tdTree})
	}
	// Props: wells/stalls, sprinkled inside the built-up area.
	for i := 0; i < props; i++ {
		x, y := tdDiskPoint(plan.cx, plan.cy, rad*0.85, r)
		plan.lots = append(plan.lots, tdLot{x: x, y: y, w: cfg.roofSize * 0.5, h: cfg.roofSize * 0.5, kind: tdProp})
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

// tdDiskPoint returns a seeded point uniformly-ish within radius rad of (cx,cy).
func tdDiskPoint(cx, cy, rad float64, r *rng) (float64, float64) {
	ang := r.f01() * 2 * math.Pi
	rr := rad * math.Sqrt(r.f01())
	return cx + math.Cos(ang)*rr, cy + math.Sin(ang)*rr
}

// tdRingPoint returns a seeded point in the annulus [r0,r1] around (cx,cy).
func tdRingPoint(cx, cy, r0, r1 float64, r *rng) (float64, float64) {
	ang := r.f01() * 2 * math.Pi
	rr := r0 + (r1-r0)*r.f01()
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

// computeTransform derives the fill-frame transform. It takes the bounding box of all
// lots (roofs + filler + streets), leaves a margin, and scales so the box fills the
// canvas. Roofs shrink as the city densifies (more lots → larger box → smaller
// scale) but a min roof-size FLOOR keeps them legible. Panic-safe: a degenerate
// (empty / zero-extent) box yields a centered identity-ish transform.
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
	for _, lt := range plan.lots {
		acc(lt.x, lt.y, math.Max(lt.w, lt.h)/2)
	}
	for _, s := range plan.streets {
		for _, p := range s.pts {
			acc(p.x, p.y, 0)
		}
	}
	// Always include the core so an empty plan still centers sensibly.
	acc(plan.cx, plan.cy, 1)

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
// shape with a ridge/texture highlight. Material comes from the era style; a subtle
// lineage tint differentiates types. Dispatches on the lot's roofType archetype.
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

	// SE drop-shadow: the roof footprint, offset down-right by ~1px (scaled a touch for
	// big roofs), painted UNDER the roof. Soft = the theme shadow tone, not black.
	sh := 1 + hw/6
	drawShadow(img, cx+sh, cy+sh, hw, hh, lt.roof, pal.shadow)

	switch lt.roof {
	case roofHut:
		drawRoofHut(img, cx, cy, hw, hh, rc)
	case roofRidge:
		drawRoofRidge(img, cx, cy, hw, hh, rc, false)
	case roofLong:
		drawRoofRidge(img, cx, cy, hw, hh, rc, true)
	case roofTemple:
		drawRoofTemple(img, cx, cy, hw, hh, rc, pal)
	case roofCamp:
		drawRoofCamp(img, cx, cy, hw, hh, rc)
	case roofStash:
		drawRoofStash(img, cx, cy, hw, hh, rc)
	case roofFlat:
		drawRoofFlat(img, cx, cy, hw, hh, rc)
	case roofWonder:
		drawRoofWonder(img, cx, cy, hw, hh, rc, pal)
	default:
		drawRoofRidge(img, cx, cy, hw, hh, rc, false)
	}
}

// drawShadow paints a soft SE drop-shadow matching the roof's rough silhouette. It
// blends the shadow tone into whatever is beneath (so it darkens the ground, not
// paints a hard slab), giving a subtle hint of height.
func drawShadow(img *image.RGBA, cx, cy, hw, hh int, rt roofType, shadow color.RGBA) {
	blendFn := func(x, y int) {
		blendPixel(img, x, y, shadow, 0.35)
	}
	switch rt {
	case roofHut, roofWonder:
		forEllipse(cx, cy, hw, hh, blendFn)
	default:
		forRect(cx, cy, hw, hh, blendFn)
	}
}

// drawRoofHut: a small round/oval thatch roof with faint radial streaks from the
// apex — the primitive dwelling read from above (locked §roof atlas).
func drawRoofHut(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	forEllipse(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.base) })
	// Radial thatch streaks: a few darker spokes from center to rim.
	const spokes = 6
	for i := 0; i < spokes; i++ {
		ang := 2 * math.Pi * float64(i) / spokes
		ex := cx + int(math.Cos(ang)*float64(hw))
		ey := cy + int(math.Sin(ang)*float64(hh))
		drawLineC(img, cx, cy, ex, ey, rc.dark)
	}
	// A bright apex pip so the cone reads as a peak.
	img.SetRGBA(cx, cy, rc.ridge)
}

// drawRoofRidge: a rectangular pitched roof — a center ridge line with two shaded
// slopes falling to the eaves. long=true elongates it (longhouse / rowhouse).
func drawRoofRidge(img *image.RGBA, cx, cy, hw, hh int, rc roofColors, long bool) {
	// Fill the rectangle with the two slopes: rows above the ridge lean light, below
	// lean dark, so the pitch reads. The ridge runs along the long axis.
	horizontalRidge := hw >= hh // ridge along the wider axis
	forRect(cx, cy, hw, hh, func(x, y int) {
		var slope color.RGBA
		if horizontalRidge {
			if y < cy {
				slope = brighten(rc.base, 0.05)
			} else {
				slope = rc.dark
			}
		} else {
			if x < cx {
				slope = brighten(rc.base, 0.05)
			} else {
				slope = rc.dark
			}
		}
		img.SetRGBA(x, y, slope)
	})
	// The ridge highlight down the center of the long axis.
	if horizontalRidge {
		drawHSpan(img, cx-hw, cx+hw, cy, rc.ridge)
	} else {
		for y := cy - hh; y <= cy+hh; y++ {
			img.SetRGBA(cx, y, rc.ridge)
		}
	}
}

// drawRoofTemple: a larger ornate symmetric roof (a stepped/tiered look) with a
// bright finial at the apex — the shrine/temple hero read (faith/knowledge/culture).
func drawRoofTemple(img *image.RGBA, cx, cy, hw, hh int, rc roofColors, pal tdPal) {
	// Base tier: full footprint.
	forRect(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.base) })
	// Inner tier: a brighter concentric rectangle for the ornate stepped look.
	ihw := maxInt(hw*2/3, 1)
	ihh := maxInt(hh*2/3, 1)
	forRect(cx, cy, ihw, ihh, func(x, y int) { img.SetRGBA(x, y, brighten(rc.base, 0.12)) })
	// A cross ridge (both axes) so it reads symmetric + civic.
	drawHSpan(img, cx-hw, cx+hw, cy, rc.ridge)
	for y := cy - hh; y <= cy+hh; y++ {
		img.SetRGBA(cx, y, rc.ridge)
	}
	// Finial: a bright accent pip at the apex (theme accent so it pops as sacred).
	img.SetRGBA(cx, cy, brighten(pal.accent, 0.10))
}

// drawRoofCamp: an open-frame / lean-to / tent — the gathering/forager/war camp. A
// triangular tent silhouette (a ridge with one open side) rather than a full roof, so
// it reads as a temporary structure among the solid dwellings.
func drawRoofCamp(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// A tent: a filled triangle peaking at the top center, widening to the base — drawn
	// as rows whose width grows toward the bottom. Open frame = leave the base row lit.
	for dy := -hh; dy <= hh; dy++ {
		// Fraction from apex(0) to base(1).
		f := float64(dy+hh) / float64(2*hh+1)
		rowHW := int(float64(hw) * f)
		col := rc.base
		if dy < 0 {
			col = brighten(rc.base, 0.05)
		}
		drawHSpan(img, cx-rowHW, cx+rowHW, cy+dy, col)
	}
	// The two tent poles / ridge as darker edges, and a bright apex.
	drawLineC(img, cx, cy-hh, cx-hw, cy+hh, rc.dark)
	drawLineC(img, cx, cy-hh, cx+hw, cy+hh, rc.dark)
	img.SetRGBA(cx, cy-hh, rc.ridge)
}

// drawRoofStash: a small square store hut — a compact filled square with a plank
// highlight, quieter than a dwelling (storage lineage).
func drawRoofStash(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Keep it small + squat: clamp to a tight square.
	s := minInt(hw, hh)
	if s < 1 {
		s = 1
	}
	forRect(cx, cy, s, s, func(x, y int) { img.SetRGBA(x, y, rc.dark) })
	// A single plank highlight across the top.
	drawHSpan(img, cx-s, cx+s, cy-s, rc.ridge)
}

// drawRoofFlat: a low flat structure — a flat slab with a thin rim, for the stone
// works. Reads distinctly from a pitched dwelling (no ridge).
func drawRoofFlat(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	forRect(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.base) })
	// A darker rim around the slab so it reads as a low walled platform.
	forRectOutline(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.dark) })
}

// drawRoofWonder: the DOMINANT central complex — a large ornate multi-part roof
// (locked #13). A big footprint with concentric ornate tiers, cross ridges, and an
// accent finial: unmistakably the grandest thing on the map.
func drawRoofWonder(img *image.RGBA, cx, cy, hw, hh int, rc roofColors, pal tdPal) {
	// Outer ornate ring (ellipse) for a grand base.
	forEllipse(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.base) })
	// A square inner hall.
	ihw := maxInt(hw*2/3, 1)
	ihh := maxInt(hh*2/3, 1)
	forRect(cx, cy, ihw, ihh, func(x, y int) { img.SetRGBA(x, y, brighten(rc.base, 0.10)) })
	// A smaller top tier.
	thw := maxInt(hw/3, 1)
	thh := maxInt(hh/3, 1)
	forRect(cx, cy, thw, thh, func(x, y int) { img.SetRGBA(x, y, brighten(pal.accent, 0.05)) })
	// Cross ridges + finial.
	drawHSpan(img, cx-hw, cx+hw, cy, rc.ridge)
	for y := cy - hh; y <= cy+hh; y++ {
		img.SetRGBA(cx, y, rc.ridge)
	}
	img.SetRGBA(cx, cy, brighten(pal.accent, 0.25))
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
