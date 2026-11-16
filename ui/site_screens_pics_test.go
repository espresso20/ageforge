//go:build mapcapture

package ui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// siteTestCells is a blank w×h grid with every cell owned by no picture.
func siteTestCells(w, h int) []simCell {
	return make([]simCell, w*h)
}

// siteTestOwn marks the cells of r as shown by picture index i (its Pic is i+1).
func siteTestOwn(cells []simCell, w int, r image.Rectangle, i int) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			cells[y*w+x].Pic = uint16(i + 1)
		}
	}
}

// siteTestEncode encodes cells and pics with the theme's colours and decodes
// the file it wrote.
func siteTestEncode(t *testing.T, cells []simCell, pics []kit.Pic, w, h int) siteScreenFile {
	t.Helper()
	raw, _, err := encodeSiteScreen(cells, pics, w, h, theme.Of(theme.RoleText), theme.Of(theme.RoleBackground), false)
	if err != nil {
		t.Fatal(err)
	}
	var f siteScreenFile
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	return f
}

// siteTestImage decodes the file's image i.
func siteTestImage(t *testing.T, f siteScreenFile, i int) image.Image {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(f.Images[i])
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return img
}

// Two pictures of the same image and one of another give three pictures in
// the file and two images.
func TestSiteScreenPicturesShareImages(t *testing.T) {
	const w, h = 12, 2
	red := image.NewRGBA(image.Rect(0, 0, 2, 2))
	blue := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for i := range red.Pix {
		red.Pix[i], blue.Pix[i] = 0, 0
	}
	for _, p := range []int{0, 1, 2, 3} {
		red.Pix[p*4+0], red.Pix[p*4+3] = 200, 255
		blue.Pix[p*4+2], blue.Pix[p*4+3] = 200, 255
	}
	sameRed := image.NewRGBA(red.Rect)
	copy(sameRed.Pix, red.Pix)
	pics := []kit.Pic{
		{Rect: image.Rect(0, 0, 2, 2), Img: red},
		{Rect: image.Rect(4, 0, 6, 2), Img: sameRed},
		{Rect: image.Rect(8, 0, 10, 2), Img: blue},
	}
	cells := siteTestCells(w, h)
	for i, p := range pics {
		siteTestOwn(cells, w, p.Rect, i)
	}
	f := siteTestEncode(t, cells, pics, w, h)
	if len(f.Pics) != 3 {
		t.Fatalf("got %d pics, want 3", len(f.Pics))
	}
	if len(f.Images) != 2 {
		t.Fatalf("got %d images, want 2", len(f.Images))
	}
	if f.Pics[0].I != f.Pics[1].I || f.Pics[0].I == f.Pics[2].I {
		t.Fatalf("pictures share images wrongly: %+v", f.Pics)
	}
}

// A 4×2-cell picture of a 2×-scaled 6×3 art is stored as the art at 6×3.
func TestSiteScreenPictureReducesToArtSize(t *testing.T) {
	const w, h = 4, 2
	art := image.NewRGBA(image.Rect(0, 0, 6, 3))
	dark := color.RGBA{10, 20, 30, 255}
	light := color.RGBA{240, 230, 220, 255}
	// A checkerboard: no 3×3 or larger block is one colour, so the reduction
	// stops at the art's own 2×2 blocks.
	for y := 0; y < 3; y++ {
		for x := 0; x < 6; x++ {
			c := dark
			if (x+y)%2 == 0 {
				c = light
			}
			art.SetRGBA(x, y, c)
		}
	}
	scaled := image.NewRGBA(image.Rect(0, 0, 12, 6))
	for y := 0; y < 6; y++ {
		for x := 0; x < 12; x++ {
			scaled.SetRGBA(x, y, art.RGBAAt(x/2, y/2))
		}
	}
	pics := []kit.Pic{{Rect: image.Rect(0, 0, 4, 2), Img: scaled}}
	cells := siteTestCells(w, h)
	siteTestOwn(cells, w, pics[0].Rect, 0)
	f := siteTestEncode(t, cells, pics, w, h)
	if len(f.Images) != 1 {
		t.Fatalf("got %d images, want 1", len(f.Images))
	}
	got := siteTestImage(t, f, 0)
	if got.Bounds().Dx() != 6 || got.Bounds().Dy() != 3 {
		t.Fatalf("image is %v, want 6x3", got.Bounds().Size())
	}
	for y := 0; y < 3; y++ {
		for x := 0; x < 6; x++ {
			r, g, b, a := got.At(x, y).RGBA()
			wr, wg, wb, wa := art.At(x, y).RGBA()
			if r != wr || g != wg || b != wb || a != wa {
				t.Fatalf("pixel (%d,%d) differs from the art", x, y)
			}
		}
	}
}

// A picture with a cell taken back by other ink is an error, and the error
// names the picture's rectangle.
func TestSiteScreenPictureWithDrawnOverCellFails(t *testing.T) {
	const w, h = 4, 2
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	pics := []kit.Pic{{Rect: image.Rect(0, 0, 4, 2), Img: img}}
	cells := siteTestCells(w, h)
	siteTestOwn(cells, w, pics[0].Rect, 0)
	cells[1*w+2].Pic = 0
	_, _, err := encodeSiteScreen(cells, pics, w, h, theme.Of(theme.RoleText), theme.Of(theme.RoleBackground), false)
	if err == nil {
		t.Fatal("a picture with a cell taken back was recorded")
	}
	if !strings.Contains(err.Error(), "drawn over") || !strings.Contains(err.Error(), "(0,0)") {
		t.Fatalf("the error does not name the picture: %v", err)
	}
}
