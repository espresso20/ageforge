package ui

import "strings"

// research_frames.go holds the frames of the Research tree's tech badges:
// a box of three rows, an octagon of five and the large ring of a tech's
// card, built from a set of corner and edge runes (round, double, dashed).

// frameRunes is a badge frame: the four corners, then the flat and the
// upright edge.
type frameRunes struct{ tl, tr, bl, br, h, v rune }

var (
	frameRound  = frameRunes{'╭', '╮', '╰', '╯', '─', '│'}
	frameDouble = frameRunes{'╔', '╗', '╚', '╝', '═', '║'}
)

// dashed is f with dashed edges: a badge that waits.
func (f frameRunes) dashed() frameRunes {
	f.h, f.v = '┄', '┆'
	return f
}

// flat is n cells of f's flat edge.
func (f frameRunes) flat(n int) string { return strings.Repeat(string(f.h), n) }

// boxRows is a boxed badge, three rows: top between the upper corners, mid
// between the uprights and foot between the lower corners.
func boxRows(f frameRunes, top, mid, foot string) []string {
	return []string{
		string(f.tl) + top + string(f.tr),
		string(f.v) + mid + string(f.v),
		string(f.bl) + foot + string(f.br),
	}
}

// octagonRows is an eight-sided badge, five rows and nine cells wide: in1
// and in3 (five cells) are the rows above and below the middle, mid (seven
// cells) the middle and foot (five cells) the bottom edge.
func octagonRows(f frameRunes, in1, mid, in3, foot string) []string {
	return []string{
		" " + string(f.tl) + f.flat(5) + string(f.tr) + " ",
		string(f.tl) + string(f.br) + in1 + string(f.bl) + string(f.tr),
		string(f.v) + mid + string(f.v),
		string(f.bl) + string(f.tr) + in3 + string(f.tl) + string(f.br),
		" " + string(f.bl) + foot + string(f.br) + " ",
	}
}

// ringRows is the large ring with nothing on it, 17 by 9: the ring of a
// tech's card. Its foot mirrors its top.
func ringRows() []string {
	return []string{
		"    ▄▄█████▄▄    ",
		"  ▄█▓▒░░░░░▒▓█▄  ",
		" █▓░         ░▓█ ",
		"▐█▒           ▒█▌",
		"▐█▒           ▒█▌",
		"▐█▒           ▒█▌",
		" █▓░         ░▓█ ",
		"  ▀█▓▒░░░░░▒▓█▀  ",
		"    ▀▀█████▀▀    ",
	}
}
