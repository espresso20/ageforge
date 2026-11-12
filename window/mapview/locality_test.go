package mapview

import (
	"image"
	"testing"
)

// A Region object near the view centre stays drawn while zooming toward it,
// through the stretch where the frames round Sol used to take the view.
func TestRegionObjectsSurviveZoomingToward(t *testing.T) {
	for _, key := range []string{"orion_nebula", "pleiades"} {
		for _, z := range []float64{20.2, 19.6, 18.6, 17.6} {
			v := nearView(t)
			at(v, z, 2, v.objAt(2, key).Pos)
			img := v.Render(tw, th)
			lit := 0
			for y := th/2 - 20; y < th/2+20; y++ {
				for x := tw/2 - 20; x < tw/2+20; x++ {
					if img.RGBAAt(x, y) != v.pal.Bg {
						lit++
					}
				}
			}
			if lit < 20 {
				t.Errorf("%s at z=%.1f: only %d lit pixels at the centre (master %d)", key, z, lit, v.m)
			}
		}
	}
}

// Zooming toward Sol still hands over to the Neighbourhood.
func TestZoomingTowardSolStillHandsOver(t *testing.T) {
	v := nearView(t)
	p, _ := v.ch.objPos(2, "sol")
	at(v, 18.4, 2, p)
	if u, o, _ := v.ch.layersAt(18.4, v.look()); u != 3 || o != 3 {
		t.Errorf("at Sol, z=18.4 draws frames %d/%d, want the Neighbourhood", u, o)
	}
}

// The Horsehead is a silhouette taller than it is wide, with a head bump
// (the top rows are narrower than the middle), a quarter of the nebula's
// radius tall, with the bright cloud behind it.
func TestHorseheadIsAHeadShapedSilhouette(t *testing.T) {
	const n = 120
	minX, maxX, minY, maxY := n, 0, n, 0
	var rows [n]int
	for j := 0; j < n; j++ {
		for i := 0; i < n; i++ {
			if horse(float64(i)/n*2-1, float64(j)/n*2-1, 7) > .5 {
				rows[j]++
				minX, maxX, minY, maxY = min(minX, i), max(maxX, i), min(minY, j), max(maxY, j)
			}
		}
	}
	w, h := maxX-minX+1, maxY-minY+1
	if h <= w {
		t.Errorf("silhouette %dx%d is not taller than wide", w, h)
	}
	if top, mid := rows[minY+h/10], rows[minY+h/2]; top >= mid {
		t.Errorf("no head bump: top row %d px, middle row %d px", top, mid)
	}
	if a, b := rows[minY+h/2], rows[minY+h-1]; a == 0 || b == 0 {
		t.Errorf("silhouette has gaps (%d, %d)", a, b)
	}
	if tall := float64(h) / n * 2 * hwin; tall < .2 || tall > .25 {
		t.Errorf("silhouette is %.2f of the nebula's radius tall, want 0.20-0.25", tall)
	}
	// the cloud behind it is brightest at the rim, not the middle
	if rim, mid := glow(hx, .85, 7), glow(0, 0, 7); rim <= mid*2 {
		t.Errorf("glow at the rim %.2f is not well above the middle %.2f", rim, mid)
	}
	if bottom := hy + hwin*(float64(maxY)/n*2-1); bottom < .6 || bottom > 1 {
		t.Errorf("the neck ends at %.2f radii down, not at the rim", bottom)
	}
}

// The Horsehead's name sits beside the silhouette, never out at the cloud's
// edge, whether or not it is selected.
func TestHorseheadLabelHugsTheSilhouette(t *testing.T) {
	v := nearView(t)
	o := objOf(v, "horsehead_nebula")
	p, _, _, _ := blankPen(v, 2, o, 70)
	box, ok := silhouetteBox(p, o)
	if !ok || box.Empty() {
		t.Fatal("no silhouette box")
	}
	for _, ringed := range []bool{false, true} {
		l := label{text: "x", mask: image.NewAlpha(image.Rect(0, 0, 20, 5)), box: extent(p, o, ringed)}
		shown, at := place([]label{l}, 400, 400)
		if len(shown) != 1 {
			t.Fatalf("ringed=%v: the label was dropped", ringed)
		}
		r := image.Rectangle{Min: at[0], Max: at[0].Add(image.Pt(20, 5))}
		near := box // what the name must hug: the silhouette, or its selection ring
		if ringed {
			cx, cy, rr := markOf(p, o)
			near = boxOf(cx, cy, rr)
		}
		gap := max(near.Min.X-r.Max.X, r.Min.X-near.Max.X, near.Min.Y-r.Max.Y, r.Min.Y-near.Max.Y)
		if gap > labelGap {
			t.Errorf("ringed=%v: label is %d art px from it, want 2", ringed, gap)
		}
	}
}
