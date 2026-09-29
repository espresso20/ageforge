package game

import (
	"maps"
	"time"

	"github.com/espresso20/ageforge/config"
)

// SessionMark is the game as its save left it: captured when a save loads,
// before offline catch-up runs, so it is what the player saw when their last
// session ended. The maps diff the live state against it for their "since
// your last visit" recap. It is never saved (the save itself is the record)
// and has no effect on the simulation.
type SessionMark struct {
	Tick      int
	SavedAt   time.Time
	Age       string
	Buildings map[string]int
	Ruins     map[string]int
	Pop       int
	// Civs holds each discovered civ's standing: "war" when at war,
	// otherwise its diplomatic status.
	Civs        map[string]string
	Routes      int
	Harbinger   string // the harbinger's name, "" when none was present
	Expeditions int    // expeditions completed
}

// sessionMarkLocked captures the mark from the restored state. Caller holds
// the write lock.
func (ge *GameEngine) sessionMarkLocked(savedAt time.Time) *SessionMark {
	m := &SessionMark{
		Tick: ge.tick, SavedAt: savedAt, Age: ge.age,
		Buildings: ge.Buildings.GetAll(), Ruins: ge.Buildings.GetAllRuins(),
		Pop: ge.Workers.TotalPop(), Civs: map[string]string{},
		Routes: ge.Trade.ActiveRouteCount(), Expeditions: ge.Military.completedCount,
	}
	for k, n := range m.Buildings {
		if n == 0 {
			delete(m.Buildings, k)
		}
	}
	for k, f := range ge.Diplomacy.GetFactionsForSave() {
		if !f.Discovered {
			continue
		}
		st := f.Status
		if f.AtWar {
			st = "war"
		}
		m.Civs[k] = st
	}
	if h := ge.harbinger; h != nil {
		if def, ok := config.HarbingerFor(h.Age); ok {
			m.Harbinger = def.Name
		}
	}
	return m
}

// clone returns a copy that shares no maps with m (nil stays nil).
func (m *SessionMark) clone() *SessionMark {
	if m == nil {
		return nil
	}
	c := *m
	c.Buildings = maps.Clone(m.Buildings)
	c.Ruins = maps.Clone(m.Ruins)
	c.Civs = maps.Clone(m.Civs)
	return &c
}
