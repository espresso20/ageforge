package mapmodel

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
)

// Fingerprint is a short hash of everything the model derived: the town's
// name, the clock and weather, the buildings, the flows and the recap. Equal
// fingerprints mean two machines laid the same state out the same way to the
// last bit.
func (m *Model) Fingerprint() string {
	h := sha256.New()
	w := func(format string, a ...any) { fmt.Fprintf(h, format, a...) }
	w("%d|%d|%s|%d\n", m.Seed, m.Tick, m.Age, m.Epoch)
	c := m.Clock
	w("clock %d %x %d:%d %x %x %x %s|%d\n", c.Day, math.Float64bits(c.TOD), c.Hour, c.Minute,
		math.Float64bits(c.Daylight), math.Float64bits(c.Twilight), math.Float64bits(c.Night), c.Phase, m.Weather.Kind)
	for _, b := range m.Buildings {
		w("b %s %d %d %d %d %x %v %x %d\n", b.Key, b.Count, b.Ruins, b.Workers, b.Capacity,
			math.Float64bits(b.Staffing), b.Legacy, math.Float64bits(b.Rate), b.Delta)
	}
	for _, l := range m.Lineages {
		w("l %s %x\n", l.Key, math.Float64bits(l.Share))
	}
	w("town %s\n", m.TownName)
	for _, f := range m.Factions {
		w("f %s %v %d\n", f.Key, f.Discovered, f.Relation)
	}
	for _, r := range m.Routes {
		w("r %s %s %v\n", r.Key, r.Civ, r.Disrupted)
	}
	a := m.Activity
	w("a %d %d %d %x %x\n", a.Routes, a.Staffed, a.Soldiers, math.Float64bits(a.Wealth), math.Float64bits(a.Traffic))
	w("c %s %x\n", m.Catastrophe.Pending, math.Float64bits(m.Catastrophe.Pressure))
	for _, s := range m.Flows.Draining {
		w("drain %s %x\n", s.Key, math.Float64bits(s.EmptyIn))
	}
	w("worst %s\n", m.Flows.Worst)
	for _, it := range m.Recap.Items {
		w("n %d %s\n", it.Kind, it.Text)
	}
	return hex.EncodeToString(h.Sum(nil)[:8])
}
