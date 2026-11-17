package ui

import (
	"cmp"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// history_data.go turns the recorder and the account's run archive into what
// the History panel draws: series, marks and age spans for a range. It reads
// only what the engine keeps and invents nothing.

// histRange is how much of the run the chart shows.
type histRange int

const (
	rangeAge histRange = iota // the age in play
	rangeRun                  // the whole run so far
	rangeAll                  // the whole run, with one past run behind it
	histRanges
)

// histGroup is a heading of the measure picker.
type histGroup int

const (
	groupPeople histGroup = iota
	groupProgress
	groupRates
)

// histMeasure is one measure the chart can show.
type histMeasure struct {
	// Key is the recorder's column; Past the column a finished run keeps it
	// in, "" when finished runs do not keep it.
	Key, Past string
	Label     string
	Group     histGroup
	// Scale turns a recorded value into the one shown; Unit follows it.
	Scale float64
	Unit  string
	// Rate marks a per-tick rate, said per second through rates.go.
	Rate bool
	// Cap is the top of a measure with a fixed range (it stays linear), 0 for
	// one without.
	Cap float64
	// Offset is added after Scale (game speed is 1 plus the bonus).
	Offset float64
}

// histSeries is one measure over time: V[i] is the recorded value of the
// bucket ending at tick T[i].
type histSeries struct {
	T []int
	V []float64
}

// histSpan is an age: its key and the ticks it covers.
type histSpan struct {
	Key        string
	Start, End int
}

// histSource is what the panel reads: the run's recorder, the filed runs and
// the age a run starts in.
type histSource struct {
	hist     *game.RunHistory
	runs     []game.RunRecord
	firstAge string
}

// tierSeries is column col of tier t from tick `from` on. A column the tier
// does not have reads empty.
func tierSeries(t *game.HistoryTier, col string, from int) histSeries {
	var s histSeries
	c := t.Cols[col]
	for i := 0; i < t.N && i < len(c); i++ {
		tk := min(t.TickAt(i), t.End)
		if tk < from {
			continue
		}
		s.T = append(s.T, tk)
		s.V = append(s.V, float64(c[i]))
	}
	return s
}

// ageStart is the tick the age in play began at: the last age mark, 0 when
// the run has not advanced.
func ageStart(marks []game.HistoryMark) int {
	start := 0
	for _, m := range marks {
		if m.Kind == game.MarkAge {
			start = max(start, m.Tick)
		}
	}
	return start
}

// window is the ticks a range shows, from the start of its span to the run's
// newest sample.
func (s histSource) window(r histRange) (from, to int) {
	if s.hist == nil {
		return 0, 0
	}
	to = s.hist.Run.End
	if r == rangeAge {
		from = ageStart(s.hist.Marks)
	}
	return from, max(to, from)
}

// tier is the finest tier that reaches back to tick from: the recent ring
// when it does, else the run tier.
func (s histSource) tier(from int) *game.HistoryTier {
	if r := &s.hist.Recent; r.N > 0 && r.TickAt(0) <= from {
		return r
	}
	return &s.hist.Run
}

// live is the run in play's series of column col over range r.
func (s histSource) live(r histRange, col string) histSeries {
	if s.hist == nil {
		return histSeries{}
	}
	from, _ := s.window(r)
	return tierSeries(s.tier(from), col, from)
}

// marks are the run in play's marks inside range r.
func (s histSource) marks(r histRange) []game.HistoryMark {
	if s.hist == nil {
		return nil
	}
	from, _ := s.window(r)
	return marksFrom(s.hist.Marks, from)
}

// marksFrom keeps the marks at or after tick from.
func marksFrom(marks []game.HistoryMark, from int) []game.HistoryMark {
	var out []game.HistoryMark
	for _, m := range marks {
		if m.Tick >= from {
			out = append(out, m)
		}
	}
	return out
}

// spans are the ages inside range r, each cut to the range. The run starts in
// firstAge unless an age mark sits at its first tick.
func (s histSource) spans(r histRange) []histSpan {
	if s.hist == nil {
		return nil
	}
	from, to := s.window(r)
	return ageSpans(s.hist.Marks, s.firstAge, from, to)
}

// ageSpans cuts the run's ages out of its age marks: the first age runs from
// tick 0 to the first mark, each mark's age to the next, the last to `to`.
func ageSpans(marks []game.HistoryMark, first string, from, to int) []histSpan {
	var ages []game.HistoryMark
	for _, m := range marks {
		if m.Kind == game.MarkAge {
			ages = append(ages, m)
		}
	}
	slices.SortStableFunc(ages, func(a, b game.HistoryMark) int { return cmp.Compare(a.Tick, b.Tick) })
	if len(ages) == 0 || ages[0].Tick > 0 {
		ages = slices.Insert(ages, 0, game.HistoryMark{Kind: game.MarkAge, Key: first})
	}
	var out []histSpan
	for i, m := range ages {
		end := to
		if i+1 < len(ages) {
			end = ages[i+1].Tick
		}
		if end <= from || m.Tick > to || m.Key == "" {
			continue
		}
		out = append(out, histSpan{Key: m.Key, Start: max(m.Tick, from), End: min(end, to)})
	}
	return out
}

// kept reports whether a filed run still has a history to draw.
func kept(rec game.RunRecord) bool {
	return rec.History != nil && rec.History.Run.N > 0
}

// past is a finished run's series of column col, lined up from its own start
// (tick 0) like the run in play. A run with no kept history, or a measure
// finished runs do not keep, reads empty.
func past(rec game.RunRecord, col string) histSeries {
	if col == "" || !kept(rec) {
		return histSeries{}
	}
	return tierSeries(&rec.History.Run, col, 0)
}

// pastSpans are a finished run's ages over its whole length.
func (s histSource) pastSpans(rec game.RunRecord) []histSpan {
	if !kept(rec) {
		return nil
	}
	return ageSpans(rec.History.Marks, s.firstAge, 0, rec.History.Run.End)
}

// resourceRates are the resource-rate columns the recorder holds, sorted, the
// food rate (a core measure) left out.
func resourceRates(h *game.RunHistory) []string {
	if h == nil {
		return nil
	}
	var keys []string
	for k := range h.Run.Cols {
		if strings.HasPrefix(k, "rate.") && k != "rate.food" {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	return keys
}

// value is the shown figure for a recorded value, per second for a rate
// (perTick is how much a recorded unit is per second, see rateFactor).
func (m histMeasure) value(v, perSecond float64) float64 {
	if m.Rate {
		return v * perSecond
	}
	return v*m.Scale + m.Offset
}

// histMeasures is the measures the picker offers, in picker order: the ten
// core ones (keys 1 to 0), then the rate of each resource the recorder holds.
// label names a resource for its rate.
func histMeasures(h *game.RunHistory, label func(resKey string) string) []histMeasure {
	out := []histMeasure{
		{Key: "population", Past: "population", Label: txtPopulation.String(), Group: groupPeople, Scale: 1, Unit: txtWorkers.String()},
		{Key: "morale", Past: "morale", Label: txtMorale.String(), Group: groupPeople, Scale: 100, Unit: "%", Cap: 100},
		{Key: "faith", Past: "faith", Label: txtFaith.String(), Group: groupPeople, Scale: 1},
		{Key: "buildings", Past: "buildings", Label: txtBuildings.String(), Group: groupProgress, Scale: 1},
		{Key: "techs", Past: "techs", Label: txtTechsShort.String(), Group: groupProgress, Scale: 1},
		{Key: "expeditions", Past: "expeditions", Label: txtExpeditions.String(), Group: groupProgress, Scale: 1},
		{Key: "know_rate", Past: "know_rate", Label: txtResearchRate.String(), Group: groupRates, Rate: true},
		{Key: "prod_all", Past: "prod_all", Label: txtAllProduction.String(), Group: groupRates, Scale: 100, Unit: "%"},
		{Key: "tick_speed", Past: "tick_speed", Label: txtGameSpeed.String(), Group: groupRates, Scale: 1, Unit: "x", Offset: 1},
		{Key: "food_rate", Past: "rate.food", Label: txtFoodRate.String(), Group: groupRates, Rate: true},
	}
	for _, k := range resourceRates(h) {
		out = append(out, histMeasure{Key: k, Label: label(strings.TrimPrefix(k, "rate.")), Group: groupRates, Rate: true})
	}
	return out
}

// rateFactor is how many of a measure's recorded per-tick units make one unit
// per second (or per minute or hour, when a second would round the figure
// away) at a rate of peak, and the unit's suffix. The recorder keeps no tick
// length, so it reads at the base tick.
func rateFactor(peak float64) (k float64, unit string) {
	ms := int(game.BaseTickInterval.Milliseconds())
	v, unit := ratePer(peak, ms)
	if peak == 0 {
		return 1 / tickSeconds(ms), unit
	}
	return v / peak, unit
}

// Measure names. The ones already said on other screens keep their rows.
var (
	txtPopulation    = vibe.E("ui.overlay_history.renderHistoryOverlay.02", "Population")
	txtWorkers       = vibe.E("ui.overlay_history.renderHistoryOverlay.03", " workers")
	txtFoodRate      = vibe.E("ui.overlay_history.renderHistoryOverlay.04", "Food rate")
	txtResearchRate  = vibe.E("ui.overlay_history.renderHistoryOverlay.05", "Research rate")
	txtFaith         = vibe.E("ui.overlay_history.renderHistoryOverlay.06", "Faith")
	txtMorale        = vibe.E("ui.overlay_history.renderHistoryOverlay.07", "Morale")
	txtAllProduction = vibe.E("ui.overlay_history.renderHistoryOverlay.08", "All production")
	txtGameSpeed     = vibe.E("ui.overlay_history.renderHistoryOverlay.09", "Game speed")
)
