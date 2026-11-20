package ui

import (
	"image"
	"math"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/window/kit"
)

// fixtureRun records a run of n samples reaching `ages` ages, through the
// recorder's own calls. Scale stretches the figures so two runs differ.
func fixtureRun(n, ages int, scale float64) *game.RunHistory {
	h := game.NewRunHistory()
	keys := rules.Core().AgeKeys()
	for i := 1; i <= n; i++ {
		f := float64(i) / float64(n)
		h.Record(10*i, map[string]float64{
			"population":  scale * 4 * math.Exp(6*f),
			"morale":      0.55 + 0.25*math.Sin(9*f),
			"faith":       scale * 30 * f * f,
			"buildings":   math.Floor(scale * 200 * f),
			"techs":       math.Floor(scale * 90 * f),
			"expeditions": math.Floor(scale * 12 * f),
			"know_rate":   scale * 2 * math.Exp(6*f),
			"prod_all":    1 + 2*f,
			"tick_speed":  0.5 * f,
			"food_rate":   scale * 3 * math.Exp(5*f),
			"rate.food":   scale * 3 * math.Exp(5*f),
			"rate.stone":  scale * 1.5 * math.Exp(4*f),
			"rate.wood":   scale * 2 * math.Exp(4.5*f),
		})
	}
	for a := 1; a < ages && a < len(keys); a++ {
		h.Mark(10*n*a/ages, game.MarkAge, keys[a], 0)
	}
	h.Mark(10*n/3, game.MarkWonder, "pyramids", 0)
	h.Mark(10*n/2, game.MarkCatastrophe, "plague", 0)
	h.Mark(10*n*3/5, game.MarkHarbinger, "plague", 0)
	h.Mark(10*n*7/10, game.MarkEvent, "boom", 0)
	h.Mark(10*n*4/5, game.MarkAway, "", 300)
	return h
}

// fixtureRecord files a run as the account keeps it: the ten core columns of
// its run tier and the four kept kinds of mark.
func fixtureRecord(id, outcome string, n, ages int, scale float64) game.RunRecord {
	src := fixtureRun(n, ages, scale)
	hist := game.NewRunHistory()
	hist.Run = game.HistoryTier{Span: src.Run.Span, N: src.Run.N, Fill: src.Run.Fill, End: src.Run.End, Cols: map[string][]float32{}}
	for _, k := range []string{"population", "rate.food", "know_rate", "faith", "morale", "prod_all", "tick_speed", "buildings", "techs", "expeditions"} {
		hist.Run.Cols[k] = src.Run.Cols[k]
	}
	for _, m := range src.Marks {
		if m.Kind != game.MarkEvent && m.Kind != game.MarkAway {
			hist.Marks = append(hist.Marks, m)
		}
	}
	keys := rules.Core().AgeKeys()
	return game.RunRecord{ID: id, Run: "Hearth", Started: 1700000000, Ended: 1700000000 + int64(n*20), Tick: 10 * n,
		Outcome: outcome, Age: keys[min(ages-1, len(keys)-1)], Built: 200, Recruited: 400, Researched: 90, Expeditions: 12, History: hist}
}

// histFixturePanel is a History panel over the given run and filed runs, laid
// out at w x h cells.
func histFixturePanel(w, h int, run *game.RunHistory, runs []game.RunRecord) *historyPanel {
	p := newHistoryPanel()
	p.src = func() (*game.RunHistory, game.RunArchiveView, string) {
		return run, game.RunArchiveView{RunID: "now", Runs: runs}, "Hearth"
	}
	p.open(game.GameState{})
	p.Layout(image.Rect(0, 0, w, h))
	return p
}

// histPaint paints the panel on a fresh buffer.
func histPaint(p *historyPanel) *kit.Buffer {
	b := kit.NewBuffer(p.w, p.h)
	p.Paint(kit.NewPainter(b))
	return b
}
