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

// TestTheRostersReadTheScript: the words of the city's parts, the movers, the
// Earth arc's looks, the station parts, the aliens and the lineages are read
// when they are asked for, so one table follows the switch both ways.
func TestTheRostersReadTheScript(t *testing.T) {
	told(t)
	words := func() map[string]string {
		out := map[string]string{}
		for f := mapmodel.CityFeature(1); f < mapmodel.NumCityFeatures; f++ {
			i := f.Info()
			out["city "+i.Key] = i.Name + "|" + i.Title + "|" + strings.Join(i.Lines, "|") + "|" + i.Line(3)
		}
		for k := mapmodel.Mover(1); k < mapmodel.Mover(40); k++ {
			i := k.Info()
			if i.Key == "" {
				continue
			}
			out["mover "+i.Key] = i.Name + "|" + i.Title + "|" + strings.Join(i.Lines, "|") + "|" + i.Line(3)
		}
		for _, l := range mapmodel.CityLooks() {
			out["look "+l.Key] = l.Name + "|" + l.Structure + "|" + l.Ambient
		}
		for _, lin := range mapmodel.LineageOrder {
			out["lineage "+lin] = mapmodel.LineageName(lin)
			for s := mapmodel.SkyScene(0); s < mapmodel.NumSkyScenes; s++ {
				out["part "+lin+" "+string(rune('0'+int(s)))] = mapmodel.SkyPartOf(s, lin).Name
			}
		}
		for i, a := range mapmodel.Aliens {
			out["alien "+string(rune('0'+i))] = a.Color.String() + "|" + a.Lines[0].String() + "|" + a.Lines[1].String()
		}
		out["war"] = mapmodel.RelWar.String()
		return out
	}
	plain := words()
	if len(plain) < 100 {
		t.Fatalf("only %d texts of the rosters read", len(plain))
	}
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

// TestTheBuilderFollowsTheSwitch: the one builder, which caches the last model
// and the world of a seed, builds again when VibeMode is switched, so the words
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
			"town":     m.Town.World.Name,
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
