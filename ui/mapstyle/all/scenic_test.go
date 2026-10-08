package all

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// TestSceneInsetIsTheChrome: every standard style says which rows of its
// full view are bars (mapstyle.Scenic), and it is right: the age is named
// above the scene, the keys are listed below it, and neither is in the
// rows between. The main menu shows those rows alone, behind the menu.
func TestSceneInsetIsTheChrome(t *testing.T) {
	reg := Registry()
	for _, name := range reg.Names() {
		for _, age := range []string{"primitive_age", "bronze_age", "medieval_age", "modern_age", "space_age", "quantum_age", "transcendent_age"} {
			for _, sz := range [][2]int{{60, 16}, {66, 20}, {95, 36}, {144, 42}} {
				w, h := sz[0], sz[1]
				st := fixture.State(fixture.Options{Age: age, Seed: 7})
				m := mapmodel.NewBuilder(nil).Model(&st, nil)
				sty, _ := reg.New(name)
				sc, ok := sty.(mapstyle.Scenic)
				if !ok {
					t.Fatalf("%s does not say where its scene is", name)
				}
				for _, o := range []mapstyle.Option{mapstyle.OptFlows, mapstyle.OptInspect, mapstyle.OptLegend, mapstyle.OptChanges} {
					sty.SetOption(o, false)
				}
				sim := tcell.NewSimulationScreen("UTF-8")
				if err := sim.Init(); err != nil {
					t.Fatal(err)
				}
				sim.SetSize(w, h)
				sty.Draw(theme.WrapScreen(sim), mapstyle.Rect{W: w, H: h}, mapstyle.Frame{Model: m, Anim: 3, Clock: 3})
				rows := make([]string, h)
				for y := range rows {
					var sb strings.Builder
					for x := 0; x < w; x++ {
						r, _, _, _ := sim.GetContent(x, y)
						sb.WriteRune(r)
					}
					rows[y] = sb.String()
				}
				sim.Fini()

				top, bottom := sc.SceneInset(w, h)
				where := fmt.Sprintf("%s in the %s at %dx%d", name, st.AgeName, w, h)
				if top < 1 || bottom < 1 || top+bottom >= h {
					t.Fatalf("%s: inset %d and %d leaves no scene", where, top, bottom)
				}
				ageWord := strings.ToLower(strings.TrimSuffix(st.AgeName, " Age"))
				if head := strings.ToLower(strings.Join(rows[:top], "\n")); !strings.Contains(head, ageWord) {
					t.Errorf("%s: the %d rows above the scene do not name the age: %q", where, top, rows[:top])
				}
				isKeys := func(row string) bool {
					return strings.Contains(row, "PgUp") || strings.Contains(row, "arrows move") || strings.Contains(row, "scroll") || strings.Contains(row, "Tab ")
				}
				if !isKeys(strings.Join(rows[h-bottom:], "\n")) {
					t.Errorf("%s: the %d rows below the scene list no keys: %q", where, bottom, rows[h-bottom:])
				}
				for y := top; y < h-bottom; y++ {
					if isKeys(rows[y]) {
						t.Errorf("%s: row %d is in the scene and lists keys: %q", where, y, rows[y])
					}
				}
			}
		}
	}
}
