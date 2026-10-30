package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui"
)

func main() {
	outputPath := flag.String("o", "docs/banner.png", "output PNG path")
	cellWidth := flag.Int("w", 10, "cell width in pixels")
	cellHeight := flag.Int("h", 20, "cell height in pixels")
	flag.Parse()

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Generate the banner grid
	grid, width, height := generateBannerGrid(*cellWidth, *cellHeight)

	// Render the grid to PNG
	img := gridToImage(grid, width, height, *cellWidth, *cellHeight)

	// Save the PNG
	f, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}

	fmt.Printf("Banner saved to %s (%d×%d)\n", *outputPath, img.Bounds().Dx(), img.Bounds().Dy())
}

// generateBannerGrid creates the menu banner grid with stars, wordmark, and tagline.
func generateBannerGrid(cellWidth, cellHeight int) (*ui.MGrid, int, int) {
	// Dimensions: aim for ~140 cells wide (about 1400 px at 10 px/cell)
	gridWidth := 140
	// Height: stars + wordmark + tagline + margins
	// wordmark is 7 cells tall, with some extra for stars and margin
	gridHeight := 30

	grid := ui.NewMGrid(gridWidth, gridHeight)

	// Get the Forge theme
	forgeTheme := theme.Forge
	pal := ui.NewMenuPalette(forgeTheme)

	// Create a seeded random generator for stars
	rnd := ui.NewMenuRand(0x9e3779b97f4a7c15)

	// Draw stars at the top
	starCount := gridWidth * 45 / 100
	if starCount > 0 {
		var stars []ui.MenuStar
		for i := 0; i < starCount; i++ {
			st := ui.MenuStar{
				X:     int(rnd.F() * float64(gridWidth)),
				Y:     int(rnd.F() * float64(6)), // Top 6 rows
				Phase: rnd.F() * 6.28,
				R:     '·',
			}
			if rnd.F() < 0.3 {
				st.R = '∙'
			}
			stars = append(stars, st)
		}
		ui.DrawMenuStars(grid, pal, stars, 0)
	}

	// Draw the wordmark
	// Position it with some margin from the top
	wordmarkY := 8
	wordmarkX := (gridWidth - (ui.MenuWordW * 1)) / 2 // Center horizontally with sx=1
	ui.DrawWordmark(grid, pal, wordmarkX, wordmarkY, 1, 0, 0, 0.5, false)

	// Draw the tagline below the wordmark
	tagline := "twenty-two ages, one terminal"
	taglineY := wordmarkY + ui.MenuWordH + 2 // After wordmark with a gap
	taglineX := (gridWidth - len(tagline)) / 2
	grid.Text(taglineX, taglineY, tagline, pal.Ink, false)

	// Draw stars at the bottom
	bottomStarCount := gridWidth * 45 / 100
	if bottomStarCount > 0 {
		var stars []ui.MenuStar
		for i := 0; i < bottomStarCount; i++ {
			st := ui.MenuStar{
				X:     int(rnd.F() * float64(gridWidth)),
				Y:     gridHeight - 6 + int(rnd.F()*float64(6)), // Bottom 6 rows
				Phase: rnd.F() * 6.28,
				R:     '·',
			}
			if rnd.F() < 0.3 {
				st.R = '∙'
			}
			stars = append(stars, st)
		}
		ui.DrawMenuStars(grid, pal, stars, 0)
	}

	return grid, gridWidth, gridHeight
}

// gridToImage converts an mGrid to an image.Image.
func gridToImage(grid *ui.MGrid, gridWidth, gridHeight, cellWidth, cellHeight int) image.Image {
	imgWidth := gridWidth * cellWidth
	imgHeight := gridHeight * cellHeight

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	// Get the Forge theme for proper background color
	forgeTheme := theme.Forge
	pal := ui.NewMenuPalette(forgeTheme)
	bgColor := tcellColorToImageColor(pal.Bg)

	// Fill with background color
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Draw each cell
	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {
			cell := grid.GetCell(x, y)
			if cell == nil {
				continue
			}

			// Get the cell's foreground color
			fgColor := tcellColorToImageColor(cell.Fg)

			// Draw a colored rectangle for this cell
			cellRect := image.Rect(x*cellWidth, y*cellHeight, (x+1)*cellWidth, (y+1)*cellHeight)
			draw.Draw(img, cellRect, &image.Uniform{fgColor}, image.Point{}, draw.Src)
		}
	}

	return img
}

// tcellColorToImageColor converts a tcell.Color to an image.color.Color.
func tcellColorToImageColor(c tcell.Color) color.Color {
	r, g, b := c.RGB()
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}
