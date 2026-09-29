package mapmodel

import (
	"strings"

	"github.com/espresso20/ageforge/game"
)

// DayTicks is one in-game day on the maps: an hour of play at 1x.
const DayTicks = 1800

// Clock is the time of day, from the game tick.
type Clock struct {
	Day          int     // 1-based day count
	TOD          float64 // time of day 0..1, 0 = midnight
	Hour, Minute int
	// Daylight is 1 at noon and 0 at night; Twilight peaks at dawn and
	// dusk; Night is what is left. They drive lit windows and skies.
	Daylight, Twilight, Night float64
	Phase                     string // "night", "dawn", "day" or "dusk"
}

// ClockAt returns the clock for a tick. Day one starts at 09:00.
func ClockAt(tick int) Clock {
	if tick < 0 {
		tick = 0
	}
	t := tick + DayTicks*3/8
	c := Clock{Day: t/DayTicks + 1, TOD: float64(t%DayTicks) / DayTicks}
	mins := int(float64(c.TOD * 1440))
	c.Hour, c.Minute = mins/60, mins%60
	c.Daylight = clamp01(smoothRange(0.21, 0.31, c.TOD) - smoothRange(0.69, 0.79, c.TOD))
	c.Twilight = bump(c.TOD, 0.25, 0.09)
	if d := bump(c.TOD, 0.75, 0.09); d > c.Twilight {
		c.Twilight = d
	}
	lit := c.Daylight
	if tw := float64(0.85 * c.Twilight); tw > lit {
		lit = tw
	}
	c.Night = clamp01(1 - lit)
	switch {
	case c.Night > 0.7:
		c.Phase = "night"
	case c.Daylight > 0.7:
		c.Phase = "day"
	case c.TOD < 0.5:
		c.Phase = "dawn"
	default:
		c.Phase = "dusk"
	}
	return c
}

// ClockAtTOD returns a clock pinned to a time of day (captures, the
// skyline's time key).
func ClockAtTOD(day int, tod float64) Clock {
	tod = tod - float64(int(tod))
	if tod < 0 {
		tod++
	}
	tick := (day-1)*DayTicks + int(float64(tod*DayTicks)) - DayTicks*3/8
	for tick < 0 {
		tick += DayTicks
	}
	return ClockAt(tick)
}

// String is "22:55".
func (c Clock) String() string {
	return string([]byte{byte('0' + c.Hour/10), byte('0' + c.Hour%10), ':', byte('0' + c.Minute/10), byte('0' + c.Minute%10)})
}

func smoothRange(a, b, t float64) float64 { return smooth(clamp01((t - a) / (b - a))) }

// bump is a smooth hump of half-width w centred on c, 1 at c, 0 beyond w.
func bump(t, c, w float64) float64 {
	d := (t - c) / w
	if d < -1 || d > 1 {
		return 0
	}
	v := 1 - float64(d*d)
	return float64(v * v)
}

// WeatherKind is the sky's weather.
type WeatherKind uint8

const (
	Clear WeatherKind = iota
	Cloudy
	Rain
	Storm
	Snow
	Smog
)

func (k WeatherKind) String() string {
	return [...]string{"clear", "cloudy", "rain", "storm", "snow", "smog"}[k]
}

// Weather is the day's weather: set by active events when one is
// weather-like, seeded per day otherwise.
type Weather struct {
	Kind WeatherKind
	// Cause names the event behind it, "" for ordinary weather.
	Cause string
}

func weatherFor(st *game.GameState, c Clock, epoch int) Weather {
	for _, e := range st.ActiveEvents {
		k := strings.ToLower(e.Key)
		switch {
		case strings.Contains(k, "storm"):
			return Weather{Kind: Storm, Cause: e.Name}
		case strings.Contains(k, "flood"), strings.Contains(k, "plague"), strings.Contains(k, "blight"), strings.Contains(k, "rain"):
			return Weather{Kind: Rain, Cause: e.Name}
		case strings.Contains(k, "winter"), strings.Contains(k, "frost"), strings.Contains(k, "ice"):
			return Weather{Kind: Snow, Cause: e.Name}
		case strings.Contains(k, "drought"), strings.Contains(k, "festival"):
			return Weather{Kind: Clear, Cause: e.Name}
		}
	}
	switch Hash(int64(c.Day), st.Seed, 17) % 20 {
	case 13, 14, 15:
		return Weather{Kind: Cloudy}
	case 16, 17:
		return Weather{Kind: Rain}
	case 18:
		if epoch == 2 || epoch == 3 {
			return Weather{Kind: Smog} // the steel and electric eras' coal haze
		}
		return Weather{Kind: Cloudy}
	case 19:
		return Weather{Kind: Storm}
	}
	return Weather{Kind: Clear}
}
