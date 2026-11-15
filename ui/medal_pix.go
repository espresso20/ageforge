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
// with a milled edge, a raised rim, a hollow field and its sign raised in the
// middle. It is drawn by code at whatever size it is given, in square blocks
// of whole pixels, so it stays pixel art at any font size.
//
// An earned coin also puts on a show, and the show grows with the tier: a
// band of light crosses a bronze coin now and then; silver adds a sparkle;
// on gold a light runs round the rim; platinum does all of it more often,
// with a second band and the studs of its rim catching the light; a
// legendary coin's rim turns through the colours without rest, and the
// large one stands on turning rays. A coin's show is a loop of poses
// (coinPoseAt), and a pose is drawn by coinPicture.

const (
	// medalArt is how many art pixels the large medal is across, and coinArt
	// a coin of the case: half of it, so a sign is drawn at its own size on
	// the small coin and doubled on the large one.
	medalArt = 56
	coinArt  = 28

	// coinStep is how long a pose of a coin's show lasts: sixteen a second.
	coinStep = mapAnimStep / 2

	glintSteps = 8  // poses the band of light takes to cross a coin
	sparkSteps = 6  // poses a sparkle takes to open and close
	rimSteps   = 16 // poses the light takes to run round the rim
	hueSteps   = 24 // poses a legendary rim takes to turn through the colours
	turnSteps  = 3  // poses a picked medal takes to turn and face the player
)

// bayer4 is the ordered dither the shading steps through.
var bayer4 = [4][4]float64{
	{0 / 16.0, 8 / 16.0, 2 / 16.0, 10 / 16.0},
	{12 / 16.0, 4 / 16.0, 14 / 16.0, 6 / 16.0},
	{3 / 16.0, 11 / 16.0, 1 / 16.0, 9 / 16.0},
	{15 / 16.0, 7 / 16.0, 13 / 16.0, 5 / 16.0},
}

var coinWhite = color.RGBA{R: 255, G: 255, B: 255, A: 255}

// metal is a medal's three tones, dark to pale, and the tone under the dark
// one that its hollows fall to.
type metal struct{ deep, dark, base, pale color.RGBA }

func rgbOf(c theme.Color) color.RGBA {
	r, g, b := c.RGB()
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

func toward(a, b color.RGBA, t float64) color.RGBA {
	t = math.Max(0, math.Min(1, t))
	m := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return color.RGBA{R: m(a.R, b.R), G: m(a.G, b.G), B: m(a.B, b.B), A: 255}
}

// coinLook is a coin as it lies: the metal it is struck in, what is struck
// on it, and how large its art is.
type coinLook struct {
	tier    config.BadgeTier
	state   medalState
	sign    string // a name of medalIcons
	crossed bool   // earned in a modified game: a strike is cut across it
	art     int    // medalArt or coinArt
}

func (l coinLook) earned() bool { return l.state == medalEarned }

// metalOf is the metal a coin is struck in: its tier's, or a dull steel for
// one not earned yet. Platinum is the cold one, a blue-white, so it is not
// taken for silver at the size of a coin of the case.
func metalOf(l coinLook) metal {
	if !l.earned() {
		return metal{
			deep: color.RGBA{R: 0x14, G: 0x17, B: 0x1e, A: 255}, dark: color.RGBA{R: 0x23, G: 0x28, B: 0x33, A: 255},
			base: color.RGBA{R: 0x38, G: 0x3f, B: 0x4c, A: 255}, pale: color.RGBA{R: 0x56, G: 0x5e, B: 0x6e, A: 255},
		}
	}
	d, b, p := theme.BadgeBronzeDark, theme.BadgeBronze, theme.BadgeBronzePale
	switch l.tier {
	case config.BadgeSilver:
		d, b, p = theme.BadgeSilverDark, theme.BadgeSilver, theme.BadgeSilverPale
	case config.BadgeGold, config.BadgeLegendary:
		d, b, p = theme.BadgeGoldDark, theme.BadgeGold, theme.BadgeGoldPale
	case config.BadgePlatinum:
		d, b, p = theme.BadgePlatinumDark, theme.BadgePlatinum, theme.BadgePlatinumPale
	}
	m := metal{dark: rgbOf(theme.BadgeHueColor(d, false)), base: rgbOf(theme.BadgeHueColor(b, false)), pale: rgbOf(theme.BadgeHueColor(p, false))}
	if l.tier == config.BadgePlatinum {
		ice := color.RGBA{R: 0x5c, G: 0x9d, B: 0xff, A: 255}
		m.dark, m.base, m.pale = toward(m.dark, ice, 0.5), toward(m.base, ice, 0.38), toward(m.pale, coinWhite, 0.6)
	}
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

// coinPose is where a coin is in its show. The zero pose is the coin at
// rest.
type coinPose struct {
	glint int8    // the band of light crossing the coin: its step, from 1
	rim   int8    // the light running round the rim: its step, from 1
	spark [3]int8 // each sparkle: how far through its opening and closing, from 1
	hue   int8    // how far a legendary rim has turned through the colours
	turn  int8    // a picked medal turning to face the player: its step, from 1
}

// coinShow is a tier's show: how many poses its loop lasts, the pose each
// sparkle opens at and the pose the light sets off round the rim (-1 for
// one the tier does not have). The band of light crosses at the start.
type coinShow struct {
	loop   int
	sparks [3]int
	rim    int
}

// coinShows are the shows, by tier: each tier has all the tier below has,
// more of it, and sooner again.
var coinShows = map[config.BadgeTier]coinShow{
	config.BadgeBronze:    {loop: 144, sparks: [3]int{-1, -1, -1}, rim: -1},
	config.BadgeSilver:    {loop: 112, sparks: [3]int{46, -1, -1}, rim: -1},
	config.BadgeGold:      {loop: 88, sparks: [3]int{20, 50, -1}, rim: 60},
	config.BadgePlatinum:  {loop: 72, sparks: [3]int{12, 28, 44}, rim: 50},
	config.BadgeLegendary: {loop: 2 * hueSteps, sparks: [3]int{10, 22, 34}, rim: -1},
}

// coinPoseAt is a coin's pose n poses into its show, which goes round for as
// long as it is watched. A coin not earned has no show, and atRest is every
// coin at rest.
func coinPoseAt(l coinLook, n int) coinPose {
	var p coinPose
	if n == atRest || !l.earned() {
		return p
	}
	show, ok := coinShows[l.tier]
	if !ok {
		show = coinShows[config.BadgeBronze]
	}
	t := (n%show.loop + show.loop) % show.loop
	if t < glintSteps {
		p.glint = int8(t + 1)
	}
	for i, at := range show.sparks {
		if at >= 0 && t >= at && t < at+sparkSteps {
			p.spark[i] = int8(t - at + 1)
		}
	}
	if show.rim >= 0 && t >= show.rim && t < show.rim+rimSteps {
		p.rim = int8(t - show.rim + 1)
	}
	if l.tier == config.BadgeLegendary {
		p.hue = int8(t / 2 % hueSteps)
	}
	return p
}

// pickedPoseAt is the pose of the large medal since poses after it was
// picked: it turns to face the player, then its show begins, so the band of
// light crosses it as it comes to rest.
func pickedPoseAt(l coinLook, since int) coinPose {
	switch {
	case since == atRest:
		return coinPose{}
	case since >= 0 && since < turnSteps:
		return coinPose{turn: int8(since + 1)}
	}
	return coinPoseAt(l, since-turnSteps)
}

// coinCanvas is the art pixels a coin is drawn on before it is scaled up: a
// pixel with no alpha is clear.
type coinCanvas struct {
	w, h int
	px   []color.RGBA
}

func (cv *coinCanvas) set(x, y int, c color.RGBA) {
	if x >= 0 && y >= 0 && x < cv.w && y < cv.h {
		cv.px[y*cv.w+x] = c
	}
}

func (cv *coinCanvas) at(x, y int) color.RGBA {
	if x < 0 || y < 0 || x >= cv.w || y >= cv.h {
		return color.RGBA{}
	}
	return cv.px[y*cv.w+x]
}

// medalPicture is a medal at rest w by h pixels, at the large size.
func medalPicture(md medal, family string, w, h int) *image.RGBA {
	return coinPicture(lookOf(md, medalSign(family), medalArt), w, h, coinPose{})
}

// lookOf is the look of a medal at a size of art: it carries its family's
// sign, or its own when it is one of the badges that have one, and one not
// earned carries a keyhole or a question mark.
func lookOf(md medal, sign string, art int) coinLook {
	if _, own := medalIcons[md.special]; own {
		sign = md.special
	}
	switch md.state {
	case medalHidden:
		sign = "question"
	case medalLocked:
		sign = "keyhole"
	}
	return coinLook{tier: md.tier, state: md.state, sign: sign, crossed: md.crossed, art: art}
}

// coinPicture is a coin in a pose, w by h pixels on a clear ground: drawn
// in art pixels, the coin in the middle, and scaled up by whole pixels to
// fit. The rim, the teeth and the sign keep their proportions at either
// size of art. What room is left round the coin is where a sparkle reaches
// and a legendary medal's rays stand.
func coinPicture(l coinLook, w, h int, pose coinPose) *image.RGBA {
	art := l.art
	k := max(1, min(w, h)/art)
	cv := &coinCanvas{w: w / k, h: h / k}
	cv.px = make([]color.RGBA, cv.w*cv.h)
	ox, oy := (cv.w-art)/2, (cv.h-art)/2
	m := metalOf(l)
	// The parts of the coin, from the edge in: the lip, the raised rim and
	// the groove under it, sized for the art.
	scale := max(1, art/coinArt) // 1 on a small coin, 2 on the large medal
	lip, rim := float64(scale), float64(3*scale)
	teeth, bite := 20.0, 0.6
	if scale > 1 {
		teeth, bite = 32, 0.9
	}
	c, r := float64(art-1)/2, float64(art)/2-1
	legend := l.tier == config.BadgeLegendary && l.earned()
	// prism is a colour of a legendary rim, a part of the way round the
	// wheel (0 to 1) from where the rim has turned to.
	prism := func(part float64, stop int) color.RGBA {
		return rgbOf(theme.PrismColor(math.Mod(part+float64(pose.hue)/hueSteps+2, 1)*360, stop, false))
	}

	if legend && min(ox, oy) >= 3*scale {
		// The rays a legendary medal stands on: sixteen thin ones, long and
		// short by turns, in the colours of its rim, turning slowly.
		const rays = 16
		reach := float64(min(ox, oy))
		turn := 2 * math.Pi / rays * float64(pose.hue) / hueSteps
		for y := 0; y < cv.h; y++ {
			for x := 0; x < cv.w; x++ {
				dx, dy := float64(x-ox)-c, float64(y-oy)-c
				d := math.Hypot(dx, dy) - r
				a := math.Atan2(dy, dx) - turn
				ray := math.Round(a / (2 * math.Pi / rays))
				long := reach
				if int(ray)&1 != 0 {
					long = reach * 0.55
				}
				off := math.Abs(a-ray*2*math.Pi/rays) * (r + d) // how far from the ray's middle, in art pixels
				if d < 2 || d > long || off > 0.8+1.3*(1-d/long) {
					continue
				}
				cv.set(x, y, prism((a+turn)/(2*math.Pi), theme.PrismPale))
			}
		}
	}

	// The band of light: how much of it falls on a point of the coin.
	g := coinGlints[config.BadgeBronze]
	if set, ok := coinGlints[l.tier]; ok {
		g = set
	}
	shine := func(dx, dy float64) float64 {
		if pose.glint == 0 {
			return 0
		}
		u := (dx + dy) / (r * math.Sqrt2) // -1 at the upper left, 1 at the lower right
		at := -1.3 + 2.6*float64(pose.glint-1)/(glintSteps-1)
		b := math.Max(0, 1-math.Abs(u-at)/g.half)
		if g.twin {
			b = math.Max(b, 0.7*math.Max(0, 1-math.Abs(u-at+2.2*g.half)/(g.half*0.45)))
		}
		return math.Round(b*3) / 3 // in steps, as the shading is
	}
	glintTint := func(dx, dy float64) color.RGBA {
		if legend {
			return toward(coinWhite, prism((dx+dy)/(4*r), theme.PrismPale), 0.6)
		}
		return coinWhite
	}
	// The light on its way round the rim, from the upper left: how much of
	// it falls at an angle. Platinum has two, on opposite sides.
	rimLight := func(a float64) float64 {
		if pose.rim == 0 {
			return 0
		}
		at := -0.75*math.Pi + 2*math.Pi*float64(pose.rim-1)/rimSteps
		best := 0.0
		for i := 0; i < g.lights; i++ {
			d := math.Abs(math.Remainder(a-at-float64(i)*math.Pi, 2*math.Pi))
			best = math.Max(best, 1-d/0.7)
		}
		return best
	}

	for y := 0; y < art; y++ {
		for x := 0; x < art; x++ {
			dx, dy := float64(x)-c, float64(y)-c
			d := math.Hypot(dx, dy)
			a := math.Atan2(dy, dx)
			// The milled edge: small teeth round the coin.
			if d > r-bite*(1+math.Cos(a*teeth))/2 {
				continue
			}
			lit := -(dx*0.6 + dy*0.8) / r // the light comes from the upper left
			b := shine(dx, dy)
			var col color.RGBA
			switch {
			case d > r-lip: // the lip
				col = m.tone(lit*0.8-0.1+b, x, y)
			case d > r-rim: // the raised rim
				col = m.tone(lit*1.3+0.25+b, x, y)
			case d > r-rim-float64(scale)/2-0.5: // the groove under it
				col = toward(m.deep, m.dark, b)
				b = 0
			default: // the field, a shallow hollow: its far side catches the light
				col = m.tone(-lit*0.55-0.35+0.25*(1-d/r)+b*1.2, x, y)
			}
			if d > r-rim {
				if legend && d <= r-lip {
					// A legendary rim runs through the colours round the coin.
					col = toward(col, prism(a/(2*math.Pi), theme.PrismBase), 0.75)
				}
				col = toward(col, coinWhite, 0.9*rimLight(a))
			}
			cv.set(ox+x, oy+y, toward(col, glintTint(dx, dy), b*g.white))
		}
	}

	// The sign, raised in the middle: its shadow first, down and to the
	// right, then the sign in the pale tone with a lit upper edge.
	icon := medalIcons[l.sign]
	at := (art - len(icon)*scale) / 2
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
							cv.set(ox+x+1, oy+y+1, m.deep)
							continue
						}
						col := m.pale
						if !on(ix, iy-1) && j == 0 {
							col = toward(m.pale, coinWhite, 0.45) // the lit edge
						} else if !on(ix, iy+1) && j == scale-1 {
							col = m.base
						}
						dx, dy := float64(x)-c, float64(y)-c
						cv.set(ox+x, oy+y, toward(col, glintTint(dx, dy), shine(dx, dy)))
					}
				}
			}
		}
	}
	if l.crossed {
		// Earned in a modified game: a strike is cut across the coin, from
		// the lower left to the upper right.
		for i := 0; i < art; i++ {
			x, y := i, art-1-i
			if math.Hypot(float64(x)-c, float64(y)-c) > r-lip {
				continue
			}
			for t := -scale; t < scale; t++ {
				cv.set(ox+x+t, oy+y, m.deep)
			}
		}
	}
	if l.tier == config.BadgePlatinum && l.earned() {
		// Platinum's studs: four blue stones set in the rim, which flare as
		// the light on the rim goes by.
		for i := 0; i < 4; i++ {
			a := float64(i) * math.Pi / 2
			sx := int(math.Round(c + (r-rim/2-lip/2)*math.Cos(a) - float64(scale-1)/2 - 0.5))
			sy := int(math.Round(c + (r-rim/2-lip/2)*math.Sin(a) - float64(scale-1)/2 - 0.5))
			for j := 0; j < 2*scale; j++ {
				for i := 0; i < 2*scale; i++ {
					col := coinStud.base
					switch {
					case i < scale && j < scale:
						col = coinStud.pale
					case i >= scale && j >= scale:
						col = coinStud.dark
					}
					cv.set(ox+sx+i, oy+sy+j, toward(col, coinWhite, rimLight(a)))
				}
			}
		}
	}

	if pose.turn > 0 {
		cv.squash(float64(pose.turn)/(turnSteps+1), m.dark)
	}
	// The sparkles, over everything: four-pointed stars on the rim that
	// open and close.
	for i, step := range pose.spark {
		if step == 0 {
			continue
		}
		arm := [sparkSteps]int{0, 1, 2, 3, 2, 1}[step-1]
		sa := coinSparks[i]
		sx := ox + int(math.Round(c+r*0.8*math.Cos(sa)))
		sy := oy + int(math.Round(c+r*0.8*math.Sin(sa)))
		tint := m.pale
		if legend {
			tint = prism(float64(i)/3, theme.PrismBase)
		}
		unit := func(ux, uy int, col color.RGBA) {
			for j := 0; j < scale; j++ {
				for i := 0; i < scale; i++ {
					cv.set(sx+ux*scale+i, sy+uy*scale+j, col)
				}
			}
		}
		for n := arm; n >= 1; n-- {
			col := toward(coinWhite, tint, float64(n)/float64(arm+1))
			unit(n, 0, col)
			unit(-n, 0, col)
			unit(0, n, col)
			unit(0, -n, col)
		}
		if arm == 3 {
			for _, dg := range [4][2]int{{1, 1}, {-1, 1}, {1, -1}, {-1, -1}} {
				unit(dg[0], dg[1], toward(coinWhite, tint, 0.5))
			}
		}
		unit(0, 0, coinWhite)
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	px, py := (w-cv.w*k)/2, (h-cv.h*k)/2
	for y := 0; y < cv.h; y++ {
		for x := 0; x < cv.w; x++ {
			col := cv.px[y*cv.w+x]
			if col.A == 0 {
				continue
			}
			for j := 0; j < k; j++ {
				for i := 0; i < k; i++ {
					img.SetRGBA(px+x*k+i, py+y*k+j, col)
				}
			}
		}
	}
	return img
}

// squash turns what is on the canvas about its upright middle, to f of its
// width, and shades it as it turns away from the light.
func (cv *coinCanvas) squash(f float64, shade color.RGBA) {
	src := append([]color.RGBA(nil), cv.px...)
	mid := float64(cv.w-1) / 2
	for y := 0; y < cv.h; y++ {
		for x := 0; x < cv.w; x++ {
			sx := int(math.Round(mid + (float64(x)-mid)/f))
			var col color.RGBA
			if sx >= 0 && sx < cv.w {
				if col = src[y*cv.w+sx]; col.A != 0 {
					col = toward(col, shade, (1-f)*0.45)
				}
			}
			cv.px[y*cv.w+x] = col
		}
	}
}

// coinGlint is how a tier's band of light looks: half its width (the coin
// is 2 across), how far it whitens the metal, whether a thin second band
// goes ahead of it, and how many lights run round the rim.
type coinGlint struct {
	half, white float64
	twin        bool
	lights      int
}

var coinGlints = map[config.BadgeTier]coinGlint{
	config.BadgeBronze:    {half: 0.22, white: 0.3},
	config.BadgeSilver:    {half: 0.26, white: 0.55},
	config.BadgeGold:      {half: 0.32, white: 0.65, lights: 1},
	config.BadgePlatinum:  {half: 0.3, white: 0.85, twin: true, lights: 2},
	config.BadgeLegendary: {half: 0.38, white: 0.8, twin: true},
}

// coinSparks is where on the rim each sparkle sits, as an angle: the upper
// left, where the light lands, then the right and the foot.
var coinSparks = [3]float64{-0.75 * math.Pi, -0.12 * math.Pi, 0.6 * math.Pi}

// coinStud is the blue stone of a platinum rim.
var coinStud = metal{dark: color.RGBA{R: 0x1e, G: 0x3a, B: 0x9a, A: 255}, base: color.RGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 255}, pale: color.RGBA{R: 0xc4, G: 0xe0, B: 0xff, A: 255}}

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

// medalIcons are the signs, 13 by 13: a # is struck, a . is the field. A
// family's sign goes by a name of medalSigns; a badge with art of its own
// (badge_emblems.go) has a sign of its own, under that art's key.
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
	specialCookieJar: { // a jar with its lid
		"....#####....",
		"...#######...",
		".............",
		".###########.",
		"#############",
		"###.#####.###",
		"#############",
		"#####.#######",
		"#############",
		"###.####.####",
		"#############",
		".###########.",
		"..#########..",
	},
	specialLedger: { // an open book
		".............",
		".#####.#####.",
		"#######.#####",
		"##...#.#...##",
		"######.######",
		"##...#.#...##",
		"######.######",
		"##...#.#...##",
		"######.######",
		"##...#.#...##",
		"#############",
		".#####.#####.",
		".............",
	},
	specialSource: { // a key
		".............",
		".............",
		"..####.......",
		".######......",
		"##....#######",
		"##....#######",
		"##....##.#.##",
		".######..#.##",
		"..####.......",
		".............",
		".............",
		".............",
		".............",
	},
	specialChains: { // two links of a chain
		".............",
		".#######.....",
		"##.....##....",
		"#.......#....",
		"#....########",
		"#...##..#...#",
		"##..#..##...#",
		".####.##....#",
		"....#.......#",
		"....##.....##",
		".....#######.",
		".............",
		".............",
	},
	specialBoxes: { // a ticked box
		".............",
		"#############",
		"#...........#",
		"#.........#.#",
		"#........##.#",
		"#.......##..#",
		"#.#....##...#",
		"#.##..##....#",
		"#..####.....#",
		"#...##......#",
		"#...........#",
		"#############",
		".............",
	},
	specialMuseum: { // a vase on a stand
		"....#####....",
		".....###.....",
		"....#####....",
		"..#########..",
		".###########.",
		".###########.",
		"..#########..",
		"...#######...",
		"....#####....",
		".....###.....",
		"....#####....",
		".............",
		".###########.",
	},
	specialWonders: { // three pyramids
		".............",
		"......#......",
		".....###.....",
		"....#####....",
		"....#####....",
		"...#######...",
		"..#########..",
		".#.#######.#.",
		"###.#####.###",
		"###.#####.###",
		"#####.#.#####",
		"#############",
		".............",
	},
	specialRogues: { // a portrait in its frame
		"#############",
		"#...........#",
		"#....###....#",
		"#...#####...#",
		"#...#####...#",
		"#....###....#",
		"#...........#",
		"#...#####...#",
		"#..#######..#",
		"#.#########.#",
		"#.#########.#",
		"#...........#",
		"#############",
	},
	specialEndings: { // a flame
		"......#......",
		".....##......",
		".....###.....",
		"....####.#...",
		"...#####.##..",
		"..######.###.",
		"..##.#######.",
		".##...######.",
		".##....#####.",
		".##....####..",
		"..##...###...",
		"...#######...",
		"....#####....",
	},
	specialUndying: { // a shield
		"#############",
		"#############",
		"######.######",
		"######.######",
		"##.........##",
		"######.######",
		"######.######",
		".#####.#####.",
		".#####.#####.",
		"..####.####..",
		"...#######...",
		"....#####....",
		"......#......",
	},
	specialTranscend: { // a ringed world
		".............",
		".........###.",
		"....####...##",
		"..########..#",
		".##########.#",
		".##########.#",
		"############.",
		"#.#########..",
		"#.#########..",
		"#..#######...",
		"##...####....",
		".###.........",
		".............",
	},
	"keyhole": { // on a medal not earned yet
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
	"question": { // on a hidden medal
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
