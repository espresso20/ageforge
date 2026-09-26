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
// FLAVOR_DUMP_N sets lines per Moment per age (default 30); FLAVOR_DUMP_SEED the
// seed. Output lines are "<moment>\t<age>\t<template id>\t<text>".
func TestDumpCorpus(t *testing.T) {
	n := 30
	if v, err := strconv.Atoi(os.Getenv("FLAVOR_DUMP_N")); err == nil && v > 0 {
		n = v
	}
	seed := int64(20260926)
	if v, err := strconv.ParseInt(os.Getenv("FLAVOR_DUMP_SEED"), 10, 64); err == nil {
		seed = v
	}
	ages := []string{"primitive_age", "bronze_age", "medieval_age", "industrial_age", "digital_age", "quantum_age", "transcendent_age"}
	if v := os.Getenv("FLAVOR_DUMP_AGES"); v != "" {
		ages = strings.Split(v, ",")
	}
	for _, m := range Moments() {
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
				fmt.Printf("%v\t%s\t%s\t%s\n", m, age, res.Template, res.Text)
			}
		}
	}
}
