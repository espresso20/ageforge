package flavor

import (
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// run_ending_test.go holds the tests specific to RunEnding, the line logged at
// every prestige. It shares every catalog lint (hygiene, era bleed, AI tells,
// register quotas, burstiness, house tics, near-duplicates). Like the harbinger
// Moments it fires once a run, so it is held to a pool floor per age rather
// than TestSkeletonFloors.

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

// TestRunEndingLinesAreClean runs every age through a Stream with the request
// the game sends (and with no subject): hygiene, no digit, no foreign era word,
// and no other harbinger's vocabulary.
func TestRunEndingLinesAreClean(t *testing.T) {
	speakers := map[string]*regexp.Regexp{}
	for age, words := range speakerWords {
		speakers[age] = regexp.MustCompile(`(?i)\b(` + strings.Join(words, "|") + `)\b`)
	}
	for _, age := range config.AgeOrder() {
		e, _ := eraOf(age)
		foreign := foreignMarkers(e)
		rng := rand.New(rand.NewSource(int64(len(age)) * 6151))
		st := NewStream()
		for _, subj := range []string{gameSubject(RunEnding, age), ""} {
			req := Request{Moment: RunEnding, Age: age, Subject: subj}
			for i := 0; i < 60; i++ {
				line := st.Line(req, rng)
				if p := hygiene(line); p != "" {
					t.Fatalf("RunEnding at %s (%+v): %s\n    %q", age, req, p, line)
				}
				if digitRE.MatchString(line) {
					t.Fatalf("RunEnding at %s: a digit in the line: %q", age, line)
				}
				scan := callerText(line, req)
				if hit := foreign.FindString(scan); hit != "" {
					t.Fatalf("era bleed: RunEnding at %s produced %q, which carries %q", age, line, hit)
				}
				for owner, re := range speakers {
					if owner == age {
						continue
					}
					if w := re.FindString(scan); w != "" && !speakerAlsoAt(speakers, age, w) {
						t.Errorf("RunEnding at %s uses %q, which belongs to the %s harbinger: %q", age, w, owner, line)
					}
				}
			}
		}
	}
}

// TestRunEndingTemplatesUseNoFigures pins the digit rule at the source.
func TestRunEndingTemplatesUseNoFigures(t *testing.T) {
	banned := []string{"{amt}", "{amt_res}", "{n}", "{ticks}", "{res}", "{res_stores}", "{res_haul}"}
	for _, tp := range templatesFor(RunEnding) {
		text := authoredText(tp)
		if digitRE.MatchString(text) {
			t.Errorf("%s carries a digit: %q", tp.ID, text)
		}
		for _, b := range banned {
			if strings.Contains(text, b) {
				t.Errorf("%s uses %s; the game prints the points", tp.ID, b)
			}
		}
	}
}

// TestRunEndingDeterminism: same seed and request, same lines; a different
// seed, different lines.
func TestRunEndingDeterminism(t *testing.T) {
	const n = 60
	run := func(req Request, seed int64) []string {
		rng := rand.New(rand.NewSource(seed))
		st := NewStream()
		out := make([]string, n)
		for i := range out {
			out[i] = st.Line(req, rng)
		}
		return out
	}
	for _, age := range []string{"modern_age", "cyberpunk_age", "space_age", "galactic_age", "transcendent_age"} {
		req := runEndRequest(age)
		a, b, c := run(req, 9001), run(req, 9001), run(req, 9002)
		differ := 0
		for i := range a {
			if a[i] != b[i] {
				t.Fatalf("RunEnding at %s: line %d differs under the same seed", age, i)
			}
			if a[i] != c[i] {
				differ++
			}
		}
		if differ < n*7/10 {
			t.Errorf("RunEnding at %s: seeds 9001 and 9002 agree on %d of %d lines", age, n-differ, n)
		}
	}
}
