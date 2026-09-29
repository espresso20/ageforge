// Package all assembles the standard map style registry. It is the one
// place that knows every style; the dashboard's "map style" setting reads
// its names, and a new style (topology, organism) joins with one line here.
package all

import (
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/roguelike"
	"github.com/espresso20/ageforge/ui/mapstyle/skyline"
)

// Registry returns a fresh registry of the standard styles. The first,
// roguelike, is the default.
func Registry() *mapstyle.Registry {
	return mapstyle.NewRegistry(roguelike.Entry(), skyline.Entry())
}
