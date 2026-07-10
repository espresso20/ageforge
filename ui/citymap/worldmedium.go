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
	styleAtlas     worldMediumStyle = iota // neutral default: crisp coast, depth-banded sea
	styleCharcoal                          // near-mono, grainy, jittered chalk coastline
	styleParchment                         // vellum, wavy ink coast, frame + compass + cartouche
	styleSatellite                         // photographic, natural land/sea transition, no outline
	styleNeon                              // black + neon wireframe, holographic grid overlay
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

// ---- chrome helpers ---------------------------------------------------------

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
