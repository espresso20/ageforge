package ui

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// rates.go is the one way a rate is said on screen. The engine's rates are
// amounts per tick; a player sees them per second, or per minute or per hour
// when a second would round them away. A tick lasts state.TickIntervalMs:
// the base tick shortened by speed bonuses, the same paused or running.

// tickSeconds is how long a tick of tickMs milliseconds lasts, in seconds. A
// tick of unknown length (0) counts as a second.
func tickSeconds(tickMs int) float64 {
	if tickMs <= 0 {
		return 1
	}
	return float64(tickMs) / 1000
}

// ratePer is a per-tick rate in the unit a player reads it in (textfmt.RateUnit).
func ratePer(perTick float64, tickMs int) (value float64, unit string) {
	return textfmt.RateUnit(perTick / tickSeconds(tickMs))
}

// rateFigure is a per-tick rate's number with its sign ("+0.45", "-3", "0")
// and its unit apart, for text that puts a word between them.
func rateFigure(perTick float64, tickMs int) (num, unit string) {
	v, unit := ratePer(perTick, tickMs)
	return textfmt.RateNumber(v), unit
}

// rateText is a per-tick rate as "+3.25/s".
func rateText(perTick float64, tickMs int) string {
	num, unit := rateFigure(perTick, tickMs)
	return num + unit
}

// rateAmount is a per-tick rate with no sign, for a quantity that has no
// direction ("food drain: 3/s").
func rateAmount(perTick float64, tickMs int) string {
	v, unit := ratePer(math.Abs(perTick), tickMs)
	return textfmt.Number(v) + unit
}

// rateTag is rateText in the colour of its sign.
func rateTag(perTick float64, tickMs int) string {
	v, unit := ratePer(perTick, tickMs)
	text, color := rateParts(v, unit)
	return "[" + color + "]" + text + "[-]"
}

// rateParts is a rate (already per unit) as it is printed and the colour it
// is printed in, apart, for callers that lay the text out in a column
// before they colour it.
func rateParts(value float64, unit string) (text, color string) {
	switch {
	case value > 0:
		color = "green"
	case value < 0:
		color = "red"
	default:
		color = "gray"
	}
	return textfmt.RateNumber(value) + unit, color
}

// tickRateInText finds a per-tick rate written in a text: an optional sign
// and a plain number, then "/tick" ("Food +1/tick" in an awakening line).
var tickRateInText = regexp.MustCompile(`([+-]?)([0-9][0-9.,]*)/tick`)

// perSecondText writes every per-tick rate in s per second (or minute, or
// hour), the rest of the text as it is. It is for hand-written lines that
// state a rate as a signed number and "/tick"; everything the code generates
// is already per second.
func perSecondText(s string, tickMs int) string {
	return tickRateInText.ReplaceAllStringFunc(s, func(m string) string {
		g := tickRateInText.FindStringSubmatch(m)
		n, err := strconv.ParseFloat(strings.ReplaceAll(g[2], ",", ""), 64)
		if err != nil {
			return m
		}
		if g[1] == "-" {
			n = -n
		}
		v, unit := ratePer(n, tickMs)
		sign := g[1]
		if v < 0 {
			sign = "-"
		}
		return sign + textfmt.Number(math.Abs(v)) + unit
	})
}

// rateSource is one part of a resource's rate: what makes it, said through
// rateText, and its size for ordering.
type rateSource struct {
	text string
	size float64 // per tick
}

// rateSources lists the non-zero parts of a rate breakdown, labelled, in the
// order the rates command shows them.
func rateSources(b game.RateBreakdown, tickMs int) []rateSource {
	var out []rateSource
	add := func(v float64, text string) {
		if v != 0 {
			out = append(out, rateSource{text, v})
		}
	}
	rt := func(v float64) string { return rateText(v, tickMs) }
	add(b.BuildingRate, vibe.Sprintf("ui.input.cmdRates.03", "Buildings: %s", rt(b.BuildingRate)))
	// What worker output bonuses add (game/engine.go, gather_rate).
	add(b.WorkerRate, vibe.Sprintf("ui.input.cmdRates.04", "Worker bonus: %s", rt(b.WorkerRate)))
	// What the techs add: their bonus on what the resource's buildings make,
	// counted after the pools, and the flat output of a tech that is its
	// first source.
	add(b.ResearchRate, vibe.Sprintf("ui.input.cmdRates.05", "Research: %s", rt(b.ResearchRate)))
	add(b.EventRate, vibe.Sprintf("ui.input.cmdRates.06", "Events: %s", rt(b.EventRate)))
	add(b.TradeRate, vibe.Sprintf("ui.input.cmdRates.07", "Trade: %s", rt(b.TradeRate)))
	add(b.BonusRate, vibe.Sprintf("ui.input.cmdRates.08", "Bonuses: %s", rt(b.BonusRate)))
	add(b.LegacyRate, vibe.Sprintf("ui.input.cmdRates.09", "Cosmic Legacy: %s", rt(b.LegacyRate)))
	add(b.MasteryRate, vibe.Sprintf("ui.input.cmdRates.10", "Era Mastery: %s", rt(b.MasteryRate)))
	add(b.PrestigeRate, vibe.Sprintf("ui.input.cmdRates.11", "Prestige: %s", rt(b.PrestigeRate)))
	add(b.FoodDrain, vibe.Sprintf("ui.input.cmdRates.12", "Drain: %s", rt(b.FoodDrain)))
	return out
}
