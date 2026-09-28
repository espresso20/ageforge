package main

import (
	"github.com/espresso20/ageforge/config"
)

// style is the per-epoch dial set. One rendering grammar, seven dials: the
// drawing never changes shape between ages, only its ink.
type style struct {
	name  string // what the view calls itself this epoch
	chip  string // the resource bus's title
	lines lineSet
	boxes lineSet // node borders (often a different weight than the wires)
	// Pill end caps for one-row nodes.
	pillL, pillR rune
	// Particles: head then trail, drawn along a wire in flow order.
	head, trail rune
	// Fibre draws a moving lit segment instead of discrete packets.
	fibre bool
	// Quantum particles appear at two places at once.
	quantum     bool
	gaugeOn     rune
	gaugeOff    rune
	gaugeFrac   []rune // partial cells, low to high; nil = whole cells only
	spark       []rune // sparkline levels, low to high
	pinIn       rune   // wire arriving at a node
	pinOut      rune   // wire leaving a node
	sever       rune   // a cut link
	alarm       rune   // alert glyph (shape carries meaning without color)
	warn        rune
	tintAge     string // age whose AgePalettes accent tints borders
	tintAmount  float64
	speed       float64 // particle speed multiplier
	ascii       bool
}

// lineSet maps a 4-bit connection mask (N=1 E=2 S=4 W=8) to a rune.
type lineSet [16]rune

const (
	mN = 1 << iota
	mE
	mS
	mW
)

func mkLines(h, v, tl, tr, bl, br, lt, rt, tt, bt, x rune) lineSet {
	var s lineSet
	s[0] = ' '
	s[mE], s[mW], s[mE|mW] = h, h, h
	s[mN], s[mS], s[mN|mS] = v, v, v
	s[mE|mS] = tl
	s[mW|mS] = tr
	s[mN|mE] = bl
	s[mN|mW] = br
	s[mN|mE|mS] = lt
	s[mN|mW|mS] = rt
	s[mE|mW|mS] = tt
	s[mE|mW|mN] = bt
	s[mN|mE|mS|mW] = x
	return s
}

var (
	linesASCII   = mkLines('-', '|', '+', '+', '+', '+', '+', '+', '+', '+', '+')
	linesLight   = mkLines('─', '│', '┌', '┐', '└', '┘', '├', '┤', '┬', '┴', '┼')
	linesRound   = mkLines('─', '│', '╭', '╮', '╰', '╯', '├', '┤', '┬', '┴', '┼')
	linesHeavy   = mkLines('━', '┃', '┏', '┓', '┗', '┛', '┣', '┫', '┳', '┻', '╋')
	linesDouble  = mkLines('═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬')
	linesDashed  = mkLines('╌', '╎', '╭', '╮', '╰', '╯', '├', '┤', '┬', '┴', '┼')
	linesDotted  = mkLines('┄', '┆', '╭', '╮', '╰', '╯', '├', '┤', '┬', '┴', '┼')
	blocks8      = []rune("▁▂▃▄▅▆▇█")
	braille      = []rune("⣀⣤⣶⣿")
	fracBlocks   = []rune("▏▎▍▌▋▊▉")
)

func styleFor(age string) style {
	ep := config.EpochForAge(age)
	switch ep {
	case "stone_era":
		// Tally-slate: the chief's scratched tallies. Pure ASCII, packets are
		// pebbles.
		return style{
			name: "TALLY STONES", chip: "STORES",
			lines: linesASCII, boxes: linesASCII, pillL: '[', pillR: ']',
			head: 'o', trail: '.', gaugeOn: '#', gaugeOff: '.',
			spark: []rune("_.-~^"), pinIn: '>', pinOut: '-', sever: 'x', alarm: '!', warn: '?',
			tintAge: "stone_age", tintAmount: 0.35, speed: 0.6, ascii: true,
		}
	case "iron_era":
		// Scribe's ledger: ruled ink lines, rounded like a quill turns.
		return style{
			name: "SCRIBE'S LEDGER", chip: "GRANARY LEDGER",
			lines: linesRound, boxes: linesRound, pillL: '(', pillR: ')',
			head: '•', trail: '·', gaugeOn: '▰', gaugeOff: '▱',
			spark: blocks8, pinIn: '▸', pinOut: '╴', sever: '✕', alarm: '‼', warn: '!',
			tintAge: "medieval_age", tintAmount: 0.35, speed: 0.8,
		}
	case "steel_era":
		// Telegraph exchange / punch-card office: double-ruled cabinets,
		// square packets like punched holes.
		return style{
			name: "TELEGRAPH EXCHANGE", chip: "COUNTING HOUSE",
			lines: linesLight, boxes: linesDouble, pillL: '╟', pillR: '╢',
			head: '▪', trail: '·', gaugeOn: '▮', gaugeOff: '▯',
			spark: blocks8, pinIn: '▶', pinOut: '╼', sever: '╳', alarm: '▲', warn: '△',
			tintAge: "industrial_age", tintAmount: 0.4, speed: 1.0,
		}
	case "electric_era":
		// Switchboard: heavy bus bars, glowing current.
		return style{
			name: "SWITCHBOARD", chip: "BUS BAR",
			lines: linesLight, boxes: linesHeavy, pillL: '┫', pillR: '┣',
			head: '●', trail: '∙', gaugeOn: '█', gaugeOff: '░',
			gaugeFrac: fracBlocks, spark: blocks8, pinIn: '▶', pinOut: '╾', sever: '╳', alarm: '▲', warn: '△',
			tintAge: "electric_age", tintAmount: 0.4, speed: 1.2,
		}
	case "digital_era":
		// Network topology: thin fibre, lit pulses instead of packets.
		return style{
			name: "NETWORK TOPOLOGY", chip: "CORE SWITCH",
			lines: linesLight, boxes: linesLight, pillL: '┤', pillR: '├',
			head: '━', trail: '─', fibre: true, gaugeOn: '█', gaugeOff: '·',
			gaugeFrac: fracBlocks, spark: braille, pinIn: '▶', pinOut: '─', sever: '╳', alarm: '▲', warn: '△',
			tintAge: "digital_age", tintAmount: 0.4, speed: 1.6,
		}
	case "neon_era":
		return style{
			name: "NEON GRID", chip: "FUSION CORE",
			lines: linesRound, boxes: linesDouble, pillL: '╡', pillR: '╞',
			head: '◉', trail: '•', gaugeOn: '▰', gaugeOff: '▱',
			spark: braille, pinIn: '▶', pinOut: '═', sever: '╳', alarm: '▲', warn: '△',
			tintAge: "cyberpunk_age", tintAmount: 0.5, speed: 1.8,
		}
	default: // cosmic_era
		// Quantum lattice: dashed entanglement, packets in superposition.
		return style{
			name: "QUANTUM LATTICE", chip: "SINGULARITY",
			lines: linesDashed, boxes: linesRound, pillL: '‹', pillR: '›',
			head: '◆', quantum: true, gaugeOn: '━', gaugeOff: '╌',
			spark: braille, pinIn: '▸', pinOut: '╌', sever: '╳', alarm: '▲', warn: '◇',
			tintAge: "quantum_age", tintAmount: 0.5, speed: 2.0,
		}
	}
}
