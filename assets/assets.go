// Package assets embeds the game's static assets (map backgrounds and
// building sprites) into the binary so they are available without a
// separate assets/ directory on disk.
package assets

import "embed"

//go:embed maps sprites/buildings
var FS embed.FS
