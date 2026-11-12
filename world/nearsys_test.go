package world

import (
	"testing"

	"github.com/espresso20/ageforge/mapmodel/fixture"
)

func TestNearSystemsAreSeeded(t *testing.T) {
	a, b := Build(model(AgeInterstellar, 7, fixture.Options{}), 7), Build(model(AgeInterstellar, 7, fixture.Options{}), 7)
	c := Build(model(AgeInterstellar, 8, fixture.Options{}), 8)
	differs := false
	for _, s := range NearStars {
		fa, fb, fc := a.System(s.Key), b.System(s.Key), c.System(s.Key)
		if fa == nil || fa.Kind != FrameSystem || fa.Anchor != s.Key {
			t.Fatalf("%s: no system", s.Key)
		}
		if a.System(s.Key) != fa {
			t.Errorf("%s: not cached", s.Key)
		}
		if len(fa.Objects) != len(fb.Objects) || len(fa.Children) != len(fb.Children) {
			t.Errorf("%s: same seed, different system", s.Key)
			continue
		}
		for i := range fa.Objects {
			if fa.Objects[i].Key != fb.Objects[i].Key || fa.Objects[i].Pos != fb.Objects[i].Pos || fa.Objects[i].Body != nil && fa.Objects[i].Body.Spec.Size != fb.Objects[i].Body.Spec.Size {
				t.Errorf("%s: object %d differs for one seed", s.Key, i)
			}
		}
		if len(fa.Objects) != len(fc.Objects) || len(fa.Children) != len(fc.Children) || fa.Objects[len(fa.Objects)-1].Pos != fc.Objects[len(fc.Objects)-1].Pos {
			differs = true
		}
	}
	if !differs {
		t.Error("two seeds laid out every system alike")
	}
	if a.System("nowhere") != nil {
		t.Error("a system for a star that is not there")
	}
}

func TestNearSystemsAreSound(t *testing.T) {
	w := Build(model(AgeInterstellar, 7, fixture.Options{}), 7)
	keys := map[string]bool{}
	for _, s := range NearStars {
		f := w.System(s.Key)
		planets := 0
		for _, o := range f.Objects {
			if keys[o.Key] {
				t.Errorf("duplicate key %s", o.Key)
			}
			keys[o.Key] = true
			if o.Kind == KindPlanet {
				planets++
				if o.Body == nil || o.Body.Orbit <= 0 || o.Body.Orbit > 40 || o.Body.Period <= 0 {
					t.Errorf("%s: bad orbit %+v", o.Key, o.Body)
				}
			}
		}
		if len(f.Children) != planets {
			t.Errorf("%s: %d planets, %d close-ups", s.Key, planets, len(f.Children))
		}
		for _, c := range f.Children {
			if c.Kind != FramePlanet || keys["f:"+c.Key] {
				t.Errorf("%s: bad close-up %s", s.Key, c.Key)
			}
			keys["f:"+c.Key] = true
		}
	}
	real := false
	for _, c := range w.System("barnards_star").Children {
		real = real || c.Key == "barnards_star_b"
	}
	if !real {
		t.Error("barnard's real planet is missing")
	}
}

func TestColoniesStandInTheirStarsSystem(t *testing.T) {
	w := Build(model(AgeInterstellar, 7, fixture.Options{}), 7)
	cols := 0
	for _, o := range w.Root.Find(FrameNeighbourhood).Objects {
		if o.Kind != KindColony {
			continue
		}
		cols++
		f := w.System(o.Label)
		settled, stations := 0, 0
		for _, p := range f.Objects {
			if p.Body != nil {
				settled += p.Body.Settled
			}
		}
		for _, c := range f.Children {
			for _, s := range c.Objects {
				if s.Kind == KindStation && s.Line == o.Line {
					stations++
				}
			}
		}
		if settled == 0 || stations == 0 {
			t.Errorf("colony %s at %s: %d settled planets, %d stations", o.Line, o.Label, settled, stations)
		}
	}
	if cols == 0 {
		t.Skip("the fixture has no colonies")
	}
}
