package ui

import (
	"os"

	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/procmap"
	"github.com/espresso20/ageforge/ui/mapstyle/roguelike"
)

// mapStyleEnv is the environment variable that picks the map's renderer:
// AGEFORGE_MAP=engine draws the Map panel and the mini map with the
// procedural engine; anything else keeps the map as it has been. This is the
// one place that reads it.
const mapStyleEnv = "AGEFORGE_MAP"

// newMapStyle is a map view of the renderer the environment picks.
func newMapStyle() mapstyle.Style { return mapStyleFor(os.Getenv(mapStyleEnv)) }

// mapStyleFor is a map view of the renderer a setting names.
func mapStyleFor(choice string) mapstyle.Style {
	if choice == "engine" {
		return procmap.New()
	}
	return roguelike.New()
}
