// Package mapview draws the sky the world package describes: every space
// scale of the Map panel, from the Local Group down to the home planet, on
// one pixel canvas with one continuous zoom. It is pure Go over an
// *image.RGBA (the window uploads and scales the result), so it is
// deterministic and testable. It imports world and nothing else of ours: no
// ui, no theme, no tview.
//
// The camera is a log zoom plus a centre. Each frame of the world's tree is
// drawn while the view width is inside its band; where two bands meet the
// parent and child are dithered into each other with a Bayer mask, not
// blended. Sprites here are plain placeholders; the art arrives frame by
// frame on top of the helpers in bayer.go.
package mapview
