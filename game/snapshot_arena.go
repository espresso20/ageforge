package game

import "slices"

// sliceArena hands out copies of many small slices from one backing array: a
// snapshot owns every slice in it (a caller may edit or append to one) at the
// cost of one allocation instead of one per slice. Count the elements with
// reserve first, then copy with clone. Each copy is capped at its own length,
// so an append by the caller moves to new memory instead of overwriting the
// copy beside it.
type sliceArena[T any] struct {
	want int
	buf  []T
}

// reserve counts src among the slices clone will copy.
func (a *sliceArena[T]) reserve(src []T) { a.want += len(src) }

// clone is slices.Clone(src), nil and empty included, out of the arena.
func (a *sliceArena[T]) clone(src []T) []T {
	if src == nil {
		return nil
	}
	if len(src) == 0 {
		return []T{}
	}
	if a.buf == nil {
		a.buf = make([]T, 0, a.want)
	}
	if cap(a.buf)-len(a.buf) < len(src) { // the count came up short
		return slices.Clone(src)
	}
	start := len(a.buf)
	a.buf = append(a.buf, src...)
	return a.buf[start:len(a.buf):len(a.buf)]
}

// take returns an empty slice with room for n elements, out of the arena, for
// the caller to append into. A nil arena makes a plain slice. Asking for more
// than reserve counted falls back to one of its own.
func (a *sliceArena[T]) take(n int) []T {
	if n <= 0 {
		return nil
	}
	if a == nil || a.want == 0 {
		return make([]T, 0, n)
	}
	if a.buf == nil {
		a.buf = make([]T, 0, a.want)
	}
	if cap(a.buf)-len(a.buf) < n {
		return make([]T, 0, n)
	}
	start := len(a.buf)
	a.buf = a.buf[:start+n]
	return a.buf[start : start : start+n]
}
