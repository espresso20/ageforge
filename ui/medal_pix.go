package ui

import (
	"image"
	"image/color"
	"math"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
)

// medal_pix.go draws a badge's medal in real pixels, for the window's
// picture layer: a struck coin in the tier's metal, lit from the upper left,
// with a milled edge, a raised rim, a hollow field and the family's sign
// raised in the middle. It is drawn by code at whatever size it is given, in
// square blocks of whole pixels, so it stays pixel art at any font size.

// medalArt is how many art pixels a medal is across.
const medalArt = 56

// bayer4 is the ordered dither the shading steps through.
var bayer4 = [4][4]float64{
	{0 / 16.0, 8 / 16.0, 2 / 16.0, 10 / 16.0},
	{12 / 16.0, 4 / 16.0, 14 / 16.0, 6 / 16.0},
	{3 / 16.0, 11 / 16.0, 1 / 16.0, 9 / 16.0},
	{15 / 16.0, 7 / 16.0, 13 / 16.0, 5 / 16.0},
}

// metal is a medal's three tones, dark to pale, and the tone under the dark
// one that its hollows fall to.
type metal struct{ deep, dark, base, pale color.RGBA }

func rgbOf(c theme.Color) color.RGBA {
	r, g, b := c.RGB()
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

func toward(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return color.RGBA{R: m(a.R, b.R), G: m(a.G, b.G), B: m(a.B, b.B), A: 255}
}

// metalOf is the metal a medal is struck in: its tier's, or a dull steel for
// one not earned yet.
func metalOf(md medal) metal {
	if md.state != medalEarned {
		return metal{
			deep: color.RGBA{R: 0x14, G: 0x17, B: 0x1e, A: 255}, dark: color.RGBA{R: 0x23, G: 0x28, B: 0x33, A: 255},
			base: color.RGBA{R: 0x38, G: 0x3f, B: 0x4c, A: 255}, pale: color.RGBA{R: 0x56, G: 0x5e, B: 0x6e, A: 255},
		}
	}
	d, b, p := theme.BadgeBronzeDark, theme.BadgeBronze, theme.BadgeBronzePale
	switch md.tier {
	case config.BadgeSilver:
		d, b, p = theme.BadgeSilverDark, theme.BadgeSilver, theme.BadgeSilverPale
	case config.BadgeGold, config.BadgeLegendary:
		d, b, p = theme.BadgeGoldDark, theme.BadgeGold, theme.BadgeGoldPale
	case config.BadgePlatinum:
		d, b, p = theme.BadgePlatinumDark, theme.BadgePlatinum, theme.BadgePlatinumPale
	}
	m := metal{dark: rgbOf(theme.BadgeHueColor(d, false)), base: rgbOf(theme.BadgeHueColor(b, false)), pale: rgbOf(theme.BadgeHueColor(p, false))}
	m.deep = toward(m.dark, color.RGBA{A: 255}, 0.45)
	return m
}

// tone picks one of a metal's tones for how lit a point is (-1 in shadow, 1
// in full light), stepping between neighbours with the dither.
func (m metal) tone(lit float64, x, y int) color.RGBA {
	steps := [4]color.RGBA{m.deep, m.dark, m.base, m.pale}
	t := (math.Max(-1, math.Min(1, lit)) + 1) / 2 * 3 // 0..3
	i := int(t)
	if i >= 3 {
		return steps[3]
	}
	if t-float64(i) > bayer4[y&3][x&3] {
		i++
	}
	return steps[i]
}

// medalPicture is a medal w by h pixels, on a clear ground, at the large
// size: the picked badge's.
func medalPicture(md medal, family string, w, h int) *image.RGBA {
	return coinPicture(md, family, w, h, medalArt)
}

// coinArt is the side of a small coin's art, in art pixels: the size the
// coins of the case are struck at. It is half the large medal's, so the
// family's sign is drawn at its own size on it.
const coinArt = 28

// coinPicture is a medal w by h pixels, on a clear ground, drawn art art
// pixels across (medalArt or coinArt) and scaled up by whole pixels to fit.
// The rim, the teeth and the sign keep their proportions at either size.
func coinPicture(md medal, family string, w, h, art int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	k := max(1, min(w, h)/art)
	ox, oy := (w-art*k)/2, (h-art*k)/2
	block := func(x, y int, c color.RGBA) {
		if x < 0 || y < 0 || x >= art || y >= art {
			return
		}
		for j := 0; j < k; j++ {
			for i := 0; i < k; i++ {
				img.SetRGBA(ox+x*k+i, oy+y*k+j, c)
			}
		}
	}
	m := metalOf(md)
	// The parts of the coin, from the edge in: the lip, the raised rim and
	// the groove under it, sized for the art.
	scale := max(1, art/coinArt) // 1 on a small coin, 2 on the large medal
	lip, rim := float64(scale), float64(3*scale)
	teeth, bite := 20.0, 0.6
	if scale > 1 {
		teeth, bite = 32, 0.9
	}
	c, r := float64(art-1)/2, float64(art)/2-1
	for y := 0; y < art; y++ {
		for x := 0; x < art; x++ {
			dx, dy := float64(x)-c, float64(y)-c
			d := math.Hypot(dx, dy)
			// The milled edge: small teeth round the coin.
			if d > r-bite*(1+math.Cos(math.Atan2(dy, dx)*teeth))/2 {
				continue
			}
			lit := -(dx*0.6 + dy*0.8) / r // the light comes from the upper left
			var col color.RGBA
			switch {
			case d > r-lip: // the lip
				col = m.tone(lit*0.8-0.1, x, y)
			case d > r-rim: // the raised rim
				col = m.tone(lit*1.3+0.25, x, y)
			case d > r-rim-float64(scale)/2-0.5: // the groove under it
				col = m.deep
			default: // the field, a shallow hollow: its far side catches the light
				col = m.tone(-lit*0.55-0.35+0.25*(1-d/r), x, y)
			}
			if md.tier == config.BadgeLegendary && md.state == medalEarned && d > r-rim && d <= r-lip {
				// A legendary rim runs through the colours round the coin.
				hue := math.Mod(math.Atan2(dy, dx)/(2*math.Pi)+1, 1)
				col = toward(col, rgbOf(theme.PrismColor(hue, 1, false)), 0.55)
			}
			block(x, y, col)
		}
	}
	// The family's sign, raised in the middle: its shadow first, down and
	// to the right, then the sign in the pale tone with a lit upper edge.
	icon := medalIcon(md, family)
	at := (art - len(icon)*scale) / 2
	if md.crossed {
		// Earned in a modified game: a strike is cut across the coin, from
		// the lower left to the upper right, when the sign is done.
		defer func() {
			for i := 0; i < art; i++ {
				x, y := i, art-1-i
				if math.Hypot(float64(x)-c, float64(y)-c) > r-lip {
					continue
				}
				for t := -scale; t < scale; t++ {
					block(x+t, y, m.deep)
				}
			}
		}()
	}
	on := func(ix, iy int) bool {
		return iy >= 0 && iy < len(icon) && ix >= 0 && ix < len(icon[iy]) && icon[iy][ix] != '.'
	}
	for pass := 0; pass < 2; pass++ {
		for iy := range icon {
			for ix := range icon[iy] {
				if !on(ix, iy) {
					continue
				}
				for j := 0; j < scale; j++ {
					for i := 0; i < scale; i++ {
						x, y := at+ix*scale+i, at+iy*scale+j
						if pass == 0 {
							block(x+1, y+1, m.deep)
							continue
						}
						col := m.pale
						if !on(ix, iy-1) && j == 0 {
							col = toward(m.pale, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 0.45) // the lit edge
						} else if !on(ix, iy+1) && j == scale-1 {
							col = m.base
						}
						block(x, y, col)
					}
				}
			}
		}
	}
	return img
}

// medalIcon is the sign struck on a medal: a keyhole on one not earned, a
// question mark on a hidden one, else its family's.
func medalIcon(md medal, family string) []string {
	switch md.state {
	case medalHidden:
		return medalIcons["hidden"]
	case medalLocked:
		return medalIcons["locked"]
	}
	return medalIcons[medalSign(family)]
}

// medalSign is the name of the sign a family's medals carry: every family
// of the case has its own, and a family the table does not know gets a star.
func medalSign(family string) string {
	if sign, ok := medalSigns[family]; ok {
		return sign
	}
	return "star"
}

// medalSigns is each badge family's sign, by the family's key.
var medalSigns = map[string]string{
	"age": "age", "swift": "swift", "wonder": "wonder", "maximalist": "maxim", "techs": "research",
	"lineage": "lineage", "domain": "payroll", "resource": "resource", "civ": "civil",
	"harbinger": "harbinger", "doom": "doom", "expedition": "expedition", "awakening": "awakening",
	"ladder": "ladder", "theme": "theme", "map": "map", "special": "star",
}

// medalIcons are the signs, 13 by 13: a # is struck, a . is the field.
var medalIcons = map[string][]string{
	"star": {
		"......#......",
		"......#......",
		".....###.....",
		".....###.....",
		"#############",
		".###########.",
		"..#########..",
		"...#######...",
		"...#######...",
		"..####.####..",
		"..###...###..",
		".###.....###.",
		".#.........#.",
	},
	"age": { // an hourglass
		".###########.",
		".###########.",
		"..#.......#..",
		"..##.....##..",
		"...###.###...",
		"....#####....",
		".....###.....",
		"....##.##....",
		"...##.#.##...",
		"..##.###.##..",
		"..#.#####.#..",
		".###########.",
		".###########.",
	},
	"swift": { // a bolt
		".......####..",
		"......####...",
		".....####....",
		"....####.....",
		"...####......",
		"..#########..",
		"..########...",
		"......####...",
		".....####....",
		"....####.....",
		"...####......",
		"..###........",
		"..#..........",
	},
	"wonder": { // a temple
		"......#......",
		"....#####....",
		"..#########..",
		"#############",
		".............",
		".##.##.##.##.",
		".##.##.##.##.",
		".##.##.##.##.",
		".##.##.##.##.",
		".##.##.##.##.",
		".............",
		"#############",
		"#############",
	},
	"maxim": { // a crown
		".............",
		"#.....#.....#",
		"#.....#.....#",
		"##...###...##",
		"##...###...##",
		"###.#####.###",
		"#############",
		"#############",
		"#############",
		".###########.",
		".............",
		".###########.",
		".###########.",
	},
	"research": { // a flask
		"....#####....",
		".....###.....",
		".....###.....",
		".....###.....",
		".....###.....",
		"....#####....",
		"...##...##...",
		"..##.....##..",
		".##.......##.",
		".###########.",
		"#############",
		"#############",
		".###########.",
	},
	"lineage": { // a tree
		"....#####....",
		"..#########..",
		".###########.",
		"#############",
		"#############",
		".###########.",
		"..#########..",
		"....#####....",
		"......#......",
		"......#......",
		"......#......",
		"....#####....",
		"...#######...",
	},
	"payroll": { // a coin
		"....#####....",
		"..#########..",
		".###.....###.",
		".##...#...##.",
		"##..#####..##",
		"##..#.#....##",
		"##..#####..##",
		"##....#.#..##",
		"##..#####..##",
		".##...#...##.",
		".###.....###.",
		"..#########..",
		"....#####....",
	},
	"resource": { // a cut gem
		".............",
		"..#########..",
		".###########.",
		"####.###.####",
		"#############",
		".###########.",
		".###########.",
		"..#########..",
		"...#######...",
		"....#####....",
		".....###.....",
		"......#......",
		".............",
	},
	"civil": { // a flag
		".##..........",
		".###########.",
		".###########.",
		".###########.",
		".###########.",
		".###########.",
		".###########.",
		".##..........",
		".##..........",
		".##..........",
		".##..........",
		".##..........",
		"####.........",
	},
	"expedition": { // a compass
		"....#####....",
		"..##..#..##..",
		".#....#....#.",
		".#...###...#.",
		"#....###....#",
		"#...#####...#",
		"####.....####",
		"#...#####...#",
		"#....###....#",
		".#...###...#.",
		".#....#....#.",
		"..##..#..##..",
		"....#####....",
	},
	"ladder": {
		"..##.....##..",
		"..##.....##..",
		"..#########..",
		"..##.....##..",
		"..##.....##..",
		"..#########..",
		"..##.....##..",
		"..##.....##..",
		"..#########..",
		"..##.....##..",
		"..##.....##..",
		"..#########..",
		"..##.....##..",
	},
	"theme": { // a painter's board
		"...#######...",
		".###########.",
		"####.###.####",
		"###...#...###",
		"####.###.####",
		"#############",
		"######...####",
		"#####.....###",
		"######...####",
		".##########..",
		"..########...",
		"....####.....",
		".............",
	},
	"map": { // a pin
		"....#####....",
		"..#########..",
		".####...####.",
		".###.....###.",
		".###.....###.",
		".####...####.",
		"..#########..",
		"...#######...",
		"....#####....",
		".....###.....",
		".....###.....",
		"......#......",
		"......#......",
	},
	"harbinger": { // an eye
		".............",
		"....#####....",
		"..#########..",
		".###.....###.",
		"###..###..###",
		"##..#####..##",
		"##..#####..##",
		"###..###..###",
		".###.....###.",
		"..#########..",
		"....#####....",
		".............",
		".............",
	},
	"doom": { // a falling star
		"..........###",
		".........####",
		".......#####.",
		"......####...",
		"....#####....",
		"...####......",
		"..#######....",
		".#########...",
		".#########...",
		".#########...",
		".#########...",
		"..#######....",
		"...#####.....",
	},
	"awakening": { // a rising sun
		"......#......",
		"......#......",
		".#....#....#.",
		"..#.......#..",
		"....#####....",
		"...#######...",
		"..#########..",
		"..#########..",
		".###########.",
		".............",
		"#############",
		".............",
		".###########.",
	},
	"locked": { // a keyhole
		"....#####....",
		"..#########..",
		".###########.",
		".####...####.",
		".###.....###.",
		".###.....###.",
		".####...####.",
		".#####.#####.",
		".####...####.",
		".####...####.",
		".###.....###.",
		"..#########..",
		"....#####....",
	},
	"hidden": { // a question mark
		"...#######...",
		"..#########..",
		"..###...###..",
		"..##.....##..",
		".........##..",
		"........###..",
		"......####...",
		".....###.....",
		".....###.....",
		".............",
		".....###.....",
		".....###.....",
		".............",
	},
}
