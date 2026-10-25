package flavor

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// run_ending_test.go holds the tests specific to RunEnding, the line logged at
// every prestige. It shares every catalog lint (hygiene, era bleed, AI tells,
// register quotas, burstiness, house tics, near-duplicates). Like the harbinger
// Moments it fires once a run, so it is held to a pool floor per age rather
// than TestSkeletonFloors. Its clean-lines, no-figures and speaker checks run in
// harbinger_test.go, inside the harbinger tests of the same kind.

// runEndRequest is what the engine sends at age: the harbinger's name in the
// Cosmic Era, nothing before it.
func runEndRequest(age string) Request {
	return Request{Moment: RunEnding, Age: age, Subject: gameSubject(RunEnding, age)}
}

// TestRunEndingPoolFloors: every age a run can end in reaches a pool wider than
// the Stream's window with the request the game sends, and a careless caller
// that drops the Subject still stays well clear of a repeat.
func TestRunEndingPoolFloors(t *testing.T) {
	const bareFloor = 24
	reachable := 0
	for _, age := range config.AgeOrder() {
		if !runEndReachable(age) {
			continue
		}
		reachable++
		req := runEndRequest(age)
		if n := len(eligible(req)); n <= streamMemory {
			t.Errorf("RunEnding at %s: %d eligible skeletons; want > %d (the Stream window)", age, n, streamMemory)
		}
		req.Subject = ""
		if n := len(eligible(req)); n < bareFloor {
			t.Errorf("RunEnding at %s with no subject: %d eligible skeletons; want >= %d", age, n, bareFloor)
		}
	}
	if reachable != 10 {
		t.Errorf("a run can end in %d ages; want 10 (the Modern Age through the Transcendent Age)", reachable)
	}
}

// TestRunEndingVoiceByEra pins who is talking. Before the Cosmic Era the pool
// has no speaker at all: no {subject} line, and none of the Cosmic Era's lines.
// In the Cosmic Era the plain winding-down pools are out of reach, most of the
// pool belongs to the Cosmic Era, and every age reaches its own figure's lines.
func TestRunEndingVoiceByEra(t *testing.T) {
	cosmic := map[string]bool{}
	for _, a := range runEndCosmicAges {
		cosmic[a] = true
	}
	for _, age := range config.AgeOrder() {
		if !runEndReachable(age) {
			continue
		}
		pool := eligible(runEndRequest(age))
		own, eraSpecific := 0, 0
		for _, tp := range pool {
			isCosmic := strings.HasPrefix(tp.ID, "run_end_cosmic") || strings.HasPrefix(tp.ID, "run_end_age")
			if !cosmic[age] {
				if isCosmic || tp.Needs&needSubject != 0 {
					t.Errorf("%s fires at %s, before the Cosmic Era, but belongs to its harbingers: %q", tp.ID, age, authoredText(tp))
				}
				continue
			}
			if strings.HasPrefix(tp.ID, "run_end_plain") || strings.HasPrefix(tp.ID, "run_end_industrial") ||
				strings.HasPrefix(tp.ID, "run_end_digital") || strings.HasPrefix(tp.ID, "run_end_space") {
				t.Errorf("%s is a winding-down line but fires at %s", tp.ID, age)
			}
			if isCosmic {
				eraSpecific++
			}
			if strings.HasPrefix(tp.ID, "run_end_age_"+age) {
				own++
			}
		}
		if !cosmic[age] {
			continue
		}
		if own < 4 {
			t.Errorf("RunEnding at %s reaches only %d lines in its harbinger's own voice; want >= 4", age, own)
		}
		if share := float64(eraSpecific) / float64(len(pool)); share < 0.70 {
			t.Errorf("RunEnding at %s: only %.0f%% of the pool is the Cosmic Era's; want >= 70%%", age, share*100)
		}
	}
}
