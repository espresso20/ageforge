package ui

import (
	"bytes"
	"image"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// chipLook is what is drawn over r: its cells and the pictures inside it.
func chipLook(st *stage, r image.Rectangle) (cells []kit.Cell, pics [][]byte) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			cells = append(cells, st.buf.At(x, y))
		}
	}
	for _, p := range st.buf.Pics {
		if p.Rect.Overlaps(r) {
			pics = append(pics, p.Img.Pix)
		}
	}
	return
}

func sameLook(ac []kit.Cell, ap [][]byte, bc []kit.Cell, bp [][]byte) bool {
	if len(ac) != len(bc) || len(ap) != len(bp) {
		return false
	}
	for i := range ac {
		if ac[i] != bc[i] {
			return false
		}
	}
	for i := range ap {
		if !bytes.Equal(ap[i], bp[i]) {
			return false
		}
	}
	return true
}

// A one-row Build key on a card that is not picked lights under the pointer,
// with the glyph tier and with pictures, on a small screen and a wide one; one
// that is off does not.
func TestOneRowBuildKeyLightsUnderThePointer(t *testing.T) {
	for _, pix := range []bool{false, true} {
		for _, wide := range []bool{false, true} {
			was := pictureCell
			if pix {
				pictureCell = func() (int, int) { return 7, 15 }
			}
			t.Cleanup(func() { pictureCell = was })
			d, st, _ := cardTown(t)
			if wide {
				d, st, _ = medievalTown(t, 279, 74)
			}
			box := d.economyTab.buildBox
			st.paint()
			tested := 0
			for _, k := range box.cardKeys.keys {
				if k.r.Dy() != 1 || k.cmd == "" {
					continue
				}
				bx, by, _, _ := box.GetRect()
				r := k.r.Add(image.Pt(bx, by))
				// At rest: the pointer on the same card, off the key.
				st.mouse(kit.Mouse{Kind: kit.MouseMove, X: bx + 2, Y: r.Min.Y})
				st.paint()
				rc, rp := chipLook(st, r)
				p := centre(d, k.r)
				st.mouse(kit.Mouse{Kind: kit.MouseMove, X: p.X, Y: p.Y})
				st.paint()
				hc, hp := chipLook(st, r)
				tested++
				if same := sameLook(rc, rp, hc, hp); same == !k.off {
					t.Errorf("pictures %v wide %v: the %q key (off %v) looks the same %v under the pointer", pix, wide, k.cmd, k.off, same)
				}
			}
			if tested == 0 {
				t.Errorf("pictures %v wide %v: no one-row Build key", pix, wide)
			}
		}
	}
}
