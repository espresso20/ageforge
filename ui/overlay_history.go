package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// historyProvider is the OverlayProvider for the history overlay.
func historyProvider(state game.GameState, w int) string {
	return renderHistoryOverlay(state, w)
}

func renderHistoryOverlay(state game.GameState, w int) string {
	if state.History == nil || len(state.History.Samples) < 2 {
		return vibe.Line("ui.overlay_history.renderHistoryOverlay.01", "\n[gray] No history yet. Play for a while and check back.[-]\n")
	}

	samples := state.History.Samples
	n := len(samples)
	firstTick := samples[0].Tick
	lastTick := samples[n-1].Tick

	// collect age marker ticks and names
	// A marker holds the age's key (and saves already written hold keys
	// too): the footer shows the age's name.
	set := state.Ruleset()
	var markerTicks []int
	var markerNames []string
	for _, m := range state.History.AgeMarkers {
		markerTicks = append(markerTicks, m.Tick)
		name := m.AgeName
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
		label  string
		color  string
		values func(game.HistorySample) float64
		unit   string
		rate   bool
	}

	metrics := []metricDef{
		{vibe.Line("ui.overlay_history.renderHistoryOverlay.02", "Population"), "[cyan]", func(s game.HistorySample) float64 { return s.Population }, vibe.Line("ui.overlay_history.renderHistoryOverlay.03", " workers"), false},
		{vibe.Line("ui.overlay_history.renderHistoryOverlay.04", "Food rate"), "[green]", func(s game.HistorySample) float64 { return s.FoodRate }, "/tick", true},
		{vibe.Line("ui.overlay_history.renderHistoryOverlay.05", "Research rate"), "[yellow]", func(s game.HistorySample) float64 { return s.KnowRate }, "/tick", true},
		{vibe.Line("ui.overlay_history.renderHistoryOverlay.06", "Faith"), "[gold]", func(s game.HistorySample) float64 { return s.Faith }, "", false},
		{vibe.Line("ui.overlay_history.renderHistoryOverlay.07", "Morale"), "[yellow]", func(s game.HistorySample) float64 { return s.Morale * 100 }, "%", false},
		{vibe.Line("ui.overlay_history.renderHistoryOverlay.08", "All production"), "[warning]", func(s game.HistorySample) float64 { return s.ProdAll * 100 }, "%", false},
		{vibe.Line("ui.overlay_history.renderHistoryOverlay.09", "Game speed"), "[gray]", func(s game.HistorySample) float64 { return s.TickSpeed }, "x", false},
	}

	var sb strings.Builder

	sb.WriteString(vibe.Line("ui.overlay_history.renderHistoryOverlay.10", "\n[gold]═══ Civilization history ═══[-]\n"))
	sb.WriteString(vibe.Sprintf("ui.overlay_history.renderHistoryOverlay.11", "[gray]Ticks %d – %d  │  %d samples  │  age advances marked │[-]\n\n",
		firstTick, lastTick, n))

	for _, m := range metrics {
		vals := make([]float64, n)
		for i, s := range samples {
			vals[i] = m.values(s)
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
		sb.WriteString(vibe.Sprintf("ui.overlay_history.renderHistoryOverlay.12", "%s%-15s[-] %s %s%s  [gray]min:%-8s max:%s[-]\n",
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
		sb.WriteString(vibe.Line("ui.overlay_history.renderHistoryOverlay.13", "[gray]  Ages: "))
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
