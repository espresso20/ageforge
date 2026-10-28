package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// people_rich.go is the People box as the rich layout (dash_layout.go) draws
// it, in the right half of the Under construction row. It says what the
// classic box says (peopleBoxText) and more, in the rows that row has:
//
//	Pop 998/2402        41% full     the headcount over the homes
//	██████████░░░░░░░░░░░░░░░░░░     the homes, as the People panel's bar
//	Serfs                    517     the four kinds, in the age's words
//	Commoners                  0       (the reserve is the second)
//	Men-at-arms      428/430 · 12    posted/posts, and the soldiers out
//	Monks                     53
//	Held by: food                    what holds the population back
//	+0.45 food a tick feeds 3 more
//
// A box with fewer rows gives up the second Held-by line, then the bar, then
// the headcount (which moves to the title, peopleRichTitle).

// peopleRichMaxW bounds the width the box is written for. A box never drawn
// is written for a huge width (fitView.changed); the box is not drawn wider
// than this in practice.
const peopleRichMaxW = 64

// stackWidths shares w cells out between counts as the People panel's bar
// does (stackBar): each count its share of the homes (or of the people, when
// there are more people than homes), every non-zero count at least one cell.
func stackWidths(counts []int, homes, w int) []int {
	total := 0
	for _, n := range counts {
		total += n
	}
	out := make([]int, len(counts))
	whole := max(homes, total)
	if whole <= 0 || w <= 0 {
		return out
	}
	used := 0
	for i, n := range counts {
		if n <= 0 {
			continue
		}
		cells := int(float64(n)/float64(whole)*float64(w) + 0.5)
		cells = max(1, min(cells, w-used))
		out[i] = cells
		used += cells
	}
	return out
}

// peopleBarText is the homes as one bar w cells wide: each kind of person in
// its colour, the empty homes dim.
func peopleBarText(kinds []peopleKind, homes, w int) string {
	counts := make([]int, len(kinds))
	for i, k := range kinds {
		counts[i] = k.count
	}
	var sb strings.Builder
	used := 0
	for i, cells := range stackWidths(counts, homes, w) {
		if cells > 0 {
			sb.WriteString(theme.Tag(kinds[i].role) + strings.Repeat("█", cells) + "[-]")
			used += cells
		}
	}
	if used < w {
		sb.WriteString(theme.Paint(theme.RoleDim, strings.Repeat("░", w-used)))
	}
	return sb.String()
}

// peopleRichTitle is the rich box's title: " People ", with the headcount
// over the housing in it when the box has fewer than 7 rows inside, too few
// for the line that carries it.
func peopleRichTitle(state game.GameState, rows int) string {
	if rows >= 7 {
		return vibe.Line("ui.people_rich.peopleRichTitle.01", " People ")
	}
	ws := state.Workers
	return fmt.Sprintf(txtPeopleBoxCount.String(), textfmt.Number(float64(ws.TotalPop)), textfmt.Number(float64(ws.MaxPop)))
}

// peopleRichHeld is the two lines on what holds the population back: what it
// is, and a note on it. With nothing holding it back they say the homes have
// room.
func peopleRichHeld(ws game.WorkerState, w int) (first, second string) {
	gray := func(s string) string { return theme.Paint(theme.RoleDim, s) }
	held := func(by string) string {
		return gray(vibe.Line("ui.people_rich.peopleRichHeld.01", "Held by: ")) + theme.Paint(theme.RoleWarning, by)
	}
	switch ws.HeldBy {
	case game.HeldHousing:
		return held("housing"), gray(fitText(vibe.Line("ui.people_rich.peopleRichHeld.02", "every home is full"), w))
	case game.HeldFood:
		lines := heldByLines(ws, w)
		if len(lines) > 1 {
			return held("food"), gray(lines[1])
		}
		return held("food"), ""
	}
	empty := max(0, ws.MaxPop-ws.TotalPop)
	note := ""
	for _, t := range []string{vibe.Sprintf("ui.people_rich.peopleRichHeld.03", "%s homes empty", textfmt.Number(float64(empty))), vibe.Sprintf("ui.people_rich.peopleRichHeld.04", "%s empty", textfmt.Number(float64(empty)))} {
		if runeLen(t) <= w {
			note = t
			break
		}
	}
	return theme.Paint(theme.RolePositive, fitText(vibe.Line("ui.people_rich.peopleRichHeld.05", "Room to grow"), w)), gray(note)
}

// peopleRichText writes the rich People box for a box w cells wide and h
// rows tall inside.
func peopleRichText(state game.GameState, w, h int) string {
	w = min(w, peopleRichMaxW)
	ws := state.Workers
	kinds := peopleKinds(state)
	var lines []string

	// The headcount over the homes, and how full they are when it fits.
	pop, housing := fmt.Sprintf("%d", ws.TotalPop), fmt.Sprintf("%d", ws.MaxPop)
	popColor := theme.Tag(theme.RoleHighlight)
	if ws.TotalPop > ws.MaxPop {
		popColor = theme.Tag(theme.RoleNegative) // the scholars moved in past the homes
	}
	// The line is measured as it is drawn, in whichever script it is written.
	say := func() string {
		return vibe.Sprintf("ui.people_rich.peopleRichText.02", "%sPop[-] %s%s[-]/%s%s[-]", theme.Tag(theme.RoleDim), popColor, pop, theme.Tag(theme.RolePositive), housing)
	}
	popLine := say()
	if visibleLen(popLine) > w {
		pop, housing = textfmt.Number(float64(ws.TotalPop)), textfmt.Number(float64(ws.MaxPop))
		popLine = say()
	}
	if ws.MaxPop > 0 {
		share := vibe.Sprintf("ui.people_rich.peopleRichText.03", "%d%% full", ws.TotalPop*100/ws.MaxPop)
		if used := visibleLen(popLine); used+2+runeLen(share) <= w {
			popLine += strings.Repeat(" ", w-used-runeLen(share)) + theme.Paint(theme.RoleDim, share)
		}
	}

	first, second := peopleRichHeld(ws, w)
	switch {
	case h >= 8:
		lines = append(lines, popLine, peopleBarText(kinds, ws.MaxPop, w))
		for _, k := range kinds {
			lines = append(lines, k.boxLine(w))
		}
		lines = append(lines, first, second)
	case h == 7:
		lines = append(lines, popLine, peopleBarText(kinds, ws.MaxPop, w))
		for _, k := range kinds {
			lines = append(lines, k.boxLine(w))
		}
		lines = append(lines, first)
	case h == 6:
		lines = append(lines, popLine)
		for _, k := range kinds {
			lines = append(lines, k.boxLine(w))
		}
		lines = append(lines, first)
	default:
		for _, k := range kinds {
			lines = append(lines, k.boxLine(w))
		}
		if h > 4 {
			lines = append(lines, first)
		}
	}
	return strings.Join(lines, "\n") + "\n"
}
