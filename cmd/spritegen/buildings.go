// cmd/spritegen/buildings.go — thin wrapper delegating to pkg/sprites
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"github.com/espresso20/ageforge/pkg/sprites"
)

// GenerateAllBuildingSprites generates 3 variations of each building sprite
// at the current era and saves them as 16×16 PNGs.
func GenerateAllBuildingSprites(outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	total := 0
	for _, b := range sprites.AllBuildings {
		for v := 0; v < 3; v++ {
			pixels := sprites.GenerateBuildingSprite(b.Lineage, b.Tier, b.AgeKey, v)
			img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
			for row := 0; row < 16; row++ {
				for col := 0; col < 16; col++ {
					pv := pixels[row][col]
					if pv == 0 {
						img.SetNRGBA(col, row, color.NRGBA{0, 0, 0, 0})
					} else {
						img.SetNRGBA(col, row, color.NRGBA{uint8(pv >> 16), uint8(pv >> 8), uint8(pv), 255})
					}
				}
			}
			suffix := ""
			if v > 0 {
				suffix = fmt.Sprintf("_v%d", v+1)
			}
			fname := filepath.Join(outDir, b.Key+suffix+".png")
			if err := savePNG(fname, img); err != nil {
				return err
			}
		}
		total++
	}
	fmt.Printf("Generated %d building sprites (%d files) → %s\n", total, total*3, outDir)
	return nil
}
