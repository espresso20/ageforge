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

	// houseProfile is the per-era ROOF SHAPE dialect (V3-B). The roof ATLAS (getRoofType) still
	// picks the archetype (hut / ridge / long / …) from the building's domain+tier — that is
	// era-independent — but the SILHOUETTE of a dwelling reads differently per era: a primitive
	// thatch hut is rounded and domed, an ancient mudbrick house is FLATTER and BLOCKIER, a
	// medieval house is STEEPER-pitched timber. drawRoof consults the profile to nudge the roof
	// proportions/shading so the same archetype reads era-appropriate. profileThatch is the
	// unchanged primitive/default look. Zero value == profileThatch (safe default).
	houseProfile roofProfile

	// wallProfile is the per-era WALL DIALECT (V3-B, locked #9). Ancient = a plain MUDBRICK
	// curtain (tan, thin, no towers); medieval = a STONE curtain (grey, thicker) studded with
	// periodic TOWERS + a gatehouse. tdAddWalls reads it to decide the wall thickness, whether to
	// emit towers, and the gate structure. Only meaningful when hasWalls is true.
	wallProfile wallProfile

	// wonderMotif is the per-era CENTERPIECE silhouette drawn for a city's dominant wonder,
	// decoupled from houseProfile (Phase 1b-i) so an age can pair (e.g.) thatch houses with a
	// megalith monument, or white-stone houses with a temple. drawRoof consults this — not the
	// house profile — to pick the wonder sprite. Zero value == wonderGeneric (the grand default
	// hall), so an untuned era keeps its current wonder.
	wonderMotif wonderMotif

	// spaceMode flips the whole GROUND read from a town-on-terrain to a station-in-the-void
	// (Phase 2c). When true, drawGround paints a deep-space VOID + STARFIELD instead of the era
	// ground tint (drawSpaceBackground), and tdAddFiller SUPPRESSES all greenery — gardens, trees,
	// groves, and ponds — because a station floating in space has no soil. Only the five
	// SPACE-AND-ABOVE ages set it (space / interstellar / galactic / quantum / transcendent); every
	// grounded era keeps the zero value (false) and its dirt/grass/deck ground untouched.
	spaceMode bool
}

// roofProfile is the per-era dwelling-roof dialect (V3-B). It shifts the ROOF SILHOUETTE of the
// house/hut archetypes without changing which archetype getRoofType assigns.
type roofProfile int

const (
	profileThatch         roofProfile = iota // primitive/default: rounded domed thatch, standard pitch
	profileMudbrick                          // ancient: flatter, blockier flat-topped mud roofs
	profileTimber                            // medieval: steeper, sharper pitched timber roofs
	profileStoneClassical                    // classical: pale white-stone body under a terracotta cap, with column fluting
	profileRowhouse                          // colonial/industrial: a TERRACE of 3–5 narrow attached units under small pitched roofs
	profileModernFlat                        // electric/atomic: a FLAT-topped modern block (flat roof slab + thin parapet rim + a rooftop vent) — the groundwork for skyscrapers
	profileGlassTower                        // modern/information/digital: a TALL glass-and-steel tower from above (cool blue-grey slab + lit window grid + a long height shadow)
	profileMetalDome                         // space: a small pale METALLIC DOME (a lit silver disc with a curved NW highlight + a rim) — the space-colony habitat dwelling
	profileSpire                             // interstellar: a TALL NARROW tapering metallic SPIRE read from above (a small bright metal core + a long SE height shadow + a lit tip) — the deep-space arcology dwelling
	profileLattice                           // quantum: a small faceted CRYSTAL / lattice NODE — a few triangular facets in SHIFTING iridescent hues (cyan/magenta/gold) around a bright core, so the dwelling reads as a glinting gem
	profileEthereal                          // transcendent: a soft glowing translucent BLOOM instead of a hard roof — a luminous white light-form (a dematerialised dwelling), the ethereal finale
)

// wallProfile is the per-era wall dialect (V3-B, locked #9): mudbrick curtain vs stone curtain +
// towers + gatehouse. Consumed by tdAddWalls.
type wallProfile int

const (
	wallNone     wallProfile = iota // no wall (primitive, industrial+)
	wallMudbrick                    // ancient/bronze: thin tan curtain, gate gaps, no towers
	wallStone                       // medieval/classical: thicker grey curtain, towers, a gatehouse
	wallTimber                      // iron: medium brown timber palisade, gate gaps, no stone towers
	wallStarFort                    // renaissance: thick earthwork curtain with ANGULAR triangular BASTION salients (no round towers)
)

// wonderMotif is the centerpiece silhouette drawn for a city's dominant wonder, decoupled from
// houseProfile (Phase 1b-i) so an age can pair (e.g.) thatch houses with a megalith monument, or
// white-stone houses with a temple. drawRoof switches on this to pick the wonder sprite.
type wonderMotif int

const (
	wonderGeneric        wonderMotif = iota // grand generic hall (drawRoofWonder)
	wonderZiggurat                          // ancient stepped pyramid (drawRoofZiggurat)
	wonderCathedral                         // medieval cruciform + spire (drawRoofCathedral)
	wonderMegalith                          // stone-age standing-stone circle (drawRoofMegalith)
	wonderTemple                            // classical columned temple + pediment (drawRoofTempleWonder)
	wonderKeep                              // iron-age fortified keep + watchtower (drawRoofKeep)
	wonderDome                              // renaissance great domed rotunda + lantern (drawRoofDome)
	wonderFactory                           // industrial great factory hall + smokestacks + soot (drawRoofFactory)
	wonderTower                             // electric art-deco SETBACK tower — concentric flat rectilinear tiers to a central mast (drawRoofTower)
	wonderSpaceNeedle                       // atomic SPACE-AGE needle — a narrow stem with a wide round SAUCER disc (a lit ring + a bright core) high up + a small mast + a long height shadow (drawRoofSpaceNeedle) — the googie centrepiece that splits atomic from electric's deco tower
	wonderSkyscraper                        // modern/digital SUPERTALL glass tower — a narrow glass slab + a lit window grid + a mast/antenna + a long height shadow (drawRoofSkyscraper)
	wonderDataHub                           // information DATA megastructure — a wide low server-farm BASE (cool grey blocks) with a central tall COMMS antenna/lattice mast + a few bright beacon/LED dabs (cyan/amber) (drawRoofDataHub) — the server-city centrepiece that splits information from modern's glass tower
	wonderFusionCore                        // fusion glowing REACTOR — concentric bright cyan rings around a white-hot central core, brightening to a bloom at the very center (drawRoofFusionCore)
	wonderLaunchpad                         // space ROCKET on a launch pad — a circular pad + a bright central rocket (capsule + fins) + gantry dabs + a scorch ring (drawRoofLaunchpad)
	wonderSpireArray                        // interstellar SPIRE CLUSTER — a ring of tall metallic spires around a tallest central spire, each throwing a long SE height shadow (drawRoofSpireArray)
	wonderRingHub                           // galactic RING-HUB megastation — concentric bright metallic orbital RINGS around a glowing central HUB, with faint spokes (drawRoofRingHub)
	wonderCrystalLattice                    // quantum CRYSTAL LATTICE — a geometric MESH of glowing nodes joined by thin iridescent lines, hues SHIFTING across the grid, a bright central node (drawRoofCrystalLattice)
	wonderAscension                         // transcendent ASCENSION OF LIGHT — a bright vertical light PILLAR at center ringed by concentric soft glowing HALOS brightening to a pure-white core, an ethereal gate (drawRoofAscension) — the luminous finale of the whole progression
)

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

// ancientCityStyle is the tuned ANCIENT preset (bronze / iron / classical — eraHubSpoke; locked
// era table row "ancient"): CLAY-TILE roofs (warm terracotta/tan), a PACKED-EARTH / PALE-STONE
// ground, MUDBRICK houses (flatter/blockier), and — the V3-B centrepiece — a MUDBRICK wall+gate
// ring around the built-up area. It starts from the default (so it keeps the tuned ground
// texture / pond / filler behaviour) and overrides only the era MOOD recipes; every tone stays a
// theme-role recipe so the whole city retints on a theme switch.
var ancientCityStyle = func() tdEraStyle {
	s := defaultTdStyle
	s.name = "ancient"

	// Clay tile: a warm terracotta fill. Background lifted toward text (a warm neutral in these
	// themes), then pulled firmly toward the clay anchor so a roof reads as fired tile, not thatch.
	s.roofBase = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.30), clayAnchor, 0.52)
	}
	s.roofDark = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.text, 0.30), clayAnchor, 0.52), 0.28)
	}
	s.lineageMix = 0.16 // keep the subtle lineage tint; the sat cap still guards the no-accent rule

	// Ground: packed earth / pale stone — drier and paler than the primitive dirt+grass. Base
	// pulled toward the mudbrick tan; alt a touch toward pale stone for a quiet dusty variation.
	s.groundBase = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.26), mudbrickAnchor, 0.34)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.22), stoneAnchor, 0.24)
	}
	// Streets: worn pale-stone/earth lanes — the packed-earth surface leaning a shade greyer than
	// the primitive tan so the gaps read against the drier ancient ground.
	s.streetCol = func(p tdPal) color.RGBA {
		packed := blend(blend(p.bg, p.dim, 0.26), mudbrickAnchor, 0.42)
		return blend(blend(packed, p.text, 0.36), stoneAnchor, 0.30)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		packed := blend(blend(p.bg, p.dim, 0.26), mudbrickAnchor, 0.42)
		surface := blend(blend(packed, p.text, 0.36), stoneAnchor, 0.30)
		return darken(surface, 0.20)
	}
	// Town-square paving: pale dressed stone, lighter/greyer than the ancient ground.
	s.pavedCol = func(p tdPal) color.RGBA {
		earthy := blend(blend(p.bg, p.dim, 0.30), mudbrickAnchor, 0.30)
		return blend(blend(earthy, p.text, 0.34), stoneAnchor, 0.40)
	}

	// Walls ON (locked #9): a MUDBRICK curtain — the mudbrick tan grounded a touch so it reads as
	// a sun-baked earthen rampart, not a roof.
	s.hasWalls = true
	s.wallProfile = wallMudbrick
	s.wallCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.30), mudbrickAnchor, 0.50)
	}

	s.houseProfile = profileMudbrick
	s.wonderMotif = wonderZiggurat // ancient centrepiece: a stepped ziggurat
	s.slotSpacing = 1.9            // a shade tighter than primitive (2.4), looser than the medieval city
	return s
}()

// ironCityStyle is the tuned IRON preset (Phase 1b-ii, +follow-up). It is an ANCIENT city that has
// learned to work metal: same mudbrick houses + packed-earth ground as bronze — but grimmer, and
// with its own monument. The roofs read COOLER/GREYER (iron accents in the fired clay), the ground a
// shade more WORKED (cooler, greyer than bronze's warm tan), the wall is a brown TIMBER PALISADE
// (wallTimber) instead of bronze's tan mudbrick curtain, and the centerpiece is a fortified KEEP +
// watchtower (wonderKeep) instead of bronze's ziggurat — so iron reads clearly apart from bronze.
// Built from ancientCityStyle for the hub-spoke form; every tone stays a theme-role recipe so the
// whole city retints.
var ironCityStyle = func() tdEraStyle {
	s := ancientCityStyle
	s.name = "iron"

	// Clay tile with IRON ACCENTS: start from the same warm clay fill as bronze, then pull it toward
	// the cool iron-grey so the roof reads as fired tile darkened by metalwork — clearly cooler +
	// grimmer than bronze's warm terracotta. roofDark leans harder into iron-grey for a cold shade.
	s.roofBase = func(p tdPal) color.RGBA {
		clay := blend(blend(p.bg, p.text, 0.30), clayAnchor, 0.48)
		return blend(clay, ironAnchor, 0.32)
	}
	s.roofDark = func(p tdPal) color.RGBA {
		clay := blend(blend(p.bg, p.text, 0.30), clayAnchor, 0.48)
		cool := blend(clay, ironAnchor, 0.50)
		return darken(cool, 0.26)
	}
	s.lineageMix = 0.14

	// Ground: packed earth, but cooler + greyer than bronze's warm tan — a worked iron-working town.
	// Same mudbrick base, then pulled toward iron-grey so the floor reads dustier/greyer.
	s.groundBase = func(p tdPal) color.RGBA {
		earthy := blend(blend(p.bg, p.dim, 0.26), mudbrickAnchor, 0.30)
		return blend(earthy, ironAnchor, 0.16)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.22), ironAnchor, 0.22)
	}

	// Wall: a brown TIMBER PALISADE (wallTimber) — the timber-brown anchor grounded so it reads as a
	// log stockade, not a tan mud rampart. Medium thickness (set in tdAddWalls), no stone towers.
	s.hasWalls = true
	s.wallProfile = wallTimber
	s.wallCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.28), timberAnchor, 0.52)
	}

	s.houseProfile = profileMudbrick // unchanged from ancient — the wall + cooler roofs carry iron
	s.wonderMotif = wonderKeep       // iron gets its OWN fortified keep + watchtower — distinct from bronze's ziggurat
	s.slotSpacing = 1.9
	return s
}()

// classicalCityStyle is the tuned CLASSICAL preset (Phase 1b-ii) — a Greco-Roman city: WHITE-STONE
// houses under TERRACOTTA roofs with column fluting (profileStoneClassical), a pale MARBLE-PAVED
// ground (lighter + cleaner than packed earth), proper grey STONE walls + towers + gatehouse
// (wallStone), and a columned TEMPLE-with-pediment wonder (wonderTemple, the Parthenon read). Built
// from ancientCityStyle for the hub-spoke ancient form, then re-skinned light + civic. Reads clearly
// apart from medieval: classical is LIGHT + WARM (white body, terracotta, columns) where medieval is
// COOL GREY (slate, cobble, gabled timber).
var classicalCityStyle = func() tdEraStyle {
	s := ancientCityStyle
	s.name = "classical"

	// Roof material: the classical HOUSE draws a two-tone (white-stone body + terracotta cap) inside
	// drawRoofStoneClassical, so roofBase here is the pale WHITE-STONE body tone (the terracotta cap
	// is derived from clayAnchor inside the sprite). Non-house roofs (temple/flat/etc.) then read as
	// pale civic stone, which suits the classical mood.
	s.roofBase = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.34), marbleAnchor, 0.52)
	}
	s.roofDark = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.text, 0.34), marbleAnchor, 0.52), 0.24)
	}
	s.lineageMix = 0.12 // whisper of lineage tint over the pale stone; sat cap guards it

	// Ground: pale marble-paved civic stone — lighter + cleaner than the ancient packed earth. Blend
	// the marble anchor strongly and lift toward the light neutral so a classical city reads as
	// dressed stone underfoot, not dirt.
	s.groundBase = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.20), marbleAnchor, 0.44)
		return blend(pale, p.text, 0.16)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.18), stoneAnchor, 0.34)
	}
	// Streets: pale flagstone lanes, cleaner + lighter than the ancient earth streets.
	s.streetCol = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.20), marbleAnchor, 0.40)
		return blend(blend(paved, p.text, 0.30), stoneAnchor, 0.24)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.20), marbleAnchor, 0.40)
		surface := blend(blend(paved, p.text, 0.30), stoneAnchor, 0.24)
		return darken(surface, 0.20)
	}
	// Town-square paving: bright dressed marble, the lightest surface in the city.
	s.pavedCol = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.16), marbleAnchor, 0.46)
		return blend(pale, p.text, 0.28)
	}

	// Wall: proper grey STONE curtain + towers + gatehouse (wallStone) — the masonry-grey granite
	// anchor grounded so it reads as cut stone.
	s.hasWalls = true
	s.wallProfile = wallStone
	s.wallCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.30), graniteAnchor, 0.50)
	}

	s.houseProfile = profileStoneClassical // white-stone body + terracotta cap + column fluting
	s.wonderMotif = wonderTemple           // classical centrepiece: a columned temple + pediment
	s.slotSpacing = 1.85
	return s
}()

// medievalCityStyle is the tuned MEDIEVAL preset (medieval / renaissance — eraCastle; locked era
// table row "castle"): SLATE/TILE roofs (grey / dark blue-grey), a COBBLE / STONE-GREY ground,
// TIMBER houses (steeper pitched), and a STONE wall+towers+gatehouse ring. Same construction as
// the ancient preset (copy the default, override the mood recipes, keep theme-derived tones).
var medievalCityStyle = func() tdEraStyle {
	s := defaultTdStyle
	s.name = "medieval"

	// Slate: a cool dark blue-grey. Background grounded toward dim (so it reads darker/cooler than
	// the warm eras), then pulled to the slate anchor.
	s.roofBase = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.22), slateAnchor, 0.56)
	}
	s.roofDark = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.dim, 0.22), slateAnchor, 0.56), 0.30)
	}
	s.lineageMix = 0.14 // a whisper of lineage tint over the cool slate; sat cap still guards it

	// Ground: cobble / stone grey — cool and neutral, the era mood stepping off the warm earth.
	s.groundBase = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.30), cobbleAnchor, 0.34)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.24), graniteAnchor, 0.22)
	}
	// Streets: paved cobble lanes — the cobble ground lifted toward the light neutral + a touch of
	// masonry grey so the gaps read as trodden paving between the wards.
	s.streetCol = func(p tdPal) color.RGBA {
		packed := blend(blend(p.bg, p.dim, 0.30), cobbleAnchor, 0.42)
		return blend(blend(packed, p.text, 0.34), graniteAnchor, 0.30)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		packed := blend(blend(p.bg, p.dim, 0.30), cobbleAnchor, 0.42)
		surface := blend(blend(packed, p.text, 0.34), graniteAnchor, 0.30)
		return darken(surface, 0.22)
	}
	// Town-square paving: dressed flagstone, a lighter cool grey than the cobble ground.
	s.pavedCol = func(p tdPal) color.RGBA {
		stony := blend(blend(p.bg, p.dim, 0.30), cobbleAnchor, 0.34)
		return blend(blend(stony, p.text, 0.36), graniteAnchor, 0.42)
	}

	// Walls ON (locked #9): a STONE curtain (+ towers + a gatehouse, emitted by tdAddWalls) — the
	// masonry-grey granite anchor grounded so it reads as cut stone.
	s.hasWalls = true
	s.wallProfile = wallStone
	s.wallCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.34), graniteAnchor, 0.52)
	}

	s.houseProfile = profileTimber
	s.wonderMotif = wonderCathedral // medieval centrepiece: a cruciform cathedral + spire
	s.slotSpacing = 1.7             // the tuned tighter default — a walled medieval town packs denser
	return s
}()

// renaissanceCityStyle is the tuned RENAISSANCE preset (Phase 1b — renaissance split off medieval,
// which it used to share). A grand ORNATE CREAM-STONE city: pale ivory ashlar houses with lead-grey
// accents (profileStoneClassical), pale DRESSED-STONE civic paving underfoot, a pale-stone EARTHWORK
// STAR-FORT wall with ANGULAR triangular BASTIONS (wallStarFort — no round towers), and a great
// domed rotunda centerpiece (wonderDome, the Florence/St-Peter's read). Built from classicalCityStyle
// (which already gives the pale-dressed-stone ground + stone walls + white-stone house profile), then
// re-skinned WARMER + LIGHTER + more monumental. Reads clearly apart from BOTH neighbours: medieval is
// COOL GREY (slate roofs, cobble, cathedral); classical is PLAIN WHITE (marble, terracotta caps,
// temple); renaissance is WARM CREAM/IVORY ashlar + a DOME + a STAR-FORT.
var renaissanceCityStyle = func() tdEraStyle {
	s := classicalCityStyle
	s.name = "renaissance"

	// Roof material: pale CREAM/IVORY ashlar — warmer + a touch lighter than classical's cooler marble
	// white, so a renaissance city reads as dressed golden-cream stone, not grey slate and not plain
	// marble. The stone-classical house sprite still draws its two-tone body; roofBase is the cream body.
	s.roofBase = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.36), creamStoneAnchor, 0.56)
	}
	s.roofDark = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.text, 0.36), creamStoneAnchor, 0.56), 0.22)
	}
	s.lineageMix = 0.12 // a whisper of lineage tint over the cream stone; sat cap guards it

	// Ground: pale DRESSED-STONE civic paving — cleaner + creamier than classical's cool marble grey,
	// a monumental piazza floor. Blend the cream anchor and lift toward the light neutral.
	s.groundBase = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.18), creamStoneAnchor, 0.46)
		return blend(pale, p.text, 0.18)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.16), stoneAnchor, 0.32)
	}
	// Streets: pale flagstone avenues, warm cream and clean.
	s.streetCol = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.18), creamStoneAnchor, 0.42)
		return blend(blend(paved, p.text, 0.30), stoneAnchor, 0.20)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.18), creamStoneAnchor, 0.42)
		surface := blend(blend(paved, p.text, 0.30), stoneAnchor, 0.20)
		return darken(surface, 0.20)
	}
	// Town-square paving: the brightest dressed cream stone in the city — a grand civic piazza.
	s.pavedCol = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.14), creamStoneAnchor, 0.50)
		return blend(pale, p.text, 0.30)
	}

	// Wall: a pale-stone EARTHWORK STAR-FORT (wallStarFort) — a thick low rampart with angular
	// triangular bastions. A dressed-cream face grounded DARKER + EARTHIER (toward dim + a touch of
	// earth) so the rampart reads as a RAISED earthwork with strong contrast against the bright cream
	// piazza floor — not the cool grey masonry of the medieval/classical curtain, and not so pale it
	// vanishes into the paving.
	s.hasWalls = true
	s.wallProfile = wallStarFort
	s.wallCol = func(p tdPal) color.RGBA {
		face := blend(blend(p.bg, p.dim, 0.44), creamStoneAnchor, 0.40)
		return blend(face, earthAnchor, 0.16)
	}

	s.houseProfile = profileStoneClassical // pale ashlar body + cap + fluting — reads as cream townhouses
	s.wonderMotif = wonderDome             // renaissance centrepiece: a great domed rotunda + lantern
	s.slotSpacing = 1.75                   // a packed, monumental civic core
	return s
}()

// colonialCityStyle is the tuned COLONIAL preset (Phase 1b-iii). A BRICK-AND-TIMBER frontier town:
// terraced ROWHOUSES under warm fired-brick roofs (profileRowhouse — a row of narrow attached units,
// earthier + redder than the ancient clay), packed-DIRT/brick lanes, modest kitchen greenery, and a
// stout TIMBER PALISADE-FORT (wallTimber, reused from iron — no new wall) ringing the settlement. The
// centrepiece is the grand generic hall (wonderGeneric) read as a colonial STATEHOUSE — no bespoke
// wonder. Built from defaultTdStyle so it keeps the tuned ground texture / pond / filler behaviour,
// then re-skinned warm brick; every tone stays a theme-role recipe so the whole town retints on a
// theme switch. Reads clearly apart from the pale-stone renaissance city (warm brick + timber vs
// cream ashlar + star-fort) AND from the not-yet-tuned default village (rowhouses + a wall + brick).
var colonialCityStyle = func() tdEraStyle {
	s := defaultTdStyle
	s.name = "colonial"

	// Brick-red roof: a warm fired-brick fill — a shade EARTHIER + redder than the ancient terracotta
	// clay. Background lifted toward text, then pulled firmly to the brick anchor so a rowhouse reads
	// as fired brick, not thatch or pale stone.
	s.roofBase = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.30), brickRedAnchor, 0.54)
	}
	s.roofDark = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.text, 0.30), brickRedAnchor, 0.54), 0.28)
	}
	s.lineageMix = 0.15 // keep the subtle lineage tint; the sat cap still guards the no-accent rule

	// Ground: packed DIRT / brick — a warm trodden earth, drier than the primitive dirt+grass but not
	// as pale as the ancient stone. Base pulled toward the dirt anchor with a whisper of brick warmth.
	s.groundBase = func(p tdPal) color.RGBA {
		earthy := blend(blend(p.bg, p.dim, 0.28), dirtAnchor, 0.40)
		return blend(earthy, brickRedAnchor, 0.10)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.24), earthAnchor, 0.26)
	}
	// Streets: packed-dirt/brick lanes — the trodden earth leaning a touch redder/warmer than the
	// ground so the gaps read as brick-edged dirt roads.
	s.streetCol = func(p tdPal) color.RGBA {
		packed := blend(blend(p.bg, p.dim, 0.28), dirtAnchor, 0.44)
		return blend(blend(packed, p.text, 0.28), brickRedAnchor, 0.16)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		packed := blend(blend(p.bg, p.dim, 0.28), dirtAnchor, 0.44)
		surface := blend(blend(packed, p.text, 0.28), brickRedAnchor, 0.16)
		return darken(surface, 0.20)
	}
	// Town-square paving: packed brick-earth, a shade lighter + warmer than the lanes.
	s.pavedCol = func(p tdPal) color.RGBA {
		earthy := blend(blend(p.bg, p.dim, 0.28), dirtAnchor, 0.36)
		return blend(blend(earthy, p.text, 0.30), brickRedAnchor, 0.18)
	}

	// Modest greenery: kitchen plots + street trees a touch drier/duller than the lush primitive green
	// (a working frontier town, not a garden village), still theme-derived so they retint.
	s.gardenCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, grassAnchor, 0.40), dirtAnchor, 0.14)
	}
	s.treeCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, grassAnchor, 0.46), p.dim, 0.10)
	}

	// Wall: a stout TIMBER PALISADE-FORT (wallTimber — reused from iron, NOT a new wall) — the
	// timber-brown anchor grounded so it reads as a log stockade. Medium thickness (set in tdAddWalls),
	// no stone towers.
	s.hasWalls = true
	s.wallProfile = wallTimber
	s.wallCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.28), timberAnchor, 0.50)
	}

	s.houseProfile = profileRowhouse // terraced brick rowhouses
	s.wonderMotif = wonderGeneric    // colonial centrepiece: the grand generic hall read as a statehouse
	s.slotSpacing = 1.6              // a packed frontier town — tighter than the ancient/medieval city
	return s
}()

// industrialCityStyle is the tuned INDUSTRIAL preset (Phase 1b-iii). A grimy RED-BRICK factory town:
// dense terraced brick rowhouses under dull TIN (corrugated grey metal) roofs (profileRowhouse reused
// from colonial), a SOOTY darkened/greyed ground (a dark soot tone blended in — clearly dirtier than
// colonial), NO walls (the age of open industry), and a great FACTORY hall + SMOKESTACKS centrepiece
// (wonderFactory). Scattered smokestacks dot the skyline (tdAddFiller). Built from defaultTdStyle,
// then re-skinned grimy + denser. Reads clearly apart from colonial: colonial is warm brick rowhouses
// + a timber palisade + greenery; industrial is grimier red-brick + tin roofs + smokestacks + a sooty
// dark ground + NO walls + DENSER.
var industrialCityStyle = func() tdEraStyle {
	s := defaultTdStyle
	s.name = "industrial"

	// Roof material: dull TIN — corrugated grey metal. The industrial HOUSE (profileRowhouse) draws its
	// brick body internally; roofBase here is the tin roof tone (a cool dull grey), so a rowhouse reads
	// as red brick under grey tin. Non-house roofs then read as grimy metal, which suits the mood.
	s.roofBase = func(p tdPal) color.RGBA {
		tin := blend(blend(p.bg, p.dim, 0.24), tinAnchor, 0.52)
		return blend(tin, sootAnchor, 0.14) // grimed with soot
	}
	s.roofDark = func(p tdPal) color.RGBA {
		tin := blend(blend(p.bg, p.dim, 0.24), tinAnchor, 0.52)
		return darken(blend(tin, sootAnchor, 0.20), 0.26)
	}
	s.lineageMix = 0.12 // a whisper of lineage tint over the grey tin; sat cap guards it

	// Ground: SOOTY — the packed dirt darkened + greyed with a dark soot tone, clearly grimier than the
	// warm colonial earth. Base pulled toward dirt then firmly toward soot so the floor reads coal-dusted.
	s.groundBase = func(p tdPal) color.RGBA {
		earthy := blend(blend(p.bg, p.dim, 0.30), dirtAnchor, 0.34)
		return blend(earthy, sootAnchor, 0.34)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(blend(p.bg, p.dim, 0.26), dirtAnchor, 0.24), sootAnchor, 0.30)
	}
	// Streets: grimy soot-dark lanes — the sooty ground lifted a touch toward the neutral so the gaps
	// read as worked coal-dusted roads, still darker than the colonial dirt streets.
	s.streetCol = func(p tdPal) color.RGBA {
		packed := blend(blend(blend(p.bg, p.dim, 0.30), dirtAnchor, 0.36), sootAnchor, 0.30)
		return blend(packed, p.text, 0.22)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		packed := blend(blend(blend(p.bg, p.dim, 0.30), dirtAnchor, 0.36), sootAnchor, 0.30)
		return darken(blend(packed, p.text, 0.22), 0.22)
	}
	// Town-square paving: soot-stained flag/brick, a shade lighter than the lanes but still grimy.
	s.pavedCol = func(p tdPal) color.RGBA {
		packed := blend(blend(blend(p.bg, p.dim, 0.28), dirtAnchor, 0.30), sootAnchor, 0.28)
		return blend(packed, p.text, 0.26)
	}

	// Greenery: sparse + sooty — the little green that survives is dull and coal-dusted, so the town
	// reads industrial, not garden. Still theme-derived so it retints.
	s.gardenCol = func(p tdPal) color.RGBA {
		return blend(blend(blend(p.bg, grassAnchor, 0.34), p.dim, 0.14), sootAnchor, 0.16)
	}
	s.treeCol = func(p tdPal) color.RGBA {
		return blend(blend(blend(p.bg, grassAnchor, 0.38), p.dim, 0.12), sootAnchor, 0.14)
	}

	// NO walls — the age of open industry (hasWalls false, wallProfile wallNone).
	s.hasWalls = false
	s.wallProfile = wallNone

	s.houseProfile = profileRowhouse // dense brick terraces under tin roofs
	s.wonderMotif = wonderFactory    // industrial centrepiece: a factory hall + smokestacks
	s.slotSpacing = 1.5              // DENSER than colonial (1.6) — packed industrial terraces
	return s
}()

// victorianCityStyle is the tuned VICTORIAN preset (V3-B ELECTRIC epoch). A genteel BROWNSTONE city:
// terraced ROWHOUSES (profileRowhouse, reused from colonial/industrial) under warm dark-CHOCOLATE
// brownstone roofs — deeper + browner than the colonial fired-brick and cleaner than the grimy
// industrial tin — over STONE-PAVED streets, dressed with GASLIT PARKS (green squares ringed with
// warm gas-lamps). NO walls (the industrial city already tore them down; the age of open boulevards).
// The centrepiece is the grand generic hall (wonderGeneric) read as a Victorian TERMINAL/MUSEUM — no
// bespoke wonder. Built from colonialCityStyle (which already gives the rowhouse profile + tuned
// filler), then re-skinned brownstone + stone-paved + park-green; every tone stays a theme-role recipe
// so the whole city retints on a theme switch. Reads clearly apart from colonial (brownstone + stone
// pavers + parks + NO wall vs warm brick + dirt lanes + a timber palisade) AND from industrial
// (genteel brownstone + parks vs grimy tin + soot + smokestacks) AND from the default village.
var victorianCityStyle = func() tdEraStyle {
	s := colonialCityStyle
	s.name = "victorian"

	// Roof material: warm dark CHOCOLATE brownstone — deeper + browner than the colonial fired brick.
	// Background lifted toward text, then pulled firmly to the brownstone anchor so a rowhouse reads as
	// dressed chocolate stone, not red brick or grey tin.
	s.roofBase = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.26), brownstoneAnchor, 0.58)
	}
	s.roofDark = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.text, 0.26), brownstoneAnchor, 0.58), 0.30)
	}
	s.lineageMix = 0.13 // a whisper of lineage tint over the brownstone; the sat cap still guards it

	// Ground: dressed STONE paving underfoot — cooler + greyer than the colonial packed dirt, a genteel
	// paved city floor. Base grounded toward the pale stone anchor with a whisper of the brownstone warmth.
	s.groundBase = func(p tdPal) color.RGBA {
		stony := blend(blend(p.bg, p.dim, 0.26), stoneAnchor, 0.40)
		return blend(stony, brownstoneAnchor, 0.08)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.22), graniteAnchor, 0.26)
	}
	// Streets: STONE-PAVED lanes — the stony ground lifted toward the light neutral + a touch of granite
	// so the gaps read as dressed-stone paving between the terraces, not dirt.
	s.streetCol = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.26), stoneAnchor, 0.44)
		return blend(blend(paved, p.text, 0.34), graniteAnchor, 0.22)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.26), stoneAnchor, 0.44)
		surface := blend(blend(paved, p.text, 0.34), graniteAnchor, 0.22)
		return darken(surface, 0.20)
	}
	// Town-square paving: the brightest dressed stone in the city — a genteel civic plaza.
	s.pavedCol = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.22), stoneAnchor, 0.48)
		return blend(pale, p.text, 0.26)
	}

	// Gaslit PARKS: manicured greenery — a touch lusher + tidier than the frontier colonial kitchen plots
	// (a genteel park square, not a working yard), still theme-derived so it retints.
	s.gardenCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, grassAnchor, 0.48), p.dim, 0.06)
	}
	s.treeCol = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.dim, 0.24), grassAnchor, 0.52), 0.08)
	}

	// NO walls — the age of open boulevards (industrial already tore the ring down).
	s.hasWalls = false
	s.wallProfile = wallNone

	s.houseProfile = profileRowhouse // terraced brownstone rowhouses
	s.wonderMotif = wonderGeneric    // victorian centrepiece: the grand generic hall read as a terminal/museum
	s.slotSpacing = 1.5              // a dense, genteel city — tighter than colonial (1.6), packed terraces
	return s
}()

// electricCityStyle is the tuned ELECTRIC preset (V3-B ELECTRIC epoch). A pale ART-DECO CONCRETE city:
// FLAT-topped modern blocks (profileModernFlat — the groundwork for skyscrapers) in warm pale concrete,
// WIDE dressed-concrete AVENUES, a subtle WARM electric-light accent (the first electric glow — kept
// muted), and — the epoch's first tall centrepiece — an ART-DECO SETBACK TOWER (wonderTower). NO walls
// (the age of open avenues). Built from defaultTdStyle so it keeps the tuned ground texture / pond /
// filler behaviour, then re-skinned pale deco concrete; every tone stays a theme-role recipe so the whole
// city retints on a theme switch. Reads clearly apart from the WARMER, ORNATE, DENSER victorian
// (brownstone rowhouses + gaslit parks) and from the default village — electric is cleaner + paler +
// flat-roofed with a stepped tower.
var electricCityStyle = func() tdEraStyle {
	s := defaultTdStyle
	s.name = "electric"

	// Roof material: pale ART-DECO CONCRETE with a subtle WARM electric-light lift — the first electric
	// glow, kept muted (a whisper of the warm gaslight anchor over the concrete) so a flat block reads as
	// warm pale deco stone catching a little artificial light, never a saturated accent.
	s.roofBase = func(p tdPal) color.RGBA {
		concrete := blend(blend(p.bg, p.text, 0.34), concreteAnchor, 0.54)
		return blend(concrete, gasGlowAnchor, 0.06) // muted warm electric lift
	}
	s.roofDark = func(p tdPal) color.RGBA {
		concrete := blend(blend(p.bg, p.text, 0.34), concreteAnchor, 0.54)
		return darken(concrete, 0.24)
	}
	s.lineageMix = 0.12 // a whisper of lineage tint over the concrete; the sat cap still guards it

	// Ground: pale dressed CONCRETE — a clean warm-grey deco plaza floor, lighter + warmer than the cool
	// industrial soot. Base grounded toward the concrete anchor, lifted a touch to the light neutral.
	s.groundBase = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.22), concreteAnchor, 0.46)
		return blend(pale, p.text, 0.14)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.18), stoneAnchor, 0.30)
	}
	// Streets: WIDE dressed-concrete AVENUES — the concrete ground lifted toward the light neutral so the
	// gaps read as broad pale boulevards, with a faint warm electric cast.
	s.streetCol = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.22), concreteAnchor, 0.44)
		lit := blend(blend(paved, p.text, 0.32), stoneAnchor, 0.16)
		return blend(lit, gasGlowAnchor, 0.05) // muted warm avenue glow
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.22), concreteAnchor, 0.44)
		surface := blend(blend(paved, p.text, 0.32), stoneAnchor, 0.16)
		return darken(surface, 0.20)
	}
	// Town-square paving: the brightest dressed concrete in the city — a grand deco forecourt.
	s.pavedCol = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.16), concreteAnchor, 0.50)
		return blend(pale, p.text, 0.28)
	}

	// NO walls — the age of open avenues.
	s.hasWalls = false
	s.wallProfile = wallNone

	s.houseProfile = profileModernFlat // flat-topped deco blocks
	s.wonderMotif = wonderTower        // electric centrepiece: an art-deco setback tower
	s.slotSpacing = 1.5                // a dense modern downtown — packed flat blocks along wide avenues
	return s
}()

// atomicCityStyle is the tuned ATOMIC preset (V3-B ELECTRIC epoch). A clean MIDCENTURY STEEL-AND-GLASS
// city: FLAT-topped modern blocks (profileModernFlat, reused from electric) in COOL PALE PASTEL concrete
// pushed hard toward steel, over clean pale streets, crowned by a googie SPACE NEEDLE (wonderSpaceNeedle —
// a slender stem + a wide flying-saucer disc). NO walls. AIRIER than electric (a suburb-and-downtown feel —
// a true zoned split is deferred FORM work; approximated here with looser density). Built from
// electricCityStyle (which gives the flat-block profile + no walls), then shifted markedly COOLER + cleaner +
// PASTEL and airier and re-crowned with the space needle. Reads clearly apart from electric: electric is
// WARM + ornate deco + DENSE + a stepped concrete TOWER; atomic is COOL steel-pastel + AIRY + a space-age
// SAUCER needle. Also apart from the default village. Every tone stays a theme-role recipe so the whole city
// retints on a theme switch.
var atomicCityStyle = func() tdEraStyle {
	s := electricCityStyle
	s.name = "atomic"

	// Roof material: COOL PALE PASTEL concrete pushed hard toward STEEL — cleaner + markedly cooler than
	// electric's warm deco concrete (no warm electric lift; a stronger steel body + a mint-pastel cast
	// instead), so a flat block reads as crisp midcentury steel-and-glass, not warm deco stone.
	s.roofBase = func(p tdPal) color.RGBA {
		concrete := blend(blend(p.bg, p.text, 0.34), blend(concreteAnchor, steelAnchor, 0.62), 0.55)
		return blend(concrete, pastelAnchor, 0.20) // cool pale pastel cast (stronger than electric)
	}
	s.roofDark = func(p tdPal) color.RGBA {
		concrete := blend(blend(p.bg, p.text, 0.34), blend(concreteAnchor, steelAnchor, 0.62), 0.55)
		return darken(blend(concrete, steelAnchor, 0.24), 0.22)
	}
	s.lineageMix = 0.10 // an even fainter lineage tint over the cool steel; the sat cap still guards it

	// Ground: cool pale PASTEL-STEEL concrete — cleaner + cooler + airier than electric's warm deco floor.
	// Base grounded toward a cool steel-forward concrete mix, lifted to the light neutral with a pastel cast.
	s.groundBase = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.20), blend(concreteAnchor, steelAnchor, 0.60), 0.44)
		return blend(blend(pale, p.text, 0.16), pastelAnchor, 0.18)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.16), steelAnchor, 0.38)
	}
	// Streets: clean pale midcentury boulevards — the cool steel-concrete lifted toward the light neutral, no
	// warm electric cast (that's electric's tell), a touch cooler + cleaner with a stronger pastel wash.
	s.streetCol = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.20), blend(concreteAnchor, steelAnchor, 0.60), 0.42)
		return blend(blend(paved, p.text, 0.34), pastelAnchor, 0.18)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.20), blend(concreteAnchor, steelAnchor, 0.60), 0.42)
		surface := blend(blend(paved, p.text, 0.34), pastelAnchor, 0.18)
		return darken(surface, 0.20)
	}
	// Town-square paving: the brightest clean pale steel-concrete in the city — a midcentury civic plaza.
	s.pavedCol = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.14), blend(concreteAnchor, steelAnchor, 0.60), 0.48)
		return blend(blend(pale, p.text, 0.30), pastelAnchor, 0.16)
	}

	// NO walls — inherited from electric, restated for clarity.
	s.hasWalls = false
	s.wallProfile = wallNone

	s.houseProfile = profileModernFlat // flat-topped midcentury blocks
	s.wonderMotif = wonderSpaceNeedle  // atomic centrepiece: a googie space needle (a saucer on a stem)
	s.slotSpacing = 1.9                // AIRIER than electric (1.5) — a suburb-and-downtown feel, looser than before
	return s
}()

// modernCityStyle is the tuned MODERN preset (V3-C DIGITAL epoch). A clean GLASS-AND-STEEL city: TALL
// glass SKYSCRAPERS (profileGlassTower — a cool blue-grey curtain-wall slab + a lit window grid + a long
// height shadow) over clean paved AVENUES, sparse greenery, NO walls, and a SUPERTALL glass tower
// centrepiece (wonderSkyscraper). Built from defaultTdStyle so it keeps the tuned ground texture / pond /
// filler behaviour, then re-skinned cool glass-blue; every tone stays a theme-role recipe so the whole
// city retints on a theme switch. Reads clearly apart from the pale-CONCRETE deco/midcentury electric +
// atomic (warm/pastel flat blocks + a stepped concrete tower) — modern is COOLER, glassier, and taller.
var modernCityStyle = func() tdEraStyle {
	s := defaultTdStyle
	s.name = "modern"

	// Roof material: cool blue-grey GLASS. The skyscraper draws its glass curtain-wall internally off rc,
	// so roofBase is the glass tone (a steely blue-grey) and non-tower roofs read as glass panels too.
	s.roofBase = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.30), glassAnchor, 0.56)
	}
	s.roofDark = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.text, 0.30), glassAnchor, 0.56), 0.24)
	}
	s.lineageMix = 0.10 // a whisper of lineage tint over the glass; the sat cap still guards it

	// Ground: clean cool CONCRETE-AND-STEEL plaza floor, a shade bluer than the atomic pastel. Base
	// grounded toward a steel/concrete mix, lifted to the light neutral.
	s.groundBase = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.22), blend(concreteAnchor, steelAnchor, 0.60), 0.46)
		return blend(pale, p.text, 0.14)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.18), steelAnchor, 0.34)
	}
	// Streets: clean pale-steel AVENUES — the cool ground lifted toward the light neutral, a touch bluer.
	s.streetCol = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.22), blend(concreteAnchor, steelAnchor, 0.60), 0.44)
		return blend(paved, p.text, 0.34)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.22), blend(concreteAnchor, steelAnchor, 0.60), 0.44)
		return darken(blend(paved, p.text, 0.34), 0.20)
	}
	// Town-square paving: the brightest clean steel-grey plaza in the city — a glass-tower forecourt.
	s.pavedCol = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.14), blend(concreteAnchor, steelAnchor, 0.60), 0.50)
		return blend(pale, p.text, 0.30)
	}

	// Greenery: SPARSE — a downtown of glass and pavement, little green. Still theme-derived so it retints.
	s.gardenCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, grassAnchor, 0.30), p.dim, 0.16)
	}
	s.treeCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, grassAnchor, 0.34), p.dim, 0.14)
	}

	// NO walls — the age of open glass towers.
	s.hasWalls = false
	s.wallProfile = wallNone

	s.houseProfile = profileGlassTower // tall glass-and-steel skyscrapers
	s.wonderMotif = wonderSkyscraper   // modern centrepiece: a supertall glass tower
	s.slotSpacing = 1.5                // a dense modern downtown — packed glass towers along wide avenues
	return s
}()

// informationCityStyle is the tuned INFORMATION preset (V3-C DIGITAL epoch). The modern glass city gone
// DENSER + COLDER, and re-crowned as a SERVER-CITY: the same glass SKYSCRAPERS (profileGlassTower) but under
// a markedly colder data-grey cast, packed tighter (slotSpacing 1.35), dotted with more low wide DATA-CENTER
// blocks (tdPropDataCenter — server farms with blinking lights), and centred on a bespoke DATA-HUB wonder
// (wonderDataHub — a wide server-farm base + a comms antenna + cyan/amber beacons) instead of another glass
// tower. Built from modernCityStyle, then shifted markedly colder + denser + re-crowned. Reads clearly apart
// from modern: modern is clean BLUE GLASS at downtown density with a glass-tower centrepiece; information is a
// denser, COLDER data-grey server-city crowned by a data hub.
var informationCityStyle = func() tdEraStyle {
	s := modernCityStyle
	s.name = "information"

	// Roof material: the glass gone markedly COLDER — pulled hard toward the cold data-grey anchor so the
	// towers read as a colder server-city curtain-wall, distinctly bluer/greyer than the modern glass.
	s.roofBase = func(p tdPal) color.RGBA {
		glass := blend(blend(p.bg, p.text, 0.30), glassAnchor, 0.52)
		return blend(glass, dataGreyAnchor, 0.30) // colder data-grey cast (stronger than before)
	}
	s.roofDark = func(p tdPal) color.RGBA {
		glass := blend(blend(p.bg, p.text, 0.30), glassAnchor, 0.52)
		return darken(blend(glass, dataGreyAnchor, 0.38), 0.24)
	}

	// Ground: a markedly colder data-grey floor — the modern steel plaza pulled hard toward the cold
	// server-grey anchor.
	s.groundBase = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.24), blend(steelAnchor, dataGreyAnchor, 0.68), 0.46)
		return blend(pale, p.text, 0.10)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.20), dataGreyAnchor, 0.46)
	}
	s.streetCol = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.24), blend(steelAnchor, dataGreyAnchor, 0.68), 0.44)
		return blend(paved, p.text, 0.28)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		paved := blend(blend(p.bg, p.dim, 0.24), blend(steelAnchor, dataGreyAnchor, 0.68), 0.44)
		return darken(blend(paved, p.text, 0.28), 0.20)
	}
	s.pavedCol = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.dim, 0.16), blend(steelAnchor, dataGreyAnchor, 0.68), 0.48)
		return blend(pale, p.text, 0.26)
	}

	s.houseProfile = profileGlassTower // denser glass towers
	s.wonderMotif = wonderDataHub      // information centrepiece: a data-hub megastructure (server farm + comms mast)
	s.slotSpacing = 1.35               // DENSER than modern (1.5) — a packed server-city, tighter than before
	return s
}()

// digitalCityStyle is the tuned DIGITAL preset (V3-C DIGITAL epoch). The information server-city gone
// SLEEK + DARKER with the epoch's FIRST NEON: the same glass SKYSCRAPERS (profileGlassTower) + supertall
// centrepiece (wonderSkyscraper), but over a DARKER hi-tech ground with restrained neon CYAN/MAGENTA
// accents in the streets, roof sheen, and props (a few NEON-SIGN dabs) — a first hint of the cyberpunk
// epoch to come, kept "first neon", not full cyberpunk. Built from informationCityStyle, then darkened +
// neon-accented. Reads clearly apart from information (which is a cold grey server-city, no neon): digital
// is darker, sleeker, and the first age with a neon glow.
var digitalCityStyle = func() tdEraStyle {
	s := informationCityStyle
	s.name = "digital"

	// Roof material: DARKER sleek glass with a faint NEON-CYAN sheen — the first neon, kept restrained (a
	// whisper of cyan over the darkened glass) so a tower reads as sleek dark glass catching a little neon,
	// never a saturated slab.
	s.roofBase = func(p tdPal) color.RGBA {
		glass := darken(blend(blend(p.bg, p.text, 0.24), glassAnchor, 0.52), 0.16) // darker sleek glass
		return blend(glass, neonCyanAnchor, 0.07)                                  // muted first-neon lift
	}
	s.roofDark = func(p tdPal) color.RGBA {
		glass := darken(blend(blend(p.bg, p.text, 0.24), glassAnchor, 0.52), 0.16)
		return darken(glass, 0.24)
	}

	// Ground: a DARKER hi-tech floor — the cold data-grey ground pulled down + a faint cool neon cast, so
	// the sleek dark streets read as the run-up to the cyberpunk night city.
	s.groundBase = func(p tdPal) color.RGBA {
		dark := darken(blend(blend(p.bg, p.dim, 0.30), dataGreyAnchor, 0.46), 0.18)
		return blend(dark, neonCyanAnchor, 0.04)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.dim, 0.26), dataGreyAnchor, 0.42), 0.20)
	}
	// Streets: sleek dark lanes lit by NEON — the dark ground lifted a touch toward the neutral with a
	// restrained cyan/magenta neon cast (the first neon streetlights).
	s.streetCol = func(p tdPal) color.RGBA {
		dark := darken(blend(blend(p.bg, p.dim, 0.30), dataGreyAnchor, 0.46), 0.14)
		lit := blend(dark, p.text, 0.20)
		return blend(lit, blend(neonCyanAnchor, neonMagentaAnchor, 0.35), 0.08) // first-neon street glow
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		dark := darken(blend(blend(p.bg, p.dim, 0.30), dataGreyAnchor, 0.46), 0.14)
		return darken(blend(dark, p.text, 0.20), 0.22)
	}
	s.pavedCol = func(p tdPal) color.RGBA {
		dark := darken(blend(blend(p.bg, p.dim, 0.24), dataGreyAnchor, 0.44), 0.12)
		lit := blend(dark, p.text, 0.24)
		return blend(lit, neonCyanAnchor, 0.06)
	}
	// Prop tone leans neon so the scattered neon-sign + data-center dabs read as the first neon glow.
	s.propCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.30), neonCyanAnchor, 0.10)
	}

	s.houseProfile = profileGlassTower // sleek dark glass towers
	s.wonderMotif = wonderSkyscraper   // digital centrepiece: a supertall glass tower
	s.slotSpacing = 1.4                // as dense as the information server-city
	return s
}()

// cyberpunkCityStyle is the tuned CYBERPUNK preset (NEON epoch). The digital age's FIRST neon pushed to
// its full, unrestrained MAX: a NIGHT CITY of very DARK ground + saturated neon CYAN/MAGENTA drenching the
// streets, paving, and roof sheen, packed into dense blocky MEGASTRUCTURES (the same profileGlassTower gone
// darker, at even tighter slotSpacing). The centrepiece stays a dark neon MEGATOWER (wonderSkyscraper),
// and the square/scatter swap the digital age's restrained neon-signs for translucent floating HOLOGRAMS
// (tdPropHologram). Built from digitalCityStyle, then darkened HARD + neon pushed from a whisper to a blaze.
// Reads clearly apart from digital (which is only the first restrained hint): cyberpunk is the full night
// city — darker ground, brighter neon, denser blocks, holograms instead of signs.
var cyberpunkCityStyle = func() tdEraStyle {
	s := digitalCityStyle
	s.name = "cyberpunk"

	// Roof material: sleek glass gone even DARKER with a much STRONGER neon-cyan sheen — where digital was a
	// whisper (0.07) of neon over lightly-darkened glass, cyberpunk drops the glass harder and blazes the
	// neon (0.20), so a megablock reads as dark glass catching a full neon wash.
	s.roofBase = func(p tdPal) color.RGBA {
		glass := darken(blend(blend(p.bg, p.text, 0.22), glassAnchor, 0.50), 0.30) // deep dark glass
		return blend(glass, neonCyanAnchor, 0.20)                                  // full neon sheen
	}
	s.roofDark = func(p tdPal) color.RGBA {
		glass := darken(blend(blend(p.bg, p.text, 0.22), glassAnchor, 0.50), 0.30)
		return blend(darken(glass, 0.22), neonMagentaAnchor, 0.10) // a magenta bruise in the shade
	}

	// Ground: a very DARK night-city floor — the digital dark ground pushed much darker with a stronger cool
	// neon cast, so the whole city sits in a saturated neon night.
	s.groundBase = func(p tdPal) color.RGBA {
		dark := darken(blend(blend(p.bg, p.dim, 0.34), dataGreyAnchor, 0.42), 0.36)
		return blend(dark, neonCyanAnchor, 0.09)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		dark := darken(blend(blend(p.bg, p.dim, 0.30), dataGreyAnchor, 0.40), 0.38)
		return blend(dark, neonMagentaAnchor, 0.07)
	}
	// Streets: dark canyons blazing with neon — the dark ground lifted a touch then drenched in a strong
	// cyan/magenta neon wash (the full night-city streetglow, no longer restrained).
	s.streetCol = func(p tdPal) color.RGBA {
		dark := darken(blend(blend(p.bg, p.dim, 0.34), dataGreyAnchor, 0.44), 0.30)
		lit := blend(dark, p.text, 0.18)
		return blend(lit, blend(neonCyanAnchor, neonMagentaAnchor, 0.45), 0.22) // full neon street wash
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		dark := darken(blend(blend(p.bg, p.dim, 0.34), dataGreyAnchor, 0.44), 0.30)
		return darken(blend(dark, p.text, 0.18), 0.26)
	}
	s.pavedCol = func(p tdPal) color.RGBA {
		dark := darken(blend(blend(p.bg, p.dim, 0.30), dataGreyAnchor, 0.42), 0.26)
		lit := blend(dark, p.text, 0.22)
		return blend(lit, neonMagentaAnchor, 0.16) // a magenta plaza glow
	}
	// Prop tone blazes neon so the scattered holograms read as bright floating projections.
	s.propCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.34), neonCyanAnchor, 0.24)
	}

	s.houseProfile = profileGlassTower // dark neon megablocks
	s.wonderMotif = wonderSkyscraper   // cyberpunk centrepiece: a dark neon megatower
	s.slotSpacing = 1.35               // even DENSER than digital (1.4) — packed megastructures
	return s
}()

// fusionCityStyle is the tuned FUSION preset (NEON epoch) — a deliberate CONTRAST to cyberpunk's dark neon
// night. A CLEAN, BRIGHT, utopian WHITE city: brilliant white structures (fusionWhiteAnchor) over pale clean
// paving, lifted with a soft pale CYAN glow (fusionCyanAnchor), spaced AIRY (a wide slotSpacing so the city
// breathes). The centrepiece is a glowing REACTOR (wonderFusionCore — concentric cyan rings around a
// white-hot bloom). Built from defaultTdStyle (so it keeps the tuned ground texture / pond / filler
// behaviour) then re-skinned white-and-cyan. Reads STRONGLY apart from cyberpunk: where cyberpunk is dark +
// saturated + dense, fusion is bright + white + airy — the two sit at opposite ends of the light scale.
var fusionCityStyle = func() tdEraStyle {
	s := defaultTdStyle
	s.name = "fusion"

	// Roof material: brilliant clean WHITE glass lifted with a pale cyan glow — a utopian white curtain-wall,
	// far brighter + cooler than any stone/marble, so the buildings read as gleaming fusion-era towers.
	s.roofBase = func(p tdPal) color.RGBA {
		white := blend(blend(p.bg, p.text, 0.30), fusionWhiteAnchor, 0.66)
		return blend(white, fusionCyanAnchor, 0.10) // soft cyan glow lift
	}
	s.roofDark = func(p tdPal) color.RGBA {
		white := blend(blend(p.bg, p.text, 0.30), fusionWhiteAnchor, 0.54)
		return darken(blend(white, fusionCyanAnchor, 0.10), 0.14) // a gentle, still-pale shade
	}
	s.lineageMix = 0.10 // keep the white clean — barely any lineage tint

	// Ground: a pale clean plaza floor — bright and airy, lifted toward white with a faint cyan cast, so the
	// city sits on gleaming paving rather than dirt.
	s.groundBase = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.text, 0.22), fusionWhiteAnchor, 0.44)
		return blend(pale, fusionCyanAnchor, 0.06)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.18), fusionWhiteAnchor, 0.34)
	}
	s.streetCol = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.text, 0.26), fusionWhiteAnchor, 0.48)
		return blend(pale, fusionCyanAnchor, 0.08)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.text, 0.26), fusionWhiteAnchor, 0.48)
		return darken(pale, 0.12)
	}
	s.pavedCol = func(p tdPal) color.RGBA {
		pale := blend(blend(p.bg, p.text, 0.24), fusionWhiteAnchor, 0.52)
		return blend(pale, fusionCyanAnchor, 0.10)
	}
	s.propCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.34), fusionCyanAnchor, 0.30)
	}

	s.houseProfile = profileGlassTower // clean white glass towers
	s.wonderMotif = wonderFusionCore   // fusion centrepiece: a glowing cyan reactor core
	s.hasWalls = false                 // a utopian open city
	s.slotSpacing = 1.6                // AIRY — the minimalist city breathes (wider than digital/cyberpunk)
	return s
}()

// spaceCityStyle is the tuned SPACE preset (NEON epoch) — a space-colony read: pale METALLIC SILVER
// (metalSilverAnchor), a COLDER sheen than fusion's warm white. Where fusion is a bright warm-white utopia,
// space is cooler, greyer, faintly blue — a habitat of pressurised metal DOMES (profileMetalDome dwellings)
// under a ROCKET LAUNCHPAD centrepiece (wonderLaunchpad). Built from fusionCityStyle (so it keeps the airy,
// open, pale layout) then re-skinned toward cold silver metal, with scattered ROCKET dabs (tdPropRocket)
// seasoning the spaceport. Reads apart from fusion: fusion is warm bright white, space is cold pale silver;
// apart from cyberpunk: cyberpunk is dark neon, space is pale metallic.
var spaceCityStyle = func() tdEraStyle {
	s := fusionCityStyle
	s.name = "space"

	// Roof material: pale cool SILVER metal — a colder metallic sheen than fusion's white, faintly blue, so
	// the domes read as pressurised metal habitats rather than white glass.
	s.roofBase = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.28), metalSilverAnchor, 0.62)
	}
	s.roofDark = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.text, 0.28), metalSilverAnchor, 0.52), 0.18)
	}

	// Ground: a cold pale metal deck — the fusion pale floor pulled toward cold silver, so the colony floor
	// reads as a plated surface, cooler + greyer than fusion's warm-white plaza.
	s.groundBase = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.20), metalSilverAnchor, 0.44)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.16), metalSilverAnchor, 0.34)
	}
	s.streetCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.18), metalSilverAnchor, 0.48)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.dim, 0.18), metalSilverAnchor, 0.48), 0.14)
	}
	s.pavedCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.16), metalSilverAnchor, 0.52)
	}
	s.propCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.30), metalSilverAnchor, 0.44)
	}

	s.houseProfile = profileMetalDome // space-colony metallic dome dwellings
	s.wonderMotif = wonderLaunchpad   // space centrepiece: a rocket on a launch pad
	s.hasWalls = false                // an open colony
	s.slotSpacing = 1.55              // airy like fusion, a touch tighter
	s.spaceMode = true                // SPACE-AND-ABOVE: void + starfield ground, no greenery (Phase 2c)
	return s
}()

// interstellarCityStyle is the tuned INTERSTELLAR preset (COSMIC epoch, first age). Built from
// spaceCityStyle but stepped OFF the pale space colony into DEEPER SPACE: the ground drops from a
// pale metal deck to a cold STARFIELD-ish deck (pulled hard toward voidAnchor — a dark blue-black
// void), while the STRUCTURES stay pale metallic silver, so bright arcology metal reads against a
// dark ground (the opposite contrast of the pale-on-pale space colony). Dwellings become tall
// tapering SPIRES (profileSpire) instead of low domes, and the centrepiece is a SPIRE CLUSTER
// (wonderSpireArray). Reads apart from space: space is pale metal on a pale deck, interstellar is
// pale metal SPIRES on a dark void deck.
var interstellarCityStyle = func() tdEraStyle {
	s := spaceCityStyle
	s.name = "interstellar"

	// Ground: a COLD DEEP-SPACE deck — the space pale floor dropped HARD toward the void so the
	// colony floor reads as a dark starfield platform, not a lit plaza. Still theme-derived (starts
	// from bg/dim) then pulled deep with voidAnchor + a faint silver fleck so it isn't a dead black.
	s.groundBase = func(p tdPal) color.RGBA {
		deep := blend(blend(p.bg, p.dim, 0.30), voidAnchor, 0.66)
		return blend(deep, metalSilverAnchor, 0.08)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.24), voidAnchor, 0.58)
	}
	s.streetCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.26), voidAnchor, 0.54)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.dim, 0.26), voidAnchor, 0.60), 0.14)
	}
	s.pavedCol = func(p tdPal) color.RGBA {
		// The plaza reads as a lit metal apron floating on the dark deck (brighter than the ground so
		// the spires sit on a defined platform), still silver-metallic.
		return blend(blend(p.bg, p.dim, 0.20), metalSilverAnchor, 0.44)
	}
	// Faint filler on the void deck leans metallic, not green (there is no soil out here).
	s.gardenCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.24), voidAnchor, 0.44)
	}
	s.treeCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.22), metalSilverAnchor, 0.34)
	}

	s.houseProfile = profileSpire    // deep-space arcology SPIRES (tall, narrow, metallic)
	s.wonderMotif = wonderSpireArray // interstellar centrepiece: a cluster of spires around a central mast
	s.hasWalls = false               // still an open colony
	s.slotSpacing = 1.45             // a touch tighter than space — a denser arcology
	s.spaceMode = true               // deep-space void + starfield ground, no greenery (Phase 2c)
	return s
}()

// galacticCityStyle is the tuned GALACTIC preset (COSMIC epoch, second age). Built from
// interstellarCityStyle (so it keeps the deep-space void deck + metallic structures) then pushed to
// a grander MEGASTRUCTURE feel: the ground brightens a shade toward energetic metal (a lit
// megastation floor rather than a bare void), dwellings return to grand metallic DOMES
// (profileMetalDome, denser than interstellar's scattered spires), and the centrepiece is the
// signature RING-HUB (wonderRingHub — concentric orbital rings around a glowing hub). Reads apart
// from interstellar: interstellar is scattered spires on a dark deck, galactic is a dense
// dome-metropolis under a ringworld megastation.
var galacticCityStyle = func() tdEraStyle {
	s := interstellarCityStyle
	s.name = "galactic"

	// Ground: the interstellar void deck lifted a shade toward energetic metal — a lit megastation
	// floor with a faint cyan-energy cast, so the galactic city reads as a powered megastructure, not
	// the darker deep-space frontier. Still void-derived (keeps the cosmic family) but brighter.
	s.groundBase = func(p tdPal) color.RGBA {
		deep := blend(blend(p.bg, p.dim, 0.28), voidAnchor, 0.52)
		lit := blend(deep, metalSilverAnchor, 0.22)
		return blend(lit, energyAnchor, 0.08)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(blend(p.bg, p.dim, 0.24), voidAnchor, 0.48), metalSilverAnchor, 0.18)
	}
	s.streetCol = func(p tdPal) color.RGBA {
		return blend(blend(blend(p.bg, p.dim, 0.24), voidAnchor, 0.44), metalSilverAnchor, 0.24)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		return darken(blend(blend(blend(p.bg, p.dim, 0.24), voidAnchor, 0.50), metalSilverAnchor, 0.20), 0.12)
	}
	s.pavedCol = func(p tdPal) color.RGBA {
		// A brighter energetic apron than interstellar — the megastation plaza catches the ring glow.
		return blend(blend(blend(p.bg, p.dim, 0.18), metalSilverAnchor, 0.46), energyAnchor, 0.10)
	}

	s.houseProfile = profileMetalDome // galactic dwellings: grand metallic domes (a dense metropolis)
	s.wonderMotif = wonderRingHub     // galactic centrepiece: the ringworld/megastation ring-hub
	s.hasWalls = false                // still open
	s.slotSpacing = 1.30              // denser than interstellar — a grand packed metropolis
	s.spaceMode = true                // deep-space void + starfield ground, no greenery (Phase 2c)
	return s
}()

// quantumCityStyle is the tuned QUANTUM preset (COSMIC epoch, THIRD age — the first of the final pair
// completing all 22). Built from galacticCityStyle (so it keeps the deep-space void deck) then re-cast as
// an IRIDESCENT CRYSTALLINE lattice: the ground darkens back toward the void (a black crystal deck) with a
// faint shifting cyan/magenta sheen, dwellings become faceted crystal NODES (profileLattice — glinting
// gems in shifting hues), and the centrepiece is a great CRYSTAL LATTICE mesh (wonderCrystalLattice). Reads
// apart from galactic: galactic is a warm lit dome-metropolis under a metal ring-hub; quantum is a cold
// black deck of shimmering iridescent crystal. The iridescence lives in the SPRITES (which cycle the three
// irid anchors by position/facet); the preset ground just sets the dark, faintly-prismatic stage.
var quantumCityStyle = func() tdEraStyle {
	s := galacticCityStyle
	s.name = "quantum"

	// Ground: a COLD BLACK CRYSTAL deck — the galactic lit floor dropped hard back toward the void so the
	// crystal structures glint against darkness, then given a faint iridescent cyan+magenta prismatic cast
	// (a shifting sheen on black glass) so the deck itself reads faintly prismatic, not a dead black. Still
	// void-derived (keeps the cosmic family), theme-derived throughout.
	s.groundBase = func(p tdPal) color.RGBA {
		deep := blend(blend(p.bg, p.dim, 0.30), voidAnchor, 0.70)
		sheen := blend(iridCyanAnchor, iridMagentaAnchor, 0.5)
		return blend(deep, sheen, 0.07)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		deep := blend(blend(p.bg, p.dim, 0.26), voidAnchor, 0.64)
		return blend(deep, iridMagentaAnchor, 0.06)
	}
	s.streetCol = func(p tdPal) color.RGBA {
		return blend(blend(blend(p.bg, p.dim, 0.26), voidAnchor, 0.60), iridCyanAnchor, 0.08)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		return darken(blend(blend(p.bg, p.dim, 0.26), voidAnchor, 0.66), 0.12)
	}
	s.pavedCol = func(p tdPal) color.RGBA {
		// The plaza reads as a faceted iridescent apron catching the crystal light — a shade brighter than
		// the black deck with a stronger prismatic sheen, so the lattice sits on a defined glinting platform.
		lit := blend(blend(p.bg, p.dim, 0.22), voidAnchor, 0.48)
		return blend(lit, blend(iridCyanAnchor, iridGoldAnchor, 0.5), 0.16)
	}
	// Faint filler on the crystal deck leans iridescent, not green (no soil out here).
	s.gardenCol = func(p tdPal) color.RGBA {
		return blend(blend(blend(p.bg, p.dim, 0.24), voidAnchor, 0.56), iridMagentaAnchor, 0.10)
	}
	s.treeCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.20), iridCyanAnchor, 0.30)
	}
	s.propCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.18), blend(iridCyanAnchor, iridMagentaAnchor, 0.5), 0.34)
	}

	s.houseProfile = profileLattice      // quantum dwellings: faceted iridescent crystal nodes
	s.wonderMotif = wonderCrystalLattice // quantum centrepiece: a glowing crystal-lattice mesh
	s.hasWalls = false                   // still open
	s.slotSpacing = 1.35                 // a touch airier than galactic — crystal spires want breathing room
	s.spaceMode = true                   // deep-space void + starfield ground, no greenery (Phase 2c)
	return s
}()

// transcendentCityStyle is the tuned TRANSCENDENT preset (COSMIC epoch, FINAL age — the 22nd, completing
// the whole progression). The ETHEREAL finale: built from galacticCityStyle then re-cast as pure LIGHT — a
// bright, pale, LUMINOUS ground (the only near-white deck in the atlas), dwellings that read as soft
// translucent light-forms (profileEthereal — a glowing bloom, not a hard roof), and a centrepiece that is
// an ASCENSION of light (wonderAscension — a rising pillar ringed by soft halos). Reads apart from every
// prior age, and MOST of all from quantum: quantum is a cold black crystal deck; transcendent is a warm
// pale luminous field — the brightest, most dematerialised city, the post-physical climax.
var transcendentCityStyle = func() tdEraStyle {
	s := galacticCityStyle
	s.name = "transcendent"

	// Ground: a PALE LUMINOUS FIELD — the deep-space deck lifted all the way UP toward a warm ethereal white,
	// so the finale reads as a place of light rather than the dark void of every prior cosmic age. Theme-derived
	// (starts from bg lifted toward text/highlight) then washed with the ether white + a soft gold cast so it
	// glows warm, not clinical.
	s.groundBase = func(p tdPal) color.RGBA {
		lift := blend(blend(p.bg, p.text, 0.34), p.highlight, 0.12)
		return blend(blend(lift, etherWhiteAnchor, 0.52), etherGoldAnchor, 0.12)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		lift := blend(blend(p.bg, p.text, 0.30), p.highlight, 0.10)
		return blend(blend(lift, etherWhiteAnchor, 0.44), etherGoldAnchor, 0.16)
	}
	s.streetCol = func(p tdPal) color.RGBA {
		// The "streets" read as faint luminous seams in the light-field — a shade softer than the ground, still pale.
		lift := blend(blend(p.bg, p.text, 0.30), p.highlight, 0.10)
		return blend(lift, etherWhiteAnchor, 0.40)
	}
	s.streetEdge = func(p tdPal) color.RGBA {
		lift := blend(blend(p.bg, p.text, 0.30), p.highlight, 0.10)
		return darken(blend(lift, etherWhiteAnchor, 0.46), 0.08)
	}
	s.pavedCol = func(p tdPal) color.RGBA {
		// The plaza is the BRIGHTEST patch — a near-pure radiant white apron the ascension rises from.
		lift := blend(blend(p.bg, p.text, 0.36), p.highlight, 0.16)
		return blend(blend(lift, etherWhiteAnchor, 0.66), etherGoldAnchor, 0.10)
	}
	// Filler on the light-field: soft pale glimmers, warm-white, not green.
	s.gardenCol = func(p tdPal) color.RGBA {
		return blend(blend(blend(p.bg, p.text, 0.28), etherWhiteAnchor, 0.40), etherGoldAnchor, 0.18)
	}
	s.treeCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.30), etherGoldAnchor, 0.30)
	}
	s.propCol = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.text, 0.24), blend(etherWhiteAnchor, etherGoldAnchor, 0.4), 0.44)
	}

	s.houseProfile = profileEthereal // transcendent dwellings: soft translucent light-form blooms
	s.wonderMotif = wonderAscension  // transcendent centrepiece: a rising ascension of light
	s.hasWalls = false               // still open
	s.slotSpacing = 1.40             // airy — light-forms float, not packed
	// spaceMode (Phase 2c): transcendent joins the SPACE-AND-ABOVE set — the whole GROUND becomes the
	// deep-space void + starfield (drawSpaceBackground), so the luminous ground* recipes above no longer
	// paint the base fill; the light-forms now read as a station floating in the void, not on a bright field.
	s.spaceMode = true
	return s
}()

// stoneAgeStyle is the tuned STONE preset (Phase 1b-i), split off organicVillageStyle so the stone
// age reads distinct from primitive. Dwellings stay THATCH (stone-age huts are still thatch, so
// houseProfile is unchanged) and there are still NO walls — the difference is a ROCKIER, cooler,
// more EXPOSED-STONE ground (a highland settlement, not a green village) and a MEGALITH centrepiece
// (a standing-stone circle) instead of the generic hall. Every tone stays a theme-role recipe so
// the whole city retints on a theme switch.
var stoneAgeStyle = func() tdEraStyle {
	s := organicVillageStyle
	s.name = "stone"

	// Ground: a clearly ROCKY GREY-BROWN highland floor — visibly cooler + more mineral than the
	// primitive earthy/green village. Where primitive leans dirt+grass, stone pulls its base toward
	// packed dirt then GREYS it HARD with the stone + granite anchors (exposed rock, little loam);
	// the alt tone drops primitive's grass cast entirely for a stronger granite fleck, so the
	// texture reads as scattered stone rather than turf. Pushed further (Phase 1b-i pass 2) so the
	// age reads distinct at thumbnail scale. Muted-natural, still theme-derived.
	s.groundBase = func(p tdPal) color.RGBA {
		earthy := blend(blend(p.bg, p.dim, 0.32), dirtAnchor, 0.22)
		rocky := blend(earthy, stoneAnchor, 0.44)
		return blend(rocky, graniteAnchor, 0.22)
	}
	s.groundAlt = func(p tdPal) color.RGBA {
		return blend(blend(p.bg, p.dim, 0.28), graniteAnchor, 0.46)
	}
	// Green filler (gardens/trees) leans GREYER + sparser for the exposed highland: pull the garden
	// green toward the stone family so the few remaining green patches read muted, not lush.
	s.gardenCol = func(p tdPal) color.RGBA {
		green := blend(blend(p.bg, p.dim, 0.20), grassAnchor, 0.34)
		return blend(green, graniteAnchor, 0.22)
	}
	s.treeCol = func(p tdPal) color.RGBA {
		green := darken(blend(blend(p.bg, p.dim, 0.30), grassAnchor, 0.46), 0.10)
		return blend(green, graniteAnchor, 0.18)
	}

	s.houseProfile = profileThatch // stone-age dwellings are still thatch (unchanged from primitive)
	s.wallProfile = wallNone
	s.hasWalls = false
	s.wonderMotif = wonderMegalith // stone centrepiece: a standing-stone circle
	s.slotSpacing = 2.2            // a touch tighter than primitive's airy 2.4, still open highland
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

	// V3-B era-material anchors (ancient + medieval). Like the earthy anchors above, these are
	// NEVER used raw — every recipe blends them against theme roles at a modest strength so a
	// dark or light theme still yields an in-family, muted tone that retints on a theme switch.
	clayAnchor     = color.RGBA{R: 0xb0, G: 0x6a, B: 0x42, A: 0xff} // warm terracotta clay tile (ancient roofs)
	mudbrickAnchor = color.RGBA{R: 0xa8, G: 0x8b, B: 0x63, A: 0xff} // sun-baked tan mudbrick (ancient ground + walls)
	slateAnchor    = color.RGBA{R: 0x4a, G: 0x52, B: 0x5e, A: 0xff} // dark blue-grey slate (medieval roofs)
	cobbleAnchor   = color.RGBA{R: 0x77, G: 0x74, B: 0x70, A: 0xff} // cool cobble/stone grey (medieval ground)
	graniteAnchor  = color.RGBA{R: 0x82, G: 0x84, B: 0x88, A: 0xff} // masonry grey (medieval stone walls + towers)

	// V3-B Phase 1b-ii anchors (iron + classical). Same discipline — never used raw, always blended
	// against theme roles + the era material so a light/dark theme retints them.
	ironAnchor   = color.RGBA{R: 0x51, G: 0x55, B: 0x59, A: 0xff} // cool iron-grey (iron roof accents / worked ground)
	timberAnchor = color.RGBA{R: 0x6e, G: 0x4c, B: 0x2f, A: 0xff} // dark palisade brown (iron timber wall)
	marbleAnchor = color.RGBA{R: 0xcf, G: 0xc9, B: 0xbc, A: 0xff} // pale warm white-stone / marble (classical houses + ground)

	// V3-B renaissance anchors (Phase 1b — renaissance split off medieval). Same discipline — never
	// used raw, always blended against theme roles + the era material so a light/dark theme retints.
	creamStoneAnchor = color.RGBA{R: 0xe3, G: 0xd8, B: 0xbf, A: 0xff} // warm cream/ivory ashlar (renaissance ornate townhouses + civic paving) — LIGHTER + warmer than marble
	leadAnchor       = color.RGBA{R: 0x8f, G: 0x93, B: 0x99, A: 0xff} // pale lead-grey (renaissance dome sheathing + stone accents)

	// V3-B colonial + industrial anchors (Phase 1b-iii). Same discipline — never used raw, always
	// blended against theme roles + neighbouring tones so a light/dark theme retints them.
	brickRedAnchor = color.RGBA{R: 0x9c, G: 0x50, B: 0x3a, A: 0xff} // warm fired brick-red (colonial roofs + industrial house walls) — earthier + redder than clay
	tinAnchor      = color.RGBA{R: 0x8c, G: 0x92, B: 0x96, A: 0xff} // dull corrugated grey tin/zinc (industrial house roofs)
	sootAnchor     = color.RGBA{R: 0x3a, G: 0x37, B: 0x33, A: 0xff} // grimy dark soot/coal (industrial ground + smokestacks + smoke)

	// V3-B ELECTRIC-epoch anchors (victorian / electric / atomic). Same discipline — never used raw,
	// always blended against theme roles + neighbouring tones so a light/dark theme retints them.
	brownstoneAnchor = color.RGBA{R: 0x6b, G: 0x45, B: 0x2f, A: 0xff} // warm dark chocolate brownstone (victorian rowhouse roofs) — deeper + browner than colonial brick-red
	gasGlowAnchor    = color.RGBA{R: 0xd8, G: 0xa8, B: 0x54, A: 0xff} // warm amber gaslight glow (victorian gas-lamp dab) — a soft flame-gold, blended so it never poster-paints
	concreteAnchor   = color.RGBA{R: 0xbf, G: 0xba, B: 0xb0, A: 0xff} // pale warm art-deco concrete/stone (electric flat roofs + avenues) — lighter + warmer than cool tin
	steelAnchor      = color.RGBA{R: 0x9a, G: 0xa1, B: 0xa8, A: 0xff} // cool clean steel-grey (atomic midcentury frames + accents) — cooler + bluer than concrete
	pastelAnchor     = color.RGBA{R: 0xcf, G: 0xd6, B: 0xd4, A: 0xff} // cool pale mint-pastel (atomic midcentury ground/roof cast) — airy, faintly green-blue

	// V3-C DIGITAL-epoch anchors (modern / information / digital). Same discipline — never used raw,
	// always blended against theme roles + neighbouring tones so a light/dark theme retints them. Glass
	// is the cool blue-grey curtain-wall of a skyscraper; the two neon anchors are the digital age's FIRST
	// neon (a restrained hint of the cyberpunk epoch to come) — used only as blended accents, never a slab.
	glassAnchor       = color.RGBA{R: 0x8c, G: 0xa2, B: 0xba, A: 0xff} // cool blue-grey curtain-wall glass (modern skyscraper faces) — steely, faintly blue
	glassLitAnchor    = color.RGBA{R: 0xd4, G: 0xe4, B: 0xf0, A: 0xff} // pale sky-lit glass sheen (the bright window-grid highlights on a glass tower)
	dataGreyAnchor    = color.RGBA{R: 0x6b, G: 0x74, B: 0x7e, A: 0xff} // cold server-farm grey (information data-center blocks + colder cast) — bluer + darker than steel
	neonCyanAnchor    = color.RGBA{R: 0x2f, G: 0xd8, B: 0xd0, A: 0xff} // bright cyan neon (digital first-neon accent) — blended only, never raw
	neonMagentaAnchor = color.RGBA{R: 0xd8, G: 0x3c, B: 0xa8, A: 0xff} // bright magenta neon (digital first-neon accent) — blended only, never raw

	// NEON-epoch anchors (cyberpunk / fusion / space). Same discipline as every anchor above — NEVER used
	// raw; every recipe blends them against theme roles + neighbouring tones so a light/dark theme retints
	// them. The three are chosen to read STRONGLY apart from each other at thumbnail scale: fusion is a
	// clean bright WHITE, its accent a pale CYAN glow (a utopian minimal city); space is a colder pale
	// SILVER metal (a colder sheen than fusion's warm white); the neon pair above already carries cyberpunk.
	fusionWhiteAnchor = color.RGBA{R: 0xf0, G: 0xf4, B: 0xf6, A: 0xff} // brilliant clean white (fusion structures + white-hot cores) — brighter + cooler than any stone/marble
	fusionCyanAnchor  = color.RGBA{R: 0x6c, G: 0xe8, B: 0xf0, A: 0xff} // bright pale cyan glow (fusion accent rings/sheen) — softer + paler than the harsh neon cyan
	metalSilverAnchor = color.RGBA{R: 0xc4, G: 0xc9, B: 0xd0, A: 0xff} // pale cool silver metal (space domes + launchpads) — a colder metallic sheen, faintly blue

	// COSMIC epoch (interstellar / galactic). Two tones for the deep-space step off the pale space colony:
	voidAnchor   = color.RGBA{R: 0x14, G: 0x18, B: 0x2a, A: 0xff} // deep starfield blue-black (interstellar ground — a cold void deck, far darker than the space colony's pale plaza)
	energyAnchor = color.RGBA{R: 0x74, G: 0xe0, B: 0xff, A: 0xff} // bright cyan-white orbital energy (galactic ring-hub core + ring glow) — an energetic megastructure glow, blended never raw

	// COSMIC epoch, SECOND pair (quantum / transcendent — the FINAL two ages, completing all 22). Same
	// discipline as every anchor above — NEVER used raw; blended against theme roles + neighbouring tones so
	// a light/dark theme retints them. QUANTUM is an IRIDESCENT crystalline lattice: three shifting facet
	// hues (cyan / magenta / gold) cycled by position/parity so the crystal city SHIMMERS rather than sits
	// one color — on a dark void-derived deck. TRANSCENDENT is ETHEREAL LIGHT: a luminous white + a soft warm
	// gold, blended translucent so buildings read as dematerialised light-forms on a pale glowing ground —
	// the brightest age, the post-physical finale.
	iridCyanAnchor    = color.RGBA{R: 0x46, G: 0xe6, B: 0xe0, A: 0xff} // iridescent facet cyan (quantum crystal — one of three shifting sheen hues)
	iridMagentaAnchor = color.RGBA{R: 0xd6, G: 0x56, B: 0xe4, A: 0xff} // iridescent facet magenta (quantum crystal — the second shifting sheen hue)
	iridGoldAnchor    = color.RGBA{R: 0xf0, G: 0xc8, B: 0x54, A: 0xff} // iridescent facet gold (quantum crystal — the third shifting sheen hue, warms the cyan/magenta cycle)
	etherWhiteAnchor  = color.RGBA{R: 0xf6, G: 0xf8, B: 0xff, A: 0xff} // luminous ethereal white (transcendent light-forms + ground bloom) — the brightest tone in the whole atlas, faintly cool
	etherGoldAnchor   = color.RGBA{R: 0xf4, G: 0xe2, B: 0xb0, A: 0xff} // soft warm halo gold (transcendent ascension rings + a warm cast on the white) — a gentle radiant warmth, never a saturated yellow
)

// iridHueFor cycles the three quantum iridescence anchors by an integer index (a ring number, a facet
// number, or an (x+y) position sum), so a crystalline surface SHIFTS hue across itself — the iridescent
// sheen. Pure helper (no locks); the anchor it returns is always BLENDED against roof/theme tones by the
// caller, never stamped raw.
func iridHueFor(i int) color.RGBA {
	switch ((i % 3) + 3) % 3 {
	case 0:
		return iridCyanAnchor
	case 1:
		return iridMagentaAnchor
	default:
		return iridGoldAnchor
	}
}

// tdStyleForEra returns the tuned preset for an era band, or defaultTdStyle for the bands not yet
// specialised. Tuned so far: ORGANIC (primitive/stone — V3-A), ANCIENT (bronze/iron/classical —
// V3-B) and MEDIEVAL (medieval/renaissance — V3-B). The remaining bands (industrial+ / modern /
// cyber / space) render a legible default city until V3-C/D tunes them.
func tdStyleForEra(e era) tdEraStyle {
	switch e {
	case eraOrganic:
		return organicVillageStyle
	case eraHubSpoke:
		return ancientCityStyle // V3-B: clay roofs, packed-earth ground, mudbrick walls+gates
	case eraCastle:
		return medievalCityStyle // V3-B: slate roofs, cobble ground, stone walls+towers+gatehouse
	default:
		return defaultTdStyle
	}
}

// ageStyles maps each age key to its citymap style preset. Per-age keying (replacing the old
// 7-band era grouping for STYLE selection) lets each age's look diverge independently in a later
// phase. Today every age maps to the exact same preset the old era-band lookup produced, so the
// render is unchanged. The era enum still drives town-form weights, flourishes, and tests — only
// the style pick moved to per-age keying.
var ageStyles = map[string]tdEraStyle{
	// organic — primitive/stone (V3-A; stone split off with its own rockier ground + megalith, 1b-i)
	"primitive_age": organicVillageStyle,
	"stone_age":     stoneAgeStyle,
	// ancient — bronze/iron/classical (V3-B; iron + classical split off with their own looks, 1b-ii)
	"bronze_age":    ancientCityStyle,
	"iron_age":      ironCityStyle,
	"classical_age": classicalCityStyle,
	// medieval — medieval/renaissance (V3-B; renaissance split off with cream stone + dome + star-fort)
	"medieval_age":    medievalCityStyle,
	"renaissance_age": renaissanceCityStyle,
	// default — every not-yet-tuned age renders the legible default city
	"colonial_age":   colonialCityStyle,
	"industrial_age": industrialCityStyle,
	// ELECTRIC epoch — victorian/electric/atomic (V3-B; distinct brownstone / art-deco / midcentury looks)
	"victorian_age":   victorianCityStyle,
	"electric_age":    electricCityStyle,
	"atomic_age":      atomicCityStyle,
	"modern_age":      modernCityStyle,
	"information_age": informationCityStyle,
	"digital_age":     digitalCityStyle,
	// NEON epoch — cyberpunk/fusion/space (distinct dark-neon / bright-white / pale-metallic looks)
	"cyberpunk_age": cyberpunkCityStyle,
	"fusion_age":    fusionCityStyle,
	"space_age":     spaceCityStyle,
	// COSMIC epoch — interstellar/galactic (distinct deep-space SPIRES / RING-HUB megastation looks)
	"interstellar_age": interstellarCityStyle,
	"galactic_age":     galacticCityStyle,
	// COSMIC epoch, FINAL pair — quantum (iridescent crystalline lattice) + transcendent (ethereal light +
	// ascension). This completes ALL 22 ages: no age maps to defaultTdStyle any more — it survives only as the
	// base preset other presets build from + the styleForAge fallback base.
	"quantum_age":      quantumCityStyle,
	"transcendent_age": transcendentCityStyle,
}

// styleForAge returns the citymap style preset for an age key, falling back to the organic village
// style for unknown ages (matching the old eraForAge default, which returned eraOrganic for age
// keys not found in the canonical order).
func styleForAge(ageKey string) tdEraStyle {
	if s, ok := ageStyles[ageKey]; ok {
		return s
	}
	return organicVillageStyle
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
	tdPlaza        // paved-stone town-square ground (drawn under the wonder/center roof)
	tdPropWell     // a stone well head (ring + dark shaft)
	tdPropFirepit  // a firepit / hearth (dark ring + ember center)
	tdPropStones   // standing stones / a totem (a couple of upright dabs)
	tdPropStall    // a market stall (awning patch)
	tdPropMegalith // stone-age: a single tall standing stone (a vertical grey slab + ground shadow)

	// V3-B era square props. Ancient set: altar / columns / braziers (+ well). Medieval set:
	// market stalls / fountain / cross-or-gallows (+ well). Each has its own small top-down draw
	// routine (drawSquareProp), so per-era squares read distinct without disturbing the primitive
	// set. Placed by tdRingProps exactly like the primitive props (deterministic ring, no overlap).
	tdPropAltar      // ancient: a low stone altar (a flat slab + a small offering dab)
	tdPropColumns    // ancient: a row of columns / colonnade (a few upright pale dabs)
	tdPropBrazier    // ancient: a fire brazier on a stand (a bright ember over a dark base)
	tdPropFountain   // medieval: a stone fountain (a paved ring + a water center)
	tdPropCross      // medieval: a market cross / gallows (an upright post with a crossbar)
	tdPropSmokestack // industrial: a tall dark factory chimney + a soot dab on top (taller than other props)
	tdPropGasLamp    // victorian: a short lamp-post with a warm amber gaslight glow (a genteel park lamp)
	tdPropDataCenter // information: a low WIDE server-farm block (flat cool grey) + a couple of cyan/amber blinking-light dabs
	tdPropNeonSign   // digital: a small bright NEON sign dab (cyan/magenta) — the epoch's first neon
	tdPropHologram   // cyberpunk: a bright TRANSLUCENT floating projection — a half-transparent cyan/magenta glowing shape blended over whatever's beneath (a hologram, not a solid)
	tdPropRocket     // space: a small ROCKET / gantry dab — a bright vertical capsule with a nose + a lit rim beside a thin gantry, seasoning the spaceport
	tdPropLightMote  // transcendent: a soft floating glowing MOTE — a small translucent warm-white bloom hovering over the light-field (an ethereal spark, blended not solid), seasoning the finale square
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
	// shape is the town's footprint SILHOUETTE (disc / organic blob) picked per AGE (tdAgeFootprint),
	// distinct from form (which arranges the wards INSIDE the outline). It selects the OUTLINE the
	// raster clips to and the wall ring follows. Kept on the plan so the wall code (tdAddWalls) reads
	// the same silhouette the block field was clipped to; future per-age silhouettes plug in here.
	shape tdFootprintShape
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
//	formCrescent   — garden-suburb CURVED streets: seeds strung along several CONCENTRIC ARCS at a
//	    lower density with deliberate green GAPS (garden squares). Long sweeping terraces, not a wheel
//	    and not a full ring. The Victorian garden suburb.
//	formBoulevard  — an orthogonal GRID CUT by 2-4 grand DIAGONAL avenues: seeds laid on a lattice
//	    with a band cleared along each diagonal so the Voronoi streets include long diagonal
//	    boulevards slicing the blocks. The Haussmann/electric-city plan.
//	formCoreSuburb — a dense downtown CORE ringed by sparse loopy SUBURBS: most seeds packed tightly
//	    inside an inner radius, the rest spread thin (large wards) in the outer ring. The atomic-age
//	    metropolis-with-sprawl.
type tdTownForm int

const (
	formOrganic tdTownForm = iota
	formRadial
	formGrid
	formRibbon
	formCrescent
	formBoulevard
	formCoreSuburb
	// tdFormCount is a sentinel = one past the last real form; loop [0,tdFormCount) to iterate
	// forms in a FIXED, deterministic order (never `range` a tdFormWeights map).
	tdFormCount
)

// tdFormWeights is a per-AGE weighting over the town forms (form → relative weight), consumed as a
// discrete distribution by tdPickTownForm. A form ABSENT from the map (or with a ≤0 weight) is
// forbidden for that age (e.g. primitive villages are NEVER radial or grid — they ramble, they are
// not planned). Modelled as a MAP (not a fixed vector) so bespoke forms — crescent, boulevard,
// core+suburb, and whatever lands next — can be added without every table entry growing a slot.
// These are deliberately TUNABLE; the only hard contract the tests lock is that PRIMITIVE is
// organic-dominant and never a wheel/grid (it never lists radial or grid). Consumers MUST iterate
// forms in a FIXED order (loop the enum 0..N and look each up) — ranging a Go map is randomized and
// would break the determinism tdPickTownForm guarantees.
type tdFormWeights map[tdTownForm]float64

// tdAgeForms is the PER-AGE form distribution table (map-overhaul-citymap Phase 2a). Each of the 22
// ages gets its OWN characteristic town form (the dominant weight), replacing the coarse per-ERA
// band table so a city's whole gestalt reads distinctly age to age while still fanning out a little
// across citySeeds. Each entry maps form→relative weight (absent form = forbidden); weights are
// relative (they need not sum to 1). Bespoke new geometries (crescent/boulevard/coreSuburb) land on
// the ELECTRIC epoch here; every other age keeps its original weights unchanged. LOCKED INVARIANT:
// primitive_age lists neither radial nor grid — villages never plan a wheel or a survey grid.
var tdAgeForms = map[string]tdFormWeights{
	"primitive_age":    {formOrganic: 0.85, formRibbon: 0.15},                   // organic (invariant: no wheel/grid)
	"stone_age":        {formOrganic: 0.85, formRibbon: 0.15},                   // organic
	"bronze_age":       {formOrganic: 0.30, formRadial: 0.60, formRibbon: 0.10}, // radial — first monument/forum cores
	"iron_age":         {formOrganic: 0.30, formRadial: 0.60, formRibbon: 0.10}, // radial
	"classical_age":    {formOrganic: 0.10, formRadial: 0.25, formGrid: 0.60, formRibbon: 0.05},
	"medieval_age":     {formOrganic: 0.45, formRadial: 0.40, formGrid: 0.05, formRibbon: 0.10},
	"renaissance_age":  {formOrganic: 0.15, formRadial: 0.65, formGrid: 0.10, formRibbon: 0.10},
	"colonial_age":     {formOrganic: 0.15, formRadial: 0.05, formGrid: 0.70, formRibbon: 0.10},
	"industrial_age":   {formOrganic: 0.10, formGrid: 0.75, formRibbon: 0.15},
	"victorian_age":    {formCrescent: 0.85, formOrganic: 0.15},                                 // crescent — garden-suburb curves
	"electric_age":     {formBoulevard: 0.85, formGrid: 0.15},                                   // boulevard — grid cut by grand diagonals
	"atomic_age":       {formCoreSuburb: 0.85, formGrid: 0.15},                                  // core+suburb — dense downtown, loose sprawl
	"modern_age":       {formOrganic: 0.05, formRadial: 0.05, formGrid: 0.85, formRibbon: 0.05}, // grid — the planned metropolis
	"information_age":  {formOrganic: 0.30, formRadial: 0.05, formGrid: 0.60, formRibbon: 0.05},
	"digital_age":      {formOrganic: 0.05, formRadial: 0.05, formGrid: 0.85, formRibbon: 0.05},
	"cyberpunk_age":    {formOrganic: 0.05, formRadial: 0.05, formGrid: 0.85, formRibbon: 0.05},
	"fusion_age":       {formOrganic: 0.20, formRadial: 0.35, formGrid: 0.40, formRibbon: 0.05},
	"space_age":        {formOrganic: 0.15, formRadial: 0.10, formGrid: 0.70, formRibbon: 0.05},
	"interstellar_age": {formOrganic: 0.55, formRadial: 0.10, formGrid: 0.30, formRibbon: 0.05},
	"galactic_age":     {formOrganic: 0.15, formRadial: 0.70, formGrid: 0.10, formRibbon: 0.05}, // radial — the galactic wheel
	"quantum_age":      {formOrganic: 0.10, formRadial: 0.10, formGrid: 0.75, formRibbon: 0.05},
	"transcendent_age": {formOrganic: 0.20, formRadial: 0.65, formGrid: 0.05, formRibbon: 0.10},
}

// tdAgeFormWeights returns the form distribution for an age key (map-overhaul-citymap Phase 2a).
// Unknown/empty keys fall back to an organic-dominant blend so a mis-keyed age still renders a
// sensible rambling town rather than a degenerate one.
func tdAgeFormWeights(ageKey string) tdFormWeights {
	if w, ok := tdAgeForms[ageKey]; ok {
		return w
	}
	return tdFormWeights{formOrganic: 0.80, formRibbon: 0.20}
}

// tdPickTownForm chooses a town's FORM deterministically from (citySeed, ageKey). It is a pure
// function — the SAME (seed, ageKey) always yields the SAME form — and it is AGE-WEIGHTED via
// tdAgeFormWeights, so different citySeeds fan out across the age-appropriate forms (no two towns
// need look alike) while each age reads characteristically (and PRIMITIVE reliably lands organic,
// never a wheel). A degenerate all-zero or negative weight vector falls back to organic. The seed
// is hashed with a distinct salt so the form roll is independent of the seed's other uses (anchor
// phase, jitter, scatter phase).
func tdPickTownForm(seed uint32, ageKey string) tdTownForm {
	w := tdAgeFormWeights(ageKey)
	// CRITICAL: iterate forms in a FIXED enum order (0..tdFormCount), looking each up in the weight
	// map — NEVER `range` the map, whose order Go randomizes (that would break determinism).
	total := 0.0
	for f := formOrganic; f < tdFormCount; f++ {
		if x := w[f]; x > 0 {
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
	for f := formOrganic; f < tdFormCount; f++ {
		x := w[f]
		if x <= 0 {
			continue
		}
		acc += x
		if roll < acc {
			return f
		}
	}
	// Float slop guard: return the last positive-weight form (walk the enum in reverse).
	for f := tdFormCount - 1; f >= formOrganic; f-- {
		if w[f] > 0 {
			return f
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

// tdFootprintShape is the PER-AGE town-silhouette family (map-overhaul-citymap Phase 2b). Distinct
// from tdTownForm (which selects the block-seed SCATTER strategy — how wards are arranged INSIDE the
// footprint): the shape selects the OUTLINE the raster clips to. Today only two exist — the plain
// disc and the ragged organic BLOB (the wobble that gives villages a rambling edge). More silhouettes
// (sprawl, rounded-rect, ring, …) arrive in the next slice; this type is the seam they plug into.
type tdFootprintShape int

const (
	shapeDisc      tdFootprintShape = iota // a clean circle of radius townR (every non-village age today)
	shapeBlob                              // the seeded organic wobble (tdOrganicRadiusAt) — village edges
	shapeSprawl                            // wide elongated oval + gentle lobes — victorian garden-suburb sprawl
	shapeRoundRect                         // slightly-elongated rounded rectangle (squircle) — a planned grid city
	shapeCoreHalo                          // broad bumpy blob — a metropolis with irregular sprawling suburbs
)

// tdAgeFootprints maps an age key to its footprint SILHOUETTE (map-overhaul-citymap Phase 2b). Only
// the two village ages get the ragged blob; every other age is a disc. Ages absent from the map (and
// the empty/unknown key) default to shapeDisc via tdAgeFootprint.
//
// Note: today the blob wobble was applied to ANY city that ROLLED formOrganic (regardless of age).
// Keying it to the two village ages is the intended per-age behavior — the only visible change is the
// rare NON-village city that happened to roll organic no longer wobbles (acceptable; it now reads as
// the clean disc its age otherwise plans).
var tdAgeFootprints = map[string]tdFootprintShape{
	"primitive_age": shapeBlob,
	"stone_age":     shapeBlob,
	"victorian_age": shapeSprawl,
	"electric_age":  shapeRoundRect,
	"atomic_age":    shapeCoreHalo,
}

// tdAgeFootprint returns the footprint shape for an age key, defaulting to shapeDisc for any age not
// in tdAgeFootprints (including the empty/unknown key). Pure.
func tdAgeFootprint(ageKey string) tdFootprintShape {
	if s, ok := tdAgeFootprints[ageKey]; ok {
		return s
	}
	return shapeDisc
}

// tdShapeRadiusAt is the town OUTLINE radius at a given angle for a footprint shape (city units): the
// single seam every silhouette funnels through. shapeDisc is the constant townR (a clean circle);
// shapeBlob defers to the existing organic wobble. Shared by the raster clip and the wall ring so the
// footprint is consistent everywhere. Pure + bounded (≤ townR).
func tdShapeRadiusAt(shape tdFootprintShape, angle, townR float64, seed uint32) float64 {
	switch shape {
	case shapeBlob:
		return tdOrganicRadiusAt(angle, townR, seed)
	case shapeSprawl:
		// A WIDE elongated oval (garden-suburb sprawl) with gentle irregular lobes. The long axis is
		// seeded so towns sprawl in different directions; a:b ≈ 1.8:1, peak reach ~1.5·townR wide.
		axis := float64(hash2(0x5B1A, 0x01, seed)) / float64(^uint32(0)) * math.Pi
		phi := angle - axis
		const a, b = 1.34, 0.74
		e := (a * b) / math.Hypot(b*math.Cos(phi), a*math.Sin(phi))
		ph2 := float64(hash2(0x5B1A, 0x02, seed)) / float64(^uint32(0)) * 2 * math.Pi
		ph3 := float64(hash2(0x5B1A, 0x03, seed)) / float64(^uint32(0)) * 2 * math.Pi
		lobe := 1 + 0.10*math.Sin(2*angle+ph2) + 0.06*math.Sin(3*angle+ph3)
		return townR * e * lobe
	case shapeRoundRect:
		// A slightly-elongated SUPERELLIPSE (squircle) → a rounded-rectangle outline for a planned city.
		axis := float64(hash2(0x5B1B, 0x01, seed)) / float64(^uint32(0)) * math.Pi
		phi := angle - axis
		const n = 4.0
		cx := math.Cos(phi) / 1.15
		cy := math.Sin(phi) / 0.9
		denom := math.Pow(math.Pow(math.Abs(cx), n)+math.Pow(math.Abs(cy), n), 1.0/n)
		if denom < 1e-6 {
			denom = 1e-6
		}
		return townR / denom
	case shapeCoreHalo:
		// A broad BUMPY blob — a metropolis whose suburbs bulge irregularly past a disc.
		ph2 := float64(hash2(0x5B1C, 0x02, seed)) / float64(^uint32(0)) * 2 * math.Pi
		ph3 := float64(hash2(0x5B1C, 0x03, seed)) / float64(^uint32(0)) * 2 * math.Pi
		ph5 := float64(hash2(0x5B1C, 0x05, seed)) / float64(^uint32(0)) * 2 * math.Pi
		lobe := 1 + 0.16*math.Sin(2*angle+ph2) + 0.11*math.Sin(3*angle+ph3) + 0.07*math.Sin(5*angle+ph5)
		if r := townR * 1.08 * lobe; r > 0.7*townR {
			return r
		}
		return 0.7 * townR
	default: // shapeDisc
		return townR
	}
}

// tdShapeMaxReach is a safe upper bound on tdShapeRadiusAt over all angles/seeds, as a multiple of
// townR. Used to bound the footprint — walls, frame fit, and the compactness test — for shapes that
// legitimately extend past the townR disc (sprawl/rect/halo). Disc + blob never exceed townR (the
// blob only bites inward), so they return 1.
func tdShapeMaxReach(shape tdFootprintShape) float64 {
	switch shape {
	case shapeSprawl:
		return 1.55 // a·(1+lobe) = 1.34·1.16
	case shapeRoundRect:
		return 1.25 // squircle corners under the seeded elongation
	case shapeCoreHalo:
		return 1.45 // 1.08·(1+0.16+0.11+0.07)
	default: // shapeDisc, shapeBlob (inward-only)
		return 1.0
	}
}

// tdShapeAreaFactor is the footprint AREA as a multiple of the unit-townR disc (πtownR²), i.e. the
// mean of (tdShapeRadiusAt/townR)² over all angles. Grid/lattice scatters use it to enlarge the cell
// size so a FIXED `need` seeds spread over the larger silhouette (a disc-density lattice would leave
// a >disc footprint's outer regions sparse). shapeDisc and shapeBlob are ~a disc in area, and sprawl
// is an ellipse of near-disc area, so all three return 1.0 (disc reduction is then exact). rect/halo
// bulge past the disc and return their measured factor. Pure.
func tdShapeAreaFactor(shape tdFootprintShape) float64 {
	switch shape {
	case shapeRoundRect:
		return 1.22
	case shapeCoreHalo:
		return 1.19
	default: // shapeDisc, shapeBlob, shapeSprawl (all ~disc area)
		return 1.0
	}
}

// tdShapeAcceptR is the per-angle placement-acceptance radius for a scatter candidate at (x,y):
// the footprint radius at that point's angle (tdShapeRadiusAt) scaled by a strategy's own edge
// INSET (e.g. 0.92 → a ground rim of ground/greenery hugs the outline). A candidate is inside iff
// hypot(x,y) ≤ this. For shapeDisc, tdShapeRadiusAt returns townR, so this reduces EXACTLY to
// inset·townR — every scatter strategy's disc behaviour is byte-identical. For sprawl/rect/halo it
// tracks the real silhouette, so seeds populate the wide/corner regions instead of stopping at a
// disc that the raster later stretches over. Pure + bounded.
func tdShapeAcceptR(shape tdFootprintShape, x, y, townR, inset float64, seed uint32) float64 {
	return inset * tdShapeRadiusAt(shape, math.Atan2(y, x), townR, seed)
}

// tdSampleInShape draws ONE point uniform-by-area inside the footprint (inset·shape outline): it
// samples uniform in the bounding disc of radius maxR and rejects candidates outside the per-angle
// footprint radius, with a bounded reject budget then a relaxed fallback so a point is ALWAYS
// returned (callers rely on this to hit an exact seed count). It advances the caller's RNG, so it is
// deterministic. For shapeDisc the accept radius is the constant inset·townR = maxR, so the FIRST
// candidate always passes and this is exactly `rad = maxR·√u, ang = 2πu` — the original disc top-up.
func tdSampleInShape(r *rng, shape tdFootprintShape, townR, inset, maxR float64, seed uint32) tdPoint {
	for tries := 0; tries < 24; tries++ {
		rad := maxR * math.Sqrt(r.f01())
		ang := r.f01() * 2 * math.Pi
		x, y := math.Cos(ang)*rad, math.Sin(ang)*rad
		if edge := tdShapeAcceptR(shape, x, y, townR, inset, seed); rad <= edge {
			return tdPoint{x, y}
		}
	}
	// Relaxed fallback: draw once more and clamp onto the outline so we always return an in-shape
	// point (keeps the caller's count exact even under an unlucky reject streak).
	rad := maxR * math.Sqrt(r.f01())
	ang := r.f01() * 2 * math.Pi
	x, y := math.Cos(ang)*rad, math.Sin(ang)*rad
	if edge := tdShapeAcceptR(shape, x, y, townR, inset, seed); rad > edge && rad > 0 {
		s := edge / rad
		x *= s
		y *= s
	}
	return tdPoint{x, y}
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

// tdInTown reports whether a city-space point lies inside the town footprint for a given SHAPE.
// shapeBlob uses the irregular blob outline (tdOrganicRadiusAt); shapeDisc uses the plain circular
// disc of radius townR (unchanged). Shared by the raster partition and Lloyd relaxation so the town
// shape is consistent everywhere. Pure.
func tdInTown(x, y, townR float64, shape tdFootprintShape, seed uint32) bool {
	d2 := x*x + y*y
	rr := tdShapeRadiusAt(shape, math.Atan2(y, x), townR, seed)
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
	// (seed, roof count, form). The FORM is picked once, deterministically + AGE-weighted, so towns
	// vary per city+age and primitive villages ramble organically instead of all reading as wheels.
	plan.form = tdPickTownForm(seed, state.Age)
	// The footprint SILHOUETTE is keyed to the AGE (villages ramble; every other age is a clean disc),
	// independently of the ward-scatter FORM above. Kept on the plan so the wall ring follows the same
	// outline the block field is clipped to.
	plan.shape = tdAgeFootprint(state.Age)
	totalRoofs := tdTotalFabricRoofs(blds)
	plan.townR = tdTownRadius(totalRoofs, cfg)
	field := tdBuildBlockField(plan.townR, plan.anchors, totalRoofs, plan.form, plan.shape, cfg, seed)
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
	tdPlaceWonders(&plan, style, cfg)
	tdEnforceMinGap(&plan, cfg)

	// (e) town squares — dress each plaza region (wonders + the wonderless center) as a paved
	// town square with a few seeded era props ringed around the roof, so the open center reads
	// as intentional rather than a bare-dirt donut.
	tdPlaceSquares(&plan, style, cfg, seed)

	// (f) filler — balanced gardens / ponds / trees / props in the leftover in-town space.
	tdAddFiller(&plan, field, style, cfg, seed)

	// (g) walls — a wall+gate ring IF the era has walls (ancient mudbrick, medieval stone+towers).
	// Primitive + industrial-and-later: none (open sprawl).
	if style.hasWalls {
		tdAddWalls(&plan, style, seed)
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
func tdScatterSeedsFor(form tdTownForm, townR float64, anchors []tdAnchor, B int, shape tdFootprintShape, cfg tdConfig, seed uint32) []tdPoint {
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
		return tdScatterRadial(seeds, townR, need, shape, cfg, seed)
	case formGrid:
		return tdScatterGrid(seeds, townR, need, shape, cfg, seed)
	case formRibbon:
		return tdScatterRibbon(seeds, townR, need, shape, cfg, seed)
	case formCrescent:
		return tdScatterCrescent(seeds, townR, need, shape, cfg, seed)
	case formBoulevard:
		return tdScatterBoulevard(seeds, townR, need, shape, cfg, seed)
	case formCoreSuburb:
		return tdScatterCoreSuburb(seeds, townR, need, shape, cfg, seed)
	default: // formOrganic
		return tdScatterOrganic(seeds, townR, need, shape, cfg, seed)
	}
}

// tdScatterRadial is the ORIGINAL golden-angle phyllotaxis scatter — now the formRadial strategy
// ONLY (map-overhaul-citymap). Free seeds fill the disc on a golden-angle spiral: this is exactly
// what puts the Voronoi ward BOUNDARIES on radial spokes and, with the pinned-center region acting
// as a hub, reads as a wagon wheel with a ring road. That look is intentional for the radial form
// (a monument/forum-planned town) but is no longer the default for every town. `seeds` already
// holds the pinned anchors; this appends the free seeds. Pure over (seed, need).
func tdScatterRadial(seeds []tdPoint, townR float64, need int, shape tdFootprintShape, cfg tdConfig, seed uint32) []tdPoint {
	phase := float64(hash2(0x5eed, 0x1a, seed)) / float64(^uint32(0)) * 2 * math.Pi
	// The spiral fills up to ~0.92 of the footprint radius AT EACH ANGLE so the outermost blocks sit
	// inside the rim (a ground/greenery ring hugs the silhouette). For shapeDisc the per-angle radius
	// is the constant townR, so this reduces to 0.92·townR·√frac exactly.
	for k := 0; k < need; k++ {
		frac := (float64(k) + 0.5) / float64(maxInt(need, 1))
		ang := float64(len(seeds))*goldenAngle + phase
		reach := tdShapeAcceptR(shape, math.Cos(ang), math.Sin(ang), townR, 0.92, seed)
		r := reach * math.Sqrt(frac)
		// Nudge the innermost few outward so they don't collide with the pinned center anchor.
		if r < cfg.townBaseRadius*0.35 {
			r = cfg.townBaseRadius*0.35 + r*0.5
		}
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
func tdScatterOrganic(seeds []tdPoint, townR float64, need int, shape tdFootprintShape, cfg tdConfig, seed uint32) []tdPoint {
	nPinned := len(seeds) // the anchor seeds already appended by tdScatterSeedsFor
	// Sampling extent covers the whole silhouette (maxReach·townR); free seeds then get clamped to the
	// per-angle footprint radius. For shapeBlob/shapeDisc maxReach is 1 → rr = 0.94·townR as before.
	rr := 0.94 * townR * tdShapeMaxReach(shape)
	// Target spacing from area: minDist ≈ 0.7·√(area / seeds). The 0.7 leaves the reject pass room to
	// actually place the target count before it must relax.
	area := math.Pi * rr * rr
	minDist := 0.7 * math.Sqrt(area/float64(need+len(seeds)))
	minD2 := minDist * minDist
	r := newRNG(hash2(0x0B10, uint32(need), seed) | 1)
	// sample draws one uniform-in-extent candidate: uniform by area over the bounding disc, then
	// clamped to the SHAPE outline at that angle so a candidate never lands past the rambling edge.
	// For shapeBlob the outline is tdOrganicRadiusAt (0.94 inset) exactly as before; for shapeDisc the
	// clamp radius is the constant 0.94·townR (never trips, since rad ≤ that) → the plain disc fill.
	sample := func() tdPoint {
		rad := rr * math.Sqrt(r.f01())
		ang := r.f01() * 2 * math.Pi
		x, y := math.Cos(ang)*rad, math.Sin(ang)*rad
		if edge := tdShapeAcceptR(shape, x, y, townR, 0.94, seed); rad > edge {
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
func tdScatterGrid(seeds []tdPoint, townR float64, need int, shape tdFootprintShape, cfg tdConfig, seed uint32) []tdPoint {
	rr := 0.92 * townR
	// Choose a cell size so ~need nodes fall inside the FOOTPRINT: footprintArea ≈ need·cell² → cell ≈
	// √(π·rr²·areaFactor / need). areaFactor enlarges the cell for shapes bigger than a disc (rect/
	// halo) so a fixed `need` seeds spread across the whole silhouette instead of packing the centre.
	// A mild 1.08 loosening keeps a few spare nodes for the centers-out trim. For shapeDisc areaFactor
	// is 1 → identical cell.
	cell := 1.08 * math.Sqrt(math.Pi*rr*rr*tdShapeAreaFactor(shape)/float64(need))
	if cell < 1e-3 {
		cell = 1e-3
	}
	// A seeded sub-cell phase offset so the grid origin (and thus the whole lattice) shifts per
	// city; kept within one cell so the lattice still spans the footprint.
	phx := (float64(hash2(0x671D, 0x11, seed))/float64(^uint32(0)) - 0.5) * cell
	phy := (float64(hash2(0x671D, 0x12, seed))/float64(^uint32(0)) - 0.5) * cell
	// Enumerate lattice nodes across the bounding square (sized to the shape's max reach so corners/
	// wide regions get nodes), keep the ones INSIDE THE SHAPE with jitter, and order them by distance
	// from the core so a trim to `need` keeps the central lattice.
	maxR := rr * tdShapeMaxReach(shape) // outer bound of the lattice box (= rr for a disc)
	type gnode struct {
		p  tdPoint
		d2 float64
	}
	var nodes []gnode
	half := int(math.Ceil(maxR/cell)) + 1
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
			if edge := tdShapeAcceptR(shape, px, py, townR, 0.92, seed); d2 > edge*edge {
				continue
			}
			nodes = append(nodes, gnode{p: tdPoint{px, py}, d2: d2})
		}
	}
	sort.SliceStable(nodes, func(a, b int) bool { return nodes[a].d2 < nodes[b].d2 })
	for i := 0; i < need && i < len(nodes); i++ {
		seeds = append(seeds, nodes[i].p)
	}
	// Top up (rare: only if the footprint couldn't host `need` in-shape nodes) so the free-seed count
	// stays exactly `need` and banded stability is preserved. Sample in the bounding disc, reject
	// outside the shape (a bounded budget, then a relaxed accept, so the count is always met).
	for placed := minInt(need, len(nodes)); placed < need; placed++ {
		seeds = append(seeds, tdSampleInShape(r, shape, townR, 0.92, maxR, seed))
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
func tdScatterRibbon(seeds []tdPoint, townR float64, need int, shape tdFootprintShape, cfg tdConfig, seed uint32) []tdPoint {
	// Axis direction: a seeded angle, so different ribbon towns run different ways.
	axis := float64(hash2(0x21B0, 0x33, seed)) / float64(^uint32(0)) * 2 * math.Pi
	ax, ay := math.Cos(axis), math.Sin(axis)
	// Perpendicular (lateral) direction.
	px, py := -ay, ax
	// Reach along the road follows the footprint radius IN THE AXIS DIRECTION (so a sprawl town's
	// ribbon runs the full length of its long axis); lateral spread scales the same way. For shapeDisc
	// the axis radius is townR → alongMax = 0.90·townR, latMax = 0.28·townR, exactly as before.
	axisReach := tdShapeRadiusAt(shape, axis, townR, seed)
	alongMax := 0.90 * axisReach // reach most of the footprint along the road
	latMax := 0.28 * axisReach   // a modest lateral spread → a road, not a blob
	r := newRNG(hash2(0x21B1, uint32(need), seed) | 1)
	for k := 0; k < need; k++ {
		// Even, centers-out spacing along the axis in [-alongMax, alongMax], with a small wobble.
		frac := (float64(k)+0.5)/float64(need)*2 - 1 // -1..1
		wobble := (r.f01() - 0.5) * (2 * alongMax / float64(need)) * 0.8
		along := frac*alongMax + wobble
		lat := (r.f01()*2 - 1) * latMax
		x := ax*along + px*lat
		y := ay*along + py*lat
		// Clamp inside the footprint at this point's angle so the raster partition stays whole (streets
		// connected). For shapeDisc the bound is the constant 0.95·townR → identical to the old clamp.
		if edge := tdShapeAcceptR(shape, x, y, townR, 0.95, seed); math.Hypot(x, y) > edge {
			d := math.Hypot(x, y)
			s := edge / d
			x *= s
			y *= s
		}
		seeds = append(seeds, tdPoint{x, y})
	}
	return seeds
}

// tdScatterCrescent is the GARDEN-SUBURB curved-street scatter — the formCrescent strategy
// (map-overhaul-citymap Phase 2a). Free seeds are strung along several CONCENTRIC ARCS (crescents)
// at increasing radii: the Voronoi ward boundaries then follow those sweeping curves, so the town
// reads as long CURVED terraces — a genteel Victorian garden suburb — rather than a grid or a wheel.
// Two things make it a crescent and NOT a ring/wheel: (1) each arc spans only a partial sweep
// (~0.62 turn), never the full circle, and (2) each arc DELIBERATELY skips 2-3 segments, leaving
// seed-free GAPS that the raster paints as green GARDEN SQUARES. Effective density is a touch lower
// than the grid (the seeds spread along long curves), so wards come out as sweeping terraces with
// leafy gaps between them.
//
// Determinism + a fixed free-seed COUNT: exactly `need` free seeds are placed (banded stability is
// preserved). Seeds are apportioned across arcs proportionally to each arc's radius (outer crescents
// are longer, so they carry more terraces); the arc a given index lands on and its along-arc position
// are index-stable, with a small seeded angular wobble. `seeds` already holds the pinned anchors;
// this appends the free seeds inside ~0.92·townR so a ground rim shows. Pure over (seed, need);
// bounded → panic-safe.
func tdScatterCrescent(seeds []tdPoint, townR float64, need int, shape tdFootprintShape, cfg tdConfig, seed uint32) []tdPoint {
	rr := 0.92 * townR
	// The terrace STACK spans the footprint's max reach so a wide silhouette (sprawl) gets terraces
	// across its whole width; each seed is then clipped to the per-angle outline. For shapeDisc
	// maxReach is 1 → reachR = rr and every quantity below reduces to the original disc geometry.
	reachR := rr * tdShapeMaxReach(shape)
	if need <= 0 {
		return seeds
	}
	// GARDEN-SUBURB terraces. The crescents are nested arcs that share a FAR pivot OFF to one side of
	// town (in the seeded "grain" direction u), so inside the town they read as roughly-parallel SWEEPS
	// all curving the same way — like Nash terraces / a garden suburb — NOT concentric rings around the
	// plaza (which read radial). Each terrace's nearest approach to the centre is spread along the grain
	// so the terraces stack across the whole town.
	nArc := 2 + int(math.Sqrt(float64(need))/2.0)
	if nArc < 3 {
		nArc = 3
	}
	if nArc > 8 {
		nArc = 8
	}
	basePhase := float64(hash2(0xC2E5, 0x01, seed)) / float64(^uint32(0)) * 2 * math.Pi
	ux, uy := math.Cos(basePhase), math.Sin(basePhase) // grain direction; the pivot lies this way
	// A FAR pivot (several town-radii away) so each terrace's arc is nearly straight over the town —
	// the terraces become gently-curved, near-PARALLEL sweeps (like contour lines / Nash terraces),
	// not spokes converging on the centre. (A near pivot made terraces cross the centre → radial.)
	farDist := 4.6 * reachR
	cX, cY := ux*farDist, uy*farDist // shared far pivot of every terrace
	r := newRNG(hash2(0xC2E5, uint32(need), seed) | 1)
	// nearAt(a): where terrace a crosses the grain axis, spread from -0.74reachR..+0.74reachR through
	// centre (so the stack spans the whole silhouette width, not just a disc).
	nearAt := func(a int) float64 {
		if nArc <= 1 {
			return 0
		}
		t := float64(a)/float64(nArc-1)*2 - 1 // -1..1
		return t * 0.74 * reachR
	}
	// Weight middle terraces heavier (they cross more of the town → longer crescents host more homes).
	weights := make([]float64, nArc)
	wsum := 0.0
	for a := 0; a < nArc; a++ {
		w := 1.0 - 0.55*math.Abs(nearAt(a))/reachR
		if w < 0.2 {
			w = 0.2
		}
		weights[a] = w
		wsum += w
	}
	perArc := make([]int, nArc)
	placedTot := 0
	for a := 0; a < nArc; a++ {
		perArc[a] = int(float64(need) * weights[a] / wsum)
		placedTot += perArc[a]
	}
	for a := 0; placedTot < need; a = (a + 1) % nArc { // hand out the rounding remainder
		perArc[a]++
		placedTot++
	}
	baseAng := basePhase + math.Pi // from the far pivot back toward the town: the crescent's mid-sweep
	for a := 0; a < nArc; a++ {
		cnt := perArc[a]
		if cnt <= 0 {
			continue
		}
		// Radius so the arc's near point sits at nearAt(a) along the grain (|farDist-R| = |nearAt|).
		R := farDist - nearAt(a)
		// With a far pivot, a small angular span already covers the whole town width (arc length ≈
		// R·span). Keep it wide enough to reach both rims but not so wide the ends curl back.
		span := 0.46 + 0.22*r.f01()
		// One seeded garden-square GAP per terrace: a t-window kept seed-free (a leafy square).
		gap := -0.18 + 0.36*r.f01()
		const gapHalf = 0.11
		for k := 0; k < cnt; k++ {
			t := (float64(k)+0.5)/float64(cnt) - 0.5 // -0.5..0.5 along the sweep
			if math.Abs(t-gap) < gapHalf {           // nudge off the garden square
				if t < gap {
					t = gap - gapHalf
				} else {
					t = gap + gapHalf
				}
			}
			ang := baseAng + t*span
			ang += (r.f01() - 0.5) * (span / float64(cnt)) * 0.6 // gentle wobble, not a mechanical arc
			rw := R * (1 + (r.f01()-0.5)*0.03)
			x := cX + math.Cos(ang)*rw
			y := cY + math.Sin(ang)*rw
			// Clip stragglers to the footprint outline at their angle (constant rr for shapeDisc).
			if edge := tdShapeAcceptR(shape, x, y, townR, 0.92, seed); math.Hypot(x, y) > edge {
				d := math.Hypot(x, y)
				s := edge / d
				x *= s
				y *= s
			}
			seeds = append(seeds, tdPoint{x, y})
		}
	}
	return seeds
}

// tdScatterBoulevard is the GRID-CUT-BY-GRAND-DIAGONALS scatter — the formBoulevard strategy
// (map-overhaul-citymap Phase 2a). It starts from a regular in-disc GRID of seeds (like
// tdScatterGrid: orthogonal-ish wards) and then CARVES 2-4 straight DIAGONAL corridors across the
// town by removing every seed within a narrow band of each diagonal line. Those cleared bands become
// long DIAGONAL AVENUES in the Voronoi street web, slicing the orthogonal blocks at an angle — the
// Haussmann/electric-city boulevard plan. Removed seeds are TOPPED UP (pushed just aside, off the
// avenue) so the free-seed count stays exactly `need` and banded stability holds.
//
// The diagonals run through/near the centre at seeded angles in the ~30-60° band (never axis-aligned,
// so they read as diagonals cutting the grid, not just extra grid streets). Deterministic; the grid
// build is shared logic with tdScatterGrid. Pure over (seed, need); bounded → panic-safe.
func tdScatterBoulevard(seeds []tdPoint, townR float64, need int, shape tdFootprintShape, cfg tdConfig, seed uint32) []tdPoint {
	rr := 0.92 * townR
	reachR := rr * tdShapeMaxReach(shape) // outer extent of the lattice + avenue offsets (= rr for a disc)
	// Same cell sizing as the grid form so the orthogonal fabric matches a plain grid town; areaFactor
	// enlarges the cell for >disc shapes so `need` seeds spread across the whole silhouette.
	cell := 1.08 * math.Sqrt(math.Pi*rr*rr*tdShapeAreaFactor(shape)/float64(need))
	if cell < 1e-3 {
		cell = 1e-3
	}
	phx := (float64(hash2(0xB0DE, 0x11, seed))/float64(^uint32(0)) - 0.5) * cell
	phy := (float64(hash2(0xB0DE, 0x12, seed))/float64(^uint32(0)) - 0.5) * cell
	// 2-4 grand diagonals, each a line through a point near centre at a seeded angle in the 30-60°
	// band (mirrored to the other diagonal quadrant too). A seed within `avenueHalf` of ANY diagonal
	// line is removed to clear the avenue.
	nDiag := 3 + int(hash2(0xB0DE, 0x20, seed)%2) // 3..4 — always a real boulevard system
	type diagLine struct {
		nx, ny float64 // unit normal of the line
		c      float64 // line: nx*x + ny*y = c  (offset from origin)
	}
	diags := make([]diagLine, 0, nDiag)
	for d := 0; d < nDiag; d++ {
		// Angle in [30°,60°], alternating sign so avenues cross like an X, plus a seeded jitter.
		base := math.Pi/6 + (math.Pi/6)*float64(hash2(0xB0DE, uint32(0x30+d), seed))/float64(^uint32(0))
		if d%2 == 1 {
			base = math.Pi - base // mirror into the other diagonal direction
		}
		ax, ay := math.Cos(base), math.Sin(base) // avenue direction
		nx, ny := -ay, ax                        // its normal
		// Offset the line off dead-centre by a seeded fraction of the reach so avenues don't all
		// cross at one point (a small spread of parallel-ish grand avenues that span the silhouette).
		off := (float64(hash2(0xB0DE, uint32(0x40+d), seed))/float64(^uint32(0)) - 0.5) * 0.5 * reachR
		diags = append(diags, diagLine{nx: nx, ny: ny, c: off})
	}
	avenueHalf := 0.62 * cell // clear a band a bit over one cell wide → a clear avenue, grid intact
	onAvenue := func(x, y float64) bool {
		for _, dl := range diags {
			if math.Abs(dl.nx*x+dl.ny*y-dl.c) < avenueHalf {
				return true
			}
		}
		return false
	}
	// Build the in-shape jittered grid, dropping nodes that fall on an avenue; order centres-out. The
	// lattice box reaches the shape's max extent so a >disc footprint's corners get nodes.
	maxR := rr * tdShapeMaxReach(shape)
	type gnode struct {
		p  tdPoint
		d2 float64
	}
	var nodes []gnode
	half := int(math.Ceil(maxR/cell)) + 1
	if half > 64 {
		half = 64
	}
	rng := newRNG(hash2(0xB0DF, uint32(need), seed) | 1)
	jit := cell * 0.22 // a touch less jitter than the plain grid so avenues read cleanly
	for gy := -half; gy <= half; gy++ {
		for gx := -half; gx <= half; gx++ {
			nx := float64(gx)*cell + phx
			ny := float64(gy)*cell + phy
			jx := (float64(hash2(uint32(gx*2+1000), uint32(gy*2+1000), seed))/float64(^uint32(0)) - 0.5) * 2 * jit
			jy := (float64(hash2(uint32(gx*2+2000), uint32(gy*2+2000), seed))/float64(^uint32(0)) - 0.5) * 2 * jit
			px, py := nx+jx, ny+jy
			d2 := px*px + py*py
			if edge := tdShapeAcceptR(shape, px, py, townR, 0.92, seed); d2 > edge*edge {
				continue // outside the footprint (constant 0.92·townR for shapeDisc)
			}
			if onAvenue(px, py) {
				continue // this node is in an avenue → leave it clear
			}
			nodes = append(nodes, gnode{p: tdPoint{px, py}, d2: d2})
		}
	}
	sort.SliceStable(nodes, func(a, b int) bool { return nodes[a].d2 < nodes[b].d2 })
	for i := 0; i < need && i < len(nodes); i++ {
		seeds = append(seeds, nodes[i].p)
	}
	// Top up if we're short (avenues removed some, or a small footprint): drop the extras just OFF the
	// avenues AND inside the shape (uniform reject) so the count is exactly `need` without refilling the
	// boulevards. On an unlucky budget the LAST in-shape candidate is kept (no extra RNG draws), so for
	// shapeDisc — where every sample is in-shape — this matches the original disc top-up byte-for-byte.
	for placed := minInt(need, len(nodes)); placed < need; placed++ {
		var p tdPoint
		for tries := 0; tries < 24; tries++ {
			rad := maxR * math.Sqrt(rng.f01())
			ang := rng.f01() * 2 * math.Pi
			cx, cy := math.Cos(ang)*rad, math.Sin(ang)*rad
			edge := tdShapeAcceptR(shape, cx, cy, townR, 0.92, seed)
			if rad > edge {
				continue // outside the footprint — resample (shapeDisc never hits this)
			}
			p = tdPoint{cx, cy} // remember the last in-shape candidate
			if !onAvenue(cx, cy) {
				break // off the avenue → take it
			}
		}
		seeds = append(seeds, p)
	}
	return seeds
}

// tdScatterCoreSuburb is the DENSE-CORE / SPARSE-SUBURB scatter — the formCoreSuburb strategy
// (map-overhaul-citymap Phase 2a). It splits the free seeds into a packed downtown CORE and a thin
// suburban RING: ~65% of seeds are dropped DENSELY (tight blue-noise spacing) inside an inner core
// radius (~0.42·townR), and the remaining ~35% SPARSELY (large spacing, aggressive reject) across
// the outer ring out to the rim. The raster then reads as a compact packed downtown of small wards
// encircled by big loose low-density suburban blocks — the atomic-age metropolis with sprawl.
//
// Both passes place a fixed number so the total free-seed count is exactly `need` (banded stability
// holds); a relax tail guarantees the count even under unlucky packing. `seeds` already holds the
// pinned anchors. Pure over (seed, need); bounded attempt budgets → panic-safe.
func tdScatterCoreSuburb(seeds []tdPoint, townR float64, need int, shape tdFootprintShape, cfg tdConfig, seed uint32) []tdPoint {
	nPinned := len(seeds) // anchors already appended
	coreR := 0.38 * townR
	ringOuter := 0.92 * townR                     // leave a ground rim (the disc-reduction reference radius)
	ringMax := ringOuter * tdShapeMaxReach(shape) // suburb annulus reaches the silhouette (= ringOuter for a disc)
	nCore := (need * 72) / 100
	if nCore < 1 && need > 0 {
		nCore = 1 // at least seed the downtown
	}
	nRing := need - nCore
	rng := newRNG(hash2(0xC03E, uint32(need), seed) | 1)

	// firstFree indexes the first FREE seed so the reject passes ignore the pinned anchors (the core
	// packs right up to and around the centre, no ring-around-the-anchor void).
	firstFree := nPinned

	// --- dense CORE: tight blue-noise inside coreR ---
	if nCore > 0 {
		area := math.Pi * coreR * coreR
		minDist := 0.52 * math.Sqrt(area/float64(nCore)) // very tight downtown packing
		minD2 := minDist * minDist
		placed := 0
		maxAtt := (nCore + 1) * 40
		for att := 0; placed < nCore && att < maxAtt; att++ {
			rad := coreR * math.Sqrt(rng.f01())
			ang := rng.f01() * 2 * math.Pi
			p := tdPoint{math.Cos(ang) * rad, math.Sin(ang) * rad}
			ok := true
			for si := firstFree; si < len(seeds); si++ {
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
		for ; placed < nCore; placed++ { // relax tail → exact count
			rad := coreR * math.Sqrt(rng.f01())
			ang := rng.f01() * 2 * math.Pi
			seeds = append(seeds, tdPoint{math.Cos(ang) * rad, math.Sin(ang) * rad})
		}
	}

	// --- sparse SUBURB: large-spacing blue-noise in the annulus (coreR, ringMax), clipped to shape ---
	if nRing > 0 {
		// annArea scales with the footprint (areaFactor) so a >disc silhouette (halo) gets even LOOSER
		// suburban spacing — big low-density wards out to the bulging rim. For shapeDisc: factor 1.
		annArea := math.Pi * (ringOuter*ringOuter - coreR*coreR) * tdShapeAreaFactor(shape)
		// Deliberately LOOSE: bigger spacing target than an even fill → big suburban wards.
		minDist := 1.45 * math.Sqrt(annArea/float64(nRing))
		minD2 := minDist * minDist
		placed := 0
		maxAtt := (nRing + 1) * 40
		for att := 0; placed < nRing && att < maxAtt; att++ {
			// Uniform by area in the annulus out to the SHAPE reach: r = √(core² + u·(ringMax²−core²)),
			// then reject anything past the per-angle footprint outline. For shapeDisc ringMax = ringOuter
			// and the outline test is the constant 0.92·townR (never rejects) → identical to the old ring.
			u := rng.f01()
			rad := math.Sqrt(coreR*coreR + u*(ringMax*ringMax-coreR*coreR))
			ang := rng.f01() * 2 * math.Pi
			p := tdPoint{math.Cos(ang) * rad, math.Sin(ang) * rad}
			ok := true
			if edge := tdShapeAcceptR(shape, p.x, p.y, townR, 0.92, seed); rad > edge {
				ok = false // outside the footprint silhouette
			}
			// Reject against the SUBURB seeds only (indices from the start of the ring pass), so the
			// sparse spacing is enforced among suburbs without being blocked by the dense core.
			if ok {
				for si := firstFree + nCore; si < len(seeds); si++ {
					dx, dy := p.x-seeds[si].x, p.y-seeds[si].y
					if dx*dx+dy*dy < minD2 {
						ok = false
						break
					}
				}
			}
			if ok {
				seeds = append(seeds, p)
				placed++
			}
		}
		for ; placed < nRing; placed++ { // relax tail → exact count, clamped into the shape
			u := rng.f01()
			rad := math.Sqrt(coreR*coreR + u*(ringMax*ringMax-coreR*coreR))
			ang := rng.f01() * 2 * math.Pi
			x, y := math.Cos(ang)*rad, math.Sin(ang)*rad
			if edge := tdShapeAcceptR(shape, x, y, townR, 0.92, seed); rad > edge && rad > 0 {
				s := edge / rad
				x *= s
				y *= s
			}
			seeds = append(seeds, tdPoint{x, y})
		}
	}
	return seeds
}

// tdBuildBlockField builds the whole Voronoi block field: it sizes the raster to the town disc,
// scatters the block seeds by the town FORM (organic / radial / grid / ribbon), Lloyd-relaxes
// them, RASTER-partitions the disc by nearest seed, then derives the street (boundary) cells and
// per-block interior cell lists. Central seeds (the plaza anchors) are flagged so their regions
// stay OPEN and stay pinned through relaxation. Pure + deterministic; panic-safe (the grid is
// capped and every loop is bounded).
func tdBuildBlockField(townR float64, anchors []tdAnchor, nRoofs int, form tdTownForm, shape tdFootprintShape, cfg tdConfig, seed uint32) blockField {
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
	seeds := tdScatterSeedsFor(form, townR, anchors, B, shape, cfg, seed)
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
	seeds = tdLloyd(seeds, nPinned, gridN, cfg.cellSize, f.origin, townR, shape, seed, cfg.lloydPasses)
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
	// for shapeDisc; shapeBlob uses the irregular BLOB outline (tdInTown) so the silhouette rambles
	// rather than reading as a clean radial disc.
	f.nearest = make([]int, gridN*gridN)
	f.street = make([]bool, gridN*gridN)
	for gy := 0; gy < gridN; gy++ {
		for gx := 0; gx < gridN; gx++ {
			c := gy*gridN + gx
			p := f.cellCenter(gx, gy)
			if !tdInTown(p.x, p.y, townR, shape, seed) {
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
// for shapeDisc, the irregular BLOB outline (tdInTown) for shapeBlob — so relaxed centroids stay
// inside the same rambling town shape. Pure + deterministic; a seed that loses its whole region
// (rare) keeps its position. Panic-safe: bounded loops, guarded division.
func tdLloyd(seeds []tdPoint, nPinned, gridN int, cellSize, origin, townR float64, shape tdFootprintShape, seed uint32, passes int) []tdPoint {
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
				if !tdInTown(px, py, townR, shape, seed) {
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
func tdPlaceWonders(plan *topPlan, style tdEraStyle, cfg tdConfig) {
	for i, a := range plan.anchors {
		if !a.wonder {
			// A city with NO built wonder normally leaves its center a modest dressed square. But an
			// era whose IDENTITY is a monument (stone → a megalith circle, Phase 1b-i) must show that
			// centerpiece anyway, so the age reads distinct even when the civ has built no wonder.
			// Draw an UNLABELLED megalith monument roof at the center anchor, sized modestly so it sits
			// inside the small central square without needing a reserved plaza region. It reads through
			// the roofWonder → wonderMegalith dispatch exactly like a built wonder's centerpiece would.
			if i == 0 && style.wonderMotif == wonderMegalith {
				mScale := tdWonderScale(0) * 0.62 // modest — fits the wonderless center square
				plan.lots = append(plan.lots, tdLot{
					x: a.cx, y: a.cy, w: cfg.roofSize * mScale, h: cfg.roofSize * mScale,
					kind: tdRoof, domain: "monument", category: "monument",
					roof: roofWonder, prom: 600, // prominent but below a true labeled wonder
				})
			}
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

// tdSquarePropsFor returns the town-square prop palette for an era, keyed off the style's house
// profile (the era discriminator V3-B threads everywhere):
//   - STONE (wonderMegalith): MEGALITHS + standing stones, a well + firepit around a wonder (a
//     stone-circle forecourt); a megalith + well for a bare center. Checked FIRST off wonderMotif
//     because stone shares profileThatch with primitive — the motif is what distinguishes them.
//   - PRIMITIVE/default (profileThatch): a well, a firepit, standing stones/totem, and a market
//     stall around a wonder; a well + firepit for a bare center.
//   - ANCIENT (profileMudbrick): an ALTAR, COLUMNS, BRAZIERS + a well around a wonder (a temple
//     forecourt); a well + altar for a bare center.
//   - MEDIEVAL (profileTimber): MARKET STALLS, a FOUNTAIN, a well + a market CROSS/gallows around a
//     wonder (a market square); a well + fountain for a bare center.
//
// Every set keeps the deterministic ring placement + no-overlap-with-roof (tdRingProps).
func tdSquarePropsFor(style tdEraStyle) tdSquareProps {
	// Stone shares profileThatch with primitive, so it can't be told apart by the house profile;
	// its megalith motif is the discriminator. Dress its square with standing stones (Phase 1b-i).
	if style.wonderMotif == wonderMegalith {
		return tdSquareProps{
			wonder: []tdLotKind{tdPropMegalith, tdPropStones, tdPropWell, tdPropFirepit},
			center: []tdLotKind{tdPropMegalith, tdPropWell},
		}
	}
	// Industrial shares profileRowhouse with colonial, so it can't be told apart by the house profile;
	// its factory motif is the discriminator. Dress its square with SMOKESTACKS (Phase 1b-iii) so the
	// factory forecourt reads industrial, not the colonial statehouse green.
	if style.wonderMotif == wonderFactory {
		return tdSquareProps{
			wonder: []tdLotKind{tdPropSmokestack, tdPropWell, tdPropSmokestack, tdPropStall},
			center: []tdLotKind{tdPropSmokestack, tdPropWell},
		}
	}
	// Victorian shares BOTH profileRowhouse (colonial/industrial) AND wonderGeneric (colonial), so neither
	// discriminator tells it apart — its style NAME is the tag. Dress its square as a GASLIT PARK: gas-lamps
	// ringing a genteel forecourt (a fountain + a lamp for the modest centre), so the square reads Victorian,
	// not the colonial frontier green.
	if style.name == "victorian" {
		return tdSquareProps{
			wonder: []tdLotKind{tdPropGasLamp, tdPropFountain, tdPropGasLamp, tdPropStall},
			center: []tdLotKind{tdPropGasLamp, tdPropFountain},
		}
	}
	// Modern / information / digital all share the profileGlassTower dwelling (information + digital now carry
	// bespoke wonder motifs, but the HOUSE profile is shared), so the style NAME is the tag. Information dresses
	// its square with DATA CENTERS (a server forecourt); digital dresses its square with NEON SIGNS (the first
	// neon plaza). Modern keeps the default.
	if style.name == "information" {
		return tdSquareProps{
			wonder: []tdLotKind{tdPropDataCenter, tdPropWell, tdPropDataCenter, tdPropStall},
			center: []tdLotKind{tdPropDataCenter, tdPropWell},
		}
	}
	if style.name == "digital" {
		return tdSquareProps{
			wonder: []tdLotKind{tdPropNeonSign, tdPropDataCenter, tdPropNeonSign, tdPropWell},
			center: []tdLotKind{tdPropNeonSign, tdPropDataCenter},
		}
	}
	// Cyberpunk shares profileGlassTower + wonderSkyscraper with modern/information/digital, so the style
	// NAME is the tag. Dress its square with translucent HOLOGRAMS (the full night-city projection plaza).
	if style.name == "cyberpunk" {
		return tdSquareProps{
			wonder: []tdLotKind{tdPropHologram, tdPropNeonSign, tdPropHologram, tdPropWell},
			center: []tdLotKind{tdPropHologram, tdPropNeonSign},
		}
	}
	// Fusion has its own wonderFusionCore motif, but no dedicated props beyond a clean well — a minimalist
	// utopian square. Space shares its airy layout but is discriminated by NAME: dress its square with ROCKET
	// dabs (the spaceport forecourt).
	if style.name == "space" {
		return tdSquareProps{
			wonder: []tdLotKind{tdPropRocket, tdPropWell, tdPropRocket, tdPropStall},
			center: []tdLotKind{tdPropRocket, tdPropWell},
		}
	}
	// TRANSCENDENT (the finale) dresses its square with floating ethereal LIGHT MOTES — soft translucent
	// sparks over the luminous field, so the finale forecourt reads as pure light, not a paved plaza with
	// hard furniture. Name-gated (the two cosmic-second-pair ages share the profileEthereal/profileLattice
	// dwellings but want distinct square dressing).
	if style.name == "transcendent" {
		return tdSquareProps{
			wonder: []tdLotKind{tdPropLightMote, tdPropWell, tdPropLightMote, tdPropLightMote},
			center: []tdLotKind{tdPropLightMote, tdPropWell},
		}
	}
	// QUANTUM dresses its square with the epoch's NEON sign dabs — bright iridescent points that echo the
	// crystal city's shifting sheen, a spare cold-crystal forecourt (name-gated like transcendent).
	if style.name == "quantum" {
		return tdSquareProps{
			wonder: []tdLotKind{tdPropNeonSign, tdPropWell, tdPropNeonSign, tdPropStall},
			center: []tdLotKind{tdPropNeonSign, tdPropWell},
		}
	}
	switch style.houseProfile {
	case profileMudbrick: // ancient (bronze / iron)
		return tdSquareProps{
			wonder: []tdLotKind{tdPropAltar, tdPropColumns, tdPropBrazier, tdPropWell},
			center: []tdLotKind{tdPropWell, tdPropAltar},
		}
	case profileStoneClassical: // classical: columns-forward (a Greco-Roman forum)
		return tdSquareProps{
			wonder: []tdLotKind{tdPropColumns, tdPropAltar, tdPropWell},
			center: []tdLotKind{tdPropColumns, tdPropWell},
		}
	case profileTimber: // medieval
		return tdSquareProps{
			wonder: []tdLotKind{tdPropStall, tdPropFountain, tdPropWell, tdPropCross},
			center: []tdLotKind{tdPropWell, tdPropFountain},
		}
	default: // primitive / not-yet-tuned
		return tdSquareProps{
			wonder: []tdLotKind{tdPropWell, tdPropFirepit, tdPropStones, tdPropStall},
			center: []tdLotKind{tdPropWell, tdPropFirepit},
		}
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
			ringR := smallR * 0.55
			if i == 0 && style.wonderMotif == wonderMegalith {
				// A megalith monument now occupies this center (tdPlaceWonders). Grow the plaza + push
				// the prop ring OUTSIDE the monument footprint so the standing-stone dabs frame it
				// rather than overlapping the circle.
				monHalf := cfg.roofSize * tdWonderScale(0) * 0.62 / 2
				if want := monHalf + cfg.roofSize*0.9; smallR < want {
					smallR = want
				}
				ringR = monHalf + cfg.roofSize*0.7
			}
			plan.lots = append(plan.lots, tdLot{
				x: a.cx, y: a.cy, w: smallR * 2, h: smallR * 2, kind: tdPlaza,
			})
			tdRingProps(plan, a.cx, a.cy, ringR, props.center, uint32(i), seed)
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
	// the town rather than swamping it. STONE (wonderMegalith) is an EXPOSED HIGHLAND — thin out the
	// green (fewer gardens + trees) so it reads as rocky ground, not a wooded village; keep some so
	// it isn't barren.
	greenScale := 1.0
	if style.wonderMotif == wonderMegalith {
		greenScale = 0.5
	}
	gardens := int(dens * math.Sqrt(float64(roofN)) * 1.0 * greenScale)
	trees := int(dens * math.Sqrt(float64(roofN)) * 0.9 * greenScale)
	props := int(dens * math.Sqrt(float64(roofN)) * 0.6)
	ponds := 1 + int(dens*math.Sqrt(float64(roofN))*0.18)
	if ponds > 5 {
		ponds = 5 // a few, never many
	}

	// SPACE-AND-ABOVE (Phase 2c): a station in the void has no soil — SUPPRESS all greenery. Gardens,
	// street trees, edge groves, and decorative ponds are all zeroed for space-mode ages; only the
	// built structures + props (wells/stalls) and paved squares stay. Grove placement is gated on the
	// same flag further down.
	if style.spaceMode {
		gardens = 0
		trees = 0
		ponds = 0
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

	// Scattered STANDING STONES (stone age, Phase 1b-i): a handful of megalith dabs dotted through
	// the settlement (not only the central circle) so the megalithic theme reads across the whole
	// town at thumbnail scale. Deterministic (same seeded pick-without-replacement), a bit larger
	// than a plain prop so they stand out, but capped at 2–4 so they season rather than swamp.
	if style.wonderMotif == wonderMegalith {
		nStones := 2 + int(r.f01()*3) // 2..4
		if nStones > 4 {
			nStones = 4
		}
		for i := 0; i < nStones; i++ {
			p, ok := pick()
			if !ok {
				break
			}
			plan.lots = append(plan.lots, tdLot{
				x: p.x, y: p.y, w: cfg.roofSize * 0.7, h: cfg.roofSize * 0.7, kind: tdPropMegalith,
			})
		}
	}

	// Scattered SMOKESTACKS (industrial, Phase 1b-iii): a handful of tall-chimney dabs dotted through
	// the factory town (not only the central works) so the industrial skyline reads across the whole
	// town at thumbnail scale. Same deterministic seeded pick-without-replacement + 2–4 cap as the
	// stone-age megalith scatter, gated on the factory motif so only industrial towns get them.
	if style.wonderMotif == wonderFactory {
		nStacks := 2 + int(r.f01()*3) // 2..4
		if nStacks > 4 {
			nStacks = 4
		}
		for i := 0; i < nStacks; i++ {
			p, ok := pick()
			if !ok {
				break
			}
			plan.lots = append(plan.lots, tdLot{
				x: p.x, y: p.y, w: cfg.roofSize * 0.6, h: cfg.roofSize * 0.6, kind: tdPropSmokestack,
			})
		}
	}

	// Scattered GASLIT PARKS (victorian, ELECTRIC epoch): a handful of gas-lamps + a couple of extra green
	// park patches dotted through the city (not only the central square) so the genteel gaslit-boulevard
	// theme reads across the whole town at thumbnail scale. Same deterministic seeded pick-without-replacement
	// + small cap as the stone/megalith + industrial/smokestack scatters, gated on the style NAME so only
	// victorian towns get them (victorian shares its motif + profile with colonial/industrial).
	if style.name == "victorian" {
		nLamps := 3 + int(r.f01()*3) // 3..5 — enough to season the boulevards
		if nLamps > 5 {
			nLamps = 5
		}
		for i := 0; i < nLamps; i++ {
			p, ok := pick()
			if !ok {
				break
			}
			plan.lots = append(plan.lots, tdLot{
				x: p.x, y: p.y, w: cfg.roofSize * 0.5, h: cfg.roofSize * 0.5, kind: tdPropGasLamp,
			})
		}
		nParks := 1 + int(r.f01()*2) // 1..2 small green park patches
		if nParks > 2 {
			nParks = 2
		}
		for i := 0; i < nParks; i++ {
			p, ok := pick()
			if !ok {
				break
			}
			plan.lots = append(plan.lots, tdLot{
				x: p.x, y: p.y, w: cfg.roofSize * 1.1, h: cfg.roofSize * 1.1, kind: tdGarden,
			})
		}
	}

	// Scattered DATA CENTERS (information, DIGITAL epoch): a generous handful of low wide server-farm blocks
	// dotted through the city (not only the central square) so the SERVER-CITY theme reads strongly across the
	// whole town at thumbnail scale. Same deterministic seeded pick-without-replacement as the
	// industrial/smokestack scatter, but a heavier 3–6 cap (denser than before) so information reads clearly as
	// a server-city vs plain modern glass. Gated on the style NAME so only information towns get them (it shares
	// its glass-tower PROFILE with modern/digital, though its wonder motif is now the bespoke data hub).
	if style.name == "information" {
		nDC := 3 + int(r.f01()*4) // 3..6
		if nDC > 6 {
			nDC = 6
		}
		for i := 0; i < nDC; i++ {
			p, ok := pick()
			if !ok {
				break
			}
			plan.lots = append(plan.lots, tdLot{
				x: p.x, y: p.y, w: cfg.roofSize * 0.6, h: cfg.roofSize * 0.6, kind: tdPropDataCenter,
			})
		}
	}

	// Scattered NEON SIGNS + a couple of DATA CENTERS (digital, DIGITAL epoch): a handful of small neon-sign
	// dabs (the epoch's first neon) plus a data center or two, dotted through the sleek dark city so the
	// first-neon read carries across the whole town at thumbnail scale. Same seeded pick-without-replacement +
	// small caps, gated on the style NAME so only digital towns get them.
	if style.name == "digital" {
		nNeon := 3 + int(r.f01()*3) // 3..5 — enough to season the streets
		if nNeon > 5 {
			nNeon = 5
		}
		for i := 0; i < nNeon; i++ {
			p, ok := pick()
			if !ok {
				break
			}
			plan.lots = append(plan.lots, tdLot{
				x: p.x, y: p.y, w: cfg.roofSize * 0.5, h: cfg.roofSize * 0.5, kind: tdPropNeonSign,
			})
		}
		nDC := 1 + int(r.f01()*2) // 1..2 data centers
		if nDC > 2 {
			nDC = 2
		}
		for i := 0; i < nDC; i++ {
			p, ok := pick()
			if !ok {
				break
			}
			plan.lots = append(plan.lots, tdLot{
				x: p.x, y: p.y, w: cfg.roofSize * 0.6, h: cfg.roofSize * 0.6, kind: tdPropDataCenter,
			})
		}
	}

	// Scattered HOLOGRAMS (cyberpunk, NEON epoch): a handful of translucent floating projections dotted
	// through the dark night city so the hologram read carries across the whole town at thumbnail scale.
	// Same deterministic seeded pick-without-replacement + small cap as the neon/data-center scatter, gated on
	// the style NAME so only cyberpunk towns get them (it shares its motif + profile with modern/digital).
	if style.name == "cyberpunk" {
		nHolo := 3 + int(r.f01()*4) // 3..6 — enough to blaze the night city
		if nHolo > 6 {
			nHolo = 6
		}
		for i := 0; i < nHolo; i++ {
			p, ok := pick()
			if !ok {
				break
			}
			plan.lots = append(plan.lots, tdLot{
				x: p.x, y: p.y, w: cfg.roofSize * 0.55, h: cfg.roofSize * 0.55, kind: tdPropHologram,
			})
		}
	}

	// Scattered ROCKETS (space, NEON epoch): a LIGHT dusting of rocket/gantry dabs across the colony so the
	// spaceport read carries beyond the central pad. Same seeded pick-without-replacement + a small 1..3 cap,
	// gated on the style NAME so only space towns get them.
	if style.name == "space" {
		nRocket := 1 + int(r.f01()*3) // 1..3 — a light seasoning, not a launch field
		if nRocket > 3 {
			nRocket = 3
		}
		for i := 0; i < nRocket; i++ {
			p, ok := pick()
			if !ok {
				break
			}
			plan.lots = append(plan.lots, tdLot{
				x: p.x, y: p.y, w: cfg.roofSize * 0.6, h: cfg.roofSize * 0.6, kind: tdPropRocket,
			})
		}
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
	if style.wonderMotif == wonderMegalith {
		groveCount = 1 + int(r.f01()*2) // 1..2 — a sparse, exposed highland, not a wooded village
	}
	if style.spaceMode {
		groveCount = 0 // no forests ring a space station (Phase 2c)
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

// ---- walls (locked #9; ancient mudbrick + medieval stone) -------------------

// tdWallExtra is one non-segment wall feature (a TOWER or a GATEHOUSE), carried as its own lot so
// the renderer can draw it prominently. Towers/gatehouses are only emitted for the STONE wall
// (medieval); the mudbrick wall (ancient) is a plain curtain.
const (
	tdWallTower     tdLotKind = iota + 100 // a wall tower (stone wall only): a fat masonry block
	tdWallGatehouse                        // a gatehouse flanking a stone-wall gate: two towers + lintel
	tdWallBastion                          // an ANGULAR arrowhead bastion salient (renaissance star-fort only): a diamond jut-out, NOT a round drum
)

// tdWallRadiusAt is the wall RING radius at a given angle (city units). The wall follows the
// (ragged) town OUTLINE just OUTSIDE the outermost ward: for any NON-disc shape it rides the town's
// outline profile (tdShapeRadiusAt) so the rampart is ragged like the town it encloses; a plain disc
// uses a plain circle. Sized from the built-up footprint (so buildings stay INSIDE) but capped to the
// town disc so the ring never leaves the bounded canvas. Pure.
func tdWallRadiusAt(angle, footR, townR float64, shape tdFootprintShape, seed uint32) float64 {
	// Ring the built-up edge with a small margin so the outermost roofs sit inside the wall.
	r := footR * 1.12
	// Follow the ragged outline for any NON-disc shape: modulate the footprint ring by the town's
	// own outline profile at this angle (tdShapeRadiusAt/townR ∈ [floor,1]), so the rampart bites
	// inward on the town's bays and bulges on its peninsulas — a ragged wall around a ragged town.
	if shape != shapeDisc && townR > 0 {
		prof := tdShapeRadiusAt(shape, angle, townR, seed) / townR // 0..1 outline profile at this angle
		r = footR * (1.04 + 0.16*prof)
	}
	// Keep the ring inside the bounded town disc (+ a hair) so the wall never flies off-canvas.
	if lim := townR * 1.14; r > lim {
		r = lim
	}
	return r
}

// tdAddWalls rings the built-up area with a WALL + GATES that follow the (ragged) town outline
// just outside the outermost wards (locked #9, V3-B). It is DETERMINISTIC (seeded), BOUNDED (the
// ring is capped to the town disc), and preserves STREET CONNECTIVITY: the wall is a ring of
// segment lots with GAPS where the town's main/longest streets reach the rim — a small gate
// structure sits at each gap and the street EXITS through it. The street-cell network itself is
// never touched (the connectivity guarantee of the Voronoi model holds); the wall only rings it.
//
// Two dialects by wall profile:
//   - wallMudbrick (ancient): a thin tan curtain, four gates, no towers.
//   - wallStone (medieval): a thicker grey curtain, periodic TOWERS, and a GATEHOUSE at the main
//     gate (the longest street's exit).
//   - wallStarFort (renaissance): a thick earthwork rampart with periodic ANGULAR BASTION salients
//     (arrowhead jut-outs) instead of round towers, and plain gate blocks (no round gatehouse).
func tdAddWalls(plan *topPlan, style tdEraStyle, seed uint32) {
	footR := tdFootprintRadius(plan)
	if footR <= 0 {
		return
	}
	shape := plan.shape
	townR := plan.townR
	prof := style.wallProfile
	if prof == wallNone {
		prof = wallMudbrick // hasWalls set but no profile → a plain curtain (safe default)
	}

	// GATES follow the STREETS: a street must be able to exit the town, so a gate opens where a
	// main street reaches the wall ring. Rank the town's exit directions by how far their street
	// cells reach from the core (the longest radial streets are the main roads), dedupe by angle,
	// and open a gate at each. This guarantees every gate sits on a real street so connectivity
	// through the wall is preserved by construction. Fall back to cardinal gates if (degenerate) no
	// street reaches the rim.
	gateAngles := tdGateAngles(plan, footR, prof, seed)

	// Segment the ring finely so the curtain reads continuous. A segment whose angle falls within a
	// gate's arc is DROPPED (the gap) and replaced by the gate structure; the rest are wall.
	const segs = 48
	r := newRNG(hash2(0x3a11, uint32(len(plan.lots)), seed) | 1)
	phase := r.f01() * 2 * math.Pi
	// Wall thickness (city units): mudbrick ~thin, timber ~medium, stone ~a hair thicker, star-fort
	// ~thickest (a low broad earthwork rampart).
	segHalf := 0.85
	switch prof {
	case wallTimber:
		segHalf = 0.95 // between mudbrick (0.85) and stone (1.05) — a stout log palisade
	case wallStone:
		segHalf = 1.05
	case wallStarFort:
		segHalf = 1.10 // a thick low earthwork rampart (renaissance)
	}
	// Gate half-arc: the angular gap a gate opens in the ring (wide enough for a street to pass).
	gateArc := 0.16 // radians each side of the gate center
	// Tower cadence (stone only): a tower every few segments around the ring.
	const towerEvery = 6
	// Bastion cadence (star-fort only): an angular salient every few segments, spaced wider than the
	// stone towers so the pointed bastions read as distinct star points, not a dense studding.
	const bastionEvery = 8

	inGate := func(ang float64) (center float64, isGate bool) {
		for _, ga := range gateAngles {
			d := angDiff(ang, ga)
			if d < gateArc {
				return ga, true
			}
		}
		return 0, false
	}

	for i := 0; i < segs; i++ {
		ang := phase + 2*math.Pi*float64(i)/float64(segs)
		rad := tdWallRadiusAt(ang, footR, townR, shape, seed)
		if _, isGate := inGate(ang); isGate {
			continue // leave a GAP in the curtain here — the gate structure is placed below
		}
		x := plan.cx + math.Cos(ang)*rad
		y := plan.cy + math.Sin(ang)*rad
		plan.lots = append(plan.lots, tdLot{x: x, y: y, w: segHalf * 2, h: segHalf * 2, kind: tdWall})
		// Stone walls carry periodic ROUND towers between the gates.
		if prof == wallStone && i%towerEvery == 0 {
			plan.lots = append(plan.lots, tdLot{x: x, y: y, w: segHalf * 3.0, h: segHalf * 3.0, kind: tdWallTower})
		}
		// Star-fort walls carry periodic ANGULAR BASTIONS — arrowhead salients pushed OUTWARD from the
		// curtain so the trace reads as a pointed-bastion star (never a round tower). Kept off the gate
		// arcs so a bastion never blocks a gate. Pushed out ~1.5 seg beyond the rampart; the renderer
		// clamps every pixel, so it stays panic-safe even if a salient reaches the canvas edge.
		if prof == wallStarFort && i%bastionEvery == 0 {
			brad := rad + segHalf*1.5
			bx := plan.cx + math.Cos(ang)*brad
			by := plan.cy + math.Sin(ang)*brad
			plan.lots = append(plan.lots, tdLot{x: bx, y: by, w: segHalf * 3.0, h: segHalf * 3.0, kind: tdWallBastion})
		}
	}

	// Gate structures: a small gate block AT each gate gap so the opening reads as a real gate the
	// street passes through, not just a missing wall segment. The FIRST gate (the longest street,
	// the main road) gets a GATEHOUSE on a stone wall — flanking towers + a lintel across the gap.
	for gi, ga := range gateAngles {
		rad := tdWallRadiusAt(ga, footR, townR, shape, seed)
		gx := plan.cx + math.Cos(ga)*rad
		gy := plan.cy + math.Sin(ga)*rad
		plan.lots = append(plan.lots, tdLot{x: gx, y: gy, w: segHalf * 2, h: segHalf * 2, kind: tdGate})
		if prof == wallStone {
			// Flanking gate towers just to either side of the opening (tangent to the ring).
			tangent := ga + math.Pi/2
			off := (gateArc + 0.05) * rad
			for _, s := range []float64{-1, 1} {
				fx := gx + math.Cos(tangent)*off*s
				fy := gy + math.Sin(tangent)*off*s
				k := tdWallTower
				if gi == 0 {
					k = tdWallGatehouse // the main gate is a full gatehouse
				}
				plan.lots = append(plan.lots, tdLot{x: fx, y: fy, w: segHalf * 3.0, h: segHalf * 3.0, kind: k})
			}
		}
	}
}

// tdGateAngles picks the wall's gate directions from the town's MAIN STREETS so a gate always
// opens where a street reaches the rim (streets exit through gates → connectivity preserved). It
// buckets the street cells by angle around the core, takes the angle of the FARTHEST-reaching cell
// in each occupied bucket (the streets that actually run out to the wall), sorts those exit
// directions by reach (longest = the main road first), dedupes near-duplicate angles, and returns
// the top few. A stone wall gets 4 gates, a mudbrick wall 4 as well (a modest ancient town). If no
// street reaches near the rim (degenerate/tiny), it falls back to seeded ~cardinal gates so the
// ring is never gate-less. Deterministic (pure over the plan + seed).
func tdGateAngles(plan *topPlan, footR float64, prof wallProfile, seed uint32) []float64 {
	want := 4
	type exit struct {
		ang   float64
		reach float64
	}
	const buckets = 16
	best := make([]exit, buckets)
	for i := range best {
		best[i].reach = -1
	}
	for _, p := range plan.streetCells {
		dx, dy := p.x-plan.cx, p.y-plan.cy
		reach := math.Hypot(dx, dy)
		// Only streets that run OUT toward the wall are exit candidates (near the built-up rim).
		if reach < footR*0.6 {
			continue
		}
		a := math.Atan2(dy, dx)
		b := int((a + math.Pi) / (2 * math.Pi) * buckets)
		if b < 0 {
			b = 0
		}
		if b >= buckets {
			b = buckets - 1
		}
		if reach > best[b].reach {
			best[b] = exit{ang: a, reach: reach}
		}
	}
	var exits []exit
	for _, e := range best {
		if e.reach > 0 {
			exits = append(exits, e)
		}
	}
	// Longest-reaching exits first (the main roads); stable tiebreak by angle.
	sort.SliceStable(exits, func(i, j int) bool {
		if exits[i].reach != exits[j].reach {
			return exits[i].reach > exits[j].reach
		}
		return exits[i].ang < exits[j].ang
	})
	// Dedupe gates that sit too close together (keep them spread around the ring).
	var out []float64
	for _, e := range exits {
		ok := true
		for _, g := range out {
			if angDiff(e.ang, g) < 0.5 {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, e.ang)
		}
		if len(out) >= want {
			break
		}
	}
	// Fallback: no street reached the rim (tiny/degenerate town) → seeded ~cardinal gates so the
	// ring still opens and a street (which always reaches the rim in a real town) can exit.
	if len(out) == 0 {
		r := newRNG(hash2(0x6a7e, uint32(len(plan.streetCells)), seed) | 1)
		ph := r.f01() * 2 * math.Pi
		for i := 0; i < want; i++ {
			out = append(out, ph+2*math.Pi*float64(i)/float64(want))
		}
	}
	return out
}

// angDiff is the smallest absolute angular distance between two angles (radians), in [0,π].
func angDiff(a, b float64) float64 {
	d := math.Mod(math.Abs(a-b), 2*math.Pi)
	if d > math.Pi {
		d = 2*math.Pi - d
	}
	return d
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
	// Walls / gates / towers / gatehouses ring the built-up edge just outside the roofs (V3-B). Fit
	// them too so the whole enceinte stays BOUNDED on-canvas (locked #9: the wall inside the canvas)
	// rather than being flung off the frame edge.
	for _, lt := range plan.lots {
		switch lt.kind {
		case tdWall, tdGate, tdWallTower, tdWallGatehouse, tdWallBastion:
			acc(lt.x, lt.y, math.Max(lt.w, lt.h)/2)
		}
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
// cosmicSceneFor returns the bespoke scene renderer for a cosmic-scale age, if one exists.
// These ages abandon the city renderer entirely and draw a zoomed-out scene instead: at
// space-scale and above a top-down "city" is the wrong lens, so we pull the camera back to
// show the civilization at planetary (and, in later slices, interstellar/galactic) scale.
// Ages without a bespoke scene fall through to the normal city renderer.
func cosmicSceneFor(ageKey string) (func(img *image.RGBA, state game.GameState, w, h int, seed uint32), bool) {
	switch ageKey {
	case "space_age":
		return drawPlanetScene, true
	case "interstellar_age":
		return drawStarSystemScene, true
	case "galactic_age":
		return drawGalaxyScene, true
	case "quantum_age":
		return drawCosmicWebScene, true
	}
	return nil, false
}

func renderTopDown(img *image.RGBA, state game.GameState, w, h int, seed uint32) layoutGeometry {
	// Cosmic-scale ages replace the city entirely with a zoomed-out scene (the home planet from
	// orbit, a starfield, and so on). Draw it and return an empty geometry so the overlay pass
	// stamps NO city landmarks/labels over the scene.
	if scene, ok := cosmicSceneFor(state.Age); ok {
		scene(img, state, w, h, seed)
		return layoutGeometry{}
	}

	style := styleForAge(state.Age)
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
		case tdPropWell, tdPropFirepit, tdPropStones, tdPropStall,
			tdPropAltar, tdPropColumns, tdPropBrazier, tdPropFountain, tdPropCross,
			tdPropSmokestack, tdPropGasLamp, tdPropDataCenter, tdPropNeonSign,
			tdPropHologram, tdPropRocket, tdPropLightMote:
			drawSquareProp(img, xf, lt, style, pal)
		}
	}

	// Walls + gates + towers last among pixels (if any) so the ring crowns the built-up edge.
	// Towers/gatehouses draw AFTER the curtain so they sit proud of it. Every tone is the era wall
	// recipe (theme-derived → retints); the gate reads lighter (an opening) and the tower darker +
	// larger (a solid mass) so the ring reads as a real fortification, not a dotted circle.
	if style.hasWalls {
		wallCol := style.wallCol(pal)
		gateCol := brighten(wallCol, 0.28)
		towerCol := darken(wallCol, 0.14)
		towerCap := brighten(wallCol, 0.10)
		// Timber palisade (iron): a darker vertical post streak down the middle of each segment so the
		// brown curtain reads as upright logs, not a smooth rampart. Cheap + bounds-checked.
		postCol := darken(wallCol, 0.22)
		for _, lt := range plan.lots {
			if lt.kind == tdWall {
				cx, cy := xf.px(lt.x, lt.y)
				rad := xf.ext(lt.w / 2)
				drawBlock(img, cx, cy, rad, wallCol)
				if style.wallProfile == wallTimber {
					for dy := -rad; dy <= rad; dy++ {
						setPixel(img, cx, cy+dy, postCol) // a central log seam
					}
				}
			}
		}
		for _, lt := range plan.lots {
			switch lt.kind {
			case tdGate:
				cx, cy := xf.px(lt.x, lt.y)
				drawBlock(img, cx, cy, xf.ext(lt.w/2), gateCol)
			case tdWallTower, tdWallGatehouse:
				cx, cy := xf.px(lt.x, lt.y)
				rad := xf.ext(lt.w / 2)
				fillDisc(img, cx, cy, rad, towerCol) // a fat masonry drum
				setPixel(img, cx, cy, towerCap)      // a lit crenellation cap
				if lt.kind == tdWallGatehouse {
					// A gatehouse reads a touch taller: a second ring of cap dabs (crenellations).
					forRectOutline(cx, cy, rad, rad, func(x, y int) { setPixel(img, x, y, towerCap) })
				}
			case tdWallBastion:
				// A star-fort ANGULAR bastion: an arrowhead salient pointing OUTWARD from the core, never
				// a round drum. Drawn as a filled DIAMOND whose OUTER vertex is stretched along the radial
				// (the pointed tip) so it reads as a triangular jut-out. All writes clamp → panic-safe.
				cx, cy := xf.px(lt.x, lt.y)
				ccx, ccy := xf.px(plan.cx, plan.cy)
				rad := xf.ext(lt.w / 2)
				drawBastion(img, cx, cy, ccx, ccy, rad, towerCol, towerCap)
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
	// SPACE-AND-ABOVE (Phase 2c): the five cosmic ages read as a station floating in the void, not a
	// town on terrain — swap the ground tint for a deep-space VOID + STARFIELD and return.
	if style.spaceMode {
		drawSpaceBackground(img, style, pal, seed, w, h)
		return
	}
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

// drawSpaceBackground paints the SPACE-AND-ABOVE ground (Phase 2c): a deep near-black VOID with a
// sparse deterministic STARFIELD and a couple of very faint NEBULA washes, so the cosmic ages read
// as a station/platform floating in space rather than a town on terrain. Called from drawGround when
// style.spaceMode is set; replaces the ground tint entirely.
//
// Every color is theme-derived so it retints on a theme switch: the void starts from the theme
// background pulled dark toward black; the nebula tints from accent/highlight; the stars brighten
// toward highlight/white. Deterministic (seeded), panic-safe (clipped setters), exact output size —
// every pixel is written by the void pass before the stars/nebula stipple over it.
func drawSpaceBackground(img *image.RGBA, style tdEraStyle, pal tdPal, seed uint32, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	b := img.Bounds()

	// VOID: a deep near-black space fill, theme-derived. Pull the background toward the dim role, then
	// hard toward black so it reads as a clearly dark void (not the grey deck). A very subtle vertical
	// gradient (top darkest, bottom lifted a hair toward the void-blue) gives depth while staying calm.
	voidTop := darken(blend(pal.bg, pal.dim, 0.30), 0.72)
	voidBot := blend(darken(blend(pal.bg, pal.dim, 0.30), 0.62), voidAnchor, 0.35)

	// NEBULA: 1–2 soft broad washes for depth. Seeded centers; a faint accent/highlight tint blended in
	// with a smooth falloff. Kept very subtle (peak ~0.12) so it adds atmosphere without muddying the
	// void. Deterministic per seed.
	type nebula struct {
		cx, cy, r float64
		tint      color.RGBA
		peak      float64
	}
	nebN := 1 + int(hash2(0x4EB0, seed, 0x51)%2) // 1..2
	nebs := make([]nebula, 0, nebN)
	for i := 0; i < nebN; i++ {
		hx := hashUnit(uint32(i)*2+1, 0x1A, seed)
		hy := hashUnit(uint32(i)*2+2, 0x2B, seed)
		hr := hashUnit(uint32(i)+7, 0x3C, seed)
		tintPick := hashUnit(uint32(i)+11, 0x4D, seed)
		tint := pal.accent
		if tintPick > 0.5 {
			tint = pal.highlight
		}
		nebs = append(nebs, nebula{
			cx:   hx * float64(w),
			cy:   hy * float64(h),
			r:    (0.35 + 0.25*hr) * float64(w), // broad — a third to half the canvas wide
			tint: tint,
			peak: 0.08 + 0.04*tintPick, // ~0.08..0.12
		})
	}

	for y := 0; y < h; y++ {
		py := b.Min.Y + y
		if py < b.Min.Y || py >= b.Max.Y {
			continue
		}
		vt := float64(y) / float64(h) // 0 (top) .. ~1 (bottom)
		row := blend(voidTop, voidBot, vt)
		for x := 0; x < w; x++ {
			px := b.Min.X + x
			if px < b.Min.X || px >= b.Max.X {
				continue
			}
			c := row
			// Faint nebula wash: smooth quadratic falloff from each center, capped subtle.
			for _, n := range nebs {
				dx := float64(x) - n.cx
				dy := float64(y) - n.cy
				d2 := dx*dx + dy*dy
				r2 := n.r * n.r
				if d2 < r2 {
					f := 1 - d2/r2 // 1 at center → 0 at edge
					c = blend(c, n.tint, n.peak*f*f)
				}
			}
			// A whisper of per-pixel value noise so the void isn't a dead flat wash (very small).
			flick := texHash(uint32(x), uint32(y), seed^0x5AC3)
			if flick > 0.85 {
				c = brighten(c, 0.02)
			}
			img.SetRGBA(px, py, c)
		}
	}

	// STARFIELD: scatter sparse deterministic stars over the void. Budget ~ (w*h)/45 candidates; each a
	// 1px bright dab (blend the local void toward highlight/white by a varied brightness), most faint with
	// a few bright; ~5% are 2px "bright stars". Tasteful, not a snowstorm. Bounds-checked via setPixel.
	starWhite := blend(pal.highlight, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, 0.55)
	candidates := (w * h) / 45
	for i := 0; i < candidates; i++ {
		si := uint32(i)
		sx := int(hash2(si, 0xA1, seed) % uint32(w))
		sy := int(hash2(si, 0xB2, seed) % uint32(h))
		br := hashUnit(si, 0xC3, seed) // brightness roll
		// Most stars faint: bias the distribution so only a minority read bright. Skip the very
		// dimmest rolls entirely so the field stays sparse rather than a uniform haze.
		if br < 0.45 {
			continue
		}
		amt := 0.35 + 0.65*((br-0.45)/0.55) // 0.35 (faint) .. 1.0 (bright)
		base := blend(voidTop, starWhite, amt)
		setPixel(img, b.Min.X+sx, b.Min.Y+sy, base)
		// A handful (~5%) are 2px bright stars — a small NE dab beside the core.
		if br > 0.95 {
			setPixel(img, b.Min.X+sx+1, b.Min.Y+sy, blend(voidTop, starWhite, amt*0.7))
			setPixel(img, b.Min.X+sx, b.Min.Y+sy-1, blend(voidTop, starWhite, amt*0.7))
		}
	}
}

// Planet-scene palette anchors: a planet reads BLUE/GREEN/WHITE regardless of theme, so the globe
// tones are near-fixed (a faint theme-accent nudge is applied where noted, but the world must still
// read as a world). Kept package-level and deterministic.
var (
	planetOcean = color.RGBA{R: 0x1b, G: 0x40, B: 0x78, A: 0xff} // deep ocean blue
	planetShoal = color.RGBA{R: 0x2c, G: 0x63, B: 0x8f, A: 0xff} // shallow/coastal blue (land–sea seam)
	planetLandL = color.RGBA{R: 0x4a, G: 0x74, B: 0x3c, A: 0xff} // lowland green
	planetLandH = color.RGBA{R: 0x7a, G: 0x64, B: 0x3e, A: 0xff} // upland brown (mountains/desert)
	planetIce   = color.RGBA{R: 0xe8, G: 0xf1, B: 0xf6, A: 0xff} // polar ice / cloud white
	planetAtmo  = color.RGBA{R: 0x9c, G: 0xd8, B: 0xff, A: 0xff} // bright atmosphere-rim blue
)

// drawPlanetScene renders the SPACE-AGE scene: the civilization's home planet seen from orbit,
// ringed by satellites and an orbital station, floating on the void+starfield. It abandons the
// city renderer entirely — there is no top-down city at this scale.
//
// The globe is drawn as a lit SPHERE (not a flat disc): per interior pixel we reconstruct a
// surface normal from the disc coordinates (nz = sqrt(1 - nx^2 - ny^2)) and shade by a fixed light
// direction, so the lit hemisphere is bright and the far limb falls through a soft TERMINATOR into
// a near-black night side. Over that we lay procedural CONTINENTS (two octaves of value-noise,
// thresholded so land clusters instead of speckling), POLAR ICE CAPS (high-latitude whitening),
// soft CLOUD bands, and a thin bright ATMOSPHERE rim on the lit limb. Everything is seeded and
// panic-safe (every write clips via setPixel/clamped math).
// planetAspectY squashes the planet + orbits VERTICALLY so they read ROUND in the terminal. The
// citymap image is cols×rows*2 px, streamed via half-blocks assuming a 1:2 terminal cell; on cells
// TALLER than that a circle-in-image renders as a tall egg. This factor pre-compensates. Tune it if
// the planet still looks off: LOWER = shorter/wider, HIGHER (→1.0) = taller. 1.0 = no correction
// (correct for an exact 1:2 cell).
const planetAspectY = 0.62

func drawPlanetScene(img *image.RGBA, state game.GameState, w, h int, seed uint32) {
	if w <= 0 || h <= 0 {
		return
	}

	// BACKDROP: the shared void + starfield (deep space, seeded stars). Reuse the space style/pal so
	// the void tone matches the rest of the cosmic-era treatment.
	pal := newTdPal()
	drawSpaceBackground(img, styleForAge("space_age"), pal, seed, w, h)

	b := img.Bounds()
	minWH := w
	if h < minWH {
		minWH = h
	}
	fmin := float64(minWH)

	// PLANET geometry: centered a touch off-center so the composition isn't dead-centered, radius a
	// bit over a quarter of the short side. Clamp radius so a tiny canvas still yields >=1px.
	cx := float64(b.Min.X) + 0.50*float64(w)
	cy := float64(b.Min.Y) + 0.52*float64(h)
	R := 0.28 * fmin
	if R < 1 {
		R = 1
	}

	// LIGHT direction (unit, in screen space with a +z toward the viewer): lit from the upper-left,
	// tilted slightly toward the camera so a sliver of the near face stays bright. Seed nudges it a
	// hair so different civs light differently, but it stays an upper-left key light.
	lang := -2.3 + (hashUnit(0x11, 0x22, seed)-0.5)*0.7 // radians, upper-left-ish
	lz := 0.55
	lx := math.Cos(lang) * math.Sqrt(1-lz*lz)
	ly := math.Sin(lang) * math.Sqrt(1-lz*lz)

	// A faint theme-accent nudge for the ocean so the world picks up a whisper of the era mood
	// without ceasing to read as blue.
	ocean := blend(planetOcean, pal.accent, 0.10)

	// Vertical radius: squashed so the disc reads round in the terminal (see planetAspectY).
	radY := R * planetAspectY
	// Bounding box of the (elliptical) disc, clamped to the image; iterate only there.
	x0 := int(math.Floor(cx - R - 2))
	x1 := int(math.Ceil(cx + R + 2))
	y0 := int(math.Floor(cy - radY - 2))
	y1 := int(math.Ceil(cy + radY + 2))
	if x0 < b.Min.X {
		x0 = b.Min.X
	}
	if y0 < b.Min.Y {
		y0 = b.Min.Y
	}
	if x1 > b.Max.X {
		x1 = b.Max.X
	}
	if y1 > b.Max.Y {
		y1 = b.Max.Y
	}

	invR := 1.0 / R
	invRadY := 1.0 / radY // normalize the vertical by the squashed radius so the disc is an ellipse
	// Continent noise scale: features span a good fraction of the globe (land clusters, not confetti).
	// Scale off the radius so the look is stable across map/minimap sizes.
	nScale := 3.2 / R
	// Terminator softness (in dot-product units): wider band → softer day/night transition.
	const termSoft = 0.42

	for py := y0; py < y1; py++ {
		dyf := (float64(py) - cy) * invRadY // -1..1 across the (squashed) disc vertically
		for px := x0; px < x1; px++ {
			dxf := (float64(px) - cx) * invR
			r2 := dxf*dxf + dyf*dyf
			if r2 > 1.0 {
				continue // outside the globe
			}
			// Surface normal of the sphere at this pixel (nz toward viewer).
			nz := math.Sqrt(1.0 - r2)
			nx, ny := dxf, dyf

			// --- SURFACE ALBEDO: ocean vs. continents -------------------------------------
			// Sample value-noise in a lightly latitude-warped screen space so land forms clustered
			// masses. Two octaves via texHash on a coarse lattice with bilinear-ish smoothing.
			lat := ny // -1 (north) .. 1 (south)
			sxf := (float64(px)-cx)*nScale + 32.0
			syf := (float64(py)-cy)*nScale*(1.0+0.25*lat*lat) + 32.0
			n := valueNoise(sxf, syf, seed^0x9111)*0.65 +
				valueNoise(sxf*2.03+11.0, syf*2.03+7.0, seed^0x5223)*0.35
			// Latitude bias: a touch more land in the mid-latitudes, more ocean at the equator, so the
			// masses don't ring the whole globe uniformly.
			landScore := n + 0.10*math.Abs(lat) - 0.06

			var albedo color.RGBA
			switch {
			case landScore < 0.50:
				albedo = ocean
			case landScore < 0.54:
				albedo = planetShoal // thin coastal seam
			default:
				// Land: greener in the lowlands, browner on the "high" (noisier) ground.
				up := (landScore - 0.54) / 0.46 // 0..~1
				if up > 1 {
					up = 1
				}
				albedo = blend(planetLandL, planetLandH, up*up)
			}

			// --- POLAR ICE CAPS: whiten the high-latitude caps (top & bottom of the disc) ---
			absLat := math.Abs(lat)
			if absLat > 0.72 {
				capF := (absLat - 0.72) / 0.28 // 0 at cap edge → 1 at pole
				if capF > 1 {
					capF = 1
				}
				// Ragged cap edge via noise so it isn't a clean band.
				capF *= 0.6 + 0.4*valueNoise(sxf*1.7, syf*1.7+50.0, seed^0x3C1D)
				albedo = blend(albedo, planetIce, clampF(capF, 0, 1)*0.9)
			}

			// --- CLOUDS: a few faint white bands/wisps over both land and sea ----------------
			cloudBand := math.Sin(lat*3.0 + 1.2)                            // broad latitude banding
			cloudTex := valueNoise(sxf*0.9+70.0, syf*0.9-40.0, seed^0x77A5) // wispy modulation
			cloud := (cloudBand*0.5 + 0.5) * cloudTex                       // 0..1
			cloud = (cloud - 0.55) / 0.45                                   // threshold → sparse
			if cloud > 0 {
				albedo = blend(albedo, planetIce, clampF(cloud, 0, 1)*0.45)
			}

			// --- SHADING: day/night terminator makes it read as a 3D sphere ------------------
			// Lambert term, remapped through a soft terminator: full-bright on the lit hemisphere,
			// fading to near-black on the night limb.
			ndl := nx*lx + ny*ly + nz*lz                 // -1 (night) .. 1 (noon)
			lit := smoothstepF(-termSoft, termSoft, ndl) // 0 night .. 1 day
			// Day side: gentle brighten toward the subsolar point; night side: crush toward black.
			shade := 0.06 + 0.94*lit // ambient floor so the night side isn't pure black
			out := scaleRGB(albedo, shade)
			// A subtle specular sky-glint on the ocean near the subsolar point (only where lit & wet).
			if albedo == ocean && ndl > 0.86 {
				out = blend(out, planetAtmo, (ndl-0.86)/0.14*0.35)
			}

			// --- ATMOSPHERE RIM: a thin bright blue ring just inside the LIT limb -------------
			// Near the edge (r2 high) AND on the lit side, add a rim glow. Fades quickly inward.
			edge := math.Sqrt(r2)
			if edge > 0.90 && lit > 0.15 {
				rim := (edge - 0.90) / 0.10 // 0..1 across the rim band
				out = blend(out, planetAtmo, clampF(rim, 0, 1)*0.6*lit)
			}

			setPixel(img, px, py, out)
		}
	}

	// SATELLITES + ORBITS: a few tilted elliptical orbit rings (thin, dim), each carrying a satellite
	// or two, plus one slightly larger station module. Drawn AFTER the planet so they read as
	// foreground hardware; seeded so positions are stable per civ.
	drawOrbitsAndSatellites(img, cx, cy, R, fmin, pal, seed)
}

// drawOrbitsAndSatellites paints 2–4 faint tilted orbit ellipses around the planet and dabs
// satellites (a small bright cluster, optionally a tiny solar-panel cross) plus one larger station
// module onto them. All seeded and panic-safe. cx,cy = planet center (pixel space), R = planet
// radius, fmin = min(w,h) for scale.
func drawOrbitsAndSatellites(img *image.RGBA, cx, cy, R, fmin float64, pal tdPal, seed uint32) {
	orbitCol := blend(pal.dim, pal.highlight, 0.30) // faint, cool ring line
	satBright := blend(pal.highlight, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, 0.5)
	panelCol := blend(pal.accent, satBright, 0.4)

	nOrbits := 2 + int(hash2(0x0B17, seed, 0x33)%3) // 2..4
	// Track where satellites will go so we can drop the station on the outermost ring.
	for oi := 0; oi < nOrbits; oi++ {
		si := uint32(oi)
		// Semi-axes: rings sit outside the planet, growing outward; tilt (squash the minor axis) and
		// a rotation give varied ellipses. Scale off the planet radius so it holds at any size.
		ra := R * (1.20 + 0.45*float64(oi) + 0.10*hashUnit(si, 0xA1, seed))
		rb := ra * (0.28 + 0.40*hashUnit(si, 0xB2, seed)) // flattened (viewed near edge-on)
		rot := hashUnit(si, 0xC3, seed) * math.Pi         // ring plane rotation
		cosR, sinR := math.Cos(rot), math.Sin(rot)

		// Draw the ring outline by stepping the parametric ellipse. Step count scales with size so a
		// big canvas gets a smooth ring and a tiny one still closes.
		steps := int(ra * 2.2)
		if steps < 40 {
			steps = 40
		}
		if steps > 900 {
			steps = 900
		}
		for s := 0; s < steps; s++ {
			t := float64(s) / float64(steps) * 2 * math.Pi
			ex := ra * math.Cos(t)
			ey := rb * math.Sin(t)
			// Rotate the ellipse in-plane.
			rx := ex*cosR - ey*sinR
			ry := ex*sinR + ey*cosR
			px := int(cx + rx)
			py := int(cy + ry*planetAspectY) // squash vertically to match the planet disc
			// Only the ring portion in FRONT of the globe (lower half of the tilt) reads a touch
			// brighter; the far arc is dimmer, hinting the ring passes behind the planet.
			c := orbitCol
			if ey > 0 {
				c = brighten(orbitCol, 0.15)
			} else {
				c = darken(orbitCol, 0.20)
			}
			setPixel(img, px, py, c)
		}

		// Place 1–2 satellites on this ring at seeded parameter angles.
		nSat := 1 + int(hash2(si, seed, 0x5E)%2) // 1..2
		for k := 0; k < nSat; k++ {
			kk := uint32(oi*7 + k + 1)
			t := hashUnit(kk, 0xD4, seed) * 2 * math.Pi
			ex := ra * math.Cos(t)
			ey := rb * math.Sin(t)
			rx := ex*cosR - ey*sinR
			ry := ex*sinR + ey*cosR
			sx := int(cx + rx)
			sy := int(cy + ry*planetAspectY)
			drawSatellite(img, sx, sy, satBright, panelCol, false)
		}
	}

	// ONE station module: a slightly larger cluster on the outermost ring at a fixed-ish angle.
	{
		oi := nOrbits - 1
		si := uint32(oi)
		ra := R * (1.20 + 0.45*float64(oi) + 0.10*hashUnit(si, 0xA1, seed))
		rb := ra * (0.28 + 0.40*hashUnit(si, 0xB2, seed))
		rot := hashUnit(si, 0xC3, seed) * math.Pi
		cosR, sinR := math.Cos(rot), math.Sin(rot)
		t := hashUnit(0x5747, 0x10, seed) * 2 * math.Pi
		ex := ra * math.Cos(t)
		ey := rb * math.Sin(t)
		rx := ex*cosR - ey*sinR
		ry := ex*sinR + ey*cosR
		stx := int(cx + rx)
		sty := int(cy + ry*planetAspectY)
		drawSatellite(img, stx, sty, satBright, panelCol, true)
	}
}

// drawSatellite dabs a single satellite (or, if station, a slightly larger module) at (x,y): a
// bright core with a tiny cross of solar panels. Panic-safe via setPixel. On a tiny canvas this
// still leaves at least a recognizable bright dot.
func drawSatellite(img *image.RGBA, x, y int, core, panel color.RGBA, station bool) {
	// Bright body.
	setPixel(img, x, y, core)
	if station {
		// A 2px-wide body block for the station.
		setPixel(img, x+1, y, core)
		setPixel(img, x, y+1, core)
		setPixel(img, x+1, y+1, core)
	}
	// Solar-panel cross: dabs one cell out on each axis (the "wings").
	setPixel(img, x-1, y, panel)
	setPixel(img, x+1+bi(station), y, panel)
	setPixel(img, x, y-1, panel)
	setPixel(img, x, y+1+bi(station), panel)
	if station {
		// Longer wings for the station.
		setPixel(img, x-2, y, panel)
		setPixel(img, x+2+bi(station), y, panel)
	}
}

// ---------------------------------------------------------------------------------------------
// INTERSTELLAR scene: the civilization pulled back one more level, from a single planet to its
// whole STAR SYSTEM — a central sun with orbiting worlds, an asteroid belt, and probes streaking
// outward toward neighbor stars. The 2nd cosmic scene (after the space-age planet), sharing its
// craft: crisp per-pixel sphere shading, procedural surfaces, deep-space void, NO soft halos.
// ---------------------------------------------------------------------------------------------

// starSurfacePalette / planet hues: like the planet scene these read as celestial bodies regardless
// of theme, so the tones are near-fixed (a faint accent nudge is applied on the ocean world where
// noted). Kept package-level and deterministic.
var (
	starCore = color.RGBA{R: 0xff, G: 0xf6, B: 0xe4, A: 0xff} // white-hot photosphere core
	starMid  = color.RGBA{R: 0xff, G: 0xd0, B: 0x5a, A: 0xff} // yellow mid-latitudes
	starLimb = color.RGBA{R: 0xf2, G: 0x8a, B: 0x24, A: 0xff} // orange cooler limb
	starSpot = color.RGBA{R: 0x9c, G: 0x51, B: 0x14, A: 0xff} // darker sunspot umbra

	// Planet surface anchors (rocky / ocean / mars-like / gas-giant bands).
	worldRockL = color.RGBA{R: 0x6d, G: 0x6a, B: 0x66, A: 0xff} // grey rock lowland
	worldRockH = color.RGBA{R: 0x9a, G: 0x96, B: 0x8e, A: 0xff} // lighter cratered upland
	worldSeaLo = color.RGBA{R: 0x16, G: 0x3a, B: 0x6e, A: 0xff} // deep ocean
	worldSeaHi = color.RGBA{R: 0x3f, G: 0x8a, B: 0x64, A: 0xff} // green landmass on the ocean world
	worldRedL  = color.RGBA{R: 0x8f, G: 0x40, B: 0x24, A: 0xff} // mars rust lowland
	worldRedH  = color.RGBA{R: 0xc2, G: 0x71, B: 0x3e, A: 0xff} // brighter oxide upland
	worldGasA  = color.RGBA{R: 0xc8, G: 0xa2, B: 0x6a, A: 0xff} // gas-giant warm band
	worldGasB  = color.RGBA{R: 0x8a, G: 0x67, B: 0x46, A: 0xff} // gas-giant dark band
	ringColor  = color.RGBA{R: 0xcf, G: 0xc2, B: 0xa4, A: 0xff} // thin planetary ring (icy tan)
)

// planetKind enumerates the four VARIED interstellar worlds; each shades its surface differently.
type planetKind int

const (
	worldRocky planetKind = iota // grey cratered
	worldOcean                   // blue ocean + green land
	worldMars                    // red oxide
	worldGas                     // banded gas giant (latitude bands)
)

// drawStarSystemScene renders the INTERSTELLAR-AGE scene: a star system — a central sun with a
// handful of orbiting worlds, an asteroid belt, and outbound probes — floating on the deep-space
// void+starfield. Like the planet scene it abandons the city renderer entirely; at this scale a
// top-down city is meaningless. Every round body (the star and each planet) applies planetAspectY
// so it reads ROUND in the terminal (see the const's note). Seeded and panic-safe throughout.
func drawStarSystemScene(img *image.RGBA, state game.GameState, w, h int, seed uint32) {
	if w <= 0 || h <= 0 {
		return
	}

	// BACKDROP: the shared void + starfield. Reuse the interstellar style/pal so the void tone
	// matches the rest of the cosmic-era treatment.
	pal := newTdPal()
	drawSpaceBackground(img, styleForAge("interstellar_age"), pal, seed, w, h)

	b := img.Bounds()
	minWH := w
	if h < minWH {
		minWH = h
	}
	fmin := float64(minWH)

	// STAR at the system barycenter, nudged a touch up-left of center so the composition isn't dead
	// centered and the outbound probes have room to streak toward the lower-right frame.
	scx := float64(b.Min.X) + 0.44*float64(w)
	scy := float64(b.Min.Y) + 0.48*float64(h)
	starR := 0.14 * fmin
	if starR < 1 {
		starR = 1
	}

	drawStar(img, scx, scy, starR, pal, seed)

	// PLANETS on elliptical orbits at increasing radii. Deterministic count (3..5) and a fixed roster
	// of VARIED kinds so every system shows contrast (rocky / ocean / mars / gas giant, cycling). Each
	// world is a small sphere-shaded globe lit from the star (light points from the world TOWARD the
	// star). Orbit y is squashed by planetAspectY so the rings read round in the terminal.
	nPlanets := 3 + int(hash2(0x9142, seed, 0x7B)%3) // 3..5
	kinds := []planetKind{worldRocky, worldOcean, worldMars, worldGas}
	// Orbit radii step outward from just beyond the star; the outermost fits inside the short side.
	orbit0 := starR * 1.9
	orbitStep := (0.46*fmin - orbit0) / float64(nPlanets)
	if orbitStep < starR*0.7 {
		orbitStep = starR * 0.7
	}
	// Remember planet orbit radii so the asteroid belt can sit between two of them.
	orbitRs := make([]float64, nPlanets)
	for i := 0; i < nPlanets; i++ {
		si := uint32(i) + 1
		orbR := orbit0 + orbitStep*float64(i)
		orbitRs[i] = orbR
		// Angle around the star; spread deterministically so worlds don't line up.
		ang := hashUnit(si, 0x2C1F, seed)*2*math.Pi + float64(i)*1.7
		px := scx + orbR*math.Cos(ang)
		py := scy + orbR*math.Sin(ang)*planetAspectY // squash the orbit to match the terminal

		// World radius: varies with kind and a seeded jitter; gas giants biggest, rocky smallest. Scale
		// off the short side so it holds at minimap size.
		kind := kinds[i%len(kinds)]
		baseR := 0.032 * fmin
		switch kind {
		case worldGas:
			baseR = 0.055 * fmin
		case worldOcean:
			baseR = 0.040 * fmin
		case worldMars:
			baseR = 0.036 * fmin
		}
		pr := baseR * (0.82 + 0.36*hashUnit(si, 0x3D2E, seed))
		if pr < 1.2 {
			pr = 1.2
		}

		// Light direction for this world points from the world toward the star (screen space; +z toward
		// the viewer for a little near-face fill). Normalize in the SQUASHED frame so the terminator sits
		// correctly on the round-in-terminal disc.
		ldx := scx - px
		ldy := (scy - py) / planetAspectY // undo the squash so the light vector matches disc coords
		llen := math.Hypot(ldx, ldy)
		if llen < 1e-6 {
			ldx, ldy, llen = 1, 0, 1
		}
		lz := 0.45
		s2 := math.Sqrt(1 - lz*lz)
		lx := ldx / llen * s2
		ly := ldy / llen * s2

		// A thin ring on the gas giant (and only there), drawn UNDER the globe's near half is complex;
		// keep it simple and legible: draw the far ring arc, then the globe, then the near ring arc.
		hasRing := kind == worldGas && hashUnit(si, 0x5AA1, seed) > 0.35
		if hasRing {
			drawPlanetRing(img, px, py, pr, false, seed^si) // far arc (behind the globe)
		}
		drawMiniGlobe(img, px, py, pr, kind, lx, ly, lz, pal, seed^(si*0x1111))
		if hasRing {
			drawPlanetRing(img, px, py, pr, true, seed^si) // near arc (in front)
		}
	}

	// ASTEROID BELT: a faint ring of tiny scattered dots between two adjacent planet orbits (squashed).
	// Pick a gap in the middle of the system so it sits visually between worlds.
	if nPlanets >= 2 {
		gap := nPlanets / 2
		if gap < 1 {
			gap = 1
		}
		inner := orbitRs[gap-1]
		outer := orbitRs[gap]
		drawAsteroidBelt(img, scx, scy, (inner+outer)*0.5, (outer-inner)*0.5, pal, seed)
	}

	// PROBES/SHIPS: a few bright streaks (a bright head + a short fading tail) heading OUTWARD from the
	// system toward the frame edges — the interstellar reach. Drawn last so they read as foreground.
	drawProbes(img, scx, scy, fmin, w, h, pal, seed)
}

// drawStar paints the central sun: a crisp sphere-shaded photosphere with procedural granulation, a
// warm white-hot-core → yellow → orange-limb gradient, a few darker sunspots, and a TIGHT bright
// limb (a 1px corona rim, NO big soft halo). Applies planetAspectY so it reads round in-terminal.
func drawStar(img *image.RGBA, cx, cy, R float64, pal tdPal, seed uint32) {
	b := img.Bounds()
	radY := R * planetAspectY
	invR := 1.0 / R
	invRadY := 1.0 / radY
	x0 := int(math.Floor(cx - R - 2))
	x1 := int(math.Ceil(cx + R + 2))
	y0 := int(math.Floor(cy - radY - 2))
	y1 := int(math.Ceil(cy + radY + 2))
	if x0 < b.Min.X {
		x0 = b.Min.X
	}
	if y0 < b.Min.Y {
		y0 = b.Min.Y
	}
	if x1 > b.Max.X {
		x1 = b.Max.X
	}
	if y1 > b.Max.Y {
		y1 = b.Max.Y
	}
	// Granulation noise scale — fine convective cells across the surface; scale off R for stability.
	gScale := 5.0 / R
	for py := y0; py < y1; py++ {
		dyf := (float64(py) - cy) * invRadY
		for px := x0; px < x1; px++ {
			dxf := (float64(px) - cx) * invR
			r2 := dxf*dxf + dyf*dyf
			if r2 > 1.0 {
				continue
			}
			edge := math.Sqrt(r2) // 0 center .. 1 limb
			// Warm gradient: white-hot core → yellow → orange limb, driven by radius.
			var base color.RGBA
			if edge < 0.55 {
				base = blend(starCore, starMid, edge/0.55)
			} else {
				base = blend(starMid, starLimb, (edge-0.55)/0.45)
			}
			// Photosphere granulation: two octaves of value-noise brighten/darken the cell texture.
			gx := (float64(px)-cx)*gScale + 40.0
			gy := (float64(py)-cy)*gScale + 40.0
			gn := valueNoise(gx, gy, seed^0x6C0F)*0.6 + valueNoise(gx*2.1+9, gy*2.1+3, seed^0x2D71)*0.4
			out := base
			if gn > 0.55 {
				out = brighten(out, (gn-0.55)*0.9)
			} else if gn < 0.42 {
				out = darken(out, (0.42-gn)*0.7)
			}
			// SUNSPOTS: a handful of darker umbral patches where a separate low-frequency noise dips very
			// low; ragged-edged so they don't read as clean circles. Kept off the very limb.
			if edge < 0.85 {
				sp := valueNoise(gx*0.6+100, gy*0.6-70, seed^0x51C3)
				if sp < 0.22 {
					sf := (0.22 - sp) / 0.22 // 0..1 depth
					out = blend(out, starSpot, clampF(sf, 0, 1)*0.85)
				}
			}
			// TIGHT bright limb: a 1–2px hot rim right at the edge (no soft outer glow).
			if edge > 0.90 {
				rim := (edge - 0.90) / 0.10
				out = blend(out, blend(starMid, starCore, 0.4), clampF(rim, 0, 1)*0.5)
			}
			setPixel(img, px, py, out)
		}
	}
	// A whisper of corona: a single faint 1px rim just OUTSIDE the disc, not a broad halo. Step the
	// squashed ellipse once at ~1.03R and dab a dim warm point.
	coron := blend(darken(starLimb, 0.35), pal.bg, 0.35)
	steps := int(R * 6)
	if steps < 48 {
		steps = 48
	}
	if steps > 1400 {
		steps = 1400
	}
	for s := 0; s < steps; s++ {
		t := float64(s) / float64(steps) * 2 * math.Pi
		ex := R * 1.04 * math.Cos(t)
		ey := R * 1.04 * math.Sin(t) * planetAspectY
		setPixel(img, int(cx+ex), int(cy+ey), coron)
	}
}

// drawMiniGlobe paints a small sphere-shaded planet at (cx,cy) with pixel radius R, shaded like the
// planet scene (per-pixel surface normal, soft terminator to a dark night limb) but with a compact
// per-KIND surface: grey cratered rock, blue/green ocean world, red mars-like, or a banded gas
// giant (horizontal latitude bands). Lit by (lx,ly,lz). Applies planetAspectY so it reads round in
// the terminal. This is the interstellar worlds' self-contained globe (kept separate from
// drawPlanetScene so that scene's byte output — and its tests — stay untouched).
func drawMiniGlobe(img *image.RGBA, cx, cy, R float64, kind planetKind, lx, ly, lz float64, pal tdPal, seed uint32) {
	if R < 1 {
		R = 1
	}
	b := img.Bounds()
	radY := R * planetAspectY
	invR := 1.0 / R
	invRadY := 1.0 / radY
	x0 := int(math.Floor(cx - R - 2))
	x1 := int(math.Ceil(cx + R + 2))
	y0 := int(math.Floor(cy - radY - 2))
	y1 := int(math.Ceil(cy + radY + 2))
	if x0 < b.Min.X {
		x0 = b.Min.X
	}
	if y0 < b.Min.Y {
		y0 = b.Min.Y
	}
	if x1 > b.Max.X {
		x1 = b.Max.X
	}
	if y1 > b.Max.Y {
		y1 = b.Max.Y
	}
	nScale := 3.4 / R
	const termSoft = 0.40
	// The ocean world picks up a faint era-accent nudge, like the planet scene's home world.
	seaLo := blend(worldSeaLo, pal.accent, 0.10)
	for py := y0; py < y1; py++ {
		dyf := (float64(py) - cy) * invRadY
		for px := x0; px < x1; px++ {
			dxf := (float64(px) - cx) * invR
			r2 := dxf*dxf + dyf*dyf
			if r2 > 1.0 {
				continue
			}
			nz := math.Sqrt(1.0 - r2)
			nx, ny := dxf, dyf
			lat := ny
			sxf := (float64(px)-cx)*nScale + 24.0
			syf := (float64(py)-cy)*nScale + 24.0

			var albedo color.RGBA
			switch kind {
			case worldGas:
				// Banded gas giant: alternate warm/dark bands by LATITUDE, wobbled by noise so the belts
				// aren't ruler-straight. No land/sea — just flowing bands.
				band := math.Sin(lat*7.0 + valueNoise(sxf*0.7, syf*0.7, seed^0x71B3)*2.4)
				albedo = blend(worldGasB, worldGasA, band*0.5+0.5)
			case worldOcean:
				n := valueNoise(sxf, syf, seed^0x9111)*0.65 +
					valueNoise(sxf*2.03+11, syf*2.03+7, seed^0x5223)*0.35
				if n+0.08*math.Abs(lat) < 0.52 {
					albedo = seaLo
				} else {
					albedo = worldSeaHi
				}
				// Small polar ice.
				if math.Abs(lat) > 0.80 {
					albedo = blend(albedo, planetIce, 0.7)
				}
			case worldMars:
				n := valueNoise(sxf, syf, seed^0x9111)*0.6 +
					valueNoise(sxf*2.1+5, syf*2.1+9, seed^0x5223)*0.4
				albedo = blend(worldRedL, worldRedH, clampF((n-0.35)/0.4, 0, 1))
				if math.Abs(lat) > 0.82 {
					albedo = blend(albedo, planetIce, 0.6) // tiny bright polar cap
				}
			default: // worldRocky
				n := valueNoise(sxf, syf, seed^0x9111)*0.6 +
					valueNoise(sxf*2.2+3, syf*2.2+6, seed^0x5223)*0.4
				albedo = blend(worldRockL, worldRockH, clampF((n-0.3)/0.45, 0, 1))
				// Crater specks: a scatter of tiny dark dots via a high-freq noise dip.
				if valueNoise(sxf*3.1+60, syf*3.1-20, seed^0x2C7D) < 0.20 {
					albedo = darken(albedo, 0.30)
				}
			}

			// SHADING: soft terminator from the star's light → dark night limb.
			ndl := nx*lx + ny*ly + nz*lz
			lit := smoothstepF(-termSoft, termSoft, ndl)
			shade := 0.05 + 0.95*lit
			out := scaleRGB(albedo, shade)
			setPixel(img, px, py, out)
		}
	}
}

// drawPlanetRing paints a thin planetary ring around a globe at (cx,cy) with globe radius R. The
// ring is an ellipse (major axis ~2.0R, tilted flat) squashed by planetAspectY; `near` selects the
// front arc (drawn over the globe) vs the far arc (drawn behind). Seeded, panic-safe.
func drawPlanetRing(img *image.RGBA, cx, cy, R float64, near bool, seed uint32) {
	ra := R * 2.05
	rb := ra * 0.34 // flattened, viewed near edge-on
	steps := int(ra * 3)
	if steps < 60 {
		steps = 60
	}
	if steps > 1200 {
		steps = 1200
	}
	col := ringColor
	for s := 0; s < steps; s++ {
		t := float64(s) / float64(steps) * 2 * math.Pi
		ey := rb * math.Sin(t)
		// Front arc = lower half (ey>0), back arc = upper half. Draw only the requested side.
		if near != (ey > 0) {
			continue
		}
		ex := ra * math.Cos(t)
		px := int(cx + ex)
		py := int(cy + ey*planetAspectY)
		c := col
		if ey > 0 {
			c = brighten(col, 0.10)
		} else {
			c = darken(col, 0.20)
		}
		setPixel(img, px, py, c)
	}
}

// drawAsteroidBelt scatters a faint ring of tiny dots around (cx,cy) at mean radius meanR with radial
// spread half, squashed by planetAspectY. Deterministic count and positions; each dot a dim 1px
// speck. Panic-safe via setPixel.
func drawAsteroidBelt(img *image.RGBA, cx, cy, meanR, half float64, pal tdPal, seed uint32) {
	if meanR <= 0 {
		return
	}
	dot := blend(pal.dim, pal.text, 0.35)
	n := int(meanR * 2.6)
	if n < 40 {
		n = 40
	}
	if n > 1200 {
		n = 1200
	}
	for i := 0; i < n; i++ {
		si := uint32(i) + 1
		ang := hashUnit(si, 0xA51F, seed) * 2 * math.Pi
		rr := meanR + (hashUnit(si, 0xB62E, seed)-0.5)*2*half
		px := cx + rr*math.Cos(ang)
		py := cy + rr*math.Sin(ang)*planetAspectY
		// Vary brightness so the belt shimmers rather than reading as a solid line.
		br := hashUnit(si, 0xC73D, seed)
		c := dot
		if br > 0.8 {
			c = brighten(dot, 0.25)
		} else if br < 0.35 {
			c = darken(dot, 0.25)
		}
		setPixel(img, int(px), int(py), c)
	}
}

// drawProbes paints 2–4 outbound probe streaks: each a bright head with a short fading tail pointing
// FROM the system OUTWARD toward a frame edge — the interstellar reach. Positions/directions seeded
// so a system's probes are stable. Panic-safe via setPixel.
func drawProbes(img *image.RGBA, scx, scy, fmin float64, w, h int, pal tdPal, seed uint32) {
	head := blend(pal.highlight, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, 0.6)
	nProbes := 2 + int(hash2(0x71B0, seed, 0x2D)%3) // 2..4
	tailLen := 0.10 * fmin
	if tailLen < 4 {
		tailLen = 4
	}
	for i := 0; i < nProbes; i++ {
		si := uint32(i) + 1
		ang := hashUnit(si, 0xD10F, seed) * 2 * math.Pi
		dx := math.Cos(ang)
		dy := math.Sin(ang) * planetAspectY // travel in the squashed frame so it matches orbits
		dl := math.Hypot(dx, dy)
		if dl < 1e-6 {
			continue
		}
		dx, dy = dx/dl, dy/dl
		// Head sits WELL outside the belt, partway toward the edge, so the probe reads as leaving.
		dist := (0.30 + 0.14*hashUnit(si, 0xE21E, seed)) * fmin
		hx := scx + dx*dist
		hy := scy + dy*dist
		// Bright 2px head.
		setPixel(img, int(hx), int(hy), head)
		setPixel(img, int(hx)+1, int(hy), blend(head, pal.highlight, 0.4))
		setPixel(img, int(hx), int(hy)+1, blend(head, pal.highlight, 0.4))
		// Fading tail pointing BACK toward the star (i.e., behind the outbound head).
		steps := int(tailLen)
		if steps < 3 {
			steps = 3
		}
		for s := 1; s <= steps; s++ {
			f := float64(s) / float64(steps) // 0 (near head) .. 1 (tail end)
			tx := hx - dx*float64(s)
			ty := hy - dy*float64(s)
			c := blend(head, pal.bg, f) // fade toward the void
			setPixel(img, int(tx), int(ty), c)
		}
	}
}

// ---------------------------------------------------------------------------------------------
// GALACTIC scene: the civilization pulled back one final level, from a single star system to its
// whole SPIRAL GALAXY — a bright dense core, logarithmic spiral arms of thousands of individual
// stars, dark dust lanes, and a few HII knots — seen at a tilt on the deep-space void. The 3rd
// cosmic scene (after the planet and the star system), sharing their craft: DEEPLY procedural,
// seed-deterministic, panic-safe (every write clipped through setPixel), NO soft blurry smears —
// the structure is carried by crisp individual star dabs, not a painted blob.
// ---------------------------------------------------------------------------------------------

// Galaxy palette anchors: a galaxy reads GOLD-CORE / BLUE-WHITE-ARMS regardless of theme (young hot
// stars populate the arms, an old warm population fills the bulge), so the tones are near-fixed with
// only a faint era-accent nudge where noted. Package-level and deterministic.
var (
	galaxyCoreHot  = color.RGBA{R: 0xff, G: 0xf3, B: 0xd6, A: 0xff} // white-gold core center
	galaxyCoreWarm = color.RGBA{R: 0xff, G: 0xcf, B: 0x82, A: 0xff} // amber bulge population
	galaxyArmStar  = color.RGBA{R: 0xdc, G: 0xe8, B: 0xff, A: 0xff} // blue-white young arm star
	galaxyArmHot   = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // rare bright standout star
	galaxyDust     = color.RGBA{R: 0x1a, G: 0x12, B: 0x1e, A: 0xff} // dark dust lane (warm-brown black)
	galaxyHII      = color.RGBA{R: 0xff, G: 0x74, B: 0x9c, A: 0xff} // pink HII / star-forming knot
)

// drawGalaxyScene renders the GALACTIC-AGE scene: a tilted spiral galaxy floating on the deep-space
// void+starfield. Like the planet and star-system scenes it abandons the city renderer — at galaxy
// scale a top-down city is meaningless.
//
// Stars are generated in the galaxy's OWN disc plane as (r, θ) polar samples, then mapped to the
// screen through: an in-plane inclination squash (the artistic tilt → the disc reads as an ellipse),
// a position-angle rotation, and finally the planetAspectY vertical squash so the whole disc (and the
// round core) read correctly in the tall terminal cell (see planetAspectY). The CORE is a dense
// gaussian cloud of warm stars over a faint underlying glow; the ARMS are 2–4 logarithmic spirals
// (r = a·e^(bθ)) walked in θ with gaussian cross-scatter, populated by blue-white stars; DUST LANES
// darken a thin band along each arm's inner edge; a few sparse pink HII knots dot the arms. Star
// count scales with canvas area and is capped, so a minimap still shows a recognizable spiral with
// fewer dabs. Deterministic and panic-safe throughout.
func drawGalaxyScene(img *image.RGBA, state game.GameState, w, h int, seed uint32) {
	if w <= 0 || h <= 0 {
		return
	}

	// BACKDROP: the shared void + a sparse far-star field. Reuse the galactic style/pal so the void tone
	// matches the rest of the cosmic-era treatment. The galaxy's own stars dominate over these.
	pal := newTdPal()
	drawSpaceBackground(img, styleForAge("galactic_age"), pal, seed, w, h)

	b := img.Bounds()
	minWH := w
	if h < minWH {
		minWH = h
	}
	fmin := float64(minWH)

	// GALAXY geometry: centered a touch off-center so the composition isn't dead-centered. Overall disc
	// radius ~0.42 of the short side (in the galaxy plane, BEFORE the tilt/aspect squashes).
	cx := float64(b.Min.X) + 0.50*float64(w)
	cy := float64(b.Min.Y) + 0.52*float64(h)
	R := 0.42 * fmin
	if R < 1 {
		R = 1
	}

	// TILT: an artistic inclination squashes the disc's minor axis so we see it at an angle (not face-on,
	// not edge-on). A seeded position-angle rotates that ellipse so different civs face differently. The
	// inclination is applied IN-PLANE (before rotation); planetAspectY is applied AFTER (terminal squash).
	incl := 0.50 + 0.14*hashUnit(0x6A1A, 0x11, seed) // minor-axis fraction ~0.50..0.64 (moderate tilt)
	rot := hashUnit(0x6A1A, 0x22, seed) * math.Pi    // disc position angle
	cosR, sinR := math.Cos(rot), math.Sin(rot)

	// project maps a point (gr,gth) in the galaxy DISC plane (polar) to integer screen pixels, applying
	// the inclination squash, the position-angle rotation, and finally planetAspectY. Returns the pixel
	// plus the mapped float coords for any caller that wants sub-pixel work. Kept as a closure so the
	// core, the arms, and the dust all share one identical transform.
	project := func(gr, gth float64) (int, int) {
		gx := gr * math.Cos(gth)
		gy := gr * math.Sin(gth) * incl // inclination: flatten the minor axis in-plane
		// Position-angle rotation of the (already-inclined) ellipse.
		rx := gx*cosR - gy*sinR
		ry := gx*sinR + gy*cosR
		px := cx + rx
		py := cy + ry*planetAspectY // terminal squash so the disc/core read round in the cell
		return int(px), int(py)
	}

	// dab paints a star: a 1px core plus, for the brighter ones, a small NE/SW cross so a few stars read
	// as standouts. Alpha-blend-free (opaque write) via setPixel — crisp, no soft halo. size 0 = 1px,
	// size 1 = a 5px plus, size 2 = a brighter 5px plus with lit neighbours.
	dab := func(px, py int, c color.RGBA, size int) {
		setPixel(img, px, py, c)
		if size >= 1 {
			dim := blend(c, pal.bg, 0.45)
			setPixel(img, px+1, py, dim)
			setPixel(img, px-1, py, dim)
			setPixel(img, px, py+1, dim)
			setPixel(img, px, py-1, dim)
		}
		if size >= 2 {
			setPixel(img, px+1, py, blend(c, galaxyArmHot, 0.3))
			setPixel(img, px-1, py, blend(c, galaxyArmHot, 0.3))
			setPixel(img, px, py+1, blend(c, galaxyArmHot, 0.3))
			setPixel(img, px, py-1, blend(c, galaxyArmHot, 0.3))
			setPixel(img, px+1, py+1, blend(c, pal.bg, 0.5))
			setPixel(img, px-1, py-1, blend(c, pal.bg, 0.5))
		}
	}

	// A faint theme-accent nudge for the arm stars so the galaxy picks up a whisper of the era mood
	// without ceasing to read as blue-white.
	armStar := blend(galaxyArmStar, pal.accent, 0.10)

	// ------------------------------------------------------------------------------------------
	// UNDERGLOW: a very faint radial bulge glow UNDER the core stars — subtle, so the stars carry it
	// (NOT a solid blur). Painted only within the inner disc, additively toward the warm core, with a
	// steep falloff. Iterated over the elliptical bounding box of the inner region, clipped.
	// ------------------------------------------------------------------------------------------
	glowR := R * 0.34 // the bulge glow reaches ~a third of the disc
	{
		radY := glowR * planetAspectY
		x0 := int(math.Floor(cx - glowR - 2))
		x1 := int(math.Ceil(cx + glowR + 2))
		y0 := int(math.Floor(cy - radY - 2))
		y1 := int(math.Ceil(cy + radY + 2))
		if x0 < b.Min.X {
			x0 = b.Min.X
		}
		if y0 < b.Min.Y {
			y0 = b.Min.Y
		}
		if x1 > b.Max.X {
			x1 = b.Max.X
		}
		if y1 > b.Max.Y {
			y1 = b.Max.Y
		}
		invGR := 1.0 / glowR
		invGRadY := 1.0 / radY
		for py := y0; py < y1; py++ {
			dyf := (float64(py) - cy) * invGRadY
			for px := x0; px < x1; px++ {
				dxf := (float64(px) - cx) * invGR
				d2 := dxf*dxf + dyf*dyf
				if d2 > 1.0 {
					continue
				}
				f := 1 - math.Sqrt(d2) // 1 center → 0 edge
				f = f * f * f          // steep falloff so the glow stays tight to the center
				if f <= 0.002 {
					continue
				}
				cur := img.RGBAAt(px, py)
				setPixel(img, px, py, blend(cur, galaxyCoreWarm, clampF(f*0.55, 0, 1)))
			}
		}
	}

	// ------------------------------------------------------------------------------------------
	// SPIRAL ARMS + DUST LANES: 2..4 logarithmic-spiral arms, evenly phased. Each arm is walked in θ;
	// at each step we place a small cluster of stars scattered around the arm centerline with a
	// gaussian-ish perpendicular jitter (denser on the spine, sparse off it). A thin dust band is
	// darkened just INSIDE each arm (toward smaller r) for depth. Star budget scales with area, capped.
	// ------------------------------------------------------------------------------------------
	nArms := 2 + int(hash2(0x6A1A, seed, 0x33)%3) // 2..4 arms
	// b (spiral tightness): larger → looser winding. a: base radius so the arm starts just outside the
	// core. Both seeded a hair so the winding differs per civ.
	spiralB := 0.24 + 0.10*hashUnit(0x6A1A, 0x44, seed) // pitch
	spiralA := R * 0.10
	if spiralA < 0.5 {
		spiralA = 0.5
	}
	// θ sweeps from the core outward until r reaches the disc edge: r = a·e^(b·θ) → θmax = ln(R/a)/b.
	thetaMax := math.Log(R/spiralA) / spiralB
	if thetaMax < 0.5 {
		thetaMax = 0.5
	}
	if thetaMax > 7.0 { // ~1.1 turns cap so arms don't over-wind into mush
		thetaMax = 7.0
	}
	// Angular step + per-step star cluster size both scale with canvas area so a minimap stays sparse but
	// a full render is dense. Total dabs are implicitly capped by the θ range × cluster size × arms.
	areaScale := float64(w*h) / (440.0 * 300.0) // 1.0 at the reference dump size
	if areaScale > 1.6 {
		areaScale = 1.6
	}
	dTheta := 0.05 / math.Max(0.35, areaScale) // finer stepping on bigger canvases
	clusterN := int(3 + 5*areaScale)           // stars scattered per θ step per arm
	if clusterN < 2 {
		clusterN = 2
	}
	armWidth := R * 0.11 // base perpendicular scatter width (grows slightly with r below)

	for a := 0; a < nArms; a++ {
		armPhase := float64(a) / float64(nArms) * 2 * math.Pi
		armSeed := seed ^ (uint32(a+1) * 0x9E37)
		step := 0
		for th := 0.15; th < thetaMax; th += dTheta {
			step++
			r := spiralA * math.Exp(spiralB*th)
			if r > R {
				break
			}
			baseTh := th + armPhase
			// DUST LANE: a thin dark band just inside the arm spine (toward the core), placed a fraction of
			// the arm width inward. Only from mid-disc outward (the core swamps any inner dust). Sparse dabs
			// so it threads rather than paints a solid ring.
			if r > R*0.16 && (step%2 == 0) {
				dOff := -armWidth * (0.55 + 0.25*hashUnit(uint32(step), 0xD0, armSeed))
				dr := r + dOff
				if dr > 0 {
					// perpendicular offset ≈ tangential displacement / r in angle terms
					dpx, dpy := project(dr, baseTh)
					setPixel(img, dpx, dpy, blend(img.RGBAAt(dpx, dpy), galaxyDust, 0.55))
					setPixel(img, dpx, dpy+1, blend(img.RGBAAt(dpx, dpy+1), galaxyDust, 0.35))
				}
			}
			// Arm width grows a little with radius (arms fan out), then density thins outward.
			wHere := armWidth * (0.7 + 0.9*(r/R))
			// Outward density taper: fewer stars far out so the disc fades at the rim.
			taper := 1.0 - 0.55*(r/R)
			nHere := int(float64(clusterN) * (0.6 + 0.4*taper))
			if nHere < 1 {
				nHere = 1
			}
			for k := 0; k < nHere; k++ {
				kk := uint32(step*31 + k*7 + 1)
				// Gaussian-ish perpendicular jitter: average two uniforms centered at 0 → a soft triangular
				// bump (denser on the spine). Convert the linear offset to an angular offset (/r).
				j1 := hashUnit(kk, 0xA0, armSeed) - 0.5
				j2 := hashUnit(kk, 0xB0, armSeed) - 0.5
				perp := (j1 + j2) * wHere // -w..w, peaked at 0
				// A little jitter along the arm too so steps don't quantize into visible rings.
				alongJ := (hashUnit(kk, 0xC0, armSeed) - 0.5) * dTheta * 1.4
				rr := r + perp // perpendicular scatter enters as a radial offset (cross-arm width)
				if rr <= 0 {
					continue
				}
				px, py := project(rr, baseTh+alongJ)
				// Star color: bluer on the spine, warming slightly toward the core; brightness varies.
				spineF := 1.0 - math.Min(1.0, math.Abs(perp)/wHere) // 1 on spine → 0 at edge
				coreMix := clampF(1.0-r/(R*0.5), 0, 1) * 0.5        // warm the inner arm toward the bulge
				col := blend(armStar, galaxyCoreWarm, coreMix)
				br := hashUnit(kk, 0xE0, armSeed)
				size := 0
				switch {
				case br > 0.985: // rare bright standout
					col = blend(col, galaxyArmHot, 0.7)
					size = 2
				case br > 0.90:
					col = brighten(col, 0.10)
					size = 1
				case br < 0.30:
					// dim field star: pull toward the void so the arm has depth, skip the very dimmest
					// off-spine so between-arm gaps stay dark.
					if spineF < 0.35 {
						continue
					}
					col = blend(col, pal.bg, 0.45)
				}
				dab(px, py, col, size)
			}
			// HII REGION: a rare pink knot on the arm (a tiny crisp cluster, not a glow), sparse.
			if hashUnit(uint32(step), 0xF0, armSeed) > 0.965 && r > R*0.20 {
				hpx, hpy := project(r, baseTh)
				dab(hpx, hpy, galaxyHII, 1)
				setPixel(img, hpx, hpy, blend(galaxyHII, galaxyArmHot, 0.4))
			}
		}
	}

	// ------------------------------------------------------------------------------------------
	// CORE / BULGE: a dense gaussian cloud of WARM stars over the underglow, brightest dead center.
	// Radial density falls off gaussianly so the bulge is tightly packed at the center and thins into
	// the arms. Placed LAST so the bright core sits over the arm stars that pass behind it. Count scales
	// with area and is capped.
	// ------------------------------------------------------------------------------------------
	coreR := R * 0.30
	coreN := int(700 * areaScale)
	if coreN > 1400 {
		coreN = 1400
	}
	if coreN < 60 {
		coreN = 60
	}
	for i := 0; i < coreN; i++ {
		si := uint32(i) + 1
		// Radius: gaussian-ish via averaging two uniforms then squaring → strong central concentration.
		u1 := hashUnit(si, 0x10, seed^0xC0DE)
		u2 := hashUnit(si, 0x20, seed^0xC0DE)
		rr := ((u1 + u2) * 0.5)
		rr = rr * rr * coreR // bias hard toward the center
		ang := hashUnit(si, 0x30, seed^0xC0DE) * 2 * math.Pi
		px, py := project(rr, ang)
		// Center-weighted color + brightness: white-gold at the very center → amber outward.
		cf := clampF(1.0-rr/coreR, 0, 1)
		col := blend(galaxyCoreWarm, galaxyCoreHot, cf*cf)
		br := hashUnit(si, 0x40, seed^0xC0DE)
		size := 0
		if br > 0.97 || rr < coreR*0.10 {
			size = 1
			col = brighten(col, 0.06)
		}
		if br > 0.995 {
			size = 2
		}
		dab(px, py, col, size)
	}
	// A tiny brilliant nucleus dab dead center so the very heart reads hottest.
	ncx, ncy := project(0, 0)
	setPixel(img, ncx, ncy, galaxyCoreHot)
	setPixel(img, ncx+1, ncy, blend(galaxyCoreHot, galaxyCoreWarm, 0.4))
	setPixel(img, ncx-1, ncy, blend(galaxyCoreHot, galaxyCoreWarm, 0.4))
	setPixel(img, ncx, ncy+1, blend(galaxyCoreHot, galaxyCoreWarm, 0.4))
	setPixel(img, ncx, ncy-1, blend(galaxyCoreHot, galaxyCoreWarm, 0.4))
}

// Cosmic-web palette anchors: at the LARGEST scale the universe is a filamentary web of galaxy
// CLUSTERS strung on threads of dark matter, with vast empty VOIDS between. The quantum lens tints
// the whole field with an IRIDESCENT sheen (cyan↔magenta↔gold cycling across position). The three
// iridescence anchors already exist package-level (iridCyanAnchor / iridMagentaAnchor /
// iridGoldAnchor); these are the web-specific structural tones. Package-level and deterministic.
var (
	webNodeCore = color.RGBA{R: 0xff, G: 0xfb, B: 0xf2, A: 0xff} // brilliant cluster-core dab (near-white, faintly warm)
	webGalaxy   = color.RGBA{R: 0xe8, G: 0xdc, B: 0xff, A: 0xff} // a single tiny galaxy in a cluster (cool pale)
	webFarGal   = color.RGBA{R: 0x6c, G: 0x60, B: 0x8c, A: 0xff} // very faint far-background galaxy speck in the void
)

// webIrid returns the iridescent hue for a point (ux,uy) in unit field coordinates [0,1]. The three
// quantum anchors (cyan→magenta→gold) are cycled by a smooth diagonal phase across the field so the
// web's clusters shift hue by position — the crystalline "quantum lens" — while staying crisp (this
// is a pure color pick, not a blur). Pure helper, no locks.
func webIrid(ux, uy float64) color.RGBA {
	// Phase sweeps ~1.3 full cycles across the diagonal so both ends of the field differ in hue.
	ph := (ux*0.62 + uy*0.38) * 1.3
	ph -= math.Floor(ph) // wrap into [0,1)
	seg := ph * 3.0      // 0..3 across the three anchors
	i := int(seg)
	f := seg - float64(i)
	f = f * f * (3 - 2*f) // smoothstep the crossfade so bands don't hard-edge
	return blend(iridHueFor(i), iridHueFor(i+1), f)
}

// drawCosmicWebScene renders the QUANTUM-AGE scene: the COSMIC WEB — the largest-scale structure of
// the universe, one zoom-out beyond the galactic spiral. Bright galaxy CLUSTERS sit as knots on a
// network of glowing FILAMENTS, separated by dark intergalactic VOIDS, the whole field washed in a
// shifting IRIDESCENT (quantum) sheen. Like the other cosmic scenes it abandons the city renderer —
// at this scale a top-down city is meaningless.
//
// Structure is built from real geometry, not a vague glow: a handful of seeded ATTRACTOR points pull
// ~10–24 cluster NODES into groups (so empty voids emerge naturally between them); each node is a
// tight knot of tiny galaxy dabs around a brilliant core, hued by webIrid at its position. Each node
// links to its 2–3 NEAREST neighbours with a wavy FILAMENT — a thread of faint dots that brightens
// toward the endpoints, with a few tiny galaxies strung along it; the filament hue interpolates
// between its two endpoints' iridescent hues. A very faint large-scale web glow sits UNDER the
// structure for depth. Node/galaxy counts scale with canvas area and are capped, so a minimap still
// shows a recognizable (sparser) web. Deterministic and panic-safe throughout (every write clips).
func drawCosmicWebScene(img *image.RGBA, state game.GameState, w, h int, seed uint32) {
	if w <= 0 || h <= 0 {
		return
	}
	_ = state // scene depends only on the (name-derived) seed, like the other cosmic scenes

	// BACKDROP: a deep intergalactic VOID — darker/emptier than the galaxy backdrop. Reuse the shared
	// space background (void + very sparse far-star field) with the quantum style so the void tone
	// matches the era; the web's own nodes/filaments dominate over it. Then darken it a touch further
	// so the intergalactic gaps read as truly empty next to the galaxy scene.
	pal := newTdPal()
	drawSpaceBackground(img, styleForAge("quantum_age"), pal, seed, w, h)

	b := img.Bounds()
	minWH := w
	if h < minWH {
		minWH = h
	}
	fmin := float64(minWH)
	fw, fh := float64(w), float64(h)

	// Pull the whole void a shade deeper toward black so intergalactic space reads emptier than the
	// galaxy backdrop. A cheap single pass over the existing pixels (no new gradient).
	for y := 0; y < h; y++ {
		py := b.Min.Y + y
		for x := 0; x < w; x++ {
			px := b.Min.X + x
			img.SetRGBA(px, py, darken(img.RGBAAt(px, py), 0.22))
		}
	}

	// A few very faint FAR-BACKGROUND galaxies scattered through the void (distinct from the star
	// field: dim purple-grey specks, occasionally a 2px smudge) — depth cues in the emptiness.
	farN := (w * h) / 900
	for i := 0; i < farN; i++ {
		si := uint32(i) + 1
		fx := int(hash2(si, 0x7A, seed) % uint32(w))
		fy := int(hash2(si, 0x8B, seed) % uint32(h))
		amt := hashUnit(si, 0x9C, seed)
		if amt < 0.5 {
			continue // keep them sparse
		}
		c := blend(img.RGBAAt(b.Min.X+fx, b.Min.Y+fy), webFarGal, 0.25+0.30*amt)
		setPixel(img, b.Min.X+fx, b.Min.Y+fy, c)
		if amt > 0.9 { // rare tiny elongated far-galaxy
			setPixel(img, b.Min.X+fx+1, b.Min.Y+fy, blend(c, webFarGal, 0.4))
		}
	}

	// ------------------------------------------------------------------------------------------
	// NODES (galaxy clusters): scatter cluster nodes CLUSTERED around a few attractor points so the
	// dark voids fall out naturally. Node count scales with canvas area, clamped to ~10..24 at the
	// full dump size (fewer on a minimap). Each node stores its screen position, its iridescent hue,
	// and a size in galaxy-dabs.
	// ------------------------------------------------------------------------------------------
	areaScale := fw * fh / (440.0 * 300.0) // 1.0 at the reference dump size
	if areaScale > 1.6 {
		areaScale = 1.6
	}
	nNodes := int(14 * areaScale)
	if nNodes > 24 {
		nNodes = 24
	}
	if nNodes < 5 { // a minimap still shows a recognizable (small) web
		nNodes = 5
	}
	// Attractors: 2..4 pull-points, placed with a margin so clusters sit inside the frame. Nodes are
	// scattered on a gaussian-ish spread around a randomly chosen attractor → groups + voids.
	nAttr := 2 + int(hash2(0xCEB, seed, 0x01)%3) // 2..4
	type pt struct{ x, y float64 }
	attr := make([]pt, nAttr)
	for i := 0; i < nAttr; i++ {
		attr[i] = pt{
			x: (0.16 + 0.68*hashUnit(uint32(i)+1, 0x11, seed)) * fw,
			y: (0.16 + 0.68*hashUnit(uint32(i)+1, 0x22, seed)) * fh,
		}
	}
	type node struct {
		x, y float64
		hue  color.RGBA
		size int // galaxy dabs in the knot
	}
	nodes := make([]node, 0, nNodes)
	spread := 0.16 * fmin // cluster tightness around an attractor
	for i := 0; i < nNodes; i++ {
		si := uint32(i) + 1
		a := attr[int(hash2(si, 0x33, seed)%uint32(nAttr))]
		// Two averaged uniforms → a soft central bump (denser near the attractor); a random angle.
		rr := (hashUnit(si, 0x44, seed) + hashUnit(si, 0x45, seed)) * 0.5
		ang := hashUnit(si, 0x55, seed) * 2 * math.Pi
		x := a.x + math.Cos(ang)*rr*spread
		y := a.y + math.Sin(ang)*rr*spread*planetAspectY // squash the scatter so clusters read round
		// Keep inside the frame with a small margin.
		x = clampF(x, 4, fw-5)
		y = clampF(y, 4, fh-5)
		hue := webIrid(x/fw, y/fh)
		sz := 4 + int(hashUnit(si, 0x66, seed)*6) // 4..9 galaxy dabs
		nodes = append(nodes, node{x: x, y: y, hue: hue, size: sz})
	}

	// ------------------------------------------------------------------------------------------
	// UNDER-GLOW: a very faint large-scale web glow beneath the structure — a soft brightening where
	// nodes cluster, so the field has depth. Kept low intensity (the crisp nodes/filaments carry the
	// scene). Accumulated per node with a steep falloff, clipped to a small bounding box each.
	// ------------------------------------------------------------------------------------------
	glowR := 0.15 * fmin
	if glowR >= 2 {
		radY := glowR * planetAspectY
		for _, n := range nodes {
			gTint := blend(n.hue, webNodeCore, 0.25)
			x0 := int(math.Floor(n.x - glowR))
			x1 := int(math.Ceil(n.x + glowR))
			y0 := int(math.Floor(n.y - radY))
			y1 := int(math.Ceil(n.y + radY))
			if x0 < b.Min.X {
				x0 = b.Min.X
			}
			if y0 < b.Min.Y {
				y0 = b.Min.Y
			}
			if x1 > b.Max.X {
				x1 = b.Max.X
			}
			if y1 > b.Max.Y {
				y1 = b.Max.Y
			}
			invR := 1.0 / glowR
			invRadY := 1.0 / radY
			for py := y0; py < y1; py++ {
				dyf := (float64(py) - n.y) * invRadY
				for px := x0; px < x1; px++ {
					dxf := (float64(px) - n.x) * invR
					d2 := dxf*dxf + dyf*dyf
					if d2 > 1.0 {
						continue
					}
					f := 1 - math.Sqrt(d2)
					f = f * f * f // steep — stays a whisper, not a smudge
					if f <= 0.003 {
						continue
					}
					setPixel(img, px, py, blend(img.RGBAAt(px, py), gTint, clampF(f*0.11, 0, 1)))
				}
			}
		}
	}

	// tinyGalaxy stamps a single faint galaxy dab: a 1px core tinted toward the iridescent hue, with
	// (for the brighter ones) a faint neighbour so a few read as small smudges. Crisp opaque writes.
	tinyGalaxy := func(px, py int, hue color.RGBA, bright float64) {
		c := blend(webGalaxy, hue, 0.45)
		c = blend(c, webNodeCore, 0.25*bright)
		setPixel(img, px, py, c)
		if bright > 0.72 {
			dim := blend(c, pal.bg, 0.5)
			setPixel(img, px+1, py, dim)
			setPixel(img, px, py+1, dim)
		}
	}

	// ------------------------------------------------------------------------------------------
	// FILAMENTS: connect each node to its 2..3 NEAREST neighbour nodes with a glowing thread. To
	// avoid drawing every edge twice, only draw i→j when j>i (the pairing is symmetric anyway). Each
	// filament is a slightly WAVY line of faint dots (a low-amplitude sine perpendicular to the run),
	// BRIGHTENING toward the two node endpoints, with a few tiny galaxies strung along it. The dot
	// hue lerps between the two endpoints' iridescent hues along the run.
	// ------------------------------------------------------------------------------------------
	for i := range nodes {
		ni := nodes[i]
		// Rank the other nodes by distance to ni; keep the nearest 2..3.
		type nb struct {
			j int
			d float64
		}
		cand := make([]nb, 0, len(nodes)-1)
		for j := range nodes {
			if j == i {
				continue
			}
			dx := nodes[j].x - ni.x
			dy := nodes[j].y - ni.y
			cand = append(cand, nb{j: j, d: dx*dx + dy*dy})
		}
		sort.Slice(cand, func(a, b int) bool { return cand[a].d < cand[b].d })
		links := 2 + int(hash2(uint32(i)+1, seed, 0x77)%2) // 2..3 nearest neighbours
		if links > len(cand) {
			links = len(cand)
		}
		for c := 0; c < links; c++ {
			j := cand[c].j
			if j <= i {
				continue // draw each undirected edge once
			}
			nj := nodes[j]
			dx := nj.x - ni.x
			dy := nj.y - ni.y
			L := math.Hypot(dx, dy)
			if L < 1 {
				continue
			}
			// Unit direction + a perpendicular for the waviness.
			ux, uy := dx/L, dy/L
			perpx, perpy := -uy, ux
			// Filament seed keys the wave phase/amplitude so each thread wobbles differently.
			fseed := seed ^ (uint32(i+1) * 0x9E37) ^ (uint32(j+1) * 0x85EB)
			waveAmp := (0.02 + 0.03*hashUnit(fseed, 0x01, seed)) * fmin // gentle, ~2..5% of the short side
			waveK := 1.5 + 2.5*hashUnit(fseed, 0x02, seed)              // ~1.5..4 humps along the run
			wavePh := hashUnit(fseed, 0x03, seed) * 2 * math.Pi
			// Step ~1px along the run so the thread is continuous but crisp.
			steps := int(L)
			if steps < 1 {
				steps = 1
			}
			for s := 0; s <= steps; s++ {
				t := float64(s) / float64(steps) // 0 at ni → 1 at nj
				// Waviness: a sine bump that vanishes at both endpoints (so it meets the nodes cleanly).
				env := math.Sin(t * math.Pi) // 0 at ends, 1 mid
				off := waveAmp * env * math.Sin(t*waveK*math.Pi+wavePh)
				fx := ni.x + ux*L*t + perpx*off
				fy := ni.y + uy*L*t + perpy*off
				px := int(fx + 0.5)
				py := int(fy + 0.5)
				// Brightness: brightens toward BOTH endpoints (a shallow U), dimmest mid-run.
				edge := 1 - env            // 1 at ends → 0 mid
				bright := 0.30 + 0.55*edge // ~0.30 mid .. 0.85 near nodes
				hue := blend(ni.hue, nj.hue, t)
				thread := blend(pal.bg, hue, 0.35+0.45*edge)
				thread = blend(thread, webNodeCore, 0.10*bright)
				// Skip a fraction of mid-run dots so the thread reads as a faint dotted filament, not a
				// solid bright line — keeps the voids feeling empty and the nodes dominant.
				if edge < 0.35 && (hashUnit(uint32(s+1), 0x0D, fseed) > 0.6) {
					continue
				}
				setPixel(img, px, py, thread)
			}
			// String a few TINY galaxies along the filament (not on the endpoints).
			galN := 1 + int(hashUnit(fseed, 0x04, seed)*3) // 1..3
			for g := 0; g < galN; g++ {
				t := 0.2 + 0.6*hashUnit(uint32(g)+1, 0x05, fseed) // keep off the ends
				env := math.Sin(t * math.Pi)
				off := waveAmp * env * math.Sin(t*waveK*math.Pi+wavePh)
				fx := ni.x + ux*L*t + perpx*off
				fy := ni.y + uy*L*t + perpy*off
				hue := blend(ni.hue, nj.hue, t)
				tinyGalaxy(int(fx+0.5), int(fy+0.5), hue, 0.4+0.4*hashUnit(uint32(g)+1, 0x06, fseed))
			}
		}
	}

	// ------------------------------------------------------------------------------------------
	// NODE KNOTS: draw the cluster nodes LAST so their bright cores sit over any filament dots that
	// pass behind them. Each node is a tight scatter of tiny galaxies around a brilliant core dab,
	// all in the node's iridescent hue. The core is a small round knot (aspect-squashed so it reads
	// round in the terminal cell).
	// ------------------------------------------------------------------------------------------
	for _, n := range nodes {
		cx, cy := n.x, n.y
		knotR := (2.0 + 0.5*float64(n.size)) // scatter radius grows a hair with the cluster size
		// Tiny galaxies scattered in the knot (gaussian-ish toward the center).
		for k := 0; k < n.size; k++ {
			kk := uint32(k)*31 + 1
			rr := (hashUnit(kk, 0x01, seed) + hashUnit(kk, 0x02, seed)) * 0.5
			rr = rr * rr * knotR // bias toward the center
			ang := hashUnit(kk, 0x03, seed) * 2 * math.Pi
			gx := int(cx + math.Cos(ang)*rr + 0.5)
			gy := int(cy + math.Sin(ang)*rr*planetAspectY + 0.5) // squash so the knot reads round
			tinyGalaxy(gx, gy, n.hue, 0.5+0.5*hashUnit(kk, 0x04, seed))
		}
		// A small round bright CORE dab (aspect-squashed). A tight filled ellipse: core-white at the
		// very center fading into the node's iridescent hue at the rim — crisp, no soft halo.
		coreR := 1.6 + 0.15*float64(n.size)
		coreRY := coreR * planetAspectY
		x0 := int(math.Floor(cx - coreR))
		x1 := int(math.Ceil(cx + coreR))
		y0 := int(math.Floor(cy - coreRY))
		y1 := int(math.Ceil(cy + coreRY))
		invCR := 1.0 / coreR
		invCRY := 1.0 / math.Max(coreRY, 0.001)
		for py := y0; py <= y1; py++ {
			dyf := (float64(py) - cy) * invCRY
			for px := x0; px <= x1; px++ {
				dxf := (float64(px) - cx) * invCR
				d2 := dxf*dxf + dyf*dyf
				if d2 > 1.0 {
					continue
				}
				f := 1 - d2 // 1 center → 0 rim
				col := blend(n.hue, webNodeCore, clampF(f*f, 0, 1))
				setPixel(img, px, py, col)
			}
		}
	}
}

// bi returns 1 if b else 0 — a tiny helper for offsetting the station's larger footprint.
func bi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// clampF clamps v to [lo,hi].
func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// smoothstepF returns the smoothstep of x in [edge0,edge1] → [0,1].
func smoothstepF(edge0, edge1, x float64) float64 {
	if edge1 <= edge0 {
		if x < edge0 {
			return 0
		}
		return 1
	}
	t := clampF((x-edge0)/(edge1-edge0), 0, 1)
	return t * t * (3 - 2*t)
}

// scaleRGB multiplies an RGB color's channels by s (clamped to [0,1] per channel), preserving
// alpha. Used to darken the planet toward its night side without shifting hue.
func scaleRGB(c color.RGBA, s float64) color.RGBA {
	if s < 0 {
		s = 0
	}
	return color.RGBA{
		R: uint8(clampF(float64(c.R)*s, 0, 255)),
		G: uint8(clampF(float64(c.G)*s, 0, 255)),
		B: uint8(clampF(float64(c.B)*s, 0, 255)),
		A: c.A,
	}
}

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
	case tdPropMegalith:
		// A single TALL standing stone: a vertical grey slab, wider at the base, with a soft ground
		// shadow — clearly larger/taller than the 3-dab tdPropStones. Grey stone tones blended with
		// propCol/pavedCol like the neighbouring props, so it retints with the theme.
		stone := blend(blend(prop, paved, 0.4), graniteAnchor, 0.35)
		shadow := darken(stone, 0.45)
		// Ground shadow: a low dark ellipse at the base (south of the slab).
		forEllipse(cx, cy+rad, maxInt(rad, 1), maxInt(rad/2, 1), func(x, y int) { blendPixel(img, x, y, shadow, 0.35) })
		// The slab: a tall vertical rect (taller than wide), a touch wider at the base.
		fillRectC(img, cx, cy, maxInt(rad/2, 0), rad, stone)
		fillRectC(img, cx, cy+rad, maxInt(rad, 1), 0, blend(stone, shadow, 0.4)) // splayed base
		// Lit NW face + shaded SE edge for upright volume; a bright crown dab on top.
		for dy := -rad; dy <= rad; dy++ {
			setPixel(img, cx-maxInt(rad/2, 0), cy+dy, brighten(stone, 0.12))
			setPixel(img, cx+maxInt(rad/2, 0), cy+dy, darken(stone, 0.15))
		}
		setPixel(img, cx, cy-rad, brighten(stone, 0.16))
	case tdPropStall:
		cloth := blend(prop, pal.text, 0.20)
		fillRectC(img, cx, cy, rad, rad, cloth)
		drawHSpan(img, cx-rad, cx+rad, cy-rad, brighten(cloth, 0.16))
	case tdPropAltar:
		// A low stone altar: a flat pale slab with a small dark offering dab on top.
		stone := blend(prop, paved, 0.45)
		fillRectC(img, cx, cy, rad, maxInt(rad-1, 0), brighten(stone, 0.08))
		setPixel(img, cx, cy, darken(prop, 0.4))
	case tdPropColumns:
		// A colonnade: a short row of upright pale column dabs.
		col := brighten(blend(prop, paved, 0.5), 0.10)
		for _, dx := range []int{-rad, 0, rad} {
			drawBlock(img, cx+dx, cy, 0, col)
			setPixel(img, cx+dx, cy-1, brighten(col, 0.12)) // a lit capital
		}
	case tdPropBrazier:
		// A fire brazier: a dark stand under a bright ember.
		fillDisc(img, cx, cy, rad, darken(prop, 0.5))
		ember := brighten(blend(prop, color.RGBA{R: 0xd0, G: 0x6a, B: 0x24, A: 0xff}, 0.7), 0.12)
		setPixel(img, cx, cy, ember)
		setPixel(img, cx, cy-1, ember)
	case tdPropFountain:
		// A stone fountain: a paved ring with a water-tone center.
		fillDisc(img, cx, cy, rad, brighten(paved, 0.08))
		water := blend(pal.bg, waterAnchor, 0.55)
		setPixel(img, cx, cy, water)
		setPixel(img, cx, cy-1, brighten(water, 0.14))
	case tdPropCross:
		// A market cross / gallows: an upright post with a crossbar.
		post := blend(prop, paved, 0.30)
		for dy := -rad; dy <= rad; dy++ {
			setPixel(img, cx, cy+dy, post)
		}
		drawHSpan(img, cx-rad, cx+rad, cy-rad+1, post)
	case tdPropSmokestack:
		// A single TALL factory chimney: a dark vertical column rising well ABOVE the other props (a
		// scatter smokestack seasoning the skyline), with a lit NW rim, a splayed base, and a soft dark
		// soot dab drifting off its top. Dark soot tones blended with propCol/pavedCol like the
		// neighbouring props so it retints with the theme.
		stack := blend(blend(prop, paved, 0.30), sootAnchor, 0.50)
		stackLit := brighten(stack, 0.16)
		soot := blend(stack, sootAnchor, 0.40)
		tall := rad * 2 // taller than a normal prop (which spans ±rad)
		// Ground shadow at the base (south of the column).
		forEllipse(cx, cy+rad, maxInt(rad, 1), maxInt(rad/2, 1), func(x, y int) { blendPixel(img, x, y, darken(stack, 0.4), 0.35) })
		// The column: a thin tall vertical rect from base (cy+rad) up to the top (cy-tall).
		fillRectC(img, cx, (cy+rad-tall)/2, maxInt(rad/3, 0), (cy+rad+tall)/2, stack)
		fillRectC(img, cx, cy+rad, maxInt(rad/2, 1), 0, blend(stack, darken(stack, 0.4), 0.4)) // splayed base
		for dy := -tall; dy <= rad; dy++ {
			setPixel(img, cx-maxInt(rad/3, 0), cy+dy, stackLit)            // lit NW edge
			setPixel(img, cx+maxInt(rad/3, 0), cy+dy, darken(stack, 0.16)) // shaded SE edge
		}
		// Soot dab drifting off the chimney top.
		puffR := maxInt(rad/2, 1)
		forEllipse(cx+puffR, cy-tall-puffR/2, puffR, puffR, func(x, y int) { blendPixel(img, x, y, soot, 0.55) })
		setPixel(img, cx, cy-tall, brighten(stack, 0.12)) // lit chimney lip
	case tdPropGasLamp:
		// A genteel VICTORIAN gas-lamp: a short dark iron lamp-post (a thin vertical stem, taller than a
		// plain prop but well short of a smokestack) capped by a warm amber GLOW dab — a soft flame-gold
		// halo, blended (never raw) so it reads as gaslight, not a saturated accent. Dark iron tones pulled
		// toward propCol/pavedCol like the neighbours, so it retints with the theme. Bounds-safe: setPixel /
		// blendPixel / forEllipse only.
		post := darken(blend(prop, paved, 0.20), 0.30) // dark iron stem
		tall := rad + rad/2                            // a short post — above props, below smokestacks
		// The stem: a thin vertical column from base (cy+rad) up to the lamp head (cy-tall).
		for dy := -tall; dy <= rad; dy++ {
			setPixel(img, cx, cy+dy, post)
		}
		// The glow: a small warm amber halo at the lamp head, a soft lit core over a dimmer ring.
		glow := blend(brighten(prop, 0.10), gasGlowAnchor, 0.60)
		glowR := maxInt(rad/2, 1)
		forEllipse(cx, cy-tall, glowR, glowR, func(x, y int) { blendPixel(img, x, y, glow, 0.55) })
		setPixel(img, cx, cy-tall, brighten(glow, 0.18)) // bright flame core
	case tdPropDataCenter:
		// A low WIDE server-farm block: a flat cool-grey slab a touch wider than tall (a data center reads
		// long and low, unlike the tall smokestack/lamp), lit N/W + shaded S/E, with a couple of tiny
		// BLINKING-LIGHT dabs (cyan + amber) on its face — the server-rack LEDs. Cool data-grey tones pulled
		// toward propCol/pavedCol so it retints with the theme. Bounds-safe: fillRectC / setPixel / blendPixel.
		slab := blend(blend(prop, paved, 0.30), dataGreyAnchor, 0.50)
		slabLit := brighten(slab, 0.12)
		slabDark := darken(slab, 0.20)
		bhw := maxInt(rad, 1)   // wide
		bhh := maxInt(rad/2, 1) // low
		forRect(cx, cy, bhw, bhh, func(x, y int) {
			if x <= cx && y <= cy {
				setPixel(img, x, y, slabLit)
			} else if x > cx && y > cy {
				setPixel(img, x, y, slabDark)
			} else {
				setPixel(img, x, y, slab)
			}
		})
		forRectOutline(cx, cy, bhw, bhh, func(x, y int) { setPixel(img, x, y, slabDark) })
		// Blinking-light dabs: a cyan and an amber LED on the slab face, blended so they glow without
		// poster-painting. Restrained — a couple of pixels, the data-center tell.
		blendPixel(img, cx-bhw/2, cy, blend(slab, neonCyanAnchor, 0.72), 0.85)
		blendPixel(img, cx+bhw/2, cy, blend(slab, gasGlowAnchor, 0.66), 0.80)
		blendPixel(img, cx, cy-bhh/2, blend(slab, neonCyanAnchor, 0.66), 0.75)
	case tdPropNeonSign:
		// A small bright NEON sign: a short dark post capped by a small bright cyan-or-magenta neon dab —
		// the digital epoch's FIRST neon, kept small + blended so it glows without going full cyberpunk. The
		// hue alternates by position parity so a town gets a mix of cyan + magenta signs. Bounds-safe.
		post := darken(blend(prop, paved, 0.20), 0.28)
		tall := rad + rad/3 // a short sign-post — above props, below lamps
		for dy := -tall; dy <= rad; dy++ {
			setPixel(img, cx, cy+dy, post)
		}
		// Alternate cyan / magenta by the lot's pixel parity so the mix reads varied but deterministic.
		hue := neonCyanAnchor
		if (cx+cy)&1 == 0 {
			hue = neonMagentaAnchor
		}
		glow := blend(brighten(prop, 0.10), hue, 0.66)
		glowR := maxInt(rad/2, 1)
		forEllipse(cx, cy-tall, glowR, glowR, func(x, y int) { blendPixel(img, x, y, glow, 0.60) })
		setPixel(img, cx, cy-tall, brighten(glow, 0.20)) // bright neon core
	case tdPropHologram:
		// A bright TRANSLUCENT floating projection — the cyberpunk hologram. Unlike every other prop this
		// paints NOTHING solid: it BLENDS a bright cyan-or-magenta glow (~0.5) over whatever's beneath, so the
		// scene shows THROUGH it (a half-transparent projected shape, not an object). A tall soft column of
		// light rising off a base, brighter at the hovering head, with a faint emitter dab on the ground. The
		// hue alternates by pixel parity so a district gets a mix of cyan + magenta holograms. Bounds-safe:
		// blendPixel / forEllipse only (both clipped).
		hue := neonCyanAnchor
		if (cx+cy)&1 == 0 {
			hue = neonMagentaAnchor
		}
		proj := brighten(hue, 0.10)
		tall := rad + rad/2 // a tall floating projection — above the sign-posts
		// The soft column of light: a translucent vertical shaft, faint low, brightening as it rises.
		for dy := -tall; dy <= 0; dy++ {
			f := float64(-dy) / float64(tall+1) // 0 at the base → 1 at the head
			blendPixel(img, cx, cy+dy, proj, 0.28+0.22*f)
			if rad >= 3 { // a little width so it reads as a shape, not a wire
				blendPixel(img, cx-1, cy+dy, proj, 0.16+0.14*f)
				blendPixel(img, cx+1, cy+dy, proj, 0.16+0.14*f)
			}
		}
		// The hovering HEAD: a bright translucent bloom where the projection focuses.
		headR := maxInt(rad/2, 1)
		forEllipse(cx, cy-tall, headR, headR, func(x, y int) {
			fx := float64(x-cx) / float64(headR)
			fy := float64(y-(cy-tall)) / float64(headR)
			blendPixel(img, x, y, proj, 0.55*(1-0.5*(fx*fx+fy*fy)))
		})
		blendPixel(img, cx, cy-tall, brighten(proj, 0.20), 0.7) // the bright focal core
		// The emitter footprint: a faint ring where the projector sits on the ground.
		forEllipse(cx, cy, maxInt(rad/2, 1), maxInt(rad/3, 1), func(x, y int) { blendPixel(img, x, y, proj, 0.14) })
	case tdPropRocket:
		// A small ROCKET / gantry dab seasoning the spaceport — a bright vertical CAPSULE (a lit metallic body
		// with a pointed nose + a lit NW rim) standing beside a thin dark GANTRY tower. Metallic silver tones
		// pulled toward propCol so it retints with the theme; a warm exhaust glow at the base. Bounds-safe:
		// setPixel / blendPixel / forEllipse only.
		body := blend(brighten(prop, 0.10), metalSilverAnchor, 0.55) // bright metal capsule
		bodyLit := brighten(body, 0.16)                              // sunlit NW rim
		bodyDark := darken(body, 0.18)                               // shaded SE rim
		gantry := darken(blend(prop, metalSilverAnchor, 0.30), 0.34) // dark lattice tower
		tall := rad + rad/2                                          // a tall capsule — above props, below smokestacks
		// Ground shadow + a faint warm exhaust bloom at the base (south of the capsule).
		forEllipse(cx, cy+rad, maxInt(rad, 1), maxInt(rad/2, 1), func(x, y int) { blendPixel(img, x, y, darken(body, 0.4), 0.30) })
		blendPixel(img, cx, cy+rad, blend(bodyLit, gasGlowAnchor, 0.6), 0.5) // exhaust glow
		// The capsule body: a thin bright vertical column from base (cy+rad/2) up toward the nose (cy-tall).
		for dy := -tall; dy <= rad/2; dy++ {
			setPixel(img, cx, cy+dy, body)
			setPixel(img, cx-1, cy+dy, bodyLit)  // lit NW edge
			setPixel(img, cx+1, cy+dy, bodyDark) // shaded SE edge
		}
		// The pointed NOSE: a bright cap converging at the top.
		setPixel(img, cx, cy-tall, brighten(bodyLit, 0.14))
		setPixel(img, cx, cy-tall-1, brighten(bodyLit, 0.20))
		// The GANTRY: a thin dark lattice tower standing just to the west of the rocket.
		gx := cx - maxInt(rad/2+1, 2)
		for dy := -tall; dy <= rad; dy++ {
			setPixel(img, gx, cy+dy, gantry)
		}
		for _, dy := range []int{-tall + tall/3, 0, rad / 2} { // a few cross-arms reaching toward the rocket
			drawHSpan(img, gx, cx-1, cy+dy, gantry)
		}
	case tdPropLightMote:
		// A soft floating glowing MOTE seasoning the transcendent light-field — like the hologram it paints
		// NOTHING solid: it BLENDS a small translucent warm-white bloom over whatever's beneath, hovering a
		// little above the ground, so it reads as an ethereal spark of light, not an object. A faint radial
		// falloff + a bright soft core; the ether tones are pulled toward propCol so it retints. Bounds-safe:
		// blendPixel only.
		mote := blend(brighten(prop, 0.10), blend(etherWhiteAnchor, etherGoldAnchor, 0.35), 0.7)
		lift := maxInt(rad/2, 1) // the mote hovers a touch above its ground point
		mr := maxInt(rad, 1)
		for dy := -mr; dy <= mr; dy++ {
			for dx := -mr; dx <= mr; dx++ {
				fx := float64(dx) / float64(mr)
				fy := float64(dy) / float64(mr)
				d2 := fx*fx + fy*fy
				if d2 > 1.0 {
					continue
				}
				blendPixel(img, cx+dx, cy-lift+dy, mote, 0.50*(1-d2))
			}
		}
		blendPixel(img, cx, cy-lift, brighten(mote, 0.14), 0.8) // the bright soft core
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
		// A hut is a dwelling: read its era silhouette. Mudbrick huts are flat-topped blocks;
		// classical huts are white-stone blocks; primitive/timber keep the rounded domed thatch cap.
		switch style.houseProfile {
		case profileMudbrick:
			drawRoofMudbrick(img, cx, cy, hw, hh, rc)
		case profileStoneClassical:
			drawRoofStoneClassical(img, cx, cy, hw, hh, rc)
		case profileRowhouse:
			drawRoofRowhouse(img, cx, cy, hw, hh, rc)
		case profileModernFlat:
			drawRoofModernFlat(img, cx, cy, hw, hh, rc)
		case profileGlassTower:
			drawRoofGlassTower(img, cx, cy, hw, hh, rc)
		case profileMetalDome:
			drawRoofMetalDome(img, cx, cy, hw, hh, rc)
		case profileSpire:
			drawRoofSpire(img, cx, cy, hw, hh, rc)
		case profileLattice:
			drawRoofLattice(img, cx, cy, hw, hh, rc)
		case profileEthereal:
			drawRoofEthereal(img, cx, cy, hw, hh, rc)
		default:
			drawRoofHut(img, cx, cy, hw, hh, rc)
		}
	case roofRidge, roofLong:
		drawRoofHouse(img, cx, cy, hw, hh, rc, style.houseProfile)
	case roofTemple:
		drawRoofTemple(img, cx, cy, hw, hh, rc)
	case roofCamp:
		drawRoofCamp(img, cx, cy, hw, hh, rc)
	case roofStash:
		drawRoofStash(img, cx, cy, hw, hh, rc)
	case roofFlat:
		drawRoofFlat(img, cx, cy, hw, hh, rc)
	case roofWonder:
		// The era WONDER silhouette (locked #13, V3-B), now keyed off the dedicated wonderMotif
		// field (Phase 1b-i) rather than the house profile — so an age can pair thatch houses with
		// a megalith monument. Ancient = a stepped ZIGGURAT, medieval = a CATHEDRAL/KEEP with a
		// spire, stone = a MEGALITH stone circle; every other era keeps the ornate default complex.
		switch style.wonderMotif {
		case wonderZiggurat:
			drawRoofZiggurat(img, cx, cy, hw, hh, rc)
		case wonderCathedral:
			drawRoofCathedral(img, cx, cy, hw, hh, rc)
		case wonderMegalith:
			drawRoofMegalith(img, cx, cy, hw, hh, rc)
		case wonderTemple:
			drawRoofTempleWonder(img, cx, cy, hw, hh, rc)
		case wonderKeep:
			drawRoofKeep(img, cx, cy, hw, hh, rc)
		case wonderDome:
			drawRoofDome(img, cx, cy, hw, hh, rc)
		case wonderFactory:
			drawRoofFactory(img, cx, cy, hw, hh, rc)
		case wonderTower:
			drawRoofTower(img, cx, cy, hw, hh, rc)
		case wonderSpaceNeedle:
			drawRoofSpaceNeedle(img, cx, cy, hw, hh, rc)
		case wonderSkyscraper:
			drawRoofSkyscraper(img, cx, cy, hw, hh, rc)
		case wonderDataHub:
			drawRoofDataHub(img, cx, cy, hw, hh, rc)
		case wonderFusionCore:
			drawRoofFusionCore(img, cx, cy, hw, hh, rc)
		case wonderLaunchpad:
			drawRoofLaunchpad(img, cx, cy, hw, hh, rc)
		case wonderSpireArray:
			drawRoofSpireArray(img, cx, cy, hw, hh, rc)
		case wonderRingHub:
			drawRoofRingHub(img, cx, cy, hw, hh, rc)
		case wonderCrystalLattice:
			drawRoofCrystalLattice(img, cx, cy, hw, hh, rc)
		case wonderAscension:
			drawRoofAscension(img, cx, cy, hw, hh, rc)
		default:
			drawRoofWonder(img, cx, cy, hw, hh, rc)
		}
	default:
		drawRoofHouse(img, cx, cy, hw, hh, rc, style.houseProfile)
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

// drawRoofHouse renders a rectangular DWELLING roof in the era dialect (V3-B). It shares the
// pitched-rectangle base with drawRoofRidge (both slopes shaded off the ridge) but nudges the
// silhouette per profile so the same archetype reads era-appropriate:
//   - profileThatch  (primitive/default): the plain two-slope ridge, unchanged.
//   - profileMudbrick (ancient): a FLATTER, BLOCKIER roof — a broad flat inner deck with only a
//     thin shaded eave, so it reads as a low mud/adobe roof, not a steep pitch.
//   - profileTimber  (medieval): a STEEPER pitch — a narrow, sharply-shaded ridge with the slopes
//     darkening fast toward the eaves, so it reads as a tall timber-framed gable.
func drawRoofHouse(img *image.RGBA, cx, cy, hw, hh int, rc roofColors, prof roofProfile) {
	switch prof {
	case profileMudbrick:
		drawRoofMudbrick(img, cx, cy, hw, hh, rc)
	case profileTimber:
		drawRoofTimber(img, cx, cy, hw, hh, rc)
	case profileStoneClassical:
		drawRoofStoneClassical(img, cx, cy, hw, hh, rc)
	case profileRowhouse:
		drawRoofRowhouse(img, cx, cy, hw, hh, rc)
	case profileModernFlat:
		drawRoofModernFlat(img, cx, cy, hw, hh, rc)
	case profileGlassTower:
		drawRoofGlassTower(img, cx, cy, hw, hh, rc)
	case profileMetalDome:
		drawRoofMetalDome(img, cx, cy, hw, hh, rc)
	case profileSpire:
		drawRoofSpire(img, cx, cy, hw, hh, rc)
	case profileLattice:
		drawRoofLattice(img, cx, cy, hw, hh, rc)
	case profileEthereal:
		drawRoofEthereal(img, cx, cy, hw, hh, rc)
	default:
		drawRoofRidge(img, cx, cy, hw, hh, rc)
	}
}

// drawRoofMudbrick: the ANCIENT dwelling — a FLAT-TOPPED, blocky mud/adobe roof read from above.
// A full base rectangle in the shaded tone, a broad flat inner DECK in the lit base tone (most of
// the roof), and only a thin darker rim as the parapet/eave — no pitched ridge line, so it reads
// low and blocky, distinct from a pitched house. Serves the ancient hut too (a mud house has no
// dome). Base-derived tones only (no accent).
func drawRoofMudbrick(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	forRect(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.dark) }) // parapet/eave shadow
	dhw := maxInt(hw-1, 0)
	dhh := maxInt(hh-1, 0)
	forRect(cx, cy, dhw, dhh, func(x, y int) { img.SetRGBA(x, y, rc.base) }) // broad flat deck
	// A faint lit corner (NW) hint so the flat deck isn't a dead slab, kept base-derived.
	lhw := maxInt(hw/2, 0)
	lhh := maxInt(hh/2, 0)
	forRect(cx-hw+lhw/2+1, cy-hh+lhh/2+1, maxInt(lhw/2, 0), maxInt(lhh/2, 0), func(x, y int) {
		img.SetRGBA(x, y, rc.ridge)
	})
}

// drawRoofTimber: the MEDIEVAL dwelling — a STEEP pitched timber roof read from above. Like the
// ridge roof but the pitch reads sharper: the two slopes darken fast away from a NARROW bright
// ridge (a tall gable throws a hard light/shade split), so it reads as a steep timber roof rather
// than the ancient flat deck or the primitive gentle pitch. Ridge is base-derived (rc.ridge).
func drawRoofTimber(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	horizontalRidge := hw >= hh
	forRect(cx, cy, hw, hh, func(x, y int) {
		// Distance from the ridge line as a fraction of the half-span across the pitch; the slope
		// darkens with that distance so the pitch reads STEEP (fast falloff), and the two sides
		// split light (north/west lit, south/east shaded).
		var frac float64
		var lit bool
		if horizontalRidge {
			if hh > 0 {
				frac = float64(absInt(y-cy)) / float64(hh)
			}
			lit = y <= cy
		} else {
			if hw > 0 {
				frac = float64(absInt(x-cx)) / float64(hw)
			}
			lit = x <= cx
		}
		slope := rc.base
		if lit {
			// Lit side: base near the ridge, easing toward dark at the eave (a steep, fast falloff).
			slope = blend(rc.base, rc.dark, frac*0.55)
		} else {
			// Shaded side: already dark near the ridge, fully dark at the eave.
			slope = blend(blend(rc.base, rc.dark, 0.5), rc.dark, frac*0.7)
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

// drawRoofStoneClassical: the CLASSICAL dwelling (Phase 1b-ii) — a pale WHITE-STONE house body under
// a TERRACOTTA roof cap, with faint vertical COLUMN FLUTING on the front. Modeled on the blocky
// mudbrick roof (a flat-topped stone block) but two-tone: the body is the passed rc (which the
// classical style sets to a pale marble tone) and the roof cap is a warm terracotta band derived
// from clayAnchor blended with rc so it retints with the theme. The fluting is a few dim vertical
// seams down the lit body, a whisper of column texture. Base-derived tones only (no accent).
func drawRoofStoneClassical(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// White-stone body: the full footprint in the pale base, a thin shaded eave rim.
	forRect(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, rc.dark) }) // eave shadow rim
	dhw := maxInt(hw-1, 0)
	dhh := maxInt(hh-1, 0)
	forRect(cx, cy, dhw, dhh, func(x, y int) { img.SetRGBA(x, y, rc.base) }) // pale stone body

	// Terracotta roof cap: a warm reddish clay band across the TOP (north) portion of the block, so
	// the dwelling reads as a tiled roof over a white-stone house. Derived from clayAnchor blended
	// with the body so it stays in-family + retints.
	terra := blend(rc.base, clayAnchor, 0.62)
	terraDark := darken(terra, 0.24)
	capHH := maxInt(hh/2, 0)
	forRect(cx, cy-hh+capHH/2+1, dhw, maxInt(capHH/2, 0), func(x, y int) { img.SetRGBA(x, y, terra) })
	drawHSpan(img, cx-dhw, cx+dhw, cy-hh+1, terraDark) // a darker ridge line at the roof's top edge

	// Column fluting: a few dim vertical seams down the lower (front) half of the pale body, a hint
	// of a colonnade facade. Kept faint (a shade darker than the body) so it textures, not stripes.
	flute := darken(rc.base, 0.12)
	step := maxInt(hw/2, 1)
	for fx := cx - hw + 1; fx <= cx+hw-1; fx += step {
		for y := cy; y <= cy+dhh; y++ {
			setPixel(img, fx, y, flute)
		}
	}
}

// drawRoofRowhouse: the COLONIAL/INDUSTRIAL dwelling (Phase 1b-iii) — a TERRACE of narrow attached
// units read from above, so a house reads as a row of townhouses rather than one block. The footprint
// is divided along its LONG axis into 3–5 equal narrow units separated by thin dark dividing SEAMS
// (party walls); each unit is a small PITCHED roof, lit on the north slope + shaded on the south, with
// a bright ridge line down its spine. Base/dark/ridge-derived tones only (no accent), so it retints
// with the era material (warm brick for colonial, dull tin for industrial). Bounds-safe: every write
// goes through forRect/drawHSpan/setPixel (all clipped), so it never panics at any footprint.
func drawRoofRowhouse(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	seam := darken(rc.dark, 0.30) // dark party-wall seam between attached units

	// Fill the whole terrace first as a pitched block: north slope lit (base), south slope shaded
	// (dark), so even the smallest footprint reads as a roofed row. The per-unit seams + ridges are
	// stamped over this.
	longAxisH := hw >= hh // the row of units runs along the wider axis
	forRect(cx, cy, hw, hh, func(x, y int) {
		lit := y <= cy // north slope lit, south slope shaded (a pitched terrace roof)
		if lit {
			img.SetRGBA(x, y, rc.base)
		} else {
			img.SetRGBA(x, y, rc.dark)
		}
	})

	if longAxisH {
		// Units tile left→right across the width; each is a narrow vertical strip. Pick 3–5 units by
		// how wide the footprint is (a wider lot = more units), floored so a tiny lot still shows ≥2.
		fullW := 2*hw + 1
		units := fullW / 3
		if units < 2 {
			units = 2
		}
		if units > 5 {
			units = 5
		}
		x0 := cx - hw
		for u := 0; u <= units; u++ {
			sx := x0 + u*fullW/units
			// Dividing seam (party wall) at every internal boundary.
			if u > 0 && u < units {
				for y := cy - hh; y <= cy+hh; y++ {
					setPixel(img, sx, y, seam)
				}
			}
		}
		// Per-unit ridge: a bright horizontal spine across each unit's own width along the crest (cy).
		for u := 0; u < units; u++ {
			ua := x0 + u*fullW/units
			ub := x0 + (u+1)*fullW/units
			drawHSpan(img, ua+1, ub-1, cy, rc.ridge)
		}
	} else {
		// Tall lot: units tile top→bottom, each a narrow horizontal strip; ridge runs vertically.
		fullH := 2*hh + 1
		units := fullH / 3
		if units < 2 {
			units = 2
		}
		if units > 5 {
			units = 5
		}
		y0 := cy - hh
		for u := 0; u <= units; u++ {
			sy := y0 + u*fullH/units
			if u > 0 && u < units {
				drawHSpan(img, cx-hw, cx+hw, sy, seam)
			}
		}
		for u := 0; u < units; u++ {
			ua := y0 + u*fullH/units
			ub := y0 + (u+1)*fullH/units
			for y := ua + 1; y < ub; y++ {
				setPixel(img, cx, y, rc.ridge)
			}
		}
	}
}

// drawRoofModernFlat: the ELECTRIC/ATOMIC dwelling (V3-B ELECTRIC epoch) — a FLAT-topped modern block
// read from above: a flat roof SLAB filling the footprint, a thin darker PARAPET rim around the edge (a
// raised roof-wall), a subtle lit NW edge for a clean modern sheen, and a small central rooftop VENT /
// detail (an AC / stair-head dab). Reads as a flat-roofed mid-rise, NOT a pitched house — the groundwork
// for later skyscrapers. Base/dark/ridge-derived tones only (no accent), so it retints with the era
// material (warm concrete for electric, cool pastel steel for atomic). Bounds-safe: every write goes
// through forRect+setPixel / forRectOutline+setPixel / drawHSpan (all clipped), so it never panics.
func drawRoofModernFlat(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// The flat roof slab: the full footprint in the base tone, lit a touch on the N/W so the block reads
	// as a clean modern deck rather than a dead slab.
	forRect(cx, cy, hw, hh, func(x, y int) {
		if x <= cx && y <= cy {
			setPixel(img, x, y, brighten(rc.base, 0.06)) // lit NW quadrant
		} else {
			setPixel(img, x, y, rc.base)
		}
	})
	// Thin darker PARAPET rim around the edge (a raised roof-wall), with a brighter N/W lip for the
	// modern lit edge.
	forRectOutline(cx, cy, hw, hh, func(x, y int) { setPixel(img, x, y, rc.dark) })
	drawHSpan(img, cx-hw, cx+hw, cy-hh, rc.ridge) // bright parapet lip along the north edge
	for y := cy - hh; y <= cy+hh; y++ {
		setPixel(img, cx-hw, y, rc.ridge) // and down the west edge
	}
	// Central rooftop VENT / detail: a small dark square (an AC unit / stair-head), a touch off-centre so
	// the deck doesn't read perfectly symmetric.
	vhw := maxInt(hw/3, 0)
	vhh := maxInt(hh/3, 0)
	vent := darken(rc.dark, 0.14)
	forRect(cx, cy, vhw, vhh, func(x, y int) { setPixel(img, x, y, vent) })
	setPixel(img, cx-vhw, cy-vhh, brighten(rc.base, 0.10)) // a lit NW corner on the vent housing
}

// drawRoofGlassTower: the MODERN/INFORMATION/DIGITAL dwelling (V3-C) — a TALL glass-and-steel tower read
// from above. Modeled on the flat modern block but re-read as a HIGH-RISE: a cool blue-grey GLASS slab
// filling the footprint, a lit WINDOW-GRID sheen (a couple of bright vertical mullions crossed by a
// couple of bright floor bands), and — the key height cue — a STRONG extended SE drop-shadow beyond the
// footprint, so the block reads TALL, not flat. Glass tones are blended with the passed roof colors so
// the tower retints on a theme switch and stays in the era mood. Bounds-safe: every write goes through
// fillRectC / forRect+setPixel / drawHSpan / blendPixel (all clipped), so it never panics at any footprint.
func drawRoofGlassTower(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Cool blue-grey glass + a pale sky-lit sheen pulled toward the (already theme/lineage-tinted) roof
	// colors, so the whole tower retints yet reads as a steely glass curtain-wall.
	glass := blend(rc.base, glassAnchor, 0.52)     // the cool glass curtain-wall face
	glassDark := blend(rc.dark, glassAnchor, 0.40) // shaded S/E face + parapet
	sheen := blend(glass, glassLitAnchor, 0.55)    // bright sky-lit window highlights
	frame := darken(glassDark, 0.16)               // dark mullion / edge frame

	// EXTENDED HEIGHT SHADOW: a longer SE drop beyond the footprint sells the high-rise — a tall glass
	// mass throws a longer shadow than the flat blocks around it. Blended into the ground so it darkens
	// rather than paints a slab. Stepped out with the offset so it reads as a raking shadow.
	towerShadow := darken(glassDark, 0.34)
	for dy := 1; dy <= maxInt(hh/2, 1); dy++ {
		drawHSpan(img, cx-hw+dy, cx+hw+dy, cy+hh+dy, towerShadow)
	}

	// THE GLASS SLAB: the full footprint in the glass tone, lit N/W + shaded S/E so the tower reads as a
	// raised curtain-wall mass rather than a flat deck.
	forRect(cx, cy, hw, hh, func(x, y int) {
		if x <= cx && y <= cy {
			setPixel(img, x, y, brighten(glass, 0.05)) // lit NW curtain-wall
		} else if x > cx && y > cy {
			setPixel(img, x, y, glassDark) // shaded SE face
		} else {
			setPixel(img, x, y, glass)
		}
	})
	// Dark edge FRAME (the steel structure at the roof rim), with a bright N/W lip for the glare.
	forRectOutline(cx, cy, hw, hh, func(x, y int) { setPixel(img, x, y, frame) })
	drawHSpan(img, cx-hw, cx+hw, cy-hh, sheen) // bright glare along the north parapet
	for y := cy - hh; y <= cy+hh; y++ {
		setPixel(img, cx-hw, y, sheen) // and down the west face
	}

	// WINDOW-GRID SHEEN: a couple of bright vertical mullion lines crossed by a couple of bright floor
	// bands, so the slab reads as a lit glass grid (a skyscraper's window lattice), not a plain panel.
	for _, fx := range []float64{-0.45, 0.10, 0.55} {
		gx := cx + int(float64(hw)*fx)
		for y := cy - hh + 1; y <= cy+hh-1; y++ {
			blendPixel(img, gx, y, sheen, 0.5)
		}
	}
	for _, fy := range []float64{-0.35, 0.25} {
		gy := cy + int(float64(hh)*fy)
		drawHSpan(img, cx-hw+1, cx+hw-1, gy, blend(glass, sheen, 0.5))
	}
	setPixel(img, cx-hw+1, cy-hh+1, brighten(sheen, 0.12)) // a bright glint on the NW crown
}

// drawRoofSkyscraper: the MODERN/INFORMATION/DIGITAL wonder (V3-C) — a SUPERTALL glass tower read from
// above. Where the deco wonderTower is a WIDE stepped concrete pyramid, this is a NARROW sheer glass
// SLAB: a tall thin cool-glass rectangle (footprint pinched inward on the long axis so it reads slender)
// with a bright WINDOW GRID, a central MAST / ANTENNA dab rising off the crown, and — the height cue — a
// LONG raking SE shadow well beyond the footprint. Reads clearly apart from the wide deco tower, the
// round dome, the blocky keep, the colonnade temple, and the factory hall+chimneys. Glass tones are
// blended with the passed roof colors so it retints on a theme switch. Bounds-safe: every write goes
// through fillRectC / forRect+setPixel / drawHSpan / setPixel / blendPixel (all clipped), never panics.
func drawRoofSkyscraper(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	glass := blend(rc.base, glassAnchor, 0.54)     // the cool glass curtain-wall face
	glassDark := blend(rc.dark, glassAnchor, 0.42) // shaded S/E face
	sheen := blend(glass, glassLitAnchor, 0.60)    // bright sky-lit window highlights
	frame := darken(glassDark, 0.18)               // dark steel edge frame
	crown := brighten(sheen, 0.14)                 // the bright mast / antenna catching light

	// A NARROW slender footprint: pinch the wider axis in so the wonder reads as a sheer supertall slab,
	// not the wide stepped deco pyramid. (Floored so it never collapses at tiny sizes.)
	shw, shh := hw, hh
	if hw >= hh {
		shw = maxInt(hw*3/5, 1)
	} else {
		shh = maxInt(hh*3/5, 1)
	}

	// LONG HEIGHT SHADOW: a raking SE drop well beyond the footprint — a supertall throws the longest
	// shadow on the map. Blended into the ground so it darkens rather than paints a slab.
	towerShadow := darken(glassDark, 0.36)
	for dy := 1; dy <= maxInt(hh, 2); dy++ {
		drawHSpan(img, cx-shw+dy, cx+shw+dy, cy+shh+dy, towerShadow)
	}

	// THE SLAB: the slender footprint in glass, lit N/W + shaded S/E for a sheer curtain-wall mass.
	forRect(cx, cy, shw, shh, func(x, y int) {
		if x <= cx && y <= cy {
			setPixel(img, x, y, brighten(glass, 0.06))
		} else if x > cx && y > cy {
			setPixel(img, x, y, glassDark)
		} else {
			setPixel(img, x, y, glass)
		}
	})
	forRectOutline(cx, cy, shw, shh, func(x, y int) { setPixel(img, x, y, frame) })
	drawHSpan(img, cx-shw, cx+shw, cy-shh, sheen) // bright glare along the north parapet
	for y := cy - shh; y <= cy+shh; y++ {
		setPixel(img, cx-shw, y, sheen) // lit west face
	}

	// WINDOW GRID: bright vertical mullions + a few floor bands, denser than a house so it reads as a
	// glittering high-rise curtain-wall.
	for _, fx := range []float64{-0.4, 0.2} {
		gx := cx + int(float64(shw)*fx)
		for y := cy - shh + 1; y <= cy+shh-1; y++ {
			blendPixel(img, gx, y, sheen, 0.55)
		}
	}
	bandEvery := maxInt(shh/3, 1)
	for gy := cy - shh + bandEvery; gy < cy+shh; gy += bandEvery {
		drawHSpan(img, cx-shw+1, cx+shw-1, gy, blend(glass, sheen, 0.45))
	}

	// CENTRAL MAST / ANTENNA: a thin bright finial rising off the crown — the broadcast spire that tops a
	// supertall, and the strongest "this is the tallest thing on the map" cue.
	mastH := maxInt(minInt(hw, hh)*3/4, 2)
	for dy := -mastH; dy <= 0; dy++ {
		setPixel(img, cx, cy-shh+dy, crown)
	}
	setPixel(img, cx, cy-shh-mastH, brighten(crown, 0.18)) // a bright beacon at the mast tip
	setPixel(img, cx-shw+1, cy-shh+1, crown)               // a glint on the NW crown
}

// drawRoofDataHub: the INFORMATION wonder (Phase 1i) — a DATA megastructure read from above. Where the
// modern wonderSkyscraper is a single slender GLASS slab, this is a WIDE LOW server-farm complex: a broad
// flat base of cool grey server BLOCKS gridded by dark cooling channels, a central tall COMMS antenna /
// lattice MAST rising off the core (with a raking SE mast shadow), and a scatter of bright beacon/LED dabs
// (cyan + amber) blinking across the farm. Reads clearly apart from the slender glass skyscraper (tall + thin
// + windowed), the deco tower (square tiers), and the space needle (a saucer on a stem). Cold data-grey +
// cyan/amber beacons are BLENDED with the passed roof colors so it retints on a theme switch. Bounds-safe:
// every write goes through forRect / drawHSpan / setPixel / blendPixel (all clipped), so it never panics.
func drawRoofDataHub(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	grey := blend(rc.base, dataGreyAnchor, 0.58)                      // the cold server-block face
	greyLit := brighten(grey, 0.14)                                   // sunlit NW block edge
	greyDark := blend(rc.dark, dataGreyAnchor, 0.44)                  // shaded SE block + cooling channels
	channel := darken(greyDark, 0.22)                                 // dark cooling-channel gaps between blocks
	mast := blend(rc.dark, dataGreyAnchor, 0.60)                      // the dark comms lattice mast
	mastLit := brighten(mast, 0.18)                                   // sunlit NW rim of the mast
	ledCyan := blend(brighten(rc.base, 0.10), fusionCyanAnchor, 0.62) // a cool cyan beacon dab
	ledAmber := blend(brighten(rc.base, 0.10), gasGlowAnchor, 0.55)   // a warm amber beacon dab

	// THE BASE: a wide low server-farm block filling the footprint, lit N/W + shaded S/E, so the complex sits
	// as a raised low mass rather than a flat slab.
	forRect(cx, cy, hw, hh, func(x, y int) {
		if x <= cx || y <= cy {
			setPixel(img, x, y, grey)
		} else {
			setPixel(img, x, y, greyDark)
		}
	})

	// SERVER-BLOCK GRID: a few dark cooling channels running both axes so the base reads as rows of server
	// racks / cooling aisles, not a plain slab. Vertical aisles every ~third, one horizontal spine.
	aisleEvery := maxInt(hw/2, 2)
	for gx := cx - hw + aisleEvery; gx < cx+hw; gx += aisleEvery {
		for y := cy - hh; y <= cy+hh; y++ {
			setPixel(img, gx, y, channel)
		}
	}
	rowEvery := maxInt(hh/2, 2)
	for gy := cy - hh + rowEvery; gy < cy+hh; gy += rowEvery {
		drawHSpan(img, cx-hw, cx+hw, gy, channel)
	}
	// A lit NW parapet edge so the farm's north face catches light.
	drawHSpan(img, cx-hw, cx+hw, cy-hh, greyLit)
	for y := cy - hh; y <= cy+hh; y++ {
		setPixel(img, cx-hw, y, greyLit)
	}

	// CENTRAL COMMS MAST: a bold dark lattice tower standing at the farm's core — a broad shaded stalk (a
	// small filled footing block) rising to a thin tall antenna, lit on its NW edge, with a bright beacon tip.
	footR := maxInt(minInt(hw, hh)/4, 1)
	forRect(cx, cy, footR, footR, func(x, y int) {
		if x <= cx {
			setPixel(img, x, y, mastLit)
		} else {
			setPixel(img, x, y, mast)
		}
	})
	mastH := maxInt(minInt(hw, hh)*3/4, 2)
	for dy := -mastH; dy <= 0; dy++ {
		setPixel(img, cx, cy+dy, mast)
		setPixel(img, cx-1, cy+dy, mastLit)
	}
	// A short cross-arm high on the mast so it reads as a comms lattice, not a plain pole.
	armW := maxInt(footR, 1)
	drawHSpan(img, cx-armW, cx+armW, cy-mastH*2/3, mast)
	// RAKING SE MAST SHADOW below the base so the tall lattice reads as standing well above the low farm.
	for dy := 1; dy <= maxInt(hh/2, 1); dy++ {
		blendPixel(img, cx+dy, cy+hh+dy, channel, 0.5)
	}
	setPixel(img, cx, cy-mastH, brighten(ledCyan, 0.16))   // a bright cyan beacon at the mast tip
	setPixel(img, cx, cy-mastH-1, brighten(ledCyan, 0.24)) // the pinpoint above the tip

	// BEACON / LED DABS: a deterministic scatter of small blinking lights across the farm — cyan + amber
	// alternating — so the server-city reads as a live data hub, not a dead grey block. Placed on a fixed
	// lattice of offsets from center (scaled to the footprint), each clamped via blendPixel.
	leds := []struct {
		fx, fy float64
		warm   bool
	}{
		{-0.60, -0.55, false}, {0.55, -0.60, true}, {-0.65, 0.45, true},
		{0.62, 0.50, false}, {-0.30, 0.66, false}, {0.28, -0.30, true},
	}
	for _, l := range leds {
		lx := cx + int(float64(hw)*l.fx)
		ly := cy + int(float64(hh)*l.fy)
		c := ledCyan
		if l.warm {
			c = ledAmber
		}
		blendPixel(img, lx, ly, c, 0.85)
		blendPixel(img, lx, ly-1, brighten(c, 0.18), 0.6) // a soft glow above each dab
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

// drawRoofZiggurat: the ANCIENT wonder (locked #13, V3-B) — a stepped pyramid temple read from
// above as a set of CONCENTRIC SQUARE TIERS shrinking toward a bright central shrine, so it reads
// as a ziggurat's terraces (Ur / Mesopotamian temple-mount) rather than the rounded default
// complex. Each tier is a shade lighter than the one below (terraces catching light as they rise);
// the top is the base-derived ridge tone (no accent — the sat cap keeps even a gold-lineage wonder
// an earthy clay). A cross axis on the base grounds it as a symmetric monument.
func drawRoofZiggurat(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	const tiers = 4
	for t := 0; t < tiers; t++ {
		f := float64(t) / float64(tiers) // 0 (base) .. →1 (top)
		thw := maxInt(int(float64(hw)*(1-f)), 1)
		thh := maxInt(int(float64(hh)*(1-f)), 1)
		// Lower tiers dark, rising tiers lighten toward the base tone; the top tier is the ridge.
		col := blend(rc.dark, rc.base, f/(1-1.0/tiers))
		if t == tiers-1 {
			col = rc.ridge
		}
		forRect(cx, cy, thw, thh, func(x, y int) { img.SetRGBA(x, y, col) })
		// A thin darker step edge on each tier's south+east so the terraces read as raised steps.
		drawHSpan(img, cx-thw, cx+thw, cy+thh, darken(col, 0.18))
		for y := cy - thh; y <= cy+thh; y++ {
			img.SetRGBA(cx+thw, y, darken(col, 0.18))
		}
	}
	// Symmetric cross axes across the base tier ground it as a monument.
	drawHSpan(img, cx-hw, cx+hw, cy, blend(rc.base, rc.ridge, 0.4))
}

// drawRoofKeep: the IRON-AGE wonder (Phase 1b-ii follow-up) — a fortified KEEP / great hall read
// from above so iron stops looking like tinted bronze. A solid blocky stone-and-timber HALL fills
// the footprint (lit NW, shaded SE), a timber roof-beam runs its long axis, and a taller, brighter
// square WATCHTOWER rises from one end capped by a dark CRENELLATED battlement, with a small dark
// gate on the south face. The blocky hall+tower silhouette reads clearly apart from the stepped
// ziggurat, the colonnade temple, the cruciform cathedral, and the stone-circle megalith. Stone
// (granite/stone anchors) + timber (timberAnchor) are BLENDED with the passed roof colors, so the
// keep retints on a theme switch and stays in the era mood. Bounds-safe: every write goes through
// forRect/fillRectC/setPixel/drawHSpan (all clipped), so it never panics at any footprint.
func drawRoofKeep(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Stone-and-timber palette pulled toward the (already theme/lineage-tinted) roof colors.
	stone := blend(rc.base, blend(graniteAnchor, stoneAnchor, 0.5), 0.50)
	stoneLit := brighten(stone, 0.16) // sunlit tower stone
	stoneDark := blend(rc.dark, graniteAnchor, 0.42)
	timber := blend(rc.base, timberAnchor, 0.50) // roof beam + gate framing
	batt := darken(stoneDark, 0.30)              // dark crenellations + gate

	horizontal := hw >= hh

	// THE HALL: a solid blocky body filling the footprint, lit N/W and shaded S/E so it reads as a
	// raised stone hall rather than a flat slab.
	forRect(cx, cy, hw, hh, func(x, y int) {
		if x <= cx || y <= cy { // NW-lit
			setPixel(img, x, y, stone)
		} else {
			setPixel(img, x, y, stoneDark)
		}
	})
	// A timber roof-beam along the hall's long axis.
	if horizontal {
		drawHSpan(img, cx-hw, cx+hw, cy, timber)
	} else {
		for y := cy - hh; y <= cy+hh; y++ {
			setPixel(img, cx, y, timber)
		}
	}

	// THE WATCHTOWER: a taller, narrower square block seated at ONE END of the hall (west if the
	// hall runs E-W, north otherwise), drawn brighter so it reads as a raised keep above the hall.
	var twx, twy, twhw, twhh int
	if horizontal {
		twhw, twhh = maxInt(hw/3, 2), maxInt(hh, 2)
		twx, twy = cx-hw+twhw, cy
	} else {
		twhw, twhh = maxInt(hw, 2), maxInt(hh/3, 2)
		twx, twy = cx, cy-hh+twhh
	}
	fillRectC(img, twx, twy, twhw, twhh, stoneLit)
	// Crenellated battlement: dark notches around the tower rim (alternating pixels).
	forRectOutline(twx, twy, twhw, twhh, func(x, y int) {
		if (x+y)&1 == 0 {
			setPixel(img, x, y, batt)
		}
	})
	setPixel(img, twx, twy, batt) // a dark keep-top / arrow-slit for depth

	// GATE: a small dark doorway on the hall's south face, offset onto the long axis.
	if horizontal {
		setPixel(img, cx+hw/3, cy+hh, batt)
	} else {
		setPixel(img, cx+hw, cy+hh/3, batt)
	}
}

// drawRoofDome: the RENAISSANCE wonder (Phase 1b) — a great DOMED ROTUNDA read from above (the
// Florence Duomo / St-Peter's read), so renaissance stops looking like a tinted medieval city. A
// square stone DRUM fills the footprint; a large filled lead-grey/cream DISC (the dome) sits on it,
// ringed by a couple of darker CONCENTRIC RIBS and eight radial RIB LINES converging on the crown so
// the curvature reads; a small bright square LANTERN / cupola dab caps the very center. The round
// domed silhouette reads clearly apart from the stepped ziggurat, the colonnade+pediment temple, the
// cruciform+spire cathedral, the blocky keep+tower, and the open stone-circle megalith. Lead + stone
// (leadAnchor / stoneAnchor) are BLENDED with the passed roof colors so the dome retints on a theme
// switch and stays in the cream era mood. Bounds-safe: every write goes through
// fillRectC/fillDisc/forEllipse/setPixel (all clipped), so it never panics at any footprint.
func drawRoofDome(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Cream-stone drum + pale lead dome tones pulled toward the (already theme/lineage-tinted) roof
	// colors, so the whole rotunda retints yet reads as pale lead-and-cream stone.
	stone := blend(rc.base, blend(creamStoneAnchor, stoneAnchor, 0.4), 0.50) // the square drum below the dome
	stoneDark := blend(rc.dark, stoneAnchor, 0.40)
	lead := blend(rc.base, blend(leadAnchor, stoneAnchor, 0.35), 0.55) // the pale lead dome shell
	leadLit := brighten(lead, 0.16)                                    // NW-lit crown of the dome
	rib := darken(lead, 0.22)                                          // darker concentric + radial ribs
	lantern := brighten(leadLit, 0.16)                                 // bright cupola / lantern at the crown

	// THE DRUM: a solid square stone body filling the footprint, lit N/W + shaded S/E so the rotunda
	// sits on a raised base rather than a flat slab.
	forRect(cx, cy, hw, hh, func(x, y int) {
		if x <= cx || y <= cy {
			setPixel(img, x, y, stone)
		} else {
			setPixel(img, x, y, stoneDark)
		}
	})

	// THE DOME: a large filled disc centered on the drum, radius the inner ~85% of the smaller extent
	// so a rim of drum shows around it. Lit NW / shaded SE across the disc for a domed sheen.
	rad := maxInt(minInt(hw, hh)*17/20, 1)
	forEllipse(cx, cy, rad, rad, func(x, y int) {
		if x <= cx || y <= cy {
			setPixel(img, x, y, lead)
		} else {
			setPixel(img, x, y, darken(lead, 0.12))
		}
	})

	// CONCENTRIC RIBS: a couple of darker rings stepping in toward the crown so the shell reads curved.
	// Each ring is stamped as a rim (walk the circle) — a raised concentric rib of the cupola.
	for _, f := range []float64{0.66, 0.40} {
		rr := maxInt(int(float64(rad)*f), 1)
		for i := 0; i < 48; i++ {
			ang := 2 * math.Pi * float64(i) / 48
			setPixel(img, cx+int(math.Round(math.Cos(ang)*float64(rr))), cy+int(math.Round(math.Sin(ang)*float64(rr))), rib)
		}
	}

	// RADIAL RIBS: eight spokes from the crown out to the dome rim, the ribs of the cupola.
	for i := 0; i < 8; i++ {
		ang := 2 * math.Pi * float64(i) / 8
		for r := 1; r <= rad; r++ {
			setPixel(img, cx+int(math.Round(math.Cos(ang)*float64(r))), cy+int(math.Round(math.Sin(ang)*float64(r))), rib)
		}
	}

	// LANTERN / CUPOLA: a small bright square cap at the very crown, the light-well atop the dome.
	lhw := maxInt(rad/5, 1)
	fillRectC(img, cx, cy, lhw, lhw, leadLit)
	setPixel(img, cx, cy, lantern)
}

// drawRoofFusionCore: the FUSION wonder (NEON epoch) — a glowing REACTOR seen from above: a set of
// CONCENTRIC BRIGHT CYAN RINGS brightening inward toward a WHITE-HOT central BLOOM. Unlike the renaissance
// dome (which is a solid drum + shell + radial/concentric RIBS + a lantern cap) this is a radiant target of
// GLOWING rings with an incandescent core — no drum, no ribs, no mast — so it reads clearly apart from the
// dome, the skyscraper slab, the deco tower, and the factory hall. White + cyan (fusionWhiteAnchor /
// fusionCyanAnchor) are BLENDED with the passed roof colors so it retints on a theme switch. Bounds-safe:
// every write goes through fillDisc / forEllipse / setPixel / blendPixel (all clipped), so it never panics.
func drawRoofFusionCore(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Clean cyan glow + a white-hot bloom pulled toward the (already theme/lineage-tinted) roof colors, so the
	// whole reactor retints yet reads as a radiant cyan-and-white core.
	cyan := blend(rc.base, fusionCyanAnchor, 0.55)    // the glowing ring cyan
	cyanDim := blend(rc.dark, fusionCyanAnchor, 0.34) // the outer, dimmer ring
	white := blend(rc.base, fusionWhiteAnchor, 0.70)  // the white-hot inner tone
	bloom := brighten(white, 0.20)                    // the incandescent center

	rad := maxInt(minInt(hw, hh), 1)

	// A faint containment DECK: a soft dim-cyan disc filling the footprint so the rings sit on a glowing
	// pad rather than bare ground (lit NW / shaded SE for a shallow bowl).
	forEllipse(cx, cy, rad, rad, func(x, y int) {
		if x <= cx || y <= cy {
			setPixel(img, x, y, blend(cyanDim, white, 0.15))
		} else {
			setPixel(img, x, y, cyanDim)
		}
	})

	// CONCENTRIC GLOWING RINGS: stepping inward from the rim, each ring a filled disc a shade brighter than
	// the last, so the reactor reads as a radiant target brightening toward the center. Drawn largest-first
	// so each smaller, brighter disc overpaints the middle of the previous ring, leaving a bright rim band.
	for _, step := range []struct {
		f float64
		c color.RGBA
	}{
		{0.82, cyan},
		{0.58, blend(cyan, white, 0.40)},
		{0.34, white},
	} {
		rr := maxInt(int(float64(rad)*step.f), 1)
		fillDisc(img, cx, cy, rr, step.c)
	}

	// The WHITE-HOT BLOOM: a small incandescent core at the very center, with a soft translucent halo bleeding
	// out so the center reads as glowing, not a hard dot.
	coreR := maxInt(rad/4, 1)
	fillDisc(img, cx, cy, coreR, bloom)
	haloR := maxInt(rad/2, 1)
	forEllipse(cx, cy, haloR, haloR, func(x, y int) {
		fx := float64(x-cx) / float64(haloR)
		fy := float64(y-cy) / float64(haloR)
		blendPixel(img, x, y, bloom, 0.45*(1-(fx*fx+fy*fy)))
	})
	setPixel(img, cx, cy, brighten(bloom, 0.15)) // the brightest pinpoint
}

// drawRoofMetalDome: the SPACE dwelling (NEON epoch) — a small pale METALLIC habitat DOME seen from above:
// a lit silver DISC with a curved NW HIGHLIGHT sweeping across the crown and a darker rim, so it reads as a
// pressurised metal dome rather than a flat disc. Cooler + greyer than the fusion white. Silver
// (metalSilverAnchor) is BLENDED with the passed roof colors so it retints on a theme switch. Bounds-safe:
// forEllipse / setPixel only (both clipped), so it never panics at any footprint.
func drawRoofMetalDome(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	metal := blend(rc.base, metalSilverAnchor, 0.60)     // the pale silver shell
	metalLit := brighten(metal, 0.16)                    // the NW-lit crown sweep
	metalDark := blend(rc.dark, metalSilverAnchor, 0.40) // the shaded SE flank
	rim := darken(metalDark, 0.16)                       // the dark base rim

	rad := maxInt(minInt(hw, hh), 1)

	// THE DOME DISC: a filled silver disc, lit toward the NW crown and shaded toward the SE flank so it reads
	// domed (a curved metal shell), not a flat coin.
	forEllipse(cx, cy, rad, rad, func(x, y int) {
		fx := float64(x-cx) / float64(rad)
		fy := float64(y-cy) / float64(rad)
		d := fx + fy // -2 NW (lit) .. +2 SE (shaded)
		switch {
		case d < -0.5:
			setPixel(img, x, y, metalLit)
		case d > 0.6:
			setPixel(img, x, y, metalDark)
		default:
			setPixel(img, x, y, metal)
		}
	})

	// CURVED NW HIGHLIGHT: a bright specular arc sweeping across the upper-left of the shell (the sun glinting
	// off curved metal) — a thin crescent one band in from the rim, on the NW side only.
	hlR := maxInt(rad*7/10, 1)
	for i := 0; i < 40; i++ {
		ang := math.Pi + math.Pi*float64(i)/40 // the upper-left arc of the circle (clipped to the NW quadrant below)
		hx := cx + int(math.Round(math.Cos(ang)*float64(hlR)))
		hy := cy + int(math.Round(math.Sin(ang)*float64(hlR)))
		if hx <= cx && hy <= cy { // NW quadrant only
			setPixel(img, hx, hy, brighten(metalLit, 0.12))
		}
	}

	// THE RIM: a darker base ring around the shell so the dome sits on a defined footprint.
	for i := 0; i < 48; i++ {
		ang := 2 * math.Pi * float64(i) / 48
		setPixel(img, cx+int(math.Round(math.Cos(ang)*float64(rad))), cy+int(math.Round(math.Sin(ang)*float64(rad))), rim)
	}
	setPixel(img, cx, cy, brighten(metal, 0.10)) // a small lit apex vent
}

// drawRoofSpire: the INTERSTELLAR dwelling (COSMIC epoch) — a TALL NARROW tapering metallic SPIRE read from
// above. A slender arcology needle throws a long SE height shadow (like the supertall), stands as a thin
// bright metal shaft tapering to a lit tip, and sits on a small dark metal base pad so it reads as a
// planted spire rather than a floating sliver. Base-derived silver tones (no accent) so it retints with the
// roof material.
func drawRoofSpire(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	metal := blend(rc.base, metalSilverAnchor, 0.58)     // the pale metal shaft
	metalLit := brighten(metal, 0.18)                    // the sunlit NW edge + tip
	metalDark := blend(rc.dark, metalSilverAnchor, 0.36) // the shaded SE edge + base
	base := darken(metalDark, 0.14)                      // the dark planted base pad

	rad := maxInt(minInt(hw, hh), 1)

	// LONG HEIGHT SHADOW: a thin raking SE drop well beyond the footprint — a spire is the tallest thing in
	// the block, so it throws a long shadow. Blended into the ground so it darkens rather than paints a slab.
	shadow := darken(metalDark, 0.34)
	for d := 1; d <= maxInt(hh, 2); d++ {
		blendPixel(img, cx+d, cy+d, shadow, 0.42)
		blendPixel(img, cx+d+1, cy+d, shadow, 0.30)
	}

	// THE BASE PAD: a small dark metal disc under the spire so it reads planted, lit NW / shaded SE.
	baseR := maxInt(rad/2, 1)
	forEllipse(cx, cy, baseR, baseR, func(x, y int) {
		if x <= cx && y <= cy {
			setPixel(img, x, y, metalDark)
		} else {
			setPixel(img, x, y, base)
		}
	})

	// THE SHAFT: a thin vertical metal needle rising well above center, lit on the NW edge + shaded on the SE
	// edge so the narrow tower reads as a round tapering volume, not a flat line. It NARROWS toward the top
	// (the lit column is a pixel wide near the tip, wider at the base).
	tall := maxInt(rad*9/5, 3) // the shaft rises well above the base
	for dy := -tall; dy <= 0; dy++ {
		frac := float64(-dy) / float64(tall) // 0 at base .. 1 at tip
		setPixel(img, cx, cy+dy, metal)
		if frac < 0.7 { // only the lower/mid shaft is wide enough for lit + shaded flanks
			setPixel(img, cx-1, cy+dy, metalLit)
			setPixel(img, cx+1, cy+dy, metalDark)
		}
	}
	// THE LIT TIP: a bright pinnacle beacon crowning the spire.
	setPixel(img, cx, cy-tall, brighten(metalLit, 0.16))
	setPixel(img, cx, cy-tall-1, brighten(metalLit, 0.24))
}

// drawRoofSpireArray: the INTERSTELLAR wonder (COSMIC epoch) — a CLUSTER of tall metallic spires: a ring of
// shorter satellite spires around a single TALLEST central spire, each throwing a long SE height shadow. The
// centrepiece scale-up of the dwelling spire — a deep-space arcology core. Base-derived silver tones.
func drawRoofSpireArray(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	pad := blend(rc.base, metalSilverAnchor, 0.40)       // the pale plated base apron
	padDark := blend(rc.dark, metalSilverAnchor, 0.30)   // shaded apron
	metal := blend(rc.base, metalSilverAnchor, 0.60)     // the bright spire metal
	metalLit := brighten(metal, 0.18)                    // sunlit NW edge + tips
	metalDark := blend(rc.dark, metalSilverAnchor, 0.38) // shaded SE edge
	shadow := darken(padDark, 0.30)                      // the raking spire shadows

	rad := maxInt(minInt(hw, hh), 1)

	// THE BASE APRON: a filled metal disc filling the footprint, lit NW / shaded SE — the arcology platform
	// the spires rise from.
	forEllipse(cx, cy, rad, rad, func(x, y int) {
		if x <= cx || y <= cy {
			setPixel(img, x, y, pad)
		} else {
			setPixel(img, x, y, padDark)
		}
	})

	// spire draws one needle at (sx, sy) rising to height h, with a long SE shadow + lit tip.
	spire := func(sx, sy, h int) {
		for d := 1; d <= h*2/3; d++ { // shadow raking SE, proportional to height
			blendPixel(img, sx+d, sy+d, shadow, 0.40)
		}
		for dy := -h; dy <= 0; dy++ {
			frac := float64(-dy) / float64(maxInt(h, 1))
			setPixel(img, sx, sy+dy, metal)
			if frac < 0.7 {
				setPixel(img, sx-1, sy+dy, metalLit)
				setPixel(img, sx+1, sy+dy, metalDark)
			}
		}
		setPixel(img, sx, sy-h, brighten(metalLit, 0.16)) // lit tip
		setPixel(img, sx, sy-h-1, brighten(metalLit, 0.22))
	}

	// SATELLITE SPIRES: a ring of shorter needles around center, drawn before the central one so the tallest
	// core overpaints in front. Placed on a circle strictly inside the apron.
	ringR := maxInt(rad*3/5, 1)
	satH := maxInt(rad*6/5, 2)
	for i := 0; i < 5; i++ {
		ang := 2*math.Pi*float64(i)/5 - math.Pi/2 // start at the top, evenly spaced
		sx := cx + int(math.Round(math.Cos(ang)*float64(ringR)))
		sy := cy + int(math.Round(math.Sin(ang)*float64(ringR)))
		spire(sx, sy, satH)
	}
	// THE CENTRAL SPIRE: the tallest needle at the apron center, drawn last (in front).
	spire(cx, cy, maxInt(rad*2, 3))
}

// drawRoofRingHub: the GALACTIC wonder (COSMIC epoch) — THE signature ringworld / orbital MEGASTATION read
// from above: 2–3 concentric bright metallic RING outlines at increasing radii, a glowing energetic central
// HUB, and faint metallic SPOKES from the hub out to the rings. Reads clearly apart from the launchpad /
// fusion-core / dome / skyscraper. Metallic silver rings + a cyan-energy hub, all base/theme-derived so the
// whole megastation retints with the roof material.
func drawRoofRingHub(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	deck := blend(rc.dark, metalSilverAnchor, 0.24)           // a dim metal deck behind the rings
	ringMetal := blend(rc.base, metalSilverAnchor, 0.62)      // the bright metal ring bands
	ringLit := brighten(ringMetal, 0.16)                      // sunlit NW arc of each ring
	spoke := blend(ringMetal, energyAnchor, 0.30)             // faint energetic spokes
	hub := blend(rc.base, energyAnchor, 0.55)                 // the glowing energetic hub
	hubCore := brighten(blend(hub, energyAnchor, 0.40), 0.14) // the bright hub core

	rad := maxInt(minInt(hw, hh), 1)

	// A dim containment DECK filling the footprint so the rings sit on a defined station platform (lit NW /
	// shaded SE for a shallow dish), not on bare ground.
	forEllipse(cx, cy, rad, rad, func(x, y int) {
		if x <= cx || y <= cy {
			setPixel(img, x, y, brighten(deck, 0.06))
		} else {
			setPixel(img, x, y, deck)
		}
	})

	// FAINT SPOKES: four thin energetic radials from the hub out toward the outer ring (N/E/S/W), drawn
	// under the ring bands so the rings overpaint the spoke ends cleanly.
	outerR := maxInt(rad*9/10, 1)
	for _, dir := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		for d := 1; d <= outerR; d++ {
			blendPixel(img, cx+dir[0]*d, cy+dir[1]*d, spoke, 0.50)
		}
	}

	// CONCENTRIC RING BANDS: three bright metal ring outlines at increasing radii, each a hollow circle
	// (drawn as a circle outline, NOT a filled disc, so the deck + spokes show through the gaps between
	// rings). The NW arc of each is lit brighter for a metallic sheen.
	for _, f := range []float64{0.42, 0.66, 0.90} {
		rr := maxInt(int(float64(rad)*f), 1)
		steps := maxInt(rr*6, 24)
		for i := 0; i < steps; i++ {
			ang := 2 * math.Pi * float64(i) / float64(steps)
			rx := cx + int(math.Round(math.Cos(ang)*float64(rr)))
			ry := cy + int(math.Round(math.Sin(ang)*float64(rr)))
			c := ringMetal
			if rx <= cx && ry <= cy { // NW arc catches the light
				c = ringLit
			}
			setPixel(img, rx, ry, c)
		}
	}

	// THE GLOWING HUB: a bright energetic core disc at the very center with a soft halo bleeding out, so the
	// station heart reads as a powered core, not a metal dot.
	hubR := maxInt(rad/4, 1)
	fillDisc(img, cx, cy, hubR, hub)
	haloR := maxInt(rad*2/5, 1)
	forEllipse(cx, cy, haloR, haloR, func(x, y int) {
		fx := float64(x-cx) / float64(haloR)
		fy := float64(y-cy) / float64(haloR)
		blendPixel(img, x, y, hubCore, 0.45*(1-(fx*fx+fy*fy)))
	})
	setPixel(img, cx, cy, brighten(hubCore, 0.16)) // the brightest pinpoint at the hub center
}

// drawRoofLattice: the QUANTUM dwelling (COSMIC epoch) — a small faceted CRYSTAL / lattice NODE read from
// above. Instead of a roof it draws a little geometric GEM: four triangular facets radiating from a bright
// central core, each facet a DIFFERENT iridescent hue (cyan / magenta / gold, cycled by facet index) so the
// crystal SHIMMERS rather than sits one color, with a bright glinting core. The iridescent anchors are
// BLENDED with the passed roof colors so it retints on a theme switch. Bounds-safe: setPixel / blendPixel
// only (both clipped), so it never panics at any footprint.
func drawRoofLattice(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	rad := maxInt(minInt(hw, hh), 1)

	// A faint dark facet backing so the gem sits on a defined footprint (a shadowed crystal base), lit NW.
	base := darken(blend(rc.dark, iridCyanAnchor, 0.20), 0.10)
	forEllipse(cx, cy, rad, rad, func(x, y int) {
		if x <= cx && y <= cy {
			blendPixel(img, x, y, brighten(base, 0.08), 0.9)
		} else {
			blendPixel(img, x, y, base, 0.9)
		}
	})

	// FOUR TRIANGULAR FACETS: a diamond of four quadrant triangles (N / E / S / W), each filled a different
	// shifting iridescent hue blended into the roof tone, so adjacent facets read as a faceted, prismatic
	// crystal catching the light differently on each face. Painted by scanning the footprint and picking the
	// facet from the pixel's quadrant + parity.
	for y := cy - rad; y <= cy+rad; y++ {
		for x := cx - rad; x <= cx+rad; x++ {
			ddx := x - cx
			ddy := y - cy
			// Diamond mask: |dx|+|dy| <= rad → inside the crystal facet body.
			if absInt(ddx)+absInt(ddy) > rad {
				continue
			}
			// Facet index from quadrant (N=0, E=1, S=2, W=3), nudged by parity so the sheen shifts within a
			// facet too — the shimmer.
			var q int
			switch {
			case absInt(ddx) >= absInt(ddy) && ddx >= 0:
				q = 1 // E
			case absInt(ddx) >= absInt(ddy):
				q = 3 // W
			case ddy < 0:
				q = 0 // N
			default:
				q = 2 // S
			}
			hue := iridHueFor(q + ((absInt(ddx) + absInt(ddy)) & 1))
			facet := blend(rc.base, hue, 0.62)
			if ddx+ddy < 0 { // NW faces catch more light
				facet = brighten(facet, 0.12)
			}
			setPixel(img, x, y, facet)
		}
	}

	// The bright glinting CORE + facet ridge lines from the core out to the four points, so the gem reads
	// cut, not a blob. Core is the brightest pinpoint.
	core := brighten(blend(rc.base, blend(iridCyanAnchor, iridGoldAnchor, 0.5), 0.5), 0.16)
	for d := 0; d <= rad; d++ {
		ridge := brighten(blend(rc.base, iridHueFor(d), 0.5), 0.10)
		setPixel(img, cx, cy-d, ridge) // N edge
		setPixel(img, cx, cy+d, ridge) // S edge
		setPixel(img, cx-d, cy, ridge) // W edge
		setPixel(img, cx+d, cy, ridge) // E edge
	}
	setPixel(img, cx, cy, core)
	setPixel(img, cx-1, cy-1, brighten(core, 0.10)) // a lit NW glint off the core
}

// drawRoofCrystalLattice: the QUANTUM wonder (COSMIC epoch) — a large geometric CRYSTAL LATTICE MESH read
// from above: a grid/web of glowing NODES joined by thin IRIDESCENT lines, the hues SHIFTING across the mesh
// (cyan → magenta → gold by node position) around a bright CENTRAL node. Unlike the ring-hub (concentric
// metal rings + a hub) or the fusion core (radiant filled discs) this is an open LATTICE of discrete nodes +
// connecting struts — no rings, no filled target — so it reads clearly apart from every other wonder. The
// three iridescent anchors are BLENDED with the passed roof colors so the whole lattice retints on a theme
// switch. Bounds-safe: fillDisc / setPixel / blendPixel / drawLineC (all clipped), so it never panics.
func drawRoofCrystalLattice(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	rad := maxInt(minInt(hw, hh), 1)

	// A faint dark crystalline DECK filling the footprint so the mesh sits on a defined black-glass platform
	// (lit NW / shaded SE for a shallow facet), not on bare ground.
	deck := darken(blend(rc.dark, iridMagentaAnchor, 0.18), 0.08)
	forEllipse(cx, cy, rad, rad, func(x, y int) {
		if x <= cx || y <= cy {
			blendPixel(img, x, y, brighten(deck, 0.06), 0.85)
		} else {
			blendPixel(img, x, y, deck, 0.85)
		}
	})

	// LATTICE NODES: a small diamond grid of nodes at lattice coordinates (a 5-node diamond: center + 4 mid +
	// 4 outer corners), each a little glowing disc whose hue SHIFTS by its lattice index so the mesh cycles
	// cyan → magenta → gold across itself. Built as offsets in units of the node step.
	step := maxInt(rad/2, 1)
	type node struct {
		dx, dy, idx int
	}
	nodes := []node{
		{0, 0, 0},                                    // center (drawn brightest, last)
		{0, -1, 1}, {1, 0, 2}, {0, 1, 3}, {-1, 0, 4}, // the four mid nodes (N/E/S/W)
		{-1, -1, 5}, {1, -1, 6}, {1, 1, 7}, {-1, 1, 8}, // the four diagonal corner nodes
	}

	// STRUTS FIRST: thin iridescent lines joining each outer/mid node back toward the center + around the ring,
	// drawn under the nodes so the nodes cap the strut ends cleanly. The strut hue is the blend of its two
	// endpoints' hues, kept faint (a glowing wire, not a solid bar).
	nodePx := func(n node) (int, int) { return cx + n.dx*step, cy + n.dy*step }
	for _, n := range nodes[1:] {
		nx, ny := nodePx(n)
		strut := blend(rc.base, iridHueFor(n.idx), 0.44)
		drawLineC(img, cx, cy, nx, ny, strut) // spoke to center
	}
	// Ring struts joining adjacent outer nodes (a woven mesh perimeter).
	ringOrder := []node{nodes[1], nodes[6], nodes[2], nodes[7], nodes[3], nodes[8], nodes[4], nodes[5]}
	for i := range ringOrder {
		ax, ay := nodePx(ringOrder[i])
		bx, by := nodePx(ringOrder[(i+1)%len(ringOrder)])
		strut := blend(rc.base, iridHueFor(i), 0.36)
		drawLineC(img, ax, ay, bx, by, strut)
	}

	// THE GLOWING NODES: each a small disc a shade brighter than its strut, hue cycling by lattice index, with
	// a faint halo so it reads as a light-emitting node. Outer/mid nodes first, then the bright central node.
	nodeR := maxInt(rad/6, 1)
	for _, n := range nodes[1:] {
		nx, ny := nodePx(n)
		glow := brighten(blend(rc.base, iridHueFor(n.idx), 0.6), 0.10)
		fillDisc(img, nx, ny, nodeR, glow)
		setPixel(img, nx, ny, brighten(glow, 0.14))
	}
	// THE CENTRAL NODE: the brightest, a white-cored iridescent bloom with a soft halo — the lattice heart.
	centerGlow := brighten(blend(rc.base, blend(iridCyanAnchor, iridMagentaAnchor, 0.5), 0.66), 0.14)
	fillDisc(img, cx, cy, maxInt(nodeR+1, 1), centerGlow)
	haloR := maxInt(rad/2, 1)
	for y := cy - haloR; y <= cy+haloR; y++ {
		for x := cx - haloR; x <= cx+haloR; x++ {
			fx := float64(x-cx) / float64(haloR)
			fy := float64(y-cy) / float64(haloR)
			if fx*fx+fy*fy <= 1.0 {
				blendPixel(img, x, y, centerGlow, 0.40*(1-(fx*fx+fy*fy)))
			}
		}
	}
	setPixel(img, cx, cy, brighten(centerGlow, 0.18)) // the brightest pinpoint at the lattice heart
}

// drawRoofEthereal: the TRANSCENDENT dwelling (COSMIC epoch, the finale) — a soft glowing translucent BLOOM
// instead of a hard roof. The dwelling has DEMATERIALISED into light: a luminous white light-form that
// paints NOTHING solid — it BLENDS a soft warm-white glow over whatever is beneath, brightest at the center
// and fading to nothing at the rim (a radial falloff), so the ground shows THROUGH it. The ethereal white +
// a soft gold cast are BLENDED with the roof tone so it retints on a theme switch. Bounds-safe: blendPixel
// only (clipped), so it never panics at any footprint.
func drawRoofEthereal(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	rad := maxInt(minInt(hw, hh), 1)
	glow := blend(blend(rc.base, etherWhiteAnchor, 0.70), etherGoldAnchor, 0.14)
	core := brighten(glow, 0.14)

	// The SOFT BLOOM: a radial translucent glow, strongest at the center, fading to zero at the rim, so the
	// light-form reads as a hovering pool of light rather than a solid disc. A gentle NW lift so it isn't flat.
	for y := cy - rad; y <= cy+rad; y++ {
		for x := cx - rad; x <= cx+rad; x++ {
			fx := float64(x-cx) / float64(rad)
			fy := float64(y-cy) / float64(rad)
			d2 := fx*fx + fy*fy
			if d2 > 1.0 {
				continue
			}
			t := 0.62 * (1 - d2) // translucent, brightest at center
			if fx+fy < 0 {       // NW half a touch brighter — a soft lit crown
				t += 0.06 * (1 - d2)
			}
			blendPixel(img, x, y, glow, t)
		}
	}
	// A bright soft CORE pinpoint — the focus of the light-form (still translucent, not a hard dot).
	blendPixel(img, cx, cy, core, 0.85)
	blendPixel(img, cx-1, cy-1, brighten(core, 0.10), 0.55) // a faint NW glint
}

// drawRoofAscension: the TRANSCENDENT wonder (COSMIC epoch) — the ASCENSION OF LIGHT, the luminous climax of
// the whole 22-age progression. A bright vertical PILLAR / beam of light rises at center, ringed by
// concentric soft glowing HALOS radiating outward (translucent blends, brightening inward to a PURE-WHITE
// core) — an ethereal gate. Unlike the crystal lattice (a hard node mesh), the ring-hub (metal rings), or
// the fusion core (filled cyan discs), this is a soft translucent WHITE bloom + a rising vertical beam — no
// hard geometry — so it reads clearly apart from every other wonder as pure light. The ethereal white + soft
// gold are BLENDED with the roof tone so it retints on a theme switch. Bounds-safe: blendPixel / setPixel /
// forEllipse only (all clipped), so it never panics.
func drawRoofAscension(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	rad := maxInt(minInt(hw, hh), 1)
	halo := blend(blend(rc.base, etherWhiteAnchor, 0.60), etherGoldAnchor, 0.16) // the warm glowing halo
	white := blend(rc.base, etherWhiteAnchor, 0.82)                              // the near-pure-white inner tone
	bloom := brighten(white, 0.16)                                               // the incandescent core + beam

	// CONCENTRIC SOFT HALOS: several translucent rings radiating out from the center, each a soft glowing
	// annulus (a radial falloff peaking at the ring radius), brightening toward the center — an ethereal gate
	// of light. Drawn as translucent blends so the ground glows through, not a stack of solid discs.
	for _, f := range []float64{0.92, 0.68, 0.44} {
		rr := maxInt(int(float64(rad)*f), 1)
		for y := cy - rr; y <= cy+rr; y++ {
			for x := cx - rr; x <= cx+rr; x++ {
				fx := float64(x-cx) / float64(rr)
				fy := float64(y-cy) / float64(rr)
				d2 := fx*fx + fy*fy
				if d2 > 1.0 {
					continue
				}
				// Brighter inner halos; each fades from its center outward.
				c := blend(halo, white, 1-f)
				blendPixel(img, x, y, c, 0.34*(1-d2))
			}
		}
	}

	// THE RISING PILLAR / BEAM: a bright vertical column of light rising well above center, brightest at its
	// base (the source) and fading as it ascends, with a faint width so it reads as a beam, not a wire —
	// the ascension itself. Translucent so it glows rather than paints a bar.
	tall := maxInt(rad*9/5, 3)
	for dy := -tall; dy <= 0; dy++ {
		frac := float64(-dy) / float64(tall) // 0 at base .. 1 at top
		t := 0.75 * (1 - 0.7*frac)
		blendPixel(img, cx, cy+dy, bloom, t)
		if frac < 0.75 { // a soft flanking glow low + mid, tapering out near the top
			blendPixel(img, cx-1, cy+dy, white, t*0.55)
			blendPixel(img, cx+1, cy+dy, white, t*0.55)
		}
	}

	// THE PURE-WHITE CORE: a small incandescent gate at center with a soft halo bleeding out, so the source of
	// the ascension reads as pure light.
	coreR := maxInt(rad/4, 1)
	forEllipse(cx, cy, coreR, coreR, func(x, y int) {
		fx := float64(x-cx) / float64(maxInt(coreR, 1))
		fy := float64(y-cy) / float64(maxInt(coreR, 1))
		blendPixel(img, x, y, bloom, 0.85*(1-0.4*(fx*fx+fy*fy)))
	})
	// The lit tip crowning the ascending beam + the brightest pinpoint at the gate.
	setPixel(img, cx, cy-tall, brighten(bloom, 0.10))
	setPixel(img, cx, cy-tall-1, brighten(bloom, 0.18))
	blendPixel(img, cx, cy, brighten(bloom, 0.14), 0.9)
}

// drawRoofLaunchpad: the SPACE wonder (NEON epoch) — a rocket on a LAUNCH PAD seen from above: a circular
// PAD, a bright central ROCKET (a lit metallic capsule with a pointed nose + a pair of splayed FINS), a few
// GANTRY dabs flanking it, and a dark SCORCH RING blasted into the pad around the base. Reads clearly apart
// from the fusion core (radiant rings), the skyscraper slab, and the dome: this is a pad + a discrete
// rocket + fins + scorch. Silver + a warm exhaust glow (metalSilverAnchor / gasGlowAnchor) are BLENDED with
// the passed roof colors so it retints on a theme switch. Bounds-safe: fillDisc / forEllipse / setPixel /
// drawHSpan / blendPixel only (all clipped), so it never panics.
func drawRoofLaunchpad(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	pad := blend(rc.base, metalSilverAnchor, 0.42)                    // the pale plated pad
	padDark := blend(rc.dark, metalSilverAnchor, 0.30)                // shaded pad + rim
	rocket := blend(brighten(rc.base, 0.10), metalSilverAnchor, 0.58) // bright metal capsule
	rocketLit := brighten(rocket, 0.16)                               // sunlit NW rim
	rocketDark := darken(rocket, 0.18)                                // shaded SE rim
	scorch := darken(padDark, 0.30)                                   // dark blast scorch
	gantry := darken(pad, 0.34)                                       // dark gantry lattice

	rad := maxInt(minInt(hw, hh), 1)

	// THE PAD: a filled silver disc filling the footprint, lit NW / shaded SE so it reads as a raised plated
	// apron rather than flat ground.
	forEllipse(cx, cy, rad, rad, func(x, y int) {
		if x <= cx || y <= cy {
			setPixel(img, x, y, pad)
		} else {
			setPixel(img, x, y, padDark)
		}
	})

	// SCORCH RING: a dark blast ring stamped into the pad around the rocket base (where the exhaust burns).
	scorchR := maxInt(rad*3/5, 1)
	for i := 0; i < 56; i++ {
		ang := 2 * math.Pi * float64(i) / 56
		setPixel(img, cx+int(math.Round(math.Cos(ang)*float64(scorchR))), cy+int(math.Round(math.Sin(ang)*float64(scorchR))), scorch)
	}

	// GANTRY dabs: a couple of dark lattice service towers flanking the rocket (W + E of the pad center).
	for _, dx := range []int{-rad * 3 / 5, rad * 3 / 5} {
		gx := cx + dx
		for dy := -rad / 2; dy <= rad/3; dy++ {
			setPixel(img, gx, cy+dy, gantry)
		}
		setPixel(img, gx, cy-rad/2-1, brighten(gantry, 0.14)) // a lit tip light
	}

	// THE ROCKET: a bright vertical capsule standing at pad center — a lit metal body from the pad up toward a
	// pointed nose, lit on the NW edge + shaded on the SE edge for round volume.
	tall := maxInt(rad*4/5, 2) // the capsule rises well above the pad center
	for dy := -tall; dy <= rad/3; dy++ {
		setPixel(img, cx, cy+dy, rocket)
		setPixel(img, cx-1, cy+dy, rocketLit)
		setPixel(img, cx+1, cy+dy, rocketDark)
	}
	// The pointed NOSE converging above the body.
	setPixel(img, cx, cy-tall, brighten(rocketLit, 0.12))
	setPixel(img, cx, cy-tall-1, brighten(rocketLit, 0.20))
	// The splayed FINS: two short diagonal dabs at the base flaring out to the SW + SE.
	for d := 1; d <= maxInt(rad/3, 1); d++ {
		setPixel(img, cx-1-d, cy+rad/3+d, rocketDark) // SW fin
		setPixel(img, cx+1+d, cy+rad/3+d, rocketDark) // SE fin
	}
	// A warm EXHAUST glow blooming under the fins at the pad center.
	forEllipse(cx, cy+rad/3, maxInt(rad/3, 1), maxInt(rad/4, 1), func(x, y int) {
		blendPixel(img, x, y, blend(rocketLit, gasGlowAnchor, 0.65), 0.45)
	})
}

// drawRoofFactory: the INDUSTRIAL wonder (Phase 1b-iii) — a great FACTORY HALL read from above: a big
// rectangular red-brick/tin hall filling the footprint (a sawtooth roof suggested by parallel ridge
// lines), with 2–3 tall SMOKESTACKS standing along its NORTH edge — each a dark vertical chimney with
// a lighter sunlit rim and a soft dark SOOT/smoke puff drifting off its top. The blocky hall + the row
// of chimneys + smoke reads clearly apart from the stepped ziggurat, the colonnade+pediment temple,
// the cruciform+spire cathedral, the domed rotunda, the blocky keep+tower, and the open megalith.
// Brick + tin + soot (brickRedAnchor / tinAnchor / sootAnchor) are BLENDED with the passed roof colors
// so the factory retints on a theme switch and stays in the grimy era mood. Bounds-safe: every write
// goes through fillRectC/forRect/drawHSpan/setPixel/blendPixel (all clipped), so it never panics.
func drawRoofFactory(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Brick hall body + dull-tin roof + dark soot tones pulled toward the (already theme/lineage-tinted)
	// roof colors, so the whole works retints yet reads as a grimy brick-and-tin factory.
	brick := blend(rc.base, brickRedAnchor, 0.46)                   // the red-brick hall walls
	tin := blend(rc.base, blend(tinAnchor, sootAnchor, 0.30), 0.50) // dull grey tin roof panels
	tinLit := brighten(tin, 0.12)                                   // sunlit tin ridge
	tinDark := blend(rc.dark, sootAnchor, 0.35)                     // shaded tin + eave
	stack := blend(rc.dark, sootAnchor, 0.55)                       // dark soot chimney
	stackLit := brighten(stack, 0.18)                               // sunlit NW rim of a stack
	smoke := blend(rc.dark, sootAnchor, 0.42)                       // soft dark smoke puff

	// THE HALL: a solid brick body filling the footprint, lit N/W + shaded S/E, so the works sit as a
	// raised mass rather than a flat slab.
	forRect(cx, cy, hw, hh, func(x, y int) {
		if x <= cx || y <= cy {
			setPixel(img, x, y, brick)
		} else {
			setPixel(img, x, y, darken(brick, 0.14))
		}
	})
	// TIN ROOF: a broad tin deck over most of the hall (inset a brick rim), then a SAWTOOTH read — a
	// few parallel bright ridge lines across the deck so it reads as a factory shed roof, not a plain slab.
	dhw := maxInt(hw-1, 0)
	dhh := maxInt(hh-1, 0)
	forRect(cx, cy, dhw, dhh, func(x, y int) {
		if y <= cy {
			setPixel(img, x, y, tin)
		} else {
			setPixel(img, x, y, tinDark)
		}
	})
	sawEvery := maxInt(hh/2, 2)
	for y := cy - dhh; y <= cy+dhh; y += sawEvery {
		drawHSpan(img, cx-dhw, cx+dhw, y, tinLit)
	}

	// SMOKESTACKS: 2–3 tall chimneys standing along the hall's NORTH edge, spaced across the width. Each
	// is a short dark vertical column rising from the roofline, with a lit NW rim, and a soft dark soot
	// puff blended above its top so the skyline reads industrial.
	stacks := 2
	if hw >= 8 {
		stacks = 3
	}
	stackHW := maxInt(hw/12, 0) // half-width of a chimney column
	stackH := maxInt(hh*3/4, 2) // how far a chimney rises above the roofline (tall)
	topY := cy - hh - stackH    // the chimney top (may clip above the footprint; setPixel clamps)
	for i := 0; i < stacks; i++ {
		// Even spread across the north edge, kept inside the hall width.
		var sx int
		if stacks == 1 {
			sx = cx
		} else {
			sx = cx - hw + (2*hw)*i/(stacks-1)
		}
		// The chimney column: from the north roofline up to topY.
		fillRectC(img, sx, (cy-hh+topY)/2, stackHW, (cy-hh-topY)/2, stack)
		// Lit NW rim + shaded SE edge for upright volume.
		for y := topY; y <= cy-hh; y++ {
			setPixel(img, sx-stackHW, y, stackLit)
			setPixel(img, sx+stackHW, y, darken(stack, 0.16))
		}
		// SOOT / smoke puff: a soft dark blob drifting up-right off the chimney top.
		puffR := maxInt(stackHW+1, 1)
		forEllipse(sx+puffR, topY-puffR, puffR, puffR, func(x, y int) { blendPixel(img, x, y, smoke, 0.55) })
		setPixel(img, sx, topY, brighten(stack, 0.10)) // a lit chimney lip
	}
}

// drawRoofTower: the ELECTRIC/ATOMIC wonder (V3-B ELECTRIC epoch) — an ART-DECO SETBACK tower read from
// above as a set of CONCENTRIC FLAT RECTILINEAR TIERS stepping in toward a small central SPIRE / MAST.
// Like the ziggurat it steps, but the read is deliberately CLEANER + FLATTER + PALER: crisp square
// setbacks in pale CONCRETE (not earthen terraces), each tier LIGHTER as it rises (a tower catching more
// light up top), with a thin dark shadow lip on each tier's south+east so the setbacks read as sheer
// drops — a strong sense of HEIGHT. A central mast/finial dab crowns it. Reads clearly apart from the
// earthen-stepped ziggurat, the round dome, the blocky keep+tower, the colonnade temple, and the factory
// hall+chimneys. Concrete/steel (concreteAnchor / steelAnchor) are BLENDED with the passed roof colors so
// the tower retints on a theme switch and stays in the era mood. Bounds-safe: every write goes through
// fillRectC / forRect+setPixel / drawHSpan / setPixel (all clipped), so it never panics at any footprint.
func drawRoofTower(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Pale concrete + cool steel tones pulled toward the (already theme/lineage-tinted) roof colors.
	concrete := blend(rc.base, blend(concreteAnchor, steelAnchor, 0.30), 0.52) // pale deco concrete face
	concreteDark := blend(rc.dark, steelAnchor, 0.34)                          // shaded setback edge
	crown := brighten(concrete, 0.18)                                          // the bright top tier / mast catching light

	// An extra deep shadow lip UNDER the tower base (south-east) sells the height — a taller mass throws a
	// longer shadow than the pitched houses around it. Blended into the ground, so it darkens rather than
	// paints a slab.
	baseShadow := darken(concreteDark, 0.30)
	for dy := 1; dy <= maxInt(hh/3, 1); dy++ {
		drawHSpan(img, cx-hw+dy, cx+hw+dy, cy+hh+dy, baseShadow)
	}

	// CONCENTRIC FLAT SETBACK TIERS: crisp square tiers shrinking toward the crown, each a shade LIGHTER as
	// it rises. A thin darker step lip on each tier's south+east edge makes the setbacks read as sheer drops.
	const tiers = 4
	for t := 0; t < tiers; t++ {
		f := float64(t) / float64(tiers) // 0 (base) .. →1 (top)
		thw := maxInt(int(float64(hw)*(1-f)), 1)
		thh := maxInt(int(float64(hh)*(1-f)), 1)
		// Rising tiers lighten from the shaded concrete toward the bright crown.
		col := blend(concreteDark, concrete, f/(1-1.0/tiers))
		if t == tiers-1 {
			col = crown
		}
		fillRectC(img, cx, cy, thw, thh, col)
		// Sheer-drop step lip on the south + east faces of each tier.
		drawHSpan(img, cx-thw, cx+thw, cy+thh, darken(col, 0.22))
		for y := cy - thh; y <= cy+thh; y++ {
			setPixel(img, cx+thw, y, darken(col, 0.22))
		}
		// A crisp lit NW corner on each tier for the clean deco sheen.
		setPixel(img, cx-thw, cy-thh, brighten(col, 0.14))
	}

	// CENTRAL SPIRE / MAST: a small bright vertical finial at the very crown — the deco tower's mast.
	mastH := maxInt(minInt(hw, hh)/2, 1)
	for dy := -mastH; dy <= 0; dy++ {
		setPixel(img, cx, cy+dy, crown)
	}
	setPixel(img, cx, cy-mastH, brighten(crown, 0.16)) // a lit mast tip
	setPixel(img, cx, cy, crown)
}

// drawRoofSpaceNeedle: the ATOMIC wonder (Phase 1i) — a googie / SPACE-AGE space needle read from above.
// Where the deco wonderTower is a WIDE stack of concentric flat concrete squares, this is a slender vertical
// STEM crowned by a wide round flying-SAUCER disc: a small footing pad at the base, a thin bright stem rising
// north, a broad steel saucer DISC near the top (a shaded rim, a lit NW arc, a bright energetic core + a soft
// halo), and a short mast finial above — plus a LONG raking SE height shadow (a lone tall needle throws the
// map's longest shadow). Reads clearly apart from the deco tower (square tiers), the round renaissance dome
// (a solid hemisphere), and the glass skyscraper (a slab + window grid). Steel + a cool energetic core are
// BLENDED with the passed roof colors so it retints on a theme switch. Bounds-safe: every write goes through
// setPixel / fillDisc / forEllipse / drawHSpan / blendPixel (all clipped), so it never panics at any footprint.
func drawRoofSpaceNeedle(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	steel := blend(rc.base, steelAnchor, 0.56)                            // the pale steel needle body / saucer face
	steelLit := brighten(steel, 0.18)                                     // sunlit NW rim / lit stem edge
	steelDark := blend(rc.dark, steelAnchor, 0.40)                        // shaded SE rim / footing
	stem := blend(steel, steelLit, 0.5)                                   // the bright thin stem catching light
	core := blend(rc.base, blend(pastelAnchor, energyAnchor, 0.45), 0.62) // the cool lit saucer core
	coreLit := brighten(blend(core, energyAnchor, 0.30), 0.14)            // the bright core pinpoint / ring

	rad := maxInt(minInt(hw, hh), 1)

	// LONG HEIGHT SHADOW: a raking SE drop well beyond the footprint — a lone tall needle casts the longest
	// shadow on the map. Blended into the ground so it darkens rather than paints a slab.
	needleShadow := darken(steelDark, 0.34)
	shR := maxInt(rad/2, 1) // the shadow tracks the saucer's width, not the whole footprint
	for dy := 1; dy <= maxInt(hh, 2); dy++ {
		drawHSpan(img, cx-shR+dy, cx+shR+dy, cy+hh+dy, needleShadow)
	}

	// FOOTING PAD: a small shaded steel disc at the base so the needle plants on a launch pad, not bare ground.
	padR := maxInt(rad/2, 1)
	forEllipse(cx, cy+hh-padR, padR, maxInt(padR/2, 1), func(x, y int) {
		if x <= cx || y <= cy {
			setPixel(img, x, y, steel)
		} else {
			setPixel(img, x, y, steelDark)
		}
	})

	// THE STEM: a thin bright vertical shaft rising from the footing up to the saucer height — the needle's
	// spine, lit on the NW edge + shaded on the SE edge for round volume.
	saucerCY := cy - hh + maxInt(rad*2/5, 1) // the saucer rides high on the stem (upper third)
	for y := saucerCY; y <= cy+hh-1; y++ {
		setPixel(img, cx, y, stem)
		setPixel(img, cx-1, y, steelLit)
		setPixel(img, cx+1, y, steelDark)
	}

	// THE SAUCER DISC: a WIDE round platter near the top — the space-age tell. A flat steel disc (lit NW /
	// shaded SE), a bright lit RING one band in, then a glowing energetic CORE with a soft halo at the hub.
	saucerR := maxInt(rad*4/5, 2)     // wide — the widest thing on the needle
	saucerH := maxInt(saucerR*3/5, 1) // squashed vertically so it reads as a disc seen from above
	forEllipse(cx, saucerCY, saucerR, saucerH, func(x, y int) {
		if x <= cx && y <= saucerCY {
			setPixel(img, x, y, steelLit)
		} else if x > cx && y > saucerCY {
			setPixel(img, x, y, steelDark)
		} else {
			setPixel(img, x, y, steel)
		}
	})
	// A bright lit RING band just inside the rim (an outline circle so the saucer face shows through inside).
	ringR := maxInt(saucerR*3/4, 1)
	ringH := maxInt(saucerH*3/4, 1)
	forEllipse(cx, saucerCY, ringR, ringH, func(x, y int) {
		// keep only the ~1px rim band: skip the solid interior
		fx := float64(x-cx) / float64(maxInt(ringR, 1))
		fy := float64(y-saucerCY) / float64(maxInt(ringH, 1))
		if fx*fx+fy*fy < 0.55 {
			return
		}
		c := blend(steel, coreLit, 0.5)
		if x <= cx && y <= saucerCY {
			c = coreLit
		}
		setPixel(img, x, y, c)
	})
	// THE CORE: a bright energetic hub disc at the saucer center with a soft halo bleeding out, so the needle
	// heart reads as a lit observation pod, not a metal dot.
	coreR := maxInt(saucerR/3, 1)
	fillDisc(img, cx, saucerCY, coreR, core)
	haloR := maxInt(saucerR/2, 1)
	forEllipse(cx, saucerCY, haloR, haloR, func(x, y int) {
		fx := float64(x-cx) / float64(haloR)
		fy := float64(y-saucerCY) / float64(haloR)
		blendPixel(img, x, y, coreLit, 0.45*(1-(fx*fx+fy*fy)))
	})
	setPixel(img, cx, saucerCY, brighten(coreLit, 0.16)) // the brightest pinpoint at the saucer hub

	// MAST FINIAL: a short thin bright spire rising off the saucer crown — the needle's antenna tip.
	mastH := maxInt(rad/3, 1)
	for dy := 1; dy <= mastH; dy++ {
		setPixel(img, cx, saucerCY-saucerH-dy, stem)
	}
	setPixel(img, cx, saucerCY-saucerH-mastH, brighten(coreLit, 0.20)) // a lit beacon at the mast tip
}

// drawRoofCathedral: the MEDIEVAL wonder (locked #13, V3-B) — a tall CATHEDRAL / KEEP read from
// above: a long cruciform nave (a broad body with a shorter transept crossing it) topped by a
// central SPIRE dab, so it reads as a great church/keep rather than the rounded default complex.
// All tones base/dark/ridge-derived (no accent — the spire is a base-derived lighten, not a
// saturated finial). The cross-plan + spire make it unmistakably a cathedral from above.
func drawRoofCathedral(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// The NAVE: a long body along the wider axis. The TRANSEPT: a shorter arm across it. Together a
	// cross plan. Slate slopes shade off each ridge.
	horizontal := hw >= hh
	naveHW, naveHH := hw, maxInt(hh*3/5, 1)
	tranHW, tranHH := maxInt(hw*3/5, 1), hh
	if !horizontal {
		naveHW, naveHH = maxInt(hw*3/5, 1), hh
		tranHW, tranHH = hw, maxInt(hh*3/5, 1)
	}
	// Nave body (pitched: lit north/west, shaded south/east).
	forRect(cx, cy, naveHW, naveHH, func(x, y int) {
		lit := y <= cy
		if !horizontal {
			lit = x <= cx
		}
		if lit {
			img.SetRGBA(x, y, rc.base)
		} else {
			img.SetRGBA(x, y, rc.dark)
		}
	})
	// Transept arm.
	forRect(cx, cy, tranHW, tranHH, func(x, y int) {
		lit := x <= cx
		if !horizontal {
			lit = y <= cy
		}
		if lit {
			img.SetRGBA(x, y, rc.base)
		} else {
			img.SetRGBA(x, y, rc.dark)
		}
	})
	// Ridge lines along both arms.
	drawHSpan(img, cx-naveHW, cx+naveHW, cy, rc.ridge)
	for y := cy - tranHH; y <= cy+tranHH; y++ {
		img.SetRGBA(cx, y, rc.ridge)
	}
	// Central SPIRE: a small bright base-derived dab at the crossing (a steeple seen from above).
	shw := maxInt(hw/5, 1)
	shh := maxInt(hh/5, 1)
	forRect(cx, cy, shw, shh, func(x, y int) { img.SetRGBA(x, y, rc.ridge) })
	setPixel(img, cx, cy, brighten(rc.ridge, 0.12))
}

// drawRoofMegalith: the STONE-AGE wonder (Phase 1b-i) — a megalithic monument read from above as a
// rough RING of upright standing stones around the footprint, a couple of lintel-topped TRILITHON
// pairs (a cap stone bridging two uprights, for the Stonehenge read), and a low central ALTAR/hearth
// stone. The grey stone palette is graniteAnchor/stoneAnchor BLENDED with the passed roof color set
// (like the ziggurat pulls its terraces from rc) so the monument retints on a theme switch and stays
// in-family with the era mood rather than reading as raw grey. Deterministic + bounds-checked (every
// write goes through setPixel/drawBlock/fillDisc), so it is panic-safe at any footprint.
func drawRoofMegalith(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Stone tones: the grey anchors pulled toward the (already theme/lineage-tinted, sat-capped)
	// roof base + shaded slope, so the megalith reads grey-stone yet retints with the theme. Kept
	// PALE so the standing stones read with strong contrast against the dark mound below.
	stone := brighten(blend(rc.base, blend(graniteAnchor, stoneAnchor, 0.5), 0.62), 0.10)
	stoneLit := brighten(stone, 0.16) // NW-lit crown of an upright
	stoneDark := blend(rc.dark, graniteAnchor, 0.45)
	lintel := blend(stone, stoneLit, 0.5) // cap stones a touch lighter so the trilithons read

	// MASS FIRST: a filled earthen mound + a paved inner ring, so the monument has the SOLID visual
	// weight of the ziggurat/cathedral (it fills its footprint) instead of reading as scattered dots.
	// Kept DARK so the pale standing stones on top read with strong contrast (the circle must show).
	mound := darken(blend(blend(rc.dark, stoneDark, 0.5), earthAnchor, 0.20), 0.24) // a low dark turf/earth platform
	forEllipse(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, mound) })
	inner := blend(mound, stone, 0.24) // a paved henge ditch/bank one band in, still darker than a stone
	forEllipse(cx, cy, maxInt(hw*3/4, 1), maxInt(hh*3/4, 1), func(x, y int) { img.SetRGBA(x, y, inner) })

	// Standing-stone size scales with the footprint but is floored so a stone always reads as a slab
	// (not a single dot) even on a modest monument.
	sw := maxInt((hw+hh)/6, 1) // stone half-width (drawBlock size)
	// Perimeter ring of 8 upright slabs, at a radius just inside the footprint so they sit on the
	// monument's rim (not spilling past it). Deterministic angles; every stone bounds-checked.
	const uprights = 8
	rx := float64(hw) * 0.80
	ry := float64(hh) * 0.80
	for i := 0; i < uprights; i++ {
		ang := 2 * math.Pi * float64(i) / float64(uprights)
		sx := cx + int(math.Round(math.Cos(ang)*rx))
		sy := cy + int(math.Round(math.Sin(ang)*ry))
		drawUpright(img, sx, sy, sw, stone, stoneLit, stoneDark)
	}
	// Two TRILITHON pairs (lintel-topped uprights) on the east + west flanks: two close uprights
	// bridged by a horizontal cap stone, so the monument reads as Stonehenge, not just a dot ring.
	tw := maxInt(hw/2, 1)
	gap := maxInt(sw+1, 2)
	for _, sign := range []int{-1, 1} {
		bx := cx + sign*tw
		by := cy
		// Two uprights straddling the flank point.
		drawUpright(img, bx, by-gap, sw, stone, stoneLit, stoneDark)
		drawUpright(img, bx, by+gap, sw, stone, stoneLit, stoneDark)
		// The cap stone bridging them (a bold vertical bar of lintel tone spanning the pair).
		for dy := -gap; dy <= gap; dy++ {
			for dx := -1; dx <= 1; dx++ {
				setPixel(img, bx+dx, by+dy, lintel)
			}
		}
		setPixel(img, bx, by, brighten(lintel, 0.12))
	}
	// Central ALTAR / hearth stone: a bold flat slab with a darker core, anchoring the ring's heart.
	ahw := maxInt(hw/3, 1)
	ahh := maxInt(hh/4, 1)
	fillRectC(img, cx, cy, ahw, ahh, blend(stone, stoneDark, 0.30))
	fillRectC(img, cx, cy, maxInt(ahw/2, 0), maxInt(ahh/2, 0), darken(stoneDark, 0.10))
}

// drawUpright paints one megalith standing stone from above: a small grey block, lit on the NW
// crown and shadowed on the SE base, with a soft ground shadow so it reads as an upright slab, not a
// flat dab. Bounds-checked via setPixel/drawBlock. size is the block half-extent (>=1).
func drawUpright(img *image.RGBA, cx, cy, size int, stone, lit, dark color.RGBA) {
	// Soft ground shadow one row south so the slab reads as standing, not painted flat.
	drawBlock(img, cx+1, cy+size+1, maxInt(size-1, 0), dark)
	drawBlock(img, cx, cy, size, stone)
	// Lit NW crown + shaded SE base edge for a hint of upright volume.
	for d := -size; d <= size; d++ {
		setPixel(img, cx+d, cy-size, lit)  // lit top edge
		setPixel(img, cx+d, cy+size, dark) // shadowed base edge
	}
	setPixel(img, cx-size, cy-size, brighten(lit, 0.10))
}

// drawBastion paints one ANGULAR star-fort bastion salient from above (renaissance wallStarFort): an
// arrowhead pointing OUTWARD from the town core (ccx,ccy) — the opposite of a round tower drum. It is a
// filled DIAMOND (a rotated square: |dx|+|dy| <= rad) whose OUTER vertex along the outward radial is
// stretched to a sharp tip, so it reads as a triangular pointed bastion. A lit cap dab marks the tip.
// Every write clamps via setPixel, so it is panic-safe at any position/size (including off-canvas).
func drawBastion(img *image.RGBA, cx, cy, ccx, ccy, rad int, fill, cap color.RGBA) {
	if rad < 1 {
		rad = 1
	}
	// Outward radial unit (from core toward the bastion). Degenerate (bastion at the core) → point east.
	ux, uy := float64(cx-ccx), float64(cy-ccy)
	n := math.Hypot(ux, uy)
	if n < 1e-6 {
		ux, uy, n = 1, 0, 1
	}
	ux, uy = ux/n, uy/n
	tip := int(math.Round(float64(rad) * 1.6)) // how far the pointed tip juts past the diamond body
	// Filled diamond body: a rotated square (angular, not round) centered on the bastion.
	for dy := -rad; dy <= rad; dy++ {
		for dx := -rad; dx <= rad; dx++ {
			if absInt(dx)+absInt(dy) <= rad {
				setPixel(img, cx+dx, cy+dy, fill)
			}
		}
	}
	// The SALIENT TIP: a short filled wedge from the body out along the outward radial, tapering to a
	// point, so the bastion reads as an arrowhead jutting from the rampart (not a symmetric dot).
	for r := 0; r <= tip; r++ {
		f := 1.0 - float64(r)/float64(tip+1) // taper the wedge half-width toward the tip
		hwid := int(math.Round(float64(rad) * f))
		bx := cx + int(math.Round(ux*float64(r)))
		by := cy + int(math.Round(uy*float64(r)))
		// Lay the wedge cross-section perpendicular to the radial.
		px, py := -uy, ux // perpendicular unit
		for w := -hwid; w <= hwid; w++ {
			setPixel(img, bx+int(math.Round(px*float64(w))), by+int(math.Round(py*float64(w))), fill)
		}
	}
	// A lit cap dab right at the tip so the salient point catches light.
	setPixel(img, cx+int(math.Round(ux*float64(tip))), cy+int(math.Round(uy*float64(tip))), cap)
}

// drawRoofTempleWonder: the CLASSICAL wonder (Phase 1b-ii) — a Greco-Roman TEMPLE read from above: a
// pale STONE platform (stylobate), a COLONNADE of vertical pale columns ringing the perimeter, and a
// warm TERRACOTTA gabled roof over the inner cella crowned by a peaked RIDGE (the pediment read).
// Modeled on drawRoofCathedral's structure (a filled body + a crowning line) but produces
// colonnade + pediment, not cruciform + spire, so it reads unmistakably as the Parthenon. Stone tones
// blend marbleAnchor + the roof clay from clayAnchor against the passed rc so it retints with theme.
// Bounds-checked (setPixel/forRect/drawHSpan), panic-safe at any footprint.
func drawRoofTempleWonder(img *image.RGBA, cx, cy, hw, hh int, rc roofColors) {
	// Pale stone tones for the platform + columns; a warm terracotta for the pedimented roof.
	stone := blend(rc.base, marbleAnchor, 0.55)
	stoneLit := brighten(stone, 0.14)
	stoneDark := blend(rc.dark, marbleAnchor, 0.35)
	terra := blend(rc.base, clayAnchor, 0.60)
	terraRidge := brighten(terra, 0.14)

	// STYLOBATE: the full stone platform (a stepped base), a lighter inner deck so the steps read.
	forRect(cx, cy, hw, hh, func(x, y int) { img.SetRGBA(x, y, stoneDark) })
	forRect(cx, cy, maxInt(hw-1, 0), maxInt(hh-1, 0), func(x, y int) { img.SetRGBA(x, y, stone) })

	// CELLA + TERRACOTTA ROOF: the inner temple body along the wider axis under a warm tiled roof,
	// with a peaked ridge line (the pediment). The roof covers the inner ~55% so the colonnade shows
	// around it.
	horizontal := hw >= hh
	rhw, rhh := maxInt(hw*11/20, 1), maxInt(hh*11/20, 1)
	forRect(cx, cy, rhw, rhh, func(x, y int) {
		// Gable shading: lit north/west of the ridge, shaded south/east, so the pediment reads pitched.
		lit := y <= cy
		if !horizontal {
			lit = x <= cx
		}
		if lit {
			img.SetRGBA(x, y, terra)
		} else {
			img.SetRGBA(x, y, darken(terra, 0.20))
		}
	})
	// Pediment ridge along the temple's long axis.
	if horizontal {
		drawHSpan(img, cx-rhw, cx+rhw, cy, terraRidge)
	} else {
		for y := cy - rhh; y <= cy+rhh; y++ {
			setPixel(img, cx, y, terraRidge)
		}
	}

	// COLONNADE: a ring of vertical pale column dabs around the perimeter of the platform, standing
	// proud of the roof so the temple reads as columned. Deterministic spacing; bounds-checked.
	colStep := maxInt((hw+hh)/6, 2)
	// Top + bottom rows of columns (the long colonnades).
	for x := cx - hw + 1; x <= cx+hw-1; x += colStep {
		drawTempleColumn(img, x, cy-hh+1, stoneLit, stoneDark)
		drawTempleColumn(img, x, cy+hh-1, stoneLit, stoneDark)
	}
	// Left + right rows (the end colonnades).
	for y := cy - hh + 1; y <= cy+hh-1; y += colStep {
		drawTempleColumn(img, cx-hw+1, y, stoneLit, stoneDark)
		drawTempleColumn(img, cx+hw-1, y, stoneLit, stoneDark)
	}
}

// drawTempleColumn paints one temple column from above: a small lit stone dab over a darker base
// shadow, so the colonnade reads as upright columns. Bounds-checked via setPixel.
func drawTempleColumn(img *image.RGBA, cx, cy int, lit, dark color.RGBA) {
	setPixel(img, cx, cy+1, dark) // base shadow
	setPixel(img, cx, cy, lit)    // lit column shaft
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
	// Cosmic-scale ages render a zoomed-out scene, not a city — no city-center marker, no landmark
	// roofs, and not even the corner title. The overlay stays empty so nothing is stamped over the
	// planet/starfield.
	if _, ok := cosmicSceneFor(state.Age); ok {
		return plan
	}
	occupied := map[int]bool{}
	plan.addBuildingLabels(geo, cols, rows, occupied)
	plan.addCityCenterLabel(geo, cols, rows, occupied)
	plan.addTitle(state, cols, rows)
	return plan
}
