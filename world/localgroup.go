package world

// The Local Group: the void between galaxies, the Milky Way, its neighbours
// and a scatter of far ones. Moved from mapgen's c.void. The original lays
// everything out in fractions of the window; here the same fractions fill a
// nominal canvas (lgW x lgH cells, isotropic) and the canvas is scaled to
// kly, so the picture keeps its arrangement and the frame is marked
// schematic (Andromeda is 2500 kly away, but sits 1590 kly from the Milky
// Way on the page).

const (
	lgW   = 200 // nominal canvas, in cells
	lgH   = 100 // and rows, isotropic
	lgKly = 15  // kly per nominal cell
)

func lgVec(x, y int) Vec { return Vec{float64(x-lgW/2) * lgKly, float64(y-lgH/2) * lgKly} }

func localGroup(seed int64) *Frame {
	f := &Frame{
		Kind: FrameLocalGroup, Key: "local_group", Unit: UnitKly,
		Radius: lgW / 2 * lgKly, Band: Band{Min: 400, Max: 6000}, Schem: true,
	}
	W, H := lgW, lgH
	mwX, mwY, mwR := W*26/100, H*46/100, min(W*23/100, H*36/100)
	m31X, m31Y, m31R := W*79/100, H*30/100, W*14/100
	m33X, m33Y, m33R := W*88/100, H*72/100, W*6/100
	lmcX, lmcY := W*46/100, H*80/100
	smcX, smcY := W*56/100, H*88/100
	type major struct{ x, y, r int }
	majors := []major{{mwX, mwY, mwR}, {m31X, m31Y, m31R}, {m33X, m33Y, m33R}, {lmcX, lmcY, 10}, {smcX, smcY, 8}}

	gal := func(key string, x, y, r int, look GalaxyLook, tone Tone, pick bool) Object {
		look.R = r * lgKly
		return Object{Kind: KindGalaxy, Key: key, Label: key, Inspect: key, Pos: lgVec(x, y), Size: float64(r * lgKly),
			Tone: tone, Weight: 255, Galaxy: &look, Pick: pick}
	}
	f.Objects = append(f.Objects,
		gal("milky_way", mwX, mwY, mwR, LookMilkyWay, ToneStar, true),
		gal("andromeda", m31X, m31Y, m31R, LookAndromeda, ToneStar, true),
		gal("triangulum", m33X, m33Y, m33R, LookTriangulum, ToneStar, true),
	)
	for i, p := range [2]struct {
		key          string
		x, y, rx, ry int
	}{{"large_magellanic_cloud", lmcX, lmcY, 9, 6}, {"small_magellanic_cloud", smcX, smcY, 6, 4}} {
		o := gal(p.key, p.x, p.y, p.rx, GalaxyLook{Q: 256 * p.ry / p.rx, Stars: 14}, ToneCool, true) // irregular: no arms
		_ = i
		f.Objects = append(f.Objects, o)
	}
	// Sol, in the Orion spur of the Milky Way
	mw := LookMilkyWay
	mw.R = mwR
	sx, sy, _ := ArmXY(mw, mwX, mwY, 0, 170)
	f.Objects = append(f.Objects, Object{Kind: KindStar, Key: "sol", Label: "sol", Pos: lgVec(sx, sy), Tone: ToneSun, Weight: 255,
		Star: &Look{Tone: ToneSun, Size: 1, Weight: 255, Extra: ToneWhite}})
	// the lane of stars from the Milky Way toward Andromeda, ending short of it
	x0, y0, x1, y1 := mwX+mwR+4, mwY-mwR/4, m31X-m31R-8, m31Y+m31R/8
	f.Objects = append(f.Objects, Object{Kind: KindFeature, Key: "lane_to_andromeda", Pos: lgVec(x0, y0), To: lgVec(x1, y1), Tone: ToneStar, Weight: 110})
	// far galaxies, faint
	for i := 0; i < 24 && len(majors) < 16; i++ {
		x, y := Roll(W, seed, 7070, int64(i)), Roll(H, seed, 7071, int64(i))
		free := true
		for _, m := range majors {
			if Sq(x-m.x)+Sq(y-m.y) < Sq(m.r+14) {
				free = false
			}
		}
		if !free {
			continue
		}
		majors = append(majors, major{x, y, 6})
		ph, amp := Roll(Turn/2, seed, 7072, int64(i)), 45+Roll(50, seed, 7073, int64(i))
		look := GalaxyLook{Phi: ph}
		tone := ToneSun
		r := 5
		switch i % 3 {
		case 0: // edge-on
			look.Q, r = 32, 8
		case 1: // a small spiral
			look.Q, look.Arms, r, tone = 171, 2, 6, ToneCool
		default: // an elliptical
			look.Q = 205
		}
		o := gal("far_"+itoa(i), x, y, r, look, tone, false)
		o.Weight = uint8(min(amp, 255))
		f.Objects = append(f.Objects, o)
	}
	return f
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [8]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	return string(b[n:])
}
