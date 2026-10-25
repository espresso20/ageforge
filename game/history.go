package game

import "slices"

const historyMaxSamples = 300
const historySampleInterval = 10 // every 10 ticks (~20s at 1x, ~13s at 1.5x)

// HistorySample is a point-in-time snapshot of key metrics.
type HistorySample struct {
	Tick       int     `json:"tick"`
	Population float64 `json:"population"`
	FoodRate   float64 `json:"food_rate"`
	KnowRate   float64 `json:"know_rate"`
	Faith      float64 `json:"faith"`
	ProdAll    float64 `json:"prod_all"`
	TickSpeed  float64 `json:"tick_speed"`
	Morale     float64 `json:"morale"`
	AgeOrder   int     `json:"age_order"`
}

// AgeMarker records when an age advance occurred.
type AgeMarker struct {
	Tick    int    `json:"tick"`
	AgeName string `json:"age_name"`
}

// HistoryCollector accumulates periodic metric samples and age-advance events.
type HistoryCollector struct {
	Samples    []HistorySample `json:"samples"`
	AgeMarkers []AgeMarker     `json:"age_markers"`
}

// Clone returns a deep copy (nil for nil), for snapshots that must not move
// as the collector keeps sampling.
func (h *HistoryCollector) Clone() *HistoryCollector {
	if h == nil {
		return nil
	}
	return &HistoryCollector{
		Samples:    slices.Clone(h.Samples),
		AgeMarkers: slices.Clone(h.AgeMarkers),
	}
}

// NewHistoryCollector returns an empty collector.
func NewHistoryCollector() *HistoryCollector {
	return &HistoryCollector{}
}

// Sample records a snapshot. Called from doTick() every historySampleInterval ticks.
func (h *HistoryCollector) Sample(tick int, s HistorySample) {
	if h == nil {
		return
	}
	s.Tick = tick
	if len(h.Samples) >= historyMaxSamples {
		h.Samples = h.Samples[1:] // ring: drop oldest
	}
	h.Samples = append(h.Samples, s)
}

// MarkAge records an age advance event and prunes markers outside the sample window.
func (h *HistoryCollector) MarkAge(tick int, ageName string) {
	if h == nil {
		return
	}
	h.AgeMarkers = append(h.AgeMarkers, AgeMarker{Tick: tick, AgeName: ageName})
	// keep only markers that fall within the current sample window
	if len(h.Samples) > 0 {
		oldest := h.Samples[0].Tick
		kept := h.AgeMarkers[:0]
		for _, m := range h.AgeMarkers {
			if m.Tick >= oldest {
				kept = append(kept, m)
			}
		}
		h.AgeMarkers = kept
	}
}

// recordHistorySample takes the sample for tick from the engine's state now,
// as live play does every historySampleInterval ticks. It reads engine fields
// directly, with no GetState() call. Callers hold the write lock.
func (ge *GameEngine) recordHistorySample(tick int) {
	ageOrderMap := ge.progress.GetAgeOrder()
	ageOrder := ageOrderMap[ge.age]
	snap := ge.Resources.Snapshot()
	foodRate := 0.0
	knowRate := 0.0
	if fr, ok := snap["food"]; ok {
		foodRate = fr.Rate
	}
	knowRate = ge.researchPerTick
	faith := snap["faith"].Amount
	ge.History.Sample(tick, HistorySample{
		Population: float64(ge.popLocked()),
		FoodRate:   foodRate,
		KnowRate:   knowRate,
		Faith:      faith,
		ProdAll:    ge.bonusPoolLocked(ge.buildResolver(), "production_all").Applied,
		TickSpeed:  ge.tickSpeedBonus,
		Morale:     ge.morale,
		AgeOrder:   ageOrder,
	})
}

// sampleOfflineTicks takes the samples live play would have taken on the ticks
// (from, to] that offline catch-up passed, so the samples stay evenly spaced in
// ticks. The History panel draws samples by index and places age markers by
// tick, and the two agree only over an even spacing. Each sample is taken at the
// end of the step it falls in. Only the ticks the ring keeps are taken: the
// last historyMaxSamples samples of the span, which ends at end. Earlier ones
// would be dropped before anything reads them.
func (ge *GameEngine) sampleOfflineTicks(from, to, end int) {
	keep := end - historyMaxSamples*historySampleInterval
	for t := max(from, keep) + 1; t <= to; t++ {
		if t%historySampleInterval == 0 {
			ge.recordHistorySample(t)
		}
	}
}
