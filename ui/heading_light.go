package ui

import (
	"image"
	"image/color"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/espresso20/ageforge/window/fonts"
	"github.com/espresso20/ageforge/window/kit"
)

// heading_light.go sets big type: the panel's own pixel font at a whole
// multiple of its size, its strokes kept thin. The light heading is twice
// the size with one-pixel strokes, two cells by two rows a letter. It is
// struck as a picture on the font's own pixel grid and the window stretches
// it over the cells, so it is the same at any cell size.

// The font's cell in pixels, the grid a heading is struck on.
const (
	headCW = 7
	headCH = 15
)

// headFace is the pixel font, parsed once.
var headFace struct {
	once sync.Once
	face font.Face
	base int
}

// headMask is r in the font as the pixels of one cell.
func headMask(r rune) (m [headCH][headCW]bool) {
	headFace.once.Do(func() {
		sf, err := opentype.Parse(fonts.Departure)
		if err != nil {
			return
		}
		f, err := opentype.NewFace(sf, &opentype.FaceOptions{Size: 11, DPI: 72, Hinting: font.HintingFull})
		if err != nil {
			return
		}
		mt := f.Metrics()
		asc, desc := mt.Ascent.Round(), mt.Descent.Round()
		headFace.face, headFace.base = f, (headCH-(asc+desc))/2+asc
	})
	if headFace.face == nil {
		return
	}
	tmp := image.NewAlpha(image.Rect(0, 0, headCW, headCH))
	d := font.Drawer{Dst: tmp, Src: image.Opaque, Face: headFace.face, Dot: fixed.P(0, headFace.base)}
	d.DrawString(string(r))
	for y := 0; y < headCH; y++ {
		for x := 0; x < headCW; x++ {
			m[y][x] = tmp.AlphaAt(x, y).A >= 128
		}
	}
	return m
}

// headKey is a piece of big type as the cache knows it.
type headKey struct {
	s     string
	n, wt int
	fg    color.RGBA
}

var headCache struct {
	sync.Mutex
	set map[headKey]*image.RGBA
}

// headWidth is the cells a heading takes: two a letter.
func headWidth(s string) int { return 2 * len([]rune(s)) }

// bigImage is s in the pixel font with the font's pixels n apart and every
// stroke wt thick, in fg on a transparent ground: n cells by n rows a letter.
func bigImage(s string, n, wt int, fg color.RGBA) *image.RGBA {
	k := headKey{s, n, wt, fg}
	headCache.Lock()
	defer headCache.Unlock()
	if img, ok := headCache.set[k]; ok {
		return img
	}
	if headCache.set == nil || len(headCache.set) > 512 {
		headCache.set = map[headKey]*image.RGBA{}
	}
	rs := []rune(s)
	img := image.NewRGBA(image.Rect(0, 0, n*headCW*len(rs), n*headCH))
	for i, r := range rs {
		if r == ' ' {
			continue
		}
		mk := headMask(r)
		on := func(x, y int) bool { return x >= 0 && y >= 0 && x < headCW && y < headCH && mk[y][x] }
		ox := n * headCW * i
		dot := func(x, y int) {
			for d := 0; d < wt*wt; d++ {
				if dx := x + d%wt; dx >= 0 && dx < n*headCW {
					img.SetRGBA(ox+dx, y+d/wt, fg)
				}
			}
		}
		for py := 0; py < headCH; py++ {
			for px := 0; px < headCW; px++ {
				if !on(px, py) {
					continue
				}
				// A stroke runs to each lit neighbour: right, down, or
				// across a corner where neither side joins it.
				right, down := on(px+1, py), on(px, py+1)
				fall := !right && !down && on(px+1, py+1)
				back := !on(px-1, py) && !down && on(px-1, py+1)
				dot(n*px, n*py)
				for t := 1; t < n; t++ {
					if right {
						dot(n*px+t, n*py)
					}
					if down {
						dot(n*px, n*py+t)
					}
					if fall {
						dot(n*px+t, n*py+t)
					}
					if back {
						dot(n*px-t, n*py+t)
					}
				}
			}
		}
	}
	headCache.set[k] = img
	return img
}

// inkBounds is the box of pixels img actually lights.
func inkBounds(img *image.RGBA) (r image.Rectangle) {
	r = image.Rectangle{Min: img.Rect.Max, Max: img.Rect.Min}
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			if img.RGBAAt(x, y).A != 0 {
				r.Min.X, r.Min.Y = min(r.Min.X, x), min(r.Min.Y, y)
				r.Max.X, r.Max.Y = max(r.Max.X, x+1), max(r.Max.Y, y+1)
			}
		}
	}
	if r.Min.X > r.Max.X {
		return image.Rectangle{}
	}
	return r
}

// paintBig sets s in big type n rows tall, strokes wt thick, with its top
// left at (x, y), and returns the cell after it.
func paintBig(p *kit.Painter, x, y int, s string, n, wt int, fg color.RGBA) int {
	w := n * len([]rune(s))
	if w > 0 {
		p.Picture(image.Rect(x, y, x+w, y+n), bigImage(s, n, wt, fg))
	}
	return x + w
}

// paintHeading sets s as a light heading: two rows tall, one-pixel strokes.
func paintHeading(p *kit.Painter, x, y int, s string, fg color.RGBA) int {
	return paintBig(p, x, y, s, 2, 1, fg)
}
