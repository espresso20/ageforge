// cmd/spritegen/main.go — generates 8×8 pixel art domain icons + sprite sheet + preview
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// rgba converts a packed 0xRRGGBB value to color.RGBA (alpha=255), or transparent if 0.
func rgba(v uint32) color.RGBA {
	if v == 0 {
		return color.RGBA{0, 0, 0, 0}
	}
	return color.RGBA{
		R: uint8((v >> 16) & 0xff),
		G: uint8((v >> 8) & 0xff),
		B: uint8(v & 0xff),
		A: 255,
	}
}

type sprite struct {
	name   string
	pixels [8][8]uint32
}

// Color palette aliases used in sprite definitions.
const (
	// food
	wheat  = 0xe8c840
	stem   = 0x6a4a20

	// wood
	dkGreen = 0x2d6b1a
	trunk   = 0x6b3a1a

	// stone
	gray = 0x7a7a7a
	hiGray = 0xaaaaaa

	// iron
	dkIron = 0x4a4a5a

	// knowledge
	brown = 0x8b4513
	pages = 0xf0f0e0

	// military
	silver = 0xc0c0c0

	// faith / trade
	gold = 0xd4a017

	// culture
	warmWood = 0xc87941

	// energy
	electric = 0xffe040

	// metallurgy
	metalGray = 0x888888

	// advanced
	cyan = 0x40c0ff

	// city_center
	stone2 = 0x8a7a6a

	// wonder
	sandstone = 0xc8a840
	sandHi    = 0xe8c860

	// storage
	woodBrown = 0x8b5a2b
	ltWood    = 0xc87941
)

var sprites = []sprite{
	{
		name: "food",
		pixels: [8][8]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, wheat, 0, wheat, 0, 0, 0},
			{0, wheat, wheat, wheat, wheat, wheat, 0, 0},
			{0, 0, wheat, 0, wheat, 0, 0, 0},
			{0, 0, stem, 0, stem, 0, 0, 0},
			{0, 0, stem, 0, stem, 0, 0, 0},
			{0, 0, stem, 0, stem, 0, 0, 0},
			{0, stem, stem, 0, stem, stem, 0, 0},
		},
	},
	{
		name: "wood",
		pixels: [8][8]uint32{
			{0, 0, dkGreen, 0, 0, 0, 0, 0},
			{0, dkGreen, dkGreen, dkGreen, 0, 0, 0, 0},
			{dkGreen, dkGreen, dkGreen, dkGreen, dkGreen, 0, 0, 0},
			{0, dkGreen, dkGreen, dkGreen, dkGreen, dkGreen, 0, 0},
			{0, 0, dkGreen, dkGreen, dkGreen, 0, 0, 0},
			{0, 0, 0, trunk, 0, 0, 0, 0},
			{0, 0, 0, trunk, 0, 0, 0, 0},
			{0, 0, trunk, trunk, trunk, 0, 0, 0},
		},
	},
	{
		name: "stone",
		pixels: [8][8]uint32{
			{0, 0, gray, gray, gray, 0, 0, 0},
			{0, gray, hiGray, hiGray, gray, gray, 0, 0},
			{gray, gray, hiGray, gray, gray, gray, gray, 0},
			{gray, gray, gray, gray, gray, gray, gray, 0},
			{gray, gray, gray, gray, gray, gray, gray, 0},
			{0, gray, gray, gray, gray, gray, 0, 0},
			{0, 0, gray, gray, gray, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	{
		name: "iron",
		pixels: [8][8]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, dkIron, dkIron, 0, dkIron, dkIron, 0, 0},
			{0, dkIron, dkIron, dkIron, dkIron, dkIron, 0, 0},
			{0, 0, dkIron, dkIron, dkIron, 0, 0, 0},
			{0, dkIron, dkIron, dkIron, dkIron, dkIron, 0, 0},
			{0, dkIron, dkIron, dkIron, dkIron, dkIron, 0, 0},
			{dkIron, dkIron, 0, 0, 0, dkIron, dkIron, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	{
		name: "knowledge",
		pixels: [8][8]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, brown, brown, brown, brown, brown, 0, 0},
			{0, brown, pages, pages, pages, brown, 0, 0},
			{0, brown, pages, pages, pages, brown, 0, 0},
			{0, brown, pages, pages, pages, brown, 0, 0},
			{0, brown, brown, brown, brown, brown, 0, 0},
			{0, brown, brown, brown, brown, brown, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	{
		name: "military",
		pixels: [8][8]uint32{
			{silver, 0, 0, 0, 0, 0, 0, silver},
			{0, silver, 0, 0, 0, 0, silver, 0},
			{0, 0, silver, 0, 0, silver, 0, 0},
			{0, 0, 0, silver, silver, 0, 0, 0},
			{0, 0, silver, 0, 0, silver, 0, 0},
			{0, silver, 0, 0, 0, 0, silver, 0},
			{silver, 0, 0, 0, 0, 0, 0, silver},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	{
		name: "faith",
		pixels: [8][8]uint32{
			{0, 0, 0, gold, 0, 0, 0, 0},
			{0, 0, 0, gold, 0, 0, 0, 0},
			{gold, gold, gold, gold, gold, gold, gold, 0},
			{0, 0, 0, gold, 0, 0, 0, 0},
			{0, 0, 0, gold, 0, 0, 0, 0},
			{0, 0, 0, gold, 0, 0, 0, 0},
			{0, 0, 0, gold, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	{
		name: "culture",
		pixels: [8][8]uint32{
			{0, warmWood, warmWood, warmWood, 0, 0, 0, 0},
			{warmWood, 0, 0, 0, warmWood, 0, 0, 0},
			{warmWood, 0, warmWood, 0, warmWood, warmWood, 0, 0},
			{warmWood, 0, warmWood, 0, warmWood, warmWood, 0, 0},
			{warmWood, 0, warmWood, 0, warmWood, warmWood, 0, 0},
			{0, warmWood, warmWood, warmWood, warmWood, 0, 0, 0},
			{0, 0, warmWood, 0, 0, 0, 0, 0},
			{0, warmWood, warmWood, warmWood, 0, 0, 0, 0},
		},
	},
	{
		name: "trade",
		pixels: [8][8]uint32{
			{0, 0, 0, gold, 0, 0, 0, 0},
			{0, 0, 0, gold, 0, 0, 0, 0},
			{0, gold, 0, gold, 0, gold, 0, 0},
			{gold, gold, gold, gold, gold, gold, gold, 0},
			{gold, 0, 0, 0, 0, 0, gold, 0},
			{0, gold, 0, 0, 0, gold, 0, 0},
			{0, 0, gold, gold, gold, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	{
		name: "energy",
		pixels: [8][8]uint32{
			{0, 0, 0, electric, electric, electric, 0, 0},
			{0, 0, electric, electric, 0, 0, 0, 0},
			{0, electric, electric, 0, 0, 0, 0, 0},
			{electric, electric, electric, electric, electric, 0, 0, 0},
			{0, 0, 0, electric, electric, 0, 0, 0},
			{0, 0, 0, 0, electric, electric, 0, 0},
			{0, 0, 0, 0, 0, electric, electric, 0},
			{0, 0, 0, 0, 0, 0, electric, 0},
		},
	},
	{
		name: "metallurgy",
		pixels: [8][8]uint32{
			{0, 0, metalGray, 0, metalGray, 0, 0, 0},
			{0, metalGray, metalGray, metalGray, metalGray, metalGray, 0, 0},
			{metalGray, metalGray, 0, metalGray, 0, metalGray, metalGray, 0},
			{0, metalGray, metalGray, metalGray, metalGray, metalGray, 0, 0},
			{metalGray, metalGray, 0, metalGray, 0, metalGray, metalGray, 0},
			{0, metalGray, metalGray, metalGray, metalGray, metalGray, 0, 0},
			{0, 0, metalGray, 0, metalGray, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	{
		name: "advanced",
		pixels: [8][8]uint32{
			{0, 0, 0, cyan, 0, 0, 0, 0},
			{0, 0, cyan, cyan, cyan, 0, 0, 0},
			{0, cyan, 0, cyan, 0, cyan, 0, 0},
			{cyan, cyan, 0, 0, 0, cyan, cyan, 0},
			{0, cyan, 0, cyan, 0, cyan, 0, 0},
			{0, 0, cyan, cyan, cyan, 0, 0, 0},
			{0, 0, 0, cyan, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	{
		name: "city_center",
		pixels: [8][8]uint32{
			{0, stone2, 0, stone2, 0, stone2, 0, 0},
			{0, stone2, stone2, stone2, stone2, stone2, 0, 0},
			{0, stone2, stone2, stone2, stone2, stone2, 0, 0},
			{0, stone2, 0, stone2, 0, stone2, 0, 0},
			{stone2, stone2, stone2, stone2, stone2, stone2, stone2, 0},
			{stone2, 0, stone2, 0, stone2, 0, stone2, 0},
			{stone2, stone2, stone2, stone2, stone2, stone2, stone2, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	{
		name: "wonder",
		pixels: [8][8]uint32{
			{0, 0, 0, sandstone, 0, 0, 0, 0},
			{0, 0, sandstone, sandHi, sandstone, 0, 0, 0},
			{0, 0, sandstone, sandstone, sandstone, 0, 0, 0},
			{0, sandstone, sandstone, sandstone, sandstone, sandstone, 0, 0},
			{0, sandstone, sandstone, sandstone, sandstone, sandstone, 0, 0},
			{sandstone, sandstone, sandstone, sandstone, sandstone, sandstone, sandstone, 0},
			{sandstone, sandstone, sandstone, sandstone, sandstone, sandstone, sandstone, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	{
		name: "storage",
		pixels: [8][8]uint32{
			{0, woodBrown, woodBrown, woodBrown, woodBrown, woodBrown, 0, 0},
			{woodBrown, ltWood, ltWood, ltWood, ltWood, ltWood, woodBrown, 0},
			{woodBrown, woodBrown, woodBrown, woodBrown, woodBrown, woodBrown, woodBrown, 0},
			{woodBrown, 0, 0, 0, 0, 0, woodBrown, 0},
			{woodBrown, 0, woodBrown, woodBrown, woodBrown, 0, woodBrown, 0},
			{woodBrown, 0, 0, 0, 0, 0, woodBrown, 0},
			{woodBrown, woodBrown, woodBrown, woodBrown, woodBrown, woodBrown, woodBrown, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
}

func spriteToImage(s sprite) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			c := rgba(s.pixels[y][x])
			img.SetNRGBA(x, y, color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A})
		}
	}
	return img
}

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func buildSpriteSheet(imgs []*image.NRGBA) *image.NRGBA {
	n := len(imgs)
	sheet := image.NewNRGBA(image.Rect(0, 0, n*8, 8))
	for i, img := range imgs {
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				sheet.SetNRGBA(i*8+x, y, img.NRGBAAt(x, y))
			}
		}
	}
	return sheet
}

const (
	previewScale  = 8  // each pixel becomes 8×8
	previewPerRow = 5
	previewGap    = 4
	spriteDisplay = 8 * previewScale // 64px
)

func buildPreview(imgs []*image.NRGBA) *image.NRGBA {
	n := len(imgs)
	rows := (n + previewPerRow - 1) / previewPerRow

	cellW := spriteDisplay + previewGap
	cellH := spriteDisplay + previewGap

	totalW := previewPerRow*cellW - previewGap
	totalH := rows*cellH - previewGap

	preview := image.NewNRGBA(image.Rect(0, 0, totalW, totalH))

	// Fill background transparent (default for NRGBA).
	for i, img := range imgs {
		col := i % previewPerRow
		row := i / previewPerRow
		offX := col * cellW
		offY := row * cellH

		// Scale 8× by repeating pixels.
		for py := 0; py < 8; py++ {
			for px := 0; px < 8; px++ {
				c := img.NRGBAAt(px, py)
				for sy := 0; sy < previewScale; sy++ {
					for sx := 0; sx < previewScale; sx++ {
						preview.SetNRGBA(offX+px*previewScale+sx, offY+py*previewScale+sy, c)
					}
				}
			}
		}
	}
	return preview
}

func main() {
	outDir := "assets/sprites"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output dir: %v\n", err)
		os.Exit(1)
	}

	imgs := make([]*image.NRGBA, len(sprites))

	for i, s := range sprites {
		img := spriteToImage(s)
		imgs[i] = img
		path := fmt.Sprintf("%s/%s.png", outDir, s.name)
		if err := savePNG(path, img); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save %s: %v\n", path, err)
			os.Exit(1)
		}
	}

	// Sprite sheet: all 15 sprites in a single row (120×8).
	sheet := buildSpriteSheet(imgs)
	if err := savePNG(outDir+"/spritesheet.png", sheet); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save spritesheet: %v\n", err)
		os.Exit(1)
	}

	// Preview: 5 per row, scaled 8×, with gaps.
	preview := buildPreview(imgs)
	if err := savePNG(outDir+"/preview.png", preview); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save preview: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %d sprites → %s/\n", len(sprites), outDir)
	fmt.Printf("  spritesheet.png  %dx8 px\n", len(sprites)*8)
	fmt.Printf("  preview.png      scaled 8× grid (%d per row)\n", previewPerRow)
	for _, s := range sprites {
		fmt.Printf("  %s.png\n", s.name)
	}
}
