package ui

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/fonts"
	"github.com/espresso20/ageforge/window/kit"
)

// screenshot.go turns the cells of a terminal screen into a picture: each
// cell is a block of pixels with its background, and its letter drawn in Go
// Mono (a font that ships inside golang.org/x/image, so nothing is fetched).
// The block elements and the box-drawing lines are drawn as shapes, so that
// they meet their neighbours without gaps; any other mark the font lacks is
// left as a blank cell, never a tofu box. The same cells always give the same
// picture.

// Cell size of a screen picture in pixels, and the air around the edge.
const (
	shotCellW = 12
	shotCellH = 24
	shotPad   = 8
)

// shotFonts are the two faces of Go Mono, parsed once.
type shotFonts struct {
	reg, bold         *sfnt.Font
	regFace, boldFace font.Face
}

func loadShotFonts() (*shotFonts, error) {
	var f shotFonts
	var err error
	if f.reg, err = opentype.Parse(gomono.TTF); err != nil {
		return nil, err
	}
	if f.bold, err = opentype.Parse(gomonobold.TTF); err != nil {
		return nil, err
	}
	// Go Mono is 0.6 em wide, so a 12 px cell is a 20 px em.
	opt := &opentype.FaceOptions{Size: float64(shotCellW) / 0.6, DPI: 72, Hinting: font.HintingFull}
	if f.regFace, err = opentype.NewFace(f.reg, opt); err != nil {
		return nil, err
	}
	if f.boldFace, err = opentype.NewFace(f.bold, opt); err != nil {
		return nil, err
	}
	return &f, nil
}

// shotBlend is a mixed t of the way from a to b.
func shotBlend(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t + 0.5) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), 255}
}

// ScreenImage is the picture of a buffer's cells, each cellW by cellH
// pixels, with shotPad pixels of the screen's own background colour
// round the edge. Go Mono is sized for a 12 by 24 cell; other sizes scale the
// shapes but not the letters.
func ScreenImage(b *kit.Buffer, cellW, cellH int) *image.RGBA {
	return screenImage(b, cellW, cellH, loadShotFonts)
}

// The window's own cell: 7 by 15 pixels, Departure Mono on an 11 pixel em.
const (
	pixelCellW = 7
	pixelCellH = 15
)

// ScreenImagePixel is the picture of a buffer in the window's own cell and
// face, so a letter is the size it is in the game.
func ScreenImagePixel(b *kit.Buffer) *image.RGBA {
	return screenImage(b, pixelCellW, pixelCellH, loadPixelFonts)
}

// loadPixelFonts is Departure Mono at its native size; it has no bold, so the
// one face serves both.
func loadPixelFonts() (*shotFonts, error) {
	sf, err := opentype.Parse(fonts.Departure)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(sf, &opentype.FaceOptions{Size: 11, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil, err
	}
	return &shotFonts{reg: sf, bold: sf, regFace: face, boldFace: face}, nil
}

func screenImage(b *kit.Buffer, cellW, cellH int, load func() (*shotFonts, error)) *image.RGBA {
	cells, cols, rows := b.Cells, b.W, b.H
	defBg := theme.RGBA(theme.RoleBackground)
	w, h := cols*cellW+2*shotPad, rows*cellH+2*shotPad
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	edge := defBg
	// Padding takes the colour of the cell nearest to it, so that a screen
	// painted in some other ground does not wear a frame.
	for i := range img.Pix {
		img.Pix[i] = 0
	}
	fill := func(r image.Rectangle, c color.RGBA) {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				img.SetRGBA(x, y, c)
			}
		}
	}
	fill(img.Bounds(), edge)
	if len(cells) < cols*rows {
		return img
	}
	sf, ferr := load()

	type cellPaint struct{ fg, bg color.RGBA }
	paints := make([]cellPaint, cols*rows)
	for i := 0; i < cols*rows; i++ {
		f, b := cells[i].Fg, cells[i].Bg
		if cells[i].Clear() {
			b = defBg // a hole for the window's picture: the page's ground, in a still
		}
		if cells[i].Attr&kit.Reverse != 0 {
			f, b = b, f
		}
		if cells[i].Attr&kit.Dim != 0 {
			f = shotBlend(f, b, 0.4)
		}
		paints[i] = cellPaint{f, b}
	}
	origin := func(cx, cy int) image.Point { return image.Pt(shotPad+cx*cellW, shotPad+cy*cellH) }
	for cy := 0; cy < rows; cy++ {
		for cx := 0; cx < cols; cx++ {
			o := origin(cx, cy)
			fill(image.Rect(o.X, o.Y, o.X+cellW, o.Y+cellH), paints[cy*cols+cx].bg)
		}
	}
	// The padding wears the colour of the nearest cell.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x >= shotPad && x < w-shotPad && y >= shotPad && y < h-shotPad {
				x = w - shotPad - 1
				continue
			}
			cx := min(max((x-shotPad)/cellW, 0), cols-1)
			cy := min(max((y-shotPad)/cellH, 0), rows-1)
			if x < shotPad {
				cx = 0
			}
			if y < shotPad {
				cy = 0
			}
			img.SetRGBA(x, y, paints[cy*cols+cx].bg)
		}
	}

	drawShotPics(img, b, cellW, cellH) // text on a picture is written over it
	for cy := 0; cy < rows; cy++ {
		for cx := 0; cx < cols; cx++ {
			c := cells[cy*cols+cx]
			if c.Cont || c.R == ' ' || c.R == 0 {
				continue
			}
			r := c.R
			p := paints[cy*cols+cx]
			o := origin(cx, cy)
			cell := image.Rect(o.X, o.Y, o.X+cellW, o.Y+cellH)
			if c.Big != 0 { // the top left of a group at twice the size
				cell.Max = cell.Max.Add(image.Pt(cellW, cellH))
				if c.Wide {
					cell.Max.X += 2 * cellW
				}
			}
			switch {
			case r >= 0x2580 && r <= 0x259F:
				drawBlockElement(img, cell, r, p.fg, p.bg)
			case r >= 0x2500 && r <= 0x257F:
				drawBoxLine(img, cell, r, p.fg)
			case symbolShapes[r] != nil:
				drawSymbol(img, cell, symbolShapes[r], p.fg)
			case ferr == nil && c.Big != 0:
				drawBigGlyph(img, sf, cell, r, p.fg, c.Attr&kit.Bold != 0)
			case ferr == nil:
				drawGlyph(img, sf, cell, r, p.fg, c.Attr&kit.Bold != 0)
			}
		}
	}
	return img
}

// drawShotPics draws the buffer's pictures over the cells that still own
// them, each stretched (nearest neighbour) over its rectangle of cells and
// cut to the owning cells, as the window draws them.
func drawShotPics(img *image.RGBA, b *kit.Buffer, cellW, cellH int) {
	for cy := 0; cy < b.H; cy++ {
		for cx := 0; cx < b.W; cx++ {
			pi := int(b.Cells[cy*b.W+cx].Pic)
			if pi == 0 || pi > len(b.Pics) {
				continue
			}
			pic := b.Pics[pi-1]
			sb := pic.Img.Bounds()
			dw, dh := pic.Rect.Dx()*cellW, pic.Rect.Dy()*cellH
			if dw == 0 || dh == 0 {
				continue
			}
			for py := 0; py < cellH; py++ {
				for px := 0; px < cellW; px++ {
					sx := sb.Min.X + ((cx-pic.Rect.Min.X)*cellW+px)*sb.Dx()/dw
					sy := sb.Min.Y + ((cy-pic.Rect.Min.Y)*cellH+py)*sb.Dy()/dh
					src := pic.Img.RGBAAt(sx, sy)
					if src.A == 0 {
						continue
					}
					x, y := shotPad+cx*cellW+px, shotPad+cy*cellH+py
					if src.A == 255 {
						img.SetRGBA(x, y, src)
						continue
					}
					dst := img.RGBAAt(x, y)
					a := uint32(src.A)
					m := func(s, d uint8) uint8 { return uint8((uint32(s)*255 + uint32(d)*(255-a)) / 255) }
					img.SetRGBA(x, y, color.RGBA{m(src.R, dst.R), m(src.G, dst.G), m(src.B, dst.B), 255})
				}
			}
		}
	}
}

// symbolShapes are the marks the epoch screen uses that Go Mono has no shape
// for (the feat icons, the epochs' icons), each as a test of whether a point
// is inside it, in a box from -1 to 1 each way. drawSymbol draws them in a
// square as wide as the cell, centred.
var symbolShapes = map[rune]func(x, y float64) bool{
	'★': func(x, y float64) bool { return inStar(x, y, 5, 1, 0.42) },
	'✦': func(x, y float64) bool { return inStar(x, y, 4, 1, 0.3) },
	'✸': func(x, y float64) bool { return inStar(x, y, 8, 1, 0.55) },
	'◆': func(x, y float64) bool { return math.Abs(x)+math.Abs(y) <= 1 },
	'◈': func(x, y float64) bool {
		d := math.Abs(x) + math.Abs(y)
		return d <= 1 && d > 0.72 || d <= 0.4
	},
	'▣': func(x, y float64) bool {
		m := math.Max(math.Abs(x), math.Abs(y))
		return m <= 0.9 && m > 0.65 || m <= 0.35
	},
	'◉': func(x, y float64) bool {
		r := math.Hypot(x, y)
		return r <= 0.95 && r > 0.7 || r <= 0.4
	},
	'⚙': func(x, y float64) bool {
		r := math.Hypot(x, y)
		if r > 1 || r <= 0.4 {
			return false
		}
		if r <= 0.75 {
			return true
		}
		// Eight teeth, each 22 degrees wide.
		a := math.Mod(math.Atan2(y, x)+2*math.Pi, math.Pi/4)
		return math.Abs(a-math.Pi/8) <= 0.19
	},
	'⚡': func(x, y float64) bool {
		return inPolygon(x, y, [][2]float64{{0.3, -1}, {-0.6, 0.1}, {-0.05, 0.1}, {-0.3, 1}, {0.6, -0.15}, {0.05, -0.15}})
	},
	'⚔': func(x, y float64) bool {
		if math.Abs(x) > 0.95 || math.Abs(y) > 0.95 {
			return false
		}
		blade := math.Abs(math.Abs(x)-math.Abs(y)) <= 0.2
		// A short guard across each blade, low on the picture.
		guard := y > 0.2 && y < 0.4 && math.Abs(math.Abs(x)-y) <= 0.5 && math.Abs(x) < 0.7
		return blade || guard
	},
	'▾': func(x, y float64) bool { return inPolygon(x, y, [][2]float64{{-0.8, -0.5}, {0.8, -0.5}, {0, 0.6}}) },
}

// inStar says whether a point is inside an n-pointed star, one point up,
// with the given outer and inner radii.
func inStar(x, y float64, n int, outer, inner float64) bool {
	pts := make([][2]float64, 0, 2*n)
	for i := 0; i < 2*n; i++ {
		r := outer
		if i%2 == 1 {
			r = inner
		}
		a := -math.Pi/2 + float64(i)*math.Pi/float64(n)
		pts = append(pts, [2]float64{r * math.Cos(a), r * math.Sin(a)})
	}
	return inPolygon(x, y, pts)
}

// inPolygon is the even-odd test.
func inPolygon(x, y float64, pts [][2]float64) bool {
	in := false
	for i, j := 0, len(pts)-1; i < len(pts); j, i = i, i+1 {
		a, b := pts[i], pts[j]
		if (a[1] > y) != (b[1] > y) && x < (b[0]-a[0])*(y-a[1])/(b[1]-a[1])+a[0] {
			in = !in
		}
	}
	return in
}

// drawSymbol draws a shape in a square the width of the cell, centred, with
// its edge smoothed over 4 by 4 samples per pixel.
func drawSymbol(img *image.RGBA, cell image.Rectangle, in func(x, y float64) bool, fg color.RGBA) {
	w, h := cell.Dx(), cell.Dy()
	half := float64(w) / 2 * 0.92
	cx, cy := float64(w)/2, float64(h)/2
	const n = 4
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			hit := 0
			for sy := 0; sy < n; sy++ {
				for sx := 0; sx < n; sx++ {
					x := (float64(px) + (float64(sx)+0.5)/n - cx) / half
					y := (float64(py) + (float64(sy)+0.5)/n - cy) / half
					if math.Abs(x) <= 1 && math.Abs(y) <= 1 && in(x, y) {
						hit++
					}
				}
			}
			if hit > 0 {
				at := image.Pt(cell.Min.X+px, cell.Min.Y+py)
				img.SetRGBA(at.X, at.Y, shotBlend(img.RGBAAt(at.X, at.Y), fg, float64(hit)/(n*n)))
			}
		}
	}
}

// drawGlyph draws one letter in its cell, or nothing where Go Mono has none.
func drawGlyph(img *image.RGBA, f *shotFonts, cell image.Rectangle, r rune, fg color.RGBA, bold bool) {
	sf, face := f.reg, f.regFace
	if bold {
		sf, face = f.bold, f.boldFace
	}
	var buf sfnt.Buffer
	if idx, err := sf.GlyphIndex(&buf, r); err != nil || idx == 0 {
		return
	}
	m := face.Metrics()
	asc, desc := m.Ascent.Round(), m.Descent.Round()
	base := cell.Min.Y + (cell.Dy()-(asc+desc))/2 + asc
	d := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(fg),
		Face: face,
		Dot:  fixed.P(cell.Min.X, base),
	}
	d.DrawString(string(r))
}

// drawBigGlyph draws a letter in a cell twice the size of the font's own: the
// letter at its size, stretched to twice that without smoothing.
func drawBigGlyph(img *image.RGBA, f *shotFonts, cell image.Rectangle, r rune, fg color.RGBA, bold bool) {
	half := image.Rect(0, 0, cell.Dx()/2, cell.Dy()/2)
	tmp := image.NewRGBA(half)
	drawGlyph(tmp, f, half, r, fg, bold)
	for y := cell.Min.Y; y < cell.Max.Y; y++ {
		for x := cell.Min.X; x < cell.Max.X; x++ {
			src := tmp.RGBAAt((x-cell.Min.X)/2, (y-cell.Min.Y)/2)
			if src.A == 0 {
				continue
			}
			dst := img.RGBAAt(x, y)
			a := uint32(src.A) // the pixel is premultiplied
			m := func(s, d uint8) uint8 { return uint8(uint32(s) + uint32(d)*(255-a)/255) }
			img.SetRGBA(x, y, color.RGBA{m(src.R, dst.R), m(src.G, dst.G), m(src.B, dst.B), 255})
		}
	}
}

// blockMask says which quadrants a block element fills: bit 0 upper left,
// 1 upper right, 2 lower left, 3 lower right.
func quadrants(r rune) (mask int, ok bool) {
	const ul, ur, ll, lr = 1, 2, 4, 8
	switch r {
	case 0x2596:
		return ll, true
	case 0x2597:
		return lr, true
	case 0x2598:
		return ul, true
	case 0x2599:
		return ul | ll | lr, true
	case 0x259A:
		return ul | lr, true
	case 0x259B:
		return ul | ur | ll, true
	case 0x259C:
		return ul | ur | lr, true
	case 0x259D:
		return ur, true
	case 0x259E:
		return ur | ll, true
	case 0x259F:
		return ur | ll | lr, true
	}
	return 0, false
}

// drawBlockElement draws U+2580 to U+259F edge to edge: the halves and
// eighths as rectangles, the shades as a mix of the two colours, the
// quadrants as quarters.
func drawBlockElement(img *image.RGBA, cell image.Rectangle, r rune, fg, bg color.RGBA) {
	w, h := cell.Dx(), cell.Dy()
	rect := func(x0, y0, x1, y1 int, c color.RGBA) {
		for y := cell.Min.Y + y0; y < cell.Min.Y+y1; y++ {
			for x := cell.Min.X + x0; x < cell.Min.X+x1; x++ {
				img.SetRGBA(x, y, c)
			}
		}
	}
	eighth := func(total, n int) int { return (total*n + 4) / 8 }
	switch {
	case r == 0x2580:
		rect(0, 0, w, h/2, fg)
	case r >= 0x2581 && r <= 0x2587: // lower one eighth to seven eighths
		rect(0, h-eighth(h, int(r-0x2580)), w, h, fg)
	case r == 0x2588:
		rect(0, 0, w, h, fg)
	case r >= 0x2589 && r <= 0x258F: // left seven eighths to one eighth
		rect(0, 0, eighth(w, int(0x2590-r)), h, fg)
	case r == 0x2590:
		rect(w/2, 0, w, h, fg)
	case r >= 0x2591 && r <= 0x2593: // light, medium, dark shade
		rect(0, 0, w, h, shotBlend(bg, fg, float64(r-0x2590)*0.25))
	case r == 0x2594:
		rect(0, 0, w, eighth(h, 1), fg)
	case r == 0x2595:
		rect(w-eighth(w, 1), 0, w, h, fg)
	default:
		mask, _ := quadrants(r)
		if mask&1 != 0 {
			rect(0, 0, w/2, h/2, fg)
		}
		if mask&2 != 0 {
			rect(w/2, 0, w, h/2, fg)
		}
		if mask&4 != 0 {
			rect(0, h/2, w/2, h, fg)
		}
		if mask&8 != 0 {
			rect(w/2, h/2, w, h, fg)
		}
	}
}

// arms is the four arms of a box-drawing cell: left, right, up, down, each
// 0 (none), 1 (light), 2 (heavy) or 3 (double).
type arms [4]uint8

// boxArms are the straight and double-line marks: the light and heavy family
// in Unicode order from U+2500, and the double family from U+2550.
var boxArms = map[rune]arms{}

func init() {
	// l, r, u, d for U+2500..U+254F, in order. Dashed marks (2504 to 250B,
	// 254C to 254F) are drawn as the solid line they stand for.
	light := []arms{
		{1, 1, 0, 0}, {2, 2, 0, 0}, {0, 0, 1, 1}, {0, 0, 2, 2}, // 2500-2503
		{1, 1, 0, 0}, {2, 2, 0, 0}, {0, 0, 1, 1}, {0, 0, 2, 2}, // 2504-2507
		{1, 1, 0, 0}, {2, 2, 0, 0}, {0, 0, 1, 1}, {0, 0, 2, 2}, // 2508-250B
		{0, 1, 0, 1}, {0, 2, 0, 1}, {0, 1, 0, 2}, {0, 2, 0, 2}, // 250C-250F
		{1, 0, 0, 1}, {2, 0, 0, 1}, {1, 0, 0, 2}, {2, 0, 0, 2}, // 2510-2513
		{0, 1, 1, 0}, {0, 2, 1, 0}, {0, 1, 2, 0}, {0, 2, 2, 0}, // 2514-2517
		{1, 0, 1, 0}, {2, 0, 1, 0}, {1, 0, 2, 0}, {2, 0, 2, 0}, // 2518-251B
		{0, 1, 1, 1}, {0, 2, 1, 1}, {0, 1, 2, 1}, {0, 1, 1, 2}, // 251C-251F
		{0, 1, 2, 2}, {0, 2, 2, 1}, {0, 2, 1, 2}, {0, 2, 2, 2}, // 2520-2523
		{1, 0, 1, 1}, {2, 0, 1, 1}, {1, 0, 2, 1}, {1, 0, 1, 2}, // 2524-2527
		{1, 0, 2, 2}, {2, 0, 2, 1}, {2, 0, 1, 2}, {2, 0, 2, 2}, // 2528-252B
		{1, 1, 0, 1}, {2, 1, 0, 1}, {1, 2, 0, 1}, {2, 2, 0, 1}, // 252C-252F
		{1, 1, 0, 2}, {2, 1, 0, 2}, {1, 2, 0, 2}, {2, 2, 0, 2}, // 2530-2533
		{1, 1, 1, 0}, {2, 1, 1, 0}, {1, 2, 1, 0}, {2, 2, 1, 0}, // 2534-2537
		{1, 1, 2, 0}, {2, 1, 2, 0}, {1, 2, 2, 0}, {2, 2, 2, 0}, // 2538-253B
		{1, 1, 1, 1}, {2, 1, 1, 1}, {1, 2, 1, 1}, {2, 2, 1, 1}, // 253C-253F
		{1, 1, 2, 1}, {1, 1, 1, 2}, {1, 1, 2, 2}, {2, 1, 2, 1}, // 2540-2543
		{1, 2, 2, 1}, {2, 1, 1, 2}, {1, 2, 1, 2}, {2, 2, 2, 1}, // 2544-2547
		{2, 2, 1, 2}, {2, 1, 2, 2}, {1, 2, 2, 2}, {2, 2, 2, 2}, // 2548-254B
		{1, 1, 0, 0}, {2, 2, 0, 0}, {0, 0, 1, 1}, {0, 0, 2, 2}, // 254C-254F
	}
	for i, a := range light {
		boxArms[0x2500+rune(i)] = a
	}
	double := []arms{
		{3, 3, 0, 0}, {0, 0, 3, 3}, // 2550-2551
		{0, 3, 0, 1}, {0, 1, 0, 3}, {0, 3, 0, 3}, // 2552-2554
		{3, 0, 0, 1}, {1, 0, 0, 3}, {3, 0, 0, 3}, // 2555-2557
		{0, 3, 1, 0}, {0, 1, 3, 0}, {0, 3, 3, 0}, // 2558-255A
		{3, 0, 1, 0}, {1, 0, 3, 0}, {3, 0, 3, 0}, // 255B-255D
		{0, 3, 1, 1}, {0, 1, 3, 3}, {0, 3, 3, 3}, // 255E-2560
		{3, 0, 1, 1}, {1, 0, 3, 3}, {3, 0, 3, 3}, // 2561-2563
		{3, 3, 0, 1}, {1, 1, 0, 3}, {3, 3, 0, 3}, // 2564-2566
		{3, 3, 1, 0}, {1, 1, 3, 0}, {3, 3, 3, 0}, // 2567-2569
		{3, 3, 1, 1}, {1, 1, 3, 3}, {3, 3, 3, 3}, // 256A-256C
	}
	for i, a := range double {
		boxArms[0x2550+rune(i)] = a
	}
	// 2574 to 257F: the halves of a line, and light meeting heavy.
	tail := []arms{
		{1, 0, 0, 0}, {0, 0, 1, 0}, {0, 1, 0, 0}, {0, 0, 0, 1},
		{2, 0, 0, 0}, {0, 0, 2, 0}, {0, 2, 0, 0}, {0, 0, 0, 2},
		{1, 2, 0, 0}, {0, 0, 1, 2}, {2, 1, 0, 0}, {0, 0, 2, 1},
	}
	for i, a := range tail {
		boxArms[0x2574+rune(i)] = a
	}
}

// boxArcs are the rounded corners: which two arms each joins.
var boxArcs = map[rune]arms{
	0x256D: {0, 1, 0, 1}, // ╭
	0x256E: {1, 0, 0, 1}, // ╮
	0x256F: {1, 0, 1, 0}, // ╯
	0x2570: {0, 1, 1, 0}, // ╰
}

// drawBoxLine draws a box-drawing mark as lines from the cell's centre to
// its edges. Light lines are 2 px, heavy 4 px, and double lines two light
// ones 2 px apart. A mark that is not drawn here (the diagonals) is blank.
func drawBoxLine(img *image.RGBA, cell image.Rectangle, r rune, fg color.RGBA) {
	w, h := cell.Dx(), cell.Dy()
	if a, ok := boxArcs[r]; ok {
		drawArc(img, cell, a, fg)
		return
	}
	a, ok := boxArms[r]
	if !ok {
		return
	}
	const light, heavy, gap = 2, 4, 2
	cx, cy := w/2, h/2
	// reach is how far the lines across an arm spread from the centre, so
	// that the arm closes the corner it makes.
	width := func(k uint8) int {
		switch k {
		case 2:
			return heavy
		case 3:
			return 2*light + gap
		case 1:
			return light
		}
		return 0
	}
	rect := func(x0, y0, x1, y1 int) {
		for y := max(y0, 0); y < min(y1, h); y++ {
			for x := max(x0, 0); x < min(x1, w); x++ {
				img.SetRGBA(cell.Min.X+x, cell.Min.Y+y, fg)
			}
		}
	}
	// strokes lists the offsets (from the centre line, start and end) of the
	// lines an arm of kind k is made of.
	strokes := func(k uint8) [][2]int {
		switch k {
		case 1:
			return [][2]int{{-light / 2, light / 2}}
		case 2:
			return [][2]int{{-heavy / 2, heavy / 2}}
		case 3:
			o := (light + gap) / 2
			return [][2]int{{-o - light/2, -o + light/2}, {o - light/2, o + light/2}}
		}
		return nil
	}
	ext := func(a, b uint8) int { return max(width(a), width(b)) / 2 }
	// horizontal arms reach the centre plus half the width of the vertical
	// arms, and the other way round.
	vext := ext(a[2], a[3])
	hext := ext(a[0], a[1])
	for _, s := range strokes(a[0]) { // left
		rect(0, cy+s[0], cx+vext, cy+s[1])
	}
	for _, s := range strokes(a[1]) { // right
		rect(cx-vext, cy+s[0], w, cy+s[1])
	}
	for _, s := range strokes(a[2]) { // up
		rect(cx+s[0], 0, cx+s[1], cy+hext)
	}
	for _, s := range strokes(a[3]) { // down
		rect(cx+s[0], cy-hext, cx+s[1], h)
	}
}

// drawArc draws a rounded corner joining two arms, as a quarter of an
// ellipse centred on the far corner of the cell, 2 px thick, averaged over
// 4 by 4 samples per pixel so that its edge is smooth.
func drawArc(img *image.RGBA, cell image.Rectangle, a arms, fg color.RGBA) {
	w, h := float64(cell.Dx()), float64(cell.Dy())
	// The ellipse is centred on the corner opposite the bend.
	ex, ey := 0.0, 0.0
	if a[0] == 0 { // joins right: the bend is at the left, centre on the right
		ex = w
	}
	if a[2] == 0 { // joins down: centre on the bottom
		ey = h
	}
	rx, ry := w/2, h/2
	const t, n = 2.0, 4
	for py := 0; py < cell.Dy(); py++ {
		for px := 0; px < cell.Dx(); px++ {
			hit := 0
			for sy := 0; sy < n; sy++ {
				for sx := 0; sx < n; sx++ {
					x := float64(px) + (float64(sx)+0.5)/n
					y := float64(py) + (float64(sy)+0.5)/n
					// Only the quarter facing the cell.
					if (ex == 0 && x > rx) || (ex == w && x < rx) || (ey == 0 && y > ry) || (ey == h && y < ry) {
						continue
					}
					// Distance from the curve, in pixels: how far d is from
					// 1 over how fast d grows per pixel there.
					ux, uy := (x-ex)/rx, (y-ey)/ry
					d := math.Hypot(ux, uy)
					if d == 0 {
						continue
					}
					grad := math.Hypot(ux/rx, uy/ry) / d
					if math.Abs(d-1)/grad <= t/2 {
						hit++
					}
				}
			}
			if hit == 0 {
				continue
			}
			at := image.Pt(cell.Min.X+px, cell.Min.Y+py)
			bg := img.RGBAAt(at.X, at.Y)
			img.SetRGBA(at.X, at.Y, shotBlend(bg, fg, float64(hit)/(n*n)))
		}
	}
}
