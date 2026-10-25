package config

import (
	"testing"

	"github.com/rivo/uniseg"
)

// TestNoLaneHoldsMoreThanThreeTechsAnAge: a lane takes three techs an age at
// most. Each tech's age, lane, kind and prerequisites are pinned by name in
// testdata/golden_tables.json.
func TestNoLaneHoldsMoreThanThreeTechsAnAge(t *testing.T) {
	perLane := map[string]int{}
	for _, def := range Technologies() {
		perLane[def.Age+" "+def.Lane]++
	}
	for laneOfAge, n := range perLane {
		if n > 3 {
			t.Errorf("%s holds %d techs: a lane takes three an age at most", laneOfAge, n)
		}
	}
}

// TestDesignedTechsCarryTheirOwnArt: every tech has an emblem and a letter
// code of its own, written on the tech and not borrowed
// from its lane or made from its name by default. An emblem is one glyph
// that takes exactly one cell, and no two techs of a lane share one; the
// code's first letter is the tech's letter in the plain glyph tier.
func TestDesignedTechsCarryTheirOwnArt(t *testing.T) {
	raw := map[string]TechDef{}
	for _, def := range rawTechnologies() {
		raw[def.Key] = def
	}
	inLane := map[string]map[string]string{}
	for _, def := range Technologies() {
		if raw[def.Key].Emblem == "" {
			t.Errorf("%s sets no emblem of its own: it would wear its lane's", def.Key)
		}
		rs := []rune(def.Emblem)
		if len(rs) != 1 || uniseg.StringWidth(def.Emblem) != 1 {
			t.Errorf("%s has the emblem %q: %d glyphs, %d cells wide; an emblem is one glyph in one cell", def.Key, def.Emblem, len(rs), uniseg.StringWidth(def.Emblem))
		}
		if len(rs) == 1 && (rs[0] >= 0x1F000 || rs[0] == 0xFE0F) {
			t.Errorf("%s has the emblem %q, from the emoji planes: terminals draw those two cells wide", def.Key, def.Emblem)
		}
		if inLane[def.Lane] == nil {
			inLane[def.Lane] = map[string]string{}
		}
		if other, taken := inLane[def.Lane][def.Emblem]; taken {
			t.Errorf("%s and %s, both in the %s lane, share the emblem %q", other, def.Key, def.Lane, def.Emblem)
		}
		inLane[def.Lane][def.Emblem] = def.Key
		if !validTechCode(def.Code) || def.Code[0] < 'A' || def.Code[0] > 'Z' {
			t.Errorf("%s has the code %q: no letter for the plain glyph tier", def.Key, def.Code)
		}
	}
}
