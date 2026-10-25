package fixture

import (
	"slices"
	"testing"

	"github.com/espresso20/ageforge/config"
)

func TestReachedEras(t *testing.T) {
	st := State(Options{Age: "transcendent_age", Seed: 7})
	ReachedEras(&st, len(config.Epochs()))
	if st.Stats.AgesReached != nil {
		t.Errorf("every era: the record was cut to %v", st.Stats.AgesReached)
	}
	ReachedEras(&st, 2)
	var want []string
	for _, e := range config.Epochs()[:2] {
		want = append(want, e.Ages...)
	}
	want = append(want, "transcendent_age")
	if !slices.Equal(st.Stats.AgesReached, want) {
		t.Errorf("two eras: the record is %v, want %v", st.Stats.AgesReached, want)
	}
}
