package citymap

import (
	"image"
	"image/color"
	"math"

	"github.com/espresso20/ageforge/theme"
)

// worldmedium.go is Phase B of the world map: the CARTOGRAPHIC MEDIUM. Phase A built
// ONE seeded world (worldmodel.go) and rendered it in a single neutral atlas look for
// every age (drawWorldModel). This file introduces the abstraction that lets the SAME
// continent be re-drawn in each era's map STYLE — charcoal cave-drawing, aged parchment,
// orbital satellite photo, holographic neon — the terminal-native take on the
// ageforge-online `medium.ts` idea.
//
// The shape:
//   - worldMedium bundles a mediumPalette (all the tones a world render needs) with a
//     draw function. The draw function receives the image, the world model, and the
//     resolved palette, and paints the terrain layer however that medium wants. It does
//     NOT draw the civ dots / labels — those stay in renderWorldImage's overlay pass so
//     every medium shares one dot/label pipeline (a medium re-skins the LAND, not the
//     player's markers).
//   - mediumForAge(ageKey) picks the medium. Four ages get a bespoke, dramatically
//     different medium (primitive→charcoal, medieval→parchment, modern→satellite,
//     cyberpunk→neon); EVERY OTHER age gets the default neutral atlas — behavior-
//     preserving, the exact look Phase A shipped.
//
// Determinism + safety: a medium is a pure function of (model, palette). The palettes are
// built from the active theme where it makes sense (so a theme switch still nudges tones)
// but each medium anchors HARD to its own identity — a parchment is cream-and-ink
// regardless of theme, a neon projection is black-and-cyan regardless of theme — because
// the whole point of Phase B is that the mediums read as DIFFERENT MAPS, not tinted
// variants of one. Every write clips to the canvas (setPixel / bounds-checked helpers),
// so tiny / odd / zero canvases never panic; an empty model paints its background and
// returns.

// mediumPalette is the full tone set a world render consumes. Not every medium uses every
// field (a satellite photo has no drawn coastline; a neon grid has no biome anchors), but
// carrying them all in one struct keeps mediumForAge's construction uniform and lets the
// default atlas keep its exact Phase-A tones. Biome-indexed land tones live in `land`
// (indexed by biome) so a medium can look up the fill for any classified pixel.
type mediumPalette struct {
	background color.RGBA             // whole-frame base (behind everything a medium draws)
	oceanDeep  color.RGBA             // deepest sea
	oceanShelf color.RGBA             // shallow shelf near the coast
	coast      color.RGBA             // the drawn shoreline stroke (mediums without one ignore it)
	river      color.RGBA             // watercourses
	relief     color.RGBA             // relief symbols (mountains/hills), lifted per medium
	reliefAlt  color.RGBA             // secondary relief tone (snow caps / forest / accent glyphs)
	shadow     color.RGBA             // grounding shadow under relief marks
	land       [biomeCount]color.RGBA // per-biome land fill
}

// worldMediumStyle is the coastline / chrome flavor tag. It is informational (each
// medium's draw function bakes its own technique) but kept on the struct so tests and
// future callers can branch on "does this medium draw a hard outline?" without poking at
// function identity.
type worldMediumStyle uint8

const (
	styleAtlas       worldMediumStyle = iota // neutral default: crisp coast, depth-banded sea
	styleCharcoal                            // near-mono, grainy, jittered chalk coastline
	styleParchment                           // vellum, wavy ink coast, frame + compass + cartouche
	styleSatellite                           // photographic, natural land/sea transition, no outline
	styleNeon                                // black + neon wireframe, holographic grid overlay
	stylePetroglyph                          // light sandstone rock carving: pecked stipple, carved-groove coast
	styleClayTablet                          // wet terracotta tablet: pressed-groove coast, wedge stamps
	styleHide                                // tanned leather with burnt edges: soot + ochre ink
	styleMosaic                              // Roman tesserae: quantized tile grid + grout, Greek-key border
	styleCopperplate                         // renaissance engraving: fine ruled sea-hatch, cross-hatched relief, blank land
	styleNautical                            // colonial sea-chart: rhumb-line net over the water, sounding stipple
	styleOrdnance                            // Victorian ordnance sheet: brown contour isolines + printed grid on pale paper
	styleLithograph                          // hand-tinted lithograph: pastel biome washes over fine black outline
)

// worldMedium is a cartographic medium: how a world is drawn for one era. `palette` is the
// resolved tone set; `style` tags the flavor; `draw` paints the terrain layer for this
// medium (the civ/label overlay is added afterward by renderWorldImage).
type worldMedium struct {
	name    string
	style   worldMediumStyle
	palette mediumPalette
	draw    func(img *image.RGBA, m *worldModel, med worldMedium)
}

// mediumForAge returns the cartographic medium for an age. FOUR ages get a bespoke medium
// that reads dramatically different from the others and from the default; every other age
// falls through to the neutral atlas (the Phase-A look), so nothing that isn't one of the
// four validation ages changes. The palette is built from the active theme inside each
// constructor, so this is resolved fresh per render (a theme switch retints where a medium
// allows it).
func mediumForAge(ageKey string) worldMedium {
	switch ageKey {
	case "primitive_age":
		return charcoalMedium()
	case "stone_age":
		return petroglyphMedium()
	case "bronze_age":
		return clayTabletMedium()
	case "iron_age":
		return hideMedium()
	case "classical_age":
		return mosaicMedium()
	case "renaissance_age":
		return copperplateMedium()
	case "colonial_age":
		return nauticalMedium()
	case "industrial_age":
		return ordnanceMedium()
	case "victorian_age":
		return lithographMedium()
	case "medieval_age":
		return parchmentMedium()
	case "modern_age":
		return satelliteMedium()
	case "cyberpunk_age":
		return neonMedium()
	default:
		return atlasMedium(ageKey)
	}
}

// ---- default: neutral atlas -------------------------------------------------

// atlasMedium is the DEFAULT medium: the exact neutral atlas Phase A shipped. Its palette
// is atlasBiomeTones (built from the age's own palette so it still hue-shifts per age),
// and its draw function is drawWorldModelAtlas — the former body of drawWorldModel, moved
// here verbatim so every non-validation age renders byte-for-byte as before.
func atlasMedium(ageKey string) worldMedium {
	_, hueShift := ageInfo(ageKey)
	pal := buildPalette(hueShift)
	tones := atlasBiomeTones(pal)

	var mp mediumPalette
	mp.background = rgba(theme.Color(theme.RoleBackground))
	for bi := biome(0); bi < biomeCount; bi++ {
		mp.land[bi] = tones[bi]
	}
	mp.oceanDeep = tones[biomeDeepWater]
	mp.oceanShelf = tones[biomeShallowWater]
	mp.coast = brighten(tones[biomeSand], 0.10)
	// The atlas river tone is the palette's shallow-water pushed to a bright blue (unchanged
	// from Phase A) — carried here so the shared atlas draw can read it off the palette.
	mp.river = brighten(blend(pal.bShallowWater, color.RGBA{R: 0x55, G: 0x8f, B: 0xd8, A: 0xff}, 0.55), 0.14)
	mp.relief = brighten(tones[biomeMountain], 0.14)
	mp.reliefAlt = brighten(tones[biomeSnow], 0.10)
	mp.shadow = pal.shadow

	return worldMedium{
		name:  "atlas",
		style: styleAtlas,
		draw: func(img *image.RGBA, m *worldModel, med worldMedium) {
			drawWorldModelAtlas(img, pal, m, med.palette)
		},
		palette: mp,
	}
}

// ---- charcoal: primitive cave-drawing ---------------------------------------

// charcoalMedium is a near-monochrome cave sketch (→ primitive_age). Near-black slate sea,
// charcoal-grey land under a rough grainy/hatched texture, a chalky off-white ROUGH
// coastline (jittered, not smooth), mountains as hand-scratched carets, rivers as faint
// chalk lines, and a rough dark stone border. It anchors to its own greys/ochre HARD — a
// cave drawing is a cave drawing regardless of theme — so it reads unmistakably primitive.
func charcoalMedium() worldMedium {
	// Fixed charcoal identity (theme-independent by design — this is the point of a medium).
	slate := color.RGBA{R: 0x14, G: 0x14, B: 0x17, A: 0xff} // near-black slate sea
	slateDeep := color.RGBA{R: 0x0c, G: 0x0c, B: 0x0f, A: 0xff}
	rock := color.RGBA{R: 0x3a, G: 0x37, B: 0x33, A: 0xff} // charcoal-grey land
	rockDark := color.RGBA{R: 0x2b, G: 0x28, B: 0x25, A: 0xff}
	chalk := color.RGBA{R: 0xd8, G: 0xcf, B: 0xba, A: 0xff} // chalky off-white
	ochre := color.RGBA{R: 0xb2, G: 0x8a, B: 0x54, A: 0xff} // ochre pigment (relief)

	var mp mediumPalette
	mp.background = slate
	mp.oceanDeep = slateDeep
	mp.oceanShelf = slate
	mp.coast = chalk
	mp.river = blend(rock, chalk, 0.45) // faint chalk line
	mp.relief = ochre
	mp.reliefAlt = chalk
	mp.shadow = rockDark
	// Land is basically one charcoal grey; the texture (not hue) carries all the read.
	for bi := biome(0); bi < biomeCount; bi++ {
		mp.land[bi] = rock
	}
	mp.land[biomeForest] = blend(rock, rockDark, 0.35) // damp ground scratched darker
	mp.land[biomeMountain] = blend(rock, rockDark, 0.55)
	mp.land[biomeSnow] = blend(rock, rockDark, 0.55)

	return worldMedium{name: "charcoal", style: styleCharcoal, palette: mp, draw: drawCharcoal}
}

// drawCharcoal paints the cave-drawing medium. Technique: flat slate sea (depth barely
// banded — a cave sea is just dark), charcoal land grained with a stable per-pixel hash so
// it reads hand-shaded, a JITTERED chalk coastline (each shore pixel nudged by a hash so
// the line is rough, not a clean vector), rivers as faint dotted chalk, mountains as
// scratched ochre carets, and a rough dark border scratched around the frame.
func drawCharcoal(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	seed := uint32(0xCA0A1)

	// 1) Base: slate sea (a whisper of depth), charcoal land with a grain/hatch texture.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := m.field.at(x, y)
			if bi >= biomeCount {
				bi = biomeGrass
			}
			var col color.RGBA
			if bi == biomeDeepWater || bi == biomeShallowWater {
				// Barely-banded slate: a touch darker in the deep so the sea isn't dead flat.
				e := m.elevAtPx(x, y)
				depth := 0.0
				if m.seaLevel > 0 {
					depth = clamp01((m.seaLevel - e) / m.seaLevel)
				}
				col = blend(p.oceanShelf, p.oceanDeep, depth*0.6)
			} else {
				col = p.land[bi]
				// Grain: a stable per-pixel jitter, darkened by a value-noise field so the land
				// reads as rough charcoal shading rather than a flat grey. A diagonal hatch
				// (every few px on x+y) scratches in a few darker strokes.
				g := hashUnit(uint32(x), uint32(y), seed)
				n := valueNoise(float64(x)*0.18, float64(y)*0.18, seed^0x33)
				shade := 0.10 + 0.22*n + 0.10*g
				col = darken(col, shade)
				if (x+y)%5 == 0 && g > 0.55 {
					col = darken(col, 0.18) // hatch stroke
				}
			}
			setPixel(img, x, y, col)
		}
	}

	// 2) Rough chalk coastline: every shore pixel (land with a water 4-neighbour), but
	//    JITTERED — the mark is nudged by a per-pixel hash so the line wobbles like a finger
	//    drawn in soot, and only ~75% of shore pixels are stamped so it reads sketched/broken.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			if isLandPx(m.field, x-1, y) && isLandPx(m.field, x+1, y) &&
				isLandPx(m.field, x, y-1) && isLandPx(m.field, x, y+1) {
				continue
			}
			hgate := hashUnit(uint32(x), uint32(y), seed^0xC0A5)
			if hgate < 0.25 {
				continue // broken line — skip a quarter of shore pixels
			}
			// Jitter the stamp by up to 1px so the coast is rough.
			jx := x + int(math.Round((hashUnit(uint32(x), uint32(y), seed^0x1A)-0.5)*1.6))
			jy := y + int(math.Round((hashUnit(uint32(y), uint32(x), seed^0x2B)-0.5)*1.6))
			setPixel(img, jx, jy, p.coast)
		}
	}

	// 3) Rivers: faint dotted chalk lines (skip alternating steps for a broken hand-drawn read).
	step := 0
	putRiver := func(x, y int, c color.RGBA) {
		if !isLandPx(m.field, x, y) {
			return
		}
		step++
		if step%2 == 0 {
			return // dotted
		}
		setPixel(img, x, y, c)
	}
	for _, r := range m.rivers {
		for i := 0; i+1 < len(r.pts); i++ {
			a, bb := r.pts[i], r.pts[i+1]
			strokeThickLineFunc(a.x, a.y, bb.x, bb.y, 0, p.river, putRiver)
		}
	}

	// 4) Relief: mountains as scratched ochre carets (an apex + two flanks, no fill), a rough
	//    hand-drawn "^". Hills a shorter tick. Forests skipped — a cave map keeps only the big
	//    landmarks. Gate to land so nothing scratches into the sea.
	putR := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, a := range m.reliefs {
		switch a.kind {
		case reliefMountain:
			putR(a.x, a.y-1, p.relief)   // apex
			putR(a.x-1, a.y, p.relief)   // left flank
			putR(a.x+1, a.y, p.relief)   // right flank
			putR(a.x-2, a.y+1, p.relief) // splayed feet → a scratched caret
			putR(a.x+2, a.y+1, p.relief)
		case reliefHill:
			putR(a.x, a.y, p.relief) // a short tick
			putR(a.x-1, a.y+1, p.relief)
			putR(a.x+1, a.y+1, p.relief)
		}
	}

	// 5) Rough stone border: a scratched dark rim a couple of px thick, broken by a hash so
	//    it reads hand-drawn, not a clean rule.
	drawRoughBorder(img, med.palette.shadow, 2, seed^0xB0DE)
}

// ---- parchment: medieval antique map ----------------------------------------

// parchmentMedium is an aged illuminated map (→ medieval_age). Warm cream vellum for BOTH
// land and sea (the sea a shade darker/bluer tan), wavy dark-ink-brown coastlines, faint
// sepia biome hints on land, drawn mountain peaks + tree symbols, thin ink rivers — plus
// the CHROME that makes it read antique: an ornate frame, a compass rose in a corner, and
// a cartouche box behind the title. Sepia everything. Theme-independent — parchment is
// parchment.
func parchmentMedium() worldMedium {
	vellum := color.RGBA{R: 0xe6, G: 0xd7, B: 0xb0, A: 0xff}    // warm cream land
	vellumSea := color.RGBA{R: 0xcf, G: 0xc2, B: 0xa2, A: 0xff} // sea: a duller, bluer tan
	vellumDeep := color.RGBA{R: 0xbe, G: 0xb2, B: 0x96, A: 0xff}
	ink := color.RGBA{R: 0x5a, G: 0x3d, B: 0x22, A: 0xff}   // dark ink-brown
	sepia := color.RGBA{R: 0x9a, G: 0x7c, B: 0x50, A: 0xff} // faint biome sepia
	forestSepia := color.RGBA{R: 0x6f, G: 0x6a, B: 0x3e, A: 0xff}

	var mp mediumPalette
	mp.background = vellum
	mp.oceanDeep = vellumDeep
	mp.oceanShelf = vellumSea
	mp.coast = ink
	mp.river = blend(ink, vellumSea, 0.30) // thin muted ink
	mp.relief = ink
	mp.reliefAlt = forestSepia
	mp.shadow = blend(ink, vellum, 0.45)
	for bi := biome(0); bi < biomeCount; bi++ {
		mp.land[bi] = vellum
	}
	// Faint sepia hints so land isn't a dead flat cream: forests greenish-sepia, uplands a
	// touch browner, sand a shade warmer.
	mp.land[biomeForest] = blend(vellum, forestSepia, 0.30)
	mp.land[biomeRock] = blend(vellum, sepia, 0.28)
	mp.land[biomeMountain] = blend(vellum, sepia, 0.40)
	mp.land[biomeSnow] = blend(vellum, sepia, 0.20)
	mp.land[biomeSand] = blend(vellum, sepia, 0.16)

	return worldMedium{name: "parchment", style: styleParchment, palette: mp, draw: drawParchment}
}

// drawParchment paints the antique-map medium. Technique: flat vellum sea/land (with a
// subtle aged blotch so the paper isn't sterile), WAVY ink coastlines (each shore pixel
// offset along a low-freq sine so the outline undulates like a hand-inked shore), faint
// sepia biome hints, drawn mountain peaks (little filled triangles) and tree glyphs, thin
// ink rivers — then the chrome: an ornate double frame, a compass rose, and a cartouche
// box top-left behind where the title lands.
func drawParchment(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	seed := uint32(0x9A2C4)

	// 1) Base: vellum land + tan sea, with a faint aged blotch (low-freq noise) so the paper
	//    has gentle mottling rather than a flat wash.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := m.field.at(x, y)
			if bi >= biomeCount {
				bi = biomeGrass
			}
			var col color.RGBA
			if bi == biomeDeepWater || bi == biomeShallowWater {
				e := m.elevAtPx(x, y)
				depth := 0.0
				if m.seaLevel > 0 {
					depth = clamp01((m.seaLevel - e) / m.seaLevel)
				}
				col = blend(p.oceanShelf, p.oceanDeep, depth*0.7)
			} else {
				col = p.land[bi]
			}
			// Aged mottle: a very soft darken keyed off broad noise so the sheet looks foxed.
			blot := valueNoise(float64(x)*0.06, float64(y)*0.06, seed^0xF0)
			col = darken(col, 0.05+0.06*blot)
			setPixel(img, x, y, col)
		}
	}

	// 2) Wavy ink coastline: each shore pixel is offset along a low-frequency sine (of its
	//    position) so the drawn shore undulates like hand-inking, and it's stamped 2px so the
	//    line reads as a confident pen stroke. Gate the offset write to stay on-canvas.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			if isLandPx(m.field, x-1, y) && isLandPx(m.field, x+1, y) &&
				isLandPx(m.field, x, y-1) && isLandPx(m.field, x, y+1) {
				continue
			}
			// Undulate: a small sine offset so the ink wanders off the exact boundary.
			ox := int(math.Round(math.Sin(float64(x)*0.30+float64(y)*0.11) * 1.2))
			oy := int(math.Round(math.Cos(float64(y)*0.30+float64(x)*0.11) * 1.2))
			setPixel(img, x+ox, y+oy, p.coast)
			setPixel(img, x, y, p.coast) // keep a solid core so the line never breaks
		}
	}

	// 3) Rivers: thin muted ink threads over land.
	putRiver := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, r := range m.rivers {
		for i := 0; i+1 < len(r.pts); i++ {
			a, bb := r.pts[i], r.pts[i+1]
			strokeThickLineFunc(a.x, a.y, bb.x, bb.y, 0, p.river, putRiver)
		}
	}

	// 4) Relief: drawn map symbols — little filled ink mountain peaks and tree glyphs, the
	//    illuminated-map vocabulary. Gate to land.
	putR := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, a := range m.reliefs {
		switch a.kind {
		case reliefMountain:
			// A small filled peak: apex, then a widening base — a drawn triangle.
			putR(a.x, a.y-1, p.relief)
			putR(a.x-1, a.y, p.relief)
			putR(a.x, a.y, p.relief)
			putR(a.x+1, a.y, p.relief)
			putR(a.x-1, a.y+1, p.shadow) // hachure feet
			putR(a.x+1, a.y+1, p.shadow)
		case reliefForest:
			// A tree glyph: a little sepia lollipop (crown + stem).
			putR(a.x, a.y-1, p.reliefAlt)
			putR(a.x-1, a.y, p.reliefAlt)
			putR(a.x+1, a.y, p.reliefAlt)
			putR(a.x, a.y, p.reliefAlt)
			putR(a.x, a.y+1, p.shadow) // stem
		case reliefHill:
			// A low sepia mound.
			putR(a.x, a.y, blend(p.relief, p.background, 0.35))
			putR(a.x-1, a.y+1, p.shadow)
			putR(a.x+1, a.y+1, p.shadow)
		}
	}

	// 5) Chrome — the illuminated-map dressing.
	drawParchmentFrame(img, p.coast, p.shadow)            // ornate double frame
	drawCompassRose(img, p.coast, p.relief, p.background) // corner compass
}

// ---- satellite: modern orbital imagery --------------------------------------

// satelliteMedium is a photographic orbital view (→ modern_age). Realistic deep→shallow
// ocean blues, natural land — vegetation greens, brown/grey mountains, white snow, tan
// arid coasts — soft NATURAL blending (no drawn outline; the coast is just where land meets
// sea), and faint cloud wisps. Clean and modern; minimal chrome. It anchors to realistic
// earth tones HARD so it reads as a satellite photo, not a tinted atlas.
func satelliteMedium() worldMedium {
	var mp mediumPalette
	// Real ocean blues.
	mp.oceanDeep = color.RGBA{R: 0x10, G: 0x2a, B: 0x4c, A: 0xff}
	mp.oceanShelf = color.RGBA{R: 0x2f, G: 0x6f, B: 0x9e, A: 0xff}
	mp.background = mp.oceanDeep
	// Natural land: vegetation + earth.
	mp.land[biomeSand] = color.RGBA{R: 0xcf, G: 0xbb, B: 0x82, A: 0xff}     // arid coast tan
	mp.land[biomeGrass] = color.RGBA{R: 0x5f, G: 0x83, B: 0x3f, A: 0xff}    // grass green
	mp.land[biomeForest] = color.RGBA{R: 0x2f, G: 0x52, B: 0x2c, A: 0xff}   // dark forest
	mp.land[biomeRock] = color.RGBA{R: 0x7a, G: 0x6f, B: 0x5c, A: 0xff}     // brown upland
	mp.land[biomeMountain] = color.RGBA{R: 0x8f, G: 0x88, B: 0x7e, A: 0xff} // grey rock
	mp.land[biomeSnow] = color.RGBA{R: 0xed, G: 0xf1, B: 0xf4, A: 0xff}     // snow white
	mp.land[biomeDeepWater] = mp.oceanDeep
	mp.land[biomeShallowWater] = mp.oceanShelf
	mp.coast = mp.land[biomeSand] // used only for the soft coastal lightening
	mp.river = color.RGBA{R: 0x3a, G: 0x62, B: 0x8f, A: 0xff}
	mp.relief = mp.land[biomeMountain]
	mp.reliefAlt = mp.land[biomeSnow]
	mp.shadow = color.RGBA{R: 0x1a, G: 0x22, B: 0x1a, A: 0xff}

	return worldMedium{name: "satellite", style: styleSatellite, palette: mp, draw: drawSatelliteWorld}
}

// drawSatelliteWorld paints the orbital-photo medium. Technique: depth-banded realistic ocean,
// natural land tones with a soft per-pixel vegetation/terrain jitter (so it looks like
// real varied ground rather than flat swatches), NO drawn coastline (a couple-px natural
// beach/surf lightening at the land/sea edge instead of a vector stroke), rivers as thin
// natural blue, snow/rock relief blended softly into the terrain, and a few faint cloud
// wisps drifting over the frame.
func drawSatelliteWorld(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	seed := uint32(0x5A7E1)

	// 1) Base: realistic depth-banded sea + natural land with a soft terrain jitter.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := m.field.at(x, y)
			if bi >= biomeCount {
				bi = biomeGrass
			}
			var col color.RGBA
			if bi == biomeDeepWater || bi == biomeShallowWater {
				e := m.elevAtPx(x, y)
				depth := 0.0
				if m.seaLevel > 0 {
					depth = clamp01((m.seaLevel - e) / m.seaLevel)
				}
				col = blend(p.oceanShelf, p.oceanDeep, depth)
			} else {
				col = p.land[bi]
				// Natural variation: fine noise lightens/darkens the ground a hair so vegetation
				// and terrain read organic (a photo is never one flat color).
				n := valueNoise(float64(x)*0.22, float64(y)*0.22, seed^0x9)
				if n > 0.5 {
					col = brighten(col, (n-0.5)*0.20)
				} else {
					col = darken(col, (0.5-n)*0.20)
				}
			}
			setPixel(img, x, y, col)
		}
	}

	// 2) Soft coast: NO vector outline. Instead, land pixels at the shore get a gentle
	//    beach/surf lightening (a natural bright rim) and shallow-sea pixels adjacent to land
	//    lift toward a surf tone — a photographic transition, not a drawn line.
	surf := brighten(p.oceanShelf, 0.22)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if isLandPx(m.field, x, y) {
				// Shore land: lighten slightly if it borders water (a bright beach edge).
				if !isLandPx(m.field, x-1, y) || !isLandPx(m.field, x+1, y) ||
					!isLandPx(m.field, x, y-1) || !isLandPx(m.field, x, y+1) {
					setPixel(img, x, y, brighten(img.RGBAAt(b.Min.X+x, b.Min.Y+y), 0.14))
				}
				continue
			}
			// Shallow sea bordering land → a soft surf lift (only shallow, keeps deep sea calm).
			if m.field.at(x, y) == biomeShallowWater {
				if isLandPx(m.field, x-1, y) || isLandPx(m.field, x+1, y) ||
					isLandPx(m.field, x, y-1) || isLandPx(m.field, x, y+1) {
					cur := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
					setPixel(img, x, y, blend(cur, surf, 0.35))
				}
			}
		}
	}

	// 3) Rivers: thin natural blue, gated to land.
	putRiver := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, r := range m.rivers {
		for i := 0; i+1 < len(r.pts); i++ {
			a, bb := r.pts[i], r.pts[i+1]
			strokeThickLineFunc(a.x, a.y, bb.x, bb.y, 0, p.river, putRiver)
		}
	}

	// 4) Relief: soft, photographic. Snow caps brighten, rock darkens — blended into the
	//    terrain (no hard symbols; a satellite photo shows landform, not icons).
	for _, a := range m.reliefs {
		if !isLandPx(m.field, a.x, a.y) {
			continue
		}
		cur := img.RGBAAt(b.Min.X+a.x, b.Min.Y+a.y)
		switch a.kind {
		case reliefMountain:
			if m.field.at(a.x, a.y) == biomeSnow {
				setPixel(img, a.x, a.y, blend(cur, p.reliefAlt, 0.40))
				setPixel(img, a.x, a.y-1, blend(img.RGBAAt(b.Min.X+a.x, b.Min.Y+maxInt(a.y-1, 0)), p.reliefAlt, 0.30))
			} else {
				setPixel(img, a.x, a.y, darken(cur, 0.12))       // shaded rock face
				setPixel(img, a.x-1, a.y-1, brighten(cur, 0.10)) // lit ridge
			}
		case reliefHill:
			setPixel(img, a.x, a.y, darken(cur, 0.06))
		}
	}

	// 5) Faint clouds: a couple of soft white wisps drifting over the frame (low-freq noise
	//    thresholded high, feathered), so it reads as a live orbital shot.
	drawClouds(img, seed^0xC10D)
}

// ---- neon: cyberpunk holographic projection ---------------------------------

// neonMedium is a holographic projection (→ cyberpunk_age). Near-black background, land as
// a glowing NEON WIREFRAME — bright cyan/magenta coastline glow, contour + grid lines over
// the land, rivers as bright data-lines — a faint holographic grid across the WHOLE frame,
// relief as neon glyphs, and a HUD-ish frame. High-tech, dark, luminous. Anchors to
// black + cyan/magenta HARD.
func neonMedium() worldMedium {
	var mp mediumPalette
	black := color.RGBA{R: 0x05, G: 0x07, B: 0x0c, A: 0xff}
	cyan := color.RGBA{R: 0x22, G: 0xe6, B: 0xd8, A: 0xff}
	magenta := color.RGBA{R: 0xe4, G: 0x33, B: 0xb0, A: 0xff}
	mp.background = black
	mp.oceanDeep = black
	mp.oceanShelf = color.RGBA{R: 0x08, G: 0x0e, B: 0x18, A: 0xff} // barely-lit sea grid bed
	mp.coast = cyan                                                // glowing coastline
	mp.river = color.RGBA{R: 0x4a, G: 0xa8, B: 0xff, A: 0xff}      // bright data-line blue
	mp.relief = magenta                                            // neon relief glyphs
	mp.reliefAlt = cyan
	mp.shadow = color.RGBA{R: 0x10, G: 0x2a, B: 0x30, A: 0xff} // dim cyan land wash
	// Land is a very dark wireframe bed; the glow lines carry the read, not the fill.
	for bi := biome(0); bi < biomeCount; bi++ {
		mp.land[bi] = color.RGBA{R: 0x0a, G: 0x16, B: 0x1c, A: 0xff}
	}
	mp.land[biomeMountain] = color.RGBA{R: 0x14, G: 0x0e, B: 0x1e, A: 0xff} // faint magenta bed for highlands
	mp.land[biomeSnow] = color.RGBA{R: 0x14, G: 0x0e, B: 0x1e, A: 0xff}

	return worldMedium{name: "neon", style: styleNeon, palette: mp, draw: drawNeon}
}

// drawNeon paints the holographic medium. Technique: near-black frame, a faint cyan
// holographic GRID over the whole canvas, a very dark land "bed" tinted per elevation,
// contour lines glowing cyan along elevation isolines over the land, a bright cyan/magenta
// coastline glow (the shore, plus a 1px bloom just outside it into the sea), rivers as
// bright data-lines, relief as magenta neon glyphs, and a HUD corner-bracket frame.
func drawNeon(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	cyan := p.coast

	// 1) Base: black sea, dark land bed. A subtle elevation lift so the land isn't a flat
	//    slab (higher ground glows a hair more toward the bed's tint).
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := m.field.at(x, y)
			if bi >= biomeCount {
				bi = biomeGrass
			}
			var col color.RGBA
			if bi == biomeDeepWater || bi == biomeShallowWater {
				col = p.oceanDeep
			} else {
				col = p.land[bi]
				e := m.elevAtPx(x, y)
				if e > m.seaLevel {
					t := (e - m.seaLevel) / (1 - m.seaLevel)
					col = blend(col, p.shadow, clamp01(t)*0.5) // higher ground → brighter cyan bed
				}
			}
			setPixel(img, x, y, col)
		}
	}

	// 2) Holographic grid over the WHOLE frame: dim cyan lines every gridStep px, so the
	//    projection reads as a HUD surface even over the sea. Kept faint (blended) so it's a
	//    substrate, not clutter.
	gridStep := 8
	gridCol := blend(p.background, cyan, 0.14)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x%gridStep == 0 || y%gridStep == 0 {
				cur := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
				setPixel(img, x, y, blend(cur, gridCol, 0.6))
			}
		}
	}

	// 3) Contour lines on the land: quantize elevation into bands; a pixel on a band boundary
	//    (its band differs from the +x or +y neighbour's) glows dim cyan — topographic isolines
	//    over the wireframe land.
	contour := blend(p.background, cyan, 0.45)
	bandOf := func(x, y int) int {
		if !isLandPx(m.field, x, y) {
			return -1
		}
		e := m.elevAtPx(x, y)
		return int((e - m.seaLevel) / (1 - m.seaLevel) * 6)
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bnd := bandOf(x, y)
			if bnd < 0 {
				continue
			}
			if bandOf(x+1, y) != bnd || bandOf(x, y+1) != bnd {
				cur := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
				setPixel(img, x, y, blend(cur, contour, 0.7))
			}
		}
	}

	// 4) Coastline glow: the shore pixels bright cyan, PLUS a 1px magenta bloom just OUTSIDE
	//    the shore (into the sea) so the continent edge blooms like a projected hologram.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			shore := !isLandPx(m.field, x-1, y) || !isLandPx(m.field, x+1, y) ||
				!isLandPx(m.field, x, y-1) || !isLandPx(m.field, x, y+1)
			if !shore {
				continue
			}
			setPixel(img, x, y, cyan)
			// Bloom into the adjacent sea pixels (magenta halo).
			bloom := blend(p.background, p.relief, 0.5)
			for _, nb := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+nb[0], y+nb[1]
				if !isLandPx(m.field, nx, ny) {
					cur := img.RGBAAt(b.Min.X+clampInt(nx, 0, w-1), b.Min.Y+clampInt(ny, 0, h-1))
					setPixel(img, nx, ny, blend(cur, bloom, 0.5))
				}
			}
		}
	}

	// 5) Rivers: bright data-lines over the land.
	putRiver := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, r := range m.rivers {
		for i := 0; i+1 < len(r.pts); i++ {
			a, bb := r.pts[i], r.pts[i+1]
			strokeThickLineFunc(a.x, a.y, bb.x, bb.y, 0, p.river, putRiver)
		}
	}

	// 6) Relief: neon glyphs. Mountains a bright magenta chevron; hills a small cyan tick.
	putR := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, a := range m.reliefs {
		switch a.kind {
		case reliefMountain:
			putR(a.x, a.y-1, p.relief)
			putR(a.x-1, a.y, p.relief)
			putR(a.x+1, a.y, p.relief)
		case reliefHill:
			putR(a.x, a.y, p.reliefAlt)
		}
	}

	// 7) HUD frame: bright cyan corner brackets (not a full box — a targeting-reticle read).
	drawHUDBrackets(img, cyan)
}

// ---- petroglyph: stone-age rock carving -------------------------------------

// petroglyphMedium is a rock carving (→ stone_age). Deliberately the ANTI-charcoal: where
// charcoal is a dark slate sketch, this is a LIGHT warm SANDSTONE surface. A buff/tan rock
// face grained with stone speckle; land is shown by STIPPLED / PECKED texture (clusters of
// small hash-placed dots) rather than a solid fill; the coast is a carved GROOVE (a recessed
// dark line with a lighter lip beside it); relief is angular pecked mountain glyphs in
// red-ochre / iron-oxide pigment; rivers are pecked dotted grooves. Chrome: a rough,
// irregular rock-edge border. Anchors to warm buff + red-ochre HARD — a carving is a
// carving regardless of theme.
func petroglyphMedium() worldMedium {
	// Fixed sandstone identity (theme-independent by design).
	sand := color.RGBA{R: 0xcb, G: 0xa9, B: 0x77, A: 0xff}     // warm buff rock face (land bed)
	sandSea := color.RGBA{R: 0xb0, G: 0x94, B: 0x69, A: 0xff}  // sea: a duller, shaded rock recess
	sandDeep := color.RGBA{R: 0x97, G: 0x7d, B: 0x57, A: 0xff} // deeper recess
	peck := color.RGBA{R: 0x6b, G: 0x4f, B: 0x2f, A: 0xff}     // dark pecked dot (chipped stone)
	groove := color.RGBA{R: 0x4c, G: 0x37, B: 0x22, A: 0xff}   // recessed carved-line shadow
	lip := color.RGBA{R: 0xe4, G: 0xcb, B: 0x9d, A: 0xff}      // sunlit lip beside a groove
	ochre := color.RGBA{R: 0xa8, G: 0x45, B: 0x2b, A: 0xff}    // red-ochre / iron-oxide pigment (relief)

	var mp mediumPalette
	mp.background = sand
	mp.oceanDeep = sandDeep
	mp.oceanShelf = sandSea
	mp.coast = groove
	mp.reliefAlt = lip
	mp.river = groove
	mp.relief = ochre
	mp.shadow = peck
	// Land is one warm rock; the pecked stipple (not hue) carries the read.
	for bi := biome(0); bi < biomeCount; bi++ {
		mp.land[bi] = sand
	}

	return worldMedium{name: "petroglyph", style: stylePetroglyph, palette: mp, draw: drawPetroglyph}
}

// drawPetroglyph paints the rock-carving medium. Technique: a light buff rock face grained
// with a stable stone speckle everywhere; the SEA sits as a slightly darker recessed rock
// (barely banded); the LAND is peppered with pecked dots — dense clusters of small dark
// chips placed by a coordinate hash so the continent reads as stippled/pecked rather than
// filled; the coastline is a carved GROOVE (a dark recessed line with a bright sunlit lip on
// the seaward side); rivers are pecked dotted grooves; relief is angular red-ochre mountain
// glyphs pecked into the rock; and a rough irregular rock-edge border rims the frame.
func drawPetroglyph(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	seed := uint32(0x9E720)

	// 1) Base: warm buff rock face with a fine stone speckle; sea a shaded recess.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := m.field.at(x, y)
			if bi >= biomeCount {
				bi = biomeGrass
			}
			var col color.RGBA
			if bi == biomeDeepWater || bi == biomeShallowWater {
				e := m.elevAtPx(x, y)
				depth := 0.0
				if m.seaLevel > 0 {
					depth = clamp01((m.seaLevel - e) / m.seaLevel)
				}
				col = blend(p.oceanShelf, p.oceanDeep, depth*0.6)
			} else {
				col = p.land[bi]
			}
			// Stone speckle: a fine per-pixel hash grain so the rock never reads as a flat wash.
			g := hashUnit(uint32(x), uint32(y), seed^0x5EED)
			if g > 0.80 {
				col = brighten(col, 0.10) // a bright mineral fleck
			} else if g < 0.16 {
				col = darken(col, 0.12) // a dark pit
			}
			setPixel(img, x, y, col)
		}
	}

	// 2) Pecked stipple over the LAND: clusters of small dark chips. A coarse cluster field
	//    (low-freq noise) decides WHERE pecking is dense; within a dense zone a per-pixel hash
	//    decides which pixels actually get a chip, so the land reads as hand-pecked stone, not
	//    a solid fill. Gated to land so the sea recess stays smooth.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			cluster := valueNoise(float64(x)*0.30, float64(y)*0.30, seed^0xC1)
			// Density ramps with the cluster field: bare in the gaps, thick in the clusters.
			dens := 0.14 + 0.42*cluster
			if hashUnit(uint32(x), uint32(y), seed^0xDA7) < dens {
				setPixel(img, x, y, p.shadow) // a pecked chip
			}
		}
	}

	// 3) Carved-groove coastline: every shore pixel gets a recessed dark groove, with a bright
	//    sunlit LIP stamped on the SEAWARD neighbour — a double line that reads as a chiseled
	//    channel, not a drawn stroke.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			if isLandPx(m.field, x-1, y) && isLandPx(m.field, x+1, y) &&
				isLandPx(m.field, x, y-1) && isLandPx(m.field, x, y+1) {
				continue
			}
			setPixel(img, x, y, p.coast) // the recessed groove itself
			// Sunlit lip: brighten the seaward neighbours so the groove looks carved (recess + lip).
			for _, nb := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+nb[0], y+nb[1]
				if !isLandPx(m.field, nx, ny) {
					setPixel(img, nx, ny, p.reliefAlt)
				}
			}
		}
	}

	// 4) Rivers: pecked dotted grooves over the land (skip alternating steps for a chipped read).
	step := 0
	putRiver := func(x, y int, c color.RGBA) {
		if !isLandPx(m.field, x, y) {
			return
		}
		step++
		if step%2 == 0 {
			return // dotted / pecked
		}
		setPixel(img, x, y, c)
	}
	for _, r := range m.rivers {
		for i := 0; i+1 < len(r.pts); i++ {
			a, bb := r.pts[i], r.pts[i+1]
			strokeThickLineFunc(a.x, a.y, bb.x, bb.y, 0, p.river, putRiver)
		}
	}

	// 5) Relief: angular pecked mountain glyphs in red-ochre pigment — a hard zig-zag apex,
	//    not a filled peak. Hills a short ochre notch. Forests skipped (a carving keeps the big
	//    landmarks). Gate to land.
	putR := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, a := range m.reliefs {
		switch a.kind {
		case reliefMountain:
			// An angular pecked chevron: apex, two long flanks stepping down — sharp, carved.
			putR(a.x, a.y-2, p.relief)
			putR(a.x-1, a.y-1, p.relief)
			putR(a.x+1, a.y-1, p.relief)
			putR(a.x-2, a.y, p.relief)
			putR(a.x+2, a.y, p.relief)
		case reliefHill:
			putR(a.x, a.y-1, p.relief) // a short ochre notch
			putR(a.x-1, a.y, p.relief)
			putR(a.x+1, a.y, p.relief)
		}
	}

	// 6) Rough rock-edge border: an irregular chipped rim, broken by a hash so it reads as the
	//    ragged edge of a rock face rather than a ruled frame.
	drawRockEdgeBorder(img, p.coast, p.shadow, seed^0xB0DE)
}

// ---- clay tablet: bronze-age cuneiform tablet -------------------------------

// clayTabletMedium is a wet-clay cuneiform tablet (→ bronze_age). A UNIFORM terracotta clay
// surface — the land carries NO biome color (it is all one clay); biome is conveyed by the
// DENSITY of impressed marks, not hue. Everything is IMPRESSED/incised: the coastline is a
// deep pressed GROOVE (a dark shadow in the channel plus a lighter raised lip beside it, so
// it reads 3D-pressed into the clay); relief is wedge / triangle stamps; rivers are incised
// channels. Chrome: a rounded clay tablet edge with one broken-off corner chip. Anchors to
// orange-brown clay + monochrome pressing HARD.
func clayTabletMedium() worldMedium {
	// Fixed wet-clay identity (theme-independent by design).
	clay := color.RGBA{R: 0xbf, G: 0x7a, B: 0x48, A: 0xff}     // uniform terracotta clay
	claySea := color.RGBA{R: 0xa5, G: 0x66, B: 0x3a, A: 0xff}  // sea area: clay pressed a touch deeper
	clayDeep := color.RGBA{R: 0x8c, G: 0x55, B: 0x30, A: 0xff} // deepest press
	press := color.RGBA{R: 0x6e, G: 0x40, B: 0x22, A: 0xff}    // shadow inside an impression
	pressDk := color.RGBA{R: 0x55, G: 0x30, B: 0x19, A: 0xff}  // deep groove shadow
	raised := color.RGBA{R: 0xe1, G: 0xa5, B: 0x6f, A: 0xff}   // sunlit raised lip / ridge of a press
	edge := color.RGBA{R: 0x4d, G: 0x2c, B: 0x17, A: 0xff}     // tablet edge shadow

	var mp mediumPalette
	mp.background = clay
	mp.oceanDeep = clayDeep
	mp.oceanShelf = claySea
	mp.coast = pressDk
	mp.reliefAlt = raised
	mp.river = press
	mp.relief = press
	mp.shadow = edge
	// Land is one clay — the impressed-mark DENSITY (below) carries the biome read, not hue.
	for bi := biome(0); bi < biomeCount; bi++ {
		mp.land[bi] = clay
	}

	return worldMedium{name: "clay_tablet", style: styleClayTablet, palette: mp, draw: drawClayTablet}
}

// drawClayTablet paints the wet-clay-tablet medium. Technique: a uniform terracotta clay
// surface (sea pressed a touch deeper) with a fine clay-grain jitter; over the LAND, biome is
// conveyed by the DENSITY of tiny impressed marks — forest/mountain/rock stamp denser than
// grass/sand — so the terrain reads through pressing, never hue; the coastline is a deep
// pressed GROOVE, drawn as a dark shadow channel on the shore with a lighter RAISED lip on
// the seaward side so it reads pushed 3D into the clay; relief is small wedge/triangle
// stamps; rivers are incised channels; and the frame is a rounded tablet edge with one
// broken-off corner chip.
func drawClayTablet(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	seed := uint32(0xC1A70)

	// biomeMarkDensity maps a land biome to how thickly it is stippled with impressed marks.
	// This is the ONLY channel that conveys terrain on the clay (all land is one color), so the
	// ramp is wide: open grass/sand barely marked, forest/rock/mountain pressed heavily.
	markDensity := func(bi biome) float64 {
		switch bi {
		case biomeForest:
			return 0.34
		case biomeRock:
			return 0.26
		case biomeMountain, biomeSnow:
			return 0.40
		case biomeGrass:
			return 0.10
		case biomeSand:
			return 0.05
		default:
			return 0.0
		}
	}

	// 1) Base: uniform clay (sea pressed a shade deeper) with a fine clay-grain jitter so the
	//    surface reads worked rather than flat.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := m.field.at(x, y)
			if bi >= biomeCount {
				bi = biomeGrass
			}
			var col color.RGBA
			if bi == biomeDeepWater || bi == biomeShallowWater {
				e := m.elevAtPx(x, y)
				depth := 0.0
				if m.seaLevel > 0 {
					depth = clamp01((m.seaLevel - e) / m.seaLevel)
				}
				col = blend(p.oceanShelf, p.oceanDeep, depth*0.7)
			} else {
				col = p.land[bi]
			}
			n := valueNoise(float64(x)*0.5, float64(y)*0.5, seed^0x6)
			col = blend(col, darken(col, 0.10), n*0.5) // fine worked-clay grain
			setPixel(img, x, y, col)
		}
	}

	// 2) Impressed-mark density over LAND: each pixel gets a small pressed dot (a dark press
	//    with a 1px raised highlight above it) at a probability keyed by the biome's density,
	//    so terrain reads purely through how thickly the clay is stippled.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			bi := m.field.at(x, y)
			if hashUnit(uint32(x), uint32(y), seed^0x1417) < markDensity(bi) {
				setPixel(img, x, y, p.river)       // pressed shadow (river tone == press shadow)
				setPixel(img, x, y-1, p.reliefAlt) // a tiny raised lip above the press
			}
		}
	}

	// 3) Pressed-groove coastline: the shore pixel becomes a deep groove shadow, and the
	//    SEAWARD neighbour gets a raised lip — a dark channel + light ridge that reads pushed
	//    into the clay in 3D. A second inland pixel deepens the channel so the groove has body.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			if isLandPx(m.field, x-1, y) && isLandPx(m.field, x+1, y) &&
				isLandPx(m.field, x, y-1) && isLandPx(m.field, x, y+1) {
				continue
			}
			setPixel(img, x, y, p.coast) // deep groove shadow on the shore
			for _, nb := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+nb[0], y+nb[1]
				if !isLandPx(m.field, nx, ny) {
					setPixel(img, nx, ny, p.reliefAlt) // raised sunlit lip on the seaward side
				}
			}
		}
	}

	// 4) Rivers: incised channels — a pressed shadow line with a faint raised lip alongside.
	putRiver := func(x, y int, c color.RGBA) {
		if !isLandPx(m.field, x, y) {
			return
		}
		setPixel(img, x, y, c)
		setPixel(img, x, y-1, p.reliefAlt) // raised bank
	}
	for _, r := range m.rivers {
		for i := 0; i+1 < len(r.pts); i++ {
			a, bb := r.pts[i], r.pts[i+1]
			strokeThickLineFunc(a.x, a.y, bb.x, bb.y, 0, p.river, putRiver)
		}
	}

	// 5) Relief: wedge / triangle stamps — the cuneiform vocabulary. Mountains a filled wedge
	//    (pressed shadow) with a raised top edge; hills a single small wedge. Gate to land.
	putR := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, a := range m.reliefs {
		switch a.kind {
		case reliefMountain:
			// A pressed wedge: apex raised, body a dark impression widening down.
			putR(a.x, a.y-1, p.reliefAlt) // raised apex
			putR(a.x-1, a.y, p.relief)
			putR(a.x, a.y, p.relief)
			putR(a.x+1, a.y, p.relief)
			putR(a.x, a.y+1, p.shadow) // deep foot
		case reliefHill:
			putR(a.x, a.y, p.relief) // a single small wedge
			putR(a.x, a.y-1, p.reliefAlt)
		}
	}

	// 6) Chrome: rounded clay tablet edge (a beveled rim) with one broken-off corner chip.
	drawClayTabletEdge(img, p.shadow, p.reliefAlt, seed^0xC00E)
}

// ---- hide: iron-age tanned leather map --------------------------------------

// hideMedium is a tanned leather / hide map (→ iron_age). A warm MID-BROWN leather (redder /
// warmer than the clay terracotta so the two don't read alike) with a pore/grain mottle and
// darker organic blotches. The SIGNATURE is BURNT scorched edges: a dark charred vignette all
// around the frame, heaviest in the corners. The drawing is soot-black + blood-ochre INK on
// the hide — a confident dark ink coastline stroke, simple ink triangle relief, thin ink
// rivers. Anchors to warm leather + burnt-edge + ink HARD.
func hideMedium() worldMedium {
	// Fixed tanned-hide identity (theme-independent by design). Redder/warmer than clay.
	hideCol := color.RGBA{R: 0x9c, G: 0x5f, B: 0x38, A: 0xff}    // warm mid-brown leather
	hideSea := color.RGBA{R: 0x86, G: 0x51, B: 0x31, A: 0xff}    // sea: a shade deeper hide
	hideDeep := color.RGBA{R: 0x70, G: 0x43, B: 0x28, A: 0xff}   // deepest hide
	soot := color.RGBA{R: 0x24, G: 0x18, B: 0x11, A: 0xff}       // soot-black ink (coast/relief)
	bloodOchre := color.RGBA{R: 0x8f, G: 0x33, B: 0x22, A: 0xff} // blood-ochre ink accent
	blotch := color.RGBA{R: 0x6f, G: 0x42, B: 0x27, A: 0xff}     // darker organic hide blotch

	var mp mediumPalette
	mp.background = hideCol
	mp.oceanDeep = hideDeep
	mp.oceanShelf = hideSea
	mp.coast = soot
	mp.river = blend(soot, hideCol, 0.30) // thin muted ink
	mp.relief = soot
	mp.reliefAlt = bloodOchre
	mp.shadow = blotch
	// Land is one leather; the pore mottle + blotches (not hue) carry the surface read.
	for bi := biome(0); bi < biomeCount; bi++ {
		mp.land[bi] = hideCol
	}

	return worldMedium{name: "hide", style: styleHide, palette: mp, draw: drawHide}
}

// drawHide paints the tanned-leather medium. Technique: a warm mid-brown hide with a fine
// pore/grain mottle plus broad darker organic blotches (low-freq noise) so it reads as tanned
// skin; a confident soot-black INK coastline stroke (stamped 2px, solid — a pen on leather);
// thin muted ink rivers; simple soot ink triangle relief with a blood-ochre accent; and the
// SIGNATURE burnt scorched vignette — a dark charred darkening that grows toward every edge
// and piles heaviest in the corners, so the map looks singed at the margins.
func drawHide(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	seed := uint32(0x41DE0)

	// 1) Base: warm hide with a pore mottle + broad organic blotches.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := m.field.at(x, y)
			if bi >= biomeCount {
				bi = biomeGrass
			}
			var col color.RGBA
			if bi == biomeDeepWater || bi == biomeShallowWater {
				e := m.elevAtPx(x, y)
				depth := 0.0
				if m.seaLevel > 0 {
					depth = clamp01((m.seaLevel - e) / m.seaLevel)
				}
				col = blend(p.oceanShelf, p.oceanDeep, depth*0.6)
			} else {
				col = p.land[bi]
			}
			// Fine pore grain.
			g := hashUnit(uint32(x), uint32(y), seed^0x9)
			if g > 0.82 {
				col = brighten(col, 0.07)
			} else if g < 0.18 {
				col = darken(col, 0.08)
			}
			// Broad organic blotches (foxed/tanned unevenness).
			blot := valueNoise(float64(x)*0.07, float64(y)*0.07, seed^0xB1)
			if blot > 0.60 {
				col = blend(col, p.shadow, (blot-0.60)/0.40*0.35)
			}
			setPixel(img, x, y, col)
		}
	}

	// 2) Coastline: a confident soot-black ink stroke, stamped 2px so it reads as a deliberate
	//    pen line on the hide (solid, not broken).
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			if isLandPx(m.field, x-1, y) && isLandPx(m.field, x+1, y) &&
				isLandPx(m.field, x, y-1) && isLandPx(m.field, x, y+1) {
				continue
			}
			setPixel(img, x, y, p.coast)
			// Thicken toward the sea by one pixel so the ink has weight.
			for _, nb := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+nb[0], y+nb[1]
				if !isLandPx(m.field, nx, ny) {
					setPixel(img, nx, ny, p.coast)
					break
				}
			}
		}
	}

	// 3) Rivers: thin muted ink threads over land.
	putRiver := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, r := range m.rivers {
		for i := 0; i+1 < len(r.pts); i++ {
			a, bb := r.pts[i], r.pts[i+1]
			strokeThickLineFunc(a.x, a.y, bb.x, bb.y, 0, p.river, putRiver)
		}
	}

	// 4) Relief: simple soot ink triangles, with a blood-ochre accent on the peak so the ink
	//    map has one warm pigment. Hills a small ochre tick. Gate to land.
	putR := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, a := range m.reliefs {
		switch a.kind {
		case reliefMountain:
			putR(a.x, a.y-1, p.reliefAlt) // ochre apex
			putR(a.x-1, a.y, p.relief)    // soot base
			putR(a.x, a.y, p.relief)
			putR(a.x+1, a.y, p.relief)
		case reliefHill:
			putR(a.x, a.y, p.reliefAlt) // a small ochre tick
		}
	}

	// 5) SIGNATURE chrome: burnt scorched vignette — a charred darkening ramping toward every
	//    edge, heaviest in the corners, so the hide looks singed at the margins.
	drawScorchedEdges(img, seed^0xF12E)
}

// ---- mosaic: classical-age Roman tesserae -----------------------------------

// mosaicMedium is a Roman tesserae mosaic (→ classical_age). The ENTIRE map is built from
// small square TILES: the world is quantized to a tile grid (~4px cells), each tile one
// slightly hash-jittered tone, with thin dark GROUT lines between tiles. Sea = blue/teal
// tesserae; land = terracotta / cream / olive tesserae keyed by biome; the coastline is a run
// of white/black tesserae; relief = darker tile clusters; rivers = blue tile runs. Chrome: a
// classical Greek-key / meander geometric border. The visible TILE GRID + grout is the
// signature — nothing else in the system is tiled. Anchors to mosaic tesserae palette HARD.
func mosaicMedium() worldMedium {
	var mp mediumPalette
	// Tesserae palette — glassy blues for sea, warm stone for land.
	mp.oceanDeep = color.RGBA{R: 0x1d, G: 0x54, B: 0x6b, A: 0xff}           // deep teal-blue tesserae
	mp.oceanShelf = color.RGBA{R: 0x37, G: 0x86, B: 0x99, A: 0xff}          // shallow teal tesserae
	mp.background = color.RGBA{R: 0x14, G: 0x3f, B: 0x52, A: 0xff}          // grout-dark base behind tiles
	mp.land[biomeSand] = color.RGBA{R: 0xe4, G: 0xd2, B: 0xa6, A: 0xff}     // cream tesserae
	mp.land[biomeGrass] = color.RGBA{R: 0x8f, G: 0x93, B: 0x4e, A: 0xff}    // olive tesserae
	mp.land[biomeForest] = color.RGBA{R: 0x53, G: 0x6b, B: 0x38, A: 0xff}   // deep olive/green tesserae
	mp.land[biomeRock] = color.RGBA{R: 0xb4, G: 0x7c, B: 0x50, A: 0xff}     // terracotta tesserae
	mp.land[biomeMountain] = color.RGBA{R: 0x8a, G: 0x55, B: 0x3c, A: 0xff} // dark terracotta tesserae
	mp.land[biomeSnow] = color.RGBA{R: 0xea, G: 0xe3, B: 0xd6, A: 0xff}     // pale marble tesserae
	mp.land[biomeDeepWater] = mp.oceanDeep
	mp.land[biomeShallowWater] = mp.oceanShelf
	mp.coast = color.RGBA{R: 0xf1, G: 0xea, B: 0xdc, A: 0xff}     // white/cream coast tesserae
	mp.river = color.RGBA{R: 0x3f, G: 0x92, B: 0xb8, A: 0xff}     // blue river tesserae
	mp.relief = color.RGBA{R: 0x5b, G: 0x38, B: 0x28, A: 0xff}    // dark tile cluster (relief)
	mp.reliefAlt = color.RGBA{R: 0x20, G: 0x1c, B: 0x18, A: 0xff} // near-black coast tesserae accent
	mp.shadow = color.RGBA{R: 0x11, G: 0x2a, B: 0x33, A: 0xff}    // grout line color
	return worldMedium{name: "mosaic", style: styleMosaic, palette: mp, draw: drawMosaic}
}

// drawMosaic paints the tesserae-mosaic medium. Technique: the world is quantized to a TILE
// grid — each tile samples the biome at its center, fills the whole cell with that biome's
// tesserae tone (hash-jittered per tile so no two tiles are quite alike), and leaves a thin
// dark GROUT gap on the tile's top/left edges so the whole map reads as laid tile. The
// coastline is retiled as a run of white/near-black tesserae; relief tiles darken; rivers lay
// a blue tile run; and a classical Greek-key / meander border frames the frame. The visible
// tile+grout grid is the signature.
func drawMosaic(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	seed := uint32(0x70551)
	const tile = 4 // tesserae cell size in px (grid + grout)

	// tileBiomeAt classifies a tile by the biome at its CENTER pixel (clamped on-canvas), so
	// each tile is one clean terrain class rather than a blur of several.
	tileBiomeAt := func(tx, ty int) biome {
		cx := clampInt(tx+tile/2, 0, w-1)
		cy := clampInt(ty+tile/2, 0, h-1)
		bi := m.field.at(cx, cy)
		if bi >= biomeCount {
			bi = biomeGrass
		}
		return bi
	}
	// tileIsLand reports whether a tile's center is land (for the coast/relief retiling).
	tileIsLand := func(tx, ty int) bool {
		cx := clampInt(tx+tile/2, 0, w-1)
		cy := clampInt(ty+tile/2, 0, h-1)
		return isLandPx(m.field, cx, cy)
	}

	// fillTile paints a whole tile cell in tone `col`, leaving a 1px dark grout gap on the
	// tile's top and left edges so adjacent tiles read as separate tesserae.
	fillTile := func(tx, ty int, col color.RGBA) {
		for yy := ty; yy < ty+tile && yy < h; yy++ {
			for xx := tx; xx < tx+tile && xx < w; xx++ {
				if xx == tx || yy == ty {
					setPixel(img, xx, yy, p.shadow) // grout on the top/left edges
					continue
				}
				setPixel(img, xx, yy, col)
			}
		}
	}

	// 1) Lay the base field as tiles: each tile one biome tesserae tone, hash-jittered so the
	//    mosaic isn't mechanically uniform. Sea tiles depth-shade a touch toward deep.
	for ty := 0; ty < h; ty += tile {
		for tx := 0; tx < w; tx += tile {
			bi := tileBiomeAt(tx, ty)
			col := p.land[bi]
			if bi == biomeDeepWater || bi == biomeShallowWater {
				e := m.elevAtPx(clampInt(tx+tile/2, 0, w-1), clampInt(ty+tile/2, 0, h-1))
				depth := 0.0
				if m.seaLevel > 0 {
					depth = clamp01((m.seaLevel - e) / m.seaLevel)
				}
				col = blend(p.oceanShelf, p.oceanDeep, depth)
			}
			// Per-tile jitter: nudge each tesserae lighter/darker so the surface glints.
			j := hashUnit(uint32(tx), uint32(ty), seed^0x7)
			if j > 0.5 {
				col = brighten(col, (j-0.5)*0.22)
			} else {
				col = darken(col, (0.5-j)*0.22)
			}
			fillTile(tx, ty, col)
		}
	}

	// 2) Coastline: retile the shore tiles as a run of white/near-black tesserae (alternating
	//    by tile parity for a classic banded mosaic border). A shore tile is a land tile with a
	//    water tile 4-neighbour.
	for ty := 0; ty < h; ty += tile {
		for tx := 0; tx < w; tx += tile {
			if !tileIsLand(tx, ty) {
				continue
			}
			shore := !tileIsLand(tx-tile, ty) || !tileIsLand(tx+tile, ty) ||
				!tileIsLand(tx, ty-tile) || !tileIsLand(tx, ty+tile)
			if !shore {
				continue
			}
			col := p.coast
			if ((tx/tile)+(ty/tile))%2 == 0 {
				col = p.reliefAlt // the near-black tesserae in the alternating run
			}
			fillTile(tx, ty, col)
		}
	}

	// 3) Rivers: lay a run of blue tesserae along each course (snap each river point to its
	//    tile). Gate to land tiles so the run stops at the coast.
	for _, r := range m.rivers {
		for _, pt := range r.pts {
			tx := (pt.x / tile) * tile
			ty := (pt.y / tile) * tile
			if tileIsLand(tx, ty) {
				fillTile(tx, ty, p.river)
			}
		}
	}

	// 4) Relief: darker tile clusters — retile each relief anchor's tile (and, for mountains, a
	//    couple of neighbours) to a dark tone so uplands read as a knot of darker tesserae.
	for _, a := range m.reliefs {
		tx := (a.x / tile) * tile
		ty := (a.y / tile) * tile
		if !tileIsLand(tx, ty) {
			continue
		}
		switch a.kind {
		case reliefMountain:
			fillTile(tx, ty, p.relief)
			if tileIsLand(tx, ty-tile) {
				fillTile(tx, ty-tile, p.relief) // a taller dark cluster
			}
		case reliefHill:
			fillTile(tx, ty, darken(p.land[biomeRock], 0.18))
		}
	}

	// 5) Chrome: a classical Greek-key / meander geometric border around the frame.
	drawGreekKeyBorder(img, p.coast, p.shadow)
}

// ---- copperplate: renaissance engraving -------------------------------------

// copperplateMedium is a fine copperplate ENGRAVING (→ renaissance_age). Cream/ivory paper
// where EVERYTHING is fine engraved LINES (burin strokes), never a solid fill. The SEA is
// fine RULED horizontal line-shading — dense near the coast, thinning to open paper offshore
// — so water reads as ruled hatching, not a tint. The LAND is mostly blank ivory paper, with
// finely CROSS-HATCHED relief only where hills/mountains sit. The coastline is one crisp fine
// engraved line with a thin band of hatch-shading on the sea side. Rivers are thin single
// engraved lines. Chrome: an elegant engraved neat-line border (a fine double rule) with small
// graticule tick marks along the margins. Anchors to cream paper + fine dark-ink line HARD —
// an engraving is line-shaded and precise, never the solid cream fills + wavy coast of
// parchment.
func copperplateMedium() worldMedium {
	// Fixed engraving identity (theme-independent by design).
	ivory := color.RGBA{R: 0xef, G: 0xe7, B: 0xd4, A: 0xff} // cream/ivory paper (land + open sea)
	ivoryLo := color.RGBA{R: 0xe4, G: 0xda, B: 0xc4, A: 0xff}
	burin := color.RGBA{R: 0x33, G: 0x2a, B: 0x1e, A: 0xff}   // fine dark engraving ink
	burinLt := color.RGBA{R: 0x6a, G: 0x5c, B: 0x46, A: 0xff} // lighter burin (thin sea rule)

	var mp mediumPalette
	mp.background = ivory
	mp.oceanDeep = ivory // sea is paper too — the RULE lines carry the water read, not a fill
	mp.oceanShelf = ivory
	mp.coast = burin
	mp.river = burin
	mp.relief = burin
	mp.reliefAlt = burinLt
	mp.shadow = ivoryLo
	// Land is blank ivory paper everywhere — relief hatching (not hue) carries the read.
	for bi := biome(0); bi < biomeCount; bi++ {
		mp.land[bi] = ivory
	}

	return worldMedium{name: "copperplate", style: styleCopperplate, palette: mp, draw: drawCopperplate}
}

// drawCopperplate paints the engraving medium. Technique: the whole sheet starts as blank
// ivory paper (a faint plate-tone mottle so it isn't sterile); the SEA is ruled with fine
// horizontal engraved lines every few rows, their darkness keyed to distance-from-coast
// (approximated by depth) so hatching is dense at the shore and thins to open paper offshore;
// the LAND stays blank except where relief sits, which gets fine diagonal CROSS-HATCHING; the
// coastline is one crisp burin line with a thin extra hatch band on the sea side; rivers are
// thin single engraved lines; and an engraved neat-line border with graticule ticks frames it.
func drawCopperplate(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	seed := uint32(0xC09E7)

	// 1) Base: blank ivory paper with a very faint plate-tone mottle (broad low-freq noise) so
	//    the sheet reads as inked paper, not a flat swatch. Land and sea are both paper here —
	//    the sea's RULE lines (step 2) are what distinguish water.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			col := p.background
			blot := valueNoise(float64(x)*0.05, float64(y)*0.05, seed^0xF0)
			col = darken(col, 0.02+0.04*blot)
			setPixel(img, x, y, col)
		}
	}

	// 2) Ruled sea-hatch: fine horizontal engraved lines over the WATER, every `spacing` rows.
	//    A line pixel's ink strength ramps with depth (deep = darker/denser feel) AND we skip the
	//    line entirely in the deepest open water on alternate bands so hatching THINS offshore and
	//    packs near the coast — the classic engraved-sea read. Broken slightly by a hash so the
	//    rule looks bitten by a burin, not laser-straight.
	const spacing = 3
	for y := 0; y < h; y++ {
		if y%spacing != 0 {
			continue
		}
		for x := 0; x < w; x++ {
			if isLandPx(m.field, x, y) {
				continue
			}
			e := m.elevAtPx(x, y)
			depth := 0.0
			if m.seaLevel > 0 {
				depth = clamp01((m.seaLevel - e) / m.seaLevel)
			}
			// Shallow (near coast, depth→0) hatches densely; deep (depth→1) thins: drop every
			// other band in the deeper half so offshore opens toward blank paper.
			if depth > 0.45 && (y/spacing)%2 == 1 {
				continue
			}
			if hashUnit(uint32(x), uint32(y), seed^0x51D) < 0.10 {
				continue // a tiny break so the rule reads engraved
			}
			// Line ink: firmer near the coast (shallow), fainter in the deep.
			ink := blend(p.reliefAlt, p.coast, clamp01(1.0-depth*0.8))
			setPixel(img, x, y, ink)
		}
	}

	// 3) Relief cross-hatching on the LAND: only where hills/mountains sit does the blank paper
	//    get fine diagonal hatch strokes (two crossing diagonal families for mountains → true
	//    cross-hatch; one family for hills → single hatch). A small radius around each anchor so
	//    the shading has area. Gate every stamp to land so hatch never bleeds into the sea.
	hatchAt := func(cx, cy, rad int, cross bool) {
		for dy := -rad; dy <= rad; dy++ {
			for dx := -rad; dx <= rad; dx++ {
				x, y := cx+dx, cy+dy
				if !isLandPx(m.field, x, y) {
					continue
				}
				// Diagonal family A: ink where (x+y) lands on a stroke line.
				a := (x + y) % 3
				bfam := (x - y) % 3
				on := a == 0
				if cross && bfam == 0 {
					on = true
				}
				if !on {
					continue
				}
				setPixel(img, x, y, p.relief)
			}
		}
	}
	for _, a := range m.reliefs {
		switch a.kind {
		case reliefMountain:
			hatchAt(a.x, a.y, 2, true) // dense cross-hatch
		case reliefHill:
			hatchAt(a.x, a.y, 1, false) // light single hatch
		}
	}

	// 4) Coastline: one crisp fine burin line at the shore, PLUS a thin one-pixel hatch band on
	//    the SEAWARD neighbour so the coast reads engraved (line + a whisker of shore-hatch).
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			if isLandPx(m.field, x-1, y) && isLandPx(m.field, x+1, y) &&
				isLandPx(m.field, x, y-1) && isLandPx(m.field, x, y+1) {
				continue
			}
			setPixel(img, x, y, p.coast) // the crisp engraved shore
			for _, nb := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+nb[0], y+nb[1]
				if !isLandPx(m.field, nx, ny) {
					setPixel(img, nx, ny, blend(p.background, p.coast, 0.55)) // thin shore-hatch band
				}
			}
		}
	}

	// 5) Rivers: thin single engraved lines over the land.
	putRiver := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, r := range m.rivers {
		for i := 0; i+1 < len(r.pts); i++ {
			a, bb := r.pts[i], r.pts[i+1]
			strokeThickLineFunc(a.x, a.y, bb.x, bb.y, 0, p.river, putRiver)
		}
	}

	// 6) Chrome: an engraved neat-line border (fine double rule) with graticule tick marks.
	drawEngravedNeatline(img, p.coast, p.reliefAlt)
}

// ---- nautical: colonial sea-chart -------------------------------------------

// nauticalMedium is a NAUTICAL sea-chart where the SEA is the subject (→ colonial_age). Aged
// tan paper; across the WATER a web of RHUMB LINES (loxodromes) radiates from 2–3 compass-node
// points and crisscrosses the sea. Near coasts, shoal STIPPLE and small sounding dot-marks
// (dots, not numbers) speckle the shallows. The LAND interior is left plain/blank tan — not
// the focus. The coastline is a firm ink line edged with tiny depth ticks. Chrome: a chart-
// style border. The geometric rhumb-line net over the sea is the signature. Anchors to aged
// tan + ink-line net HARD.
func nauticalMedium() worldMedium {
	// Fixed sea-chart identity (theme-independent by design).
	tan := color.RGBA{R: 0xd9, G: 0xc6, B: 0x9e, A: 0xff}     // aged tan chart paper (land)
	tanSea := color.RGBA{R: 0xc7, G: 0xb8, B: 0x97, A: 0xff}  // sea: a shade cooler/greyer tan
	tanDeep := color.RGBA{R: 0xb6, G: 0xa8, B: 0x8a, A: 0xff} // deeper water tan
	ink := color.RGBA{R: 0x46, G: 0x37, B: 0x22, A: 0xff}     // firm chart ink (coast, rhumb nodes)
	rhumb := color.RGBA{R: 0x7d, G: 0x5b, B: 0x38, A: 0xff}   // thin rhumb-line ink (over water)
	sound := color.RGBA{R: 0x5c, G: 0x49, B: 0x30, A: 0xff}   // sounding dot / shoal stipple

	var mp mediumPalette
	mp.background = tan
	mp.oceanDeep = tanDeep
	mp.oceanShelf = tanSea
	mp.coast = ink
	mp.river = blend(ink, tanSea, 0.35)
	mp.relief = sound    // relief is understated on a sea-chart (soundings matter, not peaks)
	mp.reliefAlt = rhumb // reuse as the rhumb-line tone
	mp.shadow = sound
	// Land is one plain tan — the interior is not the subject; only the sea net + shore read.
	for bi := biome(0); bi < biomeCount; bi++ {
		mp.land[bi] = tan
	}

	return worldMedium{name: "nautical", style: styleNautical, palette: mp, draw: drawNautical}
}

// drawNautical paints the sea-chart medium. Technique: aged tan paper (sea a shade cooler,
// faintly depth-banded) with a soft paper mottle; a web of RHUMB LINES drawn from 2–3 seeded
// compass nodes out along 16 compass bearings, each line stamped ONLY over water (it stops at
// the coast) so the sea is netted with loxodromes; shoal STIPPLE + small sounding dots speckle
// the shallow water near the coast; the coastline is a firm ink line with tiny outward depth
// ticks; the land is left plain; and a chart-style border frames it. The rhumb net is the read.
func drawNautical(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	seed := uint32(0x0CEA7)

	// 1) Base: tan paper, sea faintly cooler + depth-banded, with a soft mottle.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := m.field.at(x, y)
			if bi >= biomeCount {
				bi = biomeGrass
			}
			var col color.RGBA
			if bi == biomeDeepWater || bi == biomeShallowWater {
				e := m.elevAtPx(x, y)
				depth := 0.0
				if m.seaLevel > 0 {
					depth = clamp01((m.seaLevel - e) / m.seaLevel)
				}
				col = blend(p.oceanShelf, p.oceanDeep, depth*0.5)
			} else {
				col = p.land[bi]
			}
			blot := valueNoise(float64(x)*0.06, float64(y)*0.06, seed^0xB1)
			col = darken(col, 0.03+0.05*blot)
			setPixel(img, x, y, col)
		}
	}

	// 2) Rhumb-line net: 2–3 seeded compass NODES on the water; from each, draw thin lines along
	//    16 compass bearings clear across the frame — but stamp each line ONLY on water pixels so
	//    the net covers the sea and stops at every coast. This is the signature: a geometric web
	//    of loxodromes over the water.
	putSea := func(x, y int, c color.RGBA) {
		if x < 0 || y < 0 || x >= w || y >= h {
			return
		}
		if isLandPx(m.field, x, y) {
			return // rhumb lines live on the water only
		}
		setPixel(img, x, y, c)
	}
	nNodes := 3
	if minInt(w, h) < 60 {
		nNodes = 2 // fewer nodes on a small canvas so the net stays legible
	}
	span := w + h // long enough to cross the frame from any interior node
	for ni := 0; ni < nNodes; ni++ {
		// Place the node at a seeded position; nudge to a water pixel if it landed on land so its
		// lines actually seed from the sea (search a small spiral; fall back to the raw point).
		nx := int(hashUnit(uint32(ni), 0x0DE1, seed) * float64(w))
		ny := int(hashUnit(uint32(ni), 0x0DE2, seed) * float64(h))
		if isLandPx(m.field, nx, ny) {
			for rr := 2; rr <= 12 && isLandPx(m.field, nx, ny); rr += 2 {
				nx2 := clampInt(nx+int((hashUnit(uint32(ni), uint32(rr), seed)-0.5)*float64(rr)*4), 0, w-1)
				ny2 := clampInt(ny+int((hashUnit(uint32(rr), uint32(ni), seed)-0.5)*float64(rr)*4), 0, h-1)
				nx, ny = nx2, ny2
			}
		}
		// A node hub mark (small ink pip) so the net reads as radiating from compass roses.
		putSea(nx, ny, p.coast)
		for dir := 0; dir < 16; dir++ {
			ang := float64(dir) * (math.Pi / 8.0)
			ex := int(math.Round(math.Cos(ang) * float64(span)))
			ey := int(math.Round(math.Sin(ang) * float64(span)))
			strokeThickLineFunc(nx, ny, nx+ex, ny+ey, 0, p.reliefAlt, putSea)
		}
	}

	// 3) Shoal stipple + sounding dots: over SHALLOW water bordering land, scatter small dot
	//    marks (soundings) and a light shoal stipple by a coordinate hash — dots, never numerals.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if m.field.at(x, y) != biomeShallowWater {
				continue
			}
			nearLand := isLandPx(m.field, x-1, y) || isLandPx(m.field, x+1, y) ||
				isLandPx(m.field, x, y-1) || isLandPx(m.field, x, y+1) ||
				isLandPx(m.field, x-2, y) || isLandPx(m.field, x+2, y)
			g := hashUnit(uint32(x), uint32(y), seed^0x5A0)
			if nearLand && g > 0.86 {
				setPixel(img, x, y, p.shadow) // a sounding dot right off the shore
			} else if g > 0.955 {
				setPixel(img, x, y, blend(m.paperSample(img, x, y), p.shadow, 0.5)) // sparse shoal fleck
			}
		}
	}

	// 4) Coastline: a firm ink line at the shore with tiny outward depth ticks (one seaward pip
	//    every few shore pixels) so the coast reads as a charted, sounded shore.
	tick := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			if isLandPx(m.field, x-1, y) && isLandPx(m.field, x+1, y) &&
				isLandPx(m.field, x, y-1) && isLandPx(m.field, x, y+1) {
				continue
			}
			setPixel(img, x, y, p.coast)
			tick++
			if tick%4 == 0 {
				for _, nb := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := x+nb[0], y+nb[1]
					if !isLandPx(m.field, nx, ny) {
						setPixel(img, nx, ny, p.coast) // a small depth tick reaching into the water
						break
					}
				}
			}
		}
	}

	// 5) Rivers: thin muted ink threads on the (otherwise plain) land — a chart still notes the
	//    river mouths feeding the coast.
	putRiver := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, r := range m.rivers {
		for i := 0; i+1 < len(r.pts); i++ {
			a, bb := r.pts[i], r.pts[i+1]
			strokeThickLineFunc(a.x, a.y, bb.x, bb.y, 0, p.river, putRiver)
		}
	}

	// 6) Chrome: a chart-style border (a graduated double rule with tick divisions).
	drawChartBorder(img, p.coast, p.shadow)
}

// paperSample returns the current pixel color at (x,y) clamped on-canvas — a tiny helper so a
// medium can blend a mark toward whatever paper/mottle already sits there (keeps sparse marks
// from looking pasted on a flat swatch).
func (m *worldModel) paperSample(img *image.RGBA, x, y int) color.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return color.RGBA{A: 0xff}
	}
	return img.RGBAAt(b.Min.X+clampInt(x, 0, w-1), b.Min.Y+clampInt(y, 0, h-1))
}

// ---- ordnance: Victorian ordnance-survey sheet ------------------------------

// ordnanceMedium is a Victorian ORDNANCE-SURVEY topographic sheet (→ industrial_age). Pale
// printed off-white/buff paper with a faint pale-green lowland tint. The SIGNATURE is fine
// brown CONTOUR ISOLINES following elevation (elevation quantized into bands; each band
// boundary stroked as a thin muted-brown line) PLUS a printed thin black GRID graticule over
// the whole sheet. Thin blue rivers, small black hachure relief ticks. Chrome: a printed
// neat-line border with grid ticks. Brown contour isolines + printed grid on drab pale paper
// is the read — distinct from neon (cyan contours on BLACK) and engraving (hatch, no grid).
// Anchors to pale buff paper + brown contours + black grid HARD.
func ordnanceMedium() worldMedium {
	// Fixed ordnance-sheet identity (theme-independent by design).
	buff := color.RGBA{R: 0xea, G: 0xe6, B: 0xda, A: 0xff}     // pale printed off-white/buff paper
	lowGreen := color.RGBA{R: 0xd8, G: 0xdf, B: 0xc4, A: 0xff} // faint pale-green lowland tint
	upBuff := color.RGBA{R: 0xe3, G: 0xd8, B: 0xc4, A: 0xff}   // slightly warmer buff for uplands
	contour := color.RGBA{R: 0x9a, G: 0x6b, B: 0x40, A: 0xff}  // muted brown contour ink
	grid := color.RGBA{R: 0x3a, G: 0x36, B: 0x30, A: 0xff}     // thin printed black grid
	blue := color.RGBA{R: 0x4c, G: 0x74, B: 0x9a, A: 0xff}     // thin printed river blue

	var mp mediumPalette
	mp.background = buff
	mp.oceanDeep = color.RGBA{R: 0xc4, G: 0xd2, B: 0xd6, A: 0xff}  // pale printed water (drab blue-grey)
	mp.oceanShelf = color.RGBA{R: 0xd4, G: 0xde, B: 0xdf, A: 0xff} // paler shelf
	mp.coast = grid
	mp.river = blue
	mp.relief = grid       // black hachure ticks
	mp.reliefAlt = contour // brown contour tone
	mp.shadow = grid
	// Land tints: lowland faint green, uplands a warmer buff, so the base paper hints at
	// elevation even before the contour lines are drawn (contours + grid carry the real read).
	for bi := biome(0); bi < biomeCount; bi++ {
		mp.land[bi] = buff
	}
	mp.land[biomeGrass] = lowGreen
	mp.land[biomeForest] = blend(lowGreen, contour, 0.12)
	mp.land[biomeSand] = buff
	mp.land[biomeRock] = upBuff
	mp.land[biomeMountain] = blend(upBuff, contour, 0.10)
	mp.land[biomeSnow] = buff

	return worldMedium{name: "ordnance", style: styleOrdnance, palette: mp, draw: drawOrdnance}
}

// drawOrdnance paints the ordnance-survey medium. Technique: pale buff/green printed paper
// (very faint print mottle); fine brown CONTOUR ISOLINES where a pixel's quantized elevation
// band differs from its +x/+y neighbour's (topographic isolines over the land); a printed thin
// black GRID graticule every gridStep px across the WHOLE sheet (the km-grid); the coast as a
// firm printed line; thin printed-blue rivers; small black hachure ticks at relief; and a
// printed neat-line border with grid ticks. Brown contours + black grid on drab paper is the
// signature.
func drawOrdnance(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	seed := uint32(0x02D5A)

	// 1) Base: pale printed paper — lowland-green / upland-buff tint by biome, sea a drab pale
	//    blue-grey, with a very faint print mottle so it reads as a printed sheet.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := m.field.at(x, y)
			if bi >= biomeCount {
				bi = biomeGrass
			}
			var col color.RGBA
			if bi == biomeDeepWater || bi == biomeShallowWater {
				e := m.elevAtPx(x, y)
				depth := 0.0
				if m.seaLevel > 0 {
					depth = clamp01((m.seaLevel - e) / m.seaLevel)
				}
				col = blend(p.oceanShelf, p.oceanDeep, depth*0.6)
			} else {
				col = p.land[bi]
			}
			blot := valueNoise(float64(x)*0.09, float64(y)*0.09, seed^0xF0)
			col = darken(col, 0.015+0.03*blot)
			setPixel(img, x, y, col)
		}
	}

	// 2) Contour isolines: quantize land elevation into bands; a land pixel whose band differs
	//    from its +x or +y neighbour sits on a band boundary → stroke it muted brown. This is the
	//    topographic-contour signature.
	const bands = 7
	bandOf := func(x, y int) int {
		if !isLandPx(m.field, x, y) {
			return -1
		}
		e := m.elevAtPx(x, y)
		t := (e - m.seaLevel) / (1 - m.seaLevel)
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		return int(t * float64(bands))
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bnd := bandOf(x, y)
			if bnd < 0 {
				continue
			}
			if bandOf(x+1, y) != bnd || bandOf(x, y+1) != bnd {
				cur := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
				setPixel(img, x, y, blend(cur, p.reliefAlt, 0.85)) // fine brown contour
			}
		}
	}

	// 3) Printed grid graticule over the WHOLE sheet: thin black lines every gridStep px (the
	//    km-grid). Kept light (blended) so it's a printed reference, not a heavy overlay.
	gridStep := 10
	if minInt(w, h) < 60 {
		gridStep = 6
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x%gridStep == 0 || y%gridStep == 0 {
				cur := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
				setPixel(img, x, y, blend(cur, p.shadow, 0.28))
			}
		}
	}

	// 4) Coastline: a firm printed line at the shore (black), so the land/sea edge reads crisp
	//    over the grid.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			if isLandPx(m.field, x-1, y) && isLandPx(m.field, x+1, y) &&
				isLandPx(m.field, x, y-1) && isLandPx(m.field, x, y+1) {
				continue
			}
			setPixel(img, x, y, p.coast)
		}
	}

	// 5) Rivers: thin printed blue, gated to land.
	putRiver := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, r := range m.rivers {
		for i := 0; i+1 < len(r.pts); i++ {
			a, bb := r.pts[i], r.pts[i+1]
			strokeThickLineFunc(a.x, a.y, bb.x, bb.y, 0, p.river, putRiver)
		}
	}

	// 6) Relief: small black hachure ticks — a couple of short downhill strokes at each peak/hill
	//    (the OS hachure convention), understated over the contours. Gate to land.
	putR := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, a := range m.reliefs {
		switch a.kind {
		case reliefMountain:
			putR(a.x, a.y, p.relief)
			putR(a.x-1, a.y+1, p.relief)
			putR(a.x+1, a.y+1, p.relief)
		case reliefHill:
			putR(a.x, a.y, p.relief)
		}
	}

	// 7) Chrome: a printed neat-line border with grid ticks (matches the sheet's printed grid).
	drawOrdnanceBorder(img, p.shadow, gridStep)
}

// ---- lithograph: Victorian hand-tinted lithograph ---------------------------

// lithographMedium is a hand-TINTED LITHOGRAPH (→ victorian_age). An aged cream base carries a
// fine black litho OUTLINE (coastline + relief), and the SIGNATURE is soft PASTEL tint washes
// filling regions by biome — pale desaturated pink / green / yellow / blue — slightly
// mottled/uneven as if hand-colored, bleeding a hair past the outlines. Small shaded hill
// glyphs, thin blue rivers. Chrome: a decorative Victorian double-rule border with small corner
// flourishes. Soft multi-color pastel washes over a fine outline is the read — distinct from
// parchment's monochrome sepia via the color. Anchors to cream + pastel biome washes + black
// outline HARD.
func lithographMedium() worldMedium {
	// Fixed lithograph identity (theme-independent by design). PASTEL biome washes are the point.
	cream := color.RGBA{R: 0xf0, G: 0xe9, B: 0xd8, A: 0xff}    // aged cream base
	litho := color.RGBA{R: 0x2e, G: 0x28, B: 0x22, A: 0xff}    // fine black litho outline
	seaBlue := color.RGBA{R: 0xb8, G: 0xcd, B: 0xd6, A: 0xff}  // pale desaturated sea blue
	seaBlueD := color.RGBA{R: 0xa2, G: 0xba, B: 0xc6, A: 0xff} // slightly deeper sea wash

	var mp mediumPalette
	mp.background = cream
	mp.oceanDeep = seaBlueD
	mp.oceanShelf = seaBlue
	mp.coast = litho
	mp.river = color.RGBA{R: 0x6f, G: 0x92, B: 0xac, A: 0xff} // soft blue river
	mp.relief = litho
	mp.reliefAlt = color.RGBA{R: 0xc9, G: 0xa9, B: 0x86, A: 0xff} // soft tan hill-shade glyph
	mp.shadow = blend(litho, cream, 0.4)
	// Pastel biome washes — pale, desaturated, hand-tinted tones. Everything starts near cream
	// then leans a hair toward its pastel so the color reads soft, never saturated.
	for bi := biome(0); bi < biomeCount; bi++ {
		mp.land[bi] = cream
	}
	mp.land[biomeGrass] = color.RGBA{R: 0xcf, G: 0xd9, B: 0xb0, A: 0xff}    // pale pastel green
	mp.land[biomeForest] = color.RGBA{R: 0xb4, G: 0xc7, B: 0x9c, A: 0xff}   // deeper pastel green
	mp.land[biomeSand] = color.RGBA{R: 0xe8, G: 0xdb, B: 0xac, A: 0xff}     // pale pastel yellow
	mp.land[biomeRock] = color.RGBA{R: 0xe0, G: 0xc6, B: 0xc0, A: 0xff}     // pale pastel pink
	mp.land[biomeMountain] = color.RGBA{R: 0xd6, G: 0xba, B: 0xb6, A: 0xff} // deeper pastel pink
	mp.land[biomeSnow] = color.RGBA{R: 0xef, G: 0xec, B: 0xe6, A: 0xff}     // near-white

	return worldMedium{name: "lithograph", style: styleLithograph, palette: mp, draw: drawLithograph}
}

// drawLithograph paints the hand-tinted-lithograph medium. Technique: the base is laid as soft
// PASTEL biome washes (sea a pale blue), each wash slightly mottled by low-freq noise so the
// color looks hand-brushed; the wash is then BLED a hair past the true biome edge — a land
// pixel adjacent to a differently-tinted neighbour picks up a touch of that neighbour's wash —
// so tints bleed like hand-coloring over the plate; a fine black litho OUTLINE strokes the
// coastline; thin soft-blue rivers; small shaded hill glyphs; and a decorative Victorian
// double-rule border with corner flourishes frames it. The pastel washes are the signature.
func drawLithograph(img *image.RGBA, m *worldModel, med worldMedium) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || m == nil {
		return
	}
	p := med.palette
	seed := uint32(0x1170A)

	// biomeWash returns the pastel wash tone for a pixel's biome (sea → pale blue, depth-banded).
	biomeWash := func(x, y int) color.RGBA {
		bi := m.field.at(x, y)
		if bi >= biomeCount {
			bi = biomeGrass
		}
		if bi == biomeDeepWater || bi == biomeShallowWater {
			e := m.elevAtPx(x, y)
			depth := 0.0
			if m.seaLevel > 0 {
				depth = clamp01((m.seaLevel - e) / m.seaLevel)
			}
			return blend(p.oceanShelf, p.oceanDeep, depth*0.7)
		}
		return p.land[bi]
	}

	// 1) Base: pastel washes, softly mottled so each color reads hand-brushed (uneven density).
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			col := biomeWash(x, y)
			// Hand-brushed unevenness: low-freq noise lightens/darkens the wash a hair.
			n := valueNoise(float64(x)*0.11, float64(y)*0.11, seed^0x77)
			if n > 0.5 {
				col = brighten(col, (n-0.5)*0.14)
			} else {
				col = darken(col, (0.5-n)*0.12)
			}
			setPixel(img, x, y, col)
		}
	}

	// 2) Bleed: hand-coloring runs a hair past the outline. A LAND pixel that borders a
	//    differently-tinted land neighbour picks up a soft touch of that neighbour's wash, so the
	//    color edges look bled rather than mechanically clipped. Gated to land↔land so the sea
	//    wash and the coastline stay crisp (the outline in step 3 re-asserts the shore).
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			here := m.field.at(x, y)
			for _, nb := range [4][2]int{{1, 0}, {0, 1}} {
				nx, ny := x+nb[0], y+nb[1]
				if !isLandPx(m.field, nx, ny) {
					continue
				}
				there := m.field.at(nx, ny)
				if there == here {
					continue
				}
				// Bleed a little of each side's wash across the shared edge.
				cur := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
				setPixel(img, x, y, blend(cur, p.land[there], 0.22))
				curN := img.RGBAAt(b.Min.X+nx, b.Min.Y+ny)
				setPixel(img, nx, ny, blend(curN, p.land[here], 0.22))
			}
		}
	}

	// 3) Coastline: a fine black litho outline at the shore (stamped just the shore pixel so the
	//    outline stays fine, re-asserting a crisp edge over the bled washes).
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !isLandPx(m.field, x, y) {
				continue
			}
			if isLandPx(m.field, x-1, y) && isLandPx(m.field, x+1, y) &&
				isLandPx(m.field, x, y-1) && isLandPx(m.field, x, y+1) {
				continue
			}
			setPixel(img, x, y, p.coast)
		}
	}

	// 4) Rivers: thin soft-blue threads over the land.
	putRiver := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, r := range m.rivers {
		for i := 0; i+1 < len(r.pts); i++ {
			a, bb := r.pts[i], r.pts[i+1]
			strokeThickLineFunc(a.x, a.y, bb.x, bb.y, 0, p.river, putRiver)
		}
	}

	// 5) Relief: small shaded hill glyphs — a fine black outline peak with a soft tan hill-shade
	//    body, the hand-tinted-map vocabulary. Hills a small tan mound. Gate to land.
	putR := func(x, y int, c color.RGBA) {
		if isLandPx(m.field, x, y) {
			setPixel(img, x, y, c)
		}
	}
	for _, a := range m.reliefs {
		switch a.kind {
		case reliefMountain:
			putR(a.x, a.y-1, p.relief) // black outline apex
			putR(a.x-1, a.y, p.relief) // outline flanks
			putR(a.x+1, a.y, p.relief)
			putR(a.x, a.y, p.reliefAlt) // tan shaded body
		case reliefHill:
			putR(a.x, a.y, p.reliefAlt) // a small tan mound
			putR(a.x, a.y+1, p.shadow)
		}
	}

	// 6) Chrome: a decorative Victorian double-rule border with small corner flourishes.
	drawVictorianBorder(img, p.coast, p.reliefAlt)
}

// ---- chrome helpers ---------------------------------------------------------

// drawEngravedNeatline draws the copperplate medium's chrome: a fine engraved neat-line border
// (a double rule — an outer fine line + an inner fine line a few px in) with small graticule
// TICK marks stepping along all four margins, in the engraving ink. Precise, thin, no ornament
// — the plate mark of an engraved map. `ink` is the fine line, `tickCol` the lighter tick.
func drawEngravedNeatline(img *image.RGBA, ink, tickCol color.RGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 8 || h < 8 {
		return
	}
	rect := func(inset int, col color.RGBA) {
		for x := inset; x < w-inset; x++ {
			setPixel(img, x, inset, col)
			setPixel(img, x, h-1-inset, col)
		}
		for y := inset; y < h-inset; y++ {
			setPixel(img, inset, y, col)
			setPixel(img, w-1-inset, y, col)
		}
	}
	rect(1, ink) // outer fine rule
	rect(4, ink) // inner fine rule
	// Graticule ticks: short marks every `step` px in the 3px channel between the two rules,
	// along all four margins — the engraved minute-marks of a neat-line.
	step := 8
	if minInt(w, h) < 60 {
		step = 5
	}
	for x := 4; x < w-4; x += step {
		setPixel(img, x, 2, tickCol)
		setPixel(img, x, 3, tickCol)
		setPixel(img, x, h-3, tickCol)
		setPixel(img, x, h-4, tickCol)
	}
	for y := 4; y < h-4; y += step {
		setPixel(img, 2, y, tickCol)
		setPixel(img, 3, y, tickCol)
		setPixel(img, w-3, y, tickCol)
		setPixel(img, w-4, y, tickCol)
	}
}

// drawChartBorder draws the nautical medium's chrome: a chart-style GRADUATED border — an outer
// ink rule and an inner ink rule, with the channel between them divided into alternating
// filled/blank ticks (the graduated-scale margin of a sea-chart). `ink` the rules + fill,
// `light` the blank-tick backing.
func drawChartBorder(img *image.RGBA, ink, light color.RGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 10 || h < 10 {
		return
	}
	rect := func(inset int, col color.RGBA) {
		for x := inset; x < w-inset; x++ {
			setPixel(img, x, inset, col)
			setPixel(img, x, h-1-inset, col)
		}
		for y := inset; y < h-inset; y++ {
			setPixel(img, inset, y, col)
			setPixel(img, w-1-inset, y, col)
		}
	}
	rect(0, ink) // outer rule
	rect(3, ink) // inner rule
	// Graduated ticks in the 2px channel (rows/cols 1..2): alternate ink/light every 4px so the
	// margin reads as a graduated scale bar around the whole chart.
	const grad = 4
	fill := func(x, y int, on bool) {
		c := light
		if on {
			c = ink
		}
		setPixel(img, x, y, c)
	}
	for x := 1; x < w-1; x++ {
		on := (x/grad)%2 == 0
		fill(x, 1, on)
		fill(x, h-2, on)
	}
	for y := 1; y < h-1; y++ {
		on := (y/grad)%2 == 0
		fill(1, y, on)
		fill(w-2, y, on)
	}
}

// drawOrdnanceBorder draws the ordnance medium's chrome: a printed neat-line border (a firm
// single black rule at the edge) with small grid TICKS stepping inward at the same interval as
// the sheet's printed grid, so the frame reads as the trimmed margin of a printed OS sheet.
// `ink` the printed line/ticks, `gridStep` the grid interval to align the ticks to.
func drawOrdnanceBorder(img *image.RGBA, ink color.RGBA, gridStep int) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 8 || h < 8 {
		return
	}
	if gridStep < 2 {
		gridStep = 2
	}
	// Firm printed neat-line at the very edge.
	for x := 0; x < w; x++ {
		setPixel(img, x, 0, ink)
		setPixel(img, x, h-1, ink)
	}
	for y := 0; y < h; y++ {
		setPixel(img, 0, y, ink)
		setPixel(img, w-1, y, ink)
	}
	// Grid ticks: a short inward mark wherever the printed grid meets the margin, so the frame
	// and the interior grid line up like a real sheet.
	for x := 0; x < w; x += gridStep {
		setPixel(img, x, 1, ink)
		setPixel(img, x, 2, ink)
		setPixel(img, x, h-2, ink)
		setPixel(img, x, h-3, ink)
	}
	for y := 0; y < h; y += gridStep {
		setPixel(img, 1, y, ink)
		setPixel(img, 2, y, ink)
		setPixel(img, w-2, y, ink)
		setPixel(img, w-3, y, ink)
	}
}

// drawVictorianBorder draws the lithograph medium's chrome: a decorative Victorian DOUBLE-RULE
// border (a bold outer rule + a fine inner rule) with a small ornamental FLOURISH at each
// corner — a little stepped bracket of the accent tone — so the frame reads as a decorative
// tinted-plate border, distinct from parchment's plainer double frame. `ink` the rules,
// `accent` the corner flourish tone.
func drawVictorianBorder(img *image.RGBA, ink, accent color.RGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 10 || h < 10 {
		return
	}
	rect := func(inset int, col color.RGBA) {
		for x := inset; x < w-inset; x++ {
			setPixel(img, x, inset, col)
			setPixel(img, x, h-1-inset, col)
		}
		for y := inset; y < h-inset; y++ {
			setPixel(img, inset, y, col)
			setPixel(img, w-1-inset, y, col)
		}
	}
	rect(0, ink) // bold outer rule
	rect(1, ink) // second row → a bold 2px outer band
	rect(4, ink) // fine inner rule
	// Corner flourishes: a small stepped bracket in the accent tone, hugging each corner between
	// the two rules — a hint of Victorian ornament.
	flourish := func(cx, cy, sx, sy int) {
		pts := [][2]int{{2, 2}, {3, 2}, {2, 3}, {3, 3}, {5, 2}, {2, 5}}
		for _, d := range pts {
			setPixel(img, cx+sx*d[0], cy+sy*d[1], accent)
		}
	}
	flourish(0, 0, 1, 1)
	flourish(w-1, 0, -1, 1)
	flourish(0, h-1, 1, -1)
	flourish(w-1, h-1, -1, -1)
}

// drawRoughBorder scratches a broken dark rim `thick` px in from every edge, gated by a
// hash so ~70% of border pixels are stamped — a hand-drawn stone frame, not a clean rule.
func drawRoughBorder(img *image.RGBA, col color.RGBA, thick int, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || thick <= 0 {
		return
	}
	on := func(x, y int) {
		if hashUnit(uint32(x), uint32(y), seed) < 0.30 {
			return // broken
		}
		setPixel(img, x, y, col)
	}
	for t := 0; t < thick; t++ {
		for x := 0; x < w; x++ {
			on(x, t)
			on(x, h-1-t)
		}
		for y := 0; y < h; y++ {
			on(t, y)
			on(w-1-t, y)
		}
	}
}

// drawParchmentFrame draws an ornate DOUBLE ink frame: an outer solid rule and an inner
// rule a few px in, with little corner ticks — the border of an illuminated map.
func drawParchmentFrame(img *image.RGBA, ink, shadow color.RGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 8 || h < 8 {
		return
	}
	rect := func(inset int, col color.RGBA) {
		for x := inset; x < w-inset; x++ {
			setPixel(img, x, inset, col)
			setPixel(img, x, h-1-inset, col)
		}
		for y := inset; y < h-inset; y++ {
			setPixel(img, inset, y, col)
			setPixel(img, w-1-inset, y, col)
		}
	}
	rect(1, ink)    // outer rule
	rect(4, shadow) // inner rule (lighter)
	// Corner ticks: short diagonals joining the two rules → an ornate corner.
	for _, c := range [4][2]int{{1, 1}, {w - 2, 1}, {1, h - 2}, {w - 2, h - 2}} {
		sx := 1
		if c[0] > w/2 {
			sx = -1
		}
		sy := 1
		if c[1] > h/2 {
			sy = -1
		}
		for k := 0; k <= 3; k++ {
			setPixel(img, c[0]+sx*k, c[1]+sy*k, ink)
		}
	}
}

// drawCompassRose draws a small 8-point compass rose in the bottom-right, inset from the
// frame: a filled diamond of ink rays over a faint disc, with a bright center. The classic
// antique-map orientation mark.
func drawCompassRose(img *image.RGBA, ink, accent, paper color.RGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 40 || h < 40 {
		return
	}
	r := clampInt(minInt(w, h)/9, 5, 16)
	cx := w - r - 8
	cy := h - r - 8
	// Faint backing disc so the rose reads on a busy shore.
	disc := blend(paper, ink, 0.10)
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				setPixel(img, cx+dx, cy+dy, disc)
			}
		}
	}
	// Eight rays: N/E/S/W long, diagonals shorter — drawn as tapering ink lines.
	long := r
	diag := r * 7 / 10
	ray := func(ex, ey, length int) {
		for k := 0; k <= length; k++ {
			// Taper: near the center thicker (2px), tips thin.
			x := cx + ex*k
			y := cy + ey*k
			setPixel(img, x, y, ink)
			if k < length/2 {
				setPixel(img, x+(ey), y+(ex), blend(ink, paper, 0.5)) // faint flank for a diamond body
			}
		}
	}
	ray(0, -1, long) // N
	ray(0, 1, long)  // S
	ray(1, 0, long)  // E
	ray(-1, 0, long) // W
	ray(1, -1, diag)
	ray(-1, -1, diag)
	ray(1, 1, diag)
	ray(-1, 1, diag)
	// Bright center pip.
	setPixel(img, cx, cy, accent)
}

// drawClouds lays a few faint white cloud wisps over the frame: broad low-freq noise
// thresholded high and feathered so only sparse soft patches show — the satellite medium's
// "live orbital shot" cue. Clouds sit over sea and land alike (weather ignores the coast).
func drawClouds(img *image.RGBA, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return
	}
	white := color.RGBA{R: 0xf4, G: 0xf6, B: 0xf8, A: 0xff}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n := valueNoise(float64(x)*0.05, float64(y)*0.05, seed)
			if n <= 0.66 {
				continue
			}
			// Feather: strength ramps from 0 at the threshold to a soft cap.
			t := (n - 0.66) / 0.34 * 0.35
			cur := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
			setPixel(img, x, y, blend(cur, white, clamp01(t)))
		}
	}
}

// drawHUDBrackets draws bright cyan corner brackets (an L in each corner) — a targeting-
// reticle / HUD read for the neon medium, lighter-touch than a full frame.
func drawHUDBrackets(img *image.RGBA, cyan color.RGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 12 || h < 12 {
		return
	}
	arm := clampInt(minInt(w, h)/8, 4, 20)
	inset := 2
	// Each corner: two arms forming an L.
	corner := func(cx, cy, sx, sy int) {
		for k := 0; k < arm; k++ {
			setPixel(img, cx+sx*k, cy, cyan)
			setPixel(img, cx, cy+sy*k, cyan)
		}
	}
	corner(inset, inset, 1, 1)
	corner(w-1-inset, inset, -1, 1)
	corner(inset, h-1-inset, 1, -1)
	corner(w-1-inset, h-1-inset, -1, -1)
}

// drawRockEdgeBorder rims the frame with an irregular chipped rock edge for the petroglyph
// medium: a variable-thickness dark groove whose depth wanders by a per-position hash, with a
// brighter chipped fleck scattered along it — the ragged margin of a rock face, not a ruled
// frame. `groove` is the recess shadow, `chip` the lighter chipped stone fleck.
func drawRockEdgeBorder(img *image.RGBA, groove, chip color.RGBA, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return
	}
	// For each border pixel, the edge "bites" in by a hash-driven depth (0..2 px); a fraction
	// of those bites get a bright chip fleck so the rim reads chiseled.
	stamp := func(x, y int, depthAxis, sign int) {
		// depthAxis 0 = vertical edge (bite along x), 1 = horizontal edge (bite along y).
		d := int(hashUnit(uint32(x), uint32(y), seed) * 3.0) // 0..2
		for k := 0; k <= d; k++ {
			var px, py int
			if depthAxis == 0 {
				px, py = x+sign*k, y
			} else {
				px, py = x, y+sign*k
			}
			col := groove
			if hashUnit(uint32(px), uint32(py), seed^0xFECC) > 0.78 {
				col = chip // a bright chipped fleck
			}
			setPixel(img, px, py, col)
		}
	}
	for x := 0; x < w; x++ {
		stamp(x, 0, 1, 1)    // top edge bites downward
		stamp(x, h-1, 1, -1) // bottom edge bites upward
	}
	for y := 0; y < h; y++ {
		stamp(0, y, 0, 1)    // left edge bites rightward
		stamp(w-1, y, 0, -1) // right edge bites leftward
	}
}

// drawClayTabletEdge draws a rounded clay-tablet rim (a beveled edge: a dark shadow rule with
// a raised highlight just inside it, corners rounded off) plus ONE broken-off corner chip — a
// wedge of the tablet missing at a seeded corner, filled with the edge shadow so it reads as a
// clean break. `shadow` is the edge recess, `lip` the sunlit raised bevel.
func drawClayTabletEdge(img *image.RGBA, shadow, lip color.RGBA, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 6 || h < 6 {
		return
	}
	// Beveled rim: an outer shadow rule + an inner raised lip, with the four corners rounded
	// (skip the exact corner pixels so the tablet reads with soft edges, not sharp square ones).
	rounded := func(x, y int) bool {
		// Round a 2px radius at each corner.
		return (x <= 1 && y <= 1) || (x >= w-2 && y <= 1) ||
			(x <= 1 && y >= h-2) || (x >= w-2 && y >= h-2)
	}
	for x := 0; x < w; x++ {
		if !rounded(x, 0) {
			setPixel(img, x, 0, shadow)
			setPixel(img, x, 1, lip)
		}
		if !rounded(x, h-1) {
			setPixel(img, x, h-1, shadow)
			setPixel(img, x, h-2, lip)
		}
	}
	for y := 0; y < h; y++ {
		if !rounded(0, y) {
			setPixel(img, 0, y, shadow)
			setPixel(img, 1, y, lip)
		}
		if !rounded(w-1, y) {
			setPixel(img, w-1, y, shadow)
			setPixel(img, w-2, y, lip)
		}
	}
	// One broken-off corner chip: pick a corner by seed, gouge a small triangular wedge out of
	// it (filled with the edge shadow so it reads as a fresh break in the clay).
	corner := int(hashUnit(7, 0x0FFC, seed) * 4.0) // 0..3
	chip := clampInt(minInt(w, h)/6, 3, 10)
	sx, sy := 0, 0
	switch corner {
	case 1:
		sx = w - 1
	case 2:
		sy = h - 1
	case 3:
		sx, sy = w-1, h-1
	}
	dirx := 1
	if sx > 0 {
		dirx = -1
	}
	diry := 1
	if sy > 0 {
		diry = -1
	}
	for i := 0; i < chip; i++ {
		for j := 0; j < chip-i; j++ {
			setPixel(img, sx+dirx*i, sy+diry*j, shadow)
		}
	}
}

// drawScorchedEdges lays the hide medium's SIGNATURE burnt vignette: every pixel is darkened
// toward soot-black by a strength that grows from 0 in the interior to a heavy char at the
// frame edge, weighted so the CORNERS scorch hardest (edge distance combines both axes). A
// hash flicker roughens the burn so the char reads as an organic scorch, not a clean gradient.
func drawScorchedEdges(img *image.RGBA, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return
	}
	char := color.RGBA{R: 0x12, G: 0x0b, B: 0x07, A: 0xff}
	// The burn ramps in over this fraction of the half-span from each edge.
	marginX := float64(w) * 0.32
	marginY := float64(h) * 0.32
	if marginX < 1 {
		marginX = 1
	}
	if marginY < 1 {
		marginY = 1
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Distance INTO the frame from the nearest vertical/horizontal edge, normalized so
			// 0 = at an edge, 1 = at/inside the margin. Multiplying the two axis terms makes the
			// corners (small on BOTH axes) darken hardest.
			dxE := math.Min(float64(x), float64(w-1-x)) / marginX
			dyE := math.Min(float64(y), float64(h-1-y)) / marginY
			ex := 1 - clamp01(dxE)
			ey := 1 - clamp01(dyE)
			// Combine: heavy where either edge is close, heaviest in the corners.
			burn := 1 - (1-ex)*(1-ey)
			if burn <= 0.001 {
				continue
			}
			// Hash flicker so the scorch line is ragged.
			burn *= 0.55 + 0.45*hashUnit(uint32(x), uint32(y), seed)
			burn = clamp01(burn) * 0.9
			cur := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
			setPixel(img, x, y, blend(cur, char, burn))
		}
	}
}

// drawGreekKeyBorder draws a classical Greek-key / meander border around the frame for the
// mosaic medium: a two-rule band (outer + inner) with a repeating interlocking-key motif
// stamped along the top and bottom edges — the signature geometric mosaic frame. `tile` is
// the light tesserae tone (the key strokes), `grout` the dark backing rule.
func drawGreekKeyBorder(img *image.RGBA, tile, grout color.RGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 16 || h < 16 {
		// Too small for a meander — just lay a plain double rule so there's still a frame.
		for x := 0; x < w; x++ {
			setPixel(img, x, 0, grout)
			setPixel(img, x, h-1, grout)
		}
		for y := 0; y < h; y++ {
			setPixel(img, 0, y, grout)
			setPixel(img, w-1, y, grout)
		}
		return
	}
	// Backing band: a solid grout rule at the very edge, framing the light key motif inside it.
	rule := func(inset int, col color.RGBA) {
		for x := inset; x < w-inset; x++ {
			setPixel(img, x, inset, col)
			setPixel(img, x, h-1-inset, col)
		}
		for y := inset; y < h-inset; y++ {
			setPixel(img, inset, y, col)
			setPixel(img, w-1-inset, y, col)
		}
	}
	rule(0, grout)
	rule(4, grout)
	// Greek-key motif in the 3px channel between the two rules (rows 1..3 top / bottom). Each
	// repeat is a small interlocking hook; stamped in the light tesserae tone.
	const period = 6
	drawKeyUnit := func(ox, oy, sy int) {
		// A compact meander hook within a period×3 box, mirrored by sy (+1 top, -1 bottom).
		set := func(dx, dy int) { setPixel(img, ox+dx, oy+sy*dy, tile) }
		set(0, 0)
		set(1, 0)
		set(2, 0)
		set(2, 1)
		set(2, 2)
		set(1, 2)
		set(0, 2)
		set(4, 1) // the interlock tooth reaching into the next cell
	}
	for x := 5; x+period < w-5; x += period {
		drawKeyUnit(x, 1, 1)    // top band, growing downward
		drawKeyUnit(x, h-4, -1) // bottom band, mirrored upward
	}
}
