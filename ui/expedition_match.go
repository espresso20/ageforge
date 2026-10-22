package ui

import (
	"strings"
	"unicode"
)

// expedition_match.go finds the offer a typed name means: `expedition
// hills` sends the one offer with "hills" in its title, and asks which one
// when several have.

// nameWords are the words of a title or of what was typed: lower case,
// punctuation gone, and "the", "of", "a" and "an" left out.
func nameWords(s string) []string {
	var out []string
	for _, w := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		switch w {
		case "the", "of", "a", "an":
		default:
			out = append(out, w)
		}
	}
	return out
}

// matchOffers returns the indexes of the titles that input means: those it
// names exactly, else those that start with it, else those in which every
// word typed begins a word of the title. The first of the three that finds
// any decides. Nothing for an empty input.
func matchOffers(titles []string, input string) []int {
	typed := nameWords(input)
	if len(typed) == 0 {
		return nil
	}
	want := strings.Join(typed, " ")
	words := make([][]string, len(titles))
	for i, t := range titles {
		words[i] = nameWords(t)
	}
	var exact, starts, parts []int
	for i, tw := range words {
		joined := strings.Join(tw, " ")
		switch {
		case joined == want:
			exact = append(exact, i)
		case strings.HasPrefix(joined, want):
			starts = append(starts, i)
		}
		all := true
		for _, w := range typed {
			found := false
			for _, t := range tw {
				if strings.HasPrefix(t, w) {
					found = true
					break
				}
			}
			all = all && found
		}
		if all {
			parts = append(parts, i)
		}
	}
	switch {
	case len(exact) > 0:
		return exact
	case len(starts) > 0:
		return starts
	}
	return parts
}
