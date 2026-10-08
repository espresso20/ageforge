package main

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// tokens.go finds, in a line of game text, the pieces a rewrite must carry
// over unchanged: format verbs, template slots, style tags, key names,
// commands in quotes, line breaks and glyphs. The export lists them in the
// keep column and the import holds a rewrite to them.

// tokenKind says what a kept piece is, which decides how strictly the import
// compares it.
type tokenKind int

const (
	tokVerb  tokenKind = iota // %s, %d, %.1f: the order matters
	tokSlot                   // {name}, ~
	tokTag                    // [red], [-], [gold::b], [count[]
	tokKey                    // Enter, Esc, Ctrl+K
	tokQuote                  // 'help': a command the player types
	tokBreak                  // a line break
	tokGlyph                  // ★, →, ·
)

type kept struct {
	kind tokenKind
	text string
	at   int // byte offset in the line
}

var (
	// A format verb. The space flag is left out on purpose so "50% of" is
	// not read as one.
	verbRe = regexp.MustCompile(`%(?:%|[-+#0]*(?:\d+|\*)?(?:\.(?:\d+|\*))?[sdvqwfxXcTtbeEgGU])`)
	// A template slot: {name}, {count}, {res_stores}.
	slotRe = regexp.MustCompile(`\{[A-Za-z_][A-Za-z0-9_]*\}`)
	// A bracketed word tview would take for a style tag, and its escaped
	// form ("[count[]").
	tagRe = regexp.MustCompile(`\[[A-Za-z0-9#:\-]*(?:\]|\[\])`)
	// The names of keys.
	keyRe = regexp.MustCompile(`\b(?:Ctrl\+[A-Za-z]|Enter|Esc|Tab|Shift|PgUp|PgDn|Backspace)\b`)
	// A command quoted in a sentence: Type 'catastrophe' to choose.
	quoteRe = regexp.MustCompile(`(?:^|[\s(])('[a-z][^'\n]{0,40}')(?:[\s.,;:)!?]|$)`)
)

// tokensOf lists the kept pieces of a line in the order they appear. slotMark
// adds the flavor catalogs' ~ to the slots.
func tokensOf(s string, slotMark bool) []kept {
	var out []kept
	taken := make([]bool, len(s))
	add := func(kind tokenKind, lo, hi int) {
		for i := lo; i < hi; i++ {
			if taken[i] {
				return
			}
		}
		for i := lo; i < hi; i++ {
			taken[i] = true
		}
		out = append(out, kept{kind, s[lo:hi], lo})
	}
	for _, m := range tagRe.FindAllStringIndex(s, -1) {
		add(tokTag, m[0], m[1])
	}
	for _, m := range verbRe.FindAllStringIndex(s, -1) {
		add(tokVerb, m[0], m[1])
	}
	for _, m := range slotRe.FindAllStringIndex(s, -1) {
		add(tokSlot, m[0], m[1])
	}
	for _, m := range keyRe.FindAllStringIndex(s, -1) {
		add(tokKey, m[0], m[1])
	}
	for _, m := range quoteRe.FindAllStringSubmatchIndex(s, -1) {
		add(tokQuote, m[2], m[3])
	}
	lead, trail := edges(s)
	glyphFrom := -1 // start of the run of glyphs being read
	flush := func(end int) {
		if glyphFrom >= 0 {
			add(tokGlyph, glyphFrom, end)
			glyphFrom = -1
		}
	}
	for i, r := range s {
		if !taken[i] && isGlyph(r) {
			if glyphFrom < 0 {
				glyphFrom = i
			}
			continue
		}
		flush(i)
		switch {
		case taken[i]:
		case r == '~' && slotMark:
			add(tokSlot, i, i+1)
		case r == '\n' && i >= len(lead) && i < len(s)-len(trail):
			add(tokBreak, i, i+1)
		}
	}
	flush(len(s))
	// In order of appearance.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].at < out[j-1].at; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// isGlyph reports whether r is a symbol the game draws rather than a letter
// or punctuation of a sentence.
func isGlyph(r rune) bool {
	if r < 0x80 {
		return false
	}
	switch r {
	case '·', '•':
		return true
	case '…', '–', '—', '‘', '’', '“', '”', '′', '″', '°':
		return false
	}
	if r >= 0xE000 && r <= 0xF8FF { // private use: Nerd Font icons
		return true
	}
	return unicode.IsSymbol(r) || unicode.Is(unicode.Braille, r)
}

// edges returns the white space a line starts and ends with.
func edges(s string) (lead, trail string) {
	body := strings.TrimLeftFunc(s, unicode.IsSpace)
	lead = s[:len(s)-len(body)]
	body2 := strings.TrimRightFunc(body, unicode.IsSpace)
	trail = body[len(body2):]
	return lead, trail
}

// keepColumn writes a line's kept pieces for the sheet.
func keepColumn(toks []kept) string {
	parts := make([]string, 0, len(toks))
	for _, t := range toks {
		if t.kind == tokBreak {
			parts = append(parts, `\n`)
			continue
		}
		parts = append(parts, t.text)
	}
	return strings.Join(parts, " ")
}

// bare returns a line with its tags, verbs and slots taken out: the words a
// reader reads.
func bare(s string) string {
	s = tagRe.ReplaceAllString(s, " ")
	s = verbRe.ReplaceAllString(s, " ")
	s = slotRe.ReplaceAllString(s, " ")
	return s
}

var wordRe = regexp.MustCompile(`\pL[\pL']*\pL|\pL`)

// shortWords are real words and abbreviations with no vowel in them, which
// would otherwise read as art.
var shortWords = map[string]bool{"pts": true, "bld": true, "hrs": true, "ctrl": true, "pgdn": true, "vs": true, "tv": true, "dr": true, "mr": true, "st": true, "nd": true, "rd": true, "th": true}

// realWord reports whether a run of letters is a word: two letters or more,
// a vowel among them, and no letter three times in a row. A sprite's rows
// ("ppppp", "FF S S FF", "xXxX") fail it.
func realWord(w string) bool {
	rs := []rune(w)
	if len(rs) < 2 {
		return false
	}
	if shortWords[strings.ToLower(w)] {
		return true
	}
	vowel := false
	for i, r := range rs {
		if i >= 2 && unicode.ToLower(r) == unicode.ToLower(rs[i-1]) && unicode.ToLower(r) == unicode.ToLower(rs[i-2]) {
			return false
		}
		if r >= 0x80 || strings.ContainsRune("aeiouyAEIOUY", r) {
			vowel = true
		}
	}
	return vowel
}

// hasWord reports whether a line holds at least one word.
func hasWord(s string) bool {
	for _, w := range wordRe.FindAllString(bare(s), -1) {
		if realWord(w) {
			return true
		}
	}
	return false
}

// countWords counts the words of a line, tags, verbs and slots left out.
func countWords(s string) int {
	n := 0
	for _, f := range strings.Fields(bare(s)) {
		if strings.IndexFunc(f, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) >= 0 {
			n++
		}
	}
	return n
}

// colorNames are the names a style tag may carry: tcell's colors and the
// theme's roles. It is the same set the game's own tag guard reads.
var colorNames = func() map[string]bool {
	m := map[string]bool{}
	for k := range tcell.ColorNames {
		m[k] = true
	}
	for k := range theme.TagNames() {
		m[k] = true
	}
	return m
}()

// isStyleTag reports whether a bracketed word is a style tag the game means
// ("[red]", "[-]", "[gold::b]"), as opposed to a word in brackets.
func isStyleTag(tok string) bool {
	if !strings.HasPrefix(tok, "[") || !strings.HasSuffix(tok, "]") || strings.HasSuffix(tok, "[]") {
		return false
	}
	content := tok[1 : len(tok)-1]
	if content == "" {
		return false
	}
	parts := strings.SplitN(content, ":", 4)
	if len(parts) > 3 {
		return true
	}
	color := func(p string) bool {
		if p == "" || p == "-" {
			return true
		}
		if p[0] == '#' {
			return len(p) == 7
		}
		return colorNames[p]
	}
	if !color(parts[0]) {
		return false
	}
	return len(parts) < 2 || color(parts[1])
}
