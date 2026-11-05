package mapgen

import "testing"

func TestReduceQuadPatterns(t *testing.T) {
	g, y := P(HEarth, 30), P(HCore, 250)
	cases := []struct {
		name string
		px   [4]Paint
		r    rune
		fg   Paint
		bg   Paint
	}{
		{"flat", [4]Paint{y, y, y, y}, '█', y, y},
		{"top left", [4]Paint{y, g, g, g}, '▘', y, g},
		{"top row", [4]Paint{y, y, g, g}, '▀', y, g},
		{"left column", [4]Paint{y, g, y, g}, '▌', y, g},
		{"diagonal", [4]Paint{y, g, g, y}, '▚', y, g},
	}
	for _, c := range cases {
		r, fg, bg := reduceQuad(c.px, g)
		if r != c.r || fg != c.fg || bg != c.bg {
			t.Errorf("%s: got %q %v %v, want %q %v %v", c.name, r, fg, bg, c.r, c.fg, c.bg)
		}
	}
}

// A third colour joins the nearer of the two that fill most of the block.
func TestReduceQuadMapsThirdColour(t *testing.T) {
	g, y, o := P(HEarth, 30), P(HCore, 250), P(HCore, 240)
	r, fg, bg := reduceQuad([4]Paint{y, o, g, g}, g)
	if r != '▀' {
		t.Errorf("got %q, want the two core pixels as one colour (▀)", r)
	}
	if bg != g || fg.Hue != HCore || fg.Weight != 245 {
		t.Errorf("colours %v %v, want core over ground", fg, bg)
	}
}

func TestPixelsResolveOwnersAndDisc(t *testing.T) {
	g := newGrid(10, 6)
	px := NewPixels(10, 6)
	sun := px.Owner(Thing{Kind: KindTerrain, Key: "sun"}, 4)
	ring := px.Owner(Thing{Kind: KindRoad, Key: "orbit"}, 2)
	px.Disc(10*pxUnit, 6*pxUnit, 3*pxUnit, sun, func(nx, ny int) Paint { return P(HCore, 250) })
	px.Set(2, 2, P(HOrbit, 175), ring)
	px.Set(10, 6, P(HOrbit, 175), ring)
	px.Resolve(g)
	if th := g.ThingAt(5, 3); th.Key != "sun" {
		t.Errorf("disc centre cell reports %q, want sun", th.Key)
	}
	if th := g.ThingAt(1, 1); th.Key != "orbit" {
		t.Errorf("ring pixel cell reports %q, want orbit", th.Key)
	}
	if c := g.At(1, 1); c.R == 0 || c.R == ' ' {
		t.Errorf("ring pixel cell is %q, want a quadrant glyph", c.R)
	}
	if th := g.ThingAt(8, 5); th.Key != "" {
		t.Errorf("empty cell reports %q", th.Key)
	}
}
