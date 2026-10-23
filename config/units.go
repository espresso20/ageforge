package config

// Names for big numbers, so a cost reads as 75.4 * Million and not as an
// e-notation literal. They are untyped constants, so 75.4 * Million is worked
// out exactly at compile time and is the same float64 as the old literal.
//
// The names go up in steps of a thousand and match the suffixes the game
// prints (pkg/textfmt): K, M, B, T, Q, Qi, Sx, Sp, Oc, No, Dc. Write a number
// from a million up as <m> * <Unit> with m between 1 and 1000 (2.5 *
// Quadrillion). Below a million, write plain digits with an underscore every
// three digits from five digits up (22_700, 950_000, 4_500.5).
const (
	Thousand    = 1000.0             // K
	Million     = 1000 * Thousand    // M
	Billion     = 1000 * Million     // B
	Trillion    = 1000 * Billion     // T
	Quadrillion = 1000 * Trillion    // Q
	Quintillion = 1000 * Quadrillion // Qi
	Sextillion  = 1000 * Quintillion // Sx
	Septillion  = 1000 * Sextillion  // Sp
	Octillion   = 1000 * Septillion  // Oc
	Nonillion   = 1000 * Octillion   // No
	Decillion   = 1000 * Nonillion   // Dc
)
