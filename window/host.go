package window

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

// Host is the screen tview draws into. It is a simulation screen, wrapped so
// theme colors resolve (theme.WrapScreen) and so the window can read the cells
// while tview shows a frame: tcell hands out its cell slice without a lock, and
// Show rewrites it, so every call that rewrites the cells takes mu and the
// window reads under it.
type Host struct {
	tcell.SimulationScreen
	mu sync.Mutex
}

// NewHost is an initialised, empty host. Its size is set by the window.
func NewHost() (*Host, error) {
	h := &Host{SimulationScreen: tcell.NewSimulationScreen("UTF-8")}
	if err := h.SimulationScreen.Init(); err != nil {
		return nil, err
	}
	return h, nil
}

// Init does nothing: NewHost has done it, and tview calls it again when it starts.
func (h *Host) Init() error { return nil }

func (h *Host) Show() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.SimulationScreen.Show()
}

func (h *Host) Sync() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.SimulationScreen.Sync()
}

func (h *Host) Fini() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.SimulationScreen.Fini()
}

// SetGrid sets the grid to cols by rows and tells tview, which lays out again.
func (h *Host) SetGrid(cols, rows int) {
	h.mu.Lock()
	h.SimulationScreen.SetSize(cols, rows)
	h.mu.Unlock()
	_ = h.SimulationScreen.PostEvent(tcell.NewEventResize(cols, rows))
}

// frame is a copy of the cells the screen shows now.
type frame struct {
	cells         []tcell.SimCell
	cols, rows    int
	curX, curY    int
	cursorShowing bool
}

// snapshot copies the cells into f (reusing its storage).
func (h *Host) snapshot(f *frame) {
	h.mu.Lock()
	defer h.mu.Unlock()
	cells, w, rows := h.SimulationScreen.GetContents()
	f.cells = append(f.cells[:0], cells...)
	f.cols, f.rows = w, rows
	f.curX, f.curY, f.cursorShowing = h.SimulationScreen.GetCursor()
}
