package termimg

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"image"
	"image/color"
	"io"
	"strconv"
	"strings"
	"testing"
)

func tiny() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 3, 7))
	for y := 0; y < 7; y++ {
		for x := 0; x < 3; x++ {
			c := color.RGBA{255, 0, 0, 255}
			if x == 1 || y == 6 {
				c = color.RGBA{0, 0, 255, 255}
			}
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

func TestSixelGolden(t *testing.T) {
	got := string(EncodeSixel(tiny()))
	want := "\x1bP0;1;0q\"1;1;3;7#0;2;0;0;100#1;2;100;0;0#0?~$#1~?~-#0@@@-\x1b\\"
	if got != want {
		t.Errorf("sixel =\n%q\nwant\n%q", got, want)
	}
}

// decodeSixel reads back what EncodeSixel writes: w, h and the colour of every pixel.
func decodeSixel(t *testing.T, s string) (int, int, [][3]int, []int) {
	t.Helper()
	s = strings.TrimSuffix(strings.TrimPrefix(s, "\x1bP0;1;0q"), "\x1b\\")
	var w, h int
	if _, err := sscanRaster(s, &w, &h); err != nil {
		t.Fatal(err)
	}
	s = s[strings.IndexAny(s, "#"):]
	pal := map[int][3]int{}
	px := make([]int, w*h)
	for i := range px {
		px[i] = -1
	}
	cur, x, band := 0, 0, 0
	for i := 0; i < len(s); {
		switch c := s[i]; {
		case c == '#':
			j := i + 1
			for j < len(s) && s[j] >= '0' && s[j] <= '9' {
				j++
			}
			n, _ := strconv.Atoi(s[i+1 : j])
			if j < len(s) && s[j] == ';' { // definition
				k := j
				for k < len(s) && (s[k] == ';' || s[k] >= '0' && s[k] <= '9') {
					k++
				}
				f := strings.Split(s[j+1:k], ";")
				var rgb [3]int
				for q := 0; q < 3; q++ {
					rgb[q], _ = strconv.Atoi(f[q+1])
				}
				pal[n] = rgb
				j = k
			}
			cur = n
			i = j
		case c == '$':
			x = 0
			i++
		case c == '-':
			x, band = 0, band+1
			i++
		case c == '!':
			j := i + 1
			for s[j] >= '0' && s[j] <= '9' {
				j++
			}
			n, _ := strconv.Atoi(s[i+1 : j])
			put(px, w, h, band, x, n, s[j], cur)
			x += n
			i = j + 1
		default:
			put(px, w, h, band, x, 1, c, cur)
			x++
			i++
		}
	}
	cols := make([][3]int, len(px))
	for i, p := range px {
		if p >= 0 {
			cols[i] = pal[p]
		}
	}
	return w, h, cols, px
}

func sscanRaster(s string, w, h *int) (int, error) {
	i := strings.Index(s, "\"1;1;")
	f := strings.SplitN(s[i+5:], "#", 2)[0]
	p := strings.Split(f, ";")
	var err error
	if *w, err = strconv.Atoi(p[0]); err != nil {
		return 0, err
	}
	*h, err = strconv.Atoi(p[1])
	return 0, err
}

func put(px []int, w, h, band, x, n int, ch byte, col int) {
	for k := 0; k < n; k++ {
		for r := 0; r < 6; r++ {
			if (ch-63)>>r&1 == 1 && band*6+r < h && x+k < w {
				px[(band*6+r)*w+x+k] = col
			}
		}
	}
}

func TestSixelRoundTrip(t *testing.T) {
	img := tiny()
	w, h, cols, _ := decodeSixel(t, string(EncodeSixel(img)))
	if w != 3 || h != 7 {
		t.Fatalf("size %dx%d", w, h)
	}
	for y := 0; y < 7; y++ {
		for x := 0; x < 3; x++ {
			want := [3]int{100, 0, 0}
			if x == 1 || y == 6 {
				want = [3]int{0, 0, 100}
			}
			if cols[y*3+x] != want {
				t.Fatalf("pixel %d,%d = %v, want %v", x, y, cols[y*3+x], want)
			}
		}
	}
}

// A smooth ramp must not band into a few flat steps: the dithered picture
// uses many levels and stays close to the original on average.
func TestSixelRampDithers(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 256, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 256; x++ {
			v := uint8(4 + x/16) // 16 levels of dark blue-grey, 16 px wide each
			img.SetRGBA(x, y, color.RGBA{v, v, v + 6, 255})
		}
	}
	pal, idx := Quantize(img)
	var sum, want float64
	for y := 0; y < 24; y++ {
		for x := 0; x < 256; x++ {
			p := pal[idx[y*256+x]]
			sum += float64(p[2])
			want += float64(img.RGBAAt(x, y).B)
		}
	}
	if d := (sum - want) / (256 * 24); d > 1 || d < -1 {
		t.Errorf("mean blue drifts by %.2f", d)
	}
}

func TestKittyGolden(t *testing.T) {
	img := tiny()
	got := string(EncodeKitty(img, 2, 1))
	pre := "\x1b_Ga=T,f=24,o=z,q=2,C=1,i=7,p=1,s=3,v=7,c=2,r=1,m=0;"
	if !strings.HasPrefix(got, pre) || !strings.HasSuffix(got, "\x1b\\") {
		t.Fatalf("kitty = %q", got)
	}
	data := strings.TrimSuffix(strings.TrimPrefix(got, pre), "\x1b\\")
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	px, _ := io.ReadAll(zr)
	if len(px) != 3*7*3 || !bytes.Equal(px[:6], []byte{255, 0, 0, 0, 0, 255}) {
		t.Errorf("pixels = %v", px)
	}
	if got, want := string(KittyDelete()), "\x1b_Ga=d,d=I,q=2,i=7\x1b\\"; got != want {
		t.Errorf("delete = %q", got)
	}
}

func TestKittyChunks(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	var r uint32 = 1
	for i := range img.Pix {
		r = r*1664525 + 1013904223
		img.Pix[i] = uint8(r >> 24)
	}
	got := string(EncodeKitty(img, 10, 5))
	parts := strings.Split(strings.TrimSuffix(got, "\x1b\\"), "\x1b\\")
	if len(parts) < 3 {
		t.Fatalf("%d chunks", len(parts))
	}
	for i, p := range parts {
		body := p[strings.IndexByte(p, ';')+1:]
		if len(body) > kittyChunk {
			t.Errorf("chunk %d is %d bytes", i, len(body))
		}
		more := strings.Contains(p[:strings.IndexByte(p, ';')], "m=1")
		if more != (i < len(parts)-1) {
			t.Errorf("chunk %d more=%v", i, more)
		}
		if i > 0 && !strings.HasPrefix(p, "\x1b_Gm=") {
			t.Errorf("chunk %d header %q", i, p[:12])
		}
	}
}
