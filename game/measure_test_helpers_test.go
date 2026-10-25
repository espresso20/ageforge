package game

// setFaithMeasure puts ge's faith measure and its faith in place, for tests.
// The caller holds ge.mu or owns ge alone.
func (ge *GameEngine) setFaithMeasure(held float64, m FaithSave) {
	ge.Resources.UnlockResource("faith")
	ge.Resources.LoadAmounts(map[string]float64{"faith": held})
	ge.faithMeasure = m
}

// setCultureMeasure puts ge's culture measure and its culture in place, for
// tests. The caller holds ge.mu or owns ge alone.
func (ge *GameEngine) setCultureMeasure(held float64, m CultureSave) {
	ge.Resources.UnlockResource("culture")
	ge.Resources.LoadAmounts(map[string]float64{"culture": held})
	ge.cultureMeasure = m
}
