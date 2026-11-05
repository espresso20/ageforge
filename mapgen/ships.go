package mapgen

// Ships in the space scenes: a solid hull of one or two pixels with a short
// straight glow behind it, flying back and forth between two bodies on a
// straight line, never more than a pixel a frame. Everything about a ship
// (route, kind, timing) comes from the seed.

// shipKind is one sort of ship. The key is the traffic the cursor names; speed
// is in 16ths of a pixel a frame along the longer axis of the route.
type shipKind struct {
	key   string
	hull  Paint
	glow  Hue
	size  int // hull pixels
	tail  int // glow pixels
	speed int
}

var shipKinds = [...]shipKind{
	{"shuttle", P(HWhite, 255), HEngine, 1, 1, 9},
	{"cargo_hauler", P(HSteel, 255), HPlasma, 2, 2, 6},
	{"generation_ship", P(HCyan, 255), HEngine, 2, 1, 4},
}

// shipBody is a body the ships fly between or keep out of: its centre and
// radius in 16ths of a pixel.
type shipBody struct{ x, y, r int }

func (b shipBody) covers(px, py int) bool {
	dx, dy := px*pxUnit+pxUnit/2-b.x, py*pxUnit+pxUnit/2-b.y
	return dx*dx+4*dy*dy <= b.r*b.r
}

// shipCount is how many ships a seed has.
func shipCount(seed int64) int { return 2 + roll(3, seed, 5100) }

// shipPose is where ship i is at frame anim: its pixel position and heading in
// 16ths of a pixel (16 along the longer axis). ok is false for a route too
// short to fly.
func shipPose(seed int64, anim, i int, bodies []shipBody) (k shipKind, x, y, hx, hy int, ok bool) {
	a := roll(len(bodies), seed, 5101, int64(i))
	b := (a + 1 + roll(len(bodies)-1, seed, 5102, int64(i))) % len(bodies)
	k = shipKinds[roll(len(shipKinds), seed, 5103, int64(i))]
	dx, dy := bodies[b].x-bodies[a].x, bodies[b].y-bodies[a].y
	L := max(abs(dx), abs(dy))
	if L < 6*pxUnit {
		return k, 0, 0, 0, 0, false
	}
	s := (anim*k.speed + roll(1<<20, seed, 5104, int64(i))) % (2 * L)
	hx, hy = dx*pxUnit/L, dy*pxUnit/L
	if s > L {
		s, hx, hy = 2*L-s, -hx, -hy
	}
	return k, bodies[a].x + dx*s/L, bodies[a].y + dy*s/L, hx, hy, true
}

// Ships draws the seed's ships between the bodies. They are hidden inside any
// body or keep-out (so they pass behind the planets and the sun) and where
// skip says so.
func (p *Pixels) Ships(seed int64, anim int, bodies, keep []shipBody, skip func(x, y int) bool) {
	if len(bodies) < 2 {
		return
	}
	for i := 0; i < shipCount(seed); i++ {
		k, x, y, hx, hy, ok := shipPose(seed, anim, i, bodies)
		if !ok {
			continue
		}
		own := p.Owner(Thing{Kind: KindAgent, Key: k.key}, 6)
		for j := 0; j < k.size+k.tail; j++ {
			c := k.hull
			if j >= k.size {
				c = P(k.glow, 200-70*(j-k.size))
			}
			px, py := floorDiv(x-hx*j, pxUnit), floorDiv(y-hy*j, pxUnit)
			if skip != nil && skip(px, py) || coveredBy(px, py, bodies) || coveredBy(px, py, keep) {
				continue
			}
			p.Set(px, py, c, own)
		}
	}
}

func coveredBy(px, py int, bs []shipBody) bool {
	for _, b := range bs {
		if b.covers(px, py) {
			return true
		}
	}
	return false
}
