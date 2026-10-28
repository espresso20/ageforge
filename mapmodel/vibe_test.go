package mapmodel_test

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/vibe"
)

// told puts every text the code writes itself in guillemets while VibeMode is
// on, so a text that follows the switch is plain to find.
func told(t *testing.T) {
	t.Helper()
	was := vibe.On()
	t.Cleanup(func() { vibe.Set(was) })
	vibe.Set(false)
	t.Cleanup(vibe.UseRewriterForTest(func(id, plain string) string { return "«" + plain + "»" }))
}

// TestTheLineageNamesReadTheScript: the words of the lineages are read when
// they are asked for, so one table follows the switch both ways.
func TestTheLineageNamesReadTheScript(t *testing.T) {
	told(t)
	words := func() map[string]string {
		out := map[string]string{}
		for _, lin := range mapmodel.LineageOrder {
			out["lineage "+lin] = mapmodel.LineageName(lin)
		}
		out["war"] = mapmodel.RelWar.String()
		return out
	}
	plain := words()
	for k, v := range plain {
		if strings.Contains(v, "«") {
			t.Errorf("%s reads %q with VibeMode off", k, v)
		}
	}
	vibe.Set(true)
	for k, v := range words() {
		if v == plain[k] || !strings.Contains(v, "«") {
			t.Errorf("%s reads %q with VibeMode on, was %q", k, v, plain[k])
		}
	}
	vibe.Set(false)
	for k, v := range words() {
		if v != plain[k] {
			t.Errorf("%s reads %q with VibeMode off again, was %q", k, v, plain[k])
		}
	}
}

// TestTheBuilderFollowsTheSwitch: the one builder, which caches the last model,
// builds again when VibeMode is switched, so the words
// a model holds (the news, what is stuck, the name of the town) are in the
// script now in force, and plain again when it is off.
func TestTheBuilderFollowsTheSwitch(t *testing.T) {
	told(t)
	st := fixture.State(fixture.Options{Age: "bronze_age", Seed: 7, Catastrophe: true, Harbinger: true, Wars: 1, Idle: 3})
	prev := fixture.State(fixture.Options{Age: "stone_age", Seed: 7})
	b := mapmodel.NewBuilder(nil)
	since := mapmodel.VisitOf(b.Build(&prev, nil))
	words := func() map[string]string {
		m := b.Model(&st, since)
		out := map[string]string{
			"town":     m.TownName,
			"stuck":    m.Flows.Worst,
			"headline": m.Recap.Headline(400),
		}
		for i, it := range m.Recap.Items {
			out["news "+string(rune('0'+i))] = it.Text
		}
		return out
	}
	plain := words()
	for _, k := range []string{"town", "stuck", "headline", "news 0"} {
		if plain[k] == "" {
			t.Fatalf("no %s in the model: %v", k, plain)
		}
	}
	for k, v := range plain {
		if strings.Contains(v, "«") {
			t.Errorf("%s reads %q with VibeMode off", k, v)
		}
	}
	vibe.Set(true)
	on := words()
	for k, v := range on {
		if !strings.Contains(v, "«") {
			t.Errorf("%s reads %q with VibeMode on, was %q", k, v, plain[k])
		}
	}
	vibe.Set(false)
	for k, v := range words() {
		if v != plain[k] {
			t.Errorf("%s reads %q with VibeMode off again, was %q", k, v, plain[k])
		}
	}
	// The name is a draw of the seed that does not move: only the words it is
	// made of do.
	if got := strings.Replace(strings.Replace(on["town"], "«", "", 1), "»", "", 1); got != plain["town"] {
		t.Errorf("the town is %q with VibeMode on and %q off: the draw moved", on["town"], plain["town"])
	}
}
