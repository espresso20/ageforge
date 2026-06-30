// Package assets embeds the game's remaining static assets into the binary so
// they are available without a separate assets/ directory on disk.
//
// History: this package used to embed ~51MB of realistic map-background photos
// (assets/maps) and ~2.9MB of dead per-building sprite PNGs (assets/sprites/
// buildings) for the retired MapV4 renderer. Both were deleted in the map
// rewrite (the new ui/citymap renderer is fully procedural and generates any
// sprites it needs in-code via pkg/sprites). What remains here are the small
// loose domain/resource preview sprites under sprites/.
package assets

import "embed"

//go:embed sprites/*.png
var FS embed.FS
