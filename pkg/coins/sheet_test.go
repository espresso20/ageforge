package coins_test

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/pkg/coins"
)

// TestWriteSheet writes one age's sheet into COINS_SHEET for key COINS_AGE; skipped unless both are set.
func TestWriteSheet(t *testing.T) {
	dir, key := os.Getenv("COINS_SHEET"), os.Getenv("COINS_AGE")
	if dir == "" || key == "" {
		t.Skip("COINS_SHEET and COINS_AGE not set")
	}
	name := strings.TrimSuffix(key, "_age")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	pic := func(level int, seed uint32, size int) *image.RGBA {
		return coins.Render(key, "", level, seed, size)
	}
	levels := sheet(1104, 304)
	for l := 1; l <= coins.Levels; l++ {
		put(levels, pic(l, 0, 128), pic(l, 0, 48), 16+(l-1)*144, 16)
	}
	put(levels, pic(coins.Legendary, 0, 208), pic(coins.Legendary, 0, 48), 880, 16)
	savePNG(t, filepath.Join(dir, name+"-levels.png"), levels)
	copies := sheet(880, 432)
	var big []*image.RGBA
	for s := uint32(1); s <= 12; s++ {
		b := pic(5, s, 128)
		big = append(big, b)
		put(copies, b, pic(5, s, 48), 16+int(s-1)%6*144, 16+int(s-1)/6*208)
	}
	savePNG(t, filepath.Join(dir, name+"-copies.png"), copies)
	var sb strings.Builder
	for l := 1; l < coins.Levels; l++ {
		fmt.Fprintf(&sb, "level %d to %d at 48 px: %.1f%%\n", l, l+1, 100*paintedDiffer(pic(l, 0, 48), pic(l+1, 0, 48)))
	}
	lo, sum, n := 1.0, 0.0, 0
	for i := range big {
		for j := i + 1; j < len(big); j++ {
			d := paintedDiffer(big[i], big[j])
			lo, sum, n = math.Min(lo, d), sum+d, n+1
		}
	}
	fmt.Fprintf(&sb, "copies at 128 px, %d pairs: minimum %.1f%%, mean %.1f%%\n", n, 100*lo, 100*sum/float64(n))
	if err := os.WriteFile(filepath.Join(dir, name+"-numbers.txt"), []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// paintedDiffer is the share of painted pixels (alpha > 0 in either) that differ.
func paintedDiffer(a, b *image.RGBA) float64 {
	diff, painted := 0, 0
	for i := 0; i < len(a.Pix); i += 4 {
		if a.Pix[i+3] == 0 && b.Pix[i+3] == 0 {
			continue
		}
		painted++
		if string(a.Pix[i:i+4]) != string(b.Pix[i:i+4]) {
			diff++
		}
	}
	return float64(diff) / float64(max(painted, 1))
}

func sheet(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0x0b, 0x0e, 0x14, 255}), image.Point{}, draw.Src)
	return img
}

func put(c *image.RGBA, big, small *image.RGBA, x, y int) {
	bw, bh := big.Bounds().Dx(), big.Bounds().Dy()
	draw.Draw(c, image.Rect(x, y, x+bw, y+bh), big, image.Point{}, draw.Over)
	sx, sy := x+(bw-48)/2, y+bh+16
	draw.Draw(c, image.Rect(sx, sy, sx+48, sy+48), small, image.Point{}, draw.Over)
}

func savePNG(t *testing.T, path string, img *image.RGBA) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
