package styletest

// What changes in a glyph seen in a mirror. A glyph that is in neither
// table is its own mirror image.
var (
	mirrorLR = map[rune]rune{'╱': '╲', '╲': '╱', '◄': '►', '►': '◄', '◤': '◥', '◥': '◤', '◣': '◢', '◢': '◣'}
	mirrorTB = map[rune]rune{'╱': '╲', '╲': '╱', '▲': '▼', '▼': '▲', '◤': '◣', '◣': '◤', '◥': '◢', '◢': '◥', '▀': '▄', '▄': '▀'}
)

// MirrorRune is r seen in a mirror: across the vertical axis (lr) or the
// horizontal one. The mandala's petal points, the core's rays and the
// half-block pixels turn with the mirror; every other glyph it draws is its
// own mirror image.
func MirrorRune(r rune, lr bool) rune {
	table := mirrorTB
	if lr {
		table = mirrorLR
	}
	if m, ok := table[r]; ok {
		return m
	}
	return r
}
