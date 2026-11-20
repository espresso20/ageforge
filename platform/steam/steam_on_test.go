//go:build steam

package steam

import (
	"reflect"
	"testing"
)

type fake struct{ calls []string }

func (f *fake) runCallbacks()           { f.calls = append(f.calls, "cb") }
func (f *fake) setAchievement(n string) { f.calls = append(f.calls, "set:"+n) }
func (f *fake) storeStats()             { f.calls = append(f.calls, "store") }

func reset(t *testing.T, m map[string]string) *fake {
	t.Helper()
	oldA, oldB, oldQ := achievements, backend, queue
	t.Cleanup(func() { achievements, backend, queue = oldA, oldB, oldQ })
	f := &fake{}
	achievements, backend, queue = m, f, nil
	return f
}

func TestUnlockQueuesOnlyMappedBadges(t *testing.T) {
	f := reset(t, map[string]string{"stone_age": "ACH_STONE"})
	Unlock("unmapped")
	Unlock("stone_age")
	Unlock("stone_age")
	if len(f.calls) != 0 {
		t.Fatalf("Unlock called Steam: %v", f.calls)
	}
	Frame()
	want := []string{"cb", "set:ACH_STONE", "set:ACH_STONE", "store"}
	if !reflect.DeepEqual(f.calls, want) {
		t.Fatalf("calls = %v, want %v", f.calls, want)
	}
	f.calls = nil
	Frame()
	if want := []string{"cb"}; !reflect.DeepEqual(f.calls, want) {
		t.Fatalf("second frame = %v, want %v", f.calls, want)
	}
}

func TestEmptyMapUnlocksNothing(t *testing.T) {
	f := reset(t, map[string]string{})
	Unlock("stone_age")
	Frame()
	if want := []string{"cb"}; !reflect.DeepEqual(f.calls, want) {
		t.Fatalf("calls = %v, want %v", f.calls, want)
	}
}

func TestFrameWithoutSteamKeepsQueue(t *testing.T) {
	reset(t, map[string]string{"a": "A"})
	backend = nil
	Unlock("a")
	Frame()
	if len(queue) != 1 {
		t.Fatalf("queue = %v, want it kept until Steam starts", queue)
	}
}
