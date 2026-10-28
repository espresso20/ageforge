package mapstyle

import (
	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/vibe"
)

// visitor.go says what the cursor says on the rare visitor: a small visitor
// far from home, or an alien's craft. Either reports KindAlien, which the game
// hears to hand out the secret badge.

var (
	visitorTitles = map[string]vibe.Entry{
		"visitor": vibe.E("ui.mapstyle.roguelike.visitor.visitorAt.02", "Visitor"),
		"alien":   vibe.E("ui.mapstyle.roguelike.visitor.visitorAt.01", "Unknown craft"),
	}
	visitorLines = map[string][]vibe.Entry{
		"visitor": {
			vibe.E("ui.mapstyle.roguelike.visitor.walkerLines.01", "It waves, then it's gone."),
			vibe.E("ui.mapstyle.roguelike.visitor.walkerLines.02", "A small visitor, far from home."),
			vibe.E("ui.mapstyle.roguelike.visitor.walkerLines.03", "It seems lost, and fine with that."),
		},
		"alien": {
			vibe.E("ui.mapstyle.roguelike.visitor.saucerLines.01", "Not one of ours."),
			vibe.E("ui.mapstyle.roguelike.visitor.saucerLines.02", "It hums, then it's gone."),
			vibe.E("ui.mapstyle.roguelike.visitor.saucerLines.03", "Lights in the sky. Nobody will believe you."),
		},
	}
)

// describeVisitor is the visitor, or the alien's craft, under the cursor.
func describeVisitor(th mapgen.Thing) Inspection {
	lines := visitorLines[th.Key]
	return Inspection{
		Title: visitorTitles[th.Key].String(),
		Lines: []string{lines[th.Ref%len(lines)].String()},
		Kind:  KindAlien,
	}
}
