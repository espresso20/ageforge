package ui

import (
	"image"
	"image/color"
	"math"
	"strings"

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

// medalPicture is a medal w by h pixels, on a clear ground.
func medalPicture(md medal, family string, w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	k := max(1, min(w, h)/medalArt)
	ox, oy := (w-medalArt*k)/2, (h-medalArt*k)/2
	block := func(x, y int, c color.RGBA) {
		for j := 0; j < k; j++ {
			for i := 0; i < k; i++ {
				img.SetRGBA(ox+x*k+i, oy+y*k+j, c)
			}
		}
	}
	m := metalOf(md)
	const c, r = float64(medalArt-1) / 2, float64(medalArt)/2 - 1
	for y := 0; y < medalArt; y++ {
		for x := 0; x < medalArt; x++ {
			dx, dy := float64(x)-c, float64(y)-c
			d := math.Hypot(dx, dy)
			// The milled edge: thirty-two small teeth round the coin.
			if d > r-0.9*(1+math.Cos(math.Atan2(dy, dx)*32))/2 {
				continue
			}
			lit := -(dx*0.6 + dy*0.8) / r // the light comes from the upper left
			var col color.RGBA
			switch {
			case d > r-2: // the lip
				col = m.tone(lit*0.8-0.1, x, y)
			case d > r-6: // the raised rim
				col = m.tone(lit*1.3+0.25, x, y)
			case d > r-7: // the groove under it
				col = m.deep
			default: // the field, a shallow hollow: its far side catches the light
				col = m.tone(-lit*0.55-0.35+0.25*(1-d/r), x, y)
			}
			if md.tier == config.BadgeLegendary && md.state == medalEarned && d > r-6 && d <= r-2 {
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
	const scale = 2
	at := (medalArt - len(icon)*scale) / 2
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
	f := strings.ToLower(family)
	for _, k := range medalIconOrder {
		if strings.Contains(f, k) {
			return medalIcons[k]
		}
	}
	return medalIcons["star"]
}

// medalIconOrder is the family words looked for, in order.
var medalIconOrder = []string{"age", "swift", "wonder", "maxim", "research", "lineage", "payroll", "resource", "civil", "expedition", "ladder", "theme", "map"}

// medalIcons are the signs, thirteen by thirteen.
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
