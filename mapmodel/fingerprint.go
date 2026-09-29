package mapmodel

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"math"
)

// Fingerprint is a short hash of everything the model derived: placement
// (town tiles, wonder plots, skyline lots), the clock and weather, the
// flows and the recap. Equal fingerprints mean two machines laid the same
// state out the same way to the last bit. The smoke harness records it next
// to the engine's state digest, so the cross-machine determinism check
// covers the maps too.
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
		w("l %s %d %x\n", l.Key, l.Tiles, math.Float64bits(l.Share))
	}
	tw := m.Town.World
	if tw != nil {
		w("world %d %d %s %v\n", tw.CX, tw.CY, tw.Name, tw.Coastal)
		hashTerrain(h, tw)
	}
	for _, t := range m.Town.Tiles {
		w("t %d %d %s %d %v %v %v\n", t.X, t.Y, t.Key, t.Ord, t.Legacy, t.Ruin, t.Fresh)
	}
	for _, t := range m.Town.Wonders {
		w("w %s %d %d %v\n", t.Key, t.At.X, t.At.Y, t.Built)
	}
	for _, d := range m.Skyline.Districts {
		w("d %d %d %d %v %d\n", d.Age, d.X0, d.W, d.Bay, d.BayX)
	}
	for _, l := range m.Skyline.Lots {
		w("s %s %d %d %d %d %x\n", l.Key, l.Copy, l.Row, l.X, l.Age, l.Seed)
	}
	for _, f := range m.Factions {
		w("f %s %d %v %d\n", f.Key, f.Site, f.Discovered, f.Relation)
	}
	for _, r := range m.Routes {
		w("r %s %s %v %s\n", r.Key, r.Civ, r.Disrupted, r.Mode)
	}
	a := m.Activity
	w("a %d %d %d %x %x\n", a.Routes, a.Staffed, a.Soldiers, math.Float64bits(a.Wealth), math.Float64bits(a.Traffic))
	w("c %s %x\n", m.Catastrophe.Pending, math.Float64bits(m.Catastrophe.Pressure))
	for _, s := range m.Flows.Full {
		w("full %s %x\n", s.Key, math.Float64bits(s.Fill))
	}
	for _, s := range m.Flows.Draining {
		w("drain %s %x\n", s.Key, math.Float64bits(s.EmptyIn))
	}
	w("worst %s\n", m.Flows.Worst)
	for _, it := range m.Recap.Items {
		w("n %d %s\n", it.Kind, it.Text)
	}
	return hex.EncodeToString(h.Sum(nil)[:8])
}

func hashTerrain(h hash.Hash, w *World) {
	buf := make([]byte, len(w.T))
	for i, t := range w.T {
		buf[i] = byte(t)
	}
	h.Write(buf)
	for _, s := range w.Sites {
		fmt.Fprintf(h, "%d,%d;", s.X, s.Y)
	}
}
