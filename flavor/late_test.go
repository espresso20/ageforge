package flavor

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// late_test.go holds the tests for the digital-and-later voice. A review of a
// generated corpus found the late ages reading like a medieval village: door
// bars, lamps, servants, a council sitting before it was light. Two things
// caused it, and each has a test here: pre-modern set dressing sitting in pools
// the late ages can reach, and too few lines written for those ages at all.

// lateAges returns every age in the digital or cosmic bucket.
func lateAges() []string {
	var out []string
	for _, age := range config.AgeOrder() {
		if e, ok := eraOf(age); ok && (e == eraDigital || e == eraCosmic) {
			out = append(out, age)
		}
	}
	return out
}

// lateSetDressing is the pre-modern vocabulary a review found narrating the late
// ages: door bars, lamps, servants, a council at first light, a map case, wells,
// river crossings on foot. Each must stay an era marker whose homes stop before
// the digital bucket, so eraMarkers cannot be quietly loosened to let one back.
var lateSetDressing = []string{
	"bar", "barred", "lamp", "lamps", "lantern", "candle", "torch", "servant",
	"servants", "household", "council", "the well", "map case", "cart", "carts",
	"first light", "dawn", "dusk", "before it was light", "after dark",
	"outer houses", "runner", "chest", "chests", "bench", "coin", "ink",
	"water skin", "walking staff", "river", "hills", "camp", "crossing", "miles",
	"trail", "dried meat", "wet pack", "off a local", "walked with them",
	"paces", "no moon", "in the dirt", "dog", "lane", "washing", "mud", "spade",
}

// TestLateSetDressingIsMarked pins the list above to eraMarkers, then checks
// every authored string the late ages can reach (sentences and slot banks) for
// any of it. TestNoEraBleedInAuthoredText already enforces the markers; this
// test exists so that removing one from eraMarkers fails loudly here.
func TestLateSetDressingIsMarked(t *testing.T) {
	late := map[era]bool{eraDigital: true, eraCosmic: true}
	for _, w := range lateSetDressing {
		homes, ok := eraMarkers[w]
		if !ok {
			t.Errorf("%q is late-age set dressing but is not an era marker", w)
			continue
		}
		for _, h := range homes {
			if late[h] {
				t.Errorf("%q is marked as at home in the %s era; it must stop before the digital bucket", w, eraName(h))
			}
		}
	}
	checked := 0
	for _, u := range authoredProse() {
		if !u.eras[eraDigital] && !u.eras[eraCosmic] {
			continue
		}
		checked++
		for _, w := range lateSetDressing {
			if markerRE(w).MatchString(u.text) {
				t.Errorf("%s can fire in a late age but carries %q: %q", u.where, w, strings.TrimSpace(u.text))
			}
		}
	}
	t.Logf("%d late-reachable authored strings checked against %d set-dressing words", checked, len(lateSetDressing))
}

// minLateEraShare is the floor on era-voiced lines in digital-and-later ages.
const minLateEraShare = 0.50

// lateEraShare returns the share of drawn lines that came from an era-gated
// pool, for one Moment at one age, over the requests the engine actually sends,
// through a Stream (the engine's path).
func lateEraShare(m Moment, age string) float64 {
	gatedID := map[string]bool{}
	for _, tpl := range templatesFor(m) {
		if len(tpl.Eras) > 0 {
			gatedID[tpl.ID] = true
		}
	}
	rng := rand.New(rand.NewSource(int64(len(age))*977 + int64(m)))
	st := NewStream()
	reqs := gameRequests(m, age)
	const n = 3000
	gated := 0
	for i := 0; i < n; i++ {
		if gatedID[st.Generate(reqs[i%len(reqs)], rng).Template] {
			gated++
		}
	}
	return float64(gated) / n
}

// TestLateEraVoiceShare holds every digital-or-later age to at least half its
// drawn lines coming from a pool written for that era, for every Moment.
func TestLateEraVoiceShare(t *testing.T) {
	for _, age := range lateAges() {
		var parts []string
		for _, m := range Moments() {
			if !harbReachable(m, age) {
				continue // no false prophets this late, so nothing to discredit
			}
			share := lateEraShare(m, age)
			parts = append(parts, fmt.Sprintf("%s %2.0f%%", strings.TrimPrefix(strings.TrimPrefix(m.String(), "Expedition"), "Encounter"), share*100))
			if share < minLateEraShare {
				t.Errorf("%v at %s: %.0f%% of drawn lines are era-voiced; want >= %.0f%%",
					m, age, share*100, minLateEraShare*100)
			}
		}
		t.Logf("%-18s %s", age, strings.Join(parts, "  "))
	}
}
