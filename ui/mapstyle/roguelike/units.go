package roguelike

// Names for the big numbers in this package, so a cost reads in words and
// not in e-notation. config has the same set as exported names
// (config.Quadrillion); this package cannot import config, so it keeps its own.
const (
	million  = 1_000_000.0
	billion  = 1000 * million
	trillion = 1000 * billion
)
