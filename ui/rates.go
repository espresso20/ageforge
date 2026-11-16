package ui

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/pkg/textfmt"
)

// rates.go says a rate the way the dashboard shows it: per second, or per
// minute when a second would round it away. The engine's rates are per tick.

const (
	unitSecond = "/s"
	unitMinute = "/min"
	// perMinuteBelow is the size of a per-second rate under which the
	// dashboard turns to per minute.
	perMinuteBelow = 0.01
)

// secondly is a per-tick rate in a second.
func secondly(rate float64, tickMs int) float64 {
	if tickMs <= 0 {
		tickMs = 1000
	}
	return rate * 1000 / float64(tickMs)
}

// perSecond is a per-tick rate in the unit the dashboard shows it in: per
// second (the rate times 1000 over the tick's length in milliseconds), or
// per minute when its size is under perMinuteBelow a second. A tick of
// unknown length (0) counts as a second.
func perSecond(rate float64, tickMs int) (value float64, unit string) {
	value = secondly(rate, tickMs)
	if a := math.Abs(value); a != 0 && a < perMinuteBelow {
		return value * 60, unitMinute
	}
	return value, unitSecond
}

// rateUnits are the suffixes of a number as textfmt prints it, with what
// each stands for.
var rateUnits = map[string]float64{
	"K": 1e3, "M": 1e6, "B": 1e9, "T": 1e12, "Q": 1e15,
	"Qi": 1e18, "Sx": 1e21, "Sp": 1e24, "Oc": 1e27, "No": 1e30, "Dc": 1e33,
}

// tickRateInText finds a per-tick rate written in a description: a sign, a
// number with its suffix, then what it is a rate of ("+1.2K iron ore/tick").
var tickRateInText = regexp.MustCompile(`([+-])([0-9][0-9.,]*)(Qi|Sx|Sp|Oc|No|Dc|K|M|B|T|Q)? (\pL[\pL' -]*?)/tick`)

// perSecondText writes every per-tick rate in s as per second (or per
// minute), the rest of the text as it is.
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
		v, unit := perSecond(n, tickMs)
		sign := "+"
		if v < 0 {
			sign = "-"
		}
		return sign + textfmt.Number(math.Abs(v)) + " " + g[4] + unit
	})
}
