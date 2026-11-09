package mapstyle

import (
	"image"
	"image/color"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/theme"
)

// A Stage is where a window that can show pictures takes a map's image: the
// view says where the picture goes and how to build it, and the window builds
// it at its own pixel size, when it has the time. The cells under the picture
// are drawn with theme.Transparent as their ground so the image shows through,
// and the map's labels, cursor and brackets stay cells on top.
type Stage interface {
	// Place says that slot shows p. It is called on every draw, and is cheap
	// when nothing has changed: the window builds again only when Key changes.
	Place(slot string, p Placement)
	// Clear says that slot shows no image any more.
	Clear(slot string)
}

// A Placement is one image on the stage.
type Placement struct {
	Cells image.Rectangle // where it goes, in screen cells
	// Key says what the picture is made of, the size of a cell aside: it
	// changes when the scale, the size, the scene or the colours do.
	Key uint64
	// Build draws the picture for cells of cw by ch pixels. It runs on its own
	// goroutine, so it holds only what was copied when the view published it:
	// no theme, no engine, no view.
	Build func(cw, ch int) *image.RGBA
}

// imaged says whether the view draws this zoom as an image when a stage is there.
func (v *view) imaged(z int) bool { return v.stars && z == zCosmos }

// placeVoid puts the Galactic void on the stage and clears its cells to the
// transparent ground. The palette's colours are resolved here, on the UI
// goroutine; the build gets only the copies.
func (v *view) placeVoid(cv *Canvas, st Stage, slot string, x, y, w, h int, seed int64) {
	var hue [mapgen.NumHues]color.RGBA
	key := fnvMix(0xcbf29ce484222325, uint64(seed), uint64(w), uint64(h))
	rgba := func(c tcell.Color) color.RGBA {
		r, g, b := c.RGB()
		return color.RGBA{uint8(max(r, 0)), uint8(max(g, 0)), uint8(max(b, 0)), 255}
	}
	bg := rgba(v.pal.bg)
	key = fnvMix(key, uint64(bg.R)<<16|uint64(bg.G)<<8|uint64(bg.B))
	for i, c := range v.pal.hue {
		hue[i] = rgba(c)
		key = fnvMix(key, uint64(hue[i].R)<<16|uint64(hue[i].G)<<8|uint64(hue[i].B))
	}
	cells := image.Rect(cv.R.X+x, cv.R.Y+y, cv.R.X+x+w, cv.R.Y+y+h)
	st.Place(slot, Placement{Cells: cells, Key: key, Build: func(cw, ch int) *image.RGBA {
		return (&mapgen.Scene{Seed: seed}).VoidImage(w, h, cw, ch, bg, func(h mapgen.Hue) color.RGBA { return hue[h] })
	}})
	clear := tcell.StyleDefault.Foreground(v.pal.bg).Background(theme.Transparent)
	for yy := 0; yy < h; yy++ {
		for xx := 0; xx < w; xx++ {
			cv.Put(x+xx, y+yy, ' ', clear)
		}
	}
}

func fnvMix(h uint64, vs ...uint64) uint64 {
	for _, v := range vs {
		h = (h ^ v) * 0x100000001b3
	}
	return h
}
