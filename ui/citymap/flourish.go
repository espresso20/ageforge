package citymap

import (
	"image"
	"image/color"
	"math"
)

// flourish.go is the P3 per-age polish: small, era-keyed ambient touches that make
// ages feel distinct beyond their layout silhouette, painted over the terrain but
// under the structure so they stay background. Coherence first — every flourish is
// deliberately faint (low blend amounts, sparse coverage) and pulled from the active
// theme palette so it retints. Only a few eras get a treatment; the rest stay clean.
//
//   - orbital  (space → transcendent): a sparse starfield speckle in the margins —
//     the late-game void around the orbital city.
//   - campus   (information → fusion):  a faint neon edge-glow — the cyber districts
//     bleeding light to the frame.
//   - zonedGrid / cityBlocks (industrial → modern): a few soft smoke wisps drifting
//     up from the production sprawl.
//
// All coordinates are clipped to the image; the field is seed-driven so a given age
// always produces the same flourish.

// drawEraFlourish dispatches the era's ambient treatment, if any. seed reuses the
// age's terrain seed so the flourish is stable and co-varies with the terrain.
func drawEraFlourish(img *image.RGBA, pal terrainPalette, e era, seed uint32) {
	b := img.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		return
	}
	switch e {
	case eraOrbital:
		drawStarfield(img, pal, seed)
	case eraCampus:
		drawNeonEdge(img, pal)
	case eraZonedGrid, eraCityBlocks:
		drawSmokeWisps(img, pal, seed)
	}
}

// drawStarfield speckles a sparse scatter of bright points across the whole field,
// densest near the margins (where the orbital layout leaves empty void) and thinning
// toward the center where the rings sit. Points are the peak/accent tone brightened,
// so they read as stars and retint with the theme. Deterministic from seed.
func drawStarfield(img *image.RGBA, pal terrainPalette, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	star := brighten(pal.peak, 0.55)
	cx, cy := float64(w)/2, float64(h)/2
	maxD := math.Hypot(cx, cy)
	if maxD < 1 {
		maxD = 1
	}
	// Budget roughly one candidate per 60px, capped, so the speckle is sparse.
	budget := clampInt((w*h)/60, 0, 600)
	for i := 0; i < budget; i++ {
		// Two hashes per point: position, then a keep-roll weighted by edge distance.
		hx := hash2(uint32(i), 0x5a5a, seed)
		hy := hash2(uint32(i), 0xa5a5, seed)
		x := int(hx % uint32(w))
		y := int(hy % uint32(h))
		// Keep more points the farther from center (edge-weighted).
		d := math.Hypot(float64(x)-cx, float64(y)-cy) / maxD // 0 center .. 1 corner
		keep := 0.15 + 0.85*d                                // center sparse, edges denser
		if hashUnit(uint32(i), 0x1234, seed) > keep {
			continue
		}
		if x >= b.Min.X && x < b.Max.X && y >= b.Min.Y && y < b.Max.Y {
			img.SetRGBA(x, y, star)
		}
	}
}

// drawNeonEdge blends a thin highlight-tinted glow into the outermost few pixel
// bands of the frame, strongest at the very edge and fading inward, so the campus/
// cyber eras read as light-bleeding. Uses the building/highlight tone blended over
// the existing terrain pixel (so it glows without erasing the land underneath).
func drawNeonEdge(img *image.RGBA, pal terrainPalette) {
	b := img.Bounds()
	const bands = 3
	neon := pal.building // RoleHighlight-derived
	for band := 0; band < bands; band++ {
		// Outer band glows most; t is the blend strength for this ring.
		t := 0.30 * (1 - float64(band)/float64(bands))
		x0, y0 := b.Min.X+band, b.Min.Y+band
		x1, y1 := b.Max.X-1-band, b.Max.Y-1-band
		if x0 > x1 || y0 > y1 {
			break
		}
		// Top and bottom edges of this ring.
		for x := x0; x <= x1; x++ {
			blendPixel(img, x, y0, neon, t)
			blendPixel(img, x, y1, neon, t)
		}
		// Left and right edges of this ring.
		for y := y0; y <= y1; y++ {
			blendPixel(img, x0, y, neon, t)
			blendPixel(img, x1, y, neon, t)
		}
	}
}

// drawSmokeWisps drifts a handful of soft vertical smudges up from the lower-middle
// of the canvas (the production sprawl) toward the top, each a short column of the
// shadow/dim tone blended faintly over the terrain so it reads as haze, not a wall.
// Count and placement are seed-driven for stability.
func drawSmokeWisps(img *image.RGBA, pal terrainPalette, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 8 || h < 8 {
		return
	}
	smoke := blend(pal.shadow, pal.highland, 0.25)
	// A few wisps spread across the width.
	wisps := clampInt(w/24, 2, 6)
	for i := 0; i < wisps; i++ {
		// Base x spread across the canvas with a seeded jitter.
		baseX := (i+1)*w/(wisps+1) + int(hash2(uint32(i), 0x77, seed)%uint32(maxInt(w/12, 1))) - w/24
		// Rise from ~70% height up to ~20% height, wandering slightly as it climbs.
		yStart := int(float64(h) * 0.70)
		yEnd := int(float64(h) * 0.20)
		drift := 0.0
		for y := yStart; y >= yEnd; y-- {
			// Wander: accumulate a small seeded horizontal drift per row.
			drift += (hashUnit(uint32(i*1000+y), 0xbeef, seed) - 0.5) * 0.6
			x := baseX + int(math.Round(drift))
			// Fade with height (thinner/fainter near the top).
			prog := float64(yStart-y) / math.Max(1, float64(yStart-yEnd)) // 0 bottom .. 1 top
			t := 0.22 * (1 - prog)
			blendPixel(img, x, y, smoke, t)
			// A touch of width low down for body.
			if prog < 0.4 {
				blendPixel(img, x-1, y, smoke, t*0.6)
				blendPixel(img, x+1, y, smoke, t*0.6)
			}
		}
	}
}

// blendPixel mixes color c into the existing pixel at (x,y) by t in [0,1], clipped
// to the image. Used by the flourishes so they tint the terrain rather than erase
// it (a flat SetRGBA would punch a hole through the land).
func blendPixel(img *image.RGBA, x, y int, c color.RGBA, t float64) {
	b := img.Bounds()
	if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
		return
	}
	img.SetRGBA(x, y, blend(img.RGBAAt(x, y), c, t))
}
