package skyline

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// TestConstructionFitsTheEra: what goes up on the frontier while buildings
// are queued follows the era. Adam saw a tower crane in the Primitive Age:
// now the first ages raise poles and stick frames, the Iron Age to the
// Colonial timber scaffolding and wooden jib cranes, and only the
// Industrial Age on has tower cranes.
func TestConstructionFitsTheEra(t *testing.T) {
	cases := []struct {
		age       string
		want, not string // runes the markers must include, and must not
	}{
		{"primitive_age", "╱╲┬", "╫═┴┌┐├┤┊"},
		{"bronze_age", "╱╲┬", "╫═┴┌┐├┤┊"},
		{"iron_age", "┌┬├┤┊", "╫═┴╱╲"},
		{"medieval_age", "┌┬├┤┊", "╫═┴╱╲"},
		{"colonial_age", "┌┬├┤┊", "╫═┴╱╲"},
		{"industrial_age", "╫═┴", "╱╲┊"},
		{"cyberpunk_age", "╫═┴", "╱╲┊"},
	}
	for _, c := range cases {
		st := fixture.State(fixture.Options{Age: c.age, Seed: 3})
		st.BuildQueue = nil
		idle := mapmodel.NewBuilder(catalog()).Build(&st, nil)
		st.BuildQueue = []game.BuildQueueSnapshot{
			{Name: "A", TicksLeft: 30, TotalTicks: 40}, {Name: "B", TicksLeft: 10, TotalTicks: 40},
			{Name: "C", TicksLeft: 20, TotalTicks: 40}, {Name: "D", TicksLeft: 5, TotalTicks: 40}}
		busy := mapmodel.NewBuilder(catalog()).Build(&st, nil)
		a := draw(newView(), idle, 160, 48, 6, mapmodel.TierUnicode, false)
		b := draw(newView(), busy, 160, 48, 6, mapmodel.TierUnicode, false)
		var got strings.Builder
		for y := 2; y < 47; y++ {
			for x := 0; x < 160; x++ {
				ra, _, _, _ := a.GetContent(x, y)
				rb, _, _, _ := b.GetContent(x, y)
				if ra != rb {
					got.WriteRune(rb)
				}
			}
		}
		marks := got.String()
		if !strings.ContainsAny(marks, c.want) {
			t.Errorf("%s: the frontier shows none of %q: %q", c.age, c.want, marks)
		}
		if strings.ContainsAny(marks, c.not) {
			t.Errorf("%s: the frontier shows one of %q, from another era: %q", c.age, c.not, marks)
		}
		a.Fini()
		b.Fini()
	}
}
