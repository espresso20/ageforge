package skyline

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// portrait.go is one wonder alone: the gallery of the Wonders panel draws
// every wonder the town has built with the skyline's own renderer (the same
// forms, the same palette, the night light and the close level's fine
// detail) on a plain backdrop, as large as the panel holds. It is a
// mapstyle.Portraitist.

// portraitDusk is the hour the portrait is lit at: the light of the lamps
// is on and the sky is still blue.
const portraitDusk = 0.80

// Portrait draws the wonder key into r, every cell of it, and reports false
// for a key that is no wonder of the model's catalogue. The backdrop is that
// wonder's own age's: its sky at dusk, clear, and a plain strip of ground.
func (v *view) Portrait(scr tcell.Screen, r mapstyle.Rect, f mapstyle.Frame, key string) bool {
	m := f.Model
	if m == nil || r.W < 8 || r.H < 6 {
		return false
	}
	def := m.Catalog.Defs[key]
	if def == nil || !def.Wonder {
		return false
	}
	W, H := r.W, r.H
	v.fb.reset(W, H)
	pm := *m // the world as that age had it: its epoch, a clear dusk
	pm.AgeIdx, pm.Age = def.Age, m.Catalog.Ages[def.Age]
	pm.AgeName = m.Catalog.AgeNames[def.Age]
	pm.Epoch = m.Catalog.EpochOfAge(def.Age)
	pm.Clock = mapmodel.ClockAtTOD(3, portraitDusk)
	pm.Weather = mapmodel.Weather{}
	pm.Catastrophe = mapmodel.Catastrophe{}
	_, mp := v.palettes(m)
	s := &scene{v: v, m: &pm, fb: &v.fb, tier: f.Tier, anim: f.Anim, W: W, H: H, S: H, sel: -1, detail: true,
		lw: W, ls: H, sw: W, sh: H, zm: zMid}
	s.p = newPal(&pm)
	s.mp = mp
	s.band = bandOf(pm.AgeIdx)
	s.city, s.inCity = mapmodel.CityLookAt(pm.AgeIdx)
	s.groundY = max(4, H-3)
	s.sky()
	s.celestial()

	// the wonder, as large as the room holds: the sprite is asked for fewer
	// rows until it fits the width
	rows := clampInt(s.groundY-1, wonderMinRows, 60)
	spr := wonderArt(key, rows)
	for spr.w > W-2 && rows > wonderMinRows {
		rows--
		spr = wonderArt(key, rows)
	}
	fam := familyOf(def.Age)
	if wf, ok := wonderFamily[key]; ok {
		fam = wf
	}
	lot := mapmodel.Lot{Key: key, Lineage: mapmodel.LinWonder, Age: def.Age, Wonder: true, X: W / 2,
		Seed: uint64(hash(len(key), int(key[0]), 17))}
	lv := lotView{ml: &lot, b: pm.Building(key), spr: spr, fam: fam, depth: dWonder, staff: 1, x0: W/2 - spr.w/2}
	lay := &layout{m: &pm, groundY: s.groundY, ref: s.groundY, scale: 1, width: W, lots: []lotView{lv}, byX: []int32{0}, maxW: spr.w}
	s.lay = lay
	s.vis = []int{0}
	s.lots()
	s.portraitGround(def.Age)
	v.fb.blit(scr, r, f.Tier)
	return true
}

// portraitGround is the strip of ground under a portrait: the age's own
// grass, cobble or slab, a road and a verge.
func (s *scene) portraitGround(age int) {
	p := s.p
	gy := s.groundY
	rh, mh := roadHues(age)
	rc, mc := p.hue(rh, mLit, 0), p.hue(mh, mLit, 0)
	top, under := theme.SkyGroundGrass, theme.SkyGroundSoil
	switch {
	case age > 11:
		top, under = theme.SkyGroundSlab, theme.SkyGroundSlabDark
	case age > 7:
		top, under = theme.SkyGroundBrick, theme.SkyGroundBrickDark
	case age > 3:
		top, under = theme.SkyGroundCobble, theme.SkyGroundCobbleDark
	}
	vb := p.hue(theme.SkyVergeDark, mLit, 0)
	for x := 0; x < s.W; x++ {
		s.fb.set(x, s.Y(gy), '▀', p.hue(top, mLit, 0), p.hue(under, mLit, 0), dGround)
		s.fb.set(x, s.Y(gy+1), ' ', mc, rc, dGround)
		for y := gy + 2; y < s.S; y++ {
			s.fb.fill(x, s.Y(y), vb, dGround)
		}
	}
}
