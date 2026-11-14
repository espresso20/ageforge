package kit

import (
	"strings"

	"github.com/rivo/uniseg"
)

// lines.go breaks tagged text into the lines it is painted as. The rules for
// where a line may break are Unicode's (UAX 14, through uniseg), and the way
// a long line wraps matches the text views the screens were written for, so
// a layout that counted cells still fits.

// Wrap is what happens to a line longer than its box.
type Wrap uint8

const (
	WrapNone Wrap = iota // cut at the edge
	WrapChar             // carry on below, breaking anywhere
	WrapWord             // carry on below, breaking between words where it can
)

// Align is where a line sits in its box.
type Align uint8

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

// tabSize is the cells between tab stops.
const tabSize = 4

// A glyph is one grapheme cluster, placed in a style.
type glyph struct {
	s      string
	w      int
	gross  int // bytes of the tagged text it stands for, tags before it included
	st     Style
	region string
	brk    uint8 // after it a line: 0 goes on, 1 may break, 2 must
}

// A Line is one painted line of text.
type Line struct {
	g []glyph
	w int // cells
	// from and to are the line's bytes in the tagged text.
	from, to int
}

// Width is the cells the line takes.
func (l *Line) Width() int { return l.w }

// glyphs steps through the text one cluster at a time.
func glyphs(p *parsed, base Style) []glyph {
	out := make([]glyph, 0, len(p.plain))
	var (
		st     = base
		region string
		mi     int
		off    int
		state  = -1
		rest   = p.plain
	)
	for len(rest) > 0 {
		var c string
		var b int
		c, rest, b, state = uniseg.StepString(rest, state)
		g := glyph{s: c, w: b >> uniseg.ShiftWidth, gross: len(c)}
		// The marks at this cluster's start change its style; one inside the
		// cluster (a tag between a letter and its mark) waits for the next.
		for ; mi < len(p.marks) && p.marks[mi].at <= off; mi++ {
			g.gross += p.marks[mi].skip
			st, region = p.marks[mi].st, p.marks[mi].region
		}
		g.st, g.region = st, region
		for ; mi < len(p.marks) && p.marks[mi].at < off+len(c); mi++ {
			g.gross += p.marks[mi].skip
			st, region = p.marks[mi].st, p.marks[mi].region
		}
		switch b & uniseg.MaskLine {
		case uniseg.LineCanBreak:
			g.brk = 1
		case uniseg.LineMustBreak:
			// The end of the text is not a break unless it ends in one.
			if rest != "" || uniseg.HasTrailingLineBreakInString(c) {
				g.brk = 2
			}
		}
		off += len(c)
		out = append(out, g)
	}
	return out
}

// layout breaks tagged text into lines no wider than width (0: any width).
func layout(s string, base Style, width int, wrap Wrap) []Line {
	if s == "" {
		return nil
	}
	p := parse(s, base)
	gs := glyphs(&p, base)
	if width <= 0 {
		wrap = WrapNone
	}
	var (
		lines             []Line
		start, pos        int // the line's first glyph, and its first byte
		w, bytes          int // the line so far
		opt, optW, optLen int // the last place it may break: glyph index, cells, bytes
	)
	emit := func(end, cells, n int) {
		lines = append(lines, Line{g: gs[start:end], w: cells, from: pos, to: pos + n})
		start, pos = end, pos+n
	}
	for i := range gs {
		g := &gs[i]
		if g.s == "\t" {
			g.w = tabSize - w%tabSize
		}
		if wrap != WrapNone && w+g.w > width {
			if optW == 0 {
				emit(i, w, bytes)
				w, bytes = 0, 0
			} else {
				emit(opt, optW, optLen)
				w, bytes = w-optW, bytes-optLen
			}
			opt, optW, optLen = 0, 0, 0
		}
		w += g.w
		bytes += g.gross
		switch g.brk {
		case 1:
			if wrap == WrapWord {
				opt, optW, optLen = i+1, w, bytes
			}
		case 2:
			emit(i+1, w, bytes)
			w, bytes, opt, optW, optLen = 0, 0, 0, 0, 0
		}
	}
	// What is left is the last line; after a final line break it is empty.
	return append(lines, Line{g: gs[start:], w: w, from: pos, to: len(s)})
}

// WordWrap breaks tagged text into lines of at most width cells, between
// words where it can. Each line keeps its tags, and line breaks are dropped.
func WordWrap(s string, width int) []string {
	if width <= 0 {
		return nil
	}
	if s == "" {
		return []string{""}
	}
	ls := layout(s, Style{}, width, WrapWord)
	out := make([]string, len(ls))
	for i := range ls {
		out[i] = strings.TrimRight(s[ls[i].from:ls[i].to], "\n\r")
	}
	return out
}

// printable reports whether a cluster is drawn: line breaks and other
// control characters are not.
func printable(g *glyph) bool { return g.s != "" && g.s[0] >= ' ' }
