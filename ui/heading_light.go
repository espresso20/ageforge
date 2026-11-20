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

// heading_light.go sets the light heading: the panel's own pixel font at
// twice its size, every stroke one pixel thick. A letter takes two cells by
// two rows. It is struck as a picture on the font's own pixel grid and the
// window stretches it over the cells, so it is the same at any cell size.

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

// headKey is a heading as the cache knows it.
type headKey struct {
	s  string
	fg color.RGBA
}

var headCache struct {
	sync.Mutex
	set map[headKey]*image.RGBA
}

// headWidth is the cells a heading takes: two a letter.
func headWidth(s string) int { return 2 * len([]rune(s)) }

// headImage is s drawn light in fg on a transparent ground.
func headImage(s string, fg color.RGBA) *image.RGBA {
	k := headKey{s, fg}
	headCache.Lock()
	defer headCache.Unlock()
	if img, ok := headCache.set[k]; ok {
		return img
	}
	if headCache.set == nil || len(headCache.set) > 512 {
		headCache.set = map[headKey]*image.RGBA{}
	}
	rs := []rune(s)
	img := image.NewRGBA(image.Rect(0, 0, 2*headCW*len(rs), 2*headCH))
	for i, r := range rs {
		if r == ' ' {
			continue
		}
		mk := headMask(r)
		on := func(x, y int) bool { return x >= 0 && y >= 0 && x < headCW && y < headCH && mk[y][x] }
		ox := 2 * headCW * i
		set := func(x, y int) {
			if x >= 0 && x < 2*headCW {
				img.SetRGBA(ox+x, y, fg)
			}
		}
		for py := 0; py < headCH; py++ {
			for px := 0; px < headCW; px++ {
				if !on(px, py) {
					continue
				}
				set(2*px, 2*py)
				right, down := on(px+1, py), on(px, py+1)
				if right {
					set(2*px+1, 2*py)
				}
				if down {
					set(2*px, 2*py+1)
				}
				if !right && !down && on(px+1, py+1) {
					set(2*px+1, 2*py+1)
				}
				if !on(px-1, py) && !down && on(px-1, py+1) {
					set(2*px-1, 2*py+1)
				}
			}
		}
	}
	headCache.set[k] = img
	return img
}

// paintHeading sets s as a light heading with its top left at (x, y) and
// returns the cell after it.
func paintHeading(p *kit.Painter, x, y int, s string, fg color.RGBA) int {
	w := headWidth(s)
	if w > 0 {
		p.Picture(image.Rect(x, y, x+w, y+2), headImage(s, fg))
	}
	return x + w
}
