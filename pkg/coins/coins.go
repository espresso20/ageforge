// Package coins draws the pixel-art pieces of the coin cabinet. It is pure
// drawing: standard library only, no files, no registration at start-up.
package coins

import (
	"container/list"
	"image"
	"sync"
)

// Levels is how many levels a coin has; Legendary is the one above them.
const Levels = 6
const Legendary = Levels + 1

// entry is what an age contributes: its builder for levels 1 to Levels (lod 1
// is the small version, seed 0 the plain reference copy) and its legendary.
type entry struct {
	key    string
	build  func(level, lod int, seed uint32) *spec
	legend func() *spec
}

// table is the fixed list of ages with drawn money, in display order.
var table = [...]entry{
	{"classical_age", classical, classicalLegend},
	{"medieval_age", medieval, medievalLegend},
	{"victorian_age", victorian, victorianLegend},
	{"modern_age", modern, modernLegend},
	{"cyberpunk_age", cyberpunk, cyberpunkLegend},
	{"fusion_age", fusion, fusionLegend},
}

func find(age string) *entry {
	for i := range table {
		if table[i].key == age {
			return &table[i]
		}
	}
	return nil
}

// Ages returns the keys of the ages that have drawn money, in a fixed order.
func Ages() []string {
	keys := make([]string, len(table))
	for i := range table {
		keys[i] = table[i].key
	}
	return keys
}

// Has reports whether age has drawn money.
func Has(age string) bool { return find(age) != nil }

// Render draws one piece, size x size pixels. level is 1..Levels, or
// Legendary. Seed 0 is the plain reference copy; another seed strikes the same
// piece again with its own variation (a legendary has no variation, so its
// seed is ignored). typ does not yet change the picture. Render returns nil
// for an age without drawn money or a level out of range.
func Render(age, typ string, level int, seed uint32, size int) *image.RGBA {
	e := find(age)
	if e == nil || level < 1 || level > Legendary || size < 1 {
		return nil
	}
	if level == Legendary {
		return render(e.legend(), size)
	}
	lod := 0
	if size <= 64 {
		lod = 1
	}
	return render(e.build(level, lod, seed), size)
}

// Mint is a bounded cache of rendered pieces, safe for concurrent use.
type Mint struct {
	mu    sync.Mutex
	limit int
	order *list.List // front is the most recently used
	byKey map[mintKey]*list.Element
}

type mintKey struct {
	age, typ    string
	level, size int
	seed        uint32
}

type mintItem struct {
	key mintKey
	img *image.RGBA
}

// NewMint returns a cache holding at most limit pieces (at least one).
func NewMint(limit int) *Mint {
	return &Mint{limit: max(limit, 1), order: list.New(), byKey: map[mintKey]*list.Element{}}
}

// Piece returns the rendered piece, drawing it on first use and dropping the
// least recently used piece past the limit. The image is shared: callers must
// not write to it. It is nil where Render is nil.
func (m *Mint) Piece(age, typ string, level int, seed uint32, size int) *image.RGBA {
	k := mintKey{age, typ, level, size, seed}
	m.mu.Lock()
	if el, ok := m.byKey[k]; ok {
		m.order.MoveToFront(el)
		img := el.Value.(*mintItem).img
		m.mu.Unlock()
		return img
	}
	m.mu.Unlock()
	img := Render(age, typ, level, seed, size)
	if img == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if el, ok := m.byKey[k]; ok { // another caller drew it meanwhile
		m.order.MoveToFront(el)
		return el.Value.(*mintItem).img
	}
	m.byKey[k] = m.order.PushFront(&mintItem{k, img})
	for m.order.Len() > m.limit {
		last := m.order.Back()
		delete(m.byKey, last.Value.(*mintItem).key)
		m.order.Remove(last)
	}
	return img
}
