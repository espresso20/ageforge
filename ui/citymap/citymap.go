// Package citymap renders AgeForge's capital city as a theme-aware, procedurally
// generated map streamed to the terminal through half-block characters.
//
// This is the P1 foundation of the map overhaul (see
// design-and-architecture/map-overhaul.md). It replaces the retired MapV4
// (ui/map.go), which composited 51MB of realistic photo backgrounds and was
// invisible to the theme system. Here the terrain is generated from a small
// self-contained FBM noise field and tinted entirely from the active theme
// palette, so switching themes retints the whole map live. Buildings appear as
// simple bright markers in a center-out scatter.
//
// What is intentionally NOT here yet (it's P2/P3): per-age layout strategies,
// roads, 2.5D building volumes, trade routes, and civ-edge markers.
//
// Rendering model: on each draw we allocate an *image.RGBA at (cols × rows*2)
// pixels, fill it (terrain + markers), and stream it via '▄' half-blocks where
// each terminal cell's background is the upper pixel and foreground the lower
// pixel. The image is cached and only regenerated when the cache key
// (age, width, height, building count, theme key) changes — so an idle redraw is
// cheap, but a theme switch (new theme key) or an age advance invalidates it.
package citymap

import (
	"image"
	"sync"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CityMap is the overlay widget renderer. It satisfies the OverlayManager widget
// contract via Build (returns the primitive) and Refresh (stores a fresh state
// snapshot). All rendering reads the stored snapshot under the mutex — Refresh
// only stores, it never touches the engine, honoring the lock rules.
type CityMap struct {
	mu    sync.Mutex
	state game.GameState

	// Cache of the last rendered image, plus the key that produced it.
	cachedImg      *image.RGBA
	cachedW        int
	cachedH        int
	cachedAge      string
	cachedBld      int
	cachedThemeKey string
}

// NewCityMap creates a CityMap renderer.
func NewCityMap() *CityMap { return &CityMap{} }

// Build returns the tview.Primitive that draws the map. The draw closure reads
// the latest stored snapshot each frame and (re)generates the image on demand.
func (c *CityMap) Build(state game.GameState) tview.Primitive {
	c.mu.Lock()
	c.state = state
	c.mu.Unlock()

	box := tview.NewBox().SetBorder(false)
	box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		if width <= 0 || height <= 0 {
			return x, y, width, height
		}

		img := c.imageFor(width, height)
		streamHalfBlocks(screen, img, x, y, width, height)
		return x, y, width, height
	})
	return box
}

// Refresh stores a fresh state snapshot. Safe to call from any goroutine; it does
// no rendering and acquires no engine locks.
func (c *CityMap) Refresh(state game.GameState) {
	c.mu.Lock()
	c.state = state
	c.mu.Unlock()
}

// imageFor returns the rendered image for a (width,height) terminal area,
// regenerating it only when the cache key changes. width/height are in terminal
// cells; the image is (width × height*2) pixels because each cell is one
// half-block (two stacked pixels).
func (c *CityMap) imageFor(width, height int) *image.RGBA {
	imgW := width
	imgH := height * 2

	c.mu.Lock()
	defer c.mu.Unlock()

	st := c.state
	bld := builtBuildingCount(st)
	themeKey := theme.Active().Key

	if c.cachedImg != nil &&
		c.cachedW == imgW && c.cachedH == imgH &&
		c.cachedAge == st.Age && c.cachedBld == bld &&
		c.cachedThemeKey == themeKey {
		return c.cachedImg
	}

	img := renderImage(st, imgW, imgH)
	c.cachedImg = img
	c.cachedW = imgW
	c.cachedH = imgH
	c.cachedAge = st.Age
	c.cachedBld = bld
	c.cachedThemeKey = themeKey
	return img
}

// renderImage composites the full map: procedural theme-tinted terrain, then the
// palace + building markers on top. It reads the active theme via the palette
// helpers, so the output reflects whatever theme is active at call time.
func renderImage(state game.GameState, w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	seed, hueShift := ageInfo(state.Age)
	pal := buildPalette(hueShift)

	drawTerrain(img, pal, seed)

	keys := builtBuildingKeys(state)
	markers := layoutMarkers(w, h, keys)
	drawMarkers(img, pal, markers)

	return img
}

// streamHalfBlocks writes img to the screen using '▄' half-block characters.
// Each terminal row maps to two pixel rows: the upper pixel becomes the cell
// background and the lower pixel the foreground, so one cell shows two stacked
// colors. termW/termH are in cells; the image must be at least termW×(termH*2)px.
func streamHalfBlocks(screen tcell.Screen, img *image.RGBA, offX, offY, termW, termH int) {
	b := img.Bounds()
	for row := 0; row < termH; row++ {
		upperY := row * 2
		lowerY := row*2 + 1
		for col := 0; col < termW; col++ {
			if col >= b.Dx() || lowerY >= b.Dy() {
				continue
			}
			upper := img.RGBAAt(b.Min.X+col, b.Min.Y+upperY)
			lower := img.RGBAAt(b.Min.X+col, b.Min.Y+lowerY)
			bg := tcell.NewRGBColor(int32(upper.R), int32(upper.G), int32(upper.B))
			fg := tcell.NewRGBColor(int32(lower.R), int32(lower.G), int32(lower.B))
			screen.SetContent(offX+col, offY+row, '▄', nil,
				tcell.StyleDefault.Background(bg).Foreground(fg))
		}
	}
}
