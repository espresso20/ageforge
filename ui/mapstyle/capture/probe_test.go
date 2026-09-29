//go:build mapcapture

package capture

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapmodel"
)

func TestProbeModels(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "map_captures", "states")
	b := mapmodel.NewBuilder(nil)
	for _, age := range config.AgeOrder() {
		st, err := LoadState(filepath.Join(dir, age+".json.gz"))
		if err != nil {
			t.Log(age, err)
			continue
		}
		prev, _ := LoadState(filepath.Join(dir, "prev_"+age+".json.gz"))
		pm := b.Build(&prev, nil)
		t0 := time.Now()
		m := b.Build(&st, mapmodel.VisitOf(pm))
		legacy := 0
		for _, tt := range m.Town.Tiles {
			if tt.Legacy {
				legacy++
			}
		}
		t.Logf("%-17s %v bld=%d types=%d tiles=%d legacy=%d lots=%d routes=%d civs=%d/%d staffed=%d idle=%d traffic=%.2f wealth=%.2f new=%d full=%d worst=%q",
			age, time.Since(t0).Round(time.Microsecond), m.TotalBuildings(), len(m.Buildings), len(m.Town.Tiles), legacy, len(m.Skyline.Lots),
			len(m.Routes), m.DiscoveredCount(), len(m.Factions), m.Workers.Staffed, m.Workers.Idle, m.Activity.Traffic, m.Activity.Wealth,
			m.Recap.NewCount, len(m.Flows.Full), m.Flows.Worst)
		t.Logf("   news: %s", m.Recap.Headline(160))
	}
}
