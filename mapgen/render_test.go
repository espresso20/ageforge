package mapgen

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// The three ages the lab has rows for: one early ground age, one late Earth
// city age, one space age.
var testAges = []string{"bronze_age", "modern_age", "space_age"}

const (
	noon     = 225 + 1800*5
	midnight = 1125 + 1800*5
)

func model(st game.GameState) *mapmodel.Model {
	return mapmodel.NewBuilder(testCatalog).Build(&st, nil)
}

func state(age string, seed int64, tick int) game.GameState {
	return fixture.State(fixture.Options{Age: age, Seed: seed, Tick: tick})
}

// TestRenderDeterministic: the same input draws the same grid, cell for cell
// and thing for thing, from a fresh generator or a warm one.
func TestRenderDeterministic(t *testing.T) {
	for _, age := range testAges {
		for _, tick := range []int{noon, midnight} {
			m := model(state(age, 7, tick))
			var warm Generator
			warm.Render(m, 80, 24, 3)
			a := warm.Render(m, 120, 36, 5)
			b := (&Generator{}).Render(m, 120, 36, 5)
			if !reflect.DeepEqual(a.Cells, b.Cells) || !reflect.DeepEqual(a.Hit, b.Hit) || !reflect.DeepEqual(a.Things, b.Things) {
				t.Errorf("%s tick %d: two renders of one state differ", age, tick)
			}
		}
	}
}

// TestRenderMoves: traffic and water move with the frame, and the seed makes
// a different picture.
func TestRenderMoves(t *testing.T) {
	for _, age := range testAges {
		m := model(state(age, 7, noon))
		var gen Generator
		if reflect.DeepEqual(gen.Render(m, 120, 36, 0).Cells, gen.Render(m, 120, 36, 40).Cells) {
			t.Errorf("%s: frames 0 and 40 are the same picture", age)
		}
		other := model(state(age, 8, noon))
		if reflect.DeepEqual(gen.Render(m, 120, 36, 0).Cells, gen.Render(other, 120, 36, 0).Cells) {
			t.Errorf("%s: seeds 7 and 8 are the same picture", age)
		}
	}
}

// TestNightIsDarker: the same place at midnight is dimmer than at noon, and
// still has lit windows or lamps.
func TestNightIsDarker(t *testing.T) {
	for _, age := range testAges {
		day := (&Generator{}).Render(model(state(age, 7, noon)), 120, 36, 0)
		night := (&Generator{}).Render(model(state(age, 7, midnight)), 120, 36, 0)
		if brightness(night) >= brightness(day) {
			t.Errorf("%s: night (%d) is not darker than day (%d)", age, brightness(night), brightness(day))
		}
	}
}

func brightness(g *Grid) int {
	sum := 0
	for _, c := range g.Cells {
		sum += int(c.Bg.R) + int(c.Bg.G) + int(c.Bg.B)
	}
	return sum / len(g.Cells)
}

// worldFrame draws the whole world, so cell positions are world positions.
func worldFrame(gen *Generator, m *mapmodel.Model) *frame {
	f := gen.frame(m, 0)
	f.draw()
	return f
}

// TestNothingStandsInWaterOrOverlaps: every building cell is on dry ground
// (on the ground scenes), and no cell of one building is another's.
func TestNothingStandsInWaterOrOverlaps(t *testing.T) {
	for _, age := range testAges {
		var gen Generator
		f := worldFrame(&gen, model(state(age, 7, noon)))
		parcels := map[int]string{}
		for _, pl := range f.placed {
			if prev, taken := parcels[pl.lot]; taken {
				t.Fatalf("%s: parcel %d holds both %s and %s", age, pl.lot, prev, pl.b.Key)
			}
			parcels[pl.lot] = pl.b.Key
			if f.fields != nil {
				f.plan.Lots[pl.lot].Rect.cells(func(x, y int) {
					if f.fields.Biome[at(x, y)].Water() {
						t.Fatalf("%s: %s stands in water at (%d,%d)", age, pl.b.Key, x, y)
					}
				})
			}
		}
		for i := range f.g.Cells {
			th := f.g.Things[f.g.Hit[i]]
			if th.Kind != KindBuilding || th.Ruin {
				continue
			}
			if !f.busy[i] {
				t.Fatalf("%s: a building cell at (%d,%d) is not on a parcel", age, i%worldW, i/worldW)
			}
		}
	}
}

// TestEveryBuildingOnARoad: the road that serves each parcel in use is laid
// as far as the parcel's door, and the laid roads are one piece.
func TestEveryBuildingOnARoad(t *testing.T) {
	for _, age := range testAges {
		var gen Generator
		f := worldFrame(&gen, model(state(age, 7, noon)))
		laid := map[Pt]bool{}
		for i, r := range f.plan.Roads {
			for _, c := range r.Cells[:f.shown[i]] {
				laid[c] = true
			}
		}
		for _, pl := range f.placed {
			if door := f.plan.Lots[pl.lot].Door; !laid[door] {
				t.Fatalf("%s: the door of %s, %v, is on no laid road", age, pl.b.Key, door)
			}
		}
		reached := map[Pt]bool{}
		start := f.plan.Roads[0].Cells[0]
		queue := []Pt{start}
		reached[start] = true
		for head := 0; head < len(queue); head++ {
			for _, d := range neighbours {
				n := Pt{queue[head].X + d[0], queue[head].Y + d[1]}
				if laid[n] && !reached[n] {
					reached[n] = true
					queue = append(queue, n)
				}
			}
		}
		if len(reached) != len(laid) {
			t.Errorf("%s: %d laid road cells, %d reached from the square", age, len(laid), len(reached))
		}
	}
}

// standing lists where each building kind's parcels are, and which kind rank
// holds each parcel.
func standing(gen *Generator, st game.GameState) (at map[string]int, rankAt map[int]int) {
	m := model(st)
	f := gen.frame(m, 0)
	at, rankAt = map[string]int{}, map[int]int{}
	for _, pl := range f.placed {
		if !pl.ruin {
			at[fmt.Sprintf("%s/%d", pl.b.Key, pl.instance)] = pl.lot
		}
		rankAt[pl.lot] = pl.b.Rank
	}
	return at, rankAt
}

// TestPlacementNeverMoves builds a town one building at a time, new kinds and
// extra copies of old ones in any order, and checks that no building moves:
// each parcel in use stays where it was, unless a newer kind has since
// taken the parcel over.
func TestPlacementNeverMoves(t *testing.T) {
	for _, age := range testAges {
		for _, seed := range []int64{7, 23} {
			st := state(age, seed, noon)
			defs := testCatalog.Rules.BuildingMap()
			var kinds []string
			for _, d := range testCatalog.ByAge[testCatalog.AgeIdx[age]] {
				if !d.Wonder {
					kinds = append(kinds, d.Key)
				}
			}
			slices.Sort(kinds)
			var gen Generator
			prev, prevRank := standing(&gen, st)
			for step := 0; step < 60; step++ {
				k := kinds[roll(len(kinds), seed, 90, int64(step))]
				bs, ok := st.Buildings[k]
				if !ok {
					d := defs[k]
					bs = game.BuildingState{Name: d.Name, Category: d.Category, AgeKey: d.RequiredAge, WorkerCapacity: d.WorkerCapacity, Unlocked: true}
				}
				bs.Count += 1 + roll(3, seed, 91, int64(step))
				st.Buildings[k] = bs
				cur, curRank := standing(&gen, st)
				for key, lot := range prev {
					now, held := cur[key]
					switch {
					case held && now != lot:
						t.Fatalf("%s/%d step %d: %s moved from parcel %d to %d", age, seed, step, key, lot, now)
					case !held && curRank[lot] <= prevRank[lot]:
						t.Fatalf("%s/%d step %d: %s left parcel %d for no newer kind", age, seed, step, key, lot)
					}
				}
				prev, prevRank = cur, curRank
			}
		}
	}
}

// TestSizeOnlyGrows: a parcel's building only gets bigger as copies are added.
func TestSizeOnlyGrows(t *testing.T) {
	last := 0
	for n := 0; n <= 200; n++ {
		parcels, tier := growthFor(n)
		if tier < last {
			t.Fatalf("%d copies: tier %d after %d", n, tier, last)
		}
		last = tier
		if p2, _ := growthFor(n + 1); p2 < parcels {
			t.Fatalf("%d copies: parcels fell from %d to %d", n, parcels, p2)
		}
	}
}

// TestInspectIndex: building and wonder cells point at the kind drawn there.
func TestInspectIndex(t *testing.T) {
	for _, age := range testAges {
		m := model(state(age, 7, noon))
		g := (&Generator{}).Render(m, 120, 36, 0)
		seen := map[Kind]int{}
		for y := 0; y < g.H; y++ {
			for x := 0; x < g.W; x++ {
				th := g.ThingAt(x, y)
				seen[th.Kind]++
				if th.Kind == KindBuilding && m.Building(th.Key) == nil {
					t.Fatalf("%s: (%d,%d) shows a building %q the model does not have", age, x, y, th.Key)
				}
			}
		}
		if seen[KindBuilding] == 0 || seen[KindTerrain] == 0 {
			t.Errorf("%s: the grid shows %v", age, seen)
		}
	}
}

// TestEveryAgeDraws: every age of the game draws, at both sizes, with its
// buildings in view.
func TestEveryAgeDraws(t *testing.T) {
	var gen Generator
	for i, age := range testCatalog.Ages {
		m := model(state(age, 7, noon))
		for _, size := range [][2]int{{120, 36}, {80, 24}} {
			g := gen.Render(m, size[0], size[1], i)
			if g.W != size[0] || g.H != size[1] {
				t.Fatalf("%s: drew %dx%d, asked for %dx%d", age, g.W, g.H, size[0], size[1])
			}
			shown := 0
			for _, th := range g.Hit {
				if g.Things[th].Kind == KindBuilding {
					shown++
				}
			}
			if shown == 0 {
				t.Errorf("%s at %dx%d: no building in view", age, size[0], size[1])
			}
		}
	}
}
