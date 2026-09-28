package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// StateJSON is the engine's full save snapshot as JSON with the wall-clock
// fields zeroed, so two engines that played the same run produce the same
// bytes. JSON writes each float64 as its shortest round-trip decimal, so equal
// bytes mean equal bits. It is for determinism checks (StateDigest, the smoke
// harness's cross-architecture fingerprint), not for saving.
func (ge *GameEngine) StateJSON() ([]byte, error) {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	s := ge.buildSaveSnapshot()
	s.Timestamp = time.Time{}
	if s.Stats != nil {
		s.Stats.GameStarted = time.Time{}
	}
	return json.Marshal(s)
}

// StateDigest is a short hash of StateJSON: equal digests mean the two
// engines are in the same state to the last bit. A state JSON cannot encode
// (a NaN or an infinity) digests as "unencodable", which never matches a
// healthy run.
func (ge *GameEngine) StateDigest() string {
	b, err := ge.StateJSON()
	if err != nil {
		return "unencodable"
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}
