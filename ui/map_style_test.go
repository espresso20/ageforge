package ui

import (
	"fmt"
	"strings"
	"testing"
)

// TestMapStyleSwitch: AGEFORGE_MAP=engine picks the procedural engine for the
// Map panel and the mini map, and anything else keeps today's map.
func TestMapStyleSwitch(t *testing.T) {
	kind := func(env string, set bool) string {
		if set {
			t.Setenv(mapStyleEnv, env)
		} else {
			t.Setenv(mapStyleEnv, "")
		}
		return fmt.Sprintf("%T", newMapStyle())
	}
	if got := kind("engine", true); !strings.Contains(got, "procmap") {
		t.Errorf("AGEFORGE_MAP=engine draws with %s", got)
	}
	for _, env := range []string{"", "roguelike", "Engine", "yes"} {
		if got := kind(env, true); !strings.Contains(got, "roguelike") {
			t.Errorf("AGEFORGE_MAP=%q draws with %s, want today's map", env, got)
		}
	}
	t.Setenv(mapStyleEnv, "engine")
	if p := newMapPanel(newMapViews()); !strings.Contains(fmt.Sprintf("%T", p.view), "procmap") {
		t.Errorf("the Map panel draws with %T", p.view)
	}
	if m := newMiniMap(newMapViews()); !strings.Contains(fmt.Sprintf("%T", m.view), "procmap") {
		t.Errorf("the mini map draws with %T", m.view)
	}
}
