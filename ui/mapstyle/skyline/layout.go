package skyline

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/mapmodel"
)

// layout.go dresses the model's placed lots (mapmodel.Skyline) in sprites.
// Placement is the model's: every silhouette is centred on its Lot.X, so a
// building never moves; the sprite, its height and its finish are ours. The
// layout is cached per (model, scene height) and the sprites per (type,
// seed, height), so a frame only touches the lots on screen.

type lotView struct {
	ml       *mapmodel.Lot
	b        *mapmodel.Building
	spr      *sprite
	x0       int // world column of the sprite's left edge
	fam, vr  int
	staff    float64 // 0..1: lights and smoke follow it
	producer bool
	depth    uint8
	variant  int // 0..2: the copy's own height, windows and roof (dress.go)
}

type layout struct {
	m       *mapmodel.Model
	groundY int
	ref     int // the ground row the buildings are sized for (the mid view's)
	scale   float64
	lots    []lotView // draw order: back rows first, taller first, wonders last
	byX     []int32   // lot indexes sorted by x0
	maxW    int
	prof    []int   // skyline height per world column
	tallest []int32 // the tallest lot per world column, -1 for none
	width   int
	news    []int // world columns of lots new since the last visit
	index   map[lotID]int
	// sites are the wonders still being built (the scaffolding on their plots);
	// they are no lots: nothing can be inspected on them and the model has
	// nothing placed there.
	sites []lotView
}

type lotID struct {
	key string
	cp  int
}

type spriteKey struct {
	key  string
	seed uint64
	h    int
}

// sceneScale sizes buildings to the scene: 1 at a 45-row terminal.
func sceneScale(groundY int) float64 {
	s := float64(groundY-1) / 36
	if s < 0.42 {
		s = 0.42
	}
	if s > 1.25 {
		s = 1.25
	}
	return s
}

func staffOf(m *mapmodel.Model, b *mapmodel.Building) float64 {
	switch {
	case b == nil:
		return 0.45
	case b.Wonder:
		return 1
	case b.Lineage == mapmodel.LinHousing:
		if m.Workers.MaxPop > 0 {
			return clamp01(float64(m.Workers.Pop) / float64(m.Workers.MaxPop))
		}
		return 0.5
	case b.Staffing >= 0:
		return b.Staffing
	}
	return 0.45
}

func (v *view) layoutFor(m *mapmodel.Model, groundY int) *layout {
	return v.layoutSized(m, groundY, groundY)
}

// layoutSized is layoutFor with the buildings sized for a scene whose
// ground is at row ref: the far zoom composes a taller frame (ground at
// groundY) in which a building is still as tall as the screen's own height
// makes it.
func (v *view) layoutSized(m *mapmodel.Model, groundY, ref int) *layout {
	if v.lay != nil && v.lay.m == m && v.lay.groundY == groundY && v.lay.ref == ref {
		return v.lay
	}
	if v.sprites == nil || len(v.sprites) > 6000 {
		v.sprites = map[spriteKey]*sprite{}
	}
	lay := &layout{m: m, groundY: groundY, ref: ref, scale: sceneScale(ref), width: max(m.Skyline.Width, 1)}
	maxH := max(2, ref-3)
	// Every district's landmark is the tallest thing in it, at least half as
	// tall again as the tallest building: the buildings of an age stop short of
	// what that age's wonder stands at (districtCap).
	wonderKey := wonderKeys(m.Catalog)
	capOf := func(age int) int {
		if age < 0 || age >= len(wonderKey) || wonderKey[age] == "" {
			return maxH
		}
		return min(maxH, districtCap(wonderRows(wonderKey[age], lay.scale, ref)))
	}
	for i := range m.Skyline.Lots {
		ml := &m.Skyline.Lots[i]
		def := m.Catalog.Defs[ml.Key]
		if def == nil {
			continue
		}
		b := m.Building(ml.Key)
		lv := lotView{ml: ml, b: b, fam: familyOf(ml.Age), vr: int(hash(int(ml.Seed), 7) % 3),
			staff: staffOf(m, b), depth: rowDepth(ml.Row), variant: variantOf(ml.Seed)}
		count := 1
		if b != nil {
			count = b.Count
		}
		var sk spriteKey
		var fm form
		if ml.Wonder {
			if fam, ok := wonderFamily[ml.Key]; ok {
				lv.fam = fam
			}
			lv.depth = dWonder
			lv.vr = 0 // a landmark wears its period's standard finish, whatever the game's seed
			sk = spriteKey{ml.Key, 0, wonderRows(ml.Key, lay.scale, ref)}
		} else {
			fm = formFor(def)
			lv.producer = fm.smoke
			h := float64(fm.height * lay.scale)
			h *= (0.9 + 0.2*hashf(int(ml.Seed), 11)) * variantHeight[lv.variant]
			h *= 1 + 0.05*mapmodel.Log2(1+float64(count))
			if ml.Row > 0 {
				h *= 1.12
			}
			sk = spriteKey{ml.Key, ml.Seed, softCap(clampInt(int(h+0.5), 2, maxH), capOf(ml.Age))}
		}
		spr := v.sprites[sk]
		if spr == nil {
			if ml.Wonder {
				spr = wonderArt(ml.Key, sk.h)
			} else {
				spr = dressedUnder(fm, ml, lv.variant, sk.h, capOf(ml.Age), maxH)
				if hash(int(ml.Seed), 5)%2 == 0 {
					spr = spr.mirror()
				}
			}
			v.sprites[sk] = spr
		}
		lv.spr = spr
		lv.x0 = ml.X - spr.w/2
		lay.maxW = max(lay.maxW, spr.w)
		lay.lots = append(lay.lots, lv)
	}
	sort.SliceStable(lay.lots, func(i, j int) bool {
		a, b := &lay.lots[i], &lay.lots[j]
		if a.depth != b.depth {
			return a.depth > b.depth
		}
		if a.spr.h != b.spr.h {
			return a.spr.h > b.spr.h
		}
		return a.ml.Seed < b.ml.Seed
	})
	lay.index = make(map[lotID]int, len(lay.lots))
	for i := range lay.lots {
		ml := lay.lots[i].ml
		lay.index[lotID{ml.Key, ml.Copy}] = i
		if ml.New {
			lay.news = append(lay.news, ml.X)
		}
	}
	v.layoutSites(lay)
	lay.byX = make([]int32, len(lay.lots))
	for i := range lay.byX {
		lay.byX[i] = int32(i)
	}
	sort.SliceStable(lay.byX, func(i, j int) bool { return lay.lots[lay.byX[i]].x0 < lay.lots[lay.byX[j]].x0 })
	lay.prof = make([]int, lay.width)
	lay.tallest = make([]int32, lay.width)
	for i := range lay.tallest {
		lay.tallest[i] = -1
	}
	for i := range lay.lots {
		l := &lay.lots[i]
		for x := 0; x < l.spr.w; x++ {
			wx := l.x0 + x
			if wx < 0 || wx >= lay.width {
				continue
			}
			if t := l.spr.topAt(x); t >= 0 {
				if h := l.spr.h - t; h > lay.prof[wx] {
					lay.prof[wx] = h
					lay.tallest[wx] = int32(i)
				}
			}
		}
	}
	v.lay = lay
	return lay
}

// visible appends to out the lots overlapping world columns [x0, x1), in
// draw order.
func (l *layout) visible(x0, x1 int, out []int) []int {
	out = out[:0]
	lo := sort.Search(len(l.byX), func(i int) bool { return l.lots[l.byX[i]].x0 >= x0-l.maxW })
	for i := lo; i < len(l.byX); i++ {
		lv := &l.lots[l.byX[i]]
		if lv.x0 >= x1 {
			break
		}
		if lv.x0+lv.spr.w > x0 {
			out = append(out, int(l.byX[i]))
		}
	}
	sort.Ints(out)
	return out
}

// find returns the index of the lot for (key, copy), or -1.
func (l *layout) find(key string, cp int) int {
	if i, ok := l.index[lotID{key, cp}]; ok {
		return i
	}
	return -1
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// wonderKeys is the key of each age's wonder, by age index ("" for none).
func wonderKeys(c *mapmodel.Catalog) []string {
	out := make([]string, len(c.ByAge))
	for a, ds := range c.ByAge {
		for _, d := range ds {
			if d.Wonder {
				out[a] = d.Key
			}
		}
	}
	return out
}

// districtCap is the most rows a building of a district may stand when its
// wonder is rows tall: the wonder is at least half as tall again.
func districtCap(wonderRows int) int { return max(2, wonderRows*2/3) }

// softCap bends a building's height so it stays under capH without a row of
// flat roofs: it is unchanged under 60% of the cap, then eases toward the cap
// (reaching it only for ever taller buildings), keeping their order.
func softCap(h, capH int) int {
	knee := capH * 6 / 10
	if h <= knee {
		return h
	}
	span := float64(capH - knee)
	if span <= 0 {
		return min(h, capH)
	}
	v := float64(knee) + span*(1-math.Exp(-float64(h-knee)/span))
	return clampInt(int(v+0.5), 2, capH)
}

// dressedUnder draws a building h rows tall dressed for its variant, never
// over capH rows: a form that comes out taller than it was asked to (a spire,
// a stack) is asked again for less.
func dressedUnder(fm form, ml *mapmodel.Lot, variant, h, capH, maxH int) *sprite {
	build := func(req int) *sprite {
		return dress(fm.fn(newRnd(ml.Seed), req), variant, ml.Seed, ml.Age, maxH)
	}
	spr := build(h)
	for req := h - 1; spr.h > capH && req >= 2; req-- {
		spr = build(req)
	}
	if spr.h > capH {
		spr = spr.cropTop(capH)
	}
	return spr
}

// wonderIn is where a district's landmark stands, in columns from its east
// edge, and wonderBayIn the grand lighthouse's, in from the bay's east end:
// mapmodel places the lots, this is the same spot for a wonder that is not a
// lot yet (a test holds the two together).
const (
	wonderIn    = 18
	wonderBayIn = 16
)

// wonderX is the world column a wonder's plot is centred on in district d.
func wonderX(d mapmodel.District, key string) int {
	if d.Bay && key == "grand_lighthouse" {
		return d.BayX + wonderBayIn
	}
	return d.X0 + d.LandW - wonderIn
}

// layoutSites puts the scaffolding of every wonder under way on its plot.
func (v *view) layoutSites(lay *layout) {
	m := lay.m
	if len(m.Wonders) == 0 {
		return
	}
	lots := make([]mapmodel.Lot, 0, len(m.Wonders))
	steel := m.Catalog.AgeIdx["industrial_age"]
	for _, w := range m.Wonders {
		if w.Built || w.Age < 0 || w.Age >= len(m.Skyline.Districts) {
			continue
		}
		key := w.Key
		rows := wonderRows(key, lay.scale, lay.ref)
		built := siteRows(w.Progress, rows)
		sk := spriteKey{key + "/site", uint64(built), rows}
		spr := v.sprites[sk]
		if spr == nil {
			spr = scaffoldArt(key, rows, built, w.Age >= steel)
			v.sprites[sk] = spr
		}
		x := wonderX(m.Skyline.Districts[w.Age], key)
		lots = append(lots, mapmodel.Lot{Key: key, Lineage: mapmodel.LinWonder, Age: w.Age, X: x, Wonder: true,
			Seed: uint64(hash(len(key), int(key[0]), 17))})
		lv := lotView{ml: &lots[len(lots)-1], b: m.Building(key), spr: spr, fam: familyOf(w.Age), depth: dWonder, x0: x - spr.w/2}
		if fam, ok := wonderFamily[key]; ok {
			lv.fam = fam
		}
		lay.sites = append(lay.sites, lv)
	}
}
