package game

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
