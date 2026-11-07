package termimg

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"image"
	"strconv"
)

// kittyImageID is the one image the game keeps on screen.
const kittyImageID = 7

const kittyChunk = 4096 // base64 bytes per escape

// EncodeKitty is img as kitty graphics sequences that transmit and show it
// (id kittyImageID, placement 1) at the cursor, over cols by rows cells. The
// pixels go as zlib-compressed RGB.
func EncodeKitty(img *image.RGBA, cols, rows int) []byte {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	var raw bytes.Buffer
	zw, _ := zlib.NewWriterLevel(&raw, zlib.BestSpeed)
	row := make([]byte, 3*w)
	for y := 0; y < h; y++ {
		src := img.Pix[y*img.Stride : y*img.Stride+4*w]
		for x := 0; x < w; x++ {
			row[3*x], row[3*x+1], row[3*x+2] = src[4*x], src[4*x+1], src[4*x+2]
		}
		zw.Write(row)
	}
	zw.Close()
	enc := make([]byte, base64.StdEncoding.EncodedLen(raw.Len()))
	base64.StdEncoding.Encode(enc, raw.Bytes())
	out := make([]byte, 0, len(enc)+len(enc)/kittyChunk*8+128)
	for first := true; len(enc) > 0; first = false {
		n := min(len(enc), kittyChunk)
		out = append(out, "\x1b_G"...)
		if first {
			out = append(out, "a=T,f=24,o=z,q=2,C=1,i="...)
			out = strconv.AppendInt(out, kittyImageID, 10)
			out = append(out, ",p=1,s="...)
			out = strconv.AppendInt(out, int64(w), 10)
			out = append(out, ",v="...)
			out = strconv.AppendInt(out, int64(h), 10)
			out = append(out, ",c="...)
			out = strconv.AppendInt(out, int64(cols), 10)
			out = append(out, ",r="...)
			out = strconv.AppendInt(out, int64(rows), 10)
			out = append(out, ',')
		}
		if n < len(enc) {
			out = append(out, "m=1;"...)
		} else {
			out = append(out, "m=0;"...)
		}
		out = append(out, enc[:n]...)
		out = append(out, "\x1b\\"...)
		enc = enc[n:]
	}
	return out
}

// KittyDelete is the sequence that removes the image and frees its data.
func KittyDelete() []byte {
	return []byte("\x1b_Ga=d,d=I,q=2,i=" + strconv.Itoa(kittyImageID) + "\x1b\\")
}
