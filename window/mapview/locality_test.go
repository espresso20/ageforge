package mapview

import "testing"

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
// (the top rows are narrower than the middle), and bright cloud behind it.
func TestHorseheadIsAHeadShapedSilhouette(t *testing.T) {
	const n = 120
	minX, maxX, minY, maxY := n, 0, n, 0
	var rows [n]int
	bright := 0
	for j := 0; j < n; j++ {
		for i := 0; i < n; i++ {
			d, f := horse(float64(i)/n*2-1, float64(j)/n*2-1, 7)
			if d > .5 {
				rows[j]++
				minX, maxX, minY, maxY = min(minX, i), max(maxX, i), min(minY, j), max(maxY, j)
			} else if f > .3 {
				bright++
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
	if bright < 500 {
		t.Errorf("only %d bright cloud pixels behind the silhouette", bright)
	}
	if a, b := rows[minY+h/2], rows[minY+h-1]; a == 0 || b == 0 {
		t.Errorf("silhouette has gaps (%d, %d)", a, b)
	}
}
