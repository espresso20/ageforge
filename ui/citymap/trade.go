package citymap

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/espresso20/ageforge/game"
)

// trade.go is the pixel half of the P3 trade-route weave: it draws each ACTIVE
// trade route as a dashed connector line running from the city center out to a
// point on the map border, distinct from a road (roads are solid; lanes dash). The
// matching border LABELS are text and live in the overlay pass (overlay.go); this
// file returns the border endpoints so the overlay can tag them.
//
// Geometry plumbing also lives here: layoutGeometry is the small bundle of
// cell/pixel anchors (district centroids, palace, trade ends) that the structure
// and trade passes hand to the overlay so every label lands exactly where the
// pixels did.

// districtCentroid is a labeled district's anchor in PIXEL space, plus the data the
// overlay needs to name it: its lineage/category (the label text + color are derived
// from these at draw time via lineageLabel/lineageColor, so the label retints with
// the theme) and raw building count (used to prioritize labels when space is tight).
type districtCentroid struct {
	px, py     int
	lineageKey string
	category   string
	count      int
}

// tradeEnd is one active trade lane's border endpoint in PIXEL space, with the
// route's display name for the overlay tag.
type tradeEnd struct {
	px, py int
	name   string
}

// layoutGeometry is the per-frame bundle of anchors shared between the pixel passes
// and the text overlay. Pixels are in image space (cols × rows*2); the overlay
// halves the y to land on cells. palaceX/Y is the palace center.
type layoutGeometry struct {
	palaceX, palaceY int
	districts        []districtCentroid
	tradeEnds        []tradeEnd
}

// maxTradeLanes caps how many active routes we draw so the lane fan stays legible —
// beyond this the border turns to spaghetti. The most-relevant routes (by cycles
// completed, then name) are kept; the rest are summarized by the Trade overlay, not
// the map.
const maxTradeLanes = 8

// drawTradeLanes draws up to maxTradeLanes dashed connector lines from the city
// center to evenly-spaced points on the map border, one per active (non-disrupted-
// preferred) route, and returns their border endpoints for the overlay to label.
// Lanes are dashed so they read as trade routes, not roads (which are solid). The
// color is a positive/highlight tint pulled from the palette so it retints with the
// theme. Disrupted routes are still drawn (the player should see a stalled lane) but
// dimmed toward the road tone.
func drawTradeLanes(img *image.RGBA, pal terrainPalette, state game.GameState, cx, cy int) []tradeEnd {
	routes := selectTradeRoutes(state, maxTradeLanes)
	if len(routes) == 0 {
		return nil
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	ends := make([]tradeEnd, 0, len(routes))
	n := len(routes)
	for i, rt := range routes {
		// Fan the lanes around the compass, biased to the sides/top so they don't all
		// pile onto the bottom where the title/legend tends to sit. Start at -120° and
		// sweep ~300° so endpoints spread across top, left, and right borders.
		frac := 0.0
		if n > 1 {
			frac = float64(i) / float64(n)
		}
		ang := -2.0/3.0*math.Pi + frac*(5.0/3.0*math.Pi)
		ex, ey := borderPoint(cx, cy, ang, w, h)

		lc := pal.tradeLane
		if rt.disrupted {
			// A stalled lane: pull the lane color toward the (quieter) road tone.
			lc = blend(pal.tradeLane, pal.road, 0.6)
		}
		drawDashedLine(img, cx, cy, ex, ey, lc)
		ends = append(ends, tradeEnd{px: ex, py: ey, name: rt.name})
	}
	return ends
}

// tradeRoutePick is the minimal route info the map needs.
type tradeRoutePick struct {
	name      string
	disrupted bool
	cycles    int
}

// selectTradeRoutes picks up to max active routes, preferring the most-established
// (more completed cycles) and breaking ties by name for stability, so the same
// routes draw frame-to-frame rather than reshuffling.
func selectTradeRoutes(state game.GameState, max int) []tradeRoutePick {
	src := state.Trade.ActiveRoutes
	if len(src) == 0 || max <= 0 {
		return nil
	}
	picks := make([]tradeRoutePick, 0, len(src))
	for _, r := range src {
		name := r.Name
		if name == "" {
			name = r.Key
		}
		picks = append(picks, tradeRoutePick{name: name, disrupted: r.Disrupted, cycles: r.CyclesDone})
	}
	sort.SliceStable(picks, func(i, j int) bool {
		if picks[i].cycles != picks[j].cycles {
			return picks[i].cycles > picks[j].cycles
		}
		return picks[i].name < picks[j].name
	})
	if len(picks) > max {
		picks = picks[:max]
	}
	return picks
}

// borderPoint casts a ray from (cx,cy) at angle ang and returns the pixel where it
// meets the image border, clamped just inside the edge so the endpoint (and any
// label anchored to it) stays on-canvas. Pure trig + clamp; never out of bounds.
func borderPoint(cx, cy int, ang float64, w, h int) (int, int) {
	if w <= 0 || h <= 0 {
		return clampInt(cx, 0, maxInt(w-1, 0)), clampInt(cy, 0, maxInt(h-1, 0))
	}
	dx := math.Cos(ang)
	dy := math.Sin(ang)
	// Distance to each candidate edge along the ray; take the nearest positive hit.
	best := math.MaxFloat64
	if dx > 1e-9 {
		if t := (float64(w-1) - float64(cx)) / dx; t > 0 && t < best {
			best = t
		}
	} else if dx < -1e-9 {
		if t := (0 - float64(cx)) / dx; t > 0 && t < best {
			best = t
		}
	}
	if dy > 1e-9 {
		if t := (float64(h-1) - float64(cy)) / dy; t > 0 && t < best {
			best = t
		}
	} else if dy < -1e-9 {
		if t := (0 - float64(cy)) / dy; t > 0 && t < best {
			best = t
		}
	}
	if best == math.MaxFloat64 {
		best = 0
	}
	ex := roundPx(float64(cx) + dx*best)
	ey := roundPx(float64(cy) + dy*best)
	return clampInt(ex, 0, w-1), clampInt(ey, 0, h-1)
}

// drawDashedLine rasterizes a dashed Bresenham line between two pixel points: it
// paints a run of pixels, skips a gap, and repeats, so the lane reads as a dotted
// trade route distinct from the solid roads. Clipped to the image bounds.
func drawDashedLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	b := img.Bounds()
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
	err := dx + dy
	// Dash pattern: paint dashOn pixels, skip dashOff. Tuned to read as a lane at
	// the coarse half-block resolution without looking like noise.
	const dashOn, dashOff = 2, 2
	step := 0
	for {
		if step%(dashOn+dashOff) < dashOn {
			if x0 >= b.Min.X && x0 < b.Max.X && y0 >= b.Min.Y && y0 < b.Max.Y {
				img.SetRGBA(x0, y0, c)
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
		step++
	}
}

// maxInt returns the larger of two ints (companion to minInt in layout.go).
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
