package citymap

import (
	"github.com/espresso20/ageforge/config"
)

// layout.go holds the era-band keying and the small shared generator utilities the
// citymap render path still uses. It once owned the P2 isometric structure layer
// (per-age placement strategies + MST road generation → 2.5D volumes), but that path
// was retired by the citymap v3 top-down rewrite
// (design-and-architecture/city-synthesis.md). topdown.go now grows the whole city as
// count-scaled top-down roofs and keys its era presets off eraForAge; the placement
// strategies, the district/placement structs, and the layout seed they used are gone.
//
// What remains:
//   - era + eraForAge: the seven era bands topdown.go's tdStyleForEra looks up (and the
//     worldmap-adjacent tests exercise);
//   - buildingMeta: the pure key→lineage/category lookup gatherBuildings (topdown.go's
//     gather step) reads;
//   - roadSeg: the segment type the shared drawRoad rasterizer + pathfind speak;
//   - rng: the tiny deterministic PRNG topdown.go seeds all its placement/jitter from;
//   - minInt: a shared int helper used by the top-down atlas.

// era is one of the seven era bands. Each groups a run of the 22 ages into a shared
// visual band; topdown.go's tdStyleForEra maps the band to a top-down era preset
// (V3-A tunes the organic band; the rest use a legible default until V3-B+).
type era int

const (
	eraOrganic    era = iota // primitive, stone — earthy village, winding dirt lanes
	eraHubSpoke              // bronze, iron, classical — ancient (radial-grid, clay)
	eraCastle                // medieval, renaissance — stone walls + towers
	eraZonedGrid             // colonial, industrial, victorian — grid, no walls
	eraCityBlocks            // electric, atomic, modern — avenues + superblocks
	eraCampus                // information, digital, cyberpunk, fusion — neon megablocks
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
// An unknown age (test stub) defaults to organic — the gentlest band.
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

// roadSeg is a straight road drawn between two pixel endpoints (Bresenham). Shared
// by the top-down street rasterizer (topdown.go) and the pathfinder.
type roadSeg struct {
	x0, y0, x1, y1 int
}

// buildingMeta resolves a building key to its lineage key and category via the
// canonical config table. Pure data, no locks (config.BuildingByKey is a plain
// map build), so it is safe to call from the render path. Unknown keys return
// ("","",false). Read by gatherBuildings (topdown.go's gather step).
func buildingMeta(byKey map[string]config.BuildingDef, key string) (lineageKey, category string, ok bool) {
	d, ok := byKey[key]
	if !ok {
		return "", "", false
	}
	return d.LineageKey, d.Category, true
}

// rng is a tiny deterministic PRNG (xorshift32) for jittering placements. Pure
// and seeded, so a given seed always produces the same city. topdown.go threads one
// of these through its whole plan generation so the layout is stable frame-to-frame.
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

// minInt returns the smaller of two ints.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
