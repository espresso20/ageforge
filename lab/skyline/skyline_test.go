package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/theme"
)

// The saved states live next to the package; tests run from its directory.
const testStates = "states"

func testNames(t *testing.T) []string {
	var out []string
	for _, n := range availableStates(testStates) {
		if !strings.HasPrefix(n, "prev_") {
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		t.Skip("no saved states; run go run ./lab/skyline -gen")
	}
	return out
}

// Every state renders at every size, theme and camera without panicking,
// and fills exactly the frame it was given.
func TestRenderAllStatesAllSizes(t *testing.T) {
	sizes := [][2]int{{40, 15}, {80, 24}, {100, 30}, {160, 45}, {230, 62}}
	themes := []string{"forge", "daylight", "monochrome"}
	for _, n := range testNames(t) {
		for _, sz := range sizes {
			for _, tk := range themes {
				th, _ := theme.ByKey(tk)
				compact := sz[0] <= 40
				sh := sceneRows(sz[1])
				if compact {
					sh = sz[1]
				}
				w, err := loadWorld(testStates, n, sh)
				if err != nil {
					t.Fatal(err)
				}
				for _, cam := range []int{-50, 0, w.W / 2, camFor(w, sz[0], "end"), w.W + 10} {
					for _, tod := range []float64{0.1, 0.26, 0.5, 0.75} {
						v := View{W: sz[0], H: sz[1], Cam: cam, TOD: tod, Theme: th, Cursor: sz[0] / 2, Weather: -1, ShowNew: true, Frame: 7}
						var fb *FB
						if compact {
							fb = renderCompact(w, v)
						} else {
							fb = render(w, v)
						}
						if fb.W != sz[0] || fb.H != sz[1] || len(fb.C) != sz[0]*sz[1] {
							t.Fatalf("%s %v: frame is %dx%d", n, sz, fb.W, fb.H)
						}
					}
				}
			}
		}
	}
}

// The same state, view and frame always draws the same cells.
func TestDeterministic(t *testing.T) {
	names := testNames(t)
	th, _ := theme.ByKey("forge")
	for _, n := range names[:min(4, len(names))] {
		a, _ := loadWorld(testStates, n, 42)
		b, _ := loadWorld(testStates, n, 42)
		v := View{W: 160, H: 45, Cam: camFor(a, 160, "end"), TOD: 0.9, Theme: th, Cursor: -1, Weather: -1, Frame: 12}
		fa, fb := render(a, v), render(b, v)
		for i := range fa.C {
			if fa.C[i] != fb.C[i] {
				t.Fatalf("%s: cell %d differs between two renders", n, i)
			}
		}
	}
}

// Adding a building adds a silhouette and moves nothing else: every lot
// that existed before keeps its column and depth.
func TestPlacementIsStable(t *testing.T) {
	for _, n := range testNames(t) {
		st, err := loadState(testStates, n)
		if err != nil {
			t.Fatal(err)
		}
		before := buildWorld(st, nil, 42)
		pos := map[string][2]int{}
		for _, l := range before.Lots {
			pos[l.Key+"#"+string(rune('0'+l.Copy))] = [2]int{l.X + l.Spr.W/2, l.Row}
		}
		// grow every type the player owns threefold (new copies appear)
		for k, b := range st.Buildings {
			if b.Count > 0 {
				b.Count *= 3
				st.Buildings[k] = b
			}
		}
		after := buildWorld(st, nil, 42)
		for _, l := range after.Lots {
			if p, ok := pos[l.Key+"#"+string(rune('0'+l.Copy))]; ok && (abs(p[0]-(l.X+l.Spr.W/2)) > 1 || p[1] != l.Row) {
				t.Fatalf("%s: %s copy %d moved from %v to (%d,%d)", n, l.Key, l.Copy, p, l.X, l.Row)
			}
		}
	}
}

func BenchmarkRenderFrame(b *testing.B) {
	th, _ := theme.ByKey("forge")
	w, err := loadWorld(testStates, "cyberpunk_age", sceneRows(45))
	if err != nil {
		b.Skip()
	}
	v := View{W: 160, H: 45, Cam: camFor(w, 160, "end"), TOD: 0.95, Theme: th, Cursor: -1, Weather: -1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Frame = i
		render(w, v)
	}
}

func BenchmarkBuildWorld(b *testing.B) {
	st, err := loadState(testStates, "cyberpunk_age")
	if err != nil {
		b.Skip()
	}
	for i := 0; i < b.N; i++ {
		buildWorld(st, nil, 42)
	}
}

func TestMain(m *testing.M) {
	if _, err := os.Stat(filepath.Join(testStates)); err != nil {
		os.Exit(0)
	}
	os.Exit(m.Run())
}
