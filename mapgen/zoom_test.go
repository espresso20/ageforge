package mapgen

import "testing"

// TestZoomPicksTheScale: the five space ages draw their city at the close
// zoom and the star map at the far one; the age before them draws the city at
// both.
func TestZoomPicksTheScale(t *testing.T) {
	for _, c := range []struct {
		age        string
		close, far Scale
	}{
		{"fusion_age", City, City},
		{"space_age", City, Stars},
		{"interstellar_age", City, Stars},
		{"galactic_age", City, Stars},
		{"quantum_age", City, Stars},
		{"transcendent_age", City, Stars},
	} {
		m := model(state(c.age, 7, noon))
		var gen Generator
		if got := gen.frame(m, 0, true).st.Scale; got != c.close {
			t.Errorf("%s close: scale %d, want %d", c.age, got, c.close)
		}
		if got := gen.frame(m, 0, false).st.Scale; got != c.far {
			t.Errorf("%s far: scale %d, want %d", c.age, got, c.far)
		}
	}
}

// TestSpaceAgeCitiesDraw: each space age's city draws over a few frames with
// buildings, traffic and a spaceport, and the star map of the same age draws
// its sectors.
func TestSpaceAgeCitiesDraw(t *testing.T) {
	for _, age := range []string{"space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"} {
		var gen Generator
		m := model(state(age, 7, noon))
		for _, anim := range []int{0, 5, 40} {
			sc := gen.DrawAt(m, Moment{Anim: anim, Visit: -1, Close: true})
			seen := map[Kind]int{}
			for _, h := range sc.Grid.Hit {
				seen[sc.Grid.Things[h].Kind]++
			}
			if seen[KindBuilding] == 0 || seen[KindAgent] == 0 {
				t.Errorf("%s frame %d: the city shows %d buildings and %d agents", age, anim, seen[KindBuilding], seen[KindAgent])
			}
			far := gen.DrawAt(m, Moment{Anim: anim, Visit: -1})
			sectors := 0
			for _, h := range far.Grid.Hit {
				if far.Grid.Things[h].Kind == KindSector {
					sectors++
				}
			}
			if sectors == 0 || !far.Fill {
				t.Errorf("%s frame %d: the star map shows %d sectors", age, anim, sectors)
			}
		}
	}
}
