package mapmodel

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

const framesPerHour = 60 * 60 * 8

// sightings lists the visits starting in the first n frames.
func sightings(seed int64, late bool, n int) []Sighting {
	var out []Sighting
	for b := 0; b*SightingBlock < n; b++ {
		if s, ok := sightingIn(seed, late, b); ok && s.Start < n {
			out = append(out, s)
		}
	}
	return out
}

// TestSightingSchedule: the visitor's schedule is a pure function of the
// seed, the age and the frame. From the Space Age it comes about once every
// 60 to 120 minutes for 10 to 30 seconds; before it, about once in 10 hours.
func TestSightingSchedule(t *testing.T) {
	for _, seed := range []int64{1, 7, 42, -3} {
		late := sightings(seed, true, 100*framesPerHour)
		if n := len(late); n < 100*60/120 || n > 100*60/60 {
			t.Errorf("seed %d: %d visits in 100 hours from the Space Age, want one per 60 to 120 minutes", seed, n)
		}
		early := sightings(seed, false, 1000*framesPerHour)
		if n := len(early); n < 70 || n > 130 {
			t.Errorf("seed %d: %d visits in 1000 hours before the Space Age, want about 100", seed, n)
		}
		kinds := map[SightingKind]int{}
		for i, s := range late {
			kinds[s.Kind]++
			if s.Frames < 10*8 || s.Frames > 30*8 {
				t.Errorf("seed %d: a visit lasts %d frames", seed, s.Frames)
			}
			if i > 0 && s.Start < late[i-1].Start+late[i-1].Frames {
				t.Errorf("seed %d: visits overlap at frame %d", seed, s.Start)
			}
			// SightingAt sees it through its whole length, and not a frame more
			for _, f := range []int{s.Start, s.Start + s.Frames/2, s.Start + s.Frames - 1} {
				if got, ok := SightingAt(seed, true, f); !ok || got != s {
					t.Errorf("seed %d: frame %d inside a visit reports %v %v", seed, f, got, ok)
				}
			}
			for _, f := range []int{s.Start - 1, s.Start + s.Frames} {
				if _, ok := SightingAt(seed, true, f); ok {
					t.Errorf("seed %d: frame %d outside a visit reports one", seed, f)
				}
			}
		}
		for _, k := range []SightingKind{SightFlyby, SightHover, SightWalker} {
			if kinds[k] == 0 {
				t.Errorf("seed %d: no %d visits in 100 hours", seed, k)
			}
		}
		// the joke visits are among the scheduled blocks, nothing new
		for _, s := range early {
			if l, ok := sightingIn(seed, true, s.Start/SightingBlock); !ok || l != s {
				t.Errorf("seed %d: an early visit at %d is not the late schedule's", seed, s.Start)
			}
		}
		// deterministic: the same answers again
		again := sightings(seed, true, 100*framesPerHour)
		for i := range late {
			if late[i] != again[i] {
				t.Fatalf("seed %d: visit %d changed between calls", seed, i)
			}
		}
		if s := NextSighting(seed, true, 0); s != late[0] {
			t.Errorf("seed %d: NextSighting %v, want %v", seed, s, late[0])
		}
	}
	if _, ok := SightingAt(1, true, -5); ok {
		t.Error("a visit before the map opened")
	}
	if a, b := sightings(1, true, 100*framesPerHour), sightings(2, true, 100*framesPerHour); a[0] == b[0] {
		t.Error("two seeds share a schedule")
	}
}

// TestModelSighting: the model asks for the late schedule from the Space Age
// on and the long-odds one before it.
func TestModelSighting(t *testing.T) {
	cat := NewCatalog()
	sa := cat.SpaceAge()
	if sa < 0 || config.AgeOrder()[sa] != "space_age" {
		t.Fatalf("SpaceAge %d", sa)
	}
	for _, c := range []struct {
		age  int
		late bool
	}{{0, false}, {sa - 1, false}, {sa, true}, {len(cat.Ages) - 1, true}} {
		m := &Model{Seed: 9, AgeIdx: c.age, Catalog: cat}
		for f := 0; f < 20*SightingBlock; f += 97 {
			got, ok := m.SightingAt(f)
			want, wok := SightingAt(9, c.late, f)
			if got != want || ok != wok {
				t.Fatalf("age %d frame %d: model %v %v, schedule %v %v", c.age, f, got, ok, want, wok)
			}
		}
	}
}
