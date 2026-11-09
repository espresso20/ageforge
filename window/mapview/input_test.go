package mapview

import (
	"image"
	"math"
	"testing"
	"time"

	"github.com/espresso20/ageforge/world"
)

const tw, th = 300, 170

// settle runs the easing to rest.
func settle(v *View) {
	for i := 0; i < 600; i++ {
		v.Update(16 * time.Millisecond)
	}
}

// at puts the camera at z, centred on pos of chain frame fi (which must be the master there).
func at(v *View, z float64, fi int, pos world.Vec) {
	v.z, v.tz = z, z
	v.m = v.ch.master(z)
	v.off = v.ch.centreIn(fi, v.m, pos)
	v.dirty = true
}

// zFor is a zoom inside chain frame fi alone (no crossfade), middle of its stretch.
func zFor(v *View, fi int) (float64, bool) {
	lo, hi := v.ch.zRange()
	var zs []float64
	for z := hi; z >= lo; z -= 0.02 {
		if u, o, _ := v.ch.layers(z); u == o && v.ch.master(z) == fi {
			zs = append(zs, z)
		}
	}
	if len(zs) == 0 {
		return 0, false
	}
	return zs[len(zs)/2], true
}

// centreOf is a sensible centre for chain frame fi: Sol where the frame has it.
func centreOf(v *View, fi int) world.Vec {
	p, _ := v.ch.objPos(fi, "sol")
	return p
}

// screenOf is where chain frame fi draws o on a canvas tw by th.
func screenOf(v *View, fi int, o *world.Object) (float64, float64) {
	p := v.penFor(&image.RGBA{Rect: image.Rect(0, 0, tw, th)}, fi, v.snapped(tw))
	return p.at(o.Pos)
}

func TestWheelKeepsThePointUnderTheCursor(t *testing.T) {
	for _, age := range []string{"galactic_age", "information_age"} {
		v := New()
		v.Set(build(t, age), 1, testPal)
		for fi := range v.ch {
			z, ok := zFor(v, fi)
			if !ok {
				continue
			}
			at(v, z, fi, centreOf(v, fi))
			x, y := 175, 95
			before, _, _ := v.cursorPoint(x, y, tw, th)
			hm := v.m
			v.Wheel(x, y, tw, th, 1)
			settle(v)
			after, _, _ := v.cursorPoint(x, y, tw, th)
			want := v.ch.centreIn(hm, v.m, before)
			s := v.width() / v.ch[v.m].mx / tw // frame units per art pixel
			if math.Hypot(after.X-want.X, after.Y-want.Y) > 1.5*s {
				t.Errorf("%s/%s: the point under the cursor moved from %v to %v", age, v.ch[fi].f.Key, want, after)
			}
			if v.z == v.tz && v.tz >= z {
				t.Errorf("%s/%s: the wheel did not zoom in", age, v.ch[fi].f.Key)
			}
		}
	}
}

func TestDragMovesTheSkyByThePointerDistance(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	z, _ := zFor(v, 1)
	at(v, z, 1, centreOf(v, 1))
	o := &v.ch[v.m].f.Objects[0]
	x0, y0 := screenOf(v, v.m, o)
	v.Pan(13, -7, tw)
	x1, y1 := screenOf(v, v.m, o)
	if math.Abs(x1-x0-13) > 1 || math.Abs(y1-y0+7) > 1 {
		t.Errorf("object moved by %.2f,%.2f, the pointer by 13,-7", x1-x0, y1-y0)
	}
}

func TestPickRoundTrips(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	check := func(z float64, label string) {
		under, over, _ := v.ch.layers(z)
		for _, fi := range [...]int{under, over} {
			for j := range v.ch[fi].f.Objects {
				o := &v.ch[fi].f.Objects[j]
				if !o.Pick {
					continue
				}
				at(v, z, v.m, v.off)
				// centre near the object so it is on the canvas
				v.off = v.ch.centreIn(fi, v.m, o.Pos)
				x, y := screenOf(v, fi, o)
				if x < 0 || y < 0 || x >= tw || y >= th {
					continue
				}
				got, ok := v.near(int(math.Round(x)), int(math.Round(y)), tw, th)
				if !ok {
					t.Errorf("%s: nothing picked at %s/%s", label, v.ch[fi].f.Key, o.Key)
					continue
				}
				gx, gy := screenOf(v, got.fi, v.objAt(got.fi, got.key))
				if math.Hypot(gx-x, gy-y) > 1.5 {
					t.Errorf("%s: %s/%s picked %s/%s, %.1f px away", label, v.ch[fi].f.Key, o.Key, v.ch[got.fi].f.Key, got.key, math.Hypot(gx-x, gy-y))
				}
			}
		}
	}
	for fi := range v.ch {
		if z, ok := zFor(v, fi); ok {
			check(z, v.ch[fi].f.Key)
		}
	}
	for fi := 1; fi < len(v.ch); fi++ { // inside each handoff two layers are on screen
		check((v.ch[fi].lo+v.ch[fi].hi)/2, v.ch[fi].f.Key+" mid-handoff")
		check(v.ch[fi].lo+0.1*(v.ch[fi].hi-v.ch[fi].lo), v.ch[fi].f.Key+" crossfade")
	}
}

func TestClickSelectsAndEmptyClickClears(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	var got []State
	v.OnChange(func(s State) { got = append(got, s) })
	z, _ := zFor(v, 1)
	at(v, z, 1, centreOf(v, 1))
	var o *world.Object
	for j := range v.ch[v.m].f.Objects {
		if c := &v.ch[v.m].f.Objects[j]; c.Pick {
			o = c
			break
		}
	}
	if o == nil {
		t.Skip("no pickable object")
	}
	v.off = o.Pos
	x, y := screenOf(v, v.m, o)
	v.Click(int(math.Round(x)), int(math.Round(y)), tw, th)
	if len(got) == 0 || got[len(got)-1].Sel == nil {
		t.Fatalf("a click on %s reported no selection: %v", o.Key, got)
	}
	v.Click(0, 0, tw, th)
	if got[len(got)-1].Sel != nil {
		t.Error("a click on empty sky kept the selection")
	}
}

type boxFace struct{}

func (boxFace) Mask(s string) *image.Alpha {
	m := image.NewAlpha(image.Rect(0, 0, 6*len(s), 9))
	for i := range m.Pix {
		m.Pix[i] = 255
	}
	return m
}

func TestLabelsYieldToTheHigherPriority(t *testing.T) {
	f := boxFace{}
	ls := []label{
		{x: 10, y: 50, r: 2, text: "low", mask: f.Mask("low"), prio: 10},
		{x: 12, y: 52, r: 2, text: "high", mask: f.Mask("high"), prio: 90},
		{x: 150, y: 100, r: 2, text: "far", mask: f.Mask("far"), prio: 5},
	}
	shown, at := place(ls, tw, th)
	names := map[string]image.Point{}
	for i, l := range shown {
		names[l.text] = at[i]
	}
	if _, ok := names["high"]; !ok {
		t.Error("the higher priority label was dropped")
	}
	if _, ok := names["low"]; ok {
		t.Error("the lower priority label overlapped and was kept")
	}
	if _, ok := names["far"]; !ok {
		t.Error("a label clear of the rest was dropped")
	}
	for i := range shown {
		for j := i + 1; j < len(shown); j++ {
			a := image.Rectangle{Min: at[i], Max: at[i].Add(shown[i].mask.Bounds().Size())}
			b := image.Rectangle{Min: at[j], Max: at[j].Add(shown[j].mask.Bounds().Size())}
			if a.Overlaps(b) {
				t.Errorf("%s and %s overlap", shown[i].text, shown[j].text)
			}
		}
	}
}

func TestLabelFlipsLeftAtTheEdgeAndDropsOffCanvas(t *testing.T) {
	f := boxFace{}
	shown, at := place([]label{{x: 290, y: 20, r: 2, text: "edge", mask: f.Mask("edge"), prio: 1}}, tw, th)
	if len(shown) != 1 || at[0].X+24 > tw {
		t.Errorf("edge label not flipped inside: %v %v", shown, at)
	}
	if s, _ := place([]label{{x: 10, y: 20, r: 2, text: "wide label wide", mask: f.Mask("wide label wide"), prio: 1}}, 40, th); len(s) != 0 {
		t.Error("a label that fits nowhere was kept")
	}
}

func TestLabelsReachTheCanvas(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	v.SetFace(boxFace{})
	v.SetNames(func(k string) string { return k })
	z, _ := zFor(v, 0)
	at(v, z, 0, centreOf(v, 0))
	with := v.Render(tw, th)
	a := append([]byte(nil), with.Pix...)
	v.SetFace(nil)
	without := v.Render(tw, th)
	same := true
	for i := range a {
		same = same && a[i] == without.Pix[i]
	}
	if same {
		t.Error("labels on and off drew the same canvas at the Local Group")
	}
}
