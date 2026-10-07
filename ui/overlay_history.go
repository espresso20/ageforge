package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
)

// historyProvider is the OverlayProvider for the history overlay.
func historyProvider(state game.GameState, w int) string {
	return renderHistoryOverlay(state, w)
}

func renderHistoryOverlay(state game.GameState, w int) string {
	if state.History == nil || len(state.History.Samples) < 2 {
		return "\n[gray] No history yet. Play for a while and check back.[-]\n"
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

	// 85% overlay, minus y-axis label (8 chars), border/padding (6 chars)
	graphW := int(float64(w)*0.85) - 14
	if graphW < 20 {
		graphW = 20
	}
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
		{"Population", "[cyan]", func(s game.HistorySample) float64 { return s.Population }, " workers", false},
		{"Food rate", "[green]", func(s game.HistorySample) float64 { return s.FoodRate }, "/tick", true},
		{"Knowledge rate", "[yellow]", func(s game.HistorySample) float64 { return s.KnowRate }, "/tick", true},
		{"Faith", "[gold]", func(s game.HistorySample) float64 { return s.Faith }, "", false},
		{"Morale", "[yellow]", func(s game.HistorySample) float64 { return s.Morale * 100 }, "%", false},
		{"All production", "[warning]", func(s game.HistorySample) float64 { return s.ProdAll * 100 }, "%", false},
		{"Game speed", "[gray]", func(s game.HistorySample) float64 { return s.TickSpeed }, "x", false},
	}

	var sb strings.Builder

	sb.WriteString("\n[gold]═══ Civilization history ═══[-]\n")
	sb.WriteString(fmt.Sprintf("[gray]Ticks %d – %d  │  %d samples  │  age advances marked │[-]\n\n",
		firstTick, lastTick, n))

	for _, m := range metrics {
		vals := make([]float64, n)
		for i, s := range samples {
			vals[i] = m.values(s)
		}

		minV, maxV := vals[0], vals[0]
		for _, v := range vals {
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}

		cur := vals[n-1]
		prev := vals[n-2]
		trend := "[white]→[-]"
		if cur > prev*1.02 {
			trend = "[green]↑[-]"
		}
		if cur < prev*0.98 {
			trend = "[red]↓[-]"
		}

		curStr := fmtVal(cur)
		if m.rate && cur >= 0 {
			curStr = "+" + curStr
		}

		// header line
		sb.WriteString(fmt.Sprintf("%s%-15s[-] %s %s%s  [gray]min:%-8s max:%s[-]\n",
			m.color, m.label, trend, curStr, m.unit,
			fmtVal(minV), fmtVal(maxV)))

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
		sb.WriteString("[gray]  Ages: ")
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
	top, mid, bottom = fmtVal(maxV), fmtVal((maxV+minV)/2), fmtVal(minV)
	if mid == bottom || mid == top {
		mid = ""
	}
	if top == bottom {
		top = ""
	}
	return top, mid, bottom
}
