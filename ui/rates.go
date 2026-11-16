package ui

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/pkg/textfmt"
)

// rates.go is the one way a rate is said on screen. The engine's rates are
// amounts per tick; a player sees them per second, or per minute or per hour
// when a second would round them away. A tick lasts state.TickIntervalMs:
// the base tick shortened by speed bonuses, the same paused or running.

const (
	unitSecond = "/s"
	unitMinute = "/min"
	unitHour   = "/h"
)

// rateFloor is the size under which a rate moves to the next unit.
const rateFloor = 0.1

// tickSeconds is how long a tick of tickMs milliseconds lasts, in seconds. A
// tick of unknown length (0) counts as a second.
func tickSeconds(tickMs int) float64 {
	if tickMs <= 0 {
		return 1
	}
	return float64(tickMs) / 1000
}

// ratePer is a per-tick rate in the unit a player reads it in: per second
// when its size is at least rateFloor a second, else per minute when that is
// at least rateFloor, else per hour. Zero is zero a second.
func ratePer(perTick float64, tickMs int) (value float64, unit string) {
	value = perTick / tickSeconds(tickMs)
	switch a := math.Abs(value); {
	case a == 0 || a >= rateFloor:
		return value, unitSecond
	case a*60 >= rateFloor:
		return value * 60, unitMinute
	}
	return value * 3600, unitHour
}

// rateFigure is a per-tick rate's number with its sign ("+0.45", "-3", "0")
// and its unit apart, for text that puts a word between them.
func rateFigure(perTick float64, tickMs int) (num, unit string) {
	v, unit := ratePer(perTick, tickMs)
	return signedRate(v), unit
}

// signedRate is a rate's number as it is printed: the sign always, a zero
// bare.
func signedRate(v float64) string {
	if v > 0 {
		return "+" + textfmt.Number(v)
	}
	return textfmt.Number(v)
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
	return signedRate(value) + unit, color
}

// rateUnits are the suffixes of a number as textfmt prints it, with what
// each stands for.
var rateUnits = map[string]float64{
	"K": 1e3, "M": 1e6, "B": 1e9, "T": 1e12, "Q": 1e15,
	"Qi": 1e18, "Sx": 1e21, "Sp": 1e24, "Oc": 1e27, "No": 1e30, "Dc": 1e33,
}

// tickRateInText finds a per-tick rate written in a text: an optional sign,
// a number with its suffix, an optional name of what it is a rate of, then
// the unit ("+1.2K iron ore/tick", "Food -3/tick").
var tickRateInText = regexp.MustCompile(`([+-]?)([0-9][0-9.,]*)(Qi|Sx|Sp|Oc|No|Dc|K|M|B|T|Q)?(?: (\pL[\pL' -]*?))?/tick`)

// perSecondText writes every per-tick rate in s per second (or minute, or
// hour), the rest of the text as it is.
func perSecondText(s string, tickMs int) string {
	return tickRateInText.ReplaceAllStringFunc(s, func(m string) string {
		g := tickRateInText.FindStringSubmatch(m)
		n, err := strconv.ParseFloat(strings.ReplaceAll(g[2], ",", ""), 64)
		if err != nil {
			return m
		}
		if mult, ok := rateUnits[g[3]]; ok {
			n *= mult
		}
		if g[1] == "-" {
			n = -n
		}
		v, unit := ratePer(n, tickMs)
		sign := g[1]
		if v < 0 {
			sign = "-"
		}
		name := ""
		if g[4] != "" {
			name = " " + g[4]
		}
		return sign + textfmt.Number(math.Abs(v)) + name + unit
	})
}
