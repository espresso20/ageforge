package ui

import (
	"image"
	"sync"

	"github.com/espresso20/ageforge/window/kit"
)

// prim.go is what the screens' code asks of the things it puts on a page and
// of the running app. The kit lays out, paints and routes the mouse; this is
// the rest: a rectangle on the screen, a focus flag, who has the keyboard,
// and the loop every change to the UI goes through.

// A Primitive is a kit widget the screens can place and give the keyboard.
type Primitive interface {
	kit.Widget
	GetRect() (x, y, w, h int)
	SetRect(x, y, w, h int)
	// Focus gives the primitive the keyboard; a container hands it on to
	// one of its own through delegate.
	Focus(delegate func(p Primitive))
	HasFocus() bool
	Blur()
}

// primBase is the part every primitive has: where it is on the screen, and
// whether it has been given the keyboard.
type primBase struct {
	x, y, w, h int
	focused    bool
}

// newPrimBase is a base with a size before anything places it (15 by 10):
// code that reads a size too early has always got this one.
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

// focusTarget is the widget in tree that holds the keyboard: the primitive
// whose own flag is set, or tree itself when none is.
func focusTarget(tree kit.Widget) kit.Widget {
	at := tree
	walkWidgets(tree, func(w kit.Widget) {
		if f, ok := w.(interface{ boxFocused() bool }); ok && f.boxFocused() {
			at = w
		}
	})
	return at
}

// keyTo gives p a key as if p were the whole screen: past the captures on
// the way down to whatever in p has the keyboard, then to that and outward.
// It is how one screen hands a key on to another (the dashboard to its
// command bar), and how a test presses a key on a primitive that is on no
// page. Nothing is laid out: p is taken as it stands.
func keyTo(p Primitive, k kit.Key) bool {
	path := widgetPath(p, focusTarget(p))
	for _, w := range path {
		if f, ok := w.(kit.KeyFilter); ok {
			var on bool
			if k, on = f.FilterKey(k); !on {
				return true
			}
		}
	}
	for i := len(path) - 1; i >= 0; i-- {
		if h, ok := path[i].(kit.KeyHandler); ok && h.Key(k) {
			return true
		}
	}
	return false
}

// widgetPath is the widgets from root down to target, both included; nil
// when target is not in root.
func widgetPath(root, target kit.Widget) []kit.Widget {
	if root == target {
		return []kit.Widget{root}
	}
	if p, ok := root.(kit.Parent); ok {
		for _, c := range p.Children() {
			if rest := widgetPath(c, target); rest != nil {
				return append([]kit.Widget{root}, rest...)
			}
		}
	}
	return nil
}

// paintAlone lays p out alone on a screen of w by h cells and paints it: for
// a picture of one primitive (the epoch card), and for tests.
func paintAlone(p Primitive, w, h int) *kit.Buffer {
	p.SetRect(0, 0, w, h)
	root := kit.NewRoot(p)
	root.Resize(w, h)
	b := kit.NewBuffer(w, h)
	root.Paint(b)
	return b
}

// An appLoop is the running app: who has the keyboard, the one goroutine
// every change to the UI happens on, and the frame the window shows. Keys,
// the mouse and queued updates are events; each is handled in turn, and
// after it the screen is painted again.
type appLoop struct {
	mu   sync.Mutex // guards root, focus and the hooks
	root Primitive
	tree *kit.Root // the root's kit tree
	// focus is the primitive that was last given the keyboard.
	focus   Primitive
	capture func(*kit.Key) *kit.Key
	before  func(w, h int)
	goBack  func() bool // whether a right click may be taken for Esc now

	events chan func()   // what the loop has to do, in order
	done   chan struct{} // closed by Stop
	once   sync.Once

	cols, rows int        // the screen, in cells (the loop's own)
	back       kit.Buffer // the frame being painted (the loop's own)
	frameMu    sync.Mutex
	front      kit.Buffer // the last frame painted, for the window
}

func newAppLoop() *appLoop {
	return &appLoop{events: make(chan func(), 1024), done: make(chan struct{})}
}

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
func (a *appLoop) SetRoot(root Primitive, _ bool) *appLoop {
	a.mu.Lock()
	a.root, a.tree = root, kit.NewRoot(root)
	a.mu.Unlock()
	return a.SetFocus(root)
}

// SetInputCapture sets what sees every key before the root does. It returns
// the key to go on with, or nil to stop it there.
func (a *appLoop) SetInputCapture(fn func(*kit.Key) *kit.Key) *appLoop {
	a.mu.Lock()
	a.capture = fn
	a.mu.Unlock()
	return a
}

// SetBack sets what says whether a right click means Back now. Where it
// does, the click presses Esc, so it does whatever Esc does there; with
// nothing set a right click does nothing.
func (a *appLoop) SetBack(fn func() bool) *appLoop {
	a.mu.Lock()
	a.goBack = fn
	a.mu.Unlock()
	return a
}

// SetBeforeDrawFunc sets what is told the screen's size before each paint.
func (a *appLoop) SetBeforeDrawFunc(fn func(w, h int)) *appLoop {
	a.mu.Lock()
	a.before = fn
	a.mu.Unlock()
	return a
}

// Run handles events until Stop. It paints once first, so there is a frame
// before anything happens.
func (a *appLoop) Run() error {
	a.draw()
	for {
		select {
		case <-a.done:
			return nil
		case f := <-a.events:
			f()
		}
	}
}

// Stop ends Run. It is safe to call more than once, and from any goroutine.
func (a *appLoop) Stop() { a.once.Do(func() { close(a.done) }) }

// post hands the loop something to do. It reports false once the loop has
// stopped.
func (a *appLoop) post(f func()) bool {
	select {
	case <-a.done:
		return false
	case a.events <- f:
		return true
	}
}

// QueueUpdate runs f on the loop's goroutine and waits for it. It must not
// be called from that goroutine: f would wait for itself.
func (a *appLoop) QueueUpdate(f func()) *appLoop {
	ran := make(chan struct{})
	if a.post(func() { f(); close(ran) }) {
		select {
		case <-ran:
		case <-a.done:
		}
	}
	return a
}

// QueueUpdateDraw is QueueUpdate, with the screen painted afterward.
func (a *appLoop) QueueUpdateDraw(f func()) *appLoop {
	return a.QueueUpdate(func() { f(); a.draw() })
}

// ForceDraw paints the screen now. It is for the loop's own goroutine.
func (a *appLoop) ForceDraw() *appLoop {
	a.draw()
	return a
}

// Resize, Key and Mouse are the window's events, queued for the loop. A key
// or a mouse event never waits: if the loop is so far behind that its queue
// is full, the event is dropped. A resize is rare and must not be lost, so
// it waits for room.
func (a *appLoop) Resize(cols, rows int) {
	a.post(func() {
		a.cols, a.rows = cols, rows
		a.draw()
	})
}

func (a *appLoop) Key(k kit.Key) {
	a.offer(func() {
		a.key(k)
		a.draw()
	})
}

func (a *appLoop) Mouse(m kit.Mouse) {
	a.offer(func() {
		tree := a.rootTree()
		if tree == nil {
			return
		}
		changed := tree.Mouse(m)
		// A right click that nothing on the screen took is Back: Esc, where
		// the app says Esc goes back.
		if m.Kind == kit.MousePress && m.Button == kit.ButtonRight && !tree.Took() {
			a.mu.Lock()
			goBack := a.goBack
			a.mu.Unlock()
			if goBack != nil && goBack() {
				a.key(kit.Key{Code: kit.KeyEsc})
				changed = true
			}
		}
		if changed {
			a.draw()
		}
	})
}

// offer hands the loop something to do if it has room for it now.
func (a *appLoop) offer(f func()) {
	select {
	case a.events <- f:
	default:
	}
}

func (a *appLoop) rootTree() *kit.Root {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.tree
}

// key gives the screen a key: the app's capture first, then the root's tree,
// which hands it to whatever has the keyboard.
func (a *appLoop) key(k kit.Key) {
	a.mu.Lock()
	capture, root, tree := a.capture, a.root, a.tree
	a.mu.Unlock()
	if capture != nil {
		out := capture(&k)
		if out == nil {
			return
		}
		k = *out
	}
	if k.Ctrl('c') {
		a.Stop() // Ctrl+C leaves the game, as it always has
		return
	}
	if root == nil || !root.HasFocus() {
		return
	}
	tree.Focus(focusTarget(root))
	tree.Key(k)
}

// draw paints the root for the screen's size and makes that the frame the
// window shows.
func (a *appLoop) draw() {
	a.mu.Lock()
	before, root, tree := a.before, a.root, a.tree
	a.mu.Unlock()
	if root == nil || a.cols <= 0 || a.rows <= 0 {
		return
	}
	if before != nil {
		before(a.cols, a.rows)
	}
	root.SetRect(0, 0, a.cols, a.rows)
	tree.Resize(a.cols, a.rows)
	tree.Paint(&a.back)
	a.frameMu.Lock()
	a.front, a.back = a.back, a.front
	a.frameMu.Unlock()
}

// Frame copies the last frame painted into dst, with its text cursor and
// the shape of the mouse pointer. The window
// calls it from its own thread.
func (a *appLoop) Frame(dst *kit.Buffer) {
	a.frameMu.Lock()
	defer a.frameMu.Unlock()
	dst.W, dst.H = a.front.W, a.front.H
	dst.Cells = append(dst.Cells[:0], a.front.Cells...)
	dst.Cursor, dst.CursorOn = a.front.Cursor, a.front.CursorOn
	dst.Pointer = a.front.Pointer
	dst.Pics = append(dst.Pics[:0], a.front.Pics...)
}

// Size is the screen in cells, as the last frame was painted.
func (a *appLoop) Size() image.Point {
	a.frameMu.Lock()
	defer a.frameMu.Unlock()
	return image.Pt(a.front.W, a.front.H)
}
