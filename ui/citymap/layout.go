package citymap

import (
	"image/color"
	"sort"

	"github.com/espresso20/ageforge/config"
)

// layout.go is the P2 structure layer: per-age placement strategies plus road
// generation. Each strategy turns the player's actual built buildings — grouped
// into lineage districts — into 2.5D building placements and a set of road
// segments, deterministically seeded so the city is stable frame to frame. The
// palace stays central in every strategy; districts arrange around it per era.
//
// Eras group the 22 ages into the seven silhouettes from the map-overhaul design
// doc. The strategy is chosen by era; the terrain + hue still shift per age
// (terrain.go), so two ages in the same era share a silhouette but not a look.

// era is one of the seven layout silhouettes.
type era int

const (
	eraOrganic    era = iota // primitive, stone — clustered scatter + footpaths
	eraHubSpoke              // bronze, iron, classical — roads radiating from palace
	eraCastle                // medieval, renaissance — keep + wall ring + 4 quarters
	eraZonedGrid             // colonial, industrial, victorian — production/residential split
	eraCityBlocks            // electric, modern, atomic — regular blocks + avenues
	eraCampus                // information, digital, cyberpunk, fusion — cluster pods
	eraOrbital               // space, interstellar, galactic, quantum, transcendent — rings
)

// eraForAge maps an age key to its era band via the canonical age order, so the
// grouping is driven by age position and any future age slots in sensibly. The
// band cut points follow the design doc / P2 spec:
//
//	0-1   organic      (primitive, stone)
//	2-4   hub-and-spoke (bronze, iron, classical)
//	5-6   castle        (medieval, renaissance)
//	7-9   zoned grid    (colonial, industrial, victorian)
//	10-12 city blocks   (electric, atomic, modern)  [orders 10,11,12]
//	13-16 campus        (information, digital, cyberpunk, fusion)
//	17+   orbital       (space, interstellar, galactic, quantum, transcendent)
//
// An unknown age (test stub) defaults to organic — the gentlest strategy.
func eraForAge(ageKey string) era {
	order := config.AgeOrder()
	idx := -1
	for i, k := range order {
		if k == ageKey {
			idx = i
			break
		}
	}
	if idx < 0 {
		return eraOrganic
	}
	switch {
	case idx <= 1:
		return eraOrganic
	case idx <= 4:
		return eraHubSpoke
	case idx <= 6:
		return eraCastle
	case idx <= 9:
		return eraZonedGrid
	case idx <= 12:
		return eraCityBlocks
	case idx <= 16:
		return eraCampus
	default:
		return eraOrbital
	}
}

// importance tiers a placement for sizing and z-order.
type importance int

const (
	impNormal importance = iota
	impWonder
	impPalace
)

// placement is a single 2.5D building volume to draw: a pixel center, a footprint
// half-size (0 → 1px, 1 → 3×3, …), the lineage/category color, and its tier. It
// also carries the building's identity (key/name/lineage/category/tier) so the
// overlay pass can name the marker with the building's own config.BuildingDef.Name
// — one named marker per built building type — rather than a lineage banner. The
// palace placement leaves the identity fields zero (it is labeled "City Center").
type placement struct {
	cx, cy int
	size   int
	col    color.RGBA
	tier   importance

	// Building identity (empty for the palace). The label text is name; the label
	// color is derived from lineageKey/category via lineageColor at draw time so it
	// retints with the theme. ltier (LineageTier) + size prioritize labels when space
	// is tight (higher tier / bigger volume gets its name first per cluster).
	key        string
	name       string
	lineageKey string
	category   string
	ltier      int
}

// roadSeg is a straight road drawn between two pixel endpoints (Bresenham).
type roadSeg struct {
	x0, y0, x1, y1 int
}

// buildingItem is one distinct BUILT building type (Count > 0) the map draws as
// its own marker. It carries everything a placement needs to render + be named:
// the config key, the human Name, the lineage/category (for color + clustering),
// the instance count (nudges volume size), and the LineageTier (label priority).
type buildingItem struct {
	key      string
	name     string
	category string
	count    int
	tier     int // LineageTier — higher = more advanced; used to rank labels
}

// district is one lineage's (or special category's) buildings, grouped together so
// they cluster + share a color on the map. Each member of buildings draws as its
// OWN marker (one per distinct built type) within the district's region; the
// district itself contributes the shared lineage color (col) and the placement
// region the era strategy assigns it. lineageKey/category identify the group; col
// is the lineage color every member volume uses.
type district struct {
	lineageKey string
	category   string
	col        color.RGBA
	buildings  []buildingItem
}

// buildingMeta resolves a building key to its lineage key and category via the
// canonical config table. Pure data, no locks (config.BuildingByKey is a plain
// map build), so it is safe to call from the render path. Cached per call site by
// the caller; here it is a thin lookup. Unknown keys return ("","",false).
func buildingMeta(byKey map[string]config.BuildingDef, key string) (lineageKey, category string, ok bool) {
	d, ok := byKey[key]
	if !ok {
		return "", "", false
	}
	return d.LineageKey, d.Category, true
}

// builtDistricts groups the player's built buildings by lineage into districts,
// sorted for determinism. Crucially this no longer collapses a lineage to a few
// representatives: EVERY distinct built building type (Count > 0) becomes its own
// buildingItem inside its lineage's district, so the map draws one named marker per
// type. Same-lineage buildings land in the same district (so they cluster + share
// a color); storage/wonder/diplomacy/monument group by category so they read as a
// distinct neighborhood. keys must be sorted; counts maps key→instance count;
// byKey supplies each building's Name + LineageTier (pure data, no locks).
func builtDistricts(byKey map[string]config.BuildingDef, keys []string, counts map[string]int) []district {
	// Aggregate by a grouping key: production lineages group by lineageKey; the
	// special categories group by category so e.g. all wonders cluster together.
	groups := map[string]*district{}
	order := make([]string, 0) // preserve first-seen order before the stable sort
	for _, k := range keys {
		lineage, category, _ := buildingMeta(byKey, k)
		gkey := lineage
		switch category {
		case "wonder", "monument", "storage", "diplomacy":
			gkey = "cat:" + category
		}
		if gkey == "" {
			gkey = "misc"
		}
		d := groups[gkey]
		if d == nil {
			d = &district{
				lineageKey: lineage,
				category:   category,
				col:        lineageColor(lineage, category),
			}
			groups[gkey] = d
			order = append(order, gkey)
		}
		def := byKey[k]
		name := def.Name
		if name == "" {
			name = titleCaseKey(k) // unknown/test key: a sensible label rather than blank
		}
		d.buildings = append(d.buildings, buildingItem{
			key:      k,
			name:     name,
			category: category,
			count:    counts[k],
			tier:     def.LineageTier,
		})
	}
	sort.Strings(order)

	out := make([]district, 0, len(order))
	for _, g := range order {
		out = append(out, *groups[g])
	}
	return out
}

// layoutSeed derives a deterministic seed from the age plus the set of built
// building keys, so the layout is stable while the same buildings are present but
// reshuffles sensibly when the empire changes. FNV-1a over age + sorted keys.
func layoutSeed(ageKey string, keys []string) uint32 {
	var h uint32 = 2166136261
	mix := func(s string) {
		for i := 0; i < len(s); i++ {
			h ^= uint32(s[i])
			h *= 16777619
		}
		h ^= '|'
		h *= 16777619
	}
	mix(ageKey)
	for _, k := range keys {
		mix(k)
	}
	return h | 1
}

// rng is a tiny deterministic PRNG (xorshift32) for jittering placements. Pure
// and seeded, so a given (age, building set) always produces the same city.
type rng struct{ s uint32 }

func newRNG(seed uint32) *rng {
	if seed == 0 {
		seed = 1
	}
	return &rng{s: seed}
}

func (r *rng) next() uint32 {
	r.s ^= r.s << 13
	r.s ^= r.s >> 17
	r.s ^= r.s << 5
	return r.s
}

// f01 returns a float in [0,1).
func (r *rng) f01() float64 { return float64(r.next()) / float64(^uint32(0)) }

// span returns an int in [-half, half].
func (r *rng) span(half int) int {
	if half <= 0 {
		return 0
	}
	return int(r.next()%uint32(2*half+1)) - half
}

// The per-era placement STRATEGIES that used to live here (organic scatter, hub-and-
// spoke, castle, zoned grid, city blocks, campus, orbital) were superseded by the
// count-driven city synthesizer in citygen.go (citymap v2). The draw path now builds
// a whole cityPlan per era via generateCityPlan instead of scattering one marker per
// built type and stringing MST roads between them. The era enum + eraForAge remain
// here as the era-band keying the synthesizer's eraStyle presets look up; the
// district grouping (builtDistricts) and the small helpers below are still used by
// the synthesizer's gather step and by the retained volume/geometry primitives.

// minInt returns the smaller of two ints.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
