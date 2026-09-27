//go:build flavordump

package flavor

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
)

// dump_test.go is a review tool, not a test. It prints a sample corpus the way a
// player meets it (consecutive lines through one Stream, per Moment, per age) so
// a person can read the catalog in bulk before shipping a change to it:
//
//	go test -tags flavordump ./flavor -run TestDumpCorpus -v > corpus.txt
//
// FLAVOR_DUMP_N sets lines per Moment per age (default 30; harbinger Moments
// default to 5 per age, per kind, since each firing writes one line);
// FLAVOR_DUMP_SEED the seed; FLAVOR_DUMP_AGES a comma-separated age list;
// FLAVOR_DUMP_ONLY=harbinger or =core to dump one family of Moments.
// Output lines are "<moment>\t<age>\t<kind>\t<template id>\t<text>".
func TestDumpCorpus(t *testing.T) {
	n, nSet := 30, false
	if v, err := strconv.Atoi(os.Getenv("FLAVOR_DUMP_N")); err == nil && v > 0 {
		n, nSet = v, true
	}
	seed := int64(20260926)
	if v, err := strconv.ParseInt(os.Getenv("FLAVOR_DUMP_SEED"), 10, 64); err == nil {
		seed = v
	}
	ages := []string{"primitive_age", "bronze_age", "medieval_age", "industrial_age", "digital_age", "quantum_age", "transcendent_age"}
	// The harbinger sample walks the humanizer-review ages: one per speaker
	// family, early to late.
	harbAges := []string{"primitive_age", "bronze_age", "medieval_age", "industrial_age", "atomic_age", "cyberpunk_age", "galactic_age", "transcendent_age"}
	if v := os.Getenv("FLAVOR_DUMP_AGES"); v != "" {
		ages = strings.Split(v, ",")
		harbAges = ages
	}
	only := os.Getenv("FLAVOR_DUMP_ONLY")
	for _, m := range Moments() {
		harb := isRare(m)
		if (only == "harbinger" && !harb) || (only == "core" && harb) {
			continue
		}
		if harb {
			per := 5
			if nSet {
				per = n
			}
			for ai, age := range harbAges {
				if !harbReachable(m, age) {
					continue
				}
				rng := rand.New(rand.NewSource(seed + int64(ai)))
				st := NewStream()
				for _, kind := range harbKinds(m, age) {
					for i := 0; i < per; i++ {
						req := Request{Moment: m, Age: age, Kind: kind, Subject: gameSubject(m, age)}
						res := st.Generate(req, rng)
						fmt.Printf("%v\t%s\t%s\t%s\t%s\n", m, age, kind, res.Template, res.Text)
					}
				}
			}
			continue
		}
		for ai, age := range ages {
			// A different stream per age, so the corpus shows more of the
			// catalog than one seed replayed through a shared ungated pool.
			rng := rand.New(rand.NewSource(seed + int64(ai)))
			st := NewStream()
			kinds := gameKinds[m]
			for i := 0; i < n; i++ {
				req := Request{
					Moment: m, Age: age, Tone: allTones[i%len(allTones)], Kind: kinds[i%len(kinds)],
					Subject: "Merchant Guild", Resource: "gold", Amount: 120,
				}
				res := st.Generate(req, rng)
				fmt.Printf("%v\t%s\t%s\t%s\t%s\n", m, age, req.Kind, res.Template, res.Text)
			}
		}
	}
}
