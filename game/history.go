package game

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math"
	"slices"
)

const (
	historyInterval  = 10  // ticks between samples (~20s at 1x, ~13s at 1.5x)
	historyRecentLen = 300 // buckets in the recent ring
	historyRunMax    = 512 // buckets the run tier may hold before it merges
	historyMarksMax  = 512
)

// The kinds of mark a run's history carries.
const (
	MarkAge         = "age"         // Key: the age's key
	MarkWonder      = "wonder"      // Key: the building's key
	MarkCatastrophe = "catastrophe" // Key: the epoch's key, or the Last Passage's own
	MarkHarbinger   = "harbinger"   // Key: the epoch the thread pointed at
	MarkEvent       = "event"       // Key: the epoch event's key
	MarkAway        = "away"        // Tick: where the catch-up began; Ticks: the time credited
)

// historyCounts names the columns whose bucket value is the last sample, not
// the mean: whole-number counts that only grow within a run.
var historyCounts = map[string]bool{"buildings": true, "techs": true, "expeditions": true}

// HistoryMark is a dated event of the run.
type HistoryMark struct {
	Tick  int    `json:"t"`
	Kind  string `json:"k"`
	Key   string `json:"key,omitempty"`
	Ticks int    `json:"n,omitempty"`
}

// HistoryTier is one resolution of the run's measures: N buckets of Span
// ticks each, as columns keyed by measure. The last bucket holds Fill samples
// so far (a full bucket holds Span/historyInterval). A column that starts late
// reads zero before it.
type HistoryTier struct {
	Span int
	N    int
	Fill int
	End  int // tick of the newest sample
	Cols map[string][]float32
}

// Len is the number of buckets.
func (t *HistoryTier) Len() int { return t.N }

// TickAt is the tick bucket i ends at (the open bucket's nominal end).
func (t *HistoryTier) TickAt(i int) int {
	if t.N == 0 {
		return 0
	}
	return t.End + t.Span - t.Fill*historyInterval - (t.N-1-i)*t.Span
}

// add folds one sample into the tier. A ring tier (limit buckets of one
// sample each) drops its oldest bucket when full; the run tier merges
// neighbours in pairs and doubles the span.
func (t *HistoryTier) add(tick int, vals map[string]float64, limit int, ring bool) {
	if t.N == 0 || t.Fill >= t.Span/historyInterval {
		if t.N == limit {
			t.shrink(ring)
		}
		for k, c := range t.Cols {
			t.Cols[k] = append(c, 0)
		}
		t.N++
		t.Fill = 0
	}
	for k, v := range vals {
		c := t.Cols[k]
		if c == nil {
			c = make([]float32, t.N)
		}
		if last := &c[t.N-1]; historyCounts[k] {
			*last = float32(v)
		} else {
			*last = float32(float64(*last) + (v-float64(*last))/float64(t.Fill+1))
		}
		t.Cols[k] = c
	}
	t.Fill++
	t.End = tick
}

// shrink makes room. A ring tier drops its oldest bucket. The run tier merges
// neighbours in pairs and doubles the span: a mean column takes the mean of the
// pair weighted by the samples each bucket holds (the open last bucket may hold
// fewer than a full one), a count column the later value. An odd bucket stands
// alone as the new last bucket. End and the first bucket's start do not move,
// and Fill is the samples the new last bucket holds.
func (t *HistoryTier) shrink(ring bool) {
	if ring {
		for k, c := range t.Cols {
			t.Cols[k] = c[:copy(c, c[1:])]
		}
		t.N--
		return
	}
	full := t.Span / historyInterval
	pairs, odd := t.N/2, t.N%2 == 1
	keep := (t.N + 1) / 2
	for k, c := range t.Cols {
		for i := 0; i < pairs; i++ {
			a, b := c[2*i], c[2*i+1]
			if !historyCounts[k] {
				wb := full
				if 2*i+1 == t.N-1 {
					wb = t.Fill
				}
				b = float32((float64(a)*float64(full) + float64(b)*float64(wb)) / float64(full+wb))
			}
			c[i] = b
		}
		if odd {
			c[pairs] = c[t.N-1]
		}
		t.Cols[k] = c[:keep]
	}
	if !odd {
		t.Fill += full
	}
	t.N = keep
	t.Span *= 2
}

// thin keeps at most limit buckets, merging as the run tier does when full.
func (t *HistoryTier) thin(limit int) {
	for t.N > limit && limit > 0 {
		t.shrink(false)
	}
}

func (t *HistoryTier) clone() HistoryTier {
	out := *t
	out.Cols = make(map[string][]float32, len(t.Cols))
	for k, c := range t.Cols {
		out.Cols[k] = slices.Clone(c)
	}
	return out
}

// historyTierJSON is a tier as saved: each column is its float32 values packed
// little-endian and base64-encoded, so large numbers do not print as decimals.
type historyTierJSON struct {
	Span int               `json:"span"`
	N    int               `json:"n"`
	Fill int               `json:"fill"`
	End  int               `json:"end"`
	Cols map[string]string `json:"cols,omitempty"`
}

func (t HistoryTier) MarshalJSON() ([]byte, error) {
	out := historyTierJSON{Span: t.Span, N: t.N, Fill: t.Fill, End: t.End, Cols: make(map[string]string, len(t.Cols))}
	for k, c := range t.Cols {
		b := make([]byte, 4*len(c))
		for i, v := range c {
			binary.LittleEndian.PutUint32(b[4*i:], math.Float32bits(v))
		}
		out.Cols[k] = base64.StdEncoding.EncodeToString(b)
	}
	return json.Marshal(out)
}

// UnmarshalJSON reads a tier back. A tier that does not read (a column of the
// wrong length, bad base64) comes back empty: the history is a record, and a
// damaged one must not stop a save from loading.
func (t *HistoryTier) UnmarshalJSON(data []byte) error {
	var in historyTierJSON
	*t = HistoryTier{Span: historyInterval, Cols: map[string][]float32{}}
	if json.Unmarshal(data, &in) != nil || in.Span < historyInterval || in.N < 0 {
		return nil
	}
	cols := make(map[string][]float32, len(in.Cols))
	for k, s := range in.Cols {
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil || len(b) != 4*in.N {
			return nil
		}
		c := make([]float32, in.N)
		for i := range c {
			c[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[4*i:]))
		}
		cols[k] = c
	}
	*t = HistoryTier{Span: in.Span, N: in.N, Fill: in.Fill, End: in.End, Cols: cols}
	return nil
}

// RunHistory is the recorder: the recent ring at the sampling cadence, the
// whole run in at most historyRunMax buckets, and the run's marks.
type RunHistory struct {
	Recent HistoryTier   `json:"recent"`
	Run    HistoryTier   `json:"run"`
	Marks  []HistoryMark `json:"marks,omitempty"`
}

// NewRunHistory returns an empty recorder.
func NewRunHistory() *RunHistory {
	return &RunHistory{
		Recent: HistoryTier{Span: historyInterval, Cols: map[string][]float32{}},
		Run:    HistoryTier{Span: historyInterval, Cols: map[string][]float32{}},
	}
}

// Record takes the sample for tick into both tiers.
func (h *RunHistory) Record(tick int, vals map[string]float64) {
	h.Recent.add(tick, vals, historyRecentLen, true)
	h.Run.add(tick, vals, historyRunMax, false)
}

// Mark files a dated event. Past historyMarksMax the oldest event goes, then
// the oldest away mark; ages, wonders, catastrophes and harbingers stay.
func (h *RunHistory) Mark(tick int, kind, key string, ticks int) {
	h.Marks = append(h.Marks, HistoryMark{Tick: tick, Kind: kind, Key: key, Ticks: ticks})
	if len(h.Marks) <= historyMarksMax {
		return
	}
	for _, kind := range []string{MarkEvent, MarkAway} {
		if i := slices.IndexFunc(h.Marks, func(m HistoryMark) bool { return m.Kind == kind }); i >= 0 {
			h.Marks = slices.Delete(h.Marks, i, i+1)
			return
		}
	}
}

// Clone is a deep copy, nil for nil.
func (h *RunHistory) Clone() *RunHistory {
	if h == nil {
		return nil
	}
	return &RunHistory{Recent: h.Recent.clone(), Run: h.Run.clone(), Marks: slices.Clone(h.Marks)}
}

// History returns a copy of the run's recorder. The panel fetches it while it
// is open; the state snapshot does not carry it.
func (ge *GameEngine) History() *RunHistory {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.history.Clone()
}

// historyValues is every measure now. A resource's rate is a column once the
// run has met it (the resource is unlocked). Callers hold the lock.
func (ge *GameEngine) historyValues() map[string]float64 {
	snap := ge.Resources.Snapshot()
	owned := 0
	for _, c := range ge.Buildings.counts {
		owned += c
	}
	vals := map[string]float64{
		"population":  float64(ge.popLocked()),
		"food_rate":   snap["food"].Rate,
		"know_rate":   ge.researchPerTick,
		"faith":       snap["faith"].Amount,
		"morale":      ge.morale,
		"prod_all":    ge.bonusPoolLocked(ge.buildResolver(), "production_all").Applied,
		"tick_speed":  ge.tickSpeedBonus,
		"buildings":   float64(owned),
		"techs":       float64(ge.Research.ResearchedCount()),
		"expeditions": float64(ge.Military.returnsMade),
	}
	for key, r := range snap {
		if r.Unlocked {
			vals["rate."+key] = r.Rate
		}
	}
	return vals
}

// recordHistoryTicks takes the samples for the ticks (from, to] from the
// engine's state now, one per historyInterval, as live play takes one on each
// such tick. Offline catch-up calls it once a step, so the span holds the
// step's end values. Callers hold the write lock.
func (ge *GameEngine) recordHistoryTicks(from, to int) {
	var vals map[string]float64
	for t := from/historyInterval*historyInterval + historyInterval; t <= to; t += historyInterval {
		if vals == nil {
			vals = ge.historyValues()
		}
		ge.history.Record(t, vals)
	}
}
