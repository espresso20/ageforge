package ui

import "testing"

func testTerrain(w, h int) *expTerrain {
	return &expTerrain{w: w, h: h, r: make([]rune, w*h), k: make([]xk, w*h), water: make([]bool, w*h)}
}
func TestExpeditionRouteAvoidsWaterAndIsAdjacent(t *testing.T) {
	tr := testTerrain(9, 7)
	for y := 0; y < 7; y++ {
		tr.water[y*9+4] = true
	}
	tr.water[3*9+4] = false
	a, b := pt{1, 3}, pt{7, 3}
	route := routeCells(tr, a, b, 42, 99)
	if route[0] != a || route[len(route)-1] != b {
		t.Fatal("route endpoints changed")
	}
	for i, p := range route {
		if tr.isWater(p.x, p.y) {
			t.Fatalf("route crosses water at %+v", p)
		}
		if i > 0 && (abs(p.x-route[i-1].x) > 1 || abs(p.y-route[i-1].y) > 1) {
			t.Fatalf("nonadjacent cells %+v %+v", route[i-1], p)
		}
	}
}
func TestExpeditionRoutePrefersOpenGround(t *testing.T) {
	tr := testTerrain(12, 7)
	for y := 2; y <= 4; y++ {
		tr.k[y*12+5] = xHill
		tr.k[y*12+6] = xHill
		tr.k[y*12+7] = xHill
	}
	a, b := pt{1, 3}, pt{10, 3}
	route := routeCells(tr, a, b, 20, 101)
	crossed := false
	for _, p := range route {
		if tr.k[p.y*tr.w+p.x] == xHill {
			crossed = true
		}
	}
	if crossed {
		t.Fatalf("route crossed hills band: %v", route)
	}
}
func TestExpeditionRouteDeterministicAndMonotonicProgress(t *testing.T) {
	tr := testTerrain(20, 12)
	a, b := pt{2, 2}, pt{17, 9}
	first := routeCells(tr, a, b, 777, 500)
	second := routeCells(tr, a, b, 777, 500)
	if len(first) != len(second) {
		t.Fatal("route lengths differ")
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatal("route changed")
		}
	}
	last := -1
	for _, progress := range []float64{0, .2, .4, .7, 1} {
		k := clamp(int(float64(len(first)-1)*progress), 0, len(first)-1)
		if k < last {
			t.Fatal("caravan regressed")
		}
		last = k
	}
}
