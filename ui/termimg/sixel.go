package termimg

import (
	"image"
	"sort"
	"strconv"
)

// Sixel encoding: a 256-colour palette cut from the picture's own colours
// (median cut), the picture dithered onto it (Floyd-Steinberg, so the soft
// gradients of a galaxy do not band), then the sixel bands.

const (
	qbits = 6 // bits per channel the palette search works in
	qn    = 1 << (3 * qbits)
)

type palBox struct {
	bins   []int32 // histogram bins in the box
	pop    int
	lo, hi [3]int // channel ranges, in qbits
}

func qkey(r, g, b uint8) int {
	return int(r>>(8-qbits))<<(2*qbits) | int(g>>(8-qbits))<<qbits | int(b>>(8-qbits))
}

// paletteOf cuts at most n colours from img.
func paletteOf(img *image.RGBA, n int) [][3]uint8 {
	var cnt [qn]uint32
	var sum [qn][3]uint32
	pix := img.Pix
	for i := 0; i+3 < len(pix); i += 4 {
		r, g, b := pix[i], pix[i+1], pix[i+2]
		k := qkey(r, g, b)
		cnt[k]++
		s := &sum[k]
		s[0] += uint32(r)
		s[1] += uint32(g)
		s[2] += uint32(b)
	}
	chan3 := func(k int32, c int) int { return int(k>>(uint(2-c)*qbits)) & (1<<qbits - 1) }
	var all []int32
	pop := 0
	for k := range cnt {
		if cnt[k] > 0 {
			all = append(all, int32(k))
			pop += int(cnt[k])
		}
	}
	measure := func(bins []int32) palBox {
		b := palBox{bins: bins, lo: [3]int{1 << qbits, 1 << qbits, 1 << qbits}}
		for _, k := range bins {
			b.pop += int(cnt[k])
			for c := 0; c < 3; c++ {
				v := chan3(k, c)
				b.lo[c], b.hi[c] = min(b.lo[c], v), max(b.hi[c], v)
			}
		}
		return b
	}
	boxes := []palBox{measure(all)}
	for len(boxes) < n {
		best, score := -1, 0
		for i, b := range boxes {
			if len(b.bins) < 2 {
				continue
			}
			rng := max(b.hi[0]-b.lo[0], b.hi[1]-b.lo[1], b.hi[2]-b.lo[2]) + 1
			if s := b.pop * rng; s > score {
				best, score = i, s
			}
		}
		if best < 0 {
			break
		}
		b := boxes[best]
		axis := 0
		for c := 1; c < 3; c++ {
			if b.hi[c]-b.lo[c] > b.hi[axis]-b.lo[axis] {
				axis = c
			}
		}
		sort.Slice(b.bins, func(i, j int) bool { return chan3(b.bins[i], axis) < chan3(b.bins[j], axis) })
		half, acc := b.pop/2, 0
		cut := 1
		for i, k := range b.bins {
			acc += int(cnt[k])
			if acc >= half {
				cut = min(max(i+1, 1), len(b.bins)-1)
				break
			}
		}
		boxes[best] = measure(b.bins[:cut])
		boxes = append(boxes, measure(b.bins[cut:]))
	}
	pal := make([][3]uint8, len(boxes))
	for i, b := range boxes {
		var s [3]uint64
		var c uint64
		for _, k := range b.bins {
			c += uint64(cnt[k])
			for ch := 0; ch < 3; ch++ {
				s[ch] += uint64(sum[k][ch])
			}
		}
		if c == 0 {
			continue
		}
		pal[i] = [3]uint8{uint8(s[0] / c), uint8(s[1] / c), uint8(s[2] / c)}
	}
	return pal
}

// Quantize maps img onto at most 256 colours, dithered, and returns the palette
// and the palette index of every pixel.
func Quantize(img *image.RGBA) ([][3]uint8, []uint8) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	pal := paletteOf(img, 256)
	lut := make([]int16, qn) // 0 unknown, else index+1
	nearest := func(r, g, bl int) int {
		k := qkey(uint8(r), uint8(g), uint8(bl))
		if v := lut[k]; v != 0 {
			return int(v) - 1
		}
		// search from the bin's middle
		cr, cg, cb := r>>(8-qbits)<<(8-qbits)+1<<(7-qbits), g>>(8-qbits)<<(8-qbits)+1<<(7-qbits), bl>>(8-qbits)<<(8-qbits)+1<<(7-qbits)
		best, bd := 0, 1<<30
		for i, p := range pal {
			dr, dg, db := cr-int(p[0]), cg-int(p[1]), cb-int(p[2])
			if d := dr*dr*3 + dg*dg*4 + db*db*2; d < bd {
				best, bd = i, d
			}
		}
		lut[k] = int16(best + 1)
		return best
	}
	idx := make([]uint8, w*h)
	cur := make([]int32, 3*(w+2))
	nxt := make([]int32, 3*(w+2))
	for y := 0; y < h; y++ {
		row := img.Pix[y*img.Stride : y*img.Stride+4*w]
		for i := range nxt {
			nxt[i] = 0
		}
		for x := 0; x < w; x++ {
			var v [3]int
			for c := 0; c < 3; c++ {
				v[c] = min(255, max(0, int(row[4*x+c])+int(cur[3*(x+1)+c])>>4))
			}
			k := nearest(v[0], v[1], v[2])
			idx[y*w+x] = uint8(k)
			p := pal[k]
			for c := 0; c < 3; c++ {
				e := int32(v[c] - int(p[c]))
				cur[3*(x+2)+c] += e * 7
				nxt[3*x+c] += e * 3
				nxt[3*(x+1)+c] += e * 5
				nxt[3*(x+2)+c] += e
			}
		}
		cur, nxt = nxt, cur
	}
	return pal, idx
}

// EncodeSixel is img as a sixel sequence (DCS ... ST), 1:1 pixels.
func EncodeSixel(img *image.RGBA) []byte {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	pal, idx := Quantize(img)
	out := make([]byte, 0, w*h/8)
	out = append(out, "\x1bP0;1;0q\"1;1;"...)
	out = strconv.AppendInt(out, int64(w), 10)
	out = append(out, ';')
	out = strconv.AppendInt(out, int64(h), 10)
	for i, p := range pal {
		out = append(out, '#')
		out = strconv.AppendInt(out, int64(i), 10)
		out = append(out, ";2;"...)
		for c := 0; c < 3; c++ {
			out = strconv.AppendInt(out, (int64(p[c])*100+127)/255, 10)
			if c < 2 {
				out = append(out, ';')
			}
		}
	}
	bits := make([]uint8, len(pal)*w)
	maxX := make([]int, len(pal))
	stamp := make([]int, len(pal))
	used := make([]int, 0, len(pal))
	for y0 := 0; y0 < h; y0 += 6 {
		used = used[:0]
		for r := 0; r < 6 && y0+r < h; r++ {
			rowIdx := idx[(y0+r)*w : (y0+r+1)*w]
			for x, c := range rowIdx {
				if stamp[c] != y0+1 {
					stamp[c] = y0 + 1
					used = append(used, int(c))
				}
				bits[int(c)*w+x] |= 1 << r
				maxX[c] = max(maxX[c], x)
			}
		}
		sort.Ints(used)
		for n, c := range used {
			out = append(out, '#')
			out = strconv.AppendInt(out, int64(c), 10)
			run, last := 0, byte(0)
			flush := func() {
				switch {
				case run == 0:
				case run > 3:
					out = append(out, '!')
					out = strconv.AppendInt(out, int64(run), 10)
					out = append(out, last)
				default:
					for ; run > 0; run-- {
						out = append(out, last)
					}
				}
				run = 0
			}
			row := bits[c*w : c*w+maxX[c]+1]
			for x, bv := range row {
				ch := 63 + bv
				if ch != last {
					flush()
					last = ch
				}
				run++
				row[x] = 0
			}
			flush()
			if n < len(used)-1 {
				out = append(out, '$')
			}
			maxX[c] = 0
		}
		out = append(out, '-')
	}
	return append(out, "\x1b\\"...)
}
