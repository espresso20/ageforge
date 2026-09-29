package capture

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/gdamore/tcell/v2"
)

// PNG writes a screen as a colour thumbnail for review: each cell is 6x12
// pixels, blocks and shades drawn as themselves, other glyphs as a small
// mark in their ink. It shows composition and colour, not letterforms; the
// HTML captures are the real render.
func PNG(path string, s tcell.SimulationScreen) error {
	s.Show()
	cells, w, h := s.GetContents()
	const cw, ch = 6, 12
	img := image.NewRGBA(image.Rect(0, 0, w*cw, h*ch))
	rgb := func(c tcell.Color) color.RGBA {
		r, g, b := c.RGB()
		if r < 0 {
			return color.RGBA{0, 0, 0, 255}
		}
		return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			fg, bg, _ := c.Style.Decompose()
			r := ' '
			if len(c.Runes) > 0 {
				r = c.Runes[0]
			}
			f, b := rgb(fg), rgb(bg)
			for py := 0; py < ch; py++ {
				for px := 0; px < cw; px++ {
					col := b
					switch r {
					case ' ':
					case '█':
						col = f
					case '▀':
						if py < ch/2 {
							col = f
						}
					case '▄':
						if py >= ch/2 {
							col = f
						}
					case '▌':
						if px < cw/2 {
							col = f
						}
					case '▐':
						if px >= cw/2 {
							col = f
						}
					case '░':
						if (px+py)%4 == 0 {
							col = f
						}
					case '▒':
						if (px+py)%2 == 0 {
							col = f
						}
					case '▓':
						if (px+py)%4 != 0 {
							col = f
						}
					case '─', '═':
						if py == ch/2 || (r == '═' && py == ch/2+2) {
							col = f
						}
					case '│', '║', '╫':
						if px == cw/2 || (r != '│' && px == cw/2-2) {
							col = f
						}
					default:
						if px >= 1 && px < cw-1 && py >= 3 && py < ch-2 && (px+py)%2 == 0 {
							col = f
						}
					}
					img.SetRGBA(x*cw+px, y*ch+py, col)
				}
			}
		}
	}
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	return png.Encode(fh, img)
}
