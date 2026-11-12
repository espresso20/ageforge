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
