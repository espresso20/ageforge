package ui

import "strings"

// String is the grid as lines of text, trailing spaces kept.
func (g *glyphGrid[K]) String() string {
	var sb strings.Builder
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			sb.WriteRune(g.c[y*g.w+x].r)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// Lines are the rows as text, trailing spaces trimmed.
func (g *glyphGrid[K]) Lines() []string {
	lines := strings.Split(strings.TrimRight(g.String(), "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return lines
}
