package window

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/espresso20/ageforge/window/kit"
)

// pointer.go turns the mouse into events at cells of the grid, for the
// screens the game draws with its own kit. The map of space reads the mouse
// itself (skymouse.go), in pixels; the two do not meet, because a kit screen
// takes only the events over its own cells.

// cellPointer is what the window remembers of the mouse between frames.
type cellPointer struct {
	x, y  int     // the cell the pointer was last over
	in    bool    // it was over the grid
	wheel float64 // wheel turn not yet sent, in rows
}

// buttons are the mouse buttons the kit knows.
var buttons = [...]struct {
	e ebiten.MouseButton
	k kit.MouseButton
}{{ebiten.MouseButtonLeft, kit.ButtonLeft}, {ebiten.MouseButtonRight, kit.ButtonRight}, {ebiten.MouseButtonMiddle, kit.ButtonMiddle}}

// pointerShapes are the system's pointers for the shapes the kit names.
var pointerShapes = [...]ebiten.CursorShapeType{
	kit.PointerArrow: ebiten.CursorShapeDefault,
	kit.PointerHand:  ebiten.CursorShapePointer,
	kit.PointerText:  ebiten.CursorShapeText,
}

// cellAt is the cell under device pixel px, py and whether it is in the grid.
func (g *Game) cellAt(px, py int) (x, y int, in bool) {
	if g.cw == 0 || g.ch == 0 {
		return 0, 0, false
	}
	ox, oy := (g.pw-g.cols*g.cw)/2, (g.ph-g.rows*g.ch)/2
	x = int(math.Floor(float64(px-ox) / float64(g.cw)))
	y = int(math.Floor(float64(py-oy) / float64(g.ch)))
	return x, y, x >= 0 && y >= 0 && x < g.cols && y < g.rows
}

// pointer reads the mouse for this frame and sends what happened.
func (g *Game) pointer() {
	send := g.o.Screen.Mouse
	p := &g.ptr
	x, y, in := g.cellAt(ebiten.CursorPosition())
	if !in || !ebiten.IsFocused() {
		if p.in {
			p.in = false
			send(kit.Mouse{Kind: kit.MouseLeave, X: -1, Y: -1})
		}
		return
	}
	m := mods{
		ctrl:  ebiten.IsKeyPressed(ebiten.KeyControl),
		alt:   ebiten.IsKeyPressed(ebiten.KeyAlt),
		shift: ebiten.IsKeyPressed(ebiten.KeyShift),
	}
	ev := kit.Mouse{X: x, Y: y, Mod: m.kit()}
	if !p.in || x != p.x || y != p.y {
		p.x, p.y, p.in = x, y, true
		ev.Kind = kit.MouseMove
		send(ev)
	}
	for _, b := range buttons {
		ev.Button = b.k
		if inpututil.IsMouseButtonJustPressed(b.e) {
			ev.Kind = kit.MousePress
			send(ev)
		}
		if inpututil.IsMouseButtonJustReleased(b.e) {
			ev.Kind = kit.MouseRelease
			send(ev)
		}
	}
	// The wheel: whole rows are sent, and the rest is kept for the next
	// frame, so a trackpad's small turns add up. Up is negative.
	if _, wy := ebiten.Wheel(); wy != 0 {
		p.wheel -= math.Max(-3, math.Min(3, wy))
		if rows := int(p.wheel); rows != 0 {
			p.wheel -= float64(rows)
			send(kit.Mouse{Kind: kit.MouseWheel, X: x, Y: y, Wheel: rows, Mod: ev.Mod})
		}
	}
}
