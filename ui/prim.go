package ui

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/window/kit"
)

// prim.go is what the screens' code asks of the things it puts on a page and
// of the running app, in the game's own types. The kit lays out, paints and
// routes the mouse; this is the rest: a rectangle on the screen, a focus
// flag, who has the keyboard, and a turn on the UI goroutine. tview's
// Application still runs the loop underneath, and sees one primitive, the
// root (tviewRoot).

// A Primitive is a kit widget the screens can place and give the keyboard.
// Draw and InputHandler are how a primitive is driven on its own, by the root
// and by tests: both go with the terminal screen.
type Primitive interface {
	kit.Widget
	GetRect() (x, y, w, h int)
	SetRect(x, y, w, h int)
	// Focus gives the primitive the keyboard; a container hands it on to
	// one of its own through delegate.
	Focus(delegate func(p Primitive))
	HasFocus() bool
	Blur()
	Draw(scr tcell.Screen)
	InputHandler() func(*tcell.EventKey, func(Primitive))
}

// primBase is the part every primitive has: where it is on the screen, and
// whether it has been given the keyboard.
type primBase struct {
	x, y, w, h int
	focused    bool
}

// newPrimBase is a base with the size a box has before anything places it
// (15 by 10, tview's): code that reads a size too early has always got this.
func newPrimBase() *primBase { return &primBase{w: 15, h: 10} }

func (b *primBase) GetRect() (x, y, w, h int) { return b.x, b.y, b.w, b.h }

// GetInnerRect is the same rectangle: a primitive with a border of its own
// answers for it itself.
func (b *primBase) GetInnerRect() (x, y, w, h int) { return b.x, b.y, b.w, b.h }

func (b *primBase) SetRect(x, y, w, h int) { b.x, b.y, b.w, b.h = x, y, w, h }
func (b *primBase) Focus(func(Primitive))  { b.focused = true }
func (b *primBase) Blur()                  { b.focused = false }
func (b *primBase) HasFocus() bool         { return b.focused }

// boxFocused reports whether this very primitive was given the keyboard
// (HasFocus also answers for what is inside a container).
func (b *primBase) boxFocused() bool { return b.focused }

// An appLoop is the running app as the screens see it: who has the keyboard,
// and a turn on the UI goroutine. The focus is kept here; the loop, the
// screen and the queue are tview's until the terminal goes.
type appLoop struct {
	tv   *tview.Application
	mu   sync.Mutex
	root Primitive
	// focus is the primitive that was last given the keyboard.
	focus Primitive
}

func newAppLoop() *appLoop { return &appLoop{tv: tview.NewApplication()} }

// SetFocus gives p the keyboard: what had it is told it lost it, and p may
// hand it on to something inside itself.
func (a *appLoop) SetFocus(p Primitive) *appLoop {
	a.mu.Lock()
	if a.focus != nil {
		a.focus.Blur()
	}
	a.focus = p
	a.mu.Unlock()
	if p != nil {
		p.Focus(func(q Primitive) { a.SetFocus(q) })
	}
	return a
}

// GetFocus is the primitive that has the keyboard.
func (a *appLoop) GetFocus() Primitive {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.focus
}

// SetRoot makes root the whole screen, and gives it the keyboard.
func (a *appLoop) SetRoot(root Primitive, fullscreen bool) *appLoop {
	a.mu.Lock()
	a.root = root
	a.mu.Unlock()
	a.tv.SetRoot(&tviewRoot{Box: tview.NewBox(), app: a}, fullscreen)
	return a.SetFocus(root)
}

func (a *appLoop) rootPrim() Primitive {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.root
}

func (a *appLoop) SetScreen(s tcell.Screen) *appLoop {
	a.tv.SetScreen(s)
	return a
}

func (a *appLoop) Run() error { return a.tv.Run() }
func (a *appLoop) Stop()      { a.tv.Stop() }

// QueueUpdateDraw runs f on the UI goroutine and draws afterward;
// QueueUpdate runs it and does not draw.
func (a *appLoop) QueueUpdateDraw(f func()) *appLoop {
	a.tv.QueueUpdateDraw(f)
	return a
}

func (a *appLoop) QueueUpdate(f func()) *appLoop {
	a.tv.QueueUpdate(f)
	return a
}

func (a *appLoop) ForceDraw() *appLoop {
	a.tv.ForceDraw()
	return a
}

// QueueEvent hands the loop an event as if the screen had sent it.
func (a *appLoop) QueueEvent(ev tcell.Event) *appLoop {
	a.tv.QueueEvent(ev)
	return a
}

// SetInputCapture sets what sees every key before the root does.
func (a *appLoop) SetInputCapture(fn func(*tcell.EventKey) *tcell.EventKey) *appLoop {
	a.tv.SetInputCapture(fn)
	return a
}

// SetBeforeDrawFunc sets what is called with the screen before each draw.
func (a *appLoop) SetBeforeDrawFunc(fn func(scr tcell.Screen) bool) *appLoop {
	a.tv.SetBeforeDrawFunc(fn)
	return a
}

// A tviewRoot is the one primitive tview's Application sees: it stands for
// the app's root, drawing it in the screen's rectangle and handing it the
// keys while something in it has the keyboard.
type tviewRoot struct {
	*tview.Box
	app *appLoop
}

func (r *tviewRoot) Draw(scr tcell.Screen) {
	root := r.app.rootPrim()
	if root == nil {
		return
	}
	// The cursor shows only if this frame's draw puts it somewhere.
	scr.HideCursor()
	root.SetRect(r.GetRect())
	root.Draw(scr)
}

func (r *tviewRoot) HasFocus() bool { return true }

func (r *tviewRoot) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return func(ev *tcell.EventKey, _ func(tview.Primitive)) {
		root := r.app.rootPrim()
		if root == nil || !root.HasFocus() {
			return
		}
		if h := root.InputHandler(); h != nil {
			h(ev, func(p Primitive) { r.app.SetFocus(p) })
		}
	}
}

// keysIn is a key handler as a Primitive hands one out, for a handler that
// reads the kit's keys: tcell's keys stop here.
func keysIn(fn func(k *kit.Key)) func(*tcell.EventKey, func(Primitive)) {
	return func(ev *tcell.EventKey, _ func(Primitive)) {
		k := kitKey(ev)
		fn(&k)
	}
}
