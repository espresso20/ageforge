package ui

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/theme"
)

// banner.go draws the README banner (docs/banner.png, made by
// cmd/bannerpng): the town page's wordmark, doubled, in a band of night sky
// with the line under it, on the Forge theme. It is drawn by the menu's own
// code into an mGrid and only then put on pixels, so the banner and the menu
// cannot drift apart. The clock stands at nought and nothing strikes, so the
// picture is the same on every run.

// The banner's grid: four rows of sky, a blank row, the wordmark (seven
// rows), a blank row, the line, a blank row and four rows of sky.
const (
	bannerW      = 140
	bannerH      = 19
	bannerSky    = 4
	bannerWordY  = bannerSky + 1
	bannerTagY   = bannerWordY + menuWordH + 1
	bannerLowSky = bannerTagY + 2
)

// bannerGrid is the banner as cells, on the palette it was drawn in.
func bannerGrid(ages int) (*mGrid, *menuPalette) {
	pal := newMenuPalette(theme.Forge)
	g := newMGrid(bannerW, bannerH)
	rnd := menuRand{s: 0x9e3779b97f4a7c15}
	high := newMenuStars(&rnd, bannerW, bannerSky)
	low := newMenuStars(&rnd, bannerW, bannerSky)
	for i := range low {
		low[i].y += bannerLowSky
	}
	drawMenuStars(g, pal, high, 0)
	drawMenuStars(g, pal, low, 0)
	drawWordmark(g, pal, centred(bannerW, menuWordW*2), bannerWordY, 2, 0, 0, 0.5, false)
	tag := menuTagline(ages)
	g.textOn(centred(bannerW, runeLen(tag)), bannerTagY, tag, pal.ink, pal.bg, false)
	return g, pal
}

// BannerImage draws the README banner, its tagline counting ages, with each cell cellW by cellH pixels.
// A block is a solid block of its colour, a star a dot, and text is set in
// bannerFont, centred over the cells it takes.
func BannerImage(ages, cellW, cellH int) *image.RGBA {
	g, pal := bannerGrid(ages)
	img := image.NewRGBA(image.Rect(0, 0, g.w*cellW, g.h*cellH))
	dot := max(1, min(cellW/5, cellH/8))
	box := func(x, y, w, h int, c tcell.Color) {
		draw.Draw(img, image.Rect(x, y, x+w, y+h), image.NewUniform(rgba(c)), image.Point{}, draw.Src)
	}
	for y := 0; y < g.h; y++ {
		first, last := -1, -1
		for x := 0; x < g.w; x++ {
			c := g.c[y*g.w+x]
			bg := pal.bg
			if c.hasBg {
				bg = c.bg
			}
			px, py := x*cellW, y*cellH
			box(px, py, cellW, cellH, bg)
			switch c.r {
			case ' ':
			case '█':
				box(px, py, cellW, cellH, c.fg)
			case '·', '∙':
				s := dot
				if c.r == '∙' {
					s = 2 * dot
				}
				box(px+(cellW-s)/2, py+(cellH-s)/2, s, s, c.fg)
			default:
				if first < 0 {
					first = x
				}
				last = x
			}
		}
		if first < 0 {
			continue
		}
		// The text on the row, set as one run so its letters keep their
		// spacing whatever the cell's width.
		run := g.c[y*g.w+first : y*g.w+last+1]
		adv := 6 * dot
		x0 := (first+last+1)*cellW/2 - len(run)*adv/2 + dot/2
		y0 := y*cellH + (cellH-8*dot)/2
		for i, c := range run {
			rows, ok := bannerFont[c.r]
			if !ok {
				continue
			}
			for ry, row := range rows {
				for rx := range row {
					if row[rx] == '#' {
						box(x0+i*adv+rx*dot, y0+ry*dot, dot, dot, c.fg)
					}
				}
			}
		}
	}
	return img
}

// IconImage draws a square icon (1024x1024) with AGE and FORGE centred on
// the navy Forge background, with the menu's fire gradient and stars in
// the margins. It is drawn by the menu's own code into an mGrid and only
// then put on pixels, so the icon will never drift apart from the menu.
// The clock stands at nought and nothing strikes, so the picture is the
// same on every run.
func IconImage(size int) *image.RGBA {
	g, pal := iconGrid()
	img := image.NewRGBA(image.Rect(0, 0, g.w*size/64, g.h*size/64))
	cellW, cellH := size/64, size/64
	dot := max(1, min(cellW/5, cellH/8))
	box := func(x, y, w, h int, c tcell.Color) {
		draw.Draw(img, image.Rect(x, y, x+w, y+h), image.NewUniform(rgba(c)), image.Point{}, draw.Src)
	}
	for y := 0; y < g.h; y++ {
		first, last := -1, -1
		for x := 0; x < g.w; x++ {
			c := g.c[y*g.w+x]
			bg := pal.bg
			if c.hasBg {
				bg = c.bg
			}
			px, py := x*cellW, y*cellH
			box(px, py, cellW, cellH, bg)
			switch c.r {
			case ' ':
			case '█':
				box(px, py, cellW, cellH, c.fg)
			case '·', '∙':
				s := dot
				if c.r == '∙' {
					s = 2 * dot
				}
				box(px+(cellW-s)/2, py+(cellH-s)/2, s, s, c.fg)
			default:
				if first < 0 {
					first = x
				}
				last = x
			}
		}
		if first < 0 {
			continue
		}
		// The text on the row, set as one run so its letters keep their
		// spacing whatever the cell's width.
		run := g.c[y*g.w+first : y*g.w+last+1]
		adv := 6 * dot
		x0 := (first+last+1)*cellW/2 - len(run)*adv/2 + dot/2
		y0 := y*cellH + (cellH-8*dot)/2
		for i, c := range run {
			rows, ok := bannerFont[c.r]
			if !ok {
				continue
			}
			for ry, row := range rows {
				for rx := range row {
					if row[rx] == '#' {
						box(x0+i*adv+rx*dot, y0+ry*dot, dot, dot, c.fg)
					}
				}
			}
		}
	}
	return img
}

// iconGrid is the icon as cells, on the palette it was drawn in: AGE and
// FORGE on the navy Forge background with the fire gradient, stars in the
// margins.
func iconGrid() (*mGrid, *menuPalette) {
	pal := newMenuPalette(theme.Forge)
	const (
		iconW  = 64
		iconH  = 64
		margin = 2
		lineH  = 14 // menuWordH * 2
		gap    = 1
		totalH = lineH + gap + lineH
		topY   = (iconH - totalH) / 2
	)
	g := newMGrid(iconW, iconH)
	rnd := menuRand{s: 0x9e3779b97f4a7c15}
	stars := newMenuStars(&rnd, iconW, margin)
	bottomStars := newMenuStars(&rnd, iconW, margin)
	for i := range bottomStars {
		bottomStars[i].y += iconH - margin
	}
	drawMenuStars(g, pal, stars, 0)
	drawMenuStars(g, pal, bottomStars, 0)

	// Draw top line (AGE) on the first line
	ageGlyphs := [3]rune{'A', 'G', 'E'}
	age := buildWord(ageGlyphs[:])
	ageW := len(age[0])
	ageSx := 2 // 2x scale
	ageX0 := (iconW - ageW*ageSx) / 2
	drawPixelWordmark(g, pal, age, ageX0, topY, ageSx, iconW)

	// Draw bottom line (FORGE) on the second line
	forgeGlyphs := [5]rune{'F', 'O', 'R', 'G', 'E'}
	forge := buildWord(forgeGlyphs[:])
	forgeW := len(forge[0])
	forgeSx := 2 // 2x scale
	forgeX0 := (iconW - forgeW*forgeSx) / 2
	drawPixelWordmark(g, pal, forge, forgeX0, topY+lineH+gap, forgeSx, iconW)

	return g, pal
}

// buildWord constructs the rows of a word from menuFont glyphs.
func buildWord(glyphs []rune) [menuWordH]string {
	var rows [menuWordH]string
	for i, r := range glyphs {
		if i > 0 {
			for y := range rows {
				rows[y] += "."
			}
		}
		if glyph, ok := menuFont[r]; ok {
			for y := range rows {
				rows[y] += glyph[y]
			}
		}
	}
	return rows
}

// drawPixelWordmark draws the pixel words (AGE, FORGE) with the fire gradient
// across the full icon width.
func drawPixelWordmark(g *mGrid, pal *menuPalette, rows [menuWordH]string, x0, y0, sx, iconW int) {
	w := len(rows[0])
	for py := 0; py < menuWordH; py++ {
		for px := 0; px < w; px++ {
			if rows[py][px] != '#' {
				continue
			}
			// Apply the gradient across the full icon width
			u := float64((x0 + px*sx)) / float64(iconW-1)
			col := rampAt(&pal.heat, u)
			for d := 0; d < sx; d++ {
				g.put(x0+px*sx+d, y0+py, '█', col)
			}
		}
	}
}

func rgba(c tcell.Color) color.RGBA {
	r, g, b := c.RGB()
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

// bannerFont is a five by eight face for the banner's line: the letters it
// uses, the comma and the hyphen. The last row is for descenders.
var bannerFont = map[rune][8]string{
	' ': {".....", ".....", ".....", ".....", ".....", ".....", ".....", "....."},
	',': {".....", ".....", ".....", ".....", ".....", ".##..", "..#..", ".#..."},
	'-': {".....", ".....", ".....", ".....", ".###.", ".....", ".....", "....."},
	'a': {".....", ".....", ".###.", "....#", ".####", "#...#", ".####", "....."},
	'e': {".....", ".....", ".###.", "#...#", "#####", "#....", ".###.", "....."},
	'g': {".....", ".....", ".####", "#...#", "#...#", ".####", "....#", ".###."},
	'i': {"..#..", ".....", ".##..", "..#..", "..#..", "..#..", ".###.", "....."},
	'l': {".##..", "..#..", "..#..", "..#..", "..#..", "..#..", ".###.", "....."},
	'm': {".....", ".....", "##.#.", "#.#.#", "#.#.#", "#.#.#", "#.#.#", "....."},
	'n': {".....", ".....", "#.##.", "##..#", "#...#", "#...#", "#...#", "....."},
	'o': {".....", ".....", ".###.", "#...#", "#...#", "#...#", ".###.", "....."},
	'r': {".....", ".....", "#.##.", "##..#", "#....", "#....", "#....", "....."},
	's': {".....", ".....", ".####", "#....", ".###.", "....#", "####.", "....."},
	't': {".#...", ".#...", "####.", ".#...", ".#...", ".#..#", "..##.", "....."},
	'w': {".....", ".....", "#...#", "#...#", "#.#.#", "#.#.#", ".#.#.", "....."},
	'y': {".....", ".....", "#...#", "#...#", "#...#", ".####", "....#", ".###."},
}
