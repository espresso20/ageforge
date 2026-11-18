package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
)

// historyProvider is the OverlayProvider for the history overlay. It fetches
// the run's recorder from the engine each time the panel is drawn.
func historyProvider(engine *game.GameEngine) OverlayProvider {
	return func(state game.GameState, w int) string {
		return renderHistoryOverlay(state, engine.History(), w)
	}
}

// txtMorale is said the same way on several screens, so it has one id.
var txtMorale = rowid.E("ui.overlay_history.renderHistoryOverlay.07", "Morale")

func renderHistoryOverlay(state game.GameState, hist *game.RunHistory, w int) string {
	if hist == nil || hist.Recent.Len() < 2 {
		return rowid.Line("ui.overlay_history.renderHistoryOverlay.01", "\n[gray] No history yet. Play for a while and check back.[-]\n")
	}

	recent := &hist.Recent
	n := recent.Len()
	firstTick := recent.TickAt(0)
	lastTick := recent.TickAt(n - 1)

	// collect age marker ticks and names
	// A marker holds the age's key (and saves already written hold keys
	// too): the footer shows the age's name.
	set := state.Ruleset()
	var markerTicks []int
	var markerNames []string
	for _, m := range hist.Marks {
		if m.Kind != game.MarkAge || m.Tick < firstTick {
			continue
		}
		markerTicks = append(markerTicks, m.Tick)
		name := m.Key
		if def, ok := set.Age(name); ok {
			name = def.Name
		}
		markerNames = append(markerNames, name)
	}

	// The panel's text width, minus y-axis label (8 chars) and the gap after it (6 chars)
	graphW := max(overlayTextWidth(w)-14, 20)
	const graphH = 4 // rows tall per graph

	ageCols := AgeMarkerCols(markerTicks, firstTick, lastTick, graphW)

	type metricDef struct {
		label string
		color string
		key   string
		scale float64
		unit  string
		rate  bool
	}

	metrics := []metricDef{
		{rowid.Line("ui.overlay_history.renderHistoryOverlay.02", "Population"), "[cyan]", "population", 1, rowid.Line("ui.overlay_history.renderHistoryOverlay.03", " workers"), false},
		{rowid.Line("ui.overlay_history.renderHistoryOverlay.04", "Food rate"), "[green]", "food_rate", 1, "/tick", true},
		{rowid.Line("ui.overlay_history.renderHistoryOverlay.05", "Research rate"), "[yellow]", "know_rate", 1, "/tick", true},
		{rowid.Line("ui.overlay_history.renderHistoryOverlay.06", "Faith"), "[gold]", "faith", 1, "", false},
		{txtMorale.String(), "[yellow]", "morale", 100, "%", false},
		{rowid.Line("ui.overlay_history.renderHistoryOverlay.08", "All production"), "[warning]", "prod_all", 100, "%", false},
		{rowid.Line("ui.overlay_history.renderHistoryOverlay.09", "Game speed"), "[gray]", "tick_speed", 1, "x", false},
	}

	var sb strings.Builder

	sb.WriteString(rowid.Line("ui.overlay_history.renderHistoryOverlay.10", "\n[gold]═══ Civilization history ═══[-]\n"))
	sb.WriteString(rowid.Sprintf("ui.overlay_history.renderHistoryOverlay.11", "[gray]Ticks %d – %d  │  %d samples  │  age advances marked │[-]\n\n",
		firstTick, lastTick, n))

	for _, m := range metrics {
		vals := make([]float64, n)
		for i, v := range recent.Cols[m.key] {
			vals[i] = float64(v) * m.scale
		}

		minV, maxV := slices.Min(vals), slices.Max(vals)

		cur := vals[n-1]
		prev := vals[n-2]
		trend := "[white]→[-]"
		if cur > prev*1.02 {
			trend = "[green]↑[-]"
		}
		if cur < prev*0.98 {
			trend = "[red]↓[-]"
		}

		curStr := textfmt.Number(cur)
		if m.rate && cur >= 0 {
			curStr = "+" + curStr
		}

		// header line
		sb.WriteString(rowid.Sprintf("ui.overlay_history.renderHistoryOverlay.12", "%s%-15s[-] %s %s%s  [gray]min:%-8s max:%s[-]\n",
			m.color, m.label, trend, curStr, m.unit,
			textfmt.Number(minV), textfmt.Number(maxV)))

		// graph lines with y-axis labels
		lines := BrailleLine(vals, graphW, graphH, ageCols)
		top, mid, bottom := historyAxisLabels(minV, maxV)
		yLabels := []string{
			fmt.Sprintf("[gray]%-6s[-]", top),
			"      ",
			fmt.Sprintf("[gray]%-6s[-]", mid),
			"      ",
		}
		for i, line := range lines {
			sb.WriteString(fmt.Sprintf("  %s %s%s[-]\n", yLabels[i], m.color, line))
		}
		sb.WriteString(fmt.Sprintf("  [gray]%-6s %s[-]\n", bottom, strings.Repeat("─", graphW)))
		sb.WriteString("\n")
	}

	// x-axis age labels
	if len(markerNames) > 0 {
		sb.WriteString(rowid.Line("ui.overlay_history.renderHistoryOverlay.13", "[gray]  Ages: "))
		for i, name := range markerNames {
			if i > 0 {
				sb.WriteString(" → ")
			}
			sb.WriteString(name)
		}
		sb.WriteString("[-]\n")
	}

	return sb.String()
}

// historyAxisLabels is a graph's three y-axis labels: its top, its middle
// and its base line. A label that would only repeat the one below it is
// left blank, so a flat series (drawn along the base line) is labeled once
// and a series that barely moves does not print the same number twice.
func historyAxisLabels(minV, maxV float64) (top, mid, bottom string) {
	top, mid, bottom = textfmt.Number(maxV), textfmt.Number((maxV+minV)/2), textfmt.Number(minV)
	if mid == bottom || mid == top {
		mid = ""
	}
	if top == bottom {
		top = ""
	}
	return top, mid, bottom
}
