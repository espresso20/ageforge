package flavor

import "slices"

// StreamState is a Stream's memory in a form that can be saved. The memory is
// not just cosmetic: how many times Generate redraws depends on it, and every
// redraw spends rng draws, so a game that reloads with an empty Stream draws a
// different random stream from the one it saved. Round-tripping the state
// keeps a loaded game on the same stream.
type StreamState struct {
	Recent  []string `json:"recent,omitempty"`
	At      int      `json:"at,omitempty"`
	Topics  []string `json:"topics,omitempty"`
	TopicAt int      `json:"topic_at,omitempty"`
	Slots   []string `json:"slots,omitempty"`
	SlotAt  int      `json:"slot_at,omitempty"`
}

// State returns a copy of the Stream's memory, or nil for a Stream that has
// not generated anything yet (nothing to save).
func (s *Stream) State() *StreamState {
	if s == nil || !slices.ContainsFunc(s.recent, func(x string) bool { return x != "" }) {
		return nil // every line lands in recent, so an empty ring means nothing to save
	}
	return &StreamState{
		Recent:  append([]string(nil), s.recent...),
		At:      s.at,
		Topics:  append([]string(nil), s.topics...),
		TopicAt: s.tat,
		Slots:   append([]string(nil), s.slots...),
		SlotAt:  s.sat,
	}
}

// StreamFromState rebuilds a Stream from saved state. A nil state, or one
// whose rings are not the sizes this build uses (an edited save, or a save
// from a build with different window sizes), gives a fresh Stream.
func StreamFromState(st *StreamState) *Stream {
	s := NewStream()
	if st == nil ||
		len(st.Recent) != streamMemory || len(st.Topics) != topicMemory || len(st.Slots) != slotMemory ||
		st.At < 0 || st.At >= streamMemory || st.TopicAt < 0 || st.TopicAt >= topicMemory ||
		st.SlotAt < 0 || st.SlotAt >= slotMemory {
		return s
	}
	s.recent = append([]string(nil), st.Recent...)
	s.at = st.At
	s.topics = append([]string(nil), st.Topics...)
	s.tat = st.TopicAt
	s.slots = append([]string(nil), st.Slots...)
	s.sat = st.SlotAt
	return s
}
