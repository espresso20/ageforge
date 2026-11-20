// Package fonts holds the pixel font the game draws its large headings in.
// The window embeds the whole folder for its own faces; this is the one file
// the drawn panels need as bytes.
package fonts

import _ "embed"

// Departure is Departure Mono, an OpenType font.
//
//go:embed DepartureMono-Regular.otf
var Departure []byte
