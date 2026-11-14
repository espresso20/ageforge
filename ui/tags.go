package ui

import (
	"strings"

	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// tags.go keeps bracketed words visible.
//
// The kit reads any [word] as a color tag and removes it, so "Usage: sell
// <building> [count]" used to render as "Usage: sell <building> ", and a
// civilization's "[allied]" status vanished. safeTags runs at every place
// player text reaches a text view (the log, every panel, toasts, the status
// strip): it keeps the color tags the UI writes on purpose and escapes every
// other [word] so it prints literally.

// knownColorName reports whether name is a color an inline tag may use: one
// the theme owns (a role, a legacy alias) or a CSS color name.
func knownColorName(name string) bool {
	if _, ok := theme.TagRole(name); ok {
		return true
	}
	_, ok := theme.CSSRGBA(name)
	return ok
}

// isColorPart reports whether p is a valid fg or bg component of a style tag:
// empty, "-", "#rrggbb" or a known color name.
func isColorPart(p string) bool {
	if p == "" || p == "-" {
		return true
	}
	if p[0] == '#' {
		if len(p) != 7 {
			return false
		}
		for _, c := range p[1:] {
			if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
				return false
			}
		}
		return true
	}
	return knownColorName(p)
}

// tviewEatsTag reports whether the kit would parse "[content]" as a style tag
// and remove it from the output. Only letters, digits, '#', ':' and '-' can
// form one, and a name may not start with a digit.
func tviewEatsTag(content string) bool {
	if content == "" {
		return false
	}
	for _, c := range content {
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' ||
			c == '#' || c == ':' || c == '-'
		if !ok {
			return false
		}
	}
	fg := content
	if i := strings.IndexByte(content, ':'); i >= 0 {
		fg = content[:i]
	}
	if fg != "" && fg[0] >= '0' && fg[0] <= '9' {
		return false
	}
	return true
}

// isStyleTag reports whether "[content]" is a color/style tag the UI means:
// every fg/bg component is a known color, "-", empty or a hex color.
func isStyleTag(content string) bool {
	parts := strings.SplitN(content, ":", 4)
	if len(parts) > 3 {
		return true // URL tag ([::url]); never produced by player text
	}
	if !isColorPart(parts[0]) {
		return false
	}
	if len(parts) > 1 && !isColorPart(parts[1]) {
		return false
	}
	return true
}

// safeTags escapes every [word] in s that the kit would otherwise eat as a
// color tag, and leaves real color tags alone. Already-escaped text
// ("[count[]") is left as is, so calling it twice is harmless.
func safeTags(s string) string {
	if !strings.Contains(s, "[") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 8)
	i := 0
	for i < len(s) {
		if s[i] != '[' {
			b.WriteByte(s[i])
			i++
			continue
		}
		// Find the closing bracket, stopping at a nested '['.
		j := i + 1
		for j < len(s) && s[j] != ']' && s[j] != '[' {
			j++
		}
		if j >= len(s) || s[j] == '[' {
			b.WriteByte('[')
			i++
			continue
		}
		content := s[i+1 : j]
		if tviewEatsTag(content) && !isStyleTag(content) {
			// the escape: "[word[]" prints as "[word]".
			b.WriteString("[" + content + "[]")
		} else {
			b.WriteString(s[i : j+1])
		}
		i = j + 1
	}
	return b.String()
}

// pluralize returns word or its regular plural for n ("building",
// "buildings"). It is textfmt.Plural for nouns that just take an "s".
func pluralize(word string, n int) string {
	return textfmt.Plural(n, word, word+"s")
}

// lit escapes s completely so every bracket in it prints literally. Use it
// for text that must never carry color tags (usage lines, player-typed
// names).
func lit(s string) string {
	return kit.Escape(s)
}
