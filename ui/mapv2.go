package ui

import (
	"encoding/binary"
	"hash/fnv"
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/espresso20/ageforge/game"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MapV2 renders a procedurally generated terrain map using half-block characters
// for 2x vertical resolution. Each terminal cell maps to 2 pixel rows via '▄'.
type MapV2 struct {
	mu    sync.Mutex
	state game.GameState
}

// NewMapV2 creates a new MapV2 instance.
func NewMapV2() *MapV2 { return &MapV2{} }

// Refresh updates the stored game state. Safe to call from any goroutine.
func (m *MapV2) Refresh(state game.GameState) {
	m.mu.Lock()
	m.state = state
	m.mu.Unlock()
}

// Build returns a tview.Primitive that renders the map using SetDrawFunc.
func (m *MapV2) Build(state game.GameState) tview.Primitive {
	m.mu.Lock()
	m.state = state
	m.mu.Unlock()

	box := tview.NewBox()
	box.SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
		m.mu.Lock()
		s := m.state
		m.mu.Unlock()

		if w <= 0 || h <= 0 {
			return x, y, w, h
		}

		// Each terminal cell = 1 pixel wide, 2 pixels tall (half-block technique)
		pixW := w
		pixH := h * 2
		img := generateMapV2Image(s, pixW, pixH)
		renderHalfBlock(screen, img, x, y, w, h)
		return x, y, w, h
	})
	return box
}

// renderHalfBlock renders an image.RGBA to the terminal using '▄' half-block chars.
// Upper half of cell = background color, lower half = foreground color of '▄'.
// This gives 2x vertical resolution compared to normal character rendering.
func renderHalfBlock(screen tcell.Screen, img *image.RGBA, offX, offY, termW, termH int) {
	for row := 0; row < termH; row++ {
		for col := 0; col < termW; col++ {
			upperY := row * 2
			lowerY := row*2 + 1

			upperC := img.RGBAAt(col, upperY)
			lowerC := img.RGBAAt(col, lowerY)

			bgColor := tcell.NewRGBColor(int32(upperC.R), int32(upperC.G), int32(upperC.B))
			fgColor := tcell.NewRGBColor(int32(lowerC.R), int32(lowerC.G), int32(lowerC.B))

			style := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
			screen.SetContent(offX+col, offY+row, '▄', nil, style)
		}
	}
}

// generateMapV2Image produces the full map image for the given state and pixel dimensions.
func generateMapV2Image(state game.GameState, pixW, pixH int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, pixW, pixH))

	seed := mapV2Seed(state)
	era := ageEraV2(state.Age)
	pal := getV2Palette(era)

	drawV2Terrain(img, seed, pal, pixW, pixH)
	drawV2Water(img, seed, pal, pixW, pixH)
	drawV2Vegetation(img, seed, pal, era, pixW, pixH)
	drawV2Roads(img, pal, era, pixW, pixH)
	drawV2Buildings(img, state, pal, era, pixW, pixH)
	drawV2Label(img, state, era, pixW, pixH)

	return img
}

// ── Noise ──────────────────────────────────────────────────────────────────

func v2Hash(x, y int, seed uint64) float64 {
	h := fnv.New64a()
	var buf [24]byte
	binary.LittleEndian.PutUint64(buf[0:], seed)
	binary.LittleEndian.PutUint64(buf[8:], uint64(x+1000000))
	binary.LittleEndian.PutUint64(buf[16:], uint64(y+1000000))
	h.Write(buf[:])
	return float64(h.Sum64()%10000) / 10000.0
}

func v2Noise(x, y int, seed uint64, scale float64) float64 {
	fx := float64(x) / scale
	fy := float64(y) / scale

	x0 := int(math.Floor(fx))
	y0 := int(math.Floor(fy))
	x1 := x0 + 1
	y1 := y0 + 1

	tx := fx - math.Floor(fx)
	ty := fy - math.Floor(fy)

	// Smooth step
	tx = tx * tx * (3 - 2*tx)
	ty = ty * ty * (3 - 2*ty)

	v00 := v2Hash(x0, y0, seed)
	v10 := v2Hash(x1, y0, seed)
	v01 := v2Hash(x0, y1, seed)
	v11 := v2Hash(x1, y1, seed)

	return v00*(1-tx)*(1-ty) + v10*tx*(1-ty) + v01*(1-tx)*ty + v11*tx*ty
}

// v2FBM returns multi-octave fractal Brownian motion noise in [0, 1].
func v2FBM(x, y int, seed uint64) float64 {
	n := v2Noise(x, y, seed, 40.0) * 0.5
	n += v2Noise(x, y, seed+100, 20.0) * 0.3
	n += v2Noise(x, y, seed+200, 10.0) * 0.2
	return n
}

// ── Palette ────────────────────────────────────────────────────────────────

// V2Palette holds all terrain and feature colors for one era.
type V2Palette struct {
	Grassland, Plains, Forest, Hills, Mountains color.RGBA
	Ocean, Coast, Desert, Snow                  color.RGBA
	Road                                        color.RGBA
	City                                        color.RGBA
	Wonder                                      color.RGBA
	Tree                                        color.RGBA
}

func rgba(r, g, b uint8) color.RGBA { return color.RGBA{R: r, G: g, B: b, A: 255} }

func getV2Palette(era int) V2Palette {
	switch era {
	case 0: // primitive
		return V2Palette{
			Grassland: rgba(58, 107, 42), Plains: rgba(110, 130, 58),
			Forest: rgba(30, 74, 16), Hills: rgba(90, 106, 48),
			Mountains: rgba(58, 46, 32), Ocean: rgba(10, 42, 90),
			Coast: rgba(26, 74, 122), Desert: rgba(160, 128, 64),
			Snow: rgba(192, 200, 204), Road: rgba(140, 120, 80),
			City: rgba(200, 160, 80), Wonder: rgba(255, 215, 0),
			Tree: rgba(20, 56, 10),
		}
	case 1, 2: // ancient/medieval
		return V2Palette{
			Grassland: rgba(50, 95, 38), Plains: rgba(100, 118, 50),
			Forest: rgba(26, 66, 14), Hills: rgba(82, 96, 42),
			Mountains: rgba(64, 52, 38), Ocean: rgba(8, 38, 82),
			Coast: rgba(22, 66, 110), Desert: rgba(148, 118, 58),
			Snow: rgba(180, 188, 192), Road: rgba(120, 100, 68),
			City: rgba(180, 148, 72), Wonder: rgba(255, 215, 0),
			Tree: rgba(18, 50, 10),
		}
	case 3: // industrial
		return V2Palette{
			Grassland: rgba(62, 86, 44), Plains: rgba(100, 110, 56),
			Forest: rgba(34, 60, 20), Hills: rgba(78, 88, 50),
			Mountains: rgba(72, 64, 52), Ocean: rgba(12, 40, 78),
			Coast: rgba(24, 62, 106), Desert: rgba(148, 120, 60),
			Snow: rgba(170, 175, 178), Road: rgba(100, 92, 80),
			City: rgba(160, 148, 130), Wonder: rgba(255, 215, 0),
			Tree: rgba(28, 48, 16),
		}
	case 4, 5: // modern
		return V2Palette{
			Grassland: rgba(52, 100, 44), Plains: rgba(108, 124, 56),
			Forest: rgba(28, 68, 16), Hills: rgba(84, 98, 48),
			Mountains: rgba(68, 58, 44), Ocean: rgba(8, 36, 88),
			Coast: rgba(20, 68, 116), Desert: rgba(156, 126, 62),
			Snow: rgba(188, 196, 200), Road: rgba(90, 90, 90),
			City: rgba(140, 160, 180), Wonder: rgba(255, 215, 0),
			Tree: rgba(22, 56, 12),
		}
	case 6, 7: // digital/cyberpunk
		return V2Palette{
			Grassland: rgba(20, 42, 20), Plains: rgba(28, 50, 28),
			Forest: rgba(10, 30, 10), Hills: rgba(22, 36, 22),
			Mountains: rgba(24, 24, 28), Ocean: rgba(4, 14, 40),
			Coast: rgba(8, 24, 60), Desert: rgba(50, 40, 18),
			Snow: rgba(60, 70, 80), Road: rgba(0, 200, 200),
			City: rgba(0, 220, 180), Wonder: rgba(255, 215, 0),
			Tree: rgba(0, 60, 0),
		}
	default: // space/cosmic
		return V2Palette{
			Grassland: rgba(12, 14, 24), Plains: rgba(16, 18, 30),
			Forest: rgba(8, 10, 18), Hills: rgba(14, 16, 26),
			Mountains: rgba(30, 28, 40), Ocean: rgba(4, 6, 16),
			Coast: rgba(8, 10, 24), Desert: rgba(20, 16, 10),
			Snow: rgba(40, 44, 60), Road: rgba(60, 80, 120),
			City: rgba(80, 120, 200), Wonder: rgba(255, 215, 0),
			Tree: rgba(0, 0, 0),
		}
	}
}

// ── Draw stages ────────────────────────────────────────────────────────────

func drawV2Terrain(img *image.RGBA, seed uint64, pal V2Palette, pixW, pixH int) {
	cx, cy := pixW/2, pixH/2

	for py := 0; py < pixH; py++ {
		for px := 0; px < pixW; px++ {
			n := v2FBM(px, py, seed)

			dx := float64(px-cx) / float64(pixW/2)
			dy := float64(py-cy) / float64(pixH/2)
			dist := math.Sqrt(dx*dx + dy*dy)

			var c color.RGBA
			switch {
			case dist > 0.85:
				c = pal.Ocean
			case dist > 0.78:
				c = pal.Coast
			case n > 0.72:
				c = pal.Mountains
			case n > 0.60:
				c = pal.Hills
			case n > 0.45:
				c = pal.Forest
			case n > 0.30:
				c = pal.Grassland
			default:
				c = pal.Plains
			}
			img.SetRGBA(px, py, c)
		}
	}
}

func drawV2Water(img *image.RGBA, seed uint64, pal V2Palette, pixW, pixH int) {
	// Add subtle wave variation on ocean pixels by slightly shifting blue component
	for py := 0; py < pixH; py++ {
		for px := 0; px < pixW; px++ {
			c := img.RGBAAt(px, py)
			if c == pal.Ocean {
				wave := v2Hash(px/4, py/4, seed+999)
				shift := uint8(wave * 20)
				img.SetRGBA(px, py, color.RGBA{
					R: c.R,
					G: c.G + shift/3,
					B: c.B + shift,
					A: 255,
				})
			}
		}
	}
}

func drawV2Vegetation(img *image.RGBA, seed uint64, pal V2Palette, era, pixW, pixH int) {
	if era >= 6 {
		return // no vegetation in digital+ eras
	}
	density := 0.35 - float64(era)*0.04

	for py := 0; py < pixH; py++ {
		for px := 0; px < pixW; px++ {
			existing := img.RGBAAt(px, py)
			if existing == pal.Forest {
				n := v2Hash(px*3, py*3, seed+500)
				if n < density {
					darkened := color.RGBA{
						R: uint8(float64(existing.R) * 0.8),
						G: uint8(float64(existing.G) * 0.85),
						B: uint8(float64(existing.B) * 0.8),
						A: 255,
					}
					img.SetRGBA(px, py, darkened)
				}
			}
		}
	}
}

func drawV2Roads(img *image.RGBA, pal V2Palette, era, pixW, pixH int) {
	cx, cy := pixW/2, pixH/2
	roadW := 1 + era/3
	if roadW > 3 {
		roadW = 3
	}

	// Horizontal road
	for px := cx - pixW/4; px < cx+pixW/4; px++ {
		for dw := -roadW; dw <= roadW; dw++ {
			if py := cy + dw; py >= 0 && py < pixH {
				img.SetRGBA(px, py, pal.Road)
			}
		}
	}
	// Vertical road
	for py := cy - pixH/4; py < cy+pixH/4; py++ {
		for dw := -roadW; dw <= roadW; dw++ {
			if px := cx + dw; px >= 0 && px < pixW {
				img.SetRGBA(px, py, pal.Road)
			}
		}
	}
}

func drawV2Buildings(img *image.RGBA, state game.GameState, pal V2Palette, era, pixW, pixH int) {
	cx, cy := pixW/2, pixH/2

	builtCount := 0
	for _, bs := range state.Buildings {
		if bs.Count > 0 {
			builtCount++
		}
	}
	cityR := 8 + builtCount*2
	if cityR > pixW/3 {
		cityR = pixW / 3
	}

	// City center marker
	sz := 4 + era
	if sz > 10 {
		sz = 10
	}
	for py := cy - sz; py <= cy+sz; py++ {
		for px := cx - sz; px <= cx+sz; px++ {
			if px >= 0 && px < pixW && py >= 0 && py < pixH {
				img.SetRGBA(px, py, pal.City)
			}
		}
	}

	bSize := 3 + era/2
	if bSize > 6 {
		bSize = 6
	}
	spacing := bSize * 3

	slot := 0
	for _, bs := range state.Buildings {
		if bs.Count <= 0 {
			continue
		}

		angle := float64(slot) * 0.7 // golden angle approx
		radius := float64(spacing) + float64(slot/6)*float64(spacing)
		if radius > float64(cityR) {
			radius = float64(cityR)
		}

		bx := cx + int(math.Cos(angle)*radius)
		by := cy + int(math.Sin(angle)*radius)

		c := domainColorV2(bs.WorkerDomain)

		for py := by - bSize/2; py <= by+bSize/2; py++ {
			for px := bx - bSize/2; px <= bx+bSize/2; px++ {
				if px >= 0 && px < pixW && py >= 0 && py < pixH {
					img.SetRGBA(px, py, c)
				}
			}
		}
		slot++
	}
}

func drawV2Label(img *image.RGBA, state game.GameState, era, pixW, pixH int) {
	// Thin status bar at the very bottom (last 2 pixel rows)
	barColor := rgba(10, 10, 10)
	for px := 0; px < pixW; px++ {
		img.SetRGBA(px, pixH-2, barColor)
		img.SetRGBA(px, pixH-1, barColor)
	}
}

// ── Domain colors ──────────────────────────────────────────────────────────

func domainColorV2(domain string) color.RGBA {
	switch domain {
	case "food":
		return rgba(212, 160, 80)
	case "lumber":
		return rgba(80, 160, 80)
	case "masonry":
		return rgba(160, 160, 160)
	case "metallurgy":
		return rgba(224, 112, 48)
	case "energy":
		return rgba(240, 208, 32)
	case "military":
		return rgba(224, 64, 64)
	case "knowledge":
		return rgba(96, 176, 240)
	case "faith":
		return rgba(240, 240, 192)
	case "trade":
		return rgba(240, 192, 64)
	case "engineering":
		return rgba(64, 208, 208)
	case "hacker":
		return rgba(64, 240, 64)
	case "astronaut":
		return rgba(240, 240, 240)
	default:
		return rgba(180, 180, 180)
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────

func mapV2Seed(state game.GameState) uint64 {
	h := fnv.New64a()
	h.Write([]byte("mapv2"))
	return h.Sum64()
}

func ageEraV2(ageKey string) int {
	switch ageKey {
	case "primitive_age", "stone_age":
		return 0
	case "bronze_age", "iron_age", "classical_age":
		return 1
	case "medieval_age", "renaissance_age", "colonial_age":
		return 2
	case "industrial_age", "victorian_age", "electric_age":
		return 3
	case "atomic_age", "modern_age", "information_age":
		return 4
	case "digital_age":
		return 5
	case "cyberpunk_age", "fusion_age":
		return 6
	case "space_age", "interstellar_age":
		return 7
	default:
		return 8
	}
}
