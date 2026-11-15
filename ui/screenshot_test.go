package ui

import (
	"image"
	"image/color"
	"testing"

	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// shotCells is cols by rows blank cells, white on black, with set drawing
// into them.
func shotCells(cols, rows int, set func(x, y int) (rune, kit.Style, bool)) *kit.Buffer {
	base := kit.Style{}.Foreground(kit.RGB(255, 255, 255)).Background(kit.RGB(0, 0, 0))
	b := kit.NewBuffer(cols, rows)
	p := kit.NewPainter(b)
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			p.Put(x, y, ' ', base)
			if r, st, ok := set(x, y); ok {
				p.Put(x, y, r, st)
			}
		}
	}
	return b
}

// cellPixels counts the pixels of one cell that are white and that are black.
func cellPixels(img *image.RGBA, x, y int) (white, black, other int) {
	ox, oy := shotPad+x*shotCellW, shotPad+y*shotCellH
	for py := oy; py < oy+shotCellH; py++ {
		for px := ox; px < ox+shotCellW; px++ {
			switch img.RGBAAt(px, py) {
			case color.RGBA{255, 255, 255, 255}:
				white++
			case color.RGBA{0, 0, 0, 255}:
				black++
			default:
				other++
			}
		}
	}
	return
}

func TestScreenImageSizeAndLetters(t *testing.T) {
	cells := shotCells(4, 2, func(x, y int) (rune, kit.Style, bool) {
		if x == 1 && y == 0 {
			return 'W', kit.Style{}.Foreground(kit.RGB(255, 255, 255)).Background(kit.RGB(0, 0, 0)), true
		}
		return 0, kit.Style{}, false
	})
	img := ScreenImage(cells, shotCellW, shotCellH)
	if b := img.Bounds(); b.Dx() != 4*shotCellW+2*shotPad || b.Dy() != 2*shotCellH+2*shotPad {
		t.Fatalf("size = %v", b)
	}
	if w, _, _ := cellPixels(img, 1, 0); w == 0 {
		t.Error("a cell holding W has no foreground pixels")
	}
	if w, b, o := cellPixels(img, 0, 0); w != 0 || o != 0 || b != shotCellW*shotCellH {
		t.Errorf("a blank cell is not one colour: %d %d %d", w, b, o)
	}
	again := ScreenImage(cells, shotCellW, shotCellH)
	if string(img.Pix) != string(again.Pix) {
		t.Error("the same cells gave two pictures")
	}
}

func TestScreenImageBlocksFillTheirCells(t *testing.T) {
	st := kit.Style{}.Foreground(kit.RGB(255, 255, 255)).Background(kit.RGB(0, 0, 0))
	area := shotCellW * shotCellH
	for r, want := range map[rune]int{
		'█': area, '▀': area / 2, '▄': area / 2, '▌': area / 2, '▐': area / 2,
		'▘': area / 4, '▗': area / 4, '▚': area / 2, '▙': area * 3 / 4,
	} {
		img := ScreenImage(shotCells(1, 1, func(int, int) (rune, kit.Style, bool) { return r, st, true }), shotCellW, shotCellH)
		if w, _, o := cellPixels(img, 0, 0); w != want || o != 0 {
			t.Errorf("%c: %d white pixels, %d other, want %d white", r, w, o, want)
		}
	}
	// A shade is a mix: no pixel is pure foreground or pure background.
	img := ScreenImage(shotCells(1, 1, func(int, int) (rune, kit.Style, bool) { return '▒', st, true }), shotCellW, shotCellH)
	if w, b, o := cellPixels(img, 0, 0); w != 0 || b != 0 || o != area {
		t.Errorf("shade: %d white, %d black, %d mixed", w, b, o)
	}
}

func TestScreenImageLinesMeetAcrossCells(t *testing.T) {
	st := kit.Style{}.Foreground(kit.RGB(255, 255, 255)).Background(kit.RGB(0, 0, 0))
	img := ScreenImage(shotCells(3, 1, func(int, int) (rune, kit.Style, bool) { return '─', st, true }), shotCellW, shotCellH)
	y := shotPad + shotCellH/2
	for x := shotPad; x < shotPad+3*shotCellW; x++ {
		if img.RGBAAt(x, y) != (color.RGBA{255, 255, 255, 255}) {
			t.Fatalf("gap in the line at x=%d", x)
		}
	}
}

func TestScreenImageDrawsNoTofu(t *testing.T) {
	st := kit.Style{}.Foreground(kit.RGB(255, 255, 255)).Background(kit.RGB(0, 0, 0))
	// Marks Go Mono has no shape for stay blank, and so does a diagonal.
	for _, r := range []rune{'🙂', '╱'} {
		img := ScreenImage(shotCells(1, 1, func(int, int) (rune, kit.Style, bool) { return r, st, true }), shotCellW, shotCellH)
		if _, b, _ := cellPixels(img, 0, 0); b != shotCellW*shotCellH {
			t.Errorf("%c is not blank", r)
		}
	}
}

func TestScreenImageDrawsTheEpochIcons(t *testing.T) {
	st := kit.Style{}.Foreground(kit.RGB(255, 255, 255)).Background(kit.RGB(0, 0, 0))
	for r := range symbolShapes {
		img := ScreenImage(shotCells(1, 1, func(int, int) (rune, kit.Style, bool) { return r, st, true }), shotCellW, shotCellH)
		if _, b, _ := cellPixels(img, 0, 0); b == shotCellW*shotCellH {
			t.Errorf("%c is blank", r)
		}
	}
	for _, e := range func() []config.EpochDef { st := game.NewGameEngine().GetState(); return st.Ruleset().Eras() }() {
		for _, r := range e.Icon {
			if symbolShapes[r] == nil && !fontHas(r) {
				t.Errorf("the icon %c of %s has no shape and is not in the font", r, e.Name)
			}
		}
	}
}

func fontHas(r rune) bool {
	f, err := opentype.Parse(gomono.TTF)
	if err != nil {
		return false
	}
	var b sfnt.Buffer
	i, _ := f.GlyphIndex(&b, r)
	return i != 0
}

func TestEpochPictureIsStillAndNotFlat(t *testing.T) {
	st := lateHallState(t, "galactic_age")
	a, b := epochPicture(st), epochPicture(st)
	if a == nil || b == nil {
		t.Fatal("no picture")
	}
	if string(a.Pix) != string(b.Pix) {
		t.Error("two pictures of one state differ")
	}
	if a.Bounds().Dx() != epochPicCols*shotCellW+2*shotPad {
		t.Errorf("width = %d", a.Bounds().Dx())
	}
	first := a.RGBAAt(shotPad+5*shotCellW, shotPad+5*shotCellH)
	flat := true
	for i := 0; i+4 <= len(a.Pix) && flat; i += 4 {
		flat = a.Pix[i] == first.R && a.Pix[i+1] == first.G && a.Pix[i+2] == first.B
	}
	if flat {
		t.Error("the picture is all one colour")
	}
}

func TestScreenImageDrawsPictures(t *testing.T) {
	b := shotCells(10, 6, func(x, y int) (rune, kit.Style, bool) { return 0, kit.Style{}, false })
	red := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for i := 0; i < len(red.Pix); i += 4 {
		red.Pix[i], red.Pix[i+3] = 255, 255
	}
	p := kit.NewPainter(b)
	p.Picture(image.Rect(2, 1, 6, 3), red)
	isRed := func(img *image.RGBA, cx, cy int) bool {
		c := img.RGBAAt(shotPad+cx*shotCellW+shotCellW/2, shotPad+cy*shotCellH+shotCellH/2)
		return c.R == 255 && c.G == 0 && c.B == 0
	}
	img := ScreenImage(b, shotCellW, shotCellH)
	if !isRed(img, 3, 1) || !isRed(img, 5, 2) {
		t.Fatal("the picture's cells are not red")
	}
	if isRed(img, 1, 1) || isRed(img, 6, 1) || isRed(img, 3, 3) {
		t.Fatal("a cell outside the picture is red")
	}
	p.Put(3, 1, 'x', kit.Style{}.Foreground(kit.RGB(255, 255, 255)).Background(kit.RGB(0, 0, 0)))
	img = ScreenImage(b, shotCellW, shotCellH)
	if isRed(img, 3, 1) {
		t.Fatal("an inked cell still shows the picture")
	}
	if !isRed(img, 4, 1) {
		t.Fatal("the picture's other cells lost it")
	}
}
