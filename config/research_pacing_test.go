package config

import (
	"testing"
)

// TestNoGateAsksForKnowledge: knowledge left the age gates when the wonders
// got their keystones. An age's knowledge goes to research, and the one tech
// the advance needs is the wonder's.
func TestNoGateAsksForKnowledge(t *testing.T) {
	for _, a := range Ages() {
		if v, ok := a.ResourceReqs["knowledge"]; ok {
			t.Errorf("the gate into the %s asks for %v knowledge", a.Name, v)
		}
	}
}
