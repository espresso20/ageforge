package kit

import (
	"strings"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/theme"
)

// tags.go reads the inline tags the game's strings carry. The syntax is the
// one the strings were written in (the one the game's earlier text library used), so no string changes:
//
//	[fg:bg:attrs]   any part may be empty; "-" puts a part back to the base
//	[#rrggbb]       a literal color, in either place
//	[::b]           attributes: b bold, i italic, u underline, d dim, r reverse,
//	                s strike, l blink; the capital turns one off
//	["name"]        a region: the text after it, up to the next region tag,
//	                can be clicked by name; [""] ends it
//	[word[]         an escape: prints "[word]"
//
// A bracketed word that is not a known color is still a tag (it gives the
// text color), as it always was: text that wants a "[" writes the escape.

// A mark is where the style changes in the plain text, and how many bytes of
// the tagged text were tags there.
type mark struct {
	at     int // offset in the plain text
	skip   int // tagged bytes that are not text
	st     Style
	region string
}

// parsed is a tagged string taken apart: its text, and the marks in it.
type parsed struct {
	plain string
	marks []mark
}

func isAlnum(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

func isHex(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F'
}

func hexVal(b byte) uint8 {
	switch {
	case b >= 'a':
		return b - 'a' + 10
	case b >= 'A':
		return b - 'A' + 10
	}
	return b - '0'
}

// tagColor is the color a tag names. A name nobody knows gives the default
// for its place, the text color or the ground.
func tagColor(name string, ground bool) Color {
	if name[0] == '#' {
		return RGB(hexVal(name[1])<<4|hexVal(name[2]), hexVal(name[3])<<4|hexVal(name[4]), hexVal(name[5])<<4|hexVal(name[6]))
	}
	if r, ok := theme.TagRole(name); ok {
		return Role(r)
	}
	if c, ok := theme.CSSRGBA(name); ok {
		return RGB(c.R, c.G, c.B)
	}
	if ground {
		return Role(theme.RoleBackground)
	}
	return Role(theme.RoleText)
}

// colorField reads a color at s[i:]: a name, a "#rrggbb" or "-". It returns
// the offset after it, which is at a "]" or ":" when ok.
func colorField(s string, i int) (end int, name string, ok bool) {
	j := i
	switch {
	case s[i] == '-':
		j = i + 1
	case s[i] == '#':
		for j = i + 1; j < len(s) && isHex(s[j]); j++ {
		}
		if j-i != 7 {
			return 0, "", false
		}
	case isAlnum(s[i]):
		for j = i + 1; j < len(s) && isAlnum(s[j]); j++ {
		}
		if s[i] >= '0' && s[i] <= '9' {
			return 0, "", false
		}
	default:
		return 0, "", false
	}
	if j >= len(s) || (s[j] != ']' && s[j] != ':') {
		return 0, "", false
	}
	return j, s[i:j], true
}

var attrBits = map[byte]Attr{'b': Bold, 'i': Italic, 'l': Blink, 'd': Dim, 's': Strike, 'r': Reverse, 'u': Underline}

// oneTag reads the tag that starts at s[0] == '['. n is its length, 0 when it
// is not a tag.
func oneTag(s string, st, base Style, region string) (n int, out Style, outRegion string) {
	if len(s) < 3 {
		return 0, st, region
	}
	i := 1
	if s[i] == '"' { // a region
		j := i + 1
		for j < len(s) && (isAlnum(s[j]) || strings.IndexByte("_,;: -.", s[j]) >= 0) {
			j++
		}
		if j+1 < len(s) && s[j] == '"' && s[j+1] == ']' {
			return j + 2, st, s[i+1 : j]
		}
		return 0, st, region
	}
	out = st
	// the foreground
	if s[i] != ':' {
		end, name, ok := colorField(s, i)
		if !ok {
			return 0, st, region
		}
		if name == "-" {
			out.Fg = base.Fg
		} else {
			out.Fg = tagColor(name, false)
		}
		i = end
	}
	if s[i] == ']' {
		return i + 1, out, region
	}
	i++ // the ":" before the ground
	if i >= len(s) {
		return 0, st, region
	}
	if s[i] != ':' && s[i] != ']' {
		end, name, ok := colorField(s, i)
		if !ok {
			return 0, st, region
		}
		if name == "-" {
			out.Bg = base.Bg
		} else {
			out.Bg = tagColor(name, true)
		}
		i = end
	}
	if s[i] == ']' {
		return i + 1, out, region
	}
	i++ // the ":" before the attributes
	if i >= len(s) {
		return 0, st, region
	}
	switch {
	case s[i] == ']' || s[i] == ':':
	case s[i] == '-':
		out.Attr = base.Attr
		i++
	case strings.IndexByte("buildsrBUILDSR", s[i]) >= 0:
		for ; i < len(s) && strings.IndexByte("buildsrBUILDSR", s[i]) >= 0; i++ {
			if c := s[i]; c >= 'a' {
				out.Attr |= attrBits[c]
			} else {
				out.Attr &^= attrBits[c+('a'-'A')]
			}
		}
	default:
		return 0, st, region
	}
	if i >= len(s) {
		return 0, st, region
	}
	if s[i] == ']' {
		return i + 1, out, region
	}
	if s[i] != ':' {
		return 0, st, region
	}
	// A fourth part is a link, which the kit reads past: "-" or any text.
	i++
	if i < len(s) && s[i] == '-' {
		if i+1 < len(s) && s[i+1] == ']' {
			return i + 2, out, region
		}
		return 0, st, region
	}
	if j := strings.IndexByte(s[i:], ']'); j >= 0 {
		return i + j + 1, out, region
	}
	return 0, st, region
}

// tagRun reads every tag in a row from s[0] == '['.
func tagRun(s string, st, base Style, region string) (n int, out Style, outRegion string) {
	out, outRegion = st, region
	for n < len(s) && s[n] == '[' {
		l, o, r := oneTag(s[n:], out, base, outRegion)
		if l == 0 {
			break
		}
		n, out, outRegion = n+l, o, r
	}
	return n, out, outRegion
}

// escapeEnd reports whether s starts with an escaped tag, "[word[]" with any
// number of "[" before the "]", and where the word and the brackets end.
func escapeEnd(s string) (word, end int, ok bool) {
	j := 1
	for j < len(s) && s[j] != '[' && s[j] != ']' {
		j++
	}
	if j == 1 || j >= len(s) || s[j] != '[' {
		return 0, 0, false
	}
	k := j
	for k < len(s) && s[k] == '[' {
		k++
	}
	if k >= len(s) || s[k] != ']' {
		return 0, 0, false
	}
	return j, k + 1, true
}

// parse takes a tagged string apart. base is the style the text starts in and
// goes back to on "-".
func parse(s string, base Style) parsed {
	if strings.IndexByte(s, '[') < 0 {
		return parsed{plain: s}
	}
	var (
		plain  strings.Builder
		marks  []mark
		st     = base
		region string
	)
	skip := func(n int) {
		if k := len(marks) - 1; k >= 0 && marks[k].at == plain.Len() {
			marks[k].skip += n
			marks[k].st, marks[k].region = st, region
			return
		}
		marks = append(marks, mark{at: plain.Len(), skip: n, st: st, region: region})
	}
	for i := 0; i < len(s); {
		if s[i] != '[' {
			j := strings.IndexByte(s[i:], '[')
			if j < 0 {
				j = len(s) - i
			}
			plain.WriteString(s[i : i+j])
			i += j
			continue
		}
		if n, o, r := tagRun(s[i:], st, base, region); n > 0 {
			st, region = o, r
			skip(n)
			i += n
			if i >= len(s) || s[i] != '[' {
				continue
			}
		}
		if word, end, ok := escapeEnd(s[i:]); ok {
			// "[word[[]" prints "[word[]": one "[" of the closing run goes.
			plain.WriteString(s[i : i+word])
			skip(1)
			plain.WriteString(s[i+word+1 : i+end])
			i += end
			continue
		}
		plain.WriteByte('[')
		i++
	}
	return parsed{plain: plain.String(), marks: marks}
}

// Strip is s without its tags: the text as it prints.
func Strip(s string) string { return parse(s, Style{}).plain }

// Width is how many cells s takes on one line, tags not counted.
func Width(s string) int { return uniseg.StringWidth(Strip(s)) }

// Escape makes s print as it is written, by escaping whatever in it would
// read as a tag.
func Escape(s string) string {
	if strings.IndexByte(s, '[') < 0 {
		return s
	}
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		out.WriteByte(s[i])
		if s[i] != '[' {
			continue
		}
		// "[", then the characters a tag may hold, then any "[", then "]":
		// that would be a tag (or an escape of one), so one more "[" goes in.
		j := i + 1
		for j < len(s) && (isAlnum(s[j]) || strings.IndexByte("_,;: -.\"#", s[j]) >= 0) {
			j++
		}
		k := j
		for k < len(s) && s[k] == '[' {
			k++
		}
		if j > i+1 && k < len(s) && s[k] == ']' {
			out.WriteString(s[i+1 : k])
			out.WriteString("[]")
			i = k
		}
	}
	return out.String()
}
